#!/usr/bin/env bash
# Prove the go-vuln-scan matrix shards are a DISJOINT, COMPLETE partition of ./...
# AND that every shard declared in the WORKFLOW actually matches a package.
#
# Why this exists
# ---------------
# The Go Vulnerability Scan gate shards govulncheck across a matrix so no single
# runner has to build the whole-program call graph (which OOMs a free 7 GB
# runner). Sharding is findings-equivalent to one unsharded scan ONLY IF:
#   (1) the shards are pairwise disjoint,
#   (2) their union is exactly ./... , and
#   (3) every shard declared in the WORKFLOW matches at least one package.
# A package in NO shard is never scanned (silent false negative). A package in
# TWO shards is scanned twice. A shard matching nothing is a typo that burns a
# runner and reports no coverage at all.
#
# THE SHARD LIST IS READ FROM THE WORKFLOW, NOT COPIED HERE.
# This guard used to keep its own `SHARDS=(...)` array, so it verified that copy
# and could not observe the matrix it exists to protect. Measured on 2026-10-05
# (t_76147140): deleting the `tests-and-sim` shard from
# `.github/workflows/security.yaml` left this guard printing "RESULT: shard
# partition is EXACT" and exiting 0; adding a shard whose patterns match zero
# packages was equally invisible, which made the "matched 0 packages" check
# below unreachable for any shard the workflow actually declares. Both defects
# are now covered by .github/tests/falsify_govulncheck_shard_partition.sh, which
# mutates the WORKFLOW rather than mutating this test.
#
# All properties are measured against the REAL package list from `go list ./...`,
# not asserted from reading the YAML. A guard that reformats the matrix is not a
# guard; this one resolves the patterns the workflow uses.
#
# Usage: bash .github/tests/test_govulncheck_shard_partition.sh
# Exit 0 = exact partition; non-zero = the failing property is printed.
set -uo pipefail

cd "$(dirname "$0")/../.." || exit 2

WORKFLOW=".github/workflows/security.yaml"
READER=".github/scripts/govulncheck_shards.py"

if [ ! -f "$READER" ]; then
  echo "FATAL: missing shard reader: $READER"
  exit 2
fi

# Read the matrix out of the workflow. A reader that cannot resolve it is FATAL,
# never a pass: an unreadable matrix would otherwise satisfy every property
# below vacuously.
if ! shards_out="$(python "$READER" --workflow "$WORKFLOW" --list 2> .shards-read.err)"; then
  echo "FATAL: could not read the shard matrix from $WORKFLOW:"
  sed 's/^/  /' .shards-read.err | head -10
  rm -f .shards-read.err
  exit 2
fi

n_shards="$(printf '%s\n' "$shards_out" | grep -c '[^[:space:]]')"
if [ "$n_shards" -eq 0 ]; then
  echo "FATAL: the reader returned 0 shards - the partition check would pass vacuously"
  rm -f .shards-read.err
  exit 2
fi
echo "  shards declared in $WORKFLOW: $n_shards"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"; rm -f .shards-read.err' EXIT
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

# Collect each shard's package list. Patterns come from the workflow, so a shard
# added, removed or re-patterned there is measured on the very next run.
# Strip a trailing CR first: this repo's checkout carries CRLF endings, and a
# pattern left with a "\r" matches nothing, which would make every assertion
# below measure the line ending instead of the matrix.
while IFS="$(printf '\t')" read -r name pats; do
  [ -n "$name" ] || continue
  pats="${pats%$'\r'}"
  # Strip the surrounding YAML quotes the reader preserves verbatim.
  pats="${pats%\"}"
  pats="${pats#\"}"
  # shellcheck disable=SC2086 # word splitting is the point: multiple patterns
  go list $pats > "$tmp/shard-$name.txt" 2>"$tmp/shard-$name.err"
  n="$(wc -l < "$tmp/shard-$name.txt" | tr -d ' ')"
  printf '  shard %-14s %5s packages\n' "$name" "$n"
  if [ "$n" -eq 0 ]; then
    echo "FAIL: workflow shard '$name' matched 0 packages (pattern typo? a v0 package is not scanned)"
    sed 's/^/  /' "$tmp/shard-$name.err" | head -5
    fail=1
  fi
done <<EOF
$shards_out
EOF

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
  echo "RESULT: shard partition is EXACT (complete + disjoint; all $n_shards workflow shards non-empty) over $total packages"
else
  echo "RESULT: shard partition is BROKEN - the matrix does not cover ./... exactly"
fi
exit "$fail"
