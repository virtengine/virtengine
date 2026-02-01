// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-2D: Offering sync state persistence for chain-to-Waldur synchronization.
// This file manages the sync state including checksum tracking, version history,
// error recording, and dead-letter queue handling.
package provider_daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// OfferingSyncState tracks sync state for all offerings managed by this provider.
type OfferingSyncState struct {
	// ProviderAddress is the provider this state belongs to.
	ProviderAddress string `json:"provider_address"`

	// Records maps offering IDs to their sync records.
	Records map[string]*OfferingSyncRecord `json:"records"`

	// DeadLetterQueue contains offerings that failed max retries.
	DeadLetterQueue []*DeadLetterItem `json:"dead_letter_queue,omitempty"`

	// LastReconcileAt is when the last reconciliation ran.
	LastReconcileAt *time.Time `json:"last_reconcile_at,omitempty"`

	// Metrics tracks sync operation metrics.
	Metrics *SyncMetrics `json:"metrics"`

	// UpdatedAt is when this state was last modified.
	UpdatedAt time.Time `json:"updated_at"`
}

// OfferingSyncRecord tracks sync state for a single offering.
type OfferingSyncRecord struct {
	// OfferingID is the on-chain offering ID.
	OfferingID string `json:"offering_id"`

	// WaldurUUID is the corresponding Waldur offering UUID.
	WaldurUUID string `json:"waldur_uuid,omitempty"`

	// State is the current sync state.
	State SyncState `json:"state"`

	// ChainVersion is the current on-chain version.
	ChainVersion uint64 `json:"chain_version"`

	// SyncedVersion is the version that was last synced to Waldur.
	SyncedVersion uint64 `json:"synced_version"`

	// Checksum is the hash of the synced data.
	Checksum string `json:"checksum"`

	// LastSyncedAt is when the offering was last successfully synced.
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`

	// LastAttemptAt is when sync was last attempted.
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// LastError is the most recent error message.
	LastError string `json:"last_error,omitempty"`

	// RetryCount is the number of consecutive retry attempts.
	RetryCount int `json:"retry_count"`

	// NextRetryAt is when the next retry should be attempted.
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`

	// CreatedAt is when this record was first created.
	CreatedAt time.Time `json:"created_at"`
}

// SyncState represents the sync state of an offering.
type SyncState string

const (
	// SyncStatePending indicates sync is pending.
	SyncStatePending SyncState = "pending"

	// SyncStateSynced indicates offering is in sync with Waldur.
	SyncStateSynced SyncState = "synced"

	// SyncStateFailed indicates last sync attempt failed.
	SyncStateFailed SyncState = "failed"

	// SyncStateRetrying indicates sync is being retried.
	SyncStateRetrying SyncState = "retrying"

	// SyncStateDeadLettered indicates sync failed permanently.
	SyncStateDeadLettered SyncState = "dead_lettered"

	// SyncStateOutOfSync indicates chain version > synced version.
	SyncStateOutOfSync SyncState = "out_of_sync"
)

// DeadLetterItem represents an offering that failed to sync after max retries.
type DeadLetterItem struct {
	// OfferingID is the offering that failed.
	OfferingID string `json:"offering_id"`

	// Action is the action that failed (create, update, disable).
	Action string `json:"action"`

	// LastError is the final error.
	LastError string `json:"last_error"`

	// RetryCount is the number of attempts made.
	RetryCount int `json:"retry_count"`

	// FirstAttemptAt is when the first attempt was made.
	FirstAttemptAt time.Time `json:"first_attempt_at"`

	// DeadLetteredAt is when the item was dead-lettered.
	DeadLetteredAt time.Time `json:"dead_lettered_at"`

	// Checksum is the offering checksum at time of failure.
	Checksum string `json:"checksum"`

	// ChainVersion is the chain version that failed to sync.
	ChainVersion uint64 `json:"chain_version"`
}

// SyncMetrics tracks aggregate sync metrics.
type SyncMetrics struct {
	// TotalSyncs is the total number of sync operations.
	TotalSyncs int64 `json:"total_syncs"`

	// SuccessfulSyncs is the count of successful syncs.
	SuccessfulSyncs int64 `json:"successful_syncs"`

	// FailedSyncs is the count of failed sync attempts.
	FailedSyncs int64 `json:"failed_syncs"`

	// DeadLettered is the count of dead-lettered offerings.
	DeadLettered int64 `json:"dead_lettered"`

	// DriftDetections is the count of drift detections.
	DriftDetections int64 `json:"drift_detections"`

	// ReconciliationsRun is the count of reconciliation cycles.
	ReconciliationsRun int64 `json:"reconciliations_run"`

	// LastSyncDurationMs is the duration of the last sync in milliseconds.
	LastSyncDurationMs int64 `json:"last_sync_duration_ms"`

	// AverageSyncDurationMs is the rolling average sync duration.
	AverageSyncDurationMs float64 `json:"average_sync_duration_ms"`
}

// NewOfferingSyncState creates a new sync state for a provider.
func NewOfferingSyncState(providerAddress string) *OfferingSyncState {
	return &OfferingSyncState{
		ProviderAddress: providerAddress,
		Records:         make(map[string]*OfferingSyncRecord),
		DeadLetterQueue: make([]*DeadLetterItem, 0),
		Metrics:         &SyncMetrics{},
		UpdatedAt:       time.Now().UTC(),
	}
}

// GetRecord retrieves a sync record by offering ID.
func (s *OfferingSyncState) GetRecord(offeringID string) *OfferingSyncRecord {
	return s.Records[offeringID]
}

// GetOrCreateRecord retrieves or creates a sync record for an offering.
func (s *OfferingSyncState) GetOrCreateRecord(offeringID string) *OfferingSyncRecord {
	if record, exists := s.Records[offeringID]; exists {
		return record
	}
	record := &OfferingSyncRecord{
		OfferingID:   offeringID,
		State:        SyncStatePending,
		ChainVersion: 1,
		CreatedAt:    time.Now().UTC(),
	}
	s.Records[offeringID] = record
	return record
}

// MarkSynced marks an offering as successfully synced.
func (s *OfferingSyncState) MarkSynced(offeringID, waldurUUID, checksum string, version uint64) {
	record := s.GetOrCreateRecord(offeringID)
	now := time.Now().UTC()
	record.WaldurUUID = waldurUUID
	record.State = SyncStateSynced
	record.SyncedVersion = version
	record.ChainVersion = version
	record.Checksum = checksum
	record.LastSyncedAt = &now
	record.LastError = ""
	record.RetryCount = 0
	record.NextRetryAt = nil
	s.UpdatedAt = now
	s.Metrics.SuccessfulSyncs++
	s.Metrics.TotalSyncs++
}

// MarkFailed marks an offering sync as failed. Returns true if the item was dead-lettered.
func (s *OfferingSyncState) MarkFailed(offeringID, errorMsg string, maxRetries int, baseBackoff, maxBackoff time.Duration) bool {
	record := s.GetOrCreateRecord(offeringID)
	now := time.Now().UTC()
	record.State = SyncStateFailed
	record.LastAttemptAt = &now
	record.LastError = errorMsg
	record.RetryCount++

	deadLettered := false

	// Calculate next retry with exponential backoff
	if record.RetryCount <= maxRetries {
		record.State = SyncStateRetrying
		backoff := baseBackoff * time.Duration(1<<uint(record.RetryCount-1))
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
		nextRetry := now.Add(backoff)
		record.NextRetryAt = &nextRetry
	} else {
		// Move to dead letter queue
		s.DeadLetter(offeringID, "update", record.Checksum, record.ChainVersion)
		deadLettered = true
	}

	s.UpdatedAt = now
	s.Metrics.FailedSyncs++
	s.Metrics.TotalSyncs++
	return deadLettered
}

// MarkOutOfSync marks an offering as out of sync with chain.
func (s *OfferingSyncState) MarkOutOfSync(offeringID string, chainVersion uint64, checksum string) {
	record := s.GetOrCreateRecord(offeringID)
	record.State = SyncStateOutOfSync
	record.ChainVersion = chainVersion
	record.Checksum = checksum
	s.UpdatedAt = time.Now().UTC()
	s.Metrics.DriftDetections++
}

// DeadLetter moves an offering to the dead letter queue.
func (s *OfferingSyncState) DeadLetter(offeringID, action, checksum string, chainVersion uint64) {
	record := s.GetRecord(offeringID)
	if record == nil {
		return
	}

	now := time.Now().UTC()
	item := &DeadLetterItem{
		OfferingID:     offeringID,
		Action:         action,
		LastError:      record.LastError,
		RetryCount:     record.RetryCount,
		FirstAttemptAt: record.CreatedAt,
		DeadLetteredAt: now,
		Checksum:       checksum,
		ChainVersion:   chainVersion,
	}

	s.DeadLetterQueue = append(s.DeadLetterQueue, item)
	record.State = SyncStateDeadLettered
	s.UpdatedAt = now
	s.Metrics.DeadLettered++
}

// NeedsSyncOfferings returns offering IDs that need syncing.
func (s *OfferingSyncState) NeedsSyncOfferings() []string {
	var ids []string
	now := time.Now().UTC()

	for id, record := range s.Records {
		switch record.State {
		case SyncStatePending, SyncStateOutOfSync:
			ids = append(ids, id)
		case SyncStateRetrying:
			if record.NextRetryAt != nil && now.After(*record.NextRetryAt) {
				ids = append(ids, id)
			}
		case SyncStateFailed:
			if record.NextRetryAt != nil && now.After(*record.NextRetryAt) {
				ids = append(ids, id)
			}
		}
	}

	return ids
}

// ReprocessDeadLetter attempts to reprocess a dead-lettered offering.
func (s *OfferingSyncState) ReprocessDeadLetter(offeringID string) bool {
	// Find and remove from dead letter queue
	for i, item := range s.DeadLetterQueue {
		if item.OfferingID == offeringID {
			s.DeadLetterQueue = append(s.DeadLetterQueue[:i], s.DeadLetterQueue[i+1:]...)
			break
		}
	}

	// Reset the record for retry
	record := s.GetRecord(offeringID)
	if record != nil {
		record.State = SyncStatePending
		record.RetryCount = 0
		record.LastError = ""
		record.NextRetryAt = nil
		s.UpdatedAt = time.Now().UTC()
		return true
	}

	return false
}

// RecordReconciliation records that a reconciliation cycle ran.
func (s *OfferingSyncState) RecordReconciliation() {
	now := time.Now().UTC()
	s.LastReconcileAt = &now
	s.Metrics.ReconciliationsRun++
	s.UpdatedAt = now
}

// OfferingSyncStateStore persists offering sync state to disk.
type OfferingSyncStateStore struct {
	path string
	mu   sync.RWMutex
}

// NewOfferingSyncStateStore creates a new state store.
func NewOfferingSyncStateStore(path string) *OfferingSyncStateStore {
	return &OfferingSyncStateStore{path: path}
}

// Load reads state from disk or returns a new state.
func (s *OfferingSyncStateStore) Load(providerAddress string) (*OfferingSyncState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewOfferingSyncState(providerAddress), nil
		}
		return nil, fmt.Errorf("read offering sync state: %w", err)
	}

	var state OfferingSyncState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode offering sync state: %w", err)
	}

	// Initialize maps if nil
	if state.Records == nil {
		state.Records = make(map[string]*OfferingSyncRecord)
	}
	if state.DeadLetterQueue == nil {
		state.DeadLetterQueue = make([]*DeadLetterItem, 0)
	}
	if state.Metrics == nil {
		state.Metrics = &SyncMetrics{}
	}

	return &state, nil
}

// Save writes state to disk atomically.
func (s *OfferingSyncStateStore) Save(state *OfferingSyncState) error {
	if state == nil {
		return fmt.Errorf("state is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	state.UpdatedAt = time.Now().UTC()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode offering sync state: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	// Atomic write via temp file
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write state tmp: %w", err)
	}

	if err := os.Rename(tmp, s.path); err != nil {
		return fmt.Errorf("replace state: %w", err)
	}

	return nil
}

// Delete removes the state file.
func (s *OfferingSyncStateStore) Delete() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete state: %w", err)
	}
	return nil
}

