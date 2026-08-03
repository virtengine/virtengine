package crypto

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/encryption/types"
)

func TestGenerateKeyPair(t *testing.T) {
	kp1, err := GenerateKeyPair()
	require.NoError(t, err)
	require.NotNil(t, kp1)

	// Key should be 32 bytes
	assert.Len(t, kp1.PublicKey, 32)
	assert.Len(t, kp1.PrivateKey, 32)

	// Keys should not be all zeros
	assert.NotEqual(t, [32]byte{}, kp1.PublicKey)
	assert.NotEqual(t, [32]byte{}, kp1.PrivateKey)
	assert.NotEqual(t, [ed25519.PublicKeySize]byte{}, kp1.SigningPublicKey)
	assert.NotEqual(t, [ed25519.PrivateKeySize]byte{}, kp1.SigningPrivateKey)
	assert.NotEqual(t, kp1.PublicKey[:], kp1.SigningPublicKey[:])
	assert.True(t, bytes.Equal(kp1.SigningPublicKey[:], ed25519.PrivateKey(kp1.SigningPrivateKey[:]).Public().(ed25519.PublicKey)))

	// Each generation should produce different keys
	kp2, err := GenerateKeyPair()
	require.NoError(t, err)
	assert.NotEqual(t, kp1.PublicKey, kp2.PublicKey)
	assert.NotEqual(t, kp1.PrivateKey, kp2.PrivateKey)
}

func TestKeyPair_Fingerprint(t *testing.T) {
	kp, err := GenerateKeyPair()
	require.NoError(t, err)

	fp := kp.Fingerprint()

	// Fingerprint should be hex-encoded
	assert.NotEmpty(t, fp)
	assert.Len(t, fp, types.KeyFingerprintSize*2)

	// Same key should produce same fingerprint
	assert.Equal(t, fp, kp.Fingerprint())
}

func TestGenerateNonce(t *testing.T) {
	nonce1, err := GenerateNonce()
	require.NoError(t, err)
	assert.Len(t, nonce1, 24)

	nonce2, err := GenerateNonce()
	require.NoError(t, err)

	// Each nonce should be unique
	assert.NotEqual(t, nonce1, nonce2)
}

func TestCreateAndOpenEnvelope(t *testing.T) {
	// Generate sender and recipient key pairs
	sender, err := GenerateKeyPair()
	require.NoError(t, err)

	recipient, err := GenerateKeyPair()
	require.NoError(t, err)

	// Test data
	plaintext := []byte("Hello, this is a secret message!")

	// Create envelope
	envelope, err := CreateEnvelope(plaintext, recipient.PublicKey[:], sender)
	require.NoError(t, err)
	require.NotNil(t, envelope)

	// Validate envelope structure
	assert.Equal(t, types.EnvelopeVersionV2, envelope.Version)
	assert.Equal(t, types.AlgorithmX25519XSalsa20Poly1305, envelope.AlgorithmID)
	assert.Len(t, envelope.RecipientKeyIDs, 1)
	assert.Len(t, envelope.RecipientPublicKeys, 1)
	assert.Equal(t, recipient.PublicKey[:], envelope.RecipientPublicKeys[0])
	assert.Len(t, envelope.Nonce, types.XSalsa20NonceSize)
	assert.NotEmpty(t, envelope.Ciphertext)
	assert.NotEmpty(t, envelope.SenderSignature)
	assert.Equal(t, sender.PublicKey[:], envelope.SenderPubKey)
	assert.Equal(t, sender.SigningPublicKey[:], envelope.SenderSigningPubKey)

	// Ciphertext should be different from plaintext
	assert.NotEqual(t, plaintext, envelope.Ciphertext)

	// Open envelope with recipient's private key
	decrypted, err := OpenEnvelope(envelope, recipient.PrivateKey[:])
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestCreateEnvelope_InvalidRecipientKey(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)

	plaintext := []byte("test")

	// Wrong key size
	_, err = CreateEnvelope(plaintext, make([]byte, 16), sender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid recipient public key size")
}

func TestCreateEnvelope_RejectsMismatchedSenderKeys(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)
	recipient, err := GenerateKeyPair()
	require.NoError(t, err)

	t.Run("X25519", func(t *testing.T) {
		mismatched := *sender
		mismatched.PublicKey[0] ^= 1
		_, err := CreateEnvelope([]byte("test"), recipient.PublicKey[:], &mismatched)
		require.ErrorContains(t, err, "X25519 public key does not match")
	})

	t.Run("Ed25519", func(t *testing.T) {
		mismatched := *sender
		mismatched.SigningPublicKey[0] ^= 1
		_, err := CreateEnvelope([]byte("test"), recipient.PublicKey[:], &mismatched)
		require.ErrorContains(t, err, "Ed25519 public key does not match")
	})
}

func TestCreateEnvelope_RejectsLowOrderRecipientAndMismatchedID(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)
	recipient, err := GenerateKeyPair()
	require.NoError(t, err)

	_, err = CreateEnvelope([]byte("test"), make([]byte, 32), sender)
	require.ErrorContains(t, err, "low-order recipient public key")

	_, err = CreateEnvelopeWithRecipient([]byte("test"), RecipientInfo{
		PublicKey: recipient.PublicKey[:],
		KeyID:     strings.Repeat("0", types.KeyFingerprintSize*2),
	}, sender)
	require.ErrorContains(t, err, "key ID does not match")

	_, err = CreateMultiRecipientEnvelope([]byte("test"), [][]byte{recipient.PublicKey[:], recipient.PublicKey[:]}, sender)
	require.ErrorContains(t, err, "duplicate recipient public key")
}

func TestValidateEnvelopeSignatureWithExpectedKey(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)
	recipient, err := GenerateKeyPair()
	require.NoError(t, err)
	envelope, err := CreateEnvelope([]byte("test"), recipient.PublicKey[:], sender)
	require.NoError(t, err)

	valid, err := ValidateEnvelopeSignatureWithExpectedKey(envelope, sender.SigningPublicKey[:])
	require.NoError(t, err)
	require.True(t, valid)

	otherSender, err := GenerateKeyPair()
	require.NoError(t, err)
	_, err = ValidateEnvelopeSignatureWithExpectedKey(envelope, otherSender.SigningPublicKey[:])
	require.ErrorContains(t, err, "does not match trusted key")
}

func TestOpenEnvelope_WrongKey(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)

	recipient, err := GenerateKeyPair()
	require.NoError(t, err)

	wrongRecipient, err := GenerateKeyPair()
	require.NoError(t, err)

	plaintext := []byte("secret message")

	// Create envelope for correct recipient
	envelope, err := CreateEnvelope(plaintext, recipient.PublicKey[:], sender)
	require.NoError(t, err)

	// Try to open with wrong key
	_, err = OpenEnvelope(envelope, wrongRecipient.PrivateKey[:])
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decrypt")
}

func TestOpenEnvelope_NilEnvelope(t *testing.T) {
	recipient, err := GenerateKeyPair()
	require.NoError(t, err)

	_, err = OpenEnvelope(nil, recipient.PrivateKey[:])
	require.Error(t, err)
}

func TestOpenEnvelope_InvalidPrivateKeySize(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)

	recipient, err := GenerateKeyPair()
	require.NoError(t, err)

	plaintext := []byte("test")

	envelope, err := CreateEnvelope(plaintext, recipient.PublicKey[:], sender)
	require.NoError(t, err)

	// Wrong private key size
	_, err = OpenEnvelope(envelope, make([]byte, 16))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid private key size")
}

func TestCreateMultiRecipientEnvelope(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)

	// Create multiple recipients
	recipients := make([]*KeyPair, 3)
	recipientPubKeys := make([][]byte, 3)
	for i := range recipients {
		recipients[i], err = GenerateKeyPair()
		require.NoError(t, err)
		recipientPubKeys[i] = recipients[i].PublicKey[:]
	}

	plaintext := []byte("Multi-recipient secret message")

	// Create envelope
	envelope, err := CreateMultiRecipientEnvelope(plaintext, recipientPubKeys, sender)
	require.NoError(t, err)
	require.NotNil(t, envelope)

	// Validate structure
	assert.Equal(t, types.EnvelopeVersionV2, envelope.Version)
	assert.Len(t, envelope.RecipientKeyIDs, 3)
	assert.Empty(t, envelope.EncryptedKeys)
	assert.Len(t, envelope.WrappedKeys, 3)
	assert.Len(t, envelope.RecipientPublicKeys, 3)
	for _, wrappedKey := range envelope.WrappedKeys {
		assert.Equal(t, types.WrappedKeyAlgorithmX25519NaClBox, wrappedKey.Algorithm)
		assert.Len(t, wrappedKey.EphemeralPubKey, 32)
		assert.Len(t, wrappedKey.WrappedKey, 24+32+16)
	}
	assert.NotEqual(t, envelope.WrappedKeys[0].EphemeralPubKey, envelope.WrappedKeys[1].EphemeralPubKey)

	// Check mode metadata
	mode, ok := envelope.GetMetadata("_mode")
	assert.True(t, ok)
	assert.Equal(t, "multi-recipient", mode)

	// Each recipient should be able to decrypt
	for i, recipient := range recipients {
		decrypted, err := OpenEnvelope(envelope, recipient.PrivateKey[:])
		require.NoError(t, err, "recipient %d failed to decrypt", i)
		assert.Equal(t, plaintext, decrypted)
	}
}

func TestCreateMultiRecipientEnvelope_SingleRecipient(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)

	recipient, err := GenerateKeyPair()
	require.NoError(t, err)

	plaintext := []byte("test")

	// Single recipient should use simple envelope
	envelope, err := CreateMultiRecipientEnvelope(plaintext, [][]byte{recipient.PublicKey[:]}, sender)
	require.NoError(t, err)

	// Should not have multi-recipient mode set
	_, ok := envelope.GetMetadata("_mode")
	assert.False(t, ok)

	// Should still be decryptable
	decrypted, err := OpenEnvelope(envelope, recipient.PrivateKey[:])
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestCreateMultiRecipientEnvelope_NoRecipients(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)

	_, err = CreateMultiRecipientEnvelope([]byte("test"), [][]byte{}, sender)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one recipient")
}

func TestOpenEnvelope_VersionedKeyID(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)

	recipient, err := GenerateKeyPair()
	require.NoError(t, err)

	plaintext := []byte("versioned envelope")

	envelope, err := CreateEnvelopeWithRecipient(plaintext, RecipientInfo{
		PublicKey:  recipient.PublicKey[:],
		KeyVersion: 2,
	}, sender)
	require.NoError(t, err)

	require.Len(t, envelope.RecipientKeyIDs, 1)
	assert.Contains(t, envelope.RecipientKeyIDs[0], ":v2")

	decrypted, err := OpenEnvelope(envelope, recipient.PrivateKey[:])
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestValidateEnvelopeSignature(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)

	recipient, err := GenerateKeyPair()
	require.NoError(t, err)

	plaintext := []byte("test message")

	envelope, err := CreateEnvelope(plaintext, recipient.PublicKey[:], sender)
	require.NoError(t, err)

	// Valid signature
	valid, err := ValidateEnvelopeSignature(envelope)
	require.NoError(t, err)
	assert.True(t, valid)

	// Tamper with ciphertext
	envelope.Ciphertext = append(envelope.Ciphertext, byte(0))
	valid, err = ValidateEnvelopeSignature(envelope)
	require.NoError(t, err)
	assert.False(t, valid)
}

func TestEnvelopeV2SignatureBindsSecurityFields(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)
	recipient1, err := GenerateKeyPair()
	require.NoError(t, err)
	recipient2, err := GenerateKeyPair()
	require.NoError(t, err)

	envelope, err := CreateMultiRecipientEnvelopeWithRecipients([]byte("signed payload"), []RecipientInfo{
		{PublicKey: recipient1.PublicKey[:], KeyVersion: 1},
		{PublicKey: recipient2.PublicKey[:], KeyVersion: 2},
	}, sender)
	require.NoError(t, err)

	tests := []struct {
		name   string
		mutate func(*types.EncryptedPayloadEnvelope)
	}{
		{name: "algorithm", mutate: func(e *types.EncryptedPayloadEnvelope) { e.AlgorithmID += "-forged" }},
		{name: "algorithm version", mutate: func(e *types.EncryptedPayloadEnvelope) { e.AlgorithmVersion++ }},
		{name: "payload nonce", mutate: func(e *types.EncryptedPayloadEnvelope) { e.Nonce[0] ^= 1 }},
		{name: "ciphertext", mutate: func(e *types.EncryptedPayloadEnvelope) { e.Ciphertext[0] ^= 1 }},
		{name: "wrapped key", mutate: func(e *types.EncryptedPayloadEnvelope) { e.WrappedKeys[0].WrappedKey[24] ^= 1 }},
		{name: "wrapped key algorithm", mutate: func(e *types.EncryptedPayloadEnvelope) { e.WrappedKeys[0].Algorithm += "-forged" }},
		{name: "ephemeral key", mutate: func(e *types.EncryptedPayloadEnvelope) { e.WrappedKeys[0].EphemeralPubKey[0] ^= 1 }},
		{name: "recipient removal", mutate: func(e *types.EncryptedPayloadEnvelope) {
			e.RecipientKeyIDs = e.RecipientKeyIDs[:1]
			e.RecipientPublicKeys = e.RecipientPublicKeys[:1]
			e.WrappedKeys = e.WrappedKeys[:1]
		}},
		{name: "recipient reorder", mutate: func(e *types.EncryptedPayloadEnvelope) {
			e.RecipientKeyIDs[0], e.RecipientKeyIDs[1] = e.RecipientKeyIDs[1], e.RecipientKeyIDs[0]
			e.RecipientPublicKeys[0], e.RecipientPublicKeys[1] = e.RecipientPublicKeys[1], e.RecipientPublicKeys[0]
			e.WrappedKeys[0], e.WrappedKeys[1] = e.WrappedKeys[1], e.WrappedKeys[0]
		}},
		{name: "recipient substitution", mutate: func(e *types.EncryptedPayloadEnvelope) {
			e.RecipientPublicKeys[0][0] ^= 1
			e.RecipientKeyIDs[0] = types.ComputeKeyFingerprint(e.RecipientPublicKeys[0]) + ":v1"
			e.WrappedKeys[0].RecipientID = e.RecipientKeyIDs[0]
		}},
		{name: "sender encryption key", mutate: func(e *types.EncryptedPayloadEnvelope) { e.SenderPubKey[0] ^= 1 }},
		{name: "forged sender signing key", mutate: func(e *types.EncryptedPayloadEnvelope) { e.SenderSigningPubKey[0] ^= 1 }},
		{name: "metadata", mutate: func(e *types.EncryptedPayloadEnvelope) { e.Metadata["purpose"] = "forged" }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := cloneEnvelope(t, envelope)
			test.mutate(mutated)
			valid, err := ValidateEnvelopeSignature(mutated)
			require.NoError(t, err)
			assert.False(t, valid)
		})
	}
}

func TestOpenEnvelopeAuthenticatesBeforeDecrypting(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)
	recipient, err := GenerateKeyPair()
	require.NoError(t, err)
	envelope, err := CreateEnvelope([]byte("authenticated"), recipient.PublicKey[:], sender)
	require.NoError(t, err)
	envelope.Ciphertext[0] ^= 1

	_, err = OpenEnvelope(envelope, recipient.PrivateKey[:])
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signature")
}

func TestEd25519EnvelopeSignatureIsDeterministic(t *testing.T) {
	seed := bytes.Repeat([]byte{0x42}, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)
	envelope := &types.EncryptedPayloadEnvelope{
		Version:             types.EnvelopeVersionV2,
		AlgorithmID:         types.AlgorithmX25519XSalsa20Poly1305,
		AlgorithmVersion:    types.AlgorithmVersionV1,
		RecipientKeyIDs:     []string{"recipient:v7"},
		RecipientPublicKeys: [][]byte{bytes.Repeat([]byte{1}, 32)},
		Nonce:               bytes.Repeat([]byte{2}, 24),
		Ciphertext:          []byte("fixture ciphertext"),
		SenderPubKey:        bytes.Repeat([]byte{3}, 32),
		SenderSigningPubKey: append([]byte(nil), privateKey.Public().(ed25519.PublicKey)...),
		Metadata:            map[string]string{"context": "T5-08"},
	}

	signature1, err := signEnvelope(envelope, privateKey)
	require.NoError(t, err)
	signature2, err := signEnvelope(envelope, privateKey)
	require.NoError(t, err)
	assert.Equal(t, signature1, signature2)
	envelope.SenderSignature = signature1
	valid, err := ValidateEnvelopeSignature(envelope)
	require.NoError(t, err)
	assert.True(t, valid)
}

func TestEnvelopeV2VersionedRecipientIDsAndLegacyV1(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)
	recipient1, err := GenerateKeyPair()
	require.NoError(t, err)
	recipient2, err := GenerateKeyPair()
	require.NoError(t, err)

	envelope, err := CreateMultiRecipientEnvelopeWithRecipients([]byte("rotation"), []RecipientInfo{
		{PublicKey: recipient1.PublicKey[:], KeyVersion: 4},
		{PublicKey: recipient2.PublicKey[:], KeyVersion: 5},
	}, sender)
	require.NoError(t, err)
	assert.Contains(t, envelope.RecipientKeyIDs[0], ":v4")
	assert.Contains(t, envelope.RecipientKeyIDs[1], ":v5")
	for _, recipient := range []*KeyPair{recipient1, recipient2} {
		plaintext, err := OpenEnvelope(envelope, recipient.PrivateKey[:])
		require.NoError(t, err)
		assert.Equal(t, []byte("rotation"), plaintext)
	}

	legacy, err := CreateEnvelope([]byte("migrate me"), recipient1.PublicKey[:], sender)
	require.NoError(t, err)
	legacy.Version = types.EnvelopeVersionV1
	valid, err := ValidateEnvelopeSignature(legacy)
	require.Error(t, err)
	assert.False(t, valid)
	_, err = OpenEnvelope(legacy, recipient1.PrivateKey[:])
	require.Error(t, err)
	plaintext, err := OpenUnauthenticatedLegacyEnvelopeV1(legacy, recipient1.PrivateKey[:])
	require.NoError(t, err)
	assert.Equal(t, []byte("migrate me"), plaintext)
}

func cloneEnvelope(t *testing.T, envelope *types.EncryptedPayloadEnvelope) *types.EncryptedPayloadEnvelope {
	t.Helper()
	encoded, err := json.Marshal(envelope)
	require.NoError(t, err)
	var cloned types.EncryptedPayloadEnvelope
	require.NoError(t, json.Unmarshal(encoded, &cloned))
	return &cloned
}

func TestValidateEnvelopeSignature_NoSignature(t *testing.T) {
	envelope := &types.EncryptedPayloadEnvelope{
		SenderSignature: nil,
	}

	_, err := ValidateEnvelopeSignature(envelope)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no signature")
}

func TestValidateEnvelopeSignature_NilEnvelope(t *testing.T) {
	_, err := ValidateEnvelopeSignature(nil)
	require.Error(t, err)
}

func TestLargePayload(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)

	recipient, err := GenerateKeyPair()
	require.NoError(t, err)

	// Create a large payload (1MB)
	plaintext := make([]byte, 1024*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	envelope, err := CreateEnvelope(plaintext, recipient.PublicKey[:], sender)
	require.NoError(t, err)

	decrypted, err := OpenEnvelope(envelope, recipient.PrivateKey[:])
	require.NoError(t, err)
	assert.True(t, bytes.Equal(plaintext, decrypted))
}

func TestEmptyPayload(t *testing.T) {
	sender, err := GenerateKeyPair()
	require.NoError(t, err)

	recipient, err := GenerateKeyPair()
	require.NoError(t, err)

	plaintext := []byte{}

	envelope, err := CreateEnvelope(plaintext, recipient.PublicKey[:], sender)
	require.NoError(t, err)

	decrypted, err := OpenEnvelope(envelope, recipient.PrivateKey[:])
	require.NoError(t, err)
	assert.Empty(t, decrypted)
}
