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
	"errors"
	"math"
	"sync"
	"time"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/common/rscthrottler"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/objectio"
)

// MemoryAdmissionNoWait prevents a transaction from waiting for workspace or
// locks retained by its own earlier statements. A nil operator preserves the
// isolated embedding/test contract.
func (proc *Process) MemoryAdmissionNoWait() bool {
	if proc == nil || proc.Base == nil || proc.GetTxnOperator() == nil {
		return false
	}
	options := proc.GetTxnOperator().TxnOptions()
	return options.ByBegin || !options.Autocommit
}

// ImportMemory is shared by the local pipelines of one execution. Admission
// happens before execution locks; the compiler closes its root reference after
// operator/job cleanup. Active leases keep their promises alive independently.
type ImportMemory struct {
	WorkingSet *rscthrottler.WorkingSet
	UnitBytes  int64
	// Initial layout bounds. Runtime owners read the monotonic WorkingSet
	// capacities after a wider record has grown the complete reservation.
	SourceBytes int64
	OutputBytes int64
	IOBytes     int64
	RecordBytes int64
	GrowRecord  func(context.Context, int64) (int64, error)
	insertFiles importInsertFiles
}

// Operators can reset while an output frame is still queued in another scope.
// The statement owns emitted files until MergeBlock records them in the txn.
type importInsertFiles struct {
	mu       sync.Mutex
	fs       fileservice.FileService
	pending  map[objectio.ObjectNameShort]struct{}
	capacity int64
	closed   bool
}

// Allow for map storage, proportionally unused slots and growth overlap.
const insertFileCreditBytes = 128

func (m *ImportMemory) TrackInsertFile(fs fileservice.FileService, name objectio.ObjectNameShort) error {
	files := &m.insertFiles
	files.mu.Lock()
	defer files.mu.Unlock()
	if files.closed {
		return rscthrottler.ErrWorkingSetClosed
	}
	if int64(len(files.pending)) == files.capacity {
		if files.capacity > math.MaxInt64/(2*insertFileCreditBytes) {
			return moerr.NewInternalErrorNoCtx("insert file ownership capacity overflow")
		}
		next := max(int64(16), 2*files.capacity)
		if err := m.WorkingSet.TryAcquireData((next - files.capacity) * insertFileCreditBytes); err != nil {
			return err
		}
		if files.pending == nil {
			files.pending = make(map[objectio.ObjectNameShort]struct{}, next)
			files.fs = fs
		}
		files.capacity = next
	}
	files.pending[name] = struct{}{}
	return nil
}

func (m *ImportMemory) AcceptInsertFile(name objectio.ObjectNameShort) {
	files := &m.insertFiles
	files.mu.Lock()
	delete(files.pending, name)
	files.mu.Unlock()
}

// CleanupInsertFiles runs after all pipeline readers and writers have joined.
// Its count also detects a successful execution with unregistered outputs.
func (m *ImportMemory) CleanupInsertFiles(ctx context.Context) (int, error) {
	files := &m.insertFiles
	files.mu.Lock()
	pending, fs, capacity := files.pending, files.fs, files.capacity
	files.pending, files.fs, files.capacity, files.closed = nil, nil, 0, true
	files.mu.Unlock()
	if capacity > 0 {
		defer m.WorkingSet.ReleaseData(capacity * insertFileCreditBytes)
	}
	if len(pending) == 0 {
		return 0, nil
	}
	cleanupCtx, cancel := context.WithTimeoutCause(context.WithoutCancel(ctx), 10*time.Second, moerr.CauseCleanUpUselessFiles)
	defer cancel()
	var firstErr error
	names := make([]string, 0, 128)
	for name := range pending {
		if err := cleanupCtx.Err(); err != nil {
			return len(pending), errors.Join(firstErr, context.Cause(cleanupCtx))
		}
		names = append(names, objectio.BuildObjectName(name.Segmentid(), name.Num()).String())
		if len(names) == cap(names) {
			if err := fileservice.DeleteBatchBestEffort(cleanupCtx, fs, names); firstErr == nil {
				firstErr = err
			}
			names = names[:0]
		}
	}
	return len(pending), errors.Join(firstErr, fileservice.DeleteBatchBestEffort(cleanupCtx, fs, names))
}

func (m *ImportMemory) AcquireInput(ctx context.Context) (*rscthrottler.InputLease, error) {
	if m == nil || m.WorkingSet == nil {
		return nil, nil
	}
	return m.WorkingSet.AcquireInput(ctx, 0)
}
func (m *ImportMemory) AcquireOutput(ctx context.Context) (*rscthrottler.InputLease, error) {
	if m == nil || m.WorkingSet == nil {
		return nil, nil
	}
	return m.WorkingSet.AcquireInput(ctx, 1)
}
func (m *ImportMemory) AcquireIO(ctx context.Context) (*rscthrottler.WorkspaceLease, error) {
	if m == nil || m.WorkingSet == nil {
		return nil, nil
	}
	return m.WorkingSet.AcquireIO(ctx, m.WorkingSet.IOBytes())
}
