// Copyright 2023 Matrix Origin
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

package plan

import (
	"strings"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
	"github.com/matrixorigin/matrixone/pkg/sql/parsers/tree"
)

// RenameColumn Can change a column name but not its definition.
// More convenient than CHANGE to rename a column without changing its definition.
func RenameColumn(ctx CompilerContext, alterPlan *plan.AlterTable, spec *tree.AlterTableRenameColumnClause, alterCtx *AlterTableContext) error {
	tableDef := alterPlan.CopyTableDef

	// get the old column name
	oldColName := spec.OldColumnName.ColName()
	oldColNameOrigin := spec.OldColumnName.ColNameOrigin()

	// get the new column name
	newColName := spec.NewColumnName.ColName()
	newColNameOrigin := spec.NewColumnName.ColNameOrigin()

	// Check whether original column has existed.
	oldCol := FindColumn(tableDef.Cols, oldColName)
	if oldCol == nil || oldCol.Hidden {
		return moerr.NewBadFieldError(ctx.GetContext(), oldColNameOrigin, alterPlan.TableDef.Name)
	}

	if oldColNameOrigin == newColNameOrigin {
		return nil
	}

	// Check if the new column name is valid and conflicts with internal hidden columns
	if err := CheckColumnNameValid(ctx.GetContext(), newColName); err != nil {
		return err
	}

	// If you want to rename the original column name to new name, you need to first check if the new name already exists.
	if newColName != oldColName {
		if FindColumn(tableDef.Cols, newColName) != nil {
			return moerr.NewErrDupFieldName(ctx.GetContext(), newColNameOrigin)
		}

		// If the column name of the table changes, it is necessary to check if it is associated
		// with the index key. If it is an index key column, column name replacement is required.
		for _, indexInfo := range alterPlan.CopyTableDef.Indexes {
			for j, partCol := range indexInfo.Parts {
				partCol = catalog.ResolveAlias(partCol)
				if partCol == oldColName {
					indexInfo.Parts[j] = newColName
					break
				}
			}
		}

		primaryKeyDef := alterPlan.CopyTableDef.Pkey
		for j, partCol := range primaryKeyDef.Names {
			if partCol == oldColName {
				primaryKeyDef.Names[j] = newColName
				break
			}
		}
		// handle cluster by key in modify column
		handleClusterByKey(ctx.GetContext(), alterPlan, newColName, oldColName)
	}

	for i, col := range tableDef.Cols {
		if strings.EqualFold(col.Name, oldColName) {
			colDef := DeepCopyColDef(col)
			colDef.Name = newColName
			colDef.OriginName = newColNameOrigin
			tableDef.Cols[i] = colDef
			break
		}
	}

	delete(alterCtx.alterColMap, oldColName)
	alterCtx.alterColMap[newColName] = selectExpr{
		sexprType: columnName,
		sexprStr:  oldColName,
	}

	if tmpCol, ok := alterCtx.changColDefMap[oldCol.ColId]; ok {
		tmpCol.Name = newColName
		tmpCol.OriginName = newColNameOrigin
	}

	alterCtx.UpdateSqls = append(alterCtx.UpdateSqls,
		getSqlForRenameColumn(alterPlan.Database,
			alterPlan.TableDef.Name,
			oldColNameOrigin,
			newColNameOrigin)...)

	return nil
}

// AlterColumn ALTER ... SET DEFAULT or ALTER ... DROP DEFAULT specify a new default value for a column or remove the old default value, respectively.
// If the old default is removed and the column can be NULL, the new default is NULL. If the column cannot be NULL, MySQL assigns a default value
func AlterColumn(ctx CompilerContext, alterPlan *plan.AlterTable, spec *tree.AlterTableAlterColumnClause, alterCtx *AlterTableContext) error {
	tableDef := alterPlan.CopyTableDef

	// get the original column name
	originalColName := spec.ColumnName.ColName()

	// Check whether original column has existed.
	originalCol := FindColumn(tableDef.Cols, originalColName)
	if originalCol == nil || originalCol.Hidden {
		return moerr.NewBadFieldError(ctx.GetContext(), spec.ColumnName.ColNameOrigin(), alterPlan.TableDef.Name)
	}

	for i, col := range tableDef.Cols {
		if strings.EqualFold(col.Name, originalCol.Name) {
			colDef := DeepCopyColDef(col)
			if spec.OptionType == tree.AlterColumnOptionSetDefault {
				tmpColumnDef := tree.NewColumnTableDef(spec.ColumnName, nil, []tree.ColumnAttribute{spec.DefaultExpr})
				defer func() {
					tmpColumnDef.Free()
				}()
				defaultValue, err := buildDefaultExpr(tmpColumnDef, colDef.Typ, ctx.GetProcess())
				if err != nil {
					return err
				}
				defaultValue.NullAbility = colDef.Default.NullAbility
				colDef.Default = defaultValue
			} else if spec.OptionType == tree.AlterColumnOptionDropDefault {
				colDef.Default.Expr = nil
				colDef.Default.OriginString = ""
			}
			tableDef.Cols[i] = colDef
			break
		}
	}
	return nil
}

// OrderByColumn Currently, Mo only performs semantic checks on alter table order by
// and does not implement the function of changing the physical storage order of data in the table
func OrderByColumn(ctx CompilerContext, alterPlan *plan.AlterTable, spec *tree.AlterTableOrderByColumnClause, alterCtx *AlterTableContext) error {
	tableDef := alterPlan.CopyTableDef
	for _, order := range spec.AlterOrderByList {
		// get the original column name
		originalColName := order.Column.ColName()
		// Check whether original column has existed.
		originalCol := FindColumn(tableDef.Cols, originalColName)
		if originalCol == nil || originalCol.Hidden {
			return moerr.NewBadFieldError(ctx.GetContext(), order.Column.ColNameOrigin(), alterPlan.TableDef.Name)
		}
	}
	return nil
}
