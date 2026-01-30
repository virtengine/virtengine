# VirtEngine On-Call Runbooks

This directory contains operational runbooks for VirtEngine on-call engineers. Each runbook provides step-by-step procedures for handling common alerts and incidents.

## Quick Reference

| Alert | Severity | Runbook | Owner |
|-------|----------|---------|-------|
| NodeDown | Critical | [node-down.md](node-down.md) | Platform Team |
| BlockProductionStalled | Critical | [block-stalled.md](block-stalled.md) | Platform Team |
| LowValidatorCount | Critical | [low-validators.md](low-validators.md) | Platform Team |
| VEIDInferenceNonDeterministic | Critical | [veid-non-deterministic.md](veid-non-deterministic.md) | ML Team |
| HighErrorRate | Warning | [high-error-rate.md](high-error-rate.md) | Platform Team |
| SLOBudgetBurning | Warning | [slo-budget-burning.md](slo-budget-burning.md) | SRE Team |
| ProviderDeploymentFailures | Warning | [provider-deployment.md](provider-deployment.md) | Provider Team |

## On-Call Responsibilities

### Primary On-Call
- Respond to pages within 15 minutes
- Acknowledge all alerts in PagerDuty/Slack
- Follow runbook procedures
- Escalate as needed
- Document incident timeline

### Secondary On-Call
- Backup for primary
- Available for escalation
- Shadow for training

## Escalation Path

1. **L1 - Primary On-Call**: First responder, follows runbooks
2. **L2 - Secondary On-Call**: Complex issues requiring additional expertise
3. **L3 - Team Lead/Manager**: Policy decisions, customer communication
4. **L4 - Executive**: Major outages, external communication

## Contact Information

| Role | Primary | Secondary |
|------|---------|-----------|
| Platform On-Call | @platform-oncall | [PagerDuty Schedule] |
| ML Team | @ml-team | [Slack: #ml-alerts] |
| SRE Team | @sre-team | [Slack: #sre-alerts] |
| Security | @security-team | [security@virtengine.io] |

## Before You Start

1. Ensure you have access to:
   - Grafana dashboards
   - Prometheus/Alertmanager
   - SSH access to nodes
   - Kubectl/infrastructure access
   - Log aggregation (Loki)
   - Tracing (Tempo)

2. Bookmark these dashboards:
   - [Chain Health](http://grafana:3000/d/chain-health)
   - [SLO Overview](http://grafana:3000/d/slo-overview)
   - [VEID Dashboard](http://grafana:3000/d/veid)
   - [Provider Dashboard](http://grafana:3000/d/provider)

## Incident Response Process

### 1. Acknowledge
- Acknowledge the alert within 15 minutes
- Post in #incidents channel with initial assessment

### 2. Assess
- Determine scope and impact
- Identify affected services
- Check for related alerts

### 3. Mitigate
- Follow runbook procedures
- Prioritize service restoration
- Document actions taken

### 4. Communicate
- Update status page (if customer-facing)
- Notify stakeholders
- Regular updates every 30 minutes

### 5. Resolve
- Verify service restoration
- Close alert/incident
- Schedule postmortem if needed

## Common Commands

### Check Node Status
```bash
# SSH to node
ssh user@virtengine-node

# Check service status
systemctl status virtengined

# Check recent logs
journalctl -u virtengined -n 100

# Check block height
virtengined status | jq '.SyncInfo.latest_block_height'
```

### Check Validator Status
```bash
# List validators
virtengined q staking validators

# Check validator signing info
virtengined q slashing signing-info $(virtengined tendermint show-validator)
```

### Provider Daemon Commands
```bash
# Check provider status
provider-daemon status

# List active deployments
provider-daemon list deployments

# Check bid engine
provider-daemon bid-engine status
```

### Kubernetes Commands
```bash
# List pods
kubectl get pods -n virtengine

# Check pod logs
kubectl logs -n virtengine <pod-name>

# Describe pod
kubectl describe pod -n virtengine <pod-name>

# Restart deployment
kubectl rollout restart deployment -n virtengine <deployment-name>
```

## Useful Queries

### Prometheus

```promql
# Error rate by service
sum(rate(virtengine_errors_total[5m])) by (service)

# P95 latency
histogram_quantile(0.95, sum(rate(virtengine_api_request_duration_seconds_bucket[5m])) by (le))

# Error budget consumption
1 - (sum(rate(virtengine_api_requests_total{status!~"5.."}[28d])) / sum(rate(virtengine_api_requests_total[28d])))
```

### Loki

```logql
# Errors in last hour
{service="virtengine-node"} |= "ERROR" | json | line_format "{{.msg}}"

# Trace correlation
{service="virtengine-node"} | json | trace_id != ""

# Slow queries
{service="virtengine-node"} |= "slow_query" | json | duration > 5s
```

## Index

- [node-down.md](node-down.md) - Node is unresponsive
- [block-stalled.md](block-stalled.md) - Block production stopped
- [low-validators.md](low-validators.md) - Insufficient validators
- [veid-non-deterministic.md](veid-non-deterministic.md) - ML inference mismatch
- [high-error-rate.md](high-error-rate.md) - Elevated error rates
- [slo-budget-burning.md](slo-budget-burning.md) - SLO budget depletion
- [provider-deployment.md](provider-deployment.md) - Deployment failures
- [consensus-failure.md](consensus-failure.md) - Consensus issues
- [network-partition.md](network-partition.md) - P2P network issues
- [database-issues.md](database-issues.md) - State database problems
