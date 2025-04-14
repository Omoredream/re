package Utils

import (
	"slices"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"git.wkr.moe/web3/solana-helper/errcode"
)

func Try(do retry.RetryableFunc, errCheck func(err error) error, successDo func(spendTime time.Duration), failDo func(err error)) (err error) {
	spendTime := gtime.FuncCost(func() {
		err = do()
	})

	if err != nil && errCheck != nil {
		err = errCheck(err)
	}
	if err != nil {
		if failDo != nil {
			failDo(err)
		}
	} else {
		if successDo != nil {
			successDo(spendTime)
		}
	}

	return
}

func Retry(ctx g.Ctx, do retry.RetryableFunc, options ...retry.Option) error {
	options = slices.Insert(options, 0,
		retry.Attempts(3),
		retry.Delay(2*time.Second),
		retry.DelayType(func(n uint, err error, config *retry.Config) time.Duration {
			return time.Duration(n) * retry.FixedDelay(n, err, config)
		}),
		retry.LastErrorOnly(true),
		retry.RetryIf(func(err error) bool {
			return !gerror.HasCode(err, errcode.FatalError)
		}),
		retry.OnRetry(func(n uint, err error) {
			g.Log("retry").Errorf(ctx, "第 %d 次失败, %+v", n+1, err)
		}),
	)
	return retry.Do(do, options...)
}
