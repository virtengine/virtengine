package cli_test

import (
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gogoproto/proto"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
)

const (
	oneYear  = 365 * 24 * 60 * 60
	tenHours = 10 * 60 * 60
	oneHour  = 60 * 60
)

// createGrant creates a new basic allowance fee grant from granter to grantee.
func (s *FeegrantCLITestSuite) createGrant(granter, grantee sdk.Address) {
	args := cli.TestFlags().
		With(
			grantee.String(),
		).
		WithFrom(granter.String()).
		WithSpendLimit(sdk.NewCoin("uve", sdkmath.NewInt(100)).String()).
		WithExpiration(getFormattedExpiration(oneYear)).
		WithBroadcastModeSync().
		WithSkipConfirm().
		WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(100))))

	out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cli.GetTxFeegrantGrantCmd(), args...)
	s.Require().NoError(err)

	var resp sdk.TxResponse
	s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
	s.Require().Equal(resp.Code, uint32(0))
}

func (s *FeegrantCLITestSuite) TestNewCmdFeeGrant() {
	granter := s.accounts[0]
	alreadyExistedGrantee := s.addedGrantee
	cctx := s.cctx

	fromAddr, fromName, _, err := client.GetFromFields(s.baseCtx, s.kr, granter.String())
	s.Require().Equal(fromAddr, granter)
	s.Require().NoError(err)

	commonArgs := cli.TestFlags().
		WithBroadcastModeSync().
		WithSkipConfirm().
		WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10))))

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"wrong grantee address",
			cli.TestFlags().
				With(
					granter.String(),
					"wrong_grantee",
				).
				WithFrom(granter.String()).
				WithSpendLimit("100uve").
				Append(commonArgs),
			true, 0, nil,
		},
		{
			"wrong granter key name",
			cli.TestFlags().
				With(
					"ve16dun6ehcc86e03wreqqww89ey569wuj49x029x",
				).
				WithFrom("invalid_granter").
				WithSpendLimit("100uve").
				Append(commonArgs),
			true, 0, nil,
		},
		{
			"valid basic fee grant",
			cli.TestFlags().
				With(
					"ve1nph3cfzk6trsmfxkeu943nvach5qw4vwvkgceh",
				).
				WithFrom(granter.String()).
				WithSpendLimit("100uve").
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid basic fee grant with granter key name",
			cli.TestFlags().
				With(
					"ve16dun6ehcc86e03wreqqww89ey569wuj49x029x",
				).
				WithFrom(fromName).
				WithSpendLimit("100uve").
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid basic fee grant with amino",
			cli.TestFlags().
				With(
					"ve1v57fx2l2rt6ehujuu99u2fw05779m5e2qmwk9l",
				).
				WithFrom(granter.String()).
				WithSpendLimit("100uve").
				WithSignMode(cflags.SignModeLegacyAminoJSON).
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid basic fee grant without spend limit",
			cli.TestFlags().
				With(
					"ve17h5lzptx3ghvsuhk7wx4c4hnl7rsswxj9772kn",
				).
				WithFrom(granter.String()).
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid basic fee grant without expiration",
			cli.TestFlags().
				With(
					"ve16dlc38dcqt0uralyd8hksxyrny6kaeqfw3f6wu",
				).
				WithFrom(granter.String()).
				WithSpendLimit("100uve").
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid basic fee grant without spend-limit and expiration",
			cli.TestFlags().
				With(
					"ve1ku40qup9vwag4wtf8cls9mkszxfthakl6t69h0",
				).
				WithFrom(granter.String()).
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
		{
			"try to add existed grant",
			cli.TestFlags().
				With(
					alreadyExistedGrantee.String(),
				).
				WithFrom(granter.String()).
				WithSpendLimit("100uve").
				Append(commonArgs),
			false, 18, &sdk.TxResponse{},
		},
		{
			"invalid number of args(periodic fee grant)",
			cli.TestFlags().
				With(
					"ve1nph3cfzk6trsmfxkeu943nvach5qw4vwvkgceh",
				).
				WithFrom(granter.String()).
				WithSpendLimit("100uve").
				WithPeriodLimit("10uve").
				WithExpiration(getFormattedExpiration(tenHours)).
				Append(commonArgs),
			true, 0, nil,
		},
		{
			"period mentioned and period limit omitted, invalid periodic grant",
			cli.TestFlags().
				With(
					"ve1nph3cfzk6trsmfxkeu943nvach5qw4vwvkgceh",
				).
				WithFrom(granter.String()).
				WithSpendLimit("100uve").
				WithPeriod(tenHours).
				WithExpiration(getFormattedExpiration(tenHours)).
				Append(commonArgs),
			true, 0, nil,
		},
		{
			"period cannot be greater than the actual expiration(periodic fee grant)",
			cli.TestFlags().
				With(
					"ve1nph3cfzk6trsmfxkeu943nvach5qw4vwvkgceh",
				).
				WithFrom(granter.String()).
				WithSpendLimit("100uve").
				WithPeriodLimit("10uve").
				WithPeriod(tenHours).
				WithExpiration(getFormattedExpiration(oneHour)).
				Append(commonArgs),
			true, 0, nil,
		},
		{
			"valid periodic fee grant",
			cli.TestFlags().
				With(
					"ve1nph3cfzk6trsmfxkeu943nvach5qw4vwvkgceh",
				).
				WithFrom(granter.String()).
				WithSpendLimit("100uve").
				WithPeriodLimit("10uve").
				WithPeriod(oneHour).
				WithExpiration(getFormattedExpiration(tenHours)).
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid periodic fee grant without spend-limit",
			cli.TestFlags().
				With(
					"ve1vevyks8pthkscvgazc97qyfjt40m6g9x960ht0",
				).
				WithFrom(granter.String()).
				WithPeriodLimit("10uve").
				WithPeriod(oneHour).
				WithExpiration(getFormattedExpiration(tenHours)).
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid periodic fee grant without expiration",
			cli.TestFlags().
				With(
					"ve1vevyks8pthkscvgazc97qyfjt40m6g9x960ht0",
				).
				WithFrom(granter.String()).
				WithSpendLimit("100uve").
				WithPeriodLimit("10uve").
				WithPeriod(oneHour).
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
		{
			"valid periodic fee grant without spend-limit and expiration",
			cli.TestFlags().
				With(
					"ve12jfxk5q7273sv9v7cz8f4jl477q83jl6h98sqw",
				).
				WithFrom(granter.String()).
				WithPeriodLimit("10uve").
				WithPeriod(oneHour).
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
		{
			"invalid expiration",
			cli.TestFlags().
				With(
					"ve1vevyks8pthkscvgazc97qyfjt40m6g9x960ht0",
				).
				WithFrom(granter.String()).
				WithPeriodLimit("10uve").
				WithPeriod(oneHour).
				WithExpiration("invalid").
				Append(commonArgs),
			true, 0, nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.ctx, cctx, cli.GetTxFeegrantGrantCmd(), tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *FeegrantCLITestSuite) TestNewCmdRevokeFeegrant() {
	granter := s.addedGranter
	grantee := s.addedGrantee
	cctx := s.cctx

	commonArgs := cli.TestFlags().
		WithBroadcastModeSync().
		WithSkipConfirm().
		WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10))))

	// Create new fee grant specifically to test amino.
	aminoGrantee, err := sdk.AccAddressFromBech32("ve16ydaqh0fcnh4qt7a3jme4mmztm2qel5a6966q8")
	s.Require().NoError(err)
	s.createGrant(granter, aminoGrantee)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectedCode uint32
		respType     proto.Message
	}{
		{
			"invalid grantee",
			cli.TestFlags().
				With(
					grantee.String(),
				).
				WithFrom("wrong_granter").
				Append(commonArgs),
			true, 0, nil,
		},
		{
			"invalid grantee",
			cli.TestFlags().
				With(
					"wrong_grantee",
				).
				WithFrom(granter.String()).
				Append(commonArgs),
			true, 0, nil,
		},
		{
			"Valid revoke",
			cli.TestFlags().
				With(
					grantee.String(),
				).
				WithFrom(granter.String()).
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
		{
			"Valid revoke with amino",
			cli.TestFlags().
				With(
					aminoGrantee.String(),
				).
				WithFrom(granter.String()).
				WithSignMode(cflags.SignModeLegacyAminoJSON).
				Append(commonArgs),
			false, 0, &sdk.TxResponse{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.ctx, cctx, cli.GetTxFeegrantRevokeCmd(), tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *FeegrantCLITestSuite) TestTxWithFeeGrant() {
	// s.T().Skip() // TODO to re-enable in #12274

	cctx := s.cctx
	granter := s.addedGranter

	// creating an account manually (This account won't be exist in state)
	k, _, err := s.baseCtx.Keyring.NewMnemonic("grantee", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	pub, err := k.GetPubKey()
	s.Require().NoError(err)
	grantee := sdk.AccAddress(pub.Address())
	fee := sdk.NewCoin("uve", sdkmath.NewInt(100))

	args := cli.TestFlags().
		With(
			grantee.String(),
		).
		WithFrom(granter.String()).
		WithBroadcastModeSync().
		WithSkipConfirm().
		WithSpendLimit(fee.String()).
		WithFees(sdk.NewCoins(sdk.NewCoin("uve", sdkmath.NewInt(10)))).
		WithExpiration(getFormattedExpiration(oneYear))

	var res sdk.TxResponse
	out, err := clitestutil.ExecTestCLICmd(s.ctx, cctx, cli.GetTxFeegrantGrantCmd(), args...)
	s.Require().NoError(err)
	s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), &res), out.String())

	testcases := []struct {
		name       string
		flags      []string
		expErrCode uint32
	}{
		{
			name:  "granted fee allowance for an account which is not in state and creating any tx with it by using --fee-granter shouldn't fail",
			flags: cli.TestFlags().WithFrom(grantee.String()).WithFeeGranter(granter),
		},
		{
			name:       "--fee-payer should also sign the tx (direct)",
			flags:      cli.TestFlags().WithFrom(grantee.String()).WithFeePayer(granter),
			expErrCode: 4,
		},
		{
			name:       "--fee-payer should also sign the tx (amino-json)",
			flags:      cli.TestFlags().WithFrom(grantee.String()).WithFeePayer(granter).WithSignMode(cflags.SignModeLegacyAminoJSON),
			expErrCode: 4,
		},
		{
			name:  "use --fee-payer and --fee-granter together works",
			flags: cli.TestFlags().WithFrom(grantee.String()).WithFeePayer(grantee).WithFeeGranter(granter),
		},
	}

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxGovSubmitLegacyProposalCmd()

			pArgs := cli.TestFlags().
				WithTitle("Text Proposal").
				WithDescription("No desc").
				WithProposalType(govv1beta1.ProposalTypeText).
				WithSkipConfirm().
				Append(tc.flags)

			_, err = clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cmd, pArgs...)
			s.Require().NoError(err)

			var resp sdk.TxResponse
			s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())
		})
	}
}

func (s *FeegrantCLITestSuite) TestFilteredFeeAllowance() {
	granter := s.addedGranter
	k, _, err := s.baseCtx.Keyring.NewMnemonic("grantee1", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	pub, err := k.GetPubKey()
	s.Require().NoError(err)
	grantee := sdk.AccAddress(pub.Address())

	cctx := s.cctx

	args := cli.TestFlags().
		WithBroadcastModeSync().
		WithSkipConfirm().
		WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100))))

	spendLimit := sdk.NewCoin("uve", sdkmath.NewInt(1000))

	allowMsgs := strings.Join([]string{sdk.MsgTypeURL(&govv1beta1.MsgSubmitProposal{}), sdk.MsgTypeURL(&govv1.MsgVoteWeighted{})}, ",")

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"invalid granter address",
			cli.TestFlags().
				With("ve1nph3cfzk6trsmfxkeu943nvach5qw4vwvkgceh").
				WithFrom("not an address").
				WithAllowedMsgs(allowMsgs).
				WithSpendLimit(spendLimit.String()).
				Append(args),
			true, &sdk.TxResponse{}, 0,
		},
		{
			"invalid grantee address",
			cli.TestFlags().
				With("not an address").
				WithFrom(granter.String()).
				WithAllowedMsgs(allowMsgs).
				WithSpendLimit(spendLimit.String()).
				Append(args),
			true, &sdk.TxResponse{}, 0,
		},
		{
			"valid filter fee grant",
			cli.TestFlags().
				With(grantee.String()).
				WithFrom(granter.String()).
				WithAllowedMsgs(allowMsgs).
				WithSpendLimit(spendLimit.String()).
				Append(args),
			false, &sdk.TxResponse{}, 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.ctx, cctx, cli.GetTxFeegrantGrantCmd(), tc.args...)

			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}

	// exec filtered fee allowance
	cases := []struct {
		name         string
		malleate     func() error
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"valid proposal tx",
			func() error {
				pArgs := cli.TestFlags().
					WithTitle("Text Proposal").
					WithDescription("No desc").
					WithProposalType(govv1beta1.ProposalTypeText).
					WithFeeGranter(granter).
					WithFrom(grantee.String()).
					Append(args)

				out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cli.GetTxGovSubmitLegacyProposalCmd(), pArgs...)
				s.Require().NoError(err)
				var resp sdk.TxResponse
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), &resp), out.String())

				return err
			},
			&sdk.TxResponse{},
			0,
		},
		{
			"valid weighted_vote tx",
			func() error {
				sArgs := cli.TestFlags().
					With(
						"0",
						"yes",
					).
					WithFrom(grantee.String()).
					Append(args)

				out, err := clitestutil.ExecTestCLICmd(s.ctx, cctx, cli.GetTxGovWeightedVoteCmd(), sArgs...)

				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), &sdk.TxResponse{}), out.String())

				return err
			},
			&sdk.TxResponse{},
			2,
		},
		{
			"should fail with unauthorized msgs",
			func() error {
				sArgs := cli.TestFlags().
					With("ve14cm33pvnrv2497tyt8sp9yavhmw83nwewvqmk0").
					WithFrom(grantee.String()).
					WithSpendLimit("100uve").
					WithFeeGranter(granter).
					Append(args)

				out, err := clitestutil.ExecTestCLICmd(s.ctx, cctx, cli.GetTxFeegrantGrantCmd(), sArgs...)
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), &sdk.TxResponse{}), out.String())

				return err
			},
			&sdk.TxResponse{},
			7,
		},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			err := tc.malleate()
			s.Require().NoError(err)
		})
	}
}

func getFormattedExpiration(duration int64) string {
	return time.Now().Add(time.Duration(duration) * time.Second).Format(time.RFC3339)
}
