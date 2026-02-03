package cli_test

import (
	"context"
	"fmt"

	"cosmossdk.io/x/feegrant"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
)

func (s *FeegrantCLITestSuite) TestCmdGetFeeGrant() {
	granter := s.addedGranter
	grantee := s.addedGrantee

	testCases := []struct {
		name         string
		args         []string
		expectErrMsg string
		expectErr    bool
		respType     *feegrant.QueryAllowanceResponse
		resp         *feegrant.Grant
	}{
		{
			"wrong granter",
			[]string{
				"wrong_granter",
				grantee.String(),
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
			},
			"decoding bech32 failed",
			true, nil, nil,
		},
		{
			"wrong grantee",
			[]string{
				granter.String(),
				"wrong_grantee",
				fmt.Sprintf("--%s=json", cflags.FlagOutput),
			},
			"decoding bech32 failed",
			true, nil, nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryFeeGrantCmd()
			out, err := clitestutil.ExecTestCLICmd(context.Background(), s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *FeegrantCLITestSuite) TestCmdGetFeeGrantsByGrantee() {
	grantee := s.addedGrantee
	cctx := s.cctx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		resp         *feegrant.QueryAllowancesResponse
		expectLength int
	}{
		{
			"wrong grantee",
			cli.TestFlags().
				With("wrong_grantee").
				WithOutputJSON(),
			true, nil, 0,
		},
		{
			"valid req",
			cli.TestFlags().
				With(grantee.String()).
				WithOutputJSON(),
			false, &feegrant.QueryAllowancesResponse{}, 1,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryFeeGrantsByGranteeCmd()

			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), tc.resp), out.String())
			}
		})
	}
}

func (s *FeegrantCLITestSuite) TestCmdGetFeeGrantsByGranter() {
	granter := s.addedGranter
	cctx := s.cctx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		resp         *feegrant.QueryAllowancesByGranterResponse
		expectLength int
	}{
		{
			"wrong granter",
			cli.TestFlags().
				With("wrong_granter").
				WithOutputJSON(),
			true, nil, 0,
		},
		{
			"valid req",
			cli.TestFlags().
				With(granter.String()).
				WithOutputJSON(),
			false, &feegrant.QueryAllowancesByGranterResponse{}, 1,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryFeeGrantsByGranterCmd()
			out, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), tc.resp), out.String())
			}
		})
	}
}
