# VirtEngine On-Call Runbook

**Version:** 1.0.0  
**Last Updated:** 2026-01-30  
**Owner:** SRE Team

---

## Quick Reference Card

### Emergency Contacts

| Role | Contact | Hours |
|------|---------|-------|
| Primary On-Call | PagerDuty | 24/7 |
| Secondary On-Call | PagerDuty escalation | 24/7 |
| Security Team | security@virtengine.com | 24/7 |
| Infrastructure Lead | Slack: #sre-escalation | Business hours |

### Critical Dashboards

| Dashboard | URL |
|-----------|-----|
| Chain Health | https://grafana.virtengine.com/d/chain-health |
| VEID Scoring | https://grafana.virtengine.com/d/veid-scoring |
| Marketplace | https://grafana.virtengine.com/d/marketplace |
| HPC Scheduling | https://grafana.virtengine.com/d/hpc-scheduling |
| Error Budget | https://grafana.virtengine.com/d/error-budget |

### Key Commands

```bash
# Check chain status
virtengine status

# Check node sync status
curl -s http://localhost:26657/status | jq '.result.sync_info'

# Check validator status
virtengine query staking validators --status bonded | head -50

# Check mempool
curl -s http://localhost:26657/num_unconfirmed_txs | jq

# Check provider daemon health
curl -s http://localhost:8443/health

# Check recent errors
journalctl -u virtengine --since "10 minutes ago" | grep -i error
```

---

## Severity Classification

| Severity | Definition | Response Time | Examples |
|----------|------------|---------------|----------|
| **SEV-1** | Complete outage | 5 min | Chain halt, data breach |
| **SEV-2** | Major degradation | 15 min | VEID scoring down, marketplace stalled |
| **SEV-3** | Minor degradation | 1 hour | Elevated error rates |
| **SEV-4** | Minimal impact | 24 hours | Non-critical bugs |

---

## Common Incidents

### 1. Chain Halted

**Alert:** `ChainHalted`  
**Severity:** SEV-1

#### Symptoms
- No new blocks for > 30 seconds
- All transactions failing
- Consensus timeouts in logs

#### Quick Diagnosis

```bash
# 1. Check current block height and time since last block
curl -s http://localhost:26657/status | jq '.result.sync_info'

# 2. Check consensus state
curl -s http://localhost:26657/dump_consensus_state | jq '.result.round_state.height_vote_set'

# 3. Check peer connectivity
curl -s http://localhost:26657/net_info | jq '.result.n_peers'

# 4. Check for crash logs
journalctl -u virtengine -n 200 --no-pager | grep -E "(panic|error|fatal)"
```

#### Resolution Steps

1. **Check if < 1/3 validators offline:**
   - Coordinate with validator operators via Discord
   - Restart affected validators
   - Chain should resume automatically

2. **Check if ≥ 1/3 validators offline:**
   - Declare SEV-1 incident
   - Emergency validator coordination
   - Identify root cause (network, bug, attack)
   - Coordinate simultaneous restart

3. **If state corrupted:**
   ```bash
   # Stop all validators
   systemctl stop virtengine
   
   # Rollback to last good state
   virtengine rollback
   
   # Coordinate restart at same height
   ```

---

### 2. VEID Scoring Degraded

**Alert:** `VEIDScoringLatencyHigh`, `VEIDMLInferenceFailureRate`  
**Severity:** SEV-2

#### Symptoms
- Identity scores taking > 5 minutes
- High error rate on scoring transactions
- ML inference failures in logs

#### Quick Diagnosis

```bash
# 1. Check model status
virtengine query veid model-status

# 2. Check scoring queue depth
virtengine query veid pending-scores --count

# 3. Check validator score agreement
virtengine query veid recent-scores --limit 10

# 4. Check ML inference logs
journalctl -u ml-inference --since "10 minutes ago" | grep -E "(error|timeout)"

# 5. Check GPU status (if applicable)
nvidia-smi
```

#### Resolution Steps

1. **If single validator issue:**
   - Restart ML model container on affected validator
   - Verify model loads correctly
   - Monitor for recovery

2. **If model version mismatch:**
   ```bash
   # Check model versions across validators
   for node in node-0 node-1 node-2; do
     echo "$node: $(ssh $node 'virtengine query veid model-version')"
   done
   
   # Update model to latest version
   virtengine tx veid update-model --from operator
   ```

3. **If queue backlog:**
   - Temporarily increase inference resources
   - Consider rate limiting new submissions
   - Scale inference pods if on Kubernetes

---

### 3. Marketplace Order Backlog

**Alert:** `MarketOrderBacklog`, `NoBidsReceived`  
**Severity:** SEV-2

#### Symptoms
- Orders stuck in PENDING state
- No bids received for orders
- Provider daemon errors

#### Quick Diagnosis

```bash
# 1. Check pending orders
virtengine query market orders --state pending --limit 50

# 2. Check active providers
virtengine query provider list --status active | wc -l

# 3. Check provider daemon health
curl -s http://provider-daemon:8443/health | jq

# 4. Check bid engine logs
kubectl logs -l app=provider-daemon --tail=100 | grep -i "bid\|error"

# 5. Check provider capacity
virtengine query provider capacity --provider <addr>
```

#### Resolution Steps

1. **If no providers online:**
   - Contact provider operators
   - Check network/chain connectivity
   - Consider temporary marketplace pause

2. **If providers not bidding:**
   ```bash
   # Check bid engine configuration
   kubectl get configmap provider-daemon-config -o yaml
   
   # Restart bid engine
   kubectl rollout restart deployment/provider-daemon
   ```

3. **If orders malformed:**
   - Identify pattern in failing orders
   - Notify affected customers
   - Fix SDK bug if applicable

---

### 4. HPC Job Failures Elevated

**Alert:** `HPCJobFailureRateHigh`  
**Severity:** SEV-3

#### Symptoms
- Job failure rate above 10%
- Jobs failing at specific providers
- Resource allocation errors

#### Quick Diagnosis

```bash
# 1. Check recent job failures
virtengine query hpc jobs --state failed --limit 20

# 2. Analyze failure reasons
virtengine query hpc jobs --state failed --limit 100 | \
  jq -r '.jobs[].failure_reason' | sort | uniq -c | sort -rn

# 3. Check provider-specific failure rates
virtengine query hpc jobs --state failed --provider <addr> --limit 20

# 4. Check SLURM/K8s cluster status
kubectl get nodes -o wide
squeue -a
```

#### Resolution Steps

1. **If provider-specific:**
   - Notify provider operator
   - Consider temporary provider delist
   - Route jobs to healthy providers

2. **If resource exhaustion:**
   - Scale provider capacity
   - Adjust job scheduling priorities
   - Implement better queueing

3. **If user error patterns:**
   - Improve input validation
   - Update documentation
   - Add helpful error messages

---

### 5. High Error Rate

**Alert:** `HighErrorRate`, `CriticalErrorRate`  
**Severity:** SEV-2 / SEV-3

#### Symptoms
- Error rate > 10 errors/sec
- Multiple module errors
- Retryable errors increasing

#### Quick Diagnosis

```bash
# 1. Check error distribution by module
curl -s http://localhost:9090/api/v1/query?query=sum(rate(virtengine_errors_total[5m]))by(module) | jq

# 2. Check error distribution by code
curl -s http://localhost:9090/api/v1/query?query=topk(10,sum(rate(virtengine_errors_total[5m]))by(code)) | jq

# 3. Check recent error logs
journalctl -u virtengine --since "5 minutes ago" | grep -i error | tail -50

# 4. Check for recent deployments
git log --oneline -10
kubectl rollout history deployment/virtengine
```

#### Resolution Steps

1. **Identify root cause:**
   - Check recent deployments/config changes
   - Review error patterns
   - Correlate with infrastructure events

2. **If deployment-related:**
   ```bash
   # Rollback deployment
   kubectl rollout undo deployment/virtengine
   
   # Or revert config
   git revert <commit>
   ```

3. **If resource-related:**
   - Scale resources if needed
   - Check for memory leaks
   - Review timeout configurations

---

### 6. Database Connection Issues

**Alert:** `PostgresConnectionPoolExhausted`  
**Severity:** SEV-2

#### Symptoms
- Connection timeouts
- "too many connections" errors
- Slow queries

#### Quick Diagnosis

```bash
# 1. Check connection count
psql -c "SELECT count(*) FROM pg_stat_activity;"

# 2. Check connection states
psql -c "SELECT state, count(*) FROM pg_stat_activity GROUP BY state;"

# 3. Check long-running queries
psql -c "SELECT pid, now() - query_start AS duration, query 
         FROM pg_stat_activity 
         WHERE state = 'active' 
         ORDER BY duration DESC LIMIT 10;"

# 4. Check for locks
psql -c "SELECT * FROM pg_locks WHERE NOT granted;"
```

#### Resolution Steps

1. **Kill long-running queries:**
   ```sql
   SELECT pg_terminate_backend(pid) 
   FROM pg_stat_activity 
   WHERE state = 'active' 
   AND now() - query_start > interval '5 minutes';
   ```

2. **Increase connection pool:**
   ```bash
   # Update config and restart
   kubectl set env deployment/virtengine POSTGRES_MAX_CONNECTIONS=500
   ```

3. **Scale database:**
   - Add read replicas
   - Increase instance size
   - Implement connection pooler (PgBouncer)

---

### 7. Node Out of Sync

**Alert:** `NodeOutOfSync`, `NodeBehind`  
**Severity:** SEV-3

#### Symptoms
- Node in fast sync mode
- Block height behind network
- Stale query results

#### Quick Diagnosis

```bash
# 1. Check sync status
curl -s http://localhost:26657/status | jq '.result.sync_info'

# 2. Check block height vs network
NETWORK_HEIGHT=$(curl -s https://rpc.virtengine.com/status | jq -r '.result.sync_info.latest_block_height')
LOCAL_HEIGHT=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
echo "Network: $NETWORK_HEIGHT, Local: $LOCAL_HEIGHT, Behind: $((NETWORK_HEIGHT - LOCAL_HEIGHT))"

# 3. Check peer count
curl -s http://localhost:26657/net_info | jq '.result.n_peers'

# 4. Check disk space
df -h /var/lib/virtengine
```

#### Resolution Steps

1. **If < 100 blocks behind:**
   - Wait for automatic catch-up
   - Monitor sync progress

2. **If significantly behind:**
   ```bash
   # Option 1: State sync (faster)
   # Stop node, clear data, enable state sync, restart
   
   # Option 2: Snapshot restore
   wget https://snapshots.virtengine.com/latest.tar.gz
   tar -xzf latest.tar.gz -C /var/lib/virtengine/data
   systemctl restart virtengine
   ```

3. **If disk space issue:**
   - Run pruning
   - Archive old data
   - Expand disk

---

### 8. Memory Leak / OOM

**Alert:** `NodeMemoryHigh`  
**Severity:** SEV-3

#### Symptoms
- Memory usage > 90%
- OOM kills in logs
- Service restarts

#### Quick Diagnosis

```bash
# 1. Check memory usage
free -h
ps aux --sort=-%mem | head -10

# 2. Check for OOM kills
dmesg | grep -i "oom\|killed" | tail -20

# 3. Check service memory
systemctl status virtengine
cat /proc/$(pidof virtengine)/status | grep -E "(VmRSS|VmSize)"

# 4. Check Go memory stats (if pprof enabled)
curl http://localhost:6060/debug/pprof/heap > heap.pprof
go tool pprof heap.pprof
```

#### Resolution Steps

1. **Immediate mitigation:**
   ```bash
   # Restart service to free memory
   systemctl restart virtengine
   
   # Or gracefully restart with drain
   kubectl drain node-1 --ignore-daemonsets
   kubectl uncordon node-1
   ```

2. **Long-term fix:**
   - Identify memory leak
   - Update Go garbage collection settings
   - Increase memory limits
   - Add memory monitoring

---

## Incident Response Checklist

### When Alert Fires

- [ ] Acknowledge alert within 5 minutes
- [ ] Check relevant dashboard
- [ ] Determine severity level
- [ ] Create incident Slack channel (if SEV-1/SEV-2)
- [ ] Notify stakeholders as needed

### During Incident

- [ ] Update incident channel every 15 minutes
- [ ] Document actions taken
- [ ] Request help if needed
- [ ] Keep status page updated

### After Resolution

- [ ] Confirm service is restored
- [ ] Update status page to resolved
- [ ] Schedule postmortem (within 48 hours)
- [ ] Create action items
- [ ] Update this runbook if needed

---

## Escalation Contacts

| Level | When to Escalate | Contact |
|-------|------------------|---------|
| Level 1 | Can't diagnose in 15 min | Secondary on-call |
| Level 2 | Can't resolve in 1 hour | Engineering lead |
| Level 3 | Customer impact > 2 hours | VP Engineering |
| Security | Any security concern | security@virtengine.com |

---

## Useful Links

- [SLI/SLO/SLA Documentation](../sre/SLI_SLO_SLA.md)
- [Incident Response Process](../sre/INCIDENT_RESPONSE.md)
- [Error Code Reference](../errors/ERROR_CODES.md)
- [Architecture Overview](../../_docs/architecture.md)
- [Playbooks and SLOs](../../_docs/slos-and-playbooks.md)

---

## Appendix: Log Queries

### Loki Queries

```logql
# All errors in last hour
{job="virtengine-node"} |= "error" | json | level="error"

# VEID scoring errors
{job="virtengine-node"} |= "veid" |= "error" | json

# Provider daemon deployment failures
{job="provider-daemon"} |= "deployment" |= "failed" | json

# HPC job failures
{job="virtengine-node"} |= "hpc" |= "job_failed" | json

# Panic/crash events
{job=~"virtengine.*"} |~ "panic|fatal|crash"
```

### Prometheus Queries

```promql
# Error rate by module
sum(rate(virtengine_errors_total[5m])) by (module)

# SLO compliance
1 - (sum(rate(virtengine_errors_total[28d])) / sum(rate(virtengine_requests_total[28d])))

# Error budget burn rate
(rate(virtengine_errors_total[1h]) / rate(virtengine_requests_total[1h])) / (1 - 0.999)

# Top error codes
topk(10, sum(rate(virtengine_errors_total[5m])) by (code))
```

---

**Document Owner:** SRE Team  
**Next Review:** 2026-04-30
