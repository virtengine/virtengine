package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/encryption/crypto"
	"github.com/virtengine/virtengine/x/encryption/types"
)

func TestCheckEnvelopeAccess(t *testing.T) {
	ctx, k := setupKeeper(t)

	// Generate two key pairs for testing
	recipientKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)

	nonRecipientKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)

	// Create addresses
	recipientAddr := sdk.AccAddress([]byte("recipient_address__"))
	nonRecipientAddr := sdk.AccAddress([]byte("non_recipient_addr_"))

	// Register recipient key
	recipientFingerprint, err := k.RegisterRecipientKey(
		ctx,
		recipientAddr,
		recipientKeyPair.PublicKey[:],
		types.DefaultAlgorithm(),
		"test-recipient",
	)
	require.NoError(t, err)

	// Register non-recipient key
	_, err = k.RegisterRecipientKey(
		ctx,
		nonRecipientAddr,
		nonRecipientKeyPair.PublicKey[:],
		types.DefaultAlgorithm(),
		"test-non-recipient",
	)
	require.NoError(t, err)

	// Create an envelope with only recipient as authorized
	plaintext := []byte("sensitive data")
	senderKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)

	envelope, err := crypto.CreateEnvelope(plaintext, recipientKeyPair.PublicKey[:], senderKeyPair)
	require.NoError(t, err)

	// Update envelope to use correct fingerprint
	envelope.RecipientKeyIDs = []string{recipientFingerprint}

	t.Run("access granted to recipient", func(t *testing.T) {
		err := k.CheckEnvelopeAccess(ctx, envelope, recipientAddr)
		require.NoError(t, err)
	})

	t.Run("access denied to non-recipient", func(t *testing.T) {
		err := k.CheckEnvelopeAccess(ctx, envelope, nonRecipientAddr)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is not a recipient")
	})

	t.Run("access denied to nil envelope", func(t *testing.T) {
		err := k.CheckEnvelopeAccess(ctx, nil, recipientAddr)
		require.Error(t, err)
		require.Contains(t, err.Error(), "envelope cannot be nil")
	})

	t.Run("access denied to empty requester", func(t *testing.T) {
		err := k.CheckEnvelopeAccess(ctx, envelope, sdk.AccAddress{})
		require.Error(t, err)
		require.Contains(t, err.Error(), "requester address is required")
	})
}

func TestCheckEnvelopeAccessByFingerprint(t *testing.T) {
	ctx, k := setupKeeper(t)

	recipientKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)

	recipientAddr := sdk.AccAddress([]byte("recipient_address__"))

	recipientFingerprint, err := k.RegisterRecipientKey(
		ctx,
		recipientAddr,
		recipientKeyPair.PublicKey[:],
		types.DefaultAlgorithm(),
		"test-recipient",
	)
	require.NoError(t, err)

	plaintext := []byte("sensitive data")
	senderKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)

	envelope, err := crypto.CreateEnvelope(plaintext, recipientKeyPair.PublicKey[:], senderKeyPair)
	require.NoError(t, err)
	envelope.RecipientKeyIDs = []string{recipientFingerprint}

	t.Run("access granted with valid fingerprint", func(t *testing.T) {
		err := k.CheckEnvelopeAccessByFingerprint(ctx, envelope, recipientFingerprint)
		require.NoError(t, err)
	})

	t.Run("access denied with invalid fingerprint", func(t *testing.T) {
		err := k.CheckEnvelopeAccessByFingerprint(ctx, envelope, "invalid_fingerprint")
		require.Error(t, err)
		require.Contains(t, err.Error(), "is not a recipient")
	})

	t.Run("access denied with revoked key", func(t *testing.T) {
		// Revoke the key
		err := k.RevokeRecipientKey(ctx, recipientAddr, recipientFingerprint)
		require.NoError(t, err)

		err = k.CheckEnvelopeAccessByFingerprint(ctx, envelope, recipientFingerprint)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is revoked")
	})
}

func TestValidateAndCheckAccess(t *testing.T) {
	ctx, k := setupKeeper(t)

	recipientKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)

	recipientAddr := sdk.AccAddress([]byte("recipient_address__"))

	recipientFingerprint, err := k.RegisterRecipientKey(
		ctx,
		recipientAddr,
		recipientKeyPair.PublicKey[:],
		types.DefaultAlgorithm(),
		"test-recipient",
	)
	require.NoError(t, err)

	plaintext := []byte("sensitive data")
	senderKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)

	envelope, err := crypto.CreateEnvelope(plaintext, recipientKeyPair.PublicKey[:], senderKeyPair)
	require.NoError(t, err)
	envelope.RecipientKeyIDs = []string{recipientFingerprint}

	t.Run("valid envelope with authorized access", func(t *testing.T) {
		err := k.ValidateAndCheckAccess(ctx, envelope, recipientAddr)
		require.NoError(t, err)
	})

	t.Run("invalid envelope fails validation before access check", func(t *testing.T) {
		invalidEnvelope := &types.EncryptedPayloadEnvelope{
			Version: 0, // Invalid version
		}
		err := k.ValidateAndCheckAccess(ctx, invalidEnvelope, recipientAddr)
		require.Error(t, err)
		require.Contains(t, err.Error(), "version cannot be zero")
	})
}

func TestGetEnvelopeRecipients(t *testing.T) {
	ctx, k := setupKeeper(t)

	// Register multiple recipients
	keyPair1, err := crypto.GenerateKeyPair()
	require.NoError(t, err)
	keyPair2, err := crypto.GenerateKeyPair()
	require.NoError(t, err)

	addr1 := sdk.AccAddress([]byte("recipient1_address_"))
	addr2 := sdk.AccAddress([]byte("recipient2_address_"))

	fingerprint1, err := k.RegisterRecipientKey(ctx, addr1, keyPair1.PublicKey[:], types.DefaultAlgorithm(), "recipient1")
	require.NoError(t, err)

	fingerprint2, err := k.RegisterRecipientKey(ctx, addr2, keyPair2.PublicKey[:], types.DefaultAlgorithm(), "recipient2")
	require.NoError(t, err)

	// Create envelope with both recipients
	envelope := types.NewEncryptedPayloadEnvelope()
	envelope.RecipientKeyIDs = []string{fingerprint1, fingerprint2}
	envelope.Nonce = []byte("test_nonce_12345")
	envelope.Ciphertext = []byte("encrypted_data")
	envelope.SenderPubKey = keyPair1.PublicKey[:]
	envelope.SenderSignature = []byte("signature")

	recipients, err := k.GetEnvelopeRecipients(ctx, envelope)
	require.NoError(t, err)
	require.Len(t, recipients, 2)
	require.Contains(t, recipients, addr1)
	require.Contains(t, recipients, addr2)
}

func TestEnforceEncryptedPayloadRequired(t *testing.T) {
	_, k := setupKeeper(t)

	t.Run("nil envelope returns error", func(t *testing.T) {
		err := k.EnforceEncryptedPayloadRequired(nil, "test_field")
		require.Error(t, err)
		require.Contains(t, err.Error(), "test_field")
		require.Contains(t, err.Error(), "required but missing")
	})

	t.Run("empty ciphertext returns error", func(t *testing.T) {
		envelope := types.NewEncryptedPayloadEnvelope()
		envelope.RecipientKeyIDs = []string{"fingerprint"}
		err := k.EnforceEncryptedPayloadRequired(envelope, "test_field")
		require.Error(t, err)
		require.Contains(t, err.Error(), "ciphertext is empty")
	})

	t.Run("no recipients returns error", func(t *testing.T) {
		envelope := types.NewEncryptedPayloadEnvelope()
		envelope.Ciphertext = []byte("encrypted")
		err := k.EnforceEncryptedPayloadRequired(envelope, "test_field")
		require.Error(t, err)
		require.Contains(t, err.Error(), "no recipients specified")
	})

	t.Run("valid envelope passes", func(t *testing.T) {
		envelope := types.NewEncryptedPayloadEnvelope()
		envelope.Ciphertext = []byte("encrypted")
		envelope.RecipientKeyIDs = []string{"fingerprint"}
		err := k.EnforceEncryptedPayloadRequired(envelope, "test_field")
		require.NoError(t, err)
	})
}
