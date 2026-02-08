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

	recipients := buildRecipientInfos(keyInfo, req.Recipients)

	// Create encrypted envelope
	var envelope *enctypes.EncryptedPayloadEnvelope
	if len(recipients) > 1 {
		envelope, err = enccrypto.CreateMultiRecipientEnvelopeWithRecipients(req.Plaintext, recipients, senderKeyPair)
	} else {
		envelope, err = enccrypto.CreateEnvelopeWithRecipient(req.Plaintext, recipients[0], senderKeyPair)
	}
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

	backendRef := ""
	backendName := ""
	if putResp.ContentAddress != nil {
		backendRef = putResp.ContentAddress.BackendRef
		backendName = string(putResp.ContentAddress.Backend)
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
		Backend:         backendName,
		BackendRef:      backendRef,
	}

	if putResp.ContentAddress != nil {
		metadata.ContentAddressHash = putResp.ContentAddress.Hash
		metadata.ContentAddressSize = putResp.ContentAddress.Size
		metadata.ContentAddressAlgorithm = putResp.ContentAddress.Algorithm
		metadata.ContentAddressVersion = putResp.ContentAddress.Version
	}

	// Store metadata
	s.mu.Lock()
	s.metadata[blobID] = metadata
	s.mu.Unlock()

	return &EncryptedBlob{
		Metadata:    *metadata,
		Envelope:    envelope,
		BackendPath: backendRef,
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
	contentAddress := s.resolveContentAddress(metadata, blobID)

	getReq := &artifact_store.GetRequest{
		ContentAddress: contentAddress,
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
	deleteAddress := s.resolveContentAddress(metadata, blobID)

	deleteReq := &artifact_store.DeleteRequest{
		ContentAddress:    deleteAddress,
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

// KeyManager returns the key manager used by the store.
func (s *EncryptedBlobStore) KeyManager() *keys.KeyManager {
	return s.keyMgr
}

// Reencrypt rewraps an existing blob with a new key.
func (s *EncryptedBlobStore) Reencrypt(ctx context.Context, blobID BlobID, oldKey, newKey *keys.KeyInfo) (*EncryptedBlob, error) {
	if oldKey == nil || newKey == nil {
		return nil, NewVaultError("Reencrypt", ErrInvalidRequest, "keys required")
	}

	plaintext, metadata, err := s.Retrieve(ctx, blobID)
	if err != nil {
		return nil, err
	}

	envelope, err := s.loadEnvelope(ctx, metadata, blobID)
	if err != nil {
		return nil, err
	}

	recipients := rebuildRecipients(envelope, oldKey, newKey)
	if len(recipients) == 0 {
		recipients = buildRecipientInfos(newKey, nil)
	}

	senderKeyPair, err := enccrypto.GenerateKeyPair()
	if err != nil {
		return nil, NewVaultError("Reencrypt", ErrEncryptionFailed, fmt.Sprintf("failed to generate sender keypair: %v", err))
	}

	var newEnvelope *enctypes.EncryptedPayloadEnvelope
	if len(recipients) > 1 {
		newEnvelope, err = enccrypto.CreateMultiRecipientEnvelopeWithRecipients(plaintext, recipients, senderKeyPair)
	} else {
		newEnvelope, err = enccrypto.CreateEnvelopeWithRecipient(plaintext, recipients[0], senderKeyPair)
	}
	if err != nil {
		return nil, NewVaultError("Reencrypt", ErrEncryptionFailed, fmt.Sprintf("failed to create envelope: %v", err))
	}

	envelopeBytes, err := json.Marshal(newEnvelope)
	if err != nil {
		return nil, NewVaultError("Reencrypt", ErrEncryptionFailed, fmt.Sprintf("failed to marshal envelope: %v", err))
	}

	contentHash := sha256.Sum256(plaintext)
	envelopeHash := sha256.Sum256(envelopeBytes)

	artifactMeta := &artifact_store.EncryptionMetadata{
		AlgorithmID:     string(newEnvelope.AlgorithmID),
		RecipientKeyIDs: newEnvelope.RecipientKeyIDs,
		EnvelopeHash:    envelopeHash[:],
		SenderKeyID:     hex.EncodeToString(newEnvelope.SenderPubKey),
	}

	retentionTag := &artifact_store.RetentionTag{
		PolicyID: metadata.RetentionPolicy,
		Owner:    metadata.Owner,
	}
	if metadata.ExpiresAt != nil {
		retentionTag.ExpiresAt = metadata.ExpiresAt
	}

	putReq := &artifact_store.PutRequest{
		Data:               envelopeBytes,
		ContentHash:        contentHash[:],
		EncryptionMetadata: artifactMeta,
		RetentionTag:       retentionTag,
		Owner:              metadata.Owner,
		ArtifactType:       string(metadata.Scope),
		Metadata:           metadata.Tags,
	}

	putResp, err := s.backend.Put(ctx, putReq)
	if err != nil {
		return nil, NewVaultError("Reencrypt", ErrStorageBackend, fmt.Sprintf("backend store failed: %v", err))
	}

	backendRef := ""
	backendName := ""
	if putResp.ContentAddress != nil {
		backendRef = putResp.ContentAddress.BackendRef
		backendName = string(putResp.ContentAddress.Backend)
	}

	metadata.KeyID = newKey.ID
	metadata.KeyVersion = newKey.Version
	metadata.ContentHash = contentHash[:]
	metadata.EncryptedSize = int64(len(envelopeBytes))
	metadata.Backend = backendName
	metadata.BackendRef = backendRef

	if putResp.ContentAddress != nil {
		metadata.ContentAddressHash = putResp.ContentAddress.Hash
		metadata.ContentAddressSize = putResp.ContentAddress.Size
		metadata.ContentAddressAlgorithm = putResp.ContentAddress.Algorithm
		metadata.ContentAddressVersion = putResp.ContentAddress.Version
	}

	s.mu.Lock()
	s.metadata[blobID] = metadata
	s.mu.Unlock()

	return &EncryptedBlob{
		Metadata:    *metadata,
		Envelope:    newEnvelope,
		BackendPath: backendRef,
	}, nil
}

func (s *EncryptedBlobStore) resolveContentAddress(metadata *BlobMetadata, blobID BlobID) *artifact_store.ContentAddress {
	backendRef := metadata.BackendRef
	if backendRef == "" {
		backendRef = string(blobID)
	}
	backendName := s.backend.Backend()
	if metadata.Backend != "" {
		backendName = artifact_store.BackendType(metadata.Backend)
	}
	contentAddress := &artifact_store.ContentAddress{
		Version:    metadata.ContentAddressVersion,
		Hash:       metadata.ContentAddressHash,
		Algorithm:  metadata.ContentAddressAlgorithm,
		Size:       metadata.ContentAddressSize,
		Backend:    backendName,
		BackendRef: backendRef,
	}
	if len(contentAddress.Hash) == 0 || contentAddress.BackendRef == "" {
		contentAddress = &artifact_store.ContentAddress{
			Version:    artifact_store.ContentAddressVersion,
			Hash:       metadata.ContentHash,
			Algorithm:  "sha256",
			Size:       safeUint64FromInt64(metadata.EncryptedSize),
			Backend:    backendName,
			BackendRef: backendRef,
		}
	}
	return contentAddress
}

func (s *EncryptedBlobStore) loadEnvelope(ctx context.Context, metadata *BlobMetadata, blobID BlobID) (*enctypes.EncryptedPayloadEnvelope, error) {
	contentAddress := s.resolveContentAddress(metadata, blobID)
	getResp, err := s.backend.Get(ctx, &artifact_store.GetRequest{ContentAddress: contentAddress})
	if err != nil {
		return nil, NewVaultError("Retrieve", ErrStorageBackend, fmt.Sprintf("backend get failed: %v", err))
	}
	var envelope enctypes.EncryptedPayloadEnvelope
	if err := json.Unmarshal(getResp.Data, &envelope); err != nil {
		return nil, NewVaultError("Retrieve", ErrDecryptionFailed, fmt.Sprintf("failed to unmarshal envelope: %v", err))
	}
	return &envelope, nil
}

func buildRecipientInfos(keyInfo *keys.KeyInfo, extra []Recipient) []enccrypto.RecipientInfo {
	recipients := make([]enccrypto.RecipientInfo, 0, 1+len(extra))
	seen := map[string]bool{}
	if keyInfo != nil {
		fingerprint := enctypes.ComputeKeyFingerprint(keyInfo.PublicKey[:])
		seen[fingerprint] = true
		recipients = append(recipients, enccrypto.RecipientInfo{
			PublicKey:  keyInfo.PublicKey[:],
			KeyID:      enctypes.FormatRecipientKeyID(fingerprint, keyInfo.Version),
			KeyVersion: keyInfo.Version,
		})
	}
	for _, rec := range extra {
		if len(rec.PublicKey) == 0 {
			continue
		}
		fingerprint := enctypes.ComputeKeyFingerprint(rec.PublicKey)
		if seen[fingerprint] {
			continue
		}
		seen[fingerprint] = true

		keyID := rec.KeyID
		if keyID == "" {
			keyID = enctypes.FormatRecipientKeyID(fingerprint, rec.KeyVersion)
		}
		recipients = append(recipients, enccrypto.RecipientInfo{
			PublicKey:  rec.PublicKey,
			KeyID:      keyID,
			KeyVersion: rec.KeyVersion,
		})
	}

	return recipients
}

func rebuildRecipients(envelope *enctypes.EncryptedPayloadEnvelope, oldKey, newKey *keys.KeyInfo) []enccrypto.RecipientInfo {
	if envelope == nil || newKey == nil {
		return nil
	}
	recipientKeyIDs := make(map[string]string, len(envelope.RecipientKeyIDs))
	for i, keyID := range envelope.RecipientKeyIDs {
		if len(envelope.RecipientPublicKeys) > i {
			fingerprint := enctypes.ComputeKeyFingerprint(envelope.RecipientPublicKeys[i])
			recipientKeyIDs[fingerprint] = keyID
		}
	}

	oldFingerprint := ""
	if oldKey != nil {
		oldFingerprint = enctypes.ComputeKeyFingerprint(oldKey.PublicKey[:])
	}

	recipients := make([]enccrypto.RecipientInfo, 0, len(envelope.RecipientPublicKeys)+1)
	for _, pubKey := range envelope.RecipientPublicKeys {
		fingerprint := enctypes.ComputeKeyFingerprint(pubKey)
		if oldFingerprint != "" && fingerprint == oldFingerprint {
			continue
		}
		recipients = append(recipients, enccrypto.RecipientInfo{
			PublicKey: pubKey,
			KeyID:     recipientKeyIDs[fingerprint],
		})
	}

	newFingerprint := enctypes.ComputeKeyFingerprint(newKey.PublicKey[:])
	recipients = append(recipients, enccrypto.RecipientInfo{
		PublicKey:  newKey.PublicKey[:],
		KeyID:      enctypes.FormatRecipientKeyID(newFingerprint, newKey.Version),
		KeyVersion: newKey.Version,
	})

	return recipients
}

func safeUint64FromInt64(value int64) uint64 {
	if value <= 0 {
		return 0
	}
	return uint64(value)
}
