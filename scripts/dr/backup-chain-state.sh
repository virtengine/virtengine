#!/bin/bash
# scripts/dr/backup-chain-state.sh
# Automated chain state backup for disaster recovery
#
# Usage: ./backup-chain-state.sh [options]
#   --snapshot-only    Create snapshot without uploading
#   --upload-only      Upload existing snapshots
#   --verify           Verify backup integrity
#   --restore HEIGHT   Restore to specific height
#
# Environment variables:
#   NODE_HOME                Path to node home directory (default: /opt/virtengine)
#   SNAPSHOT_DIR             Local snapshot directory (default: /data/snapshots)
#   DR_BUCKET                S3 bucket for backups (required for upload)
#   RETENTION_COUNT          Number of local snapshots to keep (default: 10)
#   SNAPSHOT_INTERVAL        Blocks between snapshots (default: 1000)
#   SNAPSHOT_SIGNING_KEY     Path to PEM private key for snapshot signing
#   SNAPSHOT_VERIFY_PUBKEY   Path to PEM public key for snapshot verification
#   SNAPSHOT_SIGNING_REQUIRED Whether signing is required (default: 1)
#   SNAPSHOT_SIGNATURE_ALG   Digest alg for openssl (default: sha256)
#   VIRTENGINE_CMD           Override virtengine binary (default: virtengine)
#   SYSTEMCTL_CMD            Override systemctl command (default: systemctl)
#   ALERT_WEBHOOK            Optional webhook for backup/restore notifications
#   ALERT_WEBHOOK_TIMEOUT    Webhook timeout seconds (default: 5)
#   RESTORE_AUTO_APPROVE     Skip restore delay (default: 0)
#   RESTORE_SKIP_SERVICE     Skip systemctl stop/start (default: 0)
#   RESTORE_MAX_WAIT         Seconds to wait for sync check (default: 300)
#   RESTORE_FALLBACK_ENABLED Allow fallback to older snapshots (default: 1)

set -euo pipefail

# Configuration
NODE_HOME="${NODE_HOME:-/opt/virtengine}"
SNAPSHOT_DIR="${SNAPSHOT_DIR:-/data/snapshots}"
DR_BUCKET="${DR_BUCKET:-}"
RETENTION_COUNT="${RETENTION_COUNT:-10}"
SNAPSHOT_INTERVAL="${SNAPSHOT_INTERVAL:-1000}"
SNAPSHOT_SIGNING_KEY="${SNAPSHOT_SIGNING_KEY:-}"
SNAPSHOT_VERIFY_PUBKEY="${SNAPSHOT_VERIFY_PUBKEY:-}"
SNAPSHOT_SIGNING_REQUIRED="${SNAPSHOT_SIGNING_REQUIRED:-1}"
SNAPSHOT_SIGNATURE_ALG="${SNAPSHOT_SIGNATURE_ALG:-sha256}"
VIRTENGINE_CMD="${VIRTENGINE_CMD:-virtengine}"
SYSTEMCTL_CMD="${SYSTEMCTL_CMD:-systemctl}"
ALERT_WEBHOOK="${ALERT_WEBHOOK:-}"
ALERT_WEBHOOK_TIMEOUT="${ALERT_WEBHOOK_TIMEOUT:-5}"
RESTORE_AUTO_APPROVE="${RESTORE_AUTO_APPROVE:-0}"
RESTORE_SKIP_SERVICE="${RESTORE_SKIP_SERVICE:-0}"
RESTORE_MAX_WAIT="${RESTORE_MAX_WAIT:-300}"
RESTORE_FALLBACK_ENABLED="${RESTORE_FALLBACK_ENABLED:-1}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[$(date -u +%FT%TZ)] INFO:${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[$(date -u +%FT%TZ)] WARN:${NC} $*"
}

log_error() {
    echo -e "${RED}[$(date -u +%FT%TZ)] ERROR:${NC} $*" >&2
}

notify_webhook() {
    local event="$1"
    local status="$2"
    local message="$3"
    local snapshot_name="${4:-}"

    if [ -z "$ALERT_WEBHOOK" ]; then
        return 0
    fi

    if ! command -v curl > /dev/null 2>&1; then
        log_warn "curl not available, skipping webhook notification"
        return 0
    fi

    local payload
    payload=$(cat << EOF
{
  "event": "${event}",
  "status": "${status}",
  "message": "${message}",
  "snapshot": "${snapshot_name}",
  "timestamp": "$(date -u +%FT%TZ)",
  "host": "$(hostname)",
  "region": "$(get_region)"
}
EOF
)

    curl -s -X POST "$ALERT_WEBHOOK" \
        -H "Content-Type: application/json" \
        -d "$payload" \
        --max-time "$ALERT_WEBHOOK_TIMEOUT" > /dev/null 2>&1 || true
}

# Get current block height
get_current_height() {
    "$VIRTENGINE_CMD" status 2>&1 | jq -r '.sync_info.latest_block_height' || echo "0"
}

# Get current app hash
get_current_app_hash() {
    "$VIRTENGINE_CMD" status 2>&1 | jq -r '.sync_info.latest_app_hash' || echo ""
}

# Determine region from hostname or metadata
get_region() {
    # Try AWS metadata
    local region=$(curl -s --connect-timeout 2 http://169.254.169.254/latest/meta-data/placement/region 2>/dev/null || true)
    
    if [ -z "$region" ]; then
        # Fall back to hostname convention: us-east-validator-0
        region=$(hostname | cut -d'-' -f1-2 2>/dev/null || echo "unknown")
    fi
    
    echo "$region"
}

# Create snapshot directory if it doesn't exist
ensure_directories() {
    mkdir -p "${SNAPSHOT_DIR}"
    mkdir -p "${NODE_HOME}/data/snapshots"
}

ensure_signing_tools() {
    if ! command -v openssl > /dev/null 2>&1; then
        log_error "openssl not found; snapshot signing/verification requires openssl"
        return 1
    fi
    return 0
}

resolve_verify_key() {
    if [ -n "$SNAPSHOT_VERIFY_PUBKEY" ]; then
        echo "$SNAPSHOT_VERIFY_PUBKEY"
        return 0
    fi
    if [ -n "$SNAPSHOT_SIGNING_KEY" ]; then
        echo "$SNAPSHOT_SIGNING_KEY"
        return 0
    fi
    return 1
}

get_signing_fingerprint() {
    local key_path="$1"
    if [ -z "$key_path" ] || [ ! -f "$key_path" ]; then
        echo "unknown"
        return 0
    fi
    openssl pkey -in "$key_path" -pubout 2>/dev/null | openssl dgst -sha256 | awk '{print $2}'
}

add_signature_metadata() {
    local snapshot_name="$1"
    if [ -z "$SNAPSHOT_SIGNING_KEY" ] || [ ! -f "$SNAPSHOT_SIGNING_KEY" ]; then
        return 0
    fi

    local key_fingerprint
    key_fingerprint=$(get_signing_fingerprint "$SNAPSHOT_SIGNING_KEY")

    if [ -f "${SNAPSHOT_DIR}/${snapshot_name}_metadata.json" ]; then
        jq ".signature = {\"file\": \"${snapshot_name}.sig\", \"algorithm\": \"${SNAPSHOT_SIGNATURE_ALG}\", \"key_fingerprint\": \"${key_fingerprint}\"}" \
            "${SNAPSHOT_DIR}/${snapshot_name}_metadata.json" > "${SNAPSHOT_DIR}/${snapshot_name}_metadata.json.tmp"
        mv "${SNAPSHOT_DIR}/${snapshot_name}_metadata.json.tmp" "${SNAPSHOT_DIR}/${snapshot_name}_metadata.json"
    fi
}

sign_snapshot() {
    local snapshot_name="$1"
    local checksum_file="${SNAPSHOT_DIR}/${snapshot_name}.sha256"
    local signature_file="${SNAPSHOT_DIR}/${snapshot_name}.sig"

    if [ "$SNAPSHOT_SIGNING_REQUIRED" != "0" ] && [ -z "$SNAPSHOT_SIGNING_KEY" ]; then
        log_error "SNAPSHOT_SIGNING_KEY is required but not set"
        return 1
    fi

    if [ -z "$SNAPSHOT_SIGNING_KEY" ]; then
        log_warn "SNAPSHOT_SIGNING_KEY not set; skipping signing"
        return 0
    fi

    if [ ! -f "$SNAPSHOT_SIGNING_KEY" ]; then
        log_error "Signing key not found: ${SNAPSHOT_SIGNING_KEY}"
        return 1
    fi

    ensure_signing_tools

    log_info "Signing snapshot checksum..."
    openssl dgst -"${SNAPSHOT_SIGNATURE_ALG}" -sign "$SNAPSHOT_SIGNING_KEY" \
        -out "$signature_file" "$checksum_file"

    log_info "Snapshot signature created: $(basename "$signature_file")"
    return 0
}

verify_snapshot_signature() {
    local snapshot_name="$1"
    local checksum_file="${SNAPSHOT_DIR}/${snapshot_name}.sha256"
    local signature_file="${SNAPSHOT_DIR}/${snapshot_name}.sig"
    local verify_key

    if [ "$SNAPSHOT_SIGNING_REQUIRED" = "0" ] && [ ! -f "$signature_file" ]; then
        log_warn "Signature verification skipped (signing not required)"
        return 0
    fi

    if [ ! -f "$signature_file" ]; then
        log_error "Signature file missing: ${signature_file}"
        return 1
    fi

    verify_key=$(resolve_verify_key || true)
    if [ -z "$verify_key" ]; then
        log_error "No verification key available (SNAPSHOT_VERIFY_PUBKEY or SNAPSHOT_SIGNING_KEY required)"
        return 1
    fi

    if [ ! -f "$verify_key" ]; then
        log_error "Verification key not found: ${verify_key}"
        return 1
    fi

    ensure_signing_tools

    if openssl dgst -"${SNAPSHOT_SIGNATURE_ALG}" -verify "$verify_key" \
        -signature "$signature_file" "$checksum_file" > /dev/null 2>&1; then
        log_info "Signature verification: PASSED"
        return 0
    fi

    log_error "Signature verification: FAILED"
    return 1
}

# Create a chain state snapshot
create_snapshot() {
    local height=$(get_current_height)
    local app_hash=$(get_current_app_hash)
    local timestamp=$(date -u +%Y%m%d_%H%M%S)
    local snapshot_name="state_${height}_${timestamp}"
    local json_exported=1
    
    log_info "Creating snapshot at height ${height}..."
    
    # Check if node is syncing
    local catching_up=$("$VIRTENGINE_CMD" status 2>/dev/null | jq -r '.sync_info.catching_up' 2>/dev/null || echo "false")
    if [ "$catching_up" == "true" ]; then
        log_warn "Node is still syncing, snapshot may be incomplete"
    fi
    
    # Export state JSON (lightweight, for genesis recreation if needed)
    log_info "Exporting state JSON..."
    if ! "$VIRTENGINE_CMD" export --height "$height" > "${SNAPSHOT_DIR}/${snapshot_name}.json" 2>/dev/null; then
        log_warn "State export failed or not supported, skipping JSON export"
        json_exported=0
        rm -f "${SNAPSHOT_DIR}/${snapshot_name}.json"
        touch "${SNAPSHOT_DIR}/${snapshot_name}.json.skipped"
    fi
    
    # Create compressed archive of data directory
    log_info "Creating data archive..."
    tar -czf "${SNAPSHOT_DIR}/${snapshot_name}_data.tar.gz" \
        -C "${NODE_HOME}" \
        --exclude="*.log" \
        --exclude="*.tmp" \
        --exclude="*.lock" \
        --exclude="cs.wal" \
        --exclude="wasm/wasm/cache" \
        data/
    
    # Generate metadata
    cat > "${SNAPSHOT_DIR}/${snapshot_name}_metadata.json" << EOF
{
    "snapshot_name": "${snapshot_name}",
    "height": ${height},
    "app_hash": "${app_hash}",
    "timestamp": "$(date -u +%FT%TZ)",
    "node_home": "${NODE_HOME}",
    "region": "$(get_region)",
    "hostname": "$(hostname)",
    "chain_id": "$("$VIRTENGINE_CMD" status 2>/dev/null | jq -r '.node_info.network' 2>/dev/null || echo 'unknown')"
}
EOF
    
    add_signature_metadata "$snapshot_name"

    # Generate checksums
    cd "${SNAPSHOT_DIR}"
    local checksum_files=("${snapshot_name}_data.tar.gz" "${snapshot_name}_metadata.json")
    if [ "$json_exported" -eq 1 ] && [ -f "${snapshot_name}.json" ] && [ ! -f "${snapshot_name}.json.skipped" ]; then
        checksum_files=("${snapshot_name}.json" "${checksum_files[@]}")
    fi
    sha256sum "${checksum_files[@]}" > "${snapshot_name}.sha256"
    
    sign_snapshot "$snapshot_name"

    log_info "Snapshot created: ${snapshot_name}"
    log_info "  Height: ${height}"
    log_info "  App Hash: ${app_hash}"
    log_info "  Size: $(du -h "${SNAPSHOT_DIR}/${snapshot_name}_data.tar.gz" | cut -f1)"

    notify_webhook "snapshot.create" "success" "Snapshot created successfully" "$snapshot_name"
    echo "${snapshot_name}"
}

# Upload snapshot to remote storage
upload_to_remote() {
    local snapshot_name="$1"
    
    if [ -z "$DR_BUCKET" ]; then
        log_warn "DR_BUCKET not set, skipping remote upload"
        return 0
    fi
    
    local region=$(get_region)
    local remote_path="${DR_BUCKET}/${region}/state"
    
    log_info "Uploading snapshot to ${remote_path}..."
    
    # Upload data archive (larger file, use intelligent tiering)
    aws s3 cp "${SNAPSHOT_DIR}/${snapshot_name}_data.tar.gz" \
        "${remote_path}/${snapshot_name}_data.tar.gz" \
        --storage-class STANDARD_IA \
        --only-show-errors
    
    # Upload metadata (small, keep in standard storage)
    aws s3 cp "${SNAPSHOT_DIR}/${snapshot_name}_metadata.json" \
        "${remote_path}/${snapshot_name}_metadata.json" \
        --only-show-errors
    
    # Upload state export if it exists
    if [ -f "${SNAPSHOT_DIR}/${snapshot_name}.json" ] && [ ! -f "${SNAPSHOT_DIR}/${snapshot_name}.json.skipped" ]; then
        aws s3 cp "${SNAPSHOT_DIR}/${snapshot_name}.json" \
            "${remote_path}/${snapshot_name}.json" \
            --storage-class STANDARD_IA \
            --only-show-errors
    fi

    if [ -f "${SNAPSHOT_DIR}/${snapshot_name}.json.skipped" ]; then
        aws s3 cp "${SNAPSHOT_DIR}/${snapshot_name}.json.skipped" \
            "${remote_path}/${snapshot_name}.json.skipped" \
            --only-show-errors
    fi
    
    # Upload checksums
    aws s3 cp "${SNAPSHOT_DIR}/${snapshot_name}.sha256" \
        "${remote_path}/${snapshot_name}.sha256" \
        --only-show-errors

    # Upload signature if present
    if [ -f "${SNAPSHOT_DIR}/${snapshot_name}.sig" ]; then
        aws s3 cp "${SNAPSHOT_DIR}/${snapshot_name}.sig" \
            "${remote_path}/${snapshot_name}.sig" \
            --only-show-errors
    fi
    
    # Verify upload
    if aws s3api head-object --bucket "$(echo "$DR_BUCKET" | sed 's|s3://||' | cut -d'/' -f1)" \
        --key "${region}/state/${snapshot_name}_data.tar.gz" > /dev/null 2>&1; then
        log_info "Upload verified successfully"
        notify_webhook "snapshot.upload" "success" "Snapshot uploaded successfully" "$snapshot_name"
    else
        log_error "Upload verification failed!"
        notify_webhook "snapshot.upload" "failure" "Snapshot upload verification failed" "$snapshot_name"
        return 1
    fi
}

# Clean up old local snapshots
cleanup_old_snapshots() {
    log_info "Cleaning up old snapshots (keeping ${RETENTION_COUNT})..."
    
    cd "${SNAPSHOT_DIR}"
    
    # Remove old data archives
    ls -t state_*_data.tar.gz 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f || true
    
    # Remove corresponding metadata and checksums
    ls -t state_*_metadata.json 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f || true
    ls -t state_*.sha256 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f || true
    ls -t state_*.sig 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f || true
    ls -t state_*.json 2>/dev/null | grep -v metadata | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f || true
    ls -t state_*.json.skipped 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f || true
    
    log_info "Cleanup complete"
}

# Verify backup integrity
verify_backup() {
    local snapshot_name="${1:-}"
    
    if [ -z "$snapshot_name" ]; then
        # Find latest snapshot
        snapshot_name=$(ls -t "${SNAPSHOT_DIR}"/state_*_data.tar.gz 2>/dev/null | head -1 | xargs -r basename | sed 's/_data.tar.gz//')
    fi
    
    if [ -z "$snapshot_name" ]; then
        log_error "No snapshot found to verify"
        return 1
    fi
    
    log_info "Verifying snapshot: ${snapshot_name}"
    
    cd "${SNAPSHOT_DIR}"
    
    # Verify checksums
    local checksum_file="${snapshot_name}.sha256"
    local verify_checksum_file="${checksum_file}"
    if [ ! -f "${snapshot_name}.json" ] && grep -q " ${snapshot_name}.json$" "${checksum_file}"; then
        if [ -f "${snapshot_name}.json.skipped" ]; then
            log_warn "State export marked as skipped; verifying remaining files"
        else
            log_warn "State export missing; verifying remaining files"
        fi
        verify_checksum_file="$(mktemp)"
        grep -v " ${snapshot_name}.json$" "${checksum_file}" > "${verify_checksum_file}"
    fi

    if sha256sum -c "${verify_checksum_file}"; then
        log_info "Local checksum verification: PASSED"
    else
        log_error "Local checksum verification: FAILED"
        [ "${verify_checksum_file}" != "${checksum_file}" ] && rm -f "${verify_checksum_file}"
        return 1
    fi
    [ "${verify_checksum_file}" != "${checksum_file}" ] && rm -f "${verify_checksum_file}"

    # Verify signature
    if ! verify_snapshot_signature "$snapshot_name"; then
        log_error "Signature verification failed"
        return 1
    fi
    
    # Test archive integrity
    if tar -tzf "${snapshot_name}_data.tar.gz" > /dev/null 2>&1; then
        log_info "Archive integrity test: PASSED"
    else
        log_error "Archive integrity test: FAILED"
        return 1
    fi
    
    # If remote bucket is configured, verify remote copy
    if [ -n "$DR_BUCKET" ]; then
        log_info "Verifying remote backup..."
        local region=$(get_region)
        
        # Download remote checksum
        aws s3 cp "${DR_BUCKET}/${region}/state/${snapshot_name}.sha256" /tmp/remote.sha256 --only-show-errors
        if aws s3 ls "${DR_BUCKET}/${region}/state/${snapshot_name}.sig" > /dev/null 2>&1; then
            aws s3 cp "${DR_BUCKET}/${region}/state/${snapshot_name}.sig" /tmp/remote.sig --only-show-errors
        fi
        
        if diff -q "${snapshot_name}.sha256" /tmp/remote.sha256 > /dev/null 2>&1; then
            log_info "Remote checksum match: PASSED"
        else
            log_error "Remote checksum match: FAILED"
            rm -f /tmp/remote.sha256
            rm -f /tmp/remote.sig
            return 1
        fi

        if [ -f /tmp/remote.sig ]; then
            cp /tmp/remote.sig "${snapshot_name}.sig"
            if ! verify_snapshot_signature "$snapshot_name"; then
                log_error "Remote signature verification failed"
                rm -f /tmp/remote.sha256 /tmp/remote.sig
                return 1
            fi
            rm -f /tmp/remote.sig
        fi

        rm -f /tmp/remote.sha256
    fi
    
    log_info "Backup verification complete: ALL PASSED"
    return 0
}

# Restore from snapshot
restore_snapshot() {
    local target_height="$1"
    local region=$(get_region)

    if [ -z "$target_height" ]; then
        log_error "Target height required for restore"
        return 1
    fi

    log_info "Restoring to height: ${target_height}"

    local snapshot_name=""
    local candidates=()

    while IFS= read -r line; do
        [ -n "$line" ] && candidates+=("$line")
    done < <(ls -t "${SNAPSHOT_DIR}"/state_${target_height}_*_data.tar.gz 2>/dev/null | sed 's/_data.tar.gz//' | xargs -r -n1 basename)

    if [ ${#candidates[@]} -eq 0 ] && [ -n "$DR_BUCKET" ]; then
        log_info "Snapshot not found locally, checking remote..."
        while IFS= read -r line; do
            [ -n "$line" ] && candidates+=("$line")
        done < <(aws s3 ls "${DR_BUCKET}/${region}/state/" | awk '{print $4}' | grep "state_${target_height}_" | grep "_data.tar.gz" | sed 's/_data.tar.gz//' || true)
    fi

    if [ ${#candidates[@]} -eq 0 ]; then
        log_error "No snapshot found for height ${target_height}"
        return 1
    fi

    for candidate in "${candidates[@]}"; do
        snapshot_name="$candidate"
        if [ ! -f "${SNAPSHOT_DIR}/${snapshot_name}_data.tar.gz" ] && [ -n "$DR_BUCKET" ]; then
            log_info "Downloading snapshot ${snapshot_name} from remote..."
            aws s3 cp "${DR_BUCKET}/${region}/state/${snapshot_name}_data.tar.gz" "${SNAPSHOT_DIR}/" --only-show-errors
            aws s3 cp "${DR_BUCKET}/${region}/state/${snapshot_name}.sha256" "${SNAPSHOT_DIR}/" --only-show-errors
            aws s3 cp "${DR_BUCKET}/${region}/state/${snapshot_name}_metadata.json" "${SNAPSHOT_DIR}/" --only-show-errors
            aws s3 cp "${DR_BUCKET}/${region}/state/${snapshot_name}.sig" "${SNAPSHOT_DIR}/" --only-show-errors 2>/dev/null || true
        fi

        log_info "Validating snapshot candidate: ${snapshot_name}"
        if verify_backup "$snapshot_name"; then
            log_info "Snapshot validation passed: ${snapshot_name}"
            break
        fi

        log_warn "Snapshot validation failed: ${snapshot_name}"
        snapshot_name=""
    done

    if [ -z "$snapshot_name" ] && [ "$RESTORE_FALLBACK_ENABLED" != "0" ]; then
        log_warn "No valid snapshot found at requested height; attempting fallback to latest valid snapshot"
        for candidate in $(ls -t "${SNAPSHOT_DIR}"/state_*_data.tar.gz 2>/dev/null | sed 's/_data.tar.gz//' | xargs -r -n1 basename); do
            if verify_backup "$candidate"; then
                snapshot_name="$candidate"
                log_warn "Falling back to snapshot: ${snapshot_name}"
                break
            fi
        done
    fi

    if [ -z "$snapshot_name" ]; then
        log_error "No valid snapshot found for restore"
        notify_webhook "snapshot.restore" "failure" "No valid snapshot available for restore" ""
        return 1
    fi

    log_info "Using snapshot: ${snapshot_name}"

    log_warn "This will REPLACE the current chain data!"
    if [ "$RESTORE_AUTO_APPROVE" = "0" ]; then
        log_warn "Press Ctrl+C within 10 seconds to abort..."
        sleep 10
    fi

    if [ "$RESTORE_SKIP_SERVICE" = "0" ]; then
        log_info "Stopping virtengine service..."
        "$SYSTEMCTL_CMD" stop virtengine || true
    else
        log_warn "RESTORE_SKIP_SERVICE=1 set; skipping service stop"
    fi

    log_info "Backing up current data..."
    local backup_dir=""
    if [ -d "${NODE_HOME}/data" ]; then
        backup_dir="${NODE_HOME}/data.backup.$(date +%s)"
        mv "${NODE_HOME}/data" "$backup_dir"
    fi

    local restore_dir="${NODE_HOME}/data.restore.$(date +%s)"
    mkdir -p "$restore_dir"

    log_info "Extracting snapshot..."
    if ! tar -xzf "${SNAPSHOT_DIR}/${snapshot_name}_data.tar.gz" -C "$restore_dir"; then
        log_error "Snapshot extraction failed"
        if [ -n "$backup_dir" ] && [ -d "$backup_dir" ]; then
            rm -rf "${NODE_HOME}/data"
            mv "$backup_dir" "${NODE_HOME}/data"
        fi
        notify_webhook "snapshot.restore" "failure" "Snapshot extraction failed" "$snapshot_name"
        return 1
    fi

    if [ ! -d "${restore_dir}/data" ]; then
        log_error "Snapshot data directory missing after extraction"
        if [ -n "$backup_dir" ] && [ -d "$backup_dir" ]; then
            rm -rf "${NODE_HOME}/data"
            mv "$backup_dir" "${NODE_HOME}/data"
        fi
        notify_webhook "snapshot.restore" "failure" "Snapshot data directory missing after extraction" "$snapshot_name"
        return 1
    fi

    mv "${restore_dir}/data" "${NODE_HOME}/data"
    rm -rf "$restore_dir"

    if [ "$RESTORE_SKIP_SERVICE" = "0" ]; then
        log_info "Starting virtengine service..."
        "$SYSTEMCTL_CMD" start virtengine
    else
        log_warn "RESTORE_SKIP_SERVICE=1 set; skipping service start"
    fi

    log_info "Monitoring sync progress..."
    local max_wait="${RESTORE_MAX_WAIT}"
    local waited=0

    while [ $waited -lt $max_wait ]; do
        if "$VIRTENGINE_CMD" status 2>&1 | jq -e '.sync_info' > /dev/null 2>&1; then
            local current_height=$(get_current_height)
            local catching_up=$("$VIRTENGINE_CMD" status 2>&1 | jq -r '.sync_info.catching_up')

            log_info "Current height: ${current_height}, Catching up: ${catching_up}"

            if [ "$catching_up" == "false" ]; then
                log_info "Restore complete! Node is synced."
                notify_webhook "snapshot.restore" "success" "Restore completed and node is synced" "$snapshot_name"
                return 0
            fi
        fi

        sleep 5
        waited=$((waited + 5))
    done

    log_warn "Node is still syncing after ${max_wait} seconds"
    log_info "Continue monitoring with: virtengine status"

    notify_webhook "snapshot.restore" "warning" "Restore completed but node still syncing" "$snapshot_name"
    return 0
}

# List available snapshots
list_snapshots() {
    log_info "Local snapshots in ${SNAPSHOT_DIR}:"
    echo ""
    
    if ls "${SNAPSHOT_DIR}"/state_*_metadata.json 1> /dev/null 2>&1; then
        for meta in "${SNAPSHOT_DIR}"/state_*_metadata.json; do
            local name=$(jq -r '.snapshot_name' "$meta")
            local height=$(jq -r '.height' "$meta")
            local timestamp=$(jq -r '.timestamp' "$meta")
            local data_file="${SNAPSHOT_DIR}/${name}_data.tar.gz"
            local size="N/A"
            
            if [ -f "$data_file" ]; then
                size=$(du -h "$data_file" | cut -f1)
            fi
            
            printf "  %-40s Height: %-10s Size: %-8s %s\n" "$name" "$height" "$size" "$timestamp"
        done
    else
        echo "  No local snapshots found"
    fi
    
    if [ -n "$DR_BUCKET" ]; then
        echo ""
        log_info "Remote snapshots in ${DR_BUCKET}/$(get_region)/state/:"
        echo ""
        
        aws s3 ls "${DR_BUCKET}/$(get_region)/state/" 2>/dev/null | grep "_data.tar.gz" | while read -r line; do
            local size=$(echo "$line" | awk '{print $3}')
            local name=$(echo "$line" | awk '{print $4}' | sed 's/_data.tar.gz//')
            local date=$(echo "$line" | awk '{print $1, $2}')
            
            # Convert bytes to human readable
            if [ "$size" -gt 1073741824 ]; then
                size="$(echo "scale=1; $size/1073741824" | bc)G"
            elif [ "$size" -gt 1048576 ]; then
                size="$(echo "scale=1; $size/1048576" | bc)M"
            else
                size="${size}B"
            fi
            
            printf "  %-40s Size: %-8s %s\n" "$name" "$size" "$date"
        done || echo "  No remote snapshots found or bucket not accessible"
    fi
}

# Main execution
main() {
    local action="full"  # default: create and upload
    local target_height=""
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --snapshot-only)
                action="snapshot"
                shift
                ;;
            --upload-only)
                action="upload"
                shift
                ;;
            --verify)
                action="verify"
                shift
                ;;
            --restore)
                action="restore"
                target_height="$2"
                shift 2
                ;;
            --list)
                action="list"
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [options]"
                echo ""
                echo "Options:"
                echo "  --snapshot-only    Create snapshot without uploading"
                echo "  --upload-only      Upload existing snapshots"
                echo "  --verify           Verify backup integrity"
                echo "  --restore HEIGHT   Restore to specific height"
                echo "  --list             List available snapshots"
                echo "  --help             Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    ensure_directories
    
    case $action in
        full)
            local snapshot_name=$(create_snapshot)
            upload_to_remote "$snapshot_name"
            cleanup_old_snapshots
            verify_backup "$snapshot_name"
            ;;
        snapshot)
            local snapshot_name=$(create_snapshot)
            cleanup_old_snapshots
            ;;
        upload)
            # Find latest snapshot and upload
            local latest=$(ls -t "${SNAPSHOT_DIR}"/state_*_data.tar.gz 2>/dev/null | head -1 | xargs -r basename | sed 's/_data.tar.gz//')
            if [ -n "$latest" ]; then
                upload_to_remote "$latest"
            else
                log_error "No snapshot found to upload"
                exit 1
            fi
            ;;
        verify)
            verify_backup ""
            ;;
        restore)
            restore_snapshot "$target_height"
            ;;
        list)
            list_snapshots
            ;;
    esac
}

main "$@"
