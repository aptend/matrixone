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
	"sync"
	"time"

	metricv2 "github.com/matrixorigin/matrixone/pkg/util/metric/v2"
	"go.uber.org/zap"
)

const (
	pipelineCleanupWarnInterval       = 10 * time.Second
	pipelineCleanupWarnBurstCount     = int64(3)
	pipelineCleanupWarnSampleInterval = int64(100)
)

var pipelineCleanupWarnLimiter = newCleanupWarnLimiter()

type cleanupWarnLimiter struct {
	mu     sync.Mutex
	states map[string]*cleanupWarnState
}

type cleanupWarnState struct {
	count      int64
	suppressed int64
	lastLog    time.Time
}

func newCleanupWarnLimiter() *cleanupWarnLimiter {
	return &cleanupWarnLimiter{
		states: make(map[string]*cleanupWarnState),
	}
}

func (l *cleanupWarnLimiter) allow(key string) (bool, int64, int64) {
	if key == "" {
		key = "pipeline_cleanup"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	state, ok := l.states[key]
	if !ok {
		state = &cleanupWarnState{}
		l.states[key] = state
	}

	state.count++
	count := state.count
	now := time.Now()

	shouldLog := count <= pipelineCleanupWarnBurstCount ||
		count%pipelineCleanupWarnSampleInterval == 0 ||
		now.Sub(state.lastLog) >= pipelineCleanupWarnInterval
	if !shouldLog {
		state.suppressed++
		return false, count, 0
	}

	suppressed := state.suppressed
	state.suppressed = 0
	state.lastLog = now
	return true, count, suppressed
}

func WarnPipelineCleanupf(proc *Process, key string, format string, args ...any) {
	allowed, occurrence, suppressed := allowPipelineCleanupWarning(proc, key)
	if !allowed {
		return
	}

	format += " occurrence=%d"
	args = append(args, occurrence)
	if suppressed > 0 {
		format += " suppressed=%d"
		args = append(args, suppressed)
	}

	proc.Warnf(proc.Ctx, format, args...)
}

// WarnPipelineCleanup emits a structured warning under the same global bound
// as WarnPipelineCleanupf. Keys must be fixed constants because they are also
// metric labels.
func WarnPipelineCleanup(proc *Process, key string, msg string, fields ...zap.Field) {
	allowed, occurrence, suppressed := allowPipelineCleanupWarning(proc, key)
	if !allowed {
		return
	}
	emitPipelineCleanupWarning(proc, msg, occurrence, suppressed, fields)
}

// WarnPipelineCleanupLazy avoids building expensive diagnostic fields, such
// as call stacks, when the global warning bound suppresses the event. afterBuild
// runs after fields capture and before the logger, allowing cancellation-owner
// diagnostics to preserve pre-cancel state without delaying cancellation on I/O.
func WarnPipelineCleanupLazy(
	proc *Process,
	key string,
	msg string,
	buildFields func() []zap.Field,
	afterBuild func(),
) bool {
	allowed, occurrence, suppressed := allowPipelineCleanupWarning(proc, key)
	if !allowed {
		return false
	}
	var fields []zap.Field
	if buildFields != nil {
		fields = buildFields()
	}
	if afterBuild != nil {
		afterBuild()
	}
	emitPipelineCleanupWarning(proc, msg, occurrence, suppressed, fields)
	return true
}

func emitPipelineCleanupWarning(
	proc *Process,
	msg string,
	occurrence int64,
	suppressed int64,
	fields []zap.Field,
) {
	fields = append(fields, zap.Int64("occurrence", occurrence))
	if suppressed > 0 {
		fields = append(fields, zap.Int64("suppressed", suppressed))
	}
	proc.Warn(proc.Ctx, msg, fields...)
}

func allowPipelineCleanupWarning(proc *Process, key string) (bool, int64, int64) {
	if key == "" {
		key = "pipeline_cleanup"
	}
	metricv2.PipelineCleanupEventCounter.WithLabelValues(key).Inc()

	if proc == nil {
		return false, 0, 0
	}

	allowed, occurrence, suppressed := pipelineCleanupWarnLimiter.allow(key)
	if !allowed {
		return false, occurrence, 0
	}
	return true, occurrence, suppressed
}
