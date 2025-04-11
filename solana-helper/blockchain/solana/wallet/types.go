package wallet

import (
	"github.com/gagliardetto/solana-go"

	Account "git.wkr.moe/web3/solana-helper/blockchain/solana/account"
)

type WatchWallet struct {
	Account Account.Account
}

type HostedWallet struct {
	WatchWallet
	PrivateKey solana.PrivateKey
}
