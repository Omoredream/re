package instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"
	ProgramToken "github.com/gagliardetto/solana-go/programs/token"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	Token "git.wkr.moe/web3/solana-helper/blockchain/solana/token"
	Wallet "git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

type TransferToken struct {
	Owner    Wallet.HostedWallet
	Sender   Address.TokenAccountAddress
	Receiver Address.TokenAccountAddress
	Amount   decimal.Decimal
	Token    Token.Token
}

type TransferTokens []TransferToken

func (tx TransferToken) ToIx() (ix solana.Instruction, err error) {
	ix, err = ProgramToken.NewTransferInstruction(
		lamports.Token2Lamports(tx.Amount, tx.Token.Info.Decimalx),
		tx.Sender.PublicKey,
		tx.Receiver.PublicKey,
		tx.Owner.Account.Address.PublicKey,
		[]solana.PublicKey{},
	).ValidateAndBuild()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx TransferToken) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs TransferTokens) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs TransferTokens) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
