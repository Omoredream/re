package rpcs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	Utils "git.wkr.moe/web3/solana-helper/utils"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	Parse "git.wkr.moe/web3/solana-helper/blockchain/solana/parse"
	Token "git.wkr.moe/web3/solana-helper/blockchain/solana/token"
)

func (pool *RPCs) getTokenInfo(ctx g.Ctx, address Address.TokenAddress) (token Token.Info, err error) {
	accountInfo, err := pool.GetAccountInfo(ctx, address.AccountAddress)
	if err != nil {
		err = gerror.Wrapf(err, "获取代币信息失败")
		return
	}

	if accountInfo == nil {
		err = gerror.Newf("获取代币信息为空")
		return
	}

	token, err = Parse.ParseTokenInfo(accountInfo)
	if err != nil {
		err = gerror.Wrapf(err, "解析代币信息失败")
		return
	}

	return
}

func (pool *RPCs) getTokensInfo(ctx g.Ctx, addresses []Address.TokenAddress) (tokens []Token.Info, err error) {
	tokens = make([]Token.Info, len(addresses))
	for i := 0; i < len(addresses); i += 100 {
		var accountsInfo [][]byte
		accountsInfo, err = pool.getAccountsInfo(ctx, Utils.Map(addresses[i:min(i+100, len(addresses))], func(address Address.TokenAddress) (accountAddress Address.AccountAddress) {
			accountAddress = address.AccountAddress
			return
		}))
		if err != nil {
			err = gerror.Wrapf(err, "批量获取代币信息失败")
			return
		}

		for j := range accountsInfo {
			if accountsInfo[j] == nil {
				err = gerror.Newf("获取代币信息为空")
				return
			}

			tokens[i+j], err = Parse.ParseTokenInfo(accountsInfo[j])
			if err != nil {
				err = gerror.Wrapf(err, "解析代币信息失败")
				return
			}
		}
	}

	return
}
