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

// TxCreateDeploymentExec is an alias for ExecDeploymentCreate
func TxCreateDeploymentExec(ctx context.Context, cctx client.Context, sdlPath string, args ...string) (sdktest.BufferWriter, error) {
	allArgs := append([]string{sdlPath}, args...)
	return ExecDeploymentCreate(ctx, cctx, allArgs...)
}

// TxUpdateDeploymentExec is an alias for ExecDeploymentUpdate
func TxUpdateDeploymentExec(ctx context.Context, cctx client.Context, sdlPath string, args ...string) (sdktest.BufferWriter, error) {
	allArgs := append([]string{sdlPath}, args...)
	return ExecDeploymentUpdate(ctx, cctx, allArgs...)
}

// TxCloseDeploymentExec is an alias for ExecDeploymentClose
func TxCloseDeploymentExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecDeploymentClose(ctx, cctx, args...)
}

// TxCloseGroupExec is an alias for ExecDeploymentGroupClose
func TxCloseGroupExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecDeploymentGroupClose(ctx, cctx, args...)
}

// QueryDeploymentsExec is an alias for ExecQueryDeployments
func QueryDeploymentsExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecQueryDeployments(ctx, cctx, args...)
}

// QueryDeploymentExec is an alias for ExecQueryDeployment
func QueryDeploymentExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecQueryDeployment(ctx, cctx, args...)
}

// QueryGroupExec is an alias for ExecQueryGroup
func QueryGroupExec(ctx context.Context, cctx client.Context, args ...string) (sdktest.BufferWriter, error) {
	return ExecQueryGroup(ctx, cctx, args...)
}

// TxDepositDeploymentExec is used for testing deployment deposit tx
func TxDepositDeploymentExec(ctx context.Context, cctx client.Context, amount interface{}, args ...string) (sdktest.BufferWriter, error) {
	allArgs := append([]string{"deployment", amount.(interface{ String() string }).String()}, args...)
	return ExecTestCLICmd(ctx, cctx, cli.GetTxEscrowDeposit(), allArgs...)
}
