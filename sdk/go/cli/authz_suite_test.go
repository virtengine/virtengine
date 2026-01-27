package cli_test

import (
	"context"
	"fmt"
	"io"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	authz "github.com/cosmos/cosmos-sdk/x/authz/module"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclitestutil "github.com/cosmos/cosmos-sdk/x/gov/client/testutil"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gogoproto/proto"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	ev1 "github.com/virtengine/virtengine/sdk/go/node/escrow/v1"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

var (
	typeMsgSend           = banktypes.SendAuthorization{}.MsgTypeURL()
	typeMsgVote           = sdk.MsgTypeURL(&govv1.MsgVote{})
	typeMsgSubmitProposal = sdk.MsgTypeURL(&govv1.MsgSubmitProposal{})
	typeMsgDeposit        = sdk.MsgTypeURL(&ev1.DepositAuthorization{})
)

type AuthzCLITestSuite struct {
	CLITestSuite

	grantee []sdk.AccAddress
	addrs   []sdk.AccAddress
}

func (s *AuthzCLITestSuite) SetupSuite() {
	s.encCfg = sdkutil.MakeEncodingConfig(gov.AppModuleBasic{}, bank.AppModuleBasic{}, authz.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithClient(testutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	ctx := context.WithValue(context.Background(), cli.ContextTypeAddressCodec, s.encCfg.SigningOptions.AddressCodec)
	s.ctx = context.WithValue(ctx, cli.ContextTypeValidatorCodec, s.encCfg.SigningOptions.ValidatorAddressCodec)

	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := testutil.NewMockCometRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c).WithLegacyAmino(s.encCfg.Amino)
	}
	s.cctx = ctxGen().WithSignModeStr("direct")

	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	s.grantee = make([]sdk.AccAddress, 6)

	s.addrs = make([]sdk.AccAddress, 1)
	s.addrs[0] = s.createAccount("validator address")

	// Send some funds to the new account.
	// Create new account in the keyring.
	s.grantee[0] = s.createAccount("grantee1")
	s.msgSendExec(s.grantee[0])

	// create a proposal with deposit
	_, err := govclitestutil.MsgSubmitLegacyProposal(s.cctx, val[0].Address.String(),
		"Text Proposal 1", "Where is the title!?", govv1beta1.ProposalTypeText,
		fmt.Sprintf("--%s=%s", cflags.FlagDeposit, sdk.NewCoin("uakt", govv1.DefaultMinDepositTokens).String()))
	s.Require().NoError(err)

	// Create new account in the keyring.
	s.grantee[1] = s.createAccount("grantee2")
	// Send some funds to the new account.
	s.msgSendExec(s.grantee[1])

	// grant send authorization to grantee2
	out, err := clitestutil.ExecCreateGrant(
		s.ctx,
		s.cctx,
		cli.TestFlags().With(
			s.grantee[1].String(),
			"send").
			WithSpendLimit("100uakt").
			WithFrom(val[0].Address.String()).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))).
			WithExpiration(fmt.Sprintf("%d", time.Now().Add(time.Minute*time.Duration(120)).Unix()))...,
	)
	s.Require().NoError(err)

	var response sdk.TxResponse
	s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())

	// Create new account in the keyring.
	s.grantee[2] = s.createAccount("grantee3")

	// grant send authorization to grantee3
	_, err = clitestutil.ExecCreateGrant(
		s.ctx,
		s.cctx,
		cli.TestFlags().With(
			s.grantee[2].String(),
			"send").
			WithSpendLimit("100uakt").
			WithFrom(val[0].Address.String()).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))).
			WithExpiration(fmt.Sprintf("%d", time.Now().Add(time.Minute*time.Duration(120)).Unix()))...,
	)
	s.Require().NoError(err)

	// Create new accounts in the keyring.
	s.grantee[3] = s.createAccount("grantee4")
	s.msgSendExec(s.grantee[3])

	s.grantee[4] = s.createAccount("grantee5")
	s.grantee[5] = s.createAccount("grantee6")

	// grant send authorization with allow list to grantee4
	out, err = clitestutil.ExecCreateGrant(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				s.grantee[2].String(),
				"send",
			).
			WithSpendLimit("100uakt").
			WithFrom(val[0].Address.String()).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10)))).
			WithAllowList(s.grantee[4].String()).
			WithExpiration(fmt.Sprintf("%d", time.Now().Add(time.Minute*time.Duration(120)).Unix()))...,
	)
	s.Require().NoError(err)
	s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
}

func (s *AuthzCLITestSuite) createAccount(uid string) sdk.AccAddress {
	// Create a new account in the keyring.
	k, _, err := s.cctx.Keyring.NewMnemonic(uid, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	addr, err := k.GetAddress()
	s.Require().NoError(err)

	return addr
}

func (s *AuthzCLITestSuite) msgSendExec(grantee sdk.AccAddress) {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	// Send some funds to the new account.
	out, err := clitestutil.ExecSend(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				grantee.String(),
				val[0].Address.String(),
				sdk.NewCoins(
					sdk.NewCoin("uakt", sdkmath.NewInt(200)),
				).String(),
			).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(10))))...,
	)
	s.Require().NoError(err)
	s.Require().Contains(out.String(), `"code":0`)
}

func (s *AuthzCLITestSuite) TestCLITxGrantAuthorization() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	grantee := s.grantee[0]

	twoHours := time.Now().Add(time.Minute * 120).Unix()
	pastHour := time.Now().Add(-time.Minute * 60).Unix()

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		expErrMsg string
	}{
		{
			"Invalid granter Address",
			cli.TestFlags().
				With(
					"grantee_addr",
					"send",
				).
				WithSpendLimit("100uakt").
				WithFrom("granter").
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"key not found",
		},
		{
			"Invalid grantee Address",
			cli.TestFlags().
				With(
					"grantee_addr",
					"send",
				).
				WithSpendLimit("100uakt").
				WithFrom(val[0].Address.String()).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"invalid separator index",
		},
		{
			"Invalid spend limit",
			cli.TestFlags().
				With(
					grantee.String(),
					"send",
				).
				WithSpendLimit("0stake").
				WithFrom(val[0].Address.String()).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"spend-limit should be greater than zero",
		},
		{
			"Invalid expiration time",
			cli.TestFlags().
				With(
					grantee.String(),
					"send",
				).
				WithSpendLimit("100uakt").
				WithFrom(val[0].Address.String()).
				WithExpiration(fmt.Sprintf("%d", pastHour)),
			true,
			"",
		},
		{
			"fail with error invalid msg-type",
			cli.TestFlags().
				With(
					grantee.String(),
					"generic",
				).
				WithMsgType("invalid-msg-type").
				WithSpendLimit("100uakt").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			false,
			"",
		},
		{
			"invalid bond denom for tx delegate authorization allowed validators",
			cli.TestFlags().
				With(
					grantee.String(),
					"delegate",
				).
				WithAllowedValidators(sdk.ValAddress(s.addrs[0]).String()).
				WithSpendLimit("100uakt").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"invalid denom",
		},
		{
			"invalid bond denom for tx delegate authorization deny validators",
			cli.TestFlags().
				With(
					grantee.String(),
					"delegate",
				).
				WithDenyValidators(sdk.ValAddress(s.addrs[0]).String()).
				WithSpendLimit("100xyz").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"invalid denom",
		},
		{
			"invalid bond denom for tx undelegate authorization",
			cli.TestFlags().
				With(
					grantee.String(),
					"unbond",
				).
				WithAllowedValidators(sdk.ValAddress(s.addrs[0]).String()).
				WithSpendLimit("100xyz").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"invalid denom",
		},
		{
			"invalid bond denom for tx redelegate authorization",
			cli.TestFlags().
				With(
					grantee.String(),
					"redelegate",
				).
				WithAllowedValidators(sdk.ValAddress(s.addrs[0]).String()).
				WithSpendLimit("100xyz").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"invalid denom",
		},
		{
			"invalid decimal coin expression with more than single coin",
			cli.TestFlags().
				With(
					grantee.String(),
					"delegate",
				).
				WithAllowedValidators(sdk.ValAddress(s.addrs[0]).String()).
				WithSpendLimit("100uakt,20xyz").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"invalid decimal coin expression",
		},
		{
			"invalid authorization type",
			cli.TestFlags().
				With(
					grantee.String(),
					"invalid authz type",
				),
			true,
			"invalid authorization type",
		},
		{
			"Valid tx send authorization",
			cli.TestFlags().
				With(
					grantee.String(),
					"send",
				).
				WithSpendLimit("100uakt").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			false,
			"",
		},
		{
			"Valid tx send authorization with allow list",
			cli.TestFlags().
				With(
					grantee.String(),
					"send",
				).
				WithAllowList(s.grantee[1].String()).
				WithSpendLimit("100uakt").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			false,
			"",
		},
		{
			"Invalid tx send authorization with duplicate allow list",
			cli.TestFlags().
				With(
					grantee.String(),
					"send",
				).
				WithAllowList(fmt.Sprintf("%s,%s", s.grantee[1], s.grantee[1])).
				WithSpendLimit("100uakt").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"duplicate address",
		},
		{
			"Valid tx generic authorization",
			cli.TestFlags().
				With(
					grantee.String(),
					"generic",
				).
				WithMsgType(typeMsgVote).
				WithSpendLimit("100uakt").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			false,
			"",
		},
		{
			"fail when granter = grantee",
			cli.TestFlags().
				With(
					grantee.String(),
					"generic",
				).
				WithMsgType(typeMsgVote).
				WithSpendLimit("100uakt").
				WithSkipConfirm().
				WithFrom(grantee.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"grantee and granter should be different",
		},
		{
			"Valid tx with amino",
			cli.TestFlags().
				With(
					grantee.String(),
					"generic",
				).
				WithMsgType(typeMsgVote).
				WithSignMode(cflags.SignModeLegacyAminoJSON).
				WithSpendLimit("100uakt").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecCreateGrant(s.ctx, s.cctx, tc.args...)
			if tc.expectErr {
				s.Require().Error(err, out)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				var txResp sdk.TxResponse
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
			}
		})
	}
}

func (s *AuthzCLITestSuite) TestCmdRevokeAuthorizations() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	grantee := s.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	// send-authorization
	_, err := clitestutil.ExecCreateGrant(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				grantee.String(),
				"send",
			).
			WithSpendLimit("100uakt").
			WithSkipConfirm().
			WithFrom(val[0].Address.String()).
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithExpiration(fmt.Sprintf("%d", twoHours))...,
	)
	s.Require().NoError(err)

	// generic-authorization
	_, err = clitestutil.ExecCreateGrant(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				grantee.String(),
				"generic",
			).
			WithMsgType(typeMsgVote).
			WithSpendLimit("100uakt").
			WithSkipConfirm().
			WithFrom(val[0].Address.String()).
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithExpiration(fmt.Sprintf("%d", twoHours))...,
	)
	s.Require().NoError(err)

	// generic-authorization used for amino testing
	_, err = clitestutil.ExecCreateGrant(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				grantee.String(),
				"generic",
			).
			WithMsgType(typeMsgSubmitProposal).
			WithSpendLimit("100uakt").
			WithSkipConfirm().
			WithFrom(val[0].Address.String()).
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithExpiration(fmt.Sprintf("%d", twoHours)).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...,
	)
	s.Require().NoError(err)
	testCases := []struct {
		name      string
		args      []string
		respType  proto.Message
		expectErr bool
	}{
		{
			"invalid grantee address",
			cli.TestFlags().
				With(
					"invalid grantee",
					typeMsgSend,
				).
				WithFrom(val[0].Address.String()).
				WithGenerateOnly(),
			nil,
			true,
		},
		{
			"invalid granter address",
			cli.TestFlags().
				With(
					grantee.String(),
					typeMsgSend,
				).
				WithFrom("granter").
				WithGenerateOnly(),
			nil,
			true,
		},
		{
			"Valid tx send authorization",
			cli.TestFlags().
				With(
					grantee.String(),
					typeMsgSend,
				).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithBroadcastModeSync(),
			&sdk.TxResponse{},
			false,
		},
		{
			"Valid tx generic authorization",
			cli.TestFlags().
				With(
					grantee.String(),
					typeMsgVote,
				).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithBroadcastModeSync(),
			&sdk.TxResponse{},
			false,
		},
		{
			"Valid tx with amino",
			cli.TestFlags().
				With(
					grantee.String(),
					typeMsgVote,
				).
				WithFrom(val[0].Address.String()).
				WithSkipConfirm().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithBroadcastModeSync().
				WithSignMode(cflags.SignModeLegacyAminoJSON),
			&sdk.TxResponse{},
			false,
		},
	}
	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecRevokeAuthz(s.ctx, s.cctx, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *AuthzCLITestSuite) TestCLITxGrantDepositAuthorization() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	grantee := s.grantee[0]

	twoHours := time.Now().Add(time.Minute * 120).Unix()
	pastHour := time.Now().Add(-time.Minute * 60).Unix()

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		expErrMsg string
	}{
		{
			"Invalid granter Address",
			cli.TestFlags().
				With(
					"grantee_addr",
					"deposit",
				).
				WithSpendLimit("100uakt").
				WithFrom("granter").
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"key not found",
		},
		{
			"Invalid grantee Address",
			cli.TestFlags().
				With(
					"grantee_addr",
					"deposit",
				).
				WithSpendLimit("100uakt").
				WithFrom(val[0].Address.String()).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"invalid separator index",
		},
		{
			"Invalid scope",
			cli.TestFlags().
				With(
					grantee.String(),
					"deposit",
				).
				WithScope([]string{"test"}).
				WithSpendLimit("0stake").
				WithFrom(val[0].Address.String()).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"invalid scope",
		},
		{
			"Invalid spend limit",
			cli.TestFlags().
				With(
					grantee.String(),
					"deposit",
				).
				WithScope([]string{"deployment"}).
				WithSpendLimit("0stake").
				WithFrom(val[0].Address.String()).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			true,
			"spend-limit should be greater than zero",
		},
		{
			"Invalid expiration time",
			cli.TestFlags().
				With(
					grantee.String(),
					"deposit",
				).
				WithScope([]string{"deployment"}).
				WithSpendLimit("100uakt").
				WithFrom(val[0].Address.String()).
				WithExpiration(fmt.Sprintf("%d", pastHour)),
			true,
			"",
		},
		{
			"Valid tx deposit authorization (uakt)",
			cli.TestFlags().
				With(
					grantee.String(),
					"deposit",
				).
				WithMsgType(typeMsgDeposit).
				WithScope([]string{"deployment"}).
				WithSpendLimit("1000000uakt").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			false,
			"",
		},
		{
			"Valid tx deposit authorization (akt)",
			cli.TestFlags().
				With(
					grantee.String(),
					"deposit",
				).
				WithMsgType(typeMsgDeposit).
				WithScope([]string{"deployment"}).
				WithSpendLimit("1akt").
				WithSkipConfirm().
				WithFrom(val[0].Address.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithExpiration(fmt.Sprintf("%d", twoHours)),
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecCreateGrant(s.ctx, s.cctx, tc.args...)
			if tc.expectErr {
				s.Require().Error(err, out)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				var txResp sdk.TxResponse
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
			}
		})
	}
}

func (s *AuthzCLITestSuite) TestExecAuthorizationWithExpiration() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	grantee := s.grantee[0]
	tenSeconds := time.Now().Add(time.Second * time.Duration(10)).Unix()

	_, err := clitestutil.ExecCreateGrant(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				grantee.String(),
				"generic",
			).
			WithMsgType(typeMsgVote).
			WithSkipConfirm().
			WithFrom(val[0].Address.String()).
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithExpiration(fmt.Sprintf("%d", tenSeconds))...,
	)
	s.Require().NoError(err)
	// msg vote
	voteTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.gov.v1.MsgVote","proposal_id":"1","voter":"%s","option":"VOTE_OPTION_YES"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val[0].Address.String())
	execMsg := sdktestutil.WriteToNewTempFile(s.T(), voteTx)
	defer execMsg.Close()

	// waiting for authorization to expire
	time.Sleep(12 * time.Second)

	out, err := clitestutil.ExecAuthorization(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(execMsg.Name()).
			WithFrom(grantee.String()).
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithSkipConfirm()...,
	)
	s.Require().NoError(err)
	var response sdk.TxResponse
	s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
}

func (s *AuthzCLITestSuite) TestNewExecGenericAuthorized() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	grantee := s.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := clitestutil.ExecCreateGrant(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				grantee.String(),
				"generic",
			).
			WithMsgType(typeMsgVote).
			WithSkipConfirm().
			WithFrom(val[0].Address.String()).
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithExpiration(fmt.Sprintf("%d", twoHours))...,
	)

	s.Require().NoError(err)

	// msg vote
	voteTx := fmt.Sprintf(`{"body":{"messages":[{"@type":"/cosmos.gov.v1.MsgVote","proposal_id":"1","voter":"%s","option":"VOTE_OPTION_YES"}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`, val[0].Address.String())
	execMsg := sdktestutil.WriteToNewTempFile(s.T(), voteTx)
	defer execMsg.Close()

	testCases := []struct {
		name      string
		args      []string
		respType  proto.Message
		expectErr bool
	}{
		{
			"fail invalid grantee",
			cli.TestFlags().
				With(execMsg.Name()).
				WithFrom("grantee").
				WithBroadcastModeSync().
				WithGenerateOnly(),
			nil,
			true,
		},
		{
			"fail invalid json path",
			cli.TestFlags().
				With("/invalid/file.txt").
				WithFrom(grantee.String()).
				WithBroadcastModeSync(),
			nil,
			true,
		},
		{
			"valid txn",
			cli.TestFlags().
				With(execMsg.Name()).
				WithFrom(grantee.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithSkipConfirm(),
			&sdk.TxResponse{},
			false,
		},
		{
			"valid tx with amino",
			cli.TestFlags().
				With(execMsg.Name()).
				WithFrom(grantee.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithSkipConfirm().
				WithSignMode(cflags.SignModeLegacyAminoJSON),
			&sdk.TxResponse{},
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecAuthorization(context.Background(), s.cctx, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}

func (s *AuthzCLITestSuite) TestNewExecGrantAuthorized() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	grantee := s.grantee[0]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := clitestutil.ExecCreateGrant(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				grantee.String(),
				"send",
			).
			WithSpendLimit("12testtoken").
			WithSkipConfirm().
			WithFrom(val[0].Address.String()).
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithExpiration(fmt.Sprintf("%d", twoHours))...,
	)
	s.Require().NoError(err)

	tokens := sdk.NewCoins(
		sdk.NewCoin("testtoken", sdkmath.NewInt(12)),
	)
	normalGeneratedTx, err := clitestutil.ExecSend(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				val[0].Address.String(),
				grantee.String(),
				tokens.String(),
			).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithGenerateOnly()...,
	)
	s.Require().NoError(err)
	execMsg := sdktestutil.WriteToNewTempFile(s.T(), normalGeneratedTx.String())
	defer execMsg.Close()

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
	}{
		{
			"valid txn",
			cli.TestFlags().
				With(execMsg.Name()).
				WithFrom(grantee.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithSkipConfirm(),
			false,
			"",
		},
		{
			"error over spent",
			cli.TestFlags().
				With(execMsg.Name()).
				WithFrom(grantee.String()).
				WithBroadcastModeSync().
				WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
				WithSkipConfirm(),
			false,
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cctx := s.cctx

			var response sdk.TxResponse
			out, err := clitestutil.ExecAuthorization(s.ctx, cctx, tc.args...)
			switch {
			case tc.expectErrMsg != "":
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
				s.Require().Contains(response.RawLog, tc.expectErrMsg)

			case tc.expectErr:
				s.Require().Error(err)

			default:
				s.Require().NoError(err)
				s.Require().NoError(cctx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
			}
		})
	}
}

func (s *AuthzCLITestSuite) TestExecSendAuthzWithAllowList() {
	val := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	grantee := s.grantee[3]

	allowedAddr := s.grantee[4]
	notAllowedAddr := s.grantee[5]
	twoHours := time.Now().Add(time.Minute * time.Duration(120)).Unix()

	_, err := clitestutil.ExecCreateGrant(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				grantee.String(),
				"send",
			).
			WithSpendLimit("100uakt").
			WithSkipConfirm().
			WithFrom(val[0].Address.String()).
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithExpiration(fmt.Sprintf("%d", twoHours)).
			WithAllowList(allowedAddr.String())...,
	)
	s.Require().NoError(err)

	tokens := sdk.NewCoins(
		sdk.NewCoin("uakt", sdkmath.NewInt(12)),
	)

	validGeneratedTx, err := clitestutil.ExecSend(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				val[0].Address.String(),
				grantee.String(),
				tokens.String()).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithGenerateOnly()...,
	)
	s.Require().NoError(err)
	execMsg := sdktestutil.WriteToNewTempFile(s.T(), validGeneratedTx.String())
	defer execMsg.Close()

	invalidGeneratedTx, err := clitestutil.ExecSend(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				val[0].Address.String(),
				notAllowedAddr.String(),
				tokens.String()).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithGenerateOnly()...,
	)

	s.Require().NoError(err)
	execMsg1 := sdktestutil.WriteToNewTempFile(s.T(), invalidGeneratedTx.String())
	defer execMsg1.Close()

	// test sending to allowed addresses
	args := cli.TestFlags().
		With(execMsg.Name()).
		WithFrom(grantee.String()).
		WithBroadcastModeSync().
		WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
		WithSkipConfirm()

	var response sdk.TxResponse
	out, err := clitestutil.ExecAuthorization(s.ctx, s.cctx, args...)
	s.Require().NoError(err)
	s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())

	// test sending to not allowed address
	args = cli.TestFlags().
		With(execMsg1.Name()).
		WithFrom(grantee.String()).
		WithBroadcastModeSync().
		WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
		WithSkipConfirm()

	out, err = clitestutil.ExecAuthorization(s.ctx, s.cctx, args...)
	s.Require().NoError(err)
	s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
}
