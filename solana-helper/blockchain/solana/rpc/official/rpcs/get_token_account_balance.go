package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/parser"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"

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
		accountsInfo, err = pool.getAccountsInfo(ctx, lo.Map(addresses[i:min(i+100, len(addresses))], func(address Address.TokenAccountAddress, _ int) (accountAddress Address.AccountAddress) {
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

			var tokenAccountInfo Parser.TTokenAccountInfo
			tokenAccountInfo, err = Parser.ParseTokenAccountInfo(accountsInfo[j])
			if err != nil {
				err = gerror.Wrapf(err, "解析代币账户失败")
				return
			}

			balances[i+j] = lamports.Lamports2Token(tokenAccountInfo.Token, tokens[i+j].Info.Decimalx)
		}
	}

	return
}
