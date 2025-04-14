package officialWebSocket

import (
	"context"
	"net/http"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	rpc_ws "github.com/gagliardetto/solana-go/rpc/ws"

	"git.wkr.moe/web3/solana-helper/consts"
)

func (node *RPC) connect(ctx g.Ctx) (client *rpc_ws.Client, err error) {
	ctx = context.WithValue(ctx, consts.CtxRPC, node.Name())

	g.Log().Debugf(ctx, "正在尝试连接到 RPC")

	client, err = rpc_ws.ConnectWithOptions(ctx, node.url, &rpc_ws.Options{
		HttpHeader: func() (httpHeaders http.Header) {
			httpHeaders = make(http.Header, len(node.headers))
			for i := range node.headers {
				httpHeaders[i] = []string{node.headers[i]}
			}
			return
		}(),
		ShortID: node.shortID,
	})
	if err != nil {
		err = gerror.Wrapf(err, "连接到 RPC 失败")
		return
	}

	g.Log().Debugf(ctx, "连接到 RPC 成功")

	return
}
