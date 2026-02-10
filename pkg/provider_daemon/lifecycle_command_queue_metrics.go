// Package provider_daemon implements provider-side services for VirtEngine.
//
// VE-34E: Prometheus metrics for lifecycle command queue.
package provider_daemon

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// LifecycleQueueMetrics captures queue and reconciliation metrics.
type LifecycleQueueMetrics struct {
	QueueDepth        *prometheus.GaugeVec
	RetriesTotal      *prometheus.CounterVec
	CommandsTotal     *prometheus.CounterVec
	ReconcileRuns     *prometheus.CounterVec
	ReconcileCommands *prometheus.CounterVec
}

// NewLifecycleQueueMetrics registers lifecycle queue metrics.
func NewLifecycleQueueMetrics() *LifecycleQueueMetrics {
	return &LifecycleQueueMetrics{
		QueueDepth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: "provider_daemon",
				Subsystem: "lifecycle_queue",
				Name:      "depth",
				Help:      "Current number of lifecycle commands by status",
			},
			[]string{"status"},
		),
		RetriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "provider_daemon",
				Subsystem: "lifecycle_queue",
				Name:      "retries_total",
				Help:      "Total number of lifecycle command retries",
			},
			[]string{"action"},
		),
		CommandsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "provider_daemon",
				Subsystem: "lifecycle_queue",
				Name:      "commands_total",
				Help:      "Total lifecycle commands by action and outcome",
			},
			[]string{"action", "outcome"},
		),
		ReconcileRuns: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "provider_daemon",
				Subsystem: "lifecycle_reconcile",
				Name:      "runs_total",
				Help:      "Total reconciliation cycles by outcome",
			},
			[]string{"outcome"},
		),
		ReconcileCommands: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "provider_daemon",
				Subsystem: "lifecycle_reconcile",
				Name:      "commands_total",
				Help:      "Total reconciled lifecycle commands by action and outcome",
			},
			[]string{"action", "outcome"},
		),
	}
}
