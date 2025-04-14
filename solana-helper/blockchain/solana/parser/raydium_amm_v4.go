package Parser

import (
	"github.com/gagliardetto/binary"
	"github.com/gogf/gf/v2/errors/gerror"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/lp"
)

func ParseRaydiumAmmV4Pool(data []byte) (result LP.RaydiumAmmV4Pool, err error) {
	err = bin.NewBinDecoder(data).Decode(&result)
	if err != nil {
		err = gerror.Wrapf(err, "解析 Raydium AMM V4 Pool 失败")
		return
	}

	return
}
