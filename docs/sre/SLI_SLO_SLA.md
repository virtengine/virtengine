# VirtEngine SLI/SLO/SLA Framework

## Table of Contents
1. [Overview](#overview)
2. [Service Level Indicators (SLIs)](#service-level-indicators-slis)
3. [Service Level Objectives (SLOs)](#service-level-objectives-slos)
4. [Service Level Agreements (SLAs)](#service-level-agreements-slas)
5. [Error Budget Policy](#error-budget-policy)
6. [Measurement and Monitoring](#measurement-and-monitoring)

---

## Overview

This document defines the Service Level Indicators (SLIs), Service Level Objectives (SLOs), and Service Level Agreements (SLAs) for all VirtEngine services. These metrics form the foundation of our Site Reliability Engineering (SRE) practices.

### Service Categories

1. **Blockchain Node Services** - Core consensus and transaction processing
2. **Provider Daemon Services** - Resource provisioning and workload management
3. **Benchmark Services** - Performance measurement and reliability scoring
4. **API Services** - gRPC and REST endpoints

### Key Principles

- **User-Centric**: SLIs measure what users experience
- **Achievable**: SLOs are ambitious but realistic
- **Measurable**: All metrics are automatically collected
- **Actionable**: SLO violations trigger clear responses
- **Error Budget Driven**: Balances reliability with innovation velocity

---

## Service Level Indicators (SLIs)

### 1. VirtEngine Blockchain Node

#### 1.1 Availability SLIs

**SLI-NODE-001: Node Uptime**
- **Definition**: Percentage of time the node process is running and responsive
- **Measurement**: Health check endpoint returns 200 OK
- **Collection**: Every 30 seconds
- **Formula**: `(successful_checks / total_checks) * 100`

**SLI-NODE-002: Consensus Participation**
- **Definition**: Percentage of blocks where validator successfully participates
- **Measurement**: Signed blocks / expected blocks
- **Collection**: Per block (via Tendermint metrics)
- **Formula**: `(blocks_signed / blocks_expected) * 100`

**SLI-NODE-003: Transaction Processing Availability**
- **Definition**: Percentage of time node accepts and processes transactions
- **Measurement**: Successful transaction submissions
- **Collection**: Per transaction attempt
- **Formula**: `(successful_tx / total_tx_attempts) * 100`

#### 1.2 Latency SLIs

**SLI-NODE-004: Block Production Latency**
- **Definition**: Time to produce and commit blocks
- **Measurement**: Block commit timestamp delta
- **Collection**: Per block
- **Percentiles**: P50, P95, P99

**SLI-NODE-005: Transaction Confirmation Latency**
- **Definition**: Time from transaction submission to inclusion in block
- **Measurement**: Block inclusion time - submission time
- **Collection**: Per transaction
- **Percentiles**: P50, P95, P99

**SLI-NODE-006: Query Response Latency**
- **Definition**: gRPC/REST API query response time
- **Measurement**: Request duration
- **Collection**: Per query
- **Percentiles**: P50, P95, P99

#### 1.3 Quality SLIs

**SLI-NODE-007: Transaction Success Rate**
- **Definition**: Percentage of transactions that execute successfully
- **Measurement**: Successful executions / total transactions
- **Collection**: Per transaction
- **Formula**: `(successful_tx / total_tx) * 100`
- **Exclusions**: User errors (insufficient funds, invalid signature)

**SLI-NODE-008: State Sync Success Rate**
- **Definition**: Percentage of successful state synchronizations
- **Measurement**: Successful syncs / sync attempts
- **Collection**: Per sync operation
- **Formula**: `(successful_syncs / total_sync_attempts) * 100`

#### 1.4 Throughput SLIs

**SLI-NODE-009: Transaction Throughput**
- **Definition**: Number of transactions processed per second
- **Measurement**: Transactions per block / block time
- **Collection**: Per block
- **Target**: Sustainable TPS under load

---

### 2. Provider Daemon Services

#### 2.1 Availability SLIs

**SLI-PROV-001: Provider Daemon Uptime**
- **Definition**: Percentage of time provider daemon is operational
- **Measurement**: Health check endpoint success
- **Collection**: Every 30 seconds
- **Formula**: `(successful_checks / total_checks) * 100`

**SLI-PROV-002: Bid Engine Availability**
- **Definition**: Percentage of time bid engine processes orders
- **Measurement**: Successful bid placements / opportunities
- **Collection**: Per order event
- **Formula**: `(bids_placed / bid_opportunities) * 100`

**SLI-PROV-003: Manifest Service Availability**
- **Definition**: Percentage of time manifest parsing is available
- **Measurement**: Successful manifest validations
- **Collection**: Per manifest submission
- **Formula**: `(successful_validations / total_validations) * 100`

#### 2.2 Latency SLIs

**SLI-PROV-004: Deployment Provisioning Latency**
- **Definition**: Time from bid acceptance to workload running
- **Measurement**: Workload start time - bid acceptance time
- **Collection**: Per deployment
- **Percentiles**: P50, P95, P99
- **Target**: < 300 seconds (5 minutes)

**SLI-PROV-005: Bid Placement Latency**
- **Definition**: Time to analyze order and place bid
- **Measurement**: Bid submission time - order event time
- **Collection**: Per bid
- **Percentiles**: P50, P95, P99

**SLI-PROV-006: Usage Metering Latency**
- **Definition**: Time to collect and submit usage metrics
- **Measurement**: On-chain submission time - collection time
- **Collection**: Per metering cycle
- **Percentiles**: P50, P95, P99

#### 2.3 Quality SLIs

**SLI-PROV-007: Deployment Success Rate**
- **Definition**: Percentage of deployments that start successfully
- **Measurement**: Successful deployments / attempted deployments
- **Collection**: Per deployment
- **Formula**: `(successful_deployments / total_deployments) * 100`
- **Target**: > 99.0% (aligns with reliability score calculation)

**SLI-PROV-008: Deployment Stability**
- **Definition**: Percentage of deployments running without failures
- **Measurement**: MTBF (Mean Time Between Failures)
- **Collection**: Continuous uptime tracking
- **Formula**: `total_uptime / (total_uptime + total_downtime)`

**SLI-PROV-009: Resource Metering Accuracy**
- **Definition**: Accuracy of usage metrics vs actual consumption
- **Measurement**: Reported metrics validation
- **Collection**: Sample audits
- **Target**: > 99.9% accuracy

#### 2.4 Capacity SLIs

**SLI-PROV-010: Resource Utilization**
- **Definition**: Percentage of allocated resources in use
- **Measurement**: Used resources / total allocated resources
- **Collection**: Every 5 minutes
- **Dimensions**: CPU, Memory, Storage, GPU, Network

**SLI-PROV-011: Workload Capacity**
- **Definition**: Number of concurrent workloads supported
- **Measurement**: Active deployments / max capacity
- **Collection**: Continuous
- **Target**: < 80% for headroom

---

### 3. Benchmark Services

#### 3.1 Availability SLIs

**SLI-BENCH-001: Benchmark Daemon Uptime**
- **Definition**: Percentage of time benchmark daemon is operational
- **Measurement**: Process health check
- **Collection**: Every 30 seconds
- **Formula**: `(successful_checks / total_checks) * 100`

**SLI-BENCH-002: Benchmark Execution Availability**
- **Definition**: Percentage of scheduled benchmarks that execute
- **Measurement**: Completed benchmarks / scheduled benchmarks
- **Collection**: Per schedule cycle
- **Formula**: `(completed_benchmarks / scheduled_benchmarks) * 100`

#### 3.2 Latency SLIs

**SLI-BENCH-003: Benchmark Completion Latency**
- **Definition**: Time to complete benchmark suite
- **Measurement**: End time - start time
- **Collection**: Per benchmark run
- **Percentiles**: P50, P95, P99

**SLI-BENCH-004: Report Submission Latency**
- **Definition**: Time to submit benchmark report on-chain
- **Measurement**: Block inclusion time - completion time
- **Collection**: Per report
- **Percentiles**: P50, P95, P99

#### 3.3 Quality SLIs

**SLI-BENCH-005: Benchmark Success Rate**
- **Definition**: Percentage of benchmarks that complete successfully
- **Measurement**: Successful runs / total attempts
- **Collection**: Per benchmark
- **Formula**: `(successful_runs / total_runs) * 100`

**SLI-BENCH-006: Challenge Response Rate**
- **Definition**: Percentage of challenges responded to within timeout
- **Measurement**: Timely responses / total challenges
- **Collection**: Per challenge
- **Formula**: `(timely_responses / total_challenges) * 100`
- **Timeout**: Configurable via module params

---

### 4. API Services (gRPC & REST)

#### 4.1 Availability SLIs

**SLI-API-001: API Availability**
- **Definition**: Percentage of API requests that succeed
- **Measurement**: 2xx/3xx responses / total requests
- **Collection**: Per request
- **Formula**: `(success_responses / total_requests) * 100`
- **Exclusions**: 4xx client errors

**SLI-API-002: Endpoint Availability**
- **Definition**: Per-endpoint availability tracking
- **Measurement**: Successful responses per endpoint
- **Collection**: Per request, grouped by endpoint
- **Dimensions**: By endpoint path and method

#### 4.2 Latency SLIs

**SLI-API-003: Request Latency**
- **Definition**: Time to process and return API requests
- **Measurement**: Response time - request time
- **Collection**: Per request
- **Percentiles**: P50, P95, P99
- **Dimensions**: By endpoint

**SLI-API-004: Query Latency**
- **Definition**: Time to execute blockchain queries
- **Measurement**: Query execution time
- **Collection**: Per query
- **Percentiles**: P50, P95, P99

#### 4.3 Quality SLIs

**SLI-API-005: Error Rate**
- **Definition**: Percentage of requests resulting in server errors
- **Measurement**: 5xx responses / total requests
- **Collection**: Per request
- **Formula**: `(5xx_responses / total_requests) * 100`

**SLI-API-006: Query Correctness**
- **Definition**: Percentage of queries returning correct state
- **Measurement**: Validated responses / total queries
- **Collection**: Sample validation
- **Target**: 100% (deterministic blockchain state)

---

## Service Level Objectives (SLOs)

### SLO Measurement Windows

- **Rolling Window**: 28-day rolling window for trend analysis
- **Calendar Window**: Monthly calendar window for business reporting
- **Error Budget Period**: 28 days (4 weeks)

### Tier Classification

- **Tier 0 (Critical)**: Core consensus and transaction processing
- **Tier 1 (High)**: Provider daemon, API services
- **Tier 2 (Standard)**: Benchmark services, auxiliary features

---

### 1. Blockchain Node SLOs

#### Tier 0 - Critical Services

| SLO ID | Metric | Target | Measurement Window | Error Budget |
|--------|--------|--------|-------------------|--------------|
| SLO-NODE-001 | Node Uptime | 99.95% | 28 days | 21 minutes |
| SLO-NODE-002 | Consensus Participation | 99.90% | 28 days | 40 minutes |
| SLO-NODE-003 | Transaction Processing Availability | 99.90% | 28 days | 40 minutes |
| SLO-NODE-004 | Transaction Success Rate | 99.50% | 28 days | 2 hours |
| SLO-NODE-005 | Transaction Confirmation (P95) | < 10 seconds | 28 days | 5% requests |
| SLO-NODE-006 | Query Response (P95) | < 2 seconds | 28 days | 5% requests |
| SLO-NODE-007 | Block Production (P95) | < 7 seconds | 28 days | 5% blocks |
| SLO-NODE-008 | Transaction Throughput | > 50 TPS | 28 days | N/A |

#### Rationale

- **99.95% uptime** = 21 minutes downtime/month - Allows for brief maintenance/upgrades
- **99.90% consensus** = Allows missing ~144 blocks/month (6s block time)
- **P95 latency targets** = Balances user experience with realistic performance
- **50 TPS minimum** = Based on load testing baseline

---

### 2. Provider Daemon SLOs

#### Tier 1 - High Priority

| SLO ID | Metric | Target | Measurement Window | Error Budget |
|--------|--------|--------|-------------------|--------------|
| SLO-PROV-001 | Provider Daemon Uptime | 99.90% | 28 days | 40 minutes |
| SLO-PROV-002 | Bid Engine Availability | 99.50% | 28 days | 2 hours |
| SLO-PROV-003 | Deployment Success Rate | 99.00% | 28 days | 4 hours |
| SLO-PROV-004 | Deployment Stability (Uptime) | 99.50% | Per deployment | N/A |
| SLO-PROV-005 | Provisioning Latency (P95) | < 300 seconds | 28 days | 5% requests |
| SLO-PROV-006 | Bid Placement Latency (P95) | < 5 seconds | 28 days | 5% requests |
| SLO-PROV-007 | Resource Utilization | 50-80% | Real-time | N/A |
| SLO-PROV-008 | MTBF (Mean Time Between Failures) | > 24 hours | Per provider | N/A |

#### Rationale

- **99.00% deployment success** = Aligns with reliability score provisioning target
- **300 seconds provisioning** = Matches penalty-free threshold in reliability calculation
- **99.50% deployment uptime** = Balances high reliability with operational flexibility
- **24 hour MTBF** = Provides 1000 bonus points in reliability score

---

### 3. Benchmark Service SLOs

#### Tier 2 - Standard

| SLO ID | Metric | Target | Measurement Window | Error Budget |
|--------|--------|--------|-------------------|--------------|
| SLO-BENCH-001 | Benchmark Daemon Uptime | 99.50% | 28 days | 2 hours |
| SLO-BENCH-002 | Benchmark Execution Availability | 99.00% | 28 days | 4 hours |
| SLO-BENCH-003 | Benchmark Success Rate | 95.00% | 28 days | 8 hours |
| SLO-BENCH-004 | Challenge Response Rate | 99.00% | 28 days | 4 hours |
| SLO-BENCH-005 | Benchmark Completion (P95) | < 600 seconds | 28 days | 5% runs |
| SLO-BENCH-006 | Report Submission (P95) | < 30 seconds | 28 days | 5% reports |

#### Rationale

- **Lower availability targets** = Benchmark failures don't immediately impact users
- **95% success rate** = Acknowledges environmental variability in performance testing
- **99% challenge response** = Critical for anomaly detection and fraud prevention

---

### 4. API Service SLOs

#### Tier 1 - High Priority

| SLO ID | Metric | Target | Measurement Window | Error Budget |
|--------|--------|--------|-------------------|--------------|
| SLO-API-001 | API Availability | 99.90% | 28 days | 40 minutes |
| SLO-API-002 | Error Rate | < 0.50% | 28 days | N/A |
| SLO-API-003 | Request Latency (P95) | < 2 seconds | 28 days | 5% requests |
| SLO-API-004 | Query Latency (P95) | < 1 second | 28 days | 5% requests |
| SLO-API-005 | Query Correctness | 100% | 28 days | 0 errors |

#### Rationale

- **99.90% availability** = Matches blockchain node availability for consistency
- **0.50% error rate** = Aggressive target for high-quality API experience
- **100% correctness** = Non-negotiable for blockchain state queries

---

## Service Level Agreements (SLAs)

SLAs define contractual obligations and consequences for SLO violations. These apply to VirtEngine network participants.

### 1. Validator Node SLA

**Applies to**: Validators running consensus nodes

#### Availability Commitment
- **Target**: 99.50% uptime per epoch (monthly)
- **Measurement**: Consensus participation rate
- **Exclusions**: Planned maintenance with 24-hour notice

#### Consequences of SLA Violation
- **99.00% - 99.49%**: Warning notification
- **98.00% - 98.99%**: Slash 0.5% of staked tokens
- **< 98.00%**: Slash 1.0% of staked tokens + 7-day jail period

#### Credits/Compensation
- Not applicable (blockchain-enforced slashing)

---

### 2. Provider SLA

**Applies to**: Providers offering compute resources

#### Deployment Success Rate Commitment
- **Target**: 95.00% deployment success rate
- **Measurement**: Successful deployments / total bids won
- **Exclusions**: Invalid manifests, insufficient resources

#### Deployment Uptime Commitment
- **Target**: 99.00% uptime per deployment
- **Measurement**: MTBF and total uptime percentage
- **Exclusions**: Consumer-initiated terminations

#### Provisioning Speed Commitment
- **Target**: P95 < 300 seconds
- **Measurement**: Time to workload running state
- **Exclusions**: GPU deployments (higher variance)

#### Consequences of SLA Violation

**Deployment Success Rate**
- **90% - 94.99%**: Reliability score penalty (reduced trust score)
- **< 90%**: Temporary suspension from bid engine (24 hours)

**Deployment Uptime**
- **95% - 98.99%**: Reduced reliability score
- **< 95%**: Dispute eligible, potential slashing

**Provisioning Speed**
- **P95 300-600 seconds**: Reliability score penalty (up to 2000 points)
- **P95 > 600 seconds**: Review for provider flagging

#### Credits/Compensation
- Consumers eligible for dispute resolution
- Escrow refunds based on SLA violation severity
- Provider reliability score impacts future bid competitiveness

---

### 3. Network API SLA

**Applies to**: Public RPC/API endpoints operated by foundation or partners

#### Availability Commitment
- **Target**: 99.50% availability
- **Measurement**: 2xx/3xx responses excluding 4xx client errors
- **Exclusions**: DDoS attacks, force majeure events

#### Latency Commitment
- **Target**: P95 < 3 seconds
- **Measurement**: Request duration
- **Exclusions**: Complex queries (state exports, historical queries)

#### Rate Limiting
- **Free Tier**: 10 requests/second per IP
- **Premium Tier**: 100 requests/second per API key

#### Consequences of SLA Violation
- **< 99.50% monthly**: Service credits for premium tier users
- **Credit**: 5% of monthly fee per 0.1% below SLA
- **Maximum Credit**: 100% of monthly fee

---

### 4. Benchmark Service SLA

**Applies to**: Benchmark daemon operators

#### Execution Commitment
- **Target**: 95.00% of scheduled benchmarks executed
- **Measurement**: Completed runs / scheduled runs
- **Exclusions**: Provider maintenance windows

#### Challenge Response Commitment
- **Target**: 99.00% challenges responded to within timeout
- **Measurement**: Timely responses / total challenges
- **Exclusions**: Network outages

#### Consequences of SLA Violation
- **< 95% execution**: Warning and investigation
- **< 90% execution**: Benchmark operator review
- **< 99% challenge response**: Anomaly flag and potential provider impact

---

## Error Budget Policy

### Error Budget Calculation

**Formula**:
```
Error Budget = (1 - SLO_Target) × Total_Time_Period
```

**Example** (99.90% SLO over 28 days):
```
Error Budget = (1 - 0.9990) × 28 days × 24 hours × 60 minutes
            = 0.001 × 40,320 minutes
            = 40.32 minutes (2,419 seconds)
```

### Error Budget Tracking

Error budgets are tracked in real-time per service:

```go
type ErrorBudget struct {
    ServiceID        string
    SLOID           string
    BudgetPeriod    time.Duration  // 28 days
    TargetSLO       float64        // 0.9990
    BudgetMinutes   float64        // 40.32
    ConsumedMinutes float64        // Current consumption
    RemainingPct    float64        // Remaining budget %
    LastReset       time.Time
    Status          BudgetStatus   // Healthy, Warning, Critical, Depleted
}
```

### Error Budget Status Levels

| Status | Remaining Budget | Action |
|--------|-----------------|--------|
| Healthy | > 50% | Normal operations, innovation encouraged |
| Warning | 25% - 50% | Caution, prioritize stability |
| Critical | 5% - 25% | Freeze non-critical changes, incident focus |
| Depleted | < 5% | Change freeze, emergency fixes only |

### Error Budget Policy Actions

#### Healthy (> 50%)
- ✅ Feature releases allowed
- ✅ Experimental features permitted
- ✅ Aggressive deployment cadence
- ✅ Performance optimizations

#### Warning (25% - 50%)
- ⚠️ Feature releases require SRE approval
- ⚠️ Experimental features postponed
- ⚠️ Reduce deployment frequency
- ✅ Bug fixes and stability improvements

#### Critical (5% - 25%)
- 🚫 Feature releases frozen
- 🚫 Only critical bug fixes
- ⚠️ Emergency changes require VP approval
- ✅ Reliability improvements prioritized

#### Depleted (< 5%)
- 🚫 Complete change freeze
- 🚫 Only emergency security/stability fixes
- ⚠️ All changes require executive approval
- ✅ Post-incident remediation

### Error Budget Burn Rate Alerts

**Fast Burn**: Consuming budget faster than sustainable rate

```
Burn Rate = (Budget Consumed in 1 hour) / (Budget per hour)

Sustainable Rate = 1.0
Warning Threshold = 2.0x (consuming 2x faster than sustainable)
Critical Threshold = 5.0x (consuming 5x faster than sustainable)
```

**Alert Rules**:
- **2x burn rate for 1 hour** → Page on-call SRE
- **5x burn rate for 15 minutes** → Page incident commander
- **10x burn rate for 5 minutes** → Automatic rollback trigger

---

## Measurement and Monitoring

### Data Collection Architecture

```
┌─────────────────┐
│  Service Nodes  │
│  (Instrumented) │
└────────┬────────┘
         │ Metrics/Traces/Logs
         │
         ▼
┌─────────────────┐
│   Prometheus    │ ◄─── Scrape /metrics endpoints
│  (Time Series)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│    Grafana      │ ◄─── SLI/SLO Dashboards
│   (Visualize)   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Alertmanager   │ ◄─── SLO violation alerts
│  (Alert Route)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   PagerDuty/    │ ◄─── Incident escalation
│   Slack/Email   │
└─────────────────┘
```

### Instrumentation Requirements

All services must:
1. ✅ Expose `/metrics` endpoint in Prometheus format
2. ✅ Export structured logs with correlation IDs
3. ✅ Emit OpenTelemetry traces for requests
4. ✅ Report SLI metrics every 30-60 seconds
5. ✅ Include service/version labels on all metrics

### Key Metric Labels

```
service_name="virtengine-node"
service_tier="tier0"
environment="production"
region="us-west-2"
version="v1.2.3"
```

### SLO Monitoring Dashboards

**Required Dashboards**:
1. **Executive SLO Dashboard** - High-level SLO status
2. **Per-Service SLO Dashboard** - Detailed service metrics
3. **Error Budget Dashboard** - Budget consumption and burn rate
4. **Latency Distribution Dashboard** - P50/P95/P99 heatmaps
5. **Incident Impact Dashboard** - SLO impact during incidents

### Alerting Rules

See `docs/sre/ALERTING_RULES.md` for detailed alert configurations.

**Critical Alerts**:
- SLO budget depleted (< 5%)
- Fast burn rate (> 5x)
- Consensus participation drop
- API error rate spike

---

## Review and Updates

### SLO Review Schedule

- **Weekly**: Error budget review in SRE sync
- **Monthly**: SLO achievement review in service review
- **Quarterly**: SLO target adjustment review
- **Annually**: Complete SLI/SLO framework revision

### SLO Adjustment Criteria

Adjust SLOs when:
- ✅ Consistently exceeding targets (>99.99% achieved)
- ⚠️ Consistently missing targets (<90% achieved)
- 🔄 Service architecture changes significantly
- 📊 User expectations shift
- 💰 Cost/benefit analysis justifies change

### Change Management

1. Propose SLO change with justification
2. Review impact on error budgets
3. Stakeholder approval (Engineering + Product)
4. Update monitoring and alerts
5. Communicate to teams
6. Monitor for 30 days
7. Retrospective on impact

---

## References

- [Error Budget Policy](ERROR_BUDGET_POLICY.md)
- [Alerting Rules](ALERTING_RULES.md)
- [SRE Runbooks](runbooks/README.md)
- [Incident Response](INCIDENT_RESPONSE.md)
- [Capacity Planning](CAPACITY_PLANNING.md)
- [Reliability Scoring](../../x/benchmark/types/reliability.go)

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-29
**Version**: 1.0.0
**Next Review**: 2026-04-29
