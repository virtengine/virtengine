#!/usr/bin/env bash
# Reproduce the static-analysis portion of the VirtEngine audited security suite.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
REPORT_DIR="${ROOT_DIR}/tests/security/reports"
mkdir -p "${REPORT_DIR}"
: "${GOTOOLCHAIN:=go1.25.9+auto}"
export GOTOOLCHAIN

DATE="$(date -u +%Y%m%d_%H%M%S)"
REPORT_FILE="${REPORT_DIR}/static_analysis_${DATE}.txt"

exec > >(tee "${REPORT_FILE}") 2>&1

echo "========================================"
echo "VirtEngine Static Security Analysis"
echo "========================================"
echo "repo: ${ROOT_DIR}"
echo "report: ${REPORT_FILE}"
echo "go_toolchain: ${GOTOOLCHAIN}"
echo

cd "${ROOT_DIR}"

echo "=== gosec (security suite) ==="
go run github.com/securego/gosec/v2/cmd/gosec@v2.25.0 ./tests/security/...
echo

echo "=== go vet (security suite) ==="
go vet ./tests/security/...
echo

echo "static analysis complete"
