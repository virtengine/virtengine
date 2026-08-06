#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT_DIR"

DRY_RUN=0
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=1
fi

run_cmd() {
  echo "+ $*"
  if [[ "$DRY_RUN" -eq 0 ]]; then
    "$@"
  fi
}

INTEGRATION_PATTERN='Test(NodeRegistrationAndPruning|JobLifecycleSubmitScheduleRunCompleteSettle|Accounting|Billing|UsageSnapshotValidation|WorkloadTemplate|AdapterIntegration|SLURMDeploymentKind)'
E2E_PATTERN='Test(HPCJobLifecycleE2E|HPCFullLifecycleTestSuite)'
E2E_FILES=(
  tests/e2e/hpc_telemetry_helpers_test.go
  tests/e2e/hpc_lifecycle_test.go
  tests/e2e/hpc_full_lifecycle_test.go
  tests/e2e/hpc_billing_test.go
)

run_cmd go test -tags e2e.integration ./tests/integration/hpc -run "$INTEGRATION_PATTERN" -count=1
run_cmd go test -tags e2e.integration "${E2E_FILES[@]}" -run "$E2E_PATTERN" -count=1

if [[ "${VE_RUN_KIND_HARNESS:-0}" == "1" ]]; then
  run_cmd go test -tags e2e.integration ./tests/integration/hpc -run 'TestSLURMDeploymentKind' -count=1
else
  echo "Skipping TestSLURMDeploymentKind; set VE_RUN_KIND_HARNESS=1 to run the real kind/helm cluster harness."
fi
