package ProjectArb

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"
	ProgramComputeBudget "github.com/gagliardetto/solana-go/programs/compute-budget"

	"git.wkr.moe/web3/solana-helper/errcode"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/compute_budget"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/kamino"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"

	interConsts "git.wkr.moe/web3/solana-helper/project/arb/internal/consts"

	interInstruction "git.wkr.moe/web3/solana-helper/project/arb/internal/instruction"
)

func (arb *Arb) refactorTxs(
	ctx context.Context,
	swapTx *solana.Transaction,
	tradeSize decimal.Decimal, profitAsSOL decimal.Decimal, foundSlot uint64,
) (
	spamTx *solana.Transaction,
	jitoTxs []*solana.Transaction,
	err error,
) {
	ixs, blockhash, _, addressLookupTables, err := arb.officialPool.UnpackTransaction(ctx, swapTx)
	if err != nil {
		err = gerror.Wrapf(err, "拆解交易失败")
		return
	}
	if addressLookupTables == nil {
		addressLookupTables = make(map[solana.PublicKey]solana.PublicKeySlice, 1)
	}
	if arb.alt != nil {
		addressLookupTables[arb.alt.Address.PublicKey] = arb.alt.AddressLookupTable
	}

	needFlashLoan := arb.flashLoan.IsPositive() && tradeSize.GreaterThanOrEqual(arb.flashLoan)
	needSpam := arb.enableSpam && profitAsSOL.GreaterThanOrEqual(arb.spamProfitMin)
	needJito := arb.enableJito && profitAsSOL.GreaterThanOrEqual(arb.jitoProfitMin)

	setCULimitIx, setCUPriceIx, createAssociatedTokenAccountIxs, swapIx, beforeArbIx, beforeFlashLoanIx, afterFlashLoanIx, afterArbIxSpam, afterArbIxJito, jitoTipIx, jitoTipIxs, tipPayer, err := arb.refactorIxs(ctx, ixs, tradeSize, profitAsSOL, foundSlot, needFlashLoan, needSpam, needJito)
	if err != nil {
		err = gerror.Wrapf(err, "拆解指令失败")
		return
	}

	if needSpam {
		ixs = make([]solana.Instruction, 0,
			1+ // setCULimitIx
				1+ //setCUPriceIx
				len(createAssociatedTokenAccountIxs)+
				1+ //beforeArbIx
				1+ //beforeFlashLoanIx
				1+ //swapIx
				1+ //afterFlashLoanIx
				1, //afterArbIxSpam
		)

		ixs = append(ixs, setCULimitIx)
		if setCUPriceIx != nil {
			ixs = append(ixs, setCUPriceIx)
		}
		ixs = append(ixs, createAssociatedTokenAccountIxs...)
		ixs = append(ixs, beforeArbIx)
		if needFlashLoan {
			afterFlashLoanIx.BorrowIxIndex = uint8(len(ixs))
			ixs = append(ixs, beforeFlashLoanIx)
		}
		ixs = append(ixs, swapIx)
		if needFlashLoan {
			err = afterFlashLoanIx.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}
		}
		ixs = append(ixs, afterArbIxSpam)

		var swapTx *solana.Transaction
		swapTx, err = arb.officialPool.PackTransaction(ctx, ixs, blockhash, arb.wallet.Account.Address, addressLookupTables)
		if err != nil {
			err = gerror.Wrapf(err, "组装交易失败")
			return
		}

		err = arb.officialPool.SignTransaction(ctx, swapTx, []Wallet.HostedWallet{arb.wallet})
		if err != nil {
			err = gerror.Wrapf(err, "签名交易失败")
			return
		}
		ctx = context.WithValue(ctx, consts.CtxTransaction, swapTx.Signatures[0].String())

		_, err = utils.SerializeTransaction(ctx, swapTx, true)
		if err != nil {
			err = gerror.WrapCodef(errcode.IgnoreError, err, "检查交易限制发现问题")
			return
		}

		spamTx = swapTx
	}

	if needJito {
		ixs = make([]solana.Instruction, 0,
			1+ // setCULimitIx
				len(createAssociatedTokenAccountIxs)+
				1+ //beforeArbIx
				1+ //beforeFlashLoanIx
				1+ //swapIx
				1+ //afterFlashLoanIx
				1+ //afterArbIxJito
				1, //jitoTipIx
		)

		ixs = append(ixs, setCULimitIx)
		ixs = append(ixs, createAssociatedTokenAccountIxs...)
		ixs = append(ixs, beforeArbIx)
		if needFlashLoan {
			afterFlashLoanIx.BorrowIxIndex = uint8(len(ixs))
			ixs = append(ixs, beforeFlashLoanIx)
		}
		ixs = append(ixs, swapIx)
		if needFlashLoan {
			err = afterFlashLoanIx.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}
		}
		ixs = append(ixs, afterArbIxJito)
		if !arb.thirdTipPayer && !arb.dynamicTip {
			ixs = append(ixs, jitoTipIx)
		}

		var swapTx *solana.Transaction
		swapTx, err = arb.officialPool.PackTransaction(ctx, ixs, blockhash, arb.wallet.Account.Address, addressLookupTables)
		if err != nil {
			err = gerror.Wrapf(err, "组装交易失败")
			return
		}

		err = arb.officialPool.SignTransaction(ctx, swapTx, []Wallet.HostedWallet{arb.wallet})
		if err != nil {
			err = gerror.Wrapf(err, "签名交易失败")
			return
		}
		ctx = context.WithValue(ctx, consts.CtxTransaction, swapTx.Signatures[0].String())

		_, err = utils.SerializeTransaction(ctx, swapTx, true)
		if err != nil {
			err = gerror.WrapCodef(errcode.IgnoreError, err, "检查交易限制发现问题")
			return
		}

		jitoTxs = append(jitoTxs, swapTx)

		if arb.thirdTipPayer { // 另起一个小费交易
			var payTipTx *solana.Transaction
			payTipTx, err = arb.officialPool.PackTransaction(ctx, jitoTipIxs, blockhash, tipPayer.Account.Address, addressLookupTables)
			if err != nil {
				err = gerror.Wrapf(err, "组装交易失败")
				return
			}

			err = arb.officialPool.SignTransaction(ctx, payTipTx, []Wallet.HostedWallet{tipPayer})
			if err != nil {
				err = gerror.Wrapf(err, "签名交易失败")
				return
			}

			jitoTxs = append(jitoTxs, payTipTx)
		}
	}

	return
}

func (arb *Arb) refactorIxs(
	ctx context.Context,
	ixs []solana.Instruction,
	tradeSize decimal.Decimal, profitAsSOL decimal.Decimal, foundSlot uint64,
	needFlashLoan, needSpam, needJito bool,
) (
	setCULimitIx solana.Instruction,
	setCUPriceIx solana.Instruction,
	createAssociatedTokenAccountIxs []solana.Instruction,
	swapIx solana.Instruction,
	beforeArbIx solana.Instruction,
	beforeFlashLoanIx solana.Instruction,
	afterFlashLoanIx InstructionKamino.FlashRepay,
	afterArbIxSpam solana.Instruction,
	afterArbIxJito solana.Instruction,
	jitoTipIx solana.Instruction,
	jitoTipIxs []solana.Instruction,
	tipPayer Wallet.HostedWallet,
	err error,
) {
	var setCULimitIx_ InstructionComputeBudget.SetLimit

	for _, ix := range ixs {
		switch ix.ProgramID() {
		case consts.ComputeBudgetProgramAddress.PublicKey:
			var ok bool
			ok, err = InstructionComputeBudget.IsSetLimit(ix)
			if err != nil {
				err = gerror.Wrapf(err, "解析交易失败")
				return
			}
			if ok {
				err = setCULimitIx_.Deserialize(ix)
				if err != nil {
					err = gerror.Wrapf(err, "解析交易失败")
					return
				}

				if setCULimitIx_.Limit == ProgramComputeBudget.MAX_COMPUTE_UNIT_LIMIT {
					err = gerror.NewCodef(errcode.IgnoreError, "CU 异常")
					return
				}

				continue
			}
		case consts.ATAProgramAddress.PublicKey:
			var ok bool
			ok, err = InstructionToken.IsCreateAssociatedTokenAccount(ix)
			if err != nil {
				err = gerror.Wrapf(err, "解析交易失败")
				return
			}
			if ok {
				createAssociatedTokenAccountIxs = append(createAssociatedTokenAccountIxs, ix)
				continue
			}
		case consts.JupiterProgramV6Address.PublicKey:
			swapIx = ix
		default:
			err = gerror.Newf("存在非预期指令 %s", ix.ProgramID().String())
			return
		}
	}

	if setCULimitIx_.Limit == 0 {
		setCULimitIx_.Limit = 40_0000
	} else if arb.extraCULimit > 0 {
		setCULimitIx_.Limit += arb.extraCULimit
		if needFlashLoan {
			setCULimitIx_.Limit += 7_0000 // 4_2115 + 2_7758, borrow 越靠后, borrow 的 cu 越低
		}
	}
	setCULimitIx, err = setCULimitIx_.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	if needSpam {
		if arb.priorityFeeFromProfitBps.IsPositive() {
			setCUPriceIx, err = InstructionComputeBudget.SetPrice{
				Price: uint64(profitAsSOL.
					Mul(arb.priorityFeeFromProfitBps).
					Shift(-4 + 6). // Alias Div(100_00).Mul(1000_000).
					Div(decimal.NewFromInt32(int32(setCULimitIx_.Limit))).
					IntPart()),
			}.ToIx()
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}
		}
	}

	if swapIx == nil {
		err = gerror.Newf("无法找到 Swap 指令")
		return
	}

	beforeArbIx, err = interInstruction.BeforeArb{
		Searcher:      arb.wallet.Account.Address,
		TokenAccount:  arb.tradeTokenAccountAddress,
		BalanceCacher: arb.programBalanceCacherAddress,
	}.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	if needFlashLoan {
		beforeFlashLoanIx, err = InstructionKamino.FlashBorrow{
			Token:                    arb.tradeToken,
			User:                     arb.wallet.Account.Address,
			UserTokenAccount:         arb.tradeTokenAccountAddress,
			Market:                   consts.KaminoJitoMarket,
			MarketAuthority:          consts.KaminoJitoMarketAuthority,
			MarketReserve:            consts.KaminoJitoMarketSOLReserve,
			MarketReserveLiquidity:   consts.KaminoJitoMarketSOLReserveLiquidity,
			MarketReserveFeeReceiver: consts.KaminoJitoMarketSOLReserveFeeReceiver,
			Amount:                   tradeSize,
		}.ToIx()
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}

		afterFlashLoanIx.Token = arb.tradeToken
		afterFlashLoanIx.User = arb.wallet.Account.Address
		afterFlashLoanIx.UserTokenAccount = arb.tradeTokenAccountAddress
		afterFlashLoanIx.Market = consts.KaminoJitoMarket
		afterFlashLoanIx.MarketAuthority = consts.KaminoJitoMarketAuthority
		afterFlashLoanIx.MarketReserve = consts.KaminoJitoMarketSOLReserve
		afterFlashLoanIx.MarketReserveLiquidity = consts.KaminoJitoMarketSOLReserveLiquidity
		afterFlashLoanIx.MarketReserveFeeReceiver = consts.KaminoJitoMarketSOLReserveFeeReceiver
		afterFlashLoanIx.Amount = tradeSize
	}

	if needSpam {
		afterArbIxSpam, err = interInstruction.AfterArb{
			Searcher:             arb.wallet.Account.Address,
			TokenAccount:         arb.tradeTokenAccountAddress,
			BalanceCacher:        arb.programBalanceCacherAddress,
			TipAccount:           arb.wallet.Account.Address,
			UnwrapTipWSOLAccount: arb.programBUnwrapTipWSOLAddress,
			ThirdPayer:           false,
			TipBps:               100_00,
			TipMax:               decimal.Zero,
			FoundSlot:            foundSlot,
			Location:             arb.region,
			Node:                 Utils.XOR([]byte(ctx.Value(consts.CtxRPC).(string)), []byte("6f91c859-788d-4cef-af2d-97f0c31d9394")),
		}.ToIx()
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}
	}

	if needJito {
		var jitoTipAccount Address.AccountAddress
		jitoTipAccount, err = arb.jitoPool.GetRandomTipAccount(ctx)
		if err != nil {
			err = gerror.Wrapf(err, "获取 Jito 小费账户失败")
			return
		}

		tipFromProfitBps := &arb.tipFromProfitBps[0]
		for i := range arb.tipFromProfitBps {
			if profitAsSOL.GreaterThanOrEqual(arb.tipFromProfitBps[i].MinProfit) {
				tipFromProfitBps = &arb.tipFromProfitBps[i]
			} else {
				break
			}
		}

		afterArbIxJito_ := interInstruction.AfterArb{
			Searcher:             arb.wallet.Account.Address,
			TokenAccount:         arb.tradeTokenAccountAddress,
			BalanceCacher:        arb.programBalanceCacherAddress,
			TipAccount:           jitoTipAccount,
			UnwrapTipWSOLAccount: arb.programBUnwrapTipWSOLAddress,
			ThirdPayer:           arb.thirdTipPayer,
			TipBps:               tipFromProfitBps.BpsU16,
			TipMax:               arb.tipMax,
			FoundSlot:            foundSlot,
			Location:             arb.region,
			Node:                 Utils.XOR([]byte(ctx.Value(consts.CtxRPC).(string)), []byte("6f91c859-788d-4cef-af2d-97f0c31d9394")),
		}

		if arb.thirdTipPayer {
			tipPayer = arb.thirdTipPayers[grand.Intn(len(arb.thirdTipPayers))]
			afterArbIxJito_.TipAccount = tipPayer.Account.Address
		} else if !arb.dynamicTip {
			afterArbIxJito_.TipAccount = arb.wallet.Account.Address
		}

		var tipStatic decimal.Decimal
		if !arb.dynamicTip {
			afterArbIxJito_.TipBps = 0
			tipStatic = profitAsSOL.
				Mul(tipFromProfitBps.BpsDec).
				Shift(-4) // Alias Div(100_00)
			if tipStatic.LessThan(interConsts.JitoTipMin) {
				tipStatic = interConsts.JitoTipMin.Copy()
			}
			if tipStatic.GreaterThan(arb.tipMax) {
				tipStatic = arb.tipMax.Copy()
			}
			afterArbIxJito_.TipMax = tipStatic
		}

		afterArbIxJito, err = afterArbIxJito_.ToIx()
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}

		if !arb.thirdTipPayer && !arb.dynamicTip {
			jitoTipIx, err = Instruction.Transfer{
				Sender:   arb.wallet,
				Receiver: jitoTipAccount,
				Amount:   tipStatic,
			}.ToIx()
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}
		}

		if arb.thirdTipPayer { // 另起一个小费交易
			jitoTipIxs = make([]solana.Instruction, 0, 3)

			cuLimit := InstructionComputeBudget.SetLimit{
				Limit: 150,
			}

			if arb.dynamicTip {
				err = interInstruction.PayTip{
					Payer:      tipPayer.Account.Address,
					TipAccount: jitoTipAccount,
					Searcher:   arb.wallet.Account.Address,
				}.AppendIx(&jitoTipIxs)
				if err != nil {
					err = gerror.Wrapf(err, "构建交易失败")
					return
				}
				cuLimit.Limit += 6468 // 5880 * 1.1
			} else {
				err = Instruction.Transfer{
					Sender:   tipPayer,
					Receiver: jitoTipAccount,
					Amount:   tipStatic,
				}.AppendIx(&jitoTipIxs)
				if err != nil {
					err = gerror.Wrapf(err, "构建交易失败")
					return
				}
				cuLimit.Limit += 150

				err = Instruction.Transfer{
					Sender:   tipPayer,
					Receiver: arb.wallet.Account.Address,
					Amount:   lamports.Lamports2SOL(2_039_280).Add(consts.SignFee),
				}.AppendIx(&jitoTipIxs)
				if err != nil {
					err = gerror.Wrapf(err, "构建交易失败")
					return
				}
				cuLimit.Limit += 150
			}

			err = cuLimit.AppendIx(&jitoTipIxs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}
		}
	}

	return
}
