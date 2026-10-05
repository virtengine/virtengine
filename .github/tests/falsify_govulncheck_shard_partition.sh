#!/usr/bin/env bash
# Falsify the shard-partition guard: prove it goes RED on each way the matrix can
# actually break, and GREEN only on the real partition.
#
# A guard that has only ever been seen green is decoration. The failure modes it
# exists to catch are (a) a package that falls in NO shard (silent false
# negative - the dangerous one), (b) a package in TWO shards, and (c) a shard
# declared in the workflow that matches nothing.
#
# THE MUTATIONS ARE APPLIED TO THE WORKFLOW, NOT TO THE TEST.
# The previous version of this falsifier sed-ed its own copy of the guard, which
# is exactly the defect the guard had: the guard carried a private copy of the
# shard matrix, so mutating the guard's copy proved only that the copy was
# self-consistent. Measured on 2026-10-05 (t_76147140) against the pre-fix
# guard: deleting the `tests-and-sim` shard from security.yaml left it printing
# "RESULT: shard partition is EXACT" and exiting 0, and adding a zero-coverage
# shard was equally invisible - so the old falsifier reported PASS on a guard
# that could not see the defect it was written to catch. Each mutation below now
# edits .github/workflows/security.yaml and the UNMODIFIED guard must reject it.
#
# A no-op control is required to pass: a byte-identical rewrite must leave the
# guard GREEN, otherwise "everything is caught" would also be reported by a guard
# that is simply always red. Two further controls matter here: the workflow must
# be restored byte-for-byte afterwards, and an EMPTY matrix (no shards at all)
# must be rejected rather than passing vacuously.
#
# Usage: bash .github/tests/falsify_govulncheck_shard_partition.sh
set -uo pipefail

cd "$(dirname "$0")/../.." || exit 2
TEST=.github/tests/test_govulncheck_shard_partition.sh
WORKFLOW=.github/workflows/security.yaml
MUTPY=.github/tests/_mutate_shard_matrix.py

if [ ! -f "$MUTPY" ]; then
  echo "FATAL: missing matrix mutator: $MUTPY"
  exit 2
fi

tmp="$(mktemp -d)"
restore() { git checkout -- "$WORKFLOW" 2>/dev/null; rm -rf "$tmp"; }
trap restore EXIT

pass=0
fail=0
report() {
  if [ "$1" -eq 0 ]; then
    echo "  PASS  $2"
    pass=$((pass + 1))
  else
    echo "  FAIL  $2"
    fail=$((fail + 1))
  fi
}

# apply <case> <expect_inert:yes|no> : leaves the workflow mutated. Refuses an
# UNEXPECTEDLY inert mutation (no byte change) for the defect cases, so a broken
# anchor can never be credited as a caught defect. The no-op control is the one
# case that is REQUIRED to be inert.
apply() {
  local case="$1"
  local expect_inert="${2:-no}"
  if ! python "$MUTPY" "$case" > "$tmp/mut.log" 2>&1; then
    echo "    mutator error:"; sed 's/^/      /' "$tmp/mut.log" | head -5; return 1
  fi
  if grep -q 'changed=True' "$tmp/mut.log"; then
    if [ "$expect_inert" = "yes" ]; then
      echo "    control case was NOT inert - the 'noop' case changed the file"
      return 1
    fi
    return 0
  fi
  if [ "$expect_inert" = "yes" ]; then
    return 0
  fi
  echo "    INERT mutation (workflow unchanged) - fix the anchor"
  return 1
}

echo "== baseline: the real guard on the real workflow must be GREEN =="
bash "$TEST" > "$tmp/base.out" 2>&1
rc=$?
report "$rc" "unmodified workflow exits 0"
if [ "$rc" -ne 0 ]; then
  sed 's/^/    /' "$tmp/base.out" | tail -10
fi

echo
echo "== M0 CONTROL: a byte-identical rewrite must stay GREEN =="
# The control is REQUIRED to be inert, and REQUIRED to leave the guard green.
# Together they prove "every mutation is caught" is a real measurement and not
# an artifact of a guard that is simply always red.
if apply noop yes; then
  bash "$TEST" > "$tmp/m0.out" 2>&1; rc=$?
  git checkout -- "$WORKFLOW"
  report "$([ "$rc" -eq 0 ] && echo 0 || echo 1)" "no-op rewrite leaves the guard green (rc=$rc)"
else
  report 1 "M0 control could not be applied"
fi

echo
echo "== M1: delete a shard from the workflow (creates UNCOVERED packages) =="
if apply drop-shard; then
  bash "$TEST" > "$tmp/m1.out" 2>&1; rc=$?
  git checkout -- "$WORKFLOW"
  named=$(grep -c "in NO shard" "$tmp/m1.out" || true)
  report "$([ "$rc" -ne 0 ] && [ "$named" -ge 1 ] && echo 0 || echo 1)" \
    "deleted shard reported under COMPLETE (rc=$rc, named=$named)"
  if [ "$named" -lt 1 ]; then sed 's/^/    /' "$tmp/m1.out" | tail -12; fi
else
  report 1 "M1 mutation could not be applied"
fi

echo
echo "== M2: re-pattern a shard so its packages fall into no shard =="
# Drop ./util/... from tests-and-sim: those packages then belong to no shard,
# which is the silent-false-negative class, reached through the WORKFLOW.
if apply drop-util; then
  bash "$TEST" > "$tmp/m2.out" 2>&1; rc=$?
  git checkout -- "$WORKFLOW"
  named=$(grep -c "in NO shard" "$tmp/m2.out" || true)
  report "$([ "$rc" -ne 0 ] && [ "$named" -ge 1 ] && echo 0 || echo 1)" \
    "uncovered package reported under COMPLETE (rc=$rc, named=$named)"
  if [ "$named" -lt 1 ]; then sed 's/^/    /' "$tmp/m2.out" | tail -12; fi
else
  report 1 "M2 mutation could not be applied"
fi

echo
echo "== M3: point a shard's patterns at a directory that does not exist =="
# platform -> ./infra/... : a shard that matches nothing.
if apply repoint-platform; then
  bash "$TEST" > "$tmp/m3.out" 2>&1; rc=$?
  git checkout -- "$WORKFLOW"
  named=$(grep -c "matched 0 packages" "$tmp/m3.out" || true)
  report "$([ "$rc" -ne 0 ] && [ "$named" -ge 1 ] && echo 0 || echo 1)" \
    "empty shard reported (rc=$rc, named=$named)"
  if [ "$named" -lt 1 ]; then sed 's/^/    /' "$tmp/m3.out" | tail -12; fi
else
  report 1 "M3 mutation could not be applied"
fi

echo
echo "== M4: duplicate a directory into a second shard (creates an OVERLAP) =="
# tooling also takes ./util/... , which tests-and-sim already owns.
if apply duplicate-util; then
  bash "$TEST" > "$tmp/m4.out" 2>&1; rc=$?
  git checkout -- "$WORKFLOW"
  named=$(grep -c "more than one shard" "$tmp/m4.out" || true)
  report "$([ "$rc" -ne 0 ] && [ "$named" -ge 1 ] && echo 0 || echo 1)" \
    "overlapping package reported under DISJOINT (rc=$rc, named=$named)"
  if [ "$named" -lt 1 ]; then sed 's/^/    /' "$tmp/m4.out" | tail -12; fi
else
  report 1 "M4 mutation could not be applied"
fi

echo
echo "== M5: empty the include list (no shards declared at all) =="
if apply empty-matrix; then
  bash "$TEST" > "$tmp/m5.out" 2>&1; rc=$?
  git checkout -- "$WORKFLOW"
  # Must NOT be a vacuous pass: 0 shards means nothing is scanned at all.
  report "$([ "$rc" -ne 0 ] && echo 0 || echo 1)" \
    "an empty matrix is rejected, not vacuously green (rc=$rc)"
  if [ "$rc" -eq 0 ]; then sed 's/^/    /' "$tmp/m5.out" | tail -8; fi
else
  report 1 "M5 mutation could not be applied"
fi

echo
echo "== post-run: the workflow must be restored byte-for-byte =="
git checkout -- "$WORKFLOW"
if git diff --quiet -- "$WORKFLOW"; then
  report 0 "workflow restored (git diff clean)"
else
  report 1 "workflow left mutated - the falsifier leaked a change into the repo"
fi

echo
echo "RESULT: $pass passed, $fail failed"
[ "$fail" -eq 0 ] || exit 1