#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/mainnet/genesis-validate.sh \
  --genesis <path> \
  [--checks <path>] \
  [--home <virtengine-home>]

Runs deterministic validation checks against a genesis.json file.
USAGE
}

GENESIS=""
CHECKS="config/mainnet/genesis-checks.json"
HOME_DIR=""

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
    --home)
      HOME_DIR="$2"
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

if [[ -z "$GENESIS" ]]; then
  echo "ERROR: --genesis is required" >&2
  usage
  exit 1
fi

if [[ ! -f "$GENESIS" ]]; then
  echo "ERROR: genesis file not found: $GENESIS" >&2
  exit 1
fi

if [[ ! -f "$CHECKS" ]]; then
  echo "ERROR: checks file not found: $CHECKS" >&2
  exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq is required" >&2
  exit 1
fi

failures=0

while IFS= read -r entry; do
  path=$(echo "$entry" | jq -c '.path')
  expected=$(echo "$entry" | jq -c '.value')
  optional=$(echo "$entry" | jq -r '.optional // false')

  actual=$(jq -c --argjson path "$path" 'getpath($path)' "$GENESIS" 2>/dev/null || echo "__MISSING__")

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
done < <(jq -c '.[]' "$CHECKS")

if [[ $failures -gt 0 ]]; then
  echo "Genesis validation failed: $failures mismatch(es)" >&2
  exit 1
fi

echo "Genesis checks passed"

if command -v virtengine >/dev/null 2>&1; then
  if [[ -n "$HOME_DIR" ]]; then
    if virtengine genesis validate-genesis --home "$HOME_DIR" >/dev/null 2>&1; then
      echo "virtengine validate-genesis: ok"
    elif virtengine genesis validate --home "$HOME_DIR" >/dev/null 2>&1; then
      echo "virtengine validate: ok"
    else
      echo "WARNING: virtengine validate command failed" >&2
    fi
  else
    if virtengine genesis validate-genesis >/dev/null 2>&1; then
      echo "virtengine validate-genesis: ok"
    elif virtengine genesis validate >/dev/null 2>&1; then
      echo "virtengine validate: ok"
    else
      echo "WARNING: virtengine validate command failed" >&2
    fi
  fi
fi