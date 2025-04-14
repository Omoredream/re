package tests

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/samber/lo"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/compute_budget"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/token2022"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
)

func TestCloseTokenAccount(t *testing.T) {
	err := testCloseTokenAccount(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testCloseTokenAccount(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	wallet := mainWallet

	now := gtime.Now()

	for {
		ctx := ctx

		ixs := make([]solana.Instruction, 0)

		cuLimit := InstructionComputeBudget.SetLimit{
			Limit: 150,
		}
		err = cuLimit.AppendIx(&ixs)
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}

		for tokenAddress, tokenAccount := range wallet.Account.Tokens {
			var token Token.Token
			token, err = officialPool.TokenCacheGet(ctx, tokenAddress)
			if err != nil {
				err = gerror.Wrapf(err, "获取代币失败")
				return
			}

			var recentTxs []*rpc.TransactionSignature
			recentTxs, err = officialPool.GetAddressTransactions(ctx, tokenAccount.Address.AccountAddress, lo.ToPtr(1))
			if err != nil {
				err = gerror.Wrapf(err, "获取最近交易失败")
				return
			}

			if !tokenAccount.Token.IsZero() && (len(recentTxs) > 0 && now.Time.Sub(recentTxs[0].BlockTime.Time()) <= 3*gtime.D) {
				g.Log().Infof(ctx, "有余额 且 有近期活动, 不可销户: %s", token.DisplayName())
				delete(wallet.Account.Tokens, tokenAddress)
				continue
			} else if !tokenAccount.Token.IsZero() {
				g.Log().Infof(ctx, "有余额 但 无近期活动, 需手动销户: %s", token.DisplayName())
				delete(wallet.Account.Tokens, tokenAddress)
				continue
			} else if len(recentTxs) > 0 && now.Time.Sub(recentTxs[0].BlockTime.Time()) <= 3*gtime.D {
				g.Log().Infof(ctx, "无余额 但 有近期活动, 不可销户: %s", token.DisplayName())
				delete(wallet.Account.Tokens, tokenAddress)
				continue
			}

			err = InstructionToken.CloseAccount{
				TokenAccount: tokenAccount.Address,
				Owner:        wallet.Account.Address,
				Beneficiary:  wallet.Account.Address,
			}.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}
			cuLimit.Limit += 2916

			var tx *solana.Transaction
			tx, err = officialPool.SignInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
			if err != nil {
				err = gerror.Wrapf(err, "签名交易失败")
				return
			}

			var txRaw []byte
			txRaw, err = utils.SerializeTransaction(ctx, tx, false)
			if err != nil {
				err = gerror.Wrapf(err, "编码交易失败")
				return
			}

			if len(txRaw) > consts.MaxTxSize {
				ixs = ixs[:len(ixs)-1]
				cuLimit.Limit -= 2916
				break
			}

			delete(wallet.Account.Tokens, tokenAddress)
			g.Log().Infof(ctx, "可销户: %s", token.DisplayName())
		}

		for nftAddress, nftAccount := range wallet.Account.NFTs {
			var token Token.Token
			token, err = officialPool.TokenCacheGet(ctx, nftAddress)
			if err != nil {
				err = gerror.Wrapf(err, "获取代币失败")
				return
			}

			var recentTxs []*rpc.TransactionSignature
			recentTxs, err = officialPool.GetAddressTransactions(ctx, nftAccount.Address.AccountAddress, lo.ToPtr(1))
			if err != nil {
				err = gerror.Wrapf(err, "获取最近交易失败")
				return
			}

			if !nftAccount.Token.IsZero() && (len(recentTxs) > 0 && now.Time.Sub(recentTxs[0].BlockTime.Time()) <= 3*gtime.D) {
				g.Log().Infof(ctx, "有余额 且 有近期活动, 不可销户: %s", token.DisplayName())
				delete(wallet.Account.NFTs, nftAddress)
				continue
			} else if !nftAccount.Token.IsZero() {
				g.Log().Infof(ctx, "有余额 但 无近期活动, 需手动销户: %s", token.DisplayName())
				delete(wallet.Account.NFTs, nftAddress)
				continue
			} else if len(recentTxs) > 0 && now.Time.Sub(recentTxs[0].BlockTime.Time()) <= 3*gtime.D {
				g.Log().Infof(ctx, "无余额 但 有近期活动, 不可销户: %s", token.DisplayName())
				delete(wallet.Account.NFTs, nftAddress)
				continue
			}

			err = InstructionToken.CloseAccount{
				TokenAccount: nftAccount.Address,
				Owner:        wallet.Account.Address,
				Beneficiary:  wallet.Account.Address,
			}.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}
			cuLimit.Limit += 2916

			var tx *solana.Transaction
			tx, err = officialPool.SignInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
			if err != nil {
				err = gerror.Wrapf(err, "签名交易失败")
				return
			}

			var txRaw []byte
			txRaw, err = utils.SerializeTransaction(ctx, tx, false)
			if err != nil {
				err = gerror.Wrapf(err, "编码交易失败")
				return
			}

			if len(txRaw) > consts.MaxTxSize {
				ixs = ixs[:len(ixs)-1]
				cuLimit.Limit -= 2916
				break
			}

			delete(wallet.Account.NFTs, nftAddress)
			g.Log().Infof(ctx, "可销户: %s", token.DisplayName())
		}

		for tokenAddress, tokenAccount := range wallet.Account.Tokens2022 {
			var token Token.Token
			token, err = officialPool.TokenCacheGet(ctx, tokenAddress)
			if err != nil {
				err = gerror.Wrapf(err, "获取代币失败")
				return
			}

			var recentTxs []*rpc.TransactionSignature
			recentTxs, err = officialPool.GetAddressTransactions(ctx, tokenAccount.Address.AccountAddress, lo.ToPtr(1))
			if err != nil {
				err = gerror.Wrapf(err, "获取最近交易失败")
				return
			}

			if !tokenAccount.Token.IsZero() && (len(recentTxs) > 0 && now.Time.Sub(recentTxs[0].BlockTime.Time()) <= 3*gtime.D) {
				g.Log().Infof(ctx, "有余额 且 有近期活动, 不可销户: %s", token.DisplayName())
				delete(wallet.Account.Tokens2022, tokenAddress)
				continue
			} else if !tokenAccount.Token.IsZero() {
				g.Log().Infof(ctx, "有余额 但 无近期活动, 需手动销户: %s", token.DisplayName())
				delete(wallet.Account.Tokens2022, tokenAddress)
				continue
			} else if len(recentTxs) > 0 && now.Time.Sub(recentTxs[0].BlockTime.Time()) <= 3*gtime.D {
				g.Log().Infof(ctx, "无余额 但 有近期活动, 不可销户: %s", token.DisplayName())
				delete(wallet.Account.Tokens2022, tokenAddress)
				continue
			}

			err = InstructionToken2022.CloseAccount{
				TokenAccount: tokenAccount.Address,
				Owner:        wallet.Account.Address,
				Beneficiary:  wallet.Account.Address,
			}.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}
			cuLimit.Limit += 1332

			var tx *solana.Transaction
			tx, err = officialPool.SignInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
			if err != nil {
				err = gerror.Wrapf(err, "签名交易失败")
				return
			}

			var txRaw []byte
			txRaw, err = utils.SerializeTransaction(ctx, tx, false)
			if err != nil {
				err = gerror.Wrapf(err, "编码交易失败")
				return
			}

			if len(txRaw) > consts.MaxTxSize {
				ixs = ixs[:len(ixs)-1]
				cuLimit.Limit -= 1332
				break
			}

			delete(wallet.Account.Tokens2022, tokenAddress)
			g.Log().Infof(ctx, "可销户: %s", token.DisplayName())
		}

		for nftAddress, nftAccount := range wallet.Account.NFTs2022 {
			var token Token.Token
			token, err = officialPool.TokenCacheGet(ctx, nftAddress)
			if err != nil {
				err = gerror.Wrapf(err, "获取代币失败")
				return
			}

			var recentTxs []*rpc.TransactionSignature
			recentTxs, err = officialPool.GetAddressTransactions(ctx, nftAccount.Address.AccountAddress, lo.ToPtr(1))
			if err != nil {
				err = gerror.Wrapf(err, "获取最近交易失败")
				return
			}

			if !nftAccount.Token.IsZero() && (len(recentTxs) > 0 && now.Time.Sub(recentTxs[0].BlockTime.Time()) <= 3*gtime.D) {
				g.Log().Infof(ctx, "有余额 且 有近期活动, 不可销户: %s", token.DisplayName())
				delete(wallet.Account.NFTs2022, nftAddress)
				continue
			} else if !nftAccount.Token.IsZero() {
				g.Log().Infof(ctx, "有余额 但 无近期活动, 需手动销户: %s", token.DisplayName())
				delete(wallet.Account.NFTs2022, nftAddress)
				continue
			} else if len(recentTxs) > 0 && now.Time.Sub(recentTxs[0].BlockTime.Time()) <= 3*gtime.D {
				g.Log().Infof(ctx, "无余额 但 有近期活动, 不可销户: %s", token.DisplayName())
				delete(wallet.Account.NFTs2022, nftAddress)
				continue
			}

			err = InstructionToken2022.CloseAccount{
				TokenAccount: nftAccount.Address,
				Owner:        wallet.Account.Address,
				Beneficiary:  wallet.Account.Address,
			}.AppendIx(&ixs)
			if err != nil {
				err = gerror.Wrapf(err, "构建交易失败")
				return
			}
			cuLimit.Limit += 1332

			var tx *solana.Transaction
			tx, err = officialPool.SignInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
			if err != nil {
				err = gerror.Wrapf(err, "签名交易失败")
				return
			}

			var txRaw []byte
			txRaw, err = utils.SerializeTransaction(ctx, tx, false)
			if err != nil {
				err = gerror.Wrapf(err, "编码交易失败")
				return
			}

			if len(txRaw) > consts.MaxTxSize {
				ixs = ixs[:len(ixs)-1]
				cuLimit.Limit -= 1332
				break
			}

			delete(wallet.Account.Tokens2022, nftAddress)
			g.Log().Infof(ctx, "可销户: %s", token.DisplayName())
		}

		if len(ixs) == 1 {
			break
		}
		ixs[0], err = cuLimit.ToIx()
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}

		var tx *solana.Transaction
		tx, err = officialPool.SignInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
		if err != nil {
			err = gerror.Wrapf(err, "签名交易失败")
			return
		}

		var txRaw string
		txRaw, err = utils.SerializeTransactionBase64(ctx, tx, true)
		if err != nil {
			err = gerror.Wrapf(err, "编码交易失败")
			return
		}
		g.Log().Infof(ctx, "构造交易 %s", txRaw)

		_, _, err = officialPool.SimulateTransaction(ctx, tx)
		if err != nil {
			err = gerror.Wrapf(err, "模拟交易失败")
			return
		}

		var txHash solana.Signature
		txHash, err = officialPool.SendTransaction(ctx, tx)
		if err != nil {
			err = gerror.Wrapf(err, "发送交易失败")
			return
		}
		ctx = context.WithValue(ctx, consts.CtxTransaction, txHash.String())
		g.Log().Infof(ctx, "已发送交易")

		var spent time.Duration
		spent, err = officialPool.WaitConfirmTransactionByHTTP(ctx, txHash, tx)
		if err != nil {
			err = gerror.Wrapf(err, "等待交易确认失败")
			return
		}
		g.Log().Infof(ctx, "交易耗时 %s", spent)
	}

	return
}
