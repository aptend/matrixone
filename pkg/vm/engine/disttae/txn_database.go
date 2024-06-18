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
	"runtime/debug"
	"strconv"
	"strings"

	"go.uber.org/zap"

	"github.com/matrixorigin/matrixone/pkg/logutil"
	txn2 "github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/client"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/defines"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/disttae/cache"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

var _ engine.Database = new(txnDatabase)

func (db *txnDatabase) getTxn() *Transaction {
	return db.op.GetWorkspace().(*Transaction)
}

func (db *txnDatabase) getEng() *Engine {
	return db.op.GetWorkspace().(*Transaction).engine
}

func (db *txnDatabase) Relations(ctx context.Context) ([]string, error) {
	var rels []string
	//first get all delete tables
	seenTables := make(map[string]any)
	db.getTxn().deletedTableMap.Range(func(k, _ any) bool {
		key := k.(tableKey)
		if key.databaseId == db.databaseId {
			seenTables[key.name] = nil
		}
		return true
	})
	db.getTxn().createMap.Range(func(k, _ any) bool {
		key := k.(tableKey)
		if key.databaseId == db.databaseId {
			//if the table is deleted, do not save it.
			if _, exist := seenTables[key.name]; !exist {
				rels = append(rels, key.name)
				seenTables[key.name] = nil
			}
		}
		return true
	})
	accountId, err := defines.GetAccountId(ctx)
	if err != nil {
		return nil, err
	}
	var catache *cache.CatalogCache
	if !db.op.IsSnapOp() {
		catache = db.getTxn().engine.getLatestCatalogCache()
	} else {
		catache, err = db.getTxn().engine.getOrCreateSnapCatalogCache(
			ctx,
			types.TimestampToTS(db.op.SnapshotTS()))
		if err != nil {
			return nil, err
		}
	}
	tbls, _ := catache.Tables(accountId, db.databaseId, db.op.SnapshotTS())
	for _, tbl := range tbls {
		//if the table is deleted or modified in txn, skip
		if _, exist := seenTables[tbl]; !exist {
			rels = append(rels, tbl)
		}
	}
	return rels, nil
}

func (db *txnDatabase) getTableNameById(ctx context.Context, id uint64) (string, error) {
	tblName := ""
	//first check the tableID is deleted or not
	deleted := false
	db.getTxn().deletedTableMap.Range(func(k, v any) bool {
		key := k.(tableKey)
		val := v.(uint64)
		if key.databaseId == db.databaseId && val == id {
			deleted = true
			return false
		}
		return true
	})
	if deleted {
		return "", nil
	}
	db.getTxn().createMap.Range(func(k, v any) bool {
		key := k.(tableKey)
		val := v.(*txnTable)
		if key.databaseId == db.databaseId && val.tableId == id {
			tblName = key.name
			return false
		}
		return true
	})

	if tblName == "" {
		accountId, err := defines.GetAccountId(ctx)
		if err != nil {
			return "", err
		}
		var catache *cache.CatalogCache
		if !db.op.IsSnapOp() {
			catache = db.getTxn().engine.getLatestCatalogCache()
		} else {
			catache, err = db.getTxn().engine.getOrCreateSnapCatalogCache(
				ctx,
				types.TimestampToTS(db.op.SnapshotTS()))
			if err != nil {
				return "", err
			}
		}
		tbls, tblIds := catache.Tables(
			accountId, db.databaseId, db.op.SnapshotTS())
		for idx, tblId := range tblIds {
			if tblId == id {
				tblName = tbls[idx]
				break
			}
		}
	}
	return tblName, nil
}

func (db *txnDatabase) getRelationById(ctx context.Context, id uint64) (string, engine.Relation, error) {
	tblName, err := db.getTableNameById(ctx, id)
	if err != nil {
		return "", nil, err
	}
	if tblName == "" {
		return "", nil, nil
	}
	rel, _ := db.Relation(ctx, tblName, nil)
	return tblName, rel, nil
}

func (db *txnDatabase) Relation(ctx context.Context, name string, proc any) (engine.Relation, error) {
	logDebugf(db.op.Txn(), "txnDatabase.Relation table %s", name)
	txn := db.getTxn()
	if txn.op.Status() == txn2.TxnStatus_Aborted {
		return nil, moerr.NewTxnClosedNoCtx(txn.op.Txn().ID)
	}
	accountId, err := defines.GetAccountId(ctx)
	if err != nil {
		return nil, err
	}
	key := genTableKey(accountId, name, db.databaseId)
	// check the table is deleted or not
	if _, exist := db.getTxn().deletedTableMap.Load(key); exist {
		return nil, moerr.NewParseError(ctx, "table %q does not exist", name)
	}

	p := db.getTxn().proc
	if proc != nil {
		p = proc.(*process.Process)
	}

	rel := db.getTxn().getCachedTable(ctx, key)
	if rel != nil {
		rel.proc.Store(p)
		return rel, nil
	}

	// get relation from the txn created tables cache: created by this txn
	if v, ok := db.getTxn().createMap.Load(key); ok {
		v.(*txnTable).proc.Store(p)
		return v.(*txnTable), nil
	}

	// special tables
	if db.databaseName == catalog.MO_CATALOG {
		switch name {
		case catalog.MO_DATABASE:
			id := uint64(catalog.MO_DATABASE_ID)
			defs := catalog.MoDatabaseTableDefs
			return db.openSysTable(p, id, name, defs), nil
		case catalog.MO_TABLES:
			id := uint64(catalog.MO_TABLES_ID)
			defs := catalog.MoTablesTableDefs
			return db.openSysTable(p, id, name, defs), nil
		case catalog.MO_COLUMNS:
			id := uint64(catalog.MO_COLUMNS_ID)
			defs := catalog.MoColumnsTableDefs
			return db.openSysTable(p, id, name, defs), nil
		}
	}

	item, err := db.getTableItem(
		ctx,
		accountId,
		name,
		db.getTxn().engine,
	)
	if err != nil {
		return nil, err
	}

	tbl := newTxnTableWithItem(
		db,
		item,
		p,
	)

	db.getTxn().tableCache.tableMap.Store(key, tbl)
	return tbl, nil
}

func (db *txnDatabase) Delete(ctx context.Context, name string) error {
	_, err := db.deleteTable(ctx, name, false, false)
	return err
}

func (db *txnDatabase) deleteTable(ctx context.Context, name string, forAlter bool, useAlterNote bool) ([]engine.TableDef, error) {
	// useAlterNote means that the no table is really deleted, just for alter table
	var id uint64
	var rowid types.Rowid
	var rowids []types.Rowid
	var colPKs []string
	var defs []engine.TableDef
	if db.op.IsSnapOp() {
		return nil, moerr.NewInternalErrorNoCtx("delete table in snapshot transaction")
	}
	accountId, err := defines.GetAccountId(ctx)
	if err != nil {
		return nil, err
	}
	txn := db.getTxn()
	k := genTableKey(accountId, name, db.databaseId)
	if v, ok := txn.createMap.Load(k); ok {
		txn.createMap.Delete(k)
		table := v.(*txnTable)
		id = table.tableId
		rowid = table.rowid
		rowids = table.rowids
		defs = table.defs
		colPKs = getColPks(id, table.tableDef.Cols)
	} else if v, ok := txn.tableCache.tableMap.Load(k); ok {
		table := v.(*txnTable)
		id = table.tableId
		txn.tableCache.tableMap.Delete(k)
		rowid = table.rowid
		rowids = table.rowids
		defs = table.defs
		colPKs = getColPks(id, table.tableDef.Cols)
	} else {
		item := &cache.TableItem{
			Name:       name,
			DatabaseId: db.databaseId,
			AccountId:  accountId,
			Ts:         db.op.SnapshotTS(),
		}
		if ok := txn.engine.getLatestCatalogCache().GetTable(item); !ok {
			return nil, moerr.GetOkExpectedEOB()
		}
		id = item.Id
		rowid = item.Rowid
		rowids = item.Rowids
		defs = item.Defs
		colPKs = getColPks(id, item.TableDef.Cols)
	}
	if len(rowids) != len(colPKs) {
		return nil, moerr.NewInternalErrorNoCtx("delete table failed %v, %v", len(rowids), len(colPKs))
	}

	{ // delete the row from mo_tables

		var packer *types.Packer
		put := txn.engine.packerPool.Get(&packer)
		defer put.Put()
		bat, err := genDropTableTuple(rowid, accountId, id, db.databaseId,
			name, db.databaseName, txn.proc.Mp(), packer)
		if err != nil {
			return nil, err
		}
		if bat = txn.deleteBatch(bat, catalog.MO_CATALOG_ID, catalog.MO_TABLES_ID); bat.RowCount() > 0 {
			// the deleted table is not created by this txn
			note := noteForDrop(id, name)
			if useAlterNote {
				note = noteForAlterDel(id, name)
			}
			if _, err := txn.WriteBatch(
				DELETE, note, accountId, catalog.MO_CATALOG_ID, catalog.MO_TABLES_ID,
				catalog.MO_CATALOG, catalog.MO_TABLES, bat, txn.tnStores[0]); err != nil {
				bat.Clean(txn.proc.Mp())
				return nil, err
			}
		} else if !forAlter {
			// An insert batch for mo_tables is cancelled, the dml on this table should be eliminated?
			// The answer for forAlter as true is NO, because later a table with the same tableId will be created.
			// The answer for forAlter as false is YES, because the table is really deleted, which is triggered by delete & truncate
			txn.tablesInVain = append(txn.tablesInVain, id)
		}
	}

	{ // delete rows from mo_columns
		bat, err := genDropColumnTuples(rowids, colPKs, txn.proc.Mp())
		if err != nil {
			return nil, err
		}

		if bat = txn.deleteBatch(bat, catalog.MO_CATALOG_ID, catalog.MO_COLUMNS_ID); bat.RowCount() > 0 {
			note := noteForDrop(id, name)
			if useAlterNote {
				note = noteForAlterDel(id, name)
			}
			if _, err = txn.WriteBatch(
				DELETE, note, accountId, catalog.MO_CATALOG_ID, catalog.MO_COLUMNS_ID,
				catalog.MO_CATALOG, catalog.MO_COLUMNS, bat, txn.tnStores[0]); err != nil {
				bat.Clean(txn.proc.Mp())
				return nil, err
			}
		}
	}
	txn.deletedTableMap.Store(k, id)
	return defs, nil
}

func (db *txnDatabase) Truncate(ctx context.Context, name string) (uint64, error) {
	if db.op.IsSnapOp() {
		return 0, moerr.NewInternalErrorNoCtx("truncate table in snapshot transaction")
	}
	newId, err := db.getTxn().allocateID(ctx)
	if err != nil {
		return 0, err
	}

	defs, err := db.deleteTable(ctx, name, false, false)
	if err != nil {
		return 0, err
	}

	if err := db.createWithID(ctx, name, newId, defs, false); err != nil {
		return 0, err
	}

	return newId, nil
}

func (db *txnDatabase) GetDatabaseId(ctx context.Context) string {
	return strconv.FormatUint(db.databaseId, 10)
}

func (db *txnDatabase) GetCreateSql(ctx context.Context) string {
	return db.databaseCreateSql
}

func (db *txnDatabase) IsSubscription(ctx context.Context) bool {
	return db.databaseType == catalog.SystemDBTypeSubscription
}

func (db *txnDatabase) Create(ctx context.Context, name string, defs []engine.TableDef) error {
	if db.op.IsSnapOp() {
		return moerr.NewInternalErrorNoCtx("create table in snapshot transaction")
	}
	tableId, err := db.getTxn().allocateID(ctx)
	if err != nil {
		return err
	}
	return db.createWithID(ctx, name, tableId, defs, false)
}

func (db *txnDatabase) createWithID(
	ctx context.Context,
	name string, tableId uint64, defs []engine.TableDef, useAlterNote bool,
) error {
	if db.op.IsSnapOp() {
		return moerr.NewInternalErrorNoCtx("create table in snapshot transaction")
	}
	accountId, userId, roleId, err := getAccessInfo(ctx)
	if err != nil {
		return err
	}
	m := db.getTxn().proc.Mp()

	// 1. inspect and **modify** defs, and construct columns
	cols, err := genColumnsFromDefs(accountId, name, db.databaseName, tableId, db.databaseId, defs)
	if err != nil {
		return err
	}
	tbl := new(txnTable)

	{ // prepare table information
		// 2.1 prepare basic table information
		tbl.db = db
		tbl.tableName = name
		tbl.tableId = tableId
		tbl.accountId = accountId
		tbl.rowid = types.DecodeFixed[types.Rowid](types.EncodeSlice([]uint64{tableId}))
		for _, def := range defs {
			switch defVal := def.(type) {
			case *engine.PropertiesDef:
				for _, property := range defVal.Properties {
					switch strings.ToLower(property.Key) {
					case catalog.SystemRelAttr_Comment:
						tbl.comment = property.Value
					case catalog.SystemRelAttr_Kind:
						tbl.relKind = property.Value
					case catalog.SystemRelAttr_CreateSQL:
						tbl.createSql = property.Value
					default:
					}
				}
			case *engine.ViewDef:
				tbl.viewdef = defVal.View
			case *engine.CommentDef:
				tbl.comment = defVal.Comment
			case *engine.PartitionDef:
				tbl.partitioned = defVal.Partitioned
				tbl.partition = defVal.Partition
			case *engine.ConstraintDef:
				tbl.constraint, err = defVal.MarshalBinary()
				if err != nil {
					return err
				}
			}
		}
		// 2.2 prepare columns related information
		tbl.primaryIdx = -1
		tbl.primarySeqnum = -1
		tbl.clusterByIdx = -1
		for i, col := range cols {
			if col.constraintType == catalog.SystemColPKConstraint {
				tbl.primaryIdx = i
				tbl.primarySeqnum = i
			}
			if col.isClusterBy == 1 {
				tbl.clusterByIdx = i
			}
		}

		// 2.3 prepare holistic table def
		tbl.defs = defs
		tbl.GetTableDef(ctx) // generate tbl.tableDef
	}

	{ // 3. Write create table batch, update tbl.rowiod

		db := tbl.db
		var packer *types.Packer
		put := db.getEng().packerPool.Get(&packer)
		m := db.getTxn().proc.Mp()
		defer put.Put()
		bat, err := genCreateTableTuple(
			tbl, accountId, userId, roleId,
			tbl.tableName, tbl.tableId, db.databaseId, db.databaseName, m, packer)
		if err != nil {
			return err
		}
		note := noteForCreate(tbl.tableId, tbl.tableName)
		if useAlterNote {
			note = noteForAlterIns(tbl.tableId, tbl.tableName)
		}
		rowidVec, err := db.getTxn().WriteBatch(INSERT, note, accountId, catalog.MO_CATALOG_ID, catalog.MO_TABLES_ID,
			catalog.MO_CATALOG, catalog.MO_TABLES, bat, db.getTxn().tnStores[0])
		if err != nil {
			bat.Clean(m)
			return err
		}
		tbl.rowid = vector.GetFixedAt[types.Rowid](rowidVec, 0)
	}

	{ // 4. Write create column batch
		bat, err := genCreateColumnTuples(cols, m)
		if err != nil {
			return err
		}
		note := noteForCreate(tbl.tableId, tbl.tableName)
		if useAlterNote {
			note = noteForAlterIns(tbl.tableId, tbl.tableName)
		}
		rowidVec, err := db.getTxn().WriteBatch(
			INSERT, note, 0, catalog.MO_CATALOG_ID, catalog.MO_COLUMNS_ID,
			catalog.MO_CATALOG, catalog.MO_COLUMNS, bat, db.getTxn().tnStores[0])
		if err != nil {
			bat.Clean(m)
			return err
		}
		for i := 0; i < rowidVec.Length(); i++ {
			tbl.rowids = append(tbl.rowids, vector.GetFixedAt[types.Rowid](rowidVec, i))
		}
	}

	// 5. handle map cache
	key := genTableKey(accountId, name, db.databaseId)
	db.getTxn().addCreateTable(key, tbl)
	//CORNER CASE
	//begin;
	//create table t1(a int);
	//drop table t1; //t1 is in deleteTableMap now.
	//select * from t1; //t1 does not exist.
	//create table t1(a int); //t1 does not exist. t1 can be created again.
	//	t1 needs be deleted from deleteTableMap
	db.getTxn().deletedTableMap.Delete(key)
	return nil
}

func (db *txnDatabase) openSysTable(p *process.Process, id uint64, name string,
	defs []engine.TableDef) engine.Relation {
	item := db.getEng().getLatestCatalogCache().GetTableById(db.databaseId, id)
	tbl := &txnTable{
		//AccountID for mo_tables, mo_database, mo_columns is always 0.
		accountId:     0,
		db:            db,
		tableId:       id,
		tableName:     name,
		defs:          defs,
		primaryIdx:    item.PrimaryIdx,
		primarySeqnum: item.PrimarySeqnum,
		clusterByIdx:  -1,
	}
	switch name {
	case catalog.MO_DATABASE:
		tbl.constraint = catalog.MoDatabaseConstraint
	case catalog.MO_TABLES:
		tbl.constraint = catalog.MoTableConstraint
	case catalog.MO_COLUMNS:
		tbl.constraint = catalog.MoColumnConstraint
	}
	tbl.GetTableDef(context.TODO())
	tbl.proc.Store(p)
	return tbl
}

func (db *txnDatabase) getTableItem(
	ctx context.Context,
	accountID uint32,
	name string,
	engine *Engine,
) (cache.TableItem, error) {
	item := cache.TableItem{
		Name:       name,
		DatabaseId: db.databaseId,
		AccountId:  accountID,
		Ts:         db.op.SnapshotTS(),
	}

	c, err := getCatalogCache(
		ctx,
		engine,
		db.op,
	)
	if err != nil {
		return cache.TableItem{}, err
	}

	if ok := c.GetTable(&item); !ok {
		logutil.Debugf("txnDatabase.Relation table %q(acc %d db %d) does not exist",
			name,
			accountID,
			db.databaseId)
		if strings.Contains(name, "_copy_") {
			stackInfo := debug.Stack()
			logutil.Error(moerr.NewParseError(context.Background(), "table %q does not exists", name).Error(), zap.String("Stack Trace", string(stackInfo)))
		}
		return cache.TableItem{}, moerr.NewParseError(ctx, "table %q does not exist", name)
	}
	return item, nil
}

func getCatalogCache(
	ctx context.Context,
	engine *Engine,
	op client.TxnOperator,
) (*cache.CatalogCache, error) {
	var cache *cache.CatalogCache
	var err error
	if !op.IsSnapOp() {
		cache = engine.getLatestCatalogCache()
	} else {
		cache, err = engine.getOrCreateSnapCatalogCache(
			ctx,
			types.TimestampToTS(op.SnapshotTS()))
		if err != nil {
			return nil, err
		}
	}
	return cache, nil
}
