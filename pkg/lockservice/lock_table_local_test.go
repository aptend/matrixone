// Copyright 2023 Matrix Origin
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

package lockservice

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/common/stopper"
	pb "github.com/matrixorigin/matrixone/pkg/pb/lock"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/util/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloseLocalLockTable(t *testing.T) {
	table := uint64(10)
	getRunner(false)(
		t,
		table,
		func(ctx context.Context, s *service, lt *localLockTable) {
			rows := newTestRows(1)
			txnID := newTestTxnID(1)
			mustAddTestLock(
				t,
				ctx,
				s,
				table,
				txnID,
				rows,
				pb.Granularity_Row)
			lt.close()
			lt.mu.Lock()
			defer lt.mu.Unlock()
			assert.True(t, lt.mu.closed)
			assert.Equal(t, 0, lt.mu.store.Len())
		},
	)
}

func TestCloseLocalLockTableWithBlockedWaiter(t *testing.T) {
	runLockServiceTests(
		t,
		[]string{"s1"},
		func(_ *lockTableAllocator, s []*service) {
			tableID := uint64(10)

			l := s[0]
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Second*10)
			defer cancel()

			mustAddTestLock(
				t,
				ctx,
				l,
				tableID,
				[]byte{1},
				[][]byte{{1}},
				pb.Granularity_Row)

			var wg sync.WaitGroup
			wg.Add(2)
			// txn2 wait txn1 or txn3
			go func() {
				defer wg.Done()
				_, err := l.Lock(
					ctx,
					tableID,
					[][]byte{{1}},
					[]byte{2},
					newTestRowExclusiveOptions(),
				)
				require.Equal(t, ErrLockTableNotFound, err)
			}()

			// txn3 wait txn2 or txn1
			go func() {
				defer wg.Done()
				_, err := l.Lock(
					ctx,
					tableID,
					[][]byte{{1}},
					[]byte{3},
					newTestRowExclusiveOptions(),
				)
				require.Equal(t, ErrLockTableNotFound, err)
			}()

			v, err := l.getLockTable(0, tableID)
			require.NoError(t, err)
			lt := v.(*localLockTable)
			for {
				lt.mu.RLock()
				lock, ok := lt.mu.store.Get([]byte{1})
				require.True(t, ok)
				lt.mu.RUnlock()
				if lock.waiters.size() == 2 {
					break
				}
				time.Sleep(time.Millisecond * 10)
			}

			v.close()
			wg.Wait()
		})
}

func TestMergeRangeWithNoConflict(t *testing.T) {
	cases := []struct {
		txnID         string
		existsLock    [][][]byte
		waitOnLock    [][]byte
		existsWaiters [][]string
		newLock       [][]byte
		mergedLocks   [][]byte
		mergedWaiters [][]string
		flags         []byte
	}{
		{
			txnID:         "[] + [1, 2] = [1, 2]",
			existsLock:    [][][]byte{},
			newLock:       [][]byte{{1}, {2}},
			mergedLocks:   [][]byte{{1}, {2}},
			mergedWaiters: [][]string{nil},
			flags:         []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:         "[1] + [2,3] = [1, 2, 3]",
			existsLock:    [][][]byte{{{1}}},
			newLock:       [][]byte{{2}, {3}},
			mergedLocks:   [][]byte{{1}, {2}, {3}},
			waitOnLock:    [][]byte{{1}},
			existsWaiters: [][]string{{"1"}},
			mergedWaiters: [][]string{{"1"}, nil},
			flags:         []byte{flagLockRow, flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:         "[1] + [1,3] = [1, 3]",
			existsLock:    [][][]byte{{{1}}},
			newLock:       [][]byte{{1}, {3}},
			mergedLocks:   [][]byte{{1}, {3}},
			waitOnLock:    [][]byte{{1}},
			existsWaiters: [][]string{{"1"}},
			mergedWaiters: [][]string{{"1"}},
			flags:         []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:         "[1] + [2] + [1, 3] = [1, 3]",
			existsLock:    [][][]byte{{{1}}, {{2}}},
			newLock:       [][]byte{{1}, {3}},
			mergedLocks:   [][]byte{{1}, {3}},
			waitOnLock:    [][]byte{{1}, {2}},
			existsWaiters: [][]string{{"1"}, {"2"}},
			mergedWaiters: [][]string{{"1", "2"}},
			flags:         []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:         "[1] + [2] + [3] + [1, 3] = [1, 3]",
			existsLock:    [][][]byte{{{1}}, {{2}}, {{3}}},
			newLock:       [][]byte{{1}, {3}},
			mergedLocks:   [][]byte{{1}, {3}},
			waitOnLock:    [][]byte{{1}, {2}, {3}},
			existsWaiters: [][]string{{"1"}, {"2"}, {"3"}},
			mergedWaiters: [][]string{{"1", "2", "3"}},
			flags:         []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:         "[1] + [2] + [3] + [4] + [1, 3] = [1, 3] + [4]",
			existsLock:    [][][]byte{{{1}}, {{2}}, {{3}}, {{4}}},
			newLock:       [][]byte{{1}, {3}},
			mergedLocks:   [][]byte{{1}, {3}, {4}},
			waitOnLock:    [][]byte{{1}, {2}, {3}, {4}},
			existsWaiters: [][]string{{"1"}, {"2"}, {"3"}, {"4"}},
			mergedWaiters: [][]string{{"1", "2", "3"}, {"4"}},
			flags:         []byte{flagLockRangeStart, flagLockRangeEnd, flagLockRow},
		},

		{
			txnID:         "[1, 2] + [3, 4] = [1, 2] + [3, 4]",
			existsLock:    [][][]byte{{{1}, {2}}},
			newLock:       [][]byte{{3}, {4}},
			mergedLocks:   [][]byte{{1}, {2}, {3}, {4}},
			waitOnLock:    [][]byte{{2}},
			existsWaiters: [][]string{{"1"}},
			mergedWaiters: [][]string{{"1"}, nil},
			flags:         []byte{flagLockRangeStart, flagLockRangeEnd, flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[3, 4] + [1, 2] = [1, 2] + [3, 4]",
			existsLock:  [][][]byte{{{3}, {4}}},
			newLock:     [][]byte{{1}, {2}},
			mergedLocks: [][]byte{{1}, {2}, {3}, {4}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd, flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[1, 4] + [1, 3] = [1, 4]",
			existsLock:  [][][]byte{{{1}, {4}}},
			newLock:     [][]byte{{1}, {3}},
			mergedLocks: [][]byte{{1}, {4}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[1, 4] + [1, 4] = [1, 4]",
			existsLock:  [][][]byte{{{1}, {4}}},
			newLock:     [][]byte{{1}, {4}},
			mergedLocks: [][]byte{{1}, {4}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[1, 4] + [1, 5] = [1, 5]",
			existsLock:  [][][]byte{{{1}, {4}}},
			newLock:     [][]byte{{1}, {5}},
			mergedLocks: [][]byte{{1}, {5}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[2, 4] + [1, 5] = [1, 5]",
			existsLock:  [][][]byte{{{2}, {4}}},
			newLock:     [][]byte{{1}, {5}},
			mergedLocks: [][]byte{{1}, {5}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[1, 4] + [2, 5] = [1, 5]",
			existsLock:  [][][]byte{{{1}, {4}}},
			newLock:     [][]byte{{2}, {5}},
			mergedLocks: [][]byte{{1}, {5}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[2, 5] + [1, 4] = [1, 5]",
			existsLock:  [][][]byte{{{2}, {5}}},
			newLock:     [][]byte{{1}, {4}},
			mergedLocks: [][]byte{{1}, {5}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[1, 5] + [2, 5] = [1, 5]",
			existsLock:  [][][]byte{{{1}, {5}}},
			newLock:     [][]byte{{2}, {5}},
			mergedLocks: [][]byte{{1}, {5}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[2, 5] + [1, 5] = [1, 5]",
			existsLock:  [][][]byte{{{2}, {5}}},
			newLock:     [][]byte{{1}, {5}},
			mergedLocks: [][]byte{{1}, {5}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[2, 6] + [1, 5] = [1, 6]",
			existsLock:  [][][]byte{{{2}, {6}}},
			newLock:     [][]byte{{1}, {5}},
			mergedLocks: [][]byte{{1}, {6}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[1, 5] + [2, 6] = [1, 6]",
			existsLock:  [][][]byte{{{1}, {5}}},
			newLock:     [][]byte{{2}, {6}},
			mergedLocks: [][]byte{{1}, {6}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[5, 6] + [1, 5] = [1, 6]",
			existsLock:  [][][]byte{{{5}, {6}}},
			newLock:     [][]byte{{1}, {5}},
			mergedLocks: [][]byte{{1}, {6}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[1, 5] + [5, 6] = [1, 6]",
			existsLock:  [][][]byte{{{1}, {5}}},
			newLock:     [][]byte{{5}, {6}},
			mergedLocks: [][]byte{{1}, {6}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:       "[2, 3] + [1, 4] = [1, 4]",
			existsLock:  [][][]byte{{{2}, {3}}, {{1}, {4}}},
			newLock:     [][]byte{{1}, {4}},
			mergedLocks: [][]byte{{1}, {4}},
			flags:       []byte{flagLockRangeStart, flagLockRangeEnd},
		},

		{
			txnID:         "[1, 2] + [3, 4] + [5] + [6] + [1, 5] = [1, 5] + [6]",
			existsLock:    [][][]byte{{{1}, {2}}, {{3}, {4}}, {{5}}, {{6}}},
			newLock:       [][]byte{{1}, {5}},
			mergedLocks:   [][]byte{{1}, {5}, {6}},
			waitOnLock:    [][]byte{{2}, {4}, {5}},
			existsWaiters: [][]string{{"1", "2"}, {"3", "4"}, {"5"}},
			mergedWaiters: [][]string{{"1", "2", "3", "4", "5"}, nil},
			flags:         []byte{flagLockRangeStart, flagLockRangeEnd, flagLockRow},
		},
	}

	runLockServiceTests(
		t,
		[]string{"s1"},
		func(_ *lockTableAllocator, s []*service) {
			l := s[0]
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Second*10)
			defer cancel()

			table := uint64(10)
			for _, c := range cases {
				stopper := stopper.NewStopper("")
				v, err := l.getLockTableWithCreate(0, table, nil, pb.Sharding_None)
				require.NoError(t, err)
				lt := v.(*localLockTable)

				for _, rows := range c.existsLock {
					opts := pb.LockOptions{}
					if len(rows) > 1 {
						opts.Granularity = pb.Granularity_Range
					}
					_, err := l.Lock(ctx, table, rows, []byte(c.txnID), opts)
					require.NoError(t, err)
				}
				for i, lock := range c.waitOnLock {
					lt.mu.Lock()
					lock, ok := lt.mu.store.Get(lock)
					if !ok {
						panic(ok)
					}
					var wg sync.WaitGroup
					for _, txnID := range c.existsWaiters[i] {
						w := acquireWaiter(pb.WaitTxn{TxnID: []byte(txnID)}, "", nil)
						w.setStatus(blocking)
						lock.waiters.put(w)
						wg.Add(1)
						require.NoError(t, stopper.RunTask(func(ctx context.Context) {
							wg.Done()
							w.wait(ctx, getLogger(""))
							w.close("", nil)
						}))
					}
					wg.Wait()
					lt.mu.Unlock()
				}

				opts := pb.LockOptions{}
				opts.Granularity = pb.Granularity_Range
				_, err = l.Lock(ctx, table, c.newLock, []byte(c.txnID), opts)
				require.NoError(t, err)

				lt.mu.Lock()
				var keys [][]byte
				var flags []byte
				idx := 0
				lt.mu.store.Iter(func(b []byte, l Lock) bool {
					keys = append(keys, b)
					flags = append(flags, l.value)
					if !l.isLockRangeStart() {
						if len(c.mergedWaiters) == 0 {
							assert.Equal(t, 0, l.waiters.size())
						} else {
							var waitTxns []string
							l.waiters.iter(func(v *waiter) bool {
								waitTxns = append(waitTxns, string(v.txn.TxnID))
								return true
							})
							require.Equal(t, c.mergedWaiters[idx], waitTxns)
							idx++
						}
					}
					return true
				})
				lt.mu.Unlock()
				require.Equal(t, c.mergedLocks, keys)
				for idx, v := range flags {
					assert.NotEqual(t, 0, v&c.flags[idx])
				}

				txn := l.activeTxnHolder.getActiveTxn([]byte(c.txnID), false, "")
				require.NotNil(t, txn)
				fn := func(values [][]byte) [][]byte {
					sort.Slice(values, func(i, j int) bool {
						return bytes.Compare(values[i], values[j]) < 0
					})
					return values
				}
				assert.Equal(t, fn(c.mergedLocks), fn(txn.getHoldLocksLocked(0).tableKeys[table].slice().all()))

				assert.NoError(t, l.Unlock(ctx, []byte(c.txnID), timestamp.Timestamp{}))
				stopper.Stop()
				table++
			}
		})
}

func TestLocalLockTableMultipleRowLocksCannotMissIfFoundSelfTxn(t *testing.T) {
	runLockServiceTests(
		t,
		[]string{"s1"},
		func(_ *lockTableAllocator, s []*service) {
			tableID := uint64(10)
			l := s[0]
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Second*10)
			defer cancel()

			mustAddTestLock(
				t,
				ctx,
				l,
				tableID,
				[]byte{2},
				[][]byte{{1}},
				pb.Granularity_Row)

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				mustAddTestLock(
					t,
					ctx,
					l,
					tableID,
					[]byte{1},
					[][]byte{{1}},
					pb.Granularity_Row)
			}()
			go func() {
				defer wg.Done()
				waitWaiters(t, l, tableID, []byte{1}, 1)
				mustAddTestLock(
					t,
					ctx,
					l,
					tableID,
					[]byte{1},
					[][]byte{{1}, {2}},
					pb.Granularity_Row)
			}()

			waitWaiters(t, l, tableID, []byte{1}, 2)
			require.NoError(t, l.Unlock(ctx, []byte{2}, timestamp.Timestamp{}))

			wg.Wait()
			v, err := l.getLockTable(0, tableID)
			require.NoError(t, err)
			lt := v.(*localLockTable)
			lt.mu.Lock()
			defer lt.mu.Unlock()
			require.Equal(t, 2, lt.mu.store.Len())
		})
}

func TestIssue9856(t *testing.T) {
	runLockServiceTests(
		t,
		[]string{"s1"},
		func(alloc *lockTableAllocator, s []*service) {
			tableID := uint64(10)

			l := s[0]
			ctx := context.Background()
			option := pb.LockOptions{
				Granularity: pb.Granularity_Range,
				Mode:        pb.LockMode_Exclusive,
				Policy:      pb.WaitPolicy_Wait,
			}

			values := `{"start": "073a150a3a153100", "end": "083a15083a1608c000", "mode": "Exclusive"}
			{"start": "013a15093a150100", "end": "033a15093a160bb800", "mode": "Exclusive"}
			{"start": "013a15013a150100", "end": "033a15013a160bb800", "mode": "Exclusive"}
			{"start": "093a15053a150100", "end": "0a3a15053a160bb800", "mode": "Exclusive"}
			{"start": "053a15043a160ba300", "end": "083a15013a160bb800", "mode": "Exclusive"}
			{"start": "093a15023a150100", "end": "0a3a15023a160baf00", "mode": "Exclusive"}
			{"start": "013a15073a150c00", "end": "043a15063a160bb800", "mode": "Exclusive"}
			{"start": "093a15083a150100", "end": "0a3a15083a160bb800", "mode": "Exclusive"}
			{"start": "053a15023a1608b300", "end": "073a15023a160bb800", "mode": "Exclusive"}
			{"start": "013a15033a150100", "end": "043a15013a160bb800", "mode": "Exclusive"}
			{"start": "053a15093a150100", "end": "083a15053a160bb800", "mode": "Exclusive"}
			{"start": "013a15063a150100", "end": "043a15043a160bb800", "mode": "Exclusive"}
			{"start": "053a15083a150100", "end": "083a15043a160bb800", "mode": "Exclusive"}
			{"start": "013a15043a1605d500", "end": "043a15033a160bb800", "mode": "Exclusive"}
			{"start": "053a15063a150100", "end": "073a15053a160bb800", "mode": "Exclusive"}
			{"start": "013a150a3a1605db00", "end": "053a15013a1602b200", "mode": "Exclusive"}
			{"start": "083a15083a1608c100", "end": "093a15013a16059800", "mode": "Exclusive"}
			{"start": "013a15093a160b8600", "end": "043a15083a160bb800", "mode": "Exclusive"}
			{"start": "093a15013a16059900", "end": "0a3a15013a16031f00", "mode": "Exclusive"}
			{"start": "093a15063a1602e000", "end": "0a3a15063a160bb800", "mode": "Exclusive"}
			{"start": "053a15053a150100", "end": "083a15023a16055200", "mode": "Exclusive"}
			{"start": "013a15083a150100", "end": "043a15073a16057300", "mode": "Exclusive"}
			{"start": "013a15063a1605a300", "end": "043a15053a160bb800", "mode": "Exclusive"}
			{"start": "093a15073a160b7000", "end": "0a3a15073a1608ff00", "mode": "Exclusive"}
			{"start": "073a15053a150100", "end": "083a15023a160bb800", "mode": "Exclusive"}
			{"start": "053a15033a16058b00", "end": "073a15033a160bb800", "mode": "Exclusive"}
			{"start": "033a15093a150100", "end": "043a15073a160bb800", "mode": "Exclusive"}
			{"start": "013a15023a150100", "end": "033a15023a160bb800", "mode": "Exclusive"}
			{"start": "013a15073a150100", "end": "023a15073a160bb800", "mode": "Exclusive"}
			{"start": "093a15093a1605d800", "end": "0a3a150a3a1602af00", "mode": "Exclusive"}
			{"start": "013a150a3a150100", "end": "023a150a3a160bb800", "mode": "Exclusive"}
			{"start": "053a15073a150100", "end": "083a15033a160bb800", "mode": "Exclusive"}
			{"start": "093a15033a150100", "end": "0a3a15033a160bb800", "mode": "Exclusive"}
			{"start": "013a15053a150100", "end": "033a15053a160bb800", "mode": "Exclusive"}
			{"start": "053a15083a1602ed00", "end": "073a15083a160b7c00", "mode": "Exclusive"}
			{"start": "023a15023a16056900", "end": "043a15013a1602ed00", "mode": "Exclusive"}
			{"start": "0a3a150a3a1602b000", "end": "0a3a150a3a160bb800", "mode": "Exclusive"}
			{"start": "053a15043a16026300", "end": "053a15043a160ba200", "mode": "Exclusive"}
			{"start": "053a15013a1602b300", "end": "073a15013a160b4200", "mode": "Exclusive"}
			{"start": "013a15033a160b7e00", "end": "043a15023a160bb800", "mode": "Exclusive"}
			{"start": "023a15053a160b3d00", "end": "043a15043a1608ca00", "mode": "Exclusive"}
			{"start": "073a15083a160b7d00", "end": "083a15053a16090700", "mode": "Exclusive"}
			{"start": "053a150a3a150100", "end": "083a15063a160bb800", "mode": "Exclusive"}
			{"start": "093a15043a16088800", "end": "0a3a15043a16060700", "mode": "Exclusive"}
			{"start": "053a15023a150100", "end": "073a15013a160bb800", "mode": "Exclusive"}`
			for _, r := range strings.Split(values, "\n") {
				v := &target{}
				json.MustUnmarshal([]byte(r), v)
				_, err := l.Lock(ctx, tableID, [][]byte{[]byte(v.Start), []byte(v.End)}, []byte("txn1"), option)
				require.NoError(t, err)
				vv, err := l.getLockTable(0, tableID)
				require.NoError(t, err)
				lt := vv.(*localLockTable)
				lt.mu.Lock()
				var keys []string
				lt.mu.store.Iter(func(b []byte, l Lock) bool {
					keys = append(keys, fmt.Sprintf("%s(%p)", string(b), l.holders))
					return true
				})
				lt.mu.Unlock()
			}
		},
	)
}

func TestRangeLockConflict(t *testing.T) {
	runLockServiceTests(
		t,
		[]string{"s1"},
		func(_ *lockTableAllocator, s []*service) {
			l := s[0]
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Second*1000)
			defer cancel()

			tableID := uint64(10)
			txnID1 := []byte{1}
			txnID2 := []byte{2}

			cases := []struct {
				rows        [][]byte
				g           pb.Granularity
				hasConflict bool
				ranges      [][]byte
			}{
				{
					rows:        [][]byte{{3}},
					g:           pb.Granularity_Row,
					hasConflict: false,
					ranges:      [][]byte{{1}, {2}},
				},
				{
					rows:        [][]byte{{3}},
					g:           pb.Granularity_Row,
					hasConflict: true,
					ranges:      [][]byte{{1}, {3}},
				},
				{
					rows:        [][]byte{{3}},
					g:           pb.Granularity_Row,
					hasConflict: true,
					ranges:      [][]byte{{1}, {4}},
				},
				{
					rows:        [][]byte{{3}},
					g:           pb.Granularity_Row,
					hasConflict: true,
					ranges:      [][]byte{{3}, {4}},
				},
				{
					rows:        [][]byte{{3}},
					g:           pb.Granularity_Row,
					hasConflict: false,
					ranges:      [][]byte{{4}, {5}},
				},
				{
					rows:        [][]byte{{3}, {5}},
					g:           pb.Granularity_Range,
					hasConflict: false,
					ranges:      [][]byte{{1}, {2}},
				},
				{
					rows:        [][]byte{{3}, {5}},
					g:           pb.Granularity_Range,
					hasConflict: true,
					ranges:      [][]byte{{1}, {3}},
				},
				{
					rows:        [][]byte{{3}, {5}},
					g:           pb.Granularity_Range,
					hasConflict: true,
					ranges:      [][]byte{{1}, {4}},
				},
				{
					rows:        [][]byte{{3}, {5}},
					g:           pb.Granularity_Range,
					hasConflict: true,
					ranges:      [][]byte{{3}, {4}},
				},
				{
					rows:        [][]byte{{3}, {5}},
					g:           pb.Granularity_Range,
					hasConflict: true,
					ranges:      [][]byte{{3}, {5}},
				},
				{
					rows:        [][]byte{{3}, {5}},
					g:           pb.Granularity_Range,
					hasConflict: true,
					ranges:      [][]byte{{3}, {6}},
				},
				{
					rows:        [][]byte{{3}, {5}},
					g:           pb.Granularity_Range,
					hasConflict: true,
					ranges:      [][]byte{{5}, {6}},
				},
				{
					rows:        [][]byte{{3}, {5}},
					g:           pb.Granularity_Range,
					hasConflict: false,
					ranges:      [][]byte{{6}, {7}},
				},
			}

			for _, c := range cases {
				mustAddTestLock(
					t,
					ctx,
					l,
					tableID,
					txnID1,
					c.rows,
					c.g)

				var wg sync.WaitGroup
				wg.Add(1)
				fn := func() {
					defer func() {
						require.NoError(t, l.Unlock(ctx, txnID2, timestamp.Timestamp{}))
						wg.Done()
					}()
					mustAddTestLock(
						t,
						ctx,
						l,
						tableID,
						txnID2,
						c.ranges,
						pb.Granularity_Range)
				}

				if !c.hasConflict {
					fn()
					require.NoError(t, l.Unlock(ctx, txnID1, timestamp.Timestamp{}))
				} else {
					go fn()
					waitWaiters(t, l, tableID, c.rows[0], 1)
					require.NoError(t, l.Unlock(ctx, txnID1, timestamp.Timestamp{}))
				}

				wg.Wait()
			}
		})
}

func TestLockedTSIsLastCommittedTS(t *testing.T) {
	runLockServiceTests(
		t,
		[]string{"s1"},
		func(_ *lockTableAllocator, s []*service) {
			l := s[0]
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Second*10)
			defer cancel()

			tableID := uint64(10)
			v, err := l.getLockTableWithCreate(0, tableID, nil, pb.Sharding_None)
			require.NoError(t, err)
			lt := v.(*localLockTable)
			lt.mu.Lock()
			lt.mu.tableCommittedAt = timestamp.Timestamp{PhysicalTime: 1}
			lt.mu.Unlock()

			txnID := []byte{1}
			mustAddTestLock(
				t,
				ctx,
				l,
				tableID,
				txnID,
				[][]byte{{1}},
				pb.Granularity_Row)
			require.NoError(t, l.Unlock(ctx, txnID, timestamp.Timestamp{PhysicalTime: 0}))
			lt.mu.Lock()
			require.Equal(t, timestamp.Timestamp{PhysicalTime: 1}, lt.mu.tableCommittedAt)
			lt.mu.Unlock()

			txnID = []byte{2}
			mustAddTestLock(
				t,
				ctx,
				l,
				tableID,
				txnID,
				[][]byte{{1}},
				pb.Granularity_Row)
			require.NoError(t, l.Unlock(ctx, txnID, timestamp.Timestamp{PhysicalTime: 2}))
			lt.mu.Lock()
			require.Equal(t, timestamp.Timestamp{PhysicalTime: 2}, lt.mu.tableCommittedAt)
			lt.mu.Unlock()

			txnID = []byte{3}
			mustAddTestLock(
				t,
				ctx,
				l,
				tableID,
				txnID,
				[][]byte{{1}},
				pb.Granularity_Row)
			require.NoError(t, l.Unlock(ctx, txnID, timestamp.Timestamp{PhysicalTime: 1}))
			lt.mu.Lock()
			require.Equal(t, timestamp.Timestamp{PhysicalTime: 2}, lt.mu.tableCommittedAt)
			lt.mu.Unlock()

			txnID = []byte{4}
			res, err := l.Lock(ctx, tableID, [][]byte{{1}}, txnID, pb.LockOptions{
				Granularity: pb.Granularity_Row,
				Mode:        pb.LockMode_Exclusive,
				Policy:      pb.WaitPolicy_Wait,
			})
			require.NoError(t, err)
			require.Equal(t, timestamp.Timestamp{PhysicalTime: 2}, res.Timestamp)
		})
}

func TestLockedTSIsLastCommittedTSWithRange(t *testing.T) {
	runLockServiceTests(
		t,
		[]string{"s1"},
		func(_ *lockTableAllocator, s []*service) {
			l := s[0]
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Second*10)
			defer cancel()

			tableID := uint64(10)
			v, err := l.getLockTableWithCreate(0, tableID, nil, pb.Sharding_None)
			require.NoError(t, err)
			lt := v.(*localLockTable)
			lt.mu.Lock()
			lt.mu.tableCommittedAt = timestamp.Timestamp{PhysicalTime: 1}
			lt.mu.Unlock()

			txnID := []byte{1}
			mustAddTestLock(
				t,
				ctx,
				l,
				tableID,
				txnID,
				[][]byte{{1}, {2}},
				pb.Granularity_Range)
			require.NoError(t, l.Unlock(ctx, txnID, timestamp.Timestamp{PhysicalTime: 0}))
			lt.mu.Lock()
			require.Equal(t, timestamp.Timestamp{PhysicalTime: 1}, lt.mu.tableCommittedAt)
			lt.mu.Unlock()

			txnID = []byte{2}
			mustAddTestLock(
				t,
				ctx,
				l,
				tableID,
				txnID,
				[][]byte{{1}, {2}},
				pb.Granularity_Range)
			require.NoError(t, l.Unlock(ctx, txnID, timestamp.Timestamp{PhysicalTime: 2}))
			lt.mu.Lock()
			require.Equal(t, timestamp.Timestamp{PhysicalTime: 2}, lt.mu.tableCommittedAt)
			lt.mu.Unlock()

			txnID = []byte{3}
			mustAddTestLock(
				t,
				ctx,
				l,
				tableID,
				txnID,
				[][]byte{{1}, {2}},
				pb.Granularity_Range)
			require.NoError(t, l.Unlock(ctx, txnID, timestamp.Timestamp{PhysicalTime: 1}))
			lt.mu.Lock()
			require.Equal(t, timestamp.Timestamp{PhysicalTime: 2}, lt.mu.tableCommittedAt)
			lt.mu.Unlock()

			txnID = []byte{4}
			res, err := l.Lock(ctx, tableID, [][]byte{{1}, {2}}, txnID, pb.LockOptions{
				Granularity: pb.Granularity_Range,
				Mode:        pb.LockMode_Exclusive,
				Policy:      pb.WaitPolicy_Wait,
			})
			require.NoError(t, err)
			require.Equal(t, timestamp.Timestamp{PhysicalTime: 2}, res.Timestamp)
		})
}

func Test15608(t *testing.T) {
	runLockServiceTests(
		t,
		[]string{"s1"},
		func(_ *lockTableAllocator, s []*service) {
			s1 := s[0]
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Second*10)
			defer cancel()

			option := newTestRowExclusiveOptions()
			rows := newTestRows(1)
			txn1 := newTestTxnID(1)
			txn2 := newTestTxnID(2)
			txn3 := newTestTxnID(3)
			table := uint64(10)

			// txn1 hold lock
			_, err := s1.Lock(ctx, table, rows, txn1, option)
			require.NoError(t, err, err)

			v, err := s1.getLockTable(0, table)
			require.NoError(t, err)
			lt := v.(*localLockTable)
			lt.options.beforeCloseFirstWaiter = func(c *lockContext) {
				c.txn.Unlock()
				defer c.txn.Lock()

				// txn3 hold lock
				_, err = s1.Lock(ctx, table, rows, txn3, option)
				require.NoError(t, err, err)
			}

			// txn2 wait for lock, is first waiter
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				option := newTestRowExclusiveOptions()
				_, _ = s1.Lock(ctx, table, rows, txn2, option)
			}()

			waitWaiters(t, s1, table, rows[0], 1)

			// unlock txn1 and txn2
			require.NoError(t, s1.Unlock(ctx, txn2, timestamp.Timestamp{}))
			require.NoError(t, s1.Unlock(ctx, txn1, timestamp.Timestamp{}))

			wg.Wait()

			checkLock(t, lt, rows[0], [][]byte{txn3}, nil, nil)
			require.NoError(t, s1.Unlock(ctx, txn3, timestamp.Timestamp{}))
		})
}

func TestLocalNeedUpgrade(t *testing.T) {
	runLockServiceTestsWithAdjustConfig(
		t,
		[]string{"s1"},
		time.Second*10,
		func(_ *lockTableAllocator, s []*service) {
			table := uint64(1)
			s1 := s[0]
			ctx, cancel := context.WithTimeout(context.Background(),
				time.Second*10)
			defer cancel()
			rows := newTestRows(1, 2, 3, 4, 5)
			txnID := newTestTxnID(1)
			_, err := s1.Lock(ctx, table, rows, txnID, pb.LockOptions{
				Granularity: pb.Granularity_Row,
				Mode:        pb.LockMode_Exclusive,
				Policy:      pb.WaitPolicy_Wait,
			})
			assert.Error(t, err)
			assert.True(t, moerr.IsMoErrCode(err, moerr.ErrLockNeedUpgrade))
		},
		func(c *Config) {
			c.MaxLockRowCount = 3
			c.MaxFixedSliceSize = 4
		},
	)
}

func TestCannotHungIfRangeConflictWithRowMultiTimes(t *testing.T) {
	runLockServiceTests(
		t,
		[]string{"s1"},
		func(_ *lockTableAllocator, s []*service) {
			l := s[0]
			ctx, cancel := context.WithTimeout(
				context.Background(),
				time.Second*10,
			)
			defer cancel()

			tableID := uint64(10)
			add := func(
				txn []byte,
				rows [][]byte,
				g pb.Granularity,
			) {
				mustAddTestLock(
					t,
					ctx,
					l,
					tableID,
					txn,
					rows,
					g,
				)
			}

			// workflow
			//
			// txn1 lock k4
			//      k4: holder(txn1) waiters()
			//
			// txn3 lock [k1, k4], wait at k4
			//      k4: holder(txn1) waiters(txn3)
			//
			// txn1 unlock, notify txn3 ------------------|
			//      k4: holder() waiters(txn3)            |
			//                                            |
			// txn2 lock range k2, k3                     |
			//      k2: holder(txn2) waiters()            |
			//      k3: holder(txn2) waiters()            |
			//      k4: holder() waiters(txn3)            |
			//                                            |
			// txn3 lock [k1, k4] retry, wait at k2 <-----|
			//      k2: holder(txn2) waiters(txn3)
			//      k3: holder(txn2) waiters()
			//      k4: holder() waiters(txn3)
			//
			// txn4 lock k2, wait at k2 --------------------------------|
			//      k2: holder(txn2) waiters(txn3, txn4)                |
			//      k3: holder(txn2) waiters()                          |
			//      k4: holder() waiters(txn3)                          |
			//                                                          |
			//                                                          |
			// txn2 unlock, notify txn3, txn4                           |
			//      k2: holder(txn2) waiters(txn3, txn4) -> deleted     |
			//      k3: holder(txn2) waiters()           -> deleted     |
			//      k4: holder() waiters(txn3)                          |
			//                                                          |
			// txn4 lock k2 retry <-------------------------------------|
			//      k2: holder(txn4) waiters(txn3)
			//      k4: holder() waiters(txn3)
			//
			// txn3 lock [k1, k4] retry, wait at k2
			//      k2: holder(txn4) waiters(txn3)
			//      k4: holder() waiters(txn3)
			//
			// txn4 lock k4, wait txn3
			//      k2: holder(txn4) waiters()
			//      k4: holder() waiters(txn3, txn4)

			// txn1 hold row1
			txn1 := []byte{1}
			txn2 := []byte{2}
			txn3 := []byte{3}
			txn4 := []byte{4}

			key2 := newTestRows(2)
			key4 := newTestRows(4)
			range23 := newTestRows(2, 3)
			range14 := newTestRows(1, 4)

			txn2Locked := make(chan struct{})
			txn4WaitAt2 := make(chan struct{})
			txn4GetLockAt1 := make(chan struct{})
			startTxn3 := make(chan struct{})
			txn3WaitAt2 := make(chan struct{})
			txn3WaitAt2Again := make(chan struct{})
			txn3WaitAt4 := make(chan struct{})
			txn3NotifiedAt4 := make(chan struct{})
			var once sync.Once

			// txn1 lock k4
			add(txn1, key4, pb.Granularity_Row)
			close(startTxn3)

			v, err := l.getLockTable(0, tableID)
			require.NoError(t, err)
			lt := v.(*localLockTable)
			txn3WaitTimes := 0
			lt.options.beforeWait = func(c *lockContext) func() {
				if bytes.Equal(c.txn.txnID, txn3) {
					return func() {
						if txn3WaitTimes == 0 {
							// txn3 wait at key4
							close(txn3WaitAt4)
							txn3WaitTimes++
							return
						}

						if txn3WaitTimes == 1 {
							close(txn3WaitAt2)
							txn3WaitTimes++
							return
						}

						if txn3WaitTimes == 2 {
							// step10: txn4 retry lock and wait at key2 again
							close(txn3WaitAt2Again)
							txn3WaitTimes++
							return
						}
					}
				}

				if bytes.Equal(c.txn.txnID, txn4) {
					return func() {
						once.Do(func() {
							close(txn4WaitAt2)
						})
					}
				}

				return func() {}
			}

			txn3NotifiedTimes := 0
			lt.options.afterWait = func(c *lockContext) func() {
				if bytes.Equal(c.txn.txnID, txn3) {
					return func() {
						if txn3NotifiedTimes == 0 {
							// txn1 closed and txn3 get notified
							close(txn3NotifiedAt4)
							txn3NotifiedTimes++
							<-txn2Locked
							return
						}

						if txn3NotifiedTimes == 1 {
							<-txn4GetLockAt1
						}
					}
				}
				return func() {}
			}

			var wg sync.WaitGroup
			wg.Add(5)

			go func() {
				defer wg.Done()
				<-startTxn3
				// txn3 lock range [k1, k4]
				add(txn3, range14, pb.Granularity_Range)
			}()

			go func() {
				defer wg.Done()
				<-txn3WaitAt4
				// txn1 unlock
				require.NoError(t, l.Unlock(ctx, txn1, timestamp.Timestamp{}))
			}()

			go func() {
				defer wg.Done()
				<-txn3NotifiedAt4
				// txn2 lock range [k3, k3]
				add(txn2, range23, pb.Granularity_Range)
				close(txn2Locked)
			}()

			go func() {
				defer wg.Done()
				<-txn3WaitAt2
				// txn4 lock k2
				add(txn4, key2, pb.Granularity_Row)
				close(txn4GetLockAt1)
				<-txn3WaitAt2Again

				// txn4 lock k4
				add(txn4, key4, pb.Granularity_Row)

				require.NoError(t, l.Unlock(ctx, txn4, timestamp.Timestamp{}))
			}()

			go func() {
				defer wg.Done()
				<-txn4WaitAt2
				require.NoError(t, l.Unlock(ctx, txn2, timestamp.Timestamp{}))
			}()

			wg.Wait()
		},
	)
}

func TestUnlockLockNotHeldByCurrentTxn(t *testing.T) {
	table := uint64(10)
	getRunner(false)(
		t,
		table,
		func(ctx context.Context, s *service, lt *localLockTable) {
			// Create two different transactions
			txn1 := newTestTxnID(1)
			txn2 := newTestTxnID(2)
			rows := newTestRows(1)

			// Add lock with txn1
			mustAddTestLock(
				t,
				ctx,
				s,
				table,
				txn1,
				rows,
				pb.Granularity_Row)

			// Create a cowSlice for the unlock call
			ls, err := newCowSlice(lt.fsp, rows)
			require.NoError(t, err)
			defer ls.close()

			// Get txn2 and add it to the active txn holder
			txn2Active := s.activeTxnHolder.getActiveTxn(txn2, true, "")
			require.NotNil(t, txn2Active)

			// Add the same bind to txn2's hold locks to ensure bind is not changed
			txn2Active.Lock()
			err = txn2Active.lockAdded(0, lt.bind, rows, lt.logger)
			require.NoError(t, err)
			txn2Active.Unlock()

			// Try to unlock with txn2
			// This should trigger the fatal error
			require.Panics(t, func() {
				lt.unlock(
					txn2Active,
					ls,
					timestamp.Timestamp{},
				)
			}, "should panic when trying to unlock a lock not held by current transaction")
		},
	)
}

type target struct {
	Start string `json:"start"`
	End   string `json:"end"`
}
