package client

import (
	"context"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	jsonrpcclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	"golang.org/x/sync/errgroup"

	"github.com/cosmos/cosmos-sdk/client"
)

type RPCClient interface {
	client.CometRPC
	Akash(ctx context.Context) (*Akash, error)
}

type rpcClient struct {
	*rpchttp.HTTP
	rpc *jsonrpcclient.Client

	group *errgroup.Group
	ctx   context.Context
}

var _ client.CometRPC = (*rpcClient)(nil)

// NewClient allows for setting a custom http client (See New).
// An error is returned on invalid remote. The function panics when remote is nil.
func NewClient(ctx context.Context, remote string) (RPCClient, error) {
	httpClient, err := NewHTTPClient(ctx, remote)
	if err != nil {
		return nil, err
	}

	cl, err := rpchttp.NewWithClient(remote, "/websocket", httpClient)
	if err != nil {
		return nil, err
	}

	rc, err := jsonrpcclient.NewWithHTTPClient(remote, httpClient)
	if err != nil {
		return nil, err
	}

	group, ctx := errgroup.WithContext(ctx)

	rpc := &rpcClient{
		HTTP:  cl,
		rpc:   rc,
		group: group,
		ctx:   ctx,
	}

	group.Go(func() error {
		err := cl.Start()
		if err != nil {
			return err
		}

		<-ctx.Done()

		return cl.Stop()
	})

	return rpc, nil
}

func (cl *rpcClient) Akash(ctx context.Context) (*Akash, error) {
	result := &Akash{}
	_, err := cl.rpc.Call(ctx, "akash", map[string]interface{}{}, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
