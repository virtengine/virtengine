package data_vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	keyMgr  keys.VaultKeyManager
	cipher  BlobCipher
	mu      sync.RWMutex
	initErr error

	// metadata stores blob metadata indexed by blob ID
	metadata map[BlobID]*BlobMetadata
}

// NewEncryptedBlobStore creates a new encrypted blob store
func NewEncryptedBlobStore(backend artifact_store.ArtifactStore, keyMgr keys.VaultKeyManager) *EncryptedBlobStore {
	store, err := NewEncryptedBlobStoreWithError(backend, keyMgr)
	if err != nil {
		return &EncryptedBlobStore{backend: backend, keyMgr: keyMgr, metadata: make(map[BlobID]*BlobMetadata), initErr: err}
	}
	return store
}

type blobMetadataPersistence interface {
	LoadVaultMetadata() (map[BlobID]*BlobMetadata, error)
	SaveVaultMetadata(map[BlobID]*BlobMetadata) error
}

type blobTransactionBackend interface {
	PutVaultBlob(context.Context, *artifact_store.PutRequest, BlobID, func(*artifact_store.PutResponse) *BlobMetadata) (*artifact_store.PutResponse, error)
	DeleteVaultBlob(context.Context, *artifact_store.DeleteRequest, BlobID) error
}

// DurableVaultArtifactStore is the production storage contract. In addition
// to durable encrypted objects it atomically persists vault metadata, so a
// restart cannot orphan a ciphertext or lose its key/version binding.
type DurableVaultArtifactStore interface {
	artifact_store.ArtifactStore
	Durable() bool
	LoadVaultMetadata() (map[BlobID]*BlobMetadata, error)
	SaveVaultMetadata(map[BlobID]*BlobMetadata) error
	PutVaultBlob(context.Context, *artifact_store.PutRequest, BlobID, func(*artifact_store.PutResponse) *BlobMetadata) (*artifact_store.PutResponse, error)
	DeleteVaultBlob(context.Context, *artifact_store.DeleteRequest, BlobID) error
}

// NewEncryptedBlobStoreWithError creates a store and restores durable blob metadata when supported.
func NewEncryptedBlobStoreWithError(backend artifact_store.ArtifactStore, keyMgr keys.VaultKeyManager) (*EncryptedBlobStore, error) {
	return NewEncryptedBlobStoreWithCipher(backend, keyMgr, nil)
}

// NewEncryptedBlobStoreWithCipher creates a store using a non-exportable
// envelope cipher when supplied. The legacy nil cipher path remains for
// fixture compatibility only.
func NewEncryptedBlobStoreWithCipher(backend artifact_store.ArtifactStore, keyMgr keys.VaultKeyManager, blobCipher BlobCipher) (*EncryptedBlobStore, error) {
	if backend == nil || keyMgr == nil {
		return nil, errors.New("artifact backend and key manager are required")
	}
	store := &EncryptedBlobStore{
		backend:  backend,
		keyMgr:   keyMgr,
		cipher:   blobCipher,
		metadata: make(map[BlobID]*BlobMetadata),
	}
	if persistence, ok := backend.(blobMetadataPersistence); ok {
		if _, transactional := backend.(blobTransactionBackend); !transactional {
			return nil, errors.New("durable vault metadata backend requires recoverable blob transactions")
		}
		metadata, err := persistence.LoadVaultMetadata()
		if err != nil {
			return nil, fmt.Errorf("restore vault metadata: %w", err)
		}
		store.metadata = metadata
	}
	return store, nil
}

// Store encrypts and stores a blob in the artifact store
func (s *EncryptedBlobStore) Store(ctx context.Context, req *UploadRequest) (*EncryptedBlob, error) {
	if s.initErr != nil {
		return nil, s.initErr
	}
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

	var envelope *enctypes.EncryptedPayloadEnvelope
	var envelopeBytes []byte
	var artifactMeta *artifact_store.EncryptionMetadata
	if s.cipher != nil {
		envelopeBytes, artifactMeta, envelope, err = s.cipher.Encrypt(ctx, keyInfo, req)
		if err != nil {
			return nil, NewVaultError("Store", ErrEncryptionFailed, fmt.Sprintf("KMS envelope encryption failed: %v", err))
		}
	} else {
		senderKeyPair, keyErr := enccrypto.GenerateKeyPair()
		if keyErr != nil {
			return nil, NewVaultError("Store", ErrEncryptionFailed, fmt.Sprintf("failed to generate sender keypair: %v", keyErr))
		}
		recipients := buildRecipientInfos(keyInfo, req.Recipients)
		if len(recipients) > 1 {
			envelope, err = enccrypto.CreateMultiRecipientEnvelopeWithRecipients(req.Plaintext, recipients, senderKeyPair)
		} else {
			envelope, err = enccrypto.CreateEnvelopeWithRecipient(req.Plaintext, recipients[0], senderKeyPair)
		}
		if err != nil {
			return nil, NewVaultError("Store", ErrEncryptionFailed, fmt.Sprintf("failed to create envelope: %v", err))
		}
		envelopeBytes, err = json.Marshal(envelope)
		if err != nil {
			return nil, NewVaultError("Store", ErrEncryptionFailed, fmt.Sprintf("failed to marshal envelope: %v", err))
		}
		envelopeHash := sha256.Sum256(envelopeBytes)
		artifactMeta = &artifact_store.EncryptionMetadata{AlgorithmID: string(envelope.AlgorithmID), RecipientKeyIDs: envelope.RecipientKeyIDs, EnvelopeHash: envelopeHash[:], SenderKeyID: hex.EncodeToString(envelope.SenderPubKey)}
	}

	// Compute content hash of plaintext
	contentHash := sha256.Sum256(req.Plaintext)

	// Generate blob ID (hash of content + timestamp for uniqueness)
	blobIDBytes := sha256.Sum256(append(contentHash[:], []byte(time.Now().String())...))
	blobID := BlobID(hex.EncodeToString(blobIDBytes[:]))

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

	var metadata *BlobMetadata
	metadataFactory := func(putResp *artifact_store.PutResponse) *BlobMetadata {
		backendRef := ""
		backendName := ""
		if putResp != nil && putResp.ContentAddress != nil {
			backendRef = putResp.ContentAddress.BackendRef
			backendName = string(putResp.ContentAddress.Backend)
		}
		created := &BlobMetadata{
			ID: blobID, Scope: req.Scope, KeyID: keyInfo.ID, KeyVersion: keyInfo.Version,
			ContentHash: contentHash[:], Size: int64(len(req.Plaintext)), EncryptedSize: int64(len(envelopeBytes)),
			Owner: req.Owner, OrgID: req.OrgID, CreatedAt: time.Now().UTC(), ExpiresAt: req.ExpiresAt,
			RetentionPolicy: req.RetentionPolicy, Tags: cloneStringMap(req.Tags), AuditOperationID: req.AuditOperationID,
			Backend: backendName, BackendRef: backendRef,
		}
		if putResp != nil && putResp.ContentAddress != nil {
			created.ContentAddressHash = append([]byte(nil), putResp.ContentAddress.Hash...)
			created.ContentAddressSize = putResp.ContentAddress.Size
			created.ContentAddressAlgorithm = putResp.ContentAddress.Algorithm
			created.ContentAddressVersion = putResp.ContentAddress.Version
		}
		metadata = created
		return cloneBlobMetadataValue(created)
	}

	var putResp *artifact_store.PutResponse
	if transactional, ok := s.backend.(blobTransactionBackend); ok {
		putResp, err = transactional.PutVaultBlob(ctx, putReq, blobID, metadataFactory)
	} else {
		putResp, err = s.backend.Put(ctx, putReq)
		if err == nil {
			metadataFactory(putResp)
		}
	}
	if err != nil {
		if errors.Is(err, ErrReconciliationRequired) && metadata != nil {
			s.mu.Lock()
			s.metadata[blobID] = metadata
			s.mu.Unlock()
			return &EncryptedBlob{Metadata: *cloneBlobMetadataValue(metadata), Envelope: envelope, BackendPath: metadata.BackendRef}, NewVaultError("Store", err, fmt.Sprintf("backend store requires reconciliation: %v", err))
		}
		return nil, NewVaultError("Store", err, fmt.Sprintf("backend store failed: %v", err))
	}

	// Store metadata
	s.mu.Lock()
	s.metadata[blobID] = metadata
	if _, transactional := s.backend.(blobTransactionBackend); !transactional {
		if err := s.persistMetadataLocked(); err != nil {
			delete(s.metadata, blobID)
			s.mu.Unlock()
			return nil, NewVaultError("Store", ErrStorageBackend, fmt.Sprintf("persist metadata: %v", err))
		}
	}
	s.mu.Unlock()

	return &EncryptedBlob{
		Metadata:    *cloneBlobMetadataValue(metadata),
		Envelope:    envelope,
		BackendPath: metadata.BackendRef,
	}, nil
}

// Retrieve retrieves and decrypts a blob from the artifact store
func (s *EncryptedBlobStore) Retrieve(ctx context.Context, blobID BlobID) ([]byte, *BlobMetadata, error) {
	if s.initErr != nil {
		return nil, nil, s.initErr
	}
	if err := s.refreshMetadata(); err != nil {
		return nil, nil, NewVaultError("Retrieve", ErrStorageBackend, fmt.Sprintf("refresh metadata: %v", err))
	}
	// Get metadata
	s.mu.RLock()
	metadata, exists := s.metadata[blobID]
	metadata = cloneBlobMetadataValue(metadata)
	s.mu.RUnlock()

	if !exists {
		return nil, nil, NewVaultError("Retrieve", ErrBlobNotFound, string(blobID))
	}

	// Check expiration
	if metadata.ExpiresAt != nil && time.Now().After(*metadata.ExpiresAt) {
		return nil, nil, NewVaultError("Retrieve", ErrBlobExpired, string(blobID))
	}

	var keyInfo *keys.KeyInfo
	var err error
	if s.cipher == nil {
		keyInfo, err = s.keyMgr.GetKey(keys.Scope(metadata.Scope), metadata.KeyID)
		if err != nil {
			return nil, nil, NewVaultError("Retrieve", ErrInvalidKey, fmt.Sprintf("failed to get key %s: %v", metadata.KeyID, err))
		}
	}

	// Retrieve encrypted envelope from backend
	// Note: We need to reverse lookup by blob ID to content address
	// For now, we'll need to enhance this with a mapping
	// This is a simplified implementation
	contentAddress := s.resolveContentAddress(metadata, blobID)

	getReq := &artifact_store.GetRequest{
		ContentAddress:    contentAddress,
		RequestingAccount: metadata.Owner,
	}

	getResp, err := s.backend.Get(ctx, getReq)
	if err != nil {
		return nil, nil, NewVaultError("Retrieve", ErrStorageBackend, fmt.Sprintf("backend get failed: %v", err))
	}

	if s.cipher != nil {
		plaintext, cipherErr := s.cipher.Decrypt(ctx, metadata, getResp.Data)
		if cipherErr != nil {
			return nil, nil, NewVaultError("Retrieve", ErrDecryptionFailed, fmt.Sprintf("KMS envelope decryption failed: %v", cipherErr))
		}
		computedHash := sha256.Sum256(plaintext)
		if hex.EncodeToString(computedHash[:]) != hex.EncodeToString(metadata.ContentHash) {
			return nil, nil, NewVaultError("Retrieve", ErrDecryptionFailed, "content hash mismatch")
		}
		return plaintext, cloneBlobMetadataValue(metadata), nil
	}

	// Unmarshal legacy envelope
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

	return plaintext, cloneBlobMetadataValue(metadata), nil
}

// GetMetadata retrieves blob metadata without decrypting content
func (s *EncryptedBlobStore) GetMetadata(blobID BlobID) (*BlobMetadata, error) {
	if err := s.refreshMetadata(); err != nil {
		return nil, NewVaultError("GetMetadata", ErrStorageBackend, fmt.Sprintf("refresh metadata: %v", err))
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	metadata, exists := s.metadata[blobID]
	if !exists {
		return nil, NewVaultError("GetMetadata", ErrBlobNotFound, string(blobID))
	}

	// Return a copy to prevent mutation
	return cloneBlobMetadataValue(metadata), nil
}

// Delete marks a blob for deletion
func (s *EncryptedBlobStore) Delete(ctx context.Context, blobID BlobID) error {
	if err := s.refreshMetadata(); err != nil {
		return NewVaultError("Delete", ErrStorageBackend, fmt.Sprintf("refresh metadata: %v", err))
	}
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

	var err error
	if transactional, ok := s.backend.(blobTransactionBackend); ok {
		err = transactional.DeleteVaultBlob(ctx, deleteReq, blobID)
	} else {
		err = s.backend.Delete(ctx, deleteReq)
	}
	if err != nil {
		if errors.Is(err, ErrReconciliationRequired) {
			s.mu.Lock()
			delete(s.metadata, blobID)
			s.mu.Unlock()
		}
		return NewVaultError("Delete", err, fmt.Sprintf("backend delete failed: %v", err))
	}

	// Remove metadata
	s.mu.Lock()
	delete(s.metadata, blobID)
	if _, transactional := s.backend.(blobTransactionBackend); !transactional {
		if err := s.persistMetadataLocked(); err != nil {
			s.metadata[blobID] = metadata
			s.mu.Unlock()
			return NewVaultError("Delete", ErrStorageBackend, fmt.Sprintf("persist metadata: %v", err))
		}
	}
	s.mu.Unlock()

	return nil
}

// ListByScope lists all blobs in a scope
func (s *EncryptedBlobStore) ListByScope(scope Scope) ([]*BlobMetadata, error) {
	if err := s.refreshMetadata(); err != nil {
		return nil, NewVaultError("ListByScope", ErrStorageBackend, fmt.Sprintf("refresh metadata: %v", err))
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*BlobMetadata
	for _, metadata := range s.metadata {
		if metadata.Scope == scope {
			result = append(result, cloneBlobMetadataValue(metadata))
		}
	}

	return result, nil
}

func (s *EncryptedBlobStore) refreshMetadata() error {
	persistence, ok := s.backend.(blobMetadataPersistence)
	if !ok {
		return nil
	}
	metadata, err := persistence.LoadVaultMetadata()
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.metadata = metadata
	s.mu.Unlock()
	return nil
}

// Close closes the encrypted blob store
func (s *EncryptedBlobStore) Close() error {
	var closeErrors []error
	if closer, ok := s.backend.(interface{ Close() error }); ok {
		closeErrors = append(closeErrors, closer.Close())
	}
	if closer, ok := s.keyMgr.(interface{ Close() error }); ok {
		closeErrors = append(closeErrors, closer.Close())
	}
	return errors.Join(closeErrors...)
}

// KeyManager returns the key manager used by the store.
func (s *EncryptedBlobStore) KeyManager() keys.VaultKeyManager {
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
	if s.cipher != nil {
		return s.reencryptWithCipher(ctx, blobID, metadata, plaintext, newKey)
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
	candidate := cloneBlobMetadata(s.metadata)
	candidate[blobID] = cloneBlobMetadataValue(metadata)
	if err := s.persistMetadataValueLocked(candidate); err != nil {
		s.mu.Unlock()
		return nil, NewVaultError("Reencrypt", ErrStorageBackend, fmt.Sprintf("persist metadata: %v", err))
	}
	s.metadata = candidate
	s.mu.Unlock()

	return &EncryptedBlob{
		Metadata:    *cloneBlobMetadataValue(metadata),
		Envelope:    newEnvelope,
		BackendPath: backendRef,
	}, nil
}

func (s *EncryptedBlobStore) reencryptWithCipher(ctx context.Context, blobID BlobID, metadata *BlobMetadata, plaintext []byte, newKey *keys.KeyInfo) (*EncryptedBlob, error) {
	request := &UploadRequest{Scope: metadata.Scope, Plaintext: plaintext, Owner: metadata.Owner, OrgID: metadata.OrgID, RetentionPolicy: metadata.RetentionPolicy, ExpiresAt: metadata.ExpiresAt, Tags: metadata.Tags, AuditOperationID: metadata.AuditOperationID}
	encoded, encryptionMeta, _, err := s.cipher.Encrypt(ctx, newKey, request)
	if err != nil {
		return nil, NewVaultError("Reencrypt", ErrEncryptionFailed, fmt.Sprintf("KMS envelope encryption failed: %v", err))
	}
	tag := &artifact_store.RetentionTag{PolicyID: metadata.RetentionPolicy, Owner: metadata.Owner}
	if metadata.ExpiresAt != nil {
		tag.ExpiresAt = metadata.ExpiresAt
	}
	response, err := s.backend.Put(ctx, &artifact_store.PutRequest{Data: encoded, ContentHash: metadata.ContentHash, EncryptionMetadata: encryptionMeta, RetentionTag: tag, Owner: metadata.Owner, ArtifactType: string(metadata.Scope), Metadata: metadata.Tags})
	if err != nil {
		return nil, NewVaultError("Reencrypt", ErrStorageBackend, fmt.Sprintf("backend store failed: %v", err))
	}
	metadata.KeyID, metadata.KeyVersion, metadata.EncryptedSize = newKey.ID, newKey.Version, int64(len(encoded))
	if response.ContentAddress != nil {
		metadata.Backend, metadata.BackendRef = string(response.ContentAddress.Backend), response.ContentAddress.BackendRef
		metadata.ContentAddressHash = append([]byte(nil), response.ContentAddress.Hash...)
		metadata.ContentAddressSize, metadata.ContentAddressAlgorithm, metadata.ContentAddressVersion = response.ContentAddress.Size, response.ContentAddress.Algorithm, response.ContentAddress.Version
	}
	s.mu.Lock()
	candidate := cloneBlobMetadata(s.metadata)
	candidate[blobID] = cloneBlobMetadataValue(metadata)
	if err := s.persistMetadataValueLocked(candidate); err != nil {
		s.mu.Unlock()
		return nil, NewVaultError("Reencrypt", ErrStorageBackend, fmt.Sprintf("persist metadata: %v", err))
	}
	s.metadata = candidate
	s.mu.Unlock()
	return &EncryptedBlob{Metadata: *cloneBlobMetadataValue(metadata), BackendPath: metadata.BackendRef}, nil
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
	getResp, err := s.backend.Get(ctx, &artifact_store.GetRequest{ContentAddress: contentAddress, RequestingAccount: metadata.Owner})
	if err != nil {
		return nil, NewVaultError("Retrieve", ErrStorageBackend, fmt.Sprintf("backend get failed: %v", err))
	}
	var envelope enctypes.EncryptedPayloadEnvelope
	if err := json.Unmarshal(getResp.Data, &envelope); err != nil {
		return nil, NewVaultError("Retrieve", ErrDecryptionFailed, fmt.Sprintf("failed to unmarshal envelope: %v", err))
	}
	return &envelope, nil
}

func (s *EncryptedBlobStore) persistMetadataLocked() error {
	return s.persistMetadataValueLocked(s.metadata)
}

func (s *EncryptedBlobStore) persistMetadataValueLocked(metadata map[BlobID]*BlobMetadata) error {
	if persistence, ok := s.backend.(blobMetadataPersistence); ok {
		return persistence.SaveVaultMetadata(metadata)
	}
	return nil
}

func cloneBlobMetadataValue(metadata *BlobMetadata) *BlobMetadata {
	if metadata == nil {
		return nil
	}
	return cloneBlobMetadata(map[BlobID]*BlobMetadata{metadata.ID: metadata})[metadata.ID]
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
