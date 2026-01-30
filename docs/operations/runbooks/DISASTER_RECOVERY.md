# Disaster Recovery Procedures

**Version:** 1.0.0  
**Last Updated:** 2026-01-30  
**Owner:** SRE Team

---

## Table of Contents

1. [Overview](#overview)
2. [DR Strategy](#dr-strategy)
3. [Recovery Time Objectives](#recovery-time-objectives)
4. [Disaster Scenarios](#disaster-scenarios)
5. [Recovery Procedures](#recovery-procedures)
6. [DR Testing](#dr-testing)
7. [Communication Plan](#communication-plan)

---

## Overview

This document outlines disaster recovery procedures for VirtEngine production infrastructure. The goal is to ensure business continuity and minimize data loss in catastrophic failure scenarios.

### Scope

| Component | DR Coverage | RPO | RTO |
|-----------|-------------|-----|-----|
| Blockchain (validators) | Distributed by design | 0 (consensus) | 15 min |
| Provider daemons | Multi-region | 1 hour | 30 min |
| Monitoring infrastructure | Hot standby | 5 min | 15 min |
| Supporting databases | Continuous backup | 15 min | 1 hour |

### Key Principles

1. **Blockchain is inherently distributed** - No single point of failure
2. **State is on-chain** - Recovery focuses on rejoining network
3. **Providers are independent** - Each provider manages own DR
4. **Keys are critical** - Key backup is paramount

---

## DR Strategy

### Architecture Overview

```
                    Primary Region                    │        DR Region
                                                      │
┌─────────────┐   ┌─────────────┐   ┌─────────────┐  │  ┌─────────────┐
│ Validator 1 │   │ Validator 2 │   │ Validator 3 │  │  │ Validator 4 │
└─────────────┘   └─────────────┘   └─────────────┘  │  └─────────────┘
                                                      │
┌─────────────┐   ┌─────────────┐                    │  ┌─────────────┐
│ Provider 1  │   │ Provider 2  │                    │  │ Provider 3  │
└─────────────┘   └─────────────┘                    │  └─────────────┘
                                                      │
┌─────────────┐                                       │  ┌─────────────┐
│ Monitoring  │ ────────── Replication ──────────────│──│  DR Monitor │
└─────────────┘                                       │  └─────────────┘
                                                      │
┌─────────────┐                                       │  ┌─────────────┐
│  Database   │ ────────── Streaming Rep ────────────│──│  DB Replica │
└─────────────┘                                       │  └─────────────┘
```

### Backup Locations

| Data Type | Primary Location | Backup Location | Frequency |
|-----------|-----------------|-----------------|-----------|
| Validator keys | HSM / encrypted | Offline cold storage | Initial + rotation |
| Chain data | Local SSD | S3-compatible storage | Hourly snapshots |
| Config files | Git repository | S3 versioned bucket | On change |
| Monitoring data | TimescaleDB | DR region replica | Continuous |

---

## Recovery Time Objectives

### By Component

| Component | RTO | RPO | Priority |
|-----------|-----|-----|----------|
| Validator node | 15 min | 0 (chain state) | P0 |
| Provider daemon | 30 min | 1 hour (usage data) | P1 |
| RPC endpoints | 15 min | 0 | P1 |
| Monitoring | 30 min | 5 min | P2 |
| Supporting services | 1 hour | 15 min | P2 |

### By Scenario

| Scenario | Expected Recovery Time |
|----------|----------------------|
| Single node failure | 5-15 minutes |
| Availability zone failure | 15-30 minutes |
| Region failure | 30-60 minutes |
| Complete data center loss | 1-4 hours |
| Key compromise | 24-48 hours |

---

## Disaster Scenarios

### Scenario 1: Single Validator Failure

**Impact**: Reduced voting power, no service interruption

**Detection**:
```bash
# Alert: ValidatorDown
# Check validator status
virtengine query staking validator $(virtengine keys show operator --bech val -a)
```

**Recovery**:
```bash
# Option A: Restart on same hardware
sudo systemctl restart virtengine

# Option B: Failover to standby
# 1. Stop failed validator
ssh failed-validator 'sudo systemctl stop virtengine'

# 2. Start standby with same keys
ssh standby-validator 'sudo systemctl start virtengine'

# Note: Ensure only ONE validator running with these keys
```

### Scenario 2: Availability Zone Failure

**Impact**: Multiple validators/providers affected, potential degradation

**Detection**:
- Cloud provider status page
- Multiple simultaneous alerts
- Network partition symptoms

**Recovery**:
```bash
# 1. Assess impact
for node in $(cat nodes-in-affected-az.txt); do
  echo "$node: $(curl -s http://$node:26657/status 2>/dev/null || echo 'UNREACHABLE')"
done

# 2. If < 1/3 validators affected
# Chain continues, no immediate action required
# Bring up replacement nodes in other AZs

# 3. If ≥ 1/3 validators affected
# See emergency coordination procedures
```

### Scenario 3: Region Failure

**Impact**: Significant portion of infrastructure unavailable

**Recovery Procedure**:

1. **Activate DR region** (if not already active)
   ```bash
   # Verify DR infrastructure ready
   for node in $(cat dr-nodes.txt); do
     ssh $node 'virtengine status'
   done
   ```

2. **Update DNS** (if using regional routing)
   ```bash
   # Update Route53 or equivalent
   aws route53 change-resource-record-sets \
       --hosted-zone-id $ZONE_ID \
       --change-batch file://failover-dns.json
   ```

3. **Scale up DR capacity**
   ```bash
   # Scale provider daemon replicas
   kubectl scale deployment provider-daemon --replicas=5 -n virtengine-system
   ```

### Scenario 4: Database Corruption/Loss

**Impact**: Historical data unavailable, analytics affected

**Recovery**:
```bash
# 1. Stop services writing to database
kubectl scale deployment virtengine-api --replicas=0

# 2. Restore from backup
pg_restore -h localhost -U virtengine -d virtengine_db \
    /backups/virtengine_db_$(date +%Y%m%d).dump

# 3. Apply WAL logs for point-in-time recovery
pg_wal_replay -D /var/lib/postgresql/data \
    --recovery-target-time="2026-01-30 10:00:00"

# 4. Restart services
kubectl scale deployment virtengine-api --replicas=3
```

### Scenario 5: Key Compromise

**Impact**: Security breach, potential unauthorized transactions

**Recovery** (see [INCIDENT_RESPONSE.md](INCIDENT_RESPONSE.md) PLAY-SEC-001):

1. Immediately revoke compromised key
2. Generate new keys
3. Rotate all dependent credentials
4. Re-register with new keys

---

## Recovery Procedures

### Validator Recovery

#### From Complete Node Loss

```bash
# 1. Provision new hardware/VM
# Requirements: 16 cores, 64GB RAM, 2TB NVMe

# 2. Install VirtEngine
wget https://github.com/virtengine/virtengine/releases/download/v1.0.0/virtengine_linux_amd64.tar.gz
tar -xzf virtengine_linux_amd64.tar.gz
sudo mv virtengine /usr/local/bin/

# 3. Initialize node
virtengine init "validator-name" --chain-id virtengine-1

# 4. Download genesis
wget -O ~/.virtengine/config/genesis.json \
    https://raw.githubusercontent.com/virtengine/networks/main/virtengine-1/genesis.json

# 5. Restore keys from secure backup
# CRITICAL: Ensure old node is STOPPED before restoring keys
gpg -d validator-keys-backup.tar.gz.gpg | tar -xzf - -C ~/.virtengine/

# 6. Restore configuration
cp /backups/config/config.toml ~/.virtengine/config/
cp /backups/config/app.toml ~/.virtengine/config/

# 7. Option A: State sync (fastest)
# Configure state-sync in config.toml
virtengine start

# 7. Option B: Snapshot restore
wget https://snapshots.virtengine.com/virtengine-1/latest.tar.lz4
lz4 -d latest.tar.lz4 | tar -xf - -C ~/.virtengine/data/
virtengine start

# 8. Verify validator signing
virtengine query slashing signing-info $(virtengine tendermint show-validator)
```

#### From Key Backup Only

```bash
# If chain data is unrecoverable but keys are safe

# 1. Set up new node (steps 1-4 above)

# 2. Restore validator key only
gpg -d validator-priv-key-backup.gpg > ~/.virtengine/config/priv_validator_key.json
gpg -d validator-node-key-backup.gpg > ~/.virtengine/config/node_key.json

# 3. Clear any existing state
virtengine tendermint unsafe-reset-all

# 4. State sync from network
# Configure state-sync in config.toml with trust height/hash
virtengine start

# 5. Verify rejoined network
virtengine status
```

### Provider Daemon Recovery

```bash
# 1. Provision infrastructure

# 2. Install provider daemon
wget https://github.com/virtengine/virtengine/releases/download/v1.0.0/provider-daemon_linux_amd64.tar.gz
tar -xzf provider-daemon_linux_amd64.tar.gz
sudo mv provider-daemon /usr/local/bin/

# 3. Restore configuration
cp /backups/provider-daemon/config.yaml ~/.provider-daemon/

# 4. Restore keys
gpg -d provider-keys-backup.tar.gz.gpg | tar -xzf - -C ~/.provider-daemon/

# 5. Restore TLS certificates
cp /backups/provider-daemon/tls/* ~/.provider-daemon/tls/

# 6. Start daemon
sudo systemctl start provider-daemon

# 7. Verify health
curl -k https://localhost:8443/health

# 8. Resume operations
virtengine tx provider set-status --status ACTIVE --from provider

# 9. Recover any pending usage reports
ls /var/lib/provider-daemon/failed_reports/
provider-daemon usage submit --force
```

### Monitoring Infrastructure Recovery

```bash
# 1. Activate DR monitoring
ssh dr-monitoring 'sudo systemctl start prometheus grafana alertmanager'

# 2. Update DNS to point to DR monitoring
aws route53 change-resource-record-sets ...

# 3. Verify all targets discovered
curl -s http://dr-monitoring:9090/api/v1/targets | jq '.data.activeTargets | length'

# 4. Verify alerts routing
# Trigger test alert
curl -X POST http://dr-monitoring:9093/api/v1/alerts \
    -H "Content-Type: application/json" \
    -d '[{"labels":{"alertname":"DRTest","severity":"info"}}]'
```

### Database Recovery

#### Point-in-Time Recovery

```bash
# 1. Stop services
kubectl scale deployment virtengine-api --replicas=0

# 2. Create recovery configuration
cat > /tmp/recovery.conf << EOF
restore_command = 'gsutil cp gs://backup-bucket/wal/%f %p'
recovery_target_time = '2026-01-30 10:00:00 UTC'
recovery_target_action = 'promote'
EOF

# 3. Copy to data directory
cp /tmp/recovery.conf /var/lib/postgresql/data/

# 4. Start PostgreSQL in recovery mode
sudo systemctl restart postgresql

# 5. Wait for recovery to complete
tail -f /var/log/postgresql/postgresql-14-main.log

# 6. Verify data integrity
psql -c "SELECT COUNT(*) FROM important_table;"

# 7. Restart services
kubectl scale deployment virtengine-api --replicas=3
```

#### Full Restore from Backup

```bash
# 1. Stop PostgreSQL
sudo systemctl stop postgresql

# 2. Clear existing data
rm -rf /var/lib/postgresql/data/*

# 3. Restore base backup
gsutil cp gs://backup-bucket/base/latest.tar.gz /tmp/
tar -xzf /tmp/latest.tar.gz -C /var/lib/postgresql/data/

# 4. Apply WAL files
# PostgreSQL will do this automatically on startup if configured

# 5. Start PostgreSQL
sudo systemctl start postgresql

# 6. Verify
psql -c "SELECT pg_is_in_recovery();"
# Should return 'f' (false) once recovery complete
```

---

## DR Testing

### Test Schedule

| Test Type | Frequency | Duration | Scope |
|-----------|-----------|----------|-------|
| Backup verification | Daily | Automated | All backups |
| Component failover | Monthly | 2 hours | Single component |
| AZ failover | Quarterly | 4 hours | Full AZ |
| Full DR exercise | Annually | 8 hours | Complete DR |

### Backup Verification Test

```bash
#!/bin/bash
# backup-verify.sh - Run daily

set -e

echo "=== Backup Verification Test ==="
echo "Date: $(date)"

# Verify validator key backups exist and are valid
echo "Checking validator key backups..."
for backup in /backups/validator-keys/*.gpg; do
    echo -n "  $backup: "
    if gpg --list-packets "$backup" > /dev/null 2>&1; then
        echo "OK"
    else
        echo "FAILED"
        exit 1
    fi
done

# Verify chain snapshot exists and is recent
echo "Checking chain snapshots..."
LATEST_SNAPSHOT=$(ls -t /backups/chain-snapshots/*.tar.lz4 | head -1)
SNAPSHOT_AGE=$(( ($(date +%s) - $(stat -c %Y "$LATEST_SNAPSHOT")) / 3600 ))
if [ $SNAPSHOT_AGE -gt 24 ]; then
    echo "ERROR: Latest snapshot is $SNAPSHOT_AGE hours old"
    exit 1
fi
echo "  Latest snapshot: $LATEST_SNAPSHOT ($SNAPSHOT_AGE hours old)"

# Verify database backup
echo "Checking database backups..."
LATEST_DB=$(ls -t /backups/database/*.dump | head -1)
DB_AGE=$(( ($(date +%s) - $(stat -c %Y "$LATEST_DB")) / 3600 ))
if [ $DB_AGE -gt 24 ]; then
    echo "ERROR: Latest database backup is $DB_AGE hours old"
    exit 1
fi
echo "  Latest database backup: $LATEST_DB ($DB_AGE hours old)"

# Test restore to temporary location
echo "Testing database restore..."
createdb test_restore_$(date +%Y%m%d)
pg_restore -d test_restore_$(date +%Y%m%d) "$LATEST_DB"
psql -d test_restore_$(date +%Y%m%d) -c "SELECT COUNT(*) FROM blocks;"
dropdb test_restore_$(date +%Y%m%d)
echo "  Database restore: OK"

echo "=== All backup verification tests passed ==="
```

### Component Failover Test

```markdown
## Monthly Failover Test Procedure

### Pre-Test
1. Notify stakeholders 24 hours in advance
2. Schedule during low-traffic period
3. Have rollback plan ready

### Test Procedure
1. Select component for testing (rotate monthly)
2. Simulate failure (stop service)
3. Verify automatic/manual failover
4. Measure recovery time
5. Verify service restored

### Post-Test
1. Document results
2. Compare to RTO targets
3. Update procedures if needed
4. Return to normal operations
```

### Full DR Exercise

```markdown
## Annual DR Exercise Plan

### Objectives
- Validate all DR procedures
- Train team on disaster recovery
- Identify gaps in documentation

### Scope
- Simulate complete primary region failure
- Recover all services in DR region
- Verify data integrity
- Test communication procedures

### Timeline
1. T-0: Declare exercise started
2. T+15 min: Complete initial assessment
3. T+1 hour: DR infrastructure activated
4. T+2 hours: Core services recovered
5. T+4 hours: Full service restoration
6. T+6 hours: Failback to primary
7. T+8 hours: Exercise complete

### Success Criteria
- RTO targets met for all components
- RPO targets met (no data loss beyond RPO)
- All runbook steps validated
- Communication plan executed
```

---

## Communication Plan

### Internal Communication

| Audience | Channel | Timing | Content |
|----------|---------|--------|---------|
| On-call | PagerDuty | Immediate | Alert details |
| SRE team | Slack #incident | Immediate | Incident declared |
| Engineering | Slack #engineering | 15 min | Status update |
| Leadership | Email + Slack | 30 min | Impact summary |

### External Communication

| Audience | Channel | Timing | Content |
|----------|---------|--------|---------|
| Validators | Discord #validators | 15 min | Technical status |
| Providers | Discord #providers | 15 min | Impact & ETA |
| Users | Status page | 30 min | Service status |
| Press | PR team | 4+ hours | If needed |

### Communication Templates

**Status Page - DR Activated**:
```
[MAJOR OUTAGE] Primary Region Failure

We are experiencing a major infrastructure outage affecting our primary region.
Our disaster recovery procedures have been activated.

Impact: [Description]
Current Status: Recovering services in DR region
ETA: [Estimated restoration time]

We will provide updates every 30 minutes.
```

**Status Page - Recovery In Progress**:
```
[MAJOR OUTAGE - UPDATE] Recovery In Progress

Services are being restored in our disaster recovery region.

Restored:
- ✅ Blockchain consensus
- ✅ RPC endpoints
- 🔄 Provider services (in progress)
- ⏳ Historical data (pending)

ETA for full restoration: [Time]
```

**Status Page - Resolved**:
```
[RESOLVED] Primary Region Failure

All services have been restored. The system is operating normally.

Duration: [Start time] to [End time]
Impact: [Summary of impact]

A post-incident report will be published within 48 hours.

We apologize for any inconvenience caused.
```

---

## Appendix: DR Checklist

### Pre-Disaster Checklist (Quarterly Review)

- [ ] All backup jobs running successfully
- [ ] DR infrastructure provisioned and tested
- [ ] Keys backed up securely (multiple locations)
- [ ] Runbooks up to date
- [ ] Contact lists current
- [ ] Communication templates ready

### During-Disaster Checklist

- [ ] Incident declared and roles assigned
- [ ] Assess scope and impact
- [ ] Activate appropriate DR procedure
- [ ] Communicate status to stakeholders
- [ ] Execute recovery steps
- [ ] Verify service restoration
- [ ] Document all actions

### Post-Disaster Checklist

- [ ] Confirm all services stable
- [ ] Update status page to resolved
- [ ] Conduct initial debrief
- [ ] Schedule post-mortem
- [ ] Create follow-up tickets
- [ ] Update runbooks with lessons learned
- [ ] Replenish DR resources

---

**Document Owner:** SRE Team  
**Next Review:** 2026-04-30
