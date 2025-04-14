package tests

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/kamino"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
)

func TestFlashLoan(t *testing.T) {
	err := testFlashLoan(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testFlashLoan(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	tokenAccountAddress, err := wallet.Account.Address.FindAssociatedTokenAccountAddress(consts.SOL.Address)
	//tokenAccountAddress, err := wallet.Account.Address.FindAssociatedTokenAccountAddress(consts.USDC.Address)
	if err != nil {
		err = gerror.Wrapf(err, "派生代币账户失败")
		return
	}

	ixs := make([]solana.Instruction, 0)

	err = InstructionKamino.FlashBorrow{
		Token:                    consts.SOL,
		User:                     wallet.Account.Address,
		UserTokenAccount:         tokenAccountAddress,
		Market:                   consts.KaminoMainMarket,
		MarketAuthority:          consts.KaminoMainMarketAuthority,
		MarketReserve:            consts.KaminoMainMarketSOLReserve,
		MarketReserveLiquidity:   consts.KaminoMainMarketSOLReserveLiquidity,
		MarketReserveFeeReceiver: consts.KaminoMainMarketSOLReserveFeeReceiver,
		Amount:                   decimal.NewFromFloat(100),
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	err = InstructionKamino.FlashRepay{
		Token:                    consts.SOL,
		User:                     wallet.Account.Address,
		UserTokenAccount:         tokenAccountAddress,
		Market:                   consts.KaminoMainMarket,
		MarketAuthority:          consts.KaminoMainMarketAuthority,
		MarketReserve:            consts.KaminoMainMarketSOLReserve,
		MarketReserveLiquidity:   consts.KaminoMainMarketSOLReserveLiquidity,
		MarketReserveFeeReceiver: consts.KaminoMainMarketSOLReserveFeeReceiver,
		Amount:                   decimal.NewFromFloat(100),
		BorrowIxIndex:            0,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	//err = InstructionKamino.FlashBorrow{
	//	Token:                    consts.SOL,
	//	User:                     wallet.Account.Address,
	//	UserTokenAccount:         tokenAccountAddress,
	//	Market:                   consts.KaminoJitoMarket,
	//	MarketAuthority:          consts.KaminoJitoMarketAuthority,
	//	MarketReserve:            consts.KaminoJitoMarketSOLReserve,
	//	MarketReserveLiquidity:   consts.KaminoJitoMarketSOLReserveLiquidity,
	//	MarketReserveFeeReceiver: consts.KaminoJitoMarketSOLReserveFeeReceiver,
	//	Amount:                   decimal.NewFromFloat(100),
	//}.AppendIx(&ixs)
	//if err != nil {
	//	err = gerror.Wrapf(err, "构建交易失败")
	//	return
	//}
	//
	//err = InstructionKamino.FlashRepay{
	//	Token:                    consts.SOL,
	//	User:                     wallet.Account.Address,
	//	UserTokenAccount:         tokenAccountAddress,
	//	Market:                   consts.KaminoJitoMarket,
	//	MarketAuthority:          consts.KaminoJitoMarketAuthority,
	//	MarketReserve:            consts.KaminoJitoMarketSOLReserve,
	//	MarketReserveLiquidity:   consts.KaminoJitoMarketSOLReserveLiquidity,
	//	MarketReserveFeeReceiver: consts.KaminoJitoMarketSOLReserveFeeReceiver,
	//	Amount:                   decimal.NewFromFloat(100),
	//	BorrowIxIndex:            0,
	//}.AppendIx(&ixs)
	//if err != nil {
	//	err = gerror.Wrapf(err, "构建交易失败")
	//	return
	//}

	//err = InstructionKamino.FlashBorrow{
	//	Token:                    consts.USDC,
	//	User:                     wallet.Account.Address,
	//	UserTokenAccount:         tokenAccountAddress,
	//	Market:                   consts.KaminoMainMarket,
	//	MarketAuthority:          consts.KaminoMainMarketAuthority,
	//	MarketReserve:            consts.KaminoMainMarketUSDCReserve,
	//	MarketReserveLiquidity:   consts.KaminoMainMarketUSDCReserveLiquidity,
	//	MarketReserveFeeReceiver: consts.KaminoMainMarketUSDCReserveFeeReceiver,
	//	Amount:                   decimal.NewFromFloat(10000),
	//}.AppendIx(&ixs)
	//if err != nil {
	//	err = gerror.Wrapf(err, "构建交易失败")
	//	return
	//}
	//
	//err = InstructionKamino.FlashRepay{
	//	Token:                    consts.USDC,
	//	User:                     wallet.Account.Address,
	//	UserTokenAccount:         tokenAccountAddress,
	//	Market:                   consts.KaminoMainMarket,
	//	MarketAuthority:          consts.KaminoMainMarketAuthority,
	//	MarketReserve:            consts.KaminoMainMarketUSDCReserve,
	//	MarketReserveLiquidity:   consts.KaminoMainMarketUSDCReserveLiquidity,
	//	MarketReserveFeeReceiver: consts.KaminoMainMarketUSDCReserveFeeReceiver,
	//	Amount:                   decimal.NewFromFloat(10000),
	//	BorrowIxIndex:            0,
	//}.AppendIx(&ixs)
	//if err != nil {
	//	err = gerror.Wrapf(err, "构建交易失败")
	//	return
	//}

	//err = InstructionKamino.FlashBorrow{
	//	Token:                    consts.USDC,
	//	User:                     wallet.Account.Address,
	//	UserTokenAccount:         tokenAccountAddress,
	//	Market:                   consts.KaminoJLPMarket,
	//	MarketAuthority:          consts.KaminoJLPMarketAuthority,
	//	MarketReserve:            consts.KaminoJLPMarketUSDCReserve,
	//	MarketReserveLiquidity:   consts.KaminoJLPMarketUSDCReserveLiquidity,
	//	MarketReserveFeeReceiver: consts.KaminoJLPMarketUSDCReserveFeeReceiver,
	//	Amount:                   decimal.NewFromFloat(10000),
	//}.AppendIx(&ixs)
	//if err != nil {
	//	err = gerror.Wrapf(err, "构建交易失败")
	//	return
	//}
	//
	//err = InstructionKamino.FlashRepay{
	//	Token:                    consts.USDC,
	//	User:                     wallet.Account.Address,
	//	UserTokenAccount:         tokenAccountAddress,
	//	Market:                   consts.KaminoJLPMarket,
	//	MarketAuthority:          consts.KaminoJLPMarketAuthority,
	//	MarketReserve:            consts.KaminoJLPMarketUSDCReserve,
	//	MarketReserveLiquidity:   consts.KaminoJLPMarketUSDCReserveLiquidity,
	//	MarketReserveFeeReceiver: consts.KaminoJLPMarketUSDCReserveFeeReceiver,
	//	Amount:                   decimal.NewFromFloat(10000),
	//	BorrowIxIndex:            0,
	//}.AppendIx(&ixs)
	//if err != nil {
	//	err = gerror.Wrapf(err, "构建交易失败")
	//	return
	//}

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
	return

	txHash, err := officialPool.SendTransaction(ctx, tx)
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
