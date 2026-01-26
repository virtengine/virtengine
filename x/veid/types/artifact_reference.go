package types

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// ============================================================================
// On-Chain Artifact Reference Types (VE-218)
// ============================================================================
//
// These types represent on-chain references to off-chain encrypted artifacts.
// Raw biometric images and sensitive data are NEVER stored on-chain.
// Only content-addressed hashes and metadata are stored in the blockchain state.

// ArtifactReferenceVersion is the current format version
const ArtifactReferenceVersion uint32 = 1

// ChunkManifestVersion is the current chunk manifest format version
const ChunkManifestVersion uint32 = 1

// StorageBackend represents the type of off-chain storage backend
type StorageBackend string

const (
	// StorageBackendWaldur represents Waldur DB/object store
	StorageBackendWaldur StorageBackend = "waldur"

	// StorageBackendIPFS represents IPFS distributed storage
	StorageBackendIPFS StorageBackend = "ipfs"
)

// AllStorageBackends returns all valid storage backends
func AllStorageBackends() []StorageBackend {
	return []StorageBackend{
		StorageBackendWaldur,
		StorageBackendIPFS,
	}
}

// IsValidStorageBackend checks if a storage backend is valid
func IsValidStorageBackend(b StorageBackend) bool {
	for _, valid := range AllStorageBackends() {
		if b == valid {
			return true
		}
	}
	return false
}

// ContentAddressReference is the on-chain pointer to off-chain encrypted data
// This is what gets stored in the blockchain state
type ContentAddressReference struct {
	// Version is the format version
	Version uint32 `json:"version"`

	// Hash is the SHA-256 hash of the original unencrypted artifact
	// Used for content verification after decryption
	// Stored as 32 bytes (256 bits)
	Hash []byte `json:"hash"`

	// EncryptedHash is the SHA-256 hash of the encrypted artifact
	// Used for integrity verification during retrieval
	EncryptedHash []byte `json:"encrypted_hash"`

	// Algorithm identifies the hash algorithm (always sha256)
	Algorithm string `json:"algorithm"`

	// Size is the original unencrypted size in bytes
	Size uint64 `json:"size"`

	// EncryptedSize is the encrypted artifact size in bytes
	EncryptedSize uint64 `json:"encrypted_size"`

	// Backend indicates which backend stores this artifact
	Backend StorageBackend `json:"backend"`

	// BackendRef is the backend-specific reference
	// For Waldur: object ID or path
	// For IPFS: CID string
	BackendRef string `json:"backend_ref"`
}

// NewContentAddressReference creates a new content address reference
func NewContentAddressReference(
	originalHash []byte,
	encryptedHash []byte,
	size uint64,
	encryptedSize uint64,
	backend StorageBackend,
	backendRef string,
) *ContentAddressReference {
	return &ContentAddressReference{
		Version:       ArtifactReferenceVersion,
		Hash:          originalHash,
		EncryptedHash: encryptedHash,
		Algorithm:     "sha256",
		Size:          size,
		EncryptedSize: encryptedSize,
		Backend:       backend,
		BackendRef:    backendRef,
	}
}

// HashHex returns the original content hash as a hexadecimal string
func (c *ContentAddressReference) HashHex() string {
	return hex.EncodeToString(c.Hash)
}

// EncryptedHashHex returns the encrypted content hash as a hexadecimal string
func (c *ContentAddressReference) EncryptedHashHex() string {
	return hex.EncodeToString(c.EncryptedHash)
}

// Validate validates the content address reference
func (c *ContentAddressReference) Validate() error {
	if c.Version == 0 {
		return ErrInvalidPayload.Wrap("version cannot be zero")
	}
	if c.Version > ArtifactReferenceVersion {
		return ErrInvalidPayload.Wrapf("unsupported version %d (max: %d)", c.Version, ArtifactReferenceVersion)
	}
	if len(c.Hash) != 32 {
		return ErrInvalidPayload.Wrapf("invalid hash length: got %d, want 32", len(c.Hash))
	}
	if len(c.EncryptedHash) != 32 {
		return ErrInvalidPayload.Wrapf("invalid encrypted_hash length: got %d, want 32", len(c.EncryptedHash))
	}
	if c.Algorithm != "sha256" {
		return ErrInvalidPayload.Wrapf("unsupported algorithm: %s", c.Algorithm)
	}
	if c.Size == 0 {
		return ErrInvalidPayload.Wrap("size cannot be zero")
	}
	if c.EncryptedSize == 0 {
		return ErrInvalidPayload.Wrap("encrypted_size cannot be zero")
	}
	if !IsValidStorageBackend(c.Backend) {
		return ErrInvalidPayload.Wrapf("invalid backend: %s", c.Backend)
	}
	if c.BackendRef == "" {
		return ErrInvalidPayload.Wrap("backend_ref cannot be empty")
	}
	return nil
}

// ChunkReference represents a single chunk in a chunked artifact
type ChunkReference struct {
	// Index is the chunk's position in the sequence (0-based)
	Index uint32 `json:"index"`

	// Hash is the SHA-256 hash of the encrypted chunk content
	Hash []byte `json:"hash"`

	// Size is the chunk size in bytes
	Size uint64 `json:"size"`

	// Offset is the byte offset in the original artifact
	Offset uint64 `json:"offset"`

	// BackendRef is the chunk-specific backend reference (e.g., IPFS CID)
	BackendRef string `json:"backend_ref"`
}

// HashHex returns the chunk hash as a hexadecimal string
func (c *ChunkReference) HashHex() string {
	return hex.EncodeToString(c.Hash)
}

// ChunkManifestReference describes how an artifact is split into chunks
// Stored on-chain for deterministic reconstruction of fragmented artifacts
type ChunkManifestReference struct {
	// Version is the manifest format version
	Version uint32 `json:"version"`

	// TotalSize is the total encrypted artifact size in bytes
	TotalSize uint64 `json:"total_size"`

	// ChunkSize is the standard chunk size (last chunk may be smaller)
	ChunkSize uint64 `json:"chunk_size"`

	// ChunkCount is the total number of chunks
	ChunkCount uint32 `json:"chunk_count"`

	// Chunks contains ordered chunk references
	// Ordering is deterministic based on chunk index
	Chunks []ChunkReference `json:"chunks"`

	// RootHash is the Merkle root of all chunk hashes
	// Used for integrity verification
	RootHash []byte `json:"root_hash"`

	// Algorithm is the hash algorithm used
	Algorithm string `json:"algorithm"`
}

// NewChunkManifestReference creates a new chunk manifest reference
func NewChunkManifestReference(totalSize, chunkSize uint64) *ChunkManifestReference {
	chunkCount := uint32((totalSize + chunkSize - 1) / chunkSize)
	return &ChunkManifestReference{
		Version:    ChunkManifestVersion,
		TotalSize:  totalSize,
		ChunkSize:  chunkSize,
		ChunkCount: chunkCount,
		Chunks:     make([]ChunkReference, 0, chunkCount),
		Algorithm:  "sha256",
	}
}

// AddChunk adds a chunk to the manifest
func (m *ChunkManifestReference) AddChunk(chunk ChunkReference) error {
	if uint32(len(m.Chunks)) >= m.ChunkCount {
		return ErrInvalidPayload.Wrap("cannot add more chunks than declared count")
	}
	if chunk.Index != uint32(len(m.Chunks)) {
		return ErrInvalidPayload.Wrapf("expected chunk index %d, got %d", len(m.Chunks), chunk.Index)
	}
	m.Chunks = append(m.Chunks, chunk)
	return nil
}

// ComputeRootHash computes the Merkle root hash of all chunks
func (m *ChunkManifestReference) ComputeRootHash() []byte {
	if len(m.Chunks) == 0 {
		return nil
	}

	// Simple Merkle tree: hash all chunk hashes together
	hashes := make([]byte, 0, len(m.Chunks)*32)
	for _, chunk := range m.Chunks {
		hashes = append(hashes, chunk.Hash...)
	}
	root := sha256.Sum256(hashes)
	m.RootHash = root[:]
	return m.RootHash
}

// RootHashHex returns the root hash as a hexadecimal string
func (m *ChunkManifestReference) RootHashHex() string {
	return hex.EncodeToString(m.RootHash)
}

// Validate validates the chunk manifest reference
func (m *ChunkManifestReference) Validate() error {
	if m.Version == 0 {
		return ErrInvalidPayload.Wrap("version cannot be zero")
	}
	if m.Version > ChunkManifestVersion {
		return ErrInvalidPayload.Wrapf("unsupported version %d (max: %d)", m.Version, ChunkManifestVersion)
	}
	if m.TotalSize == 0 {
		return ErrInvalidPayload.Wrap("total_size cannot be zero")
	}
	if m.ChunkSize == 0 {
		return ErrInvalidPayload.Wrap("chunk_size cannot be zero")
	}
	if m.ChunkCount == 0 {
		return ErrInvalidPayload.Wrap("chunk_count cannot be zero")
	}
	if uint32(len(m.Chunks)) != m.ChunkCount {
		return ErrInvalidPayload.Wrapf("chunk count mismatch: got %d, want %d", len(m.Chunks), m.ChunkCount)
	}
	if len(m.RootHash) != 32 {
		return ErrInvalidPayload.Wrapf("invalid root_hash length: got %d, want 32", len(m.RootHash))
	}

	// Validate individual chunks
	var totalChunkSize uint64
	for i, chunk := range m.Chunks {
		if chunk.Index != uint32(i) {
			return ErrInvalidPayload.Wrapf("chunk %d has wrong index %d", i, chunk.Index)
		}
		if len(chunk.Hash) != 32 {
			return ErrInvalidPayload.Wrapf("chunk %d has invalid hash length: %d", i, len(chunk.Hash))
		}
		totalChunkSize += chunk.Size
	}

	if totalChunkSize != m.TotalSize {
		return ErrInvalidPayload.Wrapf("chunk sizes don't sum to total: got %d, want %d", totalChunkSize, m.TotalSize)
	}

	return nil
}

// EncryptionEnvelopeMetadata contains on-chain encryption metadata
// The actual encrypted payload is stored off-chain
type EncryptionEnvelopeMetadata struct {
	// AlgorithmID identifies the encryption algorithm used
	AlgorithmID string `json:"algorithm_id"`

	// RecipientKeyIDs are the fingerprints of intended recipients' public keys
	RecipientKeyIDs []string `json:"recipient_key_ids"`

	// EnvelopeHash is the SHA-256 hash of the full encryption envelope
	EnvelopeHash []byte `json:"envelope_hash"`

	// SenderKeyID is the fingerprint of the sender's public key
	SenderKeyID string `json:"sender_key_id"`
}

// NewEncryptionEnvelopeMetadata creates new encryption envelope metadata
func NewEncryptionEnvelopeMetadata(
	algorithmID string,
	recipientKeyIDs []string,
	envelopeHash []byte,
	senderKeyID string,
) *EncryptionEnvelopeMetadata {
	return &EncryptionEnvelopeMetadata{
		AlgorithmID:     algorithmID,
		RecipientKeyIDs: recipientKeyIDs,
		EnvelopeHash:    envelopeHash,
		SenderKeyID:     senderKeyID,
	}
}

// Validate validates the encryption envelope metadata
func (e *EncryptionEnvelopeMetadata) Validate() error {
	if e.AlgorithmID == "" {
		return ErrInvalidPayload.Wrap("algorithm_id cannot be empty")
	}
	if len(e.RecipientKeyIDs) == 0 {
		return ErrInvalidPayload.Wrap("recipient_key_ids cannot be empty")
	}
	if len(e.EnvelopeHash) != 32 {
		return ErrInvalidPayload.Wrapf("invalid envelope_hash length: got %d, want 32", len(e.EnvelopeHash))
	}
	return nil
}

// IdentityArtifactReference is the on-chain reference to an off-chain encrypted identity artifact
// This is stored in the blockchain state by the veid module keeper
type IdentityArtifactReference struct {
	// Version is the format version
	Version uint32 `json:"version"`

	// ReferenceID is a unique identifier for this reference
	ReferenceID string `json:"reference_id"`

	// AccountAddress is the account this artifact belongs to
	AccountAddress string `json:"account_address"`

	// ArtifactType identifies what kind of artifact this is
	ArtifactType ArtifactType `json:"artifact_type"`

	// ContentAddress points to the artifact data
	ContentAddress *ContentAddressReference `json:"content_address"`

	// ChunkManifest describes chunking (if applicable, for IPFS)
	ChunkManifest *ChunkManifestReference `json:"chunk_manifest,omitempty"`

	// EncryptionMetadata contains encryption information
	EncryptionMetadata *EncryptionEnvelopeMetadata `json:"encryption_metadata"`

	// RetentionPolicyID references the retention policy
	RetentionPolicyID string `json:"retention_policy_id,omitempty"`

	// CreatedAt is when this reference was created (block time)
	CreatedAt time.Time `json:"created_at"`

	// CreatedAtBlock is the block height when created
	CreatedAtBlock int64 `json:"created_at_block"`

	// SourceScopeID references the scope this artifact was derived from
	SourceScopeID string `json:"source_scope_id,omitempty"`

	// Revoked indicates if this artifact has been revoked
	Revoked bool `json:"revoked"`

	// RevokedAt is when the artifact was revoked
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// RevokedReason is the reason for revocation
	RevokedReason string `json:"revoked_reason,omitempty"`

	// Metadata contains optional additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewIdentityArtifactReference creates a new identity artifact reference
func NewIdentityArtifactReference(
	referenceID string,
	accountAddress string,
	artifactType ArtifactType,
	contentAddress *ContentAddressReference,
	encryptionMeta *EncryptionEnvelopeMetadata,
	createdAt time.Time,
	blockHeight int64,
) *IdentityArtifactReference {
	return &IdentityArtifactReference{
		Version:            ArtifactReferenceVersion,
		ReferenceID:        referenceID,
		AccountAddress:     accountAddress,
		ArtifactType:       artifactType,
		ContentAddress:     contentAddress,
		EncryptionMetadata: encryptionMeta,
		CreatedAt:          createdAt,
		CreatedAtBlock:     blockHeight,
		Revoked:            false,
		Metadata:           make(map[string]string),
	}
}

// Validate validates the identity artifact reference
func (a *IdentityArtifactReference) Validate() error {
	if a.Version == 0 {
		return ErrInvalidPayload.Wrap("version cannot be zero")
	}
	if a.Version > ArtifactReferenceVersion {
		return ErrInvalidPayload.Wrapf("unsupported version %d (max: %d)", a.Version, ArtifactReferenceVersion)
	}
	if a.ReferenceID == "" {
		return ErrInvalidPayload.Wrap("reference_id cannot be empty")
	}
	if a.AccountAddress == "" {
		return ErrInvalidPayload.Wrap("account_address cannot be empty")
	}
	if !IsValidArtifactType(a.ArtifactType) {
		return ErrInvalidPayload.Wrapf("invalid artifact_type: %s", a.ArtifactType)
	}
	if a.ContentAddress == nil {
		return ErrInvalidPayload.Wrap("content_address cannot be nil")
	}
	if err := a.ContentAddress.Validate(); err != nil {
		return ErrInvalidPayload.Wrapf("invalid content_address: %v", err)
	}
	if a.EncryptionMetadata == nil {
		return ErrInvalidPayload.Wrap("encryption_metadata cannot be nil")
	}
	if err := a.EncryptionMetadata.Validate(); err != nil {
		return ErrInvalidPayload.Wrapf("invalid encryption_metadata: %v", err)
	}

	// Validate chunk manifest if present
	if a.ChunkManifest != nil {
		if err := a.ChunkManifest.Validate(); err != nil {
			return ErrInvalidPayload.Wrapf("invalid chunk_manifest: %v", err)
		}
	}

	return nil
}

// SetChunkManifest sets the chunk manifest for chunked storage
func (a *IdentityArtifactReference) SetChunkManifest(manifest *ChunkManifestReference) {
	a.ChunkManifest = manifest
}

// SetRetentionPolicy sets the retention policy ID
func (a *IdentityArtifactReference) SetRetentionPolicy(policyID string) {
	a.RetentionPolicyID = policyID
}

// SetMetadata sets a metadata key-value pair
func (a *IdentityArtifactReference) SetMetadata(key, value string) {
	if a.Metadata == nil {
		a.Metadata = make(map[string]string)
	}
	a.Metadata[key] = value
}

// GetMetadata gets a metadata value by key
func (a *IdentityArtifactReference) GetMetadata(key string) (string, bool) {
	if a.Metadata == nil {
		return "", false
	}
	val, ok := a.Metadata[key]
	return val, ok
}

// IsChunked returns true if the artifact is stored in chunks
func (a *IdentityArtifactReference) IsChunked() bool {
	return a.ChunkManifest != nil && a.ChunkManifest.ChunkCount > 1
}

// Revoke marks the artifact as revoked
func (a *IdentityArtifactReference) Revoke(reason string, revokedAt time.Time) {
	a.Revoked = true
	a.RevokedAt = &revokedAt
	a.RevokedReason = reason
}

// IsRevoked returns true if the artifact is revoked
func (a *IdentityArtifactReference) IsRevoked() bool {
	return a.Revoked
}
