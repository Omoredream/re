package jitoHTTP

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/consts"
)

func (node *RPC) test(ctx g.Ctx) (err error) {
	ctx = context.WithValue(ctx, consts.CtxRPC, node.Name())

	err = node.updateTipAccounts(ctx)
	if err != nil {
		err = gerror.Wrapf(err, "测试 RPC 连通性失败")
		return
	}

	node.tipAccountsMutex.RLock()
	defer node.tipAccountsMutex.RUnlock()

	g.Log().Debugf(ctx, "RPC 提供 %d 个小费账户", len(node.tipAccounts))

	return
}
