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

die() {
  echo "ERROR: $*" >&2
  exit 1
}

trim() {
  printf '%s' "$1" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
}

is_placeholder_path() {
  local value
  value=$(trim "$1")
  [[ -z "$value" ]] && return 0
  [[ "$value" =~ ^\<.*\>$ ]] && return 0
  return 1
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
      die "unknown argument: $1"
      ;;
  esac
done

for file in "$CHECKLIST" "$PACKET"; do
  [[ -f "$file" ]] || die "file not found: $file"
done

command -v sha256sum >/dev/null 2>&1 || die "sha256sum is required"

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
fi

while IFS='|' read -r _ evidence_id artifact owner location hash _; do
  evidence_id=$(trim "$evidence_id")
  artifact=$(trim "$artifact")
  owner=$(trim "$owner")
  location=$(trim "$location")
  hash=$(trim "$hash")

  [[ -n "$evidence_id" ]] || continue
  [[ "$evidence_id" != "Evidence ID" && "$evidence_id" != "---" ]] || continue

  if [[ -z "$artifact" || -z "$owner" ]]; then
    echo "FAIL: malformed evidence row in $PACKET for $evidence_id" >&2
    failures=$((failures + 1))
    continue
  fi

  if is_placeholder_path "$location"; then
    echo "FAIL: missing evidence location for $evidence_id ($artifact)" >&2
    failures=$((failures + 1))
    continue
  fi

  if [[ "$hash" == "Pending" ]]; then
    if [[ "$ALLOW_PENDING" != "true" ]]; then
      echo "FAIL: pending evidence hash for $evidence_id ($artifact)" >&2
      failures=$((failures + 1))
    fi
    continue
  fi

  if [[ ! "$hash" =~ ^[0-9a-fA-F]{64}$ ]]; then
    echo "FAIL: invalid hash value for $evidence_id ($artifact): $hash" >&2
    failures=$((failures + 1))
    continue
  fi

  if [[ ! -f "$location" ]]; then
    echo "FAIL: missing evidence file for $evidence_id ($artifact): $location" >&2
    failures=$((failures + 1))
    continue
  fi

  actual=$(sha256sum "$location" | awk '{print $1}')
  if [[ "$actual" != "$hash" ]]; then
    echo "FAIL: hash mismatch for $evidence_id ($artifact)" >&2
    echo "  path:     $location" >&2
    echo "  expected: $hash" >&2
    echo "  actual:   $actual" >&2
    failures=$((failures + 1))
  fi
done < "$PACKET"

if [[ $failures -gt 0 ]]; then
  echo "Pre-launch checklist verification failed: $failures issue(s)" >&2
  exit 1
fi

echo "Pre-launch checklist verification passed"
