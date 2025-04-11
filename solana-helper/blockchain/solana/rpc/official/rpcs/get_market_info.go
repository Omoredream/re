package rpcs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	ProgramSerum "github.com/gagliardetto/solana-go/programs/serum"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	Parse "git.wkr.moe/web3/solana-helper/blockchain/solana/parse"
)

func (pool *RPCs) getMarketInfo(ctx g.Ctx, address Address.AccountAddress) (market ProgramSerum.MarketV2, err error) {
	accountInfo, err := pool.GetAccountInfo(ctx, address)
	if err != nil {
		err = gerror.Wrapf(err, "获取市场信息失败")
		return
	}

	if accountInfo == nil {
		err = gerror.Newf("获取市场信息为空")
		return
	}

	market, err = Parse.ParseMarketInfo(accountInfo)
	if err != nil {
		err = gerror.Wrapf(err, "解析市场信息失败")
		return
	}

	return
}
