package jitoRPCs

import (
	"git.wkr.moe/web3/solana-helper/rpcinterface"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jito/rpc"
)

type RPCs struct {
	httpRPCs rpcinterface.RPCs[*jitoHTTP.RPC]
}

func New() (pool RPCs) {
	pool = RPCs{
		httpRPCs: rpcinterface.News[*jitoHTTP.RPC](),
	}

	return
}

func (pool *RPCs) AddRPC(rpc ...*jitoHTTP.RPC) {
	pool.httpRPCs.AddRPC(rpc...)
}

func (pool *RPCs) Count() int {
	return pool.httpRPCs.Count()
}

func (pool *RPCs) GetNames() (names []string) {
	names = make([]string, 0, pool.httpRPCs.Count())
	pool.httpRPCs.Each(func(chosen *jitoHTTP.RPC) {
		names = append(names, chosen.Name())
	})
	return
}
