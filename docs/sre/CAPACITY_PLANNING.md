# Capacity Planning Framework

## Overview

Capacity planning ensures VirtEngine services have sufficient resources to meet demand while maintaining SLO targets. This document defines the framework for proactive capacity management.

## Table of Contents
1. [Capacity Planning Principles](#capacity-planning-principles)
2. [Resource Metrics](#resource-metrics)
3. [Forecasting Models](#forecasting-models)
4. [Capacity Thresholds](#capacity-thresholds)
5. [Planning Process](#planning-process)
6. [Implementation](#implementation)

---

## Capacity Planning Principles

### 1. Headroom for Growth
- **Target**: Maintain 20-30% headroom above peak usage
- **Rationale**: Absorbs traffic spikes, new deployments, failover scenarios

### 2. N+2 Redundancy
- **Blockchain Nodes**: Support losing 2 validators without SLO impact
- **Provider Nodes**: Support 2 provider failures without capacity constraints
- **API Servers**: Support 2 server failures with degraded performance only

### 3. Organic vs Inorganic Growth
- **Organic**: Gradual user adoption (predictable)
- **Inorganic**: Marketing campaigns, partnerships, viral events (unpredictable)
- **Plan for both**: Baseline + surge capacity

### 4. Lead Time Awareness
- **Cloud Resources**: Minutes to hours
- **Bare Metal**: Weeks to months
- **Plan ahead**: 3-6 month horizon

### 5. Cost Optimization
- **Right-sizing**: Don't over-provision
- **Autoscaling**: Use dynamic scaling where possible
- **Reserved capacity**: Lock in discounts for baseline

---

## Resource Metrics

### 1. Blockchain Node Capacity

#### CPU Metrics
```
# Current usage
node_cpu_usage_percent

# Transaction processing capacity
tx_processing_capacity = CPU_cores × 1000 TPS_per_core

# Headroom
cpu_headroom_percent = (CPU_capacity - CPU_peak_usage) / CPU_capacity × 100

Target: ≥ 30% headroom
```

#### Memory Metrics
```
# State size growth
blockchain_state_size_gb

# Memory usage
node_memory_usage_gb

# Growth rate
memory_growth_rate_gb_per_month = Δ(memory_usage) / Δ(time)

# Projected capacity exhaustion
months_until_exhaustion = (total_memory - current_usage) / growth_rate
```

#### Storage Metrics
```
# Blockchain data size
blockchain_data_size_gb

# Storage usage
disk_usage_percent

# Growth rate
storage_growth_rate_gb_per_day

# Projected exhaustion
days_until_full = (total_disk - current_usage) / growth_rate

Alert: < 90 days until full
```

#### Network Metrics
```
# Bandwidth usage
network_bandwidth_mbps

# Transaction broadcast capacity
broadcast_capacity_tps

# Peer connections
peer_connection_count

Target: < 70% of network bandwidth
```

---

### 2. Provider Daemon Capacity

#### Workload Capacity
```
# Concurrent deployments
active_deployment_count

# Maximum capacity
max_deployment_capacity

# Utilization
deployment_utilization = active_deployments / max_capacity × 100

Target: 50-80% utilization
Warning: > 80%
Critical: > 90%
```

#### Resource Allocation
```
# Per-resource tracking
cpu_allocated_cores
cpu_available_cores

memory_allocated_gb
memory_available_gb

storage_allocated_gb
storage_available_gb

gpu_allocated_units
gpu_available_units

# Headroom
resource_headroom = (available - allocated) / total × 100

Target: ≥ 20% headroom per resource type
```

#### Bid Capacity
```
# Bid processing rate
bids_per_second

# Queue depth
bid_queue_depth

# Processing latency
bid_processing_latency_p95

Target: Bid queue depth < 100
Target: P95 latency < 5s
```

---

### 3. API Service Capacity

#### Request Capacity
```
# Requests per second
api_requests_per_second

# Server capacity
api_capacity_rps = servers × capacity_per_server

# Utilization
api_utilization = current_rps / capacity_rps × 100

Target: 50-70% utilization
Warning: > 80%
```

#### Connection Pool
```
# Database connections
db_connection_pool_active
db_connection_pool_max

# Pool utilization
pool_utilization = active / max × 100

Target: < 80%
```

---

## Forecasting Models

### 1. Linear Regression (Baseline)

For predictable, gradual growth:

```python
# Linear model: y = mx + b
import numpy as np
from sklearn.linear_model import LinearRegression

# Historical data
timestamps = [...] # Unix timestamps
usage = [...] # Resource usage values

# Fit model
model = LinearRegression()
model.fit(timestamps.reshape(-1, 1), usage)

# Forecast 90 days ahead
future_time = current_time + (90 * 24 * 3600)
forecast = model.predict([[future_time]])
```

**Use for**:
- Storage growth
- State size growth
- Baseline traffic growth

---

### 2. Exponential Growth (Adoption Curve)

For user adoption scenarios:

```python
# Exponential model: y = a × e^(bx)
import numpy as np
from scipy.optimize import curve_fit

def exponential(x, a, b):
    return a * np.exp(b * x)

# Fit to historical data
params, _ = curve_fit(exponential, timestamps, usage)

# Forecast
forecast = exponential(future_time, *params)
```

**Use for**:
- User growth
- Transaction volume growth
- Network effects

---

### 3. Seasonal Decomposition (Cyclic Patterns)

For cyclic usage patterns:

```python
from statsmodels.tsa.seasonal import seasonal_decompose

# Decompose time series
result = seasonal_decompose(usage, model='additive', period=7*24) # Weekly cycle

# Components
trend = result.trend
seasonal = result.seasonal
residual = result.resid

# Forecast = trend + seasonal pattern
```

**Use for**:
- API request patterns (daily/weekly cycles)
- Provider workload patterns
- On-chain activity patterns

---

### 4. Machine Learning (Complex Patterns)

For complex, multi-variate forecasting:

```python
from sklearn.ensemble import RandomForestRegressor

# Features
features = [
    'day_of_week',
    'hour_of_day',
    'user_count',
    'deployment_count',
    'tx_volume',
]

# Train model
model = RandomForestRegressor(n_estimators=100)
model.fit(X_train, y_train)

# Predict future capacity needs
forecast = model.predict(X_future)
```

**Use for**:
- Multi-dimensional capacity planning
- Correlation-based forecasting
- Anomaly-aware predictions

---

## Capacity Thresholds

### Alert Levels

| Metric | Healthy | Warning | Critical | Emergency |
|--------|---------|---------|----------|-----------|
| **CPU Usage** | < 70% | 70-85% | 85-95% | > 95% |
| **Memory Usage** | < 70% | 70-85% | 85-95% | > 95% |
| **Disk Usage** | < 70% | 70-85% | 85-95% | > 95% |
| **Deployment Utilization** | 50-80% | 80-90% | 90-95% | > 95% |
| **API Utilization** | < 70% | 70-85% | 85-95% | > 95% |
| **Time to Exhaustion** | > 90 days | 60-90 days | 30-60 days | < 30 days |

### Actionable Thresholds

**Warning (70-85%)**:
- Begin capacity review
- Accelerate forecasting analysis
- Prepare provisioning plan
- No immediate action required

**Critical (85-95%)**:
- Initiate provisioning process
- Implement temporary optimizations
- Daily monitoring
- Stakeholder notification

**Emergency (> 95%)**:
- Immediate provisioning (emergency budget)
- Implement rate limiting
- Scale horizontally immediately
- Incident declared

---

## Planning Process

### Quarterly Capacity Review

**Timeline**: 6 weeks before quarter start

#### Week 1-2: Data Collection
- [ ] Gather historical usage data (past 12 months)
- [ ] Collect growth metrics (users, transactions, deployments)
- [ ] Document known future events (launches, campaigns)
- [ ] Review current capacity and utilization

#### Week 3: Forecasting
- [ ] Run forecasting models (linear, exponential, seasonal)
- [ ] Project growth for next 6 months
- [ ] Calculate capacity requirements
- [ ] Identify resource gaps

#### Week 4: Planning
- [ ] Define provisioning needs (servers, storage, bandwidth)
- [ ] Cost estimation and budget approval
- [ ] Vendor selection (cloud, bare metal, co-location)
- [ ] Create provisioning timeline

#### Week 5: Review & Approval
- [ ] Present plan to engineering leadership
- [ ] Present plan to finance (budget approval)
- [ ] Revise based on feedback
- [ ] Get executive sign-off

#### Week 6: Implementation Prep
- [ ] Purchase orders submitted
- [ ] Provisioning scripts prepared
- [ ] Monitoring configured
- [ ] Runbooks updated

---

### Monthly Capacity Check-In

**Duration**: 1 hour meeting

**Agenda**:
1. Review current utilization (10 min)
2. Compare actuals vs forecast (10 min)
3. Update growth projections (15 min)
4. Identify risks and blockers (15 min)
5. Adjust plan if needed (10 min)

**Attendees**: SRE team, Engineering leads

---

### Weekly Capacity Monitoring

**Duration**: 15 minutes in SRE standup

**Checklist**:
- [ ] Any metrics in critical zone?
- [ ] Forecast still accurate?
- [ ] Any surprise growth?
- [ ] Lead time concerns?

---

## Implementation

### Capacity Tracking System

```go
// pkg/sre/capacity/capacity.go

package capacity

import (
	"context"
	"time"
)

type Tracker struct {
	metrics map[string]*Metric
}

type Metric struct {
	Name           string
	CurrentValue   float64
	Capacity       float64
	Utilization    float64
	GrowthRate     float64 // per day
	ProjectedFull  *time.Time
	WarningThreshold  float64 // 0.7 = 70%
	CriticalThreshold float64 // 0.85 = 85%
}

func (t *Tracker) RecordMetric(ctx context.Context, name string, value float64) {
	// Implementation
}

func (t *Tracker) GetUtilization(name string) float64 {
	metric := t.metrics[name]
	return metric.CurrentValue / metric.Capacity
}

func (t *Tracker) ProjectExhaustion(name string) *time.Time {
	metric := t.metrics[name]

	if metric.GrowthRate <= 0 {
		return nil // Not growing
	}

	remainingCapacity := metric.Capacity - metric.CurrentValue
	daysUntilFull := remainingCapacity / metric.GrowthRate

	exhaustionDate := time.Now().Add(time.Duration(daysUntilFull * 24) * time.Hour)
	return &exhaustionDate
}
```

---

### Forecasting Script

```python
#!/usr/bin/env python3
# scripts/capacity-forecast.py

import pandas as pd
import numpy as np
from sklearn.linear_model import LinearRegression
import matplotlib.pyplot as plt
from datetime import datetime, timedelta

def load_metrics(metric_name, days=90):
    """Load historical metrics from Prometheus"""
    # Query Prometheus API
    # Return DataFrame with timestamp and value
    pass

def forecast_linear(df, days_ahead=90):
    """Linear regression forecast"""
    X = np.arange(len(df)).reshape(-1, 1)
    y = df['value'].values

    model = LinearRegression()
    model.fit(X, y)

    # Forecast
    future_X = np.arange(len(df), len(df) + days_ahead).reshape(-1, 1)
    forecast = model.predict(future_X)

    return forecast

def calculate_exhaustion(df, forecast, capacity):
    """Calculate when capacity will be exhausted"""
    all_values = np.concatenate([df['value'].values, forecast])

    for i, value in enumerate(all_values):
        if value >= capacity:
            days_until_full = i - len(df)
            return days_until_full

    return None  # Won't exhaust in forecast period

def main():
    # Metrics to forecast
    metrics = [
        ('disk_usage_gb', 1000),  # (metric_name, capacity)
        ('memory_usage_gb', 64),
        ('deployment_count', 100),
    ]

    for metric_name, capacity in metrics:
        df = load_metrics(metric_name)
        forecast = forecast_linear(df, days_ahead=90)
        days_until_full = calculate_exhaustion(df, forecast, capacity)

        print(f"\n{metric_name}:")
        print(f"  Current: {df['value'].iloc[-1]:.2f}")
        print(f"  Capacity: {capacity}")
        print(f"  Utilization: {df['value'].iloc[-1] / capacity * 100:.1f}%")

        if days_until_full:
            print(f"  ⚠️  Days until full: {days_until_full}")
        else:
            print(f"  ✅ Won't exhaust in 90 days")

if __name__ == '__main__':
    main()
```

---

### Capacity Dashboard

**Grafana Dashboard**: `capacity-planning.json`

**Panels**:
1. **Resource Utilization Gauges**
   - CPU, Memory, Disk, Network
   - Color-coded: Green (< 70%), Yellow (70-85%), Red (> 85%)

2. **Growth Trends**
   - 30-day, 90-day, 12-month trends
   - Linear regression trendlines

3. **Time to Exhaustion**
   - Days until each resource is full
   - Sorted by urgency

4. **Forecast vs Actual**
   - Compare previous forecasts to actuals
   - Forecast accuracy tracking

5. **Headroom by Service**
   - Per-service capacity headroom
   - Identify bottlenecks

---

### Alerting Rules

```yaml
# alerting/capacity-alerts.yml

groups:
  - name: capacity_planning
    interval: 5m
    rules:
      # Disk space running out
      - alert: DiskSpaceWarning
        expr: (disk_usage_gb / disk_capacity_gb) > 0.70
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "Disk space at {{ $value | humanizePercentage }}"
          description: "Disk will be full in {{ $labels.days_until_full }} days"

      - alert: DiskSpaceCritical
        expr: (disk_usage_gb / disk_capacity_gb) > 0.85
        for: 30m
        labels:
          severity: critical
        annotations:
          summary: "Disk space critically low"
          description: "Immediate action required"

      # Memory exhaustion
      - alert: MemoryCapacityWarning
        expr: (memory_usage_gb / memory_capacity_gb) > 0.70
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "Memory usage at {{ $value | humanizePercentage }}"

      # Deployment capacity
      - alert: DeploymentCapacityWarning
        expr: (active_deployments / max_deployments) > 0.80
        for: 30m
        labels:
          severity: warning
        annotations:
          summary: "Deployment capacity at {{ $value | humanizePercentage }}"

      # Forecast-based alerts
      - alert: CapacityExhaustionImminent
        expr: capacity_days_until_full < 30
        for: 1h
        labels:
          severity: critical
        annotations:
          summary: "Capacity will be exhausted in {{ $value }} days"
          description: "Resource {{ $labels.resource }} needs provisioning"
```

---

## Capacity Planning Scenarios

### Scenario 1: Mainnet Launch

**Event**: Public mainnet launch
**Expected Impact**: 10x transaction volume, 5x new users

**Planning**:
1. **Baseline Capacity** (current):
   - 3 validator nodes (100 TPS each = 300 TPS total)
   - 10 provider nodes (10 concurrent deployments each)
   - 2 API servers (1000 RPS each)

2. **Projected Need** (10x tx volume):
   - 30 validator nodes (3000 TPS) → scale from 3 to 30
   - 50 provider nodes → scale from 10 to 50
   - 10 API servers (10,000 RPS) → scale from 2 to 10

3. **Provisioning Timeline**:
   - Week -4: Order hardware/reserve cloud capacity
   - Week -2: Configure and test nodes
   - Week -1: Deploy to staging, load test
   - Week 0: Go live with full capacity

4. **Fallback Plan**:
   - Rate limiting if capacity exceeded
   - Queue system for deployments
   - CDN for static API responses

---

### Scenario 2: Viral Growth Event

**Event**: Unexpected 100x traffic spike (social media, news)
**Timeline**: < 24 hours notice

**Response**:
1. **Hour 0**: Detect spike in monitoring
2. **Hour 1**: Activate emergency capacity
   - Auto-scale cloud resources
   - Enable rate limiting
   - Redirect to CDN
3. **Hour 2**: Provision additional servers
4. **Hour 6**: Load test new capacity
5. **Hour 12**: Monitor and adjust
6. **Day 2**: Review and optimize

**Requirements**:
- Cloud infrastructure for rapid scaling
- Rate limiting configuration ready
- CDN integration configured
- On-call team available 24/7

---

### Scenario 3: Gradual Organic Growth

**Event**: Steady 10% month-over-month growth
**Timeline**: Ongoing

**Planning**:
1. **Monthly Review**:
   - Track growth rate
   - Update forecasts
   - Adjust thresholds

2. **Quarterly Provisioning**:
   - Add capacity every quarter
   - Stay ahead of curve
   - Maintain 20-30% headroom

3. **Annual Architecture Review**:
   - Optimize for scale
   - Evaluate cost efficiency
   - Plan major upgrades

---

## Cost Optimization

### Right-Sizing Strategy

**Principle**: Match resources to actual needs

**Process**:
1. Analyze historical usage patterns
2. Identify over-provisioned resources
3. Downsize where safe (maintain headroom)
4. Reinvest savings in automation

**Example**:
- API server running at 20% CPU → downsize instance type
- Database with 500GB allocated, 200GB used → reduce storage
- Savings: $5000/month → invest in monitoring tools

---

### Autoscaling Policy

**Kubernetes Horizontal Pod Autoscaling**:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-server-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-server
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Percent
          value: 100
          periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
```

---

## References

- [Google SRE Book - Demand Forecasting](https://sre.google/sre-book/handling-overload/)
- [SLI/SLO/SLA Definitions](SLI_SLO_SLA.md)
- [Performance Budgets](PERFORMANCE_BUDGETS.md)
- [Incident Response for Capacity Issues](INCIDENT_RESPONSE.md)

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-29
**Version**: 1.0.0
**Next Review**: 2026-04-29
