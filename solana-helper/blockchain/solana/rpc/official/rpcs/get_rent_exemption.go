package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func (pool *RPCs) GetRentExemption(ctx g.Ctx, size uint64) (balance decimal.Decimal, err error) {
	getRentExemptionResult, err := pool.httpGetRentExemption(ctx, size)
	if err != nil {
		err = gerror.Wrapf(err, "获取最新区块失败")
		return
	}

	balance = lamports.Lamports2SOL(getRentExemptionResult)

	return
}
