// Package marketplace provides types for the marketplace on-chain module.
//
// VE-14D: On-chain sync status tracking for bidirectional Waldur offering sync.
// This file defines types for tracking sync status on-chain, enabling drift
// detection and reconciliation between VirtEngine and Waldur.
package marketplace

import (
	"time"
)

// SyncDirection represents the direction of synchronization.
type SyncDirection string

const (
	// SyncDirectionToWaldur indicates chain-to-Waldur sync (offerings pushed to Waldur).
	SyncDirectionToWaldur SyncDirection = "to_waldur"

	// SyncDirectionFromWaldur indicates Waldur-to-chain sync (offerings ingested from Waldur).
	SyncDirectionFromWaldur SyncDirection = "from_waldur"

	// SyncDirectionBidirectional indicates both directions are active.
	SyncDirectionBidirectional SyncDirection = "bidirectional"
)

// SyncStatusState represents the overall sync status state.
type SyncStatusState string

const (
	// SyncStatusStateHealthy indicates sync is working normally.
	SyncStatusStateHealthy SyncStatusState = "healthy"

	// SyncStatusStateDegraded indicates some sync operations are failing.
	SyncStatusStateDegraded SyncStatusState = "degraded"

	// SyncStatusStateFailed indicates sync is completely failed.
	SyncStatusStateFailed SyncStatusState = "failed"

	// SyncStatusStateDisabled indicates sync is disabled.
	SyncStatusStateDisabled SyncStatusState = "disabled"

	// SyncStatusStateReconciling indicates active reconciliation in progress.
	SyncStatusStateReconciling SyncStatusState = "reconciling"
)

// OfferingSyncStatus tracks the on-chain sync status for a provider's offerings.
// This is stored on-chain to enable cross-node visibility of sync state.
type OfferingSyncStatus struct {
	// ProviderAddress is the provider this status belongs to.
	ProviderAddress string `json:"provider_address"`

	// Direction is the sync direction for this provider.
	Direction SyncDirection `json:"direction"`

	// State is the overall sync status state.
	State SyncStatusState `json:"state"`

	// WaldurCustomerUUID is the linked Waldur customer UUID.
	WaldurCustomerUUID string `json:"waldur_customer_uuid,omitempty"`

	// ToWaldurStats contains chain-to-Waldur sync statistics.
	ToWaldurStats *SyncDirectionStats `json:"to_waldur_stats,omitempty"`

	// FromWaldurStats contains Waldur-to-chain ingestion statistics.
	FromWaldurStats *SyncDirectionStats `json:"from_waldur_stats,omitempty"`

	// LastReconciliationAt is when the last reconciliation completed.
	LastReconciliationAt *time.Time `json:"last_reconciliation_at,omitempty"`

	// LastDriftCheckAt is when the last drift detection check ran.
	LastDriftCheckAt *time.Time `json:"last_drift_check_at,omitempty"`

	// DriftCount is the number of offerings currently out of sync.
	DriftCount uint32 `json:"drift_count"`

	// DeadLetterCount is the number of offerings in dead letter queue.
	DeadLetterCount uint32 `json:"dead_letter_count"`

	// EnabledAt is when sync was enabled for this provider.
	EnabledAt *time.Time `json:"enabled_at,omitempty"`

	// UpdatedAt is when this status was last modified.
	UpdatedAt time.Time `json:"updated_at"`
}

// SyncDirectionStats contains statistics for a sync direction.
type SyncDirectionStats struct {
	// TotalOfferings is the total number of offerings tracked.
	TotalOfferings uint32 `json:"total_offerings"`

	// SyncedOfferings is the number of offerings in sync.
	SyncedOfferings uint32 `json:"synced_offerings"`

	// PendingOfferings is the number of offerings pending sync.
	PendingOfferings uint32 `json:"pending_offerings"`

	// FailedOfferings is the number of offerings with failed sync.
	FailedOfferings uint32 `json:"failed_offerings"`

	// DeadLetteredOfferings is the number of dead-lettered offerings.
	DeadLetteredOfferings uint32 `json:"dead_lettered_offerings"`

	// TotalSyncAttempts is the total number of sync attempts.
	TotalSyncAttempts uint64 `json:"total_sync_attempts"`

	// SuccessfulSyncs is the count of successful sync operations.
	SuccessfulSyncs uint64 `json:"successful_syncs"`

	// FailedSyncs is the count of failed sync attempts.
	FailedSyncs uint64 `json:"failed_syncs"`

	// LastSyncAt is when the last sync operation completed.
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`

	// LastSuccessAt is when the last successful sync completed.
	LastSuccessAt *time.Time `json:"last_success_at,omitempty"`

	// LastFailureAt is when the last sync failure occurred.
	LastFailureAt *time.Time `json:"last_failure_at,omitempty"`

	// LastError is the most recent error message.
	LastError string `json:"last_error,omitempty"`

	// AverageSyncDurationMs is the average sync duration in milliseconds.
	AverageSyncDurationMs float64 `json:"average_sync_duration_ms"`
}

// OfferingSyncRecord tracks the sync state of a single offering on-chain.
type OfferingSyncRecord struct {
	// OfferingID is the on-chain offering ID.
	OfferingID string `json:"offering_id"`

	// ProviderAddress is the provider's address.
	ProviderAddress string `json:"provider_address"`

	// WaldurUUID is the corresponding Waldur offering UUID (if synced).
	WaldurUUID string `json:"waldur_uuid,omitempty"`

	// SyncState is the current sync state.
	SyncState OfferingSyncState `json:"sync_state"`

	// Direction is the sync direction for this offering.
	Direction SyncDirection `json:"direction"`

	// ChainVersion is the on-chain offering version.
	ChainVersion uint64 `json:"chain_version"`

	// WaldurVersion is a hash representing the Waldur offering state.
	WaldurVersion string `json:"waldur_version,omitempty"`

	// ChainChecksum is the checksum of on-chain data.
	ChainChecksum string `json:"chain_checksum"`

	// WaldurChecksum is the checksum of Waldur data.
	WaldurChecksum string `json:"waldur_checksum,omitempty"`

	// InSync indicates if chain and Waldur are in sync.
	InSync bool `json:"in_sync"`

	// LastSyncedAt is when the offering was last successfully synced.
	LastSyncedAt *time.Time `json:"last_synced_at,omitempty"`

	// LastAttemptAt is when sync was last attempted.
	LastAttemptAt *time.Time `json:"last_attempt_at,omitempty"`

	// RetryCount is the number of consecutive retry attempts.
	RetryCount uint32 `json:"retry_count"`

	// NextRetryAt is when the next retry should be attempted.
	NextRetryAt *time.Time `json:"next_retry_at,omitempty"`

	// LastError is the most recent sync error.
	LastError string `json:"last_error,omitempty"`

	// CreatedAt is when this record was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this record was last modified.
	UpdatedAt time.Time `json:"updated_at"`
}

// OfferingSyncState represents the sync state of an individual offering.
type OfferingSyncState string

const (
	// OfferingSyncStatePending indicates sync is pending.
	OfferingSyncStatePending OfferingSyncState = "pending"

	// OfferingSyncStateSynced indicates offering is in sync.
	OfferingSyncStateSynced OfferingSyncState = "synced"

	// OfferingSyncStateFailed indicates last sync attempt failed.
	OfferingSyncStateFailed OfferingSyncState = "failed"

	// OfferingSyncStateRetrying indicates sync is being retried.
	OfferingSyncStateRetrying OfferingSyncState = "retrying"

	// OfferingSyncStateDeadLettered indicates sync failed permanently.
	OfferingSyncStateDeadLettered OfferingSyncState = "dead_lettered"

	// OfferingSyncStateDrifted indicates offering is out of sync.
	OfferingSyncStateDrifted OfferingSyncState = "drifted"

	// OfferingSyncStateSkipped indicates offering was intentionally skipped.
	OfferingSyncStateSkipped OfferingSyncState = "skipped"
)

// SyncEvent represents a sync event for audit trail.
type SyncEvent struct {
	// EventID is the unique event identifier.
	EventID string `json:"event_id"`

	// OfferingID is the offering involved.
	OfferingID string `json:"offering_id"`

	// ProviderAddress is the provider's address.
	ProviderAddress string `json:"provider_address"`

	// Direction is the sync direction.
	Direction SyncDirection `json:"direction"`

	// EventType is the type of sync event.
	EventType SyncEventType `json:"event_type"`

	// OldState is the previous sync state.
	OldState OfferingSyncState `json:"old_state,omitempty"`

	// NewState is the new sync state.
	NewState OfferingSyncState `json:"new_state"`

	// WaldurUUID is the Waldur offering UUID (if applicable).
	WaldurUUID string `json:"waldur_uuid,omitempty"`

	// ErrorMessage is the error message (if failed).
	ErrorMessage string `json:"error_message,omitempty"`

	// RetryCount is the current retry count.
	RetryCount uint32 `json:"retry_count"`

	// Timestamp is when the event occurred.
	Timestamp time.Time `json:"timestamp"`

	// Metadata contains additional event metadata.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SyncEventType represents the type of sync event.
type SyncEventType string

const (
	// SyncEventCreated is emitted when sync record is created.
	SyncEventCreated SyncEventType = "created"

	// SyncEventSynced is emitted when sync succeeds.
	SyncEventSynced SyncEventType = "synced"

	// SyncEventFailed is emitted when sync fails.
	SyncEventFailed SyncEventType = "failed"

	// SyncEventRetrying is emitted when sync is being retried.
	SyncEventRetrying SyncEventType = "retrying"

	// SyncEventDeadLettered is emitted when sync is dead-lettered.
	SyncEventDeadLettered SyncEventType = "dead_lettered"

	// SyncEventDriftDetected is emitted when drift is detected.
	SyncEventDriftDetected SyncEventType = "drift_detected"

	// SyncEventReconciled is emitted when reconciliation completes.
	SyncEventReconciled SyncEventType = "reconciled"

	// SyncEventReprocessed is emitted when a dead-letter is reprocessed.
	SyncEventReprocessed SyncEventType = "reprocessed"
)

// ReconciliationReport contains the results of a reconciliation run.
type ReconciliationReport struct {
	// ReportID is the unique report identifier.
	ReportID string `json:"report_id"`

	// ProviderAddress is the provider this report is for.
	ProviderAddress string `json:"provider_address"`

	// Direction is the sync direction reconciled.
	Direction SyncDirection `json:"direction"`

	// StartedAt is when reconciliation started.
	StartedAt time.Time `json:"started_at"`

	// CompletedAt is when reconciliation completed.
	CompletedAt time.Time `json:"completed_at"`

	// Duration is how long reconciliation took.
	DurationMs int64 `json:"duration_ms"`

	// TotalOfferings is the total offerings checked.
	TotalOfferings uint32 `json:"total_offerings"`

	// InSyncCount is offerings already in sync.
	InSyncCount uint32 `json:"in_sync_count"`

	// DriftedCount is offerings with detected drift.
	DriftedCount uint32 `json:"drifted_count"`

	// RepairedCount is offerings successfully repaired.
	RepairedCount uint32 `json:"repaired_count"`

	// FailedCount is offerings that failed to repair.
	FailedCount uint32 `json:"failed_count"`

	// NewOfferingsCount is newly discovered offerings.
	NewOfferingsCount uint32 `json:"new_offerings_count"`

	// DeadLetteredCount is offerings moved to dead letter.
	DeadLetteredCount uint32 `json:"dead_lettered_count"`

	// Success indicates if reconciliation completed successfully.
	Success bool `json:"success"`

	// Error is the error message if reconciliation failed.
	Error string `json:"error,omitempty"`

	// DriftDetails contains details about detected drift.
	DriftDetails []DriftDetail `json:"drift_details,omitempty"`
}

// DriftDetail describes a detected drift for an offering.
type DriftDetail struct {
	// OfferingID is the on-chain offering ID.
	OfferingID string `json:"offering_id"`

	// WaldurUUID is the Waldur offering UUID.
	WaldurUUID string `json:"waldur_uuid,omitempty"`

	// DriftType is the type of drift detected.
	DriftType DriftType `json:"drift_type"`

	// ChainValue is the on-chain value.
	ChainValue string `json:"chain_value,omitempty"`

	// WaldurValue is the Waldur value.
	WaldurValue string `json:"waldur_value,omitempty"`

	// Field is the field that drifted.
	Field string `json:"field,omitempty"`

	// Repaired indicates if the drift was repaired.
	Repaired bool `json:"repaired"`

	// Error is the repair error (if failed).
	Error string `json:"error,omitempty"`
}

// DriftType represents the type of drift detected.
type DriftType string

const (
	// DriftTypeNameMismatch indicates offering name differs.
	DriftTypeNameMismatch DriftType = "name_mismatch"

	// DriftTypeDescriptionMismatch indicates description differs.
	DriftTypeDescriptionMismatch DriftType = "description_mismatch"

	// DriftTypePricingMismatch indicates pricing differs.
	DriftTypePricingMismatch DriftType = "pricing_mismatch"

	// DriftTypeStateMismatch indicates state differs.
	DriftTypeStateMismatch DriftType = "state_mismatch"

	// DriftTypeAttributesMismatch indicates attributes differ.
	DriftTypeAttributesMismatch DriftType = "attributes_mismatch"

	// DriftTypeMissingInWaldur indicates offering missing in Waldur.
	DriftTypeMissingInWaldur DriftType = "missing_in_waldur"

	// DriftTypeMissingOnChain indicates offering missing on-chain.
	DriftTypeMissingOnChain DriftType = "missing_on_chain"

	// DriftTypeChecksumMismatch indicates overall checksum differs.
	DriftTypeChecksumMismatch DriftType = "checksum_mismatch"
)

// SyncConfiguration contains on-chain sync configuration for a provider.
type SyncConfiguration struct {
	// ProviderAddress is the provider this configuration is for.
	ProviderAddress string `json:"provider_address"`

	// Enabled indicates if sync is enabled.
	Enabled bool `json:"enabled"`

	// Direction is the sync direction(s) enabled.
	Direction SyncDirection `json:"direction"`

	// WaldurCustomerUUID is the linked Waldur customer UUID.
	WaldurCustomerUUID string `json:"waldur_customer_uuid"`

	// SyncIntervalSeconds is how often to run reconciliation.
	SyncIntervalSeconds uint32 `json:"sync_interval_seconds"`

	// MaxRetries is the max retries before dead-lettering.
	MaxRetries uint32 `json:"max_retries"`

	// AutoRepairDrift enables automatic drift repair.
	AutoRepairDrift bool `json:"auto_repair_drift"`

	// CategoryMapping maps on-chain categories to Waldur category UUIDs.
	CategoryMapping map[string]string `json:"category_mapping,omitempty"`

	// EnabledAt is when sync was enabled.
	EnabledAt *time.Time `json:"enabled_at,omitempty"`

	// LastModifiedAt is when configuration was last modified.
	LastModifiedAt time.Time `json:"last_modified_at"`

	// ModifiedBy is who last modified the configuration.
	ModifiedBy string `json:"modified_by,omitempty"`
}

// NewOfferingSyncStatus creates a new sync status for a provider.
func NewOfferingSyncStatus(providerAddress string) *OfferingSyncStatus {
	now := time.Now().UTC()
	return &OfferingSyncStatus{
		ProviderAddress: providerAddress,
		Direction:       SyncDirectionBidirectional,
		State:           SyncStatusStateDisabled,
		ToWaldurStats:   &SyncDirectionStats{},
		FromWaldurStats: &SyncDirectionStats{},
		UpdatedAt:       now,
	}
}

// Enable enables sync for the provider.
func (s *OfferingSyncStatus) Enable(direction SyncDirection, waldurCustomerUUID string) {
	now := time.Now().UTC()
	s.State = SyncStatusStateHealthy
	s.Direction = direction
	s.WaldurCustomerUUID = waldurCustomerUUID
	s.EnabledAt = &now
	s.UpdatedAt = now
}

// Disable disables sync for the provider.
func (s *OfferingSyncStatus) Disable() {
	s.State = SyncStatusStateDisabled
	s.UpdatedAt = time.Now().UTC()
}

// RecordSync records a sync operation.
func (s *OfferingSyncStatus) RecordSync(direction SyncDirection, success bool, err error) {
	now := time.Now().UTC()
	var stats *SyncDirectionStats

	switch direction {
	case SyncDirectionToWaldur:
		stats = s.ToWaldurStats
	case SyncDirectionFromWaldur:
		stats = s.FromWaldurStats
	default:
		return
	}

	if stats == nil {
		return
	}

	stats.TotalSyncAttempts++
	stats.LastSyncAt = &now

	if success {
		stats.SuccessfulSyncs++
		stats.LastSuccessAt = &now
		stats.LastError = ""
	} else {
		stats.FailedSyncs++
		stats.LastFailureAt = &now
		if err != nil {
			stats.LastError = err.Error()
		}
	}

	s.UpdatedAt = now
	s.updateState()
}

// RecordReconciliation records a reconciliation run.
func (s *OfferingSyncStatus) RecordReconciliation(report *ReconciliationReport) {
	now := time.Now().UTC()
	s.LastReconciliationAt = &now
	s.DriftCount = report.DriftedCount - report.RepairedCount
	s.DeadLetterCount += report.DeadLetteredCount
	s.UpdatedAt = now
	s.updateState()
}

// updateState updates the overall state based on statistics.
func (s *OfferingSyncStatus) updateState() {
	if s.State == SyncStatusStateDisabled {
		return
	}

	// Check failure rates
	totalStats := &SyncDirectionStats{}
	if s.ToWaldurStats != nil {
		totalStats.TotalSyncAttempts += s.ToWaldurStats.TotalSyncAttempts
		totalStats.FailedSyncs += s.ToWaldurStats.FailedSyncs
	}
	if s.FromWaldurStats != nil {
		totalStats.TotalSyncAttempts += s.FromWaldurStats.TotalSyncAttempts
		totalStats.FailedSyncs += s.FromWaldurStats.FailedSyncs
	}

	if totalStats.TotalSyncAttempts == 0 {
		s.State = SyncStatusStateHealthy
		return
	}

	failureRate := float64(totalStats.FailedSyncs) / float64(totalStats.TotalSyncAttempts)
	if failureRate > 0.5 {
		s.State = SyncStatusStateFailed
	} else if failureRate > 0.1 || s.DriftCount > 10 || s.DeadLetterCount > 5 {
		s.State = SyncStatusStateDegraded
	} else {
		s.State = SyncStatusStateHealthy
	}
}

// IsHealthy returns true if sync is in a healthy state.
func (s *OfferingSyncStatus) IsHealthy() bool {
	return s.State == SyncStatusStateHealthy
}

// GetSuccessRate returns the overall sync success rate.
func (s *OfferingSyncStatus) GetSuccessRate() float64 {
	var total, success uint64

	if s.ToWaldurStats != nil {
		total += s.ToWaldurStats.TotalSyncAttempts
		success += s.ToWaldurStats.SuccessfulSyncs
	}
	if s.FromWaldurStats != nil {
		total += s.FromWaldurStats.TotalSyncAttempts
		success += s.FromWaldurStats.SuccessfulSyncs
	}

	if total == 0 {
		return 1.0
	}
	return float64(success) / float64(total)
}

// NewOfferingSyncRecord creates a new sync record for an offering.
func NewOfferingSyncRecord(offeringID, providerAddress string, direction SyncDirection) *OfferingSyncRecord {
	now := time.Now().UTC()
	return &OfferingSyncRecord{
		OfferingID:      offeringID,
		ProviderAddress: providerAddress,
		SyncState:       OfferingSyncStatePending,
		Direction:       direction,
		ChainVersion:    1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// MarkSynced marks the offering as successfully synced.
func (r *OfferingSyncRecord) MarkSynced(waldurUUID, chainChecksum, waldurChecksum string) {
	now := time.Now().UTC()
	r.WaldurUUID = waldurUUID
	r.SyncState = OfferingSyncStateSynced
	r.ChainChecksum = chainChecksum
	r.WaldurChecksum = waldurChecksum
	r.InSync = true
	r.LastSyncedAt = &now
	r.LastAttemptAt = &now
	r.RetryCount = 0
	r.NextRetryAt = nil
	r.LastError = ""
	r.UpdatedAt = now
}

// MarkFailed marks the offering sync as failed.
func (r *OfferingSyncRecord) MarkFailed(err error, maxRetries uint32) bool {
	now := time.Now().UTC()
	r.LastAttemptAt = &now
	r.RetryCount++
	r.InSync = false
	r.UpdatedAt = now

	if err != nil {
		r.LastError = err.Error()
	}

	if r.RetryCount >= maxRetries {
		r.SyncState = OfferingSyncStateDeadLettered
		return true
	}

	r.SyncState = OfferingSyncStateRetrying
	// Exponential backoff: 30s * 2^(retryCount-1), max 1 hour
	backoffSeconds := 30 * (1 << (r.RetryCount - 1))
	if backoffSeconds > 3600 {
		backoffSeconds = 3600
	}
	nextRetry := now.Add(time.Duration(backoffSeconds) * time.Second)
	r.NextRetryAt = &nextRetry
	return false
}

// MarkDrifted marks the offering as drifted.
func (r *OfferingSyncRecord) MarkDrifted(newChecksum string) {
	now := time.Now().UTC()
	r.SyncState = OfferingSyncStateDrifted
	r.InSync = false
	if r.Direction == SyncDirectionToWaldur {
		r.ChainChecksum = newChecksum
	} else {
		r.WaldurChecksum = newChecksum
	}
	r.UpdatedAt = now
}

// ResetForRetry resets the record to allow retry.
func (r *OfferingSyncRecord) ResetForRetry() {
	now := time.Now().UTC()
	r.SyncState = OfferingSyncStatePending
	r.RetryCount = 0
	r.NextRetryAt = nil
	r.LastError = ""
	r.UpdatedAt = now
}
