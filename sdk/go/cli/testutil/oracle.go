package testutil

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	sdktest "github.com/cosmos/cosmos-sdk/testutil"

	"github.com/virtengine/virtengine/sdk/go/cli"
)

// ExecOracleFeedPrice is used for testing oracle feed price tx
func ExecOracleFeedPrice(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxOracleFeedPriceCmd(), args...)
}

// ExecQueryOraclePrices is used for testing oracle prices query
func ExecQueryOraclePrices(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetOraclePricesCmd(), args...)
}

// ExecQueryOracleParams is used for testing oracle params query
func ExecQueryOracleParams(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetOracleParamsCmd(), args...)
}

// ExecQueryOraclePriceFeedConfig is used for testing oracle price feed config query
func ExecQueryOraclePriceFeedConfig(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetOraclePriceFeedConfigCmd(), args...)
}

