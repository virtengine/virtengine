// Package provider_daemon implements the provider daemon for VirtEngine.
//
// SCALE-002: Prometheus metrics for horizontal scaling observability
package provider_daemon

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// metricsOnce ensures metrics are only registered once
	metricsOnce sync.Once

	// Provider daemon scaling metrics
	providerDaemonOrdersReceived = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "provider_daemon",
			Name:      "orders_received_total",
			Help:      "Total number of orders received by this instance",
		},
		[]string{"instance_id"},
	)

	providerDaemonOrdersProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "provider_daemon",
			Name:      "orders_processed_total",
			Help:      "Total number of orders processed by this instance",
		},
		[]string{"instance_id", "status"},
	)

	providerDaemonOrdersSkipped = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "provider_daemon",
			Name:      "orders_skipped_total",
			Help:      "Total number of orders skipped",
		},
		[]string{"instance_id", "reason"},
	)

	providerDaemonBidsSubmitted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "provider_daemon",
			Name:      "bids_submitted_total",
			Help:      "Total number of bids successfully submitted",
		},
		[]string{"instance_id"},
	)

	providerDaemonBidsFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "provider_daemon",
			Name:      "bids_failed_total",
			Help:      "Total number of failed bid submissions",
		},
		[]string{"instance_id", "error_type"},
	)

	providerDaemonPendingOrders = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "provider_daemon",
			Name:      "pending_orders",
			Help:      "Current number of pending orders in the queue",
		},
		[]string{"instance_id"},
	)

	providerDaemonActiveLeases = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "provider_daemon",
			Name:      "active_leases",
			Help:      "Current number of active leases managed by this instance",
		},
		[]string{"instance_id"},
	)

	providerDaemonClaimConflicts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "provider_daemon",
			Name:      "claim_conflicts_total",
			Help:      "Total number of bid claim conflicts in distributed deduplication",
		},
		[]string{"instance_id"},
	)

	providerDaemonBidDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "provider_daemon",
			Name:      "bid_duration_seconds",
			Help:      "Time taken to process a bid",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"instance_id", "status"},
	)

	providerDaemonPartitionAssignment = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "provider_daemon",
			Name:      "partition_assignment",
			Help:      "Partition assignment for this instance (1 = active)",
		},
		[]string{"instance_id", "partition", "total_instances"},
	)

	// Scaling health metrics
	scalingInstanceInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scaling",
			Name:      "instance_info",
			Help:      "Information about this scaling instance (value is always 1)",
		},
		[]string{"instance_id", "partition_mode", "dedup_enabled"},
	)

	scalingHealthy = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "scaling",
			Name:      "healthy",
			Help:      "Whether this instance is healthy (1 = healthy, 0 = unhealthy)",
		},
		[]string{"instance_id"},
	)

	// RPC and connection metrics for full nodes
	virtengineRPCRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "virtengine",
			Name:      "rpc_requests_total",
			Help:      "Total number of RPC requests",
		},
		[]string{"method", "status"},
	)

	virtengineRPCDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "virtengine",
			Name:      "rpc_duration_seconds",
			Help:      "RPC request duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"method"},
	)

	virtengineActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "virtengine",
			Name:      "active_websocket_connections",
			Help:      "Current number of active WebSocket connections",
		},
	)

	// State sync metrics
	stateSyncSnapshotLastCreated = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "virtengine",
			Name:      "statesync_snapshot_last_created_timestamp",
			Help:      "Timestamp of last state sync snapshot creation",
		},
	)

	stateSyncSnapshotHeight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "virtengine",
			Name:      "statesync_snapshot_height",
			Help:      "Block height of most recent state sync snapshot",
		},
	)

	stateSyncChunkFetchFailures = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "virtengine",
			Name:      "statesync_chunk_fetch_failures_total",
			Help:      "Total number of state sync chunk fetch failures",
		},
	)

	// Marketplace metrics
	marketplaceOpenOrders = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "virtengine",
			Name:      "marketplace_order_count",
			Help:      "Current count of marketplace orders by status",
		},
		[]string{"status"},
	)

	marketplaceActiveLeases = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "virtengine",
			Name:      "active_leases_total",
			Help:      "Total number of active leases across all providers",
		},
	)

	// Cross-region metrics
	crossRegionRPCDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "virtengine",
			Name:      "cross_region_rpc_duration_seconds",
			Help:      "Cross-region RPC latency in seconds",
			Buckets:   []float64{0.05, 0.1, 0.2, 0.3, 0.5, 0.75, 1.0, 2.0, 5.0},
		},
		[]string{"source_region", "target_region"},
	)
)

// ScalingMetricsCollector provides methods to update scaling metrics
type ScalingMetricsCollector struct {
	instanceID string
}

// NewScalingMetricsCollector creates a new metrics collector for the given instance
func NewScalingMetricsCollector(instanceID string) *ScalingMetricsCollector {
	return &ScalingMetricsCollector{
		instanceID: instanceID,
	}
}

// RecordOrderReceived increments the orders received counter
func (m *ScalingMetricsCollector) RecordOrderReceived() {
	providerDaemonOrdersReceived.WithLabelValues(m.instanceID).Inc()
}

// RecordOrderProcessed increments the orders processed counter
func (m *ScalingMetricsCollector) RecordOrderProcessed(status string) {
	providerDaemonOrdersProcessed.WithLabelValues(m.instanceID, status).Inc()
}

// RecordOrderSkipped increments the orders skipped counter
func (m *ScalingMetricsCollector) RecordOrderSkipped(reason string) {
	providerDaemonOrdersSkipped.WithLabelValues(m.instanceID, reason).Inc()
}

// RecordBidSubmitted increments the bids submitted counter
func (m *ScalingMetricsCollector) RecordBidSubmitted() {
	providerDaemonBidsSubmitted.WithLabelValues(m.instanceID).Inc()
}

// RecordBidFailed increments the bids failed counter
func (m *ScalingMetricsCollector) RecordBidFailed(errorType string) {
	providerDaemonBidsFailed.WithLabelValues(m.instanceID, errorType).Inc()
}

// SetPendingOrders sets the current pending orders gauge
func (m *ScalingMetricsCollector) SetPendingOrders(count float64) {
	providerDaemonPendingOrders.WithLabelValues(m.instanceID).Set(count)
}

// SetActiveLeases sets the current active leases gauge
func (m *ScalingMetricsCollector) SetActiveLeases(count float64) {
	providerDaemonActiveLeases.WithLabelValues(m.instanceID).Set(count)
}

// RecordClaimConflict increments the claim conflicts counter
func (m *ScalingMetricsCollector) RecordClaimConflict() {
	providerDaemonClaimConflicts.WithLabelValues(m.instanceID).Inc()
}

// ObserveBidDuration records the duration of a bid operation
func (m *ScalingMetricsCollector) ObserveBidDuration(seconds float64, status string) {
	providerDaemonBidDuration.WithLabelValues(m.instanceID, status).Observe(seconds)
}

// SetPartitionAssignment records partition assignment info
func (m *ScalingMetricsCollector) SetPartitionAssignment(partition, totalInstances string) {
	providerDaemonPartitionAssignment.WithLabelValues(m.instanceID, partition, totalInstances).Set(1)
}

// SetInstanceInfo records instance metadata
func (m *ScalingMetricsCollector) SetInstanceInfo(partitionMode string, dedupEnabled bool) {
	dedupStr := "false"
	if dedupEnabled {
		dedupStr = "true"
	}
	scalingInstanceInfo.WithLabelValues(m.instanceID, partitionMode, dedupStr).Set(1)
}

// SetHealthy sets the healthy status of this instance
func (m *ScalingMetricsCollector) SetHealthy(healthy bool) {
	val := float64(0)
	if healthy {
		val = 1
	}
	scalingHealthy.WithLabelValues(m.instanceID).Set(val)
}

// RecordRPCRequest records an RPC request
func RecordRPCRequest(method, status string) {
	virtengineRPCRequests.WithLabelValues(method, status).Inc()
}

// ObserveRPCDuration records the duration of an RPC request
func ObserveRPCDuration(method string, seconds float64) {
	virtengineRPCDuration.WithLabelValues(method).Observe(seconds)
}

// SetActiveConnections sets the number of active WebSocket connections
func SetActiveConnections(count float64) {
	virtengineActiveConnections.Set(count)
}

// RecordStateSyncSnapshot records a new state sync snapshot
func RecordStateSyncSnapshot(height float64, timestamp float64) {
	stateSyncSnapshotHeight.Set(height)
	stateSyncSnapshotLastCreated.Set(timestamp)
}

// RecordStateSyncChunkFailure records a state sync chunk fetch failure
func RecordStateSyncChunkFailure() {
	stateSyncChunkFetchFailures.Inc()
}

// SetMarketplaceOrderCount sets the count of marketplace orders by status
func SetMarketplaceOrderCount(status string, count float64) {
	marketplaceOpenOrders.WithLabelValues(status).Set(count)
}

// SetActiveLeaseCount sets the total active lease count
func SetActiveLeaseCount(count float64) {
	marketplaceActiveLeases.Set(count)
}

// ObserveCrossRegionLatency records cross-region RPC latency
func ObserveCrossRegionLatency(sourceRegion, targetRegion string, seconds float64) {
	crossRegionRPCDuration.WithLabelValues(sourceRegion, targetRegion).Observe(seconds)
}

