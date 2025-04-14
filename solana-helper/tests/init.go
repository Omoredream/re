package tests

import (
	_ "git.wkr.moe/web3/solana-helper"

	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jito/rpcs"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/rpcs"
	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/official/rpcs"
)

var (
	ctx          g.Ctx
	officialPool officialRPCs.RPCs
	jupiterPool  jupiterRPCs.RPCs
	jitoPool     jitoRPCs.RPCs
)

func init() {
	var err error

	ctx = context.Background()

	// 日志

	logConfig := g.Log().GetConfig()
	logConfig.CtxKeys = []any{
		consts.CtxRPC,
		consts.CtxWallet,
		consts.CtxAccount,
		consts.CtxAddress,
		consts.CtxDerivation,
		consts.CtxToken,
		consts.CtxTransaction,
		consts.CtxBundle,
	}
	logConfig.Level = glog.LEVEL_ALL & ^glog.LEVEL_DEBU
	logConfig.StStatus = 0
	err = g.Log().SetConfig(logConfig)
	if err != nil {
		g.Log().Fatal(ctx, err)
	}

	err = g.Log("retry").SetConfig(logConfig)
	if err != nil {
		g.Log().Fatal(ctx, err)
	}
	g.Log("retry").SetStdoutPrint(false)

	err = g.Log("scheduler").SetConfig(logConfig)
	if err != nil {
		g.Log().Fatal(ctx, err)
	}

	// Official RPC

	officialPool = officialRPCs.New()

	officialHttpRPCs, err := getOfficialHttpRPCs(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "批量初始化 HTTP RPC 失败, %+v", err)
	}
	officialPool.AddHttpRPC(officialHttpRPCs...)

	officialWebsocketRPCs, err := getOfficialWebsocketRPCs(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "批量初始化 WebSocket RPC 失败, %+v", err)
	}
	officialPool.AddWebsocketRPC(officialWebsocketRPCs...)

	// Jupiter Swap API

	jupiterPool = jupiterRPCs.New()

	jupiterHttpRPCs, err := getJupiterHttpRPCs(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "批量初始化 Jupiter Swap API 失败, %+v", err)
	}
	jupiterPool.AddRPC(jupiterHttpRPCs...)

	// Jito JSON-RPC

	jitoPool = jitoRPCs.New()

	jitoHttpRPCs, err := getJitoHttpRPCs(ctx)
	if err != nil {
		g.Log().Fatalf(ctx, "批量初始化 Jito JSON-RPC 失败, %+v", err)
	}
	jitoPool.AddRPC(jitoHttpRPCs...)
}
