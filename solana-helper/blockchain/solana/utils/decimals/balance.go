package decimals

import (
	"strings"

	"github.com/gogf/gf/v2/text/gstr"
	"github.com/shopspring/decimal"
)

func DisplayBalance(balance decimal.Decimal) (result string) {
	signPart := balance.Sign()
	numberPart := balance.Abs()

	intPart := numberPart.Truncate(0)
	intString := intPart.String()
	intLength := len(intString)

	fractionalPart := numberPart.Sub(intPart)
	fractionalString := fractionalPart.Coefficient().String()
	fractionalLength := len(fractionalString)
	fractionalLeftZeroLength := -int(fractionalPart.Exponent()) - fractionalLength

	if signPart == -1 {
		result += "-"
	}

	result += intString[0 : (intLength-1)%3+1]
	for i := (intLength-1)%3 + 1; i < intLength; i += 3 {
		result += ","
		result += intString[i : i+3]
	}

	if fractionalPart.IsZero() {
		return
	} else {
		result += "."
	}

	resultFractionalLeftZeroLength := 0
	if fractionalLeftZeroLength > 2 {
		result += "0"
		resultFractionalLeftZeroLength++
		resultFractionalLeftZero := strings.Builder{}
		for fractionalLeftZeroLength != 0 {
			resultFractionalLeftZero.WriteRune([]rune{'₀', '₁', '₂', '₃', '₄', '₅', '₆', '₇', '₈', '₉'}[fractionalLeftZeroLength%10])
			fractionalLeftZeroLength /= 10
			resultFractionalLeftZeroLength++
		}
		result += gstr.Reverse(resultFractionalLeftZero.String())
	} else {
		result += gstr.Repeat("0", fractionalLeftZeroLength)
		resultFractionalLeftZeroLength += fractionalLeftZeroLength
	}

	if resultFractionalLeftZeroLength+fractionalLength <= 8 {
		result += fractionalString
	} else {
		resultFractionalLength := 8 - resultFractionalLeftZeroLength
		result += gstr.TrimRight(fractionalString[:resultFractionalLength], "0")
	}

	return
}
