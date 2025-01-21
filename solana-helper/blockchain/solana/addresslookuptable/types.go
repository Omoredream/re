package addresslookuptable

import (
	"github.com/gagliardetto/solana-go"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type AddressLookupTable struct {
	Address            Address.AccountAddress
	AddressLookupTable solana.PublicKeySlice
}

func AddressLookupTablesToAddressLookupTableMap(alts []AddressLookupTable) (altMap map[solana.PublicKey]solana.PublicKeySlice) {
	altMap = make(map[solana.PublicKey]solana.PublicKeySlice)
	for _, alt := range alts {
		altMap[alt.Address.PublicKey] = alt.AddressLookupTable
	}
	return
}
