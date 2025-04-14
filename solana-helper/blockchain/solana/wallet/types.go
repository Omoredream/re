package Wallet

import (
	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/account"
)

type WatchWallet struct {
	Account Account.Account
}

type HostedWallet struct {
	WatchWallet
	PrivateKey solana.PrivateKey
}
