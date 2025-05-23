// Copyright 2025 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package merge

import (
	"context"
	"testing"
	"time"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/index"
	"github.com/stretchr/testify/require"
)

func TestSimulator(t *testing.T) {
	clock := newFakeClock()

	scatalog := NewSCatalog()
	sexecutor := NewSExecutor(clock, scatalog)

	sched := NewMergeScheduler(
		5*time.Second,
		scatalog,
		sexecutor,
		clock,
	)
	sched.Start()
	defer sched.Stop()

	scatalog.AddData(SData{
		stats: newTestObjectStats(t, 100, 200, 20*1024, 50, 0, objectio.NewSegmentid(), 0),
	})

	scatalog.AddData(SData{
		stats: newTestObjectStats(t, 100, 200, 30*1024, 42, 0, objectio.NewSegmentid(), 0),
	})

	ticker := time.NewTicker(200 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				clock.Advance(30 * time.Second)
				// answer := sched.Query(scatalog.hero)
				// t.Logf("answer: %+v", answer)
			}
		}
	}()

	time.Sleep(10 * time.Second)
	cancel()
	ticker.Stop()
}

func TestSplitZM(t *testing.T) {
	zm := index.NewZM(types.T_int32, 0)
	zm.Update(int32(1))
	zm.Update(int32(20))
	zmSplit := splitZM(zm, []int{1, 1, 1})
	constantZMCount := 0
	for _, zm := range zmSplit {
		if IsConstantZM(zm) {
			constantZMCount++
		}
	}
	require.Equal(t, 2, constantZMCount)
}
