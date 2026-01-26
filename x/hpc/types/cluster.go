// Package types contains types for the HPC module.
//
// VE-500: SLURM cluster lifecycle module - HPC Offering and Cluster types
package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ClusterState represents the state of an HPC cluster
type ClusterState string

const (
	// ClusterStatePending indicates the cluster is pending registration
	ClusterStatePending ClusterState = "pending"

	// ClusterStateActive indicates the cluster is active and accepting jobs
	ClusterStateActive ClusterState = "active"

	// ClusterStateDraining indicates the cluster is draining (not accepting new jobs)
	ClusterStateDraining ClusterState = "draining"

	// ClusterStateOffline indicates the cluster is offline
	ClusterStateOffline ClusterState = "offline"

	// ClusterStateDeregistered indicates the cluster has been deregistered
	ClusterStateDeregistered ClusterState = "deregistered"
)

// IsValidClusterState checks if the state is valid
func IsValidClusterState(s ClusterState) bool {
	switch s {
	case ClusterStatePending, ClusterStateActive, ClusterStateDraining, ClusterStateOffline, ClusterStateDeregistered:
		return true
	default:
		return false
	}
}

// HPCCluster represents a SLURM cluster registered on-chain
type HPCCluster struct {
	// ClusterID is the unique identifier for the cluster
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the provider that owns this cluster
	ProviderAddress string `json:"provider_address"`

	// Name is a human-readable name for the cluster
	Name string `json:"name"`

	// Description provides details about the cluster
	Description string `json:"description,omitempty"`

	// State is the current state of the cluster
	State ClusterState `json:"state"`

	// Partitions lists available SLURM partitions
	Partitions []Partition `json:"partitions"`

	// TotalNodes is the total number of nodes in the cluster
	TotalNodes int32 `json:"total_nodes"`

	// AvailableNodes is the number of nodes currently available
	AvailableNodes int32 `json:"available_nodes"`

	// Region is the geographic region of the cluster
	Region string `json:"region"`

	// ClusterMetadata contains additional cluster information
	ClusterMetadata ClusterMetadata `json:"cluster_metadata"`

	// SLURMVersion is the version of SLURM running
	SLURMVersion string `json:"slurm_version"`

	// KubernetesClusterID links to the K8s cluster if deployed via K8s
	KubernetesClusterID string `json:"kubernetes_cluster_id,omitempty"`

	// CreatedAt is when the cluster was registered
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the cluster was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// BlockHeight is when the cluster was registered
	BlockHeight int64 `json:"block_height"`
}

// Partition represents a SLURM partition/queue
type Partition struct {
	// Name is the partition name
	Name string `json:"name"`

	// Nodes is the number of nodes in this partition
	Nodes int32 `json:"nodes"`

	// MaxRuntime is the maximum runtime in seconds
	MaxRuntime int64 `json:"max_runtime"`

	// DefaultRuntime is the default runtime in seconds
	DefaultRuntime int64 `json:"default_runtime"`

	// MaxNodes is the maximum nodes per job
	MaxNodes int32 `json:"max_nodes"`

	// Features lists node features (e.g., "gpu", "highmem")
	Features []string `json:"features,omitempty"`

	// Priority is the partition priority (higher = more priority)
	Priority int32 `json:"priority"`

	// State is the partition state
	State string `json:"state"`
}

// ClusterMetadata contains additional cluster metadata
type ClusterMetadata struct {
	// TotalCPUCores across all nodes
	TotalCPUCores int64 `json:"total_cpu_cores"`

	// TotalMemoryGB across all nodes
	TotalMemoryGB int64 `json:"total_memory_gb"`

	// TotalGPUs across all nodes
	TotalGPUs int64 `json:"total_gpus,omitempty"`

	// GPUTypes lists available GPU types
	GPUTypes []string `json:"gpu_types,omitempty"`

	// InterconnectType is the interconnect type (e.g., "infiniband", "ethernet")
	InterconnectType string `json:"interconnect_type"`

	// StorageType is the storage system type
	StorageType string `json:"storage_type"`

	// TotalStorageGB is the total storage capacity
	TotalStorageGB int64 `json:"total_storage_gb"`
}

// Validate validates a cluster
func (c *HPCCluster) Validate() error {
	if c.ClusterID == "" {
		return ErrInvalidCluster.Wrap("cluster_id cannot be empty")
	}

	if len(c.ClusterID) > 64 {
		return ErrInvalidCluster.Wrap("cluster_id exceeds maximum length")
	}

	if _, err := sdk.AccAddressFromBech32(c.ProviderAddress); err != nil {
		return ErrInvalidCluster.Wrap("invalid provider address")
	}

	if c.Name == "" {
		return ErrInvalidCluster.Wrap("name cannot be empty")
	}

	if !IsValidClusterState(c.State) {
		return ErrInvalidCluster.Wrapf("invalid cluster state: %s", c.State)
	}

	if c.TotalNodes < 1 {
		return ErrInvalidCluster.Wrap("total_nodes must be at least 1")
	}

	if c.AvailableNodes > c.TotalNodes {
		return ErrInvalidCluster.Wrap("available_nodes cannot exceed total_nodes")
	}

	if c.Region == "" {
		return ErrInvalidCluster.Wrap("region cannot be empty")
	}

	return nil
}

// HPCOffering represents an HPC service offering
type HPCOffering struct {
	// OfferingID is the unique identifier for the offering
	OfferingID string `json:"offering_id"`

	// ClusterID is the cluster this offering is for
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the provider offering this service
	ProviderAddress string `json:"provider_address"`

	// Name is a human-readable name
	Name string `json:"name"`

	// Description provides details
	Description string `json:"description,omitempty"`

	// QueueOptions lists available queue/partition options
	QueueOptions []QueueOption `json:"queue_options"`

	// Pricing contains pricing information
	Pricing HPCPricing `json:"pricing"`

	// RequiredIdentityThreshold is the minimum identity score required
	RequiredIdentityThreshold int32 `json:"required_identity_threshold"`

	// MaxRuntimeSeconds is the maximum job runtime
	MaxRuntimeSeconds int64 `json:"max_runtime_seconds"`

	// PreconfiguredWorkloads lists available preconfigured workloads
	PreconfiguredWorkloads []PreconfiguredWorkload `json:"preconfigured_workloads,omitempty"`

	// SupportsCustomWorkloads indicates if custom workloads are allowed
	SupportsCustomWorkloads bool `json:"supports_custom_workloads"`

	// Active indicates if the offering is active
	Active bool `json:"active"`

	// CreatedAt is when the offering was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the offering was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// BlockHeight is when the offering was recorded
	BlockHeight int64 `json:"block_height"`
}

// QueueOption represents a queue/partition option in an offering
type QueueOption struct {
	// PartitionName is the SLURM partition name
	PartitionName string `json:"partition_name"`

	// DisplayName is a user-friendly name
	DisplayName string `json:"display_name"`

	// MaxNodes is the max nodes allowed
	MaxNodes int32 `json:"max_nodes"`

	// MaxRuntime is the max runtime in seconds
	MaxRuntime int64 `json:"max_runtime"`

	// Features lists required features
	Features []string `json:"features,omitempty"`

	// PriceMultiplier adjusts base pricing
	PriceMultiplier string `json:"price_multiplier"`
}

// HPCPricing contains HPC pricing information
type HPCPricing struct {
	// BaseNodeHourPrice is the base price per node-hour
	BaseNodeHourPrice string `json:"base_node_hour_price"`

	// CPUCoreHourPrice is the price per CPU core-hour
	CPUCoreHourPrice string `json:"cpu_core_hour_price"`

	// GPUHourPrice is the price per GPU-hour
	GPUHourPrice string `json:"gpu_hour_price,omitempty"`

	// MemoryGBHourPrice is the price per GB-hour of memory
	MemoryGBHourPrice string `json:"memory_gb_hour_price"`

	// StorageGBPrice is the price per GB of storage
	StorageGBPrice string `json:"storage_gb_price"`

	// NetworkGBPrice is the price per GB of network transfer
	NetworkGBPrice string `json:"network_gb_price"`

	// Currency is the pricing currency (token denom)
	Currency string `json:"currency"`

	// MinimumCharge is the minimum charge for any job
	MinimumCharge string `json:"minimum_charge"`
}

// PreconfiguredWorkload represents a pre-approved workload
type PreconfiguredWorkload struct {
	// WorkloadID is the unique identifier
	WorkloadID string `json:"workload_id"`

	// Name is the workload name
	Name string `json:"name"`

	// Description describes the workload
	Description string `json:"description"`

	// ContainerImage is the container image
	ContainerImage string `json:"container_image"`

	// DefaultCommand is the default command
	DefaultCommand string `json:"default_command,omitempty"`

	// RequiredResources are the required resources
	RequiredResources JobResources `json:"required_resources"`

	// Category is the workload category (e.g., "ml-training", "simulation")
	Category string `json:"category"`

	// Version is the workload version
	Version string `json:"version"`
}

// Validate validates an HPC offering
func (o *HPCOffering) Validate() error {
	if o.OfferingID == "" {
		return ErrInvalidOffering.Wrap("offering_id cannot be empty")
	}

	if len(o.OfferingID) > 64 {
		return ErrInvalidOffering.Wrap("offering_id exceeds maximum length")
	}

	if o.ClusterID == "" {
		return ErrInvalidOffering.Wrap("cluster_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(o.ProviderAddress); err != nil {
		return ErrInvalidOffering.Wrap("invalid provider address")
	}

	if o.Name == "" {
		return ErrInvalidOffering.Wrap("name cannot be empty")
	}

	if len(o.QueueOptions) == 0 {
		return ErrInvalidOffering.Wrap("at least one queue option is required")
	}

	if o.RequiredIdentityThreshold < 0 || o.RequiredIdentityThreshold > 100 {
		return ErrInvalidOffering.Wrap("required_identity_threshold must be between 0 and 100")
	}

	if o.MaxRuntimeSeconds < 60 {
		return ErrInvalidOffering.Wrap("max_runtime_seconds must be at least 60")
	}

	return nil
}
