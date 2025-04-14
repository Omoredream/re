package InstructionAddressLookupTable

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
)

var extendDiscriminator = Utils.Uint32ToBytesL(2)

type Extend struct {
	Funder             Address.AccountAddress
	AddressLookupTable Address.AccountAddress
	Owner              Address.AccountAddress
	Addresses          []Address.AccountAddress
}

type Extends []Extend

func (tx Extend) ToIx() (ix solana.Instruction, err error) {
	ix, err = Instruction.Custom{
		ProgramID: consts.ALTProgramAddress,
		Accounts: []*solana.AccountMeta{
			tx.AddressLookupTable.Meta().WRITE(),
			tx.Owner.Meta().SIGNER(),
			tx.Funder.Meta().SIGNER().WRITE(),
			consts.SystemProgramAddress.Meta(),
		},
		Discriminator: extendDiscriminator,
		Data: Utils.Append(
			Utils.Vec64TToBytes(func(address Address.AccountAddress) (b []byte) {
				return address.Bytes()
			}, tx.Addresses...),
		),
	}.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx Extend) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs Extends) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs Extends) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
