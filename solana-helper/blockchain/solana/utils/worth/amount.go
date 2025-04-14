package worth

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/exchange/binance"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
)

func amount(ctx g.Ctx, symbol string, worth decimal.Decimal) (amount decimal.Decimal, err error) {
	price, err := Binance.GetPrice(ctx, symbol)
	if err != nil {
		err = gerror.Wrapf(err, "获取最新价格失败")
		return
	}

	amount = worth.Div(price)

	return
}

func SOLAmount(ctx g.Ctx, worth decimal.Decimal) (decimal.Decimal, error) {
	return amount(ctx, "SOLUSDT", worth)
}

func USDCAmount(ctx g.Ctx, worth decimal.Decimal) (decimal.Decimal, error) {
	return amount(ctx, "USDCUSDT", worth)
}

func USDTAmount(ctx g.Ctx, worth decimal.Decimal) (decimal.Decimal, error) {
	return worth.Copy(), nil
}

func TokenAmount(ctx g.Ctx, worth decimal.Decimal, token Token.Token) (decimal.Decimal, error) {
	switch token.Address {
	case consts.SOL.Address:
		return SOLAmount(ctx, worth)
	case consts.USDC.Address:
		return USDCAmount(ctx, worth)
	case consts.USDT.Address:
		return USDTAmount(ctx, worth)
	default:
		return decimal.Zero, gerror.Newf("未知的交易所代币 %s", token.DisplayName())
	}
}
