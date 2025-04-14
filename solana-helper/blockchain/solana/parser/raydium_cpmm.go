package Parser

import (
	"github.com/gagliardetto/binary"
	"github.com/gogf/gf/v2/errors/gerror"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/lp"
)

func ParseRaydiumCpmmPool(data []byte) (result LP.RaydiumCpmmPool, err error) {
	err = bin.NewBorshDecoder(data).Decode(&result)
	if err != nil {
		err = gerror.Wrapf(err, "解析 Raydium CPMM Pool 失败")
		return
	}

	return
}
