package InstructionToken2022

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

type Transfer struct {
	Owner    Wallet.HostedWallet
	Sender   Address.TokenAccountAddress
	Receiver Address.TokenAccountAddress
	Amount   decimal.Decimal
	Token    Token.Token
}

type Transfers []Transfer

func (tx Transfer) ToIx() (ix solana.Instruction, err error) {
	ix, err = Instruction.Custom{
		ProgramID: consts.Token2022ProgramAddress,
		Accounts: []*solana.AccountMeta{
			tx.Sender.Meta().WRITE(),
			tx.Receiver.Meta().WRITE(),
			tx.Owner.Account.Address.Meta().SIGNER().WRITE(),
		},
		Discriminator: []byte{0x03},
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

func (tx Transfer) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs Transfers) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs Transfers) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
