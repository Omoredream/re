package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
)

func (pool *RPCs) GetTransaction(ctx g.Ctx, signature solana.Signature) (transaction *utils.Transaction, err error) {
	getTransactionResult, err := pool.httpGetTransaction(ctx, signature)
	if err != nil {
		if gerror.Equal(err, rpc.ErrNotFound) {
			err = nil
			return
		}
		err = gerror.Wrapf(err, "获取交易失败")
		return
	}

	transaction_, err := getTransactionResult.Transaction.GetTransaction()
	if err != nil {
		err = gerror.Wrapf(err, "解析交易失败")
		return
	}

	transaction = &utils.Transaction{
		Transaction: *transaction_,
		Meta:        *getTransactionResult.Meta,
	}

	return
}
