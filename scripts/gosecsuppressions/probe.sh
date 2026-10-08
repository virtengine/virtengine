#!/usr/bin/env bash
# PROBE: proves the guard in scripts/gosecsuppressions rejects the trap and
# accepts the fix, and that its ratchet cannot be defeated by trading one
# unreadable annotation for another.
#
# It writes real trees, runs the real binary, and asserts real exit codes. It
# does not import the guard's functions, because the claim under test is what the
# CI STEP sees -- the process exit code -- not what a unit test in the same
# package sees.
#
# Usage: bash scripts/gosecsuppressions/probe.sh
# Exit:  0 all assertions hold, 1 at least one assertion failed.
set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

failures=0
bin="$WORK/guard"

# Two host quirks are handled here, because both made an earlier run of this
# probe report 7 phantom failures against a build that had actually succeeded:
#
#   - `go build -o guard` writes EXACTLY the name given. It does not append
#     `.exe`, so the result is an extensionless Windows PE, and MSYS does not
#     report such a file as executable, so `[ -x ]` is the wrong existence test.
#   - native `go` cannot write through an MSYS `/tmp/...` path, so the -o target
#     is handed over in native (cygpath -m) form and invoked via the MSYS path.
build_out="$(cd "$REPO_ROOT/scripts/gosecsuppressions" && GOWORK=off go build -o "$(cygpath -m "$WORK/guard")" . 2>&1)" && built=1 || built=0
if [ "$built" -ne 1 ]; then
  echo "FATAL: could not build the guard" >&2
  echo "$build_out" >&2
  exit 1
fi
if [ ! -f "$bin" ]; then
  echo "FATAL: built the guard but no binary appeared at $bin" >&2
  exit 1
fi
if ! "$bin" -h >/dev/null 2>&1 && [ $? -ne 2 ]; then
  echo "FATAL: the binary at $bin exists but is not executable by this shell" >&2
  exit 1
fi
echo "built: $bin"

# Every probe tree grades against ONE baseline name, and both helpers use it.
# An earlier version recorded to "recorded.baseline" while the grade step used
# the guard's DEFAULT path, so four cases failed on "missing baseline" instead of
# on the behaviour they were written to test -- a failure that looks like a guard
# defect and is not one.
BASELINE_NAME="probe.baseline"

# run <tree> [args...] -> prints the exit code, keeps output in $WORK/{out,err}
#
# The guard is a NATIVE binary, so every path handed to it is converted to native
# form. An MSYS /tmp/... path makes it exit 2 with "GetFileAttributesEx ... The
# system cannot find the path specified" -- which is indistinguishable from the
# real exit-2 conditions unless the message is asserted too. That bit us: cases
# 5 and 6 originally "passed" for that wrong reason, so assert_msg exists and
# every exit-2 assertion names the message it expects.
run() {
  local tree
  tree="$(cygpath -m "$1")"
  shift
  "$bin" -root "$tree" -baseline "$BASELINE_NAME" "$@" >"$WORK/out" 2>"$WORK/err"
  echo $?
}

# assert_msg <what> <needle> -- a required substring in the combined output.
assert_msg() {
  local name="$1" needle="$2"
  if cat "$WORK/out" "$WORK/err" | grep -q -- "$needle"; then
    echo "  PASS  $name"
  else
    echo "  FAIL  $name: output did not contain: $needle"
    cat "$WORK/out" "$WORK/err" | sed 's/^/        | /' | head -10
    failures=$((failures + 1))
  fi
}

# A tree holding one already-recorded (accepted) annotation plus whatever the
# probe adds, so the case under test is graded against a NON-empty baseline --
# otherwise "no baseline at all" would be a trivially different situation.
base_tree() {
  local dir="$1"
  mkdir -p "$dir"
  printf 'package p\n\nfunc recorded() int {\n\t//nolint:gosec // G115: pre-existing, already in the baseline\n\treturn 1\n}\n' > "$dir/recorded.go"
}

# baseline_for <tree> -> records the tree's current state as its own baseline.
# The -baseline path is relative and resolves against -root, so this must hand
# over the same native path `run` does -- otherwise the recording silently fails
# and every later case grades against a baseline that was never written.
baseline_for() {
  local tree
  tree="$(cygpath -m "$1")"
  "$bin" -root "$tree" -baseline "$BASELINE_NAME" -write-baseline >"$WORK/baseline-out" 2>"$WORK/baseline-err"
  local rc=$?
  if [ "$rc" -ne 0 ]; then
    echo "  FATAL: could not record a baseline for $tree (exit $rc)"
    cat "$WORK/baseline-err" | sed 's/^/        | /'
    exit 1
  fi
}

expect() {
  local name="$1" want="$2" got="$3"
  if [ "$got" = "$want" ]; then
    echo "  PASS  $name (exit $got)"
  else
    echo "  FAIL  $name: expected exit $want, got $got"
    sed 's/^/        | /' "$WORK/out" "$WORK/err" | head -20
    failures=$((failures + 1))
  fi
}

echo
echo "1. THE TRAP: a new file shipping //nolint:gosec-only suppressions"
t="$WORK/t1"; base_tree "$t"; baseline_for "$t"
cat > "$t/newfile.go" <<'EOF'
package p

func ship() int {
	//nolint:gosec // G115: looks justified, invisible to the gate
	return 1
}
EOF
expect "new //nolint-only annotation FAILS the guard" 1 "$(run "$t")"
grep -q "gosec-triage.md section 9.5" "$WORK/err" \
  && echo "  PASS  the documented error message is present" \
  || { echo "  FAIL  error message does not point at the documented fix"; failures=$((failures + 1)); }

echo
echo "2. THE FIX: the same file, #nosec first, passes"
cat > "$t/newfile.go" <<'EOF'
package p

func ship() int {
	// #nosec G115 -- bounded by the clamp above
	return 1
}
EOF
expect "the readable form PASSES" 0 "$(run "$t")"

echo
echo "3. THE TRADE a bare count ratchet cannot catch: ship one, delete one"
t="$WORK/t3"; base_tree "$t"; baseline_for "$t"
printf 'package p\n\nfunc ship() int {\n\t//nolint:gosec // G115: brand new\n\treturn 1\n}\n' > "$t/newfile.go"
printf 'package p\n' > "$t/recorded.go"   # remove one recorded annotation
out="$(run "$t")"
expect "count is unchanged (370-style trap) yet the guard FAILS" 1 "$out"
if grep -q "recorded baseline" "$WORK/out"; then
  echo "  PASS  the failure names the baseline rule, not a count"
else
  echo "  FAIL  failure did not name the rule"
  failures=$((failures + 1))
fi

echo
echo "4. THE IMPROVEMENT: deleting an annotation passes and is reported"
t="$WORK/t4"; base_tree "$t"; baseline_for "$t"
printf 'package p\n' > "$t/recorded.go"
expect "an improvement PASSES" 0 "$(run "$t")"
grep -q "improved:" "$WORK/out" \
  && echo "  PASS  the improvement is reported, not silently absorbed" \
  || { echo "  FAIL  improvement was not reported"; failures=$((failures + 1)); }

echo
echo "5. FAIL-CLOSED: a missing baseline is NOT a pass"
t="$WORK/t5"; base_tree "$t"
expect "missing baseline exits 2, not 0" 2 "$(run "$t")"
assert_msg "and it fails for the RIGHT reason (unreadable baseline)" "read baseline"
assert_msg "and it explains grading against no record would pass everything" "would pass everything"

echo
echo "5b. FAIL-CLOSED: an EMPTY baseline is NOT a pass either"
# A different defect from a missing file: an empty record would grade every
# annotation as new debt (safe), but if the emptiness check were ever dropped it
# would also make a 0-annotation tree look recorded. Asserted separately because
# the message differs, and asserting the wrong message is what made this case
# vacuous on an earlier run.
t="$WORK/t5b"; base_tree "$t"
: > "$t/$BASELINE_NAME"
expect "empty baseline exits 2, not 0" 2 "$(run "$t")"
assert_msg "and it names the empty record specifically" "the record is empty, not a record of zero"
# It must ALSO say the tree is not empty, so the failure is diagnosable from
# the log instead of reading as an unexplained refusal.
assert_msg "and it reports how much the tree actually holds" "unreadable annotation(s)"

echo
echo "6. FAIL-CLOSED: an unparseable tree is NOT a pass"
t="$WORK/t6"; base_tree "$t"; baseline_for "$t"
printf 'package p\n\nfunc f( { this is not go\n' > "$t/broken.go"
expect "parse error exits 2, not 0" 2 "$(run "$t")"
assert_msg "and it fails for the RIGHT reason (parse error)" "broken.go"

echo
echo "7. THE OVERRIDE is one-sided: it cannot be raised to buy headroom"
t="$WORK/t7"; base_tree "$t"; baseline_for "$t"
printf 'package p\n\nfunc ship() int {\n\t//nolint:gosec // G115: new\n\treturn 1\n}\n' > "$t/newfile.go"
code="$(GOSEC_SUPPRESSION_MAX=99 run "$t")"
expect "GOSEC_SUPPRESSION_MAX=99 does not buy an allowance" 1 "$code"
grep -q "REFUSED" "$WORK/err" \
  && echo "  PASS  the refusal is explicit in the log" \
  || { echo "  FAIL  the refusal was silent"; failures=$((failures + 1)); }

echo
if [ "$failures" -eq 0 ]; then
  echo "PROBE PASSED: 7/7 groups"
  exit 0
fi
echo "PROBE FAILED: $failures assertion(s)" >&2
exit 1