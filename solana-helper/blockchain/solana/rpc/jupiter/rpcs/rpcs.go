package jupiterRPCs

import (
	"git.wkr.moe/web3/solana-helper/rpcinterface"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/rpc"
)

type RPCs struct {
	httpRPCs rpcinterface.RPCs[*jupiterHTTP.RPC]
}

func New() (pool RPCs) {
	pool = RPCs{
		httpRPCs: rpcinterface.News[*jupiterHTTP.RPC](),
	}

	return
}

func (pool *RPCs) AddRPC(rpc ...*jupiterHTTP.RPC) {
	pool.httpRPCs.AddRPC(rpc...)
}

func (pool *RPCs) Count() int {
	return pool.httpRPCs.Count()
}

func (pool *RPCs) GetNames() (names []string) {
	names = make([]string, 0, pool.httpRPCs.Count())
	pool.httpRPCs.Each(func(chosen *jupiterHTTP.RPC) {
		names = append(names, chosen.Name())
	})
	return
}

func (pool *RPCs) Get(fp string) *jupiterHTTP.RPC {
	return pool.httpRPCs.Get(fp)
}
