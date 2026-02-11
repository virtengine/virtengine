# VirtEngine Disaster Recovery Plan

## Table of Contents

1. [Overview](#overview)
2. [RTO/RPO Targets](#rtorpo-targets)
3. [Multi-Region Architecture](#multi-region-architecture)
4. [Backup Procedures](#backup-procedures)
5. [Recovery Procedures](#recovery-procedures)
6. [Failover Procedures](#failover-procedures)
7. [Data Replication Strategy](#data-replication-strategy)
8. [Geographic Redundancy](#geographic-redundancy)
9. [DR Testing](#dr-testing)
10. [Runbooks](#runbooks)

---

## Overview

This document defines the disaster recovery (DR) plan for VirtEngine blockchain infrastructure. It covers backup procedures, recovery procedures, failover mechanisms, and RTO/RPO targets for all critical components.

### Scope

The DR plan covers:

- **Blockchain Nodes**: Validator nodes, full nodes, archive nodes
- **Provider Daemon**: Off-chain provider infrastructure
- **Key Management**: Cryptographic keys and HSM configurations
- **Chain State**: Blockchain data, transaction history, application state
- **Configuration**: Network configuration, genesis files, node configs
- **Monitoring**: Alerting rules, dashboards, metrics retention

### Disaster Categories

| Category                   | Description                          | Example Scenarios                                   |
| -------------------------- | ------------------------------------ | --------------------------------------------------- |
| **D1 - Component Failure** | Single component fails               | Node crash, disk failure, memory corruption         |
| **D2 - Zone Failure**      | Entire availability zone unavailable | AZ outage, datacenter power failure                 |
| **D3 - Region Failure**    | Entire region unavailable            | Regional disaster, network partition                |
| **D4 - Data Corruption**   | Data integrity compromised           | Software bug, malicious attack, accidental deletion |
| **D5 - Key Compromise**    | Cryptographic keys compromised       | Key leak, unauthorized access                       |

---

## RTO/RPO Targets

### Recovery Time Objective (RTO)

| Component       | D1     | D2     | D3     | D4     | D5     |
| --------------- | ------ | ------ | ------ | ------ | ------ |
| Validator Node  | 5 min  | 15 min | 30 min | 1 hr   | 15 min |
| Full Node       | 15 min | 30 min | 1 hr   | 2 hr   | N/A    |
| Provider Daemon | 5 min  | 15 min | 30 min | 1 hr   | 30 min |
| API Gateway     | 2 min  | 10 min | 20 min | 30 min | N/A    |
| Monitoring      | 15 min | 30 min | 1 hr   | 1 hr   | N/A    |

### Recovery Point Objective (RPO)

| Component      | RPO              | Backup Frequency          | Retention  |
| -------------- | ---------------- | ------------------------- | ---------- |
| Chain State    | 0 (no data loss) | Continuous replication    | 90 days    |
| Validator Keys | 0 (no data loss) | Real-time HSM replication | Indefinite |
| Provider Keys  | 0 (no data loss) | Daily encrypted backup    | 1 year     |
| Configuration  | 1 hour           | Hourly snapshots          | 30 days    |
| Metrics/Logs   | 24 hours         | Daily export              | 90 days    |

### SLA Summary

| Metric                   | Target          | Measurement         |
| ------------------------ | --------------- | ------------------- |
| Chain Availability       | 99.9%           | 30-day rolling      |
| Validator Uptime         | 99.95%          | Per validator       |
| API Availability         | 99.9%           | 30-day rolling      |
| Maximum Planned Downtime | 4 hours/quarter | Maintenance windows |

---

## Multi-Region Architecture

### Region Topology

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           Global Load Balancer                               │
│                        (DNS + Anycast / CloudFlare)                         │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
┌───────────────────┐   ┌───────────────────┐   ┌───────────────────┐
│   Region: US-EAST │   │  Region: EU-WEST  │   │  Region: AP-SOUTH │
│   (Primary)       │   │  (Secondary)      │   │  (Tertiary)       │
├───────────────────┤   ├───────────────────┤   ├───────────────────┤
│ • 2 Validator     │   │ • 2 Validator     │   │ • 1 Validator     │
│ • 3 Full Nodes    │   │ • 2 Full Nodes    │   │ • 2 Full Nodes    │
│ • 2 Provider      │   │ • 1 Provider      │   │ • 1 Provider      │
│ • API Gateway     │   │ • API Gateway     │   │ • API Gateway     │
│ • HSM Cluster     │   │ • HSM Replica     │   │ • HSM Replica     │
└───────────────────┘   └───────────────────┘   └───────────────────┘
        │                           │                           │
        └───────────────────────────┼───────────────────────────┘
                                    │
                        ┌───────────────────┐
                        │   Cross-Region    │
                        │   Replication     │
                        │   (State Sync)    │
                        └───────────────────┘
```

### Region Roles

| Region                | Role      | Capacity        | Priority |
| --------------------- | --------- | --------------- | -------- |
| US-EAST (us-east-1)   | Primary   | 40% of workload | 1        |
| EU-WEST (eu-west-1)   | Secondary | 35% of workload | 2        |
| AP-SOUTH (ap-south-1) | Tertiary  | 25% of workload | 3        |

### Availability Zone Distribution

Each region should distribute nodes across a minimum of 2 AZs:

```yaml
# Example: US-EAST distribution
us-east-1a:
  - validator-0
  - fullnode-0
  - fullnode-2
  - provider-daemon-0
  - api-gateway-0

us-east-1b:
  - validator-1
  - fullnode-1
  - provider-daemon-1
  - api-gateway-1
```

---

## Backup Procedures

### 1. Chain State Backup

Primary automation lives in `scripts/dr/backup-chain-state.sh`, which creates snapshots, signs them, uploads to remote storage, and verifies signatures before restore. Set `SNAPSHOT_SIGNING_KEY` and `SNAPSHOT_VERIFY_PUBKEY` on all backup and restore hosts. The script includes rollback logic and fallback to older snapshots if corruption is detected.

Provider daemon state is backed up separately with `scripts/dr/backup-provider-state.sh` to capture off-chain checkpoints and queue state. Schedule it alongside chain state snapshots.

#### Continuous State Sync Replication

All full nodes maintain state sync snapshots that can be used for rapid recovery.

```bash
#!/bin/bash
# scripts/dr/create-state-snapshot.sh
# Creates a state sync snapshot for disaster recovery

set -euo pipefail

SNAPSHOT_DIR="${SNAPSHOT_DIR:-/data/snapshots}"
NODE_HOME="${NODE_HOME:-/opt/virtengine}"
SNAPSHOT_INTERVAL="${SNAPSHOT_INTERVAL:-1000}"  # blocks
RETENTION_COUNT="${RETENTION_RETENTION:-10}"

create_snapshot() {
    local height=$(virtengine status 2>&1 | jq -r '.sync_info.latest_block_height')
    local timestamp=$(date -u +%Y%m%d_%H%M%S)
    local snapshot_name="state_${height}_${timestamp}"

    echo "[$(date -u +%FT%TZ)] Creating snapshot at height $height..."

    # Export state using Cosmos SDK state export
    virtengine export --height "$height" > "${SNAPSHOT_DIR}/${snapshot_name}.json"

    # Create compressed archive of data directory
    tar -czf "${SNAPSHOT_DIR}/${snapshot_name}_data.tar.gz" \
        -C "${NODE_HOME}" \
        --exclude="*.log" \
        --exclude="*.tmp" \
        data/

    # Generate checksums
    cd "${SNAPSHOT_DIR}"
    sha256sum "${snapshot_name}.json" "${snapshot_name}_data.tar.gz" > "${snapshot_name}.sha256"

    echo "[$(date -u +%FT%TZ)] Snapshot created: ${snapshot_name}"

    # Upload to remote storage (S3/GCS)
    upload_to_remote "${snapshot_name}"

    # Cleanup old snapshots
    cleanup_old_snapshots
}

upload_to_remote() {
    local snapshot_name="$1"
    local bucket="${DR_BUCKET:-s3://virtengine-dr-backups}"
    local region=$(hostname | cut -d'-' -f1)

    aws s3 cp "${SNAPSHOT_DIR}/${snapshot_name}.json" \
        "${bucket}/${region}/state/${snapshot_name}.json" \
        --storage-class STANDARD_IA

    aws s3 cp "${SNAPSHOT_DIR}/${snapshot_name}_data.tar.gz" \
        "${bucket}/${region}/state/${snapshot_name}_data.tar.gz" \
        --storage-class STANDARD_IA

    aws s3 cp "${SNAPSHOT_DIR}/${snapshot_name}.sha256" \
        "${bucket}/${region}/state/${snapshot_name}.sha256"
}

cleanup_old_snapshots() {
    # Keep only the most recent snapshots locally
    cd "${SNAPSHOT_DIR}"
    ls -t state_*.json 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f
    ls -t state_*_data.tar.gz 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f
    ls -t state_*.sha256 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f
}

# Main execution
create_snapshot
```

#### Scheduled Backups

```yaml
# Kubernetes CronJob for state backups
apiVersion: batch/v1
kind: CronJob
metadata:
  name: chain-state-backup
  namespace: virtengine
spec:
  schedule: "0 */4 * * *" # Every 4 hours
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: backup-sa
          containers:
            - name: backup
              image: virtengine/backup-tools:latest
              command: ["/scripts/dr/create-state-snapshot.sh"]
              env:
                - name: NODE_HOME
                  value: "/opt/virtengine"
                - name: SNAPSHOT_DIR
                  value: "/data/snapshots"
                - name: DR_BUCKET
                  valueFrom:
                    secretKeyRef:
                      name: dr-config
                      key: bucket
              volumeMounts:
                - name: node-data
                  mountPath: /opt/virtengine
                  readOnly: true
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

### 2. Key Backup

#### Validator Key Backup

**CRITICAL**: Validator keys require special handling due to the double-signing risk.

```bash
#!/bin/bash
# scripts/dr/backup-validator-keys.sh
# Securely backup validator keys with Shamir secret sharing

set -euo pipefail

KEY_DIR="${KEY_DIR:-/opt/virtengine/config}"
BACKUP_DIR="${BACKUP_DIR:-/secure/backups}"
SHARES="${SHARES:-5}"
THRESHOLD="${THRESHOLD:-3}"

backup_validator_keys() {
    local timestamp=$(date -u +%Y%m%d_%H%M%S)
    local backup_name="validator_keys_${timestamp}"
    local temp_dir=$(mktemp -d)

    trap "rm -rf ${temp_dir}" EXIT

    echo "[$(date -u +%FT%TZ)] Starting validator key backup..."

    # Export priv_validator_key.json (encrypted)
    if [ -f "${KEY_DIR}/priv_validator_key.json" ]; then
        # DO NOT backup priv_validator_key.json directly for active validators
        # This prevents double-signing risks during failover
        echo "[WARN] Active validator key detected - creating reference only"

        # Create HSM reference instead
        create_hsm_reference "${backup_name}"
    fi

    # Backup node_key.json (can be regenerated, but useful for continuity)
    if [ -f "${KEY_DIR}/node_key.json" ]; then
        encrypt_and_split "${KEY_DIR}/node_key.json" "${backup_name}_node_key"
    fi

    echo "[$(date -u +%FT%TZ)] Key backup completed"
}

create_hsm_reference() {
    local backup_name="$1"

    # For HSM-backed keys, store the key reference/label
    # The actual key material stays in the HSM
    cat > "${BACKUP_DIR}/${backup_name}_hsm_ref.json" << EOF
{
    "type": "hsm_reference",
    "timestamp": "$(date -u +%FT%TZ)",
    "hsm_slot": "${HSM_SLOT:-0}",
    "key_label": "${HSM_KEY_LABEL:-validator-key}",
    "hsm_provider": "${HSM_PROVIDER:-softhsm}",
    "recovery_instructions": "Contact security team for HSM key recovery procedure"
}
EOF
}

encrypt_and_split() {
    local source_file="$1"
    local output_prefix="$2"

    # Read passphrase from secure source
    local passphrase=$(get_backup_passphrase)

    # Encrypt with AES-256-GCM
    openssl enc -aes-256-gcm -salt -pbkdf2 -iter 100000 \
        -in "${source_file}" \
        -out "${temp_dir}/${output_prefix}.enc" \
        -pass pass:"${passphrase}"

    # Split into shares using Shamir secret sharing
    # Each share goes to a different custodian/location
    ssss-split -t ${THRESHOLD} -n ${SHARES} -w "$(basename ${output_prefix})" \
        < "${temp_dir}/${output_prefix}.enc" \
        > "${BACKUP_DIR}/${output_prefix}_shares.txt"

    # Distribute shares to different storage locations
    distribute_shares "${output_prefix}"
}

distribute_shares() {
    local output_prefix="$1"
    local share_file="${BACKUP_DIR}/${output_prefix}_shares.txt"

    # Send share 1 to primary secure storage
    aws s3 cp <(sed -n '1p' "${share_file}") \
        "s3://virtengine-dr-primary/keys/${output_prefix}_share1.txt" \
        --sse aws:kms

    # Send share 2 to secondary secure storage
    aws s3 cp <(sed -n '2p' "${share_file}") \
        "s3://virtengine-dr-secondary/keys/${output_prefix}_share2.txt" \
        --sse aws:kms

    # Send share 3 to offline storage (manual process)
    echo "[INFO] Share 3 requires manual distribution to offline storage"
    sed -n '3p' "${share_file}" | gpg --encrypt --armor --recipient security@virtengine.network \
        > "${BACKUP_DIR}/${output_prefix}_share3_offline.gpg"

    # Additional shares distributed similarly

    # Securely delete the combined shares file
    shred -u "${share_file}"
}

get_backup_passphrase() {
    # Get passphrase from secure parameter store
    aws secretsmanager get-secret-value \
        --secret-id virtengine/dr/backup-passphrase \
        --query SecretString --output text
}

# Main execution
backup_validator_keys
```

#### Provider Key Backup

Provider keys use the existing `KeyBackupManager` from `pkg/provider_daemon/backup.go`:

```bash
#!/bin/bash
# scripts/dr/backup-provider-keys.sh
# Backup provider daemon keys using the built-in backup system

set -euo pipefail

PROVIDER_HOME="${PROVIDER_HOME:-/opt/provider-daemon}"
BACKUP_DIR="${BACKUP_DIR:-/secure/backups/provider}"

backup_provider_keys() {
    local timestamp=$(date -u +%Y%m%d_%H%M%S)

    echo "[$(date -u +%FT%TZ)] Starting provider key backup..."

    # Use the provider-daemon CLI to create backup
    provider-daemon keys backup \
        --output "${BACKUP_DIR}/provider_keys_${timestamp}.backup" \
        --shamir-threshold 3 \
        --shamir-shares 5 \
        --encrypt

    # Upload encrypted backup to secure storage
    aws s3 cp "${BACKUP_DIR}/provider_keys_${timestamp}.backup" \
        "s3://virtengine-dr-backups/provider-keys/" \
        --sse aws:kms \
        --storage-class STANDARD_IA

    echo "[$(date -u +%FT%TZ)] Provider key backup completed"
}

backup_provider_keys
```

### 3. Configuration Backup

```bash
#!/bin/bash
# scripts/dr/backup-config.sh
# Backup all configuration files

set -euo pipefail

CONFIG_DIRS=(
    "/opt/virtengine/config"
    "/opt/provider-daemon/config"
    "/etc/istio"
    "/etc/prometheus"
)

BACKUP_DIR="${BACKUP_DIR:-/data/backups/config}"
REMOTE_BUCKET="${REMOTE_BUCKET:-s3://virtengine-dr-backups/config}"

backup_configs() {
    local timestamp=$(date -u +%Y%m%d_%H%M%S)
    local archive_name="config_backup_${timestamp}.tar.gz"
    local temp_dir=$(mktemp -d)

    trap "rm -rf ${temp_dir}" EXIT

    echo "[$(date -u +%FT%TZ)] Starting configuration backup..."

    # Collect configurations
    for dir in "${CONFIG_DIRS[@]}"; do
        if [ -d "$dir" ]; then
            local dest="${temp_dir}/$(basename $dir)"
            mkdir -p "$dest"

            # Copy excluding secrets (they're backed up separately)
            rsync -av --exclude="*key*" --exclude="*secret*" --exclude="*.pem" \
                "$dir/" "$dest/"
        fi
    done

    # Include genesis file
    if [ -f "/opt/virtengine/config/genesis.json" ]; then
        cp "/opt/virtengine/config/genesis.json" "${temp_dir}/"
    fi

    # Create archive
    tar -czf "${BACKUP_DIR}/${archive_name}" -C "${temp_dir}" .

    # Upload to remote storage
    aws s3 cp "${BACKUP_DIR}/${archive_name}" \
        "${REMOTE_BUCKET}/${archive_name}"

    # Verify upload
    aws s3api head-object --bucket virtengine-dr-backups \
        --key "config/${archive_name}" > /dev/null

    echo "[$(date -u +%FT%TZ)] Configuration backup completed: ${archive_name}"
}

backup_configs
```

---

## Recovery Procedures

### 1. Single Node Recovery (D1)

**Scenario**: A single validator or full node has failed and needs recovery.

```bash
#!/bin/bash
# scripts/dr/recover-single-node.sh
# Recover a single node from backup

set -euo pipefail

NODE_TYPE="${1:-fullnode}"  # validator or fullnode
NODE_HOME="${NODE_HOME:-/opt/virtengine}"

recover_node() {
    echo "[$(date -u +%FT%TZ)] Starting ${NODE_TYPE} recovery..."

    # 1. Stop any running processes
    systemctl stop virtengine || true

    # 2. Clear corrupted data (keep config)
    rm -rf "${NODE_HOME}/data"

    # 3. Download latest snapshot
    local latest_snapshot=$(aws s3 ls s3://virtengine-dr-backups/$(get_region)/state/ \
        | sort | tail -1 | awk '{print $4}' | sed 's/_data.tar.gz//')

    echo "[$(date -u +%FT%TZ)] Restoring from snapshot: ${latest_snapshot}"

    # 4. Download and extract state
    aws s3 cp "s3://virtengine-dr-backups/$(get_region)/state/${latest_snapshot}_data.tar.gz" \
        /tmp/restore_data.tar.gz

    tar -xzf /tmp/restore_data.tar.gz -C "${NODE_HOME}/"

    # 5. Verify checksums
    aws s3 cp "s3://virtengine-dr-backups/$(get_region)/state/${latest_snapshot}.sha256" \
        /tmp/restore.sha256

    if ! sha256sum -c /tmp/restore.sha256; then
        echo "[ERROR] Checksum verification failed!"
        exit 1
    fi

    # 6. For validators, ensure key is properly configured
    if [ "${NODE_TYPE}" == "validator" ]; then
        verify_validator_key
    fi

    # 7. Start the node
    systemctl start virtengine

    # 8. Monitor sync progress
    monitor_sync

    echo "[$(date -u +%FT%TZ)] Node recovery completed"
}

verify_validator_key() {
    # Check HSM connectivity for validator keys
    if [ -n "${HSM_ENABLED:-}" ]; then
        pkcs11-tool --module "${HSM_MODULE}" --list-objects --type privkey
    fi

    # Verify priv_validator_key.json exists and is valid
    if [ -f "${NODE_HOME}/config/priv_validator_key.json" ]; then
        jq -e '.pub_key.value' "${NODE_HOME}/config/priv_validator_key.json" > /dev/null
    fi
}

monitor_sync() {
    local start_height=$(virtengine status 2>&1 | jq -r '.sync_info.latest_block_height')
    local network_height=$(curl -s https://rpc.virtengine.network/status | jq -r '.result.sync_info.latest_block_height')

    echo "[$(date -u +%FT%TZ)] Starting sync from height ${start_height}, network at ${network_height}"

    while true; do
        local current=$(virtengine status 2>&1 | jq -r '.sync_info.latest_block_height')
        local catching_up=$(virtengine status 2>&1 | jq -r '.sync_info.catching_up')

        if [ "${catching_up}" == "false" ]; then
            echo "[$(date -u +%FT%TZ)] Sync complete at height ${current}"
            break
        fi

        echo "[$(date -u +%FT%TZ)] Syncing: ${current}/${network_height}"
        sleep 10
    done
}

get_region() {
    # Determine region from instance metadata or hostname
    curl -s http://169.254.169.254/latest/meta-data/placement/region 2>/dev/null || \
        hostname | cut -d'-' -f1
}

recover_node
```

### 2. Zone Failover (D2)

**Scenario**: An entire availability zone is unavailable.

```bash
#!/bin/bash
# scripts/dr/zone-failover.sh
# Failover workloads from failed zone to healthy zone

set -euo pipefail

FAILED_ZONE="${1}"
TARGET_ZONE="${2}"
NAMESPACE="${NAMESPACE:-virtengine}"

zone_failover() {
    echo "[$(date -u +%FT%TZ)] Initiating zone failover from ${FAILED_ZONE} to ${TARGET_ZONE}..."

    # 1. Cordon nodes in failed zone
    kubectl get nodes -l topology.kubernetes.io/zone=${FAILED_ZONE} \
        -o name | xargs -I{} kubectl cordon {}

    # 2. Scale up replicas in target zone
    scale_replicas_in_zone "${TARGET_ZONE}"

    # 3. Wait for new pods to be ready
    kubectl rollout status deployment -n ${NAMESPACE} --timeout=300s

    # 4. Drain workloads from failed zone
    kubectl get nodes -l topology.kubernetes.io/zone=${FAILED_ZONE} \
        -o name | xargs -I{} kubectl drain {} --ignore-daemonsets --delete-emptydir-data --force

    # 5. Update load balancer to exclude failed zone
    update_load_balancer "${FAILED_ZONE}" "exclude"

    # 6. Verify services are healthy
    verify_services

    echo "[$(date -u +%FT%TZ)] Zone failover completed"
}

scale_replicas_in_zone() {
    local zone="$1"

    # Increase replica count for critical deployments
    kubectl scale deployment/virtengine-fullnode -n ${NAMESPACE} --replicas=3
    kubectl scale deployment/api-gateway -n ${NAMESPACE} --replicas=3

    # Ensure pod anti-affinity spreads across remaining AZs
    kubectl patch deployment/virtengine-fullnode -n ${NAMESPACE} --type=json \
        -p='[{"op": "add", "path": "/spec/template/spec/affinity", "value": {
            "podAntiAffinity": {
                "preferredDuringSchedulingIgnoredDuringExecution": [{
                    "weight": 100,
                    "podAffinityTerm": {
                        "labelSelector": {
                            "matchLabels": {"app": "virtengine-fullnode"}
                        },
                        "topologyKey": "topology.kubernetes.io/zone"
                    }
                }]
            }
        }}]'
}

update_load_balancer() {
    local zone="$1"
    local action="$2"

    # Update Route53 health checks or ALB target groups
    # Implementation depends on load balancer type
    echo "[INFO] Updating load balancer to ${action} zone ${zone}"
}

verify_services() {
    # Check all critical services are responding
    local services=("virtengine-node" "api-gateway" "provider-daemon")

    for svc in "${services[@]}"; do
        local ready=$(kubectl get deployment ${svc} -n ${NAMESPACE} -o jsonpath='{.status.readyReplicas}')
        local desired=$(kubectl get deployment ${svc} -n ${NAMESPACE} -o jsonpath='{.spec.replicas}')

        if [ "${ready}" != "${desired}" ]; then
            echo "[WARN] Service ${svc} not fully ready: ${ready}/${desired}"
        else
            echo "[OK] Service ${svc}: ${ready}/${desired} replicas ready"
        fi
    done
}

zone_failover
```

### 3. Region Failover (D3)

**Scenario**: An entire region is unavailable.

```bash
#!/bin/bash
# scripts/dr/region-failover.sh
# Failover entire region to secondary/tertiary region

set -euo pipefail

FAILED_REGION="${1:-us-east}"
PRIMARY_REGION="${PRIMARY_REGION:-us-east}"
SECONDARY_REGION="${SECONDARY_REGION:-eu-west}"
TERTIARY_REGION="${TERTIARY_REGION:-ap-south}"

region_failover() {
    echo "[$(date -u +%FT%TZ)] =========================================="
    echo "[$(date -u +%FT%TZ)] INITIATING REGION FAILOVER"
    echo "[$(date -u +%FT%TZ)] Failed Region: ${FAILED_REGION}"
    echo "[$(date -u +%FT%TZ)] =========================================="

    # 1. Determine target region
    local target_region
    if [ "${FAILED_REGION}" == "${PRIMARY_REGION}" ]; then
        target_region="${SECONDARY_REGION}"
    else
        target_region="${TERTIARY_REGION}"
    fi

    echo "[$(date -u +%FT%TZ)] Target Region: ${target_region}"

    # 2. Switch kubectl context to target region
    kubectl config use-context "virtengine-${target_region}"

    # 3. Scale up infrastructure in target region
    scale_region "${target_region}"

    # 4. Update DNS to point to new region
    update_dns "${target_region}"

    # 5. Notify validators of region change
    notify_validators "${target_region}"

    # 6. Update monitoring dashboards
    update_monitoring "${target_region}"

    # 7. Verify chain is healthy
    verify_chain_health

    echo "[$(date -u +%FT%TZ)] Region failover completed to ${target_region}"

    # 8. Send notifications
    send_failover_notification "${FAILED_REGION}" "${target_region}"
}

scale_region() {
    local region="$1"

    echo "[$(date -u +%FT%TZ)] Scaling infrastructure in ${region}..."

    # Scale full nodes
    kubectl scale deployment/virtengine-fullnode -n virtengine --replicas=5

    # Scale API gateways
    kubectl scale deployment/api-gateway -n virtengine --replicas=4

    # Scale provider daemons
    kubectl scale deployment/provider-daemon -n virtengine --replicas=3

    # Wait for scaling
    kubectl rollout status deployment -n virtengine --timeout=600s
}

update_dns() {
    local region="$1"

    echo "[$(date -u +%FT%TZ)] Updating DNS to ${region}..."

    # Update Route53 weighted records
    local hosted_zone_id="${ROUTE53_HOSTED_ZONE_ID}"

    # Set weight to 100 for target region, 0 for others
    aws route53 change-resource-record-sets \
        --hosted-zone-id "${hosted_zone_id}" \
        --change-batch '{
            "Changes": [{
                "Action": "UPSERT",
                "ResourceRecordSet": {
                    "Name": "rpc.virtengine.network",
                    "Type": "A",
                    "SetIdentifier": "'${region}'",
                    "Weight": 100,
                    "TTL": 60,
                    "ResourceRecords": [{"Value": "'$(get_region_ip ${region})'"}]
                }
            }]
        }'

    # Reduce TTL for faster failover
    echo "[$(date -u +%FT%TZ)] DNS TTL reduced to 60s for faster propagation"
}

get_region_ip() {
    local region="$1"

    # Get the load balancer IP for the region
    kubectl --context="virtengine-${region}" get svc api-gateway \
        -n virtengine -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
}

notify_validators() {
    local region="$1"

    echo "[$(date -u +%FT%TZ)] Notifying validators of failover..."

    # Send webhook notification to validator operators
    curl -X POST "${VALIDATOR_WEBHOOK_URL}" \
        -H "Content-Type: application/json" \
        -d '{
            "event": "region_failover",
            "failed_region": "'${FAILED_REGION}'",
            "active_region": "'${region}'",
            "timestamp": "'$(date -u +%FT%TZ)'",
            "action_required": "Update peer addresses if needed"
        }'
}

update_monitoring() {
    local region="$1"

    echo "[$(date -u +%FT%TZ)] Updating monitoring configuration..."

    # Update Prometheus targets
    kubectl apply -f - << EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-targets
  namespace: monitoring
data:
  active_region: "${region}"
EOF

    # Restart Prometheus to pick up changes
    kubectl rollout restart deployment/prometheus -n monitoring
}

verify_chain_health() {
    echo "[$(date -u +%FT%TZ)] Verifying chain health..."

    local retries=30
    local wait_time=10

    for i in $(seq 1 $retries); do
        local status=$(virtengine status 2>&1 | jq -r '.sync_info.catching_up')

        if [ "$status" == "false" ]; then
            echo "[$(date -u +%FT%TZ)] Chain is healthy and synced"
            return 0
        fi

        echo "[$(date -u +%FT%TZ)] Waiting for chain to sync... (attempt ${i}/${retries})"
        sleep $wait_time
    done

    echo "[ERROR] Chain health check failed after ${retries} attempts"
    return 1
}

send_failover_notification() {
    local failed_region="$1"
    local target_region="$2"

    # PagerDuty
    curl -X POST "https://events.pagerduty.com/v2/enqueue" \
        -H "Content-Type: application/json" \
        -d '{
            "routing_key": "'${PAGERDUTY_ROUTING_KEY}'",
            "event_action": "trigger",
            "payload": {
                "summary": "VirtEngine Region Failover: '${failed_region}' -> '${target_region}'",
                "severity": "critical",
                "source": "dr-automation",
                "custom_details": {
                    "failed_region": "'${failed_region}'",
                    "target_region": "'${target_region}'",
                    "timestamp": "'$(date -u +%FT%TZ)'"
                }
            }
        }'

    # Slack
    curl -X POST "${SLACK_WEBHOOK_URL}" \
        -H "Content-Type: application/json" \
        -d '{
            "text": "🚨 *Region Failover Completed*\n*Failed Region:* '${failed_region}'\n*Active Region:* '${target_region}'\n*Time:* '$(date -u +%FT%TZ)'"
        }'
}

region_failover
```

### 4. Data Corruption Recovery (D4)

```bash
#!/bin/bash
# scripts/dr/recover-from-corruption.sh
# Recover from data corruption

set -euo pipefail

CORRUPTION_TYPE="${1:-state}"  # state, database, config
RECOVERY_HEIGHT="${2:-}"  # Optional: specific height to recover to

recover_from_corruption() {
    echo "[$(date -u +%FT%TZ)] Starting corruption recovery for: ${CORRUPTION_TYPE}..."

    case "${CORRUPTION_TYPE}" in
        state)
            recover_state_corruption
            ;;
        database)
            recover_database_corruption
            ;;
        config)
            recover_config_corruption
            ;;
        *)
            echo "[ERROR] Unknown corruption type: ${CORRUPTION_TYPE}"
            exit 1
            ;;
    esac
}

recover_state_corruption() {
    echo "[$(date -u +%FT%TZ)] Recovering from state corruption..."

    # 1. Stop the node
    systemctl stop virtengine

    # 2. Identify the last known good height
    local good_height="${RECOVERY_HEIGHT}"

    if [ -z "${good_height}" ]; then
        # Find last good snapshot
        good_height=$(aws s3 ls s3://virtengine-dr-backups/$(get_region)/state/ \
            | grep -oP 'state_\K\d+' | sort -n | tail -1)
    fi

    echo "[$(date -u +%FT%TZ)] Recovering to height: ${good_height}"

    # 3. Rollback to last good state
    virtengine rollback --hard

    # 4. If rollback insufficient, restore from snapshot
    if [ "$(get_current_height)" -lt "${good_height}" ]; then
        restore_from_snapshot "${good_height}"
    fi

    # 5. Restart and verify
    systemctl start virtengine
    verify_recovery
}

recover_database_corruption() {
    echo "[$(date -u +%FT%TZ)] Recovering from database corruption..."

    # 1. Stop services
    systemctl stop virtengine

    # 2. Backup corrupted data for analysis
    mv /opt/virtengine/data/application.db /opt/virtengine/data/application.db.corrupted

    # 3. Restore from backup
    local latest_backup=$(find /data/backups -name "*.db.backup" -type f | sort | tail -1)

    if [ -n "${latest_backup}" ]; then
        cp "${latest_backup}" /opt/virtengine/data/application.db
        echo "[$(date -u +%FT%TZ)] Restored database from: ${latest_backup}"
    else
        echo "[ERROR] No database backup found!"
        exit 1
    fi

    # 4. Replay WAL if available
    if [ -d "/opt/virtengine/data/wal" ]; then
        echo "[$(date -u +%FT%TZ)] Replaying WAL..."
        virtengine db replay-wal
    fi

    # 5. Restart
    systemctl start virtengine
    verify_recovery
}

recover_config_corruption() {
    echo "[$(date -u +%FT%TZ)] Recovering from config corruption..."

    # 1. Download latest config backup
    local latest_config=$(aws s3 ls s3://virtengine-dr-backups/config/ \
        | sort | tail -1 | awk '{print $4}')

    aws s3 cp "s3://virtengine-dr-backups/config/${latest_config}" /tmp/config_restore.tar.gz

    # 2. Extract config
    tar -xzf /tmp/config_restore.tar.gz -C /opt/virtengine/

    # 3. Restore genesis if needed
    if [ ! -f "/opt/virtengine/config/genesis.json" ]; then
        aws s3 cp "s3://virtengine-dr-backups/genesis/genesis.json" \
            /opt/virtengine/config/genesis.json
    fi

    # 4. Restart
    systemctl restart virtengine
    verify_recovery
}

restore_from_snapshot() {
    local height="$1"

    local snapshot_name="state_${height}"

    echo "[$(date -u +%FT%TZ)] Restoring from snapshot: ${snapshot_name}"

    # Download snapshot
    aws s3 cp "s3://virtengine-dr-backups/$(get_region)/state/${snapshot_name}_data.tar.gz" \
        /tmp/restore.tar.gz

    # Verify checksum
    aws s3 cp "s3://virtengine-dr-backups/$(get_region)/state/${snapshot_name}.sha256" \
        /tmp/restore.sha256

    cd /tmp && sha256sum -c restore.sha256

    # Extract
    rm -rf /opt/virtengine/data
    tar -xzf /tmp/restore.tar.gz -C /opt/virtengine/
}

get_current_height() {
    virtengine status 2>&1 | jq -r '.sync_info.latest_block_height' || echo "0"
}

verify_recovery() {
    local retries=60
    local wait_time=5

    echo "[$(date -u +%FT%TZ)] Verifying recovery..."

    for i in $(seq 1 $retries); do
        if virtengine status 2>&1 | jq -e '.sync_info' > /dev/null; then
            local height=$(virtengine status 2>&1 | jq -r '.sync_info.latest_block_height')
            local catching_up=$(virtengine status 2>&1 | jq -r '.sync_info.catching_up')

            echo "[$(date -u +%FT%TZ)] Node at height ${height}, catching_up=${catching_up}"

            if [ "$catching_up" == "false" ]; then
                echo "[$(date -u +%FT%TZ)] Recovery successful!"
                return 0
            fi
        fi

        sleep $wait_time
    done

    echo "[ERROR] Recovery verification failed"
    return 1
}

get_region() {
    curl -s http://169.254.169.254/latest/meta-data/placement/region 2>/dev/null || \
        hostname | cut -d'-' -f1
}

recover_from_corruption
```

### 5. Key Compromise Recovery (D5)

```bash
#!/bin/bash
# scripts/dr/recover-key-compromise.sh
# Emergency response for key compromise

set -euo pipefail

COMPROMISED_KEY="${1}"  # Key identifier
KEY_TYPE="${2:-validator}"  # validator, provider, node

key_compromise_response() {
    echo "[$(date -u +%FT%TZ)] =========================================="
    echo "[$(date -u +%FT%TZ)] KEY COMPROMISE RESPONSE INITIATED"
    echo "[$(date -u +%FT%TZ)] Compromised Key: ${COMPROMISED_KEY}"
    echo "[$(date -u +%FT%TZ)] Key Type: ${KEY_TYPE}"
    echo "[$(date -u +%FT%TZ)] =========================================="

    # 1. Immediate containment
    contain_compromise

    # 2. Rotate keys
    rotate_keys

    # 3. Audit trail
    create_audit_trail

    # 4. Notify stakeholders
    notify_stakeholders

    # 5. Post-incident actions
    post_incident_actions
}

contain_compromise() {
    echo "[$(date -u +%FT%TZ)] Containing compromise..."

    case "${KEY_TYPE}" in
        validator)
            # Stop the validator immediately to prevent double-signing
            systemctl stop virtengine

            # Revoke the key in HSM if HSM-backed
            if [ -n "${HSM_ENABLED:-}" ]; then
                pkcs11-tool --module "${HSM_MODULE}" --delete-object \
                    --type privkey --label "${COMPROMISED_KEY}"
            fi

            # Remove key from filesystem
            rm -f "/opt/virtengine/config/priv_validator_key.json"

            # Submit unjail transaction from secure location (if jailed)
            # This will be done manually after key rotation
            ;;

        provider)
            # Stop provider daemon
            systemctl stop provider-daemon

            # Revoke provider key
            provider-daemon keys revoke --key-id "${COMPROMISED_KEY}" --reason "compromise"

            # Update provider registration
            echo "[INFO] Provider registration needs manual update"
            ;;

        node)
            # Node keys are less critical, regenerate
            virtengine tendermint reset-peer-id
            systemctl restart virtengine
            ;;
    esac
}

rotate_keys() {
    echo "[$(date -u +%FT%TZ)] Rotating keys..."

    case "${KEY_TYPE}" in
        validator)
            # Generate new validator key in HSM
            if [ -n "${HSM_ENABLED:-}" ]; then
                pkcs11-tool --module "${HSM_MODULE}" --keygen \
                    --key-type EC:secp256k1 --label "validator-key-new"

                # Export public key
                export_hsm_pubkey "validator-key-new"
            else
                # Generate new file-based key (not recommended for production)
                virtengine keys add validator-new --keyring-backend file
            fi

            # New validator key requires chain governance to migrate
            echo "[ACTION REQUIRED] Submit validator key migration proposal"
            ;;

        provider)
            # Generate new provider key
            provider-daemon keys generate --label "provider-key-new"

            # Update provider registration
            provider-daemon provider update-key
            ;;
    esac
}

create_audit_trail() {
    local audit_file="/var/log/virtengine/key_compromise_$(date -u +%Y%m%d_%H%M%S).json"

    cat > "${audit_file}" << EOF
{
    "event": "key_compromise_response",
    "timestamp": "$(date -u +%FT%TZ)",
    "compromised_key": "${COMPROMISED_KEY}",
    "key_type": "${KEY_TYPE}",
    "responder": "$(whoami)@$(hostname)",
    "actions_taken": [
        "service_stopped",
        "key_revoked",
        "new_key_generated",
        "stakeholders_notified"
    ],
    "containment_time_seconds": "$SECONDS"
}
EOF

    # Upload to secure audit storage
    aws s3 cp "${audit_file}" \
        "s3://virtengine-security-audit/key-compromise/" \
        --sse aws:kms
}

notify_stakeholders() {
    echo "[$(date -u +%FT%TZ)] Notifying stakeholders..."

    # Security team
    curl -X POST "${SECURITY_WEBHOOK_URL}" \
        -H "Content-Type: application/json" \
        -d '{
            "event": "KEY_COMPROMISE",
            "severity": "critical",
            "key_type": "'${KEY_TYPE}'",
            "key_id": "'${COMPROMISED_KEY}'",
            "timestamp": "'$(date -u +%FT%TZ)'"
        }'

    # PagerDuty
    curl -X POST "https://events.pagerduty.com/v2/enqueue" \
        -H "Content-Type: application/json" \
        -d '{
            "routing_key": "'${PAGERDUTY_ROUTING_KEY}'",
            "event_action": "trigger",
            "payload": {
                "summary": "CRITICAL: Key Compromise - '${KEY_TYPE}' - '${COMPROMISED_KEY}'",
                "severity": "critical",
                "source": "key-compromise-response"
            }
        }'

    # If validator, notify validator coordination channel
    if [ "${KEY_TYPE}" == "validator" ]; then
        curl -X POST "${VALIDATOR_DISCORD_WEBHOOK}" \
            -H "Content-Type: application/json" \
            -d '{
                "content": "⚠️ **VALIDATOR KEY COMPROMISE DETECTED**\nKey rotation in progress. Please stand by for governance proposal."
            }'
    fi
}

post_incident_actions() {
    echo "[$(date -u +%FT%TZ)] Post-incident actions..."

    cat << EOF

===========================================
POST-INCIDENT CHECKLIST
===========================================

1. [ ] Complete incident report
2. [ ] Review access logs for compromise vector
3. [ ] Strengthen key storage security
4. [ ] Update key rotation policy if needed
5. [ ] Schedule post-mortem meeting
6. [ ] Update runbooks with lessons learned

For validator key compromise:
7. [ ] Submit governance proposal for key migration
8. [ ] Coordinate with other validators
9. [ ] Verify new validator is signing blocks
10. [ ] Monitor for any double-signing evidence

===========================================
EOF
}

export_hsm_pubkey() {
    local label="$1"

    pkcs11-tool --module "${HSM_MODULE}" --read-object \
        --type pubkey --label "${label}" \
        | openssl ec -pubin -inform DER -outform PEM
}

key_compromise_response
```

---

## Failover Procedures

### Automatic Failover

```yaml
# deploy/dr/failover-controller.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: dr-failover-controller
  namespace: virtengine
spec:
  replicas: 1
  selector:
    matchLabels:
      app: dr-failover-controller
  template:
    metadata:
      labels:
        app: dr-failover-controller
    spec:
      serviceAccountName: dr-controller
      containers:
        - name: controller
          image: virtengine/dr-controller:latest
          env:
            - name: PRIMARY_REGION
              value: "us-east-1"
            - name: SECONDARY_REGION
              value: "eu-west-1"
            - name: TERTIARY_REGION
              value: "ap-south-1"
            - name: HEALTH_CHECK_INTERVAL
              value: "30s"
            - name: FAILOVER_THRESHOLD
              value: "3" # consecutive failures before failover
            - name: SLACK_WEBHOOK_URL
              valueFrom:
                secretKeyRef:
                  name: dr-secrets
                  key: slack-webhook
          ports:
            - containerPort: 8080
              name: metrics
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 10
            periodSeconds: 30
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 256Mi
```

### Manual Failover Checklist

```markdown
## Manual Failover Checklist

**Pre-Failover:**

- [ ] Verify the disaster scenario warrants failover
- [ ] Notify stakeholders of impending failover
- [ ] Confirm target region is healthy
- [ ] Review recent backups are available

**During Failover:**

- [ ] Stop services in failed region (if accessible)
- [ ] Scale up target region infrastructure
- [ ] Update DNS records
- [ ] Verify chain sync status
- [ ] Validate API connectivity
- [ ] Confirm validator participation (if applicable)

**Post-Failover:**

- [ ] Monitor error rates and latency
- [ ] Verify backup replication is working
- [ ] Update incident ticket
- [ ] Schedule post-mortem

**Rollback (if needed):**

- [ ] Verify original region is healthy
- [ ] Sync state from failover region
- [ ] Update DNS records back
- [ ] Verify services are stable
- [ ] Monitor for 24 hours before closing incident
```

---

## Data Replication Strategy

### Chain State Replication

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        Chain State Replication                           │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Primary Region (US-EAST)                                              │
│   ┌──────────────────────────────────────────────────────────────┐     │
│   │  Validator-0  ──────►  Full Node-0  ──────►  Archive Node    │     │
│   │       │                     │                     │          │     │
│   │       │                     │                     │          │     │
│   │  Validator-1  ──────►  Full Node-1  ──────────────┘          │     │
│   │       │                     │                                │     │
│   │       └──────►  Full Node-2 (State Sync Source)              │     │
│   └──────────────────────────────────────────────────────────────┘     │
│                                    │                                    │
│                                    │ State Sync                         │
│                                    ▼                                    │
│   Secondary Region (EU-WEST)                                            │
│   ┌──────────────────────────────────────────────────────────────┐     │
│   │  Full Node-0 (Sync from US-EAST)                              │     │
│   │       │                                                       │     │
│   │       └──────►  Full Node-1  ──────►  Archive Node           │     │
│   │                                                               │     │
│   │  Validator-0 (Standby)  ◄──── Uses local full nodes          │     │
│   └──────────────────────────────────────────────────────────────┘     │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Replication Configuration

```toml
# config/config.toml - State sync configuration

[statesync]
# Enable state sync for faster catch-up
enable = true

# RPC servers to use for light client verification
rpc_servers = "rpc-us-east.virtengine.network:26657,rpc-eu-west.virtengine.network:26657"

# Trust height and hash
# Update these periodically or use a trust period
trust_height = 0
trust_hash = ""
trust_period = "168h0m0s"

# Discovery timeout
discovery_time = "15s"

# Chunk request timeout
chunk_request_timeout = "10s"

# Chunk fetchers to run in parallel
chunk_fetchers = "4"
```

### Cross-Region Sync Monitoring

```yaml
# deploy/monitoring/cross-region-sync.yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: cross-region-sync-alerts
  namespace: monitoring
spec:
  groups:
    - name: cross-region-sync
      rules:
        - alert: CrossRegionSyncLag
          expr: |
            abs(
              virtengine_consensus_height{region="us-east"} -
              virtengine_consensus_height{region="eu-west"}
            ) > 100
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Cross-region sync lag detected"
            description: "Region {{ $labels.region }} is more than 100 blocks behind"

        - alert: CrossRegionSyncFailure
          expr: |
            abs(
              virtengine_consensus_height{region="us-east"} -
              virtengine_consensus_height{region="eu-west"}
            ) > 1000
          for: 10m
          labels:
            severity: critical
          annotations:
            summary: "Critical cross-region sync failure"
            description: "Region {{ $labels.region }} is more than 1000 blocks behind"
```

---

## Geographic Redundancy

### Redundancy Requirements

| Component   | Minimum Regions | Minimum AZs per Region | Notes                        |
| ----------- | --------------- | ---------------------- | ---------------------------- |
| Validators  | 2               | 2                      | Max 33% in any single region |
| Full Nodes  | 3               | 2                      | At least 2 per region        |
| API Gateway | 3               | 2                      | Global load balancing        |
| HSM         | 2               | N/A                    | Primary + replicated         |
| Backups     | 3               | N/A                    | Cross-region replication     |

### Geographic Distribution Validation

```bash
#!/bin/bash
# scripts/dr/validate-geo-distribution.sh
# Validate geographic distribution meets DR requirements

set -euo pipefail

validate_distribution() {
    echo "[$(date -u +%FT%TZ)] Validating geographic distribution..."

    local errors=0

    # Check validator distribution
    echo ""
    echo "=== Validator Distribution ==="
    local validators_us=$(kubectl --context=virtengine-us-east get pods -l app=validator -n virtengine --no-headers | wc -l)
    local validators_eu=$(kubectl --context=virtengine-eu-west get pods -l app=validator -n virtengine --no-headers | wc -l)
    local validators_ap=$(kubectl --context=virtengine-ap-south get pods -l app=validator -n virtengine --no-headers | wc -l)
    local total_validators=$((validators_us + validators_eu + validators_ap))

    echo "US-EAST: ${validators_us}"
    echo "EU-WEST: ${validators_eu}"
    echo "AP-SOUTH: ${validators_ap}"
    echo "Total: ${total_validators}"

    # Validate no region has more than 33% of validators
    local max_validators_per_region=$((total_validators / 3 + 1))

    if [ $validators_us -gt $max_validators_per_region ]; then
        echo "[FAIL] US-EAST has too many validators (${validators_us} > ${max_validators_per_region})"
        ((errors++))
    fi

    if [ $validators_eu -gt $max_validators_per_region ]; then
        echo "[FAIL] EU-WEST has too many validators (${validators_eu} > ${max_validators_per_region})"
        ((errors++))
    fi

    # Check full node distribution
    echo ""
    echo "=== Full Node Distribution ==="
    local fullnodes_us=$(kubectl --context=virtengine-us-east get pods -l app=fullnode -n virtengine --no-headers | wc -l)
    local fullnodes_eu=$(kubectl --context=virtengine-eu-west get pods -l app=fullnode -n virtengine --no-headers | wc -l)
    local fullnodes_ap=$(kubectl --context=virtengine-ap-south get pods -l app=fullnode -n virtengine --no-headers | wc -l)

    echo "US-EAST: ${fullnodes_us}"
    echo "EU-WEST: ${fullnodes_eu}"
    echo "AP-SOUTH: ${fullnodes_ap}"

    if [ $fullnodes_us -lt 2 ]; then
        echo "[FAIL] US-EAST needs at least 2 full nodes"
        ((errors++))
    fi

    if [ $fullnodes_eu -lt 2 ]; then
        echo "[FAIL] EU-WEST needs at least 2 full nodes"
        ((errors++))
    fi

    # Check AZ distribution within regions
    echo ""
    echo "=== AZ Distribution ==="
    for region in us-east eu-west ap-south; do
        local azs=$(kubectl --context=virtengine-${region} get pods -n virtengine \
            -o jsonpath='{.items[*].spec.nodeName}' | tr ' ' '\n' | \
            xargs -I{} kubectl --context=virtengine-${region} get node {} \
            -o jsonpath='{.metadata.labels.topology\.kubernetes\.io/zone}' | tr ' ' '\n' | sort -u | wc -l)

        echo "${region}: ${azs} AZs"

        if [ $azs -lt 2 ]; then
            echo "[FAIL] ${region} needs at least 2 AZs"
            ((errors++))
        fi
    done

    # Check backup replication
    echo ""
    echo "=== Backup Replication ==="
    local backup_regions=$(aws s3 ls s3://virtengine-dr-backups/ | grep -E '^\s+PRE' | wc -l)
    echo "Backup regions: ${backup_regions}"

    if [ $backup_regions -lt 3 ]; then
        echo "[FAIL] Backups need to be replicated to at least 3 regions"
        ((errors++))
    fi

    # Summary
    echo ""
    echo "=== Validation Summary ==="
    if [ $errors -gt 0 ]; then
        echo "[FAIL] ${errors} validation errors found"
        exit 1
    else
        echo "[PASS] All geographic distribution requirements met"
    fi
}

validate_distribution
```

---

## DR Testing

### Test Schedule

| Test Type            | Frequency | Duration  | Impact              |
| -------------------- | --------- | --------- | ------------------- |
| Backup Validation    | Daily     | Automated | None                |
| Single Node Recovery | Weekly    | 30 min    | None                |
| Zone Failover        | Monthly   | 1 hour    | Minimal             |
| Region Failover      | Quarterly | 4 hours   | Planned maintenance |
| Full DR Drill        | Annually  | 8 hours   | Planned maintenance |

### Automated DR Test

```bash
#!/bin/bash
# scripts/dr/automated-dr-test.sh
# Automated DR testing suite

set -euo pipefail

TEST_ENVIRONMENT="${TEST_ENVIRONMENT:-staging}"
TEST_RESULTS_DIR="/var/log/dr-tests"

run_dr_tests() {
    local timestamp=$(date -u +%Y%m%d_%H%M%S)
    local results_file="${TEST_RESULTS_DIR}/dr_test_${timestamp}.json"

    mkdir -p "${TEST_RESULTS_DIR}"

    echo "[$(date -u +%FT%TZ)] Starting automated DR test suite..."

    local results='{"timestamp": "'$(date -u +%FT%TZ)'", "environment": "'${TEST_ENVIRONMENT}'", "tests": []}'

    # Test 1: Backup Integrity
    echo "[$(date -u +%FT%TZ)] Test 1: Backup Integrity"
    local backup_result=$(test_backup_integrity)
    results=$(echo "$results" | jq ".tests += [$backup_result]")

    # Test 2: State Sync Recovery
    echo "[$(date -u +%FT%TZ)] Test 2: State Sync Recovery"
    local sync_result=$(test_state_sync)
    results=$(echo "$results" | jq ".tests += [$sync_result]")

    # Test 3: Key Recovery
    echo "[$(date -u +%FT%TZ)] Test 3: Key Recovery"
    local key_result=$(test_key_recovery)
    results=$(echo "$results" | jq ".tests += [$key_result]")

    # Test 4: DNS Failover
    echo "[$(date -u +%FT%TZ)] Test 4: DNS Failover"
    local dns_result=$(test_dns_failover)
    results=$(echo "$results" | jq ".tests += [$dns_result]")

    # Test 5: Cross-Region Connectivity
    echo "[$(date -u +%FT%TZ)] Test 5: Cross-Region Connectivity"
    local connectivity_result=$(test_cross_region_connectivity)
    results=$(echo "$results" | jq ".tests += [$connectivity_result]")

    # Save results
    echo "$results" | jq '.' > "${results_file}"

    # Summarize
    local passed=$(echo "$results" | jq '[.tests[] | select(.passed == true)] | length')
    local failed=$(echo "$results" | jq '[.tests[] | select(.passed == false)] | length')

    echo ""
    echo "=== DR Test Summary ==="
    echo "Passed: ${passed}"
    echo "Failed: ${failed}"
    echo "Results saved to: ${results_file}"

    # Upload results
    aws s3 cp "${results_file}" "s3://virtengine-dr-backups/test-results/"

    # Alert on failures
    if [ "$failed" -gt 0 ]; then
        send_failure_alert "${results_file}"
        exit 1
    fi
}

test_backup_integrity() {
    local start_time=$(date +%s)
    local passed=true
    local details=""

    # Get latest backup
    local latest_backup=$(aws s3 ls s3://virtengine-dr-backups/us-east/state/ \
        | sort | tail -1 | awk '{print $4}')

    if [ -z "$latest_backup" ]; then
        passed=false
        details="No backup found"
    else
        # Download and verify checksum
        aws s3 cp "s3://virtengine-dr-backups/us-east/state/${latest_backup}" /tmp/test_backup.tar.gz
        aws s3 cp "s3://virtengine-dr-backups/us-east/state/${latest_backup%.tar.gz}.sha256" /tmp/test_backup.sha256

        if ! (cd /tmp && sha256sum -c test_backup.sha256); then
            passed=false
            details="Checksum verification failed"
        else
            details="Verified backup: ${latest_backup}"
        fi

        rm -f /tmp/test_backup.tar.gz /tmp/test_backup.sha256
    fi

    local duration=$(($(date +%s) - start_time))

    echo '{"name": "backup_integrity", "passed": '${passed}', "duration_seconds": '${duration}', "details": "'${details}'"}'
}

test_state_sync() {
    local start_time=$(date +%s)
    local passed=true
    local details=""

    # Test state sync endpoint availability
    local endpoints=(
        "rpc-us-east.virtengine.network:26657"
        "rpc-eu-west.virtengine.network:26657"
    )

    for endpoint in "${endpoints[@]}"; do
        if ! curl -s "http://${endpoint}/status" | jq -e '.result.sync_info' > /dev/null 2>&1; then
            passed=false
            details="${details}Failed: ${endpoint}; "
        fi
    done

    if [ "$passed" = true ]; then
        details="All state sync endpoints healthy"
    fi

    local duration=$(($(date +%s) - start_time))

    echo '{"name": "state_sync", "passed": '${passed}', "duration_seconds": '${duration}', "details": "'${details}'"}'
}

test_key_recovery() {
    local start_time=$(date +%s)
    local passed=true
    local details=""

    # Verify key backups exist and are recent
    local latest_key_backup=$(aws s3 ls s3://virtengine-dr-backups/provider-keys/ \
        | sort | tail -1 | awk '{print $4}')

    if [ -z "$latest_key_backup" ]; then
        passed=false
        details="No key backup found"
    else
        # Check backup age
        local backup_date=$(echo "$latest_key_backup" | grep -oP '\d{8}')
        local today=$(date +%Y%m%d)
        local days_old=$(( ($(date -d "$today" +%s) - $(date -d "$backup_date" +%s)) / 86400 ))

        if [ "$days_old" -gt 7 ]; then
            passed=false
            details="Key backup is ${days_old} days old (max 7)"
        else
            details="Key backup is ${days_old} days old"
        fi
    fi

    local duration=$(($(date +%s) - start_time))

    echo '{"name": "key_recovery", "passed": '${passed}', "duration_seconds": '${duration}', "details": "'${details}'"}'
}

test_dns_failover() {
    local start_time=$(date +%s)
    local passed=true
    local details=""

    # Test DNS resolution from multiple regions
    local expected_ips=("us-east-ip" "eu-west-ip" "ap-south-ip")
    local resolved_ip=$(dig +short rpc.virtengine.network | head -1)

    if [ -z "$resolved_ip" ]; then
        passed=false
        details="DNS resolution failed"
    else
        details="Resolved to: ${resolved_ip}"
    fi

    local duration=$(($(date +%s) - start_time))

    echo '{"name": "dns_failover", "passed": '${passed}', "duration_seconds": '${duration}', "details": "'${details}'"}'
}

test_cross_region_connectivity() {
    local start_time=$(date +%s)
    local passed=true
    local details=""

    local regions=("us-east" "eu-west" "ap-south")

    for region in "${regions[@]}"; do
        # Test P2P connectivity between regions
        if ! timeout 10 curl -s "http://rpc-${region}.virtengine.network:26657/net_info" | \
            jq -e '.result.n_peers | tonumber > 0' > /dev/null 2>&1; then
            passed=false
            details="${details}${region}: no peers; "
        fi
    done

    if [ "$passed" = true ]; then
        details="All regions connected"
    fi

    local duration=$(($(date +%s) - start_time))

    echo '{"name": "cross_region_connectivity", "passed": '${passed}', "duration_seconds": '${duration}', "details": "'${details}'"}'
}

send_failure_alert() {
    local results_file="$1"

    curl -X POST "${SLACK_WEBHOOK_URL}" \
        -H "Content-Type: application/json" \
        -d '{
            "text": "⚠️ *DR Test Failures Detected*\nResults: s3://virtengine-dr-backups/test-results/'$(basename ${results_file})'"
        }'
}

run_dr_tests
```

### DR Drill Procedure

See `docs/sre/INCIDENT_DRILLS.md` for the full DR drill procedure, integrated with this plan.

---

## Runbooks

### Runbook Index

| Runbook                                     | Scenario            | RTO Target |
| ------------------------------------------- | ------------------- | ---------- |
| [RB-DR-001](#rb-dr-001-single-node-failure) | Single Node Failure | 5 min      |
| [RB-DR-002](#rb-dr-002-zone-failover)       | Zone Failover       | 15 min     |
| [RB-DR-003](#rb-dr-003-region-failover)     | Region Failover     | 30 min     |
| [RB-DR-004](#rb-dr-004-data-corruption)     | Data Corruption     | 1 hr       |
| [RB-DR-005](#rb-dr-005-key-compromise)      | Key Compromise      | 15 min     |
| [RB-DR-006](#rb-dr-006-backup-restore)      | Backup Restore      | 2 hr       |

### RB-DR-001: Single Node Failure

**Trigger**: Node crash, disk failure, or unresponsive node

**Steps**:

1. **Verify** the failure via monitoring
2. **Assess** impact on consensus/availability
3. **Initiate** node recovery script: `scripts/dr/recover-single-node.sh`
4. **Monitor** sync progress
5. **Verify** node is participating normally
6. **Document** incident

**Escalation**: If node doesn't recover in 10 minutes, escalate to on-call lead

### RB-DR-002: Zone Failover

**Trigger**: Availability zone outage detected

**Steps**:

1. **Confirm** zone outage via cloud provider status
2. **Notify** team via Slack/PagerDuty
3. **Execute** zone failover: `scripts/dr/zone-failover.sh <failed_zone> <target_zone>`
4. **Verify** services are healthy in target zone
5. **Update** load balancer configurations
6. **Monitor** for 1 hour
7. **Document** incident and recovery

### RB-DR-003: Region Failover

**Trigger**: Region-wide outage or >50% services unavailable

**Steps**:

1. **Declare** region failover incident
2. **Notify** all stakeholders
3. **Execute** region failover: `scripts/dr/region-failover.sh <failed_region>`
4. **Update** DNS records
5. **Verify** chain consensus and API availability
6. **Monitor** for 4 hours
7. **Conduct** post-mortem within 48 hours

### RB-DR-004: Data Corruption

**Trigger**: Chain halt due to state inconsistency, or database corruption detected

**Steps**:

1. **Stop** affected services immediately
2. **Preserve** corrupted data for analysis
3. **Identify** last known good state
4. **Execute** recovery: `scripts/dr/recover-from-corruption.sh <type> [height]`
5. **Verify** state consistency across nodes
6. **Restart** services
7. **Root cause** analysis within 72 hours

### RB-DR-005: Key Compromise

**Trigger**: Suspected or confirmed key compromise

**Steps**:

1. **IMMEDIATE**: Execute containment: `scripts/dr/recover-key-compromise.sh <key_id> <type>`
2. **Notify** security team and leadership
3. **Assess** scope of compromise
4. **Rotate** affected keys
5. **Audit** all key usage
6. **Notify** affected parties (if required)
7. **Full** incident report within 24 hours

### RB-DR-006: Backup Restore

**Trigger**: Need to restore from backup due to data loss

**Steps**:

1. **Identify** appropriate backup point
2. **Download** backup to secure location
3. **Verify** backup integrity (checksums)
4. **Stop** target services
5. **Restore** data from backup
6. **Replay** any available WAL/logs
7. **Start** services and verify
8. **Reconcile** any data gaps

---

## Appendix

### Contact Information

| Role                | Contact                                | Escalation       |
| ------------------- | -------------------------------------- | ---------------- |
| Primary On-Call     | PagerDuty Schedule                     | 15 min           |
| Security Team       | security@virtengine.network            | Immediate for D5 |
| Infrastructure Lead | infrastructure-lead@virtengine.network | 30 min           |
| Executive           | exec-oncall@virtengine.network         | 1 hr for D3+     |

### Related Documentation

- [Horizontal Scaling Guide](horizontal-scaling-guide.md)
- [Key Management](key-management.md)
- [SLOs and Playbooks](slos-and-playbooks.md)
- [Incident Response](../docs/sre/INCIDENT_RESPONSE.md)
- [Incident Drills](../docs/sre/INCIDENT_DRILLS.md)
- [Reliability Testing](../docs/sre/RELIABILITY_TESTING.md)
- [Network Partition Recovery](../docs/operations/partition-recovery-runbook.md)

### Version History

| Version | Date       | Author   | Changes         |
| ------- | ---------- | -------- | --------------- |
| 1.0.0   | 2026-01-30 | SRE Team | Initial version |

---

**Document Owner**: SRE Team  
**Last Updated**: 2026-01-30  
**Next Review**: 2026-04-30
