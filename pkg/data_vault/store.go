package data_vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/artifact_store"
	"github.com/virtengine/virtengine/pkg/data_vault/keys"
	enccrypto "github.com/virtengine/virtengine/x/encryption/crypto"
	enctypes "github.com/virtengine/virtengine/x/encryption/types"
)

// EncryptedBlobStore wraps an artifact_store with encryption using the vault's key management
type EncryptedBlobStore struct {
	backend artifact_store.ArtifactStore
	keyMgr  *keys.KeyManager
	mu      sync.RWMutex

	// metadata stores blob metadata indexed by blob ID
	metadata map[BlobID]*BlobMetadata
}

// NewEncryptedBlobStore creates a new encrypted blob store
func NewEncryptedBlobStore(backend artifact_store.ArtifactStore, keyMgr *keys.KeyManager) *EncryptedBlobStore {
	return &EncryptedBlobStore{
		backend:  backend,
		keyMgr:   keyMgr,
		metadata: make(map[BlobID]*BlobMetadata),
	}
}

// Store encrypts and stores a blob in the artifact store
func (s *EncryptedBlobStore) Store(ctx context.Context, req *UploadRequest) (*EncryptedBlob, error) {
	if req == nil {
		return nil, NewVaultError("Store", ErrInvalidRequest, "request cannot be nil")
	}

	if len(req.Plaintext) == 0 {
		return nil, NewVaultError("Store", ErrInvalidRequest, "plaintext cannot be empty")
	}

	if req.Owner == "" {
		return nil, NewVaultError("Store", ErrInvalidRequest, "owner cannot be empty")
	}

	// Get the active encryption key for this scope
	keyInfo, err := s.keyMgr.GetActiveKey(keys.Scope(req.Scope))
	if err != nil {
		return nil, NewVaultError("Store", err, "failed to get encryption key")
	}

	// Generate sender key pair for this blob
	senderKeyPair, err := enccrypto.GenerateKeyPair()
	if err != nil {
		return nil, NewVaultError("Store", ErrEncryptionFailed, fmt.Sprintf("failed to generate sender keypair: %v", err))
	}

	// Create encrypted envelope
	envelope, err := enccrypto.CreateEnvelope(req.Plaintext, keyInfo.PublicKey[:], senderKeyPair)
	if err != nil {
		return nil, NewVaultError("Store", ErrEncryptionFailed, fmt.Sprintf("failed to create envelope: %v", err))
	}

	// Serialize envelope for storage
	envelopeBytes, err := json.Marshal(envelope)
	if err != nil {
		return nil, NewVaultError("Store", ErrEncryptionFailed, fmt.Sprintf("failed to marshal envelope: %v", err))
	}

	// Compute content hash of plaintext
	contentHash := sha256.Sum256(req.Plaintext)

	// Generate blob ID (hash of content + timestamp for uniqueness)
	blobIDBytes := sha256.Sum256(append(contentHash[:], []byte(time.Now().String())...))
	blobID := BlobID(hex.EncodeToString(blobIDBytes[:]))

	// Compute envelope hash for metadata
	envelopeHash := sha256.Sum256(envelopeBytes)

	// Create artifact store metadata
	artifactMeta := &artifact_store.EncryptionMetadata{
		AlgorithmID:     string(envelope.AlgorithmID),
		RecipientKeyIDs: envelope.RecipientKeyIDs,
		EnvelopeHash:    envelopeHash[:],
		SenderKeyID:     hex.EncodeToString(envelope.SenderPubKey),
	}

	retentionTag := &artifact_store.RetentionTag{
		PolicyID: req.RetentionPolicy,
		Owner:    req.Owner,
	}
	if req.ExpiresAt != nil {
		retentionTag.ExpiresAt = req.ExpiresAt
	}

	// Store in artifact backend
	putReq := &artifact_store.PutRequest{
		Data:               envelopeBytes,
		ContentHash:        contentHash[:],
		EncryptionMetadata: artifactMeta,
		RetentionTag:       retentionTag,
		Owner:              req.Owner,
		ArtifactType:       string(req.Scope),
		Metadata:           req.Tags,
	}

	putResp, err := s.backend.Put(ctx, putReq)
	if err != nil {
		return nil, NewVaultError("Store", ErrStorageBackend, fmt.Sprintf("backend store failed: %v", err))
	}

	// Create blob metadata
	metadata := &BlobMetadata{
		ID:              blobID,
		Scope:           req.Scope,
		KeyID:           keyInfo.ID,
		KeyVersion:      keyInfo.Version,
		ContentHash:     contentHash[:],
		Size:            int64(len(req.Plaintext)),
		EncryptedSize:   int64(len(envelopeBytes)),
		Owner:           req.Owner,
		OrgID:           req.OrgID,
		CreatedAt:       time.Now(),
		ExpiresAt:       req.ExpiresAt,
		RetentionPolicy: req.RetentionPolicy,
		Tags:            req.Tags,
	}

	// Store metadata
	s.mu.Lock()
	s.metadata[blobID] = metadata
	s.mu.Unlock()

	return &EncryptedBlob{
		Metadata:    *metadata,
		Envelope:    envelope,
		BackendPath: putResp.ContentAddress.BackendRef,
	}, nil
}

// Retrieve retrieves and decrypts a blob from the artifact store
func (s *EncryptedBlobStore) Retrieve(ctx context.Context, blobID BlobID) ([]byte, *BlobMetadata, error) {
	// Get metadata
	s.mu.RLock()
	metadata, exists := s.metadata[blobID]
	s.mu.RUnlock()

	if !exists {
		return nil, nil, NewVaultError("Retrieve", ErrBlobNotFound, string(blobID))
	}

	// Check expiration
	if metadata.ExpiresAt != nil && time.Now().After(*metadata.ExpiresAt) {
		return nil, nil, NewVaultError("Retrieve", ErrBlobExpired, string(blobID))
	}

	// Get the decryption key (try current and historical versions)
	keyInfo, err := s.keyMgr.GetKey(keys.Scope(metadata.Scope), metadata.KeyID)
	if err != nil {
		return nil, nil, NewVaultError("Retrieve", ErrInvalidKey, fmt.Sprintf("failed to get key %s: %v", metadata.KeyID, err))
	}

	// Retrieve encrypted envelope from backend
	// Note: We need to reverse lookup by blob ID to content address
	// For now, we'll need to enhance this with a mapping
	// This is a simplified implementation
	getReq := &artifact_store.GetRequest{
		ContentAddress: &artifact_store.ContentAddress{
			Backend:    s.backend.Backend(),
			BackendRef: string(blobID), // Simplified - would need proper mapping
		},
	}

	getResp, err := s.backend.Get(ctx, getReq)
	if err != nil {
		return nil, nil, NewVaultError("Retrieve", ErrStorageBackend, fmt.Sprintf("backend get failed: %v", err))
	}

	// Unmarshal envelope
	var envelope enctypes.EncryptedPayloadEnvelope
	if err := json.Unmarshal(getResp.Data, &envelope); err != nil {
		return nil, nil, NewVaultError("Retrieve", ErrDecryptionFailed, fmt.Sprintf("failed to unmarshal envelope: %v", err))
	}

	// Decrypt envelope
	plaintext, err := enccrypto.OpenEnvelope(&envelope, keyInfo.PrivateKey[:])
	if err != nil {
		return nil, nil, NewVaultError("Retrieve", ErrDecryptionFailed, fmt.Sprintf("failed to decrypt envelope: %v", err))
	}

	// Verify content hash
	computedHash := sha256.Sum256(plaintext)
	if hex.EncodeToString(computedHash[:]) != hex.EncodeToString(metadata.ContentHash) {
		return nil, nil, NewVaultError("Retrieve", ErrDecryptionFailed, "content hash mismatch")
	}

	return plaintext, metadata, nil
}

// GetMetadata retrieves blob metadata without decrypting content
func (s *EncryptedBlobStore) GetMetadata(blobID BlobID) (*BlobMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metadata, exists := s.metadata[blobID]
	if !exists {
		return nil, NewVaultError("GetMetadata", ErrBlobNotFound, string(blobID))
	}

	// Return a copy to prevent mutation
	metadataCopy := *metadata
	return &metadataCopy, nil
}

// Delete marks a blob for deletion
func (s *EncryptedBlobStore) Delete(ctx context.Context, blobID BlobID) error {
	// Get metadata to find backend path
	s.mu.RLock()
	metadata, exists := s.metadata[blobID]
	s.mu.RUnlock()

	if !exists {
		return NewVaultError("Delete", ErrBlobNotFound, string(blobID))
	}

	// Delete from backend
	deleteReq := &artifact_store.DeleteRequest{
		ContentAddress: &artifact_store.ContentAddress{
			Backend:    s.backend.Backend(),
			BackendRef: string(blobID), // Simplified mapping
		},
		RequestingAccount: metadata.Owner,
		Force:             true,
	}

	if err := s.backend.Delete(ctx, deleteReq); err != nil {
		return NewVaultError("Delete", ErrStorageBackend, fmt.Sprintf("backend delete failed: %v", err))
	}

	// Remove metadata
	s.mu.Lock()
	delete(s.metadata, blobID)
	s.mu.Unlock()

	return nil
}

// ListByScope lists all blobs in a scope
func (s *EncryptedBlobStore) ListByScope(scope Scope) ([]*BlobMetadata, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*BlobMetadata
	for _, metadata := range s.metadata {
		if metadata.Scope == scope {
			metadataCopy := *metadata
			result = append(result, &metadataCopy)
		}
	}

	return result, nil
}

// Close closes the encrypted blob store
func (s *EncryptedBlobStore) Close() error {
	// Nothing to clean up for now
	return nil
}
