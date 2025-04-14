package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/lp"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/parser"
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

	raydiumCpmmPool, err = Parser.ParseRaydiumCpmmPool(accountInfo)
	if err != nil {
		err = gerror.Wrapf(err, "解析 Raydium CPMM Pool 失败")
		return
	}

	return
}
