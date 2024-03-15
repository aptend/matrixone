package main

import (
	"context"

	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/defines"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/blockio"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
)

func TestNewObjectReader2() {
	ctx := context.Background()
	name := "018e0838-be27-72c5-a166-aa6f120859ad_00000"

	fsDir := "/Users/aptend/code/matrixone"
	c := fileservice.Config{
		Name:    defines.LocalFileServiceName,
		Backend: "DISK",
		DataDir: fsDir,
	}
	service, err := fileservice.NewFileService(ctx, c, nil)
	reader, err := blockio.NewFileReader(service, name)
	if err != nil {
		return
	}
	// 这里需要  load meta 来看具体的信息
	bats, clear, err := reader.LoadAllColumns(ctx, []uint16{0, 1, 13}, common.DefaultAllocator)
	defer clear()
	if err != nil {
		logutil.Infof("load all columns failed: %v", err)
		return
	}
	/*name1, err := EncodeNameFromString(reader.GetName())
	assert.Nil(t, err)
	location := objectio.BuildLocation(name1, *reader.GetObjectReader().GetMetaExtent(), 51, 1)
	_, err = blockio.LoadTombstoneColumns(context.Background(), []uint16{0}, nil, service, location, nil)*/
	//applyDelete(bats[0], bb)
	// zm, err := reader.LoadZoneMaps(ctx, []uint16{0, 1, objectio.SEQNUM_COMMITTS}, 0, nil)
	// logutil.Infof("zm is %v-%v", zm[0].GetMax(), zm[0].GetMin())
	// bf, w, _ := reader.LoadOneBF(ctx, 0)
	// logutil.Infof("bf is %v, w is %v, err is %v", bf.String(), w, err)
	ts := types.TS{}
	for _, bat := range bats {
		for _, vec := range bat.Vecs {
			logutil.Infof("vec is %v", vec.GetType())
		}
		for i := 2600; i < bat.Vecs[0].Length(); i++ {
			//ts.Unmarshal(bats[0].Vecs[1].GetRawBytesAt(i))
			num := types.DecodeInt32(bat.Vecs[0].GetRawBytesAt(i))
			num1 := types.DecodeInt32(bat.Vecs[1].GetRawBytesAt(i))
			ts.Unmarshal(bat.Vecs[2].GetRawBytesAt(i))
			logutil.Infof("line %v: num is %d-%d, cmmit is %v", i, num, num1, ts.ToString())
		}
		//logutil.Infof("bats[0].Vecs[1].String() is %v", bat.Vecs[2].String())
	}
}

func TestNewObjectReader1() {
	ctx := context.Background()
	name := "018e3c0c-6ead-775a-8e65-faef3d690efa_00000"

	fsDir := "/Users/aptend/code/matrixone"
	c := fileservice.Config{
		Name:    defines.LocalFileServiceName,
		Backend: "DISK",
		DataDir: fsDir,
	}
	service, err := fileservice.NewFileService(ctx, c, nil)
	if err != nil {
		return
	}
	reader, err := blockio.NewFileReader(service, name)
	if err != nil {
		return
	}
	// dedicated deltaloc 读取 rowid, ts, pk
	bats, clear, err := reader.LoadDeleteAllColumns(ctx, []uint16{0, 1, 2}, common.DefaultAllocator)
	defer clear()
	if err != nil {
		logutil.Infof("load all columns failed: %v", err)
		return
	}

	for _, bat := range bats[:1] {
		for _, vec := range bat.Vecs {
			logutil.Infof("vec is %v", vec.GetType())
		}
		for i := 2600; i < bat.Vecs[0].Length(); i++ {
			//ts.Unmarshal(bats[0].Vecs[1].GetRawBytesAt(i))
			rowid := types.Rowid(bat.Vecs[0].GetRawBytesAt(i))
			committs := types.TS(bat.Vecs[1].GetRawBytesAt(i))
			t, _, _, _ := types.DecodeTuple(bat.Vecs[2].GetRawBytesAt(i))
			pkstr := t.ErrString(nil)
			logutil.Infof("line %v: num is %v-%v, cmmit is %v", i, rowid, committs, pkstr)
		}
		//logutil.Infof("bats[0].Vecs[1].String() is %v", bat.Vecs[2].String())
	}
}

func main() {
	// TestNewObjectReader2()
	TestNewObjectReader1()
}
