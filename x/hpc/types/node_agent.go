// Package types contains types for the HPC module.
//
// VE-500: HPC node agent heartbeat and protocol types
package types

import (
	"errors"
	"time"
)

// NodeState represents the state of a compute node
type NodeState string

const (
	// NodeStateUnknown indicates an unknown node state
	NodeStateUnknown NodeState = "unknown"

	// NodeStatePending indicates the node is pending registration
	NodeStatePending NodeState = "pending"

	// NodeStateActive indicates the node is active and available
	NodeStateActive NodeState = "active"

	// NodeStateStale indicates the node has missed heartbeats
	NodeStateStale NodeState = "stale"

	// NodeStateDraining indicates the node is draining jobs
	NodeStateDraining NodeState = "draining"

	// NodeStateDrained indicates the node has drained all jobs
	NodeStateDrained NodeState = "drained"

	// NodeStateOffline indicates the node is offline
	NodeStateOffline NodeState = "offline"

	// NodeStateDeregistered indicates the node has been deregistered
	NodeStateDeregistered NodeState = "deregistered"
)

// IsValidNodeState checks if the state is valid
func IsValidNodeState(s NodeState) bool {
	switch s {
	case NodeStateUnknown, NodeStatePending, NodeStateActive, NodeStateStale,
		NodeStateDraining, NodeStateDrained, NodeStateOffline, NodeStateDeregistered:
		return true
	default:
		return false
	}
}

// HealthStatus represents the health status of a node
type HealthStatus string

const (
	// HealthStatusHealthy indicates the node is healthy
	HealthStatusHealthy HealthStatus = "healthy"

	// HealthStatusDegraded indicates the node is degraded
	HealthStatusDegraded HealthStatus = "degraded"

	// HealthStatusUnhealthy indicates the node is unhealthy
	HealthStatusUnhealthy HealthStatus = "unhealthy"

	// HealthStatusDraining indicates the node is draining
	HealthStatusDraining HealthStatus = "draining"

	// HealthStatusOffline indicates the node is offline
	HealthStatusOffline HealthStatus = "offline"
)

// IsValidHealthStatus checks if the health status is valid
func IsValidHealthStatus(s HealthStatus) bool {
	switch s {
	case HealthStatusHealthy, HealthStatusDegraded, HealthStatusUnhealthy,
		HealthStatusDraining, HealthStatusOffline:
		return true
	default:
		return false
	}
}

// NodeIdentity represents the identity of a node agent
type NodeIdentity struct {
	// NodeID is the unique identifier for the node
	NodeID string `json:"node_id"`

	// ClusterID is the parent cluster identifier
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the owning provider's blockchain address
	ProviderAddress string `json:"provider_address"`

	// AgentPubkey is the base64-encoded Ed25519 public key
	AgentPubkey string `json:"agent_pubkey"`

	// Hostname is the FQDN or short hostname
	Hostname string `json:"hostname,omitempty"`

	// HardwareFingerprint is SHA256 of hardware identifiers
	HardwareFingerprint string `json:"hardware_fingerprint,omitempty"`

	// RegistrationNonce is a one-time use nonce (base64, 32 bytes)
	RegistrationNonce string `json:"registration_nonce,omitempty"`

	// RegistrationTimestamp is when the node was registered
	RegistrationTimestamp time.Time `json:"registration_timestamp"`

	// ProviderSignature is the provider's signature of the identity blob
	ProviderSignature string `json:"provider_signature,omitempty"`
}

// Validate validates a node identity
func (n *NodeIdentity) Validate() error {
	if n.NodeID == "" {
		return errors.New("node_id required")
	}
	if len(n.NodeID) < 4 || len(n.NodeID) > 64 {
		return errors.New("node_id must be 4-64 characters")
	}
	if n.ClusterID == "" {
		return errors.New("cluster_id required")
	}
	if n.ProviderAddress == "" {
		return errors.New("provider_address required")
	}
	if n.AgentPubkey == "" {
		return errors.New("agent_pubkey required")
	}
	return nil
}

// NodeHeartbeat represents a heartbeat from a node agent
type NodeHeartbeat struct {
	// NodeID is the node identifier
	NodeID string `json:"node_id"`

	// ClusterID is the cluster identifier
	ClusterID string `json:"cluster_id"`

	// SequenceNumber is monotonically increasing
	SequenceNumber uint64 `json:"sequence_number"`

	// Timestamp is the heartbeat timestamp (RFC3339)
	Timestamp time.Time `json:"timestamp"`

	// AgentVersion is the agent version (semver)
	AgentVersion string `json:"agent_version"`

	// Capacity contains capacity information
	Capacity NodeCapacity `json:"capacity"`

	// Health contains health information
	Health NodeHealth `json:"health"`

	// Latency contains latency measurements
	Latency NodeLatency `json:"latency"`

	// Jobs contains job information
	Jobs NodeJobs `json:"jobs"`

	// Services contains service status
	Services NodeServices `json:"services"`
}

// NodeCapacity contains node capacity information
type NodeCapacity struct {
	// CPUCoresTotal is the total CPU cores
	CPUCoresTotal int32 `json:"cpu_cores_total"`

	// CPUCoresAvailable is available CPU cores
	CPUCoresAvailable int32 `json:"cpu_cores_available"`

	// CPUCoresAllocated is allocated CPU cores
	CPUCoresAllocated int32 `json:"cpu_cores_allocated"`

	// MemoryGBTotal is total memory in GB
	MemoryGBTotal int32 `json:"memory_gb_total"`

	// MemoryGBAvailable is available memory in GB
	MemoryGBAvailable int32 `json:"memory_gb_available"`

	// MemoryGBAllocated is allocated memory in GB
	MemoryGBAllocated int32 `json:"memory_gb_allocated"`

	// GPUsTotal is total GPUs
	GPUsTotal int32 `json:"gpus_total"`

	// GPUsAvailable is available GPUs
	GPUsAvailable int32 `json:"gpus_available"`

	// GPUsAllocated is allocated GPUs
	GPUsAllocated int32 `json:"gpus_allocated"`

	// GPUType is the GPU type
	GPUType string `json:"gpu_type,omitempty"`

	// StorageGBTotal is total storage in GB
	StorageGBTotal int32 `json:"storage_gb_total"`

	// StorageGBAvailable is available storage in GB
	StorageGBAvailable int32 `json:"storage_gb_available"`

	// StorageGBAllocated is allocated storage in GB
	StorageGBAllocated int32 `json:"storage_gb_allocated"`
}

// NodeHealth contains node health information
type NodeHealth struct {
	// Status is the health status
	Status HealthStatus `json:"status"`

	// UptimeSeconds is the node uptime
	UptimeSeconds int64 `json:"uptime_seconds"`

	// LoadAverage1m is 1-minute load average (fixed-point, 6 decimals)
	LoadAverage1m string `json:"load_average_1m"`

	// LoadAverage5m is 5-minute load average (fixed-point, 6 decimals)
	LoadAverage5m string `json:"load_average_5m"`

	// LoadAverage15m is 15-minute load average (fixed-point, 6 decimals)
	LoadAverage15m string `json:"load_average_15m"`

	// CPUUtilizationPercent is CPU utilization (0-100)
	CPUUtilizationPercent int32 `json:"cpu_utilization_percent"`

	// MemoryUtilizationPercent is memory utilization (0-100)
	MemoryUtilizationPercent int32 `json:"memory_utilization_percent"`

	// GPUUtilizationPercent is GPU utilization (0-100)
	GPUUtilizationPercent int32 `json:"gpu_utilization_percent,omitempty"`

	// GPUMemoryUtilizationPercent is GPU memory utilization (0-100)
	GPUMemoryUtilizationPercent int32 `json:"gpu_memory_utilization_percent,omitempty"`

	// DiskIOUtilizationPercent is disk I/O utilization (0-100)
	DiskIOUtilizationPercent int32 `json:"disk_io_utilization_percent"`

	// NetworkUtilizationPercent is network utilization (0-100)
	NetworkUtilizationPercent int32 `json:"network_utilization_percent"`

	// TemperatureCelsius is CPU temperature
	TemperatureCelsius int32 `json:"temperature_celsius,omitempty"`

	// GPUTemperatureCelsius is GPU temperature
	GPUTemperatureCelsius int32 `json:"gpu_temperature_celsius,omitempty"`

	// ErrorCount24h is errors in the last 24 hours
	ErrorCount24h int32 `json:"error_count_24h"`

	// WarningCount24h is warnings in the last 24 hours
	WarningCount24h int32 `json:"warning_count_24h"`

	// LastErrorMessage is the last error message (max 256 chars)
	LastErrorMessage string `json:"last_error_message,omitempty"`

	// SLURMState is the SLURM node state (idle|allocated|mixed|down|drain|unknown)
	SLURMState string `json:"slurm_state,omitempty"`
}

// NodeLatency contains latency measurements
type NodeLatency struct {
	// Measurements contains latency measurements to other nodes
	Measurements []LatencyProbe `json:"measurements,omitempty"`

	// GatewayLatencyUs is latency to provider daemon in microseconds
	GatewayLatencyUs int64 `json:"gateway_latency_us"`

	// ChainLatencyMs is estimated latency to chain in milliseconds
	ChainLatencyMs int64 `json:"chain_latency_ms"`

	// AvgClusterLatencyUs is average cluster latency in microseconds
	AvgClusterLatencyUs int64 `json:"avg_cluster_latency_us"`
}

// LatencyProbe represents a latency measurement to another node
type LatencyProbe struct {
	// TargetNodeID is the target node
	TargetNodeID string `json:"target_node_id"`

	// LatencyUs is the measured latency in microseconds
	LatencyUs int64 `json:"latency_us"`

	// PacketLossPercent is the packet loss percentage (0-100)
	PacketLossPercent int32 `json:"packet_loss_percent"`

	// MeasuredAt is when the measurement was taken
	MeasuredAt time.Time `json:"measured_at"`
}

// NodeJobs contains job information
type NodeJobs struct {
	// RunningCount is the number of running jobs
	RunningCount int32 `json:"running_count"`

	// PendingCount is the number of pending jobs
	PendingCount int32 `json:"pending_count"`

	// Completed24h is jobs completed in last 24 hours
	Completed24h int32 `json:"completed_24h"`

	// Failed24h is jobs failed in last 24 hours
	Failed24h int32 `json:"failed_24h"`

	// ActiveJobIDs is the list of active SLURM job IDs
	ActiveJobIDs []string `json:"active_job_ids,omitempty"`
}

// NodeServices contains service status
type NodeServices struct {
	// SLURMDRunning indicates if slurmd is running
	SLURMDRunning bool `json:"slurmd_running"`

	// SLURMDVersion is the slurmd version
	SLURMDVersion string `json:"slurmd_version,omitempty"`

	// MungeRunning indicates if munge is running
	MungeRunning bool `json:"munge_running"`

	// ContainerRuntime is the container runtime (singularity|docker|podman|none)
	ContainerRuntime string `json:"container_runtime,omitempty"`

	// ContainerRuntimeVersion is the container runtime version
	ContainerRuntimeVersion string `json:"container_runtime_version,omitempty"`
}

// HeartbeatAuth contains authentication for a heartbeat
type HeartbeatAuth struct {
	// Signature is the base64 Ed25519 signature
	Signature string `json:"signature"`

	// Nonce is a base64 16-byte nonce
	Nonce string `json:"nonce"`

	// Timestamp is the Unix timestamp of the signature
	Timestamp int64 `json:"timestamp,omitempty"`
}

// HeartbeatResponse is the response to a heartbeat
type HeartbeatResponse struct {
	// Accepted indicates if the heartbeat was accepted
	Accepted bool `json:"accepted"`

	// SequenceAck is the acknowledged sequence number
	SequenceAck uint64 `json:"sequence_ack"`

	// Timestamp is the response timestamp
	Timestamp time.Time `json:"timestamp"`

	// NextHeartbeatSeconds is the suggested next interval
	NextHeartbeatSeconds int32 `json:"next_heartbeat_seconds"`

	// Commands are commands for the node to execute
	Commands []NodeCommand `json:"commands,omitempty"`

	// ConfigUpdates are configuration updates
	ConfigUpdates *HeartbeatConfigUpdate `json:"config_updates,omitempty"`

	// Errors are any error responses
	Errors []HeartbeatError `json:"errors,omitempty"`
}

// NodeCommand represents a command for the node agent
type NodeCommand struct {
	// CommandID is the unique command identifier
	CommandID string `json:"command_id"`

	// Type is the command type (drain|resume|shutdown|update_agent|run_diagnostic)
	Type string `json:"type"`

	// Parameters contains command parameters
	Parameters map[string]string `json:"parameters,omitempty"`

	// Deadline is the command deadline
	Deadline time.Time `json:"deadline,omitempty"`
}

// HeartbeatConfigUpdate contains configuration updates
type HeartbeatConfigUpdate struct {
	// SamplingIntervalSeconds is the new sampling interval
	SamplingIntervalSeconds int32 `json:"sampling_interval_seconds,omitempty"`

	// LatencyProbeTargets are nodes to measure latency to
	LatencyProbeTargets []string `json:"latency_probe_targets,omitempty"`

	// MetricsRetentionHours is the metrics retention period
	MetricsRetentionHours int32 `json:"metrics_retention_hours,omitempty"`
}

// HeartbeatError represents an error in heartbeat processing
type HeartbeatError struct {
	// Code is the error code
	Code string `json:"code"`

	// Message is the error message
	Message string `json:"message"`
}

// Validate validates a node heartbeat
func (h *NodeHeartbeat) Validate() error {
	if h.NodeID == "" {
		return ErrInvalidHeartbeat.Wrap("node_id required")
	}
	if h.ClusterID == "" {
		return ErrInvalidHeartbeat.Wrap("cluster_id required")
	}
	if h.SequenceNumber == 0 {
		return ErrInvalidHeartbeat.Wrap("sequence_number must be > 0")
	}

	// Timestamp validation (not too far in past or future)
	now := time.Now()
	if h.Timestamp.Before(now.Add(-5 * time.Minute)) {
		return ErrInvalidHeartbeat.Wrap("timestamp too old")
	}
	if h.Timestamp.After(now.Add(1 * time.Minute)) {
		return ErrInvalidHeartbeat.Wrap("timestamp in future")
	}

	if err := h.Capacity.Validate(); err != nil {
		return ErrInvalidHeartbeat.Wrapf("capacity: %s", err.Error())
	}

	if err := h.Health.Validate(); err != nil {
		return ErrInvalidHeartbeat.Wrapf("health: %s", err.Error())
	}

	return nil
}

// Validate validates node capacity
func (c *NodeCapacity) Validate() error {
	if c.CPUCoresTotal < 1 {
		return errors.New("cpu_cores_total must be >= 1")
	}
	if c.CPUCoresAvailable < 0 || c.CPUCoresAvailable > c.CPUCoresTotal {
		return errors.New("cpu_cores_available out of range")
	}
	if c.MemoryGBTotal < 1 {
		return errors.New("memory_gb_total must be >= 1")
	}
	if c.MemoryGBAvailable < 0 || c.MemoryGBAvailable > c.MemoryGBTotal {
		return errors.New("memory_gb_available out of range")
	}
	return nil
}

// Validate validates node health
func (h *NodeHealth) Validate() error {
	if !IsValidHealthStatus(h.Status) {
		return errors.New("invalid health status")
	}
	if h.CPUUtilizationPercent < 0 || h.CPUUtilizationPercent > 100 {
		return errors.New("cpu_utilization_percent out of range")
	}
	if h.MemoryUtilizationPercent < 0 || h.MemoryUtilizationPercent > 100 {
		return errors.New("memory_utilization_percent out of range")
	}
	if len(h.LastErrorMessage) > 256 {
		return errors.New("last_error_message exceeds 256 characters")
	}
	return nil
}

// TTL and expiry configuration
const (
	// DefaultHeartbeatInterval is the default heartbeat interval
	DefaultHeartbeatInterval = 30 * time.Second

	// DefaultHeartbeatTimeout is the time before node marked stale
	DefaultHeartbeatTimeout = 120 * time.Second

	// DefaultOfflineThreshold is the time before node marked offline
	DefaultOfflineThreshold = 300 * time.Second

	// DefaultDeregistrationDelay is the time before automatic deregistration
	DefaultDeregistrationDelay = 3600 * time.Second

	// MinHeartbeatInterval is the minimum heartbeat interval
	MinHeartbeatInterval = 10 * time.Second

	// MaxHeartbeatInterval is the maximum heartbeat interval
	MaxHeartbeatInterval = 120 * time.Second

	// MinHeartbeatTimeout is the minimum heartbeat timeout
	MinHeartbeatTimeout = 60 * time.Second

	// MaxHeartbeatTimeout is the maximum heartbeat timeout
	MaxHeartbeatTimeout = 600 * time.Second
)

// SamplingConfig defines metrics sampling configuration
type SamplingConfig struct {
	// BaseIntervalSeconds is the base sampling interval
	BaseIntervalSeconds int32 `json:"base_interval_seconds"`

	// CapacityIntervalSeconds is the capacity sampling interval
	CapacityIntervalSeconds int32 `json:"capacity_interval_seconds"`

	// HealthIntervalSeconds is the health sampling interval
	HealthIntervalSeconds int32 `json:"health_interval_seconds"`

	// UtilizationIntervalSeconds is the utilization sampling interval
	UtilizationIntervalSeconds int32 `json:"utilization_interval_seconds"`

	// LatencyIntervalSeconds is the latency sampling interval
	LatencyIntervalSeconds int32 `json:"latency_interval_seconds"`

	// LatencyProbeCount is the number of latency probes per measurement
	LatencyProbeCount int32 `json:"latency_probe_count"`

	// LatencyProbeTimeoutMs is the latency probe timeout in milliseconds
	LatencyProbeTimeoutMs int32 `json:"latency_probe_timeout_ms"`

	// MetricsBufferSize is the metrics buffer size
	MetricsBufferSize int32 `json:"metrics_buffer_size"`

	// BatchSubmitIntervalSeconds is the batch submit interval
	BatchSubmitIntervalSeconds int32 `json:"batch_submit_interval_seconds"`

	// BatchMaxSize is the maximum batch size
	BatchMaxSize int32 `json:"batch_max_size"`
}

// DefaultSamplingConfig returns the default sampling configuration
func DefaultSamplingConfig() SamplingConfig {
	return SamplingConfig{
		BaseIntervalSeconds:        30,
		CapacityIntervalSeconds:    30,
		HealthIntervalSeconds:      30,
		UtilizationIntervalSeconds: 15,
		LatencyIntervalSeconds:     300,
		LatencyProbeCount:          3,
		LatencyProbeTimeoutMs:      1000,
		MetricsBufferSize:          100,
		BatchSubmitIntervalSeconds: 60,
		BatchMaxSize:               50,
	}
}

// Minimum sampling intervals (in seconds)
const (
	// MinCapacitySamplingInterval is the minimum capacity sampling interval
	MinCapacitySamplingInterval = 10

	// MinHealthSamplingInterval is the minimum health sampling interval
	MinHealthSamplingInterval = 10

	// MinUtilizationSamplingInterval is the minimum utilization sampling interval
	MinUtilizationSamplingInterval = 5

	// MinLatencySamplingInterval is the minimum latency sampling interval
	MinLatencySamplingInterval = 60

	// MinJobCountSamplingInterval is the minimum job count sampling interval
	MinJobCountSamplingInterval = 30

	// MinTemperatureSamplingInterval is the minimum temperature sampling interval
	MinTemperatureSamplingInterval = 30
)
