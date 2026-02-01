package cli_test

import (
	"bytes"
	"context"
	"io"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"

	"cosmossdk.io/x/nft/module"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

type SlashingCLITestSuite struct {
	CLITestSuite

	pub  types.PubKey
	addr sdk.AccAddress
}

func (s *SlashingCLITestSuite) SetupSuite() {
	s.encCfg = sdkutil.MakeEncodingConfig(module.AppModuleBasic{})
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

	k, _, err := s.cctx.Keyring.NewMnemonic("NewValidator", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	pub, err := k.GetPubKey()
	s.Require().NoError(err)

	s.pub = pub
	s.addr = sdk.AccAddress(pub.Address())
}

func (s *SlashingCLITestSuite) TestGetCmdQuerySigningInfo() {
	pubKeyBz, err := s.encCfg.Codec.MarshalInterfaceJSON(s.pub)
	s.Require().NoError(err)
	pubKeyStr := string(pubKeyBz)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{"invalid address", []string{"foo"}, true},
		{
			"valid address (json output)",
			cli.TestFlags().
				With(pubKeyStr).
				WithHeight(1).
				WithOutputJSON(),
			false,
		},
		{
			"valid address (text output)",
			cli.TestFlags().
				With(pubKeyStr).
				WithHeight(1).
				WithOutputText(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQuerySlashingSigningInfoCmd()
			cctx := s.cctx

			_, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *SlashingCLITestSuite) TestGetCmdQueryParams() {
	testCases := []struct {
		name string
		args []string
	}{
		{
			"json output",
			cli.TestFlags().
				WithOutputJSON(),
		},
		{
			"text output",
			cli.TestFlags().
				WithOutputText(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQuerySlashingParamsCmd()
			cctx := s.cctx

			_, err := clitestutil.ExecTestCLICmd(context.Background(), cctx, cmd, tc.args...)
			s.Require().NoError(err)
		})
	}
}

