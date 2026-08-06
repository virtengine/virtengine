#!/usr/bin/env bash
# Reproduce the secret-detection portion of the VirtEngine audited security suite.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
REPORT_DIR="${ROOT_DIR}/tests/security/reports"
mkdir -p "${REPORT_DIR}"
: "${GOTOOLCHAIN:=go1.25.9+auto}"
export GOTOOLCHAIN

DATE="$(date -u +%Y%m%d_%H%M%S)"
REPORT_FILE="${REPORT_DIR}/secret_scan_${DATE}.txt"

GITLEAKS_BIN="${GITLEAKS_BIN:-}"
GITLEAKS_VERSION="${GITLEAKS_VERSION:-v8.30.1}"

exec > >(tee "${REPORT_FILE}") 2>&1

echo "========================================"
echo "VirtEngine Secret Detection Scan"
echo "========================================"
echo "repo: ${ROOT_DIR}"
echo "report: ${REPORT_FILE}"
echo "go_toolchain: ${GOTOOLCHAIN}"
echo

cd "${ROOT_DIR}"

if [[ -n "${GITLEAKS_BIN}" ]]; then
	if [[ -x "${GITLEAKS_BIN}" ]]; then
		GITLEAKS_CMD=("${GITLEAKS_BIN}")
	elif command -v "${GITLEAKS_BIN}" >/dev/null 2>&1; then
		GITLEAKS_CMD=("${GITLEAKS_BIN}")
	else
		echo "gitleaks binary not found at ${GITLEAKS_BIN}" >&2
		exit 1
	fi
	GITLEAKS_SOURCE="${GITLEAKS_BIN}"
elif command -v gitleaks >/dev/null 2>&1; then
	GITLEAKS_CMD=(gitleaks)
	GITLEAKS_SOURCE="gitleaks"
else
	GITLEAKS_CMD=(go run "github.com/zricethezav/gitleaks/v8@${GITLEAKS_VERSION}")
	GITLEAKS_SOURCE="go run github.com/zricethezav/gitleaks/v8@${GITLEAKS_VERSION}"
fi

echo "=== gitleaks current-tree scan ==="
echo "scanner: ${GITLEAKS_SOURCE}"
"${GITLEAKS_CMD[@]}" dir . --config .gitleaks.toml --redact --report-format json --report-path "${REPORT_DIR}/gitleaks_${DATE}.json"
echo

echo "secret scan complete"
