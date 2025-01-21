package rpc

import (
	"net/http"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"

	"git.wkr.moe/web3/solana-helper/errcode"

	jupiterTypes "git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/types"
)

type SwapMode string

const (
	ExactIn  SwapMode = "ExactIn"
	ExactOut SwapMode = "ExactOut"
)

type slippageBps = uint16
type platformFeeBps = uint8

type GetQuoteOptionalArgs struct {
	SlippageBps                   *slippageBps    // 滑点, 万分位, 如 50 = 0.5%
	SwapMode                      *SwapMode       // 交易模式, 精确输入/精确输出
	Dexes                         []string        // 白名单 Dex
	ExcludeDexes                  []string        // 黑名单 Dex
	RestrictIntermediateTokens    *bool           // 仅使用流动性稳定的代币作为中转代币
	OnlyDirectRoutes              *bool           // 不使用中转代币
	AsLegacyTransaction           *bool           // 使用传统交易
	PlatformFeeBps                *platformFeeBps // 平台抽成, 用途不明
	MaxAccounts                   *uint           // 交易中涉及账户的最大数量, 避免交易过大
	AutoSlippage                  *bool           // 自动滑点
	MaxAutoSlippageBps            *slippageBps    // 自动滑点的最大值
	AutoSlippageCollisionUsdValue *uint64         // 自动滑点的美元影响因子, 用途不明
}

type GetQuoteResponse struct {
	InputMint            string      `json:"inputMint"`
	InAmount             uint64      `json:"inAmount,string"`
	OutputMint           string      `json:"outputMint"`
	OutAmount            uint64      `json:"outAmount,string"`
	OtherAmountThreshold uint64      `json:"otherAmountThreshold,string"`
	SwapMode             SwapMode    `json:"swapMode"`
	SlippageBps          slippageBps `json:"slippageBps"`
	PlatformFee          *struct {
		Amount *uint64         `json:"amount,string"`
		FeeBps *platformFeeBps `json:"feeBps"`
	} `json:"platformFee,omitempty"`
	PriceImpactPct float64 `json:"priceImpactPct,string"`
	RoutePlan      []struct {
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
	} `json:"routePlan"`
	ContextSlot *uint64  `json:"contextSlot,omitempty"`
	TimeTaken   *float64 `json:"timeTaken,omitempty"`
}

func (node *RPC) GetQuote(ctx g.Ctx, inputMint string, outputMint string, amount uint64, optionalArgs ...GetQuoteOptionalArgs) (quote GetQuoteResponse, err error) {
	step := "获取报价"
	queryParams := map[string]any{
		"inputMint":  inputMint,
		"outputMint": outputMint,
		"amount":     amount,
	}
	if len(optionalArgs) > 0 {
		if optionalArgs[0].SlippageBps != nil {
			queryParams["slippageBps"] = *optionalArgs[0].SlippageBps
		}
		if optionalArgs[0].SwapMode != nil {
			queryParams["swapMode"] = *optionalArgs[0].SwapMode
		}
		if optionalArgs[0].Dexes != nil {
			queryParams["dexes"] = gstr.Join(optionalArgs[0].Dexes, ",")
		}
		if optionalArgs[0].ExcludeDexes != nil {
			queryParams["excludeDexes"] = gstr.Join(optionalArgs[0].ExcludeDexes, ",")
		}
		if optionalArgs[0].RestrictIntermediateTokens != nil {
			queryParams["restrictIntermediateTokens"] = *optionalArgs[0].RestrictIntermediateTokens
		}
		if optionalArgs[0].OnlyDirectRoutes != nil {
			queryParams["onlyDirectRoutes"] = *optionalArgs[0].OnlyDirectRoutes
		}
		if optionalArgs[0].AsLegacyTransaction != nil {
			queryParams["asLegacyTransaction"] = *optionalArgs[0].AsLegacyTransaction
		}
		if optionalArgs[0].PlatformFeeBps != nil {
			queryParams["platformFeeBps"] = *optionalArgs[0].PlatformFeeBps
		}
		if optionalArgs[0].MaxAccounts != nil {
			queryParams["maxAccounts"] = *optionalArgs[0].MaxAccounts
		}
		if optionalArgs[0].AutoSlippage != nil {
			queryParams["autoSlippage"] = *optionalArgs[0].AutoSlippage
		}
		if optionalArgs[0].MaxAutoSlippageBps != nil {
			queryParams["maxAutoSlippageBps"] = *optionalArgs[0].MaxAutoSlippageBps
		}
		if optionalArgs[0].AutoSlippageCollisionUsdValue != nil {
			queryParams["autoSlippageCollisionUsdValue"] = *optionalArgs[0].AutoSlippageCollisionUsdValue
		}
	}
	resp, err := node.client.R().
		SetQueryParamsAnyType(queryParams).
		Get("/quote")
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

	var result GetQuoteResponse
	err = gjson.DecodeTo(resp.Bytes(), &result)
	if err != nil {
		err = gerror.Wrapf(err, "解析%s响应失败", step)
		return
	}

	quote = result

	return
}
