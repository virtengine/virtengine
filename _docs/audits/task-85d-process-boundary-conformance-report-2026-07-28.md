# Task 85D Process-Boundary Conformance Report

Date: 2026-07-28
Status: `engineering_complete_external_blocked`

This report records local engineering conformance for Task 85D only. It does not claim live-chain, real production IdP, production CA, production HSM/KMS, production model/runtime/dataset, or retained production app-hash evidence.

## Scope

- Hardened helper-process protocols for web evidence, deterministic inference receipts, and keeper-level consensus receipts.
- Strengthened web evidence stored-record assertions for verified status, exact evidence pointers, privacy-safe metadata, issuer lineage, score history, and wallet score lineage.
- Changed the deterministic inference fixture to derive receipt inputs, features, score, output digest, and lineage digest from an explicit bounded golden input.
- Changed the four-validator consensus harness so a child process owns the Ed25519 receipt private key and returns only public registration data plus canonical signed receipt bytes.
- Modernized old VEID integration flow tests so disabled ordinary `UpdateVerificationStatus` and `UpdateScore` messages are asserted as rejected, then continuation state is seeded directly as a fixture without weakening production.
- Hardened `scripts/task85d-preflight.ps1` for generated-contract applicability, task-scoped lint/diff checks, PowerShell parse checks, named diagnostic skips, Git Bash expectations, and WSL fail-closed behavior.

## Engineering Evidence Classification

The local web, inference, and consensus child processes are deterministic engineering fixtures. They prove process-boundary behavior, private-key custody in a separate OS process, canonical sign bytes, committed keeper state transitions, idempotent retries, strict receipt bytes, real vote-extension production/verification, and strict quorum handling. They are not external-provider certification.

## RED Evidence

Command:

```text
$env:GOCACHE=(Resolve-Path .cache).Path + '\go-build'; go test ./tests/integration/veid -run "TestTask85D" -count=1
```

Result:

```text
FAIL: TestTask85DWebEvidenceIssuerProcessBoundaryConformance
expected wallet tier "standard", actual "basic"
```

Classification: stronger wallet lineage assertion RED. The test was corrected to assert the wallet's verified-status-aware `TierFromScore` path instead of the older identity-record tier helper.

Command:

```text
$env:GOCACHE=(Resolve-Path .cache).Path + '\go-build'; go test ./tests/integration/veid -count=1
```

Result:

```text
FAIL: TestVEIDRegistrationVerificationAuthorizationFlow
scorer not healthy: ML inference failed
```

Classification: production fail-closed RED. The old integration flow implicitly selected a scorer path that is now fail-closed. The test now continues with deterministic fixture values after separately proving ordinary score mutation messages are rejected.

Command:

```text
golangci-lint run --timeout=10m ./tests/integration/veid ./x/veid/keeper
```

Result:

```text
tests\integration\veid\task85d_process_boundary_test.go:1082:19: G115: integer overflow conversion int -> uint32 (gosec)
x\veid\keeper\task85d_consensus_conformance_test.go:455:109: task85DConsensusRegisterHelper - result 1 (error) is always nil (unparam)
```

Classification: Task 85D lint RED. The golden-input risk count is now explicitly bounded and annotated, and the consensus helper registration function no longer returns an always-nil error.

## GREEN Evidence

Command:

```text
$env:GOCACHE=(Resolve-Path .cache).Path + '\go-build'; go test ./tests/integration/veid -run "TestTask85D" -count=1
```

Result:

```text
ok  	github.com/virtengine/virtengine/tests/integration/veid	4.642s
```

Command:

```text
$env:GOCACHE=(Resolve-Path .cache).Path + '\go-build'; go test ./x/veid/keeper -run "TestTask85D" -count=1
```

Result:

```text
ok  	github.com/virtengine/virtengine/x/veid/keeper	0.634s
```

Command:

```text
$env:GOCACHE=(Resolve-Path .cache).Path + '\go-build'; go test ./tests/integration/veid -count=1
```

Result:

```text
ok  	github.com/virtengine/virtengine/tests/integration/veid	5.108s
```

Command:

```text
$env:GOCACHE=(Resolve-Path .cache).Path + '\go-build'; go test ./x/veid/types ./x/veid/keeper ./cmd/inference-sidecar ./pkg/inference -count=1
```

Result:

```text
ok  	github.com/virtengine/virtengine/x/veid/types	0.149s
ok  	github.com/virtengine/virtengine/x/veid/keeper	22.281s
ok  	github.com/virtengine/virtengine/cmd/inference-sidecar	1.093s
ok  	github.com/virtengine/virtengine/pkg/inference	2.008s
```

Command:

```text
$env:GOCACHE=(Resolve-Path .cache).Path + '\go-build'; go vet ./tests/integration/veid ./x/veid/types ./x/veid/keeper ./cmd/inference-sidecar ./pkg/inference
```

Result: exit code 0, no output.

Command:

```text
python .github/tests/test_inference_deployment_policy.py
python .github/scripts/validate_inference_deployment_policy.py
node scripts/validate-agents-docs.mjs
```

Result:

```text
Ran 22 tests in 1.241s
OK
PASS inference deployment policy
AGENTS documentation validation passed (9 files).
```

Command:

```text
$env:GOLANGCI_LINT_CACHE=(Resolve-Path .cache).Path + '\golangci-lint'; golangci-lint run --timeout=10m ./tests/integration/veid ./x/veid/keeper
```

Result:

```text
Task 85D files: no findings after fixes.
Existing non-Task-85D package findings remain in x/veid/keeper:
consensus_verifier.go, evidence_pipeline.go, fallback_handler.go, scoring.go, verification_pipeline.go unused-code findings.
```

## Preflight Evidence

Command:

```text
pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/task85d-preflight.ps1 -SkipRace
```

Result:

```text
ok  	github.com/virtengine/virtengine/tests/integration/veid	4.852s
ok  	github.com/virtengine/virtengine/x/veid/types	0.137s
ok  	github.com/virtengine/virtengine/x/veid/keeper	1.440s
ok  	github.com/virtengine/virtengine/cmd/inference-sidecar	1.291s
ok  	github.com/virtengine/virtengine/pkg/inference	0.774s
ok  	github.com/virtengine/virtengine/tests/integration/veid	5.167s
ok  	github.com/virtengine/virtengine/x/veid/types	0.155s
ok  	github.com/virtengine/virtengine/x/veid/keeper	23.195s
ok  	github.com/virtengine/virtengine/cmd/inference-sidecar	1.165s
ok  	github.com/virtengine/virtengine/pkg/inference	1.858s
Ran 22 tests in 1.851s
OK
PASS inference deployment policy
AGENTS documentation validation passed (9 files).
task-scoped golangci-lint: 0 issues.
generated contract drift check: not applicable to current Task 85D files
DIAGNOSTIC SKIP: WSL race test was not run.
This output is NOT release evidence for Task 85D.
Task 85D preflight passed.
```

Command:

```text
pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/task85d-preflight.ps1
```

Result:

```text
All gates before WSL race passed, including focused/full tests, policy validation, docs validation, vet, task-scoped lint, task-scoped diff check, and generated drift applicability.
WSL race focused Task 85D tests failed:
Windows Subsystem for Linux has no installed distributions.
WSL race focused Task 85D tests failed: WSL is installed but no Linux distribution is available; full Task 85D race preflight fails closed. Use -SkipRace only for diagnostic non-release evidence.
```

## Local Conformance Summary

- Web issuer children own deterministic Ed25519 private keys and return only public registration data plus signed SSO/email/SMS/social attestations.
- Parent-owned account wallet keys remain independently generated; issuer private keys never leave child responses.
- Exact idempotent retries return the same response and do not mutate state.
- Negative web cases cover wrong account, wrong chain, wrong scope, wrong type, changed payload, nonce replay, stale evidence, expired evidence, revoked key, and rotated unregistered key. Each failure asserts no store mutation.
- Inference child owns deterministic receipt signer private key and returns only public key plus canonical signed `types.InferenceReceipt` bytes.
- Two fresh inference children produce identical canonical bytes and outputs for the same golden input; a changed golden input remains deterministic but changes input/feature/output/lineage digests, score, and receipt bytes.
- Consensus conformance uses four independent committed keeper stores. Three validators with 90 total power independently stage the identical child-issued quorum receipt through real `ExtendVote`; the 10-power validator stages a separate valid dissenting receipt. The dissent cannot reach strict quorum.
- Finalization uses carried receipt bytes and a fresh keeper over each store, proving no local receipt buffer is required by `SubmitConsensusVerification`.

## External Blockers

External certification prerequisites are tracked separately in `_docs/task-85d-external-prerequisite-certification-ledger.md`. Race evidence remains blocked locally if WSL has no Linux distribution installed; `-SkipRace` output is diagnostic only and not release evidence.
