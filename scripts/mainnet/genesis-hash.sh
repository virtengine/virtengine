#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/mainnet/genesis-hash.sh \
  --genesis <path> \
  [--output <path>] \
  [--label <filename>]

Outputs the deterministic SHA-256 hash of the canonical JSON form of the
provided genesis file. When used with --output, the script writes a standard
sha256sum-formatted line for publication.
USAGE
}

die() {
  echo "ERROR: $*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required"
}

GENESIS=""
OUTPUT=""
LABEL=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --genesis)
      GENESIS="$2"
      shift 2
      ;;
    --output)
      OUTPUT="$2"
      shift 2
      ;;
    --label)
      LABEL="$2"
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

require_cmd jq
require_cmd sha256sum

if [[ -z "$LABEL" ]]; then
  LABEL=$(basename "$GENESIS")
fi

tmp_canonical=$(mktemp)
trap 'rm -f "$tmp_canonical"' EXIT

jq -S . "$GENESIS" > "$tmp_canonical"
hash=$(sha256sum "$tmp_canonical" | awk '{print $1}')

if [[ -n "$OUTPUT" ]]; then
  printf '%s  %s\n' "$hash" "$LABEL" > "$OUTPUT"
fi

printf '%s\n' "$hash"
