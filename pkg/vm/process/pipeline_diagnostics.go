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
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
)

// PipelineCancellationSource identifies the owner that requested a remote
// pipeline cancellation. These values are diagnostic data, not metric labels.
type PipelineCancellationSource string

const (
	PipelineCancelStopSending              PipelineCancellationSource = "stop_sending"
	PipelineCancelStopSendingBeforePublish PipelineCancellationSource = "stop_sending_before_pipeline_publish"
	PipelineCancelRPCSessionClosed         PipelineCancellationSource = "rpc_session_closed"
)

// PipelineCancellationCause preserves why a remote pipeline was canceled.
// It remains cancellation-shaped so graceful StopSending normalization keeps
// its behavior while context.Cause retains the owner.
type PipelineCancellationCause struct {
	source   PipelineCancellationSource
	streamID uint64
	detail   error
}

func NewPipelineCancellationCause(
	source PipelineCancellationSource,
	streamID uint64,
	detail error,
) error {
	return &PipelineCancellationCause{source: source, streamID: streamID, detail: detail}
}

func (e *PipelineCancellationCause) Error() string {
	if e == nil {
		return context.Canceled.Error()
	}
	if e.detail != nil {
		return fmt.Sprintf(
			"pipeline canceled: source=%s stream_id=%d detail=%v",
			e.source, e.streamID, e.detail)
	}
	return fmt.Sprintf(
		"pipeline canceled: source=%s stream_id=%d",
		e.source, e.streamID)
}

func (e *PipelineCancellationCause) Unwrap() error { return context.Canceled }

func (e *PipelineCancellationCause) Source() PipelineCancellationSource {
	if e == nil {
		return ""
	}
	return e.source
}

func (e *PipelineCancellationCause) StreamID() uint64 {
	if e == nil {
		return 0
	}
	return e.streamID
}

func (e *PipelineCancellationCause) Detail() error {
	if e == nil {
		return nil
	}
	return e.detail
}

type pipelineCancellationDiagnosticState struct {
	mu sync.Mutex

	id          uint64
	parent      *pipelineCancellationDiagnosticState
	directOwner bool
	caller      string
	source      string
	streamID    uint64
	cause       string
	reported    atomic.Bool
}

type pipelineDiagnosticContextKey struct{}

var pipelineDiagnosticKey pipelineDiagnosticContextKey

var nextPipelineContextDiagnosticID atomic.Uint64

// pipelineDiagnosticContext keeps diagnostics beside the cancel context
// without wrapping context.Cause. Preserving the original concrete error is a
// protocol contract: remote error encoding special-cases *moerr.Error.
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

var errPipelineContextReplaced = errors.New("pipeline context replaced")

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

func sameCancellationCause(left, right error) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	leftType := reflect.TypeOf(left)
	if leftType != reflect.TypeOf(right) || !leftType.Comparable() {
		return false
	}
	return left == right
}

func cancellationDiagnosticOwner(cause error) (string, uint64) {
	source := "direct_process_cancel"
	streamID := uint64(0)
	var cancellationCause *PipelineCancellationCause
	if errors.As(cause, &cancellationCause) {
		source = string(cancellationCause.Source())
		streamID = cancellationCause.StreamID()
	}
	return source, streamID
}

func (proc *Process) buildPipelineContextWithDiagnostics(
	parent context.Context,
) (context.Context, context.CancelCauseFunc) {
	cancelCtx, rawCancel := context.WithCancelCause(parent)
	ctx := &pipelineDiagnosticContext{Context: cancelCtx}
	ctx.diagnostics.id = nextPipelineContextDiagnosticID.Add(1)
	ctx.diagnostics.parent, _ = parent.Value(pipelineDiagnosticKey).(*pipelineCancellationDiagnosticState)

	cancel := func(cause error) {
		if cause == errPipelineContextReplaced {
			rawCancel(nil)
			return
		}
		if cause == nil {
			cause = context.Canceled
		}
		ctx.diagnostics.mu.Lock()
		defer ctx.diagnostics.mu.Unlock()
		if cancelCtx.Err() != nil {
			return
		}
		caller := pipelineCancellationCaller()
		rawCancel(cause)
		actualCause := context.Cause(cancelCtx)
		if !sameCancellationCause(actualCause, cause) {
			return
		}
		ctx.diagnostics.directOwner = true
		ctx.diagnostics.caller = caller
		ctx.diagnostics.source, ctx.diagnostics.streamID = cancellationDiagnosticOwner(cause)
		if parentCause := context.Cause(parent); parentCause != nil &&
			sameCancellationCause(actualCause, parentCause) {
			ctx.diagnostics.source = "direct_or_parent_race"
		}
		ctx.diagnostics.cause = cause.Error()
	}
	return ctx, cancel
}

// PipelineContextDiagnosticID returns the stable diagnostic identity assigned
// to a pipeline context. Zero means the context was not built by Process.
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
) (id uint64, parent *pipelineCancellationDiagnosticState, direct bool, caller string, source string, streamID uint64, cause string) {
	if diagnostics == nil {
		return 0, nil, false, "", "", 0, ""
	}
	diagnostics.mu.Lock()
	defer diagnostics.mu.Unlock()
	return diagnostics.id,
		diagnostics.parent,
		diagnostics.directOwner,
		diagnostics.caller,
		diagnostics.source,
		diagnostics.streamID,
		diagnostics.cause
}

// PipelineCancellationDiagnostic formats the first cancellation owner retained
// by a pipeline context. It is called only from rate-limited anomaly logs.
func PipelineCancellationDiagnostic(ctx context.Context) string {
	if ctx == nil {
		return "owner=none"
	}
	cause := context.Cause(ctx)
	if cause == nil {
		return "owner=none"
	}

	diagnostics, _ := ctx.Value(pipelineDiagnosticKey).(*pipelineCancellationDiagnosticState)
	if diagnostics == nil {
		return fmt.Sprintf("owner=parent_context cause=%q", cause.Error())
	}
	id, parent, directOwner, caller, source, streamID, diagnosticCause :=
		cancellationDiagnosticStateSnapshot(diagnostics)
	if !directOwner {
		for ancestor := parent; ancestor != nil; {
			ancestorID, next, ancestorDirect, ancestorCaller, ancestorSource,
				ancestorStreamID, ancestorCause := cancellationDiagnosticStateSnapshot(ancestor)
			if ancestorDirect {
				return fmt.Sprintf(
					"owner=parent_pipeline context_id=%d parent_context_id=%d inherited_owner=%s inherited_caller=%s inherited_stream_id=%d inherited_cause=%q cause=%q",
					id,
					ancestorID,
					ancestorSource,
					ancestorCaller,
					ancestorStreamID,
					ancestorCause,
					cause.Error(),
				)
			}
			ancestor = next
		}
		return fmt.Sprintf(
			"owner=parent_context context_id=%d cause=%q",
			id,
			cause.Error(),
		)
	}
	return fmt.Sprintf(
		"owner=%s context_id=%d caller=%s stream_id=%d cause=%q",
		source,
		id,
		caller,
		streamID,
		diagnosticCause,
	)
}

// ClaimPipelineCancellationDiagnostic reports whether the caller owns the
// single anomaly report associated with this pipeline context. The global
// warning limiter remains the cross-query bound.
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
