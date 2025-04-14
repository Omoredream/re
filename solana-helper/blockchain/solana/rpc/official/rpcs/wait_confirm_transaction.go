package officialRPCs

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.wkr.moe/web3/solana-helper/consts"
)

// WaitConfirmTransactionByHTTP 若传入 tx, 则会自动每秒重发
func (pool *RPCs) WaitConfirmTransactionByHTTP(ctx g.Ctx, signature solana.Signature, tx ...*solana.Transaction) (spent time.Duration, err error) {
	ctx = context.WithValue(ctx, consts.CtxTransaction, signature.String())

	timeout := time.NewTimer(1 * time.Minute)
	startTime := gtime.Now()

	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-timeout.C:
			err = gerror.Newf("等待交易确认超时")
			return
		case <-time.After(1 * time.Second):
			var status *rpc.SignatureStatusesResult
			status, err = pool.httpGetTransactionStatus(ctx, signature)
			if err != nil {
				err = gerror.Wrapf(err, "等待交易上链失败")
				return
			}

			if status != nil {
				if status.Err != nil {
					err = gerror.Newf("链上错误: %v", status.Err)
					return
				}

				spent = gtime.Now().Sub(startTime)
				return
			} else if len(tx) > 0 && tx[0] != nil {
				_, err := pool.httpSendTransaction(ctx, tx[0], false)
				if err != nil {
					err = gerror.Wrapf(err, "发送交易失败")
					g.Log("background").Warningf(ctx, "%v", err)
				}
			}
		}
	}
}

func (pool *RPCs) WaitConfirmTransactionByWebSocket(ctx g.Ctx, signature solana.Signature) (err error) {
	ctx = context.WithValue(ctx, consts.CtxTransaction, signature.String())

	err = pool.websocketWaitForConfirmation(ctx, signature)
	if err != nil {
		err = gerror.Wrapf(err, "等待交易上链失败")
		return
	}

	return
}
