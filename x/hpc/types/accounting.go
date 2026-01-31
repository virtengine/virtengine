// Package types contains types for the HPC module.
//
// VE-5A: Usage accounting types for deterministic billing and rewards
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountingRecordStatus indicates the status of an accounting record
type AccountingRecordStatus string

const (
	// AccountingStatusPending indicates the record is pending finalization
	AccountingStatusPending AccountingRecordStatus = "pending"

	// AccountingStatusFinalized indicates the record is finalized
	AccountingStatusFinalized AccountingRecordStatus = "finalized"

	// AccountingStatusDisputed indicates the record is under dispute
	AccountingStatusDisputed AccountingRecordStatus = "disputed"

	// AccountingStatusSettled indicates the record is settled
	AccountingStatusSettled AccountingRecordStatus = "settled"

	// AccountingStatusCorrected indicates the record was corrected
	AccountingStatusCorrected AccountingRecordStatus = "corrected"
)

// IsValidAccountingRecordStatus checks if the status is valid
func IsValidAccountingRecordStatus(s AccountingRecordStatus) bool {
	switch s {
	case AccountingStatusPending, AccountingStatusFinalized, AccountingStatusDisputed,
		AccountingStatusSettled, AccountingStatusCorrected:
		return true
	default:
		return false
	}
}

// HPCAccountingRecord represents a detailed accounting record for HPC usage
type HPCAccountingRecord struct {
	// RecordID is the unique identifier
	RecordID string `json:"record_id"`

	// JobID is the HPC job ID
	JobID string `json:"job_id"`

	// ClusterID is the HPC cluster
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the provider
	ProviderAddress string `json:"provider_address"`

	// CustomerAddress is the customer
	CustomerAddress string `json:"customer_address"`

	// OfferingID is the HPC offering used
	OfferingID string `json:"offering_id"`

	// SchedulerType is the scheduler type (SLURM, MOAB, OOD)
	SchedulerType string `json:"scheduler_type"`

	// SchedulerJobID is the native scheduler job ID
	SchedulerJobID string `json:"scheduler_job_id,omitempty"`

	// UsageMetrics contains detailed usage metrics
	UsageMetrics HPCDetailedMetrics `json:"usage_metrics"`

	// BillableAmount is the calculated billable amount
	BillableAmount sdk.Coins `json:"billable_amount"`

	// BillableBreakdown shows cost per resource type
	BillableBreakdown BillableBreakdown `json:"billable_breakdown"`

	// AppliedDiscounts tracks discounts applied
	AppliedDiscounts []AppliedDiscount `json:"applied_discounts,omitempty"`

	// AppliedCaps tracks any caps applied
	AppliedCaps []AppliedCap `json:"applied_caps,omitempty"`

	// ProviderReward is the reward allocated to provider
	ProviderReward sdk.Coins `json:"provider_reward"`

	// PlatformFee is the platform fee
	PlatformFee sdk.Coins `json:"platform_fee"`

	// SignedUsageRecords are the signed usage record IDs
	SignedUsageRecords []string `json:"signed_usage_records"`

	// Status is the record status
	Status AccountingRecordStatus `json:"status"`

	// DisputeID links to dispute if disputed
	DisputeID string `json:"dispute_id,omitempty"`

	// CorrectedFromID links to original if corrected
	CorrectedFromID string `json:"corrected_from_id,omitempty"`

	// CorrectionReason explains the correction
	CorrectionReason string `json:"correction_reason,omitempty"`

	// SettlementID links to settlement record
	SettlementID string `json:"settlement_id,omitempty"`

	// InvoiceID links to invoice
	InvoiceID string `json:"invoice_id,omitempty"`

	// PeriodStart is the start of the accounting period
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the end of the accounting period
	PeriodEnd time.Time `json:"period_end"`

	// FormulaVersion is the billing formula version used
	FormulaVersion string `json:"formula_version"`

	// CalculationHash is deterministic hash of calculation inputs
	CalculationHash string `json:"calculation_hash"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`

	// FinalizedAt is when the record was finalized
	FinalizedAt *time.Time `json:"finalized_at,omitempty"`

	// SettledAt is when the record was settled
	SettledAt *time.Time `json:"settled_at,omitempty"`

	// BlockHeight is when the record was created
	BlockHeight int64 `json:"block_height"`
}

// HPCDetailedMetrics contains comprehensive usage metrics
type HPCDetailedMetrics struct {
	// Wall clock time in seconds
	WallClockSeconds int64 `json:"wall_clock_seconds"`

	// Queue time in seconds (time spent waiting)
	QueueTimeSeconds int64 `json:"queue_time_seconds"`

	// CPU core-seconds consumed
	CPUCoreSeconds int64 `json:"cpu_core_seconds"`

	// CPU time in seconds (actual CPU execution time)
	CPUTimeSeconds int64 `json:"cpu_time_seconds"`

	// Memory GB-seconds consumed
	MemoryGBSeconds int64 `json:"memory_gb_seconds"`

	// Peak memory in bytes
	MemoryBytesMax int64 `json:"memory_bytes_max"`

	// GPU seconds consumed
	GPUSeconds int64 `json:"gpu_seconds,omitempty"`

	// GPU type used
	GPUType string `json:"gpu_type,omitempty"`

	// Storage GB-hours consumed
	StorageGBHours int64 `json:"storage_gb_hours"`

	// Network bytes in
	NetworkBytesIn int64 `json:"network_bytes_in"`

	// Network bytes out
	NetworkBytesOut int64 `json:"network_bytes_out"`

	// Node hours consumed
	NodeHours sdkmath.LegacyDec `json:"node_hours"`

	// Nodes used
	NodesUsed int32 `json:"nodes_used"`

	// Energy consumption in joules (if available)
	EnergyJoules int64 `json:"energy_joules,omitempty"`

	// Submit time (for queue time calculation)
	SubmitTime time.Time `json:"submit_time"`

	// Start time (when job started running)
	StartTime *time.Time `json:"start_time,omitempty"`

	// End time (when job completed)
	EndTime *time.Time `json:"end_time,omitempty"`
}

// CalculateQueueTime calculates queue time from submit and start times
func (m *HPCDetailedMetrics) CalculateQueueTime() int64 {
	if m.StartTime == nil {
		return 0
	}
	queueTime := m.StartTime.Sub(m.SubmitTime)
	if queueTime < 0 {
		return 0
	}
	return int64(queueTime.Seconds())
}

// CalculateWallClock calculates wall clock from start and end times
func (m *HPCDetailedMetrics) CalculateWallClock() int64 {
	if m.StartTime == nil || m.EndTime == nil {
		return m.WallClockSeconds
	}
	wallClock := m.EndTime.Sub(*m.StartTime)
	if wallClock < 0 {
		return 0
	}
	return int64(wallClock.Seconds())
}

// ToLegacyMetrics converts to HPCUsageMetrics for backward compatibility
func (m *HPCDetailedMetrics) ToLegacyMetrics() HPCUsageMetrics {
	return HPCUsageMetrics{
		WallClockSeconds: m.WallClockSeconds,
		CPUCoreSeconds:   m.CPUCoreSeconds,
		MemoryGBSeconds:  m.MemoryGBSeconds,
		GPUSeconds:       m.GPUSeconds,
		StorageGBHours:   m.StorageGBHours,
		NetworkBytesIn:   m.NetworkBytesIn,
		NetworkBytesOut:  m.NetworkBytesOut,
		NodeHours:        m.NodeHours.TruncateInt64(),
		NodesUsed:        m.NodesUsed,
	}
}

// BillableBreakdown shows cost breakdown by resource type
type BillableBreakdown struct {
	// CPU cost
	CPUCost sdk.Coin `json:"cpu_cost"`

	// Memory cost
	MemoryCost sdk.Coin `json:"memory_cost"`

	// GPU cost
	GPUCost sdk.Coin `json:"gpu_cost"`

	// Storage cost
	StorageCost sdk.Coin `json:"storage_cost"`

	// Network cost
	NetworkCost sdk.Coin `json:"network_cost"`

	// Node cost (flat node-hour rate)
	NodeCost sdk.Coin `json:"node_cost"`

	// Queue penalty (for long queue times, may be credit)
	QueuePenalty sdk.Coin `json:"queue_penalty,omitempty"`

	// Subtotal before adjustments
	Subtotal sdk.Coins `json:"subtotal"`
}

// AppliedDiscount represents a discount applied to billing
type AppliedDiscount struct {
	// DiscountID is the discount identifier
	DiscountID string `json:"discount_id"`

	// DiscountType is the type (volume, loyalty, promo, etc.)
	DiscountType string `json:"discount_type"`

	// Description describes the discount
	Description string `json:"description"`

	// DiscountBps is the discount in basis points
	DiscountBps uint32 `json:"discount_bps"`

	// DiscountAmount is the absolute discount amount
	DiscountAmount sdk.Coins `json:"discount_amount"`

	// AppliedTo indicates what the discount was applied to
	AppliedTo string `json:"applied_to"`
}

// AppliedCap represents a billing cap that was applied
type AppliedCap struct {
	// CapID is the cap identifier
	CapID string `json:"cap_id"`

	// CapType is the type (daily, weekly, monthly, per-job)
	CapType string `json:"cap_type"`

	// Description describes the cap
	Description string `json:"description"`

	// CapAmount is the cap limit
	CapAmount sdk.Coins `json:"cap_amount"`

	// OriginalAmount is the amount before cap
	OriginalAmount sdk.Coins `json:"original_amount"`

	// CappedAmount is the amount saved by cap
	CappedAmount sdk.Coins `json:"capped_amount"`
}

// CalculateHash computes a deterministic hash of the accounting inputs
func (r *HPCAccountingRecord) CalculateHash() string {
	hashInput := struct {
		JobID           string             `json:"job_id"`
		ClusterID       string             `json:"cluster_id"`
		UsageMetrics    HPCDetailedMetrics `json:"usage_metrics"`
		PeriodStart     int64              `json:"period_start"`
		PeriodEnd       int64              `json:"period_end"`
		FormulaVersion  string             `json:"formula_version"`
		SignedRecords   []string           `json:"signed_records"`
	}{
		JobID:          r.JobID,
		ClusterID:      r.ClusterID,
		UsageMetrics:   r.UsageMetrics,
		PeriodStart:    r.PeriodStart.Unix(),
		PeriodEnd:      r.PeriodEnd.Unix(),
		FormulaVersion: r.FormulaVersion,
		SignedRecords:  r.SignedUsageRecords,
	}

	data, _ := json.Marshal(hashInput)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// Validate validates the accounting record
func (r *HPCAccountingRecord) Validate() error {
	if r.RecordID == "" {
		return ErrInvalidJobAccounting.Wrap("record_id cannot be empty")
	}

	if len(r.RecordID) > 64 {
		return ErrInvalidJobAccounting.Wrap("record_id exceeds maximum length")
	}

	if r.JobID == "" {
		return ErrInvalidJobAccounting.Wrap("job_id cannot be empty")
	}

	if r.ClusterID == "" {
		return ErrInvalidJobAccounting.Wrap("cluster_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(r.ProviderAddress); err != nil {
		return ErrInvalidJobAccounting.Wrap("invalid provider address")
	}

	if _, err := sdk.AccAddressFromBech32(r.CustomerAddress); err != nil {
		return ErrInvalidJobAccounting.Wrap("invalid customer address")
	}

	if !r.BillableAmount.IsValid() {
		return ErrInvalidJobAccounting.Wrap("billable_amount must be valid")
	}

	if !r.ProviderReward.IsValid() {
		return ErrInvalidJobAccounting.Wrap("provider_reward must be valid")
	}

	if !r.PlatformFee.IsValid() {
		return ErrInvalidJobAccounting.Wrap("platform_fee must be valid")
	}

	if !IsValidAccountingRecordStatus(r.Status) {
		return ErrInvalidJobAccounting.Wrapf("invalid status: %s", r.Status)
	}

	if r.PeriodEnd.Before(r.PeriodStart) {
		return ErrInvalidJobAccounting.Wrap("period_end must be after period_start")
	}

	if r.FormulaVersion == "" {
		return ErrInvalidJobAccounting.Wrap("formula_version cannot be empty")
	}

	return nil
}

// Finalize transitions the record to finalized status
func (r *HPCAccountingRecord) Finalize(now time.Time) error {
	if r.Status != AccountingStatusPending {
		return ErrInvalidJobAccounting.Wrapf("can only finalize pending records, current: %s", r.Status)
	}
	r.Status = AccountingStatusFinalized
	r.FinalizedAt = &now
	r.CalculationHash = r.CalculateHash()
	return nil
}

// MarkDisputed marks the record as disputed
func (r *HPCAccountingRecord) MarkDisputed(disputeID string) error {
	if r.Status == AccountingStatusSettled {
		return ErrInvalidJobAccounting.Wrap("cannot dispute settled record")
	}
	r.Status = AccountingStatusDisputed
	r.DisputeID = disputeID
	return nil
}

// Settle marks the record as settled
func (r *HPCAccountingRecord) Settle(settlementID string, now time.Time) error {
	if r.Status == AccountingStatusDisputed {
		return ErrInvalidJobAccounting.Wrap("cannot settle disputed record")
	}
	r.Status = AccountingStatusSettled
	r.SettlementID = settlementID
	r.SettledAt = &now
	return nil
}

// SchedulerAccountingData represents raw accounting data from a scheduler
type SchedulerAccountingData struct {
	// SchedulerType is the scheduler type (SLURM, MOAB, OOD)
	SchedulerType string `json:"scheduler_type"`

	// SchedulerJobID is the native job ID
	SchedulerJobID string `json:"scheduler_job_id"`

	// VirtEngineJobID is the VirtEngine job ID
	VirtEngineJobID string `json:"virtengine_job_id"`

	// RawMetrics contains the raw metrics from the scheduler
	RawMetrics map[string]interface{} `json:"raw_metrics"`

	// NormalizedMetrics contains normalized metrics
	NormalizedMetrics HPCDetailedMetrics `json:"normalized_metrics"`

	// AccountingExtractTime is when the data was extracted
	AccountingExtractTime time.Time `json:"accounting_extract_time"`

	// Signature is the provider's signature on the raw metrics
	Signature string `json:"signature,omitempty"`
}

// NormalizationRule defines how to normalize a scheduler metric
type NormalizationRule struct {
	// SourceField is the field name in raw metrics
	SourceField string `json:"source_field"`

	// TargetField is the target field in normalized metrics
	TargetField string `json:"target_field"`

	// ConversionFactor is the multiplier to apply
	ConversionFactor sdkmath.LegacyDec `json:"conversion_factor"`

	// Unit is the source unit
	Unit string `json:"unit"`

	// TargetUnit is the target unit
	TargetUnit string `json:"target_unit"`
}

// AccountingAggregation represents aggregated accounting for billing periods
type AccountingAggregation struct {
	// AggregationID is the unique identifier
	AggregationID string `json:"aggregation_id"`

	// CustomerAddress is the customer
	CustomerAddress string `json:"customer_address"`

	// ProviderAddress is the provider
	ProviderAddress string `json:"provider_address"`

	// ClusterID is the cluster (optional, for cluster-level aggregation)
	ClusterID string `json:"cluster_id,omitempty"`

	// PeriodStart is the aggregation period start
	PeriodStart time.Time `json:"period_start"`

	// PeriodEnd is the aggregation period end
	PeriodEnd time.Time `json:"period_end"`

	// TotalJobs is the number of jobs in this period
	TotalJobs int64 `json:"total_jobs"`

	// TotalCPUCoreHours is aggregated CPU core-hours
	TotalCPUCoreHours sdkmath.LegacyDec `json:"total_cpu_core_hours"`

	// TotalGPUHours is aggregated GPU-hours
	TotalGPUHours sdkmath.LegacyDec `json:"total_gpu_hours"`

	// TotalMemoryGBHours is aggregated memory GB-hours
	TotalMemoryGBHours sdkmath.LegacyDec `json:"total_memory_gb_hours"`

	// TotalStorageGBHours is aggregated storage GB-hours
	TotalStorageGBHours sdkmath.LegacyDec `json:"total_storage_gb_hours"`

	// TotalNodeHours is aggregated node-hours
	TotalNodeHours sdkmath.LegacyDec `json:"total_node_hours"`

	// TotalBillableAmount is the aggregated billable amount
	TotalBillableAmount sdk.Coins `json:"total_billable_amount"`

	// TotalDiscounts is the aggregated discount amount
	TotalDiscounts sdk.Coins `json:"total_discounts"`

	// AccountingRecordIDs lists the records in this aggregation
	AccountingRecordIDs []string `json:"accounting_record_ids"`

	// CreatedAt is when the aggregation was created
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is when the aggregation was created
	BlockHeight int64 `json:"block_height"`
}

// Validate validates the accounting aggregation
func (a *AccountingAggregation) Validate() error {
	if a.AggregationID == "" {
		return fmt.Errorf("aggregation_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(a.CustomerAddress); err != nil {
		return fmt.Errorf("invalid customer address: %w", err)
	}

	if _, err := sdk.AccAddressFromBech32(a.ProviderAddress); err != nil {
		return fmt.Errorf("invalid provider address: %w", err)
	}

	if a.PeriodEnd.Before(a.PeriodStart) {
		return fmt.Errorf("period_end must be after period_start")
	}

	return nil
}
