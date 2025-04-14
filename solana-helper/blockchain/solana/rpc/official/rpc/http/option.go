package officialHTTP

import (
	"github.com/imroc/req/v3"
)

type Option func(*req.Client)

func OptionCustomQueryParams(params map[string]string) func(client *req.Client) {
	return func(client *req.Client) {
		client.SetCommonQueryParams(params)
	}
}

func OptionCustomHeaders(headers map[string]string) func(client *req.Client) {
	return func(client *req.Client) {
		client.SetCommonHeaders(headers)
	}
}

func OptionCustomUA(ua string) func(client *req.Client) {
	return func(client *req.Client) {
		client.SetUserAgent(ua)
	}
}

func OptionBrowser(client *req.Client) {
	client.ImpersonateChrome()
}
