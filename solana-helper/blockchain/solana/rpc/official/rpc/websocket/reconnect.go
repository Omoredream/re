package officialWebSocket

import (
	"context"

	"github.com/avast/retry-go/v4"
	rpc_ws "github.com/gagliardetto/solana-go/rpc/ws"

	"git.wkr.moe/web3/solana-helper/consts"
	"git.wkr.moe/web3/solana-helper/utils"
)

func (node *RPC) Reconnect() {
	ctx := context.Background()
	ctx = context.WithValue(ctx, consts.CtxRPC, node.Name())

	if !node.reconnecting {
		node.reconnecting = true
		node.client.Close()
		var client *rpc_ws.Client
		_ = Utils.Retry(ctx, func() (err error) {
			err = Utils.Try(func() (err error) {
				client, err = node.connect(ctx)
				return
			}, nil, node.Success, node.Fail)
			if err != nil {
				return
			}

			err = Utils.Try(func() (err error) {
				//err = node.test(ctx)
				return
			}, nil, node.Success, node.Fail)
			if err != nil {
				return
			}

			return
		}, retry.Attempts(0))
		node.client = client
		node.reconnecting = false
	}
}
