package testutil

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	sdktest "github.com/cosmos/cosmos-sdk/testutil"

	"github.com/virtengine/virtengine/sdk/go/cli"
)

// ExecCreateBid is used for testing create bid tx
func ExecCreateBid(ctx context.Context, cctx client.Context, extraArgs ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxMarketBidCreateCmd(), extraArgs...)
}

// ExecCloseBid is used for testing close bid tx
func ExecCloseBid(ctx context.Context, cctx client.Context, extraArgs ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxMarketBidCloseCmd(), extraArgs...)
}

// ExecCreateLease is used for creating a lease
func ExecCreateLease(ctx context.Context, cctx client.Context, extraArgs ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxMarketLeaseCreateCmd(), extraArgs...)
}

// ExecCloseLease is used for testing close order tx
func ExecCloseLease(ctx context.Context, cctx client.Context, extraArgs ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxMarketLeaseCloseCmd(), extraArgs...)
}

// ExecQueryOrders is used for testing orders query
func ExecQueryOrders(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryMarketOrdersCmd(), args...)
}

// ExecQueryOrder is used for testing order query
func ExecQueryOrder(ctx context.Context, cctx client.Context, extraArgs ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryMarketOrderCmd(), extraArgs...)
}

// QueryOrdersExec is an alias for ExecQueryOrders
func QueryOrdersExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecQueryOrders(ctx, cctx, args...)
}

// QueryOrderExec is an alias for ExecQueryOrder
func QueryOrderExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecQueryOrder(ctx, cctx, args...)
}

// ExecQueryBids is used for testing bids query
func ExecQueryBids(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryMarketBidsCmd(), args...)
}

// ExecQueryBid is used for testing bid query
func ExecQueryBid(ctx context.Context, cctx client.Context, extraArgs ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryMarketBidCmd(), extraArgs...)
}

// ExecQueryLeases is used for testing leases query
func ExecQueryLeases(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryMarketLeasesCmd(), args...)
}

// ExecQueryLease is used for testing lease query
func ExecQueryLease(ctx context.Context, cctx client.Context, extraArgs ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryMarketLeaseCmd(), extraArgs...)
}
