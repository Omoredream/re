package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"
)

func (pool *RPCs) SignInstructions(ctx g.Ctx, ixs []solana.Instruction, signers []Wallet.HostedWallet, feePayer Wallet.HostedWallet) (tx *solana.Transaction, err error) {
	if len(ixs) == 0 {
		err = gerror.Newf("交易指令为空")
		return
	}

	blockhash, err := pool.GetLatestBlockhash(ctx)
	if err != nil {
		err = gerror.Wrapf(err, "获取最新区块失败")
		return
	}

	tx, err = pool.PackTransaction(ctx, ixs, blockhash, feePayer.Account.Address, nil)
	if err != nil {
		err = gerror.Wrapf(err, "构造交易失败")
		return
	}

	err = pool.SignTransaction(ctx, tx, signers)
	if err != nil {
		err = gerror.Wrapf(err, "签名交易失败")
		return
	}

	return
}
