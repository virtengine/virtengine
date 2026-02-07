package keeper_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Model Hash Governance Test Suite
// ============================================================================

type ModelHashGovernanceTestSuite struct {
	suite.Suite
	ctx        sdk.Context
	keeper     keeper.Keeper
	cdc        codec.Codec
	stateStore store.CommitMultiStore
}

func TestModelHashGovernanceTestSuite(t *testing.T) {
	suite.Run(t, new(ModelHashGovernanceTestSuite))
}

func (s *ModelHashGovernanceTestSuite) SetupTest() {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	s.ctx = s.createContextWithStore(storeKey)
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority")

	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	err = s.keeper.SetModelParams(s.ctx, &types.ModelParams{
		RequiredModelTypes:       []string{string(types.ModelTypeFaceVerification)},
		ActivationDelayBlocks:    100,
		MaxModelAgeDays:          365,
		AllowedRegistrars:        []string{"virtengine1registrar1"},
		ValidatorSyncGracePeriod: 50,
		ModelUpdateQuorum:        67,
		EnableGovernanceUpdates:  true,
	})
	s.Require().NoError(err)
}

func (s *ModelHashGovernanceTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		s.T().Fatalf("failed to load latest version: %v", err)
	}
	s.stateStore = stateStore

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())
	return ctx
}

func (s *ModelHashGovernanceTestSuite) TearDownTest() {
	CloseStoreIfNeeded(s.stateStore)
}

// registerAndActivateModel registers and activates a model for testing.
func (s *ModelHashGovernanceTestSuite) registerAndActivateModel(modelID, name, hash string) {
	version := "1.0.0"
	modelType := string(types.ModelTypeFaceVerification)
	registrar := "virtengine1registrar1"
	model := &types.MLModelInfo{
		ModelID:      modelID,
		Name:         name,
		Version:      version,
		ModelType:    modelType,
		SHA256Hash:   hash,
		Description:  "test model",
		RegisteredBy: registrar,
	}
	err := s.keeper.RegisterModel(s.ctx, model)
	s.Require().NoError(err)

	err = s.keeper.ActivatePendingModel(s.ctx, modelType, modelID)
	s.Require().NoError(err)
}

// ============================================================================
// Test ValidateModelForScoring
// ============================================================================

func (s *ModelHashGovernanceTestSuite) TestValidateModelForScoring() {
	modelType := string(types.ModelTypeFaceVerification)
	modelHash := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

	s.registerAndActivateModel(
		"model_gov_001",
		"Face Verification v1",
		modelHash,
	)

	// Create a temp dir with a test model file to compute its hash
	tmpDir := s.T().TempDir()
	modelDir := filepath.Join(tmpDir, "model")
	s.Require().NoError(os.MkdirAll(modelDir, 0o755))
	s.Require().NoError(os.WriteFile(filepath.Join(modelDir, "saved_model.pb"), []byte("test-model-data"), 0o600))

	// Compute the hash of the test model
	localHash, err := keeper.ComputeLocalModelHash(modelDir)
	s.Require().NoError(err)
	s.Require().NotEmpty(localHash)

	// This should fail because the local hash won't match the on-chain hash
	err = s.keeper.ValidateModelForScoring(s.ctx, modelType, modelDir)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrModelHashMismatch)
}

func (s *ModelHashGovernanceTestSuite) TestValidateModelForScoring_EmptyModelType() {
	err := s.keeper.ValidateModelForScoring(s.ctx, "", "/some/path")
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidModelType)
}

func (s *ModelHashGovernanceTestSuite) TestValidateModelForScoring_EmptyPath() {
	err := s.keeper.ValidateModelForScoring(s.ctx, string(types.ModelTypeFaceVerification), "")
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidModelInfo)
}

func (s *ModelHashGovernanceTestSuite) TestValidateModelForScoring_NoActiveModel() {
	// No model registered — should fail
	err := s.keeper.ValidateModelForScoring(s.ctx, string(types.ModelTypeFaceVerification), "/some/path")
	s.Require().Error(err)
}

// ============================================================================
// Test ComputeLocalModelHash
// ============================================================================

func (s *ModelHashGovernanceTestSuite) TestComputeLocalModelHash() {
	tmpDir := s.T().TempDir()

	// Create model files
	s.Require().NoError(os.WriteFile(filepath.Join(tmpDir, "saved_model.pb"), []byte("model-proto-data"), 0o600))
	s.Require().NoError(os.MkdirAll(filepath.Join(tmpDir, "variables"), 0o755))
	s.Require().NoError(os.WriteFile(filepath.Join(tmpDir, "variables", "variables.data-00000-of-00001"), []byte("weights"), 0o600))
	s.Require().NoError(os.WriteFile(filepath.Join(tmpDir, "variables", "variables.index"), []byte("index"), 0o600))

	hash, err := keeper.ComputeLocalModelHash(tmpDir)
	s.Require().NoError(err)
	s.Require().Len(hash, 64) // SHA-256 hex

	// Same content should produce same hash (determinism)
	hash2, err := keeper.ComputeLocalModelHash(tmpDir)
	s.Require().NoError(err)
	s.Require().Equal(hash, hash2)
}

func (s *ModelHashGovernanceTestSuite) TestComputeLocalModelHash_ExcludesMetadata() {
	tmpDir := s.T().TempDir()

	s.Require().NoError(os.WriteFile(filepath.Join(tmpDir, "saved_model.pb"), []byte("data"), 0o600))
	hash1, err := keeper.ComputeLocalModelHash(tmpDir)
	s.Require().NoError(err)

	// Adding metadata files should not change hash
	s.Require().NoError(os.WriteFile(filepath.Join(tmpDir, "MODEL_HASH.txt"), []byte("hash-info"), 0o600))
	s.Require().NoError(os.WriteFile(filepath.Join(tmpDir, "export_metadata.json"), []byte("{}"), 0o600))
	s.Require().NoError(os.WriteFile(filepath.Join(tmpDir, "manifest.json"), []byte("{}"), 0o600))

	hash2, err := keeper.ComputeLocalModelHash(tmpDir)
	s.Require().NoError(err)
	s.Require().Equal(hash1, hash2)
}

func (s *ModelHashGovernanceTestSuite) TestComputeLocalModelHash_NonexistentPath() {
	_, err := keeper.ComputeLocalModelHash("/nonexistent/path")
	s.Require().Error(err)
}

func (s *ModelHashGovernanceTestSuite) TestComputeLocalModelHash_NotADirectory() {
	tmpFile := filepath.Join(s.T().TempDir(), "file.txt")
	s.Require().NoError(os.WriteFile(tmpFile, []byte("data"), 0o600))
	_, err := keeper.ComputeLocalModelHash(tmpFile)
	s.Require().Error(err)
}

// ============================================================================
// Test IsModelHashApproved
// ============================================================================

func (s *ModelHashGovernanceTestSuite) TestIsModelHashApproved() {
	modelType := string(types.ModelTypeFaceVerification)
	expectedHash := "e1f2a3b4c5d6e1f2a3b4c5d6e1f2a3b4c5d6e1f2a3b4c5d6e1f2a3b4c5d6e1f2"

	s.registerAndActivateModel(
		"model_approved_001",
		"Approved Model",
		expectedHash,
	)

	// Correct hash should be approved
	s.Require().True(s.keeper.IsModelHashApproved(s.ctx, modelType, expectedHash))

	// Wrong hash should not be approved
	wrongHash := "f1f2a3b4c5d6e1f2a3b4c5d6e1f2a3b4c5d6e1f2a3b4c5d6e1f2a3b4c5d6e1f2"
	s.Require().False(s.keeper.IsModelHashApproved(s.ctx, modelType, wrongHash))
}

// ============================================================================
// Test GetActiveModelHash
// ============================================================================

func (s *ModelHashGovernanceTestSuite) TestGetActiveModelHash() {
	modelType := string(types.ModelTypeFaceVerification)
	expectedHash := "d1e2f3a4b5c6d1e2f3a4b5c6d1e2f3a4b5c6d1e2f3a4b5c6d1e2f3a4b5c6d1e2"

	// No active model yet
	hash, found := s.keeper.GetActiveModelHash(s.ctx, modelType)
	s.Require().False(found)
	s.Require().Empty(hash)

	// Register and activate
	s.registerAndActivateModel(
		"model_hash_001",
		"Hash Test Model",
		expectedHash,
	)

	hash, found = s.keeper.GetActiveModelHash(s.ctx, modelType)
	s.Require().True(found)
	s.Require().Equal(expectedHash, hash)
}

// ============================================================================
// Test EnsureModelGovernanceCompliance
// ============================================================================

func (s *ModelHashGovernanceTestSuite) TestEnsureModelGovernanceCompliance_Synced() {
	modelType := string(types.ModelTypeFaceVerification)
	expectedHash := "c1d2e3f4a5b6c1d2e3f4a5b6c1d2e3f4a5b6c1d2e3f4a5b6c1d2e3f4a5b6c1d2"

	s.registerAndActivateModel(
		"model_compliance_001",
		"Compliance Model",
		expectedHash,
	)

	validatorAddr := "virtengine1val1"
	versions := map[string]string{
		modelType: expectedHash,
	}

	report, err := s.keeper.EnsureModelGovernanceCompliance(s.ctx, validatorAddr, versions)
	s.Require().NoError(err)
	s.Require().NotNil(report)
	s.Require().True(report.IsSynced)
	s.Require().Empty(report.MismatchedModels)
}

func (s *ModelHashGovernanceTestSuite) TestEnsureModelGovernanceCompliance_Mismatched() {
	modelType := string(types.ModelTypeFaceVerification)
	expectedHash := "b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2d3e4f5a6b1c2"

	s.registerAndActivateModel(
		"model_compliance_002",
		"Compliance Model 2",
		expectedHash,
	)

	validatorAddr := "virtengine1val2"
	wrongHash := "fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1"
	versions := map[string]string{
		modelType: wrongHash,
	}

	report, err := s.keeper.EnsureModelGovernanceCompliance(s.ctx, validatorAddr, versions)
	s.Require().NoError(err)
	s.Require().NotNil(report)
	s.Require().False(report.IsSynced)
	s.Require().Contains(report.MismatchedModels, modelType)
}

func (s *ModelHashGovernanceTestSuite) TestEnsureModelGovernanceCompliance_EmptyValidator() {
	_, err := s.keeper.EnsureModelGovernanceCompliance(s.ctx, "", map[string]string{"x": "y"})
	s.Require().Error(err)
}

func (s *ModelHashGovernanceTestSuite) TestEnsureModelGovernanceCompliance_EmptyVersions() {
	_, err := s.keeper.EnsureModelGovernanceCompliance(s.ctx, "virtengine1val1", map[string]string{})
	s.Require().Error(err)
}

// ============================================================================
// Test Full Governance Lifecycle
// ============================================================================

func (s *ModelHashGovernanceTestSuite) TestFullGovernanceLifecycle() {
	modelType := string(types.ModelTypeFaceVerification)
	modelHash := "1111111111111111111111111111111111111111111111111111111111111111"
	newModelHash := "2222222222222222222222222222222222222222222222222222222222222222"

	// Step 1: Register initial model
	s.registerAndActivateModel(
		"model_lifecycle_v1",
		"Trust Score v1",
		modelHash,
	)

	// Verify active hash
	hash, found := s.keeper.GetActiveModelHash(s.ctx, modelType)
	s.Require().True(found)
	s.Require().Equal(modelHash, hash)

	// Step 2: Register new model version
	newModel := &types.MLModelInfo{
		ModelID:      "model_lifecycle_v2",
		Name:         "Trust Score v2",
		Version:      "2.0.0",
		ModelType:    modelType,
		SHA256Hash:   newModelHash,
		Description:  "Improved model",
		RegisteredBy: "virtengine1registrar1",
	}
	err := s.keeper.RegisterModel(s.ctx, newModel)
	s.Require().NoError(err)

	// Step 3: Propose model update
	proposal := &types.ModelUpdateProposal{
		Title:           "Upgrade to v2",
		Description:     "Improved accuracy",
		ModelType:       modelType,
		NewModelID:      "model_lifecycle_v2",
		NewModelHash:    newModelHash,
		ProposerAddress: "virtengine1registrar1",
	}
	err = s.keeper.ProposeModelUpdate(s.ctx, proposal)
	s.Require().NoError(err)

	// Step 4: Approve proposal
	err = s.keeper.ApproveModelProposal(s.ctx, modelType, 1)
	s.Require().NoError(err)

	// Step 5: Advance block height and process activations
	s.ctx = s.ctx.WithBlockHeight(300) // Beyond activation delay of 100
	err = s.keeper.ProcessPendingActivations(s.ctx)
	s.Require().NoError(err)

	// Step 6: Verify new model is active
	hash, found = s.keeper.GetActiveModelHash(s.ctx, modelType)
	s.Require().True(found)
	s.Require().Equal(newModelHash, hash)

	// Old hash should no longer be approved
	s.Require().False(s.keeper.IsModelHashApproved(s.ctx, modelType, modelHash))
	// New hash should be approved
	s.Require().True(s.keeper.IsModelHashApproved(s.ctx, modelType, newModelHash))

	// Step 7: Check history
	history := s.keeper.GetModelHistory(s.ctx, modelType)
	s.Require().NotEmpty(history)
}
