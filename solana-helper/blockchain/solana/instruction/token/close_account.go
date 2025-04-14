package InstructionToken

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"
	ProgramToken "github.com/gagliardetto/solana-go/programs/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type CloseAccount struct {
	TokenAccount Address.TokenAccountAddress
	Owner        Address.AccountAddress
	Beneficiary  Address.AccountAddress
}

type CloseAccounts []CloseAccount

func (tx CloseAccount) ToIx() (ix solana.Instruction, err error) {
	ix, err = ProgramToken.NewCloseAccountInstruction(
		tx.TokenAccount.PublicKey,
		tx.Beneficiary.PublicKey,
		tx.Owner.PublicKey,
		nil,
	).ValidateAndBuild()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx CloseAccount) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs CloseAccounts) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs CloseAccounts) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
