package testutil

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"

	"github.com/virtengine/virtengine/sdk/go/cli"
)

func ExecCreateGrant(ctx context.Context, cctx client.Context, args ...string) (testutil.BufferWriter, error) {
	cmd := cli.GetTxAuthzGrantAuthorizationCmd()
	return ExecTestCLICmd(ctx, cctx, cmd, args...)
}

func ExecRevokeAuthz(ctx context.Context, cctx client.Context, args ...string) (testutil.BufferWriter, error) {
	cmd := cli.GetTxAuthzRevokeAuthorizationCmd()
	return ExecTestCLICmd(ctx, cctx, cmd, args...)
}

func ExecAuthorization(ctx context.Context, cctx client.Context, args ...string) (testutil.BufferWriter, error) {
	cmd := cli.GetTxAuthzExecAuthorizationCmd()
	return ExecTestCLICmd(ctx, cctx, cmd, args...)
}
