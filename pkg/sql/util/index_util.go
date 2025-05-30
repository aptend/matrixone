// Copyright 2021 Matrix Origin
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

package util

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"github.com/matrixorigin/matrixone/pkg/catalog"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/container/vector"
	"github.com/matrixorigin/matrixone/pkg/sql/plan/function"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

var SerialWithCompacted = serialWithCompacted
var SerialWithoutCompacted = serialWithoutCompacted
var CompactSingleIndexCol = compactSingleIndexCol
var CompactPrimaryCol = compactPrimaryCol

type PackerList struct {
	ps []*types.Packer
}

func (list *PackerList) Free() {
	for _, p := range list.ps {
		if p != nil {
			p.Close()
		}
	}
	list.ps = nil
}

func (list *PackerList) PackerCount() int {
	return len(list.ps)
}

func BuildIndexTableName(ctx context.Context, unique bool) (string, error) {
	var name string
	if unique {
		name = catalog.UniqueIndexTableNamePrefix
	} else {
		name = catalog.SecondaryIndexTableNamePrefix
	}
	id, err := uuid.NewV7()
	if err != nil {
		return "", moerr.NewInternalError(ctx, "newuuid failed")
	}
	name += id.String()
	return name, nil
}

// IsIndexTableName checks if the given table name is an index table name with a valid UUID.
func IsIndexTableName(tableName string) bool {
	if catalog.IsUniqueIndexTable(tableName) {
		// Strip the prefix and check if the remaining part is a valid UUID
		uuidPart := strings.TrimPrefix(tableName, catalog.UniqueIndexTableNamePrefix)
		_, err := uuid.Parse(uuidPart)
		return err == nil
	} else if catalog.IsSecondaryIndexTable(tableName) {
		// Strip the prefix and check if the remaining part is a valid UUID
		uuidPart := strings.TrimPrefix(tableName, catalog.SecondaryIndexTableNamePrefix)
		_, err := uuid.Parse(uuidPart)
		return err == nil
	}
	return false // Not an index table name
}

// BuildUniqueKeyBatch used in test to validate
// serialWithCompacted(), compactSingleIndexCol() and compactPrimaryCol()
func BuildUniqueKeyBatch(
	vecs []*vector.Vector,
	attrs []string,
	parts []string,
	originTablePrimaryKey string,
	proc *process.Process,
	packers *PackerList,
) (*batch.Batch, int, error) {
	var (
		b               *batch.Batch
		err             error
		isCompoundIndex bool
		bitMap          *nulls.Nulls
	)

	if originTablePrimaryKey == "" {
		b = &batch.Batch{
			Attrs: make([]string, 1),
			Vecs:  make([]*vector.Vector, 1),
		}
		b.Attrs[0] = catalog.IndexTableIndexColName
	} else {
		b = &batch.Batch{
			Attrs: make([]string, 2),
			Vecs:  make([]*vector.Vector, 2),
		}
		b.Attrs[0] = catalog.IndexTableIndexColName
		b.Attrs[1] = catalog.IndexTablePrimaryColName
	}
	if len(parts) > 1 {
		isCompoundIndex = true
	}
	if isCompoundIndex {
		cIndexVecMap := make(map[string]*vector.Vector)
		for num, attrName := range attrs {
			for _, name := range parts {
				if attrName == name {
					cIndexVecMap[name] = vecs[num]
				}
			}
		}
		vs := make([]*vector.Vector, 0)
		for _, part := range parts {
			v := cIndexVecMap[part]
			vs = append(vs, v)
		}
		b.Vecs[0] = vector.NewVec(types.T_varchar.ToType())
		bitMap, err = serialWithCompacted(vs, b.Vecs[0], proc, packers)
	} else {
		var vec *vector.Vector
		for i, name := range attrs {
			if parts[0] == name {
				vec = vecs[i]
				break
			}
		}
		b.Vecs[0] = vector.NewVec(*vec.GetType())
		bitMap, err = compactSingleIndexCol(vec, b.Vecs[0], proc)
	}

	if len(b.Attrs) > 1 {
		var vec *vector.Vector
		for i, name := range attrs {
			if originTablePrimaryKey == name {
				vec = vecs[i]
			}
		}
		b.Vecs[1] = vector.NewVec(*vec.GetType())
		err = compactPrimaryCol(vec, nil, bitMap, proc)
	}

	if err != nil {
		b.Clean(proc.Mp())
		return nil, -1, err
	}
	b.SetRowCount(b.Vecs[0].Length())
	return b, b.RowCount(), nil
}

// SerialWithCompacted have a similar function named Serial
// SerialWithCompacted function is used by BuildUniqueKeyBatch
// when vs have null value, the function will ignore the row in
// the vs
// for example:
// input vec is [[1, 1, 1], [2, 2, null], [3, 3, 3]]
// result vec is [serial(1, 2, 3), serial(1, 2, 3)]
// result bitmap is [2]
func serialWithCompacted(
	vs []*vector.Vector,
	vec *vector.Vector,
	proc *process.Process,
	packers *PackerList,
) (*nulls.Nulls, error) {
	// resolve vs
	length := vs[0].Length()
	val := make([][]byte, 0, length)
	if length > cap(packers.ps) {
		for _, p := range packers.ps {
			if p != nil {
				p.Close()
			}
		}
		packers.ps = types.NewPackerArray(length)
	}
	defer func() {
		for i := 0; i < length; i++ {
			packers.ps[i].Reset()
		}
	}()
	bitMap := new(nulls.Nulls)

	ps := packers.ps
	for _, v := range vs {
		vNull := v.GetNulls()
		hasNull := v.HasNull()
		switch v.GetType().Oid {
		case types.T_bool:
			s := vector.MustFixedColNoTypeCheck[bool](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeBool(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeBool(b)
				}
			}
		case types.T_bit:
			s := vector.MustFixedColNoTypeCheck[uint64](v)
			for i, b := range s {
				if nulls.Contains(v.GetNulls(), uint64(i)) {
					nulls.Add(bitMap, uint64(i))
				} else {
					ps[i].EncodeUint64(b)
				}
			}
		case types.T_int8:
			s := vector.MustFixedColNoTypeCheck[int8](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeInt8(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeInt8(b)
				}
			}
		case types.T_int16:
			s := vector.MustFixedColNoTypeCheck[int16](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeInt16(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeInt16(b)
				}
			}
		case types.T_int32:
			s := vector.MustFixedColNoTypeCheck[int32](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeInt32(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeInt32(b)
				}
			}
		case types.T_int64:
			s := vector.MustFixedColNoTypeCheck[int64](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeInt64(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeInt64(b)
				}
			}
		case types.T_uint8:
			s := vector.MustFixedColNoTypeCheck[uint8](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeUint8(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeUint8(b)
				}
			}
		case types.T_uint16:
			s := vector.MustFixedColNoTypeCheck[uint16](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeUint16(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeUint16(b)
				}
			}
		case types.T_uint32:
			s := vector.MustFixedColNoTypeCheck[uint32](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeUint32(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeUint32(b)
				}
			}
		case types.T_uint64:
			s := vector.MustFixedColNoTypeCheck[uint64](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeUint64(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeUint64(b)
				}
			}
		case types.T_float32:
			s := vector.MustFixedColNoTypeCheck[float32](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeFloat32(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeFloat32(b)
				}
			}
		case types.T_float64:
			s := vector.MustFixedColNoTypeCheck[float64](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeFloat64(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeFloat64(b)
				}
			}
		case types.T_date:
			s := vector.MustFixedColNoTypeCheck[types.Date](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeDate(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeDate(b)
				}
			}
		case types.T_time:
			s := vector.MustFixedColNoTypeCheck[types.Time](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeTime(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeTime(b)
				}
			}
		case types.T_datetime:
			s := vector.MustFixedColNoTypeCheck[types.Datetime](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeDatetime(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeDatetime(b)
				}
			}
		case types.T_timestamp:
			s := vector.MustFixedColNoTypeCheck[types.Timestamp](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeTimestamp(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeTimestamp(b)
				}
			}
		case types.T_enum:
			s := vector.MustFixedColNoTypeCheck[types.Enum](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeEnum(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeEnum(b)
				}
			}
		case types.T_decimal64:
			s := vector.MustFixedColNoTypeCheck[types.Decimal64](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeDecimal64(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeDecimal64(b)
				}
			}
		case types.T_decimal128:
			s := vector.MustFixedColNoTypeCheck[types.Decimal128](v)
			if hasNull {
				for i, b := range s {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeDecimal128(b)
					}
				}
			} else {
				for i, b := range s {
					ps[i].EncodeDecimal128(b)
				}
			}
		case types.T_json, types.T_char, types.T_varchar, types.T_binary, types.T_varbinary, types.T_blob, types.T_text,
			types.T_array_float32, types.T_array_float64, types.T_datalink:
			// NOTE 1: We will consider T_array as bytes here just like JSON, VARBINARY and BLOB.
			// If not, we need to define arrayType in types/tuple.go as arrayF32TypeCode, arrayF64TypeCode etc
			// NOTE 2: vs is []string and not []byte. vs[i] is not of form "[1,2,3]". It is binary string of []float32{1,2,3}
			// NOTE 3: This class is mainly used by PreInsertUnique which gets triggered before inserting into column having
			// Unique Key or Primary Key constraint. Vector cannot be UK or PK.
			vs, area := vector.MustVarlenaRawData(v)
			if hasNull {
				for i := range vs {
					if nulls.Contains(vNull, uint64(i)) {
						nulls.Add(bitMap, uint64(i))
					} else {
						ps[i].EncodeStringType(vs[i].GetByteSlice(area))
					}
				}
			} else {
				for i := range vs {
					ps[i].EncodeStringType(vs[i].GetByteSlice(area))
				}
			}
		}
	}

	for i := 0; i < length; i++ {
		if !nulls.Contains(bitMap, uint64(i)) {
			val = append(val, ps[i].GetBuf())
		}
	}

	err := vector.AppendBytesList(vec, val, nil, proc.Mp())

	return bitMap, err
}

// serialWithoutCompacted is similar to serialWithCompacted and builtInSerial
// serialWithoutCompacted function is used by Secondary Index to support rows containing null entries
// for example:
// input vec is [[1, 1, 1], [2, 2, null], [3, 3, 3]]
// result vec is [serial(1, 2, 3), serial(1, 2, null), serial(1, 2, 3)]
// result bitmap is [] (empty)
// Here we are keeping the same function signature of serialWithCompacted so that we can duplicate the same code of
// `preinsertunique` in `preinsertsecondaryindex`
func serialWithoutCompacted(
	vs []*vector.Vector,
	vec *vector.Vector,
	proc *process.Process,
	packers *PackerList,
) (*nulls.Nulls, error) {
	if len(vs) == 0 {
		// return empty bitmap
		return new(nulls.Nulls), nil
	}

	rowCount := vs[0].Length()
	if rowCount > cap(packers.ps) {
		for _, p := range packers.ps {
			if p != nil {
				p.Close()
			}
		}
		packers.ps = types.NewPackerArray(rowCount)
	}
	defer func() {
		for i := 0; i < rowCount; i++ {
			packers.ps[i].Reset()
		}
	}()

	ps := packers.ps
	for _, v := range vs {
		if v.IsConstNull() {
			for i := 0; i < v.Length(); i++ {
				ps[i].EncodeNull()
			}
			continue
		}
		function.SerialHelper(v, nil, ps, true)
	}

	for i := 0; i < rowCount; i++ {
		if err := vector.AppendBytes(vec, ps[i].GetBuf(), false, proc.Mp()); err != nil {
			return nil, err
		}
	}

	return new(nulls.Nulls), nil
}

func compactSingleIndexCol(
	v *vector.Vector,
	vec *vector.Vector,
	proc *process.Process,
) (*nulls.Nulls, error) {
	var err error

	hasNull := v.HasNull()
	if !hasNull {
		err = vector.GetUnionAllFunction(*v.GetType(), proc.GetMPool())(vec, v)
		return v.GetNulls(), err
	}
	length := v.Length()
	switch v.GetType().Oid {
	case types.T_bool:
		s := vector.MustFixedColNoTypeCheck[bool](v)
		ns := make([]bool, 0, length)
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_bit:
		s := vector.MustFixedColNoTypeCheck[uint64](v)
		ns := make([]uint64, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_int8:
		s := vector.MustFixedColNoTypeCheck[int8](v)
		ns := make([]int8, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_int16:
		s := vector.MustFixedColNoTypeCheck[int16](v)
		ns := make([]int16, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_int32:
		s := vector.MustFixedColNoTypeCheck[int32](v)
		ns := make([]int32, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_int64:
		s := vector.MustFixedColNoTypeCheck[int64](v)
		ns := make([]int64, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_uint8:
		s := vector.MustFixedColNoTypeCheck[uint8](v)
		ns := make([]uint8, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_uint16:
		s := vector.MustFixedColNoTypeCheck[uint16](v)
		ns := make([]uint16, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_uint32:
		s := vector.MustFixedColNoTypeCheck[uint32](v)
		ns := make([]uint32, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		vec = vector.NewVec(*v.GetType())
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_uint64:
		s := vector.MustFixedColNoTypeCheck[uint64](v)
		ns := make([]uint64, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_float32:
		s := vector.MustFixedColNoTypeCheck[float32](v)
		ns := make([]float32, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_float64:
		s := vector.MustFixedColNoTypeCheck[float64](v)
		ns := make([]float64, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		vec = vector.NewVec(*v.GetType())
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_date:
		s := vector.MustFixedColNoTypeCheck[types.Date](v)
		ns := make([]types.Date, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		vec = vector.NewVec(*v.GetType())
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_time:
		s := vector.MustFixedColNoTypeCheck[types.Time](v)
		ns := make([]types.Time, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_datetime:
		s := vector.MustFixedColNoTypeCheck[types.Datetime](v)
		ns := make([]types.Datetime, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_timestamp:
		s := vector.MustFixedColNoTypeCheck[types.Timestamp](v)
		ns := make([]types.Timestamp, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_enum:
		s := vector.MustFixedColNoTypeCheck[types.Enum](v)
		ns := make([]types.Enum, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_decimal64:
		s := vector.MustFixedColNoTypeCheck[types.Decimal64](v)
		ns := make([]types.Decimal64, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_decimal128:
		s := vector.MustFixedColNoTypeCheck[types.Decimal128](v)
		ns := make([]types.Decimal128, 0, len(s))
		for i, b := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_json, types.T_char, types.T_varchar, types.T_binary, types.T_varbinary, types.T_blob,
		types.T_array_float32, types.T_array_float64:
		s, area := vector.MustVarlenaRawData(v)
		ns := make([][]byte, 0, len(s))
		for i := range s {
			if !nulls.Contains(v.GetNulls(), uint64(i)) {
				ns = append(ns, s[i].GetByteSlice(area))
			}
		}
		err = vector.AppendBytesList(vec, ns, nil, proc.Mp())
	}
	return v.GetNulls(), err
}

func compactPrimaryCol(
	v *vector.Vector,
	vec *vector.Vector,
	bitMap *nulls.Nulls,
	proc *process.Process,
) error {
	var err error

	if bitMap.IsEmpty() {
		err = vector.GetUnionAllFunction(*v.GetType(), proc.GetMPool())(vec, v)
		return err
	}
	length := v.Length()
	switch v.GetType().Oid {
	case types.T_bool:
		s := vector.MustFixedColNoTypeCheck[bool](v)
		ns := make([]bool, 0, length)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_bit:
		s := vector.MustFixedColNoTypeCheck[uint64](v)
		ns := make([]uint64, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_int8:
		s := vector.MustFixedColNoTypeCheck[int8](v)
		ns := make([]int8, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_int16:
		s := vector.MustFixedColNoTypeCheck[int16](v)
		ns := make([]int16, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_int32:
		s := vector.MustFixedColNoTypeCheck[int32](v)
		ns := make([]int32, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_int64:
		s := vector.MustFixedColNoTypeCheck[int64](v)
		ns := make([]int64, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_uint8:
		s := vector.MustFixedColNoTypeCheck[uint8](v)
		ns := make([]uint8, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_uint16:
		s := vector.MustFixedColNoTypeCheck[uint16](v)
		ns := make([]uint16, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_uint32:
		s := vector.MustFixedColNoTypeCheck[uint32](v)
		ns := make([]uint32, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_uint64:
		s := vector.MustFixedColNoTypeCheck[uint64](v)
		ns := make([]uint64, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_float32:
		s := vector.MustFixedColNoTypeCheck[float32](v)
		ns := make([]float32, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_float64:
		s := vector.MustFixedColNoTypeCheck[float64](v)
		ns := make([]float64, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_date:
		s := vector.MustFixedColNoTypeCheck[types.Date](v)
		ns := make([]types.Date, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_time:
		s := vector.MustFixedColNoTypeCheck[types.Time](v)
		ns := make([]types.Time, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_datetime:
		s := vector.MustFixedColNoTypeCheck[types.Datetime](v)
		ns := make([]types.Datetime, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_timestamp:
		s := vector.MustFixedColNoTypeCheck[types.Timestamp](v)
		ns := make([]types.Timestamp, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_enum:
		s := vector.MustFixedColNoTypeCheck[types.Enum](v)
		ns := make([]types.Enum, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_decimal64:
		s := vector.MustFixedColNoTypeCheck[types.Decimal64](v)
		ns := make([]types.Decimal64, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_decimal128:
		s := vector.MustFixedColNoTypeCheck[types.Decimal128](v)
		ns := make([]types.Decimal128, 0)
		for i, b := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, b)
			}
		}
		err = vector.AppendFixedList(vec, ns, nil, proc.Mp())
	case types.T_json, types.T_char, types.T_varchar, types.T_binary, types.T_varbinary, types.T_blob,
		types.T_array_float32, types.T_array_float64:
		s, area := vector.MustVarlenaRawData(v)
		ns := make([][]byte, 0, len(s))
		for i := range s {
			if !nulls.Contains(bitMap, uint64(i)) {
				ns = append(ns, s[i].GetByteSlice(area))
			}
		}
		err = vector.AppendBytesList(vec, ns, nil, proc.Mp())
	}
	return err
}
