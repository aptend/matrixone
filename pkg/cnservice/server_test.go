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

package cnservice

import (
	"context"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matrixorigin/matrixone/pkg/bootstrap"
	"github.com/matrixorigin/matrixone/pkg/common/moerr"
	"github.com/matrixorigin/matrixone/pkg/common/morpc"
	"github.com/matrixorigin/matrixone/pkg/common/morpc/mock_morpc"
	"github.com/matrixorigin/matrixone/pkg/defines"
	"github.com/matrixorigin/matrixone/pkg/fileservice"
	"github.com/matrixorigin/matrixone/pkg/lockservice"
	"github.com/matrixorigin/matrixone/pkg/logservice"
	"github.com/matrixorigin/matrixone/pkg/pb/metadata"
	"github.com/matrixorigin/matrixone/pkg/pb/pipeline"
	qclient "github.com/matrixorigin/matrixone/pkg/queryservice/client"
	"github.com/matrixorigin/matrixone/pkg/txn/client"
	"github.com/matrixorigin/matrixone/pkg/udf"
	"github.com/matrixorigin/matrixone/pkg/util/address"
	"github.com/matrixorigin/matrixone/pkg/vm/engine"
)

func Test_InitServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &Config{
		UUID:     "dd1dccb4-4d3c-41f8-b482-5251dc7a41bf",
		PortBase: 18000,
	}

	srv := &service{
		metadata: metadata.CNStore{
			UUID: cfg.UUID,
		},
		cfg: cfg,
		responsePool: &sync.Pool{
			New: func() any {
				return &pipeline.Message{}
			},
		},
		addressMgr: address.NewAddressManager(cfg.ServiceHost, cfg.PortBase),
	}
	srv.addressMgr.Register(0)

	WithTaskStorageFactory(nil)(srv)
	handler := func(
		ctx context.Context,
		cnAddr string,
		message morpc.Message,
		cs morpc.ClientSession,
		engine engine.Engine,
		fs fileservice.FileService,
		lockService lockservice.LockService,
		queryClient qclient.QueryClient,
		hakeeper logservice.CNHAKeeperClient,
		udfService udf.Service,
		cli client.TxnClient,
		aicm *defines.AutoIncrCacheManager,
		mAcquirer func() morpc.Message) error {
		return nil
	}
	WithMessageHandle(handler)(srv)

	require.Equal(t, srv.ID(), cfg.UUID)
	require.Equal(t, srv.SQLAddress(), cfg.SQLAddress)

	msg := &pipeline.Message{}

	srv.releaseMessage(msg)
	message := srv.acquireMessage()
	require.Equal(t, message.(*pipeline.Message).Sid, msg.Sid)

	var err error
	ctx := context.TODO()
	session := mock_morpc.NewMockClientSession(ctrl)
	msg.Cmd = pipeline.Method_PipelineMessage
	session.EXPECT().CreateCache(ctx, uint64(0)).Return(&testMessageCache{}, nil).Times(2)

	msg.Sid = pipeline.Status_WaitingNext
	err = srv.handleRequest(
		ctx,
		morpc.RPCMessage{
			Ctx:     ctx,
			Cancel:  func() {},
			Message: msg,
		},
		0,
		session,
	)
	require.Nil(t, err)

	msg.Sid = pipeline.Status_Last
	err = srv.handleRequest(
		ctx,
		morpc.RPCMessage{
			Ctx:     ctx,
			Cancel:  func() {},
			Message: msg,
		},
		0,
		session,
	)
	require.Nil(t, err)
}

type testMessageCache struct {
	cache []morpc.Message
}

func (c *testMessageCache) Add(val morpc.Message) error {
	c.cache = append(c.cache, val)
	return nil
}

func (c *testMessageCache) Len() (int, error) {
	return len(c.cache), nil
}

func (c *testMessageCache) Pop() (morpc.Message, bool, error) {
	if len(c.cache) == 0 {
		return nil, false, nil
	}
	ret := c.cache[0]
	c.cache = c.cache[1:]
	return ret, true, nil
}

func (c *testMessageCache) Close() {
}

var _ bootstrap.Service = new(testBootService)

type testBootService struct {
	choice int
}

func (boot *testBootService) Bootstrap(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (boot *testBootService) BootstrapUpgrade(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (boot *testBootService) MaybeUpgradeTenant(ctx context.Context, tenantFetchFunc func() (int32, string, error), txnOp client.TxnOperator) (bool, error) {
	if boot.choice == 1 {
		return false, moerr.NewInternalErrorNoCtx("return_err")
	}
	return true, nil
}

func (boot *testBootService) UpgradeTenant(ctx context.Context, tenantName string, retryCount uint32, isALLAccount bool) (bool, error) {
	if boot.choice == 1 {
		return false, moerr.NewInternalErrorNoCtx("return_err")
	}
	return true, nil
}

func (boot *testBootService) GetFinalVersion() string {
	return "2.0.0"
}

func (boot *testBootService) GetFinalVersionOffset() int32 {
	//TODO implement me
	panic("implement me")
}

func (boot *testBootService) Close() error {
	//TODO implement me
	panic("implement me")
}

func Test_tenant(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	boot := &testBootService{}

	sv := &service{
		bootstrapService: boot,
	}

	err := sv.CheckTenantUpgrade(ctx, 3)
	assert.Nil(t, err)

	err = sv.UpgradeTenant(ctx, "acc3", 1, true)
	assert.Nil(t, err)

	boot.choice = 1
	err = sv.CheckTenantUpgrade(ctx, 3)
	assert.Error(t, err)

	err = sv.UpgradeTenant(ctx, "acc3", 1, true)
	assert.Error(t, err)
}
