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
	"sync"
	"testing"
	"time"
)

func TestClock(t *testing.T) {
	clock := newFakeClock()
	ticker := clock.NewTicker(time.Second * 3)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		t.Log("ticker started")
		i := 0
		for i < 3 {
			select {
			case <-ticker.Chan():
				t.Logf("ticker fired %d times", i)
				i++
			}
		}
	}()

	go func() {
		clock.BlockUntil(1) // wait for the first ticker to fire
		clock.Advance(5 * time.Second)
		clock.BlockUntil(1) // wait for the first ticker to fire
		clock.Advance(5 * time.Second)
		clock.BlockUntil(1) // wait for the first ticker to fire
		clock.Advance(5 * time.Second)
		clock.BlockUntil(1) // wait for the first ticker to fire
		clock.Advance(5 * time.Second)
	}()

	wg.Wait()
}

func TestSimulator(t *testing.T) {
	clock := newFakeClock()
	sched := NewMergeScheduler(
		5*time.Second,
		&dummyCatalogSource{settingsFn: noDefaultSettings},
		&dummyExecutor{},
		clock,
	)
	sched.Start()
	defer sched.Stop()

	clock.BlockUntil(1)
	clock.Advance(5 * time.Second)
	clock.Advance(5 * time.Second)
	time.Sleep(time.Second * 1)
	clock.Advance(5 * time.Second)
	clock.Advance(5 * time.Second)
	time.Sleep(time.Second * 1)
	clock.Advance(5 * time.Second)
}
