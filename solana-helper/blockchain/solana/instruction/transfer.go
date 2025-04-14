package Instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"
	ProgramSystem "github.com/gagliardetto/solana-go/programs/system"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

type Transfer struct {
	Sender   Wallet.HostedWallet
	Receiver Address.AccountAddress
	Amount   decimal.Decimal
}

type Transfers []Transfer

func (tx Transfer) ToIx() (ix solana.Instruction, err error) {
	ix, err = ProgramSystem.NewTransferInstruction(
		lamports.SOL2Lamports(tx.Amount),
		tx.Sender.Account.Address.PublicKey,
		tx.Receiver.PublicKey,
	).ValidateAndBuild()
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
