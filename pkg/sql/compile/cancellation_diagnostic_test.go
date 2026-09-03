// Copyright 2026 Matrix Origin
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

package compile

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/matrixorigin/matrixone/pkg/perfcounter"
	"github.com/matrixorigin/matrixone/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestWrapRunSQLTrackerCancelDelegatesForInternalExecutor(t *testing.T) {
	proc := testutil.NewProcess(t)
	proc.ReplaceTopCtx(perfcounter.AttachTxnExecutorKey(context.Background()))
	queryCtx := proc.Base.GetContextBase().BuildQueryCtx(proc.GetTopContext())
	proc.BuildPipelineContext(queryCtx)
	c := &Compile{proc: proc}

	var token atomic.Uint64
	token.Store(27261)
	var calls atomic.Int32
	wrapped := wrapRunSQLTrackerCancel(c, func() {
		calls.Add(1)
	}, "select 1", &token)

	wrapped()
	require.Equal(t, int32(1), calls.Load())
}

func TestWrapRunSQLTrackerCancelSkipsOrdinaryExecutor(t *testing.T) {
	proc := testutil.NewProcess(t)
	c := &Compile{proc: proc}

	var token atomic.Uint64
	var calls atomic.Int32
	original := func() {
		calls.Add(1)
	}
	wrapper := wrapRunSQLTrackerCancel(c, original, "select 1", &token)

	wrapper()
	require.Equal(t, int32(1), calls.Load())
}
