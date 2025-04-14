package jupiterHTTP

import (
	"context"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/gogf/gf/v2/container/gset"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtimer"
	"github.com/imroc/req/v3"

	"git.wkr.moe/web3/solana-helper/rpcinterface"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"
)

type RPC struct {
	*rpcinterface.RPC
	client        *req.Client
	scheduler     *gtimer.Timer
	supportTokens *gset.StrSet

	fingerprint  string
	tooManyToken bool
	arbDisabled  bool
}

func New(ctx g.Ctx, semaphore chan struct{}, name string, url string, cooldownIntervalMill int64, maxRunningThreads uint) (node *RPC, err error) {
	client := req.C().
		//SetProxyURL("socks5://127.0.0.1:10801").
		SetTimeout(10 * time.Second).
		EnableInsecureSkipVerify().
		SetBaseURL(url)

	node = &RPC{
		RPC:           rpcinterface.New(name, cooldownIntervalMill, maxRunningThreads),
		client:        client,
		scheduler:     gtimer.New(),
		supportTokens: gset.NewStrSet(true),
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
	}, retry.Attempts(5), retry.Delay(5*time.Second))
	if err != nil {
		err = gerror.Wrapf(err, "初始化 RPC 失败")
		return
	}

	node.scheduler.AddSingleton(ctx, 1*time.Minute, func(ctx context.Context) {
		err := node.updateSupportTokens(ctx)
		if err != nil {
			g.Log("scheduler").Errorf(ctx, "%+v", err)
		}
	})

	return
}

func (node *RPC) Fingerprint() string {
	return node.fingerprint
}

func (node *RPC) Weight() float64 {
	return 0
}

func (node *RPC) TokensCount() int {
	return node.supportTokens.Size()
}

func (node *RPC) IsTooManyToken() bool {
	return node.tooManyToken
}

func (node *RPC) IsArbDisabled() bool {
	return node.arbDisabled
}
