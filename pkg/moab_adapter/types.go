// Package moab_adapter implements the MOAB workload manager adapter for VirtEngine.
//
// VE-917: MOAB workload manager using Waldur
package moab_adapter

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Error definitions
var (
	// ErrMOABNotConnected is returned when MOAB is not connected
	ErrMOABNotConnected = errors.New("MOAB not connected")

	// ErrJobNotFound is returned when a job is not found
	ErrJobNotFound = errors.New("job not found in MOAB")

	// ErrJobSubmissionFailed is returned when job submission fails
	ErrJobSubmissionFailed = errors.New("job submission failed")

	// ErrInvalidJobSpec is returned when a job spec is invalid
	ErrInvalidJobSpec = errors.New("invalid job specification")

	// ErrJobCancellationFailed is returned when job cancellation fails
	ErrJobCancellationFailed = errors.New("job cancellation failed")

	// ErrQueueNotFound is returned when a queue is not found
	ErrQueueNotFound = errors.New("queue not found")

	// ErrInvalidCredentials is returned when authentication fails
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// MOABJobState represents the state of a MOAB job
type MOABJobState string

const (
	// MOABJobStateIdle indicates job is idle/pending
	MOABJobStateIdle MOABJobState = "Idle"

	// MOABJobStateStarting indicates job is starting
	MOABJobStateStarting MOABJobState = "Starting"

	// MOABJobStateRunning indicates job is running
	MOABJobStateRunning MOABJobState = "Running"

	// MOABJobStateCompleted indicates job completed successfully
	MOABJobStateCompleted MOABJobState = "Completed"

	// MOABJobStateRemoved indicates job was removed
	MOABJobStateRemoved MOABJobState = "Removed"

	// MOABJobStateHold indicates job is on hold
	MOABJobStateHold MOABJobState = "Hold"

	// MOABJobStateSuspended indicates job is suspended
	MOABJobStateSuspended MOABJobState = "Suspended"

	// MOABJobStateVacated indicates job was vacated
	MOABJobStateVacated MOABJobState = "Vacated"

	// MOABJobStateCancelled indicates job was cancelled
	MOABJobStateCancelled MOABJobState = "Cancelled"

	// MOABJobStateDeferred indicates job is deferred
	MOABJobStateDeferred MOABJobState = "Deferred"

	// MOABJobStateFailed indicates job failed
	MOABJobStateFailed MOABJobState = "Failed"
)

// MOABConfig configures the MOAB adapter
type MOABConfig struct {
	// ServerHost is the MOAB server hostname
	ServerHost string `json:"server_host"`

	// ServerPort is the MOAB server port
	ServerPort int `json:"server_port"`

	// UseTLS enables TLS connection
	UseTLS bool `json:"use_tls"`

	// AuthMethod is the authentication method (password, key, kerberos)
	AuthMethod string `json:"auth_method"`

	// Username is the authentication username
	Username string `json:"-"` // Never log this

	// Password is the authentication password
	Password string `json:"-"` // Never log this

	// DefaultQueue is the default job queue
	DefaultQueue string `json:"default_queue"`

	// DefaultAccount is the default account for job submission
	DefaultAccount string `json:"default_account"`

	// JobPollInterval is how often to poll for job status
	JobPollInterval time.Duration `json:"job_poll_interval"`

	// ConnectionTimeout is the connection timeout
	ConnectionTimeout time.Duration `json:"connection_timeout"`

	// MaxRetries is the maximum retry attempts
	MaxRetries int `json:"max_retries"`

	// WaldurIntegration enables Waldur HPC integration
	WaldurIntegration bool `json:"waldur_integration"`

	// WaldurEndpoint is the Waldur API endpoint
	WaldurEndpoint string `json:"waldur_endpoint,omitempty"`
}

// DefaultMOABConfig returns the default MOAB configuration
func DefaultMOABConfig() MOABConfig {
	return MOABConfig{
		ServerHost:        "moab-server",
		ServerPort:        42559,
		UseTLS:            true,
		AuthMethod:        "password",
		DefaultQueue:      "batch",
		DefaultAccount:    "default",
		JobPollInterval:   time.Second * 15,
		ConnectionTimeout: time.Second * 30,
		MaxRetries:        3,
		WaldurIntegration: true,
	}
}

// MOABJobSpec defines a MOAB job specification
type MOABJobSpec struct {
	// JobName is the job name
	JobName string `json:"job_name"`

	// Queue is the MOAB queue (class)
	Queue string `json:"queue"`

	// Account is the account for job submission
	Account string `json:"account,omitempty"`

	// Nodes is the number of nodes requested
	Nodes int32 `json:"nodes"`

	// ProcsPerNode is processors per node
	ProcsPerNode int32 `json:"procs_per_node"`

	// MemoryMB is memory per node in MB
	MemoryMB int64 `json:"memory_mb"`

	// GPUs is the number of GPUs
	GPUs int32 `json:"gpus,omitempty"`

	// GPUType is the GPU type
	GPUType string `json:"gpu_type,omitempty"`

	// WallTimeLimit is wall time limit in seconds
	WallTimeLimit int64 `json:"wall_time_limit"`

	// WorkingDirectory is the working directory
	WorkingDirectory string `json:"working_directory"`

	// Executable is the executable path
	Executable string `json:"executable"`

	// Arguments are command arguments
	Arguments []string `json:"arguments,omitempty"`

	// Environment contains environment variables
	Environment map[string]string `json:"environment,omitempty"`

	// InputFiles are input file references
	InputFiles []string `json:"input_files,omitempty"`

	// OutputFile is the stdout file
	OutputFile string `json:"output_file,omitempty"`

	// ErrorFile is the stderr file
	ErrorFile string `json:"error_file,omitempty"`

	// EmailNotify enables email notification
	EmailNotify bool `json:"email_notify,omitempty"`

	// EmailAddress is the notification email
	EmailAddress string `json:"email_address,omitempty"`

	// Features are node feature requirements
	Features []string `json:"features,omitempty"`

	// Dependencies are job dependencies
	Dependencies []string `json:"dependencies,omitempty"`

	// Priority is the job priority (-1024 to 1024)
	Priority int32 `json:"priority,omitempty"`

	// Reservation is the reservation name
	Reservation string `json:"reservation,omitempty"`

	// Array is the job array specification (e.g., "1-10", "1,5,10")
	Array string `json:"array,omitempty"`
}

// Validate validates the job spec
func (s *MOABJobSpec) Validate() error {
	if s.JobName == "" {
		return errors.New("job name is required")
	}
	if s.Queue == "" {
		return errors.New("queue is required")
	}
	if s.Nodes < 1 {
		return errors.New("nodes must be at least 1")
	}
	if s.ProcsPerNode < 1 {
		return errors.New("procs_per_node must be at least 1")
	}
	if s.WallTimeLimit < 1 {
		return errors.New("wall_time_limit must be at least 1 second")
	}
	if s.Executable == "" {
		return errors.New("executable is required")
	}
	return nil
}

// ToMsubScript generates an msub-compatible script
func (s *MOABJobSpec) ToMsubScript() string {
	script := "#!/bin/bash\n"
	script += "#MSUB -N " + s.JobName + "\n"
	script += "#MSUB -q " + s.Queue + "\n"
	if s.Account != "" {
		script += "#MSUB -A " + s.Account + "\n"
	}
	script += "#MSUB -l nodes=" + formatInt(s.Nodes)
	if s.ProcsPerNode > 0 {
		script += ":ppn=" + formatInt(s.ProcsPerNode)
	}
	if s.GPUs > 0 {
		script += ":gpus=" + formatInt(s.GPUs)
	}
	script += "\n"
	if s.MemoryMB > 0 {
		script += "#MSUB -l mem=" + formatInt64(s.MemoryMB) + "mb\n"
	}
	script += "#MSUB -l walltime=" + formatDuration(s.WallTimeLimit) + "\n"
	if s.WorkingDirectory != "" {
		script += "#MSUB -d " + s.WorkingDirectory + "\n"
	}
	if s.OutputFile != "" {
		script += "#MSUB -o " + s.OutputFile + "\n"
	}
	if s.ErrorFile != "" {
		script += "#MSUB -e " + s.ErrorFile + "\n"
	}
	if s.EmailNotify && s.EmailAddress != "" {
		script += "#MSUB -m abe\n"
		script += "#MSUB -M " + s.EmailAddress + "\n"
	}
	if len(s.Dependencies) > 0 {
		for _, dep := range s.Dependencies {
			script += "#MSUB -l depend=" + dep + "\n"
		}
	}
	if s.Array != "" {
		script += "#MSUB -t " + s.Array + "\n"
	}
	// Environment variables
	for k, v := range s.Environment {
		script += "export " + k + "=\"" + v + "\"\n"
	}
	script += "\n"
	// Command
	script += s.Executable
	for _, arg := range s.Arguments {
		script += " " + arg
	}
	script += "\n"
	return script
}

// formatInt formats int32 to string
func formatInt(n int32) string {
	return fmt.Sprintf("%d", n)
}

// formatInt64 formats int64 to string
func formatInt64(n int64) string {
	return fmt.Sprintf("%d", n)
}

// formatDuration formats seconds to HH:MM:SS
func formatDuration(seconds int64) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// MOABJob represents a submitted MOAB job
type MOABJob struct {
	// MOABJobID is the MOAB job ID
	MOABJobID string `json:"moab_job_id"`

	// VirtEngineJobID is the VirtEngine job ID
	VirtEngineJobID string `json:"virtengine_job_id"`

	// Spec is the job specification
	Spec *MOABJobSpec `json:"spec"`

	// State is the current job state
	State MOABJobState `json:"state"`

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
	UsageMetrics *MOABUsageMetrics `json:"usage_metrics,omitempty"`

	// StatusMessage contains status details
	StatusMessage string `json:"status_message,omitempty"`

	// QueuePosition is position in queue (for pending jobs)
	QueuePosition int32 `json:"queue_position,omitempty"`

	// EligibleTime is when job becomes eligible
	EligibleTime *time.Time `json:"eligible_time,omitempty"`

	// CompletionCode is the MOAB completion code
	CompletionCode string `json:"completion_code,omitempty"`
}

// MOABUsageMetrics contains MOAB job usage metrics for VE rewards integration
type MOABUsageMetrics struct {
	// WallClockSeconds is wall clock time used
	WallClockSeconds int64 `json:"wall_clock_seconds"`

	// CPUTimeSeconds is CPU time used
	CPUTimeSeconds int64 `json:"cpu_time_seconds"`

	// MaxRSSBytes is max resident set size
	MaxRSSBytes int64 `json:"max_rss_bytes"`

	// MaxVMSizeBytes is max virtual memory size
	MaxVMSizeBytes int64 `json:"max_vm_size_bytes"`

	// GPUSeconds is GPU time used
	GPUSeconds int64 `json:"gpu_seconds,omitempty"`

	// SUSUsed is service units used (for billing)
	SUSUsed float64 `json:"sus_used"`

	// NodeHours is node hours used
	NodeHours float64 `json:"node_hours"`

	// EnergyJoules is energy consumed in joules
	EnergyJoules int64 `json:"energy_joules,omitempty"`
}

// MOABClient is the interface for MOAB operations
type MOABClient interface {
	// Connect connects to the MOAB server
	Connect(ctx context.Context) error

	// Disconnect disconnects from MOAB
	Disconnect() error

	// IsConnected checks if connected
	IsConnected() bool

	// SubmitJob submits a job using msub
	SubmitJob(ctx context.Context, spec *MOABJobSpec) (string, error)

	// CancelJob cancels a job using mjobctl -c
	CancelJob(ctx context.Context, moabJobID string) error

	// HoldJob puts a job on hold using mjobctl -h
	HoldJob(ctx context.Context, moabJobID string) error

	// ReleaseJob releases a held job using mjobctl -u
	ReleaseJob(ctx context.Context, moabJobID string) error

	// GetJobStatus gets job status using checkjob
	GetJobStatus(ctx context.Context, moabJobID string) (*MOABJob, error)

	// GetJobAccounting gets job accounting data
	GetJobAccounting(ctx context.Context, moabJobID string) (*MOABUsageMetrics, error)

	// ListQueues lists available queues using mdiag -q
	ListQueues(ctx context.Context) ([]QueueInfo, error)

	// ListNodes lists nodes using mdiag -n
	ListNodes(ctx context.Context) ([]NodeInfo, error)

	// GetClusterInfo gets cluster information using mdiag -s
	GetClusterInfo(ctx context.Context) (*ClusterInfo, error)

	// GetReservations lists reservations using mdiag -r
	GetReservations(ctx context.Context) ([]ReservationInfo, error)
}

// QueueInfo contains queue (class) information
type QueueInfo struct {
	Name         string   `json:"name"`
	State        string   `json:"state"`
	MaxNodes     int32    `json:"max_nodes"`
	MaxWalltime  int64    `json:"max_walltime"`
	DefaultNodes int32    `json:"default_nodes"`
	Priority     int32    `json:"priority"`
	Features     []string `json:"features,omitempty"`
	IdleJobs     int32    `json:"idle_jobs"`
	RunningJobs  int32    `json:"running_jobs"`
	HeldJobs     int32    `json:"held_jobs"`
}

// NodeInfo contains node information
type NodeInfo struct {
	Name         string   `json:"name"`
	State        string   `json:"state"`
	Processors   int32    `json:"processors"`
	MemoryMB     int64    `json:"memory_mb"`
	GPUs         int32    `json:"gpus,omitempty"`
	GPUType      string   `json:"gpu_type,omitempty"`
	Features     []string `json:"features,omitempty"`
	Load         float64  `json:"load"`
	AllocatedCPU int32    `json:"allocated_cpu"`
	AllocatedMem int64    `json:"allocated_mem"`
}

// ClusterInfo contains cluster information
type ClusterInfo struct {
	Name               string `json:"name"`
	TotalNodes         int32  `json:"total_nodes"`
	IdleNodes          int32  `json:"idle_nodes"`
	BusyNodes          int32  `json:"busy_nodes"`
	DownNodes          int32  `json:"down_nodes"`
	TotalProcessors    int32  `json:"total_processors"`
	IdleProcessors     int32  `json:"idle_processors"`
	RunningJobs        int32  `json:"running_jobs"`
	IdleJobs           int32  `json:"idle_jobs"`
	ActiveReservations int32  `json:"active_reservations"`
}

// ReservationInfo contains reservation information
type ReservationInfo struct {
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Nodes     []string  `json:"nodes"`
	Owner     string    `json:"owner"`
	State     string    `json:"state"`
}

// JobStatusReport contains job status for on-chain reporting
type JobStatusReport struct {
	// ProviderAddress is the provider address
	ProviderAddress string `json:"provider_address"`

	// VirtEngineJobID is the VirtEngine job ID
	VirtEngineJobID string `json:"virtengine_job_id"`

	// MOABJobID is the MOAB job ID
	MOABJobID string `json:"moab_job_id"`

	// State is the job state
	State MOABJobState `json:"state"`

	// StatusMessage is the status message
	StatusMessage string `json:"status_message,omitempty"`

	// ExitCode is the exit code
	ExitCode int32 `json:"exit_code,omitempty"`

	// UsageMetrics are the usage metrics
	UsageMetrics *MOABUsageMetrics `json:"usage_metrics,omitempty"`

	// Timestamp is when the report was created
	Timestamp time.Time `json:"timestamp"`

	// Signature is the provider's signature
	Signature string `json:"signature"`
}

// Hash generates a hash for signing
func (r *JobStatusReport) Hash() []byte {
	data := struct {
		ProviderAddress string `json:"provider_address"`
		VirtEngineJobID string `json:"virtengine_job_id"`
		MOABJobID       string `json:"moab_job_id"`
		State           string `json:"state"`
		ExitCode        int32  `json:"exit_code"`
		Timestamp       int64  `json:"timestamp"`
	}{
		ProviderAddress: r.ProviderAddress,
		VirtEngineJobID: r.VirtEngineJobID,
		MOABJobID:       r.MOABJobID,
		State:           string(r.State),
		ExitCode:        r.ExitCode,
		Timestamp:       r.Timestamp.Unix(),
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	hash := sha256.Sum256(bytes)
	return hash[:]
}

// VERewardsData contains data for VirtEngine rewards integration
type VERewardsData struct {
	// JobID is the VirtEngine job ID
	JobID string `json:"job_id"`

	// ClusterID is the HPC cluster ID
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the provider address
	ProviderAddress string `json:"provider_address"`

	// CustomerAddress is the customer address
	CustomerAddress string `json:"customer_address"`

	// SchedulerType is the scheduler type (MOAB)
	SchedulerType string `json:"scheduler_type"`

	// Usage contains usage metrics
	Usage *MOABUsageMetrics `json:"usage"`

	// StartTime is when the job started
	StartTime time.Time `json:"start_time"`

	// EndTime is when the job ended
	EndTime time.Time `json:"end_time"`

	// CompletionStatus is the completion status
	CompletionStatus string `json:"completion_status"`
}
