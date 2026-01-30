package keeper

import (
	"fmt"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/veid/types"
)

// Common error message formats for marshaling operations
const (
	errMsgMarshalFailed = "failed to marshal: %v"
)

// ============================================================================
// Pipeline Version Management (VE-219)
// ============================================================================

// SetPipelineVersion stores a pipeline version in the state
func (k Keeper) SetPipelineVersion(ctx sdk.Context, pv *types.PipelineVersion) error {
	if err := pv.Validate(); err != nil {
		return types.ErrInvalidPipelineVersion.Wrapf("validation failed: %v", err)
	}

	store := ctx.KVStore(k.skey)
	key := types.PipelineVersionKey(pv.Version)

	bz, err := k.cdc.Marshal(pv)
	if err != nil {
		return types.ErrInvalidPipelineVersion.Wrapf(errMsgMarshalFailed, err)
	}

	store.Set(key, bz)

	k.Logger(ctx).Info("pipeline version stored",
		"version", pv.Version,
		"image_hash", pv.ImageHash,
		"status", pv.Status,
	)

	return nil
}

// GetPipelineVersion retrieves a pipeline version by version string
func (k Keeper) GetPipelineVersion(ctx sdk.Context, version string) (*types.PipelineVersion, bool) {
	store := ctx.KVStore(k.skey)
	key := types.PipelineVersionKey(version)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var pv types.PipelineVersion
	if err := k.cdc.Unmarshal(bz, &pv); err != nil {
		k.Logger(ctx).Error("failed to unmarshal pipeline version",
			"version", version,
			"error", err,
		)
		return nil, false
	}

	return &pv, true
}

// GetActivePipelineVersion returns the currently active pipeline version
func (k Keeper) GetActivePipelineVersion(ctx sdk.Context) (*types.PipelineVersion, error) {
	store := ctx.KVStore(k.skey)
	key := types.ActivePipelineVersionKey()

	bz := store.Get(key)
	if bz == nil {
		return nil, types.ErrNoPipelineVersionActive
	}

	versionStr := string(bz)
	pv, found := k.GetPipelineVersion(ctx, versionStr)
	if !found {
		return nil, types.ErrPipelineVersionNotFound.Wrapf("active version %s not found", versionStr)
	}

	return pv, nil
}

// SetActivePipelineVersion sets the active pipeline version
func (k Keeper) SetActivePipelineVersion(ctx sdk.Context, version string) error {
	// Verify the version exists
	pv, found := k.GetPipelineVersion(ctx, version)
	if !found {
		return types.ErrPipelineVersionNotFound.Wrapf("version %s not found", version)
	}

	// Update the version status to active
	now := ctx.BlockTime()
	pv.Status = string(types.PipelineVersionStatusActive)
	pv.ActivatedAt = &now
	pv.ActivatedAtHeight = ctx.BlockHeight()

	if err := k.SetPipelineVersion(ctx, pv); err != nil {
		return err
	}

	// Store the active version reference
	store := ctx.KVStore(k.skey)
	key := types.ActivePipelineVersionKey()
	store.Set(key, []byte(version))

	k.Logger(ctx).Info("active pipeline version set",
		"version", version,
		"block_height", ctx.BlockHeight(),
	)

	return nil
}

// ActivatePipelineVersion activates a pipeline version for use by validators
func (k Keeper) ActivatePipelineVersion(ctx sdk.Context, version string) error {
	// Get the current active version (if any) and deprecate it
	currentActive, err := k.GetActivePipelineVersion(ctx)
	if err == nil && currentActive != nil {
		now := ctx.BlockTime()
		currentActive.Status = string(types.PipelineVersionStatusDeprecated)
		currentActive.DeprecatedAt = &now
		if err := k.SetPipelineVersion(ctx, currentActive); err != nil {
			return fmt.Errorf("failed to deprecate current version: %w", err)
		}
	}

	// Activate the new version
	return k.SetActivePipelineVersion(ctx, version)
}

// RegisterPipelineVersion registers a new pipeline version (pending activation)
func (k Keeper) RegisterPipelineVersion(
	ctx sdk.Context,
	version string,
	imageHash string,
	imageRef string,
	modelManifest types.ModelManifest,
) (*types.PipelineVersion, error) {
	// Check if version already exists
	if _, found := k.GetPipelineVersion(ctx, version); found {
		return nil, types.ErrPipelineVersionAlreadyExists.Wrapf("version %s already exists", version)
	}

	// Store the model manifest
	if err := k.SetModelManifest(ctx, &modelManifest); err != nil {
		return nil, err
	}

	// Create the pipeline version
	pv := types.NewPipelineVersion(
		version,
		imageHash,
		imageRef,
		modelManifest,
		ctx.BlockTime(),
		ctx.BlockHeight(),
	)

	if err := k.SetPipelineVersion(ctx, pv); err != nil {
		return nil, err
	}

	return pv, nil
}

// ListPipelineVersions returns all registered pipeline versions
func (k Keeper) ListPipelineVersions(ctx sdk.Context) []*types.PipelineVersion {
	store := ctx.KVStore(k.skey)
	prefix := types.PipelineVersionPrefixKey()
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	versions := make([]*types.PipelineVersion, 0)
	for ; iterator.Valid(); iterator.Next() {
		var pv types.PipelineVersion
		if err := k.cdc.Unmarshal(iterator.Value(), &pv); err != nil {
			continue
		}
		versions = append(versions, &pv)
	}

	return versions
}

// ============================================================================
// Model Manifest Management
// ============================================================================

// SetModelManifest stores a model manifest
func (k Keeper) SetModelManifest(ctx sdk.Context, mm *types.ModelManifest) error {
	if err := mm.Validate(); err != nil {
		return types.ErrInvalidModelManifest.Wrapf("validation failed: %v", err)
	}

	store := ctx.KVStore(k.skey)
	key := types.ModelManifestKey(mm.ManifestHash)

	bz, err := k.cdc.Marshal(mm)
	if err != nil {
		return types.ErrInvalidModelManifest.Wrapf(errMsgMarshalFailed, err)
	}

	store.Set(key, bz)
	return nil
}

// GetModelManifest retrieves a model manifest by hash
func (k Keeper) GetModelManifest(ctx sdk.Context, manifestHash string) (*types.ModelManifest, bool) {
	store := ctx.KVStore(k.skey)
	key := types.ModelManifestKey(manifestHash)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var mm types.ModelManifest
	if err := k.cdc.Unmarshal(bz, &mm); err != nil {
		return nil, false
	}

	return &mm, true
}

// ============================================================================
// Pipeline Execution Record Management
// ============================================================================

// SetPipelineExecutionRecord stores an execution record
func (k Keeper) SetPipelineExecutionRecord(
	ctx sdk.Context,
	requestID string,
	record *types.PipelineExecutionRecord,
) error {
	store := ctx.KVStore(k.skey)
	key := types.PipelineExecutionRecordKey(requestID)

	bz, err := k.cdc.Marshal(record)
	if err != nil {
		return types.ErrPipelineExecutionFailed.Wrapf(errMsgMarshalFailed, err)
	}

	store.Set(key, bz)
	return nil
}

// GetPipelineExecutionRecord retrieves an execution record by request ID
func (k Keeper) GetPipelineExecutionRecord(
	ctx sdk.Context,
	requestID string,
) (*types.PipelineExecutionRecord, bool) {
	store := ctx.KVStore(k.skey)
	key := types.PipelineExecutionRecordKey(requestID)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var record types.PipelineExecutionRecord
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return nil, false
	}

	return &record, true
}

// SetPipelineExecutionByValidator stores an execution record by validator
func (k Keeper) SetPipelineExecutionByValidator(
	ctx sdk.Context,
	validatorAddress sdk.ValAddress,
	requestID string,
	record *types.PipelineExecutionRecord,
) error {
	store := ctx.KVStore(k.skey)
	key := types.PipelineExecutionByValidatorKey(validatorAddress.Bytes(), requestID)

	bz, err := k.cdc.Marshal(record)
	if err != nil {
		return types.ErrPipelineExecutionFailed.Wrapf(errMsgMarshalFailed, err)
	}

	store.Set(key, bz)
	return nil
}

// GetPipelineExecutionByValidator retrieves an execution record by validator and request
func (k Keeper) GetPipelineExecutionByValidator(
	ctx sdk.Context,
	validatorAddress sdk.ValAddress,
	requestID string,
) (*types.PipelineExecutionRecord, bool) {
	store := ctx.KVStore(k.skey)
	key := types.PipelineExecutionByValidatorKey(validatorAddress.Bytes(), requestID)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var record types.PipelineExecutionRecord
	if err := k.cdc.Unmarshal(bz, &record); err != nil {
		return nil, false
	}

	return &record, true
}

// ============================================================================
// Pipeline Version Verification
// ============================================================================

// VerifyPipelineVersion checks if a validator is using the correct pipeline version
func (k Keeper) VerifyPipelineVersion(
	ctx sdk.Context,
	pipelineVersion string,
	imageHash string,
	modelManifestHash string,
) error {
	// Get the active pipeline version
	activePV, err := k.GetActivePipelineVersion(ctx)
	if err != nil {
		return err
	}

	// Verify version matches
	if activePV.Version != pipelineVersion {
		return types.ErrPipelineVersionMismatch.Wrapf(
			"expected version %s, got %s",
			activePV.Version,
			pipelineVersion,
		)
	}

	// Verify image hash matches
	if activePV.ImageHash != imageHash {
		return types.ErrPipelineVersionMismatch.Wrapf(
			"image hash mismatch for version %s",
			pipelineVersion,
		)
	}

	// Verify model manifest hash matches
	if activePV.ModelManifest.ManifestHash != modelManifestHash {
		return types.ErrModelManifestMismatch.Wrapf(
			"model manifest hash mismatch for version %s",
			pipelineVersion,
		)
	}

	return nil
}

// ComparePipelineExecutions compares two execution records for consensus
func (k Keeper) ComparePipelineExecutions(
	ctx sdk.Context,
	proposerRecord *types.PipelineExecutionRecord,
	validatorRecord *types.PipelineExecutionRecord,
) (*types.PipelineComparisonResult, error) {
	result := types.CompareExecutionRecords(proposerRecord, validatorRecord)

	if !result.Match {
		k.Logger(ctx).Warn("pipeline execution mismatch",
			"differences", result.Differences,
		)
	}

	return result, nil
}

// PipelineExecutionParams contains parameters for recording a pipeline execution
type PipelineExecutionParams struct {
	RequestID           string
	ValidatorAddress    sdk.ValAddress
	PipelineVersion     string
	ImageHash           string
	ModelManifestHash   string
	InputHash           string
	OutputHash          string
	ExecutionDurationMs int64
}

// RecordPipelineExecution records a pipeline execution for consensus verification
func (k Keeper) RecordPipelineExecution(
	ctx sdk.Context,
	params PipelineExecutionParams,
) (*types.PipelineExecutionRecord, error) {
	// Verify pipeline version is valid
	if err := k.VerifyPipelineVersion(ctx, params.PipelineVersion, params.ImageHash, params.ModelManifestHash); err != nil {
		return nil, err
	}

	// Create execution record
	record := types.NewPipelineExecutionRecord(
		params.PipelineVersion,
		params.ImageHash,
		params.ModelManifestHash,
		ctx.BlockTime(),
	)
	record.InputHash = params.InputHash
	record.OutputHash = params.OutputHash
	record.ExecutionDurationMs = params.ExecutionDurationMs
	record.DeterminismVerified = true

	// Store by request ID
	if err := k.SetPipelineExecutionRecord(ctx, params.RequestID, record); err != nil {
		return nil, err
	}

	// Store by validator
	if err := k.SetPipelineExecutionByValidator(ctx, params.ValidatorAddress, params.RequestID, record); err != nil {
		return nil, err
	}

	return record, nil
}

// ============================================================================
// Conformance Test Results
// ============================================================================

// ConformanceTestResult represents the result of a conformance test
type ConformanceTestResult struct {
	// TestID is the unique identifier for this test
	TestID string `json:"test_id" protobuf:"bytes,1,opt,name=test_id,json=testId,proto3"`

	// PipelineVersion is the version being tested
	PipelineVersion string `json:"pipeline_version" protobuf:"bytes,2,opt,name=pipeline_version,json=pipelineVersion,proto3"`

	// TestInputHash is the hash of the test inputs
	TestInputHash string `json:"test_input_hash" protobuf:"bytes,3,opt,name=test_input_hash,json=testInputHash,proto3"`

	// ExpectedOutputHash is the expected output hash
	ExpectedOutputHash string `json:"expected_output_hash" protobuf:"bytes,4,opt,name=expected_output_hash,json=expectedOutputHash,proto3"`

	// ActualOutputHash is the actual output hash
	ActualOutputHash string `json:"actual_output_hash" protobuf:"bytes,5,opt,name=actual_output_hash,json=actualOutputHash,proto3"`

	// Passed indicates if the test passed
	Passed bool `json:"passed" protobuf:"varint,6,opt,name=passed,proto3"`

	// ValidatorAddress is the validator that ran the test
	ValidatorAddress string `json:"validator_address" protobuf:"bytes,7,opt,name=validator_address,json=validatorAddress,proto3"`

	// TestedAt is when the test was run
	TestedAt time.Time `json:"tested_at" protobuf:"bytes,8,opt,name=tested_at,json=testedAt,proto3,stdtime"`

	// BlockHeight is the block at which the test was run
	BlockHeight int64 `json:"block_height" protobuf:"varint,9,opt,name=block_height,json=blockHeight,proto3"`
}

// SetConformanceTestResult stores a conformance test result
func (k Keeper) SetConformanceTestResult(ctx sdk.Context, result *ConformanceTestResult) error {
	store := ctx.KVStore(k.skey)
	key := types.PipelineConformanceResultKey(result.TestID)

	bz, err := k.cdc.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal conformance result: %w", err)
	}

	store.Set(key, bz)
	return nil
}

// GetConformanceTestResult retrieves a conformance test result
func (k Keeper) GetConformanceTestResult(ctx sdk.Context, testID string) (*ConformanceTestResult, bool) {
	store := ctx.KVStore(k.skey)
	key := types.PipelineConformanceResultKey(testID)

	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}

	var result ConformanceTestResult
	if err := k.cdc.Unmarshal(bz, &result); err != nil {
		return nil, false
	}

	return &result, true
}

// ListConformanceTestResults returns all conformance test results
func (k Keeper) ListConformanceTestResults(ctx sdk.Context) []*ConformanceTestResult {
	store := ctx.KVStore(k.skey)
	prefix := types.PipelineConformanceResultPrefixKey()
	iterator := storetypes.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()

	results := make([]*ConformanceTestResult, 0)
	for ; iterator.Valid(); iterator.Next() {
		var result ConformanceTestResult
		if err := k.cdc.Unmarshal(iterator.Value(), &result); err != nil {
			continue
		}
		results = append(results, &result)
	}

	return results
}

// ============================================================================
// Proto.Message interface stubs for ConformanceTestResult
// ============================================================================

// ConformanceTestResult proto stubs
func (*ConformanceTestResult) ProtoMessage()            {}
func (m *ConformanceTestResult) Reset()                 { *m = ConformanceTestResult{} }
func (m *ConformanceTestResult) String() string         { return fmt.Sprintf("%+v", *m) }
