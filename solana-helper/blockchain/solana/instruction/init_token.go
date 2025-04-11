package instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"
	ProgramToken "github.com/gagliardetto/solana-go/programs/token"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type InitToken struct {
	Token           Address.TokenAddress
	Decimals        uint8
	MintAuthority   Address.AccountAddress
	FreezeAuthority Address.AccountAddress
}

type InitTokens []InitToken

func (tx InitToken) ToIx() (ix solana.Instruction, err error) {
	ix, err = ProgramToken.NewInitializeMint2Instruction(
		tx.Decimals,
		tx.MintAuthority.PublicKey,
		tx.FreezeAuthority.PublicKey,
		tx.Token.PublicKey,
	).ValidateAndBuild()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx InitToken) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs InitTokens) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs InitTokens) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
