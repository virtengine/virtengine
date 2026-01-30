#!/bin/bash
# scripts/dr/backup-keys.sh
# Secure key backup for disaster recovery
#
# This script handles backup of validator keys, provider keys, and node keys
# with support for HSM-backed keys and Shamir secret sharing.
#
# Usage: ./backup-keys.sh [options]
#   --type TYPE       Key type: validator, provider, node, all (default: all)
#   --output DIR      Output directory for backups
#   --shamir          Use Shamir secret sharing (3-of-5)
#   --verify          Verify existing backups
#   --list            List backed up keys
#
# Environment variables:
#   KEY_DIR           Key directory (default: /opt/virtengine/config)
#   BACKUP_DIR        Backup output directory
#   HSM_ENABLED       Set to 1 if using HSM
#   HSM_MODULE        Path to PKCS#11 module
#   AWS_REGION        AWS region for secrets storage

set -euo pipefail

# Configuration
KEY_DIR="${KEY_DIR:-/opt/virtengine/config}"
BACKUP_DIR="${BACKUP_DIR:-/secure/backups}"
HSM_ENABLED="${HSM_ENABLED:-0}"
HSM_MODULE="${HSM_MODULE:-/usr/lib/softhsm/libsofthsm2.so}"

# Colors
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

# Ensure backup directory exists with proper permissions
ensure_directories() {
    mkdir -p "${BACKUP_DIR}"
    chmod 700 "${BACKUP_DIR}"
}

# Generate a secure passphrase
generate_passphrase() {
    openssl rand -base64 32
}

# Get passphrase from AWS Secrets Manager or generate new
get_or_create_passphrase() {
    local key_type="$1"
    local secret_name="virtengine/dr/backup-passphrase-${key_type}"
    
    # Try to get existing passphrase
    local passphrase=$(aws secretsmanager get-secret-value \
        --secret-id "$secret_name" \
        --query SecretString \
        --output text 2>/dev/null || true)
    
    if [ -z "$passphrase" ]; then
        log_info "Creating new passphrase for ${key_type} backups"
        passphrase=$(generate_passphrase)
        
        # Store in Secrets Manager
        aws secretsmanager create-secret \
            --name "$secret_name" \
            --secret-string "$passphrase" \
            --description "Backup passphrase for VirtEngine ${key_type} keys" \
            --tags Key=Purpose,Value=DR Key=KeyType,Value="${key_type}" \
            > /dev/null 2>&1 || \
        aws secretsmanager put-secret-value \
            --secret-id "$secret_name" \
            --secret-string "$passphrase" \
            > /dev/null 2>&1
    fi
    
    echo "$passphrase"
}

# Backup validator keys
backup_validator_keys() {
    local timestamp=$(date -u +%Y%m%d_%H%M%S)
    local backup_name="validator_keys_${timestamp}"
    
    log_info "Backing up validator keys..."
    
    if [ "$HSM_ENABLED" == "1" ]; then
        backup_hsm_validator_keys "$backup_name"
    else
        backup_file_validator_keys "$backup_name"
    fi
}

# Backup HSM-backed validator keys (reference only)
backup_hsm_validator_keys() {
    local backup_name="$1"
    
    log_info "Creating HSM key reference backup..."
    
    # List keys in HSM
    local key_list=$(pkcs11-tool --module "${HSM_MODULE}" --list-objects --type privkey 2>/dev/null || true)
    
    if [ -z "$key_list" ]; then
        log_warn "No HSM keys found or HSM not accessible"
        return 0
    fi
    
    # Create HSM reference file
    cat > "${BACKUP_DIR}/${backup_name}_hsm_ref.json" << EOF
{
    "type": "hsm_reference",
    "timestamp": "$(date -u +%FT%TZ)",
    "backup_name": "${backup_name}",
    "hsm_module": "${HSM_MODULE}",
    "hsm_slot": "${HSM_SLOT:-0}",
    "keys": [],
    "recovery_instructions": [
        "1. Contact security team for HSM access",
        "2. Use HSM vendor recovery procedures",
        "3. Restore from HSM backup media if needed",
        "4. Re-initialize validator with HSM key"
    ],
    "security_note": "HSM keys are not exported for security. This file contains references only."
}
EOF
    
    # Get key labels
    local labels=$(echo "$key_list" | grep -oP 'label:\s+\K.*' || true)
    
    # Update JSON with key labels
    if [ -n "$labels" ]; then
        local keys_json=$(echo "$labels" | jq -R -s 'split("\n") | map(select(length > 0))')
        jq ".keys = $keys_json" "${BACKUP_DIR}/${backup_name}_hsm_ref.json" > /tmp/hsm_ref.json
        mv /tmp/hsm_ref.json "${BACKUP_DIR}/${backup_name}_hsm_ref.json"
    fi
    
    log_info "HSM reference backup created: ${backup_name}_hsm_ref.json"
    log_warn "Remember: HSM keys are NOT exported. This is a reference file only."
}

# Backup file-based validator keys
backup_file_validator_keys() {
    local backup_name="$1"
    local priv_key="${KEY_DIR}/priv_validator_key.json"
    local node_key="${KEY_DIR}/node_key.json"
    local priv_state="${KEY_DIR}/priv_validator_state.json"
    
    # WARNING: Backing up priv_validator_key.json is risky for active validators
    # as it could lead to double-signing if restored incorrectly
    
    if [ -f "$priv_key" ]; then
        log_warn "CRITICAL: priv_validator_key.json backup is sensitive!"
        log_warn "Improper restoration can cause double-signing and slashing!"
        
        # Get passphrase
        local passphrase=$(get_or_create_passphrase "validator")
        
        # Create encrypted backup
        local temp_archive=$(mktemp)
        tar -czf "$temp_archive" -C "${KEY_DIR}" \
            priv_validator_key.json \
            priv_validator_state.json 2>/dev/null || \
            tar -czf "$temp_archive" -C "${KEY_DIR}" priv_validator_key.json
        
        # Encrypt with AES-256-GCM
        openssl enc -aes-256-gcm -salt -pbkdf2 -iter 100000 \
            -in "$temp_archive" \
            -out "${BACKUP_DIR}/${backup_name}_encrypted.tar.gz.enc" \
            -pass pass:"$passphrase"
        
        rm -f "$temp_archive"
        
        # Generate checksum
        sha256sum "${BACKUP_DIR}/${backup_name}_encrypted.tar.gz.enc" \
            > "${BACKUP_DIR}/${backup_name}.sha256"
        
        # Create metadata
        cat > "${BACKUP_DIR}/${backup_name}_metadata.json" << EOF
{
    "type": "validator_keys",
    "timestamp": "$(date -u +%FT%TZ)",
    "backup_name": "${backup_name}",
    "encryption": "AES-256-GCM",
    "kdf": "PBKDF2-100000",
    "includes_state": $([ -f "$priv_state" ] && echo "true" || echo "false"),
    "validator_address": "$(jq -r '.address' "$priv_key" 2>/dev/null || echo 'unknown')",
    "pub_key_type": "$(jq -r '.pub_key.type' "$priv_key" 2>/dev/null || echo 'unknown')",
    "warning": "DO NOT restore this key while another validator instance is running!",
    "passphrase_location": "AWS Secrets Manager: virtengine/dr/backup-passphrase-validator"
}
EOF
        
        log_info "Validator key backup created: ${backup_name}"
        
        # Clear passphrase from memory (best effort)
        unset passphrase
    else
        log_warn "priv_validator_key.json not found in ${KEY_DIR}"
    fi
}

# Backup node keys
backup_node_keys() {
    local timestamp=$(date -u +%Y%m%d_%H%M%S)
    local backup_name="node_keys_${timestamp}"
    local node_key="${KEY_DIR}/node_key.json"
    
    log_info "Backing up node keys..."
    
    if [ -f "$node_key" ]; then
        local passphrase=$(get_or_create_passphrase "node")
        
        # Encrypt node key
        openssl enc -aes-256-gcm -salt -pbkdf2 -iter 100000 \
            -in "$node_key" \
            -out "${BACKUP_DIR}/${backup_name}_encrypted.json.enc" \
            -pass pass:"$passphrase"
        
        # Generate checksum
        sha256sum "${BACKUP_DIR}/${backup_name}_encrypted.json.enc" \
            > "${BACKUP_DIR}/${backup_name}.sha256"
        
        # Create metadata
        cat > "${BACKUP_DIR}/${backup_name}_metadata.json" << EOF
{
    "type": "node_key",
    "timestamp": "$(date -u +%FT%TZ)",
    "backup_name": "${backup_name}",
    "encryption": "AES-256-GCM",
    "node_id": "$(jq -r '.priv_key.value' "$node_key" 2>/dev/null | head -c 40 || echo 'unknown')",
    "note": "Node keys can be regenerated if lost. This backup is for convenience."
}
EOF
        
        log_info "Node key backup created: ${backup_name}"
        unset passphrase
    else
        log_warn "node_key.json not found in ${KEY_DIR}"
    fi
}

# Backup provider daemon keys
backup_provider_keys() {
    local timestamp=$(date -u +%Y%m%d_%H%M%S)
    local backup_name="provider_keys_${timestamp}"
    local provider_home="${PROVIDER_HOME:-/opt/provider-daemon}"
    
    log_info "Backing up provider daemon keys..."
    
    # Check if provider-daemon CLI is available
    if command -v provider-daemon &> /dev/null; then
        # Use built-in backup command
        provider-daemon keys backup \
            --output "${BACKUP_DIR}/${backup_name}.backup" \
            --encrypt \
            2>/dev/null || {
                log_warn "provider-daemon backup command failed, using manual backup"
                backup_provider_keys_manual "$backup_name" "$provider_home"
            }
    else
        backup_provider_keys_manual "$backup_name" "$provider_home"
    fi
}

# Manual backup of provider keys
backup_provider_keys_manual() {
    local backup_name="$1"
    local provider_home="$2"
    local key_files=(
        "keys/provider.key"
        "config/provider.yaml"
        "config/client.yaml"
    )
    
    local temp_dir=$(mktemp -d)
    trap "rm -rf $temp_dir" EXIT
    
    # Collect key files
    for file in "${key_files[@]}"; do
        if [ -f "${provider_home}/${file}" ]; then
            mkdir -p "$(dirname "${temp_dir}/${file}")"
            cp "${provider_home}/${file}" "${temp_dir}/${file}"
        fi
    done
    
    # Check if we found any files
    if [ -z "$(ls -A "$temp_dir")" ]; then
        log_warn "No provider key files found in ${provider_home}"
        return 0
    fi
    
    local passphrase=$(get_or_create_passphrase "provider")
    
    # Create encrypted archive
    tar -czf - -C "$temp_dir" . | \
        openssl enc -aes-256-gcm -salt -pbkdf2 -iter 100000 \
            -out "${BACKUP_DIR}/${backup_name}_encrypted.tar.gz.enc" \
            -pass pass:"$passphrase"
    
    # Generate checksum
    sha256sum "${BACKUP_DIR}/${backup_name}_encrypted.tar.gz.enc" \
        > "${BACKUP_DIR}/${backup_name}.sha256"
    
    # Create metadata
    cat > "${BACKUP_DIR}/${backup_name}_metadata.json" << EOF
{
    "type": "provider_keys",
    "timestamp": "$(date -u +%FT%TZ)",
    "backup_name": "${backup_name}",
    "encryption": "AES-256-GCM",
    "provider_home": "${provider_home}",
    "files_included": $(printf '%s\n' "${key_files[@]}" | jq -R -s 'split("\n") | map(select(length > 0))')
}
EOF
    
    log_info "Provider key backup created: ${backup_name}"
    unset passphrase
}

# Split backup using Shamir Secret Sharing
create_shamir_shares() {
    local backup_file="$1"
    local threshold="${2:-3}"
    local total_shares="${3:-5}"
    
    log_info "Creating Shamir secret shares (${threshold}-of-${total_shares})..."
    
    if ! command -v ssss-split &> /dev/null; then
        log_error "ssss-split not found. Install ssss package."
        return 1
    fi
    
    local backup_name=$(basename "$backup_file" | sed 's/_encrypted.*//')
    local shares_dir="${BACKUP_DIR}/${backup_name}_shares"
    mkdir -p "$shares_dir"
    chmod 700 "$shares_dir"
    
    # Split the encrypted backup
    ssss-split -t "$threshold" -n "$total_shares" -w "${backup_name}" \
        < "$backup_file" \
        | while read -r share; do
            local index=$(echo "$share" | cut -d'-' -f2)
            echo "$share" > "${shares_dir}/share_${index}.txt"
            chmod 600 "${shares_dir}/share_${index}.txt"
        done
    
    # Create share metadata
    cat > "${shares_dir}/shares_metadata.json" << EOF
{
    "type": "shamir_shares",
    "timestamp": "$(date -u +%FT%TZ)",
    "original_backup": "$(basename $backup_file)",
    "threshold": ${threshold},
    "total_shares": ${total_shares},
    "shares": [
$(for i in $(seq 1 $total_shares); do
    echo "        {"
    echo "            \"index\": $i,"
    echo "            \"file\": \"share_${i}.txt\","
    echo "            \"checksum\": \"$(sha256sum "${shares_dir}/share_${i}.txt" 2>/dev/null | cut -d' ' -f1 || echo 'N/A')\""
    if [ $i -lt $total_shares ]; then
        echo "        },"
    else
        echo "        }"
    fi
done)
    ],
    "recovery_instructions": [
        "1. Collect at least ${threshold} shares from different custodians",
        "2. Run: cat share_1.txt share_2.txt share_N.txt | ssss-combine -t ${threshold}",
        "3. Decrypt the restored backup with the stored passphrase"
    ]
}
EOF
    
    log_info "Shamir shares created in: ${shares_dir}"
    log_info "Distribute shares to different secure locations!"
    
    # Suggest distribution
    echo ""
    log_info "Recommended share distribution:"
    echo "  Share 1: Primary secure storage (on-premises safe)"
    echo "  Share 2: Secondary secure storage (bank safe deposit)"
    echo "  Share 3: Offline storage (air-gapped system)"
    echo "  Share 4: Executive custody"
    echo "  Share 5: Security team custody"
}

# Upload backup to remote storage
upload_backup() {
    local backup_name="$1"
    local bucket="${DR_BUCKET:-}"
    
    if [ -z "$bucket" ]; then
        log_warn "DR_BUCKET not set, skipping remote upload"
        return 0
    fi
    
    log_info "Uploading backup to ${bucket}..."
    
    # Upload all files with the backup name prefix
    for file in "${BACKUP_DIR}/${backup_name}"*; do
        if [ -f "$file" ]; then
            aws s3 cp "$file" "${bucket}/keys/$(basename $file)" \
                --sse aws:kms \
                --storage-class STANDARD_IA \
                --only-show-errors
        fi
    done
    
    log_info "Upload complete"
}

# Verify backup integrity
verify_backup() {
    local backup_type="${1:-all}"
    local errors=0
    
    log_info "Verifying backups..."
    
    # Find all checksum files
    for sha_file in "${BACKUP_DIR}"/*.sha256; do
        if [ -f "$sha_file" ]; then
            local backup_name=$(basename "$sha_file" .sha256)
            
            # Filter by type if specified
            if [ "$backup_type" != "all" ] && [[ ! "$backup_name" =~ ^${backup_type} ]]; then
                continue
            fi
            
            log_info "Verifying: ${backup_name}"
            
            if (cd "${BACKUP_DIR}" && sha256sum -c "$sha_file" > /dev/null 2>&1); then
                echo "  ✓ Checksum: PASS"
            else
                echo "  ✗ Checksum: FAIL"
                ((errors++))
            fi
            
            # Check metadata exists
            if [ -f "${BACKUP_DIR}/${backup_name}_metadata.json" ]; then
                echo "  ✓ Metadata: Present"
            else
                echo "  ✗ Metadata: Missing"
                ((errors++))
            fi
        fi
    done
    
    if [ $errors -gt 0 ]; then
        log_error "Verification failed with $errors errors"
        return 1
    fi
    
    log_info "All backups verified successfully"
    return 0
}

# List backups
list_backups() {
    log_info "Available key backups in ${BACKUP_DIR}:"
    echo ""
    
    for meta_file in "${BACKUP_DIR}"/*_metadata.json; do
        if [ -f "$meta_file" ]; then
            local backup_name=$(jq -r '.backup_name' "$meta_file" 2>/dev/null || basename "$meta_file" _metadata.json)
            local backup_type=$(jq -r '.type' "$meta_file" 2>/dev/null || "unknown")
            local timestamp=$(jq -r '.timestamp' "$meta_file" 2>/dev/null || "unknown")
            
            printf "  %-40s Type: %-15s %s\n" "$backup_name" "$backup_type" "$timestamp"
        fi
    done
    
    if [ ! -f "${BACKUP_DIR}"/*_metadata.json 2>/dev/null ]; then
        echo "  No backups found"
    fi
}

# Clean old backups (keep last N)
cleanup_old_backups() {
    local keep="${1:-10}"
    local backup_type="${2:-all}"
    
    log_info "Cleaning up old backups (keeping ${keep} per type)..."
    
    for type in validator_keys node_keys provider_keys; do
        if [ "$backup_type" != "all" ] && [ "$backup_type" != "$type" ]; then
            continue
        fi
        
        # Get sorted list of backups
        local backups=$(ls -t "${BACKUP_DIR}/${type}_"*_metadata.json 2>/dev/null | tail -n +$((keep + 1)))
        
        for meta_file in $backups; do
            local backup_name=$(basename "$meta_file" _metadata.json)
            log_info "Removing old backup: ${backup_name}"
            
            rm -f "${BACKUP_DIR}/${backup_name}"*
        done
    done
}

# Main execution
main() {
    local key_type="all"
    local use_shamir=false
    local action="backup"
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --type)
                key_type="$2"
                shift 2
                ;;
            --output)
                BACKUP_DIR="$2"
                shift 2
                ;;
            --shamir)
                use_shamir=true
                shift
                ;;
            --verify)
                action="verify"
                shift
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
                echo "  --type TYPE       Key type: validator, provider, node, all (default: all)"
                echo "  --output DIR      Output directory for backups"
                echo "  --shamir          Use Shamir secret sharing (3-of-5)"
                echo "  --verify          Verify existing backups"
                echo "  --list            List backed up keys"
                echo "  --cleanup         Remove old backups (keep last 10)"
                echo "  --help            Show this help message"
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
            case $key_type in
                validator)
                    backup_validator_keys
                    ;;
                provider)
                    backup_provider_keys
                    ;;
                node)
                    backup_node_keys
                    ;;
                all)
                    backup_validator_keys
                    backup_node_keys
                    backup_provider_keys
                    ;;
                *)
                    log_error "Unknown key type: $key_type"
                    exit 1
                    ;;
            esac
            
            # Create Shamir shares if requested
            if [ "$use_shamir" = true ]; then
                for enc_file in "${BACKUP_DIR}"/*_encrypted.*.enc; do
                    if [ -f "$enc_file" ]; then
                        create_shamir_shares "$enc_file"
                    fi
                done
            fi
            
            # Verify after backup
            verify_backup "$key_type"
            ;;
        verify)
            verify_backup "$key_type"
            ;;
        list)
            list_backups
            ;;
        cleanup)
            cleanup_old_backups 10 "$key_type"
            ;;
    esac
}

main "$@"
