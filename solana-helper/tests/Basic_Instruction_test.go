package tests

import (
	"context"
	"maps"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/account"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/compute_budget"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func TestIxReorganize(t *testing.T) {
	err := testIxReorganize(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testIxReorganize(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	ixs := make([]solana.Instruction, 0)
	alts := make(map[solana.PublicKey]solana.PublicKeySlice)

	for _, txRawBase64 := range []string{
		"这里",
	} {
		var tx *solana.Transaction
		tx, err = utils.DeserializeTransactionBase64(txRawBase64)
		if err != nil {
			err = gerror.Wrapf(err, "解析交易失败")
			return
		}

		var ixs_ []solana.Instruction
		var alt map[solana.PublicKey]solana.PublicKeySlice
		ixs_, _, _, alt, err = officialPool.UnpackTransaction(ctx, tx)
		if err != nil {
			err = gerror.Wrapf(err, "拆解交易失败")
			return
		}

		for _, ix := range ixs_ {
			if ix.ProgramID() == consts.ComputeBudgetProgramAddress.PublicKey {
				continue
			}
			ixs = append(ixs, ix)
		}

		maps.Copy(alts, alt)
	}

	err = InstructionToken.Burns{
		{
			Amount:       wallet.Account.Tokens[Address.NewFromBase58("2PLc1Cbm7wJo8vKon54wGtxb1QcoyFnTut86BYpbdHKk").AsTokenAddress()].Token.Sub(decimal.NewFromFloat(0)),
			TokenAccount: wallet.Account.Tokens[Address.NewFromBase58("2PLc1Cbm7wJo8vKon54wGtxb1QcoyFnTut86BYpbdHKk").AsTokenAddress()].Address,
			Token:        lo.Must(officialPool.TokenCacheGet(ctx, Address.NewFromBase58("2PLc1Cbm7wJo8vKon54wGtxb1QcoyFnTut86BYpbdHKk").AsTokenAddress())),
			Owner:        wallet.Account.Address,
		},
	}.AppendIxs(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	err = InstructionToken.CloseAccounts{
		{
			TokenAccount: wallet.Account.Tokens[Address.NewFromBase58("2PLc1Cbm7wJo8vKon54wGtxb1QcoyFnTut86BYpbdHKk").AsTokenAddress()].Address,
			Owner:        wallet.Account.Address,
			Beneficiary:  wallet.Account.Address,
		},
	}.AppendIxs(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	err = InstructionComputeBudget.SetLimit{
		Limit: 12_0000 + 4659*1 + 2916*1 + 150,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	tx, err := officialPool.PackTransaction(ctx, ixs, solana.Hash{}, wallet.Account.Address, alts)
	if err != nil {
		err = gerror.Wrapf(err, "构造交易失败")
		return
	}

	err = officialPool.SignTransaction(ctx, tx, []Wallet.HostedWallet{wallet})
	if err != nil {
		err = gerror.Wrapf(err, "签名交易失败")
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

func TestIxCustom(t *testing.T) {
	err := testIxCustom(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testIxCustom(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	ixs := make([]solana.Instruction, 0)

	wsolTokenAccount, err := officialPool.GetAccountToken(ctx, wallet.Account.Address, consts.SOL.Address)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包代币账户失败")
		return
	}
	if wsolTokenAccount == nil {
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
		wsolTokenAccount = &tokenAccount

		err = Instruction.Transfer{
			Sender:   wallet,
			Receiver: wsolTokenAccount.Address.AccountAddress,
			Amount:   decimal.NewFromFloat(1),
		}.AppendIx(&ixs)
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}

		err = Instruction.SyncSOL{
			TokenAccount: wsolTokenAccount.Address,
		}.AppendIx(&ixs)
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}
	}

	mint1Token := Address.NewFromBase58("CGahKWnAp4yP4n9YT3DjuPumPAAeW7QUXX62SREgG2PT").AsTokenAddress()

	mint1TokenAccount, err := officialPool.GetAccountToken(ctx, wallet.Account.Address, mint1Token)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包代币账户失败")
		return
	}
	if mint1TokenAccount == nil {
		var tokenAccount Account.TokenAccount
		tokenAccount, err = InstructionToken.CreateAssociatedTokenAccount{
			Funder: wallet.Account.Address,
			Owner:  wallet.Account.Address,
			Token:  mint1Token,
		}.AppendIx(&ixs)
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}
		mint1TokenAccount = &tokenAccount
	}

	mint2Token := Address.NewFromBase58("AknMqhKAdwr5tXUev54d4tK2VvBNFKAkFneqQkty25nW").AsTokenAddress()

	mint2TokenAccount, err := officialPool.GetAccountToken(ctx, wallet.Account.Address, mint2Token)
	if err != nil {
		err = gerror.Wrapf(err, "获取钱包代币账户失败")
		return
	}
	if mint2TokenAccount == nil {
		var tokenAccount Account.TokenAccount
		tokenAccount, err = InstructionToken.CreateAssociatedTokenAccount{
			Funder: wallet.Account.Address,
			Owner:  wallet.Account.Address,
			Token:  mint2Token,
		}.AppendIx(&ixs)
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}
		mint2TokenAccount = &tokenAccount
	}

	arbProgram := Address.NewFromBase58("MoneyymapoTpHK5zNmo877RwgNN74Wx7r6bS3aS7Buq").AsProgramAddress()

	arbEventAuthority, _, err := arbProgram.FindProgramDerivedAddress([][]byte{
		[]byte("__event_authority"),
	})
	if err != nil {
		err = gerror.Wrapf(err, "派生 Arb 日志鉴权账户失败")
		return
	}

	raydiumAmmV4Program := Address.NewFromBase58("HWy1jotHpo6UqeQxx49dpYYdQB8wj9Qk9MdxwjLvDHB8").AsProgramAddress()

	raydiumAmmV4Authority, err := raydiumAmmV4Program.CreateProgramDerivedAddress([][]byte{
		[]byte("amm authority"),
	}, 252)
	if err != nil {
		err = gerror.Wrapf(err, "派生 Raydium AMM V4 鉴权账户失败")
		return
	}

	raydiumCpmmProgram := Address.NewFromBase58("CPMDWBwJDtYax9qW7AyRuVC19Cc4L4Vcy4n2BHAbHkCW").AsProgramAddress()

	raydiumCpmmAuthority, _, err := raydiumCpmmProgram.FindProgramDerivedAddress([][]byte{
		[]byte("vault_and_lp_mint_auth_seed"),
	})
	if err != nil {
		err = gerror.Wrapf(err, "派生 Raydium CPMM 鉴权账户失败")
		return
	}

	err = Instruction.Custom{
		ProgramID: arbProgram,
		Accounts: solana.AccountMetaSlice{
			arbEventAuthority.Meta(), // event_authority
			arbProgram.Meta(),        // program
			// Raydium AMM V4
			raydiumAmmV4Program.Meta(),
			consts.TokenProgramAddress.Meta(), // token_program
			Address.NewFromBase58("2FnyFeEwG1wXNBFKgc6RrN9YZk63hGKW5E1Mh8mXBiFo").Meta().WRITE(), // amm
			raydiumAmmV4Authority.Meta(), // amm_authority
			Address.NewFromBase58("3zcZ4GtZJHS2MXUVEFWzp3ZGHGQviqPPNHH2r74TFhU5").Meta().WRITE(), // amm_coin_vault
			Address.NewFromBase58("HqFHPeJJ6om4fZBKRHXiE2qwmytecdDGtC6fL7tHW3oQ").Meta().WRITE(), // amm_pc_vault
			wsolTokenAccount.Address.Meta().WRITE(),                                              // user_token_source
			mint1TokenAccount.Address.Meta().WRITE(),                                             // user_token_destination
			wallet.Account.Address.Meta().SIGNER().WRITE(),                                       // user_source_owner
			// Raydium AMM V4
			raydiumAmmV4Program.Meta(),
			consts.TokenProgramAddress.Meta(), // token_program
			Address.NewFromBase58("2FnyFeEwG1wXNBFKgc6RrN9YZk63hGKW5E1Mh8mXBiFo").Meta().WRITE(), // amm
			raydiumAmmV4Authority.Meta(), // amm_authority
			Address.NewFromBase58("3zcZ4GtZJHS2MXUVEFWzp3ZGHGQviqPPNHH2r74TFhU5").Meta().WRITE(), // amm_coin_vault
			Address.NewFromBase58("HqFHPeJJ6om4fZBKRHXiE2qwmytecdDGtC6fL7tHW3oQ").Meta().WRITE(), // amm_pc_vault
			mint1TokenAccount.Address.Meta().WRITE(),                                             // user_token_source
			wsolTokenAccount.Address.Meta().WRITE(),                                              // user_token_destination
			wallet.Account.Address.Meta().SIGNER().WRITE(),                                       // user_source_owner
			// Raydium CPMM
			raydiumCpmmProgram.Meta(),
			wallet.Account.Address.Meta().SIGNER().WRITE(),                                       // payer
			raydiumCpmmAuthority.Meta(),                                                          // authority
			Address.NewFromBase58("9zSzfkYy6awexsHvmggeH36pfVUdDGyCcwmjT3AQPBj6").Meta(),         // amm_config
			Address.NewFromBase58("8mfKzG7kGnorSPGE1mxk1qGebqKpZ6ZQmBjDD6fU1sqn").Meta().WRITE(), // pool_state
			wsolTokenAccount.Address.Meta().WRITE(),                                              // input_token_account
			mint2TokenAccount.Address.Meta().WRITE(),                                             // output_token_account
			Address.NewFromBase58("AdhpaT5BN5R5uxMnNjzzdf3jC9WYxRPAczTEQpwdVeuR").Meta().WRITE(), // input_vault
			Address.NewFromBase58("D8DkgUuxDsmfPz5rMi6dsG4v68seZAFFfG5fKsKyKKaj").Meta().WRITE(), // output_vault
			consts.TokenProgramAddress.Meta(),                                                    // input_token_program
			consts.TokenProgramAddress.Meta(),                                                    // output_token_program
			consts.SOL.Address.Meta(),                                                            // input_token_mint
			mint2Token.Meta(),                                                                    // output_token_mint
			Address.NewFromBase58("CTE88ArKnp9kQ9kgdxAbfDtdAwt37zpKr3MG4CSKvgYk").Meta().WRITE(), // observation_state
			// Raydium CPMM
			raydiumCpmmProgram.Meta(),
			wallet.Account.Address.Meta().SIGNER().WRITE(),                                       // payer
			raydiumCpmmAuthority.Meta(),                                                          // authority
			Address.NewFromBase58("9zSzfkYy6awexsHvmggeH36pfVUdDGyCcwmjT3AQPBj6").Meta(),         // amm_config
			Address.NewFromBase58("8mfKzG7kGnorSPGE1mxk1qGebqKpZ6ZQmBjDD6fU1sqn").Meta().WRITE(), // pool_state
			mint2TokenAccount.Address.Meta().WRITE(),                                             // input_token_account
			wsolTokenAccount.Address.Meta().WRITE(),                                              // output_token_account
			Address.NewFromBase58("D8DkgUuxDsmfPz5rMi6dsG4v68seZAFFfG5fKsKyKKaj").Meta().WRITE(), // input_vault
			Address.NewFromBase58("AdhpaT5BN5R5uxMnNjzzdf3jC9WYxRPAczTEQpwdVeuR").Meta().WRITE(), // output_vault
			consts.TokenProgramAddress.Meta(),                                                    // input_token_program
			consts.TokenProgramAddress.Meta(),                                                    // output_token_program
			mint2Token.Meta(),                                                                    // input_token_mint
			consts.SOL.Address.Meta(),                                                            // output_token_mint
			Address.NewFromBase58("CTE88ArKnp9kQ9kgdxAbfDtdAwt37zpKr3MG4CSKvgYk").Meta().WRITE(), // observation_state
		},
		Discriminator: arbProgram.GetDiscriminator("global", "swap"),
		Data: Utils.Append(
			Utils.Uint64ToBytesL(lamports.SOL2Lamports(decimal.NewFromFloat(0.1))), // amount_in
			Utils.VecTToBytes(Utils.ByteArrayToBytes, // routes
				Utils.Append(
					Utils.EnumToBytes(1, nil), // dex
					Utils.Uint8ToBytesL(0),    // input_index
					Utils.Uint8ToBytesL(1),    // output_index
				),
				Utils.Append(
					Utils.EnumToBytes(1, nil), // dex
					Utils.Uint8ToBytesL(1),    // input_index
					Utils.Uint8ToBytesL(0),    // output_index
				),
				Utils.Append(
					Utils.EnumToBytes(2, nil), // dex
					Utils.Uint8ToBytesL(0),    // input_index
					Utils.Uint8ToBytesL(1),    // output_index
				),
				Utils.Append(
					Utils.EnumToBytes(2, nil), // dex
					Utils.Uint8ToBytesL(1),    // input_index
					Utils.Uint8ToBytesL(0),    // output_index
				),
			),
		),
	}.AppendIx(&ixs)
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

	spent, err := officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
	if err != nil {
		err = gerror.Wrapf(err, "等待交易确认失败")
		return
	}
	g.Log().Infof(ctx, "交易耗时 %s", spent)

	return
}
