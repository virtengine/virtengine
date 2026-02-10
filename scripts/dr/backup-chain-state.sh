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
#   NODE_HOME          Path to node home directory (default: /opt/virtengine)
#   SNAPSHOT_DIR       Local snapshot directory (default: /data/snapshots)
#   DR_BUCKET          S3 bucket for backups (required for upload)
#   RETENTION_COUNT    Number of local snapshots to keep (default: 10)
#   SNAPSHOT_INTERVAL  Blocks between snapshots (default: 1000)

set -euo pipefail

# Configuration
NODE_HOME="${NODE_HOME:-/opt/virtengine}"
SNAPSHOT_DIR="${SNAPSHOT_DIR:-/data/snapshots}"
DR_BUCKET="${DR_BUCKET:-}"
RETENTION_COUNT="${RETENTION_COUNT:-10}"
SNAPSHOT_INTERVAL="${SNAPSHOT_INTERVAL:-1000}"

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

# Get current block height
get_current_height() {
    virtengine status 2>&1 | jq -r '.sync_info.latest_block_height' || echo "0"
}

# Get current app hash
get_current_app_hash() {
    virtengine status 2>&1 | jq -r '.sync_info.latest_app_hash' || echo ""
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

# Create a chain state snapshot
create_snapshot() {
    local height=$(get_current_height)
    local app_hash=$(get_current_app_hash)
    local timestamp=$(date -u +%Y%m%d_%H%M%S)
    local snapshot_name="state_${height}_${timestamp}"
    
    log_info "Creating snapshot at height ${height}..."
    
    # Check if node is syncing
    local catching_up=$(virtengine status 2>&1 | jq -r '.sync_info.catching_up')
    if [ "$catching_up" == "true" ]; then
        log_warn "Node is still syncing, snapshot may be incomplete"
    fi
    
    # Export state JSON (lightweight, for genesis recreation if needed)
    log_info "Exporting state JSON..."
    if ! virtengine export --height "$height" > "${SNAPSHOT_DIR}/${snapshot_name}.json" 2>/dev/null; then
        log_warn "State export failed or not supported, skipping JSON export"
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
    "chain_id": "$(virtengine status 2>&1 | jq -r '.node_info.network' || echo 'unknown')"
}
EOF
    
    # Generate checksums
    cd "${SNAPSHOT_DIR}"
    sha256sum "${snapshot_name}.json" "${snapshot_name}_data.tar.gz" "${snapshot_name}_metadata.json" 2>/dev/null \
        > "${snapshot_name}.sha256" || \
        sha256sum "${snapshot_name}_data.tar.gz" "${snapshot_name}_metadata.json" > "${snapshot_name}.sha256"
    
    log_info "Snapshot created: ${snapshot_name}"
    log_info "  Height: ${height}"
    log_info "  App Hash: ${app_hash}"
    log_info "  Size: $(du -h "${SNAPSHOT_DIR}/${snapshot_name}_data.tar.gz" | cut -f1)"
    
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
    
    # Upload checksums
    aws s3 cp "${SNAPSHOT_DIR}/${snapshot_name}.sha256" \
        "${remote_path}/${snapshot_name}.sha256" \
        --only-show-errors
    
    # Verify upload
    if aws s3api head-object --bucket "$(echo $DR_BUCKET | sed 's|s3://||' | cut -d'/' -f1)" \
        --key "${region}/state/${snapshot_name}_data.tar.gz" > /dev/null 2>&1; then
        log_info "Upload verified successfully"
    else
        log_error "Upload verification failed!"
        return 1
    fi
}

# Clean up old local snapshots
cleanup_old_snapshots() {
    log_info "Cleaning up old snapshots (keeping ${RETENTION_COUNT})..."
    
    cd "${SNAPSHOT_DIR}"
    
    # Remove old data archives
    ls -t state_*_data.tar.gz 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f
    
    # Remove corresponding metadata and checksums
    ls -t state_*_metadata.json 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f
    ls -t state_*.sha256 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f
    ls -t state_*.json 2>/dev/null | grep -v metadata | tail -n +$((RETENTION_COUNT + 1)) | xargs -r rm -f
    
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
    if sha256sum -c "${snapshot_name}.sha256"; then
        log_info "Local checksum verification: PASSED"
    else
        log_error "Local checksum verification: FAILED"
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
        
        if diff -q "${snapshot_name}.sha256" /tmp/remote.sha256 > /dev/null 2>&1; then
            log_info "Remote checksum match: PASSED"
        else
            log_error "Remote checksum match: FAILED"
            rm -f /tmp/remote.sha256
            return 1
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
    
    # Find matching snapshot
    local snapshot_name=""
    
    # Check local first
    snapshot_name=$(ls -t "${SNAPSHOT_DIR}"/state_${target_height}_*_data.tar.gz 2>/dev/null | head -1 | xargs -r basename | sed 's/_data.tar.gz//')
    
    # If not local, try remote
    if [ -z "$snapshot_name" ] && [ -n "$DR_BUCKET" ]; then
        log_info "Snapshot not found locally, checking remote..."
        snapshot_name=$(aws s3 ls "${DR_BUCKET}/${region}/state/" | grep "state_${target_height}_" | grep "_data.tar.gz" | head -1 | awk '{print $4}' | sed 's/_data.tar.gz//')
        
        if [ -n "$snapshot_name" ]; then
            log_info "Downloading snapshot from remote..."
            aws s3 cp "${DR_BUCKET}/${region}/state/${snapshot_name}_data.tar.gz" "${SNAPSHOT_DIR}/" --only-show-errors
            aws s3 cp "${DR_BUCKET}/${region}/state/${snapshot_name}.sha256" "${SNAPSHOT_DIR}/" --only-show-errors
            aws s3 cp "${DR_BUCKET}/${region}/state/${snapshot_name}_metadata.json" "${SNAPSHOT_DIR}/" --only-show-errors
        fi
    fi
    
    if [ -z "$snapshot_name" ]; then
        log_error "No snapshot found for height ${target_height}"
        return 1
    fi
    
    log_info "Using snapshot: ${snapshot_name}"
    
    # Verify snapshot
    if ! verify_backup "$snapshot_name"; then
        log_error "Snapshot verification failed, aborting restore"
        return 1
    fi
    
    # Confirm restore
    log_warn "This will REPLACE the current chain data!"
    log_warn "Press Ctrl+C within 10 seconds to abort..."
    sleep 10
    
    # Stop the node
    log_info "Stopping virtengine service..."
    systemctl stop virtengine || true
    
    # Backup current data (just in case)
    log_info "Backing up current data..."
    if [ -d "${NODE_HOME}/data" ]; then
        mv "${NODE_HOME}/data" "${NODE_HOME}/data.backup.$(date +%s)"
    fi
    
    # Extract snapshot
    log_info "Extracting snapshot..."
    tar -xzf "${SNAPSHOT_DIR}/${snapshot_name}_data.tar.gz" -C "${NODE_HOME}/"
    
    # Start the node
    log_info "Starting virtengine service..."
    systemctl start virtengine
    
    # Monitor sync
    log_info "Monitoring sync progress..."
    local max_wait=300
    local waited=0
    
    while [ $waited -lt $max_wait ]; do
        if virtengine status 2>&1 | jq -e '.sync_info' > /dev/null 2>&1; then
            local current_height=$(get_current_height)
            local catching_up=$(virtengine status 2>&1 | jq -r '.sync_info.catching_up')
            
            log_info "Current height: ${current_height}, Catching up: ${catching_up}"
            
            if [ "$catching_up" == "false" ]; then
                log_info "Restore complete! Node is synced."
                return 0
            fi
        fi
        
        sleep 5
        waited=$((waited + 5))
    done
    
    log_warn "Node is still syncing after ${max_wait} seconds"
    log_info "Continue monitoring with: virtengine status"
    
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
