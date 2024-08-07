// Copyright 2022 Matrix Origin
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

package disttae

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/perfcounter"
	plan2 "github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/testutil"
	"github.com/matrixorigin/matrixone/pkg/txn/trace"
	v2 "github.com/matrixorigin/matrixone/pkg/util/metric/v2"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/disttae/logtailreplay"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/blockio"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/index"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
	"go.uber.org/zap"
)

// -----------------------------------------------------------------
// ------------------------ withFilterMixin ------------------------
// -----------------------------------------------------------------

func (mixin *withFilterMixin) reset() {
	mixin.filterState.evaluated = false
	mixin.filterState.filter = blockio.BlockReadFilter{}
	mixin.columns.pkPos = -1
	mixin.columns.indexOfFirstSortedColumn = -1
	mixin.columns.seqnums = nil
	mixin.columns.colTypes = nil
	mixin.sels = nil
}

// when the reader.Read is called for a new block, it will always
// call tryUpdate to update the seqnums
// NOTE: here we assume the tryUpdate is always called with the same cols
// for all blocks and it will only be updated once
func (mixin *withFilterMixin) tryUpdateColumns(cols []string, blkCnt int) {
	if len(cols) == len(mixin.columns.seqnums) {
		return
	}
	if len(mixin.columns.seqnums) != 0 {
		panic(moerr.NewInternalErrorNoCtx("withFilterMixin tryUpdate called with different cols"))
	}

	// record the column selectivity
	chit, ctotal := len(cols), len(mixin.tableDef.Cols)
	v2.TaskSelColumnTotal.Add(float64(ctotal))
	v2.TaskSelColumnHit.Add(float64(ctotal - chit))
	blockio.RecordColumnSelectivity(mixin.proc.GetService(), chit, ctotal)

	mixin.columns.seqnums = make([]uint16, len(cols))
	mixin.columns.colTypes = make([]types.Type, len(cols))
	// mixin.columns.colNulls = make([]bool, len(cols))
	mixin.columns.pkPos = -1
	mixin.columns.indexOfFirstSortedColumn = -1
	for i, column := range cols {
		column = strings.ToLower(column)
		if column == catalog.Row_ID {
			mixin.columns.seqnums[i] = objectio.SEQNUM_ROWID
			mixin.columns.colTypes[i] = objectio.RowidType
		} else {
			if plan2.GetSortOrderByName(mixin.tableDef, column) == 0 {
				mixin.columns.indexOfFirstSortedColumn = i
			}
			colIdx := mixin.tableDef.Name2ColIndex[column]
			colDef := mixin.tableDef.Cols[colIdx]
			mixin.columns.seqnums[i] = uint16(colDef.Seqnum)

			if mixin.tableDef.Pkey != nil && mixin.tableDef.Pkey.PkeyColName == column {
				// primary key is in the cols
				mixin.columns.pkPos = i
			}
			mixin.columns.colTypes[i] = types.T(colDef.Typ.Id).ToType()
		}
	}

	if mixin.columns.pkPos != -1 {
		// here we will select the primary key column from the vectors, and
		// use the search function to find the offset of the primary key.
		// it returns the offset of the primary key in the pk vector.
		// if the primary key is not found, it returns empty slice
		mixin.filterState.evaluated = true
		//mixin.filterState.filter = filter
		mixin.filterState.seqnums = []uint16{mixin.columns.seqnums[mixin.columns.pkPos]}
		mixin.filterState.colTypes = mixin.columns.colTypes[mixin.columns.pkPos : mixin.columns.pkPos+1]

		// records how many blks one reader needs to read when having filter
		objectio.BlkReadStats.BlksByReaderStats.Record(1, blkCnt)
	}
}

// -----------------------------------------------------------------
// ------------------------ emptyReader ----------------------------
// -----------------------------------------------------------------

func (r *emptyReader) SetFilterZM(objectio.ZoneMap) {
}

func (r *emptyReader) GetOrderBy() []*plan.OrderBySpec {
	return nil
}

func (r *emptyReader) SetOrderBy([]*plan.OrderBySpec) {
}

func (r *emptyReader) Close() error {
	return nil
}

func (r *emptyReader) Read(_ context.Context, _ []string,
	_ *plan.Expr, _ *mpool.MPool, _ engine.VectorPool) (*batch.Batch, error) {
	return nil, nil
}

// -----------------------------------------------------------------
// ------------------------ blockReader ----------------------------
// -----------------------------------------------------------------

func newBlockReader(
	ctx context.Context,
	tableDef *plan.TableDef,
	ts timestamp.Timestamp,
	blks []*objectio.BlockInfo,
	filterExpr *plan.Expr,
	filter blockio.BlockReadFilter,
	fs fileservice.FileService,
	proc *process.Process,
) *blockReader {
	for _, blk := range blks {
		trace.GetService(proc.GetService()).TxnReadBlock(
			proc.GetTxnOperator(),
			tableDef.TblId,
			blk.BlockID[:])
	}
	r := &blockReader{
		withFilterMixin: withFilterMixin{
			ctx:      ctx,
			fs:       fs,
			ts:       ts,
			proc:     proc,
			tableDef: tableDef,
		},
		blks: blks,
	}
	r.filterState.expr = filterExpr
	r.filterState.filter = filter
	return r
}

func (r *blockReader) Close() error {
	r.withFilterMixin.reset()
	r.blks = nil
	r.buffer = nil
	return nil
}

func (r *blockReader) SetFilterZM(zm objectio.ZoneMap) {
	if !r.filterZM.IsInited() {
		r.filterZM = zm.Clone()
		return
	}
	if r.desc && r.filterZM.CompareMax(zm) < 0 {
		r.filterZM = zm.Clone()
		return
	}
	if !r.desc && r.filterZM.CompareMin(zm) > 0 {
		r.filterZM = zm.Clone()
		return
	}
}

func (r *blockReader) GetOrderBy() []*plan.OrderBySpec {
	return r.OrderBy
}

func (r *blockReader) SetOrderBy(orderby []*plan.OrderBySpec) {
	r.OrderBy = orderby
}

func (r *blockReader) needReadBlkByZM(i int) bool {
	zm := r.blockZMS[i]
	if !r.filterZM.IsInited() || !zm.IsInited() {
		return true
	}
	if r.desc {
		return r.filterZM.CompareMax(zm) <= 0
	} else {
		return r.filterZM.CompareMin(zm) >= 0
	}
}

func (r *blockReader) getBlockZMs() {
	orderByCol, _ := r.OrderBy[0].Expr.Expr.(*plan.Expr_Col)
	orderByColIDX := int(r.tableDef.Cols[int(orderByCol.Col.ColPos)].Seqnum)

	r.blockZMS = make([]index.ZM, len(r.blks))
	var objDataMeta objectio.ObjectDataMeta
	var location objectio.Location
	for i := range r.blks {
		location = r.blks[i].MetaLocation()
		if !objectio.IsSameObjectLocVsMeta(location, objDataMeta) {
			objMeta, err := objectio.FastLoadObjectMeta(r.ctx, &location, false, r.fs)
			if err != nil {
				panic("load object meta error when ordered scan!")
			}
			objDataMeta = objMeta.MustDataMeta()
		}
		blkMeta := objDataMeta.GetBlockMeta(uint32(location.ID()))
		r.blockZMS[i] = blkMeta.ColumnMeta(uint16(orderByColIDX)).ZoneMap()
	}
}

func (r *blockReader) sortBlockList() {
	helper := make([]*blockSortHelper, len(r.blks))
	for i := range r.blks {
		helper[i] = &blockSortHelper{}
		helper[i].blk = r.blks[i]
		helper[i].zm = r.blockZMS[i]
	}
	if r.desc {
		sort.Slice(helper, func(i, j int) bool {
			zm1 := helper[i].zm
			if !zm1.IsInited() {
				return true
			}
			zm2 := helper[j].zm
			if !zm2.IsInited() {
				return false
			}
			return zm1.CompareMax(zm2) > 0
		})
	} else {
		sort.Slice(helper, func(i, j int) bool {
			zm1 := helper[i].zm
			if !zm1.IsInited() {
				return true
			}
			zm2 := helper[j].zm
			if !zm2.IsInited() {
				return false
			}
			return zm1.CompareMin(zm2) < 0
		})
	}

	for i := range helper {
		r.blks[i] = helper[i].blk
		r.blockZMS[i] = helper[i].zm
	}
}

func (r *blockReader) deleteFirstNBlocks(n int) {
	r.blks = r.blks[n:]
	if len(r.OrderBy) > 0 {
		r.blockZMS = r.blockZMS[n:]
	}
}

func (r *blockReader) Read(
	ctx context.Context,
	cols []string,
	_ *plan.Expr,
	mp *mpool.MPool,
	vp engine.VectorPool,
) (bat *batch.Batch, err error) {
	start := time.Now()
	defer func() {
		v2.TxnBlockReaderDurationHistogram.Observe(time.Since(start).Seconds())
	}()

	// for ordered scan, sort blocklist by zonemap info, and then filter by zonemap
	if len(r.OrderBy) > 0 {
		if !r.sorted {
			r.desc = r.OrderBy[0].Flag&plan.OrderBySpec_DESC != 0
			r.getBlockZMs()
			r.sortBlockList()
			r.sorted = true
		}
		i := 0
		for i < len(r.blks) {
			if r.needReadBlkByZM(i) {
				break
			}
			i++
		}
		r.deleteFirstNBlocks(i)
	}
	// if the block list is empty, return nil
	if len(r.blks) == 0 {
		return nil, nil
	}

	// move to the next block at the end of this call
	defer func() {
		r.deleteFirstNBlocks(1)
		r.buffer = r.buffer[:0]
		r.currentStep++
	}()

	// get the current block to be read
	blockInfo := r.blks[0]

	// try to update the columns
	// the columns is only updated once for all blocks
	r.tryUpdateColumns(cols, len(r.blks))

	filter := r.withFilterMixin.filterState.filter

	// if any null expr is found in the primary key (composite primary keys), quick return
	if r.filterState.hasNull {
		return nil, nil
	}

	if !r.dontPrefetch {
		//prefetch some objects
		for len(r.steps) > 0 && r.steps[0] == r.currentStep {
			// always true for now, will optimize this in the future
			prefetchFile := r.scanType == SMALL || r.scanType == LARGE || r.scanType == NORMAL
			if filter.Valid && blockInfo.Sorted {
				err = blockio.BlockPrefetch(r.withFilterMixin.proc.GetService(), r.filterState.seqnums, r.fs, [][]*objectio.BlockInfo{r.infos[0]}, prefetchFile)
			} else {
				err = blockio.BlockPrefetch(r.withFilterMixin.proc.GetService(), r.columns.seqnums, r.fs, [][]*objectio.BlockInfo{r.infos[0]}, prefetchFile)
			}
			if err != nil {
				return nil, err
			}
			r.infos = r.infos[1:]
			r.steps = r.steps[1:]
		}
	}

	statsCtx, numRead, numHit := r.ctx, int64(0), int64(0)
	if filter.Valid {
		// try to store the blkReadStats CounterSet into ctx, so that
		// it can record the mem cache hit stats when call MemCache.Read() later soon.
		statsCtx, numRead, numHit = r.prepareGatherStats()
	}

	// read the block
	var policy fileservice.Policy
	if r.scanType == LARGE || r.scanType == NORMAL {
		policy = fileservice.SkipMemoryCacheWrites
	}

	bat, err = blockio.BlockRead(
		statsCtx, r.withFilterMixin.proc.GetService(), blockInfo, r.buffer, r.columns.seqnums, r.columns.colTypes, r.ts,
		r.filterState.seqnums,
		r.filterState.colTypes,
		filter,
		r.fs, mp, vp, policy,
	)
	if err != nil {
		return nil, err
	}

	if filter.Valid {
		// we collect mem cache hit related statistics info for blk read here
		r.gatherStats(numRead, numHit)
	}

	bat.SetAttributes(cols)

	if blockInfo.Sorted && r.columns.indexOfFirstSortedColumn != -1 {
		bat.GetVector(int32(r.columns.indexOfFirstSortedColumn)).SetSorted(true)
	}

	if logutil.GetSkip1Logger().Core().Enabled(zap.DebugLevel) {
		logutil.Debug(testutil.OperatorCatchBatch("block reader", bat))
	}
	return bat, nil
}

func (r *blockReader) prepareGatherStats() (context.Context, int64, int64) {
	ctx := perfcounter.WithCounterSet(r.ctx, objectio.BlkReadStats.CounterSet)
	return ctx, objectio.BlkReadStats.CounterSet.FileService.Cache.Read.Load(),
		objectio.BlkReadStats.CounterSet.FileService.Cache.Hit.Load()
}

func (r *blockReader) gatherStats(lastNumRead, lastNumHit int64) {
	numRead := objectio.BlkReadStats.CounterSet.FileService.Cache.Read.Load()
	numHit := objectio.BlkReadStats.CounterSet.FileService.Cache.Hit.Load()

	curNumRead := numRead - lastNumRead
	curNumHit := numHit - lastNumHit

	if curNumRead > curNumHit {
		objectio.BlkReadStats.BlkCacheHitStats.Record(0, 1)
	} else {
		objectio.BlkReadStats.BlkCacheHitStats.Record(1, 1)
	}

	objectio.BlkReadStats.EntryCacheHitStats.Record(int(curNumHit), int(curNumRead))
}

// -----------------------------------------------------------------
// ---------------------- blockMergeReader -------------------------
// -----------------------------------------------------------------

func newBlockMergeReader(
	ctx context.Context,
	txnTable *txnTable,
	memFilter memPKFilter,
	blockFilter blockio.BlockReadFilter,
	ts timestamp.Timestamp,
	dirtyBlks []*objectio.BlockInfo,
	filterExpr *plan.Expr,
	txnOffset int, // Transaction writes offset used to specify the starting position for reading data.
	fs fileservice.FileService,
	proc *process.Process,
) *blockMergeReader {
	r := &blockMergeReader{
		table:     txnTable,
		txnOffset: txnOffset,
		blockReader: newBlockReader(
			ctx,
			txnTable.GetTableDef(ctx),
			ts,
			dirtyBlks,
			filterExpr,
			blockFilter,
			fs,
			proc,
		),
		pkFilter:   memFilter,
		deletaLocs: make(map[string][]objectio.Location),
	}
	return r
}

func (r *blockMergeReader) Close() error {
	r.table = nil
	return r.blockReader.Close()
}

func (r *blockMergeReader) prefetchDeletes() error {
	//load delta locations for r.blocks.
	r.table.getTxn().blockId_tn_delete_metaLoc_batch.RLock()
	defer r.table.getTxn().blockId_tn_delete_metaLoc_batch.RUnlock()

	if !r.loaded {
		for _, info := range r.blks {
			bats, ok := r.table.getTxn().blockId_tn_delete_metaLoc_batch.data[info.BlockID]

			if !ok {
				return nil
			}
			for _, bat := range bats {
				vs, area := vector.MustVarlenaRawData(bat.GetVector(0))
				for i := range vs {
					location, err := blockio.EncodeLocationFromString(vs[i].UnsafeGetString(area))
					if err != nil {
						return err
					}
					r.deletaLocs[location.Name().String()] =
						append(r.deletaLocs[location.Name().String()], location)
				}
			}
		}

		// Get Single Col pk index
		for idx, colDef := range r.tableDef.Cols {
			if colDef.Name == r.tableDef.Pkey.PkeyColName {
				r.pkidx = idx
				break
			}
		}
		r.loaded = true
	}

	//prefetch the deletes
	for name, locs := range r.deletaLocs {
		pref, err := blockio.BuildPrefetchParams(r.fs, locs[0])
		if err != nil {
			return err
		}
		for _, loc := range locs {
			//rowid + pk
			pref.AddBlockWithType([]uint16{0, uint16(r.pkidx)}, []uint16{loc.ID()}, uint16(objectio.SchemaTombstone))

		}
		delete(r.deletaLocs, name)
		return blockio.PrefetchWithMerged(r.withFilterMixin.proc.GetService(), pref)
	}
	return nil
}

func (r *blockMergeReader) loadDeletes(ctx context.Context, cols []string) error {
	if len(r.blks) == 0 {
		return nil
	}
	info := r.blks[0]

	r.tryUpdateColumns(cols, len(r.blks))
	// load deletes from txn.blockId_dn_delete_metaLoc_batch
	err := r.table.LoadDeletesForBlock(info.BlockID, &r.buffer)
	if err != nil {
		return err
	}

	state, err := r.table.getPartitionState(ctx)
	if err != nil {
		return err
	}
	ts := types.TimestampToTS(r.ts)

	var iter logtailreplay.RowsIter
	if r.pkFilter.delIterFactory != nil {
		iter = r.pkFilter.delIterFactory(info.BlockID)
	}

	if iter != nil {
		for iter.Next() {
			entry := iter.Entry()
			if !entry.Deleted {
				continue
			}
			_, offset := entry.RowID.Decode()
			r.buffer = append(r.buffer, int64(offset))
		}
	} else {
		iter = state.NewRowsIter(ts, &info.BlockID, true)
		currlen := len(r.buffer)
		for iter.Next() {
			entry := iter.Entry()
			_, offset := entry.RowID.Decode()
			r.buffer = append(r.buffer, int64(offset))
		}
		v2.TaskLoadMemDeletesPerBlockHistogram.Observe(float64(len(r.buffer) - currlen))
	}

	iter.Close()

	txnOffset := r.txnOffset
	if r.table.db.op.IsSnapOp() {
		txnOffset = r.table.getTxn().GetSnapshotWriteOffset()
	}

	//TODO:: if r.table.writes is a map , the time complexity could be O(1)
	//load deletes from txn.writes for the specified block
	r.table.getTxn().forEachTableWrites(
		r.table.db.databaseId,
		r.table.tableId,
		txnOffset, func(entry Entry) {
			if entry.isGeneratedByTruncate() {
				return
			}
			if (entry.typ == DELETE || entry.typ == DELETE_TXN) && entry.fileName == "" {
				vs := vector.MustFixedCol[types.Rowid](entry.bat.GetVector(0))
				for _, v := range vs {
					id, offset := v.Decode()
					if id == info.BlockID {
						r.buffer = append(r.buffer, int64(offset))
					}
				}
			}
		})
	//load deletes from txn.deletedBlocks.
	txn := r.table.getTxn()
	txn.deletedBlocks.getDeletedOffsetsByBlock(&info.BlockID, &r.buffer)
	return nil
}

func (r *blockMergeReader) Read(
	ctx context.Context,
	cols []string,
	expr *plan.Expr,
	mp *mpool.MPool,
	vp engine.VectorPool,
) (*batch.Batch, error) {
	start := time.Now()
	defer func() {
		v2.TxnBlockMergeReaderDurationHistogram.Observe(time.Since(start).Seconds())
	}()

	//prefetch deletes for r.blks
	if err := r.prefetchDeletes(); err != nil {
		return nil, err
	}
	//load deletes for the specified block
	if err := r.loadDeletes(ctx, cols); err != nil {
		return nil, err
	}
	return r.blockReader.Read(ctx, cols, expr, mp, vp)
}

// -----------------------------------------------------------------
// ------------------------ mergeReader ----------------------------
// -----------------------------------------------------------------

func NewMergeReader(readers []engine.Reader) *mergeReader {
	return &mergeReader{
		rds: readers,
	}
}

func (r *mergeReader) SetFilterZM(zm objectio.ZoneMap) {
	for i := range r.rds {
		r.rds[i].SetFilterZM(zm)
	}
}

func (r *mergeReader) GetOrderBy() []*plan.OrderBySpec {
	for i := range r.rds {
		if r.rds[i].GetOrderBy() != nil {
			return r.rds[i].GetOrderBy()
		}
	}
	return nil
}

func (r *mergeReader) SetOrderBy(orderby []*plan.OrderBySpec) {
	for i := range r.rds {
		r.rds[i].SetOrderBy(orderby)
	}
}

func (r *mergeReader) Close() error {
	return nil
}

func (r *mergeReader) Read(
	ctx context.Context,
	cols []string,
	expr *plan.Expr,
	mp *mpool.MPool,
	vp engine.VectorPool,
) (*batch.Batch, error) {
	start := time.Now()
	defer func() {
		v2.TxnMergeReaderDurationHistogram.Observe(time.Since(start).Seconds())
	}()

	if len(r.rds) == 0 {
		return nil, nil
	}
	for len(r.rds) > 0 {
		bat, err := r.rds[0].Read(ctx, cols, expr, mp, vp)
		if err != nil {
			for _, rd := range r.rds {
				rd.Close()
			}
			return nil, err
		}
		if bat == nil {
			r.rds = r.rds[1:]
		}
		if bat != nil {
			if logutil.GetSkip1Logger().Core().Enabled(zap.DebugLevel) {
				logutil.Debug(testutil.OperatorCatchBatch("merge reader", bat))
			}
			return bat, nil
		}
	}
	return nil, nil
}
