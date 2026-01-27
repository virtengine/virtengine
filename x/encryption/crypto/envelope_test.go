package crypto

import (
	"bytes"
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
	assert.Equal(t, types.EnvelopeVersion, envelope.Version)
	assert.Equal(t, types.AlgorithmX25519XSalsa20Poly1305, envelope.AlgorithmID)
	assert.Len(t, envelope.RecipientKeyIDs, 1)
	assert.Len(t, envelope.Nonce, types.XSalsa20NonceSize)
	assert.NotEmpty(t, envelope.Ciphertext)
	assert.NotEmpty(t, envelope.SenderSignature)
	assert.Equal(t, sender.PublicKey[:], envelope.SenderPubKey)

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
	assert.Equal(t, types.EnvelopeVersion, envelope.Version)
	assert.Len(t, envelope.RecipientKeyIDs, 3)
	assert.Len(t, envelope.EncryptedKeys, 3)

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
