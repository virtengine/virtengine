package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Scope Type Tests (VE-1006: Test Coverage)
// ============================================================================

func TestIsValidScopeType(t *testing.T) {
	tests := []struct {
		scopeType types.ScopeType
		expected  bool
	}{
		{types.ScopeTypeIDDocument, true},
		{types.ScopeTypeSelfie, true},
		{types.ScopeTypeFaceVideo, true},
		{types.ScopeTypeBiometric, true},
		{types.ScopeTypeBiometricHardware, true},
		{types.ScopeTypeDeviceAttestation, true},
		{types.ScopeTypeSSOMetadata, true},
		{types.ScopeTypeEmailProof, true},
		{types.ScopeTypeSMSProof, true},
		{types.ScopeTypeDomainVerify, true},
		{types.ScopeTypeADSSO, true},
		{types.ScopeType("invalid"), false},
		{types.ScopeType(""), false},
	}

	for _, tc := range tests {
		t.Run(string(tc.scopeType), func(t *testing.T) {
			assert.Equal(t, tc.expected, types.IsValidScopeType(tc.scopeType))
		})
	}
}

func TestAllScopeTypes(t *testing.T) {
	scopeTypes := types.AllScopeTypes()
	assert.Len(t, scopeTypes, 11) // All 11 scope types

	// Verify all types are present
	assert.Contains(t, scopeTypes, types.ScopeTypeIDDocument)
	assert.Contains(t, scopeTypes, types.ScopeTypeSelfie)
	assert.Contains(t, scopeTypes, types.ScopeTypeFaceVideo)
	assert.Contains(t, scopeTypes, types.ScopeTypeBiometric)
	assert.Contains(t, scopeTypes, types.ScopeTypeBiometricHardware)
	assert.Contains(t, scopeTypes, types.ScopeTypeDeviceAttestation)
	assert.Contains(t, scopeTypes, types.ScopeTypeSSOMetadata)
	assert.Contains(t, scopeTypes, types.ScopeTypeEmailProof)
	assert.Contains(t, scopeTypes, types.ScopeTypeSMSProof)
	assert.Contains(t, scopeTypes, types.ScopeTypeDomainVerify)
	assert.Contains(t, scopeTypes, types.ScopeTypeADSSO)
}

func TestScopeTypeWeight(t *testing.T) {
	tests := []struct {
		scopeType types.ScopeType
		expected  uint32
	}{
		{types.ScopeTypeIDDocument, 30}, // Highest weight
		{types.ScopeTypeFaceVideo, 25},
		{types.ScopeTypeSelfie, 20},
		{types.ScopeTypeBiometric, 20},
		{types.ScopeTypeBiometricHardware, 22},
		{types.ScopeTypeDeviceAttestation, 12},
		{types.ScopeTypeDomainVerify, 15},
		{types.ScopeTypeADSSO, 12},
		{types.ScopeTypeEmailProof, 10},
		{types.ScopeTypeSMSProof, 10},
		{types.ScopeTypeSSOMetadata, 5}, // Lowest weight
		{types.ScopeType("unknown"), 0},
	}

	for _, tc := range tests {
		t.Run(string(tc.scopeType), func(t *testing.T) {
			assert.Equal(t, tc.expected, types.ScopeTypeWeight(tc.scopeType))
		})
	}
}

func TestScopeTypeDescription(t *testing.T) {
	tests := []struct {
		scopeType types.ScopeType
		contains  string
	}{
		{types.ScopeTypeIDDocument, "Government-issued"},
		{types.ScopeTypeSelfie, "Selfie photo"},
		{types.ScopeTypeFaceVideo, "liveness"},
		{types.ScopeTypeBiometric, "Biometric"},
		{types.ScopeTypeBiometricHardware, "hardware"},
		{types.ScopeTypeDeviceAttestation, "Device integrity"},
		{types.ScopeTypeSSOMetadata, "SSO provider"},
		{types.ScopeTypeEmailProof, "Email"},
		{types.ScopeTypeSMSProof, "Phone number"},
		{types.ScopeTypeDomainVerify, "Domain ownership"},
		{types.ScopeTypeADSSO, "Active Directory"},
		{types.ScopeType("unknown"), "Unknown"},
	}

	for _, tc := range tests {
		t.Run(string(tc.scopeType), func(t *testing.T) {
			desc := types.ScopeTypeDescription(tc.scopeType)
			assert.Contains(t, desc, tc.contains)
		})
	}
}

// ============================================================================
// Identity Scope Tests
// ============================================================================

// createValidEnvelope creates a valid EncryptedPayloadEnvelope for testing
func createValidEnvelope() encryptiontypes.EncryptedPayloadEnvelope {
	return encryptiontypes.EncryptedPayloadEnvelope{
		Version:          1,
		AlgorithmID:      encryptiontypes.AlgorithmX25519XSalsa20Poly1305,
		AlgorithmVersion: encryptiontypes.AlgorithmVersionV1,
		RecipientKeyIDs:  []string{"key-fingerprint-001"},
		Nonce:            make([]byte, encryptiontypes.XSalsa20NonceSize),
		Ciphertext:       []byte("encrypted-data-placeholder"),
		SenderSignature:  []byte("sender-signature"),
		SenderPubKey:     make([]byte, encryptiontypes.X25519PublicKeySize),
	}
}

// createValidUploadMetadata creates a valid UploadMetadata for testing
func createValidUploadMetadata() types.UploadMetadata {
	return types.UploadMetadata{
		Salt:              make([]byte, 32),
		ClientID:          "approved-client-001",
		ClientSignature:   []byte("client-signature"),
		UserSignature:     []byte("user-signature"),
		PayloadHash:       make([]byte, 32),
		DeviceFingerprint: "device-fingerprint-hash-abc123",
	}
}

func TestNewIdentityScope(t *testing.T) {
	now := time.Now()
	envelope := createValidEnvelope()
	metadata := createValidUploadMetadata()

	scope := types.NewIdentityScope(
		"scope-001",
		types.ScopeTypeIDDocument,
		envelope,
		metadata,
		now,
	)

	require.NotNil(t, scope)
	assert.Equal(t, "scope-001", scope.ScopeID)
	assert.Equal(t, types.ScopeTypeIDDocument, scope.ScopeType)
	assert.Equal(t, types.ScopeSchemaVersion, scope.Version)
	assert.Equal(t, types.VerificationStatusPending, scope.Status)
	assert.Equal(t, now, scope.UploadedAt)
	assert.False(t, scope.Revoked)
	assert.Nil(t, scope.VerifiedAt)
	assert.Nil(t, scope.ExpiresAt)
}

func TestIdentityScope_Validate(t *testing.T) {
	now := time.Now()
	envelope := createValidEnvelope()
	metadata := createValidUploadMetadata()

	tests := []struct {
		name        string
		modify      func(*types.IdentityScope)
		expectError bool
		errContains string
	}{
		{
			name:        "valid scope",
			modify:      func(s *types.IdentityScope) {},
			expectError: false,
		},
		{
			name: "empty scope ID",
			modify: func(s *types.IdentityScope) {
				s.ScopeID = ""
			},
			expectError: true,
			errContains: "scope_id cannot be empty",
		},
		{
			name: "invalid scope type",
			modify: func(s *types.IdentityScope) {
				s.ScopeType = types.ScopeType("invalid")
			},
			expectError: true,
			errContains: "invalid scope type",
		},
		{
			name: "zero version",
			modify: func(s *types.IdentityScope) {
				s.Version = 0
			},
			expectError: true,
			errContains: "unsupported version",
		},
		{
			name: "version too high",
			modify: func(s *types.IdentityScope) {
				s.Version = types.ScopeSchemaVersion + 1
			},
			expectError: true,
			errContains: "unsupported version",
		},
		{
			name: "invalid verification status",
			modify: func(s *types.IdentityScope) {
				s.Status = types.VerificationStatus("invalid")
			},
			expectError: true,
			errContains: "invalid status",
		},
		{
			name: "zero uploaded_at",
			modify: func(s *types.IdentityScope) {
				s.UploadedAt = time.Time{}
			},
			expectError: true,
			errContains: "uploaded_at cannot be zero",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			scope := types.NewIdentityScope("scope-001", types.ScopeTypeIDDocument, envelope, metadata, now)
			tc.modify(scope)

			err := scope.Validate()
			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestIdentityScope_IsActive(t *testing.T) {
	now := time.Now()
	envelope := createValidEnvelope()
	metadata := createValidUploadMetadata()

	t.Run("active scope", func(t *testing.T) {
		scope := types.NewIdentityScope("scope-001", types.ScopeTypeSelfie, envelope, metadata, now)
		assert.True(t, scope.IsActive())
	})

	t.Run("revoked scope", func(t *testing.T) {
		scope := types.NewIdentityScope("scope-001", types.ScopeTypeSelfie, envelope, metadata, now)
		scope.Revoked = true
		assert.False(t, scope.IsActive())
	})

	t.Run("expired scope", func(t *testing.T) {
		scope := types.NewIdentityScope("scope-001", types.ScopeTypeSelfie, envelope, metadata, now)
		pastTime := now.Add(-24 * time.Hour)
		scope.ExpiresAt = &pastTime
		assert.False(t, scope.IsActive())
	})

	t.Run("not yet expired scope", func(t *testing.T) {
		scope := types.NewIdentityScope("scope-001", types.ScopeTypeSelfie, envelope, metadata, now)
		futureTime := now.Add(24 * time.Hour)
		scope.ExpiresAt = &futureTime
		assert.True(t, scope.IsActive())
	})
}

func TestIdentityScope_IsVerified(t *testing.T) {
	now := time.Now()
	envelope := createValidEnvelope()
	metadata := createValidUploadMetadata()

	t.Run("verified scope", func(t *testing.T) {
		scope := types.NewIdentityScope("scope-001", types.ScopeTypeSelfie, envelope, metadata, now)
		scope.Status = types.VerificationStatusVerified
		scope.VerifiedAt = &now
		assert.True(t, scope.IsVerified())
	})

	t.Run("verified status but no verified_at", func(t *testing.T) {
		scope := types.NewIdentityScope("scope-001", types.ScopeTypeSelfie, envelope, metadata, now)
		scope.Status = types.VerificationStatusVerified
		scope.VerifiedAt = nil
		assert.False(t, scope.IsVerified())
	})

	t.Run("pending scope", func(t *testing.T) {
		scope := types.NewIdentityScope("scope-001", types.ScopeTypeSelfie, envelope, metadata, now)
		assert.False(t, scope.IsVerified())
	})
}

func TestIdentityScope_CanBeVerified(t *testing.T) {
	now := time.Now()
	envelope := createValidEnvelope()
	metadata := createValidUploadMetadata()

	tests := []struct {
		name     string
		status   types.VerificationStatus
		revoked  bool
		expected bool
	}{
		{"pending scope", types.VerificationStatusPending, false, true},
		{"in progress scope", types.VerificationStatusInProgress, false, true},
		{"verified scope", types.VerificationStatusVerified, false, false},
		{"rejected scope", types.VerificationStatusRejected, false, false},
		{"expired scope", types.VerificationStatusExpired, false, false},
		{"revoked pending scope", types.VerificationStatusPending, true, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			scope := types.NewIdentityScope("scope-001", types.ScopeTypeSelfie, envelope, metadata, now)
			scope.Status = tc.status
			scope.Revoked = tc.revoked
			assert.Equal(t, tc.expected, scope.CanBeVerified())
		})
	}
}

func TestIdentityScope_String(t *testing.T) {
	now := time.Now()
	envelope := createValidEnvelope()
	metadata := createValidUploadMetadata()

	scope := types.NewIdentityScope("scope-001", types.ScopeTypeIDDocument, envelope, metadata, now)
	str := scope.String()

	assert.Contains(t, str, "scope-001")
	assert.Contains(t, str, "id_document")
	assert.Contains(t, str, "pending")
}

// ============================================================================
// Scope Ref Tests
// ============================================================================

func TestNewScopeRef(t *testing.T) {
	now := time.Now()
	envelope := createValidEnvelope()
	metadata := createValidUploadMetadata()

	scope := types.NewIdentityScope("scope-001", types.ScopeTypeFaceVideo, envelope, metadata, now)
	scope.Status = types.VerificationStatusVerified

	ref := types.NewScopeRef(scope)

	assert.Equal(t, "scope-001", ref.ScopeID)
	assert.Equal(t, types.ScopeTypeFaceVideo, ref.ScopeType)
	assert.Equal(t, types.VerificationStatusVerified, ref.Status)
	assert.Equal(t, now.Unix(), ref.UploadedAt)
}
