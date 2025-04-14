package Instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"
	ProgramSystem "github.com/gagliardetto/solana-go/programs/system"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

type CreateAccount struct {
	Funder  Address.AccountAddress
	Owner   Address.AccountAddress
	Account Address.AccountAddress
	Size    uint64
	Balance decimal.Decimal
}

type CreateAccounts []CreateAccount

func (tx CreateAccount) ToIx() (ix solana.Instruction, err error) {
	ix, err = ProgramSystem.NewCreateAccountInstruction(
		lamports.SOL2Lamports(tx.Balance),
		tx.Size,
		tx.Owner.PublicKey,
		tx.Funder.PublicKey,
		tx.Account.PublicKey,
	).ValidateAndBuild()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx CreateAccount) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs CreateAccounts) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs CreateAccounts) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
