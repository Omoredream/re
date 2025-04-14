package jupiterHTTP

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/types"
)

func (node *RPC) test(ctx g.Ctx) (err error) {
	ctx = context.WithValue(ctx, consts.CtxRPC, node.Name())

	err = node.updateSupportTokens(ctx)
	if err != nil {
		err = gerror.Wrapf(err, "测试 RPC 连通性失败")
		return
	}

	pumpFunTokensCount := 0
	node.supportTokens.Iterator(func(supportToken string) bool {
		if gstr.HasSuffix(supportToken, "pump") {
			pumpFunTokensCount++
		}
		return true
	})

	g.Log().Debugf(ctx, "RPC 支持 %d 个常规代币, 以及 %d 个 PumpFun 代币", node.supportTokens.Size()-pumpFunTokensCount, pumpFunTokensCount)
	node.tooManyToken = node.supportTokens.Size() > 500

	_, err = node.GetQuote(ctx, consts.SOL.Address.String(), consts.SOL.Address.String(), 10000)
	if err != nil {
		if node.arbDisabled = gerror.HasCode(err, jupiterTypes.CircularArbitrageIsDisabled); node.arbDisabled {
			err = nil
		} else {
			return
		}
	}

	return
}
