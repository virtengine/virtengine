#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/mainnet/validate-gentx.sh \
  --gentx-dir <dir> \
  [--constraints <path>]

Validates gentx files against mainnet constraints.
USAGE
}

GENTX_DIR=""
CONSTRAINTS="config/mainnet/gentx-constraints.json"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --gentx-dir)
      GENTX_DIR="$2"
      shift 2
      ;;
    --constraints)
      CONSTRAINTS="$2"
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

if [[ ! -f "$CONSTRAINTS" ]]; then
  echo "ERROR: constraints file not found: $CONSTRAINTS" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq is required" >&2
  exit 1
fi

if ! command -v python3 >/dev/null 2>&1; then
  echo "ERROR: python3 is required" >&2
  exit 1
fi

dec_ge() {
  python3 - <<PY
from decimal import Decimal
import sys
lhs = Decimal("$1")
rhs = Decimal("$2")
print("true" if lhs >= rhs else "false")
PY
}

dec_le() {
  python3 - <<PY
from decimal import Decimal
import sys
lhs = Decimal("$1")
rhs = Decimal("$2")
print("true" if lhs <= rhs else "false")
PY
}

min_self=$(jq -r '.min_self_delegation' "$CONSTRAINTS")
bond_denom=$(jq -r '.bond_denom' "$CONSTRAINTS")
min_rate=$(jq -r '.commission.min_rate' "$CONSTRAINTS")
max_rate=$(jq -r '.commission.max_rate' "$CONSTRAINTS")
max_change_rate=$(jq -r '.commission.max_change_rate' "$CONSTRAINTS")

failures=0

shopt -s nullglob
files=("$GENTX_DIR"/*.json)
if [[ ${#files[@]} -eq 0 ]]; then
  echo "ERROR: no gentx files found in $GENTX_DIR" >&2
  exit 1
fi

for file in "${files[@]}"; do
  msg='(.body.messages[0].value // .body.messages[0])'
  amount=$(jq -r "$msg.amount.amount // $msg.amount[0].amount // empty" "$file")
  denom=$(jq -r "$msg.amount.denom // $msg.amount[0].denom // empty" "$file")
  rate=$(jq -r "$msg.commission.rate // empty" "$file")
  max_rate_val=$(jq -r "$msg.commission.max_rate // empty" "$file")
  max_change_val=$(jq -r "$msg.commission.max_change_rate // empty" "$file")
  min_self_val=$(jq -r "$msg.min_self_delegation // empty" "$file")
  moniker=$(jq -r "$msg.description.moniker // empty" "$file")
  identity=$(jq -r "$msg.description.identity // empty" "$file")
  website=$(jq -r "$msg.description.website // empty" "$file")

  if [[ -z "$amount" || -z "$denom" ]]; then
    echo "FAIL: $file missing amount/denom" >&2
    failures=$((failures + 1))
    continue
  fi

  if [[ "$denom" != "$bond_denom" ]]; then
    echo "FAIL: $file denom $denom != $bond_denom" >&2
    failures=$((failures + 1))
  fi

  if [[ -n "$min_self_val" && "$min_self_val" =~ ^[0-9]+$ ]]; then
    if [[ "$min_self_val" -lt "$min_self" ]]; then
      echo "FAIL: $file min_self_delegation $min_self_val < $min_self" >&2
      failures=$((failures + 1))
    fi
  else
    echo "FAIL: $file missing min_self_delegation" >&2
    failures=$((failures + 1))
  fi

  if [[ "$amount" =~ ^[0-9]+$ && "$amount" -lt "$min_self" ]]; then
    echo "FAIL: $file stake amount $amount < $min_self" >&2
    failures=$((failures + 1))
  fi

  if [[ -n "$rate" ]]; then
    if [[ "$(dec_ge "$rate" "$min_rate")" != "true" ]]; then
      echo "FAIL: $file commission rate $rate < $min_rate" >&2
      failures=$((failures + 1))
    fi
  else
    echo "FAIL: $file missing commission rate" >&2
    failures=$((failures + 1))
  fi

  if [[ -n "$max_rate_val" ]]; then
    if [[ "$(dec_le "$max_rate_val" "$max_rate")" != "true" ]]; then
      echo "FAIL: $file commission max_rate $max_rate_val > $max_rate" >&2
      failures=$((failures + 1))
    fi
  else
    echo "FAIL: $file missing commission max_rate" >&2
    failures=$((failures + 1))
  fi

  if [[ -n "$max_change_val" ]]; then
    if [[ "$(dec_le "$max_change_val" "$max_change_rate")" != "true" ]]; then
      echo "FAIL: $file commission max_change_rate $max_change_val > $max_change_rate" >&2
      failures=$((failures + 1))
    fi
  else
    echo "FAIL: $file missing commission max_change_rate" >&2
    failures=$((failures + 1))
  fi

  if [[ -z "$moniker" || -z "$identity" || -z "$website" ]]; then
    echo "FAIL: $file missing validator description fields" >&2
    failures=$((failures + 1))
  fi

done

if [[ $failures -gt 0 ]]; then
  echo "Gentx validation failed: $failures issue(s)" >&2
  exit 1
fi

echo "Gentx validation passed"