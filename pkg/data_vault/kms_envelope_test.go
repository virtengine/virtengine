package data_vault

import (
	"context"
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
)

type testKMSProvider struct{ key []byte }

func (p testKMSProvider) GenerateDataKey(_ context.Context, _ string, _ map[string]string) (KMSDataKey, error) {
	wrapped := sha256.Sum256(p.key)
	return KMSDataKey{Plaintext: append([]byte(nil), p.key...), Wrapped: wrapped[:]}, nil
}
func (p testKMSProvider) DecryptDataKey(_ context.Context, _ string, wrapped []byte, _ map[string]string) ([]byte, error) {
	expected := sha256.Sum256(p.key)
	if string(wrapped) != string(expected[:]) {
		return nil, context.Canceled
	}
	return append([]byte(nil), p.key...), nil
}

func TestKMSEnvelopeCipherStoreRetrieveAndReencrypt(t *testing.T) {
	manager := keys.NewKeyManager()
	require.NoError(t, manager.Initialize())
	cipher, err := NewKMSEnvelopeCipher(testKMSProvider{key: []byte("0123456789abcdef0123456789abcdef")})
	require.NoError(t, err)
	store, err := NewEncryptedBlobStoreWithCipher(newMemoryArtifactStore(), manager, cipher)
	require.NoError(t, err)
	stored, err := store.Store(context.Background(), &UploadRequest{Scope: ScopeVEID, Owner: "owner", Plaintext: []byte("identity-document")})
	require.NoError(t, err)
	retrieved, _, err := store.Retrieve(context.Background(), stored.Metadata.ID)
	require.NoError(t, err)
	require.Equal(t, []byte("identity-document"), retrieved)
	oldKey, err := manager.GetActiveKey(keys.ScopeVEID)
	require.NoError(t, err)
	require.NoError(t, manager.RotateKey(keys.ScopeVEID, 0))
	rotation, err := manager.GetRotationStatus(keys.ScopeVEID)
	require.NoError(t, err)
	newKey, err := manager.GetKey(keys.ScopeVEID, rotation.NewKeyID)
	require.NoError(t, err)
	_, err = store.Reencrypt(context.Background(), stored.Metadata.ID, oldKey, newKey)
	require.NoError(t, err)
	retrieved, metadata, err := store.Retrieve(context.Background(), stored.Metadata.ID)
	require.NoError(t, err)
	require.Equal(t, []byte("identity-document"), retrieved)
	require.Equal(t, newKey.ID, metadata.KeyID)
}
