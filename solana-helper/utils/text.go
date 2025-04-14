package Utils

import (
	"github.com/fatih/color"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
)

func UDColorThird(before decimal.Decimal, after decimal.Decimal, show decimal.Decimal) string {
	return UDColor(show, after.Cmp(before))
}

func UDColorDiff(before decimal.Decimal, after decimal.Decimal) string {
	return UDColor(after, after.Cmp(before))
}

func UDColorSelf(n decimal.Decimal) string {
	return UDColor(n, n.Sign())
}

func UDColor(n decimal.Decimal, cmp int) string {
	var c int
	switch cmp {
	case 1:
		c = glog.COLOR_RED
	case -1:
		c = glog.COLOR_GREEN
	case 0:
		c = glog.COLOR_BLACK
	}
	return color.New(color.Attribute(c)).Sprint(decimals.DisplayBalance(n))
}
