// Copyright 2021 - 2022 Matrix Origin
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

package ctl

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

// parameter should be "DbName.TableName obj1,obj2,obj3..."
func parseArg(parameter string) (db, tbl string, targets []objectio.ObjectStats, err error) {
	parameters := strings.Split(parameter, " ")
	if len(parameters) != 2 {
		err = moerr.NewInternalErrorNoCtx("handleMerge: invalid parameter")
		return
	}
	address := strings.Split(parameters[0], ".")
	if len(address) != 2 {
		err = moerr.NewInternalErrorNoCtx("handleMerge: invalid parameter")
		return
	}
	db, tbl = address[0], address[1]
	objstrs := strings.Split(parameters[1], ",")
	for _, objstrs := range objstrs {
		parts := strings.Split(objstrs, "_")
		uuid, err := types.ParseUuid(parts[0])
		if err != nil {
			return "", "", nil, err
		}
		num, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", "", nil, err
		}
		objectname := objectio.BuildObjectName(&uuid, uint16(num))

		obj := objectio.NewObjectStats()
		objectio.SetObjectStatsObjectName(obj, objectname)
		targets = append(targets, *obj)
	}
	return
}

func handleMerge(
	proc *process.Process,
	st serviceType,
	parameter string,
	sender requestSender) (Result, error) {
	txnOp := proc.TxnOperator
	if proc.TxnOperator == nil {
		return Result{}, moerr.NewInternalError(proc.Ctx, "handleFlush: txn operator is nil")
	}

	db, tbl, targets, err := parseArg(parameter)
	if err != nil {
		return Result{}, err
	}
	database, err := proc.SessionInfo.StorageEngine.Database(proc.Ctx, db, txnOp)
	if err != nil {
		return Result{}, err
	}
	rel, err := database.Relation(proc.Ctx, tbl, nil)
	if err != nil {
		return Result{}, err
	}
	createdObjs, err := rel.MergeObjects(proc.Ctx, targets)
	if err != nil {
		return Result{}, err
	}

	var dataBuilder strings.Builder
	for _, o := range createdObjs {
		stat := objectio.ObjectStats(o)
		detail := fmt.Sprintf("%v, rows %v, blks %v, osize %v, csize %v\n",
			stat.ObjectName().String(), stat.Rows(), stat.BlkCnt(),
			common.HumanReadableBytes(int(stat.OriginSize())),
			common.HumanReadableBytes(int(stat.Size())))
		dataBuilder.WriteString(detail)
	}
	return Result{
		Method: MergeObjectsMethod,
		Data:   dataBuilder.String(),
	}, nil
}