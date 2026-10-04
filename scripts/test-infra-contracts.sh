#!/usr/bin/env bash
# Copyright 2026 VirtEngine contributors.
# SPDX-License-Identifier: Apache-2.0
#
# Run the offline infra/workflow contract tests in the infra/tests module.
#
# Why this exists: infra/tests is a SEPARATE Go module (infra/tests/go.mod), so
# the root module cannot reach it — `go list ./...` from the repo root returns no
# infra package and `go test ./infra/tests/...` fails with "directory prefix
# infra\tests does not contain main module". Until a gate ran it, every test in
# this module passed locally and in CI no matter what it asserted, so a contract
# regression could come back silently. quality-gate.yaml runs this script.
#
# Two deliberate choices:
#
#   * -count=1 — these tests read .github/workflows/*.yaml from disk at run time,
#     so the go test result cache could otherwise serve a stale PASS for an
#     edited workflow. Never trust a cached result from this script.
#
#   * -run with an explicit allowlist — the rest of the package is Terratest
#     against real AWS (see infra/tests/README.md: it creates billable
#     resources and needs credentials). Selecting the offline contract tests
#     keeps this gate cheap and credential-free.
#
# The gate fails CLOSED: each named contract test must actually report PASS, so a
# typo in the -run pattern, a renamed test, or a filter that matches nothing
# exits non-zero instead of green having checked nothing.
set -euo pipefail

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
module="$repo/infra/tests"
log="${VE_INFRA_CONTRACT_LOG:-$(mktemp "${TMPDIR:-/tmp}/infra-contracts.XXXXXX")}"

# Every test the gate exists to run. Keep in sync with the names below.
required=(
  TestTerraformContractsAreFailClosed
  TestWorkflowContractsUseReviewedPlansAndInfraOwnedAutomation
  TestDRToolsRePinBranchIsStable
)
pattern="^($(IFS='|'; echo "${required[*]}"))$"

# No AWS credentials in this gate. The Terratest suite probes for them on every
# test; disabling IMDS keeps that probe from hanging on the runner's link-local
# metadata address.
export AWS_EC2_METADATA_DISABLED=true
export AWS_REGION="${AWS_REGION:-us-east-1}"

cd "$module"

echo "==> go test (offline infra + workflow contracts): $pattern"
# GOWORK=off: the root go.work spans this module, but the gate must test the
# module as committed rather than a workspace-resolved dependency set.
GOWORK=off go test -v -count=1 -timeout 10m -run "$pattern" ./... 2>&1 | tee "$log"

for test_name in "${required[@]}"; do
  # `go test -v` prints "--- PASS: <name> (0.00s)" — the name is followed by a
  # SPACE then the elapsed time. Anchoring on the trailing space (rather than
  # the paren) is what makes the check exact: without it the pattern silently
  # matches nothing and this gate fails on a healthy tree, i.e. it would report
  # red while having verified nothing.
  if ! grep -qE "^--- PASS: ${test_name} " "$log"; then
    echo "::error::${test_name} did not report PASS — the contract gate ran nothing for it" >&2
    echo "::error::check the -run pattern and the test name; a gate that cannot fail is not a gate" >&2
    exit 1
  fi
done

# A reported FAIL or SKIP for any required test is a failure even if go test
# somehow exited 0 (e.g. a future edit that skips instead of asserting).
for test_name in "${required[@]}"; do
  if grep -qE "^--- (FAIL|SKIP): ${test_name} " "$log"; then
    echo "::error::${test_name} reported FAIL/SKIP — the contract gate is not green" >&2
    exit 1
  fi
done

echo "infra + workflow contract tests passed: ${#required[@]}/${#required[@]}"
