package artifact_store

import (
	"context"
	"io"
)

// ArtifactStore defines the interface for storing and retrieving identity artifacts.
// This is the primary abstraction for off-chain encrypted storage with on-chain references.
//
// Implementations must ensure:
//   - All artifacts are encrypted before storage
//   - Content-addressed retrieval is deterministic
//   - Retention tags are respected
//   - Only authorized callers can access artifacts
type ArtifactStore interface {
	// Put stores an encrypted artifact and returns its content address.
	// The artifact data must already be encrypted using an EncryptedPayloadEnvelope.
	// Returns the content address for on-chain storage.
	Put(ctx context.Context, req *PutRequest) (*PutResponse, error)

	// Get retrieves an encrypted artifact by its content address.
	// The caller is responsible for decryption using the appropriate private key.
	Get(ctx context.Context, req *GetRequest) (*GetResponse, error)

	// Delete removes an artifact by its content address.
	// This is a privileged operation requiring authorization.
	Delete(ctx context.Context, req *DeleteRequest) error

	// Exists checks if an artifact exists at the given content address.
	Exists(ctx context.Context, address *ContentAddress) (bool, error)

	// GetChunk retrieves a specific chunk of a chunked artifact.
	// For non-chunked artifacts, this returns an error.
	GetChunk(ctx context.Context, address *ContentAddress, chunkIndex uint32) (*ChunkData, error)

	// ListByOwner lists all artifacts owned by a specific account.
	// Returns artifact references (not the data itself).
	ListByOwner(ctx context.Context, owner string, pagination *Pagination) (*ListResponse, error)

	// UpdateRetention updates the retention tag for an artifact.
	UpdateRetention(ctx context.Context, address *ContentAddress, tag *RetentionTag) error

	// PurgeExpired removes all expired artifacts based on the current time/block.
	// Returns the number of artifacts purged.
	PurgeExpired(ctx context.Context, currentBlock int64) (int, error)

	// GetMetrics returns storage metrics for monitoring.
	GetMetrics(ctx context.Context) (*StorageMetrics, error)

	// Health checks if the backend is healthy.
	Health(ctx context.Context) error

	// Backend returns the backend type.
	Backend() BackendType
}

// StreamingArtifactStore extends ArtifactStore with streaming support
// for large artifacts that don't fit in memory.
type StreamingArtifactStore interface {
	ArtifactStore

	// PutStream stores a large artifact using streaming.
	// The reader should provide encrypted data.
	PutStream(ctx context.Context, req *PutStreamRequest) (*PutResponse, error)

	// GetStream retrieves a large artifact as a stream.
	// The caller is responsible for decryption.
	GetStream(ctx context.Context, address *ContentAddress) (io.ReadCloser, error)
}

// ChunkedArtifactStore extends ArtifactStore with chunking support
// for fragmented storage (primarily for IPFS backend).
type ChunkedArtifactStore interface {
	ArtifactStore

	// PutChunked stores an artifact in chunks according to the manifest.
	// Returns the content address and computed chunk manifest.
	PutChunked(ctx context.Context, data []byte, chunkSize uint64, meta *PutMetadata) (*PutResponse, *ChunkManifest, error)

	// GetChunked retrieves and reassembles a chunked artifact.
	// Uses the manifest for deterministic reconstruction.
	GetChunked(ctx context.Context, manifest *ChunkManifest) ([]byte, error)

	// VerifyChunks verifies all chunks match the manifest hashes.
	VerifyChunks(ctx context.Context, manifest *ChunkManifest) error
}

// PutRequest contains parameters for storing an artifact
type PutRequest struct {
	// Data is the encrypted artifact data
	Data []byte

	// ContentHash is the hash of the original unencrypted data (for verification)
	ContentHash []byte

	// EncryptionMetadata contains encryption envelope information
	EncryptionMetadata *EncryptionMetadata

	// RetentionTag specifies retention policy
	RetentionTag *RetentionTag

	// Owner is the account address that owns this artifact
	Owner string

	// ArtifactType identifies the type of artifact
	ArtifactType string

	// Metadata contains optional additional metadata
	Metadata map[string]string
}

// Validate validates the put request
func (r *PutRequest) Validate() error {
	if len(r.Data) == 0 {
		return ErrInvalidInput.Wrap("data cannot be empty")
	}
	if r.EncryptionMetadata == nil {
		return ErrInvalidInput.Wrap("encryption_metadata cannot be nil")
	}
	if r.Owner == "" {
		return ErrInvalidInput.Wrap("owner cannot be empty")
	}
	if r.ArtifactType == "" {
		return ErrInvalidInput.Wrap("artifact_type cannot be empty")
	}
	return nil
}

// PutResponse contains the result of a Put operation
type PutResponse struct {
	// ContentAddress is the content-addressed reference
	ContentAddress *ContentAddress

	// ArtifactReference is the complete artifact reference for on-chain storage
	ArtifactReference *ArtifactReference
}

// PutMetadata contains metadata for Put operations
type PutMetadata struct {
	// EncryptionMetadata contains encryption information
	EncryptionMetadata *EncryptionMetadata

	// RetentionTag specifies retention policy
	RetentionTag *RetentionTag

	// Owner is the account address that owns this artifact
	Owner string

	// ArtifactType identifies the type of artifact
	ArtifactType string

	// Metadata contains optional additional metadata
	Metadata map[string]string
}

// PutStreamRequest contains parameters for streaming storage
type PutStreamRequest struct {
	// Reader provides the encrypted artifact data
	Reader io.Reader

	// Size is the expected total size (if known)
	Size int64

	// ContentHash is the hash of the original unencrypted data (for verification)
	ContentHash []byte

	// EncryptionMetadata contains encryption envelope information
	EncryptionMetadata *EncryptionMetadata

	// RetentionTag specifies retention policy
	RetentionTag *RetentionTag

	// Owner is the account address that owns this artifact
	Owner string

	// ArtifactType identifies the type of artifact
	ArtifactType string

	// Metadata contains optional additional metadata
	Metadata map[string]string
}

// GetRequest contains parameters for retrieving an artifact
type GetRequest struct {
	// ContentAddress is the artifact's content address
	ContentAddress *ContentAddress

	// RequestingAccount is the account requesting access (for authorization)
	RequestingAccount string

	// AuthToken is an optional authentication token
	AuthToken string
}

// Validate validates the get request
func (r *GetRequest) Validate() error {
	if r.ContentAddress == nil {
		return ErrInvalidInput.Wrap("content_address cannot be nil")
	}
	if err := r.ContentAddress.Validate(); err != nil {
		return err
	}
	return nil
}

// GetResponse contains the retrieved artifact data
type GetResponse struct {
	// Data is the encrypted artifact data
	Data []byte

	// ContentAddress is the verified content address
	ContentAddress *ContentAddress

	// ChunkManifest describes chunking (if applicable)
	ChunkManifest *ChunkManifest
}

// DeleteRequest contains parameters for deleting an artifact
type DeleteRequest struct {
	// ContentAddress is the artifact's content address
	ContentAddress *ContentAddress

	// RequestingAccount is the account requesting deletion (for authorization)
	RequestingAccount string

	// AuthToken is an optional authentication token
	AuthToken string

	// Force indicates if deletion should be forced even if not expired
	Force bool
}

// Validate validates the delete request
func (r *DeleteRequest) Validate() error {
	if r.ContentAddress == nil {
		return ErrInvalidInput.Wrap("content_address cannot be nil")
	}
	if err := r.ContentAddress.Validate(); err != nil {
		return err
	}
	if r.RequestingAccount == "" {
		return ErrInvalidInput.Wrap("requesting_account cannot be empty")
	}
	return nil
}

// ChunkData contains a single chunk's data
type ChunkData struct {
	// Index is the chunk index
	Index uint32

	// Data is the chunk data
	Data []byte

	// Hash is the chunk hash for verification
	Hash []byte
}

// Verify verifies the chunk data matches its hash
func (c *ChunkData) Verify() error {
	if c == nil {
		return ErrInvalidInput.Wrap("chunk data is nil")
	}

	// Compute hash
	addr := NewContentAddress(c.Data, "", "")
	if len(addr.Hash) != len(c.Hash) {
		return ErrHashMismatch.Wrap("hash length mismatch")
	}
	for i := range addr.Hash {
		if addr.Hash[i] != c.Hash[i] {
			return ErrHashMismatch.Wrap("chunk hash verification failed")
		}
	}
	return nil
}

// Pagination contains pagination parameters
type Pagination struct {
	// Offset is the starting offset
	Offset uint64

	// Limit is the maximum number of results
	Limit uint64

	// OrderBy specifies the sort field
	OrderBy string

	// Descending indicates descending order
	Descending bool
}

// ListResponse contains a list of artifact references
type ListResponse struct {
	// References are the matching artifact references
	References []*ArtifactReference

	// Total is the total count (ignoring pagination)
	Total uint64

	// HasMore indicates if more results are available
	HasMore bool
}

// StorageMetrics contains storage utilization metrics
type StorageMetrics struct {
	// TotalArtifacts is the total number of artifacts stored
	TotalArtifacts uint64

	// TotalBytes is the total storage used in bytes
	TotalBytes uint64

	// TotalChunks is the total number of chunks stored
	TotalChunks uint64

	// ExpiredArtifacts is the number of expired artifacts awaiting cleanup
	ExpiredArtifacts uint64

	// BackendType is the backend type
	BackendType BackendType

	// BackendStatus is the backend-specific status
	BackendStatus map[string]string
}
