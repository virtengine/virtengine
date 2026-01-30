// Package provider_daemon implements the VirtEngine provider daemon.
//
// VE-4D: HPCScheduler interface - unified abstraction for HPC schedulers
package provider_daemon

import (
	"context"
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// HPCScheduler defines the unified interface for HPC scheduler operations.
// This interface abstracts SLURM, MOAB, and OOD adapters behind a common API.
type HPCScheduler interface {
	// Type returns the scheduler type
	Type() HPCSchedulerType

	// Start starts the scheduler adapter
	Start(ctx context.Context) error

	// Stop stops the scheduler adapter
	Stop() error

	// IsRunning checks if the scheduler is running
	IsRunning() bool

	// SubmitJob submits a job to the scheduler
	SubmitJob(ctx context.Context, job *hpctypes.HPCJob) (*HPCSchedulerJob, error)

	// CancelJob cancels a job
	CancelJob(ctx context.Context, virtEngineJobID string) error

	// GetJobStatus gets the current job status
	GetJobStatus(ctx context.Context, virtEngineJobID string) (*HPCSchedulerJob, error)

	// GetJobAccounting gets job accounting/usage metrics
	GetJobAccounting(ctx context.Context, virtEngineJobID string) (*HPCSchedulerMetrics, error)

	// ListActiveJobs lists all active (non-terminal) jobs
	ListActiveJobs(ctx context.Context) ([]*HPCSchedulerJob, error)

	// RegisterLifecycleCallback registers a callback for job lifecycle events
	RegisterLifecycleCallback(cb HPCJobLifecycleCallback)

	// CreateStatusReport creates a signed status report for on-chain submission
	CreateStatusReport(job *HPCSchedulerJob) (*HPCStatusReport, error)
}

// HPCSchedulerJob represents a job tracked by the scheduler abstraction layer
type HPCSchedulerJob struct {
	// VirtEngineJobID is the on-chain job ID
	VirtEngineJobID string `json:"virtengine_job_id"`

	// SchedulerJobID is the native scheduler job ID (SLURM/MOAB/OOD)
	SchedulerJobID string `json:"scheduler_job_id"`

	// SchedulerType is the type of scheduler
	SchedulerType HPCSchedulerType `json:"scheduler_type"`

	// State is the unified job state
	State HPCJobState `json:"state"`

	// StateMessage contains additional state details
	StateMessage string `json:"state_message,omitempty"`

	// ExitCode is the job exit code (if completed)
	ExitCode int32 `json:"exit_code,omitempty"`

	// NodeList is the list of allocated nodes
	NodeList []string `json:"node_list,omitempty"`

	// SubmitTime is when the job was submitted
	SubmitTime time.Time `json:"submit_time"`

	// StartTime is when the job started running
	StartTime *time.Time `json:"start_time,omitempty"`

	// EndTime is when the job ended
	EndTime *time.Time `json:"end_time,omitempty"`

	// Metrics contains usage metrics
	Metrics *HPCSchedulerMetrics `json:"metrics,omitempty"`

	// OriginalJob is a reference to the original HPC job
	OriginalJob *hpctypes.HPCJob `json:"-"`
}

// HPCJobState represents the unified job state across all schedulers
type HPCJobState string

const (
	// HPCJobStatePending indicates job is pending/queued
	HPCJobStatePending HPCJobState = "pending"

	// HPCJobStateQueued indicates job is queued in scheduler
	HPCJobStateQueued HPCJobState = "queued"

	// HPCJobStateStarting indicates job is starting
	HPCJobStateStarting HPCJobState = "starting"

	// HPCJobStateRunning indicates job is running
	HPCJobStateRunning HPCJobState = "running"

	// HPCJobStateSuspended indicates job is suspended/paused
	HPCJobStateSuspended HPCJobState = "suspended"

	// HPCJobStateCompleted indicates job completed successfully
	HPCJobStateCompleted HPCJobState = "completed"

	// HPCJobStateFailed indicates job failed
	HPCJobStateFailed HPCJobState = "failed"

	// HPCJobStateCancelled indicates job was cancelled
	HPCJobStateCancelled HPCJobState = "cancelled"

	// HPCJobStateTimeout indicates job timed out
	HPCJobStateTimeout HPCJobState = "timeout"
)

// IsTerminal checks if the job state is terminal
func (s HPCJobState) IsTerminal() bool {
	switch s {
	case HPCJobStateCompleted, HPCJobStateFailed, HPCJobStateCancelled, HPCJobStateTimeout:
		return true
	default:
		return false
	}
}

// ToChainState converts to x/hpc job state
func (s HPCJobState) ToChainState() hpctypes.JobState {
	switch s {
	case HPCJobStatePending:
		return hpctypes.JobStatePending
	case HPCJobStateQueued:
		return hpctypes.JobStateQueued
	case HPCJobStateStarting, HPCJobStateRunning:
		return hpctypes.JobStateRunning
	case HPCJobStateSuspended:
		return hpctypes.JobStateRunning // No suspended state in chain
	case HPCJobStateCompleted:
		return hpctypes.JobStateCompleted
	case HPCJobStateFailed:
		return hpctypes.JobStateFailed
	case HPCJobStateCancelled:
		return hpctypes.JobStateCancelled
	case HPCJobStateTimeout:
		return hpctypes.JobStateTimeout
	default:
		return hpctypes.JobStatePending
	}
}

// HPCSchedulerMetrics contains usage metrics from the scheduler
type HPCSchedulerMetrics struct {
	// WallClockSeconds is wall clock time in seconds
	WallClockSeconds int64 `json:"wall_clock_seconds"`

	// CPUTimeSeconds is CPU time used in seconds
	CPUTimeSeconds int64 `json:"cpu_time_seconds"`

	// CPUCoreSeconds is CPU core-seconds used
	CPUCoreSeconds int64 `json:"cpu_core_seconds"`

	// MemoryBytesMax is maximum memory used in bytes
	MemoryBytesMax int64 `json:"memory_bytes_max"`

	// MemoryGBSeconds is memory GB-seconds used
	MemoryGBSeconds int64 `json:"memory_gb_seconds"`

	// GPUSeconds is GPU time used in seconds
	GPUSeconds int64 `json:"gpu_seconds,omitempty"`

	// NodesUsed is the number of nodes used
	NodesUsed int32 `json:"nodes_used"`

	// NodeHours is node-hours used
	NodeHours float64 `json:"node_hours"`

	// StorageGBHours is storage GB-hours used
	StorageGBHours int64 `json:"storage_gb_hours,omitempty"`

	// NetworkBytesIn is inbound network bytes
	NetworkBytesIn int64 `json:"network_bytes_in,omitempty"`

	// NetworkBytesOut is outbound network bytes
	NetworkBytesOut int64 `json:"network_bytes_out,omitempty"`

	// EnergyJoules is energy consumed in joules
	EnergyJoules int64 `json:"energy_joules,omitempty"`

	// SchedulerSpecific contains scheduler-specific metrics
	SchedulerSpecific map[string]interface{} `json:"scheduler_specific,omitempty"`
}

// ToChainMetrics converts to x/hpc usage metrics
func (m *HPCSchedulerMetrics) ToChainMetrics() hpctypes.HPCUsageMetrics {
	return hpctypes.HPCUsageMetrics{
		WallClockSeconds: m.WallClockSeconds,
		CPUCoreSeconds:   m.CPUCoreSeconds,
		MemoryGBSeconds:  m.MemoryGBSeconds,
		GPUSeconds:       m.GPUSeconds,
		StorageGBHours:   m.StorageGBHours,
		NetworkBytesIn:   m.NetworkBytesIn,
		NetworkBytesOut:  m.NetworkBytesOut,
		NodeHours:        int64(m.NodeHours),
		NodesUsed:        m.NodesUsed,
	}
}

// HPCJobLifecycleEvent represents a job lifecycle event
type HPCJobLifecycleEvent string

const (
	// HPCJobEventSubmitted is fired when job is submitted
	HPCJobEventSubmitted HPCJobLifecycleEvent = "submitted"

	// HPCJobEventQueued is fired when job is queued
	HPCJobEventQueued HPCJobLifecycleEvent = "queued"

	// HPCJobEventStarted is fired when job starts running
	HPCJobEventStarted HPCJobLifecycleEvent = "started"

	// HPCJobEventCompleted is fired when job completes
	HPCJobEventCompleted HPCJobLifecycleEvent = "completed"

	// HPCJobEventFailed is fired when job fails
	HPCJobEventFailed HPCJobLifecycleEvent = "failed"

	// HPCJobEventCancelled is fired when job is cancelled
	HPCJobEventCancelled HPCJobLifecycleEvent = "cancelled"

	// HPCJobEventTimeout is fired when job times out
	HPCJobEventTimeout HPCJobLifecycleEvent = "timeout"

	// HPCJobEventSuspended is fired when job is suspended
	HPCJobEventSuspended HPCJobLifecycleEvent = "suspended"

	// HPCJobEventResumed is fired when job is resumed
	HPCJobEventResumed HPCJobLifecycleEvent = "resumed"

	// HPCJobEventStateChanged is fired on any state change
	HPCJobEventStateChanged HPCJobLifecycleEvent = "state_changed"
)

// HPCJobLifecycleCallback is called during job lifecycle events
type HPCJobLifecycleCallback func(job *HPCSchedulerJob, event HPCJobLifecycleEvent, prevState HPCJobState)

// HPCStatusReport contains status report for on-chain submission
type HPCStatusReport struct {
	// ProviderAddress is the provider address
	ProviderAddress string `json:"provider_address"`

	// VirtEngineJobID is the VirtEngine job ID
	VirtEngineJobID string `json:"virtengine_job_id"`

	// SchedulerJobID is the scheduler job ID
	SchedulerJobID string `json:"scheduler_job_id"`

	// SchedulerType is the scheduler type
	SchedulerType HPCSchedulerType `json:"scheduler_type"`

	// State is the job state
	State HPCJobState `json:"state"`

	// StateMessage is the status message
	StateMessage string `json:"state_message,omitempty"`

	// ExitCode is the exit code
	ExitCode int32 `json:"exit_code,omitempty"`

	// Metrics are the usage metrics
	Metrics *HPCSchedulerMetrics `json:"metrics,omitempty"`

	// Timestamp is when the report was created
	Timestamp time.Time `json:"timestamp"`

	// Signature is the provider's signature (hex encoded)
	Signature string `json:"signature"`
}

// HPCSchedulerError represents an error from the scheduler
type HPCSchedulerError struct {
	// Code is the error code
	Code HPCErrorCode `json:"code"`

	// Message is the error message
	Message string `json:"message"`

	// Retryable indicates if the error is retryable
	Retryable bool `json:"retryable"`

	// Cause is the underlying error
	Cause error `json:"-"`
}

func (e *HPCSchedulerError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *HPCSchedulerError) Unwrap() error {
	return e.Cause
}

// HPCErrorCode represents error codes for HPC operations
type HPCErrorCode string

const (
	// HPCErrorCodeConnectionFailed indicates connection failure
	HPCErrorCodeConnectionFailed HPCErrorCode = "connection_failed"

	// HPCErrorCodeAuthenticationFailed indicates authentication failure
	HPCErrorCodeAuthenticationFailed HPCErrorCode = "authentication_failed"

	// HPCErrorCodeJobNotFound indicates job not found
	HPCErrorCodeJobNotFound HPCErrorCode = "job_not_found"

	// HPCErrorCodeJobSubmissionFailed indicates job submission failure
	HPCErrorCodeJobSubmissionFailed HPCErrorCode = "job_submission_failed"

	// HPCErrorCodeJobCancellationFailed indicates job cancellation failure
	HPCErrorCodeJobCancellationFailed HPCErrorCode = "job_cancellation_failed"

	// HPCErrorCodeInvalidJobSpec indicates invalid job specification
	HPCErrorCodeInvalidJobSpec HPCErrorCode = "invalid_job_spec"

	// HPCErrorCodeResourceUnavailable indicates resource unavailable
	HPCErrorCodeResourceUnavailable HPCErrorCode = "resource_unavailable"

	// HPCErrorCodeQuotaExceeded indicates quota exceeded
	HPCErrorCodeQuotaExceeded HPCErrorCode = "quota_exceeded"

	// HPCErrorCodeTimeout indicates operation timeout
	HPCErrorCodeTimeout HPCErrorCode = "timeout"

	// HPCErrorCodeInternal indicates internal error
	HPCErrorCodeInternal HPCErrorCode = "internal"
)

// IsRetryable returns true if the error code is typically retryable
func (c HPCErrorCode) IsRetryable() bool {
	switch c {
	case HPCErrorCodeConnectionFailed, HPCErrorCodeTimeout, HPCErrorCodeResourceUnavailable:
		return true
	default:
		return false
	}
}

// NewHPCSchedulerError creates a new HPCSchedulerError
func NewHPCSchedulerError(code HPCErrorCode, message string, cause error) *HPCSchedulerError {
	return &HPCSchedulerError{
		Code:      code,
		Message:   message,
		Retryable: code.IsRetryable(),
		Cause:     cause,
	}
}
