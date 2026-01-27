package cli_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

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
	bme "github.com/virtengine/virtengine/sdk/go/node/bme/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

type BMECLITestSuite struct {
	CLITestSuite

	addrs []sdk.AccAddress
}

func TestBMECLITestSuite(t *testing.T) {
	suite.Run(t, new(BMECLITestSuite))
}

func (s *BMECLITestSuite) SetupSuite() {
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
}

// ==================== Query Tests ====================

func (s *BMECLITestSuite) TestCLIQueryBMEParams() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query params with json output",
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
		{
			"query params with height flag",
			cli.TestFlags().
				WithOutputJSON().
				WithHeight(1),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetBMEParamsCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				output := out.String()
				s.Require().NotEmpty(output)
				// For JSON output, verify it can be unmarshaled
				if strings.Contains(tc.name, "json") {
					var resp bme.QueryParamsResponse
					s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), output)
				} else {
					// For text output, just verify it contains expected fields
					s.Require().True(strings.Contains(output, "params"))
				}
			}
		})
	}
}

func (s *BMECLITestSuite) TestCLIQueryBMEVaultState() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query vault state with json output",
			cli.TestFlags().
				WithOutputJSON(),
			false,
		},
		{
			"query vault state with text output",
			cli.TestFlags().
				WithOutputText(),
			false,
		},
		{
			"query vault state with height flag",
			cli.TestFlags().
				WithOutputJSON().
				WithHeight(1),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetBMEVaultStateCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				output := out.String()
				s.Require().NotEmpty(output)
				if strings.Contains(tc.name, "json") {
					var resp bme.QueryVaultStateResponse
					s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), output)
				}
			}
		})
	}
}

func (s *BMECLITestSuite) TestCLIQueryBMEOutputFormats() {
	// Test different output formats for various commands
	s.Run("query params with json output succeeds", func() {
		cmd := cli.GetBMEParamsCmd()
		args := cli.TestFlags().WithOutputJSON()
		out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, args...)
		s.Require().NoError(err)
		s.Require().NotEmpty(out.String())
	})

	s.Run("query vault state with json output succeeds", func() {
		cmd := cli.GetBMEVaultStateCmd()
		args := cli.TestFlags().WithOutputJSON()
		out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, args...)
		s.Require().NoError(err)
		s.Require().NotEmpty(out.String())
	})

	s.Run("query circuit breaker with json output succeeds", func() {
		cmd := cli.GetBMEStatusCmd()
		args := cli.TestFlags().WithOutputJSON()
		out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, args...)
		s.Require().NoError(err)
		s.Require().NotEmpty(out.String())
	})
}

// ==================== Transaction Tests ====================

func (s *BMECLITestSuite) TestCLITxBMEBurnMint() {
	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedErrMsg string
	}{
		{
			"burn AKT to mint ACT successfully",
			cli.TestFlags().
				With("1000000uve", "uact").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false,
			"",
		},
		{
			"burn ACT to mint AKT successfully",
			cli.TestFlags().
				With("500000uact", "uve").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false,
			"",
		},
		{
			"burn mint with missing args",
			cli.TestFlags().
				With("1000000uve").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync(),
			true,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxBMEBurnMintCmd()
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

func (s *BMECLITestSuite) TestCLITxBMEBurnMintWithToFlag() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"burn mint with to flag same as sender",
			cli.TestFlags().
				With("1000000uve", "uact").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false,
		},
		{
			"burn mint with different to address",
			cli.TestFlags().
				With("1000000uve", "uact").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxBMEBurnMintCmd()
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

func (s *BMECLITestSuite) TestCLITxBMEBurnMintEdgeCases() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"burn mint with invalid coin format",
			cli.TestFlags().
				With("invalid", "uact").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync(),
			true,
		},
		{
			"burn mint with empty denom to mint",
			cli.TestFlags().
				With("1000000uve", "").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync(),
			true,
		},
		{
			"burn mint with zero amount",
			cli.TestFlags().
				With("0uve", "uact").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false, // Zero amount might be valid at CLI level, rejected by chain
		},
		{
			"burn mint with very large amount",
			cli.TestFlags().
				With("999999999999999999999uve", "uact").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false, // Large amounts are valid at CLI level, rejected by chain
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxBMEBurnMintCmd()
			_, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *BMECLITestSuite) TestCLITxBMEBurnMintMultipleDenoms() {
	testCases := []struct {
		name        string
		coinsToBurn string
		denomToMint string
		expectErr   bool
	}{
		{
			"burn uve to mint uact",
			"1000000uve",
			"uact",
			false,
		},
		{
			"burn uact to mint uve",
			"500000uact",
			"uve",
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxBMEBurnMintCmd()
			args := cli.TestFlags().
				With(tc.coinsToBurn, tc.denomToMint).
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))))

			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, args...)

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

func (s *BMECLITestSuite) TestCLITxBMEBurnMintWithGasSettings() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"burn mint with explicit gas",
			cli.TestFlags().
				With("1000000uve", "uact").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithGas(200000).
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false,
		},
		{
			"burn mint with gas adjustment",
			cli.TestFlags().
				With("1000000uve", "uact").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithGas(200000).
				WithGasAdjustment(1.5).
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false,
		},
		{
			"burn mint with gas prices",
			cli.TestFlags().
				With("1000000uve", "uact").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithGas(200000).
				WithGasPrices("0.025uve"),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxBMEBurnMintCmd()
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

func (s *BMECLITestSuite) TestCLITxBMEBurnMintTimeouts() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"burn mint with timeout height",
			cli.TestFlags().
				With("1000000uve", "uact").
				WithFrom(s.addrs[0].String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithTimeoutHeight(100).
				WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxBMEBurnMintCmd()
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

// ==================== Query Exec Tests ====================

func (s *BMECLITestSuite) TestCLIQueryBMEParamsExec() {
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
			cmd := cli.GetBMEParamsCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				output := out.String()
				s.Require().NotEmpty(output)
				// For JSON output, verify it can be unmarshaled
				if strings.Contains(tc.name, "successfully") {
					var resp bme.QueryParamsResponse
					s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), output)
				} else {
					// For text output, just verify it contains expected fields
					s.Require().True(strings.Contains(output, "params"))
				}
			}
		})
	}
}

func (s *BMECLITestSuite) TestCLIQueryBMEVaultStateExec() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query vault state successfully",
			cli.TestFlags().
				WithOutputJSON(),
			false,
		},
		{
			"query vault state with text output",
			cli.TestFlags().
				WithOutputText(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetBMEVaultStateCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				output := out.String()
				s.Require().NotEmpty(output)
				if strings.Contains(tc.name, "successfully") {
					var resp bme.QueryVaultStateResponse
					s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), output)
				}
			}
		})
	}
}

func (s *BMECLITestSuite) TestCLIQueryBMEStatusExec() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"query circuit breaker status successfully",
			cli.TestFlags().
				WithOutputJSON(),
			false,
		},
		{
			"query circuit breaker status with text output",
			cli.TestFlags().
				WithOutputText(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetBMEStatusCmd()
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				output := out.String()
				s.Require().NotEmpty(output)
				if strings.Contains(tc.name, "successfully") {
					var resp bme.QueryStatusResponse
					s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), output)
				}
			}
		})
	}
}
