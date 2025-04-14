package InstructionComputeBudget

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"
	ProgramComputeBudget "github.com/gagliardetto/solana-go/programs/compute-budget"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"
)

type SetLimit struct {
	Limit uint32
}

type SetLimits []SetLimit

func (tx SetLimit) ToIx() (ix solana.Instruction, err error) {
	ix, err = ProgramComputeBudget.NewSetComputeUnitLimitInstruction(
		min(tx.Limit, ProgramComputeBudget.MAX_COMPUTE_UNIT_LIMIT),
	).ValidateAndBuild()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx SetLimit) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs SetLimits) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs SetLimits) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}

func IsSetLimit(ix solana.Instruction) (ok bool, err error) {
	if ix.ProgramID() != consts.ComputeBudgetProgramAddress.PublicKey {
		ok = false
		return
	}

	data, err := ix.Data()
	if err != nil {
		err = gerror.Wrapf(err, "无法读取指令数据")
		return
	}

	if Utils.BytesLToUint8(data[0:1]) != ProgramComputeBudget.Instruction_SetComputeUnitLimit {
		ok = false
		return
	}

	ok = true
	return
}

func (tx *SetLimit) Deserialize(ix solana.Instruction) (err error) {
	data, err := ix.Data()
	if err != nil {
		err = gerror.Wrapf(err, "无法读取指令数据")
		return
	}

	tx.Limit = Utils.BytesLToUint32(data[1:5])
	return
}
