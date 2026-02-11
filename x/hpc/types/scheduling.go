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

	// State is the current node state (pending, active, stale, draining, offline, deregistered)
	State NodeState `json:"state"`

	// HealthStatus is the health status (healthy, degraded, unhealthy, draining, offline)
	HealthStatus HealthStatus `json:"health_status"`

	// AgentPubkey is the base64-encoded Ed25519 public key of the node agent
	AgentPubkey string `json:"agent_pubkey,omitempty"`

	// HardwareFingerprint is SHA256 hash of hardware identifiers for attestation
	HardwareFingerprint string `json:"hardware_fingerprint,omitempty"`

	// AgentVersion is the version of the node agent software
	AgentVersion string `json:"agent_version,omitempty"`

	// LastSequenceNumber is the last heartbeat sequence number received
	LastSequenceNumber uint64 `json:"last_sequence_number"`

	// LastHeartbeat is the last heartbeat time
	LastHeartbeat time.Time `json:"last_heartbeat"`

	// LastHealthyAt is when the node was last in healthy state
	LastHealthyAt *time.Time `json:"last_healthy_at,omitempty"`

	// StateChangedAt is when the state last changed
	StateChangedAt time.Time `json:"state_changed_at"`

	// JoinedAt is when the node joined the cluster
	JoinedAt time.Time `json:"joined_at"`

	// UpdatedAt is when metadata was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// DeregisteredAt is when the node was deregistered (if applicable)
	DeregisteredAt *time.Time `json:"deregistered_at,omitempty"`

	// DeregistrationReason is why the node was deregistered
	DeregistrationReason string `json:"deregistration_reason,omitempty"`

	// Capacity contains current capacity snapshot from last heartbeat
	Capacity *NodeCapacity `json:"capacity,omitempty"`

	// Health contains current health snapshot from last heartbeat
	Health *NodeHealth `json:"health,omitempty"`

	// Hardware contains static hardware details
	Hardware *NodeHardware `json:"hardware,omitempty"`

	// Topology contains topology and fabric details
	Topology *NodeTopology `json:"topology,omitempty"`

	// Locality contains locality details beyond region/datacenter
	Locality *NodeLocality `json:"locality,omitempty"`

	// MissedHeartbeatCount is consecutive missed heartbeats
	MissedHeartbeatCount int32 `json:"missed_heartbeat_count"`

	// TotalHeartbeats is total heartbeats received since registration
	TotalHeartbeats uint64 `json:"total_heartbeats"`

	// TotalMissedHeartbeats is total missed heartbeats since registration
	TotalMissedHeartbeats uint64 `json:"total_missed_heartbeats"`

	// BlockHeight is when the metadata was recorded
	BlockHeight int64 `json:"block_height"`
}

// NodeStateAuditEntry records a node state transition for audit trail
type NodeStateAuditEntry struct {
	// NodeID is the node identifier
	NodeID string `json:"node_id"`

	// ClusterID is the cluster identifier
	ClusterID string `json:"cluster_id"`

	// PreviousState is the state before transition
	PreviousState NodeState `json:"previous_state"`

	// NewState is the state after transition
	NewState NodeState `json:"new_state"`

	// Reason is why the state changed
	Reason string `json:"reason"`

	// TriggeredBy indicates who/what triggered the change (system, provider, heartbeat)
	TriggeredBy string `json:"triggered_by"`

	// Timestamp is when the transition occurred
	Timestamp time.Time `json:"timestamp"`

	// BlockHeight is when the transition was recorded
	BlockHeight int64 `json:"block_height"`

	// Details contains additional context
	Details map[string]string `json:"details,omitempty"`
}

// ValidNodeStateTransitions defines allowed state transitions for nodes
var ValidNodeStateTransitions = map[NodeState][]NodeState{
	NodeStatePending:      {NodeStateActive, NodeStateOffline, NodeStateDeregistered},
	NodeStateActive:       {NodeStateStale, NodeStateDraining, NodeStateOffline, NodeStateDeregistered},
	NodeStateStale:        {NodeStateActive, NodeStateOffline, NodeStateDeregistered},
	NodeStateDraining:     {NodeStateDrained, NodeStateOffline, NodeStateDeregistered},
	NodeStateDrained:      {NodeStateActive, NodeStateOffline, NodeStateDeregistered},
	NodeStateOffline:      {NodeStateActive, NodeStatePending, NodeStateDeregistered},
	NodeStateDeregistered: {}, // Terminal state
}

// IsValidNodeStateTransition checks if a node state transition is valid
func IsValidNodeStateTransition(from, to NodeState) bool {
	allowed, ok := ValidNodeStateTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// IsTerminalNodeState checks if the node state is terminal
func IsTerminalNodeState(s NodeState) bool {
	return s == NodeStateDeregistered
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

	// PriorityScore is the priority score (fixed-point, 6 decimals)
	PriorityScore string `json:"priority_score,omitempty"`

	// FairShareScore is the fair-share score (fixed-point, 6 decimals)
	FairShareScore string `json:"fair_share_score,omitempty"`

	// AgeScore is the age-based priority score (fixed-point, 6 decimals)
	AgeScore string `json:"age_score,omitempty"`

	// JobSizeScore is the size-based priority score (fixed-point, 6 decimals)
	JobSizeScore string `json:"job_size_score,omitempty"`

	// PartitionScore is the partition priority score (fixed-point, 6 decimals)
	PartitionScore string `json:"partition_score,omitempty"`

	// PreemptionPlanned indicates if preemption is required to place the job
	PreemptionPlanned bool `json:"preemption_planned,omitempty"`

	// PreemptedJobIDs lists jobs planned for preemption
	PreemptedJobIDs []string `json:"preempted_job_ids,omitempty"`

	// BackfillUsed indicates if backfill scheduling was used
	BackfillUsed bool `json:"backfill_used,omitempty"`

	// BackfillWindowSeconds is the backfill window applied
	BackfillWindowSeconds int64 `json:"backfill_window_seconds,omitempty"`

	// QuotaBurstUsed indicates burst quota usage
	QuotaBurstUsed bool `json:"quota_burst_used,omitempty"`

	// QuotaReason explains quota usage or denial
	QuotaReason string `json:"quota_reason,omitempty"`

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

	// PriorityScore is the priority score (fixed-point, 6 decimals)
	PriorityScore string `json:"priority_score,omitempty"`

	// FairShareScore is the fair-share score (fixed-point, 6 decimals)
	FairShareScore string `json:"fair_share_score,omitempty"`

	// AgeScore is the age-based priority score (fixed-point, 6 decimals)
	AgeScore string `json:"age_score,omitempty"`

	// JobSizeScore is the size-based priority score (fixed-point, 6 decimals)
	JobSizeScore string `json:"job_size_score,omitempty"`

	// PartitionScore is the partition priority score (fixed-point, 6 decimals)
	PartitionScore string `json:"partition_score,omitempty"`

	// PreemptionPossible indicates if preemption could satisfy capacity
	PreemptionPossible bool `json:"preemption_possible,omitempty"`

	// QuotaBurstUsed indicates burst quota usage for this candidate
	QuotaBurstUsed bool `json:"quota_burst_used,omitempty"`

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
