package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"
)

func (pool *RPCs) GetLatestBlockhash(ctx g.Ctx) (blockhash solana.Hash, err error) {
	blockhash, err = pool.httpGetLatestBlockhash(ctx)
	if err != nil {
		err = gerror.Wrapf(err, "获取最新区块失败")
		return
	}

	return
}
