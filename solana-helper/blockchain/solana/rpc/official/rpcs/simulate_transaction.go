package officialRPCs

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
)

func (pool *RPCs) SimulateTransaction(ctx g.Ctx, tx *solana.Transaction) (logs []string, cu uint64, err error) {
	txRaw, err := utils.SerializeTransactionBase64(ctx, tx, true)
	if err != nil {
		err = gerror.Wrapf(err, "编码交易失败")
		return
	}

	g.Log().Debugf(ctx, "构造交易: %s", tx.Signatures[0].String())
	g.Log().Debugf(ctx, "原始交易: %s", txRaw)
	//g.Log().Debugf(ctx, "可视化交易: %s", tx.String())

	ctx = context.WithValue(ctx, consts.CtxTransaction, tx.Signatures[0].String())

	status, err := pool.httpSimulateTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "模拟交易失败")
		return
	}

	logs = status.Logs
	cu = *status.UnitsConsumed

	return
}
