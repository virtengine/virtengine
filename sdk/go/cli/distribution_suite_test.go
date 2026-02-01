package cli_test

import (
	"bytes"
	"context"
	"io"
	"strings"

	"cosmossdk.io/depinject"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/virtengine/virtengine/sdk/go/cli"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

type DistributionCLITestSuite struct {
	CLITestSuite
}

func (s *DistributionCLITestSuite) SetupSuite() {
	s.encCfg = sdkutil.MakeEncodingConfig(gov.AppModuleBasic{}, bank.AppModuleBasic{})
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
		WithSignModeStr("direct")

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

	cfg, err := network.DefaultConfigWithAppConfig(
		depinject.Configs(
			testutil.DistributionAppConfig,
			depinject.Provide(sdkutil.BuildCustomSigners)))
	s.Require().NoError(err)

	genesisState := cfg.GenesisState

	var mintData minttypes.GenesisState
	s.Require().NoError(cfg.Codec.UnmarshalJSON(genesisState[minttypes.ModuleName], &mintData))

	inflation := sdkmath.LegacyMustNewDecFromStr("1.0")
	mintData.Minter.Inflation = inflation
	mintData.Params.InflationMin = inflation
	mintData.Params.InflationMax = inflation

	mintDataBz, err := cfg.Codec.MarshalJSON(&mintData)
	s.Require().NoError(err)
	genesisState[minttypes.ModuleName] = mintDataBz

	cfg.GenesisState = genesisState
}

func (s *DistributionCLITestSuite) TestGetCmdQueryParams() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			cli.TestFlags().
				WithOutputJSON(),
			`{"community_tax":"0","base_proposer_reward":"0","bonus_proposer_reward":"0","withdraw_addr_enabled":false}`,
		},
		{
			"text output",
			cli.TestFlags().
				WithOutputText(),
			`base_proposer_reward: "0"
bonus_proposer_reward: "0"
community_tax: "0"
withdraw_addr_enabled: false`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryDistributionParamsCmd()

			out, err := clitestutil.ExecTestCLICmd(context.Background(), s.cctx, cmd, tc.args...)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *DistributionCLITestSuite) TestGetCmdQueryValidatorDistributionInfo() {
	addr := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	val := sdk.ValAddress(addr[0].Address.String())

	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"invalid val address",
			cli.TestFlags().
				With("invalid address").
				WithOutputJSON(),
			true,
		},
		{
			"json output",
			cli.TestFlags().
				With(val.String()).
				WithOutputJSON(),
			false,
		},
		{
			"text output",
			cli.TestFlags().
				With(val.String()).
				WithOutputText(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryDistributionValidatorDistributionInfoCmd()

			_, err := clitestutil.ExecTestCLICmd(context.Background(), s.cctx, cmd, tc.args...)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *DistributionCLITestSuite) TestGetCmdQueryValidatorOutstandingRewards() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"invalid validator address",
			cli.TestFlags().
				With("foo").
				WithHeight(3),
			true,
			"",
		},
		{
			"json output",
			cli.TestFlags().
				With(sdk.ValAddress(val[0].Address).String()).
				WithHeight(3).
				WithOutputJSON(),
			false,
			`{"rewards":[]}`,
		},
		{
			"text output",
			cli.TestFlags().
				With(sdk.ValAddress(val[0].Address).String()).
				WithHeight(3).
				WithOutputText(),
			false,
			`rewards: []`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryDistributionValidatorOutstandingRewardsCmd()

			out, err := clitestutil.ExecTestCLICmd(context.Background(), s.cctx, cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
			}
		})
	}
}

func (s *DistributionCLITestSuite) TestGetCmdQueryValidatorCommission() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"invalid validator address",
			cli.TestFlags().
				With("foo").
				WithHeight(3),
			true,
			"",
		},
		{
			"json output",
			cli.TestFlags().
				With(sdk.ValAddress(val[0].Address).String()).
				WithHeight(3).
				WithOutputJSON(),
			false,
			`{"commission":[]}`,
		},
		{
			"text output",
			cli.TestFlags().
				With(sdk.ValAddress(val[0].Address).String()).
				WithHeight(3).
				WithOutputText(),
			false,
			`commission: []`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryDistributionValidatorCommissionCmd()

			out, err := clitestutil.ExecTestCLICmd(context.Background(), s.cctx, cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
			}
		})
	}
}

func (s *DistributionCLITestSuite) TestGetCmdQueryValidatorSlashes() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"invalid validator address",
			cli.TestFlags().
				With(
					"foo",
					"1",
					"3",
				).
				WithHeight(3),
			true,
			"",
		},
		{
			"invalid start height",
			cli.TestFlags().
				With(
					sdk.ValAddress(val[0].Address).String(),
					"-1",
					"3",
				).
				WithHeight(3),
			true,
			"",
		},
		{
			"invalid end height",
			cli.TestFlags().
				With(
					sdk.ValAddress(val[0].Address).String(),
					"1",
					"-3",
				).
				WithHeight(3),
			true,
			"",
		},
		{
			"json output",
			cli.TestFlags().
				With(
					sdk.ValAddress(val[0].Address).String(),
					"1",
					"3",
				).
				WithHeight(3).
				WithOutputJSON(),
			false,
			"{\"slashes\":[],\"pagination\":null}",
		},
		{
			"text output",
			cli.TestFlags().
				With(
					sdk.ValAddress(val[0].Address).String(),
					"1",
					"3",
				).
				WithHeight(3).
				WithOutputText(),
			false,
			"pagination: null\nslashes: []",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryDistributionValidatorSlashesCmd()

			out, err := clitestutil.ExecTestCLICmd(context.Background(), s.cctx, cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
			}
		})
	}
}

func (s *DistributionCLITestSuite) TestGetCmdQueryDelegatorRewards() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	addr := val[0].Address
	valAddr := sdk.ValAddress(addr)

	testCases := []struct {
		name           string
		args           []string
		expectErr      bool
		expectedOutput string
	}{
		{
			"invalid delegator address",
			cli.TestFlags().
				With(
					"foo",
					valAddr.String(),
				).
				WithHeight(5),
			true,
			"",
		},
		{
			"invalid validator address",
			cli.TestFlags().
				With(
					addr.String(),
					"foo",
				).
				WithHeight(5),
			true,
			"",
		},
		{
			"json output",
			cli.TestFlags().
				With(
					addr.String(),
				).
				WithHeight(5).
				WithOutputJSON(),
			false,
			`{"rewards":[],"total":[]}`,
		},
		{
			"json output (specific validator)",
			cli.TestFlags().
				With(
					addr.String(),
					valAddr.String(),
				).
				WithHeight(5).
				WithOutputJSON(),
			false,
			`{"rewards":[]}`,
		},
		{
			"text output",
			cli.TestFlags().
				With(
					addr.String(),
				).
				WithHeight(5).
				WithOutputText(),
			false,
			`rewards: []
total: []`,
		},
		{
			"text output (specific validator)",
			cli.TestFlags().
				With(
					addr.String(),
					valAddr.String(),
				).
				WithHeight(5).
				WithOutputText(),
			false,
			`rewards: []`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryDistributionDelegatorRewardsCmd()

			out, err := clitestutil.ExecTestCLICmd(context.Background(), s.cctx, cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
			}
		})
	}
}

func (s *DistributionCLITestSuite) TestGetCmdQueryCommunityPool() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"json output",
			cli.TestFlags().
				WithHeight(3).
				WithOutputJSON(),
			`{"pool":[]}`,
		},
		{
			"text output",
			cli.TestFlags().
				WithHeight(3).
				WithOutputText(),
			`pool: []`,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryDistributionCommunityPoolCmd()

			out, err := clitestutil.ExecTestCLICmd(context.Background(), s.cctx, cmd, tc.args...)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}

func (s *DistributionCLITestSuite) TestNewWithdrawRewardsCmd() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name                 string
		args                 []string
		expectErr            bool
		respType             proto.Message
		expectedResponseType []string
	}{
		{
			"invalid validator address",
			cli.TestFlags().
				With(
					val[0].Address.String(),
				).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			true,
			nil,
			[]string{},
		},
		{
			"valid transaction",
			cli.TestFlags().
				With(
					sdk.ValAddress(val[0].Address).String(),
				).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false,
			&sdk.TxResponse{},
			[]string{
				"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse",
			},
		},
		{
			"valid transaction (with commission)",
			cli.TestFlags().
				With(
					sdk.ValAddress(val[0].Address).String(),
				).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithCommission().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false,
			&sdk.TxResponse{},
			[]string{
				"/cosmos.distribution.v1beta1.MsgWithdrawDelegatorRewardResponse",
				"/cosmos.distribution.v1beta1.MsgWithdrawValidatorCommissionResponse",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			bz, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cli.GetTxDistributionWithdrawRewardsCmd(), tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(bz.Bytes(), tc.respType), string(bz.Bytes()))
			}
		})
	}
}

func (s *DistributionCLITestSuite) TestNewSetWithdrawAddrCmd() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"invalid withdraw address",
			cli.TestFlags().
				With("foo").
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			true, nil,
		},
		{
			"valid transaction",
			cli.TestFlags().
				With(val[0].Address.String()).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cli.GetTxDistributionSetWithdrawAddrCmd(), tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *DistributionCLITestSuite) TestNewFundCommunityPoolCmd() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"invalid funding amount",
			cli.TestFlags().
				With("-43foocoin").
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			true, nil,
		},
		{
			"valid transaction",
			cli.TestFlags().
				With(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(5431))).String()).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cli.GetTxDistributionFundCommunityPoolCmd(), tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

