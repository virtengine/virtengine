#!/usr/bin/env bash
# Regression tests for the fail-closed security-summary result checker.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CHECKER="$ROOT/.github/scripts/check_security_gate_results.sh"

if [ ! -f "$CHECKER" ]; then
  echo "FAIL: missing checker: $CHECKER"
  exit 1
fi

PASS=0
FAIL=0
ok() { echo "  PASS  $*"; PASS=$((PASS + 1)); }
bad() { echo "  FAIL  $*"; FAIL=$((FAIL + 1)); }

successes=(success success success success success success success success success)
if bash "$CHECKER" "${successes[@]}" >/dev/null 2>&1; then
  ok "all nine required security gates succeeded"
else
  bad "all-success result set was rejected"
fi

expect_rejected() {
  local name="$1"
  shift
  if bash "$CHECKER" "$@" >/dev/null 2>&1; then
    bad "$name was accepted as green"
  else
    ok "$name failed closed"
  fi
}

# Simulate a failed govulncheck matrix (one shard errored) and missing gates.
expect_rejected "one failed shard / failed scan job" success success failure success success success success success success
expect_rejected "cancelled required gate" success success success success cancelled success success success success
expect_rejected "skipped required gate" success success skipped success success success success success success
expect_rejected "unknown result" success success maybe success success success success success success
expect_rejected "empty result" success success "" success success success success success success
expect_rejected "missing shard aggregate (only eight gate results)" success success success success success success success success

printf '\nRESULT: %d passed, %d failed\n' "$PASS" "$FAIL"
[ "$FAIL" -eq 0 ]
