package officialRPCs

import (
	"context"
	"errors"
	"net/url"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/samber/lo"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.wkr.moe/web3/solana-helper/errcode"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/official/rpc/http"
)

func (pool *RPCs) http(ctx g.Ctx, do func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error)) (err error) {
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
	})

	return
}

func (pool *RPCs) httpAll(ctx g.Ctx, do func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error)) (err error) {
	if pool.httpRPCs.Count() == 0 {
		err = gerror.WrapCode(errcode.FatalError, err, "无法获取可用节点")
		return
	}

	wg := sync.WaitGroup{}
	errs := make(chan error, pool.httpRPCs.Count())
	ctx, finish := context.WithCancel(ctx)
	pool.httpRPCs.Each(func(chosen *officialHTTP.RPC) {
		wg.Add(1)
		go func(chosen *officialHTTP.RPC) {
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

func (pool *RPCs) httpGetAccountInfo(ctx g.Ctx, address Address.AccountAddress) (out *rpc.GetAccountInfoResult, err error) {
	ctx = context.WithValue(ctx, consts.CtxAddress, address.String())

	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out, err = httpRPC.Client().GetAccountInfoWithOpts(ctx, address.PublicKey, &rpc.GetAccountInfoOpts{
			Commitment: rpc.CommitmentConfirmed,
		})
		if err != nil {
			if gerror.Equal(err, rpc.ErrNotFound) {
				err = gerror.WrapCode(errcode.FatalError, err)
			}
		}
		return
	})

	return
}

func (pool *RPCs) httpGetMultipleAccounts(ctx g.Ctx, addresses ...Address.AccountAddress) (out *rpc.GetMultipleAccountsResult, err error) {
	ctx = context.WithValue(ctx, consts.CtxAddress, addresses)

	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out, err = httpRPC.Client().GetMultipleAccountsWithOpts(ctx, lo.Map(addresses, func(address Address.AccountAddress, _ int) solana.PublicKey {
			return address.PublicKey
		}), &rpc.GetMultipleAccountsOpts{
			Commitment: rpc.CommitmentConfirmed,
		})
		return
	})

	return
}

func (pool *RPCs) httpGetBalance(ctx g.Ctx, address Address.AccountAddress) (out *rpc.GetBalanceResult, err error) {
	ctx = context.WithValue(ctx, consts.CtxWallet, address.String())

	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out, err = httpRPC.Client().GetBalance(ctx, address.PublicKey, rpc.CommitmentConfirmed)
		return
	})

	return
}

func (pool *RPCs) httpGetTokenAccountsByOwner(ctx g.Ctx, address Address.AccountAddress, conf *rpc.GetTokenAccountsConfig) (out *rpc.GetTokenAccountsResult, err error) {
	ctx = context.WithValue(ctx, consts.CtxWallet, address.String())

	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out, err = httpRPC.Client().GetTokenAccountsByOwner(ctx, address.PublicKey, conf, &rpc.GetTokenAccountsOpts{
			Commitment: rpc.CommitmentConfirmed,
		})
		return
	})

	return
}

func (pool *RPCs) httpGetTokenLargestAccounts(ctx g.Ctx, address Address.TokenAddress) (out *rpc.GetTokenLargestAccountsResult, err error) {
	ctx = context.WithValue(ctx, consts.CtxToken, address.String())

	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out, err = httpRPC.Client().GetTokenLargestAccounts(ctx, address.PublicKey, rpc.CommitmentConfirmed)
		return
	})

	return
}

func (pool *RPCs) httpGetLatestBlockhash(ctx g.Ctx) (out solana.Hash, err error) {
	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		if httpRPC.Version()[0] > 1 ||
			(httpRPC.Version()[0] == 1 && httpRPC.Version()[1] >= 9) {
			var out_ *rpc.GetLatestBlockhashResult
			out_, err = httpRPC.Client().GetLatestBlockhash(ctx, rpc.CommitmentConfirmed)
			if err != nil {
				return
			}

			out = out_.Value.Blockhash
		} else {
			var out_ *rpc.GetRecentBlockhashResult
			out_, err = httpRPC.Client().GetRecentBlockhash(ctx, rpc.CommitmentConfirmed)
			if err != nil {
				return
			}

			out = out_.Value.Blockhash
		}

		return
	})

	return
}

func (pool *RPCs) httpGetSlot(ctx g.Ctx) (out uint64, err error) {
	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out, err = httpRPC.Client().GetSlot(ctx, rpc.CommitmentConfirmed)
		return
	})

	return
}

func (pool *RPCs) httpGetTransaction(ctx g.Ctx, signature solana.Signature) (out *rpc.GetTransactionResult, err error) {
	ctx = context.WithValue(ctx, consts.CtxTransaction, signature.String())

	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out, err = httpRPC.Client().GetTransaction(ctx, signature, &rpc.GetTransactionOpts{
			MaxSupportedTransactionVersion: lo.ToPtr(uint64(0)),
			Commitment:                     rpc.CommitmentConfirmed,
		})
		return
	})

	return
}

func (pool *RPCs) httpGetTokenAccountBalance(ctx g.Ctx, address Address.TokenAccountAddress) (out *rpc.GetTokenAccountBalanceResult, err error) {
	ctx = context.WithValue(ctx, consts.CtxAccount, address.String())

	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out, err = httpRPC.Client().GetTokenAccountBalance(ctx, address.PublicKey, rpc.CommitmentConfirmed)
		return
	})

	return
}

func (pool *RPCs) httpGetSignaturesForAddress(ctx g.Ctx, address Address.AccountAddress, limit *int, before *solana.Signature, after *solana.Signature) (out []*rpc.TransactionSignature, err error) {
	ctx = context.WithValue(ctx, consts.CtxAccount, address.String())

	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out, err = httpRPC.Client().GetSignaturesForAddressWithOpts(ctx, address.PublicKey, &rpc.GetSignaturesForAddressOpts{
			Limit:      limit,
			Before:     lo.FromPtrOr(before, solana.Signature{}),
			Until:      lo.FromPtrOr(after, solana.Signature{}),
			Commitment: rpc.CommitmentConfirmed,
		})
		return
	})

	return
}

func (pool *RPCs) httpGetTransactionStatus(ctx g.Ctx, signature solana.Signature) (out *rpc.SignatureStatusesResult, err error) {
	ctx = context.WithValue(ctx, consts.CtxTransaction, signature.String())

	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out_, err := httpRPC.Client().GetSignatureStatuses(ctx, true, signature)
		if err != nil {
			return
		}

		if len(out_.Value) > 0 {
			out = out_.Value[0]
		}
		return
	})

	return
}

func (pool *RPCs) httpGetRentExemption(ctx g.Ctx, size uint64) (out uint64, err error) {
	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out, err = httpRPC.Client().GetMinimumBalanceForRentExemption(ctx, size, rpc.CommitmentConfirmed)
		return
	})

	return
}

func (pool *RPCs) httpSimulateTransaction(ctx g.Ctx, tx *solana.Transaction) (out *rpc.SimulateTransactionResult, err error) {
	err = pool.http(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out_, err := httpRPC.Client().SimulateTransactionWithOpts(ctx, tx, &rpc.SimulateTransactionOpts{
			SigVerify:              false,
			Commitment:             rpc.CommitmentProcessed,
			ReplaceRecentBlockhash: false,
		})
		if err != nil {
			return
		}

		if out_.Value.Err != nil {
			err = gerror.Newf("交易执行失败")
			if errs, ok := out_.Value.Err.(map[string]any); ok {
				if len(errs) > 0 {
					for errName, errDetail := range errs {
						if errName == "InstructionError" {
							if len(out_.Value.Logs) < 3 {
								err = gerror.Wrapf(err, "%s", gstr.Join(out_.Value.Logs, "\n"))
							} else {
								err = gerror.Wrapf(err, "%s\n%s", out_.Value.Logs[len(out_.Value.Logs)-3], out_.Value.Logs[len(out_.Value.Logs)-1])
							}
						} else {
							err = gerror.Wrapf(err, "%s: %v", errName, errDetail)
						}
					}
				}
			} else {
				err = gerror.Wrapf(err, "%v", out_.Value.Err)
			}
			if err != nil {
				err = gerror.WrapCode(errcode.FatalError, err)
				return
			}
		}

		out = out_.Value
		return
	})

	return
}

func (pool *RPCs) httpSendTransaction(ctx g.Ctx, tx *solana.Transaction, skipSimulate bool) (out solana.Signature, err error) {
	ctx = context.WithValue(ctx, consts.CtxTransaction, tx.Signatures[0].String())

	err = pool.httpAll(ctx, func(ctx g.Ctx, httpRPC *officialHTTP.RPC) (err error) {
		out_, err := httpRPC.Client().SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
			SkipPreflight:       skipSimulate,
			PreflightCommitment: rpc.CommitmentProcessed,
		})
		if err != nil {
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				err = gerror.Wrapf(urlErr.Err, "%s %q 失败", urlErr.Op, urlErr.URL)
				return
			}
			return
		}

		out = out_ // 防止速度较慢的节点替换早已成功的结果
		return
	})

	return
}
