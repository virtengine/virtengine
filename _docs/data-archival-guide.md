# Data Archival Operations Guide

This guide provides step-by-step instructions for operators managing the VirtEngine data archival system.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Initial Setup](#initial-setup)
3. [Daily Operations](#daily-operations)
4. [Monitoring](#monitoring)
5. [Troubleshooting](#troubleshooting)
6. [Maintenance](#maintenance)

## Prerequisites

- Running VirtEngine validator node
- Access to cloud storage (AWS S3, Azure, or local filesystem)
- Archive encryption keys configured
- Understanding of retention policies

## Initial Setup

### 1. Configure Archive Backend

Choose your archive backend based on requirements:

#### Option A: AWS S3 Glacier (Recommended for Production)

```bash
# Install AWS CLI
pip install awscli

# Configure credentials
aws configure

# Create S3 bucket with lifecycle policy
aws s3api create-bucket \
    --bucket virtengine-archives-${VALIDATOR_MONIKER} \
    --region us-east-1

# Configure lifecycle transition to Glacier
aws s3api put-bucket-lifecycle-configuration \
    --bucket virtengine-archives-${VALIDATOR_MONIKER} \
    --lifecycle-configuration file://lifecycle.json
```

`lifecycle.json`:
```json
{
  "Rules": [
    {
      "Id": "TransitionToGlacier",
      "Status": "Enabled",
      "Transitions": [
        {
          "Days": 30,
          "StorageClass": "GLACIER"
        },
        {
          "Days": 365,
          "StorageClass": "DEEP_ARCHIVE"
        }
      ]
    }
  ]
}
```

#### Option B: Local Filesystem (Development/Testing)

```bash
# Create archive directories
mkdir -p /var/lib/virtengine/archives/{data,metadata}
chmod 700 /var/lib/virtengine/archives

# Configure in app.toml
cat >> ~/.virtengine/config/app.toml <<EOF
[archive]
backend = "filesystem"
archive_dir = "/var/lib/virtengine/archives/data"
metadata_dir = "/var/lib/virtengine/archives/metadata"
restore_ttl_seconds = 86400
enable_integrity_checks = true
EOF
```

### 2. Generate Archive Encryption Keys

```bash
# Generate archive master key
virtengine keys add archive-master --keyring-backend file

# Derive archive encryption key
virtengine keys derive-archive-key \
    --master-key archive-master \
    --key-id "v1.0.0" \
    --output archive-key.json

# Store securely (use HSM in production)
chmod 400 archive-key.json
```

### 3. Enable Archival System

```bash
# Submit governance proposal to enable archival
virtengine tx gov submit-proposal param-change \
    --title "Enable Data Archival System" \
    --description "Enable automated data archival for cost optimization" \
    --changes '[{
        "subspace": "veid",
        "key": "ArchivalConfig",
        "value": {
            "enabled": true,
            "auto_archive": true,
            "archival_check_interval_blocks": 100,
            "default_archive_tier": "standard",
            "max_archives_per_block": 10,
            "min_age_for_archival_blocks": 12960,
            "min_age_for_archival_seconds": 7776000,
            "restore_ttl_seconds": 86400,
            "enable_integrity_checks": true,
            "integrity_check_interval_blocks": 14400
        }
    }]' \
    --deposit 10000000uvirt \
    --from ${VALIDATOR_KEY}

# Vote on proposal
virtengine tx gov vote 1 yes --from ${VALIDATOR_KEY}
```

### 4. Verify Configuration

```bash
# Check archival configuration
virtengine query veid archival-config

# Test archive backend health
virtengine query veid archive-health

# Check eligibility criteria
virtengine query veid archival-eligibility
```

## Daily Operations

### Morning Routine

```bash
# 1. Check archival metrics
virtengine query veid archive-metrics

# 2. Review overnight archival operations
virtengine query veid archive-operations \
    --since $(date -d '24 hours ago' +%s) \
    --limit 100

# 3. Check for failed archives
virtengine query veid failed-archives --limit 10

# 4. Verify integrity check results
virtengine query veid integrity-check-results \
    --since $(date -d '24 hours ago' +%s)

# 5. Monitor restore queue
virtengine query veid restore-queue
```

### Monitoring Dashboard

Create a monitoring script (`monitor-archival.sh`):

```bash
#!/bin/bash

while true; do
    clear
    echo "=== VirtEngine Archival System Monitor ==="
    echo ""
    
    echo "Archive Metrics:"
    virtengine query veid archive-metrics --output json | jq .
    echo ""
    
    echo "Restore Queue:"
    virtengine query veid restore-queue --output json | jq .
    echo ""
    
    echo "Recent Failures:"
    virtengine query veid failed-archives --limit 5 --output json | jq .
    echo ""
    
    echo "Storage Usage:"
    du -sh /var/lib/virtengine/archives/
    echo ""
    
    echo "Last updated: $(date)"
    sleep 60
done
```

Run monitoring dashboard:

```bash
chmod +x monitor-archival.sh
./monitor-archival.sh
```

### Handling Restore Requests

When users request restored archives:

```bash
# 1. Check restore request
virtengine query veid restore-request <request-id>

# 2. Monitor restore progress
watch -n 10 'virtengine query veid restore-status <archive-id>'

# 3. Verify completion
virtengine query veid archive-status <archive-id>

# 4. Check restored data availability
virtengine query veid restored-archives --limit 10
```

## Monitoring

### Key Metrics

#### 1. Archive Creation Rate

```bash
# Get archives created in last 24 hours
virtengine query veid archive-operations \
    --since $(date -d '24 hours ago' +%s) \
    --count-only

# Expected: 50-200 archives/day for medium validator
```

#### 2. Storage Usage

```bash
# Check total archive storage
virtengine query veid archive-metrics | jq '.total_bytes'

# Check by tier
virtengine query veid archive-metrics | jq '.archives_by_tier'

# Local filesystem usage
df -h /var/lib/virtengine/archives/
```

#### 3. Restore Performance

```bash
# Average restore time by tier
virtengine query veid restore-metrics

# Current restore queue depth
virtengine query veid restore-queue | jq '.pending_restoration'
```

### Alert Thresholds

Set up alerts for:

| Metric | Warning | Critical |
|--------|---------|----------|
| Failed archives | > 5 | > 20 |
| Restore queue depth | > 50 | > 100 |
| Storage usage (local) | > 80% | > 95% |
| Integrity check failures | > 2 | > 10 |
| Archive creation failures | > 5/hour | > 20/hour |

### Prometheus Metrics

Export metrics to Prometheus:

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'virtengine-archival'
    static_configs:
      - targets: ['localhost:26660']
    metrics_path: '/metrics'
    relabel_configs:
      - source_labels: [__name__]
        regex: 'virtengine_veid_archive_.*'
        action: keep
```

Example metrics:
- `virtengine_veid_archive_total`
- `virtengine_veid_archive_by_tier`
- `virtengine_veid_archive_restore_duration_seconds`
- `virtengine_veid_archive_failed_total`

## Troubleshooting

### Issue: Archives Stuck in "Archiving" Status

**Symptoms:**
```bash
virtengine query veid archive-status abc123
# status: "archiving"
# created_at: 2 hours ago
```

**Diagnosis:**
```bash
# Check backend connectivity
virtengine query veid archive-health

# Check logs
tail -f ~/.virtengine/virtengine.log | grep -i archive

# Check backend credentials
aws s3 ls s3://virtengine-archives-${VALIDATOR_MONIKER}/ || echo "Failed"
```

**Resolution:**
```bash
# Retry archive operation
virtengine tx veid retry-archive abc123 --from ${VALIDATOR_KEY}

# If still stuck, cancel and re-archive
virtengine tx veid cancel-archive abc123 --from ${VALIDATOR_KEY}
virtengine tx veid archive-artifact --content-hash <hash> --from ${VALIDATOR_KEY}
```

### Issue: Restore Timeout

**Symptoms:**
```
Error: restore request timed out
```

**Diagnosis:**
```bash
# Check restore status
virtengine query veid restore-status <archive-id>

# Check backend restore job
aws glacier describe-job \
    --account-id - \
    --vault-name virtengine-archives \
    --job-id <job-id>
```

**Resolution:**
```bash
# For Glacier Deep Archive, restoration can take 12-48 hours
# Use expedited tier for urgent restores:
virtengine tx veid restore-archive \
    --archive-id <archive-id> \
    --restore-tier expedited \
    --from ${VALIDATOR_KEY}
```

### Issue: Integrity Check Failures

**Symptoms:**
```
Error: integrity verification failed: checksum mismatch
```

**Diagnosis:**
```bash
# Get archive details
virtengine query veid archive <archive-id>

# Check if corruption is widespread
virtengine query veid failed-integrity-checks --limit 100

# Verify backend storage
aws s3api head-object \
    --bucket virtengine-archives-${VALIDATOR_MONIKER} \
    --key archives/<archive-id>.dat
```

**Resolution:**
```bash
# If single archive corrupted, restore from backup
virtengine tx veid restore-from-backup \
    --archive-id <archive-id> \
    --backup-source s3://virtengine-archives-backup/ \
    --from ${VALIDATOR_KEY}

# If multiple archives corrupted, investigate storage backend
# Check for hardware failures, network issues, etc.
```

### Issue: High Archive Failure Rate

**Symptoms:**
```bash
virtengine query veid archive-metrics | jq '.archives_by_status.failed'
# Output: 50 (expected: < 5)
```

**Diagnosis:**
```bash
# Get failure reasons
virtengine query veid failed-archives --limit 50 | jq '.[].error'

# Common causes:
# - Backend connectivity issues
# - Insufficient permissions
# - Rate limiting
# - Storage quota exceeded
```

**Resolution:**
```bash
# Check backend credentials
aws sts get-caller-identity

# Check IAM permissions
aws iam get-user-policy --user-name virtengine-archival --policy-name ArchiveAccess

# Increase rate limits
aws s3api put-bucket-request-payment \
    --bucket virtengine-archives-${VALIDATOR_MONIKER} \
    --request-payment-configuration Payer=Requester

# Increase storage quota (if applicable)
```

## Maintenance

### Weekly Maintenance

```bash
#!/bin/bash
# weekly-maintenance.sh

echo "Starting weekly archival maintenance..."

# 1. Purge expired archives
echo "Purging expired archives..."
virtengine tx veid purge-expired-archives --from ${VALIDATOR_KEY}

# 2. Run integrity checks on sample
echo "Running integrity checks..."
virtengine tx veid check-archive-integrity \
    --sample-size 100 \
    --from ${VALIDATOR_KEY}

# 3. Clean up expired restores
echo "Cleaning up expired restores..."
virtengine tx veid cleanup-expired-restores --from ${VALIDATOR_KEY}

# 4. Update archival index
echo "Rebuilding archival index..."
virtengine tx veid rebuild-archive-index --from ${VALIDATOR_KEY}

# 5. Generate report
echo "Generating weekly report..."
virtengine query veid archive-report \
    --since $(date -d '7 days ago' +%s) \
    --output json > weekly-archive-report.json

echo "Maintenance complete. Report saved to weekly-archive-report.json"
```

Schedule with cron:

```bash
# Run every Sunday at 2 AM
0 2 * * 0 /home/validator/weekly-maintenance.sh
```

### Monthly Maintenance

```bash
#!/bin/bash
# monthly-maintenance.sh

echo "Starting monthly archival maintenance..."

# 1. Full integrity check
echo "Running full integrity check (may take hours)..."
virtengine tx veid check-archive-integrity \
    --full \
    --from ${VALIDATOR_KEY}

# 2. Archive rotation (move standard -> glacier -> deep archive)
echo "Rotating archives to deeper tiers..."
virtengine tx veid rotate-archive-tiers --from ${VALIDATOR_KEY}

# 3. Generate compliance report
echo "Generating compliance report..."
virtengine query veid compliance-report \
    --since $(date -d '30 days ago' +%s) \
    --output json > monthly-compliance-report.json

# 4. Backup archive encryption keys
echo "Backing up encryption keys..."
virtengine keys export archive-master > archive-master-backup.json
gpg --encrypt --recipient validator@virtengine.com archive-master-backup.json
rm archive-master-backup.json

# 5. Test disaster recovery
echo "Testing disaster recovery procedures..."
TEST_ARCHIVE=$(virtengine query veid archives --limit 1 | jq -r '.[0].archive_id')
virtengine tx veid restore-archive \
    --archive-id ${TEST_ARCHIVE} \
    --restore-tier expedited \
    --from ${VALIDATOR_KEY}
sleep 300  # Wait for restoration
virtengine query veid get-archived-artifact ${TEST_ARCHIVE}

echo "Monthly maintenance complete."
```

### Key Rotation

Rotate archive encryption keys annually:

```bash
#!/bin/bash
# rotate-archive-keys.sh

echo "Starting archive key rotation..."

# 1. Generate new archive key
NEW_KEY_ID="v2.0.0"
virtengine keys derive-archive-key \
    --master-key archive-master \
    --key-id ${NEW_KEY_ID} \
    --output archive-key-${NEW_KEY_ID}.json

# 2. Submit governance proposal to activate new key
virtengine tx gov submit-proposal param-change \
    --title "Rotate Archive Encryption Key" \
    --description "Activate new archive encryption key ${NEW_KEY_ID}" \
    --changes '[{
        "subspace": "veid",
        "key": "ArchiveEncryptionKeyID",
        "value": "'${NEW_KEY_ID}'"
    }]' \
    --deposit 10000000uvirt \
    --from ${VALIDATOR_KEY}

# 3. Wait for proposal to pass and activate
echo "Waiting for governance proposal..."

# 4. Verify new key is active
virtengine query veid archival-config | jq '.archive_encryption_key_id'

# 5. (Optional) Re-encrypt existing archives with new key
# This is a background process that can take days/weeks
virtengine tx veid reencrypt-archives \
    --old-key-id "v1.0.0" \
    --new-key-id ${NEW_KEY_ID} \
    --from ${VALIDATOR_KEY}

echo "Key rotation initiated. Monitor progress with:"
echo "virtengine query veid reencryption-progress"
```

### Backup Strategy

Set up automated backups:

```bash
#!/bin/bash
# backup-archives.sh

BACKUP_BUCKET="s3://virtengine-archives-backup-${VALIDATOR_MONIKER}"
SOURCE_BUCKET="s3://virtengine-archives-${VALIDATOR_MONIKER}"

# 1. Sync archives to backup bucket
aws s3 sync ${SOURCE_BUCKET}/ ${BACKUP_BUCKET}/ \
    --storage-class GLACIER \
    --exclude "*.tmp"

# 2. Export on-chain archival index
virtengine query veid export-archive-index > archive-index.json
aws s3 cp archive-index.json ${BACKUP_BUCKET}/indexes/$(date +%Y%m%d)-index.json

# 3. Backup encryption keys (encrypted)
gpg --encrypt --recipient validator@virtengine.com archive-key.json
aws s3 cp archive-key.json.gpg ${BACKUP_BUCKET}/keys/archive-key.json.gpg

echo "Backup complete to ${BACKUP_BUCKET}"
```

Schedule daily backups:

```bash
# Run every day at 3 AM
0 3 * * * /home/validator/backup-archives.sh
```

## Performance Tuning

### Optimize Archival Rate

```bash
# Increase max archives per block (requires governance)
virtengine tx gov submit-proposal param-change \
    --title "Increase Archival Throughput" \
    --description "Increase max archives per block from 10 to 50" \
    --changes '[{
        "subspace": "veid",
        "key": "ArchivalConfig.max_archives_per_block",
        "value": 50
    }]' \
    --deposit 10000000uvirt \
    --from ${VALIDATOR_KEY}
```

### Optimize Storage Costs

```bash
# Adjust archival age threshold (archive sooner)
virtengine tx gov submit-proposal param-change \
    --title "Reduce Archival Age Threshold" \
    --description "Archive artifacts after 30 days instead of 90" \
    --changes '[{
        "subspace": "veid",
        "key": "ArchivalConfig.min_age_for_archival_seconds",
        "value": 2592000
    }]' \
    --deposit 10000000uvirt \
    --from ${VALIDATOR_KEY}
```

### Optimize Restore Performance

```bash
# Use expedited restore tier by default for critical operations
# Configure in app.toml:
[archive]
default_restore_tier = "expedited"  # Faster but more expensive
```

## Compliance Audits

### Generate Compliance Report

```bash
# GDPR compliance report
virtengine query veid compliance-report \
    --regulation GDPR \
    --since $(date -d '1 year ago' +%s) \
    --output json > gdpr-compliance-report.json

# HIPAA compliance report
virtengine query veid compliance-report \
    --regulation HIPAA \
    --since $(date -d '6 years ago' +%s) \
    --output json > hipaa-compliance-report.json

# CCPA compliance report
virtengine query veid compliance-report \
    --regulation CCPA \
    --since $(date -d '1 year ago' +%s) \
    --output json > ccpa-compliance-report.json
```

### Audit Trail

Query audit trail for specific archive:

```bash
# Get all operations on an archive
virtengine query veid archive-audit-trail <archive-id>

# Output:
# - Archive creation timestamp and block
# - All restore requests (with requester addresses)
# - All access operations
# - Integrity check results
# - Legal hold events
# - Deletion events
```

## Security Best Practices

1. **Encrypt Keys at Rest**: Use gpg or HSM for archive keys
2. **Restrict Access**: Limit who can submit archival transactions
3. **Monitor Access Patterns**: Alert on suspicious restore requests
4. **Regular Integrity Checks**: Run weekly integrity verification
5. **Backup Regularly**: Daily backup to separate storage provider
6. **Rotate Keys**: Annual key rotation
7. **Audit Logs**: Review audit logs monthly
8. **Disaster Recovery Testing**: Quarterly DR tests

## Additional Resources

- [Data Retention Policy](./data-retention-policy.md)
- [Compliance Framework](./compliance-framework.md)
- [API Documentation](./api-docs/archival-api.md)
- [Troubleshooting Guide](./troubleshooting/archival-issues.md)

## Support

For issues not covered in this guide:

- **Community Forum**: https://forum.virtengine.com/c/archival
- **Discord**: #validator-support
- **Email**: support@virtengine.com
- **Emergency**: On-call validator support (SLA customers only)
