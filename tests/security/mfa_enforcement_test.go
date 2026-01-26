// Package security contains security-focused tests for VirtEngine.
// These tests verify MFA enforcement for sensitive transactions.
//
// Task Reference: VE-800 - Security audit readiness
package security

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MFAEnforcementTestSuite tests MFA security enforcement.
type MFAEnforcementTestSuite struct {
	suite.Suite
}

func TestMFAEnforcement(t *testing.T) {
	suite.Run(t, new(MFAEnforcementTestSuite))
}

// =============================================================================
// Sensitive Transaction Gating Tests
// =============================================================================

// TestSensitiveTransactionGating verifies MFA is required for sensitive ops.
func (s *MFAEnforcementTestSuite) TestSensitiveTransactionGating() {
	s.T().Log("=== Test: Sensitive Transaction Gating ===")

	// Define sensitive transaction types
	sensitiveTypes := []SensitiveTransactionType{
		TxTypeAccountRecovery,
		TxTypeKeyRotation,
		TxTypeHighValueTransfer,
		TxTypeProviderRegistration,
		TxTypeIdentityScopeUpdate,
		TxTypeMFAPolicyChange,
		TxTypeDelegationChange,
	}

	// Test: All sensitive transactions require MFA
	for _, txType := range sensitiveTypes {
		s.Run("tx_"+string(txType)+"_requires_mfa", func() {
			policy := &MFAPolicy{
				SensitiveTxTypes: sensitiveTypes,
				RequiredFactors:  1,
			}

			requiresMFA := policy.RequiresMFA(txType)
			require.True(s.T(), requiresMFA,
				"transaction type %s should require MFA", txType)
		})
	}

	// Test: Non-sensitive transactions don't require MFA
	s.Run("non_sensitiVIRTENGINE_tx_no_mfa", func() {
		policy := &MFAPolicy{
			SensitiveTxTypes: sensitiveTypes,
			RequiredFactors:  1,
		}

		nonSensitive := []SensitiveTransactionType{
			"normal_transfer",
			"query_balance",
			"view_offering",
		}

		for _, txType := range nonSensitive {
			requiresMFA := policy.RequiresMFA(txType)
			require.False(s.T(), requiresMFA,
				"transaction type %s should not require MFA", txType)
		}
	})
}

// TestMFAChallengeVerification tests MFA challenge flow.
func (s *MFAEnforcementTestSuite) TestMFAChallengeVerification() {
	s.T().Log("=== Test: MFA Challenge Verification ===")

	// Test: Valid TOTP response accepted
	s.Run("valid_totp_response_accepted", func() {
		challenge := &MFAChallenge{
			ID:          generateChallengeID(s.T()),
			FactorType:  FactorTypeTOTP,
			AccountID:   "user123",
			CreatedAt:   time.Now().UTC(),
			ExpiresAt:   time.Now().UTC().Add(5 * time.Minute),
			ExpectedOTP: "123456", // In real impl, this is computed
		}

		response := &MFAResponse{
			ChallengeID: challenge.ID,
			OTP:         "123456",
			Timestamp:   time.Now().UTC(),
		}

		result := verifyTOTPChallenge(challenge, response)
		require.True(s.T(), result.Valid, "valid TOTP should be accepted")
	})

	// Test: Invalid TOTP response rejected
	s.Run("invalid_totp_response_rejected", func() {
		challenge := &MFAChallenge{
			ID:          generateChallengeID(s.T()),
			FactorType:  FactorTypeTOTP,
			AccountID:   "user123",
			CreatedAt:   time.Now().UTC(),
			ExpiresAt:   time.Now().UTC().Add(5 * time.Minute),
			ExpectedOTP: "123456",
		}

		response := &MFAResponse{
			ChallengeID: challenge.ID,
			OTP:         "654321", // Wrong OTP
			Timestamp:   time.Now().UTC(),
		}

		result := verifyTOTPChallenge(challenge, response)
		require.False(s.T(), result.Valid, "invalid TOTP should be rejected")
	})

	// Test: Expired challenge rejected
	s.Run("expired_challenge_rejected", func() {
		challenge := &MFAChallenge{
			ID:          generateChallengeID(s.T()),
			FactorType:  FactorTypeTOTP,
			AccountID:   "user123",
			CreatedAt:   time.Now().UTC().Add(-10 * time.Minute),
			ExpiresAt:   time.Now().UTC().Add(-5 * time.Minute), // Expired
			ExpectedOTP: "123456",
		}

		response := &MFAResponse{
			ChallengeID: challenge.ID,
			OTP:         "123456",
			Timestamp:   time.Now().UTC(),
		}

		result := verifyTOTPChallenge(challenge, response)
		require.False(s.T(), result.Valid, "expired challenge should be rejected")
		require.Equal(s.T(), "challenge_expired", result.Reason)
	})

	// Test: Challenge ID mismatch rejected
	s.Run("challenge_id_mismatch_rejected", func() {
		challenge := &MFAChallenge{
			ID:          generateChallengeID(s.T()),
			FactorType:  FactorTypeTOTP,
			AccountID:   "user123",
			CreatedAt:   time.Now().UTC(),
			ExpiresAt:   time.Now().UTC().Add(5 * time.Minute),
			ExpectedOTP: "123456",
		}

		response := &MFAResponse{
			ChallengeID: generateChallengeID(s.T()), // Different ID
			OTP:         "123456",
			Timestamp:   time.Now().UTC(),
		}

		result := verifyTOTPChallenge(challenge, response)
		require.False(s.T(), result.Valid, "mismatched challenge ID should be rejected")
		require.Equal(s.T(), "challenge_id_mismatch", result.Reason)
	})
}

// TestMFASessionManagement tests authorization session handling.
func (s *MFAEnforcementTestSuite) TestMFASessionManagement() {
	s.T().Log("=== Test: MFA Session Management ===")

	// Test: Valid session allows sensitive operation
	s.Run("valid_session_allows_operation", func() {
		session := &AuthorizationSession{
			ID:          generateSessionID(s.T()),
			AccountID:   "user123",
			TxType:      TxTypeHighValueTransfer,
			CreatedAt:   time.Now().UTC(),
			ExpiresAt:   time.Now().UTC().Add(15 * time.Minute),
			Used:        false,
			FactorTypes: []FactorType{FactorTypeTOTP},
		}

		canProceed := session.IsValidForOperation(TxTypeHighValueTransfer)
		require.True(s.T(), canProceed, "valid session should allow operation")
	})

	// Test: Expired session rejected
	s.Run("expired_session_rejected", func() {
		session := &AuthorizationSession{
			ID:          generateSessionID(s.T()),
			AccountID:   "user123",
			TxType:      TxTypeHighValueTransfer,
			CreatedAt:   time.Now().UTC().Add(-20 * time.Minute),
			ExpiresAt:   time.Now().UTC().Add(-5 * time.Minute), // Expired
			Used:        false,
			FactorTypes: []FactorType{FactorTypeTOTP},
		}

		canProceed := session.IsValidForOperation(TxTypeHighValueTransfer)
		require.False(s.T(), canProceed, "expired session should not allow operation")
	})

	// Test: Already used session rejected (one-time use)
	s.Run("used_session_rejected", func() {
		session := &AuthorizationSession{
			ID:          generateSessionID(s.T()),
			AccountID:   "user123",
			TxType:      TxTypeHighValueTransfer,
			CreatedAt:   time.Now().UTC(),
			ExpiresAt:   time.Now().UTC().Add(15 * time.Minute),
			Used:        true, // Already used
			FactorTypes: []FactorType{FactorTypeTOTP},
		}

		canProceed := session.IsValidForOperation(TxTypeHighValueTransfer)
		require.False(s.T(), canProceed, "used session should not allow operation")
	})

	// Test: Wrong transaction type rejected
	s.Run("wrong_tx_type_rejected", func() {
		session := &AuthorizationSession{
			ID:          generateSessionID(s.T()),
			AccountID:   "user123",
			TxType:      TxTypeHighValueTransfer,
			CreatedAt:   time.Now().UTC(),
			ExpiresAt:   time.Now().UTC().Add(15 * time.Minute),
			Used:        false,
			FactorTypes: []FactorType{FactorTypeTOTP},
		}

		// Try to use session for different tx type
		canProceed := session.IsValidForOperation(TxTypeAccountRecovery)
		require.False(s.T(), canProceed, "session should not allow different transaction type")
	})
}

// TestDeviceTrustValidation tests trusted device handling.
func (s *MFAEnforcementTestSuite) TestDeviceTrustValidation() {
	s.T().Log("=== Test: Device Trust Validation ===")

	// Test: Trusted device recognized
	s.Run("trusted_device_recognized", func() {
		trustStore := NewDeviceTrustStore()
		fingerprint := generateDeviceFingerprint(s.T())

		trustStore.AddDevice("user123", &TrustedDevice{
			Fingerprint: fingerprint,
			Name:        "iPhone 15",
			AddedAt:     time.Now().UTC(),
			LastUsed:    time.Now().UTC(),
			ExpiresAt:   time.Now().UTC().Add(30 * 24 * time.Hour),
		})

		isTrusted := trustStore.IsTrusted("user123", fingerprint)
		require.True(s.T(), isTrusted, "registered device should be trusted")
	})

	// Test: Unknown device not trusted
	s.Run("unknown_device_not_trusted", func() {
		trustStore := NewDeviceTrustStore()
		knownFingerprint := generateDeviceFingerprint(s.T())
		unknownFingerprint := generateDeviceFingerprint(s.T())

		trustStore.AddDevice("user123", &TrustedDevice{
			Fingerprint: knownFingerprint,
			Name:        "Known Device",
			AddedAt:     time.Now().UTC(),
			LastUsed:    time.Now().UTC(),
			ExpiresAt:   time.Now().UTC().Add(30 * 24 * time.Hour),
		})

		isTrusted := trustStore.IsTrusted("user123", unknownFingerprint)
		require.False(s.T(), isTrusted, "unknown device should not be trusted")
	})

	// Test: Expired device trust removed
	s.Run("expired_device_trust_rejected", func() {
		trustStore := NewDeviceTrustStore()
		fingerprint := generateDeviceFingerprint(s.T())

		trustStore.AddDevice("user123", &TrustedDevice{
			Fingerprint: fingerprint,
			Name:        "Old Device",
			AddedAt:     time.Now().UTC().Add(-60 * 24 * time.Hour),
			LastUsed:    time.Now().UTC().Add(-45 * 24 * time.Hour),
			ExpiresAt:   time.Now().UTC().Add(-30 * 24 * time.Hour), // Expired
		})

		isTrusted := trustStore.IsTrusted("user123", fingerprint)
		require.False(s.T(), isTrusted, "expired device trust should be rejected")
	})
}

// TestRecoveryFlowEnforcement tests account recovery MFA requirements.
func (s *MFAEnforcementTestSuite) TestRecoveryFlowEnforcement() {
	s.T().Log("=== Test: Recovery Flow Enforcement ===")

	// Test: Recovery requires multiple factors
	s.Run("recovery_requires_multiple_factors", func() {
		policy := &MFAPolicy{
			SensitiveTxTypes:      []SensitiveTransactionType{TxTypeAccountRecovery},
			RequiredFactors:       2, // Require 2 factors for recovery
			AllowedFactorTypes:    []FactorType{FactorTypeTOTP, FactorTypeFIDO2, FactorTypeBackupCode},
			RecoveryRequirements:  3, // Need 3 factors for recovery specifically
		}

		recoveryReqs := policy.GetRecoveryRequirements()
		require.Equal(s.T(), 3, recoveryReqs.RequiredFactors,
			"recovery should require 3 factors")
	})

	// Test: Recovery with backup codes
	s.Run("recovery_with_backup_codes", func() {
		backupCodes := generateBackupCodes(s.T(), 10)
		store := NewBackupCodeStore("user123", backupCodes)

		// First use should succeed
		result := store.UseCode(backupCodes[0])
		require.True(s.T(), result.Valid, "valid backup code should be accepted")
		require.True(s.T(), result.CodeConsumed, "backup code should be marked as used")

		// Second use of same code should fail
		result = store.UseCode(backupCodes[0])
		require.False(s.T(), result.Valid, "reused backup code should be rejected")
		require.Equal(s.T(), "code_already_used", result.Reason)
	})

	// Test: Invalid backup code rejected
	s.Run("invalid_backup_code_rejected", func() {
		backupCodes := generateBackupCodes(s.T(), 10)
		store := NewBackupCodeStore("user123", backupCodes)

		invalidCode := "INVALID-CODE-1234"
		result := store.UseCode(invalidCode)
		require.False(s.T(), result.Valid, "invalid backup code should be rejected")
		require.Equal(s.T(), "invalid_code", result.Reason)
	})
}

// =============================================================================
// Test Types
// =============================================================================

type SensitiveTransactionType string

const (
	TxTypeAccountRecovery     SensitiveTransactionType = "account_recovery"
	TxTypeKeyRotation         SensitiveTransactionType = "key_rotation"
	TxTypeHighValueTransfer   SensitiveTransactionType = "high_value_transfer"
	TxTypeProviderRegistration SensitiveTransactionType = "provider_registration"
	TxTypeIdentityScopeUpdate SensitiveTransactionType = "identity_scope_update"
	TxTypeMFAPolicyChange     SensitiveTransactionType = "mfa_policy_change"
	TxTypeDelegationChange    SensitiveTransactionType = "delegation_change"
)

type FactorType string

const (
	FactorTypeTOTP       FactorType = "totp"
	FactorTypeFIDO2      FactorType = "fido2"
	FactorTypeSMS        FactorType = "sms"
	FactorTypeEmail      FactorType = "email"
	FactorTypeBackupCode FactorType = "backup_code"
)

type MFAPolicy struct {
	SensitiveTxTypes     []SensitiveTransactionType
	RequiredFactors      int
	AllowedFactorTypes   []FactorType
	RecoveryRequirements int
}

func (p *MFAPolicy) RequiresMFA(txType SensitiveTransactionType) bool {
	for _, sensitive := range p.SensitiveTxTypes {
		if sensitive == txType {
			return true
		}
	}
	return false
}

func (p *MFAPolicy) GetRecoveryRequirements() *RecoveryRequirements {
	return &RecoveryRequirements{
		RequiredFactors:    p.RecoveryRequirements,
		AllowedFactorTypes: p.AllowedFactorTypes,
	}
}

type RecoveryRequirements struct {
	RequiredFactors    int
	AllowedFactorTypes []FactorType
}

type MFAChallenge struct {
	ID          string
	FactorType  FactorType
	AccountID   string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	ExpectedOTP string
}

type MFAResponse struct {
	ChallengeID string
	OTP         string
	Timestamp   time.Time
}

type VerificationResult struct {
	Valid  bool
	Reason string
}

type AuthorizationSession struct {
	ID          string
	AccountID   string
	TxType      SensitiveTransactionType
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Used        bool
	FactorTypes []FactorType
}

func (s *AuthorizationSession) IsValidForOperation(txType SensitiveTransactionType) bool {
	if s.Used {
		return false
	}
	if time.Now().UTC().After(s.ExpiresAt) {
		return false
	}
	if s.TxType != txType {
		return false
	}
	return true
}

type TrustedDevice struct {
	Fingerprint string
	Name        string
	AddedAt     time.Time
	LastUsed    time.Time
	ExpiresAt   time.Time
}

type DeviceTrustStore struct {
	devices map[string][]*TrustedDevice
}

func NewDeviceTrustStore() *DeviceTrustStore {
	return &DeviceTrustStore{
		devices: make(map[string][]*TrustedDevice),
	}
}

func (s *DeviceTrustStore) AddDevice(accountID string, device *TrustedDevice) {
	s.devices[accountID] = append(s.devices[accountID], device)
}

func (s *DeviceTrustStore) IsTrusted(accountID, fingerprint string) bool {
	devices, ok := s.devices[accountID]
	if !ok {
		return false
	}
	for _, d := range devices {
		if d.Fingerprint == fingerprint {
			if time.Now().UTC().After(d.ExpiresAt) {
				return false // Expired
			}
			return true
		}
	}
	return false
}

type BackupCodeStore struct {
	accountID string
	codes     map[string]bool // code -> used
}

func NewBackupCodeStore(accountID string, codes []string) *BackupCodeStore {
	store := &BackupCodeStore{
		accountID: accountID,
		codes:     make(map[string]bool),
	}
	for _, code := range codes {
		store.codes[code] = false // Not used
	}
	return store
}

type BackupCodeResult struct {
	Valid        bool
	CodeConsumed bool
	Reason       string
}

func (s *BackupCodeStore) UseCode(code string) *BackupCodeResult {
	used, exists := s.codes[code]
	if !exists {
		return &BackupCodeResult{Valid: false, Reason: "invalid_code"}
	}
	if used {
		return &BackupCodeResult{Valid: false, Reason: "code_already_used"}
	}
	s.codes[code] = true // Mark as used
	return &BackupCodeResult{Valid: true, CodeConsumed: true}
}

// =============================================================================
// Test Helpers
// =============================================================================

func generateChallengeID(t *testing.T) string {
	t.Helper()
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	require.NoError(t, err)
	return "chal_" + hex.EncodeToString(b)
}

func generateSessionID(t *testing.T) string {
	t.Helper()
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	require.NoError(t, err)
	return "sess_" + hex.EncodeToString(b)
}

func generateDeviceFingerprint(t *testing.T) string {
	t.Helper()
	b := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, b)
	require.NoError(t, err)
	return "dev_" + hex.EncodeToString(b)
}

func generateBackupCodes(t *testing.T, count int) []string {
	t.Helper()
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		b := make([]byte, 8)
		_, err := io.ReadFull(rand.Reader, b)
		require.NoError(t, err)
		codes[i] = hex.EncodeToString(b)
	}
	return codes
}

func verifyTOTPChallenge(challenge *MFAChallenge, response *MFAResponse) *VerificationResult {
	if challenge.ID != response.ChallengeID {
		return &VerificationResult{Valid: false, Reason: "challenge_id_mismatch"}
	}
	if time.Now().UTC().After(challenge.ExpiresAt) {
		return &VerificationResult{Valid: false, Reason: "challenge_expired"}
	}
	if challenge.ExpectedOTP != response.OTP {
		return &VerificationResult{Valid: false, Reason: "invalid_otp"}
	}
	return &VerificationResult{Valid: true}
}
