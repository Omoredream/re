package jupiterRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (pool *RPCs) GetRandomSupportToken(ctx g.Ctx, excludes ...string) (token string, err error) {
	token, err = pool.httpGetRandomSupportToken(ctx, excludes...)
	if err != nil {
		err = gerror.Wrapf(err, "随机选择支持代币失败")
		return
	}

	return
}
