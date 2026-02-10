# VirtEngine SLOs and Incident Playbooks

**Version:** 1.0.0  
**Date:** 2026-01-24  
**Task Reference:** VE-804

---

## Table of Contents

1. [Service Level Objectives](#service-level-objectives)
2. [Incident Playbooks](#incident-playbooks)
3. [On-Call Procedures](#on-call-procedures)
4. [Communication Templates](#communication-templates)

---

## Service Level Objectives

### SLO Summary

| Service | Metric | Target | Measurement Window |
|---------|--------|--------|-------------------|
| Chain Availability | Uptime | 99.9% | 30-day rolling |
| Chain Availability | Block time | < 6 seconds | P99 hourly |
| Identity Scoring | Processing time | < 5 minutes | P95 |
| Identity Scoring | Availability | 99.5% | 30-day rolling |
| Marketplace | Order fulfillment | < 10 minutes | P95 |
| Marketplace | Availability | 99.5% | 30-day rolling |
| HPC Scheduling | Job scheduling | < 15 minutes | P95 |
| HPC Scheduling | Availability | 99.0% | 30-day rolling |
| Provider Daemon | API availability | 99.5% | 30-day rolling |
| Provider Daemon | Health check latency | < 30 seconds | P99 |

---

### Chain Availability

#### SLO-CHAIN-001: Block Production Uptime

- **Objective**: Chain produces blocks 99.9% of the time
- **Measurement**: `(successful_blocks / expected_blocks) * 100`
- **Target**: ≥ 99.9%
- **Window**: 30-day rolling
- **Error Budget**: 43.2 minutes/month

```promql
# PromQL for SLO measurement
1 - (
  sum(rate(tendermint_consensus_rounds_total{result="timeout"}[30d]))
  /
  sum(rate(tendermint_consensus_rounds_total[30d]))
)
```

#### SLO-CHAIN-002: Block Time P99

- **Objective**: 99th percentile block time under 6 seconds
- **Measurement**: Histogram of block production time
- **Target**: P99 ≤ 6 seconds
- **Window**: Hourly

```promql
histogram_quantile(0.99,
  sum(rate(tendermint_consensus_block_time_seconds_bucket[1h])) by (le)
)
```

#### SLO-CHAIN-003: Transaction Finality

- **Objective**: Transactions finalize within 2 blocks
- **Measurement**: Blocks between submission and inclusion
- **Target**: P99 ≤ 2 blocks
- **Window**: Daily

---

### Identity Scoring

#### SLO-VEID-001: Scoring Processing Time

- **Objective**: Identity scores computed within 5 minutes
- **Measurement**: Time from upload to score availability
- **Target**: P95 ≤ 5 minutes
- **Window**: Daily

```promql
histogram_quantile(0.95,
  sum(rate(veid_scoring_duration_seconds_bucket[24h])) by (le)
)
```

#### SLO-VEID-002: Scoring Availability

- **Objective**: Identity scoring service available 99.5% of the time
- **Measurement**: Successful score requests / total requests
- **Target**: ≥ 99.5%
- **Window**: 30-day rolling
- **Error Budget**: 3.6 hours/month

#### SLO-VEID-003: Score Accuracy

- **Objective**: ML model maintains accuracy baseline
- **Measurement**: Weekly test dataset evaluation
- **Target**: F1 ≥ 0.95 on verification dataset
- **Window**: Weekly

---

### Marketplace

#### SLO-MARKET-001: Order Fulfillment Time

- **Objective**: Orders receive allocation within 10 minutes
- **Measurement**: Time from order creation to allocation
- **Target**: P95 ≤ 10 minutes
- **Window**: Daily

```promql
histogram_quantile(0.95,
  sum(rate(market_order_fulfillment_seconds_bucket[24h])) by (le)
)
```

#### SLO-MARKET-002: Marketplace Availability

- **Objective**: Marketplace operations available 99.5% of the time
- **Measurement**: Successful operations / total operations
- **Target**: ≥ 99.5%
- **Window**: 30-day rolling

#### SLO-MARKET-003: Bid Response Time

- **Objective**: Provider bids received within 2 minutes
- **Measurement**: Time from order broadcast to first bid
- **Target**: P90 ≤ 2 minutes
- **Window**: Daily

---

### HPC Scheduling

#### SLO-HPC-001: Job Scheduling Time

- **Objective**: Jobs scheduled within 15 minutes
- **Measurement**: Time from submission to scheduled state
- **Target**: P95 ≤ 15 minutes
- **Window**: Daily

```promql
histogram_quantile(0.95,
  sum(rate(hpc_job_scheduling_seconds_bucket[24h])) by (le)
)
```

#### SLO-HPC-002: HPC Availability

- **Objective**: HPC submission available 99.0% of the time
- **Measurement**: Successful submissions / total submissions
- **Target**: ≥ 99.0%
- **Window**: 30-day rolling
- **Error Budget**: 7.2 hours/month

#### SLO-HPC-003: Job Completion Reliability

- **Objective**: Jobs complete without infrastructure failures
- **Measurement**: Jobs succeeded / jobs started
- **Target**: ≥ 99.0% (excluding user errors)
- **Window**: 30-day rolling

---

## Incident Playbooks

### INC-001: Chain Halted

**Severity**: Critical  
**Impact**: All chain operations unavailable  
**Detection**: Alertmanager fires `ChainHalted` alert

#### Symptoms
- No new blocks produced for > 30 seconds
- Tendermint consensus rounds timing out
- All RPC queries return stale data

#### Diagnosis Steps

```bash
# 1. Check validator status
curl http://localhost:26657/status | jq '.result.sync_info'

# 2. Check consensus state
curl http://localhost:26657/dump_consensus_state | jq '.result.round_state'

# 3. Check validator connectivity
virtengine query staking validators --status bonded | grep -c 'status: BOND_STATUS_BONDED'

# 4. Check for crash logs
journalctl -u virtengine -n 100 --no-pager
```

#### Resolution Steps

1. **If < 1/3 validators offline**:
   - Coordinate with validator operators
   - Restart affected validators
   - Chain should resume automatically

2. **If ≥ 1/3 validators offline**:
   - Emergency validator coordination (Discord/Telegram)
   - Identify root cause (network, bug, attack)
   - Coordinate simultaneous restart if needed

3. **If chain state corrupted**:
   - Stop all validators
   - Identify last good block height
   - Rollback to good state: `virtengine rollback`
   - Coordinate restart from same height

#### Post-Incident
- RCA within 48 hours
- Update runbooks with findings
- Evaluate monitoring gaps

---

### INC-002: Identity Scoring Degraded

**Severity**: High  
**Impact**: New identity verifications delayed  
**Detection**: Alertmanager fires `VEIDScoringLatencyHigh` or `VEIDScoringErrorRate` alert

#### Symptoms
- Identity score requests timing out
- Score computation taking > 5 minutes
- High error rate on scoring transactions

#### Diagnosis Steps

```bash
# 1. Check validator ML model status
virtengine query veid model-status

# 2. Check scoring queue depth
virtengine query veid pending-scores --count

# 3. Check validator consensus on scores
virtengine query veid recent-scores --limit 10 | jq '.scores[].validators'

# 4. Check model inference latency
curl http://localhost:9090/api/v1/query?query=veid_inference_duration_seconds
```

#### Resolution Steps

1. **If single validator issue**:
   - Validator should restart ML model container
   - Remove validator from scoring if repeated issues

2. **If systemic latency**:
   - Check for model version skew
   - Verify all validators on same model hash
   - Consider scaling inference resources

3. **If queue backlog**:
   - Temporarily increase validator scoring capacity
   - Consider rate limiting new submissions

---

### INC-003: Marketplace Order Fulfillment Stalled

**Severity**: High  
**Impact**: Customers unable to provision resources  
**Detection**: Alertmanager fires `MarketOrderBacklog` alert

#### Symptoms
- Orders stuck in PENDING state
- No bids received for orders
- Provider daemon errors

#### Diagnosis Steps

```bash
# 1. Check pending orders
virtengine query market orders --state pending --limit 50

# 2. Check active providers
virtengine query provider list --status active | wc -l

# 3. Check provider daemon health
curl http://provider-daemon:8443/health

# 4. Check bid engine logs
kubectl logs -l app=provider-daemon --tail=100 | grep -i "bid\|error"
```

#### Resolution Steps

1. **If no providers online**:
   - Contact provider operators
   - Check for network/chain issues affecting providers
   - Consider temporary marketplace pause

2. **If providers online but not bidding**:
   - Check bid engine configuration
   - Verify offering matches available capacity
   - Review bid filtering rules

3. **If orders malformed**:
   - Identify pattern in failing orders
   - Notify affected customers
   - Consider platform-side fix if SDK bug

---

### INC-004: HPC Job Failures Elevated

**Severity**: Medium  
**Impact**: Increased job failures for customers  
**Detection**: Alertmanager fires `HPCJobFailureRateHigh` alert

#### Symptoms
- Job failure rate above baseline
- Jobs failing at specific providers
- Resource allocation errors

#### Diagnosis Steps

```bash
# 1. Check recent job failures
virtengine query hpc jobs --state failed --limit 20

# 2. Analyze failure reasons
virtengine query hpc jobs --state failed --limit 100 | jq '.jobs[].failure_reason' | sort | uniq -c | sort -rn

# 3. Check provider-specific failure rates
virtengine query hpc jobs --state failed --provider <addr>

# 4. Check SLURM/K8s cluster status
kubectl get nodes -o wide
squeue -a --format="%.18i %.9P %.8j %.8u %.8T %.10M %.9l %.6D %R"
```

#### Resolution Steps

1. **If provider-specific**:
   - Notify provider operator
   - Consider temporary provider delist
   - Route jobs to healthy providers

2. **If resource exhaustion**:
   - Scale provider capacity
   - Adjust job scheduling priorities
   - Implement better capacity planning

3. **If user error patterns**:
   - Improve input validation
   - Update documentation
   - Add helpful error messages

---

### INC-005: Encryption Key Compromise Suspected

**Severity**: Critical  
**Impact**: Potential data confidentiality breach  
**Detection**: Manual report or anomaly detection

#### Symptoms
- Unauthorized decryption attempts logged
- Key usage from unexpected IP addresses
- User/provider reports suspicious activity

#### Immediate Actions

```bash
# 1. Revoke compromised key immediately
virtengine tx encryption revoke-key \
    --fingerprint <compromised_key_fingerprint> \
    --reason "Security incident" \
    --from admin

# 2. Notify affected parties
# Use communication template below

# 3. Audit key usage
virtengine query encryption key-usage-log \
    --fingerprint <key_fingerprint> \
    --since "2026-01-01T00:00:00Z"
```

#### Investigation Steps

1. **Identify scope**:
   - What data was encrypted to this key?
   - What timeframe of potential exposure?
   - What transactions involved this key?

2. **Trace source**:
   - Review access logs
   - Check for malware/phishing
   - Interview key holder

3. **Contain impact**:
   - Re-encrypt affected data with new keys
   - Reset any derived credentials
   - Force re-authentication for affected users

---

## On-Call Procedures

### On-Call Rotation

- **Primary**: First responder, 15-minute response SLA
- **Secondary**: Backup if primary unavailable, 30-minute response
- **Escalation**: Engineering lead for Critical severity

### Escalation Path

```
Alert Fired
    │
    ▼
Primary On-Call (15 min)
    │
    ├── Acknowledged & Handling → Continue
    │
    └── No Response (15 min)
            │
            ▼
        Secondary On-Call (30 min)
            │
            ├── Acknowledged & Handling → Continue
            │
            └── No Response (30 min)
                    │
                    ▼
                Engineering Lead
                    │
                    ▼
                Incident Commander (if Critical)
```

### Severity Definitions

| Severity | Impact | Response Time | Examples |
|----------|--------|---------------|----------|
| Critical | Service completely unavailable | 15 min | Chain halt, data breach |
| High | Major feature unavailable | 1 hour | Identity scoring down, marketplace stalled |
| Medium | Degraded service | 4 hours | Elevated error rates, slow performance |
| Low | Minor issue | 24 hours | Non-critical bugs, cosmetic issues |

### Incident Commander Checklist

- [ ] Acknowledge incident in alert system
- [ ] Create incident channel (#inc-YYYYMMDD-description)
- [ ] Assign roles: IC, Tech Lead, Comms Lead
- [ ] Set incident timeline: check-ins every 15-30 min
- [ ] Post initial status update
- [ ] Coordinate resolution
- [ ] Confirm service restored
- [ ] Schedule post-incident review
- [ ] Complete RCA document

---

## Communication Templates

### Status Page Update - Investigating

```
[INVESTIGATING] <Service> Degradation

We are investigating reports of degraded performance for <service>.

Impact: <description of user impact>
Start Time: <HH:MM UTC>

We will provide updates every 30 minutes or as new information becomes available.
```

### Status Page Update - Identified

```
[IDENTIFIED] <Service> Degradation

We have identified the cause of the <service> degradation and are implementing a fix.

Root Cause: <brief technical summary>
Impact: <description of user impact>
ETA: <estimated resolution time>

We will update when the fix is deployed.
```

### Status Page Update - Resolved

```
[RESOLVED] <Service> Degradation

The <service> degradation has been resolved. All systems are operating normally.

Duration: <start time> to <end time> (<X> hours <Y> minutes)
Impact: <description of what was affected>
Resolution: <brief summary of fix>

We apologize for any inconvenience. A detailed post-mortem will be published within 48 hours.
```

### User Notification - Security Incident

```
Subject: Important Security Notice - Action Required

Dear VirtEngine User,

We are writing to inform you of a security incident that may affect your account.

What Happened:
<Clear description of the incident>

What Information Was Involved:
<List of affected data types>

What We Are Doing:
<Actions taken to address the incident>

What You Should Do:
1. <Specific action item>
2. <Specific action item>
3. <Specific action item>

For More Information:
Contact our security team at security@virtengine.com

We take your security seriously and apologize for any concern this may cause.

The VirtEngine Security Team
```

---

## Appendix: Alert Definitions

```yaml
# Prometheus AlertManager rules
groups:
  - name: chain
    rules:
      - alert: ChainHalted
        expr: time() - tendermint_consensus_latest_block_time > 30
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Chain halted - no new blocks"
          
      - alert: ChainBlockTimeSlow
        expr: histogram_quantile(0.99, tendermint_consensus_block_time_seconds_bucket) > 6
        for: 5m
        labels:
          severity: high

  - name: veid
    rules:
      - alert: VEIDScoringLatencyHigh
        expr: histogram_quantile(0.95, veid_scoring_duration_seconds_bucket) > 300
        for: 5m
        labels:
          severity: high
          
      - alert: VEIDScoringErrorRate
        expr: rate(veid_scoring_errors_total[5m]) > 0.1
        for: 5m
        labels:
          severity: high

  - name: market
    rules:
      - alert: MarketOrderBacklog
        expr: market_orders_pending > 100
        for: 10m
        labels:
          severity: high

  - name: hpc
    rules:
      - alert: HPCJobFailureRateHigh
        expr: rate(hpc_jobs_failed_total[1h]) / rate(hpc_jobs_total[1h]) > 0.1
        for: 15m
        labels:
          severity: medium
```
