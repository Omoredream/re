package consts

import (
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	ProgramMetaplexTokenMetadata "github.com/gagliardetto/metaplex-go/clients/token-metadata"
	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
)

const (
	CtxRPC         = "rpc"
	CtxWallet      = "wallet"
	CtxAccount     = "account"
	CtxAddress     = "address"
	CtxDerivation  = "derivation"
	CtxToken       = "token"
	CtxTransaction = "transaction"
	CtxBundle      = "bundle"

	CtxValueJupiterHttpRPC = "jupiterHttpRPC"
)

var (
	SignFee           = decimal.NewFromFloat(0.000_005_000)
	MaxTxSize         = 1232
	MaxTxAccountCount = 64
)

var (
	NullAddress                    = Address.NewFromBytes32(solana.PublicKey{})
	SystemProgramAddress           = Address.NewFromBytes32(solana.SystemProgramID).AsProgramAddress()
	SysVarRentAddress              = Address.NewFromBytes32(solana.SysVarRentPubkey).AsProgramAddress()
	SysVarInstructionsAddress      = Address.NewFromBytes32(solana.SysVarInstructionsPubkey).AsProgramAddress()
	SysVarSlotHashesAddress        = Address.NewFromBytes32(solana.SysVarSlotHashesPubkey).AsProgramAddress()
	ComputeBudgetProgramAddress    = Address.NewFromBytes32(solana.ComputeBudget).AsProgramAddress()
	Ed25519SigVerifyProgramAddress = Address.NewFromBase58("Ed25519SigVerify111111111111111111111111111").AsProgramAddress()
	ALTProgramAddress              = Address.NewFromBase58("AddressLookupTab1e1111111111111111111111111").AsProgramAddress()
	TokenProgramAddress            = Address.NewFromBytes32(solana.TokenProgramID).AsProgramAddress()
	ATAProgramAddress              = Address.NewFromBytes32(solana.SPLAssociatedTokenAccountProgramID).AsProgramAddress()
	Token2022ProgramAddress        = Address.NewFromBase58("TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb").AsProgramAddress()
	TokenMetadataProgramAddress    = Address.NewFromBytes32(solana.TokenMetadataProgramID).AsProgramAddress()
	MemoProgramAddress             = Address.NewFromBytes32(solana.MemoProgramID).AsProgramAddress()

	StateCompressionProgramAddress   = Address.NewFromBase58("cmtDvXumGCrqC1Age74AVPhSRVXJMd8PJS91L8KbNCK").AsProgramAddress()
	NoopProgramAddress               = Address.NewFromBase58("noopb9bkMVfRPU8AsbpTUg8AQkHtKwMYZiFUjNRtMmV").AsProgramAddress()
	RaydiumProgramV4Address          = Address.NewFromBase58("675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8").AsProgramAddress()
	RaydiumCreatePoolChargingAddress = Address.NewFromBase58("7YttLkHDoNj9wyDur5pM1ejNaAvT9X4eqaYcHQqtj2G5").AsTokenAddress()
	OpenBookProgramAddress           = Address.NewFromBase58("srmqPvymJeFKQ4zGQed1GFppgkRHL9kaELCbyksJtPX").AsProgramAddress()
	JupiterProgramV6Address          = Address.NewFromBase58("JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4").AsProgramAddress()
	MeteoraDLMMProgramAddress        = Address.NewFromBase58("LBUZKhRxPF3XUpBCjp4YzTKgLccjZhTSDM9YuVaPwxo").AsProgramAddress()
	BubblegumProgramAddress          = Address.NewFromBase58("BGUMAp9Gq7iTEuizy4pqaxsTyUCBK68MDfK752saRPUY").AsProgramAddress()
	MetaplexCoreProgramAddress       = Address.NewFromBase58("CoREENxT6tW1HoK8ypY1SxRMZTcVPm7R94rH4PZNhX7d").AsProgramAddress()
	KaminoLendingProgramAddress      = Address.NewFromBase58("KLend2g3cP87fffoy8q1mQqGKjrxjC8boSyAYavgmjD").AsProgramAddress()

	solAddress  = Address.NewFromBase58("So11111111111111111111111111111111111111112").AsTokenAddress()
	usdcAddress = Address.NewFromBase58("EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v").AsTokenAddress()
	usdtAddress = Address.NewFromBase58("Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB").AsTokenAddress()
)

var (
	SOL = Token.Token{
		Address: solAddress,
		Info: Token.Info{
			Supply:   decimal.Zero,
			Decimals: 9,
			Decimalx: 9,
		},
		Metadata: &Token.Metadata{
			Name:                 "Wrapped SOL",
			Symbol:               "SOL",
			Uri:                  "",
			SellerFeeBasisPoints: 0,
			TokenStandard:        lo.ToPtr(ProgramMetaplexTokenMetadata.TokenStandardFungible),
		},
		TokenStandard: ProgramMetaplexTokenMetadata.TokenStandardFungible,
	}
	USDC = Token.Token{
		Address: usdcAddress,
		Info: Token.Info{
			Supply:   decimal.Zero,
			Decimals: 6,
			Decimalx: 6,
		},
		Metadata: &Token.Metadata{
			Name:                 "USD Coin",
			Symbol:               "USDC",
			Uri:                  "",
			SellerFeeBasisPoints: 0,
			TokenStandard:        lo.ToPtr(ProgramMetaplexTokenMetadata.TokenStandardFungible),
		},
		TokenStandard: ProgramMetaplexTokenMetadata.TokenStandardFungible,
	}
	USDT = Token.Token{
		Address: usdtAddress,
		Info: Token.Info{
			Supply:   decimal.Zero,
			Decimals: 6,
			Decimalx: 6,
		},
		Metadata: &Token.Metadata{
			Name:                 "USDT",
			Symbol:               "USDT",
			Uri:                  "",
			SellerFeeBasisPoints: 0,
			TokenStandard:        lo.ToPtr(ProgramMetaplexTokenMetadata.TokenStandardFungible),
		},
		TokenStandard: ProgramMetaplexTokenMetadata.TokenStandardFungible,
	}
)
