package rpcs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	Utils "git.wkr.moe/web3/solana-helper/utils"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	Parse "git.wkr.moe/web3/solana-helper/blockchain/solana/parse"
)

func (pool *RPCs) getTokenAccountOwner(ctx g.Ctx, address Address.TokenAccountAddress) (owner Address.AccountAddress, err error) {
	accountInfo, err := pool.GetAccountInfo(ctx, address.AccountAddress)
	if err != nil {
		err = gerror.Wrapf(err, "获取代币账户信息失败")
		return
	}

	if accountInfo == nil {
		err = gerror.Newf("获取代币账户信息为空")
		return
	}

	tokenAccountInfo, err := Parse.ParseTokenAccountInfo(accountInfo)
	if err != nil {
		err = gerror.Wrapf(err, "解析代币账户信息失败")
		return
	}

	owner = tokenAccountInfo.OwnerAddress

	return
}

func (pool *RPCs) getTokenAccountsOwner(ctx g.Ctx, addresses []Address.TokenAccountAddress) (owners []Address.AccountAddress, err error) {
	owners = make([]Address.AccountAddress, len(addresses))
	for i := 0; i < len(addresses); i += 100 {
		var accountsInfo [][]byte
		accountsInfo, err = pool.getAccountsInfo(ctx, Utils.Map(addresses[i:min(i+100, len(addresses))], func(address Address.TokenAccountAddress) (accountAddress Address.AccountAddress) {
			accountAddress = address.AccountAddress
			return
		}))
		if err != nil {
			err = gerror.Wrapf(err, "批量获取代币账户信息失败")
			return
		}

		for j := range accountsInfo {
			if accountsInfo[j] == nil {
				err = gerror.Newf("获取代币账户信息为空")
				return
			}

			var tokenAccountInfo Parse.TTokenAccountInfo
			tokenAccountInfo, err = Parse.ParseTokenAccountInfo(accountsInfo[j])
			if err != nil {
				err = gerror.Wrapf(err, "解析代币账户信息失败")
				return
			}

			owners[i+j] = tokenAccountInfo.OwnerAddress
		}
	}

	return
}
