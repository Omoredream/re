package officialRPCs

import (
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
)

func (pool *RPCs) SubNewLP(ctx g.Ctx, lps chan utils.RaydiumLP) (err error) {
	signatures := make(chan solana.Signature, 0xff)
	defer close(signatures)
	go func() {
		for {
			err = pool.websocketSubAddressTransactions(ctx, consts.RaydiumCreatePoolChargingAddress.AccountAddress, signatures)
			if err != nil {
				g.Log().Errorf(ctx, "订阅 Raydium 创建 LP 抽成账户失败, %+v", err)
			}
		}
	}()
	for {
		signature := <-signatures

		go func() {
			var lp utils.RaydiumLP
			lp, err = pool.ParseRaydiumLP(ctx, signature)
			if err != nil {
				g.Log().Errorf(ctx, "解析 LP 失败, %+v", err)
				return
			}

			g.Log().Debugf(ctx, "新 LP: %s", lp.String())

			lps <- lp
		}()
	}
}
