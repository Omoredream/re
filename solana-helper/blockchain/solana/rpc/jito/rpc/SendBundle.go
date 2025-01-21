package rpc

import (
	"net/http"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/errcode"
)

func (node *RPC) SendBundle(ctx g.Ctx, txs []string) (bundleId string, err error) {
	if len(txs) > 5 {
		err = gerror.Newf("单个捆绑包中交易过多")
		return
	}
	step := "发送捆绑包"
	resp, err := node.client().R().
		SetBodyJsonMarshal(apiRequest{
			JsonRPC: "2.0",
			ID:      1,
			Method:  "sendBundle",
			Params: []any{
				txs,
				map[string]string{
					"encoding": "base64",
				},
			},
		}).
		Post("/api/v1/bundles")
	if err != nil {
		err = gerror.Wrapf(err, "发送%s请求失败", step)
		return
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest && resp.StatusCode != http.StatusTooManyRequests {
		err = gerror.Newf("HTTP %d, %s", resp.StatusCode, resp.Status)
		err = gerror.Wrapf(err, "服务器响应%s请求失败", step)
		return
	}

	var result apiResponse[string]
	err = gjson.DecodeTo(resp.Bytes(), &result)
	if err != nil {
		err = gerror.Wrapf(err, "解析%s响应失败", step)
		return
	}

	if result.Error != nil {
		err = gerror.NewCode(gcode.New(result.Error.Code, result.Error.Message, nil))
		switch result.Error.Code {
		case -32602, // bundle contains an expired blockhash / bundle contains an already processed transaction
			-32603: // transaction has too many account locks
			err = gerror.WrapCode(errcode.FatalError, err)
		case -32097: // Rate limit exceeded. Limit: x per second for txn requests
			g.Log().Warningf(ctx, "触发 Jito 频率限制, %v", err)
		}
		err = gerror.Wrapf(err, "服务器响应%s请求失败", step)
		return
	}

	bundleId = *result.Result

	return
}
