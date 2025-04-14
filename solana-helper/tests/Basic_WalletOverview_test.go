package tests

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/worth"
)

func TestAccountOverview(t *testing.T) {
	err := testAccountOverview(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testAccountOverview(ctx g.Ctx) (err error) {
	address := Address.NewFromBase58("888jyPHuucwtYFepqQY7F55xXeBP4P3pp8QC8Awai888")

	account, err := officialPool.GetAccountOverview(ctx, address)
	if err != nil {
		err = gerror.Wrapf(err, "查询钱包失败")
		return
	}

	g.Log().Infof(ctx, "钱包: %s", account.Address)
	g.Log().Infof(ctx, "|- SOL: %s ($%s)", decimals.DisplayBalance(account.SOL), decimals.DisplayBalance(worth.IgnoreErr(worth.SOLWorth(ctx, account.SOL))))
	for tokenAddress, tokenBalance := range account.Tokens {
		var token Token.Token
		token, err = officialPool.TokenCacheGet(ctx, tokenAddress)
		if err != nil {
			err = gerror.Wrapf(err, "获取钱包 %s 代币 %s 失败", address, tokenAddress)
			return
		}

		g.Log().Infof(ctx, "|- 代币 %s: %s", token.DisplayName(), decimals.DisplayBalance(tokenBalance.Token))
	}
	for nftAddress, nftBalance := range account.NFTs {
		var token Token.Token
		token, err = officialPool.TokenCacheGet(ctx, nftAddress)
		if err != nil {
			err = gerror.Wrapf(err, "获取钱包 %s NFT %s 失败", address, nftBalance)
			return
		}

		g.Log().Infof(ctx, "|- NFT %s: %s", token.DisplayName(), decimals.DisplayBalance(nftBalance.Token))
	}

	return
}

func TestWalletOverview(t *testing.T) {
	err := testWalletOverview(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testWalletOverview(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	g.Log().Infof(ctx, "钱包: %s", wallet.Account.Address)
	g.Log().Infof(ctx, "|- SOL: %s ($%s)", decimals.DisplayBalance(wallet.Account.SOL), decimals.DisplayBalance(worth.IgnoreErr(worth.SOLWorth(ctx, wallet.Account.SOL))))
	for tokenAddress, tokenBalance := range wallet.Account.Tokens {
		var token Token.Token
		token, err = officialPool.TokenCacheGet(ctx, tokenAddress)
		if err != nil {
			err = gerror.Wrapf(err, "获取钱包 %s 代币 %s 失败", wallet.Account, tokenAddress)
			return
		}

		g.Log().Infof(ctx, "|- 代币 %s: %s", token.DisplayName(), decimals.DisplayBalance(tokenBalance.Token))
	}
	for nftAddress, nftBalance := range wallet.Account.NFTs {
		var token Token.Token
		token, err = officialPool.TokenCacheGet(ctx, nftAddress)
		if err != nil {
			err = gerror.Wrapf(err, "获取钱包 %s NFT %s 失败", wallet.Account, nftBalance)
			return
		}

		g.Log().Infof(ctx, "|- NFT %s: %s", token.DisplayName(), decimals.DisplayBalance(nftBalance.Token))
	}

	return
}

func TestWalletsOverview(t *testing.T) {
	err := testWalletsOverview(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testWalletsOverview(ctx g.Ctx) (err error) {
	for derivation := uint32(0); derivation < 150; derivation++ {
		ctx := context.WithValue(ctx, consts.CtxDerivation, derivation)

		var wallet Wallet.HostedWallet
		wallet, err = officialPool.NewWalletFromMnemonic(ctx, testWalletMnemonic, derivation)
		if err != nil {
			err = gerror.Wrapf(err, "导入钱包失败")
			return
		}

		g.Log().Infof(ctx, "钱包: %s", wallet.Account.Address)
		if wallet.Account.SOL.IsPositive() {
			g.Log().Infof(ctx, "|- SOL: %s ($%s)", decimals.DisplayBalance(wallet.Account.SOL), decimals.DisplayBalance(worth.IgnoreErr(worth.SOLWorth(ctx, wallet.Account.SOL))))
		}
		for tokenAddress, tokenBalance := range wallet.Account.Tokens {
			if !tokenBalance.Token.IsPositive() {
				continue
			}

			var token Token.Token
			token, err = officialPool.TokenCacheGet(ctx, tokenAddress)
			if err != nil {
				err = gerror.Wrapf(err, "获取钱包 %s 代币 %s 失败", wallet.Account, tokenAddress)
				return
			}

			g.Log().Infof(ctx, "|- 代币 %s: %s", token.DisplayName(), decimals.DisplayBalance(tokenBalance.Token))
		}
		for nftAddress, nftBalance := range wallet.Account.NFTs {
			if !nftBalance.Token.IsPositive() {
				continue
			}

			var token Token.Token
			token, err = officialPool.TokenCacheGet(ctx, nftAddress)
			if err != nil {
				err = gerror.Wrapf(err, "获取钱包 %s NFT %s 失败", wallet.Account, nftBalance)
				return
			}

			g.Log().Infof(ctx, "|- NFT %s: %s", token.DisplayName(), decimals.DisplayBalance(nftBalance.Token))
		}
		for tokenAddress, tokenBalance := range wallet.Account.Tokens2022 {
			if !tokenBalance.Token.IsPositive() {
				continue
			}

			var token Token.Token
			token, err = officialPool.TokenCacheGet(ctx, tokenAddress)
			if err != nil {
				err = gerror.Wrapf(err, "获取钱包 %s 代币 %s 失败", wallet.Account, tokenAddress)
				return
			}

			g.Log().Infof(ctx, "|- 代币 %s: %s", token.DisplayName(), decimals.DisplayBalance(tokenBalance.Token))
		}
		for nftAddress, nftBalance := range wallet.Account.NFTs2022 {
			if !nftBalance.Token.IsPositive() {
				continue
			}

			var token Token.Token
			token, err = officialPool.TokenCacheGet(ctx, nftAddress)
			if err != nil {
				err = gerror.Wrapf(err, "获取钱包 %s NFT %s 失败", wallet.Account, nftBalance)
				return
			}

			g.Log().Infof(ctx, "|- NFT %s: %s", token.DisplayName(), decimals.DisplayBalance(nftBalance.Token))
		}
	}

	return
}
