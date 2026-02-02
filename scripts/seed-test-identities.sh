#!/bin/bash
#
# VirtEngine Test Identity Seeding Script
# Seeds test accounts with pre-verified VEID identities for localnet testing
#
# Usage:
#   ./scripts/seed-test-identities.sh
#
# This script patches the genesis file to include:
#   - Pre-verified identity records with scores
#   - Approved test capture clients
#   - Disabled MFA/identity gating for testing
#
# Run AFTER init-chain.sh but BEFORE starting the node
#

set -e

# Configuration
HOME_DIR="${HOME}/.virtengine"
GENESIS="${HOME_DIR}/config/genesis.json"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log() {
    echo -e "${BLUE}[seed-identities]${NC} $1"
}

success() {
    echo -e "${GREEN}[seed-identities]${NC} $1"
}

error() {
    echo -e "${RED}[seed-identities] ERROR:${NC} $1" >&2
    exit 1
}

warn() {
    echo -e "${YELLOW}[seed-identities] WARNING:${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    if [ ! -f "${GENESIS}" ]; then
        error "Genesis file not found at ${GENESIS}. Run init-chain.sh first."
    fi

    if ! command -v jq &> /dev/null; then
        error "jq is required but not installed."
    fi
}

# Get account address by name
get_account_address() {
    local account="$1"
    virtengine --home "${HOME_DIR}" keys show "${account}" -a --keyring-backend=test 2>/dev/null || echo ""
}

# Add identity record to genesis
add_identity_record() {
    local address="$1"
    local score="$2"
    local tier="$3"
    local status="${4:-verified}"
    
    log "Adding identity record for ${address} with score ${score}"
    
    # Create identity record JSON
    local timestamp
    timestamp=$(date +%s)
    
    local record
    record=$(cat <<EOF
{
    "account_address": "${address}",
    "current_score": ${score},
    "tier": "${tier}",
    "status": "${status}",
    "created_at": ${timestamp},
    "updated_at": ${timestamp},
    "score_version": "test-seed-v1.0.0",
    "scopes": []
}
EOF
)

    # Add to genesis
    jq --argjson record "${record}" \
       '.app_state.veid.identity_records += [$record]' \
       "${GENESIS}" > "${GENESIS}.tmp" && mv "${GENESIS}.tmp" "${GENESIS}"
}

# Add approved test client
add_test_client() {
    local client_id="$1"
    local client_name="$2"
    
    log "Adding approved test client: ${client_id}"
    
    local timestamp
    timestamp=$(date +%s)
    
    # Generate deterministic public key (32 bytes of 0x42 for testing)
    local pubkey
    pubkey=$(printf '42%.0s' {1..64})
    
    local client
    client=$(cat <<EOF
{
    "client_id": "${client_id}",
    "name": "${client_name}",
    "public_key": "${pubkey}",
    "algorithm": "Ed25519",
    "active": true,
    "registered_at": ${timestamp}
}
EOF
)

    # Add to genesis
    jq --argjson client "${client}" \
       '.app_state.veid.approved_clients += [$client]' \
       "${GENESIS}" > "${GENESIS}.tmp" && mv "${GENESIS}.tmp" "${GENESIS}"
}

# Disable identity and MFA gating for testing
disable_gating() {
    log "Disabling identity and MFA gating for testing..."
    
    jq '.app_state.mfa.params.require_at_least_one_factor = false |
        .app_state.mfa.sensitive_tx_configs = [] |
        .app_state.mktplace.params.enable_mfa_gating = false |
        .app_state.mktplace.params.enable_identity_gating = false |
        .app_state.mktplace.params.mfa_configs = [] |
        .app_state.mktplace.mfa_configs = []' \
        "${GENESIS}" > "${GENESIS}.tmp" && mv "${GENESIS}.tmp" "${GENESIS}"
}

# Determine tier from score
get_tier_for_score() {
    local score="$1"
    if [ "${score}" -ge 85 ]; then
        echo "trusted"
    elif [ "${score}" -ge 60 ]; then
        echo "verified"
    elif [ "${score}" -ge 30 ]; then
        echo "standard"
    elif [ "${score}" -ge 1 ]; then
        echo "basic"
    else
        echo "unverified"
    fi
}

# Main seeding logic
seed_identities() {
    log "Seeding test identities..."
    
    # Initialize empty arrays if not present
    jq '.app_state.veid.identity_records = (.app_state.veid.identity_records // []) |
        .app_state.veid.approved_clients = (.app_state.veid.approved_clients // [])' \
        "${GENESIS}" > "${GENESIS}.tmp" && mv "${GENESIS}.tmp" "${GENESIS}"
    
    # Seed provider account with score 80 (meets provider requirement of 70)
    local provider_addr
    provider_addr=$(get_account_address "provider")
    if [ -n "${provider_addr}" ]; then
        add_identity_record "${provider_addr}" 80 "verified" "verified"
    else
        warn "Provider account not found, skipping"
    fi
    
    # Seed operator account with score 85 (meets validator requirement)
    local operator_addr
    operator_addr=$(get_account_address "operator")
    if [ -n "${operator_addr}" ]; then
        add_identity_record "${operator_addr}" 85 "trusted" "verified"
    else
        warn "Operator account not found, skipping"
    fi
    
    # Seed alice with score 60 (standard customer)
    local alice_addr
    alice_addr=$(get_account_address "alice")
    if [ -n "${alice_addr}" ]; then
        add_identity_record "${alice_addr}" 60 "verified" "verified"
    else
        warn "Alice account not found, skipping"
    fi
    
    # Seed bob with score 50 (basic customer)
    local bob_addr
    bob_addr=$(get_account_address "bob")
    if [ -n "${bob_addr}" ]; then
        add_identity_record "${bob_addr}" 50 "standard" "verified"
    else
        warn "Bob account not found, skipping"
    fi
    
    # Seed charlie with score 30 (low score for testing rejection paths)
    local charlie_addr
    charlie_addr=$(get_account_address "charlie")
    if [ -n "${charlie_addr}" ]; then
        add_identity_record "${charlie_addr}" 30 "standard" "verified"
    else
        warn "Charlie account not found, skipping"
    fi
    
    # Add test capture clients
    add_test_client "ve-test-capture-app" "VirtEngine Test Capture App"
    add_test_client "ve-e2e-capture-app" "VirtEngine E2E Test Capture App"
    add_test_client "ve-e2e-onboarding-app" "VirtEngine E2E Onboarding Test Capture App"
}

# Print summary
print_summary() {
    echo ""
    success "Test identities seeded successfully!"
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo "                    Test Account VEID Scores                    "
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    
    for account in provider operator alice bob charlie; do
        local addr
        addr=$(get_account_address "${account}")
        if [ -n "${addr}" ]; then
            local score
            score=$(jq -r ".app_state.veid.identity_records[] | select(.account_address == \"${addr}\") | .current_score" "${GENESIS}" 2>/dev/null || echo "N/A")
            printf "  %-12s %s  (Score: %s)\n" "${account}:" "${addr}" "${score}"
        fi
    done
    
    echo ""
    echo "  Approved Test Clients:"
    jq -r '.app_state.veid.approved_clients[] | "    - \(.client_id)"' "${GENESIS}" 2>/dev/null || echo "    None"
    
    echo ""
    echo "  Gating Status:"
    local mfa_disabled
    mfa_disabled=$(jq -r '.app_state.mfa.params.require_at_least_one_factor' "${GENESIS}" 2>/dev/null || echo "unknown")
    local identity_disabled
    identity_disabled=$(jq -r '.app_state.mktplace.params.enable_identity_gating' "${GENESIS}" 2>/dev/null || echo "unknown")
    echo "    MFA Required: ${mfa_disabled}"
    echo "    Identity Gating: ${identity_disabled}"
    
    echo ""
    echo "═══════════════════════════════════════════════════════════════"
    echo ""
    log "You can now start the chain with: virtengine start"
}

# Main
main() {
    log "VirtEngine Test Identity Seeding"
    log "================================"
    
    check_prerequisites
    seed_identities
    disable_gating
    print_summary
}

main "$@"
