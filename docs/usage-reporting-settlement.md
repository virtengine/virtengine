# Usage Reporting and Settlement Integration

## Overview

VirtEngine's usage reporting system connects marketplace allocations to settlement and invoicing through a comprehensive pipeline that includes:

1. **Scheduled Usage Collection** - Periodic metrics collection from workloads
2. **On-Chain Submission** - Signed usage reports submitted to the blockchain
3. **Settlement Pipeline** - Conversion to billable line items
4. **Dispute Resolution** - Time-windowed dispute and correction workflow
5. **Waldur Reconciliation** - Cross-validation with Waldur platform metrics
6. **Anomaly Detection** - Automated fraud and error detection

## Architecture

```
┌─────────────────────┐
│   Usage Meter       │
│   (per workload)    │
└─────────┬───────────┘
          │ ResourceMetrics
          ▼
┌─────────────────────┐
│ Scheduled Collector │
│   (hourly cron)     │
└─────────┬───────────┘
          │ UsageRecord
          ▼
┌─────────────────────┐     ┌──────────────────┐
│ Settlement Pipeline │────▶│ Anomaly Detector │
│                     │     └──────────────────┘
└─────────┬───────────┘              │
          │                          ▼
          │              ┌──────────────────┐
          │              │  Alert Manager   │
          │              └──────────────────┘
          ▼
┌─────────────────────┐
│  Chain Submitter    │
│  (MsgRecordUsage)   │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐     ┌──────────────────┐
│  x/settlement       │◀────│ Dispute Window   │
│  (on-chain)         │     │ (24 hours)       │
└─────────┬───────────┘     └──────────────────┘
          │
          ▼
┌─────────────────────┐
│  Escrow Settlement  │
│  (fund transfer)    │
└─────────────────────┘
```

## Configuration

### Settlement Pipeline Configuration

```go
type SettlementConfig struct {
    // SettlementInterval is the interval for periodic settlements
    SettlementInterval time.Duration // default: 1 hour
    
    // DisputeWindow is the time window for disputes after usage is reported
    DisputeWindow time.Duration // default: 24 hours
    
    // ReconciliationInterval is the interval for Waldur reconciliation
    ReconciliationInterval time.Duration // default: 6 hours
    
    // MaxPendingRecords triggers settlement when exceeded
    MaxPendingRecords int // default: 100
    
    // RetryAttempts for failed submissions
    RetryAttempts int // default: 3
}
```

### Provider Daemon Configuration

Add to `provider-daemon.yaml`:

```yaml
settlement:
  enabled: true
  settlement_interval: "1h"
  dispute_window: "24h"
  reconciliation_interval: "6h"
  max_pending_records: 100
  
chain_submitter:
  enabled: true
  gas_limit: 200000
  gas_price: "0.025uvirt"
  batch_size: 10
  batch_interval: "1m"
  
waldur_reconciler:
  enabled: true
  discrepancy_threshold: 10.0  # percentage
  auto_correct: false
  auto_correct_threshold: 5.0

alerts:
  enabled: true
  default_ttl: "24h"
  max_alerts: 10000
```

## Usage Reporting Schedule

### Hourly Collection (Default)

| Time | Action |
|------|--------|
| XX:00 | Collect metrics for all active workloads |
| XX:01 | Process metrics to usage records |
| XX:02 | Detect anomalies |
| XX:03 | Submit to chain (batched) |
| XX:05 | Update settlement pipeline |

### Daily Settlement Cycle

| Time | Action |
|------|--------|
| 00:00 UTC | Begin daily settlement cycle |
| 00:15 UTC | Process unsettled usage records |
| 00:30 UTC | Generate billable line items |
| 01:00 UTC | Submit settlement transactions |
| 02:00 UTC | Waldur reconciliation |

## Dispute Workflow

### Creating a Dispute

```bash
virtengine tx settlement dispute-usage \
  --usage-id usage-1234567890 \
  --reason "CPU usage appears inflated" \
  --expected-cpu-hours 10 \
  --from customer
```

### Dispute Lifecycle

1. **Created** - Dispute is created within the dispute window
2. **Pending** - Awaiting review
3. **Reviewing** - Under active investigation
4. **Resolved** - Dispute accepted, correction applied
5. **Rejected** - Dispute denied
6. **Expired** - Dispute window passed

### Dispute Window

- Default: **24 hours** after usage record submission
- Usage cannot be settled while disputed
- Corrections generate new usage records with reference to original

## Reconciliation Process

### Provider vs Waldur Metrics

The reconciler compares:

| Metric | Provider Source | Waldur Source |
|--------|-----------------|---------------|
| CPU | Container metrics | Component usage |
| Memory | Container metrics | Component usage |
| Storage | Volume stats | Resource limits |
| GPU | NVIDIA metrics | Component usage |
| Network | CNI metrics | Traffic stats |

### Discrepancy Handling

| Difference | Severity | Action |
|------------|----------|--------|
| 0-5% | Low | Log only |
| 5-10% | Medium | Alert generated |
| 10-25% | High | Alert + review |
| >25% | Critical | Settlement blocked |

## Anomaly Detection

### Detected Anomaly Types

| Type | Description | Severity |
|------|-------------|----------|
| `duration_too_short` | Record < 1 minute | Medium |
| `duration_too_long` | Record > 25 hours | High |
| `negative_values` | Negative metric values | Critical |
| `cpu_variance` | CPU usage variance > 50% | Variable |
| `memory_variance` | Memory usage variance > 50% | Variable |
| `zero_duration_with_usage` | Duration 0 but usage > 0 | Critical |

### Fraud Detection

The fraud checker evaluates:

- Timestamp validity (no future timestamps)
- Duration bounds
- Resource ratio limits
- Signature verification

## Metrics and Monitoring

### Key Metrics

```
# Counter metrics
provider_daemon_usage_records_collected_total
provider_daemon_usage_records_submitted_total
provider_daemon_submission_failures_total
provider_daemon_settlements_processed_total
provider_daemon_disputes_created_total
provider_daemon_anomalies_detected_total

# Gauge metrics
provider_daemon_pending_records
provider_daemon_active_disputes
provider_daemon_reconciliation_score
```

### Prometheus Integration

```yaml
# prometheus.yml scrape config
- job_name: 'provider-daemon'
  static_configs:
    - targets: ['localhost:9090']
  metrics_path: '/metrics'
```

### Alert Rules

```yaml
# alertmanager rules
groups:
  - name: usage-reporting
    rules:
      - alert: HighDisputeRate
        expr: rate(provider_daemon_disputes_created_total[1h]) > 0.1
        for: 5m
        labels:
          severity: warning
          
      - alert: ReconciliationFailure
        expr: provider_daemon_reconciliation_score < 50
        for: 10m
        labels:
          severity: critical
          
      - alert: SubmissionBacklog
        expr: provider_daemon_pending_records > 500
        for: 5m
        labels:
          severity: warning
```

## Operational Procedures

### Starting the Settlement Pipeline

```bash
# Start provider daemon with settlement enabled
provider-daemon start \
  --settlement.enabled=true \
  --chain.rpc=http://localhost:26657
```

### Manual Settlement Trigger

```bash
# Force settlement for an order
virtengine tx settlement settle-order \
  --order-id order-1234567890 \
  --from provider
```

### Checking Reconciliation Status

```bash
# Get reconciliation status
virtengine query settlement reconciliation-status \
  --allocation-id alloc-1234567890
```

### Viewing Active Disputes

```bash
# List active disputes
virtengine query settlement disputes \
  --status pending \
  --provider provider-address
```

## Error Recovery

### Submission Failures

1. Failed submissions are automatically retried (3 attempts)
2. After max retries, records are re-queued
3. Alert generated for persistent failures

### Settlement Conflicts

1. Disputed usage blocks settlement
2. After dispute resolution, settlement resumes
3. Corrections create adjustment records

### Reconciliation Errors

1. Waldur unavailable: Continue with provider data only
2. Large discrepancy: Block settlement, generate critical alert
3. Auto-correct: Only for discrepancies < 5%

## API Reference

### Submit Usage Report

```protobuf
message MsgRecordUsage {
  string sender = 1;
  string order_id = 2;
  string lease_id = 3;
  uint64 usage_units = 4;
  string usage_type = 5;
  int64 period_start = 6;
  int64 period_end = 7;
  cosmos.base.v1beta1.DecCoin unit_price = 8;
  bytes signature = 9;
}
```

### Settle Order

```protobuf
message MsgSettleOrder {
  string sender = 1;
  string order_id = 2;
  repeated string usage_record_ids = 3;
  bool is_final = 4;
}
```

### Query Usage Records

```bash
virtengine query settlement usage-records \
  --order-id order-123 \
  --settled=false
```

## Security Considerations

1. **Signature Verification** - All usage reports must be signed by the provider
2. **Rate Limiting** - Submission rate limited to prevent spam
3. **Dispute Authentication** - Only customer or provider can create disputes
4. **Correction Audit** - All corrections are logged with signatures

## Troubleshooting

### Common Issues

| Issue | Cause | Resolution |
|-------|-------|------------|
| Settlement stuck | Active dispute | Resolve dispute first |
| High anomaly rate | Misconfigured thresholds | Adjust threshold values |
| Reconciliation failing | Waldur connectivity | Check Waldur endpoint |
| Submission timeout | Chain congestion | Increase timeout/retry |

### Debug Logging

```bash
# Enable debug logging
provider-daemon start --log-level=debug

# Check settlement logs
grep "settlement-pipeline" /var/log/provider-daemon.log
```
