#!/bin/bash
# scripts/dr/backup-provider-state.sh
# Backup and restore provider daemon state with integrity verification.
#
# Usage: ./backup-provider-state.sh [options]
#   --backup           Create a provider state backup (default)
#   --verify           Verify latest backup integrity
#   --restore NAME     Restore provider state from backup name
#   --list             List available backups
#   --cleanup          Remove old backups (keep last RETENTION_COUNT)
#
# Environment variables:
#   PROVIDER_HOME            Provider daemon home (default: /opt/provider-daemon)
#   PROVIDER_SNAPSHOT_DIR    Local backup directory (default: /data/provider-snapshots)
#   DR_BUCKET                S3 bucket for backups (optional)
#   RETENTION_COUNT          Number of local backups to keep (default: 10)
#   SNAPSHOT_SIGNING_KEY     Path to PEM private key for signing
#   SNAPSHOT_VERIFY_PUBKEY   Path to PEM public key for verification
#   SNAPSHOT_SIGNING_REQUIRED Whether signing is required (default: 1)
#   SNAPSHOT_SIGNATURE_ALG   Digest alg for openssl (default: sha256)
#   ALERT_WEBHOOK            Optional webhook for notifications
#   ALERT_WEBHOOK_TIMEOUT    Webhook timeout seconds (default: 5)
#   RESTORE_AUTO_APPROVE     Skip restore delay (default: 0)
#   RESTORE_SKIP_SERVICE     Skip systemctl stop/start (default: 0)
#   SYSTEMCTL_CMD            Override systemctl command (default: systemctl)

set -euo pipefail

PROVIDER_HOME="${PROVIDER_HOME:-/opt/provider-daemon}"
PROVIDER_SNAPSHOT_DIR="${PROVIDER_SNAPSHOT_DIR:-/data/provider-snapshots}"
DR_BUCKET="${DR_BUCKET:-}"
RETENTION_COUNT="${RETENTION_COUNT:-10}"
SNAPSHOT_SIGNING_KEY="${SNAPSHOT_SIGNING_KEY:-}"
SNAPSHOT_VERIFY_PUBKEY="${SNAPSHOT_VERIFY_PUBKEY:-}"
SNAPSHOT_SIGNING_REQUIRED="${SNAPSHOT_SIGNING_REQUIRED:-1}"
SNAPSHOT_SIGNATURE_ALG="${SNAPSHOT_SIGNATURE_ALG:-sha256}"
ALERT_WEBHOOK="${ALERT_WEBHOOK:-}"
ALERT_WEBHOOK_TIMEOUT="${ALERT_WEBHOOK_TIMEOUT:-5}"
RESTORE_AUTO_APPROVE="${RESTORE_AUTO_APPROVE:-0}"
RESTORE_SKIP_SERVICE="${RESTORE_SKIP_SERVICE:-0}"
SYSTEMCTL_CMD="${SYSTEMCTL_CMD:-systemctl}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[$(date -u +%FT%TZ)] INFO:${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[$(date -u +%FT%TZ)] WARN:${NC} $*"
}

log_error() {
    echo -e "${RED}[$(date -u +%FT%TZ)] ERROR:${NC} $*" >&2
}

get_region() {
    local region
    region=$(curl -s --connect-timeout 2 http://169.254.169.254/latest/meta-data/placement/region 2>/dev/null || true)
    if [ -z "$region" ]; then
        region=$(hostname | cut -d'-' -f1-2 2>/dev/null || echo "unknown")
    fi
    echo "$region"
}

notify_webhook() {
    local event="$1"
    local status="$2"
    local message="$3"
    local backup_name="${4:-}"

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
  "backup": "${backup_name}",
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

ensure_directories() {
    mkdir -p "$PROVIDER_SNAPSHOT_DIR"
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
    local backup_name="$1"
    if [ -z "$SNAPSHOT_SIGNING_KEY" ] || [ ! -f "$SNAPSHOT_SIGNING_KEY" ]; then
        return 0
    fi

    local key_fingerprint
    key_fingerprint=$(get_signing_fingerprint "$SNAPSHOT_SIGNING_KEY")

    if [ -f "${PROVIDER_SNAPSHOT_DIR}/${backup_name}_metadata.json" ]; then
        jq ".signature = {\"file\": \"${backup_name}.sig\", \"algorithm\": \"${SNAPSHOT_SIGNATURE_ALG}\", \"key_fingerprint\": \"${key_fingerprint}\"}" \
            "${PROVIDER_SNAPSHOT_DIR}/${backup_name}_metadata.json" > "${PROVIDER_SNAPSHOT_DIR}/${backup_name}_metadata.json.tmp"
        mv "${PROVIDER_SNAPSHOT_DIR}/${backup_name}_metadata.json.tmp" "${PROVIDER_SNAPSHOT_DIR}/${backup_name}_metadata.json"
    fi
}

sign_backup() {
    local backup_name="$1"
    local checksum_file="${PROVIDER_SNAPSHOT_DIR}/${backup_name}.sha256"
    local signature_file="${PROVIDER_SNAPSHOT_DIR}/${backup_name}.sig"

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

    log_info "Signing provider backup checksum..."
    openssl dgst -"${SNAPSHOT_SIGNATURE_ALG}" -sign "$SNAPSHOT_SIGNING_KEY" \
        -out "$signature_file" "$checksum_file"

    log_info "Provider backup signature created: $(basename "$signature_file")"
}

verify_signature() {
    local backup_name="$1"
    local checksum_file="${PROVIDER_SNAPSHOT_DIR}/${backup_name}.sha256"
    local signature_file="${PROVIDER_SNAPSHOT_DIR}/${backup_name}.sig"
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

create_backup() {
    local timestamp
    timestamp=$(date -u +%Y%m%d_%H%M%S)
    local backup_name="provider_state_${timestamp}"

    local state_dir="${PROVIDER_HOME}/data"
    local config_dir="${PROVIDER_HOME}/config"

    log_info "Creating provider state backup..."

    if [ ! -d "$state_dir" ] && [ ! -d "$config_dir" ]; then
        log_error "Provider data/config directories not found in ${PROVIDER_HOME}"
        return 1
    fi

    local tmp_dir
    tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT

    if [ -d "$state_dir" ]; then
        mkdir -p "$tmp_dir/data"
        cp -R "$state_dir"/* "$tmp_dir/data/" 2>/dev/null || true
    fi

    if [ -d "$config_dir" ]; then
        mkdir -p "$tmp_dir/config"
        cp -R "$config_dir"/* "$tmp_dir/config/" 2>/dev/null || true
    fi

    tar -czf "${PROVIDER_SNAPSHOT_DIR}/${backup_name}.tar.gz" -C "$tmp_dir" .

    cat > "${PROVIDER_SNAPSHOT_DIR}/${backup_name}_metadata.json" << EOF
{
  "backup_name": "${backup_name}",
  "timestamp": "$(date -u +%FT%TZ)",
  "provider_home": "${PROVIDER_HOME}",
  "state_dir": "${state_dir}",
  "config_dir": "${config_dir}",
  "hostname": "$(hostname)",
  "region": "$(get_region)"
}
EOF

    add_signature_metadata "$backup_name"

    (cd "$PROVIDER_SNAPSHOT_DIR" && sha256sum "${backup_name}.tar.gz" "${backup_name}_metadata.json" > "${backup_name}.sha256")

    sign_backup "$backup_name"

    log_info "Provider state backup created: ${backup_name}"
    notify_webhook "provider.backup" "success" "Provider state backup created" "$backup_name"
    echo "$backup_name"
}

upload_backup() {
    local backup_name="$1"

    if [ -z "$DR_BUCKET" ]; then
        log_warn "DR_BUCKET not set, skipping remote upload"
        return 0
    fi

    local region
    region=$(get_region)
    local remote_path="${DR_BUCKET}/${region}/provider"

    log_info "Uploading provider backup to ${remote_path}..."

    aws s3 cp "${PROVIDER_SNAPSHOT_DIR}/${backup_name}.tar.gz" \
        "${remote_path}/${backup_name}.tar.gz" \
        --storage-class STANDARD_IA \
        --only-show-errors

    aws s3 cp "${PROVIDER_SNAPSHOT_DIR}/${backup_name}_metadata.json" \
        "${remote_path}/${backup_name}_metadata.json" \
        --only-show-errors

    aws s3 cp "${PROVIDER_SNAPSHOT_DIR}/${backup_name}.sha256" \
        "${remote_path}/${backup_name}.sha256" \
        --only-show-errors

    if [ -f "${PROVIDER_SNAPSHOT_DIR}/${backup_name}.sig" ]; then
        aws s3 cp "${PROVIDER_SNAPSHOT_DIR}/${backup_name}.sig" \
            "${remote_path}/${backup_name}.sig" \
            --only-show-errors
    fi

    notify_webhook "provider.backup.upload" "success" "Provider state backup uploaded" "$backup_name"
}

verify_backup() {
    local backup_name="$1"

    if [ -z "$backup_name" ]; then
        backup_name=$(ls -t "${PROVIDER_SNAPSHOT_DIR}"/provider_state_*.tar.gz 2>/dev/null | head -1 | xargs -r basename | sed 's/.tar.gz//')
    fi

    if [ -z "$backup_name" ]; then
        log_error "No provider backup found to verify"
        return 1
    fi

    log_info "Verifying provider backup: ${backup_name}"

    if (cd "$PROVIDER_SNAPSHOT_DIR" && sha256sum -c "${backup_name}.sha256" > /dev/null 2>&1); then
        log_info "Checksum verification: PASSED"
    else
        log_error "Checksum verification: FAILED"
        return 1
    fi

    if ! verify_signature "$backup_name"; then
        return 1
    fi

    if tar -tzf "${PROVIDER_SNAPSHOT_DIR}/${backup_name}.tar.gz" > /dev/null 2>&1; then
        log_info "Archive integrity test: PASSED"
    else
        log_error "Archive integrity test: FAILED"
        return 1
    fi

    log_info "Provider backup verification complete: ALL PASSED"
    return 0
}

restore_backup() {
    local backup_name="$1"

    if [ -z "$backup_name" ]; then
        log_error "Backup name required for restore"
        return 1
    fi

    log_info "Restoring provider state from ${backup_name}"

    if [ ! -f "${PROVIDER_SNAPSHOT_DIR}/${backup_name}.tar.gz" ] && [ -n "$DR_BUCKET" ]; then
        local region
        region=$(get_region)
        log_info "Downloading provider backup from remote..."
        aws s3 cp "${DR_BUCKET}/${region}/provider/${backup_name}.tar.gz" "${PROVIDER_SNAPSHOT_DIR}/" --only-show-errors
        aws s3 cp "${DR_BUCKET}/${region}/provider/${backup_name}.sha256" "${PROVIDER_SNAPSHOT_DIR}/" --only-show-errors
        aws s3 cp "${DR_BUCKET}/${region}/provider/${backup_name}_metadata.json" "${PROVIDER_SNAPSHOT_DIR}/" --only-show-errors
        aws s3 cp "${DR_BUCKET}/${region}/provider/${backup_name}.sig" "${PROVIDER_SNAPSHOT_DIR}/" --only-show-errors 2>/dev/null || true
    fi

    if ! verify_backup "$backup_name"; then
        log_error "Backup verification failed, aborting restore"
        return 1
    fi

    log_warn "This will REPLACE the current provider daemon data/config!"
    if [ "$RESTORE_AUTO_APPROVE" = "0" ]; then
        log_warn "Press Ctrl+C within 10 seconds to abort..."
        sleep 10
    fi

    if [ "$RESTORE_SKIP_SERVICE" = "0" ]; then
        log_info "Stopping provider-daemon service..."
        "$SYSTEMCTL_CMD" stop provider-daemon || true
    else
        log_warn "RESTORE_SKIP_SERVICE=1 set; skipping service stop"
    fi

    local backup_dir="${PROVIDER_HOME}/data.backup.$(date +%s)"
    if [ -d "${PROVIDER_HOME}/data" ]; then
        mv "${PROVIDER_HOME}/data" "$backup_dir"
    fi

    local config_backup_dir="${PROVIDER_HOME}/config.backup.$(date +%s)"
    if [ -d "${PROVIDER_HOME}/config" ]; then
        mv "${PROVIDER_HOME}/config" "$config_backup_dir"
    fi

    local restore_dir
    restore_dir=$(mktemp -d)

    if ! tar -xzf "${PROVIDER_SNAPSHOT_DIR}/${backup_name}.tar.gz" -C "$restore_dir"; then
        log_error "Provider backup extraction failed"
        if [ -d "$backup_dir" ]; then
            mv "$backup_dir" "${PROVIDER_HOME}/data"
        fi
        if [ -d "$config_backup_dir" ]; then
            mv "$config_backup_dir" "${PROVIDER_HOME}/config"
        fi
        rm -rf "$restore_dir"
        notify_webhook "provider.restore" "failure" "Provider backup extraction failed" "$backup_name"
        return 1
    fi

    if [ -d "$restore_dir/data" ]; then
        mv "$restore_dir/data" "${PROVIDER_HOME}/data"
    fi

    if [ -d "$restore_dir/config" ]; then
        mv "$restore_dir/config" "${PROVIDER_HOME}/config"
    fi

    rm -rf "$restore_dir"

    if [ "$RESTORE_SKIP_SERVICE" = "0" ]; then
        log_info "Starting provider-daemon service..."
        "$SYSTEMCTL_CMD" start provider-daemon || true
    else
        log_warn "RESTORE_SKIP_SERVICE=1 set; skipping service start"
    fi

    notify_webhook "provider.restore" "success" "Provider state restore completed" "$backup_name"
    log_info "Provider state restore complete"
}

list_backups() {
    log_info "Available provider backups in ${PROVIDER_SNAPSHOT_DIR}:"
    echo ""
    if ls "${PROVIDER_SNAPSHOT_DIR}"/provider_state_*_metadata.json 1> /dev/null 2>&1; then
        for meta in "${PROVIDER_SNAPSHOT_DIR}"/provider_state_*_metadata.json; do
            local name
            name=$(jq -r '.backup_name' "$meta" 2>/dev/null || basename "$meta" _metadata.json)
            local timestamp
            timestamp=$(jq -r '.timestamp' "$meta" 2>/dev/null || "unknown")
            local archive="${PROVIDER_SNAPSHOT_DIR}/${name}.tar.gz"
            local size="N/A"
            if [ -f "$archive" ]; then
                size=$(du -h "$archive" | cut -f1)
            fi
            printf "  %-40s Size: %-8s %s\n" "$name" "$size" "$timestamp"
        done
    else
        echo "  No provider backups found"
    fi
}

cleanup_old_backups() {
    log_info "Cleaning up old provider backups (keeping ${RETENTION_COUNT})..."
    ls -t "${PROVIDER_SNAPSHOT_DIR}"/provider_state_*.tar.gz 2>/dev/null | tail -n +$((RETENTION_COUNT + 1)) | while read -r archive; do
        local name
        name=$(basename "$archive" .tar.gz)
        rm -f "${PROVIDER_SNAPSHOT_DIR}/${name}.tar.gz" \
            "${PROVIDER_SNAPSHOT_DIR}/${name}.sha256" \
            "${PROVIDER_SNAPSHOT_DIR}/${name}.sig" \
            "${PROVIDER_SNAPSHOT_DIR}/${name}_metadata.json"
    done
}

main() {
    local action="backup"
    local restore_name=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            --backup)
                action="backup"
                shift
                ;;
            --verify)
                action="verify"
                shift
                ;;
            --restore)
                action="restore"
                restore_name="$2"
                shift 2
                ;;
            --list)
                action="list"
                shift
                ;;
            --cleanup)
                action="cleanup"
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [options]"
                echo ""
                echo "Options:"
                echo "  --backup           Create provider state backup (default)"
                echo "  --verify           Verify latest provider backup"
                echo "  --restore NAME     Restore provider state from backup name"
                echo "  --list             List available backups"
                echo "  --cleanup          Remove old backups (keep last RETENTION_COUNT)"
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
        backup)
            local backup_name
            backup_name=$(create_backup)
            upload_backup "$backup_name"
            cleanup_old_backups
            verify_backup "$backup_name"
            ;;
        verify)
            verify_backup ""
            ;;
        restore)
            restore_backup "$restore_name"
            ;;
        list)
            list_backups
            ;;
        cleanup)
            cleanup_old_backups
            ;;
    esac
}

main "$@"
