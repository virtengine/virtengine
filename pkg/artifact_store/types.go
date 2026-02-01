package artifact_store

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Version constants
const (
	// ContentAddressVersion is the current content address format version
	ContentAddressVersion uint32 = 1

	// ChunkManifestVersion is the current chunk manifest format version
	ChunkManifestVersion uint32 = 1

	// ArtifactReferenceVersion is the current artifact reference format version
	ArtifactReferenceVersion uint32 = 1
)

// Backend type constants
const (
	// BackendWaldur represents the Waldur backend storage
	BackendWaldur BackendType = "waldur"

	// BackendIPFS represents the IPFS backend storage
	BackendIPFS BackendType = "ipfs"
)

// BackendType represents the type of storage backend
type BackendType string

// String returns the string representation of the backend type
func (b BackendType) String() string {
	return string(b)
}

// IsValid checks if the backend type is valid
func (b BackendType) IsValid() bool {
	switch b {
	case BackendWaldur, BackendIPFS:
		return true
	default:
		return false
	}
}

// ContentAddress represents a content-addressed reference to an artifact
// This is the on-chain pointer to off-chain encrypted data
type ContentAddress struct {
	// Version is the format version
	Version uint32 `json:"version"`

	// Hash is the SHA-256 hash of the artifact content
	// This is 32 bytes (256 bits)
	Hash []byte `json:"hash"`

	// Algorithm identifies the hash algorithm used (always sha256 for now)
	Algorithm string `json:"algorithm"`

	// Size is the original unencrypted size in bytes
	Size uint64 `json:"size"`

	// Backend indicates which backend stores this artifact
	Backend BackendType `json:"backend"`

	// BackendRef is the backend-specific reference
	// For Waldur: object ID or path
	// For IPFS: CID string
	BackendRef string `json:"backend_ref"`
}

// NewContentAddress creates a new content address from raw content
func NewContentAddress(content []byte, backend BackendType, backendRef string) *ContentAddress {
	hash := sha256.Sum256(content)
	return &ContentAddress{
		Version:    ContentAddressVersion,
		Hash:       hash[:],
		Algorithm:  "sha256",
		Size:       uint64(len(content)),
		Backend:    backend,
		BackendRef: backendRef,
	}
}

// NewContentAddressFromHash creates a new content address from a pre-computed hash
func NewContentAddressFromHash(hash []byte, size uint64, backend BackendType, backendRef string) *ContentAddress {
	return &ContentAddress{
		Version:    ContentAddressVersion,
		Hash:       hash,
		Algorithm:  "sha256",
		Size:       size,
		Backend:    backend,
		BackendRef: backendRef,
	}
}

// HashHex returns the hash as a hexadecimal string
func (c *ContentAddress) HashHex() string {
	return hex.EncodeToString(c.Hash)
}

// Validate validates the content address
func (c *ContentAddress) Validate() error {
	if c.Version == 0 {
		return ErrInvalidContentAddress.Wrap("version cannot be zero")
	}
	if c.Version > ContentAddressVersion {
		return ErrInvalidContentAddress.Wrapf("unsupported version %d (max: %d)", c.Version, ContentAddressVersion)
	}
	if len(c.Hash) != 32 {
		return ErrInvalidContentAddress.Wrapf("invalid hash length: got %d, want 32", len(c.Hash))
	}
	if c.Algorithm != "sha256" {
		return ErrInvalidContentAddress.Wrapf("unsupported algorithm: %s", c.Algorithm)
	}
	if !c.Backend.IsValid() {
		return ErrInvalidContentAddress.Wrapf("invalid backend: %s", c.Backend)
	}
	if c.BackendRef == "" {
		return ErrInvalidContentAddress.Wrap("backend_ref cannot be empty")
	}
	return nil
}

// Equals checks if two content addresses are equal
func (c *ContentAddress) Equals(other *ContentAddress) bool {
	if c == nil || other == nil {
		return c == other
	}
	if len(c.Hash) != len(other.Hash) {
		return false
	}
	for i := range c.Hash {
		if c.Hash[i] != other.Hash[i] {
			return false
		}
	}
	return c.Algorithm == other.Algorithm &&
		c.Backend == other.Backend &&
		c.BackendRef == other.BackendRef
}

// ChunkInfo represents information about a single chunk
type ChunkInfo struct {
	// Index is the chunk's position in the sequence (0-based)
	Index uint32 `json:"index"`

	// Hash is the SHA-256 hash of the chunk content
	Hash []byte `json:"hash"`

	// Size is the chunk size in bytes
	Size uint64 `json:"size"`

	// Offset is the byte offset in the original artifact
	Offset uint64 `json:"offset"`

	// BackendRef is the chunk-specific backend reference (e.g., IPFS CID)
	BackendRef string `json:"backend_ref"`
}

// HashHex returns the chunk hash as a hexadecimal string
func (c *ChunkInfo) HashHex() string {
	return hex.EncodeToString(c.Hash)
}

// ChunkManifest describes how an artifact is split into chunks for storage
// This enables deterministic reconstruction of fragmented artifacts
type ChunkManifest struct {
	// Version is the manifest format version
	Version uint32 `json:"version"`

	// TotalSize is the total artifact size in bytes
	TotalSize uint64 `json:"total_size"`

	// ChunkSize is the standard chunk size (last chunk may be smaller)
	ChunkSize uint64 `json:"chunk_size"`

	// ChunkCount is the total number of chunks
	ChunkCount uint32 `json:"chunk_count"`

	// Chunks contains ordered chunk information
	// Ordering is deterministic based on chunk index
	Chunks []ChunkInfo `json:"chunks"`

	// RootHash is the Merkle root of all chunk hashes
	// Used for integrity verification
	RootHash []byte `json:"root_hash"`

	// Algorithm is the hash algorithm used
	Algorithm string `json:"algorithm"`
}

// NewChunkManifest creates a new chunk manifest
func NewChunkManifest(totalSize, chunkSize uint64) *ChunkManifest {
	//nolint:gosec // G115: chunkCount is bounded by reasonable file sizes
	chunkCount := uint32((totalSize + chunkSize - 1) / chunkSize)
	return &ChunkManifest{
		Version:    ChunkManifestVersion,
		TotalSize:  totalSize,
		ChunkSize:  chunkSize,
		ChunkCount: chunkCount,
		Chunks:     make([]ChunkInfo, 0, chunkCount),
		Algorithm:  "sha256",
	}
}

// AddChunk adds a chunk to the manifest
func (m *ChunkManifest) AddChunk(chunk ChunkInfo) error {
	//nolint:gosec // G115: len(m.Chunks) bounded by m.ChunkCount
	if uint32(len(m.Chunks)) >= m.ChunkCount {
		return ErrInvalidChunkManifest.Wrap("cannot add more chunks than declared count")
	}
	//nolint:gosec // G115: len(m.Chunks) bounded by m.ChunkCount
	if chunk.Index != uint32(len(m.Chunks)) {
		return ErrInvalidChunkManifest.Wrapf("expected chunk index %d, got %d", len(m.Chunks), chunk.Index)
	}
	m.Chunks = append(m.Chunks, chunk)
	return nil
}

// ComputeRootHash computes the Merkle root hash of all chunks
func (m *ChunkManifest) ComputeRootHash() []byte {
	if len(m.Chunks) == 0 {
		return nil
	}

	// Simple Merkle tree: hash all chunk hashes together
	// For a more sophisticated implementation, use a proper Merkle tree
	hashes := make([]byte, 0, len(m.Chunks)*32)
	for _, chunk := range m.Chunks {
		hashes = append(hashes, chunk.Hash...)
	}
	root := sha256.Sum256(hashes)
	m.RootHash = root[:]
	return m.RootHash
}

// RootHashHex returns the root hash as a hexadecimal string
func (m *ChunkManifest) RootHashHex() string {
	return hex.EncodeToString(m.RootHash)
}

// Serialize serializes the manifest to bytes for storage
// Uses a simple binary format: version + totalSize + chunkSize + chunkCount + rootHash + chunks
func (m *ChunkManifest) Serialize() []byte {
	// Estimate size: header (4+8+8+4+32) + chunks (4+32+8+8+variable string per chunk)
	buf := make([]byte, 0, 56+len(m.Chunks)*100)

	// Write header
	buf = appendUint32(buf, m.Version)
	buf = appendUint64(buf, m.TotalSize)
	buf = appendUint64(buf, m.ChunkSize)
	buf = appendUint32(buf, m.ChunkCount)
	buf = append(buf, m.RootHash...)

	// Write chunks
	for _, chunk := range m.Chunks {
		buf = appendUint32(buf, chunk.Index)
		buf = append(buf, chunk.Hash...)
		buf = appendUint64(buf, chunk.Size)
		buf = appendUint64(buf, chunk.Offset)
		buf = appendUint32(buf, uint32(len(chunk.BackendRef)))
		buf = append(buf, []byte(chunk.BackendRef)...)
	}

	return buf
}

// appendUint32 appends a uint32 to a byte slice in big-endian format
func appendUint32(buf []byte, v uint32) []byte {
	return append(buf, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// appendUint64 appends a uint64 to a byte slice in big-endian format
func appendUint64(buf []byte, v uint64) []byte {
	return append(buf, byte(v>>56), byte(v>>48), byte(v>>40), byte(v>>32),
		byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

// Validate validates the chunk manifest
func (m *ChunkManifest) Validate() error {
	if m.Version == 0 {
		return ErrInvalidChunkManifest.Wrap("version cannot be zero")
	}
	if m.Version > ChunkManifestVersion {
		return ErrInvalidChunkManifest.Wrapf("unsupported version %d (max: %d)", m.Version, ChunkManifestVersion)
	}
	if m.TotalSize == 0 {
		return ErrInvalidChunkManifest.Wrap("total_size cannot be zero")
	}
	if m.ChunkSize == 0 {
		return ErrInvalidChunkManifest.Wrap("chunk_size cannot be zero")
	}
	if m.ChunkCount == 0 {
		return ErrInvalidChunkManifest.Wrap("chunk_count cannot be zero")
	}
	//nolint:gosec // G115: len(m.Chunks) bounded by m.ChunkCount
	if uint32(len(m.Chunks)) != m.ChunkCount {
		return ErrInvalidChunkManifest.Wrapf("chunk count mismatch: got %d, want %d", len(m.Chunks), m.ChunkCount)
	}
	if len(m.RootHash) != 32 {
		return ErrInvalidChunkManifest.Wrapf("invalid root_hash length: got %d, want 32", len(m.RootHash))
	}

	// Validate individual chunks
	var totalChunkSize uint64
	for i, chunk := range m.Chunks {
		//nolint:gosec // G115: i bounded by m.ChunkCount
		if chunk.Index != uint32(i) {
			return ErrInvalidChunkManifest.Wrapf("chunk %d has wrong index %d", i, chunk.Index)
		}
		if len(chunk.Hash) != 32 {
			return ErrInvalidChunkManifest.Wrapf("chunk %d has invalid hash length: %d", i, len(chunk.Hash))
		}
		totalChunkSize += chunk.Size
	}

	if totalChunkSize != m.TotalSize {
		return ErrInvalidChunkManifest.Wrapf("chunk sizes don't sum to total: got %d, want %d", totalChunkSize, m.TotalSize)
	}

	return nil
}

// RetentionTag represents retention metadata for an artifact
type RetentionTag struct {
	// PolicyID references the retention policy
	PolicyID string `json:"policy_id"`

	// ExpiresAt is when the artifact should be deleted
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// ExpiresAtBlock is the block height for expiration
	ExpiresAtBlock *int64 `json:"expires_at_block,omitempty"`

	// DeleteOnExpiry indicates if the artifact should be auto-deleted
	DeleteOnExpiry bool `json:"delete_on_expiry"`

	// Owner is the account address that owns this artifact
	Owner string `json:"owner"`

	// CreatedAt is when the retention tag was created
	CreatedAt time.Time `json:"created_at"`
}

// NewRetentionTag creates a new retention tag
func NewRetentionTag(policyID, owner string, deleteOnExpiry bool) *RetentionTag {
	return &RetentionTag{
		PolicyID:       policyID,
		DeleteOnExpiry: deleteOnExpiry,
		Owner:          owner,
		CreatedAt:      time.Now().UTC(),
	}
}

// SetExpiration sets the expiration time
func (r *RetentionTag) SetExpiration(expiresAt time.Time) {
	r.ExpiresAt = &expiresAt
}

// SetExpirationBlock sets the expiration block height
func (r *RetentionTag) SetExpirationBlock(blockHeight int64) {
	r.ExpiresAtBlock = &blockHeight
}

// IsExpired checks if the retention has expired based on current time
func (r *RetentionTag) IsExpired(now time.Time) bool {
	if r.ExpiresAt == nil {
		return false
	}
	return now.After(*r.ExpiresAt)
}

// IsExpiredAtBlock checks if the retention has expired based on block height
func (r *RetentionTag) IsExpiredAtBlock(blockHeight int64) bool {
	if r.ExpiresAtBlock == nil {
		return false
	}
	return blockHeight >= *r.ExpiresAtBlock
}

// EncryptionMetadata contains encryption-related metadata for an artifact
type EncryptionMetadata struct {
	// AlgorithmID identifies the encryption algorithm used
	AlgorithmID string `json:"algorithm_id"`

	// RecipientKeyIDs are the fingerprints of intended recipients' public keys
	RecipientKeyIDs []string `json:"recipient_key_ids"`

	// EnvelopeHash is the hash of the full encryption envelope
	EnvelopeHash []byte `json:"envelope_hash"`

	// SenderKeyID is the fingerprint of the sender's public key
	SenderKeyID string `json:"sender_key_id"`
}

// ArtifactReference is the on-chain reference to an off-chain encrypted artifact
// This is what gets stored on the blockchain
type ArtifactReference struct {
	// Version is the format version
	Version uint32 `json:"version"`

	// ReferenceID is a unique identifier for this reference
	ReferenceID string `json:"reference_id"`

	// ContentAddress points to the artifact data
	ContentAddress *ContentAddress `json:"content_address"`

	// ChunkManifest describes chunking (if applicable)
	ChunkManifest *ChunkManifest `json:"chunk_manifest,omitempty"`

	// EncryptionMetadata contains encryption information
	EncryptionMetadata *EncryptionMetadata `json:"encryption_metadata"`

	// RetentionTag contains lifecycle metadata
	RetentionTag *RetentionTag `json:"retention_tag,omitempty"`

	// AccountAddress is the account this artifact belongs to
	AccountAddress string `json:"account_address"`

	// ArtifactType identifies what kind of artifact this is
	ArtifactType string `json:"artifact_type"`

	// CreatedAt is when this reference was created
	CreatedAt time.Time `json:"created_at"`

	// CreatedAtBlock is the block height when created
	CreatedAtBlock int64 `json:"created_at_block"`

	// Metadata contains optional additional metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewArtifactReference creates a new artifact reference
func NewArtifactReference(
	referenceID string,
	contentAddress *ContentAddress,
	encryptionMeta *EncryptionMetadata,
	accountAddress string,
	artifactType string,
	blockHeight int64,
) *ArtifactReference {
	return &ArtifactReference{
		Version:            ArtifactReferenceVersion,
		ReferenceID:        referenceID,
		ContentAddress:     contentAddress,
		EncryptionMetadata: encryptionMeta,
		AccountAddress:     accountAddress,
		ArtifactType:       artifactType,
		CreatedAt:          time.Now().UTC(),
		CreatedAtBlock:     blockHeight,
		Metadata:           make(map[string]string),
	}
}

// Validate validates the artifact reference
func (a *ArtifactReference) Validate() error {
	if a.Version == 0 {
		return ErrInvalidArtifactReference.Wrap("version cannot be zero")
	}
	if a.Version > ArtifactReferenceVersion {
		return ErrInvalidArtifactReference.Wrapf("unsupported version %d (max: %d)", a.Version, ArtifactReferenceVersion)
	}
	if a.ReferenceID == "" {
		return ErrInvalidArtifactReference.Wrap("reference_id cannot be empty")
	}
	if a.ContentAddress == nil {
		return ErrInvalidArtifactReference.Wrap("content_address cannot be nil")
	}
	if err := a.ContentAddress.Validate(); err != nil {
		return ErrInvalidArtifactReference.Wrapf("invalid content_address: %v", err)
	}
	if a.EncryptionMetadata == nil {
		return ErrInvalidArtifactReference.Wrap("encryption_metadata cannot be nil")
	}
	if a.AccountAddress == "" {
		return ErrInvalidArtifactReference.Wrap("account_address cannot be empty")
	}
	if a.ArtifactType == "" {
		return ErrInvalidArtifactReference.Wrap("artifact_type cannot be empty")
	}

	// Validate chunk manifest if present
	if a.ChunkManifest != nil {
		if err := a.ChunkManifest.Validate(); err != nil {
			return ErrInvalidArtifactReference.Wrapf("invalid chunk_manifest: %v", err)
		}
	}

	return nil
}

// SetChunkManifest sets the chunk manifest for chunked storage
func (a *ArtifactReference) SetChunkManifest(manifest *ChunkManifest) {
	a.ChunkManifest = manifest
}

// SetRetentionTag sets the retention tag
func (a *ArtifactReference) SetRetentionTag(tag *RetentionTag) {
	a.RetentionTag = tag
}

// SetMetadata sets a metadata key-value pair
func (a *ArtifactReference) SetMetadata(key, value string) {
	if a.Metadata == nil {
		a.Metadata = make(map[string]string)
	}
	a.Metadata[key] = value
}

// GetMetadata gets a metadata value by key
func (a *ArtifactReference) GetMetadata(key string) (string, bool) {
	if a.Metadata == nil {
		return "", false
	}
	val, ok := a.Metadata[key]
	return val, ok
}

// IsChunked returns true if the artifact is stored in chunks
func (a *ArtifactReference) IsChunked() bool {
	return a.ChunkManifest != nil && a.ChunkManifest.ChunkCount > 1
}

