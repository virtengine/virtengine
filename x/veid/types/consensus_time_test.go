package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIdentityScopeIsActiveAtUsesExplicitConsensusTime(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC)
	scope := IdentityScope{ExpiresAt: &expiresAt}

	require.True(t, scope.IsActiveAt(expiresAt.Add(-time.Nanosecond)))
	require.True(t, scope.IsActiveAt(expiresAt))
	require.False(t, scope.IsActiveAt(expiresAt.Add(time.Nanosecond)))
	scope.Revoked = true
	require.False(t, scope.IsActiveAt(expiresAt.Add(-time.Hour)))
}

func TestWebVerificationActivityUsesExplicitConsensusTime(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2040, 1, 1, 0, 0, 0, 0, time.UTC)
	require.True(t, (&SSOLinkageMetadata{Status: SSOStatusVerified, ExpiresAt: &expiresAt}).IsActiveAt(expiresAt))
	require.False(t, (&SSOLinkageMetadata{Status: SSOStatusVerified, ExpiresAt: &expiresAt}).IsActiveAt(expiresAt.Add(time.Nanosecond)))
	require.True(t, (&DomainVerificationRecord{Status: DomainStatusVerified, ExpiresAt: &expiresAt}).IsActiveAt(expiresAt))
	require.False(t, (&DomainVerificationRecord{Status: DomainStatusVerified, ExpiresAt: &expiresAt}).IsActiveAt(expiresAt.Add(time.Nanosecond)))
}

func TestCalculateDomainScoreUsesProvidedTimeForExpiry(t *testing.T) {
	t.Parallel()

	verifiedAt := time.Date(2038, 1, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := verifiedAt.Add(365 * 24 * time.Hour)
	record := &DomainVerificationRecord{
		Status:     DomainStatusVerified,
		VerifiedAt: &verifiedAt,
		ExpiresAt:  &expiresAt,
	}

	require.NotZero(t, CalculateDomainScore(record, DefaultDomainScoringWeight(), expiresAt))
	require.Zero(t, CalculateDomainScore(record, DefaultDomainScoringWeight(), expiresAt.Add(time.Nanosecond)))
}

func TestPrivacyProofConstructorsUseExplicitConsensusTime(t *testing.T) {
	t.Parallel()

	at := time.Date(2042, 6, 7, 8, 9, 10, 0, time.FixedZone("test", 3600))
	duration := 2 * time.Hour

	age := NewAgeProofAt("age", "subject", 18, duration, at)
	residency := NewResidencyProofAt("residency", "subject", "US", duration, at)
	score := NewScoreThresholdProofAt("score", "subject", 75, duration, at)
	request := NewSelectiveDisclosureRequestAt("request", "requester", "subject", []ClaimType{ClaimTypeAgeOver18}, "test", duration, duration, at)
	proof := NewSelectiveDisclosureProofAt("proof", "subject", []ClaimType{ClaimTypeAgeOver18}, ProofSchemeRangeProof, duration, at)
	result := NewProofVerificationResultAt(true, []ClaimType{ClaimTypeAgeOver18}, "verifier", at)

	expected := at.UTC()
	require.Equal(t, expected, age.CreatedAt)
	require.Equal(t, expected.Add(duration), age.ValidUntil)
	require.Equal(t, expected, residency.CreatedAt)
	require.Equal(t, expected, score.CreatedAt)
	require.Equal(t, expected, request.CreatedAt)
	require.Equal(t, expected, proof.CreatedAt)
	require.Equal(t, expected, result.VerifiedAt)
}
