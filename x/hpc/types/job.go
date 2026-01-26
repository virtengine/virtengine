// Package types contains types for the HPC module.
//
// VE-500: SLURM cluster lifecycle module - Job types and accounting schema
package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// JobState represents the state of an HPC job
type JobState string

const (
	// JobStatePending indicates the job is pending
	JobStatePending JobState = "pending"

	// JobStateQueued indicates the job is queued in SLURM
	JobStateQueued JobState = "queued"

	// JobStateRunning indicates the job is running
	JobStateRunning JobState = "running"

	// JobStateCompleted indicates the job completed successfully
	JobStateCompleted JobState = "completed"

	// JobStateFailed indicates the job failed
	JobStateFailed JobState = "failed"

	// JobStateCancelled indicates the job was cancelled
	JobStateCancelled JobState = "cancelled"

	// JobStateTimeout indicates the job timed out
	JobStateTimeout JobState = "timeout"
)

// IsValidJobState checks if the state is valid
func IsValidJobState(s JobState) bool {
	switch s {
	case JobStatePending, JobStateQueued, JobStateRunning, JobStateCompleted, JobStateFailed, JobStateCancelled, JobStateTimeout:
		return true
	default:
		return false
	}
}

// IsTerminalJobState checks if the state is terminal
func IsTerminalJobState(s JobState) bool {
	switch s {
	case JobStateCompleted, JobStateFailed, JobStateCancelled, JobStateTimeout:
		return true
	default:
		return false
	}
}

// HPCJob represents an HPC job request
type HPCJob struct {
	// JobID is the unique identifier for the job
	JobID string `json:"job_id"`

	// OfferingID is the HPC offering this job uses
	OfferingID string `json:"offering_id"`

	// ClusterID is the cluster this job runs on
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the provider handling this job
	ProviderAddress string `json:"provider_address"`

	// CustomerAddress is the customer who submitted the job
	CustomerAddress string `json:"customer_address"`

	// SLURMJobID is the SLURM job ID assigned by the cluster
	SLURMJobID string `json:"slurm_job_id,omitempty"`

	// State is the current job state
	State JobState `json:"state"`

	// QueueName is the SLURM queue/partition
	QueueName string `json:"queue_name"`

	// WorkloadSpec defines the workload
	WorkloadSpec JobWorkloadSpec `json:"workload_spec"`

	// Resources are the requested resources
	Resources JobResources `json:"resources"`

	// DataReferences are references to input data
	DataReferences []DataReference `json:"data_references,omitempty"`

	// EncryptedInputsPointer is the pointer to encrypted inputs
	EncryptedInputsPointer string `json:"encrypted_inputs_pointer,omitempty"`

	// EncryptedOutputsPointer is the pointer to encrypted outputs
	EncryptedOutputsPointer string `json:"encrypted_outputs_pointer,omitempty"`

	// MaxRuntimeSeconds is the maximum runtime
	MaxRuntimeSeconds int64 `json:"max_runtime_seconds"`

	// AgreedPrice is the agreed price for the job
	AgreedPrice sdk.Coins `json:"agreed_price"`

	// EscrowID links to the escrow for this job
	EscrowID string `json:"escrow_id,omitempty"`

	// SchedulingDecisionID links to the scheduling decision
	SchedulingDecisionID string `json:"scheduling_decision_id,omitempty"`

	// StatusMessage contains status details
	StatusMessage string `json:"status_message,omitempty"`

	// ExitCode is the job exit code (if completed)
	ExitCode int32 `json:"exit_code,omitempty"`

	// CreatedAt is when the job was created
	CreatedAt time.Time `json:"created_at"`

	// QueuedAt is when the job was queued
	QueuedAt *time.Time `json:"queued_at,omitempty"`

	// StartedAt is when the job started running
	StartedAt *time.Time `json:"started_at,omitempty"`

	// CompletedAt is when the job completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// BlockHeight is when the job was recorded
	BlockHeight int64 `json:"block_height"`
}

// JobWorkloadSpec defines the workload for an HPC job
type JobWorkloadSpec struct {
	// ContainerImage is the container image to use
	ContainerImage string `json:"container_image"`

	// Command is the command to run
	Command string `json:"command"`

	// Arguments are command arguments
	Arguments []string `json:"arguments,omitempty"`

	// Environment contains environment variables (non-sensitive)
	Environment map[string]string `json:"environment,omitempty"`

	// WorkingDirectory is the working directory
	WorkingDirectory string `json:"working_directory,omitempty"`

	// PreconfiguredWorkloadID references a preconfigured workload
	PreconfiguredWorkloadID string `json:"preconfigured_workload_id,omitempty"`

	// IsPreconfigured indicates if using a preconfigured workload
	IsPreconfigured bool `json:"is_preconfigured"`
}

// JobResources defines resource requirements for an HPC job
type JobResources struct {
	// Nodes is the number of nodes requested
	Nodes int32 `json:"nodes"`

	// CPUCoresPerNode is cores per node
	CPUCoresPerNode int32 `json:"cpu_cores_per_node"`

	// MemoryGBPerNode is memory per node in GB
	MemoryGBPerNode int32 `json:"memory_gb_per_node"`

	// GPUsPerNode is GPUs per node
	GPUsPerNode int32 `json:"gpus_per_node,omitempty"`

	// StorageGB is storage required
	StorageGB int32 `json:"storage_gb"`

	// GPUType is the required GPU type
	GPUType string `json:"gpu_type,omitempty"`
}

// DataReference references external data
type DataReference struct {
	// ReferenceID is a unique identifier for this reference
	ReferenceID string `json:"reference_id"`

	// Type is the data source type (e.g., "s3", "ipfs", "http")
	Type string `json:"type"`

	// URI is the data location (may be encrypted)
	URI string `json:"uri"`

	// Encrypted indicates if the URI is encrypted
	Encrypted bool `json:"encrypted"`

	// Checksum is the data checksum
	Checksum string `json:"checksum,omitempty"`

	// SizeBytes is the data size
	SizeBytes int64 `json:"size_bytes,omitempty"`
}

// Validate validates a job
func (j *HPCJob) Validate() error {
	if j.JobID == "" {
		return ErrInvalidJob.Wrap("job_id cannot be empty")
	}

	if len(j.JobID) > 64 {
		return ErrInvalidJob.Wrap("job_id exceeds maximum length")
	}

	if j.OfferingID == "" {
		return ErrInvalidJob.Wrap("offering_id cannot be empty")
	}

	if j.ClusterID == "" {
		return ErrInvalidJob.Wrap("cluster_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(j.ProviderAddress); err != nil {
		return ErrInvalidJob.Wrap("invalid provider address")
	}

	if _, err := sdk.AccAddressFromBech32(j.CustomerAddress); err != nil {
		return ErrInvalidJob.Wrap("invalid customer address")
	}

	if !IsValidJobState(j.State) {
		return ErrInvalidJob.Wrapf("invalid job state: %s", j.State)
	}

	if j.QueueName == "" {
		return ErrInvalidJob.Wrap("queue_name cannot be empty")
	}

	if j.WorkloadSpec.ContainerImage == "" && !j.WorkloadSpec.IsPreconfigured {
		return ErrInvalidJob.Wrap("container_image required for custom workloads")
	}

	if j.Resources.Nodes < 1 {
		return ErrInvalidJob.Wrap("nodes must be at least 1")
	}

	if j.MaxRuntimeSeconds < 60 {
		return ErrInvalidJob.Wrap("max_runtime_seconds must be at least 60")
	}

	return nil
}

// JobAccounting represents accounting data for an HPC job
type JobAccounting struct {
	// JobID is the job identifier
	JobID string `json:"job_id"`

	// ClusterID is the cluster that ran the job
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the provider
	ProviderAddress string `json:"provider_address"`

	// CustomerAddress is the customer
	CustomerAddress string `json:"customer_address"`

	// UsageMetrics contains usage metrics
	UsageMetrics HPCUsageMetrics `json:"usage_metrics"`

	// TotalCost is the total computed cost
	TotalCost sdk.Coins `json:"total_cost"`

	// ProviderReward is the reward for the provider
	ProviderReward sdk.Coins `json:"provider_reward"`

	// NodeRewards are rewards for individual nodes
	NodeRewards []NodeReward `json:"node_rewards,omitempty"`

	// PlatformFee is the platform fee
	PlatformFee sdk.Coins `json:"platform_fee"`

	// SettlementStatus indicates if this has been settled
	SettlementStatus string `json:"settlement_status"`

	// SettlementID links to the settlement record
	SettlementID string `json:"settlement_id,omitempty"`

	// SignedUsageRecordIDs are the signed usage record IDs
	SignedUsageRecordIDs []string `json:"signed_usage_record_ids"`

	// JobCompletionStatus indicates how the job completed
	JobCompletionStatus JobState `json:"job_completion_status"`

	// CreatedAt is when the accounting record was created
	CreatedAt time.Time `json:"created_at"`

	// FinalizedAt is when the accounting was finalized
	FinalizedAt *time.Time `json:"finalized_at,omitempty"`

	// BlockHeight is when the record was created
	BlockHeight int64 `json:"block_height"`
}

// HPCUsageMetrics contains usage metrics for an HPC job
type HPCUsageMetrics struct {
	// WallClockSeconds is the wall clock time in seconds
	WallClockSeconds int64 `json:"wall_clock_seconds"`

	// CPUCoreSeconds is CPU core-seconds used
	CPUCoreSeconds int64 `json:"cpu_core_seconds"`

	// MemoryGBSeconds is memory GB-seconds used
	MemoryGBSeconds int64 `json:"memory_gb_seconds"`

	// GPUSeconds is GPU-seconds used
	GPUSeconds int64 `json:"gpu_seconds,omitempty"`

	// StorageGBHours is storage GB-hours used
	StorageGBHours int64 `json:"storage_gb_hours"`

	// NetworkBytesIn is inbound network bytes
	NetworkBytesIn int64 `json:"network_bytes_in"`

	// NetworkBytesOut is outbound network bytes
	NetworkBytesOut int64 `json:"network_bytes_out"`

	// NodeHours is the node-hours used
	NodeHours int64 `json:"node_hours"`

	// NodesUsed is the number of nodes used
	NodesUsed int32 `json:"nodes_used"`
}

// NodeReward represents a reward for a specific node
type NodeReward struct {
	// NodeID is the node identifier
	NodeID string `json:"node_id"`

	// ProviderAddress is the provider/operator address
	ProviderAddress string `json:"provider_address"`

	// Amount is the reward amount
	Amount sdk.Coins `json:"amount"`

	// ContributionWeight is the node's contribution weight (fixed-point, 6 decimals)
	ContributionWeight string `json:"contribution_weight"`

	// UsageSeconds is the seconds this node was used
	UsageSeconds int64 `json:"usage_seconds"`
}

// Validate validates job accounting
func (ja *JobAccounting) Validate() error {
	if ja.JobID == "" {
		return ErrInvalidJobAccounting.Wrap("job_id cannot be empty")
	}

	if ja.ClusterID == "" {
		return ErrInvalidJobAccounting.Wrap("cluster_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(ja.ProviderAddress); err != nil {
		return ErrInvalidJobAccounting.Wrap("invalid provider address")
	}

	if _, err := sdk.AccAddressFromBech32(ja.CustomerAddress); err != nil {
		return ErrInvalidJobAccounting.Wrap("invalid customer address")
	}

	if !ja.TotalCost.IsValid() {
		return ErrInvalidJobAccounting.Wrap("total_cost must be valid")
	}

	return nil
}
