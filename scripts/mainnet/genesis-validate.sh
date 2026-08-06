#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/mainnet/genesis-validate.sh \
  --genesis <path> \
  [--checks <path>] \
  [--allocations <path>] \
  [--home <virtengine-home>] \
  [--binary <virtengine-path>]

Runs deterministic validation checks against a genesis.json file.
USAGE
}

die() {
  echo "ERROR: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required"
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

compare_supply_totals() {
  local allocations_path=$1
  declare -A expected_totals=()
  declare -A actual_totals=()

  while IFS= read -r coin; do
    local denom amount
    denom=$(echo "$coin" | jq -r '.denom // empty')
    amount=$(echo "$coin" | jq -r '.amount // empty')
    [[ -n "$denom" && "$amount" =~ ^[0-9]+$ ]] || die "invalid allocation coin entry in $allocations_path"
    expected_totals["$denom"]=$(( ${expected_totals["$denom"]:-0} + amount ))
  done < <(jq -c '([.accounts[]?.coins[]?] + [.community_pool.coins[]?])[]?' "$allocations_path")

  while IFS= read -r coin; do
    local denom amount
    denom=$(echo "$coin" | jq -r '.denom // empty')
    amount=$(echo "$coin" | jq -r '.amount // empty')
    [[ -n "$denom" && "$amount" =~ ^[0-9]+$ ]] || die "invalid bank supply coin entry in $GENESIS"
    actual_totals["$denom"]=$(( ${actual_totals["$denom"]:-0} + amount ))
  done < <(jq -c '.app_state.bank.supply[]?' "$GENESIS")

  for denom in "${!expected_totals[@]}"; do
    if [[ "${actual_totals[$denom]:-0}" -ne "${expected_totals[$denom]}" ]]; then
      echo "FAIL: bank supply for $denom expected ${expected_totals[$denom]}, got ${actual_totals[$denom]:-0}" >&2
      SUPPLY_FAILURES=$((SUPPLY_FAILURES + 1))
    fi
  done

  for denom in "${!actual_totals[@]}"; do
    if [[ -z "${expected_totals[$denom]+set}" ]]; then
      echo "FAIL: unexpected supply denom present in genesis: $denom=${actual_totals[$denom]}" >&2
      SUPPLY_FAILURES=$((SUPPLY_FAILURES + 1))
    fi
  done
}

compare_community_pool_state() {
  local allocations_path=$1 expected actual
  expected=$(jq -c -S '
    (.community_pool.coins // [])
    | map({
        denom: .denom,
        amount: (
          .amount
          | tostring
          | if contains(".") then . else . + ".000000000000000000" end
        )
      })
  ' "$allocations_path")
  actual=$(jq -c -S '.app_state.distribution.fee_pool.community_pool // []' "$GENESIS")

  if [[ "$expected" != "$actual" ]]; then
    echo "FAIL: distribution community pool state does not match allocations" >&2
    echo "  expected: $expected" >&2
    echo "  actual:   $actual" >&2
    SUPPLY_FAILURES=$((SUPPLY_FAILURES + 1))
  fi
}

GENESIS=""
CHECKS="config/mainnet/genesis-checks.json"
ALLOCATIONS=""
HOME_DIR=""
VIRTENGINE_BIN="${VIRTENGINE_BIN:-virtengine}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --genesis)
      GENESIS="$2"
      shift 2
      ;;
    --checks)
      CHECKS="$2"
      shift 2
      ;;
    --allocations)
      ALLOCATIONS="$2"
      shift 2
      ;;
    --home)
      HOME_DIR="$2"
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

[[ -n "$GENESIS" ]] || {
  usage
  die "--genesis is required"
}
[[ -f "$GENESIS" ]] || die "genesis file not found: $GENESIS"
[[ -f "$CHECKS" ]] || die "checks file not found: $CHECKS"

if [[ -n "$ALLOCATIONS" ]]; then
  [[ -f "$ALLOCATIONS" ]] || die "allocations file not found: $ALLOCATIONS"
fi

require_cmd jq

failures=0
SUPPLY_FAILURES=0

while IFS= read -r entry; do
  check_type=$(echo "$entry" | jq -r '.type // "equals"')
  path=$(echo "$entry" | jq -c '.path')
  optional=$(echo "$entry" | jq -r '.optional // false')
  actual=$(jq -c --argjson path "$path" 'getpath($path)' "$GENESIS" 2>/dev/null || echo "__MISSING__")

  case "$check_type" in
    equals)
      expected=$(echo "$entry" | jq -c '.value')
      if [[ "$actual" == "__MISSING__" || "$actual" == "null" ]]; then
        if [[ "$optional" == "true" ]]; then
          continue
        fi
        echo "FAIL: missing path $path" >&2
        failures=$((failures + 1))
        continue
      fi
      if [[ "$actual" != "$expected" ]]; then
        echo "FAIL: path $path expected $expected, got $actual" >&2
        failures=$((failures + 1))
      fi
      ;;
    exists)
      if [[ "$actual" == "__MISSING__" || "$actual" == "null" ]]; then
        if [[ "$optional" == "true" ]]; then
          continue
        fi
        echo "FAIL: missing required path $path" >&2
        failures=$((failures + 1))
      fi
      ;;
    not_placeholder)
      if [[ "$actual" == "__MISSING__" || "$actual" == "null" ]]; then
        if [[ "$optional" == "true" ]]; then
          continue
        fi
        echo "FAIL: missing required path $path" >&2
        failures=$((failures + 1))
        continue
      fi
      actual_value=$(jq -r --argjson path "$path" 'getpath($path)' "$GENESIS")
      if is_placeholder_string "$actual_value"; then
        echo "FAIL: placeholder value at path $path: $actual_value" >&2
        failures=$((failures + 1))
      fi
      ;;
    *)
      echo "FAIL: unsupported check type '$check_type' in $CHECKS" >&2
      failures=$((failures + 1))
      ;;
  esac
done < <(jq -c '.[]' "$CHECKS")

declare -A seen_placeholder_values=()
while IFS= read -r value; do
  value=$(trim "$value")
  [[ -n "$value" ]] || continue
  if is_placeholder_string "$value"; then
    if [[ -z "${seen_placeholder_values[$value]+set}" ]]; then
      echo "FAIL: placeholder string present in genesis payload: $value" >&2
      seen_placeholder_values["$value"]=1
      failures=$((failures + 1))
    fi
  fi
done < <(jq -r '.. | strings' "$GENESIS")

if [[ -n "$ALLOCATIONS" ]]; then
  compare_supply_totals "$ALLOCATIONS"
  compare_community_pool_state "$ALLOCATIONS"
  failures=$((failures + SUPPLY_FAILURES))
fi

if [[ $failures -gt 0 ]]; then
  echo "Genesis validation failed: $failures issue(s)" >&2
  exit 1
fi

echo "Genesis checks passed"

if command -v "$VIRTENGINE_BIN" >/dev/null 2>&1; then
  validate_home="$HOME_DIR"
  cleanup_home=""
  if [[ -z "$validate_home" ]]; then
    mkdir -p ./.cache
    validate_home=$(mktemp -d "./.cache/genesis-validate.XXXXXX")
    cleanup_home="$validate_home"
    mkdir -p "$validate_home/config"
    cp "$GENESIS" "$validate_home/config/genesis.json"
  fi

  if "$VIRTENGINE_BIN" genesis validate --home "$validate_home" >/dev/null 2>&1; then
    echo "virtengine genesis validate: ok"
  else
    [[ -n "$cleanup_home" ]] && rm -rf "$cleanup_home"
    die "virtengine genesis validate failed"
  fi

  if [[ -n "$cleanup_home" ]]; then
    rm -rf "$cleanup_home"
  fi
fi
