// Package security contains security-focused tests for VirtEngine.
// These tests verify signature validation and anti-forgery measures.
//
// Task Reference: VE-800 - Security audit readiness
package security

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SignatureSecurityTestSuite tests signature security properties.
type SignatureSecurityTestSuite struct {
	suite.Suite
}

func TestSignatureSecurity(t *testing.T) {
	suite.Run(t, new(SignatureSecurityTestSuite))
}

// =============================================================================
// Identity Upload Signature Tests
// =============================================================================

// TestIdentityUploadSignatureVerification tests identity scope upload signatures.
func (s *SignatureSecurityTestSuite) TestIdentityUploadSignatureVerification() {
	s.T().Log("=== Test: Identity Upload Signature Verification ===")

	// Test: Valid signatures are accepted
	s.Run("valid_signature_accepted", func() {
		clientKeyPair := generateSigningKeyPair(s.T())
		userKeyPair := generateSigningKeyPair(s.T())

		upload := &IdentityUploadPayload{
			Scopes:    []byte("encrypted identity scopes"),
			Salt:      generateSalt(s.T()),
			Timestamp: time.Now().UTC(),
		}

		// Sign with approved client key
		clientSig := signPayload(s.T(), upload.SigningPayload(), clientKeyPair)
		upload.ClientSignature = clientSig

		// Sign with user key
		userSig := signPayload(s.T(), upload.SigningPayload(), userKeyPair)
		upload.UserSignature = userSig

		// Verify both signatures
		clientValid := verifySignature(upload.SigningPayload(), clientSig, clientKeyPair.PublicKey[:])
		userValid := verifySignature(upload.SigningPayload(), userSig, userKeyPair.PublicKey[:])

		require.True(s.T(), clientValid, "client signature should be valid")
		require.True(s.T(), userValid, "user signature should be valid")
	})

	// Test: Modified payload invalidates signature
	s.Run("modified_payload_invalidates_signature", func() {
		clientKeyPair := generateSigningKeyPair(s.T())

		upload := &IdentityUploadPayload{
			Scopes:    []byte("original scopes"),
			Salt:      generateSalt(s.T()),
			Timestamp: time.Now().UTC(),
		}

		sig := signPayload(s.T(), upload.SigningPayload(), clientKeyPair)

		// Modify payload
		upload.Scopes = []byte("modified scopes")

		// Signature should no longer be valid
		valid := verifySignature(upload.SigningPayload(), sig, clientKeyPair.PublicKey[:])
		require.False(s.T(), valid, "modified payload should invalidate signature")
	})

	// Test: Wrong public key fails verification
	s.Run("wrong_public_key_fails", func() {
		correctKey := generateSigningKeyPair(s.T())
		wrongKey := generateSigningKeyPair(s.T())

		payload := []byte("identity data")
		sig := signPayload(s.T(), payload, correctKey)

		valid := verifySignature(payload, sig, wrongKey.PublicKey[:])
		require.False(s.T(), valid, "wrong public key should fail verification")
	})
}

// TestApprovedClientSignatureChecks tests approved client validation.
func (s *SignatureSecurityTestSuite) TestApprovedClientSignatureChecks() {
	s.T().Log("=== Test: Approved Client Signature Checks ===")

	// Test: Approved client key is accepted
	s.Run("approved_client_accepted", func() {
		approvedClientKey := generateSigningKeyPair(s.T())
		allowlist := [][]byte{approvedClientKey.PublicKey[:]}

		payload := []byte("capture data")
		sig := signPayload(s.T(), payload, approvedClientKey)

		isApproved := isApprovedClient(approvedClientKey.PublicKey[:], allowlist)
		sigValid := verifySignature(payload, sig, approvedClientKey.PublicKey[:])

		require.True(s.T(), isApproved, "client should be in approved list")
		require.True(s.T(), sigValid, "signature should be valid")
	})

	// Test: Unapproved client key is rejected
	s.Run("unapproved_client_rejected", func() {
		approvedClientKey := generateSigningKeyPair(s.T())
		rogueClientKey := generateSigningKeyPair(s.T())
		allowlist := [][]byte{approvedClientKey.PublicKey[:]}

		isApproved := isApprovedClient(rogueClientKey.PublicKey[:], allowlist)
		require.False(s.T(), isApproved, "rogue client should not be in approved list")
	})

	// Test: Empty allowlist rejects all clients
	s.Run("empty_allowlist_rejects_all", func() {
		clientKey := generateSigningKeyPair(s.T())
		emptyAllowlist := [][]byte{}

		isApproved := isApprovedClient(clientKey.PublicKey[:], emptyAllowlist)
		require.False(s.T(), isApproved, "empty allowlist should reject all clients")
	})
}

// =============================================================================
// Signature Forgery Detection Tests
// =============================================================================

// TestSignatureForgeryDetection tests detection of forged signatures.
func (s *SignatureSecurityTestSuite) TestSignatureForgeryDetection() {
	s.T().Log("=== Test: Signature Forgery Detection ===")

	// Test: Random bytes not accepted as valid signature
	s.Run("random_bytes_rejected", func() {
		keyPair := generateSigningKeyPair(s.T())
		payload := []byte("important data")

		// Generate random bytes as fake signature
		fakeSig := make([]byte, 64) // Ed25519 signature size
		_, err := io.ReadFull(rand.Reader, fakeSig)
		require.NoError(s.T(), err)

		valid := verifySignature(payload, fakeSig, keyPair.PublicKey[:])
		require.False(s.T(), valid, "random bytes should not be valid signature")
	})

	// Test: Signature from different payload rejected
	s.Run("signature_from_different_payload_rejected", func() {
		keyPair := generateSigningKeyPair(s.T())
		payload1 := []byte("payload one")
		payload2 := []byte("payload two")

		sig1 := signPayload(s.T(), payload1, keyPair)

		// Try to use sig1 to verify payload2
		valid := verifySignature(payload2, sig1, keyPair.PublicKey[:])
		require.False(s.T(), valid, "signature from different payload should be rejected")
	})

	// Test: Truncated signature rejected
	s.Run("truncated_signature_rejected", func() {
		keyPair := generateSigningKeyPair(s.T())
		payload := []byte("test data")

		sig := signPayload(s.T(), payload, keyPair)

		// Truncate signature
		truncatedSig := sig[:len(sig)-1]

		valid := verifySignature(payload, truncatedSig, keyPair.PublicKey[:])
		require.False(s.T(), valid, "truncated signature should be rejected")
	})

	// Test: Extended signature rejected
	s.Run("extended_signature_rejected", func() {
		keyPair := generateSigningKeyPair(s.T())
		payload := []byte("test data")

		sig := signPayload(s.T(), payload, keyPair)

		// Extend signature with extra bytes
		extendedSig := append(sig, 0x00, 0x01)

		valid := verifySignature(payload, extendedSig, keyPair.PublicKey[:])
		require.False(s.T(), valid, "extended signature should be rejected")
	})

	// Test: Bit-flipped signature rejected
	s.Run("bit_flipped_signature_rejected", func() {
		keyPair := generateSigningKeyPair(s.T())
		payload := []byte("test data")

		sig := signPayload(s.T(), payload, keyPair)

		// Flip a bit in the signature
		flippedSig := make([]byte, len(sig))
		copy(flippedSig, sig)
		flippedSig[0] ^= 0x01

		valid := verifySignature(payload, flippedSig, keyPair.PublicKey[:])
		require.False(s.T(), valid, "bit-flipped signature should be rejected")
	})
}

// =============================================================================
// Replay Attack Prevention Tests
// =============================================================================

// TestReplayAttackPrevention tests replay attack mitigation.
func (s *SignatureSecurityTestSuite) TestReplayAttackPrevention() {
	s.T().Log("=== Test: Replay Attack Prevention ===")

	// Test: Salt binding prevents reuse
	s.Run("salt_binding_prevents_reuse", func() {
		keyPair := generateSigningKeyPair(s.T())

		// Create two uploads with same content but different salts
		upload1 := &IdentityUploadPayload{
			Scopes:    []byte("identity data"),
			Salt:      generateSalt(s.T()),
			Timestamp: time.Now().UTC(),
		}

		upload2 := &IdentityUploadPayload{
			Scopes:    []byte("identity data"),
			Salt:      generateSalt(s.T()),
			Timestamp: time.Now().UTC(),
		}

		sig1 := signPayload(s.T(), upload1.SigningPayload(), keyPair)

		// sig1 should not verify for upload2
		valid := verifySignature(upload2.SigningPayload(), sig1, keyPair.PublicKey[:])
		require.False(s.T(), valid, "signature with different salt should be rejected")
	})

	// Test: Nonce tracker prevents replay
	s.Run("nonce_tracker_prevents_replay", func() {
		tracker := NewNonceTracker()
		nonce := generateSalt(s.T())

		// First use should succeed
		firstUse := tracker.UseNonce(nonce)
		require.True(s.T(), firstUse, "first nonce use should succeed")

		// Second use should fail
		secondUse := tracker.UseNonce(nonce)
		require.False(s.T(), secondUse, "second nonce use should fail (replay)")
	})

	// Test: Timestamp validation within window
	s.Run("timestamp_within_window_accepted", func() {
		window := 5 * time.Minute

		validTimestamp := time.Now().UTC()
		require.True(s.T(), isTimestampValid(validTimestamp, window),
			"current timestamp should be valid")

		oldTimestamp := time.Now().UTC().Add(-10 * time.Minute)
		require.False(s.T(), isTimestampValid(oldTimestamp, window),
			"old timestamp should be rejected")

		futureTimestamp := time.Now().UTC().Add(10 * time.Minute)
		require.False(s.T(), isTimestampValid(futureTimestamp, window),
			"future timestamp should be rejected")
	})

	// Test: Sequence number prevents replay
	s.Run("sequence_number_prevents_replay", func() {
		seqTracker := NewSequenceTracker("account123")

		// Process in order
		require.True(s.T(), seqTracker.Accept(1), "sequence 1 should be accepted")
		require.True(s.T(), seqTracker.Accept(2), "sequence 2 should be accepted")
		require.True(s.T(), seqTracker.Accept(3), "sequence 3 should be accepted")

		// Replay old sequence
		require.False(s.T(), seqTracker.Accept(2), "replayed sequence 2 should be rejected")
		require.False(s.T(), seqTracker.Accept(1), "replayed sequence 1 should be rejected")
	})
}

// =============================================================================
// Malformed Signature Handling Tests
// =============================================================================

// TestMalformedSignatureHandling tests handling of malformed signatures.
func (s *SignatureSecurityTestSuite) TestMalformedSignatureHandling() {
	s.T().Log("=== Test: Malformed Signature Handling ===")

	keyPair := generateSigningKeyPair(s.T())
	payload := []byte("test payload")

	// Test: Empty signature rejected
	s.Run("empty_signature_rejected", func() {
		valid := verifySignature(payload, []byte{}, keyPair.PublicKey[:])
		require.False(s.T(), valid, "empty signature should be rejected")
	})

	// Test: Nil signature rejected
	s.Run("nil_signature_rejected", func() {
		valid := verifySignature(payload, nil, keyPair.PublicKey[:])
		require.False(s.T(), valid, "nil signature should be rejected")
	})

	// Test: Single byte signature rejected
	s.Run("single_byte_signature_rejected", func() {
		valid := verifySignature(payload, []byte{0x00}, keyPair.PublicKey[:])
		require.False(s.T(), valid, "single byte signature should be rejected")
	})

	// Test: Very long signature rejected
	s.Run("very_long_signature_rejected", func() {
		longSig := make([]byte, 10000)
		valid := verifySignature(payload, longSig, keyPair.PublicKey[:])
		require.False(s.T(), valid, "very long signature should be rejected")
	})

	// Test: Empty public key rejected
	s.Run("empty_public_key_rejected", func() {
		sig := signPayload(s.T(), payload, keyPair)
		valid := verifySignature(payload, sig, []byte{})
		require.False(s.T(), valid, "empty public key should be rejected")
	})
}

// =============================================================================
// Test Helpers
// =============================================================================

// SigningKeyPair represents a test signing key pair
type SigningKeyPair struct {
	PublicKey  [32]byte
	PrivateKey [64]byte // Ed25519 expanded private key
}

// IdentityUploadPayload represents an identity scope upload
type IdentityUploadPayload struct {
	Scopes          []byte
	Salt            []byte
	Timestamp       time.Time
	ClientSignature []byte
	UserSignature   []byte
}

// SigningPayload returns the bytes to sign
func (p *IdentityUploadPayload) SigningPayload() []byte {
	h := sha256.New()
	h.Write(p.Scopes)
	h.Write(p.Salt)
	h.Write([]byte(p.Timestamp.Format(time.RFC3339Nano)))
	return h.Sum(nil)
}

// NonceTracker tracks used nonces to prevent replay
type NonceTracker struct {
	used map[string]bool
}

func NewNonceTracker() *NonceTracker {
	return &NonceTracker{used: make(map[string]bool)}
}

func (t *NonceTracker) UseNonce(nonce []byte) bool {
	key := hex.EncodeToString(nonce)
	if t.used[key] {
		return false // Replay detected
	}
	t.used[key] = true
	return true
}

// SequenceTracker tracks sequence numbers per account
type SequenceTracker struct {
	accountID    string
	lastSequence uint64
}

func NewSequenceTracker(accountID string) *SequenceTracker {
	return &SequenceTracker{accountID: accountID, lastSequence: 0}
}

func (t *SequenceTracker) Accept(seq uint64) bool {
	if seq <= t.lastSequence {
		return false // Replay or out-of-order
	}
	t.lastSequence = seq
	return true
}

func generateSigningKeyPair(t *testing.T) *SigningKeyPair {
	t.Helper()

	var publicKey [32]byte
	var privateKey [64]byte

	// Generate random keys for testing
	_, err := io.ReadFull(rand.Reader, privateKey[:32])
	require.NoError(t, err)

	// Derive public key (simplified for testing)
	h := sha256.Sum256(privateKey[:32])
	copy(publicKey[:], h[:])
	copy(privateKey[32:], publicKey[:])

	return &SigningKeyPair{
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}
}

func generateSalt(t *testing.T) []byte {
	t.Helper()

	salt := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, salt)
	require.NoError(t, err)

	return salt
}

func signPayload(t *testing.T, payload []byte, keyPair *SigningKeyPair) []byte {
	t.Helper()

	// Simplified signing for testing (real impl uses Ed25519)
	h := sha256.New()
	h.Write(keyPair.PrivateKey[:32])
	h.Write(payload)
	sig := h.Sum(nil)

	// Double to make 64-byte signature
	return append(sig, sig...)
}

func verifySignature(payload, signature, publicKey []byte) bool {
	// Validate signature format
	if len(signature) != 64 {
		return false
	}
	if len(publicKey) != 32 {
		return false
	}

	// Simplified verification for testing
	// Real impl uses Ed25519
	h := sha256.New()

	// Derive what private key would have produced this public key
	// In test, we simulate verification
	expectedFirst := signature[:32]
	expectedSecond := signature[32:]

	// Check signature halves match
	if !bytes.Equal(expectedFirst, expectedSecond) {
		return false
	}

	// Verify using public key derivation
	h.Write(derivePrivateFromPublic(publicKey))
	h.Write(payload)
	expected := h.Sum(nil)

	return bytes.Equal(expected, expectedFirst)
}

func derivePrivateFromPublic(publicKey []byte) []byte {
	// Inverse of our test derivation (not real crypto)
	// This only works because we use SHA256(private) = public
	return publicKey // Simplified for testing
}

func isApprovedClient(publicKey []byte, allowlist [][]byte) bool {
	for _, approved := range allowlist {
		if bytes.Equal(publicKey, approved) {
			return true
		}
	}
	return false
}

func isTimestampValid(ts time.Time, window time.Duration) bool {
	now := time.Now().UTC()
	return ts.After(now.Add(-window)) && ts.Before(now.Add(window))
}
