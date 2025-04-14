package AddressLookupTable

import (
	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type AddressLookupTable struct {
	Address            Address.AccountAddress
	AddressLookupTable solana.PublicKeySlice
}

type AddressLookupTables []AddressLookupTable

func (alts AddressLookupTables) ToAddressLookupTableMap() (altMap map[solana.PublicKey]solana.PublicKeySlice) {
	altMap = make(map[solana.PublicKey]solana.PublicKeySlice)
	for _, alt := range alts {
		altMap[alt.Address.PublicKey] = alt.AddressLookupTable
	}
	return
}
