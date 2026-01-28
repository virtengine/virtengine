package artifact_store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"sync"
	"time"
)

// IPFSConfig contains configuration for the IPFS backend
type IPFSConfig struct {
	// Endpoint is the IPFS API endpoint URL (e.g., "localhost:5001")
	Endpoint string `json:"endpoint"`

	// GatewayURL is the IPFS HTTP gateway URL for retrieval
	GatewayURL string `json:"gateway_url"`

	// PinningService is the pinning service to use (optional)
	PinningService string `json:"pinning_service,omitempty"`

	// MaxRetries is the maximum number of retries for failed requests
	MaxRetries int `json:"max_retries"`

	// Timeout is the request timeout
	Timeout time.Duration `json:"timeout"`

	// ChunkSize is the default chunk size for large artifacts
	ChunkSize uint64 `json:"chunk_size"`

	// EnablePinning indicates if content should be pinned
	EnablePinning bool `json:"enable_pinning"`

	// UseStubClient indicates if stub client should be used (for testing)
	// WARNING: Stub client uses fake CIDs and in-memory storage.
	// NEVER use in production - data will be lost on restart.
	UseStubClient bool `json:"use_stub_client"`

	// ValidateCIDs enables CID validation for stored/retrieved content.
	// When true, stub CIDs are rejected. Should be true in production.
	ValidateCIDs bool `json:"validate_cids"`
}

// DefaultIPFSConfig returns a default configuration for testing with stub client.
// WARNING: This uses in-memory storage with fake CIDs. Use ProductionIPFSConfig for real deployments.
func DefaultIPFSConfig() *IPFSConfig {
	return &IPFSConfig{
		Endpoint:      "localhost:5001",
		GatewayURL:    "http://localhost:8080",
		MaxRetries:    3,
		Timeout:       60 * time.Second,
		ChunkSize:     256 * 1024, // 256KB chunks
		EnablePinning: true,
		UseStubClient: true,  // Default to stub for tests
		ValidateCIDs:  false, // Allow stub CIDs in test mode
	}
}

// ProductionIPFSConfig returns a configuration for production use with real IPFS.
// This configuration:
//   - Connects to a real IPFS node
//   - Enables CID validation (rejects fake stub CIDs)
//   - Enables pinning for persistence
//   - Uses sensible timeouts and retry policies
func ProductionIPFSConfig(endpoint string) *IPFSConfig {
	return &IPFSConfig{
		Endpoint:      endpoint,
		GatewayURL:    "https://ipfs.io",
		MaxRetries:    3,
		Timeout:       60 * time.Second,
		ChunkSize:     256 * 1024, // 256KB chunks
		EnablePinning: true,
		UseStubClient: false, // Use real IPFS client
		ValidateCIDs:  true,  // Validate CID format in production
	}
}

// Validate validates the configuration
func (c *IPFSConfig) Validate() error {
	if c.Endpoint == "" {
		return ErrInvalidInput.Wrap("ipfs endpoint cannot be empty")
	}
	if c.ChunkSize == 0 {
		return ErrInvalidInput.Wrap("chunk_size cannot be zero")
	}
	// Warn if using stub client with CID validation disabled in what looks like production
	if !c.UseStubClient && !c.ValidateCIDs {
		// This is a configuration that uses real IPFS but doesn't validate CIDs
		// It's allowed but unusual - production should validate CIDs
	}
	return nil
}

// IPFSBackend implements ArtifactStore and ChunkedArtifactStore using IPFS
// The implementation stores encrypted chunks in IPFS with on-chain CID references.
//
// Features:
//   - Content-addressed storage using IPFS CIDs
//   - Automatic chunking for large artifacts
//   - Chunk manifests for deterministic reconstruction
//   - Optional pinning for persistence
//   - Supports both real IPFS nodes and stub client for testing
//   - CID validation to prevent fake/stub CIDs in production
//
// Security:
//   - All data is encrypted before being stored in IPFS
//   - CIDs reference encrypted data only
//   - No sensitive data is logged
//   - CID validation prevents stub CID injection in production
type IPFSBackend struct {
	config       *IPFSConfig
	client       IPFSClient
	cidValidator *CIDValidator

	// mu protects concurrent access to the internal state
	mu sync.RWMutex

	// artifactIndex maps content hash (hex) -> artifact metadata
	// This is kept in-memory for fast lookups; production would use a database
	artifactIndex map[string]*ipfsStoredArtifact

	// chunkIndex maps chunk CID -> parent artifact hash
	chunkIndex map[string]string

	// metrics tracks storage metrics
	metrics *StorageMetrics

	// ownerIndex maps owner -> list of content hashes
	ownerIndex map[string][]string
}

// ipfsStoredArtifact represents artifact metadata in the index
type ipfsStoredArtifact struct {
	cid        string
	size       uint64
	reference  *ArtifactReference
	manifest   *ChunkManifest
	storedAt   time.Time
	accessedAt time.Time
	pinned     bool
}

// NewIPFSBackend creates a new IPFS backend instance
func NewIPFSBackend(config *IPFSConfig) (*IPFSBackend, error) {
	if config == nil {
		config = DefaultIPFSConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	var client IPFSClient
	var err error

	if config.UseStubClient {
		// Use stub client for testing
		client = NewStubIPFSClient()
	} else {
		// Create real IPFS client
		clientConfig := &RealIPFSClientConfig{
			Endpoint:   config.Endpoint,
			Timeout:    config.Timeout,
			MaxRetries: config.MaxRetries,
		}
		client, err = NewRealIPFSClient(clientConfig)
		if err != nil {
			return nil, ErrBackendUnavailable.Wrapf("failed to connect to IPFS: %v", err)
		}
	}

	// Create CID validator based on configuration
	var cidValidator *CIDValidator
	if config.ValidateCIDs {
		cidValidator = NewCIDValidator()
	} else {
		cidValidator = NewTestCIDValidator()
	}

	return &IPFSBackend{
		config:        config,
		client:        client,
		cidValidator:  cidValidator,
		artifactIndex: make(map[string]*ipfsStoredArtifact),
		chunkIndex:    make(map[string]string),
		ownerIndex:    make(map[string][]string),
		metrics: &StorageMetrics{
			BackendType:   BackendIPFS,
			BackendStatus: make(map[string]string),
		},
	}, nil
}

// NewIPFSBackendWithClient creates a new IPFS backend with a custom client
// This is useful for testing with mock clients
func NewIPFSBackendWithClient(config *IPFSConfig, client IPFSClient) (*IPFSBackend, error) {
	if config == nil {
		config = DefaultIPFSConfig()
	}

	if client == nil {
		return nil, ErrInvalidInput.Wrap("client cannot be nil")
	}

	// Create CID validator based on configuration
	var cidValidator *CIDValidator
	if config.ValidateCIDs {
		cidValidator = NewCIDValidator()
	} else {
		cidValidator = NewTestCIDValidator()
	}

	return &IPFSBackend{
		config:        config,
		client:        client,
		cidValidator:  cidValidator,
		artifactIndex: make(map[string]*ipfsStoredArtifact),
		chunkIndex:    make(map[string]string),
		ownerIndex:    make(map[string][]string),
		metrics: &StorageMetrics{
			BackendType:   BackendIPFS,
			BackendStatus: make(map[string]string),
		},
	}, nil
}

// Backend returns the backend type
func (i *IPFSBackend) Backend() BackendType {
	return BackendIPFS
}

// Put stores an encrypted artifact and returns its content address
func (i *IPFSBackend) Put(ctx context.Context, req *PutRequest) (*PutResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Store data in IPFS and get real CID
	cid, err := i.client.Add(ctx, req.Data)
	if err != nil {
		return nil, ErrBackendUnavailable.Wrapf("ipfs add failed: %v", err)
	}

	// Validate the CID returned from IPFS
	if err := i.cidValidator.ValidateCID(cid); err != nil {
		return nil, err
	}

	// Pin if enabled
	if i.config.EnablePinning {
		if err := i.client.Pin(ctx, cid); err != nil {
			// Log warning but don't fail - data is already stored
			// In production, this would be logged properly
			_ = err
		}
	}

	// Compute content address
	hash := sha256.Sum256(req.Data)
	hashHex := hex.EncodeToString(hash[:])

	contentAddr := &ContentAddress{
		Version:    ContentAddressVersion,
		Hash:       hash[:],
		Algorithm:  "sha256",
		Size:       uint64(len(req.Data)),
		Backend:    BackendIPFS,
		BackendRef: cid,
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

	// Store artifact metadata in index
	i.mu.Lock()
	defer i.mu.Unlock()

	i.artifactIndex[hashHex] = &ipfsStoredArtifact{
		cid:        cid,
		size:       uint64(len(req.Data)),
		reference:  artifactRef,
		storedAt:   time.Now().UTC(),
		accessedAt: time.Now().UTC(),
		pinned:     i.config.EnablePinning,
	}

	// Update owner index
	i.ownerIndex[req.Owner] = append(i.ownerIndex[req.Owner], hashHex)

	// Update metrics
	i.metrics.TotalArtifacts++
	i.metrics.TotalBytes += uint64(len(req.Data))

	return &PutResponse{
		ContentAddress:    contentAddr,
		ArtifactReference: artifactRef,
	}, nil
}

// Get retrieves an encrypted artifact by its content address
func (i *IPFSBackend) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	hashHex := hex.EncodeToString(req.ContentAddress.Hash)

	i.mu.RLock()
	artifact, exists := i.artifactIndex[hashHex]
	i.mu.RUnlock()

	if !exists {
		return nil, ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
	}

	// Check retention
	if artifact.reference.RetentionTag != nil {
		if artifact.reference.RetentionTag.IsExpired(time.Now().UTC()) {
			return nil, ErrRetentionExpired.Wrap("artifact retention has expired")
		}
	}

	// Retrieve data from IPFS using the CID
	data, err := i.client.Cat(ctx, artifact.cid)
	if err != nil {
		return nil, ErrBackendUnavailable.Wrapf("ipfs cat failed: %v", err)
	}

	// Verify hash
	computedHash := sha256.Sum256(data)
	if !bytesEqual(computedHash[:], req.ContentAddress.Hash) {
		return nil, ErrHashMismatch.Wrap("retrieved data hash does not match request")
	}

	// Update access time
	i.mu.Lock()
	artifact.accessedAt = time.Now().UTC()
	i.mu.Unlock()

	return &GetResponse{
		Data:           data,
		ContentAddress: artifact.reference.ContentAddress,
		ChunkManifest:  artifact.manifest,
	}, nil
}

// Delete removes an artifact by its content address
// In IPFS, this unpins the data allowing garbage collection
func (i *IPFSBackend) Delete(ctx context.Context, req *DeleteRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	hashHex := hex.EncodeToString(req.ContentAddress.Hash)

	i.mu.Lock()
	defer i.mu.Unlock()

	artifact, exists := i.artifactIndex[hashHex]
	if !exists {
		return ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
	}

	// Check authorization
	if artifact.reference.AccountAddress != req.RequestingAccount && !req.Force {
		return ErrAuthenticationFailed.Wrap("not authorized to delete artifact")
	}

	// Unpin from IPFS (allows garbage collection)
	if artifact.pinned {
		if err := i.client.Unpin(ctx, artifact.cid); err != nil {
			// Log warning but continue with deletion from index
			_ = err
		}
	}

	// Delete chunks if present
	if artifact.manifest != nil {
		for _, chunk := range artifact.manifest.Chunks {
			if err := i.client.Unpin(ctx, chunk.BackendRef); err != nil {
				_ = err
			}
			delete(i.chunkIndex, chunk.BackendRef)
			i.metrics.TotalChunks--
		}
	}

	// Remove from owner index
	if hashes, ok := i.ownerIndex[artifact.reference.AccountAddress]; ok {
		newHashes := make([]string, 0, len(hashes)-1)
		for _, h := range hashes {
			if h != hashHex {
				newHashes = append(newHashes, h)
			}
		}
		i.ownerIndex[artifact.reference.AccountAddress] = newHashes
	}

	// Update metrics
	i.metrics.TotalArtifacts--
	i.metrics.TotalBytes -= artifact.size

	// Delete from index
	delete(i.artifactIndex, hashHex)

	return nil
}

// Exists checks if an artifact exists at the given content address
func (i *IPFSBackend) Exists(ctx context.Context, address *ContentAddress) (bool, error) {
	if address == nil {
		return false, ErrInvalidInput.Wrap("address cannot be nil")
	}

	hashHex := hex.EncodeToString(address.Hash)

	i.mu.RLock()
	artifact, exists := i.artifactIndex[hashHex]
	i.mu.RUnlock()

	if !exists {
		return false, nil
	}

	// Optionally verify the CID is still pinned in IPFS
	if i.config.EnablePinning {
		pinned, err := i.client.IsPinned(ctx, artifact.cid)
		if err != nil {
			// If we can't check, assume it exists based on index
			return true, nil
		}
		return pinned, nil
	}

	return true, nil
}

// GetChunk retrieves a specific chunk of a chunked artifact
func (i *IPFSBackend) GetChunk(ctx context.Context, address *ContentAddress, chunkIndex uint32) (*ChunkData, error) {
	if address == nil {
		return nil, ErrInvalidInput.Wrap("address cannot be nil")
	}

	hashHex := hex.EncodeToString(address.Hash)

	i.mu.RLock()
	artifact, exists := i.artifactIndex[hashHex]
	i.mu.RUnlock()

	if !exists {
		return nil, ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
	}

	if artifact.manifest == nil {
		return nil, ErrInvalidInput.Wrap("artifact is not chunked")
	}

	if chunkIndex >= artifact.manifest.ChunkCount {
		return nil, ErrChunkNotFound.Wrapf("chunk index %d out of range (max: %d)", chunkIndex, artifact.manifest.ChunkCount-1)
	}

	chunkInfo := artifact.manifest.Chunks[chunkIndex]

	// Retrieve chunk from IPFS
	chunkData, err := i.client.Cat(ctx, chunkInfo.BackendRef)
	if err != nil {
		return nil, ErrChunkNotFound.Wrapf("failed to retrieve chunk: %v", err)
	}

	return &ChunkData{
		Index: chunkIndex,
		Data:  chunkData,
		Hash:  chunkInfo.Hash,
	}, nil
}

// ListByOwner lists all artifacts owned by a specific account
func (i *IPFSBackend) ListByOwner(ctx context.Context, owner string, pagination *Pagination) (*ListResponse, error) {
	if owner == "" {
		return nil, ErrInvalidInput.Wrap("owner cannot be empty")
	}

	i.mu.RLock()
	defer i.mu.RUnlock()

	hashes, exists := i.ownerIndex[owner]
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
		if artifact, ok := i.artifactIndex[h]; ok {
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
func (i *IPFSBackend) UpdateRetention(ctx context.Context, address *ContentAddress, tag *RetentionTag) error {
	if address == nil {
		return ErrInvalidInput.Wrap("address cannot be nil")
	}

	hashHex := hex.EncodeToString(address.Hash)

	i.mu.Lock()
	defer i.mu.Unlock()

	artifact, exists := i.artifactIndex[hashHex]
	if !exists {
		return ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
	}

	artifact.reference.SetRetentionTag(tag)
	return nil
}

// PurgeExpired removes all expired artifacts based on the current time/block
func (i *IPFSBackend) PurgeExpired(ctx context.Context, currentBlock int64) (int, error) {
	now := time.Now().UTC()

	i.mu.Lock()
	defer i.mu.Unlock()

	toDelete := make([]string, 0)

	for hashHex, artifact := range i.artifactIndex {
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
		artifact := i.artifactIndex[hashHex]

		// Unpin from IPFS
		if artifact.pinned {
			if err := i.client.Unpin(ctx, artifact.cid); err != nil {
				_ = err // Log but continue
			}
		}

		// Delete chunks if present
		if artifact.manifest != nil {
			for _, chunk := range artifact.manifest.Chunks {
				if err := i.client.Unpin(ctx, chunk.BackendRef); err != nil {
					_ = err
				}
				delete(i.chunkIndex, chunk.BackendRef)
				i.metrics.TotalChunks--
			}
		}

		// Remove from owner index
		if hashes, ok := i.ownerIndex[artifact.reference.AccountAddress]; ok {
			newHashes := make([]string, 0, len(hashes)-1)
			for _, h := range hashes {
				if h != hashHex {
					newHashes = append(newHashes, h)
				}
			}
			i.ownerIndex[artifact.reference.AccountAddress] = newHashes
		}

		// Update metrics
		i.metrics.TotalBytes -= artifact.size

		delete(i.artifactIndex, hashHex)
	}

	i.metrics.TotalArtifacts -= uint64(len(toDelete))
	return len(toDelete), nil
}

// GetMetrics returns storage metrics for monitoring
func (i *IPFSBackend) GetMetrics(ctx context.Context) (*StorageMetrics, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Count expired artifacts
	now := time.Now().UTC()
	expired := uint64(0)
	for _, artifact := range i.artifactIndex {
		if artifact.reference.RetentionTag != nil {
			if artifact.reference.RetentionTag.IsExpired(now) {
				expired++
			}
		}
	}

	// Get IPFS version for status
	version, _ := i.client.Version(ctx)

	return &StorageMetrics{
		TotalArtifacts:   i.metrics.TotalArtifacts,
		TotalBytes:       i.metrics.TotalBytes,
		TotalChunks:      i.metrics.TotalChunks,
		ExpiredArtifacts: expired,
		BackendType:      BackendIPFS,
		BackendStatus: map[string]string{
			"endpoint":     i.config.Endpoint,
			"gateway":      i.config.GatewayURL,
			"pinning":      boolToString(i.config.EnablePinning),
			"chunk_size":   uintToString(i.config.ChunkSize),
			"total_cids":   uintToString(i.metrics.TotalArtifacts),
			"total_pins":   uintToString(i.countPinned()),
			"ipfs_version": version,
		},
	}, nil
}

// countPinned counts pinned artifacts
func (i *IPFSBackend) countPinned() uint64 {
	count := uint64(0)
	for _, artifact := range i.artifactIndex {
		if artifact.pinned {
			count++
		}
	}
	return count
}

// Health checks if the backend is healthy
func (i *IPFSBackend) Health(ctx context.Context) error {
	return i.client.IsHealthy(ctx)
}

// Ensure IPFSBackend implements ArtifactStore
var _ ArtifactStore = (*IPFSBackend)(nil)

// PutChunked stores an artifact in chunks according to the manifest
func (i *IPFSBackend) PutChunked(ctx context.Context, data []byte, chunkSize uint64, meta *PutMetadata) (*PutResponse, *ChunkManifest, error) {
	if len(data) == 0 {
		return nil, nil, ErrInvalidInput.Wrap("data cannot be empty")
	}
	if chunkSize == 0 {
		chunkSize = i.config.ChunkSize
	}
	if meta == nil {
		return nil, nil, ErrInvalidInput.Wrap("metadata cannot be nil")
	}

	totalSize := uint64(len(data))
	manifest := NewChunkManifest(totalSize, chunkSize)

	// Split data into chunks and store each in IPFS
	for offset := uint64(0); offset < totalSize; offset += chunkSize {
		end := offset + chunkSize
		if end > totalSize {
			end = totalSize
		}

		chunkData := data[offset:end]
		chunkHash := sha256.Sum256(chunkData)

		// Store chunk in IPFS
		chunkCID, err := i.client.Add(ctx, chunkData)
		if err != nil {
			return nil, nil, ErrBackendUnavailable.Wrapf("failed to store chunk: %v", err)
		}

		// Validate the chunk CID
		if err := i.cidValidator.ValidateCID(chunkCID); err != nil {
			return nil, nil, err
		}

		// Pin chunk if enabled
		if i.config.EnablePinning {
			if err := i.client.Pin(ctx, chunkCID); err != nil {
				_ = err // Log but continue
			}
		}

		// Add to manifest
		chunkInfo := ChunkInfo{
			//nolint:gosec // G115: offset/chunkSize bounded by chunk count
			Index:      uint32(offset / chunkSize),
			Hash:       chunkHash[:],
			Size:       uint64(len(chunkData)),
			Offset:     offset,
			BackendRef: chunkCID,
		}
		if err := manifest.AddChunk(chunkInfo); err != nil {
			return nil, nil, err
		}

		// Track chunk in index
		i.mu.Lock()
		i.chunkIndex[chunkCID] = hex.EncodeToString(chunkHash[:])
		i.metrics.TotalChunks++
		i.mu.Unlock()
	}

	// Compute root hash
	manifest.ComputeRootHash()

	// Create content address for the whole artifact
	hash := sha256.Sum256(data)
	hashHex := hex.EncodeToString(hash[:])

	// Store manifest in IPFS and get its CID
	manifestData := manifest.Serialize()
	manifestCID, err := i.client.Add(ctx, manifestData)
	if err != nil {
		return nil, nil, ErrBackendUnavailable.Wrapf("failed to store manifest: %v", err)
	}

	// Validate the manifest CID
	if err := i.cidValidator.ValidateCID(manifestCID); err != nil {
		return nil, nil, err
	}

	if i.config.EnablePinning {
		if err := i.client.Pin(ctx, manifestCID); err != nil {
			_ = err
		}
	}

	contentAddr := &ContentAddress{
		Version:    ContentAddressVersion,
		Hash:       hash[:],
		Algorithm:  "sha256",
		Size:       totalSize,
		Backend:    BackendIPFS,
		BackendRef: manifestCID,
	}

	// Create artifact reference
	artifactRef := NewArtifactReference(
		hashHex,
		contentAddr,
		meta.EncryptionMetadata,
		meta.Owner,
		meta.ArtifactType,
		0, // Block height is set by caller
	)

	artifactRef.SetChunkManifest(manifest)

	if meta.RetentionTag != nil {
		artifactRef.SetRetentionTag(meta.RetentionTag)
	}

	for k, v := range meta.Metadata {
		artifactRef.SetMetadata(k, v)
	}

	// Store artifact metadata in index
	i.mu.Lock()
	defer i.mu.Unlock()

	i.artifactIndex[hashHex] = &ipfsStoredArtifact{
		cid:        manifestCID,
		size:       totalSize,
		reference:  artifactRef,
		manifest:   manifest,
		storedAt:   time.Now().UTC(),
		accessedAt: time.Now().UTC(),
		pinned:     i.config.EnablePinning,
	}

	// Update owner index
	i.ownerIndex[meta.Owner] = append(i.ownerIndex[meta.Owner], hashHex)

	// Update metrics
	i.metrics.TotalArtifacts++
	i.metrics.TotalBytes += totalSize

	return &PutResponse{
		ContentAddress:    contentAddr,
		ArtifactReference: artifactRef,
	}, manifest, nil
}

// GetChunked retrieves and reassembles a chunked artifact
func (i *IPFSBackend) GetChunked(ctx context.Context, manifest *ChunkManifest) ([]byte, error) {
	if manifest == nil {
		return nil, ErrInvalidInput.Wrap("manifest cannot be nil")
	}

	if err := manifest.Validate(); err != nil {
		return nil, err
	}

	// Reassemble data from chunks in order
	data := make([]byte, 0, manifest.TotalSize)

	for _, chunkInfo := range manifest.Chunks {
		// Retrieve chunk from IPFS
		chunkData, err := i.client.Cat(ctx, chunkInfo.BackendRef)
		if err != nil {
			return nil, ErrChunkNotFound.Wrapf("failed to retrieve chunk %s: %v", chunkInfo.BackendRef, err)
		}

		// Verify chunk hash
		computedHash := sha256.Sum256(chunkData)
		if !bytesEqual(computedHash[:], chunkInfo.Hash) {
			return nil, ErrHashMismatch.Wrapf("chunk %d hash verification failed", chunkInfo.Index)
		}

		data = append(data, chunkData...)
	}

	if uint64(len(data)) != manifest.TotalSize {
		return nil, ErrChunkReassemblyFailed.Wrapf("reassembled size %d doesn't match expected %d", len(data), manifest.TotalSize)
	}

	return data, nil
}

// VerifyChunks verifies all chunks match the manifest hashes
func (i *IPFSBackend) VerifyChunks(ctx context.Context, manifest *ChunkManifest) error {
	if manifest == nil {
		return ErrInvalidInput.Wrap("manifest cannot be nil")
	}

	for _, chunkInfo := range manifest.Chunks {
		// Retrieve chunk from IPFS
		chunkData, err := i.client.Cat(ctx, chunkInfo.BackendRef)
		if err != nil {
			return ErrChunkNotFound.Wrapf("failed to retrieve chunk %s: %v", chunkInfo.BackendRef, err)
		}

		computedHash := sha256.Sum256(chunkData)
		if !bytesEqual(computedHash[:], chunkInfo.Hash) {
			return ErrHashMismatch.Wrapf("chunk %d hash verification failed", chunkInfo.Index)
		}
	}

	return nil
}

// Ensure IPFSBackend implements ChunkedArtifactStore
var _ ChunkedArtifactStore = (*IPFSBackend)(nil)

// uintToString converts uint64 to string
func uintToString(n uint64) string {
	return hex.EncodeToString([]byte{
		byte(n >> 56), byte(n >> 48), byte(n >> 40), byte(n >> 32),
		byte(n >> 24), byte(n >> 16), byte(n >> 8), byte(n),
	})
}

// IPFSStreamingBackend extends IPFSBackend with streaming support
type IPFSStreamingBackend struct {
	*IPFSBackend
}

// NewIPFSStreamingBackend creates a new IPFS streaming backend
func NewIPFSStreamingBackend(config *IPFSConfig) (*IPFSStreamingBackend, error) {
	base, err := NewIPFSBackend(config)
	if err != nil {
		return nil, err
	}
	return &IPFSStreamingBackend{IPFSBackend: base}, nil
}

// PutStream stores a large artifact using streaming
func (i *IPFSStreamingBackend) PutStream(ctx context.Context, req *PutStreamRequest) (*PutResponse, error) {
	// Read all data from stream
	// Note: For very large files, consider implementing true streaming to IPFS
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

	return i.Put(ctx, putReq)
}

// GetStream retrieves a large artifact as a stream
func (i *IPFSStreamingBackend) GetStream(ctx context.Context, address *ContentAddress) (io.ReadCloser, error) {
	if address == nil {
		return nil, ErrInvalidInput.Wrap("address cannot be nil")
	}

	hashHex := hex.EncodeToString(address.Hash)

	i.mu.RLock()
	artifact, exists := i.artifactIndex[hashHex]
	i.mu.RUnlock()

	if !exists {
		return nil, ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
	}

	// Use the client's streaming interface
	return i.client.CatStream(ctx, artifact.cid)
}

// Ensure IPFSStreamingBackend implements StreamingArtifactStore
var _ StreamingArtifactStore = (*IPFSStreamingBackend)(nil)
