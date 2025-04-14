package officialHTTP

import (
	"io"
	"net/http"

	"github.com/imroc/req/v3"
)

type clientWrapper struct {
	req *req.Client
}

func (c *clientWrapper) Do(rawReq *http.Request) (rawResp *http.Response, err error) {
	r := c.req.R()
	r.Method = rawReq.Method
	r.RawURL = rawReq.URL.String()
	r.Headers = rawReq.Header
	r.Body = make([]byte, rawReq.ContentLength)
	r.GetBody = func() (io.ReadCloser, error) {
		return rawReq.Body, nil
	}
	if rawReq.Close {
		r.EnableCloseConnection()
	}
	if rawReq.Host != "" {
		r.SetHeader("Host", rawReq.Host)
	}
	if ctx := rawReq.Context(); ctx != nil {
		r.SetContext(ctx)
	}

	resp := r.Do()
	if resp.Err != nil {
		err = resp.Err
		return
	}
	rawResp = resp.Response

	return
}

func (c *clientWrapper) CloseIdleConnections() {
	c.req.CloseIdleConnections()
	return
}
