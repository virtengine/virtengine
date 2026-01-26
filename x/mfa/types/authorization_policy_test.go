package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pkg.akt.dev/node/x/mfa/types"
)

// ============================================================================
// Authorization Policy Tests (VE-221: Authorization Policy)
// ============================================================================

func TestAuthorizationPolicy_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		policy  *types.AuthorizationPolicy
		wantErr bool
	}{
		{
			name: "valid threshold policy",
			policy: types.NewThresholdAuthorizationPolicy(
				"policy-1",
				"cosmos1abc...",
				1000000,
				"uve",
				types.AuthReqMFA,
				now,
			),
			wantErr: false,
		},
		{
			name: "valid multi-trigger policy",
			policy: types.NewAuthorizationPolicy(
				"policy-2",
				"cosmos1abc...",
				[]types.AuthorizationPolicyTrigger{
					{
						TriggerType: types.PolicyTriggerThreshold,
						Threshold: &types.ThresholdConfig{
							Amount:         5000000,
							Denom:          "uve",
							PerTransaction: true,
						},
					},
					{
						TriggerType: types.PolicyTriggerCategory,
						Categories:  []types.SensitiveTransactionType{types.SensitiveTxLargeWithdrawal},
					},
				},
				types.AuthReqBiometricAndMFA,
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty policy ID",
			policy: &types.AuthorizationPolicy{
				Version:        types.AuthorizationPolicyVersion,
				PolicyID:       "",
				AccountAddress: "cosmos1abc...",
				Triggers: []types.AuthorizationPolicyTrigger{
					{
						TriggerType: types.PolicyTriggerCategory,
						Categories:  []types.SensitiveTransactionType{types.SensitiveTxLargeWithdrawal},
					},
				},
				Requirements: types.AuthReqMFA,
				CreatedAt:    now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty account address",
			policy: &types.AuthorizationPolicy{
				Version:        types.AuthorizationPolicyVersion,
				PolicyID:       "policy-1",
				AccountAddress: "",
				Triggers: []types.AuthorizationPolicyTrigger{
					{
						TriggerType: types.PolicyTriggerCategory,
						Categories:  []types.SensitiveTransactionType{types.SensitiveTxLargeWithdrawal},
					},
				},
				Requirements: types.AuthReqMFA,
				CreatedAt:    now,
			},
			wantErr: true,
		},
		{
			name: "invalid - no triggers",
			policy: &types.AuthorizationPolicy{
				Version:        types.AuthorizationPolicyVersion,
				PolicyID:       "policy-1",
				AccountAddress: "cosmos1abc...",
				Triggers:       []types.AuthorizationPolicyTrigger{},
				Requirements:   types.AuthReqMFA,
				CreatedAt:      now,
			},
			wantErr: true,
		},
		{
			name: "invalid - threshold with no config",
			policy: &types.AuthorizationPolicy{
				Version:        types.AuthorizationPolicyVersion,
				PolicyID:       "policy-1",
				AccountAddress: "cosmos1abc...",
				Triggers: []types.AuthorizationPolicyTrigger{
					{
						TriggerType: types.PolicyTriggerThreshold,
						Threshold:   nil,
					},
				},
				Requirements: types.AuthReqMFA,
				CreatedAt:    now,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero threshold amount",
			policy: &types.AuthorizationPolicy{
				Version:        types.AuthorizationPolicyVersion,
				PolicyID:       "policy-1",
				AccountAddress: "cosmos1abc...",
				Triggers: []types.AuthorizationPolicyTrigger{
					{
						TriggerType: types.PolicyTriggerThreshold,
						Threshold: &types.ThresholdConfig{
							Amount:         0,
							Denom:          "uve",
							PerTransaction: true,
						},
					},
				},
				Requirements: types.AuthReqMFA,
				CreatedAt:    now,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.policy.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestThresholdConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  types.ThresholdConfig
		wantErr bool
	}{
		{
			name: "valid per-transaction threshold",
			config: types.ThresholdConfig{
				Amount:         1000000,
				Denom:          "uve",
				PerTransaction: true,
			},
			wantErr: false,
		},
		{
			name: "valid rolling window threshold",
			config: types.ThresholdConfig{
				Amount:                5000000,
				Denom:                 "uve",
				PerTransaction:        false,
				WindowDurationSeconds: 3600,
			},
			wantErr: false,
		},
		{
			name: "invalid - zero amount",
			config: types.ThresholdConfig{
				Amount:         0,
				Denom:          "uve",
				PerTransaction: true,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty denom",
			config: types.ThresholdConfig{
				Amount:         1000000,
				Denom:          "",
				PerTransaction: true,
			},
			wantErr: true,
		},
		{
			name: "invalid - rolling window without duration",
			config: types.ThresholdConfig{
				Amount:                1000000,
				Denom:                 "uve",
				PerTransaction:        false,
				WindowDurationSeconds: 0,
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

func TestPolicyTriggerTypes(t *testing.T) {
	// Test all trigger types are valid
	for _, triggerType := range types.AllPolicyTriggerTypes() {
		assert.True(t, types.IsValidPolicyTriggerType(triggerType), "AllPolicyTriggerTypes returned invalid type: %s", triggerType)
	}

	// Test invalid trigger type
	assert.False(t, types.IsValidPolicyTriggerType("invalid"), "IsValidPolicyTriggerType should return false for invalid type")
}

func TestAuthorizationRequirements(t *testing.T) {
	// Test all requirements are valid
	for _, req := range types.AllAuthorizationRequirements() {
		assert.True(t, types.IsValidAuthorizationRequirement(req), "AllAuthorizationRequirements returned invalid requirement: %s", req)
	}

	// Test invalid requirement
	assert.False(t, types.IsValidAuthorizationRequirement("invalid"), "IsValidAuthorizationRequirement should return false for invalid requirement")
}

func TestAuthorizationPolicy_HasThresholdTrigger(t *testing.T) {
	now := time.Now()

	thresholdPolicy := types.NewThresholdAuthorizationPolicy(
		"policy-1",
		"cosmos1abc...",
		1000000,
		"uve",
		types.AuthReqMFA,
		now,
	)

	assert.True(t, thresholdPolicy.HasThresholdTrigger(), "HasThresholdTrigger should return true for threshold policy")

	categoryPolicy := types.NewAuthorizationPolicy(
		"policy-2",
		"cosmos1abc...",
		[]types.AuthorizationPolicyTrigger{
			{
				TriggerType: types.PolicyTriggerCategory,
				Categories:  []types.SensitiveTransactionType{types.SensitiveTxLargeWithdrawal},
			},
		},
		types.AuthReqMFA,
		now,
	)

	assert.False(t, categoryPolicy.HasThresholdTrigger(), "HasThresholdTrigger should return false for category-only policy")
}

func TestAuthorizationResult(t *testing.T) {
	result := types.NewAuthorizationResult(false)
	result.RequiredAction = types.AuthReqBiometric
	result.AddTriggeredPolicy("policy-1", "threshold exceeded")
	result.AddTriggeredPolicy("policy-2", "category match")

	assert.False(t, result.Authorized, "Result should not be authorized")
	assert.Len(t, result.TriggeredPolicies, 2, "Expected 2 triggered policies")
	assert.Len(t, result.TriggerReasons, 2, "Expected 2 trigger reasons")
}

func TestAuthorizationAuditEvent(t *testing.T) {
	now := time.Now()

	event := types.NewAuthorizationAuditEvent(
		"event-1",
		"cosmos1abc...",
		types.SensitiveTxHighValueOrder,
		12345,
		now,
	)

	assert.Equal(t, "event-1", event.EventID)
	assert.Equal(t, types.SensitiveTxHighValueOrder, event.TransactionType)

	// Test validation
	err := event.Validate()
	require.NoError(t, err, "Valid event should not return error")

	// Test invalid event
	invalidEvent := &types.AuthorizationAuditEvent{
		EventID:        "",
		AccountAddress: "cosmos1abc...",
		Timestamp:      now,
	}
	err = invalidEvent.Validate()
	require.Error(t, err, "Invalid event should return error")
}

func TestFrequencyConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  types.FrequencyConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: types.FrequencyConfig{
				MaxTransactions:       10,
				WindowDurationSeconds: 3600,
			},
			wantErr: false,
		},
		{
			name: "invalid - zero max transactions",
			config: types.FrequencyConfig{
				MaxTransactions:       0,
				WindowDurationSeconds: 3600,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero window duration",
			config: types.FrequencyConfig{
				MaxTransactions:       10,
				WindowDurationSeconds: 0,
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

func TestTimeWindowConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  types.TimeWindowConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: types.TimeWindowConfig{
				StartHourUTC: 22,
				EndHourUTC:   6,
				DaysOfWeek:   []uint8{0, 6}, // Weekend
			},
			wantErr: false,
		},
		{
			name: "invalid - start hour > 23",
			config: types.TimeWindowConfig{
				StartHourUTC: 24,
				EndHourUTC:   6,
			},
			wantErr: true,
		},
		{
			name: "invalid - end hour > 23",
			config: types.TimeWindowConfig{
				StartHourUTC: 22,
				EndHourUTC:   24,
			},
			wantErr: true,
		},
		{
			name: "invalid - day > 6",
			config: types.TimeWindowConfig{
				StartHourUTC: 22,
				EndHourUTC:   6,
				DaysOfWeek:   []uint8{7},
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
