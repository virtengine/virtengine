#!/usr/bin/env bash
# Prove the go-vuln-scan matrix shards are a DISJOINT, COMPLETE partition of ./...
#
# Why this exists
# ---------------
# The Go Vulnerability Scan gate shards govulncheck across a matrix so no single
# runner has to build the whole-program call graph (which OOMs a free 7 GB
# runner). Sharding is findings-equivalent to one unsharded scan ONLY IF:
#   (1) the shards are pairwise disjoint, and
#   (2) their union is exactly ./...
# If a package falls in NO shard it is never scanned (silent false negative).
# If it falls in TWO shards it is scanned twice (wasteful, and can mask an
# allowlist divergence between shards).
#
# Both properties are measured here against the REAL package list from
# `go list ./...`, not asserted from reading the YAML. A guard that reformats
# the matrix is not a guard; this one resolves the patterns the workflow uses.
#
# Usage: bash .github/tests/test_govulncheck_shard_partition.sh
# Exit 0 = exact partition; non-zero = the failing property is printed.
set -uo pipefail

cd "$(dirname "$0")/../.." || exit 2

# The shard patterns, in the same order as the workflow matrix. Kept in this
# test on purpose: if the workflow changes, this test must be updated in the
# same commit, and that is a useful coupling (the test IS the spec).
SHARDS=(
  "core:. ./x/... ./app/... ./client/... ./pubsub/..."
  "platform:./pkg/..."
  "tooling:./cmd/... ./scripts/... ./tools/... ./docgen/..."
  "tests-and-sim:./tests/... ./testutil/... ./sim/... ./upgrades/... ./sdk/... ./util/..."
)

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
fail=0

echo "== every package in ./... (go list) =="
if ! go list ./... > "$tmp/all.txt" 2>"$tmp/all.err"; then
  echo "FATAL: 'go list ./...' failed - cannot verify the partition:"
  sed 's/^/  /' "$tmp/all.err" | head -20
  exit 2
fi
total="$(wc -l < "$tmp/all.txt" | tr -d ' ')"
echo "  packages: $total"
if [ "$total" -eq 0 ]; then
  echo "FATAL: go list returned 0 packages - the partition check would pass vacuously"
  exit 2
fi

# Collect each shard's package list.
for entry in "${SHARDS[@]}"; do
  name="${entry%%:*}"
  pats="${entry#*:}"
  # shellcheck disable=SC2086 # word splitting is the point: multiple patterns
  go list $pats > "$tmp/shard-$name.txt" 2>"$tmp/shard-$name.err"
  n="$(wc -l < "$tmp/shard-$name.txt" | tr -d ' ')"
  printf '  shard %-14s %5s packages\n' "$name" "$n"
  if [ "$n" -eq 0 ]; then
    echo "FAIL: shard '$name' matched 0 packages (pattern typo? a v0 package is not scanned)"
    sed 's/^/  /' "$tmp/shard-$name.err" | head -5
    fail=1
  fi
done

echo
echo "== property 1: COMPLETE (union == ./...) =="
cat "$tmp"/shard-*.txt | sort -u > "$tmp/union.txt"
comm -23 "$tmp/all.txt" "$tmp/union.txt" > "$tmp/missing.txt" || true
missing="$(wc -l < "$tmp/missing.txt" | tr -d ' ')"
if [ "$missing" -ne 0 ]; then
  echo "FAIL: $missing package(s) are in NO shard - they would never be scanned:"
  sed 's/^/  /' "$tmp/missing.txt" | head -30
  fail=1
else
  echo "  OK - all $total packages are covered by at least one shard"
fi

echo
echo "== property 2: disjoint (no package in two shards) =="
cat "$tmp"/shard-*.txt | sort | uniq -d > "$tmp/dup.txt"
dups="$(wc -l < "$tmp/dup.txt" | tr -d ' ')"
if [ "$dups" -ne 0 ]; then
  echo "FAIL: $dups package(s) appear in more than one shard:"
  sed 's/^/  /' "$tmp/dup.txt" | head -30
  fail=1
else
  echo "  OK - no package is scanned by two shards"
fi

echo
if [ "$fail" -eq 0 ]; then
  echo "RESULT: shard partition is EXACT (complete + disjoint) over $total packages"
else
  echo "RESULT: shard partition is BROKEN - the matrix does not cover ./... exactly"
fi
exit "$fail"
