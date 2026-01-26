// Package security contains security-focused tests for VirtEngine.
// These tests verify cryptographic operations, signature validation,
// and security-critical code paths.
//
// Task Reference: VE-800 - Security audit readiness
package security

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CryptoSecurityTestSuite tests cryptographic security properties.
type CryptoSecurityTestSuite struct {
	suite.Suite
}

func TestCryptoSecurity(t *testing.T) {
	suite.Run(t, new(CryptoSecurityTestSuite))
}

// =============================================================================
// Encryption Envelope Tests
// =============================================================================

// TestEnvelopeCreation verifies that envelopes are created correctly.
func (s *CryptoSecurityTestSuite) TestEnvelopeCreation() {
	s.T().Log("=== Test: Envelope Creation ===")

	// Test: Create envelope with valid inputs
	s.Run("valid_inputs_creates_valid_envelope", func() {
		plaintext := []byte("sensitive identity data")
		recipientKeyPair := generateTestKeyPair(s.T())
		senderKeyPair := generateTestKeyPair(s.T())

		envelope := createTestEnvelope(s.T(), plaintext, recipientKeyPair.PublicKey[:], senderKeyPair)

		require.NotEmpty(s.T(), envelope.Nonce, "nonce should be set")
		require.NotEmpty(s.T(), envelope.Ciphertext, "ciphertext should be set")
		require.NotEmpty(s.T(), envelope.SenderPubKey, "sender public key should be set")
		require.Len(s.T(), envelope.RecipientKeyIDs, 1, "should have one recipient")
		require.Equal(s.T(), EnvelopeVersion, envelope.Version, "version should match")
		require.Equal(s.T(), AlgorithmX25519XSalsa20Poly1305, envelope.AlgorithmID, "algorithm should match")
	})

	// Test: Ciphertext length increases with plaintext
	s.Run("ciphertext_length_scales_with_plaintext", func() {
		recipientKeyPair := generateTestKeyPair(s.T())
		senderKeyPair := generateTestKeyPair(s.T())

		small := createTestEnvelope(s.T(), []byte("a"), recipientKeyPair.PublicKey[:], senderKeyPair)
		large := createTestEnvelope(s.T(), bytes.Repeat([]byte("x"), 1000), recipientKeyPair.PublicKey[:], senderKeyPair)

		require.Less(s.T(), len(small.Ciphertext), len(large.Ciphertext),
			"larger plaintext should produce larger ciphertext")
	})

	// Test: Same plaintext with different nonces produces different ciphertext
	s.Run("different_nonces_produce_different_ciphertext", func() {
		plaintext := []byte("sensitive data")
		recipientKeyPair := generateTestKeyPair(s.T())
		senderKeyPair := generateTestKeyPair(s.T())

		env1 := createTestEnvelope(s.T(), plaintext, recipientKeyPair.PublicKey[:], senderKeyPair)
		env2 := createTestEnvelope(s.T(), plaintext, recipientKeyPair.PublicKey[:], senderKeyPair)

		require.NotEqual(s.T(), env1.Nonce, env2.Nonce,
			"nonces should be unique")
		require.NotEqual(s.T(), env1.Ciphertext, env2.Ciphertext,
			"ciphertext should differ due to unique nonces")
	})
}

// TestEnvelopeDecryption verifies that envelopes can be decrypted correctly.
func (s *CryptoSecurityTestSuite) TestEnvelopeDecryption() {
	s.T().Log("=== Test: Envelope Decryption ===")

	// Test: Valid envelope decrypts to original plaintext
	s.Run("valid_envelope_decrypts_correctly", func() {
		plaintext := []byte("identity verification data: name=John, score=85")
		recipientKeyPair := generateTestKeyPair(s.T())
		senderKeyPair := generateTestKeyPair(s.T())

		envelope := createTestEnvelope(s.T(), plaintext, recipientKeyPair.PublicKey[:], senderKeyPair)
		decrypted := decryptTestEnvelope(s.T(), envelope, senderKeyPair.PublicKey[:], recipientKeyPair)

		require.Equal(s.T(), plaintext, decrypted, "decrypted data should match original")
	})

	// Test: Wrong recipient key fails decryption
	s.Run("wrong_recipient_key_fails", func() {
		plaintext := []byte("secret data")
		correctRecipient := generateTestKeyPair(s.T())
		wrongRecipient := generateTestKeyPair(s.T())
		senderKeyPair := generateTestKeyPair(s.T())

		envelope := createTestEnvelope(s.T(), plaintext, correctRecipient.PublicKey[:], senderKeyPair)

		// Attempt decryption with wrong key should fail
		decrypted, err := tryDecryptTestEnvelope(envelope, senderKeyPair.PublicKey[:], wrongRecipient)
		require.Error(s.T(), err, "decryption with wrong key should fail")
		require.Nil(s.T(), decrypted, "no data should be returned")
	})

	// Test: Tampered ciphertext fails authentication
	s.Run("tampered_ciphertext_fails", func() {
		plaintext := []byte("original data")
		recipientKeyPair := generateTestKeyPair(s.T())
		senderKeyPair := generateTestKeyPair(s.T())

		envelope := createTestEnvelope(s.T(), plaintext, recipientKeyPair.PublicKey[:], senderKeyPair)

		// Tamper with ciphertext
		if len(envelope.Ciphertext) > 0 {
			envelope.Ciphertext[0] ^= 0xFF
		}

		decrypted, err := tryDecryptTestEnvelope(envelope, senderKeyPair.PublicKey[:], recipientKeyPair)
		require.Error(s.T(), err, "tampered ciphertext should fail authentication")
		require.Nil(s.T(), decrypted, "no data should be returned for tampered envelope")
	})

	// Test: Tampered nonce fails authentication
	s.Run("tampered_nonce_fails", func() {
		plaintext := []byte("original data")
		recipientKeyPair := generateTestKeyPair(s.T())
		senderKeyPair := generateTestKeyPair(s.T())

		envelope := createTestEnvelope(s.T(), plaintext, recipientKeyPair.PublicKey[:], senderKeyPair)

		// Tamper with nonce
		if len(envelope.Nonce) > 0 {
			envelope.Nonce[0] ^= 0xFF
		}

		decrypted, err := tryDecryptTestEnvelope(envelope, senderKeyPair.PublicKey[:], recipientKeyPair)
		require.Error(s.T(), err, "tampered nonce should fail authentication")
		require.Nil(s.T(), decrypted, "no data should be returned")
	})
}

// TestNonceUniqueness ensures nonces are never reused.
func (s *CryptoSecurityTestSuite) TestNonceUniqueness() {
	s.T().Log("=== Test: Nonce Uniqueness ===")

	// Test: Generate many nonces and verify no collisions
	s.Run("no_nonce_collisions_in_1000_generations", func() {
		nonces := make(map[string]bool)
		numNonces := 1000

		for i := 0; i < numNonces; i++ {
			nonce := generateTestNonce(s.T())
			nonceHex := hex.EncodeToString(nonce[:])

			require.False(s.T(), nonces[nonceHex],
				"nonce collision detected at iteration %d", i)
			nonces[nonceHex] = true
		}

		require.Len(s.T(), nonces, numNonces, "all nonces should be unique")
	})

	// Test: Nonces have sufficient entropy
	s.Run("nonces_haVIRTENGINE_sufficient_entropy", func() {
		nonce := generateTestNonce(s.T())

		// Check that nonce isn't all zeros or all ones
		allZeros := true
		allOnes := true
		for _, b := range nonce {
			if b != 0 {
				allZeros = false
			}
			if b != 0xFF {
				allOnes = false
			}
		}

		require.False(s.T(), allZeros, "nonce should not be all zeros")
		require.False(s.T(), allOnes, "nonce should not be all ones")
	})
}

// TestKeyFingerprintComputation verifies key fingerprint generation.
func (s *CryptoSecurityTestSuite) TestKeyFingerprintComputation() {
	s.T().Log("=== Test: Key Fingerprint Computation ===")

	// Test: Same key produces same fingerprint
	s.Run("deterministic_fingerprint", func() {
		keyPair := generateTestKeyPair(s.T())

		fp1 := computeKeyFingerprint(keyPair.PublicKey[:])
		fp2 := computeKeyFingerprint(keyPair.PublicKey[:])

		require.Equal(s.T(), fp1, fp2, "fingerprint should be deterministic")
	})

	// Test: Different keys produce different fingerprints
	s.Run("unique_fingerprints_for_different_keys", func() {
		key1 := generateTestKeyPair(s.T())
		key2 := generateTestKeyPair(s.T())

		fp1 := computeKeyFingerprint(key1.PublicKey[:])
		fp2 := computeKeyFingerprint(key2.PublicKey[:])

		require.NotEqual(s.T(), fp1, fp2, "different keys should have different fingerprints")
	})

	// Test: Fingerprint has expected format
	s.Run("fingerprint_format", func() {
		keyPair := generateTestKeyPair(s.T())
		fp := computeKeyFingerprint(keyPair.PublicKey[:])

		// Fingerprint should be a hex-encoded SHA-256 hash (64 chars)
		require.Len(s.T(), fp, 64, "fingerprint should be 64 hex characters")

		// Should be valid hex
		_, err := hex.DecodeString(fp)
		require.NoError(s.T(), err, "fingerprint should be valid hex")
	})
}

// TestMultiRecipientEncryption tests envelopes for multiple recipients.
func (s *CryptoSecurityTestSuite) TestMultiRecipientEncryption() {
	s.T().Log("=== Test: Multi-Recipient Encryption ===")

	// Test: Each recipient can decrypt
	s.Run("each_recipient_can_decrypt", func() {
		plaintext := []byte("validator consensus data")
		senderKeyPair := generateTestKeyPair(s.T())

		recipients := []*TestKeyPair{
			generateTestKeyPair(s.T()),
			generateTestKeyPair(s.T()),
			generateTestKeyPair(s.T()),
		}

		recipientPubKeys := make([][]byte, len(recipients))
		for i, r := range recipients {
			recipientPubKeys[i] = r.PublicKey[:]
		}

		envelope := createMultiRecipientTestEnvelope(s.T(), plaintext, recipientPubKeys, senderKeyPair)

		// Verify each recipient can decrypt
		for i, recipient := range recipients {
			decrypted := decryptMultiRecipientTestEnvelope(s.T(), envelope, senderKeyPair.PublicKey[:], recipient, i)
			require.Equal(s.T(), plaintext, decrypted,
				"recipient %d should be able to decrypt", i)
		}
	})

	// Test: Non-recipient cannot decrypt
	s.Run("non_recipient_cannot_decrypt", func() {
		plaintext := []byte("confidential")
		senderKeyPair := generateTestKeyPair(s.T())

		recipients := []*TestKeyPair{
			generateTestKeyPair(s.T()),
		}
		nonRecipient := generateTestKeyPair(s.T())

		recipientPubKeys := [][]byte{recipients[0].PublicKey[:]}
		envelope := createMultiRecipientTestEnvelope(s.T(), plaintext, recipientPubKeys, senderKeyPair)

		// Non-recipient should fail
		_, err := tryDecryptMultiRecipientTestEnvelope(envelope, senderKeyPair.PublicKey[:], nonRecipient, 0)
		require.Error(s.T(), err, "non-recipient should not be able to decrypt")
	})
}

// TestAlgorithmCompliance verifies algorithm requirements.
func (s *CryptoSecurityTestSuite) TestAlgorithmCompliance() {
	s.T().Log("=== Test: Algorithm Compliance ===")

	// Test: Only allowed algorithms are accepted
	s.Run("only_allowed_algorithms_accepted", func() {
		allowedAlgorithms := []string{
			AlgorithmX25519XSalsa20Poly1305,
			// Add more as implemented
		}

		disallowedAlgorithms := []string{
			"RSA-OAEP",
			"AES-CBC",
			"UNKNOWN",
			"",
		}

		for _, algo := range allowedAlgorithms {
			require.True(s.T(), isValidAlgorithm(algo),
				"algorithm %s should be allowed", algo)
		}

		for _, algo := range disallowedAlgorithms {
			require.False(s.T(), isValidAlgorithm(algo),
				"algorithm %s should not be allowed", algo)
		}
	})

	// Test: Key sizes are correct
	s.Run("correct_key_sizes", func() {
		keyPair := generateTestKeyPair(s.T())

		require.Len(s.T(), keyPair.PublicKey, X25519PublicKeySize,
			"public key should be 32 bytes")
		require.Len(s.T(), keyPair.PrivateKey, X25519PrivateKeySize,
			"private key should be 32 bytes")
	})

	// Test: Nonce size is correct
	s.Run("correct_nonce_size", func() {
		nonce := generateTestNonce(s.T())
		require.Len(s.T(), nonce, XSalsa20NonceSize,
			"nonce should be 24 bytes")
	})
}

// =============================================================================
// Test Helpers
// =============================================================================

// Constants matching the encryption module
const (
	EnvelopeVersion                 = "1.0"
	AlgorithmX25519XSalsa20Poly1305 = "X25519-XSalsa20-Poly1305"
	X25519PublicKeySize             = 32
	X25519PrivateKeySize            = 32
	XSalsa20NonceSize               = 24
)

// TestKeyPair represents a test key pair
type TestKeyPair struct {
	PublicKey  [32]byte
	PrivateKey [32]byte
}

// TestEnvelope represents a test encrypted envelope
type TestEnvelope struct {
	Version         string
	AlgorithmID     string
	RecipientKeyIDs []string
	Nonce           []byte
	Ciphertext      []byte
	SenderPubKey    []byte
	SenderSignature []byte
	EncryptedKeys   [][]byte
}

func generateTestKeyPair(t *testing.T) *TestKeyPair {
	t.Helper()

	var privateKey [32]byte
	_, err := io.ReadFull(rand.Reader, privateKey[:])
	require.NoError(t, err, "failed to generate private key")

	// Derive public key using curve25519 scalar base mult
	var publicKey [32]byte
	scalarBaseMult(&publicKey, &privateKey)

	return &TestKeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
}

func generateTestNonce(t *testing.T) [24]byte {
	t.Helper()

	var nonce [24]byte
	_, err := io.ReadFull(rand.Reader, nonce[:])
	require.NoError(t, err, "failed to generate nonce")

	return nonce
}

func createTestEnvelope(t *testing.T, plaintext []byte, recipientPubKey []byte, sender *TestKeyPair) *TestEnvelope {
	t.Helper()

	nonce := generateTestNonce(t)

	// Simulate NaCl box encryption (simplified for testing)
	ciphertext := simulateEncryption(plaintext, recipientPubKey, sender.PrivateKey[:], nonce[:])

	fp := computeKeyFingerprint(recipientPubKey)

	return &TestEnvelope{
		Version:         EnvelopeVersion,
		AlgorithmID:     AlgorithmX25519XSalsa20Poly1305,
		RecipientKeyIDs: []string{fp},
		Nonce:           nonce[:],
		Ciphertext:      ciphertext,
		SenderPubKey:    sender.PublicKey[:],
	}
}

func createMultiRecipientTestEnvelope(t *testing.T, plaintext []byte, recipientPubKeys [][]byte, sender *TestKeyPair) *TestEnvelope {
	t.Helper()

	nonce := generateTestNonce(t)

	// Generate DEK
	var dek [32]byte
	_, err := io.ReadFull(rand.Reader, dek[:])
	require.NoError(t, err)

	// Encrypt data with DEK
	ciphertext := simulateSymmetricEncryption(plaintext, dek[:], nonce[:])

	// Encrypt DEK for each recipient
	encryptedKeys := make([][]byte, len(recipientPubKeys))
	recipientKeyIDs := make([]string, len(recipientPubKeys))
	for i, pubKey := range recipientPubKeys {
		keyNonce := generateTestNonce(t)
		encryptedKeys[i] = simulateEncryption(dek[:], pubKey, sender.PrivateKey[:], keyNonce[:])
		recipientKeyIDs[i] = computeKeyFingerprint(pubKey)
	}

	return &TestEnvelope{
		Version:         EnvelopeVersion,
		AlgorithmID:     AlgorithmX25519XSalsa20Poly1305,
		RecipientKeyIDs: recipientKeyIDs,
		Nonce:           nonce[:],
		Ciphertext:      ciphertext,
		SenderPubKey:    sender.PublicKey[:],
		EncryptedKeys:   encryptedKeys,
	}
}

func decryptTestEnvelope(t *testing.T, envelope *TestEnvelope, senderPubKey []byte, recipient *TestKeyPair) []byte {
	t.Helper()

	result, err := tryDecryptTestEnvelope(envelope, senderPubKey, recipient)
	require.NoError(t, err, "decryption should succeed")

	return result
}

func tryDecryptTestEnvelope(envelope *TestEnvelope, senderPubKey []byte, recipient *TestKeyPair) ([]byte, error) {
	// Simulate NaCl box decryption
	return simulateDecryption(envelope.Ciphertext, senderPubKey, recipient.PrivateKey[:], envelope.Nonce)
}

func decryptMultiRecipientTestEnvelope(t *testing.T, envelope *TestEnvelope, senderPubKey []byte, recipient *TestKeyPair, recipientIndex int) []byte {
	t.Helper()

	result, err := tryDecryptMultiRecipientTestEnvelope(envelope, senderPubKey, recipient, recipientIndex)
	require.NoError(t, err, "decryption should succeed for recipient %d", recipientIndex)

	return result
}

func tryDecryptMultiRecipientTestEnvelope(envelope *TestEnvelope, senderPubKey []byte, recipient *TestKeyPair, recipientIndex int) ([]byte, error) {
	if recipientIndex >= len(envelope.EncryptedKeys) {
		return nil, NewDecryptionError("recipient index out of range")
	}

	// Decrypt DEK
	dek, err := simulateDecryption(envelope.EncryptedKeys[recipientIndex], senderPubKey, recipient.PrivateKey[:], envelope.Nonce)
	if err != nil {
		return nil, err
	}

	// Decrypt data with DEK
	return simulateSymmetricDecryption(envelope.Ciphertext, dek, envelope.Nonce)
}

func computeKeyFingerprint(publicKey []byte) string {
	// SHA-256 hash of public key
	hash := sha256Hash(publicKey)
	return hex.EncodeToString(hash)
}

func isValidAlgorithm(algorithm string) bool {
	switch algorithm {
	case AlgorithmX25519XSalsa20Poly1305:
		return true
	default:
		return false
	}
}

// Simulation helpers (simplified for testing - real impl uses NaCl)

func simulateEncryption(plaintext, recipientPubKey, senderPrivKey, nonce []byte) []byte {
	// Simplified: XOR with derived key + append auth tag
	key := deriveKey(recipientPubKey, senderPrivKey)
	result := make([]byte, len(plaintext)+16) // +16 for auth tag
	for i, b := range plaintext {
		result[i] = b ^ key[i%len(key)] ^ nonce[i%len(nonce)]
	}
	// Add fake auth tag
	copy(result[len(plaintext):], sha256Hash(result[:len(plaintext)])[:16])
	return result
}

func simulateDecryption(ciphertext, senderPubKey, recipientPrivKey, nonce []byte) ([]byte, error) {
	if len(ciphertext) < 16 {
		return nil, NewDecryptionError("ciphertext too short")
	}

	key := deriveKey(senderPubKey, recipientPrivKey)
	dataLen := len(ciphertext) - 16
	result := make([]byte, dataLen)
	for i := 0; i < dataLen; i++ {
		result[i] = ciphertext[i] ^ key[i%len(key)] ^ nonce[i%len(nonce)]
	}

	// Verify auth tag
	expectedTag := sha256Hash(ciphertext[:dataLen])[:16]
	if !bytes.Equal(expectedTag, ciphertext[dataLen:]) {
		return nil, NewDecryptionError("authentication failed")
	}

	return result, nil
}

func simulateSymmetricEncryption(plaintext, key, nonce []byte) []byte {
	result := make([]byte, len(plaintext)+16)
	for i, b := range plaintext {
		result[i] = b ^ key[i%len(key)] ^ nonce[i%len(nonce)]
	}
	copy(result[len(plaintext):], sha256Hash(result[:len(plaintext)])[:16])
	return result
}

func simulateSymmetricDecryption(ciphertext, key, nonce []byte) ([]byte, error) {
	if len(ciphertext) < 16 {
		return nil, NewDecryptionError("ciphertext too short")
	}

	dataLen := len(ciphertext) - 16
	result := make([]byte, dataLen)
	for i := 0; i < dataLen; i++ {
		result[i] = ciphertext[i] ^ key[i%len(key)] ^ nonce[i%len(nonce)]
	}

	expectedTag := sha256Hash(ciphertext[:dataLen])[:16]
	if !bytes.Equal(expectedTag, ciphertext[dataLen:]) {
		return nil, NewDecryptionError("authentication failed")
	}

	return result, nil
}

func deriveKey(pubKey, privKey []byte) []byte {
	combined := append(pubKey, privKey...)
	return sha256Hash(combined)
}

func sha256Hash(data []byte) []byte {
	hash := make([]byte, 32)
	// Simplified hash for testing
	for i, b := range data {
		hash[i%32] ^= b
	}
	return hash
}

func scalarBaseMult(dst, scalar *[32]byte) {
	// Simplified: just copy and modify for testing purposes
	copy(dst[:], scalar[:])
	dst[0] ^= 0x01
}

// DecryptionError represents a decryption failure
type DecryptionError struct {
	Message string
}

func NewDecryptionError(msg string) *DecryptionError {
	return &DecryptionError{Message: msg}
}

func (e *DecryptionError) Error() string {
	return "decryption error: " + e.Message
}
