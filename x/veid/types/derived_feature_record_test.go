package types_test

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Derived Feature Verification Record Tests (VE-217)
// ============================================================================

func TestDerivedFeatureVerificationRecord_New(t *testing.T) {
	now := time.Now()

	record := types.NewDerivedFeatureVerificationRecord(
		"record-001",
		"ve1test1234567890",
		"request-001",
		"1.0.0",
		"model-hash-abc",
		now,
		100,
		"ve1validator123",
	)

	require.NotNil(t, record)
	assert.Equal(t, "record-001", record.RecordID)
	assert.Equal(t, "ve1test1234567890", record.AccountAddress)
	assert.Equal(t, "request-001", record.RequestID)
	assert.Equal(t, types.DerivedFeatureVerificationRecordVersion, record.Version)
	assert.Equal(t, types.VerificationResultStatusFailed, record.Status) // Default
	assert.False(t, record.Finalized)
	assert.Empty(t, record.FeatureReferences)
}

func TestDerivedFeatureVerificationRecord_Validate(t *testing.T) {
	now := time.Now()

	record := types.NewDerivedFeatureVerificationRecord(
		"record-001",
		"ve1test1234567890",
		"request-001",
		"1.0.0",
		"model-hash-abc",
		now,
		100,
		"ve1validator123",
	)

	// Add valid feature reference
	hash := sha256.Sum256([]byte("feature-data"))
	ref := types.NewDerivedFeatureReference(
		"face_embedding",
		hash[:],
		"scope-001",
		30,
	)
	ref.SetMatch(95)
	record.AddFeatureReference(ref)
	record.ComputeCompositeHash()
	record.SetSuccess(95, 98)

	err := record.Validate()
	require.NoError(t, err)
}

func TestDerivedFeatureVerificationRecord_Validate_Errors(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name   string
		modify func(*types.DerivedFeatureVerificationRecord)
		errMsg string
	}{
		{
			name: "empty record ID",
			modify: func(r *types.DerivedFeatureVerificationRecord) {
				r.RecordID = ""
			},
			errMsg: "record_id cannot be empty",
		},
		{
			name: "empty account address",
			modify: func(r *types.DerivedFeatureVerificationRecord) {
				r.AccountAddress = ""
			},
			errMsg: "account_address cannot be empty",
		},
		{
			name: "empty request ID",
			modify: func(r *types.DerivedFeatureVerificationRecord) {
				r.RequestID = ""
			},
			errMsg: "request_id cannot be empty",
		},
		{
			name: "empty model version",
			modify: func(r *types.DerivedFeatureVerificationRecord) {
				r.ModelVersion = ""
			},
			errMsg: "model_version cannot be empty",
		},
		{
			name: "score exceeds max",
			modify: func(r *types.DerivedFeatureVerificationRecord) {
				r.Score = 150
			},
			errMsg: "exceeds maximum",
		},
		{
			name: "confidence exceeds 100",
			modify: func(r *types.DerivedFeatureVerificationRecord) {
				r.Confidence = 150
			},
			errMsg: "exceeds 100",
		},
		{
			name: "invalid feature hash length",
			modify: func(r *types.DerivedFeatureVerificationRecord) {
				r.FeatureReferences = []types.DerivedFeatureReference{
					{
						FeatureType:   "face",
						FeatureHash:   []byte("short"),
						SourceScopeID: "scope-001",
						MatchResult:   types.FeatureMatchResultMatch,
					},
				}
			},
			errMsg: "hash must be 32 bytes",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			record := types.NewDerivedFeatureVerificationRecord(
				"record-001",
				"ve1test1234567890",
				"request-001",
				"1.0.0",
				"model-hash",
				now,
				100,
				"ve1validator123",
			)

			tc.modify(record)
			err := record.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

func TestDerivedFeatureVerificationRecord_FeatureReferences(t *testing.T) {
	now := time.Now()

	record := types.NewDerivedFeatureVerificationRecord(
		"record-001",
		"ve1test1234567890",
		"request-001",
		"1.0.0",
		"model-hash-abc",
		now,
		100,
		"ve1validator123",
	)

	// Add multiple feature references
	hash1 := sha256.Sum256([]byte("face-embedding"))
	ref1 := types.NewDerivedFeatureReference("face_embedding", hash1[:], "scope-001", 30)
	ref1.SetMatch(95)
	record.AddFeatureReference(ref1)

	hash2 := sha256.Sum256([]byte("doc-face"))
	ref2 := types.NewDerivedFeatureReference("document_face", hash2[:], "scope-002", 25)
	ref2.SetMatch(90)
	record.AddFeatureReference(ref2)

	hash3 := sha256.Sum256([]byte("doc-hash"))
	ref3 := types.NewDerivedFeatureReference("document_hash", hash3[:], "scope-002", 30)
	ref3.SetNoMatch()
	record.AddFeatureReference(ref3)

	assert.Len(t, record.FeatureReferences, 3)

	// Test GetMatchingFeatures
	matching := record.GetMatchingFeatures()
	assert.Len(t, matching, 2)

	// Test GetFailedFeatures
	failed := record.GetFailedFeatures()
	assert.Len(t, failed, 1)
	assert.Equal(t, types.FeatureMatchResultNoMatch, failed[0].MatchResult)
}

func TestDerivedFeatureVerificationRecord_ComputeCompositeHash(t *testing.T) {
	now := time.Now()

	record := types.NewDerivedFeatureVerificationRecord(
		"record-001",
		"ve1test1234567890",
		"request-001",
		"1.0.0",
		"model-hash-abc",
		now,
		100,
		"ve1validator123",
	)

	hash1 := sha256.Sum256([]byte("face-embedding"))
	ref1 := types.NewDerivedFeatureReference("face_embedding", hash1[:], "scope-001", 30)
	record.AddFeatureReference(ref1)

	// Compute composite hash
	record.ComputeCompositeHash()

	assert.NotNil(t, record.CompositeHash)
	assert.Len(t, record.CompositeHash, 32)

	// Same inputs should produce same hash
	originalHash := record.CompositeHash
	record.ComputeCompositeHash()
	assert.Equal(t, originalHash, record.CompositeHash)
}

func TestDerivedFeatureVerificationRecord_WeightedScore(t *testing.T) {
	now := time.Now()

	record := types.NewDerivedFeatureVerificationRecord(
		"record-001",
		"ve1test1234567890",
		"request-001",
		"1.0.0",
		"model-hash-abc",
		now,
		100,
		"ve1validator123",
	)

	// Add features with weights
	hash1 := sha256.Sum256([]byte("feature1"))
	ref1 := types.NewDerivedFeatureReference("feature1", hash1[:], "scope-001", 50)
	ref1.SetMatch(100) // 100% score with weight 50
	record.AddFeatureReference(ref1)

	hash2 := sha256.Sum256([]byte("feature2"))
	ref2 := types.NewDerivedFeatureReference("feature2", hash2[:], "scope-002", 50)
	ref2.SetMatch(80) // 80% score with weight 50
	record.AddFeatureReference(ref2)

	// Weighted average: (100*50 + 80*50) / (50+50) = 90
	score := record.ComputeWeightedScore()
	assert.Equal(t, uint32(90), score)
}

func TestDerivedFeatureVerificationRecord_Consensus(t *testing.T) {
	now := time.Now()

	record := types.NewDerivedFeatureVerificationRecord(
		"record-001",
		"ve1test1234567890",
		"request-001",
		"1.0.0",
		"model-hash-abc",
		now,
		100,
		"ve1validator123",
	)

	// Add consensus votes
	vote1 := types.ConsensusVote{
		ValidatorAddress: "val1",
		Agreed:           true,
		ComputedScore:    90,
		VotedAt:          now,
		BlockHeight:      101,
	}
	record.AddConsensusVote(vote1)

	vote2 := types.ConsensusVote{
		ValidatorAddress: "val2",
		Agreed:           true,
		ComputedScore:    90,
		VotedAt:          now,
		BlockHeight:      101,
	}
	record.AddConsensusVote(vote2)

	vote3 := types.ConsensusVote{
		ValidatorAddress: "val3",
		Agreed:           false,
		ComputedScore:    85,
		VotedAt:          now,
		BlockHeight:      101,
	}
	record.AddConsensusVote(vote3)

	assert.Equal(t, 2, record.CountAgreements())
	assert.Equal(t, 1, record.CountDisagreements())

	// Test consensus with 3 validators (need > 2/3 = > 2)
	assert.False(t, record.HasConsensus(3)) // 2 > 2 is false

	// Add another agreeing vote
	vote4 := types.ConsensusVote{
		ValidatorAddress: "val4",
		Agreed:           true,
		ComputedScore:    90,
		VotedAt:          now,
		BlockHeight:      101,
	}
	record.AddConsensusVote(vote4)

	// Now with 4 validators: 3 > (4*2/3) = 3 > 2 = true
	assert.True(t, record.HasConsensus(4))
}

func TestDerivedFeatureVerificationRecord_Finalize(t *testing.T) {
	now := time.Now()

	record := types.NewDerivedFeatureVerificationRecord(
		"record-001",
		"ve1test1234567890",
		"request-001",
		"1.0.0",
		"model-hash-abc",
		now,
		100,
		"ve1validator123",
	)

	assert.False(t, record.Finalized)

	finalizedAt := now.Add(time.Hour)
	record.Finalize(finalizedAt, 200)

	assert.True(t, record.Finalized)
	assert.NotNil(t, record.FinalizedAt)
	assert.Equal(t, finalizedAt.Unix(), record.FinalizedAt.Unix())
	assert.NotNil(t, record.FinalizedAtBlock)
	assert.Equal(t, int64(200), *record.FinalizedAtBlock)
}

func TestDerivedFeatureReference_SetResults(t *testing.T) {
	hash := sha256.Sum256([]byte("test"))
	ref := types.NewDerivedFeatureReference("test", hash[:], "scope-001", 30)

	// Initially skipped
	assert.Equal(t, types.FeatureMatchResultSkipped, ref.MatchResult)
	assert.Equal(t, uint32(0), ref.MatchScore)

	// Set match
	ref.SetMatch(95)
	assert.Equal(t, types.FeatureMatchResultMatch, ref.MatchResult)
	assert.Equal(t, uint32(95), ref.MatchScore)

	// Set no match
	ref.SetNoMatch()
	assert.Equal(t, types.FeatureMatchResultNoMatch, ref.MatchResult)
	assert.Equal(t, uint32(0), ref.MatchScore)

	// Set partial
	ref.SetPartial(50)
	assert.Equal(t, types.FeatureMatchResultPartial, ref.MatchResult)
	assert.Equal(t, uint32(50), ref.MatchScore)

	// Set error
	ref.SetError()
	assert.Equal(t, types.FeatureMatchResultError, ref.MatchResult)
	assert.Equal(t, uint32(0), ref.MatchScore)
}

func TestFeatureMatchResult_Validation(t *testing.T) {
	validResults := types.AllFeatureMatchResults()

	for _, r := range validResults {
		assert.True(t, types.IsValidFeatureMatchResult(r))
	}

	assert.False(t, types.IsValidFeatureMatchResult("invalid"))
	assert.False(t, types.IsValidFeatureMatchResult(""))
}
