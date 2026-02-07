package data_vault

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/virtengine/virtengine/pkg/observability"
)

// VaultMetrics tracks data vault metrics.
type VaultMetrics struct {
	AccessTotal       *prometheus.CounterVec
	AccessDeniedTotal *prometheus.CounterVec
	AuditFailures     prometheus.Counter
}

var (
	vaultMetricsOnce sync.Once
	vaultMetrics     *VaultMetrics
)

// NewVaultMetrics registers vault metrics in the global registry.
func NewVaultMetrics() *VaultMetrics {
	vaultMetricsOnce.Do(func() {
		reg := observability.GetRegistry()

		accessTotal := prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "virtengine",
			Subsystem: "data_vault",
			Name:      "access_total",
			Help:      "Total data vault access attempts",
		}, []string{"scope", "action", "success"})

		accessDenied := prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "virtengine",
			Subsystem: "data_vault",
			Name:      "access_denied_total",
			Help:      "Total denied access attempts",
		}, []string{"scope", "action"})

		auditFailures := prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "virtengine",
			Subsystem: "data_vault",
			Name:      "audit_failures_total",
			Help:      "Total audit logging failures",
		})

		reg.MustRegister(accessTotal, accessDenied, auditFailures)

		vaultMetrics = &VaultMetrics{
			AccessTotal:       accessTotal,
			AccessDeniedTotal: accessDenied,
			AuditFailures:     auditFailures,
		}
	})

	return vaultMetrics
}

// AccessAnomalyDetector detects repeated access failures in a window.
type AccessAnomalyDetector struct {
	threshold int
	window    time.Duration
	mu        sync.Mutex
	failures  map[string][]time.Time
	onAlert   func(key string, count int)
}

// NewAccessAnomalyDetector creates an anomaly detector.
func NewAccessAnomalyDetector(threshold int, window time.Duration, onAlert func(key string, count int)) *AccessAnomalyDetector {
	return &AccessAnomalyDetector{
		threshold: threshold,
		window:    window,
		failures:  make(map[string][]time.Time),
		onAlert:   onAlert,
	}
}

// RecordFailure records a failed access attempt.
func (d *AccessAnomalyDetector) RecordFailure(key string) {
	if d == nil || d.threshold <= 0 || d.window <= 0 {
		return
	}
	now := time.Now().UTC()

	d.mu.Lock()
	defer d.mu.Unlock()

	history := d.failures[key]
	cutoff := now.Add(-d.window)
	filtered := history[:0]
	for _, ts := range history {
		if ts.After(cutoff) {
			filtered = append(filtered, ts)
		}
	}
	filtered = append(filtered, now)
	d.failures[key] = filtered

	if len(filtered) >= d.threshold && d.onAlert != nil {
		d.onAlert(key, len(filtered))
	}
}
