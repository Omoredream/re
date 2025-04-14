package tests

import (
	"context"
	"crypto/ed25519"
	"slices"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/compute_budget"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"
)

func TestLaunchMyNFT(t *testing.T) {
	err := testLaunchMyNFT(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testLaunchMyNFT(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF, true)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	fleetWallets, err := officialPool.NewWalletsFromMnemonic(ctx, testWalletMnemonic, g.Cfg().MustGet(nil, "project.launchMyNFT.derivationFrom", 0).Uint32(), g.Cfg().MustGet(nil, "project.launchMyNFT.derivationTo", 50).Uint32(), true)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	var nftWallets = make([]Wallet.HostedWallet, len(fleetWallets))
	for i, _ := range fleetWallets {
		nftWallets[i], err = officialPool.NewWalletFromPrivateKey(ctx, ed25519.NewKeyFromSeed(grand.B(32)), true)
		if err != nil {
			err = gerror.Wrapf(err, "导入钱包失败")
			return
		}
	}

	interval := g.Cfg().MustGet(nil, "project.launchMyNFT.interval", 100*time.Millisecond).Duration()

	launchMyNFTProgram := Address.NewFromBase58("F9SixdqdmEBP5kprp2gZPZNeMmfHJRCTMFjN22dx3akf").AsProgramAddress()
	launchMyNFTMintCoreDiscriminator := launchMyNFTProgram.GetDiscriminator("global", "mint_core")

	machineAddress := Address.NewFromBase58("3uNU3oVskN23u2i76aPJ9tN75cK8FDQCShNdjHsCuV9R")
	collectionAddress := Address.NewFromBase58("57p7WnyUyeQRFe3Kx3WKPeXn1SU1sHhCKMrPKXxQc9oB")
	fundingAddress := Address.NewFromBase58("Cazm1kQ5JPppYmgLfvJag6hxkmrbViskhsRUCEFKcXf3")
	coreBankAddress, _, err := launchMyNFTProgram.FindProgramDerivedAddress([][]byte{
		[]byte("core_bank"),
		machineAddress.Bytes(),
	})
	if err != nil {
		err = gerror.Wrapf(err, "派生地址失败")
		return
	}

	for {
		for i, fleetWallet := range fleetWallets {
			nftWallet := nftWallets[i]
			ctx := context.WithValue(ctx, consts.CtxWallet, nftWallet.Account.Address.ShortString())

			// mint tx

			var totalMints Address.AccountAddress
			totalMints, _, err = launchMyNFTProgram.FindProgramDerivedAddress([][]byte{
				[]byte("TotalMints"),
				fleetWallet.Account.Address.Bytes(),
				machineAddress.Bytes(),
				slices.Collect(slices.Chunk([]byte("Public"), 32))[0],
			})
			if err != nil {
				err = gerror.Wrapf(err, "派生地址失败")
				return
			}

			ixs := make([]solana.Instruction, 0)

			err = Instruction.Transfer{
				Sender:   mainWallet,
				Receiver: fleetWallet.Account.Address,
				Amount:   decimal.NewFromFloat(0.01), // >= 0.0037272+0.0009744+0.00415 + 0.000890880 + 0.000005000 + 0.00011
			}.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}

			err = Instruction.Custom{
				ProgramID: launchMyNFTProgram,
				Accounts: solana.AccountMetaSlice{
					fleetWallet.Account.Address.Meta().SIGNER().WRITE(),                                  // buyer
					fleetWallet.Account.Address.Meta().SIGNER(),                                          // requiredSigner
					machineAddress.Meta().WRITE(),                                                        // machine
					totalMints.Meta().WRITE(),                                                            // totalMints
					fundingAddress.Meta().WRITE(),                                                        // wallet
					Address.NewFromBase58("33nQCgievSd3jJLSWFBefH3BJRN7h6sAoS82VFFdJGF5").Meta().WRITE(), // wallet2
					consts.SystemProgramAddress.Meta(),
					launchMyNFTProgram.Meta(), //.WRITE(), // buyerWlTokenWallet, optional
					launchMyNFTProgram.Meta(), // wlNftMetadata, optional
					launchMyNFTProgram.Meta(), //.WRITE(), // mintsPerWlNft, optional
					launchMyNFTProgram.Meta(), //.WRITE(), // whitelistMint, optional
					launchMyNFTProgram.Meta(), //.WRITE(), // buyerPaymentTokenWallet, optional
					consts.SysVarInstructionsAddress.Meta(),
					consts.TokenProgramAddress.Meta(),
					consts.Token2022ProgramAddress.Meta(),
					consts.TokenMetadataProgramAddress.Meta(),
					consts.StateCompressionProgramAddress.Meta(),
					consts.NoopProgramAddress.Meta(),
					consts.SysVarSlotHashesAddress.Meta(),
					consts.BubblegumProgramAddress.Meta(),
					launchMyNFTProgram.Meta(), // wlNftTree, optional
					launchMyNFTProgram.Meta(), //.WRITE(), // wlNftTreeAuthority, optional
					consts.ATAProgramAddress.Meta(),
					consts.MetaplexCoreProgramAddress.Meta(),
					collectionAddress.Meta().WRITE(),                  // collection
					nftWallet.Account.Address.Meta().SIGNER().WRITE(), // asset
					// remaining
					consts.Token2022ProgramAddress.Meta(),
					fundingAddress.Meta().WRITE(),
					fundingAddress.Meta().WRITE(),
					coreBankAddress.Meta().WRITE(),
				},
				Discriminator: launchMyNFTMintCoreDiscriminator,
				Data: Utils.Append(
					Utils.VecTToBytes[struct{}](nil),             // proof
					Utils.Uint64ToBytesL(0),                      // expectedPrice
					Utils.Uint32ToBytesL(1),                      // amount
					Utils.OptionTToBytes[struct{}](nil),          // cnftWlEntry
					Utils.OptionTToBytes(Utils.Uint8ToBytesL, 1), // targetPhase, Public round
				),
			}.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}

			err = InstructionComputeBudget.SetLimit{
				Limit: 150 + 17_5000 + 150,
			}.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}

			var blockhash solana.Hash
			blockhash, err = officialPool.GetLatestBlockhash(ctx)
			if err != nil {
				err = gerror.Wrapf(err, "获取最新区块失败")
				return
			}

			var mintTx *solana.Transaction
			mintTx, err = officialPool.PackTransaction(ctx, ixs, blockhash, mainWallet.Account.Address, nil)
			if err != nil {
				err = gerror.Wrapf(err, "构造交易失败")
				return
			}

			err = officialPool.SignTransaction(ctx, mintTx, []Wallet.HostedWallet{mainWallet, fleetWallet, nftWallet})
			if err != nil {
				err = gerror.Wrapf(err, "签名交易失败")
				return
			}

			// tip tx

			var jitoTipAccount Address.AccountAddress
			jitoTipAccount, err = jitoPool.GetRandomTipAccount(ctx)
			if err != nil {
				err = gerror.Wrapf(err, "获取 Jito 小费账户失败")
				return
			}

			ixs = make([]solana.Instruction, 0)

			err = Instruction.Transfer{
				Sender:   fleetWallet,
				Receiver: jitoTipAccount,
				Amount:   decimal.NewFromFloat(0.000_11),
			}.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}

			err = InstructionComputeBudget.SetLimit{
				Limit: 150 + 150,
			}.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}

			var tipTx *solana.Transaction
			tipTx, err = officialPool.PackTransaction(ctx, ixs, blockhash, fleetWallet.Account.Address, nil)
			if err != nil {
				err = gerror.Wrapf(err, "构造交易失败")
				return
			}

			err = officialPool.SignTransaction(ctx, tipTx, []Wallet.HostedWallet{fleetWallet})
			if err != nil {
				err = gerror.Wrapf(err, "签名交易失败")
				return
			}

			var bundleId string
			bundleId, err = jitoPool.SendBundle(ctx, mintTx, tipTx)
			if err != nil {
				err = gerror.Wrapf(err, "广播捆绑交易失败")
				g.Log().Warningf(ctx, "%v", err)
				err = nil
				continue
			}
			ctx = context.WithValue(ctx, consts.CtxBundle, bundleId)
			g.Log().Noticef(ctx, "已发送交易")
			time.Sleep(interval)
		}
	}
}
