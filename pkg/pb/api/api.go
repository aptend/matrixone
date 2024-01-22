// Copyright 2021 - 2022 Matrix Origin
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
package api

import (
	fmt "fmt"

	"github.com/matrixorigin/matrixone/pkg/container/nulls"
	"github.com/matrixorigin/matrixone/pkg/pb/plan"
)

var (
	OpMethodName = map[OpCode]string{
		OpCode_OpPing:          "Ping",
		OpCode_OpFlush:         "Flush",
		OpCode_OpCheckpoint:    "Checkpoint",
		OpCode_OpInspect:       "Inspect",
		OpCode_OpAddFaultPoint: "AddFaultPoint",
		OpCode_OpBackup:        "Backup",
		OpCode_OpTraceSpan:     "TraceSpan",
	}
)

func NewUpdatePolicyReq(minRowQ, maxObjOnerune, maxRowsMerged int, hints ...MergeHint) *AlterTableReq {
	return &AlterTableReq{
		Kind: AlterKind_UpdatePolicy,
		Operation: &AlterTableReq_UpdatePolicy{
			&AlterTablePolicy{
				MinRowsQuailifed: uint32(minRowQ),
				MaxObjOnerun:     uint32(maxObjOnerune),
				MaxRowsMergedObj: uint32(maxRowsMerged),
				Hints:            hints,
			},
		},
	}
}

func NewUpdateConstraintReq(did, tid uint64, cstr string) *AlterTableReq {
	return &AlterTableReq{
		DbId:    did,
		TableId: tid,
		Kind:    AlterKind_UpdateConstraint,
		Operation: &AlterTableReq_UpdateCstr{
			&AlterTableConstraint{Constraints: []byte(cstr)},
		},
	}
}

func NewUpdateCommentReq(did, tid uint64, comment string) *AlterTableReq {
	return &AlterTableReq{
		DbId:    did,
		TableId: tid,
		Kind:    AlterKind_UpdateComment,
		Operation: &AlterTableReq_UpdateComment{
			&AlterTableComment{Comment: comment},
		},
	}
}

func NewRenameTableReq(did, tid uint64, old, new string) *AlterTableReq {
	return &AlterTableReq{
		DbId:    did,
		TableId: tid,
		Kind:    AlterKind_RenameTable,
		Operation: &AlterTableReq_RenameTable{
			&AlterTableRenameTable{OldName: old, NewName: new},
		},
	}
}

func NewAddColumnReq(did, tid uint64, name string, typ *plan.Type, insertAt int32) *AlterTableReq {
	return &AlterTableReq{
		DbId:    did,
		TableId: tid,
		Kind:    AlterKind_AddColumn,
		Operation: &AlterTableReq_AddColumn{
			&AlterTableAddColumn{
				Column: &plan.ColDef{
					Name: name,
					Typ:  typ,
					Default: &plan.Default{
						NullAbility:  true,
						Expr:         nil,
						OriginString: "",
					},
				},
				InsertPosition: insertAt,
			},
		},
	}
}

func NewRemoveColumnReq(did, tid uint64, idx, seqnum uint32) *AlterTableReq {
	return &AlterTableReq{
		DbId:    did,
		TableId: tid,
		Kind:    AlterKind_DropColumn,
		Operation: &AlterTableReq_DropColumn{
			&AlterTableDropColumn{
				LogicalIdx:  idx,
				SequenceNum: seqnum,
			},
		},
	}
}

func NewAddPartitionReq(did, tid uint64, partitionDef *plan.PartitionByDef) *AlterTableReq {
	return &AlterTableReq{
		DbId:    did,
		TableId: tid,
		Kind:    AlterKind_AddPartition,
		Operation: &AlterTableReq_AddPartition{
			AddPartition: &AlterTableAddPartition{
				PartitionDef: partitionDef,
			},
		},
	}
}

func NewRenameColumnReq(did, tid uint64, oldname, newname string, seqnum uint32) *AlterTableReq {
	return &AlterTableReq{
		DbId:    did,
		TableId: tid,
		Kind:    AlterKind_RenameColumn,
		Operation: &AlterTableReq_RenameCol{
			&AlterTableRenameCol{
				OldName:     oldname,
				NewName:     newname,
				SequenceNum: seqnum,
			},
		},
	}
}
func (m *SyncLogTailReq) MarshalBinary() ([]byte, error) {
	return m.Marshal()
}

func (m *SyncLogTailReq) UnmarshalBinary(data []byte) error {
	return m.Unmarshal(data)
}

func (m *SyncLogTailResp) MarshalBinary() ([]byte, error) {
	return m.Marshal()
}

func (m *SyncLogTailResp) UnmarshalBinary(data []byte) error {
	return m.Unmarshal(data)
}

func (m *PrecommitWriteCmd) MarshalBinary() ([]byte, error) {
	return m.Marshal()
}

func (m *PrecommitWriteCmd) UnmarshalBinary(data []byte) error {
	return m.Unmarshal(data)
}

func NewBlkTransferBooking(size int) *BlkTransferBooking {
	mappings := make([]BlkTransMap, size)
	for i := 0; i < size; i++ {
		mappings[i] = BlkTransMap{
			M: make(map[int32]TransDestPos),
		}
	}
	return &BlkTransferBooking{
		Mappings: mappings,
	}
}

func (b *BlkTransferBooking) Clean() {
	for i := 0; i < len(b.Mappings); i++ {
		b.Mappings[i] = BlkTransMap{
			M: make(map[int32]TransDestPos),
		}
	}
}
func (b *BlkTransferBooking) AddSortPhaseMapping(idx int, originRowCnt int, deletes *nulls.Nulls, mapping []int32) {
	// TODO: remove panic check
	if mapping != nil {
		deletecnt := 0
		if deletes != nil {
			deletecnt = deletes.GetCardinality()
		}
		if len(mapping) != originRowCnt-deletecnt {
			panic(fmt.Sprintf("mapping length %d != originRowCnt %d - deletes %s", len(mapping), originRowCnt, deletes))
		}
		// mapping sortedVec[i] = originalVec[sortMapping[i]]
		// transpose it, originalVec[sortMapping[i]] = sortedVec[i]
		// [9 4 8 5 2 6 0 7 3 1](orignVec)  -> [6 9 4 8 1 3 5 7 2 0](sortedVec)
		// [0 1 2 3 4 5 6 7 8 9](sortedVec) -> [0 1 2 3 4 5 6 7 8 9](originalVec)
		// TODO: use a more efficient way to transpose, in place
		transposedMapping := make([]int32, len(mapping))
		for sortedPos, originalPos := range mapping {
			transposedMapping[originalPos] = int32(sortedPos)
		}
		mapping = transposedMapping
	}
	posInVecApplyDeletes := 0
	targetMapping := b.Mappings[idx].M
	for origRow := 0; origRow < originRowCnt; origRow++ {
		if deletes != nil && deletes.Contains(uint64(origRow)) {
			// this row has been deleted, skip its mapping
			continue
		}
		if mapping == nil {
			// no sort phase, the mapping is 1:1, just use posInVecApplyDeletes
			targetMapping[int32(origRow)] = TransDestPos{Idx: -1, Row: int32(posInVecApplyDeletes)}
		} else {
			targetMapping[int32(origRow)] = TransDestPos{Idx: -1, Row: mapping[posInVecApplyDeletes]}
		}
		posInVecApplyDeletes++
	}
}

func (b *BlkTransferBooking) UpdateMappingAfterMerge(mapping, fromLayout, toLayout []uint32) {
	bisectHaystack := make([]uint32, 0, len(toLayout)+1)
	bisectHaystack = append(bisectHaystack, 0)
	for _, x := range toLayout {
		bisectHaystack = append(bisectHaystack, bisectHaystack[len(bisectHaystack)-1]+x)
	}

	// given toLayout and a needle, find its corresponding block index and row index in the block
	// For example, toLayout [8192, 8192, 1024], needle = 0 -> (0, 0); needle = 8192 -> (1, 0); needle = 8193 -> (1, 1)
	bisectPinpoint := func(needle uint32) (int, uint32) {
		i, j := 0, len(bisectHaystack)
		for i < j {
			m := (i + j) / 2
			if bisectHaystack[m] > needle {
				j = m
			} else {
				i = m + 1
			}
		}
		// bisectHaystack[i] is the first number > needle, so the needle falls into i-1 th block
		blkIdx := i - 1
		rows := needle - bisectHaystack[blkIdx]
		return blkIdx, rows
	}

	var totalHandledRows int32

	for _, mcontainer := range b.Mappings {
		m := mcontainer.M
		var curTotal int32   // index in the flatten src array
		var destTotal uint32 // index in the flatten merged array
		for srcRow := range m {
			curTotal = totalHandledRows + m[srcRow].Row
			if mapping == nil {
				destTotal = uint32(curTotal)
			} else {
				destTotal = mapping[curTotal]
			}
			destBlkIdx, destRowIdx := bisectPinpoint(destTotal)
			m[srcRow] = TransDestPos{Idx: int32(destBlkIdx), Row: int32(destRowIdx)}
		}
		totalHandledRows += int32(len(m))
	}
}
