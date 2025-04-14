package jupiterRPCs

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/errcode"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/rpc"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/types"
)

func (pool *RPCs) MultiStep(ctx g.Ctx, targetTokens []string, do func(ctx g.Ctx) (err error)) (fingerprint string, err error) {
	chosen, ok := ctx.Value(consts.CtxValueJupiterHttpRPC).(*jupiterHTTP.RPC) // 提供固定使用某一节点的方案
	if !ok {
		chosen, err = pool.httpRPCs.ChooseOne(func(chosen *jupiterHTTP.RPC) bool {
			return chosen.Support(targetTokens)
		})
		if err != nil {
			err = gerror.WrapCodef(errcode.FatalError, err, "无法获取可用节点")
			return
		}
		defer chosen.EndThread()
	}
	fingerprint = chosen.Fingerprint()
	ctx = context.WithValue(ctx, consts.CtxValueJupiterHttpRPC, chosen)

	ctx = context.WithValue(ctx, consts.CtxRPC, chosen.Name())

	err = do(ctx)
	return
}

func (pool *RPCs) http(ctx g.Ctx, targetTokens []string, do func(ctx g.Ctx, httpRPC *jupiterHTTP.RPC) (err error)) (err error) {
	err = Utils.Retry(ctx, func() (err error) {
		chosen, ok := ctx.Value(consts.CtxValueJupiterHttpRPC).(*jupiterHTTP.RPC) // 提供固定使用某一节点的方案
		if !ok {
			chosen, err = pool.httpRPCs.ChooseOne(func(chosen *jupiterHTTP.RPC) bool {
				return chosen.Support(targetTokens)
			})
			if err != nil {
				err = gerror.WrapCode(errcode.FatalError, err, "无法获取可用节点")
				return
			}
			defer chosen.EndThread()
		}

		ctx = context.WithValue(ctx, consts.CtxRPC, chosen.Name())

		err = Utils.Try(func() (err error) {
			err = do(ctx, chosen)
			return
		}, nil, chosen.Success, chosen.Fail)
		if err != nil {
			return
		}

		return
	})

	return
}

func (pool *RPCs) httpAll(ctx g.Ctx, targetTokens []string, do func(ctx g.Ctx, httpRPC *jupiterHTTP.RPC) (err error)) (err error) {
	if pool.httpRPCs.Count() == 0 {
		err = gerror.WrapCode(errcode.FatalError, err, "无法获取可用节点")
		return
	}

	wg := sync.WaitGroup{}
	errs := make(chan error, pool.httpRPCs.Count())
	ctx, finish := context.WithCancel(ctx)
	pool.httpRPCs.Each(func(chosen *jupiterHTTP.RPC) {
		if !chosen.Support(targetTokens) {
			return
		}
		wg.Add(1)
		go func(chosen *jupiterHTTP.RPC) {
			defer wg.Done()
			ctx := context.WithValue(ctx, consts.CtxRPC, chosen.Name())
			err := Utils.Retry(ctx, func() (err error) {
				err = Utils.Try(func() (err error) {
					err = do(ctx, chosen)
					return
				}, func(err error) error {
					if gerror.HasCode(err, errcode.PreemptedError) || gerror.Is(err, context.Canceled) {
						return nil
					} else {
						return err
					}
				}, chosen.Success, chosen.Fail)
				if err != nil {
					return
				}

				return
			})
			errs <- err
			if err == nil {
				finish()
			}
		}(chosen)
	})

	go func() {
		wg.Wait() // 所有 RPC 都报错结束的情况
		finish()
	}()

	<-ctx.Done() // 有 RPC 正常结束的情况
	wg.Wait()

	close(errs)

	for err_ := range errs {
		if err_ != nil {
			err = err_
		}
	}

	return
}

func (pool *RPCs) httpGetRandomSupportToken(ctx g.Ctx, excludes ...string) (out string, err error) {
	err = pool.http(ctx, nil, func(ctx g.Ctx, httpRPC *jupiterHTTP.RPC) (err error) {
		out, err = httpRPC.GetRandomSupportToken(excludes...)
		return
	})

	return
}

func (pool *RPCs) httpGetQuote(ctx g.Ctx, inputMint string, outputMint string, amount uint64, optionalArgs ...jupiterTypes.QuoteOption) (out jupiterHTTP.GetQuoteResponse, err error) {
	err = pool.http(ctx, []string{inputMint, outputMint}, func(ctx g.Ctx, httpRPC *jupiterHTTP.RPC) (err error) {
		out, err = httpRPC.GetQuote(ctx, inputMint, outputMint, amount, optionalArgs...)
		return
	})

	return
}

func (pool *RPCs) httpCreateSwapTransaction(ctx g.Ctx, userPublicKey string, quoteResponse jupiterHTTP.GetQuoteResponse, optionalArgs ...jupiterTypes.SwapOption) (out jupiterHTTP.CreateSwapTransactionResponse, err error) {
	var targetTokens []string
	for _, routePlan := range quoteResponse.RoutePlan {
		targetTokens = append(targetTokens, routePlan.SwapInfo.InputMint, routePlan.SwapInfo.OutputMint)
	}
	targetTokens = gset.NewStrSetFrom(targetTokens).Slice()
	err = pool.http(ctx, targetTokens, func(ctx g.Ctx, httpRPC *jupiterHTTP.RPC) (err error) {
		out, err = httpRPC.CreateSwapTransaction(ctx, userPublicKey, quoteResponse, optionalArgs...)
		return
	})

	return
}
