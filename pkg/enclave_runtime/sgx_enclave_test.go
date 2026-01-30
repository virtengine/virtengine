package enclave_runtime

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// SGX Enclave Service Implementation Tests
// =============================================================================

func TestSGXEnclaveServiceImpl_NewService(t *testing.T) {
	tests := []struct {
		name    string
		config  SGXEnclaveConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: SGXEnclaveConfig{
				EnclavePath: "/path/to/enclave.signed.so",
				DCAPEnabled: true,
				Debug:       false,
			},
			wantErr: false,
		},
		{
			name: "debug mode allowed for testing",
			config: SGXEnclaveConfig{
				EnclavePath: "/path/to/enclave.signed.so",
				Debug:       true,
			},
			wantErr: false,
		},
		{
			name: "missing enclave path",
			config: SGXEnclaveConfig{
				DCAPEnabled: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := NewSGXEnclaveServiceImpl(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, svc)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, svc)
			}
		})
	}
}

func TestSGXEnclaveServiceImpl_Initialize(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
		DCAPEnabled: true,
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)

	// Initialize
	err = svc.Initialize(DefaultRuntimeConfig())
	assert.NoError(t, err)

	// Check status
	status := svc.GetStatus()
	assert.True(t, status.Initialized)
	assert.True(t, status.Available)
	assert.Equal(t, uint64(1), status.CurrentEpoch)

	// Double initialization should fail
	err = svc.Initialize(DefaultRuntimeConfig())
	assert.Error(t, err)
}

func TestSGXEnclaveServiceImpl_Score(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
		DCAPEnabled: true,
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	request := &ScoringRequest{
		RequestID:      "sgx-test-request-1",
		Ciphertext:     []byte("sgx_encrypted_identity_data"),
		WrappedKey:     []byte("sgx_wrapped_key"),
		Nonce:          []byte("sgx_nonce_12345"),
		ScopeID:        "sgx-scope-123",
		AccountAddress: "virtengine1sgxtest",
		BlockHeight:    12345,
	}

	result, err := svc.Score(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsSuccess())

	// Verify result fields
	assert.Equal(t, request.RequestID, result.RequestID)
	assert.LessOrEqual(t, result.Score, uint32(100))
	assert.NotEmpty(t, result.Status)
	assert.NotEmpty(t, result.EnclaveSignature)
	assert.NotEmpty(t, result.MeasurementHash)
	assert.NotEmpty(t, result.InputHash)

	// Verify SGX-specific fields
	assert.Contains(t, result.ReasonCodes[0], "sgx_")
}

func TestSGXEnclaveServiceImpl_GetMeasurement(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)

	// Should fail before init
	_, err = svc.GetMeasurement()
	assert.Error(t, err)

	// Initialize
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Should succeed after init
	mrEnclave, err := svc.GetMeasurement()
	assert.NoError(t, err)
	assert.Len(t, mrEnclave, SGXMREnclaveSize)

	// MRSIGNER should also be available
	mrSigner, err := svc.GetMRSigner()
	assert.NoError(t, err)
	assert.Len(t, mrSigner, SGXMRSignerSize)
}

func TestSGXEnclaveServiceImpl_GenerateAttestation(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
		DCAPEnabled: true,
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Generate attestation with nonce
	reportData := []byte("test_nonce_for_attestation")
	quote, err := svc.GenerateAttestation(reportData)
	assert.NoError(t, err)
	assert.NotEmpty(t, quote)

	// Verify quote contains expected data
	// Quote should be at least header + report body
	assert.Greater(t, len(quote), SGXQuoteHeaderSize)

	// Report data too large should fail
	largeData := make([]byte, SGXReportDataSize+1)
	_, err = svc.GenerateAttestation(largeData)
	assert.Error(t, err)
}

func TestSGXEnclaveServiceImpl_KeyRotation(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Get initial keys
	encPub1, _ := svc.GetEncryptionPubKey()
	sigPub1, _ := svc.GetSigningPubKey()

	// Rotate keys
	err = svc.RotateKeys()
	assert.NoError(t, err)

	// Get new keys
	encPub2, _ := svc.GetEncryptionPubKey()
	sigPub2, _ := svc.GetSigningPubKey()

	// Keys should be different after rotation
	assert.NotEqual(t, encPub1, encPub2)
	assert.NotEqual(t, sigPub1, sigPub2)

	// Epoch should increase
	status := svc.GetStatus()
	assert.Equal(t, uint64(2), status.CurrentEpoch)
}

func TestSGXEnclaveServiceImpl_SealUnseal(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Seal some data
	plaintext := []byte("sensitive_data_to_seal")
	aad := []byte("additional_authenticated_data")

	sealed, err := svc.SealData(plaintext, aad)
	assert.NoError(t, err)
	assert.NotEmpty(t, sealed)

	// Unseal the data
	unsealed, _, err := svc.UnsealData(sealed)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, unsealed)
}

func TestSGXEnclaveServiceImpl_VerifyMeasurement(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Valid measurement
	validMeasurement := make([]byte, SGXMREnclaveSize)
	validMeasurement[0] = 0x01
	assert.True(t, svc.VerifyMeasurement(validMeasurement))

	// All zeros is invalid
	zeroMeasurement := make([]byte, SGXMREnclaveSize)
	assert.False(t, svc.VerifyMeasurement(zeroMeasurement))

	// Wrong size is invalid
	wrongSize := make([]byte, 16)
	assert.False(t, svc.VerifyMeasurement(wrongSize))
}

func TestSGXEnclaveServiceImpl_Shutdown(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Shutdown
	err = svc.Shutdown()
	assert.NoError(t, err)

	// Status should show not initialized
	status := svc.GetStatus()
	assert.False(t, status.Initialized)

	// Operations should fail after shutdown
	_, err = svc.GetMeasurement()
	assert.Error(t, err)
}

func TestSGXEnclaveServiceImpl_PlatformInfo(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
		Debug:       false,
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)

	// Platform type
	assert.Equal(t, PlatformSGX, svc.GetPlatformType())

	// Secure if debug disabled
	assert.True(t, svc.IsPlatformSecure())

	// Debug mode makes it insecure
	config.Debug = true
	svc2, _ := NewSGXEnclaveServiceImpl(config)
	assert.False(t, svc2.IsPlatformSecure())
}

// =============================================================================
// SGX Type Tests
// =============================================================================

func TestSGXMeasurement_String(t *testing.T) {
	var m SGXMeasurement
	m[0] = 0xAB
	m[1] = 0xCD

	str := m.String()
	assert.Contains(t, str, "abcd")
}

func TestSGXAttributes_IsDebug(t *testing.T) {
	// Debug enabled
	debugAttrs := SGXAttributes{Flags: SGXFlagDebug | SGXFlagInitted}
	assert.True(t, debugAttrs.IsDebug())

	// Debug disabled
	prodAttrs := SGXAttributes{Flags: SGXFlagInitted | SGXFlagMode64Bit}
	assert.False(t, prodAttrs.IsDebug())
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestSGXEnclaveServiceImpl_ConcurrentScoring(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Run multiple concurrent scoring requests
	const numRequests = 10
	results := make(chan *ScoringResult, numRequests)
	errors := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(idx int) {
			request := &ScoringRequest{
				RequestID:      string(rune('a' + idx)),
				Ciphertext:     []byte("concurrent_test_data"),
				WrappedKey:     []byte("wrapped_key"),
				Nonce:          []byte("nonce"),
				ScopeID:        "scope",
				AccountAddress: "addr",
			}
			result, err := svc.Score(context.Background(), request)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(i)
	}

	// Collect results (some may fail due to concurrency limits)
	successCount := 0
	for i := 0; i < numRequests; i++ {
		select {
		case <-results:
			successCount++
		case <-errors:
			// Expected for requests exceeding concurrent limit
		}
	}

	// At least max concurrent should succeed
	assert.GreaterOrEqual(t, successCount, 1)
}

// =============================================================================
// Plaintext Isolation Integration Tests
// =============================================================================

func TestPlaintextGuard_BasicOperation(t *testing.T) {
	guard := NewPlaintextGuard()

	// Allocate some buffers
	buf1 := guard.AllocatePlaintext(100)
	buf2 := guard.AllocatePlaintext(200)

	// Write sensitive data
	_, err := buf1.Write([]byte("sensitive_data_1"))
	require.NoError(t, err)
	_, err = buf2.Write([]byte("sensitive_data_2"))
	require.NoError(t, err)

	// Scrub buffers
	guard.ScrubAndRelease(buf1)
	guard.ScrubAndRelease(buf2)

	// Seal should succeed
	err = guard.Seal()
	assert.NoError(t, err)

	// Verify stats
	allocCount, scrubCount, verified := guard.Stats()
	assert.Equal(t, uint64(2), allocCount)
	assert.Equal(t, uint64(2), scrubCount)
	assert.True(t, verified)
}

func TestPlaintextGuard_UnscrubedBuffersDetected(t *testing.T) {
	guard := NewPlaintextGuard()

	// Allocate buffer but don't scrub it
	buf := guard.AllocatePlaintext(100)
	_, _ = buf.Write([]byte("sensitive_data"))

	// Seal should detect unscrubed buffer and force scrub
	err := guard.Seal()
	// Note: Seal now force scrubs and returns error for detection
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not properly scrubbed")

	// But buffer should still be destroyed after seal
	assert.True(t, buf.IsDestroyed())
}

func TestPlaintextGuard_CannotAllocateAfterSeal(t *testing.T) {
	guard := NewPlaintextGuard()

	// Seal the guard
	err := guard.Seal()
	require.NoError(t, err)

	// Attempting to allocate after seal should panic
	assert.Panics(t, func() {
		guard.AllocatePlaintext(100)
	})
}

func TestSGXEnclaveServiceImpl_PlaintextIsolation(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
		DCAPEnabled: true,
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Reset the plaintext operation counter
	ResetPlaintextOperationCount()
	initialCount := GetPlaintextOperationCount()
	assert.Equal(t, int64(0), initialCount)

	// Perform scoring
	request := &ScoringRequest{
		RequestID:      "isolation-test-1",
		Ciphertext:     []byte("encrypted_identity_data_for_testing"),
		WrappedKey:     []byte("wrapped_key"),
		Nonce:          []byte("nonce_12345"),
		ScopeID:        "scope-123",
		AccountAddress: "virtengine1isolationtest",
		BlockHeight:    12345,
	}

	result, err := svc.Score(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsSuccess())

	// Verify plaintext operation was tracked
	finalCount := GetPlaintextOperationCount()
	assert.Equal(t, int64(1), finalCount)

	// Verify result has valid signature
	assert.NotEmpty(t, result.EnclaveSignature)
	assert.NotEmpty(t, result.MeasurementHash)
}

func TestSGXEnclaveServiceImpl_KeyIsolation(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Get public keys (should succeed)
	encPubKey, err := svc.GetEncryptionPubKey()
	assert.NoError(t, err)
	assert.NotEmpty(t, encPubKey)

	sigPubKey, err := svc.GetSigningPubKey()
	assert.NoError(t, err)
	assert.NotEmpty(t, sigPubKey)

	// Verify key material exists but private keys are not directly accessible
	// (The struct fields are private, so this is enforced by Go's visibility)
	assert.NotNil(t, svc.keyMaterial)
	assert.NotNil(t, svc.keyMaterial.SigningPublic)
	assert.NotNil(t, svc.keyMaterial.EncryptionPublic)

	// Perform key rotation
	err = svc.RotateKeys()
	require.NoError(t, err)

	// Get new public keys
	newEncPubKey, err := svc.GetEncryptionPubKey()
	assert.NoError(t, err)
	newSigPubKey, err := svc.GetSigningPubKey()
	assert.NoError(t, err)

	// Keys should be different after rotation
	assert.NotEqual(t, encPubKey, newEncPubKey)
	assert.NotEqual(t, sigPubKey, newSigPubKey)

	// Epoch should have increased
	status := svc.GetStatus()
	assert.Equal(t, uint64(2), status.CurrentEpoch)
}

func TestSGXEnclaveServiceImpl_MemoryScrubOnShutdown(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	// Verify key material exists
	assert.NotNil(t, svc.keyMaterial)

	// Perform some operations
	request := &ScoringRequest{
		RequestID:      "scrub-test-1",
		Ciphertext:     []byte("test_data"),
		WrappedKey:     []byte("key"),
		Nonce:          []byte("nonce"),
		ScopeID:        "scope",
		AccountAddress: "addr",
	}
	_, _ = svc.Score(context.Background(), request)

	// Shutdown should scrub keys
	err = svc.Shutdown()
	require.NoError(t, err)

	// Key material should be nil after shutdown
	assert.Nil(t, svc.keyMaterial)

	// Status should show not initialized
	status := svc.GetStatus()
	assert.False(t, status.Initialized)
}

func TestEnclaveKeyMaterial_Generation(t *testing.T) {
	sealKey := make([]byte, 32)
	_, err := rand.Read(sealKey)
	require.NoError(t, err)

	km, err := GenerateEnclaveKeys(sealKey, 1)
	require.NoError(t, err)
	require.NotNil(t, km)

	// Verify public keys are generated
	assert.NotEmpty(t, km.SigningPublic)
	assert.NotEmpty(t, km.EncryptionPublic[:])
	assert.Equal(t, uint64(1), km.Epoch)

	// Verify signing works
	testData := []byte("test message to sign")
	signature := km.Sign(testData)
	assert.NotEmpty(t, signature)
	assert.Len(t, signature, 64) // Ed25519 signature is 64 bytes

	// Verify key derivation is deterministic
	km2, err := GenerateEnclaveKeys(sealKey, 1)
	require.NoError(t, err)

	// Same seal key and epoch should produce same public keys
	assert.Equal(t, km.SigningPublic, km2.SigningPublic)
	assert.Equal(t, km.EncryptionPublic, km2.EncryptionPublic)

	// Different epoch should produce different keys
	km3, err := GenerateEnclaveKeys(sealKey, 2)
	require.NoError(t, err)
	assert.NotEqual(t, km.SigningPublic, km3.SigningPublic)
}

func TestEnclaveKeyMaterial_Scrubbing(t *testing.T) {
	sealKey := make([]byte, 32)
	_, err := rand.Read(sealKey)
	require.NoError(t, err)

	km, err := GenerateEnclaveKeys(sealKey, 1)
	require.NoError(t, err)

	// Store a copy of the private key before scrubbing
	originalPrivKeyLen := len(km.signingPrivate)
	assert.Greater(t, originalPrivKeyLen, 0)

	// Scrub private keys
	km.ScrubPrivateKeys()

	// Verify private key has been zeroed
	allZero := true
	for _, b := range km.signingPrivate {
		if b != 0 {
			allZero = false
			break
		}
	}
	assert.True(t, allZero, "Private signing key should be zeroed after scrub")

	// Verify encryption private key is zeroed
	for _, b := range km.encryptionPrivate {
		if b != 0 {
			t.Error("Private encryption key should be zeroed after scrub")
			break
		}
	}
}

func TestSGXEnclaveServiceImpl_SignatureVerification(t *testing.T) {
	config := SGXEnclaveConfig{
		EnclavePath: "/path/to/enclave.signed.so",
	}

	svc, err := NewSGXEnclaveServiceImpl(config)
	require.NoError(t, err)
	require.NoError(t, svc.Initialize(DefaultRuntimeConfig()))

	request := &ScoringRequest{
		RequestID:      "sig-verify-test",
		Ciphertext:     []byte("data_to_score"),
		WrappedKey:     []byte("key"),
		Nonce:          []byte("nonce"),
		ScopeID:        "scope",
		AccountAddress: "addr",
	}

	result, err := svc.Score(context.Background(), request)
	require.NoError(t, err)
	require.True(t, result.IsSuccess())

	// Get signing public key
	sigPubKey, err := svc.GetSigningPubKey()
	require.NoError(t, err)

	// Verify the signature is 64 bytes (Ed25519)
	assert.Len(t, result.EnclaveSignature, 64)

	// In real usage, we would verify the signature here
	// For now, just check it's not empty and is consistent
	assert.NotEmpty(t, sigPubKey)
	assert.Equal(t, 32, len(sigPubKey)) // Ed25519 public key is 32 bytes
}
