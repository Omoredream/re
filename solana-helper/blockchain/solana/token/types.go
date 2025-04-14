package Token

import (
	"github.com/shopspring/decimal"

	ProgramMetaplexTokenMetadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type Token struct {
	Address       Address.TokenAddress
	Info          Info
	Metadata      *Metadata
	TokenStandard ProgramMetaplexTokenMetadata.TokenStandard
}

func (token *Token) String() string {
	return token.Address.String()
}

func (token *Token) DisplayName() (name string) {
	if token.Metadata != nil && (token.Metadata.Symbol != "" || token.Metadata.Name != "") {
		if token.IsToken() {
			if token.Metadata.Symbol != "" {
				name += "$" + token.Metadata.Symbol
			}
			if token.Metadata.Name != "" {
				name += "(" + token.Metadata.Name + ")"
			}
		} else {
			if token.Metadata.Name != "" {
				name += "#" + token.Metadata.Name
			}
			if token.Metadata.Symbol != "" {
				name += "<" + token.Metadata.Symbol + ">"
			}
		}
		name += "[" + token.Address.ShortString() + "]"
	} else {
		name = token.Address.ShortString()
	}
	return
}

func (token *Token) IsToken() bool {
	return token.TokenStandard == ProgramMetaplexTokenMetadata.TokenStandardFungible
}

type Info struct {
	Supply   decimal.Decimal
	Decimals uint8
	Decimalx int32
}

type Metadata struct {
	Name                 string
	Symbol               string
	Uri                  string
	SellerFeeBasisPoints uint16
	TokenStandard        *ProgramMetaplexTokenMetadata.TokenStandard
}
