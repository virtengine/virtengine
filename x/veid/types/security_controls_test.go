package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Security Controls Tests (VE-225: Security Controls)
// ============================================================================

func TestTokenTypes(t *testing.T) {
	// Test all types are valid
	for _, tokenType := range types.AllTokenTypes() {
		assert.True(t, types.IsValidTokenType(tokenType), "AllTokenTypes returned invalid type: %s", tokenType)
	}

	// Test invalid type
	assert.False(t, types.IsValidTokenType("invalid"), "IsValidTokenType should return false for invalid type")
}

func TestTokenMapping_Creation(t *testing.T) {
	now := time.Now()

	mapping := types.NewTokenMapping(
		"token-abc123",
		types.TokenTypeIdentityRef,
		"internal-ref-xyz",
		"cosmos1abc...",
		now,
		3600, // 1 hour TTL
	)

	require.NotNil(t, mapping)
	assert.Equal(t, "token-abc123", mapping.Token)
	assert.Equal(t, types.TokenTypeIdentityRef, mapping.TokenType)
	assert.Equal(t, "internal-ref-xyz", mapping.InternalReference)
	assert.Equal(t, "cosmos1abc...", mapping.AccountAddress)
	assert.True(t, mapping.CreatedAt.Equal(now))
	assert.True(t, mapping.ExpiresAt.After(now))
	assert.False(t, mapping.IsRevoked)
}

func TestTokenMapping_IsValid(t *testing.T) {
	now := time.Now()

	// Valid mapping
	validMapping := types.NewTokenMapping(
		"token-123",
		types.TokenTypeIdentityRef,
		"internal-ref",
		"cosmos1abc...",
		now,
		3600, // 1 hour TTL
	)
	assert.True(t, validMapping.IsValid(now), "Mapping should be valid")

	// Expired mapping
	expiredMapping := types.NewTokenMapping(
		"token-456",
		types.TokenTypeIdentityRef,
		"internal-ref",
		"cosmos1abc...",
		now.Add(-2*time.Hour), // Created 2 hours ago
		3600,                  // 1 hour TTL
	)
	assert.False(t, expiredMapping.IsValid(now), "Expired mapping should not be valid")

	// Revoked mapping
	revokedMapping := types.NewTokenMapping(
		"token-789",
		types.TokenTypeIdentityRef,
		"internal-ref",
		"cosmos1abc...",
		now,
		3600,
	)
	revokedMapping.Revoke(now)
	assert.False(t, revokedMapping.IsValid(now), "Revoked mapping should not be valid")
}

func TestTokenMapping_RecordUsage(t *testing.T) {
	now := time.Now()

	mapping := types.NewTokenMapping(
		"token-123",
		types.TokenTypeIdentityRef,
		"internal-ref",
		"cosmos1abc...",
		now,
		3600,
	)

	assert.Equal(t, uint64(0), mapping.UsageCount)
	assert.Nil(t, mapping.LastUsedAt)

	// Record usage
	usageTime := now.Add(time.Hour)
	mapping.RecordUsage(usageTime)

	assert.Equal(t, uint64(1), mapping.UsageCount)
	require.NotNil(t, mapping.LastUsedAt)
	assert.True(t, mapping.LastUsedAt.Equal(usageTime))

	// Record more usage
	mapping.RecordUsage(usageTime.Add(time.Hour))
	assert.Equal(t, uint64(2), mapping.UsageCount)
}

func TestTokenMapping_Revoke(t *testing.T) {
	now := time.Now()

	mapping := types.NewTokenMapping(
		"token-123",
		types.TokenTypeIdentityRef,
		"internal-ref",
		"cosmos1abc...",
		now,
		3600,
	)

	assert.False(t, mapping.IsRevoked)
	assert.Nil(t, mapping.RevokedAt)

	// Revoke the mapping
	revokeTime := now.Add(time.Hour)
	mapping.Revoke(revokeTime)

	assert.True(t, mapping.IsRevoked)
	require.NotNil(t, mapping.RevokedAt)
	assert.True(t, mapping.RevokedAt.Equal(revokeTime))
}

func TestGenerateToken(t *testing.T) {
	now := time.Now()

	token1 := types.GenerateToken("ref-123", "salt1", now)
	token2 := types.GenerateToken("ref-123", "salt1", now)
	token3 := types.GenerateToken("ref-456", "salt1", now)
	token4 := types.GenerateToken("ref-123", "salt2", now)

	// Should be 64 hex characters (SHA256)
	assert.Len(t, token1, 64)

	// Same inputs should produce same token
	assert.Equal(t, token1, token2)

	// Different inputs should produce different tokens
	assert.NotEqual(t, token1, token3)
	assert.NotEqual(t, token1, token4)
}

func TestPseudonymTypes(t *testing.T) {
	// Test all types are valid
	allTypes := types.AllPseudonymTypes()
	assert.NotEmpty(t, allTypes)

	// Check expected types exist
	assert.Contains(t, allTypes, types.PseudonymTypeAccount)
	assert.Contains(t, allTypes, types.PseudonymTypeValidator)
	assert.Contains(t, allTypes, types.PseudonymTypeProvider)
	assert.Contains(t, allTypes, types.PseudonymTypeSession)
}

func TestPseudonymize(t *testing.T) {
	tests := []struct {
		name           string
		identifier     string
		identType      types.PseudonymType
		salt           string
		preservePrefix int
	}{
		{
			name:           "account with prefix",
			identifier:     "cosmos1abc123def456",
			identType:      types.PseudonymTypeAccount,
			salt:           "test-salt",
			preservePrefix: 4,
		},
		{
			name:           "validator no prefix",
			identifier:     "cosmosvaloper1xyz",
			identType:      types.PseudonymTypeValidator,
			salt:           "test-salt",
			preservePrefix: 0,
		},
		{
			name:           "session with prefix",
			identifier:     "session-abc123",
			identType:      types.PseudonymTypeSession,
			salt:           "different-salt",
			preservePrefix: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pseudonym := types.Pseudonymize(tt.identifier, tt.identType, tt.salt, tt.preservePrefix)

			require.NotNil(t, pseudonym)
			assert.Equal(t, tt.identType, pseudonym.Type)
			assert.NotEmpty(t, pseudonym.Value)

			// Should not contain the full original identifier
			assert.NotEqual(t, tt.identifier, pseudonym.Value)

			// Should be deterministic
			pseudonym2 := types.Pseudonymize(tt.identifier, tt.identType, tt.salt, tt.preservePrefix)
			assert.Equal(t, pseudonym.Value, pseudonym2.Value)

			// If prefix is preserved, check it
			if tt.preservePrefix > 0 && len(tt.identifier) >= tt.preservePrefix {
				expectedPrefix := tt.identifier[:tt.preservePrefix] + "_"
				assert.True(t, len(pseudonym.Value) > len(expectedPrefix))
			}
		})
	}
}

func TestDefaultPseudonymizationConfig(t *testing.T) {
	config := types.DefaultPseudonymizationConfig()

	require.NotNil(t, config)
	assert.True(t, config.Enabled)
	assert.Greater(t, config.RotationIntervalSeconds, int64(0))
	assert.Greater(t, config.PreservePrefix, 0)
}

func TestRetentionRule_Creation(t *testing.T) {
	now := time.Now()

	rule := types.NewRetentionRule(
		"rule-123",
		types.ArtifactTypeRawImage,
		types.RetentionTypeDuration,
		90*24*3600, // 90 days in seconds
		true,       // auto delete
		now,
	)

	require.NotNil(t, rule)
	assert.Equal(t, types.RetentionRuleVersion, rule.Version)
	assert.Equal(t, "rule-123", rule.RuleID)
	assert.Equal(t, types.ArtifactTypeRawImage, rule.ArtifactType)
	assert.Equal(t, types.RetentionTypeDuration, rule.RetentionType)
	assert.Equal(t, int64(90*24*3600), rule.RetentionDurationSeconds)
	assert.True(t, rule.AutoDelete)
	assert.True(t, rule.IsEnabled)
}

func TestRetentionRule_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		rule    *types.RetentionRule
		wantErr bool
	}{
		{
			name: "valid duration rule",
			rule: types.NewRetentionRule(
				"rule-1",
				types.ArtifactTypeRawImage,
				types.RetentionTypeDuration,
				7*24*3600, // 7 days
				true,
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty rule ID",
			rule: &types.RetentionRule{
				Version:                  types.RetentionRuleVersion,
				RuleID:                   "",
				ArtifactType:             types.ArtifactTypeRawImage,
				RetentionType:            types.RetentionTypeDuration,
				RetentionDurationSeconds: 7 * 24 * 3600,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero duration for duration type",
			rule: &types.RetentionRule{
				Version:                  types.RetentionRuleVersion,
				RuleID:                   "rule-1",
				ArtifactType:             types.ArtifactTypeRawImage,
				RetentionType:            types.RetentionTypeDuration,
				RetentionDurationSeconds: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultRetentionRules(t *testing.T) {
	now := time.Now()
	rules := types.DefaultRetentionRules(now)

	require.NotEmpty(t, rules, "DefaultRetentionRules should return at least one rule")

	// Validate all default rules
	for _, rule := range rules {
		err := rule.Validate()
		require.NoError(t, err, "Default rule %s should be valid", rule.RuleID)
	}

	// Check specific rules exist
	ruleMap := make(map[types.ArtifactType]*types.RetentionRule)
	for i := range rules {
		ruleMap[rules[i].ArtifactType] = &rules[i]
	}

	assert.Contains(t, ruleMap, types.ArtifactTypeRawImage, "Default rules should include raw image rule")
	assert.Contains(t, ruleMap, types.ArtifactTypeFaceEmbedding, "Default rules should include face embedding rule")
}

func TestRetentionEnforcementResult(t *testing.T) {
	now := time.Now()

	result := types.NewRetentionEnforcementResult("enforcement-1", now, 12345)

	require.NotNil(t, result)
	assert.Equal(t, "enforcement-1", result.EnforcementID)
	assert.True(t, result.RunAt.Equal(now))
	assert.Equal(t, int64(12345), result.BlockHeight)
	assert.NotNil(t, result.ByType)

	// Test tracking counts
	result.ArtifactsScanned = 100
	result.ArtifactsExpired = 50
	result.ArtifactsDeleted = 45
	result.ArtifactsFailed = 5

	assert.Equal(t, uint64(100), result.ArtifactsScanned)
	assert.Equal(t, uint64(50), result.ArtifactsExpired)
	assert.Equal(t, uint64(45), result.ArtifactsDeleted)
	assert.Equal(t, uint64(5), result.ArtifactsFailed)
}

func TestSecurityAuditEvent_Creation(t *testing.T) {
	now := time.Now()

	event := types.NewSecurityAuditEvent(
		"event-1",
		"token_created",
		"cosmos1abc...",
		"create",
		"success",
		"test-salt",
		now,
		12345,
	)

	require.NotNil(t, event)
	assert.Equal(t, "event-1", event.EventID)
	assert.Equal(t, "token_created", event.EventType)
	assert.NotEmpty(t, event.AccountPseudonym) // Should be pseudonymized
	assert.NotEqual(t, "cosmos1abc...", event.AccountPseudonym)
	assert.Equal(t, "create", event.Action)
	assert.Equal(t, "success", event.Outcome)
	assert.True(t, event.Timestamp.Equal(now))
	assert.Equal(t, int64(12345), event.BlockHeight)
}

func TestSecurityAuditEvent_AddMetadata(t *testing.T) {
	now := time.Now()

	event := types.NewSecurityAuditEvent(
		"event-1",
		"token_created",
		"cosmos1abc...",
		"create",
		"success",
		"test-salt",
		now,
		12345,
	)

	assert.Empty(t, event.Metadata["key1"])

	event.AddMetadata("key1", "value1")
	event.AddMetadata("key2", "value2")

	assert.Equal(t, "value1", event.Metadata["key1"])
	assert.Equal(t, "value2", event.Metadata["key2"])
}
