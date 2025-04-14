package LP

import (
	"github.com/gagliardetto/binary"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type RaydiumAmmV4Pool struct {
	Status             uint64
	Nonce              uint64
	OrderNum           uint64
	Depth              uint64
	CoinDecimals       uint64
	PcDecimals         uint64
	State              uint64
	ResetFlag          uint64
	MinSize            uint64
	VolMaxCutRatio     uint64
	AmountWave         uint64
	CoinLotSize        uint64
	PcLotSize          uint64
	MinPriceMultiplier uint64
	MaxPriceMultiplier uint64
	SysDecimalValue    uint64
	Fees               RaydiumAmmV4Fees
	StateData          RaydiumAmmV4StateData
	CoinVault          Address.TokenAccountAddress
	PcVault            Address.TokenAccountAddress
	CoinVaultMint      Address.TokenAddress
	PcVaultMint        Address.TokenAddress
	LPMint             Address.TokenAddress
	OpenOrders         Address.AccountAddress
	Market             Address.AccountAddress
	MarketProgram      Address.ProgramAddress
	TargetOrders       Address.AccountAddress
	Padding1           [8]uint64
	AmmOwner           Address.AccountAddress
	LPAmount           uint64
	ClientOrderId      uint64
	RecentEpoch        uint64
	Padding2           [1]uint64
}

type RaydiumAmmV4Fees struct {
	MinSeparateNumerator   uint64
	MinSeparateDenominator uint64
	TradeFeeNumerator      uint64
	TradeFeeDenominator    uint64
	PnlNumerator           uint64
	PnlDenominator         uint64
	SwapFeeNumerator       uint64
	SwapFeeDenominator     uint64
}

type RaydiumAmmV4StateData struct {
	NeedTakePnlCoin     uint64
	NeedTakePnlPc       uint64
	TotalPnlPc          uint64
	TotalPnlCoin        uint64
	PoolOpenTime        uint64
	Padding             [2]uint64
	OrderbookToInitTime uint64
	SwapCoinInAmount    bin.Uint128
	SwapPcOutAmount     bin.Uint128
	SwapAccPcFee        uint64
	SwapPcInAmount      bin.Uint128
	SwapCoinOutAmount   bin.Uint128
	SwapAccCoinFee      uint64
}
