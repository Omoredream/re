package jupiterTypes

import (
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type SwapOption func() (string, any)

// 不自动 wrap/unwrap SOL
var SwapOptionUseWSOL SwapOption = func() (string, any) {
	return "wrapAndUnwrapSol", false
}

// 不使用公共的中间代币账户
var SwapOptionUseSelfTokenAccounts SwapOption = func() (string, any) {
	return "useSharedAccounts", false
}

// 抽成账户, 似乎是与 quote 时的 PlatformFeeBps 参数配合使用
func SwapOptionFeeAccount(feeAccount Address.AccountAddress) SwapOption {
	return func() (string, any) {
		return "feeAccount", feeAccount.String()
	}
}

// 标记账户, 用于跟踪特定交易
func SwapOptionTrackingAccount(trackingAccount Address.AccountAddress) SwapOption {
	return func() (string, any) {
		return "trackingAccount", trackingAccount.String()
	}
}

// 优先 CU 单价, 与总价二选一 (CU 限制默认为 1400000)
func SwapOptionComputeUnitPriceMicroLamports(cuPrice uint64) SwapOption {
	return func() (string, any) {
		return "computeUnitPriceMicroLamports", cuPrice
	}
}

// 优先 CU 总价, 与单价二选一 (CU 限制默认为 1400000)
func SwapOptionPrioritizationFeeLamports(cPrice uint64) SwapOption {
	return func() (string, any) {
		return "prioritizationFeeLamports", cPrice
	}
}

// 使用传统交易
var SwapOptionAsLegacyTransaction SwapOption = func() (string, any) {
	return "asLegacyTransaction", true
}

// B 代币的接收账户, 默认为钱包的 ATA 账户
func SwapOptionDestinationTokenAccount(destinationTokenAccount Address.TokenAccountAddress) SwapOption {
	return func() (string, any) {
		return "destinationTokenAccount", destinationTokenAccount.String()
	}
}

// 动态 CU 限制, 会增加响应耗时
var SwapOptionDynamicComputeUnitLimit SwapOption = func() (string, any) {
	return "dynamicComputeUnitLimit", true
}

// 跳过账户检查, 如 wSOL 账户、 ATA 账户是否存在等
var SwapOptionSkipUserAccountsRpcCalls SwapOption = func() (string, any) {
	return "skipUserAccountsRpcCalls", true
}

// 动态滑点
var SwapOptionDynamicSlippage SwapOption = func() (string, any) {
	return "dynamicSlippage", true
}
