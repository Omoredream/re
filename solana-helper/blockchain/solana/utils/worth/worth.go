package worth

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/exchange/binance"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
)

func worth(ctx g.Ctx, symbol string, amount decimal.Decimal) (worth decimal.Decimal, err error) {
	price, err := Binance.GetPrice(ctx, symbol)
	if err != nil {
		err = gerror.Wrapf(err, "获取最新价格失败")
		return
	}

	worth = amount.Mul(price)

	return
}

func SOLWorth(ctx g.Ctx, amount decimal.Decimal) (decimal.Decimal, error) {
	return worth(ctx, "SOLUSDT", amount)
}

func USDCWorth(ctx g.Ctx, amount decimal.Decimal) (decimal.Decimal, error) {
	return worth(ctx, "USDCUSDT", amount)
}

func USDTWorth(ctx g.Ctx, amount decimal.Decimal) (decimal.Decimal, error) {
	return amount.Copy(), nil
}

func TokenWorth(ctx g.Ctx, amount decimal.Decimal, token Token.Token) (decimal.Decimal, error) {
	switch token.Address {
	case consts.SOL.Address:
		return SOLWorth(ctx, amount)
	case consts.USDC.Address:
		return USDCWorth(ctx, amount)
	case consts.USDT.Address:
		return USDTWorth(ctx, amount)
	default:
		return decimal.Zero, gerror.Newf("未知的交易所代币 %s", token.DisplayName())
	}
}

func IgnoreErr(amount decimal.Decimal, err error) decimal.Decimal {
	return amount
}
