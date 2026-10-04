#!/usr/bin/env bash
# Hermetic regression tests for .github/scripts/run_fuzz_targets.sh.
#
# A fake `go` on PATH stands in for the toolchain, so these assertions are fast
# and deterministic - no compilation, no network. That is the right level here:
# the original defect was an ARGUMENT bug (`-fuzz=.` matching 10-14 targets and
# aborting before fuzzing), so the thing to pin is the arguments the runner hands
# to `go test`, not the fuzzer itself.
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
RUNNER="$ROOT/.github/scripts/run_fuzz_targets.sh"

PASS=0
FAIL=0
ok() { echo "  PASS  $*"; PASS=$((PASS + 1)); }
bad() { echo "  FAIL  $*"; FAIL=$((FAIL + 1)); }

if [ ! -f "$RUNNER" ]; then
  echo "FAIL: missing runner: $RUNNER"
  exit 1
fi

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
mkdir -p "$WORK/bin"
CALLS="$WORK/calls.log"
: > "$CALLS"
export CALLS

# Stub toolchain: `list` reports the packages under test, `test -list` reports the
# fuzz targets, and a real `test` invocation records its arguments (one per line)
# and exits with the code the case configured.
cat > "$WORK/bin/go" <<'FAKE'
#!/usr/bin/env bash
cmd="${1:-}"
shift || true
case "$cmd" in
  list)
    printf '%s\n' ${FAKE_PACKAGES:-pkg/a}
    ;;
  test)
    if [ "${1:-}" = "-list" ]; then
      printf '%s\n' ${FAKE_TARGETS:-}
      printf 'ok  \t%s\t0.00s\n' "${2:-pkg}"
      exit 0
    fi
    printf '%s\n' "$*" >> "$CALLS"
    printf 'fuzz: elapsed: 3s, execs: 100 (rate: 33/s), new interesting: 0\n'
    for arg in "$@"; do
      case "$arg" in
        "${FAKE_FAIL_TARGET:-\$^}")
          echo "--- FAIL: ${arg}"
          exit 1
          ;;
      esac
    done
    exit "${FAKE_TEST_EXIT:-0}"
    ;;
esac
exit 0
FAKE
chmod +x "$WORK/bin/go"
export PATH="$WORK/bin:$PATH"

run_runner() {
  : > "$CALLS"
  bash "$RUNNER" "$@"
}

# --- case 1: the defect itself -------------------------------------------------
# A package with several fuzz targets must be fuzzed target by target, and the
# runner must never emit `-fuzz=.` (which matches more than one target and aborts).
export FAKE_PACKAGES="pkg/a"
export FAKE_TARGETS="FuzzAlpha
FuzzBeta
FuzzGamma"
export FAKE_FAIL_TARGET=""
if run_runner 5s ./x/example/... > "$WORK/case1.out" 2>&1; then
  ok "multi-target package exits 0"
else
  bad "multi-target package exited non-zero: $(tail -3 "$WORK/case1.out" | tr '\n' ' ')"
fi

# Fixed-string searches: the recorded call is e.g.
#   -run ^$ -fuzz ^FuzzAlpha$ -fuzztime=5s pkg/a
# so a grep pattern would treat the ^ and $ as anchors and never match.
if grep -qF -- "-fuzz=. " "$CALLS" || grep -qE -- "-fuzz=\.$" "$CALLS"; then
  bad "-fuzz=. was still passed to go test (the original defect)"
else
  ok "never passes a bare -fuzz=. pattern"
fi

for t in FuzzAlpha FuzzBeta FuzzGamma; do
  if grep -qF -- "-fuzz ^${t}\$" "$CALLS"; then
    ok "fuzzed ${t} with a single-target pattern"
  else
    bad "${t} was never fuzzed one-target-at-a-time"
  fi
done

if [ "$(wc -l < "$CALLS")" -eq 3 ]; then
  ok "exactly one go test invocation per fuzz target"
else
  bad "expected 3 go test invocations, saw $(wc -l < "$CALLS")"
fi

if grep -q "^fuzz targets: 3/3 fuzzed clean" "$WORK/case1.out"; then
  ok "reports 3/3 targets fuzzed clean"
else
  bad "did not report 3/3 fuzzed clean: $(tail -2 "$WORK/case1.out" | tr '\n' ' ')"
fi

# --- case 1b: the budget is per step, not per target ----------------------------
# `-fuzztime` used to be the whole step's budget. Applied per target it would be
# multiplied by ~50 targets and blow the 60-minute job timeout on the nightly
# schedule, so the budget must be divided across the targets found.
if run_runner 30s ./x/example/... > "$WORK/case1b.out" 2>&1; then
  if [ "$(grep -cF -- "-fuzztime=10s" "$CALLS")" -eq 3 ]; then
    ok "30s budget over 3 targets is divided to 10s each"
  else
    bad "30s budget was not divided across targets: $(tr '\n' ' ' < "$CALLS")"
  fi
else
  bad "30s budget run failed"
fi

if run_runner 1m30s ./x/example/... > "$WORK/case1c.out" 2>&1; then
  if [ "$(grep -cF -- "-fuzztime=30s" "$CALLS")" -eq 3 ]; then
    ok "1m30s budget parses and divides to 30s each"
  else
    bad "compound duration 1m30s was not parsed: $(tr '\n' ' ' < "$CALLS")"
  fi
else
  bad "1m30s budget run failed: $(tail -2 "$WORK/case1c.out" | tr '\n' ' ')"
fi

# A budget too small to divide still has to fuzz: the floor wins.
if run_runner 4s ./x/example/... > "$WORK/case1d.out" 2>&1; then
  if [ "$(grep -cF -- "-fuzztime=5s" "$CALLS")" -eq 3 ]; then
    ok "a tiny budget floors at 5s per target instead of under-fuzzing"
  else
    bad "floor not applied: $(tr '\n' ' ' < "$CALLS")"
  fi
else
  bad "tiny-budget run failed"
fi

# --- case 2: a crashing target must fail the step -------------------------------
export FAKE_FAIL_TARGET="^FuzzBeta$"
if run_runner 5s ./x/example/... > "$WORK/case2.out" 2>&1; then
  bad "a crashing fuzz target did not fail the step"
else
  ok "a crashing fuzz target fails the step"
fi
if grep -q "::error::fuzz target FuzzBeta failed" "$WORK/case2.out"; then
  ok "names the crashing target"
else
  bad "crash was not attributed to a named target"
fi
if grep -qF -- "-fuzz ^FuzzGamma\$" "$CALLS"; then
  ok "kept fuzzing the remaining targets after a crash"
else
  bad "stopped at the first crashing target"
fi
export FAKE_FAIL_TARGET=""

# --- case 3: a package with no fuzz targets is not a silent pass ----------------
export FAKE_TARGETS=""
if run_runner 5s ./x/empty/... > "$WORK/case3.out" 2>&1; then
  bad "a package with no fuzz target was accepted as green"
else
  ok "a package with no fuzz target fails (decorative step detected)"
fi
if grep -q "this fuzz step is decorative" "$WORK/case3.out"; then
  ok "says out loud that the step was decorative"
else
  bad "did not explain why it failed"
fi

# --- case 4: usage and bad input ---------------------------------------------------
run_runner > "$WORK/case4.out" 2>&1
rc=$?
if [ "$rc" -eq 2 ]; then
  ok "no arguments exits 2 (usage)"
else
  bad "no arguments exited ${rc}, expected 2"
fi

run_runner 5s > "$WORK/case5.out" 2>&1
rc=$?
if [ "$rc" -eq 2 ]; then
  ok "missing package pattern exits 2 (usage)"
else
  bad "missing package pattern exited ${rc}, expected 2"
fi

run_runner 5x ./x/example/... > "$WORK/case6.out" 2>&1
rc=$?
if [ "$rc" -eq 2 ] && grep -q "unrecognised time budget" "$WORK/case6.out"; then
  ok "an unparseable budget is rejected, not silently fuzzed"
else
  bad "unparseable budget exited ${rc} without a diagnosis"
fi

run_runner 0s ./x/example/... > "$WORK/case7.out" 2>&1
rc=$?
if [ "$rc" -eq 2 ]; then
  ok "a zero budget is rejected"
else
  bad "zero budget exited ${rc}, expected 2"
fi

echo
echo "test_run_fuzz_targets.sh: ${PASS} passed, ${FAIL} failed"
[ "$FAIL" -eq 0 ]
