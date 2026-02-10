// Package crypto provides fuzz tests for envelope encryption operations.
// These tests use Go's native fuzzing support (Go 1.18+) to discover edge cases
// and potential vulnerabilities in cryptographic operations.
//
// Run with: go test -fuzz=. -fuzztime=30s ./x/encryption/crypto/...
//
// Task Reference: VE-2022 - Security audit preparation
package crypto

import (
	"bytes"
	"testing"

	"github.com/virtengine/virtengine/x/encryption/types"
)

// FuzzCreateEnvelope tests envelope creation with arbitrary plaintext.
// This fuzz test verifies that:
// 1. Envelope creation succeeds for any valid plaintext
// 2. The ciphertext differs from plaintext
// 3. The envelope can be decrypted to recover original plaintext
// 4. No panics occur with arbitrary input
func FuzzCreateEnvelope(f *testing.F) {
	// Seed corpus with various plaintext sizes and patterns
	f.Add([]byte(""))                                 // Empty
	f.Add([]byte("a"))                                // Single byte
	f.Add([]byte("Hello, World!"))                    // Simple ASCII
	f.Add([]byte("🔐 Unicode encryption test 密码"))     // Unicode
	f.Add(bytes.Repeat([]byte{0x00}, 100))            // Null bytes
	f.Add(bytes.Repeat([]byte{0xFF}, 100))            // High bytes
	f.Add(bytes.Repeat([]byte("x"), 10000))           // Large payload
	f.Add([]byte{0x00, 0x01, 0x02, 0x03, 0x04})       // Binary sequence
	f.Add([]byte("\x00\x00\x00\x00\x00\x00\x00\x00")) // All zeros

	f.Fuzz(func(t *testing.T, plaintext []byte) {
		// Generate fresh key pairs for each iteration
		sender, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("failed to generate sender key pair: %v", err)
		}

		recipient, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("failed to generate recipient key pair: %v", err)
		}

		// Create envelope
		envelope, err := CreateEnvelope(plaintext, recipient.PublicKey[:], sender)
		if err != nil {
			t.Fatalf("failed to create envelope: %v", err)
		}

		// Verify envelope structure
		if envelope == nil {
			t.Fatal("envelope is nil")
		}

		if envelope.Version != types.EnvelopeVersion {
			t.Errorf("unexpected version: got %d, want %d", envelope.Version, types.EnvelopeVersion)
		}

		if envelope.AlgorithmID != types.AlgorithmX25519XSalsa20Poly1305 {
			t.Errorf("unexpected algorithm: got %s, want %s", envelope.AlgorithmID, types.AlgorithmX25519XSalsa20Poly1305)
		}

		if len(envelope.Nonce) != types.XSalsa20NonceSize {
			t.Errorf("unexpected nonce size: got %d, want %d", len(envelope.Nonce), types.XSalsa20NonceSize)
		}

		// Ciphertext should never equal plaintext (unless empty)
		if len(plaintext) > 0 && bytes.Equal(plaintext, envelope.Ciphertext) {
			t.Error("ciphertext equals plaintext - encryption failed")
		}

		// Ciphertext should be larger than plaintext (includes auth tag)
		if len(envelope.Ciphertext) < len(plaintext) {
			t.Errorf("ciphertext smaller than plaintext: %d < %d", len(envelope.Ciphertext), len(plaintext))
		}

		// Decrypt and verify round-trip
		decrypted, err := OpenEnvelope(envelope, recipient.PrivateKey[:])
		if err != nil {
			t.Fatalf("failed to decrypt envelope: %v", err)
		}

		if !bytes.Equal(plaintext, decrypted) {
			t.Errorf("decrypted data does not match original plaintext")
		}
	})
}

// FuzzOpenEnvelope tests decryption with corrupted ciphertext.
// This fuzz test verifies that:
// 1. Corrupted ciphertext fails authentication
// 2. No panics occur with malformed input
// 3. Error handling is robust
func FuzzOpenEnvelope(f *testing.F) {
	// Seed corpus with various corruption patterns
	f.Add([]byte{0x00})                   // Single null byte
	f.Add([]byte{0xFF})                   // Single high byte
	f.Add([]byte("corrupted"))            // ASCII corruption
	f.Add(bytes.Repeat([]byte{0x41}, 50)) // Repeated bytes

	f.Fuzz(func(t *testing.T, corruption []byte) {
		if len(corruption) == 0 {
			return // Skip empty corruption
		}

		// Create a valid envelope first
		sender, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("failed to generate sender key pair: %v", err)
		}

		recipient, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("failed to generate recipient key pair: %v", err)
		}

		plaintext := []byte("original secret message")
		envelope, err := CreateEnvelope(plaintext, recipient.PublicKey[:], sender)
		if err != nil {
			t.Fatalf("failed to create envelope: %v", err)
		}

		// Corrupt the ciphertext
		originalCiphertext := make([]byte, len(envelope.Ciphertext))
		copy(originalCiphertext, envelope.Ciphertext)

		// XOR corruption into ciphertext
		for i := 0; i < len(corruption) && i < len(envelope.Ciphertext); i++ {
			envelope.Ciphertext[i] ^= corruption[i]
		}

		// Only test if ciphertext was actually modified
		if bytes.Equal(originalCiphertext, envelope.Ciphertext) {
			return
		}

		// Decryption should fail with corrupted ciphertext
		_, err = OpenEnvelope(envelope, recipient.PrivateKey[:])
		if err == nil {
			t.Error("decryption succeeded with corrupted ciphertext - authentication failed")
		}
	})
}

// FuzzNonceUniqueness verifies that nonces are unique across many encryptions.
// This is critical for security - nonce reuse breaks XSalsa20 security.
func FuzzNonceUniqueness(f *testing.F) {
	// Seed with various iteration counts
	f.Add(10)
	f.Add(50)
	f.Add(100)

	f.Fuzz(func(t *testing.T, iterations int) {
		// Limit iterations to avoid timeout
		if iterations < 2 || iterations > 1000 {
			return
		}

		sender, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("failed to generate sender key pair: %v", err)
		}

		recipient, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("failed to generate recipient key pair: %v", err)
		}

		plaintext := []byte("test message for nonce uniqueness")
		nonces := make(map[string]bool)

		for i := 0; i < iterations; i++ {
			envelope, err := CreateEnvelope(plaintext, recipient.PublicKey[:], sender)
			if err != nil {
				t.Fatalf("failed to create envelope %d: %v", i, err)
			}

			nonceStr := string(envelope.Nonce)
			if nonces[nonceStr] {
				t.Fatalf("nonce collision detected at iteration %d", i)
			}
			nonces[nonceStr] = true
		}
	})
}

// FuzzMultiRecipientEnvelope tests multi-recipient encryption with varying recipient counts.
func FuzzMultiRecipientEnvelope(f *testing.F) {
	// Seed with various recipient counts
	f.Add(1, []byte("secret"))
	f.Add(2, []byte("two recipients"))
	f.Add(5, []byte("five recipients"))
	f.Add(10, []byte("ten recipients"))

	f.Fuzz(func(t *testing.T, recipientCount int, plaintext []byte) {
		// Limit recipients to reasonable range
		if recipientCount < 1 || recipientCount > 20 {
			return
		}

		sender, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("failed to generate sender key pair: %v", err)
		}

		// Generate recipient key pairs
		recipients := make([]*KeyPair, recipientCount)
		recipientPubKeys := make([][]byte, recipientCount)
		for i := range recipients {
			recipients[i], err = GenerateKeyPair()
			if err != nil {
				t.Fatalf("failed to generate recipient key pair %d: %v", i, err)
			}
			recipientPubKeys[i] = recipients[i].PublicKey[:]
		}

		// Create multi-recipient envelope
		envelope, err := CreateMultiRecipientEnvelope(plaintext, recipientPubKeys, sender)
		if err != nil {
			t.Fatalf("failed to create multi-recipient envelope: %v", err)
		}

		// Verify structure
		if len(envelope.RecipientKeyIDs) != recipientCount {
			t.Errorf("unexpected recipient count: got %d, want %d", len(envelope.RecipientKeyIDs), recipientCount)
		}

		// Each recipient should be able to decrypt
		for i, recipient := range recipients {
			decrypted, err := OpenEnvelope(envelope, recipient.PrivateKey[:])
			if err != nil {
				t.Fatalf("recipient %d failed to decrypt: %v", i, err)
			}

			if !bytes.Equal(plaintext, decrypted) {
				t.Errorf("recipient %d: decrypted data does not match original plaintext", i)
			}
		}

		// Wrong key should fail to decrypt
		wrongRecipient, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("failed to generate wrong recipient key pair: %v", err)
		}

		_, err = OpenEnvelope(envelope, wrongRecipient.PrivateKey[:])
		if err == nil {
			t.Error("wrong recipient was able to decrypt - security failure")
		}
	})
}

// FuzzInvalidKeySize tests handling of invalid key sizes.
// This verifies proper input validation.
func FuzzInvalidKeySize(f *testing.F) {
	// Seed with various invalid key sizes
	f.Add(0)
	f.Add(1)
	f.Add(16)
	f.Add(31)
	f.Add(33)
	f.Add(64)

	f.Fuzz(func(t *testing.T, keySize int) {
		// Skip valid key size
		if keySize == 32 {
			return
		}

		// Limit to reasonable range
		if keySize < 0 || keySize > 256 {
			return
		}

		sender, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("failed to generate sender key pair: %v", err)
		}

		invalidKey := make([]byte, keySize)
		_, err = CreateEnvelope([]byte("test"), invalidKey, sender)
		if err == nil {
			t.Errorf("expected error for invalid key size %d, got nil", keySize)
		}
	})
}

// FuzzKeyPairGeneration tests that key pair generation produces valid keys.
func FuzzKeyPairGeneration(f *testing.F) {
	// Seed with iteration counts
	f.Add(5)
	f.Add(20)
	f.Add(50)

	f.Fuzz(func(t *testing.T, iterations int) {
		if iterations < 1 || iterations > 100 {
			return
		}

		keyPairs := make([]*KeyPair, iterations)
		for i := 0; i < iterations; i++ {
			kp, err := GenerateKeyPair()
			if err != nil {
				t.Fatalf("failed to generate key pair %d: %v", i, err)
			}

			// Verify key sizes
			if len(kp.PublicKey) != 32 {
				t.Errorf("key %d: invalid public key size: %d", i, len(kp.PublicKey))
			}
			if len(kp.PrivateKey) != 32 {
				t.Errorf("key %d: invalid private key size: %d", i, len(kp.PrivateKey))
			}

			// Keys should not be all zeros
			allZero := true
			for _, b := range kp.PublicKey {
				if b != 0 {
					allZero = false
					break
				}
			}
			if allZero {
				t.Errorf("key %d: public key is all zeros", i)
			}

			// Fingerprint should be consistent
			fp1 := kp.Fingerprint()
			fp2 := kp.Fingerprint()
			if fp1 != fp2 {
				t.Errorf("key %d: fingerprint not consistent", i)
			}

			keyPairs[i] = kp
		}

		// Check for duplicate keys
		seen := make(map[[32]byte]bool)
		for i, kp := range keyPairs {
			if seen[kp.PublicKey] {
				t.Errorf("key %d: duplicate public key detected", i)
			}
			seen[kp.PublicKey] = true
		}
	})
}

// FuzzAlgorithmEncryption tests the Algorithm interface implementation.
func FuzzAlgorithmEncryption(f *testing.F) {
	// Seed with various plaintext
	f.Add([]byte("algorithm test"))
	f.Add([]byte(""))
	f.Add(bytes.Repeat([]byte{0xAB}, 1000))

	f.Fuzz(func(t *testing.T, plaintext []byte) {
		algo := NewX25519XSalsa20Poly1305()

		// Verify algorithm properties
		if algo.ID() != types.AlgorithmX25519XSalsa20Poly1305 {
			t.Errorf("unexpected algorithm ID: %s", algo.ID())
		}
		if algo.KeySize() != 32 {
			t.Errorf("unexpected key size: %d", algo.KeySize())
		}
		if algo.NonceSize() != 24 {
			t.Errorf("unexpected nonce size: %d", algo.NonceSize())
		}

		// Generate key pairs using algorithm
		sender, err := algo.GenerateKeyPair()
		if err != nil {
			t.Fatalf("failed to generate sender key pair: %v", err)
		}

		recipient, err := algo.GenerateKeyPair()
		if err != nil {
			t.Fatalf("failed to generate recipient key pair: %v", err)
		}

		// Encrypt
		ciphertext, nonce, err := algo.Encrypt(plaintext, recipient.PublicKey[:], sender)
		if err != nil {
			t.Fatalf("encryption failed: %v", err)
		}

		if len(nonce) != algo.NonceSize() {
			t.Errorf("nonce size mismatch: got %d, want %d", len(nonce), algo.NonceSize())
		}

		// Decrypt
		decrypted, err := algo.Decrypt(ciphertext, nonce, sender.PublicKey[:], recipient.PrivateKey[:])
		if err != nil {
			t.Fatalf("decryption failed: %v", err)
		}

		if !bytes.Equal(plaintext, decrypted) {
			t.Error("decrypted data does not match plaintext")
		}
	})
}
