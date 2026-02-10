// Package types contains types for the HPC module.
//
// VE-500: HPC cluster template and QoS policy types
package types

import (
	"regexp"
	"time"
)

// ClusterTemplate represents a cluster configuration template
type ClusterTemplate struct {
	// TemplateID is the unique identifier for the template
	TemplateID string `json:"template_id"`

	// TemplateName is a human-readable name
	TemplateName string `json:"template_name"`

	// TemplateVersion is the semver version string
	TemplateVersion string `json:"template_version"`

	// Description provides details about the template
	Description string `json:"description,omitempty"`

	// Partitions lists the SLURM partition configurations
	Partitions []PartitionConfig `json:"partitions"`

	// QoSPolicies lists Quality of Service policies
	QoSPolicies []QoSPolicy `json:"qos_policies,omitempty"`

	// HardwareClasses defines available hardware classifications
	HardwareClasses HardwareClasses `json:"hardware_classes"`

	// ResourceLimits defines cluster-wide resource limits
	ResourceLimits ResourceLimits `json:"resource_limits"`

	// SchedulingPolicy defines scheduler configuration
	SchedulingPolicy SchedulingPolicy `json:"scheduling_policy"`

	// MaintenanceWindows defines scheduled maintenance periods
	MaintenanceWindows []MaintenanceWindow `json:"maintenance_windows,omitempty"`

	// CreatedAt is when the template was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the template was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// PartitionConfig represents a SLURM partition configuration
type PartitionConfig struct {
	// Name is the SLURM partition name
	Name string `json:"name"`

	// DisplayName is a user-friendly name
	DisplayName string `json:"display_name,omitempty"`

	// Nodes is the number of nodes in this partition
	Nodes int32 `json:"nodes"`

	// MaxNodesPerJob is the maximum nodes per job
	MaxNodesPerJob int32 `json:"max_nodes_per_job"`

	// MaxRuntimeSeconds is the maximum runtime in seconds
	MaxRuntimeSeconds int64 `json:"max_runtime_seconds"`

	// DefaultRuntimeSeconds is the default runtime in seconds
	DefaultRuntimeSeconds int64 `json:"default_runtime_seconds"`

	// Priority is the partition priority (0-1000)
	Priority int32 `json:"priority"`

	// Features lists node features
	Features []string `json:"features,omitempty"`

	// State is the partition state (up|down|drain|inactive)
	State string `json:"state"`

	// AccessControl defines access restrictions
	AccessControl *PartitionAccessControl `json:"access_control,omitempty"`
}

// PartitionAccessControl defines partition access restrictions
type PartitionAccessControl struct {
	// AllowGroups lists groups allowed to use this partition
	AllowGroups []string `json:"allow_groups,omitempty"`

	// DenyGroups lists groups denied from using this partition
	DenyGroups []string `json:"deny_groups,omitempty"`

	// RequireReservation indicates if reservation is required
	RequireReservation bool `json:"require_reservation"`
}

// QoSPolicy represents a Quality of Service policy
type QoSPolicy struct {
	// Name is the QoS name
	Name string `json:"name"`

	// Priority is the priority adjustment factor
	Priority int32 `json:"priority"`

	// MaxJobsPerUser is max running jobs per user (0=unlimited)
	MaxJobsPerUser int32 `json:"max_jobs_per_user"`

	// MaxSubmitJobsPerUser is max submitted jobs per user (0=unlimited)
	MaxSubmitJobsPerUser int32 `json:"max_submit_jobs_per_user"`

	// MaxWallDurationSeconds is the max wall-clock duration
	MaxWallDurationSeconds int64 `json:"max_wall_duration_seconds"`

	// MaxCPUsPerUser is max CPUs per user (0=unlimited)
	MaxCPUsPerUser int32 `json:"max_cpus_per_user"`

	// MaxGPUsPerUser is max GPUs per user (0=unlimited)
	MaxGPUsPerUser int32 `json:"max_gpus_per_user"`

	// MaxMemoryGBPerUser is max memory per user in GB (0=unlimited)
	MaxMemoryGBPerUser int32 `json:"max_memory_gb_per_user"`

	// PreemptMode is the preemption behavior (off|suspend|requeue|cancel)
	PreemptMode string `json:"preempt_mode"`

	// UsageFactor is the fair-share usage multiplier (fixed-point, 6 decimals)
	UsageFactor string `json:"usage_factor"`
}

// HardwareClasses contains hardware classification definitions
type HardwareClasses struct {
	// CPUClasses lists CPU node classes
	CPUClasses []CPUClass `json:"cpu_classes,omitempty"`

	// GPUClasses lists GPU node classes
	GPUClasses []GPUClass `json:"gpu_classes,omitempty"`

	// StorageClasses lists storage classes
	StorageClasses []StorageClass `json:"storage_classes,omitempty"`

	// NetworkClasses lists network classes
	NetworkClasses []NetworkClass `json:"network_classes,omitempty"`
}

// CPUClass represents a CPU hardware class
type CPUClass struct {
	// ClassID is the unique class identifier
	ClassID string `json:"class_id"`

	// Description describes the class
	Description string `json:"description,omitempty"`

	// CoresPerNode is the cores per node
	CoresPerNode int32 `json:"cores_per_node"`

	// MemoryGBPerNode is memory per node in GB
	MemoryGBPerNode int32 `json:"memory_gb_per_node"`

	// CPUModel is the CPU model name
	CPUModel string `json:"cpu_model"`

	// CPUGeneration is the CPU generation
	CPUGeneration string `json:"cpu_generation,omitempty"`

	// ThreadsPerCore is threads per core (1 or 2)
	ThreadsPerCore int32 `json:"threads_per_core"`

	// NUMANodes is the NUMA topology
	NUMANodes int32 `json:"numa_nodes"`

	// Features lists additional features
	Features []string `json:"features,omitempty"`
}

// GPUClass represents a GPU hardware class
type GPUClass struct {
	// ClassID is the unique class identifier
	ClassID string `json:"class_id"`

	// Description describes the class
	Description string `json:"description,omitempty"`

	// GPUModel is the GPU model name
	GPUModel string `json:"gpu_model"`

	// GPUCountPerNode is GPUs per node
	GPUCountPerNode int32 `json:"gpu_count_per_node"`

	// GPUMemoryGB is GPU memory in GB
	GPUMemoryGB int32 `json:"gpu_memory_gb"`

	// CUDAComputeCapability is the CUDA compute level
	CUDAComputeCapability string `json:"cuda_compute_capability,omitempty"`

	// NVLinkEnabled indicates NVLink support
	NVLinkEnabled bool `json:"nvlink_enabled"`

	// MIGSupported indicates Multi-Instance GPU support
	MIGSupported bool `json:"mig_supported"`

	// MIGProfiles lists available MIG profiles
	MIGProfiles []string `json:"mig_profiles,omitempty"`

	// Features lists additional features
	Features []string `json:"features,omitempty"`
}

// StorageClass represents a storage class
type StorageClass struct {
	// ClassID is the unique class identifier
	ClassID string `json:"class_id"`

	// Description describes the class
	Description string `json:"description,omitempty"`

	// StorageType is the storage type (nvme|ssd|hdd|lustre|gpfs|cephfs)
	StorageType string `json:"storage_type"`

	// CapacityTB is total capacity in TB
	CapacityTB int64 `json:"capacity_tb"`

	// IOPSRead is estimated read IOPS
	IOPSRead int64 `json:"iops_read,omitempty"`

	// IOPSWrite is estimated write IOPS
	IOPSWrite int64 `json:"iops_write,omitempty"`

	// BandwidthGbps is bandwidth in Gbps
	BandwidthGbps int64 `json:"bandwidth_gbps,omitempty"`

	// IsShared indicates if storage is shared
	IsShared bool `json:"is_shared"`

	// MountPath is the mount path
	MountPath string `json:"mount_path,omitempty"`
}

// NetworkClass represents a network class
type NetworkClass struct {
	// ClassID is the unique class identifier
	ClassID string `json:"class_id"`

	// Description describes the class
	Description string `json:"description,omitempty"`

	// NetworkType is the network type (infiniband|ethernet|roce)
	NetworkType string `json:"network_type"`

	// BandwidthGbps is bandwidth in Gbps
	BandwidthGbps int64 `json:"bandwidth_gbps"`

	// LatencyUs is estimated latency in microseconds
	LatencyUs int64 `json:"latency_us,omitempty"`

	// RDMAEnabled indicates RDMA support
	RDMAEnabled bool `json:"rdma_enabled"`
}

// ResourceLimits defines cluster-wide resource limits
type ResourceLimits struct {
	// MaxNodesPerJob is max nodes per job
	MaxNodesPerJob int32 `json:"max_nodes_per_job"`

	// MaxCPUsPerJob is max CPUs per job
	MaxCPUsPerJob int32 `json:"max_cpus_per_job"`

	// MaxGPUsPerJob is max GPUs per job
	MaxGPUsPerJob int32 `json:"max_gpus_per_job"`

	// MaxMemoryGBPerJob is max memory per job in GB
	MaxMemoryGBPerJob int32 `json:"max_memory_gb_per_job"`

	// MaxWallTimeSeconds is max wall time in seconds
	MaxWallTimeSeconds int64 `json:"max_wall_time_seconds"`

	// MaxJobsPerUser is max jobs per user
	MaxJobsPerUser int32 `json:"max_jobs_per_user"`

	// MaxConcurrentJobsPerUser is max concurrent jobs per user
	MaxConcurrentJobsPerUser int32 `json:"max_concurrent_jobs_per_user"`
}

// SchedulingPolicy defines scheduler configuration
type SchedulingPolicy struct {
	// SchedulerType is the scheduler type (slurm|pbs|custom)
	SchedulerType string `json:"scheduler_type"`

	// BackfillEnabled indicates if backfill scheduling is enabled
	BackfillEnabled bool `json:"backfill_enabled"`

	// PreemptionEnabled indicates if preemption is enabled
	PreemptionEnabled bool `json:"preemption_enabled"`

	// FairShareEnabled indicates if fair-share is enabled
	FairShareEnabled bool `json:"fair_share_enabled"`

	// FairShareDecayHalfLifeDays is the fair-share decay half-life
	FairShareDecayHalfLifeDays int32 `json:"fair_share_decay_half_life_days"`

	// PriorityWeightAge is the age priority weight
	PriorityWeightAge int32 `json:"priority_weight_age"`

	// PriorityWeightFairShare is the fair-share priority weight
	PriorityWeightFairShare int32 `json:"priority_weight_fair_share"`

	// PriorityWeightJobSize is the job size priority weight
	PriorityWeightJobSize int32 `json:"priority_weight_job_size"`

	// PriorityWeightPartition is the partition priority weight
	PriorityWeightPartition int32 `json:"priority_weight_partition"`

	// PriorityWeightQoS is the QoS priority weight
	PriorityWeightQoS int32 `json:"priority_weight_qos"`
}

// MaintenanceWindow defines a scheduled maintenance period
type MaintenanceWindow struct {
	// Name is the maintenance window name
	Name string `json:"name"`

	// StartCron is the cron expression for start time
	StartCron string `json:"start_cron"`

	// DurationHours is the duration in hours
	DurationHours int32 `json:"duration_hours"`

	// Timezone is the timezone (e.g., "UTC")
	Timezone string `json:"timezone"`
}

// Validation regex patterns
var (
	partitionNameRegex   = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]{0,62}$`)
	templateNameRegex    = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9-]{2,63}$`)
	semverRegex          = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	validPartitionStates = map[string]bool{"up": true, "down": true, "drain": true, "inactive": true}
	validPreemptModes    = map[string]bool{"off": true, "suspend": true, "requeue": true, "cancel": true}
	validStorageTypes    = map[string]bool{"nvme": true, "ssd": true, "hdd": true, "lustre": true, "gpfs": true, "cephfs": true}
	validNetworkTypes    = map[string]bool{"infiniband": true, "ethernet": true, "roce": true}
	validSchedulerTypes  = map[string]bool{"slurm": true, "pbs": true, "custom": true}
)

// Validate validates a cluster template
func (t *ClusterTemplate) Validate() error {
	if t.TemplateName == "" {
		return ErrInvalidClusterTemplate.Wrap("template_name required")
	}
	if !templateNameRegex.MatchString(t.TemplateName) {
		return ErrInvalidClusterTemplate.Wrap("template_name must be 3-64 chars, start with letter, alphanumeric with hyphens")
	}

	if t.TemplateVersion != "" && !semverRegex.MatchString(t.TemplateVersion) {
		return ErrInvalidClusterTemplate.Wrap("template_version must be valid semver")
	}

	if len(t.Description) > 1024 {
		return ErrInvalidClusterTemplate.Wrap("description exceeds 1024 characters")
	}

	if len(t.Partitions) == 0 {
		return ErrInvalidClusterTemplate.Wrap("at least one partition required")
	}

	for i, p := range t.Partitions {
		if err := p.Validate(); err != nil {
			return ErrInvalidClusterTemplate.Wrapf("partition[%d]: %s", i, err.Error())
		}
	}

	for i, q := range t.QoSPolicies {
		if err := q.Validate(); err != nil {
			return ErrInvalidClusterTemplate.Wrapf("qos_policy[%d]: %s", i, err.Error())
		}
	}

	return nil
}

// Validate validates a partition configuration
func (p *PartitionConfig) Validate() error {
	if p.Name == "" {
		return ErrInvalidPartition.Wrap("name required")
	}
	if !partitionNameRegex.MatchString(p.Name) {
		return ErrInvalidPartition.Wrap("invalid partition name format")
	}

	if p.Nodes < 1 {
		return ErrInvalidPartition.Wrap("nodes must be >= 1")
	}

	if p.MaxNodesPerJob > p.Nodes {
		return ErrInvalidPartition.Wrap("max_nodes_per_job cannot exceed nodes")
	}

	if p.MaxRuntimeSeconds < 60 {
		return ErrInvalidPartition.Wrap("max_runtime_seconds must be >= 60")
	}

	if p.DefaultRuntimeSeconds > p.MaxRuntimeSeconds {
		return ErrInvalidPartition.Wrap("default_runtime_seconds cannot exceed max_runtime_seconds")
	}

	if p.Priority < 0 || p.Priority > 1000 {
		return ErrInvalidPartition.Wrap("priority must be 0-1000")
	}

	if !validPartitionStates[p.State] {
		return ErrInvalidPartition.Wrapf("invalid state: %s", p.State)
	}

	return nil
}

// Validate validates a QoS policy
func (q *QoSPolicy) Validate() error {
	if q.Name == "" {
		return ErrInvalidQoSPolicy.Wrap("name required")
	}

	if q.MaxJobsPerUser < 0 {
		return ErrInvalidQoSPolicy.Wrap("max_jobs_per_user cannot be negative")
	}

	if q.MaxSubmitJobsPerUser < 0 {
		return ErrInvalidQoSPolicy.Wrap("max_submit_jobs_per_user cannot be negative")
	}

	if q.MaxCPUsPerUser < 0 {
		return ErrInvalidQoSPolicy.Wrap("max_cpus_per_user cannot be negative")
	}

	if q.MaxGPUsPerUser < 0 {
		return ErrInvalidQoSPolicy.Wrap("max_gpus_per_user cannot be negative")
	}

	if q.PreemptMode != "" && !validPreemptModes[q.PreemptMode] {
		return ErrInvalidQoSPolicy.Wrapf("invalid preempt_mode: %s", q.PreemptMode)
	}

	return nil
}

// Validate validates a CPU class
func (c *CPUClass) Validate() error {
	if c.ClassID == "" {
		return ErrInvalidHardwareClass.Wrap("class_id required")
	}
	if c.CoresPerNode < 1 {
		return ErrInvalidHardwareClass.Wrap("cores_per_node must be >= 1")
	}
	if c.MemoryGBPerNode < 1 {
		return ErrInvalidHardwareClass.Wrap("memory_gb_per_node must be >= 1")
	}
	if c.CPUModel == "" {
		return ErrInvalidHardwareClass.Wrap("cpu_model required")
	}
	if c.ThreadsPerCore != 0 && c.ThreadsPerCore != 1 && c.ThreadsPerCore != 2 {
		return ErrInvalidHardwareClass.Wrap("threads_per_core must be 1 or 2")
	}
	return nil
}

// Validate validates a GPU class
func (g *GPUClass) Validate() error {
	if g.ClassID == "" {
		return ErrInvalidHardwareClass.Wrap("class_id required")
	}
	if g.GPUModel == "" {
		return ErrInvalidHardwareClass.Wrap("gpu_model required")
	}
	if g.GPUCountPerNode < 1 {
		return ErrInvalidHardwareClass.Wrap("gpu_count_per_node must be >= 1")
	}
	if g.GPUMemoryGB < 1 {
		return ErrInvalidHardwareClass.Wrap("gpu_memory_gb must be >= 1")
	}
	return nil
}

// Validate validates a storage class
func (s *StorageClass) Validate() error {
	if s.ClassID == "" {
		return ErrInvalidHardwareClass.Wrap("class_id required")
	}
	if !validStorageTypes[s.StorageType] {
		return ErrInvalidHardwareClass.Wrapf("invalid storage_type: %s", s.StorageType)
	}
	if s.CapacityTB < 0 {
		return ErrInvalidHardwareClass.Wrap("capacity_tb cannot be negative")
	}
	return nil
}

// Validate validates a network class
func (n *NetworkClass) Validate() error {
	if n.ClassID == "" {
		return ErrInvalidHardwareClass.Wrap("class_id required")
	}
	if !validNetworkTypes[n.NetworkType] {
		return ErrInvalidHardwareClass.Wrapf("invalid network_type: %s", n.NetworkType)
	}
	if n.BandwidthGbps < 1 {
		return ErrInvalidHardwareClass.Wrap("bandwidth_gbps must be >= 1")
	}
	return nil
}

// Validate validates a scheduling policy
func (s *SchedulingPolicy) Validate() error {
	if s.SchedulerType != "" && !validSchedulerTypes[s.SchedulerType] {
		return ErrInvalidSchedulingPolicy.Wrapf("invalid scheduler_type: %s", s.SchedulerType)
	}
	if s.FairShareDecayHalfLifeDays < 0 {
		return ErrInvalidSchedulingPolicy.Wrap("fair_share_decay_half_life_days cannot be negative")
	}
	return nil
}
