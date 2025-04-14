package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address_lookup_table"
)

var addressLookupTableCacheDuration = g.Cfg().MustGet(nil, "cache.addressLookupTable.duration", "1d").Duration() // 地址查找表可能会更新一些信息, 所以给予 1d 有效期

func (pool *RPCs) AddressLookupTableCacheGet(ctx g.Ctx, address Address.AccountAddress) (addressLookupTable AddressLookupTable.AddressLookupTable, err error) {
	addressLookupTableV, err := pool.addressLookupTableCache.GetOrSetFuncLock(ctx, address, func(ctx g.Ctx) (addressLookupTable any, err error) {
		return pool.getAddressLookupTable(ctx, address)
	}, addressLookupTableCacheDuration)
	if err != nil { // 当使用内存缓存时, 此处 `err` 只可能由 `get` 函数引发
		err = gerror.Wrapf(err, "获取地址查找表失败")
		return
	}
	err = addressLookupTableV.Struct(&addressLookupTable)
	if err != nil {
		err = gerror.Wrapf(err, "反序列化数据失败")
		return
	}

	return
}

func (pool *RPCs) AddressLookupTableCacheGets(ctx g.Ctx, addresses []Address.AccountAddress) (addressLookupTables []AddressLookupTable.AddressLookupTable, err error) {
	addressLookupTables = make([]AddressLookupTable.AddressLookupTable, len(addresses))
	noCacheAddressLookupTables := make([]Address.AccountAddress, 0, len(addresses))
	noCacheAddressLookupTablesIndex := make([]int, 0, len(addresses))
	for i := range addresses {
		addressLookupTableV := pool.addressLookupTableCache.MustGet(ctx, addresses[i])
		if !addressLookupTableV.IsNil() {
			err = addressLookupTableV.Struct(&addressLookupTables[i])
			if err != nil {
				err = gerror.Wrapf(err, "反序列化数据失败")
				return
			}
		} else {
			noCacheAddressLookupTables = append(noCacheAddressLookupTables, addresses[i])
			noCacheAddressLookupTablesIndex = append(noCacheAddressLookupTablesIndex, i)
		}
	}

	needCacheAddressLookupTables, err := pool.getAddressLookupTables(ctx, noCacheAddressLookupTables)
	if err != nil {
		err = gerror.Wrapf(err, "获取未缓存地址查找表失败")
		return
	}
	pool.AddressLookupTableCacheSets(ctx, needCacheAddressLookupTables)
	for i, noCacheAddressLookupTableIndex := range noCacheAddressLookupTablesIndex {
		addressLookupTables[noCacheAddressLookupTableIndex] = needCacheAddressLookupTables[i]
	}

	return
}

func (pool *RPCs) AddressLookupTableCacheSet(ctx g.Ctx, addressLookupTable AddressLookupTable.AddressLookupTable) {
	_ = pool.addressLookupTableCache.Set(ctx, addressLookupTable.Address, addressLookupTable, addressLookupTableCacheDuration)

	return
}

func (pool *RPCs) AddressLookupTableCacheSets(ctx g.Ctx, addressLookupTables []AddressLookupTable.AddressLookupTable) {
	_ = pool.addressLookupTableCache.SetMap(ctx, func() (data map[any]any) {
		data = make(map[any]any, len(addressLookupTables))
		for i := range addressLookupTables {
			data[addressLookupTables[i].Address] = addressLookupTables[i]
		}

		return
	}(), addressLookupTableCacheDuration)

	return
}
