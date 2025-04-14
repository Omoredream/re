package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/parser"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
)

func (pool *RPCs) getToken(ctx g.Ctx, address Address.TokenAddress) (token Token.Token, err error) {
	tokenMetadataAddress, err := address.FindTokenMetadataAddress()
	if err != nil {
		err = gerror.Wrapf(err, "无法找到代币元数据地址")
		return
	}

	var tokenInfo Token.Info
	tokenInfo, err = pool.getTokenInfo(ctx, address)
	if err != nil {
		err = gerror.Wrapf(err, "获取代币信息失败")
		return
	}

	var tokenMetadata *Token.Metadata
	tokenMetadata, err = pool.getTokenMetadata(ctx, tokenMetadataAddress)
	if err != nil {
		err = gerror.Wrapf(err, "获取代币元数据失败")
		return
	}

	token = Parser.ParseToken(address, tokenInfo, tokenMetadata)

	return
}

func (pool *RPCs) getTokens(ctx g.Ctx, addresses []Address.TokenAddress) (tokens []Token.Token, err error) {
	if len(addresses) == 0 {
		return
	}
	tokens = make([]Token.Token, len(addresses))
	for i := 0; i < len(addresses); i += 50 {
		var accountsAddress []Address.AccountAddress
		accountsAddress, err = Utils.MapsWithErr(addresses[i:min(i+50, len(addresses))], func(address Address.TokenAddress) (accountsAddress []Address.AccountAddress, err error) {
			accountsAddress = make([]Address.AccountAddress, 2)

			accountsAddress[0] = address.AccountAddress

			var tokenMetaAddress Address.TokenMetadataAddress
			tokenMetaAddress, err = address.FindTokenMetadataAddress()
			if err != nil {
				err = gerror.Wrapf(err, "无法找到代币元数据地址")
				return
			}
			accountsAddress[1] = tokenMetaAddress.AccountAddress

			return
		})
		if err != nil {
			err = gerror.Wrapf(err, "批量寻找代币元数据地址失败")
			return
		}

		var accountsInfo [][]byte
		accountsInfo, err = pool.getAccountsInfo(ctx, accountsAddress)
		if err != nil {
			err = gerror.Wrapf(err, "批量获取代币失败")
			return
		}

		for j := 0; j < len(accountsInfo); j += 2 {
			if accountsInfo[j] == nil {
				err = gerror.Newf("获取代币信息为空")
				return
			}

			var tokenInfo Token.Info
			tokenInfo, err = Parser.ParseTokenInfo(accountsInfo[j])
			if err != nil {
				err = gerror.Wrapf(err, "解析代币信息失败")
				return
			}

			var tokenMetadata *Token.Metadata

			if accountsInfo[j+1] != nil {
				tokenMetadata = &Token.Metadata{}
				*tokenMetadata, err = Parser.ParseTokenMetadata(accountsInfo[j+1])
				if err != nil {
					err = gerror.Wrapf(err, "解析代币元数据失败")
					return
				}
			}

			tokens[i+j/2] = Parser.ParseToken(addresses[i+j/2], tokenInfo, tokenMetadata)
		}
	}

	return
}
