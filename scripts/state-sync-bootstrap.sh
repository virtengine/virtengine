#!/bin/bash
# SCALE-002: State Sync Bootstrap Script
# Enables fast node bootstrap using state sync for horizontal scaling
#
# Usage: ./state-sync-bootstrap.sh [options]
#   --rpc-servers    Comma-separated list of RPC endpoints
#   --trust-height   Specific trust height (optional, auto-detected if not provided)
#   --trust-hash     Specific trust hash (optional, auto-detected if not provided)
#   --home           Node home directory (default: ~/.virtengine)
#   --chain-id       Chain ID (default: virtengine-1)
#   --dry-run        Show configuration without applying

set -euo pipefail

# Default configuration
NODE_HOME="${NODE_HOME:-$HOME/.virtengine}"
CHAIN_ID="${CHAIN_ID:-virtengine-1}"
RPC_SERVERS="${RPC_SERVERS:-}"
TRUST_HEIGHT="${TRUST_HEIGHT:-}"
TRUST_HASH="${TRUST_HASH:-}"
DRY_RUN=false
BLOCKS_BEHIND=1000  # Trust a block this far behind latest

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

usage() {
    cat << EOF
VirtEngine State Sync Bootstrap Script

Enables fast node synchronization using CometBFT state sync.
New validators can sync in ~20 minutes instead of hours/days.

Usage: $0 [OPTIONS]

Options:
    --rpc-servers    Comma-separated RPC endpoints (required)
                     Example: https://rpc1.virtengine.network:443,https://rpc2.virtengine.network:443
    --trust-height   Specific block height to trust (optional, auto-detected)
    --trust-hash     Block hash at trust height (optional, auto-detected)
    --home           Node home directory (default: ~/.virtengine)
    --chain-id       Chain ID (default: virtengine-1)
    --dry-run        Preview configuration without applying
    -h, --help       Show this help message

Examples:
    # Auto-detect trust height from RPC
    $0 --rpc-servers https://rpc1.virtengine.network:443,https://rpc2.virtengine.network:443

    # Specify trust height and hash manually
    $0 --rpc-servers https://rpc.virtengine.network:443 \\
       --trust-height 5000000 \\
       --trust-hash ABCD1234...

    # Dry run to see configuration
    $0 --rpc-servers https://rpc.virtengine.network:443 --dry-run

Environment Variables:
    NODE_HOME       Node home directory
    CHAIN_ID        Chain ID
    RPC_SERVERS     RPC endpoints
    TRUST_HEIGHT    Trust height
    TRUST_HASH      Trust hash

EOF
    exit 0
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --rpc-servers)
            RPC_SERVERS="$2"
            shift 2
            ;;
        --trust-height)
            TRUST_HEIGHT="$2"
            shift 2
            ;;
        --trust-hash)
            TRUST_HASH="$2"
            shift 2
            ;;
        --home)
            NODE_HOME="$2"
            shift 2
            ;;
        --chain-id)
            CHAIN_ID="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            log_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Validate required parameters
if [[ -z "$RPC_SERVERS" ]]; then
    log_error "RPC servers are required. Use --rpc-servers or set RPC_SERVERS environment variable."
    exit 1
fi

# Get the first RPC server for queries
FIRST_RPC="${RPC_SERVERS%%,*}"

log_info "State Sync Bootstrap for VirtEngine"
log_info "=================================="
log_info "Node Home: $NODE_HOME"
log_info "Chain ID: $CHAIN_ID"
log_info "RPC Servers: $RPC_SERVERS"

# Check if node home exists
if [[ ! -d "$NODE_HOME/config" ]]; then
    log_error "Node home directory not found: $NODE_HOME/config"
    log_error "Please initialize the node first with: virtengine init <moniker> --chain-id $CHAIN_ID"
    exit 1
fi

# Function to fetch from RPC
fetch_rpc() {
    local endpoint="$1"
    local url="${FIRST_RPC}${endpoint}"
    
    # Handle both http and https
    if command -v curl &> /dev/null; then
        curl -s --fail --max-time 10 "$url"
    elif command -v wget &> /dev/null; then
        wget -q -O - --timeout=10 "$url"
    else
        log_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
}

# Fetch latest block height
fetch_latest_height() {
    log_info "Fetching latest block height from $FIRST_RPC..."
    
    local response
    response=$(fetch_rpc "/block") || {
        log_error "Failed to fetch latest block from RPC"
        exit 1
    }
    
    local height
    height=$(echo "$response" | jq -r '.result.block.header.height // .block.header.height' 2>/dev/null)
    
    if [[ -z "$height" || "$height" == "null" ]]; then
        log_error "Failed to parse block height from response"
        exit 1
    fi
    
    echo "$height"
}

# Fetch block hash at specific height
fetch_block_hash() {
    local height="$1"
    log_info "Fetching block hash at height $height..."
    
    local response
    response=$(fetch_rpc "/block?height=$height") || {
        log_error "Failed to fetch block at height $height"
        exit 1
    }
    
    local hash
    hash=$(echo "$response" | jq -r '.result.block_id.hash // .block_id.hash' 2>/dev/null)
    
    if [[ -z "$hash" || "$hash" == "null" ]]; then
        log_error "Failed to parse block hash from response"
        exit 1
    fi
    
    echo "$hash"
}

# Verify RPC connectivity
verify_rpc() {
    log_info "Verifying RPC connectivity..."
    
    for rpc in ${RPC_SERVERS//,/ }; do
        log_info "  Checking $rpc..."
        if curl -s --fail --max-time 5 "$rpc/health" > /dev/null 2>&1; then
            log_success "  $rpc is healthy"
        else
            log_warn "  $rpc may be unavailable"
        fi
    done
}

# Auto-detect trust height and hash if not provided
if [[ -z "$TRUST_HEIGHT" || -z "$TRUST_HASH" ]]; then
    log_info "Auto-detecting trust height and hash..."
    
    verify_rpc
    
    LATEST_HEIGHT=$(fetch_latest_height)
    log_info "Latest block height: $LATEST_HEIGHT"
    
    if [[ -z "$TRUST_HEIGHT" ]]; then
        # Trust a block ~1000 blocks behind latest for safety
        TRUST_HEIGHT=$((LATEST_HEIGHT - BLOCKS_BEHIND))
        log_info "Using trust height: $TRUST_HEIGHT (${BLOCKS_BEHIND} blocks behind latest)"
    fi
    
    if [[ -z "$TRUST_HASH" ]]; then
        TRUST_HASH=$(fetch_block_hash "$TRUST_HEIGHT")
        log_info "Trust hash: $TRUST_HASH"
    fi
fi

# Display configuration
log_info ""
log_info "State Sync Configuration"
log_info "========================"
log_info "RPC Servers:  $RPC_SERVERS"
log_info "Trust Height: $TRUST_HEIGHT"
log_info "Trust Hash:   $TRUST_HASH"
log_info ""

if [[ "$DRY_RUN" == "true" ]]; then
    log_info "Dry run mode - configuration preview only"
    log_info ""
    log_info "Add to config.toml [statesync] section:"
    cat << EOF

[statesync]
enable = true
rpc_servers = "$RPC_SERVERS"
trust_height = $TRUST_HEIGHT
trust_hash = "$TRUST_HASH"
trust_period = "168h"
discovery_time = "15s"
chunk_fetchers = "4"
chunk_request_timeout = "10s"

EOF
    exit 0
fi

# Backup existing config
CONFIG_FILE="$NODE_HOME/config/config.toml"
BACKUP_FILE="$CONFIG_FILE.backup.$(date +%Y%m%d_%H%M%S)"

log_info "Backing up config to $BACKUP_FILE..."
cp "$CONFIG_FILE" "$BACKUP_FILE"

# Update config.toml with state sync settings
log_info "Updating config.toml with state sync settings..."

# Use sed to update the statesync section
# This handles both cases where settings exist or need to be added

# Enable state sync
sed -i.tmp 's/^enable = false$/enable = true/' "$CONFIG_FILE"

# Update or add RPC servers
if grep -q "^rpc_servers = " "$CONFIG_FILE"; then
    sed -i.tmp "s|^rpc_servers = .*|rpc_servers = \"$RPC_SERVERS\"|" "$CONFIG_FILE"
else
    sed -i.tmp "/^\[statesync\]/a rpc_servers = \"$RPC_SERVERS\"" "$CONFIG_FILE"
fi

# Update or add trust height
if grep -q "^trust_height = " "$CONFIG_FILE"; then
    sed -i.tmp "s/^trust_height = .*/trust_height = $TRUST_HEIGHT/" "$CONFIG_FILE"
else
    sed -i.tmp "/^\[statesync\]/a trust_height = $TRUST_HEIGHT" "$CONFIG_FILE"
fi

# Update or add trust hash
if grep -q "^trust_hash = " "$CONFIG_FILE"; then
    sed -i.tmp "s/^trust_hash = .*/trust_hash = \"$TRUST_HASH\"/" "$CONFIG_FILE"
else
    sed -i.tmp "/^\[statesync\]/a trust_hash = \"$TRUST_HASH\"" "$CONFIG_FILE"
fi

# Update trust period
if grep -q "^trust_period = " "$CONFIG_FILE"; then
    sed -i.tmp 's/^trust_period = .*/trust_period = "168h"/' "$CONFIG_FILE"
fi

# Update chunk fetchers for faster sync
if grep -q "^chunk_fetchers = " "$CONFIG_FILE"; then
    sed -i.tmp 's/^chunk_fetchers = .*/chunk_fetchers = "4"/' "$CONFIG_FILE"
fi

# Cleanup temp files
rm -f "$CONFIG_FILE.tmp"

log_success "Config updated successfully"

# Ask about resetting chain data
log_info ""
log_warn "State sync requires clearing existing chain data."
log_warn "This will NOT delete your validator keys or node key."

read -p "Reset chain data now? (y/N) " -n 1 -r
echo

if [[ $REPLY =~ ^[Yy]$ ]]; then
    log_info "Resetting chain data (keeping keys and address book)..."
    
    if command -v virtengine &> /dev/null; then
        virtengine tendermint unsafe-reset-all --home "$NODE_HOME" --keep-addr-book
    else
        log_warn "virtengine command not found, manually clearing data..."
        rm -rf "$NODE_HOME/data/"
        mkdir -p "$NODE_HOME/data"
        log_info "Data directory cleared"
    fi
    
    log_success "Chain data reset complete"
fi

log_info ""
log_success "State sync configuration complete!"
log_info ""
log_info "Next steps:"
log_info "  1. Start the node: virtengine start --home $NODE_HOME"
log_info "  2. Monitor sync progress in logs"
log_info "  3. State sync typically takes 15-30 minutes"
log_info ""
log_info "Monitor sync status:"
log_info "  virtengine status 2>&1 | jq '.sync_info'"
log_info ""
log_info "Expected stages:"
log_info "  1. Discovering snapshots from peers"
log_info "  2. Downloading snapshot chunks"
log_info "  3. Applying snapshot to state"
log_info "  4. Catching up remaining blocks"
log_info "  5. Sync complete (catching_up: false)"
