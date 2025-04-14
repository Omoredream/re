package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (pool *RPCs) GetLatestSlot(ctx g.Ctx) (slot uint64, err error) {
	slot, err = pool.httpGetSlot(ctx)
	if err != nil {
		err = gerror.Wrapf(err, "获取最新插槽失败")
		return
	}

	return
}
