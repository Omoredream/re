package instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"
	ProgramComputeBudget "github.com/gagliardetto/solana-go/programs/compute-budget"
)

type SetCUPrice struct {
	Price uint64
}

type SetCUPrices []SetCUPrice

func (tx SetCUPrice) ToIx() (ix solana.Instruction, err error) {
	ix, err = ProgramComputeBudget.NewSetComputeUnitPriceInstruction(
		tx.Price,
	).ValidateAndBuild()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx SetCUPrice) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs SetCUPrices) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs SetCUPrices) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
