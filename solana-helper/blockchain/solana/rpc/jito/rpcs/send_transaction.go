package jitoRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"
)

func (pool *RPCs) SendTransaction(ctx g.Ctx, tx *solana.Transaction, asBundle bool) (txHash solana.Signature, bundleId *string, err error) {
	txHash, bundleId, err = pool.httpSendTransaction(ctx, tx, asBundle)
	if err != nil {
		err = gerror.Wrapf(err, "广播交易失败")
		return
	}

	return
}
