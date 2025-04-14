package types

import (
	"github.com/shopspring/decimal"
)

type TipFromProfitBps struct {
	MinProfit decimal.Decimal
	BpsU16    uint16
	BpsDec    decimal.Decimal
}
