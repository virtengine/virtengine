# Task 31J: Production Metrics Dashboards

**vibe-kanban ID:** `f91b5c13-22df-4dc3-87b9-b8ec16f63cd9`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31J |
| **Title** | feat(observability): Production metrics dashboards |
| **Priority** | P1 |
| **Wave** | 2 |
| **Estimated LOC** | 2000 |
| **Duration** | 2-3 weeks |
| **Dependencies** | 31B (Distributed Tracing), 31I (Load Testing) |
| **Blocking** | None |

---

## Problem Statement

Production operations require comprehensive metrics dashboards for:
- Real-time system health monitoring
- Capacity planning and trend analysis
- Incident detection and response
- SLO tracking and reporting
- Business metrics visibility

While Prometheus metrics may exist, no dashboards are configured for production use.

### Current State Analysis

```
docker-compose.observability.yaml  ✅ Basic Prometheus/Grafana
deploy/observability/              ⚠️  Incomplete configuration
grafana/dashboards/                ❌ No provisioned dashboards
Alerting rules:                    ❌ None defined
```

---

## Acceptance Criteria

### AC-1: Infrastructure Dashboards
- [ ] Kubernetes cluster overview
- [ ] Node health and resource utilization
- [ ] Pod status and restarts
- [ ] Network I/O and latency
- [ ] Storage utilization

### AC-2: Application Dashboards
- [ ] VirtEngine node metrics (block height, sync lag)
- [ ] Transaction throughput and latency
- [ ] Module-specific metrics (VEID, Market, Escrow)
- [ ] Provider daemon metrics
- [ ] ML inference metrics

### AC-3: Business Dashboards
- [ ] Active users and VEID verifications
- [ ] Order volume and value
- [ ] Provider utilization rates
- [ ] Revenue metrics (escrow settlements)
- [ ] SLA compliance metrics

### AC-4: Alerting Configuration
- [ ] Critical alerts (service down, consensus failure)
- [ ] Warning alerts (high latency, resource pressure)
- [ ] Business alerts (SLA breach risk)
- [ ] PagerDuty/Opsgenie integration
- [ ] Alert runbook links

---

## Technical Requirements

### Grafana Dashboard - VirtEngine Overview

```json
// deploy/observability/grafana/dashboards/virtengine-overview.json
{
  "annotations": {
    "list": []
  },
  "editable": true,
  "fiscalYearStartMonth": 0,
  "graphTooltip": 1,
  "id": null,
  "links": [],
  "liveNow": false,
  "panels": [
    {
      "title": "Network Status",
      "type": "stat",
      "gridPos": { "h": 4, "w": 6, "x": 0, "y": 0 },
      "targets": [
        {
          "expr": "cometbft_consensus_height",
          "legendFormat": "Block Height"
        }
      ],
      "options": {
        "colorMode": "value",
        "graphMode": "area",
        "textMode": "auto"
      },
      "fieldConfig": {
        "defaults": {
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "green", "value": null }
            ]
          },
          "unit": "none"
        }
      }
    },
    {
      "title": "Active Validators",
      "type": "gauge",
      "gridPos": { "h": 4, "w": 6, "x": 6, "y": 0 },
      "targets": [
        {
          "expr": "cometbft_consensus_validators_power / cometbft_consensus_total_voting_power * 100"
        }
      ],
      "options": {
        "orientation": "auto",
        "showThresholdLabels": false,
        "showThresholdMarkers": true
      },
      "fieldConfig": {
        "defaults": {
          "min": 0,
          "max": 100,
          "thresholds": {
            "mode": "absolute",
            "steps": [
              { "color": "red", "value": null },
              { "color": "yellow", "value": 60 },
              { "color": "green", "value": 80 }
            ]
          },
          "unit": "percent"
        }
      }
    },
    {
      "title": "Transaction Throughput",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 12, "x": 0, "y": 4 },
      "targets": [
        {
          "expr": "rate(cometbft_consensus_total_txs[5m])",
          "legendFormat": "TPS"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "custom": {
            "drawStyle": "line",
            "lineInterpolation": "smooth",
            "fillOpacity": 20
          },
          "unit": "ops"
        }
      }
    },
    {
      "title": "Block Time",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 12, "x": 12, "y": 4 },
      "targets": [
        {
          "expr": "histogram_quantile(0.99, rate(cometbft_consensus_block_interval_seconds_bucket[5m]))",
          "legendFormat": "P99"
        },
        {
          "expr": "histogram_quantile(0.50, rate(cometbft_consensus_block_interval_seconds_bucket[5m]))",
          "legendFormat": "Median"
        }
      ]
    },
    {
      "title": "VEID Submissions",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 8, "x": 0, "y": 12 },
      "targets": [
        {
          "expr": "rate(virtengine_veid_submissions_total[1h])",
          "legendFormat": "Submissions/hr"
        },
        {
          "expr": "rate(virtengine_veid_verifications_total[1h])",
          "legendFormat": "Verifications/hr"
        }
      ]
    },
    {
      "title": "Market Orders",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 8, "x": 8, "y": 12 },
      "targets": [
        {
          "expr": "sum(virtengine_market_orders_active)",
          "legendFormat": "Active Orders"
        },
        {
          "expr": "sum(virtengine_market_leases_active)",
          "legendFormat": "Active Leases"
        }
      ]
    },
    {
      "title": "Escrow Balance",
      "type": "stat",
      "gridPos": { "h": 8, "w": 8, "x": 16, "y": 12 },
      "targets": [
        {
          "expr": "sum(virtengine_escrow_balance_total)",
          "legendFormat": "Total Escrow"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "currencyUSD"
        }
      }
    }
  ],
  "refresh": "30s",
  "schemaVersion": 38,
  "style": "dark",
  "tags": ["virtengine", "overview"],
  "templating": { "list": [] },
  "time": { "from": "now-6h", "to": "now" },
  "timepicker": {},
  "timezone": "browser",
  "title": "VirtEngine Overview",
  "uid": "virtengine-overview",
  "version": 1,
  "weekStart": ""
}
```

### Provider Daemon Dashboard

```json
// deploy/observability/grafana/dashboards/provider-daemon.json
{
  "title": "Provider Daemon",
  "uid": "provider-daemon",
  "tags": ["virtengine", "provider"],
  "panels": [
    {
      "title": "Active Providers",
      "type": "stat",
      "gridPos": { "h": 4, "w": 6, "x": 0, "y": 0 },
      "targets": [
        {
          "expr": "sum(virtengine_provider_status{status=\"active\"})"
        }
      ]
    },
    {
      "title": "Bid Success Rate",
      "type": "gauge",
      "gridPos": { "h": 4, "w": 6, "x": 6, "y": 0 },
      "targets": [
        {
          "expr": "sum(rate(virtengine_provider_bids_won[1h])) / sum(rate(virtengine_provider_bids_total[1h])) * 100"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percent",
          "min": 0,
          "max": 100
        }
      }
    },
    {
      "title": "Workload Status",
      "type": "piechart",
      "gridPos": { "h": 8, "w": 6, "x": 12, "y": 0 },
      "targets": [
        {
          "expr": "sum by (state) (virtengine_provider_workloads)",
          "legendFormat": "{{state}}"
        }
      ]
    },
    {
      "title": "Resource Utilization - CPU",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 12, "x": 0, "y": 4 },
      "targets": [
        {
          "expr": "avg(virtengine_provider_cpu_utilization)",
          "legendFormat": "Average"
        },
        {
          "expr": "max(virtengine_provider_cpu_utilization)",
          "legendFormat": "Peak"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percentunit",
          "max": 1
        }
      }
    },
    {
      "title": "Resource Utilization - Memory",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 12, "x": 12, "y": 4 },
      "targets": [
        {
          "expr": "avg(virtengine_provider_memory_utilization)",
          "legendFormat": "Average"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "percentunit"
        }
      }
    },
    {
      "title": "Lease Operations Latency",
      "type": "heatmap",
      "gridPos": { "h": 8, "w": 24, "x": 0, "y": 12 },
      "targets": [
        {
          "expr": "sum(increase(virtengine_provider_lease_operation_seconds_bucket[5m])) by (le, operation)",
          "format": "heatmap"
        }
      ]
    }
  ]
}
```

### Alerting Rules

```yaml
# deploy/observability/prometheus/rules/virtengine.yml
groups:
  - name: virtengine.critical
    rules:
      - alert: ConsensusStalled
        expr: increase(cometbft_consensus_height[5m]) == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Consensus has stalled - no new blocks in 5 minutes"
          runbook_url: https://docs.virtengine.io/runbooks/consensus-stalled
          
      - alert: ValidatorDown
        expr: count(cometbft_consensus_validator_ready == 1) < (count(cometbft_consensus_validator_ready) * 0.67)
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "More than 1/3 of validators are down"
          
      - alert: EscrowBalanceLow
        expr: virtengine_escrow_balance_total < 1000000000  # 1000 tokens in minor units
        for: 10m
        labels:
          severity: critical
        annotations:
          summary: "Escrow balance critically low"

  - name: virtengine.warning
    rules:
      - alert: HighTransactionLatency
        expr: histogram_quantile(0.99, rate(virtengine_tx_latency_seconds_bucket[5m])) > 5
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "P99 transaction latency exceeds 5 seconds"
          
      - alert: VEIDVerificationQueue
        expr: virtengine_veid_pending_verifications > 100
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "VEID verification queue backlog"
          
      - alert: ProviderHighFailureRate
        expr: rate(virtengine_provider_workload_failures[1h]) / rate(virtengine_provider_workload_starts[1h]) > 0.1
        for: 15m
        labels:
          severity: warning
        annotations:
          summary: "Provider experiencing high workload failure rate (>10%)"
          
      - alert: MLInferenceTimeout
        expr: rate(virtengine_ml_inference_timeout_total[5m]) > 0.01
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "ML inference timeouts detected"

  - name: virtengine.slo
    rules:
      - alert: SLOVEIDResponseTime
        expr: |
          (
            sum(rate(virtengine_veid_submission_seconds_bucket{le="10"}[1h]))
            / sum(rate(virtengine_veid_submission_seconds_count[1h]))
          ) < 0.99
        for: 1h
        labels:
          severity: warning
          slo: veid_response_time
        annotations:
          summary: "VEID response time SLO at risk (<99% within 10s)"
          
      - alert: SLOOrderMatchingSuccess
        expr: |
          (
            sum(rate(virtengine_market_orders_matched[24h]))
            / sum(rate(virtengine_market_orders_created[24h]))
          ) < 0.95
        for: 2h
        labels:
          severity: warning
          slo: order_matching
        annotations:
          summary: "Order matching success rate below 95%"
```

### Dashboard Provisioning

```yaml
# deploy/observability/grafana/provisioning/dashboards/dashboards.yml
apiVersion: 1

providers:
  - name: 'VirtEngine'
    orgId: 1
    folder: 'VirtEngine'
    folderUid: 'virtengine'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 30
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards/virtengine

---
# deploy/observability/grafana/provisioning/datasources/datasources.yml
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false

  - name: Jaeger
    type: jaeger
    access: proxy
    url: http://jaeger:16686
    editable: false

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    editable: false
```

### Business Metrics Dashboard

```json
// deploy/observability/grafana/dashboards/business-metrics.json
{
  "title": "Business Metrics",
  "uid": "business-metrics",
  "tags": ["virtengine", "business"],
  "panels": [
    {
      "title": "Daily Active Users",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 12, "x": 0, "y": 0 },
      "targets": [
        {
          "expr": "count(increase(virtengine_user_transactions_total[1d]) > 0)",
          "legendFormat": "DAU"
        }
      ]
    },
    {
      "title": "VEID Verifications by Tier",
      "type": "piechart",
      "gridPos": { "h": 8, "w": 6, "x": 12, "y": 0 },
      "targets": [
        {
          "expr": "sum by (tier) (virtengine_veid_verified_total)",
          "legendFormat": "Tier {{tier}}"
        }
      ]
    },
    {
      "title": "New Verifications (24h)",
      "type": "stat",
      "gridPos": { "h": 8, "w": 6, "x": 18, "y": 0 },
      "targets": [
        {
          "expr": "increase(virtengine_veid_verified_total[24h])"
        }
      ]
    },
    {
      "title": "Order Volume",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 12, "x": 0, "y": 8 },
      "targets": [
        {
          "expr": "sum(increase(virtengine_market_orders_created[1d]))",
          "legendFormat": "Orders Created"
        },
        {
          "expr": "sum(increase(virtengine_market_orders_matched[1d]))",
          "legendFormat": "Orders Matched"
        }
      ]
    },
    {
      "title": "Total Value Locked",
      "type": "stat",
      "gridPos": { "h": 8, "w": 12, "x": 12, "y": 8 },
      "targets": [
        {
          "expr": "sum(virtengine_escrow_balance_total)",
          "legendFormat": "TVL"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "currencyUSD",
          "decimals": 2
        }
      }
    },
    {
      "title": "Revenue (Settlements)",
      "type": "timeseries",
      "gridPos": { "h": 8, "w": 24, "x": 0, "y": 16 },
      "targets": [
        {
          "expr": "sum(increase(virtengine_escrow_settlements_total[1d]))",
          "legendFormat": "Daily Settlements"
        }
      ],
      "fieldConfig": {
        "defaults": {
          "unit": "currencyUSD"
        }
      }
    }
  ]
}
```

---

## Directory Structure

```
deploy/observability/
├── grafana/
│   ├── provisioning/
│   │   ├── dashboards/
│   │   │   └── dashboards.yml
│   │   ├── datasources/
│   │   │   └── datasources.yml
│   │   └── alerting/
│   │       └── alerting.yml
│   └── dashboards/
│       ├── virtengine-overview.json
│       ├── provider-daemon.json
│       ├── veid-module.json
│       ├── market-module.json
│       ├── ml-inference.json
│       └── business-metrics.json
├── prometheus/
│   ├── prometheus.yml
│   └── rules/
│       ├── virtengine.yml
│       └── infrastructure.yml
├── alertmanager/
│   └── alertmanager.yml
└── docker-compose.yaml
```

---

## Testing Requirements

### Unit Tests
- Dashboard JSON validation
- Alert rule syntax validation

### Integration Tests
- Prometheus scrape targets reachable
- Grafana API access
- Alert firing verification

### Visual Validation
- Manual review of all dashboards
- Mobile responsiveness check
- Dark/light mode verification

---

## Success Metrics

| Metric | Target |
|--------|--------|
| Dashboard coverage | All critical metrics visualized |
| Alert coverage | All SLOs have alerts |
| MTTR improvement | < 15 min for P1 incidents |
| False positive rate | < 5% of alerts |
