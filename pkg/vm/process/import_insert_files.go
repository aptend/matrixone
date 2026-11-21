// Copyright 2026 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package process

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/common/rscthrottler"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/objectio"
)

// InsertFileOwnerBytes is charged to each controlled writer's admitted output
// slot. Larger sets of object names use the same disk and FD budget as metadata.
const InsertFileOwnerBytes = 2048

const insertOutputReceipt = uint64(0x3e034d4f494e5301)

type InsertFileJournal interface {
	Reserve(uint64) error
	Write([]byte) (int, error)
	ReadAt([]byte, int64) (int, error)
	Close() error
}

type importInsertFiles struct {
	mu     sync.Mutex
	owners []*InsertFileOwner
	closed bool
}

// InsertFileOwner keeps the identities of uploads separate from the serialized
// stats. A damaged stats spool cannot erase the identities needed to abort it.
// Each producer sends complete object frames to one FIFO pipeline edge.
type InsertFileOwner struct {
	mu         sync.Mutex
	id         uint64
	fs         fileservice.FileService
	inline     [64]objectio.ObjectNameShort
	journal    InsertFileJournal
	count      int64
	published  int64
	accepted   int64
	accepting  bool
	cleaned    int64
	unrecorded [][]objectio.ObjectStats
}

func (m *ImportMemory) NewInsertFileOwner(fs fileservice.FileService) (*InsertFileOwner, error) {
	files := &m.insertFiles
	files.mu.Lock()
	defer files.mu.Unlock()
	if files.closed {
		return nil, rscthrottler.ErrWorkingSetClosed
	}
	owner := &InsertFileOwner{id: uint64(len(files.owners) + 1), fs: fs}
	files.owners = append(files.owners, owner)
	return owner, nil
}

// Retain transfers a bounded batch of completed results. On a journal error
// the original, still-owned suffix is retained for cleanup without copying it.
func (o *InsertFileOwner) Retain(stats []objectio.ObjectStats, open func() (InsertFileJournal, error)) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	for i := range stats {
		if err := o.record(*stats[i].ObjectName().Short(), open); err != nil {
			o.unrecorded = append(o.unrecorded, stats[i:])
			return err
		}
	}
	return nil
}

func (o *InsertFileOwner) record(name objectio.ObjectNameShort, open func() (InsertFileJournal, error)) error {
	if o.count < int64(len(o.inline)) {
		o.inline[o.count] = name
	} else {
		if o.journal == nil {
			journal, err := open()
			if err != nil {
				return err
			}
			o.journal = journal
		}
		if err := o.journal.Reserve(uint64(len(name))); err != nil {
			return err
		}
		n, err := o.journal.Write(name[:])
		if err != nil {
			return err
		}
		if n != len(name) {
			return io.ErrShortWrite
		}
	}
	o.count++
	return nil
}

func (o *InsertFileOwner) RetainUnpublished(stats ...[]objectio.ObjectStats) {
	o.mu.Lock()
	defer o.mu.Unlock()
	for _, objects := range stats {
		if len(objects) > 0 {
			o.unrecorded = append(o.unrecorded, objects)
		}
	}
}

func (o *InsertFileOwner) Publish(name objectio.ObjectNameShort) ([]byte, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.published >= o.count {
		return nil, moerr.NewInternalErrorNoCtx("untracked controlled INSERT object")
	}
	expected, err := o.readName(o.published)
	if err != nil {
		return nil, err
	}
	if expected != name {
		return nil, moerr.NewInternalErrorNoCtx("controlled INSERT metadata does not match its uploaded object")
	}
	var receipt [24]byte
	binary.LittleEndian.PutUint64(receipt[:8], insertOutputReceipt)
	binary.LittleEndian.PutUint64(receipt[8:16], o.id)
	binary.LittleEndian.PutUint64(receipt[16:], uint64(o.published))
	o.published++
	return receipt[:], nil
}

// WriteInsertOutput validates the receipt before calling Relation.Write. The
// only transition after successful workspace registration is an infallible
// cursor advance. Different producers can reach the merger in any order.
func (m *ImportMemory) WriteInsertOutput(receipt []byte, write func() error) error {
	if len(receipt) != 24 || binary.LittleEndian.Uint64(receipt[:8]) != insertOutputReceipt {
		return write()
	}
	id, index := binary.LittleEndian.Uint64(receipt[8:16]), binary.LittleEndian.Uint64(receipt[16:])
	files := &m.insertFiles
	files.mu.Lock()
	if files.closed || id == 0 || id > uint64(len(files.owners)) {
		files.mu.Unlock()
		return moerr.NewInternalErrorNoCtx("invalid controlled INSERT producer")
	}
	owner := files.owners[id-1]
	files.mu.Unlock()

	owner.mu.Lock()
	if index != uint64(owner.accepted) || owner.accepted >= owner.published || owner.accepting {
		owner.mu.Unlock()
		return moerr.NewInternalErrorNoCtx("out-of-order controlled INSERT object")
	}
	owner.accepting = true
	owner.mu.Unlock()
	accepted := false
	defer func() {
		owner.mu.Lock()
		if accepted {
			owner.accepted++
		}
		owner.accepting = false
		owner.mu.Unlock()
	}()
	err := write()
	accepted = err == nil
	return err
}

// CleanupUnpublished may run during producer reset while a downstream owner
// is still processing a frame. Only the suffix that was never sent is eligible.
func (o *InsertFileOwner) CleanupUnpublished(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	cursor := o.published
	err := o.deleteRange(ctx, &cursor, o.count)
	if err == nil {
		o.count = o.published
	}
	return errors.Join(err, o.deleteUnrecorded(ctx))
}

func (o *InsertFileOwner) readName(index int64) (name objectio.ObjectNameShort, err error) {
	if index < int64(len(o.inline)) {
		return o.inline[index], nil
	}
	_, err = o.journal.ReadAt(name[:], (index-int64(len(o.inline)))*int64(len(name)))
	return name, err
}

func (o *InsertFileOwner) deleteRange(ctx context.Context, cursor *int64, end int64) error {
	var firstErr error
	for index := *cursor; index < end && ctx.Err() == nil; {
		var names [64]string
		n := 0
		var readErr error
		for i := index; i < min(end, index+int64(len(names))); i++ {
			var name objectio.ObjectNameShort
			name, readErr = o.readName(i)
			if readErr != nil {
				break
			}
			names[n] = objectio.BuildObjectName(name.Segmentid(), name.Num()).String()
			n++
		}
		if err := fileservice.DeleteBatchBestEffort(ctx, o.fs, names[:n]); firstErr == nil {
			firstErr = err
		}
		index += int64(n)
		if firstErr == nil {
			*cursor = index
		}
		if readErr != nil {
			return errors.Join(firstErr, readErr)
		}
	}
	return errors.Join(firstErr, context.Cause(ctx))
}

func (o *InsertFileOwner) deleteUnrecorded(ctx context.Context) error {
	for len(o.unrecorded) > 0 {
		objects := o.unrecorded[0]
		var names [64]string
		n := min(len(objects), len(names))
		for i := range n {
			names[i] = objects[i].ObjectName().String()
		}
		if err := fileservice.DeleteBatchBestEffort(ctx, o.fs, names[:n]); err != nil {
			return err
		}
		if n < len(objects) {
			o.unrecorded[0] = objects[n:]
		} else {
			o.unrecorded[0] = nil
			o.unrecorded = o.unrecorded[1:]
		}
	}
	return nil
}

// CleanupInsertFiles is called after all pipelines join. It retires only
// unaccepted objects and retains the failed suffix for a later attempt.
func (m *ImportMemory) CleanupInsertFiles(ctx context.Context) (pending int, err error) {
	files := &m.insertFiles
	files.mu.Lock()
	defer files.mu.Unlock()
	files.closed = true
	cleanupCtx, cancel := context.WithTimeoutCause(context.WithoutCancel(ctx), 10*time.Second, moerr.CauseCleanUpUselessFiles)
	defer cancel()
	for _, owner := range files.owners {
		owner.mu.Lock()
		owner.cleaned = max(owner.cleaned, owner.accepted)
		pending += int(owner.count - owner.cleaned)
		for _, stats := range owner.unrecorded {
			pending += len(stats)
		}
		deleteErr := errors.Join(owner.deleteRange(cleanupCtx, &owner.cleaned, owner.count), owner.deleteUnrecorded(cleanupCtx))
		if deleteErr == nil && owner.journal != nil {
			deleteErr = owner.journal.Close()
			owner.journal = nil
		}
		owner.mu.Unlock()
		err = errors.Join(err, deleteErr)
	}
	if err == nil {
		files.owners = nil
	}
	return pending, err
}

// FinishInsertFileCleanup gives the terminal statement path a second attempt
// against the same retained owners. A persistent storage error is surfaced
// while all CN-local spool resources and the admission ticket can be released.
func (m *ImportMemory) FinishInsertFileCleanup(ctx context.Context) (int, error) {
	pending, err := m.CleanupInsertFiles(ctx)
	if err == nil {
		return pending, nil
	}
	if _, err = m.CleanupInsertFiles(ctx); err == nil {
		return pending, nil
	}
	files := &m.insertFiles
	files.mu.Lock()
	defer files.mu.Unlock()
	for _, owner := range files.owners {
		if owner.journal != nil {
			err = errors.Join(err, owner.journal.Close())
			owner.journal = nil
		}
	}
	files.owners = nil
	return pending, err
}
