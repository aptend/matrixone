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
	"github.com/matrixorigin/matrixone/pkg/vm/process"
	"go.uber.org/zap"
)

const cancellationDiagnosticLog = "internal SQL cancellation diagnostic"

func isInternalCancellationDiagnostic(proc *process.Process, err error) bool {
	return proc != nil && isScopeCancellationError(err) &&
		perfcounter.IsInternalExecutor(proc.GetTopContext())
}

func cancellationContextFields(prefix string, ctx context.Context) []zap.Field {
	if ctx == nil {
		return []zap.Field{zap.Bool(prefix+"-present", false)}
	}
	return []zap.Field{
		zap.Bool(prefix+"-present", true),
		zap.String(prefix+"-err", cancellationErrorString(ctx.Err())),
		zap.String(prefix+"-cause", cancellationErrorString(context.Cause(ctx))),
	}
}

func cancellationErrorString(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
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
	fields := make([]zap.Field, 0, 14+len(extra))
	fields = append(fields,
		zap.String("stage", stage),
		zap.String("error", cancellationErrorString(err)),
	)
	fields = append(fields, cancellationContextFields("pipeline-context", pipelineCtx)...)
	fields = append(fields, cancellationContextFields("query-context", queryCtx)...)
	fields = append(fields, cancellationContextFields("top-context", topCtx)...)
	fields = append(fields, extra...)
	proc.Error(context.Background(), cancellationDiagnosticLog, fields...)
}

func wrapRunSQLTrackerCancel(
	c *Compile,
	cancel context.CancelFunc,
	sqlText string,
	token *atomic.Uint64,
) context.CancelFunc {
	if cancel == nil || c == nil || c.proc == nil ||
		!perfcounter.IsInternalExecutor(c.proc.GetTopContext()) {
		return cancel
	}
	return func() {
		queryCtx, _ := process.GetQueryCtxFromProc(c.proc)
		fields := []zap.Field{
			zap.String("stage", "run-sql-tracker-cancel"),
			zap.Uint64("run-sql-token", token.Load()),
			zap.String("sql", commonutil.Abbreviate(sqlText, 500)),
			zap.ByteString("cancel-call-stack", debug.Stack()),
		}
		fields = append(fields, cancellationContextFields("pipeline-context", c.proc.Ctx)...)
		fields = append(fields, cancellationContextFields("query-context-before-cancel", queryCtx)...)
		fields = append(fields, cancellationContextFields("top-context", c.proc.GetTopContext())...)
		c.proc.Error(context.Background(), cancellationDiagnosticLog, fields...)
		cancel()
	}
}
