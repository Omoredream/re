package InstructionToken

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"
	ProgramToken "github.com/gagliardetto/solana-go/programs/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type SetAuthority struct {
	Token         Address.TokenAddress
	AuthorityType ProgramToken.AuthorityType
	OldAuthority  Address.AccountAddress
	NewAuthority  *Address.AccountAddress
}

type SetAuthoritys []SetAuthority

func (tx SetAuthority) ToIx() (ix solana.Instruction, err error) {
	builder := ProgramToken.NewSetAuthorityInstructionBuilder()
	builder.SetAuthorityType(tx.AuthorityType)
	if tx.NewAuthority != nil {
		builder.SetNewAuthority(tx.NewAuthority.PublicKey)
	}
	builder.SetSubjectAccount(tx.Token.PublicKey)
	builder.SetAuthorityAccount(tx.OldAuthority.PublicKey)
	ix, err = builder.ValidateAndBuild()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx SetAuthority) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs SetAuthoritys) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs SetAuthoritys) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
