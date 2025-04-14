package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go/rpc"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

func (pool *RPCs) GetAccountInfo(ctx g.Ctx, address Address.AccountAddress) (accountInfo []byte, err error) {
	getAccountInfoResult, err := pool.httpGetAccountInfo(ctx, address)
	if err != nil {
		if gerror.Is(err, rpc.ErrNotFound) {
			err = nil
			return
		}
		err = gerror.Wrapf(err, "获取账户信息失败")
		return
	}

	if getAccountInfoResult.Value == nil {
		return
	}

	accountInfo = getAccountInfoResult.Value.Data.GetBinary()

	return
}

func (pool *RPCs) getAccountsInfo(ctx g.Ctx, addresses []Address.AccountAddress) (accountsInfo [][]byte, err error) {
	if len(addresses) > 100 {
		err = gerror.Newf("批量获取账户信息输入长度超出限制")
		return
	}

	getMultipleAccountsResult, err := pool.httpGetMultipleAccounts(ctx, addresses...)
	if err != nil {
		err = gerror.Wrapf(err, "批量获取账户信息失败")
		return
	}

	if len(addresses) != len(getMultipleAccountsResult.Value) {
		err = gerror.Newf("批量获取账户信息输入输出长度不一致")
		return
	}

	accountsInfo = make([][]byte, len(addresses))
	for i := range getMultipleAccountsResult.Value {
		if getMultipleAccountsResult.Value[i] == nil {
			continue
		}

		accountsInfo[i] = getMultipleAccountsResult.Value[i].Data.GetBinary()
	}

	return
}
