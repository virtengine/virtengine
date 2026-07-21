package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
)

func TestSelectBestEnrollment_NoActive(t *testing.T) {
	enrollments := []mfatypes.FactorEnrollment{
		{
			FactorType: mfatypes.FactorTypeTOTP,
			FactorID:   "totp-1",
			Status:     mfatypes.EnrollmentStatusPending,
			EnrolledAt: 1000,
		},
		{
			FactorType: mfatypes.FactorTypeFIDO2,
			FactorID:   "fido-1",
			Status:     mfatypes.EnrollmentStatusRevoked,
			EnrolledAt: 2000,
		},
	}

	_, ok := selectBestEnrollment(enrollments)
	assert.False(t, ok, "should return false when no active enrollments")
}

func TestSelectBestEnrollment_Empty(t *testing.T) {
	_, ok := selectBestEnrollment(nil)
	assert.False(t, ok, "should return false for nil enrollments")

	_, ok = selectBestEnrollment([]mfatypes.FactorEnrollment{})
	assert.False(t, ok, "should return false for empty enrollments")
}

func TestSelectBestEnrollment_SingleActive(t *testing.T) {
	enrollments := []mfatypes.FactorEnrollment{
		{
			FactorType: mfatypes.FactorTypeTOTP,
			FactorID:   "totp-1",
			Status:     mfatypes.EnrollmentStatusActive,
			EnrolledAt: 1000,
			VerifiedAt: 1100,
		},
	}

	best, ok := selectBestEnrollment(enrollments)
	require.True(t, ok)
	assert.Equal(t, mfatypes.FactorTypeTOTP, best.FactorType)
	assert.Equal(t, "totp-1", best.FactorID)
}

func TestSelectBestEnrollment_PrefersHigherSecurity(t *testing.T) {
	// FIDO2 has higher security than TOTP
	enrollments := []mfatypes.FactorEnrollment{
		{
			FactorType: mfatypes.FactorTypeTOTP,
			FactorID:   "totp-1",
			Status:     mfatypes.EnrollmentStatusActive,
			EnrolledAt: 2000,
			VerifiedAt: 2100,
		},
		{
			FactorType: mfatypes.FactorTypeFIDO2,
			FactorID:   "fido-1",
			Status:     mfatypes.EnrollmentStatusActive,
			EnrolledAt: 1000,
			VerifiedAt: 1100,
		},
	}

	best, ok := selectBestEnrollment(enrollments)
	require.True(t, ok)
	assert.Equal(t, mfatypes.FactorTypeFIDO2, best.FactorType,
		"should prefer FIDO2 (higher security) over TOTP")
}

func TestSelectBestEnrollment_SameSecurityPrefersRecentVerification(t *testing.T) {
	// Two TOTP enrollments, same security level — pick the one verified more recently
	enrollments := []mfatypes.FactorEnrollment{
		{
			FactorType: mfatypes.FactorTypeTOTP,
			FactorID:   "totp-old",
			Status:     mfatypes.EnrollmentStatusActive,
			EnrolledAt: 1000,
			VerifiedAt: 1100,
		},
		{
			FactorType: mfatypes.FactorTypeTOTP,
			FactorID:   "totp-new",
			Status:     mfatypes.EnrollmentStatusActive,
			EnrolledAt: 2000,
			VerifiedAt: 2100,
		},
	}

	best, ok := selectBestEnrollment(enrollments)
	require.True(t, ok)
	assert.Equal(t, "totp-new", best.FactorID,
		"should prefer enrollment verified more recently")
}

func TestSelectBestEnrollment_FiltersInactive(t *testing.T) {
	enrollments := []mfatypes.FactorEnrollment{
		{
			FactorType: mfatypes.FactorTypeFIDO2,
			FactorID:   "fido-revoked",
			Status:     mfatypes.EnrollmentStatusRevoked,
			EnrolledAt: 3000,
			VerifiedAt: 3100,
		},
		{
			FactorType: mfatypes.FactorTypeTOTP,
			FactorID:   "totp-active",
			Status:     mfatypes.EnrollmentStatusActive,
			EnrolledAt: 1000,
			VerifiedAt: 1100,
		},
		{
			FactorType: mfatypes.FactorTypeHardwareKey,
			FactorID:   "hw-expired",
			Status:     mfatypes.EnrollmentStatusExpired,
			EnrolledAt: 2000,
			VerifiedAt: 2100,
		},
	}

	best, ok := selectBestEnrollment(enrollments)
	require.True(t, ok)
	assert.Equal(t, "totp-active", best.FactorID,
		"should only consider active enrollments")
}

func TestResolveSensitiveTxType_EmptyString(t *testing.T) {
	result := resolveSensitiveTxType("")
	assert.Equal(t, mfatypes.SensitiveTxUnspecified, result)
}

func TestResolveSensitiveTxType_MarketplaceActions(t *testing.T) {
	tests := []struct {
		action   string
		expected mfatypes.SensitiveTransactionType
	}{
		{"place_order", mfatypes.SensitiveTxHighValueOrder},
		{"modify_order", mfatypes.SensitiveTxHighValueOrder},
		{"place_bid", mfatypes.SensitiveTxHighValueOrder},
		{"accept_bid", mfatypes.SensitiveTxHighValueOrder},
		{"settlement", mfatypes.SensitiveTxHighValueOrder},
	}

	for _, tc := range tests {
		t.Run(tc.action, func(t *testing.T) {
			result := resolveSensitiveTxType(tc.action)
			assert.Equal(t, tc.expected, result,
				"action %q should resolve to SensitiveTxHighValueOrder", tc.action)
		})
	}
}

func TestResolveSensitiveTxType_SpecialActions(t *testing.T) {
	result := resolveSensitiveTxType("create_offering")
	assert.Equal(t, mfatypes.SensitiveTxFirstOfferingCreate, result)

	result = resolveSensitiveTxType("key_rotation")
	assert.Equal(t, mfatypes.SensitiveTxKeyRotation, result)
}

func TestResolveSensitiveTxType_UnknownAction(t *testing.T) {
	result := resolveSensitiveTxType("some_unknown_action")
	assert.Equal(t, mfatypes.SensitiveTxUnspecified, result)
}

func TestResolveSensitiveTxType_ValidStringTypes(t *testing.T) {
	// These should be resolved via SensitiveTransactionTypeFromString
	result := resolveSensitiveTxType("account_recovery")
	assert.Equal(t, mfatypes.SensitiveTxAccountRecovery, result)

	result = resolveSensitiveTxType("large_withdrawal")
	assert.Equal(t, mfatypes.SensitiveTxLargeWithdrawal, result)
}

func TestHasVerifiedActiveScope_Empty(t *testing.T) {
	now := time.Now()
	assert.False(t, hasVerifiedActiveScope(nil, now), "nil scopes should return false")
	assert.False(t, hasVerifiedActiveScope([]veidtypes.IdentityScope{}, now),
		"empty scopes should return false")
}

func TestHasVerifiedActiveScope_ActiveAndVerified(t *testing.T) {
	now := time.Now()
	scopes := []veidtypes.IdentityScope{
		{
			ScopeID:    "scope-1",
			Status:     veidtypes.VerificationStatusVerified,
			VerifiedAt: &now,
			Revoked:    false,
		},
	}

	assert.True(t, hasVerifiedActiveScope(scopes, now),
		"verified and active scope should return true")
}

func TestHasVerifiedActiveScope_RevokedScope(t *testing.T) {
	now := time.Now()
	scopes := []veidtypes.IdentityScope{
		{
			ScopeID:    "scope-1",
			Status:     veidtypes.VerificationStatusVerified,
			VerifiedAt: &now,
			Revoked:    true,
		},
	}

	assert.False(t, hasVerifiedActiveScope(scopes, now),
		"revoked scope should not be considered active")
}

func TestHasVerifiedActiveScope_PendingScope(t *testing.T) {
	now := time.Now()
	scopes := []veidtypes.IdentityScope{
		{
			ScopeID: "scope-1",
			Status:  veidtypes.VerificationStatusPending,
			Revoked: false,
		},
	}

	assert.False(t, hasVerifiedActiveScope(scopes, now),
		"pending scope should not be considered verified")
}

func TestHasVerifiedActiveScope_ExpiredScope(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	scopes := []veidtypes.IdentityScope{
		{
			ScopeID:    "scope-1",
			Status:     veidtypes.VerificationStatusVerified,
			VerifiedAt: &now,
			ExpiresAt:  &past,
			Revoked:    false,
		},
	}

	assert.False(t, hasVerifiedActiveScope(scopes, now),
		"expired scope should not be considered active")
}

func TestHasVerifiedActiveScope_MixedScopes(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)

	scopes := []veidtypes.IdentityScope{
		{
			ScopeID: "scope-revoked",
			Status:  veidtypes.VerificationStatusVerified,
			Revoked: true,
		},
		{
			ScopeID:   "scope-expired",
			Status:    veidtypes.VerificationStatusVerified,
			ExpiresAt: &past,
			Revoked:   false,
		},
		{
			ScopeID: "scope-pending",
			Status:  veidtypes.VerificationStatusPending,
			Revoked: false,
		},
		{
			ScopeID:    "scope-valid",
			Status:     veidtypes.VerificationStatusVerified,
			VerifiedAt: &now,
			Revoked:    false,
		},
	}

	assert.True(t, hasVerifiedActiveScope(scopes, now),
		"should return true when at least one scope is verified and active")
}
