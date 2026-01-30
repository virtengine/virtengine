// Package types provides fuzz tests for MFA (Multi-Factor Authentication) types.
// These tests use Go's native fuzzing support (Go 1.18+) to discover edge cases
// and potential vulnerabilities in MFA validation logic.
//
// Run with: go test -fuzz=. -fuzztime=30s ./x/mfa/types/...
//
// Task Reference: QUALITY-001 - Fuzz Testing Implementation
package types

import (
	"encoding/json"
	"testing"
	"time"
)

// FuzzFactorTypeFromString tests factor type parsing with arbitrary input.
// This fuzz test verifies that:
// 1. Parsing never panics regardless of input
// 2. Invalid types return appropriate errors
// 3. Valid types are parsed correctly
func FuzzFactorTypeFromString(f *testing.F) {
	// Valid factor types
	f.Add("totp")
	f.Add("fido2")
	f.Add("sms")
	f.Add("email")
	f.Add("veid")
	f.Add("trusted_device")
	f.Add("hardware_key")
	// Invalid types
	f.Add("")
	f.Add("invalid")
	f.Add("TOTP")        // Case sensitivity
	f.Add("totp ")       // Trailing space
	f.Add(" fido2")      // Leading space
	f.Add("unknown_type")
	f.Add("123")
	f.Add("\x00\x01\x02") // Binary data

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		ft, err := FactorTypeFromString(input)

		if err == nil {
			// If parsing succeeded, verify consistency
			if !ft.IsValid() {
				t.Errorf("parsed factor type %q is not valid", input)
			}

			// String conversion should be consistent
			str := ft.String()
			reparsed, reErr := FactorTypeFromString(str)
			if reErr != nil {
				t.Errorf("failed to reparse factor type string %q: %v", str, reErr)
			}
			if reparsed != ft {
				t.Errorf("reparsed type mismatch: got %v, want %v", reparsed, ft)
			}
		}
	})
}

// FuzzFactorTypeProperties tests factor type properties.
func FuzzFactorTypeProperties(f *testing.F) {
	// Test all uint8 values
	f.Add(uint8(0))
	f.Add(uint8(1))
	f.Add(uint8(2))
	f.Add(uint8(3))
	f.Add(uint8(4))
	f.Add(uint8(5))
	f.Add(uint8(6))
	f.Add(uint8(7))
	f.Add(uint8(255))

	f.Fuzz(func(t *testing.T, typeVal uint8) {
		ft := FactorType(typeVal)

		// Should never panic
		_ = ft.IsValid()
		_ = ft.String()
		_ = ft.GetSecurityLevel()
		_ = ft.RequiresOffChainVerification()

		// If valid, security level should be non-zero
		if ft.IsValid() {
			level := ft.GetSecurityLevel()
			if level < FactorSecurityLevelLow || level > FactorSecurityLevelHigh {
				t.Errorf("invalid security level %d for factor type %d", level, typeVal)
			}
		}
	})
}

// FuzzFactorEnrollmentValidate tests factor enrollment validation.
func FuzzFactorEnrollmentValidate(f *testing.F) {
	f.Add("cosmos1addr", uint8(1), "factor-001", "label", uint8(1), int64(1234567890), int64(1234567890))
	f.Add("", uint8(0), "", "", uint8(0), int64(0), int64(0))
	f.Add("addr", uint8(255), "id", "l", uint8(4), int64(-1), int64(-1))

	f.Fuzz(func(t *testing.T, address string, factorType uint8, factorID, label string, status uint8, enrolledAt, lastUsedAt int64) {
		enrollment := &FactorEnrollment{
			AccountAddress: address,
			FactorType:     FactorType(factorType),
			FactorID:       factorID,
			Label:          label,
			Status:         FactorEnrollmentStatus(status),
			EnrolledAt:     enrolledAt,
			LastUsedAt:     lastUsedAt,
		}

		// Should never panic
		_ = enrollment.Validate()
		_ = enrollment.IsActive()
	})
}

// FuzzMFAPolicyValidate tests MFA policy validation.
func FuzzMFAPolicyValidate(f *testing.F) {
	validPolicy := &MFAPolicy{
		AccountAddress: "cosmos1test",
		Enabled:        true,
		RequiredFactors: []FactorCombination{
			{Factors: []FactorType{FactorTypeTOTP}},
		},
	}
	validJSON, _ := json.Marshal(validPolicy)
	f.Add(validJSON)

	f.Add([]byte("{}"))
	f.Add([]byte("null"))
	f.Add([]byte(`{"account_address": ""}`))
	f.Add([]byte(`{"enabled": true, "required_factors": []}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var policy MFAPolicy
		if err := json.Unmarshal(data, &policy); err != nil {
			return
		}

		// Should never panic
		_ = policy.Validate()
	})
}

// FuzzFactorCombinationValidate tests factor combination validation.
func FuzzFactorCombinationValidate(f *testing.F) {
	// Test with combinations of factor bits
	f.Add(uint8(1), uint8(0), uint8(0), uint8(0)) // Single factor required
	f.Add(uint8(1), uint8(2), uint8(0), uint8(0)) // Two factors required
	f.Add(uint8(1), uint8(2), uint8(5), uint8(0)) // Three factors
	f.Add(uint8(0), uint8(0), uint8(0), uint8(0)) // Empty

	f.Fuzz(func(t *testing.T, f1, f2, f3, f4 uint8) {
		// Build required factors
		var required []FactorType
		for _, ft := range []uint8{f1, f2, f3, f4} {
			if ft > 0 && FactorType(ft).IsValid() {
				required = append(required, FactorType(ft))
			}
		}

		combination := FactorCombination{Factors: required}

		// Should never panic
		_ = combination.Validate()
		_ = combination.HasFactor(FactorTypeTOTP)
		_ = combination.HasFactor(FactorTypeFIDO2)
		_ = combination.GetSecurityLevel()
	})
}

// FuzzChallengeValidate tests challenge validation.
func FuzzChallengeValidate(f *testing.F) {
	f.Add("challenge-001", "cosmos1addr", uint8(1), "factor-001", int64(1234567890), int64(1234567895), uint8(1))
	f.Add("", "", uint8(0), "", int64(0), int64(0), uint8(0))
	f.Add("id", "addr", uint8(255), "fid", int64(-1), int64(-1), uint8(5))

	f.Fuzz(func(t *testing.T, challengeID, address string, factorType uint8, factorID string, createdAt, expiresAt int64, status uint8) {
		challenge := &Challenge{
			ChallengeID:    challengeID,
			AccountAddress: address,
			FactorType:     FactorType(factorType),
			FactorID:       factorID,
			ChallengeData:  []byte("random-data"),
			CreatedAt:      createdAt,
			ExpiresAt:      expiresAt,
			Status:         ChallengeStatus(status),
		}

		// Should never panic
		_ = challenge.Validate()
		_ = challenge.IsPending()
		_ = challenge.IsExpired(time.Now())
	})
}

// FuzzChallengeStatusProperties tests challenge status properties.
func FuzzChallengeStatusProperties(f *testing.F) {
	f.Add(uint8(0))
	f.Add(uint8(1))
	f.Add(uint8(2))
	f.Add(uint8(3))
	f.Add(uint8(4))
	f.Add(uint8(5))
	f.Add(uint8(255))

	f.Fuzz(func(t *testing.T, statusVal uint8) {
		status := ChallengeStatus(statusVal)

		// Should never panic
		_ = status.String()
	})
}

// FuzzFactorEnrollmentStatusProperties tests enrollment status properties.
func FuzzFactorEnrollmentStatusProperties(f *testing.F) {
	f.Add(uint8(0))
	f.Add(uint8(1))
	f.Add(uint8(2))
	f.Add(uint8(3))
	f.Add(uint8(4))
	f.Add(uint8(255))

	f.Fuzz(func(t *testing.T, statusVal uint8) {
		status := FactorEnrollmentStatus(statusVal)

		// Should never panic
		_ = status.IsValid()
		_ = status.String()
	})
}

// FuzzTrustedDevicePolicyValidate tests trusted device policy validation.
func FuzzTrustedDevicePolicyValidate(f *testing.F) {
	f.Add(true, int64(86400), uint32(5), true)
	f.Add(false, int64(0), uint32(0), false)
	f.Add(true, int64(-1), uint32(0), true)

	f.Fuzz(func(t *testing.T, enabled bool, trustDuration int64, maxDevices uint32, requireReauth bool) {
		policy := &TrustedDevicePolicy{
			Enabled:                   enabled,
			TrustDuration:             trustDuration,
			MaxTrustedDevices:         maxDevices,
			RequireReauthForSensitive: requireReauth,
		}

		// Should never panic
		_ = policy.Validate()
	})
}

// FuzzSensitiveTransactionTypeString tests transaction type string conversion.
func FuzzSensitiveTransactionTypeString(f *testing.F) {
	f.Add(uint8(0))
	f.Add(uint8(1))
	f.Add(uint8(2))
	f.Add(uint8(3))
	f.Add(uint8(4))
	f.Add(uint8(5))
	f.Add(uint8(255))

	f.Fuzz(func(t *testing.T, typeVal uint8) {
		txType := SensitiveTransactionType(typeVal)

		// Should never panic
		str := txType.String()

		// String should not be empty
		if str == "" {
			t.Error("transaction type string is empty")
		}
	})
}

// FuzzComputeFactorFingerprint tests factor fingerprint computation.
func FuzzComputeFactorFingerprint(f *testing.F) {
	f.Add(uint8(1), []byte("credential-data"))
	f.Add(uint8(0), []byte{})
	f.Add(uint8(255), make([]byte, 1000))

	f.Fuzz(func(t *testing.T, factorType uint8, credential []byte) {
		ft := FactorType(factorType)

		// Should never panic
		fp := ComputeFactorFingerprint(ft, credential)

		// Fingerprint should always be 32 bytes
		if len(fp) != 32 {
			t.Errorf("expected 32-byte fingerprint, got %d", len(fp))
		}

		// Same inputs should produce same fingerprint
		fp2 := ComputeFactorFingerprint(ft, credential)
		for i := range fp {
			if fp[i] != fp2[i] {
				t.Error("fingerprint not deterministic")
				break
			}
		}
	})
}

// FuzzComputeDeviceFingerprint tests device fingerprint computation.
func FuzzComputeDeviceFingerprint(f *testing.F) {
	f.Add("Device Name", []byte("device-key"))
	f.Add("", []byte{})
	f.Add("Long Device Name With Special Chars!@#$%", make([]byte, 256))

	f.Fuzz(func(t *testing.T, deviceName string, deviceKey []byte) {
		// Should never panic
		fp := ComputeDeviceFingerprint(deviceName, deviceKey)

		// Fingerprint should always be 32 bytes
		if len(fp) != 32 {
			t.Errorf("expected 32-byte fingerprint, got %d", len(fp))
		}

		// Same inputs should produce same fingerprint
		fp2 := ComputeDeviceFingerprint(deviceName, deviceKey)
		for i := range fp {
			if fp[i] != fp2[i] {
				t.Error("fingerprint not deterministic")
				break
			}
		}
	})
}

// FuzzComputeChallengeID tests challenge ID computation.
func FuzzComputeChallengeID(f *testing.F) {
	f.Add("cosmos1addr", uint8(1), int64(1234567890))
	f.Add("", uint8(0), int64(0))
	f.Add("address", uint8(255), int64(-1))

	f.Fuzz(func(t *testing.T, address string, factorType uint8, nonce int64) {
		ft := FactorType(factorType)

		// Should never panic
		cid := ComputeChallengeID(address, ft, nonce)

		// Challenge ID should always be 24 bytes
		if len(cid) != 24 {
			t.Errorf("expected 24-byte challenge ID, got %d", len(cid))
		}

		// Same inputs should produce same challenge ID
		cid2 := ComputeChallengeID(address, ft, nonce)
		for i := range cid {
			if cid[i] != cid2[i] {
				t.Error("challenge ID not deterministic")
				break
			}
		}
	})
}

// FuzzGenesisStateValidate tests genesis state validation.
func FuzzGenesisStateValidate(f *testing.F) {
	defaultGenesis := DefaultGenesisState()
	genesisJSON, _ := json.Marshal(defaultGenesis)
	f.Add(genesisJSON)

	f.Add([]byte("{}"))
	f.Add([]byte("null"))
	f.Add([]byte(`{"policies": null, "enrollments": null}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var gs GenesisState
		if err := json.Unmarshal(data, &gs); err != nil {
			return
		}

		// Should never panic
		_ = gs.Validate()
	})
}
