package tests

import (
	"context"
	"net"
	"sync"

	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gmutex"

	"git.wkr.moe/web3/solana-helper/consts"
	Utils "git.wkr.moe/web3/solana-helper/utils"

	jitoHTTP "git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/jito/rpc"
)

type jitoHttpRPCFeature = string

const (
	jitoHttpMultiNetFeature jitoHttpRPCFeature = "多线"
)

type jitoHttpRPCConfig struct {
	Name   string   `json:"name"`
	URL    []string `json:"url"`
	Config struct {
		CooldownIntervalMill int64 `json:"cooldownIntervalMill"`
		MaxRunningThreads    uint  `json:"maxRunningThreads"`
	} `json:"config"`
	Location string      `json:"location"`
	Region   string      `json:"region"`
	Feature  gset.StrSet `json:"feature"`
	Offline  bool        `json:"offline"`
}

func getJitoHttpRPCs(ctx g.Ctx) (rpcs []*jitoHTTP.RPC, err error) {
	rpcs = make([]*jitoHTTP.RPC, 0)
	mu := &gmutex.Mutex{}
	wg := sync.WaitGroup{}
	semaphore := make(chan struct{}, 10)

	IPs := Utils.Map(g.Cfg().MustGet(nil, "rpcs.jito.config.ip").Strings(), func(s string) (ip net.IP) {
		return net.ParseIP(s)
	})
	UUID := g.Cfg().MustGet(nil, "rpcs.jito.config.uuid").String()

	var (
		offline = g.Cfg().MustGet(nil, "rpcs.jito.filter.offline", false).Bool()

		multiNet = g.Cfg().MustGet(nil, "rpcs.jito.filter.多线", true).Bool()
	)

	var rpcConfigs []jitoHttpRPCConfig
	err = g.Cfg(g.Cfg().MustGet(nil, "rpcs.jito.httpRPCConfigFile").String()).MustGet(nil, "rpc").Structs(&rpcConfigs)
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
		if !multiNet && rpcConfig.Feature.Contains(jitoHttpMultiNetFeature) {
			continue
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

			if len(IPs) == 0 {
				IPs = append(IPs, nil)
			}

			for _, ip := range IPs {
				if UUID != "" {
					rpc, err := jitoHTTP.New(ctx,
						rpcConfig.Name+"-"+ip.String()+"-"+UUID, rpcConfig.URL[0],
						rpcConfig.Config.CooldownIntervalMill, rpcConfig.Config.MaxRunningThreads,
						UUID, &ip,
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
				}

				rpc, err := jitoHTTP.New(ctx,
					rpcConfig.Name+"-"+ip.String(), rpcConfig.URL[0],
					rpcConfig.Config.CooldownIntervalMill, rpcConfig.Config.MaxRunningThreads,
					"", &ip,
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
			}
		}()
	}

	wg.Wait()

	if len(rpcs) == 0 {
		err = gerror.Newf("未创建任何可用的 Jito JSON-RPC")
		return
	}

	g.Log().Infof(ctx, "导入了 %d 个 Jito JSON-RPC", len(rpcs))

	return
}
