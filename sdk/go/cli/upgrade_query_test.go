package cli_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"

	"cosmossdk.io/x/upgrade"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/virtengine/virtengine/sdk/go/cli"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

func TestGetCurrentPlanCmd(t *testing.T) {
	encCfg := sdkutil.MakeEncodingConfig(upgrade.AppModuleBasic{})
	kr := keyring.NewInMemory(encCfg.Codec)
	baseCtx := client.Context{}.
		WithKeyring(kr).
		WithTxConfig(encCfg.TxConfig).
		WithCodec(encCfg.Codec).
		WithClient(testutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			name:         "json output",
			args:         cli.TestFlags().WithOutputJSON(),
			expCmdOutput: `[--output=json]`,
		},
		{
			name:         "text output",
			args:         cli.TestFlags().WithOutputText(),
			expCmdOutput: `[--output=text]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := cli.GetQueryUpgradeCurrentPlanCmd()
			cmd.SetOut(io.Discard)
			require.NotNil(t, cmd)

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			require.NoError(t, client.SetCmdClientContextHandler(baseCtx, cmd))

			require.Contains(t, fmt.Sprint(cmd), "plan [] [] get upgrade plan (if one exists)")
			require.Contains(t, fmt.Sprint(cmd), tc.expCmdOutput)
		})
	}
}

func TestGetAppliedPlanCmd(t *testing.T) {
	encCfg := sdkutil.MakeEncodingConfig(upgrade.AppModuleBasic{})
	kr := keyring.NewInMemory(encCfg.Codec)
	baseCtx := client.Context{}.
		WithKeyring(kr).
		WithTxConfig(encCfg.TxConfig).
		WithCodec(encCfg.Codec).
		WithClient(testutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			name:         "json output",
			args:         cli.TestFlags().With("test-upgrade").WithOutputJSON(),
			expCmdOutput: `[test-upgrade --output=json]`,
		},
		{
			name:         "text output",
			args:         cli.TestFlags().With("test-upgrade").WithOutputText(),
			expCmdOutput: `[test-upgrade --output=text]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := cli.GetQueryUpgradeAppliedPlanCmd()
			cmd.SetOut(io.Discard)
			require.NotNil(t, cmd)

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			require.NoError(t, client.SetCmdClientContextHandler(baseCtx, cmd))

			require.Contains(t, fmt.Sprint(cmd), "applied [upgrade-name] [] [] block header for height at which a completed upgrade was applied")
			require.Contains(t, fmt.Sprint(cmd), tc.expCmdOutput)
		})
	}
}

func TestGetModuleVersionsCmd(t *testing.T) {
	encCfg := sdkutil.MakeEncodingConfig(upgrade.AppModuleBasic{})
	kr := keyring.NewInMemory(encCfg.Codec)
	baseCtx := client.Context{}.
		WithKeyring(kr).
		WithTxConfig(encCfg.TxConfig).
		WithCodec(encCfg.Codec).
		WithClient(testutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	testCases := []struct {
		msg          string
		args         []string
		expCmdOutput string
	}{
		{
			msg:          "test full query with json output",
			args:         cli.TestFlags().WithHeight(1).WithOutputJSON(),
			expCmdOutput: `--height=1 --output=json`,
		},
		{
			msg:          "test full query with text output",
			args:         cli.TestFlags().WithHeight(1).WithOutputText(),
			expCmdOutput: `--height=1 --output=text`,
		},
		{
			msg:          "test single module",
			args:         cli.TestFlags().With("bank").WithHeight(1),
			expCmdOutput: `bank --height=1`,
		},
		{
			msg:          "test non-existent module",
			args:         cli.TestFlags().With("abcdefg").WithHeight(1),
			expCmdOutput: `abcdefg --height=1`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.msg, func(t *testing.T) {
			ctx := svrcmd.CreateExecuteContext(context.Background())

			cmd := cli.GetQueryUpgradeModuleVersionsCmd()
			cmd.SetOut(io.Discard)
			require.NotNil(t, cmd)

			cmd.SetContext(ctx)
			cmd.SetArgs(tc.args)

			require.NoError(t, client.SetCmdClientContextHandler(baseCtx, cmd))

			require.Contains(t, fmt.Sprint(cmd), "module_versions [optional module_name] [] [] get the list of module versions")
			require.Contains(t, fmt.Sprint(cmd), tc.expCmdOutput)
		})
	}
}

