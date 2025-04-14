package LP

import (
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type RaydiumCpmmPool struct {
	Discriminator      [8]byte `json:"-"`
	AmmConfig          Address.AccountAddress
	PoolCreator        Address.AccountAddress
	Token0Vault        Address.TokenAccountAddress
	Token1Vault        Address.TokenAccountAddress
	LPMint             Address.TokenAddress
	Token0Mint         Address.TokenAddress
	Token1Mint         Address.TokenAddress
	Token0Program      Address.ProgramAddress
	Token1Program      Address.ProgramAddress
	ObservationKey     Address.AccountAddress
	AutoBump           uint8
	Status             uint8
	LPMintDecimals     uint8
	Mint0Decimals      uint8
	Mint1Decimals      uint8
	LPSupply           uint64
	ProtocolFeesToken0 uint64
	ProtocolFeesToken1 uint64
	FundFeesToken0     uint64
	FundFeesToken1     uint64
	OpenTime           uint64
	RecentEpoch        uint64
	Padding            [31]uint64
}
