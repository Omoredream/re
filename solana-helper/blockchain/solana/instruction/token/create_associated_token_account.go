package InstructionToken

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"
	ProgramAssociatedTokenAccount "github.com/gagliardetto/solana-go/programs/associated-token-account"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/account"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type CreateAssociatedTokenAccount struct {
	Funder Address.AccountAddress
	Owner  Address.AccountAddress
	Token  Address.TokenAddress
}

type CreateAssociatedTokenAccounts []CreateAssociatedTokenAccount

func (tx CreateAssociatedTokenAccount) ToIx() (ix solana.Instruction, associatedTokenAccount Account.TokenAccount, err error) {
	ix, err = ProgramAssociatedTokenAccount.NewCreateInstruction(
		tx.Funder.PublicKey,
		tx.Owner.PublicKey,
		tx.Token.PublicKey,
	).ValidateAndBuild()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	associatedTokenAccount = Account.TokenAccount{
		Address:      Address.NewFromBytes32(ix.Accounts()[1].PublicKey).AsTokenAccountAddress(),
		TokenAddress: tx.Token,
		OwnerAddress: tx.Owner,
		Token:        decimal.Zero,
	}

	return
}

func (tx CreateAssociatedTokenAccount) AppendIx(ixs *[]solana.Instruction) (associatedTokenAccount Account.TokenAccount, err error) {
	ix, associatedTokenAccount, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs CreateAssociatedTokenAccounts) ToIxs() (ixs []solana.Instruction, associatedTokenAccounts []Account.TokenAccount, err error) {
	ixs = make([]solana.Instruction, len(txs))
	associatedTokenAccounts = make([]Account.TokenAccount, len(txs))
	for i, tx := range txs {
		ixs[i], associatedTokenAccounts[i], err = tx.ToIx()
		if err != nil {
			err = gerror.Wrapf(err, "将交易批量转换为基本指令失败")
			return
		}
	}

	return
}

func (txs CreateAssociatedTokenAccounts) AppendIxs(ixs *[]solana.Instruction) (associatedTokenAccounts []Account.TokenAccount, err error) {
	ixs_, associatedTokenAccounts, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}

func IsCreateAssociatedTokenAccount(ix solana.Instruction) (ok bool, err error) {
	if ix.ProgramID() != consts.ATAProgramAddress.PublicKey {
		ok = false
		return
	}

	data, err := ix.Data()
	if err != nil {
		err = gerror.Wrapf(err, "无法读取指令数据")
		return
	}

	if Utils.BytesLToUint8(data[0:1]) != 0x01 {
		ok = false
		return
	}

	ok = true
	return
}
