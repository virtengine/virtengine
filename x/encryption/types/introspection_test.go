package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/encryption/types"
)

func TestExtractEnvelopeMetadata(t *testing.T) {
	t.Run("nil envelope returns error", func(t *testing.T) {
		_, err := types.ExtractEnvelopeMetadata(nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "envelope cannot be nil")
	})

	t.Run("valid envelope extracts metadata", func(t *testing.T) {
		envelope := types.NewEncryptedPayloadEnvelope()
		envelope.RecipientKeyIDs = []string{"recipient1", "recipient2"}
		envelope.Ciphertext = []byte("encrypted_data_here")
		envelope.Nonce = []byte("nonce_12345")
		envelope.SenderPubKey = []byte("sender_public_key_32_bytes_long!")
		envelope.SenderSignature = []byte("signature")
		envelope.Metadata = map[string]string{
			"purpose": "test",
		}

		metadata, err := types.ExtractEnvelopeMetadata(envelope)
		require.NoError(t, err)
		require.NotNil(t, metadata)

		require.Equal(t, types.EnvelopeVersion, metadata.Version)
		require.Equal(t, types.DefaultAlgorithm(), metadata.AlgorithmID)
		require.Len(t, metadata.RecipientKeyIDs, 2)
		require.Contains(t, metadata.RecipientKeyIDs, "recipient1")
		require.Contains(t, metadata.RecipientKeyIDs, "recipient2")
		require.Equal(t, len(envelope.Ciphertext), metadata.CiphertextSize)
		require.Equal(t, len(envelope.Nonce), metadata.NonceSize)
		require.True(t, metadata.HasSignature)
		require.NotEmpty(t, metadata.SenderPubKeyFingerprint)
		require.NotEmpty(t, metadata.EnvelopeHash)
		require.Equal(t, "test", metadata.Metadata["purpose"])
	})

	t.Run("envelope without signature", func(t *testing.T) {
		envelope := types.NewEncryptedPayloadEnvelope()
		envelope.RecipientKeyIDs = []string{"recipient1"}
		envelope.Ciphertext = []byte("data")
		envelope.Nonce = []byte("nonce")
		envelope.SenderPubKey = []byte("sender_key_32_bytes_long_enough!")
		// No signature

		metadata, err := types.ExtractEnvelopeMetadata(envelope)
		require.NoError(t, err)
		require.False(t, metadata.HasSignature)
	})
}

func TestEnvelopeMetadataString(t *testing.T) {
	envelope := types.NewEncryptedPayloadEnvelope()
	envelope.RecipientKeyIDs = []string{"recipient1", "recipient2"}
	envelope.Ciphertext = make([]byte, 1024)
	envelope.Nonce = []byte("nonce")
	envelope.SenderPubKey = []byte("sender")
	envelope.SenderSignature = []byte("sig")

	metadata, err := types.ExtractEnvelopeMetadata(envelope)
	require.NoError(t, err)

	str := metadata.String()
	require.Contains(t, str, "Envelope{")
	require.Contains(t, str, "version=1")
	require.Contains(t, str, "recipients=2")
	require.Contains(t, str, "size=1024 bytes")
}

func TestEnvelopeValidationResult(t *testing.T) {
	t.Run("add error marks as invalid", func(t *testing.T) {
		result := &types.EnvelopeValidationResult{
			Valid: true,
		}

		result.AddError("test error")
		require.False(t, result.Valid)
		require.Len(t, result.Errors, 1)
		require.Equal(t, "test error", result.Errors[0])
	})

	t.Run("add warning keeps valid status", func(t *testing.T) {
		result := &types.EnvelopeValidationResult{
			Valid: true,
		}

		result.AddWarning("test warning")
		require.True(t, result.Valid)
		require.Len(t, result.Warnings, 1)
		require.Equal(t, "test warning", result.Warnings[0])
	})

	t.Run("set recipient status", func(t *testing.T) {
		result := &types.EnvelopeValidationResult{}

		status := types.RecipientStatus{
			KeyFound:  true,
			KeyActive: true,
			Address:   "virtengine1abc",
			Message:   "key is valid",
		}

		result.SetRecipientStatus("fingerprint1", status)
		require.Len(t, result.RecipientStatus, 1)

		retrievedStatus, ok := result.RecipientStatus["fingerprint1"]
		require.True(t, ok)
		require.True(t, retrievedStatus.KeyFound)
		require.True(t, retrievedStatus.KeyActive)
		require.Equal(t, "virtengine1abc", retrievedStatus.Address)
	})
}

func TestRecipientStatus(t *testing.T) {
	t.Run("active key", func(t *testing.T) {
		status := types.RecipientStatus{
			KeyFound:  true,
			KeyActive: true,
			Address:   "virtengine1xyz",
		}

		require.True(t, status.KeyFound)
		require.True(t, status.KeyActive)
		require.False(t, status.KeyExpired)
		require.False(t, status.KeyDeprecated)
	})

	t.Run("revoked key", func(t *testing.T) {
		status := types.RecipientStatus{
			KeyFound:  true,
			KeyActive: false,
			Address:   "virtengine1xyz",
			Message:   "key has been revoked",
		}

		require.True(t, status.KeyFound)
		require.False(t, status.KeyActive)
	})

	t.Run("key not found", func(t *testing.T) {
		status := types.RecipientStatus{
			KeyFound: false,
			Message:  "key not registered",
		}

		require.False(t, status.KeyFound)
	})
}
