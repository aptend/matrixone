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

package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/google/uuid"
)

/*

[SegUUID-16bytes] [ObjectOffset-2bytes] [BlockOffset-2bytes] [RowOffset-4bytes]
--------------------------------------- s3 file name
------------------------------------------------------------ block id
------------------------------------------------------------------------------ rowid

*/

const ObjectBytesSize = 18

type ObjectBytes = [ObjectBytesSize]byte

func (r Rowid) Less(than Rowid) bool {
	return bytes.Compare(r[:], than[:]) < 0
}

func (r Rowid) Equal(to Rowid) bool {
	return bytes.Equal(r[:], to[:])
}

func (r Rowid) GetBlockid() Blockid {
	return *(*Blockid)(r[:BlockidSize])
	// return (Blockid)(r[:BlockidSize])
}

func (r Rowid) GetSegid() Uuid {
	return *(*Uuid)(r[:UuidSize])
	// return (Blockid)(r[:BlockidSize])
}

func (r *Rowid) GetBlockidUnsafe() *Blockid {
	return (*Blockid)(unsafe.Pointer(&r[0]))
}

func (r Rowid) GetObject() ObjectBytes {
	return *(*ObjectBytes)(r[:ObjectBytesSize])
}

func (r Rowid) GetObjectString() string {
	uuid := (*uuid.UUID)(r[:16])
	s := binary.BigEndian.Uint16(r[16:18])
	return fmt.Sprintf("%s-%d", uuid.String(), s)
}

func RandomRowid() Rowid {
	var r Rowid
	u := uuid.New()
	copy(r[:], u[:])
	return r
}

func (b Blockid) Compare(other Blockid) int {
	return bytes.Compare(b[:], other[:])
}

func (b *Blockid) IsEmpty() bool {
	for _, v := range b[:] {
		if v != 0 {
			return false
		}
	}
	return true
}

func (b *Blockid) String() string {
	uuid := (*uuid.UUID)(b[:16])
	filen, blkn := b.Offsets()
	return fmt.Sprintf("%s-%d-%d", uuid.String(), filen, blkn)
}

func (b *Blockid) ShortString() string {
	filen, blkn := b.Offsets()
	return fmt.Sprintf("%d-%d", filen, blkn)
}

func (b *Blockid) Offsets() (uint16, uint16) {
	filen := binary.BigEndian.Uint16(b[16:18])
	blkn := binary.BigEndian.Uint16(b[18:20])
	return filen, blkn
}
