package ProjectArb

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/gogf/gf/v2/container/gvar"
	"github.com/gogf/gf/v2/database/gredis"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/rpc"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/types"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func (arb *Arb) findProfitByJupiter(ctx context.Context, timepiece *Utils.Timepiece) (jupiterFp string, swapTx *solana.Transaction, tradeSize decimal.Decimal, profitAsSOL decimal.Decimal, foundSlot uint64, err error) {
	jupiterFp, err = arb.jupiterPool.MultiStep(ctx, []string{arb.tradeToken.Address.String()}, func(ctx g.Ctx) (err error) {
		jupiterHTTPRPCName := ctx.Value(consts.CtxRPC).(string)

		timepiece.StepStart("选择代币")
		tokenExcludes, err := arb.tokenBlacklists.HKeys(ctx, jupiterHTTPRPCName)
		if err != nil {
			err = gerror.Wrapf(err, "拉取中间代币黑名单失败")
			return
		}
		tokenExcludes = append(tokenExcludes, arb.tradeToken.Address.String())

		tokenMiddleAddressS, err := arb.jupiterPool.GetRandomSupportToken(ctx, tokenExcludes...)
		if err != nil {
			err = gerror.Wrapf(err, "随机选择中间代币失败")
			return
		}
		tokenMiddleAddress := Address.NewFromBase58(tokenMiddleAddressS).AsTokenAddress()

		if arb.debug {
			var tokenMiddle Token.Token
			tokenMiddle, err = arb.officialPool.TokenCacheGet(ctx, tokenMiddleAddress)
			if err != nil {
				err = gerror.Wrapf(err, "查询代币失败")
				return
			}
			ctx = context.WithValue(ctx, consts.CtxToken, tokenMiddle.DisplayName())
		} else {
			ctx = context.WithValue(ctx, consts.CtxToken, tokenMiddleAddress.String())
		}
		timepiece.StepFinish("选择代币")

		var quote *jupiterHTTP.GetQuoteResponse
		for tradeSizeRound := range arb.tradeSizeRounds {
			tradeSize = arb.tradeSize[tradeSizeRound]
			timepiece.StepStart("寻找A-B路由")
			var quote1 jupiterHTTP.GetQuoteResponse
			quote1, err = arb.jupiterPool.GetQuote(ctx, arb.tradeToken.Address, tokenMiddleAddress, lamports.Token2Lamports(tradeSize, arb.tradeToken.Info.Decimalx), arb.quoteOptions...)
			timepiece.StepFinish("寻找A-B路由")
			if err != nil {
				if gerror.HasCode(err, jupiterTypes.NoRoutesFound) ||
					gerror.HasCode(err, jupiterTypes.CouldNotFindAnyRoute) {
					_, err = arb.tokenBlacklists.HSet(ctx, jupiterHTTPRPCName, map[string]any{
						tokenMiddleAddress.String(): gtime.Now().String() + "|无路由",
					})
					if err != nil {
						err = gerror.Wrapf(err, "添加无路由的中间代币到黑名单失败")
						return
					}
					err = Utils.RedisHPExpire(ctx, arb.tokenBlacklists, jupiterHTTPRPCName, (10 * time.Second).Milliseconds(), tokenMiddleAddress.String())
					if err != nil {
						err = gerror.Wrapf(err, "代币设置临时黑名单有效期失败")
						return
					}

					if arb.debug {
						g.Log().Warningf(ctx, "无路由: %s", tokenMiddleAddress)
					}
					err = nil
					return
				} else if gerror.HasCode(err, jupiterTypes.TokenNotTradable) {
					_, err = arb.tokenBlacklists.HSet(ctx, jupiterHTTPRPCName, map[string]any{
						tokenMiddleAddress.String(): gtime.Now().String() + "|不可交易",
					})
					if err != nil {
						err = gerror.Wrapf(err, "添加不可交易的中间代币到黑名单失败")
						return
					}
					err = Utils.RedisHPExpire(ctx, arb.tokenBlacklists, jupiterHTTPRPCName, (10 * time.Second).Milliseconds(), tokenMiddleAddress.String())
					if err != nil {
						err = gerror.Wrapf(err, "代币设置临时黑名单有效期失败")
						return
					}

					if arb.debug {
						g.Log().Warningf(ctx, "不可交易: %s", tokenMiddleAddress)
					}
					err = nil
					return
				} else if gerror.HasCode(err, jupiterTypes.RoutePlanDoesNotConsumeAllTheAmount) {
					continue
				}
				err = gerror.Wrapf(err, "生成 A-B 交易意向失败")
				return
			}
			//g.Dump(quote1)

			timepiece.StepStart("寻找B-A路由")
			var quote2 jupiterHTTP.GetQuoteResponse
			quote2, err = arb.jupiterPool.GetQuote(ctx, tokenMiddleAddress, arb.tradeToken.Address, quote1.OutAmount, arb.quoteOptions...)
			timepiece.StepFinish("寻找B-A路由")
			if err != nil {
				if gerror.HasCode(err, jupiterTypes.NoRoutesFound) ||
					gerror.HasCode(err, jupiterTypes.CouldNotFindAnyRoute) {
					_, err = arb.tokenBlacklists.HSet(ctx, jupiterHTTPRPCName, map[string]any{
						tokenMiddleAddress.String(): gtime.Now().String() + "|无路由",
					})
					if err != nil {
						err = gerror.Wrapf(err, "添加无路由的中间代币到黑名单失败")
						return
					}
					err = Utils.RedisHPExpire(ctx, arb.tokenBlacklists, jupiterHTTPRPCName, (10 * time.Second).Milliseconds(), tokenMiddleAddress.String())
					if err != nil {
						err = gerror.Wrapf(err, "代币设置临时黑名单有效期失败")
						return
					}

					if arb.debug {
						g.Log().Warningf(ctx, "无路由: %s", tokenMiddleAddress)
					}
					err = nil
					return
				} else if gerror.HasCode(err, jupiterTypes.TokenNotTradable) {
					_, err = arb.tokenBlacklists.HSet(ctx, jupiterHTTPRPCName, map[string]any{
						tokenMiddleAddress.String(): gtime.Now().String() + "|不可交易",
					})
					if err != nil {
						err = gerror.Wrapf(err, "添加不可交易的中间代币到黑名单失败")
						return
					}
					err = Utils.RedisHPExpire(ctx, arb.tokenBlacklists, jupiterHTTPRPCName, (10 * time.Second).Milliseconds(), tokenMiddleAddress.String())
					if err != nil {
						err = gerror.Wrapf(err, "代币设置临时黑名单有效期失败")
						return
					}

					if arb.debug {
						g.Log().Warningf(ctx, "不可交易: %s", tokenMiddleAddress)
					}
					err = nil
					return
				} else if gerror.HasCode(err, jupiterTypes.RoutePlanDoesNotConsumeAllTheAmount) {
					continue
				}
				err = gerror.Wrapf(err, "生成 B-A 交易意向失败")
				return
			}
			//g.Dump(quote2)

			timepiece.StepStart("利润判断")
			balanceAfter := lamports.Lamports2Token(quote2.OutAmount, arb.tradeToken.Info.Decimalx)

			profit := balanceAfter.Sub(tradeSize)

			if !profit.IsPositive() {
				continue
			}

			if arb.tradeToken.Address != consts.SOL.Address { // todo
				//profitAsSOL, err = worth.SOLAmount(ctx, profit)
				//if err != nil {
				//	err = gerror.Wrapf(err, "计算币本位利润失败")
				//	return
				//}
				//profitAsSOL = profitAsSOL.Mul(0.98) // 给 SOL 价格下跌留足空间
				profitAsSOL = profit.Copy()
			} else {
				profitAsSOL = profit.Copy()
			}

			profitAsSOL = profitAsSOL.Sub(consts.SignFee) // swap 交易签名
			if arb.enableJito && arb.thirdTipPayer {
				profitAsSOL = profitAsSOL.Sub(consts.SignFee) // 代付交易签名
			}

			if !profitAsSOL.IsPositive() {
				continue
			}

			if (arb.enableSpam && profitAsSOL.LessThan(arb.spamProfitMin)) &&
				(arb.enableJito && profitAsSOL.LessThan(arb.jitoProfitMin)) {
				continue
			}
			timepiece.StepFinish("利润判断")

			timepiece.StepStart("组装路由")
			quote = new(jupiterHTTP.GetQuoteResponse)
			*quote = quote1
			quote.OutputMint = quote2.OutputMint
			quote.OutAmount = 0
			quote.OtherAmountThreshold = quote.OutAmount
			quote.RoutePlan = append(quote.RoutePlan, quote2.RoutePlan...)

			err = trimQuote(quote)
			if err != nil {
				err = gerror.Wrapf(err, "修剪交易意向失败")
				return
			}
			timepiece.StepFinish("组装路由")

			timepiece.StepStart("机会查重")
			quoteMap := fmt.Sprintf("%s:%s", decimals.DisplayBalance(profitAsSOL), quote.RoutePlan.String())
			var quoteMapExisted *gvar.Var
			quoteMapExisted, err = arb.txBlacklists.Set(ctx, quoteMap, jupiterHTTPRPCName, gredis.SetOption{
				TTLOption: gredis.TTLOption{
					PX: lo.ToPtr((800 * time.Millisecond).Milliseconds()), // 2 slots
				},
				NX: true,
			})
			if err != nil {
				err = gerror.Wrapf(err, "查询机会黑名单失败")
				return
			}
			timepiece.StepFinish("机会查重")
			if quoteMapExisted.IsEmpty() {
				quote = nil // 重复机会
				return
			}

			if arb.debug {
				g.Log().Info(ctx, quoteMap)
			}

			break
		}
		if err != nil {
			if gerror.HasCode(err, jupiterTypes.RoutePlanDoesNotConsumeAllTheAmount) {
				_, err = arb.tokenBlacklists.HSet(ctx, jupiterHTTPRPCName, map[string]any{
					tokenMiddleAddress.String(): gtime.Now().String() + "|数量超池",
				})
				if err != nil {
					err = gerror.Wrapf(err, "添加数量超池的中间代币到黑名单失败")
					return
				}
				err = Utils.RedisHPExpire(ctx, arb.tokenBlacklists, jupiterHTTPRPCName, (1 * time.Minute).Milliseconds(), tokenMiddleAddress.String())
				if err != nil {
					err = gerror.Wrapf(err, "代币设置临时黑名单有效期失败")
					return
				}

				if arb.debug {
					g.Log().Warningf(ctx, "数量超池: %s", tokenMiddleAddress)
				}
				err = nil
				return
			}
			err = gerror.Wrapf(err, "生成 A-B-A 交易意向失败")
			return
		}
		if quote == nil {
			return // 无套利机会
		}
		if quote.ContextSlot != nil {
			foundSlot = *quote.ContextSlot
		}

		timepiece.StepStart("生成交易")
		swapTx, err = arb.jupiterPool.CreateSwapTransaction(
			ctx,
			arb.wallet.Account.Address, *quote,
			jupiterTypes.SwapOptionUseWSOL,
			jupiterTypes.SwapOptionUseSelfTokenAccounts,             // 套利时无法开启
			jupiterTypes.SwapOptionComputeUnitPriceMicroLamports(0), // 不使用优先费
			jupiterTypes.SwapOptionDynamicComputeUnitLimit,          // 动态 CU 限制
		)
		if err != nil {
			if gerror.HasCode(err, jupiterTypes.CircularArbitrageIsDisabled) {
				g.Log().Warningf(ctx, "节点 %s 禁止套利", jupiterHTTPRPCName)
				return
			}
			err = gerror.Wrapf(err, "根据意向生成交易失败")
			return
		}
		timepiece.StepFinish("生成交易")

		return
	})
	if err != nil {
		err = gerror.Wrapf(err, "通过 Jupiter 寻找套利机会失败")
		return
	}

	return
}

type routeBalance struct {
	Balance uint64
	Max     uint64
	MaxAt   int
}

func trimQuote(quote *jupiterHTTP.GetQuoteResponse) (err error) {
	const tinyBalance uint64 = 5

	routeBalances := map[string]*routeBalance{
		quote.InputMint: {
			Balance: quote.InAmount,
			Max:     quote.InAmount,
			MaxAt:   -1,
		},
	}

	for i, routePlan := range quote.RoutePlan {
		routeBalances[routePlan.SwapInfo.InputMint].Balance -= routePlan.SwapInfo.InAmount

		if _, ok := routeBalances[routePlan.SwapInfo.OutputMint]; !ok {
			routeBalances[routePlan.SwapInfo.OutputMint] = &routeBalance{
				Balance: 0,
				Max:     0,
				MaxAt:   -1,
			}
		}
		routeBalances[routePlan.SwapInfo.OutputMint].Balance += routePlan.SwapInfo.OutAmount

		if Utils.MapFilterCount(routeBalances, func(balance *routeBalance) bool {
			return balance.Balance > tinyBalance
		}) > 1 { // 剩余多个代币
			continue
		} // 只剩1个代币, 说明已完成一轮完整的跳板

		// A-B-xxx 的 B, 或 A-B1-C-B2-xxx 的 B2(B2>B1)
		if routeBalances[routePlan.SwapInfo.OutputMint].Balance > routeBalances[routePlan.SwapInfo.OutputMint].Max {
			routeBalances[routePlan.SwapInfo.OutputMint].Max = routeBalances[routePlan.SwapInfo.OutputMint].Balance
			routeBalances[routePlan.SwapInfo.OutputMint].MaxAt = i
			continue // 只要发生增长即为预期路由, 无需精简, 继续遍历即可
		} // A-B1-C-B2-xxx 的 B2 存在无用功, 不如直接 A-B1-xxx

		if routeBalances[routePlan.SwapInfo.OutputMint].MaxAt == -1 {
			err = gerror.Newf("路径判断错误, %v", quote.RoutePlan)
			return
		}
		if quote.RoutePlan[routeBalances[routePlan.SwapInfo.OutputMint].MaxAt].SwapInfo.OutputMint != quote.RoutePlan[i].SwapInfo.OutputMint {
			err = gerror.Newf("路径解析错误, %v", quote.RoutePlan)
			return
		}

		slip := float64(routeBalances[routePlan.SwapInfo.OutputMint].Max) / float64(routeBalances[routePlan.SwapInfo.OutputMint].Balance)
		// 清除 [MaxAt+1,i]
		quote.RoutePlan = append(quote.RoutePlan[:routeBalances[routePlan.SwapInfo.OutputMint].MaxAt+1], quote.RoutePlan[i+1:]...)
		// 等比例放大
		for j := routeBalances[routePlan.SwapInfo.OutputMint].MaxAt + 1; j < len(quote.RoutePlan); j++ {
			quote.RoutePlan[j].SwapInfo.InAmount = uint64(math.Floor(float64(quote.RoutePlan[j].SwapInfo.InAmount) * slip))   // 尽量小, 避免Out-In出现负数
			quote.RoutePlan[j].SwapInfo.OutAmount = uint64(math.Ceil(float64(quote.RoutePlan[j].SwapInfo.OutAmount) * slip))  // 尽量大, 避免Out-In出现负数
			quote.RoutePlan[j].SwapInfo.FeeAmount = uint64(math.Round(float64(quote.RoutePlan[j].SwapInfo.FeeAmount) * slip)) // 对于计算来说不太重要
		}
		return trimQuote(quote)
	}
	if Utils.MapFilterCount(routeBalances, func(balance *routeBalance) bool {
		return balance.Balance > tinyBalance
	}) > 1 { // 剩余多个代币
		err = gerror.Newf("路径余额检查错误, %v", quote.RoutePlan)
		return
	}

	return
}
