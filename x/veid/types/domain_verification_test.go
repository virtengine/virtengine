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
				types.DomainMethodDNSTXT,
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
				types.DomainMethodHTTPWellKnown,
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
				types.DomainMethodEmailAdmin,
				now,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty domain ID",
			record: &types.DomainVerificationRecord{
				Version:  types.DomainVerificationVersion,
				DomainID: "",
				Owner:    "cosmos1abc...",
				Domain:   "example.com",
				Method:   types.DomainMethodDNSTXT,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty owner",
			record: &types.DomainVerificationRecord{
				Version:  types.DomainVerificationVersion,
				DomainID: "domain-1",
				Owner:    "",
				Domain:   "example.com",
				Method:   types.DomainMethodDNSTXT,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty domain",
			record: &types.DomainVerificationRecord{
				Version:  types.DomainVerificationVersion,
				DomainID: "domain-1",
				Owner:    "cosmos1abc...",
				Domain:   "",
				Method:   types.DomainMethodDNSTXT,
			},
			wantErr: true,
		},
		{
			name: "invalid - invalid method",
			record: &types.DomainVerificationRecord{
				Version:  types.DomainVerificationVersion,
				DomainID: "domain-1",
				Owner:    "cosmos1abc...",
				Domain:   "example.com",
				Method:   "invalid_method",
			},
			wantErr: true,
		},
		{
			name: "invalid - domain with protocol",
			record: &types.DomainVerificationRecord{
				Version:  types.DomainVerificationVersion,
				DomainID: "domain-1",
				Owner:    "cosmos1abc...",
				Domain:   "https://example.com",
				Method:   types.DomainMethodDNSTXT,
			},
			wantErr: true,
		},
		{
			name: "invalid - domain with path",
			record: &types.DomainVerificationRecord{
				Version:  types.DomainVerificationVersion,
				DomainID: "domain-1",
				Owner:    "cosmos1abc...",
				Domain:   "example.com/path",
				Method:   types.DomainMethodDNSTXT,
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
	future := now.Add(24 * time.Hour)

	tests := []struct {
		name      string
		challenge *types.DomainVerificationChallenge
		wantErr   bool
	}{
		{
			name: "valid DNS TXT challenge",
			challenge: types.NewDomainVerificationChallenge(
				"challenge-1",
				"domain-1",
				types.DomainMethodDNSTXT,
				"_virtengine-verify.example.com",
				"virtengine-verify=abc123def456",
				now,
				future,
			),
			wantErr: false,
		},
		{
			name: "valid HTTP well-known challenge",
			challenge: types.NewDomainVerificationChallenge(
				"challenge-2",
				"domain-2",
				types.DomainMethodHTTPWellKnown,
				"https://example.com/.well-known/virtengine-verify.txt",
				"abc123def456",
				now,
				future,
			),
			wantErr: false,
		},
		{
			name: "invalid - empty challenge ID",
			challenge: &types.DomainVerificationChallenge{
				Version:       types.DomainChallengeVersion,
				ChallengeID:   "",
				DomainID:      "domain-1",
				Method:        types.DomainMethodDNSTXT,
				ChallengeKey:  "_virtengine-verify.example.com",
				ExpectedValue: "abc123",
				CreatedAt:     now,
				ExpiresAt:     future,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty challenge key",
			challenge: &types.DomainVerificationChallenge{
				Version:       types.DomainChallengeVersion,
				ChallengeID:   "challenge-1",
				DomainID:      "domain-1",
				Method:        types.DomainMethodDNSTXT,
				ChallengeKey:  "",
				ExpectedValue: "abc123",
				CreatedAt:     now,
				ExpiresAt:     future,
			},
			wantErr: true,
		},
		{
			name: "invalid - empty expected value",
			challenge: &types.DomainVerificationChallenge{
				Version:       types.DomainChallengeVersion,
				ChallengeID:   "challenge-1",
				DomainID:      "domain-1",
				Method:        types.DomainMethodDNSTXT,
				ChallengeKey:  "_virtengine-verify.example.com",
				ExpectedValue: "",
				CreatedAt:     now,
				ExpiresAt:     future,
			},
			wantErr: true,
		},
		{
			name: "invalid - expired challenge",
			challenge: &types.DomainVerificationChallenge{
				Version:       types.DomainChallengeVersion,
				ChallengeID:   "challenge-1",
				DomainID:      "domain-1",
				Method:        types.DomainMethodDNSTXT,
				ChallengeKey:  "_virtengine-verify.example.com",
				ExpectedValue: "abc123",
				CreatedAt:     now.Add(-48 * time.Hour),
				ExpiresAt:     now.Add(-24 * time.Hour),
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

func TestDomainVerificationRecord_IsVerified(t *testing.T) {
	now := time.Now()

	verified := &types.DomainVerificationRecord{Status: types.DomainStatusVerified, VerifiedAt: &now}
	assert.True(t, verified.IsVerified(), "Record should be verified")

	pending := &types.DomainVerificationRecord{Status: types.DomainStatusPending}
	assert.False(t, pending.IsVerified(), "Pending record should not be verified")

	expired := &types.DomainVerificationRecord{Status: types.DomainStatusExpired}
	assert.False(t, expired.IsVerified(), "Expired record should not be verified")
}

func TestDomainVerificationRecord_GetHashedDomain(t *testing.T) {
	now := time.Now()

	record := types.NewDomainVerificationRecord(
		"domain-1",
		"cosmos1abc...",
		"example.com",
		types.DomainMethodDNSTXT,
		now,
	)

	hashed := record.GetHashedDomain()

	// Should be 64 hex characters
	assert.Len(t, hashed, 64, "GetHashedDomain() length should be 64")

	// Should be lowercase hex
	assert.Equal(t, strings.ToLower(hashed), hashed, "GetHashedDomain() should return lowercase hex")
}

func TestDomainVerificationRecord_CanBeReverified(t *testing.T) {
	now := time.Now()
	reverifyAt := now.Add(30 * 24 * time.Hour)

	canReverify := &types.DomainVerificationRecord{
		Status:               types.DomainStatusVerified,
		NextReverificationAt: &reverifyAt,
	}

	assert.True(t, canReverify.CanBeReverified(reverifyAt.Add(time.Hour)), "Record should be eligible for reverification after the reverification date")
	assert.False(t, canReverify.CanBeReverified(reverifyAt.Add(-time.Hour)), "Record should not be eligible for reverification before the reverification date")

	// Pending records cannot be reverified
	pending := &types.DomainVerificationRecord{Status: types.DomainStatusPending}
	assert.False(t, pending.CanBeReverified(now), "Pending records should not be eligible for reverification")
}
