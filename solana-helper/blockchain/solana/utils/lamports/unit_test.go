package lamports_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func TestDecimal(t *testing.T) {
	assert.Equal(t, decimal.NewFromFloat(0.00001).String(), "0.00001")
	assert.Equal(t, decimal.NewFromFloat(0.1).String(), "0.1")
}

func TestLamports2SOL(t *testing.T) {
	assert.Equal(t, decimals.DisplayBalance(lamports.Lamports2SOL(12345)), "0.0₄12345")
}

func TestSOL2Lamports(t *testing.T) {
	var f decimal.Decimal
	var ok error

	f, ok = decimal.NewFromString("0.000001234")
	assert.NoError(t, ok)
	assert.Equal(t, lamports.SOL2Lamports(f), uint64(1234))

	f, ok = decimal.NewFromString("0.0000012344")
	assert.NoError(t, ok)
	assert.Equal(t, lamports.SOL2Lamports(f), uint64(1234))

	f, ok = decimal.NewFromString("0.0000012345")
	assert.NoError(t, ok)
	assert.Equal(t, lamports.SOL2Lamports(f), uint64(1235))
}
