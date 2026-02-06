# Task 31N: SLA Monitoring Alerts

**vibe-kanban ID:** `77a37fa9-b3e9-40e9-a7a6-b8eb6d46e21f`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31N |
| **Title** | feat(sla): SLA monitoring alerts |
| **Priority** | P1 |
| **Wave** | 3 |
| **Estimated LOC** | 2500 |
| **Duration** | 2-3 weeks |
| **Dependencies** | 31J (Metrics Dashboards), Provider Daemon |
| **Blocking** | None |

---

## Problem Statement

Providers commit to SLAs (uptime, latency, availability) in the marketplace. Currently:
- No automated SLA monitoring
- No breach detection
- No automatic remediation
- No customer notifications on SLA violations

This creates risk for customers and providers.

### Current State Analysis

```
_docs/slos-and-playbooks.md     ✅ SLO documentation exists
pkg/provider_daemon/            ⚠️  Basic health checks only
pkg/sla/                        ❌ Does not exist
Alerting for SLA breaches:      ❌ None
```

---

## Acceptance Criteria

### AC-1: SLA Definition Framework
- [ ] Define SLA metrics schema (uptime, latency, throughput)
- [ ] SLA tier configurations (Bronze, Silver, Gold)
- [ ] Per-provider SLA commitments
- [ ] SLA calculation windows (5m, 1h, 24h, 30d)

### AC-2: SLA Monitoring Engine
- [ ] Real-time SLA metric collection
- [ ] Rolling window calculations
- [ ] Breach detection logic
- [ ] Multi-metric aggregation

### AC-3: Alerting and Notifications
- [ ] Alert rules for SLA breach risk (< 5% buffer)
- [ ] Alert rules for SLA breach
- [ ] Customer notification on breach
- [ ] Provider notification on degradation
- [ ] Escalation workflows

### AC-4: Remediation Actions
- [ ] Auto-scaling triggers
- [ ] Failover initiation
- [ ] Escrow credit calculation
- [ ] Incident report generation

---

## Technical Requirements

### SLA Types

```go
// pkg/sla/types.go

package sla

import (
    "time"
    
    "github.com/shopspring/decimal"
)

type SLAMetricType string

const (
    MetricUptime     SLAMetricType = "uptime"        // Percentage (99.9%)
    MetricLatencyP50 SLAMetricType = "latency_p50"   // Milliseconds
    MetricLatencyP99 SLAMetricType = "latency_p99"   // Milliseconds
    MetricThroughput SLAMetricType = "throughput"    // Requests/sec
    MetricErrorRate  SLAMetricType = "error_rate"    // Percentage
)

type SLATier string

const (
    TierBronze   SLATier = "bronze"
    TierSilver   SLATier = "silver"
    TierGold     SLATier = "gold"
    TierPlatinum SLATier = "platinum"
)

type SLADefinition struct {
    ID           string
    ProviderID   string
    Tier         SLATier
    Metrics      []SLAMetric
    EffectiveFrom time.Time
    EffectiveTo  *time.Time
    
    // Compensation
    CreditPolicy CreditPolicy
}

type SLAMetric struct {
    Type      SLAMetricType
    Target    decimal.Decimal  // e.g., 99.9 for uptime
    Window    time.Duration    // Calculation window (e.g., 30 days)
    Comparison ComparisonType  // gte, lte, eq
}

type ComparisonType string

const (
    ComparisonGTE ComparisonType = "gte"  // >= (for uptime)
    ComparisonLTE ComparisonType = "lte"  // <= (for latency)
)

type CreditPolicy struct {
    // Credit percentage based on breach severity
    Brackets []CreditBracket
    MaxCredit decimal.Decimal  // Maximum credit percentage
}

type CreditBracket struct {
    MinDeviation decimal.Decimal  // e.g., 0.1 (0.1% below SLA)
    MaxDeviation decimal.Decimal  // e.g., 1.0 (1% below SLA)
    CreditPercent decimal.Decimal // e.g., 10 (10% credit)
}

// Default SLA tiers
var DefaultSLATiers = map[SLATier][]SLAMetric{
    TierBronze: {
        {Type: MetricUptime, Target: decimal.NewFromFloat(99.0), Window: 30 * 24 * time.Hour, Comparison: ComparisonGTE},
        {Type: MetricLatencyP99, Target: decimal.NewFromInt(1000), Window: 24 * time.Hour, Comparison: ComparisonLTE},
    },
    TierSilver: {
        {Type: MetricUptime, Target: decimal.NewFromFloat(99.5), Window: 30 * 24 * time.Hour, Comparison: ComparisonGTE},
        {Type: MetricLatencyP99, Target: decimal.NewFromInt(500), Window: 24 * time.Hour, Comparison: ComparisonLTE},
        {Type: MetricErrorRate, Target: decimal.NewFromFloat(1.0), Window: 24 * time.Hour, Comparison: ComparisonLTE},
    },
    TierGold: {
        {Type: MetricUptime, Target: decimal.NewFromFloat(99.9), Window: 30 * 24 * time.Hour, Comparison: ComparisonGTE},
        {Type: MetricLatencyP99, Target: decimal.NewFromInt(200), Window: 24 * time.Hour, Comparison: ComparisonLTE},
        {Type: MetricErrorRate, Target: decimal.NewFromFloat(0.1), Window: 24 * time.Hour, Comparison: ComparisonLTE},
    },
    TierPlatinum: {
        {Type: MetricUptime, Target: decimal.NewFromFloat(99.99), Window: 30 * 24 * time.Hour, Comparison: ComparisonGTE},
        {Type: MetricLatencyP99, Target: decimal.NewFromInt(100), Window: 24 * time.Hour, Comparison: ComparisonLTE},
        {Type: MetricErrorRate, Target: decimal.NewFromFloat(0.01), Window: 24 * time.Hour, Comparison: ComparisonLTE},
    },
}
```

### SLA Monitor

```go
// pkg/sla/monitor.go

package sla

import (
    "context"
    "sync"
    "time"
    
    "github.com/shopspring/decimal"
)

type Monitor struct {
    metricsCollector MetricsCollector
    slaStore         SLAStore
    alerter          Alerter
    remediator       Remediator
    
    // State
    mu               sync.RWMutex
    currentStatus    map[string]*SLAStatus  // leaseID -> status
    
    // Config
    checkInterval    time.Duration
    warningBuffer    decimal.Decimal  // Alert when within X% of breach
}

type MetricsCollector interface {
    GetMetric(ctx context.Context, leaseID string, metric SLAMetricType, window time.Duration) (decimal.Decimal, error)
    GetHistoricalMetrics(ctx context.Context, leaseID string, metric SLAMetricType, start, end time.Time) ([]MetricDataPoint, error)
}

type SLAStatus struct {
    LeaseID      string
    ProviderID   string
    CustomerID   string
    SLADefinition *SLADefinition
    
    // Current state
    Metrics      map[SLAMetricType]*MetricStatus
    OverallStatus HealthStatus
    
    // Breach tracking
    ActiveBreach  *BreachRecord
    BreachHistory []BreachRecord
    
    LastChecked  time.Time
}

type MetricStatus struct {
    Type         SLAMetricType
    CurrentValue decimal.Decimal
    Target       decimal.Decimal
    Status       HealthStatus
    LastUpdated  time.Time
}

type HealthStatus string

const (
    StatusHealthy  HealthStatus = "healthy"
    StatusWarning  HealthStatus = "warning"   // Within buffer of breach
    StatusBreached HealthStatus = "breached"
)

type BreachRecord struct {
    ID           string
    LeaseID      string
    MetricType   SLAMetricType
    TargetValue  decimal.Decimal
    ActualValue  decimal.Decimal
    StartedAt    time.Time
    ResolvedAt   *time.Time
    Duration     time.Duration
    CreditAmount decimal.Decimal
}

func NewMonitor(
    collector MetricsCollector,
    store SLAStore,
    alerter Alerter,
    remediator Remediator,
) *Monitor {
    return &Monitor{
        metricsCollector: collector,
        slaStore:         store,
        alerter:          alerter,
        remediator:       remediator,
        currentStatus:    make(map[string]*SLAStatus),
        checkInterval:    1 * time.Minute,
        warningBuffer:    decimal.NewFromFloat(0.05),  // 5% buffer
    }
}

func (m *Monitor) Start(ctx context.Context) {
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
    leases, err := m.slaStore.GetActiveLeases(ctx)
    if err != nil {
        // Log error
        return
    }
    
    for _, lease := range leases {
        go m.checkLease(ctx, lease)
    }
}

func (m *Monitor) checkLease(ctx context.Context, lease LeaseInfo) {
    sla, err := m.slaStore.GetSLADefinition(ctx, lease.ProviderID)
    if err != nil {
        return
    }
    
    status := &SLAStatus{
        LeaseID:       lease.ID,
        ProviderID:    lease.ProviderID,
        CustomerID:    lease.CustomerID,
        SLADefinition: sla,
        Metrics:       make(map[SLAMetricType]*MetricStatus),
        OverallStatus: StatusHealthy,
        LastChecked:   time.Now(),
    }
    
    for _, metric := range sla.Metrics {
        currentValue, err := m.metricsCollector.GetMetric(ctx, lease.ID, metric.Type, metric.Window)
        if err != nil {
            continue
        }
        
        metricStatus := m.evaluateMetric(metric, currentValue)
        status.Metrics[metric.Type] = metricStatus
        
        // Update overall status
        if metricStatus.Status == StatusBreached {
            status.OverallStatus = StatusBreached
        } else if metricStatus.Status == StatusWarning && status.OverallStatus != StatusBreached {
            status.OverallStatus = StatusWarning
        }
    }
    
    // Update state
    m.mu.Lock()
    oldStatus := m.currentStatus[lease.ID]
    m.currentStatus[lease.ID] = status
    m.mu.Unlock()
    
    // Handle status changes
    m.handleStatusChange(ctx, oldStatus, status)
}

func (m *Monitor) evaluateMetric(metric SLAMetric, currentValue decimal.Decimal) *MetricStatus {
    status := &MetricStatus{
        Type:         metric.Type,
        CurrentValue: currentValue,
        Target:       metric.Target,
        LastUpdated:  time.Now(),
    }
    
    var breached bool
    var warning bool
    
    switch metric.Comparison {
    case ComparisonGTE:
        breached = currentValue.LessThan(metric.Target)
        warningThreshold := metric.Target.Add(metric.Target.Mul(m.warningBuffer))
        warning = currentValue.LessThan(warningThreshold)
    case ComparisonLTE:
        breached = currentValue.GreaterThan(metric.Target)
        warningThreshold := metric.Target.Sub(metric.Target.Mul(m.warningBuffer))
        warning = currentValue.GreaterThan(warningThreshold)
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

func (m *Monitor) handleStatusChange(ctx context.Context, old, new *SLAStatus) {
    if old == nil {
        return
    }
    
    // Check for new breaches
    for metricType, newMetric := range new.Metrics {
        oldMetric, exists := old.Metrics[metricType]
        
        if newMetric.Status == StatusBreached && (!exists || oldMetric.Status != StatusBreached) {
            // New breach detected
            m.handleBreach(ctx, new, metricType, newMetric)
        } else if newMetric.Status == StatusWarning && (!exists || oldMetric.Status == StatusHealthy) {
            // Warning state entered
            m.handleWarning(ctx, new, metricType, newMetric)
        } else if newMetric.Status == StatusHealthy && exists && oldMetric.Status == StatusBreached {
            // Breach resolved
            m.handleBreachResolved(ctx, new, metricType)
        }
    }
}

func (m *Monitor) handleBreach(ctx context.Context, status *SLAStatus, metricType SLAMetricType, metric *MetricStatus) {
    // Create breach record
    breach := &BreachRecord{
        ID:          generateID(),
        LeaseID:     status.LeaseID,
        MetricType:  metricType,
        TargetValue: metric.Target,
        ActualValue: metric.CurrentValue,
        StartedAt:   time.Now(),
    }
    
    m.slaStore.CreateBreach(ctx, breach)
    
    // Alert
    m.alerter.SendAlert(ctx, Alert{
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
    
    // Notify customer
    m.alerter.NotifyCustomer(ctx, status.CustomerID, CustomerNotification{
        Type:    "sla_breach",
        Title:   "SLA Breach Detected",
        Message: fmt.Sprintf("An SLA breach has been detected for your service. Credit compensation will be applied automatically."),
        LeaseID: status.LeaseID,
    })
    
    // Trigger remediation
    m.remediator.HandleBreach(ctx, breach)
}

func (m *Monitor) handleWarning(ctx context.Context, status *SLAStatus, metricType SLAMetricType, metric *MetricStatus) {
    // Alert provider
    m.alerter.SendAlert(ctx, Alert{
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
    
    // Trigger proactive remediation if configured
    m.remediator.HandleWarning(ctx, status, metricType)
}

func (m *Monitor) handleBreachResolved(ctx context.Context, status *SLAStatus, metricType SLAMetricType) {
    // Update breach record
    breach, err := m.slaStore.GetActiveBreach(ctx, status.LeaseID, metricType)
    if err != nil || breach == nil {
        return
    }
    
    now := time.Now()
    breach.ResolvedAt = &now
    breach.Duration = now.Sub(breach.StartedAt)
    
    // Calculate credit
    breach.CreditAmount = m.calculateCredit(status.SLADefinition, breach)
    
    m.slaStore.UpdateBreach(ctx, breach)
    
    // Alert
    m.alerter.SendAlert(ctx, Alert{
        Severity: AlertInfo,
        Title:    fmt.Sprintf("SLA Breach Resolved: %s", metricType),
        Message:  fmt.Sprintf("SLA breach resolved after %s. Credit: %s", breach.Duration, breach.CreditAmount),
        Labels: map[string]string{
            "lease_id": status.LeaseID,
            "metric":   string(metricType),
        },
    })
    
    // Apply credit to escrow
    m.remediator.ApplyCredit(ctx, breach)
}
```

### Alerter

```go
// pkg/sla/alerter.go

package sla

import (
    "context"
    "fmt"
)

type AlertSeverity string

const (
    AlertInfo     AlertSeverity = "info"
    AlertWarning  AlertSeverity = "warning"
    AlertCritical AlertSeverity = "critical"
)

type Alert struct {
    Severity AlertSeverity
    Title    string
    Message  string
    Labels   map[string]string
}

type CustomerNotification struct {
    Type    string
    Title   string
    Message string
    LeaseID string
}

type Alerter interface {
    SendAlert(ctx context.Context, alert Alert) error
    NotifyCustomer(ctx context.Context, customerID string, notif CustomerNotification) error
    NotifyProvider(ctx context.Context, providerID string, notif ProviderNotification) error
}

type AlerterImpl struct {
    prometheus    PrometheusClient
    pagerduty     *PagerDutyClient
    notifications NotificationService
}

func (a *AlerterImpl) SendAlert(ctx context.Context, alert Alert) error {
    // Send to Prometheus AlertManager
    if err := a.prometheus.SendAlert(ctx, PrometheusAlert{
        Labels: map[string]string{
            "alertname": alert.Title,
            "severity":  string(alert.Severity),
        },
        Annotations: map[string]string{
            "summary":     alert.Title,
            "description": alert.Message,
        },
    }); err != nil {
        return err
    }
    
    // For critical alerts, also page on-call
    if alert.Severity == AlertCritical && a.pagerduty != nil {
        return a.pagerduty.CreateIncident(ctx, PagerDutyIncident{
            Title:    alert.Title,
            Body:     alert.Message,
            Severity: "critical",
            Service:  "virtengine-sla",
        })
    }
    
    return nil
}

func (a *AlerterImpl) NotifyCustomer(ctx context.Context, customerID string, notif CustomerNotification) error {
    return a.notifications.Send(ctx, Notification{
        UserAddress: customerID,
        Type:        NotificationType(notif.Type),
        Title:       notif.Title,
        Body:        notif.Message,
        Data: map[string]string{
            "lease_id": notif.LeaseID,
        },
        Channels: []Channel{ChannelPush, ChannelEmail, ChannelInApp},
    })
}
```

### Prometheus Alert Rules

```yaml
# deploy/observability/prometheus/rules/sla.yml
groups:
  - name: sla.alerts
    rules:
      - alert: SLAUptimeBreach
        expr: |
          (1 - (
            sum(rate(virtengine_provider_workload_uptime_seconds[30d])) by (provider_id, lease_id)
            / sum(rate(virtengine_provider_workload_expected_uptime_seconds[30d])) by (provider_id, lease_id)
          )) * 100 > 0.1
        for: 5m
        labels:
          severity: critical
          sla_metric: uptime
        annotations:
          summary: "SLA Uptime breach for {{ $labels.lease_id }}"
          description: "Uptime is {{ $value | printf \"%.2f\" }}% below SLA target"
          
      - alert: SLALatencyBreach
        expr: |
          histogram_quantile(0.99, 
            sum(rate(virtengine_provider_request_latency_bucket[1h])) by (le, lease_id)
          ) > 500
        for: 5m
        labels:
          severity: critical
          sla_metric: latency_p99
        annotations:
          summary: "SLA P99 Latency breach for {{ $labels.lease_id }}"
          description: "P99 latency is {{ $value | printf \"%.0f\" }}ms, exceeds 500ms SLA"
          
      - alert: SLAUptimeWarning
        expr: |
          (1 - (
            sum(rate(virtengine_provider_workload_uptime_seconds[30d])) by (provider_id, lease_id)
            / sum(rate(virtengine_provider_workload_expected_uptime_seconds[30d])) by (provider_id, lease_id)
          )) * 100 > 0.05
        for: 5m
        labels:
          severity: warning
          sla_metric: uptime
        annotations:
          summary: "SLA Uptime approaching threshold for {{ $labels.lease_id }}"
          description: "Uptime buffer is {{ $value | printf \"%.2f\" }}%"
```

---

## Directory Structure

```
pkg/sla/
├── types.go              # SLA types
├── monitor.go            # SLA monitoring engine
├── alerter.go            # Alert sending
├── remediator.go         # Remediation actions
├── credit.go             # Credit calculation
└── store/
    └── postgres.go       # SLA storage

portal/src/app/sla/
├── page.tsx              # SLA dashboard
└── [leaseId]/
    └── page.tsx          # Lease SLA details
```

---

## Testing Requirements

### Unit Tests
- SLA metric evaluation
- Credit calculation
- Alert rule logic

### Integration Tests
- Full breach detection flow
- Notification delivery
- Credit application

### Load Tests
- Monitor performance with 10k+ leases
- Alert storm handling

---

## Success Metrics

| Metric | Target |
|--------|--------|
| Breach detection time | < 1 minute |
| Alert delivery time | < 30 seconds |
| False positive rate | < 1% |
| Credit accuracy | 100% |
