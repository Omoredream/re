package officialHTTP

import (
	"context"
	"net"
	"slices"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/imroc/req/v3"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/jsonrpc"

	"git.wkr.moe/web3/solana-helper/rpcinterface"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"

	"git.wkr.moe/web3/solana-helper/blockchain/solana/rpc/official/rpc/types"
)

type RPC struct {
	*rpcinterface.RPC
	client *rpc.Client

	fingerprint string
	version     []uint
}

func New(ctx g.Ctx, semaphore chan struct{}, name string, url string, cooldownIntervalMill int64, maxRunningThreads uint, ip *net.IP, options ...Option) (node *RPC, err error) {
	client := req.C().
		//SetProxyURL("socks5://127.0.0.1:10801").
		DisableAutoReadResponse().
		DisableAutoDecode()
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
		SetResponseHeaderTimeout(1 * time.Second)

	for _, option := range options {
		option(client)
	}

	node = &RPC{
		RPC: rpcinterface.New(string(officialRPCTypes.HTTP)+"/"+name, cooldownIntervalMill, maxRunningThreads),
		client: rpc.NewWithCustomRPCClient(jsonrpc.NewClientWithOpts(url, &jsonrpc.RPCClientOpts{
			HTTPClient: &clientWrapper{client},
		})),
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

	return
}

//func (node *RPC) Fingerprint() string {
//	return node.fingerprint
//}

func (node *RPC) Client() *rpc.Client {
	return node.client
}

func (node *RPC) Version() []uint {
	return slices.Clone(node.version)
}
