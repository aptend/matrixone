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
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/container/types"
	"github.com/matrixorigin/matrixone/pkg/objectio"
	"github.com/matrixorigin/matrixone/pkg/pb/timestamp"
	"github.com/matrixorigin/matrixone/pkg/vm/engine/tae/index"
	"github.com/stretchr/testify/require"
)

func TestBasicSimulator(t *testing.T) {
	player := NewSimPlayer()
	player.sexec.SetLogEnabled(true)
	player.Start()
	player.ResetPace(100*time.Millisecond, 120*time.Second)
	defer player.Stop()

	const K = 1024
	sid := objectio.NewSegmentid

	// add 4 lv0 small objects
	data := []SData{
		{
			stats: newTestObjectStats(t, 100, 200, 24*K, 50, 0, sid(), 0),
		},
		{
			stats: newTestObjectStats(t, 100, 200, 43*K, 50, 0, sid(), 0),
		},
		{
			stats: newTestObjectStats(t, 50, 300, 20*K, 50, 0, sid(), 0),
		},
		{
			stats: newTestObjectStats(t, 150, 400, 20*K, 50, 0, sid(), 0),
		},
	}

	tombstones := []STombstone{
		{
			SData: SData{
				stats: newTestObjectStats(t, 0, 0, 30*K, 20, 0, sid(), 0),
			},
			distro: map[objectio.ObjectId]int{
				data[1].stats.ObjectLocation().ObjectId(): 14,
				data[2].stats.ObjectLocation().ObjectId(): 6,
			},
		},
		{
			SData: SData{
				stats: newTestObjectStats(t, 0, 0, 50*K, 40, 0, sid(), 0),
			},
			distro: map[objectio.ObjectId]int{
				data[0].stats.ObjectLocation().ObjectId(): 8,
				data[3].stats.ObjectLocation().ObjectId(): 2,
				objectio.NewObjectid():                    20,
			},
		},
		{
			SData: SData{
				stats: newTestObjectStats(t, 0, 0, 3*K, 1, 0, sid(), 0),
			},
			distro: map[objectio.ObjectId]int{
				data[0].stats.ObjectLocation().ObjectId(): 1,
			},
		},

		{
			SData: SData{
				stats: newTestObjectStats(t, 0, 0, 3*K, 1, 0, sid(), 0),
			},
			distro: map[objectio.ObjectId]int{
				data[0].stats.ObjectLocation().ObjectId(): 1,
			},
		},
	}

	for _, data := range data {
		player.AddData(data)
	}
	for _, tombstone := range tombstones {
		player.AddTombstone(tombstone)
	}

	time.Sleep(3 * time.Second)
	t.Logf("report: %v", player.ReportString())

}

func constantCount(zms []index.ZM) int {
	constantZMCount := 0
	for _, zm := range zms {
		if IsConstantZM(zm) {
			constantZMCount++
		}
	}
	return constantZMCount
}

func TestSplitZM(t *testing.T) {
	zm := index.NewZM(types.T_int32, 0)
	zm.Update(int32(1))
	zm.Update(int32(20))
	{
		zmSplit := splitZM(zm, []int{1, 1, 1})
		require.Equal(t, 2, constantCount(zmSplit))
	}
	{
		zmSplit := splitZM(zm, []int{100, 100, 100})
		require.Equal(t, 0, constantCount(zmSplit))
	}
}

func TestUpdateStringTypeZM(t *testing.T) {
	zm := index.NewZM(types.T_varchar, 0)
	zm.Update([]byte("12345"))
	zm.Update([]byte("12346"))
	{
		zmSplit := splitZM(zm, []int{1, 1, 1})
		require.Equal(t, 2, constantCount(zmSplit))
	}

	{
		zmSplit := splitZM(zm, []int{100, 100, 100})
		require.Equal(t, 0, constantCount(zmSplit))
	}
}

func ExtractDataInput(filename string, beginTime time.Time) (sdata []SData, err error) {
	zoutFilePath := "/root/matrixone/zmtest/statement-info.out"
	ctx := context.Background()

	content, err := os.ReadFile(zoutFilePath)
	if err != nil {
		return nil, moerr.NewInternalErrorf(ctx, "Failed to read file: %v", err)
	}
	var data []map[string]any
	json.Unmarshal(content, &data)

	var line map[string]any
	baseTs := beginTime.UTC().UnixNano()
	firstSeqTime := int64(0)
	for i := range data {
		json.Unmarshal([]byte(data[i]["line"].(string)), &line)
		ts, _ := timestamp.ParseTimestamp(line["ts"].(string))
		if firstSeqTime == 0 {
			firstSeqTime = ts.PhysicalTime
		}
		statsbs, _ := base64.StdEncoding.DecodeString(line["stats"].(string))
		stat := objectio.ObjectStats(statsbs)

		sdata = append(sdata, SData{
			stats:      &stat,
			createTime: types.BuildTS(baseTs+ts.PhysicalTime-firstSeqTime, 0),
		})
	}
	return
}

func TestSimulatorOnStatementInfo(t *testing.T) {
	player := NewSimPlayer()
	player.ResetPace(100*time.Millisecond, 60*time.Second)

	filename := "/root/matrixone/zmtest/statement-info.out"
	sdata, err := ExtractDataInput(filename, player.sclok.Now())
	t.Log(sdata[0].createTime.ToTimestamp().ToStdTime(), sdata[len(sdata)-1].createTime.ToTimestamp().ToStdTime())
	require.NoError(t, err)

	player.SetEventSource(sdata, nil)
	player.Start()
	defer player.Stop()

	require.NoError(t, err)

	time.Sleep(5 * time.Second)

	t.Logf("report: %v", player.ReportString())
}
