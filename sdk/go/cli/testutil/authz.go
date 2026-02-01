package testutil

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

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

// TxGrantAuthorizationExec is used for testing grant authorization tx
func TxGrantAuthorizationExec(ctx context.Context, cctx client.Context, grantee sdk.AccAddress, args ...string) (testutil.BufferWriter, error) {
	allArgs := append([]string{grantee.String(), "deposit"}, args...)
	return ExecCreateGrant(ctx, cctx, allArgs...)
}

// TxRevokeAuthorizationExec is used for testing revoke authorization tx
func TxRevokeAuthorizationExec(ctx context.Context, cctx client.Context, grantee sdk.AccAddress, args ...string) (testutil.BufferWriter, error) {
	allArgs := append([]string{grantee.String(), "deposit"}, args...)
	return ExecRevokeAuthz(ctx, cctx, allArgs...)
}
