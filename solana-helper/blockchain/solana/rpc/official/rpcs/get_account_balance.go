package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func (pool *RPCs) getAccountBalance(ctx g.Ctx, address Address.AccountAddress) (balance decimal.Decimal, err error) {
	getBalanceResult, err := pool.httpGetBalance(ctx, address)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包余额失败")
		return
	}

	balance = lamports.Lamports2SOL(getBalanceResult.Value)

	return
}
