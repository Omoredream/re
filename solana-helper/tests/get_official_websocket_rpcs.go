package tests

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gmutex"

	"git.wkr.moe/web3/solana-helper/consts"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/official/rpc/websocket"
)

type officialWebsocketRPCFeature = string

const (
	officialWebsocketTritonFeature   officialWebsocketRPCFeature = "Triton"
	officialWebsocketHeliusFeature   officialWebsocketRPCFeature = "Helius"
	officialWebsocketCDNFeature      officialWebsocketRPCFeature = "CDN"
	officialWebsocketMultiNetFeature officialWebsocketRPCFeature = "多线"
)

type officialWebsocketRPCConfig struct {
	Name   string   `json:"name"`
	URL    []string `json:"url"`
	Config struct {
		CooldownIntervalMill int64             `json:"cooldownIntervalMill"`
		MaxRunningThreads    uint              `json:"maxRunningThreads"`
		ShortID              bool              `json:"shortID"`
		Headers              map[string]string `json:"headers"`
	} `json:"config"`
	Location string      `json:"location"`
	Region   string      `json:"region"`
	Feature  gset.StrSet `json:"feature"`
	Offline  bool        `json:"offline"`
}

func getOfficialWebsocketRPCs(ctx g.Ctx) (rpcs []*officialWebSocket.RPC, err error) {
	if g.Cfg().MustGet(nil, "rpcs.official", nil).IsEmpty() {
		return
	}

	rpcs = make([]*officialWebSocket.RPC, 0)
	mu := &gmutex.Mutex{}
	wg := sync.WaitGroup{}
	semaphore := make(chan struct{}, g.Cfg().MustGet(nil, "rpcs.official.importParallel", 10).Int())

	var (
		offline = g.Cfg().MustGet(nil, "rpcs.official.filter.offline", false).Bool()

		cdn      = g.Cfg().MustGet(nil, "rpcs.official.filter.CDN", true).Bool()
		multiNet = g.Cfg().MustGet(nil, "rpcs.official.filter.多线", true).Bool()
	)

	var rpcConfigs []officialWebsocketRPCConfig
	err = g.Cfg(g.Cfg().MustGet(nil, "rpcs.official.websocketRPCConfigFile").String()).MustGet(nil, "rpc").Structs(&rpcConfigs)
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
		if !cdn && rpcConfig.Feature.Contains(officialWebsocketCDNFeature) {
			continue
		}
		if !multiNet && rpcConfig.Feature.Contains(officialWebsocketMultiNetFeature) {
			continue
		}

		if gset.NewStrSetFrom(rpcConfig.URL).Size() != len(rpcConfig.URL) {
			err = gerror.Newf("RPC %s 存在重复 URL", rpcConfig.Name)
			return
		}

		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
			}()
			ctx := context.WithValue(ctx, consts.CtxRPC, rpcConfig.Name)

			rpc, err := officialWebSocket.New(ctx, semaphore,
				rpcConfig.Name, rpcConfig.URL[0],
				rpcConfig.Config.CooldownIntervalMill, rpcConfig.Config.MaxRunningThreads,
				rpcConfig.Config.Headers, rpcConfig.Config.ShortID,
			)
			if err != nil {
				if !rpcConfig.Offline {
					g.Log().Errorf(ctx, "创建 RPC 失败, %+v", err)
				}
				return
			} else if rpcConfig.Offline {
				g.Log().Warningf(ctx, "RPC 恢复在线")
			}

			mu.LockFunc(func() {
				rpcs = append(rpcs, rpc)
			})
		}()
	}

	wg.Wait()

	if len(rpcs) == 0 {
		err = gerror.Newf("未创建任何可用的 WebSocket RPC")
		return
	}

	g.Log().Infof(ctx, "导入了 %d 个 WebSocket RPC", len(rpcs))

	return
}
