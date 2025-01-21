package rpcs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	Utils "git.wkr.moe/web3/solana-helper/utils"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	Parse "git.wkr.moe/web3/solana-helper/blockchain/solana/parse"
	Token "git.wkr.moe/web3/solana-helper/blockchain/solana/token"
)

func (pool *RPCs) getTokenMetadata(ctx g.Ctx, address Address.TokenMetadataAddress) (token *Token.Metadata, err error) {
	accountInfo, err := pool.GetAccountInfo(ctx, address.AccountAddress)
	if err != nil {
		err = gerror.Wrapf(err, "获取代币元数据失败")
		return
	}

	if accountInfo == nil {
		return
	}

	token = &Token.Metadata{}
	*token, err = Parse.ParseTokenMetadata(accountInfo)
	if err != nil {
		err = gerror.Wrapf(err, "解析代币元数据失败")
		return
	}

	return
}

func (pool *RPCs) getTokensMetadata(ctx g.Ctx, addresses []Address.TokenMetadataAddress) (tokens []*Token.Metadata, err error) {
	tokens = make([]*Token.Metadata, len(addresses))
	for i := 0; i < len(addresses); i += 100 {
		var accountsInfo [][]byte
		accountsInfo, err = pool.getAccountsInfo(ctx, Utils.Map(addresses[i:min(i+100, len(addresses))], func(address Address.TokenMetadataAddress) (accountAddress Address.AccountAddress) {
			accountAddress = address.AccountAddress
			return
		}))
		if err != nil {
			err = gerror.Wrapf(err, "批量获取代币元数据失败")
			return
		}

		for j := range accountsInfo {
			if accountsInfo[j] == nil {
				continue
			}

			tokens[i+j] = &Token.Metadata{}
			*tokens[i+j], err = Parse.ParseTokenMetadata(accountsInfo[j])
			if err != nil {
				err = gerror.Wrapf(err, "解析代币元数据失败")
				return
			}
		}
	}

	return
}
