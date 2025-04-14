package jitoHTTP

import (
	"net/http"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (node *RPC) GetTipAccounts(ctx g.Ctx) (tipAccounts []string, err error) {
	step := "获取小费账户列表"
	resp, err := node.client().R().
		SetBodyJsonMarshal(apiRequest{
			JsonRPC: "2.0",
			ID:      1,
			Method:  "getTipAccounts",
			Params:  []any{},
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

	var result apiResponse[[]string]
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

	tipAccounts = *result.Result

	return
}
