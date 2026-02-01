// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-3D: Waldur ingestion state persistence for Waldur-to-chain synchronization.
// This file manages the ingestion state including checksum tracking, version history,
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

// WaldurIngestState tracks ingestion state for all Waldur offerings.
type WaldurIngestState struct {
	// WaldurCustomerUUID is the Waldur customer this state belongs to.
	WaldurCustomerUUID string `json:"waldur_customer_uuid"`

	// ProviderAddress is the corresponding on-chain provider address.
	ProviderAddress string `json:"provider_address"`

	// Records maps Waldur offering UUIDs to their ingestion records.
	Records map[string]*WaldurIngestRecord `json:"records"`

	// DeadLetterQueue contains offerings that failed max retries.
	DeadLetterQueue []*IngestDeadLetterItem `json:"dead_letter_queue,omitempty"`

	// LastReconcileAt is when the last reconciliation ran.
	LastReconcileAt *time.Time `json:"last_reconcile_at,omitempty"`

	// LastIngestAt is when the last ingestion job ran.
	LastIngestAt *time.Time `json:"last_ingest_at,omitempty"`

	// Cursor is the pagination cursor for resuming ingestion.
	Cursor *IngestCursor `json:"cursor,omitempty"`

	// Metrics tracks ingestion operation metrics.
	Metrics *IngestStateMetrics `json:"metrics"`

	// UpdatedAt is when this state was last modified.
	UpdatedAt time.Time `json:"updated_at"`

	mu sync.RWMutex
}

// WaldurIngestRecord tracks ingestion state for a single Waldur offering.
type WaldurIngestRecord struct {
	// WaldurUUID is the Waldur offering UUID.
	WaldurUUID string `json:"waldur_uuid"`

	// ChainOfferingID is the on-chain offering ID (after first ingest).
	ChainOfferingID string `json:"chain_offering_id,omitempty"`

	// State is the current ingestion state.
	State IngestRecordState `json:"state"`

	// WaldurChecksum is the checksum of Waldur data at last ingest.
	WaldurChecksum string `json:"waldur_checksum"`

	// ChainVersion is the on-chain offering version.
	ChainVersion uint64 `json:"chain_version"`

	// LastIngestedAt is when the offering was last successfully ingested.
	LastIngestedAt *time.Time `json:"last_ingested_at,omitempty"`

	// LastAttemptAt is when ingestion was last attempted.
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// LastWaldurModified is when the Waldur offering was last modified.
	LastWaldurModified *time.Time `json:"last_waldur_modified,omitempty"`

	// LastError is the most recent error message.
	LastError string `json:"last_error,omitempty"`

	// RetryCount is the number of consecutive retry attempts.
	RetryCount int `json:"retry_count"`

	// NextRetryAt is when the next retry should be attempted.
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`

	// ProviderAddress is the resolved provider address.
	ProviderAddress string `json:"provider_address"`

	// Category is the resolved category.
	Category string `json:"category"`

	// OfferingName is cached for logging.
	OfferingName string `json:"offering_name"`

	// CreatedAt is when this record was first created.
	CreatedAt time.Time `json:"created_at"`
}

// IngestRecordState represents the ingestion state of a Waldur offering.
type IngestRecordState string

const (
	// IngestRecordStatePending indicates ingestion is pending.
	IngestRecordStatePending IngestRecordState = "pending"

	// IngestRecordStateIngested indicates offering is ingested on-chain.
	IngestRecordStateIngested IngestRecordState = "ingested"

	// IngestRecordStateFailed indicates last ingestion attempt failed.
	IngestRecordStateFailed IngestRecordState = "failed"

	// IngestRecordStateRetrying indicates ingestion is being retried.
	IngestRecordStateRetrying IngestRecordState = "retrying"

	// IngestRecordStateDeadLettered indicates ingestion failed permanently.
	IngestRecordStateDeadLettered IngestRecordState = "dead_lettered"

	// IngestRecordStateOutOfSync indicates Waldur data changed since last ingest.
	IngestRecordStateOutOfSync IngestRecordState = "out_of_sync"

	// IngestRecordStateSkipped indicates offering was intentionally skipped.
	IngestRecordStateSkipped IngestRecordState = "skipped"

	// IngestRecordStateDeprecated indicates the Waldur offering was archived.
	IngestRecordStateDeprecated IngestRecordState = "deprecated"
)

// IngestDeadLetterItem represents an offering that failed to ingest after max retries.
type IngestDeadLetterItem struct {
	// WaldurUUID is the Waldur offering UUID.
	WaldurUUID string `json:"waldur_uuid"`

	// OfferingName is the offering name for reference.
	OfferingName string `json:"offering_name"`

	// Action is the action that failed (create, update, deprecate).
	Action string `json:"action"`

	// LastError is the final error.
	LastError string `json:"last_error"`

	// RetryCount is the number of attempts made.
	RetryCount int `json:"retry_count"`

	// FirstAttemptAt is when the first attempt was made.
	FirstAttemptAt time.Time `json:"first_attempt_at"`

	// DeadLetteredAt is when the item was dead-lettered.
	DeadLetteredAt time.Time `json:"dead_lettered_at"`

	// WaldurChecksum is the checksum at time of failure.
	WaldurChecksum string `json:"waldur_checksum"`
}

// IngestCursor tracks pagination state for resuming ingestion.
type IngestCursor struct {
	// Page is the current page number.
	Page int `json:"page"`

	// PageSize is the number of items per page.
	PageSize int `json:"page_size"`

	// TotalPages is the estimated total pages.
	TotalPages int `json:"total_pages,omitempty"`

	// ProcessedCount is the number of offerings processed.
	ProcessedCount int `json:"processed_count"`

	// LastProcessedUUID is the UUID of the last processed offering.
	LastProcessedUUID string `json:"last_processed_uuid,omitempty"`

	// StartedAt is when this ingestion run started.
	StartedAt time.Time `json:"started_at"`
}

// IngestStateMetrics tracks aggregate ingestion metrics.
type IngestStateMetrics struct {
	// TotalIngests is the total number of ingestion operations.
	TotalIngests int64 `json:"total_ingests"`

	// SuccessfulIngests is the count of successful ingestions.
	SuccessfulIngests int64 `json:"successful_ingests"`

	// FailedIngests is the count of failed ingestion attempts.
	FailedIngests int64 `json:"failed_ingests"`

	// DeadLettered is the count of dead-lettered offerings.
	DeadLettered int64 `json:"dead_lettered"`

	// Skipped is the count of skipped offerings.
	Skipped int64 `json:"skipped"`

	// DriftDetections is the count of drift detections.
	DriftDetections int64 `json:"drift_detections"`

	// ReconciliationsRun is the count of reconciliation cycles.
	ReconciliationsRun int64 `json:"reconciliations_run"`

	// OfferingsCreated is the count of offerings created on-chain.
	OfferingsCreated int64 `json:"offerings_created"`

	// OfferingsUpdated is the count of offerings updated on-chain.
	OfferingsUpdated int64 `json:"offerings_updated"`

	// OfferingsDeprecated is the count of offerings deprecated.
	OfferingsDeprecated int64 `json:"offerings_deprecated"`

	// LastIngestDurationMs is the duration of the last ingest in milliseconds.
	LastIngestDurationMs int64 `json:"last_ingest_duration_ms"`

	// AverageIngestDurationMs is the rolling average ingest duration.
	AverageIngestDurationMs float64 `json:"average_ingest_duration_ms"`
}

// NewWaldurIngestState creates a new ingestion state.
func NewWaldurIngestState(waldurCustomerUUID, providerAddress string) *WaldurIngestState {
	return &WaldurIngestState{
		WaldurCustomerUUID: waldurCustomerUUID,
		ProviderAddress:    providerAddress,
		Records:            make(map[string]*WaldurIngestRecord),
		DeadLetterQueue:    make([]*IngestDeadLetterItem, 0),
		Metrics:            &IngestStateMetrics{},
		UpdatedAt:          time.Now().UTC(),
	}
}

// GetRecord retrieves an ingestion record by Waldur UUID.
func (s *WaldurIngestState) GetRecord(waldurUUID string) *WaldurIngestRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Records[waldurUUID]
}

// GetOrCreateRecord retrieves or creates an ingestion record.
func (s *WaldurIngestState) GetOrCreateRecord(waldurUUID, offeringName string) *WaldurIngestRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	if record, exists := s.Records[waldurUUID]; exists {
		return record
	}

	record := &WaldurIngestRecord{
		WaldurUUID:      waldurUUID,
		State:           IngestRecordStatePending,
		OfferingName:    offeringName,
		ProviderAddress: s.ProviderAddress,
		CreatedAt:       time.Now().UTC(),
	}
	s.Records[waldurUUID] = record
	return record
}

// MarkIngested marks an offering as successfully ingested.
func (s *WaldurIngestState) MarkIngested(waldurUUID, chainOfferingID, checksum string, version uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := s.Records[waldurUUID]
	if record == nil {
		record = &WaldurIngestRecord{
			WaldurUUID: waldurUUID,
			CreatedAt:  time.Now().UTC(),
		}
		s.Records[waldurUUID] = record
	}

	now := time.Now().UTC()
	record.ChainOfferingID = chainOfferingID
	record.State = IngestRecordStateIngested
	record.WaldurChecksum = checksum
	record.ChainVersion = version
	record.LastIngestedAt = &now
	record.LastError = ""
	record.RetryCount = 0
	record.NextRetryAt = nil
	s.UpdatedAt = now
	s.Metrics.SuccessfulIngests++
	s.Metrics.TotalIngests++
}

// MarkFailed marks an ingestion as failed. Returns true if the item was dead-lettered.
func (s *WaldurIngestState) MarkFailed(waldurUUID, errorMsg string, maxRetries int, baseBackoff, maxBackoff time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := s.Records[waldurUUID]
	if record == nil {
		return false
	}

	now := time.Now().UTC()
	record.State = IngestRecordStateFailed
	record.LastAttemptAt = &now
	record.LastError = errorMsg
	record.RetryCount++

	deadLettered := false

	// Calculate next retry with exponential backoff
	if record.RetryCount <= maxRetries {
		record.State = IngestRecordStateRetrying
		backoff := baseBackoff * time.Duration(1<<uint(record.RetryCount-1))
		if backoff > maxBackoff {
			backoff = maxBackoff
		}
		nextRetry := now.Add(backoff)
		record.NextRetryAt = &nextRetry
	} else {
		// Move to dead letter queue
		s.addToDeadLetter(record)
		deadLettered = true
	}

	s.UpdatedAt = now
	s.Metrics.FailedIngests++
	s.Metrics.TotalIngests++
	return deadLettered
}

// MarkOutOfSync marks an offering as out of sync with Waldur.
func (s *WaldurIngestState) MarkOutOfSync(waldurUUID, newChecksum string, waldurModified time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := s.Records[waldurUUID]
	if record == nil {
		return
	}

	record.State = IngestRecordStateOutOfSync
	record.WaldurChecksum = newChecksum
	record.LastWaldurModified = &waldurModified
	s.UpdatedAt = time.Now().UTC()
	s.Metrics.DriftDetections++
}

// MarkSkipped marks an offering as intentionally skipped.
func (s *WaldurIngestState) MarkSkipped(waldurUUID, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := s.Records[waldurUUID]
	if record == nil {
		return
	}

	record.State = IngestRecordStateSkipped
	record.LastError = reason
	s.UpdatedAt = time.Now().UTC()
	s.Metrics.Skipped++
}

// MarkDeprecated marks an offering as deprecated (archived in Waldur).
func (s *WaldurIngestState) MarkDeprecated(waldurUUID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	record := s.Records[waldurUUID]
	if record == nil {
		return
	}

	record.State = IngestRecordStateDeprecated
	s.UpdatedAt = time.Now().UTC()
	s.Metrics.OfferingsDeprecated++
}

// addToDeadLetter adds a record to the dead letter queue (must hold lock).
func (s *WaldurIngestState) addToDeadLetter(record *WaldurIngestRecord) {
	now := time.Now().UTC()
	item := &IngestDeadLetterItem{
		WaldurUUID:     record.WaldurUUID,
		OfferingName:   record.OfferingName,
		Action:         "ingest",
		LastError:      record.LastError,
		RetryCount:     record.RetryCount,
		FirstAttemptAt: record.CreatedAt,
		DeadLetteredAt: now,
		WaldurChecksum: record.WaldurChecksum,
	}

	s.DeadLetterQueue = append(s.DeadLetterQueue, item)
	record.State = IngestRecordStateDeadLettered
	s.Metrics.DeadLettered++
}

// NeedsIngestOfferings returns Waldur UUIDs that need ingestion.
func (s *WaldurIngestState) NeedsIngestOfferings() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var uuids []string
	now := time.Now().UTC()

	for uuid, record := range s.Records {
		switch record.State {
		case IngestRecordStatePending, IngestRecordStateOutOfSync:
			uuids = append(uuids, uuid)
		case IngestRecordStateRetrying, IngestRecordStateFailed:
			if record.NextRetryAt != nil && now.After(*record.NextRetryAt) {
				uuids = append(uuids, uuid)
			}
		}
	}

	return uuids
}

// ReprocessDeadLetter attempts to reprocess a dead-lettered offering.
func (s *WaldurIngestState) ReprocessDeadLetter(waldurUUID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and remove from dead letter queue
	for i, item := range s.DeadLetterQueue {
		if item.WaldurUUID == waldurUUID {
			s.DeadLetterQueue = append(s.DeadLetterQueue[:i], s.DeadLetterQueue[i+1:]...)
			break
		}
	}

	// Reset the record for retry
	record := s.Records[waldurUUID]
	if record != nil {
		record.State = IngestRecordStatePending
		record.RetryCount = 0
		record.LastError = ""
		record.NextRetryAt = nil
		s.UpdatedAt = time.Now().UTC()
		return true
	}

	return false
}

// RecordReconciliation records that a reconciliation cycle ran.
func (s *WaldurIngestState) RecordReconciliation() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	s.LastReconcileAt = &now
	s.Metrics.ReconciliationsRun++
	s.UpdatedAt = now
}

// UpdateCursor updates the pagination cursor.
func (s *WaldurIngestState) UpdateCursor(cursor *IngestCursor) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Cursor = cursor
	s.UpdatedAt = time.Now().UTC()
}

// ResetCursor resets the pagination cursor.
func (s *WaldurIngestState) ResetCursor() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Cursor = nil
	now := time.Now().UTC()
	s.LastIngestAt = &now
	s.UpdatedAt = now
}

// GetStats returns summary statistics.
func (s *WaldurIngestState) GetStats() IngestStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := IngestStats{
		TotalRecords: len(s.Records),
		DeadLettered: len(s.DeadLetterQueue),
	}

	for _, record := range s.Records {
		switch record.State {
		case IngestRecordStatePending:
			stats.Pending++
		case IngestRecordStateIngested:
			stats.Ingested++
		case IngestRecordStateFailed, IngestRecordStateRetrying:
			stats.Failed++
		case IngestRecordStateOutOfSync:
			stats.OutOfSync++
		case IngestRecordStateSkipped:
			stats.Skipped++
		case IngestRecordStateDeprecated:
			stats.Deprecated++
		}
	}

	return stats
}

// IngestStats provides summary statistics.
type IngestStats struct {
	TotalRecords int `json:"total_records"`
	Pending      int `json:"pending"`
	Ingested     int `json:"ingested"`
	Failed       int `json:"failed"`
	OutOfSync    int `json:"out_of_sync"`
	Skipped      int `json:"skipped"`
	Deprecated   int `json:"deprecated"`
	DeadLettered int `json:"dead_lettered"`
}

// WaldurIngestStateStore persists ingestion state to disk.
type WaldurIngestStateStore struct {
	path string
	mu   sync.RWMutex
}

// NewWaldurIngestStateStore creates a new state store.
func NewWaldurIngestStateStore(path string) *WaldurIngestStateStore {
	return &WaldurIngestStateStore{path: path}
}

// Load reads state from disk or returns a new state.
func (s *WaldurIngestStateStore) Load(waldurCustomerUUID, providerAddress string) (*WaldurIngestState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewWaldurIngestState(waldurCustomerUUID, providerAddress), nil
		}
		return nil, fmt.Errorf("read waldur ingest state: %w", err)
	}

	var state WaldurIngestState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("decode waldur ingest state: %w", err)
	}

	// Initialize maps if nil
	if state.Records == nil {
		state.Records = make(map[string]*WaldurIngestRecord)
	}
	if state.DeadLetterQueue == nil {
		state.DeadLetterQueue = make([]*IngestDeadLetterItem, 0)
	}
	if state.Metrics == nil {
		state.Metrics = &IngestStateMetrics{}
	}

	return &state, nil
}

// Save writes state to disk atomically.
func (s *WaldurIngestStateStore) Save(state *WaldurIngestState) error {
	if state == nil {
		return fmt.Errorf("state is nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	state.mu.RLock()
	state.UpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(state, "", "  ")
	state.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("encode waldur ingest state: %w", err)
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
func (s *WaldurIngestStateStore) Delete() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete state: %w", err)
	}
	return nil
}

