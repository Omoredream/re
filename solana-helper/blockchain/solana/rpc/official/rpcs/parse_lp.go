package officialRPCs

import (
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"github.com/gagliardetto/solana-go"
	ProgramSerum "github.com/gagliardetto/solana-go/programs/serum"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/lp"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/parser"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func (pool *RPCs) ParseRaydiumLP(ctx g.Ctx, genesisTransaction solana.Signature) (lp utils.RaydiumLP, err error) {
	var transaction *utils.Transaction
	for transaction == nil {
		transaction, err = pool.GetTransaction(ctx, genesisTransaction)
		if err != nil {
			g.Log().Errorf(ctx, "获取交易失败, %+v", err)
			err = nil
		}
	}

	accounts := transaction.Transaction.Message.AccountKeys
	accounts = append(accounts, transaction.Meta.LoadedAddresses.Writable...) // 不确定是否应当拼接上
	accounts = append(accounts, transaction.Meta.LoadedAddresses.ReadOnly...) // 不确定是否应当拼接上
	ixs := transaction.Transaction.Message.Instructions
	for i := range transaction.Meta.InnerInstructions {
		ixs = append(ixs, transaction.Meta.InnerInstructions[i].Instructions...)
	}
	for _, ix := range ixs {
		if accounts[ix.ProgramIDIndex] != consts.RaydiumProgramV4Address.PublicKey {
			continue
		}

		if ix.Data[0] != 1 {
			continue
		}

		var args LP.RaydiumLiquidityPoolV4Initialize2
		args, err = Parser.ParseRaydiumLiquidityPoolV4Initialize2(ix.Data[1:])
		if err != nil {
			err = gerror.Wrapf(err, "解析程序调用参数失败")
			return
		}

		idAddress := Address.NewFromBytes32(accounts[ix.Accounts[4]])
		authorityAddress := Address.NewFromBytes32(accounts[ix.Accounts[5]])
		openOrdersAddress := Address.NewFromBytes32(accounts[ix.Accounts[6]])
		targetOrdersAddress := Address.NewFromBytes32(accounts[ix.Accounts[12]])
		baseVaultAddress := Address.NewFromBytes32(accounts[ix.Accounts[10]]).AsTokenAccountAddress()
		quoteVaultAddress := Address.NewFromBytes32(accounts[ix.Accounts[11]]).AsTokenAccountAddress()
		marketAddress := Address.NewFromBytes32(accounts[ix.Accounts[16]])
		lpMintAddress := Address.NewFromBytes32(accounts[ix.Accounts[7]]).AsTokenAddress()
		baseMintAddress := Address.NewFromBytes32(accounts[ix.Accounts[8]]).AsTokenAddress()
		quoteMintAddress := Address.NewFromBytes32(accounts[ix.Accounts[9]]).AsTokenAddress()

		var tokens []Token.Token
		tokens, err = pool.TokenCacheGets(ctx, []Address.TokenAddress{baseMintAddress, quoteMintAddress})
		if err != nil {
			err = gerror.Wrapf(err, "批量获取代币失败")
			return
		}

		baseToken := tokens[0]  // 左向代币
		quoteToken := tokens[1] // 右向代币

		var market ProgramSerum.MarketV2 // OpenBook 市场
		market, err = pool.getMarketInfo(ctx, marketAddress)
		if err != nil {
			err = gerror.Wrapf(err, "获取市场信息失败")
			return
		}

		marketBidsAddress := Address.NewFromBytes32(market.Bids)
		marketAsksAddress := Address.NewFromBytes32(market.Asks)
		marketEventQueueAddress := Address.NewFromBytes32(market.EventQueue)
		marketBaseVaultAddress := Address.NewFromBytes32(market.BaseVault).AsTokenAccountAddress()
		marketQuoteVaultAddress := Address.NewFromBytes32(market.QuoteVault).AsTokenAccountAddress()

		var marketAuthorityAddress Address.AccountAddress
		marketAuthorityAddress, err = utils.FindOpenBookAssociatedAuthorityAddress(consts.OpenBookProgramAddress, marketAddress)
		if err != nil {
			err = gerror.Wrapf(err, "派生 OpenBook Authority 地址失败")
			return
		}

		baseBalance := lamports.Lamports2Token(args.InitCoinAmount, baseToken.Info.Decimalx)
		quoteBalance := lamports.Lamports2Token(args.InitPcAmount, quoteToken.Info.Decimalx)

		initialPrice := quoteBalance.Div(baseBalance)

		lp = utils.RaydiumLP{
			Id:                  idAddress,
			Authority:           authorityAddress,
			OpenOrders:          openOrdersAddress,
			TargetOrders:        targetOrdersAddress,
			BaseVault:           baseVaultAddress,
			QuoteVault:          quoteVaultAddress,
			Market:              marketAddress,
			MarketBids:          marketBidsAddress,
			MarketAsks:          marketAsksAddress,
			MarketEventQueue:    marketEventQueueAddress,
			MarketBaseVault:     marketBaseVaultAddress,
			MarketQuoteVault:    marketQuoteVaultAddress,
			MarketAuthority:     marketAuthorityAddress,
			LPMint:              lpMintAddress,
			BaseToken:           baseToken,
			QuoteToken:          quoteToken,
			InitialBaseBalance:  baseBalance,
			InitialQuoteBalance: quoteBalance,
			InitialPrice:        initialPrice,
			InitialLiquidity:    quoteBalance,
			InitialFdv:          baseToken.Info.Supply.Mul(initialPrice),
			OpenTime:            gtime.NewFromTimeStamp(int64(args.OpenTime)),
		}

		return
	}

	err = gerror.Newf("未找到任何 LP 创建交易")

	return
}
