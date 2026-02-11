// Package keeper implements the HPC module keeper.
//
// VE-78C: Scheduling fairness metrics.
package keeper

import (
	"encoding/json"
	"math"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/virtengine/virtengine/x/hpc/types"
)

// GetSchedulingMetrics retrieves scheduling metrics for a cluster and queue.
func (k Keeper) GetSchedulingMetrics(ctx sdk.Context, clusterID, queueName string) (types.SchedulingMetrics, bool) {
	store := ctx.KVStore(k.skey)
	bz := store.Get(types.GetSchedulingMetricsKey(clusterID, queueName))
	if bz == nil {
		return types.SchedulingMetrics{}, false
	}

	var metrics types.SchedulingMetrics
	if err := json.Unmarshal(bz, &metrics); err != nil {
		return types.SchedulingMetrics{}, false
	}
	return metrics, true
}

// SetSchedulingMetrics stores scheduling metrics.
func (k Keeper) SetSchedulingMetrics(ctx sdk.Context, metrics types.SchedulingMetrics) error {
	store := ctx.KVStore(k.skey)
	bz, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	store.Set(types.GetSchedulingMetricsKey(metrics.ClusterID, metrics.QueueName), bz)
	return nil
}

// WithSchedulingMetrics iterates over scheduling metrics.
func (k Keeper) WithSchedulingMetrics(ctx sdk.Context, fn func(types.SchedulingMetrics) bool) {
	store := ctx.KVStore(k.skey)
	iter := storetypes.KVStorePrefixIterator(store, types.SchedulingMetricsPrefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var metrics types.SchedulingMetrics
		if err := json.Unmarshal(iter.Value(), &metrics); err != nil {
			continue
		}
		if fn(metrics) {
			break
		}
	}
}

func (k Keeper) updateSchedulingMetrics(ctx sdk.Context, job *types.HPCJob, decision *types.SchedulingDecision) {
	if job == nil || decision == nil {
		return
	}

	queueName := job.QueueName
	if queueName == "" {
		queueName = defaultQueueName
	}

	if decision.SelectedClusterID != "" {
		k.updateSchedulingMetricsForKey(ctx, decision.SelectedClusterID, queueName, decision)
	}
	k.updateSchedulingMetricsForKey(ctx, types.SchedulingMetricsAllClusters, queueName, decision)
}

func (k Keeper) updateSchedulingMetricsForKey(ctx sdk.Context, clusterID, queueName string, decision *types.SchedulingDecision) {
	if clusterID == "" || queueName == "" || decision == nil {
		return
	}

	metrics, found := k.GetSchedulingMetrics(ctx, clusterID, queueName)
	if !found {
		metrics = types.SchedulingMetrics{
			ClusterID: clusterID,
			QueueName: queueName,
		}
	}

	applyDecisionToMetrics(&metrics, decision)
	_ = k.SetSchedulingMetrics(ctx, metrics)
}

func (k Keeper) recordQuotaDenied(ctx sdk.Context, queueName string) {
	if queueName == "" {
		queueName = defaultQueueName
	}

	metrics, found := k.GetSchedulingMetrics(ctx, types.SchedulingMetricsAllClusters, queueName)
	if !found {
		metrics = types.SchedulingMetrics{
			ClusterID: types.SchedulingMetricsAllClusters,
			QueueName: queueName,
		}
	}
	metrics.QuotaDenied++
	metrics.LastDecisionAt = ctx.BlockTime()
	metrics.BlockHeight = ctx.BlockHeight()
	_ = k.SetSchedulingMetrics(ctx, metrics)
}

func applyDecisionToMetrics(metrics *types.SchedulingMetrics, decision *types.SchedulingDecision) {
	if metrics == nil || decision == nil {
		return
	}

	count := metrics.TotalDecisions
	metrics.AvgLatencyScore = updateAverageFixed(metrics.AvgLatencyScore, count, parseFixedPoint(decision.LatencyScore))
	metrics.AvgCapacityScore = updateAverageFixed(metrics.AvgCapacityScore, count, parseFixedPoint(decision.CapacityScore))
	metrics.AvgCombinedScore = updateAverageFixed(metrics.AvgCombinedScore, count, parseFixedPoint(decision.CombinedScore))
	metrics.AvgPriorityScore = updateAverageFixed(metrics.AvgPriorityScore, count, parseFixedPoint(decision.PriorityScore))
	metrics.AvgFairShareScore = updateAverageFixed(metrics.AvgFairShareScore, count, parseFixedPoint(decision.FairShareScore))
	metrics.AvgAgeScore = updateAverageFixed(metrics.AvgAgeScore, count, parseFixedPoint(decision.AgeScore))
	metrics.AvgJobSizeScore = updateAverageFixed(metrics.AvgJobSizeScore, count, parseFixedPoint(decision.JobSizeScore))
	metrics.AvgPartitionScore = updateAverageFixed(metrics.AvgPartitionScore, count, parseFixedPoint(decision.PartitionScore))

	metrics.TotalDecisions = count + 1
	if decision.PreemptionPlanned {
		metrics.PreemptionPlanned++
	}
	if decision.BackfillUsed {
		metrics.BackfillUsed++
	}
	if decision.QuotaBurstUsed {
		metrics.QuotaBurstUsed++
	}
	metrics.LastDecisionAt = decision.CreatedAt
	metrics.BlockHeight = decision.BlockHeight
}

func updateAverageFixed(prev string, count uint64, value int64) string {
	if count == 0 {
		return formatFixedPoint(value)
	}
	prevVal := parseFixedPoint(prev)
	if count > math.MaxInt64-1 {
		return formatFixedPoint(prevVal)
	}
	count64 := int64(count)
	newVal := (prevVal*count64 + value) / (count64 + 1)
	return formatFixedPoint(newVal)
}
