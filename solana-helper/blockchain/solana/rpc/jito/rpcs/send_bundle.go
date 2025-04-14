package jitoRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"
)

func (pool *RPCs) SendBundle(ctx g.Ctx, txs ...*solana.Transaction) (bundleId string, err error) {
	bundleId, err = pool.httpSendBundle(ctx, txs)
	if err != nil {
		err = gerror.Wrapf(err, "广播捆绑交易失败")
		return
	}

	return
}
