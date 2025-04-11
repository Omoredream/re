package instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	Token "git.wkr.moe/web3/solana-helper/blockchain/solana/token"
)

type HarvestWithheldToken struct {
	Token        Token.Token
	TokenAccount Address.TokenAccountAddress
}

type HarvestWithheldTokens []HarvestWithheldToken

func (tx HarvestWithheldToken) ToIx() (ix solana.Instruction, err error) {
	ix, err = Custom{
		ProgramID: consts.Token2022ProgramAddress,
		Accounts: []*solana.AccountMeta{
			tx.Token.Address.Meta().WRITE(),
			tx.TokenAccount.Meta().WRITE(),
		},
		Discriminator: []byte{0x1a, 0x04},
	}.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx HarvestWithheldToken) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs HarvestWithheldTokens) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs HarvestWithheldTokens) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
