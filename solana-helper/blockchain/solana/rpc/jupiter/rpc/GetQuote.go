package jupiterHTTP

import (
	"fmt"
	"net/http"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/samber/lo"

	"git.wkr.moe/web3/solana-helper/errcode"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/types"
)

type GetQuoteResponse struct {
	InputMint            string `json:"inputMint"`
	InAmount             uint64 `json:"inAmount,string"`
	OutputMint           string `json:"outputMint"`
	OutAmount            uint64 `json:"outAmount,string"`
	OtherAmountThreshold uint64 `json:"otherAmountThreshold,string"`
	SwapMode             string `json:"swapMode"`
	SlippageBps          uint16 `json:"slippageBps"`
	PlatformFee          *struct {
		Amount *uint64 `json:"amount,string"`
		FeeBps *uint8  `json:"feeBps"`
	} `json:"platformFee,omitempty"`
	PriceImpactPct float64   `json:"priceImpactPct,string"`
	RoutePlan      RoutePlan `json:"routePlan"`
	ContextSlot    *uint64   `json:"contextSlot,omitempty"`
	TimeTaken      *float64  `json:"timeTaken,omitempty"`
}

type RoutePlan []RoutePlanStep

func (routePlan RoutePlan) String() string {
	return gstr.Join(lo.Map(routePlan, func(routePlanStep RoutePlanStep, index int) string {
		return routePlanStep.String()
	}), ",")
}

type RoutePlanStep struct {
	SwapInfo struct {
		AmmKey     string  `json:"ammKey"`
		Label      *string `json:"label,omitempty"`
		InputMint  string  `json:"inputMint"`
		OutputMint string  `json:"outputMint"`
		InAmount   uint64  `json:"inAmount,string"`
		OutAmount  uint64  `json:"outAmount,string"`
		FeeAmount  uint64  `json:"feeAmount,string"`
		FeeMint    string  `json:"feeMint"`
	} `json:"swapInfo"`
	Percent uint8 `json:"percent"`
}

func (routePlanStep RoutePlanStep) String() string {
	if routePlanStep.Percent == 100 {
		return fmt.Sprintf("%s->%s", routePlanStep.SwapInfo.InputMint, routePlanStep.SwapInfo.OutputMint)
	} else {
		return fmt.Sprintf("%s-%d%%->%s", routePlanStep.SwapInfo.InputMint, routePlanStep.Percent, routePlanStep.SwapInfo.OutputMint)
	}
}

func (node *RPC) GetQuote(ctx g.Ctx, inputMint string, outputMint string, amount uint64, optionalArgs ...jupiterTypes.QuoteOption) (quote GetQuoteResponse, err error) {
	step := "获取报价"
	queryParams := map[string]any{
		"inputMint":  inputMint,
		"outputMint": outputMint,
		"amount":     amount,
	}
	for _, optionalArg := range optionalArgs {
		paramName, paramValue := optionalArg()
		queryParams[paramName] = paramValue
	}

	resp, err := node.client.R().
		SetQueryParamsAnyType(queryParams).
		Get("/quote")
	if err != nil {
		err = gerror.WrapCodef(errcode.NetworkError, err, "发送%s请求失败", step)
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
			case
				jupiterTypes.NotSupported,
				jupiterTypes.CircularArbitrageIsDisabled:
				err = gerror.WrapCode(errcode.FatalError, err)
			case
				jupiterTypes.NoRoutesFound,
				jupiterTypes.CouldNotFindAnyRoute,
				jupiterTypes.TokenNotTradable,
				jupiterTypes.RoutePlanDoesNotConsumeAllTheAmount:
				err = gerror.WrapCode(errcode.FatalError, err)
				err = gerror.WrapCode(errcode.CoolDownLessError, err) // 与 Jup 节点本身无关的错误, 无需内部冷却
			}
		} else {
			err = gerror.NewCodef(errcode.NetworkError, "HTTP %d, %s: %s", resp.StatusCode, resp.Status, resp.String())
		}
		err = gerror.Wrapf(err, "服务器响应%s请求失败", step)
		return
	}

	var result GetQuoteResponse
	err = gjson.DecodeTo(resp.Bytes(), &result)
	if err != nil {
		err = gerror.Wrapf(err, "解析%s响应失败", step)
		return
	}

	quote = result

	return
}
