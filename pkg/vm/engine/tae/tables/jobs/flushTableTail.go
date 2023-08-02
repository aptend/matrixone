package jobs

import (
	"context"
	"fmt"
	"time"
	"unsafe"

	"github.com/matrixorigin/matrixone/pkg/common/bitmap"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/util/fault"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/blockio"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/catalog"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/containers"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/db/dbutils"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/iface/handle"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/iface/txnif"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/mergesort"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/tables/txnentries"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/tasks"
	"go.uber.org/zap"
)

var FlushTableTailTaskFactory = func(
	metas []*catalog.BlockEntry, rt *dbutils.Runtime,
) tasks.TxnTaskFactory {
	return func(ctx *tasks.Context, txn txnif.AsyncTxn) (tasks.Task, error) {
		return NewFlushTableTailTask(ctx, txn, metas, rt)
	}
}

type flushTableTailTask struct {
	*tasks.BaseTask
	txn txnif.AsyncTxn
	rt  *dbutils.Runtime

	scopes []common.ID
	schema *catalog.Schema

	rel  handle.Relation
	dbid uint64

	// record the row mapping from deleted blocks to created blocks
	transMappings *txnentries.BlkTransferBooking

	ablksMetas        []*catalog.BlockEntry
	nblksMetas        []*catalog.BlockEntry
	ablksHandles      []handle.Block
	nblksHandles      []handle.Block
	createdBlkHandles []handle.Block

	createdMergedObjectName  string
	createdDeletesObjectName string

	mergeRowsCnt, ablksDeletesCnt, nblksDeletesCnt int
}

func NewFlushTableTailTask(
	ctx *tasks.Context,
	txn txnif.AsyncTxn,
	blks []*catalog.BlockEntry,
	rt *dbutils.Runtime,
) (task *flushTableTailTask, err error) {
	task = &flushTableTailTask{
		txn: txn,
		rt:  rt,
	}
	meta := blks[0]
	dbId := meta.GetSegment().GetTable().GetDB().ID
	task.dbid = dbId
	database, err := txn.UnsafeGetDatabase(dbId)
	if err != nil {
		return
	}
	tableId := meta.GetSegment().GetTable().ID
	rel, err := database.UnsafeGetRelation(tableId)
	task.rel = rel
	if err != nil {
		return
	}
	task.schema = rel.Schema().(*catalog.Schema)
	seg, err := rel.GetSegment(&meta.GetSegment().ID)
	if err != nil {
		return
	}

	for _, blk := range blks {
		task.scopes = append(task.scopes, *blk.AsCommonID())
		var hdl handle.Block
		if hdl, err = seg.GetBlock(blk.ID); err != nil {
			return
		}
		if blk.IsAppendable() {
			task.ablksMetas = append(task.ablksMetas, blk)
			task.ablksHandles = append(task.ablksHandles, hdl)
		} else {
			task.nblksMetas = append(task.nblksMetas, blk)
			task.nblksHandles = append(task.nblksHandles, hdl)
		}
	}
	task.transMappings = txnentries.NewBlkTransferBooking(len(task.ablksHandles))

	task.BaseTask = tasks.NewBaseTask(task, tasks.DataCompactionTask, ctx)

	return
}

// Scopes is used in conflict checking in scheduler. For ScopedTask interface
func (task *flushTableTailTask) Scopes() []common.ID { return task.scopes }

// Name is for ScopedTask interface
func (task *flushTableTailTask) Name() string {
	return fmt.Sprintf("[%d]FT-%s", task.ID(), task.schema.Name)
}

func (task *flushTableTailTask) Execute(ctx context.Context) (err error) {
	task.rt.Throttle.AcquireCompactionQuota()
	defer task.rt.Throttle.ReleaseCompactionQuota()

	logutil.Info("[Start]", common.OperationField(task.Name()),
		common.OperandField(len(task.ablksHandles)+len(task.nblksHandles)))

	phaseDesc := ""
	defer func() {
		if err != nil {
			logutil.Error("[DoneWithErr]", common.OperationField(task.Name()),
				common.AnyField("error", err),
				common.AnyField("phase", phaseDesc),
			)
		}
	}()
	now := time.Now()

	/////////////////////
	//// phase seperator
	///////////////////

	phaseDesc = "1-flushing appendable blocks for snapshot"
	snapshotSubtasks, err := task.flushAblksForSnapshot(ctx)
	if err != nil {
		return
	}
	defer releaseFlushBlkTasks(snapshotSubtasks)

	/////////////////////
	//// phase seperator
	///////////////////

	phaseDesc = "1-write all deletes from nablks"
	// just collect deletes, do not soft delete it, leave that to merge task.
	deleteTask, emptyMap, err := task.flushAllDeletesFromNBlks(ctx)
	if err != nil {
		return
	}
	if deleteTask != nil && deleteTask.delta != nil {
		defer deleteTask.delta.Close()
	}
	/////////////////////
	//// phase seperator
	///////////////////

	phaseDesc = "1-merge ablocks"
	// merge ablocks, no need to wait, it is a sync procedure, that is why put it
	// after flushAblksForSnapshot and flushAllDeletesFromNBlks
	if err = task.mergeAblks(ctx); err != nil {
		return
	}

	/////////////////////
	//// phase seperator
	///////////////////
	phaseDesc = "1-waiting flushing appendable blocks for snapshot"
	// wait flush tasks
	if err = task.waitFlushAblkForSnapshot(ctx, snapshotSubtasks); err != nil {
		return
	}

	/////////////////////
	//// phase seperator
	///////////////////

	phaseDesc = "1-wait flushing all deletes from nablks"
	if err = task.waitFlushAllDeletesFromNAblks(ctx, deleteTask, emptyMap); err != nil {
		return
	}

	phaseDesc = "1-wait LogTxnEntry"
	txnEntry := txnentries.NewFlushTableTailEntry(
		task.txn,
		task.ID(),
		task.transMappings,
		task.rel.GetMeta().(*catalog.TableEntry),
		task.ablksMetas,
		task.nblksMetas,
		task.ablksHandles,
		task.nblksHandles,
		task.createdBlkHandles,
		task.createdDeletesObjectName,
		task.createdMergedObjectName,
		task.rt,
	)
	task.rel.GetDB()
	readset := make([]*common.ID, 0, len(task.ablksMetas)+len(task.nblksMetas))
	for _, blk := range task.ablksMetas {
		readset = append(readset, blk.AsCommonID())
	}
	for _, blk := range task.nblksMetas {
		readset = append(readset, blk.AsCommonID())
	}
	if err = task.txn.LogTxnEntry(
		task.dbid,
		task.rel.ID(),
		txnEntry,
		readset,
	); err != nil {
		return
	}
	/////////////////////

	logutil.Info("[End]", common.OperationField(task.Name()),
		zap.Int("ablks", len(task.ablksHandles)),
		zap.Int("ablks-deletes", task.ablksDeletesCnt),
		zap.Int("ablks-merge-rows", task.mergeRowsCnt),
		zap.Int("nblks", len(task.nblksHandles)),
		zap.Int("nblks-deletes", task.nblksDeletesCnt),
		common.DurationField(time.Since(now)))

	sleep, name, exist := fault.TriggerFault("slow_flush")
	if exist && name == task.schema.Name {
		time.Sleep(time.Duration(sleep) * time.Second)
	}
	return
}

// prepareAblkSortedData read the data from appendable blocks, sort them if sort key exists
func (task *flushTableTailTask) prepareAblkSortedData(ctx context.Context, blkidx int, idxs []int, sortKeyPos int) (bat *containers.Batch, empty bool, err error) {
	if len(idxs) <= 0 {
		logutil.Infof("[FlushTabletail] no mergeable columns")
		return nil, true, nil
	}
	blk := task.ablksHandles[blkidx]

	views, err := blk.GetColumnDataByIds(ctx, idxs)
	if err != nil {
		return
	}
	bat = containers.NewBatch()
	rowCntBeforeApplyDelete := views.Columns[0].Length()
	deletes := views.DeleteMask
	views.ApplyDeletes()
	defer views.Close()
	for i, colidx := range idxs {
		colview := views.Columns[i]
		if colview == nil {
			empty = true
			return
		}
		vec := colview.Orphan()
		if vec.Length() == 0 {
			empty = true
			vec.Close()
			bat.Close()
			return
		}
		bat.AddVector(task.schema.ColDefs[colidx].Name, vec)
	}

	if deletes != nil {
		task.ablksDeletesCnt += deletes.GetCardinality()
	}

	var sortMapping []int32
	if sortKeyPos >= 0 {
		logutil.Infof("flushtabletail sort blk on %s", bat.Attrs[sortKeyPos])
		sortMapping, err = mergesort.SortBlockColumns(bat.Vecs, sortKeyPos, task.rt.VectorPool.Transient)
		if err != nil {
			return
		}
	}
	task.transMappings.AddSortPhaseMapping(blkidx, rowCntBeforeApplyDelete, deletes, sortMapping)
	return
}

// mergeAblks merge the data from appendable blocks, and write the merged data to new block,
// recording row mapping in blkTransferBooking struct
func (task *flushTableTailTask) mergeAblks(ctx context.Context) (err error) {
	if len(task.ablksMetas) == 0 {
		return nil
	}
	// create new object to hold merged blocks
	var toSegmentEntry *catalog.SegmentEntry
	var toSegmentHandle handle.Segment
	if toSegmentHandle, err = task.rel.CreateNonAppendableSegment(false); err != nil {
		return
	}
	toSegmentEntry = toSegmentHandle.GetMeta().(*catalog.SegmentEntry)
	toSegmentEntry.SetSorted()

	// prepare columns idx and sortKey to read sorted batch
	schema := task.schema
	seqnums := make([]uint16, 0, len(schema.ColDefs))
	readColIdxs := make([]int, 0, len(schema.ColDefs))
	sortKeyIdx := -1
	sortKeyPos := -1
	if schema.HasSortKey() {
		sortKeyIdx = schema.GetSingleSortKeyIdx()
	}
	for i, def := range schema.ColDefs {
		if def.IsPhyAddr() {
			continue
		}
		readColIdxs = append(readColIdxs, def.Idx)
		if def.Idx == sortKeyIdx {
			sortKeyPos = i
		}
		seqnums = append(seqnums, def.SeqNum)
	}

	// read from ablocks
	readedBats := make([]*containers.Batch, 0, len(task.ablksHandles))
	for _, block := range task.ablksHandles {
		err = block.Prefetch(readColIdxs)
		if err != nil {
			return
		}
	}
	for i := range task.ablksHandles {
		bat, empty, err := task.prepareAblkSortedData(ctx, i, readColIdxs, sortKeyPos)
		if err != nil {
			return err
		}
		if empty {
			continue
		}
		readedBats = append(readedBats, bat)
	}

	for _, bat := range readedBats {
		defer bat.Close()
	}

	// prepare merge
	// pick the sort key or first column to run first merge, determing the ordering
	sortVecs := make([]containers.Vector, 0, len(readedBats))
	// fromLayout describes the layout of the input batch, which is a list of batch length
	fromLayout := make([]uint32, 0, len(readedBats))
	// toLayout describes the layout of the output batch, i.e. [8192, 8192, 8192, 4242]
	toLayout := make([]uint32, 0, len(readedBats))
	totalRowCnt := 0
	if sortKeyPos < 0 {
		// no pk, just pick the first column to reshape
		sortKeyPos = 0
	}
	for _, bat := range readedBats {
		vec := bat.Vecs[sortKeyPos]
		fromLayout = append(fromLayout, uint32(vec.Length()))
		totalRowCnt += vec.Length()
		sortVecs = append(sortVecs, vec)
	}
	task.mergeRowsCnt = totalRowCnt
	rowsLeft := totalRowCnt
	for rowsLeft > 0 {
		if rowsLeft > int(schema.BlockMaxRows) {
			toLayout = append(toLayout, schema.BlockMaxRows)
			rowsLeft -= int(schema.BlockMaxRows)
		} else {
			toLayout = append(toLayout, uint32(rowsLeft))
			break
		}
	}

	// do first sort
	allocSz := totalRowCnt * 4
	node, err := common.DefaultAllocator.Alloc(allocSz)
	if err != nil {
		panic(err)
	}
	defer common.DefaultAllocator.Free(node)
	// sortedIdx is used to shuffle other columns according to the order of the sort key
	sortedIdx := unsafe.Slice((*uint32)(unsafe.Pointer(&node[0])), totalRowCnt)
	orderedVecs, mapping := mergeColumns(sortVecs, &sortedIdx, true, fromLayout, toLayout, schema.HasSortKey(), task.rt.VectorPool.Transient)
	for _, vec := range orderedVecs {
		defer vec.Close()
	}
	task.transMappings.UpdateMappingAfterMerge(mapping, fromLayout, toLayout)

	// create blks to hold sorted data
	writtenBatches := make([]*containers.Batch, 0, len(orderedVecs))
	task.createdBlkHandles = make([]handle.Block, 0, len(orderedVecs))
	for i := range orderedVecs {
		blk, err := toSegmentHandle.CreateNonAppendableBlock(
			new(objectio.CreateBlockOpt).WithFileIdx(0).WithBlkIdx(uint16(i)))
		if err != nil {
			return err
		}
		task.createdBlkHandles = append(task.createdBlkHandles, blk)
		writtenBatches = append(writtenBatches, containers.NewBatch())
	}

	// make all columns ordered and prepared writtenBatches
	vecs := make([]containers.Vector, 0, len(readedBats))
	for i, idx := range readColIdxs {
		// skip rowid and sort(reshape) column in the first run
		if schema.ColDefs[idx].IsPhyAddr() {
			continue
		}
		if i == sortKeyPos {
			for i, vec := range orderedVecs {
				writtenBatches[i].AddVector(schema.ColDefs[idx].Name, vec)
			}
			continue
		}
		vecs = vecs[:0]
		for _, bat := range readedBats {
			vecs = append(vecs, bat.Vecs[i])
		}
		vecs, _ := mergeColumns(vecs, &sortedIdx, false, fromLayout, toLayout, schema.HasSortKey(), task.rt.VectorPool.Transient)

		for i := range vecs {
			defer vecs[i].Close()
		}
		for i, vec := range vecs {
			writtenBatches[i].AddVector(schema.ColDefs[idx].Name, vec)
		}
	}

	// write!
	name := objectio.BuildObjectName(&toSegmentEntry.ID, 0)
	writer, err := blockio.NewBlockWriterNew(task.rt.Fs.Service, name, schema.Version, seqnums)
	if err != nil {
		return err
	}
	if schema.HasPK() {
		pkIdx := schema.GetSingleSortKeyIdx()
		writer.SetPrimaryKey(uint16(pkIdx))
	}
	for _, bat := range writtenBatches {
		_, err = writer.WriteBatch(containers.ToCNBatch(bat))
		if err != nil {
			return err
		}
	}
	writtenBlocks, _, err := writer.Sync(ctx)
	if err != nil {
		return err
	}
	task.createdMergedObjectName = name.String()

	// update new status for created blocks
	var metaLoc objectio.Location
	for i, block := range writtenBlocks {
		metaLoc = blockio.EncodeLocation(name, block.GetExtent(), uint32(writtenBatches[i].Length()), block.GetID())
		if err = task.createdBlkHandles[i].UpdateMetaLoc(metaLoc); err != nil {
			return err
		}
		if err = task.createdBlkHandles[i].GetMeta().(*catalog.BlockEntry).GetBlockData().Init(); err != nil {
			return err
		}
	}

	// soft delete all ablks
	for _, blk := range task.ablksHandles {
		seg := blk.GetSegment()
		if err = seg.SoftDeleteBlock(blk.ID()); err != nil {
			return err
		}
	}

	return nil
}

// flushAblksForSnapshot schedule io task to flush ablocks for snapshot read. this function will not release any data in io task
func (task *flushTableTailTask) flushAblksForSnapshot(ctx context.Context) (subtasks []*flushBlkTask, err error) {
	defer func() {
		if err != nil {
			releaseFlushBlkTasks(subtasks)
		}
	}()
	subtasks = make([]*flushBlkTask, len(task.ablksMetas))
	// fire flush task
	for i, blk := range task.ablksMetas {
		var data, deletes *containers.Batch
		var dataVer *containers.BatchWithVersion
		blkData := blk.GetBlockData()
		if dataVer, err = blkData.CollectAppendInRange(types.TS{}, task.txn.GetStartTS(), true); err != nil {
			return
		}
		data = dataVer.Batch
		if data == nil || data.Length() == 0 {
			// the new appendable block might has no data when we flush the table, just skip it
			// In previous impl, runner will only pass non-empty blk to NewCompactBlackTask
			continue
		}
		// do not close data, leave that to wait phase
		if deletes, err = blkData.CollectDeleteInRange(ctx, types.TS{}, task.txn.GetStartTS(), true); err != nil {
			return
		}
		if deletes != nil {
			// make sure every batch in deltaloc object is sorted by rowid
			mergesort.SortBlockColumns(deletes.Vecs, 0, task.rt.VectorPool.Transient)
		}

		ablockTask := NewFlushBlkTask(
			tasks.WaitableCtx,
			dataVer.Version,
			dataVer.Seqnums,
			blkData.GetFs(),
			blk,
			data,
			deletes,
		)
		if err = task.rt.Scheduler.Schedule(ablockTask); err != nil {
			return
		}
		subtasks[i] = ablockTask
	}
	return
}

// waitFlushAblkForSnapshot waits all io tasks about flushing ablock for snapshot read, update locations
func (task *flushTableTailTask) waitFlushAblkForSnapshot(ctx context.Context, subtasks []*flushBlkTask) (err error) {
	for i, subtask := range subtasks {
		if subtask == nil {
			continue
		}
		if err = subtask.WaitDone(); err != nil {
			return
		}
		metaLocABlk := blockio.EncodeLocation(
			subtask.name,
			subtask.blocks[0].GetExtent(),
			uint32(subtask.data.Length()),
			subtask.blocks[0].GetID(),
		)
		if err = task.ablksHandles[i].UpdateMetaLoc(metaLocABlk); err != nil {
			return
		}
		if subtask.delta == nil {
			continue
		}
		deltaLoc := blockio.EncodeLocation(
			subtask.name,
			subtask.blocks[1].GetExtent(),
			uint32(subtask.delta.Length()),
			subtask.blocks[1].GetID())

		if err = task.ablksHandles[i].UpdateDeltaLoc(deltaLoc); err != nil {
			return err
		}
	}
	return nil
}

// flushAllDeletesFromNBlks collects all deletes from nblks and flush them into one block
func (task *flushTableTailTask) flushAllDeletesFromNBlks(ctx context.Context) (subtask *flushDeletesTask, emtpyDelBlkIdx *bitmap.Bitmap, err error) {
	var bufferBatch *containers.Batch
	defer func() {
		if err != nil && bufferBatch != nil {
			bufferBatch.Close()
		}
	}()
	for i, blk := range task.nblksMetas {
		blkData := blk.GetBlockData()
		var deletes *containers.Batch
		if deletes, err = blkData.CollectDeleteInRange(ctx, types.TS{}, task.txn.GetStartTS(), true); err != nil {
			return
		}
		if deletes == nil || deletes.Length() == 0 {
			if emtpyDelBlkIdx == nil {
				emtpyDelBlkIdx = &bitmap.Bitmap{}
				emtpyDelBlkIdx.InitWithSize(len(task.nblksMetas))
			}
			emtpyDelBlkIdx.Add(uint64(i))
			continue
		}
		if blkData.Rows() == deletes.Length() {
			logutil.Infof("flushtabletail blk %s should be deleted,  its all rows are deleted", blk.ID.String())
		}
		if blk.HasDropCommitted() {
			logutil.Infof("flushtabletail blk %s has been deleted, that's nice anyway", blk.ID.String())
		}
		if bufferBatch == nil {
			bufferBatch = makeDeletesTempBatch(deletes, task.rt.VectorPool.Transient)
		}
		task.nblksDeletesCnt += deletes.Length()
		// deletes is closed by Extend
		bufferBatch.Extend(deletes)
	}
	if bufferBatch != nil {
		// make sure every batch in deltaloc object is sorted by rowid
		mergesort.SortBlockColumns(bufferBatch.Vecs, 0, task.rt.VectorPool.Transient)
		subtask = NewFlushDeletesTask(tasks.WaitableCtx, task.rt.Fs, bufferBatch)
		if err = task.rt.Scheduler.Schedule(subtask); err != nil {
			return
		}
	}
	return
}

// waitFlushAllDeletesFromNAblks waits all io tasks about flushing deletes from nblks, update locations but skip those in emtpyDelBlkIdx
func (task *flushTableTailTask) waitFlushAllDeletesFromNAblks(ctx context.Context, subtask *flushDeletesTask, emtpyDelBlkIdx *bitmap.Bitmap) (err error) {
	if subtask == nil {
		return
	}
	if err = subtask.WaitDone(); err != nil {
		return err
	}
	task.createdDeletesObjectName = subtask.name.String()
	deltaLoc := blockio.EncodeLocation(
		subtask.name,
		subtask.blocks[0].GetExtent(),
		uint32(subtask.delta.Length()),
		subtask.blocks[0].GetID())
	logutil.Infof("[FlushTabletail] task %d update %s for approximate %d blks", task.ID(), deltaLoc, len(task.nblksHandles))
	for i, hdl := range task.nblksHandles {
		if emtpyDelBlkIdx != nil && emtpyDelBlkIdx.Contains(uint64(i)) {
			continue
		}
		if err = hdl.UpdateDeltaLoc(deltaLoc); err != nil {
			return err
		}
	}
	return
}

func makeDeletesTempBatch(template *containers.Batch, pool *containers.VectorPool) *containers.Batch {
	bat := containers.NewBatchWithCapacity(len(template.Attrs))
	for i, name := range template.Attrs {
		bat.AddVector(name, pool.GetVector(template.Vecs[i].GetType()))
	}
	return bat
}

func releaseFlushBlkTasks(subtasks []*flushBlkTask) {
	for _, subtask := range subtasks {
		if subtask != nil && subtask.data != nil {
			subtask.data.Close()
		}
		if subtask != nil && subtask.delta != nil {
			subtask.delta.Close()
		}
	}
}
