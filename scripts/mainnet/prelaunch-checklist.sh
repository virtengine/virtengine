#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/mainnet/prelaunch-checklist.sh \
  [--checklist <path>] \
  [--packet <path>] \
  [--allow-pending] \
  [--allow-unchecked]

Runs automated checks against the mainnet launch readiness checklist and
launch packet evidence hashes.
USAGE
}

CHECKLIST="_docs/operations/mainnet-launch-readiness-checklist.md"
PACKET="_docs/operations/mainnet-launch-packet.md"
ALLOW_PENDING="false"
ALLOW_UNCHECKED="false"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --checklist)
      CHECKLIST="$2"
      shift 2
      ;;
    --packet)
      PACKET="$2"
      shift 2
      ;;
    --allow-pending)
      ALLOW_PENDING="true"
      shift 1
      ;;
    --allow-unchecked)
      ALLOW_UNCHECKED="true"
      shift 1
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

for file in "$CHECKLIST" "$PACKET"; do
  if [[ ! -f "$file" ]]; then
    echo "ERROR: file not found: $file" >&2
    exit 1
  fi
done

failures=0

if [[ "$ALLOW_UNCHECKED" != "true" ]]; then
  if grep -n "\[ \]" "$CHECKLIST" >/dev/null 2>&1; then
    echo "FAIL: unchecked checklist items found in $CHECKLIST" >&2
    grep -n "\[ \]" "$CHECKLIST" >&2
    failures=$((failures + 1))
  fi
fi

if [[ "$ALLOW_PENDING" != "true" ]]; then
  if grep -n "| Pending |" "$CHECKLIST" >/dev/null 2>&1; then
    echo "FAIL: pending sign-offs found in $CHECKLIST" >&2
    grep -n "| Pending |" "$CHECKLIST" >&2
    failures=$((failures + 1))
  fi
  if grep -n "| Pending |" "$PACKET" >/dev/null 2>&1; then
    echo "FAIL: pending evidence entries found in $PACKET" >&2
    grep -n "| Pending |" "$PACKET" >&2
    failures=$((failures + 1))
  fi
fi

if ! command -v sha256sum >/dev/null 2>&1; then
  echo "ERROR: sha256sum is required" >&2
  exit 1
fi

while IFS='|' read -r col1 col2 col3 col4 col5 col6; do
  hash=$(echo "$col5" | tr -d ' ')
  path=$(echo "$col4" | sed 's/^ *//;s/ *$//')

  if [[ "$hash" =~ ^[0-9a-fA-F]{64}$ ]]; then
    if [[ ! -f "$path" ]]; then
      echo "FAIL: missing evidence file $path" >&2
      failures=$((failures + 1))
      continue
    fi

    actual=$(sha256sum "$path" | awk '{print $1}')
    if [[ "$actual" != "$hash" ]]; then
      echo "FAIL: hash mismatch for $path" >&2
      echo "  expected: $hash" >&2
      echo "  actual:   $actual" >&2
      failures=$((failures + 1))
    fi
  fi
done < <(grep -E "\| [0-9a-fA-F]{64} \|" "$PACKET")

if [[ $failures -gt 0 ]]; then
  echo "Pre-launch checklist verification failed: $failures issue(s)" >&2
  exit 1
fi

echo "Pre-launch checklist verification passed"