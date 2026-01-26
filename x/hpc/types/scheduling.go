// Package types contains types for the HPC module.
//
// VE-503: Proximity-based mini-supercomputer clustering
package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NodeMetadata contains metadata about a compute node for proximity-based clustering
type NodeMetadata struct {
	// NodeID is the unique identifier for the node
	NodeID string `json:"node_id"`

	// ClusterID is the cluster this node belongs to
	ClusterID string `json:"cluster_id"`

	// ProviderAddress is the provider operating this node
	ProviderAddress string `json:"provider_address"`

	// Region is the geographic region
	Region string `json:"region"`

	// Datacenter is the datacenter identifier
	Datacenter string `json:"datacenter,omitempty"`

	// LatencyMeasurements contains latency measurements to other nodes
	LatencyMeasurements []LatencyMeasurement `json:"latency_measurements,omitempty"`

	// AvgLatencyMs is the average latency to other nodes in ms
	AvgLatencyMs int64 `json:"avg_latency_ms"`

	// NetworkBandwidthMbps is the network bandwidth in Mbps
	NetworkBandwidthMbps int64 `json:"network_bandwidth_mbps"`

	// Resources contains node resource capacity
	Resources NodeResources `json:"resources"`

	// Active indicates if the node is active
	Active bool `json:"active"`

	// LastHeartbeat is the last heartbeat time
	LastHeartbeat time.Time `json:"last_heartbeat"`

	// JoinedAt is when the node joined the cluster
	JoinedAt time.Time `json:"joined_at"`

	// UpdatedAt is when metadata was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// BlockHeight is when the metadata was recorded
	BlockHeight int64 `json:"block_height"`
}

// LatencyMeasurement contains latency measurement to another node
type LatencyMeasurement struct {
	// TargetNodeID is the target node
	TargetNodeID string `json:"target_node_id"`

	// LatencyMs is the measured latency in milliseconds
	LatencyMs int64 `json:"latency_ms"`

	// MeasuredAt is when the measurement was taken
	MeasuredAt time.Time `json:"measured_at"`
}

// NodeResources contains node resource capacity
type NodeResources struct {
	// CPUCores is the number of CPU cores
	CPUCores int32 `json:"cpu_cores"`

	// MemoryGB is the memory in GB
	MemoryGB int32 `json:"memory_gb"`

	// GPUs is the number of GPUs
	GPUs int32 `json:"gpus,omitempty"`

	// GPUType is the GPU type
	GPUType string `json:"gpu_type,omitempty"`

	// StorageGB is the storage in GB
	StorageGB int32 `json:"storage_gb"`
}

// Validate validates node metadata
func (n *NodeMetadata) Validate() error {
	if n.NodeID == "" {
		return ErrInvalidNodeMetadata.Wrap("node_id cannot be empty")
	}

	if len(n.NodeID) > 64 {
		return ErrInvalidNodeMetadata.Wrap("node_id exceeds maximum length")
	}

	if n.ClusterID == "" {
		return ErrInvalidNodeMetadata.Wrap("cluster_id cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(n.ProviderAddress); err != nil {
		return ErrInvalidNodeMetadata.Wrap("invalid provider address")
	}

	if n.Region == "" {
		return ErrInvalidNodeMetadata.Wrap("region cannot be empty")
	}

	return nil
}

// SchedulingDecision records the decision trail for job scheduling
type SchedulingDecision struct {
	// DecisionID is the unique identifier
	DecisionID string `json:"decision_id"`

	// JobID is the job being scheduled
	JobID string `json:"job_id"`

	// SelectedClusterID is the cluster selected for the job
	SelectedClusterID string `json:"selected_cluster_id"`

	// CandidateClusters are the clusters that were considered
	CandidateClusters []ClusterCandidate `json:"candidate_clusters"`

	// DecisionReason explains why the cluster was selected
	DecisionReason string `json:"decision_reason"`

	// IsFallback indicates if this was a fallback decision
	IsFallback bool `json:"is_fallback"`

	// FallbackReason explains why fallback was used
	FallbackReason string `json:"fallback_reason,omitempty"`

	// LatencyScore is the latency score for the selected cluster (fixed-point, 6 decimals)
	LatencyScore string `json:"latency_score"`

	// CapacityScore is the capacity score for the selected cluster (fixed-point, 6 decimals)
	CapacityScore string `json:"capacity_score"`

	// CombinedScore is the combined selection score (fixed-point, 6 decimals)
	CombinedScore string `json:"combined_score"`

	// CreatedAt is when the decision was made
	CreatedAt time.Time `json:"created_at"`

	// BlockHeight is when the decision was recorded
	BlockHeight int64 `json:"block_height"`
}

// ClusterCandidate represents a candidate cluster considered for scheduling
type ClusterCandidate struct {
	// ClusterID is the cluster ID
	ClusterID string `json:"cluster_id"`

	// Region is the cluster region
	Region string `json:"region"`

	// AvgLatencyMs is the average latency for this cluster
	AvgLatencyMs int64 `json:"avg_latency_ms"`

	// AvailableNodes is the number of available nodes
	AvailableNodes int32 `json:"available_nodes"`

	// LatencyScore is the latency score (fixed-point, 6 decimals)
	LatencyScore string `json:"latency_score"`

	// CapacityScore is the capacity score (fixed-point, 6 decimals)
	CapacityScore string `json:"capacity_score"`

	// CombinedScore is the combined score (fixed-point, 6 decimals)
	CombinedScore string `json:"combined_score"`

	// Eligible indicates if the cluster was eligible
	Eligible bool `json:"eligible"`

	// IneligibilityReason explains why the cluster wasn't eligible
	IneligibilityReason string `json:"ineligibility_reason,omitempty"`
}

// Validate validates a scheduling decision
func (s *SchedulingDecision) Validate() error {
	if s.DecisionID == "" {
		return ErrInvalidSchedulingDecision.Wrap("decision_id cannot be empty")
	}

	if s.JobID == "" {
		return ErrInvalidSchedulingDecision.Wrap("job_id cannot be empty")
	}

	if s.SelectedClusterID == "" {
		return ErrInvalidSchedulingDecision.Wrap("selected_cluster_id cannot be empty")
	}

	if s.DecisionReason == "" {
		return ErrInvalidSchedulingDecision.Wrap("decision_reason cannot be empty")
	}

	return nil
}
