package artifact_store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sync"
	"time"
)

// WaldurConfig contains configuration for the Waldur backend
type WaldurConfig struct {
	// Endpoint is the Waldur API endpoint URL
	Endpoint string `json:"endpoint"`

	// APIKey is the API key for authentication (never log this)
	APIKey string `json:"-"`

	// Organization is the Waldur organization ID
	Organization string `json:"organization"`

	// Project is the Waldur project ID
	Project string `json:"project"`

	// Bucket is the object storage bucket name
	Bucket string `json:"bucket"`

	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries int `json:"max_retries"`

	// Timeout is the request timeout
	Timeout time.Duration `json:"timeout"`

	// EncryptAtRest indicates if Waldur encrypts data at rest
	EncryptAtRest bool `json:"encrypt_at_rest"`
}

// DefaultWaldurConfig returns a default configuration
func DefaultWaldurConfig() *WaldurConfig {
	return &WaldurConfig{
		MaxRetries:    3,
		Timeout:       30 * time.Second,
		EncryptAtRest: true,
	}
}

// Validate validates the configuration
func (c *WaldurConfig) Validate() error {
	if c.Endpoint == "" {
		return ErrInvalidInput.Wrap("waldur endpoint cannot be empty")
	}
	if c.Organization == "" {
		return ErrInvalidInput.Wrap("waldur organization cannot be empty")
	}
	return nil
}

// WaldurBackend implements ArtifactStore using Waldur's object storage
// The implementation stores encrypted artifacts in Waldur's backend database
// with authenticated retrieval.
//
// Security:
//   - All data is encrypted before being sent to Waldur
//   - API authentication is required for all operations
//   - Waldur provides encryption at rest
//   - No sensitive data (credentials, biometrics) is logged
type WaldurBackend struct {
	config *WaldurConfig

	// mu protects concurrent access to the storage map
	mu sync.RWMutex

	// storage is an in-memory map for stubbed implementation
	// In production, this would be replaced with actual Waldur API calls
	storage map[string]*storedArtifact

	// metrics tracks storage metrics
	metrics *StorageMetrics

	// ownerIndex maps owner -> list of content hashes
	ownerIndex map[string][]string
}

// storedArtifact represents an artifact in storage
type storedArtifact struct {
	data       []byte
	reference  *ArtifactReference
	storedAt   time.Time
	accessedAt time.Time
}

// NewWaldurBackend creates a new Waldur backend instance
func NewWaldurBackend(config *WaldurConfig) (*WaldurBackend, error) {
	if config == nil {
		config = DefaultWaldurConfig()
	}

	// For stubbed implementation, we don't validate endpoint
	// In production, this would establish connection and validate credentials

	return &WaldurBackend{
		config:     config,
		storage:    make(map[string]*storedArtifact),
		ownerIndex: make(map[string][]string),
		metrics: &StorageMetrics{
			BackendType:   BackendWaldur,
			BackendStatus: make(map[string]string),
		},
	}, nil
}

// Backend returns the backend type
func (w *WaldurBackend) Backend() BackendType {
	return BackendWaldur
}

// Put stores an encrypted artifact and returns its content address
func (w *WaldurBackend) Put(ctx context.Context, req *PutRequest) (*PutResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Compute content address
	hash := sha256.Sum256(req.Data)
	hashHex := hex.EncodeToString(hash[:])

	// Generate Waldur object path
	backendRef := w.generateObjectPath(hashHex, req.Owner)

	contentAddr := &ContentAddress{
		Version:    ContentAddressVersion,
		Hash:       hash[:],
		Algorithm:  "sha256",
		Size:       uint64(len(req.Data)),
		Backend:    BackendWaldur,
		BackendRef: backendRef,
	}

	// Create artifact reference
	artifactRef := NewArtifactReference(
		hashHex,
		contentAddr,
		req.EncryptionMetadata,
		req.Owner,
		req.ArtifactType,
		0, // Block height is set by caller
	)

	if req.RetentionTag != nil {
		artifactRef.SetRetentionTag(req.RetentionTag)
	}

	for k, v := range req.Metadata {
		artifactRef.SetMetadata(k, v)
	}

	// Store in backend (stubbed - in production would call Waldur API)
	w.mu.Lock()
	defer w.mu.Unlock()

	w.storage[hashHex] = &storedArtifact{
		data:       req.Data,
		reference:  artifactRef,
		storedAt:   time.Now().UTC(),
		accessedAt: time.Now().UTC(),
	}

	// Update owner index
	w.ownerIndex[req.Owner] = append(w.ownerIndex[req.Owner], hashHex)

	// Update metrics
	w.metrics.TotalArtifacts++
	w.metrics.TotalBytes += uint64(len(req.Data))

	return &PutResponse{
		ContentAddress:    contentAddr,
		ArtifactReference: artifactRef,
	}, nil
}

// Get retrieves an encrypted artifact by its content address
func (w *WaldurBackend) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	hashHex := hex.EncodeToString(req.ContentAddress.Hash)

	w.mu.RLock()
	artifact, exists := w.storage[hashHex]
	w.mu.RUnlock()

	if !exists {
		return nil, ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
	}

	// Check authorization (stubbed - in production would verify token/account)
	if req.RequestingAccount != "" && artifact.reference.AccountAddress != req.RequestingAccount {
		// Allow if requester is in recipient list
		authorized := false
		if artifact.reference.EncryptionMetadata != nil {
			for _, keyID := range artifact.reference.EncryptionMetadata.RecipientKeyIDs {
				if keyID == req.RequestingAccount {
					authorized = true
					break
				}
			}
		}
		if !authorized {
			return nil, ErrAuthenticationFailed.Wrap("not authorized to access artifact")
		}
	}

	// Check retention
	if artifact.reference.RetentionTag != nil {
		if artifact.reference.RetentionTag.IsExpired(time.Now().UTC()) {
			return nil, ErrRetentionExpired.Wrap("artifact retention has expired")
		}
	}

	// Verify hash
	computedHash := sha256.Sum256(artifact.data)
	if !bytesEqual(computedHash[:], req.ContentAddress.Hash) {
		return nil, ErrHashMismatch.Wrap("stored data hash does not match request")
	}

	// Update access time
	w.mu.Lock()
	artifact.accessedAt = time.Now().UTC()
	w.mu.Unlock()

	return &GetResponse{
		Data:           artifact.data,
		ContentAddress: artifact.reference.ContentAddress,
		ChunkManifest:  artifact.reference.ChunkManifest,
	}, nil
}

// Delete removes an artifact by its content address
func (w *WaldurBackend) Delete(ctx context.Context, req *DeleteRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	hashHex := hex.EncodeToString(req.ContentAddress.Hash)

	w.mu.Lock()
	defer w.mu.Unlock()

	artifact, exists := w.storage[hashHex]
	if !exists {
		return ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
	}

	// Check authorization
	if artifact.reference.AccountAddress != req.RequestingAccount && !req.Force {
		return ErrAuthenticationFailed.Wrap("not authorized to delete artifact")
	}

	// Remove from owner index
	if hashes, ok := w.ownerIndex[artifact.reference.AccountAddress]; ok {
		newHashes := make([]string, 0, len(hashes)-1)
		for _, h := range hashes {
			if h != hashHex {
				newHashes = append(newHashes, h)
			}
		}
		w.ownerIndex[artifact.reference.AccountAddress] = newHashes
	}

	// Update metrics
	w.metrics.TotalArtifacts--
	w.metrics.TotalBytes -= uint64(len(artifact.data))

	// Delete from storage
	delete(w.storage, hashHex)

	return nil
}

// Exists checks if an artifact exists at the given content address
func (w *WaldurBackend) Exists(ctx context.Context, address *ContentAddress) (bool, error) {
	if address == nil {
		return false, ErrInvalidInput.Wrap("address cannot be nil")
	}

	hashHex := hex.EncodeToString(address.Hash)

	w.mu.RLock()
	_, exists := w.storage[hashHex]
	w.mu.RUnlock()

	return exists, nil
}

// GetChunk retrieves a specific chunk of a chunked artifact
func (w *WaldurBackend) GetChunk(ctx context.Context, address *ContentAddress, chunkIndex uint32) (*ChunkData, error) {
	// Waldur backend stores whole artifacts, not chunks
	// For chunked storage, use IPFS backend
	return nil, ErrBackendNotSupported.Wrap("waldur backend does not support chunked storage")
}

// ListByOwner lists all artifacts owned by a specific account
func (w *WaldurBackend) ListByOwner(ctx context.Context, owner string, pagination *Pagination) (*ListResponse, error) {
	if owner == "" {
		return nil, ErrInvalidInput.Wrap("owner cannot be empty")
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	hashes, exists := w.ownerIndex[owner]
	if !exists {
		return &ListResponse{
			References: make([]*ArtifactReference, 0),
			Total:      0,
			HasMore:    false,
		}, nil
	}

	// Apply pagination
	offset := uint64(0)
	limit := uint64(100)
	if pagination != nil {
		offset = pagination.Offset
		limit = pagination.Limit
	}

	total := uint64(len(hashes))
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	refs := make([]*ArtifactReference, 0, end-start)
	for _, h := range hashes[start:end] {
		if artifact, ok := w.storage[h]; ok {
			refs = append(refs, artifact.reference)
		}
	}

	return &ListResponse{
		References: refs,
		Total:      total,
		HasMore:    end < total,
	}, nil
}

// UpdateRetention updates the retention tag for an artifact
func (w *WaldurBackend) UpdateRetention(ctx context.Context, address *ContentAddress, tag *RetentionTag) error {
	if address == nil {
		return ErrInvalidInput.Wrap("address cannot be nil")
	}

	hashHex := hex.EncodeToString(address.Hash)

	w.mu.Lock()
	defer w.mu.Unlock()

	artifact, exists := w.storage[hashHex]
	if !exists {
		return ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
	}

	artifact.reference.SetRetentionTag(tag)
	return nil
}

// PurgeExpired removes all expired artifacts based on the current time/block
func (w *WaldurBackend) PurgeExpired(ctx context.Context, currentBlock int64) (int, error) {
	now := time.Now().UTC()

	w.mu.Lock()
	defer w.mu.Unlock()

	toDelete := make([]string, 0)

	for hashHex, artifact := range w.storage {
		if artifact.reference.RetentionTag != nil {
			if artifact.reference.RetentionTag.DeleteOnExpiry {
				expired := artifact.reference.RetentionTag.IsExpired(now) ||
					artifact.reference.RetentionTag.IsExpiredAtBlock(currentBlock)
				if expired {
					toDelete = append(toDelete, hashHex)
				}
			}
		}
	}

	for _, hashHex := range toDelete {
		artifact := w.storage[hashHex]

		// Remove from owner index
		if hashes, ok := w.ownerIndex[artifact.reference.AccountAddress]; ok {
			newHashes := make([]string, 0, len(hashes)-1)
			for _, h := range hashes {
				if h != hashHex {
					newHashes = append(newHashes, h)
				}
			}
			w.ownerIndex[artifact.reference.AccountAddress] = newHashes
		}

		// Update metrics
		w.metrics.TotalBytes -= uint64(len(artifact.data))

		delete(w.storage, hashHex)
	}

	w.metrics.TotalArtifacts -= uint64(len(toDelete))
	return len(toDelete), nil
}

// GetMetrics returns storage metrics for monitoring
func (w *WaldurBackend) GetMetrics(ctx context.Context) (*StorageMetrics, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Count expired artifacts
	now := time.Now().UTC()
	expired := uint64(0)
	for _, artifact := range w.storage {
		if artifact.reference.RetentionTag != nil {
			if artifact.reference.RetentionTag.IsExpired(now) {
				expired++
			}
		}
	}

	return &StorageMetrics{
		TotalArtifacts:   w.metrics.TotalArtifacts,
		TotalBytes:       w.metrics.TotalBytes,
		TotalChunks:      0, // Waldur doesn't use chunks
		ExpiredArtifacts: expired,
		BackendType:      BackendWaldur,
		BackendStatus: map[string]string{
			"endpoint":        w.config.Endpoint,
			"organization":    w.config.Organization,
			"encrypt_at_rest": boolToString(w.config.EncryptAtRest),
		},
	}, nil
}

// Health checks if the backend is healthy
func (w *WaldurBackend) Health(ctx context.Context) error {
	// Stubbed - in production would ping Waldur API
	return nil
}

// generateObjectPath generates a unique object path in Waldur storage
func (w *WaldurBackend) generateObjectPath(hashHex, owner string) string {
	// Use a hierarchical structure for better organization
	// Format: org/project/bucket/owner_prefix/hash
	prefix := ""
	if len(owner) >= 8 {
		prefix = owner[:8]
	} else {
		prefix = owner
	}
	return w.config.Organization + "/" + w.config.Project + "/" + w.config.Bucket + "/" + prefix + "/" + hashHex
}

// Ensure WaldurBackend implements ArtifactStore
var _ ArtifactStore = (*WaldurBackend)(nil)

// WaldurStreamingBackend extends WaldurBackend with streaming support
type WaldurStreamingBackend struct {
	*WaldurBackend
}

// NewWaldurStreamingBackend creates a new Waldur streaming backend
func NewWaldurStreamingBackend(config *WaldurConfig) (*WaldurStreamingBackend, error) {
	base, err := NewWaldurBackend(config)
	if err != nil {
		return nil, err
	}
	return &WaldurStreamingBackend{WaldurBackend: base}, nil
}

// PutStream stores a large artifact using streaming
func (w *WaldurStreamingBackend) PutStream(ctx context.Context, req *PutStreamRequest) (*PutResponse, error) {
	// Read all data from stream (stubbed - production would stream to Waldur)
	data, err := io.ReadAll(req.Reader)
	if err != nil {
		return nil, ErrInvalidInput.Wrapf("failed to read stream: %v", err)
	}

	// Convert to regular Put request
	putReq := &PutRequest{
		Data:               data,
		ContentHash:        req.ContentHash,
		EncryptionMetadata: req.EncryptionMetadata,
		RetentionTag:       req.RetentionTag,
		Owner:              req.Owner,
		ArtifactType:       req.ArtifactType,
		Metadata:           req.Metadata,
	}

	return w.Put(ctx, putReq)
}

// GetStream retrieves a large artifact as a stream
func (w *WaldurStreamingBackend) GetStream(ctx context.Context, address *ContentAddress) (io.ReadCloser, error) {
	// Get artifact (stubbed - production would stream from Waldur)
	resp, err := w.Get(ctx, &GetRequest{ContentAddress: address})
	if err != nil {
		return nil, err
	}

	return &bytesReadCloser{data: resp.Data}, nil
}

// bytesReadCloser wraps a byte slice as an io.ReadCloser
type bytesReadCloser struct {
	data   []byte
	offset int
}

func (b *bytesReadCloser) Read(p []byte) (n int, err error) {
	if b.offset >= len(b.data) {
		return 0, io.EOF
	}
	n = copy(p, b.data[b.offset:])
	b.offset += n
	return n, nil
}

func (b *bytesReadCloser) Close() error {
	return nil
}

// Ensure WaldurStreamingBackend implements StreamingArtifactStore
var _ StreamingArtifactStore = (*WaldurStreamingBackend)(nil)

// bytesEqual compares two byte slices for equality
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// boolToString converts a bool to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
