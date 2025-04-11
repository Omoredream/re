package rpcs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	jupiterHTTP "git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/rpc"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
)

func (pool *RPCs) CreateSwapTransaction(ctx g.Ctx, wallet Address.AccountAddress, quote jupiterHTTP.GetQuoteResponse) (swapTx *solana.Transaction, err error) {
	swap, err := pool.httpCreateSwapTransaction(ctx, wallet.String(), quote)
	if err != nil {
		err = gerror.Wrapf(err, "创建兑换交易失败")
		return
	}

	swapTx, err = utils.DeserializeTransactionBase64(swap.SwapTransaction)
	if err != nil {
		err = gerror.Wrapf(err, "解析交易失败")
		return
	}

	return
}
