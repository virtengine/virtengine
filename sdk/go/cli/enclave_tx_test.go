package cli_test

import (
	"encoding/hex"

	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/sdk/go/cli"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
)

func (s *EnclaveCLITestSuite) TestTxEnclaveRegisterCmd() {
	accounts := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	measurementHash := hex.EncodeToString(make([]byte, 32))
	encryptionPubKey := hex.EncodeToString([]byte{0x01, 0x02, 0x03, 0x04})
	signingPubKey := hex.EncodeToString([]byte{0x0A, 0x0B, 0x0C})
	attestationQuote := hex.EncodeToString([]byte("quote"))

	commonArgs := cli.TestFlags().
		WithFrom(accounts[0].Address.String()).
		WithSkipConfirm().
		WithBroadcastModeSync().
		WithFees(sdk.NewCoins(sdk.NewInt64Coin("uve", 10))).
		WithChainID("test-chain")

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"missing required flags",
			commonArgs,
			true,
		},
		{
			"invalid tee type",
			cli.TestFlags().
				WithFlag("tee-type", "INVALID").
				WithFlag("measurement-hash", measurementHash).
				WithFlag("encryption-pubkey", encryptionPubKey).
				WithFlag("signing-pubkey", signingPubKey).
				WithFlag("attestation-quote", attestationQuote).
				Append(commonArgs),
			true,
		},
		{
			"invalid measurement hash",
			cli.TestFlags().
				WithFlag("tee-type", "SGX").
				WithFlag("measurement-hash", "not-hex").
				WithFlag("encryption-pubkey", encryptionPubKey).
				WithFlag("signing-pubkey", signingPubKey).
				WithFlag("attestation-quote", attestationQuote).
				Append(commonArgs),
			true,
		},
		{
			"invalid encryption pubkey",
			cli.TestFlags().
				WithFlag("tee-type", "SGX").
				WithFlag("measurement-hash", measurementHash).
				WithFlag("encryption-pubkey", "xyz").
				WithFlag("signing-pubkey", signingPubKey).
				WithFlag("attestation-quote", attestationQuote).
				Append(commonArgs),
			true,
		},
		{
			"invalid signing pubkey",
			cli.TestFlags().
				WithFlag("tee-type", "SGX").
				WithFlag("measurement-hash", measurementHash).
				WithFlag("encryption-pubkey", encryptionPubKey).
				WithFlag("signing-pubkey", "xyz").
				WithFlag("attestation-quote", attestationQuote).
				Append(commonArgs),
			true,
		},
		{
			"invalid attestation quote",
			cli.TestFlags().
				WithFlag("tee-type", "SGX").
				WithFlag("measurement-hash", measurementHash).
				WithFlag("encryption-pubkey", encryptionPubKey).
				WithFlag("signing-pubkey", signingPubKey).
				WithFlag("attestation-quote", "not-hex").
				Append(commonArgs),
			true,
		},
		{
			"valid register",
			cli.TestFlags().
				WithFlag("tee-type", "SGX").
				WithFlag("measurement-hash", measurementHash).
				WithFlag("encryption-pubkey", encryptionPubKey).
				WithFlag("signing-pubkey", signingPubKey).
				WithFlag("attestation-quote", attestationQuote).
				Append(commonArgs),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cli.GetTxEnclaveRegisterCmd(), tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var response sdk.TxResponse
				s.Require().NoError(err, out.String())
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
			}
		})
	}
}

func (s *EnclaveCLITestSuite) TestTxEnclaveRotateCmd() {
	accounts := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	encryptionPubKey := hex.EncodeToString([]byte{0x01, 0x02, 0x03, 0x04})
	signingPubKey := hex.EncodeToString([]byte{0x0A, 0x0B, 0x0C})
	attestationQuote := hex.EncodeToString([]byte("quote"))
	newMeasurement := hex.EncodeToString(make([]byte, 32))

	commonArgs := cli.TestFlags().
		WithFrom(accounts[0].Address.String()).
		WithSkipConfirm().
		WithBroadcastModeSync().
		WithFees(sdk.NewCoins(sdk.NewInt64Coin("uve", 10))).
		WithChainID("test-chain")

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"missing required flags",
			commonArgs,
			true,
		},
		{
			"invalid encryption pubkey",
			cli.TestFlags().
				WithFlag("encryption-pubkey", "xyz").
				WithFlag("signing-pubkey", signingPubKey).
				WithFlag("attestation-quote", attestationQuote).
				WithFlag("overlap-blocks", 100).
				Append(commonArgs),
			true,
		},
		{
			"invalid signing pubkey",
			cli.TestFlags().
				WithFlag("encryption-pubkey", encryptionPubKey).
				WithFlag("signing-pubkey", "xyz").
				WithFlag("attestation-quote", attestationQuote).
				WithFlag("overlap-blocks", 100).
				Append(commonArgs),
			true,
		},
		{
			"invalid attestation quote",
			cli.TestFlags().
				WithFlag("encryption-pubkey", encryptionPubKey).
				WithFlag("signing-pubkey", signingPubKey).
				WithFlag("attestation-quote", "not-hex").
				WithFlag("overlap-blocks", 100).
				Append(commonArgs),
			true,
		},
		{
			"overlap blocks must be positive",
			cli.TestFlags().
				WithFlag("encryption-pubkey", encryptionPubKey).
				WithFlag("signing-pubkey", signingPubKey).
				WithFlag("attestation-quote", attestationQuote).
				WithFlag("overlap-blocks", 0).
				Append(commonArgs),
			true,
		},
		{
			"invalid new measurement hash length",
			cli.TestFlags().
				WithFlag("encryption-pubkey", encryptionPubKey).
				WithFlag("signing-pubkey", signingPubKey).
				WithFlag("attestation-quote", attestationQuote).
				WithFlag("overlap-blocks", 100).
				WithFlag("new-measurement-hash", "abcd").
				Append(commonArgs),
			true,
		},
		{
			"valid rotate",
			cli.TestFlags().
				WithFlag("encryption-pubkey", encryptionPubKey).
				WithFlag("signing-pubkey", signingPubKey).
				WithFlag("attestation-quote", attestationQuote).
				WithFlag("overlap-blocks", 100).
				WithFlag("new-measurement-hash", newMeasurement).
				Append(commonArgs),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cli.GetTxEnclaveRotateCmd(), tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var response sdk.TxResponse
				s.Require().NoError(err, out.String())
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
			}
		})
	}
}

func (s *EnclaveCLITestSuite) TestTxEnclaveProposeMeasurementCmd() {
	accounts := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	measurementHash := hex.EncodeToString(make([]byte, 32))

	commonArgs := cli.TestFlags().
		WithFrom(accounts[0].Address.String()).
		WithSkipConfirm().
		WithBroadcastModeSync().
		WithFees(sdk.NewCoins(sdk.NewInt64Coin("uve", 10))).
		WithChainID("test-chain")

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"missing required flags",
			commonArgs,
			true,
		},
		{
			"invalid measurement hash",
			cli.TestFlags().
				WithFlag("measurement-hash", "not-hex").
				WithFlag("tee-type", "SGX").
				WithFlag("description", "test").
				Append(commonArgs),
			true,
		},
		{
			"invalid tee type",
			cli.TestFlags().
				WithFlag("measurement-hash", measurementHash).
				WithFlag("tee-type", "INVALID").
				WithFlag("description", "test").
				Append(commonArgs),
			true,
		},
		{
			"missing description",
			cli.TestFlags().
				WithFlag("measurement-hash", measurementHash).
				WithFlag("tee-type", "SGX").
				Append(commonArgs),
			true,
		},
		{
			"valid propose measurement",
			cli.TestFlags().
				WithFlag("measurement-hash", measurementHash).
				WithFlag("tee-type", "SGX").
				WithFlag("description", "test measurement").
				Append(commonArgs),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cli.GetTxEnclaveProposeMeasurementCmd(), tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var response sdk.TxResponse
				s.Require().NoError(err, out.String())
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
			}
		})
	}
}

func (s *EnclaveCLITestSuite) TestTxEnclaveRevokeMeasurementCmd() {
	accounts := sdktestutil.CreateKeyringAccounts(s.T(), s.kr, 1)

	measurementHash := hex.EncodeToString(make([]byte, 32))

	commonArgs := cli.TestFlags().
		WithFrom(accounts[0].Address.String()).
		WithSkipConfirm().
		WithBroadcastModeSync().
		WithFees(sdk.NewCoins(sdk.NewInt64Coin("uve", 10))).
		WithChainID("test-chain")

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"missing required flags",
			commonArgs,
			true,
		},
		{
			"invalid measurement hash",
			cli.TestFlags().
				WithFlag("measurement-hash", "not-hex").
				WithFlag("reason", "bad measurement").
				Append(commonArgs),
			true,
		},
		{
			"missing reason",
			cli.TestFlags().
				WithFlag("measurement-hash", measurementHash).
				Append(commonArgs),
			true,
		},
		{
			"valid revoke measurement",
			cli.TestFlags().
				WithFlag("measurement-hash", measurementHash).
				WithFlag("reason", "security issue").
				Append(commonArgs),
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			out, err := clitestutil.ExecTestCLICmd(s.ctx, s.cctx, cli.GetTxEnclaveRevokeMeasurementCmd(), tc.args...)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var response sdk.TxResponse
				s.Require().NoError(err, out.String())
				s.Require().NoError(s.cctx.Codec.UnmarshalJSON(out.Bytes(), &response), out.String())
			}
		})
	}
}
