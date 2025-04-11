package rpcs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.wkr.moe/web3/solana-helper/consts"
	Utils "git.wkr.moe/web3/solana-helper/utils"

	Account "git.wkr.moe/web3/solana-helper/blockchain/solana/account"
	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	Parse "git.wkr.moe/web3/solana-helper/blockchain/solana/parse"
	Token "git.wkr.moe/web3/solana-helper/blockchain/solana/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func (pool *RPCs) GetAccountToken(ctx g.Ctx, address Address.AccountAddress, tokenAddress Address.TokenAddress) (token *Account.TokenAccount, err error) {
	getTokenAccountsByOwnerResult, err := pool.httpGetTokenAccountsByOwner(ctx, address, &rpc.GetTokenAccountsConfig{
		Mint: tokenAddress.PublicKey.ToPointer(),
	})
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包代币失败")
		return
	}

	if len(getTokenAccountsByOwnerResult.Value) == 0 {
		return
	} else if len(getTokenAccountsByOwnerResult.Value) != 1 {
		err = gerror.Newf("返回数据数量不符合预期")
		return
	}

	tokenAccountInfo, err := Parse.ParseTokenAccountInfo(getTokenAccountsByOwnerResult.Value[0].Account.Data.GetBinary())
	if err != nil {
		err = gerror.Wrapf(err, "解析钱包代币失败")
		return
	}

	var token_ Token.Token
	token_, err = pool.TokenCacheGet(ctx, tokenAddress)
	if err != nil {
		err = gerror.Wrapf(err, "获取代币失败")
		return
	}

	if token_.IsToken() {
		tokenBalance := lamports.Lamports2Token(tokenAccountInfo.Token, token_.Info.Decimalx)
		if tokenBalance.Sign() >= 0 {
			token = &Account.TokenAccount{
				Address:      Address.NewFromBytes32(getTokenAccountsByOwnerResult.Value[0].Pubkey).AsTokenAccountAddress(),
				TokenAddress: tokenAddress,
				OwnerAddress: address,
				Token:        tokenBalance,
			}
			return
		}
	}

	err = gerror.Newf("目标代币不符合要求")

	return
}

func (pool *RPCs) getAccountTokens(ctx g.Ctx, address Address.AccountAddress, isToken2022 bool) (tokens map[Address.TokenAddress]Account.TokenAccount, nfts map[Address.TokenAddress]Account.TokenAccount, err error) {
	getTokenAccountsByOwnerResult, err := pool.httpGetTokenAccountsByOwner(ctx, address, &rpc.GetTokenAccountsConfig{
		ProgramId: func() *solana.PublicKey {
			if isToken2022 {
				return consts.Token2022ProgramAddress.PublicKey.ToPointer()
			} else {
				return consts.TokenProgramAddress.PublicKey.ToPointer()
			}
		}(),
	})
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包代币失败")
		return
	}

	tokenAccountsInfo := make([]Parse.TTokenAccountInfo, len(getTokenAccountsByOwnerResult.Value))
	for i := range getTokenAccountsByOwnerResult.Value {
		tokenAccountsInfo[i], err = Parse.ParseTokenAccountInfo(getTokenAccountsByOwnerResult.Value[i].Account.Data.GetBinary())
		if err != nil {
			err = gerror.Wrapf(err, "解析钱包代币失败")
			return
		}
	}

	tokens_, err := pool.TokenCacheGets(ctx, Utils.Map(tokenAccountsInfo, func(tokenAccountInfo Parse.TTokenAccountInfo) Address.TokenAddress {
		return tokenAccountInfo.TokenAddress
	}))
	if err != nil {
		err = gerror.Wrapf(err, "批量获取代币失败")
		return
	}

	tokens = make(map[Address.TokenAddress]Account.TokenAccount, len(tokenAccountsInfo))
	nfts = make(map[Address.TokenAddress]Account.TokenAccount, len(tokenAccountsInfo))
	for i := range tokenAccountsInfo {
		token := tokens_[i]
		if token.IsToken() {
			tokenBalance := lamports.Lamports2Token(tokenAccountsInfo[i].Token, token.Info.Decimalx)
			tokens[token.Address] = Account.TokenAccount{
				Address:      Address.NewFromBytes32(getTokenAccountsByOwnerResult.Value[i].Pubkey).AsTokenAccountAddress(),
				TokenAddress: tokenAccountsInfo[i].TokenAddress,
				OwnerAddress: address,
				Token:        tokenBalance,
			}
		} else {
			tokenBalance := lamports.Lamports2Token(tokenAccountsInfo[i].Token, token.Info.Decimalx)
			nfts[token.Address] = Account.TokenAccount{
				Address:      Address.NewFromBytes32(getTokenAccountsByOwnerResult.Value[i].Pubkey).AsTokenAccountAddress(),
				TokenAddress: tokenAccountsInfo[i].TokenAddress,
				OwnerAddress: address,
				Token:        tokenBalance,
			}
		}
	}

	return
}
