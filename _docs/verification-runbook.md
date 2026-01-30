# VEID Verification Service Runbook

This runbook provides operational procedures for managing the VEID verification infrastructure.

## Table of Contents

1. [Key Rotation Procedures](#key-rotation-procedures)
2. [Incident Response](#incident-response)
3. [Recovery Procedures](#recovery-procedures)
4. [Maintenance Tasks](#maintenance-tasks)

---

## Key Rotation Procedures

### Scheduled Key Rotation

Keys should be rotated before expiration (default: 90 days). The system will warn when keys are 80+ days old.

#### Pre-Rotation Checklist

- [ ] Verify new key storage capacity
- [ ] Ensure monitoring is active
- [ ] Notify dependent services of upcoming rotation
- [ ] Confirm Vault/HSM is accessible
- [ ] Review current rotation status (no rotation in progress)

#### Rotation Steps

```bash
# 1. Check current key status
curl -s http://signer:8080/api/v1/keys | jq .

# 2. Verify health
curl -s http://signer:8080/healthz

# 3. Initiate rotation via API
curl -X POST http://signer:8080/api/v1/keys/rotate \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "rotation",
    "initiated_by": "ops-team",
    "notes": "Scheduled rotation"
  }'

# 4. Monitor rotation progress
curl -s http://signer:8080/api/v1/rotations/{rotation_id}

# 5. Verify new key is active
curl -s http://signer:8080/api/v1/keys/active

# 6. Complete rotation after overlap period
curl -X POST http://signer:8080/api/v1/rotations/{rotation_id}/complete
```

#### Post-Rotation Verification

```bash
# Verify attestation signing works
curl -X POST http://signer:8080/api/v1/sign/test

# Check metrics
curl -s http://signer:9090/metrics | grep signer_active_keys

# Review audit logs for rotation events
curl -s http://signer:8080/api/v1/audit?event_type=key_rotated
```

### Emergency Key Rotation (Key Compromise)

If a signing key is suspected to be compromised, follow this emergency procedure.

#### Immediate Actions

```bash
# 1. IMMEDIATELY revoke the compromised key
curl -X POST http://signer:8080/api/v1/keys/{key_id}/revoke \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "compromised",
    "initiated_by": "security-team"
  }'

# 2. Initiate emergency rotation (no overlap period)
curl -X POST http://signer:8080/api/v1/keys/rotate \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "compromised",
    "initiated_by": "security-team",
    "emergency": true,
    "notes": "Key compromise incident"
  }'

# 3. Verify new key is active
curl -s http://signer:8080/api/v1/keys/active
```

#### Follow-up Actions

1. **Create incident ticket** with details of compromise
2. **Review audit logs** for unauthorized key usage
3. **Identify affected attestations** signed with compromised key
4. **Notify stakeholders** about potential impact
5. **Update security procedures** to prevent recurrence

---

## Incident Response

### Incident Severity Levels

| Level | Description | Response Time |
|-------|-------------|---------------|
| P1 - Critical | Service down, key compromise | Immediate |
| P2 - High | Degraded service, high error rate | < 1 hour |
| P3 - Medium | Elevated latency, partial failures | < 4 hours |
| P4 - Low | Minor issues, monitoring alerts | < 24 hours |

### P1: Service Unavailable

**Symptoms:**
- Health endpoint returns unhealthy
- No attestations being signed
- Connection errors from dependent services

**Investigation:**
```bash
# 1. Check pod/container status
kubectl get pods -n virtengine-veid -l app=verification-signer

# 2. Check logs
kubectl logs -n virtengine-veid -l app=verification-signer --tail=100

# 3. Check health endpoint
curl -v http://signer:8080/healthz

# 4. Check dependencies
# Redis
redis-cli -h redis ping

# Vault
curl -s https://vault:8200/v1/sys/health
```

**Resolution:**
1. If pods not running: Check resource limits, node capacity
2. If Vault unreachable: Check Vault status, renew token
3. If Redis unreachable: Check Redis cluster health
4. If key storage error: Check storage backend connectivity

### P1: Suspected Key Compromise

**Symptoms:**
- Unauthorized attestations detected
- Key accessed from unknown source
- Security alert triggered

**Immediate Actions:**
1. Follow [Emergency Key Rotation](#emergency-key-rotation-key-compromise)
2. Preserve all logs for forensics
3. Block suspicious IP addresses

**Investigation:**
```bash
# Review audit logs
curl -s "http://signer:8080/api/v1/audit?event_type=key_accessed&limit=100"

# Check for unusual signing patterns
curl -s "http://signer:8080/api/v1/audit?event_type=attestation_signed&limit=1000" | \
  jq 'group_by(.request.ip_address) | map({ip: .[0].request.ip_address, count: length})'
```

### P2: High Error Rate

**Symptoms:**
- Error rate > 1% on signing operations
- Increased latency
- Rate limit rejections spiking

**Investigation:**
```bash
# Check error metrics
curl -s http://signer:9090/metrics | grep signer_errors_total

# Check latency
curl -s http://signer:9090/metrics | grep signer_sign_latency_seconds

# Review recent errors in logs
kubectl logs -n virtengine-veid -l app=verification-signer --since=5m | grep -i error
```

**Resolution:**
1. If storage errors: Check Vault/HSM connectivity
2. If rate limit errors: Review limits, check for attack
3. If latency spikes: Scale service, check Redis performance

### P2: Nonce Store Issues

**Symptoms:**
- Nonce validation failures
- Replay attack alerts
- Nonce store full warnings

**Investigation:**
```bash
# Check nonce store stats
curl -s http://signer:8080/api/v1/nonce/stats

# Check Redis memory
redis-cli -h redis INFO memory

# Review nonce rejection reasons
curl -s http://signer:9090/metrics | grep nonce_rejected_total
```

**Resolution:**
1. Run manual cleanup if needed
2. Increase Redis memory if near capacity
3. Adjust nonce window if too short
4. Review clock synchronization across nodes

---

## Recovery Procedures

### Restore from Backup

If key storage is corrupted or lost:

```bash
# 1. Stop all signer instances
kubectl scale deployment verification-signer -n virtengine-veid --replicas=0

# 2. Restore Vault data from backup
vault operator raft snapshot restore <backup-file>

# 3. Verify key data
vault kv list secret/veid/signer/keys

# 4. Restart signers
kubectl scale deployment verification-signer -n virtengine-veid --replicas=3

# 5. Verify health
curl -s http://signer:8080/healthz

# 6. Test signing
curl -X POST http://signer:8080/api/v1/sign/test
```

### Generate New Initial Key

If no keys exist (e.g., after full data loss):

```bash
# 1. Stop all signer instances
kubectl scale deployment verification-signer -n virtengine-veid --replicas=0

# 2. Clear any corrupt state
redis-cli -h redis KEYS "virtengine:veid:*" | xargs redis-cli -h redis DEL

# 3. Start a single instance to generate initial key
kubectl scale deployment verification-signer -n virtengine-veid --replicas=1

# 4. Wait for initialization
kubectl logs -n virtengine-veid -l app=verification-signer -f

# 5. Verify key generated
curl -s http://signer:8080/api/v1/keys/active

# 6. Scale back up
kubectl scale deployment verification-signer -n virtengine-veid --replicas=3
```

### Redis Failover

If Redis master fails:

```bash
# 1. Check sentinel status
redis-cli -h redis-sentinel SENTINEL get-master-addr-by-name mymaster

# 2. If failover not automatic, trigger manually
redis-cli -h redis-sentinel SENTINEL failover mymaster

# 3. Verify new master
redis-cli -h redis-sentinel SENTINEL get-master-addr-by-name mymaster

# 4. Check signer reconnection
curl -s http://signer:8080/healthz
```

---

## Maintenance Tasks

### Daily Tasks

- [ ] Review error rate metrics
- [ ] Check key age (rotation needed?)
- [ ] Review abuse score alerts
- [ ] Verify audit log backups

### Weekly Tasks

- [ ] Review rate limit effectiveness
- [ ] Check nonce store capacity
- [ ] Review and rotate Vault tokens
- [ ] Test disaster recovery procedure

### Monthly Tasks

- [ ] Review and update rate limit thresholds
- [ ] Capacity planning review
- [ ] Security audit review
- [ ] Update runbook if needed

### Nonce Cleanup (Manual)

```bash
# Trigger manual cleanup
curl -X POST http://signer:8080/api/v1/nonce/cleanup

# Check cleanup results
curl -s http://signer:8080/api/v1/nonce/stats
```

### Audit Log Rotation

```bash
# Check audit log size
du -sh /var/log/virtengine/audit/

# Trigger rotation (if file-based)
kill -USR1 $(pgrep verification-signer)

# Archive old logs
tar -czf audit-$(date +%Y%m%d).tar.gz /var/log/virtengine/audit/*.log.*
aws s3 cp audit-*.tar.gz s3://virtengine-audit-archive/
```

### Metrics Cleanup

```bash
# Check Prometheus retention
curl -s http://prometheus:9090/api/v1/status/runtimeinfo | jq .data.storageRetention

# If needed, adjust retention in Prometheus config
```

---

## Contacts

| Role | Contact |
|------|---------|
| On-Call Engineer | pagerduty.com/oncall |
| Security Team | security@virtengine.io |
| Platform Team | platform@virtengine.io |

## Related Documents

- [Verification Deployment Guide](verification-deployment.md)
- [VEID Architecture Overview](veid-architecture.md)
- [Security Incident Response Plan](security-incident-response.md)
