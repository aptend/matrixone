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
	"fmt"
	"sync/atomic"
)

// PipelineTerminalDiagnosticOwner is fixed-size metadata attached only to
// internal SQL terminal sends. Operator fields come from the existing plan and
// avoid per-operator diagnostic allocation or synchronization.
type PipelineTerminalDiagnosticOwner struct {
	Kind              string
	OperatorID        int32
	ParallelID        int32
	OperatorIdx       int
	PipelineContextID uint64
}

func (owner PipelineTerminalDiagnosticOwner) enabled() bool {
	return owner.Kind != "" && owner.PipelineContextID != 0
}

func (owner PipelineTerminalDiagnosticOwner) String() string {
	if !owner.enabled() {
		return "owner=unknown"
	}
	return fmt.Sprintf(
		"owner=%s operator_id=%d parallel_id=%d operator_idx=%d pipeline_context_id=%d",
		owner.Kind,
		owner.OperatorID,
		owner.ParallelID,
		owner.OperatorIdx,
		owner.PipelineContextID,
	)
}

type pipelineEdgeDiagnosticState struct {
	id                uint64
	generation        uint64
	terminalAttempts  uint64
	doneTerminalEvent PipelineEventType
	doneTerminalOwner PipelineTerminalDiagnosticOwner
}

var nextPipelineEdgeDiagnosticID atomic.Uint64

func (e *PipelineEdge) ensureEdgeDiagnosticsLocked() *pipelineEdgeDiagnosticState {
	if e.diagnostics == nil {
		e.diagnostics = &pipelineEdgeDiagnosticState{
			id:         nextPipelineEdgeDiagnosticID.Add(1),
			generation: 1,
		}
	}
	return e.diagnostics
}

func (e *PipelineEdge) recordTerminalAttemptLocked(
	owner PipelineTerminalDiagnosticOwner,
) {
	if e.diagnostics == nil && !owner.enabled() {
		return
	}
	e.ensureEdgeDiagnosticsLocked().terminalAttempts++
}

func (e *PipelineEdge) recordDoneTerminalLocked(
	signal PipelineSignal,
	owner PipelineTerminalDiagnosticOwner,
) {
	if e.diagnostics == nil && !owner.enabled() {
		return
	}
	diagnostics := e.ensureEdgeDiagnosticsLocked()
	diagnostics.doneTerminalEvent = signal.EventType
	diagnostics.doneTerminalOwner = owner
}

func (e *PipelineEdge) resetEdgeDiagnosticsLocked() {
	if e.diagnostics == nil {
		return
	}
	e.diagnostics.generation++
	e.diagnostics.terminalAttempts = 0
	e.diagnostics.doneTerminalEvent = EventData
	e.diagnostics.doneTerminalOwner = PipelineTerminalDiagnosticOwner{}
}

// PipelineEdgeDiagnosticSnapshot is a lock-consistent view captured only after
// an anomalous terminal send fails.
type PipelineEdgeDiagnosticSnapshot struct {
	ID                uint64
	Generation        uint64
	ExpectedEnds      int
	RecordedEnds      int
	TerminalAttempts  uint64
	DoneTerminalEvent PipelineEventType
	DoneTerminalOwner PipelineTerminalDiagnosticOwner
	FatalTerminal     bool
	FatalEvent        PipelineEventType
	FatalErr          error
	FatalDelivered    int
	FatalRemaining    int
	DoneClosed        bool
	AbortClosed       bool
	ChannelLen        int
	ChannelCap        int
}

func PipelineEdgeDiagnostics(e *WaitRegister) PipelineEdgeDiagnosticSnapshot {
	if e == nil {
		return PipelineEdgeDiagnosticSnapshot{}
	}
	e.initTerminalState()
	e.terminalMu.Lock()
	defer e.terminalMu.Unlock()

	snapshot := PipelineEdgeDiagnosticSnapshot{
		ExpectedEnds:   e.expectedEndCountLocked(),
		RecordedEnds:   e.endRecorded,
		FatalTerminal:  e.fatalTerminal,
		FatalDelivered: e.fatalDelivered,
		FatalRemaining: e.fatalRemaining,
		DoneClosed:     e.doneClosed,
		AbortClosed:    e.abortClosed,
		ChannelLen:     len(e.Ch2),
		ChannelCap:     cap(e.Ch2),
	}
	if e.diagnostics != nil {
		snapshot.ID = e.diagnostics.id
		snapshot.Generation = e.diagnostics.generation
		snapshot.TerminalAttempts = e.diagnostics.terminalAttempts
		snapshot.DoneTerminalEvent = e.diagnostics.doneTerminalEvent
		snapshot.DoneTerminalOwner = e.diagnostics.doneTerminalOwner
	}
	if e.fatalTerminal {
		snapshot.FatalEvent = e.fatalSignal.EventType
		snapshot.FatalErr = e.terminalErr
	}
	return snapshot
}
