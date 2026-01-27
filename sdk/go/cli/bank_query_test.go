package cli_test

import (
	"context"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/client"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

func (s *BankCLITestSuite) TestGetBalancesCmd() {
	accounts := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name         string
		ctxGen       func() client.Context
		args         []string
		expectResult proto.Message
		expectErr    bool
	}{
		{
			"valid query",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&types.QueryAllBalancesResponse{})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			cli.TestFlags().
				With(accounts[0].Address.String()).
				WithOutputJSON(),
			&types.QueryAllBalancesResponse{},
			false,
		},
		{
			"valid query with denom",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&types.QueryBalanceResponse{
					Balance: &sdk.Coin{},
				})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			cli.TestFlags().
				With(accounts[0].Address.String()).
				WithDenom("photon").
				WithOutputJSON(),
			&sdk.Coin{},
			false,
		},
		{
			"invalid Address",
			func() client.Context {
				return s.baseCtx
			},
			cli.TestFlags().
				With("foo"),
			nil,
			true,
		},
		{
			"invalid denom",
			func() client.Context {
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Code: 1,
				})
				return s.baseCtx.WithClient(c)
			},
			cli.TestFlags().
				With(accounts[0].Address.String()).
				WithDenom("foo").
				WithOutputJSON(),
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryBankBalancesCmd()
			out, err := clitestutil.ExecTestCLICmd(context.Background(), tc.ctxGen(), cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), tc.expectResult))
			}
		})
	}
}

func (s *BankCLITestSuite) TestGetSpendableBalancesCmd() {
	accounts := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name         string
		ctxGen       func() client.Context
		args         []string
		expectResult proto.Message
		expectErr    bool
	}{
		{
			"valid query",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&types.QuerySpendableBalancesResponse{})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			[]string{
				accounts[0].Address.String(),
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
			},
			&types.QuerySpendableBalancesResponse{},
			false,
		},
		{
			"valid query with denom flag",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&types.QuerySpendableBalanceByDenomRequest{})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			[]string{
				accounts[0].Address.String(),
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
				fmt.Sprintf("--%s=photon", cflags.FlagDenom),
			},
			&types.QuerySpendableBalanceByDenomResponse{},
			false,
		},
		{
			"invalid Address",
			func() client.Context {
				return s.baseCtx
			},
			[]string{
				"foo",
			},
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryBankSpendableBalancesCmd()
			out, err := clitestutil.ExecTestCLICmd(context.Background(), tc.ctxGen(), cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), tc.expectResult))
			}
		})
	}
}

func (s *BankCLITestSuite) TestGetCmdDenomsMetadata() {
	testCases := []struct {
		name         string
		ctxGen       func() client.Context
		args         []string
		expectResult proto.Message
		expectErr    bool
	}{
		{
			"valid query",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&types.QueryDenomsMetadataResponse{})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			[]string{
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
			},
			&types.QueryDenomsMetadataResponse{},
			false,
		},
		{
			"valid query with denom",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&types.QueryDenomMetadataResponse{})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			[]string{
				fmt.Sprintf("--%s=photon", cflags.FlagDenom),
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
			},
			&types.QueryDenomMetadataResponse{},
			false,
		},
		{
			"invalid query with denom",
			func() client.Context {
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Code: 1,
				})
				return s.baseCtx.WithClient(c)
			},
			[]string{
				fmt.Sprintf("--%s=foo", cflags.FlagDenom),
			},
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryBankDenomsMetadataCmd()
			out, err := clitestutil.ExecTestCLICmd(context.Background(), tc.ctxGen(), cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), tc.expectResult))
				s.Require().NoError(err)
			}
		})
	}
}

func (s *BankCLITestSuite) TestGetCmdQueryTotalSupply() {
	testCases := []struct {
		name         string
		ctxGen       func() client.Context
		args         []string
		expectResult proto.Message
		expectErr    bool
	}{
		{
			"valid query",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&types.QueryTotalSupplyResponse{})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			[]string{
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
			},
			&types.QueryTotalSupplyResponse{},
			false,
		},
		{
			"valid query with denom",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&types.QuerySupplyOfResponse{
					Amount: sdk.Coin{},
				})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			[]string{
				fmt.Sprintf("--%s=photon", cflags.FlagDenom),
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
			},
			&sdk.Coin{},
			false,
		},
		{
			"invalid query with denom",
			func() client.Context {
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Code: 1,
				})
				return s.baseCtx.WithClient(c)
			},
			[]string{
				fmt.Sprintf("--%s=foo", cflags.FlagDenom),
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
			},
			nil,
			true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryBankTotalSupplyCmd()
			out, err := clitestutil.ExecTestCLICmd(context.Background(), tc.ctxGen(), cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), tc.expectResult))
				s.Require().NoError(err)
			}
		})
	}
}

func (s *BankCLITestSuite) TestGetCmdQuerySendEnabled() {
	testCases := []struct {
		name         string
		ctxGen       func() client.Context
		args         []string
		expectResult proto.Message
		expectErr    bool
	}{
		{
			"valid query",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&types.QuerySendEnabledResponse{
					SendEnabled: []*types.SendEnabled{},
				})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			[]string{
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
			},
			&types.QuerySendEnabledResponse{},
			false,
		},
		{
			"valid query with denoms",
			func() client.Context {
				bz, _ := s.encCfg.Codec.Marshal(&types.QuerySendEnabledResponse{
					SendEnabled: []*types.SendEnabled{},
				})
				c := testutil.NewMockCometRPC(abci.ResponseQuery{
					Value: bz,
				})
				return s.baseCtx.WithClient(c)
			},
			[]string{
				"photon",
				"stake",
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
			},
			&types.QuerySendEnabledResponse{},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryBankSendEnabledCmd()
			out, err := clitestutil.ExecTestCLICmd(context.Background(), tc.ctxGen(), cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), tc.expectResult))
				s.Require().NoError(err)
			}
		})
	}
}
