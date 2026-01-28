package keeper

import (
	"bytes"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Test VoteExtension Creation
// ============================================================================

func TestNewVoteExtension(t *testing.T) {
	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

	require.Equal(t, uint8(VoteExtensionVersion), extension.Version)
	require.Equal(t, int64(1000), extension.Height)
	require.Equal(t, "virt1val123", extension.ValidatorAddress)
	require.Equal(t, "v1.0.0", extension.ModelVersion)
	require.Empty(t, extension.VerificationResults)
	require.False(t, extension.Timestamp.IsZero())
}

func TestVoteExtension_AddResult(t *testing.T) {
	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

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

	extension.AddResult(result)

	require.Len(t, extension.VerificationResults, 1)

	extResult := extension.VerificationResults[0]
	require.Equal(t, "req-123", extResult.RequestID)
	require.Equal(t, uint32(85), extResult.Score)
	require.Equal(t, types.VerificationResultStatusSuccess, extResult.Status)
	require.NotEmpty(t, extResult.ResultHash)
}

func TestVoteExtension_AddMultipleResults(t *testing.T) {
	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

	for i := 0; i < 5; i++ {
		inputHash := sha256.Sum256([]byte("test-input-" + string(rune('a'+i))))
		result := types.VerificationResult{
			RequestID:    "req-" + string(rune('a'+i)),
			Score:        uint32(80 + i),
			Status:       types.VerificationResultStatusSuccess,
			ModelVersion: "v1.0.0",
			InputHash:    inputHash[:],
		}
		extension.AddResult(result)
	}

	require.Len(t, extension.VerificationResults, 5)

	// Verify each result
	for i := 0; i < 5; i++ {
		require.Equal(t, "req-"+string(rune('a'+i)), extension.VerificationResults[i].RequestID)
		require.Equal(t, uint32(80+i), extension.VerificationResults[i].Score)
	}
}

// ============================================================================
// Test VoteExtension Serialization
// ============================================================================

func TestVoteExtension_Marshal(t *testing.T) {
	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

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
	extension.AddResult(result)

	bz, err := extension.Marshal()
	require.NoError(t, err)
	require.NotEmpty(t, bz)

	// Verify it's valid JSON
	require.True(t, bz[0] == '{', "should be valid JSON object")
}

func TestVoteExtension_UnmarshalVoteExtension(t *testing.T) {
	original := NewVoteExtension(1000, "virt1val123", "v1.0.0")

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
	original.AddResult(result)

	bz, err := original.Marshal()
	require.NoError(t, err)

	decoded, err := UnmarshalVoteExtension(bz)
	require.NoError(t, err)
	require.NotNil(t, decoded)

	require.Equal(t, original.Version, decoded.Version)
	require.Equal(t, original.Height, decoded.Height)
	require.Equal(t, original.ValidatorAddress, decoded.ValidatorAddress)
	require.Equal(t, original.ModelVersion, decoded.ModelVersion)
	require.Len(t, decoded.VerificationResults, 1)
}

func TestVoteExtension_UnmarshalInvalid(t *testing.T) {
	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "empty bytes",
			data: []byte{},
		},
		{
			name: "invalid json",
			data: []byte("not json"),
		},
		{
			name: "partial json",
			data: []byte("{\"version\":1,"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := UnmarshalVoteExtension(tc.data)
			require.Error(t, err)
		})
	}
}

func TestVoteExtension_RoundTrip(t *testing.T) {
	// Create extension with multiple results
	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

	results := []types.VerificationResult{
		{
			RequestID:    "req-1",
			Score:        85,
			Status:       types.VerificationResultStatusSuccess,
			ModelVersion: "v1.0.0",
			InputHash:    []byte("hash1hash1hash1hash1hash1hash1"),
		},
		{
			RequestID:    "req-2",
			Score:        72,
			Status:       types.VerificationResultStatusPartial,
			ModelVersion: "v1.0.0",
			InputHash:    []byte("hash2hash2hash2hash2hash2hash2"),
		},
		{
			RequestID:    "req-3",
			Score:        0,
			Status:       types.VerificationResultStatusFailed,
			ModelVersion: "v1.0.0",
			InputHash:    []byte("hash3hash3hash3hash3hash3hash3"),
		},
	}

	for _, r := range results {
		extension.AddResult(r)
	}

	// Marshal and unmarshal
	bz, err := extension.Marshal()
	require.NoError(t, err)

	decoded, err := UnmarshalVoteExtension(bz)
	require.NoError(t, err)

	// Verify all results preserved
	require.Len(t, decoded.VerificationResults, 3)

	for i, extResult := range decoded.VerificationResults {
		require.Equal(t, results[i].RequestID, extResult.RequestID)
		require.Equal(t, results[i].Score, extResult.Score)
		require.Equal(t, results[i].Status, extResult.Status)
	}
}

// ============================================================================
// Test VoteExtensionResult
// ============================================================================

func TestVoteExtensionResult_InputHashTruncation(t *testing.T) {
	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

	// Full SHA256 hash (32 bytes)
	inputHash := sha256.Sum256([]byte("test-input"))

	result := types.VerificationResult{
		RequestID: "req-123",
		Score:     85,
		Status:    types.VerificationResultStatusSuccess,
		InputHash: inputHash[:],
	}

	extension.AddResult(result)

	// Verify truncation to 8 bytes
	require.Len(t, extension.VerificationResults[0].InputHash, 8)
	require.True(t, bytes.Equal(inputHash[:8], extension.VerificationResults[0].InputHash))
}

func TestVoteExtensionResult_ShortInputHash(t *testing.T) {
	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

	// Short hash (less than 8 bytes)
	shortHash := []byte("short")

	result := types.VerificationResult{
		RequestID: "req-123",
		Score:     85,
		Status:    types.VerificationResultStatusSuccess,
		InputHash: shortHash,
	}

	extension.AddResult(result)

	// Should keep original short hash
	require.Len(t, extension.VerificationResults[0].InputHash, 5)
	require.True(t, bytes.Equal(shortHash, extension.VerificationResults[0].InputHash))
}

func TestVoteExtensionResult_ResultHash(t *testing.T) {
	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

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

	extension.AddResult(result)

	// Verify result hash is computed
	require.NotEmpty(t, extension.VerificationResults[0].ResultHash)
	require.Len(t, extension.VerificationResults[0].ResultHash, 32)

	// Verify it matches ComputeResultHash
	expectedHash := ComputeResultHash(result)
	require.True(t, bytes.Equal(expectedHash, extension.VerificationResults[0].ResultHash))
}

// ============================================================================
// Test VoteExtension Size
// ============================================================================

func TestVoteExtension_Size(t *testing.T) {
	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

	// Add 10 results
	for i := 0; i < 10; i++ {
		inputHash := sha256.Sum256([]byte("test-input-" + string(rune('a'+i))))
		result := types.VerificationResult{
			RequestID:      "req-" + string(rune('a'+i)),
			AccountAddress: "virt1abc123",
			Score:          uint32(80 + i),
			Status:         types.VerificationResultStatusSuccess,
			ModelVersion:   "v1.0.0",
			InputHash:      inputHash[:],
			BlockHeight:    1000,
		}
		extension.AddResult(result)
	}

	bz, err := extension.Marshal()
	require.NoError(t, err)

	// Vote extensions should be reasonably sized
	// With 10 results, should be under 10KB
	require.Less(t, len(bz), 10*1024, "vote extension should be under 10KB")
}

// ============================================================================
// Test VoteExtension Empty Cases
// ============================================================================

func TestVoteExtension_Empty(t *testing.T) {
	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

	require.Empty(t, extension.VerificationResults)

	bz, err := extension.Marshal()
	require.NoError(t, err)

	decoded, err := UnmarshalVoteExtension(bz)
	require.NoError(t, err)
	require.Empty(t, decoded.VerificationResults)
}

// ============================================================================
// Test VoteExtension Timestamp
// ============================================================================

func TestVoteExtension_Timestamp(t *testing.T) {
	// NewVoteExtension uses a deterministic timestamp (Unix epoch) for consensus safety
	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")

	// The default timestamp should be Unix epoch (zero time)
	require.True(t, extension.Timestamp.Equal(time.Unix(0, 0).UTC()))

	// Test NewVoteExtensionWithTime with custom timestamp
	customTime := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)
	extensionWithTime := NewVoteExtensionWithTime(1000, "virt1val123", "v1.0.0", customTime)
	require.True(t, extensionWithTime.Timestamp.Equal(customTime))
}

// ============================================================================
// Test VoteExtension Version
// ============================================================================

func TestVoteExtensionVersion(t *testing.T) {
	require.Equal(t, 1, VoteExtensionVersion, "current version should be 1")

	extension := NewVoteExtension(1000, "virt1val123", "v1.0.0")
	require.Equal(t, uint8(1), extension.Version)
}

// ============================================================================
// Test VoteExtensionHandler
// ============================================================================

func TestNewVoteExtensionHandler(t *testing.T) {
	keyProvider := newMockKeyProvider("validator1")
	handler := NewVoteExtensionHandler(nil, keyProvider)

	require.NotNil(t, handler)
	require.Nil(t, handler.keeper)
	require.NotNil(t, handler.keyProvider)
}
