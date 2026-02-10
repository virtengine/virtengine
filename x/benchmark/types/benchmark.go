// Package types contains types for the Benchmark module.
//
// VE-600, VE-601: Benchmark types and schema
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// MetricSchemaVersion is the current version of the metric schema
const MetricSchemaVersion = "1.0.0"

// SupportedMetricSchemaVersions lists all supported schema versions
var SupportedMetricSchemaVersions = []string{"1.0.0"}

// CPUMetrics contains CPU benchmark metrics
type CPUMetrics struct {
	// SingleCoreScore is the single-core performance score (normalized 0-10000)
	SingleCoreScore int64 `json:"single_core_score"`

	// MultiCoreScore is the multi-core performance score (normalized 0-10000)
	MultiCoreScore int64 `json:"multi_core_score"`

	// CoreCount is the number of CPU cores
	CoreCount int32 `json:"core_count"`

	// ThreadCount is the number of threads
	ThreadCount int32 `json:"thread_count"`

	// BaseFrequencyMHz is the base clock frequency
	BaseFrequencyMHz int64 `json:"base_frequency_mhz"`

	// BoostFrequencyMHz is the boost clock frequency
	BoostFrequencyMHz int64 `json:"boost_frequency_mhz,omitempty"`
}

// MemoryMetrics contains memory benchmark metrics
type MemoryMetrics struct {
	// TotalGB is the total memory in GB
	TotalGB int64 `json:"total_gb"`

	// BandwidthMBps is the memory bandwidth in MB/s
	BandwidthMBps int64 `json:"bandwidth_mbps"`

	// LatencyNs is the memory latency in nanoseconds
	LatencyNs int64 `json:"latency_ns"`

	// Score is the normalized memory score (0-10000)
	Score int64 `json:"score"`
}

// DiskMetrics contains disk I/O benchmark metrics
type DiskMetrics struct {
	// ReadIOPS is the read I/O operations per second
	ReadIOPS int64 `json:"read_iops"`

	// WriteIOPS is the write I/O operations per second
	WriteIOPS int64 `json:"write_iops"`

	// ReadThroughputMBps is read throughput in MB/s
	ReadThroughputMBps int64 `json:"read_throughput_mbps"`

	// WriteThroughputMBps is write throughput in MB/s
	WriteThroughputMBps int64 `json:"write_throughput_mbps"`

	// TotalStorageGB is total storage capacity in GB
	TotalStorageGB int64 `json:"total_storage_gb"`

	// Score is the normalized disk score (0-10000)
	Score int64 `json:"score"`
}

// NetworkMetrics contains network benchmark metrics
type NetworkMetrics struct {
	// ThroughputMbps is network throughput in Mbps
	ThroughputMbps int64 `json:"throughput_mbps"`

	// LatencyMs is RTT latency in milliseconds (fixed-point: value * 1000)
	LatencyMs int64 `json:"latency_ms"`

	// PacketLossRate is packet loss rate (fixed-point: value * 1000000, 0-1000000)
	PacketLossRate int64 `json:"packet_loss_rate"`

	// ReferenceEndpoint is the endpoint used for testing
	ReferenceEndpoint string `json:"reference_endpoint"`

	// Score is the normalized network score (0-10000)
	Score int64 `json:"score"`
}

// GPUMetrics contains GPU benchmark metrics (optional)
type GPUMetrics struct {
	// Present indicates whether GPU metrics are included
	Present bool `json:"present"`

	// DeviceCount is the number of GPUs
	DeviceCount int32 `json:"device_count,omitempty"`

	// DeviceType is the GPU type (e.g., "nvidia-a100")
	DeviceType string `json:"device_type,omitempty"`

	// TotalMemoryGB is total GPU memory in GB
	TotalMemoryGB int64 `json:"total_memory_gb,omitempty"`

	// ComputeScore is the normalized compute score (0-10000)
	ComputeScore int64 `json:"compute_score,omitempty"`

	// MemoryBandwidthGBps is memory bandwidth in GB/s
	MemoryBandwidthGBps int64 `json:"memory_bandwidth_gbps,omitempty"`
}

// BenchmarkMetrics contains all benchmark metrics
type BenchmarkMetrics struct {
	// SchemaVersion is the version of the metric schema
	SchemaVersion string `json:"schema_version"`

	// CPU contains CPU metrics
	CPU CPUMetrics `json:"cpu"`

	// Memory contains memory metrics
	Memory MemoryMetrics `json:"memory"`

	// Disk contains disk I/O metrics
	Disk DiskMetrics `json:"disk"`

	// Network contains network metrics
	Network NetworkMetrics `json:"network"`

	// GPU contains optional GPU metrics
	GPU GPUMetrics `json:"gpu,omitempty"`
}

// NodeMetadata contains metadata about the benchmarked node
type NodeMetadata struct {
	// NodeID is the unique identifier for the node
	NodeID string `json:"node_id"`

	// Hostname is the node hostname (sanitized)
	Hostname string `json:"hostname,omitempty"`

	// Region is the geographic region
	Region string `json:"region"`

	// Zone is the availability zone
	Zone string `json:"zone,omitempty"`

	// Datacenter is the datacenter identifier
	Datacenter string `json:"datacenter,omitempty"`

	// OSType is the operating system type
	OSType string `json:"os_type"`

	// KernelVersion is the kernel version
	KernelVersion string `json:"kernel_version,omitempty"`
}

// BenchmarkReport is a signed benchmark report from a provider
type BenchmarkReport struct {
	// ReportID is the unique identifier for this report
	ReportID string `json:"report_id"`

	// ProviderAddress is the provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// ClusterID is the cluster being benchmarked
	ClusterID string `json:"cluster_id"`

	// OfferingID is the optional offering being benchmarked
	OfferingID string `json:"offering_id,omitempty"`

	// NodeMetadata contains node information
	NodeMetadata NodeMetadata `json:"node_metadata"`

	// SuiteVersion is the benchmark suite version
	SuiteVersion string `json:"suite_version"`

	// SuiteHash is the hash of the benchmark suite for reproducibility
	SuiteHash string `json:"suite_hash"`

	// Metrics contains the benchmark metrics
	Metrics BenchmarkMetrics `json:"metrics"`

	// SummaryScore is the normalized summary score (0-10000, fixed-point)
	SummaryScore int64 `json:"summary_score"`

	// Timestamp is when the benchmark was conducted
	Timestamp time.Time `json:"timestamp"`

	// Signature is the provider's signature over the report
	Signature string `json:"signature"`

	// PublicKey is the public key used for signing
	PublicKey string `json:"public_key"`

	// BlockHeight is when the report was submitted on-chain
	BlockHeight int64 `json:"block_height,omitempty"`

	// ChallengeID is set if this report is a challenge response
	ChallengeID string `json:"challenge_id,omitempty"`
}

// Hash computes the hash of the benchmark report for signing
func (r *BenchmarkReport) Hash() (string, error) {
	// Create a copy without signature for hashing
	copy := *r
	copy.Signature = ""
	copy.BlockHeight = 0

	data, err := json.Marshal(copy)
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// Validate validates the benchmark report
func (r *BenchmarkReport) Validate() error {
	if r.ReportID == "" {
		return ErrInvalidBenchmark.Wrap("report_id cannot be empty")
	}

	if len(r.ReportID) > 64 {
		return ErrInvalidBenchmark.Wrap("report_id exceeds maximum length")
	}

	if r.ProviderAddress == "" {
		return ErrInvalidBenchmark.Wrap("provider_address cannot be empty")
	}

	if r.ClusterID == "" {
		return ErrInvalidBenchmark.Wrap("cluster_id cannot be empty")
	}

	if r.SuiteVersion == "" {
		return ErrInvalidBenchmark.Wrap("suite_version cannot be empty")
	}

	if r.Signature == "" {
		return ErrInvalidBenchmark.Wrap("signature cannot be empty")
	}

	if r.PublicKey == "" {
		return ErrInvalidBenchmark.Wrap("public_key cannot be empty")
	}

	if !isSchemaVersionSupported(r.Metrics.SchemaVersion) {
		return ErrInvalidMetricSchema.Wrapf("unsupported schema version: %s", r.Metrics.SchemaVersion)
	}

	// Validate metric bounds
	if err := r.validateMetricBounds(); err != nil {
		return err
	}

	return nil
}

// validateMetricBounds validates that metrics are within acceptable bounds
func (r *BenchmarkReport) validateMetricBounds() error {
	// CPU score bounds: 0-10000
	if r.Metrics.CPU.SingleCoreScore < 0 || r.Metrics.CPU.SingleCoreScore > 10000 {
		return ErrOutOfBounds.Wrap("single_core_score out of bounds (0-10000)")
	}
	if r.Metrics.CPU.MultiCoreScore < 0 || r.Metrics.CPU.MultiCoreScore > 10000 {
		return ErrOutOfBounds.Wrap("multi_core_score out of bounds (0-10000)")
	}

	// Memory bounds
	if r.Metrics.Memory.Score < 0 || r.Metrics.Memory.Score > 10000 {
		return ErrOutOfBounds.Wrap("memory score out of bounds (0-10000)")
	}
	if r.Metrics.Memory.TotalGB <= 0 {
		return ErrOutOfBounds.Wrap("total_gb must be positive")
	}

	// Disk bounds
	if r.Metrics.Disk.Score < 0 || r.Metrics.Disk.Score > 10000 {
		return ErrOutOfBounds.Wrap("disk score out of bounds (0-10000)")
	}

	// Network bounds
	if r.Metrics.Network.Score < 0 || r.Metrics.Network.Score > 10000 {
		return ErrOutOfBounds.Wrap("network score out of bounds (0-10000)")
	}
	if r.Metrics.Network.PacketLossRate < 0 || r.Metrics.Network.PacketLossRate > 1000000 {
		return ErrOutOfBounds.Wrap("packet_loss_rate out of bounds (0-1000000)")
	}

	// GPU bounds (if present)
	if r.Metrics.GPU.Present {
		if r.Metrics.GPU.ComputeScore < 0 || r.Metrics.GPU.ComputeScore > 10000 {
			return ErrOutOfBounds.Wrap("gpu compute_score out of bounds (0-10000)")
		}
	}

	// Summary score bounds
	if r.SummaryScore < 0 || r.SummaryScore > 10000 {
		return ErrOutOfBounds.Wrap("summary_score out of bounds (0-10000)")
	}

	return nil
}

// isSchemaVersionSupported checks if a schema version is supported
func isSchemaVersionSupported(version string) bool {
	for _, v := range SupportedMetricSchemaVersions {
		if v == version {
			return true
		}
	}
	return false
}
