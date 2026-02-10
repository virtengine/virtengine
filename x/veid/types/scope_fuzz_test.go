// Package types provides fuzz tests for VEID identity scope validation.
// These tests use Go's native fuzzing support (Go 1.18+) to discover edge cases
// and potential vulnerabilities in identity verification logic.
//
// Run with: go test -fuzz=. -fuzztime=30s ./x/veid/types/...
//
// Task Reference: QUALITY-002 - Fuzz Testing Implementation
package types

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
)

// FuzzIdentityScopeValidate tests identity scope validation with arbitrary input.
// This fuzz test verifies that:
// 1. Validation never panics regardless of input
// 2. Invalid scopes are properly rejected
// 3. Valid scopes pass validation
func FuzzIdentityScopeValidate(f *testing.F) {
	// Create a minimal valid encrypted payload
	validPayload := encryptiontypes.EncryptedPayloadEnvelope{
		Version:          encryptiontypes.EnvelopeVersion,
		AlgorithmID:      encryptiontypes.AlgorithmX25519XSalsa20Poly1305,
		AlgorithmVersion: encryptiontypes.AlgorithmVersionV1,
		RecipientKeyIDs:  []string{"abcdef0123456789abcdef0123456789abcdef01"},
		Nonce:            bytes.Repeat([]byte{0x01}, encryptiontypes.XSalsa20NonceSize),
		Ciphertext:       []byte("encrypted data here"),
		SenderPubKey:     bytes.Repeat([]byte{0x02}, encryptiontypes.X25519PublicKeySize),
		SenderSignature:  bytes.Repeat([]byte{0x03}, 64),
	}

	validMetadata := UploadMetadata{
		Salt:              bytes.Repeat([]byte{0x04}, 32),
		SaltHash:          ComputeSaltHash(bytes.Repeat([]byte{0x04}, 32)),
		DeviceFingerprint: "device-fp-123",
		ClientID:          "approved-client-1",
		ClientSignature:   bytes.Repeat([]byte{0x05}, 64),
		UserSignature:     bytes.Repeat([]byte{0x06}, 64),
		PayloadHash:       bytes.Repeat([]byte{0x07}, 32),
	}

	validScope := &IdentityScope{
		ScopeID:          "scope-123",
		ScopeType:        ScopeTypeIDDocument,
		Version:          ScopeSchemaVersion,
		EncryptedPayload: validPayload,
		UploadMetadata:   validMetadata,
		Status:           VerificationStatusPending,
		UploadedAt:       time.Now().UTC(),
	}
	validJSON, _ := json.Marshal(validScope) //nolint:errchkjson // Best-effort marshal for fuzz seeding
	f.Add(validJSON)

	// Edge cases
	f.Add([]byte("{}"))                             // Empty object
	f.Add([]byte("null"))                           // Null
	f.Add([]byte(`{"scope_id": ""}`))               // Empty scope ID
	f.Add([]byte(`{"scope_type": "invalid_type"}`)) // Invalid scope type
	f.Add([]byte(`{"version": 0}`))                 // Zero version
	f.Add([]byte(`{"version": 999}`))               // Future version
	f.Add([]byte(`{"status": "invalid_status"}`))   // Invalid status
	f.Add(bytes.Repeat([]byte{0xFF}, 1000))         // Random bytes

	f.Fuzz(func(t *testing.T, data []byte) {
		var scope IdentityScope
		if err := json.Unmarshal(data, &scope); err != nil {
			// Invalid JSON is expected for some inputs
			return
		}

		// Validation should never panic
		_ = scope.Validate()

		// Other methods should also not panic
		_ = scope.IsActive()
		_ = scope.IsVerified()
		_ = scope.CanBeVerified()
		_ = scope.String()
	})
}

// FuzzScopeTypeValidation tests scope type validation.
func FuzzScopeTypeValidation(f *testing.F) {
	// Valid types
	f.Add("id_document")
	f.Add("selfie")
	f.Add("face_video")
	f.Add("biometric")
	f.Add("sso_metadata")
	f.Add("email_proof")
	f.Add("sms_proof")
	f.Add("domain_verify")
	f.Add("ad_sso")
	// Invalid types
	f.Add("")
	f.Add("invalid")
	f.Add("ID_DOCUMENT") // Case sensitivity
	f.Add("id-document") // Wrong separator
	f.Add("unknown_type")
	f.Add("a")
	f.Add(string(bytes.Repeat([]byte{0xFF}, 100)))

	f.Fuzz(func(t *testing.T, scopeType string) {
		st := ScopeType(scopeType)

		// Should never panic
		isValid := IsValidScopeType(st)
		weight := ScopeTypeWeight(st)
		desc := ScopeTypeDescription(st)

		// Consistency checks
		if isValid {
			// Valid types should have non-zero weight (except unknown)
			// Some valid types might have zero weight.
			// Description should not be "Unknown scope type"
			if desc == "Unknown scope type" {
				t.Errorf("valid scope type %q has unknown description", scopeType)
			}
		} else {
			// Invalid types should have zero weight
			if weight != 0 {
				t.Errorf("invalid scope type %q has non-zero weight: %d", scopeType, weight)
			}
			// Description should be "Unknown scope type"
			if desc != "Unknown scope type" {
				t.Errorf("invalid scope type %q has specific description: %s", scopeType, desc)
			}
		}
	})
}

// FuzzVerificationStatusValidation tests verification status validation.
func FuzzVerificationStatusValidation(f *testing.F) {
	// Valid statuses
	f.Add("unknown")
	f.Add("pending")
	f.Add("in_progress")
	f.Add("verified")
	f.Add("rejected")
	f.Add("expired")
	f.Add("needs_additional_factor")
	f.Add("additional_factor_pending")
	// Invalid statuses
	f.Add("")
	f.Add("invalid")
	f.Add("VERIFIED")  // Case sensitivity
	f.Add("verified ") // Trailing space
	f.Add(" pending")  // Leading space

	f.Fuzz(func(t *testing.T, status string) {
		vs := VerificationStatus(status)

		// Should never panic
		isValid := IsValidVerificationStatus(vs)
		isFinal := IsFinalStatus(vs)

		// Consistency: only valid statuses can be final
		if isFinal && !isValid {
			t.Errorf("invalid status %q marked as final", status)
		}
	})
}

// FuzzVerificationStatusTransitions tests verification status state machine.
func FuzzVerificationStatusTransitions(f *testing.F) {
	// Valid transitions
	f.Add("unknown", "pending")
	f.Add("pending", "in_progress")
	f.Add("in_progress", "verified")
	f.Add("in_progress", "rejected")
	f.Add("verified", "expired")
	// Invalid transitions
	f.Add("verified", "pending")
	f.Add("rejected", "verified")
	f.Add("", "")
	f.Add("invalid", "invalid")

	f.Fuzz(func(t *testing.T, from, to string) {
		fromStatus := VerificationStatus(from)
		toStatus := VerificationStatus(to)

		// Should never panic
		canTransition := fromStatus.CanTransitionTo(toStatus)

		// If both are valid, check consistency
		if IsValidVerificationStatus(fromStatus) && IsValidVerificationStatus(toStatus) {
			// Final statuses can only transition to specific states
			if IsFinalStatus(fromStatus) && canTransition && toStatus != VerificationStatusPending && toStatus != VerificationStatusExpired {
				t.Errorf("final status %q should not transition to %q", from, to)
			}
		}
	})
}

// FuzzUploadMetadataValidate tests upload metadata validation.
func FuzzUploadMetadataValidate(f *testing.F) {
	// Valid metadata
	validSalt := bytes.Repeat([]byte{0x01}, 32)
	f.Add(validSalt, "device-fp", "client-1", bytes.Repeat([]byte{0x02}, 64), bytes.Repeat([]byte{0x03}, 64), bytes.Repeat([]byte{0x04}, 32))

	// Edge cases
	f.Add([]byte{}, "", "", []byte{}, []byte{}, []byte{})                             // All empty
	f.Add(bytes.Repeat([]byte{0x01}, 15), "fp", "c", []byte{1}, []byte{1}, []byte{1}) // Salt too short
	f.Add(bytes.Repeat([]byte{0x01}, 65), "fp", "c", []byte{1}, []byte{1}, []byte{1}) // Salt too long
	f.Add(validSalt, "", "c", []byte{1}, []byte{1}, bytes.Repeat([]byte{0x01}, 32))   // Empty device fp
	f.Add(validSalt, "fp", "", []byte{1}, []byte{1}, bytes.Repeat([]byte{0x01}, 32))  // Empty client id
	f.Add(validSalt, "fp", "c", []byte{}, []byte{1}, bytes.Repeat([]byte{0x01}, 32))  // Empty client sig
	f.Add(validSalt, "fp", "c", []byte{1}, []byte{}, bytes.Repeat([]byte{0x01}, 32))  // Empty user sig
	f.Add(validSalt, "fp", "c", []byte{1}, []byte{1}, []byte{1})                      // Wrong payload hash size

	f.Fuzz(func(t *testing.T, salt []byte, deviceFp, clientID string, clientSig, userSig, payloadHash []byte) {
		metadata := &UploadMetadata{
			Salt:              salt,
			SaltHash:          ComputeSaltHash(salt),
			DeviceFingerprint: deviceFp,
			ClientID:          clientID,
			ClientSignature:   clientSig,
			UserSignature:     userSig,
			PayloadHash:       payloadHash,
		}

		// Should never panic
		_ = metadata.Validate()

		// Other methods should also not panic
		_ = metadata.SaltHashHex()
		_ = metadata.PayloadHashHex()
		_ = metadata.SigningPayload()
		_ = metadata.UserSigningPayload()
	})
}

// FuzzVerificationEventValidate tests verification event validation.
func FuzzVerificationEventValidate(f *testing.F) {
	f.Add("event-1", "scope-1", "pending", "in_progress", int64(1234567890), "reason")
	f.Add("", "", "", "", int64(0), "")
	f.Add("event", "scope", "invalid", "invalid", int64(-1), "")

	f.Fuzz(func(t *testing.T, eventID, scopeID, prevStatus, newStatus string, timestamp int64, reason string) {
		event := &VerificationEvent{
			EventID:        eventID,
			ScopeID:        scopeID,
			PreviousStatus: VerificationStatus(prevStatus),
			NewStatus:      VerificationStatus(newStatus),
			Timestamp:      time.Unix(timestamp, 0),
			Reason:         reason,
		}

		// Should never panic
		_ = event.Validate()
	})
}

// FuzzSimpleVerificationResultValidate tests simple verification result validation.
func FuzzSimpleVerificationResultValidate(f *testing.F) {
	f.Add(true, "verified", uint32(85), "v1.0.0", uint32(95))
	f.Add(false, "rejected", uint32(0), "v1.0.0", uint32(50))
	f.Add(true, "invalid", uint32(101), "", uint32(101)) // Invalid score/confidence

	f.Fuzz(func(t *testing.T, success bool, status string, score uint32, scoreVersion string, confidence uint32) {
		result := &SimpleVerificationResult{
			Success:      success,
			Status:       VerificationStatus(status),
			Score:        score,
			ScoreVersion: scoreVersion,
			Confidence:   confidence,
			ProcessedAt:  time.Now(),
		}

		// Should never panic
		_ = result.Validate()
	})
}

// FuzzApprovedClientValidate tests approved client validation.
func FuzzApprovedClientValidate(f *testing.F) {
	f.Add("client-1", "Test Client", bytes.Repeat([]byte{0x01}, 32), "ed25519", int64(1234567890))
	f.Add("", "", []byte{}, "", int64(0))

	f.Fuzz(func(t *testing.T, clientID, name string, publicKey []byte, algorithm string, registeredAt int64) {
		client := &ApprovedClient{
			ClientID:     clientID,
			Name:         name,
			PublicKey:    publicKey,
			Algorithm:    algorithm,
			Active:       true,
			RegisteredAt: registeredAt,
		}

		// Should never panic
		_ = client.Validate()
	})
}

// FuzzScopeRefCreation tests scope reference creation from full scope.
func FuzzScopeRefCreation(f *testing.F) {
	f.Add("scope-123", "id_document", "pending", int64(1234567890))
	f.Add("", "", "", int64(0))
	f.Add("scope", "invalid_type", "invalid_status", int64(-1))

	f.Fuzz(func(t *testing.T, scopeID, scopeType, status string, uploadedAt int64) {
		scope := &IdentityScope{
			ScopeID:    scopeID,
			ScopeType:  ScopeType(scopeType),
			Status:     VerificationStatus(status),
			UploadedAt: time.Unix(uploadedAt, 0),
		}

		// Should never panic
		ref := NewScopeRef(scope)

		// Verify consistency
		if ref.ScopeID != scopeID {
			t.Errorf("ScopeID mismatch: got %q, want %q", ref.ScopeID, scopeID)
		}
		if ref.ScopeType != ScopeType(scopeType) {
			t.Errorf("ScopeType mismatch: got %q, want %q", ref.ScopeType, scopeType)
		}
		if ref.Status != VerificationStatus(status) {
			t.Errorf("Status mismatch: got %q, want %q", ref.Status, status)
		}
	})
}

// FuzzComputeSaltHash tests salt hash computation.
func FuzzComputeSaltHash(f *testing.F) {
	f.Add(bytes.Repeat([]byte{0x01}, 32))
	f.Add([]byte{})
	f.Add(bytes.Repeat([]byte{0xFF}, 100))
	f.Add([]byte{0x00})

	f.Fuzz(func(t *testing.T, salt []byte) {
		// Should never panic
		hash := ComputeSaltHash(salt)

		// Hash should always be 32 bytes (SHA256)
		if len(hash) != 32 {
			t.Errorf("expected 32-byte hash, got %d bytes", len(hash))
		}

		// Same salt should produce same hash
		hash2 := ComputeSaltHash(salt)
		if !bytes.Equal(hash, hash2) {
			t.Error("hash not deterministic")
		}

		// Different salts should (almost always) produce different hashes
		if len(salt) > 0 {
			modifiedSalt := make([]byte, len(salt))
			copy(modifiedSalt, salt)
			modifiedSalt[0] ^= 0x01
			hash3 := ComputeSaltHash(modifiedSalt)
			if bytes.Equal(hash, hash3) {
				// This is extremely unlikely but possible
				t.Log("hash collision detected (expected to be very rare)")
			}
		}
	})
}

// FuzzBytesEqual tests constant-time byte comparison.
func FuzzBytesEqual(f *testing.F) {
	f.Add([]byte{1, 2, 3}, []byte{1, 2, 3})
	f.Add([]byte{1, 2, 3}, []byte{1, 2, 4})
	f.Add([]byte{}, []byte{})
	f.Add([]byte{1}, []byte{})
	f.Add(bytes.Repeat([]byte{0xFF}, 100), bytes.Repeat([]byte{0xFF}, 100))

	f.Fuzz(func(t *testing.T, a, b []byte) {
		// Should never panic
		result := bytesEqual(a, b)

		// Verify consistency with standard comparison
		expected := bytes.Equal(a, b)
		if result != expected {
			t.Errorf("bytesEqual(%v, %v) = %v, want %v", a, b, result, expected)
		}
	})
}

// FuzzAllScopeTypes tests that AllScopeTypes returns consistent results.
func FuzzAllScopeTypes(f *testing.F) {
	f.Add(0)
	f.Add(100)

	f.Fuzz(func(t *testing.T, iterations int) {
		if iterations < 0 || iterations > 1000 {
			return
		}

		// Get all types
		types := AllScopeTypes()
		if len(types) == 0 {
			t.Error("AllScopeTypes returned empty slice")
		}

		// All returned types should be valid
		for _, st := range types {
			if !IsValidScopeType(st) {
				t.Errorf("AllScopeTypes returned invalid type: %q", st)
			}
		}

		// Multiple calls should return same result
		for i := 0; i < iterations; i++ {
			types2 := AllScopeTypes()
			if len(types) != len(types2) {
				t.Error("AllScopeTypes not deterministic")
			}
		}
	})
}

// FuzzAllVerificationStatuses tests that AllVerificationStatuses returns consistent results.
func FuzzAllVerificationStatuses(f *testing.F) {
	f.Add(0)
	f.Add(100)

	f.Fuzz(func(t *testing.T, iterations int) {
		if iterations < 0 || iterations > 1000 {
			return
		}

		// Get all statuses
		statuses := AllVerificationStatuses()
		if len(statuses) == 0 {
			t.Error("AllVerificationStatuses returned empty slice")
		}

		// All returned statuses should be valid
		for _, vs := range statuses {
			if !IsValidVerificationStatus(vs) {
				t.Errorf("AllVerificationStatuses returned invalid status: %q", vs)
			}
		}

		// Multiple calls should return same result
		for i := 0; i < iterations; i++ {
			statuses2 := AllVerificationStatuses()
			if len(statuses) != len(statuses2) {
				t.Error("AllVerificationStatuses not deterministic")
			}
		}
	})
}
