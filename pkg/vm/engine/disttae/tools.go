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
	"slices"
	"time"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/defines"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/pb/api"
	"github.com/matrixorigin/matrixone/pkg/pb/metadata"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/pb/txn"
	"github.com/matrixorigin/matrixone/pkg/txn/trace"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/blockio"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

func genCreateDatabaseTuple(sql string, accountId, userId, roleId uint32,
	name string, databaseId uint64, typ string,
	m *mpool.MPool, packer *types.Packer,
) (*batch.Batch, error) {
	bat := batch.NewWithSize(len(catalog.MoDatabaseSchema))
	bat.Attrs = append(bat.Attrs, catalog.MoDatabaseSchema...)
	bat.SetRowCount(1)

	packer.Reset()
	var err error
	defer func() {
		packer.Reset()
		if err != nil {
			bat.Clean(m)
		}
	}()

	{
		idx := catalog.MO_DATABASE_DAT_ID_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoDatabaseTypes[idx]) // dat_id
		if err = vector.AppendFixed(bat.Vecs[idx], databaseId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_DATABASE_DAT_NAME_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoDatabaseTypes[idx]) // datname
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(name), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_DATABASE_DAT_CATALOG_NAME_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoDatabaseTypes[idx]) // dat_catalog_name
		val := []byte(catalog.SystemCatalogName)
		if name == catalog.MO_CATALOG { // historic debt
			val = []byte(catalog.MO_CATALOG)
		}
		if err = vector.AppendBytes(bat.Vecs[idx], val, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_DATABASE_CREATESQL_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoDatabaseTypes[idx])                     // dat_createsql
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(sql), false, m); err != nil { // TODO
			return nil, err
		}
		idx = catalog.MO_DATABASE_OWNER_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoDatabaseTypes[idx]) // owner
		if err = vector.AppendFixed(bat.Vecs[idx], roleId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_DATABASE_CREATOR_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoDatabaseTypes[idx]) // creator
		if err = vector.AppendFixed(bat.Vecs[idx], userId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_DATABASE_CREATED_TIME_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoDatabaseTypes[idx]) // created_time
		if err = vector.AppendFixed(bat.Vecs[idx], types.Timestamp(time.Now().UnixMicro()+types.GetUnixEpochSecs()), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_DATABASE_ACCOUNT_ID_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoDatabaseTypes[idx]) // account_id
		if err = vector.AppendFixed(bat.Vecs[idx], accountId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_DATABASE_DAT_TYPE_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoDatabaseTypes[idx])                     // dat_type
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(typ), false, m); err != nil { // TODO
			return nil, err
		}

		idx = catalog.MO_DATABASE_CPKEY_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoDatabaseTypes[idx]) // cpkey
		packer.EncodeUint32(accountId)
		packer.EncodeStringType([]byte(name))
		if err = vector.AppendBytes(bat.Vecs[idx], packer.Bytes(), false, m); err != nil {
			return nil, err
		}
	}
	return bat, nil
}

func genDropDatabaseTuple(
	rowid types.Rowid, accid uint32, datid uint64, name string,
	m *mpool.MPool, packer *types.Packer,
) (*batch.Batch, error) {
	bat := batch.NewWithSize(4)
	bat.Attrs = append([]string{catalog.Row_ID, catalog.CPrimaryKeyColName}, catalog.MoDatabaseSchema[:2]...)
	bat.SetRowCount(1)
	var err error
	packer.Reset()
	defer func() {
		packer.Reset()
		if err != nil {
			bat.Clean(m)
		}
	}()

	//add the rowid vector as the first one in the batch
	rowidVec := vector.NewVec(types.T_Rowid.ToType())
	if err = vector.AppendFixed(rowidVec, rowid, false, m); err != nil {
		return nil, err
	}
	bat.Vecs[0] = rowidVec

	// add the cpkey vector as the second one in the batch
	cpkVec := vector.NewVec(types.T_varchar.ToType()) // cpkey of accid+dbname
	packer.EncodeUint32(accid)
	packer.EncodeStringType([]byte(name))
	if err = vector.AppendBytes(cpkVec, packer.Bytes(), false, m); err != nil {
		return nil, err
	}
	bat.Vecs[1] = cpkVec

	// add supplementary info to generate ddl cmd for TN handler
	{
		bat.Vecs[2] = vector.NewVec(catalog.MoDatabaseTypes[catalog.MO_DATABASE_DAT_ID_IDX]) // dat_id
		if err = vector.AppendFixed(bat.Vecs[2], datid, false, m); err != nil {
			return nil, err
		}
		bat.Vecs[3] = vector.NewVec(catalog.MoDatabaseTypes[catalog.MO_DATABASE_DAT_NAME_IDX]) // datname
		if err = vector.AppendBytes(bat.Vecs[3], []byte(name), false, m); err != nil {
			return nil, err
		}
	}

	return bat, nil
}

func genTableAlterTuple(constraint [][]byte, m *mpool.MPool) (*batch.Batch, error) {
	bat := batch.NewWithSize(1)
	bat.Attrs = append(bat.Attrs, catalog.SystemRelAttr_Constraint)
	bat.SetRowCount(1)

	var err error
	defer func() {
		if err != nil {
			bat.Clean(m)
		}
	}()

	idx := catalog.MO_TABLES_ALTER_TABLE
	bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[catalog.MO_TABLES_CONSTRAINT_IDX]) // constraint
	for i := 0; i < len(constraint); i++ {
		if err = vector.AppendBytes(bat.Vecs[idx], constraint[i], false, m); err != nil {
			return nil, err
		}
	}
	return bat, nil
}

// genCreateTableTuple yields a batch for insertion into mo_tables.
func genCreateTableTuple(tbl *txnTable, accountId, userId, roleId uint32, name string,
	tableId uint64, databaseId uint64, databaseName string, m *mpool.MPool, packer *types.Packer) (*batch.Batch, error) {
	bat := batch.NewWithSize(len(catalog.MoTablesSchema))
	bat.Attrs = append(bat.Attrs, catalog.MoTablesSchema...)
	bat.SetRowCount(1)

	var err error
	packer.Reset()
	defer func() {
		packer.Reset()
		if err != nil {
			bat.Clean(m)
		}
	}()

	{
		idx := catalog.MO_TABLES_REL_ID_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // rel_id
		if err = vector.AppendFixed(bat.Vecs[idx], tableId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_REL_NAME_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // relname
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(name), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_RELDATABASE_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // reldatabase
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(databaseName), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_RELDATABASE_ID_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // reldatabase_id
		if err = vector.AppendFixed(bat.Vecs[idx], databaseId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_RELPERSISTENCE_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // relpersistence
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(""), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_RELKIND_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // relkind
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(tbl.relKind), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_REL_COMMENT_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // rel_comment
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(tbl.comment), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_REL_CREATESQL_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // rel_createsql
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(tbl.createSql), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_CREATED_TIME_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // created_time
		if err = vector.AppendFixed(bat.Vecs[idx], types.Timestamp(time.Now().Unix()), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_CREATOR_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // creator
		if err = vector.AppendFixed(bat.Vecs[idx], userId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_OWNER_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // owner
		if err = vector.AppendFixed(bat.Vecs[idx], roleId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_ACCOUNT_ID_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // account_id
		if err = vector.AppendFixed(bat.Vecs[idx], accountId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_PARTITIONED_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // partitioned
		if err = vector.AppendFixed(bat.Vecs[idx], tbl.partitioned, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_PARTITION_INFO_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // partition_info
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(tbl.partition), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_VIEWDEF_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // viewdef
		if err := vector.AppendBytes(bat.Vecs[idx], []byte(tbl.viewdef), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_CONSTRAINT_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // constraint
		if err = vector.AppendBytes(bat.Vecs[idx], tbl.constraint, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_VERSION_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // schema_version
		if err = vector.AppendFixed(bat.Vecs[idx], uint32(tbl.version), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_CATALOG_VERSION_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // catalog version
		if err = vector.AppendFixed(bat.Vecs[idx], catalog.CatalogVersion_Curr, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_CPKEY_IDX
		bat.Vecs[idx] = vector.NewVec(catalog.MoTablesTypes[idx]) // cpkey
		packer.EncodeUint32(accountId)
		packer.EncodeStringType([]byte(databaseName))
		packer.EncodeStringType([]byte(name))
		if err = vector.AppendBytes(bat.Vecs[idx], packer.Bytes(), false, m); err != nil {
			return nil, err
		}
	}
	return bat, nil
}

// genCreateColumnTuples yields a batch for insertion into mo_columns.
func genCreateColumnTuples(cols []column, m *mpool.MPool) (*batch.Batch, error) {
	bat := batch.NewWithSize(len(catalog.MoColumnsSchema))
	bat.Attrs = append(bat.Attrs, catalog.MoColumnsSchema...)
	bat.SetRowCount(len(cols))

	var err error
	defer func() {
		if err != nil {
			bat.Clean(m)
		}
	}()
	for i := 0; i <= catalog.MO_COLUMNS_MAXIDX; i++ {
		bat.Vecs[i] = vector.NewVec(catalog.MoColumnsTypes[i])
		bat.Vecs[i].PreExtend(len(cols), m)
	}
	for _, col := range cols {
		idx := catalog.MO_COLUMNS_ATT_UNIQ_NAME_IDX
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(genColumnPrimaryKey(col.tableId, col.name)),
			false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ACCOUNT_ID_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], col.accountId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_DATABASE_ID_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], col.databaseId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_DATABASE_IDX
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(col.databaseName), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_RELNAME_ID_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], col.tableId, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_RELNAME_IDX
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(col.tableName), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATTNAME_IDX
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(col.name), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATTTYP_IDX
		if err = vector.AppendBytes(bat.Vecs[idx], col.typ, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATTNUM_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], col.num, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_LENGTH_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], col.typLen, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATTNOTNULL_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], col.notNull, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATTHASDEF_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], col.hasDef, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_DEFAULT_IDX
		if err = vector.AppendBytes(bat.Vecs[idx], col.defaultExpr, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATTISDROPPED_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], int8(0), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_CONSTRAINT_TYPE_IDX
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(col.constraintType), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_IS_UNSIGNED_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], int8(0), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_IS_AUTO_INCREMENT_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], col.isAutoIncrement, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_COMMENT_IDX
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(col.comment), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_IS_HIDDEN_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], col.isHidden, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_HAS_UPDATE_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], col.hasUpdate, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_UPDATE_IDX
		if err = vector.AppendBytes(bat.Vecs[idx], col.updateExpr, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_IS_CLUSTERBY
		if err = vector.AppendFixed(bat.Vecs[idx], col.isClusterBy, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_SEQNUM_IDX
		if err = vector.AppendFixed(bat.Vecs[idx], col.seqnum, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_COLUMNS_ATT_ENUM_IDX
		if err = vector.AppendBytes(bat.Vecs[idx], []byte(col.enumValues), false, m); err != nil {
			return nil, err
		}
	}
	return bat, nil
}

// genDropColumnTuple generates the batch for deletion on mo_columns.
// the batch has rowid vector.
func genDropColumnTuples(rowids []types.Rowid, pks []string, m *mpool.MPool) (*batch.Batch, error) {
	bat := batch.NewWithSize(2)
	bat.Attrs = []string{catalog.Row_ID, catalog.CPrimaryKeyColName}
	bat.SetRowCount(len(rowids))

	var err error
	defer func() {
		if err != nil {
			bat.Clean(m)
		}
	}()

	rowidVec := vector.NewVec(types.T_Rowid.ToType())
	for _, rowid := range rowids {
		if err = vector.AppendFixed(rowidVec, rowid, false, m); err != nil {
			return nil, err
		}
	}
	bat.Vecs[0] = rowidVec

	pkVec := vector.NewVec(types.T_varchar.ToType())
	for _, s := range pks {
		if err = vector.AppendBytes(pkVec, []byte(s), false, m); err != nil {
			return nil, err
		}
	}
	bat.Vecs[1] = pkVec
	return bat, nil
}

// genDropTableTuple generates the batch for deletion on mo_tables.
// the batch has rowid vector.
func genDropTableTuple(rowid types.Rowid, accid uint32, id, databaseId uint64, name, databaseName string,
	m *mpool.MPool, packer *types.Packer) (*batch.Batch, error) {
	bat := batch.NewWithSize(6)
	bat.Attrs = append([]string{catalog.Row_ID, catalog.CPrimaryKeyColName}, catalog.MoTablesSchema[:4]...)
	bat.SetRowCount(1)

	var err error
	packer.Reset()
	defer func() {
		packer.Reset()
		if err != nil {
			bat.Clean(m)
		}
	}()

	//add the rowid vector as the first one in the batch
	rowidVec := vector.NewVec(types.T_Rowid.ToType())
	if err = vector.AppendFixed(rowidVec, rowid, false, m); err != nil {
		return nil, err
	}
	bat.Vecs[0] = rowidVec

	cpkVec := vector.NewVec(types.T_varchar.ToType()) // cpkey of acc_id + db_name + tbl_name
	packer.EncodeUint32(accid)
	packer.EncodeStringType([]byte(databaseName))
	packer.EncodeStringType([]byte(name))
	if err = vector.AppendBytes(cpkVec, packer.Bytes(), false, m); err != nil {
		return nil, err
	}
	bat.Vecs[1] = cpkVec

	{
		off := 2
		idx := catalog.MO_TABLES_REL_ID_IDX
		bat.Vecs[idx+off] = vector.NewVec(catalog.MoTablesTypes[idx]) // rel_id
		if err = vector.AppendFixed(bat.Vecs[idx+off], id, false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_REL_NAME_IDX
		bat.Vecs[idx+off] = vector.NewVec(catalog.MoTablesTypes[idx]) // relname
		if err = vector.AppendBytes(bat.Vecs[idx+off], []byte(name), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_RELDATABASE_IDX
		bat.Vecs[idx+off] = vector.NewVec(catalog.MoTablesTypes[idx]) // reldatabase
		if err = vector.AppendBytes(bat.Vecs[idx+off], []byte(databaseName), false, m); err != nil {
			return nil, err
		}
		idx = catalog.MO_TABLES_RELDATABASE_ID_IDX
		bat.Vecs[idx+off] = vector.NewVec(catalog.MoTablesTypes[idx]) // reldatabase_id
		if err = vector.AppendFixed(bat.Vecs[idx+off], databaseId, false, m); err != nil {
			return nil, err
		}
	}

	return bat, nil
}

func newColumnExpr(pos int, typ plan.Type, name string) *plan.Expr {
	return &plan.Expr{
		Typ: typ,
		Expr: &plan.Expr_Col{
			Col: &plan.ColRef{
				Name:   name,
				ColPos: int32(pos),
			},
		},
	}
}

func genWriteReqs(ctx context.Context, txnCommit *Transaction) ([]txn.TxnRequest, error) {
	writes, tablesInVain, op := txnCommit.writes, txnCommit.tablesInVain, txnCommit.op
	var pkChkByTN int8
	if v := ctx.Value(defines.PkCheckByTN{}); v != nil {
		pkChkByTN = v.(int8)
	}
	var tnID string
	var tn metadata.TNService
	entries := make([]*api.Entry, 0, len(writes))
	for _, e := range writes {
		if tnID == "" {
			tnID = e.tnStore.ServiceID
			tn = e.tnStore
		}
		if tnID != "" && tnID != e.tnStore.ServiceID {
			panic(fmt.Sprintf("txnCommit contains entries from different TNs, %s != %s", tnID, e.tnStore.ServiceID))
		}
		if e.bat == nil || e.bat.IsEmpty() {
			continue
		}
		if slices.Contains(tablesInVain, e.tableId) { // cancel dml and alter request
			continue
		}
		e.pkChkByTN = pkChkByTN
		pe, err := toPBEntry(e)
		if err != nil {
			return nil, err
		}
		// --sql
		// create table t (a int);
		// begin;
		// alter table t comment 'will come back';
		// drop table t;
		// commit;
		//
		// the txn wrote a delete & insert batch due to alter, and the insert batch was cancelled by dropping.
		// the table should be dropped in TN, so we need to reset the delete batch to normal delete.
		isAlter, typ, id, name := noteSplitAlter(e.note)
		if isAlter && typ == DELETE && slices.Contains(tablesInVain, id) {
			// reset to normal delete, this will lead to dropping table in TN
			e.note = noteForDrop(id, name)
		} else if isAlter {
			// To tell TN, this is an update due to alter, do not touch catalog
			pe.TableName = "alter"
		}
		entries = append(entries, pe)
	}

	trace.GetService().TxnCommit(op, entries)
	if len(entries) == 0 {
		return nil, nil
	}
	reqs := make([]txn.TxnRequest, 0, len(entries))
	payload, err := types.Encode(&api.PrecommitWriteCmd{EntryList: entries})
	if err != nil {
		return nil, err
	}
	for _, info := range tn.Shards {
		reqs = append(reqs, txn.TxnRequest{
			CNRequest: &txn.CNOpRequest{
				OpCode:  uint32(api.OpCode_OpPreCommit),
				Payload: payload,
				Target: metadata.TNShard{
					TNShardRecord: metadata.TNShardRecord{
						ShardID: info.ShardID,
					},
					ReplicaID: info.ReplicaID,
					Address:   tn.TxnServiceAddress,
				},
			},
			Options: &txn.TxnRequestOptions{
				RetryCodes: []int32{
					// tn shard not found
					int32(moerr.ErrTNShardNotFound),
				},
				RetryInterval: int64(time.Second),
			},
		})
	}
	return reqs, nil
}

func toPBEntry(e Entry) (*api.Entry, error) {
	var ebat *batch.Batch

	if e.typ == INSERT {
		ebat = batch.NewWithSize(0)
		if e.bat.Attrs[0] == catalog.BlockMeta_MetaLoc {
			ebat.Vecs = e.bat.Vecs
			ebat.Attrs = e.bat.Attrs
		} else {
			//e.bat.Vecs[0] is rowid vector
			ebat.Vecs = e.bat.Vecs[1:]
			ebat.Attrs = e.bat.Attrs[1:]
		}
	} else {
		ebat = e.bat
	}
	typ := api.Entry_Insert
	if e.typ == DELETE {
		typ = api.Entry_Delete
		// ddl drop bat includes extra information to generate command in TN
		if e.tableId != catalog.MO_TABLES_ID &&
			e.tableId != catalog.MO_DATABASE_ID {
			ebat = batch.NewWithSize(0)
			if e.fileName == "" {
				if len(e.bat.Vecs) != 2 {
					panic(fmt.Sprintf("e.bat should contain 2 vectors, "+
						"one is rowid vector, the other is pk vector,"+
						"database name = %s, table name = %s", e.databaseName, e.tableName))
				}
				ebat.Vecs = e.bat.Vecs[:2]
				ebat.Attrs = e.bat.Attrs[:2]
			} else {
				ebat.Vecs = e.bat.Vecs[:1]
				ebat.Attrs = e.bat.Attrs[:1]
			}
		}

	} else if e.typ == ALTER {
		typ = api.Entry_Alter
	}
	bat, err := toPBBatch(ebat)
	if err != nil {
		return nil, err
	}
	return &api.Entry{
		Bat:          bat,
		EntryType:    typ,
		TableId:      e.tableId,
		DatabaseId:   e.databaseId,
		TableName:    e.tableName,
		DatabaseName: e.databaseName,
		FileName:     e.fileName,
		PkCheckByTn:  int32(e.pkChkByTN),
	}, nil
}

func toPBBatch(bat *batch.Batch) (*api.Batch, error) {
	rbat := new(api.Batch)
	rbat.Attrs = bat.Attrs
	for _, vec := range bat.Vecs {
		pbVector, err := vector.VectorToProtoVector(vec)
		if err != nil {
			return nil, err
		}
		rbat.Vecs = append(rbat.Vecs, pbVector)
	}
	return rbat, nil
}

// genColumnsFromDefs generates column struct from TableDef.
//
// NOTE: 1. it will modify the input TableDef.
// 2. it is usually used in creating new table.
// 3. It will append rowid column as the last column, which is **incorrect** if we want impl alter column gracefully.
func genColumnsFromDefs(accountId uint32, tableName, databaseName string,
	tableId, databaseId uint64, defs []engine.TableDef) ([]column, error) {
	{
		mp := make(map[string]int)
		for i, def := range defs {
			if attr, ok := def.(*engine.AttributeDef); ok {
				mp[attr.Attr.Name] = i
			}
		}
		for _, def := range defs {
			if constraintDef, ok := def.(*engine.ConstraintDef); ok {
				for _, ct := range constraintDef.Cts {
					if pkdef, ok2 := ct.(*engine.PrimaryKeyDef); ok2 {
						pos := mp[pkdef.Pkey.PkeyColName]
						attr, _ := defs[pos].(*engine.AttributeDef)
						attr.Attr.Primary = true
					}
				}
			}

			if clusterByDef, ok := def.(*engine.ClusterByDef); ok {
				attr, _ := defs[mp[clusterByDef.Name]].(*engine.AttributeDef)
				attr.Attr.ClusterBy = true
			}
		}
	}
	var num int32 = 1
	cols := make([]column, 0, len(defs))
	for _, def := range defs {
		attrDef, ok := def.(*engine.AttributeDef)
		if !ok || attrDef.Attr.Name == catalog.Row_ID {
			continue
		}
		typ, err := types.Encode(&attrDef.Attr.Type)
		if err != nil {
			return nil, err
		}
		col := column{
			typ:          typ,
			typLen:       int32(len(typ)),
			accountId:    accountId,
			tableId:      tableId,
			databaseId:   databaseId,
			name:         attrDef.Attr.Name,
			tableName:    tableName,
			databaseName: databaseName,
			num:          num,
			comment:      attrDef.Attr.Comment,
			seqnum:       uint16(num - 1),
			enumValues:   attrDef.Attr.EnumVlaues,
		}
		attrDef.Attr.ID = uint64(num)
		attrDef.Attr.Seqnum = uint16(num - 1)
		if attrDef.Attr.Default != nil {
			if !attrDef.Attr.Default.NullAbility {
				col.notNull = 1
			}
			defaultExpr, err := types.Encode(attrDef.Attr.Default)
			if err != nil {
				return nil, err
			}
			if len(defaultExpr) > 0 {
				col.hasDef = 1
				col.defaultExpr = defaultExpr
			}
		}
		if attrDef.Attr.OnUpdate != nil {
			expr, err := types.Encode(attrDef.Attr.OnUpdate)
			if err != nil {
				return nil, err
			}
			if len(expr) > 0 {
				col.hasUpdate = 1
				col.updateExpr = expr
			}
		}
		if attrDef.Attr.IsHidden {
			col.isHidden = 1
		}
		if attrDef.Attr.AutoIncrement {
			col.isAutoIncrement = 1
		}
		if attrDef.Attr.Primary {
			col.constraintType = catalog.SystemColPKConstraint
		} else {
			col.constraintType = catalog.SystemColNoConstraint
		}
		if attrDef.Attr.ClusterBy {
			col.isClusterBy = 1
		}

		cols = append(cols, col)
		num++
	}

	// add rowid column
	rowidTyp := types.T_Rowid.ToType()
	typ, _ := types.Encode(&rowidTyp)
	cols = append(cols, column{
		typ:          typ,
		typLen:       int32(len(typ)),
		accountId:    accountId,
		tableId:      tableId,
		databaseId:   databaseId,
		name:         catalog.Row_ID,
		tableName:    tableName,
		databaseName: databaseName,
		num:          int32(len(cols) + 1),
		seqnum:       uint16(len(cols)),
		isHidden:     1,
		notNull:      1,
	})

	return cols, nil
}

func getSql(ctx context.Context) string {
	if v := ctx.Value(defines.SqlKey{}); v != nil {
		return v.(string)
	}
	return ""
}
func getTyp(ctx context.Context) string {
	if v := ctx.Value(defines.DatTypKey{}); v != nil {
		return v.(string)
	}
	return ""
}

func getAccessInfo(ctx context.Context) (uint32, uint32, uint32, error) {
	var accountId, userId, roleId uint32
	var err error

	accountId, err = defines.GetAccountId(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	userId = defines.GetUserId(ctx)
	roleId = defines.GetRoleId(ctx)
	return accountId, userId, roleId, nil
}

func genDatabaseKey(id uint32, name string) databaseKey {
	return databaseKey{
		name:      name,
		accountId: id,
	}
}

func genTableKey(id uint32, name string, databaseId uint64) tableKey {
	return tableKey{
		name:       name,
		databaseId: databaseId,
		accountId:  id,
	}
}

// fillRandomRowidAndZeroTs modifies the input batch and returns the proto batch as a shallow copy.
func fillRandomRowidAndZeroTs(bat *batch.Batch, m *mpool.MPool) (*api.Batch, error) {
	var attrs []string
	vecs := make([]*vector.Vector, 0, 2)

	{
		vec := vector.NewVec(types.T_Rowid.ToType())
		for i := 0; i < bat.RowCount(); i++ {
			val := types.RandomRowid()
			if err := vector.AppendFixed(vec, val, false, m); err != nil {
				vec.Free(m)
				return nil, err
			}
		}
		vecs = append(vecs, vec)
		attrs = append(attrs, catalog.Row_ID)
	}
	{
		var val types.TS

		vec := vector.NewVec(types.T_TS.ToType())
		for i := 0; i < bat.RowCount(); i++ {
			if err := vector.AppendFixed(vec, val, false, m); err != nil {
				vecs[0].Free(m)
				vec.Free(m)
				return nil, err
			}
		}
		vecs = append(vecs, vec)
		attrs = append(attrs, catalog.TableTailAttrCommitTs)
	}
	bat.Vecs = append(vecs, bat.Vecs...)
	bat.Attrs = append(attrs, bat.Attrs...)
	return batch.BatchToProtoBatch(bat)
}

func genColumnPrimaryKey(tableId uint64, name string) string {
	return fmt.Sprintf("%v-%v", tableId, name)
}

func getColPks(tid uint64, cols []*plan.ColDef) []string {
	pks := make([]string, 0, len(cols))
	for _, col := range cols {
		pks = append(pks, genColumnPrimaryKey(tid, col.Name))
	}
	return pks
}

func transferIval[T int32 | int64](v T, oid types.T) (bool, any) {
	switch oid {
	case types.T_bit:
		return true, uint64(v)
	case types.T_int8:
		return true, int8(v)
	case types.T_int16:
		return true, int16(v)
	case types.T_int32:
		return true, int32(v)
	case types.T_int64:
		return true, int64(v)
	case types.T_uint8:
		return true, uint8(v)
	case types.T_uint16:
		return true, uint16(v)
	case types.T_uint32:
		return true, uint32(v)
	case types.T_uint64:
		return true, uint64(v)
	case types.T_float32:
		return true, float32(v)
	case types.T_float64:
		return true, float64(v)
	default:
		return false, nil
	}
}

func transferUval[T uint32 | uint64](v T, oid types.T) (bool, any) {
	switch oid {
	case types.T_bit:
		return true, uint64(v)
	case types.T_int8:
		return true, int8(v)
	case types.T_int16:
		return true, int16(v)
	case types.T_int32:
		return true, int32(v)
	case types.T_int64:
		return true, int64(v)
	case types.T_uint8:
		return true, uint8(v)
	case types.T_uint16:
		return true, uint16(v)
	case types.T_uint32:
		return true, uint32(v)
	case types.T_uint64:
		return true, uint64(v)
	case types.T_float32:
		return true, float32(v)
	case types.T_float64:
		return true, float64(v)
	default:
		return false, nil
	}
}

func transferFval(v float32, oid types.T) (bool, any) {
	switch oid {
	case types.T_float32:
		return true, float32(v)
	case types.T_float64:
		return true, float64(v)
	default:
		return false, nil
	}
}

func transferDval(v float64, oid types.T) (bool, any) {
	switch oid {
	case types.T_float32:
		return true, float32(v)
	case types.T_float64:
		return true, float64(v)
	default:
		return false, nil
	}
}

func transferSval(v string, oid types.T) (bool, any) {
	switch oid {
	case types.T_json:
		return true, []byte(v)
	case types.T_char, types.T_varchar:
		return true, []byte(v)
	case types.T_text, types.T_blob:
		return true, []byte(v)
	case types.T_binary, types.T_varbinary:
		return true, []byte(v)
	case types.T_uuid:
		var uv types.Uuid
		copy(uv[:], []byte(v)[:])
		return true, uv
		//TODO: should we add T_array for this code?
	default:
		return false, nil
	}
}

func transferBval(v bool, oid types.T) (bool, any) {
	switch oid {
	case types.T_bool:
		return true, v
	default:
		return false, nil
	}
}

func transferDateval(v int32, oid types.T) (bool, any) {
	switch oid {
	case types.T_date:
		return true, types.Date(v)
	default:
		return false, nil
	}
}

func transferTimeval(v int64, oid types.T) (bool, any) {
	switch oid {
	case types.T_time:
		return true, types.Time(v)
	default:
		return false, nil
	}
}

func transferDatetimeval(v int64, oid types.T) (bool, any) {
	switch oid {
	case types.T_datetime:
		return true, types.Datetime(v)
	default:
		return false, nil
	}
}

func transferTimestampval(v int64, oid types.T) (bool, any) {
	switch oid {
	case types.T_timestamp:
		return true, types.Timestamp(v)
	default:
		return false, nil
	}
}

func transferDecimal64val(v int64, oid types.T) (bool, any) {
	switch oid {
	case types.T_decimal64:
		return true, types.Decimal64(v)
	default:
		return false, nil
	}
}

func transferDecimal128val(a, b int64, oid types.T) (bool, any) {
	switch oid {
	case types.T_decimal128:
		return true, types.Decimal128{B0_63: uint64(a), B64_127: uint64(b)}
	default:
		return false, nil
	}
}

func groupBlocksToObjects(blkInfos []*objectio.BlockInfo, dop int) ([][]*objectio.BlockInfo, []int) {
	var infos [][]*objectio.BlockInfo
	objMap := make(map[string]int)
	lenObjs := 0
	for _, blkInfo := range blkInfos {
		//block := catalog.DecodeBlockInfo(blkInfos[i])
		objName := blkInfo.MetaLocation().Name().String()
		if idx, ok := objMap[objName]; ok {
			infos[idx] = append(infos[idx], blkInfo)
		} else {
			objMap[objName] = lenObjs
			lenObjs++
			infos = append(infos, []*objectio.BlockInfo{blkInfo})
		}
	}
	steps := make([]int, len(infos))
	currentBlocks := 0
	for i := range infos {
		steps[i] = (currentBlocks-PREFETCH_THRESHOLD)/dop - PREFETCH_ROUNDS
		if steps[i] < 0 {
			steps[i] = 0
		}
		currentBlocks += len(infos[i])
	}
	return infos, steps
}

func newBlockReaders(ctx context.Context, fs fileservice.FileService, tblDef *plan.TableDef,
	ts timestamp.Timestamp, num int, expr *plan.Expr, filter blockio.BlockReadFilter,
	proc *process.Process) []*blockReader {
	rds := make([]*blockReader, num)
	for i := 0; i < num; i++ {
		rds[i] = newBlockReader(
			ctx, tblDef, ts, nil, expr, filter, fs, proc,
		)
	}
	return rds
}

func distributeBlocksToBlockReaders(rds []*blockReader, numOfReaders int, numOfBlocks int, infos [][]*objectio.BlockInfo, steps []int) []*blockReader {
	readerIndex := 0
	for i := range infos {
		//distribute objects and steps for prefetch
		rds[readerIndex].steps = append(rds[readerIndex].steps, steps[i])
		rds[readerIndex].infos = append(rds[readerIndex].infos, infos[i])
		for j := range infos[i] {
			//distribute block
			rds[readerIndex].blks = append(rds[readerIndex].blks, infos[i][j])
			readerIndex++
			readerIndex = readerIndex % numOfReaders
		}
	}
	scanType := NORMAL
	if numOfBlocks < numOfReaders*SMALLSCAN_THRESHOLD {
		scanType = SMALL
	} else if (numOfReaders * LARGESCAN_THRESHOLD) <= numOfBlocks {
		scanType = LARGE
	}
	for i := range rds {
		rds[i].scanType = scanType
	}
	return rds
}
