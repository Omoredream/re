package tests

import (
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/project/arb"
)

func TestArb(t *testing.T) {
	err := testArb(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testArb(ctx g.Ctx) (err error) {
	arb, err := ProjectArb.New(ctx, &officialPool, &jupiterPool, &jitoPool, Region, testWalletWIF, testWalletMnemonic)
	if err != nil {
		err = gerror.Wrapf(err, "初始化套利项目失败")
		return
	}

	arb.Loop(ctx)
	return
}
