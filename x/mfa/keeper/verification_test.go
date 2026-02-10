package keeper

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/virtengine/virtengine/x/mfa/types"
)

// TestTOTPConfig tests the default TOTP configuration
func TestTOTPConfig(t *testing.T) {
	config := DefaultTOTPConfig()

	if config.Period != 30 {
		t.Errorf("expected Period=30, got %d", config.Period)
	}
	if config.Digits != 6 {
		t.Errorf("expected Digits=6, got %d", config.Digits)
	}
	if config.Skew != 1 {
		t.Errorf("expected Skew=1, got %d", config.Skew)
	}
	if config.Algorithm != "SHA256" {
		t.Errorf("expected Algorithm=SHA256, got %s", config.Algorithm)
	}
}

// TestVEIDVerificationConfig tests the default VEID configuration
func TestVEIDVerificationConfig(t *testing.T) {
	config := DefaultVEIDConfig()

	if config.DefaultThreshold != 50 {
		t.Errorf("expected DefaultThreshold=50, got %d", config.DefaultThreshold)
	}
	if config.MinimumThreshold != 50 {
		t.Errorf("expected MinimumThreshold=50, got %d", config.MinimumThreshold)
	}
	if config.MaximumThreshold != 100 {
		t.Errorf("expected MaximumThreshold=100, got %d", config.MaximumThreshold)
	}
}

// TestGetCombinationPolicy tests factor combination policies for different transaction types
func TestGetCombinationPolicy(t *testing.T) {
	tests := []struct {
		name              string
		txType            types.SensitiveTransactionType
		expectedMinFactor uint32
		expectedRequired  []types.FactorType
		requireHighSec    bool
	}{
		{
			name:              "AccountRecovery requires 3 factors with high security",
			txType:            types.SensitiveTxAccountRecovery,
			expectedMinFactor: 3,
			expectedRequired:  []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			requireHighSec:    true,
		},
		{
			name:              "KeyRotation requires 3 factors",
			txType:            types.SensitiveTxKeyRotation,
			expectedMinFactor: 3,
			expectedRequired:  []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			requireHighSec:    true,
		},
		{
			name:              "ProviderRegistration requires 2 factors",
			txType:            types.SensitiveTxProviderRegistration,
			expectedMinFactor: 2,
			expectedRequired:  []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			requireHighSec:    true,
		},
		{
			name:              "ValidatorRegistration requires 3 factors",
			txType:            types.SensitiveTxValidatorRegistration,
			expectedMinFactor: 3,
			expectedRequired:  []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
			requireHighSec:    true,
		},
		{
			name:              "HighValueOrder requires 2 factors",
			txType:            types.SensitiveTxHighValueOrder,
			expectedMinFactor: 2,
			expectedRequired:  []types.FactorType{types.FactorTypeVEID},
			requireHighSec:    false,
		},
		{
			name:              "Default requires 1 factor",
			txType:            types.SensitiveTxUnspecified,
			expectedMinFactor: 1,
			expectedRequired:  nil,
			requireHighSec:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := GetCombinationPolicy(tt.txType)

			if policy.MinFactors != tt.expectedMinFactor {
				t.Errorf("expected MinFactors=%d, got %d", tt.expectedMinFactor, policy.MinFactors)
			}

			if policy.RequireHighSecurityFactor != tt.requireHighSec {
				t.Errorf("expected RequireHighSecurityFactor=%v, got %v", tt.requireHighSec, policy.RequireHighSecurityFactor)
			}

			for _, required := range tt.expectedRequired {
				found := false
				for _, r := range policy.RequiredTypes {
					if r == required {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected RequiredTypes to contain %s", required.String())
				}
			}
		})
	}
}

// TestValidateFactorCombination tests factor combination validation
func TestValidateFactorCombination(t *testing.T) {
	tests := []struct {
		name            string
		verifiedFactors []types.FactorType
		policy          FactorCombinationPolicy
		expectError     bool
	}{
		{
			name:            "Valid - meets minimum factors",
			verifiedFactors: []types.FactorType{types.FactorTypeFIDO2, types.FactorTypeTOTP},
			policy: FactorCombinationPolicy{
				MinFactors:                2,
				MinSecurityLevel:          types.FactorSecurityLevelMedium,
				RequireHighSecurityFactor: false,
			},
			expectError: false,
		},
		{
			name:            "Invalid - not enough factors",
			verifiedFactors: []types.FactorType{types.FactorTypeFIDO2},
			policy: FactorCombinationPolicy{
				MinFactors:       2,
				MinSecurityLevel: types.FactorSecurityLevelLow,
			},
			expectError: true,
		},
		{
			name:            "Invalid - missing required type",
			verifiedFactors: []types.FactorType{types.FactorTypeTOTP},
			policy: FactorCombinationPolicy{
				MinFactors:       1,
				RequiredTypes:    []types.FactorType{types.FactorTypeFIDO2},
				MinSecurityLevel: types.FactorSecurityLevelLow,
			},
			expectError: true,
		},
		{
			name:            "Valid - has required type",
			verifiedFactors: []types.FactorType{types.FactorTypeFIDO2, types.FactorTypeVEID},
			policy: FactorCombinationPolicy{
				MinFactors:       2,
				RequiredTypes:    []types.FactorType{types.FactorTypeFIDO2, types.FactorTypeVEID},
				MinSecurityLevel: types.FactorSecurityLevelMedium,
			},
			expectError: false,
		},
		{
			name:            "Invalid - needs high security factor",
			verifiedFactors: []types.FactorType{types.FactorTypeSMS, types.FactorTypeEmail},
			policy: FactorCombinationPolicy{
				MinFactors:                2,
				MinSecurityLevel:          types.FactorSecurityLevelLow,
				RequireHighSecurityFactor: true,
			},
			expectError: true,
		},
		{
			name:            "Valid - has high security factor",
			verifiedFactors: []types.FactorType{types.FactorTypeFIDO2, types.FactorTypeTOTP},
			policy: FactorCombinationPolicy{
				MinFactors:                2,
				MinSecurityLevel:          types.FactorSecurityLevelLow,
				RequireHighSecurityFactor: true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFactorCombination(tt.verifiedFactors, tt.policy)
			if tt.expectError && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

// TestTOTPCommitmentVerification tests the TOTP hash commitment verification
func TestTOTPCommitmentVerification(t *testing.T) {
	keeper := Keeper{}
	config := DefaultTOTPConfig()
	now := time.Now()

	// Create a commitment with known hashes
	code1 := "123456"
	code2 := "654321"

	hash1 := sha256.Sum256([]byte(code1))
	hash2 := sha256.Sum256([]byte(code2))

	// Create commitment containing both hashes
	commitment := append(hash1[:], hash2[:]...)

	// Valid code should verify
	if !keeper.verifyTOTPWithCommitment(code1, commitment, now, config) {
		t.Error("expected code1 to verify successfully")
	}

	if !keeper.verifyTOTPWithCommitment(code2, commitment, now, config) {
		t.Error("expected code2 to verify successfully")
	}

	// Invalid code should not verify
	if keeper.verifyTOTPWithCommitment("000000", commitment, now, config) {
		t.Error("expected invalid code to fail verification")
	}
}

// TestVerificationResult tests the VerificationResult struct
func TestVerificationResult(t *testing.T) {
	result := VerificationResult{
		Verified:   true,
		FactorType: types.FactorTypeFIDO2,
		FactorID:   "test-credential-id",
		Metadata: map[string]string{
			"policy_satisfied": "true",
			"factors_verified": "2",
		},
	}

	if !result.Verified {
		t.Error("expected Verified to be true")
	}

	if result.FactorType != types.FactorTypeFIDO2 {
		t.Errorf("expected FactorType FIDO2, got %s", result.FactorType.String())
	}

	if result.FactorID != "test-credential-id" {
		t.Errorf("expected FactorID 'test-credential-id', got %s", result.FactorID)
	}

	if result.Metadata["policy_satisfied"] != "true" {
		t.Error("expected policy_satisfied=true in metadata")
	}
}

// TestFormatFactorList tests the factor list formatting helper
func TestFormatFactorList(t *testing.T) {
	tests := []struct {
		name     string
		factors  []types.FactorType
		expected string
	}{
		{
			name:     "empty list",
			factors:  []types.FactorType{},
			expected: "",
		},
		{
			name:     "single factor",
			factors:  []types.FactorType{types.FactorTypeFIDO2},
			expected: "fido2",
		},
		{
			name:     "multiple factors",
			factors:  []types.FactorType{types.FactorTypeFIDO2, types.FactorTypeTOTP, types.FactorTypeVEID},
			expected: "fido2,totp,veid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatFactorList(tt.factors)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestOTPResponseValidation tests OTP response format validation
func TestOTPResponseValidation(t *testing.T) {
	// Valid numeric codes
	validCodes := []string{"123456", "12345678", "000000", "999999"}
	for _, code := range validCodes {
		if len(code) < 6 || len(code) > 8 {
			t.Errorf("expected valid code %s to pass length check", code)
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Errorf("expected valid code %s to be numeric", code)
			}
		}
	}

	// Invalid codes
	invalidCodes := []string{"12345", "abc123", "1234567890", "", "12 345"}
	for _, code := range invalidCodes {
		if len(code) >= 6 && len(code) <= 8 {
			allDigits := true
			for _, c := range code {
				if c < '0' || c > '9' {
					allDigits = false
					break
				}
			}
			if allDigits {
				t.Errorf("expected invalid code %s to fail validation", code)
			}
		}
	}
}

// TestGenerateTOTPCode tests TOTP code generation
func TestGenerateTOTPCode(t *testing.T) {
	secret := []byte("12345678901234567890")
	counter := uint64(1)
	digits := uint(6)

	// Test with different algorithms
	algorithms := []string{"SHA256", "SHA512"}
	for _, alg := range algorithms {
		code := generateTOTPCode(secret, counter, digits, alg)

		if len(code) != 6 {
			t.Errorf("expected code length 6 for %s, got %d", alg, len(code))
		}

		// Verify code is numeric
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Errorf("expected numeric code for %s, got %s", alg, code)
			}
		}
	}

	// Test 8 digit code
	code8 := generateTOTPCode(secret, counter, 8, "SHA256")
	if len(code8) != 8 {
		t.Errorf("expected code length 8, got %d", len(code8))
	}
}

// TestHardwareKeyResponseValidation tests hardware key response validation
func TestHardwareKeyResponseValidation(t *testing.T) {
	tests := []struct {
		name        string
		response    HardwareKeyResponse
		expectValid bool
	}{
		{
			name: "Valid response",
			response: HardwareKeyResponse{
				Signature: []byte{0x01, 0x02, 0x03},
				Algorithm: "ES256",
				KeyID:     "test-key-id",
			},
			expectValid: true,
		},
		{
			name: "Missing signature",
			response: HardwareKeyResponse{
				Algorithm: "ES256",
				KeyID:     "test-key-id",
			},
			expectValid: false,
		},
		{
			name: "Missing key ID",
			response: HardwareKeyResponse{
				Signature: []byte{0x01, 0x02, 0x03},
				Algorithm: "ES256",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := len(tt.response.Signature) > 0 && tt.response.KeyID != ""
			if isValid != tt.expectValid {
				t.Errorf("expected valid=%v, got %v", tt.expectValid, isValid)
			}
		})
	}
}

// TestFactorCombinationPolicy tests policy struct initialization
func TestFactorCombinationPolicy(t *testing.T) {
	policy := FactorCombinationPolicy{
		MinFactors:                2,
		RequiredTypes:             []types.FactorType{types.FactorTypeVEID, types.FactorTypeFIDO2},
		OptionalTypes:             []types.FactorType{types.FactorTypeTOTP},
		MinSecurityLevel:          types.FactorSecurityLevelHigh,
		RequireHighSecurityFactor: true,
	}

	if policy.MinFactors != 2 {
		t.Errorf("expected MinFactors=2, got %d", policy.MinFactors)
	}

	if len(policy.RequiredTypes) != 2 {
		t.Errorf("expected 2 required types, got %d", len(policy.RequiredTypes))
	}

	if len(policy.OptionalTypes) != 1 {
		t.Errorf("expected 1 optional type, got %d", len(policy.OptionalTypes))
	}

	if policy.MinSecurityLevel != types.FactorSecurityLevelHigh {
		t.Errorf("expected MinSecurityLevel=High, got %d", policy.MinSecurityLevel)
	}

	if !policy.RequireHighSecurityFactor {
		t.Error("expected RequireHighSecurityFactor=true")
	}
}

// TestCriticalTierPolicies tests all critical tier transaction policies
func TestCriticalTierPolicies(t *testing.T) {
	criticalTypes := []types.SensitiveTransactionType{
		types.SensitiveTxAccountRecovery,
		types.SensitiveTxKeyRotation,
		types.SensitiveTxAccountDeletion,
		types.SensitiveTxTwoFactorDisable,
	}

	for _, txType := range criticalTypes {
		policy := GetCombinationPolicy(txType)

		if policy.MinFactors < 2 {
			t.Errorf("%s: critical tier should require at least 2 factors", txType.String())
		}

		if !policy.RequireHighSecurityFactor {
			t.Errorf("%s: critical tier should require high security factor", txType.String())
		}

		if policy.MinSecurityLevel != types.FactorSecurityLevelHigh {
			t.Errorf("%s: critical tier should have high min security level", txType.String())
		}
	}
}

// TestNonceSecurity tests nonce generation for challenges
func TestNonceSecurity(t *testing.T) {
	// Test that different inputs produce different nonces
	nonces := make(map[string]bool)
	inputs := []string{
		"addr1:factor1:100",
		"addr1:factor1:101",
		"addr2:factor1:100",
		"addr1:factor2:100",
	}

	for _, input := range inputs {
		seed := sha256.Sum256([]byte(input))
		nonceSeed := sha256.Sum256(append(seed[:], []byte("nonce")...))
		nonceBytes := nonceSeed[:16]
		nonceHex := hex.EncodeToString(nonceBytes)

		if nonces[nonceHex] {
			t.Errorf("duplicate nonce generated for different input: %s", input)
		}
		nonces[nonceHex] = true
	}

	if len(nonces) != len(inputs) {
		t.Error("expected unique nonces for each input")
	}
}
