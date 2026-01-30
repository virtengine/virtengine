package types_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// SSO Verification Tests (VE-222: SSO Verification Scope)
// ============================================================================

func TestSSOLinkageMetadata_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		metadata *types.SSOLinkageMetadata
		wantErr  bool
	}{
		{
			name: "valid Google SSO",
			metadata: types.NewSSOLinkageMetadata(
				"linkage-123",
				types.SSOProviderGoogle,
				"https://accounts.google.com",
				"123456789",
				"random-nonce",
				now,
			),
			wantErr: false,
		},
		{
			name: "valid Microsoft SSO",
			metadata: types.NewSSOLinkageMetadata(
				"linkage-456",
				types.SSOProviderMicrosoft,
				"https://login.microsoftonline.com/tenant",
				"user-oid-12345",
				"nonce-abc",
				now,
			),
			wantErr: false,
		},
		{
			name: "valid GitHub SSO",
			metadata: types.NewSSOLinkageMetadata(
				"linkage-789",
				types.SSOProviderGitHub,
				"https://github.com",
				"12345678",
				"nonce-def",
				now,
			),
			wantErr: false,
		},
		{
			name: "valid custom OIDC provider",
			metadata: types.NewSSOLinkageMetadata(
				"linkage-custom",
				types.SSOProviderOIDC,
				"https://auth.custom.com",
				"custom-sub-123",
				"nonce-ghi",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty linkage ID",
			metadata: &types.SSOLinkageMetadata{
				Version:     types.SSOVerificationVersion,
				LinkageID:   "",
				Provider:    types.SSOProviderGoogle,
				Issuer:      "https://accounts.google.com",
				SubjectHash: types.HashSubjectID("123456789"),
				Nonce:       "nonce",
				VerifiedAt:  now,
				Status:      types.SSOStatusVerified,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty issuer",
			metadata: &types.SSOLinkageMetadata{
				Version:     types.SSOVerificationVersion,
				LinkageID:   "linkage-123",
				Provider:    types.SSOProviderGoogle,
				Issuer:      "",
				SubjectHash: types.HashSubjectID("123456789"),
				Nonce:       "nonce",
				VerifiedAt:  now,
				Status:      types.SSOStatusVerified,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty subject hash",
			metadata: &types.SSOLinkageMetadata{
				Version:     types.SSOVerificationVersion,
				LinkageID:   "linkage-123",
				Provider:    types.SSOProviderGoogle,
				Issuer:      "https://accounts.google.com",
				SubjectHash: "",
				Nonce:       "nonce",
				VerifiedAt:  now,
				Status:      types.SSOStatusVerified,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid provider type",
			metadata: &types.SSOLinkageMetadata{
				Version:     types.SSOVerificationVersion,
				LinkageID:   "linkage-123",
				Provider:    "invalid_provider",
				Issuer:      "https://example.com",
				SubjectHash: types.HashSubjectID("123"),
				Nonce:       "nonce",
				VerifiedAt:  now,
				Status:      types.SSOStatusVerified,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metadata.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHashSubjectID(t *testing.T) {
	tests := []struct {
		name      string
		subjectID string
	}{
		{
			name:      "Google subject",
			subjectID: "123456789012345678901",
		},
		{
			name:      "Microsoft subject",
			subjectID: "user-object-id-12345",
		},
		{
			name:      "GitHub subject",
			subjectID: "12345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := types.HashSubjectID(tt.subjectID)

			// Hash should be 64 hex characters (SHA256)
			assert.Len(t, hash, 64, "HashSubjectID() length should be 64")

			// Hash should be deterministic
			hash2 := types.HashSubjectID(tt.subjectID)
			assert.Equal(t, hash, hash2, "HashSubjectID() should be deterministic")

			// Different inputs should produce different hashes
			hash3 := types.HashSubjectID(tt.subjectID + "x")
			assert.NotEqual(t, hash, hash3, "HashSubjectID() should produce different hashes for different inputs")
		})
	}
}

func TestSSOProviderTypes(t *testing.T) {
	// Test all provider types are valid
	for _, provider := range types.AllSSOProviderTypes() {
		assert.True(t, types.IsValidSSOProviderType(provider), "AllSSOProviderTypes returned invalid type: %s", provider)
	}

	// Test invalid provider type
	assert.False(t, types.IsValidSSOProviderType("invalid"), "IsValidSSOProviderType should return false for invalid type")
}

func TestSSOVerificationChallenge_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		challenge *types.SSOVerificationChallenge
		wantErr   bool
	}{
		{
			name: "valid challenge",
			challenge: types.NewSSOVerificationChallenge(
				"challenge-123",
				"cosmos1abc...",
				types.SSOProviderGoogle,
				"random-nonce-value",
				now,
				900, // 15 minutes TTL
			),
			wantErr: false,
		},
		{
			name: "invalid - empty challenge ID",
			challenge: &types.SSOVerificationChallenge{
				ChallengeID:    "",
				AccountAddress: "cosmos1abc...",
				Provider:       types.SSOProviderGoogle,
				Nonce:          "random-nonce",
				CreatedAt:      now,
				ExpiresAt:      now.Add(15 * time.Minute),
				Status:         types.SSOStatusPending,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty nonce",
			challenge: &types.SSOVerificationChallenge{
				ChallengeID:    "challenge-123",
				AccountAddress: "cosmos1abc...",
				Provider:       types.SSOProviderGoogle,
				Nonce:          "",
				CreatedAt:      now,
				ExpiresAt:      now.Add(15 * time.Minute),
				Status:         types.SSOStatusPending,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty account address",
			challenge: &types.SSOVerificationChallenge{
				ChallengeID:    "challenge-123",
				AccountAddress: "",
				Provider:       types.SSOProviderGoogle,
				Nonce:          "random-nonce",
				CreatedAt:      now,
				ExpiresAt:      now.Add(15 * time.Minute),
				Status:         types.SSOStatusPending,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.challenge.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSSOVerificationChallenge_IsExpired(t *testing.T) {
	now := time.Now()

	notExpired := types.NewSSOVerificationChallenge(
		"challenge-123",
		"cosmos1abc...",
		types.SSOProviderGoogle,
		"random-nonce",
		now,
		900, // 15 minutes TTL
	)
	assert.False(t, notExpired.IsExpired(now), "Challenge should not be expired")

	// Create an expired challenge
	expired := types.NewSSOVerificationChallenge(
		"challenge-456",
		"cosmos1abc...",
		types.SSOProviderGoogle,
		"random-nonce",
		now.Add(-20*time.Minute), // Created 20 minutes ago
		900,                      // 15 minutes TTL (so expired)
	)
	assert.True(t, expired.IsExpired(now), "Challenge should be expired")
}

func TestSSOLinkageMetadata_IsActive(t *testing.T) {
	now := time.Now()

	// Active linkage
	active := types.NewSSOLinkageMetadata(
		"linkage-123",
		types.SSOProviderGoogle,
		"https://accounts.google.com",
		"subject-123",
		"nonce",
		now.Add(-1*time.Hour),
	)
	assert.True(t, active.IsActive(), "Linkage should be active")

	// Revoked linkage
	revoked := types.NewSSOLinkageMetadata(
		"linkage-456",
		types.SSOProviderGoogle,
		"https://accounts.google.com",
		"subject-456",
		"nonce",
		now.Add(-1*time.Hour),
	)
	revoked.Status = types.SSOStatusRevoked
	assert.False(t, revoked.IsActive(), "Revoked linkage should not be active")

	// Expired linkage
	expiredTime := now.Add(-30 * time.Minute)
	expiredLinkage := types.NewSSOLinkageMetadata(
		"linkage-789",
		types.SSOProviderGoogle,
		"https://accounts.google.com",
		"subject-789",
		"nonce",
		now.Add(-1*time.Hour),
	)
	expiredLinkage.ExpiresAt = &expiredTime
	assert.False(t, expiredLinkage.IsActive(), "Expired linkage should not be active")
}

func TestSSOLinkageMetadata_String(t *testing.T) {
	now := time.Now()

	metadata := types.NewSSOLinkageMetadata(
		"linkage-123",
		types.SSOProviderGoogle,
		"https://accounts.google.com",
		"subject-123",
		"nonce",
		now,
	)

	str := metadata.String()
	assert.Contains(t, str, "linkage-123", "String should contain linkage ID")
	assert.Contains(t, str, "google", "String should contain provider")
	assert.Contains(t, strings.ToLower(str), "verified", "String should contain status")
}

func TestSSOScoringWeights(t *testing.T) {
	weights := types.DefaultSSOScoringWeights()
	assert.NotEmpty(t, weights, "Default SSO scoring weights should not be empty")

	// Check that Google provider has a weight
	googleWeight := types.GetSSOScoringWeight(types.SSOProviderGoogle)
	assert.Greater(t, googleWeight, uint32(0), "Google provider should have positive weight")

	// Check that invalid provider returns 0
	invalidWeight := types.GetSSOScoringWeight("invalid")
	assert.Equal(t, uint32(0), invalidWeight, "Invalid provider should return 0 weight")
}
