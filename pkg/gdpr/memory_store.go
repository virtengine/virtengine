package gdpr

import (
	"context"
	"sync"
)

// MemoryStore is a simple in-memory implementation of ExportStore.
type MemoryStore struct {
	mu        sync.RWMutex
	exports   map[string]*DataExportRequest
	deletions map[string]*DeletionRequest
	data      map[string][]byte
}

// NewMemoryStore creates a new MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		exports:   make(map[string]*DataExportRequest),
		deletions: make(map[string]*DeletionRequest),
		data:      make(map[string][]byte),
	}
}

// SaveExportRequest stores an export request.
func (s *MemoryStore) SaveExportRequest(_ context.Context, req *DataExportRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.exports[req.ID] = req
	return nil
}

// SaveDeletionRequest stores a deletion request.
func (s *MemoryStore) SaveDeletionRequest(_ context.Context, req *DeletionRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deletions[req.ID] = req
	return nil
}

// SaveExportData stores export payload bytes.
func (s *MemoryStore) SaveExportData(_ context.Context, requestID string, data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[requestID] = data
	return nil
}

// GetExportRequest retrieves an export request by ID.
func (s *MemoryStore) GetExportRequest(requestID string) (*DataExportRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	req, ok := s.exports[requestID]
	return req, ok
}

// GetDeletionRequest retrieves a deletion request by ID.
func (s *MemoryStore) GetDeletionRequest(requestID string) (*DeletionRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	req, ok := s.deletions[requestID]
	return req, ok
}

// GetExportData retrieves stored export bytes.
func (s *MemoryStore) GetExportData(requestID string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.data[requestID]
	return data, ok
}
