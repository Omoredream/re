package jitoRPCs

import (
	"context"
	"sync"

	"github.com/avast/retry-go/v4"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/errcode"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jito/rpc"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jito/types"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
)

func (pool *RPCs) http(ctx g.Ctx, do func(ctx g.Ctx, httpRPC *jitoHTTP.RPC) (err error)) (err error) {
	err = Utils.Retry(ctx, func() (err error) {
		chosen, err := pool.httpRPCs.ChooseOne()
		if err != nil {
			err = gerror.WrapCode(errcode.FatalError, err, "无法获取可用节点")
			return
		}
		defer chosen.EndThread()

		ctx = context.WithValue(ctx, consts.CtxRPC, chosen.Name())

		err = Utils.Try(func() (err error) {
			err = do(ctx, chosen)
			return
		}, nil, chosen.Success, chosen.Fail)
		if err != nil {
			return
		}

		return
	}, retry.Delay(0))

	return
}

func (pool *RPCs) httpAll(ctx g.Ctx, do func(ctx g.Ctx, httpRPC *jitoHTTP.RPC) (err error)) (err error) {
	if pool.httpRPCs.Count() == 0 {
		err = gerror.WrapCode(errcode.FatalError, err, "无法获取可用节点")
		return
	}

	wg := sync.WaitGroup{}
	errs := make(chan error, pool.httpRPCs.Count())
	ctx, finish := context.WithCancel(ctx)
	pool.httpRPCs.Each(func(chosen *jitoHTTP.RPC) {
		wg.Add(1)
		go func(chosen *jitoHTTP.RPC) {
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

func (pool *RPCs) httpGetRandomTipAccount(ctx g.Ctx) (out Address.AccountAddress, err error) {
	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *jitoHTTP.RPC) (err error) {
		out, err = httpRPC.GetRandomTipAccount()
		return
	})

	return
}

func (pool *RPCs) httpSendTransaction(ctx g.Ctx, tx *solana.Transaction, asBundle bool) (out1 solana.Signature, out2 *string, err error) {
	ctx = context.WithValue(ctx, consts.CtxTransaction, tx.Signatures[0].String())
	txRaw, err := utils.SerializeTransactionBase64(ctx, tx, true)
	if err != nil {
		err = gerror.Wrapf(err, "编码交易失败")
		return
	}

	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *jitoHTTP.RPC) (err error) {
		out1_, out2_, err := httpRPC.SendTransaction(ctx, txRaw, asBundle)
		if err != nil {
			return
		}

		out1, err = solana.SignatureFromBase58(out1_)
		if err != nil {
			err = gerror.Wrapf(err, "反序列化交易ID失败")
			return
		}

		out2 = out2_
		return
	})

	return
}

func (pool *RPCs) httpSendBundle(ctx g.Ctx, txs []*solana.Transaction) (out string, err error) {
	txRaws := make([]string, 0, len(txs))
	for _, tx := range txs {
		var txRaw string
		txRaw, err = utils.SerializeTransactionBase64(ctx, tx, true)
		if err != nil {
			err = gerror.Wrapf(err, "编码交易失败")
			return
		}
		txRaws = append(txRaws, txRaw)
	}

	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *jitoHTTP.RPC) (err error) {
		out, err = httpRPC.SendBundle(ctx, txRaws)
		return
	})

	return
}

func (pool *RPCs) httpGetBundlesStatus(ctx g.Ctx, bundleIds ...string) (out []*jitoTypes.GetBundleStatusesResponse, err error) {
	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *jitoHTTP.RPC) (err error) {
		out, err = httpRPC.GetBundleStatuses(ctx, bundleIds)
		return
	})

	return
}

func (pool *RPCs) httpGetInflightBundlesStatus(ctx g.Ctx, bundleIds ...string) (out []jitoTypes.GetInflightBundleStatusesResponse, err error) {
	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *jitoHTTP.RPC) (err error) {
		out, err = httpRPC.GetInflightBundleStatuses(ctx, bundleIds)
		return
	})

	return
}
