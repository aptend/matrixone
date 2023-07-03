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

package sm

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/matrixorigin/matrixone/pkg/logutil"
)

const (
	Created int32 = iota
	Running
	ReceiverStopped
	PrepareStop
	Stopped
)

type safeQueue struct {
	queue     chan any
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	state     atomic.Int32
	pending   atomic.Int64
	batchSize int
	onItemsCB OnItemsCB
}

func NewSafeQueue(queueSize, batchSize int, onItem OnItemsCB) *safeQueue {
	q := &safeQueue{
		queue:     make(chan any, queueSize),
		batchSize: batchSize,
		onItemsCB: onItem,
	}
	q.state.Store(Created)
	q.ctx, q.cancel = context.WithCancel(context.Background())
	return q
}

func (q *safeQueue) Start(tags ...string) {
	q.state.Store(Running)
	q.wg.Add(1)
	items := make([]any, 0, q.batchSize)
	go func() {
		var buf [64]byte
		n := runtime.Stack(buf[:], false)
		idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
		id, _ := strconv.Atoi(idField)
		logutil.Infof("Slow commit queue %v on %d", tags, id)
		moment := time.Now()
		defer q.wg.Done()
		var wake, handle time.Duration
		for {
			select {
			case <-q.ctx.Done():
				return
			case item := <-q.queue:
				wake = time.Since(moment)
				if q.onItemsCB == nil {
					continue
				}
				items = append(items, item)
			Left:
				for i := 0; i < q.batchSize-1; i++ {
					select {
					case item = <-q.queue:
						items = append(items, item)
					default:
						break Left
					}
				}
				cnt := len(items)
				moment = time.Now()
				q.onItemsCB(items...)
				handle = time.Since(moment)
				if wake > 1*time.Second && len(tags) > 0 {
					logutil.Infof("Slow commit queue fire %v, elapse %v, cnt %d", tags, wake, cnt)
				}
				if handle > 1*time.Second && len(tags) > 0 {
					logutil.Infof("Slow commit queue onItemsCB %v, elapse %v, cnt %d", tags, handle, cnt)
				}
				items = items[:0]
				q.pending.Add(-1 * int64(cnt))
				moment = time.Now()
			}
		}
	}()
}

func (q *safeQueue) Stop() {
	q.stopReceiver()
	q.waitStop()
	close(q.queue)
}

func (q *safeQueue) stopReceiver() {
	state := q.state.Load()
	if state >= ReceiverStopped {
		return
	}
	q.state.CompareAndSwap(state, ReceiverStopped)
}

func (q *safeQueue) waitStop() {
	if q.state.Load() <= Running {
		panic("logic error")
	}
	if q.state.Load() == Stopped {
		return
	}
	if q.state.CompareAndSwap(ReceiverStopped, PrepareStop) {
		for q.pending.Load() != 0 {
			runtime.Gosched()
		}
		q.cancel()
	}
	q.wg.Wait()
}

func (q *safeQueue) Enqueue(item any) (any, error) {
	if q.state.Load() != Running {
		return item, ErrClose
	}
	q.pending.Add(1)
	if q.state.Load() != Running {
		q.pending.Add(-1)
		return item, ErrClose
	}
	q.queue <- item
	return item, nil
}
