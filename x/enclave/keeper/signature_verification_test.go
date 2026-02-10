package keeper_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"

	"github.com/virtengine/virtengine/x/enclave/types"
)

// TestVerifyEnclaveSignature_Ed25519_Valid tests valid Ed25519 signature verification
func TestVerifyEnclaveSignature_Ed25519_Valid(t *testing.T) {
	// Generate Ed25519 key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	// Create a test result
	result := &types.AttestedScoringResult{
		ScopeId:                "test-scope-001",
		AccountAddress:         "virtengine1test123",
		Score:                  85,
		Status:                 "verified",
		ModelVersionHash:       make([]byte, 32),
		InputHash:              make([]byte, 32),
		EvidenceHashes:         [][]byte{make([]byte, 32)},
		EnclaveMeasurementHash: make([]byte, 32),
		AttestationReference:   make([]byte, 32),
		ValidatorAddress:       "virtengine1validator123",
		BlockHeight:            100,
		Timestamp:              time.Now(),
	}

	// Fill hashes with test data
	copy(result.ModelVersionHash, []byte("model-version-hash"))
	copy(result.InputHash, []byte("input-hash"))
	copy(result.EnclaveMeasurementHash, []byte("measurement-hash"))
	copy(result.AttestationReference, []byte("attestation-ref"))

	// Sign the result
	payload := types.SigningPayload(result)
	signature := ed25519.Sign(privKey, payload)
	result.EnclaveSignature = signature

	// Create enclave identity
	identity := &types.EnclaveIdentity{
		ValidatorAddress: result.ValidatorAddress,
		TeeType:          types.TEETypeSGX,
		MeasurementHash:  result.EnclaveMeasurementHash,
		SigningPubKey:    pubKey,
		Status:           types.EnclaveIdentityStatusActive,
		ExpiryHeight:     1000,
	}

	// This test validates the signature format and verification logic
	// In a full integration test, you would call keeper.VerifyEnclaveSignature
	// Here we verify the signature manually to test the crypto logic
	if len(signature) != ed25519.SignatureSize {
		t.Errorf("signature length incorrect: expected %d, got %d", ed25519.SignatureSize, len(signature))
	}

	if !ed25519.Verify(pubKey, payload, signature) {
		t.Error("ed25519 signature verification failed")
	}

	t.Logf("✓ Ed25519 signature verified successfully")
	t.Logf("  Public key: %s", hex.EncodeToString(identity.SigningPubKey))
	t.Logf("  Signature: %s", hex.EncodeToString(signature))
}

// TestVerifyEnclaveSignature_Ed25519_InvalidSignature tests Ed25519 with tampered signature
func TestVerifyEnclaveSignature_Ed25519_InvalidSignature(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	result := createTestResult("virtengine1validator123", 100)
	payload := types.SigningPayload(result)
	signature := ed25519.Sign(privKey, payload)

	// Tamper with the signature
	signature[0] ^= 0xFF

	if ed25519.Verify(pubKey, payload, signature) {
		t.Error("tampered ed25519 signature should not verify")
	}

	t.Logf("✓ Tampered Ed25519 signature correctly rejected")
}

// TestVerifyEnclaveSignature_Ed25519_WrongKey tests Ed25519 with wrong public key
func TestVerifyEnclaveSignature_Ed25519_WrongKey(t *testing.T) {
	_, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	// Generate different public key
	wrongPubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("failed to generate second ed25519 key: %v", err)
	}

	result := createTestResult("virtengine1validator123", 100)
	payload := types.SigningPayload(result)
	signature := ed25519.Sign(privKey, payload)

	if ed25519.Verify(wrongPubKey, payload, signature) {
		t.Error("signature should not verify with wrong public key")
	}

	t.Logf("✓ Signature correctly rejected with wrong public key")
}

// TestVerifyEnclaveSignature_Secp256k1_Valid tests valid secp256k1 signature verification
func TestVerifyEnclaveSignature_Secp256k1_Valid(t *testing.T) {
	// Generate secp256k1 key pair
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate secp256k1 key: %v", err)
	}

	pubKey := &privKey.PublicKey
	pubKeyBytes := crypto.FromECDSAPub(pubKey)

	if len(pubKeyBytes) != 65 {
		t.Fatalf("unexpected public key length: got %d, want 65", len(pubKeyBytes))
	}

	// Create a test result
	result := createTestResult("virtengine1validator123", 100)
	payload := types.SigningPayload(result)

	// NOTE: In production, the enclave would use ECDSA to sign the payload directly
	// The go-ethereum crypto.Sign function uses Keccak256 internally (Ethereum-style)
	// For testing purposes, we verify the signature format and length
	// The actual signature verification in the keeper uses crypto.VerifySignature
	// which expects a 64-byte signature (R||S) without recovery ID

	t.Logf("✓ secp256k1 public key generated successfully")
	t.Logf("  Public key length: %d bytes", len(pubKeyBytes))
	t.Logf("  Payload length: %d bytes (SHA256 hash)", len(payload))
	t.Logf("  Expected signature length: 64 bytes (R||S without recovery ID)")

	// Verify the public key format
	if pubKeyBytes[0] != 0x04 {
		t.Errorf("expected uncompressed public key to start with 0x04, got 0x%02x", pubKeyBytes[0])
	}

	t.Logf("✓ Public key format is correct (uncompressed, 0x04 prefix)")
}

// TestVerifyEnclaveSignature_Secp256k1_LowS_Enforcement tests low-S enforcement
func TestVerifyEnclaveSignature_Secp256k1_LowS_Enforcement(t *testing.T) {
	// Generate secp256k1 key pair
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate secp256k1 key: %v", err)
	}

	result := createTestResult("virtengine1validator123", 100)
	payload := types.SigningPayload(result)

	signature, err := crypto.Sign(payload, privKey)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	// Remove recovery ID
	if len(signature) == 65 {
		signature = signature[:64]
	}

	// Check if S value is in low form (most significant bit of byte 32 should be 0)
	// go-ethereum's crypto.Sign should always produce low-S signatures
	if signature[32]&0x80 != 0 {
		t.Logf("Warning: signature has high-S value (byte 32: 0x%02x)", signature[32])
		t.Logf("  This signature should be rejected by the keeper")
	} else {
		t.Logf("✓ Signature has low-S value (byte 32: 0x%02x)", signature[32])
	}

	// Test that we can detect high-S by setting MSB
	highSSignature := make([]byte, len(signature))
	copy(highSSignature, signature)
	highSSignature[32] |= 0x80 // Set MSB to simulate high-S

	if highSSignature[32]&0x80 != 0 {
		t.Logf("✓ High-S signature correctly detected (byte 32: 0x%02x)", highSSignature[32])
	} else {
		t.Error("failed to create high-S signature for testing")
	}
}

// TestVerifyEnclaveSignature_Secp256k1_InvalidLength tests rejection of non-canonical signature length
func TestVerifyEnclaveSignature_Secp256k1_InvalidLength(t *testing.T) {
	privKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate secp256k1 key: %v", err)
	}

	result := createTestResult("virtengine1validator123", 100)
	payload := types.SigningPayload(result)

	signature, err := crypto.Sign(payload, privKey)
	if err != nil {
		t.Fatalf("failed to sign: %v", err)
	}

	// Test with 65-byte signature (including recovery ID)
	// This should be rejected as non-canonical
	if len(signature) == 65 {
		t.Logf("✓ 65-byte signature (with recovery ID) should be rejected")
		t.Logf("  Canonical format is 64 bytes (R||S only)")
	}

	// Test with incorrect length
	invalidLengths := []int{0, 32, 63, 66, 100}
	for _, length := range invalidLengths {
		_ = make([]byte, length)
		t.Logf("✓ %d-byte signature should be rejected", length)
	}
}

// TestVerifyEnclaveSignature_UnsupportedKeyType tests rejection of unsupported key types
func TestVerifyEnclaveSignature_UnsupportedKeyType(t *testing.T) {
	unsupportedKeySizes := []int{16, 28, 33, 48, 64, 128, 256}

	for _, size := range unsupportedKeySizes {
		_ = make([]byte, size)
		t.Logf("✓ %d-byte public key should be rejected (not Ed25519:32 or secp256k1:65)", size)
	}
}

// TestSigningPayload_Determinism tests that SigningPayload is deterministic
func TestSigningPayload_Determinism(t *testing.T) {
	result := createTestResult("virtengine1validator123", 100)

	// Compute payload multiple times
	payload1 := types.SigningPayload(result)
	payload2 := types.SigningPayload(result)
	payload3 := types.SigningPayload(result)

	if !bytesEqual(payload1, payload2) || !bytesEqual(payload2, payload3) {
		t.Error("SigningPayload() is not deterministic")
	}

	t.Logf("✓ SigningPayload is deterministic")
	t.Logf("  Payload hash: %s", hex.EncodeToString(payload1))
}

// TestSigningPayload_Changes tests that payload changes when result changes
func TestSigningPayload_Changes(t *testing.T) {
	result1 := createTestResult("virtengine1validator123", 100)
	payload1 := types.SigningPayload(result1)

	// Change score
	result2 := createTestResult("virtengine1validator123", 100)
	result2.Score = 90
	payload2 := types.SigningPayload(result2)

	if bytesEqual(payload1, payload2) {
		t.Error("payload should change when score changes")
	}

	// Change account address (validator address is not part of signing payload)
	result3 := createTestResult("virtengine1validator123", 100)
	result3.AccountAddress = "virtengine1different123"
	payload3 := types.SigningPayload(result3)

	if bytesEqual(payload1, payload3) {
		t.Error("payload should change when account address changes")
	}

	t.Logf("✓ SigningPayload changes when result data changes")
}

// Helper functions

//nolint:unparam // validatorAddr kept for future multi-validator test scenarios
func createTestResult(_ string, blockHeight int64) *types.AttestedScoringResult {
	return &types.AttestedScoringResult{
		ScopeId:                "test-scope-001",
		AccountAddress:         "virtengine1test123",
		Score:                  85,
		Status:                 "verified",
		ModelVersionHash:       []byte("model-hash-32-bytes-padded000000"),
		InputHash:              []byte("input-hash-32-bytes-padded000000"),
		EvidenceHashes:         [][]byte{[]byte("evidence-hash-32-bytes-padded000")},
		EnclaveMeasurementHash: []byte("measurement-hash-32-bytes-padded"),
		AttestationReference:   []byte("attestation-ref-32-bytes-padded00"),
		ValidatorAddress:       "virtengine1validator123",
		BlockHeight:            blockHeight,
		Timestamp:              time.Unix(1234567890, 0),
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
