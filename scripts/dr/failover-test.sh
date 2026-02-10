#!/usr/bin/env bash
# scripts/dr/failover-test.sh
# Docker Compose-based DR failover validation.
set -euo pipefail

CHAIN_ID="${CHAIN_ID:-virtengine-dr-test-1}"
WORK_DIR="${DR_WORK_DIR:-.dr-failover}"
COMPOSE_FILE="scripts/dr/docker-compose.failover.yaml"

log() { echo "[dr-failover] $*"; }
fail() { echo "[dr-failover] ERROR: $*" >&2; exit 1; }

if ! command -v jq >/dev/null 2>&1; then
  fail "jq not found in PATH"
fi

abs_path() {
  local target="$1"
  if command -v python >/dev/null 2>&1; then
    python - <<'PY' "$target"
import os, sys
print(os.path.abspath(sys.argv[1]))
PY
  else
    pwsh -NoProfile -Command "Resolve-Path '$target' | Select-Object -ExpandProperty Path"
  fi
}

wait_for_rpc() {
  local url="$1"
  local label="$2"
  for i in $(seq 1 60); do
    if curl -sf --connect-timeout 5 "${url}/status" >/dev/null 2>&1; then
      log "${label} RPC is ready"
      return 0
    fi
    sleep 2
  done
  fail "${label} RPC did not become ready"
}

rpc_height() {
  curl -sf --connect-timeout 5 "$1/status" | jq -r '.result.sync_info.latest_block_height' 2>/dev/null || echo "0"
}

rpc_app_hash() {
  curl -sf --connect-timeout 5 "$1/status" | jq -r '.result.sync_info.latest_app_hash' 2>/dev/null || echo ""
}

cleanup() {
  log "Cleaning up docker compose services..."
  DR_WORK_DIR="$DR_WORK_DIR" docker compose -f "$COMPOSE_FILE" down -v --remove-orphans >/dev/null 2>&1 || true
}

trap cleanup EXIT

main() {
  export DR_WORK_DIR
  DR_WORK_DIR=$(abs_path "$WORK_DIR")
  export DR_WORK_DIR

  log "Initializing failover cluster..."
  WORK_DIR="$DR_WORK_DIR" CHAIN_ID="$CHAIN_ID" scripts/dr/init-failover-cluster.sh

  log "Building virtengine image..."
  docker build -t virtengine:dr-test -f _build/Dockerfile.virtengine .

  log "Starting failover cluster..."
  DR_WORK_DIR="$DR_WORK_DIR" docker compose -f "$COMPOSE_FILE" up -d --remove-orphans

  wait_for_rpc "http://localhost:26657" "Primary"
  wait_for_rpc "http://localhost:26667" "Secondary"
  wait_for_rpc "http://localhost:26677" "Witness"

  log "Capturing baseline heights..."
  base_height=$(rpc_height "http://localhost:26667")

  log "Simulating primary failure..."
  DR_WORK_DIR="$DR_WORK_DIR" docker compose -f "$COMPOSE_FILE" stop virtengine-dr-primary

  log "Verifying secondary continues consensus..."
  for i in $(seq 1 15); do
    sleep 2
    new_height=$(rpc_height "http://localhost:26667")
    if [ "$new_height" -gt "$base_height" ]; then
      break
    fi
  done
  if [ "$new_height" -le "$base_height" ]; then
    fail "Secondary did not advance blocks after primary failure"
  fi

  log "Verifying state consistency between secondary and witness..."
  secondary_height=$(rpc_height "http://localhost:26667")
  witness_height=$(rpc_height "http://localhost:26677")
  secondary_hash=$(rpc_app_hash "http://localhost:26667")
  witness_hash=$(rpc_app_hash "http://localhost:26677")

  height_diff=$((secondary_height - witness_height))
  if [ "$height_diff" -gt 1 ] || [ "$height_diff" -lt -1 ]; then
    fail "Height divergence too large (secondary=${secondary_height}, witness=${witness_height})"
  fi
  if [ -n "$secondary_hash" ] && [ "$secondary_hash" != "$witness_hash" ]; then
    fail "App hash mismatch between secondary and witness"
  fi

  log "Exporting genesis from secondary node..."
  docker exec virtengine-dr-secondary virtengine export --home /var/lib/virtengine > "${DR_WORK_DIR}/export.json"

  log "Preparing restore node from exported genesis..."
  restore_home="${DR_WORK_DIR}/restore"
  rm -rf "$restore_home"
  virtengine --home "$restore_home" genesis init restore --chain-id "$CHAIN_ID" >/dev/null
  cp "${DR_WORK_DIR}/export.json" "${restore_home}/config/genesis.json"

  restore_config="${restore_home}/config/config.toml"
  restore_app="${restore_home}/config/app.toml"
  secondary_id=$(virtengine --home "${DR_WORK_DIR}/secondary" tendermint show-node-id)
  witness_id=$(virtengine --home "${DR_WORK_DIR}/witness" tendermint show-node-id)
  restore_peers="${secondary_id}@virtengine-dr-secondary:26666,${witness_id}@virtengine-dr-witness:26676"
  sed -i.bak "s#laddr = \"tcp://127.0.0.1:26657\"#laddr = \"tcp://0.0.0.0:26687\"#" "$restore_config"
  sed -i.bak "s#laddr = \"tcp://0.0.0.0:26656\"#laddr = \"tcp://0.0.0.0:26686\"#" "$restore_config"
  sed -i.bak "s#proxy_app = \"tcp://127.0.0.1:26658\"#proxy_app = \"tcp://0.0.0.0:26688\"#" "$restore_config"
  sed -i.bak "s#persistent_peers = \"\"#persistent_peers = \"${restore_peers}\"#" "$restore_config"

  if [ -f "$restore_app" ]; then
    sed -i.bak "s#address = \"tcp://0.0.0.0:1317\"#address = \"tcp://0.0.0.0:1327\"#" "$restore_app"
    sed -i.bak "s#address = \"0.0.0.0:9090\"#address = \"0.0.0.0:9100\"#" "$restore_app"
  fi
  rm -f "${restore_home}/config/"*.bak

  log "Starting restore node..."
  DR_WORK_DIR="$DR_WORK_DIR" docker compose -f "$COMPOSE_FILE" --profile restore up -d virtengine-dr-restore
  wait_for_rpc "http://localhost:26687" "Restore"
  DR_WORK_DIR="$DR_WORK_DIR" docker compose -f "$COMPOSE_FILE" --profile restore stop virtengine-dr-restore

  log "Testing backup snapshot restoration..."
  docker exec virtengine-dr-secondary tar -czf /var/lib/virtengine/latest-snapshot.tar.gz -C /var/lib/virtengine data
  docker cp virtengine-dr-secondary:/var/lib/virtengine/latest-snapshot.tar.gz "${DR_WORK_DIR}/latest-snapshot.tar.gz"

  rm -rf "${restore_home}/data"
  tar -xzf "${DR_WORK_DIR}/latest-snapshot.tar.gz" -C "${restore_home}"

  DR_WORK_DIR="$DR_WORK_DIR" docker compose -f "$COMPOSE_FILE" --profile restore up -d virtengine-dr-restore
  wait_for_rpc "http://localhost:26687" "Restore-from-snapshot"
  DR_WORK_DIR="$DR_WORK_DIR" docker compose -f "$COMPOSE_FILE" --profile restore stop virtengine-dr-restore

  log "DR failover test completed successfully."
}

main "$@"
