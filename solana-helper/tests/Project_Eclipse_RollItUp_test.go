package tests

import (
	"context"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gogf/gf/v2/encoding/gbase64"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/grand"
	"github.com/imroc/req/v3"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/tkhq/go-sdk/pkg/apikey"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/account"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

func TestEclipseRollItUp(t *testing.T) {
	err := testEclipseRollItUp(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "%+v", err)
	}
}

func testEclipseRollItUp(ctx g.Ctx) (err error) {
	appWalletAddress := Address.NewFromBase58("F3e4NJjxq89Ja6vdgZXqnC1Vbfz8GFxGMkNDa396Q5Bi")

	turnKey, err := apikey.FromTurnkeyPrivateKey("755fed40d82af9203023fbec12d67a2b533e23d40aea2f60ff0da4e95d13dab7", apikey.SchemeP256)
	if err != nil {
		err = gerror.Wrapf(err, "初始化签名钱包失败")
		return
	}

	appWalletTokenAccountAddress, err := appWalletAddress.FindAssociatedToken2022AccountAddress(Address.NewFromBase58("AKEWE7Bgh87GPp171b4cJPSSZfmZwQ3KaqYqXoKLNAEE").AsTokenAddress())
	if err != nil {
		err = gerror.Wrapf(err, "派生代理钱包代币地址失败")
		return
	}

	turnKeySign := func(data []byte, apiKey *apikey.Key, client *req.Client) (signature []byte, err error) {
		payload := gjson.MustEncode(g.MapStrAny{
			"type":           "ACTIVITY_TYPE_SIGN_RAW_PAYLOAD_V2",
			"organizationId": "6c469059-3e37-4902-a672-a798157ad838",
			"timestampMs":    gtime.TimestampMilliStr(),
			"parameters": g.MapStrAny{
				"signWith":     appWalletAddress.String(),
				"payload":      fmt.Sprintf("%x", data),
				"encoding":     "PAYLOAD_ENCODING_HEXADECIMAL",
				"hashFunction": "HASH_FUNCTION_NOT_APPLICABLE",
			},
		})

		payloadStamp, err := apikey.Stamp(payload, apiKey)
		if err != nil {
			err = gerror.Wrapf(err, "生成签名请求签名失败")
			return
		}

		resp, err := client.R().
			SetHeader("x-stamp", payloadStamp).
			SetHeader("x-client-version", "@turnkey/http@2.15.0").
			SetBodyJsonBytes(payload).
			Post("https://api.turnkey.com/public/v1/submit/sign_raw_payload")
		if err != nil {
			err = gerror.Wrapf(err, "发送签名请求失败")
			return
		}
		if resp.GetStatusCode() != 200 {
			err = gerror.Newf("服务端响应签名请求失败, %s", resp.String())
			return
		}

		respJson, err := gjson.DecodeToJson(resp.Bytes())
		if err != nil {
			err = gerror.Wrapf(err, "解析签名响应失败")
			return
		}

		signature, err = hex.DecodeString(respJson.Get("activity.result.signRawPayloadResult.r").String() + respJson.Get("activity.result.signRawPayloadResult.s").String())
		if err != nil {
			err = gerror.Wrapf(err, "解析签名失败")
			return
		}
		if len(signature) == 0 {
			err = gerror.Newf("服务端响应签名请求失败, %s", resp.String())
			return
		}

		return
	}

	client := req.C().
		SetCommonHeader("origin", "https://rollitup.fun").
		SetCommonHeader("referer", "https://rollitup.fun/").
		SetProxyURL("socks5://127.0.0.1:10801").
		ImpersonateChrome()

	for {
		var poolTokenAccount *Account.TokenAccount
		poolTokenAccount, err = officialPool.GetAccountToken(ctx, Address.NewFromBase58("Cr9Hvcymgsfz1Dx9LfYkaouACU2PYYbZyiwThj5o98hc"), Address.NewFromBase58("AKEWE7Bgh87GPp171b4cJPSSZfmZwQ3KaqYqXoKLNAEE").AsTokenAddress())
		if err != nil {
			err = gerror.Wrapf(err, "获取奖池失败")
			return
		}

		var appWalletTokenAccount *Account.TokenAccount
		appWalletTokenAccount, err = officialPool.GetAccountToken(ctx, appWalletTokenAccountAddress.AccountAddress, Address.NewFromBase58("AKEWE7Bgh87GPp171b4cJPSSZfmZwQ3KaqYqXoKLNAEE").AsTokenAddress())
		if err != nil {
			err = gerror.Wrapf(err, "获取余额失败")
			return
		}

		maxProfit := poolTokenAccount.Token.Mul(decimal.NewFromFloat(0.05))
		if maxProfit.LessThan(decimal.NewFromFloat(50)) {
			g.Log().Warningf(ctx, "池子 %.2f USDC 可供掏取的单次利润 %.2f USDC 过小", poolTokenAccount.Token.InexactFloat64(), maxProfit.InexactFloat64())
			break
		}

		var (
			predictedNumber uint8 = 94 // 与 multiplier 对应
		)
		multiplier := 15.84 // 最大 15.84
		betAmount := decimal.Min(appWalletTokenAccount.Token, maxProfit.RoundFloor(-1).Div(decimal.NewFromFloat(multiplier))).InexactFloat64()
		g.Log().Infof(ctx, "最大利润 %.2f USDC, 倍率 x%.2f, 下注数量 %.2f USDC", maxProfit.InexactFloat64(), multiplier, betAmount)

		generateTxAuth := fmt.Sprintf("%s:%s", gtime.TimestampMilliStr(), strings.Join(lo.Map(grand.B(32), func(item uint8, index int) string {
			return strconv.Itoa(int(item))
		}), ""))

		var generateTxAuthSignature []byte
		generateTxAuthSignature, err = turnKeySign([]byte(generateTxAuth), turnKey, client)
		if err != nil {
			err = gerror.Wrapf(err, "签名生成交易令牌失败")
			return
		}

		var generateTxResp *req.Response
		generateTxResp, err = client.R().
			SetHeader("x-auth-nonce", gbase64.EncodeToString(generateTxAuthSignature)).
			SetHeader("x-auth-data", generateTxAuth).
			SetHeader("x-auth-id", appWalletAddress.String()).
			SetBodyJsonMarshal(g.MapStrAny{
				"params": g.MapStrAny{
					"amount":   lamports.Token2Lamports(decimal.NewFromFloat(betAmount), 6),
					"ratio":    multiplier,
					"selector": predictedNumber,
					"account":  appWalletAddress.String(),
					"tokenId":  appWalletTokenAccountAddress.String(),
				},
			}).
			Post("https://rollitup.fun/api/rpc")
		if err != nil {
			err = gerror.Wrapf(err, "发送生成交易请求失败")
			return
		}
		if generateTxResp.GetStatusCode() != 200 {
			err = gerror.Newf("服务端响应生成交易请求失败, %s", generateTxResp.String())
			return
		}
		var generateTxRespJson *gjson.Json
		generateTxRespJson, err = gjson.DecodeToJson(generateTxResp.Bytes())
		if err != nil {
			err = gerror.Wrapf(err, "解析生成交易响应失败")
			return
		}

		generatedTx := generateTxRespJson.Get("result").String()
		if len(generatedTx) == 0 {
			err = gerror.Newf("服务端响应生成交易请求失败, %s", generateTxResp.String())
			return
		}

		var tx *solana.Transaction
		tx, err = utils.DeserializeTransactionBase64(generatedTx)
		if err != nil {
			err = gerror.Wrapf(err, "解析交易失败")
			return
		}

		var (
			ixs []solana.Instruction
		)
		ixs, _, _, _, err = officialPool.UnpackTransaction(ctx, tx)
		if err != nil {
			err = gerror.Wrapf(err, "拆解交易失败")
			return
		}

		diceNumber := lo.Must(ixs[0].Data())[16]
		diceRoll := lo.Must(ixs[0].Data())[17]
		g.Log().Noticef(ctx, "赢钱线: %d, 掷出: %d", diceNumber, diceRoll)

		if predictedNumber != diceNumber {
			err = gerror.Newf("预期赢钱线 %d 与实际 %d 不同", predictedNumber, diceNumber)
			return
		}
		if diceNumber > diceRoll {
			continue
		}

		var messageContent []byte
		messageContent, err = tx.Message.MarshalBinary()
		if err != nil {
			err = gerror.Wrapf(err, "编码待签名的交易失败")
			return
		}

		var signature []byte
		signature, err = turnKeySign(messageContent, turnKey, client)
		if err != nil {
			err = gerror.Wrapf(err, "签名交易失败")
			return
		}

		tx.Signatures[0] = solana.Signature(signature)

		var txRaw string
		txRaw, err = utils.SerializeTransactionBase64(ctx, tx, false)
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

		var submitResp *req.Response
		submitResp, err = client.R().
			SetBodyJsonMarshal(g.MapStrAny{
				"betAmount":       betAmount,
				"multiplier":      multiplier,
				"predictedNumber": predictedNumber,
				"userWallet":      appWalletAddress.String(),
				"won":             true,
				"actualNumber":    diceRoll,
			}).
			Post("https://rollitup.fun/api/bet")
		if err != nil {
			err = gerror.Wrapf(err, "发送提交结果请求失败")
			return
		}
		if submitResp.GetStatusCode() != 200 {
			err = gerror.Newf("服务端响应提交结果请求失败, %s", submitResp.String())
			return
		}
		var submitRespJson *gjson.Json
		submitRespJson, err = gjson.DecodeToJson(submitResp.Bytes())
		if err != nil {
			err = gerror.Wrapf(err, "解析提交结果响应失败")
			return
		}

		g.Log().Noticef(ctx, "已下注 %d 次, 下注量 %.2f USDC, 总利润 %.2f USDC, 余额 %.2f USDC", submitRespJson.Get("user.number_of_bets").Int(), submitRespJson.Get("user.volume").Float64(), submitRespJson.Get("user.profit").Float64(), submitRespJson.Get("user.balance").Float64())
	}

	return
}
