package instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"

	interConsts "git.wkr.moe/web3/solana-helper/project/arb/internal/consts"
)

var afterArbDiscriminator = interConsts.ArbProgramAddress.GetDiscriminator("global", "after_arb")

type AfterArb struct {
	Searcher             Address.AccountAddress
	TokenAccount         Address.TokenAccountAddress
	BalanceCacher        Address.AccountAddress
	TipAccount           Address.AccountAddress
	UnwrapTipWSOLAccount Address.TokenAccountAddress
	ThirdPayer           bool
	TipBps               uint16
	TipMax               decimal.Decimal
	FoundSlot            uint64
	Location             string
	Node                 []byte
}

type AfterArbs []AfterArb

func (tx AfterArb) ToIx() (ix solana.Instruction, err error) {
	ix, err = Instruction.Custom{
		ProgramID: interConsts.ArbProgramAddress,
		Accounts: []*solana.AccountMeta{
			tx.Searcher.Meta().SIGNER().WRITE(),
			tx.TokenAccount.Meta().WRITE(),
			tx.BalanceCacher.Meta().WRITE(),
			tx.TipAccount.Meta().WRITE(),
			tx.UnwrapTipWSOLAccount.Meta().WRITE(),

			consts.SystemProgramAddress.Meta(),
			consts.TokenProgramAddress.Meta(),
			consts.SOL.Address.Meta(),
			interConsts.ArbEventAuthority.Meta(),
			interConsts.ArbProgramAddress.Meta(),
		},
		Discriminator: afterArbDiscriminator,
		Data: Utils.Append(
			Utils.BoolToBytes(tx.ThirdPayer),
			Utils.Uint16ToBytesL(tx.TipBps),
			Utils.Uint64ToBytesL(lamports.SOL2Lamports(tx.TipMax)),
			Utils.Uint64ToBytesL(tx.FoundSlot),
			Utils.StringToBytes(tx.Location),
			Utils.BytesToBytes(tx.Node),
		),
	}.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx AfterArb) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs AfterArbs) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs AfterArbs) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
