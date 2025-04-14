package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"
)

func (pool *RPCs) SendInstructions(ctx g.Ctx, ixs []solana.Instruction, signers []Wallet.HostedWallet, feePayer Wallet.HostedWallet) (txHash solana.Signature, err error) {
	tx, err := pool.SignInstructions(ctx, ixs, signers, feePayer)
	if err != nil {
		err = gerror.Wrapf(err, "构造交易失败")
		return
	}

	txHash, err = pool.SendTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "完成交易失败")
		return
	}

	return
}
