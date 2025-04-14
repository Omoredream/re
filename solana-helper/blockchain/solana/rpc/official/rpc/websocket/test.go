package officialWebSocket

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	rpc_ws "github.com/gagliardetto/solana-go/rpc/ws"

	"git.wkr.moe/web3/solana-helper/consts"
)

func (node *RPC) test(ctx g.Ctx) (err error) {
	ctx = context.WithValue(ctx, consts.CtxRPC, node.Name())

	subscribeBlockResult, err := node.client.BlockSubscribe(rpc_ws.NewBlockSubscribeFilterAll(), nil)
	if err != nil {
		err = gerror.Wrapf(err, "测试 RPC 连通性失败")
		return
	}
	defer subscribeBlockResult.Unsubscribe()

	resp, err := subscribeBlockResult.Recv()
	if err != nil {
		err = gerror.Wrapf(err, "接收区块订阅数据失败")
		return
	}

	if resp.Value.Err != nil {
		err = gerror.Newf("%+v", resp.Value.Err)
		return
	}

	g.Log().Debugf(ctx, "RPC 区块高度 %d, 哈希 %s", *resp.Value.Block.BlockHeight, resp.Value.Block.Blockhash)

	return
}
