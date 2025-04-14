package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gagliardetto/binary"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/shopspring/decimal"

	"github.com/gagliardetto/solana-go"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/instruction"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/wallet"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/decimals"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/utils/lamports"
)

type Pump struct {
	VirtualReservesToken uint64
	VirtualReservesSol   uint64
	RealReservesToken    uint64
	RealReservesSol      uint64
	TotalSupply          uint64
	Complete             bool
}

func TestPumpBuy(t *testing.T) {
	var err error

	tokenAddress := Address.NewFromBase58("B1TD2HX7JZSXyWRJUjGBhqYm9zwMV9ECjHuZFYZFm7z4").AsTokenAddress()

	token, err := officialPool.TokenCacheGet(ctx, tokenAddress)
	if err != nil {
		g.Log().Fatalf(ctx, "获取代币信息失败, %+v", err)
	}

	request, err := g.Client().SetHeaderMap(map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
		"Origin":     "https://www.pump.fun",
		"Referer":    "https://www.pump.fun/",
	}).Get(ctx, fmt.Sprintf("https://client-api-2-74b1891ee9f9.herokuapp.com/coins/%s", tokenAddress.String()))
	if err != nil {
		g.Log().Fatalf(ctx, "获取 Pump 代币信息失败, %+v", err)
	}
	if request.StatusCode != http.StatusOK {
		err = gerror.Newf("HTTP %d, %s", request.StatusCode, request.Status)
		g.Log().Fatalf(ctx, "获取 Pump 代币信息失败, %+v", err)
	}
	response, err := gjson.DecodeToJson(request.ReadAll())
	if err != nil {
		g.Log().Fatalf(ctx, "解析 Pump 代币信息失败, %+v", err)
	}
	bondingCurve := Address.NewFromBase58(response.Get("bonding_curve").String())
	associatedBondingCurve := Address.NewFromBase58(response.Get("associated_bonding_curve").String())

	wallet, err := officialPool.NewWalletFromWIF(ctx, testWalletWIF)
	if err != nil {
		g.Log().Fatalf(ctx, "导入钱包失败, %+v", err)
	}

	tokenAccount, err := wallet.Account.Address.FindAssociatedTokenAccountAddress(tokenAddress)
	if err != nil {
		g.Log().Fatalf(ctx, "寻找代币账户地址失败, %+v", err)
	}

	getAccountInfoResult, err := officialPool.GetAccountInfo(ctx, bondingCurve)
	if err != nil {
		g.Log().Fatalf(ctx, "获取 Pump 代币价格失败, %+v", err)
	}

	buySolWanna := decimal.NewFromFloat(0.1)

	var pump Pump
	err = bin.NewBinDecoder(getAccountInfoResult[8:]).Decode(&pump)
	if err != nil {
		g.Log().Fatalf(ctx, "解析余额信息失败, %+v", err)
	}
	virtualReservesToken := lamports.Lamports2Token(pump.VirtualReservesToken, token.Info.Decimalx)
	virtualReservesSol := lamports.Lamports2SOL(pump.VirtualReservesSol)

	virtualReservesSolAfterBuy := virtualReservesSol.Add(buySolWanna)
	virtualReservesTokenAfterBuy := virtualReservesToken.Mul(virtualReservesSol).Div(virtualReservesSolAfterBuy).Round(token.Info.Decimalx)
	buyToken := virtualReservesToken.Sub(virtualReservesTokenAfterBuy)
	g.Log().Infof(ctx, "%s SOL => %s %s", decimals.DisplayBalance(buySolWanna), decimals.DisplayBalance(buyToken), token.DisplayName())

	ixs := make([]solana.Instruction, 0)

	err = Instruction.Custom{
		ProgramID: Address.NewFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ1d4M5uBEwF6P").AsProgramAddress(),
		Accounts: []*solana.AccountMeta{
			Address.NewFromBase58("4wTV1YmiEkRvAtNtsSGPtUrqRYQMe5SKy2uB4Jjaxnjf").Meta(),
			Address.NewFromBase58("CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM").Meta().WRITE(),
			tokenAddress.Meta(),
			bondingCurve.Meta().WRITE(),
			associatedBondingCurve.Meta().WRITE(),
			tokenAccount.Meta().WRITE(),
			wallet.Account.Address.Meta().SIGNER().WRITE(),
			consts.SystemProgramAddress.Meta(),
			consts.TokenProgramAddress.Meta(),
			consts.SysVarRentAddress.Meta(),
			Address.NewFromBase58("Ce6TQqeHC9p8KetsN6JsjHK7UTZk7nasjjnr7XxXp9F1").Meta(),
			Address.NewFromBase58("6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P").AsProgramAddress().Meta(),
		},
		Discriminator: Utils.Uint64ToBytesL(16927863322537952870), // 买
		//Discriminator: Utils.Uint64ToBytesL(12502976635542562355), // 卖
		Data: Utils.Append(
			Utils.Uint64ToBytesL(lamports.Token2Lamports(buyToken, token.Info.Decimalx)),
			Utils.Uint64ToBytesL(lamports.SOL2Lamports(buySolWanna.Mul(decimal.NewFromFloat(1.05)))), // 5% 滑点
		),
	}.AppendIx(&ixs)
	if err != nil {
		err = gerror.Wrapf(err, "构建交易失败")
		return
	}

	txHash, err := officialPool.SendInstructions(ctx, ixs, []Wallet.HostedWallet{wallet}, wallet)
	if err != nil {
		g.Log().Fatalf(ctx, "发送交易失败, %+v", err)
	}

	g.Log().Infof(ctx, "交易ID: %s", txHash)
}
