package tests

import (
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

func TestTokenHolders(t *testing.T) {
	err := testTokenHolders(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testTokenHolders(ctx g.Ctx) (err error) {
	address := Address.NewFromBase58("7atgF8KQo4wJrD5ATGX7t1V2zVvykPJbFfNeVf1icFv1").AsTokenAddress()

	token, err := officialPool.TokenCacheGet(ctx, address)
	if err != nil {
		err = gerror.Wrapf(err, "查询代币失败")
		return
	}

	g.Log().Infof(ctx, "代币: %+v", token)

	tokenHolders, err := officialPool.GetTokenHolders(ctx, token)
	if err != nil {
		err = gerror.Wrapf(err, "查询代币持有者失败")
		return
	}

	for _, tokenHolder := range tokenHolders {
		g.Log().Infof(ctx, "持有者: %+v", tokenHolder)
	}

	return
}
