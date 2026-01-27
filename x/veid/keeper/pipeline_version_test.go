package keeper

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

	"github.com/virtengine/virtengine/x/veid/types"
)

// setupPipelineTestKeeper creates a test keeper for pipeline version tests
func setupPipelineTestKeeper(t *testing.T) (Keeper, sdk.Context) {
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

	// Register types needed for pipeline version
	types.RegisterInterfaces(registry)

	keeper := NewKeeper(cdc, storeKey, "authority")

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())

	return keeper, ctx
}

// createTestModelManifest creates a test model manifest
func createTestModelManifest(t *testing.T) types.ModelManifest {
	models := []types.ModelInfo{
		{
			Name:        "deepface_facenet512",
			Version:     "1.0.0",
			WeightsHash: "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			Framework:   "tensorflow",
			Purpose:     string(types.ModelPurposeFaceRecognition),
			InputShape:  []int32{1, 160, 160, 3},
			OutputShape: []int32{1, 512},
		},
		{
			Name:        "craft_text_detection",
			Version:     "1.0.0",
			WeightsHash: "sha256:b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3",
			Framework:   "pytorch",
			Purpose:     string(types.ModelPurposeTextDetection),
		},
		{
			Name:        "unet_face_extraction",
			Version:     "1.0.0",
			WeightsHash: "sha256:c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
			Framework:   "tensorflow",
			Purpose:     string(types.ModelPurposeFaceExtraction),
		},
	}

	return *types.NewModelManifest("1.0.0", models, time.Now().UTC())
}

// TestPipelineVersionRegistration tests registering a new pipeline version
func TestPipelineVersionRegistration(t *testing.T) {
	keeper, ctx := setupPipelineTestKeeper(t)

	manifest := createTestModelManifest(t)

	// Register a new pipeline version
	pv, err := keeper.RegisterPipelineVersion(
		ctx,
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		manifest,
	)
	if err != nil {
		t.Fatalf("failed to register pipeline version: %v", err)
	}

	if pv.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", pv.Version)
	}

	if pv.Status != string(types.PipelineVersionStatusPending) {
		t.Errorf("expected status pending, got %s", pv.Status)
	}

	// Verify it can be retrieved
	retrieved, found := keeper.GetPipelineVersion(ctx, "1.0.0")
	if !found {
		t.Fatal("failed to retrieve registered pipeline version")
	}

	if retrieved.ImageHash != pv.ImageHash {
		t.Errorf("image hash mismatch: expected %s, got %s", pv.ImageHash, retrieved.ImageHash)
	}
}

// TestPipelineVersionDuplicateRegistration tests duplicate registration is rejected
func TestPipelineVersionDuplicateRegistration(t *testing.T) {
	keeper, ctx := setupPipelineTestKeeper(t)

	manifest := createTestModelManifest(t)

	// Register first version
	_, err := keeper.RegisterPipelineVersion(
		ctx,
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		manifest,
	)
	if err != nil {
		t.Fatalf("failed to register pipeline version: %v", err)
	}

	// Try to register same version again
	_, err = keeper.RegisterPipelineVersion(
		ctx,
		"1.0.0",
		"sha256:different0000000000000000000000000000000000000000000000000000000",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		manifest,
	)
	if err == nil {
		t.Error("expected error for duplicate registration, got nil")
	}
}

// TestActivePipelineVersion tests setting and getting active version
func TestActivePipelineVersion(t *testing.T) {
	keeper, ctx := setupPipelineTestKeeper(t)

	manifest := createTestModelManifest(t)

	// Register a pipeline version
	_, err := keeper.RegisterPipelineVersion(
		ctx,
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		manifest,
	)
	if err != nil {
		t.Fatalf("failed to register pipeline version: %v", err)
	}

	// Initially no active version
	_, err = keeper.GetActivePipelineVersion(ctx)
	if err == nil {
		t.Error("expected error when no active version, got nil")
	}

	// Activate the version
	err = keeper.ActivatePipelineVersion(ctx, "1.0.0")
	if err != nil {
		t.Fatalf("failed to activate pipeline version: %v", err)
	}

	// Now should have active version
	active, err := keeper.GetActivePipelineVersion(ctx)
	if err != nil {
		t.Fatalf("failed to get active version: %v", err)
	}

	if active.Version != "1.0.0" {
		t.Errorf("expected active version 1.0.0, got %s", active.Version)
	}

	if active.Status != string(types.PipelineVersionStatusActive) {
		t.Errorf("expected status active, got %s", active.Status)
	}

	if active.ActivatedAt == nil {
		t.Error("expected ActivatedAt to be set")
	}
}

// TestPipelineVersionSuccession tests version succession
func TestPipelineVersionSuccession(t *testing.T) {
	keeper, ctx := setupPipelineTestKeeper(t)

	manifest := createTestModelManifest(t)

	// Register and activate v1.0.0
	_, err := keeper.RegisterPipelineVersion(
		ctx,
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
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

	// Register v2.0.0
	_, err = keeper.RegisterPipelineVersion(
		ctx,
		"2.0.0",
		"sha256:b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3",
		"ghcr.io/virtengine/veid-pipeline:v2.0.0",
		manifest,
	)
	if err != nil {
		t.Fatalf("failed to register v2.0.0: %v", err)
	}

	// Activate v2.0.0 - should deprecate v1.0.0
	err = keeper.ActivatePipelineVersion(ctx, "2.0.0")
	if err != nil {
		t.Fatalf("failed to activate v2.0.0: %v", err)
	}

	// Check v1.0.0 is now deprecated
	v1, found := keeper.GetPipelineVersion(ctx, "1.0.0")
	if !found {
		t.Fatal("v1.0.0 should still exist")
	}

	if v1.Status != string(types.PipelineVersionStatusDeprecated) {
		t.Errorf("expected v1.0.0 status deprecated, got %s", v1.Status)
	}

	if v1.DeprecatedAt == nil {
		t.Error("expected DeprecatedAt to be set")
	}

	// Check v2.0.0 is active
	active, err := keeper.GetActivePipelineVersion(ctx)
	if err != nil {
		t.Fatalf("failed to get active version: %v", err)
	}

	if active.Version != "2.0.0" {
		t.Errorf("expected active version 2.0.0, got %s", active.Version)
	}
}

// TestPipelineVersionVerification tests pipeline version verification
func TestPipelineVersionVerification(t *testing.T) {
	keeper, ctx := setupPipelineTestKeeper(t)

	manifest := createTestModelManifest(t)

	// Register and activate a version
	pv, err := keeper.RegisterPipelineVersion(
		ctx,
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		manifest,
	)
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	err = keeper.ActivatePipelineVersion(ctx, "1.0.0")
	if err != nil {
		t.Fatalf("failed to activate: %v", err)
	}

	// Verify correct version should pass
	err = keeper.VerifyPipelineVersion(
		ctx,
		"1.0.0",
		pv.ImageHash,
		manifest.ManifestHash,
	)
	if err != nil {
		t.Errorf("verification should pass: %v", err)
	}

	// Verify wrong version should fail
	err = keeper.VerifyPipelineVersion(
		ctx,
		"2.0.0",
		pv.ImageHash,
		manifest.ManifestHash,
	)
	if err == nil {
		t.Error("verification with wrong version should fail")
	}

	// Verify wrong image hash should fail
	err = keeper.VerifyPipelineVersion(
		ctx,
		"1.0.0",
		"sha256:wrong00000000000000000000000000000000000000000000000000000000000",
		manifest.ManifestHash,
	)
	if err == nil {
		t.Error("verification with wrong image hash should fail")
	}

	// Verify wrong manifest hash should fail
	err = keeper.VerifyPipelineVersion(
		ctx,
		"1.0.0",
		pv.ImageHash,
		"wrongmanifesthash",
	)
	if err == nil {
		t.Error("verification with wrong manifest hash should fail")
	}
}

// TestPipelineExecutionRecording tests recording execution records
func TestPipelineExecutionRecording(t *testing.T) {
	keeper, ctx := setupPipelineTestKeeper(t)

	manifest := createTestModelManifest(t)

	// Register and activate a version
	pv, err := keeper.RegisterPipelineVersion(
		ctx,
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"ghcr.io/virtengine/veid-pipeline:v1.0.0",
		manifest,
	)
	if err != nil {
		t.Fatalf("failed to register: %v", err)
	}

	err = keeper.ActivatePipelineVersion(ctx, "1.0.0")
	if err != nil {
		t.Fatalf("failed to activate: %v", err)
	}

	// Record an execution
	validatorAddr := sdk.ValAddress([]byte("validator1"))
	record, err := keeper.RecordPipelineExecution(
		ctx,
		PipelineExecutionParams{
			RequestID:           "request-123",
			ValidatorAddress:    validatorAddr,
			PipelineVersion:     "1.0.0",
			ImageHash:           pv.ImageHash,
			ModelManifestHash:   manifest.ManifestHash,
			InputHash:           "inputhash123",
			OutputHash:          "outputhash456",
			ExecutionDurationMs: 250, // 250ms
		},
	)
	if err != nil {
		t.Fatalf("failed to record execution: %v", err)
	}

	if record.PipelineVersion != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", record.PipelineVersion)
	}

	if record.InputHash != "inputhash123" {
		t.Errorf("expected input hash inputhash123, got %s", record.InputHash)
	}

	if record.ExecutionDurationMs != 250 {
		t.Errorf("expected duration 250, got %d", record.ExecutionDurationMs)
	}

	// Retrieve by request ID
	retrieved, found := keeper.GetPipelineExecutionRecord(ctx, "request-123")
	if !found {
		t.Fatal("failed to retrieve execution record")
	}

	if retrieved.OutputHash != "outputhash456" {
		t.Errorf("expected output hash outputhash456, got %s", retrieved.OutputHash)
	}

	// Retrieve by validator
	validatorRecord, found := keeper.GetPipelineExecutionByValidator(ctx, validatorAddr, "request-123")
	if !found {
		t.Fatal("failed to retrieve by validator")
	}

	if validatorRecord.PipelineVersion != record.PipelineVersion {
		t.Error("validator record should match request record")
	}
}

// TestPipelineExecutionComparison tests comparing execution records
func TestPipelineExecutionComparison(t *testing.T) {
	keeper, ctx := setupPipelineTestKeeper(t)

	now := time.Now().UTC()

	// Create two matching records
	record1 := types.NewPipelineExecutionRecord(
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"manifesthash123",
		now,
	)
	record1.InputHash = "inputhash"
	record1.OutputHash = "outputhash"

	record2 := types.NewPipelineExecutionRecord(
		"1.0.0",
		"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
		"manifesthash123",
		now,
	)
	record2.InputHash = "inputhash"
	record2.OutputHash = "outputhash"

	// Compare matching records
	result, err := keeper.ComparePipelineExecutions(ctx, record1, record2)
	if err != nil {
		t.Fatalf("comparison error: %v", err)
	}

	if !result.Match {
		t.Errorf("matching records should match, differences: %v", result.Differences)
	}

	// Modify output hash
	record2.OutputHash = "differentoutput"
	result, err = keeper.ComparePipelineExecutions(ctx, record1, record2)
	if err != nil {
		t.Fatalf("comparison error: %v", err)
	}

	if result.Match {
		t.Error("records with different output should not match")
	}

	if len(result.Differences) == 0 {
		t.Error("expected differences to be recorded")
	}
}

// TestListPipelineVersions tests listing all versions
func TestListPipelineVersions(t *testing.T) {
	keeper, ctx := setupPipelineTestKeeper(t)

	manifest := createTestModelManifest(t)

	// Register multiple versions
	versions := []string{"1.0.0", "1.1.0", "2.0.0"}
	for _, v := range versions {
		_, err := keeper.RegisterPipelineVersion(
			ctx,
			v,
			"sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2",
			"ghcr.io/virtengine/veid-pipeline:v"+v,
			manifest,
		)
		if err != nil {
			t.Fatalf("failed to register %s: %v", v, err)
		}
	}

	// List all versions
	list := keeper.ListPipelineVersions(ctx)
	if len(list) != 3 {
		t.Errorf("expected 3 versions, got %d", len(list))
	}
}

// TestConformanceTestResults tests conformance result storage
func TestConformanceTestResults(t *testing.T) {
	keeper, ctx := setupPipelineTestKeeper(t)

	result := &ConformanceTestResult{
		TestID:             "test-001",
		PipelineVersion:    "1.0.0",
		TestInputHash:      "inputhash",
		ExpectedOutputHash: "expectedhash",
		ActualOutputHash:   "expectedhash",
		Passed:             true,
		ValidatorAddress:   "validator1",
		TestedAt:           time.Now().UTC(),
		BlockHeight:        100,
	}

	// Store result
	err := keeper.SetConformanceTestResult(ctx, result)
	if err != nil {
		t.Fatalf("failed to store result: %v", err)
	}

	// Retrieve result
	retrieved, found := keeper.GetConformanceTestResult(ctx, "test-001")
	if !found {
		t.Fatal("failed to retrieve result")
	}

	if !retrieved.Passed {
		t.Error("expected test to pass")
	}

	if retrieved.PipelineVersion != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", retrieved.PipelineVersion)
	}

	// List results
	results := keeper.ListConformanceTestResults(ctx)
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// TestModelManifestStorage tests model manifest storage
func TestModelManifestStorage(t *testing.T) {
	keeper, ctx := setupPipelineTestKeeper(t)

	manifest := createTestModelManifest(t)

	// Store manifest
	err := keeper.SetModelManifest(ctx, &manifest)
	if err != nil {
		t.Fatalf("failed to store manifest: %v", err)
	}

	// Retrieve manifest
	retrieved, found := keeper.GetModelManifest(ctx, manifest.ManifestHash)
	if !found {
		t.Fatal("failed to retrieve manifest")
	}

	if len(retrieved.Models) != len(manifest.Models) {
		t.Errorf("model count mismatch: expected %d, got %d",
			len(manifest.Models), len(retrieved.Models))
	}
}
