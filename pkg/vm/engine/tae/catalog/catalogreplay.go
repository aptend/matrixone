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
package catalog

import (
	"context"
	"fmt"

	pkgcatalog "github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/containers"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/iface/data"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/iface/txnif"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/txn/txnbase"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/wal"
)

const (
	Backup_Object_Offset uint16 = 1000
)

//#region Replay WAL related

func (catalog *Catalog) ReplayCmd(
	txncmd txnif.TxnCmd,
	dataFactory DataFactory,
	observer wal.ReplayObserver) {
	switch txncmd.GetType() {
	case txnbase.IOET_WALTxnCommand_Composed:
		cmds := txncmd.(*txnbase.ComposedCmd)
		for _, cmds := range cmds.Cmds {
			catalog.ReplayCmd(cmds, dataFactory, observer)
		}
	case IOET_WALTxnCommand_Database:
		cmd := txncmd.(*EntryCommand[*EmptyMVCCNode, *DBNode])
		catalog.onReplayUpdateDatabase(cmd, observer)
	case IOET_WALTxnCommand_Table:
		cmd := txncmd.(*EntryCommand[*TableMVCCNode, *TableNode])
		catalog.onReplayUpdateTable(cmd, dataFactory, observer)
	case IOET_WALTxnCommand_Object:
		cmd := txncmd.(*EntryCommand[*ObjectMVCCNode, *ObjectNode])
		catalog.onReplayUpdateObject(cmd, dataFactory, observer)
	case IOET_WALTxnCommand_Block:
		cmd := txncmd.(*EntryCommand[*MetadataMVCCNode, *BlockNode])
		catalog.onReplayUpdateBlock(cmd, dataFactory, observer)
	case IOET_WALTxnCommand_Segment:
		// segment is deprecated
		return
	default:
		panic("unsupport")
	}
}

func (catalog *Catalog) onReplayUpdateDatabase(cmd *EntryCommand[*EmptyMVCCNode, *DBNode], _ wal.ReplayObserver) {
	catalog.OnReplayDBID(cmd.ID.DbID)
	var err error
	un := cmd.mvccNode

	db, err := catalog.GetDatabaseByID(cmd.ID.DbID)
	if err != nil {
		db = NewReplayDBEntry()
		db.ID = cmd.ID.DbID
		db.catalog = catalog
		db.DBNode = cmd.node
		db.Insert(un)
		err = catalog.AddEntryLocked(db, un.GetTxn(), false)
		if err != nil {
			panic(err)
		}
		return
	}

	dbun := db.SearchNodeLocked(un)
	if dbun == nil {
		db.Insert(un)
	} else {
		return
		// panic(fmt.Sprintf("logic err: duplicate node %v and %v", dbun.String(), un.String()))
	}
}

func (catalog *Catalog) onReplayUpdateTable(cmd *EntryCommand[*TableMVCCNode, *TableNode], dataFactory DataFactory, _ wal.ReplayObserver) {
	catalog.OnReplayTableID(cmd.ID.TableID)
	// prepareTS := cmd.GetTs()
	// if prepareTS.LessEq(catalog.GetCheckpointed().MaxTS) {
	// 	if observer != nil {
	// 		observer.OnStaleIndex(idx)
	// 	}
	// 	return
	// }
	db, err := catalog.GetDatabaseByID(cmd.ID.DbID)
	if err != nil {
		panic(err)
	}
	tbl, err := db.GetTableEntryByID(cmd.ID.TableID)

	un := cmd.mvccNode
	if err != nil {
		tbl = NewReplayTableEntry()
		tbl.ID = cmd.ID.TableID
		tbl.db = db
		tbl.tableData = dataFactory.MakeTableFactory()(tbl)
		tbl.TableNode = cmd.node
		tbl.TableNode.schema.Store(un.BaseNode.Schema)
		tbl.Insert(un)
		err = db.AddEntryLocked(tbl, un.GetTxn(), true)
		if err != nil {
			logutil.Warn(catalog.SimplePPString(common.PPL3))
			panic(err)
		}
		return
	}
	tblun := tbl.SearchNodeLocked(un)
	if tblun == nil {
		tbl.Insert(un) //TODO isvalid
		if tbl.isColumnChangedInSchema() {
			tbl.FreezeAppend()
		}
		schema := un.BaseNode.Schema
		tbl.TableNode.schema.Store(schema)
		// alter table rename
		if schema.Extra.OldName != "" && un.DeletedAt.IsEmpty() {
			err := tbl.db.RenameTableInTxn(schema.Extra.OldName, schema.Name, tbl.ID, schema.AcInfo.TenantID, un.GetTxn(), true)
			if err != nil {
				logutil.Warn(schema.String())
				panic(err)
			}
		}
	}

}

func (catalog *Catalog) onReplayUpdateObject(
	cmd *EntryCommand[*ObjectMVCCNode, *ObjectNode],
	dataFactory DataFactory,
	_ wal.ReplayObserver) {
	catalog.OnReplayObjectID(cmd.node.SortHint)

	db, err := catalog.GetDatabaseByID(cmd.ID.DbID)
	if err != nil {
		// a db is dropped before checkpoint end
		// and its tables are flushed after the checkpoint end,
		// it is normal to for WAL to miss the db
		if moerr.IsMoErrCode(err, moerr.OkExpectedEOB) {
			return
		}
		panic(err)
	}
	rel, err := db.GetTableEntryByID(cmd.ID.TableID)
	if err != nil {
		if moerr.IsMoErrCode(err, moerr.OkExpectedEOB) {
			return
		}
		panic(err)
	}
	var obj *ObjectEntry
	if cmd.mvccNode.CreatedAt.Equal(&txnif.UncommitTS) {
		obj = NewReplayObjectEntry()
		obj.table = rel
		obj.ObjectNode = *cmd.node
		obj.SortHint = catalog.NextObject()
		obj.EntryMVCCNode = *cmd.mvccNode.EntryMVCCNode
		obj.CreateNode = *cmd.mvccNode.TxnMVCCNode
		cmd.mvccNode.TxnMVCCNode = &obj.CreateNode
		cmd.mvccNode.EntryMVCCNode = &obj.EntryMVCCNode
		obj.ObjectMVCCNode = *cmd.mvccNode.BaseNode
		obj.remainingRows = &common.FixedSampleIII[int]{}
		obj.ObjectState = ObjectState_Create_ApplyCommit
		rel.AddEntryLocked(obj)
	}
	if cmd.mvccNode.DeletedAt.Equal(&txnif.UncommitTS) {
		obj, err = rel.GetObjectByID(cmd.ID.ObjectID())
		if err != nil {
			panic(fmt.Sprintf("obj %v not existed, table:\n%v", cmd.ID.String(), rel.StringWithLevel(3)))
		}
		obj.EntryMVCCNode = *cmd.mvccNode.EntryMVCCNode
		obj.DeleteNode = *cmd.mvccNode.TxnMVCCNode
		obj.ObjectMVCCNode = *cmd.mvccNode.BaseNode
		cmd.mvccNode.TxnMVCCNode = &obj.DeleteNode
		cmd.mvccNode.EntryMVCCNode = &obj.EntryMVCCNode
		obj.ObjectState = ObjectState_Delete_ApplyCommit
	}

	if obj.objData == nil {
		obj.objData = dataFactory.MakeObjectFactory()(obj)
	} else {
		deleteAt := obj.GetDeleteAt()
		if !obj.IsAppendable() || (obj.IsAppendable() && !deleteAt.IsEmpty()) {
			obj.objData.TryUpgrade()
			obj.objData.UpgradeAllDeleteChain()
		}
	}
}

func (catalog *Catalog) onReplayUpdateBlock(
	cmd *EntryCommand[*MetadataMVCCNode, *BlockNode],
	dataFactory DataFactory,
	_ wal.ReplayObserver) {
	// catalog.OnReplayBlockID(cmd.ID.BlockID)
	db, err := catalog.GetDatabaseByID(cmd.ID.DbID)
	if err != nil {
		if moerr.IsMoErrCode(err, moerr.OkExpectedEOB) {
			return
		}
		panic(err)
	}
	tbl, err := db.GetTableEntryByID(cmd.ID.TableID)
	if err != nil {
		if moerr.IsMoErrCode(err, moerr.OkExpectedEOB) {
			return
		}
		panic(err)
	}
	if !cmd.mvccNode.BaseNode.DeltaLoc.IsEmpty() {
		obj, err := tbl.GetObjectByID(cmd.ID.ObjectID())
		if obj == nil {
			logutil.Fatalf("obj %v not found, mvcc node: %v", cmd.ID.String(), cmd.mvccNode.String())
			return
		}
		if err != nil {
			panic(err)
		}
		tombstone := tbl.GetOrCreateTombstone(obj, dataFactory.MakeTombstoneFactory())
		_, blkOffset := cmd.ID.BlockID.Offsets()
		tombstone.ReplayDeltaLoc(cmd.mvccNode, blkOffset)
	}
}

//#endregion

//#region Replay Checkpoint related

func (catalog *Catalog) onReplayCreateDB(
	dbid uint64, name string, txnNode *txnbase.TxnMVCCNode,
	tenantID, userID, roleID uint32, createAt types.Timestamp, createSql, datType string) {
	catalog.OnReplayDBID(dbid)
	db, _ := catalog.GetDatabaseByID(dbid)
	if db != nil {
		dbCreatedAt := db.GetCreatedAtLocked()
		if !dbCreatedAt.Equal(&txnNode.End) {
			panic(moerr.NewInternalErrorNoCtx("logic err expect %s, get %s",
				txnNode.End.ToString(), dbCreatedAt.ToString()))
		}
		return
	}
	db = NewReplayDBEntry()
	db.catalog = catalog
	db.ID = dbid
	db.DBNode = &DBNode{
		acInfo: accessInfo{
			TenantID: tenantID,
			UserID:   userID,
			RoleID:   roleID,
			CreateAt: createAt,
		},
		createSql: createSql,
		datType:   datType,
		name:      name,
	}
	_ = catalog.AddEntryLocked(db, nil, true)
	un := &MVCCNode[*EmptyMVCCNode]{
		EntryMVCCNode: &EntryMVCCNode{
			CreatedAt: txnNode.End,
		},
		TxnMVCCNode: txnNode,
	}
	db.Insert(un)
}

func (catalog *Catalog) onReplayCreateTable(dbid, tid uint64, schema *Schema, txnNode *txnbase.TxnMVCCNode, dataFactory DataFactory) {
	catalog.OnReplayTableID(tid)
	db, err := catalog.GetDatabaseByID(dbid)
	if err != nil {
		// logutil.Info(catalog.SimplePPString(common.PPL3))
		panic(err)
	}
	tbl, _ := db.GetTableEntryByID(tid)
	if tbl != nil {
		tblCreatedAt := tbl.GetCreatedAtLocked()
		if tblCreatedAt.Greater(&txnNode.End) {
			panic(moerr.NewInternalErrorNoCtxf("logic err expect %s, get %s", txnNode.End.ToString(), tblCreatedAt.ToString()))
		}
		// alter table
		un := &MVCCNode[*TableMVCCNode]{
			EntryMVCCNode: &EntryMVCCNode{
				CreatedAt: tblCreatedAt,
			},
			TxnMVCCNode: txnNode,
			BaseNode: &TableMVCCNode{
				Schema: schema,
			},
		}
		tbl.Insert(un)
		if tbl.isColumnChangedInSchema() {
			tbl.FreezeAppend()
		}
		tbl.TableNode.schema.Store(schema)
		if schema.Extra.OldName != "" {
			logutil.Infof("replay rename %v from %v -> %v", tid, schema.Extra.OldName, schema.Name)
			err := tbl.db.RenameTableInTxn(schema.Extra.OldName, schema.Name, tbl.ID, schema.AcInfo.TenantID, un.GetTxn(), true)
			if err != nil {
				logutil.Warn(schema.String())
				panic(err)
			}
		}

		return
	}
	tbl = NewReplayTableEntry()
	tbl.TableNode = &TableNode{}
	tbl.TableNode.schema.Store(schema)
	tbl.db = db
	tbl.ID = tid
	tbl.tableData = dataFactory.MakeTableFactory()(tbl)
	_ = db.AddEntryLocked(tbl, nil, true)
	un := &MVCCNode[*TableMVCCNode]{
		EntryMVCCNode: &EntryMVCCNode{
			CreatedAt: txnNode.End,
		},
		TxnMVCCNode: txnNode,
		BaseNode: &TableMVCCNode{
			Schema: schema,
		},
	}
	tbl.Insert(un)
}

func (catalog *Catalog) OnReplayObjectBatch(objectInfo *containers.Batch, dataFactory DataFactory, forSys bool) {
	for i := 0; i < objectInfo.Length(); i++ {
		tid := objectInfo.GetVectorByName(SnapshotAttr_TID).Get(i).(uint64)
		if pkgcatalog.IsSystemTable(tid) != forSys {
			continue
		}
		dbid := objectInfo.GetVectorByName(SnapshotAttr_DBID).Get(i).(uint64)
		objectNode := ReadObjectInfoTuple(objectInfo, i)
		sid := objectNode.ObjectName().ObjectId()
		txnNode := txnbase.ReadTuple(objectInfo, i)
		entryNode := ReadEntryNodeTuple(objectInfo, i)
		state := objectInfo.GetVectorByName(ObjectAttr_State).Get(i).(bool)
		entryState := ES_Appendable
		if !state {
			entryState = ES_NotAppendable
		}
		catalog.onReplayCheckpointObject(dbid, tid, sid, objectNode, entryNode, txnNode, entryState, dataFactory)
	}
}

func (catalog *Catalog) onReplayCheckpointObject(
	dbid, tbid uint64,
	objid *types.Objectid,
	objNode *ObjectMVCCNode,
	entryNode *EntryMVCCNode,
	txnNode *txnbase.TxnMVCCNode,
	state EntryState,
	dataFactory DataFactory,
) {
	db, err := catalog.GetDatabaseByID(dbid)
	if err != nil {
		// As only replay the catalog view at the end time of lastest checkpoint
		// it is normal fot deleted db or table to be missed
		if moerr.IsMoErrCode(err, moerr.OkExpectedEOB) {
			return
		}
		panic(err)
	}
	rel, err := db.GetTableEntryByID(tbid)
	if err != nil {
		if moerr.IsMoErrCode(err, moerr.OkExpectedEOB) {
			return
		}
		panic(err)
	}
	newObject := func() *ObjectEntry {
		object := NewReplayObjectEntry()
		object.table = rel
		object.ObjectNode = ObjectNode{
			state:    state,
			sorted:   state == ES_NotAppendable,
			SortHint: catalog.NextObject(),
		}
		object.EntryMVCCNode = *entryNode
		object.ObjectMVCCNode = *objNode
		object.CreateNode = *txnNode
		object.remainingRows = &common.FixedSampleIII[int]{}
		object.forcePNode = true
		object.ObjectState = ObjectState_Create_ApplyCommit
		return object
	}
	var obj *ObjectEntry
	if entryNode.CreatedAt.Equal(&txnNode.End) {
		obj = newObject()
		rel.AddEntryLocked(obj)
	}
	if entryNode.DeletedAt.Equal(&txnNode.End) {
		obj, err = rel.GetObjectByID(objid)
		if err != nil {
			panic(fmt.Sprintf("obj %v, [%v %v %v] not existed, table:\n%v", objid.String(),
				entryNode.String(), objNode.String(),
				txnNode.String(), rel.StringWithLevel(3)))
		}
		obj.EntryMVCCNode = *entryNode
		obj.ObjectMVCCNode = *objNode
		obj.DeleteNode = *txnNode
		obj.ObjectState = ObjectState_Delete_ApplyCommit
	}
	if !entryNode.CreatedAt.Equal(&txnNode.End) && !entryNode.DeletedAt.Equal(&txnNode.End) {
		// In back up, aobj is replaced with naobj and its DeleteAt is removed.
		// Before back up, txnNode.End equals DeleteAt of naobj.
		// After back up, DeleteAt is empty.
		if objid.Offset() == Backup_Object_Offset && entryNode.DeletedAt.IsEmpty() {
			obj = newObject()
			rel.AddEntryLocked(obj)
			logutil.Warnf("obj %v, tbl %v-%d delete %v, create %v, end %v",
				objid.String(), rel.fullName, rel.ID, entryNode.CreatedAt.ToString(),
				entryNode.DeletedAt.ToString(), txnNode.End.ToString())
		} else {
			panic(fmt.Sprintf("logic error: obj %v, tbl %v-%d create %v, delete %v, end %v",
				objid.String(), rel.fullName, rel.ID, entryNode.CreatedAt.ToString(),
				entryNode.DeletedAt.ToString(), txnNode.End.ToString()))
		}
	}
	if obj.objData == nil {
		obj.objData = dataFactory.MakeObjectFactory()(obj)
	} else {
		deleteAt := obj.GetDeleteAt()
		if !obj.IsAppendable() || (obj.IsAppendable() && !deleteAt.IsEmpty()) {
			obj.objData.TryUpgrade()
			obj.objData.UpgradeAllDeleteChain()
		}
	}
}

// TODO: ensure bat does not exceed the 1G limit
func readSysTableBatch(ctx context.Context, entry *TableEntry, readTxn txnif.AsyncTxn) *containers.Batch {
	schema := entry.GetLastestSchema()
	it := entry.MakeObjectIt(false)
	defer it.Release()
	colIdxes := make([]int, 0, len(schema.ColDefs))
	for _, col := range schema.ColDefs {
		if col.IsPhyAddr() {
			continue
		}
		colIdxes = append(colIdxes, col.Idx)
	}
	var bat *containers.Batch
	for it.Next() {
		obj := it.Item()
		if !obj.IsVisible(readTxn) {
			continue
		}
		for blkid := range obj.BlockCnt() {
			blkbat, err := obj.GetObjectData().GetColumnDataByIds(
				ctx,
				readTxn,
				schema,
				uint16(blkid),
				colIdxes,
				common.CheckpointAllocator)
			if err != nil {
				panic(err)
			}
			if blkbat == nil {
				panic(fmt.Sprintf("blkbat is nil, obj %v, blkid %v, seemed not fluhsed", obj.ID().String(), blkid))
			}
			if bat == nil {
				bat = blkbat
				if bat.Deletes == nil {
					bat.Deletes = nulls.NewWithSize(1024)
				}
			} else {
				len := uint64(bat.Length())
				bat.Extend(blkbat)
				if blkbat.Deletes != nil {
					blkbat.Deletes.Foreach(func(u uint64) bool {
						bat.Deletes.Set(len + u)
						return true
					})
				}
			}
		}
	}
	if bat != nil {
		bat.ShrinkDeletes()
	}
	return bat
}

func (catalog *Catalog) RelayFromSysTableObjects(
	ctx context.Context,
	readTxn txnif.AsyncTxn,
	dataFactory DataFactory,
	sortFunc func([]containers.Vector, int) error) {
	db, err := catalog.GetDatabaseByID(pkgcatalog.MO_CATALOG_ID)
	if err != nil {
		panic(err)
	}
	dbTbl, err := db.GetTableEntryByID(pkgcatalog.MO_DATABASE_ID)
	if err != nil {
		panic(err)
	}
	tableTbl, err := db.GetTableEntryByID(pkgcatalog.MO_TABLES_ID)
	if err != nil {
		panic(err)
	}
	columnTbl, err := db.GetTableEntryByID(pkgcatalog.MO_COLUMNS_ID)
	if err != nil {
		panic(err)
	}
	txnNode := &txnbase.TxnMVCCNode{
		Start:   readTxn.GetStartTS(),
		Prepare: readTxn.GetStartTS(),
		End:     readTxn.GetStartTS(),
	}

	dbBatch := readSysTableBatch(ctx, dbTbl, readTxn)
	if dbBatch != nil {
		defer dbBatch.Close()
		catalog.ReplayMODatabase(ctx, txnNode, dbBatch)
	}

	tableBatch := readSysTableBatch(ctx, tableTbl, readTxn)
	if tableBatch != nil {
		if err := sortFunc(tableBatch.Vecs, 0); err != nil {
			panic(err)
		}
		defer tableBatch.Close()
		columnBatch := readSysTableBatch(ctx, columnTbl, readTxn)
		if err := sortFunc(columnBatch.Vecs, 0); err != nil {
			panic(err)
		}
		defer columnBatch.Close()
		catalog.ReplayMOTables(ctx, txnNode, dataFactory, tableBatch, columnBatch)
	}
	logutil.Infof(catalog.SimplePPString(common.PPL3))
}

func (catalog *Catalog) ReplayMODatabase(ctx context.Context, txnNode *txnbase.TxnMVCCNode, bat *containers.Batch) {
	for i := 0; i < bat.Length(); i++ {
		dbid := bat.GetVectorByName(pkgcatalog.SystemDBAttr_ID).Get(i).(uint64)
		name := string(bat.GetVectorByName(pkgcatalog.SystemDBAttr_Name).Get(i).([]byte))
		tenantID := bat.GetVectorByName(pkgcatalog.SystemDBAttr_AccID).Get(i).(uint32)
		userID := bat.GetVectorByName(pkgcatalog.SystemDBAttr_Creator).Get(i).(uint32)
		roleID := bat.GetVectorByName(pkgcatalog.SystemDBAttr_Owner).Get(i).(uint32)
		createAt := bat.GetVectorByName(pkgcatalog.SystemDBAttr_CreateAt).Get(i).(types.Timestamp)
		createSql := string(bat.GetVectorByName(pkgcatalog.SystemDBAttr_CreateSQL).Get(i).([]byte))
		datType := string(bat.GetVectorByName(pkgcatalog.SystemDBAttr_Type).Get(i).([]byte))
		catalog.onReplayCreateDB(dbid, name, txnNode, tenantID, userID, roleID, createAt, createSql, datType)
	}
}

func (catalog *Catalog) ReplayMOTables(ctx context.Context, txnNode *txnbase.TxnMVCCNode, dataF DataFactory, tblBat, colBat *containers.Batch) {
	schemaOffset := 0
	for i := 0; i < tblBat.Length(); i++ {
		tid := tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_ID).Get(i).(uint64)
		dbid := tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_DBID).Get(i).(uint64)
		name := string(tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_Name).Get(i).([]byte))
		schema := NewEmptySchema(name)
		schemaOffset = schema.ReadFromBatch(colBat, schemaOffset, tid)
		schema.Comment = string(tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_Comment).Get(i).([]byte))
		schema.Version = tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_Version).Get(i).(uint32)
		schema.CatalogVersion = tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_CatalogVersion).Get(i).(uint32)
		schema.Partitioned = tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_Partitioned).Get(i).(int8)
		schema.Partition = string(tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_Partition).Get(i).([]byte))
		schema.Relkind = string(tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_Kind).Get(i).([]byte))
		schema.Createsql = string(tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_CreateSQL).Get(i).([]byte))
		schema.View = string(tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_ViewDef).Get(i).([]byte))
		schema.Constraint = tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_Constraint).Get(i).([]byte)
		schema.AcInfo = accessInfo{}
		schema.AcInfo.RoleID = tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_Owner).Get(i).(uint32)
		schema.AcInfo.UserID = tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_Creator).Get(i).(uint32)
		schema.AcInfo.CreateAt = tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_CreateAt).Get(i).(types.Timestamp)
		schema.AcInfo.TenantID = tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_AccID).Get(i).(uint32)
		extra := tblBat.GetVectorByName(pkgcatalog.SystemRelAttr_ExtraInfo).Get(i).([]byte)
		schema.MustRestoreExtra(extra)
		if err := schema.Finalize(true); err != nil {
			panic(err)
		}
		catalog.onReplayCreateTable(dbid, tid, schema, txnNode, dataF)
	}
}

func (catalog *Catalog) OnReplayBlockBatch(ins, insTxn, del, delTxn *containers.Batch, dataFactory DataFactory, forSys bool) {
	for i := 0; i < ins.Length(); i++ {
		tid := insTxn.GetVectorByName(SnapshotAttr_TID).Get(i).(uint64)
		if pkgcatalog.IsSystemTable(tid) != forSys {
			continue
		}
		dbid := insTxn.GetVectorByName(SnapshotAttr_DBID).Get(i).(uint64)
		appendable := ins.GetVectorByName(pkgcatalog.BlockMeta_EntryState).Get(i).(bool)
		state := ES_NotAppendable
		if appendable {
			state = ES_Appendable
		}
		blkID := ins.GetVectorByName(pkgcatalog.BlockMeta_ID).Get(i).(types.Blockid)
		sid := blkID.Object()
		metaLoc := ins.GetVectorByName(pkgcatalog.BlockMeta_MetaLoc).Get(i).([]byte)
		deltaLoc := ins.GetVectorByName(pkgcatalog.BlockMeta_DeltaLoc).Get(i).([]byte)
		txnNode := txnbase.ReadTuple(insTxn, i)
		catalog.onReplayCreateBlock(dbid, tid, sid, &blkID, state, metaLoc, deltaLoc, txnNode, dataFactory)
	}
	for i := range del.Length() {
		tid := delTxn.GetVectorByName(SnapshotAttr_TID).Get(i).(uint64)
		if pkgcatalog.IsSystemTable(tid) != forSys {
			continue
		}
		dbid := delTxn.GetVectorByName(SnapshotAttr_DBID).Get(i).(uint64)
		rid := del.GetVectorByName(AttrRowID).Get(i).(types.Rowid)
		blkID := rid.BorrowBlockID()
		sid := rid.BorrowObjectID()
		un := txnbase.ReadTuple(delTxn, i)
		metaLoc := delTxn.GetVectorByName(pkgcatalog.BlockMeta_MetaLoc).Get(i).([]byte)
		deltaLoc := delTxn.GetVectorByName(pkgcatalog.BlockMeta_DeltaLoc).Get(i).([]byte)
		catalog.onReplayDeleteBlock(dbid, tid, sid, blkID, metaLoc, deltaLoc, un)
	}
}
func (catalog *Catalog) onReplayCreateBlock(
	dbid, tid uint64,
	objid *types.Objectid,
	blkid *types.Blockid,
	_ EntryState,
	_, deltaloc objectio.Location,
	txnNode *txnbase.TxnMVCCNode,
	dataFactory DataFactory) {
	// catalog.OnReplayBlockID(blkid)
	db, err := catalog.GetDatabaseByID(dbid)
	if err != nil {
		if moerr.IsMoErrCode(err, moerr.OkExpectedEOB) {
			return
		}
		panic(err)
	}
	rel, err := db.GetTableEntryByID(tid)
	if err != nil {
		if moerr.IsMoErrCode(err, moerr.OkExpectedEOB) {
			return
		}
		panic(err)
	}
	if !deltaloc.IsEmpty() {
		obj, err := rel.GetObjectByID(objid)
		if obj == nil {
			// checkpoint-end is t100, data collecting happens at t110
			// and a new aobject at t104, flush at t105, the deltaloc on the aobject will be collected.
			return
		}
		if err != nil {
			panic(err)
		}
		tombstone := rel.GetOrCreateTombstone(obj, dataFactory.MakeTombstoneFactory())
		_, blkOffset := blkid.Offsets()
		mvccNode := &MVCCNode[*MetadataMVCCNode]{
			EntryMVCCNode: &EntryMVCCNode{},
			TxnMVCCNode:   txnNode,
			BaseNode: &MetadataMVCCNode{
				DeltaLoc: deltaloc,
			},
		}
		tombstone.ReplayDeltaLoc(mvccNode, blkOffset)
	}
}

func (catalog *Catalog) onReplayDeleteBlock(
	dbid, tid uint64,
	objid *types.Objectid,
	blkid *types.Blockid,
	metaloc,
	deltaloc objectio.Location,
	txnNode *txnbase.TxnMVCCNode,
) {
	panic("logic error")
}
func (catalog *Catalog) ReplayTableRows() {
	rows := uint64(0)
	tableProcessor := new(LoopProcessor)
	tableProcessor.ObjectFn = func(be *ObjectEntry) error {
		if !be.IsActive() {
			return nil
		}
		rows += be.GetObjectData().GetRowsOnReplay()
		return nil
	}
	tableProcessor.TombstoneFn = func(t data.Tombstone) error {
		rows -= uint64(t.GetDeleteCnt())
		return nil
	}
	processor := new(LoopProcessor)
	processor.TableFn = func(tbl *TableEntry) error {
		if tbl.db.name == pkgcatalog.MO_CATALOG {
			return nil
		}
		rows = 0
		err := tbl.RecurLoop(tableProcessor)
		if err != nil {
			panic(err)
		}
		tbl.rows.Store(rows)
		return nil
	}
	err := catalog.RecurLoop(processor)
	if err != nil {
		panic(err)
	}
}

//#endregion
