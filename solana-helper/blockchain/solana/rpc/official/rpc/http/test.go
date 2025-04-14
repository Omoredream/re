package officialHTTP

import (
	"context"
	"strconv"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"

	"github.com/gagliardetto/solana-go/rpc"

	"git.wkr.moe/web3/solana-helper/consts"
)

func (node *RPC) test(ctx g.Ctx) (err error) {
	ctx = context.WithValue(ctx, consts.CtxRPC, node.Name())

	if node.fingerprint == "" { // 指纹只在首次初始化时计算, 无需更新
		var getIdentityResult *rpc.GetIdentityResult
		getIdentityResult, err = node.client.GetIdentity(ctx)
		if err != nil {
			err = gerror.Wrapf(err, "获取 RPC 公钥失败")
			return
		}

		node.fingerprint = getIdentityResult.Identity.String()
	}

	getVersionResult, err := node.client.GetVersion(ctx)
	if err != nil {
		err = gerror.Wrapf(err, "获取 RPC 版本失败")
		return
	}

	node.version = make([]uint, 0, 3)
	for _, vers := range gstr.Split(getVersionResult.SolanaCore, ".") {
		var ver int
		ver, err = strconv.Atoi(vers)
		if err != nil {
			err = gerror.Wrapf(err, "解析版本号失败")
			return
		}
		node.version = append(node.version, uint(ver))
	}
	g.Log().Debugf(ctx, "RPC 版本 %s, 特性 %d", getVersionResult.SolanaCore, getVersionResult.FeatureSet)

	return
}
