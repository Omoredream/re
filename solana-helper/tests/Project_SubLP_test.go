package tests

import (
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gmutex"
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
)

func TestSubLP(t *testing.T) {
	var err error

	lps := make(chan utils.RaydiumLP, 0xff)
	defer close(lps)
	go func() {
		err = officialPool.SubNewLP(ctx, lps)
		if err != nil {
			g.Log().Fatalf(ctx, "订阅新 LP 失败, %+v", err)
		}
	}()

	txLock := &gmutex.Mutex{}
	sellTimes := int64(0)
	profitAmount := decimal.Zero

	logTimer := time.NewTicker(1 * time.Minute)
	for {
		select {
		case <-logTimer.C:
			g.Log().Infof(ctx, "冲 %d 次, 利润 $%s",
				sellTimes,
				Utils.UDColorSelf(profitAmount),
			)
		case lp := <-lps:
			go func() {
				if lp.QuoteToken.Address != consts.SOL.Address {
					return
				}

				cost := decimal.NewFromFloat(0.01)
				fee := decimal.NewFromFloat(0.00001)

				income, err := officialPool.SnipeAtRaydium(ctx, lp, cost, fee)
				if err != nil {
					g.Log().Errorf(ctx, "狙击 LP 失败, %+v", err)
					return
				}

				profit := income.Sub(cost)

				g.Log().Infof(ctx, "狙击 LP 获利 $%s", Utils.UDColorSelf(profit))

				txLock.LockFunc(func() {
					sellTimes++
					profitAmount = profitAmount.Add(profit.Sub(fee).Sub(fee))
				})
			}()
		}
	}
}
