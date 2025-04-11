package rpcs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	LP "git.wkr.moe/web3/solana-helper/blockchain/solana/lp"
	Parse "git.wkr.moe/web3/solana-helper/blockchain/solana/parse"
)

func (pool *RPCs) GetRaydiumCpmmPool(ctx g.Ctx, address Address.AccountAddress) (raydiumCpmmPool LP.RaydiumCpmmPool, err error) {
	accountInfo, err := pool.GetAccountInfo(ctx, address)
	if err != nil {
		err = gerror.Wrapf(err, "获取 Raydium CPMM Pool 失败")
		return
	}

	if accountInfo == nil {
		err = gerror.Newf("获取 Raydium CPMM Pool 为空")
		return
	}

	raydiumCpmmPool, err = Parse.ParseRaydiumCpmmPool(accountInfo)
	if err != nil {
		err = gerror.Wrapf(err, "解析 Raydium CPMM Pool 失败")
		return
	}

	return
}
