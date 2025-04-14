package Parser

import (
	"github.com/gagliardetto/binary"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/shopspring/decimal"

	ProgramMetaplexTokenMetadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	ProgramSerum "github.com/gagliardetto/solana-go/programs/serum"
	ProgramToken "github.com/gagliardetto/solana-go/programs/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/lp"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func ParseToken(tokenAddress Address.TokenAddress, tokenInfo Token.Info, tokenMetadata *Token.Metadata) (token Token.Token) {
	tokenStandard := ProgramMetaplexTokenMetadata.TokenStandardNonFungible
	if tokenMetadata != nil && tokenMetadata.TokenStandard != nil {
		tokenStandard = *tokenMetadata.TokenStandard
	} else {
		if tokenInfo.Supply.Cmp(decimal.New(1, 0)) == 0 {
			tokenStandard = ProgramMetaplexTokenMetadata.TokenStandardNonFungible
		} else if tokenInfo.Supply.Cmp(decimal.New(1, 0)) > 0 && tokenInfo.Decimals == 0 {
			tokenStandard = ProgramMetaplexTokenMetadata.TokenStandardFungibleAsset
		} else {
			tokenStandard = ProgramMetaplexTokenMetadata.TokenStandardFungible
		}
	}

	token = Token.Token{
		Address:       tokenAddress,
		Info:          tokenInfo,
		Metadata:      tokenMetadata,
		TokenStandard: tokenStandard,
	}

	return
}

func ParseTokenAccountInfo(data []byte) (tokenAccountInfo TTokenAccountInfo, err error) {
	var accountInfo ProgramToken.Account
	err = bin.NewBinDecoder(data).Decode(&accountInfo)
	if err != nil {
		err = gerror.Wrapf(err, "解析代币账户信息失败")
		return
	}

	tokenAccountInfo = TTokenAccountInfo{
		TokenAddress: Address.NewFromBytes32(accountInfo.Mint).AsTokenAddress(),
		OwnerAddress: Address.NewFromBytes32(accountInfo.Owner),
		Token:        accountInfo.Amount,
	}

	return
}

func ParseTokenInfo(data []byte) (token Token.Info, err error) {
	var tokenInfo ProgramToken.Mint
	err = bin.NewBinDecoder(data).Decode(&tokenInfo)
	if err != nil {
		err = gerror.Wrapf(err, "解析代币信息失败")
		return
	}

	decimalx := int32(tokenInfo.Decimals)
	token = Token.Info{
		Supply:   lamports.Lamports2Token(tokenInfo.Supply, decimalx),
		Decimals: tokenInfo.Decimals,
		Decimalx: decimalx,
	}

	return
}

func ParseTokenMetadata(data []byte) (token Token.Metadata, err error) {
	var tokenMetadata ProgramMetaplexTokenMetadata.Metadata
	err = bin.NewBorshDecoder(data).Decode(&tokenMetadata)
	if err != nil {
		err = gerror.Wrapf(err, "解析代币元数据失败")
		return
	}
	token = Token.Metadata{
		Name:                 gstr.TrimRightStr(tokenMetadata.Data.Name, "\x00"),
		Symbol:               gstr.TrimRightStr(tokenMetadata.Data.Symbol, "\x00"),
		Uri:                  gstr.TrimRightStr(tokenMetadata.Data.Uri, "\x00"),
		SellerFeeBasisPoints: tokenMetadata.Data.SellerFeeBasisPoints,
		TokenStandard:        tokenMetadata.TokenStandard,
	}

	return
}

func ParseMarketInfo(data []byte) (market ProgramSerum.MarketV2, err error) {
	err = bin.NewBinDecoder(data).Decode(&market)
	if err != nil {
		err = gerror.Wrapf(err, "解析市场信息失败")
		return
	}

	return
}

func ParseRaydiumLiquidityPoolV4Initialize2(data []byte) (args LP.RaydiumLiquidityPoolV4Initialize2, err error) {
	err = bin.NewBinDecoder(data).Decode(&args)
	if err != nil {
		err = gerror.Wrapf(err, "解析 Raydium Liquidity Pool V4 程序调用 Initialize2 参数失败")
		return
	}

	return
}
