package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/matrixorigin/matrixone/pkg/common/mpool"
	"github.com/matrixorigin/matrixone/pkg/container/batch"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/defines"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/logutil"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/util/toml"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/blockio"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/common"
)

func TestDecode3() (err error) {
	ctx := context.Background()
	name := "018f236b-5193-7ffd-8d48-6f11b26cfe51_00000"
	fsDir := "/data2/hanfeng/matrixone"
	mp, _ := mpool.NewMPool("test", 0, mpool.NoFixed)
	memCache := toml.ByteSize(1 << 20)
	c := fileservice.Config{
		Name:    defines.LocalFileServiceName,
		Backend: "DISK",
		DataDir: fsDir,
		Cache: fileservice.CacheConfig{
			MemoryCapacity: &memCache,
		},
	}

	service, err := fileservice.NewFileService(ctx, c, nil)
	if err != nil {
		return
	}

	uid, _ := types.ParseUuid("018f236b-5193-7ffd-8d48-6f11b26cfe51")
	sid := objectio.Segmentid(uid)
	objectName := objectio.BuildObjectName(&sid, 0)

	reader, err := blockio.NewFileReader(service, name)
	if err != nil {
		return
	}

	objmeta, err := reader.GetObjectReader().ReadAllMeta(ctx, mp)
	if err != nil {
		return
	}

	meta := objmeta.MustDataMeta()
	stat := fmt.Sprintf("%v blks, %v rows", meta.BlockCount(), meta.BlockHeader().Rows())
	extend := reader.GetObjectReader().GetMetaExtent()
	rows := int(meta.BlockHeader().Rows())
	infos := make([]objectio.BlockInfo, 0, meta.BlockCount())
	for i := 0; i < int(meta.BlockCount()); i++ {
		blockid := objectio.BuildObjectBlockid(objectName, uint16(i))
		brows := 8192
		if rows < 8192 {
			brows = rows
		}
		rows -= 8192
		loc := objectio.BuildLocation(objectName, *extend, uint32(brows), uint16(i))
		info := objectio.BlockInfo{
			BlockID:    *blockid,
			SegmentID:  objectName.SegmentId(),
			EntryState: false,
			Sorted:     true,
			MetaLoc:    objectio.ObjectLocation(loc),
		}
		infos = append(infos, info)
	}
	typs := []types.Type{
		types.T_uuid.ToType(),
		types.T_uuid.ToType(),
		types.T_uuid.ToType(),
		types.T_varchar.ToType(),
		types.T_int64.ToType(),
		types.T_varchar.ToType(), // user
		types.T_varchar.ToType(), // host
		types.T_varchar.ToType(), // database
		types.T_text.ToType(),    // stmt
		types.T_text.ToType(),    // stmt tag
		types.T_text.ToType(),    // stmt fg
		types.T_uuid.ToType(),
		types.T_varchar.ToType(),            // nodetype
		types.T_datetime.ToTypeWithScale(6), // request-at
		types.T_datetime.ToTypeWithScale(6), //
		types.T_uint64.ToType(),
		types.T_varchar.ToType(),
		types.T_text.ToType(), // error
		types.T_text.ToType(), // exec-plan
		types.T_int64.ToType(),
		types.T_int64.ToType(),
		types.T_text.ToType(),
		types.T_varchar.ToType(),
		types.T_varchar.ToType(),
		types.T_text.ToType(), // source type
		types.T_int64.ToType(),
		types.T_int64.ToType(), // result-cnt
	}
	cols := []uint16{}
	for i := 0; i < 27; i++ {
		// if i == 18 {
		// 	continue
		// }
		cols = append(cols, uint16(i))
	}
	// logutil.Infof("cols %v len len ", cols, len(typs), len(cols))

	read := func() {
		batches := []*batch.Batch{}
		inst := time.Now()
		for _, info := range infos {
			ins := time.Now()
			var bat *batch.Batch
			bat, err = blockio.BlockRead(
				ctx,
				&info,
				nil,
				cols,
				typs,
				types.BuildTS(1714286188916223690, 0).ToTimestamp(),
				nil,
				nil,
				nil,
				service,
				mp,
				nil,
				fileservice.Policy(0))
			if err != nil {
				return
			}
			fmt.Printf("  read blk batch cost %v\n----\n", time.Since(ins))
			batches = append(batches, bat)
		}
		logutil.Infof("%s: read %v batch cost %v", stat, len(batches), time.Since(inst))
	}

	read()

	return
}

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

	for _, bat := range bats {
		for _, vec := range bat.Vecs {
			logutil.Infof("vec is %v, len %v", vec.GetType(), vec.Length())
		}
		for i := 0; i < bat.Vecs[0].Length(); i++ {
			// for i := 0; i < 10; i++ {
			rowid := types.Rowid(bat.Vecs[0].GetRawBytesAt(i))
			committs := types.TS(bat.Vecs[1].GetRawBytesAt(i))
			t, _, _, _ := types.DecodeTuple(bat.Vecs[2].GetRawBytesAt(i))
			pkstr := t.ErrString(nil)
			if strings.HasPrefix(rowid.String(), "018e3c0c-3ed7-737b-befe-dbe524aac512-0-2") {
				fmt.Printf("line %v: rowid %v, ts %v, pkstr is %v\n", i, rowid.String(), committs.ToString(), pkstr)
			}
			// logutil.Infof("line %v: num is %v-%v, pkstr is %v", i, rowid.String(), committs.ToString(), pkstr)

		}
	}
}

func main() {
	// TestNewObjectReader2()
	// TestNewObjectReader1()
	err := TestDecode3()
	if err != nil {
		fmt.Println(err)
	}
}
