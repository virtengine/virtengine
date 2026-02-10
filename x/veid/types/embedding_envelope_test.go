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
// Embedding Envelope Tests (VE-217: Derived Feature Minimization)
// ============================================================================

func TestEmbeddingEnvelope_NewAndValidate(t *testing.T) {
	hash := sha256.Sum256([]byte("test-embedding-data"))

	envelope := types.NewEmbeddingEnvelope(
		"envelope-001",
		"ve1test1234567890",
		types.EmbeddingTypeFace,
		hash[:],
		"1.0.0",
		"model-hash-abc123",
		512,
		"scope-001",
		time.Now(),
		100,
		"ve1validator123",
	)

	require.NotNil(t, envelope)
	assert.Equal(t, "envelope-001", envelope.EnvelopeID)
	assert.Equal(t, types.EmbeddingTypeFace, envelope.EmbeddingType)
	assert.Equal(t, types.EmbeddingEnvelopeVersion, envelope.Version)
	assert.Equal(t, uint32(512), envelope.Dimension)
	assert.False(t, envelope.Revoked)

	// Validate should pass
	err := envelope.Validate()
	require.NoError(t, err)
}

func TestEmbeddingEnvelope_Validate_Errors(t *testing.T) {
	tests := []struct {
		name   string
		modify func(*types.EmbeddingEnvelope)
		errMsg string
	}{
		{
			name: "empty envelope ID",
			modify: func(e *types.EmbeddingEnvelope) {
				e.EnvelopeID = ""
			},
			errMsg: "envelope_id cannot be empty",
		},
		{
			name: "empty account address",
			modify: func(e *types.EmbeddingEnvelope) {
				e.AccountAddress = ""
			},
			errMsg: "account address cannot be empty",
		},
		{
			name: "invalid embedding type",
			modify: func(e *types.EmbeddingEnvelope) {
				e.EmbeddingType = "invalid"
			},
			errMsg: "invalid embedding_type",
		},
		{
			name: "invalid hash length",
			modify: func(e *types.EmbeddingEnvelope) {
				e.EmbeddingHash = []byte("short")
			},
			errMsg: "embedding_hash must be 32 bytes",
		},
		{
			name: "empty model version",
			modify: func(e *types.EmbeddingEnvelope) {
				e.ModelVersion = ""
			},
			errMsg: "model_version cannot be empty",
		},
		{
			name: "zero dimension",
			modify: func(e *types.EmbeddingEnvelope) {
				e.Dimension = 0
			},
			errMsg: "dimension cannot be zero",
		},
		{
			name: "empty source scope ID",
			modify: func(e *types.EmbeddingEnvelope) {
				e.SourceScopeID = ""
			},
			errMsg: "source_scope_id cannot be empty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hash := sha256.Sum256([]byte("test-data"))
			envelope := types.NewEmbeddingEnvelope(
				"envelope-001",
				"ve1test1234567890",
				types.EmbeddingTypeFace,
				hash[:],
				"1.0.0",
				"model-hash",
				512,
				"scope-001",
				time.Now(),
				100,
				"ve1validator123",
			)

			tc.modify(envelope)
			err := envelope.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

func TestEmbeddingEnvelope_Revoke(t *testing.T) {
	hash := sha256.Sum256([]byte("test-data"))
	envelope := types.NewEmbeddingEnvelope(
		"envelope-001",
		"ve1test1234567890",
		types.EmbeddingTypeFace,
		hash[:],
		"1.0.0",
		"model-hash",
		512,
		"scope-001",
		time.Now(),
		100,
		"ve1validator123",
	)

	require.False(t, envelope.Revoked)

	revokedAt := time.Now()
	envelope.Revoke("user requested", revokedAt)

	assert.True(t, envelope.Revoked)
	assert.Equal(t, "user requested", envelope.RevokedReason)
	assert.NotNil(t, envelope.RevokedAt)
}

func TestEmbeddingEnvelope_MatchesEmbedding(t *testing.T) {
	embeddingData := []byte("test-embedding-data-12345")
	hash := sha256.Sum256(embeddingData)

	envelope := types.NewEmbeddingEnvelope(
		"envelope-001",
		"ve1test1234567890",
		types.EmbeddingTypeFace,
		hash[:],
		"1.0.0",
		"model-hash",
		512,
		"scope-001",
		time.Now(),
		100,
		"ve1validator123",
	)

	// Same data should match
	assert.True(t, envelope.MatchesEmbedding(embeddingData))

	// Different data should not match
	assert.False(t, envelope.MatchesEmbedding([]byte("different-data")))

	// Empty hash should not match
	envelope.EmbeddingHash = nil
	assert.False(t, envelope.MatchesEmbedding(embeddingData))
}

func TestEmbeddingEnvelope_IsActive(t *testing.T) {
	hash := sha256.Sum256([]byte("test-data"))
	now := time.Now()

	envelope := types.NewEmbeddingEnvelope(
		"envelope-001",
		"ve1test1234567890",
		types.EmbeddingTypeFace,
		hash[:],
		"1.0.0",
		"model-hash",
		512,
		"scope-001",
		now,
		100,
		"ve1validator123",
	)

	// Should be active by default
	assert.True(t, envelope.IsActive(now))

	// Should not be active if revoked
	envelope.Revoke("test", now)
	assert.False(t, envelope.IsActive(now))

	// Reset revocation and add expired retention policy
	envelope.Revoked = false
	expiredAt := now.Add(-time.Hour)
	envelope.RetentionPolicy = &types.RetentionPolicy{
		RetentionType: types.RetentionTypeDuration,
		ExpiresAt:     &expiredAt,
	}
	assert.False(t, envelope.IsActive(now))
}

func TestEmbeddingEnvelope_ToOnChainReference(t *testing.T) {
	hash := sha256.Sum256([]byte("test-data"))
	now := time.Now()

	envelope := types.NewEmbeddingEnvelope(
		"envelope-001",
		"ve1test1234567890",
		types.EmbeddingTypeFace,
		hash[:],
		"1.0.0",
		"model-hash",
		512,
		"scope-001",
		now,
		100,
		"ve1validator123",
	)

	ref := envelope.ToOnChainReference()

	assert.Equal(t, envelope.EnvelopeID, ref.EnvelopeID)
	assert.Equal(t, envelope.AccountAddress, ref.AccountAddress)
	assert.Equal(t, envelope.EmbeddingType, ref.EmbeddingType)
	assert.Equal(t, envelope.EmbeddingHash, ref.EmbeddingHash)
	assert.Equal(t, envelope.ModelVersion, ref.ModelVersion)

	// Validate reference
	err := ref.Validate()
	require.NoError(t, err)
}

func TestComputeEmbeddingHash(t *testing.T) {
	data := []byte("test-embedding-vector-data")

	hash := types.ComputeEmbeddingHash(data)

	require.Len(t, hash, 32)

	// Same input should produce same hash
	hash2 := types.ComputeEmbeddingHash(data)
	assert.Equal(t, hash, hash2)

	// Different input should produce different hash
	hash3 := types.ComputeEmbeddingHash([]byte("different-data"))
	assert.NotEqual(t, hash, hash3)
}

func TestEmbeddingType_Validation(t *testing.T) {
	validTypes := types.AllEmbeddingTypes()

	for _, et := range validTypes {
		assert.True(t, types.IsValidEmbeddingType(et))
	}

	assert.False(t, types.IsValidEmbeddingType("invalid_type"))
	assert.False(t, types.IsValidEmbeddingType(""))
}
