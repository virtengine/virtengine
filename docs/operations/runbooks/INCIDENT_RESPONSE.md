# Incident Response Playbooks

**Version:** 1.0.0  
**Last Updated:** 2026-01-30  
**Owner:** SRE Team

---

## Table of Contents

1. [Incident Response Framework](#incident-response-framework)
2. [Severity Classification](#severity-classification)
3. [Incident Response Process](#incident-response-process)
4. [Critical Playbooks](#critical-playbooks)
5. [High Priority Playbooks](#high-priority-playbooks)
6. [Medium Priority Playbooks](#medium-priority-playbooks)
7. [Security Incident Playbooks](#security-incident-playbooks)
8. [Post-Incident Procedures](#post-incident-procedures)

---

## Incident Response Framework

### Incident Roles

| Role | Responsibility |
|------|----------------|
| **Incident Commander (IC)** | Overall incident coordination |
| **Technical Lead** | Technical diagnosis and resolution |
| **Communications Lead** | Stakeholder updates |
| **Scribe** | Document timeline and actions |

### Communication Channels

| Channel | Purpose |
|---------|---------|
| PagerDuty | Alert routing and escalation |
| Slack #incident-response | Real-time coordination |
| Slack #inc-YYYYMMDD-desc | Incident-specific channel |
| Status Page | Customer communication |

---

## Severity Classification

### SEV-1: Critical

**Definition**: Complete service outage affecting all users

**Examples**:
- Chain halted (no blocks for > 1 minute)
- Complete network partition
- Data breach confirmed
- All validators offline

**Response Requirements**:
- Response time: 5 minutes
- Page: All on-call, engineering lead
- Bridge: Start immediately
- Status page: Update within 15 minutes

### SEV-2: High

**Definition**: Major feature unavailable or severely degraded

**Examples**:
- Identity scoring completely unavailable
- Marketplace orders not processing
- Provider daemon widespread failures
- > 50% error rate on any service

**Response Requirements**:
- Response time: 15 minutes
- Page: Primary on-call
- Bridge: Within 30 minutes if not resolved
- Status page: Update within 30 minutes

### SEV-3: Medium

**Definition**: Partial degradation, workarounds available

**Examples**:
- Elevated error rates (10-50%)
- Single provider offline
- Slow query performance
- Minor feature unavailable

**Response Requirements**:
- Response time: 1 hour
- Notification: Slack alert
- Status page: Optional

### SEV-4: Low

**Definition**: Minor issues, no user impact

**Examples**:
- Cosmetic bugs
- Non-critical monitoring gaps
- Documentation issues

**Response Requirements**:
- Response time: 24 hours
- Tracking: JIRA ticket

---

## Incident Response Process

### Phase 1: Detection & Triage (0-5 minutes)

```
Alert Received
     │
     ▼
┌─────────────────┐
│ Acknowledge in  │
│ PagerDuty       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Check Dashboard │
│ & Quick Status  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Classify        │
│ Severity        │
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
  SEV-1/2   SEV-3/4
    │         │
    ▼         ▼
 Create    Work ticket
 incident  async
 channel
```

### Phase 2: Mobilization (5-15 minutes)

For SEV-1/SEV-2:

1. Create incident channel: `#inc-YYYYMMDD-<description>`
2. Post initial status:
   ```
   @here INCIDENT DECLARED
   
   Severity: SEV-1
   Service: Chain Consensus
   Impact: Chain halted, no blocks producing
   IC: @oncall-primary
   
   Status: Investigating
   ```
3. Assign roles (IC, Tech Lead, Comms)
4. Start timeline documentation

### Phase 3: Investigation (15-60 minutes)

1. Follow relevant playbook
2. Document all actions in incident channel
3. Update status every 15 minutes (SEV-1) or 30 minutes (SEV-2)
4. Escalate if needed

### Phase 4: Resolution

1. Implement fix
2. Verify fix with multiple team members
3. Monitor for recurrence
4. Communicate resolution

### Phase 5: Post-Incident

1. Confirm all systems stable
2. Update status page to resolved
3. Schedule post-mortem (within 48 hours)
4. Create follow-up tickets

---

## Critical Playbooks

### PLAY-001: Chain Halted

**Severity**: SEV-1  
**Alert**: `ChainHalted`  
**Impact**: All blockchain operations unavailable

#### Detection

```bash
# Check last block time
curl -s http://localhost:26657/status | jq '.result.sync_info.latest_block_time'

# If > 30 seconds since last block, chain is halted
```

#### Immediate Actions

1. **Acknowledge alert and declare incident**
2. **Check consensus state**

```bash
# Dump consensus state
curl -s http://localhost:26657/dump_consensus_state | jq '.result.round_state'

# Check validator votes
curl -s http://localhost:26657/dump_consensus_state | jq '.result.round_state.height_vote_set'
```

3. **Identify affected validators**

```bash
# Check missing validators
virtengine query slashing signing-infos --limit 100 | grep -B5 "missed_blocks_counter: [1-9]"

# Check validator connectivity
for validator in validator-1 validator-2 validator-3; do
  echo "$validator: $(curl -s http://$validator:26657/status | jq -r '.result.sync_info.catching_up')"
done
```

#### Resolution Steps

**Scenario A: < 1/3 validators offline**

1. Contact offline validator operators (Discord, Telegram)
2. Wait for validators to restart
3. Chain will resume automatically once quorum restored

**Scenario B: ≥ 1/3 validators offline**

1. **Emergency coordination required**
   - Post in #validator-emergency
   - Contact validator operators via all channels
   
2. **Identify root cause**
   - Network partition?
   - Common infrastructure failure?
   - Software bug?
   - Coordinated attack?

3. **Coordinated restart if needed**
   ```bash
   # All validators simultaneously:
   sudo systemctl restart virtengine
   ```

**Scenario C: State corruption**

1. **Stop all validators**
   ```bash
   sudo systemctl stop virtengine
   ```

2. **Identify last good block height**
   ```bash
   # Check for consensus failures in logs
   journalctl -u virtengine | grep -E "CONSENSUS FAILURE|panic" | tail -20
   ```

3. **Rollback to good state**
   ```bash
   # On each validator
   virtengine rollback --hard
   ```

4. **Coordinate restart at same height**
   - Verify all validators at same height
   - Restart simultaneously

#### Verification

```bash
# Verify blocks producing
watch -n 5 'curl -s http://localhost:26657/status | jq ".result.sync_info.latest_block_height"'

# Verify consensus
curl -s http://localhost:26657/consensus_state | jq '.result.round_state.height'
```

#### Escalation

- If not resolved in 15 minutes: Page engineering lead
- If not resolved in 30 minutes: Emergency all-hands

---

### PLAY-002: Complete Network Partition

**Severity**: SEV-1  
**Alert**: `NetworkPartition`, `PeerCountCritical`  
**Impact**: Validators cannot communicate, potential chain split

#### Detection

```bash
# Check peer counts across validators
for v in validator-{1..10}; do
  echo "$v: $(curl -s http://$v:26657/net_info | jq '.result.n_peers')"
done

# Check for two or more groups with different heights
```

#### Immediate Actions

1. **Identify partition topology**
   - Which validators can communicate?
   - Are there two separate "islands"?

2. **Check network infrastructure**
   ```bash
   # From each validator, check connectivity
   nc -zv validator-2 26656
   traceroute validator-2
   ```

#### Resolution Steps

1. **If infrastructure issue (AWS, GCP, etc.)**
   - Check cloud provider status page
   - Contact cloud support
   - Route traffic through backup path if available

2. **If configuration issue**
   ```bash
   # Verify peer configuration
   cat ~/.virtengine/config/config.toml | grep -A10 "\[p2p\]"
   
   # Force peer reconnection
   virtengine unsafe-reset-peers
   sudo systemctl restart virtengine
   ```

3. **If chain split occurred**
   - Identify canonical chain (highest voting power)
   - Minority chain validators must:
     ```bash
     sudo systemctl stop virtengine
     virtengine tendermint unsafe-reset-all
     # Resync from canonical chain
     ```

#### Verification

```bash
# All validators should have same height and hash
for v in validator-{1..10}; do
  curl -s http://$v:26657/status | jq -r '"\(.result.validator_info.address): \(.result.sync_info.latest_block_height) \(.result.sync_info.latest_block_hash)"'
done
```

---

### PLAY-003: Data Breach / Security Incident

**Severity**: SEV-1  
**Alert**: Manual report or anomaly detection  
**Impact**: Potential data confidentiality or integrity compromise

#### Immediate Actions (First 5 minutes)

1. **Do NOT make changes that destroy evidence**
2. **Page security team immediately**: security@virtengine.com
3. **Document what you know**

```markdown
## Initial Breach Report

Time detected: [TIMESTAMP]
Detected by: [WHO/WHAT]
Affected systems: [LIST]
Initial observations: [DETAILS]
```

#### Containment Actions

**Only with security team approval:**

1. **Isolate affected systems**
   ```bash
   # Firewall isolation
   sudo ufw deny from any to any
   sudo ufw allow from SECURITY_TEAM_IP
   ```

2. **Revoke compromised credentials**
   ```bash
   # If key compromised
   virtengine tx encryption revoke-key \
       --fingerprint <compromised_key> \
       --reason "Security incident" \
       --from admin
   ```

3. **Capture evidence**
   ```bash
   # Preserve logs
   tar -czf /secure/evidence-$(date +%Y%m%d-%H%M).tar.gz \
       /var/log/virtengine/ \
       /var/log/provider-daemon/ \
       /var/log/auth.log
   
   # Memory dump if needed
   gcore $(pidof virtengine)
   ```

#### Investigation (Security team leads)

1. Determine scope of breach
2. Identify attack vector
3. Assess data impact
4. Determine notification requirements

#### Recovery

1. Rebuild affected systems from known-good images
2. Rotate all potentially compromised credentials
3. Implement additional monitoring
4. Complete incident report

---

## High Priority Playbooks

### PLAY-010: Identity Scoring Unavailable

**Severity**: SEV-2  
**Alert**: `VEIDScoringDown`, `VEIDMLInferenceFailureRate`  
**Impact**: New identity verifications cannot be processed

#### Detection

```bash
# Check scoring metrics
curl -s http://localhost:26660/metrics | grep veid_scoring

# Check pending scores
virtengine query veid pending-scores --count
```

#### Diagnosis

```bash
# 1. Check model status on validators
for v in validator-{1..5}; do
  echo "$v: $(ssh $v 'virtengine query veid model-status')"
done

# 2. Check ML inference service
curl -s http://localhost:8080/health

# 3. Check for model version mismatch
for v in validator-{1..5}; do
  echo "$v: $(ssh $v 'sha256sum ~/.virtengine/models/veid_scorer_*.h5')"
done

# 4. Check GPU status (if applicable)
nvidia-smi
```

#### Resolution

**If single validator issue:**
```bash
# Restart ML service
ssh validator-X 'sudo systemctl restart ml-inference'

# Verify recovery
ssh validator-X 'virtengine query veid model-status'
```

**If model version mismatch:**
```bash
# Download correct model version
wget -O ~/.virtengine/models/veid_scorer_v1.0.0.h5 \
    https://models.virtengine.com/veid_scorer_v1.0.0.h5

# Verify hash
sha256sum ~/.virtengine/models/veid_scorer_v1.0.0.h5

# Restart validator
sudo systemctl restart virtengine
```

**If systemic issue:**
```bash
# Check for governance model update
virtengine query upgrade plan

# If new model required, coordinate upgrade
```

---

### PLAY-011: Marketplace Orders Not Processing

**Severity**: SEV-2  
**Alert**: `MarketOrderBacklog`, `NoBidsReceived`  
**Impact**: Customers cannot provision resources

#### Detection

```bash
# Check pending orders
virtengine query market orders --state pending --limit 100 | wc -l

# Check order age
virtengine query market orders --state pending --limit 10 | jq '.orders[].created_at'
```

#### Diagnosis

```bash
# 1. Check active providers
virtengine query provider list --status active | wc -l

# 2. Check provider daemon health
for p in provider-{1..5}; do
  echo "$p: $(curl -sk https://$p:8443/health)"
done

# 3. Check bid engine status
for p in provider-{1..5}; do
  echo "$p: $(ssh $p 'provider-daemon bids status')"
done

# 4. Check recent bids
virtengine query market bids --limit 20
```

#### Resolution

**If no providers online:**
- Contact provider operators
- Check for common infrastructure issues

**If providers not bidding:**
```bash
# Check provider config
ssh provider-X 'cat ~/.provider-daemon/config.yaml | grep -A20 bidding'

# Restart bid engine
ssh provider-X 'sudo systemctl restart provider-daemon'
```

**If orders malformed:**
```bash
# Analyze failing orders
virtengine query market orders --state pending --limit 10 | jq '.orders[].spec'

# If SDK bug, notify customers with workaround
```

---

### PLAY-012: Provider Daemon Mass Failure

**Severity**: SEV-2  
**Alert**: `ProviderDaemonDown`, `ProviderDaemonErrorRateHigh`  
**Impact**: Resource provisioning affected

#### Detection

```bash
# Check provider daemon status
for p in provider-{1..10}; do
  echo "$p: $(curl -sk https://$p:8443/health || echo 'DOWN')"
done
```

#### Diagnosis

```bash
# 1. Check common error pattern
for p in provider-{1..5}; do
  echo "=== $p ==="
  ssh $p 'journalctl -u provider-daemon --since "30 min ago" | grep -i error | tail -5'
done

# 2. Check chain connectivity from providers
for p in provider-{1..5}; do
  ssh $p 'curl -s http://localhost:26657/status | jq .result.sync_info.catching_up'
done

# 3. Check for common root cause
# - Chain upgrade?
# - API change?
# - Configuration push?
```

#### Resolution

**If chain connectivity issue:**
```bash
# Verify RPC endpoints
curl -s http://rpc.virtengine.com:26657/status

# Update provider configs if needed
```

**If software bug:**
```bash
# Rollback to previous version
ssh provider-X 'sudo systemctl stop provider-daemon'
ssh provider-X 'sudo cp /usr/local/bin/provider-daemon.bak /usr/local/bin/provider-daemon'
ssh provider-X 'sudo systemctl start provider-daemon'
```

---

## Medium Priority Playbooks

### PLAY-020: Elevated Error Rates

**Severity**: SEV-3  
**Alert**: `HighErrorRate`  
**Impact**: Degraded user experience

#### Diagnosis

```bash
# 1. Identify error distribution
curl -s http://localhost:9090/api/v1/query?query=topk(10,sum(rate(virtengine_errors_total[5m]))by(module,code)) | jq

# 2. Check recent changes
git log --oneline -10
kubectl rollout history deployment/virtengine

# 3. Correlate with infrastructure events
# Check cloud provider dashboards
```

#### Resolution

Depends on root cause:
- **Deployment issue**: Rollback
- **Resource exhaustion**: Scale up
- **External dependency**: Failover or wait

---

### PLAY-021: Node Out of Sync

**Severity**: SEV-3  
**Alert**: `NodeOutOfSync`, `NodeBehind`  
**Impact**: Stale data for users on affected node

#### Diagnosis

```bash
# Check how far behind
NETWORK=$(curl -s https://rpc.virtengine.com/status | jq -r '.result.sync_info.latest_block_height')
LOCAL=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
echo "Behind by $((NETWORK - LOCAL)) blocks"

# Check peer connectivity
curl -s http://localhost:26657/net_info | jq '.result.n_peers'

# Check disk space
df -h ~/.virtengine/data
```

#### Resolution

**If < 100 blocks behind:**
- Usually catches up automatically
- Monitor for 10 minutes

**If significantly behind:**
```bash
# Option 1: Wait for catch-up (may take hours)

# Option 2: State sync
sudo systemctl stop virtengine
# Configure state-sync in config.toml
sudo systemctl start virtengine

# Option 3: Restore from snapshot
sudo systemctl stop virtengine
wget https://snapshots.virtengine.com/latest.tar.lz4
lz4 -d latest.tar.lz4 | tar -xf - -C ~/.virtengine/data/
sudo systemctl start virtengine
```

---

### PLAY-022: Database Connection Issues

**Severity**: SEV-3  
**Alert**: `PostgresConnectionPoolExhausted`  
**Impact**: Queries failing for services using PostgreSQL

#### Diagnosis

```bash
# Check connection count
psql -c "SELECT count(*) FROM pg_stat_activity;"

# Check by state
psql -c "SELECT state, count(*) FROM pg_stat_activity GROUP BY state;"

# Find long-running queries
psql -c "SELECT pid, now() - query_start AS duration, query 
         FROM pg_stat_activity 
         WHERE state = 'active' AND now() - query_start > interval '1 minute'
         ORDER BY duration DESC LIMIT 10;"
```

#### Resolution

```bash
# 1. Kill long-running queries
psql -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity 
         WHERE state = 'active' AND now() - query_start > interval '5 minutes';"

# 2. If connection pool exhausted, increase limit
kubectl set env deployment/virtengine POSTGRES_MAX_CONNECTIONS=200

# 3. For persistent issues, add PgBouncer
```

---

## Security Incident Playbooks

### PLAY-SEC-001: Key Compromise Suspected

**Severity**: SEV-1  
**Impact**: Potential unauthorized access

#### Immediate Actions

1. **Revoke key immediately**
   ```bash
   virtengine tx encryption revoke-key \
       --fingerprint <compromised_key> \
       --reason "Suspected compromise" \
       --from admin
   ```

2. **Audit key usage**
   ```bash
   virtengine query encryption key-usage-log \
       --fingerprint <fingerprint> \
       --since "7 days ago"
   ```

3. **Notify affected parties**

#### Investigation

1. How was key potentially compromised?
2. What data was accessible with this key?
3. What is the exposure window?

#### Recovery

1. Re-encrypt affected data with new keys
2. Force re-verification for affected identities
3. Implement additional key protection measures

---

### PLAY-SEC-002: Unauthorized Access Attempt

**Severity**: SEV-2/SEV-3 (depending on success)  
**Impact**: Potential security breach

#### Detection

```bash
# Check failed auth attempts
grep "authentication failure" /var/log/auth.log | tail -50

# Check API authentication failures
grep "401\|403" /var/log/virtengine/api.log | tail -50
```

#### Response

1. Block attacking IPs
   ```bash
   sudo ufw deny from ATTACKER_IP
   ```

2. Increase authentication logging

3. If access succeeded, escalate to SEV-1

---

## Post-Incident Procedures

### Immediate (Within 2 hours)

1. Confirm all systems stable
2. Update status page to "Resolved"
3. Send summary to stakeholders
4. Create post-mortem document

### Post-Mortem (Within 48 hours)

**Template:**

```markdown
# Post-Incident Report: [INCIDENT TITLE]

## Summary
- **Incident ID**: INC-YYYYMMDD-XXX
- **Severity**: SEV-X
- **Duration**: HH:MM to HH:MM (X hours Y minutes)
- **Impact**: [Description of user impact]

## Timeline
| Time (UTC) | Event |
|------------|-------|
| HH:MM | Alert triggered |
| HH:MM | IC assigned |
| ...   | ... |
| HH:MM | Resolved |

## Root Cause
[Detailed technical explanation]

## Resolution
[What was done to fix it]

## Detection
- How was this detected?
- Could we have detected it sooner?

## Impact Analysis
- Users affected: X
- Revenue impact: $X
- Data impact: [None / Describe]

## Action Items
| Item | Owner | Due Date | Status |
|------|-------|----------|--------|
| [Action] | [Owner] | [Date] | Open |

## Lessons Learned
- What went well?
- What could be improved?
```

### Follow-up (Within 1 week)

1. Complete all P0 action items
2. Update runbooks with lessons learned
3. Conduct team retrospective
4. Track P1/P2 action items

---

**Document Owner:** SRE Team  
**Next Review:** 2026-04-30
