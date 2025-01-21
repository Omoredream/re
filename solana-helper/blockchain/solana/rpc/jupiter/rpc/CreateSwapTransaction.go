package rpc

import (
	"net/http"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/errcode"

	jupiterTypes "git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/types"
)

type CreateSwapTransactionOptionalArgs struct {
	WrapAndUnwrapSol              *bool   `json:"wrapAndUnwrapSol,omitempty"`              // 自动 wrap/unwrap SOL
	UseSharedAccounts             *bool   `json:"useSharedAccounts,omitempty"`             // 不创建独立的中间代币账户
	FeeAccount                    *string `json:"feeAccount,omitempty"`                    // 抽成账户, 似乎是与 quote 时的 PlatformFeeBps 参数配合使用
	TrackingAccount               *string `json:"trackingAccount,omitempty"`               // 标记账户, 用于跟踪特定交易
	ComputeUnitPriceMicroLamports *uint64 `json:"computeUnitPriceMicroLamports,omitempty"` // 优先 CU 单价, 与总价二选一 (CU 限制默认为 1400000)
	PrioritizationFeeLamports     *uint64 `json:"prioritizationFeeLamports,omitempty"`     // 优先 CU 总价, 与单价二选一 (CU 限制默认为 1400000)
	AsLegacyTransaction           *bool   `json:"asLegacyTransaction,omitempty"`           // 使用传统交易
	UseTokenLedger                *bool   `json:"useTokenLedger,omitempty"`                // 使用交易内 swap 前的代币流入值作为 swap in 值, 即忽视 quote 的确切值, 而使用交易内代币流入值
	DestinationTokenAccount       *string `json:"destinationTokenAccount,omitempty"`       // B 代币的接收账户, 默认为钱包的 ATA 账户
	DynamicComputeUnitLimit       *bool   `json:"dynamicComputeUnitLimit,omitempty"`       // 动态 CU 限制, 会增加响应耗时
	SkipUserAccountsRpcCalls      *bool   `json:"skipUserAccountsRpcCalls,omitempty"`      // 跳过账户检查, 如 wSOL 账户、 ATA 账户是否存在等
	DynamicSlippage               *struct {
		MinBps slippageBps `json:"minBps"`
		MaxBps slippageBps `json:"maxBps"`
	} `json:"dynamicSlippage,omitempty"` // 动态滑点
}

type createSwapTransactionRequest struct {
	UserPublicKey string `json:"userPublicKey"`
	CreateSwapTransactionOptionalArgs
	QuoteResponse GetQuoteResponse `json:"quoteResponse"`
}

type CreateSwapTransactionResponse struct {
	SwapTransaction           string  `json:"swapTransaction"`
	LastValidBlockHeight      uint64  `json:"lastValidBlockHeight"`
	PrioritizationFeeLamports *uint64 `json:"prioritizationFeeLamports"`
	DynamicSlippageReport     *struct {
		SlippageBps                  slippageBps `json:"slippageBps"`
		OtherAmount                  int32       `json:"otherAmount"`
		SimulatedIncurredSlippageBps slippageBps `json:"simulatedIncurredSlippageBps"`
	} `json:"dynamicSlippageReport"`
}

func (node *RPC) CreateSwapTransaction(ctx g.Ctx, userPublicKey string, quoteResponse GetQuoteResponse, optionalArgs ...CreateSwapTransactionOptionalArgs) (swap CreateSwapTransactionResponse, err error) {
	step := "创建兑换交易"
	request := createSwapTransactionRequest{
		UserPublicKey: userPublicKey,
		QuoteResponse: quoteResponse,
	}
	if len(optionalArgs) > 0 {
		request.CreateSwapTransactionOptionalArgs = optionalArgs[0]
	}
	resp, err := node.client.R().
		SetBodyJsonMarshal(request).
		Post("/swap")
	if err != nil {
		err = gerror.Wrapf(err, "发送%s请求失败", step)
		return
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			var result jupiterTypes.ErrorResponse
			err = gjson.DecodeTo(resp.Bytes(), &result)
			if err != nil {
				err = gerror.Wrapf(err, "解析%s响应失败", step)
				return
			}

			if result.ErrorCode != "" {
				err = gerror.NewCode(result.ErrorCode)
			} else if result.ErrorCodeOld != "" {
				err = gerror.NewCode(result.ErrorCodeOld)
			} else {
				err = gerror.New(resp.String())
			}
			err = gerror.Wrap(err, result.Error)
			switch gerror.Code(err) {
			case jupiterTypes.NotSupported,
				jupiterTypes.CircularArbitrageIsDisabled,
				jupiterTypes.NoRoutesFound,
				jupiterTypes.CouldNotFindAnyRoute,
				jupiterTypes.TokenNotTradable,
				jupiterTypes.RoutePlanDoesNotConsumeAllTheAmount:
				err = gerror.WrapCode(errcode.FatalError, err)
			}
		} else {
			err = gerror.Newf("HTTP %d, %s", resp.StatusCode, resp.Status)
		}
		err = gerror.Wrapf(err, "服务器响应%s请求失败", step)
		return
	}

	var result CreateSwapTransactionResponse
	err = gjson.DecodeTo(resp.Bytes(), &result)
	if err != nil {
		err = gerror.Wrapf(err, "解析%s响应失败", step)
		return
	}

	swap = result

	return
}
