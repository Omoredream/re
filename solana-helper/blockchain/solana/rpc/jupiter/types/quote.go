package jupiterTypes

import (
	"github.com/gogf/gf/v2/text/gstr"
)

type QuoteOption func() (string, any)

// 滑点, 万分位, 如 50 = 0.5%
func QuoteOptionSlippageBps(slippageBps uint16) QuoteOption {
	return func() (string, any) {
		return "slippageBps", slippageBps
	}
}

// 交易模式, 精确输入
var QuoteOptionSwapExactIn QuoteOption = func() (string, any) {
	return "swapMode", "ExactIn"
}

// 交易模式, 精确输出
var QuoteOptionSwapExactOut QuoteOption = func() (string, any) {
	return "swapMode", "ExactOut"
}

// 白名单 Dex
func QuoteOptionDexes(dexes []string) QuoteOption {
	return func() (string, any) {
		return "dexes", gstr.Join(dexes, ",")
	}
}

// 黑名单 Dex
func QuoteOptionExcludeDexes(dexes []string) QuoteOption {
	return func() (string, any) {
		return "excludeDexes", gstr.Join(dexes, ",")
	}
}

// 仅使用流动性稳定的代币作为中转代币
var QuoteOptionRestrictIntermediateTokens QuoteOption = func() (string, any) {
	return "restrictIntermediateTokens", true
}

// 不使用中转代币
var QuoteOptionOnlyDirectRoutes QuoteOption = func() (string, any) {
	return "onlyDirectRoutes", true
}

// 使用传统交易
var QuoteOptionAsLegacyTransaction QuoteOption = func() (string, any) {
	return "asLegacyTransaction", true
}

// 平台抽成, 用途不明
func QuoteOptionPlatformFeeBps(platformFeeBps uint8) QuoteOption {
	return func() (string, any) {
		return "platformFeeBps", platformFeeBps
	}
}

// 交易中涉及账户的最大数量, 避免交易过大
func QuoteOptionMaxAccounts(maxAccounts uint) QuoteOption {
	return func() (string, any) {
		return "maxAccounts", maxAccounts
	}
}
