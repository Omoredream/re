package tests

import (
	"github.com/gogf/gf/v2/crypto/gaes"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/mr-tron/base58"
	"github.com/samber/lo"
)

var (
	testWalletWIF      = g.Cfg().MustGet(nil, "wallets.wif").String()
	testWalletMnemonic = g.Cfg().MustGet(nil, "wallets.mnemonic").String()
	testPinataToken    = g.Cfg().MustGet(nil, "wallets.pinataToken").String()
)

func init() {
	if len(testWalletWIF) == 108 {
		testWalletWIF = base58.Encode(lo.Must(gaes.Decrypt(gbase64.MustDecodeString(testWalletWIF), []byte("0153d185f9854ef4aab9753ef725173c"), []byte("a2bf2e6ed9a94b93"))))
	}
}
