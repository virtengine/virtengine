// Package types contains types for the HPC module.
package types

// GenesisState defines the genesis state for the HPC module
type GenesisState struct {
	// Params are the module parameters
	Params Params `json:"params"`

	// Clusters are the registered HPC clusters
	Clusters []HPCCluster `json:"clusters"`

	// Offerings are the HPC offerings
	Offerings []HPCOffering `json:"offerings"`

	// Jobs are the HPC jobs
	Jobs []HPCJob `json:"jobs"`

	// JobAccountings are the job accounting records
	JobAccountings []JobAccounting `json:"job_accountings"`

	// NodeMetadatas are the node metadata records
	NodeMetadatas []NodeMetadata `json:"node_metadatas"`

	// SchedulingDecisions are the scheduling decision records
	SchedulingDecisions []SchedulingDecision `json:"scheduling_decisions"`

	// HPCRewards are the HPC reward records
	HPCRewards []HPCRewardRecord `json:"hpc_rewards"`

	// Disputes are the dispute records
	Disputes []HPCDispute `json:"disputes"`

	// ClusterSequence is the next cluster sequence
	ClusterSequence uint64 `json:"cluster_sequence"`

	// OfferingSequence is the next offering sequence
	OfferingSequence uint64 `json:"offering_sequence"`

	// JobSequence is the next job sequence
	JobSequence uint64 `json:"job_sequence"`

	// DecisionSequence is the next decision sequence
	DecisionSequence uint64 `json:"decision_sequence"`

	// DisputeSequence is the next dispute sequence
	DisputeSequence uint64 `json:"dispute_sequence"`
}

// Params defines the parameters for the HPC module
type Params struct {
	// PlatformFeeRate is the platform fee rate for HPC jobs (fixed-point, 6 decimals)
	PlatformFeeRate string `json:"platform_fee_rate"`

	// ProviderRewardRate is the provider reward rate (fixed-point, 6 decimals)
	ProviderRewardRate string `json:"provider_reward_rate"`

	// NodeRewardRate is the node operator reward rate (fixed-point, 6 decimals)
	NodeRewardRate string `json:"node_reward_rate"`

	// MinJobDurationSeconds is the minimum job duration
	MinJobDurationSeconds int64 `json:"min_job_duration_seconds"`

	// MaxJobDurationSeconds is the maximum job duration
	MaxJobDurationSeconds int64 `json:"max_job_duration_seconds"`

	// DefaultIdentityThreshold is the default identity score threshold
	DefaultIdentityThreshold int32 `json:"default_identity_threshold"`

	// ClusterHeartbeatTimeout is the cluster heartbeat timeout in seconds
	ClusterHeartbeatTimeout int64 `json:"cluster_heartbeat_timeout"`

	// NodeHeartbeatTimeout is the node heartbeat timeout in seconds
	NodeHeartbeatTimeout int64 `json:"node_heartbeat_timeout"`

	// LatencyWeightFactor is the weight for latency in scheduling (fixed-point, 6 decimals)
	LatencyWeightFactor string `json:"latency_weight_factor"`

	// CapacityWeightFactor is the weight for capacity in scheduling (fixed-point, 6 decimals)
	CapacityWeightFactor string `json:"capacity_weight_factor"`

	// MaxLatencyMs is the maximum acceptable latency for cluster selection
	MaxLatencyMs int64 `json:"max_latency_ms"`

	// DisputeResolutionPeriod is the dispute resolution period in seconds
	DisputeResolutionPeriod int64 `json:"dispute_resolution_period"`

	// RewardFormulaVersion is the current reward formula version
	RewardFormulaVersion string `json:"reward_formula_version"`

	// EnableProximityClustering enables proximity-based clustering
	EnableProximityClustering bool `json:"enable_proximity_clustering"`
}

// DefaultParams returns the default parameters
func DefaultParams() Params {
	return Params{
		PlatformFeeRate:           "50000",  // 5% (50000/1000000)
		ProviderRewardRate:        "800000", // 80% (800000/1000000)
		NodeRewardRate:            "150000", // 15% (150000/1000000)
		MinJobDurationSeconds:     60,       // 1 minute
		MaxJobDurationSeconds:     604800,   // 7 days
		DefaultIdentityThreshold:  50,       // 50/100
		ClusterHeartbeatTimeout:   300,      // 5 minutes
		NodeHeartbeatTimeout:      120,      // 2 minutes
		LatencyWeightFactor:       "600000", // 0.6 weight
		CapacityWeightFactor:      "400000", // 0.4 weight
		MaxLatencyMs:              50,       // 50ms max
		DisputeResolutionPeriod:   604800,   // 7 days
		RewardFormulaVersion:      "v1.0.0",
		EnableProximityClustering: true,
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:              DefaultParams(),
		Clusters:            []HPCCluster{},
		Offerings:           []HPCOffering{},
		Jobs:                []HPCJob{},
		JobAccountings:      []JobAccounting{},
		NodeMetadatas:       []NodeMetadata{},
		SchedulingDecisions: []SchedulingDecision{},
		HPCRewards:          []HPCRewardRecord{},
		Disputes:            []HPCDispute{},
		ClusterSequence:     1,
		OfferingSequence:    1,
		JobSequence:         1,
		DecisionSequence:    1,
		DisputeSequence:     1,
	}
}

// Validate validates the genesis state
func (gs *GenesisState) Validate() error {
	// Validate clusters
	for i, cluster := range gs.Clusters {
		if err := cluster.Validate(); err != nil {
			return ErrInvalidCluster.Wrapf("invalid cluster at index %d: %s", i, err.Error())
		}
	}

	// Validate offerings
	for i, offering := range gs.Offerings {
		if err := offering.Validate(); err != nil {
			return ErrInvalidOffering.Wrapf("invalid offering at index %d: %s", i, err.Error())
		}
	}

	// Validate jobs
	for i, job := range gs.Jobs {
		if err := job.Validate(); err != nil {
			return ErrInvalidJob.Wrapf("invalid job at index %d: %s", i, err.Error())
		}
	}

	// Validate job accountings
	for i, accounting := range gs.JobAccountings {
		if err := accounting.Validate(); err != nil {
			return ErrInvalidJobAccounting.Wrapf("invalid job accounting at index %d: %s", i, err.Error())
		}
	}

	// Validate node metadatas
	for i, node := range gs.NodeMetadatas {
		if err := node.Validate(); err != nil {
			return ErrInvalidNodeMetadata.Wrapf("invalid node metadata at index %d: %s", i, err.Error())
		}
	}

	// Validate scheduling decisions
	for i, decision := range gs.SchedulingDecisions {
		if err := decision.Validate(); err != nil {
			return ErrInvalidSchedulingDecision.Wrapf("invalid scheduling decision at index %d: %s", i, err.Error())
		}
	}

	// Validate HPC rewards
	for i, reward := range gs.HPCRewards {
		if err := reward.Validate(); err != nil {
			return ErrInvalidReward.Wrapf("invalid HPC reward at index %d: %s", i, err.Error())
		}
	}

	// Validate disputes
	for i, dispute := range gs.Disputes {
		if err := dispute.Validate(); err != nil {
			return ErrInvalidDispute.Wrapf("invalid dispute at index %d: %s", i, err.Error())
		}
	}

	return nil
}
