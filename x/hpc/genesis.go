// Package hpc implements the HPC module for VirtEngine.
package hpc

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/hpc/keeper"
	"github.com/virtengine/virtengine/x/hpc/types"
)

// InitGenesis initializes the HPC module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data *types.GenesisState) {
	// Set module parameters
	if err := k.SetParams(ctx, data.Params); err != nil {
		panic(err)
	}

	// Import clusters
	for _, cluster := range data.Clusters {
		if err := k.SetCluster(ctx, cluster); err != nil {
			panic(err)
		}
	}

	// Import offerings
	for _, offering := range data.Offerings {
		if err := k.SetOffering(ctx, offering); err != nil {
			panic(err)
		}
	}

	// Import jobs
	for _, job := range data.Jobs {
		if err := k.SetJob(ctx, job); err != nil {
			panic(err)
		}
	}

	// Import job accountings
	for _, accounting := range data.JobAccountings {
		if err := k.SetJobAccounting(ctx, accounting); err != nil {
			panic(err)
		}
	}

	// Import node metadatas
	for _, node := range data.NodeMetadatas {
		if err := k.SetNodeMetadata(ctx, node); err != nil {
			panic(err)
		}
	}

	// Import scheduling decisions
	for _, decision := range data.SchedulingDecisions {
		if err := k.SetSchedulingDecision(ctx, decision); err != nil {
			panic(err)
		}
	}

	// Import scheduling metrics
	for _, metrics := range data.SchedulingMetrics {
		if err := k.SetSchedulingMetrics(ctx, metrics); err != nil {
			panic(err)
		}
	}

	// Import HPC rewards
	for _, reward := range data.HPCRewards {
		if err := k.SetHPCReward(ctx, reward); err != nil {
			panic(err)
		}
	}

	// Import disputes
	for _, dispute := range data.Disputes {
		if err := k.SetDispute(ctx, dispute); err != nil {
			panic(err)
		}
	}

	// Set sequences
	if data.ClusterSequence > 0 {
		k.SetNextClusterSequence(ctx, data.ClusterSequence)
	}
	if data.OfferingSequence > 0 {
		k.SetNextOfferingSequence(ctx, data.OfferingSequence)
	}
	if data.JobSequence > 0 {
		k.SetNextJobSequence(ctx, data.JobSequence)
	}
	if data.DecisionSequence > 0 {
		k.SetNextDecisionSequence(ctx, data.DecisionSequence)
	}
	if data.DisputeSequence > 0 {
		k.SetNextDisputeSequence(ctx, data.DisputeSequence)
	}
}

// ExportGenesis returns the HPC module's genesis state.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params := k.GetParams(ctx)

	// Export clusters
	var clusters []types.HPCCluster
	var maxClusterSeq uint64
	k.WithClusters(ctx, func(cluster types.HPCCluster) bool {
		clusters = append(clusters, cluster)
		if seq := extractSequenceFromID(cluster.ClusterID); seq > maxClusterSeq {
			maxClusterSeq = seq
		}
		return false
	})

	// Export offerings
	var offerings []types.HPCOffering
	var maxOfferingSeq uint64
	k.WithOfferings(ctx, func(offering types.HPCOffering) bool {
		offerings = append(offerings, offering)
		if seq := extractSequenceFromID(offering.OfferingID); seq > maxOfferingSeq {
			maxOfferingSeq = seq
		}
		return false
	})

	// Export jobs
	var jobs []types.HPCJob
	var maxJobSeq uint64
	k.WithJobs(ctx, func(job types.HPCJob) bool {
		jobs = append(jobs, job)
		if seq := extractSequenceFromID(job.JobID); seq > maxJobSeq {
			maxJobSeq = seq
		}
		return false
	})

	// Export job accountings
	var jobAccountings []types.JobAccounting
	k.WithJobs(ctx, func(job types.HPCJob) bool {
		if accounting, exists := k.GetJobAccounting(ctx, job.JobID); exists {
			jobAccountings = append(jobAccountings, accounting)
		}
		return false
	})

	// Export node metadatas
	var nodeMetadatas []types.NodeMetadata
	k.WithNodeMetadatas(ctx, func(node types.NodeMetadata) bool {
		nodeMetadatas = append(nodeMetadatas, node)
		return false
	})

	// Export scheduling decisions
	var schedulingDecisions []types.SchedulingDecision
	var maxDecisionSeq uint64
	k.WithSchedulingDecisions(ctx, func(decision types.SchedulingDecision) bool {
		schedulingDecisions = append(schedulingDecisions, decision)
		if seq := extractSequenceFromID(decision.DecisionID); seq > maxDecisionSeq {
			maxDecisionSeq = seq
		}
		return false
	})

	// Export scheduling metrics
	var schedulingMetrics []types.SchedulingMetrics
	k.WithSchedulingMetrics(ctx, func(metrics types.SchedulingMetrics) bool {
		schedulingMetrics = append(schedulingMetrics, metrics)
		return false
	})

	// Export HPC rewards
	var hpcRewards []types.HPCRewardRecord
	k.WithHPCRewards(ctx, func(reward types.HPCRewardRecord) bool {
		hpcRewards = append(hpcRewards, reward)
		return false
	})

	// Export disputes
	var disputes []types.HPCDispute
	var maxDisputeSeq uint64
	k.WithDisputes(ctx, func(dispute types.HPCDispute) bool {
		disputes = append(disputes, dispute)
		if seq := extractSequenceFromID(dispute.DisputeID); seq > maxDisputeSeq {
			maxDisputeSeq = seq
		}
		return false
	})

	return &types.GenesisState{
		Params:              params,
		Clusters:            clusters,
		Offerings:           offerings,
		Jobs:                jobs,
		JobAccountings:      jobAccountings,
		NodeMetadatas:       nodeMetadatas,
		SchedulingDecisions: schedulingDecisions,
		SchedulingMetrics:   schedulingMetrics,
		HPCRewards:          hpcRewards,
		Disputes:            disputes,
		ClusterSequence:     maxClusterSeq + 1,
		OfferingSequence:    maxOfferingSeq + 1,
		JobSequence:         maxJobSeq + 1,
		DecisionSequence:    maxDecisionSeq + 1,
		DisputeSequence:     maxDisputeSeq + 1,
	}
}

// ValidateGenesis validates the genesis state
func ValidateGenesis(data *types.GenesisState) error {
	return data.Validate()
}

// extractSequenceFromID extracts the sequence number from an ID like "hpc-cluster-123"
func extractSequenceFromID(id string) uint64 {
	parts := strings.Split(id, "-")
	if len(parts) < 1 {
		return 0
	}
	seq, err := strconv.ParseUint(parts[len(parts)-1], 10, 64)
	if err != nil {
		return 0
	}
	return seq
}
