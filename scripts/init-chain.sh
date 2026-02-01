#!/bin/bash
#
# VirtEngine Chain Initialization Script
# Initializes the chain with test accounts for local development
#
# Usage:
#   ./scripts/init-chain.sh [chain_id] [genesis_account]
#
# This script:
#   1. Initializes the chain if not already initialized
#   2. Creates validator and test accounts
#   3. Configures the node for local development
#   4. Starts the virtengine node
#

set -e

# Configuration
CHAIN_ID="${1:-virtengine-localnet-1}"
GENESIS_ACCOUNT="${2:-}"
HOME_DIR="${HOME}/.virtengine"
KEYRING_BACKEND="${KEYRING_BACKEND:-test}"
DENOM="${DENOM:-uve}"
MONIKER="${MONIKER:-localvalidator}"

# Token amounts
VALIDATOR_COINS="100000000000${DENOM}"
VALIDATOR_STAKE="10000000000${DENOM}"
TEST_ACCOUNT_COINS="100000000000${DENOM}"

# Test accounts to create
TEST_ACCOUNTS="alice bob charlie provider operator"

# Deterministic mnemonics for localnet test accounts (DO NOT USE IN PRODUCTION)
# Override with environment variables: VE_MNEMONIC_VALIDATOR, VE_MNEMONIC_ALICE, etc.
declare -A DEFAULT_MNEMONICS
DEFAULT_MNEMONICS[validator]="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
DEFAULT_MNEMONICS[alice]="legal winner thank year wave sausage worth useful legal winner thank yellow"
DEFAULT_MNEMONICS[bob]="letter advice cage absurd amount doctor acoustic avoid letter advice cage above"
DEFAULT_MNEMONICS[charlie]="zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo zoo wrong"
# provider and operator use generated keys (no deterministic mnemonic needed for localnet)

log() {
    echo "[init-chain] $1"
}

error() {
    echo "[init-chain] ERROR: $1" >&2
    exit 1
}

# Wrapper: always pass --home to virtengine commands
# The binary defaults to ~/.akash but we use ~/.virtengine
ve() {
    virtengine --home "${HOME_DIR}" "$@"
}

# Check if chain is already initialized
is_initialized() {
    [ -f "${HOME_DIR}/config/genesis.json" ]
}

# Initialize the chain
init_chain() {
    log "Initializing chain with ID: ${CHAIN_ID}"

    # Clean up previous data
    rm -rf "${HOME_DIR}"

    # Initialize the chain
    ve genesis init "${MONIKER}" --chain-id "${CHAIN_ID}"
}

# Create validator account
create_key_with_mnemonic() {
    local account="$1"
    local mnemonic_var
    local mnemonic

    mnemonic_var="VE_MNEMONIC_${account^^}"
    mnemonic="${!mnemonic_var}"

    if [ -z "${mnemonic}" ]; then
        mnemonic="${DEFAULT_MNEMONICS[$account]:-}"
    fi

    if [ -n "${mnemonic}" ]; then
        printf "%s\n" "${mnemonic}" | ve keys add "${account}" --recover --keyring-backend="${KEYRING_BACKEND}"
    else
        log "  No mnemonic for '${account}', generating new key"
        ve keys add "${account}" --keyring-backend="${KEYRING_BACKEND}"
    fi
}

create_validator() {
    log "Creating validator account..."

    create_key_with_mnemonic validator

    local validator_addr
    validator_addr=$(ve keys show validator -a --keyring-backend="${KEYRING_BACKEND}")
    log "Validator address: ${validator_addr}"

    # Add validator to genesis
    ve genesis add-account "${validator_addr}" "${VALIDATOR_COINS}"

    # Create genesis transaction
    ve genesis gentx validator "${VALIDATOR_STAKE}" \
        --keyring-backend="${KEYRING_BACKEND}" \
        --chain-id="${CHAIN_ID}" \
        --min-self-delegation="1"
}

# Create test accounts
create_test_accounts() {
    log "Creating test accounts..."

    for account in ${TEST_ACCOUNTS}; do
        log "Creating account: ${account}"
        create_key_with_mnemonic "${account}"

        local addr
        addr=$(ve keys show "${account}" -a --keyring-backend="${KEYRING_BACKEND}")
        log "  ${account} address: ${addr}"

        # Add to genesis
        ve genesis add-account "${addr}" "${TEST_ACCOUNT_COINS}"
    done

    # Add genesis account if provided
    if [ -n "${GENESIS_ACCOUNT}" ]; then
        log "Adding genesis account: ${GENESIS_ACCOUNT}"
        ve genesis add-account "${GENESIS_ACCOUNT}" "${TEST_ACCOUNT_COINS}"
    fi
}

# Configure the node for local development
configure_node() {
    log "Configuring node for local development..."

    local config_dir="${HOME_DIR}/config"

    # config.toml modifications
    sed -i.bak 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:26657"#g' "${config_dir}/config.toml"
    sed -i.bak 's/timeout_commit = "5s"/timeout_commit = "1s"/g' "${config_dir}/config.toml"
    sed -i.bak 's/timeout_propose = "3s"/timeout_propose = "1s"/g' "${config_dir}/config.toml"
    sed -i.bak 's/index_all_keys = false/index_all_keys = true/g' "${config_dir}/config.toml"

    # Enable prometheus
    sed -i.bak 's/prometheus = false/prometheus = true/g' "${config_dir}/config.toml"

    # Enable CORS for local development
    sed -i.bak 's/cors_allowed_origins = \[\]/cors_allowed_origins = ["*"]/g' "${config_dir}/config.toml"

    # app.toml modifications - enable API
    if [ -f "${config_dir}/app.toml" ]; then
        # Enable REST API
        sed -i.bak 's/enable = false/enable = true/g' "${config_dir}/app.toml"
        sed -i.bak 's/swagger = false/swagger = true/g' "${config_dir}/app.toml"

        # Enable gRPC
        sed -i.bak 's/address = "0.0.0.0:9090"/address = "0.0.0.0:9090"/g' "${config_dir}/app.toml"

        # Set minimum gas prices
        sed -i.bak "s/minimum-gas-prices = \"\"/minimum-gas-prices = \"0.025${DENOM}\"/g" "${config_dir}/app.toml"
    fi

    # Cleanup backup files
    rm -f "${config_dir}/"*.bak
}

# Patch genesis for local development
# Disables MFA enforcement, identity gating, and other production-only features
# that would prevent genesis transactions from executing
patch_genesis_for_localnet() {
    log "Patching genesis for local development..."

    local genesis="${HOME_DIR}/config/genesis.json"

    # Disable MFA: set require_at_least_one_factor=false and clear sensitive_tx_configs
    jq '.app_state.mfa.params.require_at_least_one_factor = false |
        .app_state.mfa.sensitive_tx_configs = [] |
        .app_state.mktplace.params.enable_mfa_gating = false |
        .app_state.mktplace.params.enable_identity_gating = false |
        .app_state.mktplace.params.mfa_configs = [] |
        .app_state.mktplace.mfa_configs = []' \
        "${genesis}" > "${genesis}.tmp" && mv "${genesis}.tmp" "${genesis}"
}

# Seed test accounts with pre-verified VEID identity records
# Provider needs score >= 70 for registration, Operator >= 85
seed_test_identities() {
    log "Seeding test identities for localnet..."

    local genesis="${HOME_DIR}/config/genesis.json"
    # RFC3339 formatted timestamp for JSON (Go's time.Time format)
    local timestamp
    timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Get account addresses
    local provider_addr operator_addr alice_addr bob_addr charlie_addr
    provider_addr=$(ve keys show provider -a --keyring-backend="${KEYRING_BACKEND}" 2>/dev/null || echo "")
    operator_addr=$(ve keys show operator -a --keyring-backend="${KEYRING_BACKEND}" 2>/dev/null || echo "")
    alice_addr=$(ve keys show alice -a --keyring-backend="${KEYRING_BACKEND}" 2>/dev/null || echo "")
    bob_addr=$(ve keys show bob -a --keyring-backend="${KEYRING_BACKEND}" 2>/dev/null || echo "")
    charlie_addr=$(ve keys show charlie -a --keyring-backend="${KEYRING_BACKEND}" 2>/dev/null || echo "")

    # Build identity records array
    local records="[]"
    
    # Provider: score 80 (meets >=70 requirement)
    if [ -n "${provider_addr}" ]; then
        records=$(echo "${records}" | jq --arg addr "${provider_addr}" --arg ts "${timestamp}" \
            '. += [{
                "account_address": $addr,
                "current_score": 80,
                "tier": "verified",
                "score_version": "localnet-seed-v1.0.0",
                "scope_refs": [],
                "created_at": $ts,
                "updated_at": $ts,
                "flags": [],
                "locked": false,
                "locked_reason": ""
            }]')
        log "  - Provider (${provider_addr}): score=80"
    fi

    # Operator: score 85 (meets validator requirements)
    if [ -n "${operator_addr}" ]; then
        records=$(echo "${records}" | jq --arg addr "${operator_addr}" --arg ts "${timestamp}" \
            '. += [{
                "account_address": $addr,
                "current_score": 85,
                "tier": "trusted",
                "score_version": "localnet-seed-v1.0.0",
                "scope_refs": [],
                "created_at": $ts,
                "updated_at": $ts,
                "flags": [],
                "locked": false,
                "locked_reason": ""
            }]')
        log "  - Operator (${operator_addr}): score=85"
    fi

    # Alice: score 60 (verified tier)
    if [ -n "${alice_addr}" ]; then
        records=$(echo "${records}" | jq --arg addr "${alice_addr}" --arg ts "${timestamp}" \
            '. += [{
                "account_address": $addr,
                "current_score": 60,
                "tier": "verified",
                "score_version": "localnet-seed-v1.0.0",
                "scope_refs": [],
                "created_at": $ts,
                "updated_at": $ts,
                "flags": [],
                "locked": false,
                "locked_reason": ""
            }]')
        log "  - Alice (${alice_addr}): score=60"
    fi

    # Bob: score 50 (standard tier)
    if [ -n "${bob_addr}" ]; then
        records=$(echo "${records}" | jq --arg addr "${bob_addr}" --arg ts "${timestamp}" \
            '. += [{
                "account_address": $addr,
                "current_score": 50,
                "tier": "standard",
                "score_version": "localnet-seed-v1.0.0",
                "scope_refs": [],
                "created_at": $ts,
                "updated_at": $ts,
                "flags": [],
                "locked": false,
                "locked_reason": ""
            }]')
        log "  - Bob (${bob_addr}): score=50"
    fi

    # Charlie: score 20 (basic tier, below provider threshold)
    if [ -n "${charlie_addr}" ]; then
        records=$(echo "${records}" | jq --arg addr "${charlie_addr}" --arg ts "${timestamp}" \
            '. += [{
                "account_address": $addr,
                "current_score": 20,
                "tier": "basic",
                "score_version": "localnet-seed-v1.0.0",
                "scope_refs": [],
                "created_at": $ts,
                "updated_at": $ts,
                "flags": [],
                "locked": false,
                "locked_reason": ""
            }]')
        log "  - Charlie (${charlie_addr}): score=20"
    fi

    # Update genesis with identity records
    jq --argjson records "${records}" \
        '.app_state.veid.identity_records = $records' \
        "${genesis}" > "${genesis}.tmp" && mv "${genesis}.tmp" "${genesis}"

    log "Identity seeding complete!"
}

# Collect genesis transactions
collect_gentxs() {
    log "Collecting genesis transactions..."
    ve genesis collect
}

# Export test account info
export_account_info() {
    log "Exporting account information..."

    local accounts_file="${HOME_DIR}/test-accounts.json"

    {
        echo "{"
        echo '  "accounts": {'

        local first=true
        local addr
        for account in validator ${TEST_ACCOUNTS}; do
            addr=$(ve keys show "${account}" -a --keyring-backend="${KEYRING_BACKEND}")

            if [ "${first}" = "true" ]; then
                first=false
            else
                echo ","
            fi

            printf '    "%s": "%s"' "${account}" "${addr}"
        done

        echo ""
        echo "  },"
        echo "  \"chain_id\": \"${CHAIN_ID}\","
        echo "  \"denom\": \"${DENOM}\""
        echo "}"
    } > "${accounts_file}"

    log "Account info exported to: ${accounts_file}"
    cat "${accounts_file}"
}

# Start the node
start_node() {
    log "Starting VirtEngine node..."
    log "Chain ID: ${CHAIN_ID}"
    log "RPC: http://0.0.0.0:26657"
    log "gRPC: 0.0.0.0:9090"
    log "REST API: http://0.0.0.0:1317"

    exec virtengine --home "${HOME_DIR}" start \
        --pruning=nothing \
        --minimum-gas-prices="0.025${DENOM}" \
        --api.enable=true \
        --api.swagger=true \
        --api.address="tcp://0.0.0.0:1317" \
        --grpc.address="0.0.0.0:9090"
}

# Main
main() {
    log "VirtEngine Chain Initialization"
    log "================================"

    if is_initialized; then
        log "Chain already initialized, starting node..."
    else
        init_chain
        create_validator
        create_test_accounts
        patch_genesis_for_localnet
        seed_test_identities
        collect_gentxs
        configure_node
        export_account_info
        log "Chain initialization complete!"
    fi

    start_node
}

main
