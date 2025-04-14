package instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"

	interConsts "git.wkr.moe/web3/solana-helper/project/arb/internal/consts"
)

var beforeArbDiscriminator = interConsts.ArbProgramAddress.GetDiscriminator("global", "before_arb")

type BeforeArb struct {
	Searcher      Address.AccountAddress
	TokenAccount  Address.TokenAccountAddress
	BalanceCacher Address.AccountAddress
}

type BeforeArbs []BeforeArb

func (tx BeforeArb) ToIx() (ix solana.Instruction, err error) {
	ix, err = Instruction.Custom{
		ProgramID: interConsts.ArbProgramAddress,
		Accounts: []*solana.AccountMeta{
			tx.Searcher.Meta().SIGNER().WRITE(),
			tx.TokenAccount.Meta(),
			tx.BalanceCacher.Meta().WRITE(),

			consts.SystemProgramAddress.Meta(),
		},
		Discriminator: beforeArbDiscriminator,
	}.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx BeforeArb) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs BeforeArbs) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs BeforeArbs) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
