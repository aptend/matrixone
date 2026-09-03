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

package connector

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/common/reuse"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/pSpool"
	"github.com/matrixorigin/matrixone/pkg/vm"
	"github.com/matrixorigin/matrixone/pkg/vm/process"
)

var _ vm.Operator = new(Connector)

// Connector pipe connector
type Connector struct {
	ctr container

	Reg               *process.WaitRegister
	cleanupSpool      *pSpool.PipelineSpool
	allocationAccount *mpool.AllocationAccount
	diagnostics       *connectorDiagnosticState
	vm.OperatorBase
}

type connectorDiagnosticState struct {
	mu         sync.Mutex
	id         uint64
	generation uint64
	attempt    uint64
}

var nextConnectorDiagnosticID atomic.Uint64

type container struct {
	sp *pSpool.PipelineSpool
}

func (connector *Connector) GetOperatorBase() *vm.OperatorBase {
	return &connector.OperatorBase
}

func init() {
	reuse.CreatePool[Connector](
		func() *Connector {
			return &Connector{}
		},
		func(a *Connector) {
			*a = Connector{}
		},
		reuse.DefaultOptions[Connector]().
			WithEnableChecker(),
	)
}

func (connector Connector) TypeName() string {
	return opName
}

func (connector *Connector) OpType() vm.OpType {
	return vm.Connector
}

func NewArgument() *Connector {
	connector := reuse.Alloc[Connector](nil)
	connector.diagnostics = &connectorDiagnosticState{
		id: nextConnectorDiagnosticID.Add(1),
	}
	return connector
}

func (connector *Connector) beginDiagnosticGeneration() {
	if connector.diagnostics == nil {
		connector.diagnostics = &connectorDiagnosticState{
			id: nextConnectorDiagnosticID.Add(1),
		}
	}
	connector.diagnostics.mu.Lock()
	defer connector.diagnostics.mu.Unlock()
	connector.diagnostics.generation++
	connector.diagnostics.attempt = 0
}

func (connector *Connector) nextTerminalDiagnosticOwner(
	proc *process.Process,
) process.PipelineTerminalDiagnosticOwner {
	if connector.diagnostics == nil {
		connector.diagnostics = &connectorDiagnosticState{
			id: nextConnectorDiagnosticID.Add(1),
		}
	}
	connector.diagnostics.mu.Lock()
	defer connector.diagnostics.mu.Unlock()
	if connector.diagnostics.generation == 0 {
		connector.diagnostics.generation = 1
	}
	connector.diagnostics.attempt++
	pipelineContextID := uint64(0)
	if proc != nil {
		pipelineContextID = process.PipelineContextDiagnosticID(proc.Ctx)
	}
	return process.PipelineTerminalDiagnosticOwner{
		Kind:              "connector_reset",
		ID:                connector.diagnostics.id,
		Generation:        connector.diagnostics.generation,
		Attempt:           connector.diagnostics.attempt,
		PipelineContextID: pipelineContextID,
	}
}

func (connector *Connector) WithReg(reg *process.WaitRegister) *Connector {
	connector.Reg = reg
	return connector
}

func (connector *Connector) SetAllocationAccount(
	account *mpool.AllocationAccount,
) error {
	if account == nil || account.Handle() == 0 {
		return mpool.ErrAllocationAccountInvalid
	}
	if connector.allocationAccount != nil && connector.allocationAccount != account {
		return mpool.ErrAllocationAccountMismatch
	}
	connector.allocationAccount = account
	return nil
}

// ActivatesAllocationAccountLifecycle reports that Connector only participates
// in an account already required by an allocation-producing operator.
func (connector *Connector) ActivatesAllocationAccountLifecycle() bool {
	return false
}

func (connector *Connector) ClearAllocationAccount(
	account *mpool.AllocationAccount,
) error {
	if connector.allocationAccount == nil {
		return nil
	}
	if connector.allocationAccount != account {
		return mpool.ErrAllocationAccountMismatch
	}
	if connector.ctr.sp != nil {
		return mpool.ErrAllocationAccountInvariant
	}
	if connector.cleanupSpool != nil {
		connector.cleanupSpool.FinalizeAfterConsumersQuiesced()
		connector.cleanupSpool = nil
	}
	connector.allocationAccount = nil
	return nil
}

func (connector *Connector) Release() {
	if connector != nil {
		reuse.Free[Connector](connector, nil)
	}
}

func (connector *Connector) Reset(proc *process.Process, pipelineFailed bool, err error) {
	diagnosticOwner := connector.nextTerminalDiagnosticOwner(proc)
	terminalSignal := process.BuildCleanupSignal(pipelineFailed, err)
	terminalErr := terminalSignal.TerminalErr()
	signalCtx, signalCancel := context.WithTimeout(context.TODO(), process.PipelineSignalSendTimeout)
	defer signalCancel()
	signalCtx = process.WithPipelineTerminalDiagnosticOwner(signalCtx, diagnosticOwner)

	terminalDelivered := connector.sendTerminalWithLog(
		signalCtx, proc, terminalSignal, pipelineFailed, terminalErr, diagnosticOwner)

	if connector.ctr.sp != nil {
		sp := connector.ctr.sp

		if terminalSignal.EventType == process.EventEnd && terminalDelivered {
			connector.cleanupSpool = sp
		} else {
			abortErr := terminalErr
			if terminalSignal.EventType == process.EventEnd && !terminalDelivered {
				fallbackErr := process.ResolvePipelineSpoolAbortError(connector.Reg)
				connector.sendTerminalWithLog(
					signalCtx, proc, process.NewAbortSignal(fallbackErr), true, fallbackErr, diagnosticOwner)
				abortErr = fallbackErr
			}
			sp.Abort(abortErr)
			if connector.allocationAccount != nil {
				connector.cleanupSpool = sp
			} else {
				connector.cleanupSpool = nil
			}
		}
		connector.ctr.sp = nil
	} else if terminalSignal.EventType == process.EventEnd && !terminalDelivered {
		fallbackErr := process.ErrPipelineEndSignalDeliveryFailed
		connector.sendTerminalWithLog(
			signalCtx, proc, process.NewAbortSignal(fallbackErr), true, fallbackErr, diagnosticOwner)
	}
}

// sendTerminalWithLog sends a terminal signal to Reg, logging a warning on failure.
func (connector *Connector) sendTerminalWithLog(
	ctx context.Context,
	proc *process.Process,
	signal process.PipelineSignal,
	pipelineFailed bool,
	err error,
	diagnosticOwner process.PipelineTerminalDiagnosticOwner,
) bool {
	if connector.Reg == nil {
		process.WarnPipelineCleanupf(
			proc,
			"connector_cleanup_nil_reg",
			"connector cleanup skipped terminal %s signal because Reg is nil: terminal_owner=%s pipeline_failed=%t err=%v",
			signal.EventType.String(),
			diagnosticOwner.String(),
			pipelineFailed,
			err)
		return false
	}
	delivered := process.SendPipelineSignalWithContext(ctx, connector.Reg, signal)
	if delivered {
		return true
	}
	edgeState := process.PipelineEdgeDiagnostics(connector.Reg)
	queryID := ""
	pipelineCancelDiagnostic := "owner=none"
	var pipelineErr, pipelineCause, queryErr, queryCause error
	if proc != nil {
		queryID = proc.QueryId()
		pipelineCancelDiagnostic = process.PipelineCancellationDiagnostic(proc.Ctx)
		if proc.Ctx != nil {
			pipelineErr = proc.Ctx.Err()
			pipelineCause = context.Cause(proc.Ctx)
		}
		if proc.Base != nil {
			queryCtx, _ := process.GetQueryCtxFromProc(proc)
			if queryCtx != nil {
				queryErr = queryCtx.Err()
				queryCause = context.Cause(queryCtx)
			}
		}
	}
	fatalEvent := "None"
	if edgeState.FatalTerminal {
		fatalEvent = edgeState.FatalEvent.String()
	}
	firstTerminalEvent := "None"
	lastTerminalEvent := "None"
	doneTerminalEvent := "None"
	if edgeState.TerminalApplied > 0 {
		firstTerminalEvent = edgeState.FirstTerminalEvent.String()
		lastTerminalEvent = edgeState.LastTerminalEvent.String()
	}
	if edgeState.DoneClosed {
		doneTerminalEvent = edgeState.DoneTerminalEvent.String()
	}
	process.WarnPipelineCleanupf(
		proc,
		"connector_cleanup_send_terminal_signal",
		"connector cleanup failed sending terminal %s signal: timeout=%s query_id=%s terminal_owner=%s edge_id=%d edge_generation=%d channel_len=%d channel_cap=%d expected_ends=%d recorded_ends=%d terminal_attempts=%d terminal_applied=%d first_terminal_event=%s first_terminal_owner=%s last_terminal_event=%s last_terminal_owner=%s done_terminal_event=%s done_terminal_owner=%s fatal_terminal=%t fatal_event=%s fatal_err=%v fatal_delivered=%d fatal_remaining=%d done_closed=%t abort_closed=%t pipeline_failed=%t err=%v pipeline_err=%v pipeline_cause=%v pipeline_cancel=%s query_err=%v query_cause=%v",
		signal.EventType.String(),
		process.PipelineSignalSendTimeout,
		queryID,
		diagnosticOwner.String(),
		edgeState.ID,
		edgeState.Generation,
		edgeState.ChannelLen,
		edgeState.ChannelCap,
		edgeState.ExpectedEnds,
		edgeState.RecordedEnds,
		edgeState.TerminalAttempts,
		edgeState.TerminalApplied,
		firstTerminalEvent,
		edgeState.FirstTerminalOwner.String(),
		lastTerminalEvent,
		edgeState.LastTerminalOwner.String(),
		doneTerminalEvent,
		edgeState.DoneTerminalOwner.String(),
		edgeState.FatalTerminal,
		fatalEvent,
		edgeState.FatalErr,
		edgeState.FatalDelivered,
		edgeState.FatalRemaining,
		edgeState.DoneClosed,
		edgeState.AbortClosed,
		pipelineFailed,
		err,
		pipelineErr,
		pipelineCause,
		pipelineCancelDiagnostic,
		queryErr,
		queryCause)
	return false
}

// CleanupDeferredSpool reclaims spool cache memory after the paired Merge
// cleanup has returned on a normal End path. The normal path drains queued
// GetFromSpool signals; a cleanup-time timeout releases the current reference
// and leaves no receiver goroutine that can read pending signals later.
func (connector *Connector) CleanupDeferredSpool() {
	if connector.cleanupSpool == nil {
		return
	}
	if connector.allocationAccount != nil {
		connector.cleanupSpool.ReleaseReusableCacheAfterProducerQuiesced()
		return
	}
	connector.cleanupSpool.ForceCleanupAfterTerminalSignal()
	connector.cleanupSpool = nil
}

func (connector *Connector) Free(proc *process.Process, pipelineFailed bool, err error) {
}

func (connector *Connector) ExecProjection(proc *process.Process, input *batch.Batch) (*batch.Batch, error) {
	return input, nil
}
