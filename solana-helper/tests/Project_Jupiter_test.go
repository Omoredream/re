package tests

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/rpc"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func TestJupiter(t *testing.T) {
	err := testJupiter(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testJupiter(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	for _, tokenAddress := range []Address.TokenAddress{
		Address.NewFromBase58("DezXAZ8z7PnrnRJjz3wXBoRgixCa6xjnB7YaB1pPB263").AsTokenAddress(),
		Address.NewFromBase58("EKpQGSJtjMFqKZ9KQanSqYXRcF8fBopzLHYxdM65zcjm").AsTokenAddress(),
		Address.NewFromBase58("7GCihgDB8fe6KNjn2MYtkzZcRjQy3t9GHdC8uHYmW2hr").AsTokenAddress(),
	} {
		for {
			var token Token.Token
			token, err = officialPool.TokenCacheGet(ctx, tokenAddress)
			if err != nil {
				err = gerror.Wrapf(err, "查询代币失败")
				return
			}

			var quote jupiterHTTP.GetQuoteResponse
			quote, err = jupiterPool.GetQuote(ctx, tokenAddress, consts.SOL.Address, lamports.Token2Lamports(wallet.Account.Tokens[tokenAddress].Token, token.Info.Decimalx))
			if err != nil {
				err = gerror.Wrapf(err, "获取报价失败")
				return
			}

			g.Log().Infof(
				ctx,
				"%s %s => %s %s",
				token.DisplayName(), decimals.DisplayBalance(wallet.Account.Tokens[tokenAddress].Token),
				consts.SOL.DisplayName(), decimals.DisplayBalance(lamports.Lamports2SOL(quote.OutAmount)),
			)

			var swapTx *solana.Transaction
			swapTx, err = jupiterPool.CreateSwapTransaction(ctx, wallet.Account.Address, quote)
			if err != nil {
				err = gerror.Wrapf(err, "生成交易失败")
				return
			}

			err = officialPool.SignTransaction(ctx, swapTx, []Wallet.HostedWallet{wallet})
			if err != nil {
				err = gerror.Wrapf(err, "签名交易失败")
				return
			}

			var txHash solana.Signature
			txHash, err = officialPool.SendTransaction(ctx, swapTx)
			if err != nil {
				err = gerror.Wrapf(err, "发送交易失败")
				return
			}
			ctx := context.WithValue(ctx, consts.CtxTransaction, txHash.String())
			g.Log().Infof(ctx, "已发送交易")

			var spent time.Duration
			spent, err = officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
			if err != nil {
				err = gerror.Wrapf(err, "等待交易确认失败")
				g.Log().Errorf(ctx, "%v", err)
				continue
			}
			g.Log().Infof(ctx, "交易耗时 %s", spent)
			break
		}
	}

	return
}
