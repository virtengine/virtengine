// Package types contains types for the HPC module.
//
// VE-5A: Usage snapshots for audit and reconciliation
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SnapshotType indicates the type of usage snapshot
type SnapshotType string

const (
	// SnapshotTypeInterim is a periodic interim snapshot
	SnapshotTypeInterim SnapshotType = "interim"

	// SnapshotTypeFinal is the final snapshot at job completion
	SnapshotTypeFinal SnapshotType = "final"

	// SnapshotTypeReconciliation is a reconciliation snapshot
	SnapshotTypeReconciliation SnapshotType = "reconciliation"

	// SnapshotTypeDispute is a snapshot for dispute evidence
	SnapshotTypeDispute SnapshotType = "dispute"
)

// IsValidSnapshotType checks if the snapshot type is valid
func IsValidSnapshotType(t SnapshotType) bool {
	switch t {
	case SnapshotTypeInterim, SnapshotTypeFinal, SnapshotTypeReconciliation, SnapshotTypeDispute:
		return true
	default:
		return false
	}
}

// HPCUsageSnapshot represents a point-in-time snapshot of usage data
type HPCUsageSnapshot struct {
	// SnapshotID is the unique identifier
	SnapshotID string `json:"snapshot_id"`

	// JobID is the job this snapshot is for
	JobID string `json:"job_id"`

	// ClusterID is the cluster running the job
	ClusterID string `json:"cluster_id"`

	// SchedulerType is the scheduler type
	SchedulerType string `json:"scheduler_type"`

	// SchedulerJobID is the native scheduler job ID
	SchedulerJobID string `json:"scheduler_job_id,omitempty"`

	// SnapshotType indicates the snapshot type
	SnapshotType SnapshotType `json:"snapshot_type"`

	// SequenceNumber is the sequence within the job (0, 1, 2...)
	SequenceNumber uint32 `json:"sequence_number"`

	// ProviderAddress is the provider
	ProviderAddress string `json:"provider_address"`

	// CustomerAddress is the customer
	CustomerAddress string `json:"customer_address"`

	// Metrics contains the usage metrics at snapshot time
	Metrics HPCDetailedMetrics `json:"metrics"`

	// CumulativeMetrics are cumulative since job start
	CumulativeMetrics HPCDetailedMetrics `json:"cumulative_metrics"`

	// DeltaMetrics are the delta since last snapshot
	DeltaMetrics HPCDetailedMetrics `json:"delta_metrics"`

	// JobState is the job state at snapshot time
	JobState JobState `json:"job_state"`

	// SnapshotTime is when the snapshot was taken
	SnapshotTime time.Time `json:"snapshot_time"`

	// PreviousSnapshotID links to the previous snapshot
	PreviousSnapshotID string `json:"previous_snapshot_id,omitempty"`

	// ProviderSignature is the provider's signature on the snapshot
	ProviderSignature string `json:"provider_signature"`

	// ContentHash is the hash of the snapshot content
	ContentHash string `json:"content_hash"`

	// SchedulerRawData contains raw scheduler output (for reconciliation)
	SchedulerRawData string `json:"scheduler_raw_data,omitempty"`

	// CreatedAt is when the snapshot was created
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is when the snapshot was recorded on-chain
	BlockHeight int64 `json:"block_height"`
}

// CalculateContentHash calculates the hash of snapshot content
func (s *HPCUsageSnapshot) CalculateContentHash() string {
	hashInput := struct {
		JobID             string             `json:"job_id"`
		ClusterID         string             `json:"cluster_id"`
		SequenceNumber    uint32             `json:"sequence_number"`
		Metrics           HPCDetailedMetrics `json:"metrics"`
		CumulativeMetrics HPCDetailedMetrics `json:"cumulative_metrics"`
		SnapshotTime      int64              `json:"snapshot_time"`
	}{
		JobID:             s.JobID,
		ClusterID:         s.ClusterID,
		SequenceNumber:    s.SequenceNumber,
		Metrics:           s.Metrics,
		CumulativeMetrics: s.CumulativeMetrics,
		SnapshotTime:      s.SnapshotTime.Unix(),
	}

	data, err := json.Marshal(hashInput)
	if err != nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Validate validates the usage snapshot
func (s *HPCUsageSnapshot) Validate() error {
	if s.SnapshotID == "" {
		return fmt.Errorf("snapshot_id cannot be empty")
	}

	if len(s.SnapshotID) > 64 {
		return fmt.Errorf("snapshot_id exceeds maximum length")
	}

	if s.JobID == "" {
		return fmt.Errorf("job_id cannot be empty")
	}

	if s.ClusterID == "" {
		return fmt.Errorf("cluster_id cannot be empty")
	}

	if !IsValidSnapshotType(s.SnapshotType) {
		return fmt.Errorf("invalid snapshot_type: %s", s.SnapshotType)
	}

	if _, err := sdk.AccAddressFromBech32(s.ProviderAddress); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(s.CustomerAddress); err != nil {
		return fmt.Errorf("invalid customer address: %w", err)
	}

	if s.ProviderSignature == "" {
		return fmt.Errorf("provider_signature cannot be empty")
	}

	return nil
}

// ReconciliationStatus indicates the status of a reconciliation
type ReconciliationStatus string

const (
	// ReconciliationStatusPending indicates pending reconciliation
	ReconciliationStatusPending ReconciliationStatus = "pending"

	// ReconciliationStatusMatched indicates records match
	ReconciliationStatusMatched ReconciliationStatus = "matched"

	// ReconciliationStatusDiscrepancy indicates a discrepancy was found
	ReconciliationStatusDiscrepancy ReconciliationStatus = "discrepancy"

	// ReconciliationStatusResolved indicates discrepancy was resolved
	ReconciliationStatusResolved ReconciliationStatus = "resolved"

	// ReconciliationStatusFailed indicates reconciliation failed
	ReconciliationStatusFailed ReconciliationStatus = "failed"
)

// IsValidReconciliationStatus checks if the status is valid
func IsValidReconciliationStatus(s ReconciliationStatus) bool {
	switch s {
	case ReconciliationStatusPending, ReconciliationStatusMatched,
		ReconciliationStatusDiscrepancy, ReconciliationStatusResolved, ReconciliationStatusFailed:
		return true
	default:
		return false
	}
}

// HPCReconciliationRecord represents a reconciliation between scheduler and on-chain records
type HPCReconciliationRecord struct {
	// ReconciliationID is the unique identifier
	ReconciliationID string `json:"reconciliation_id"`

	// JobID is the job being reconciled
	JobID string `json:"job_id"`

	// ClusterID is the cluster
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the provider
	ProviderAddress string `json:"provider_address"`

	// SchedulerSource is the scheduler data source
	SchedulerSource ReconciliationSource `json:"scheduler_source"`

	// OnChainSource is the on-chain data source
	OnChainSource ReconciliationSource `json:"on_chain_source"`

	// Status is the reconciliation status
	Status ReconciliationStatus `json:"status"`

	// Discrepancies lists any discrepancies found
	Discrepancies []ReconciliationDiscrepancy `json:"discrepancies,omitempty"`

	// Resolution describes how discrepancies were resolved
	Resolution string `json:"resolution,omitempty"`

	// ResolutionAction is the action taken
	ResolutionAction string `json:"resolution_action,omitempty"`

	// AdjustmentRecord links to adjustment accounting record
	AdjustmentRecordID string `json:"adjustment_record_id,omitempty"`

	// ReconciliationTime is when reconciliation was performed
	ReconciliationTime time.Time `json:"reconciliation_time"`

	// ResolvedAt is when discrepancies were resolved
	ResolvedAt *time.Time `json:"resolved_at,omitempty"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is when the record was created on-chain
	BlockHeight int64 `json:"block_height"`
}

// ReconciliationSource represents a data source for reconciliation
type ReconciliationSource struct {
	// SourceType is the type (scheduler, on_chain, usage_record)
	SourceType string `json:"source_type"`

	// SourceID is the source record ID
	SourceID string `json:"source_id"`

	// ExtractTime is when data was extracted
	ExtractTime time.Time `json:"extract_time"`

	// Metrics are the metrics from this source
	Metrics HPCDetailedMetrics `json:"metrics"`

	// Hash is the content hash
	Hash string `json:"hash"`
}

// ReconciliationDiscrepancy represents a discrepancy between sources
type ReconciliationDiscrepancy struct {
	// Field is the field with discrepancy
	Field string `json:"field"`

	// SchedulerValue is the value from scheduler
	SchedulerValue string `json:"scheduler_value"`

	// OnChainValue is the value on-chain
	OnChainValue string `json:"on_chain_value"`

	// DifferencePercent is the percentage difference
	DifferencePercent string `json:"difference_percent"`

	// Severity indicates the severity (low, medium, high, critical)
	Severity string `json:"severity"`

	// ToleranceExceeded indicates if tolerance was exceeded
	ToleranceExceeded bool `json:"tolerance_exceeded"`
}

// Validate validates the reconciliation record
func (r *HPCReconciliationRecord) Validate() error {
	if r.ReconciliationID == "" {
		return fmt.Errorf("reconciliation_id cannot be empty")
	}

	if r.JobID == "" {
		return fmt.Errorf("job_id cannot be empty")
	}

	if r.ClusterID == "" {
		return fmt.Errorf("cluster_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(r.ProviderAddress); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if !IsValidReconciliationStatus(r.Status) {
		return fmt.Errorf("invalid status: %s", r.Status)
	}

	return nil
}

// HasDiscrepancies returns true if there are discrepancies
func (r *HPCReconciliationRecord) HasDiscrepancies() bool {
	return len(r.Discrepancies) > 0
}

// GetCriticalDiscrepancies returns discrepancies with critical severity
func (r *HPCReconciliationRecord) GetCriticalDiscrepancies() []ReconciliationDiscrepancy {
	var critical []ReconciliationDiscrepancy
	for _, d := range r.Discrepancies {
		if d.Severity == "critical" || d.Severity == "high" {
			critical = append(critical, d)
		}
	}
	return critical
}

// HasCriticalDiscrepancies returns true if there are any critical or high severity discrepancies
func (r *HPCReconciliationRecord) HasCriticalDiscrepancies() bool {
	for _, d := range r.Discrepancies {
		if d.Severity == "critical" || d.Severity == "high" {
			return true
		}
	}
	return false
}

// ReconciliationTolerances defines acceptable tolerances for reconciliation
type ReconciliationTolerances struct {
	// CPUCoreSecondsPercent is the tolerance for CPU core-seconds
	CPUCoreSecondsPercent float64 `json:"cpu_core_seconds_percent"`

	// MemoryGBSecondsPercent is the tolerance for memory
	MemoryGBSecondsPercent float64 `json:"memory_gb_seconds_percent"`

	// GPUSecondsPercent is the tolerance for GPU
	GPUSecondsPercent float64 `json:"gpu_seconds_percent"`

	// WallClockSecondsPercent is the tolerance for wall clock
	WallClockSecondsPercent float64 `json:"wall_clock_seconds_percent"`

	// NetworkBytesPercent is the tolerance for network
	NetworkBytesPercent float64 `json:"network_bytes_percent"`

	// StorageGBHoursPercent is the tolerance for storage
	StorageGBHoursPercent float64 `json:"storage_gb_hours_percent"`

	// TimeDriftSeconds is the allowed time drift
	TimeDriftSeconds int64 `json:"time_drift_seconds"`
}

// DefaultReconciliationTolerances returns default tolerances
func DefaultReconciliationTolerances() ReconciliationTolerances {
	return ReconciliationTolerances{
		CPUCoreSecondsPercent:   1.0, // 1% tolerance
		MemoryGBSecondsPercent:  1.0,
		GPUSecondsPercent:       0.5, // Tighter for expensive GPUs
		WallClockSecondsPercent: 0.1, // Very tight for wall clock
		NetworkBytesPercent:     5.0, // More tolerance for network
		StorageGBHoursPercent:   2.0,
		TimeDriftSeconds:        60, // 1 minute drift allowed
	}
}

// AuditTrailEntry represents an entry in the audit trail
type AuditTrailEntry struct {
	// EntryID is the unique identifier
	EntryID string `json:"entry_id"`

	// EntityType is the type of entity (job, accounting, reward, etc.)
	EntityType string `json:"entity_type"`

	// EntityID is the entity identifier
	EntityID string `json:"entity_id"`

	// Action is the action performed
	Action string `json:"action"`

	// ActorAddress is who performed the action
	ActorAddress string `json:"actor_address"`

	// ActorType is the actor type (provider, customer, system, governance)
	ActorType string `json:"actor_type"`

	// PreviousState is the previous state (JSON encoded)
	PreviousState string `json:"previous_state,omitempty"`

	// NewState is the new state (JSON encoded)
	NewState string `json:"new_state,omitempty"`

	// Reason explains the action
	Reason string `json:"reason,omitempty"`

	// RelatedEntities are related entity references
	RelatedEntities []string `json:"related_entities,omitempty"`

	// Timestamp is when the action occurred
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is when the entry was recorded
	BlockHeight int64 `json:"block_height"`

	// TxHash is the transaction hash
	TxHash string `json:"tx_hash,omitempty"`
}

// Validate validates the audit trail entry
func (e *AuditTrailEntry) Validate() error {
	if e.EntryID == "" {
		return fmt.Errorf("entry_id cannot be empty")
	}

	if e.EntityType == "" {
		return fmt.Errorf("entity_type cannot be empty")
	}

	if e.EntityID == "" {
		return fmt.Errorf("entity_id cannot be empty")
	}

	if e.Action == "" {
		return fmt.Errorf("action cannot be empty")
	}

	return nil
}
