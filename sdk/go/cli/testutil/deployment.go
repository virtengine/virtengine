package testutil

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	sdktest "github.com/cosmos/cosmos-sdk/testutil"

	"github.com/virtengine/virtengine/sdk/go/cli"
)

// ExecDeploymentCreate is used for testing create deployment tx
func ExecDeploymentCreate(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxDeploymentCreateCmd(), args...)
}

// ExecDeploymentUpdate is used for testing update deployment tx
func ExecDeploymentUpdate(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxDeploymentUpdateCmd(), args...)
}

// ExecDeploymentClose is used for testing close deployment tx
// requires --dseq, --fees
func ExecDeploymentClose(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxDeploymentCloseCmd(), args...)
}

// ExecDeploymentGroupClose is used for testing close deployment group tx
// requires --dseq, --fees
func ExecDeploymentGroupClose(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetTxDeploymentGroupCloseCmd(), args...)
}

// ExecQueryDeployments is used for testing deployments query
func ExecQueryDeployments(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryDeploymentsCmd(), args...)
}

// ExecQueryDeployment is used for testing deployment query
func ExecQueryDeployment(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryDeploymentCmd(), args...)
}

// ExecQueryGroup is used for testing group query
func ExecQueryGroup(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryDeploymentGroupCmd(), args...)
}
