package instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"
	ProgramAssociatedTokenAccount "github.com/gagliardetto/solana-go/programs/associated-token-account"

	Account "git.wkr.moe/web3/solana-helper/blockchain/solana/account"
	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type CreateTokenAccount struct {
	Creator Address.AccountAddress
	Owner   Address.AccountAddress
	Token   Address.TokenAddress
}

type CreateTokenAccounts []CreateTokenAccount

func (tx CreateTokenAccount) ToIx() (ix solana.Instruction, tokenAccount Account.TokenAccount, err error) {
	ix, err = ProgramAssociatedTokenAccount.NewCreateInstruction(
		tx.Creator.PublicKey,
		tx.Owner.PublicKey,
		tx.Token.PublicKey,
	).ValidateAndBuild()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	tokenAccount = Account.TokenAccount{
		Address:      Address.NewFromBytes32(ix.Accounts()[1].PublicKey).AsTokenAccountAddress(),
		TokenAddress: tx.Token,
		OwnerAddress: tx.Owner,
		Token:        decimal.Zero,
	}

	return
}

func (tx CreateTokenAccount) AppendIx(ixs *[]solana.Instruction) (tokenAccount Account.TokenAccount, err error) {
	ix, tokenAccount, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs CreateTokenAccounts) ToIxs() (ixs []solana.Instruction, tokenAccounts []Account.TokenAccount, err error) {
	ixs = make([]solana.Instruction, len(txs))
	tokenAccounts = make([]Account.TokenAccount, len(txs))
	for i, tx := range txs {
		ixs[i], tokenAccounts[i], err = tx.ToIx()
		if err != nil {
			err = gerror.Wrapf(err, "将交易批量转换为基本指令失败")
			return
		}
	}

	return
}

func (txs CreateTokenAccounts) AppendIxs(ixs *[]solana.Instruction) (tokenAccounts []Account.TokenAccount, err error) {
	ixs_, tokenAccounts, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
