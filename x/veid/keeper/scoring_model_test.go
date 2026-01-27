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

// Test address constants for scoring model tests - valid bech32 addresses
var (
	testScoringModelAddress1 = sdk.AccAddress([]byte("scoring_addr1_______")).String()
	testScoringModelAddress2 = sdk.AccAddress([]byte("scoring_addr2_______")).String()
)

// ============================================================================
// Scoring Model Keeper Test Suite
// ============================================================================

type ScoringModelKeeperTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.Codec
}

func TestScoringModelKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(ScoringModelKeeperTestSuite))
}

func (s *ScoringModelKeeperTestSuite) SetupTest() {
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
}

func (s *ScoringModelKeeperTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		s.T().Fatalf("failed to load latest version: %v", err)
	}

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())
	return ctx
}

// ============================================================================
// Scoring Model Version Tests
// ============================================================================

func (s *ScoringModelKeeperTestSuite) TestSetAndGetScoringModelVersion() {
	model := types.DefaultScoringModel()
	model.Version = "1.0.0"
	model.Description = "Test model"

	// Set
	err := s.keeper.SetScoringModelVersion(s.ctx, model)
	s.Require().NoError(err)

	// Get
	retrieved, found := s.keeper.GetScoringModelVersion(s.ctx, "1.0.0")
	s.Require().True(found)
	s.Require().Equal("1.0.0", retrieved.Version)
	s.Require().Equal("Test model", retrieved.Description)
	s.Require().Equal(model.Weights.FaceSimilarityWeight, retrieved.Weights.FaceSimilarityWeight)
	s.Require().Equal(model.Thresholds.MinFaceSimilarity, retrieved.Thresholds.MinFaceSimilarity)
}

func (s *ScoringModelKeeperTestSuite) TestGetScoringModelVersion_NotFound() {
	_, found := s.keeper.GetScoringModelVersion(s.ctx, "nonexistent")
	s.Require().False(found)
}

func (s *ScoringModelKeeperTestSuite) TestSetScoringModelVersion_InvalidWeights() {
	model := types.ScoringModelVersion{
		Version: "1.0.0",
		Weights: types.ScoringWeights{
			FaceSimilarityWeight: 5000, // Doesn't sum to 10000
		},
		Thresholds: types.DefaultScoringThresholds(),
		CreatedAt:  time.Now(),
	}

	err := s.keeper.SetScoringModelVersion(s.ctx, model)
	s.Require().Error(err)
}

func (s *ScoringModelKeeperTestSuite) TestListScoringModelVersions() {
	// Store multiple versions
	for _, version := range []string{"1.0.0", "1.1.0", "2.0.0"} {
		model := types.DefaultScoringModel()
		model.Version = version
		err := s.keeper.SetScoringModelVersion(s.ctx, model)
		s.Require().NoError(err)
	}

	// List
	versions := s.keeper.ListScoringModelVersions(s.ctx)
	s.Require().Len(versions, 3)

	// Check all versions are present
	versionMap := make(map[string]bool)
	for _, v := range versions {
		versionMap[v.Version] = true
	}
	s.Require().True(versionMap["1.0.0"])
	s.Require().True(versionMap["1.1.0"])
	s.Require().True(versionMap["2.0.0"])
}

// ============================================================================
// Active Scoring Model Tests
// ============================================================================

func (s *ScoringModelKeeperTestSuite) TestSetAndGetActiveScoringModel() {
	// Store a model first
	model := types.DefaultScoringModel()
	model.Version = "1.0.0"
	err := s.keeper.SetScoringModelVersion(s.ctx, model)
	s.Require().NoError(err)

	// Set as active
	err = s.keeper.SetActiveScoringModel(s.ctx, "1.0.0")
	s.Require().NoError(err)

	// Get active
	active, err := s.keeper.GetActiveScoringModel(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal("1.0.0", active.Version)
}

func (s *ScoringModelKeeperTestSuite) TestSetActiveScoringModel_NotFound() {
	err := s.keeper.SetActiveScoringModel(s.ctx, "nonexistent")
	s.Require().Error(err)
}

func (s *ScoringModelKeeperTestSuite) TestGetActiveScoringModel_Default() {
	// Without setting, should return default
	active, err := s.keeper.GetActiveScoringModel(s.ctx)
	s.Require().NoError(err)
	s.Require().Equal(types.DefaultScoringModelVersion, active.Version)
}

func (s *ScoringModelKeeperTestSuite) TestGetActiveScoringModelVersion() {
	// Store and activate a model
	model := types.DefaultScoringModel()
	model.Version = "2.0.0"
	err := s.keeper.SetScoringModelVersion(s.ctx, model)
	s.Require().NoError(err)
	err = s.keeper.SetActiveScoringModel(s.ctx, "2.0.0")
	s.Require().NoError(err)

	version := s.keeper.GetActiveScoringModelVersion(s.ctx)
	s.Require().Equal("2.0.0", version)
}

// ============================================================================
// Scoring History Tests
// ============================================================================

func (s *ScoringModelKeeperTestSuite) TestRecordAndGetScoringResult() {
	summary := &types.EvidenceSummary{
		FinalScore:   75,
		Passed:       true,
		ModelVersion: "1.0.0",
		Contributions: []types.FeatureContribution{
			{FeatureName: types.FeatureNameFaceSimilarity, RawScore: 8500, Weight: 3000, WeightedScore: 2550},
		},
		ReasonCodes:     []types.ScoringReasonCode{types.ScoringReasonSuccess},
		InputHash:       []byte("test-hash"),
		ComputedAt:      time.Now(),
		BlockHeight:     100,
		FeaturePresence: map[string]bool{"face_similarity": true},
	}

	// Record
	err := s.keeper.RecordScoringResult(s.ctx, testScoringModelAddress1, summary)
	s.Require().NoError(err)

	// Get history
	history := s.keeper.GetScoringHistory(s.ctx, testScoringModelAddress1)
	s.Require().Len(history, 1)
	s.Require().Equal(uint32(75), history[0].Score)
	s.Require().Equal("1.0.0", history[0].ModelVersion)
	s.Require().Contains(history[0].ReasonCodes, types.ScoringReasonSuccess)
}

func (s *ScoringModelKeeperTestSuite) TestScoringHistory_MultipleEntries() {
	// Record multiple scores at different blocks
	for i := 0; i < 5; i++ {
		s.ctx = s.ctx.WithBlockHeight(int64(100 + i)).WithBlockTime(time.Now().Add(time.Duration(i) * time.Hour))

		summary := &types.EvidenceSummary{
			FinalScore:   uint32(50 + i*10),
			Passed:       true,
			ModelVersion: "1.0.0",
			ReasonCodes:  []types.ScoringReasonCode{types.ScoringReasonSuccess},
			ComputedAt:   s.ctx.BlockTime(),
			BlockHeight:  s.ctx.BlockHeight(),
		}

		err := s.keeper.RecordScoringResult(s.ctx, testScoringModelAddress1, summary)
		s.Require().NoError(err)
	}

	// Get history (should be newest first)
	history := s.keeper.GetScoringHistory(s.ctx, testScoringModelAddress1)
	s.Require().Len(history, 5)

	// Verify order is newest first
	s.Require().Equal(uint32(90), history[0].Score)
	s.Require().Equal(uint32(50), history[4].Score)
}

func (s *ScoringModelKeeperTestSuite) TestScoringHistory_Paginated() {
	// Record 10 entries
	for i := 0; i < 10; i++ {
		s.ctx = s.ctx.WithBlockHeight(int64(100 + i))

		summary := &types.EvidenceSummary{
			FinalScore:   uint32(50 + i*5),
			Passed:       true,
			ModelVersion: "1.0.0",
			ComputedAt:   time.Now(),
			BlockHeight:  s.ctx.BlockHeight(),
		}

		err := s.keeper.RecordScoringResult(s.ctx, testScoringModelAddress1, summary)
		s.Require().NoError(err)
	}

	// Get first page
	page1 := s.keeper.GetScoringHistoryPaginated(s.ctx, testScoringModelAddress1, 5, 0)
	s.Require().Len(page1, 5)

	// Get second page
	page2 := s.keeper.GetScoringHistoryPaginated(s.ctx, testScoringModelAddress1, 5, 5)
	s.Require().Len(page2, 5)

	// Get beyond available
	page3 := s.keeper.GetScoringHistoryPaginated(s.ctx, testScoringModelAddress1, 5, 15)
	s.Require().Nil(page3)
}

// ============================================================================
// Version Transition Tests
// ============================================================================

func (s *ScoringModelKeeperTestSuite) TestRecordAndGetVersionTransition() {
	err := s.keeper.RecordVersionTransition(
		s.ctx,
		testScoringModelAddress1,
		"1.0.0",
		"2.0.0",
		75,
		80,
		"model upgrade",
	)
	s.Require().NoError(err)

	transitions := s.keeper.GetVersionTransitions(s.ctx, testScoringModelAddress1)
	s.Require().Len(transitions, 1)
	s.Require().Equal("1.0.0", transitions[0].FromVersion)
	s.Require().Equal("2.0.0", transitions[0].ToVersion)
	s.Require().Equal(uint32(75), transitions[0].PreviousScore)
	s.Require().Equal(uint32(80), transitions[0].NewScore)
	s.Require().Equal("model upgrade", transitions[0].TransitionReason)
}

func (s *ScoringModelKeeperTestSuite) TestVersionTransition_MultipleTransitions() {
	// Record multiple transitions
	for i := 0; i < 3; i++ {
		s.ctx = s.ctx.WithBlockHeight(int64(100 + i))

		err := s.keeper.RecordVersionTransition(
			s.ctx,
			testScoringModelAddress1,
			"1."+string(rune('0'+i))+".0",
			"1."+string(rune('0'+i+1))+".0",
			uint32(70+i*5),
			uint32(75+i*5),
			"upgrade",
		)
		s.Require().NoError(err)
	}

	transitions := s.keeper.GetVersionTransitions(s.ctx, testScoringModelAddress1)
	s.Require().Len(transitions, 3)

	// Should be newest first
	s.Require().Equal("1.3.0", transitions[0].ToVersion)
}

// ============================================================================
// Evidence Summary Tests
// ============================================================================

func (s *ScoringModelKeeperTestSuite) TestStoreAndGetEvidenceSummary() {
	summary := &types.EvidenceSummary{
		FinalScore:   82,
		Passed:       true,
		ModelVersion: "1.0.0",
		Contributions: []types.FeatureContribution{
			{
				FeatureName:     types.FeatureNameFaceSimilarity,
				RawScore:        8500,
				Weight:          3000,
				WeightedScore:   2550,
				PassedThreshold: true,
			},
			{
				FeatureName:     types.FeatureNameOCRConfidence,
				RawScore:        7800,
				Weight:          2500,
				WeightedScore:   1950,
				PassedThreshold: true,
			},
		},
		ReasonCodes: []types.ScoringReasonCode{types.ScoringReasonSuccess},
		InputHash:   []byte("input-hash-12345"),
		ComputedAt:  s.ctx.BlockTime(),
		BlockHeight: s.ctx.BlockHeight(),
		FeaturePresence: map[string]bool{
			types.FeatureNameFaceSimilarity: true,
			types.FeatureNameOCRConfidence:  true,
		},
		ThresholdsApplied: map[string]uint32{
			"min_face_similarity": 7000,
			"required_for_pass":   50,
		},
	}

	// Record (which stores summary internally)
	err := s.keeper.RecordScoringResult(s.ctx, testScoringModelAddress1, summary)
	s.Require().NoError(err)

	// Get evidence summary
	retrieved, found := s.keeper.GetEvidenceSummary(s.ctx, testScoringModelAddress1, s.ctx.BlockHeight())
	s.Require().True(found)
	s.Require().Equal(uint32(82), retrieved.FinalScore)
	s.Require().True(retrieved.Passed)
	s.Require().Equal("1.0.0", retrieved.ModelVersion)
	s.Require().Len(retrieved.Contributions, 2)
	s.Require().Equal(types.FeatureNameFaceSimilarity, retrieved.Contributions[0].FeatureName)
}

func (s *ScoringModelKeeperTestSuite) TestGetEvidenceSummary_NotFound() {
	_, found := s.keeper.GetEvidenceSummary(s.ctx, testScoringModelAddress1, 9999)
	s.Require().False(found)
}

// ============================================================================
// Compute Score Tests
// ============================================================================

func (s *ScoringModelKeeperTestSuite) TestComputeScoreWithModel() {
	// Store a model
	model := types.DefaultScoringModel()
	model.Version = "1.0.0"
	err := s.keeper.SetScoringModelVersion(s.ctx, model)
	s.Require().NoError(err)
	err = s.keeper.SetActiveScoringModel(s.ctx, "1.0.0")
	s.Require().NoError(err)

	// Create inputs
	inputs := types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{
			SimilarityScore: 8500,
			Confidence:      9000,
			Present:         true,
		},
		OCRConfidence: types.OCRConfidenceInput{
			OverallConfidence:   8000,
			ExtractedFieldCount: 8,
			ExpectedFieldCount:  10,
			Present:             true,
		},
		DocIntegrity: types.DocIntegrityInput{
			QualityScore:          8500,
			FormatValid:           true,
			TamperDetectionPassed: true,
			Present:               true,
		},
		SaltBinding: types.SaltBindingInput{
			SaltPresent:          true,
			SaltValid:            true,
			ClientSignatureValid: true,
			UserSignatureValid:   true,
		},
		AccountAddress: testScoringModelAddress1,
	}

	summary, err := s.keeper.ComputeScoreWithModel(s.ctx, inputs, "1.0.0")
	s.Require().NoError(err)
	s.Require().NotNil(summary)
	s.Require().Equal("1.0.0", summary.ModelVersion)
	s.Require().Greater(summary.FinalScore, uint32(0))
}

func (s *ScoringModelKeeperTestSuite) TestComputeScoreWithModel_UseActive() {
	// Store and activate a model
	model := types.DefaultScoringModel()
	model.Version = "2.0.0"
	err := s.keeper.SetScoringModelVersion(s.ctx, model)
	s.Require().NoError(err)
	err = s.keeper.SetActiveScoringModel(s.ctx, "2.0.0")
	s.Require().NoError(err)

	inputs := types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{SimilarityScore: 8500, Confidence: 9000, Present: true},
		AccountAddress: testScoringModelAddress1,
	}

	// Empty version string should use active model
	summary, err := s.keeper.ComputeScoreWithModel(s.ctx, inputs, "")
	s.Require().NoError(err)
	s.Require().Equal("2.0.0", summary.ModelVersion)
}

func (s *ScoringModelKeeperTestSuite) TestComputeScoreWithModel_NotFound() {
	inputs := types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{SimilarityScore: 8500, Present: true},
	}

	_, err := s.keeper.ComputeScoreWithModel(s.ctx, inputs, "nonexistent")
	s.Require().Error(err)
}

// ============================================================================
// Initialize Scoring Model Tests
// ============================================================================

func (s *ScoringModelKeeperTestSuite) TestInitializeScoringModel() {
	// Initialize
	err := s.keeper.InitializeScoringModel(s.ctx)
	s.Require().NoError(err)

	// Check active model
	activeVersion := s.keeper.GetActiveScoringModelVersion(s.ctx)
	s.Require().Equal(types.DefaultScoringModelVersion, activeVersion)

	// Get the model
	model, found := s.keeper.GetScoringModelVersion(s.ctx, types.DefaultScoringModelVersion)
	s.Require().True(found)
	s.Require().NotNil(model.ActivatedAt)
}

func (s *ScoringModelKeeperTestSuite) TestInitializeScoringModel_Idempotent() {
	// Initialize twice
	err := s.keeper.InitializeScoringModel(s.ctx)
	s.Require().NoError(err)

	err = s.keeper.InitializeScoringModel(s.ctx)
	s.Require().NoError(err)

	// Should still work
	versions := s.keeper.ListScoringModelVersions(s.ctx)
	s.Require().Len(versions, 1)
}

// ============================================================================
// Integration Test: Full Scoring Flow
// ============================================================================

func (s *ScoringModelKeeperTestSuite) TestFullScoringFlow() {
	// Initialize model
	err := s.keeper.InitializeScoringModel(s.ctx)
	s.Require().NoError(err)

	// Create comprehensive inputs
	inputs := types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{
			SimilarityScore: 8500,
			Confidence:      9000,
			Present:         true,
		},
		OCRConfidence: types.OCRConfidenceInput{
			OverallConfidence:   8000,
			ExtractedFieldCount: 8,
			ExpectedFieldCount:  10,
			Present:             true,
		},
		DocIntegrity: types.DocIntegrityInput{
			QualityScore:          8500,
			FormatValid:           true,
			TemplateMatch:         true,
			TamperDetectionPassed: true,
			ExpiryValid:           true,
			Present:               true,
		},
		SaltBinding: types.SaltBindingInput{
			SaltPresent:          true,
			SaltValid:            true,
			ClientSignatureValid: true,
			UserSignatureValid:   true,
		},
		LivenessCheck: types.LivenessCheckInput{
			LivenessScore:        8000,
			BlinkDetected:        true,
			HeadMovementDetected: true,
			VideoFrameCount:      30,
			Present:              true,
		},
		CaptureQuality: types.CaptureQualityInput{
			OverallQuality:     8000,
			LightingScore:      7500,
			FocusScore:         8500,
			AngleScore:         8000,
			ResolutionAdequate: true,
			Present:            true,
		},
	}

	// Compute and record score
	summary, err := s.keeper.ComputeAndRecordScore(s.ctx, testScoringModelAddress1, inputs)
	s.Require().NoError(err)
	s.Require().NotNil(summary)
	s.Require().True(summary.Passed)
	s.Require().Greater(summary.FinalScore, uint32(50))

	// Verify score was recorded
	score, status, found := s.keeper.GetScore(s.ctx, testScoringModelAddress1)
	s.Require().True(found)
	s.Require().Equal(summary.FinalScore, score)
	s.Require().Equal(types.AccountStatusVerified, status)

	// Verify history was recorded
	history := s.keeper.GetScoringHistory(s.ctx, testScoringModelAddress1)
	s.Require().Len(history, 1)
	s.Require().Equal(summary.FinalScore, history[0].Score)

	// Verify evidence summary was stored
	retrieved, found := s.keeper.GetEvidenceSummary(s.ctx, testScoringModelAddress1, s.ctx.BlockHeight())
	s.Require().True(found)
	s.Require().Equal(summary.FinalScore, retrieved.FinalScore)
}

func (s *ScoringModelKeeperTestSuite) TestFullScoringFlow_VersionTransition() {
	// Initialize with v1
	err := s.keeper.InitializeScoringModel(s.ctx)
	s.Require().NoError(err)

	// Compute first score
	inputs := types.ScoringInputs{
		FaceSimilarity: types.FaceSimilarityInput{SimilarityScore: 8500, Confidence: 9000, Present: true},
		DocIntegrity:   types.DocIntegrityInput{QualityScore: 8000, FormatValid: true, Present: true},
		SaltBinding:    types.SaltBindingInput{SaltPresent: true, SaltValid: true, ClientSignatureValid: true, UserSignatureValid: true},
	}

	summary1, err := s.keeper.ComputeAndRecordScore(s.ctx, testScoringModelAddress1, inputs)
	s.Require().NoError(err)
	firstScore := summary1.FinalScore

	// Upgrade to v2
	modelV2 := types.DefaultScoringModel()
	modelV2.Version = "2.0.0"
	modelV2.Weights.FaceSimilarityWeight = 4000 // Increase face weight
	modelV2.Weights.CaptureQualityWeight = 0    // Remove capture quality weight
	// Rebalance weights to sum to 10000
	modelV2.Weights.OCRConfidenceWeight = 2000
	modelV2.Weights.DocIntegrityWeight = 2000
	modelV2.Weights.SaltBindingWeight = 1000
	modelV2.Weights.LivenessCheckWeight = 1000
	err = s.keeper.SetScoringModelVersion(s.ctx, modelV2)
	s.Require().NoError(err)
	err = s.keeper.SetActiveScoringModel(s.ctx, "2.0.0")
	s.Require().NoError(err)

	// Advance block
	s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)

	// Compute second score with new model
	summary2, err := s.keeper.ComputeAndRecordScore(s.ctx, testScoringModelAddress1, inputs)
	s.Require().NoError(err)
	s.Require().Equal("2.0.0", summary2.ModelVersion)

	// Check version transition was recorded
	transitions := s.keeper.GetVersionTransitions(s.ctx, testScoringModelAddress1)
	s.Require().Len(transitions, 1)
	s.Require().Equal(types.DefaultScoringModelVersion, transitions[0].FromVersion)
	s.Require().Equal("2.0.0", transitions[0].ToVersion)
	s.Require().Equal(firstScore, transitions[0].PreviousScore)
	s.Require().Equal(summary2.FinalScore, transitions[0].NewScore)
}
