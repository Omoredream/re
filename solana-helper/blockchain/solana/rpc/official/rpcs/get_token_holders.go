package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go/rpc"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/account"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func (pool *RPCs) GetTokenHolders(ctx g.Ctx, token Token.Token) (tokenHolders []Account.Account, err error) {
	getTokenLargestAccountsResult, err := pool.httpGetTokenLargestAccounts(ctx, token.Address)
	if err != nil {
		err = gerror.Wrapf(err, "获取代币 %s 最多持有者失败", token.Address)
		return
	}

	var owners []Address.AccountAddress
	owners, err = pool.getTokenAccountsOwner(ctx, lo.Map(getTokenLargestAccountsResult.Value, func(value *rpc.TokenLargestAccountsResult, _ int) Address.TokenAccountAddress {
		return Address.NewFromBytes32(value.Address).AsTokenAccountAddress()
	}))
	if err != nil {
		err = gerror.Wrapf(err, "批量获取代币账户所有者失败")
		return
	}

	tokenHolders = make([]Account.Account, len(getTokenLargestAccountsResult.Value))
	for i := range getTokenLargestAccountsResult.Value {
		tokenHolders[i] = Account.Account{
			Address: owners[i],
			SOL:     decimal.Zero, // todo 获取余额
			Tokens: map[Address.TokenAddress]Account.TokenAccount{
				token.Address: {
					Address:      Address.NewFromBytes32(getTokenLargestAccountsResult.Value[i].Address).AsTokenAccountAddress(),
					TokenAddress: token.Address,
					OwnerAddress: owners[i],
					Token:        lamports.LamportsString2Token(getTokenLargestAccountsResult.Value[i].Amount, token.Info.Decimalx),
				},
			},
		}
	}

	return
}
