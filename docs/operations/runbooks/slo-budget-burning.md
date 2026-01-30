# Runbook: SLO Budget Burning

## Alert Details

| Field | Value |
|-------|-------|
| Alert Name | SLOBudgetBurning |
| Severity | Warning / Critical |
| Service | Multiple |
| Tier | Tier 1 |
| SLO Impact | Direct error budget consumption |

## Summary

This alert fires when an SLO's error budget is being consumed faster than sustainable. Different severity levels indicate different burn rates:

- **Warning (14.4x)**: Budget will be exhausted in ~2 days at current rate
- **Critical (6x)**: Budget will be exhausted in ~5 days at current rate

## SLO Definitions

| SLO ID | Service | Target | Error Budget (28d) |
|--------|---------|--------|-------------------|
| SLO-API-001 | API Availability | 99.9% | 0.1% (~40 min) |
| SLO-API-002 | API Latency P95 | <500ms | 0.1% requests |
| SLO-NODE-001 | Node Uptime | 99.95% | 0.05% (~21 min) |
| SLO-VEID-001 | Verification Success | 99.5% | 0.5% requests |
| SLO-MARKET-001 | Order Processing | 99.9% | 0.1% orders |

## Understanding Error Budgets

```
Error Budget = (1 - SLO Target) × Time Period

For 99.9% SLO over 28 days:
Budget = 0.1% × 28 days × 24 hours × 60 min = ~40 minutes of downtime

Burn Rate = Actual Error Rate / Allowed Error Rate
- Burn Rate 1.0 = Sustainable (will use exactly 100% budget in window)
- Burn Rate 2.0 = Will exhaust budget in 14 days
- Burn Rate 14.4 = Will exhaust budget in ~2 days
```

## Diagnostic Steps

### 1. Identify the Affected SLO

```bash
# Check Prometheus for budget status
curl -s "http://prometheus:9090/api/v1/query?query=virtengine_slo_error_budget_remaining" | jq '.data.result[]'

# Check current error rate
curl -s "http://prometheus:9090/api/v1/query?query=virtengine_api_error_rate_5m" | jq '.data.result[]'
```

### 2. Analyze Error Patterns

```promql
# Error rate over time
rate(virtengine_api_requests_total{status=~"5.."}[5m])

# Error rate by endpoint
sum(rate(virtengine_api_requests_total{status=~"5.."}[5m])) by (endpoint)

# Error rate by error code
sum(rate(virtengine_errors_total[5m])) by (error_code)
```

### 3. Check Grafana Dashboard

Navigate to: **Grafana → SLO Overview Dashboard**

Look for:
- Error budget consumption rate graph
- Top error-producing endpoints
- Correlation with deployments/changes

### 4. Correlate with Changes

```bash
# Recent deployments
kubectl get events -n virtengine --sort-by='.lastTimestamp' | tail -20

# Recent config changes
git log --oneline --since="2 hours ago" -- deploy/

# Recent alerts
curl -s http://alertmanager:9093/api/v2/alerts | jq '.[] | select(.status.state=="active")'
```

## Resolution Steps

### Scenario 1: High Error Rate from Specific Endpoint

```bash
# 1. Identify problematic endpoint
curl -s "http://prometheus:9090/api/v1/query?query=topk(5,sum(rate(virtengine_api_requests_total{status=~\"5..\"}[5m]))by(endpoint))" | jq '.data.result'

# 2. Check endpoint logs
kubectl logs -n virtengine -l app=virtengine-api --since=30m | grep "ERROR"

# 3. If due to recent deployment, consider rollback
kubectl rollout undo deployment/virtengine-api -n virtengine

# 4. If due to downstream dependency, check dependency health
curl -s http://localhost:26657/health
```

### Scenario 2: Latency SLO Burning

```bash
# 1. Check latency distribution
curl -s "http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,rate(virtengine_api_request_duration_seconds_bucket[5m]))" | jq '.data.result'

# 2. Identify slow endpoints
curl -s "http://prometheus:9090/api/v1/query?query=topk(5,histogram_quantile(0.95,sum(rate(virtengine_api_request_duration_seconds_bucket[5m]))by(endpoint,le)))" | jq '.data.result'

# 3. Check for resource constraints
kubectl top pods -n virtengine

# 4. Scale if needed
kubectl scale deployment/virtengine-api -n virtengine --replicas=5
```

### Scenario 3: Node Uptime SLO Burning

```bash
# 1. Check node health
virtengined status | jq '.NodeInfo.network, .SyncInfo.catching_up'

# 2. Check for restart patterns
journalctl -u virtengined --since "1 hour ago" | grep -E "start|stop|crash"

# 3. If frequent restarts, check resources
free -h
df -h
dmesg | tail -50

# 4. If memory pressure, increase limits
# Edit systemd unit or k8s resource limits
```

### Scenario 4: VEID Verification SLO Burning

```bash
# 1. Check verification failure rate
curl -s "http://prometheus:9090/api/v1/query?query=rate(virtengine_veid_verification_failures_total[5m])" | jq '.data.result'

# 2. Check failure reasons
grep "verification_failed" /var/log/virtengine/veid.log | tail -20 | jq '.reason'

# 3. If ML inference issues
./scripts/check-inference-health.sh

# 4. If encryption issues
./scripts/check-encryption-health.sh
```

## Mitigation Strategies

### Immediate Actions (Reduce Burn Rate)

1. **Traffic reduction**: Implement rate limiting if under attack
2. **Graceful degradation**: Return cached responses for non-critical endpoints
3. **Rollback**: Revert recent changes if correlated
4. **Scale out**: Add capacity if resource-constrained

### Medium-term Actions (Restore Budget)

1. **Fix root cause**: Address underlying issues
2. **Improve resilience**: Add retries, circuit breakers
3. **Better monitoring**: Add more granular alerts
4. **Capacity planning**: Ensure adequate headroom

## Error Budget Policy

Per [SLI_SLO_SLA.md](../../sre/SLI_SLO_SLA.md):

| Budget Remaining | Action |
|------------------|--------|
| >50% | Normal operations |
| 25-50% | Freeze risky changes |
| 10-25% | Focus on reliability |
| <10% | Emergency mode, stability only |

```bash
# Check current budget
BUDGET=$(curl -s "http://prometheus:9090/api/v1/query?query=virtengine_slo_error_budget_remaining{slo=\"api-availability\"}" | jq -r '.data.result[0].value[1]')

if (( $(echo "$BUDGET < 0.10" | bc -l) )); then
  echo "EMERGENCY: Budget below 10% - stability changes only"
elif (( $(echo "$BUDGET < 0.25" | bc -l) )); then
  echo "WARNING: Budget below 25% - focus on reliability"
fi
```

## Recovery Verification

```bash
# 1. Verify burn rate has decreased
watch -n 60 'curl -s "http://prometheus:9090/api/v1/query?query=virtengine_slo_burn_rate_1h" | jq ".data.result[0].value[1]"'

# 2. Confirm error rate is below threshold
curl -s "http://prometheus:9090/api/v1/query?query=virtengine_api_error_rate_5m" | jq '.data.result[0].value[1]'

# 3. Check budget is no longer decreasing rapidly
# Should see burn rate < 1.0 for budget recovery
```

## Escalation

**Escalate to L2 if**:
- Cannot identify root cause
- Multiple SLOs affected
- Budget below 25%

**Escalate to L3 if**:
- Budget below 10%
- Customer-reported issues
- Potential SLA breach

## Communication

### Internal Update Template

```
SLO Budget Alert Update

Affected SLO: [SLO-ID]
Current Budget: [X]%
Burn Rate: [X]x sustainable

Status: [Investigating | Mitigating | Resolved]
Impact: [Description of user impact]
Actions: [Current actions being taken]
ETA: [Expected resolution time]
```

## Post-Incident

1. Update error budget tracking spreadsheet
2. If budget < 50% consumed, schedule reliability review
3. Consider adjusting SLO if consistently difficult to meet
4. Review and improve alerting thresholds if needed

## Related Alerts

- `HighErrorRate` - Direct contributor to budget burn
- `HighLatency` - Latency SLO contributor
- `NodeDown` - Availability SLO contributor

## References

- [SLI/SLO/SLA Definitions](../../sre/SLI_SLO_SLA.md)
- [Error Budget Policy](../../sre/error-budget-policy.md)
- [Google SRE Book - SLOs](https://sre.google/sre-book/service-level-objectives/)
