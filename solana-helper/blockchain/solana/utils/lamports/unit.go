package lamports

import (
	"math/big"

	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/consts"
)

func Lamports2SOL(lamports uint64) (SOL decimal.Decimal) {
	SOL = Lamports2Token(lamports, consts.SOL.Info.Decimalx)

	return
}

func SOL2Lamports(SOL decimal.Decimal) (lamports uint64) {
	lamports = Token2Lamports(SOL, consts.SOL.Info.Decimalx)

	return
}

func Lamports2Token(lamports uint64, decimals int32) (token decimal.Decimal) {
	lamports_ := new(big.Int).SetUint64(lamports)
	token = lamportsBigInt2Token(lamports_, decimals)

	return
}

func LamportsString2Token(lamports string, decimals int32) (token decimal.Decimal) {
	lamports_, _ := new(big.Int).SetString(lamports, 10)
	token = lamportsBigInt2Token(lamports_, decimals)

	return
}

func lamportsBigInt2Token(lamports *big.Int, decimals int32) (token decimal.Decimal) {
	token = decimal.NewFromBigInt(lamports, -decimals)

	return
}

func Token2Lamports(token decimal.Decimal, decimals int32) (lamports uint64) {
	lamports = token.Shift(decimals).Round(0).BigInt().Uint64()

	return
}
