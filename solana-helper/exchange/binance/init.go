package Binance

import (
	"github.com/adshao/go-binance/v2"
	"github.com/gogf/gf/v2/os/gcache"
)

var client *binance.Client
var memoryCache *gcache.Cache

func init() {
	client = binance.NewClient("", "")
	memoryCache = gcache.NewWithAdapter(gcache.NewAdapterMemory())
}
