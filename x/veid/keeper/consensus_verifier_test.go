package keeper

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

const (
	testChainID        = "test-chain"
	testModelVersion   = "v1.0.0"
	testKeyFingerprint = "validator1"
)

// closeStoreIfNeeded closes the CommitMultiStore if it implements io.Closer.
func closeStoreIfNeeded(stateStore store.CommitMultiStore) {
	if stateStore == nil {
		return
	}
	if closer, ok := stateStore.(io.Closer); ok {
		_ = closer.Close()
	}
}

// createTestContext creates a minimal SDK context for testing
func createTestContext(t *testing.T) sdk.Context {
	t.Helper()
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	storeKey := storetypes.NewKVStoreKey("test")
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	require.NoError(t, err)
	return sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())
}

// ============================================================================
// Mock ML Scorer for Testing
// ============================================================================

type mockMLScorer struct {
	version       string
	healthy       bool
	fixedScore    uint32
	returnError   error
	scoreOverride map[string]uint32 // Allow per-input score overrides
}

func newMockMLScorer(healthy bool, fixedScore uint32) *mockMLScorer {
	return &mockMLScorer{
		version:       testModelVersion,
		healthy:       healthy,
		fixedScore:    fixedScore,
		scoreOverride: make(map[string]uint32),
	}
}

func (m *mockMLScorer) Score(input *ScoringInput) (*ScoringOutput, error) {
	if m.returnError != nil {
		return nil, m.returnError
	}

	score := m.fixedScore
	// Check for overrides based on account address
	if override, ok := m.scoreOverride[input.AccountAddress]; ok {
		score = override
	}

	return &ScoringOutput{
		Score:          score,
		ModelVersion:   m.version,
		ReasonCodes:    []types.ReasonCode{types.ReasonCodeSuccess},
		ScopeScores:    make(map[string]uint32),
		Confidence:     0.9,
		ProcessingTime: 50,
		InputHash:      input.ComputeInputHash(),
	}, nil
}

func (m *mockMLScorer) GetModelVersion() string {
	return m.version
}

func (m *mockMLScorer) IsHealthy() bool {
	return m.healthy
}

func (m *mockMLScorer) Close() error {
	return nil
}

// ============================================================================
// Mock Key Provider
// ============================================================================

type mockKeyProvider struct {
	privateKey  []byte
	fingerprint string
}

func newMockKeyProvider() *mockKeyProvider {
	key := sha256.Sum256([]byte(testKeyFingerprint))
	return &mockKeyProvider{
		privateKey:  key[:],
		fingerprint: testKeyFingerprint,
	}
}

func (m *mockKeyProvider) GetPrivateKey() ([]byte, error) {
	return m.privateKey, nil
}

func (m *mockKeyProvider) GetKeyFingerprint() string {
	return m.fingerprint
}

func (m *mockKeyProvider) Close() error {
	return nil
}

// ============================================================================
// Test ConsensusParams
// ============================================================================

func TestDefaultConsensusParams(t *testing.T) {
	params := DefaultConsensusParams()

	require.Equal(t, uint32(0), params.ScoreTolerance, "default tolerance should be 0 (exact match)")
	require.True(t, params.RequireModelMatch, "should require model version match by default")
	require.True(t, params.RequireInputHashMatch, "should require input hash match by default")
	require.Equal(t, 0.67, params.MinValidatorAgreement, "should require 2/3 validator agreement")
	require.Equal(t, int64(1000), params.MaxVerificationTimeMs, "should have 1 second max verification time")
}

// ============================================================================
// Test CompareResults
// ============================================================================

func TestCompareResults_ExactMatch(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 85)
	keyProvider := newMockKeyProvider()
	params := DefaultConsensusParams()

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)

	inputHash := sha256.Sum256([]byte("test-input"))

	proposed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	computed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	result := cv.CompareResults(proposed, computed)

	require.True(t, result.Match, "exact match should pass")
	require.Empty(t, result.Differences, "no differences expected")
	require.Equal(t, int32(0), result.ScoreDifference, "score difference should be 0")
	require.True(t, result.ModelVersionMatch, "model version should match")
	require.True(t, result.InputHashMatch, "input hash should match")
	require.True(t, result.StatusMatch, "status should match")
}

func TestCompareResults_ScoreMismatch(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 85)
	keyProvider := newMockKeyProvider()
	params := DefaultConsensusParams() // tolerance = 0

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)

	inputHash := sha256.Sum256([]byte("test-input"))

	proposed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	computed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        83, // Different score
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	result := cv.CompareResults(proposed, computed)

	require.False(t, result.Match, "should not match with different scores")
	require.NotEmpty(t, result.Differences, "should have differences")
	require.Equal(t, int32(2), result.ScoreDifference, "score difference should be 2")
	require.Equal(t, uint32(85), result.ProposedScore)
	require.Equal(t, uint32(83), result.ComputedScore)
}

func TestCompareResults_ScoreWithinTolerance(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 85)
	keyProvider := newMockKeyProvider()
	params := ConsensusParams{
		ScoreTolerance:        2, // Allow up to 2 points difference
		RequireModelMatch:     true,
		RequireInputHashMatch: true,
	}

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)

	inputHash := sha256.Sum256([]byte("test-input"))

	proposed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	computed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        83, // Within tolerance
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	result := cv.CompareResults(proposed, computed)

	require.True(t, result.Match, "should match when score difference within tolerance")
	require.Equal(t, int32(2), result.ScoreDifference)
}

func TestCompareResults_ScoreExceedsTolerance(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 85)
	keyProvider := newMockKeyProvider()
	params := ConsensusParams{
		ScoreTolerance:        2, // Allow up to 2 points difference
		RequireModelMatch:     true,
		RequireInputHashMatch: true,
	}

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)

	inputHash := sha256.Sum256([]byte("test-input"))

	proposed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	computed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        82, // Exceeds tolerance (difference of 3)
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	result := cv.CompareResults(proposed, computed)

	require.False(t, result.Match, "should not match when score exceeds tolerance")
	require.Equal(t, int32(3), result.ScoreDifference)
}

func TestCompareResults_StatusMismatch(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 85)
	keyProvider := newMockKeyProvider()
	params := DefaultConsensusParams()

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)

	inputHash := sha256.Sum256([]byte("test-input"))

	proposed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	computed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusPartial, // Different status
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	result := cv.CompareResults(proposed, computed)

	require.False(t, result.Match, "should not match with different status")
	require.False(t, result.StatusMatch)
	require.NotEmpty(t, result.Differences)
}

func TestCompareResults_ModelVersionMismatch(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 85)
	keyProvider := newMockKeyProvider()
	params := DefaultConsensusParams()

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)

	inputHash := sha256.Sum256([]byte("test-input"))

	proposed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	computed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v2.0.0", // Different model version
		InputHash:    inputHash[:],
	}

	result := cv.CompareResults(proposed, computed)

	require.False(t, result.Match, "should not match with different model version")
	require.False(t, result.ModelVersionMatch)
}

func TestCompareResults_InputHashMismatch(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 85)
	keyProvider := newMockKeyProvider()
	params := DefaultConsensusParams()

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)

	inputHash1 := sha256.Sum256([]byte("test-input-1"))
	inputHash2 := sha256.Sum256([]byte("test-input-2"))

	proposed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash1[:],
	}

	computed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash2[:], // Different input hash
	}

	result := cv.CompareResults(proposed, computed)

	require.False(t, result.Match, "should not match with different input hash")
	require.False(t, result.InputHashMatch)
}

func TestCompareResults_ModelMatchNotRequired(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 85)
	keyProvider := newMockKeyProvider()
	params := ConsensusParams{
		ScoreTolerance:        0,
		RequireModelMatch:     false, // Don't require model match
		RequireInputHashMatch: true,
	}

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)

	inputHash := sha256.Sum256([]byte("test-input"))

	proposed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	computed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v2.0.0", // Different but not required to match
		InputHash:    inputHash[:],
	}

	result := cv.CompareResults(proposed, computed)

	require.True(t, result.Match, "should match when model version not required")
}

// ============================================================================
// Test ValidateModelVersion
// ============================================================================

func TestValidateModelVersion_Match(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 85)
	keyProvider := newMockKeyProvider()
	params := DefaultConsensusParams()

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)
	ctx := createTestContext(t)

	err := cv.ValidateModelVersion(ctx, "v1.0.0")
	require.NoError(t, err, "should pass when versions match")
}

func TestValidateModelVersion_Mismatch(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 85)
	keyProvider := newMockKeyProvider()
	params := DefaultConsensusParams()

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)
	ctx := createTestContext(t)

	err := cv.ValidateModelVersion(ctx, "v2.0.0")
	require.Error(t, err, "should fail when versions don't match")
	require.Contains(t, err.Error(), "model version mismatch")
}

func TestValidateModelVersion_UnhealthyScorer(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(false, 85) // Not healthy
	keyProvider := newMockKeyProvider()
	params := DefaultConsensusParams()

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)
	ctx := createTestContext(t)

	err := cv.ValidateModelVersion(ctx, "v1.0.0")
	require.Error(t, err, "should fail when scorer is not healthy")
	require.Contains(t, err.Error(), "not healthy")
}

// ============================================================================
// Test ComputeResultHash
// ============================================================================

func TestComputeResultHash_Deterministic(t *testing.T) {
	inputHash := sha256.Sum256([]byte("test-input"))

	result := types.VerificationResult{
		RequestID:      "req-123",
		AccountAddress: "virt1abc123",
		Score:          85,
		Status:         types.VerificationResultStatusSuccess,
		ModelVersion:   "v1.0.0",
		InputHash:      inputHash[:],
		BlockHeight:    1000,
	}

	hash1 := ComputeResultHash(result)
	hash2 := ComputeResultHash(result)

	require.True(t, bytes.Equal(hash1, hash2), "hash should be deterministic")
	require.Len(t, hash1, 32, "hash should be 32 bytes (SHA256)")
}

func TestComputeResultHash_DifferentScores(t *testing.T) {
	inputHash := sha256.Sum256([]byte("test-input"))

	result1 := types.VerificationResult{
		RequestID:      "req-123",
		AccountAddress: "virt1abc123",
		Score:          85,
		Status:         types.VerificationResultStatusSuccess,
		ModelVersion:   "v1.0.0",
		InputHash:      inputHash[:],
		BlockHeight:    1000,
	}

	result2 := types.VerificationResult{
		RequestID:      "req-123",
		AccountAddress: "virt1abc123",
		Score:          84, // Different score
		Status:         types.VerificationResultStatusSuccess,
		ModelVersion:   "v1.0.0",
		InputHash:      inputHash[:],
		BlockHeight:    1000,
	}

	hash1 := ComputeResultHash(result1)
	hash2 := ComputeResultHash(result2)

	require.False(t, bytes.Equal(hash1, hash2), "different scores should produce different hashes")
}

func TestComputeResultHash_DifferentStatus(t *testing.T) {
	inputHash := sha256.Sum256([]byte("test-input"))

	result1 := types.VerificationResult{
		RequestID:      "req-123",
		AccountAddress: "virt1abc123",
		Score:          85,
		Status:         types.VerificationResultStatusSuccess,
		ModelVersion:   "v1.0.0",
		InputHash:      inputHash[:],
		BlockHeight:    1000,
	}

	result2 := types.VerificationResult{
		RequestID:      "req-123",
		AccountAddress: "virt1abc123",
		Score:          85,
		Status:         types.VerificationResultStatusPartial, // Different status
		ModelVersion:   "v1.0.0",
		InputHash:      inputHash[:],
		BlockHeight:    1000,
	}

	hash1 := ComputeResultHash(result1)
	hash2 := ComputeResultHash(result2)

	require.False(t, bytes.Equal(hash1, hash2), "different status should produce different hashes")
}

// ============================================================================
// Test VerificationMetrics
// ============================================================================

func TestVerificationMetrics_Fields(t *testing.T) {
	metrics := VerificationMetrics{
		RequestID:         "req-123",
		AccountAddress:    "virt1abc123",
		ProposerScore:     85,
		ComputedScore:     85,
		ScoreDifference:   0,
		Match:             true,
		ModelVersion:      "v1.0.0",
		ComputeTimeMs:     50,
		BlockHeight:       1000,
		ValidatorAddress:  "virt1val123",
		Timestamp:         time.Now().UTC(),
		Status:            types.VerificationResultStatusSuccess,
		InputHashMatch:    true,
		ModelVersionMatch: true,
		ScopeCount:        3,
	}

	require.Equal(t, "req-123", metrics.RequestID)
	require.Equal(t, uint32(85), metrics.ProposerScore)
	require.Equal(t, uint32(85), metrics.ComputedScore)
	require.Equal(t, int32(0), metrics.ScoreDifference)
	require.True(t, metrics.Match)
	require.True(t, metrics.InputHashMatch)
	require.True(t, metrics.ModelVersionMatch)
}

// ============================================================================
// Test VoteExtension
// ============================================================================

func TestVoteExtension_MarshalUnmarshal(t *testing.T) {
	inputHash := sha256.Sum256([]byte("test-input"))

	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

	result := types.VerificationResult{
		RequestID:      "req-123",
		AccountAddress: "virt1abc123",
		Score:          85,
		Status:         types.VerificationResultStatusSuccess,
		ModelVersion:   "v1.0.0",
		InputHash:      inputHash[:],
		BlockHeight:    1000,
	}

	extension.AddResult(result)

	// Marshal
	bz, err := extension.Marshal()
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	// Unmarshal
	decoded, err := UnmarshalVoteExtension(bz)
	require.NoError(t, err)
	require.NotNil(t, decoded)

	require.Equal(t, extension.Version, decoded.Version)
	require.Equal(t, extension.Height, decoded.Height)
	require.Equal(t, extension.ValidatorAddress, decoded.ValidatorAddress)
	require.Equal(t, extension.ModelVersion, decoded.ModelVersion)
	require.Len(t, decoded.VerificationResults, 1)

	extResult := decoded.VerificationResults[0]
	require.Equal(t, "req-123", extResult.RequestID)
	require.Equal(t, uint32(85), extResult.Score)
	require.Equal(t, types.VerificationResultStatusSuccess, extResult.Status)
}

func TestVoteExtension_InputHashTruncation(t *testing.T) {
	inputHash := sha256.Sum256([]byte("test-input")) // 32 bytes

	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

	result := types.VerificationResult{
		RequestID: "req-123",
		Score:     85,
		Status:    types.VerificationResultStatusSuccess,
		InputHash: inputHash[:], // 32 bytes
	}

	extension.AddResult(result)

	// Input hash should be truncated to 8 bytes for efficiency
	require.Len(t, extension.VerificationResults[0].InputHash, 8)
	require.True(t, bytes.Equal(inputHash[:8], extension.VerificationResults[0].InputHash))
}

func TestVoteExtension_InvalidUnmarshal(t *testing.T) {
	_, err := UnmarshalVoteExtension([]byte("invalid json"))
	require.Error(t, err)
}

// ============================================================================
// Test ComparisonResult
// ============================================================================

func TestComparisonResult_AllMatching(t *testing.T) {
	result := ComparisonResult{
		Match:             true,
		Differences:       []string{},
		ScoreDifference:   0,
		ProposedScore:     85,
		ComputedScore:     85,
		ModelVersionMatch: true,
		InputHashMatch:    true,
		StatusMatch:       true,
	}

	require.True(t, result.Match)
	require.Empty(t, result.Differences)
	require.True(t, result.ModelVersionMatch)
	require.True(t, result.InputHashMatch)
	require.True(t, result.StatusMatch)
}

func TestComparisonResult_MultipleDifferences(t *testing.T) {
	result := ComparisonResult{
		Match: false,
		Differences: []string{
			"score difference exceeds tolerance",
			"status mismatch",
			"model version mismatch",
		},
		ScoreDifference:   5,
		ProposedScore:     85,
		ComputedScore:     80,
		ModelVersionMatch: false,
		InputHashMatch:    true,
		StatusMatch:       false,
	}

	require.False(t, result.Match)
	require.Len(t, result.Differences, 3)
	require.False(t, result.ModelVersionMatch)
	require.True(t, result.InputHashMatch)
	require.False(t, result.StatusMatch)
}

// ============================================================================
// Test Edge Cases
// ============================================================================

func TestCompareResults_ZeroScores(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 0)
	keyProvider := newMockKeyProvider()
	params := DefaultConsensusParams()

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)

	inputHash := sha256.Sum256([]byte("test-input"))

	proposed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        0,
		Status:       types.VerificationResultStatusFailed,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	computed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        0,
		Status:       types.VerificationResultStatusFailed,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	result := cv.CompareResults(proposed, computed)

	require.True(t, result.Match)
	require.Equal(t, int32(0), result.ScoreDifference)
}

func TestCompareResults_MaxScores(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 100)
	keyProvider := newMockKeyProvider()
	params := DefaultConsensusParams()

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)

	inputHash := sha256.Sum256([]byte("test-input"))

	proposed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        100,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	computed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        100,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    inputHash[:],
	}

	result := cv.CompareResults(proposed, computed)

	require.True(t, result.Match)
	require.Equal(t, int32(0), result.ScoreDifference)
}

func TestCompareResults_EmptyInputHash(t *testing.T) {
	logger := log.NewNopLogger()
	scorer := newMockMLScorer(true, 85)
	keyProvider := newMockKeyProvider()
	params := DefaultConsensusParams()

	cv := NewConsensusVerifier(nil, scorer, keyProvider, params, logger)

	proposed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    []byte{},
	}

	computed := types.VerificationResult{
		RequestID:    "req-123",
		Score:        85,
		Status:       types.VerificationResultStatusSuccess,
		ModelVersion: "v1.0.0",
		InputHash:    []byte{},
	}

	result := cv.CompareResults(proposed, computed)

	require.True(t, result.Match)
	require.True(t, result.InputHashMatch)
}

// ============================================================================
// Test Logging and Hex Encoding
// ============================================================================

func TestInputHashHexEncoding(t *testing.T) {
	inputHash := sha256.Sum256([]byte("test-input"))
	hexEncoded := hex.EncodeToString(inputHash[:])

	require.Len(t, hexEncoded, 64, "hex encoded SHA256 should be 64 characters")

	// Verify we can decode back
	decoded, err := hex.DecodeString(hexEncoded)
	require.NoError(t, err)
	require.True(t, bytes.Equal(inputHash[:], decoded))
}
