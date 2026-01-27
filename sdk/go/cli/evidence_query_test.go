package cli_test

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"

	"cosmossdk.io/x/evidence"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

func TestGetQueryCmd(t *testing.T) {
	encCfg := sdkutil.MakeEncodingConfig(evidence.AppModuleBasic{})
	kr := keyring.NewInMemory(encCfg.Codec)
	baseCtx := client.Context{}.
		WithKeyring(kr).
		WithTxConfig(encCfg.TxConfig).
		WithCodec(encCfg.Codec).
		WithLegacyAmino(encCfg.Amino).
		WithClient(testutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain").
		WithSignModeStr(cflags.SignModeDirect)

	ctx := context.WithValue(context.Background(), cli.ContextTypeAddressCodec, encCfg.SigningOptions.AddressCodec)
	ctx = context.WithValue(ctx, cli.ContextTypeValidatorCodec, encCfg.SigningOptions.ValidatorAddressCodec)

	testCases := map[string]struct {
		args           []string
		ctxGen         func() client.Context
		expCmdOutput   string
		expectedOutput string
		expectErr      bool
	}{
		"non-existent evidence": {
			cli.TestFlags().
				With("DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660"),
			func() client.Context {
				bz, _ := encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return baseCtx.WithClient(c)
			},
			"DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660",
			"",
			true,
		},
		"all evidence (default pagination)": {
			cli.TestFlags().
				WithOutputText(),
			func() client.Context {
				bz, _ := encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return baseCtx.WithClient(c)
			},
			"",
			"evidence: []\npagination: null",
			false,
		},
		"all evidence (json output)": {
			cli.TestFlags().
				WithOutputJSON(),
			func() client.Context {
				bz, _ := encCfg.Codec.Marshal(&sdk.TxResponse{})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return baseCtx.WithClient(c)
			},
			"",
			`{"evidence":[],"pagination":null}`,
			false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			cmd := cli.GetQueryEvidenceCmd()

			if len(tc.args) != 0 {
				require.Contains(t, fmt.Sprint(cmd), tc.expCmdOutput)
			}

			out, err := clitestutil.ExecTestCLICmd(ctx, tc.ctxGen(), cmd, tc.args...)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Contains(t, fmt.Sprint(cmd), "evidence [] [] Query for evidence by hash or for all (paginated) submitted evidence")
			require.Contains(t, strings.TrimSpace(out.String()), tc.expectedOutput)
		})
	}
}
