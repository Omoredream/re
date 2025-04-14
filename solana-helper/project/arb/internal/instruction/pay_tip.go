package instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"

	interConsts "git.wkr.moe/web3/solana-helper/project/arb/internal/consts"
)

var payTipDiscriminator = interConsts.ArbProgramAddress.GetDiscriminator("global", "pay_tip")

type PayTip struct {
	Payer      Address.AccountAddress
	TipAccount Address.AccountAddress
	Searcher   Address.AccountAddress
}

type PayTips []PayTip

func (tx PayTip) ToIx() (ix solana.Instruction, err error) {
	ix, err = Instruction.Custom{
		ProgramID: interConsts.ArbProgramAddress,
		Accounts: []*solana.AccountMeta{
			tx.Payer.Meta().SIGNER().WRITE(),
			tx.TipAccount.Meta().WRITE(),
			tx.Searcher.Meta().WRITE(),

			consts.SystemProgramAddress.Meta(),
		},
		Discriminator: payTipDiscriminator,
		Data:          nil,
	}.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx PayTip) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs PayTips) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs PayTips) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
