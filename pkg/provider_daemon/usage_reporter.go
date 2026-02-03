package provider_daemon

import (
	"sync"
	"time"
)

// UsageReporter resolves usage data for a given allocation request.
type UsageReporter interface {
	FindLatest(allocationID string, periodStart, periodEnd *time.Time) (*UsageRecord, bool)
}

// UsageSnapshotStore keeps the latest usage record indexed by multiple identifiers.
type UsageSnapshotStore struct {
	mu    sync.RWMutex
	byKey map[string]*UsageRecord
}

// NewUsageSnapshotStore creates a new usage snapshot store.
func NewUsageSnapshotStore() *UsageSnapshotStore {
	return &UsageSnapshotStore{
		byKey: make(map[string]*UsageRecord),
	}
}

// Track stores a usage record for lookup by workload, deployment, lease, and record IDs.
func (s *UsageSnapshotStore) Track(record *UsageRecord) {
	if record == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.trackKey(record.WorkloadID, record)
	s.trackKey(record.DeploymentID, record)
	s.trackKey(record.LeaseID, record)
	s.trackKey(record.ID, record)
}

// FindLatest returns the latest usage record for the requested allocation ID.
func (s *UsageSnapshotStore) FindLatest(allocationID string, periodStart, periodEnd *time.Time) (*UsageRecord, bool) {
	if allocationID == "" {
		return nil, false
	}

	s.mu.RLock()
	record := s.byKey[allocationID]
	s.mu.RUnlock()

	if record == nil {
		return nil, false
	}

	if periodStart != nil && record.EndTime.Before(*periodStart) {
		return nil, false
	}
	if periodEnd != nil && record.StartTime.After(*periodEnd) {
		return nil, false
	}

	return record, true
}

func (s *UsageSnapshotStore) trackKey(key string, record *UsageRecord) {
	if key == "" || record == nil {
		return
	}

	current := s.byKey[key]
	if current == nil || record.EndTime.After(current.EndTime) {
		s.byKey[key] = record
	}
}
