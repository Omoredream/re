package tests

import (
	"testing"

	"github.com/gagliardetto/binary"
	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gregex"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/worth"
)

type Proof struct {
	Authority    solana.PublicKey
	Balance      bin.Uint64
	Challenge    [32]byte
	LastHash     [32]byte
	LastHashAt   bin.Int64
	LastStakeAt  bin.Int64
	Miner        solana.PublicKey
	TotalHashes  bin.Uint64
	TotalRewards bin.Uint64
}

type ProofResult struct {
	Difficulty bin.Uint64
	Reward     bin.Uint64
	Timing     bin.Int64
}

var OreV2Program = Address.NewFromBase58("oreV2ZymfyeXgNgBdqMkumTqqAprVqgBWQfoYkrtKWQ").AsProgramAddress()
var OreV2Token = Address.NewFromBase58("oreoU2P8bN6jkk3jbaiVxYnG1dCXcYxwhwyK9jSybcp").AsTokenAddress()

func TestOreV2Balance(t *testing.T) {
	var err error

	wallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		g.Log().Fatalf(ctx, "导入钱包失败, %+v", err)
	}

	//wallet, err := officialPool.NewWalletFromMnemonic(ctx, testWalletMnemonic, 0)
	//if err != nil {
	//	g.Log().Fatalf(ctx, "导入钱包失败, %+v", err)
	//}

	//wallet, err := officialPool.NewWalletFromMnemonic(ctx, testWalletMnemonic, 1)
	//if err != nil {
	//	g.Log().Fatalf(ctx, "导入钱包失败, %+v", err)
	//}

	g.Log().Infof(ctx, "钱包: %s", wallet.Account.Address)
	g.Log().Infof(ctx, "|- SOL: %s ($%s)", decimals.DisplayBalance(wallet.Account.SOL), decimals.DisplayBalance(worth.IgnoreErr(worth.SOLWorth(ctx, wallet.Account.SOL))))
	for tokenAddress, tokenBalance := range wallet.Account.Tokens {
		if tokenAddress != OreV2Token {
			continue
		}
		var token Token.Token
		token, err = officialPool.TokenCacheGet(ctx, tokenAddress)
		if err != nil {
			g.Log().Fatalf(ctx, "%s", gerror.Wrapf(err, "获取钱包 %s 代币 %s 失败", wallet.Account.Address, tokenAddress))
		}

		g.Log().Infof(ctx, "|- %s: %s", token.DisplayName(), decimals.DisplayBalance(tokenBalance.Token))
	}

	proofAddress, _, err := OreV2Program.FindProgramDerivedAddress([][]byte{[]byte("proof"), wallet.Account.Address.Bytes()})
	if err != nil {
		g.Log().Fatalf(ctx, "生成挖矿账户失败, %+v", err)
	}

	proofData, err := officialPool.GetAccountInfo(ctx, proofAddress)
	if err != nil {
		g.Log().Fatalf(ctx, "查询挖矿账户数据失败, %+v", err)
	}

	if proofData[0] != 102 {
		g.Log().Fatalf(ctx, "挖矿账户数据标识非预期, %d", proofData[0])
	}

	var proof Proof
	err = bin.NewBinDecoder(proofData[8:]).Decode(&proof)
	if err != nil {
		g.Log().Fatalf(ctx, "解析挖矿账户数据失败, %+v", err)
	}

	g.Log().Infof(ctx,
		"质押: %s ORE, 最后挖矿: %s, 最后质押: %s, 已挖 %d 次, 挖出 %s ORE, 平均 %s ORE/次",
		decimals.DisplayBalance(lamports.Lamports2Token(uint64(proof.Balance), 11)),
		gtime.Now().Sub(gtime.NewFromTimeStamp(int64(proof.LastHashAt))),
		gtime.Now().Sub(gtime.NewFromTimeStamp(int64(proof.LastStakeAt))),
		proof.TotalHashes,
		decimals.DisplayBalance(lamports.Lamports2Token(uint64(proof.TotalRewards), 11)),
		decimals.DisplayBalance(lamports.Lamports2Token(uint64(proof.TotalRewards), 11).Div(decimal.NewFromInt(max(int64(proof.TotalHashes), 1)))),
	)

	lastTransactions, err := officialPool.GetAddressTransactions(ctx, proofAddress, lo.ToPtr(10))
	if err != nil {
		g.Log().Fatalf(ctx, "查询挖矿账户最近交易失败, %+v", err)
	}

	for _, lastTransaction := range lastTransactions {
		transaction, err := officialPool.GetTransaction(ctx, lastTransaction.Signature)
		if err != nil {
			g.Log().Fatalf(ctx, "查询挖矿账户最近交易内容失败, %+v", err)
		}

		if len(transaction.Meta.LogMessages) >= 2 {
			result, err := gregex.MatchString(`^Program return: oreV2ZymfyeXgNgBdqMkumTqqAprVqgBWQfoYkrtKWQ ([A-Za-z0-9+/]+?)$`, transaction.Meta.LogMessages[len(transaction.Meta.LogMessages)-2])
			if err != nil {
				g.Log().Fatalf(ctx, "解析挖矿账户最近交易结果失败, %+v", err)
			}

			if len(result) == 2 {
				var proofResult ProofResult
				err = bin.NewBinDecoder(gbase64.MustDecodeString(result[1])).Decode(&proofResult)
				if err != nil {
					g.Log().Fatalf(ctx, "解析挖矿账户数据失败, %+v", err)
				}

				fee := lamports.Lamports2SOL(transaction.Meta.Fee)
				reward := lamports.Lamports2Token(uint64(proofResult.Reward), 11)
				var minPrice decimal.Decimal
				if reward.Sign() > 0 {
					minPrice = fee.Div(reward)
				}
				g.Log().Infof(ctx,
					"%s 前, 优先费 %s SOL, 挖出 %s ORE, 关机价 %s SOL, 难度 %d, 超时 %ds",
					gtime.Now().Sub(gtime.NewFromTime(lastTransaction.BlockTime.Time())),
					decimals.DisplayBalance(fee),
					decimals.DisplayBalance(reward),
					decimals.DisplayBalance(minPrice),
					proofResult.Difficulty,
					proofResult.Timing,
				)
			}
		}
	}
}
