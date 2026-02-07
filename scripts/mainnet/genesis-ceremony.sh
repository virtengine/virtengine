#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/mainnet/genesis-ceremony.sh \
  --gentx-dir <dir> \
  [--home <path>] \
  [--params <path>] \
  [--allocations <path>] \
  [--checks <path>] \
  [--output <dir>] \
  [--chain-id <id>] \
  [--genesis-time <rfc3339>]

Builds the mainnet genesis.json deterministically from config inputs.
USAGE
}

GENTX_DIR=""
HOME_DIR=".cache/mainnet-genesis"
PARAMS="config/mainnet/genesis-params.json"
ALLOCATIONS="config/mainnet/genesis-allocations.json"
CHECKS="config/mainnet/genesis-checks.json"
OUTPUT_DIR="artifacts/mainnet"
CHAIN_ID_OVERRIDE=""
GENESIS_TIME_OVERRIDE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --gentx-dir)
      GENTX_DIR="$2"
      shift 2
      ;;
    --home)
      HOME_DIR="$2"
      shift 2
      ;;
    --params)
      PARAMS="$2"
      shift 2
      ;;
    --allocations)
      ALLOCATIONS="$2"
      shift 2
      ;;
    --checks)
      CHECKS="$2"
      shift 2
      ;;
    --output)
      OUTPUT_DIR="$2"
      shift 2
      ;;
    --chain-id)
      CHAIN_ID_OVERRIDE="$2"
      shift 2
      ;;
    --genesis-time)
      GENESIS_TIME_OVERRIDE="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "$GENTX_DIR" ]]; then
  echo "ERROR: --gentx-dir is required" >&2
  usage
  exit 1
fi

if [[ ! -d "$GENTX_DIR" ]]; then
  echo "ERROR: gentx dir not found: $GENTX_DIR" >&2
  exit 1
fi

for file in "$PARAMS" "$ALLOCATIONS" "$CHECKS"; do
  if [[ ! -f "$file" ]]; then
    echo "ERROR: required file not found: $file" >&2
    exit 1
  fi
done

if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq is required" >&2
  exit 1
fi

if ! command -v virtengine >/dev/null 2>&1; then
  echo "ERROR: virtengine binary not found in PATH" >&2
  exit 1
fi

CHAIN_ID=$(jq -r '.chain_id' "$PARAMS")
GENESIS_TIME=$(jq -r '.genesis_time' "$PARAMS")

if [[ -n "$CHAIN_ID_OVERRIDE" ]]; then
  CHAIN_ID="$CHAIN_ID_OVERRIDE"
fi

if [[ -n "$GENESIS_TIME_OVERRIDE" ]]; then
  GENESIS_TIME="$GENESIS_TIME_OVERRIDE"
fi

mkdir -p "$HOME_DIR"

if [[ -f "$HOME_DIR/config/genesis.json" ]]; then
  echo "INFO: existing genesis found at $HOME_DIR/config/genesis.json (will overwrite)"
  rm -rf "$HOME_DIR"
  mkdir -p "$HOME_DIR"
fi

virtengine init mainnet-genesis --chain-id "$CHAIN_ID" --home "$HOME_DIR" >/dev/null

GENESIS="$HOME_DIR/config/genesis.json"

scripts/mainnet/genesis-apply-params.sh \
  --genesis "$GENESIS" \
  --params "$PARAMS" \
  --chain-id "$CHAIN_ID" \
  --genesis-time "$GENESIS_TIME"

apply_allocations() {
  local accounts
  accounts=$(jq -c '.accounts[]' "$ALLOCATIONS")
  while IFS= read -r account; do
    local address coins
    address=$(echo "$account" | jq -r '.address')
    coins=$(echo "$account" | jq -r '[.coins[] | "\(.amount)\(.denom)"] | join(",")')

    if [[ -z "$address" || -z "$coins" ]]; then
      echo "ERROR: allocation missing address/coins" >&2
      exit 1
    fi

    if echo "$account" | jq -e '.vesting' >/dev/null 2>&1; then
      local vest_amount start_time end_time start_ts end_ts
      vest_amount=$(echo "$account" | jq -r '.vesting.amount')
      start_time=$(echo "$account" | jq -r '.vesting.start_time')
      end_time=$(echo "$account" | jq -r '.vesting.end_time')

      if [[ -z "$vest_amount" || -z "$start_time" || -z "$end_time" ]]; then
        echo "ERROR: vesting entry missing amount/start/end" >&2
        exit 1
      fi

      start_ts=$(date -u -d "$start_time" +%s)
      end_ts=$(date -u -d "$end_time" +%s)

      virtengine genesis add-genesis-account \
        "$address" \
        "$coins" \
        --vesting-amount "$vest_amount" \
        --vesting-start-time "$start_ts" \
        --vesting-end-time "$end_ts" \
        --home "$HOME_DIR" >/dev/null
    else
      virtengine genesis add-genesis-account \
        "$address" \
        "$coins" \
        --home "$HOME_DIR" >/dev/null
    fi
  done <<< "$accounts"
}

apply_allocations

scripts/mainnet/validate-gentx.sh --gentx-dir "$GENTX_DIR"

mkdir -p "$HOME_DIR/config/gentx"
cp "$GENTX_DIR"/*.json "$HOME_DIR/config/gentx/"

collect_gentxs() {
  if virtengine genesis collect-gentxs --home "$HOME_DIR" >/dev/null 2>&1; then
    return 0
  fi
  if virtengine genesis collect --home "$HOME_DIR" >/dev/null 2>&1; then
    return 0
  fi
  echo "ERROR: could not run genesis collect-gentxs/collect" >&2
  exit 1
}

collect_gentxs

scripts/mainnet/genesis-validate.sh \
  --genesis "$GENESIS" \
  --checks "$CHECKS" \
  --home "$HOME_DIR"

mkdir -p "$OUTPUT_DIR"
cp "$GENESIS" "$OUTPUT_DIR/genesis.json"

hash=$(scripts/mainnet/genesis-hash.sh --genesis "$GENESIS")

echo "$hash  genesis.json" > "$OUTPUT_DIR/genesis.sha256"

echo "Genesis ceremony complete"