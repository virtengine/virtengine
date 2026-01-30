# Runbook: High Error Rate

## Alert Details

| Field | Value |
|-------|-------|
| Alert Name | HighErrorRate |
| Severity | Warning / Critical |
| Service | Multiple |
| Tier | Tier 1 |
| SLO Impact | Multiple SLOs affected |

## Summary

This alert fires when the error rate for a service exceeds threshold:
- **Warning**: Error rate > 1% for 5 minutes
- **Critical**: Error rate > 5% for 5 minutes

## Impact

- **High**: User-facing requests failing
- **Medium**: API reliability degraded
- **Medium**: Potential cascade to downstream services

## Diagnostic Steps

### 1. Identify Error Source

```bash
# Check error rate by service
curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(virtengine_errors_total[5m]))by(service)" | jq '.data.result'

# Check error rate by error code
curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(virtengine_errors_total[5m]))by(error_code)" | jq '.data.result'

# Check HTTP error codes
curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(virtengine_api_requests_total{status=~\"5..\"}[5m]))by(endpoint)" | jq '.data.result'
```

### 2. Analyze Error Patterns

```bash
# Check Grafana for error trends
# Navigate to: API Dashboard → Error Rate panel

# Check logs for error details
kubectl logs -n virtengine -l app=virtengine-api --since=15m | grep -i error | head -50

# Loki query
# {service="virtengine-api"} |= "error" | json | line_format "{{.level}} {{.error}} {{.endpoint}}"
```

### 3. Check for Recent Changes

```bash
# Recent deployments
kubectl get events -n virtengine --sort-by='.lastTimestamp' | grep -E "Pulled|Started|Created" | tail -10

# Git history
git log --oneline --since="2 hours ago"

# Config changes
kubectl get configmaps -n virtengine -o yaml | diff - /tmp/last-known-good-config.yaml
```

### 4. Check Dependencies

```bash
# Check chain node health
curl -s http://localhost:26657/health

# Check database connectivity
kubectl exec -it -n virtengine deployment/virtengine-api -- nc -zv postgres 5432

# Check external services
curl -s http://localhost:8080/health/dependencies | jq
```

## Resolution Steps

### Scenario 1: Backend Service Errors (5xx)

```bash
# 1. Check service logs
kubectl logs -n virtengine deployment/virtengine-api --tail=100 | grep -E "ERROR|FATAL"

# 2. Check resource constraints
kubectl top pods -n virtengine
kubectl describe pod -n virtengine <pod-name> | grep -A 5 "Resources:"

# 3. If OOMKilled, increase memory
kubectl patch deployment virtengine-api -n virtengine -p '{"spec":{"template":{"spec":{"containers":[{"name":"api","resources":{"limits":{"memory":"2Gi"}}}]}}}}'

# 4. If recent deployment caused it, rollback
kubectl rollout undo deployment/virtengine-api -n virtengine
```

### Scenario 2: Database Connection Errors

```bash
# 1. Check database status
kubectl get pods -n virtengine -l app=postgres

# 2. Check connection pool
kubectl exec -it -n virtengine deployment/virtengine-api -- cat /tmp/db-pool-stats

# 3. If pool exhausted, increase connections
kubectl set env deployment/virtengine-api -n virtengine DB_MAX_CONNECTIONS=100

# 4. If database down, restart
kubectl delete pod -n virtengine -l app=postgres
```

### Scenario 3: Chain Node Connection Errors

```bash
# 1. Check chain node health
curl -s http://localhost:26657/status | jq '.result.sync_info'

# 2. If node is syncing, wait or switch to another node
kubectl set env deployment/virtengine-api -n virtengine CHAIN_RPC_URL=http://node2:26657

# 3. If node crashed, restart
systemctl restart virtengined
```

### Scenario 4: Rate Limiting / Overload

```bash
# 1. Check request rate
curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(virtengine_api_requests_total[5m]))" | jq '.data.result[0].value[1]'

# 2. If spike in traffic, enable rate limiting
kubectl set env deployment/virtengine-api -n virtengine RATE_LIMIT_ENABLED=true RATE_LIMIT_RPS=100

# 3. Scale up if legitimate traffic
kubectl scale deployment/virtengine-api -n virtengine --replicas=5
```

### Scenario 5: External Dependency Failure

```bash
# 1. Identify failing dependency
curl -s http://localhost:8080/health/dependencies | jq

# 2. Check external service status
curl -I https://external-service.example.com/health

# 3. Enable circuit breaker if available
kubectl set env deployment/virtengine-api -n virtengine CIRCUIT_BREAKER_ENABLED=true

# 4. Enable graceful degradation
kubectl set env deployment/virtengine-api -n virtengine FALLBACK_MODE=true
```

## Recovery Verification

```bash
# 1. Verify error rate is decreasing
watch -n 30 'curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(virtengine_errors_total[5m]))" | jq ".data.result[0].value[1]"'

# 2. Check successful request rate
curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(virtengine_api_requests_total{status=~\"2..\"}[5m]))" | jq '.data.result[0].value[1]'

# 3. Test critical endpoints
curl -w "%{http_code}" http://localhost:8080/api/v1/health
curl -w "%{http_code}" http://localhost:8080/api/v1/status

# 4. Verify no new error patterns in logs
kubectl logs -n virtengine -l app=virtengine-api --since=5m | grep -c ERROR
```

## Error Code Reference

| Code | Description | Action |
|------|-------------|--------|
| 500 | Internal Server Error | Check application logs |
| 502 | Bad Gateway | Check upstream service |
| 503 | Service Unavailable | Check service health |
| 504 | Gateway Timeout | Check latency, increase timeout |
| ERR_DB_CONN | Database connection failed | Check database |
| ERR_CHAIN_RPC | Chain RPC failed | Check chain node |
| ERR_TIMEOUT | Operation timeout | Check slow operations |
| ERR_VALIDATION | Request validation failed | Check client input |

## Escalation

**Escalate to L2 if**:
- Error rate > 10% for 10+ minutes
- Cannot identify root cause
- Multiple services affected

**Escalate to L3 if**:
- Error rate > 25%
- Customer-reported widespread issues
- Potential data integrity issues

## Post-Incident

1. Document error patterns and root cause
2. Review and update error handling
3. Add more specific alerting if needed
4. Consider adding circuit breakers

## Related Alerts

- `SLOBudgetBurning` - Consequence of high error rate
- `HighLatency` - Often correlated
- `ServiceDown` - If errors are total failure

## References

- [Error Handling Guide](../../error-handling.md)
- [API Error Codes](../../api-error-codes.md)
- [Circuit Breaker Configuration](../../circuit-breaker.md)
