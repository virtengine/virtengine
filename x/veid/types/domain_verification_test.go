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
// Domain Verification Tests (VE-223: Domain Verification Scope)
// ============================================================================

func TestDomainVerificationRecord_Validate(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		record  *types.DomainVerificationRecord
		wantErr bool
	}{
		{
			name: "valid DNS TXT verification",
			record: types.NewDomainVerificationRecord(
				"domain-1",
				"cosmos1abc...",
				"example.com",
				types.DomainVerifyDNSTXT,
				"challenge-token-1",
				now,
			),
			wantErr: false,
		},
		{
			name: "valid HTTP well-known verification",
			record: types.NewDomainVerificationRecord(
				"domain-2",
				"cosmos1abc...",
				"subdomain.example.org",
				types.DomainVerifyHTTPWellKnown,
				"challenge-token-2",
				now,
			),
			wantErr: false,
		},
		{
			name: "valid email admin verification",
			record: types.NewDomainVerificationRecord(
				"domain-3",
				"cosmos1abc...",
				"example.io",
				types.DomainVerifyEmailAdmin,
				"challenge-token-3",
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty domain ID",
			record: func() *types.DomainVerificationRecord {
				record := types.NewDomainVerificationRecord(
					"domain-1",
					"cosmos1abc...",
					"example.com",
					types.DomainVerifyDNSTXT,
					"challenge-token-1",
					now,
				)
				record.VerificationID = ""
				return record
			}(),
			wantErr: true,
		},
		{
			name: "invalid - empty owner",
			record: func() *types.DomainVerificationRecord {
				record := types.NewDomainVerificationRecord(
					"domain-1",
					"cosmos1abc...",
					"example.com",
					types.DomainVerifyDNSTXT,
					"challenge-token-1",
					now,
				)
				record.AccountAddress = ""
				return record
			}(),
			wantErr: true,
		},
		{
			name: "invalid - empty domain",
			record: func() *types.DomainVerificationRecord {
				record := types.NewDomainVerificationRecord(
					"domain-1",
					"cosmos1abc...",
					"example.com",
					types.DomainVerifyDNSTXT,
					"challenge-token-1",
					now,
				)
				record.Domain = ""
				return record
			}(),
			wantErr: true,
		},
		{
			name: "invalid - empty domain hash",
			record: func() *types.DomainVerificationRecord {
				record := types.NewDomainVerificationRecord(
					"domain-1",
					"cosmos1abc...",
					"example.com",
					types.DomainVerifyDNSTXT,
					"challenge-token-1",
					now,
				)
				record.DomainHash = ""
				return record
			}(),
			wantErr: true,
		},
		{
			name: "invalid - invalid method",
			record: func() *types.DomainVerificationRecord {
				record := types.NewDomainVerificationRecord(
					"domain-1",
					"cosmos1abc...",
					"example.com",
					types.DomainVerifyDNSTXT,
					"challenge-token-1",
					now,
				)
				record.Method = types.DomainVerificationMethod("invalid_method")
				return record
			}(),
			wantErr: true,
		},
		{
			name: "invalid - empty challenge token",
			record: func() *types.DomainVerificationRecord {
				record := types.NewDomainVerificationRecord(
					"domain-1",
					"cosmos1abc...",
					"example.com",
					types.DomainVerifyDNSTXT,
					"challenge-token-1",
					now,
				)
				record.ChallengeToken = ""
				return record
			}(),
			wantErr: true,
		},
		{
			name: "invalid - invalid status",
			record: func() *types.DomainVerificationRecord {
				record := types.NewDomainVerificationRecord(
					"domain-1",
					"cosmos1abc...",
					"example.com",
					types.DomainVerifyDNSTXT,
					"challenge-token-1",
					now,
				)
				record.Status = types.DomainVerificationStatus("invalid")
				return record
			}(),
			wantErr: true,
		},
		{
			name: "invalid - zero created_at",
			record: func() *types.DomainVerificationRecord {
				record := types.NewDomainVerificationRecord(
					"domain-1",
					"cosmos1abc...",
					"example.com",
					types.DomainVerifyDNSTXT,
					"challenge-token-1",
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

func TestHashDomain(t *testing.T) {
	tests := []struct {
		name   string
		domain string
	}{
		{name: "simple domain", domain: "example.com"},
		{name: "subdomain", domain: "subdomain.example.com"},
		{name: "multi-level subdomain", domain: "deep.subdomain.example.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := types.HashDomain(tt.domain)

			// Hash should be 64 hex characters (SHA256)
			assert.Len(t, hash, 64, "HashDomain() length should be 64")

			// Hash should be deterministic
			hash2 := types.HashDomain(tt.domain)
			assert.Equal(t, hash, hash2, "HashDomain() should be deterministic")

			// Case insensitive - lowercase
			hash3 := types.HashDomain(strings.ToUpper(tt.domain))
			assert.Equal(t, hash, hash3, "HashDomain() should be case insensitive")

			// Different domains should produce different hashes
			hash4 := types.HashDomain(tt.domain + "x")
			assert.NotEqual(t, hash, hash4, "HashDomain() should produce different hashes for different domains")
		})
	}
}

func TestDomainVerificationMethods(t *testing.T) {
	// Test all methods are valid
	for _, method := range types.AllDomainVerificationMethods() {
		assert.True(t, types.IsValidDomainVerificationMethod(method), "AllDomainVerificationMethods returned invalid method: %s", method)
	}

	// Test invalid method
	assert.False(t, types.IsValidDomainVerificationMethod("invalid"), "IsValidDomainVerificationMethod should return false for invalid method")
}

func TestDomainVerificationStatuses(t *testing.T) {
	// Test all statuses are valid
	for _, status := range types.AllDomainVerificationStatuses() {
		assert.True(t, types.IsValidDomainVerificationStatus(status), "AllDomainVerificationStatuses returned invalid status: %s", status)
	}

	// Test invalid status
	assert.False(t, types.IsValidDomainVerificationStatus("invalid"), "IsValidDomainVerificationStatus should return false for invalid status")
}

func TestDomainVerificationChallenge_Validate(t *testing.T) {
	now := time.Now()
	ttlSeconds := int64(86400) // 24 hours

	tests := []struct {
		name      string
		challenge *types.DomainVerificationChallenge
		wantErr   bool
	}{
		{
			name: "valid DNS TXT challenge",
			challenge: types.NewDomainVerificationChallenge(
				"challenge-1",
				"cosmos1abc...",
				"example.com",
				types.DomainVerifyDNSTXT,
				"abc123def456",
				now,
				ttlSeconds,
			),
			wantErr: false,
		},
		{
			name: "valid HTTP well-known challenge",
			challenge: types.NewDomainVerificationChallenge(
				"challenge-2",
				"cosmos1def...",
				"example.org",
				types.DomainVerifyHTTPWellKnown,
				"xyz789ghi012",
				now,
				ttlSeconds,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty challenge ID",
			challenge: &types.DomainVerificationChallenge{
				ChallengeID:    "",
				AccountAddress: "cosmos1abc...",
				Domain:         "example.com",
				Method:         types.DomainVerifyDNSTXT,
				Token:          "abc123",
				ExpectedValue:  "virtengine-verification=abc123",
				CreatedAt:      now,
				ExpiresAt:      now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - empty account address",
			challenge: &types.DomainVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "",
				Domain:         "example.com",
				Method:         types.DomainVerifyDNSTXT,
				Token:          "abc123",
				ExpectedValue:  "virtengine-verification=abc123",
				CreatedAt:      now,
				ExpiresAt:      now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - empty domain",
			challenge: &types.DomainVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "cosmos1abc...",
				Domain:         "",
				Method:         types.DomainVerifyDNSTXT,
				Token:          "abc123",
				ExpectedValue:  "virtengine-verification=abc123",
				CreatedAt:      now,
				ExpiresAt:      now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid method",
			challenge: &types.DomainVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "cosmos1abc...",
				Domain:         "example.com",
				Method:         types.DomainVerificationMethod("invalid_method"),
				Token:          "abc123",
				ExpectedValue:  "virtengine-verification=abc123",
				CreatedAt:      now,
				ExpiresAt:      now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - empty token",
			challenge: &types.DomainVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "cosmos1abc...",
				Domain:         "example.com",
				Method:         types.DomainVerifyDNSTXT,
				Token:          "",
				ExpectedValue:  "virtengine-verification=",
				CreatedAt:      now,
				ExpiresAt:      now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - zero created_at",
			challenge: &types.DomainVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "cosmos1abc...",
				Domain:         "example.com",
				Method:         types.DomainVerifyDNSTXT,
				Token:          "abc123",
				ExpectedValue:  "virtengine-verification=abc123",
				CreatedAt:      time.Time{},
				ExpiresAt:      now.Add(24 * time.Hour),
			},
			wantErr: true,
		},
		{
			name: "invalid - zero expires_at",
			challenge: &types.DomainVerificationChallenge{
				ChallengeID:    "challenge-1",
				AccountAddress: "cosmos1abc...",
				Domain:         "example.com",
				Method:         types.DomainVerifyDNSTXT,
				Token:          "abc123",
				ExpectedValue:  "virtengine-verification=abc123",
				CreatedAt:      now,
				ExpiresAt:      time.Time{},
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

func TestDomainVerificationChallenge_IsExpired(t *testing.T) {
	now := time.Now()

	notExpired := &types.DomainVerificationChallenge{ExpiresAt: now.Add(24 * time.Hour)}
	assert.False(t, notExpired.IsExpired(now), "Challenge should not be expired")

	expired := &types.DomainVerificationChallenge{ExpiresAt: now.Add(-1 * time.Hour)}
	assert.True(t, expired.IsExpired(now), "Challenge should be expired")
}

func TestDomainVerificationChallenge_Paths(t *testing.T) {
	now := time.Now()
	challenge := types.NewDomainVerificationChallenge(
		"challenge-1",
		"cosmos1abc...",
		"example.com",
		types.DomainVerifyDNSTXT,
		"abc123",
		now,
		3600,
	)

	assert.Equal(t, "_virtengine-verification.example.com", challenge.GetDNSRecordName(), "GetDNSRecordName should return expected name")
	assert.Equal(t, "/.well-known/virtengine-verification/abc123", challenge.GetHTTPWellKnownPath(), "GetHTTPWellKnownPath should return expected path")
}

func TestDomainVerificationRecord_StatusTransitions(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)

	record := types.NewDomainVerificationRecord(
		"domain-1",
		"cosmos1abc...",
		"example.com",
		types.DomainVerifyDNSTXT,
		"token-123",
		now,
	)

	record.MarkVerified(now, &expiresAt)
	assert.True(t, record.IsActive(), "Verified record should be active")

	record.MarkRevoked(now.Add(1*time.Hour), "test")
	assert.False(t, record.IsActive(), "Revoked record should not be active")
	assert.Equal(t, types.DomainStatusRevoked, record.Status, "Status should be revoked")
}
