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
	"runtime/debug"
	"sync/atomic"

	commonutil "github.com/matrixorigin/matrixone/pkg/common/util"
	"github.com/matrixorigin/matrixone/pkg/perfcounter"
	txnutil "github.com/matrixorigin/matrixone/pkg/txn/util"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
	"go.uber.org/zap"
)

const cancellationDiagnosticLog = "internal SQL cancellation diagnostic"

const (
	internalSQLCancellationStageKey = "internal_sql_cancellation_stage"
	internalSQLTrackerCancelKey     = "internal_sql_tracker_cancel"
)

func isInternalCancellationDiagnostic(proc *process.Process, err error) bool {
	return proc != nil && proc.GetTopContext() != nil &&
		isScopeCancellationError(err) &&
		perfcounter.IsInternalExecutor(proc.GetTopContext())
}

func cancellationContextFields(prefix string, ctx context.Context) []zap.Field {
	if ctx == nil {
		return []zap.Field{zap.Bool(prefix+"-present", false)}
	}
	return []zap.Field{
		zap.Bool(prefix+"-present", true),
		zap.NamedError(prefix+"-err", ctx.Err()),
		zap.NamedError(prefix+"-cause", context.Cause(ctx)),
		zap.Uint64(prefix+"-pipeline-id", process.PipelineContextDiagnosticID(ctx)),
		zap.String(prefix+"-pipeline-cancel", process.PipelineCancellationDiagnostic(ctx)),
	}
}

func logInternalCancellationDiagnostic(
	proc *process.Process,
	stage string,
	err error,
	pipelineCtx context.Context,
	queryCtx context.Context,
	topCtx context.Context,
	extra ...zap.Field,
) {
	if !isInternalCancellationDiagnostic(proc, err) {
		return
	}
	fields := make([]zap.Field, 0, 18+len(extra))
	fields = append(fields,
		zap.String("stage", stage),
		zap.NamedError("error", err),
		zap.String("query-id", proc.QueryId()),
	)
	if txnOperator := proc.GetTxnOperator(); txnOperator != nil {
		fields = append(fields, txnutil.TxnIDField(txnOperator.Txn()))
	}
	fields = append(fields, cancellationContextFields("pipeline-context", pipelineCtx)...)
	fields = append(fields, cancellationContextFields("query-context", queryCtx)...)
	fields = append(fields, cancellationContextFields("top-context", topCtx)...)
	fields = append(fields, extra...)
	process.WarnPipelineCleanup(
		proc, internalSQLCancellationStageKey, cancellationDiagnosticLog, fields...)
}

// wrapRunSQLTrackerCancel records the in-process owner that can cancel an
// internal executor through transaction coordination. The original cancel is
// invoked exactly once and receives no changed arguments.
func wrapRunSQLTrackerCancel(
	c *Compile,
	cancel context.CancelFunc,
	sqlText string,
	token *atomic.Uint64,
) context.CancelFunc {
	if cancel == nil || c == nil || c.proc == nil || c.proc.GetTopContext() == nil ||
		!perfcounter.IsInternalExecutor(c.proc.GetTopContext()) {
		return cancel
	}
	// The tracker may invoke this callback from another goroutine, after the
	// Compile has advanced to a new pipeline generation. Capture the registered
	// generation now both for accurate provenance and to avoid reading mutable
	// process context fields from the callback.
	pipelineCtx := c.proc.Ctx
	queryCtx, _ := process.GetQueryCtxFromProc(c.proc)
	topCtx := c.proc.GetTopContext()
	queryID := c.proc.QueryId()
	var txnField zap.Field
	hasTxnField := false
	if txnOperator := c.proc.GetTxnOperator(); txnOperator != nil {
		txnField = txnutil.TxnIDField(txnOperator.Txn())
		hasTxnField = true
	}
	return func() {
		delegated := false
		delegate := func() {
			if !delegated {
				delegated = true
				cancel()
			}
		}
		defer delegate()
		if process.WarnPipelineCleanupLazy(
			c.proc,
			pipelineCtx,
			internalSQLTrackerCancelKey,
			cancellationDiagnosticLog,
			func() []zap.Field {
				fields := []zap.Field{
					zap.String("stage", "run-sql-tracker-cancel"),
					zap.Uint64("run-sql-token", token.Load()),
					zap.String("query-id", queryID),
					zap.String("sql", commonutil.Abbreviate(sqlText, 500)),
					zap.ByteString("cancel-call-stack", debug.Stack()),
				}
				if hasTxnField {
					fields = append(fields, txnField)
				}
				fields = append(fields, cancellationContextFields("pipeline-context", pipelineCtx)...)
				fields = append(fields, cancellationContextFields("query-context-before-cancel", queryCtx)...)
				fields = append(fields, cancellationContextFields("top-context", topCtx)...)
				return fields
			},
			delegate,
		) {
			return
		}
		delegate()
	}
}
