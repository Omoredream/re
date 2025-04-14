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
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"
)

func TestTransferToken(t *testing.T) {
	err := testTransferToken(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testTransferToken(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	payTo := Address.NewFromBase58("6tGotiReGQRXrXRaB6MaNyww9GFaDJqhkWuS8kFdCtFP")

	tokenAddress := Address.NewFromBase58("ukHH6c7mMyiWCf1b9pnWe25TSpkDDt3H5pQZgZ74J82").AsTokenAddress()

	transferAmount := decimal.NewFromFloat(0.000_01)

	ixs := make([]solana.Instruction, 0)

	walletTokenAccount, err := officialPool.GetAccountToken(ctx, wallet.Account.Address, tokenAddress)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包代币账户失败")
		return
	}
	if walletTokenAccount == nil {
		err = gerror.Newf("无法找到发送者的代币账户")
		err = gerror.Wrapf(err, "获取钱包代币账户失败")
		return
	}

	payToTokenAccount, err := officialPool.GetAccountToken(ctx, payTo, tokenAddress)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包代币账户失败")
		return
	}
	if payToTokenAccount == nil {
		var tokenAccount Account.TokenAccount
		tokenAccount, err = InstructionToken.CreateAssociatedTokenAccount{
			Funder: wallet.Account.Address,
			Owner:  payTo,
			Token:  tokenAddress,
		}.AppendIx(&ixs)
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}
		payToTokenAccount = &tokenAccount
	}

	token, err := officialPool.TokenCacheGet(ctx, tokenAddress)
	if err != nil {
		err = gerror.Wrapf(err, "获取代币失败")
		return
	}

	err = InstructionToken.Transfer{
		Owner:    wallet,
		Sender:   walletTokenAccount.Address,
		Receiver: payToTokenAccount.Address,
		Amount:   transferAmount,
		Token:    token,
	}.AppendIx(&ixs)
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
