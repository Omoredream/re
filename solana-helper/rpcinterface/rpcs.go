package rpcinterface

import (
	"context"
	"maps"
	"slices"

	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/samber/lo"

	"git.wkr.moe/web3/solana-helper/errcode"
)

type RPCs[T RPCInterface] struct {
	rpcs     map[string]T
	rpcNames *gset.StrSet
}

func News[T RPCInterface]() (pool RPCs[T]) {
	pool = RPCs[T]{
		rpcs:     make(map[string]T),
		rpcNames: gset.NewStrSet(),
	}

	return
}

func (pool *RPCs[T]) AddRPC(rpcs ...T) {
	for _, rpc := range rpcs {
		if !pool.rpcNames.AddIfNotExist(rpc.Name()) {
			g.Log().Warningf(context.Background(), "RPC %s 存在重复名称", rpc.Name())
		}

		fp := rpc.Fingerprint()
		if rpcExisted, exist := pool.rpcs[fp]; exist {
			g.Log().Warningf(context.Background(), "RPC %s 与 RPC %s 可能重复", rpc.Name(), rpcExisted.Name())
			continue
		}
		pool.rpcs[fp] = rpc
	}
}

func (pool *RPCs[T]) ChooseOne(conditions ...func(T) bool) (chosen T, err error) {
	if len(pool.rpcs) == 0 {
		err = gerror.Newf("无可用 RPC 节点")
		return
	}
	choiced := false
	for _, i := range lo.Shuffle(slices.Collect(maps.Keys(pool.rpcs))) {
		if !pool.rpcs[i].IsCoolDown() {
			conditionsResult := true
			for _, condition := range conditions {
				conditionsResult = conditionsResult && condition(pool.rpcs[i])
				if !conditionsResult {
					break
				}
			}
			if !conditionsResult {
				continue
			}

			if meetProb := grand.MeetProb(0.75); !choiced || // 未选择任何节点
				(chosen.Weight() > pool.rpcs[i].Weight() && meetProb || // 75% 概率在某节点更优时选择
					chosen.Weight() <= pool.rpcs[i].Weight() && !meetProb) { // 25% 概率选择其他节点
				if pool.rpcs[i].RunMoreThread() {
					if choiced {
						chosen.EndThread()
					}
					choiced = true
					chosen = pool.rpcs[i]
				}
			}
		}
	}
	if !choiced {
		err = gerror.NewCodef(errcode.IgnoreError, "无满足条件 RPC 节点")
		err = gerror.WrapCode(errcode.CoolDownError, err)
		return
	}

	return
}

func (pool *RPCs[T]) Count() int {
	return len(pool.rpcs)
}

func (pool *RPCs[T]) Each(do func(chosen T)) {
	for _, rpc := range pool.rpcs {
		do(rpc)
	}
}

func (pool *RPCs[T]) Get(fp string) T {
	return pool.rpcs[fp]
}
