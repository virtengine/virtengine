#!/usr/bin/env bash
# Fail closed unless every required security gate completed successfully.
# The workflow passes one result per item in security-summary.needs.
set -uo pipefail

expected=9
if [ "$#" -ne "$expected" ]; then
  echo "::error::Expected $expected security-gate results, got $#"
  exit 1
fi

failed=0
index=0
for result in "$@"; do
  index=$((index + 1))
  if [ "$result" != "success" ]; then
    echo "::error::Required security gate $index returned '$result' (expected success)"
    failed=1
  fi
done

if [ "$failed" -ne 0 ]; then
  exit 1
fi

echo "All $expected required security gates succeeded"
