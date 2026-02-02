package testutil

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	sdktest "github.com/cosmos/cosmos-sdk/testutil"

	"github.com/virtengine/virtengine/sdk/go/cli"
)

// ExecBMEBurnMint is used for testing BME burn-mint tx
func ExecBMEBurnMint(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxBMEBurnMintCmd(), args...)
}

// ExecQueryBMEParams is used for testing BME params query
func ExecQueryBMEParams(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetBMEParamsCmd(), args...)
}

// ExecQueryBMEVaultState is used for testing BME vault state query
func ExecQueryBMEVaultState(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetBMEVaultStateCmd(), args...)
}

// ExecQueryBMEStatus is used for testing BME status query
func ExecQueryBMEStatus(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetBMEStatusCmd(), args...)
}
