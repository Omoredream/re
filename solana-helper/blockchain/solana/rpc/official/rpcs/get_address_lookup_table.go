package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	ProgramLookup "github.com/gagliardetto/solana-go/programs/address-lookup-table"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address_lookup_table"
)

func (pool *RPCs) getAddressLookupTable(ctx g.Ctx, address Address.AccountAddress) (addressLookupTable AddressLookupTable.AddressLookupTable, err error) {
	accountInfo, err := pool.GetAccountInfo(ctx, address)
	if err != nil {
		err = gerror.Wrapf(err, "获取地址查找表失败")
		return
	}

	if accountInfo == nil {
		err = gerror.Newf("获取地址查找表为空")
		return
	}

	addressLookupTable_, err := ProgramLookup.DecodeAddressLookupTableState(accountInfo)
	if err != nil {
		err = gerror.Wrapf(err, "解析地址查找表失败")
		return
	}
	addressLookupTable = AddressLookupTable.AddressLookupTable{
		Address:            address,
		AddressLookupTable: addressLookupTable_.Addresses,
	}

	return
}

func (pool *RPCs) getAddressLookupTables(ctx g.Ctx, addresses []Address.AccountAddress) (addressLookupTables []AddressLookupTable.AddressLookupTable, err error) {
	if len(addresses) == 0 {
		return
	}
	addressLookupTables = make([]AddressLookupTable.AddressLookupTable, len(addresses))
	for i := 0; i < len(addresses); i += 100 {
		var accountsInfo [][]byte
		accountsInfo, err = pool.getAccountsInfo(ctx, addresses[i:min(i+100, len(addresses))])
		if err != nil {
			err = gerror.Wrapf(err, "批量获取地址查找表失败")
			return
		}

		for j := range accountsInfo {
			if accountsInfo[j] == nil {
				err = gerror.Newf("获取地址查找表为空")
				return
			}

			var addressLookupTable *ProgramLookup.AddressLookupTableState
			addressLookupTable, err = ProgramLookup.DecodeAddressLookupTableState(accountsInfo[j])
			if err != nil {
				err = gerror.Wrapf(err, "解析地址查找表失败")
				return
			}

			addressLookupTables[i+j] = AddressLookupTable.AddressLookupTable{
				Address:            addresses[i+j],
				AddressLookupTable: addressLookupTable.Addresses,
			}
		}
	}

	return
}
