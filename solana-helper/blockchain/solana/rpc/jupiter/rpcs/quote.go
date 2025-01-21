package rpcs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	jupiterHTTP "git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/rpc"
)

func (pool *RPCs) GetQuote(ctx g.Ctx, inToken Address.TokenAddress, outToken Address.TokenAddress, inLamports uint64) (quote jupiterHTTP.GetQuoteResponse, err error) {
	quote, err = pool.httpGetQuote(ctx, inToken.String(), outToken.String(), inLamports)
	if err != nil {
		err = gerror.Wrapf(err, "获取报价失败")
		return
	}

	return
}
