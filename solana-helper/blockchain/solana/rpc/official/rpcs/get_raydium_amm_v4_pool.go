package rpcs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	LP "git.wkr.moe/web3/solana-helper/blockchain/solana/lp"
	Parse "git.wkr.moe/web3/solana-helper/blockchain/solana/parse"
)

func (pool *RPCs) GetRaydiumAmmV4Pool(ctx g.Ctx, address Address.AccountAddress) (raydiumAmmV4Pool LP.RaydiumAmmV4Pool, err error) {
	accountInfo, err := pool.GetAccountInfo(ctx, address)
	if err != nil {
		err = gerror.Wrapf(err, "获取 Raydium AMM V4 Pool 失败")
		return
	}

	if accountInfo == nil {
		err = gerror.Newf("获取 Raydium AMM V4 Pool 为空")
		return
	}

	raydiumAmmV4Pool, err = Parse.ParseRaydiumAmmV4Pool(accountInfo)
	if err != nil {
		err = gerror.Wrapf(err, "解析 Raydium AMM V4 Pool 失败")
		return
	}

	return
}
