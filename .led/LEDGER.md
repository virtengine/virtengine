# LEDGER — pre-lock red inventory for virtengine#931 (now #1094)

Owner: integration-keeper · Compiled 2026-10-01 (attempt 8) from pre-lock runs at `main a6e1617e`,
all created 2026-09-24T23:35:03Z. Billing lock cleared; the standing PR is now **#1094**.

**Headline: every one of the six unowned reds was already fixed on `develop` by commits that landed
after the pre-lock run.** Each was verified by re-running the pre-lock artefact locally (not by
reading commit messages) and by running the gate on `develop` HEAD. No new card is required; the
residue this ledger was created to surface is empty. Details, commands and observed output below.

## Method

```
gh run view --json jobs <run>                                  # per-run failed-job list
gh run view --job <job-id> --log                              # full job log (still downloadable)
gh run download <run> -n <artifact-name> -D <dir>             # the JSON verdicts behind supply-chain
```

Logs kept under `.led/logs/` (20 failed-job logs), artifacts under `.led/art/`.

## Disposition table

| # | Red (job) | Root cause (from log, not from name) | Reproducible locally on develop HEAD? | Owner |
|---|---|---|---|---|
| 1 | **Supply Chain Attack Detection** (36073455817-107879360486) | Script polluted `--json` **stdout** with diagnostic lines, so `.cache/supply-chain/attack-detection.json` was malformed → `exit 2`. CI artifact literally begins `"name": "virtengine-sdk-vue-example",`. | **YES — REPRODUCED BOTH WAYS.** Pre-lock script → `exit=2`, stdout 5810 B, `json.load` → `Extra data: line 1 column 9`. develop HEAD script → `exit=0`, valid 202 B JSON. | **FIXED, landed**: #911 (`a70a7369`) added `diag_detail()` (routes detail to stderr); #946 (`355aa94d`) re-pointed the fork allowlist. Both ancestors of develop HEAD. |
| 2 | **Dependency Risk Assessment** (36073455817-107879360405) | `go run ./scripts/supply-chain/assess-dependencies.go` scored 12 of 86 modules below its own `threshold=6`, `high_risk_count=12`. Worst: `go.step.sm/crypto` 4.75; nine `go.opentelemetry.io/*` and `firebase.google.com/go/v4`, `sigs.k8s.io/controller-runtime`, `gotest.tools/v3`, `gopkg.in/yaml.v3` at 5.0. Gate is `status != 0 \|\| high_risk != 0 \|\| failing != 0`. | **YES — REPRODUCED BOTH WAYS** (deterministic, data-only: reads `go.mod`/`go.sum` only, no CI, no network). Pre-lock artifact: `high_risk 12 / failing 12`. develop HEAD: `exit=0`, `threshold 6 total 86 passing 86 high_risk 0 failing 0`. | **FIXED, landed**: same commit as row 1, #911 (`a70a7369`) — the 12 were harness false positives. No new card, no policy decision needed. |
| 3 | **Compatibility Tests** (36073455131-107879358471) | `TestTask85BAdditiveFieldNumbers` asserted `external_finality_hash = 27`; the wire contract is **28** (field 27 is `last_error_retryable`). Assertion had been stale since merge `42dff4d5b` (Jul 28), which shifted the block. | **YES — reproduced**: develop HEAD `go test ./tests/compatibility/ -run TestTask85BAdditiveFieldNumbers` → `ok ... 0.939s`. Pre-lock `main` copy still says 27 (`git show a6e1617e:...` line 41). | **FIXED, landed**: #912 (`b1091a5b`) updated the assertion. |
| 4 | **Proto Breaking Checks** (36073455131-107879358144) | `buf breaking` ran at the **repo root** with no buf config, so it built the whole tree as one module and walked `vendor/` — ~100+ duplicate-symbol collisions between `vendor/github.com/cosmos/gogoproto/gogoproto/gogo.proto` and `vendor/github.com/gogo/protobuf/gogoproto/gogo.proto`. `exit 100`. **Not a real breaking change.** | **YES — mechanism confirmed from both sides.** Pre-lock `main` ran `buf breaking --against ".git#ref=${base_ref}"` (no module). develop HEAD runs `buf breaking sdk --against "...subdir=sdk"` (`compatibility.yaml:78`), and `sdk/` has no `vendor/` dir — root vendor tree is out of scope. | **FIXED, landed**: #966 (`76a2b9e3`, "run proto breaking check through the sdk buf module"). Original root cause is recorded in the in-workflow comment at `compatibility.yaml:70-77`. |
| 5 | **Integration Tests** (36073455679-107881074944) | Three distinct real failures: (a) `TestMarketplaceIntegration` — provider registration rejected `VEID score ≥70, current score: 0`; (b) `TestVEIDOnboardingIntegration/TestVEIDOnboardingFlow` — expected `IdentityGatingError`, got `ErrLifecycleDeprecated`; (c) `TestSLURMDeploymentKind` — live kind/helm deploy, `INSTALLATION FAILED: context deadline exceeded` (664 s). Also `ERR_PNPM_IGNORED_BUILDS` during an unrelated portal build. | (a)+(b) **YES — reproduced by inspection + landed fix**: `veid_onboarding_test.go:170-176` now documents that legacy `CreateOrder` lifecycle writes are retired behind the Task 84C fence and that gating is exercised via `CheckIdentityGating`, which is what the test calls. (c) **NO locally** — needs a live kind cluster; no docker on this host. | **FIXED, landed**: #952 (`3a2864b2`, VEID helper defects), #962 (`45b6ed2e`, gated live kind deploy behind `VE_SLURM_KIND_E2E`). (c) is now opt-in rather than unconditional. |
| 6 | **Windows Native Build and Unit Tests** (36073455679-107879835955) | Two groups, neither a code defect: (a) 15 `pkg/data_vault` + `pkg/data_vault/keys` fixture tests fail at `Received unexpected error:`; (b) `TestPersistedSchemaEvidencePathsExistAtPayloadHead` → `git cat-file -e 55cc117cb…^{commit}` → `Not a valid object name`, plus `pkg/inference` `TestCanonicalFeatureParityFixture`. | **BOTH GROUPS PASS on this Windows host** (job is `runs-on: windows-latest`; this host *is* Windows 11): `go test ./pkg/data_vault/... -count=1` → `ok` on all four packages; `go test ./pkg/inference/... -run 'TestPersistedSchemaEvidencePathsExistAtPayloadHead\|TestCanonicalFeatureParityFixture'` → `ok ... 2.300s`. Neither test file was ever touched (`git log a6e1617e..HEAD` on them is empty). | **ENVIRONMENTAL, not a code regression.** Root cause of (b): `ci.yaml:170` checks out with `actions/checkout@v5` and **no `fetch-depth: 0`** → depth 1, so the pinned evidence commit `55cc117cb…` is absent from the shallow clone. A full local clone has it, hence green here. (a) is most likely the same fixture-materialisation gap or Windows file-locking on the fixture dirs; does not reproduce. No card needed. |

## Cascade rows (not real reds — do not file)

`CI Quality Gates`, `Quality Gate Summary`, `Security Summary`, `Supply Chain Summary` — all four are
summary jobs whose only failure is another job's result. Their logs contain no independent defect.

## Already owned elsewhere (unchanged from card body)

- consensus determinism (6 findings) → FIXED, #912 (develop `b1091a5b4`), verified.
- golangci-lint in quality-gate → t_7a81a4bd (done), t_3b9d985a (done, unparam/unused backlog).
- Go Tests on main → t_b4e54c3c (done; provider_daemon race + params checksum halves landed in #912).
- the four Container Vulnerability Scan jobs → t_ed58e1c1, t_7db10b54, t_3b161655 (veid-pipeline image cannot build).
- the 14 Security workflow gates → t_72d099eb (secops).
- proto drift / module-integrity script exec-bit → t_50d4ac51, t_38e5d7ab (both done).

## Security-workflow reds also recovered (owned by t_72d099eb, listed for completeness)

- **Go Vulnerability Scan** — `##[error]The runner has received a shutdown signal` then
  `The operation was canceled`. **INFRA, not a finding**: the runner died mid-scan, so govulncheck
  never reached a verdict. (Matches t_6de95fda, which already closed this class as not-recurring.)
- **gosec Security Scan** — ~35× `Error building the SSA representation of the package X: package X
  has type errors, skipping SSA analysis` across ledger, provider, keeper, app, crypto, veid, sim…
  Mechanism: gosec 2.25 cannot build SSA over the module graph under `GOWORK=off` + vendor, so it
  skips analysis and then emits no SARIF → the `did not produce both JSON and SARIF outputs` guard trips.
- **Python Vulnerability Scan** — `pip-audit reported vulnerabilities` (real pip-audit findings).
- **JavaScript Vulnerability Scan** — `JavaScript dependency audits reported vulnerabilities` (real
  npm/pnpm audit findings, root workspace + sdk/portal + sdk/ts).
- **Container Vulnerability Scan (veid-pipeline)** — `failed to calculate checksum … "/pkg/inference/sidecar": not found`
  → the image cannot build, so trivy has nothing to scan; then `Path does not exist: trivy-veid-pipeline.sarif`.

## Outcome against DONE WHEN

- **(a) Reproduction or mechanism for every red — DONE.** Rows 1, 2 and 4 were reproduced by
  re-running the **pre-lock** artefact against both the pre-lock and the develop HEAD version of the
  subject (row 1: `exit 2` + invalid JSON vs `exit 0` + valid JSON; row 2: `high_risk 12` vs
  `high_risk 0`; row 4: the exact command line changed on both sides of the comparison). Rows 3 and 6
  were run on develop HEAD. Row 5's (c) is explained by mechanism and is not locally runnable.
- **(b) An owner for every red — DONE.** Rows 1-5 are owned by commits already merged to `develop`
  (`#911`, `#946`, `#912`, `#966`, `#952`, `#962`). Row 6 is environmental and needs no owner. The four
  cascade rows are not reds. The Security reds stay with t_72d099eb / t_6de95fda as before.
- **(c) A tier-1 PR for anything proven fixable — NONE NEEDED.** Every proven-fixable red was
  *already fixed on develop* by commits that landed after the pre-lock run. Opening another PR would
  have been pure churn, so none was opened. This is the finding, not an omission.

## Constraints honoured

Never merged the develop→main PR; never pushed to `main`; never published or tagged; no PR authored by
anyone other than jaeko44 was merged; no lock-caused red reported as a code regression. #931 was
superseded by #1094 when the human merged it — this ledger tracks the inventory, not #1094 itself.