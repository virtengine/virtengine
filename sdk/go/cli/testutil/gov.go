package testutil

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
)

// ExecGovSubmitLegacyProposal creates a tx for submit legacy proposal
//
//nolint:staticcheck // we are intentionally using a deprecated flag here.
func ExecGovSubmitLegacyProposal(ctx context.Context, cctx client.Context, args ...string) (testutil.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, govcli.NewCmdSubmitLegacyProposal(), args...)
}

// MsgVote votes for a proposal
func MsgVote(clientCtx client.Context, args ...string) (testutil.BufferWriter, error) {
	return clitestutil.ExecTestCLICmd(clientCtx, govcli.NewCmdWeightedVote(), args)
}

// MsgDeposit deposits on a proposal
func MsgDeposit(clientCtx client.Context, args ...string) (testutil.BufferWriter, error) {
	return clitestutil.ExecTestCLICmd(clientCtx, govcli.NewCmdDeposit(), args)
}
