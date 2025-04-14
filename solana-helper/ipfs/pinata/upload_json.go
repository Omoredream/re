package Pinata

import (
	"net/http"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
)

type UploadJsonResult struct {
	IpfsHash    string      `json:"IpfsHash"`
	PinSize     int         `json:"PinSize"`
	Timestamp   *gtime.Time `json:"Timestamp"`
	IsDuplicate bool        `json:"isDuplicate"`
}

func (p *Pinata) UploadJson(filename string, json any) (cid string, err error) {
	resp, err := p.client.R().
		SetBodyJsonBytes(gjson.MustEncode(map[string]any{
			"pinataContent": json,
			"pinataMetadata": map[string]any{
				"name": filename,
			},
			"pinataOptions": map[string]any{
				"cidVersion": 1,
			},
		})).
		Post("/pinning/pinJSONToIPFS")
	if err != nil {
		err = gerror.Wrap(err, "上传 JSON 失败")
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = gerror.Newf("HTTP %d, %s", resp.StatusCode, resp.Status)
		return
	}

	var result UploadJsonResult
	err = gjson.DecodeTo(resp.Bytes(), &result)
	if err != nil {
		err = gerror.Wrap(err, "解析上传 JSON 结果失败")
		return
	}

	cid = result.IpfsHash

	return
}
