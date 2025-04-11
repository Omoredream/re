package instruction

import (
	"github.com/gogf/gf/v2/errors/gerror"

	ProgramMetaplexTokenMetadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	"github.com/gagliardetto/solana-go"

	Address "git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type UpdateTokenMetadata struct {
	Updater              Address.AccountAddress
	Token                Address.TokenAddress
	Name                 string
	Symbol               string
	IPFSUrl              string
	SellerFeeBasisPoints uint16
}

type UpdateTokenMetadatas []UpdateTokenMetadata

func (tx UpdateTokenMetadata) ToIx() (ix solana.Instruction, err error) {
	tokenMetadata, err := tx.Token.FindTokenMetadataAddress()
	if err != nil {
		err = gerror.Wrapf(err, "派生代币元数据地址失败")
		return
	}

	builder := ProgramMetaplexTokenMetadata.NewUpdateMetadataAccountV2InstructionBuilder()
	builder.SetUpdateMetadataAccountArgsV2(ProgramMetaplexTokenMetadata.UpdateMetadataAccountArgsV2{
		Data: &ProgramMetaplexTokenMetadata.DataV2{
			Name:                 tx.Name,
			Symbol:               tx.Symbol,
			Uri:                  tx.IPFSUrl,
			SellerFeeBasisPoints: tx.SellerFeeBasisPoints,
		},
	})
	builder.SetMetadataAccount(tokenMetadata.PublicKey)
	builder.SetUpdateAuthorityAccount(tx.Updater.PublicKey)
	ix, err = builder.ValidateAndBuild()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx UpdateTokenMetadata) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs UpdateTokenMetadatas) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs UpdateTokenMetadatas) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
