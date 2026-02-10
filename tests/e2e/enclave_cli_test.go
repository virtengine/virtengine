//go:build e2e.integration

package e2e

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/virtengine/virtengine/sdk/go/cli"
	clitestutil "github.com/virtengine/virtengine/sdk/go/cli/testutil"
	enclavetypes "github.com/virtengine/virtengine/sdk/go/node/enclave/v1"
	"github.com/virtengine/virtengine/testutil"
)

type enclaveFixtures struct {
	MeasurementHash  string `json:"measurement_hash"`
	EncryptionPubKey string `json:"encryption_pubkey"`
	SigningPubKey    string `json:"signing_pubkey"`
	AttestationQuote string `json:"attestation_quote"`
	TEEType          string `json:"tee_type"`
	ValidatorAddress string `json:"validator_address"`
	ScopeID          string `json:"scope_id"`
	BlockHeight      int64  `json:"block_height"`
}

type enclaveIntegrationTestSuite struct {
	*testutil.NetworkTestSuite
	cctx client.Context
	fx   enclaveFixtures
}

func (s *enclaveIntegrationTestSuite) SetupSuite() {
	s.NetworkTestSuite.SetupSuite()
	val := s.Network().Validators[0]
	s.cctx = val.ClientCtx

	path, err := filepath.Abs("testdata/enclave_fixtures.json")
	s.Require().NoError(err)

	data, err := os.ReadFile(path)
	s.Require().NoError(err)

	s.Require().NoError(json.Unmarshal(data, &s.fx))
}

// Naming as Test{number} just to run all tests sequentially
func (s *enclaveIntegrationTestSuite) Test1QueryParams() {
	ctx := context.Background()
	cctx := s.cctx

	resp, err := clitestutil.ExecTestCLICmd(
		ctx,
		cctx,
		cli.GetQueryEnclaveParamsCmd(),
		cli.TestFlags().WithOutputJSON()...,
	)
	s.Require().NoError(err)

	out := &enclavetypes.QueryParamsResponse{}
	s.Require().NoError(cctx.Codec.UnmarshalJSON(resp.Bytes(), out))
	s.Require().NotEmpty(out.Params.AllowedTeeTypes)
}

func (s *enclaveIntegrationTestSuite) Test2QueryMeasurements() {
	ctx := context.Background()
	cctx := s.cctx

	resp, err := clitestutil.ExecTestCLICmd(
		ctx,
		cctx,
		cli.GetQueryEnclaveMeasurementsCmd(),
		cli.TestFlags().WithOutputJSON()...,
	)
	s.Require().NoError(err)

	out := &enclavetypes.QueryMeasurementAllowlistResponse{}
	s.Require().NoError(cctx.Codec.UnmarshalJSON(resp.Bytes(), out))
}

func (s *enclaveIntegrationTestSuite) Test3QueryKeys() {
	ctx := context.Background()
	cctx := s.cctx

	resp, err := clitestutil.ExecTestCLICmd(
		ctx,
		cctx,
		cli.GetQueryEnclaveKeysCmd(),
		cli.TestFlags().WithOutputJSON()...,
	)
	s.Require().NoError(err)

	out := &enclavetypes.QueryActiveValidatorEnclaveKeysResponse{}
	s.Require().NoError(cctx.Codec.UnmarshalJSON(resp.Bytes(), out))
}

func (s *enclaveIntegrationTestSuite) Test4RegisterEnclaveInvalidAttestation() {
	ctx := context.Background()
	cctx := s.cctx

	args := cli.TestFlags().
		WithFlag("tee-type", s.fx.TEEType).
		WithFlag("measurement-hash", s.fx.MeasurementHash).
		WithFlag("encryption-pubkey", s.fx.EncryptionPubKey).
		WithFlag("signing-pubkey", s.fx.SigningPubKey).
		WithFlag("attestation-quote", s.fx.AttestationQuote).
		WithFrom(s.Network().Validators[0].Address.String()).
		WithSkipConfirm().
		WithGasAutoFlags().
		WithBroadcastModeBlock()

	_, err := clitestutil.ExecTestCLICmd(ctx, cctx, cli.GetTxEnclaveRegisterCmd(), args...)
	s.Require().Error(err)
}
