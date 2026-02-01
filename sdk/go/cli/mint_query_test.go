package cli_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/x/mint"

	"github.com/virtengine/virtengine/sdk/go/cli"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

func TestGetQueryMintParamsCmd(t *testing.T) {
	encCfg := sdkutil.MakeEncodingConfig(mint.AppModuleBasic{})
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
		name           string
		flagArgs       []string
		expCmdOutput   string
		expectedOutput string
	}{
		{
			"json output",
			cli.TestFlags().
				WithHeight(1).
				WithOutputJSON(),
			`[--height=1 --output=json]`,
			`{"mint_denom":"","inflation_rate_change":"0","inflation_max":"0","inflation_min":"0","goal_bonded":"0","blocks_per_year":"0"}`,
		},
		{
			"text output",
			cli.TestFlags().
				WithHeight(1).
				WithOutputText(),
			`[--height=1 --output=text]`,
			`blocks_per_year: "0"
goal_bonded: "0"
inflation_max: "0"
inflation_min: "0"
inflation_rate_change: "0"
mint_denom: ""`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.GetQueryMintParamsCmd()

			out, err := clitestutil.ExecTestCLICmd(context.Background(), baseCtx, cmd, tc.flagArgs...)
			require.NoError(t, err)
			require.Equal(t, tc.expectedOutput, strings.TrimSpace(out.String()))

			if len(tc.flagArgs) != 0 {
				require.Contains(t, fmt.Sprint(cmd), "params [] [] Query the current minting parameters")
				require.Contains(t, fmt.Sprint(cmd), tc.expCmdOutput)
			}
		})
	}
}

func TestGetQueryMintInflationCmd(t *testing.T) {
	encCfg := sdkutil.MakeEncodingConfig(mint.AppModuleBasic{})
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
		name           string
		flagArgs       []string
		expCmdOutput   string
		expectedOutput string
	}{
		{
			"json output",
			cli.TestFlags().
				WithHeight(1).
				WithOutputJSON(),
			`[--height=1 --output=json]`,
			`<nil>`,
		},
		{
			"text output",
			cli.TestFlags().
				WithHeight(1).
				WithOutputText(),
			`[--height=1 --output=text]`,
			`<nil>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.GetQueryMintInflationCmd()

			out, err := clitestutil.ExecTestCLICmd(context.Background(), baseCtx, cmd, tc.flagArgs...)
			require.NoError(t, err)
			require.Equal(t, tc.expectedOutput, strings.TrimSpace(out.String()))

			if len(tc.flagArgs) != 0 {
				require.Contains(t, fmt.Sprint(cmd), "inflation [] [] Query the current minting inflation value")
				require.Contains(t, fmt.Sprint(cmd), tc.expCmdOutput)
			}
		})
	}
}

func TestGetCmdQueryAnnualProvisions(t *testing.T) {
	encCfg := sdkutil.MakeEncodingConfig(mint.AppModuleBasic{})
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
		name           string
		flagArgs       []string
		expCmdOutput   string
		expectedOutput string
	}{
		{
			"json output",
			cli.TestFlags().
				WithHeight(1).
				WithOutputJSON(),
			`[--height=1 --output=json]`,
			`<nil>`,
		},
		{
			"text output",
			cli.TestFlags().
				WithHeight(1).
				WithOutputText(),
			`[--height=1 --output=text]`,
			`<nil>`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.GetQueryMintAnnualProvisionsCmd()

			out, err := clitestutil.ExecTestCLICmd(context.Background(), baseCtx, cmd, tc.flagArgs...)
			require.NoError(t, err)
			require.Equal(t, tc.expectedOutput, strings.TrimSpace(out.String()))

			if len(tc.flagArgs) != 0 {
				require.Contains(t, fmt.Sprint(cmd), "annual-provisions [] [] Query the current minting annual provisions value")
				require.Contains(t, fmt.Sprint(cmd), tc.expCmdOutput)
			}
		})
	}
}

