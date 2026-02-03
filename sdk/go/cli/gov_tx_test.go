package cli_test

import (
	"encoding/base64"
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gogoproto/proto"

	"github.com/virtengine/virtengine/sdk/go/cli"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
)

func (s *GovCLITestSuite) TestNewCmdSubmitProposal() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	// Create a legacy proposal JSON, make sure it doesn't pass this new CLI
	// command.
	invalidProp := `{
		"title": "",
		"description": "Where is the title!?",
		"type": "Text",
		"deposit": "-324foocoin"
	}`
	invalidPropFile := sdktestutil.WriteToNewTempFile(s.T(), invalidProp)
	defer func() {
		_ = invalidPropFile.Close()
	}()

	// Create a valid new proposal JSON.
	propMetadata := []byte{42}
	validProp := fmt.Sprintf(`
	{
		"messages": [
			{
				"@type": "/cosmos.gov.v1.MsgExecLegacyContent",
				"authority": "%s",
				"content": {
					"@type": "/cosmos.gov.v1beta1.TextProposal",
					"title": "My awesome title",
					"description": "My awesome description"
				}
			}
		],
		"title": "My awesome title",
		"summary": "My awesome description",
		"metadata": "%s",
		"deposit": "%s"
	}`, authtypes.NewModuleAddress(types.ModuleName), base64.StdEncoding.EncodeToString(propMetadata), sdk.NewCoin("uve", sdkmath.NewInt(5431)))
	validPropFile := sdktestutil.WriteToNewTempFile(s.T(), validProp)

	defer func() {
		_ = validPropFile.Close()
	}()

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"invalid proposal",
			cli.TestFlags().
				With(invalidPropFile.Name()).
				WithSkipConfirm().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			true, nil,
		},
		{
			"valid proposal",
			cli.TestFlags().
				With(validPropFile.Name()).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithSignMode("direct").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxGovSubmitProposalCmd()

			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *GovCLITestSuite) TestNewCmdSubmitLegacyProposal() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	invalidProp := `{
	  "title": "",
		"description": "Where is the title!?",
		"type": "Text",
	  "deposit": "-324foocoin"
	}`
	invalidPropFile := sdktestutil.WriteToNewTempFile(s.T(), invalidProp)

	defer func() {
		_ = invalidPropFile.Close()
	}()

	validProp := fmt.Sprintf(`{
	  "title": "Text Proposal",
		"description": "Hello, World!",
		"type": "Text",
	  "deposit": "%s"
	}`, sdk.NewCoin("uve", sdkmath.NewInt(5431)))
	validPropFile := sdktestutil.WriteToNewTempFile(s.T(), validProp)
	defer func() {
		_ = validPropFile.Close()
	}()

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		respType  proto.Message
	}{
		{
			"invalid proposal (file)",
			cli.TestFlags().
				WithProposal(invalidPropFile.Name()).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			true, nil,
		},
		{
			"invalid proposal",
			cli.TestFlags().
				WithDescription("Where is the title!?").
				WithProposalType(v1beta1.ProposalTypeText).
				WithDeposit(sdk.NewCoin("uve", sdkmath.NewInt(5431))).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			true, nil,
		},
		{
			"valid transaction (file)",
			cli.TestFlags().
				WithProposal(validPropFile.Name()).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithSignMode("direct").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, &sdk.TxResponse{},
		},
		{
			"valid transaction",
			cli.TestFlags().
				WithTitle("Text Proposal").
				WithDescription("'Where is the title!?").
				WithProposalType(v1beta1.ProposalTypeText).
				WithDeposit(sdk.NewCoin("uve", sdkmath.NewInt(5431))).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithSignMode("direct").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxGovSubmitLegacyProposalCmd()

			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *GovCLITestSuite) TestNewCmdDeposit() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"without proposal id",
			cli.TestFlags().
				With(sdk.NewCoin("uve", sdkmath.NewInt(10)).String()). // 10uve
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			true,
		},
		{
			"without deposit amount",
			cli.TestFlags().
				With("1").
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			true,
		},
		{
			"deposit on a proposal",
			cli.TestFlags().
				With("1", sdk.NewCoin("uve", sdkmath.NewInt(10)).String()).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithSignMode("direct").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false,
		},
	}

	for _, tc := range testCases {
		var resp sdk.TxResponse

		s.Run(tc.name, func() {
			cmd := cli.GetTxGovDepositCmd()

			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
			}
		})
	}
}

func (s *GovCLITestSuite) TestNewCmdVote() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
	}{
		{
			"invalid vote",
			[]string{},
			true, 0,
		},
		{
			"vote for invalid proposal",
			cli.TestFlags().
				With("10", "yes").
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithMetadata("AQ==").
				WithSignMode("direct").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, 3,
		},
		{
			"valid vote",
			cli.TestFlags().
				With("1", "yes").
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithSignMode("direct").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, 0,
		},
		{
			"valid vote with metadata",
			cli.TestFlags().
				With("1", "yes").
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithSignMode("direct").
				WithMetadata("AQ==").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxGovVoteCmd()
			var txResp sdk.TxResponse

			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
			}
		})
	}
}

func (s *GovCLITestSuite) TestNewCmdWeightedVote() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
	}{
		{
			"invalid vote",
			[]string{},
			true, 0,
		},
		{
			"vote for invalid proposal",
			cli.TestFlags().
				With("10", "yes").
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithMetadata("AQ==").
				WithSignMode("direct").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, 3,
		},
		{
			"valid vote",
			cli.TestFlags().
				With("1", "yes").
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithSignMode("direct").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, 0,
		},
		{
			"valid vote with metadata",
			cli.TestFlags().
				With("1", "yes").
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithSignMode("direct").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, 0,
		},
		{
			"invalid valid split vote string",
			cli.TestFlags().
				With("1", "yes/0.6,no/0.3,abstain/0.05,no_with_veto/0.05").
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithSignMode("direct").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			true, 0,
		},
		{
			"valid split vote",
			cli.TestFlags().
				With("1", "yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05").
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithSignMode("direct").
				WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))),
			false, 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxGovWeightedVoteCmd()
			var txResp sdk.TxResponse

			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
			}
		})
	}
}
