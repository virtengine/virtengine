# Five-Thread Prototype Completion Plan

**Campaign branch:** `stable-virtengine-beta`  
**Planning baseline:** `79391a3df86d85522b92e0400c6904971ecbe65d`  
**Prepared:** 2026-08-01  
**Source of scope:** `_docs/protocol-completion-continuation-plan.md`  
**Source of status:** `_docs/ralph/progress.md`

## 1. Confirmed Boundary

- Tasks 84A-85A are locally complete.
- Tasks 85B and 85C are locally complete as
  `engineering_complete_external_blocked`; their named external certification
  requirements remain open.
- Task 86A is locally implemented, although Docker-image execution, public-node
  route certification, three TypeScript harness suites, and an approved Buf
  breaking-change baseline remain open. Record these as dependency evidence,
  not as a reason to repeat 86A implementation.
- Task 85D is the next implementation task. Partial web-evidence and receipt
  verification foundations exist, but no authenticated production runtime
  currently produces signed inference receipts.
- Remaining scope is 22 implementation tasks, split exhaustively across five
  persistent threads:

  | Thread | Roadmap tasks | Count |
  | --- | --- | ---: |
  | T1 Identity Trust | 85D, 87A, 90C, 90D | 4 |
  | T2 Product Clients | 86B, 86C, 86D, 91B | 4 |
  | T3 Protocol Reliability | 87B, 87C, 87D, 90B | 4 |
  | T4 Integration and Release | 88A, 88B, 88C, 88D | 4 |
  | T5 Platform Security | 89A, 89B, 89C, 89D, 90A, 91A | 6 |
  | **Total** | **85D through 91B** | **22** |

- Milestone M follows the 22 tasks but is not part of the three-day prototype.
  Its independent review, external certification, exact-digest deployment, and
  uninterrupted 28-day observation cannot be compressed into this window.
- The three-day target is an exact-SHA, fail-closed, demonstrable prototype with
  verified contracts and integrated vertical paths. It is not
  `planned_functionality_complete`, GA, or production-certified.

## 2. Persistent Thread Model

Create five new VS Code chat threads. Give each thread exactly one master prompt
from `_docs/thread-queues/`. Each thread may launch multiple subagents for
research, tests, review, or non-overlapping implementation, but its parent
thread remains responsible for its queue, branch, commits, test evidence, and
handoffs.

Each thread follows this loop without waiting for another user prompt:

1. Read its queue and dependency ledger.
2. Select the highest-priority dependency-satisfied checkpoint.
3. Use subagents in parallel where their file ownership does not overlap.
4. Implement the checkpoint end to end.
5. Run focused tests and commit only a green checkpoint.
6. Push the thread branch and update its handoff ledger.
7. Immediately continue with the next ready checkpoint.
8. When blocked, record the exact blocker and select the first fallback item.
9. Stop only when the queue and fallback queue are exhausted or a genuine
  external prerequisite prevents all useful work.

The master prompts are:

- `_docs/thread-queues/thread-1-identity-trust.md`
- `_docs/thread-queues/thread-2-product-clients.md`
- `_docs/thread-queues/thread-3-protocol-reliability.md`
- `_docs/thread-queues/thread-4-integration-release.md`
- `_docs/thread-queues/thread-5-platform-security.md`

## 3. Branches, Worktrees, and Integration

Parallel threads must not mutate one checkout or one branch. Create sibling
worktrees from the same immutable baseline:

| Thread | Branch | Suggested worktree |
| --- | --- | --- |
| T1 | `ve/prototype-t1-identity` | `../virtengine-t1` |
| T2 | `ve/prototype-t2-product` | `../virtengine-t2` |
| T3 | `ve/prototype-t3-reliability` | `../virtengine-t3` |
| T4 | `ve/prototype-integration` | `../virtengine-t4` |
| T5 | `ve/prototype-t5-platform` | `../virtengine-t5` |

Bootstrap from the clean repository before starting the threads:

```powershell
git fetch --prune origin
$baseline = git rev-parse origin/stable-virtengine-beta
git worktree add ..\virtengine-t1 -b ve/prototype-t1-identity $baseline
git worktree add ..\virtengine-t2 -b ve/prototype-t2-product $baseline
git worktree add ..\virtengine-t3 -b ve/prototype-t3-reliability $baseline
git worktree add ..\virtengine-t4 -b ve/prototype-integration $baseline
git worktree add ..\virtengine-t5 -b ve/prototype-t5-platform $baseline
```

T4 is the only integration owner. T1, T2, T3, and T5 never push directly to
`stable-virtengine-beta` or T4's branch. They publish immutable green checkpoint
SHAs. T4 accepts one producer checkpoint at a time.

Integration follows `_docs/prototype-thread-intake-runbook.md`. T4 publishes a
frozen intake epoch base. Producers synchronize once to that base, resolve
conflicts on their own branch, and rerun their gates. Later T4 commits do not
invalidate checkpoints in the open epoch. T4 does not repair producer conflicts.

### 3.1 Producer Intake Gate

A producer checkpoint is intake-eligible only when all of the following can be
verified from a fresh clone after its push:

1. The remote producer payload descends from the frozen intake epoch base.
2. The handoff names `payload_head`, frozen baseline, planning SHA, intake epoch,
   intake base and prior accepted payload.
3. The immutable checkpoint tag targets the committed handoff descendant. Only
   declared handoff/evidence paths may differ after `payload_head`.
4. The frozen baseline remains `79391a3df86d85522b92e0400c6904971ecbe65d`;
   planning and intake SHAs are separate fields.
5. `tree_clean` is true, every test entry has a literal executable command,
  exit code `0`, result `passed`, and the required pinned tool versions.
6. The producer has pushed its annotated `checkpoint/prototype-tN/<id>` tag and
  retained the exact command output in the handoff evidence location.

The T4 intake validator must check remote-ref ancestry, tag target, handoff
range ownership, baseline identity, and literal test commands. Schema-only YAML
validation is not sufficient. A stale, dirty, placeholder, or locally-only
handoff is rejected and remains the producer's responsibility to repair.

Generated protobuf/OpenAPI/SDK output, module metadata, application wiring,
upgrade registration, and release manifests use explicit single-owner windows
coordinated by T4. Contract source may be developed in producer branches, but
generation and cross-thread registration happen only after T4 grants the window.

## 4. Commit, Checkpoint, and Tag Contract

Each thread checkpoint produces one or more atomic Conventional Commits. Stage only declared
paths; do not use `git add -A` or `git commit -a`.

```text
test(veid): cover issuer-bound evidence replay
feat(veid): enforce governed evidence signatures
fix(app): reject mismatched inference receipts
docs(veid): record task 85d conformance evidence
```

Every commit includes trailers:

```text
Campaign: 85D
Thread: T1
Checkpoint: T1-03
Handoff-From: <full-parent-sha>
```

All Go checkpoint commands use the repository-pinned Go `1.25.8` toolchain.
Every handoff records its absolute executable path or an unambiguous versioned
tool name and the exact `go version` output. A checkpoint validated with a
different Go version is diagnostic only and cannot be tagged or integrated.

Before commit, verify `git diff --cached --name-status` and
`git diff --cached --check`. Push only the owning thread branch:

```powershell
$newHead = git rev-parse HEAD
git push origin "${newHead}:refs/heads/ve/prototype-t1-identity"
```

After CI passes for each checkpoint, create and push an annotated immutable checkpoint
tag. Checkpoint tags never use the release-reserved `v*` namespace:

```powershell
git tag -a checkpoint/prototype-t1/t1-03 -m "Prototype T1 checkpoint 03" $newHead
git push origin refs/tags/checkpoint/prototype-t1/t1-03
```

Use `checkpoint/prototype-tN/<checkpoint-id>` for producer checkpoints and
`checkpoint/prototype-integration/<sequence>` for integrated checkpoints. Use
`milestone/prototype/<name>` only for T4's reproducible prototype manifest.
Never move or recreate a published tag. A correction receives a new commit and
new checkpoint identifier.

The historic mixed commits before this plan are cumulative boundaries, not
atomic task commits. Do not retroactively claim otherwise. The first execution
slot records the current HEAD as the campaign baseline.

## 5. Handoff Contract

Each thread maintains `_docs/ralph/handoffs/prototype-tN/HANDOFF.yaml`. Every
green checkpoint commits an update following the intake runbook. The core fields
are:

```yaml
campaign: three-day-prototype
thread: T1
checkpoint: T1-03
branch: ve/prototype-t1-identity
frozen_baseline: 79391a3df86d85522b92e0400c6904971ecbe65d
planning_sha: <stable-planning-sha>
intake_epoch: 1
intake_base_sha: <frozen-t4-epoch-base>
payload_head: <implementation-sha>
prior_accepted_payload: null
tree_clean: true
commits_since_prior_acceptance: []
owned_paths: []
files_changed: []
tests: []
generated_hashes: {}
migrations: []
external_evidence: []
known_failures: []
blockers: []
next_checkpoint: T1-04
```

Commands, exit codes, platform/tool versions, generated hashes, and external
non-claims are evidence. Summaries without retained results are not evidence.
The checkpoint tag identifies the handoff commit. `payload_head` identifies the
green implementation commit. The intervening range is limited to declared
handoff and retained evidence files.

If a slot cannot reach green, preserve `git diff --binary`, untracked-file names
and hashes, test output, and expected parent in the external lease record. A
designated recovery agent inherits that exact checkout. Revert a bad pushed
commit with `git revert`; never rewrite shared history.

## 6. Three-Day Parallel Queue

Every roadmap task has one owner. The queue uses small checkpoints so threads
can commit and integrate continuously instead of holding three days of changes.

| Thread | Day 1 | Day 2 | Day 3 | Fallback queue |
| --- | --- | --- | --- | --- |
| T1 Identity Trust | 85D trust inventory, envelope, runtime reader | Complete evidence vertical, workload binding, signed receipt seam | 90C compatibility, 90D decision input, integrated contract gate | Remaining evidence classes and receipt tamper vectors |
| T2 Product Clients | 86B SDK capability states and wallet sessions | 86C provider fail-closed APIs and first authoritative UI seam | 86D production provider gate/upload transport; isolated 91B deny-by-default action gate | Individual billing, usage, metrics, ticket adapters |
| T3 Protocol Reliability | 87B packet contract and atomic transitions | 87C explicit states and durable jobs/results | 87D producer telemetry and inactive 90B signed benchmark envelope | Restart, replay, conservation, and metric conformance cases |
| T4 Integration and Release | Baseline, producer intake contract, 88A inventory/RED fixtures | Integrate checkpoints; 88B required-gate matrix; 88C convergence contract | Integrate checkpoints; 88C validators; exact-SHA prototype manifest | Conflict-free intake, compile, generated drift, evidence repair |
| T5 Platform Security | 89A and 89B contracts/fixtures | 89C and 89D contracts/vectors | 90A profile fixtures and 91A policy contracts | Negative tests and cross-contract compatibility fixtures |

T1, T2, T3, and T5 should target checkpoints of two to six hours. T4 polls for
new producer checkpoints between its own queue items and integrates only green,
dependency-compatible SHAs.

T2-03 cannot be green until a legacy `ChainQuery` implementation that does not
affirmatively declare authoritative capabilities fails startup or readiness with
a typed unavailable state. Capability enforcement is opt-out only for explicit,
test-scoped mocks; production-facing defaults must fail closed.

T3-03 cannot be green until table-driven tests cover matched, mismatched,
unavailable, stale, and unresolved reconciliation states, including tolerance
boundaries, partial responses, and malformed evidence. Each non-matched state
must preserve the durable intent and block settlement mutation.

Tasks whose roadmap predecessors are incomplete may produce only isolated
contracts, validators, fixtures, and fail-closed feature gates. They must not be
wired into production, advertised, or marked complete. This lets all five
threads make useful progress without falsifying dependency completion.

### 6.1 Supplemental AI and Biometric Workstream

The model, private-evidence, biometric-uniqueness and fund-authorization audit is
specified in `_docs/protocols/veid-ai-biometric-architecture.md`. It adds 39
checkpoint-sized tasks without changing roadmap ownership:

| Thread | Supplemental IDs | Count | Primary outcome |
| --- | --- | ---: | --- |
| T1 | T1-10 through T1-18 | 9 | Canonical model schema, stage decisions, uniqueness receipts, eligibility and authenticated inference |
| T2 | T2-09 through T2-16 | 8 | Off-chain capture, client cleanup, encrypted claims, selective presentation and passkeys |
| T3 | T3-08 through T3-13 | 6 | Model, uniqueness, deletion, drift and appeal evidence |
| T4 | T4-10 through T4-16 | 7 | Model supply chain, parity CI, prohibited-path gates, generated integration and live fund-handler enforcement |
| T5 | T5-08 through T5-18 | 11 | Cryptography, OTP, durable custody, private uniqueness boundary and fund authorization |
| **Total** |  | **41** |  |

P0 prototype blockers take precedence over the earlier queue when encountered:

1. repair envelope encryption/authentication and reject arbitrary OTP;
2. freeze Python/Go feature parity and disable placeholder production models;
3. keep raw evidence and biometric templates off-chain;
4. implement signed stage/uniqueness/eligibility contracts with hard gates;
5. require transaction-bound authorization for every value-moving route;
6. add model/license/privacy/deletion evidence to T4's exact-SHA manifest.

Biometric uniqueness and face authentication remain separate. Public fuzzy
embedding hashes are prohibited. Prototype uniqueness uses contracts and a
clearly labelled non-cryptographic simulator; production requires an externally
reviewed threshold-MPC/confidential-compute design. Preferred face-backed
authentication is a device-local biometric unlocking a WebAuthn/passkey.

### 6.2 Active Thread Synchronization

If a thread already has uncommitted work when this supplemental queue lands, it
must not merge, copy, or regenerate over that dirty checkpoint. It first makes
its current item green, commits and pushes it, and records the handoff. It then
merges the updated `stable-virtengine-beta` or latest T4 integration checkpoint
into its own branch, resolves conflicts there, reruns focused gates, and begins
the first ready supplemental ID. T4 never overwrites or repairs a producer's
dirty worktree.

## 7. Cross-Thread Contract Gates

- T1 publishes Evidence Envelope v1, Runtime Policy Reader v1, Workload Binding
  v1, and signed receipt vectors. T2 consumes the client envelope; T4 integrates
  generated artifacts; T5 does not redefine identity or recovery evidence.
- T2 publishes SDK capability states, wallet-session rules, typed provider API
  availability, and upload/transaction receipt adapters. T4 owns final route and
  generated-client integration.
- T3 publishes IBC terminal states, reconciliation states, durable intent
  semantics, and bounded metric names. T4 owns deployment and required CI gates.
- T5 publishes only versioned contracts and `fixture_only` profiles until 88D
  and the relevant predecessor digests exist. Missing dependencies always yield
  typed `feature_unavailable` or disabled readiness.
- T4 publishes the current integration SHA and grants single-owner windows for
  generated files, app wiring, upgrades, charts, and manifests.

The `localnet.sh test` and `localnet.ps1 test` commands must use the same
explicit integration build tags as the canonical `make test-integration` gate;
a launcher that executes only untagged packages is not integration evidence.

Task 85D does not implement Task 90C lifecycle transitions. The T1 prototype
reader may carry compatibility fixtures for future states, but it accepts only
the exact runtime-eligible profile and never mutates model lifecycle.

## 8. Quality Gates

### Per commit

- Format only touched files.
- Run focused unit and negative/replay tests for touched packages.
- Run `go vet` for touched Go packages.
- Run targeted race tests for concurrency, replay, registry, sidecar, or keeper
  changes.
- Run `git diff --cached --check` and inspect staged paths.

### Per slot checkpoint

- Rerun the focused package tests from a clean committed SHA.
- Run applicable generation verification for protobuf, OpenAPI, SDK, module, or
  vendor changes.
- Run predecessor integration checks affected by the slot.
- Wait for CI on the exact SHA before publishing the checkpoint tag. Rapid
  pushes to the same branch can cancel prior workflow runs and are not evidence.

### Task 85D completion

Create `scripts/task85d-preflight.ps1` as part of S01 and make it fail closed.
Until it exists, the minimum explicit gate is:

```powershell
go test -count=1 ./x/veid/types ./x/veid/keeper ./x/veidregistry/... ./pkg/inference ./cmd/inference-sidecar
go test -race -count=1 ./x/veid/keeper ./pkg/inference ./cmd/inference-sidecar
go test -count=1 ./app -run 'VoteExtension|Proposal|Inference'
go test -tags=e2e.integration -count=1 ./tests/integration/... -run 'VEID|Inference'
python .github/tests/test_inference_deployment_policy.py
python .github/scripts/validate_inference_deployment_policy.py
go run ./scripts/consensusdeterminism -root .
node scripts/validate-agents-docs.mjs
pwsh scripts/agent-preflight.ps1
```

Add separate-process SSO, email, SMS, social, account-signing, inference-runtime,
restart, key/certificate rotation, outage, and receipt-tamper tests. In-process
fixtures alone do not satisfy 85D.

### Per sprint and core milestone

- Compile all Go packages with `go test -run '^$' -count=1 ./...`.
- Run full applicable preflight with no diagnostic skip variables.
- Require zero new lint, vet, race, build, generation, or test failures.
- Preserve external blockers as explicit matrix states. Mocks, unavailable
  hardware, and missing credentials never produce a certified status.

## 9. External Workstream

External onboarding runs in parallel with code sprints but cannot bypass the
DAG. The orchestrator tracks at least:

- production DEX route and fiat payout corridor certification;
- TMKMS/HSM, Kubernetes storage, multi-zone and regional DR evidence;
- real TEE hardware and current collateral;
- strong authenticators/verifier roots and threshold recovery participants;
- production KMS/HSM, durable storage, retention, hold, restore, and erasure;
- two independent provider operators with DNS/PKI control;
- Kubernetes and one VM/IaaS account with quotas and cleanup authority;
- representative benchmark hardware and independent verifier;
- consented datasets, labeling, fairness/privacy approvals;
- non-exportable issuer custody, ceremony, two wallets, two relying parties;
- privileged approvers, break-glass custody, model hosting, red team, and human
  handoff operations.

Milestone M starts only after every implementation task and mandatory profile is
eligible. Its 28-day observation is continuous and non-overlapping. Any bound
digest/configuration/profile change or telemetry gap resets the window.

## 10. Definition of Three-Day Prototype Success

The urgent prototype succeeds when:

1. all five thread branches publish green checkpoint SHAs and handoff ledgers;
2. T4 integrates the accepted checkpoints without undeclared conflicts;
3. one identity-to-client vertical uses authenticated evidence and a signed,
  profile-bound deterministic receipt or fails closed when unavailable;
4. provider/client false-success paths selected for the prototype return real
  committed state or typed unavailability;
5. IBC/reconciliation prototype states are replay-safe and restart-testable;
6. platform features without satisfied dependencies remain `fixture_only` or
  disabled and cannot be activated accidentally;
7. the exact integration SHA compiles, focused/race/negative tests pass, and a
  reproducible `virtengine.core-rc.prototype/v0` manifest records every missing
  external or GA prerequisite truthfully.

Full campaign success remains governed by the continuation plan: every roadmap
row must reach a terminal evidence state, mandatory GA profiles must be
certified, and Milestone M must pass independently. The prototype milestone does
not alter those obligations.