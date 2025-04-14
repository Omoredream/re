package Binance

import (
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/utils"
)

func GetPrice(ctx g.Ctx, symbol string) (price decimal.Decimal, err error) {
	err = Utils.Retry(ctx, func() (err error) {
		priceV, err := memoryCache.GetOrSetFuncLock(ctx, fmt.Sprintf("price/%s", symbol), func(ctx g.Ctx) (price any, err error) {
			service := client.NewListPricesService()
			result, err := service.Symbol(symbol).Do(ctx)
			if err != nil {
				err = gerror.Wrapf(err, "获取交易对 %s 最新价格失败", symbol)
				return
			}

			price, err = decimal.NewFromString(result[0].Price)
			if err != nil {
				err = gerror.Wrapf(err, "解析价格 %s 失败", result[0].Price)
				return
			}

			return
		}, 10*time.Second)
		if err != nil {
			err = gerror.Wrapf(err, "经缓存获取交易对 %s 最新价格失败", symbol)
			return
		}

		var ok bool
		price, ok = priceV.Interface().(decimal.Decimal)
		if !ok {
			err = gerror.Newf("从内存读取价格 %+v 失败", priceV.Interface())
			return
		}

		return
	})

	return
}
