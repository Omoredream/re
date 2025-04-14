package InstructionAddressLookupTable

import (
	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
)

var createDiscriminator = Utils.Uint32ToBytesL(0)

type Create struct {
	Funder Address.AccountAddress
	Owner  Address.AccountAddress
	Slot   uint64
}

type Creates []Create

func (tx Create) ToIx() (ix solana.Instruction, addressLookupTableAddress Address.AccountAddress, err error) {
	addressLookupTableAddress, bumpSeed, err := consts.ALTProgramAddress.FindProgramDerivedAddress([][]byte{
		tx.Owner.Bytes(),
		Utils.Uint64ToBytesL(tx.Slot),
	})
	if err != nil {
		err = gerror.Wrapf(err, "派生地址查找表失败")
		return
	}

	ix, err = Instruction.Custom{
		ProgramID: consts.ALTProgramAddress,
		Accounts: []*solana.AccountMeta{
			addressLookupTableAddress.Meta().WRITE(),
			tx.Owner.Meta().SIGNER(),
			tx.Funder.Meta().SIGNER().WRITE(),
			consts.SystemProgramAddress.Meta(),
		},
		Discriminator: createDiscriminator,
		Data: Utils.Append(
			Utils.Uint64ToBytesL(tx.Slot),
			Utils.Uint8ToBytesL(bumpSeed),
		),
	}.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx Create) AppendIx(ixs *[]solana.Instruction) (addressLookupTableAddress Address.AccountAddress, err error) {
	ix, addressLookupTableAddress, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs Creates) ToIxs() (ixs []solana.Instruction, addressLookupTableAddresses []Address.AccountAddress, err error) {
	ixs = make([]solana.Instruction, len(txs))
	addressLookupTableAddresses = make([]Address.AccountAddress, len(txs))
	for i, tx := range txs {
		ixs[i], addressLookupTableAddresses[i], err = tx.ToIx()
		if err != nil {
			err = gerror.Wrapf(err, "将交易批量转换为基本指令失败")
			return
		}
	}

	return
}

func (txs Creates) AppendIxs(ixs *[]solana.Instruction) (addressLookupTableAddresses []Address.AccountAddress, err error) {
	ixs_, addressLookupTableAddresses, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
