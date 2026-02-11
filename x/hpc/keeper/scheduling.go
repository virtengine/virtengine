// Package keeper implements the HPC module keeper.
//
// VE-503: Proximity-based mini-supercomputer clustering
package keeper

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// FixedPointScale is the scale for fixed-point arithmetic (6 decimals)
const FixedPointScale int64 = 1000000

const (
	defaultQueueName = "default"

	priorityWeightAge       int64 = 400000
	priorityWeightJobSize   int64 = 300000
	priorityWeightPartition int64 = 300000

	priorityMaxAgeSeconds   int64 = 3600
	backfillWindowSeconds   int64 = 1800
	preemptionEnabled              = true
	backfillEnabled                = true

	quotaMaxRunningJobsPerTenant int32 = 12
	quotaMaxQueuedJobsPerTenant  int32 = 24
	quotaBurstRunningJobs        int32 = 6
	quotaBurstQueuedJobs         int32 = 12

	quotaMinRunningNodes      int32 = 2
	quotaNodeShareDivisor     int32 = 4
	quotaBurstNodeShareDivisor int32 = 2

	quotaMinRunningCPUCores   int32 = 32
	quotaMinRunningMemoryGB   int32 = 64
	quotaMinRunningGPUs       int32 = 1
)

type tenantUsage struct {
	RunningJobs     int32
	QueuedJobs      int32
	RunningNodes    int32
	RunningCPUCores int32
	RunningMemoryGB int32
	RunningGPUs     int32
}

type tenantQuota struct {
	MaxRunningJobs     int32
	MaxQueuedJobs      int32
	MaxRunningNodes    int32
	MaxRunningCPUCores int32
	MaxRunningMemoryGB int32
	MaxRunningGPUs     int32

	BurstRunningJobs     int32
	BurstQueuedJobs      int32
	BurstRunningNodes    int32
	BurstRunningCPUCores int32
	BurstRunningMemoryGB int32
	BurstRunningGPUs     int32
}

type candidateMeta struct {
	cluster            types.HPCCluster
	queueName          string
	partitionPriority  int32
	usage              tenantUsage
	quota              tenantQuota
	quotaBurstUsed     bool
	quotaReason        string
	preemptedJobIDs    []string
	preemptionPossible bool
	backfillUsed       bool
	backfillWindow     int64
	priorityScore      int64
	fairShareScore     int64
	ageScore           int64
	jobSizeScore       int64
	partitionScore     int64
}

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

	queueName := resolveQueueName(job, &offering)
	job.QueueName = queueName

	// Get all eligible clusters for this offering
	candidates, meta := k.findEligibleClusters(ctx, job, &offering)
	if len(candidates) == 0 {
		return nil, types.ErrNoAvailableCluster
	}

	// Score and rank clusters
	k.scoreClusters(ctx, candidates, job, params, meta)

	// Sort by combined score (descending)
	sort.Slice(candidates, func(i, j int) bool {
		scoreI := parseFixedPoint(candidates[i].CombinedScore)
		scoreJ := parseFixedPoint(candidates[j].CombinedScore)
		if scoreI == scoreJ {
			return candidates[i].ClusterID < candidates[j].ClusterID
		}
		return scoreI > scoreJ
	})

	// Select best cluster
	var selectedCluster *types.ClusterCandidate
	var selectedMeta *candidateMeta
	var isFallback bool
	var fallbackReason string

	// Try to find a preferred (low-latency) cluster
	if params.EnableProximityClustering {
		for i := range candidates {
			if candidates[i].Eligible && candidates[i].AvgLatencyMs <= params.MaxLatencyMs {
				selectedCluster = &candidates[i]
				selectedMeta = meta[candidates[i].ClusterID]
				break
			}
		}
	}

	// Fallback to best available if no low-latency cluster found
	if selectedCluster == nil {
		for i := range candidates {
			if candidates[i].Eligible {
				selectedCluster = &candidates[i]
				selectedMeta = meta[candidates[i].ClusterID]
				isFallback = true
				fallbackReason = "No low-latency cluster available; selected best capacity match"
				break
			}
		}
	}

	if selectedCluster == nil {
		if schedulingBlockedByQuota(candidates) {
			k.recordQuotaDenied(ctx, queueName)
			return nil, types.ErrTenantQuotaExceeded
		}
		return nil, types.ErrNoAvailableCluster
	}

	if selectedMeta == nil {
		selectedMeta = &candidateMeta{}
	}

	// Create scheduling decision record
	seq := k.incrementSequence(ctx, types.SequenceKeyDecision)
	decision := &types.SchedulingDecision{
		DecisionID:        fmt.Sprintf("hpc-decision-%d", seq),
		JobID:             job.JobID,
		SelectedClusterID: selectedCluster.ClusterID,
		CandidateClusters: candidates,
		IsFallback:        isFallback,
		FallbackReason:    fallbackReason,
		LatencyScore:      selectedCluster.LatencyScore,
		CapacityScore:     selectedCluster.CapacityScore,
		CombinedScore:     selectedCluster.CombinedScore,
		PriorityScore:     selectedCluster.PriorityScore,
		FairShareScore:    selectedCluster.FairShareScore,
		AgeScore:          selectedCluster.AgeScore,
		JobSizeScore:      selectedCluster.JobSizeScore,
		PartitionScore:    selectedCluster.PartitionScore,
		PreemptionPlanned: selectedMeta.preemptionPossible && len(selectedMeta.preemptedJobIDs) > 0,
		PreemptedJobIDs:   selectedMeta.preemptedJobIDs,
		BackfillUsed:      selectedMeta.backfillUsed,
		BackfillWindowSeconds: selectedMeta.backfillWindow,
		QuotaBurstUsed:    selectedCluster.QuotaBurstUsed,
		QuotaReason:       selectedMeta.quotaReason,
		CreatedAt:         ctx.BlockTime(),
		BlockHeight:       ctx.BlockHeight(),
	}

	decision.DecisionReason = formatDecisionReason(selectedCluster, isFallback, decision)

	if err := k.SetSchedulingDecision(ctx, *decision); err != nil {
		return nil, err
	}

	// Update job with scheduling decision
	job.SchedulingDecisionID = decision.DecisionID
	job.ClusterID = selectedCluster.ClusterID

	// Emit scheduling events and metrics
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"hpc_scheduling_decision",
			sdk.NewAttribute("job_id", job.JobID),
			sdk.NewAttribute("cluster_id", decision.SelectedClusterID),
			sdk.NewAttribute("combined_score", decision.CombinedScore),
			sdk.NewAttribute("priority_score", decision.PriorityScore),
			sdk.NewAttribute("fair_share_score", decision.FairShareScore),
			sdk.NewAttribute("preemption_planned", fmt.Sprintf("%t", decision.PreemptionPlanned)),
			sdk.NewAttribute("backfill_used", fmt.Sprintf("%t", decision.BackfillUsed)),
			sdk.NewAttribute("quota_burst_used", fmt.Sprintf("%t", decision.QuotaBurstUsed)),
		),
	)

	if decision.PreemptionPlanned && len(decision.PreemptedJobIDs) > 0 {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"hpc_preemption_plan",
				sdk.NewAttribute("job_id", job.JobID),
				sdk.NewAttribute("cluster_id", decision.SelectedClusterID),
				sdk.NewAttribute("preempted_job_ids", strings.Join(decision.PreemptedJobIDs, ",")),
			),
		)
	}

	if decision.BackfillUsed {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				"hpc_backfill_scheduled",
				sdk.NewAttribute("job_id", job.JobID),
				sdk.NewAttribute("cluster_id", decision.SelectedClusterID),
				sdk.NewAttribute("backfill_window_seconds", fmt.Sprintf("%d", decision.BackfillWindowSeconds)),
			),
		)
	}

	k.updateSchedulingMetrics(ctx, job, decision)

	return decision, nil
}

// findEligibleClusters finds clusters eligible for the job
//
//nolint:unparam // offering kept for future offering-based cluster filtering
func (k Keeper) findEligibleClusters(ctx sdk.Context, job *types.HPCJob, offering *types.HPCOffering) ([]types.ClusterCandidate, map[string]*candidateMeta) {
	candidates := make([]types.ClusterCandidate, 0)
	meta := make(map[string]*candidateMeta)

	queueName := resolveQueueName(job, offering)

	k.WithClusters(ctx, func(cluster types.HPCCluster) bool {
		candidate := types.ClusterCandidate{
			ClusterID:      cluster.ClusterID,
			Region:         cluster.Region,
			AvailableNodes: cluster.AvailableNodes,
			Eligible:       true,
		}

		info := &candidateMeta{
			cluster:           cluster,
			queueName:         queueName,
			partitionPriority: resolvePartitionPriority(cluster, queueName),
		}

		// Check cluster state
		if cluster.State != types.ClusterStateActive {
			candidate.Eligible = false
			candidate.IneligibilityReason = "Cluster is not active"
		}

		// Check queue limits when present
		if queueLimitReason := validateQueueOptionLimits(offering, queueName, job); queueLimitReason != "" {
			candidate.Eligible = false
			candidate.IneligibilityReason = queueLimitReason
		}

		// Quota enforcement
		info.usage = k.tenantUsageForCluster(ctx, job.CustomerAddress, cluster.ClusterID)
		info.quota = defaultTenantQuotaForCluster(cluster)
		allowed, burstUsed, quotaReason := checkTenantQuota(info.usage, info.quota, job)
		info.quotaBurstUsed = burstUsed
		info.quotaReason = quotaReason
		candidate.QuotaBurstUsed = burstUsed
		if !allowed {
			candidate.Eligible = false
			candidate.IneligibilityReason = fmt.Sprintf("Tenant quota exceeded: %s", quotaReason)
		}

		// Capacity / preemption checks
		if candidate.Eligible && cluster.AvailableNodes < job.Resources.Nodes {
			if preemptionEnabled {
				required := job.Resources.Nodes - cluster.AvailableNodes
				preempted, freed := k.preemptionPlan(ctx, cluster, job, info.partitionPriority)
				if freed >= required && len(preempted) > 0 {
					info.preemptedJobIDs = preempted
					info.preemptionPossible = true
					candidate.PreemptionPossible = true
				} else {
					candidate.Eligible = false
					candidate.IneligibilityReason = fmt.Sprintf("Insufficient nodes: need %d, have %d", job.Resources.Nodes, cluster.AvailableNodes)
				}
			} else {
				candidate.Eligible = false
				candidate.IneligibilityReason = fmt.Sprintf("Insufficient nodes: need %d, have %d", job.Resources.Nodes, cluster.AvailableNodes)
			}
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

		// Backfill evaluation
		if candidate.Eligible {
			backfillUsed, backfillWindow, backfillReason := k.evaluateBackfill(ctx, cluster, job, info.partitionPriority)
			if backfillReason != "" {
				candidate.Eligible = false
				candidate.IneligibilityReason = backfillReason
			} else if backfillUsed {
				info.backfillUsed = true
				info.backfillWindow = backfillWindow
			}
		}

		candidates = append(candidates, candidate)
		meta[cluster.ClusterID] = info
		return false
	})

	return candidates, meta
}

// scoreClusters calculates scores for cluster candidates
//
//nolint:unparam // ctx kept for future state-based scoring adjustments
func (k Keeper) scoreClusters(ctx sdk.Context, candidates []types.ClusterCandidate, job *types.HPCJob, params types.Params, meta map[string]*candidateMeta) {
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

		baseScore := (latencyWeight*latencyScore + capacityWeight*capacityScore) / FixedPointScale

		metaInfo := meta[candidates[i].ClusterID]
		if metaInfo != nil {
			metaInfo.priorityScore, metaInfo.fairShareScore, metaInfo.ageScore, metaInfo.jobSizeScore, metaInfo.partitionScore =
				computePriorityAndFairShare(ctx.BlockTime(), job, metaInfo)
		}

		priorityScore := int64(0)
		fairShareScore := int64(0)
		ageScore := int64(0)
		jobSizeScore := int64(0)
		partitionScore := int64(0)
		if metaInfo != nil {
			priorityScore = metaInfo.priorityScore
			fairShareScore = metaInfo.fairShareScore
			ageScore = metaInfo.ageScore
			jobSizeScore = metaInfo.jobSizeScore
			partitionScore = metaInfo.partitionScore
		}

		combinedScore := (baseScore + priorityScore + fairShareScore) / 3

		candidates[i].LatencyScore = strconv.FormatInt(latencyScore, 10)
		candidates[i].CapacityScore = strconv.FormatInt(capacityScore, 10)
		candidates[i].CombinedScore = strconv.FormatInt(combinedScore, 10)
		candidates[i].PriorityScore = strconv.FormatInt(priorityScore, 10)
		candidates[i].FairShareScore = strconv.FormatInt(fairShareScore, 10)
		candidates[i].AgeScore = strconv.FormatInt(ageScore, 10)
		candidates[i].JobSizeScore = strconv.FormatInt(jobSizeScore, 10)
		candidates[i].PartitionScore = strconv.FormatInt(partitionScore, 10)
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

func resolveQueueName(job *types.HPCJob, offering *types.HPCOffering) string {
	if job.QueueName != "" {
		return job.QueueName
	}
	if offering != nil {
		for _, option := range offering.QueueOptions {
			if option.PartitionName != "" {
				return option.PartitionName
			}
		}
	}
	return defaultQueueName
}

func validateQueueOptionLimits(offering *types.HPCOffering, queueName string, job *types.HPCJob) string {
	if offering == nil || len(offering.QueueOptions) == 0 {
		return ""
	}

	var option *types.QueueOption
	for i := range offering.QueueOptions {
		if offering.QueueOptions[i].PartitionName == queueName || offering.QueueOptions[i].DisplayName == queueName {
			option = &offering.QueueOptions[i]
			break
		}
	}

	if option == nil {
		return ""
	}

	if option.MaxNodes > 0 && job.Resources.Nodes > option.MaxNodes {
		return fmt.Sprintf("Queue limit exceeded: max_nodes=%d", option.MaxNodes)
	}

	if option.MaxRuntime > 0 && job.MaxRuntimeSeconds > option.MaxRuntime {
		return fmt.Sprintf("Queue limit exceeded: max_runtime=%d", option.MaxRuntime)
	}

	return ""
}

func resolvePartitionPriority(cluster types.HPCCluster, queueName string) int32 {
	for _, partition := range cluster.Partitions {
		if partition.Name == queueName {
			return partition.Priority
		}
	}
	return 0
}

func (k Keeper) tenantUsageForCluster(ctx sdk.Context, tenant string, clusterID string) tenantUsage {
	var usage tenantUsage
	k.WithJobs(ctx, func(job types.HPCJob) bool {
		if job.ClusterID != clusterID || job.CustomerAddress != tenant {
			return false
		}
		if types.IsTerminalJobState(job.State) {
			return false
		}

		switch job.State {
		case types.JobStateRunning:
			usage.RunningJobs++
			usage.RunningNodes = saturatingAddInt32(usage.RunningNodes, job.Resources.Nodes)
			usage.RunningCPUCores = saturatingAddInt32(usage.RunningCPUCores, safeInt32Mul(job.Resources.Nodes, job.Resources.CPUCoresPerNode))
			usage.RunningMemoryGB = saturatingAddInt32(usage.RunningMemoryGB, safeInt32Mul(job.Resources.Nodes, job.Resources.MemoryGBPerNode))
			usage.RunningGPUs = saturatingAddInt32(usage.RunningGPUs, safeInt32Mul(job.Resources.Nodes, job.Resources.GPUsPerNode))
		case types.JobStatePending, types.JobStateQueued:
			usage.QueuedJobs++
		}

		return false
	})

	return usage
}

func defaultTenantQuotaForCluster(cluster types.HPCCluster) tenantQuota {
	maxRunningNodes := int32(0)
	if cluster.TotalNodes > 0 {
		maxRunningNodes = cluster.TotalNodes / quotaNodeShareDivisor
	}
	if maxRunningNodes < quotaMinRunningNodes {
		maxRunningNodes = quotaMinRunningNodes
	}

	var burstRunningNodes int32
	if cluster.TotalNodes > 0 {
		burstLimit := cluster.TotalNodes / quotaBurstNodeShareDivisor
		if burstLimit < maxRunningNodes {
			burstLimit = maxRunningNodes
		}
		if burstLimit > cluster.TotalNodes {
			burstLimit = cluster.TotalNodes
		}
		burstRunningNodes = burstLimit - maxRunningNodes
	}

	maxRunningCPUCores, burstRunningCPUCores := deriveQuotaFromTotal(cluster.ClusterMetadata.TotalCPUCores, quotaMinRunningCPUCores)
	maxRunningMemoryGB, burstRunningMemoryGB := deriveQuotaFromTotal(cluster.ClusterMetadata.TotalMemoryGB, quotaMinRunningMemoryGB)
	maxRunningGPUs, burstRunningGPUs := deriveQuotaFromTotal(cluster.ClusterMetadata.TotalGPUs, quotaMinRunningGPUs)

	return tenantQuota{
		MaxRunningJobs:     quotaMaxRunningJobsPerTenant,
		MaxQueuedJobs:      quotaMaxQueuedJobsPerTenant,
		MaxRunningNodes:    maxRunningNodes,
		MaxRunningCPUCores: maxRunningCPUCores,
		MaxRunningMemoryGB: maxRunningMemoryGB,
		MaxRunningGPUs:     maxRunningGPUs,
		BurstRunningJobs:   quotaBurstRunningJobs,
		BurstQueuedJobs:    quotaBurstQueuedJobs,
		BurstRunningNodes:  burstRunningNodes,
		BurstRunningCPUCores: burstRunningCPUCores,
		BurstRunningMemoryGB: burstRunningMemoryGB,
		BurstRunningGPUs:     burstRunningGPUs,
	}
}

func deriveQuotaFromTotal(total int64, min int32) (int32, int32) {
	if total <= 0 {
		return 0, 0
	}

	base := total / 4
	if base < int64(min) {
		base = int64(min)
	}
	if base > total {
		base = total
	}

	burstLimit := total / 2
	if burstLimit < base {
		burstLimit = base
	}
	if burstLimit > total {
		burstLimit = total
	}

	base32 := clampInt32(base)
	burst32 := clampInt32(burstLimit - base)
	return base32, burst32
}

func checkTenantQuota(usage tenantUsage, quota tenantQuota, job *types.HPCJob) (bool, bool, string) {
	burstUsed := false

	if quota.MaxRunningJobs > 0 {
		runningAfter := usage.RunningJobs + 1
		allowed, burst := withinQuota(runningAfter, quota.MaxRunningJobs, quota.BurstRunningJobs)
		if !allowed {
			return false, false, fmt.Sprintf("running_jobs=%d max=%d", runningAfter, quota.MaxRunningJobs)
		}
		burstUsed = burstUsed || burst
	}

	if quota.MaxQueuedJobs > 0 {
		queuedAfter := usage.QueuedJobs + 1
		allowed, burst := withinQuota(queuedAfter, quota.MaxQueuedJobs, quota.BurstQueuedJobs)
		if !allowed {
			return false, false, fmt.Sprintf("queued_jobs=%d max=%d", queuedAfter, quota.MaxQueuedJobs)
		}
		burstUsed = burstUsed || burst
	}

	if quota.MaxRunningNodes > 0 {
		nodesAfter := usage.RunningNodes + job.Resources.Nodes
		allowed, burst := withinQuota(nodesAfter, quota.MaxRunningNodes, quota.BurstRunningNodes)
		if !allowed {
			return false, false, fmt.Sprintf("running_nodes=%d max=%d", nodesAfter, quota.MaxRunningNodes)
		}
		burstUsed = burstUsed || burst
	}

	if quota.MaxRunningCPUCores > 0 {
		coresAfter := usage.RunningCPUCores + safeInt32Mul(job.Resources.Nodes, job.Resources.CPUCoresPerNode)
		allowed, burst := withinQuota(coresAfter, quota.MaxRunningCPUCores, quota.BurstRunningCPUCores)
		if !allowed {
			return false, false, fmt.Sprintf("cpu_cores=%d max=%d", coresAfter, quota.MaxRunningCPUCores)
		}
		burstUsed = burstUsed || burst
	}

	if quota.MaxRunningMemoryGB > 0 {
		memAfter := usage.RunningMemoryGB + safeInt32Mul(job.Resources.Nodes, job.Resources.MemoryGBPerNode)
		allowed, burst := withinQuota(memAfter, quota.MaxRunningMemoryGB, quota.BurstRunningMemoryGB)
		if !allowed {
			return false, false, fmt.Sprintf("memory_gb=%d max=%d", memAfter, quota.MaxRunningMemoryGB)
		}
		burstUsed = burstUsed || burst
	}

	if quota.MaxRunningGPUs > 0 {
		gpuAfter := usage.RunningGPUs + safeInt32Mul(job.Resources.Nodes, job.Resources.GPUsPerNode)
		allowed, burst := withinQuota(gpuAfter, quota.MaxRunningGPUs, quota.BurstRunningGPUs)
		if !allowed {
			return false, false, fmt.Sprintf("gpus=%d max=%d", gpuAfter, quota.MaxRunningGPUs)
		}
		burstUsed = burstUsed || burst
	}

	return true, burstUsed, ""
}

func withinQuota(value, max, burst int32) (bool, bool) {
	limit := max + burst
	if value > limit {
		return false, false
	}
	if value > max {
		return true, true
	}
	return true, false
}

func computePriorityAndFairShare(now time.Time, job *types.HPCJob, meta *candidateMeta) (int64, int64, int64, int64, int64) {
	ageScore := calculateAgeScore(now, job.CreatedAt)
	jobSizeScore := calculateJobSizeScore(job, meta.cluster)
	partitionScore := calculatePartitionScore(meta.partitionPriority)
	priorityScore := (priorityWeightAge*ageScore + priorityWeightJobSize*jobSizeScore + priorityWeightPartition*partitionScore) / FixedPointScale
	fairShareScore := calculateFairShareScore(meta.usage, meta.quota)
	return priorityScore, fairShareScore, ageScore, jobSizeScore, partitionScore
}

func calculateAgeScore(now time.Time, createdAt time.Time) int64 {
	if createdAt.IsZero() {
		return 0
	}
	ageSeconds := int64(now.Sub(createdAt).Seconds())
	if ageSeconds <= 0 {
		return 0
	}
	if ageSeconds >= priorityMaxAgeSeconds {
		return FixedPointScale
	}
	return ageSeconds * FixedPointScale / priorityMaxAgeSeconds
}

func calculateJobSizeScore(job *types.HPCJob, cluster types.HPCCluster) int64 {
	if cluster.TotalNodes <= 0 || job.Resources.Nodes <= 0 {
		return 0
	}
	total := int64(cluster.TotalNodes)
	nodes := int64(job.Resources.Nodes)
	if nodes >= total {
		return 0
	}
	return (total - nodes) * FixedPointScale / total
}

func calculatePartitionScore(priority int32) int64 {
	if priority <= 0 {
		return 0
	}
	if priority > 1000 {
		priority = 1000
	}
	return int64(priority) * FixedPointScale / 1000
}

func calculateFairShareScore(usage tenantUsage, quota tenantQuota) int64 {
	maxRatio := int64(0)
	hasQuota := false

	maxRatio = maxRatioFixed(maxRatio, usage.RunningJobs, quota.MaxRunningJobs, &hasQuota)
	maxRatio = maxRatioFixed(maxRatio, usage.RunningNodes, quota.MaxRunningNodes, &hasQuota)
	maxRatio = maxRatioFixed(maxRatio, usage.RunningCPUCores, quota.MaxRunningCPUCores, &hasQuota)
	maxRatio = maxRatioFixed(maxRatio, usage.RunningMemoryGB, quota.MaxRunningMemoryGB, &hasQuota)
	maxRatio = maxRatioFixed(maxRatio, usage.RunningGPUs, quota.MaxRunningGPUs, &hasQuota)

	if !hasQuota {
		return FixedPointScale
	}
	if maxRatio >= FixedPointScale {
		return 0
	}
	return FixedPointScale - maxRatio
}

func maxRatioFixed(current int64, value int32, max int32, hasQuota *bool) int64 {
	if max <= 0 {
		return current
	}
	*hasQuota = true
	ratio := int64(value) * FixedPointScale / int64(max)
	if ratio > current {
		return ratio
	}
	return current
}

func (k Keeper) preemptionPlan(ctx sdk.Context, cluster types.HPCCluster, job *types.HPCJob, partitionPriority int32) ([]string, int32) {
	if job.Resources.Nodes <= 0 {
		return nil, 0
	}

	jobPriority := calculatePriorityScore(ctx.BlockTime(), job, cluster, partitionPriority)
	required := job.Resources.Nodes - cluster.AvailableNodes
	if required <= 0 {
		return nil, 0
	}

	type candidateJob struct {
		job           types.HPCJob
		priorityScore int64
	}

	var running []candidateJob
	k.WithJobs(ctx, func(existing types.HPCJob) bool {
		if existing.ClusterID != cluster.ClusterID || existing.State != types.JobStateRunning {
			return false
		}
		if existing.Resources.Nodes <= 0 {
			return false
		}

		priority := calculatePriorityScore(ctx.BlockTime(), &existing, cluster, resolvePartitionPriority(cluster, existing.QueueName))
		if priority < jobPriority {
			running = append(running, candidateJob{job: existing, priorityScore: priority})
		}
		return false
	})

	sort.Slice(running, func(i, j int) bool {
		if running[i].priorityScore == running[j].priorityScore {
			return running[i].job.JobID < running[j].job.JobID
		}
		return running[i].priorityScore < running[j].priorityScore
	})

	var freed int32
	var preempted []string
	for _, entry := range running {
		preempted = append(preempted, entry.job.JobID)
		freed = saturatingAddInt32(freed, entry.job.Resources.Nodes)
		if freed >= required {
			break
		}
	}

	return preempted, freed
}

func calculatePriorityScore(now time.Time, job *types.HPCJob, cluster types.HPCCluster, partitionPriority int32) int64 {
	ageScore := calculateAgeScore(now, job.CreatedAt)
	jobSizeScore := calculateJobSizeScore(job, cluster)
	partitionScore := calculatePartitionScore(partitionPriority)
	return (priorityWeightAge*ageScore + priorityWeightJobSize*jobSizeScore + priorityWeightPartition*partitionScore) / FixedPointScale
}

func (k Keeper) evaluateBackfill(ctx sdk.Context, cluster types.HPCCluster, job *types.HPCJob, partitionPriority int32) (bool, int64, string) {
	currentPriority := calculatePriorityScore(ctx.BlockTime(), job, cluster, partitionPriority)
	highestQueued := k.highestQueuedPriority(ctx, cluster)
	if highestQueued <= currentPriority {
		return false, 0, ""
	}

	if !backfillEnabled {
		return false, 0, "Backfill disabled: higher-priority job waiting"
	}

	if job.MaxRuntimeSeconds <= 0 || job.MaxRuntimeSeconds > backfillWindowSeconds {
		return false, backfillWindowSeconds, "Backfill would delay higher-priority job"
	}

	return true, backfillWindowSeconds, ""
}

func (k Keeper) highestQueuedPriority(ctx sdk.Context, cluster types.HPCCluster) int64 {
	var highest int64
	k.WithJobs(ctx, func(job types.HPCJob) bool {
		if job.ClusterID != cluster.ClusterID {
			return false
		}
		if job.State != types.JobStatePending && job.State != types.JobStateQueued {
			return false
		}

		priority := calculatePriorityScore(ctx.BlockTime(), &job, cluster, resolvePartitionPriority(cluster, job.QueueName))
		if priority > highest {
			highest = priority
		}
		return false
	})

	return highest
}

func schedulingBlockedByQuota(candidates []types.ClusterCandidate) bool {
	if len(candidates) == 0 {
		return false
	}
	for _, candidate := range candidates {
		if candidate.Eligible {
			return false
		}
		if !strings.Contains(strings.ToLower(candidate.IneligibilityReason), "quota") {
			return false
		}
	}
	return true
}

func safeInt32Mul(a, b int32) int32 {
	if a == 0 || b == 0 {
		return 0
	}
	val := int64(a) * int64(b)
	return clampInt32(val)
}

func saturatingAddInt32(a, b int32) int32 {
	val := int64(a) + int64(b)
	return clampInt32(val)
}

func clampInt32(val int64) int32 {
	if val > math.MaxInt32 {
		return math.MaxInt32
	}
	if val < math.MinInt32 {
		return math.MinInt32
	}
	return int32(val)
}

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
func formatDecisionReason(selected *types.ClusterCandidate, isFallback bool, decision *types.SchedulingDecision) string {
	if selected == nil {
		return "No cluster selected"
	}

	var base string
	if isFallback {
		base = fmt.Sprintf("Fallback selection: cluster %s with capacity score %s",
			selected.ClusterID, selected.CapacityScore)
	} else {
		base = fmt.Sprintf("Optimal selection: cluster %s with latency %dms and combined score %s",
			selected.ClusterID, selected.AvgLatencyMs, selected.CombinedScore)
	}

	if decision == nil {
		return base
	}

	return fmt.Sprintf("%s (priority=%s fair_share=%s preemption=%t backfill=%t quota_burst=%t)",
		base,
		decision.PriorityScore,
		decision.FairShareScore,
		decision.PreemptionPlanned,
		decision.BackfillUsed,
		decision.QuotaBurstUsed,
	)
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

		// Optional resource checks when cluster metadata is available
		if resources.CPUCoresPerNode > 0 && cluster.ClusterMetadata.TotalCPUCores > 0 {
			requiredCores := int64(resources.CPUCoresPerNode) * int64(resources.Nodes)
			if requiredCores > cluster.ClusterMetadata.TotalCPUCores {
				continue
			}
		}

		if resources.MemoryGBPerNode > 0 && cluster.ClusterMetadata.TotalMemoryGB > 0 {
			requiredMemory := int64(resources.MemoryGBPerNode) * int64(resources.Nodes)
			if requiredMemory > cluster.ClusterMetadata.TotalMemoryGB {
				continue
			}
		}

		if resources.GPUsPerNode > 0 {
			requiredGPUs := int64(resources.GPUsPerNode) * int64(resources.Nodes)
			if requiredGPUs > cluster.ClusterMetadata.TotalGPUs {
				continue
			}
		}

		eligible = append(eligible, cluster)
	}

	return eligible
}
