package keeper

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	encryptioncrypto "github.com/virtengine/virtengine/x/encryption/crypto"
	"github.com/virtengine/virtengine/x/encryption/types"
)

type testReencryptionWorker struct {
	privateKeys map[string][]byte
}

func (w testReencryptionWorker) ReencryptEnvelope(ctx sdk.Context, envelope *types.EncryptedPayloadEnvelope, oldKey, newKey types.RecipientKeyRecord) (*types.EncryptedPayloadEnvelope, error) {
	privateKey, ok := w.privateKeys[oldKey.KeyFingerprint]
	if !ok {
		return nil, types.ErrDecryptionFailed.Wrap("missing private key")
	}

	plaintext, err := encryptioncrypto.OpenEnvelope(envelope, privateKey)
	if err != nil {
		return nil, err
	}

	senderKey, err := encryptioncrypto.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	recipientID := types.FormatRecipientKeyID(newKey.KeyFingerprint, newKey.KeyVersion)
	recipient := encryptioncrypto.RecipientInfo{
		PublicKey:  newKey.PublicKey,
		KeyID:      recipientID,
		KeyVersion: newKey.KeyVersion,
	}

	newEnvelope, err := encryptioncrypto.CreateEnvelopeWithRecipient(plaintext, recipient, senderKey)
	if err != nil {
		return nil, err
	}
	newEnvelope.Metadata = envelope.Metadata
	return newEnvelope, nil
}

func TestKeyRotationWithReencryption(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("address_123456789012345"))
	oldKeyPair, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)
	newKeyPair, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)

	oldFingerprint, err := k.RegisterRecipientKey(ctx, addr, oldKeyPair.PublicKey[:], types.AlgorithmX25519XSalsa20Poly1305, "old")
	require.NoError(t, err)

	senderKey, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)

	envelope, err := encryptioncrypto.CreateEnvelope([]byte("secret"), oldKeyPair.PublicKey[:], senderKey)
	require.NoError(t, err)

	hash, err := k.StoreEnvelope(ctx, envelope)
	require.NoError(t, err)

	_, err = k.RotateRecipientKey(ctx, addr, oldFingerprint, newKeyPair.PublicKey[:], types.AlgorithmX25519XSalsa20Poly1305, "new", "scheduled", 0)
	require.NoError(t, err)

	worker := testReencryptionWorker{
		privateKeys: map[string][]byte{
			oldFingerprint: oldKeyPair.PrivateKey[:],
		},
	}

	processed, err := k.ProcessReencryptionJobs(ctx, 10, worker)
	require.NoError(t, err)
	require.Equal(t, uint32(1), processed)

	record, found := k.GetEnvelope(ctx, hash)
	require.True(t, found)

	plaintext, err := encryptioncrypto.OpenEnvelope(&record.Envelope, newKeyPair.PrivateKey[:])
	require.NoError(t, err)
	require.Equal(t, "secret", string(plaintext))

	oldRecord, found := k.GetRecipientKeyByFingerprint(ctx, oldFingerprint)
	require.True(t, found)
	require.NotZero(t, oldRecord.DeprecatedAt)
}

func TestRevokedKeyRejected(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("address_123456789012345"))
	keyPair, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)

	fingerprint, err := k.RegisterRecipientKey(ctx, addr, keyPair.PublicKey[:], types.AlgorithmX25519XSalsa20Poly1305, "key")
	require.NoError(t, err)

	err = k.RevokeRecipientKey(ctx, addr, fingerprint)
	require.NoError(t, err)

	senderKey, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)

	envelope, err := encryptioncrypto.CreateEnvelope([]byte("secret"), keyPair.PublicKey[:], senderKey)
	require.NoError(t, err)

	_, err = k.ValidateEnvelopeRecipients(ctx, envelope)
	require.Error(t, err)
	require.ErrorIs(t, err, types.ErrKeyRevoked)
}

func TestKeyExpiryEnforced(t *testing.T) {
	ctx, k := setupKeeper(t)

	params := types.DefaultParams()
	params.DefaultKeyTtlSeconds = 1
	require.NoError(t, k.SetParams(ctx, params))

	addr := sdk.AccAddress([]byte("address_123456789012345"))
	keyPair, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)

	fingerprint, err := k.RegisterRecipientKey(ctx, addr, keyPair.PublicKey[:], types.AlgorithmX25519XSalsa20Poly1305, "key")
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Second))
	warnings, expired := k.HandleKeyExpiry(ctx)
	require.Equal(t, uint32(0), warnings)
	require.Equal(t, uint32(1), expired)

	record, found := k.GetRecipientKeyByFingerprint(ctx, fingerprint)
	require.True(t, found)
	require.NotZero(t, record.RevokedAt)
	require.NotZero(t, record.ExpiresAt)
}

func TestResolveRecipientKeyVersion(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("address_123456789012345"))
	key1, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)
	key2, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)

	fp1, err := k.RegisterRecipientKey(ctx, addr, key1.PublicKey[:], types.AlgorithmX25519XSalsa20Poly1305, "v1")
	require.NoError(t, err)
	_, err = k.RegisterRecipientKey(ctx, addr, key2.PublicKey[:], types.AlgorithmX25519XSalsa20Poly1305, "v2")
	require.NoError(t, err)

	senderKey, err := encryptioncrypto.GenerateKeyPair()
	require.NoError(t, err)
	envelope, err := encryptioncrypto.CreateEnvelopeWithRecipient([]byte("payload"), encryptioncrypto.RecipientInfo{
		PublicKey:  key1.PublicKey[:],
		KeyID:      types.FormatRecipientKeyID(fp1, 1),
		KeyVersion: 1,
	}, senderKey)
	require.NoError(t, err)

	resolved, found := k.ResolveRecipientKeyID(ctx, addr, envelope.RecipientKeyIDs[0])
	require.True(t, found)
	require.Equal(t, fp1, resolved.KeyFingerprint)
	require.Equal(t, uint32(1), resolved.KeyVersion)
}

func TestEphemeralKeyLifecycle(t *testing.T) {
	ctx, k := setupKeeper(t)

	addr := sdk.AccAddress([]byte("address_123456789012345"))
	sessionID, pubKey, privKey, err := k.CreateEphemeralKey(ctx, addr, 5)
	require.NoError(t, err)
	require.NotEmpty(t, sessionID)

	shared, err := k.DeriveSharedSecret(privKey, pubKey)
	require.NoError(t, err)
	require.Len(t, shared, 32)

	err = k.UseEphemeralKey(ctx, sessionID)
	require.NoError(t, err)

	err = k.UseEphemeralKey(ctx, sessionID)
	require.Error(t, err)

	sessionID2, _, _, err := k.CreateEphemeralKey(ctx, addr, 1)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(2 * time.Second))
	err = k.UseEphemeralKey(ctx, sessionID2)
	require.ErrorIs(t, err, types.ErrKeyExpired)
}
