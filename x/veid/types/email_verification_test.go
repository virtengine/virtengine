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
// Email Verification Tests (VE-224: Email Verification Scope)
// ============================================================================

func TestEmailVerificationRecord_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		record  *types.EmailVerificationRecord
		wantErr bool
	}{
		{
			name: "valid email verification",
			record: types.NewEmailVerificationRecord(
				"email-1",
				"cosmos1abc...",
				"user@example.com",
				now,
			),
			wantErr: false,
		},
		{
			name: "valid subdomain email",
			record: types.NewEmailVerificationRecord(
				"email-2",
				"cosmos1abc...",
				"user@mail.subdomain.example.org",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty email ID",
			record: &types.EmailVerificationRecord{
				Version: types.EmailVerificationVersion,
				EmailID: "",
				Owner:   "cosmos1abc...",
				Email:   "user@example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid - empty owner",
			record: &types.EmailVerificationRecord{
				Version: types.EmailVerificationVersion,
				EmailID: "email-1",
				Owner:   "",
				Email:   "user@example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid - empty email",
			record: &types.EmailVerificationRecord{
				Version: types.EmailVerificationVersion,
				EmailID: "email-1",
				Owner:   "cosmos1abc...",
				Email:   "",
			},
			wantErr: true,
		},
		{
			name: "invalid - email without @",
			record: &types.EmailVerificationRecord{
				Version: types.EmailVerificationVersion,
				EmailID: "email-1",
				Owner:   "cosmos1abc...",
				Email:   "userexample.com",
			},
			wantErr: true,
		},
		{
			name: "invalid - email without domain",
			record: &types.EmailVerificationRecord{
				Version: types.EmailVerificationVersion,
				EmailID: "email-1",
				Owner:   "cosmos1abc...",
				Email:   "user@",
			},
			wantErr: true,
		},
		{
			name: "invalid - email without local part",
			record: &types.EmailVerificationRecord{
				Version: types.EmailVerificationVersion,
				EmailID: "email-1",
				Owner:   "cosmos1abc...",
				Email:   "@example.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHashEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{name: "simple email", email: "user@example.com"},
		{name: "email with plus", email: "user+tag@example.com"},
		{name: "subdomain email", email: "user@mail.example.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := types.HashEmail(tt.email)

			// Hash should be 64 hex characters (SHA256)
			assert.Len(t, hash, 64, "HashEmail() length should be 64")

			// Hash should be deterministic
			hash2 := types.HashEmail(tt.email)
			assert.Equal(t, hash, hash2, "HashEmail() should be deterministic")

			// Case insensitive (email domain part is case insensitive)
			hash3 := types.HashEmail(strings.ToLower(tt.email))
			assert.Equal(t, hash, hash3, "HashEmail() should be case insensitive")

			// Different emails should produce different hashes
			hash4 := types.HashEmail(tt.email + "x")
			assert.NotEqual(t, hash, hash4, "HashEmail() should produce different hashes for different emails")
		})
	}
}

func TestEmailVerificationStatuses(t *testing.T) {
	// Test all statuses are valid
	for _, status := range types.AllEmailVerificationStatuses() {
		assert.True(t, types.IsValidEmailVerificationStatus(status), "AllEmailVerificationStatuses returned invalid status: %s", status)
	}

	// Test invalid status
	assert.False(t, types.IsValidEmailVerificationStatus("invalid"), "IsValidEmailVerificationStatus should return false for invalid status")
}

func TestEmailVerificationChallenge_Validate(t *testing.T) {
	now := time.Now()
	future := now.Add(types.EmailChallengeValidDuration)

	tests := []struct {
		name      string
		challenge *types.EmailVerificationChallenge
		wantErr   bool
	}{
		{
			name: "valid challenge",
			challenge: types.NewEmailVerificationChallenge(
				"challenge-1",
				"email-1",
				"verification-code-123456",
				"random-nonce-abc123",
				now,
				future,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty challenge ID",
			challenge: &types.EmailVerificationChallenge{
				Version:          types.EmailChallengeVersion,
				ChallengeID:      "",
				EmailID:          "email-1",
				VerificationCode: "code123",
				Nonce:            "nonce123",
				CreatedAt:        now,
				ExpiresAt:        future,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty email ID",
			challenge: &types.EmailVerificationChallenge{
				Version:          types.EmailChallengeVersion,
				ChallengeID:      "challenge-1",
				EmailID:          "",
				VerificationCode: "code123",
				Nonce:            "nonce123",
				CreatedAt:        now,
				ExpiresAt:        future,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty verification code",
			challenge: &types.EmailVerificationChallenge{
				Version:          types.EmailChallengeVersion,
				ChallengeID:      "challenge-1",
				EmailID:          "email-1",
				VerificationCode: "",
				Nonce:            "nonce123",
				CreatedAt:        now,
				ExpiresAt:        future,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty nonce",
			challenge: &types.EmailVerificationChallenge{
				Version:          types.EmailChallengeVersion,
				ChallengeID:      "challenge-1",
				EmailID:          "email-1",
				VerificationCode: "code123",
				Nonce:            "",
				CreatedAt:        now,
				ExpiresAt:        future,
			},
			wantErr: true,
		},
		{
			name: "invalid - expired challenge",
			challenge: &types.EmailVerificationChallenge{
				Version:          types.EmailChallengeVersion,
				ChallengeID:      "challenge-1",
				EmailID:          "email-1",
				VerificationCode: "code123",
				Nonce:            "nonce123",
				CreatedAt:        now.Add(-2 * time.Hour),
				ExpiresAt:        now.Add(-1 * time.Hour),
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

func TestEmailVerificationChallenge_IsExpired(t *testing.T) {
	now := time.Now()

	notExpired := &types.EmailVerificationChallenge{ExpiresAt: now.Add(1 * time.Hour)}
	assert.False(t, notExpired.IsExpired(now), "Challenge should not be expired")

	expired := &types.EmailVerificationChallenge{ExpiresAt: now.Add(-1 * time.Minute)}
	assert.True(t, expired.IsExpired(now), "Challenge should be expired")
}

func TestEmailVerificationChallenge_IncrementAttempts(t *testing.T) {
	now := time.Now()
	future := now.Add(types.EmailChallengeValidDuration)

	challenge := types.NewEmailVerificationChallenge(
		"challenge-1",
		"email-1",
		"code123",
		"nonce123",
		now,
		future,
	)

	assert.Equal(t, uint8(0), challenge.Attempts, "Initial attempts should be 0")

	challenge.IncrementAttempts()
	assert.Equal(t, uint8(1), challenge.Attempts, "After increment, attempts should be 1")

	// Test max attempts
	for i := 0; i < types.MaxEmailVerificationAttempts; i++ {
		challenge.IncrementAttempts()
	}

	assert.True(t, challenge.HasExceededMaxAttempts(), "Challenge should have exceeded max attempts")
}

func TestUsedNonceRecord_Validate(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name    string
		record  *types.UsedNonceRecord
		wantErr bool
	}{
		{
			name: "valid nonce record",
			record: types.NewUsedNonceRecord(
				"nonce-hash-abc123",
				"email",
				now,
				future,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty nonce hash",
			record: &types.UsedNonceRecord{
				Version:   types.UsedNonceVersion,
				NonceHash: "",
				Context:   "email",
				UsedAt:    now,
				ExpiresAt: future,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty context",
			record: &types.UsedNonceRecord{
				Version:   types.UsedNonceVersion,
				NonceHash: "abc123",
				Context:   "",
				UsedAt:    now,
				ExpiresAt: future,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.record.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestUsedNonceRecord_CanBeCleanedUp(t *testing.T) {
	now := time.Now()

	canCleanUp := &types.UsedNonceRecord{ExpiresAt: now.Add(-1 * time.Hour)}
	assert.True(t, canCleanUp.CanBeCleanedUp(now), "Expired nonce should be eligible for cleanup")

	cannotCleanUp := &types.UsedNonceRecord{ExpiresAt: now.Add(1 * time.Hour)}
	assert.False(t, cannotCleanUp.CanBeCleanedUp(now), "Non-expired nonce should not be eligible for cleanup")
}

func TestEmailVerificationRecord_IsVerified(t *testing.T) {
	now := time.Now()

	verified := &types.EmailVerificationRecord{Status: types.EmailStatusVerified, VerifiedAt: &now}
	assert.True(t, verified.IsVerified(), "Record should be verified")

	pending := &types.EmailVerificationRecord{Status: types.EmailStatusPending}
	assert.False(t, pending.IsVerified(), "Pending record should not be verified")
}

func TestEmailVerificationRecord_GetHashedEmail(t *testing.T) {
	now := time.Now()

	record := types.NewEmailVerificationRecord(
		"email-1",
		"cosmos1abc...",
		"user@example.com",
		now,
	)

	hashed := record.GetHashedEmail()

	// Should be 64 hex characters
	assert.Len(t, hashed, 64, "GetHashedEmail() length should be 64")

	// Should be lowercase hex
	assert.Equal(t, strings.ToLower(hashed), hashed, "GetHashedEmail() should return lowercase hex")
}

func TestEmailVerificationRecord_GetDomain(t *testing.T) {
	now := time.Now()

	record := types.NewEmailVerificationRecord(
		"email-1",
		"cosmos1abc...",
		"user@example.com",
		now,
	)

	domain := record.GetDomain()
	assert.Equal(t, "example.com", domain, "GetDomain() should return correct domain")

	// Test with subdomain
	record2 := types.NewEmailVerificationRecord(
		"email-2",
		"cosmos1abc...",
		"user@mail.subdomain.example.org",
		now,
	)

	domain2 := record2.GetDomain()
	assert.Equal(t, "mail.subdomain.example.org", domain2, "GetDomain() should return full domain")
}

func TestEmailVerificationRecord_BelongsToDomain(t *testing.T) {
	now := time.Now()

	record := types.NewEmailVerificationRecord(
		"email-1",
		"cosmos1abc...",
		"user@example.com",
		now,
	)

	assert.True(t, record.BelongsToDomain("example.com"), "Email should belong to example.com")
	assert.False(t, record.BelongsToDomain("other.com"), "Email should not belong to other.com")

	// Test case insensitivity
	assert.True(t, record.BelongsToDomain("EXAMPLE.COM"), "BelongsToDomain should be case insensitive")
}
