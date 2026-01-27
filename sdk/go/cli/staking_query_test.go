package cli_test

import (
	"context"
	"fmt"
	"strings"

	"github.com/cosmos/gogoproto/proto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
)

func (s *StakingCLITestSuite) TestGetCmdQueryValidator() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"with invalid address ",
			cli.TestFlags().
				With("bla").
				WithOutputJSON(),
			true,
		},
		{
			"happy case",
			cli.TestFlags().
				With(sdk.ValAddress(s.addrs[0]).String()).
				WithOutputJSON(),
			false,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingValidatorCmd()
			cctx := s.cctx
			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().NotEqual("internal", err.Error())
			} else {
				var result types.Validator
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), &result))
			}
		})
	}
}

func (s *StakingCLITestSuite) TestGetCmdQueryValidators() {
	testCases := []struct {
		name              string
		args              []string
		minValidatorCount int
	}{
		{
			"one validator case",
			cli.TestFlags().
				WithLimit(1).
				WithOutputJSON(),
			1,
		},
		{
			"multi validator case",
			cli.TestFlags().
				WithOutputJSON(),
			0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingValidatorsCmd()
			cctx := s.cctx

			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)
			s.Require().NoError(err)

			var result types.QueryValidatorsResponse
			s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), &result))
		})
	}
}

func (s *StakingCLITestSuite) TestGetCmdQueryDelegation() {
	testCases := []struct {
		name     string
		args     []string
		expErr   bool
		respType proto.Message
	}{
		{
			"with wrong delegator address",
			cli.TestFlags().
				With(
					"wrongDelAddr",
					s.addrs[1].String(),
				).
				WithOutputJSON(),
			true, nil,
		},
		{
			"with wrong validator address",
			cli.TestFlags().
				With(
					s.addrs[0].String(),
					"wrongDelAddr",
				).
				WithOutputJSON(),
			true, nil,
		},
		{
			"with json output",
			cli.TestFlags().
				With(
					s.addrs[0].String(),
					sdk.ValAddress(s.addrs[1]).String(),
				).
				WithOutputJSON(),
			false,
			&types.DelegationResponse{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingDelegationCmd()
			cctx := s.cctx

			_, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().Contains(err.Error(), "Marshal called with nil")
			}
		})
	}
}

func (s *StakingCLITestSuite) TestGetCmdQueryDelegations() {
	testCases := []struct {
		name     string
		args     []string
		expErr   bool
		respType proto.Message
	}{
		{
			"with no delegator address",
			[]string{},
			true, nil,
		},
		{
			"with wrong delegator address",
			cli.TestFlags().
				With("wrongDelAddr"),
			true, nil,
		},
		{
			"valid request (height specific)",
			cli.TestFlags().
				With(
					s.addrs[0].String(),
				).
				WithOutputJSON().
				WithHeight(1),
			false,
			&types.QueryDelegatorDelegationsResponse{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingDelegationsCmd()
			cctx := s.cctx

			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *StakingCLITestSuite) TestGetCmdQueryValidatorDelegations() {
	testCases := []struct {
		name     string
		args     []string
		expErr   bool
		respType proto.Message
	}{
		{
			"with no validator address",
			[]string{},
			true, nil,
		},
		{
			"wrong validator address",
			cli.TestFlags().
				With("wrongDelAddr"),
			true, nil,
		},
		{
			"valid request(height specific)",
			cli.TestFlags().
				With(
					s.addrs[0].String(),
				).
				WithOutputJSON().
				WithHeight(1),
			false,
			&types.QueryValidatorDelegationsResponse{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingDelegationsCmd()
			cctx := s.cctx

			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *StakingCLITestSuite) TestGetCmdQueryUnbondingDelegations() {
	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			cli.TestFlags().
				With(
					"wrongDelAddr",
				).
				WithOutputJSON(),
			true,
		},
		{
			"valid request",
			cli.TestFlags().
				With(
					s.addrs[0].String(),
				).
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingUnbondingDelegationsCmd()
			cctx := s.cctx

			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var ubds types.QueryDelegatorUnbondingDelegationsResponse
				err = s.cctx.Codec.UnmarshalJSON(out.Bytes(), &ubds)

				s.Require().NoError(err)
			}
		})
	}
}

func (s *StakingCLITestSuite) TestGetCmdQueryUnbondingDelegation() {
	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			cli.TestFlags().
				With(
					"wrongDelAddr",
					s.addrs[0].String(),
				).
				WithOutputJSON(),
			true,
		},
		{
			"wrong validator address",
			cli.TestFlags().
				With(
					s.addrs[0].String(),
					"wrongValAddr",
				).
				WithOutputJSON(),
			true,
		},
		{
			"valid request",
			cli.TestFlags().
				With(
					s.addrs[0].String(),
					sdk.ValAddress(s.addrs[1]).String(),
				).
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingUnbondingDelegationCmd()
			cctx := s.cctx

			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var ubd types.UnbondingDelegation

				err = s.cctx.Codec.UnmarshalJSON(out.Bytes(), &ubd)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *StakingCLITestSuite) TestGetCmdQueryValidatorUnbondingDelegations() {
	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong validator address",
			cli.TestFlags().
				With(
					"wrongValAddr",
				).
				WithOutputJSON(),
			true,
		},
		{
			"valid request",
			cli.TestFlags().
				With(
					sdk.ValAddress(s.addrs[0]).String(),
				).
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingValidatorUnbondingDelegationsCmd()
			cctx := s.cctx

			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var ubds types.QueryValidatorUnbondingDelegationsResponse
				err = s.cctx.Codec.UnmarshalJSON(out.Bytes(), &ubds)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *StakingCLITestSuite) TestGetCmdQueryRedelegations() {
	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			cli.TestFlags().
				With(
					"wrongdeladdr",
				).
				WithOutputJSON(),
			true,
		},
		{
			"valid request",
			cli.TestFlags().
				With(
					s.addrs[0].String(),
				).
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingRedelegationsCmd()
			cctx := s.cctx

			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)
			if tc.expErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				var redelegations types.QueryRedelegationsResponse
				err = s.cctx.Codec.UnmarshalJSON(out.Bytes(), &redelegations)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *StakingCLITestSuite) TestGetCmdQueryRedelegation() {
	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong delegator address",
			cli.TestFlags().
				With(
					"wrongdeladdr",
					sdk.ValAddress(s.addrs[0]).String(),
					sdk.ValAddress(s.addrs[1]).String(),
				).
				WithOutputJSON(),
			true,
		},
		{
			"wrong source validator address address",
			cli.TestFlags().
				With(
					s.addrs[0].String(),
					"wrongSrcValAddress",
					sdk.ValAddress(s.addrs[1]).String(),
				).
				WithOutputJSON(),
			true,
		},
		{
			"wrong destination validator address address",
			cli.TestFlags().
				With(
					s.addrs[0].String(),
					sdk.ValAddress(s.addrs[0]).String(),
					"wrongSrcValAddress",
				).
				WithOutputJSON(),
			true,
		},
		{
			"valid request",
			cli.TestFlags().
				With(
					s.addrs[0].String(),
					sdk.ValAddress(s.addrs[0]).String(),
					sdk.ValAddress(s.addrs[1]).String(),
				).
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingRedelegationCmd()
			cctx := s.cctx

			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var redelegations types.QueryRedelegationsResponse

				err = s.cctx.Codec.UnmarshalJSON(out.Bytes(), &redelegations)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *StakingCLITestSuite) TestGetCmdQueryValidatorRedelegations() {
	testCases := []struct {
		name   string
		args   []string
		expErr bool
	}{
		{
			"wrong validator address",
			cli.TestFlags().
				With(
					"wrongValAddr",
				).
				WithOutputJSON(),
			true,
		},
		{
			"valid request",
			cli.TestFlags().
				With(
					sdk.ValAddress(s.addrs[0]).String(),
				).
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingValidatorRedelegationsCmd()
			cctx := s.cctx

			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)

			if tc.expErr {
				s.Require().Error(err)
			} else {
				var redelegations types.QueryRedelegationsResponse
				err = s.cctx.Codec.UnmarshalJSON(out.Bytes(), &redelegations)
				s.Require().NoError(err)
			}
		})
	}
}

func (s *StakingCLITestSuite) TestGetCmdQueryPool() {
	testCases := []struct {
		name           string
		args           []string
		expectedOutput string
	}{
		{
			"with text",
			[]string{
				fmt.Sprintf("--%s=text", cflags.FlagOutput),
				fmt.Sprintf("--%s=1", cflags.FlagHeight),
			},
			`bonded_tokens: "0"
not_bonded_tokens: "0"`,
		},
		{
			"with json",
			[]string{
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
				fmt.Sprintf("--%s=1", cflags.FlagHeight),
			},
			`{"not_bonded_tokens":"0","bonded_tokens":"0"}`,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryStakingPoolCmd()
			cctx := s.cctx
			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)
			s.Require().NoError(err)
			s.Require().Equal(tc.expectedOutput, strings.TrimSpace(out.String()))
		})
	}
}
