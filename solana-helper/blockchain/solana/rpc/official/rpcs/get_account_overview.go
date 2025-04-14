package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/account"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

func (pool *RPCs) GetAccountOverview(ctx g.Ctx, address Address.AccountAddress) (account Account.Account, err error) {
	balance, err := pool.getAccountBalance(ctx, address)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包 %s 余额失败", address)
		return
	}

	tokens, nfts, err := pool.getAccountTokens(ctx, address, false)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包 %s 代币余额失败", address)
		return
	}

	tokens2022, nfts2022, err := pool.getAccountTokens(ctx, address, true)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包 %s 2022 标准代币余额失败", address)
		return
	}

	account = Account.Account{
		Address:    address,
		SOL:        balance,
		Tokens:     tokens,
		NFTs:       nfts,
		Tokens2022: tokens2022,
		NFTs2022:   nfts2022,
	}

	return
}
