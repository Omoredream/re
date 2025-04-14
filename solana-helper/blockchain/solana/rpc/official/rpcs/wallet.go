package officialRPCs

import (
	"crypto/ed25519"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"github.com/blocto/solana-go-sdk/pkg/hdwallet"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/gagliardetto/solana-go"
	"github.com/tyler-smith/go-bip39"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/account"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"
)

var baseDerivationPath accounts.DerivationPath

func init() {
	baseDerivationPath = accounts.DerivationPath{0x80000000 + 44, 0x80000000 + 501} // m/44'/501'
}

func (pool *RPCs) NewWalletFromWIF(ctx g.Ctx, WIF string, skipScan ...bool) (wallet Wallet.HostedWallet, err error) {
	privateKey, err := solana.PrivateKeyFromBase58(WIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包私钥失败")
		return
	}

	return pool.NewWalletFromPrivateKey(ctx, privateKey, skipScan...)
}

func (pool *RPCs) NewWalletFromMnemonic(ctx g.Ctx, mnemonic string, derivation uint32, skipScan ...bool) (wallet Wallet.HostedWallet, err error) {
	extendedKey := hdwallet.CreateMasterKey(bip39.NewSeed(mnemonic, ""))
	for _, n := range append(baseDerivationPath, 0x80000000+derivation, 0x80000000+0) { // /d'/0'
		extendedKey = hdwallet.CKDPriv(extendedKey, n)
	}

	privateKey := ed25519.NewKeyFromSeed(extendedKey.PrivateKey)

	return pool.NewWalletFromPrivateKey(ctx, privateKey, skipScan...)
}

func (pool *RPCs) NewWalletsFromMnemonic(ctx g.Ctx, mnemonic string, derivationFrom, derivationTo uint32, skipScan ...bool) (wallets []Wallet.HostedWallet, err error) {
	wallets = make([]Wallet.HostedWallet, derivationTo-derivationFrom)

	extendedKey := hdwallet.CreateMasterKey(bip39.NewSeed(mnemonic, ""))
	for _, n := range baseDerivationPath {
		extendedKey = hdwallet.CKDPriv(extendedKey, n)
	}
	for derivation := derivationFrom; derivation < derivationTo; derivation++ {
		extendedKey := extendedKey
		for _, n := range (accounts.DerivationPath{0x80000000 + derivation, 0x80000000 + 0}) { // /d'/0'
			extendedKey = hdwallet.CKDPriv(extendedKey, n)
		}

		privateKey := ed25519.NewKeyFromSeed(extendedKey.PrivateKey)

		wallets[derivation-derivationFrom], err = pool.NewWalletFromPrivateKey(ctx, privateKey, skipScan...)
	}

	return
}

func (pool *RPCs) NewWalletFromPrivateKey(ctx g.Ctx, privateKey []byte, skipScan ...bool) (wallet Wallet.HostedWallet, err error) {
	address := Address.NewFromBytes32(solana.PrivateKey(privateKey).PublicKey())

	wallet = Wallet.HostedWallet{
		WatchWallet: Wallet.WatchWallet{
			Account: Account.Account{
				Address: address,
			},
		},
		PrivateKey: privateKey,
	}

	if len(skipScan) == 0 || skipScan[0] == false {
		err = pool.UpdateWalletAssets(ctx, &(wallet.WatchWallet))
		if err != nil {
			err = gerror.Wrapf(err, "更新钱包资产失败")
			return
		}
	}

	return
}

func (pool *RPCs) NewWatchWallet(ctx g.Ctx, address string) (wallet Wallet.WatchWallet, err error) {
	wallet = Wallet.WatchWallet{
		Account: Account.Account{
			Address: Address.NewFromBase58(address),
		},
	}

	err = pool.UpdateWalletAssets(ctx, &wallet)
	if err != nil {
		err = gerror.Wrapf(err, "更新钱包资产失败")
		return
	}

	return
}

func (pool *RPCs) UpdateWalletAssets(ctx g.Ctx, wallet *Wallet.WatchWallet) (err error) {
	wallet.Account, err = pool.GetAccountOverview(ctx, wallet.Account.Address)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包资产失败")
		return
	}

	return
}

func (pool *RPCs) CreateWallet() (wallet Wallet.HostedWallet, err error) {
	privateKey, err := solana.NewRandomPrivateKey()
	if err != nil {
		err = gerror.Wrapf(err, "生成钱包私钥失败")
		return
	}

	address := Address.NewFromBytes32(privateKey.PublicKey())

	wallet = Wallet.HostedWallet{
		WatchWallet: Wallet.WatchWallet{
			Account: Account.Account{
				Address: address,
				SOL:     decimal.Zero,
				Tokens:  make(map[Address.TokenAddress]Account.TokenAccount),
			},
		},
		PrivateKey: privateKey,
	}

	return
}
