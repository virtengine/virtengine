package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Verification Status Tests (VE-1006: Test Coverage)
// ============================================================================

func TestIsValidVerificationStatus(t *testing.T) {
	tests := []struct {
		status   types.VerificationStatus
		expected bool
	}{
		{types.VerificationStatusUnknown, true},
		{types.VerificationStatusPending, true},
		{types.VerificationStatusInProgress, true},
		{types.VerificationStatusVerified, true},
		{types.VerificationStatusRejected, true},
		{types.VerificationStatusExpired, true},
		{types.VerificationStatusNeedsAdditionalFactor, true},
		{types.VerificationStatusAdditionalFactorPending, true},
		{types.VerificationStatus("invalid"), false},
		{types.VerificationStatus(""), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			assert.Equal(t, tc.expected, types.IsValidVerificationStatus(tc.status))
		})
	}
}

func TestAllVerificationStatuses(t *testing.T) {
	statuses := types.AllVerificationStatuses()
	assert.Len(t, statuses, 8)
	assert.Contains(t, statuses, types.VerificationStatusUnknown)
	assert.Contains(t, statuses, types.VerificationStatusAdditionalFactorPending)
}

func TestIsFinalStatus(t *testing.T) {
	tests := []struct {
		status   types.VerificationStatus
		expected bool
	}{
		{types.VerificationStatusVerified, true},
		{types.VerificationStatusRejected, true},
		{types.VerificationStatusExpired, true},
		{types.VerificationStatusUnknown, false},
		{types.VerificationStatusPending, false},
		{types.VerificationStatusInProgress, false},
		{types.VerificationStatusNeedsAdditionalFactor, false},
		{types.VerificationStatusAdditionalFactorPending, false},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			assert.Equal(t, tc.expected, types.IsFinalStatus(tc.status))
		})
	}
}

func TestVerificationStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     types.VerificationStatus
		to       types.VerificationStatus
		expected bool
	}{
		// From Unknown
		{"unknown to pending", types.VerificationStatusUnknown, types.VerificationStatusPending, true},
		{"unknown to verified", types.VerificationStatusUnknown, types.VerificationStatusVerified, false},

		// From Pending
		{"pending to in_progress", types.VerificationStatusPending, types.VerificationStatusInProgress, true},
		{"pending to rejected", types.VerificationStatusPending, types.VerificationStatusRejected, true},
		{"pending to expired", types.VerificationStatusPending, types.VerificationStatusExpired, true},
		{"pending to verified", types.VerificationStatusPending, types.VerificationStatusVerified, false},

		// From InProgress
		{"in_progress to verified", types.VerificationStatusInProgress, types.VerificationStatusVerified, true},
		{"in_progress to rejected", types.VerificationStatusInProgress, types.VerificationStatusRejected, true},
		{"in_progress to pending", types.VerificationStatusInProgress, types.VerificationStatusPending, true},
		{"in_progress to needs_additional_factor", types.VerificationStatusInProgress, types.VerificationStatusNeedsAdditionalFactor, true},
		{"in_progress to expired", types.VerificationStatusInProgress, types.VerificationStatusExpired, false},

		// From Verified
		{"verified to expired", types.VerificationStatusVerified, types.VerificationStatusExpired, true},
		{"verified to pending", types.VerificationStatusVerified, types.VerificationStatusPending, false},
		{"verified to rejected", types.VerificationStatusVerified, types.VerificationStatusRejected, false},

		// From Rejected (can re-verify)
		{"rejected to pending", types.VerificationStatusRejected, types.VerificationStatusPending, true},
		{"rejected to verified", types.VerificationStatusRejected, types.VerificationStatusVerified, false},

		// From Expired (can re-verify)
		{"expired to pending", types.VerificationStatusExpired, types.VerificationStatusPending, true},
		{"expired to verified", types.VerificationStatusExpired, types.VerificationStatusVerified, false},

		// From NeedsAdditionalFactor
		{"needs_additional_factor to additional_factor_pending", types.VerificationStatusNeedsAdditionalFactor, types.VerificationStatusAdditionalFactorPending, true},
		{"needs_additional_factor to rejected", types.VerificationStatusNeedsAdditionalFactor, types.VerificationStatusRejected, true},
		{"needs_additional_factor to expired", types.VerificationStatusNeedsAdditionalFactor, types.VerificationStatusExpired, true},
		{"needs_additional_factor to verified", types.VerificationStatusNeedsAdditionalFactor, types.VerificationStatusVerified, false},

		// From AdditionalFactorPending
		{"additional_factor_pending to verified", types.VerificationStatusAdditionalFactorPending, types.VerificationStatusVerified, true},
		{"additional_factor_pending to rejected", types.VerificationStatusAdditionalFactorPending, types.VerificationStatusRejected, true},
		{"additional_factor_pending to expired", types.VerificationStatusAdditionalFactorPending, types.VerificationStatusExpired, true},
		{"additional_factor_pending to needs_additional_factor", types.VerificationStatusAdditionalFactorPending, types.VerificationStatusNeedsAdditionalFactor, true},
		{"additional_factor_pending to pending", types.VerificationStatusAdditionalFactorPending, types.VerificationStatusPending, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.from.CanTransitionTo(tc.to))
		})
	}
}

// ============================================================================
// Verification Event Tests
// ============================================================================

func TestNewVerificationEvent(t *testing.T) {
	now := time.Now()
	event := types.NewVerificationEvent(
		"event-001",
		"scope-001",
		types.VerificationStatusPending,
		types.VerificationStatusInProgress,
		now,
		"Processing started",
	)

	require.NotNil(t, event)
	assert.Equal(t, "event-001", event.EventID)
	assert.Equal(t, "scope-001", event.ScopeID)
	assert.Equal(t, types.VerificationStatusPending, event.PreviousStatus)
	assert.Equal(t, types.VerificationStatusInProgress, event.NewStatus)
	assert.Equal(t, now, event.Timestamp)
	assert.Equal(t, "Processing started", event.Reason)
	assert.NotNil(t, event.Metadata)
	assert.Empty(t, event.Metadata)
}

func TestVerificationEvent_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		event       *types.VerificationEvent
		expectError bool
		errContains string
	}{
		{
			name: "valid event",
			event: &types.VerificationEvent{
				EventID:        "event-001",
				PreviousStatus: types.VerificationStatusPending,
				NewStatus:      types.VerificationStatusInProgress,
				Timestamp:      now,
			},
			expectError: false,
		},
		{
			name: "valid event without scope (account-level)",
			event: &types.VerificationEvent{
				EventID:        "event-001",
				ScopeID:        "", // Empty is valid for account-level events
				PreviousStatus: types.VerificationStatusPending,
				NewStatus:      types.VerificationStatusVerified,
				Timestamp:      now,
			},
			expectError: false,
		},
		{
			name: "empty event ID",
			event: &types.VerificationEvent{
				EventID:        "",
				PreviousStatus: types.VerificationStatusPending,
				NewStatus:      types.VerificationStatusInProgress,
				Timestamp:      now,
			},
			expectError: true,
			errContains: "event_id cannot be empty",
		},
		{
			name: "invalid previous status",
			event: &types.VerificationEvent{
				EventID:        "event-001",
				PreviousStatus: types.VerificationStatus("invalid"),
				NewStatus:      types.VerificationStatusInProgress,
				Timestamp:      now,
			},
			expectError: true,
			errContains: "invalid previous status",
		},
		{
			name: "invalid new status",
			event: &types.VerificationEvent{
				EventID:        "event-001",
				PreviousStatus: types.VerificationStatusPending,
				NewStatus:      types.VerificationStatus("invalid"),
				Timestamp:      now,
			},
			expectError: true,
			errContains: "invalid new status",
		},
		{
			name: "zero timestamp",
			event: &types.VerificationEvent{
				EventID:        "event-001",
				PreviousStatus: types.VerificationStatusPending,
				NewStatus:      types.VerificationStatusInProgress,
				Timestamp:      time.Time{},
			},
			expectError: true,
			errContains: "timestamp cannot be zero",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.event.Validate()
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// Simple Verification Result Tests
// ============================================================================

func TestSimpleVerificationResult(t *testing.T) {
	now := time.Now()

	result := types.SimpleVerificationResult{
		Success:            true,
		Status:             types.VerificationStatusVerified,
		Score:              85,
		ScoreVersion:       "v1.0.0",
		Confidence:         95,
		Details:            "All checks passed",
		Flags:              []string{},
		ProcessedAt:        now,
		ValidatorConsensus: 100,
	}

	assert.True(t, result.Success)
	assert.Equal(t, types.VerificationStatusVerified, result.Status)
	assert.Equal(t, uint32(85), result.Score)
	assert.Equal(t, "v1.0.0", result.ScoreVersion)
	assert.Equal(t, uint32(95), result.Confidence)
	assert.Equal(t, "All checks passed", result.Details)
	assert.Empty(t, result.Flags)
	assert.Equal(t, now, result.ProcessedAt)
	assert.Equal(t, uint32(100), result.ValidatorConsensus)
}

func TestSimpleVerificationResult_WithFlags(t *testing.T) {
	result := types.SimpleVerificationResult{
		Success: false,
		Status:  types.VerificationStatusRejected,
		Score:   30,
		Flags:   []string{"low_confidence", "blur_detected", "possible_spoof"},
	}

	assert.False(t, result.Success)
	assert.Equal(t, types.VerificationStatusRejected, result.Status)
	assert.Len(t, result.Flags, 3)
	assert.Contains(t, result.Flags, "low_confidence")
	assert.Contains(t, result.Flags, "blur_detected")
	assert.Contains(t, result.Flags, "possible_spoof")
}
