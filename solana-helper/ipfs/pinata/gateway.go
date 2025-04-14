package Pinata

import (
	"fmt"
)

func (p *Pinata) Gateway(cid string) (url string) {
	url = fmt.Sprintf("https://gateway.pinata.cloud/ipfs/%s", cid)
	return
}
