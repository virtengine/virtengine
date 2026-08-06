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
  [--constraints <path>] \
  [--output <dir>] \
  [--chain-id <id>] \
  [--genesis-time <rfc3339>] \
  [--binary <virtengine-path>]

Builds the mainnet genesis.json deterministically from config inputs.
USAGE
}

die() {
  echo "ERROR: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

resolve_python() {
  if command -v python3 >/dev/null 2>&1 && python3 -c "import sys" >/dev/null 2>&1; then
    PYTHON_CMD=(python3)
    return 0
  fi
  if command -v python >/dev/null 2>&1 && python -c "import sys" >/dev/null 2>&1; then
    PYTHON_CMD=(python)
    return 0
  fi
  if command -v py >/dev/null 2>&1 && py -3 -c "import sys" >/dev/null 2>&1; then
    PYTHON_CMD=(py -3)
    return 0
  fi
  die "a working Python 3 interpreter is required"
}

run_quiet() {
  local output
  if ! output=$("$@" </dev/null 2>&1); then
    [[ -n "$output" ]] && echo "$output" >&2
    return 1
  fi
}

trim() {
  printf '%s' "$1" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
}

is_placeholder_string() {
  local value=$1
  [[ -z "$value" ]] && return 1

  local lower
  lower=$(printf '%s' "$value" | tr '[:upper:]' '[:lower:]')

  [[ "$value" =~ ^\<.*\>$ ]] && return 0
  [[ "$lower" == *"placeholder"* ]] && return 0
  [[ "$lower" == *"example.com"* || "$lower" == *"example.org"* || "$lower" == *"example.net"* ]] && return 0
  [[ "$lower" == *"changeme"* || "$lower" == *"todo"* || "$lower" == *"tbd"* ]] && return 0
  [[ "$value" =~ ^(ve1|vevaloper1|virtengine1).*[xX]{4,}.*$ ]] && return 0
  return 1
}

rfc3339_to_epoch() {
  "${PYTHON_CMD[@]}" - "$1" <<'PY'
from datetime import datetime
import sys

value = sys.argv[1]
if value.endswith("Z"):
    value = value[:-1] + "+00:00"
dt = datetime.fromisoformat(value)
if dt.tzinfo is None:
    raise SystemExit("timestamp must include timezone information")
print(int(dt.timestamp()))
PY
}

canonical_json_in_place() {
  local path=$1
  local next_file="${path}.sorted"
  jq -S . "$path" > "$next_file"
  mv "$next_file" "$path"
}

canonical_json_sha() {
  local path=$1
  jq -S . "$path" | sha256sum | awk '{print $1}'
}

normalize_vesting_amount() {
  local account_json=$1
  local raw_amount vest_denom coin_count single_denom
  raw_amount=$(echo "$account_json" | jq -r '.vesting.amount // empty')
  vest_denom=$(echo "$account_json" | jq -r '.vesting.denom // empty')
  coin_count=$(echo "$account_json" | jq -r '.coins | length')

  [[ -n "$raw_amount" ]] || die "vesting amount is required when vesting block is present"

  if [[ "$raw_amount" =~ ^[0-9]+$ ]]; then
    if [[ -z "$vest_denom" ]]; then
      if [[ "$coin_count" != "1" ]]; then
        die "numeric vesting amount requires a single coin denom or explicit vesting.denom"
      fi
      single_denom=$(echo "$account_json" | jq -r '.coins[0].denom // empty')
      [[ -n "$single_denom" ]] || die "could not determine denom for vesting amount"
      vest_denom="$single_denom"
    fi
    printf '%s%s\n' "$raw_amount" "$vest_denom"
    return 0
  fi

  printf '%s\n' "$raw_amount"
}

community_pool_dec_json() {
  jq -c '
    (.community_pool.coins // [])
    | map({
        denom: .denom,
        amount: (
          .amount
          | tostring
          | if contains(".") then . else . + ".000000000000000000" end
        )
      })
  ' "$ALLOCATIONS"
}

validate_allocation_inputs() {
  local account_count failures=0 blocked_status blocked_count blocker_id blocker_summary
  blocked_status=$(jq -r '.status // empty' "$ALLOCATIONS")
  blocked_count=$(jq -r '(.blocked_allocations // []) | length' "$ALLOCATIONS")

  if [[ "$blocked_status" == "BLOCKED" || "$blocked_count" -gt 0 ]]; then
    blocker_id=$(jq -r '.blocker_id // "GENESIS-EVIDENCE"' "$ALLOCATIONS")
    blocker_summary=$(jq -r '.blocker_summary // "allocation evidence is incomplete"' "$ALLOCATIONS")
    echo "FAIL: allocations file is blocked ($blocker_id): $blocker_summary" >&2
    jq -r '.blocked_allocations // [] | .[] | "FAIL: missing allocation evidence for \(.name): \(.reason // "no reason provided")"' "$ALLOCATIONS" >&2
    die "allocation evidence is incomplete; refusing genesis assembly"
  fi

  account_count=$(jq -r '.accounts | length' "$ALLOCATIONS")
  [[ "$account_count" =~ ^[0-9]+$ ]] || die "allocations file is malformed"
  (( account_count > 0 )) || die "allocations file must contain at least one account"

  declare -A seen_addresses=()

  while IFS= read -r account; do
    local name address unique_denoms
    name=$(echo "$account" | jq -r '.name // empty')
    address=$(echo "$account" | jq -r '.address // empty')

    if [[ -z "$name" || -z "$address" ]]; then
      echo "FAIL: allocation entry missing name or address" >&2
      failures=$((failures + 1))
      continue
    fi

    if is_placeholder_string "$address"; then
      echo "FAIL: placeholder allocation address detected for $name: $address" >&2
      failures=$((failures + 1))
    fi

    if [[ -n "${seen_addresses[$address]+set}" ]]; then
      echo "FAIL: duplicate allocation address detected: $address" >&2
      failures=$((failures + 1))
    fi
    seen_addresses["$address"]=1

    if [[ ! "$address" =~ ^ve1[0-9a-z]+$ ]]; then
      echo "FAIL: allocation address is not a ve bech32 account address: $address" >&2
      failures=$((failures + 1))
    fi

    unique_denoms=$(echo "$account" | jq -r '[.coins[]?.denom] | unique | length')
    while IFS= read -r coin; do
      local denom amount
      denom=$(echo "$coin" | jq -r '.denom // empty')
      amount=$(echo "$coin" | jq -r '.amount // empty')

      if [[ -z "$denom" || -z "$amount" || ! "$amount" =~ ^[0-9]+$ || "$amount" -le 0 ]]; then
        echo "FAIL: allocation coin entry for $name is malformed" >&2
        failures=$((failures + 1))
      fi

      if is_placeholder_string "$denom"; then
        echo "FAIL: placeholder denom detected for $name: $denom" >&2
        failures=$((failures + 1))
      fi
    done < <(echo "$account" | jq -c '.coins[]?')

    if echo "$account" | jq -e '.vesting? != null' >/dev/null 2>&1; then
      local start_time end_time vesting_coin vesting_amount vesting_denom account_amount
      start_time=$(echo "$account" | jq -r '.vesting.start_time // empty')
      end_time=$(echo "$account" | jq -r '.vesting.end_time // empty')

      if [[ -z "$start_time" || -z "$end_time" ]]; then
        echo "FAIL: vesting entry for $name is missing start_time or end_time" >&2
        failures=$((failures + 1))
      else
        start_epoch=$(rfc3339_to_epoch "$start_time" 2>/dev/null || echo "__INVALID__")
        end_epoch=$(rfc3339_to_epoch "$end_time" 2>/dev/null || echo "__INVALID__")
        if [[ "$start_epoch" == "__INVALID__" || "$end_epoch" == "__INVALID__" || "$start_epoch" -ge "$end_epoch" ]]; then
          echo "FAIL: vesting schedule for $name has invalid timestamps" >&2
          failures=$((failures + 1))
        fi
      fi

      if ! vesting_coin=$(normalize_vesting_amount "$account" 2>/dev/null); then
        echo "FAIL: vesting amount for $name could not be normalized" >&2
        failures=$((failures + 1))
      else
        if [[ "$vesting_coin" =~ ^([0-9]+)([A-Za-z0-9/._-]+)$ ]]; then
          vesting_amount=${BASH_REMATCH[1]}
          vesting_denom=${BASH_REMATCH[2]}
          account_amount=$(echo "$account" | jq -r --arg denom "$vesting_denom" '[.coins[] | select(.denom == $denom) | .amount | tonumber] | add // 0')
          if [[ "$account_amount" =~ ^[0-9]+$ ]] && [[ "$vesting_amount" -gt "$account_amount" ]]; then
            echo "FAIL: vesting amount for $name exceeds allocated $vesting_denom balance" >&2
            failures=$((failures + 1))
          fi
        else
          echo "FAIL: normalized vesting amount for $name is invalid: $vesting_coin" >&2
          failures=$((failures + 1))
        fi
      fi

      if [[ "$unique_denoms" -gt 1 ]] && [[ -z "$(echo "$account" | jq -r '.vesting.denom // empty')" ]]; then
        echo "FAIL: multi-denom vesting entry for $name must include vesting.denom" >&2
        failures=$((failures + 1))
      fi
    fi
  done < <(jq -c '.accounts[]' "$ALLOCATIONS")

  if ! jq -e '.community_pool? != null' "$ALLOCATIONS" >/dev/null 2>&1; then
    echo "FAIL: allocations file must define a community_pool block" >&2
    failures=$((failures + 1))
  else
    local module_account community_address expected_module_address community_coin_count
    module_account=$(jq -r '.community_pool.module_account // empty' "$ALLOCATIONS")
    community_address=$(jq -r '.community_pool.address // empty' "$ALLOCATIONS")
    community_coin_count=$(jq -r '.community_pool.coins | length' "$ALLOCATIONS")

    if [[ "$module_account" != "distribution" ]]; then
      echo "FAIL: community_pool.module_account must be 'distribution'" >&2
      failures=$((failures + 1))
    fi

    if [[ -z "$community_address" ]]; then
      echo "FAIL: community_pool.address is required" >&2
      failures=$((failures + 1))
    elif is_placeholder_string "$community_address"; then
      echo "FAIL: placeholder community_pool.address detected: $community_address" >&2
      failures=$((failures + 1))
    elif [[ ! "$community_address" =~ ^ve1[0-9a-z]+$ ]]; then
      echo "FAIL: community_pool.address is not a ve bech32 account address: $community_address" >&2
      failures=$((failures + 1))
    else
      expected_module_address=$("$VIRTENGINE_BIN" debug module-address "$module_account" 2>/dev/null || echo "__INVALID__")
      if [[ "$expected_module_address" == "__INVALID__" || "$expected_module_address" != "$community_address" ]]; then
        echo "FAIL: community_pool.address must match the $module_account module account address ($expected_module_address)" >&2
        failures=$((failures + 1))
      fi
    fi

    if [[ -n "$community_address" && -n "${seen_addresses[$community_address]+set}" ]]; then
      echo "FAIL: duplicate allocation address detected for community_pool: $community_address" >&2
      failures=$((failures + 1))
    elif [[ -n "$community_address" ]]; then
      seen_addresses["$community_address"]=1
    fi

    if [[ ! "$community_coin_count" =~ ^[0-9]+$ || "$community_coin_count" -lt 1 ]]; then
      echo "FAIL: community_pool must contain at least one coin entry" >&2
      failures=$((failures + 1))
    fi

    while IFS= read -r coin; do
      local denom amount
      denom=$(echo "$coin" | jq -r '.denom // empty')
      amount=$(echo "$coin" | jq -r '.amount // empty')

      if [[ -z "$denom" || -z "$amount" || ! "$amount" =~ ^[0-9]+$ || "$amount" -le 0 ]]; then
        echo "FAIL: community_pool coin entry is malformed" >&2
        failures=$((failures + 1))
      fi

      if is_placeholder_string "$denom"; then
        echo "FAIL: placeholder denom detected for community_pool: $denom" >&2
        failures=$((failures + 1))
      fi
    done < <(jq -c '.community_pool.coins[]?' "$ALLOCATIONS")
  fi

  if [[ $failures -gt 0 ]]; then
    die "allocation validation failed with $failures issue(s)"
  fi
}

build_allocation_maps() {
  EXPECTED_SUPPLY_FILE=$(mktemp)
  ALLOCATION_MAP_FILE=$(mktemp)
  CLEANUP_PATHS+=("$EXPECTED_SUPPLY_FILE" "$ALLOCATION_MAP_FILE")

  jq -r '
    (
      [.accounts[] | {address, coins}]
      + (
        if .community_pool? != null then
          [{address: .community_pool.address, coins: .community_pool.coins}]
        else
          []
        end
      )
    )[]
    | .address as $address
    | .coins[]
    | [$address, .denom, .amount]
    | @tsv
  ' "$ALLOCATIONS" | tr -d '\r' > "$ALLOCATION_MAP_FILE"
  jq -r '
    (
      [.accounts[].coins[]]
      + (.community_pool.coins // [])
    )
    | group_by(.denom)[]
    | [.[0].denom, (map(.amount | tonumber) | add | tostring)]
    | @tsv
  ' "$ALLOCATIONS" | tr -d '\r' > "$EXPECTED_SUPPLY_FILE"
}

validate_gentx_funding() {
  declare -A allocation_by_address_denom=()
  declare -A required_by_address_denom=()

  while IFS=$'\t' read -r address denom amount; do
    address=${address%$'\r'}
    denom=${denom%$'\r'}
    amount=${amount%$'\r'}
    allocation_by_address_denom["$address|$denom"]=$amount
  done < "$ALLOCATION_MAP_FILE"

  while IFS= read -r file; do
    local validator_address validator_account denom amount key msg
    msg='.body.messages[0]'
    validator_address=$(jq -r '.body.messages[0].validator_address // empty' "$file")
    denom=$(jq -r "$msg.value.denom // $msg.amount.denom // $msg.amount[0].denom // $msg.value[0].denom // empty" "$file")
    amount=$(jq -r "$msg.value.amount // $msg.amount.amount // $msg.amount[0].amount // $msg.value[0].amount // empty" "$file")

    [[ -n "$validator_address" && -n "$denom" && "$amount" =~ ^[0-9]+$ ]] || die "gentx funding check could not parse $file"
    validator_account=$("$VIRTENGINE_BIN" debug bech32-convert --prefix ve "$validator_address" 2>/dev/null) || die "could not convert validator address to funding account: $validator_address"
    key="$validator_account|$denom"
    required_by_address_denom["$key"]=$(( ${required_by_address_denom["$key"]:-0} + amount ))
  done < <(printf '%s\n' "$GENTX_DIR"/*.json | sort)

  for key in "${!required_by_address_denom[@]}"; do
    local current address denom
    current=${allocation_by_address_denom["$key"]:-0}
    if [[ "$current" -lt "${required_by_address_denom[$key]}" ]]; then
      address=${key%%|*}
      denom=${key#*|}
      echo "FAIL: allocation funding for $address is ${current}${denom}, but submitted gentxs require ${required_by_address_denom[$key]}${denom}" >&2
      die "gentx self-delegation funding does not match allocations"
    fi
  done
}

write_gentx_hashes() {
  local output_path=$1
  : > "$output_path"
  while IFS= read -r file; do
    hash=$(canonical_json_sha "$file")
    printf '%s  %s\n' "$hash" "$(basename "$file")" >> "$output_path"
  done < <(printf '%s\n' "$GENTX_DIR"/*.json | sort)
}

build_manifest() {
  local output_path=$1
  local genesis_hash=$2
  local gentx_hash_file=$3
  local gentx_rows_file

  gentx_rows_file=$(mktemp)
  CLEANUP_PATHS+=("$gentx_rows_file")

  jq -Rn '
    [inputs
     | split("  ")
     | select(length >= 2)
     | {canonical_sha256: .[0], file: (.[1:] | join("  "))}]
  ' < "$gentx_hash_file" > "$gentx_rows_file"

  jq -n \
    --arg chain_id "$CHAIN_ID" \
    --arg genesis_time "$GENESIS_TIME" \
    --arg params_path "$PARAMS" \
    --arg allocations_path "$ALLOCATIONS" \
    --arg checks_path "$CHECKS" \
    --arg constraints_path "$CONSTRAINTS" \
    --arg output_dir "$OUTPUT_DIR" \
    --arg genesis_hash "$genesis_hash" \
    --arg params_hash "$(canonical_json_sha "$PARAMS")" \
    --arg allocations_hash "$(canonical_json_sha "$ALLOCATIONS")" \
    --arg checks_hash "$(canonical_json_sha "$CHECKS")" \
    --arg constraints_hash "$(canonical_json_sha "$CONSTRAINTS")" \
    --slurpfile gentx_rows "$gentx_rows_file" \
    '{
      format: "virtengine-mainnet-ceremony-manifest/v1",
      hash_mode: {
        algorithm: "sha256",
        json_inputs: "canonical-jq-sorted",
        genesis_artifact: "canonical-file-bytes"
      },
      chain_id: $chain_id,
      genesis_time: $genesis_time,
      inputs: {
        params: {path: $params_path, canonical_sha256: $params_hash},
        allocations: {path: $allocations_path, canonical_sha256: $allocations_hash},
        checks: {path: $checks_path, canonical_sha256: $checks_hash},
        constraints: {path: $constraints_path, canonical_sha256: $constraints_hash},
        gentxs: $gentx_rows[0]
      },
      outputs: {
        genesis: {
          path: ($output_dir + "/genesis.json"),
          sha256: $genesis_hash
        },
        genesis_sha256: {
          path: ($output_dir + "/genesis.sha256")
        },
        gentx_sha256: {
          path: ($output_dir + "/gentx.sha256")
        }
      }
    }' > "$output_path"
  canonical_json_in_place "$output_path"
}

GENTX_DIR=""
HOME_DIR=".cache/mainnet-genesis"
PARAMS="config/mainnet/genesis-params.json"
ALLOCATIONS="config/mainnet/genesis-allocations.json"
CHECKS="config/mainnet/genesis-checks.json"
CONSTRAINTS="config/mainnet/gentx-constraints.json"
OUTPUT_DIR="artifacts/mainnet"
CHAIN_ID_OVERRIDE=""
GENESIS_TIME_OVERRIDE=""
VIRTENGINE_BIN="${VIRTENGINE_BIN:-virtengine}"
CLEANUP_PATHS=()

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
    --constraints)
      CONSTRAINTS="$2"
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
    --binary)
      VIRTENGINE_BIN="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

trap 'for path in "${CLEANUP_PATHS[@]}"; do [[ -e "$path" ]] && rm -rf "$path"; done' EXIT

[[ -n "$GENTX_DIR" ]] || {
  usage
  die "--gentx-dir is required"
}
[[ -d "$GENTX_DIR" ]] || die "gentx dir not found: $GENTX_DIR"

for file in "$PARAMS" "$ALLOCATIONS" "$CHECKS" "$CONSTRAINTS"; do
  [[ -f "$file" ]] || die "required file not found: $file"
done

require_cmd jq
require_cmd sha256sum
require_cmd "$VIRTENGINE_BIN"
resolve_python

CHAIN_ID=$(jq -r '.chain_id' "$PARAMS")
GENESIS_TIME=$(jq -r '.genesis_time' "$PARAMS")

if [[ -n "$CHAIN_ID_OVERRIDE" ]]; then
  CHAIN_ID="$CHAIN_ID_OVERRIDE"
fi

if [[ -n "$GENESIS_TIME_OVERRIDE" ]]; then
  GENESIS_TIME="$GENESIS_TIME_OVERRIDE"
fi

validate_allocation_inputs
build_allocation_maps

mkdir -p "$HOME_DIR"

if [[ -f "$HOME_DIR/config/genesis.json" ]]; then
  echo "INFO: existing genesis found at $HOME_DIR/config/genesis.json (will overwrite)"
  rm -rf "$HOME_DIR"
  mkdir -p "$HOME_DIR"
fi

run_quiet "$VIRTENGINE_BIN" genesis init mainnet-genesis --chain-id "$CHAIN_ID" --home "$HOME_DIR" --overwrite || die "virtengine genesis init failed"

GENESIS="$HOME_DIR/config/genesis.json"

bash scripts/mainnet/genesis-apply-params.sh \
  --genesis "$GENESIS" \
  --params "$PARAMS" \
  --chain-id "$CHAIN_ID" \
  --genesis-time "$GENESIS_TIME"

while IFS= read -r account; do
  address=$(echo "$account" | jq -r '.address')
  coins=$(echo "$account" | jq -r '[.coins[] | "\(.amount)\(.denom)"] | join(",")')

  [[ -n "$address" && -n "$coins" ]] || die "allocation entry is missing address or coins"

  if echo "$account" | jq -e '.vesting? != null' >/dev/null 2>&1; then
    vesting_amount=$(normalize_vesting_amount "$account")
    start_time=$(echo "$account" | jq -r '.vesting.start_time')
    end_time=$(echo "$account" | jq -r '.vesting.end_time')
    start_ts=$(rfc3339_to_epoch "$start_time")
    end_ts=$(rfc3339_to_epoch "$end_time")

    run_quiet "$VIRTENGINE_BIN" genesis add-account \
      "$address" \
      "$coins" \
      --vesting-amount "$vesting_amount" \
      --vesting-start-time "$start_ts" \
      --vesting-end-time "$end_ts" \
      --home "$HOME_DIR" || die "failed to add vesting allocation for $address"
  else
    run_quiet "$VIRTENGINE_BIN" genesis add-account \
      "$address" \
      "$coins" \
      --home "$HOME_DIR" || die "failed to add allocation for $address"
  fi
done < <(jq -c '.accounts[]' "$ALLOCATIONS")

community_pool_address=$(jq -r '.community_pool.address // empty' "$ALLOCATIONS")
community_pool_coins=$(jq -r '[.community_pool.coins[]? | "\(.amount)\(.denom)"] | join(",")' "$ALLOCATIONS")
[[ -n "$community_pool_address" && -n "$community_pool_coins" ]] || die "community_pool is missing address or coins"

run_quiet "$VIRTENGINE_BIN" genesis add-account \
  "$community_pool_address" \
  "$community_pool_coins" \
  --home "$HOME_DIR" || die "failed to add community pool module balance"

community_pool_dec=$(community_pool_dec_json)
jq --argjson community_pool "$community_pool_dec" \
  '.app_state.distribution.fee_pool.community_pool = $community_pool' \
  "$GENESIS" > "${GENESIS}.next"
mv "${GENESIS}.next" "$GENESIS"

bash scripts/mainnet/validate-gentx.sh \
  --gentx-dir "$GENTX_DIR" \
  --constraints "$CONSTRAINTS" \
  --binary "$VIRTENGINE_BIN"

validate_gentx_funding

mkdir -p "$HOME_DIR/config/gentx"
cp "$GENTX_DIR"/*.json "$HOME_DIR/config/gentx/"

run_quiet "$VIRTENGINE_BIN" genesis collect --home "$HOME_DIR" --gentx-dir "$HOME_DIR/config/gentx" || die "virtengine genesis collect failed"
canonical_json_in_place "$GENESIS"

bash scripts/mainnet/genesis-validate.sh \
  --genesis "$GENESIS" \
  --checks "$CHECKS" \
  --allocations "$ALLOCATIONS" \
  --home "$HOME_DIR" \
  --binary "$VIRTENGINE_BIN"

mkdir -p "$OUTPUT_DIR"
cp "$GENESIS" "$OUTPUT_DIR/genesis.json"
canonical_json_in_place "$OUTPUT_DIR/genesis.json"

genesis_hash=$(sha256sum "$OUTPUT_DIR/genesis.json" | awk '{print $1}')
printf '%s  %s\n' "$genesis_hash" "genesis.json" > "$OUTPUT_DIR/genesis.sha256"

GENTX_HASH_FILE="$OUTPUT_DIR/gentx.sha256"
write_gentx_hashes "$GENTX_HASH_FILE"

MANIFEST_PATH="$OUTPUT_DIR/ceremony-manifest.json"
build_manifest "$MANIFEST_PATH" "$genesis_hash" "$GENTX_HASH_FILE"
manifest_hash=$(sha256sum "$MANIFEST_PATH" | awk '{print $1}')
printf '%s  %s\n' "$manifest_hash" "ceremony-manifest.json" > "$OUTPUT_DIR/ceremony-manifest.sha256"

echo "Genesis ceremony complete"
echo "  genesis:  $OUTPUT_DIR/genesis.json"
echo "  sha256:   $genesis_hash"
echo "  manifest: $MANIFEST_PATH"
