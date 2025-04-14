package ProjectArb

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/errcode"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jito/types"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
)

func (arb *Arb) Loop(ctx context.Context) {
	Utils.Parallel(ctx, 0, arb.parallel, func(_ int, thread int) (err error) {
		err = arb.arb(context.WithValue(ctx, consts.CtxDerivation, thread))
		if err != nil {
			err = gerror.Wrapf(err, `套利失败`)
			return
		}

		return
	})
}

func (arb *Arb) arb(ctx context.Context) (err error) {
	timepiece := Utils.NewTimepiece()

	jupiterFp, swapTx, tradeSize, profitAsSOL, foundSlot, err := arb.findProfitByJupiter(ctx, timepiece)
	if err != nil {
		if gerror.HasCode(err, errcode.CoolDownError) {
			g.Log().Warningf(ctx, "需要冷却: %v", err)
			time.Sleep(10 * time.Second)
		}
		if gerror.HasCode(err, errcode.NetworkError) {
			g.Log().Warningf(ctx, "Jupiter 节点网络异常, %v", err)
			err = gerror.WrapCode(errcode.IgnoreError, err)
		}
		err = gerror.Wrapf(err, "创建 swap 失败")
		return
	}
	if swapTx == nil {
		return
	}
	jupiterHTTPRPC := arb.jupiterPool.Get(jupiterFp)
	ctx = context.WithValue(ctx, consts.CtxRPC, jupiterHTTPRPC.Name())

	timepiece.StepStart("修改交易")
	spamTx, jitoTxs, err := arb.refactorTxs(ctx, swapTx, tradeSize, profitAsSOL, foundSlot)
	if err != nil {
		err = gerror.Wrapf(err, "重构交易失败")
		return
	}
	timepiece.StepFinish("修改交易")

	var wg sync.WaitGroup

	if arb.enableSpam && spamTx != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			timepiece.StepStart("发送 spam 交易")
			var txId solana.Signature
			txId, err = arb.officialPool.SendTransaction(ctx, spamTx, !arb.simulate)
			if err != nil {
				err = gerror.Wrapf(err, "广播交易失败")
				return
			}
			ctx = context.WithValue(ctx, consts.CtxTransaction, txId)
			sendSpentBySpam := timepiece.StepFinish("发送 spam 交易")
			if arb.metrics {
				timepiece.StepStart("spam 性能指标上报")
				_, err = arb.arbMetricsTS.Do(ctx, "TS.ADD", "sendSpentBySpam", "*", float64(sendSpentBySpam.Microseconds())/1e3)
				if err != nil {
					err = gerror.Wrapf(err, "上报性能指标失败")
					return
				}
				timepiece.StepFinish("spam 性能指标上报")
			}
		}()
	}

	if arb.enableJito && len(jitoTxs) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if arb.simulate {
				timepiece.StepStart("模拟 jito 交易")
				_, _, err = arb.officialPool.SimulateTransaction(ctx, jitoTxs[0])
				if err != nil {
					if arb.debug {
						g.Log().Warningf(ctx, "交易模拟不通过, %v", err)
					}
					err = nil
					return
				}
				timepiece.StepFinish("模拟 jito 交易")
			}

			timepiece.StepStart("发送 jito 交易")
			var bundleId string
			bundleId, err = arb.jitoPool.SendBundle(ctx, jitoTxs...)
			if err != nil {
				if gerror.HasCode(err, jitoTypes.ErrBlockhashExpired) {
					g.Log().Warningf(ctx, "Jupiter Swap API 区块滞后, %v", err)
					jupiterHTTPRPC.AddCoolDown((10 * time.Minute).Milliseconds())
					err = nil
					return
				} else if gerror.HasCode(err, jitoTypes.ErrRateLimited) {
					if arb.metrics {
						timepiece.StepStart("jito 性能指标上报")
						_, err = arb.arbMetricsTS.Do(ctx, "TS.ADD", "jitoRateLimitTimes", "*", 1)
						if err != nil {
							err = gerror.Wrapf(err, "上报性能指标失败")
							return
						}
						timepiece.StepFinish("jito 性能指标上报")
					}
				}
				err = gerror.Wrapf(err, "广播捆绑交易失败")
				return
			}
			ctx = context.WithValue(ctx, consts.CtxBundle, bundleId)
			sendSpentByJito := timepiece.StepFinish("发送 jito 交易")
			if arb.metrics {
				timepiece.StepStart("jito 性能指标上报")
				_, err = arb.arbMetricsTS.Do(ctx, "TS.ADD", "sendSpentByJito", "*", float64(sendSpentByJito.Microseconds())/1e3)
				if err != nil {
					err = gerror.Wrapf(err, "上报性能指标失败")
					return
				}
				timepiece.StepFinish("jito 性能指标上报")
			}
		}()
	}

	wg.Wait()
	g.Log().Noticef(ctx, "已发送交易, 流程耗时 %s, 预期利润 SOL %s", timepiece.Report(), decimals.DisplayBalance(profitAsSOL))

	return
}
