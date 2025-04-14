package officialRPCs

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcache"

	"git.wkr.moe/web3/solana-helper/rpcinterface"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/official/rpc/http"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/official/rpc/websocket"
)

type RPCs struct {
	httpRPCs      rpcinterface.RPCs[*officialHTTP.RPC]
	websocketRPCs rpcinterface.RPCs[*officialWebSocket.RPC]

	tokenCache              *gcache.Cache
	addressLookupTableCache *gcache.Cache
}

func New() (pool RPCs) {
	var tokenCacheAdapter gcache.Adapter
	if g.Cfg().MustGet(nil, "cache.token.medium", "memory").String() == "redis" {
		tokenCacheAdapter = gcache.NewAdapterRedis(g.Redis("token"))
	} else {
		tokenCacheAdapter = gcache.NewAdapterMemory()
	}

	var addressLookupTableCacheAdapter gcache.Adapter
	if g.Cfg().MustGet(nil, "cache.addressLookupTable.medium", "memory").String() == "redis" {
		addressLookupTableCacheAdapter = gcache.NewAdapterRedis(g.Redis("addressLookupTable"))
	} else {
		addressLookupTableCacheAdapter = gcache.NewAdapterMemory()
	}

	pool = RPCs{
		httpRPCs:      rpcinterface.News[*officialHTTP.RPC](),
		websocketRPCs: rpcinterface.News[*officialWebSocket.RPC](),

		tokenCache:              gcache.NewWithAdapter(tokenCacheAdapter),
		addressLookupTableCache: gcache.NewWithAdapter(addressLookupTableCacheAdapter),
	}

	return
}

func (pool *RPCs) AddHttpRPC(httpRPC ...*officialHTTP.RPC) {
	pool.httpRPCs.AddRPC(httpRPC...)
}

func (pool *RPCs) AddWebsocketRPC(websocketRPC ...*officialWebSocket.RPC) {
	pool.websocketRPCs.AddRPC(websocketRPC...)
}

func (pool *RPCs) HttpCount() int {
	return pool.httpRPCs.Count()
}

func (pool *RPCs) GetHttpNames() (names []string) {
	names = make([]string, 0, pool.httpRPCs.Count())
	pool.httpRPCs.Each(func(chosen *officialHTTP.RPC) {
		names = append(names, chosen.Name())
	})
	return
}

func (pool *RPCs) WebsocketCount() int {
	return pool.httpRPCs.Count()
}

func (pool *RPCs) GetWebsocketNames() (names []string) {
	names = make([]string, 0, pool.websocketRPCs.Count())
	pool.websocketRPCs.Each(func(chosen *officialWebSocket.RPC) {
		names = append(names, chosen.Name())
	})
	return
}
