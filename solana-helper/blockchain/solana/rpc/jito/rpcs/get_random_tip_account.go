package jitoRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

func (pool *RPCs) GetRandomTipAccount(ctx g.Ctx) (tipAccount Address.AccountAddress, err error) {
	tipAccount, err = pool.httpGetRandomTipAccount(ctx)
	if err != nil {
		err = gerror.Wrapf(err, "随机选择小费账户失败")
		return
	}

	return
}
