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
	"fmt"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/clusterservice"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	moruntime "github.com/matrixorigin/matrixone/pkg/common/runtime"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/defines"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/lockservice"
	"github.com/matrixorigin/matrixone/pkg/logservice"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/pb/metadata"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	pb "github.com/matrixorigin/matrixone/pkg/pb/statsinfo"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	client2 "github.com/matrixorigin/matrixone/pkg/queryservice/client"
	"github.com/matrixorigin/matrixone/pkg/sql/colexec"
	"github.com/matrixorigin/matrixone/pkg/txn/client"
	"github.com/matrixorigin/matrixone/pkg/util/errutil"
	"github.com/matrixorigin/matrixone/pkg/util/executor"
	v2 "github.com/matrixorigin/matrixone/pkg/util/metric/v2"
	"github.com/matrixorigin/matrixone/pkg/util/stack"
	"github.com/matrixorigin/matrixone/pkg/version"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/disttae/cache"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/disttae/logtailreplay"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/disttae/route"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/blockio"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
	"github.com/panjf2000/ants/v2"
)

var _ engine.Engine = new(Engine)
var ncpu = runtime.GOMAXPROCS(0)

func New(
	ctx context.Context,
	mp *mpool.MPool,
	fs fileservice.FileService,
	cli client.TxnClient,
	hakeeper logservice.CNHAKeeperClient,
	keyRouter client2.KeyRouter[pb.StatsInfoKey],
	threshold int,
) *Engine {
	cluster := clusterservice.GetMOCluster()
	services := cluster.GetAllTNServices()

	var tnID string
	if len(services) > 0 {
		tnID = services[0].ServiceID
	}

	ls, ok := moruntime.ProcessLevelRuntime().GetGlobalVariables(moruntime.LockService)
	if !ok {
		logutil.Fatalf("missing lock service")
	}

	e := &Engine{
		mp:       mp,
		fs:       fs,
		ls:       ls.(lockservice.LockService),
		hakeeper: hakeeper,
		cli:      cli,
		idGen:    hakeeper,
		tnID:     tnID,
		packerPool: fileservice.NewPool(
			128,
			func() *types.Packer {
				return types.NewPacker(mp)
			},
			func(packer *types.Packer) {
				packer.Reset()
				packer.FreeMem()
			},
			func(packer *types.Packer) {
				packer.Reset()
				packer.FreeMem()
			},
		),
	}
	e.snapCatalog = &struct {
		sync.Mutex
		snaps []*cache.CatalogCache
	}{}
	e.mu.snapParts = make(map[[2]uint64]*struct {
		sync.Mutex
		snaps []*logtailreplay.Partition
	})

	pool, err := ants.NewPool(GCPoolSize)
	if err != nil {
		panic(err)
	}
	e.gcPool = pool

	e.globalStats = NewGlobalStats(ctx, e, keyRouter)

	e.messageCenter = &process.MessageCenter{
		StmtIDToBoard: make(map[uuid.UUID]*process.MessageBoard, 64),
		RwMutex:       &sync.Mutex{},
	}

	if err := e.init(ctx); err != nil {
		panic(err)
	}

	return e
}

func (e *Engine) Create(ctx context.Context, name string, op client.TxnOperator) error {
	if op.IsSnapOp() {
		return moerr.NewInternalErrorNoCtx("create database in snapshot txn")
	}
	txn, err := txnIsValid(op)
	if err != nil {
		return err
	}
	typ := getTyp(ctx)
	sql := getSql(ctx)
	accountId, userId, roleId, err := getAccessInfo(ctx)
	if err != nil {
		return err
	}
	databaseId, err := txn.allocateID(ctx)
	if err != nil {
		return err
	}

	var packer *types.Packer
	put := e.packerPool.Get(&packer)
	defer put.Put()
	bat, err := genCreateDatabaseTuple(sql, accountId, userId, roleId,
		name, databaseId, typ, txn.proc.Mp(), packer)
	if err != nil {
		return err
	}
	// non-io operations do not need to pass context
	note := noteForCreate(uint64(accountId), name)
	if _, err = txn.WriteBatch(INSERT, note, accountId, catalog.MO_CATALOG_ID, catalog.MO_DATABASE_ID,
		catalog.MO_CATALOG, catalog.MO_DATABASE, bat, txn.tnStores[0]); err != nil {
		bat.Clean(txn.proc.Mp())
		return err
	}

	key := genDatabaseKey(accountId, name)
	txn.databaseMap.Store(key, &txnDatabase{
		op:           op,
		databaseId:   databaseId,
		databaseName: name,
	})

	txn.deletedDatabaseMap.Delete(key)
	return nil
}

func execReadSql(ctx context.Context, op client.TxnOperator, sql string) (executor.Result, error) {
	// copy from compile.go runSqlWithResult
	v, ok := moruntime.ProcessLevelRuntime().GetGlobalVariables(moruntime.InternalSQLExecutor)
	if !ok {
		panic("missing lock service")
	}
	exec := v.(executor.SQLExecutor)
	proc := op.GetWorkspace().(*Transaction).proc
	opts := executor.Options{}.
		WithDisableIncrStatement().
		WithTxn(op).
		WithTimeZone(proc.SessionInfo.TimeZone)
	return exec.Exec(ctx, sql, opts)
}

func fillTsVecForSysTableQueryBatch(bat *batch.Batch, ts types.TS, m *mpool.MPool) error {
	tsvec := vector.NewVec(types.T_TS.ToType())
	for i := 0; i < bat.RowCount(); i++ {
		if err := vector.AppendFixed(tsvec, ts, false, m); err != nil {
			tsvec.Free(m)
			return err
		}
	}
	bat.Vecs = append([]*vector.Vector{bat.Vecs[0] /*rowid*/, tsvec}, bat.Vecs[1:]...)
	return nil
}

func isColumnsBatchPerfectlySplitted(bs []*batch.Batch) bool {
	tidIdx := cache.MO_OFF + catalog.MO_COLUMNS_ATT_RELNAME_ID_IDX
	if len(bs) == 1 {
		return true
	}
	prevTableId := vector.GetFixedAt[uint64](bs[0].Vecs[tidIdx], bs[0].RowCount()-1)
	for _, b := range bs[1:] {
		firstId := vector.GetFixedAt[uint64](b.Vecs[tidIdx], 0)
		if firstId == prevTableId {
			return false
		}
		prevTableId = vector.GetFixedAt[uint64](b.Vecs[tidIdx], b.RowCount()-1)
	}
	return true
}

func (e *Engine) foundDatabaseFromStorage(
	ctx context.Context,
	accountID uint32,
	name string, op client.TxnOperator) (*cache.DatabaseItem, error) {
	now := time.Now()
	defer func() {
		logutil.Infof("yyyy  foundDatabaseFromStorage %v-%v %v", accountID, name, time.Since(now))
	}()
	sql := fmt.Sprintf(catalog.MoDatabaseAllQueryFormat, accountID, name)
	res, err := execReadSql(ctx, op, sql)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	if len(res.Batches) != 1 { // not found
		return nil, nil
	}
	if row := res.Batches[0].RowCount(); row != 1 {
		panic(fmt.Sprintf("foundDatabaseFromStorage failed: table result row cnt: %v, sql : %s", row, sql))
	}
	bat := res.Batches[0]

	ts := types.TimestampToTS(op.SnapshotTS())
	if err := fillTsVecForSysTableQueryBatch(bat, ts, res.Mp); err != nil {
		return nil, err
	}
	var ret *cache.DatabaseItem
	e.catalog.ParseDatabaseBatchAnd(bat, func(di *cache.DatabaseItem) {
		ret = di
	})
	return ret, nil
}

func (e *Engine) Database(ctx context.Context, name string,
	op client.TxnOperator) (engine.Database, error) {
	logDebugf(op.Txn(), "Engine.Database %s", name)
	txn, err := txnIsValid(op)
	if err != nil {
		return nil, err
	}
	if name == catalog.MO_CATALOG {
		db := &txnDatabase{
			op:           op,
			databaseId:   catalog.MO_CATALOG_ID,
			databaseName: name,
		}
		return db, nil
	}
	accountId, err := defines.GetAccountId(ctx)
	if err != nil {
		return nil, err
	}

	// check the database is deleted or not
	key := genDatabaseKey(accountId, name)
	if _, exist := txn.deletedDatabaseMap.Load(key); exist {
		return nil, moerr.NewParseError(ctx, "database %q does not exist", name)
	}

	if v, ok := txn.databaseMap.Load(key); ok {
		return v.(*txnDatabase), nil
	}

	item := &cache.DatabaseItem{
		Name:      name,
		AccountId: accountId,
		Ts:        txn.op.SnapshotTS(),
	}

	var catalog *cache.CatalogCache
	if !txn.op.IsSnapOp() {
		catalog = e.getLatestCatalogCache()
	} else {
		catalog, err = e.getOrCreateSnapCatalogCache(
			ctx,
			types.TimestampToTS(txn.op.SnapshotTS()))
		if err != nil {
			return nil, err
		}
	}

	if ok := catalog.GetDatabase(item); !ok {
		if !catalog.CanServe(types.TimestampToTS(op.SnapshotTS())) {
			// read batch from storage
			if item, err = e.foundDatabaseFromStorage(ctx, accountId, name, op); err != nil {
				return nil, err
			}
			if item == nil {
				return nil, moerr.GetOkExpectedEOB()
			}
		} else {
			return nil, moerr.GetOkExpectedEOB()
		}
	}

	return &txnDatabase{
		op:                op,
		databaseName:      name,
		databaseId:        item.Id,
		databaseType:      item.Typ,
		databaseCreateSql: item.CreateSql,
	}, nil
}

func (e *Engine) Databases(ctx context.Context, op client.TxnOperator) ([]string, error) {
	var dbs []string

	txn, err := txnIsValid(op)
	if err != nil {
		return nil, err
	}
	accountId, err := defines.GetAccountId(ctx)
	if err != nil {
		return nil, err
	}

	//first get all delete tables
	deleteDatabases := make(map[string]any)
	txn.deletedDatabaseMap.Range(func(k, _ any) bool {
		key := k.(databaseKey)
		if key.accountId == accountId {
			deleteDatabases[key.name] = nil
		}
		return true
	})

	txn.databaseMap.Range(func(k, _ any) bool {
		key := k.(databaseKey)
		if key.accountId == accountId {
			// if the database is deleted, do not save it.
			if _, exist := deleteDatabases[key.name]; !exist {
				dbs = append(dbs, key.name)
			}
		}
		return true
	})

	var catalog *cache.CatalogCache
	if !txn.op.IsSnapOp() {
		catalog = e.getLatestCatalogCache()
	} else {
		catalog, err = e.getOrCreateSnapCatalogCache(
			ctx,
			types.TimestampToTS(txn.op.SnapshotTS()))
		if err != nil {
			return nil, err
		}
	}
	dbsInCatalog := catalog.Databases(accountId, txn.op.SnapshotTS())
	dbsExceptDelete := removeIf[string](dbsInCatalog, func(t string) bool {
		return find[string](deleteDatabases, t)
	})
	dbs = append(dbs, dbsExceptDelete...)
	return dbs, nil
}

func (e *Engine) GetNameById(ctx context.Context, op client.TxnOperator, tableId uint64) (dbName string, tblName string, err error) {
	dbName, tblName, _, err = e.GetRelationById(ctx, op, tableId)
	return
}

func (e *Engine) GetRelationById(ctx context.Context, op client.TxnOperator, tableId uint64) (dbName, tableName string, rel engine.Relation, err error) {
	switch tableId {
	case catalog.MO_DATABASE_ID:
		db := &txnDatabase{
			op:           op,
			databaseId:   catalog.MO_CATALOG_ID,
			databaseName: catalog.MO_CATALOG,
		}
		defs := catalog.MoDatabaseTableDefs
		return catalog.MO_CATALOG, catalog.MO_DATABASE,
			db.openSysTable(nil, tableId, catalog.MO_DATABASE, defs), nil
	case catalog.MO_TABLES_ID:
		db := &txnDatabase{
			op:           op,
			databaseId:   catalog.MO_CATALOG_ID,
			databaseName: catalog.MO_CATALOG,
		}
		defs := catalog.MoTablesTableDefs
		return catalog.MO_CATALOG, catalog.MO_TABLES,
			db.openSysTable(nil, tableId, catalog.MO_TABLES, defs), nil
	case catalog.MO_COLUMNS_ID:
		db := &txnDatabase{
			op:           op,
			databaseId:   catalog.MO_CATALOG_ID,
			databaseName: catalog.MO_CATALOG,
		}
		defs := catalog.MoColumnsTableDefs
		return catalog.MO_CATALOG, catalog.MO_COLUMNS,
			db.openSysTable(nil, tableId, catalog.MO_COLUMNS, defs), nil
	}

	noRepCtx := errutil.ContextWithNoReport(ctx, true)
	dbs, err := e.Databases(ctx, op)
	if err != nil {
		return "", "", nil, err
	}
	fn := func(tryDB string) error {
		var db engine.Database
		db, err = e.Database(noRepCtx, tryDB, op)
		if err != nil {
			return err
		}
		distDb := db.(*txnDatabase)
		tableName, rel, err = distDb.getRelationById(noRepCtx, tableId)
		if err != nil {
			return err
		}
		if rel != nil {
			dbName = tryDB
		}
		return nil
	}
	for _, dbname := range dbs {
		if err := fn(dbname); err != nil {
			return "", "", nil, err
		}
		if rel != nil {
			break
		}
	}
	// everyone is able to see MO_CATALOG
	if rel == nil && !slices.Contains(dbs, catalog.MO_CATALOG) {
		if err := fn(catalog.MO_CATALOG); err != nil {
			return "", "", nil, err
		}
	}
	if rel == nil {
		accountId, _ := defines.GetAccountId(ctx)
		return "", "", nil, moerr.NewInternalError(ctx, "can not find table by id %d: accountId: %v ", tableId, accountId)
	}
	return
}

func (e *Engine) AllocateIDByKey(ctx context.Context, key string) (uint64, error) {
	return e.idGen.AllocateIDByKey(ctx, key)
}

func (e *Engine) Delete(ctx context.Context, name string, op client.TxnOperator) (err error) {
	defer func() {
		if err != nil {
			if strings.Contains(name, "sysbench_db") {
				logutil.Errorf("delete database %s failed: %v", name, err)
				logutil.Errorf("stack: %s", stack.Callers(3))
				logutil.Errorf("txnmeta %v", op.Txn().DebugString())
			}
		}
	}()

	var databaseId uint64
	var rowId types.Rowid
	var txn *Transaction
	//var db *txnDatabase
	if op.IsSnapOp() {
		return moerr.NewInternalErrorNoCtx("delete database in snapshot txn")
	}

	txn, err = txnIsValid(op)
	if err != nil {
		return err
	}

	accountId, err := defines.GetAccountId(ctx)
	if err != nil {
		return err
	}

	key := genDatabaseKey(accountId, name)
	if val, ok := txn.databaseMap.Load(key); ok {
		txn.databaseMap.Delete(key)
		database := val.(*txnDatabase)
		databaseId = database.databaseId
	} else {
		item := &cache.DatabaseItem{
			Name:      name,
			AccountId: accountId,
			Ts:        txn.op.SnapshotTS(),
		}
		if ok = e.getLatestCatalogCache().GetDatabase(item); !ok {
			return moerr.GetOkExpectedEOB()
		}
		databaseId = item.Id
	}

	dbNew := &txnDatabase{
		op:           op,
		databaseName: name,
		databaseId:   databaseId,
	}

	rels, err := dbNew.Relations(ctx)
	if err != nil {
		return err
	}
	for _, relName := range rels {
		if err := dbNew.Delete(ctx, relName); err != nil {
			return err
		}
	}

	res, err := execReadSql(ctx, op, fmt.Sprintf(catalog.MoDatabaseRowidQueryFormat, accountId, name))
	if err != nil {
		return err
	}
	if len(res.Batches) != 1 || res.Batches[0].Vecs[0].Length() != 1 {
		panic("delete table failed: query failed")
	}
	rowId = vector.GetFixedAt[types.Rowid](res.Batches[0].Vecs[0], 0)

	var packer *types.Packer
	put := e.packerPool.Get(&packer)
	defer put.Put()
	bat, err := genDropDatabaseTuple(rowId, accountId, databaseId, name, txn.proc.Mp(), packer)
	if err != nil {
		return err
	}
	dbNew.getTxn().deletedDatabaseMap.Store(key, databaseId)
	bat = txn.deleteBatch(bat, catalog.MO_CATALOG_ID, catalog.MO_DATABASE_ID)
	if bat.RowCount() == 0 {
		return nil
	}
	note := noteForDrop(uint64(accountId), name)
	if _, err := txn.WriteBatch(DELETE, note, accountId, catalog.MO_CATALOG_ID, catalog.MO_DATABASE_ID,
		catalog.MO_CATALOG, catalog.MO_DATABASE, bat, txn.tnStores[0]); err != nil {
		bat.Clean(txn.proc.Mp())
		return err
	}

	return nil
}

func (e *Engine) New(ctx context.Context, op client.TxnOperator) error {
	logDebugf(op.Txn(), "Engine.New")
	proc := process.New(
		ctx,
		e.mp,
		e.cli,
		op,
		e.fs,
		e.ls,
		e.qc,
		e.hakeeper,
		e.us,
		nil,
	)

	id := objectio.NewSegmentid()
	bytes := types.EncodeUuid(id)
	txn := &Transaction{
		op:     op,
		proc:   proc,
		engine: e,
		//meta:     op.TxnRef(),
		idGen:    e.idGen,
		tnStores: e.getTNServices(),
		tableCache: struct {
			cachedIndex int
			tableMap    *sync.Map
		}{
			tableMap: new(sync.Map),
		},
		databaseMap:        new(sync.Map),
		deletedDatabaseMap: new(sync.Map),
		createMap:          new(sync.Map),
		deletedTableMap:    new(sync.Map),
		rowId: [6]uint32{
			types.DecodeUint32(bytes[0:4]),
			types.DecodeUint32(bytes[4:8]),
			types.DecodeUint32(bytes[8:12]),
			types.DecodeUint32(bytes[12:16]),
			0,
			0,
		},
		segId: *id,
		deletedBlocks: &deletedBlocks{
			offsets: map[types.Blockid][]int64{},
		},
		cnBlkId_Pos:          map[types.Blockid]Pos{},
		batchSelectList:      make(map[*batch.Batch][]int64),
		toFreeBatches:        make(map[tableKey][]*batch.Batch),
		syncCommittedTSCount: e.cli.GetSyncLatestCommitTSTimes(),
	}

	txn.blockId_tn_delete_metaLoc_batch = struct {
		sync.RWMutex
		data map[types.Blockid][]*batch.Batch
	}{data: make(map[types.Blockid][]*batch.Batch)}

	txn.readOnly.Store(true)
	// transaction's local segment for raw batch.
	colexec.Get().PutCnSegment(id, colexec.TxnWorkSpaceIdType)
	op.AddWorkspace(txn)

	e.pClient.validLogTailMustApplied(txn.op.SnapshotTS())
	return nil
}

func (e *Engine) Nodes(
	isInternal bool, tenant string, username string, cnLabel map[string]string,
) (engine.Nodes, error) {
	var nodes engine.Nodes

	start := time.Now()
	defer func() {
		v2.TxnStatementNodesHistogram.Observe(time.Since(start).Seconds())
	}()

	cluster := clusterservice.GetMOCluster()
	var selector clusterservice.Selector

	// If the requested labels are empty, return all CN servers.
	if len(cnLabel) == 0 {
		cluster.GetCNService(selector, func(c metadata.CNService) bool {
			if c.CommitID == version.CommitID {
				nodes = append(nodes, engine.Node{
					Mcpu: ncpu,
					Id:   c.ServiceID,
					Addr: c.PipelineServiceAddress,
				})
			}
			return true
		})
		return nodes, nil
	}

	selector = clusterservice.NewSelector().SelectByLabel(cnLabel, clusterservice.EQ_Globbing)
	if isInternal || strings.ToLower(tenant) == "sys" {
		route.RouteForSuperTenant(selector, username, nil, func(s *metadata.CNService) {
			if s.CommitID == version.CommitID {
				nodes = append(nodes, engine.Node{
					Mcpu: ncpu,
					Id:   s.ServiceID,
					Addr: s.PipelineServiceAddress,
				})
			}
		})
	} else {
		route.RouteForCommonTenant(selector, nil, func(s *metadata.CNService) {
			if s.CommitID == version.CommitID {
				nodes = append(nodes, engine.Node{
					Mcpu: ncpu,
					Id:   s.ServiceID,
					Addr: s.PipelineServiceAddress,
				})
			}
		})
	}
	return nodes, nil
}

func (e *Engine) Hints() (h engine.Hints) {
	h.CommitOrRollbackTimeout = time.Minute * 5
	return
}

func (e *Engine) NewBlockReader(ctx context.Context, num int, ts timestamp.Timestamp,
	expr *plan.Expr, filter any, ranges []byte, tblDef *plan.TableDef, proc any) ([]engine.Reader, error) {
	var blockReadPKFilter blockio.BlockReadFilter
	if filter == nil {
		// remote block reader
		basePKFilter := constructBasePKFilter(expr, tblDef, proc.(*process.Process))
		blockReadPKFilter = constructBlockReadPKFilter(tblDef.Pkey.PkeyColName, basePKFilter)
		//fmt.Println("remote filter: ", basePKFilter.String(), blockReadPKFilter)
	}

	blkSlice := objectio.BlockInfoSlice(ranges)
	rds := make([]engine.Reader, num)
	blkInfos := make([]*objectio.BlockInfo, 0, blkSlice.Len())
	for i := 0; i < blkSlice.Len(); i++ {
		blkInfos = append(blkInfos, blkSlice.Get(i))
	}
	if len(blkInfos) < num || len(blkInfos) == 1 {
		for i, blk := range blkInfos {
			//FIXME::why set blk.EntryState = false ?
			blk.EntryState = false
			rds[i] = newBlockReader(
				ctx, tblDef, ts, []*objectio.BlockInfo{blk}, expr, blockReadPKFilter, e.fs, proc.(*process.Process),
			)
		}
		for j := len(blkInfos); j < num; j++ {
			rds[j] = &emptyReader{}
		}
		return rds, nil
	}

	infos, steps := groupBlocksToObjects(blkInfos, num)
	fs, err := fileservice.Get[fileservice.FileService](e.fs, defines.SharedFileServiceName)
	if err != nil {
		return nil, err
	}
	blockReaders := newBlockReaders(ctx, fs, tblDef, ts, num, expr, blockReadPKFilter, proc.(*process.Process))
	distributeBlocksToBlockReaders(blockReaders, num, len(blkInfos), infos, steps)
	for i := 0; i < num; i++ {
		rds[i] = blockReaders[i]
	}
	return rds, nil
}

func (e *Engine) getTNServices() []DNStore {
	cluster := clusterservice.GetMOCluster()
	return cluster.GetAllTNServices()
}

func (e *Engine) setPushClientStatus(ready bool) {
	e.Lock()
	defer e.Unlock()

	if ready {
		e.cli.Resume()
	} else {
		e.cli.Pause()
	}

	e.pClient.receivedLogTailTime.ready.Store(ready)
	if e.pClient.subscriber != nil {
		if ready {
			e.pClient.subscriber.setReady()
		} else {
			e.pClient.subscriber.setNotReady()
		}
	}
}

func (e *Engine) abortAllRunningTxn() {
	e.Lock()
	defer e.Unlock()
	e.cli.AbortAllRunningTxn()
}

func (e *Engine) cleanMemoryTableWithTable(dbId, tblId uint64) {
	e.Lock()
	defer e.Unlock()
	// XXX it's probably not a good way to do that.
	// after we set it to empty, actually this part of memory was not immediately released.
	// maybe a very old transaction still using that.
	delete(e.partitions, [2]uint64{dbId, tblId})
	logutil.Debugf("clean memory table of tbl[dbId: %d, tblId: %d]", dbId, tblId)
}

func (e *Engine) PushClient() *PushClient {
	return &e.pClient
}

// TryToSubscribeTable implements the LogtailEngine interface.
func (e *Engine) TryToSubscribeTable(ctx context.Context, dbID, tbID uint64) error {
	return e.PushClient().TryToSubscribeTable(ctx, dbID, tbID)
}

// UnsubscribeTable implements the LogtailEngine interface.
func (e *Engine) UnsubscribeTable(ctx context.Context, dbID, tbID uint64) error {
	return e.PushClient().UnsubscribeTable(ctx, dbID, tbID)
}

func (e *Engine) Stats(ctx context.Context, key pb.StatsInfoKey, sync bool) *pb.StatsInfo {
	return e.globalStats.Get(ctx, key, sync)
}

func (e *Engine) GetMessageCenter() any {
	return e.messageCenter
}
