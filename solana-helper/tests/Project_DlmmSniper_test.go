package tests

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/compute_budget"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func TestDlmmSniper(t *testing.T) {
	err := testDlmmSniper(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testDlmmSniper(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	amountSOL := lamports.SOL2Lamports(decimal.NewFromFloat(1.25))

	// NYC
	//targetTokenAddress := "nTYWWx9pnyoMVehGxEbf5focKJQ8EntNMWQkTMVuMAP"
	//targetDlmmPoolAddress := "H2QpxgyunkpvvixDTXARyhgKfY5d2Ty5NsD7k81efRtm"
	//targetDlmmPoolReserveXAddress := "8p9ZJoRjN9A6LEFSFHjbbQYVC3ZiHgQMaGHHycYDcwTr"
	//targetDlmmPoolReserveYAddress := "GqsWSprS3kz8PYZYtGrXQgfZpo8HuvpZjEAY8VZ3JgVD"
	//targetDlmmPoolOracleAddress := "8Bn2DJtepqRHfS83grqRsZ4vat5xK7G2TZgB32CvQ6zF"
	//targetDlmmPoolBinArraysAddresses := []string{
	//	"2sLTqV7EcPncVtk1z4UUo78ZLv48at2MzCXJTK3DarnR",
	//}

	// SFO
	targetTokenAddress := "hNVhnH9WjtZcxXzJcgxcL1BRFcFeVAzFnkGxfpMgMAP"
	targetDlmmPoolAddress := "J2Z7xEgLGVM4dvyBprSaRrgKLf6UyKYBLEkzmnkPk7Zx"
	targetDlmmPoolReserveXAddress := "BDpu87BfdmwobYb4dVwcQjd4CCKnZWDsUtv211iDZ8JP"
	targetDlmmPoolReserveYAddress := "cWAEKaj2dc3WpZqvbS4cyLzaUzLNMsvkhAFNVgGpwAG"
	targetDlmmPoolOracleAddress := "F4wSLWgmv1o2v8wj6HKbw1cVcNQJyPtjtMEp2NwKTmCL"
	targetDlmmPoolBinArraysAddresses := []string{
		"EiZFyrXWXBfPtT1F37od2WYi5KiQFSk87ZEhDdazhN3i",
	}

	wallet := mainWallet
	targetToken := Address.NewFromBase58(targetTokenAddress).AsTokenAddress()
	targetDlmmPool := Address.NewFromBase58(targetDlmmPoolAddress)
	targetDlmmPoolReserveX := Address.NewFromBase58(targetDlmmPoolReserveXAddress)
	targetDlmmPoolReserveY := Address.NewFromBase58(targetDlmmPoolReserveYAddress)
	targetDlmmPoolOracle := Address.NewFromBase58(targetDlmmPoolOracleAddress)
	targetDlmmPoolBinArrays := Address.NewsFromBase58(targetDlmmPoolBinArraysAddresses)

	wsolTokenAccount, err := wallet.Account.Address.FindAssociatedTokenAccountAddress(consts.SOL.Address)
	if err != nil {
		g.Log().Fatalf(ctx, "生成代币账户失败, %+v", err)
	}

	ixs := make([]solana.Instruction, 0)
	cuLimit := InstructionComputeBudget.SetLimit{
		Limit: 150,
	}
	err = cuLimit.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	tokenAccount, err := InstructionToken.CreateAssociatedTokenAccount{
		Funder: wallet.Account.Address,
		Owner:  wallet.Account.Address,
		Token:  targetToken,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}
	cuLimit.Limit += 3_0000

	err = Instruction.Custom{
		ProgramID: consts.MeteoraDLMMProgramAddress,
		Accounts: Utils.Append([]*solana.AccountMeta{
			targetDlmmPool.Meta().WRITE(),                                                // Lb Pair
			consts.MeteoraDLMMProgramAddress.Meta(),                                      // Bin Array Bitmap Extension
			targetDlmmPoolReserveX.Meta().WRITE(),                                        // Reserve X
			targetDlmmPoolReserveY.Meta().WRITE(),                                        // Reserve Y
			wsolTokenAccount.Meta().WRITE(),                                              // User Token In
			tokenAccount.Address.Meta().WRITE(),                                          // User Token Out
			targetToken.Meta(),                                                           // Token X Mint
			consts.SOL.Address.Meta(),                                                    // Token Y Mint
			targetDlmmPoolOracle.Meta().WRITE(),                                          // Oracle
			consts.MeteoraDLMMProgramAddress.Meta(),                                      // Host Fee In
			wallet.Account.Address.Meta().SIGNER().WRITE(),                               // User
			consts.TokenProgramAddress.Meta(),                                            // Token X Program
			consts.TokenProgramAddress.Meta(),                                            // Token Y Program
			Address.NewFromBase58("D1ZN9Wj1fRSUQfCjhvnu1hqDMT7hzjzBBpi12nVniYD6").Meta(), // Event Authority
			consts.MeteoraDLMMProgramAddress.Meta(),                                      // Program
		}, lo.Map(targetDlmmPoolBinArrays, func(targetDlmmPoolBinArray Address.AccountAddress, _ int) *solana.AccountMeta { // binArrays
			return targetDlmmPoolBinArray.Meta().WRITE()
		})),
		Discriminator: []byte{0xf8, 0xc6, 0x9e, 0x91, 0xe1, 0x75, 0x87, 0xc8}, // Swap
		Data: Utils.Append(
			Utils.Uint64ToBytesL(amountSOL), // amountIn
			Utils.Uint64ToBytesL(0),         // minAmountOut
		),
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}
	cuLimit.Limit += 30_0000

	tipAccount, err := jitoPool.GetRandomTipAccount(ctx)
	if err != nil {
		err = gerror.Wrapf(err, "随机选择小费账户失败")
		return
	}

	err = Instruction.Transfer{
		Sender:   wallet,
		Receiver: tipAccount,
		Amount:   decimal.NewFromFloat(0.1),
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}
	cuLimit.Limit += 150

	ixs[0], err = cuLimit.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	for {
		var tx *solana.Transaction
		tx, err = officialPool.SignInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
		if err != nil {
			err = gerror.Wrapf(err, "签名交易失败")
			return
		}
		ctx := context.WithValue(ctx, consts.CtxTransaction, tx.Signatures[0].String())

		var bundleId string
		if grand.MeetProb(0.5) {
			var bundleId_ *string
			_, bundleId_, err = jitoPool.SendTransaction(ctx, tx, true)
			if err != nil {
				err = gerror.Wrapf(err, "广播捆绑交易失败")
				g.Log().Warningf(ctx, "%v", err)
				continue
			}
			bundleId = *bundleId_
		} else {
			bundleId, err = jitoPool.SendBundle(ctx, tx)
			if err != nil {
				err = gerror.Wrapf(err, "广播捆绑交易失败")
				g.Log().Warningf(ctx, "%v", err)
				continue
			}
		}
		ctx = context.WithValue(ctx, consts.CtxBundle, bundleId)
		g.Log().Infof(ctx, "已发送交易")

		time.Sleep(100 * time.Millisecond)
	}

	return
}
