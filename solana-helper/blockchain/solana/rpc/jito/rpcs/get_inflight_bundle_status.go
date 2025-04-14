package jitoRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jito/types"
)

func (pool *RPCs) GetInflightBundleStatus(ctx g.Ctx, bundleId string) (bundleStatus jitoTypes.GetInflightBundleStatusesResponse, err error) {
	bundlesStatus, err := pool.httpGetInflightBundlesStatus(ctx, bundleId)
	if err != nil {
		err = gerror.Wrapf(err, "获取捆绑交易状态失败")
		return
	}

	bundleStatus = bundlesStatus[0]

	return
}

func (pool *RPCs) GetInflightBundlesStatus(ctx g.Ctx, bundlesId ...string) (bundlesStatus []jitoTypes.GetInflightBundleStatusesResponse, err error) {
	bundlesStatus, err = pool.httpGetInflightBundlesStatus(ctx, bundlesId...)
	if err != nil {
		err = gerror.Wrapf(err, "获取捆绑交易状态失败")
		return
	}

	return
}
