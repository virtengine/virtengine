#!/usr/bin/env bash
# Initialize a 3-node failover cluster for DR testing.
set -euo pipefail

CHAIN_ID="${CHAIN_ID:-virtengine-dr-test-1}"
DENOM="${DENOM:-uve}"
WORK_DIR="${WORK_DIR:-.dr-failover}"
FIXTURE_FILE="${FIXTURE_FILE:-scripts/dr/fixtures/dr-test-accounts.json}"

NODES=("primary" "secondary" "witness")
P2P_PORTS=(26656 26666 26676)
RPC_PORTS=(26657 26667 26677)
API_PORTS=(1317 1319 1321)
GRPC_PORTS=(9090 9092 9094)

log() { echo "[dr-init] $*"; }
fail() { echo "[dr-init] ERROR: $*" >&2; exit 1; }

if ! command -v virtengine >/dev/null 2>&1; then
  fail "virtengine CLI not found in PATH"
fi

if ! command -v jq >/dev/null 2>&1; then
  fail "jq not found in PATH"
fi

ve() {
  local home="$1"
  shift
  virtengine --home "$home" "$@"
}

reset_workdir() {
  rm -rf "${WORK_DIR}"
  mkdir -p "${WORK_DIR}"
}

init_node() {
  local name="$1"
  local home="${WORK_DIR}/${name}"
  rm -rf "$home"
  ve "$home" genesis init "$name" --chain-id "$CHAIN_ID" >/dev/null
}

configure_node() {
  local name="$1"
  local home="${WORK_DIR}/${name}"
  local p2p_port="$2"
  local rpc_port="$3"
  local api_port="$4"
  local grpc_port="$5"

  local config="${home}/config/config.toml"
  local app="${home}/config/app.toml"

  sed -i.bak "s#laddr = \"tcp://127.0.0.1:26657\"#laddr = \"tcp://0.0.0.0:${rpc_port}\"#" "$config"
  sed -i.bak "s#laddr = \"tcp://0.0.0.0:26656\"#laddr = \"tcp://0.0.0.0:${p2p_port}\"#" "$config"
  sed -i.bak "s#proxy_app = \"tcp://127.0.0.1:26658\"#proxy_app = \"tcp://0.0.0.0:$((rpc_port + 1))\"#" "$config"
  sed -i.bak "s#allow_duplicate_ip = false#allow_duplicate_ip = true#" "$config"

  if [ -f "$app" ]; then
    sed -i.bak "s#address = \"tcp://0.0.0.0:1317\"#address = \"tcp://0.0.0.0:${api_port}\"#" "$app"
    sed -i.bak "s#address = \"0.0.0.0:9090\"#address = \"0.0.0.0:${grpc_port}\"#" "$app"
    sed -i.bak "s#minimum-gas-prices = \"\"#minimum-gas-prices = \"0.025${DENOM}\"#" "$app"
    sed -i.bak "s#enable = false#enable = true#g" "$app"
  fi

  rm -f "${home}/config/"*.bak
}

create_validator_key() {
  local name="$1"
  local home="${WORK_DIR}/${name}"
  local mnemonic="$2"
  printf "%s\n" "$mnemonic" | ve "$home" keys add validator --recover --keyring-backend test >/dev/null
}

validator_address() {
  local name="$1"
  local home="${WORK_DIR}/${name}"
  ve "$home" keys show validator -a --keyring-backend test
}

gentx_for_node() {
  local name="$1"
  local home="${WORK_DIR}/${name}"
  local gentx_dir="${WORK_DIR}/gentx"
  mkdir -p "$gentx_dir"
  ve "$home" genesis gentx validator "10000000000${DENOM}" \
    --chain-id "$CHAIN_ID" \
    --keyring-backend test \
    --output-document "${gentx_dir}/${name}.json" >/dev/null
}

seed_fixture_accounts() {
  local home="${WORK_DIR}/primary"

  if [ ! -f "$FIXTURE_FILE" ]; then
    fail "Fixture file not found: ${FIXTURE_FILE}"
  fi

  local denom
  denom=$(jq -r '.denom' "$FIXTURE_FILE")
  [ -n "$denom" ] || denom="$DENOM"

  jq -c '.accounts[]' "$FIXTURE_FILE" | while read -r entry; do
    local name mnemonic balance addr
    name=$(echo "$entry" | jq -r '.name')
    mnemonic=$(echo "$entry" | jq -r '.mnemonic')
    balance=$(echo "$entry" | jq -r '.balance')
    printf "%s\n" "$mnemonic" | ve "$home" keys add "$name" --recover --keyring-backend test >/dev/null
    addr=$(ve "$home" keys show "$name" -a --keyring-backend test)
    ve "$home" genesis add-account "$addr" "${balance}${denom}" >/dev/null
  done
}

set_persistent_peers() {
  local peers=""
  local primary_id secondary_id witness_id
  primary_id=$(ve "${WORK_DIR}/primary" tendermint show-node-id)
  secondary_id=$(ve "${WORK_DIR}/secondary" tendermint show-node-id)
  witness_id=$(ve "${WORK_DIR}/witness" tendermint show-node-id)

  peers="${primary_id}@virtengine-dr-primary:26656,${secondary_id}@virtengine-dr-secondary:26666,${witness_id}@virtengine-dr-witness:26676"

  for name in "${NODES[@]}"; do
    local config="${WORK_DIR}/${name}/config/config.toml"
    sed -i.bak "s#persistent_peers = \"\"#persistent_peers = \"${peers}\"#" "$config"
    rm -f "${WORK_DIR}/${name}/config/"*.bak
  done
}

main() {
  reset_workdir
  log "Initializing failover cluster in ${WORK_DIR}"

  for idx in "${!NODES[@]}"; do
    init_node "${NODES[$idx]}"
    configure_node "${NODES[$idx]}" "${P2P_PORTS[$idx]}" "${RPC_PORTS[$idx]}" "${API_PORTS[$idx]}" "${GRPC_PORTS[$idx]}"
  done

  create_validator_key primary "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
  create_validator_key secondary "legal winner thank year wave sausage worth useful legal winner thank yellow"
  create_validator_key witness "letter advice cage absurd amount doctor acoustic avoid letter advice cage above"

  local primary_home="${WORK_DIR}/primary"
  for name in "${NODES[@]}"; do
    addr=$(validator_address "$name")
    ve "$primary_home" genesis add-account "$addr" "100000000000${DENOM}" >/dev/null
    gentx_for_node "$name"
  done

  seed_fixture_accounts

  mkdir -p "${primary_home}/config/gentx"
  cp "${WORK_DIR}/gentx/"*.json "${primary_home}/config/gentx/"
  ve "$primary_home" genesis collect-gentxs >/dev/null

  for name in "secondary" "witness"; do
    cp "${primary_home}/config/genesis.json" "${WORK_DIR}/${name}/config/genesis.json"
  done

  set_persistent_peers
  log "Failover cluster initialization complete."
}

main "$@"
