package InstructionKamino

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

var flashRepayDiscriminator = consts.KaminoLendingProgramAddress.GetDiscriminator("global", "flash_repay_reserve_liquidity")

type FlashRepay struct {
	Token                    Token.Token
	User                     Address.AccountAddress
	UserTokenAccount         Address.TokenAccountAddress
	Market                   Address.AccountAddress
	MarketAuthority          Address.AccountAddress
	MarketReserve            Address.AccountAddress
	MarketReserveLiquidity   Address.TokenAccountAddress
	MarketReserveFeeReceiver Address.TokenAccountAddress
	Amount                   decimal.Decimal
	BorrowIxIndex            uint8
}

type FlashRepays []FlashRepay

func (tx FlashRepay) ToIx() (ix solana.Instruction, err error) {
	ix, err = Instruction.Custom{
		ProgramID: consts.KaminoLendingProgramAddress,
		Accounts: []*solana.AccountMeta{
			tx.User.Meta().SIGNER(),                    // userTransferAuthority
			tx.MarketAuthority.Meta(),                  // lendingMarketAuthority
			tx.Market.Meta(),                           // lendingMarket
			tx.MarketReserve.Meta().WRITE(),            // reserve
			tx.Token.Address.Meta(),                    // reserveLiquidityMint
			tx.MarketReserveLiquidity.Meta().WRITE(),   // reserveDestinationLiquidity
			tx.UserTokenAccount.Meta().WRITE(),         // userSourceLiquidity
			tx.MarketReserveFeeReceiver.Meta().WRITE(), // reserveLiquidityFeeReceiver
			consts.KaminoLendingProgramAddress.Meta(),  // referrerTokenState
			consts.KaminoLendingProgramAddress.Meta(),  // referrerAccount

			consts.SysVarInstructionsAddress.Meta(),
			consts.TokenProgramAddress.Meta(),
		},
		Discriminator: flashRepayDiscriminator,
		Data: Utils.Append(
			Utils.Uint64ToBytesL(lamports.Token2Lamports(tx.Amount, tx.Token.Info.Decimalx)), // liquidityAmount
			Utils.Uint8ToBytesL(tx.BorrowIxIndex),                                            // borrowInstructionIndex
		),
	}.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx FlashRepay) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs FlashRepays) ToIxs() (ixs []solana.Instruction, err error) {
	ixs = make([]solana.Instruction, len(txs))
	for i, tx := range txs {
		ixs[i], err = tx.ToIx()
		if err != nil {
			err = gerror.Wrapf(err, "将交易批量转换为基本指令失败")
			return
		}
	}

	return
}

func (txs FlashRepays) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
