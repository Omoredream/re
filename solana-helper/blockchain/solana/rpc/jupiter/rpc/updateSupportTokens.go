package jupiterHTTP

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/container/garray"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/grand"

	"git.wkr.moe/web3/solana-helper/errcode"

	"git.wkr.moe/web3/solana-helper/consts"
)

func (node *RPC) updateSupportTokens(ctx g.Ctx) (err error) {
	ctx = context.WithValue(ctx, consts.CtxRPC, node.Name())

	getTokensResult, err := node.GetTokens(ctx)
	if err != nil {
		err = gerror.Wrapf(err, "获取 RPC 支持代币列表失败")
		return
	}
	if node.fingerprint == "" { // 指纹只在首次初始化时计算, 无需更新
		node.fingerprint = gmd5.MustEncrypt(gjson.MustEncode(garray.NewSortedStrArrayFromCopy(getTokensResult).Slice()))
	}

	node.supportTokens = gset.NewStrSetFrom(getTokensResult, true)

	return
}

func (node *RPC) Support(targetTokens []string) bool {
	for _, targetToken := range targetTokens {
		if !node.supportTokens.Contains(targetToken) {
			return false
		}
	}
	return true
}

func (node *RPC) GetRandomSupportToken(excludes ...string) (token string, err error) {
	supportTokens := node.supportTokens.Diff(gset.NewStrSetFrom(excludes, false)).Slice()

	if len(supportTokens) == 0 {
		err = gerror.NewCodef(errcode.IgnoreError, "无可用的支持代币")
		node.AddCoolDown((1 * time.Minute).Milliseconds())
		return
	}

	token = supportTokens[grand.Intn(len(supportTokens))]
	return
}
