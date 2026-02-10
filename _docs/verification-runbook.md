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

---

## SMS Verification Operations

### Overview

The SMS verification service provides phone number verification for VEID identity attestations. It includes:

- Primary/secondary SMS gateway failover (Twilio, AWS SNS)
- OTP generation with secure hashing (SHA256)
- VoIP detection and blocking
- Velocity checks and anti-fraud controls
- Signed attestations for on-chain storage

### Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │───▶│ SMS Service │───▶│  Primary    │
│             │    │             │    │  Provider   │
└─────────────┘    │             │    │  (Twilio)   │
                   │             │    └──────┬──────┘
                   │             │           │ failover
                   │             │    ┌──────▼──────┐
                   │             │───▶│  Secondary  │
                   │             │    │  Provider   │
                   │             │    │  (SNS)      │
                   └──────┬──────┘    └─────────────┘
                          │
           ┌──────────────┼──────────────┐
           ▼              ▼              ▼
    ┌──────────┐   ┌──────────┐   ┌──────────┐
    │  Redis   │   │  Signer  │   │  Auditor │
    │  Cache   │   │  Service │   │          │
    └──────────┘   └──────────┘   └──────────┘
```

### Configuration Reference

| Parameter | Default | Description |
|-----------|---------|-------------|
| `OTPLength` | 6 | Number of digits in OTP |
| `OTPTTLSeconds` | 300 | OTP validity period (5 min) |
| `MaxAttempts` | 3 | Max verification attempts |
| `MaxResends` | 3 | Max OTP resends per session |
| `ResendCooldownSeconds` | 60 | Cooldown between resends |
| `EnableVoIPBlocking` | true | Block VoIP numbers |
| `EnableVelocityChecks` | true | Enable rate limiting |
| `MaxRequestsPerPhonePerHour` | 3 | Phone velocity limit |
| `MaxRequestsPerIPPerHour` | 10 | IP velocity limit |
| `MaxRequestsPerAccountPerDay` | 10 | Account daily limit |

### SMS Gateway Health Checks

```bash
# Check primary provider (Twilio)
curl -s http://sms-service:8080/api/v1/health | jq .

# Check failover provider status
curl -s http://sms-service:8080/api/v1/providers/status

# Test SMS delivery (dry run)
curl -X POST http://sms-service:8080/api/v1/test \
  -H "Content-Type: application/json" \
  -d '{"phone": "+14155551234", "dry_run": true}'
```

### Monitoring Metrics

Key Prometheus metrics for SMS verification:

```promql
# SMS delivery rate
sum(rate(veid_sms_sent_total[5m])) by (provider, country_code)

# Verification success rate
sum(rate(veid_sms_otp_successful_total[5m])) / sum(rate(veid_sms_otp_attempts_total[5m]))

# VoIP detection rate
sum(rate(veid_sms_voip_detected_total[5m])) by (country_code)

# Rate limit hits
sum(rate(veid_sms_rate_limit_hit_total[5m])) by (type)

# Provider failover rate
sum(rate(veid_sms_provider_failover_total[5m]))

# Attestation creation rate
sum(rate(veid_sms_attestations_created_total[5m])) by (country_code)
```

### Alert Thresholds

| Alert | Threshold | Severity |
|-------|-----------|----------|
| SMS Delivery Failure Rate | > 5% | P2 |
| VoIP Detection Rate | > 20% | P3 |
| Rate Limit Hit Rate | > 10% | P3 |
| Provider Failover | Any | P3 |
| Primary Provider Down | > 1 min | P2 |
| Both Providers Down | Any | P1 |
| Attestation Failure Rate | > 1% | P2 |

### Incident Response: SMS Delivery Issues

**Symptoms:**
- High SMS delivery failure rate
- Customer reports of OTPs not received
- Provider error metrics spiking

**Investigation:**
```bash
# Check delivery status by provider
curl -s http://sms-service:9090/metrics | grep veid_sms_failed_total

# Check provider-specific errors
curl -s http://sms-service:8080/api/v1/delivery/status?status=failed&limit=100

# Check country-specific issues
curl -s http://sms-service:9090/metrics | grep veid_sms_sent_total | sort
```

**Resolution:**
1. If Twilio errors: Check Twilio console, verify API keys
2. If SNS errors: Check AWS console, verify IAM permissions
3. If country-specific: Check carrier blocklists, regional routing
4. If all providers: Check network connectivity, DNS resolution

### Incident Response: VoIP Abuse Attack

**Symptoms:**
- Spike in VoIP detection metrics
- Velocity limit hits increasing
- Same IP/device requesting multiple verifications

**Investigation:**
```bash
# Check VoIP detection breakdown
curl -s http://sms-service:9090/metrics | grep veid_sms_voip_detected_total

# Check velocity stats for suspicious IPs
curl -s http://sms-service:8080/api/v1/antifraud/stats?type=ip

# Check blocked phones
curl -s http://sms-service:8080/api/v1/antifraud/blocked/phones?limit=50
```

**Mitigation:**
```bash
# Block suspicious IP range
curl -X POST http://sms-service:8080/api/v1/antifraud/block/ip \
  -H "Content-Type: application/json" \
  -d '{"ip_range": "192.168.0.0/16", "reason": "abuse", "duration_hours": 24}'

# Block phone number
curl -X POST http://sms-service:8080/api/v1/antifraud/block/phone \
  -H "Content-Type: application/json" \
  -d '{"phone_hash": "<hash>", "reason": "abuse", "duration_hours": 168}'

# Temporarily increase velocity limits (emergency only)
# Update config and restart service
```

### Provider Failover Testing

Monthly testing of failover capabilities:

```bash
# 1. Verify current provider status
curl -s http://sms-service:8080/api/v1/providers/status

# 2. Simulate primary failure
kubectl set env deployment/sms-service -n virtengine-veid \
  SMS_PRIMARY_PROVIDER_DISABLED=true

# 3. Send test SMS
curl -X POST http://sms-service:8080/api/v1/test \
  -H "Content-Type: application/json" \
  -d '{"phone": "+14155551234"}'

# 4. Verify secondary was used
curl -s http://sms-service:9090/metrics | grep secondary_sent

# 5. Re-enable primary
kubectl set env deployment/sms-service -n virtengine-veid \
  SMS_PRIMARY_PROVIDER_DISABLED-

# 6. Verify primary is back
curl -s http://sms-service:8080/api/v1/providers/status
```

### Regional Rate Limit Adjustments

Some regions require different rate limits due to fraud patterns:

| Region | Max/Phone/Hour | Max/Account/Day | VoIP Block | Risk Multiplier |
|--------|----------------|-----------------|------------|-----------------|
| US/CA | 5 | 15 | Yes | 1.0x |
| GB | 5 | 15 | Yes | 1.0x |
| IN | 3 | 8 | Yes | 1.2x |
| PH | 2 | 5 | Yes | 1.3x |
| NG | 2 | 5 | Yes | 1.5x |

To update regional limits:

```bash
# Update config map
kubectl edit configmap sms-service-config -n virtengine-veid

# Restart to apply
kubectl rollout restart deployment/sms-service -n virtengine-veid
```

### Anti-Fraud Manual Actions

#### Block a Phone Number

```bash
curl -X POST http://sms-service:8080/api/v1/antifraud/block/phone \
  -H "Content-Type: application/json" \
  -d '{
    "phone_hash": "<sha256-hash>",
    "reason": "fraud_confirmed",
    "duration_hours": 8760
  }'
```

#### Unblock a Phone Number

```bash
curl -X DELETE http://sms-service:8080/api/v1/antifraud/block/phone \
  -H "Content-Type: application/json" \
  -d '{"phone_hash": "<sha256-hash>"}'
```

#### Review Velocity Stats

```bash
# By phone hash
curl -s http://sms-service:8080/api/v1/antifraud/velocity/phone/<hash>

# By IP hash
curl -s http://sms-service:8080/api/v1/antifraud/velocity/ip/<hash>

# By account
curl -s http://sms-service:8080/api/v1/antifraud/velocity/account/<address>
```

### Cache Maintenance

```bash
# Check Redis SMS cache stats
redis-cli -h redis INFO keyspace | grep sms

# Clear expired challenges (automatic, but can force)
redis-cli -h redis SCAN 0 MATCH "sms:challenge:*" COUNT 1000

# Check anti-fraud cache size
redis-cli -h redis DBSIZE
```

### SMS Template Updates

Templates are defined in code but can be overridden via config:

```yaml
# In ConfigMap
sms_templates:
  otp_verification:
    en: "Your VirtEngine code is: {{otp}}. Valid for {{expires_in}}."
    es: "Tu código de VirtEngine es: {{otp}}. Válido por {{expires_in}}."
    # Add more locales...
```

Apply changes:
```bash
kubectl apply -f sms-templates-configmap.yaml
kubectl rollout restart deployment/sms-service -n virtengine-veid
```

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| OTP not received | Carrier filtering | Check carrier blocklist, try alternate sender ID |
| High latency | Provider congestion | Check provider status page, consider failover |
| VoIP false positives | Carrier lookup error | Review carrier patterns, whitelist valid carriers |
| Rate limit errors | Velocity misconfiguration | Review and adjust regional limits |
| Attestation failures | Signer unavailable | Check signer health, verify key rotation |
