package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
)

var tokenCacheDuration = g.Cfg().MustGet(nil, "cache.token.duration", "1h").Duration() // 代币可能会更新一些信息, 所以给予 1h 有效期

func (pool *RPCs) TokenCacheGet(ctx g.Ctx, address Address.TokenAddress) (token Token.Token, err error) {
	tokenV, err := pool.tokenCache.GetOrSetFuncLock(ctx, address, func(ctx g.Ctx) (token any, err error) {
		return pool.getToken(ctx, address)
	}, tokenCacheDuration)
	if err != nil { // 当使用内存缓存时, 此处 `err` 只可能由 `get` 函数引发
		err = gerror.Wrapf(err, "获取代币失败")
		return
	}
	err = tokenV.Struct(&token)
	if err != nil {
		err = gerror.Wrapf(err, "反序列化数据失败")
		return
	}

	return
}

func (pool *RPCs) TokenCacheGets(ctx g.Ctx, addresses []Address.TokenAddress) (tokens []Token.Token, err error) {
	tokens = make([]Token.Token, len(addresses))
	noCacheTokens := make([]Address.TokenAddress, 0, len(addresses))
	noCacheTokensIndex := make([]int, 0, len(addresses))
	for i := range addresses {
		tokenV := pool.tokenCache.MustGet(ctx, addresses[i])
		if !tokenV.IsNil() {
			err = tokenV.Struct(&tokens[i])
			if err != nil {
				err = gerror.Wrapf(err, "反序列化数据失败")
				return
			}
		} else {
			noCacheTokens = append(noCacheTokens, addresses[i])
			noCacheTokensIndex = append(noCacheTokensIndex, i)
		}
	}

	needCacheTokens, err := pool.getTokens(ctx, noCacheTokens)
	if err != nil {
		err = gerror.Wrapf(err, "获取未缓存代币失败")
		return
	}
	pool.TokenCacheSets(ctx, needCacheTokens)
	for i, noCacheTokenIndex := range noCacheTokensIndex {
		tokens[noCacheTokenIndex] = needCacheTokens[i]
	}

	return
}

func (pool *RPCs) TokenCacheSet(ctx g.Ctx, token Token.Token) {
	_ = pool.tokenCache.Set(ctx, token.Address, token, tokenCacheDuration)

	return
}

func (pool *RPCs) TokenCacheSets(ctx g.Ctx, tokens []Token.Token) {
	_ = pool.tokenCache.SetMap(ctx, func() (data map[any]any) {
		data = make(map[any]any, len(tokens))
		for i := range tokens {
			data[tokens[i].Address] = tokens[i]
		}

		return
	}(), tokenCacheDuration)

	return
}
