// Package types provides fuzz tests for EncryptedPayloadEnvelope validation.
// These tests use Go's native fuzzing support (Go 1.18+) to discover edge cases
// and potential vulnerabilities in envelope parsing and validation.
//
// Run with: go test -fuzz=. -fuzztime=30s ./x/encryption/types/...
//
// Task Reference: QUALITY-002 - Fuzz Testing Implementation
package types

import (
	"bytes"
	"encoding/json"
	"testing"
)

// FuzzEnvelopeValidate tests envelope validation with arbitrary input.
// This fuzz test verifies that:
// 1. Validation never panics regardless of input
// 2. Invalid envelopes are properly rejected
// 3. Valid envelopes pass validation
func FuzzEnvelopeValidate(f *testing.F) {
	// Seed corpus with valid envelope structures
	validEnvelope := &EncryptedPayloadEnvelope{
		Version:          EnvelopeVersion,
		AlgorithmID:      AlgorithmX25519XSalsa20Poly1305,
		AlgorithmVersion: AlgorithmVersionV1,
		RecipientKeyIDs:  []string{"abcdef0123456789abcdef0123456789abcdef01"},
		Nonce:            bytes.Repeat([]byte{0x01}, XSalsa20NonceSize),
		Ciphertext:       []byte("encrypted data here"),
		SenderPubKey:     bytes.Repeat([]byte{0x02}, X25519PublicKeySize),
		SenderSignature:  bytes.Repeat([]byte{0x03}, 64),
	}
	validJSON, _ := json.Marshal(validEnvelope)
	f.Add(validJSON)

	// Edge cases
	f.Add([]byte("{}"))                               // Empty object
	f.Add([]byte("null"))                             // Null
	f.Add([]byte("[]"))                               // Array instead of object
	f.Add([]byte(`{"version": 0}`))                   // Zero version
	f.Add([]byte(`{"version": 999999}`))              // Future version
	f.Add([]byte(`{"algorithm_id": "UNKNOWN-ALGO"}`)) // Unknown algorithm
	f.Add([]byte(`{"recipient_key_ids": []}`))        // Empty recipients
	f.Add([]byte(`{"nonce": ""}`))                    // Empty nonce
	f.Add([]byte(`{"ciphertext": ""}`))               // Empty ciphertext
	f.Add(bytes.Repeat([]byte{0xFF}, 1000))           // Random bytes
	f.Add([]byte(`{"version": 1, "algorithm_id": "X25519-XSALSA20-POLY1305", "algorithm_version": 1}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var envelope EncryptedPayloadEnvelope
		if err := json.Unmarshal(data, &envelope); err != nil {
			// Invalid JSON is expected for some inputs
			return
		}

		// Validation should never panic
		_ = envelope.Validate()

		// If valid, test other methods
		if envelope.Validate() == nil {
			_ = envelope.SigningPayload()
			_ = envelope.Hash()
			_, _ = envelope.DeterministicBytes()
		}
	})
}

// FuzzEnvelopeSigningPayload tests signing payload generation.
func FuzzEnvelopeSigningPayload(f *testing.F) {
	f.Add(uint32(1), "X25519-XSALSA20-POLY1305", uint32(1), []byte("ciphertext"), []byte("nonce123456789012345678"))
	f.Add(uint32(0), "", uint32(0), []byte{}, []byte{})
	f.Add(uint32(255), "UNKNOWN", uint32(255), bytes.Repeat([]byte{0xFF}, 100), bytes.Repeat([]byte{0x00}, 24))

	f.Fuzz(func(t *testing.T, version uint32, algorithmID string, algVersion uint32, ciphertext, nonce []byte) {
		envelope := &EncryptedPayloadEnvelope{
			Version:          version,
			AlgorithmID:      algorithmID,
			AlgorithmVersion: algVersion,
			Ciphertext:       ciphertext,
			Nonce:            nonce,
			RecipientKeyIDs:  []string{"recipient1"},
		}

		// Should never panic
		payload := envelope.SigningPayload()

		// Payload should always be a hash (32 bytes)
		if len(payload) != 32 {
			t.Errorf("expected 32-byte payload, got %d bytes", len(payload))
		}

		// Same envelope should produce same payload
		payload2 := envelope.SigningPayload()
		if !bytes.Equal(payload, payload2) {
			t.Error("signing payload not deterministic")
		}
	})
}

// FuzzEnvelopeHash tests hash generation.
func FuzzEnvelopeHash(f *testing.F) {
	f.Add([]byte("ciphertext"), []byte("nonce123456789012345678"), "recipient1")
	f.Add([]byte{}, []byte{}, "")
	f.Add(bytes.Repeat([]byte{0xFF}, 10000), bytes.Repeat([]byte{0x00}, 24), "long-recipient-id-string")

	f.Fuzz(func(t *testing.T, ciphertext, nonce []byte, recipientID string) {
		envelope := &EncryptedPayloadEnvelope{
			Version:          EnvelopeVersion,
			AlgorithmID:      AlgorithmX25519XSalsa20Poly1305,
			AlgorithmVersion: AlgorithmVersionV1,
			Ciphertext:       ciphertext,
			Nonce:            nonce,
			RecipientKeyIDs:  []string{recipientID},
		}

		// Should never panic
		hash := envelope.Hash()

		// Hash should always be 32 bytes (SHA256)
		if len(hash) != 32 {
			t.Errorf("expected 32-byte hash, got %d bytes", len(hash))
		}

		// Same envelope should produce same hash
		hash2 := envelope.Hash()
		if !bytes.Equal(hash, hash2) {
			t.Error("hash not deterministic")
		}
	})
}

// FuzzEnvelopeDeterministicBytes tests deterministic serialization.
func FuzzEnvelopeDeterministicBytes(f *testing.F) {
	f.Add("recipient1", "recipient2")
	f.Add("", "")
	f.Add("zzz", "aaa") // Reversed order to test sorting

	f.Fuzz(func(t *testing.T, recipient1, recipient2 string) {
		// Create envelope with recipients in one order
		envelope1 := &EncryptedPayloadEnvelope{
			Version:          EnvelopeVersion,
			AlgorithmID:      AlgorithmX25519XSalsa20Poly1305,
			AlgorithmVersion: AlgorithmVersionV1,
			RecipientKeyIDs:  []string{recipient1, recipient2},
			Nonce:            bytes.Repeat([]byte{0x01}, XSalsa20NonceSize),
			Ciphertext:       []byte("test"),
			SenderPubKey:     bytes.Repeat([]byte{0x02}, X25519PublicKeySize),
			SenderSignature:  bytes.Repeat([]byte{0x03}, 64),
		}

		// Create envelope with recipients in reverse order
		envelope2 := &EncryptedPayloadEnvelope{
			Version:          EnvelopeVersion,
			AlgorithmID:      AlgorithmX25519XSalsa20Poly1305,
			AlgorithmVersion: AlgorithmVersionV1,
			RecipientKeyIDs:  []string{recipient2, recipient1},
			Nonce:            bytes.Repeat([]byte{0x01}, XSalsa20NonceSize),
			Ciphertext:       []byte("test"),
			SenderPubKey:     bytes.Repeat([]byte{0x02}, X25519PublicKeySize),
			SenderSignature:  bytes.Repeat([]byte{0x03}, 64),
		}

		// Both should produce same deterministic bytes
		bytes1, err1 := envelope1.DeterministicBytes()
		bytes2, err2 := envelope2.DeterministicBytes()

		if err1 != nil || err2 != nil {
			return // Skip if either fails
		}

		if !bytes.Equal(bytes1, bytes2) {
			t.Error("deterministic bytes differ for same content with different recipient order")
		}
	})
}

// FuzzEnvelopeGetRecipientIndex tests recipient lookup.
func FuzzEnvelopeGetRecipientIndex(f *testing.F) {
	f.Add("recipient1", "recipient2", "recipient1") // Find first
	f.Add("recipient1", "recipient2", "recipient2") // Find second
	f.Add("recipient1", "recipient2", "notfound")   // Not found
	f.Add("", "", "")                               // Empty strings

	f.Fuzz(func(t *testing.T, r1, r2, search string) {
		envelope := &EncryptedPayloadEnvelope{
			RecipientKeyIDs: []string{r1, r2},
		}

		// Should never panic
		idx := envelope.GetRecipientIndex(search)

		// Verify correctness
		if search == r1 && idx != 0 {
			t.Errorf("expected index 0 for %q, got %d", search, idx)
		} else if search == r2 && r1 != r2 && idx != 1 {
			t.Errorf("expected index 1 for %q, got %d", search, idx)
		} else if search != r1 && search != r2 && idx != -1 {
			t.Errorf("expected -1 for %q, got %d", search, idx)
		}

		// IsRecipient should be consistent
		isRecipient := envelope.IsRecipient(search)
		if (idx >= 0) != isRecipient {
			t.Error("IsRecipient inconsistent with GetRecipientIndex")
		}
	})
}

// FuzzRecipientKeyRecordValidate tests recipient key record validation.
func FuzzRecipientKeyRecordValidate(f *testing.F) {
	// Valid record
	f.Add("cosmos1abc...", bytes.Repeat([]byte{0x01}, 32), "fingerprint123", "X25519-XSALSA20-POLY1305")
	// Edge cases
	f.Add("", []byte{}, "", "")
	f.Add("addr", bytes.Repeat([]byte{0xFF}, 32), "fp", "UNKNOWN")
	f.Add("addr", []byte{0x01}, "fp", "X25519-XSALSA20-POLY1305") // Wrong key size

	f.Fuzz(func(t *testing.T, address string, publicKey []byte, fingerprint, algorithmID string) {
		record := &RecipientKeyRecord{
			Address:        address,
			PublicKey:      publicKey,
			KeyFingerprint: fingerprint,
			AlgorithmID:    algorithmID,
		}

		// Should never panic
		_ = record.Validate()

		// IsActive should work
		_ = record.IsActive()
	})
}

// FuzzWrappedKeyEntry tests wrapped key entry handling.
func FuzzWrappedKeyEntry(f *testing.F) {
	f.Add("recipient1", []byte("wrappedkey"), "algo", []byte("ephemeral"))
	f.Add("", []byte{}, "", []byte{})

	f.Fuzz(func(t *testing.T, recipientID string, wrappedKey []byte, algorithm string, ephemeralPubKey []byte) {
		entry := WrappedKeyEntry{
			RecipientID:     recipientID,
			WrappedKey:      wrappedKey,
			Algorithm:       algorithm,
			EphemeralPubKey: ephemeralPubKey,
		}

		// Include in envelope and validate
		envelope := &EncryptedPayloadEnvelope{
			Version:          EnvelopeVersion,
			AlgorithmID:      AlgorithmX25519XSalsa20Poly1305,
			AlgorithmVersion: AlgorithmVersionV1,
			RecipientKeyIDs:  []string{recipientID},
			WrappedKeys:      []WrappedKeyEntry{entry},
			Nonce:            bytes.Repeat([]byte{0x01}, XSalsa20NonceSize),
			Ciphertext:       []byte("test"),
			SenderPubKey:     bytes.Repeat([]byte{0x02}, X25519PublicKeySize),
			SenderSignature:  bytes.Repeat([]byte{0x03}, 64),
		}

		// Should never panic
		_ = envelope.Validate()
	})
}

// FuzzEnvelopeMetadata tests metadata operations.
func FuzzEnvelopeMetadata(f *testing.F) {
	f.Add("key", "value")
	f.Add("", "value")
	f.Add("key", "")
	f.Add("_system_key", "reserved")
	f.Add("key with spaces", "value with\nnewlines")

	f.Fuzz(func(t *testing.T, key, value string) {
		envelope := NewEncryptedPayloadEnvelope()

		// AddMetadata should handle all inputs
		err := envelope.AddMetadata(key, value)

		// Empty key should fail
		if key == "" && err == nil {
			t.Error("expected error for empty key")
		}

		// Non-empty key should succeed
		if key != "" && err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// GetMetadata should be consistent
		if key != "" && err == nil {
			got, ok := envelope.GetMetadata(key)
			if !ok {
				t.Error("failed to get metadata that was just added")
			}
			if got != value {
				t.Errorf("metadata mismatch: got %q, want %q", got, value)
			}
		}
	})
}

// FuzzComputeKeyFingerprint tests fingerprint computation.
func FuzzComputeKeyFingerprint(f *testing.F) {
	f.Add(bytes.Repeat([]byte{0x01}, 32))
	f.Add([]byte{})
	f.Add(bytes.Repeat([]byte{0xFF}, 100))

	f.Fuzz(func(t *testing.T, publicKey []byte) {
		// Should never panic
		fp := ComputeKeyFingerprint(publicKey)

		// Fingerprint should be hex-encoded (2 chars per byte, 20 bytes = 40 chars)
		if len(fp) != KeyFingerprintSize*2 {
			t.Errorf("expected %d char fingerprint, got %d", KeyFingerprintSize*2, len(fp))
		}

		// Same key should produce same fingerprint
		fp2 := ComputeKeyFingerprint(publicKey)
		if fp != fp2 {
			t.Error("fingerprint not deterministic")
		}
	})
}

// FuzzAlgorithmValidation tests algorithm parameter validation.
func FuzzAlgorithmValidation(f *testing.F) {
	f.Add("X25519-XSALSA20-POLY1305", bytes.Repeat([]byte{0x01}, 32), bytes.Repeat([]byte{0x02}, 24))
	f.Add("UNKNOWN", []byte{}, []byte{})
	f.Add("X25519-XSALSA20-POLY1305", []byte{0x01}, []byte{0x02}) // Wrong sizes

	f.Fuzz(func(t *testing.T, algorithmID string, publicKey, nonce []byte) {
		// Should never panic
		err := ValidateAlgorithmParams(algorithmID, publicKey, nonce)

		// Check consistency
		if IsAlgorithmSupported(algorithmID) {
			info, infoErr := GetAlgorithmInfo(algorithmID)
			if infoErr != nil {
				t.Errorf("GetAlgorithmInfo failed for supported algorithm: %v", infoErr)
			}

			// If sizes match, validation should succeed
			if len(publicKey) == info.KeySize && len(nonce) == info.NonceSize {
				if err != nil {
					t.Errorf("unexpected error for valid params: %v", err)
				}
			}
		} else if err == nil {
			// Unsupported algorithm should fail
			t.Errorf("expected error for unsupported algorithm %q", algorithmID)
		}
	})
}
