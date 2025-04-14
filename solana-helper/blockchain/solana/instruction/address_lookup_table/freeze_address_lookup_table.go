package InstructionAddressLookupTable

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	Utils "git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
)

var freezeDiscriminator = Utils.Uint32ToBytesL(1)

type Freeze struct {
	AddressLookupTable Address.AccountAddress
	Owner              Address.AccountAddress
}

type Freezes []Freeze

func (tx Freeze) ToIx() (ix solana.Instruction, err error) {
	ix, err = Instruction.Custom{
		ProgramID: consts.ALTProgramAddress,
		Accounts: []*solana.AccountMeta{
			tx.AddressLookupTable.Meta().WRITE(),
			tx.Owner.Meta().SIGNER(),
		},
		Discriminator: freezeDiscriminator,
	}.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx Freeze) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs Freezes) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs Freezes) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
