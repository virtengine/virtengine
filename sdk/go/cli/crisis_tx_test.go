package cli_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/require"

	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

func TestNewMsgVerifyInvariantTxCmd(t *testing.T) {
	encCfg := sdkutil.MakeEncodingConfig(crisis.AppModuleBasic{})
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

	accounts := sdktestutil.CreateKeyringAccounts(t, kr, 1)
	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		errString    string
		expectedCode uint32
	}{
		{
			"missing module",
			cli.TestFlags().
				With(
					"",
					"total-supply",
				).
				WithFrom(accounts[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			true, "invalid module name", 0,
		},
		{
			"missing invariant route",
			cli.TestFlags().
				With(
					"bank",
					"",
				).
				WithFrom(accounts[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			true, "invalid invariant route", 0,
		},
		{
			"valid transaction",
			cli.TestFlags().
				With(
					"bank",
					"total-supply",
				).
				WithFrom(accounts[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false, "", 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := cli.GetTxCrisisVerifyInvariantTxCmd()

			_, err := clitestutil.ExecTestCLICmd(ctx, baseCtx, cmd, tc.args...)
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errString)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
