package tests

import (
	"bufio"
	"context"
	"os"
	"sync"

	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gmutex"
	"github.com/gogf/gf/v2/os/gtime"

	"git.wkr.moe/web3/solana-helper/consts"

	jupiterHTTP "git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jupiter/rpc"
)

type jupiterHttpRPCFeature = string

const (
	jupiterHTTPTooManyTokensFeature jupiterHttpRPCFeature = "币池过多"
	jupiterHTTPArbDisabledFeature   jupiterHttpRPCFeature = "禁止套利"
)

type jupiterHttpRPCConfig struct {
	Name     string      `json:"name"`
	URL      []string    `json:"url"`
	Location string      `json:"location"`
	Region   string      `json:"region"`
	Level    string      `json:"level"`
	Feature  gset.StrSet `json:"feature"`
	Offline  bool        `json:"offline"`
}

func writeLinesToFile(filename string, lines []string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}

func getJupiterHttpRPCs(ctx g.Ctx) (rpcs []*jupiterHTTP.RPC, err error) {
	rpcs = make([]*jupiterHTTP.RPC, 0)
	mu := &gmutex.Mutex{}
	wg := sync.WaitGroup{}
	semaphore := make(chan struct{}, 10)

	var (
		lv_     = g.Cfg().MustGet(nil, "rpcs.jupiter.filter.lv?", true).Bool()
		lv0     = g.Cfg().MustGet(nil, "rpcs.jupiter.filter.lv0", true).Bool() // 毫无机会, 0
		lv1     = g.Cfg().MustGet(nil, "rpcs.jupiter.filter.lv1", true).Bool() // 机会随缘, 0 - 90
		lv2     = g.Cfg().MustGet(nil, "rpcs.jupiter.filter.lv2", true).Bool() // 略有机会, 90 - 180
		lv3     = g.Cfg().MustGet(nil, "rpcs.jupiter.filter.lv3", true).Bool() // 机会寻常, 180 - 360
		lv4     = g.Cfg().MustGet(nil, "rpcs.jupiter.filter.lv4", true).Bool() // 机会特多, 360 - 720
		lv5     = g.Cfg().MustGet(nil, "rpcs.jupiter.filter.lv5", true).Bool() // 机会究极多, 720 - 1440
		lv6     = g.Cfg().MustGet(nil, "rpcs.jupiter.filter.lv6", true).Bool() // 印钞机, 1440+
		offline = g.Cfg().MustGet(nil, "rpcs.jupiter.filter.offline", false).Bool()

		tooManyTokens = g.Cfg().MustGet(nil, "rpcs.jupiter.filter.币池过多", false).Bool()
		arbDisabled   = g.Cfg().MustGet(nil, "rpcs.jupiter.filter.禁止套利", false).Bool()
	)

	valleyTime := g.Cfg().MustGet(nil, "rpcs.jupiter.filter.valleyHourStart", 6).Int() <=
		gtime.Now().Hour() &&
		gtime.Now().Hour() <=
			g.Cfg().MustGet(nil, "rpcs.jupiter.filter.valleyHourEnd", 14).Int()

	if tooManyTokens && arbDisabled && 自适应 || 自适应 {
		wg.Add(1)
		semaphore <- struct{}{}
		go func() {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			rpc, err := jupiterHTTP.New(ctx,
				"Jupiter 官方", "https://quote-api.jup.ag/v6",
				5_000, 1,
			)
			if err != nil {
				g.Log().Errorf(ctx, "创建 RPC 失败, %+v", err)
			} else {
				mu.LockFunc(func() {
					rpcs = append(rpcs, rpc)
				})
			}
		}()
	}

	var rpcConfigs []jupiterHttpRPCConfig
	err = g.Cfg(g.Cfg().MustGet(nil, "rpcs.jupiter.httpRPCConfigFile").String()).MustGet(nil, "rpc").Structs(&rpcConfigs)
	if err != nil {
		err = gerror.Wrapf(err, "解析配置文件中的 RPC 失败")
		return
	}

	for _, rpcConfig := range rpcConfigs {
		if Region != rpcConfig.Region && rpcConfig.Region != "global" {
			continue
		}
		if !offline && rpcConfig.Offline {
			continue
		}
		if !tooManyTokens && rpcConfig.Feature.Contains(jupiterHTTPTooManyTokensFeature) {
			continue
		}
		if !arbDisabled && rpcConfig.Feature.Contains(jupiterHTTPArbDisabledFeature) {
			continue
		}
		if !valleyTime {
			if !lv_ && rpcConfig.Level == "lv?" {
				continue
			}
			if !lv0 && rpcConfig.Level == "lv0" {
				continue
			}
			if !lv1 && rpcConfig.Level == "lv1" {
				continue
			}
			if !lv2 && rpcConfig.Level == "lv2" {
				continue
			}
			if !lv3 && rpcConfig.Level == "lv3" {
				continue
			}
			if !lv4 && rpcConfig.Level == "lv4" {
				continue
			}
			if !lv5 && rpcConfig.Level == "lv5" {
				continue
			}
			if !lv6 && rpcConfig.Level == "lv6" {
				continue
			}
		}

		if gset.NewStrSetFrom(rpcConfig.URL).Size() != len(rpcConfig.URL) {
			err = gerror.Newf("RPC %s 存在重复 URL", rpcConfig.Name)
			return
		}

		wg.Add(1)
		semaphore <- struct{}{}
		go func() {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			ctx := context.WithValue(ctx, consts.CtxRPC, rpcConfig.Name)

			rpc, err := jupiterHTTP.New(ctx,
				rpcConfig.Name, rpcConfig.URL[0],
				10_000, 4,
			)
			if err != nil {
				if !rpcConfig.Offline {
					g.Log().Errorf(ctx, "创建 RPC 失败, %+v", err)
				}
				return
			} else if rpcConfig.Offline {
				g.Log().Warningf(ctx, "RPC 恢复在线")
			}

			if rpc.IsTooManyToken() && !rpcConfig.Feature.Contains(jupiterHTTPTooManyTokensFeature) {
				g.Log().Warningf(ctx, "RPC 支持代币数量 %d 过多", rpc.TokensCount())
			} else if !rpc.IsTooManyToken() && rpcConfig.Feature.Contains(jupiterHTTPTooManyTokensFeature) {
				g.Log().Warningf(ctx, "RPC 无 %s 特性", jupiterHTTPTooManyTokensFeature)
			}

			if rpc.IsArbDisabled() && !rpcConfig.Feature.Contains(jupiterHTTPArbDisabledFeature) {
				g.Log().Warningf(ctx, "RPC 禁止套利")
			} else if !rpc.IsArbDisabled() && rpcConfig.Feature.Contains(jupiterHTTPArbDisabledFeature) {
				g.Log().Warningf(ctx, "RPC 无 %s 特性", jupiterHTTPArbDisabledFeature)
			}

			mu.LockFunc(func() {
				rpcs = append(rpcs, rpc)
			})
		}()
	}

	wg.Wait()

	if len(rpcs) == 0 {
		err = gerror.Newf("未创建任何可用的 Jupiter Swap API")
		return
	} else {
		// 将 rpcs 中的 Name 写入 jupiter_http_rpcs.txt
		names := make([]string, 0)
		for _, rpc := range rpcs {
			names = append(names, rpc.Name())
		}
		err = writeLinesToFile("valid_jupiter_http_rpcs.txt", names)
		if err != nil {
			err = gerror.Wrapf(err, "写入valid_jupiter_http_rpcs.txt失败")
		}
	}
	g.Log().Infof(ctx, "导入了 %d 个 Jupiter Swap API", len(rpcs))

	return
}
