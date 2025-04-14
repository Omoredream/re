package decimals_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
)

func TestDisplayBalance(t *testing.T) {
	var f decimal.Decimal
	var ok error

	f, ok = decimal.NewFromString("0.12345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "0.12345")

	f, ok = decimal.NewFromString("-0.12345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-0.12345")

	f, ok = decimal.NewFromString("-1.12345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1.12345")

	f, ok = decimal.NewFromString("-12.12345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12.12345")

	f, ok = decimal.NewFromString("-123.12345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123.12345")

	f, ok = decimal.NewFromString("-1234.12345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1,234.12345")

	f, ok = decimal.NewFromString("-12345.12345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12,345.12345")

	f, ok = decimal.NewFromString("-123456.12345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123,456.12345")

	f, ok = decimal.NewFromString("-0.012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-0.012345")

	f, ok = decimal.NewFromString("-1.012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1.012345")

	f, ok = decimal.NewFromString("-12.012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12.012345")

	f, ok = decimal.NewFromString("-123.012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123.012345")

	f, ok = decimal.NewFromString("-1234.012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1,234.012345")

	f, ok = decimal.NewFromString("-12345.012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12,345.012345")

	f, ok = decimal.NewFromString("-123456.012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123,456.012345")

	f, ok = decimal.NewFromString("-0.0012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-0.0012345")

	f, ok = decimal.NewFromString("-1.0012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1.0012345")

	f, ok = decimal.NewFromString("-12.0012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12.0012345")

	f, ok = decimal.NewFromString("-123.0012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123.0012345")

	f, ok = decimal.NewFromString("-1234.0012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1,234.0012345")

	f, ok = decimal.NewFromString("-12345.0012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12,345.0012345")

	f, ok = decimal.NewFromString("-123456.0012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123,456.0012345")

	f, ok = decimal.NewFromString("-0.00012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-0.0₃12345")

	f, ok = decimal.NewFromString("-1.00012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1.0₃12345")

	f, ok = decimal.NewFromString("-12.00012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12.0₃12345")

	f, ok = decimal.NewFromString("-123.00012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123.0₃12345")

	f, ok = decimal.NewFromString("-1234.00012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1,234.0₃12345")

	f, ok = decimal.NewFromString("-12345.00012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12,345.0₃12345")

	f, ok = decimal.NewFromString("-123456.00012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123,456.0₃12345")

	f, ok = decimal.NewFromString("-0.00000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-0.0₉12345")

	f, ok = decimal.NewFromString("-1.00000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1.0₉12345")

	f, ok = decimal.NewFromString("-12.00000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12.0₉12345")

	f, ok = decimal.NewFromString("-123.00000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123.0₉12345")

	f, ok = decimal.NewFromString("-1234.00000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1,234.0₉12345")

	f, ok = decimal.NewFromString("-12345.00000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12,345.0₉12345")

	f, ok = decimal.NewFromString("-123456.00000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123,456.0₉12345")

	f, ok = decimal.NewFromString("-0.000000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-0.0₁₀12345")

	f, ok = decimal.NewFromString("-1.000000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1.0₁₀12345")

	f, ok = decimal.NewFromString("-12.000000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12.0₁₀12345")

	f, ok = decimal.NewFromString("-123.000000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123.0₁₀12345")

	f, ok = decimal.NewFromString("-1234.000000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1,234.0₁₀12345")

	f, ok = decimal.NewFromString("-12345.000000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12,345.0₁₀12345")

	f, ok = decimal.NewFromString("-123456.000000000012345")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123,456.0₁₀12345")

	f, ok = decimal.NewFromString("-0.0000000000123456")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-0.0₁₀12345")

	f, ok = decimal.NewFromString("-1.0000000000123456")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1.0₁₀12345")

	f, ok = decimal.NewFromString("-12.0000000000123456")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12.0₁₀12345")

	f, ok = decimal.NewFromString("-123.0000000000123456")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123.0₁₀12345")

	f, ok = decimal.NewFromString("-1234.0000000000123456")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-1,234.0₁₀12345")

	f, ok = decimal.NewFromString("-12345.0000000000123456")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-12,345.0₁₀12345")

	f, ok = decimal.NewFromString("-123456.0000000000123456")
	assert.NoError(t, ok)
	assert.Equal(t, decimals.DisplayBalance(f), "-123,456.0₁₀12345")
}
