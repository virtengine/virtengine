package artifact_store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

// MemoryBackend stores artifacts in memory for development and testing.
type MemoryBackend struct {
	mu        sync.RWMutex
	artifacts map[string][]byte
	meta      map[string]*ArtifactReference
	byOwner   map[string][]string
}

// NewMemoryBackend creates a new memory backend.
func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		artifacts: make(map[string][]byte),
		meta:      make(map[string]*ArtifactReference),
		byOwner:   make(map[string][]string),
	}
}

// Put stores an encrypted artifact.
func (m *MemoryBackend) Put(_ context.Context, req *PutRequest) (*PutResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	hash := req.ContentHash
	if len(hash) == 0 {
		sum := sha256.Sum256(req.Data)
		hash = sum[:]
	}
	backendRef := hex.EncodeToString(hash)

	address := NewContentAddressFromHash(hash, uint64(len(req.Data)), BackendIPFS, backendRef)
	if err := address.Validate(); err != nil {
		return nil, err
	}

	ref := NewArtifactReference(backendRef, address, req.EncryptionMetadata, req.Owner, req.ArtifactType, 0)
	ref.RetentionTag = req.RetentionTag
	ref.Metadata = req.Metadata

	m.mu.Lock()
	defer m.mu.Unlock()
	m.artifacts[backendRef] = append([]byte(nil), req.Data...)
	m.meta[backendRef] = ref
	m.byOwner[req.Owner] = append(m.byOwner[req.Owner], backendRef)

	return &PutResponse{
		ContentAddress:    address,
		ArtifactReference: ref,
	}, nil
}

// Get retrieves an artifact.
func (m *MemoryBackend) Get(_ context.Context, req *GetRequest) (*GetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	ref := req.ContentAddress.BackendRef

	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.artifacts[ref]
	if !ok {
		return nil, ErrArtifactNotFound
	}

	return &GetResponse{
		Data:           append([]byte(nil), data...),
		ContentAddress: req.ContentAddress,
	}, nil
}

// Delete removes an artifact.
func (m *MemoryBackend) Delete(_ context.Context, req *DeleteRequest) error {
	if req == nil || req.ContentAddress == nil {
		return ErrInvalidInput.Wrap("content address required")
	}
	ref := req.ContentAddress.BackendRef

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.artifacts[ref]; !ok {
		return ErrArtifactNotFound
	}
	delete(m.artifacts, ref)
	delete(m.meta, ref)
	for owner, refs := range m.byOwner {
		filtered := refs[:0]
		for _, item := range refs {
			if item != ref {
				filtered = append(filtered, item)
			}
		}
		if len(filtered) == 0 {
			delete(m.byOwner, owner)
		} else {
			m.byOwner[owner] = filtered
		}
	}
	return nil
}

// Exists checks if an artifact exists.
func (m *MemoryBackend) Exists(_ context.Context, address *ContentAddress) (bool, error) {
	if address == nil {
		return false, ErrInvalidInput.Wrap("content address required")
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.artifacts[address.BackendRef]
	return ok, nil
}

// GetChunk is unsupported for the memory backend.
func (m *MemoryBackend) GetChunk(_ context.Context, _ *ContentAddress, _ uint32) (*ChunkData, error) {
	return nil, ErrBackendNotSupported
}

// ListByOwner lists artifacts by owner.
func (m *MemoryBackend) ListByOwner(_ context.Context, owner string, pagination *Pagination) (*ListResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	refs := m.byOwner[owner]
	total := uint64(len(refs))
	if total == 0 {
		return &ListResponse{References: []*ArtifactReference{}, Total: 0, HasMore: false}, nil
	}

	offset := uint64(0)
	limit := total
	if pagination != nil {
		offset = pagination.Offset
		if pagination.Limit > 0 {
			limit = pagination.Limit
		}
	}

	if offset >= total {
		return &ListResponse{References: []*ArtifactReference{}, Total: total, HasMore: false}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	result := make([]*ArtifactReference, 0, end-offset)
	for _, refID := range refs[offset:end] {
		if ref := m.meta[refID]; ref != nil {
			result = append(result, ref)
		}
	}

	return &ListResponse{
		References: result,
		Total:      total,
		HasMore:    end < total,
	}, nil
}

// UpdateRetention updates the retention tag for an artifact.
func (m *MemoryBackend) UpdateRetention(_ context.Context, address *ContentAddress, tag *RetentionTag) error {
	if address == nil {
		return ErrInvalidInput.Wrap("content address required")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	ref := m.meta[address.BackendRef]
	if ref == nil {
		return ErrArtifactNotFound
	}
	ref.RetentionTag = tag
	return nil
}

// PurgeExpired removes expired artifacts.
func (m *MemoryBackend) PurgeExpired(_ context.Context, currentBlock int64) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	removed := 0
	for refID, ref := range m.meta {
		if ref.RetentionTag == nil || ref.RetentionTag.ExpiresAt == nil {
			continue
		}
		if ref.RetentionTag.ExpiresAt.Before(time.Now().UTC()) {
			delete(m.meta, refID)
			delete(m.artifacts, refID)
			removed++
		}
	}
	return removed, nil
}

// GetMetrics returns storage metrics.
func (m *MemoryBackend) GetMetrics(_ context.Context) (*StorageMetrics, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	totalBytes := uint64(0)
	for _, data := range m.artifacts {
		totalBytes += uint64(len(data))
	}
	return &StorageMetrics{
		TotalArtifacts:   uint64(len(m.artifacts)),
		TotalBytes:       totalBytes,
		TotalChunks:      0,
		ExpiredArtifacts: 0,
		BackendType:      BackendIPFS,
		BackendStatus:    map[string]string{"mode": "memory"},
	}, nil
}

// Health returns backend health.
func (m *MemoryBackend) Health(_ context.Context) error {
	return nil
}

// Backend returns the backend type.
func (m *MemoryBackend) Backend() BackendType {
	return BackendIPFS
}
