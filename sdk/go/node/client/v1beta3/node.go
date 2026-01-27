package v1beta3

import (
	"context"

	tmrpc "github.com/cometbft/cometbft/rpc/core/types"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
)

var _ NodeClient = (*node)(nil)

type node struct {
	rpc sdkclient.CometRPC
}

func newNode(cctx sdkclient.Context) *node {
	nd := &node{
		rpc: cctx.Client,
	}

	return nd
}

func (nd *node) SyncInfo(ctx context.Context) (*tmrpc.SyncInfo, error) {
	status, err := nd.rpc.Status(ctx)
	if err != nil {
		return nil, err
	}

	info := status.SyncInfo

	return &info, nil
}

func (nd *node) CurrentBlockHeight(ctx context.Context) (int64, error) {
	info, err := nd.SyncInfo(ctx)
	if err != nil {
		return 0, err
	}

	return info.LatestBlockHeight, nil
}
