// Package keeper implements the HPC module keeper.
//
// VE-503: Proximity-based mini-supercomputer clustering
package keeper

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// FixedPointScale is the scale for fixed-point arithmetic (6 decimals)
const FixedPointScale int64 = 1000000

// ============================================================================
// Scheduling Decision Management
// ============================================================================

// ScheduleJob selects the best cluster for a job using proximity-based heuristics
func (k Keeper) ScheduleJob(ctx sdk.Context, job *types.HPCJob) (*types.SchedulingDecision, error) {
	params := k.GetParams(ctx)

	offering, exists := k.GetOffering(ctx, job.OfferingID)
	if !exists {
		return nil, types.ErrOfferingNotFound
	}

	// Get all eligible clusters for this offering
	candidates := k.findEligibleClusters(ctx, job, &offering)
	if len(candidates) == 0 {
		return nil, types.ErrNoAvailableCluster
	}

	// Score and rank clusters
	k.scoreClusters(ctx, candidates, job, params)

	// Sort by combined score (descending)
	sort.Slice(candidates, func(i, j int) bool {
		scoreI := parseFixedPoint(candidates[i].CombinedScore)
		scoreJ := parseFixedPoint(candidates[j].CombinedScore)
		return scoreI > scoreJ
	})

	// Select best cluster
	var selectedCluster *types.ClusterCandidate
	var isFallback bool
	var fallbackReason string

	// Try to find a preferred (low-latency) cluster
	if params.EnableProximityClustering {
		for i := range candidates {
			if candidates[i].Eligible && candidates[i].AvgLatencyMs <= params.MaxLatencyMs {
				selectedCluster = &candidates[i]
				break
			}
		}
	}

	// Fallback to best available if no low-latency cluster found
	if selectedCluster == nil {
		for i := range candidates {
			if candidates[i].Eligible {
				selectedCluster = &candidates[i]
				isFallback = true
				fallbackReason = "No low-latency cluster available; selected best capacity match"
				break
			}
		}
	}

	if selectedCluster == nil {
		return nil, types.ErrNoAvailableCluster
	}

	// Create scheduling decision record
	seq := k.incrementSequence(ctx, types.SequenceKeyDecision)
	decision := &types.SchedulingDecision{
		DecisionID:        fmt.Sprintf("hpc-decision-%d", seq),
		JobID:             job.JobID,
		SelectedClusterID: selectedCluster.ClusterID,
		CandidateClusters: candidates,
		DecisionReason:    formatDecisionReason(selectedCluster, isFallback),
		IsFallback:        isFallback,
		FallbackReason:    fallbackReason,
		LatencyScore:      selectedCluster.LatencyScore,
		CapacityScore:     selectedCluster.CapacityScore,
		CombinedScore:     selectedCluster.CombinedScore,
		CreatedAt:         ctx.BlockTime(),
		BlockHeight:       ctx.BlockHeight(),
	}

	if err := k.SetSchedulingDecision(ctx, *decision); err != nil {
		return nil, err
	}

	// Update job with scheduling decision
	job.SchedulingDecisionID = decision.DecisionID
	job.ClusterID = selectedCluster.ClusterID

	return decision, nil
}

// findEligibleClusters finds clusters eligible for the job
//
//nolint:unparam // offering kept for future offering-based cluster filtering
func (k Keeper) findEligibleClusters(ctx sdk.Context, job *types.HPCJob, _ *types.HPCOffering) []types.ClusterCandidate {
	var candidates []types.ClusterCandidate

	k.WithClusters(ctx, func(cluster types.HPCCluster) bool {
		candidate := types.ClusterCandidate{
			ClusterID:      cluster.ClusterID,
			Region:         cluster.Region,
			AvailableNodes: cluster.AvailableNodes,
			Eligible:       true,
		}

		// Check cluster state
		if cluster.State != types.ClusterStateActive {
			candidate.Eligible = false
			candidate.IneligibilityReason = "Cluster is not active"
		}

		// Check capacity
		if cluster.AvailableNodes < job.Resources.Nodes {
			candidate.Eligible = false
			candidate.IneligibilityReason = fmt.Sprintf("Insufficient nodes: need %d, have %d", job.Resources.Nodes, cluster.AvailableNodes)
		}

		// Calculate average latency for this cluster
		nodes := k.GetNodesByCluster(ctx, cluster.ClusterID)
		if len(nodes) > 0 {
			var totalLatency int64
			activeNodes := 0
			for _, node := range nodes {
				if node.Active {
					totalLatency += node.AvgLatencyMs
					activeNodes++
				}
			}
			if activeNodes > 0 {
				candidate.AvgLatencyMs = totalLatency / int64(activeNodes)
			}
		}

		candidates = append(candidates, candidate)
		return false
	})

	return candidates
}

// scoreClusters calculates scores for cluster candidates
//
//nolint:unparam // ctx kept for future state-based scoring adjustments
func (k Keeper) scoreClusters(_ sdk.Context, candidates []types.ClusterCandidate, job *types.HPCJob, params types.Params) {
	latencyWeight := parseFixedPoint(params.LatencyWeightFactor)
	capacityWeight := parseFixedPoint(params.CapacityWeightFactor)

	// Find max values for normalization
	var maxLatency int64 = 1
	var maxCapacity int32 = 1
	for _, c := range candidates {
		if c.AvgLatencyMs > maxLatency {
			maxLatency = c.AvgLatencyMs
		}
		if c.AvailableNodes > maxCapacity {
			maxCapacity = c.AvailableNodes
		}
	}

	// Calculate scores using fixed-point arithmetic
	for i := range candidates {
		// Latency score: lower latency = higher score
		// Score = (maxLatency - latency) / maxLatency * 1000000
		latencyScore := (maxLatency - candidates[i].AvgLatencyMs) * FixedPointScale / maxLatency

		// Capacity score: higher capacity = higher score
		// Score = availableNodes / maxCapacity * 1000000
		capacityScore := int64(candidates[i].AvailableNodes) * FixedPointScale / int64(maxCapacity)

		// Combined score = latencyWeight * latencyScore + capacityWeight * capacityScore
		// All values are in fixed-point, so divide by scale once at the end
		combinedScore := (latencyWeight*latencyScore + capacityWeight*capacityScore) / FixedPointScale

		candidates[i].LatencyScore = strconv.FormatInt(latencyScore, 10)
		candidates[i].CapacityScore = strconv.FormatInt(capacityScore, 10)
		candidates[i].CombinedScore = strconv.FormatInt(combinedScore, 10)
	}
}

// GetSchedulingDecision retrieves a scheduling decision by ID
func (k Keeper) GetSchedulingDecision(ctx sdk.Context, decisionID string) (types.SchedulingDecision, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetSchedulingDecisionKey(decisionID))
	if bz == nil {
		return types.SchedulingDecision{}, false
	}

	var decision types.SchedulingDecision
	if err := json.Unmarshal(bz, &decision); err != nil {
		return types.SchedulingDecision{}, false
	}
	return decision, true
}

// SetSchedulingDecision stores a scheduling decision
func (k Keeper) SetSchedulingDecision(ctx sdk.Context, decision types.SchedulingDecision) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(decision)
	if err != nil {
		return err
	}
	store.Set(types.GetSchedulingDecisionKey(decision.DecisionID), bz)
	return nil
}

// WithSchedulingDecisions iterates over all scheduling decisions
func (k Keeper) WithSchedulingDecisions(ctx sdk.Context, fn func(types.SchedulingDecision) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.SchedulingDecisionPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var decision types.SchedulingDecision
		if err := json.Unmarshal(iter.Value(), &decision); err != nil {
			continue
		}
		if fn(decision) {
			break
		}
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// parseFixedPoint parses a fixed-point string to int64
func parseFixedPoint(s string) int64 {
	if s == "" {
		return 0
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// formatFixedPoint formats an int64 to fixed-point string
func formatFixedPoint(val int64) string {
	return strconv.FormatInt(val, 10)
}

// formatDecisionReason creates a human-readable decision reason
func formatDecisionReason(selected *types.ClusterCandidate, isFallback bool) string {
	if isFallback {
		return fmt.Sprintf("Fallback selection: cluster %s with capacity score %s",
			selected.ClusterID, selected.CapacityScore)
	}
	return fmt.Sprintf("Optimal selection: cluster %s with latency %dms and combined score %s",
		selected.ClusterID, selected.AvgLatencyMs, selected.CombinedScore)
}

// ============================================================================
// Exported Helper Functions for Testing
// ============================================================================

// SortCandidatesByScore sorts cluster candidates by score in descending order
func SortCandidatesByScore(candidates []types.ClusterCandidate) []types.ClusterCandidate {
	result := make([]types.ClusterCandidate, len(candidates))
	copy(result, candidates)
	sort.Slice(result, func(i, j int) bool {
		scoreI := parseFixedPoint(result[i].CombinedScore)
		scoreJ := parseFixedPoint(result[j].CombinedScore)
		return scoreI > scoreJ
	})
	return result
}

// CalculateAverageLatency calculates average latency from measurements
func CalculateAverageLatency(measurements []types.LatencyMeasurement) int64 {
	if len(measurements) == 0 {
		return 0
	}

	var total int64
	for _, m := range measurements {
		total += m.LatencyMs
	}

	return total / int64(len(measurements))
}

// FilterEligibleClusters filters clusters that meet job resource requirements
func FilterEligibleClusters(clusters []types.HPCCluster, resources types.JobResources) []types.HPCCluster {
	eligible := make([]types.HPCCluster, 0, len(clusters))

	for _, cluster := range clusters {
		// Skip non-active clusters
		if cluster.State != types.ClusterStateActive {
			continue
		}

		// Check node capacity (use available nodes)
		if cluster.AvailableNodes < resources.Nodes {
			continue
		}

		eligible = append(eligible, cluster)
	}

	return eligible
}
