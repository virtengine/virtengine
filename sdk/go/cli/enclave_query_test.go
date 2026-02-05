package cli_test

import (
	"context"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/virtengine/virtengine/sdk/go/cli"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	enclavetypes "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

func (s *EnclaveCLITestSuite) mockQueryContext(msg proto.Message) client.Context {
	bz, _ := s.encCfg.Codec.Marshal(msg)
	c := testutil.NewMockCometRPC(abci.ResponseQuery{Value: bz})
	return s.baseCtx.WithClient(c)
}

func (s *EnclaveCLITestSuite) TestQueryEnclaveIdentityCmd() {
	validator := "virtengine1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqd7fy4f"

	testCases := []struct {
		name      string
		ctxGen    func() client.Context
		args      []string
		expectErr bool
	}{
		{
			"missing validator address",
			func() client.Context { return s.baseCtx },
			[]string{},
			true,
		},
		{
			"valid query",
			func() client.Context {
				return s.mockQueryContext(&enclavetypes.QueryEnclaveIdentityResponse{
					Identity: &enclavetypes.EnclaveIdentity{ValidatorAddress: validator},
				})
			},
			cli.TestFlags().
				With(validator).
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryEnclaveIdentityCmd()
			out, err := clitestutil.ExecTestCLICmd(context.Background(), tc.ctxGen(), cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resp enclavetypes.QueryEnclaveIdentityResponse
				s.Require().NoError(err)
				s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), &resp))
			}
		})
	}
}

func (s *EnclaveCLITestSuite) TestQueryEnclaveKeysCmd() {
	cmd := cli.GetQueryEnclaveKeysCmd()
	out, err := clitestutil.ExecTestCLICmd(
		context.Background(),
		s.mockQueryContext(&enclavetypes.QueryActiveValidatorEnclaveKeysResponse{Identities: []enclavetypes.EnclaveIdentity{}}),
		cmd,
		cli.TestFlags().WithOutputJSON()...,
	)

	s.Require().NoError(err)
	var resp enclavetypes.QueryActiveValidatorEnclaveKeysResponse
	s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), &resp))
}

func (s *EnclaveCLITestSuite) TestQueryEnclaveMeasurementsCmd() {
	cmd := cli.GetQueryEnclaveMeasurementsCmd()
	out, err := clitestutil.ExecTestCLICmd(
		context.Background(),
		s.mockQueryContext(&enclavetypes.QueryMeasurementAllowlistResponse{Measurements: []enclavetypes.MeasurementRecord{}}),
		cmd,
		cli.TestFlags().WithOutputJSON()...,
	)

	s.Require().NoError(err)
	var resp enclavetypes.QueryMeasurementAllowlistResponse
	s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), &resp))
}

func (s *EnclaveCLITestSuite) TestQueryEnclaveMeasurementCmd() {
	measurementHash := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	testCases := []struct {
		name      string
		ctxGen    func() client.Context
		args      []string
		expectErr bool
	}{
		{
			"missing measurement hash",
			func() client.Context { return s.baseCtx },
			[]string{},
			true,
		},
		{
			"valid query",
			func() client.Context {
				return s.mockQueryContext(&enclavetypes.QueryMeasurementResponse{
					Measurement: &enclavetypes.MeasurementRecord{},
				})
			},
			cli.TestFlags().
				With(measurementHash).
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryEnclaveMeasurementCmd()
			out, err := clitestutil.ExecTestCLICmd(context.Background(), tc.ctxGen(), cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resp enclavetypes.QueryMeasurementResponse
				s.Require().NoError(err)
				s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), &resp))
			}
		})
	}
}

func (s *EnclaveCLITestSuite) TestQueryEnclaveRotationCmd() {
	validator := "virtengine1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqd7fy4f"

	testCases := []struct {
		name      string
		ctxGen    func() client.Context
		args      []string
		expectErr bool
	}{
		{
			"missing validator address",
			func() client.Context { return s.baseCtx },
			[]string{},
			true,
		},
		{
			"valid query",
			func() client.Context {
				return s.mockQueryContext(&enclavetypes.QueryKeyRotationResponse{
					Rotation: &enclavetypes.KeyRotationRecord{ValidatorAddress: validator},
				})
			},
			cli.TestFlags().
				With(validator).
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryEnclaveRotationCmd()
			out, err := clitestutil.ExecTestCLICmd(context.Background(), tc.ctxGen(), cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resp enclavetypes.QueryKeyRotationResponse
				s.Require().NoError(err)
				s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), &resp))
			}
		})
	}
}

func (s *EnclaveCLITestSuite) TestQueryEnclaveParamsCmd() {
	cmd := cli.GetQueryEnclaveParamsCmd()
	out, err := clitestutil.ExecTestCLICmd(
		context.Background(),
		s.mockQueryContext(&enclavetypes.QueryParamsResponse{Params: enclavetypes.Params{}}),
		cmd,
		cli.TestFlags().WithOutputJSON()...,
	)

	s.Require().NoError(err)
	var resp enclavetypes.QueryParamsResponse
	s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), &resp))
}

func (s *EnclaveCLITestSuite) TestQueryEnclaveValidKeysCmd() {
	cmd := cli.GetQueryEnclaveValidKeySetCmd()
	out, err := clitestutil.ExecTestCLICmd(
		context.Background(),
		s.mockQueryContext(&enclavetypes.QueryValidKeySetResponse{ValidatorKeys: []enclavetypes.ValidatorKeyInfo{}}),
		cmd,
		cli.TestFlags().WithOutputJSON()...,
	)

	s.Require().NoError(err)
	var resp enclavetypes.QueryValidKeySetResponse
	s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), &resp))
}

func (s *EnclaveCLITestSuite) TestQueryEnclaveAttestedResultCmd() {
	validator := "virtengine1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqd7fy4f"

	testCases := []struct {
		name      string
		ctxGen    func() client.Context
		args      []string
		expectErr bool
	}{
		{
			"missing required flags",
			func() client.Context { return s.baseCtx },
			[]string{},
			true,
		},
		{
			"valid query",
			func() client.Context {
				return s.mockQueryContext(&enclavetypes.QueryAttestedResultResponse{
					Result: &enclavetypes.AttestedScoringResult{
						BlockHeight:      1,
						ScopeId:          "scope-1",
						ValidatorAddress: validator,
						AccountAddress:   validator,
						Score:            1,
						Status:           "ok",
					},
				})
			},
			cli.TestFlags().
				WithFlag("block-height", 1).
				WithFlag("scope-id", "scope-1").
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryEnclaveAttestedResultCmd()
			out, err := clitestutil.ExecTestCLICmd(context.Background(), tc.ctxGen(), cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var resp enclavetypes.QueryAttestedResultResponse
				s.Require().NoError(err)
				s.Require().NoError(s.encCfg.Codec.UnmarshalJSON(out.Bytes(), &resp))
			}
		})
	}
}
