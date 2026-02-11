// Package types contains types for the HPC module.
//
// VE-78C: Scheduling fairness metrics and quota tracking.
package types

import (
	"fmt"
	"time"
)

// SchedulingMetricsAllClusters is the aggregate cluster ID for metrics.
const SchedulingMetricsAllClusters = "all"

// SchedulingMetrics captures aggregated scheduling metrics for a queue and cluster.
type SchedulingMetrics struct {
	// ClusterID is the cluster the metrics apply to (or "all" for aggregate).
	ClusterID string `json:"cluster_id"`

	// QueueName is the queue/partition name.
	QueueName string `json:"queue_name"`

	// TotalDecisions is the number of scheduling decisions recorded.
	TotalDecisions uint64 `json:"total_decisions"`

	// PreemptionPlanned is the number of decisions that planned preemption.
	PreemptionPlanned uint64 `json:"preemption_planned"`

	// BackfillUsed is the number of decisions that used backfill.
	BackfillUsed uint64 `json:"backfill_used"`

	// QuotaBurstUsed is the number of decisions that used burst quota.
	QuotaBurstUsed uint64 `json:"quota_burst_used"`

	// QuotaDenied is the number of decisions denied due to quota.
	QuotaDenied uint64 `json:"quota_denied"`

	// AvgLatencyScore is the average latency score (fixed-point, 6 decimals).
	AvgLatencyScore string `json:"avg_latency_score"`

	// AvgCapacityScore is the average capacity score (fixed-point, 6 decimals).
	AvgCapacityScore string `json:"avg_capacity_score"`

	// AvgCombinedScore is the average combined score (fixed-point, 6 decimals).
	AvgCombinedScore string `json:"avg_combined_score"`

	// AvgPriorityScore is the average priority score (fixed-point, 6 decimals).
	AvgPriorityScore string `json:"avg_priority_score"`

	// AvgFairShareScore is the average fair-share score (fixed-point, 6 decimals).
	AvgFairShareScore string `json:"avg_fair_share_score"`

	// AvgAgeScore is the average age-based score (fixed-point, 6 decimals).
	AvgAgeScore string `json:"avg_age_score"`

	// AvgJobSizeScore is the average job size score (fixed-point, 6 decimals).
	AvgJobSizeScore string `json:"avg_job_size_score"`

	// AvgPartitionScore is the average partition score (fixed-point, 6 decimals).
	AvgPartitionScore string `json:"avg_partition_score"`

	// LastDecisionAt is the last decision time.
	LastDecisionAt time.Time `json:"last_decision_at"`

	// BlockHeight is the height of the last update.
	BlockHeight int64 `json:"block_height"`
}

// Validate validates scheduling metrics.
func (m *SchedulingMetrics) Validate() error {
	if m.ClusterID == "" {
		return fmt.Errorf("cluster_id cannot be empty")
	}
	if m.QueueName == "" {
		return fmt.Errorf("queue_name cannot be empty")
	}
	return nil
}
