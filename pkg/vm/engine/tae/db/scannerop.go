// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"

	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/catalog"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/iface/txnif"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/tables/jobs"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/tasks"
)

type ScannerOp interface {
	catalog.Processor
	PreExecute() error
	PostExecute() error
}

const (
	constMergeWaitDuration = 1 * time.Minute
	constMergeMinBlks      = 5
	constHeapCapacity      = 300
	const4GBytes           = 4 * (1 << 30)
)

// min heap item
type mItem[T any] struct {
	row   int
	entry T
}

type itemSet[T any] []*mItem[T]

func (is itemSet[T]) Len() int { return len(is) }

func (is itemSet[T]) Less(i, j int) bool {
	// max heap
	return is[i].row > is[j].row
}

func (is itemSet[T]) Swap(i, j int) {
	is[i], is[j] = is[j], is[i]
}

func (is *itemSet[T]) Push(x any) {
	item := x.(*mItem[T])
	*is = append(*is, item)
}

func (is *itemSet[T]) Pop() any {
	old := *is
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*is = old[0 : n-1]
	return item
}

func (is *itemSet[T]) Clear() {
	old := *is
	*is = old[:0]
}

// heapBuilder founds out blocks to be merged via maintaining a min heap holding
// up to default 300 items.
type heapBuilder[T any] struct {
	items itemSet[T]
	cap   int
}

func (h *heapBuilder[T]) reset() {
	h.items.Clear()
}

func (h *heapBuilder[T]) push(item *mItem[T]) {
	heap.Push(&h.items, item)
	if h.items.Len() > h.cap {
		heap.Pop(&h.items)
	}
}

// copy out the items in the heap
func (h *heapBuilder[T]) finish() []T {
	ret := make([]T, h.items.Len())
	for i, item := range h.items {
		ret[i] = item.entry
	}
	return ret
}

// deletableSegBuilder founds deletable segemnts of a table.
// if a segment has no any non-dropped blocks, it can be deleted. except the
// segment has the max segment id, appender may creates block in it.
type deletableSegBuilder struct {
	segHasNonDropBlk     bool
	segRowCnt, segRowDel int
	segIsSorted          bool
	maxSegId             uint64
	segCandids           []*catalog.SegmentEntry // appendable
	nsegCandids          []*catalog.SegmentEntry // non-appendable
}

func (d *deletableSegBuilder) reset() {
	d.segHasNonDropBlk = false
	d.segIsSorted = false
	d.maxSegId = 0
	d.segCandids = d.segCandids[:0]
	d.nsegCandids = d.nsegCandids[:0]
}

func (d *deletableSegBuilder) resetForNewSeg() {
	d.segHasNonDropBlk = false
	d.segIsSorted = false
	d.segRowCnt = 0
	d.segRowDel = 0
}

// call this when a non dropped block was found when iterating blocks of a segment,
// which make the builder skip this segment
func (d *deletableSegBuilder) hintNonDropBlock() {
	d.segHasNonDropBlk = true
}

func (d *deletableSegBuilder) push(entry *catalog.SegmentEntry) {
	isAppendable := entry.IsAppendable()
	if isAppendable && d.maxSegId < entry.SortHint {
		d.maxSegId = entry.SortHint
	}
	if d.segHasNonDropBlk {
		return
	}
	// all blocks has been dropped
	if isAppendable {
		d.segCandids = append(d.segCandids, entry)
	} else {
		d.nsegCandids = append(d.nsegCandids, entry)
	}
}

// copy out segment entries expect the one with max segment id.
func (d *deletableSegBuilder) finish() []*catalog.SegmentEntry {
	sort.Slice(d.segCandids, func(i, j int) bool { return d.segCandids[i].SortHint < d.segCandids[j].SortHint })
	if last := len(d.segCandids) - 1; last >= 0 && d.segCandids[last].SortHint == d.maxSegId {
		d.segCandids = d.segCandids[:last]
	}
	if len(d.segCandids) == 0 && len(d.nsegCandids) == 0 {
		return nil
	}
	ret := make([]*catalog.SegmentEntry, len(d.segCandids)+len(d.nsegCandids))
	copy(ret[:len(d.segCandids)], d.segCandids)
	copy(ret[len(d.segCandids):], d.nsegCandids)
	if cnt := len(d.nsegCandids); cnt != 0 {
		logutil.Info("Mergeblocks deletable nseg", zap.Int("cnt", cnt))
	}
	return ret
}

type stat struct {
	ttl          time.Time
	lastTotalRow int
}

func (st *stat) String() string {
	return fmt.Sprintf("row%d[%s]", st.lastTotalRow, st.ttl)
}

// mergeLimiter consider update rate and time to decide to merge or not.
type mergeLimiter struct {
	stats               map[uint64]*stat
	mergeMaxRows        int
	blkMaxRows          int
	tableName           string
	memAvail            int
	activeMergeBlkCount int32
}

func (ml *mergeLimiter) IncActiveCount(n int) {
	atomic.AddInt32(&ml.activeMergeBlkCount, int32(n))
}

func (ml *mergeLimiter) OnExecDone(v any) {
	task := v.(tasks.MScopedTask)
	n := int32(len(task.Scopes()))
	atomic.AddInt32(&ml.activeMergeBlkCount, -n)
}

func (ml *mergeLimiter) checkMemAvail(blks int) bool {
	// by experience, it is assumed the merging 256 * 8192 rows costs 4 GB
	if ml.blkMaxRows == 0 {
		return false
	}
	quotaBlks := ml.memAvail / const4GBytes * 256 * 8192 / ml.blkMaxRows
	merging := atomic.LoadInt32(&ml.activeMergeBlkCount)
	return quotaBlks-int(merging) > blks
}

// merge immediately if it has enough rows, skip if:
// 1. has only a few rows or blocks
// 2. is actively updating, which means total rows changes obviously compared with last time
// in other cases, wait some time to merge
func (ml *mergeLimiter) canMerge(tid uint64, totalRow int, deletes int, blks int) bool {
	if !ml.checkMemAvail(blks) {
		return false
	}
	if totalRow > ml.mergeMaxRows {
		logutil.Warnf(
			"Mergeblocks %d-%s merge right now: %d rows %d blks",
			tid, ml.tableName, totalRow, blks)
		delete(ml.stats, tid)
		return true
	}

	if blks < constMergeMinBlks {
		return false
	}

	if st, ok := ml.stats[tid]; !ok {
		ml.stats[tid] = &stat{
			ttl:          ml.ttl(totalRow),
			lastTotalRow: totalRow,
		}
		return false
	} else if deletes == 0 /*no update, simple insert*/ && totalRow-st.lastTotalRow > 0 {
		// quick append is happening, wait some time...
		// it is not bad to gather 10 full blocks anyway.
		st.ttl = ml.ttl(totalRow)
		st.lastTotalRow = totalRow
		logutil.Warnf(
			"Mergeblocks resched table %d-%s, resched to %v",
			tid, ml.tableName, st.ttl)
		return false
	} else {
		// this table is quiet finally, check ttl
		return st.ttl.Before(time.Now())
	}
}

func (ml *mergeLimiter) ttl(totalRow int) time.Time {
	return time.Now().Add(time.Duration(
		(float32(constMergeWaitDuration) / float32(ml.mergeMaxRows)) *
			(float32(ml.mergeMaxRows) - float32(totalRow))))
}

// prune old stat entry
func (ml *mergeLimiter) pruneStale() {
	staleIds := make([]uint64, 0)
	t := time.Now().Add(-10 * time.Minute)
	for id, st := range ml.stats {
		if st.ttl.Before(t) {
			staleIds = append(staleIds, id)
		}
	}
	for _, id := range staleIds {
		delete(ml.stats, id)
	}
}

func (ml *mergeLimiter) String() string {
	return fmt.Sprintf("%v", ml.stats)
}

type MergeTaskBuilder struct {
	db *DB
	*catalog.LoopProcessor
	runCnt           int
	tableRowCnt      int
	tableDelete      int
	tid              uint64
	limiter          *mergeLimiter
	segBuilder       *deletableSegBuilder
	blkBuilder       *heapBuilder[*catalog.BlockEntry]
	sortedSegBuilder *heapBuilder[*catalog.SegmentEntry]
}

func newMergeTaskBuiler(db *DB) *MergeTaskBuilder {
	op := &MergeTaskBuilder{
		db:            db,
		LoopProcessor: new(catalog.LoopProcessor),
		limiter: &mergeLimiter{
			stats: make(map[uint64]*stat),
		},
		segBuilder: &deletableSegBuilder{
			segCandids:  make([]*catalog.SegmentEntry, 0),
			nsegCandids: make([]*catalog.SegmentEntry, 0),
		},
		blkBuilder: &heapBuilder[*catalog.BlockEntry]{
			items: make(itemSet[*catalog.BlockEntry], 0, constHeapCapacity),
			cap:   constHeapCapacity,
		},
		sortedSegBuilder: &heapBuilder[*catalog.SegmentEntry]{
			items: make(itemSet[*catalog.SegmentEntry], 0, 2),
			cap:   2,
		},
	}

	op.TableFn = op.onTable
	op.BlockFn = op.onBlock
	op.SegmentFn = op.onSegment
	op.PostSegmentFn = op.onPostSegment
	op.PostTableFn = op.onPostTable
	return op
}

func (s *MergeTaskBuilder) checkSortedSegs(segs []*catalog.SegmentEntry) (
	mblks []*catalog.BlockEntry, msegs []*catalog.SegmentEntry,
) {
	if len(segs) < 2 {
		return
	}
	s1, s2 := segs[0], segs[1]
	r1 := s1.Stat.Rows - s1.Stat.Dels
	r2 := s2.Stat.Rows - s2.Stat.Dels
	logutil.Infof(
		"mergeblocks ======== %v %v | %v %v | %v %v",
		s1.SortHint, s2.SortHint, r1, r2, s1.Stat.MergeIntent, s2.Stat.MergeIntent,
	)
	// skip big segment
	if r1 > s.limiter.mergeMaxRows*15 || r2 > s.limiter.mergeMaxRows*15 {
		return
	}

	// push back schedule for big gap
	if gap := math.Abs(float64(r1 - r2)); gap > float64(s.limiter.mergeMaxRows) {
		// bump intention
		s1.Stat.MergeIntent++
		s2.Stat.MergeIntent++
		factor := gap / float64(s.limiter.mergeMaxRows)
		if float64(s1.Stat.MergeIntent) < 20*factor ||
			float64(s2.Stat.MergeIntent) < 20*factor {
			return
		}
	}

	msegs = segs[:2]
	mblks = make([]*catalog.BlockEntry, 0, len(msegs)*constMergeMinBlks)
	for _, seg := range msegs {
		blkit := seg.MakeBlockIt(true)
		for ; blkit.Valid(); blkit.Next() {
			entry := blkit.Get().GetPayload()
			if !entry.IsActive() {
				continue
			}
			entry.RLock()
			if entry.IsCommitted() &&
				catalog.ActiveWithNoTxnFilter(entry.BaseEntryImpl) {
				mblks = append(mblks, entry)
			}
			entry.RUnlock()
		}
	}

	logutil.Infof("mergeblocks merge %v-%v, sorted %d and %d rows",
		s.tid, s.limiter.tableName, r1, r2)
	return
}

func (s *MergeTaskBuilder) trySchedMergeTask() {
	if s.tid == 0 {
		return
	}
	// compactable blks
	mergedBlks := s.blkBuilder.finish()
	// deletable segs
	mergedSegs := s.segBuilder.finish()
	hasDelSeg := len(mergedSegs) > 0

	if strings.HasPrefix(s.limiter.tableName, "sbtest") {
		logutil.Warnf(
			"mergeblocks onPostTable %v-%v, totalRow %v, deleteRow %v, blks %v, sortedCandidates %v",
			s.tid, s.limiter.tableName, s.tableRowCnt, s.tableDelete, len(mergedBlks), len(s.sortedSegBuilder.items))
	}

	hasMergeBlk := s.limiter.canMerge(s.tid, s.tableRowCnt, s.tableDelete, len(mergedBlks))
	if !hasMergeBlk {
		mblks, msegs := s.checkSortedSegs(s.sortedSegBuilder.finish())
		if len(mblks) > 0 && s.limiter.checkMemAvail(len(mblks)) {
			mergedBlks = mblks
			mergedSegs = append(mergedSegs, msegs...)
			hasMergeBlk = true
		}
	}

	if !hasDelSeg && !hasMergeBlk {
		return
	}

	segScopes := make([]common.ID, len(mergedSegs))
	for i, s := range mergedSegs {
		segScopes[i] = *s.AsCommonID()
	}

	// remove stale segments only
	if hasDelSeg && !hasMergeBlk {
		factory := func(ctx *tasks.Context, txn txnif.AsyncTxn) (tasks.Task, error) {
			return jobs.NewDelSegTask(ctx, txn, mergedSegs), nil
		}
		_, err := s.db.Runtime.Scheduler.ScheduleMultiScopedTxnTask(nil, tasks.DataCompactionTask, segScopes, factory)
		if err != nil {
			logutil.Warnf("[Mergeblocks] Schedule del seg errinfo=%v", err)
			return
		}
		logutil.Warnf("[Mergeblocks] Scheduled | del %d seg", len(mergedSegs))
		return
	}

	scopes := make([]common.ID, len(mergedBlks))
	for i, blk := range mergedBlks {
		scopes[i] = *blk.AsCommonID()
	}

	factory := func(ctx *tasks.Context, txn txnif.AsyncTxn) (tasks.Task, error) {
		return jobs.NewMergeBlocksTask(ctx, txn, mergedBlks, mergedSegs, nil, s.db.Runtime)
	}
	task, err := s.db.Runtime.Scheduler.ScheduleMultiScopedTxnTask(nil, tasks.DataCompactionTask, scopes, factory)
	if err != nil {
		if err != tasks.ErrScheduleScopeConflict {
			logutil.Warnf("[Mergeblocks] Schedule error info=%v", err)
		}
	} else {
		// reset flag of sorted
		for _, seg := range mergedSegs {
			seg.Stat.SameDelsStreak = 0
			seg.Stat.MergeIntent = 0
		}
		n := len(scopes)
		s.limiter.IncActiveCount(n)
		task.AddObserver(s.limiter)
		if n > constMergeMinBlks {
			n = constMergeMinBlks
		}
		logutil.Warnf("[Mergeblocks] Scheduled | Scopes=[%d],[%d]%s",
			len(segScopes), len(scopes),
			common.BlockIDArraryString(scopes[:n]))
	}
}

func (s *MergeTaskBuilder) resetForTable(entry *catalog.TableEntry) {
	s.tableRowCnt = 0
	s.tableDelete = 0
	s.tid = 0
	if entry != nil {
		s.tid = entry.ID
		schema := entry.GetLastestSchema()
		s.limiter.mergeMaxRows = determineMaxRows(
			int(schema.SegmentMaxBlocks), int(schema.BlockMaxRows))
		s.limiter.tableName = schema.Name
		s.limiter.blkMaxRows = int(schema.BlockMaxRows)
	}
	s.segBuilder.reset()
	s.blkBuilder.reset()
	s.sortedSegBuilder.reset()
}

func determineMaxRows(segMaxBlks, blkMaxRows int) int {
	fullrows := segMaxBlks * blkMaxRows
	// for first merged layer, we want at most 5 full blocks in merged segment
	maxRows := constMergeMinBlks * blkMaxRows
	if fullrows < maxRows { // for small config in unit test
		return fullrows
	}
	return maxRows
}

// PreExecute is called before each loop, refresh and print stats
func (s *MergeTaskBuilder) PreExecute() error {
	// clean stale stats for every 10min (default)
	if s.runCnt++; s.runCnt >= 120 {
		s.runCnt = 0
		s.limiter.pruneStale()
	}

	// print stats for every 50s (default)
	if s.runCnt%10 == 0 {
		logutil.Warnf("Mergeblocks stats: %s", s.limiter.String())
	}

	// fresh mem use
	if stats, err := mem.VirtualMemory(); err == nil {
		s.limiter.memAvail = int(stats.Available)
	}
	return nil
}
func (s *MergeTaskBuilder) PostExecute() error {
	if cnt := atomic.LoadInt32(&s.limiter.activeMergeBlkCount); cnt > 0 {
		logutil.Warnf(
			"Mergeblocks avail mem: %dG, current active blk: %d",
			s.limiter.memAvail/(1<<30), cnt)
	}
	logutil.Infof("mergeblocks ------------------------------------")
	return nil
}

func (s *MergeTaskBuilder) onTable(tableEntry *catalog.TableEntry) (err error) {
	if !tableEntry.IsActive() {
		err = moerr.GetOkStopCurrRecur()
	}
	s.resetForTable(tableEntry)
	return
}

func (s *MergeTaskBuilder) onPostTable(tableEntry *catalog.TableEntry) (err error) {
	// base on the info of tableEntry, we can decide whether to merge or not
	s.trySchedMergeTask()
	return
}

func (s *MergeTaskBuilder) onSegment(segmentEntry *catalog.SegmentEntry) (err error) {
	if !segmentEntry.IsActive() {
		return moerr.GetOkStopCurrRecur()
	}
	// TODO Iter non appendable segs to delete all. Typical occasion is TPCC
	s.segBuilder.resetForNewSeg()
	s.segBuilder.segIsSorted = segmentEntry.IsSorted()
	return
}

func (s *MergeTaskBuilder) onPostSegment(seg *catalog.SegmentEntry) (err error) {
	s.segBuilder.push(seg)

	if !seg.IsSorted() {
		return nil
	}

	seg.Stat.Rows = s.segBuilder.segRowCnt
	if seg.Stat.Dels == s.segBuilder.segRowDel {
		seg.Stat.SameDelsStreak++
		if seg.Stat.SameDelsStreak > 10 {
			s.sortedSegBuilder.push(&mItem[*catalog.SegmentEntry]{
				row:   s.segBuilder.segRowCnt - s.segBuilder.segRowDel,
				entry: seg,
			})
		}
	} else {
		seg.Stat.SameDelsStreak = 0
	}
	seg.Stat.Dels = s.segBuilder.segRowDel

	return nil
}

func (s *MergeTaskBuilder) onBlock(entry *catalog.BlockEntry) (err error) {
	if !entry.IsActive() {
		return
	}
	s.segBuilder.hintNonDropBlock()

	entry.RLock()
	defer entry.RUnlock()

	// Skip uncommitted entries and appendable block
	if !entry.IsCommitted() ||
		!catalog.ActiveWithNoTxnFilter(entry.BaseEntryImpl) ||
		!catalog.NonAppendableBlkFilter(entry) {
		return
	}

	// nblks in appenable segs or non-sorted non-appendable segs
	// these blks are formed by continuous append
	entry.RUnlock()
	rows := entry.GetBlockData().Rows()
	dels := entry.GetBlockData().GetTotalDeletes()
	entry.RLock()
	s.segBuilder.segRowCnt += rows
	s.segBuilder.segRowDel += dels
	if s.segBuilder.segIsSorted {
		return
	}
	s.tableRowCnt += rows
	s.tableDelete += dels
	s.blkBuilder.push(&mItem[*catalog.BlockEntry]{row: rows - dels, entry: entry})
	return nil
}
