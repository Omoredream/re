package instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	Utils "git.wkr.moe/web3/solana-helper/utils"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	Token "git.wkr.moe/web3/solana-helper/blockchain/solana/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

type BurnToken2022 struct {
	Amount       decimal.Decimal
	TokenAccount Address.TokenAccountAddress
	Token        Token.Token
	Owner        Address.AccountAddress
}

type BurnToken2022s []BurnToken2022

func (tx BurnToken2022) ToIx() (ix solana.Instruction, err error) {
	ix, err = Custom{
		ProgramID: consts.Token2022ProgramAddress,
		Accounts: []*solana.AccountMeta{
			tx.TokenAccount.Meta().WRITE(),
			tx.Token.Address.Meta().WRITE(),
			tx.Owner.Meta().WRITE().SIGNER(),
		},
		Discriminator: []byte{0x08},
		Data: Utils.Append(
			Utils.Uint64ToBytesL(lamports.Token2Lamports(tx.Amount, tx.Token.Info.Decimalx)),
		),
	}.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx BurnToken2022) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs BurnToken2022s) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs BurnToken2022s) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
