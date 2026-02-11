#!/bin/bash
# scripts/dr/dr-test.sh
# Automated disaster recovery testing suite
#
# This script validates DR procedures are working correctly.
# Run regularly (daily/weekly) to ensure DR readiness.
#
# Usage: ./dr-test.sh [options]
#   --test TEST       Run specific test (backup, restore, failover, connectivity, all)
#   --environment     Environment (staging, production) - default: staging
#   --report          Generate HTML report
#   --notify          Send Slack notification
#
# Environment variables:
#   DR_BUCKET         S3 bucket for DR backups
#   SLACK_WEBHOOK     Slack webhook for notifications

set -euo pipefail

# Configuration
TEST_ENVIRONMENT="${TEST_ENVIRONMENT:-staging}"
DR_BUCKET="${DR_BUCKET:-}"
SLACK_WEBHOOK="${SLACK_WEBHOOK:-}"
RESULTS_DIR="${RESULTS_DIR:-/var/log/dr-tests}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test results tracking
declare -A TEST_RESULTS
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
TEST_START_TIME=$(date +%s)

log_info() {
    echo -e "${GREEN}[$(date -u +%FT%TZ)] INFO:${NC} $*"
}

log_warn() {
    echo -e "${YELLOW}[$(date -u +%FT%TZ)] WARN:${NC} $*"
}

log_error() {
    echo -e "${RED}[$(date -u +%FT%TZ)] ERROR:${NC} $*" >&2
}

log_test() {
    echo -e "${BLUE}[$(date -u +%FT%TZ)] TEST:${NC} $*"
}

# Record test result
record_result() {
    local test_name="$1"
    local passed="$2"
    local duration="$3"
    local details="${4:-}"
    
    ((TOTAL_TESTS++))
    
    if [ "$passed" = true ]; then
        ((PASSED_TESTS++))
        TEST_RESULTS["$test_name"]="PASS|${duration}|${details}"
        echo -e "  ${GREEN}✓ ${test_name}${NC} (${duration}s)"
    else
        ((FAILED_TESTS++))
        TEST_RESULTS["$test_name"]="FAIL|${duration}|${details}"
        echo -e "  ${RED}✗ ${test_name}${NC} (${duration}s) - ${details}"
    fi
}

# Get region from hostname or metadata
get_region() {
    curl -s --connect-timeout 2 http://169.254.169.254/latest/meta-data/placement/region 2>/dev/null || \
        hostname | cut -d'-' -f1-2 2>/dev/null || echo "unknown"
}

# Test 1: Backup Integrity
test_backup_integrity() {
    log_test "Testing backup integrity..."
    local start=$(date +%s)
    local passed=true
    local details=""
    
    # Check local backups exist
    local local_snapshots=$(ls /data/snapshots/state_*_data.tar.gz 2>/dev/null | wc -l)
    if [ "$local_snapshots" -lt 1 ]; then
        passed=false
        details="No local snapshots found"
    fi
    
    # Check remote backups if bucket configured
    if [ -n "$DR_BUCKET" ] && [ "$passed" = true ]; then
        local remote_count=$(aws s3 ls "${DR_BUCKET}/$(get_region)/state/" 2>/dev/null | grep -c "_data.tar.gz" || echo "0")
        if [ "$remote_count" -lt 1 ]; then
            passed=false
            details="No remote backups found in ${DR_BUCKET}"
        fi
    fi
    
    # Verify latest backup checksum
    if [ "$passed" = true ]; then
        local latest=$(ls -t /data/snapshots/state_*_data.tar.gz 2>/dev/null | head -1)
        if [ -n "$latest" ]; then
            local checksum_file="${latest%_data.tar.gz}.sha256"
            if [ -f "$checksum_file" ]; then
                if ! (cd "$(dirname "$latest")" && sha256sum -c "$checksum_file" > /dev/null 2>&1); then
                    passed=false
                    details="Checksum verification failed"
                fi
            fi
        fi
    fi
    
    local duration=$(($(date +%s) - start))
    record_result "backup_integrity" "$passed" "$duration" "$details"
}

# Test 2: Backup Age
test_backup_age() {
    log_test "Testing backup freshness..."
    local start=$(date +%s)
    local passed=true
    local details=""
    local max_age_hours=24
    
    # Find latest backup
    local latest=$(ls -t /data/snapshots/state_*_metadata.json 2>/dev/null | head -1)
    
    if [ -z "$latest" ]; then
        passed=false
        details="No backup metadata found"
    else
        local backup_timestamp=$(jq -r '.timestamp' "$latest" 2>/dev/null)
        if [ -n "$backup_timestamp" ] && [ "$backup_timestamp" != "null" ]; then
            local backup_epoch=$(date -d "$backup_timestamp" +%s 2>/dev/null || echo "0")
            local now_epoch=$(date +%s)
            local age_hours=$(( (now_epoch - backup_epoch) / 3600 ))
            
            if [ "$age_hours" -gt "$max_age_hours" ]; then
                passed=false
                details="Latest backup is ${age_hours} hours old (max: ${max_age_hours})"
            else
                details="Latest backup is ${age_hours} hours old"
            fi
        fi
    fi
    
    local duration=$(($(date +%s) - start))
    record_result "backup_age" "$passed" "$duration" "$details"
}

# Test 3: Key Backup Existence
test_key_backups() {
    log_test "Testing key backups..."
    local start=$(date +%s)
    local passed=true
    local details=""
    local backup_dir="${BACKUP_DIR:-/secure/backups}"
    
    # Check validator key backup
    if [ -d "$backup_dir" ]; then
        local val_backups=$(ls "${backup_dir}"/validator_keys_*_metadata.json 2>/dev/null | wc -l)
        if [ "$val_backups" -lt 1 ]; then
            passed=false
            details="No validator key backups found"
        fi
        
        # Check provider key backup
        local prov_backups=$(ls "${backup_dir}"/provider_keys_*_metadata.json 2>/dev/null | wc -l)
        if [ "$prov_backups" -lt 1 ]; then
            if [ "$passed" = true ]; then
                details="No provider key backups found"
            else
                details="${details}; No provider key backups"
            fi
            # Provider keys are optional, so don't fail
        fi
    else
        passed=false
        details="Backup directory not found: ${backup_dir}"
    fi
    
    local duration=$(($(date +%s) - start))
    record_result "key_backups" "$passed" "$duration" "$details"
}

# Test 4: Provider State Backup
test_provider_state_backup() {
    log_test "Testing provider state backups..."
    local start=$(date +%s)
    local passed=true
    local details=""
    local provider_dir="${PROVIDER_SNAPSHOT_DIR:-/data/provider-snapshots}"

    if [ -d "$provider_dir" ]; then
        local backups=$(ls "${provider_dir}"/provider_state_*_metadata.json 2>/dev/null | wc -l)
        if [ "$backups" -lt 1 ]; then
            passed=false
            details="No provider state backups found"
        else
            if [ -x "${SCRIPT_DIR}/backup-provider-state.sh" ]; then
                if ! "${SCRIPT_DIR}/backup-provider-state.sh" --verify > /dev/null 2>&1; then
                    passed=false
                    details="Provider backup verification failed"
                else
                    details="Provider backups present and verified"
                fi
            else
                details="Provider backups present (verification script missing)"
            fi
        fi
    else
        passed=false
        details="Provider backup directory not found: ${provider_dir}"
    fi

    local duration=$(($(date +%s) - start))
    record_result "provider_state_backup" "$passed" "$duration" "$details"
}

# Test 4: State Sync Endpoints
test_state_sync_endpoints() {
    log_test "Testing state sync endpoints..."
    local start=$(date +%s)
    local passed=true
    local details=""
    
    local endpoints=(
        "rpc-us-east.virtengine.network:26657"
        "rpc-eu-west.virtengine.network:26657"
        "rpc-ap-south.virtengine.network:26657"
    )
    
    local available=0
    local failed_endpoints=""
    
    for endpoint in "${endpoints[@]}"; do
        if curl -s --connect-timeout 5 "http://${endpoint}/status" | jq -e '.result.sync_info' > /dev/null 2>&1; then
            ((available++))
        else
            failed_endpoints="${failed_endpoints}${endpoint} "
        fi
    done
    
    if [ "$available" -lt 2 ]; then
        passed=false
        details="Only ${available} of ${#endpoints[@]} endpoints available. Failed: ${failed_endpoints}"
    else
        details="${available} of ${#endpoints[@]} endpoints available"
    fi
    
    local duration=$(($(date +%s) - start))
    record_result "state_sync_endpoints" "$passed" "$duration" "$details"
}

# Test 5: Cross-Region Connectivity
test_cross_region_connectivity() {
    log_test "Testing cross-region connectivity..."
    local start=$(date +%s)
    local passed=true
    local details=""
    
    local regions=("us-east" "eu-west" "ap-south")
    local connected=0
    
    for region in "${regions[@]}"; do
        # Check P2P connectivity
        if timeout 10 curl -s "http://rpc-${region}.virtengine.network:26657/net_info" | \
            jq -e '.result.n_peers | tonumber > 0' > /dev/null 2>&1; then
            ((connected++))
        fi
    done
    
    if [ "$connected" -lt 2 ]; then
        passed=false
        details="Only ${connected} of ${#regions[@]} regions connected"
    else
        details="${connected} of ${#regions[@]} regions connected"
    fi
    
    local duration=$(($(date +%s) - start))
    record_result "cross_region_connectivity" "$passed" "$duration" "$details"
}

# Test 6: DNS Health
test_dns_health() {
    log_test "Testing DNS health..."
    local start=$(date +%s)
    local passed=true
    local details=""
    
    local dns_records=(
        "rpc.virtengine.network"
        "api.virtengine.network"
    )
    
    for record in "${dns_records[@]}"; do
        local ips=$(dig +short "$record" 2>/dev/null)
        if [ -z "$ips" ]; then
            passed=false
            details="${details}${record} failed to resolve; "
        fi
    done
    
    if [ "$passed" = true ]; then
        details="All DNS records resolving"
    fi
    
    local duration=$(($(date +%s) - start))
    record_result "dns_health" "$passed" "$duration" "$details"
}

# Test 7: S3 Bucket Access
test_s3_access() {
    log_test "Testing S3 backup bucket access..."
    local start=$(date +%s)
    local passed=true
    local details=""
    
    if [ -z "$DR_BUCKET" ]; then
        details="DR_BUCKET not configured"
        passed=false
    else
        # Test list access
        if ! aws s3 ls "$DR_BUCKET" > /dev/null 2>&1; then
            passed=false
            details="Cannot list bucket ${DR_BUCKET}"
        fi
        
        # Test write access
        local test_file="/tmp/dr_test_$$"
        echo "DR test $(date)" > "$test_file"
        if ! aws s3 cp "$test_file" "${DR_BUCKET}/test/dr_test_$(date +%s).txt" --only-show-errors 2>/dev/null; then
            passed=false
            details="Cannot write to bucket ${DR_BUCKET}"
        else
            # Cleanup
            aws s3 rm "${DR_BUCKET}/test/dr_test_$(date +%s).txt" --only-show-errors 2>/dev/null || true
        fi
        rm -f "$test_file"
    fi
    
    if [ "$passed" = true ]; then
        details="S3 bucket access OK"
    fi
    
    local duration=$(($(date +%s) - start))
    record_result "s3_access" "$passed" "$duration" "$details"
}

# Test 8: Secrets Manager Access
test_secrets_access() {
    log_test "Testing Secrets Manager access..."
    local start=$(date +%s)
    local passed=true
    local details=""
    
    # Try to read a DR secret
    if ! aws secretsmanager describe-secret \
        --secret-id "virtengine/dr/backup-passphrase-validator" \
        > /dev/null 2>&1; then
        # Secret might not exist yet, try to list secrets
        if ! aws secretsmanager list-secrets --max-results 1 > /dev/null 2>&1; then
            passed=false
            details="Cannot access Secrets Manager"
        else
            details="Secrets Manager accessible (no DR secrets yet)"
        fi
    else
        details="DR secrets accessible"
    fi
    
    local duration=$(($(date +%s) - start))
    record_result "secrets_access" "$passed" "$duration" "$details"
}

# Test 9: Node Health
test_node_health() {
    log_test "Testing local node health..."
    local start=$(date +%s)
    local passed=true
    local details=""
    
    # Check node is running
    if ! virtengine status > /dev/null 2>&1; then
        passed=false
        details="Node is not responding"
    else
        # Check sync status
        local catching_up=$(virtengine status 2>&1 | jq -r '.sync_info.catching_up')
        local height=$(virtengine status 2>&1 | jq -r '.sync_info.latest_block_height')
        
        if [ "$catching_up" = "true" ]; then
            passed=false
            details="Node is syncing (height: ${height})"
        else
            details="Node healthy at height ${height}"
        fi
    fi
    
    local duration=$(($(date +%s) - start))
    record_result "node_health" "$passed" "$duration" "$details"
}

# Test 10: Backup Restore (Dry Run)
test_backup_restore_dryrun() {
    log_test "Testing backup restore (dry run)..."
    local start=$(date +%s)
    local passed=true
    local details=""
    
    # Find latest backup
    local latest=$(ls -t /data/snapshots/state_*_data.tar.gz 2>/dev/null | head -1)
    
    if [ -z "$latest" ]; then
        passed=false
        details="No backup found to test"
    else
        # Test archive can be read
        if ! tar -tzf "$latest" > /dev/null 2>&1; then
            passed=false
            details="Backup archive is corrupted"
        else
            # Check archive contents
            local has_data=$(tar -tzf "$latest" | grep -c "^data/" || echo "0")
            if [ "$has_data" -lt 1 ]; then
                passed=false
                details="Backup does not contain data directory"
            else
                details="Backup can be restored (contains data/)"
            fi
        fi
    fi
    
    local duration=$(($(date +%s) - start))
    record_result "backup_restore_dryrun" "$passed" "$duration" "$details"
}

# Run all tests
run_all_tests() {
    log_info "Starting DR test suite..."
    log_info "Environment: ${TEST_ENVIRONMENT}"
    log_info "Region: $(get_region)"
    echo ""
    
    test_backup_integrity
    test_backup_age
    test_key_backups
    test_provider_state_backup
    test_state_sync_endpoints
    test_cross_region_connectivity
    test_dns_health
    test_s3_access
    test_secrets_access
    test_node_health
    test_backup_restore_dryrun
}

# Generate summary
generate_summary() {
    local total_duration=$(($(date +%s) - TEST_START_TIME))
    
    echo ""
    echo "=============================================="
    echo "DR Test Summary"
    echo "=============================================="
    echo "Environment: ${TEST_ENVIRONMENT}"
    echo "Region: $(get_region)"
    echo "Timestamp: $(date -u +%FT%TZ)"
    echo "Duration: ${total_duration}s"
    echo ""
    echo "Results:"
    echo "  Total:  ${TOTAL_TESTS}"
    echo "  Passed: ${PASSED_TESTS}"
    echo "  Failed: ${FAILED_TESTS}"
    echo ""
    
    if [ "$FAILED_TESTS" -gt 0 ]; then
        echo -e "${RED}OVERALL: FAILED${NC}"
        return 1
    else
        echo -e "${GREEN}OVERALL: PASSED${NC}"
        return 0
    fi
}

# Generate JSON report
generate_json_report() {
    local output_file="${RESULTS_DIR}/dr_test_$(date -u +%Y%m%d_%H%M%S).json"
    local total_duration=$(($(date +%s) - TEST_START_TIME))
    
    mkdir -p "${RESULTS_DIR}"
    
    cat > "$output_file" << EOF
{
    "timestamp": "$(date -u +%FT%TZ)",
    "environment": "${TEST_ENVIRONMENT}",
    "region": "$(get_region)",
    "duration_seconds": ${total_duration},
    "summary": {
        "total": ${TOTAL_TESTS},
        "passed": ${PASSED_TESTS},
        "failed": ${FAILED_TESTS}
    },
    "tests": [
$(
    local first=true
    for test_name in "${!TEST_RESULTS[@]}"; do
        IFS='|' read -r status duration details <<< "${TEST_RESULTS[$test_name]}"
        if [ "$first" = true ]; then
            first=false
        else
            echo ","
        fi
        echo -n "        {\"name\": \"${test_name}\", \"status\": \"${status}\", \"duration_seconds\": ${duration}, \"details\": \"${details}\"}"
    done
)
    ]
}
EOF
    
    echo ""
    log_info "Report saved to: ${output_file}"
    
    # Upload to S3 if configured
    if [ -n "$DR_BUCKET" ]; then
        aws s3 cp "$output_file" "${DR_BUCKET}/test-results/$(basename $output_file)" --only-show-errors
    fi
}

# Send Slack notification
send_slack_notification() {
    if [ -z "$SLACK_WEBHOOK" ]; then
        return 0
    fi
    
    local color
    local status
    
    if [ "$FAILED_TESTS" -gt 0 ]; then
        color="danger"
        status="FAILED"
    else
        color="good"
        status="PASSED"
    fi
    
    local message="*DR Test Results - ${TEST_ENVIRONMENT}*\n"
    message+="Status: ${status}\n"
    message+="Region: $(get_region)\n"
    message+="Passed: ${PASSED_TESTS}/${TOTAL_TESTS}\n"
    
    if [ "$FAILED_TESTS" -gt 0 ]; then
        message+="\nFailed tests:\n"
        for test_name in "${!TEST_RESULTS[@]}"; do
            IFS='|' read -r status duration details <<< "${TEST_RESULTS[$test_name]}"
            if [ "$status" = "FAIL" ]; then
                message+="• ${test_name}: ${details}\n"
            fi
        done
    fi
    
    curl -s -X POST "$SLACK_WEBHOOK" \
        -H "Content-Type: application/json" \
        -d "{
            \"attachments\": [{
                \"color\": \"${color}\",
                \"text\": \"${message}\",
                \"footer\": \"DR Test Suite\",
                \"ts\": $(date +%s)
            }]
        }" > /dev/null
}

# Main execution
main() {
    local test_type="all"
    local generate_report=false
    local send_notification=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --test)
                test_type="$2"
                shift 2
                ;;
            --environment)
                TEST_ENVIRONMENT="$2"
                shift 2
                ;;
            --report)
                generate_report=true
                shift
                ;;
            --notify)
                send_notification=true
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [options]"
                echo ""
                echo "Options:"
                echo "  --test TEST       Run specific test (backup, restore, failover, connectivity, all)"
                echo "  --environment     Environment (staging, production)"
                echo "  --report          Generate JSON report"
                echo "  --notify          Send Slack notification"
                echo "  --help            Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Run tests based on type
    case $test_type in
        all)
            run_all_tests
            ;;
        backup)
            test_backup_integrity
            test_backup_age
            test_key_backups
            test_provider_state_backup
            test_backup_restore_dryrun
            ;;
        connectivity)
            test_state_sync_endpoints
            test_cross_region_connectivity
            test_dns_health
            ;;
        infrastructure)
            test_s3_access
            test_secrets_access
            test_node_health
            ;;
        *)
            log_error "Unknown test type: $test_type"
            exit 1
            ;;
    esac
    
    # Generate summary
    local overall_result=0
    generate_summary || overall_result=$?
    
    # Generate report if requested
    if [ "$generate_report" = true ]; then
        generate_json_report
    fi
    
    # Send notification if requested
    if [ "$send_notification" = true ]; then
        send_slack_notification
    fi
    
    exit $overall_result
}

main "$@"
