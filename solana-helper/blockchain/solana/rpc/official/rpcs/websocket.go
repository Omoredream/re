package officialRPCs

import (
	"context"
	"sync"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"
	"github.com/gogf/gf/v2/os/gmutex"
	"github.com/samber/lo"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/sendAndConfirmTransaction"
	rpc_ws "github.com/gagliardetto/solana-go/rpc/ws"

	"git.wkr.moe/web3/solana-helper/errcode"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/official/rpc/websocket"
)

func (pool *RPCs) websocket(ctx g.Ctx, do func(ctx g.Ctx, websocketRPC *officialWebSocket.RPC) (err error)) (err error) {
	err = Utils.Retry(ctx, func() (err error) {
		chosen, err := pool.websocketRPCs.ChooseOne()
		if err != nil {
			err = gerror.WrapCodef(errcode.FatalError, err, "无法获取可用节点")
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

func (pool *RPCs) websocketAll(ctx g.Ctx, do func(ctx g.Ctx, websocketRPC *officialWebSocket.RPC) (err error)) (err error) {
	if pool.websocketRPCs.Count() == 0 {
		err = gerror.WrapCode(errcode.FatalError, err, "无法获取可用节点")
		return
	}

	wg := sync.WaitGroup{}
	errs := make(chan error, pool.websocketRPCs.Count())
	ctx, finish := context.WithCancel(ctx)
	pool.websocketRPCs.Each(func(chosen *officialWebSocket.RPC) {
		wg.Add(1)
		go func(chosen *officialWebSocket.RPC) {
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
						if gerror.HasCode(err, errcode.NeedReconnectError) {
							go chosen.Reconnect()
						}
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

func (pool *RPCs) websocketSubAddressTransactions(ctx g.Ctx, address Address.AccountAddress, signatures chan solana.Signature) (err error) {
	ctx = context.WithValue(ctx, consts.CtxAddress, address.String())

	received := gcache.NewWithAdapter(gcache.NewAdapterMemory())
	mu := &gmutex.Mutex{}

	err = pool.websocketAll(ctx, func(ctx g.Ctx, websocketRPC *officialWebSocket.RPC) (err error) {
		var sub *rpc_ws.LogSubscription
		sub, err = websocketRPC.Client().LogsSubscribeMentions(address.PublicKey, rpc.CommitmentProcessed)
		if err != nil {
			return
		}
		defer sub.Unsubscribe()

		for {
			select {
			case <-ctx.Done():
				err = gerror.NewCode(errcode.PreemptedError)
				return
			case <-time.After(1 * time.Minute):
				err = gerror.NewCode(errcode.NeedReconnectError, "WebSocket RPC 无响应超时")
				return
			case err = <-sub.Err():
				err = gerror.WrapCode(errcode.NeedReconnectError, err, "接收消息失败")
				return
			case got := <-sub.Response():
				transactionSignature := got.Value.Signature
				transactionHash := transactionSignature.String()
				if got.Value.Err != nil {
					g.Log().Debugf(ctx, "RPC 推送了一个失败的交易, %s, %+v", transactionHash, got.Value.Err)
					continue
				}

				g.Log().Debugf(ctx, "RPC 推送了交易, %s", transactionHash)

				var notReceived bool
				mu.LockFunc(func() { // SetIfNotExist 不是原子操作, 需要加锁
					notReceived, _ = received.SetIfNotExist(ctx, transactionHash, true, 1*time.Minute)
				})
				if notReceived {
					signatures <- transactionSignature
				}
			}
		}
	})

	return
}

func (pool *RPCs) websocketWaitForConfirmation(ctx g.Ctx, signature solana.Signature) (err error) {
	ctx = context.WithValue(ctx, consts.CtxTransaction, signature.String())

	err = pool.websocketAll(ctx, func(ctx g.Ctx, websocketRPC *officialWebSocket.RPC) (err error) {
		_, err = sendandconfirmtransaction.WaitForConfirmation(ctx, websocketRPC.Client(), signature, lo.ToPtr(1*time.Minute))
		return
	})

	return
}
