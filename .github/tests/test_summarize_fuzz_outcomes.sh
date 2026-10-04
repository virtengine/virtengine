#!/usr/bin/env bash
# Regression tests for .github/scripts/summarize_fuzz_outcomes.sh.
#
# Case 2 is a real bug that shipped: a YAML block scalar value ends with a
# newline, so an inline `grep -v '=(success|skipped)$'` matched the trailing
# empty line and warned on a run where all five fuzz steps had succeeded.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCRIPT="$ROOT/.github/scripts/summarize_fuzz_outcomes.sh"

PASS=0
FAIL=0
ok() { echo "  PASS  $*"; PASS=$((PASS + 1)); }
bad() { echo "  FAIL  $*"; FAIL=$((FAIL + 1)); }

if [ ! -f "$SCRIPT" ]; then
  echo "FAIL: missing script: $SCRIPT"
  exit 1
fi

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

# run <stdin-text> -> writes stdout to $WORK/out, body text to $WORK/summary
run() {
  : > "$WORK/summary"
  printf '%s' "$1" | bash "$SCRIPT" "$WORK/summary" > "$WORK/out" 2>&1
}

# --- case 1: all steps succeeded ------------------------------------------------
run "encryption-crypto=success
encryption-types=success
veid-types=success
market-types=success
mfa-types=success
"
if grep -q "::warning::" "$WORK/out"; then
  bad "all-success run warned"
else
  ok "all-success run is quiet"
fi
if grep -q "encryption-crypto=success" "$WORK/summary"; then
  ok "publishes the outcomes to the job summary"
else
  bad "did not publish the outcomes"
fi

# --- case 2: the shipped false positive ----------------------------------------
# This is exactly the env value the workflow produces: a trailing newline from the
# block scalar, i.e. a final empty line.
run "encryption-crypto=success
encryption-types=success
veid-types=success
market-types=success
mfa-types=success
"
if grep -q "::warning::" "$WORK/out"; then
  bad "trailing blank line still triggers a false warning"
else
  ok "a trailing blank line does not trigger a false warning"
fi

# --- case 3: a genuinely failing step is called out ------------------------------
run "encryption-crypto=failure
encryption-types=success
"
if grep -q "::warning::" "$WORK/out"; then
  ok "a failing step raises a warning"
else
  bad "a failing step was not reported"
fi
if grep -q "encryption-crypto=failure" "$WORK/out"; then
  ok "names the failing step"
else
  bad "did not name the failing step"
fi

# --- case 4: skipped is not a failure -------------------------------------------
run "encryption-crypto=skipped
encryption-types=success
"
if grep -q "::warning::" "$WORK/out"; then
  bad "a skipped step was reported as a failure"
else
  ok "a skipped step is not a failure"
fi

# --- case 5: icons/other outcomes are reported -----------------------------------
run "encryption-crypto=cancelled
"
if grep -q "::warning::" "$WORK/out"; then
  ok "a cancelled step raises a warning"
else
  bad "a cancelled step was not reported"
fi

# --- case 6: nothing to report is quiet ------------------------------------------
run ""
if grep -q "::warning::" "$WORK/out"; then
  bad "empty input warned"
else
  ok "empty input is quiet"
fi

# --- case 7: it never gates ------------------------------------------------------
run "encryption-crypto=failure
"
rc=$?
if [ "$rc" -eq 0 ]; then
  ok "exits 0 even when a step failed (reports, does not gate)"
else
  bad "exited ${rc}; it must not gate"
fi

echo
echo "test_summarize_fuzz_outcomes.sh: ${PASS} passed, ${FAIL} failed"
[ "$FAIL" -eq 0 ]
