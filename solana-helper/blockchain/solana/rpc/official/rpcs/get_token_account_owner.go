package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/samber/lo"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/parser"
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

	tokenAccountInfo, err := Parser.ParseTokenAccountInfo(accountInfo)
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
		accountsInfo, err = pool.getAccountsInfo(ctx, lo.Map(addresses[i:min(i+100, len(addresses))], func(address Address.TokenAccountAddress, _ int) (accountAddress Address.AccountAddress) {
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

			var tokenAccountInfo Parser.TTokenAccountInfo
			tokenAccountInfo, err = Parser.ParseTokenAccountInfo(accountsInfo[j])
			if err != nil {
				err = gerror.Wrapf(err, "解析代币账户信息失败")
				return
			}

			owners[i+j] = tokenAccountInfo.OwnerAddress
		}
	}

	return
}
