# HPC Routing Policy and Fallback Rules

This document describes the on-chain scheduling enforcement for HPC job placement in VirtEngine.

## Overview

VE-5B implements routing enforcement to guarantee that jobs are placed according to x/hpc scheduling decisions. This ensures:

1. **Deterministic Routing**: Jobs are routed to clusters based on on-chain scheduling decisions
2. **Audit Trail**: All routing decisions and fallbacks are recorded for accountability
3. **Fail-Closed Behavior**: In strict mode, jobs cannot be placed without valid routing decisions
4. **Explicit Fallbacks**: When fallback routing is used, it is logged and auditable

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Job Submission │───▶│ Routing Enforcer │───▶│  HPC Scheduler  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                     ┌──────────────────┐
                     │ x/hpc Scheduling │
                     │    Decision      │
                     └──────────────────┘
```

## Enforcement Modes

### Strict Mode (Default)

- Jobs **must** have a valid scheduling decision
- Jobs are rejected if the scheduled cluster is unavailable
- Cluster mismatch results in job rejection
- Stale decisions require re-scheduling

```yaml
routing:
  enforcement_mode: "strict"
  require_decision_for_submission: true
  allow_automatic_fallback: false
```

### Permissive Mode

- Jobs with missing decisions can proceed with new scheduling
- Automatic fallback to alternative clusters is allowed
- Stale decisions are automatically refreshed
- Violations are logged but jobs are not rejected

```yaml
routing:
  enforcement_mode: "permissive"
  require_decision_for_submission: false
  allow_automatic_fallback: true
```

### Audit-Only Mode

- No enforcement is applied
- All routing decisions and violations are logged
- Useful for monitoring before enabling enforcement

```yaml
routing:
  enforcement_mode: "audit_only"
  require_decision_for_submission: false
```

## Routing Flow

### 1. Pre-Submission Check

Before any job submission, the routing enforcer:

1. **Fetches Scheduling Decision**: Queries x/hpc for the scheduling decision
2. **Validates Cluster Availability**: Checks if the target cluster is active
3. **Validates Node Capacity**: Ensures sufficient nodes are available
4. **Checks Decision Staleness**: Verifies the decision is not too old

### 2. Decision Validation

Scheduling decisions are validated against:

- **Block Age**: Maximum age in blocks (default: 100 blocks ≈ 10 minutes)
- **Time Age**: Maximum age in seconds (default: 600 seconds)
- **Cluster Match**: Target cluster matches the scheduled cluster
- **Capacity**: Cluster has sufficient resources

### 3. Fallback Logic

When the primary cluster is unavailable:

1. **Strict Mode**: Job is rejected with `cluster_unavailable` violation
2. **Permissive Mode**: 
   - Request new scheduling decision
   - Route to alternative cluster
   - Record fallback with reason
3. **Audit-Only Mode**: Log warning, proceed with original decision

## Violation Types

| Type | Description | Severity |
|------|-------------|----------|
| `missing_decision` | No scheduling decision referenced | 4 |
| `stale_decision` | Decision is too old | 2 |
| `cluster_mismatch` | Job placed on wrong cluster | 5 |
| `cluster_unavailable` | Scheduled cluster is not available | 3 |
| `capacity_exceeded` | Insufficient cluster capacity | 3 |
| `unauthorized_fallback` | Fallback used without authorization | 4 |

## Audit Records

Every routing decision creates an audit record:

```json
{
  "record_id": "routing-audit-123",
  "job_id": "hpc-job-456",
  "scheduling_decision_id": "hpc-decision-789",
  "expected_cluster_id": "cluster-1",
  "actual_cluster_id": "cluster-1",
  "status": "approved",
  "reason": "Job routed to scheduled cluster",
  "is_fallback": false,
  "decision_age_blocks": 5,
  "decision_age_seconds": 30,
  "provider_address": "virtengine1...",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### Audit Record Status

- `approved`: Job routed as per scheduling decision
- `fallback`: Job used fallback routing (with reason)
- `rejected`: Job was rejected (in strict mode)
- `rescheduled`: Job was re-scheduled due to stale decision

## Configuration

### Provider Daemon Configuration

```yaml
hpc:
  enabled: true
  cluster_id: "hpc-cluster-1"
  provider_address: "virtengine1provider..."
  
  routing:
    enabled: true
    enforcement_mode: "strict"
    max_decision_age_blocks: 100
    max_decision_age_seconds: 600
    allow_automatic_fallback: true
    require_decision_for_submission: true
    auto_refresh_stale_decisions: true
    violation_alert_threshold: 5
```

### Module Parameters

On-chain parameters in x/hpc:

```json
{
  "routing_enforcement_enabled": true,
  "max_decision_age_blocks": 100,
  "max_decision_age_seconds": 600,
  "allow_provider_fallback": true
}
```

## Events

### Routing Events

Events emitted on the blockchain:

- `hpc_job_routed`: Job successfully routed
- `hpc_job_rescheduled`: Job was re-scheduled
- `hpc_routing_violation`: Routing violation detected
- `hpc_routing_violation_resolved`: Violation was resolved

### Audit Events

Events logged in the audit log:

- `routing_enforced`: Routing enforcement completed
- `routing_enforcement_failed`: Enforcement failed
- `routing_fallback_used`: Fallback routing was used
- `routing_decision_refreshed`: Stale decision was refreshed
- `routing_violation_threshold_exceeded`: Too many violations

## Monitoring

### Violation Monitoring

The routing enforcer tracks violations per provider:

```go
violationCount := enforcer.GetViolationCount("virtengine1provider")
```

When violations exceed the threshold (`violation_alert_threshold`), an alert event is logged.

### Metrics

Key metrics to monitor:

- `routing_enforcement_total`: Total routing enforcements
- `routing_violations_total`: Total violations by type
- `routing_fallbacks_total`: Total fallback routings
- `routing_decision_age_seconds`: Age of decisions at enforcement
- `routing_enforcement_latency_ms`: Enforcement latency

## Security Considerations

1. **Decision Integrity**: Scheduling decisions are stored on-chain and cannot be tampered with
2. **Provider Authorization**: Only authorized providers can submit jobs to their clusters
3. **Audit Trail**: All routing actions are auditable
4. **Fail-Closed**: In strict mode, security is prioritized over availability

## Best Practices

### For Providers

1. **Use Strict Mode in Production**: Ensures routing compliance
2. **Monitor Violations**: Set up alerts for violation thresholds
3. **Review Audit Logs**: Regularly review routing audit records
4. **Handle Stale Decisions**: Enable auto-refresh for stale decisions

### For Operators

1. **Set Appropriate Thresholds**: Balance security with usability
2. **Monitor Cluster Availability**: Ensure clusters are healthy
3. **Capacity Planning**: Ensure adequate node capacity
4. **Incident Response**: Have procedures for routing violations

## Troubleshooting

### Common Issues

**Job rejected with "missing scheduling decision"**

Cause: Job submitted without first obtaining a scheduling decision
Solution: Ensure jobs call `ScheduleJob` before submission

**Job rejected with "scheduling decision is stale"**

Cause: Too much time between scheduling and submission
Solution: Enable `auto_refresh_stale_decisions` or reduce submission delay

**Frequent fallback routing**

Cause: Scheduled clusters frequently unavailable
Solution: Review cluster health and capacity planning

**High violation count**

Cause: Provider configuration or cluster issues
Solution: Review provider configuration and cluster status

## API Reference

### RoutingEnforcer

```go
// Create enforcer
enforcer := NewRoutingEnforcer(config, querier, reporter, auditor)

// Enforce routing before submission
result, err := enforcer.EnforceRouting(ctx, job)

// Validate job placement
err := enforcer.ValidateJobPlacement(ctx, job, targetClusterID)

// Get violation count
count := enforcer.GetViolationCount(providerAddress)
```

### HPCSchedulingQuerier

```go
// Query scheduling decision
decision, err := querier.GetSchedulingDecision(ctx, decisionID)

// Request new decision
decision, err := querier.RequestNewSchedulingDecision(ctx, job)

// Get cluster status
status, err := querier.GetClusterStatus(ctx, clusterID)
```

## Changelog

### VE-5B (Current)

- Initial implementation of routing enforcement
- Added routing audit records and violations
- Implemented strict, permissive, and audit-only modes
- Added violation threshold alerting
- Documented routing policy and fallback rules
