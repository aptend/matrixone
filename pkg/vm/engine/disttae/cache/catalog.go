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

package cache

import (
	"encoding/hex"
	"sort"
	"sync"

	plan2 "github.com/matrixorigin/matrixone/pkg/sql/plan"
	"github.com/matrixorigin/matrixone/pkg/sql/util"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/compress"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/tidwall/btree"
)

func NewCatalog() *CatalogCache {
	return &CatalogCache{
		tables: &tableCache{
			data:       btree.NewBTreeG(tableItemLess),
			cpkeyIndex: btree.NewBTreeG(tableItemCPKeyLess),
		},
		databases: &databaseCache{
			data:       btree.NewBTreeG(databaseItemLess),
			cpkeyIndex: btree.NewBTreeG(databaseItemCPKeyLess),
		},
		mu: struct {
			sync.Mutex
			start types.TS
			end   types.TS
		}{start: types.MaxTs()},
	}
}

func (cc *CatalogCache) UpdateDuration(start types.TS, end types.TS) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.mu.start = start
	cc.mu.end = end
}

var _ = (&CatalogCache{}).UpdateStart

func (cc *CatalogCache) UpdateStart(ts types.TS) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.mu.start != types.MaxTs() {
		cc.mu.start = ts
	}
}

func (cc *CatalogCache) CanServe(ts types.TS) bool {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return ts.GreaterEq(&cc.mu.start) && ts.LessEq(&cc.mu.end)
}

func (cc *CatalogCache) GC(ts timestamp.Timestamp) {
	{ // table cache gc
		var items []*TableItem

		cc.tables.data.Scan(func(item *TableItem) bool {
			if len(items) > GcBuffer {
				return false
			}
			if item.Ts.Less(ts) {
				items = append(items, item)
			}
			return true
		})
		for _, item := range items {
			cc.tables.data.Delete(item)
			if !item.deleted {
				cc.tables.cpkeyIndex.Delete(item)
			}
		}
	}
	{ // database cache gc
		var items []*DatabaseItem

		cc.databases.data.Scan(func(item *DatabaseItem) bool {
			if len(items) > GcBuffer {
				return false
			}
			if item.Ts.Less(ts) {
				items = append(items, item)
			}
			return true
		})
		for _, item := range items {
			cc.databases.data.Delete(item)
			if !item.deleted {
				cc.databases.cpkeyIndex.Delete(item)
			}
		}
	}
}

func (cc *CatalogCache) Databases(accountId uint32, ts timestamp.Timestamp) []string {
	var rs []string

	key := &DatabaseItem{
		AccountId: accountId,
	}
	mp := make(map[string]struct{})
	cc.databases.data.Ascend(key, func(item *DatabaseItem) bool {
		if item.AccountId != accountId {
			return false
		}
		if item.Ts.Greater(ts) {
			return true
		}
		if _, ok := mp[item.Name]; !ok {
			mp[item.Name] = struct{}{}
			if !item.deleted {
				rs = append(rs, item.Name)
			}
		}
		return true
	})
	return rs
}

func (cc *CatalogCache) Tables(accountId uint32, databaseId uint64,
	ts timestamp.Timestamp) ([]string, []uint64) {
	var rs []string
	var rids []uint64

	key := &TableItem{
		AccountId:  accountId,
		DatabaseId: databaseId,
	}
	mp := make(map[string]struct{})
	cc.tables.data.Ascend(key, func(item *TableItem) bool {
		if item.AccountId != accountId {
			return false
		}
		if item.DatabaseId != databaseId {
			return false
		}

		if item.Ts.Greater(ts) {
			return true
		}
		if _, ok := mp[item.Name]; !ok {
			// How does this work?
			// 1. If there are two items in the same txn, non-deleted always comes first.
			// 2. if this item is deleted, the map will block the next item with the same name.
			mp[item.Name] = struct{}{}
			if !item.deleted {
				rs = append(rs, item.Name)
				rids = append(rids, item.Id)
			}
		}
		return true
	})
	return rs, rids
}

func (cc *CatalogCache) GetTableById(databaseId, tblId uint64) *TableItem {
	var rel *TableItem

	key := &TableItem{
		DatabaseId: databaseId,
	}
	// If account is much, the performance is very bad.
	cc.tables.data.Ascend(key, func(item *TableItem) bool {
		if item.Id == tblId {
			rel = item
			return false
		}
		return true
	})
	return rel
}

// GetTableByName returns the table item whose name is tableName in the database.
func (cc *CatalogCache) GetTableByName(databaseID uint64, tableName string) *TableItem {
	var rel *TableItem
	key := &TableItem{
		DatabaseId: databaseID,
	}
	cc.tables.data.Ascend(key, func(item *TableItem) bool {
		if item.Name == tableName {
			rel = item
			return false
		}
		return true
	})
	return rel
}

func (cc *CatalogCache) GetTable(tbl *TableItem) bool {
	var find bool

	cc.tables.data.Ascend(tbl, func(item *TableItem) bool {
		if item.Name != tbl.Name {
			return false
		}

		// just find once
		if !item.deleted {
			find = true
			if !item.initedByCol {
				logutil.Infof("yyyyy read incomplete table item %v", item.Name)
			}
			copyTableItem(tbl, item)
		}
		return false
	})

	return find
}

func (cc *CatalogCache) HasNewerVersion(qry *TableChangeQuery) bool {
	var find bool

	key := &TableItem{
		AccountId:  qry.AccountId,
		DatabaseId: qry.DatabaseId,
		Name:       qry.Name,
		Ts:         types.MaxTs().ToTimestamp(), // get the latest version
	}
	cc.tables.data.Ascend(key, func(item *TableItem) bool {
		if item.Name != qry.Name {
			return false
		}

		if item.Ts.Greater(qry.Ts) {
			if item.deleted || item.Id != qry.TableId {
				find = true
			}
		}
		return false
	})
	return find
}

func (cc *CatalogCache) GetDatabase(db *DatabaseItem) bool {
	var find bool

	cc.databases.data.Ascend(db, func(item *DatabaseItem) bool {
		if item.Name != db.Name {
			return false
		}

		// just find once
		if !item.deleted {
			find = true
			copyDatabaseItem(db, item)
		}
		return false
	})

	return find
}

func (cc *CatalogCache) DeleteTable(bat *batch.Batch) {
	cpks := bat.GetVector(MO_OFF + 0)
	timestamps := vector.MustFixedCol[types.TS](bat.GetVector(MO_TIMESTAMP_IDX))
	for i, ts := range timestamps {
		pk := cpks.GetBytesAt(i)
		if item, ok := cc.tables.cpkeyIndex.Get(&TableItem{CPKey: pk}); ok {
			// Note: the newItem.Id is the latest id under the name of the table,
			// not the id that can be seen at the moment ts.
			// Lucy thing is that the wrong tableid hold by delete item not used.
			newItem := &TableItem{
				deleted:    true,
				Id:         item.Id,
				Name:       item.Name,
				CPKey:      append([]byte{}, item.CPKey...),
				Rowid:      item.Rowid,
				AccountId:  item.AccountId,
				DatabaseId: item.DatabaseId,
				Ts:         ts.ToTimestamp(),
			}
			cc.tables.data.Set(newItem)
		}
	}
}

func (cc *CatalogCache) DeleteDatabase(bat *batch.Batch) {
	cpks := bat.GetVector(MO_OFF + 0)
	timestamps := vector.MustFixedCol[types.TS](bat.GetVector(MO_TIMESTAMP_IDX))
	for i, ts := range timestamps {
		pk := cpks.GetBytesAt(i)
		if item, ok := cc.databases.cpkeyIndex.Get(&DatabaseItem{CPKey: pk}); ok {
			newItem := &DatabaseItem{
				deleted:   true,
				Id:        item.Id,
				Name:      item.Name,
				Rowid:     item.Rowid,
				CPKey:     append([]byte{}, item.CPKey...),
				AccountId: item.AccountId,
				Typ:       item.Typ,
				CreateSql: item.CreateSql,
				Ts:        ts.ToTimestamp(),
			}
			cc.databases.data.Set(newItem)
		}
	}
}

func (cc *CatalogCache) InsertTable(bat *batch.Batch) {
	rowids := vector.MustFixedCol[types.Rowid](bat.GetVector(MO_ROWID_IDX))
	timestamps := vector.MustFixedCol[types.TS](bat.GetVector(MO_TIMESTAMP_IDX))
	accounts := vector.MustFixedCol[uint32](bat.GetVector(catalog.MO_TABLES_ACCOUNT_ID_IDX + MO_OFF))
	names := bat.GetVector(catalog.MO_TABLES_REL_NAME_IDX + MO_OFF)
	ids := vector.MustFixedCol[uint64](bat.GetVector(catalog.MO_TABLES_REL_ID_IDX + MO_OFF))
	databaseIds := vector.MustFixedCol[uint64](bat.GetVector(catalog.MO_TABLES_RELDATABASE_ID_IDX + MO_OFF))
	kinds := bat.GetVector(catalog.MO_TABLES_RELKIND_IDX + MO_OFF)
	comments := bat.GetVector(catalog.MO_TABLES_REL_COMMENT_IDX + MO_OFF)
	createSqls := bat.GetVector(catalog.MO_TABLES_REL_CREATESQL_IDX + MO_OFF)
	viewDefs := bat.GetVector(catalog.MO_TABLES_VIEWDEF_IDX + MO_OFF)
	partitioneds := vector.MustFixedCol[int8](bat.GetVector(catalog.MO_TABLES_PARTITIONED_IDX + MO_OFF))
	paritions := bat.GetVector(catalog.MO_TABLES_PARTITION_INFO_IDX + MO_OFF)
	constraints := bat.GetVector(catalog.MO_TABLES_CONSTRAINT_IDX + MO_OFF)
	versions := vector.MustFixedCol[uint32](bat.GetVector(catalog.MO_TABLES_VERSION_IDX + MO_OFF))
	catalogVersions := vector.MustFixedCol[uint32](bat.GetVector(catalog.MO_TABLES_CATALOG_VERSION_IDX + MO_OFF))
	pks := bat.GetVector(catalog.MO_TABLES_CPKEY_IDX + MO_OFF)
	for i, account := range accounts {
		item := new(TableItem)
		item.Id = ids[i]
		item.Name = names.GetStringAt(i)
		item.AccountId = account
		item.DatabaseId = databaseIds[i]
		item.Ts = timestamps[i].ToTimestamp()
		item.Kind = kinds.GetStringAt(i)
		item.ViewDef = viewDefs.GetStringAt(i)
		item.Constraint = append(item.Constraint, constraints.GetBytesAt(i)...)
		item.Comment = comments.GetStringAt(i)
		item.Partitioned = partitioneds[i]
		item.Partition = paritions.GetStringAt(i)
		item.CreateSql = createSqls.GetStringAt(i)
		item.Version = versions[i]
		item.CatalogVersion = catalogVersions[i]
		item.PrimaryIdx = -1
		item.PrimarySeqnum = -1
		item.ClusterByIdx = -1
		copy(item.Rowid[:], rowids[i][:])
		item.CPKey = append(item.CPKey, pks.GetBytesAt(i)...)

		cc.tables.data.Set(item)
		cc.tables.cpkeyIndex.Set(item)
		logutil.Infof("yyyyy insert table %v-%v-%s, %v", item.AccountId, item.DatabaseId, item.Name, hex.EncodeToString(pks.GetBytesAt(i)))
	}
}

func (cc *CatalogCache) InsertColumns(bat *batch.Batch) {
	var tblKey tableItemKey

	mp := make(map[tableItemKey]columns) // TableItem -> columns
	key := new(TableItem)
	rowids := vector.MustFixedCol[types.Rowid](bat.GetVector(MO_ROWID_IDX))
	// get table key info
	timestamps := vector.MustFixedCol[types.TS](bat.GetVector(MO_TIMESTAMP_IDX))
	accounts := vector.MustFixedCol[uint32](bat.GetVector(catalog.MO_COLUMNS_ACCOUNT_ID_IDX + MO_OFF))
	databaseIds := vector.MustFixedCol[uint64](bat.GetVector(catalog.MO_COLUMNS_ATT_DATABASE_ID_IDX + MO_OFF))
	tableNames := bat.GetVector(catalog.MO_COLUMNS_ATT_RELNAME_IDX + MO_OFF)
	tableIds := vector.MustFixedCol[uint64](bat.GetVector(catalog.MO_COLUMNS_ATT_RELNAME_ID_IDX + MO_OFF))
	// get columns info
	names := bat.GetVector(catalog.MO_COLUMNS_ATTNAME_IDX + MO_OFF)
	comments := bat.GetVector(catalog.MO_COLUMNS_ATT_COMMENT_IDX + MO_OFF)
	isHiddens := vector.MustFixedCol[int8](bat.GetVector(catalog.MO_COLUMNS_ATT_IS_HIDDEN_IDX + MO_OFF))
	isAutos := vector.MustFixedCol[int8](bat.GetVector(catalog.MO_COLUMNS_ATT_IS_AUTO_INCREMENT_IDX + MO_OFF))
	constraintTypes := bat.GetVector(catalog.MO_COLUMNS_ATT_CONSTRAINT_TYPE_IDX + MO_OFF)
	typs := bat.GetVector(catalog.MO_COLUMNS_ATTTYP_IDX + MO_OFF)
	hasDefs := vector.MustFixedCol[int8](bat.GetVector(catalog.MO_COLUMNS_ATTHASDEF_IDX + MO_OFF))
	defaultExprs := bat.GetVector(catalog.MO_COLUMNS_ATT_DEFAULT_IDX + MO_OFF)
	hasUpdates := vector.MustFixedCol[int8](bat.GetVector(catalog.MO_COLUMNS_ATT_HAS_UPDATE_IDX + MO_OFF))
	updateExprs := bat.GetVector(catalog.MO_COLUMNS_ATT_UPDATE_IDX + MO_OFF)
	nums := vector.MustFixedCol[int32](bat.GetVector(catalog.MO_COLUMNS_ATTNUM_IDX + MO_OFF))
	clusters := vector.MustFixedCol[int8](bat.GetVector(catalog.MO_COLUMNS_ATT_IS_CLUSTERBY + MO_OFF))
	seqnums := vector.MustFixedCol[uint16](bat.GetVector(catalog.MO_COLUMNS_ATT_SEQNUM_IDX + MO_OFF))
	enumValues := bat.GetVector(catalog.MO_COLUMNS_ATT_ENUM_IDX + MO_OFF)
	for i, account := range accounts {
		key.AccountId = account
		key.Name = tableNames.GetStringAt(i)
		key.DatabaseId = databaseIds[i]
		key.Ts = timestamps[i].ToTimestamp()
		key.Id = tableIds[i]
		tblKey.Name = key.Name
		tblKey.AccountId = key.AccountId
		tblKey.DatabaseId = key.DatabaseId
		tblKey.NodeId = key.Ts.NodeID
		tblKey.LogicalTime = key.Ts.LogicalTime
		tblKey.PhysicalTime = uint64(key.Ts.PhysicalTime)
		tblKey.Id = tableIds[i]
		if _, ok := cc.tables.data.Get(key); ok {
			col := column{
				num:             nums[i],
				name:            names.GetStringAt(i),
				comment:         comments.GetStringAt(i),
				isHidden:        isHiddens[i],
				isAutoIncrement: isAutos[i],
				hasDef:          hasDefs[i],
				hasUpdate:       hasUpdates[i],
				constraintType:  constraintTypes.GetStringAt(i),
				isClusterBy:     clusters[i],
				seqnum:          seqnums[i],
				enumValues:      enumValues.GetStringAt(i),
			}
			copy(col.rowid[:], rowids[i][:])
			col.typ = append(col.typ, typs.GetBytesAt(i)...)
			col.updateExpr = append(col.updateExpr, updateExprs.GetBytesAt(i)...)
			col.defaultExpr = append(col.defaultExpr, defaultExprs.GetBytesAt(i)...)
			mp[tblKey] = append(mp[tblKey], col)
		}
	}
	for k, cols := range mp {
		sort.Sort(cols)
		key.Name = k.Name
		key.AccountId = k.AccountId
		key.DatabaseId = k.DatabaseId
		key.Ts = timestamp.Timestamp{
			NodeID:       k.NodeId,
			PhysicalTime: int64(k.PhysicalTime),
			LogicalTime:  k.LogicalTime,
		}
		key.Id = k.Id
		item, _ := cc.tables.data.Get(key)
		coldefs := make([]engine.TableDef, 0, len(cols))
		for i, col := range cols {
			if col.constraintType == catalog.SystemColPKConstraint {
				item.PrimaryIdx = i
				item.PrimarySeqnum = int(col.seqnum)
			}
			if col.isClusterBy == 1 {
				item.ClusterByIdx = i
			}
			coldefs = append(coldefs, genTableDefOfColumn(col))
		}
		item.TableDef, item.Defs = getTableDef(item, coldefs)
		item.initedByCol = true
	}
}

func (cc *CatalogCache) InsertDatabase(bat *batch.Batch) {
	rowids := vector.MustFixedCol[types.Rowid](bat.GetVector(MO_ROWID_IDX))
	timestamps := vector.MustFixedCol[types.TS](bat.GetVector(MO_TIMESTAMP_IDX))
	accounts := vector.MustFixedCol[uint32](bat.GetVector(catalog.MO_DATABASE_ACCOUNT_ID_IDX + MO_OFF))
	names := bat.GetVector(catalog.MO_DATABASE_DAT_NAME_IDX + MO_OFF)
	ids := vector.MustFixedCol[uint64](bat.GetVector(catalog.MO_DATABASE_DAT_ID_IDX + MO_OFF))
	typs := bat.GetVector(catalog.MO_DATABASE_DAT_TYPE_IDX + MO_OFF)
	createSqls := bat.GetVector(catalog.MO_DATABASE_CREATESQL_IDX + MO_OFF)
	pks := bat.GetVector(catalog.MO_DATABASE_CPKEY_IDX + MO_OFF)
	for i, account := range accounts {
		item := new(DatabaseItem)
		item.Id = ids[i]
		item.Name = names.GetStringAt(i)
		item.AccountId = account
		item.Ts = timestamps[i].ToTimestamp()
		item.Typ = typs.GetStringAt(i)
		item.CreateSql = createSqls.GetStringAt(i)
		copy(item.Rowid[:], rowids[i][:])
		item.CPKey = append(item.CPKey, pks.GetBytesAt(i)...)
		cc.databases.data.Set(item)
		cc.databases.cpkeyIndex.Set(item)
		logutil.Infof("yyyyy insert db %v-%v cpk %v, %v", item.AccountId, item.Name, item.Rowid.ShortStringEx(), hex.EncodeToString(pks.GetBytesAt(i)))
	}
}

func genTableDefOfColumn(col column) engine.TableDef {
	var attr engine.Attribute

	attr.Name = col.name
	attr.ID = uint64(col.num)
	attr.Alg = compress.Lz4
	attr.Comment = col.comment
	attr.IsHidden = col.isHidden == 1
	attr.ClusterBy = col.isClusterBy == 1
	attr.AutoIncrement = col.isAutoIncrement == 1
	attr.Seqnum = col.seqnum
	attr.EnumVlaues = col.enumValues
	if err := types.Decode(col.typ, &attr.Type); err != nil {
		panic(err)
	}
	attr.Default = new(plan.Default)
	if col.hasDef == 1 {
		if err := types.Decode(col.defaultExpr, attr.Default); err != nil {
			panic(err)
		}
	}
	if col.hasUpdate == 1 {
		attr.OnUpdate = new(plan.OnUpdate)
		if err := types.Decode(col.updateExpr, attr.OnUpdate); err != nil {
			panic(err)
		}
	}
	if col.constraintType == catalog.SystemColPKConstraint {
		attr.Primary = true
	}
	return &engine.AttributeDef{Attr: attr}
}

func getTableDef(tblItem *TableItem, coldefs []engine.TableDef) (*plan.TableDef, []engine.TableDef) {
	var clusterByDef *plan.ClusterByDef
	var cols []*plan.ColDef
	var defs []*plan.TableDef_DefType
	var properties []*plan.Property
	var TableType string
	var Createsql string
	var partitionInfo *plan.PartitionByDef
	var viewSql *plan.ViewDef
	var foreignKeys []*plan.ForeignKeyDef
	var primarykey *plan.PrimaryKeyDef
	var indexes []*plan.IndexDef
	var refChildTbls []uint64

	tableDef := make([]engine.TableDef, 0)

	i := int32(0)
	name2index := make(map[string]int32)
	for _, def := range coldefs {
		if attr, ok := def.(*engine.AttributeDef); ok {
			name2index[attr.Attr.Name] = i
			cols = append(cols, &plan.ColDef{
				ColId: attr.Attr.ID,
				Name:  attr.Attr.Name,
				Typ: plan.Type{
					Id:          int32(attr.Attr.Type.Oid),
					Width:       attr.Attr.Type.Width,
					Scale:       attr.Attr.Type.Scale,
					AutoIncr:    attr.Attr.AutoIncrement,
					Table:       tblItem.Name,
					NotNullable: attr.Attr.Default != nil && !attr.Attr.Default.NullAbility,
					Enumvalues:  attr.Attr.EnumVlaues,
				},
				Primary:   attr.Attr.Primary,
				Default:   attr.Attr.Default,
				OnUpdate:  attr.Attr.OnUpdate,
				Comment:   attr.Attr.Comment,
				ClusterBy: attr.Attr.ClusterBy,
				Hidden:    attr.Attr.IsHidden,
				Seqnum:    uint32(attr.Attr.Seqnum),
			})
			if attr.Attr.ClusterBy {
				clusterByDef = &plan.ClusterByDef{
					Name: attr.Attr.Name,
				}
			}
			i++
		}
	}

	if tblItem.Comment != "" {
		properties = append(properties, &plan.Property{
			Key:   catalog.SystemRelAttr_Comment,
			Value: tblItem.Comment,
		})

		tableDef = append(tableDef, &engine.CommentDef{Comment: tblItem.Comment})
	}

	if tblItem.Partitioned > 0 {
		p := &plan.PartitionByDef{}
		err := p.UnMarshalPartitionInfo(([]byte)(tblItem.Partition))
		if err != nil {
			logutil.Errorf("cannot unmarshal partition metadata information: %v-%v-%v", tblItem.AccountId, tblItem.Id, tblItem.Name)
			return nil, nil
		}
		partitionInfo = p

		tableDef = append(tableDef, &engine.PartitionDef{
			Partitioned: tblItem.Partitioned,
			Partition:   tblItem.Partition,
		})
	}

	if tblItem.ViewDef != "" {
		viewSql = &plan.ViewDef{
			View: tblItem.ViewDef,
		}

		tableDef = append(tableDef, &engine.ViewDef{View: tblItem.ViewDef})
	}

	if len(tblItem.Constraint) > 0 {
		c := &engine.ConstraintDef{}
		err := c.UnmarshalBinary(tblItem.Constraint)
		if err != nil {
			logutil.Errorf("cannot unmarshal table constraint information: %v-%v-%v", tblItem.AccountId, tblItem.Id, tblItem.Name)
			return nil, nil
		}

		tableDef = append(tableDef, c)
		for _, ct := range c.Cts {
			switch k := ct.(type) {
			case *engine.IndexDef:
				indexes = k.Indexes
			case *engine.ForeignKeyDef:
				foreignKeys = k.Fkeys
			case *engine.RefChildTableDef:
				refChildTbls = k.Tables
			case *engine.PrimaryKeyDef:
				primarykey = k.Pkey
			case *engine.StreamConfigsDef:
				properties = append(properties, k.Configs...)
			}
		}
	}

	properties = append(properties, &plan.Property{
		Key:   catalog.SystemRelAttr_Kind,
		Value: tblItem.Kind,
	})
	TableType = tblItem.Kind

	props := &engine.PropertiesDef{}
	props.Properties = append(props.Properties, engine.Property{
		Key:   catalog.SystemRelAttr_Kind,
		Value: TableType,
	})

	if tblItem.CreateSql != "" {
		properties = append(properties, &plan.Property{
			Key:   catalog.SystemRelAttr_CreateSQL,
			Value: tblItem.CreateSql,
		})
		Createsql = tblItem.CreateSql

		props.Properties = append(props.Properties, engine.Property{
			Key:   catalog.SystemRelAttr_CreateSQL,
			Value: Createsql,
		})
	}

	tableDef = append(tableDef, props)
	tableDef = append(tableDef, coldefs...)
	if len(properties) > 0 {
		defs = append(defs, &plan.TableDef_DefType{
			Def: &plan.TableDef_DefType_Properties{
				Properties: &plan.PropertiesDef{
					Properties: properties,
				},
			},
		})
	}

	if primarykey != nil && primarykey.PkeyColName == catalog.CPrimaryKeyColName {
		primarykey.CompPkeyCol = plan2.GetColDefFromTable(cols, catalog.CPrimaryKeyColName)
	}
	if clusterByDef != nil && util.JudgeIsCompositeClusterByColumn(clusterByDef.Name) {
		clusterByDef.CompCbkeyCol = plan2.GetColDefFromTable(cols, clusterByDef.Name)
	}

	return &plan.TableDef{
		TblId:         tblItem.Id,
		Name:          tblItem.Name,
		Cols:          cols,
		Name2ColIndex: name2index,
		Defs:          defs,
		TableType:     TableType,
		Createsql:     Createsql,
		Pkey:          primarykey,
		ViewSql:       viewSql,
		Partition:     partitionInfo,
		Fkeys:         foreignKeys,
		RefChildTbls:  refChildTbls,
		ClusterBy:     clusterByDef,
		Indexes:       indexes,
		Version:       tblItem.Version,
	}, tableDef
}
