package keeper

import (
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	"cosmossdk.io/log"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"

	"github.com/virtengine/virtengine/x/veid/types"
)

// TestPipelineVersionConsensusIntegration tests the full consensus verification flow
// This simulates:
// 1. Proposer computes verification with a specific pipeline version
// 2. Validator recomputes and verifies the pipeline version matches
// 3. Consensus is reached when outputs match
func TestPipelineVersionConsensusIntegration(t *testing.T) {
	// Setup test environment
	keeper, ctx := setupPipelineIntegrationTestKeeper(t)

	// Register and activate a pipeline version
	manifest := createIntegrationTestManifest(t)
	
	pv, err := keeper.RegisterPipelineVersion(
		ctx,
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		manifest,
	)
	if err != nil {
		t.Fatalf("failed to register pipeline: %v", err)
	}

	err = keeper.ActivatePipelineVersion(ctx, "1.0.0")
	if err != nil {
		t.Fatalf("failed to activate pipeline: %v", err)
	}

	// Simulate proposer computing verification
	proposerAddr := sdk.ValAddress([]byte("proposer-validator"))
	requestID := "verification-request-001"

	proposerRecord, err := keeper.RecordPipelineExecution(
		ctx,
		PipelineExecutionParams{
			RequestID:           requestID,
			ValidatorAddress:    proposerAddr,
			PipelineVersion:     "1.0.0",
			ImageHash:           pv.ImageHash,
			ModelManifestHash:   manifest.ManifestHash,
			InputHash:           "inputhash_abc123",
			OutputHash:          "outputhash_xyz789",
			ExecutionDurationMs: 300, // 300ms
		},
	)
	if err != nil {
		t.Fatalf("proposer failed to record execution: %v", err)
	}

	t.Run("ValidatorWithMatchingOutput", func(t *testing.T) {
		// Simulate validator recomputing with same outputs
		validatorAddr := sdk.ValAddress([]byte("validator-1"))

		validatorRecord, err := keeper.RecordPipelineExecution(
			ctx,
			PipelineExecutionParams{
				RequestID:           requestID + "-validator",
				ValidatorAddress:    validatorAddr,
				PipelineVersion:     "1.0.0",
				ImageHash:           pv.ImageHash,
				ModelManifestHash:   manifest.ManifestHash,
				InputHash:           "inputhash_abc123",  // Same input
				OutputHash:          "outputhash_xyz789", // Same output
				ExecutionDurationMs: 280, // Similar execution time
			},
		)
		if err != nil {
			t.Fatalf("validator failed to record execution: %v", err)
		}

		// Compare execution records
		comparison, err := keeper.ComparePipelineExecutions(ctx, proposerRecord, validatorRecord)
		if err != nil {
			t.Fatalf("comparison failed: %v", err)
		}

		if !comparison.Match {
			t.Errorf("expected records to match, differences: %v", comparison.Differences)
		}
	})

	t.Run("ValidatorWithDifferentOutput", func(t *testing.T) {
		// Simulate validator with different output (non-deterministic execution)
		validatorAddr := sdk.ValAddress([]byte("validator-2"))

		validatorRecord, err := keeper.RecordPipelineExecution(
			ctx,
			PipelineExecutionParams{
				RequestID:           requestID + "-validator2",
				ValidatorAddress:    validatorAddr,
				PipelineVersion:     "1.0.0",
				ImageHash:           pv.ImageHash,
				ModelManifestHash:   manifest.ManifestHash,
				InputHash:           "inputhash_abc123",
				OutputHash:          "DIFFERENT_OUTPUT", // Different output!
				ExecutionDurationMs: 250,
			},
		)
		if err != nil {
			t.Fatalf("validator failed to record execution: %v", err)
		}

		// Compare execution records
		comparison, err := keeper.ComparePipelineExecutions(ctx, proposerRecord, validatorRecord)
		if err != nil {
			t.Fatalf("comparison failed: %v", err)
		}

		if comparison.Match {
			t.Error("expected records to NOT match due to different output")
		}

		if len(comparison.Differences) == 0 {
			t.Error("expected differences to be recorded")
		}
	})

	t.Run("ValidatorWithWrongPipelineVersion", func(t *testing.T) {
		// Simulate validator using wrong pipeline version
		validatorAddr := sdk.ValAddress([]byte("validator-3"))

		// This should fail because the version doesn't match
		_, err := keeper.RecordPipelineExecution(
			ctx,
			PipelineExecutionParams{
				RequestID:           requestID + "-validator3",
				ValidatorAddress:    validatorAddr,
				PipelineVersion:     "2.0.0", // Wrong version!
				ImageHash:           pv.ImageHash,
				ModelManifestHash:   manifest.ManifestHash,
				InputHash:           "inputhash_abc123",
				OutputHash:          "outputhash_xyz789",
				ExecutionDurationMs: 250,
			},
		)
		if err == nil {
			t.Error("expected error for wrong pipeline version")
		}
	})

	t.Run("ValidatorWithWrongImageHash", func(t *testing.T) {
		// Simulate validator using wrong image hash
		validatorAddr := sdk.ValAddress([]byte("validator-4"))

		_, err := keeper.RecordPipelineExecution(
			ctx,
			PipelineExecutionParams{
				RequestID:           requestID + "-validator4",
				ValidatorAddress:    validatorAddr,
				PipelineVersion:     "1.0.0",
				ImageHash:           "sha256:wrong00000000000000000000000000000000000000000000000000000000000", // Wrong hash!
				ModelManifestHash:   manifest.ManifestHash,
				InputHash:           "inputhash_abc123",
				OutputHash:          "outputhash_xyz789",
				ExecutionDurationMs: 250,
			},
		)
		if err == nil {
			t.Error("expected error for wrong image hash")
		}
	})
}

// TestMultiValidatorConsensus tests consensus with multiple validators
func TestMultiValidatorConsensus(t *testing.T) {
	keeper, ctx := setupPipelineIntegrationTestKeeper(t)

	// Setup pipeline
	manifest := createIntegrationTestManifest(t)
	pv, _ := keeper.RegisterPipelineVersion(
		ctx,
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		manifest,
	)
	keeper.ActivatePipelineVersion(ctx, "1.0.0")

	// Proposer's execution record
	proposerRecord := types.NewPipelineExecutionRecord(
		"1.0.0",
		pv.ImageHash,
		manifest.ManifestHash,
		ctx.BlockTime(),
	)
	proposerRecord.InputHash = "inputhash_consensus_test"
	proposerRecord.OutputHash = "outputhash_deterministic_result"
	proposerRecord.DeterminismVerified = true

	// Simulate multiple validators
	validators := []struct {
		name       string
		outputHash string
		shouldMatch bool
	}{
		{"validator-1", "outputhash_deterministic_result", true},
		{"validator-2", "outputhash_deterministic_result", true},
		{"validator-3", "outputhash_deterministic_result", true},
		{"validator-4", "outputhash_DIFFERENT", false}, // Bad validator
		{"validator-5", "outputhash_deterministic_result", true},
	}

	matchCount := 0
	mismatchCount := 0

	for _, v := range validators {
		validatorRecord := types.NewPipelineExecutionRecord(
			"1.0.0",
			pv.ImageHash,
			manifest.ManifestHash,
			ctx.BlockTime(),
		)
		validatorRecord.InputHash = "inputhash_consensus_test"
		validatorRecord.OutputHash = v.outputHash
		validatorRecord.DeterminismVerified = true

		comparison, _ := keeper.ComparePipelineExecutions(ctx, proposerRecord, validatorRecord)
		
		if comparison.Match != v.shouldMatch {
			t.Errorf("%s: expected match=%v, got %v", v.name, v.shouldMatch, comparison.Match)
		}

		if comparison.Match {
			matchCount++
		} else {
			mismatchCount++
		}
	}

	// Should have 4 matches (supermajority) and 1 mismatch
	if matchCount != 4 {
		t.Errorf("expected 4 matches, got %d", matchCount)
	}

	if mismatchCount != 1 {
		t.Errorf("expected 1 mismatch, got %d", mismatchCount)
	}

	// 4/5 = 80% > 67% threshold, consensus should pass
	consensusThreshold := 0.67
	matchRatio := float64(matchCount) / float64(len(validators))
	if matchRatio < consensusThreshold {
		t.Errorf("consensus should pass: %f >= %f", matchRatio, consensusThreshold)
	}
}

// TestPipelineVersionUpgrade tests upgrading pipeline version
func TestPipelineVersionUpgrade(t *testing.T) {
	keeper, ctx := setupPipelineIntegrationTestKeeper(t)

	manifest := createIntegrationTestManifest(t)

	// Register and activate v1.0.0
	_, err := keeper.RegisterPipelineVersion(
		ctx,
		"1.0.0",
		"sha256:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		manifest,
	)
	if err != nil {
		t.Fatalf("failed to register v1.0.0: %v", err)
	}

	err = keeper.ActivatePipelineVersion(ctx, "1.0.0")
	if err != nil {
		t.Fatalf("failed to activate v1.0.0: %v", err)
	}

	// Verify v1.0.0 is active
	active, err := keeper.GetActivePipelineVersion(ctx)
	if err != nil {
		t.Fatalf("failed to get active version: %v", err)
	}
	if active.Version != "1.0.0" {
		t.Errorf("expected active version 1.0.0, got %s", active.Version)
	}

	// Register v2.0.0 (pending)
	_, err = keeper.RegisterPipelineVersion(
		ctx,
		"2.0.0",
		"sha256:b1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0b1b2",
		"ghcr.io/virtengine/veid-pipeline:v2.0.0",
		manifest,
	)
	if err != nil {
		t.Fatalf("failed to register v2.0.0: %v", err)
	}

	// v2.0.0 should be pending
	v2, found := keeper.GetPipelineVersion(ctx, "2.0.0")
	if !found {
		t.Fatal("v2.0.0 should exist")
	}
	if v2.Status != string(types.PipelineVersionStatusPending) {
		t.Errorf("v2.0.0 should be pending, got %s", v2.Status)
	}

	// Simulate upgrade: activate v2.0.0
	err = keeper.ActivatePipelineVersion(ctx, "2.0.0")
	if err != nil {
		t.Fatalf("failed to activate v2.0.0: %v", err)
	}

	// Verify v2.0.0 is now active
	active, err = keeper.GetActivePipelineVersion(ctx)
	if err != nil {
		t.Fatalf("failed to get active version: %v", err)
	}
	if active.Version != "2.0.0" {
		t.Errorf("expected active version 2.0.0, got %s", active.Version)
	}

	// Verify v1.0.0 is deprecated
	v1, found := keeper.GetPipelineVersion(ctx, "1.0.0")
	if !found {
		t.Fatal("v1.0.0 should still exist")
	}
	if v1.Status != string(types.PipelineVersionStatusDeprecated) {
		t.Errorf("v1.0.0 should be deprecated, got %s", v1.Status)
	}

	// Validators using old version should fail verification
	err = keeper.VerifyPipelineVersion(
		ctx,
		"1.0.0", // Old version
		"sha256:a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2",
		manifest.ManifestHash,
	)
	if err == nil {
		t.Error("verification with old version should fail")
	}

	// Validators using new version should succeed
	err = keeper.VerifyPipelineVersion(
		ctx,
		"2.0.0",
		"sha256:b1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0b1b2",
		manifest.ManifestHash,
	)
	if err != nil {
		t.Errorf("verification with new version should succeed: %v", err)
	}
}

// TestConformanceTestIntegration tests conformance test recording and retrieval
func TestConformanceTestIntegration(t *testing.T) {
	keeper, ctx := setupPipelineIntegrationTestKeeper(t)

	// Record multiple conformance test results
	tests := []struct {
		testID       string
		passed       bool
		validatorAddr string
	}{
		{"face_detect_001", true, "validator-1"},
		{"face_embed_001", true, "validator-1"},
		{"ocr_extract_001", false, "validator-1"}, // One failure
		{"e2e_score_001", true, "validator-1"},
	}

	for _, tc := range tests {
		result := &ConformanceTestResult{
			TestID:             tc.testID,
			PipelineVersion:    "1.0.0",
			TestInputHash:      "inputhash_" + tc.testID,
			ExpectedOutputHash: "expectedhash_" + tc.testID,
			ActualOutputHash:   "expectedhash_" + tc.testID,
			Passed:             tc.passed,
			ValidatorAddress:   tc.validatorAddr,
			TestedAt:           ctx.BlockTime(),
			BlockHeight:        ctx.BlockHeight(),
		}

		if !tc.passed {
			result.ActualOutputHash = "differenthash"
		}

		err := keeper.SetConformanceTestResult(ctx, result)
		if err != nil {
			t.Fatalf("failed to store result for %s: %v", tc.testID, err)
		}
	}

	// Retrieve and verify results
	for _, tc := range tests {
		result, found := keeper.GetConformanceTestResult(ctx, tc.testID)
		if !found {
			t.Errorf("result not found for %s", tc.testID)
			continue
		}

		if result.Passed != tc.passed {
			t.Errorf("%s: expected passed=%v, got %v", tc.testID, tc.passed, result.Passed)
		}
	}

	// List all results
	results := keeper.ListConformanceTestResults(ctx)
	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}

	// Count passed/failed
	passedCount := 0
	for _, r := range results {
		if r.Passed {
			passedCount++
		}
	}

	if passedCount != 3 {
		t.Errorf("expected 3 passed, got %d", passedCount)
	}
}

// Helper functions

func setupPipelineIntegrationTestKeeper(t *testing.T) (Keeper, sdk.Context) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	types.RegisterInterfaces(registry)

	keeper := NewKeeper(cdc, storeKey, "authority")

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	return keeper, ctx
}

func createIntegrationTestManifest(t *testing.T) types.ModelManifest {
	models := []types.ModelInfo{
		{
			Name:        "deepface_facenet512",
			Version:     "1.0.0",
			WeightsHash: "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			Framework:   "tensorflow",
			Purpose:     string(types.ModelPurposeFaceRecognition),
		},
		{
			Name:        "craft_text_detection",
			Version:     "1.0.0",
			WeightsHash: "sha256:b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3",
			Framework:   "pytorch",
			Purpose:     string(types.ModelPurposeTextDetection),
		},
	}

	return *types.NewModelManifest("1.0.0", models, time.Now().UTC())
}
