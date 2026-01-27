package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Score Tests (VE-1006: Test Coverage)
// ============================================================================

func TestScoreThresholds(t *testing.T) {
	// Verify threshold constants match veid-flow-spec.md
	assert.Equal(t, uint32(50), types.ThresholdBasic)
	assert.Equal(t, uint32(70), types.ThresholdStandard)
	assert.Equal(t, uint32(85), types.ThresholdPremium)
	assert.Equal(t, uint32(100), types.MaxScore)
}

func TestTierConstants(t *testing.T) {
	assert.Equal(t, 0, types.TierUnverified)
	assert.Equal(t, 1, types.TierBasic)
	assert.Equal(t, 2, types.TierStandard)
	assert.Equal(t, 3, types.TierPremium)
}

// ============================================================================
// Account Status Tests
// ============================================================================

func TestIsValidAccountStatus(t *testing.T) {
	tests := []struct {
		status   types.AccountStatus
		expected bool
	}{
		{types.AccountStatusUnknown, true},
		{types.AccountStatusPending, true},
		{types.AccountStatusInProgress, true},
		{types.AccountStatusVerified, true},
		{types.AccountStatusRejected, true},
		{types.AccountStatusExpired, true},
		{types.AccountStatusNeedsAdditionalFactor, true},
		{types.AccountStatus("invalid"), false},
		{types.AccountStatus(""), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.status), func(t *testing.T) {
			assert.Equal(t, tc.expected, types.IsValidAccountStatus(tc.status))
		})
	}
}

func TestAllAccountStatuses(t *testing.T) {
	statuses := types.AllAccountStatuses()
	assert.Len(t, statuses, 7)
	assert.Contains(t, statuses, types.AccountStatusUnknown)
	assert.Contains(t, statuses, types.AccountStatusVerified)
	assert.Contains(t, statuses, types.AccountStatusNeedsAdditionalFactor)
}

// ============================================================================
// Identity Score Tests
// ============================================================================

func TestNewIdentityScore(t *testing.T) {
	now := time.Now()
	verificationInputs := []byte("test-verification-inputs")

	score := types.NewIdentityScore(
		"ve1account123",
		75,
		types.AccountStatusVerified,
		"v1.0.0",
		now,
		12345,
		verificationInputs,
	)

	require.NotNil(t, score)
	assert.Equal(t, "ve1account123", score.AccountAddress)
	assert.Equal(t, uint32(75), score.Score)
	assert.Equal(t, types.AccountStatusVerified, score.Status)
	assert.Equal(t, "v1.0.0", score.ModelVersion)
	assert.Equal(t, now, score.ComputedAt)
	assert.Equal(t, int64(12345), score.BlockHeight)
	assert.NotEmpty(t, score.VerificationHash)
	assert.Len(t, score.VerificationHash, 32) // SHA-256
}

func TestIdentityScore_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		score       *types.IdentityScore
		expectError bool
		errContains string
	}{
		{
			name: "valid score",
			score: &types.IdentityScore{
				AccountAddress: "ve1account123",
				Score:          75,
				Status:         types.AccountStatusVerified,
				ComputedAt:     now,
			},
			expectError: false,
		},
		{
			name: "empty account address",
			score: &types.IdentityScore{
				AccountAddress: "",
				Score:          75,
				Status:         types.AccountStatusVerified,
				ComputedAt:     now,
			},
			expectError: true,
			errContains: "account address cannot be empty",
		},
		{
			name: "score exceeds maximum",
			score: &types.IdentityScore{
				AccountAddress: "ve1account123",
				Score:          101,
				Status:         types.AccountStatusVerified,
				ComputedAt:     now,
			},
			expectError: true,
			errContains: "exceeds maximum",
		},
		{
			name: "invalid status",
			score: &types.IdentityScore{
				AccountAddress: "ve1account123",
				Score:          75,
				Status:         types.AccountStatus("invalid"),
				ComputedAt:     now,
			},
			expectError: true,
			errContains: "invalid status",
		},
		{
			name: "zero computed_at",
			score: &types.IdentityScore{
				AccountAddress: "ve1account123",
				Score:          75,
				Status:         types.AccountStatusVerified,
				ComputedAt:     time.Time{},
			},
			expectError: true,
			errContains: "computed_at cannot be zero",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.score.Validate()
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIdentityScore_IsExpired(t *testing.T) {
	now := time.Now()

	t.Run("no expiry set", func(t *testing.T) {
		score := &types.IdentityScore{
			AccountAddress: "ve1account123",
			Score:          75,
			Status:         types.AccountStatusVerified,
			ComputedAt:     now,
			ExpiresAt:      nil,
		}
		assert.False(t, score.IsExpired())
	})

	t.Run("not yet expired", func(t *testing.T) {
		future := now.Add(24 * time.Hour)
		score := &types.IdentityScore{
			AccountAddress: "ve1account123",
			Score:          75,
			Status:         types.AccountStatusVerified,
			ComputedAt:     now,
			ExpiresAt:      &future,
		}
		assert.False(t, score.IsExpired())
	})

	t.Run("expired", func(t *testing.T) {
		past := now.Add(-24 * time.Hour)
		score := &types.IdentityScore{
			AccountAddress: "ve1account123",
			Score:          75,
			Status:         types.AccountStatusVerified,
			ComputedAt:     now,
			ExpiresAt:      &past,
		}
		assert.True(t, score.IsExpired())
	})
}

func TestIdentityScore_IsAboveThreshold(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		score     uint32
		status    types.AccountStatus
		threshold uint32
		expected  bool
	}{
		{"verified above threshold", 75, types.AccountStatusVerified, 70, true},
		{"verified at threshold", 70, types.AccountStatusVerified, 70, true},
		{"verified below threshold", 65, types.AccountStatusVerified, 70, false},
		{"not verified above threshold", 75, types.AccountStatusPending, 70, false},
		{"not verified below threshold", 50, types.AccountStatusPending, 70, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			score := &types.IdentityScore{
				AccountAddress: "ve1account123",
				Score:          tc.score,
				Status:         tc.status,
				ComputedAt:     now,
			}
			assert.Equal(t, tc.expected, score.IsAboveThreshold(tc.threshold))
		})
	}
}

// ============================================================================
// Score History Tests
// ============================================================================

func TestNewScoreHistoryEntry(t *testing.T) {
	now := time.Now()
	entry := types.NewScoreHistoryEntry(
		75,
		types.AccountStatusVerified,
		"v1.0.0",
		now,
		12345,
		"Initial verification",
	)

	require.NotNil(t, entry)
	assert.Equal(t, uint32(75), entry.Score)
	assert.Equal(t, types.AccountStatusVerified, entry.Status)
	assert.Equal(t, "v1.0.0", entry.ModelVersion)
	assert.Equal(t, now, entry.ComputedAt)
	assert.Equal(t, int64(12345), entry.BlockHeight)
	assert.Equal(t, "Initial verification", entry.Reason)
}

func TestNewScoreHistory(t *testing.T) {
	history := types.NewScoreHistory("ve1account123")

	require.NotNil(t, history)
	assert.Equal(t, "ve1account123", history.AccountAddress)
	assert.Empty(t, history.Entries)
}

func TestScoreHistory_AddEntry(t *testing.T) {
	history := types.NewScoreHistory("ve1account123")
	now := time.Now()

	entry1 := types.ScoreHistoryEntry{
		Score:       50,
		Status:      types.AccountStatusPending,
		ComputedAt:  now.Add(-1 * time.Hour),
		BlockHeight: 100,
	}

	entry2 := types.ScoreHistoryEntry{
		Score:       75,
		Status:      types.AccountStatusVerified,
		ComputedAt:  now,
		BlockHeight: 200,
	}

	history.AddEntry(entry1)
	assert.Len(t, history.Entries, 1)
	assert.Equal(t, uint32(50), history.Entries[0].Score)

	// Second entry should be prepended (newest first)
	history.AddEntry(entry2)
	assert.Len(t, history.Entries, 2)
	assert.Equal(t, uint32(75), history.Entries[0].Score)
	assert.Equal(t, uint32(50), history.Entries[1].Score)
}

func TestScoreHistory_GetLatest(t *testing.T) {
	t.Run("empty history", func(t *testing.T) {
		history := types.NewScoreHistory("ve1account123")
		_, ok := history.GetLatest()
		assert.False(t, ok)
	})

	t.Run("with entries", func(t *testing.T) {
		history := types.NewScoreHistory("ve1account123")
		now := time.Now()

		history.AddEntry(types.ScoreHistoryEntry{Score: 50, ComputedAt: now.Add(-1 * time.Hour)})
		history.AddEntry(types.ScoreHistoryEntry{Score: 75, ComputedAt: now})

		latest, ok := history.GetLatest()
		assert.True(t, ok)
		assert.Equal(t, uint32(75), latest.Score)
	})
}

// ============================================================================
// Offering Type Tests
// ============================================================================

func TestIsValidOfferingType(t *testing.T) {
	tests := []struct {
		offeringType types.OfferingType
		expected     bool
	}{
		{types.OfferingTypeBasic, true},
		{types.OfferingTypeStandard, true},
		{types.OfferingTypePremium, true},
		{types.OfferingTypeProvider, true},
		{types.OfferingTypeValidator, true},
		{types.OfferingType("invalid"), false},
		{types.OfferingType(""), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.offeringType), func(t *testing.T) {
			assert.Equal(t, tc.expected, types.IsValidOfferingType(tc.offeringType))
		})
	}
}

func TestAllOfferingTypes(t *testing.T) {
	offeringTypes := types.AllOfferingTypes()
	assert.Len(t, offeringTypes, 5)
	assert.Contains(t, offeringTypes, types.OfferingTypeBasic)
	assert.Contains(t, offeringTypes, types.OfferingTypeValidator)
}

func TestGetRequiredScopesForOffering(t *testing.T) {
	t.Run("basic offering", func(t *testing.T) {
		req := types.GetRequiredScopesForOffering(types.OfferingTypeBasic)
		assert.Equal(t, types.OfferingTypeBasic, req.OfferingType)
		assert.Equal(t, types.ThresholdBasic, req.MinimumScore)
		assert.Contains(t, req.RequiredScopeTypes, types.ScopeTypeIDDocument)
		assert.Contains(t, req.RequiredScopeTypes, types.ScopeTypeSelfie)
		assert.False(t, req.RequiresMFA)
	})

	t.Run("standard offering", func(t *testing.T) {
		req := types.GetRequiredScopesForOffering(types.OfferingTypeStandard)
		assert.Equal(t, types.OfferingTypeStandard, req.OfferingType)
		assert.Equal(t, types.ThresholdStandard, req.MinimumScore)
		assert.Contains(t, req.RequiredScopeTypes, types.ScopeTypeEmailProof)
		assert.False(t, req.RequiresMFA)
	})

	t.Run("premium offering", func(t *testing.T) {
		req := types.GetRequiredScopesForOffering(types.OfferingTypePremium)
		assert.Equal(t, types.OfferingTypePremium, req.OfferingType)
		assert.Equal(t, types.ThresholdPremium, req.MinimumScore)
		assert.Contains(t, req.RequiredScopeTypes, types.ScopeTypeFaceVideo)
		assert.True(t, req.RequiresMFA)
	})
}
