package keeper_test

import (
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

// Test constants for model version tests
const (
	testModelID1       = "model_001"
	testModelID2       = "model_002"
	testModelHash1     = "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"
	testModelHash2     = "b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3"
	testModelHash3     = "c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
	testValidatorAddr1 = "virtengine1validator1"
	testValidatorAddr2 = "virtengine1validator2"
	testRegistrarAddr  = "virtengine1registrar1"
	testProposerAddr   = "virtengine1proposer1"
)

// ============================================================================
// Model Version Keeper Test Suite
// ============================================================================

type ModelVersionKeeperTestSuite struct {
	suite.Suite
	ctx        sdk.Context
	keeper     keeper.Keeper
	cdc        codec.Codec
	stateStore store.CommitMultiStore
}

func TestModelVersionKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(ModelVersionKeeperTestSuite))
}

func (s *ModelVersionKeeperTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	s.ctx = s.createContextWithStore(storeKey)

	// Create keeper
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority")

	// Set default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	// Set default model params
	err = s.keeper.SetModelParams(s.ctx, &types.ModelParams{
		RequiredModelTypes:       []string{string(types.ModelTypeFaceVerification)},
		ActivationDelayBlocks:    100,
		MaxModelAgeDays:          365,
		AllowedRegistrars:        []string{testRegistrarAddr},
		ValidatorSyncGracePeriod: 50,
		ModelUpdateQuorum:        67,
		EnableGovernanceUpdates:  true,
	})
	s.Require().NoError(err)
}

func (s *ModelVersionKeeperTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
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

// TearDownTest closes the IAVL store to stop background pruning goroutines
func (s *ModelVersionKeeperTestSuite) TearDownTest() {
	CloseStoreIfNeeded(s.stateStore)
}

// ============================================================================
// TestRegisterModel
// ============================================================================

func (s *ModelVersionKeeperTestSuite) TestRegisterModel() {
	testCases := []struct {
		name      string
		model     *types.MLModelInfo
		expectErr bool
		errType   error
	}{
		{
			name: "valid model registration",
			model: &types.MLModelInfo{
				ModelID:      testModelID1,
				Name:         "Face Verification Model",
				Version:      "1.0.0",
				ModelType:    string(types.ModelTypeFaceVerification),
				SHA256Hash:   testModelHash1,
				Description:  "Face verification model for identity",
				RegisteredBy: testRegistrarAddr,
			},
			expectErr: false,
		},
		{
			name: "duplicate model registration",
			model: &types.MLModelInfo{
				ModelID:      testModelID1, // Same ID as above
				Name:         "Another Model",
				Version:      "2.0.0",
				ModelType:    string(types.ModelTypeFaceVerification),
				SHA256Hash:   testModelHash2,
				Description:  "Duplicate test",
				RegisteredBy: testRegistrarAddr,
			},
			expectErr: true,
			errType:   types.ErrModelAlreadyExists,
		},
		{
			name:      "nil model",
			model:     nil,
			expectErr: true,
			errType:   types.ErrInvalidModelInfo,
		},
		{
			name: "invalid model type",
			model: &types.MLModelInfo{
				ModelID:      "model_invalid",
				Name:         "Invalid Model",
				Version:      "1.0.0",
				ModelType:    "invalid_type",
				SHA256Hash:   testModelHash1,
				RegisteredBy: testRegistrarAddr,
			},
			expectErr: true,
			errType:   types.ErrInvalidModelInfo,
		},
		{
			name: "empty model ID",
			model: &types.MLModelInfo{
				ModelID:      "",
				Name:         "No ID Model",
				Version:      "1.0.0",
				ModelType:    string(types.ModelTypeFaceVerification),
				SHA256Hash:   testModelHash1,
				RegisteredBy: testRegistrarAddr,
			},
			expectErr: true,
			errType:   types.ErrInvalidModelInfo,
		},
		{
			name: "invalid hash length",
			model: &types.MLModelInfo{
				ModelID:      "model_bad_hash",
				Name:         "Bad Hash Model",
				Version:      "1.0.0",
				ModelType:    string(types.ModelTypeFaceVerification),
				SHA256Hash:   "tooshort",
				RegisteredBy: testRegistrarAddr,
			},
			expectErr: true,
			errType:   types.ErrInvalidModelInfo,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := s.keeper.RegisterModel(s.ctx, tc.model)
			if tc.expectErr {
				s.Require().Error(err)
				if tc.errType != nil {
					s.Require().ErrorIs(err, tc.errType)
				}
			} else {
				s.Require().NoError(err)

				// Verify model was stored
				retrieved, found := s.keeper.GetModel(s.ctx, tc.model.ModelID)
				s.Require().True(found)
				s.Require().Equal(tc.model.ModelID, retrieved.ModelID)
				s.Require().Equal(tc.model.Name, retrieved.Name)
				s.Require().Equal(tc.model.SHA256Hash, retrieved.SHA256Hash)
			}
		})
	}
}

// ============================================================================
// TestGetActiveModel
// ============================================================================

func (s *ModelVersionKeeperTestSuite) TestGetActiveModel() {
	// First register and activate a model
	model := &types.MLModelInfo{
		ModelID:      testModelID1,
		Name:         "Active Test Model",
		Version:      "1.0.0",
		ModelType:    string(types.ModelTypeFaceVerification),
		SHA256Hash:   testModelHash1,
		Description:  "Test model",
		RegisteredBy: testRegistrarAddr,
		Status:       types.ModelStatusActive,
	}

	err := s.keeper.RegisterModel(s.ctx, model)
	s.Require().NoError(err)

	// Activate the model
	err = s.keeper.ActivatePendingModel(s.ctx, string(types.ModelTypeFaceVerification), testModelID1)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		modelType string
		expectErr bool
		errType   error
	}{
		{
			name:      "get active model",
			modelType: string(types.ModelTypeFaceVerification),
			expectErr: false,
		},
		{
			name:      "no active model for type",
			modelType: string(types.ModelTypeLiveness),
			expectErr: true,
			errType:   types.ErrNoActiveModel,
		},
		{
			name:      "invalid model type",
			modelType: "invalid",
			expectErr: true,
			errType:   types.ErrInvalidModelType,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			result, err := s.keeper.GetActiveModel(s.ctx, tc.modelType)
			if tc.expectErr {
				s.Require().Error(err)
				if tc.errType != nil {
					s.Require().ErrorIs(err, tc.errType)
				}
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(result)
				s.Require().Equal(testModelID1, result.ModelID)
			}
		})
	}
}

// ============================================================================
// TestValidateModelHash
// ============================================================================

func (s *ModelVersionKeeperTestSuite) TestValidateModelHash() {
	// Register and activate a model
	model := &types.MLModelInfo{
		ModelID:      testModelID1,
		Name:         "Hash Test Model",
		Version:      "1.0.0",
		ModelType:    string(types.ModelTypeFaceVerification),
		SHA256Hash:   testModelHash1,
		Description:  "Test model for hash validation",
		RegisteredBy: testRegistrarAddr,
		Status:       types.ModelStatusActive,
	}

	err := s.keeper.RegisterModel(s.ctx, model)
	s.Require().NoError(err)

	err = s.keeper.ActivatePendingModel(s.ctx, string(types.ModelTypeFaceVerification), testModelID1)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		modelType string
		hash      string
		expectErr bool
		errType   error
	}{
		{
			name:      "correct hash",
			modelType: string(types.ModelTypeFaceVerification),
			hash:      testModelHash1,
			expectErr: false,
		},
		{
			name:      "wrong hash",
			modelType: string(types.ModelTypeFaceVerification),
			hash:      testModelHash2,
			expectErr: true,
			errType:   types.ErrModelHashMismatch,
		},
		{
			name:      "invalid model type",
			modelType: "invalid",
			hash:      testModelHash1,
			expectErr: true,
			errType:   types.ErrInvalidModelType,
		},
		{
			name:      "no active model",
			modelType: string(types.ModelTypeLiveness),
			hash:      testModelHash1,
			expectErr: true,
			errType:   types.ErrNoActiveModel,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := s.keeper.ValidateModelHash(s.ctx, tc.modelType, tc.hash)
			if tc.expectErr {
				s.Require().Error(err)
				if tc.errType != nil {
					s.Require().ErrorIs(err, tc.errType)
				}
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// ============================================================================
// TestModelUpdateProposal
// ============================================================================

func (s *ModelVersionKeeperTestSuite) TestModelUpdateProposal() {
	// Register the new model first
	newModel := &types.MLModelInfo{
		ModelID:      testModelID2,
		Name:         "New Face Model",
		Version:      "2.0.0",
		ModelType:    string(types.ModelTypeFaceVerification),
		SHA256Hash:   testModelHash2,
		Description:  "Updated face verification model",
		RegisteredBy: testRegistrarAddr,
		Status:       types.ModelStatusPending,
	}

	err := s.keeper.RegisterModel(s.ctx, newModel)
	s.Require().NoError(err)

	testCases := []struct {
		name      string
		proposal  *types.ModelUpdateProposal
		expectErr bool
		errType   error
	}{
		{
			name: "valid proposal",
			proposal: &types.ModelUpdateProposal{
				Title:           "Update Face Verification Model",
				Description:     "Upgrade to v2.0.0 with improved accuracy",
				ModelType:       string(types.ModelTypeFaceVerification),
				NewModelID:      testModelID2,
				NewModelHash:    testModelHash2,
				ActivationDelay: 100,
				ProposerAddress: testProposerAddr,
			},
			expectErr: false,
		},
		{
			name:      "nil proposal",
			proposal:  nil,
			expectErr: true,
			errType:   types.ErrInvalidModelProposal,
		},
		{
			name: "model not found",
			proposal: &types.ModelUpdateProposal{
				Title:           "Non-existent Model",
				Description:     "This model doesn't exist",
				ModelType:       string(types.ModelTypeFaceVerification),
				NewModelID:      "nonexistent_model",
				NewModelHash:    testModelHash3,
				ActivationDelay: 100,
				ProposerAddress: testProposerAddr,
			},
			expectErr: true,
			errType:   types.ErrModelNotFound,
		},
		{
			name: "hash mismatch",
			proposal: &types.ModelUpdateProposal{
				Title:           "Hash Mismatch Test",
				Description:     "Wrong hash provided",
				ModelType:       string(types.ModelTypeFaceVerification),
				NewModelID:      testModelID2,
				NewModelHash:    testModelHash3, // Wrong hash
				ActivationDelay: 100,
				ProposerAddress: testProposerAddr,
			},
			expectErr: true,
			errType:   types.ErrModelHashMismatch,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Reset context for each test to avoid proposal conflicts
			s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)

			err := s.keeper.ProposeModelUpdate(s.ctx, tc.proposal)
			if tc.expectErr {
				s.Require().Error(err)
				if tc.errType != nil {
					s.Require().ErrorIs(err, tc.errType)
				}
			} else {
				s.Require().NoError(err)

				// Verify proposal was stored
				proposal, found := s.keeper.GetPendingProposal(s.ctx, tc.proposal.ModelType)
				s.Require().True(found)
				s.Require().Equal(tc.proposal.Title, proposal.Title)
				s.Require().Equal(tc.proposal.NewModelID, proposal.NewModelID)
				s.Require().Equal(types.ModelProposalStatusPending, proposal.Status)
			}
		})
	}
}

// ============================================================================
// TestActivatePendingModel
// ============================================================================

func (s *ModelVersionKeeperTestSuite) TestActivatePendingModel() {
	// Register first model
	model1 := &types.MLModelInfo{
		ModelID:      testModelID1,
		Name:         "First Model",
		Version:      "1.0.0",
		ModelType:    string(types.ModelTypeFaceVerification),
		SHA256Hash:   testModelHash1,
		Description:  "First model",
		RegisteredBy: testRegistrarAddr,
		Status:       types.ModelStatusPending,
	}

	err := s.keeper.RegisterModel(s.ctx, model1)
	s.Require().NoError(err)

	// Register second model
	model2 := &types.MLModelInfo{
		ModelID:      testModelID2,
		Name:         "Second Model",
		Version:      "2.0.0",
		ModelType:    string(types.ModelTypeFaceVerification),
		SHA256Hash:   testModelHash2,
		Description:  "Second model",
		RegisteredBy: testRegistrarAddr,
		Status:       types.ModelStatusPending,
	}

	err = s.keeper.RegisterModel(s.ctx, model2)
	s.Require().NoError(err)

	// Activate first model
	err = s.keeper.ActivatePendingModel(s.ctx, string(types.ModelTypeFaceVerification), testModelID1)
	s.Require().NoError(err)

	// Verify first model is active
	state, err := s.keeper.GetModelVersionState(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(testModelID1, state.FaceVerificationModel)

	activeModel, err := s.keeper.GetActiveModel(s.ctx, string(types.ModelTypeFaceVerification))
	s.Require().NoError(err)
	s.Require().Equal(testModelID1, activeModel.ModelID)
	s.Require().Equal(types.ModelStatusActive, activeModel.Status)

	// Activate second model (should deprecate first)
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 10)
	err = s.keeper.ActivatePendingModel(s.ctx, string(types.ModelTypeFaceVerification), testModelID2)
	s.Require().NoError(err)

	// Verify second model is now active
	state, err = s.keeper.GetModelVersionState(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(testModelID2, state.FaceVerificationModel)

	// Verify first model is deprecated
	oldModel, found := s.keeper.GetModel(s.ctx, testModelID1)
	s.Require().True(found)
	s.Require().Equal(types.ModelStatusDeprecated, oldModel.Status)

	// Test error cases
	err = s.keeper.ActivatePendingModel(s.ctx, "invalid_type", testModelID2)
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrInvalidModelType)

	err = s.keeper.ActivatePendingModel(s.ctx, string(types.ModelTypeFaceVerification), "nonexistent")
	s.Require().Error(err)
	s.Require().ErrorIs(err, types.ErrModelNotFound)
}

// ============================================================================
// TestModelVersionHistory
// ============================================================================

func (s *ModelVersionKeeperTestSuite) TestModelVersionHistory() {
	// Register and activate first model
	model1 := &types.MLModelInfo{
		ModelID:      testModelID1,
		Name:         "First Model",
		Version:      "1.0.0",
		ModelType:    string(types.ModelTypeFaceVerification),
		SHA256Hash:   testModelHash1,
		Description:  "First model",
		RegisteredBy: testRegistrarAddr,
	}

	err := s.keeper.RegisterModel(s.ctx, model1)
	s.Require().NoError(err)

	err = s.keeper.ActivatePendingModel(s.ctx, string(types.ModelTypeFaceVerification), testModelID1)
	s.Require().NoError(err)

	// Register and activate second model
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 100)

	model2 := &types.MLModelInfo{
		ModelID:      testModelID2,
		Name:         "Second Model",
		Version:      "2.0.0",
		ModelType:    string(types.ModelTypeFaceVerification),
		SHA256Hash:   testModelHash2,
		Description:  "Second model",
		RegisteredBy: testRegistrarAddr,
	}

	err = s.keeper.RegisterModel(s.ctx, model2)
	s.Require().NoError(err)

	err = s.keeper.ActivatePendingModel(s.ctx, string(types.ModelTypeFaceVerification), testModelID2)
	s.Require().NoError(err)

	// Get history
	history := s.keeper.GetModelHistory(s.ctx, string(types.ModelTypeFaceVerification))
	s.Require().NotEmpty(history)
	s.Require().GreaterOrEqual(len(history), 1)

	// Verify latest history entry (reverse order, newest first)
	latest := history[0]
	s.Require().Equal(string(types.ModelTypeFaceVerification), latest.ModelType)
	s.Require().Equal(testModelID1, latest.OldModelID)
	s.Require().Equal(testModelID2, latest.NewModelID)
}

// ============================================================================
// TestValidatorModelSync
// ============================================================================

func (s *ModelVersionKeeperTestSuite) TestValidatorModelSync() {
	// Register and activate a model
	model := &types.MLModelInfo{
		ModelID:      testModelID1,
		Name:         "Sync Test Model",
		Version:      "1.0.0",
		ModelType:    string(types.ModelTypeFaceVerification),
		SHA256Hash:   testModelHash1,
		Description:  "Model for sync testing",
		RegisteredBy: testRegistrarAddr,
	}

	err := s.keeper.RegisterModel(s.ctx, model)
	s.Require().NoError(err)

	err = s.keeper.ActivatePendingModel(s.ctx, string(types.ModelTypeFaceVerification), testModelID1)
	s.Require().NoError(err)

	testCases := []struct {
		name          string
		validatorAddr string
		modelType     string
		hash          string
		expectErr     bool
	}{
		{
			name:          "synced validator",
			validatorAddr: testValidatorAddr1,
			modelType:     string(types.ModelTypeFaceVerification),
			hash:          testModelHash1,
			expectErr:     false,
		},
		{
			name:          "out of sync validator",
			validatorAddr: testValidatorAddr2,
			modelType:     string(types.ModelTypeFaceVerification),
			hash:          testModelHash2, // Wrong hash
			expectErr:     true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := s.keeper.SyncValidatorModel(s.ctx, tc.validatorAddr, tc.modelType, tc.hash)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}

	// Test ReportValidatorModelVersions
	versions := map[string]string{
		string(types.ModelTypeFaceVerification): testModelHash1,
	}

	err = s.keeper.ReportValidatorModelVersions(s.ctx, testValidatorAddr1, versions)
	s.Require().NoError(err)

	report, found := s.keeper.GetValidatorModelReport(s.ctx, testValidatorAddr1)
	s.Require().True(found)
	s.Require().True(report.IsSynced)
	s.Require().Empty(report.MismatchedModels)

	// Report with wrong hash
	badVersions := map[string]string{
		string(types.ModelTypeFaceVerification): testModelHash2,
	}

	err = s.keeper.ReportValidatorModelVersions(s.ctx, testValidatorAddr2, badVersions)
	s.Require().NoError(err)

	report2, found := s.keeper.GetValidatorModelReport(s.ctx, testValidatorAddr2)
	s.Require().True(found)
	s.Require().False(report2.IsSynced)
	s.Require().Contains(report2.MismatchedModels, string(types.ModelTypeFaceVerification))
}

// ============================================================================
// TestModelParams
// ============================================================================

func (s *ModelVersionKeeperTestSuite) TestModelParams() {
	// Get default params
	params, err := s.keeper.GetModelParams(s.ctx)
	s.Require().NoError(err)
	s.Require().NotNil(params)

	// Update params
	newParams := &types.ModelParams{
		RequiredModelTypes:       []string{string(types.ModelTypeFaceVerification), string(types.ModelTypeLiveness)},
		ActivationDelayBlocks:    200,
		MaxModelAgeDays:          180,
		AllowedRegistrars:        []string{testRegistrarAddr, "another_registrar"},
		ValidatorSyncGracePeriod: 100,
		ModelUpdateQuorum:        75,
		EnableGovernanceUpdates:  true,
	}

	err = s.keeper.SetModelParams(s.ctx, newParams)
	s.Require().NoError(err)

	// Verify update
	retrieved, err := s.keeper.GetModelParams(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(newParams.ActivationDelayBlocks, retrieved.ActivationDelayBlocks)
	s.Require().Equal(newParams.MaxModelAgeDays, retrieved.MaxModelAgeDays)
	s.Require().Equal(len(newParams.RequiredModelTypes), len(retrieved.RequiredModelTypes))
}

// ============================================================================
// TestQueryActiveModels
// ============================================================================

func (s *ModelVersionKeeperTestSuite) TestQueryActiveModels() {
	// Register and activate models for different types
	faceModel := &types.MLModelInfo{
		ModelID:      "face_model_1",
		Name:         "Face Model",
		Version:      "1.0.0",
		ModelType:    string(types.ModelTypeFaceVerification),
		SHA256Hash:   testModelHash1,
		RegisteredBy: testRegistrarAddr,
	}

	livenessModel := &types.MLModelInfo{
		ModelID:      "liveness_model_1",
		Name:         "Liveness Model",
		Version:      "1.0.0",
		ModelType:    string(types.ModelTypeLiveness),
		SHA256Hash:   testModelHash2,
		RegisteredBy: testRegistrarAddr,
	}

	err := s.keeper.RegisterModel(s.ctx, faceModel)
	s.Require().NoError(err)
	err = s.keeper.RegisterModel(s.ctx, livenessModel)
	s.Require().NoError(err)

	err = s.keeper.ActivatePendingModel(s.ctx, string(types.ModelTypeFaceVerification), "face_model_1")
	s.Require().NoError(err)
	err = s.keeper.ActivatePendingModel(s.ctx, string(types.ModelTypeLiveness), "liveness_model_1")
	s.Require().NoError(err)

	// Query all active models
	resp, err := s.keeper.QueryActiveModels(s.ctx, nil)
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().Equal(2, len(resp.Models))
	s.Require().Equal("face_model_1", resp.State.FaceVerificationModel)
	s.Require().Equal("liveness_model_1", resp.State.LivenessModel)
}

// ============================================================================
// TestModelGenesisValidation
// ============================================================================

func (s *ModelVersionKeeperTestSuite) TestModelGenesisValidation() {
	// Test valid genesis
	validGenesis := types.ModelGenesisState{
		Models: []types.MLModelInfo{
			{
				ModelID:      testModelID1,
				Name:         "Test Model",
				Version:      "1.0.0",
				ModelType:    string(types.ModelTypeFaceVerification),
				SHA256Hash:   testModelHash1,
				RegisteredBy: testRegistrarAddr,
				Status:       types.ModelStatusActive,
			},
		},
		State: types.ModelVersionState{
			FaceVerificationModel: testModelID1,
		},
		Params: types.DefaultModelParams(),
	}

	err := validGenesis.Validate()
	s.Require().NoError(err)

	// Test invalid genesis - unknown model in state
	invalidGenesis := types.ModelGenesisState{
		Models: []types.MLModelInfo{},
		State: types.ModelVersionState{
			FaceVerificationModel: "nonexistent",
		},
		Params: types.DefaultModelParams(),
	}

	err = invalidGenesis.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "unknown")

	// Test duplicate model IDs
	duplicateGenesis := types.ModelGenesisState{
		Models: []types.MLModelInfo{
			{
				ModelID:      testModelID1,
				Name:         "First",
				Version:      "1.0.0",
				ModelType:    string(types.ModelTypeFaceVerification),
				SHA256Hash:   testModelHash1,
				RegisteredBy: testRegistrarAddr,
			},
			{
				ModelID:      testModelID1, // Duplicate
				Name:         "Second",
				Version:      "2.0.0",
				ModelType:    string(types.ModelTypeFaceVerification),
				SHA256Hash:   testModelHash2,
				RegisteredBy: testRegistrarAddr,
			},
		},
		Params: types.DefaultModelParams(),
	}

	err = duplicateGenesis.Validate()
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "duplicate")
}

// ============================================================================
// TestIsAuthorizedRegistrar
// ============================================================================

func (s *ModelVersionKeeperTestSuite) TestIsAuthorizedRegistrar() {
	// Test authorized registrar
	isAuthorized := s.keeper.IsAuthorizedRegistrar(s.ctx, testRegistrarAddr)
	s.Require().True(isAuthorized)

	// Test authority is always authorized
	isAuthorized = s.keeper.IsAuthorizedRegistrar(s.ctx, "authority")
	s.Require().True(isAuthorized)

	// Test unauthorized address
	isAuthorized = s.keeper.IsAuthorizedRegistrar(s.ctx, "random_address")
	s.Require().False(isAuthorized)

	// Clear allowed registrars - only authority should be allowed
	err := s.keeper.SetModelParams(s.ctx, &types.ModelParams{
		RequiredModelTypes:       []string{},
		ActivationDelayBlocks:    100,
		MaxModelAgeDays:          365,
		AllowedRegistrars:        []string{}, // Empty
		ValidatorSyncGracePeriod: 50,
		ModelUpdateQuorum:        67,
		EnableGovernanceUpdates:  true,
	})
	s.Require().NoError(err)

	isAuthorized = s.keeper.IsAuthorizedRegistrar(s.ctx, testRegistrarAddr)
	s.Require().False(isAuthorized)

	isAuthorized = s.keeper.IsAuthorizedRegistrar(s.ctx, "authority")
	s.Require().True(isAuthorized)
}
