package artifact_store

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/waldur"
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

	// QuotaBytes is the storage quota in bytes (0 = unlimited)
	QuotaBytes int64 `json:"quota_bytes"`

	// StreamingChunkSize is the chunk size for streaming uploads (default: 8MB)
	StreamingChunkSize int64 `json:"streaming_chunk_size"`

	// HealthCheckTimeout is the timeout for health checks
	HealthCheckTimeout time.Duration `json:"health_check_timeout"`

	// UseFallbackMemory enables in-memory fallback when Waldur is unavailable (testing only)
	UseFallbackMemory bool `json:"use_fallback_memory"`
}

// DefaultWaldurConfig returns a default configuration
func DefaultWaldurConfig() *WaldurConfig {
	return &WaldurConfig{
		MaxRetries:         3,
		Timeout:            30 * time.Second,
		EncryptAtRest:      true,
		StreamingChunkSize: 8 * 1024 * 1024, // 8MB
		HealthCheckTimeout: 10 * time.Second,
		UseFallbackMemory:  false,
	}
}

// Validate validates the configuration
func (c *WaldurConfig) Validate() error {
	if c.Endpoint == "" && !c.UseFallbackMemory {
		return ErrInvalidInput.Wrap("waldur endpoint cannot be empty")
	}
	if c.Organization == "" {
		return ErrInvalidInput.Wrap("waldur organization cannot be empty")
	}
	if c.Bucket == "" && !c.UseFallbackMemory {
		return ErrInvalidInput.Wrap("waldur bucket cannot be empty")
	}
	return nil
}

// WaldurBackend implements ArtifactStore using Waldur's object storage
// The implementation stores encrypted artifacts in Waldur's backend storage
// with authenticated retrieval.
//
// Security:
//   - All data is encrypted before being sent to Waldur
//   - API authentication is required for all operations
//   - Waldur provides encryption at rest
//   - No sensitive data (credentials, biometrics) is logged
type WaldurBackend struct {
	config *WaldurConfig

	// mu protects concurrent access to internal state
	mu sync.RWMutex

	// waldurClient is the Waldur API client
	waldurClient *waldur.Client

	// objectStorage is the Waldur object storage client
	objectStorage *waldur.ObjectStorageClient

	// fallbackStorage is an in-memory fallback for testing/development
	fallbackStorage map[string]*storedArtifact

	// ownerIndex maps owner -> list of content hashes (for fallback and caching)
	ownerIndex map[string][]string

	// metrics tracks storage metrics
	metrics *StorageMetrics

	// useFallback indicates if we're using fallback memory storage
	useFallback bool
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

	backend := &WaldurBackend{
		config:          config,
		fallbackStorage: make(map[string]*storedArtifact),
		ownerIndex:      make(map[string][]string),
		metrics: &StorageMetrics{
			BackendType:   BackendWaldur,
			BackendStatus: make(map[string]string),
		},
		useFallback: config.UseFallbackMemory,
	}

	// If fallback mode is enabled, skip Waldur client initialization
	if config.UseFallbackMemory {
		return backend, nil
	}

	// Validate configuration for production use
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// Create Waldur client configuration
	waldurCfg := waldur.DefaultConfig()
	waldurCfg.BaseURL = config.Endpoint
	waldurCfg.Token = config.APIKey
	waldurCfg.Timeout = config.Timeout
	waldurCfg.MaxRetries = config.MaxRetries

	// Initialize Waldur client
	waldurClient, err := waldur.NewClient(waldurCfg)
	if err != nil {
		return nil, ErrBackendUnavailable.Wrapf("failed to create waldur client: %v", err)
	}

	backend.waldurClient = waldurClient
	backend.objectStorage = waldur.NewObjectStorageClient(waldurClient)

	return backend, nil
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

	// Use fallback storage if enabled
	if w.useFallback {
		return w.putFallback(ctx, req, hashHex, contentAddr, artifactRef)
	}

	// Check quota before upload
	if w.config.QuotaBytes > 0 {
		if err := w.objectStorage.CheckQuota(ctx, w.config.Bucket, int64(len(req.Data))); err != nil {
			if errors.Is(err, waldur.ErrQuotaExceeded) {
				return nil, ErrStorageLimitExceeded.Wrap(err.Error())
			}
			// Log but continue if quota check fails (soft enforcement)
		}
	}

	// Upload to Waldur object storage
	uploadReq := &waldur.UploadRequest{
		Bucket:      w.config.Bucket,
		Key:         backendRef,
		Body:        bytes.NewReader(req.Data),
		Size:        int64(len(req.Data)),
		ContentType: "application/octet-stream",
		ContentHash: hashHex,
		Metadata: map[string]string{
			"owner":         req.Owner,
			"artifact_type": req.ArtifactType,
			"algorithm":     "sha256",
		},
	}

	_, err := w.objectStorage.Upload(ctx, uploadReq)
	if err != nil {
		return nil, ErrBackendUnavailable.Wrapf("failed to upload: %v", err)
	}

	// Update metrics
	w.mu.Lock()
	w.metrics.TotalArtifacts++
	w.metrics.TotalBytes += uint64(len(req.Data))
	w.ownerIndex[req.Owner] = append(w.ownerIndex[req.Owner], hashHex)
	w.mu.Unlock()

	return &PutResponse{
		ContentAddress:    contentAddr,
		ArtifactReference: artifactRef,
	}, nil
}

// putFallback stores artifact in memory (for testing/development)
func (w *WaldurBackend) putFallback(ctx context.Context, req *PutRequest, hashHex string, contentAddr *ContentAddress, artifactRef *ArtifactReference) (*PutResponse, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.fallbackStorage[hashHex] = &storedArtifact{
		data:       req.Data,
		reference:  artifactRef,
		storedAt:   time.Now().UTC(),
		accessedAt: time.Now().UTC(),
	}

	w.ownerIndex[req.Owner] = append(w.ownerIndex[req.Owner], hashHex)
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

	// Use fallback storage if enabled
	if w.useFallback {
		return w.getFallback(ctx, req, hashHex)
	}

	// Download from Waldur object storage
	downloadReq := &waldur.DownloadRequest{
		Bucket: w.config.Bucket,
		Key:    req.ContentAddress.BackendRef,
	}

	downloadResp, err := w.objectStorage.Download(ctx, downloadReq)
	if err != nil {
		if errors.Is(err, waldur.ErrObjectNotFound) {
			return nil, ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
		}
		return nil, ErrBackendUnavailable.Wrapf("download failed: %v", err)
	}
	defer func() { _ = downloadResp.Body.Close() }()

	// Read all data
	data, err := io.ReadAll(downloadResp.Body)
	if err != nil {
		return nil, ErrBackendUnavailable.Wrapf("failed to read artifact: %v", err)
	}

	// Verify hash
	computedHash := sha256.Sum256(data)
	if !bytesEqual(computedHash[:], req.ContentAddress.Hash) {
		return nil, ErrHashMismatch.Wrap("stored data hash does not match request")
	}

	return &GetResponse{
		Data:           data,
		ContentAddress: req.ContentAddress,
	}, nil
}

// getFallback retrieves from memory storage (for testing/development)
func (w *WaldurBackend) getFallback(ctx context.Context, req *GetRequest, hashHex string) (*GetResponse, error) {
	w.mu.RLock()
	artifact, exists := w.fallbackStorage[hashHex]
	w.mu.RUnlock()

	if !exists {
		return nil, ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
	}

	// Check authorization
	if req.RequestingAccount != "" && artifact.reference.AccountAddress != req.RequestingAccount {
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

	// Use fallback storage if enabled
	if w.useFallback {
		return w.deleteFallback(ctx, req, hashHex)
	}

	// Delete from Waldur object storage
	err := w.objectStorage.Delete(ctx, w.config.Bucket, req.ContentAddress.BackendRef)
	if err != nil {
		if errors.Is(err, waldur.ErrObjectNotFound) {
			return ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
		}
		return ErrBackendUnavailable.Wrapf("delete failed: %v", err)
	}

	// Update metrics (note: size not known without HEAD call)
	w.mu.Lock()
	if w.metrics.TotalArtifacts > 0 {
		w.metrics.TotalArtifacts--
	}
	// Remove from owner index if we have it cached
	for owner, hashes := range w.ownerIndex {
		for i, h := range hashes {
			if h == hashHex {
				w.ownerIndex[owner] = append(hashes[:i], hashes[i+1:]...)
				break
			}
		}
	}
	w.mu.Unlock()

	return nil
}

// deleteFallback removes from memory storage (for testing/development)
func (w *WaldurBackend) deleteFallback(ctx context.Context, req *DeleteRequest, hashHex string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	artifact, exists := w.fallbackStorage[hashHex]
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

	delete(w.fallbackStorage, hashHex)
	return nil
}

// Exists checks if an artifact exists at the given content address
func (w *WaldurBackend) Exists(ctx context.Context, address *ContentAddress) (bool, error) {
	if address == nil {
		return false, ErrInvalidInput.Wrap("address cannot be nil")
	}

	// Use fallback storage if enabled
	if w.useFallback {
		hashHex := hex.EncodeToString(address.Hash)
		w.mu.RLock()
		_, exists := w.fallbackStorage[hashHex]
		w.mu.RUnlock()
		return exists, nil
	}

	// Check Waldur object storage
	return w.objectStorage.Exists(ctx, w.config.Bucket, address.BackendRef)
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

	// For production Waldur storage, we need to list from the object storage API
	if !w.useFallback {
		// List objects with prefix matching owner
		prefix := w.generateOwnerPrefix(owner)
		listReq := &waldur.ListRequest{
			Bucket:  w.config.Bucket,
			Prefix:  prefix,
			MaxKeys: 1000,
		}

		listResp, err := w.objectStorage.List(ctx, listReq)
		if err != nil {
			return nil, ErrBackendUnavailable.Wrapf("list failed: %v", err)
		}

		// Convert to artifact references (simplified - full metadata would need HEAD calls)
		refs := make([]*ArtifactReference, 0, len(listResp.Objects))
		for _, obj := range listResp.Objects {
			ref := &ArtifactReference{
				Version:        ArtifactReferenceVersion,
				ReferenceID:    obj.Key,
				AccountAddress: owner,
				CreatedAt:      obj.CreatedAt,
				ContentAddress: &ContentAddress{
					Version:    ContentAddressVersion,
					Backend:    BackendWaldur,
					BackendRef: obj.Key,
					Size:       uint64(obj.Size),
				},
			}
			refs = append(refs, ref)
		}

		// Apply pagination
		total := uint64(len(refs))
		offset := uint64(0)
		limit := uint64(100)
		if pagination != nil {
			offset = pagination.Offset
			limit = pagination.Limit
		}

		start := offset
		if start > total {
			start = total
		}
		end := start + limit
		if end > total {
			end = total
		}

		return &ListResponse{
			References: refs[start:end],
			Total:      total,
			HasMore:    end < total || listResp.IsTruncated,
		}, nil
	}

	// Fallback storage
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
		if artifact, ok := w.fallbackStorage[h]; ok {
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

	// For production Waldur, we would update object metadata
	// For now, only fallback storage supports retention updates
	if !w.useFallback {
		// In production, this would update object metadata via API
		// Waldur object storage may not support custom metadata updates
		// This is a limitation that should be documented
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	artifact, exists := w.fallbackStorage[hashHex]
	if !exists {
		return ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
	}

	artifact.reference.SetRetentionTag(tag)
	return nil
}

// PurgeExpired removes all expired artifacts based on the current time/block
func (w *WaldurBackend) PurgeExpired(ctx context.Context, currentBlock int64) (int, error) {
	now := time.Now().UTC()

	// For production Waldur, we need to list and delete expired objects
	if !w.useFallback {
		// List all objects and check retention metadata
		// This is inefficient for large buckets - in production, use lifecycle policies
		listReq := &waldur.ListRequest{
			Bucket:  w.config.Bucket,
			MaxKeys: 1000,
		}

		listResp, err := w.objectStorage.List(ctx, listReq)
		if err != nil {
			return 0, ErrBackendUnavailable.Wrapf("list failed: %v", err)
		}

		deleted := 0
		for _, obj := range listResp.Objects {
			// Check if object has expired based on metadata
			// In production, retention would be stored in object metadata
			if expiry, ok := obj.Metadata["retention_expires_at"]; ok {
				expiryTime, err := time.Parse(time.RFC3339, expiry)
				if err == nil && now.After(expiryTime) {
					if deleteOnExpiry, ok := obj.Metadata["delete_on_expiry"]; ok && deleteOnExpiry == "true" {
						if err := w.objectStorage.Delete(ctx, w.config.Bucket, obj.Key); err == nil {
							deleted++
						}
					}
				}
			}
		}

		return deleted, nil
	}

	// Fallback storage purge
	w.mu.Lock()
	defer w.mu.Unlock()

	toDelete := make([]string, 0)

	for hashHex, artifact := range w.fallbackStorage {
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
		artifact := w.fallbackStorage[hashHex]

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

		delete(w.fallbackStorage, hashHex)
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

	// For production, get quota info from Waldur
	if !w.useFallback && w.objectStorage != nil {
		quota, err := w.objectStorage.GetQuota(ctx, w.config.Bucket)
		if err == nil {
			return &StorageMetrics{
				TotalArtifacts:   uint64(quota.ObjectCount),
				TotalBytes:       uint64(quota.UsedBytes),
				TotalChunks:      0, // Waldur doesn't use chunks
				ExpiredArtifacts: expired,
				BackendType:      BackendWaldur,
				BackendStatus: map[string]string{
					"endpoint":        w.config.Endpoint,
					"organization":    w.config.Organization,
					"bucket":          w.config.Bucket,
					"encrypt_at_rest": boolToString(w.config.EncryptAtRest),
					"quota_total":     fmt.Sprintf("%d", quota.TotalBytes),
					"quota_used":      fmt.Sprintf("%d", quota.UsedBytes),
					"quota_available": fmt.Sprintf("%d", quota.AvailableBytes),
				},
			}, nil
		}
		// Fall through to cached metrics if quota call fails
	}

	// Fallback metrics from local cache
	for _, artifact := range w.fallbackStorage {
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
			"bucket":          w.config.Bucket,
			"encrypt_at_rest": boolToString(w.config.EncryptAtRest),
		},
	}, nil
}

// Health checks if the backend is healthy
func (w *WaldurBackend) Health(ctx context.Context) error {
	// Fallback mode is always healthy
	if w.useFallback {
		return nil
	}

	// Check object storage health
	healthCtx, cancel := context.WithTimeout(ctx, w.config.HealthCheckTimeout)
	defer cancel()

	if err := w.objectStorage.Ping(healthCtx); err != nil {
		return ErrBackendUnavailable.Wrapf("health check failed: %v", err)
	}

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

// generateOwnerPrefix generates the storage prefix for an owner
func (w *WaldurBackend) generateOwnerPrefix(owner string) string {
	prefix := ""
	if len(owner) >= 8 {
		prefix = owner[:8]
	} else {
		prefix = owner
	}
	return w.config.Organization + "/" + w.config.Project + "/" + w.config.Bucket + "/" + prefix + "/"
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
	if req == nil {
		return nil, ErrInvalidInput.Wrap("request cannot be nil")
	}
	if req.Reader == nil {
		return nil, ErrInvalidInput.Wrap("reader cannot be nil")
	}
	if req.Owner == "" {
		return nil, ErrInvalidInput.Wrap("owner cannot be empty")
	}
	if req.ArtifactType == "" {
		return nil, ErrInvalidInput.Wrap("artifact_type cannot be empty")
	}
	if req.EncryptionMetadata == nil {
		return nil, ErrInvalidInput.Wrap("encryption_metadata cannot be nil")
	}

	// Use fallback mode - read and store
	if w.useFallback {
		data, err := io.ReadAll(req.Reader)
		if err != nil {
			return nil, ErrInvalidInput.Wrapf("failed to read stream: %v", err)
		}

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

	// Generate a temporary reference for streaming upload
	tempRef := fmt.Sprintf("stream-%d", time.Now().UnixNano())
	backendRef := w.generateObjectPath(tempRef, req.Owner)

	// Check quota before streaming upload
	if w.config.QuotaBytes > 0 && req.Size > 0 {
		if err := w.objectStorage.CheckQuota(ctx, w.config.Bucket, req.Size); err != nil {
			if errors.Is(err, waldur.ErrQuotaExceeded) {
				return nil, ErrStorageLimitExceeded.Wrap(err.Error())
			}
		}
	}

	// Stream upload to Waldur object storage
	uploadReq := &waldur.UploadRequest{
		Bucket:      w.config.Bucket,
		Key:         backendRef,
		Body:        req.Reader,
		Size:        req.Size,
		ContentType: "application/octet-stream",
		Metadata: map[string]string{
			"owner":         req.Owner,
			"artifact_type": req.ArtifactType,
			"algorithm":     "sha256",
		},
	}

	uploadResp, err := w.objectStorage.UploadStream(ctx, uploadReq)
	if err != nil {
		return nil, ErrBackendUnavailable.Wrapf("streaming upload failed: %v", err)
	}

	// Parse content hash from response
	hashBytes, _ := hex.DecodeString(uploadResp.ContentHash)
	if len(hashBytes) != 32 {
		hashBytes = make([]byte, 32) // fallback empty hash
	}

	contentAddr := &ContentAddress{
		Version:    ContentAddressVersion,
		Hash:       hashBytes,
		Algorithm:  "sha256",
		Size:       uint64(uploadResp.Size),
		Backend:    BackendWaldur,
		BackendRef: uploadResp.Key,
	}

	artifactRef := NewArtifactReference(
		uploadResp.ContentHash,
		contentAddr,
		req.EncryptionMetadata,
		req.Owner,
		req.ArtifactType,
		0,
	)

	if req.RetentionTag != nil {
		artifactRef.SetRetentionTag(req.RetentionTag)
	}

	for k, v := range req.Metadata {
		artifactRef.SetMetadata(k, v)
	}

	// Update metrics
	w.mu.Lock()
	w.metrics.TotalArtifacts++
	w.metrics.TotalBytes += uint64(uploadResp.Size)
	w.ownerIndex[req.Owner] = append(w.ownerIndex[req.Owner], uploadResp.ContentHash)
	w.mu.Unlock()

	return &PutResponse{
		ContentAddress:    contentAddr,
		ArtifactReference: artifactRef,
	}, nil
}

// GetStream retrieves a large artifact as a stream
func (w *WaldurStreamingBackend) GetStream(ctx context.Context, address *ContentAddress) (io.ReadCloser, error) {
	if address == nil {
		return nil, ErrInvalidInput.Wrap("address cannot be nil")
	}

	// Use fallback mode
	if w.useFallback {
		resp, err := w.Get(ctx, &GetRequest{ContentAddress: address})
		if err != nil {
			return nil, err
		}
		return &bytesReadCloser{data: resp.Data}, nil
	}

	// Stream download from Waldur object storage
	downloadReq := &waldur.DownloadRequest{
		Bucket: w.config.Bucket,
		Key:    address.BackendRef,
	}

	downloadResp, err := w.objectStorage.Download(ctx, downloadReq)
	if err != nil {
		if errors.Is(err, waldur.ErrObjectNotFound) {
			return nil, ErrArtifactNotFound.Wrapf("artifact not found: %s", address.HashHex())
		}
		return nil, ErrBackendUnavailable.Wrapf("stream download failed: %v", err)
	}

	// Wrap with hash verification for integrity
	return &hashVerifyingReader{
		reader:       downloadResp.Body,
		expectedHash: address.Hash,
	}, nil
}

// hashVerifyingReader wraps a reader and verifies hash on close
type hashVerifyingReader struct {
	reader       io.ReadCloser
	expectedHash []byte
	hasher       io.Writer
	hashSum      func() []byte
}

func (h *hashVerifyingReader) Read(p []byte) (n int, err error) {
	if h.hasher == nil {
		hasher := sha256.New()
		h.hasher = hasher
		h.hashSum = func() []byte { return hasher.Sum(nil) }
	}

	n, err = h.reader.Read(p)
	if n > 0 {
		_, _ = h.hasher.Write(p[:n])
	}
	return n, err
}

func (h *hashVerifyingReader) Close() error {
	closeErr := h.reader.Close()

	// Verify hash if we have expected hash
	if len(h.expectedHash) > 0 && h.hashSum != nil {
		computedHash := h.hashSum()
		if !bytesEqual(computedHash, h.expectedHash) {
			return ErrHashMismatch.Wrap("downloaded data hash mismatch")
		}
	}

	return closeErr
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
