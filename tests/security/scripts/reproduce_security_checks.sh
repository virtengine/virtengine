#!/usr/bin/env bash
# Run the test-backed local reproduction path described in SECURITY.md.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)"
MODE="${1:-full}"

cd "${ROOT_DIR}"

run_unit() {
	echo "=== security unit coverage ==="
	go test -tags=security ./tests/security/...
}

run_integration() {
	echo "=== security integration coverage ==="
	go test -tags='security,integration' ./tests/security/...
}

run_e2e() {
	echo "=== security e2e coverage ==="
	go test -tags='security,e2e.integration' ./tests/security/...
}

run_static() {
	echo "=== static analysis ==="
	bash ./tests/security/scripts/static_analysis.sh
}

run_deps() {
	echo "=== dependency scanning ==="
	bash ./tests/security/scripts/scan_dependencies.sh
}

run_secrets() {
	echo "=== secret scanning ==="
	bash ./tests/security/scripts/secret_scan.sh
}

case "${MODE}" in
	unit)
		run_unit
		;;
	integration)
		run_integration
		;;
	e2e)
		run_e2e
		;;
	tests)
		run_unit
		run_integration
		run_e2e
		;;
	full)
		run_unit
		run_integration
		run_e2e
		run_static
		run_deps
		run_secrets
		;;
	*)
		echo "usage: $0 {unit|integration|e2e|tests|full}" >&2
		exit 2
		;;
esac
