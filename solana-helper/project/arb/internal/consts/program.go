package consts

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

var (
	ArbProgramAddress = Address.NewFromBase58("MoneyymapoTpHK5zNmo877RwgNN74Wx7r6bS3aS7Buq").AsProgramAddress()
	ArbEventAuthority Address.AccountAddress
)

func init() {
	var err error
	defer func() {
		if err != nil {
			g.Log("init").Errorf(context.Background(), "%+v", err)
		}
	}()

	ArbEventAuthority, _, err = ArbProgramAddress.FindProgramDerivedAddress([][]byte{
		[]byte("__event_authority"),
	})
	if err != nil {
		err = gerror.Wrapf(err, "派生 Arb 日志鉴权账户失败")
		return
	}
}
