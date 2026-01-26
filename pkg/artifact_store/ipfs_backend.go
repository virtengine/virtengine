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
	// Endpoint is the IPFS API endpoint URL
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
}

// DefaultIPFSConfig returns a default configuration
func DefaultIPFSConfig() *IPFSConfig {
	return &IPFSConfig{
		Endpoint:      "http://localhost:5001",
		GatewayURL:    "http://localhost:8080",
		MaxRetries:    3,
		Timeout:       60 * time.Second,
		ChunkSize:     256 * 1024, // 256KB chunks
		EnablePinning: true,
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
//
// Security:
//   - All data is encrypted before being stored in IPFS
//   - CIDs reference encrypted data only
//   - No sensitive data is logged
type IPFSBackend struct {
	config *IPFSConfig

	// mu protects concurrent access to the storage map
	mu sync.RWMutex

	// storage is an in-memory map for stubbed implementation
	// In production, this would be replaced with actual IPFS API calls
	storage map[string]*ipfsStoredArtifact

	// chunks stores individual chunks by CID
	chunks map[string][]byte

	// metrics tracks storage metrics
	metrics *StorageMetrics

	// ownerIndex maps owner -> list of content hashes
	ownerIndex map[string][]string

	// cidCounter is used to generate unique CIDs in stubbed mode
	cidCounter uint64
}

// ipfsStoredArtifact represents an artifact in IPFS storage
type ipfsStoredArtifact struct {
	cid        string
	data       []byte
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

	// For stubbed implementation, we don't validate endpoint
	// In production, this would establish connection to IPFS node

	return &IPFSBackend{
		config:     config,
		storage:    make(map[string]*ipfsStoredArtifact),
		chunks:     make(map[string][]byte),
		ownerIndex: make(map[string][]string),
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

// generateCID generates a fake CID for stubbed implementation
// In production, this would be computed by IPFS
func (i *IPFSBackend) generateCID(data []byte) string {
	i.cidCounter++
	hash := sha256.Sum256(data)
	// Simulate CIDv1 format: Qm + base58 of hash
	return "Qm" + hex.EncodeToString(hash[:16])
}

// Put stores an encrypted artifact and returns its content address
func (i *IPFSBackend) Put(ctx context.Context, req *PutRequest) (*PutResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Generate CID for the data
	cid := i.generateCID(req.Data)

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

	// Store in backend (stubbed - in production would call IPFS API)
	i.mu.Lock()
	defer i.mu.Unlock()

	i.storage[hashHex] = &ipfsStoredArtifact{
		cid:        cid,
		data:       req.Data,
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
	artifact, exists := i.storage[hashHex]
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

	// Verify hash
	computedHash := sha256.Sum256(artifact.data)
	if !bytesEqual(computedHash[:], req.ContentAddress.Hash) {
		return nil, ErrHashMismatch.Wrap("stored data hash does not match request")
	}

	// Update access time
	i.mu.Lock()
	artifact.accessedAt = time.Now().UTC()
	i.mu.Unlock()

	return &GetResponse{
		Data:           artifact.data,
		ContentAddress: artifact.reference.ContentAddress,
		ChunkManifest:  artifact.manifest,
	}, nil
}

// Delete removes an artifact by its content address
func (i *IPFSBackend) Delete(ctx context.Context, req *DeleteRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	hashHex := hex.EncodeToString(req.ContentAddress.Hash)

	i.mu.Lock()
	defer i.mu.Unlock()

	artifact, exists := i.storage[hashHex]
	if !exists {
		return ErrArtifactNotFound.Wrapf("artifact not found: %s", hashHex)
	}

	// Check authorization
	if artifact.reference.AccountAddress != req.RequestingAccount && !req.Force {
		return ErrAuthenticationFailed.Wrap("not authorized to delete artifact")
	}

	// Delete chunks if present
	if artifact.manifest != nil {
		for _, chunk := range artifact.manifest.Chunks {
			delete(i.chunks, chunk.BackendRef)
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
	i.metrics.TotalBytes -= uint64(len(artifact.data))

	// Delete from storage
	delete(i.storage, hashHex)

	return nil
}

// Exists checks if an artifact exists at the given content address
func (i *IPFSBackend) Exists(ctx context.Context, address *ContentAddress) (bool, error) {
	if address == nil {
		return false, ErrInvalidInput.Wrap("address cannot be nil")
	}

	hashHex := hex.EncodeToString(address.Hash)

	i.mu.RLock()
	_, exists := i.storage[hashHex]
	i.mu.RUnlock()

	return exists, nil
}

// GetChunk retrieves a specific chunk of a chunked artifact
func (i *IPFSBackend) GetChunk(ctx context.Context, address *ContentAddress, chunkIndex uint32) (*ChunkData, error) {
	if address == nil {
		return nil, ErrInvalidInput.Wrap("address cannot be nil")
	}

	hashHex := hex.EncodeToString(address.Hash)

	i.mu.RLock()
	artifact, exists := i.storage[hashHex]
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

	i.mu.RLock()
	chunkData, exists := i.chunks[chunkInfo.BackendRef]
	i.mu.RUnlock()

	if !exists {
		return nil, ErrChunkNotFound.Wrapf("chunk not found: %s", chunkInfo.BackendRef)
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
		if artifact, ok := i.storage[h]; ok {
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

	artifact, exists := i.storage[hashHex]
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

	for hashHex, artifact := range i.storage {
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
		artifact := i.storage[hashHex]

		// Delete chunks if present
		if artifact.manifest != nil {
			for _, chunk := range artifact.manifest.Chunks {
				delete(i.chunks, chunk.BackendRef)
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
		i.metrics.TotalBytes -= uint64(len(artifact.data))

		delete(i.storage, hashHex)
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
	for _, artifact := range i.storage {
		if artifact.reference.RetentionTag != nil {
			if artifact.reference.RetentionTag.IsExpired(now) {
				expired++
			}
		}
	}

	return &StorageMetrics{
		TotalArtifacts:   i.metrics.TotalArtifacts,
		TotalBytes:       i.metrics.TotalBytes,
		TotalChunks:      i.metrics.TotalChunks,
		ExpiredArtifacts: expired,
		BackendType:      BackendIPFS,
		BackendStatus: map[string]string{
			"endpoint":    i.config.Endpoint,
			"gateway":     i.config.GatewayURL,
			"pinning":     boolToString(i.config.EnablePinning),
			"chunk_size":  uintToString(i.config.ChunkSize),
			"total_cids":  uintToString(i.metrics.TotalArtifacts),
			"total_pins":  uintToString(i.countPinned()),
		},
	}, nil
}

// countPinned counts pinned artifacts
func (i *IPFSBackend) countPinned() uint64 {
	count := uint64(0)
	for _, artifact := range i.storage {
		if artifact.pinned {
			count++
		}
	}
	return count
}

// Health checks if the backend is healthy
func (i *IPFSBackend) Health(ctx context.Context) error {
	// Stubbed - in production would ping IPFS node
	return nil
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

	i.mu.Lock()
	defer i.mu.Unlock()

	// Split data into chunks and store each
	for offset := uint64(0); offset < totalSize; offset += chunkSize {
		end := offset + chunkSize
		if end > totalSize {
			end = totalSize
		}

		chunkData := data[offset:end]
		chunkHash := sha256.Sum256(chunkData)
		chunkCID := i.generateCID(chunkData)

		// Store chunk
		i.chunks[chunkCID] = chunkData
		i.metrics.TotalChunks++

		// Add to manifest
		chunkInfo := ChunkInfo{
			Index:      uint32(offset / chunkSize),
			Hash:       chunkHash[:],
			Size:       uint64(len(chunkData)),
			Offset:     offset,
			BackendRef: chunkCID,
		}
		if err := manifest.AddChunk(chunkInfo); err != nil {
			return nil, nil, err
		}
	}

	// Compute root hash
	manifest.ComputeRootHash()

	// Create content address for the whole artifact
	hash := sha256.Sum256(data)
	hashHex := hex.EncodeToString(hash[:])

	// Generate CID for the manifest itself
	manifestCID := i.generateCID(manifest.RootHash)

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

	// Store artifact
	i.storage[hashHex] = &ipfsStoredArtifact{
		cid:        manifestCID,
		data:       data,
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

	i.mu.RLock()
	defer i.mu.RUnlock()

	// Reassemble data from chunks in order
	data := make([]byte, 0, manifest.TotalSize)

	for _, chunkInfo := range manifest.Chunks {
		chunkData, exists := i.chunks[chunkInfo.BackendRef]
		if !exists {
			return nil, ErrChunkNotFound.Wrapf("chunk not found: %s", chunkInfo.BackendRef)
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

	i.mu.RLock()
	defer i.mu.RUnlock()

	for _, chunkInfo := range manifest.Chunks {
		chunkData, exists := i.chunks[chunkInfo.BackendRef]
		if !exists {
			return ErrChunkNotFound.Wrapf("chunk not found: %s", chunkInfo.BackendRef)
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
	// Read all data from stream (stubbed - production would stream to IPFS)
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
	// Get artifact (stubbed - production would stream from IPFS gateway)
	resp, err := i.Get(ctx, &GetRequest{ContentAddress: address})
	if err != nil {
		return nil, err
	}

	return &bytesReadCloser{data: resp.Data}, nil
}

// Ensure IPFSStreamingBackend implements StreamingArtifactStore
var _ StreamingArtifactStore = (*IPFSStreamingBackend)(nil)
