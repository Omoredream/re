package InstructionMetaplexTokenMetadata

import (
	"github.com/gogf/gf/v2/errors/gerror"

	ProgramMetaplexTokenMetadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type Create struct {
	Creator              Address.AccountAddress
	Token                Address.TokenAddress
	MintAuthority        Address.AccountAddress
	UpdateAuthority      Address.AccountAddress
	Name                 string
	Symbol               string
	IPFSUrl              string
	SellerFeeBasisPoints uint16
}

type Creates []Create

func (tx Create) ToIx() (ix solana.Instruction, err error) {
	tokenMetadata, err := tx.Token.FindTokenMetadataAddress()
	if err != nil {
		err = gerror.Wrapf(err, "派生代币元数据地址失败")
		return
	}

	builder := ProgramMetaplexTokenMetadata.NewCreateMetadataAccountV3InstructionBuilder()
	builder.SetCreateMetadataAccountArgsV3(ProgramMetaplexTokenMetadata.CreateMetadataAccountArgsV3{
		Data: ProgramMetaplexTokenMetadata.DataV2{
			Name:                 tx.Name,
			Symbol:               tx.Symbol,
			Uri:                  tx.IPFSUrl,
			SellerFeeBasisPoints: tx.SellerFeeBasisPoints,
		},
		IsMutable: true,
	})
	builder.SetMetadataAccount(tokenMetadata.PublicKey)
	builder.SetMintAccount(tx.Token.PublicKey)
	builder.SetMintAuthorityAccount(tx.MintAuthority.PublicKey)
	builder.SetPayerAccount(tx.Creator.PublicKey)
	builder.SetUpdateAuthorityAccount(tx.UpdateAuthority.PublicKey)
	builder.SetSystemProgramAccount(consts.SystemProgramAddress.PublicKey)
	//builder.SetRentAccount(consts.SysVarRentAddress.PublicKey)
	ix, err = builder.ValidateAndBuild()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	return
}

func (tx Create) AppendIx(ixs *[]solana.Instruction) (err error) {
	ix, err := tx.ToIx()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ix)

	return
}

func (txs Creates) ToIxs() (ixs []solana.Instruction, err error) {
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

func (txs Creates) AppendIxs(ixs *[]solana.Instruction) (err error) {
	ixs_, err := txs.ToIxs()
	if err != nil {
		err = gerror.Wrapf(err, "将交易转换为基本指令失败")
		return
	}

	*ixs = append(*ixs, ixs_...)

	return
}
