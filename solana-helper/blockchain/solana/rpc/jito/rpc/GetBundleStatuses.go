package jitoHTTP

import (
	"net/http"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jito/types"
)

func (node *RPC) GetBundleStatuses(ctx g.Ctx, bundleIds []string) (bundleStatuses []*jitoTypes.GetBundleStatusesResponse, err error) {
	step := "获取捆绑包状态列表"
	resp, err := node.client().R().
		SetBodyJsonMarshal(apiRequest{
			JsonRPC: "2.0",
			ID:      1,
			Method:  "getBundleStatuses",
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

	var result apiResponse[apiResponseContext[[]jitoTypes.GetBundleStatusesResponse]]
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

	bundleStatuses = make([]*jitoTypes.GetBundleStatusesResponse, len(bundleIds))
	for i, bundleId := range bundleIds {
		for j, bundleStatus := range (*result.Result).Value {
			if bundleId == bundleStatus.BundleId {
				bundleStatuses[i] = &bundleStatus
				(*result.Result).Value = append((*result.Result).Value[:j], (*result.Result).Value[j+1:]...)
				break
			}
		}
	}

	return
}
