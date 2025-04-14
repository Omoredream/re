package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go/rpc"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

func (pool *RPCs) GetAddressTransactions(ctx g.Ctx, address Address.AccountAddress, amount *int) (transactions []*rpc.TransactionSignature, err error) {
	getSignaturesForAddressResult, err := pool.httpGetSignaturesForAddress(ctx, address, amount, nil, nil)
	if err != nil {
		if gerror.Equal(err, rpc.ErrNotFound) {
			err = nil
			return
		}
		err = gerror.Wrapf(err, "获取地址交易记录失败")
		return
	}

	transactions = getSignaturesForAddressResult

	return
}
