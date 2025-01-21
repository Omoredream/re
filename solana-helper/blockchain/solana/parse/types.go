package parse

import (
	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type TTokenAccountInfo struct {
	TokenAddress Address.TokenAddress
	OwnerAddress Address.AccountAddress
	Token        uint64
}
