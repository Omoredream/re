package jitoHTTP

import (
	"context"
	"net"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gmutex"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/imroc/req/v3"

	"git.wkr.moe/web3/solana-helper/rpcinterface"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/address"
)

type RPC struct {
	*rpcinterface.RPC
	c                *req.Client
	scheduler        *gtimer.Timer
	tipAccounts      []Address.AccountAddress
	tipAccountsMutex *gmutex.RWMutex
}

func New(ctx g.Ctx, semaphore chan struct{}, name string, url string, cooldownIntervalMill int64, maxRunningThreads uint, ip *net.IP, uuid string) (node *RPC, err error) {
	client := req.C()
	if ip != nil {
		client.
			SetDial(func(ctx context.Context, network, addr string) (net.Conn, error) {
				return (&net.Dialer{
					LocalAddr: &net.TCPAddr{
						IP: *ip,
					},
				}).Dial(network, addr)
			})
	}

	client.
		//SetProxyURL("socks5://127.0.0.1:10801").
		SetBaseURL(url)
	if uuid != "" {
		client.
			SetCommonHeader("x-jito-auth", uuid)
	}

	node = &RPC{
		RPC:              rpcinterface.New(name, cooldownIntervalMill, maxRunningThreads),
		c:                client,
		scheduler:        gtimer.New(),
		tipAccounts:      make([]Address.AccountAddress, 0),
		tipAccountsMutex: &gmutex.RWMutex{},
	}

	ctx = context.WithValue(ctx, consts.CtxRPC, node.Name())

	err = Utils.Retry(ctx, func() (err error) {
		if semaphore != nil {
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()
		}
		err = Utils.Try(func() (err error) {
			err = node.test(ctx)
			return
		}, nil, node.Success, node.Fail)
		if err != nil {
			return
		}

		return
	})
	if err != nil {
		err = gerror.Wrapf(err, "初始化 RPC 失败")
		return
	}

	node.scheduler.AddSingleton(ctx, 1*gtime.D, func(ctx context.Context) {
		err := node.updateTipAccounts(ctx)
		if err != nil {
			g.Log("scheduler").Errorf(ctx, "%+v", err)
		}
	})

	return
}

func (node *RPC) client() *req.Client {
	return node.c
}
