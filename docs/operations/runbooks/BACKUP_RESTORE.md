# Backup and Restore Procedures

**Version:** 1.0.0  
**Last Updated:** 2026-01-30  
**Owner:** SRE Team

---

## Table of Contents

1. [Backup Strategy Overview](#backup-strategy-overview)
2. [Validator Node Backups](#validator-node-backups)
3. [Provider Daemon Backups](#provider-daemon-backups)
4. [Database Backups](#database-backups)
5. [Configuration Backups](#configuration-backups)
6. [Key Management Backups](#key-management-backups)
7. [Restore Procedures](#restore-procedures)
8. [Backup Verification](#backup-verification)
9. [Backup Monitoring](#backup-monitoring)

---

## Backup Strategy Overview

### Backup Matrix

| Component | Type | Frequency | Retention | Location | Encryption |
|-----------|------|-----------|-----------|----------|------------|
| Validator keys | Full | Initial + rotation | Forever | Offline + Cloud | AES-256 |
| Chain data snapshot | Full | 6 hours | 7 days | S3/GCS | AES-256 |
| Chain data incremental | Incremental | 1 hour | 24 hours | S3/GCS | AES-256 |
| Provider keys | Full | Initial + rotation | Forever | Offline + Cloud | AES-256 |
| Provider state | Full | 1 hour | 7 days | S3/GCS | AES-256 |
| PostgreSQL | Full | Daily | 30 days | S3/GCS | AES-256 |
| PostgreSQL WAL | Continuous | Continuous | 7 days | S3/GCS | AES-256 |
| Configuration | On change | On change | 90 days | Git + S3 | AES-256 |
| Monitoring data | Full | Daily | 30 days | S3/GCS | Optional |

### Backup Locations

| Tier | Location | Purpose | Access |
|------|----------|---------|--------|
| Tier 1 | Local SSD | Fast recovery | Immediate |
| Tier 2 | Regional S3/GCS | Standard recovery | < 5 min |
| Tier 3 | Cross-region S3/GCS | DR | < 30 min |
| Tier 4 | Offline/Air-gapped | Keys only | Manual |

### Encryption Standards

- **Algorithm**: AES-256-GCM
- **Key management**: AWS KMS / GCP KMS / HashiCorp Vault
- **Key rotation**: Annual or on compromise

---

## Validator Node Backups

### Critical Files to Backup

```
~/.virtengine/
├── config/
│   ├── priv_validator_key.json    # CRITICAL - validator signing key
│   ├── node_key.json              # Node identity key
│   ├── config.toml                # Configuration
│   └── app.toml                   # App configuration
├── data/
│   ├── priv_validator_state.json  # Signing state (to prevent double-sign)
│   └── ...                        # Chain data (can be regenerated)
└── keyring-file/                  # Operator keys
    └── *.info
```

### Key Backup Procedure

> ⚠️ **CRITICAL**: Validator keys must be backed up before the validator goes online. Loss of keys means loss of staked funds if jailed.

```bash
#!/bin/bash
# backup-validator-keys.sh

set -e

BACKUP_DIR="/secure/backups/validator-keys"
DATE=$(date +%Y%m%d)
BACKUP_FILE="validator-keys-${DATE}.tar.gz"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Stop validator to ensure consistent state
sudo systemctl stop virtengine

# Create archive of critical keys
tar -czf "/tmp/${BACKUP_FILE}" \
    ~/.virtengine/config/priv_validator_key.json \
    ~/.virtengine/config/node_key.json \
    ~/.virtengine/data/priv_validator_state.json \
    ~/.virtengine/keyring-file/

# Encrypt with GPG
gpg --encrypt --recipient backup@virtengine.com "/tmp/${BACKUP_FILE}"

# Move to backup location
mv "/tmp/${BACKUP_FILE}.gpg" "${BACKUP_DIR}/"

# Upload to secure cloud storage
gsutil cp "${BACKUP_DIR}/${BACKUP_FILE}.gpg" gs://virtengine-secure-backups/validator-keys/

# Clean up
rm -f "/tmp/${BACKUP_FILE}"
shred -u ~/.virtengine/config/priv_validator_key.json.tmp 2>/dev/null || true

# Restart validator
sudo systemctl start virtengine

echo "Backup completed: ${BACKUP_FILE}.gpg"
echo "Uploaded to: gs://virtengine-secure-backups/validator-keys/"
```

### Chain Data Snapshot

```bash
#!/bin/bash
# backup-chain-data.sh

set -e

SNAPSHOT_DIR="/data/snapshots"
DATE=$(date +%Y%m%d-%H%M)
SNAPSHOT_FILE="virtengine-snapshot-${DATE}.tar.lz4"

# Create snapshot directory
mkdir -p "$SNAPSHOT_DIR"

# Option 1: Hot backup (no downtime, may have inconsistencies)
# Requires virtengine running with snapshot-interval configured

# Check latest snapshot
LATEST_HEIGHT=$(curl -s http://localhost:26657/status | jq -r '.result.sync_info.latest_block_height')
echo "Current height: $LATEST_HEIGHT"

# Export snapshot
virtengine export --height $LATEST_HEIGHT > "${SNAPSHOT_DIR}/genesis-export-${DATE}.json"

# Option 2: Cold backup (requires brief downtime)
# Stop validator
sudo systemctl stop virtengine

# Create compressed archive
tar -cf - ~/.virtengine/data | lz4 -9 > "${SNAPSHOT_DIR}/${SNAPSHOT_FILE}"

# Restart immediately
sudo systemctl start virtengine

# Upload to cloud storage
gsutil cp "${SNAPSHOT_DIR}/${SNAPSHOT_FILE}" gs://virtengine-snapshots/

# Clean up old snapshots (keep 7 days)
find "$SNAPSHOT_DIR" -name "*.tar.lz4" -mtime +7 -delete

echo "Snapshot completed: ${SNAPSHOT_FILE}"
```

### Automated Backup Cron

```bash
# /etc/cron.d/virtengine-backup

# Chain data snapshot every 6 hours
0 */6 * * * validator /opt/scripts/backup-chain-data.sh >> /var/log/backup-chain.log 2>&1

# Validator state backup (prevents double-sign on restart)
*/5 * * * * validator cp ~/.virtengine/data/priv_validator_state.json /backup/priv_validator_state.json

# Key backup on first of month (reminder to verify)
0 0 1 * * validator echo "Monthly key backup verification required" | mail -s "Validator Key Backup Reminder" ops@virtengine.com
```

---

## Provider Daemon Backups

### Critical Files

```
~/.provider-daemon/
├── config.yaml              # Configuration
├── keyring/                 # Provider keys
│   ├── provider.info
│   └── provider-encryption.info
├── tls/                     # TLS certificates
│   ├── server.crt
│   └── server.key
└── state/                   # Provider state (if not using external store)
    └── workloads.db
```

### Provider Backup Script

```bash
#!/bin/bash
# backup-provider.sh

set -e

BACKUP_DIR="/secure/backups/provider"
DATE=$(date +%Y%m%d-%H%M)
BACKUP_FILE="provider-backup-${DATE}.tar.gz"

mkdir -p "$BACKUP_DIR"

# Create backup
tar -czf "/tmp/${BACKUP_FILE}" \
    ~/.provider-daemon/config.yaml \
    ~/.provider-daemon/keyring/ \
    ~/.provider-daemon/tls/

# Encrypt
gpg --encrypt --recipient backup@virtengine.com "/tmp/${BACKUP_FILE}"

# Move to backup location
mv "/tmp/${BACKUP_FILE}.gpg" "${BACKUP_DIR}/"

# Upload to cloud
gsutil cp "${BACKUP_DIR}/${BACKUP_FILE}.gpg" gs://virtengine-secure-backups/provider/

# Also backup pending usage reports
if [ -d "/var/lib/provider-daemon/failed_reports" ]; then
    tar -czf "/tmp/failed-reports-${DATE}.tar.gz" /var/lib/provider-daemon/failed_reports/
    gsutil cp "/tmp/failed-reports-${DATE}.tar.gz" gs://virtengine-backups/provider-state/
fi

# Clean up
rm -f "/tmp/${BACKUP_FILE}" "/tmp/failed-reports-${DATE}.tar.gz"

echo "Provider backup completed: ${BACKUP_FILE}.gpg"
```

### Provider State Backup

```bash
#!/bin/bash
# backup-provider-state.sh
# Run hourly to preserve workload state

BACKUP_DIR="/backup/provider-state"
DATE=$(date +%Y%m%d-%H%M)

mkdir -p "$BACKUP_DIR"

# Export current workload state
provider-daemon workloads export --output "${BACKUP_DIR}/workloads-${DATE}.json"

# Export pending usage reports
provider-daemon usage export --pending --output "${BACKUP_DIR}/pending-usage-${DATE}.json"

# Upload to cloud
gsutil rsync "$BACKUP_DIR" gs://virtengine-backups/provider-state/

# Clean up old backups (keep 7 days)
find "$BACKUP_DIR" -name "*.json" -mtime +7 -delete
```

---

## Database Backups

### PostgreSQL Full Backup

```bash
#!/bin/bash
# backup-postgres-full.sh

set -e

BACKUP_DIR="/backup/postgresql"
DATE=$(date +%Y%m%d)
BACKUP_FILE="virtengine-db-${DATE}.dump"

mkdir -p "$BACKUP_DIR"

# Create custom format dump (supports parallel restore)
pg_dump \
    -h localhost \
    -U virtengine \
    -d virtengine_db \
    -Fc \
    -f "${BACKUP_DIR}/${BACKUP_FILE}"

# Compress
lz4 -9 "${BACKUP_DIR}/${BACKUP_FILE}" "${BACKUP_DIR}/${BACKUP_FILE}.lz4"
rm "${BACKUP_DIR}/${BACKUP_FILE}"

# Upload to cloud
gsutil cp "${BACKUP_DIR}/${BACKUP_FILE}.lz4" gs://virtengine-backups/postgresql/

# Clean up old backups (keep 30 days)
find "$BACKUP_DIR" -name "*.dump.lz4" -mtime +30 -delete

# Clean up old cloud backups
gsutil ls -l gs://virtengine-backups/postgresql/*.lz4 | \
    awk -v cutoff=$(date -d '30 days ago' +%Y-%m-%d) '$2 < cutoff {print $3}' | \
    xargs -I {} gsutil rm {}

echo "PostgreSQL backup completed: ${BACKUP_FILE}.lz4"
```

### PostgreSQL WAL Archiving

```bash
# postgresql.conf
archive_mode = on
archive_command = 'gsutil cp %p gs://virtengine-backups/postgresql/wal/%f'
archive_timeout = 300  # Archive every 5 minutes even if not full

# recovery.conf (for point-in-time recovery)
restore_command = 'gsutil cp gs://virtengine-backups/postgresql/wal/%f %p'
```

### WAL Archive Management

```bash
#!/bin/bash
# cleanup-wal-archives.sh

# Remove WAL files older than 7 days
gsutil ls -l gs://virtengine-backups/postgresql/wal/ | \
    awk -v cutoff=$(date -d '7 days ago' +%Y-%m-%d) '$2 < cutoff {print $3}' | \
    xargs -I {} gsutil rm {}

echo "WAL archive cleanup completed"
```

---

## Configuration Backups

### Git-based Configuration Backup

```bash
# All configuration should be in Git repository
# /opt/virtengine-config/

git init /opt/virtengine-config

# Add configuration files
cp ~/.virtengine/config/config.toml /opt/virtengine-config/validator/
cp ~/.virtengine/config/app.toml /opt/virtengine-config/validator/
cp ~/.provider-daemon/config.yaml /opt/virtengine-config/provider/

# Commit changes
cd /opt/virtengine-config
git add -A
git commit -m "Configuration update $(date +%Y%m%d-%H%M)"
git push origin main
```

### Configuration Change Detection

```bash
#!/bin/bash
# config-backup-on-change.sh

CONFIG_REPO="/opt/virtengine-config"
WATCH_DIRS=(
    "$HOME/.virtengine/config"
    "$HOME/.provider-daemon"
)

cd "$CONFIG_REPO"

# Check for changes
CHANGES=0
for dir in "${WATCH_DIRS[@]}"; do
    if [ -d "$dir" ]; then
        rsync -a --delete "$dir/" "${CONFIG_REPO}/$(basename $dir)/"
    fi
done

# Commit if changes
if ! git diff --quiet; then
    git add -A
    git commit -m "Auto-backup: Configuration changed $(date +%Y%m%d-%H%M)"
    git push origin main
    echo "Configuration changes backed up"
fi
```

### Systemd Timer for Config Backup

```ini
# /etc/systemd/system/config-backup.timer
[Unit]
Description=Configuration backup timer

[Timer]
OnCalendar=*:0/15
Persistent=true

[Install]
WantedBy=timers.target
```

```ini
# /etc/systemd/system/config-backup.service
[Unit]
Description=Configuration backup service

[Service]
Type=oneshot
User=validator
ExecStart=/opt/scripts/config-backup-on-change.sh
```

---

## Key Management Backups

### Key Backup Best Practices

> ⚠️ **CRITICAL SECURITY**: Keys are the most sensitive assets. Follow these practices strictly.

1. **Never store keys unencrypted**
2. **Use multiple backup locations** (cloud + offline)
3. **Test restores quarterly**
4. **Maintain access control lists**
5. **Document key holders**

### Offline Key Backup Procedure

```bash
#!/bin/bash
# offline-key-backup.sh
# Run manually on secure workstation

echo "=== Offline Key Backup Procedure ==="
echo "This should be performed on an air-gapped machine"
echo ""

# Verify we're on secure machine
read -p "Confirm this is an air-gapped machine (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Aborting. Use air-gapped machine for offline backups."
    exit 1
fi

# Mount encrypted USB drive
echo "Insert encrypted USB drive and press Enter..."
read

# Create encrypted backup
BACKUP_DATE=$(date +%Y%m%d)
BACKUP_NAME="virtengine-keys-offline-${BACKUP_DATE}"

# Create tarball
tar -czf "/tmp/${BACKUP_NAME}.tar.gz" \
    validator-keys/ \
    provider-keys/ \
    key-inventory.txt

# Encrypt with passphrase (not GPG key - for offline access)
openssl enc -aes-256-cbc -salt \
    -in "/tmp/${BACKUP_NAME}.tar.gz" \
    -out "/mnt/usb/${BACKUP_NAME}.tar.gz.enc"

# Create verification hash
sha256sum "/mnt/usb/${BACKUP_NAME}.tar.gz.enc" > "/mnt/usb/${BACKUP_NAME}.sha256"

# Secure delete temporary file
shred -u "/tmp/${BACKUP_NAME}.tar.gz"

# Unmount
umount /mnt/usb

echo ""
echo "Offline backup completed: ${BACKUP_NAME}.tar.gz.enc"
echo "Store USB drive in secure location"
echo "Record passphrase separately from USB drive"
```

### Key Inventory

```markdown
# Key Inventory

## Validator Keys

| Key Type | Fingerprint | Created | Backup Locations | Key Holders |
|----------|-------------|---------|------------------|-------------|
| Validator Consensus | ABCD1234... | 2026-01-01 | AWS KMS, Offline USB #1 | Alice, Bob |
| Validator Node | EFGH5678... | 2026-01-01 | AWS KMS, Offline USB #1 | Alice, Bob |
| Operator Account | IJKL9012... | 2026-01-01 | AWS KMS, Offline USB #1 | Alice, Bob, Carol |

## Provider Keys

| Key Type | Fingerprint | Created | Backup Locations | Key Holders |
|----------|-------------|---------|------------------|-------------|
| Provider Signing | MNOP3456... | 2026-01-01 | GCP KMS, Offline USB #2 | Dave, Eve |
| Provider Encryption | QRST7890... | 2026-01-01 | GCP KMS, Offline USB #2 | Dave, Eve |

## Backup Verification Log

| Date | Backup Location | Verified By | Result |
|------|-----------------|-------------|--------|
| 2026-01-15 | AWS KMS | Alice | ✅ Pass |
| 2026-01-15 | Offline USB #1 | Bob | ✅ Pass |
```

---

## Restore Procedures

### Validator Key Restore

```bash
#!/bin/bash
# restore-validator-keys.sh

set -e

echo "=== Validator Key Restore ==="
echo ""

# CRITICAL: Ensure old validator is STOPPED
read -p "Confirm the original validator is STOPPED (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "DANGER: Running two validators with same keys causes SLASHING"
    echo "Stop the original validator first!"
    exit 1
fi

BACKUP_FILE=$1
if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file.tar.gz.gpg>"
    exit 1
fi

# Create restore directory
RESTORE_DIR=$(mktemp -d)

# Decrypt backup
gpg --decrypt "$BACKUP_FILE" | tar -xzf - -C "$RESTORE_DIR"

# Verify contents
echo "Backup contents:"
ls -la "$RESTORE_DIR"

read -p "Proceed with restore? (yes/no): " proceed
if [ "$proceed" != "yes" ]; then
    rm -rf "$RESTORE_DIR"
    exit 0
fi

# Stop validator if running
sudo systemctl stop virtengine 2>/dev/null || true

# Backup existing keys (if any)
if [ -f ~/.virtengine/config/priv_validator_key.json ]; then
    mv ~/.virtengine/config/priv_validator_key.json \
       ~/.virtengine/config/priv_validator_key.json.old
fi

# Restore keys
cp "$RESTORE_DIR"/.virtengine/config/priv_validator_key.json ~/.virtengine/config/
cp "$RESTORE_DIR"/.virtengine/config/node_key.json ~/.virtengine/config/
cp "$RESTORE_DIR"/.virtengine/data/priv_validator_state.json ~/.virtengine/data/

# Restore keyring
cp -r "$RESTORE_DIR"/.virtengine/keyring-file/* ~/.virtengine/keyring-file/

# Set permissions
chmod 600 ~/.virtengine/config/priv_validator_key.json
chmod 600 ~/.virtengine/config/node_key.json

# Clean up
rm -rf "$RESTORE_DIR"

echo ""
echo "Keys restored successfully"
echo "Start validator with: sudo systemctl start virtengine"
```

### Chain Data Restore

```bash
#!/bin/bash
# restore-chain-data.sh

set -e

SNAPSHOT_URL=$1
if [ -z "$SNAPSHOT_URL" ]; then
    echo "Usage: $0 <snapshot-url-or-path>"
    exit 1
fi

echo "=== Chain Data Restore ==="

# Stop validator
sudo systemctl stop virtengine

# Backup existing data (in case restore fails)
mv ~/.virtengine/data ~/.virtengine/data.old

# Create new data directory
mkdir -p ~/.virtengine/data

# Download and extract snapshot
if [[ "$SNAPSHOT_URL" == http* ]] || [[ "$SNAPSHOT_URL" == gs://* ]]; then
    echo "Downloading snapshot..."
    gsutil cp "$SNAPSHOT_URL" /tmp/snapshot.tar.lz4
    SNAPSHOT_FILE="/tmp/snapshot.tar.lz4"
else
    SNAPSHOT_FILE="$SNAPSHOT_URL"
fi

echo "Extracting snapshot..."
lz4 -d "$SNAPSHOT_FILE" | tar -xf - -C ~/.virtengine/

# Restore priv_validator_state.json from backup (CRITICAL)
if [ -f ~/.virtengine/data.old/priv_validator_state.json ]; then
    cp ~/.virtengine/data.old/priv_validator_state.json ~/.virtengine/data/
    echo "Restored priv_validator_state.json"
fi

# Clean up
rm -f /tmp/snapshot.tar.lz4

echo ""
echo "Chain data restored successfully"
echo "Start validator with: sudo systemctl start virtengine"
```

### Provider Restore

```bash
#!/bin/bash
# restore-provider.sh

set -e

BACKUP_FILE=$1
if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file.tar.gz.gpg>"
    exit 1
fi

echo "=== Provider Restore ==="

# Stop provider daemon
sudo systemctl stop provider-daemon 2>/dev/null || true

# Decrypt and extract
RESTORE_DIR=$(mktemp -d)
gpg --decrypt "$BACKUP_FILE" | tar -xzf - -C "$RESTORE_DIR"

# Backup existing config
if [ -f ~/.provider-daemon/config.yaml ]; then
    mv ~/.provider-daemon/config.yaml ~/.provider-daemon/config.yaml.old
fi

# Restore configuration
cp "$RESTORE_DIR"/.provider-daemon/config.yaml ~/.provider-daemon/

# Restore keys
cp -r "$RESTORE_DIR"/.provider-daemon/keyring/* ~/.provider-daemon/keyring/

# Restore TLS certificates
cp "$RESTORE_DIR"/.provider-daemon/tls/* ~/.provider-daemon/tls/

# Set permissions
chmod 600 ~/.provider-daemon/keyring/*
chmod 600 ~/.provider-daemon/tls/server.key

# Clean up
rm -rf "$RESTORE_DIR"

echo ""
echo "Provider restored successfully"
echo "Start with: sudo systemctl start provider-daemon"
```

### Database Restore

```bash
#!/bin/bash
# restore-postgres.sh

set -e

BACKUP_FILE=$1
TARGET_TIME=$2  # Optional: for PITR

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup-file.dump.lz4> [target-time]"
    exit 1
fi

echo "=== PostgreSQL Restore ==="

# Stop services using database
kubectl scale deployment virtengine-api --replicas=0

# Drop and recreate database
psql -U postgres -c "DROP DATABASE IF EXISTS virtengine_db;"
psql -U postgres -c "CREATE DATABASE virtengine_db OWNER virtengine;"

# Decompress and restore
if [[ "$BACKUP_FILE" == *.lz4 ]]; then
    lz4 -d "$BACKUP_FILE" /tmp/restore.dump
    DUMP_FILE="/tmp/restore.dump"
else
    DUMP_FILE="$BACKUP_FILE"
fi

# Restore
pg_restore \
    -h localhost \
    -U virtengine \
    -d virtengine_db \
    -j 4 \
    "$DUMP_FILE"

# Point-in-time recovery (if target time specified)
if [ -n "$TARGET_TIME" ]; then
    echo "Applying WAL for PITR to $TARGET_TIME..."
    # Configure recovery
    cat > /var/lib/postgresql/data/recovery.signal << EOF
restore_command = 'gsutil cp gs://virtengine-backups/postgresql/wal/%f %p'
recovery_target_time = '$TARGET_TIME'
recovery_target_action = 'promote'
EOF
    
    sudo systemctl restart postgresql
    # Wait for recovery
    while psql -c "SELECT pg_is_in_recovery();" | grep -q "t"; do
        sleep 5
    done
fi

# Clean up
rm -f /tmp/restore.dump

# Restart services
kubectl scale deployment virtengine-api --replicas=3

echo ""
echo "Database restored successfully"
```

---

## Backup Verification

### Daily Verification Script

```bash
#!/bin/bash
# verify-backups.sh

set -e

REPORT_FILE="/var/log/backup-verification-$(date +%Y%m%d).log"

echo "=== Backup Verification Report ===" | tee "$REPORT_FILE"
echo "Date: $(date)" | tee -a "$REPORT_FILE"
echo "" | tee -a "$REPORT_FILE"

# Check validator key backup
echo "1. Validator Key Backups" | tee -a "$REPORT_FILE"
LATEST_KEY=$(gsutil ls -l gs://virtengine-secure-backups/validator-keys/ | tail -1)
echo "   Latest: $LATEST_KEY" | tee -a "$REPORT_FILE"

# Check chain snapshots
echo "2. Chain Snapshots" | tee -a "$REPORT_FILE"
LATEST_SNAPSHOT=$(gsutil ls -l gs://virtengine-snapshots/ | tail -1)
echo "   Latest: $LATEST_SNAPSHOT" | tee -a "$REPORT_FILE"

# Check database backups
echo "3. Database Backups" | tee -a "$REPORT_FILE"
LATEST_DB=$(gsutil ls -l gs://virtengine-backups/postgresql/*.dump.lz4 | tail -1)
echo "   Latest: $LATEST_DB" | tee -a "$REPORT_FILE"

# Check WAL archives
echo "4. WAL Archives" | tee -a "$REPORT_FILE"
WAL_COUNT=$(gsutil ls gs://virtengine-backups/postgresql/wal/ | wc -l)
echo "   Count: $WAL_COUNT files" | tee -a "$REPORT_FILE"

# Verify backup age
echo "" | tee -a "$REPORT_FILE"
echo "=== Age Verification ===" | tee -a "$REPORT_FILE"

check_age() {
    local name=$1
    local max_hours=$2
    local timestamp=$3
    
    if [ -z "$timestamp" ]; then
        echo "   $name: MISSING" | tee -a "$REPORT_FILE"
        return 1
    fi
    
    local age_hours=$(( ($(date +%s) - $(date -d "$timestamp" +%s)) / 3600 ))
    
    if [ $age_hours -gt $max_hours ]; then
        echo "   $name: STALE ($age_hours hours old, max $max_hours)" | tee -a "$REPORT_FILE"
        return 1
    else
        echo "   $name: OK ($age_hours hours old)" | tee -a "$REPORT_FILE"
        return 0
    fi
}

# Run age checks
ERRORS=0
check_age "Chain snapshot" 6 "$(echo $LATEST_SNAPSHOT | awk '{print $2}')" || ((ERRORS++))
check_age "Database backup" 24 "$(echo $LATEST_DB | awk '{print $2}')" || ((ERRORS++))

echo "" | tee -a "$REPORT_FILE"
if [ $ERRORS -gt 0 ]; then
    echo "=== VERIFICATION FAILED: $ERRORS issues ===" | tee -a "$REPORT_FILE"
    # Send alert
    mail -s "Backup Verification Failed" ops@virtengine.com < "$REPORT_FILE"
    exit 1
else
    echo "=== VERIFICATION PASSED ===" | tee -a "$REPORT_FILE"
fi
```

### Monthly Restore Test

```bash
#!/bin/bash
# monthly-restore-test.sh

echo "=== Monthly Restore Test ==="
echo "This test verifies backup integrity by performing actual restores"
echo ""

# Create isolated test environment
TEST_DIR=$(mktemp -d)
cd "$TEST_DIR"

# Test 1: Key backup restore
echo "Test 1: Validator key backup restore"
LATEST_KEY=$(gsutil ls gs://virtengine-secure-backups/validator-keys/*.gpg | tail -1)
gsutil cp "$LATEST_KEY" ./key-backup.tar.gz.gpg
gpg --decrypt key-backup.tar.gz.gpg | tar -tzf - > /dev/null
echo "   PASSED: Key backup is valid"

# Test 2: Database restore
echo "Test 2: Database restore"
LATEST_DB=$(gsutil ls gs://virtengine-backups/postgresql/*.dump.lz4 | tail -1)
gsutil cp "$LATEST_DB" ./db-backup.dump.lz4
lz4 -d db-backup.dump.lz4 db-backup.dump
createdb test_restore_$(date +%Y%m%d)
pg_restore -d test_restore_$(date +%Y%m%d) db-backup.dump
TABLES=$(psql -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';" test_restore_$(date +%Y%m%d))
dropdb test_restore_$(date +%Y%m%d)
echo "   PASSED: Database restore completed ($TABLES tables)"

# Cleanup
cd /
rm -rf "$TEST_DIR"

echo ""
echo "=== All restore tests passed ==="
```

---

## Backup Monitoring

### Prometheus Metrics

```yaml
# Custom backup metrics
- job_name: 'backup-metrics'
  static_configs:
    - targets: ['localhost:9092']
  metrics_path: /backup/metrics
```

### Backup Exporter Script

```bash
#!/bin/bash
# backup-metrics-exporter.sh
# Exposes backup metrics for Prometheus

METRICS_FILE="/var/lib/prometheus/backup_metrics.prom"

# Calculate metrics
SNAPSHOT_AGE=$(( ($(date +%s) - $(stat -c %Y /backup/snapshots/latest.tar.lz4 2>/dev/null || echo 0)) ))
DB_BACKUP_AGE=$(( ($(date +%s) - $(stat -c %Y /backup/postgresql/latest.dump.lz4 2>/dev/null || echo 0)) ))
SNAPSHOT_SIZE=$(stat -c %s /backup/snapshots/latest.tar.lz4 2>/dev/null || echo 0)
DB_SIZE=$(stat -c %s /backup/postgresql/latest.dump.lz4 2>/dev/null || echo 0)

# Write metrics
cat > "$METRICS_FILE" << EOF
# HELP backup_snapshot_age_seconds Age of latest chain snapshot in seconds
# TYPE backup_snapshot_age_seconds gauge
backup_snapshot_age_seconds $SNAPSHOT_AGE

# HELP backup_database_age_seconds Age of latest database backup in seconds
# TYPE backup_database_age_seconds gauge
backup_database_age_seconds $DB_BACKUP_AGE

# HELP backup_snapshot_size_bytes Size of latest chain snapshot in bytes
# TYPE backup_snapshot_size_bytes gauge
backup_snapshot_size_bytes $SNAPSHOT_SIZE

# HELP backup_database_size_bytes Size of latest database backup in bytes
# TYPE backup_database_size_bytes gauge
backup_database_size_bytes $DB_SIZE

# HELP backup_last_success_timestamp Unix timestamp of last successful backup
# TYPE backup_last_success_timestamp gauge
backup_last_success_timestamp{type="snapshot"} $(stat -c %Y /backup/snapshots/latest.tar.lz4 2>/dev/null || echo 0)
backup_last_success_timestamp{type="database"} $(stat -c %Y /backup/postgresql/latest.dump.lz4 2>/dev/null || echo 0)
EOF
```

### Alert Rules

```yaml
# backup_alerts.yml
groups:
  - name: backups
    rules:
      - alert: BackupSnapshotStale
        expr: backup_snapshot_age_seconds > 21600  # 6 hours
        for: 30m
        labels:
          severity: warning
        annotations:
          summary: "Chain snapshot backup is stale"
          
      - alert: BackupDatabaseStale
        expr: backup_database_age_seconds > 86400  # 24 hours
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "Database backup is stale"
          
      - alert: BackupFailed
        expr: increase(backup_failures_total[1h]) > 0
        labels:
          severity: critical
        annotations:
          summary: "Backup job failed"
```

---

## Appendix: Backup Checklist

### Daily Checklist (Automated)

- [ ] Chain snapshot created
- [ ] Database backup created
- [ ] WAL files archived
- [ ] Backup verification passed
- [ ] Metrics updated

### Weekly Checklist

- [ ] Review backup job logs
- [ ] Check storage usage
- [ ] Verify cloud uploads
- [ ] Test backup accessibility

### Monthly Checklist

- [ ] Perform restore test
- [ ] Verify key backups
- [ ] Update key inventory
- [ ] Review retention policies

### Quarterly Checklist

- [ ] Full DR exercise
- [ ] Rotate backup encryption keys
- [ ] Audit backup access logs
- [ ] Update documentation

---

**Document Owner:** SRE Team  
**Next Review:** 2026-04-30
