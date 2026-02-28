#!/usr/bin/env bash
# Reproduce the dependency-focused portion of the VirtEngine audited security suite.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
REPORT_DIR="${ROOT_DIR}/tests/security/reports"
mkdir -p "${REPORT_DIR}"
: "${GOTOOLCHAIN:=go1.25.9+auto}"
export GOTOOLCHAIN

DATE="$(date -u +%Y%m%d_%H%M%S)"
REPORT_FILE="${REPORT_DIR}/dependency_scan_${DATE}.txt"

exec > >(tee "${REPORT_FILE}") 2>&1

echo "========================================"
echo "VirtEngine Dependency Security Scan"
echo "========================================"
echo "repo: ${ROOT_DIR}"
echo "report: ${REPORT_FILE}"
echo "go_toolchain: ${GOTOOLCHAIN}"
echo

cd "${ROOT_DIR}"

echo "=== govulncheck (security suite) ==="
go run golang.org/x/vuln/cmd/govulncheck@v1.1.4 ./tests/security/...
echo

echo "=== repository dependency risk assessment ==="
go run ./scripts/supply-chain/assess-dependencies.go --report --json
echo

echo "dependency scan complete"
