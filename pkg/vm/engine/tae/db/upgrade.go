package db

import (
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/catalog"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/containers"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/txn/txnimpl"
)

func collectThreeTableBatch(c *catalog.Catalog) (dBat, tBat, cBat *containers.Batch, err error) {
	newBatch := func(schema *catalog.Schema) *containers.Batch {
		bat := containers.NewBatch()
		typs := schema.Types()
		attrs := schema.Attrs()
		for i, attr := range attrs {
			v := containers.MakeVector(typs[i], common.CheckpointAllocator)
			bat.AddVector(attr, v)
		}
		return bat
	}

	dBat = newBatch(catalog.SystemDBSchema)
	tBat = newBatch(catalog.SystemTableSchema)
	cBat = newBatch(catalog.SystemColumnSchema)

	processer := &catalog.LoopProcessor{}
	processer.DatabaseFn = func(de *catalog.DBEntry) error {
		de.RLock()
		node := de.GetLatestCommittedNodeLocked()
		de.RUnlock()
		if node.HasDropCommitted() {
			return nil
		}
		for _, col := range catalog.SystemDBSchema.ColDefs {
			if col.Name == catalog.PhyAddrColumnName {
				continue
			}
			txnimpl.FillDBRow(de, node, col.Name, dBat.GetVectorByName(col.Name))
		}
		return nil
	}
	processer.TableFn = func(te *catalog.TableEntry) error {
		te.RLock()
		node := te.GetLatestCommittedNodeLocked()
		te.RUnlock()
		if node.HasDropCommitted() {
			return nil
		}
		for _, col := range catalog.SystemTableSchema.ColDefs {
			if col.Name == catalog.PhyAddrColumnName {
				continue
			}
			txnimpl.FillTableRow(te, node, col.Name, tBat.GetVectorByName(col.Name))
		}
		for _, col := range catalog.SystemColumnSchema.ColDefs {
			if col.Name == catalog.PhyAddrColumnName {
				continue
			}
			txnimpl.FillColumnRow(te, node, col.Name, cBat.GetVectorByName(col.Name))
		}
		return nil
	}
	if err = c.RecurLoop(processer); err != nil {
		cBat.Close()
		dBat.Close()
		tBat.Close()
		return
	}
	return
}
