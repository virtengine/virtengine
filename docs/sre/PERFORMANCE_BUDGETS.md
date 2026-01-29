# Performance Budgets

## Overview

Performance budgets define acceptable performance limits for VirtEngine services. These budgets ensure that features and changes don't degrade user experience below acceptable thresholds.

## Table of Contents
1. [What are Performance Budgets?](#what-are-performance-budgets)
2. [Budget Categories](#budget-categories)
3. [Service Performance Budgets](#service-performance-budgets)
4. [Enforcement](#enforcement)
5. [Monitoring and Alerting](#monitoring-and-alerting)

---

## What are Performance Budgets?

### Definition

A **performance budget** is a set of limits imposed on metrics that affect site/service performance. It acts as a forcing function to ensure performance remains acceptable.

### Why Performance Budgets?

1. **User Experience**: Slow systems drive users away
2. **Cost Control**: Poor performance wastes resources
3. **Reliability**: Performance and reliability are linked
4. **Decision Making**: Quantify trade-offs ("Is this feature worth 100ms?")
5. **Accountability**: Clear targets for teams

### Budget Types

1. **Timing Budgets**: Latency, response time, load time
2. **Quantity Budgets**: Bundle size, request count, memory usage
3. **Rule Budgets**: Performance best practices (e.g., < 50 DB queries)

---

## Budget Categories

### 1. Latency Budgets

**Principle**: Every service has a latency budget that must not be exceeded

#### Blockchain Node
```
Block Production Latency (P95): < 7 seconds
Transaction Confirmation (P95): < 10 seconds
Query Response (P95): < 2 seconds
State Sync (P95): < 30 seconds
```

**Rationale**:
- 7s block time = network target
- 10s confirmation = 1-2 blocks tolerable wait
- 2s query = responsive API experience

#### Provider Daemon
```
Bid Placement (P95): < 5 seconds
Deployment Provisioning (P95): < 300 seconds (5 minutes)
Manifest Validation (P95): < 1 second
Usage Metering (P95): < 10 seconds
```

**Rationale**:
- 5s bid = stay competitive in auction
- 300s provisioning = penalty-free in reliability score
- 1s validation = fast feedback for users

#### API Services
```
REST API (P95): < 2 seconds
gRPC Query (P95): < 1 second
WebSocket Message (P95): < 500ms
GraphQL Query (P95): < 3 seconds
```

**Rationale**:
- 2s REST = industry standard
- 1s gRPC = RPC call expectation
- 500ms WebSocket = real-time feel

---

### 2. Throughput Budgets

**Principle**: Services must sustain minimum throughput under load

#### Blockchain Node
```
Transaction Throughput: ≥ 50 TPS (sustained)
Peak Throughput: ≥ 100 TPS (burst)
Block Processing: ≥ 500 tx/block
```

**Rationale**:
- 50 TPS = baseline from load testing
- 100 TPS = 2x headroom for spikes

#### Provider Daemon
```
Concurrent Deployments: ≥ 10 per provider
Bid Processing Rate: ≥ 20 bids/second
Event Processing: ≥ 100 events/second
```

**Rationale**:
- 10 deployments = reasonable provider capacity
- 20 bids/s = handle busy marketplace

#### API Services
```
API Requests: ≥ 1000 RPS (sustained)
Peak Requests: ≥ 5000 RPS (burst)
Concurrent Connections: ≥ 10,000
```

**Rationale**:
- 1000 RPS = moderate production load
- 5000 RPS = marketing campaign spike

---

### 3. Resource Budgets

**Principle**: Services must operate within resource constraints

#### Memory Budgets
```
Blockchain Node: < 16 GB RAM (normal operation)
Provider Daemon: < 4 GB RAM per instance
API Server: < 2 GB RAM per instance
Benchmark Daemon: < 1 GB RAM
```

**Rationale**:
- Cost control
- Prevent memory leaks from causing OOM
- Enable horizontal scaling

#### CPU Budgets
```
Blockchain Node: < 4 cores (average)
Provider Daemon: < 2 cores (average)
API Server: < 1 core (average)
Benchmark Daemon: < 0.5 core (average)
```

**Rationale**:
- Efficient resource utilization
- Predictable scaling costs

#### Storage Budgets
```
Blockchain Node: < 500 GB/year growth
Provider Daemon: < 100 GB persistent storage
API Server: < 10 GB cache storage
```

**Rationale**:
- Sustainable storage costs
- Prevent unbounded growth

#### Network Budgets
```
Blockchain Node: < 100 Mbps (average)
Provider Daemon: < 50 Mbps (average)
API Server: < 10 Mbps (average)
```

**Rationale**:
- Network bandwidth costs
- Prevent network saturation

---

### 4. Quality Budgets

**Principle**: Services must maintain quality standards

#### Error Rate Budgets
```
API Error Rate: < 0.5%
Transaction Failure Rate: < 1%
Deployment Failure Rate: < 1%
Benchmark Failure Rate: < 5%
```

**Linked to**: Error budgets from SLO framework

#### Availability Budgets
```
Blockchain Node: 99.95% uptime
Provider Daemon: 99.90% uptime
API Services: 99.90% uptime
```

**Linked to**: SLO targets

---

## Service Performance Budgets

### VirtEngine Blockchain Node

#### Critical Path: Transaction Processing

**Performance Budget**:
```
Transaction Submission → Mempool: < 100ms (P95)
Mempool → Block Proposal: < 5s (P95)
Block Proposal → Consensus: < 2s (P95)
Consensus → Finalization: < 7s (P95)
Total: < 10s (P95)
```

**Budget Allocation**:
| Stage | Budget | Current | Headroom |
|-------|--------|---------|----------|
| Validation | 100ms | 50ms | 50ms ✅ |
| Mempool | 5s | 3s | 2s ✅ |
| Consensus | 2s | 1.5s | 500ms ✅ |
| Finalization | 7s | 6s | 1s ✅ |
| **Total** | **10s** | **7.55s** | **2.45s** ✅ |

**Enforcement**:
- Load tests validate against budget
- Regressions block PR merge
- Alerts if approaching budget

---

### Provider Daemon

#### Critical Path: Deployment Provisioning

**Performance Budget**:
```
Order Event → Bid Decision: < 1s (P95)
Bid Submission → Win Notification: < 5s (P95)
Manifest Parsing → Validation: < 1s (P95)
Kubernetes Scheduling → Pod Running: < 280s (P95)
Health Check → Deployment Active: < 20s (P95)
Total: < 300s (P95)
```

**Budget Allocation**:
| Stage | Budget | Current | Headroom |
|-------|--------|---------|----------|
| Bid Decision | 1s | 500ms | 500ms ✅ |
| Bid Win | 5s | 3s | 2s ✅ |
| Validation | 1s | 200ms | 800ms ✅ |
| K8s Scheduling | 280s | 180s | 100s ✅ |
| Health Check | 20s | 10s | 10s ✅ |
| **Total** | **300s** | **193.7s** | **106.3s** ✅ |

---

### API Services

#### Critical Path: Query Request

**Performance Budget**:
```
Request Receipt → Auth: < 50ms (P95)
Auth → Query Parsing: < 50ms (P95)
Query Execution → Response Build: < 1s (P95)
Response Serialization → Send: < 100ms (P95)
Total: < 2s (P95)
```

**Budget Allocation**:
| Stage | Budget | Current | Headroom |
|-------|--------|---------|----------|
| Auth | 50ms | 20ms | 30ms ✅ |
| Parsing | 50ms | 10ms | 40ms ✅ |
| Execution | 1s | 500ms | 500ms ✅ |
| Serialization | 100ms | 50ms | 50ms ✅ |
| **Total** | **2s** | **0.58s** | **1.42s** ✅ |

---

## Enforcement

### 1. Pre-Commit Hooks

**Performance Testing in CI**:

```yaml
# .github/workflows/performance-check.yml
name: Performance Budget Check

on: [pull_request]

jobs:
  performance:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v3

      - name: Run Benchmark Suite
        run: go test -bench=. -benchmem ./...

      - name: Compare to Baseline
        run: |
          ./scripts/perf-compare.sh \
            --baseline benchmarks/baseline.json \
            --current benchmarks/current.json \
            --threshold 10  # 10% regression allowed

      - name: Block if Budget Exceeded
        if: failure()
        run: |
          echo "❌ Performance budget exceeded!"
          echo "Review the regression and optimize or update budget."
          exit 1
```

---

### 2. Load Testing Gates

**Load Test Budget Validation**:

```go
// tests/load/budget_validation_test.go

func TestTransactionLatencyBudget(t *testing.T) {
    const budget = 10 * time.Second // P95 budget

    // Run load test
    results := runLoadTest(t, 50 /* TPS */, 5*time.Minute)

    p95 := results.Latency.P95()

    if p95 > budget {
        t.Errorf("Transaction latency P95 exceeded budget: %v > %v",
            p95, budget)
    }

    // Log headroom
    headroom := budget - p95
    t.Logf("Latency headroom: %v (%.1f%%)",
        headroom, float64(headroom)/float64(budget)*100)
}
```

---

### 3. Production Monitoring

**Real-Time Budget Alerts**:

```yaml
# alerting/performance-budget-alerts.yml

groups:
  - name: performance_budgets
    interval: 1m
    rules:
      # Transaction latency budget
      - alert: TransactionLatencyBudgetExceeded
        expr: histogram_quantile(0.95, tx_confirmation_latency_seconds) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Transaction P95 latency exceeded 10s budget"
          description: "Current: {{ $value }}s"

      # API latency budget
      - alert: APILatencyBudgetExceeded
        expr: histogram_quantile(0.95, api_request_duration_seconds) > 2
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "API P95 latency exceeded 2s budget"

      # Throughput budget
      - alert: ThroughputBudgetNotMet
        expr: rate(tx_processed_total[5m]) < 50
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Transaction throughput below 50 TPS budget"
```

---

### 4. Budget Review Process

**Monthly Performance Review**:

1. **Collect Metrics** (Week 1):
   - Actual performance vs budgets
   - Headroom analysis
   - Trend analysis

2. **Identify Issues** (Week 2):
   - Budget violations
   - Degradation trends
   - Capacity concerns

3. **Plan Improvements** (Week 3):
   - Performance optimizations
   - Budget adjustments (if needed)
   - Infrastructure upgrades

4. **Update Budgets** (Week 4):
   - Tighten if consistently under-utilized
   - Loosen if unrealistic
   - Document changes

---

## Monitoring and Alerting

### Performance Dashboard

**Key Panels**:

1. **Budget Utilization Gauges**
   ```
   Latency Budget Usage: Current P95 / Budget P95
   - Green: < 70%
   - Yellow: 70-90%
   - Red: > 90%
   ```

2. **Headroom Tracking**
   ```
   Headroom = Budget - Actual

   Graph: Headroom over time
   - Positive = under budget ✅
   - Negative = over budget ❌
   ```

3. **Trend Lines**
   ```
   30-day trend of P95 latencies
   - Slope indicates degradation/improvement
   - Forecast when budget will be exceeded
   ```

4. **Budget Violations**
   ```
   Count of budget violations in past 30 days
   - By service
   - By metric
   ```

---

### Performance Budget Report

**Weekly Report Template**:

```markdown
## Performance Budget Report - Week of 2026-01-29

### Summary
- 📊 Budgets Monitored: 24
- ✅ Within Budget: 22 (91.7%)
- ⚠️ Approaching Budget: 1 (4.2%)
- ❌ Exceeded Budget: 1 (4.2%)

### Budget Status

| Service | Metric | Budget | Actual | Headroom | Status |
|---------|--------|--------|--------|----------|--------|
| Node | TX Latency P95 | 10s | 7.2s | 2.8s | ✅ |
| Provider | Provisioning P95 | 300s | 285s | 15s | ⚠️ |
| API | Query P95 | 2s | 2.1s | -0.1s | ❌ |
| Benchmark | Completion P95 | 600s | 450s | 150s | ✅ |

### Action Items
1. **API Query P95 over budget** (❌)
   - Current: 2.1s vs budget 2s
   - Root cause: Database query optimization needed
   - Owner: @backend-team
   - Due: 2026-02-05

2. **Provider Provisioning approaching budget** (⚠️)
   - Current: 285s vs budget 300s (95%)
   - Monitor closely, optimize if trend continues
   - Owner: @sre-team

### Trends
- Transaction latency improving (7.5s → 7.2s)
- API latency degrading (1.8s → 2.1s) ⚠️

### Recommendations
1. Investigate API query regression
2. Consider increasing Provider budget to 320s (more realistic)
3. Continue monitoring trends
```

---

## Budget Adjustment Policy

### When to Adjust Budgets

**Tighten Budget** (make more aggressive):
- Consistently operating at < 60% of budget
- Technology improvements enable better performance
- Competitor benchmarks show we're behind

**Loosen Budget** (make more lenient):
- Consistently exceeding budget despite optimization efforts
- Fundamental constraints discovered
- User expectations lower than budget

### Adjustment Process

1. **Proposal**: SRE team proposes budget change
2. **Justification**: Document reasoning and impact
3. **Stakeholder Review**: Engineering + Product review
4. **Approval**: VP Engineering approval required
5. **Implementation**: Update docs, monitoring, alerts
6. **Communication**: Announce to engineering team
7. **Monitoring**: Track impact for 30 days

---

## Performance Optimization Workflow

### Budget Violation Response

**When a budget is exceeded**:

1. **Acknowledge** (< 1 hour):
   - Create ticket
   - Assign owner
   - Notify stakeholders

2. **Investigate** (< 24 hours):
   - Reproduce issue
   - Profile performance
   - Identify bottleneck

3. **Plan** (< 48 hours):
   - Design optimization
   - Estimate effort
   - Get approval

4. **Implement** (< 1 week):
   - Code changes
   - Test improvements
   - Deploy to production

5. **Validate** (< 2 weeks):
   - Confirm budget met
   - Monitor stability
   - Document learnings

---

## Performance Testing Framework

### Continuous Performance Testing

```go
// tests/performance/continuous_test.go

// BenchmarkTransactionSubmission validates transaction latency budget
func BenchmarkTransactionSubmission(b *testing.B) {
    ctx := context.Background()
    client := setupTestClient(b)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        start := time.Now()

        // Submit transaction
        tx := generateTestTx(i)
        _, err := client.SubmitTx(ctx, tx)
        if err != nil {
            b.Fatalf("tx submission failed: %v", err)
        }

        latency := time.Since(start)

        // Validate against budget (100ms P95)
        if latency > 100*time.Millisecond {
            b.Logf("⚠️  TX %d latency %v exceeds 100ms budget", i, latency)
        }
    }

    // Report percentiles
    b.ReportMetric(float64(b.Elapsed())/float64(b.N)/1e6, "ms/op")
}
```

---

## References

- [SLI/SLO/SLA Framework](SLI_SLO_SLA.md)
- [Error Budget Policy](ERROR_BUDGET_POLICY.md)
- [Capacity Planning](CAPACITY_PLANNING.md)
- [Load Testing Framework](../tests/load/README.md)
- [Web Performance Budget Guide](https://web.dev/performance-budgets-101/)

---

**Document Owner**: SRE Team
**Last Updated**: 2026-01-29
**Version**: 1.0.0
**Next Review**: 2026-04-29
