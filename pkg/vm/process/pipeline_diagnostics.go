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

package process

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

type pipelineCancellationDiagnosticState struct {
	mu sync.Mutex

	id          uint64
	parent      *pipelineCancellationDiagnosticState
	directOwner bool
	caller      string
	source      string
	cause       string
	reported    atomic.Bool
}

type pipelineDiagnosticContextKey struct{}

var pipelineDiagnosticKey pipelineDiagnosticContextKey

var nextPipelineContextDiagnosticID atomic.Uint64

// pipelineDiagnosticContext records cancellation ownership beside the original
// context. It deliberately does not wrap or replace context.Cause.
type pipelineDiagnosticContext struct {
	context.Context
	diagnostics pipelineCancellationDiagnosticState
}

func (ctx *pipelineDiagnosticContext) Value(key any) any {
	if _, ok := key.(pipelineDiagnosticContextKey); ok {
		return &ctx.diagnostics
	}
	return ctx.Context.Value(key)
}

var errPipelineContextReplaced = fmt.Errorf("pipeline context replaced")

func pipelineCancellationCaller() string {
	pc, _, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown"
	}
	function := runtime.FuncForPC(pc)
	if function == nil {
		return fmt.Sprintf("unknown:%d", line)
	}
	return fmt.Sprintf("%s:%d", function.Name(), line)
}

func (proc *Process) buildPipelineContextWithDiagnostics(
	parent context.Context,
) (context.Context, context.CancelCauseFunc) {
	cancelCtx, rawCancel := context.WithCancelCause(parent)
	ctx := &pipelineDiagnosticContext{Context: cancelCtx}
	ctx.diagnostics.id = nextPipelineContextDiagnosticID.Add(1)
	ctx.diagnostics.parent, _ = parent.Value(pipelineDiagnosticKey).(*pipelineCancellationDiagnosticState)

	cancel := func(cause error) {
		// Replacing a process's old pipeline context is ordinary lifecycle work.
		// Preserve the original nil-cause cancellation and omit it as an owner.
		if cause == errPipelineContextReplaced {
			rawCancel(nil)
			return
		}

		ctx.diagnostics.mu.Lock()
		if cancelCtx.Err() != nil || ctx.diagnostics.directOwner {
			ctx.diagnostics.mu.Unlock()
			rawCancel(cause)
			return
		}

		parentCauseBefore := context.Cause(parent)
		ctx.diagnostics.directOwner = true
		ctx.diagnostics.caller = pipelineCancellationCaller()
		ctx.diagnostics.source = "direct_process_cancel"
		if parentCauseBefore != nil {
			ctx.diagnostics.source = "direct_or_parent_race"
		}
		if cause != nil {
			ctx.diagnostics.cause = cause.Error()
		} else {
			ctx.diagnostics.cause = context.Canceled.Error()
		}
		ctx.diagnostics.mu.Unlock()

		// Pass the caller's exact cause through. In particular, nil must remain
		// nil so context.Cause retains the standard context.Canceled value.
		rawCancel(cause)
		actualCause := context.Cause(cancelCtx)
		parentCauseAfter := context.Cause(parent)

		ctx.diagnostics.mu.Lock()
		if parentCauseAfter != nil {
			ctx.diagnostics.source = "direct_or_parent_race"
		}
		if actualCause != nil {
			ctx.diagnostics.cause = actualCause.Error()
		}
		ctx.diagnostics.mu.Unlock()
	}
	return ctx, cancel
}

// PipelineContextDiagnosticID returns zero for ordinary (non-instrumented)
// query pipelines and a stable process-local identity for internal SQL.
func PipelineContextDiagnosticID(ctx context.Context) uint64 {
	if ctx == nil {
		return 0
	}
	diagnostics, _ := ctx.Value(pipelineDiagnosticKey).(*pipelineCancellationDiagnosticState)
	if diagnostics == nil {
		return 0
	}
	return diagnostics.id
}

func cancellationDiagnosticStateSnapshot(
	diagnostics *pipelineCancellationDiagnosticState,
) (id uint64, parent *pipelineCancellationDiagnosticState, direct bool, caller string, source string, cause string) {
	if diagnostics == nil {
		return 0, nil, false, "", "", ""
	}
	diagnostics.mu.Lock()
	defer diagnostics.mu.Unlock()
	return diagnostics.id,
		diagnostics.parent,
		diagnostics.directOwner,
		diagnostics.caller,
		diagnostics.source,
		diagnostics.cause
}

// PipelineCancellationDiagnostic formats the first owner recorded for an
// internal SQL pipeline. It is called only from rate-limited anomaly logs.
func PipelineCancellationDiagnostic(ctx context.Context) string {
	if ctx == nil || context.Cause(ctx) == nil {
		return "owner=none"
	}

	diagnostics, _ := ctx.Value(pipelineDiagnosticKey).(*pipelineCancellationDiagnosticState)
	if diagnostics == nil {
		return fmt.Sprintf("owner=parent_context cause=%q", context.Cause(ctx).Error())
	}
	id, parent, directOwner, caller, source, diagnosticCause :=
		cancellationDiagnosticStateSnapshot(diagnostics)
	if !directOwner {
		for ancestor := parent; ancestor != nil; {
			ancestorID, next, ancestorDirect, ancestorCaller, ancestorSource, ancestorCause :=
				cancellationDiagnosticStateSnapshot(ancestor)
			if ancestorDirect {
				return fmt.Sprintf(
					"owner=parent_pipeline context_id=%d parent_context_id=%d inherited_owner=%s inherited_caller=%s inherited_cause=%q cause=%q",
					id,
					ancestorID,
					ancestorSource,
					ancestorCaller,
					ancestorCause,
					context.Cause(ctx).Error(),
				)
			}
			ancestor = next
		}
		return fmt.Sprintf(
			"owner=parent_context context_id=%d cause=%q",
			id,
			context.Cause(ctx).Error(),
		)
	}
	return fmt.Sprintf(
		"owner=%s context_id=%d caller=%s cause=%q",
		source,
		id,
		caller,
		diagnosticCause,
	)
}

// ClaimPipelineCancellationDiagnostic allows one escaping-cancellation warning
// per instrumented pipeline context. The global warning limiter remains the
// cross-query bound.
func ClaimPipelineCancellationDiagnostic(ctx context.Context) bool {
	if ctx == nil {
		return true
	}
	diagnostics, _ := ctx.Value(pipelineDiagnosticKey).(*pipelineCancellationDiagnosticState)
	if diagnostics == nil {
		return true
	}
	return diagnostics.reported.CompareAndSwap(false, true)
}
