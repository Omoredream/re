package officialWebSocket

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"

	rpc_ws "github.com/gagliardetto/solana-go/rpc/ws"

	"git.wkr.moe/web3/solana-helper/rpcinterface"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/official/rpc/types"
)

type RPC struct {
	*rpcinterface.RPC
	url          string
	headers      g.MapStrStr
	shortID      bool
	client       *rpc_ws.Client
	reconnecting bool
}

func New(ctx g.Ctx, semaphore chan struct{}, name string, url string, cooldownIntervalMill int64, maxRunningThreads uint, headers g.MapStrStr, shortID bool) (node *RPC, err error) {
	if gstr.HasPrefix(url, "http://") {
		url = "ws://" + gstr.SubStr(url, 7)
	} else if gstr.HasPrefix(url, "https://") {
		url = "wss://" + gstr.SubStr(url, 8)
	} else if gstr.HasPrefix(url, "ws://") || gstr.HasPrefix(url, "wss://") {
	} else {
		err = gerror.Newf("未知的 WebSocket URL 格式")
		return
	}

	node = &RPC{
		RPC:     rpcinterface.New(string(officialRPCTypes.WebSocket)+"/"+name, cooldownIntervalMill, maxRunningThreads),
		url:     url,
		headers: headers,
		shortID: shortID,
	}

	ctx = context.WithValue(ctx, consts.CtxRPC, node.Name())

	err = Utils.Retry(ctx, func() (err error) {
		if semaphore != nil {
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()
		}
		err = Utils.Try(func() (err error) {
			node.client, err = node.connect(ctx)
			return
		}, nil, node.Success, node.Fail)
		if err != nil {
			return
		}

		err = Utils.Try(func() (err error) {
			//err = node.test(ctx)
			return
		}, nil, node.Success, node.Fail)
		if err != nil {
			return
		}

		return
	})
	if err != nil {
		err = gerror.Wrapf(err, "初始化 RPC 失败")
		return
	}

	return
}

func (node *RPC) Client() *rpc_ws.Client {
	return node.client
}
