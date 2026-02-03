package testutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/testutil"

	"github.com/virtengine/virtengine/sdk/go/cli"
)

func execSetContext(ctx context.Context, cctx client.Context, cmd *cobra.Command, args ...string) (testutil.BufferWriter, error) {
	cmd.SetArgs(args)

	_, out := testutil.ApplyMockIO(cmd)
	cctx = cctx.WithOutput(out)

	ctx = context.WithValue(ctx, cli.ClientContextKey, &client.Context{})
	ctx = context.WithValue(ctx, server.ServerContextKey, server.NewDefaultContext())

	cmd.SetContext(ctx)
	cctx.CmdContext = ctx

	if err := cli.SetCmdClientContextHandler(cctx, cmd); err != nil {
		return nil, err
	}

	return out, nil
}

// ExecTestCLICmd builds the client context, mocks the output and executes the command.
func ExecTestCLICmd(ctx context.Context, cctx client.Context, cmd *cobra.Command, args ...string) (testutil.BufferWriter, error) {
	{
		dupFlags := make(map[string]bool)
		for _, arg := range args {
			if !strings.HasPrefix(arg, "--") {
				continue
			}

			arg = strings.TrimPrefix(arg, "--")
			tokens := strings.Split(arg, "=")

			if _, exists := dupFlags[tokens[0]]; exists {
				return nil, fmt.Errorf("test: duplicated flag \"%s\"", tokens[0])
			}

			dupFlags[tokens[0]] = true
		}
	}

	out, err := execSetContext(ctx, cctx, cmd, args...)
	if err != nil {
		return nil, err
	}

	if err := cmd.Execute(); err != nil {
		return out, err
	}

	return out, nil
}

func QueryBalancesExec(ctx context.Context, cctx client.Context, args ...string) (testutil.BufferWriter, error) {
	return ExecTestCLICmd(ctx, cctx, cli.GetQueryBankBalancesCmd(), args...)
}
