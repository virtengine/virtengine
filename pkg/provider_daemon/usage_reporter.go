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
	snapshot := *record

	s.mu.Lock()
	defer s.mu.Unlock()

	s.trackKey(snapshot.WorkloadID, &snapshot)
	s.trackKey(snapshot.DeploymentID, &snapshot)
	s.trackKey(snapshot.LeaseID, &snapshot)
	s.trackKey(snapshot.AllocationID, &snapshot)
	s.trackKey(snapshot.ID, &snapshot)
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

	snapshot := *record
	return &snapshot, true
}

func (s *UsageSnapshotStore) trackKey(key string, record *UsageRecord) {
	if key == "" || record == nil {
		return
	}

	current := s.byKey[key]
	newer := current == nil || record.EndTime.After(current.EndTime)
	if current != nil && record.EndTime.Equal(current.EndTime) {
		newer = record.CreatedAt.After(current.CreatedAt) ||
			(record.CreatedAt.Equal(current.CreatedAt) && record.ID > current.ID)
	}
	if newer || (current != nil && record.ID != "" && record.ID == current.ID) {
		s.byKey[key] = record
	}
}
