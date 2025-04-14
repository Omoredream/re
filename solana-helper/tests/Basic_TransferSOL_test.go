package tests

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func TestSimpleTransferSOL(t *testing.T) {
	err := testSimpleTransferSOL(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testSimpleTransferSOL(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	ixs := make([]solana.Instruction, 0)

	err = Instruction.Transfer{
		Sender:   wallet,
		Receiver: Address.NewFromBase58("777ZhCSUmKjcT1ErZ3J1do8rL67xoKqNgyBezXFvn777"),
		Amount:   wallet.Account.SOL.Sub(consts.SignFee),
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	var txHash solana.Signature
	txHash, err = officialPool.SendInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
	if err != nil {
		err = gerror.Wrapf(err, "发送交易失败")
		return
	}
	ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
	g.Log().Infof(ctx, "已发送交易")

	var spent time.Duration
	spent, err = officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
	if err != nil {
		err = gerror.Wrapf(err, "等待交易确认失败")
		return
	}
	g.Log().Infof(ctx, "交易耗时 %s", spent)

	return
}

func TestTransferSOL(t *testing.T) {
	err := testTransferSOL(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testTransferSOL(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	var walletCount uint32 = 500
	afterSOL := lamports.Lamports2SOL(26_711_760) // 0.00228288+0.0014616+0.01561672+0.0028536+0.00203928+0.00144768+0.001+0.00001

	const ixPerTx = 21

	for i := uint32(0); i < walletCount; {
		var fleetWallets []Wallet.HostedWallet

		fleetWallets = Utils.Grow(fleetWallets, len(fleetWallets)+ixPerTx)
		for ; i < walletCount && len(fleetWallets) < ixPerTx; i++ {
			ctx := context.WithValue(ctx, consts.CtxDerivation, i)

			var fleetWallet Wallet.HostedWallet
			fleetWallet, err = officialPool.NewWalletFromMnemonic(ctx, testWalletMnemonic, i)
			if err != nil {
				err = gerror.Wrapf(err, "派生钱包失败")
				return
			}

			ctx = context.WithValue(ctx, consts.CtxAddress, fleetWallet.Account.Address.String())

			g.Log().Infof(ctx, "SOL %s", decimals.DisplayBalance(fleetWallet.Account.SOL))

			if fleetWallet.Account.SOL.GreaterThanOrEqual(afterSOL) {
				continue
			}

			fleetWallets = append(fleetWallets, fleetWallet)
		}

		ixs := make([]solana.Instruction, 0)

		ixs, err = func() (ixs Instruction.Transfers) {
			for _, fleetWallet := range fleetWallets {
				ixs = append(ixs, Instruction.Transfer{
					Sender:   wallet,
					Receiver: fleetWallet.Account.Address,
					Amount:   afterSOL.Sub(fleetWallet.Account.SOL),
				})
			}
			return
		}().ToIxs()
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}
		if len(ixs) == 0 {
			continue
		}

		var txHash solana.Signature
		txHash, err = officialPool.SendInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
		if err != nil {
			err = gerror.Wrapf(err, "发送交易失败")
			return
		}
		ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
		g.Log().Infof(ctx, "已发送交易")

		var spent time.Duration
		spent, err = officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
		if err != nil {
			err = gerror.Wrapf(err, "等待交易确认失败")
			return
		}
		g.Log().Infof(ctx, "交易耗时 %s", spent)
	}

	return
}

func TestGatherSOL(t *testing.T) {
	err := testGatherSOL(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testGatherSOL(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	var walletCount uint32 = 500
	afterSOL := lamports.Lamports2SOL(0)

	const ixPerTx = 10

	for i := uint32(0); i < walletCount; {
		var fleetWallets []Wallet.HostedWallet

		fleetWallets = Utils.Grow(fleetWallets, len(fleetWallets)+ixPerTx)
		for ; i < walletCount && len(fleetWallets) < ixPerTx; i++ {
			ctx := context.WithValue(ctx, consts.CtxDerivation, i)

			var fleetWallet Wallet.HostedWallet
			fleetWallet, err = officialPool.NewWalletFromMnemonic(ctx, testWalletMnemonic, i)
			if err != nil {
				err = gerror.Wrapf(err, "派生钱包失败")
				return
			}

			ctx = context.WithValue(ctx, consts.CtxAddress, fleetWallet.Account.Address.String())

			g.Log().Infof(ctx, "SOL %s", decimals.DisplayBalance(fleetWallet.Account.SOL))

			if fleetWallet.Account.SOL.LessThanOrEqual(afterSOL) {
				continue
			}

			fleetWallets = append(fleetWallets, fleetWallet)
		}

		ixs := make([]solana.Instruction, 0)

		ixs, err = func() (ixs Instruction.Transfers) {
			for _, fleetWallet := range fleetWallets {
				ixs = append(ixs, Instruction.Transfer{
					Sender:   fleetWallet,
					Receiver: wallet.Account.Address,
					Amount:   fleetWallet.Account.SOL.Sub(afterSOL),
				})
			}
			if len(ixs) > 0 {
				ixs[0].Amount = ixs[0].Amount.Sub(lamports.Lamports2SOL(5000).Mul(decimal.NewFromInt(int64(len(ixs))))) // 每个签名 5000 lamports
				if ixs[0].Amount.IsZero() {
					ixs = ixs[1:]
				} else if ixs[0].Amount.IsNegative() {
					g.Log().Fatalf(ctx, "用于签名的钱包余额不足")
				}
			}
			return
		}().ToIxs()
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}
		if len(ixs) == 0 {
			continue
		}

		var txHash solana.Signature
		txHash, err = officialPool.SendInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
		if err != nil {
			err = gerror.Wrapf(err, "发送交易失败")
			return
		}
		ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
		g.Log().Infof(ctx, "已发送交易")

		var spent time.Duration
		spent, err = officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
		if err != nil {
			err = gerror.Wrapf(err, "等待交易确认失败")
			return
		}
		g.Log().Infof(ctx, "交易耗时 %s", spent)
	}

	return
}
