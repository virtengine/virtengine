package testutil

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	sdktest "github.com/cosmos/cosmos-sdk/testutil"

	"github.com/virtengine/virtengine/sdk/go/cli"
)

// ExecTxCreateProvider is used for testing create provider tx
func ExecTxCreateProvider(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxProviderCreateCmd(), args...)
}

// ExecTxUpdateProvider is used for testing update provider tx
func ExecTxUpdateProvider(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxProviderUpdateCmd(), args...)
}

// TxUpdateProviderExec is an alias for ExecTxUpdateProvider.
func TxUpdateProviderExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTxUpdateProvider(ctx, cctx, args...)
}

// ExecQueryProviders is used for testing providers query
func ExecQueryProviders(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryProvidersCmd(), args...)
}

// QueryProvidersExec is an alias for ExecQueryProviders.
func QueryProvidersExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecQueryProviders(ctx, cctx, args...)
}

// ExecQueryProvider is used for testing provider query
func ExecQueryProvider(ctx context.Context, cctx client.Context, extraArgs ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryProviderCmd(), extraArgs...)
}

// TxCreateProviderExec is an alias for ExecTxCreateProvider with config path
func TxCreateProviderExec(ctx context.Context, cctx client.Context, configPath string, flags ...string) (sdktest.BufferWriter, error) {
	allArgs := append([]string{configPath}, flags...)
	return ExecTxCreateProvider(ctx, cctx, allArgs...)
}

// TxUpdateProviderExec is an alias for ExecTxUpdateProvider with config path
func TxUpdateProviderExec(ctx context.Context, cctx client.Context, configPath string, flags ...string) (sdktest.BufferWriter, error) {
	allArgs := append([]string{configPath}, flags...)
	return ExecTxUpdateProvider(ctx, cctx, allArgs...)
}

// QueryProviderExec is an alias for ExecQueryProvider
func QueryProviderExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecQueryProvider(ctx, cctx, args...)
}
