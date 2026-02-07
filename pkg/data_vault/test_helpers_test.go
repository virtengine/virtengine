package data_vault

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/virtengine/virtengine/pkg/artifact_store"
)

type memoryArtifactStore struct {
	mu      sync.RWMutex
	backend artifact_store.BackendType
	objects map[string][]byte
	owners  map[string][]string
}

var _ artifact_store.ArtifactStore = (*memoryArtifactStore)(nil)
var _ = newMemoryArtifactStore

func newMemoryArtifactStore() *memoryArtifactStore {
	return &memoryArtifactStore{
		backend: artifact_store.BackendWaldur,
		objects: make(map[string][]byte),
		owners:  make(map[string][]string),
	}
}

func (m *memoryArtifactStore) Put(_ context.Context, req *artifact_store.PutRequest) (*artifact_store.PutResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	hash := sha256.Sum256(req.Data)
	ref := hex.EncodeToString(hash[:])
	m.mu.Lock()
	defer m.mu.Unlock()
	m.objects[ref] = req.Data
	m.owners[req.Owner] = append(m.owners[req.Owner], ref)

	addr := &artifact_store.ContentAddress{
		Version:    artifact_store.ContentAddressVersion,
		Hash:       hash[:],
		Algorithm:  "sha256",
		Size:       uint64(len(req.Data)),
		Backend:    m.backend,
		BackendRef: ref,
	}

	return &artifact_store.PutResponse{ContentAddress: addr}, nil
}

func (m *memoryArtifactStore) Get(_ context.Context, req *artifact_store.GetRequest) (*artifact_store.GetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	ref := req.ContentAddress.BackendRef
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.objects[ref]
	if !ok {
		return nil, artifact_store.ErrArtifactNotFound
	}
	return &artifact_store.GetResponse{Data: data}, nil
}

func (m *memoryArtifactStore) Delete(_ context.Context, req *artifact_store.DeleteRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	ref := req.ContentAddress.BackendRef
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objects, ref)
	return nil
}

func (m *memoryArtifactStore) Exists(_ context.Context, address *artifact_store.ContentAddress) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.objects[address.BackendRef]
	return ok, nil
}

func (m *memoryArtifactStore) GetChunk(_ context.Context, _ *artifact_store.ContentAddress, _ uint32) (*artifact_store.ChunkData, error) {
	return nil, artifact_store.ErrInvalidInput.Wrap("chunking not supported")
}

func (m *memoryArtifactStore) ListByOwner(_ context.Context, owner string, _ *artifact_store.Pagination) (*artifact_store.ListResponse, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	refs := m.owners[owner]
	references := make([]*artifact_store.ArtifactReference, 0, len(refs))
	for _, ref := range refs {
		address := &artifact_store.ContentAddress{
			Version:    artifact_store.ContentAddressVersion,
			Backend:    m.backend,
			BackendRef: ref,
		}
		references = append(references, &artifact_store.ArtifactReference{
			Version:        artifact_store.ArtifactReferenceVersion,
			ReferenceID:    ref,
			ContentAddress: address,
			AccountAddress: owner,
			ArtifactType:   "test",
			CreatedAt:      time.Now(),
		})
	}
	return &artifact_store.ListResponse{References: references}, nil
}

func (m *memoryArtifactStore) UpdateRetention(_ context.Context, _ *artifact_store.ContentAddress, _ *artifact_store.RetentionTag) error {
	return nil
}

func (m *memoryArtifactStore) PurgeExpired(_ context.Context, _ int64) (int, error) {
	return 0, nil
}

func (m *memoryArtifactStore) GetMetrics(_ context.Context) (*artifact_store.StorageMetrics, error) {
	return &artifact_store.StorageMetrics{}, nil
}

func (m *memoryArtifactStore) Health(_ context.Context) error {
	return nil
}

func (m *memoryArtifactStore) Backend() artifact_store.BackendType {
	return m.backend
}
