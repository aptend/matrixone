// Copyright 2021 Matrix Origin
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

package mergesort

import (
	"context"
	"fmt"
	"time"
	"unsafe"

	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/pb/api"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/blockio"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"go.uber.org/zap"
)

// DisposableVecPool bridge the gap between the vector pools in cn and tn
type DisposableVecPool interface {
	GetVector(*types.Type) (ret *vector.Vector, release func())
	GetMPool() *mpool.MPool
}

type MergeTaskHost interface {
	DisposableVecPool
	PrepareData() ([]*batch.Batch, []*nulls.Nulls, func(), error)
	PrepareCommitEntry() *MergeCommitEntry
	PrepareNewWriterFunc() func() *blockio.BlockWriter
}

type MergeCommitEntry struct {
	DbID               uint64
	TableID            uint64
	Tablename          string
	StartTs            types.TS
	MergedObjs         []objectio.ObjectStats // deleted
	CreatedObjectStats []objectio.ObjectStats // created
	Booking            *api.BlkTransferBooking
}

func (e *MergeCommitEntry) InitTransfermapping(blkcnt int) {
	e.Booking = api.NewBlkTransferBooking(blkcnt)
}

type MergeTicket struct {
	DbID    uint64
	TableID uint64
	Targets []objectio.ObjectStats
}

func GetMustNewWriter(
	fs fileservice.FileService,
	ver uint32, seqnums []uint16,
	sortkeyPos int, sortkeyIsPK bool,
) func() *blockio.BlockWriter {
	return func() *blockio.BlockWriter {
		name := objectio.BuildObjectNameWithObjectID(objectio.NewObjectid())
		writer, err := blockio.NewBlockWriterNew(fs, name, ver, seqnums)
		if err != nil {
			panic(err) // it is impossible
		}
		// has sortkey
		if sortkeyPos >= 0 {
			if sortkeyIsPK {
				writer.SetPrimaryKey(uint16(sortkeyPos))
			} else { // cluster by
				writer.SetSortKey(uint16(sortkeyPos))
			}
		}
		return writer
	}
}

func DoMergeAndWrite(
	ctx context.Context,
	sortkeyPos int,
	blkMaxRow int,
	mergehost MergeTaskHost,
) (err error) {
	now := time.Now()
	/*out args, keep the transfer infomation*/
	commitEntry := mergehost.PrepareCommitEntry()
	objs := ""
	for _, o := range commitEntry.MergedObjs {
		objs = fmt.Sprintf("%s%s,", objs, common.ShortObjId(*o.ObjectName().ObjectId()))
	}
	logutil.Info("[Start] Mergeblocks",
		zap.String("table", commitEntry.Tablename),
		zap.String("txn-start-ts", commitEntry.StartTs.ToString()),
		zap.String("from-objs", objs),
	)
	phaseDesc := "prepare data"
	defer func() {
		if err != nil {
			logutil.Error("[DoneWithErr] Mergeblocks",
				zap.String("table", commitEntry.Tablename),
				zap.Error(err),
				zap.String("phase", phaseDesc),
			)
		}
	}()

	// batches is read from disk, dels is read from disk and memory
	//
	// batches[i] means the i-th non-appendable block to be merged and
	// it has no rowid
	batches, dels, release, err := mergehost.PrepareData()
	if err != nil {
		return err
	}
	defer release()

	commitEntry.InitTransfermapping(len(batches))

	fromLayout := make([]uint32, 0, len(batches))
	totalRowCount := 0

	hasSortKey := sortkeyPos >= 0
	if !hasSortKey {
		sortkeyPos = 0 // no sort key, use the first column to do reshape
	}

	toSortVecs := make([]*vector.Vector, 0, len(batches))

	mpool := mergehost.GetMPool()
	// iter all block to get basic info, do shrink if needed
	for i := range batches {
		rowCntBeforeApplyDelete := batches[i].RowCount()
		del := dels[i]
		if del != nil {
			if del.Count() == rowCntBeforeApplyDelete {
				continue // this block is all deleted, do not merge
			}

			// dup vector before apply delete. old b will be freed in releaseF
			newb, err := batches[i].Dup(mpool)
			if err != nil {
				return err
			}
			defer newb.Clean(mpool) // whoever create new vector, should clean it
			batches[i] = newb
			batches[i].AntiShrink(del.ToI64Arrary())
		}
		commitEntry.Booking.AddSortPhaseMapping(i, rowCntBeforeApplyDelete, del, nil)
		fromLayout = append(fromLayout, uint32(batches[i].RowCount()))
		totalRowCount += batches[i].RowCount()
		toSortVecs = append(toSortVecs, batches[i].GetVector(int32(sortkeyPos)))
	}

	if totalRowCount == 0 {
		logutil.Info("[Done] Mergeblocks due to all deleted",
			zap.String("table", commitEntry.Tablename),
			zap.String("txn-start-ts", commitEntry.StartTs.ToString()))
		commitEntry.Booking.Clean()
		return
	}

	// -------------------------- phase 1
	phaseDesc = "merge sort, or reshape, one column"
	toLayout := arrangeToLayout(totalRowCount, blkMaxRow)

	sortedVecs, releaseF := getRetVecs(len(toLayout), toSortVecs[0].GetType(), mergehost)
	defer releaseF()

	var sortedIdx []uint32
	if hasSortKey {
		// mergesort is needed, allocate sortedidx and mapping
		allocSz := totalRowCount * 4
		// sortedIdx is used to shuffle other columns according to the order of the sort key
		sortIdxNode, err := mpool.Alloc(allocSz)
		if err != nil {
			panic(err)
		}
		// sortedidx will be used to shuffle other column, defer free
		defer mpool.Free(sortIdxNode)
		sortedIdx = unsafe.Slice((*uint32)(unsafe.Pointer(&sortIdxNode[0])), totalRowCount)

		mappingNode, err := mpool.Alloc(allocSz)
		if err != nil {
			panic(err)
		}
		mapping := unsafe.Slice((*uint32)(unsafe.Pointer(&mappingNode[0])), totalRowCount)

		// modify sortidx and mapping
		Merge(toSortVecs, sortedVecs, sortedIdx, mapping, fromLayout, toLayout, mpool)
		commitEntry.Booking.UpdateMappingAfterMerge(mapping, fromLayout, toLayout)
		// free mapping, which is never used again
		mpool.Free(mappingNode)
	} else {
		// just do reshape, keep sortedIdx nil
		Reshape(toSortVecs, sortedVecs, fromLayout, toLayout, mpool)
		commitEntry.Booking.UpdateMappingAfterMerge(nil, fromLayout, toLayout)
	}

	// -------------------------- phase 2
	phaseDesc = "merge sort, or reshape, the rest of columns"

	// prepare multiple batch
	attrs := batches[0].Attrs
	var writtenBatches []*batch.Batch
	for _, vec := range sortedVecs {
		b := batch.New(true, attrs)
		b.SetRowCount(vec.Length())
		writtenBatches = append(writtenBatches, b)
	}

	// arrange the other columns according to sortedidx, or, just reshape
	tempVecs := make([]*vector.Vector, 0, len(batches))
	for i := range attrs {
		// just put the sorted column to the write batch
		if i == sortkeyPos {
			for j, vec := range sortedVecs {
				writtenBatches[j].Vecs[i] = vec
			}
			continue
		}
		tempVecs = tempVecs[:0]

		for k, bat := range batches {
			vec := bat.Vecs[i]
			if vec.Length() == 0 || (dels[k] != nil && vec.Length() == dels[k].GetCardinality()) {
				continue
			}
			tempVecs = append(tempVecs, vec)
		}
		if len(toSortVecs) != len(tempVecs) {
			panic("mismatch length")
		}

		outvecs, release := getRetVecs(len(toLayout), tempVecs[0].GetType(), mergehost)
		defer release()

		if hasSortKey {
			Multiplex(tempVecs, outvecs, sortedIdx, fromLayout, toLayout, mpool)
		} else {
			Reshape(tempVecs, outvecs, fromLayout, toLayout, mpool)
		}

		for j, vec := range outvecs {
			writtenBatches[j].Vecs[i] = vec
		}
	}

	// -------------------------- phase 3
	phaseDesc = "new writer to write down"
	newWriterFunc := mergehost.PrepareNewWriterFunc()
	writer := newWriterFunc()
	for _, bat := range writtenBatches {
		_, err = writer.WriteBatch(bat)
		if err != nil {
			return err
		}
	}

	if _, _, err = writer.Sync(ctx); err != nil {
		return err
	}

	// no tomestone actually
	cobjstats := writer.GetObjectStats()[:objectio.SchemaTombstone]
	commitEntry.CreatedObjectStats = cobjstats
	cobj := fmt.Sprintf("%s(%v)Rows(%v)",
		common.ShortObjId(*cobjstats[0].ObjectName().ObjectId()),
		cobjstats[0].BlkCnt(),
		cobjstats[0].Rows())
	logutil.Info("[Done] Mergeblocks",
		zap.String("table", commitEntry.Tablename),
		zap.String("txn-start-ts", commitEntry.StartTs.ToString()),
		zap.String("to-objs", cobj),
		common.DurationField(time.Since(now)))

	return nil

}

// get vector from pool, and return a release function
func getRetVecs(count int, t *types.Type, vpool DisposableVecPool) (ret []*vector.Vector, releaseAll func()) {
	var fs []func()
	for i := 0; i < count; i++ {
		vec, release := vpool.GetVector(t)
		ret = append(ret, vec)
		fs = append(fs, release)
	}
	releaseAll = func() {
		for i := 0; i < count; i++ {
			fs[i]()
		}
	}
	return
}

// layout [blkMaxRow, blkMaxRow, blkMaxRow,..., blkMaxRow, totalRowCount - blkMaxRow*N]
func arrangeToLayout(totalRowCount int, blkMaxRow int) []uint32 {
	toLayout := make([]uint32, 0, totalRowCount/blkMaxRow)
	unconsumed := totalRowCount
	for unconsumed > 0 {
		if unconsumed > blkMaxRow {
			toLayout = append(toLayout, uint32(blkMaxRow))
			unconsumed -= blkMaxRow
		} else {
			toLayout = append(toLayout, uint32(unconsumed))
			unconsumed = 0
		}
	}
	return toLayout
}
