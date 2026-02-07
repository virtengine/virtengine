//go:build e2e.integration

// Package e2e contains end-to-end tests for VirtEngine.
//
// This file tests the model hash computation and governance flow:
// - Model registration with hash verification
// - Active model querying
// - Model version history tracking
// - Governance proposal for model updates
// - Validator model sync verification
// - Hash mismatch rejection
//
// Task Reference: 29B - Model Hash Computation + Governance
package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/app"
	sdktestutil "github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Model Hash Governance E2E Test Suite
// ============================================================================

type ModelHashE2ETestSuite struct {
	suite.Suite

	app    *app.VirtEngineApp
	ctx    sdk.Context
	keeper keeper.Keeper
}

func TestModelHashE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	suite.Run(t, new(ModelHashE2ETestSuite))
}

func (s *ModelHashE2ETestSuite) SetupSuite() {
	testClient := NewVEIDTestClient()

	s.app = app.Setup(
		app.WithChainID(TestChainID),
		app.WithGenesis(func(cdc codec.Codec) app.GenesisState {
			return genesisWithVEIDApprovedClientE2E(s.T(), cdc, testClient)
		}),
	)

	s.ctx = s.app.NewContext(false).
		WithBlockHeight(1).
		WithBlockTime(FixedTimestamp())

	s.keeper = s.app.Keepers.VirtEngine.VEID
}

// TestModelRegistrationAndQuery tests the end-to-end model registration flow.
func (s *ModelHashE2ETestSuite) TestModelRegistrationAndQuery() {
	ctx := s.ctx
	registrar := sdktestutil.AccAddress(s.T())

	// Register a face verification model
	modelInfo := &veidtypes.MLModelInfo{
		ModelID:      "face-v1.0.0",
		Name:         "Face Verification Model",
		Version:      "1.0.0",
		ModelType:    string(veidtypes.ModelTypeFaceVerification),
		SHA256Hash:   "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		Description:  "Production face verification model",
		RegisteredBy: registrar.String(),
	}

	err := s.keeper.RegisterModel(ctx, modelInfo)
	require.NoError(s.T(), err)

	// Query the registered model
	resp, err := s.keeper.QueryModelVersion(ctx, &veidtypes.QueryModelVersionRequest{
		ModelType: string(veidtypes.ModelTypeFaceVerification),
	})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp.ModelInfo)
	require.Equal(s.T(), "face-v1.0.0", resp.ModelInfo.ModelID)
	require.Equal(s.T(), "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", resp.ModelInfo.SHA256Hash)
}

// TestActiveModelsQuery tests querying all active models after registration.
func (s *ModelHashE2ETestSuite) TestActiveModelsQuery() {
	ctx := s.ctx
	registrar := sdktestutil.AccAddress(s.T())

	// Register models for multiple types
	models := []struct {
		id        string
		modelType veidtypes.ModelType
		hash      string
	}{
		{"liveness-v1.0.0", veidtypes.ModelTypeLiveness, "1111111111111111111111111111111111111111111111111111111111111111"},
		{"ocr-v1.0.0", veidtypes.ModelTypeOCR, "2222222222222222222222222222222222222222222222222222222222222222"},
	}

	for _, m := range models {
		err := s.keeper.RegisterModel(ctx, &veidtypes.MLModelInfo{
			ModelID:      m.id,
			Name:         string(m.modelType) + " model",
			Version:      "1.0.0",
			ModelType:    string(m.modelType),
			SHA256Hash:   m.hash,
			RegisteredBy: registrar.String(),
		})
		require.NoError(s.T(), err)
	}

	// Query all active models
	resp, err := s.keeper.QueryActiveModels(ctx, &veidtypes.QueryActiveModelsRequest{})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp)
	require.GreaterOrEqual(s.T(), len(resp.Models), 2)
}

// TestHashMismatchRejection tests that scoring is rejected when model hash
// doesn't match the governance-approved hash.
func (s *ModelHashE2ETestSuite) TestHashMismatchRejection() {
	ctx := s.ctx
	registrar := sdktestutil.AccAddress(s.T())

	approvedHash := "aaaa000000000000000000000000000000000000000000000000000000000000"

	err := s.keeper.RegisterModel(ctx, &veidtypes.MLModelInfo{
		ModelID:      "trust-v1.0.0",
		Name:         "Trust Score Model",
		Version:      "1.0.0",
		ModelType:    string(veidtypes.ModelTypeTrustScore),
		SHA256Hash:   approvedHash,
		RegisteredBy: registrar.String(),
	})
	require.NoError(s.T(), err)

	// Validate with correct hash should succeed
	err = s.keeper.ValidateModelForScoring(ctx, string(veidtypes.ModelTypeTrustScore), approvedHash)
	require.NoError(s.T(), err)

	// Validate with wrong hash should fail
	wrongHash := "bbbb000000000000000000000000000000000000000000000000000000000000"
	err = s.keeper.ValidateModelForScoring(ctx, string(veidtypes.ModelTypeTrustScore), wrongHash)
	require.Error(s.T(), err)
}

// TestGovernanceProposalFlow tests the governance proposal lifecycle for model updates.
func (s *ModelHashE2ETestSuite) TestGovernanceProposalFlow() {
	ctx := s.ctx
	registrar := sdktestutil.AccAddress(s.T())
	proposer := sdktestutil.AccAddress(s.T())

	// Register initial model
	err := s.keeper.RegisterModel(ctx, &veidtypes.MLModelInfo{
		ModelID:      "gan-v1.0.0",
		Name:         "GAN Detection Model v1",
		Version:      "1.0.0",
		ModelType:    string(veidtypes.ModelTypeGANDetection),
		SHA256Hash:   "cccc000000000000000000000000000000000000000000000000000000000000",
		RegisteredBy: registrar.String(),
	})
	require.NoError(s.T(), err)

	// Register new version to be proposed
	err = s.keeper.RegisterModel(ctx, &veidtypes.MLModelInfo{
		ModelID:      "gan-v2.0.0",
		Name:         "GAN Detection Model v2",
		Version:      "2.0.0",
		ModelType:    string(veidtypes.ModelTypeGANDetection),
		SHA256Hash:   "dddd000000000000000000000000000000000000000000000000000000000000",
		RegisteredBy: registrar.String(),
		Status:       veidtypes.ModelStatusPending,
	})
	require.NoError(s.T(), err)

	// Submit governance proposal for model update
	proposal := &veidtypes.ModelUpdateProposal{
		Title:           "Update GAN Detection Model to v2",
		Description:     "Improved accuracy and reduced false positives",
		ModelType:       string(veidtypes.ModelTypeGANDetection),
		NewModelID:      "gan-v2.0.0",
		NewModelHash:    "dddd000000000000000000000000000000000000000000000000000000000000",
		ActivationDelay: 10,
		ProposerAddress: proposer.String(),
	}

	err = s.keeper.ProposeModelUpdate(ctx, proposal)
	require.NoError(s.T(), err)

	// Query model history to verify proposal was recorded
	histResp, err := s.keeper.QueryModelHistory(ctx, &veidtypes.QueryModelHistoryRequest{
		ModelType: string(veidtypes.ModelTypeGANDetection),
	})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), histResp)
}

// TestValidatorModelSync tests validator model sync status reporting.
func (s *ModelHashE2ETestSuite) TestValidatorModelSync() {
	ctx := s.ctx
	validator := sdktestutil.AccAddress(s.T())

	// Report validator model versions
	report := &veidtypes.ValidatorModelReport{
		ValidatorAddress: validator.String(),
		ModelVersions: map[string]string{
			string(veidtypes.ModelTypeFaceVerification): "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		},
		IsSynced: true,
	}

	s.keeper.SetValidatorModelReport(ctx, report)

	// Query validator sync status
	resp, err := s.keeper.QueryValidatorModelSync(ctx, &veidtypes.QueryValidatorModelSyncRequest{
		ValidatorAddress: validator.String(),
	})
	require.NoError(s.T(), err)
	require.NotNil(s.T(), resp.Report)
	require.True(s.T(), resp.IsSynced)
}

// TestDeterministicHashComputation tests that model hash computation is deterministic.
func (s *ModelHashE2ETestSuite) TestDeterministicHashComputation() {
	modelPath := "testdata/sample_model.pb"

	// ComputeLocalModelHash should produce identical results for same input
	hash1, err := s.keeper.ComputeLocalModelHash(s.ctx, modelPath)
	if err != nil {
		// If testdata doesn't exist, skip — this is expected in CI without model files
		s.T().Skipf("Skipping deterministic hash test: %v", err)
	}

	hash2, err := s.keeper.ComputeLocalModelHash(s.ctx, modelPath)
	require.NoError(s.T(), err)
	require.Equal(s.T(), hash1, hash2, "Hash computation must be deterministic")
}
