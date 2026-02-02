package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	oracle "github.com/virtengine/virtengine/sdk/go/node/oracle/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

type OracleCLITestSuite struct {
	CLITestSuite

	addrs []sdk.AccAddress
}

func TestOracleCLITestSuite(t *testing.T) {
	suite.Run(t, new(OracleCLITestSuite))
}

func (s *OracleCLITestSuite) SetupSuite() {
	s.encCfg = sdkutil.MakeEncodingConfig()
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithLegacyAmino(s.encCfg.Amino).
		WithClient(testutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain").
		WithSignModeStr(cflags.SignModeDirect)

	var outBuf bytes.Buffer
	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := testutil.NewMockCometRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	s.cctx = ctxGen().WithOutput(&outBuf)

	ctx := context.WithValue(context.Background(), cli.ContextTypeAddressCodec, s.encCfg.SigningOptions.AddressCodec)
	s.ctx = context.WithValue(ctx, cli.ContextTypeValidatorCodec, s.encCfg.SigningOptions.ValidatorAddressCodec)

	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	s.addrs = make([]sdk.AccAddress, 1)
	s.addrs[0] = val[0].Address

	// Feed initial price data for uve/usd
	cmd := cli.GetTxOracleFeedPriceCmd()
	_, err := clitestutil.ExecTestCLICmd(
		s.ctx,
		s.cctx,
		cmd,
		cli.TestFlags().
			With("uve", "usd", "5.47", time.Now().Format(time.RFC3339Nano)).
			WithFrom(s.addrs[0].String()).
			WithBroadcastModeSync().
			WithSkipConfirm().
			WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))))...)
	s.Require().NoError(err)
}

func (s *OracleCLITestSuite) TestCLIQueryOraclePrices() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query all prices",
			cli.TestFlags().
				WithOutput("json"),
			false,
		},
		{
			"query prices with pagination",
			cli.TestFlags().
				WithOutput("json").
				WithLimit(10),
			false,
		},
		{
			"query prices with invalid pagination",
			cli.TestFlags().
				WithOutput("json").
				WithLimit(-1),
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetOraclePricesCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp oracle.QueryPricesResponse
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *OracleCLITestSuite) TestCLITxOracleFeedPrice() {
	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			"feed price successfully",
			cli.TestFlags().
				With("uve", "usd", "5.48", time.Now().Format(time.RFC3339Nano)).
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false,
			"",
		},
		{
			"feed price with missing args",
			cli.TestFlags().
				With("uve").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync(),
			true,
			"",
		},
		{
			"feed price without from address",
			cli.TestFlags().
				With("uve", "usd", "5.48", time.Now().Format(time.RFC3339Nano)).
				WithSkipConfirm().
				WithBroadcastModeSync(),
			true,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxOracleFeedPriceCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
				if tc.expectedErrMsg != "" {
					s.Require().Contains(err.Error(), tc.expectedErrMsg)
				}
			} else {
				s.Require().NoError(err)
				var resp sdk.TxResponse
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *OracleCLITestSuite) TestCLIQueryOraclePricesWithFilter() {
	testCases := []struct {
		name       string
		args       []string
		expectErr  bool
		checkCount bool
		minCount   int
	}{
		{
			"query prices with asset filter",
			cli.TestFlags().
				WithOutput("json").
				WithFlag(cflags.FlagAssetDenom, "ve"),
			false,
			false,
			0,
		},
		{
			"query prices with base denom filter",
			cli.TestFlags().
				WithOutput("json").
				WithFlag(cflags.FlagBaseDenom, "usd"),
			false,
			false,
			0,
		},
		{
			"query prices with both filters",
			cli.TestFlags().
				WithOutput("json").
				WithFlag(cflags.FlagAssetDenom, "ve").
				WithFlag(cflags.FlagBaseDenom, "usd"),
			false,
			false,
			0,
		},
		{
			"query prices with height filter",
			cli.TestFlags().
				WithOutput("json").
				WithHeight(100),
			false,
			false,
			0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetOraclePricesCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp oracle.QueryPricesResponse
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())

				if tc.checkCount {
					s.Require().GreaterOrEqual(len(resp.Prices), tc.minCount)
				}
			}
		})
	}
}

func (s *OracleCLITestSuite) TestCLITxOracleFeedPriceMultipleDenoms() {
	denoms := []struct {
		asset     string
		base      string
		price     string
		timestamp string
	}{
		{"ve", "usd", "5.48", time.Now().Format(time.RFC3339Nano)},
		{"usdc", "usd", "1.00", time.Now().Format(time.RFC3339Nano)},
	}

	for _, denom := range denoms {
		s.Run(fmt.Sprintf("feed_%s_%s", denom.asset, denom.base), func() {
			cmd := cli.GetTxOracleFeedPriceCmd()
			args := cli.TestFlags().
				With(denom.asset, denom.base, denom.price, denom.timestamp).
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))))

			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, args...)
			s.Require().NoError(err)

			var resp sdk.TxResponse
			s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
		})
	}
}

func (s *OracleCLITestSuite) TestCLITxOracleFeedPriceWithGasSettings() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"feed price with auto gas",
			cli.TestFlags().
				With("uve", "usd", "5.48", time.Now().Format(time.RFC3339Nano)).
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithGas(200000).
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false,
		},
		{
			"feed price with gas adjustment",
			cli.TestFlags().
				With("uve", "usd", "5.48", time.Now().Format(time.RFC3339Nano)).
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithGas(200000).
				WithGasAdjustment(1.5).
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false,
		},
		{
			"feed price with gas prices",
			cli.TestFlags().
				With("uve", "usd", "5.48", time.Now().Format(time.RFC3339Nano)).
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithGasPrices("0.025uve"),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxOracleFeedPriceCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp sdk.TxResponse
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *OracleCLITestSuite) TestCLIQueryOraclePricesPagination() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query with limit",
			cli.TestFlags().
				WithOutput("json").
				WithLimit(5),
			false,
		},
		{
			"query with offset",
			cli.TestFlags().
				WithOutput("json").
				WithOffset(10),
			false,
		},
		{
			"query with count total",
			cli.TestFlags().
				WithOutput("json").
				WithCountTotal(true),
			false,
		},
		{
			"query with reverse",
			cli.TestFlags().
				WithOutput("json").
				WithReverse(true),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetOraclePricesCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp oracle.QueryPricesResponse
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *OracleCLITestSuite) TestCLITxOracleFeedPriceTimeouts() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"feed price with timeout height",
			cli.TestFlags().
				With("uve", "usd", "5.48", time.Now().Format(time.RFC3339Nano)).
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))).
				WithTimeoutHeight(1000),
			false,
		},
		{
			"feed price with timeout duration and unordered",
			cli.TestFlags().
				With("uve", "usd", "5.48", time.Now().Format(time.RFC3339Nano)).
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))).
				WithTimeoutDuration(time.Minute).
				WithUnordered(true),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxOracleFeedPriceCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp sdk.TxResponse
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *OracleCLITestSuite) TestCLIQueryOracleInvalidQueries() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query with invalid output format",
			cli.TestFlags().
				WithOutput("invalid"),
			false, // SDK accepts any output format string without validation
		},
		{
			"query with negative limit",
			cli.TestFlags().
				WithOutput("json").
				WithLimit(-10),
			true,
		},
		{
			"query with negative offset",
			cli.TestFlags().
				WithOutput("json").
				WithFlag("offset", int64(-5)),
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetOraclePricesCmd()
			_, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *OracleCLITestSuite) TestCLITxOracleFeedPriceEdgeCases() {
	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			"feed price with empty denom",
			cli.TestFlags().
				With("", "usd").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync(),
			true,
			"",
		},
		{
			"feed price with empty base denom",
			cli.TestFlags().
				With("uve", "").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync(),
			true,
			"",
		},
		{
			"feed price with special characters in denom",
			cli.TestFlags().
				With("u@kt", "usd").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync(),
			true,
			"",
		},
		{
			"feed price with very long denom",
			cli.TestFlags().
				With("thisisaverylongdenomthatexceedsthemaximumallowedsizeofonetwentyeightcharactersthisisaverylongdenomthatexceedsthemaximumallowedsizeofonetwentyeightcharacters", "usd").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync(),
			true,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxOracleFeedPriceCmd()
			_, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
				if tc.expectedErrMsg != "" {
					s.Require().Contains(err.Error(), tc.expectedErrMsg)
				}
			}
		})
	}
}

func (s *OracleCLITestSuite) TestCLIQueryOracleParams() {
	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"json output",
			cli.TestFlags().
				WithOutputJSON(),
			"--output=json",
		},
		{
			"text output",
			cli.TestFlags().
				WithOutputText(),
			"--output=text",
		},
		{
			"with height flag",
			cli.TestFlags().
				WithOutputJSON().
				WithHeight(100),
			"--height=100",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetOracleParamsCmd()
			cmd.SetArgs(tc.args)
			s.Require().Contains(fmt.Sprint(cmd), strings.TrimSpace(tc.expCmdOutput))
		})
	}
}

func (s *OracleCLITestSuite) TestCLIQueryOraclePriceFeedConfigValidation() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"missing denom argument",
			cli.TestFlags().
				WithOutputJSON(),
			true,
		},
		{
			"valid denom",
			cli.TestFlags().
				With("uve").
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetOraclePriceFeedConfigCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var resp oracle.QueryPriceFeedConfigResponse
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *OracleCLITestSuite) TestCLIQueryOracleParamsExec() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query params successfully",
			cli.TestFlags().
				WithOutputJSON(),
			false,
		},
		{
			"query params with text output",
			cli.TestFlags().
				WithOutputText(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetOracleParamsCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				output := out.String()
				s.Require().NotEmpty(output)
				// For JSON output, verify it can be unmarshaled
				if strings.Contains(tc.name, "successfully") {
					var resp oracle.QueryParamsResponse
					s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), output)
				} else {
					// For text output, just verify it contains expected fields
					s.Require().True(strings.Contains(output, "params"))
				}
			}
		})
	}
}

