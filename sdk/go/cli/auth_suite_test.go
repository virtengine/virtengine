package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/virtengine/virtengine/sdk/go/cli"
	cflags "github.com/virtengine/virtengine/sdk/go/cli/flags"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	"github.com/virtengine/virtengine/sdk/go/sdkutil"
	"github.com/virtengine/virtengine/sdk/go/testutil"
)

type AuthCLITestSuite struct {
	CLITestSuite
	val  sdk.AccAddress
	val1 sdk.AccAddress
}

func (s *AuthCLITestSuite) SetupSuite() {
	s.encCfg = sdkutil.MakeEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}, gov.AppModuleBasic{})
	s.kr = keyring.NewInMemory(s.encCfg.Codec)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Codec).
		WithLegacyAmino(s.encCfg.Amino).
		WithClient(testutil.MockCometRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID("test-chain")

	var outBuf bytes.Buffer
	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Codec.Marshal(&sdk.TxResponse{})
		c := testutil.NewMockCometRPC(abci.ResponseQuery{
			Value: bz,
		})
		return s.baseCtx.WithClient(c)
	}
	s.cctx = ctxGen().WithOutput(&outBuf).WithSignModeStr("direct")

	ctx := context.WithValue(context.Background(), cli.ContextTypeAddressCodec, s.encCfg.SigningOptions.AddressCodec)
	s.ctx = context.WithValue(ctx, cli.ContextTypeValidatorCodec, s.encCfg.SigningOptions.ValidatorAddressCodec)

	kb := s.cctx.Keyring
	valAcc, _, err := kb.NewMnemonic("newAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	s.val, err = valAcc.GetAddress()
	s.Require().NoError(err)

	account1, _, err := kb.NewMnemonic("newAccount1", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	s.val1, err = account1.GetAddress()
	s.Require().NoError(err)

	account2, _, err := kb.NewMnemonic("newAccount2", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)
	pub1, err := account1.GetPubKey()
	s.Require().NoError(err)
	pub2, err := account2.GetPubKey()
	s.Require().NoError(err)

	// Create a dummy account for testing purpose
	_, _, err = kb.NewMnemonic("dummyAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	s.Require().NoError(err)

	multi := kmultisig.NewLegacyAminoPubKey(2, []cryptotypes.PubKey{pub1, pub2})
	_, err = kb.SaveMultisig("multi", multi)
	s.Require().NoError(err)
}

func (s *AuthCLITestSuite) TestCLIValidateSignatures() {
	sendTokens := sdk.NewCoins(
		sdk.NewCoin("testtoken", sdkmath.NewInt(10)),
		sdk.NewCoin("uakt", sdkmath.NewInt(10)))

	res, err := s.createBankMsg(
		s.cctx,
		s.val,
		sendTokens,
		cli.TestFlags().WithGenerateOnly()...)
	s.Require().NoError(err)

	// write  unsigned tx to file
	unsignedTx := sdktestutil.WriteToNewTempFile(s.T(), res.String())
	defer func() {
		_ = unsignedTx.Close()
	}()

	res, err = clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				unsignedTx.Name(),
			).
			WithFrom(s.val.String()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)
	signedTx, err := s.cctx.TxConfig.TxJSONDecoder()(res.Bytes())
	s.Require().NoError(err)

	signedTxFile := sdktestutil.WriteToNewTempFile(s.T(), res.String())
	defer func() {
		_ = signedTxFile.Close()
	}()

	txBuilder, err := s.cctx.TxConfig.WrapTxBuilder(signedTx)
	s.Require().NoError(err)
	_, err = clitestutil.TxValidateSignaturesExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				signedTxFile.Name(),
			)...)
	s.Require().NoError(err)

	txBuilder.SetMemo("MODIFIED TX")
	bz, err := s.cctx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	modifiedTxFile := sdktestutil.WriteToNewTempFile(s.T(), string(bz))
	defer func() {
		_ = modifiedTxFile.Close()
	}()

	_, err = clitestutil.TxValidateSignaturesExec(s.ctx, s.cctx, modifiedTxFile.Name())
	s.Require().EqualError(err, "signatures validation failed")
}

func (s *AuthCLITestSuite) TestCLISignBatch() {
	sendTokens := sdk.NewCoins(
		sdk.NewCoin("testtoken", sdkmath.NewInt(10)),
		sdk.NewCoin("uakt", sdkmath.NewInt(10)),
	)

	generatedStd, err := s.createBankMsg(
		s.cctx,
		s.val,
		sendTokens,
		fmt.Sprintf("--%s=true", cflags.FlagGenerateOnly))
	s.Require().NoError(err)

	outputFile := sdktestutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String(), 3))
	defer func() {
		_ = outputFile.Close()
	}()
	s.cctx.HomeDir = strings.Replace(s.cctx.HomeDir, "simd", "simcli", 1)

	// sign-batch file - offline is set but account-number and sequence are not
	_, err = clitestutil.TxSignBatchExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				outputFile.Name(),
			).
			WithFrom(s.val.String()).
			WithChainID(s.cctx.ChainID).
			WithOffline()...)
	s.Require().EqualError(err, "required flag(s) \"account-number\", \"sequence\" not set")

	// sign-batch file - offline and sequence is set but account-number is not set
	_, err = clitestutil.TxSignBatchExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				outputFile.Name(),
			).
			WithFrom(s.val.String()).
			WithChainID(s.cctx.ChainID).
			WithOffline().
			WithSequence(1)...)
	s.Require().EqualError(err, "required flag(s) \"account-number\" not set")

	// sign-batch file - offline and account-number is set but sequence is not set
	_, err = clitestutil.TxSignBatchExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				outputFile.Name(),
			).
			WithFrom(s.val.String()).
			WithChainID(s.cctx.ChainID).
			WithOffline().
			WithAccountNumber(1)...)
	s.Require().EqualError(err, "required flag(s) \"sequence\" not set")
}

func (s *AuthCLITestSuite) TestCLIQueryTxCmdByHash() {
	sendTokens := sdk.NewInt64Coin("uakt", 10)

	// Send coins.
	out, err := s.createBankMsg(
		s.cctx, s.val,
		sdk.NewCoins(sendTokens),
	)
	s.Require().NoError(err)

	var txRes sdk.TxResponse
	s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &txRes))

	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"not enough args",
			[]string{},
			"",
		},
		{
			"with invalid hash",
			cli.TestFlags().
				With("somethinginvalid").
				WithOutputJSON(),
			`[somethinginvalid --output=json]`,
		},
		{
			"with valid and not existing hash",
			cli.TestFlags().
				With("C7E7D3A86A17AB3A321172239F3B61357937AF0F25D9FA4D2F4DCCAD9B0D7747").
				WithOutputJSON(),
			`[C7E7D3A86A17AB3A321172239F3B61357937AF0F25D9FA4D2F4DCCAD9B0D7747 --output=json`,
		},
		{
			"happy case",
			cli.TestFlags().
				With(txRes.TxHash).
				WithOutputJSON(),
			fmt.Sprintf("%s --%s=json", txRes.TxHash, cflags.FlagOutput),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryAuthTxCmd()
			cmd.SetArgs(tc.args)

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}
		})
	}
}

func (s *AuthCLITestSuite) TestCLIQueryTxCmdByEvents() {
	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"invalid --type",
			cli.TestFlags().
				WithType("foo").
				With("bar").
				WithOutputJSON(),
			"--type=foo bar --output=json",
		},
		{
			"--type=acc_seq with no addr+seq",
			cli.TestFlags().
				WithType("acc_seq").
				With("").
				WithOutputJSON(),
			"--type=acc_seq  --output=json",
		},
		{
			"non-existing addr+seq combo",
			cli.TestFlags().
				WithType("acc_seq").
				With("foobar").
				WithOutputJSON(),
			"--type=acc_seq foobar --output=json",
		},
		{
			"--type=signature with no signature",
			cli.TestFlags().
				WithType("signature").
				With("").
				WithOutputJSON(),
			"--type=signature  --output=json",
		},
		{
			"non-existing signatures",
			cli.TestFlags().
				WithType("signature").
				With("foo").
				WithOutputJSON(),
			"--type=signature foo --output=json",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryAuthTxCmd()
			cmd.SetArgs(tc.args)

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}
		})
	}
}

func (s *AuthCLITestSuite) TestCLIQueryTxsCmdByEvents() {
	testCases := []struct {
		name         string
		args         []string
		expCmdOutput string
	}{
		{
			"fee event happy case",
			cli.TestFlags().
				WithEvents(fmt.Sprintf("tx.fee=%s", sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10))).String())).
				WithOutputJSON(),
			"",
		},
		{
			"no matching fee event",
			cli.TestFlags().
				WithEvents(fmt.Sprintf("tx.fee=%s", sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(0))).String())).
				WithOutputJSON(),
			"",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryAuthTxsByEventsCmd()

			if len(tc.args) != 0 {
				s.Require().Contains(fmt.Sprint(cmd), tc.expCmdOutput)
			}
		})
	}
}

func (s *AuthCLITestSuite) TestCLISendGenerateSignAndBroadcast() {
	sendTokens := sdk.NewCoin("uakt", sdk.TokensFromConsensusPower(10, cli.DefaultPowerReduction))

	normalGeneratedTx, err := s.createBankMsg(
		s.cctx,
		s.val,
		sdk.NewCoins(sendTokens),
		cli.TestFlags().
			WithGas(cflags.DefaultGasLimit).
			WithGenerateOnly()...)
	s.Require().NoError(err)

	txCfg := s.cctx.TxConfig

	normalGeneratedStdTx, err := txCfg.TxJSONDecoder()(normalGeneratedTx.Bytes())
	s.Require().NoError(err)

	txBuilder, err := txCfg.WrapTxBuilder(normalGeneratedStdTx)
	s.Require().NoError(err)
	s.Require().Equal(txBuilder.GetTx().GetGas(), uint64(cflags.DefaultGasLimit))
	s.Require().Equal(len(txBuilder.GetTx().GetMsgs()), 1)

	sigs, err := txBuilder.GetTx().GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Equal(0, len(sigs))

	// Test generate sendTx with --gas=$amount
	limitedGasGeneratedTx, err := s.createBankMsg(
		s.cctx,
		s.val,
		sdk.NewCoins(sendTokens),
		cli.TestFlags().
			WithGas(100).
			WithGenerateOnly()...)
	s.Require().NoError(err)

	limitedGasStdTx, err := txCfg.TxJSONDecoder()(limitedGasGeneratedTx.Bytes())
	s.Require().NoError(err)

	txBuilder, err = txCfg.WrapTxBuilder(limitedGasStdTx)
	s.Require().NoError(err)
	s.Require().Equal(txBuilder.GetTx().GetGas(), uint64(100))
	s.Require().Equal(len(txBuilder.GetTx().GetMsgs()), 1)

	sigs, err = txBuilder.GetTx().GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Equal(0, len(sigs))

	// Test generate sendTx, estimate gas
	finalGeneratedTx, err := s.createBankMsg(
		s.cctx,
		s.val,
		sdk.NewCoins(sendTokens),
		cli.TestFlags().
			WithGas(cflags.DefaultGasLimit).
			WithGenerateOnly()...)
	s.Require().NoError(err)

	finalStdTx, err := txCfg.TxJSONDecoder()(finalGeneratedTx.Bytes())
	s.Require().NoError(err)

	txBuilder, err = txCfg.WrapTxBuilder(finalStdTx)
	s.Require().NoError(err)
	s.Require().Equal(uint64(cflags.DefaultGasLimit), txBuilder.GetTx().GetGas())
	s.Require().Equal(len(finalStdTx.GetMsgs()), 1)

	// Write the output to disk
	unsignedTxFile := sdktestutil.WriteToNewTempFile(s.T(), finalGeneratedTx.String())
	defer func() {
		_ = unsignedTxFile.Close()
	}()

	// Test validate-signatures
	res, err := clitestutil.TxValidateSignaturesExec(
		s.ctx,
		s.cctx,
		unsignedTxFile.Name())
	s.Require().EqualError(err, "signatures validation failed")
	s.Require().True(strings.Contains(res.String(), fmt.Sprintf("Signers:\n  0: %v\n\nSignatures:\n\n", s.val.String())))

	// Test sign

	// Does not work in offline mode
	_, err = clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(unsignedTxFile.Name()).
			WithFrom(s.val.String()).
			WithOffline()...)
	s.Require().EqualError(err, "required flag(s) \"account-number\", \"sequence\" not set")

	// But works offline if we set account number and sequence
	s.cctx.HomeDir = strings.Replace(s.cctx.HomeDir, "simd", "simcli", 1)
	_, err = clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(unsignedTxFile.Name()).
			WithFrom(s.val.String()).
			WithOffline().
			WithAccountNumber(1).
			WithSequence(1)...)
	s.Require().NoError(err)

	// Sign transaction
	signedTx, err := clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(unsignedTxFile.Name()).
			WithFrom(s.val.String())...)
	s.Require().NoError(err)

	signedFinalTx, err := txCfg.TxJSONDecoder()(signedTx.Bytes())
	s.Require().NoError(err)

	txBuilder, err = s.cctx.TxConfig.WrapTxBuilder(signedFinalTx)
	s.Require().NoError(err)
	s.Require().Equal(len(txBuilder.GetTx().GetMsgs()), 1)

	sigs, err = txBuilder.GetTx().GetSignaturesV2()
	s.Require().NoError(err)
	s.Require().Equal(1, len(sigs))
	signers, err := txBuilder.GetTx().GetSigners()
	s.Require().NoError(err)
	s.Require().Equal(s.val.Bytes(), signers[0])

	// Write the output to disk
	signedTxFile := sdktestutil.WriteToNewTempFile(s.T(), signedTx.String())
	defer func() {
		_ = signedTxFile.Close()
	}()

	// validate Signature
	res, err = clitestutil.TxValidateSignaturesExec(
		s.ctx,
		s.cctx,
		signedTxFile.Name())
	s.Require().NoError(err)
	s.Require().True(strings.Contains(res.String(), "[OK]"))

	// Test broadcast

	// Does not work in offline mode
	_, err = clitestutil.TxBroadcastExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(signedTxFile.Name()).
			WithFrom(s.val.String()).
			WithOffline()...)
	s.Require().EqualError(err, "cannot broadcast tx during offline mode")

	// Broadcast correct transaction.
	_, err = clitestutil.TxBroadcastExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(signedTxFile.Name()).
			WithFrom(s.val.String()).
			WithBroadcastModeSync()...)
	s.Require().NoError(err)
}

func (s *AuthCLITestSuite) TestCLIMultisignInsufficientCosigners() {
	// Fetch account and a multisig info
	account1, err := s.cctx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	multisigRecord, err := s.cctx.Keyring.Key("multi")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)
	// Send coins from validator to multisig.
	_, err = s.createBankMsg(
		s.cctx,
		addr,
		sdk.NewCoins(
			sdk.NewInt64Coin("uakt", 10),
		),
	)
	s.Require().NoError(err)

	// Generate multisig transaction.
	multiGeneratedTx, err := clitestutil.ExecSend(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				addr.String(),
				s.val.String(),
				sdk.NewCoins(
					sdk.NewInt64Coin("uakt", 5),
				).String(),
			).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithGenerateOnly()...)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := sdktestutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer func() {
		_ = multiGeneratedTxFile.Close()
	}()

	// Multisign, sign with one signature
	s.cctx.HomeDir = strings.Replace(s.cctx.HomeDir, "simd", "simcli", 1)
	addr1, err := account1.GetAddress()
	s.Require().NoError(err)

	account1Signature, err := clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				multiGeneratedTxFile.Name(),
			).
			WithFrom(addr1.String()).
			WithSignMode(cflags.SignModeLegacyAminoJSON).
			WithMultisig(addr.String())...)
	s.Require().NoError(err)

	sign1File := sdktestutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer func() {
		_ = sign1File.Close()
	}()

	multiSigWith1Signature, err := clitestutil.TxMultiSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				multiGeneratedTxFile.Name(),
				multisigRecord.Name,
				sign1File.Name(),
			)...)
	s.Require().NoError(err)

	// Save tx to file
	multiSigWith1SignatureFile := sdktestutil.WriteToNewTempFile(s.T(), multiSigWith1Signature.String())
	defer func() {
		_ = multiSigWith1SignatureFile.Close()
	}()

	_, err = clitestutil.TxValidateSignaturesExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				multiSigWith1SignatureFile.Name())...)
	s.Require().Error(err)
}

func (s *AuthCLITestSuite) TestCLIEncode() {
	sendTokens := sdk.NewCoin("uakt", sdk.TokensFromConsensusPower(10, cli.DefaultPowerReduction))

	normalGeneratedTx, err := s.createBankMsg(
		s.cctx,
		s.val,
		sdk.NewCoins(sendTokens),
		cli.TestFlags().
			WithGenerateOnly().
			WithNote("deadbeef").
			WithFrom(s.val.String())...)
	s.Require().NoError(err)

	savedTxFile := sdktestutil.WriteToNewTempFile(s.T(), normalGeneratedTx.String())
	defer func() {
		_ = savedTxFile.Close()
	}()

	// Encode
	encodeExec, err := clitestutil.TxEncodeExec(
		s.ctx,
		s.cctx,
		savedTxFile.Name())
	s.Require().NoError(err)

	trimmedBase64 := strings.Trim(encodeExec.String(), "\"\n")
	// Check that the transaction decodes as expected
	decodedTx, err := clitestutil.TxDecodeExec(
		s.ctx,
		s.cctx,
		trimmedBase64)
	s.Require().NoError(err)

	txCfg := s.cctx.TxConfig
	theTx, err := txCfg.TxJSONDecoder()(decodedTx.Bytes())
	s.Require().NoError(err)
	txBuilder, err := s.cctx.TxConfig.WrapTxBuilder(theTx)
	s.Require().NoError(err)
	s.Require().Equal("deadbeef", txBuilder.GetTx().GetMemo())
}

func (s *AuthCLITestSuite) TestCLIMultisignSortSignatures() {
	// Generate 2 accounts and a multisig.
	account1, err := s.cctx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	account2, err := s.cctx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	multisigRecord, err := s.cctx.Keyring.Key("multi")
	s.Require().NoError(err)

	// Generate dummy account which is not a part of multisig.
	dummyAcc, err := s.cctx.Keyring.Key("dummyAccount")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)

	// Generate multisig transaction.
	multiGeneratedTx, err := clitestutil.ExecSend(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				addr.String(),
				s.val.String(),
				sdk.NewCoins(
					sdk.NewInt64Coin("uakt", 5),
				).String(),
			).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithGenerateOnly()...)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := sdktestutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer func() {
		_ = multiGeneratedTxFile.Close()
	}()

	// Sign with account1
	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	s.cctx.HomeDir = strings.Replace(s.cctx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				multiGeneratedTxFile.Name(),
			).
			WithFrom(addr1.String()).
			WithMultisig(addr.String()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)

	sign1File := sdktestutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer func() {
		_ = sign1File.Close()
	}()

	// Sign with account2
	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	account2Signature, err := clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				multiGeneratedTxFile.Name(),
			).
			WithFrom(addr2.String()).
			WithMultisig(addr.String()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)

	sign2File := sdktestutil.WriteToNewTempFile(s.T(), account2Signature.String())
	defer func() {
		_ = sign2File.Close()
	}()

	// Sign with dummy account
	dummyAddr, err := dummyAcc.GetAddress()
	s.Require().NoError(err)
	_, err = clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				multiGeneratedTxFile.Name(),
			).
			WithFrom(dummyAddr.String()).
			WithMultisig(addr.String()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "signing key is not a part of multisig key")

	multiSigWith2Signatures, err := clitestutil.TxMultiSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				multiGeneratedTxFile.Name(),
				multisigRecord.Name,
				sign1File.Name(),
				sign2File.Name()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile := sdktestutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())
	defer func() {
		_ = signedTxFile.Close()
	}()

	_, err = clitestutil.TxValidateSignaturesExec(
		s.ctx,
		s.cctx,
		signedTxFile.Name())
	s.Require().NoError(err)

	_, err = clitestutil.TxBroadcastExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(signedTxFile.Name()).
			WithBroadcastModeSync()...)
	s.Require().NoError(err)
}

func (s *AuthCLITestSuite) TestSignWithMultisig() {
	// Generate a account for signing.
	account1, err := s.cctx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	addr1, err := account1.GetAddress()
	s.Require().NoError(err)

	// Create an address that is not in the keyring, will be used to simulate `--multisig`
	multisig := "akash1hd6fsrvnz6qkp87s3u86ludegq97agxsqdr9ad"
	multisigAddr, err := sdk.AccAddressFromBech32(multisig)
	s.Require().NoError(err)

	// Generate a transaction for testing --multisig with an address not in the keyring.
	multisigTx, err := clitestutil.ExecSend(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				s.val.String(),
				s.val.String(),
				sdk.NewCoins(
					sdk.NewInt64Coin("uakt", 5),
				).String()).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithGenerateOnly().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10))))...)
	s.Require().NoError(err)

	// Save multi tx to file
	multiGeneratedTx2File := sdktestutil.WriteToNewTempFile(s.T(), multisigTx.String())
	defer func() {
		_ = multiGeneratedTx2File.Close()
	}()

	// Sign using multisig. We're signing a tx on behalf of the multisig address,
	// even though the tx signer is NOT the multisig address. This is fine though,
	// as the main point of this test is to test the `--multisig` flag with an address
	// that is not in the keyring.
	_, err = clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(multiGeneratedTx2File.Name()).
			WithFrom(addr1.String()).
			WithMultisig(multisigAddr.String()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)

	s.Require().Contains(err.Error(), "error getting account from keybase")
}

func (s *AuthCLITestSuite) TestCLIMultisign() {
	// Generate 2 accounts and a multisig.
	account1, err := s.cctx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	account2, err := s.cctx.Keyring.Key("newAccount2")
	s.Require().NoError(err)

	multisigRecord, err := s.cctx.Keyring.Key("multi")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)

	// Generate multisig transaction.
	multiGeneratedTx, err := clitestutil.ExecSend(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				addr.String(),
				s.val.String(),
				sdk.NewCoins(
					sdk.NewInt64Coin("uakt", 5),
				).String()).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithGenerateOnly().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10))))...,
	)
	s.Require().NoError(err)

	// Save tx to file
	multiGeneratedTxFile := sdktestutil.WriteToNewTempFile(s.T(), multiGeneratedTx.String())
	defer func() {
		_ = multiGeneratedTxFile.Close()
	}()

	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	// Sign with account1
	s.cctx.HomeDir = strings.Replace(s.cctx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(multiGeneratedTxFile.Name()).
			WithFrom(addr1.String()).
			WithMultisig(addr.String()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)

	sign1File := sdktestutil.WriteToNewTempFile(s.T(), account1Signature.String())
	defer func() {
		_ = sign1File.Close()
	}()

	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	// Sign with account2
	account2Signature, err := clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				multiGeneratedTxFile.Name()).
			WithFrom(addr2.String()).
			WithMultisig(addr.String()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)

	sign2File := sdktestutil.WriteToNewTempFile(s.T(), account2Signature.String())
	defer func() {
		_ = sign2File.Close()
	}()

	s.cctx.Offline = false
	multiSigWith2Signatures, err := clitestutil.TxMultiSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				multiGeneratedTxFile.Name(),
				multisigRecord.Name,
				sign1File.Name(),
				sign2File.Name()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)

	// Write the output to disk
	signedTxFile := sdktestutil.WriteToNewTempFile(s.T(), multiSigWith2Signatures.String())
	defer func() {
		_ = signedTxFile.Close()
	}()

	_, err = clitestutil.TxValidateSignaturesExec(s.ctx, s.cctx, signedTxFile.Name())
	s.Require().NoError(err)

	_, err = clitestutil.TxBroadcastExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(signedTxFile.Name()).
			WithBroadcastModeSync()...)
	s.Require().NoError(err)
}

func (s *AuthCLITestSuite) TestSignBatchMultisig() {
	// Fetch 2 accounts and a multisig.
	account1, err := s.cctx.Keyring.Key("newAccount1")
	s.Require().NoError(err)
	account2, err := s.cctx.Keyring.Key("newAccount2")
	s.Require().NoError(err)
	multisigRecord, err := s.cctx.Keyring.Key("multi")
	s.Require().NoError(err)

	addr, err := multisigRecord.GetAddress()
	s.Require().NoError(err)
	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin("uakt", 10)
	_, err = s.createBankMsg(
		s.cctx,
		addr,
		sdk.NewCoins(sendTokens),
	)
	s.Require().NoError(err)

	generatedStd, err := clitestutil.ExecSend(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				addr.String(),
				s.val.String(),
				sdk.NewCoins(
					sdk.NewCoin("uakt", sdkmath.NewInt(1)),
				).String(),
			).
			WithBroadcastModeSync().
			WithSkipConfirm().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			WithGenerateOnly()...,
	)
	s.Require().NoError(err)

	// Write the output to disk
	filename := sdktestutil.WriteToNewTempFile(s.T(), strings.Repeat(generatedStd.String(), 1))
	defer func() {
		_ = filename.Close()
	}()

	s.cctx.HomeDir = strings.Replace(s.cctx.HomeDir, "simd", "simcli", 1)

	addr1, err := account1.GetAddress()
	s.Require().NoError(err)
	// sign-batch file
	res, err := clitestutil.TxSignBatchExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				filename.Name(),
			).
			WithFrom(addr1.String()).
			WithChainID(s.cctx.ChainID).
			WithSignatureOnly().
			WithMultisig(addr.String()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)
	s.Require().Equal(1, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file
	file1 := sdktestutil.WriteToNewTempFile(s.T(), res.String())
	defer func() {
		_ = file1.Close()
	}()

	addr2, err := account2.GetAddress()
	s.Require().NoError(err)
	// sign-batch file with account2
	res, err = clitestutil.TxSignBatchExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				filename.Name(),
			).
			WithFrom(addr2.String()).
			WithChainID(s.cctx.ChainID).
			WithSignatureOnly().
			WithMultisig(addr.String()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)
	s.Require().Equal(1, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file2
	file2 := sdktestutil.WriteToNewTempFile(s.T(), res.String())
	defer func() {
		_ = file2.Close()
	}()
	_, err = clitestutil.TxMultiSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(
				filename.Name(),
				multisigRecord.Name,
				file1.Name(),
				file2.Name()).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)
}

func (s *AuthCLITestSuite) TestGetBroadcastCommandOfflineFlag() {
	_, err := clitestutil.TxBroadcastExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With("fsdf").
			WithOffline()...,
	)
	s.Require().EqualError(err, "cannot broadcast tx during offline mode")
}

func (s *AuthCLITestSuite) TestQueryParamsCmd() {
	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"happy case",
			cli.TestFlags().
				WithOutputJSON(),
			false,
		},
		{
			"with specific height",
			cli.TestFlags().
				WithHeight(1).
				WithOutputJSON(),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetQueryAuthParamsCmd()
			cctx := s.cctx

			out, err := clitestutil.ExecTestCLICmd(s.ctx, cctx, cmd, tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().NotEqual("internal", err.Error())
			} else {
				var authParams authtypes.Params
				s.Require().NoError(err)
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &authParams))
				s.Require().NotNil(authParams.MaxMemoCharacters)
			}
		})
	}
}

// TestTxWithoutPublicKey makes sure sending a proto tx message without the
// public key doesn't cause any error in the RPC layer (broadcast).
// See https://github.com/cosmos/cosmos-sdk/issues/7585 for more details.
func (s *AuthCLITestSuite) TestTxWithoutPublicKey() {
	txCfg := s.cctx.TxConfig

	// Create a txBuilder with an unsigned tx.
	txBuilder := txCfg.NewTxBuilder()
	msg := banktypes.NewMsgSend(
		s.val,
		s.val,
		sdk.NewCoins(
			sdk.NewCoin("uakt", sdkmath.NewInt(10)),
		))
	err := txBuilder.SetMsgs(msg)
	s.Require().NoError(err)

	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(150))))
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())

	// Create a file with the unsigned tx.
	txJSON, err := txCfg.TxJSONEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)

	unsignedTxFile := sdktestutil.WriteToNewTempFile(s.T(), string(txJSON))
	defer func() {
		_ = unsignedTxFile.Close()
	}()

	// Sign the file with the unsignedTx.
	signedTx, err := clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(unsignedTxFile.Name()).
			WithFrom(s.val.String()).
			WithOverwrite()...)
	s.Require().NoError(err)

	// Remove the signerInfo's `public_key` field manually from the signedTx.
	// Note: this method is only used for test purposes! In general, one should
	// use txBuilder and TxEncoder/TxDecoder to manipulate txs.
	var tx tx.Tx
	err = s.cctx.Codec.UnmarshalJSON(signedTx.Bytes(), &tx)
	s.Require().NoError(err)

	tx.AuthInfo.SignerInfos[0].PublicKey = nil
	// Re-encode the tx again, to another file.
	txJSON, err = s.cctx.Codec.MarshalJSON(&tx)
	s.Require().NoError(err)

	signedTxFile := sdktestutil.WriteToNewTempFile(s.T(), string(txJSON))
	defer func() {
		_ = signedTxFile.Close()
	}()
	s.Require().True(strings.Contains(string(txJSON), "\"public_key\":null"))

	// Broadcast tx, test that it shouldn't panic.
	out, err := clitestutil.TxBroadcastExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(signedTxFile.Name()).
			WithBroadcastModeSync()...)
	s.Require().NoError(err)

	var res sdk.TxResponse
	s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &res))
	s.Require().NotEqual(0, res.Code)
}

// TestSignWithMultiSignersAminoJSON tests the case where a transaction with 2
// messages which has to be signed with 2 different keys. Sign and append the
// signatures using the CLI with Amino signing mode. Finally, send the
// transaction to the blockchain.
func (s *AuthCLITestSuite) TestSignWithMultiSignersAminoJSON() {
	val0, val1 := s.val, s.val1
	val0Coin := sdk.NewCoin("test1token", sdkmath.NewInt(10))
	val1Coin := sdk.NewCoin("test2token", sdkmath.NewInt(10))

	from, err := s.cctx.Keyring.Key("newAccount1")
	s.Require().NoError(err)

	addr1, err := from.GetAddress()
	s.Require().NoError(err)

	// Creating a tx with 2 msgs from 2 signers: val0 and val1.
	// The validators need to sign with SIGN_MODE_LEGACY_AMINO_JSON,
	// because DIRECT doesn't support multi signers via the CLI.
	// Since we use amino, we don't need to pre-populate signer_infos.
	txBuilder := s.cctx.TxConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(
		banktypes.NewMsgSend(val0, addr1, sdk.NewCoins(val0Coin)),
		banktypes.NewMsgSend(val1, addr1, sdk.NewCoins(val1Coin)),
	)
	s.Require().NoError(err)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10))))
	txBuilder.SetGasLimit(testdata.NewTestGasLimit() * 2)
	signers, err := txBuilder.GetTx().GetSigners()
	s.Require().NoError(err)
	s.Require().Equal([][]byte{val0, val1}, signers)

	// Write the unsigned tx into a file.
	txJSON, err := s.cctx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	s.Require().NoError(err)
	unsignedTxFile := sdktestutil.WriteToNewTempFile(s.T(), string(txJSON))
	defer func() {
		_ = unsignedTxFile.Close()
	}()

	// Let val0 sign first the file with the unsignedTx.
	signedByVal0, err := clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(unsignedTxFile.Name()).
			WithFrom(val0.String()).
			WithOverwrite().
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)
	signedByVal0File := sdktestutil.WriteToNewTempFile(s.T(), signedByVal0.String())
	defer func() {
		_ = signedByVal0File.Close()
	}()

	// Then let val1 sign the file with signedByVal0.
	val1AccNum, val1Seq, err := s.cctx.AccountRetriever.GetAccountNumberSequence(s.cctx, val1)
	s.Require().NoError(err)

	signedTx, err := clitestutil.TxSignExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(signedByVal0File.Name()).
			WithFrom(val1.String()).
			WithOffline().
			WithAccountNumber(val1AccNum).
			WithSequence(val1Seq).
			WithSignMode(cflags.SignModeLegacyAminoJSON)...)
	s.Require().NoError(err)
	signedTxFile := sdktestutil.WriteToNewTempFile(s.T(), signedTx.String())
	defer func() {
		_ = signedTxFile.Close()
	}()

	res, err := clitestutil.TxBroadcastExec(
		s.ctx,
		s.cctx,
		cli.TestFlags().
			With(signedTxFile.Name()).
			WithBroadcastModeSync()...)
	s.Require().NoError(err)

	var txRes sdk.TxResponse
	s.Require().NoError(s.cctx.Codec.UnmarshalJSON(res.Bytes(), &txRes))
	s.Require().Equal(uint32(0), txRes.Code, txRes.RawLog)
}

func (s *AuthCLITestSuite) TestAuxSigner() {
	val0Coin := sdk.NewCoin("testtoken", sdkmath.NewInt(10))

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"error with SIGN_MODE_DIRECT_AUX and --aux unset",
			cli.TestFlags().
				WithSignMode(cflags.SignModeDirectAux),
			true,
		},
		{
			"no error with SIGN_MODE_DIRECT_AUX mode and generate-only set (ignores generate-only)",
			cli.TestFlags().
				WithSignMode(cflags.SignModeDirectAux).
				WithGenerateOnly(),
			false,
		},
		{
			"no error with SIGN_MODE_DIRECT_AUX mode and generate-only, tip flag set",
			cli.TestFlags().
				WithSignMode(cflags.SignModeDirectAux).
				WithGenerateOnly().
				WithTip(val0Coin),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxGovSubmitLegacyProposalCmd()
			_, err := clitestutil.ExecTestCLICmd(
				s.ctx,
				s.cctx,
				cmd,
				cli.TestFlags().
					WithTitle("Text Proposal").
					WithDescription("test desc").
					WithProposalType(govtypes.ProposalTypeText).
					WithFrom(s.val.String()).
					Append(tc.args)...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *AuthCLITestSuite) TestAuxToFeeWithTips() {
	// Skipping this test as it needs a simapp with the TipDecorator in post handler.
	s.T().Skip()

	require := s.Require()

	kb := s.cctx.Keyring
	acc, _, err := kb.NewMnemonic("tipperAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(err)

	tipper, err := acc.GetAddress()
	require.NoError(err)
	tipperInitialBal := sdk.NewCoin("testtoken", sdkmath.NewInt(10000))

	feePayer := s.val
	fee := sdk.NewCoin("uakt", sdkmath.NewInt(1000))
	tip := sdk.NewCoin("testtoken", sdkmath.NewInt(1000))

	_, err = s.createBankMsg(s.cctx, tipper, sdk.NewCoins(tipperInitialBal))
	require.NoError(err)

	bal := s.getBalances(s.cctx, tipper, tip.Denom)
	require.True(bal.Equal(tipperInitialBal.Amount))

	testCases := []struct {
		name               string
		tipper             sdk.AccAddress
		feePayer           sdk.AccAddress
		tip                sdk.Coin
		expectErrAux       bool
		expectErrBroadCast bool
		errMsg             string
		tipperArgs         []string
		feePayerArgs       []string
	}{
		{
			name:     "when --aux and --sign-mode = direct set: error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeDirect).
				WithTip(tip).
				WithAux(),
			expectErrAux: true,
			feePayerArgs: cli.TestFlags().
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFrom(feePayer.String()).
				WithFees(sdk.Coins{fee}),
		},
		{
			name:     "both tipper, fee payer uses AMINO: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeLegacyAminoJSON).
				WithTip(tip).
				WithAux(),
			feePayerArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeLegacyAminoJSON).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFrom(feePayer.String()).
				WithFees(sdk.Coins{fee}),
		},
		{
			name:     "tipper uses DIRECT_AUX, fee payer uses AMINO: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeDirectAux).
				WithTip(tip).
				WithAux(),
			feePayerArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeLegacyAminoJSON).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFrom(feePayer.String()).
				WithFees(sdk.Coins{fee}),
		},
		{
			name:     "--tip flag unset: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      sdk.Coin{Denom: "testtoken", Amount: sdkmath.NewInt(0)},
			tipperArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeDirectAux).
				WithAux(),
			feePayerArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeLegacyAminoJSON).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFrom(feePayer.String()).
				WithFees(sdk.Coins{fee}),
		},
		{
			name:     "legacy amino json: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeLegacyAminoJSON).
				WithTip(tip).
				WithAux(),
			feePayerArgs: cli.TestFlags().
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFrom(feePayer.String()).
				WithFees(sdk.Coins{fee}),
		},
		{
			name:     "tipper uses direct aux, fee payer uses direct: happy case",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeDirectAux).
				WithTip(tip).
				WithAux(),
			feePayerArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeDirect).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFrom(feePayer.String()).
				WithFees(sdk.Coins{fee}),
		},
		{
			name:     "chain-id mismatch: error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeDirectAux).
				WithTip(tip).
				WithAux(),
			expectErrAux: false,
			feePayerArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeDirect).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFrom(feePayer.String()).
				WithFees(sdk.Coins{fee}).
				WithChainID("foobar"),
			expectErrBroadCast: true,
		},
		{
			name:     "wrong denom in tip: error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      sdk.Coin{Denom: "testtoken", Amount: sdkmath.NewInt(0)},
			tipperArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeDirectAux).
				WithTip(sdk.Coin{Denom: "wrongDenom", Amount: sdkmath.NewInt(100)}).
				WithAux(),
			feePayerArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeDirect).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFrom(feePayer.String()).
				WithFees(sdk.Coins{fee}),
			errMsg: "insufficient funds",
		},
		{
			name:     "insufficient fees: error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      sdk.Coin{Denom: "testtoken", Amount: sdkmath.NewInt(0)},
			tipperArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeDirectAux).
				WithTip(tip).
				WithAux(),
			feePayerArgs: cli.TestFlags().
				WithSignMode(cflags.SignModeDirect).
				WithSkipConfirm().
				WithBroadcastModeSync().
				WithFrom(feePayer.String()),
			errMsg: "insufficient fees",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			cmd := cli.GetTxGovSubmitLegacyProposalCmd()
			res, err := clitestutil.ExecTestCLICmd(
				s.ctx,
				s.cctx,
				cmd,
				cli.TestFlags().
					WithTitle("test").
					WithDescription("test desc").
					WithProposalType(govtypes.ProposalTypeText).
					WithFrom(s.val.String()).
					Append(tc.tipperArgs)...)

			if tc.expectErrAux {
				require.Error(err)
			} else {
				require.NoError(err)
				genTxFile := sdktestutil.WriteToNewTempFile(s.T(), string(res.Bytes()))
				defer func() {
					_ = genTxFile.Close()
				}()

				// broadcast the tx
				res, err = clitestutil.TxAuxToFeeExec(
					s.ctx,
					s.cctx,
					genTxFile.Name(),
					tc.feePayerArgs...,
				)

				switch {
				case tc.expectErrBroadCast:
					require.Error(err)

				case tc.errMsg != "":
					require.NoError(err)

					var txRes sdk.TxResponse
					require.NoError(s.cctx.Codec.UnmarshalJSON(res.Bytes(), &txRes))

					require.Contains(txRes.RawLog, tc.errMsg)

				default:
					require.NoError(err)

					var txRes sdk.TxResponse
					require.NoError(s.cctx.Codec.UnmarshalJSON(res.Bytes(), &txRes))

					require.Equal(uint32(0), txRes.Code)
					require.NotNil(int64(0), txRes.Height)

					bal = s.getBalances(s.cctx, tipper, tc.tip.Denom)
					tipperInitialBal = tipperInitialBal.Sub(tc.tip)
					require.True(bal.Equal(tipperInitialBal.Amount))
				}
			}
		})
	}
}

func (s *AuthCLITestSuite) getBalances(cctx client.Context, addr sdk.AccAddress, denom string) sdkmath.Int {
	resp, err := clitestutil.QueryBalancesExec(
		s.ctx,
		cctx,
		cli.TestFlags().
			With(addr.String())...)
	s.Require().NoError(err)

	var balRes banktypes.QueryAllBalancesResponse
	err = cctx.Codec.UnmarshalJSON(resp.Bytes(), &balRes)
	s.Require().NoError(err)
	startTokens := balRes.Balances.AmountOf(denom)
	return startTokens
}

func (s *AuthCLITestSuite) createBankMsg(cctx client.Context, toAddr sdk.AccAddress, amount sdk.Coins, extraFlags ...string) (sdktestutil.BufferWriter, error) {
	return clitestutil.ExecSend(
		s.ctx,
		cctx,
		cli.TestFlags().
			With(
				s.val.String(),
				toAddr.String(),
				amount.String()).
			WithSkipConfirm().
			WithBroadcastModeSync().
			WithFees(sdk.NewCoins(sdk.NewCoin("uakt", sdkmath.NewInt(10)))).
			Append(extraFlags)...)
}
