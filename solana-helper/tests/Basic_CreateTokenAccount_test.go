package tests

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"
)

func TestCreateTokenAccount(t *testing.T) {
	err := testCreateTokenAccount(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testCreateTokenAccount(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	ixs := make([]solana.Instruction, 0)

	for _, addressBase58 := range []string{
		"AjQAzNbvJwoQCM8ztvAdjxhwLK4ZhWQ4JcyCV2GeLqUY",
	} {
		_, err = InstructionToken.CreateAssociatedTokenAccount{
			Funder: wallet.Account.Address,
			Owner:  wallet.Account.Address,
			Token:  Address.NewFromBase58(addressBase58).AsTokenAddress(),
		}.AppendIx(&ixs)
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}
	}

	txHash, err := officialPool.SendInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
	if err != nil {
		err = gerror.Wrapf(err, "发送交易失败")
		return
	}
	ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
	g.Log().Infof(ctx, "已发送交易")

	spent, err := officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
	if err != nil {
		err = gerror.Wrapf(err, "等待交易确认失败")
		return
	}
	g.Log().Infof(ctx, "交易耗时 %s", spent)

	return
}
