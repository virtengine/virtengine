#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/mainnet/genesis-hash.sh --genesis <path>

Outputs a deterministic SHA-256 hash of the provided genesis file.
USAGE
}

GENESIS=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --genesis)
      GENESIS="$2"
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

if ! command -v jq >/dev/null 2>&1; then
  echo "ERROR: jq is required" >&2
  exit 1
fi

if ! command -v sha256sum >/dev/null 2>&1; then
  echo "ERROR: sha256sum is required" >&2
  exit 1
fi

jq -S . "$GENESIS" | sha256sum | awk '{print $1}'