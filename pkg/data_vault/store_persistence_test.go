package data_vault

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/virtengine/virtengine/pkg/artifact_store"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
)

type metadataPersistenceArtifactStore struct {
	*memoryArtifactStore
	mu         sync.Mutex
	metadata   map[BlobID]*BlobMetadata
	failSave   bool
	beforeSave func(map[BlobID]*BlobMetadata) error
}

func newMetadataPersistenceArtifactStore() *metadataPersistenceArtifactStore {
	return &metadataPersistenceArtifactStore{
		memoryArtifactStore: newMemoryArtifactStore(),
		metadata:            make(map[BlobID]*BlobMetadata),
	}
}

func (s *metadataPersistenceArtifactStore) LoadVaultMetadata() (map[BlobID]*BlobMetadata, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return cloneBlobMetadata(s.metadata), nil
}

func (s *metadataPersistenceArtifactStore) SaveVaultMetadata(metadata map[BlobID]*BlobMetadata) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.beforeSave != nil {
		if err := s.beforeSave(metadata); err != nil {
			return err
		}
	}
	if s.failSave {
		return errors.New("metadata persistence failed")
	}
	s.metadata = cloneBlobMetadata(metadata)
	return nil
}

func (s *metadataPersistenceArtifactStore) PutVaultBlob(ctx context.Context, req *artifact_store.PutRequest, blobID BlobID, metadataFactory func(*artifact_store.PutResponse) *BlobMetadata) (*artifact_store.PutResponse, error) {
	response, err := s.Put(ctx, req)
	if err != nil {
		return nil, err
	}
	metadata := metadataFactory(response)
	s.mu.Lock()
	s.metadata[blobID] = cloneBlobMetadataValue(metadata)
	s.mu.Unlock()
	return response, nil
}

func (s *metadataPersistenceArtifactStore) DeleteVaultBlob(ctx context.Context, req *artifact_store.DeleteRequest, blobID BlobID) error {
	if err := s.Delete(ctx, req); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.metadata, blobID)
	s.mu.Unlock()
	return nil
}

func TestReencryptMetadataPersistenceFailureDoesNotSwapLiveMetadata(t *testing.T) {
	ctx := context.Background()
	keyManager := keys.NewKeyManager()
	require.NoError(t, keyManager.Initialize())
	backend := newMetadataPersistenceArtifactStore()
	store, err := NewEncryptedBlobStoreWithError(backend, keyManager)
	require.NoError(t, err)
	blob, err := store.Store(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("secret"), Owner: "owner"})
	require.NoError(t, err)
	oldKey, err := keyManager.GetActiveKey(keys.ScopeSupport)
	require.NoError(t, err)
	require.NoError(t, keyManager.RotateKey(keys.ScopeSupport, time.Hour))
	rotation, err := keyManager.GetRotationStatus(keys.ScopeSupport)
	require.NoError(t, err)
	newKey, err := keyManager.GetKey(keys.ScopeSupport, rotation.NewKeyID)
	require.NoError(t, err)

	backend.failSave = true
	_, err = store.Reencrypt(ctx, blob.Metadata.ID, oldKey, newKey)
	require.ErrorContains(t, err, "metadata persistence failed")
	metadata, err := store.GetMetadata(blob.Metadata.ID)
	require.NoError(t, err)
	require.Equal(t, oldKey.ID, metadata.KeyID)

	backend.failSave = false
	backend.beforeSave = func(candidate map[BlobID]*BlobMetadata) error {
		require.Equal(t, oldKey.ID, store.metadata[blob.Metadata.ID].KeyID)
		require.Equal(t, newKey.ID, candidate[blob.Metadata.ID].KeyID)
		return nil
	}
	_, err = store.Reencrypt(ctx, blob.Metadata.ID, oldKey, newKey)
	require.NoError(t, err)
	metadata, err = store.GetMetadata(blob.Metadata.ID)
	require.NoError(t, err)
	require.Equal(t, newKey.ID, metadata.KeyID)
}

func TestReencryptConcurrentMetadataSavesDoNotLoseUpdates(t *testing.T) {
	ctx := context.Background()
	keyManager := keys.NewKeyManager()
	require.NoError(t, keyManager.Initialize())
	backend := newMetadataPersistenceArtifactStore()
	store, err := NewEncryptedBlobStoreWithError(backend, keyManager)
	require.NoError(t, err)
	first, err := store.Store(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("first"), Owner: "owner"})
	require.NoError(t, err)
	second, err := store.Store(ctx, &UploadRequest{Scope: ScopeSupport, Plaintext: []byte("second"), Owner: "owner"})
	require.NoError(t, err)
	oldKey, err := keyManager.GetActiveKey(keys.ScopeSupport)
	require.NoError(t, err)
	require.NoError(t, keyManager.RotateKey(keys.ScopeSupport, time.Hour))
	rotation, err := keyManager.GetRotationStatus(keys.ScopeSupport)
	require.NoError(t, err)
	newKey, err := keyManager.GetKey(keys.ScopeSupport, rotation.NewKeyID)
	require.NoError(t, err)

	var waitGroup sync.WaitGroup
	errorsByBlob := make(chan error, 2)
	for _, blobID := range []BlobID{first.Metadata.ID, second.Metadata.ID} {
		waitGroup.Add(1)
		go func(id BlobID) {
			defer waitGroup.Done()
			_, reencryptErr := store.Reencrypt(ctx, id, oldKey, newKey)
			errorsByBlob <- reencryptErr
		}(blobID)
	}
	waitGroup.Wait()
	close(errorsByBlob)
	for reencryptErr := range errorsByBlob {
		require.NoError(t, reencryptErr)
	}
	for _, blobID := range []BlobID{first.Metadata.ID, second.Metadata.ID} {
		metadata, metadataErr := store.GetMetadata(blobID)
		require.NoError(t, metadataErr)
		require.Equal(t, newKey.ID, metadata.KeyID)
	}
	require.Equal(t, newKey.ID, backend.metadata[first.Metadata.ID].KeyID)
	require.Equal(t, newKey.ID, backend.metadata[second.Metadata.ID].KeyID)
}
