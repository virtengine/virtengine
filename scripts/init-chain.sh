#!/bin/sh
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
DENOM="${DENOM:-uakt}"
MONIKER="${MONIKER:-localvalidator}"

# Token amounts
VALIDATOR_COINS="100000000000${DENOM}"
VALIDATOR_STAKE="10000000000${DENOM}"
TEST_ACCOUNT_COINS="100000000000${DENOM}"

# Test accounts to create
TEST_ACCOUNTS="alice bob charlie provider operator"

log() {
    echo "[init-chain] $1"
}

error() {
    echo "[init-chain] ERROR: $1" >&2
    exit 1
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
    virtengine genesis init "${MONIKER}" --chain-id "${CHAIN_ID}"
}

# Create validator account
create_validator() {
    log "Creating validator account..."

    virtengine keys add validator --keyring-backend="${KEYRING_BACKEND}"

    local validator_addr=$(virtengine keys show validator -a --keyring-backend="${KEYRING_BACKEND}")
    log "Validator address: ${validator_addr}"

    # Add validator to genesis
    virtengine genesis add-account "${validator_addr}" ${VALIDATOR_COINS}

    # Create genesis transaction
    virtengine genesis gentx validator ${VALIDATOR_STAKE} \
        --keyring-backend="${KEYRING_BACKEND}" \
        --chain-id="${CHAIN_ID}" \
        --min-self-delegation="1"
}

# Create test accounts
create_test_accounts() {
    log "Creating test accounts..."

    for account in ${TEST_ACCOUNTS}; do
        log "Creating account: ${account}"
        virtengine keys add "${account}" --keyring-backend="${KEYRING_BACKEND}"

        local addr=$(virtengine keys show "${account}" -a --keyring-backend="${KEYRING_BACKEND}")
        log "  ${account} address: ${addr}"

        # Add to genesis
        virtengine genesis add-account "${addr}" ${TEST_ACCOUNT_COINS}
    done

    # Add genesis account if provided
    if [ -n "${GENESIS_ACCOUNT}" ]; then
        log "Adding genesis account: ${GENESIS_ACCOUNT}"
        virtengine genesis add-account "${GENESIS_ACCOUNT}" ${TEST_ACCOUNT_COINS}
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
        sed -i.bak 's/minimum-gas-prices = ""/minimum-gas-prices = "0.025'${DENOM}'"/g' "${config_dir}/app.toml"
    fi

    # Cleanup backup files
    rm -f "${config_dir}/"*.bak
}

# Collect genesis transactions
collect_gentxs() {
    log "Collecting genesis transactions..."
    virtengine genesis collect
}

# Export test account info
export_account_info() {
    log "Exporting account information..."

    local accounts_file="${HOME_DIR}/test-accounts.json"

    echo "{" > "${accounts_file}"
    echo '  "accounts": {' >> "${accounts_file}"

    local first=true
    for account in validator ${TEST_ACCOUNTS}; do
        local addr=$(virtengine keys show "${account}" -a --keyring-backend="${KEYRING_BACKEND}")

        if [ "${first}" = "true" ]; then
            first=false
        else
            echo "," >> "${accounts_file}"
        fi

        printf '    "%s": "%s"' "${account}" "${addr}" >> "${accounts_file}"
    done

    echo "" >> "${accounts_file}"
    echo "  }," >> "${accounts_file}"
    echo '  "chain_id": "'${CHAIN_ID}'",' >> "${accounts_file}"
    echo '  "denom": "'${DENOM}'"' >> "${accounts_file}"
    echo "}" >> "${accounts_file}"

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

    exec virtengine start \
        --pruning=nothing \
        --minimum-gas-prices="0.025${DENOM}" \
        --api.enable=true \
        --api.swagger=true \
        --api.address="tcp://0.0.0.0:1317" \
        --grpc.address="0.0.0.0:9090" \
        --grpc-web.address="0.0.0.0:9091"
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
        collect_gentxs
        configure_node
        export_account_info
        log "Chain initialization complete!"
    fi

    start_node
}

main
