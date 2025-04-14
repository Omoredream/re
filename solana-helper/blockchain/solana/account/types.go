package Account

import (
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type Account struct {
	Address    Address.AccountAddress
	SOL        decimal.Decimal
	Tokens     map[Address.TokenAddress]TokenAccount
	NFTs       map[Address.TokenAddress]TokenAccount
	Tokens2022 map[Address.TokenAddress]TokenAccount
	NFTs2022   map[Address.TokenAddress]TokenAccount
}

type TokenAccount struct {
	Address      Address.TokenAccountAddress
	TokenAddress Address.TokenAddress
	OwnerAddress Address.AccountAddress
	Token        decimal.Decimal
}
