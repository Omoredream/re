package jupiterHTTP

import (
	"net/http"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/errcode"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/types"
)

type CreateSwapTransactionResponse struct {
	SwapTransaction           string  `json:"swapTransaction"`
	LastValidBlockHeight      uint64  `json:"lastValidBlockHeight"`
	PrioritizationFeeLamports *uint64 `json:"prioritizationFeeLamports"`
	DynamicSlippageReport     *struct {
		SlippageBps                  uint16 `json:"slippageBps"`
		OtherAmount                  int32  `json:"otherAmount"`
		SimulatedIncurredSlippageBps uint16 `json:"simulatedIncurredSlippageBps"`
	} `json:"dynamicSlippageReport"`
}

func (node *RPC) CreateSwapTransaction(ctx g.Ctx, userPublicKey string, quoteResponse GetQuoteResponse, optionalArgs ...jupiterTypes.SwapOption) (swap CreateSwapTransactionResponse, err error) {
	step := "创建兑换交易"
	request := map[string]any{
		"userPublicKey": userPublicKey,
		"quoteResponse": quoteResponse,
	}
	for _, optionalArg := range optionalArgs {
		paramName, paramValue := optionalArg()
		request[paramName] = paramValue
	}

	resp, err := node.client.R().
		SetBodyJsonMarshal(request).
		Post("/swap")
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

	var result CreateSwapTransactionResponse
	err = gjson.DecodeTo(resp.Bytes(), &result)
	if err != nil {
		err = gerror.Wrapf(err, "解析%s响应失败", step)
		return
	}

	swap = result

	return
}
