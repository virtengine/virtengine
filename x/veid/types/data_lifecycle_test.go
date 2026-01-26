package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pkg.akt.dev/node/x/veid/types"
)

// ============================================================================
// Retention Policy Tests (VE-217: Data Lifecycle)
// ============================================================================

func TestRetentionPolicy_Duration(t *testing.T) {
	now := time.Now()
	durationSeconds := int64(7 * 24 * 60 * 60) // 7 days
	
	policy := types.NewRetentionPolicyDuration(
		"policy-001",
		durationSeconds,
		now,
		100,
		true,
	)

	require.NotNil(t, policy)
	assert.Equal(t, "policy-001", policy.PolicyID)
	assert.Equal(t, types.RetentionTypeDuration, policy.RetentionType)
	assert.Equal(t, durationSeconds, policy.DurationSeconds)
	assert.NotNil(t, policy.ExpiresAt)
	assert.True(t, policy.DeleteOnExpiry)
	assert.True(t, policy.ExtensionAllowed)

	// Validate
	err := policy.Validate()
	require.NoError(t, err)
}

func TestRetentionPolicy_BlockCount(t *testing.T) {
	now := time.Now()
	blockCount := int64(1000)
	createdAtBlock := int64(500)
	
	policy := types.NewRetentionPolicyBlockCount(
		"policy-002",
		blockCount,
		now,
		createdAtBlock,
		true,
	)

	require.NotNil(t, policy)
	assert.Equal(t, types.RetentionTypeBlockCount, policy.RetentionType)
	assert.Equal(t, blockCount, policy.BlockCount)
	assert.NotNil(t, policy.ExpiresAtBlock)
	assert.Equal(t, createdAtBlock+blockCount, *policy.ExpiresAtBlock)

	err := policy.Validate()
	require.NoError(t, err)
}

func TestRetentionPolicy_Indefinite(t *testing.T) {
	now := time.Now()
	
	policy := types.NewRetentionPolicyIndefinite(
		"policy-003",
		now,
		100,
	)

	require.NotNil(t, policy)
	assert.Equal(t, types.RetentionTypeIndefinite, policy.RetentionType)
	assert.False(t, policy.DeleteOnExpiry)
	assert.False(t, policy.ExtensionAllowed)

	err := policy.Validate()
	require.NoError(t, err)
}

func TestRetentionPolicy_UntilRevoked(t *testing.T) {
	now := time.Now()
	
	policy := types.NewRetentionPolicyUntilRevoked(
		"policy-004",
		now,
		100,
	)

	require.NotNil(t, policy)
	assert.Equal(t, types.RetentionTypeUntilRevoked, policy.RetentionType)
	assert.True(t, policy.DeleteOnExpiry)
	assert.False(t, policy.ExtensionAllowed)

	err := policy.Validate()
	require.NoError(t, err)
}

func TestRetentionPolicy_IsExpired(t *testing.T) {
	now := time.Now()
	
	// Test duration-based expiry
	expiredPolicy := types.NewRetentionPolicyDuration(
		"policy-expired",
		-3600, // Expired 1 hour ago (negative duration won't work, let's set manually)
		now.Add(-2*time.Hour),
		100,
		true,
	)
	// Manually set to expired
	expiredTime := now.Add(-time.Hour)
	expiredPolicy.ExpiresAt = &expiredTime
	
	assert.True(t, expiredPolicy.IsExpired(now))
	
	// Test non-expired policy
	futureTime := now.Add(24 * time.Hour)
	activePolicy := &types.RetentionPolicy{
		RetentionType: types.RetentionTypeDuration,
		ExpiresAt:     &futureTime,
	}
	assert.False(t, activePolicy.IsExpired(now))
	
	// Indefinite policies never expire
	indefinitePolicy := types.NewRetentionPolicyIndefinite("policy-indef", now, 100)
	assert.False(t, indefinitePolicy.IsExpired(now))
	assert.False(t, indefinitePolicy.IsExpired(now.Add(100*365*24*time.Hour)))
}

func TestRetentionPolicy_IsExpiredAtBlock(t *testing.T) {
	now := time.Now()
	
	policy := types.NewRetentionPolicyBlockCount(
		"policy-block",
		1000,
		now,
		500,
	)

	// Should expire at block 1500
	assert.False(t, policy.IsExpiredAtBlock(1499))
	assert.True(t, policy.IsExpiredAtBlock(1500))
	assert.True(t, policy.IsExpiredAtBlock(2000))
}

func TestRetentionPolicy_Extend(t *testing.T) {
	now := time.Now()
	durationSeconds := int64(7 * 24 * 60 * 60)
	
	policy := types.NewRetentionPolicyDuration(
		"policy-extend",
		durationSeconds,
		now,
		100,
		true,
	)
	
	originalExpiry := *policy.ExpiresAt

	// Should be able to extend
	assert.True(t, policy.CanExtend())
	err := policy.Extend()
	require.NoError(t, err)

	// Expiry should be extended
	assert.True(t, policy.ExpiresAt.After(originalExpiry))
	assert.Equal(t, uint32(1), policy.CurrentExtensions)

	// Extend again until max
	for i := uint32(1); i < policy.MaxExtensions; i++ {
		err = policy.Extend()
		require.NoError(t, err)
	}

	// Should not be able to extend anymore
	assert.False(t, policy.CanExtend())
	err = policy.Extend()
	require.Error(t, err)
}

func TestRetentionPolicy_Validate_Errors(t *testing.T) {
	tests := []struct {
		name   string
		policy *types.RetentionPolicy
		errMsg string
	}{
		{
			name: "empty policy ID",
			policy: &types.RetentionPolicy{
				Version:       types.RetentionPolicyVersion,
				PolicyID:      "",
				RetentionType: types.RetentionTypeDuration,
			},
			errMsg: "policy_id cannot be empty",
		},
		{
			name: "invalid retention type",
			policy: &types.RetentionPolicy{
				Version:       types.RetentionPolicyVersion,
				PolicyID:      "test",
				RetentionType: "invalid",
				CreatedAt:     time.Now(),
			},
			errMsg: "invalid retention_type",
		},
		{
			name: "duration with non-positive seconds",
			policy: &types.RetentionPolicy{
				Version:         types.RetentionPolicyVersion,
				PolicyID:        "test",
				RetentionType:   types.RetentionTypeDuration,
				DurationSeconds: 0,
				CreatedAt:       time.Now(),
			},
			errMsg: "duration_seconds must be positive",
		},
		{
			name: "block count with non-positive count",
			policy: &types.RetentionPolicy{
				Version:       types.RetentionPolicyVersion,
				PolicyID:      "test",
				RetentionType: types.RetentionTypeBlockCount,
				BlockCount:    0,
				CreatedAt:     time.Now(),
			},
			errMsg: "block_count must be positive",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.policy.Validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.errMsg)
		})
	}
}

// ============================================================================
// Data Lifecycle Rules Tests
// ============================================================================

func TestDataLifecycleRules_Default(t *testing.T) {
	rules := types.DefaultDataLifecycleRules()

	require.NotNil(t, rules)
	assert.Equal(t, types.DataLifecycleRulesVersion, rules.Version)
	assert.NotEmpty(t, rules.ArtifactPolicies)

	// Validate
	err := rules.Validate()
	require.NoError(t, err)
}

func TestDataLifecycleRules_RawImageNotOnChain(t *testing.T) {
	rules := types.DefaultDataLifecycleRules()

	// Raw images should NEVER be allowed on-chain
	assert.False(t, rules.CanStoreOnChain(types.ArtifactTypeRawImage))
	
	// Raw images must be encrypted
	assert.True(t, rules.RequiresEncryption(types.ArtifactTypeRawImage))
	
	// Raw images should be deleted after verification
	assert.True(t, rules.ShouldDeleteAfterVerification(types.ArtifactTypeRawImage))
}

func TestDataLifecycleRules_FaceEmbeddingHashOnChain(t *testing.T) {
	rules := types.DefaultDataLifecycleRules()

	// Face embedding hashes CAN be stored on-chain
	assert.True(t, rules.CanStoreOnChain(types.ArtifactTypeFaceEmbedding))
	
	// Raw embeddings still need encryption (off-chain)
	assert.True(t, rules.RequiresEncryption(types.ArtifactTypeFaceEmbedding))
	
	// Should NOT be deleted after verification (we keep the hash)
	assert.False(t, rules.ShouldDeleteAfterVerification(types.ArtifactTypeFaceEmbedding))
}

func TestDataLifecycleRules_DocumentHashOnChain(t *testing.T) {
	rules := types.DefaultDataLifecycleRules()

	// Document hashes can be on-chain
	assert.True(t, rules.CanStoreOnChain(types.ArtifactTypeDocumentHash))
	
	// Hashes don't need encryption
	assert.False(t, rules.RequiresEncryption(types.ArtifactTypeDocumentHash))
}

func TestDataLifecycleRules_GetRule(t *testing.T) {
	rules := types.DefaultDataLifecycleRules()

	rule, found := rules.GetRule(types.ArtifactTypeRawImage)
	require.True(t, found)
	assert.False(t, rule.AllowOnChain)

	_, found = rules.GetRule("nonexistent")
	assert.False(t, found)
}

func TestDataLifecycleRules_CreateRetentionPolicy(t *testing.T) {
	rules := types.DefaultDataLifecycleRules()
	now := time.Now()

	// Create policy for raw image (should have short retention)
	policy, err := rules.CreateRetentionPolicy(
		types.ArtifactTypeRawImage,
		"policy-001",
		now,
		100,
	)
	require.NoError(t, err)
	require.NotNil(t, policy)
	assert.True(t, policy.DeleteOnExpiry)

	// Create policy for verification record (should be indefinite)
	policy2, err := rules.CreateRetentionPolicy(
		types.ArtifactTypeVerificationRecord,
		"policy-002",
		now,
		100,
	)
	require.NoError(t, err)
	require.NotNil(t, policy2)
	assert.Equal(t, types.RetentionTypeIndefinite, policy2.RetentionType)
}

func TestArtifactType_Validation(t *testing.T) {
	validTypes := types.AllArtifactTypes()
	
	for _, at := range validTypes {
		assert.True(t, types.IsValidArtifactType(at))
	}
	
	assert.False(t, types.IsValidArtifactType("invalid"))
	assert.False(t, types.IsValidArtifactType(""))
}

func TestRetentionType_Validation(t *testing.T) {
	validTypes := types.AllRetentionTypes()
	
	for _, rt := range validTypes {
		assert.True(t, types.IsValidRetentionType(rt))
	}
	
	assert.False(t, types.IsValidRetentionType("invalid"))
}
