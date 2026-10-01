#!/usr/bin/env bash
# Falsify the PR-side gosec stdout contract: prove its guard goes RED when the
# contract is broken, and GREEN only on the real bytes.
#
# Why this exists
# ---------------
# `.github/scripts/gosec_pr_targets.py` has been the subject of TWO false-red +
# false-negative defects in the same step, both invisible to a green run:
#
#   1. gosec was handed changed .go FILE PATHS, which it resolves as package
#      patterns -> imports 0 files -> exits 1. (fixed by #1114)
#   2. the `note: ...` summary went to STDOUT, which the workflow consumes as the
#      gosec argument list -> gosec was handed a sentence as a path, skipped it,
#      imported 0 files, exited 0, and the fail-closed `Stats.files=0` guard
#      reported a red that was not a security finding. Any PR touching only
#      _test.go / .pb.go / vendor files could not go green. (fixed by #1115)
#
# The guard `test_pr_gosec_scan_targets.py` pins this by RUNNING the script, so it
# is a real end-to-end check rather than a source reading. But a guard that has
# only ever been seen green is decoration, so this harness removes the fix and
# REQUIRES the guard to fail - for the right reason, not by crashing.
#
# The mutation is applied to a COPY. The live script is never written, so this is
# safe to run against a dirty tree and cannot leave a mangled file behind.
#
# Usage: bash .github/tests/falsify_pr_gosec_stdout_contract.sh
set -uo pipefail

cd "$(dirname "$0")/../.." || exit 2

SCRIPT=.github/scripts/gosec_pr_targets.py
GUARD=.github/tests/test_pr_gosec_scan_targets.py
# A MSYS /tmp path is not resolvable by a NATIVE Windows python.exe ("\\tmp\\..."),
# so the temp dir is forced under the repo, where both the shell and python agree
# on the spelling. Measured, not assumed: mktemp -d alone dies with FileNotFoundError.
WORK="$(mktemp -d ./.falsify-gosec-XXXXXX)"
cleanup() { rm -rf "$WORK"; }
trap cleanup EXIT

pass=0
fail=0
ok()   { echo "  PASS  $1"; pass=$((pass + 1)); }
bad()  { echo "  FAIL  $1"; fail=$((fail + 1)); }

# ---------------------------------------------------------------------------
# M1: the note goes back to stdout (the exact #1115 defect)
# ---------------------------------------------------------------------------
mutate_m1() {
  mkdir -p "$WORK/m1"
  cp -a .github "$WORK/m1/.github"
  cp -a "$SCRIPT" "$WORK/m1/$SCRIPT"
  # Capture the status of the heredoc itself. A bare `[ $? -eq 0 ]` after it reads
  # the status of the `cp`/pipeline, not the python that ran, so a failed mutation
  # was reported as a pass. Assign-then-test, or the check is a lie.
  if ! python3 - "$WORK/m1/$SCRIPT" <<'PY'
import sys
from pathlib import Path
p = Path(sys.argv[1])
b = p.read_bytes()
fix = (b'f"(tests, generated .pb.go, vendor/testdata), e.g. {skipped[0]}",\n'
       b'              file=sys.stderr)')
broken = b'f"(tests, generated .pb.go, vendor/testdata), e.g. {skipped[0]}")'
if fix not in b:
    sys.exit(3)
# Bytes, not text: a text write would rewrite the file's line endings.
p.write_bytes(b.replace(fix, broken))
PY
  then
    return 3
  fi
  # Prove the copy really was mutated, so a silent no-op replace cannot pass.
  # Grepping for 'file=sys.stderr' would be WRONG: the script has six legitimate
  # stderr diagnostics, so that probe matches whether or not the mutation landed.
  # Pin the mutated note line itself instead.
  if grep -q 'e.g. {skipped\[0\]}",$' "$WORK/m1/$SCRIPT"; then
    echo "mutate_m1: the note still ends with the stderr redirection - no-op replace" >&2
    return 4
  fi
  return 0
}

run_guard_in() {
  # $1 = tree root containing a mutated copy. Runs THAT copy's test, which derives
  # its own REPO/SCRIPT from __file__, so it exercises the mutant and not the live
  # file. The path is built from the root only - interpolating $GUARD (which is
  # repo-relative) onto $1 doubled it and python could not open it.
  ( cd "$1" && python3 ".github/tests/test_pr_gosec_scan_targets.py" ) >"$WORK/guard.out" 2>&1
  echo $?
}

echo "== falsify_pr_gosec_stdout_contract: the PR gosec guard must be able to go red =="

# --- Baseline: on the real tree the guard must be GREEN. ---------------------
rc="$( cd . && python3 "$GUARD" >"$WORK/base.out" 2>&1; echo $? )"
if [ "$rc" -eq 0 ]; then
  ok "baseline: guard is green on the real tree (rc=0)"
else
  bad "baseline: guard is RED on the real tree (rc=$rc) - fix that first; a falsifier
       run against a broken baseline proves nothing"
  sed 's/^/       /' "$WORK/base.out" | tail -5
fi

# --- M1: the note is on stdout again. The guard MUST fail. -------------------
if ! mutate_m1; then
  bad "M1 could not be constructed - the stderr fix is not in these bytes"
else
  rc="$(run_guard_in "$WORK/m1")"
  if [ "$rc" -ne 0 ]; then
    ok "M1 (note on stdout) -> guard FAILS as required (rc=$rc)"
  else
    bad "M1 (note on stdout) -> guard still passed: the stdout contract is DECORATION,
         a regression of #1115 would ship silently"
  fi
  # It must fail for the RIGHT reason: the stdout assertions, not an import error.
  if grep -q "stdout is empty when 0 packages resolved" "$WORK/guard.out"; then
    ok "M1 failed on the stdout assertion, not on an unrelated error"
  else
    bad "M1 went red for the wrong reason - expected the stdout property to fail"
    sed 's/^/       /' "$WORK/guard.out" | tail -6
  fi
fi

# --- The live script must be untouched by this harness. -----------------------
if git diff --quiet -- "$SCRIPT" 2>/dev/null; then
  ok "live $SCRIPT is byte-identical after the run (mutations touched a copy)"
else
  bad "live $SCRIPT was modified by the falsifier - the mutation must not touch it"
fi

echo
if [ "$fail" -ne 0 ]; then
  echo "RESULT: $fail property(ies) FAILED - the guard cannot be proven real"
  exit 1
fi
echo "RESULT: the guard goes red on the stdout-contract regression and green on the real bytes"
exit 0
