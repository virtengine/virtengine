#!/usr/bin/env bash
# Post-deploy smoke test for staging/testnet environments.
set -euo pipefail

RPC_URL="${VE_RPC_URL:-}"
GRPC_URL="${VE_GRPC_URL:-}"
REST_URL="${VE_REST_URL:-}"
CHAIN_ID="${VE_CHAIN_ID:-}"
DENOM="${VE_SMOKE_DENOM:-uve}"
MNEMONIC="${VE_SMOKE_MNEMONIC:-}"
KEY_NAME="${VE_SMOKE_KEY_NAME:-smoke-test}"
RECIPIENT="${VE_SMOKE_RECIPIENT_ADDRESS:-}"
PROVIDER_ADDRESS="${VE_PROVIDER_ADDRESS:-}"

log() { echo "[smoke-test] $*"; }
fail() { echo "[smoke-test] ERROR: $*" >&2; exit 1; }

require_env() {
  local name="$1"
  local value="${!name:-}"
  if [ -z "$value" ]; then
    fail "Missing required env: ${name}"
  fi
}

require_env RPC_URL
require_env GRPC_URL
require_env REST_URL
require_env CHAIN_ID
require_env MNEMONIC
require_env RECIPIENT

if ! command -v virtengine >/dev/null 2>&1; then
  fail "virtengine CLI not found in PATH"
fi

if ! command -v jq >/dev/null 2>&1; then
  fail "jq not found in PATH"
fi

if ! command -v grpcurl >/dev/null 2>&1; then
  fail "grpcurl not found in PATH"
fi

log "Checking RPC endpoint..."
rpc_status=$(curl -sf --connect-timeout 10 "${RPC_URL}/status")
rpc_height=$(echo "$rpc_status" | grep -o '"latest_block_height":"[0-9]*"' | grep -o '[0-9]*' | head -1)
[ -n "$rpc_height" ] || fail "RPC status did not include block height"
log "RPC height: ${rpc_height}"

log "Checking REST API..."
curl -sf --connect-timeout 10 "${REST_URL}/cosmos/base/tendermint/v1beta1/node_info" >/dev/null
curl -sf --connect-timeout 10 "${REST_URL}/cosmos/base/tendermint/v1beta1/blocks/latest" >/dev/null

log "Checking gRPC endpoint..."
grpcurl -plaintext "${GRPC_URL}" list | grep -q "cosmos.base.tendermint.v1beta1.Service" || \
  fail "gRPC reflection did not list tendermint service"
grpcurl -plaintext -d '{}' "${GRPC_URL}" cosmos.base.tendermint.v1beta1.Service/GetLatestBlock >/dev/null

log "Running module queries..."
keyring_dir=$(mktemp -d)
trap 'rm -rf "$keyring_dir"' EXIT

printf "%s\n" "$MNEMONIC" | virtengine keys add "$KEY_NAME" \
  --recover \
  --keyring-backend test \
  --keyring-dir "$keyring_dir" \
  --output json > "${keyring_dir}/key.json"

from_address=$(jq -r '.address' "${keyring_dir}/key.json")
[ -n "$from_address" ] || fail "Failed to derive address from mnemonic"

curl -sf --connect-timeout 10 "${REST_URL}/virtengine/veid/v1/scope/list?account_address=${from_address}&pagination.limit=1" >/dev/null
curl -sf --connect-timeout 10 "${REST_URL}/virtengine/mfa/v1/factors/${from_address}" >/dev/null
curl -sf --connect-timeout 10 "${REST_URL}/virtengine/market/v2beta1/orders/list?filters.owner=${from_address}&pagination.limit=1" >/dev/null

settlement_service=$(grpcurl -plaintext "${GRPC_URL}" list | grep -i "settlement" | head -1 || true)
if [ -z "$settlement_service" ]; then
  fail "gRPC reflection does not expose settlement service"
fi
grpcurl -plaintext "${GRPC_URL}" list "${settlement_service}" >/dev/null

log "Submitting MsgSend transaction..."
tx_output=$(virtengine tx bank send "${from_address}" "${RECIPIENT}" "1000${DENOM}" \
  --node "${RPC_URL}" \
  --chain-id "${CHAIN_ID}" \
  --keyring-backend test \
  --keyring-dir "$keyring_dir" \
  --gas auto \
  --yes \
  --broadcast-mode block \
  --output json)

tx_code=$(echo "$tx_output" | jq -r '.code // 0')
if [ "$tx_code" != "0" ]; then
  fail "MsgSend failed: ${tx_output}"
fi

virtengine query bank balances "${RECIPIENT}" \
  --node "${RPC_URL}" \
  --output json | jq -e '.balances | length >= 0' >/dev/null

log "Checking provider registration..."
providers_payload=$(curl -sf --connect-timeout 10 "${REST_URL}/virtengine/provider/v1beta4/providers")
echo "$providers_payload" | grep -q "providers" || fail "Provider list response missing providers field"
if [ -n "$PROVIDER_ADDRESS" ]; then
  echo "$providers_payload" | grep -q "$PROVIDER_ADDRESS" || fail "Provider ${PROVIDER_ADDRESS} not found in provider list"
fi

log "Smoke test completed successfully."
