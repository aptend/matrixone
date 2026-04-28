// Copyright 2021 Matrix Origin
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

package logtail

import (
	"fmt"
	"time"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/catalog"
	"github.com/tidwall/btree"
	"go.uber.org/zap"
)

// moColumnsTableID matches catalog.MO_COLUMNS_ID. Hard-coded here to avoid an
// import cycle with the higher-level catalog package; the value is also used
// as a magic-id sentinel in handle.go (entry.ID > 3).
const moColumnsTableID = uint64(3)
const moCatalogDBID = uint64(1)

func (c *BoundTableOperator) isMoColumns() bool {
	return c.tbl != nil && c.tbl.ID == moColumnsTableID && c.tbl.GetDB() != nil && c.tbl.GetDB().ID == moCatalogDBID
}

type BoundTableOperator struct {
	from, to types.TS
	tbl      *catalog.TableEntry
	visitor  *TableLogtailRespBuilder

	dAScanCnt, dSScanCnt int
	tAScanCnt, tSScanCnt int
}

func (c *BoundTableOperator) Report() string {
	return fmt.Sprintf("dAScanCnt: %d, dSScanCnt: %d, tAScanCnt: %d, tSScanCnt: %d", c.dAScanCnt, c.dSScanCnt, c.tAScanCnt, c.tSScanCnt)
}

func (c *BoundTableOperator) recordReport(isTombstone bool, appendable bool) {
	if isTombstone {
		if appendable {
			c.tAScanCnt++
		} else {
			c.tSScanCnt++
		}
	} else {
		if appendable {
			c.dAScanCnt++
		} else {
			c.dSScanCnt++
		}
	}
}

// iterObject is allowed to yield false positive results, because ForeachMVCCNodeInRange will check the accuracy of the result.
func (c *BoundTableOperator) iterObject(from, to types.TS, isTombstone bool) error {
	var it btree.IterG[*catalog.ObjectEntry]
	if isTombstone {
		it = c.tbl.MakeTombstoneObjectIt()
	} else {
		it = c.tbl.MakeDataObjectIt()
	}
	key := &catalog.ObjectEntry{EntryMVCCNode: catalog.EntryMVCCNode{DeletedAt: to.Next()}}
	var ok bool
	if ok = it.Seek(key); !ok {
		ok = it.Last()
	}

	traceMC := c.isMoColumns()
	const slowVisit = 100 * time.Millisecond

	// after seeking, the first object could be out of the range, but false positive is allowed.
	earlybreak := false
	for ; ok; ok = it.Prev() {
		if earlybreak {
			break
		}
		obj := it.Item()
		c.recordReport(isTombstone, obj.GetAppendable())
		if obj.IsAppendable() && obj.IsCEntry() && obj.CreatedAt.LT(&from) {
			earlybreak = true
		}

		if next := obj.GetNextVersion(); obj.IsCEntry() && next != nil && next.DeletedAt.LE(&to) {
			continue
		}

		var visitStart time.Time
		if traceMC {
			visitStart = time.Now()
		}
		if err := c.visitor.VisitObj(obj); err != nil {
			return err
		}
		if traceMC {
			if d := time.Since(visitStart); d >= slowVisit {
				logutil.Info(
					"LAZY-CATALOG-MC-VISIT-SLOW",
					zap.Duration("duration", d),
					zap.Bool("tombstone", isTombstone),
					zap.Bool("appendable", obj.IsAppendable()),
					zap.Bool("centry", obj.IsCEntry()),
					zap.String("obj-id", obj.ID().String()),
					zap.String("created", obj.CreatedAt.ToString()),
					zap.String("deleted", obj.DeletedAt.ToString()),
					zap.String("from", from.ToString()),
					zap.String("to", to.ToString()),
				)
			}
		}
	}
	return nil
}

func (c *BoundTableOperator) Run() error {
	traceMC := c.isMoColumns()
	var t0, t1, t2, t3, t4 time.Time
	if traceMC {
		t0 = time.Now()
	}
	c.tbl.WaitDataObjectCommitted(c.to)
	if traceMC {
		t1 = time.Now()
	}
	if err := c.iterObject(c.from, c.to, false); err != nil {
		return err
	}
	if traceMC {
		t2 = time.Now()
	}
	c.tbl.WaitTombstoneObjectCommitted(c.to)
	if traceMC {
		t3 = time.Now()
	}
	if err := c.iterObject(c.from, c.to, true); err != nil {
		return err
	}
	if traceMC {
		t4 = time.Now()
		total := t4.Sub(t0)
		if total >= 200*time.Millisecond {
			logutil.Info(
				"LAZY-CATALOG-MC-OPERATOR-RUN",
				zap.Duration("total", total),
				zap.Duration("wait-data", t1.Sub(t0)),
				zap.Duration("iter-data", t2.Sub(t1)),
				zap.Duration("wait-tomb", t3.Sub(t2)),
				zap.Duration("iter-tomb", t4.Sub(t3)),
				zap.String("from", c.from.ToString()),
				zap.String("to", c.to.ToString()),
				zap.Int("dAScanCnt", c.dAScanCnt),
				zap.Int("dSScanCnt", c.dSScanCnt),
				zap.Int("tAScanCnt", c.tAScanCnt),
				zap.Int("tSScanCnt", c.tSScanCnt),
			)
		}
	}
	return nil
}
