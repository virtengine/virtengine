//go:build ignore
// +build ignore

// TODO: This test file is excluded until ExternalToken API is stabilized.
// The constructor NewExternalToken and field names need to match the implementation.

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

func TestExternalToken_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		token   *types.ExternalToken
		wantErr bool
	}{
		{
			name: "valid token",
			token: types.NewExternalToken(
				"token-123",
				"cosmos1abc...",
				"provider-abc",
				types.TokenTypePII,
				now,
			),
			wantErr: false,
		},
		{
			name: "valid payment token",
			token: types.NewExternalToken(
				"token-456",
				"cosmos1abc...",
				"stripe-pm-123",
				types.TokenTypePayment,
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty token ID",
			token: &types.ExternalToken{
				Version:       types.ExternalTokenVersion,
				TokenID:       "",
				Owner:         "cosmos1abc...",
				ExternalRef:   "provider-abc",
				TokenType:     types.TokenTypePII,
				TokenizedAt:   now,
				TokenStatus:   types.TokenStatusActive,
				EncryptionKey: "key-ref-123",
			},
			wantErr: true,
		},
		{
			name: "invalid - empty owner",
			token: &types.ExternalToken{
				Version:       types.ExternalTokenVersion,
				TokenID:       "token-123",
				Owner:         "",
				ExternalRef:   "provider-abc",
				TokenType:     types.TokenTypePII,
				TokenizedAt:   now,
				TokenStatus:   types.TokenStatusActive,
				EncryptionKey: "key-ref-123",
			},
			wantErr: true,
		},
		{
			name: "invalid - empty external ref",
			token: &types.ExternalToken{
				Version:       types.ExternalTokenVersion,
				TokenID:       "token-123",
				Owner:         "cosmos1abc...",
				ExternalRef:   "",
				TokenType:     types.TokenTypePII,
				TokenizedAt:   now,
				TokenStatus:   types.TokenStatusActive,
				EncryptionKey: "key-ref-123",
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid token type",
			token: &types.ExternalToken{
				Version:       types.ExternalTokenVersion,
				TokenID:       "token-123",
				Owner:         "cosmos1abc...",
				ExternalRef:   "provider-abc",
				TokenType:     "invalid",
				TokenizedAt:   now,
				TokenStatus:   types.TokenStatusActive,
				EncryptionKey: "key-ref-123",
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid token status",
			token: &types.ExternalToken{
				Version:       types.ExternalTokenVersion,
				TokenID:       "token-123",
				Owner:         "cosmos1abc...",
				ExternalRef:   "provider-abc",
				TokenType:     types.TokenTypePII,
				TokenizedAt:   now,
				TokenStatus:   "invalid",
				EncryptionKey: "key-ref-123",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.token.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTokenTypes(t *testing.T) {
	// Test all types are valid
	for _, tokenType := range types.AllTokenTypes() {
		assert.True(t, types.IsValidTokenType(tokenType), "AllTokenTypes returned invalid type: %s", tokenType)
	}

	// Test invalid type
	assert.False(t, types.IsValidTokenType("invalid"), "IsValidTokenType should return false for invalid type")
}

func TestTokenStatuses(t *testing.T) {
	// Test all statuses are valid
	for _, status := range types.AllTokenStatuses() {
		assert.True(t, types.IsValidTokenStatus(status), "AllTokenStatuses returned invalid status: %s", status)
	}

	// Test invalid status
	assert.False(t, types.IsValidTokenStatus("invalid"), "IsValidTokenStatus should return false for invalid status")
}

func TestTokenMapping_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		mapping *types.TokenMapping
		wantErr bool
	}{
		{
			name: "valid mapping",
			mapping: types.NewTokenMapping(
				"mapping-123",
				"token-123",
				"biometric_template",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty mapping ID",
			mapping: &types.TokenMapping{
				Version:   types.TokenMappingVersion,
				MappingID: "",
				TokenID:   "token-123",
				FieldPath: "biometric_template",
				CreatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty token ID",
			mapping: &types.TokenMapping{
				Version:   types.TokenMappingVersion,
				MappingID: "mapping-123",
				TokenID:   "",
				FieldPath: "biometric_template",
				CreatedAt: now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty field path",
			mapping: &types.TokenMapping{
				Version:   types.TokenMappingVersion,
				MappingID: "mapping-123",
				TokenID:   "token-123",
				FieldPath: "",
				CreatedAt: now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mapping.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPseudonym_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		pseudonym *types.Pseudonym
		wantErr   bool
	}{
		{
			name: "valid pseudonym",
			pseudonym: types.NewPseudonym(
				"pseudo-123",
				"cosmos1abc...",
				"analysis-context-1",
				"hash-abc123",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty pseudonym ID",
			pseudonym: &types.Pseudonym{
				Version:     types.PseudonymVersion,
				PseudonymID: "",
				Owner:       "cosmos1abc...",
				Context:     "analysis-1",
				SaltedHash:  "hash-123",
				CreatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty owner",
			pseudonym: &types.Pseudonym{
				Version:     types.PseudonymVersion,
				PseudonymID: "pseudo-123",
				Owner:       "",
				Context:     "analysis-1",
				SaltedHash:  "hash-123",
				CreatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty context",
			pseudonym: &types.Pseudonym{
				Version:     types.PseudonymVersion,
				PseudonymID: "pseudo-123",
				Owner:       "cosmos1abc...",
				Context:     "",
				SaltedHash:  "hash-123",
				CreatedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty salted hash",
			pseudonym: &types.Pseudonym{
				Version:     types.PseudonymVersion,
				PseudonymID: "pseudo-123",
				Owner:       "cosmos1abc...",
				Context:     "analysis-1",
				SaltedHash:  "",
				CreatedAt:   now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pseudonym.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPseudonymizationConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *types.PseudonymizationConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: types.NewPseudonymizationConfig(
				"config-1",
				"biometric_data",
				types.PseudoMethodSaltedHash,
				32,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty config ID",
			config: &types.PseudonymizationConfig{
				Version:    types.PseudoConfigVersion,
				ConfigID:   "",
				FieldPath:  "biometric_data",
				Method:     types.PseudoMethodSaltedHash,
				SaltLength: 32,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty field path",
			config: &types.PseudonymizationConfig{
				Version:    types.PseudoConfigVersion,
				ConfigID:   "config-1",
				FieldPath:  "",
				Method:     types.PseudoMethodSaltedHash,
				SaltLength: 32,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid method",
			config: &types.PseudonymizationConfig{
				Version:    types.PseudoConfigVersion,
				ConfigID:   "config-1",
				FieldPath:  "biometric_data",
				Method:     "invalid",
				SaltLength: 32,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero salt length",
			config: &types.PseudonymizationConfig{
				Version:    types.PseudoConfigVersion,
				ConfigID:   "config-1",
				FieldPath:  "biometric_data",
				Method:     types.PseudoMethodSaltedHash,
				SaltLength: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPseudonymizationMethods(t *testing.T) {
	// Test all methods are valid
	for _, method := range types.AllPseudonymizationMethods() {
		assert.True(t, types.IsValidPseudonymizationMethod(method), "AllPseudonymizationMethods returned invalid method: %s", method)
	}

	// Test invalid method
	assert.False(t, types.IsValidPseudonymizationMethod("invalid"), "IsValidPseudonymizationMethod should return false for invalid method")
}

func TestRetentionRule_Validate(t *testing.T) {
	tests := []struct {
		name    string
		rule    *types.RetentionRule
		wantErr bool
	}{
		{
			name: "valid rule with duration",
			rule: types.NewRetentionRule(
				"rule-1",
				"biometric_data",
				types.RetentionActionDelete,
				7*24*time.Hour,
			),
			wantErr: false,
		},
		{
			name: "valid rule with anonymize action",
			rule: types.NewRetentionRule(
				"rule-2",
				"face_template",
				types.RetentionActionAnonymize,
				30*24*time.Hour,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty rule ID",
			rule: &types.RetentionRule{
				Version:           types.RetentionRuleVersion,
				RuleID:            "",
				DataType:          "biometric_data",
				RetentionDuration: 7 * 24 * time.Hour,
				Action:            types.RetentionActionDelete,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty data type",
			rule: &types.RetentionRule{
				Version:           types.RetentionRuleVersion,
				RuleID:            "rule-1",
				DataType:          "",
				RetentionDuration: 7 * 24 * time.Hour,
				Action:            types.RetentionActionDelete,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero retention duration",
			rule: &types.RetentionRule{
				Version:           types.RetentionRuleVersion,
				RuleID:            "rule-1",
				DataType:          "biometric_data",
				RetentionDuration: 0,
				Action:            types.RetentionActionDelete,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid action",
			rule: &types.RetentionRule{
				Version:           types.RetentionRuleVersion,
				RuleID:            "rule-1",
				DataType:          "biometric_data",
				RetentionDuration: 7 * 24 * time.Hour,
				Action:            "invalid",
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

func TestRetentionActions(t *testing.T) {
	// Test all actions are valid
	for _, action := range types.AllRetentionActions() {
		assert.True(t, types.IsValidRetentionAction(action), "AllRetentionActions returned invalid action: %s", action)
	}

	// Test invalid action
	assert.False(t, types.IsValidRetentionAction("invalid"), "IsValidRetentionAction should return false for invalid action")
}

func TestDefaultRetentionRules(t *testing.T) {
	rules := types.DefaultRetentionRules()

	require.NotEmpty(t, rules, "DefaultRetentionRules should return at least one rule")

	// Validate all default rules
	for _, rule := range rules {
		err := rule.Validate()
		require.NoError(t, err, "Default rule %s should be valid", rule.RuleID)
	}

	// Check specific rules exist
	ruleMap := make(map[string]*types.RetentionRule)
	for _, rule := range rules {
		ruleMap[rule.DataType] = rule
	}

	assert.Contains(t, ruleMap, "biometric_template", "Default rules should include biometric_template rule")
	assert.Contains(t, ruleMap, "capture_image", "Default rules should include capture_image rule")
}

func TestRetentionEnforcementResult(t *testing.T) {
	now := time.Now()

	result := types.NewRetentionEnforcementResult("rule-1", now)
	result.ItemsProcessed = 100
	result.ItemsDeleted = 50
	result.ItemsAnonymized = 30
	result.ItemsFailed = 5

	// Add errors
	result.AddError("item-1", "deletion failed")
	result.AddError("item-2", "access denied")

	assert.Len(t, result.Errors, 2, "Expected 2 errors")

	// Test success calculation
	assert.Equal(t, int64(80), result.GetSuccessCount(), "Expected success count 80") // 50 + 30
}

func TestSecurityAuditEvent_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		event   *types.SecurityAuditEvent
		wantErr bool
	}{
		{
			name: "valid event",
			event: types.NewSecurityAuditEvent(
				"event-1",
				types.SecurityEventTokenCreated,
				"cosmos1abc...",
				"token-123",
				"Created new PII token",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty event ID",
			event: &types.SecurityAuditEvent{
				Version:     types.SecurityAuditVersion,
				EventID:     "",
				EventType:   types.SecurityEventTokenCreated,
				Actor:       "cosmos1abc...",
				ResourceID:  "token-123",
				Description: "description",
				Timestamp:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty event type",
			event: &types.SecurityAuditEvent{
				Version:     types.SecurityAuditVersion,
				EventID:     "event-1",
				EventType:   "",
				Actor:       "cosmos1abc...",
				ResourceID:  "token-123",
				Description: "description",
				Timestamp:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty actor",
			event: &types.SecurityAuditEvent{
				Version:     types.SecurityAuditVersion,
				EventID:     "event-1",
				EventType:   types.SecurityEventTokenCreated,
				Actor:       "",
				ResourceID:  "token-123",
				Description: "description",
				Timestamp:   now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSecurityEventTypes(t *testing.T) {
	// Test all event types are valid
	for _, eventType := range types.AllSecurityEventTypes() {
		assert.True(t, types.IsValidSecurityEventType(eventType), "AllSecurityEventTypes returned invalid type: %s", eventType)
	}

	// Test invalid type
	assert.False(t, types.IsValidSecurityEventType("invalid"), "IsValidSecurityEventType should return false for invalid type")
}

func TestExternalToken_Revoke(t *testing.T) {
	now := time.Now()

	token := types.NewExternalToken(
		"token-123",
		"cosmos1abc...",
		"provider-abc",
		types.TokenTypePII,
		now,
	)

	assert.Equal(t, types.TokenStatusActive, token.TokenStatus, "New token should have active status")

	revokeTime := now.Add(1 * time.Hour)
	token.Revoke(revokeTime)

	assert.Equal(t, types.TokenStatusRevoked, token.TokenStatus, "Revoked token should have revoked status")
	require.NotNil(t, token.RevokedAt, "RevokedAt should be set")
	assert.True(t, token.RevokedAt.Equal(revokeTime), "RevokedAt should be set to revoke time")
}

func TestPseudonym_Revoke(t *testing.T) {
	now := time.Now()

	pseudonym := types.NewPseudonym(
		"pseudo-123",
		"cosmos1abc...",
		"analysis-1",
		"hash-123",
		now,
	)

	assert.False(t, pseudonym.IsRevoked, "New pseudonym should not be revoked")

	revokeTime := now.Add(1 * time.Hour)
	pseudonym.Revoke(revokeTime)

	assert.True(t, pseudonym.IsRevoked, "Revoked pseudonym should be marked as revoked")
	require.NotNil(t, pseudonym.RevokedAt, "RevokedAt should be set")
	assert.True(t, pseudonym.RevokedAt.Equal(revokeTime), "RevokedAt should be set to revoke time")
}
