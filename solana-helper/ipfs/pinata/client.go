package Pinata

import (
	"github.com/imroc/req/v3"
)

type Pinata struct {
	client *req.Client
}

func NewPinata(token string) *Pinata {
	return &Pinata{
		client: req.C().
			SetCommonHeaders(map[string]string{
				"Authorization": "Bearer " + token,
			}).
			SetBaseURL("https://api.pinata.cloud"),
	}
}
