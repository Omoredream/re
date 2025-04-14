package jitoHTTP

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/grand"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

func (node *RPC) updateTipAccounts(ctx g.Ctx) (err error) {
	ctx = context.WithValue(ctx, consts.CtxRPC, node.Name())

	tipAccounts, err := node.GetTipAccounts(ctx)
	if err != nil {
		err = gerror.Wrapf(err, "获取小费账户列表失败")
		return
	}

	node.tipAccountsMutex.LockFunc(func() {
		node.tipAccounts = make([]Address.AccountAddress, 0, len(tipAccounts))
		for _, tipAccount := range tipAccounts {
			node.tipAccounts = append(node.tipAccounts, Address.NewFromBase58(tipAccount))
		}
	})

	return
}

func (node *RPC) GetRandomTipAccount() (tipAccount Address.AccountAddress, err error) {
	node.tipAccountsMutex.RLock()
	defer node.tipAccountsMutex.RUnlock()

	if len(node.tipAccounts) == 0 {
		err = gerror.Newf("无可用的小费账户")
		return
	}

	tipAccount = node.tipAccounts[grand.Intn(len(node.tipAccounts))]

	return
}
