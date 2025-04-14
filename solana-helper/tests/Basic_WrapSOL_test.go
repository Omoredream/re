package tests

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/account"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/compute_budget"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func TestWrapSOL(t *testing.T) {
	err := testWrapSOL(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testWrapSOL(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	wrapAmount := mainWallet.Account.SOL.
		Sub(decimal.NewFromFloat(1)). // 留 1 SOL
		Add(mainWallet.Account.Tokens[consts.SOL.Address].Token).
		Sub(lamports.Lamports2SOL(5_000))

	ixs := make([]solana.Instruction, 0)
	cuLimit := InstructionComputeBudget.SetLimit{
		Limit: 150,
	}
	err = cuLimit.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	walletTokenAccount, err := officialPool.GetAccountToken(ctx, wallet.Account.Address, consts.SOL.Address)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包代币账户失败")
		return
	}
	if walletTokenAccount != nil {
		err = InstructionToken.CloseAccount{
			TokenAccount: walletTokenAccount.Address,
			Owner:        wallet.Account.Address,
			Beneficiary:  wallet.Account.Address,
		}.AppendIx(&ixs)
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}
		cuLimit.Limit += 2916
		walletTokenAccount = nil
	}
	if walletTokenAccount == nil {
		var tokenAccount Account.TokenAccount
		tokenAccount, err = InstructionToken.CreateAssociatedTokenAccount{
			Funder: wallet.Account.Address,
			Owner:  wallet.Account.Address,
			Token:  consts.SOL.Address,
		}.AppendIx(&ixs)
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}
		cuLimit.Limit += 30000
		walletTokenAccount = &tokenAccount
	}

	err = Instruction.Transfer{
		Sender:   wallet,
		Receiver: walletTokenAccount.Address.AccountAddress,
		Amount:   wrapAmount,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}
	cuLimit.Limit += 150

	err = Instruction.SyncSOL{
		TokenAccount: walletTokenAccount.Address,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}
	cuLimit.Limit += 3045

	ixs[0], err = cuLimit.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	tx, err := officialPool.SignInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
	if err != nil {
		err = gerror.Wrapf(err, "构造交易失败")
		return
	}

	txRaw, err := utils.SerializeTransactionBase64(ctx, tx, false)
	if err != nil {
		err = gerror.Wrapf(err, "编码交易失败")
		return
	}
	g.Log().Infof(ctx, "构造交易 %s", txRaw)

	_, _, err = officialPool.SimulateTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "模拟交易失败")
		return
	}

	txHash, err := officialPool.SendTransaction(ctx, tx)
	if err != nil {
		err = gerror.Wrapf(err, "发送交易失败")
		return
	}
	ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
	g.Log().Infof(ctx, "已发送交易")

	spent, err := officialPool.WaitConfirmTransactionByHTTP(ctx, txHash, tx)
	if err != nil {
		err = gerror.Wrapf(err, "等待交易确认失败")
		return
	}
	g.Log().Infof(ctx, "交易耗时 %s", spent)

	return
}

func TestUnwrapSOL(t *testing.T) {
	err := testUnwrapSOL(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testUnwrapSOL(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	ixs := make([]solana.Instruction, 0)
	cuLimit := InstructionComputeBudget.SetLimit{
		Limit: 150,
	}
	err = cuLimit.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	walletTokenAccount, err := officialPool.GetAccountToken(ctx, wallet.Account.Address, consts.SOL.Address)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包代币账户失败")
		return
	}
	if walletTokenAccount == nil {
		g.Log().Infof(ctx, "无需重复 Unwrap")
		return
	}

	err = InstructionToken.CloseAccount{
		TokenAccount: walletTokenAccount.Address,
		Owner:        wallet.Account.Address,
		Beneficiary:  wallet.Account.Address,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}
	cuLimit.Limit += 2916

	ixs[0], err = cuLimit.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
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
