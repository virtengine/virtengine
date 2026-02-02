package capture_protocol

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants
const (
	testClientID    = "test-client-001"
	testClientName  = "Test Mobile App"
	testDeviceID    = "device-fingerprint-abc123"
	testSessionID   = "session-uuid-xyz789"
	testUserAddress = "virtengine1abc123def456"
)

// TestSaltValidation tests salt validation logic
func TestSaltValidation(t *testing.T) {
	t.Run("valid salt is accepted", func(t *testing.T) {
		sv := NewSaltValidator()
		salt := generateRandomSalt(t, 32)
		binding := CreateSaltBinding(salt, testDeviceID, testSessionID, time.Now().Unix())

		err := sv.ValidateSalt(binding)
		assert.NoError(t, err)
	})

	t.Run("empty salt is rejected", func(t *testing.T) {
		sv := NewSaltValidator()
		err := sv.ValidateSaltOnly(nil)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeSaltEmpty, GetErrorCode(err))
	})

	t.Run("short salt is rejected", func(t *testing.T) {
		sv := NewSaltValidator()
		salt := generateRandomSalt(t, 16) // Too short
		err := sv.ValidateSaltOnly(salt)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeSaltTooShort, GetErrorCode(err))
	})

	t.Run("weak salt (all zeros) is rejected", func(t *testing.T) {
		sv := NewSaltValidator()
		salt := make([]byte, 32) // All zeros
		err := sv.ValidateSaltOnly(salt)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeSaltWeak, GetErrorCode(err))
	})

	t.Run("weak salt (all same) is rejected", func(t *testing.T) {
		sv := NewSaltValidator()
		salt := make([]byte, 32)
		for i := range salt {
			salt[i] = 0xAB
		}
		err := sv.ValidateSaltOnly(salt)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeSaltWeak, GetErrorCode(err))
	})

	t.Run("expired salt is rejected", func(t *testing.T) {
		sv := NewSaltValidator(
			WithMaxSaltAge(5 * time.Minute),
		)
		salt := generateRandomSalt(t, 32)
		// Create binding with old timestamp
		binding := CreateSaltBinding(salt, testDeviceID, testSessionID, time.Now().Add(-10*time.Minute).Unix())

		err := sv.ValidateSalt(binding)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeSaltExpired, GetErrorCode(err))
	})

	t.Run("future salt is rejected", func(t *testing.T) {
		sv := NewSaltValidator(
			WithMaxClockSkew(30 * time.Second),
		)
		salt := generateRandomSalt(t, 32)
		// Create binding with future timestamp
		binding := CreateSaltBinding(salt, testDeviceID, testSessionID, time.Now().Add(5*time.Minute).Unix())

		err := sv.ValidateSalt(binding)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeSaltFromFuture, GetErrorCode(err))
	})

	t.Run("invalid binding hash is rejected", func(t *testing.T) {
		sv := NewSaltValidator()
		salt := generateRandomSalt(t, 32)
		binding := CreateSaltBinding(salt, testDeviceID, testSessionID, time.Now().Unix())
		// Corrupt the binding hash
		binding.BindingHash[0] ^= 0xFF

		err := sv.ValidateSalt(binding)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeBindingHashMismatch, GetErrorCode(err))
	})

	t.Run("missing device ID is rejected", func(t *testing.T) {
		sv := NewSaltValidator()
		salt := generateRandomSalt(t, 32)
		binding := CreateSaltBinding(salt, "", testSessionID, time.Now().Unix())

		err := sv.ValidateSalt(binding)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeBindingDeviceIDMissing, GetErrorCode(err))
	})

	t.Run("missing session ID is rejected", func(t *testing.T) {
		sv := NewSaltValidator()
		salt := generateRandomSalt(t, 32)
		binding := CreateSaltBinding(salt, testDeviceID, "", time.Now().Unix())

		err := sv.ValidateSalt(binding)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeBindingSessionIDMissing, GetErrorCode(err))
	})
}

// TestReplayPrevention tests anti-replay mechanisms
func TestReplayPrevention(t *testing.T) {
	t.Run("salt is marked as used after validation", func(t *testing.T) {
		sv := NewSaltValidator()
		salt := generateRandomSalt(t, 32)
		binding := CreateSaltBinding(salt, testDeviceID, testSessionID, time.Now().Unix())

		// First validation should succeed
		err := sv.ValidateSalt(binding)
		require.NoError(t, err)

		// Record the salt
		err = sv.RecordUsedSalt(salt)
		require.NoError(t, err)

		// Second validation should fail (replay)
		err = sv.ValidateSalt(binding)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeSaltReplayed, GetErrorCode(err))
	})

	t.Run("different salts are accepted", func(t *testing.T) {
		sv := NewSaltValidator()

		// First salt
		salt1 := generateRandomSalt(t, 32)
		binding1 := CreateSaltBinding(salt1, testDeviceID, testSessionID, time.Now().Unix())
		err := sv.ValidateSalt(binding1)
		require.NoError(t, err)
		_ = sv.RecordUsedSalt(salt1)

		// Second salt (different)
		salt2 := generateRandomSalt(t, 32)
		binding2 := CreateSaltBinding(salt2, testDeviceID, testSessionID, time.Now().Unix())
		err = sv.ValidateSalt(binding2)
		assert.NoError(t, err)
	})

	t.Run("expired salts are cleaned up", func(t *testing.T) {
		// Use a pointer to time so that both validator and cache see updates
		currentTime := time.Now()
		timeSource := func() time.Time { return currentTime }

		sv := NewSaltValidator(
			WithReplayWindow(1*time.Second),
			WithTimeSource(timeSource),
		)

		salt := generateRandomSalt(t, 32)
		_ = sv.RecordUsedSalt(salt)
		assert.True(t, sv.IsSaltUsed(salt))

		// Advance time past replay window - update the variable that the closure references
		currentTime = currentTime.Add(2 * time.Second)

		// Salt should now be considered not used (expired)
		assert.False(t, sv.IsSaltUsed(salt))
	})
}

// TestSignatureValidation tests signature verification
func TestSignatureValidation(t *testing.T) {
	// Generate test keys
	clientPub, clientPriv := generateEd25519KeyPair(t)
	userPub, userPriv := generateEd25519KeyPair(t)

	// Create mock registry
	registry := NewMockApprovedClientRegistry()
	registry.AddClient(&ApprovedClient{
		ClientID:     testClientID,
		Name:         testClientName,
		PublicKey:    clientPub,
		Algorithm:    AlgorithmEd25519,
		Active:       true,
		RegisteredAt: time.Now().Add(-24 * time.Hour),
	})

	t.Run("valid signatures are accepted", func(t *testing.T) {
		sv := NewSignatureValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)

		err := sv.ValidateClientSignature(payload)
		assert.NoError(t, err)

		err = sv.ValidateUserSignature(payload, testUserAddress)
		assert.NoError(t, err)
	})

	t.Run("missing client signature is rejected in strict mode", func(t *testing.T) {
		sv := NewSignatureValidator(registry, WithStrictMode(true))
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)
		payload.ClientSignature.Signature = nil

		err := sv.ValidateClientSignature(payload)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeClientSignatureMissing, GetErrorCode(err))
	})

	t.Run("unapproved client is rejected", func(t *testing.T) {
		sv := NewSignatureValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)
		payload.ClientSignature.KeyID = "unknown-client"

		err := sv.ValidateClientSignature(payload)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeClientNotApproved, GetErrorCode(err))
	})

	t.Run("wrong client key is rejected", func(t *testing.T) {
		sv := NewSignatureValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)
		// Use different public key
		wrongPub, _ := generateEd25519KeyPair(t)
		payload.ClientSignature.PublicKey = wrongPub

		err := sv.ValidateClientSignature(payload)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeClientKeyMismatch, GetErrorCode(err))
	})

	t.Run("invalid client signature is rejected", func(t *testing.T) {
		sv := NewSignatureValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)
		// Corrupt signature
		payload.ClientSignature.Signature[0] ^= 0xFF

		err := sv.ValidateClientSignature(payload)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeClientSignatureInvalid, GetErrorCode(err))
	})

	t.Run("user address mismatch is rejected", func(t *testing.T) {
		sv := NewSignatureValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)

		err := sv.ValidateUserSignature(payload, "wrong-address")
		assert.Error(t, err)
		assert.Equal(t, ErrCodeUserAddressMismatch, GetErrorCode(err))
	})

	t.Run("invalid user signature is rejected", func(t *testing.T) {
		sv := NewSignatureValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)
		// Corrupt signature
		payload.UserSignature.Signature[0] ^= 0xFF

		err := sv.ValidateUserSignature(payload, testUserAddress)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeUserSignatureInvalid, GetErrorCode(err))
	})

	t.Run("signature chain verification", func(t *testing.T) {
		sv := NewSignatureValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)

		err := sv.VerifySignatureChain(payload)
		assert.NoError(t, err)
	})

	t.Run("broken signature chain is rejected", func(t *testing.T) {
		sv := NewSignatureValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)
		// Modify the signed data to break chain
		payload.UserSignature.SignedData = []byte("wrong data")

		err := sv.VerifySignatureChain(payload)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeSignatureChainBroken, GetErrorCode(err))
	})
}

// TestKeyRotation tests key rotation support
func TestKeyRotation(t *testing.T) {
	oldPub, oldPriv := generateEd25519KeyPair(t)
	newPub, _ := generateEd25519KeyPair(t)
	userPub, userPriv := generateEd25519KeyPair(t)

	t.Run("deprecated key is valid during overlap", func(t *testing.T) {
		registry := NewMockApprovedClientRegistry()
		futureExpiry := time.Now().Add(24 * time.Hour)
		registry.AddClient(&ApprovedClient{
			ClientID:            testClientID,
			Name:                testClientName,
			PublicKey:           newPub,
			Algorithm:           AlgorithmEd25519,
			Active:              true,
			RegisteredAt:        time.Now().Add(-48 * time.Hour),
			DeprecatedKey:       oldPub,
			DeprecatedKeyExpiry: &futureExpiry,
		})

		sv := NewSignatureValidator(registry)
		// Create payload signed with OLD key
		payload := createValidPayload(t, oldPub, oldPriv, userPub, userPriv)

		err := sv.ValidateClientSignature(payload)
		assert.NoError(t, err)
	})

	t.Run("deprecated key is rejected after expiry", func(t *testing.T) {
		registry := NewMockApprovedClientRegistry()
		pastExpiry := time.Now().Add(-1 * time.Hour)
		registry.AddClient(&ApprovedClient{
			ClientID:            testClientID,
			Name:                testClientName,
			PublicKey:           newPub,
			Algorithm:           AlgorithmEd25519,
			Active:              true,
			RegisteredAt:        time.Now().Add(-48 * time.Hour),
			DeprecatedKey:       oldPub,
			DeprecatedKeyExpiry: &pastExpiry,
		})

		sv := NewSignatureValidator(registry)
		// Create payload signed with OLD key
		payload := createValidPayload(t, oldPub, oldPriv, userPub, userPriv)

		err := sv.ValidateClientSignature(payload)
		assert.Error(t, err)
		assert.Equal(t, ErrCodeClientKeyMismatch, GetErrorCode(err))
	})
}

// TestFullProtocolValidation tests the complete protocol validator
func TestFullProtocolValidation(t *testing.T) {
	clientPub, clientPriv := generateEd25519KeyPair(t)
	userPub, userPriv := generateEd25519KeyPair(t)

	registry := NewMockApprovedClientRegistry()
	registry.AddClient(&ApprovedClient{
		ClientID:     testClientID,
		Name:         testClientName,
		PublicKey:    clientPub,
		Algorithm:    AlgorithmEd25519,
		Active:       true,
		RegisteredAt: time.Now().Add(-24 * time.Hour),
	})

	t.Run("valid payload is accepted", func(t *testing.T) {
		pv := NewProtocolValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)

		result := pv.ValidatePayload(payload, testUserAddress)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.Equal(t, testClientID, result.ClientID)
		assert.Equal(t, testUserAddress, result.UserAddress)
	})

	t.Run("replay attack is prevented", func(t *testing.T) {
		pv := NewProtocolValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)

		// First submission should succeed
		result := pv.ValidatePayload(payload, testUserAddress)
		require.True(t, result.Valid)

		// Second submission (replay) should fail
		result = pv.ValidatePayload(payload, testUserAddress)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("expired payload is rejected", func(t *testing.T) {
		pv := NewProtocolValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)
		// Make payload old
		payload.Timestamp = time.Now().Add(-1 * time.Hour)
		payload.SaltBinding.Timestamp = payload.Timestamp.Unix()
		// Recompute binding hash
		payload.SaltBinding.BindingHash = payload.SaltBinding.ComputeBindingHash()

		result := pv.ValidatePayload(payload, testUserAddress)
		assert.False(t, result.Valid)
	})

	t.Run("missing protocol version is rejected", func(t *testing.T) {
		pv := NewProtocolValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)
		payload.Version = 0

		result := pv.ValidatePayload(payload, testUserAddress)
		assert.False(t, result.Valid)
	})

	t.Run("unsupported protocol version is rejected", func(t *testing.T) {
		pv := NewProtocolValidator(registry)
		payload := createValidPayload(t, clientPub, clientPriv, userPub, userPriv)
		payload.Version = 99

		result := pv.ValidatePayload(payload, testUserAddress)
		assert.False(t, result.Valid)
	})
}

// TestBindingHashComputation tests binding hash computation
func TestBindingHashComputation(t *testing.T) {
	salt := generateRandomSalt(t, 32)
	deviceID := "device-123"
	sessionID := "session-456"
	timestamp := time.Now().Unix()

	t.Run("binding hash is deterministic", func(t *testing.T) {
		hash1 := ComputeBindingHash(salt, deviceID, sessionID, timestamp)
		hash2 := ComputeBindingHash(salt, deviceID, sessionID, timestamp)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("binding hash changes with salt", func(t *testing.T) {
		hash1 := ComputeBindingHash(salt, deviceID, sessionID, timestamp)
		salt2 := generateRandomSalt(t, 32)
		hash2 := ComputeBindingHash(salt2, deviceID, sessionID, timestamp)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("binding hash changes with device ID", func(t *testing.T) {
		hash1 := ComputeBindingHash(salt, deviceID, sessionID, timestamp)
		hash2 := ComputeBindingHash(salt, "different-device", sessionID, timestamp)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("binding hash changes with session ID", func(t *testing.T) {
		hash1 := ComputeBindingHash(salt, deviceID, sessionID, timestamp)
		hash2 := ComputeBindingHash(salt, deviceID, "different-session", timestamp)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("binding hash changes with timestamp", func(t *testing.T) {
		hash1 := ComputeBindingHash(salt, deviceID, sessionID, timestamp)
		hash2 := ComputeBindingHash(salt, deviceID, sessionID, timestamp+1)
		assert.NotEqual(t, hash1, hash2)
	})
}

// TestErrorTypes tests error categorization
func TestErrorTypes(t *testing.T) {
	t.Run("salt errors are identified", func(t *testing.T) {
		assert.True(t, IsSaltError(ErrSaltEmpty))
		assert.True(t, IsSaltError(ErrSaltTooShort))
		assert.True(t, IsSaltError(ErrSaltExpired))
		assert.True(t, IsSaltError(ErrSaltReplayed))
		assert.False(t, IsSaltError(ErrClientNotApproved))
	})

	t.Run("signature errors are identified", func(t *testing.T) {
		assert.True(t, IsSignatureError(ErrClientSignatureInvalid))
		assert.True(t, IsSignatureError(ErrUserSignatureInvalid))
		assert.True(t, IsSignatureError(ErrSignatureChainBroken))
		assert.False(t, IsSignatureError(ErrSaltExpired))
	})

	t.Run("replay errors are identified", func(t *testing.T) {
		assert.True(t, IsReplayError(ErrSaltReplayed))
		assert.False(t, IsReplayError(ErrSaltExpired))
	})

	t.Run("client errors are identified", func(t *testing.T) {
		assert.True(t, IsClientError(ErrClientNotApproved))
		assert.True(t, IsClientError(ErrClientNotActive))
		assert.True(t, IsClientError(ErrClientKeyMismatch))
		assert.False(t, IsClientError(ErrUserSignatureInvalid))
	})
}

// Helper functions

func generateRandomSalt(t *testing.T, length int) []byte {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	require.NoError(t, err)
	return salt
}

func generateEd25519KeyPair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	return pub, priv
}

func createValidPayload(
	t *testing.T,
	clientPub ed25519.PublicKey,
	clientPriv ed25519.PrivateKey,
	userPub ed25519.PublicKey,
	userPriv ed25519.PrivateKey,
) CapturePayload {
	// Generate salt
	salt := generateRandomSalt(t, 32)
	now := time.Now()

	// Create payload hash
	payloadHash := sha256.Sum256([]byte("test encrypted content"))

	// Compute signing data
	clientSignedData := ComputeClientSigningData(salt, payloadHash[:])
	clientSig := ed25519.Sign(clientPriv, clientSignedData)

	userSignedData := ComputeUserSigningData(salt, payloadHash[:], clientSig)
	userSig := ed25519.Sign(userPriv, userSignedData)

	// Create metadata
	metadata := CaptureMetadata{
		DeviceFingerprint: testDeviceID,
		ClientID:          testClientID,
		ClientVersion:     "1.0.0",
		SessionID:         testSessionID,
		DocumentType:      "id_card",
		QualityScore:      85,
		CaptureTimestamp:  now.Unix(),
	}

	return CapturePayload{
		Version:     ProtocolVersion,
		PayloadHash: payloadHash[:],
		Salt:        salt,
		SaltBinding: CreateSaltBinding(salt, testDeviceID, testSessionID, now.Unix()),
		ClientSignature: SignatureProof{
			PublicKey:  clientPub,
			Signature:  clientSig,
			Algorithm:  AlgorithmEd25519,
			KeyID:      testClientID,
			SignedData: clientSignedData,
		},
		UserSignature: SignatureProof{
			PublicKey:  userPub,
			Signature:  userSig,
			Algorithm:  AlgorithmEd25519,
			KeyID:      testUserAddress,
			SignedData: userSignedData,
		},
		CaptureMetadata: metadata,
		Timestamp:       now,
	}
}
