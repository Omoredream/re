package officialRPCs

import (
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
)

// GetTokenRaydiumPrice
// price: 以 QuoteToken 为单位的代币价格
// liquidity: 以 QuoteToken 为单位的流动性
// fdv: 以 QuoteToken 为单位的市值
func (pool *RPCs) GetTokenRaydiumPrice(ctx g.Ctx, lp utils.RaydiumLP) (price, liquidity, fdv decimal.Decimal, err error) {
	balances, err := pool.GetTokenAccountsBalance(ctx, []Address.TokenAccountAddress{lp.BaseVault, lp.QuoteVault}, []Token.Token{lp.BaseToken, lp.QuoteToken})
	if err != nil {
		err = gerror.Wrapf(err, "获取 LP 左向/右向现价失败")
		return
	}

	token := balances[0]
	liquidity = balances[1]

	price = liquidity.Div(token)

	fdv = lp.BaseToken.Info.Supply.Mul(price)

	return
}

func (pool *RPCs) swapAtRaydium() {
	// todo
}

func (pool *RPCs) buyAtRaydium(ctx g.Ctx, lp utils.RaydiumLP, cost decimal.Decimal, wanna decimal.Decimal, fee decimal.Decimal) (got decimal.Decimal, err error) {
	// todo
	return wanna, nil
}

func (pool *RPCs) sellAtRaydium(ctx g.Ctx, lp utils.RaydiumLP, cost decimal.Decimal, wanna decimal.Decimal, fee decimal.Decimal) (got decimal.Decimal, err error) {
	// todo
	return wanna, nil
}

func (pool *RPCs) SnipeAtRaydium(ctx g.Ctx, lp utils.RaydiumLP, cost decimal.Decimal, fee decimal.Decimal) (income decimal.Decimal, err error) {
	if lp.QuoteToken.Address != consts.SOL.Address && lp.QuoteToken.Address != consts.USDC.Address && lp.QuoteToken.Address != consts.USDT.Address {
		err = gerror.Newf("不支持的右向代币 %s", lp.QuoteToken.DisplayName())
		return
	}

	g.Log().Debugf(ctx, "LP: %s, 开盘时间 %s, 开盘价格 %s x%s, 开盘流动性 $%s, 开盘市值 $%s",
		lp.String(),
		lp.OpenTime.String(),
		lp.QuoteToken.DisplayName(),
		decimals.DisplayBalance(lp.InitialPrice),
		decimals.DisplayBalance(lp.InitialLiquidity),
		decimals.DisplayBalance(lp.InitialFdv),
	)

	time.Sleep(lp.OpenTime.Sub(gtime.Now()))

	buyPrice, buyLiquidity, buyFdv, err := pool.GetTokenRaydiumPrice(ctx, lp)
	if err != nil {
		err = gerror.Wrapf(err, "获取 LP 当前价格失败")
		return
	}

	g.Log().Infof(ctx, "LP: %s, 现价 %s x%s, 流动性 $%s, 市值 $%s",
		lp.String(),
		lp.QuoteToken.DisplayName(),
		Utils.UDColorDiff(lp.InitialPrice, buyPrice),
		Utils.UDColorDiff(lp.InitialLiquidity, buyLiquidity),
		Utils.UDColorDiff(lp.InitialFdv, buyFdv),
	)

	wanna := cost.Div(buyPrice)

	got, err := pool.buyAtRaydium(ctx, lp, cost, wanna, fee)
	if err != nil {
		err = gerror.Wrapf(err, "买入失败")
		return
	}

	g.Log().Infof(ctx, "花费 %s x%s (+ 手续费 %s x%s) 买入 %s x%s",
		lp.QuoteToken.DisplayName(),
		decimals.DisplayBalance(cost),
		consts.SOL.DisplayName(),
		decimals.DisplayBalance(fee),
		lp.BaseToken.DisplayName(),
		decimals.DisplayBalance(got),
	)

	time.Sleep(40 * time.Second)

	sellPrice, sellLiquidity, sellFdv, err := pool.GetTokenRaydiumPrice(ctx, lp)
	if err != nil {
		err = gerror.Wrapf(err, "获取 LP 当前价格失败")
		return
	}

	g.Log().Infof(ctx, "LP: %s, 现价 %s x%s, 流动性 $%s, 市值 $%s",
		lp.String(),
		lp.QuoteToken.DisplayName(),
		Utils.UDColorDiff(buyPrice, sellPrice),
		Utils.UDColorDiff(buyLiquidity, sellLiquidity),
		Utils.UDColorDiff(buyFdv, sellFdv),
	)

	wanna = got.Mul(sellPrice)

	income, err = pool.sellAtRaydium(ctx, lp, got, wanna, fee)
	if err != nil {
		err = gerror.Wrapf(err, "卖出失败")
		return
	}

	g.Log().Infof(ctx, "卖出 %s x%s (+ 手续费 %s x%s) 收入 %s x%s",
		lp.BaseToken.DisplayName(),
		decimals.DisplayBalance(got),
		consts.SOL.DisplayName(),
		decimals.DisplayBalance(fee),
		lp.QuoteToken.DisplayName(),
		Utils.UDColorDiff(cost, income),
	)

	return
}
