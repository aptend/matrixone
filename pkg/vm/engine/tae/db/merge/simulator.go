// Copyright 2025 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package merge

import (
	"context"
	"fmt"
	"iter"
	"sync"
	"time"

	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/catalog"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/index"
	"golang.org/x/exp/constraints"

	"github.com/jonboulle/clockwork"
	"github.com/tidwall/btree"
)

// region: Clock

type Clock interface {
	clockwork.Clock
	Until(t time.Time) time.Duration
}

type Ticker interface {
	clockwork.Ticker
}

type stdClock struct {
	clockwork.Clock
}

func NewStdClock() *stdClock {
	return &stdClock{
		Clock: clockwork.NewRealClock(),
	}
}

func (c stdClock) Until(t time.Time) time.Duration {
	return time.Until(t)
}

type fakeClock struct {
	clockwork.FakeClock
}

func newFakeClock() *fakeClock {
	return &fakeClock{
		FakeClock: clockwork.NewFakeClock(),
	}
}

func (c *fakeClock) Until(t time.Time) time.Duration {
	return t.Sub(c.FakeClock.Now())
}

// endregion: Clock

// region: executor

type SExecutor struct {
	clock    Clock
	scatalog *SCatalog

	// stats
	dataMergedSize      int64
	tombstoneMergedSize int64
}

func NewSExecutor(c Clock, scatalog *SCatalog) *SExecutor {
	return &SExecutor{
		clock:    c,
		scatalog: scatalog,
	}
}

type objsInfo struct {
	totalOsize, totalCsize, totalRowCount int
	avgROsize, avgRCsize                  int
}

func sizeInfo(objs []*objectio.ObjectStats) (info objsInfo) {
	for _, stat := range objs {
		info.totalOsize += int(stat.OriginSize())
		info.totalCsize += int(stat.Size())
		info.totalRowCount += int(stat.Rows())
	}

	info.avgROsize = info.totalOsize / info.totalRowCount
	info.avgRCsize = info.totalCsize / info.totalRowCount
	return
}

func mergeDataLocked(
	stable *STable,
	task mergeTask,
	clock Clock,
) (newObjs []SData) {

	info := sizeInfo(task.objs)

	zm := index.NewZM(task.objs[0].SortKeyZoneMap().GetType(), 0)
	dels := 0
	for _, stat := range task.objs {
		zm.Update(stat.SortKeyZoneMap().GetMin())
		zm.Update(stat.SortKeyZoneMap().GetMax())
		for _, tombstone := range stable.tombstone {
			dels += tombstone.distro[stat.ObjectLocation().ObjectId()]
		}
	}

	leftRows := info.totalRowCount - dels
	objRows := common.DefaultMaxOsizeObjBytes/info.avgROsize + 100
	mergedSize := 0

	rowSplit := make([]int, 0)

	for leftRows > 0 {
		mergedSize += objRows * info.avgROsize
		if info.totalOsize-mergedSize > common.DefaultMaxOsizeObjBytes {
			rowSplit = append(rowSplit, objRows)
			leftRows -= objRows
		} else {
			rowSplit = append(rowSplit, leftRows)
			leftRows = 0
		}
	}

	createdSegId := objectio.NewSegmentid()
	zmSplit := splitZM(zm, rowSplit)

	for i, zm := range zmSplit {
		newObj := SData{
			stats:      objectio.NewObjectStats(),
			createTime: types.BuildTS(clock.Now().UnixNano(), 0),
		}
		row := rowSplit[i]
		name := objectio.BuildObjectName(createdSegId, uint16(i))
		objectio.SetObjectStatsObjectName(newObj.stats, name)
		objectio.SetObjectStatsOriginSize(newObj.stats, uint32(row*info.avgROsize))
		objectio.SetObjectStatsSize(newObj.stats, uint32(row*info.avgRCsize))
		objectio.SetObjectStatsSortKeyZoneMap(newObj.stats, zm)
		objectio.SetObjectStatsRowCnt(newObj.stats, uint32(row))

		if task.level > 0 ||
			newObj.stats.OriginSize() > common.DefaultMinOsizeQualifiedBytes {
			if task.level < 7 {
				newObj.stats.SetLevel(int8(task.level + 1))
			} else {
				newObj.stats.SetLevel(7)
			}
		}

		newObjs = append(newObjs, newObj)
	}

	for _, obj := range task.objs {
		delete(stable.data, obj.ObjectLocation().ObjectId())
	}

	for _, obj := range newObjs {
		stable.data[obj.stats.ObjectLocation().ObjectId()] = obj
	}

	logutil.Infof("mergeDataLocked: %d -> %d", len(task.objs), len(newObjs))

	return
}

func mergeTombstoneLocked(
	stable *STable,
	task mergeTask,
	clock Clock,
) (newTombstones []STombstone) {

	info := sizeInfo(task.objs)

	// target disto is tombstone waiting to be flushed
	targetDistro := btree.NewBTreeG(func(a, b struct {
		oid   objectio.ObjectId
		count int
	}) bool {
		return a.oid.Compare(&b.oid) < 0
	})

	// only keep the tombstone targeting alive data object
	for _, stat := range task.objs {
		obj := stable.tombstone[stat.ObjectLocation().ObjectId()]
		for dataid, delcnt := range obj.distro {
			if _, ok := stable.data[dataid]; ok {
				targetDistro.Set(struct {
					oid   objectio.ObjectId
					count int
				}{oid: dataid, count: delcnt})
			}
		}
	}

	iter := targetDistro.Iter()
	defer iter.Release()

	currentObjRow := 0
	currentObjOsize := 0
	currentObjCsize := 0
	mergedSize := 0

	createdSegId := objectio.NewSegmentid()
	createdObjIdx := uint16(0)

	var newTombstone STombstone
	var currentZM index.ZM

	writeTombstone := func() {
		// set stat
		newStat := newTombstone.stats
		objectio.SetObjectStatsOriginSize(newStat, uint32(currentObjOsize))
		objectio.SetObjectStatsSize(newStat, uint32(currentObjCsize))
		objectio.SetObjectStatsRowCnt(newStat, uint32(currentObjRow))
		objectio.SetObjectStatsSortKeyZoneMap(newStat, currentZM)

		newTombstones = append(newTombstones, newTombstone)
		newTombstone = STombstone{}
		createdObjIdx++
		mergedSize += currentObjOsize
		currentObjRow = 0
		currentObjOsize = 0
		currentObjCsize = 0
		currentZM = index.NewZM(types.T_Rowid, 0)
	}

	for iter.Next() {
		if newTombstone.distro == nil {
			newTombstone.stats = objectio.NewObjectStats()
			objectio.SetObjectStatsObjectName(
				newTombstone.stats,
				objectio.BuildObjectName(createdSegId, createdObjIdx),
			)
			newTombstone.createTime = types.BuildTS(clock.Now().UnixNano(), 0)
			newTombstone.distro = make(map[objectio.ObjectId]int)
			currentZM = index.NewZM(types.T_Rowid, 0)
		}
		item := iter.Item()
		currentObjRow += item.count
		currentObjOsize += item.count * info.avgROsize
		currentObjCsize += item.count * info.avgRCsize
		newTombstone.distro[item.oid] = item.count

		// tombstone's zm does not matter for merging
		currentZM.Update(types.NewRowIDWithObjectIDBlkNumAndRowID(item.oid, 0, 0))
		currentZM.Update(types.NewRowIDWithObjectIDBlkNumAndRowID(item.oid, 2, 8192))

		if currentObjRow*info.avgROsize > common.DefaultMaxOsizeObjBytes &&
			info.totalOsize-mergedSize > common.DefaultMaxOsizeObjBytes {
			writeTombstone()
		}
	}

	if currentObjRow > 0 {
		writeTombstone()
	}

	for _, obj := range task.objs {
		delete(stable.tombstone, obj.ObjectLocation().ObjectId())
	}

	for _, obj := range newTombstones {
		stable.tombstone[obj.stats.ObjectLocation().ObjectId()] = obj
	}

	return newTombstones
}

func (e *SExecutor) ExecuteFor(target catalog.MergeTable, task mergeTask) bool {
	stable := target.(*STable)
	stable.Lock()
	defer stable.Unlock()

	newCount := 0
	if task.isTombstone {
		newCount = len(mergeTombstoneLocked(stable, task, e.clock))
	} else {
		newCount = len(mergeDataLocked(stable, task, e.clock))
	}

	// baseline: 2MB oringnal size -> 150ms
	taskCost := float64(time.Millisecond) * 150 *
		float64(task.oSize) / common.Const1MBytes / 2

	e.clock.AfterFunc(time.Duration(taskCost), func() {
		if task.doneCB != nil {
			task.doneCB.f()
		}
		for range newCount {
			e.scatalog.mergeSched.OnCreateNonAppendObject(target)
		}
	})

	return true
}

func updateNumberTypeZM[T constraints.Integer | constraints.Float](
	zmSplit []index.ZM,
	ratio []float64,
	rowsSplit []int,
	l, r T,
) {
	span := float64(r - l)
	for i := range zmSplit {
		zmSplit[i].Update(l)
		if i == len(zmSplit)-1 {
			zmSplit[i].Update(r)
		} else {
			piece := T(ratio[i] * span)
			// update the right bound only if the rowscount is able to contain
			// all variants of the current interval.
			// otherwise, let's make it a constant object for simplicity
			// and let the last object hold the rest
			if float64(rowsSplit[i]) > float64(piece) {
				l = l + piece
				zmSplit[i].Update(l)
			}
			l += 1
		}
	}
}

func splitZM(zm index.ZM, rowsPerObj []int) []index.ZM {
	ratio := make([]float64, len(rowsPerObj))
	total := 0

	for _, row := range rowsPerObj {
		total += row
	}

	for i, row := range rowsPerObj {
		ratio[i] = float64(row) / float64(total)
	}

	zmSplit := make([]index.ZM, len(rowsPerObj))
	for i := range zmSplit {
		zmSplit[i] = index.NewZM(zm.GetType(), 0)
	}

	_min := zm.GetMin()
	_max := zm.GetMax()

	// case types.T_int8:
	// case types.T_int16:
	// case types.T_int32:
	// case types.T_int64:
	// case types.T_uint8:
	// case types.T_uint16:
	// case types.T_uint32:
	// case types.T_uint64:
	// case types.T_float32:
	// case types.T_float64:
	// case types.T_date:
	// case types.T_time:
	// case types.T_datetime:
	// case types.T_timestamp:
	// case types.T_enum:
	// case types.T_decimal64:
	// case types.T_decimal128:
	// case types.T_json:
	// case types.T_blob:
	// case types.T_text:
	// case types.T_char:
	switch zm.GetType() {
	case types.T_int32:
		l, r := _min.(int32), _max.(int32)
		updateNumberTypeZM(zmSplit, ratio, rowsPerObj, l, r)
	default:
		panic(fmt.Sprintf("unsupported type: %s", zm.GetType()))
	}

	return zmSplit
}

// endregion: executor

// region: Catalog

type SCatalog struct {
	mergeSched catalog.MergeNotifierOnCatalog
	hero       *STable
}

func (c *SCatalog) AddData(data SData) {
	c.hero.Lock()
	defer c.hero.Unlock()
	c.hero.data[data.stats.ObjectLocation().ObjectId()] = data
	c.mergeSched.OnCreateNonAppendObject(c.hero)
}

func (c *SCatalog) AddTombstone(tombstone STombstone) {
	c.hero.Lock()
	defer c.hero.Unlock()
	c.hero.tombstone[tombstone.stats.ObjectLocation().ObjectId()] = tombstone
	c.mergeSched.OnCreateNonAppendObject(c.hero)
}

func NewSCatalog() *SCatalog {
	return &SCatalog{
		hero: &STable{
			id:        1000,
			desc:      "1000-merge-hero",
			data:      make(map[objectio.ObjectId]SData),
			tombstone: make(map[objectio.ObjectId]STombstone),
		},
	}
}

func (c *SCatalog) InitSource() iter.Seq[catalog.MergeTable] {
	return func(yield func(catalog.MergeTable) bool) {
		yield(c.hero)
	}
}

func (c *SCatalog) SetMergeNotifier(scheduler catalog.MergeNotifierOnCatalog) {
	c.mergeSched = scheduler
}

func (c *SCatalog) GetMergeSettingsBatchFn() func() (*batch.Batch, func()) {
	return func() (*batch.Batch, func()) {
		return nil, func() {}
	}
}

type STable struct {
	sync.RWMutex
	id   uint64
	desc string

	data      map[objectio.ObjectId]SData
	tombstone map[objectio.ObjectId]STombstone
}

func (t *STable) ID() uint64 { return t.id }

func (t *STable) GetNameDesc() string { return t.desc }

func (t *STable) HasDropCommitted() bool { return false }

func (t *STable) IsSpecialBigTable() bool { return false }

func (t *STable) IterDataItem() iter.Seq[catalog.MergeDataItem] {
	return func(yield func(catalog.MergeDataItem) bool) {
		t.RLock()
		defer t.RUnlock()
		for _, obj := range t.data {
			yield(&obj)
		}
	}
}

func (t *STable) IterTombstoneItem() iter.Seq[catalog.MergeTombstoneItem] {
	return func(yield func(catalog.MergeTombstoneItem) bool) {
		t.RLock()
		defer t.RUnlock()
		for _, obj := range t.tombstone {
			yield(&obj)
		}
	}
}

type SData struct {
	stats      *objectio.ObjectStats
	createTime types.TS
}

func (o *SData) GetObjectStats() *objectio.ObjectStats {
	return o.stats
}

func (o *SData) GetCreatedAt() types.TS {
	return o.createTime
}

type STombstone struct {
	SData
	distro map[objectio.ObjectId]int
}

func (o *STombstone) ForeachRowid(
	ctx context.Context,
	reuseBatch any,
	each func(rowid types.Rowid, isNull bool, rowIdx int) error,
) error {
	for id, count := range o.distro {
		rowid := types.NewRowIDWithObjectIDBlkNumAndRowID(id, 0, 0)
		for i := 0; i < count; i++ {
			each(rowid, false, i)
		}
	}
	return nil
}

func (o *STombstone) MakeBufferBatch() (any, func()) {
	return nil, func() {}
}

// endregion: Catalog & IO
