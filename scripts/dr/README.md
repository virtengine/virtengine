# VirtEngine Disaster Recovery Scripts

This directory contains scripts for disaster recovery operations including backup, restore, failover, and testing.

## Scripts Overview

| Script | Purpose | Frequency |
|--------|---------|-----------|
| `backup-chain-state.sh` | Backup blockchain state and data | Every 4 hours (automated) |
| `backup-provider-state.sh` | Backup provider daemon state | Every 4 hours (automated) |
| `backup-keys.sh` | Backup validator and provider keys | Daily (automated) |
| `dr-test.sh` | Automated DR validation tests | Daily (automated) |

## Quick Start

### Create a Chain State Backup

```bash
# Full backup with remote upload
./backup-chain-state.sh

# Local snapshot only
./backup-chain-state.sh --snapshot-only

# List available snapshots
./backup-chain-state.sh --list

# Verify backup integrity
./backup-chain-state.sh --verify

# Restore to specific height
./backup-chain-state.sh --restore 1234567
```

### Create Key Backups

```bash
# Backup all keys
./backup-keys.sh

# Backup specific key type
./backup-keys.sh --type validator
./backup-keys.sh --type provider

# With Shamir secret sharing (3-of-5)
./backup-keys.sh --type validator --shamir

# List existing backups
./backup-keys.sh --list

# Verify backups
./backup-keys.sh --verify
```

### Create Provider State Backups

```bash
# Backup provider daemon state
./backup-provider-state.sh

# List backups
./backup-provider-state.sh --list

# Verify latest backup
./backup-provider-state.sh --verify

# Restore a backup
./backup-provider-state.sh --restore provider_state_YYYYMMDD_HHMMSS
```

### Run DR Tests

```bash
# Run all tests
./dr-test.sh

# Run specific test suite
./dr-test.sh --test backup
./dr-test.sh --test connectivity

# Generate report and notify
./dr-test.sh --report --notify
```

## Environment Variables

### Common Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DR_BUCKET` | S3 bucket for remote backups | (required for remote) |
| `SLACK_WEBHOOK` | Slack webhook for notifications | (optional) |
| `ALERT_WEBHOOK` | Webhook for backup/restore events | (optional) |
| `SNAPSHOT_SIGNING_KEY` | PEM private key for snapshot signing | (required) |
| `SNAPSHOT_VERIFY_PUBKEY` | PEM public key for verification | (required) |

### Chain State Backup

| Variable | Description | Default |
|----------|-------------|---------|
| `NODE_HOME` | VirtEngine node home directory | `/opt/virtengine` |
| `SNAPSHOT_DIR` | Local snapshot storage | `/data/snapshots` |
| `RETENTION_COUNT` | Local snapshots to keep | `10` |
| `RESTORE_AUTO_APPROVE` | Skip restore delay | `0` |
| `RESTORE_SKIP_SERVICE` | Skip systemctl stop/start | `0` |
| `RESTORE_STATUS_TIMEOUT` | Wait for status after restore (seconds) | `60` |
| `RESTORE_MAX_WAIT` | Seconds to wait for sync check | `300` |
| `RESTORE_FALLBACK_ENABLED` | Allow fallback to older snapshots | `1` |
| `RESTORE_ROLLBACK_ON_FAILURE` | Roll back to previous data on restore failure | `1` |

### Key Backup

| Variable | Description | Default |
|----------|-------------|---------|
| `KEY_DIR` | Key configuration directory | `/opt/virtengine/config` |
| `BACKUP_DIR` | Key backup output directory | `/secure/backups` |
| `HSM_ENABLED` | Enable HSM support | `0` |
| `HSM_MODULE` | PKCS#11 module path | `/usr/lib/softhsm/libsofthsm2.so` |
| `PROVIDER_HOME` | Provider daemon home | `/opt/provider-daemon` |

### Provider State Backup

| Variable | Description | Default |
|----------|-------------|---------|
| `PROVIDER_HOME` | Provider daemon home | `/opt/provider-daemon` |
| `PROVIDER_SNAPSHOT_DIR` | Provider backup storage | `/data/provider-snapshots` |
| `RETENTION_COUNT` | Backups to keep | `10` |
| `RESTORE_FALLBACK_ENABLED` | Allow fallback to latest valid backup | `1` |
| `RESTORE_ROLLBACK_ON_FAILURE` | Roll back to previous provider data/config on failure | `1` |
| `PROVIDER_HEALTHCHECK_CMD` | Optional healthcheck command post-restore | (optional) |

### Snapshot Signing Keys

Generate a dedicated signing keypair for snapshots and distribute the public key to restore hosts:

```bash
openssl genpkey -algorithm RSA -out /secure/keys/snapshot_signing.pem -pkeyopt rsa_keygen_bits:2048
openssl pkey -in /secure/keys/snapshot_signing.pem -pubout -out /secure/keys/snapshot_signing.pub
```

## Automated Scheduling

### Kubernetes CronJobs

```yaml
# Chain state backup - every 4 hours
apiVersion: batch/v1
kind: CronJob
metadata:
  name: dr-backup-chain-state
spec:
  schedule: "0 */4 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: virtengine/dr-tools:latest
            command: ["/scripts/dr/backup-chain-state.sh"]
            env:
            - name: DR_BUCKET
              value: "s3://virtengine-dr-backups"
            volumeMounts:
            - name: node-data
              mountPath: /opt/virtengine
            - name: snapshots
              mountPath: /data/snapshots
          restartPolicy: OnFailure
          volumes:
          - name: node-data
            persistentVolumeClaim:
              claimName: validator-data
          - name: snapshots
            persistentVolumeClaim:
              claimName: snapshot-storage
```

```yaml
# Key backup - daily at 2 AM
apiVersion: batch/v1
kind: CronJob
metadata:
  name: dr-backup-keys
spec:
  schedule: "0 2 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: virtengine/dr-tools:latest
            command: ["/scripts/dr/backup-keys.sh", "--type", "all"]
            volumeMounts:
            - name: keys
              mountPath: /opt/virtengine/config
              readOnly: true
            - name: backup-storage
              mountPath: /secure/backups
          restartPolicy: OnFailure
```

```yaml
# Provider state backup - every 4 hours
apiVersion: batch/v1
kind: CronJob
metadata:
  name: dr-backup-provider-state
spec:
  schedule: "30 */4 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: virtengine/dr-tools:latest
            command: ["/scripts/dr/backup-provider-state.sh"]
            env:
            - name: DR_BUCKET
              value: "s3://virtengine-dr-backups"
            volumeMounts:
            - name: provider-data
              mountPath: /opt/provider-daemon
            - name: provider-snapshots
              mountPath: /data/provider-snapshots
          restartPolicy: OnFailure
          volumes:
          - name: provider-data
            persistentVolumeClaim:
              claimName: provider-daemon-data
          - name: provider-snapshots
            persistentVolumeClaim:
              claimName: provider-snapshots
```

```yaml
# DR tests - daily at 6 AM
apiVersion: batch/v1
kind: CronJob
metadata:
  name: dr-tests
spec:
  schedule: "0 6 * * *"
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: test
            image: virtengine/dr-tools:latest
            command: ["/scripts/dr/dr-test.sh", "--report", "--notify"]
            env:
            - name: DR_BUCKET
              value: "s3://virtengine-dr-backups"
            - name: SLACK_WEBHOOK
              valueFrom:
                secretKeyRef:
                  name: dr-secrets
                  key: slack-webhook
          restartPolicy: OnFailure
```

### Systemd Timers

```ini
# /etc/systemd/system/dr-backup-chain.timer
[Unit]
Description=Chain state backup timer

[Timer]
OnCalendar=*-*-* 00/4:00:00
Persistent=true

[Install]
WantedBy=timers.target
```

```ini
# /etc/systemd/system/dr-backup-chain.service
[Unit]
Description=Chain state backup

[Service]
Type=oneshot
ExecStart=/opt/virtengine/scripts/dr/backup-chain-state.sh
Environment=DR_BUCKET=s3://virtengine-dr-backups
```

## Recovery Procedures

### Single Node Recovery

1. **Stop the node** (if running):
   ```bash
   systemctl stop virtengine
   ```

2. **Restore from backup**:
   ```bash
   ./backup-chain-state.sh --restore <HEIGHT>
   ```

3. **Monitor sync**:
   ```bash
   virtengine status | jq '.sync_info'
   ```

### Provider Daemon Recovery

1. **Stop provider daemon** (if running):
   ```bash
   systemctl stop provider-daemon
   ```

2. **Restore provider state**:
   ```bash
   ./backup-provider-state.sh --restore provider_state_YYYYMMDD_HHMMSS
   ```

3. **Validate service**:
   ```bash
   systemctl status provider-daemon
   ```

### Key Recovery

1. **Get passphrase from Secrets Manager**:
   ```bash
   aws secretsmanager get-secret-value \
     --secret-id virtengine/dr/backup-passphrase-validator \
     --query SecretString --output text
   ```

2. **Decrypt backup**:
   ```bash
   openssl enc -d -aes-256-gcm -pbkdf2 -iter 100000 \
     -in validator_keys_TIMESTAMP_encrypted.tar.gz.enc \
     -out validator_keys.tar.gz \
     -pass pass:YOUR_PASSPHRASE
   ```

3. **Extract and verify**:
   ```bash
   tar -xzf validator_keys.tar.gz
   ```

### Shamir Share Recovery

1. **Collect threshold number of shares** (minimum 3 for 3-of-5):
   ```bash
   # From different custodians
   cat share_1.txt share_3.txt share_5.txt | ssss-combine -t 3
   ```

2. **The output is the encrypted backup file**

3. **Decrypt using standard recovery procedure**

## Security Considerations

### Validator Key Backup Warning

⚠️ **CRITICAL**: Validator key backups require extreme care:

- **NEVER** restore a validator key while another instance is running
- Improper restoration can cause **double-signing** and **slashing**
- HSM-backed keys should use HSM's native backup/restore
- Consider using file-based keys only for testing

### Recommended Practices

1. **Use HSM for production validators**
2. **Distribute Shamir shares geographically**
3. **Test recovery procedures quarterly**
4. **Rotate backup encryption keys annually**
5. **Audit backup access logs**

## Monitoring and Alerts

### Key Metrics

| Metric | Alert Threshold |
|--------|-----------------|
| `dr_backup_age_hours` | > 24 hours |
| `dr_backup_size_bytes` | < 1MB (likely failed) |
| `dr_test_failures` | > 0 |
| `dr_restore_duration_seconds` | > 3600 (1 hour) |
| `dr_provider_backup_age_hours` | > 24 hours |

### Prometheus Rules

```yaml
groups:
- name: dr_alerts
  rules:
  - alert: DRBackupStale
    expr: time() - dr_last_backup_timestamp > 86400
    for: 1h
    labels:
      severity: warning
    annotations:
      summary: "DR backup is stale"
      
  - alert: DRTestFailed
    expr: dr_test_failures > 0
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "DR test failed"
```

## Related Documentation

- [Disaster Recovery Plan](../../_docs/disaster-recovery.md)
- [Business Continuity Plan](../../_docs/business-continuity.md)
- [Key Management](../../_docs/key-management.md)
- [Incident Response](../../docs/sre/INCIDENT_RESPONSE.md)
