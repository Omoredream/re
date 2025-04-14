package Pinata

import (
	"net/http"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
)

type UploadFileResult struct {
	IpfsHash    string      `json:"IpfsHash"`
	PinSize     int         `json:"PinSize"`
	Timestamp   *gtime.Time `json:"Timestamp"`
	IsDuplicate bool        `json:"isDuplicate"`
}

func (p *Pinata) UploadFile(filename string, data []byte) (cid string, err error) {
	resp, err := p.client.R().
		SetFileBytes("file", filename, data).
		SetFormData(map[string]string{
			"pinataMetadata": gjson.MustEncodeString(map[string]any{
				"name": filename,
			}),
			"pinataOptions": gjson.MustEncodeString(map[string]any{
				"cidVersion": 1,
			}),
		}).
		Post("/pinning/pinFileToIPFS")
	if err != nil {
		err = gerror.Wrap(err, "上传文件失败")
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = gerror.Newf("HTTP %d, %s", resp.StatusCode, resp.Status)
		return
	}

	var result UploadFileResult
	err = gjson.DecodeTo(resp.Bytes(), &result)
	if err != nil {
		err = gerror.Wrap(err, "解析上传文件结果失败")
		return
	}

	cid = result.IpfsHash

	return
}
