package types_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pkg.akt.dev/node/x/veid/types"
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
				"user@example.com",
				types.SSOProviderGoogle,
				"https://accounts.google.com",
				"123456789",
				now,
			),
			wantErr: false,
		},
		{
			name: "valid Microsoft SSO",
			metadata: types.NewSSOLinkageMetadata(
				"user@contoso.com",
				types.SSOProviderMicrosoft,
				"https://login.microsoftonline.com/tenant",
				"user-oid-12345",
				now,
			),
			wantErr: false,
		},
		{
			name: "valid GitHub SSO",
			metadata: types.NewSSOLinkageMetadata(
				"github_user",
				types.SSOProviderGitHub,
				"https://github.com",
				"12345678",
				now,
			),
			wantErr: false,
		},
		{
			name: "valid custom OIDC provider",
			metadata: types.NewSSOLinkageMetadata(
				"custom_user",
				types.SSOProviderOIDC,
				"https://auth.custom.com",
				"custom-sub-123",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty subject ID",
			metadata: &types.SSOLinkageMetadata{
				Version:    types.SSOLinkageVersion,
				SubjectID:  "",
				Provider:   types.SSOProviderGoogle,
				IssuerURL:  "https://accounts.google.com",
				ExternalID: "123456789",
				LinkedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty issuer URL",
			metadata: &types.SSOLinkageMetadata{
				Version:    types.SSOLinkageVersion,
				SubjectID:  "user@example.com",
				Provider:   types.SSOProviderGoogle,
				IssuerURL:  "",
				ExternalID: "123456789",
				LinkedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty external ID",
			metadata: &types.SSOLinkageMetadata{
				Version:    types.SSOLinkageVersion,
				SubjectID:  "user@example.com",
				Provider:   types.SSOProviderGoogle,
				IssuerURL:  "https://accounts.google.com",
				ExternalID: "",
				LinkedAt:   now,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid provider type",
			metadata: &types.SSOLinkageMetadata{
				Version:    types.SSOLinkageVersion,
				SubjectID:  "user@example.com",
				Provider:   "invalid_provider",
				IssuerURL:  "https://example.com",
				ExternalID: "123",
				LinkedAt:   now,
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
		name       string
		issuer     string
		externalID string
	}{
		{
			name:       "Google subject",
			issuer:     "https://accounts.google.com",
			externalID: "123456789012345678901",
		},
		{
			name:       "Microsoft subject",
			issuer:     "https://login.microsoftonline.com/tenant-id",
			externalID: "user-object-id-12345",
		},
		{
			name:       "GitHub subject",
			issuer:     "https://github.com",
			externalID: "12345678",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := types.HashSubjectID(tt.issuer, tt.externalID)

			// Hash should be 64 hex characters (SHA256)
			assert.Len(t, hash, 64, "HashSubjectID() length should be 64")

			// Hash should be deterministic
			hash2 := types.HashSubjectID(tt.issuer, tt.externalID)
			assert.Equal(t, hash, hash2, "HashSubjectID() should be deterministic")

			// Different inputs should produce different hashes
			hash3 := types.HashSubjectID(tt.issuer, tt.externalID+"x")
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
	future := now.Add(15 * time.Minute)

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
				"https://accounts.google.com",
				"random-state-value",
				"random-nonce-value",
				"https://app.example.com/callback",
				now,
				future,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty challenge ID",
			challenge: &types.SSOVerificationChallenge{
				Version:     types.SSOChallengeVersion,
				ChallengeID: "",
				VeidOwner:   "cosmos1abc...",
				Provider:    types.SSOProviderGoogle,
				State:       "random-state",
				Nonce:       "random-nonce",
				CreatedAt:   now,
				ExpiresAt:   future,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty state",
			challenge: &types.SSOVerificationChallenge{
				Version:     types.SSOChallengeVersion,
				ChallengeID: "challenge-123",
				VeidOwner:   "cosmos1abc...",
				Provider:    types.SSOProviderGoogle,
				State:       "",
				Nonce:       "random-nonce",
				CreatedAt:   now,
				ExpiresAt:   future,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty nonce",
			challenge: &types.SSOVerificationChallenge{
				Version:     types.SSOChallengeVersion,
				ChallengeID: "challenge-123",
				VeidOwner:   "cosmos1abc...",
				Provider:    types.SSOProviderGoogle,
				State:       "random-state",
				Nonce:       "",
				CreatedAt:   now,
				ExpiresAt:   future,
			},
			wantErr: true,
		},
		{
			name: "invalid - expired",
			challenge: &types.SSOVerificationChallenge{
				Version:     types.SSOChallengeVersion,
				ChallengeID: "challenge-123",
				VeidOwner:   "cosmos1abc...",
				Provider:    types.SSOProviderGoogle,
				State:       "random-state",
				Nonce:       "random-nonce",
				CreatedAt:   now.Add(-1 * time.Hour),
				ExpiresAt:   now.Add(-30 * time.Minute),
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

	notExpired := &types.SSOVerificationChallenge{
		ExpiresAt: now.Add(15 * time.Minute),
	}
	assert.False(t, notExpired.IsExpired(now), "Challenge should not be expired")

	expired := &types.SSOVerificationChallenge{
		ExpiresAt: now.Add(-1 * time.Minute),
	}
	assert.True(t, expired.IsExpired(now), "Challenge should be expired")
}

func TestSSOLinkageMetadata_IsRevoked(t *testing.T) {
	now := time.Now()

	active := &types.SSOLinkageMetadata{
		LinkedAt:  now.Add(-1 * time.Hour),
		RevokedAt: nil,
	}
	assert.False(t, active.IsRevoked(), "Linkage should not be revoked")

	revokedAt := now.Add(-30 * time.Minute)
	revoked := &types.SSOLinkageMetadata{
		LinkedAt:  now.Add(-1 * time.Hour),
		RevokedAt: &revokedAt,
	}
	assert.True(t, revoked.IsRevoked(), "Linkage should be revoked")
}

func TestSSOLinkageMetadata_GetHashedSubjectID(t *testing.T) {
	now := time.Now()

	metadata := types.NewSSOLinkageMetadata(
		"user@example.com",
		types.SSOProviderGoogle,
		"https://accounts.google.com",
		"123456789",
		now,
	)

	hashed := metadata.GetHashedSubjectID()

	// Should be 64 hex characters
	assert.Len(t, hashed, 64, "GetHashedSubjectID() length should be 64")

	// Should be lowercase hex
	assert.Equal(t, strings.ToLower(hashed), hashed, "GetHashedSubjectID() should return lowercase hex")

	// Should be deterministic
	hashed2 := metadata.GetHashedSubjectID()
	assert.Equal(t, hashed, hashed2, "GetHashedSubjectID() should be deterministic")
}
