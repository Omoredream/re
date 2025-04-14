package utils

import (
	"fmt"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/shopspring/decimal"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/token"
)

func FindOpenBookAssociatedAuthorityAddress(programAddress Address.ProgramAddress, marketAddress Address.AccountAddress) (authorityAddress Address.AccountAddress, err error) {
	seeds := [][]byte{marketAddress.Bytes()}

	for nonce := byte(0); nonce < 100; nonce++ {
		authorityAddress, err = programAddress.CreateProgramAddress(append(seeds, []byte{nonce}, make([]byte, 7)))
		if err == nil {
			return
		}
	}

	err = gerror.Newf("unable to find a valid program address")

	return
}

type RaydiumLP struct {
	Id                  Address.AccountAddress      // LP 地址
	Authority           Address.AccountAddress      // Authority 地址
	OpenOrders          Address.AccountAddress      // OpenBook 订单地址
	TargetOrders        Address.AccountAddress      // 订单地址
	BaseVault           Address.TokenAccountAddress // 左向代币金库地址
	QuoteVault          Address.TokenAccountAddress // 右向代币金库地址
	Market              Address.AccountAddress      // OpenBook LP 地址
	MarketBids          Address.AccountAddress
	MarketAsks          Address.AccountAddress
	MarketEventQueue    Address.AccountAddress
	MarketBaseVault     Address.TokenAccountAddress // 左向代币 OpenBook 金库地址
	MarketQuoteVault    Address.TokenAccountAddress // 右向代币 OpenBook 金库地址
	MarketAuthority     Address.AccountAddress      // OpenBook Authority 地址
	LPMint              Address.TokenAddress        // LP 凭证代币地址
	BaseToken           Token.Token                 // 左向代币
	QuoteToken          Token.Token                 // 右向代币
	InitialBaseBalance  decimal.Decimal             // 左向代币初始金额
	InitialQuoteBalance decimal.Decimal             // 右向代币初始金额
	InitialPrice        decimal.Decimal             // 代币初始金额比
	InitialLiquidity    decimal.Decimal             // 初始流动性
	InitialFdv          decimal.Decimal             // 初始市值
	OpenTime            *gtime.Time                 // 开盘时间
}

func (lp *RaydiumLP) String() string {
	return fmt.Sprintf("%s/%s (%s - %s)", lp.BaseToken.DisplayName(), lp.QuoteToken.DisplayName(), lp.BaseToken.String(), lp.Id.String())
}
