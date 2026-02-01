// Package slurm_adapter implements the SLURM orchestration adapter for VirtEngine.
//
// VE-501: SLURM orchestration adapter in Provider Daemon (v1)
package slurm_adapter

import (
	"context"
	"errors"
	"time"
)

// ErrSLURMNotConnected is returned when SLURM is not connected
var ErrSLURMNotConnected = errors.New("SLURM not connected")

// ErrJobNotFound is returned when a job is not found
var ErrJobNotFound = errors.New("job not found in SLURM")

// ErrJobSubmissionFailed is returned when job submission fails
var ErrJobSubmissionFailed = errors.New("job submission failed")

// ErrInvalidJobSpec is returned when a job spec is invalid
var ErrInvalidJobSpec = errors.New("invalid job specification")

// ErrJobCancellationFailed is returned when job cancellation fails
var ErrJobCancellationFailed = errors.New("job cancellation failed")

// SLURMJobState represents the state of a SLURM job
type SLURMJobState string

const (
	// SLURMJobStatePending indicates job is pending
	SLURMJobStatePending SLURMJobState = "PENDING"

	// SLURMJobStateRunning indicates job is running
	SLURMJobStateRunning SLURMJobState = "RUNNING"

	// SLURMJobStateCompleted indicates job completed successfully
	SLURMJobStateCompleted SLURMJobState = "COMPLETED"

	// SLURMJobStateFailed indicates job failed
	SLURMJobStateFailed SLURMJobState = "FAILED"

	// SLURMJobStateCancelled indicates job was cancelled
	SLURMJobStateCancelled SLURMJobState = "CANCELLED"

	// SLURMJobStateTimeout indicates job timed out
	SLURMJobStateTimeout SLURMJobState = "TIMEOUT"

	// SLURMJobStateSuspended indicates job is suspended
	SLURMJobStateSuspended SLURMJobState = "SUSPENDED"
)

// SLURMConfig configures the SLURM adapter
type SLURMConfig struct {
	// ClusterName is the SLURM cluster name
	ClusterName string `json:"cluster_name"`

	// ControllerHost is the SLURM controller hostname
	ControllerHost string `json:"controller_host"`

	// ControllerPort is the SLURM controller port
	ControllerPort int `json:"controller_port"`

	// AuthMethod is the authentication method (munge, jwt, etc.)
	AuthMethod string `json:"auth_method"`

	// AuthToken is the authentication token (if using jwt)
	AuthToken string `json:"-"` // Never log this

	// DefaultPartition is the default partition
	DefaultPartition string `json:"default_partition"`

	// JobPollInterval is how often to poll for job status
	JobPollInterval time.Duration `json:"job_poll_interval"`

	// ConnectionTimeout is the connection timeout
	ConnectionTimeout time.Duration `json:"connection_timeout"`

	// MaxRetries is the maximum retry attempts
	MaxRetries int `json:"max_retries"`
}

// DefaultSLURMConfig returns the default SLURM configuration
func DefaultSLURMConfig() SLURMConfig {
	return SLURMConfig{
		ClusterName:       "virtengine-hpc",
		ControllerHost:    "slurmctld",
		ControllerPort:    6817,
		AuthMethod:        "munge",
		DefaultPartition:  "default",
		JobPollInterval:   time.Second * 10,
		ConnectionTimeout: time.Second * 30,
		MaxRetries:        3,
	}
}

// SLURMJobSpec defines a SLURM job specification
type SLURMJobSpec struct {
	// JobName is the job name
	JobName string `json:"job_name"`

	// Partition is the SLURM partition
	Partition string `json:"partition"`

	// Nodes is the number of nodes
	Nodes int32 `json:"nodes"`

	// CPUsPerNode is CPUs per node
	CPUsPerNode int32 `json:"cpus_per_node"`

	// MemoryMB is memory per node in MB
	MemoryMB int64 `json:"memory_mb"`

	// GPUs is the number of GPUs
	GPUs int32 `json:"gpus,omitempty"`

	// GPUType is the GPU type
	GPUType string `json:"gpu_type,omitempty"`

	// TimeLimit is the time limit in minutes
	TimeLimit int64 `json:"time_limit"`

	// WorkingDirectory is the working directory
	WorkingDirectory string `json:"working_directory"`

	// ContainerImage is the container image
	ContainerImage string `json:"container_image"`

	// Command is the command to run
	Command string `json:"command"`

	// Arguments are command arguments
	Arguments []string `json:"arguments,omitempty"`

	// Environment contains environment variables
	Environment map[string]string `json:"environment,omitempty"`

	// InputFiles are input file references
	InputFiles []string `json:"input_files,omitempty"`

	// OutputDirectory is where to store outputs
	OutputDirectory string `json:"output_directory"`

	// Exclusive requests exclusive node access
	Exclusive bool `json:"exclusive"`

	// Constraints are node constraints
	Constraints []string `json:"constraints,omitempty"`
}

// Validate validates the job spec
func (s *SLURMJobSpec) Validate() error {
	if s.JobName == "" {
		return ErrInvalidJobSpec
	}
	if s.Nodes < 1 {
		return ErrInvalidJobSpec
	}
	if s.CPUsPerNode < 1 {
		return ErrInvalidJobSpec
	}
	if s.TimeLimit < 1 {
		return ErrInvalidJobSpec
	}
	if s.ContainerImage == "" && s.Command == "" {
		return ErrInvalidJobSpec
	}
	return nil
}

// SLURMJob represents a submitted SLURM job
type SLURMJob struct {
	// SLURMJobID is the SLURM job ID
	SLURMJobID string `json:"slurm_job_id"`

	// VirtEngineJobID is the VirtEngine job ID
	VirtEngineJobID string `json:"virtengine_job_id"`

	// Spec is the job specification
	Spec *SLURMJobSpec `json:"spec"`

	// State is the current job state
	State SLURMJobState `json:"state"`

	// NodeList is the list of allocated nodes
	NodeList []string `json:"node_list,omitempty"`

	// ExitCode is the job exit code
	ExitCode int32 `json:"exit_code,omitempty"`

	// StartTime is when the job started
	StartTime *time.Time `json:"start_time,omitempty"`

	// EndTime is when the job ended
	EndTime *time.Time `json:"end_time,omitempty"`

	// SubmitTime is when the job was submitted
	SubmitTime time.Time `json:"submit_time"`

	// UsageMetrics contains usage metrics
	UsageMetrics *SLURMUsageMetrics `json:"usage_metrics,omitempty"`

	// StatusMessage contains status details
	StatusMessage string `json:"status_message,omitempty"`
}

// SLURMUsageMetrics contains SLURM job usage metrics
type SLURMUsageMetrics struct {
	// WallClockSeconds is wall clock time
	WallClockSeconds int64 `json:"wall_clock_seconds"`

	// CPUTimeSeconds is CPU time used
	CPUTimeSeconds int64 `json:"cpu_time_seconds"`

	// MaxRSSBytes is max resident set size
	MaxRSSBytes int64 `json:"max_rss_bytes"`

	// MaxVMSizeBytes is max virtual memory size
	MaxVMSizeBytes int64 `json:"max_vm_size_bytes"`

	// GPUSeconds is GPU time used
	GPUSeconds int64 `json:"gpu_seconds,omitempty"`
}

// SLURMClient is the interface for SLURM operations
type SLURMClient interface {
	// Connect connects to the SLURM controller
	Connect(ctx context.Context) error

	// Disconnect disconnects from SLURM
	Disconnect() error

	// IsConnected checks if connected
	IsConnected() bool

	// SubmitJob submits a job to SLURM
	SubmitJob(ctx context.Context, spec *SLURMJobSpec) (string, error)

	// CancelJob cancels a job
	CancelJob(ctx context.Context, slurmJobID string) error

	// GetJobStatus gets job status
	GetJobStatus(ctx context.Context, slurmJobID string) (*SLURMJob, error)

	// GetJobAccounting gets job accounting data
	GetJobAccounting(ctx context.Context, slurmJobID string) (*SLURMUsageMetrics, error)

	// ListPartitions lists available partitions
	ListPartitions(ctx context.Context) ([]PartitionInfo, error)

	// ListNodes lists nodes in the cluster
	ListNodes(ctx context.Context) ([]NodeInfo, error)
}

// PartitionInfo contains partition information
type PartitionInfo struct {
	Name        string   `json:"name"`
	Nodes       int32    `json:"nodes"`
	State       string   `json:"state"`
	MaxTime     int64    `json:"max_time"`
	DefaultTime int64    `json:"default_time"`
	MaxNodes    int32    `json:"max_nodes"`
	Features    []string `json:"features,omitempty"`
}

// NodeInfo contains node information
type NodeInfo struct {
	Name       string   `json:"name"`
	State      string   `json:"state"`
	CPUs       int32    `json:"cpus"`
	MemoryMB   int64    `json:"memory_mb"`
	GPUs       int32    `json:"gpus,omitempty"`
	GPUType    string   `json:"gpu_type,omitempty"`
	Partitions []string `json:"partitions"`
	Features   []string `json:"features,omitempty"`
}

