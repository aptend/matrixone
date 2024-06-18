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

package logtail

/*

an application on logtail mgr: build reponse to SyncLogTailRequest

More docs:
https://github.com/matrixorigin/docs/blob/main/tech-notes/dnservice/ref_logtail_impl.md


Main workflow:

          +------------------+
          | CheckpointRunner |
          +------------------+
            ^         |
            | range   | ckp & newRange
            |         v
          +------------------+  newRange  +----------------+  snapshot   +--------------+
 user ->  | HandleGetLogTail | ---------> | LogtailManager | ----------> | LogtailTable |
   ^      +------------------+            +----------------+             +--------------+
   |                                                                        |
   |           +------------------+                                         |
   +---------- |   RespBuilder    |  ------------------>+-------------------+
      return   +------------------+                     |
      entries                                           |  visit
                                                        |
                                                        v
                                  +-----------------------------------+
                                  |     txnblock2                     |
                     ...          +-----------------------------------+   ...
                                  | bornTs  | ... txn100 | txn101 |.. |
                                  +-----------------+---------+-------+
                                                    |         |
                                                    |         |
                                                    |         |
                                  +-----------------+    +----+-------+     dirty blocks
                                  |                 |    |            |
                                  v                 v    v            v
                              +-------+           +-------+       +-------+
                              | BLK-1 |           | BLK-2 |       | BLK-3 |
                              +---+---+           +---+---+       +---+---+
                                  |                   |               |
                                  v                   v               v
                            [V1@t25,disk]       [V1@t17,mem]     [V1@t17,disk]
                                  |                   |               |
                                  v                   v               v
                            [V0@t12,mem]        [V0@t10,mem]     [V0@t10,disk]
                                  |                                   |
                                  v                                   v
                            [V0@t7,mem]                           [V0@t7,mem]


*/

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/matrixorigin/matrixone/pkg/container/batch"

	"github.com/matrixorigin/matrixone/pkg/objectio"

	pkgcatalog "github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/pb/api"
	"github.com/matrixorigin/matrixone/pkg/util/fault"
	v2 "github.com/matrixorigin/matrixone/pkg/util/metric/v2"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/blockio"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/catalog"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/containers"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/iface/data"
	"go.uber.org/zap"
)

const Size90M = 90 * 1024 * 1024

type CheckpointClient interface {
	CollectCheckpointsInRange(ctx context.Context, start, end types.TS) (ckpLoc string, lastEnd types.TS, err error)
	FlushTable(ctx context.Context, dbID, tableID uint64, ts types.TS) error
}

func HandleSyncLogTailReq(
	ctx context.Context,
	ckpClient CheckpointClient,
	mgr *Manager,
	c *catalog.Catalog,
	req api.SyncLogTailReq,
	canRetry bool) (resp api.SyncLogTailResp, closeCB func(), err error) {
	now := time.Now()
	logutil.Debugf("[Logtail] begin handle %+v", req)
	defer func() {
		if elapsed := time.Since(now); elapsed > 5*time.Second {
			logutil.Infof("[Logtail] long pull cost %v, %v: %+v, %v ", elapsed, canRetry, req, err)
		}
		logutil.Debugf("[Logtail] end handle %d entries[%q], err %v", len(resp.Commands), resp.CkpLocation, err)
	}()
	start := types.BuildTS(req.CnHave.PhysicalTime, req.CnHave.LogicalTime)
	end := types.BuildTS(req.CnWant.PhysicalTime, req.CnWant.LogicalTime)
	did, tid := req.Table.DbId, req.Table.TbId
	dbEntry, err := c.GetDatabaseByID(did)
	if err != nil {
		return
	}
	tableEntry, err := dbEntry.GetTableEntryByID(tid)
	if err != nil {
		return
	}
	tableEntry.RLock()
	createTS := tableEntry.GetCreatedAtLocked()
	tableEntry.RUnlock()
	if start.Less(&createTS) {
		start = createTS
	}

	ckpLoc, checkpointed, err := ckpClient.CollectCheckpointsInRange(ctx, start, end)
	if err != nil {
		return
	}

	if checkpointed.GreaterEq(&end) {
		return api.SyncLogTailResp{
			CkpLocation: ckpLoc,
		}, nil, err
	} else if ckpLoc != "" {
		start = checkpointed.Next()
	}

	visitor := NewTableLogtailRespBuilder(ctx, ckpLoc, start, end, tableEntry)
	closeCB = visitor.Close

	operator := mgr.GetTableOperator(start, end, c, did, tid, visitor)
	if err := operator.Run(); err != nil {
		return api.SyncLogTailResp{}, visitor.Close, err
	}
	resp, err = visitor.BuildResp()

	if canRetry { // check simple conditions first
		_, name, forceFlush := fault.TriggerFault("logtail_max_size")
		if (forceFlush && name == tableEntry.GetLastestSchemaLocked().Name) || resp.ProtoSize() > Size90M {
			_ = ckpClient.FlushTable(ctx, did, tid, end)
			// try again after flushing
			newResp, closeCB, err := HandleSyncLogTailReq(ctx, ckpClient, mgr, c, req, false)
			logutil.Infof("[logtail] flush result: %d -> %d err: %v", resp.ProtoSize(), newResp.ProtoSize(), err)
			return newResp, closeCB, err
		}
	}
	return
}

type RespBuilder interface {
	catalog.Processor
	BuildResp() (api.SyncLogTailResp, error)
	Close()
}

// this is used to collect ONE ROW of db or table change
func catalogEntry2Batch[
	T *catalog.DBEntry | *catalog.TableEntry,
	N *catalog.MVCCNode[*catalog.EmptyMVCCNode] | *catalog.MVCCNode[*catalog.TableMVCCNode]](
	dstBatch *containers.Batch,
	e T,
	node N,
	schema *catalog.Schema,
	fillDataRow func(e T, node N, attr string, col containers.Vector),
	rowid types.Rowid,
	commitTs types.TS,
) {
	for _, col := range schema.ColDefs {
		fillDataRow(e, node, col.Name, dstBatch.GetVectorByName(col.Name))
	}
	dstBatch.GetVectorByName(catalog.AttrRowID).Append(rowid, false)
	dstBatch.GetVectorByName(catalog.AttrCommitTs).Append(commitTs, false)
}

// CatalogLogtailRespBuilder knows how to make api-entry from block entry.
// impl catalog.Processor interface, driven by BoundTableOperator
type TableLogtailRespBuilder struct {
	ctx context.Context
	*catalog.LoopProcessor
	start, end      types.TS
	did, tid        uint64
	dname, tname    string
	checkpoint      string
	blkMetaInsBatch *containers.Batch
	blkMetaDelBatch *containers.Batch
	objectMetaBatch *containers.Batch
	dataInsBatches  map[uint32]*containers.Batch // schema version -> data batch
	dataDelBatch    *containers.Batch
}

func NewTableLogtailRespBuilder(ctx context.Context, ckp string, start, end types.TS, tbl *catalog.TableEntry) *TableLogtailRespBuilder {
	b := &TableLogtailRespBuilder{
		ctx:           ctx,
		LoopProcessor: new(catalog.LoopProcessor),
		start:         start,
		end:           end,
		checkpoint:    ckp,
	}
	b.ObjectFn = b.VisitObj
	b.TombstoneFn = b.visitDelete

	b.did = tbl.GetDB().GetID()
	b.tid = tbl.ID
	b.dname = tbl.GetDB().GetName()
	b.tname = tbl.GetLastestSchemaLocked().Name

	b.dataInsBatches = make(map[uint32]*containers.Batch)
	b.blkMetaInsBatch = makeRespBatchFromSchema(BlkMetaSchema, common.LogtailAllocator)
	b.blkMetaDelBatch = makeRespBatchFromSchema(DelSchema, common.LogtailAllocator)
	b.objectMetaBatch = makeRespBatchFromSchema(ObjectInfoSchema, common.LogtailAllocator)
	return b
}

func (b *TableLogtailRespBuilder) Close() {
	for _, vec := range b.dataInsBatches {
		if vec != nil {
			vec.Close()
		}
	}
	b.dataInsBatches = nil
	if b.dataDelBatch != nil {
		b.dataDelBatch.Close()
		b.dataDelBatch = nil
	}
	if b.blkMetaInsBatch != nil {
		b.blkMetaInsBatch.Close()
		b.blkMetaInsBatch = nil
	}
	if b.blkMetaDelBatch != nil {
		b.blkMetaDelBatch.Close()
		b.blkMetaDelBatch = nil
	}
}

func (b *TableLogtailRespBuilder) VisitObj(e *catalog.ObjectEntry) error {
	skip, err := b.visitObjMeta(e)
	if err != nil {
		return err
	}
	if skip {
		return nil
	} else {
		return b.visitObjData(e)
	}
}
func (b *TableLogtailRespBuilder) visitObjMeta(e *catalog.ObjectEntry) (bool, error) {
	mvccNodes := e.ClonePreparedInRange(b.start, b.end)
	if len(mvccNodes) == 0 {
		return false, nil
	}

	for _, node := range mvccNodes {
		if e.IsAppendable() && node.BaseNode.IsEmpty() {
			continue
		}
		visitObject(b.objectMetaBatch, e, node, false, types.TS{})
	}
	return b.skipObjectData(e, mvccNodes[len(mvccNodes)-1]), nil
}
func (b *TableLogtailRespBuilder) skipObjectData(e *catalog.ObjectEntry, lastMVCCNode *catalog.MVCCNode[*catalog.ObjectMVCCNode]) bool {
	if e.IsAppendable() {
		// appendable block has been flushed, no need to collect data
		return !lastMVCCNode.BaseNode.IsEmpty()
	} else {
		return true
	}
}
func (b *TableLogtailRespBuilder) visitObjData(e *catalog.ObjectEntry) error {
	data := e.GetObjectData()
	insBatch, err := data.CollectAppendInRange(b.start, b.end, false, common.LogtailAllocator)
	if err != nil {
		return err
	}
	if insBatch != nil && insBatch.Length() > 0 {
		dest, ok := b.dataInsBatches[insBatch.Version]
		if !ok {
			// create new dest batch
			dest = DataChangeToLogtailBatch(insBatch)
			b.dataInsBatches[insBatch.Version] = dest
		} else {
			dest.Extend(insBatch.Batch)
			// insBatch is freed, don't use anymore
		}
	}
	return nil
}
func visitObject(batch *containers.Batch, entry *catalog.ObjectEntry, node *catalog.MVCCNode[*catalog.ObjectMVCCNode], push bool, committs types.TS) {
	batch.GetVectorByName(catalog.AttrRowID).Append(objectio.HackObjid2Rowid(&entry.ID), false)
	if push {
		batch.GetVectorByName(catalog.AttrCommitTs).Append(committs, false)
	} else {
		batch.GetVectorByName(catalog.AttrCommitTs).Append(node.TxnMVCCNode.End, false)
	}
	node.BaseNode.AppendTuple(&entry.ID, batch)
	if push {
		node.TxnMVCCNode.AppendTupleWithCommitTS(batch, committs)
	} else {
		node.TxnMVCCNode.AppendTuple(batch)
	}
	if push {
		node.EntryMVCCNode.AppendTupleWithCommitTS(batch, committs)
	} else {
		node.EntryMVCCNode.AppendTuple(batch)
	}
	batch.GetVectorByName(SnapshotAttr_DBID).Append(entry.GetTable().GetDB().ID, false)
	batch.GetVectorByName(SnapshotAttr_TID).Append(entry.GetTable().ID, false)
	batch.GetVectorByName(ObjectAttr_State).Append(entry.IsAppendable(), false)
	sorted := false
	if entry.GetTable().GetLastestSchemaLocked().HasSortKey() && !entry.IsAppendable() {
		sorted = true
	}
	batch.GetVectorByName(ObjectAttr_Sorted).Append(sorted, false)
}

func (b *TableLogtailRespBuilder) visitDelete(e data.Tombstone) error {
	deletes, _, _, err := e.VisitDeletes(b.ctx, b.start, b.end, b.blkMetaInsBatch, nil, false)
	if err != nil {
		return err
	}
	if deletes != nil && deletes.Length() != 0 {
		if b.dataDelBatch == nil {
			b.dataDelBatch = deletes
		} else {
			b.dataDelBatch.Extend(deletes)
			deletes.Close()
		}
	}
	return nil
}

type TableRespKind int

const (
	TableRespKind_Data TableRespKind = iota
	TableRespKind_Blk
	TableRespKind_Obj
)

func (b *TableLogtailRespBuilder) BuildResp() (api.SyncLogTailResp, error) {
	entries := make([]*api.Entry, 0)
	tryAppendEntry := func(typ api.Entry_EntryType, kind TableRespKind, batch *containers.Batch, version uint32) error {
		if batch == nil || batch.Length() == 0 {
			return nil
		}
		bat, err := containersBatchToProtoBatch(batch)
		if err != nil {
			return err
		}

		if b.tid == pkgcatalog.MO_DATABASE_ID || b.tid == pkgcatalog.MO_TABLES_ID {
			switch kind {
			case TableRespKind_Data:
				logutil.Infof("[yyyy pull] table data [%v] %d-%s-%d: %s", typ, b.tid, b.tname, version,
					DebugBatchToString("data", batch, false, zap.InfoLevel))
			case TableRespKind_Blk:
				logutil.Infof("[yyyy pull] blk meta [%v] %d-%s: %s", typ, b.tid, b.tname,
					DebugBatchToString("blkmeta", batch, false, zap.InfoLevel))
			case TableRespKind_Obj:
				logutil.Infof("[yyyy pull] obj meta [%v] %d-%s: %s", typ, b.tid, b.tname,
					DebugBatchToString("object", batch, false, zap.InfoLevel))
			}
		}
		tableName := ""
		switch kind {
		case TableRespKind_Data:
			tableName = b.tname
			logutil.Debugf("[logtail] table data [%v] %d-%s-%d: %s", typ, b.tid, b.tname, version,
				DebugBatchToString("data", batch, false, zap.InfoLevel))
		case TableRespKind_Blk:
			tableName = fmt.Sprintf("_%d_meta", b.tid)
			logutil.Debugf("[logtail] table meta [%v] %d-%s: %s", typ, b.tid, b.tname,
				DebugBatchToString("blkmeta", batch, false, zap.InfoLevel))
		case TableRespKind_Obj:
			tableName = fmt.Sprintf("_%d_obj", b.tid)
			logutil.Debugf("[logtail] table meta [%v] %d-%s: %s", typ, b.tid, b.tname,
				DebugBatchToString("object", batch, false, zap.InfoLevel))
		}

		entry := &api.Entry{
			EntryType:    typ,
			TableId:      b.tid,
			TableName:    tableName,
			DatabaseId:   b.did,
			DatabaseName: b.dname,
			Bat:          bat,
		}
		entries = append(entries, entry)
		return nil
	}

	empty := api.SyncLogTailResp{}
	if err := tryAppendEntry(api.Entry_Insert, TableRespKind_Blk, b.blkMetaInsBatch, 0); err != nil {
		return empty, err
	}
	if err := tryAppendEntry(api.Entry_Delete, TableRespKind_Blk, b.blkMetaDelBatch, 0); err != nil {
		return empty, err
	}
	if err := tryAppendEntry(api.Entry_Insert, TableRespKind_Obj, b.objectMetaBatch, 0); err != nil {
		return empty, err
	}
	keys := make([]uint32, 0, len(b.dataInsBatches))
	for k := range b.dataInsBatches {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		if err := tryAppendEntry(api.Entry_Insert, TableRespKind_Data, b.dataInsBatches[k], k); err != nil {
			return empty, err
		}
	}
	if err := tryAppendEntry(api.Entry_Delete, TableRespKind_Data, b.dataDelBatch, 0); err != nil {
		return empty, err
	}

	return api.SyncLogTailResp{
		CkpLocation: b.checkpoint,
		Commands:    entries,
	}, nil
}
func GetMetaIdxesByVersion(ver uint32) []uint16 {
	meteIdxSchema := checkpointDataReferVersions[ver][MetaIDX]
	idxes := make([]uint16, len(meteIdxSchema.attrs))
	for attr := range meteIdxSchema.attrs {
		idxes[attr] = uint16(attr)
	}
	return idxes
}
func LoadCheckpointEntries(
	ctx context.Context,
	metLoc string,
	tableID uint64,
	tableName string,
	dbID uint64,
	dbName string,
	mp *mpool.MPool,
	fs fileservice.FileService) ([]*api.Entry, []func(), error) {
	if metLoc == "" {
		return nil, nil, nil
	}
	v2.LogtailLoadCheckpointCounter.Inc()
	now := time.Now()
	defer func() {
		v2.LogTailLoadCheckpointDurationHistogram.Observe(time.Since(now).Seconds())
	}()
	locationsAndVersions := strings.Split(metLoc, ";")

	datas := make([]*CNCheckpointData, len(locationsAndVersions)/2)

	readers := make([]*blockio.BlockReader, len(locationsAndVersions)/2)
	objectLocations := make([]objectio.Location, len(locationsAndVersions)/2)
	versions := make([]uint32, len(locationsAndVersions)/2)
	locations := make([]objectio.Location, len(locationsAndVersions)/2)
	for i := 0; i < len(locationsAndVersions); i += 2 {
		key := locationsAndVersions[i]
		version, err := strconv.ParseUint(locationsAndVersions[i+1], 10, 32)
		if err != nil {
			return nil, nil, err
		}
		location, err := blockio.EncodeLocationFromString(key)
		if err != nil {
			return nil, nil, err
		}
		locations[i/2] = location
		reader, err := blockio.NewObjectReader(fs, location)
		if err != nil {
			return nil, nil, err
		}
		readers[i/2] = reader
		err = blockio.PrefetchMeta(fs, location)
		if err != nil {
			return nil, nil, err
		}
		objectLocations[i/2] = location
		versions[i/2] = uint32(version)
	}

	for i := range objectLocations {
		data := NewCNCheckpointData()
		meteIdxSchema := checkpointDataReferVersions[versions[i]][MetaIDX]
		idxes := make([]uint16, len(meteIdxSchema.attrs))
		for attr := range meteIdxSchema.attrs {
			idxes[attr] = uint16(attr)
		}
		err := data.PrefetchMetaIdx(ctx, versions[i], idxes, objectLocations[i], fs)
		if err != nil {
			return nil, nil, err
		}
		datas[i] = data
	}

	for i := range datas {
		err := datas[i].InitMetaIdx(ctx, versions[i], readers[i], locations[i], mp)
		if err != nil {
			return nil, nil, err
		}
	}

	for i := range datas {
		err := datas[i].PrefetchMetaFrom(ctx, versions[i], locations[i], fs, tableID)
		if err != nil {
			return nil, nil, err
		}
	}

	for i := range datas {
		err := datas[i].PrefetchFrom(ctx, versions[i], fs, locations[i], tableID)
		if err != nil {
			return nil, nil, err
		}
	}

	closeCBs := make([]func(), 0)
	bats := make([][]*batch.Batch, len(locationsAndVersions)/2)
	var err error
	for i, data := range datas {
		var bat []*batch.Batch
		bat, err = data.ReadFromData(ctx, tableID, locations[i], readers[i], versions[i], mp)
		closeCBs = append(closeCBs, data.GetCloseCB(versions[i], mp))
		if err != nil {
			for j := range closeCBs {
				if closeCBs[j] != nil {
					closeCBs[j]()
				}
			}
			return nil, nil, err
		}
		bats[i] = bat
	}

	entries := make([]*api.Entry, 0)
	for i := range objectLocations {
		data := datas[i]
		ins, del, cnIns, objInfo, err := data.GetTableDataFromBats(tableID, bats[i])
		if err != nil {
			for j := range closeCBs {
				if closeCBs[j] != nil {
					closeCBs[j]()
				}
			}
			return nil, nil, err
		}
		if tableName != pkgcatalog.MO_DATABASE &&
			tableName != pkgcatalog.MO_COLUMNS &&
			tableName != pkgcatalog.MO_TABLES {
			tableName = fmt.Sprintf("_%d_meta", tableID)
		}
		if ins != nil {
			entry := &api.Entry{
				EntryType:    api.Entry_Insert,
				TableId:      tableID,
				TableName:    tableName,
				DatabaseId:   dbID,
				DatabaseName: dbName,
				Bat:          ins,
			}
			entries = append(entries, entry)
		}
		if cnIns != nil {
			entry := &api.Entry{
				EntryType:    api.Entry_Insert,
				TableId:      tableID,
				TableName:    tableName,
				DatabaseId:   dbID,
				DatabaseName: dbName,
				Bat:          cnIns,
			}
			entries = append(entries, entry)
		}
		if del != nil {
			entry := &api.Entry{
				EntryType:    api.Entry_Delete,
				TableId:      tableID,
				TableName:    tableName,
				DatabaseId:   dbID,
				DatabaseName: dbName,
				Bat:          del,
			}
			entries = append(entries, entry)
		}
		if objInfo != nil {
			entry := &api.Entry{
				EntryType:    api.Entry_Insert,
				TableId:      tableID,
				TableName:    fmt.Sprintf("_%d_obj", tableID),
				DatabaseId:   dbID,
				DatabaseName: dbName,
				Bat:          objInfo,
			}
			entries = append(entries, entry)
		}
	}
	return entries, closeCBs, nil
}
