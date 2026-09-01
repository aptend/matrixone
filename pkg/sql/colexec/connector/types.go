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
	vm.OperatorBase
}

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
	return reuse.Alloc[Connector](nil)
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
	terminalSignal := process.BuildCleanupSignal(pipelineFailed, err)
	terminalErr := terminalSignal.TerminalErr()
	signalCtx, signalCancel := context.WithTimeout(context.TODO(), process.PipelineSignalSendTimeout)
	defer signalCancel()

	terminalDelivered := connector.sendTerminalWithLog(signalCtx, proc, terminalSignal, pipelineFailed, terminalErr)

	if connector.ctr.sp != nil {
		sp := connector.ctr.sp

		if terminalSignal.EventType == process.EventEnd && terminalDelivered {
			connector.cleanupSpool = sp
		} else {
			abortErr := terminalErr
			if terminalSignal.EventType == process.EventEnd && !terminalDelivered {
				fallbackErr := process.ResolvePipelineSpoolAbortError(connector.Reg)
				connector.sendTerminalWithLog(signalCtx, proc, process.NewAbortSignal(fallbackErr), true, fallbackErr)
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
		connector.sendTerminalWithLog(signalCtx, proc, process.NewAbortSignal(fallbackErr), true, fallbackErr)
	}
}

// sendTerminalWithLog sends a terminal signal to Reg, logging a warning on failure.
func (connector *Connector) sendTerminalWithLog(ctx context.Context, proc *process.Process, signal process.PipelineSignal, pipelineFailed bool, err error) bool {
	if connector.Reg == nil {
		process.WarnPipelineCleanupf(
			proc,
			"connector_cleanup_nil_reg",
			"connector cleanup skipped terminal %s signal because Reg is nil: pipeline_failed=%t err=%v",
			signal.EventType.String(),
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
	process.WarnPipelineCleanupf(
		proc,
		"connector_cleanup_send_terminal_signal",
		"connector cleanup failed sending terminal %s signal: timeout=%s query_id=%s edge_id=%d edge_generation=%d channel_len=%d channel_cap=%d expected_ends=%d recorded_ends=%d fatal_terminal=%t fatal_event=%s fatal_err=%v fatal_delivered=%d fatal_remaining=%d done_closed=%t abort_closed=%t pipeline_failed=%t err=%v pipeline_err=%v pipeline_cause=%v pipeline_cancel=%s query_err=%v query_cause=%v",
		signal.EventType.String(),
		process.PipelineSignalSendTimeout,
		queryID,
		edgeState.ID,
		edgeState.Generation,
		edgeState.ChannelLen,
		edgeState.ChannelCap,
		edgeState.ExpectedEnds,
		edgeState.RecordedEnds,
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
