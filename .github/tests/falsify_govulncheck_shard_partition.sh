#!/usr/bin/env bash
# Falsify the shard-partition guard: prove it goes RED on each way the matrix can
# actually break, and GREEN only on the real partition.
#
# A guard that has only ever been seen green is decoration. The two failure modes
# it exists to catch are (a) a package that falls in NO shard (silent false
# negative - the dangerous one) and (b) a package in TWO shards. Both are injected
# here into copies of the test, and the test must fail for the RIGHT reason.
#
# Usage: bash .github/tests/falsify_govulncheck_shard_partition.sh
set -uo pipefail

cd "$(dirname "$0")/../.." || exit 2
TEST=.github/tests/test_govulncheck_shard_partition.sh
MUTDIR=.github/tests
tmp="$(mktemp -d)"
cleanup() { rm -rf "$tmp"; rm -f "$MUTDIR"/_mut_*.sh; }
trap cleanup EXIT

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

# Apply a sed expression; require the result to DIFFER from the original.
# Two harness traps this closes, both of which would make this falsifier
# decoration instead of evidence:
#   * the SHARDS entries are "name:patterns", so the pattern string has no
#     leading quote - a sed anchored on one matches nothing and the "mutation"
#     is a no-op. `cmp -s` catches that.
#   * the mutated copy must live INSIDE the repo: the test under mutation does
#     its own `cd "$(dirname "$0")/../.."`, so a copy in $TMPDIR fails `go list`
#     and exits 2 - a red for the wrong reason, crediting nothing.
mutate() { # $1=tag $2=sed-expr ; echoes the mutated path, or nothing if inert
  local tag="$1"
  local expr="$2"
  local path="$MUTDIR/_mut_$tag.sh"
  sed "$expr" "$TEST" > "$path"
  if cmp -s "$TEST" "$path"; then
    rm -f "$path"
    return 1
  fi
  echo "$path"
}

echo "== baseline: the real test must be GREEN =="
bash "$TEST" > "$tmp/base.out" 2>&1
rc=$?
report "$rc" "unmodified test exits 0"
if [ "$rc" -ne 0 ]; then
  sed 's/^/    /' "$tmp/base.out" | tail -10
fi

echo
echo "== M1: drop a directory from its shard (creates an UNCOVERED package) =="
# Remove ./util/... from tests-and-sim. Those packages then belong to no shard,
# which is the silent-false-negative class. The test must go RED under COMPLETE.
m1="$(mutate m1 's|tests-and-sim:\./tests/\.\.\. \./testutil/\.\.\. \./sim/\.\.\. \./upgrades/\.\.\. \./sdk/\.\.\. \./util/\.\.\.|tests-and-sim:./tests/... ./testutil/... ./sim/... ./upgrades/... ./sdk/...|')"
if [ -z "$m1" ]; then
  report 1 "M1 mutation changed the file (inert mutation - fix the anchor)"
else
  bash "$m1" > "$tmp/m1.out" 2>&1; rc=$?
  rm -f "$m1"
  named=$(grep -c "in NO shard" "$tmp/m1.out" || true)
  report "$([ "$rc" -ne 0 ] && [ "$named" -ge 1 ] && echo 0 || echo 1)" \
    "uncovered package reported under COMPLETE (rc=$rc, named=$named)"
fi

echo
echo "== M2: duplicate a directory into a second shard (creates an OVERLAP) =="
m2="$(mutate m2 's|tooling:\./cmd/\.\.\. \./scripts/\.\.\. \./tools/\.\.\. \./docgen/\.\.\.|tooling:./cmd/... ./scripts/... ./tools/... ./docgen/... ./util/...|')"
if [ -z "$m2" ]; then
  report 1 "M2 mutation changed the file (inert mutation - fix the anchor)"
else
  bash "$m2" > "$tmp/m2.out" 2>&1; rc=$?
  rm -f "$m2"
  named=$(grep -c "more than one shard" "$tmp/m2.out" || true)
  report "$([ "$rc" -ne 0 ] && [ "$named" -ge 1 ] && echo 0 || echo 1)" \
    "overlapping package reported under DISJOINT (rc=$rc, named=$named)"
fi

echo
echo "== M3: a shard pattern that matches nothing (the defect this guard already caught) =="
m3="$(mutate m3 's|"platform:\./pkg/\.\.\."|"platform:./infra/..."|')"
if [ -z "$m3" ]; then
  report 1 "M3 mutation changed the file (inert mutation - fix the anchor)"
else
  bash "$m3" > "$tmp/m3.out" 2>&1; rc=$?
  rm -f "$m3"
  named=$(grep -c "matched 0 packages" "$tmp/m3.out" || true)
  report "$([ "$rc" -ne 0 ] && [ "$named" -ge 1 ] && echo 0 || echo 1)" \
    "empty shard reported (rc=$rc, named=$named)"
fi

echo
echo "RESULT: $pass passed, $fail failed"
[ "$fail" -eq 0 ] || exit 1
