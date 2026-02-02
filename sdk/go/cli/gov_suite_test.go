package cli_test

import (
	"bytes"
	"context"
	"io"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

type GovCLITestSuite struct {
	CLITestSuite
}

func (s *GovCLITestSuite) SetupSuite() {
	s.encCfg = sdkutil.MakeEncodingConfig(gov.AppModuleBasic{})
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

	// create a proposal with deposit
	cmd := cli.GetTxGovSubmitLegacyProposalCmd()

	_, err := clitestutil.ExecTestCLICmd(
		s.ctx,
		s.cctx,
		cmd,
		cli.TestFlags().
			WithFrom(val[0].Address.String()).
			WithTitle("Text Proposal 1").
			WithDescription("Where is the title!?").
			WithProposalType(v1beta1.ProposalTypeText).
			WithDeposit(sdk.NewCoin("uve", cli.DefaultMinDepositTokens)).
			WithBroadcastModeSync().
			WithSkipConfirm()...)
	s.Require().NoError(err)

	// vote for proposal
	cmd = cli.GetTxGovVoteCmd()

	_, err = clitestutil.ExecTestCLICmd(
		s.ctx,
		s.cctx,
		cmd,
		cli.TestFlags().
			With(
				"1",
				"yes",
			).
			WithFrom(val[0].Address.String()).
			WithBroadcastModeSync().
			WithSkipConfirm()...)
	s.Require().NoError(err)

	// create a proposal without deposit
	cmd = cli.GetTxGovSubmitLegacyProposalCmd()

	_, err = clitestutil.ExecTestCLICmd(
		s.ctx,
		s.cctx,
		cmd,
		cli.TestFlags().
			WithFrom(val[0].Address.String()).
			WithTitle("Text Proposal 2").
			WithDescription("Where is the title!?").
			WithProposalType(v1beta1.ProposalTypeText).
			WithBroadcastModeSync().
			WithSkipConfirm()...)
	s.Require().NoError(err)

	// create a proposal3 with deposit
	cmd = cli.GetTxGovSubmitLegacyProposalCmd()

	_, err = clitestutil.ExecTestCLICmd(
		s.ctx,
		s.cctx,
		cmd,
		cli.TestFlags().
			WithFrom(val[0].Address.String()).
			WithTitle("Text Proposal 3").
			WithDescription("Where is the title!?").
			WithProposalType(v1beta1.ProposalTypeText).
			WithDeposit(sdk.NewCoin("uve", cli.DefaultMinDepositTokens)).
			WithBroadcastModeSync().
			WithSkipConfirm()...)
	s.Require().NoError(err)

	// vote for proposal3 as val
	cmd = cli.GetTxGovWeightedVoteCmd()

	_, err = clitestutil.ExecTestCLICmd(
		s.ctx,
		s.cctx,
		cmd,
		cli.TestFlags().
			With(
				"3",
				"yes=0.6,no=0.3,abstain=0.05,no_with_veto=0.05",
			).
			WithFrom(val[0].Address.String()).
			WithBroadcastModeSync().
			WithSkipConfirm()...)
	s.Require().NoError(err)
}
