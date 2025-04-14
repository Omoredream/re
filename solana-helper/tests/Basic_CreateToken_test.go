package tests

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	ProgramMetaplexTokenMetadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	"github.com/gagliardetto/solana-go"
	ProgramToken "github.com/gagliardetto/solana-go/programs/token"

	"git.wkr.moe/web3/solana-helper/ipfs/pinata"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/metaplex_token_metadata"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"
)

func TestCreateToken(t *testing.T) {
	err := testCreateToken(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testCreateToken(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	tokenWallet, err := officialPool.NewWalletFromMnemonic(ctx, testWalletMnemonic, 1<<16+0)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	pinata := Pinata.NewPinata(testPinataToken)

	logoCID, err := pinata.UploadFile("logo.jpg", gfile.GetBytes("logo.jpg"))
	if err != nil {
		err = gerror.Wrapf(err, "上传图标失败")
		return
	}

	metadataCID, err := pinata.UploadJson("metadata.json", map[string]any{
		"name":        "Test Coin",
		"symbol":      "TC",
		"image":       pinata.Gateway(logoCID),
		"description": "just a test coin",
		"extensions":  map[string]any{},
		"tags": []string{
			"SOLANA",
			"MEME",
		},
		"creator": map[string]any{},
	})
	if err != nil {
		err = gerror.Wrapf(err, "上传元数据失败")
		return
	}

	tokenInfo := Token.Token{
		Address: tokenWallet.Account.Address.AsTokenAddress(),
		Info: Token.Info{
			Supply:   decimal.NewFromFloat(888_888_888_888),
			Decimals: 6,
			Decimalx: 6,
		},
		Metadata: &Token.Metadata{
			Name:                 "Test Token",
			Symbol:               "TC",
			Uri:                  pinata.Gateway(metadataCID),
			SellerFeeBasisPoints: 0,
			TokenStandard:        lo.ToPtr(ProgramMetaplexTokenMetadata.TokenStandardFungible),
		},
		TokenStandard: ProgramMetaplexTokenMetadata.TokenStandardFungible,
	}

	rentExemption, err := officialPool.GetRentExemption(ctx, ProgramToken.MINT_SIZE)
	if err != nil {
		err = gerror.Wrapf(err, "获取免租余额失败")
		return
	}

	ixs := make([]solana.Instruction, 0, 3)

	err = Instruction.CreateAccount{
		Funder:  mainWallet.Account.Address,
		Owner:   consts.TokenProgramAddress.AccountAddress,
		Account: tokenInfo.Address.AccountAddress,
		Size:    ProgramToken.MINT_SIZE,
		Balance: rentExemption,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	err = InstructionToken.Init{
		Token:           tokenInfo.Address,
		Decimals:        tokenInfo.Info.Decimals,
		MintAuthority:   mainWallet.Account.Address,
		FreezeAuthority: mainWallet.Account.Address,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	err = InstructionMetaplexTokenMetadata.Create{
		Creator:              mainWallet.Account.Address,
		Token:                tokenInfo.Address,
		MintAuthority:        mainWallet.Account.Address,
		UpdateAuthority:      mainWallet.Account.Address,
		Name:                 tokenInfo.Metadata.Name,
		Symbol:               tokenInfo.Metadata.Symbol,
		IPFSUrl:              tokenInfo.Metadata.Uri,
		SellerFeeBasisPoints: tokenInfo.Metadata.SellerFeeBasisPoints,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	tokenAccount, err := InstructionToken.CreateAssociatedTokenAccount{
		Funder: mainWallet.Account.Address,
		Owner:  mainWallet.Account.Address,
		Token:  tokenInfo.Address,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	err = InstructionToken.Mint{
		Minter:   mainWallet.Account.Address,
		Receiver: tokenAccount.Address,
		Token:    tokenInfo,
		Amount:   tokenInfo.Info.Supply,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	err = InstructionToken.SetAuthority{
		Token:         tokenInfo.Address,
		AuthorityType: ProgramToken.AuthorityMintTokens,
		OldAuthority:  mainWallet.Account.Address,
		NewAuthority:  nil,
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	txHash, err := officialPool.SendInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
	if err != nil {
		err = gerror.Wrapf(err, "发送交易失败")
		return
	}
	ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
	g.Log().Infof(ctx, "已发送交易")

	spent, err := officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
	if err != nil {
		err = gerror.Wrapf(err, "等待交易确认失败")
		return
	}
	g.Log().Infof(ctx, "交易耗时 %s", spent)

	return
}
