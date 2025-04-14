package tests

import (
	"testing"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/utils"
)

func TestSnipe(t *testing.T) {
	var err error

	lp, err := officialPool.ParseRaydiumLP(ctx, solana.MustSignatureFromBase58("d6aMTVanmrGTVhe9SP5XaUVTpsVGh7sQsHNVV9PJTKiyANH4kTpRnjotPATq7kMALUDnVjnEWGHgGrPf7ejGmdV"))
	if err != nil {
		g.Log().Fatalf(ctx, "解析 LP 失败, %+v", err)
	}

	cost := decimal.NewFromFloat(0.1)
	income, err := officialPool.SnipeAtRaydium(ctx, lp, cost, decimal.NewFromFloat(0.01))
	if err != nil {
		g.Log().Fatalf(ctx, "狙击 LP 失败, %+v", err)
	}

	profit := income.Sub(cost)

	g.Log().Infof(ctx, "狙击 LP 获利 %s x%s", lp.QuoteToken.DisplayName(), Utils.UDColorSelf(profit))
}
