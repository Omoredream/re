package rpcs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	Utils "git.wkr.moe/web3/solana-helper/utils"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	Parse "git.wkr.moe/web3/solana-helper/blockchain/solana/parse"
	Token "git.wkr.moe/web3/solana-helper/blockchain/solana/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func (pool *RPCs) GetTokenAccountBalance(ctx g.Ctx, address Address.TokenAccountAddress, token Token.Token) (balance decimal.Decimal, err error) {
	getTokenAccountBalanceResult, err := pool.httpGetTokenAccountBalance(ctx, address)
	if err != nil {
		err = gerror.Wrapf(err, "获取代币账户余额失败")
		return
	}

	balance = lamports.LamportsString2Token(getTokenAccountBalanceResult.Value.Amount, token.Info.Decimalx)

	return
}

func (pool *RPCs) GetTokenAccountsBalance(ctx g.Ctx, addresses []Address.TokenAccountAddress, tokens []Token.Token) (balances []decimal.Decimal, err error) {
	balances = make([]decimal.Decimal, len(addresses))
	for i := 0; i < len(addresses); i += 100 {
		var accountsInfo [][]byte
		accountsInfo, err = pool.getAccountsInfo(ctx, Utils.Map(addresses[i:min(i+100, len(addresses))], func(address Address.TokenAccountAddress) (accountAddress Address.AccountAddress) {
			accountAddress = address.AccountAddress
			return
		}))
		if err != nil {
			err = gerror.Wrapf(err, "批量获取代币账户失败")
			return
		}

		for j := range accountsInfo {
			if accountsInfo[j] == nil {
				err = gerror.Newf("获取代币账户为空")
				return
			}

			var tokenAccountInfo Parse.TTokenAccountInfo
			tokenAccountInfo, err = Parse.ParseTokenAccountInfo(accountsInfo[j])
			if err != nil {
				err = gerror.Wrapf(err, "解析代币账户失败")
				return
			}

			balances[i+j] = lamports.Lamports2Token(tokenAccountInfo.Token, tokens[i+j].Info.Decimalx)
		}
	}

	return
}
