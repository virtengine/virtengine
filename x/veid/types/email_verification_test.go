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
				"nonce-12345",
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
				"nonce-67890",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty verification ID",
			record: types.NewEmailVerificationRecord(
				"",
				"cosmos1abc...",
				"user@example.com",
				"nonce-12345",
				now,
			),
			wantErr: true,
		},
		{
			name: "invalid - empty account address",
			record: types.NewEmailVerificationRecord(
				"email-1",
				"",
				"user@example.com",
				"nonce-12345",
				now,
			),
			wantErr: true,
		},
		{
			name: "invalid - empty email hash",
			record: func() *types.EmailVerificationRecord {
				record := types.NewEmailVerificationRecord(
					"email-1",
					"cosmos1abc...",
					"user@example.com",
					"nonce-12345",
					now,
				)
				record.EmailHash = ""
				return record
			}(),
			wantErr: true,
		},
		{
			name: "invalid - email hash wrong length",
			record: func() *types.EmailVerificationRecord {
				record := types.NewEmailVerificationRecord(
					"email-1",
					"cosmos1abc...",
					"user@example.com",
					"nonce-12345",
					now,
				)
				record.EmailHash = "abc123"
				return record
			}(),
			wantErr: true,
		},
		{
			name: "invalid - empty nonce",
			record: func() *types.EmailVerificationRecord {
				record := types.NewEmailVerificationRecord(
					"email-1",
					"cosmos1abc...",
					"user@example.com",
					"nonce-12345",
					now,
				)
				record.Nonce = ""
				return record
			}(),
			wantErr: true,
		},
		{
			name: "invalid - invalid status",
			record: func() *types.EmailVerificationRecord {
				record := types.NewEmailVerificationRecord(
					"email-1",
					"cosmos1abc...",
					"user@example.com",
					"nonce-12345",
					now,
				)
				record.Status = types.EmailVerificationStatus("invalid")
				return record
			}(),
			wantErr: true,
		},
		{
			name: "invalid - zero created_at",
			record: func() *types.EmailVerificationRecord {
				record := types.NewEmailVerificationRecord(
					"email-1",
					"cosmos1abc...",
					"user@example.com",
					"nonce-12345",
					now,
				)
				record.CreatedAt = time.Time{}
				return record
			}(),
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

func TestEmailVerificationRecord_EvidenceRequiredForVerified(t *testing.T) {
	now := time.Now()
	record := types.NewEmailVerificationRecord(
		"email-verified",
		"cosmos1abc...",
		"user@example.com",
		"nonce-12345",
		now,
	)
	record.Status = types.EmailStatusVerified
	record.VerifiedAt = &now

	require.Error(t, record.Validate())

	record.EvidenceHash = strings.Repeat("a", 64)
	record.EvidenceStorageBackend = string(types.StorageBackendWaldur)
	record.EvidenceStorageRef = "vault://email/evidence"
	require.NoError(t, record.Validate())
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
			emailHash, domainHash := types.HashEmail(tt.email)

			// Hash should be 64 hex characters (SHA256)
			assert.Len(t, emailHash, 64, "HashEmail() length should be 64")
			assert.Len(t, domainHash, 64, "HashEmail() domain hash length should be 64")

			// Hash should be deterministic
			emailHash2, domainHash2 := types.HashEmail(tt.email)
			assert.Equal(t, emailHash, emailHash2, "HashEmail() should be deterministic")
			assert.Equal(t, domainHash, domainHash2, "HashEmail() domain hash should be deterministic")

			// Case insensitive (email domain part is case insensitive)
			emailHash3, domainHash3 := types.HashEmail(strings.ToLower(tt.email))
			assert.Equal(t, emailHash, emailHash3, "HashEmail() should be case insensitive")
			assert.Equal(t, domainHash, domainHash3, "HashEmail() domain hash should be case insensitive")

			// Different emails should produce different hashes
			emailHash4, domainHash4 := types.HashEmail(tt.email + "x")
			assert.NotEqual(t, emailHash, emailHash4, "HashEmail() should produce different hashes for different emails")
			assert.NotEqual(t, domainHash, domainHash4, "HashEmail() should produce different hashes for different emails")
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
	emailHash, _ := types.HashEmail("user@example.com")
	ttlSeconds := int64(3600)
	maxAttempts := uint32(3)

	tests := []struct {
		name      string
		challenge *types.EmailVerificationChallenge
		wantErr   bool
	}{
		{
			name: "valid challenge",
			challenge: types.NewEmailVerificationChallenge(
				"challenge-1",
				"cosmos1abc...",
				emailHash,
				"random-nonce-abc123",
				now,
				ttlSeconds,
				maxAttempts,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty challenge ID",
			challenge: &types.EmailVerificationChallenge{
				ChallengeID:    "",
				AccountAddress: "cosmos1abc...",
				EmailHash:      emailHash,
				Nonce:          "nonce123",
				CreatedAt:      now,
				ExpiresAt:      now.Add(time.Duration(ttlSeconds) * time.Second),
				MaxAttempts:    maxAttempts,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty account address",
			challenge: &types.EmailVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "",
				EmailHash:      emailHash,
				Nonce:          "nonce123",
				CreatedAt:      now,
				ExpiresAt:      now.Add(time.Duration(ttlSeconds) * time.Second),
				MaxAttempts:    maxAttempts,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty email hash",
			challenge: &types.EmailVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "cosmos1abc...",
				EmailHash:      "",
				Nonce:          "nonce123",
				CreatedAt:      now,
				ExpiresAt:      now.Add(time.Duration(ttlSeconds) * time.Second),
				MaxAttempts:    maxAttempts,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty nonce",
			challenge: &types.EmailVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "cosmos1abc...",
				EmailHash:      emailHash,
				Nonce:          "",
				CreatedAt:      now,
				ExpiresAt:      now.Add(time.Duration(ttlSeconds) * time.Second),
				MaxAttempts:    maxAttempts,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero created_at",
			challenge: &types.EmailVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "cosmos1abc...",
				EmailHash:      emailHash,
				Nonce:          "nonce123",
				CreatedAt:      time.Time{},
				ExpiresAt:      now.Add(time.Duration(ttlSeconds) * time.Second),
				MaxAttempts:    maxAttempts,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero expires_at",
			challenge: &types.EmailVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "cosmos1abc...",
				EmailHash:      emailHash,
				Nonce:          "nonce123",
				CreatedAt:      now,
				ExpiresAt:      time.Time{},
				MaxAttempts:    maxAttempts,
			},
			wantErr: true,
		},
		{
			name: "invalid - zero max attempts",
			challenge: &types.EmailVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "cosmos1abc...",
				EmailHash:      emailHash,
				Nonce:          "nonce123",
				CreatedAt:      now,
				ExpiresAt:      now.Add(time.Duration(ttlSeconds) * time.Second),
				MaxAttempts:    0,
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

func TestEmailVerificationChallenge_RecordAttempt(t *testing.T) {
	now := time.Now()
	emailHash, _ := types.HashEmail("user@example.com")
	maxAttempts := uint32(2)

	challenge := types.NewEmailVerificationChallenge(
		"challenge-1",
		"email-1",
		emailHash,
		"nonce123",
		now,
		3600,
		maxAttempts,
	)

	assert.Equal(t, uint32(0), challenge.Attempts, "Initial attempts should be 0")
	assert.True(t, challenge.CanAttempt(), "Challenge should allow attempts initially")

	challenge.RecordAttempt(now.Add(1*time.Minute), false)
	assert.Equal(t, uint32(1), challenge.Attempts, "After attempt, attempts should be 1")
	assert.Equal(t, types.EmailStatusPending, challenge.Status, "Status should remain pending before max attempts")
	assert.True(t, challenge.CanAttempt(), "Challenge should allow attempts before max attempts")

	challenge.RecordAttempt(now.Add(2*time.Minute), false)
	assert.Equal(t, uint32(2), challenge.Attempts, "After second attempt, attempts should be 2")
	assert.Equal(t, types.EmailStatusFailed, challenge.Status, "Status should be failed after max attempts")
	assert.False(t, challenge.CanAttempt(), "Challenge should not allow attempts after max attempts")
}

func TestEmailVerificationChallenge_VerifyNonce(t *testing.T) {
	now := time.Now()
	emailHash, _ := types.HashEmail("user@example.com")
	challenge := types.NewEmailVerificationChallenge(
		"challenge-1",
		"cosmos1abc...",
		emailHash,
		"nonce123",
		now,
		3600,
		3,
	)

	assert.True(t, challenge.VerifyNonce("nonce123"), "VerifyNonce should accept matching nonce")
	assert.False(t, challenge.VerifyNonce("wrong"), "VerifyNonce should reject non-matching nonce")

	challenge.RecordAttempt(now.Add(1*time.Minute), true)
	assert.False(t, challenge.VerifyNonce("nonce123"), "VerifyNonce should reject consumed nonce")
}

func TestUsedNonceRecord_NewAndIsNonceUsed(t *testing.T) {
	now := time.Now()
	record := types.NewUsedNonceRecord(
		"nonce-abc123",
		now,
		"cosmos1abc...",
		"verification-1",
		7,
	)

	assert.NotEmpty(t, record.NonceHash, "NonceHash should be set")
	assert.Equal(t, "cosmos1abc...", record.AccountAddress, "AccountAddress should be set")
	assert.Equal(t, "verification-1", record.VerificationID, "VerificationID should be set")
	assert.True(t, record.ExpiresAt.After(now), "ExpiresAt should be in the future")

	used := []types.UsedNonceRecord{*record}
	assert.True(t, types.IsNonceUsed(record.NonceHash, used), "Nonce should be recognized as used")
	assert.False(t, types.IsNonceUsed("missing", used), "Unknown nonce should not be marked used")
}

func TestEmailVerificationRecord_IsActive(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	record := types.NewEmailVerificationRecord(
		"email-1",
		"cosmos1abc...",
		"user@example.com",
		"nonce-12345",
		now,
	)
	record.MarkVerified(now, &expiresAt)
	assert.True(t, record.IsActive(), "Verified record should be active")

	expiredAt := now.Add(-1 * time.Hour)
	record.MarkVerified(now, &expiredAt)
	assert.False(t, record.IsActive(), "Expired record should not be active")

	record.Status = types.EmailStatusPending
	assert.False(t, record.IsActive(), "Pending record should not be active")
}
