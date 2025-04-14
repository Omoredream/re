package tests

import (
	"context"
	"io"
	"math/rand"
	"net/http"
	"testing"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/imroc/req/v3"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
)

type ActionsProof struct {
	Transaction string  `json:"transaction"`
	Message     *string `json:"message"`
}

func TestBlinks(t *testing.T) {
	err := testBlinks(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testBlinks(ctx g.Ctx) (err error) {
	mainWallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		err = gerror.Wrapf(err, "导入钱包失败")
		return
	}

	var walletCount uint32 = 500

	var fleetWallets []Wallet.HostedWallet
	fleetWallets = append(fleetWallets, mainWallet)

	fleetWallets = Utils.Grow(fleetWallets, len(fleetWallets)+int(walletCount))
	for i := uint32(0); i < walletCount; i++ {
		ctx := context.WithValue(ctx, consts.CtxDerivation, i)

		var fleetWallet Wallet.HostedWallet
		fleetWallet, err = officialPool.NewWalletFromMnemonic(ctx, testWalletMnemonic, i)
		if err != nil {
			err = gerror.Wrapf(err, "派生钱包失败")
			return
		}

		ctx = context.WithValue(ctx, consts.CtxAddress, fleetWallet.Account.Address.String())

		g.Log().Infof(ctx, "SOL %s", decimals.DisplayBalance(fleetWallet.Account.SOL))

		if fleetWallet.Account.SOL.IsZero() { // 空钱包不参与
			continue
		}

		fleetWallets = append(fleetWallets, fleetWallet)
	}

	Utils.Parallel(ctx, 256, 0, func(_, _ int) (err error) {
		client := req.C().
			ImpersonateChrome().
			SetCommonHeaders(g.MapStrStr{
				"Origin":  "https://x.com",
				"Referer": "https://x.com/",
			})

		for {
			wallet := fleetWallets[rand.Intn(len(fleetWallets))]

			var resp *req.Response
			resp, err = client.R().
				SetBodyJsonMarshal(g.MapStrAny{
					"account": wallet.Account.Address.String(),
				}).
				Post("https://solanasummer.click/on/mint")
			//Post("https://api-mainnet.magiceden.dev/actions/mint-launchpad/class_of_2021_3_year_reunion_oe")
			//Post("https://proxy.dial.to/?url=https%3A%2F%2Fsolanasummer.click%2Fon%2Fmint")
			if err != nil {
				if gerror.Is(err, io.EOF) || gerror.Is(err, io.ErrUnexpectedEOF) {
					continue
				}
				err = gerror.Wrapf(err, "发送请求失败")
				g.Log().Errorf(ctx, "%v", err)
				continue
			}
			if resp.StatusCode != http.StatusOK {
				if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusForbidden {
					err = gerror.Newf("HTTP/%d, %s", resp.StatusCode, resp.Status)
					g.Log().Warningf(ctx, "%v", err)
				}
				continue
			}

			var data ActionsProof
			err = gjson.DecodeTo(resp.Bytes(), &data)
			if err != nil {
				err = gerror.Wrapf(err, "解析响应失败")
				g.Log().Errorf(ctx, "%v", err)
				continue
			}

			g.Log().Debugf(ctx, "未签名交易: %s", data.Transaction)

			if data.Message != nil && *data.Message != "Blink and see it in your wallet!" {
				err = gerror.Newf("返回信息异常: %s", *data.Message)
				return
			}

			tx := &solana.Transaction{}
			err = tx.UnmarshalBase64(data.Transaction)
			if err != nil {
				err = gerror.Wrapf(err, "解析交易失败")
				g.Log().Errorf(ctx, "%+v", err)
				continue
			}

			var writable bool
			writable, err = tx.Message.IsWritable(wallet.Account.Address.PublicKey)
			if err != nil {
				err = gerror.Wrapf(err, "解析账户失败")
				g.Log().Errorf(ctx, "%+v", err)
				continue
			}
			if writable {
				err = gerror.Newf("交易可写, 可能存在问题")
				return
			}

			err = officialPool.SignTransaction(ctx, tx, []Wallet.HostedWallet{wallet})
			if err != nil {
				err = gerror.Wrapf(err, "签名交易失败")
				return
			}

			var txHash solana.Signature
			txHash, err = officialPool.SendTransaction(ctx, tx)
			if err != nil {
				err = gerror.Wrapf(err, "发送交易失败")
				g.Log().Errorf(ctx, "%+v", err)
				continue
			}
			g.Log().Infof(ctx, "交易ID: %s", txHash)
		}
	})

	return
}
