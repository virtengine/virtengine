package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Identity Record Tests (VE-1006: Test Coverage)
// ============================================================================

func TestNewIdentityRecord(t *testing.T) {
	now := time.Now()
	record := types.NewIdentityRecord("ve1account123", now)

	require.NotNil(t, record)
	assert.Equal(t, "ve1account123", record.AccountAddress)
	assert.Empty(t, record.ScopeRefs)
	assert.Equal(t, uint32(0), record.CurrentScore)
	assert.Equal(t, "", record.ScoreVersion)
	assert.Equal(t, now, record.CreatedAt)
	assert.Equal(t, now, record.UpdatedAt)
	assert.Equal(t, types.IdentityTierUnverified, record.Tier)
	assert.Empty(t, record.Flags)
	assert.False(t, record.Locked)
	assert.Empty(t, record.LockedReason)
}

func TestIdentityRecord_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		record      *types.IdentityRecord
		expectError bool
		errContains string
	}{
		{
			name:        "valid record",
			record:      types.NewIdentityRecord("ve1account123", now),
			expectError: false,
		},
		{
			name: "empty account address",
			record: &types.IdentityRecord{
				AccountAddress: "",
				CurrentScore:   50,
				Tier:           types.IdentityTierBasic,
				CreatedAt:      now,
			},
			expectError: true,
			errContains: "account address cannot be empty",
		},
		{
			name: "score exceeds maximum",
			record: &types.IdentityRecord{
				AccountAddress: "ve1account123",
				CurrentScore:   101,
				Tier:           types.IdentityTierBasic,
				CreatedAt:      now,
			},
			expectError: true,
			errContains: "score cannot exceed 100",
		},
		{
			name: "invalid tier",
			record: &types.IdentityRecord{
				AccountAddress: "ve1account123",
				CurrentScore:   50,
				Tier:           types.IdentityTier("invalid_tier"),
				CreatedAt:      now,
			},
			expectError: true,
			errContains: "invalid tier",
		},
		{
			name: "zero created_at",
			record: &types.IdentityRecord{
				AccountAddress: "ve1account123",
				CurrentScore:   50,
				Tier:           types.IdentityTierBasic,
				CreatedAt:      time.Time{},
			},
			expectError: true,
			errContains: "created_at cannot be zero",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.record.Validate()
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIdentityRecord_AddScopeRef(t *testing.T) {
	now := time.Now()
	record := types.NewIdentityRecord("ve1account123", now)

	ref1 := types.ScopeRef{
		ScopeID:   "scope-001",
		ScopeType: types.ScopeTypeIDDocument,
		Status:    types.VerificationStatusPending,
	}

	ref2 := types.ScopeRef{
		ScopeID:   "scope-002",
		ScopeType: types.ScopeTypeSelfie,
		Status:    types.VerificationStatusVerified,
	}

	// Add first scope
	record.AddScopeRef(ref1)
	assert.Len(t, record.ScopeRefs, 1)
	assert.Equal(t, "scope-001", record.ScopeRefs[0].ScopeID)

	// Add second scope
	record.AddScopeRef(ref2)
	assert.Len(t, record.ScopeRefs, 2)

	// Update existing scope (same ID, different status)
	ref1Updated := types.ScopeRef{
		ScopeID:   "scope-001",
		ScopeType: types.ScopeTypeIDDocument,
		Status:    types.VerificationStatusVerified,
	}
	record.AddScopeRef(ref1Updated)
	assert.Len(t, record.ScopeRefs, 2) // Should not add duplicate

	// Check the update was applied
	found, ok := record.GetScopeRef("scope-001")
	require.True(t, ok)
	assert.Equal(t, types.VerificationStatusVerified, found.Status)
}

func TestIdentityRecord_RemoveScopeRef(t *testing.T) {
	now := time.Now()
	record := types.NewIdentityRecord("ve1account123", now)

	ref := types.ScopeRef{
		ScopeID:   "scope-001",
		ScopeType: types.ScopeTypeIDDocument,
		Status:    types.VerificationStatusPending,
	}

	record.AddScopeRef(ref)
	assert.Len(t, record.ScopeRefs, 1)

	// Remove existing scope
	removed := record.RemoveScopeRef("scope-001")
	assert.True(t, removed)
	assert.Empty(t, record.ScopeRefs)

	// Try to remove non-existent scope
	removed = record.RemoveScopeRef("scope-999")
	assert.False(t, removed)
}

func TestIdentityRecord_GetScopeRef(t *testing.T) {
	now := time.Now()
	record := types.NewIdentityRecord("ve1account123", now)

	ref := types.ScopeRef{
		ScopeID:   "scope-001",
		ScopeType: types.ScopeTypeIDDocument,
		Status:    types.VerificationStatusVerified,
	}
	record.AddScopeRef(ref)

	// Get existing scope
	found, ok := record.GetScopeRef("scope-001")
	assert.True(t, ok)
	assert.Equal(t, "scope-001", found.ScopeID)
	assert.Equal(t, types.ScopeTypeIDDocument, found.ScopeType)

	// Get non-existent scope
	_, ok = record.GetScopeRef("scope-999")
	assert.False(t, ok)
}

func TestIdentityRecord_GetScopeRefsByType(t *testing.T) {
	now := time.Now()
	record := types.NewIdentityRecord("ve1account123", now)

	record.AddScopeRef(types.ScopeRef{ScopeID: "scope-001", ScopeType: types.ScopeTypeIDDocument})
	record.AddScopeRef(types.ScopeRef{ScopeID: "scope-002", ScopeType: types.ScopeTypeSelfie})
	record.AddScopeRef(types.ScopeRef{ScopeID: "scope-003", ScopeType: types.ScopeTypeIDDocument})

	// Get ID document scopes
	idDocs := record.GetScopeRefsByType(types.ScopeTypeIDDocument)
	assert.Len(t, idDocs, 2)

	// Get selfie scopes
	selfies := record.GetScopeRefsByType(types.ScopeTypeSelfie)
	assert.Len(t, selfies, 1)

	// Get non-existent type
	biometrics := record.GetScopeRefsByType(types.ScopeTypeBiometric)
	assert.Empty(t, biometrics)
}

func TestIdentityRecord_CountVerifiedScopes(t *testing.T) {
	now := time.Now()
	record := types.NewIdentityRecord("ve1account123", now)

	assert.Equal(t, 0, record.CountVerifiedScopes())

	record.AddScopeRef(types.ScopeRef{
		ScopeID: "scope-001",
		Status:  types.VerificationStatusPending,
	})
	assert.Equal(t, 0, record.CountVerifiedScopes())

	record.AddScopeRef(types.ScopeRef{
		ScopeID: "scope-002",
		Status:  types.VerificationStatusVerified,
	})
	assert.Equal(t, 1, record.CountVerifiedScopes())

	record.AddScopeRef(types.ScopeRef{
		ScopeID: "scope-003",
		Status:  types.VerificationStatusVerified,
	})
	assert.Equal(t, 2, record.CountVerifiedScopes())
}

func TestIdentityRecord_HasVerifiedScope(t *testing.T) {
	now := time.Now()
	record := types.NewIdentityRecord("ve1account123", now)

	// No scopes yet
	assert.False(t, record.HasVerifiedScope(types.ScopeTypeIDDocument))

	// Add pending scope
	record.AddScopeRef(types.ScopeRef{
		ScopeID:   "scope-001",
		ScopeType: types.ScopeTypeIDDocument,
		Status:    types.VerificationStatusPending,
	})
	assert.False(t, record.HasVerifiedScope(types.ScopeTypeIDDocument))

	// Add verified scope of different type
	record.AddScopeRef(types.ScopeRef{
		ScopeID:   "scope-002",
		ScopeType: types.ScopeTypeSelfie,
		Status:    types.VerificationStatusVerified,
	})
	assert.False(t, record.HasVerifiedScope(types.ScopeTypeIDDocument))
	assert.True(t, record.HasVerifiedScope(types.ScopeTypeSelfie))

	// Update ID document to verified
	record.AddScopeRef(types.ScopeRef{
		ScopeID:   "scope-001",
		ScopeType: types.ScopeTypeIDDocument,
		Status:    types.VerificationStatusVerified,
	})
	assert.True(t, record.HasVerifiedScope(types.ScopeTypeIDDocument))
}

func TestIdentityRecord_IsActive(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		record   *types.IdentityRecord
		expected bool
	}{
		{
			name: "active - not locked with score",
			record: &types.IdentityRecord{
				AccountAddress: "ve1account123",
				CurrentScore:   50,
				Locked:         false,
				CreatedAt:      now,
			},
			expected: true,
		},
		{
			name: "inactive - locked",
			record: &types.IdentityRecord{
				AccountAddress: "ve1account123",
				CurrentScore:   50,
				Locked:         true,
				CreatedAt:      now,
			},
			expected: false,
		},
		{
			name: "inactive - zero score",
			record: &types.IdentityRecord{
				AccountAddress: "ve1account123",
				CurrentScore:   0,
				Locked:         false,
				CreatedAt:      now,
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.record.IsActive())
		})
	}
}

// ============================================================================
// Identity Tier Tests
// ============================================================================

func TestIsValidIdentityTier(t *testing.T) {
	tests := []struct {
		tier     types.IdentityTier
		expected bool
	}{
		{types.IdentityTierUnverified, true},
		{types.IdentityTierBasic, true},
		{types.IdentityTierStandard, true},
		{types.IdentityTierVerified, true},
		{types.IdentityTierTrusted, true},
		{types.IdentityTierPremium, true},
		{types.IdentityTier("invalid"), false},
		{types.IdentityTier(""), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.tier), func(t *testing.T) {
			assert.Equal(t, tc.expected, types.IsValidIdentityTier(tc.tier))
		})
	}
}

func TestAllIdentityTiers(t *testing.T) {
	tiers := types.AllIdentityTiers()
	assert.Len(t, tiers, 6)
	assert.Contains(t, tiers, types.IdentityTierUnverified)
	assert.Contains(t, tiers, types.IdentityTierPremium)
}

func TestComputeTierFromScore(t *testing.T) {
	tests := []struct {
		score    uint32
		expected types.IdentityTier
	}{
		{0, types.IdentityTierUnverified},
		{1, types.IdentityTierBasic},
		{29, types.IdentityTierBasic},
		{30, types.IdentityTierStandard},
		{59, types.IdentityTierStandard},
		{60, types.IdentityTierVerified},
		{84, types.IdentityTierVerified},
		{85, types.IdentityTierTrusted},
		{100, types.IdentityTierTrusted},
	}

	for _, tc := range tests {
		t.Run(string(tc.expected), func(t *testing.T) {
			assert.Equal(t, tc.expected, types.ComputeTierFromScore(tc.score))
		})
	}
}

func TestTierMinimumScore(t *testing.T) {
	tests := []struct {
		tier     types.IdentityTier
		expected uint32
	}{
		{types.IdentityTierUnverified, 0},
		{types.IdentityTierBasic, 1},
		{types.IdentityTierStandard, 30},
		{types.IdentityTierVerified, 60},
		{types.IdentityTierTrusted, 85},
		{types.IdentityTier("unknown"), 0},
	}

	for _, tc := range tests {
		t.Run(string(tc.tier), func(t *testing.T) {
			assert.Equal(t, tc.expected, types.TierMinimumScore(tc.tier))
		})
	}
}

// ============================================================================
// Simple Identity Wallet Tests
// ============================================================================

func TestNewSimpleIdentityWallet(t *testing.T) {
	now := time.Now()
	wallet := types.NewSimpleIdentityWallet(
		"wallet-001",
		"ve1owner123",
		"ve1record456",
		"fingerprint-abc",
		now,
	)

	require.NotNil(t, wallet)
	assert.Equal(t, "wallet-001", wallet.WalletID)
	assert.Equal(t, "ve1owner123", wallet.OwnerAddress)
	assert.Equal(t, "ve1record456", wallet.IdentityRecordAddress)
	assert.Equal(t, "fingerprint-abc", wallet.PublicKeyFingerprint)
	assert.Equal(t, now, wallet.CreatedAt)
	assert.True(t, wallet.Active)
}

func TestSimpleIdentityWallet_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		wallet      *types.SimpleIdentityWallet
		expectError bool
		errContains string
	}{
		{
			name: "valid wallet",
			wallet: types.NewSimpleIdentityWallet(
				"wallet-001", "ve1owner123", "ve1record456", "fingerprint-abc", now,
			),
			expectError: false,
		},
		{
			name: "empty wallet ID",
			wallet: &types.SimpleIdentityWallet{
				WalletID:              "",
				OwnerAddress:          "ve1owner123",
				IdentityRecordAddress: "ve1record456",
				PublicKeyFingerprint:  "fingerprint-abc",
				CreatedAt:             now,
			},
			expectError: true,
			errContains: "wallet_id cannot be empty",
		},
		{
			name: "empty owner address",
			wallet: &types.SimpleIdentityWallet{
				WalletID:              "wallet-001",
				OwnerAddress:          "",
				IdentityRecordAddress: "ve1record456",
				PublicKeyFingerprint:  "fingerprint-abc",
				CreatedAt:             now,
			},
			expectError: true,
			errContains: "owner address cannot be empty",
		},
		{
			name: "empty identity record address",
			wallet: &types.SimpleIdentityWallet{
				WalletID:              "wallet-001",
				OwnerAddress:          "ve1owner123",
				IdentityRecordAddress: "",
				PublicKeyFingerprint:  "fingerprint-abc",
				CreatedAt:             now,
			},
			expectError: true,
			errContains: "identity record address cannot be empty",
		},
		{
			name: "empty public key fingerprint",
			wallet: &types.SimpleIdentityWallet{
				WalletID:              "wallet-001",
				OwnerAddress:          "ve1owner123",
				IdentityRecordAddress: "ve1record456",
				PublicKeyFingerprint:  "",
				CreatedAt:             now,
			},
			expectError: true,
			errContains: "public key fingerprint cannot be empty",
		},
		{
			name: "zero created_at",
			wallet: &types.SimpleIdentityWallet{
				WalletID:              "wallet-001",
				OwnerAddress:          "ve1owner123",
				IdentityRecordAddress: "ve1record456",
				PublicKeyFingerprint:  "fingerprint-abc",
				CreatedAt:             time.Time{},
			},
			expectError: true,
			errContains: "created_at cannot be zero",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.wallet.Validate()
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
