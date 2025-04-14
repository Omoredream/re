package tests

import (
	"testing"

	"github.com/gagliardetto/binary"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/imroc/req/v3"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction/compute_budget"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"
)

func TestOkxSell(t *testing.T) {
	const nftProject = "3843304"
	const orderId = "6428214102"
	const price = 0.013

	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	client := req.C().
		ImpersonateChrome()

	for _, nftId := range []string{} {
		resp, err := client.R().
			SetBodyJsonMarshal(map[string]any{
				"nftId":                           nftId,
				"orderId":                         orderId,
				"price":                           price,
				"project":                         nftProject,
				"seller":                          mainWallet.Account.Address.String(),
				"transferCheckedInstructionsKeys": []any{},
			}).
			Post("https://www.okx.com/priapi/v1/nft/okx/solana/instructions/takeCollectionOffer")
		if err != nil {
			g.Log().Fatalf(ctx, "%+v", err)
			return
		}

		respJ, err := gjson.DecodeToJson(resp.Bytes())
		if err != nil {
			g.Log().Fatalf(ctx, "%+v", err)
			return
		}

		tx := &solana.Transaction{}
		err = tx.UnmarshalWithDecoder(bin.NewBinDecoder(respJ.Get("data.tx.data").Bytes()))
		if err != nil {
			g.Log().Fatalf(ctx, "解析交易失败, %+v", err)
		}

		var ixs []solana.Instruction
		var ixs_ []solana.Instruction
		var alt map[solana.PublicKey]solana.PublicKeySlice
		ixs_, _, _, alt, err = officialPool.UnpackTransaction(ctx, tx)
		if err != nil {
			err = gerror.Wrapf(err, "拆解交易失败")
			return
		}

		for _, ix := range ixs_ {
			if ix.ProgramID() == consts.ComputeBudgetProgramAddress.PublicKey {
				continue
			}
			ixs = append(ixs, ix)
		}

		err = InstructionComputeBudget.SetLimit{
			Limit: 10_0000 + 150,
		}.AppendIx(&ixs)
		if err != nil {
			err = gerror.Wrapf(err, "构建交易失败")
			return
		}

		tx, err = officialPool.PackTransaction(ctx, ixs, solana.Hash{}, mainWallet.Account.Address, alt)
		if err != nil {
			err = gerror.Wrapf(err, "构造交易失败")
			return
		}

		err = officialPool.SignTransaction(ctx, tx, []Wallet.HostedWallet{mainWallet})
		if err != nil {
			err = gerror.Wrapf(err, "签名交易失败")
			return
		}

		txHash, err := officialPool.SendTransaction(ctx, tx)
		if err != nil {
			err = gerror.Wrapf(err, "发送交易失败")
			return
		}
		g.Log().Infof(ctx, "已发送交易")

		spent, err := officialPool.WaitConfirmTransactionByHTTP(ctx, txHash)
		if err != nil {
			err = gerror.Wrapf(err, "等待交易确认失败")
			return
		}
		g.Log().Infof(ctx, "交易耗时 %s", spent)
	}
}
