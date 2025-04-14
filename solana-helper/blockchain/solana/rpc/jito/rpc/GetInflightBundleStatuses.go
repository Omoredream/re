package jitoHTTP

import (
	"net/http"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jito/types"
)

func (node *RPC) GetInflightBundleStatuses(ctx g.Ctx, bundleIds []string) (inflightBundleStatuses []jitoTypes.GetInflightBundleStatusesResponse, err error) {
	step := "获取处理中捆绑包状态列表"
	resp, err := node.client().R().
		SetBodyJsonMarshal(apiRequest{
			JsonRPC: "2.0",
			ID:      1,
			Method:  "getInflightBundleStatuses",
			Params: []any{
				bundleIds,
			},
		}).
		Post("/api/v1/bundles")
	if err != nil {
		err = gerror.Wrapf(err, "发送%s请求失败", step)
		return
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		err = gerror.Newf("HTTP %d, %s", resp.StatusCode, resp.Status)
		err = gerror.Wrapf(err, "服务器响应%s请求失败", step)
		return
	}

	var result apiResponse[apiResponseContext[[]jitoTypes.GetInflightBundleStatusesResponse]]
	err = gjson.DecodeTo(resp.Bytes(), &result)
	if err != nil {
		err = gerror.Wrapf(err, "解析%s响应失败", step)
		return
	}

	if result.Error != nil {
		err = gerror.Newf("[%d] %s", result.Error.Code, result.Error.Message)
		err = gerror.Wrapf(err, "服务器响应%s请求失败", step)
		return
	}

	inflightBundleStatuses = make([]jitoTypes.GetInflightBundleStatusesResponse, len(bundleIds))
	for i, bundleId := range bundleIds {
		found := false
		for j, inflightBundleStatus := range (*result.Result).Value {
			if bundleId == inflightBundleStatus.BundleId {
				found = true
				inflightBundleStatuses[i] = inflightBundleStatus
				(*result.Result).Value = append((*result.Result).Value[:j], (*result.Result).Value[j+1:]...)
				break
			}
		}
		if !found {
			err = gerror.Newf("未查询到捆绑包 %s 状态", bundleId)
			return
		}
	}

	return
}
