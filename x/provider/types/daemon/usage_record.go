// Package daemon provides on-chain types for provider daemon operations.
//
// VE-404: Usage metering + on-chain recording
package daemon

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// UsageRecordType distinguishes between periodic and final settlement records
type UsageRecordType uint8

const (
	// UsageRecordTypePeriodic is a periodic usage update
	UsageRecordTypePeriodic UsageRecordType = 1

	// UsageRecordTypeFinal is the final settlement record on termination
	UsageRecordTypeFinal UsageRecordType = 2
)

// String returns the string representation of UsageRecordType
func (t UsageRecordType) String() string {
	switch t {
	case UsageRecordTypePeriodic:
		return "periodic"
	case UsageRecordTypeFinal:
		return "final"
	default:
		return fmt.Sprintf("unknown(%d)", t)
	}
}

// UsageRecord represents a signed usage record from the provider daemon.
// This is stored on-chain and used for settlement and billing.
type UsageRecord struct {
	// RecordID is the unique identifier for this record
	RecordID string `json:"record_id"`

	// OrderID is the order this usage applies to
	OrderID string `json:"order_id"`

	// AllocationID is the allocation this usage applies to
	AllocationID string `json:"allocation_id"`

	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// RecordType indicates if this is periodic or final
	RecordType UsageRecordType `json:"record_type"`

	// TimeWindow specifies the time range for this usage record
	TimeWindow TimeWindow `json:"time_window"`

	// ResourceUsage contains the resource usage metrics
	ResourceUsage ResourceUsage `json:"resource_usage"`

	// PricingCalcInputs contains inputs used for price calculation
	PricingCalcInputs PricingCalcInputs `json:"pricing_calc_inputs"`

	// JobID is the workload job identifier (optional)
	JobID string `json:"job_id,omitempty"`

	// Signature is the daemon's cryptographic signature
	Signature DaemonSignature `json:"signature"`

	// CreatedAt is when the record was created
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is the block at which this was recorded
	BlockHeight int64 `json:"block_height"`
}

// TimeWindow represents a time range for usage metering
type TimeWindow struct {
	// StartTime is the start of the metering window
	StartTime time.Time `json:"start_time"`

	// EndTime is the end of the metering window
	EndTime time.Time `json:"end_time"`

	// DurationSeconds is the duration in seconds
	DurationSeconds int64 `json:"duration_seconds"`
}

// Validate validates the TimeWindow
func (tw TimeWindow) Validate() error {
	if tw.StartTime.IsZero() {
		return errors.New("start_time is required")
	}
	if tw.EndTime.IsZero() {
		return errors.New("end_time is required")
	}
	if tw.EndTime.Before(tw.StartTime) {
		return errors.New("end_time cannot be before start_time")
	}
	if tw.DurationSeconds < 0 {
		return errors.New("duration_seconds cannot be negative")
	}
	expectedDuration := int64(tw.EndTime.Sub(tw.StartTime).Seconds())
	if tw.DurationSeconds != expectedDuration {
		return fmt.Errorf("duration_seconds (%d) does not match time range (%d)", tw.DurationSeconds, expectedDuration)
	}
	return nil
}

// ResourceUsage contains resource usage metrics
type ResourceUsage struct {
	// CPUMillicores is the CPU usage in millicores
	CPUMillicores int64 `json:"cpu_millicores"`

	// CPUSeconds is the total CPU time in seconds
	CPUSeconds int64 `json:"cpu_seconds"`

	// MemoryBytesAvg is the average memory usage in bytes
	MemoryBytesAvg int64 `json:"memory_bytes_avg"`

	// MemoryBytesMax is the maximum memory usage in bytes
	MemoryBytesMax int64 `json:"memory_bytes_max"`

	// StorageBytesUsed is the storage used in bytes
	StorageBytesUsed int64 `json:"storage_bytes_used"`

	// StorageIOReadBytes is the total bytes read from storage
	StorageIOReadBytes int64 `json:"storage_io_read_bytes"`

	// StorageIOWriteBytes is the total bytes written to storage
	StorageIOWriteBytes int64 `json:"storage_io_write_bytes"`

	// NetworkIngressBytes is the total inbound network traffic in bytes
	NetworkIngressBytes int64 `json:"network_ingress_bytes"`

	// NetworkEgressBytes is the total outbound network traffic in bytes
	NetworkEgressBytes int64 `json:"network_egress_bytes"`

	// GPUMilliseconds is GPU usage time in milliseconds (optional)
	GPUMilliseconds int64 `json:"gpu_milliseconds,omitempty"`

	// GPUMemoryBytesMax is maximum GPU memory used (optional)
	GPUMemoryBytesMax int64 `json:"gpu_memory_bytes_max,omitempty"`
}

// Validate validates the ResourceUsage
func (ru ResourceUsage) Validate() error {
	if ru.CPUMillicores < 0 {
		return errors.New("cpu_millicores cannot be negative")
	}
	if ru.CPUSeconds < 0 {
		return errors.New("cpu_seconds cannot be negative")
	}
	if ru.MemoryBytesAvg < 0 {
		return errors.New("memory_bytes_avg cannot be negative")
	}
	if ru.MemoryBytesMax < 0 {
		return errors.New("memory_bytes_max cannot be negative")
	}
	if ru.MemoryBytesMax < ru.MemoryBytesAvg {
		return errors.New("memory_bytes_max cannot be less than memory_bytes_avg")
	}
	if ru.StorageBytesUsed < 0 {
		return errors.New("storage_bytes_used cannot be negative")
	}
	if ru.NetworkIngressBytes < 0 {
		return errors.New("network_ingress_bytes cannot be negative")
	}
	if ru.NetworkEgressBytes < 0 {
		return errors.New("network_egress_bytes cannot be negative")
	}
	return nil
}

// PricingCalcInputs contains inputs used for price calculation
type PricingCalcInputs struct {
	// PricingModel is the pricing model used
	PricingModel string `json:"pricing_model"`

	// BaseRatePerHour is the base hourly rate
	BaseRatePerHour string `json:"base_rate_per_hour"`

	// CPURatePerCore is the per-core CPU rate
	CPURatePerCore string `json:"cpu_rate_per_core"`

	// MemoryRatePerGB is the per-GB memory rate
	MemoryRatePerGB string `json:"memory_rate_per_gb"`

	// StorageRatePerGB is the per-GB storage rate
	StorageRatePerGB string `json:"storage_rate_per_gb"`

	// NetworkRatePerGB is the per-GB network rate
	NetworkRatePerGB string `json:"network_rate_per_gb"`

	// CalculatedAmount is the calculated cost for this period
	CalculatedAmount string `json:"calculated_amount"`

	// Currency is the currency/denom for the amounts
	Currency string `json:"currency"`
}

// Validate validates the PricingCalcInputs
func (p PricingCalcInputs) Validate() error {
	if p.PricingModel == "" {
		return errors.New("pricing_model is required")
	}
	if p.Currency == "" {
		return errors.New("currency is required")
	}
	return nil
}

// Validate validates the UsageRecord
func (ur UsageRecord) Validate() error {
	if ur.RecordID == "" {
		return errors.New("record_id is required")
	}
	if ur.OrderID == "" {
		return errors.New("order_id is required")
	}
	if ur.AllocationID == "" {
		return errors.New("allocation_id is required")
	}
	if ur.ProviderAddress == "" {
		return errors.New("provider_address is required")
	}
	if ur.RecordType != UsageRecordTypePeriodic && ur.RecordType != UsageRecordTypeFinal {
		return fmt.Errorf("invalid record_type: %d", ur.RecordType)
	}
	if err := ur.TimeWindow.Validate(); err != nil {
		return fmt.Errorf("time_window: %w", err)
	}
	if err := ur.ResourceUsage.Validate(); err != nil {
		return fmt.Errorf("resource_usage: %w", err)
	}
	if err := ur.PricingCalcInputs.Validate(); err != nil {
		return fmt.Errorf("pricing_calc_inputs: %w", err)
	}
	if err := ur.Signature.Validate(); err != nil {
		return fmt.Errorf("signature: %w", err)
	}
	return nil
}

// Hash computes the hash of the usage record content (excluding signature)
func (ur UsageRecord) Hash() ([]byte, error) {
	// Create a copy without signature for hashing
	recordForHash := struct {
		RecordID          string            `json:"record_id"`
		OrderID           string            `json:"order_id"`
		AllocationID      string            `json:"allocation_id"`
		ProviderAddress   string            `json:"provider_address"`
		RecordType        UsageRecordType   `json:"record_type"`
		TimeWindow        TimeWindow        `json:"time_window"`
		ResourceUsage     ResourceUsage     `json:"resource_usage"`
		PricingCalcInputs PricingCalcInputs `json:"pricing_calc_inputs"`
		JobID             string            `json:"job_id,omitempty"`
	}{
		RecordID:          ur.RecordID,
		OrderID:           ur.OrderID,
		AllocationID:      ur.AllocationID,
		ProviderAddress:   ur.ProviderAddress,
		RecordType:        ur.RecordType,
		TimeWindow:        ur.TimeWindow,
		ResourceUsage:     ur.ResourceUsage,
		PricingCalcInputs: ur.PricingCalcInputs,
		JobID:             ur.JobID,
	}

	data, err := json.Marshal(recordForHash)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal record for hashing: %w", err)
	}

	hash := sha256.Sum256(data)
	return hash[:], nil
}
