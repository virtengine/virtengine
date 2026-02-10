/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package sla

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Monitor evaluates SLA metrics, detects breaches, and triggers alerts.
type Monitor struct {
	metricsCollector MetricsCollector
	slaStore         SLAStore
	alerter          Alerter
	remediator       Remediator

	mu            sync.RWMutex
	currentStatus map[string]*SLAStatus

	checkInterval time.Duration
	warningBuffer decimal.Decimal
	nowFn         func() time.Time
}

// MonitorOption configures the SLA monitor.
type MonitorOption func(*Monitor)

// WithCheckInterval sets the monitor interval.
func WithCheckInterval(interval time.Duration) MonitorOption {
	return func(m *Monitor) {
		m.checkInterval = interval
	}
}

// WithWarningBuffer sets the warning buffer percent (e.g. 0.05 for 5%).
func WithWarningBuffer(buffer decimal.Decimal) MonitorOption {
	return func(m *Monitor) {
		m.warningBuffer = buffer
	}
}

// WithNowFn sets a custom clock for testing.
func WithNowFn(nowFn func() time.Time) MonitorOption {
	return func(m *Monitor) {
		m.nowFn = nowFn
	}
}

// NewMonitor constructs a Monitor.
func NewMonitor(
	collector MetricsCollector,
	store SLAStore,
	alerter Alerter,
	remediator Remediator,
	opts ...MonitorOption,
) *Monitor {
	monitor := &Monitor{
		metricsCollector: collector,
		slaStore:         store,
		alerter:          alerter,
		remediator:       remediator,
		currentStatus:    make(map[string]*SLAStatus),
		checkInterval:    1 * time.Minute,
		warningBuffer:    decimal.NewFromFloat(0.05),
		nowFn:            time.Now,
	}

	for _, opt := range opts {
		opt(monitor)
	}

	return monitor
}

// Start begins periodic SLA checks until the context is canceled.
func (m *Monitor) Start(ctx context.Context) {
	if m == nil {
		return
	}

	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkAllLeases(ctx)
		}
	}
}

func (m *Monitor) checkAllLeases(ctx context.Context) {
	if m.slaStore == nil || m.metricsCollector == nil {
		return
	}

	leases, err := m.slaStore.GetActiveLeases(ctx)
	if err != nil {
		return
	}

	for _, lease := range leases {
		if ctx.Err() != nil {
			return
		}
		m.checkLease(ctx, lease)
	}
}

func (m *Monitor) checkLease(ctx context.Context, lease LeaseInfo) {
	slaDef, err := m.slaStore.GetSLADefinition(ctx, lease.ProviderID)
	if err != nil || slaDef == nil {
		return
	}

	status := &SLAStatus{
		LeaseID:       lease.ID,
		ProviderID:    lease.ProviderID,
		CustomerID:    lease.CustomerID,
		SLADefinition: slaDef,
		Metrics:       make(map[SLAMetricType]*MetricStatus),
		OverallStatus: StatusHealthy,
		LastChecked:   m.nowFn(),
	}

	for _, metric := range slaDef.Metrics {
		currentValue, err := m.metricsCollector.GetMetric(ctx, lease.ID, metric.Type, metric.Window)
		if err != nil {
			continue
		}

		metricStatus := m.evaluateMetric(metric, currentValue)
		status.Metrics[metric.Type] = metricStatus

		if metricStatus.Status == StatusBreached {
			status.OverallStatus = StatusBreached
		} else if metricStatus.Status == StatusWarning && status.OverallStatus != StatusBreached {
			status.OverallStatus = StatusWarning
		}
	}

	m.mu.Lock()
	oldStatus := m.currentStatus[lease.ID]
	m.currentStatus[lease.ID] = status
	m.mu.Unlock()

	m.handleStatusChange(ctx, oldStatus, status)
}

func (m *Monitor) evaluateMetric(metric SLAMetric, currentValue decimal.Decimal) *MetricStatus {
	status := &MetricStatus{
		Type:         metric.Type,
		CurrentValue: currentValue,
		Target:       metric.Target,
		LastUpdated:  m.nowFn(),
	}

	breached := false
	warning := false

	one := decimal.NewFromInt(1)

	switch metric.Comparison {
	case ComparisonGTE:
		if currentValue.LessThan(metric.Target) {
			breached = true
		} else {
			warningThreshold := metric.Target.Mul(one.Add(m.warningBuffer))
			warning = currentValue.LessThan(warningThreshold)
		}
	case ComparisonLTE:
		if currentValue.GreaterThan(metric.Target) {
			breached = true
		} else {
			warningThreshold := metric.Target.Mul(one.Sub(m.warningBuffer))
			warning = currentValue.GreaterThan(warningThreshold)
		}
	default:
		breached = false
		warning = false
	}

	if breached {
		status.Status = StatusBreached
	} else if warning {
		status.Status = StatusWarning
	} else {
		status.Status = StatusHealthy
	}

	return status
}

func (m *Monitor) handleStatusChange(ctx context.Context, oldStatus, newStatus *SLAStatus) {
	if newStatus == nil {
		return
	}

	if oldStatus == nil {
		return
	}

	for metricType, newMetric := range newStatus.Metrics {
		oldMetric, exists := oldStatus.Metrics[metricType]

		if newMetric.Status == StatusBreached && (!exists || oldMetric.Status != StatusBreached) {
			m.handleBreach(ctx, newStatus, metricType, newMetric)
			continue
		}

		if newMetric.Status == StatusWarning && (!exists || oldMetric.Status == StatusHealthy) {
			m.handleWarning(ctx, newStatus, metricType, newMetric)
			continue
		}

		if newMetric.Status == StatusHealthy && exists && oldMetric.Status == StatusBreached {
			m.handleBreachResolved(ctx, newStatus, metricType)
		}
	}
}

func (m *Monitor) handleBreach(ctx context.Context, status *SLAStatus, metricType SLAMetricType, metric *MetricStatus) {
	breach := &BreachRecord{
		ID:          uuid.NewString(),
		LeaseID:     status.LeaseID,
		MetricType:  metricType,
		TargetValue: metric.Target,
		ActualValue: metric.CurrentValue,
		StartedAt:   m.nowFn(),
	}

	status.ActiveBreach = breach

	if m.slaStore != nil {
		_ = m.slaStore.CreateBreach(ctx, breach)
	}

	if m.alerter != nil {
		_ = m.alerter.SendAlert(ctx, Alert{
			Severity: AlertCritical,
			Title:    fmt.Sprintf("SLA Breach: %s for lease %s", metricType, status.LeaseID),
			Message: fmt.Sprintf(
				"SLA breach detected. Metric: %s, Target: %s, Current: %s",
				metricType, metric.Target, metric.CurrentValue,
			),
			Labels: map[string]string{
				"lease_id":    status.LeaseID,
				"provider_id": status.ProviderID,
				"customer_id": status.CustomerID,
				"metric":      string(metricType),
			},
		})

		_ = m.alerter.NotifyCustomer(ctx, status.CustomerID, CustomerNotification{
			Type:    "sla_breach",
			Title:   "SLA Breach Detected",
			Message: "An SLA breach has been detected for your service. Credit compensation will be applied automatically.",
			LeaseID: status.LeaseID,
		})
	}

	if m.remediator != nil {
		m.remediator.HandleBreach(ctx, breach)
	}
}

func (m *Monitor) handleWarning(ctx context.Context, status *SLAStatus, metricType SLAMetricType, metric *MetricStatus) {
	if m.alerter != nil {
		_ = m.alerter.SendAlert(ctx, Alert{
			Severity: AlertWarning,
			Title:    fmt.Sprintf("SLA Warning: %s approaching threshold", metricType),
			Message: fmt.Sprintf(
				"SLA metric approaching breach threshold. Metric: %s, Target: %s, Current: %s",
				metricType, metric.Target, metric.CurrentValue,
			),
			Labels: map[string]string{
				"lease_id":    status.LeaseID,
				"provider_id": status.ProviderID,
				"metric":      string(metricType),
			},
		})

		_ = m.alerter.NotifyProvider(ctx, status.ProviderID, ProviderNotification{
			Type:    "sla_warning",
			Title:   "SLA Warning",
			Message: "SLA metrics are approaching the breach threshold. Please investigate.",
			LeaseID: status.LeaseID,
			Metric:  metricType,
		})
	}

	if m.remediator != nil {
		m.remediator.HandleWarning(ctx, status, metricType)
	}
}

func (m *Monitor) handleBreachResolved(ctx context.Context, status *SLAStatus, metricType SLAMetricType) {
	if m.slaStore == nil {
		return
	}

	breach, err := m.slaStore.GetActiveBreach(ctx, status.LeaseID, metricType)
	if err != nil || breach == nil {
		return
	}

	now := m.nowFn()
	breach.ResolvedAt = &now
	breach.Duration = now.Sub(breach.StartedAt)

	credit, err := CalculateCredit(status.SLADefinition, breach)
	if err == nil {
		breach.CreditAmount = credit
	}

	_ = m.slaStore.UpdateBreach(ctx, breach)

	if m.alerter != nil {
		_ = m.alerter.SendAlert(ctx, Alert{
			Severity: AlertInfo,
			Title:    fmt.Sprintf("SLA Breach Resolved: %s", metricType),
			Message:  fmt.Sprintf("SLA breach resolved after %s. Credit: %s", breach.Duration, breach.CreditAmount),
			Labels: map[string]string{
				"lease_id": status.LeaseID,
				"metric":   string(metricType),
			},
		})
	}

	if m.remediator != nil {
		m.remediator.ApplyCredit(ctx, breach)
	}
}

// GetStatus returns the latest SLA status for a lease.
func (m *Monitor) GetStatus(leaseID string) (*SLAStatus, bool) {
	if m == nil {
		return nil, false
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	status, ok := m.currentStatus[leaseID]
	return status, ok
}
