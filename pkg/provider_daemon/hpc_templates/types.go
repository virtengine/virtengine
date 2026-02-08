// Package hpc_templates provides workload template resolution for HPC jobs.
//
// VE-5F: Template resolution from on-chain registry to SLURM/container configs
package hpc_templates

import (
	"time"

	hpctypes "github.com/virtengine/virtengine/x/hpc/types"
)

// ResolvedJob represents a fully resolved job configuration from a template
type ResolvedJob struct {
	// TemplateID is the source template ID
	TemplateID string

	// TemplateVersion is the source template version
	TemplateVersion string

	// JobType is the resolved job type (slurm|container|native)
	JobType string

	// SlurmScript is the generated SLURM batch script (if JobType is slurm)
	SlurmScript string

	// ContainerConfig is the container runtime config (if JobType is container)
	ContainerConfig *ContainerRuntimeConfig

	// Resources is the resolved resource requirements
	Resources ResolvedResources

	// Environment is the resolved environment variables
	Environment map[string]string

	// DataBindings is the resolved data mount points
	DataBindings []ResolvedDataBinding

	// ResolvedAt is when this job was resolved
	ResolvedAt time.Time
}

// ContainerRuntimeConfig represents container runtime configuration
type ContainerRuntimeConfig struct {
	// RuntimeType is the runtime (singularity|apptainer)
	RuntimeType string

	// Image is the container image reference
	Image string

	// ImageDigest is the verified image digest
	ImageDigest string

	// Command is the command to execute
	Command string

	// Args are the command arguments
	Args []string

	// WorkingDir is the working directory
	WorkingDir string

	// Mounts is the list of volume mounts
	Mounts []ContainerMount

	// Environment is the environment variables
	Environment map[string]string

	// PreRunScript is the script to run before main command
	PreRunScript string

	// PostRunScript is the script to run after main command
	PostRunScript string

	// SecurityOptions are security-related options
	SecurityOptions *ContainerSecurityOptions
}

// ContainerMount represents a volume mount
type ContainerMount struct {
	// Source is the host path
	Source string

	// Target is the container path
	Target string

	// ReadOnly indicates if mount is read-only
	ReadOnly bool
}

// ContainerSecurityOptions represents security options
type ContainerSecurityOptions struct {
	// AllowNetworkAccess allows network access
	AllowNetworkAccess bool

	// AllowHostMounts allows host path mounts
	AllowHostMounts bool

	// SandboxLevel is the sandboxing level (none|basic|strict)
	SandboxLevel string

	// MaxOpenFiles is the max open files limit
	MaxOpenFiles int64

	// MaxProcesses is the max processes limit
	MaxProcesses int64

	// MaxFileSize is the max file size in bytes
	MaxFileSize int64
}

// ResolvedResources represents resolved resource requirements
type ResolvedResources struct {
	// Nodes is the number of compute nodes
	Nodes int32

	// CPUsPerNode is CPUs per node
	CPUsPerNode int32

	// MemoryMBPerNode is memory per node in MB
	MemoryMBPerNode int64

	// GPUsPerNode is GPUs per node (0 if not GPU job)
	GPUsPerNode int32

	// GPUTypes are the requested GPU types
	GPUTypes []string

	// RuntimeMinutes is the runtime in minutes
	RuntimeMinutes int64

	// StorageGB is required storage in GB
	StorageGB int32

	// ExclusiveNodes indicates if exclusive node access is required
	ExclusiveNodes bool

	// NetworkRequired indicates if high-speed network is required
	NetworkRequired bool
}

// ResolvedDataBinding represents a resolved data mount point
type ResolvedDataBinding struct {
	// Name is the binding name
	Name string

	// MountPath is the path inside the container/job
	MountPath string

	// HostPath is the host path
	HostPath string

	// DataType is the type of data (input|output|scratch)
	DataType string

	// ReadOnly indicates if the mount is read-only
	ReadOnly bool
}

// UserParameters represents user-provided parameters for template instantiation
type UserParameters struct {
	// Parameters is the map of parameter name to value
	Parameters map[string]string

	// Resources allows overriding resource defaults (within template limits)
	Resources *UserResourceOverrides

	// DataMappings maps data binding names to actual host paths
	DataMappings map[string]string
}

// UserResourceOverrides allows users to override resource defaults
type UserResourceOverrides struct {
	// Nodes overrides the default number of nodes (must be within min/max)
	Nodes *int32

	// CPUsPerNode overrides the default CPUs per node
	CPUsPerNode *int32

	// MemoryMBPerNode overrides the default memory per node
	MemoryMBPerNode *int64

	// GPUsPerNode overrides the default GPUs per node
	GPUsPerNode *int32

	// RuntimeMinutes overrides the default runtime
	RuntimeMinutes *int64
}

// TemplateResolutionResult represents the result of template resolution
type TemplateResolutionResult struct {
	// Success indicates if resolution succeeded
	Success bool

	// ResolvedJob is the resolved job (if success)
	ResolvedJob *ResolvedJob

	// Error is the error message (if !success)
	Error string

	// Warnings are non-fatal warnings during resolution
	Warnings []string
}

// TemplateResolver is the interface for template resolution
type TemplateResolver interface {
	// ResolveTemplate resolves a template to a runnable job configuration
	ResolveTemplate(template *hpctypes.WorkloadTemplate, userParams *UserParameters) (*TemplateResolutionResult, error)

	// ValidateTemplate validates a template can be resolved
	ValidateTemplate(template *hpctypes.WorkloadTemplate) error
}
