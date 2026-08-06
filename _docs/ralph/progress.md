# VirtEngine Progress

Last updated: 2026-07-23
Status: **Protocol continuation Task 85C is complete at the deterministic local engineering boundary as `engineering_complete_external_blocked`. Task 85D is NEXT.** Live TMKMS/HSM, multi-zone Kubernetes storage and regional DR certification remain externally blocked. Historical 80A-83D backlog status remains below for traceability.

## Protocol Completion Continuation (2026-07-20)

| ID | Title | Priority | Status |
| --- | --- | --- | --- |
| 84A | Eliminate consensus nondeterminism and enforce deterministic ABCI++ admission | P0 | **COMPLETED** |
| 84B | Cryptographically authenticate metering and make settlement replay-safe | P0 | **COMPLETED** |
| 84C | Decide canonical market lifecycle and unify reservations | P0 | **COMPLETED** |
| 84D | Converge fraud, HPC, billing, escrow disputes, and review effects | P0 | **COMPLETED** |
| 85A | Route every provider chain mutation through the durable signed broadcaster | P0 | **COMPLETED** |
| 85B | Deliver verifiable DEX routing and a compliant fiat off-ramp | P0 | **`engineering_complete_external_blocked` (LOCALLY COMPLETE)** |
| 85C | Converge production deployment rendering and prevent validator double-signing or provider HA state loss | P0 | **`engineering_complete_external_blocked` (LOCALLY COMPLETE)** |

### 85C: Converge Production Deployment Rendering and Prevent Validator Double-Signing or Provider HA State Loss

**Locally completed:** 2026-07-23
**Status:** `engineering_complete_external_blocked`
**Commit:** Not created per orchestration instruction

- ADR-009 makes `deploy/kubernetes` the only application source. Infra base/dev/staging/prod entry points are import-only compatibility shims with byte-equivalent renders; chaos and DR assets consume canonical labels/PVCs.
- `infra/rollouts` is not application topology: stale blue/green Rollout manifests were removed and the retained rollback config is optional AnalysisTemplate policy only.
- Production renders one explicit remote-signer validator plus four horizontally scalable sentry/full nodes. Validator readiness binds chain ID, address, key fingerprint, signer endpoint, epoch, fencing token, TMKMS config digest and signer certificate digest; sentries receive no signer identity.
- Provider production renders three StatefulSet replicas with durable shared encrypted identity/HA/backup PVCs, immutable images, PDB, anti-affinity, topology spread, restricted pod security, canonical metrics port 9090, and fencing/key-aware readiness.
- Added Kubernetes Lease ownership using resource-version atomic updates and monotonic `leaseTransitions` fencing tokens. Standbys stay unready until takeover; stale owners/tokens cannot renew or submit. Shared-file leases remain an explicit compatible profile; process-local leases are rejected in production.
- A chain-client mutation guard fences bid, resource, domain, HPC and other provider mutation producers on standby replicas; authenticated metering stops on ownership loss and starts only after fenced takeover.
- The mutation file store now reloads and updates under cross-process locking, preserving atomic idempotency/sequence/reconciliation writes between replicas.
- The provider file keystore now persists Argon2id/AES-256-GCM authenticated envelopes with atomic writes, metadata/private-key integrity, restart/rotation continuity, secret-file passphrases, fingerprint binding, and fail-closed missing production identity. Hardware/Ledger/non-custodial profiles fail closed; SoftHSM is not represented as real hardware certification.
- Provider DR backup/restore now carries encrypted identity, queues/idempotency, usage sequence/proofs, mutation sequence/reconciliation, fiat reconciliation, and fencing state. The executable local restore drill passed.
- Full `scripts/task85c-preflight.ps1` passed: provider/command tests, settlement continuity, vet, lint (`0 issues`), provider build, canonical/compatibility/chaos renders and policy, AGENTS docs, provider recovery drill, whitespace and WSL race tests. Repository-wide `go test ./... -run '^$' -count=1` compilation also passed.
- Final CI enforcement gap closed: `.github/workflows/infrastructure.yaml` now triggers on canonical infra/deploy changes, the Task 85C validator, focused Task 85C docs, and the INFRA-001 summary; its validate job installs Node 20 plus checksum-pinned kubectl 1.29.0 and runs `node scripts/task85c-validate-kubernetes.mjs` with read-only permissions and no deployment credentials before downstream security, plan, apply, or drift jobs.
- Focused local validation for the CI enforcement gap passed: the kubectl 1.29.0 Linux amd64 checksum matched `dl.k8s.io`; workflow YAML parsed with PyYAML and passed `actionlint` 1.7.7; the Task 85C validator and AGENTS validator passed; and `git diff --check` passed.
- Completion evidence and release-only limitations are in `_docs/audits/task-85c-completion-report-2026-07-23.md`; architecture and recovery guidance are in `_docs/adr/ADR-009-canonical-kubernetes-rendering-and-identity.md` and `_docs/runbooks/kubernetes-identity-backup-restore-runbook.md`.
- No live TMKMS/vendor HSM/mTLS signer, real double-sign/partition drill, named Kubernetes distribution, encrypted RWO/RWX storage backend, snapshot controller, immutable-image provenance, multi-zone eviction, detached-volume or regional failover evidence exists locally. These remain mandatory release-only blockers; no production hardware or DR certification is claimed.
- **NEXT:** Task 85D.

### 85B: Deliver Verifiable DEX Routing and a Compliant Fiat Off-Ramp

**Locally completed:** 2026-07-23
**Status:** `engineering_complete_external_blocked`
**Commit:** Not created per orchestration instruction

- Production factory selection now reaches the real, exact-version Osmosis `gamm-v1beta1-cp-equal-weight-v1` equal-weight constant-product adapter through an authenticated pool provider bound to height, block hash and node identity; Uniswap V2 and Curve remain rejected/test-only for production.
- Added strict versioned DEX/payout profile states and trust, exact pool/denom/decimal/height/finality/oracle/quote/payload checks, a strict provider-neutral payout HTTP adapter, authenticated replay-safe webhooks and durable privacy-safe payout repositories.
- Added the off-chain durable provider-daemon conversion orchestrator with local lease/fencing, restart and ambiguous-outcome recovery, target-chain custody boundary, and six ordered authenticated observation stages.
- Added generated provider-signed `MsgRecordFiatConversionObservation`, consensus profile/compliance/idempotency/lineage/finality validation, Task 85A durable mutation registration, settlement/payout holds and invariants. Current authorization gates new side effects, while immutable accepted commitments allow safe post-irreversible-boundary reconciliation after governance/profile/compliance/hold changes.
- Authenticated completion deterministically moves native net value module-to-module into the internal-only fiat-custody sink, with exactly-once platform-fee, validator-fee and holdback treasury entries. It performs no synthetic provider-account or external bank/chain transfer.
- Financial cases cannot double-allocate payout exposure after an irreversible fiat boundary. Restart-safe payout binding recovery requires an exact durable provider/quote/correlation/economic match, and expired DEX/payout quotes can be replaced only before the corresponding submission boundary without releasing holds or quota.
- Added upgrade `v1.8.0`; activation disables execution, sets both profile states to `engineering_complete_external_blocked`, preserves/quarantines legacy records and verifies deterministic reconciliation.
- Production startup remains fail-closed because reviewed external custody, payout partner, secret, destination, compliance and webhook backends are not injected. Backend identifiers do not enable execution.
- Aggregate targeted tests passed across DEX, off-ramp, provider-daemon/command, app custody, settlement, `v1.8.0`, compatibility and upgrade scope. Relevant vet passed; lint reported 0 issues; provider-daemon and virtengine builds passed; TypeScript SDK lint/build/tests passed; tagged settlement integration and `e2e.upgrade` compile/worker registration passed.
- Focused WSL race tests passed for Task 85B DEX/off-ramp, provider fiat/orchestrator and settlement fiat/consensus paths.
- Generated hashes were rechecked: descriptor `1ebde455e84065fa2f53cfc90536b9aece3423c55aed1abdb323ce42a7b9fb9e`; inventory `654f503d6e84c881f62882f2a57cca3bf8077b2772bac3176c452dc6ebac50c4`; OpenAPI `9e07efd4b158589d71907f350422fbba4ab9da36a3870a3065547ef8935d4d2d`. Full Task 85B preflight requires these live hashes in this progress record and the completion report.
- Completion evidence and external blockers are recorded in `_docs/audits/task-85b-completion-evidence-2026-07-23.md`; matrices, protocol, runbook and certification ledger are in `_docs/protocols/task-85b-dex-payout-support-matrices.md`, `_docs/protocols/fiat-conversion-orchestrator-protocol.md`, `_docs/runbooks/fiat-conversion-incident-recovery.md` and `_docs/task-85b-external-prerequisite-certification-ledger.md`.
- Deterministic local fixtures prove engineering conformance only. No real Osmosis testnet execution or provider sandbox payout is claimed; both remain required external conformance evidence before certification. No production custody, provider contract/credentials, certified pool/liquidity/oracle/governance evidence, approved corridor or independently reconciled production execution exists. The Claim 4 production floor and overall `planned_functionality_complete` status remain externally blocked.
- **NEXT:** Task 85C.

### 85A: Route Every Provider Chain Mutation Through the Durable Signed Broadcaster

**Completed:** 2026-07-23
**Commit:** Not created per orchestration instruction

- Added the versioned provider mutation envelope/registry, durable queue store with `QueueStore` compatibility alias, local single-submitter lease/fencing seam, bounded metrics and readiness reporting.
- Routed provider-originated bid, usage/settlement, HPC status/accounting/node metadata, resource heartbeat, provider domain/key/lifecycle, Waldur callback and support mutations through the durable signed submitter.
- Production provider-daemon mutation paths no longer use generated `MsgClient` calls and no longer return success only because chain clients are absent; disconnected writes return typed unavailable/not-ready errors.
- SDK transactions are direct-sign-mode signed by the configured provider `KeyManager`, gas-simulated, persisted with sequence/tx bytes/hash/attempts, broadcast, inclusion-confirmed, finalized or explicitly retried/dead-lettered.
- Crash/restart, response-loss, timeout, sequence mismatch, out-of-gas, mempool, replacement/reorg, lease-loss, missing evidence, mutable idempotency and logical-state reconciliation paths are covered by focused tests.
- Production fallbacks for legacy usage broadcast clients, file Waldur/provisioning callbacks, raw HPC node metadata submission and legacy domain confirmation are fail-closed.
- Customer/owner-signed market `MsgCreateLease` and `MsgCloseLease` are intentionally not registered as provider-originated queue kinds; attempts fail closed as unknown.
- Completion evidence and release-only limitations are recorded in `_docs/audits/task-85a-completion-report-2026-07-23.md`, with the mutation inventory in `_docs/audits/task-85a-mutation-inventory-2026-07-23.md`.

### 84D: Converge Fraud, HPC, Billing, Escrow Disputes, and Review Effects

**Completed:** 2026-07-22
**Checkout:** `99973c3af6c084a915e2913016d5d08899dcdaa6`
**Commit:** Not created per orchestration instruction

- ADR-008 selects `x/settlement` as the sole canonical financial-case owner; escrow billing, fraud, HPC and review are evidence/projection adapters after `v1.7.0`.
- Added generated multi-denom `FinancialCase`, complete lifecycle messages, 10 paginated query routes, generated privacy-safe events, deterministic domain-separated case/claim/appeal IDs, one active subject-group root and bounded append-only audit lineage.
- Opening uses cached contexts and binds payout, escrow and Task 84C reservation holds. Settlement, payout/retry, reward claim, HPC accounting/reward and expiry paths reject active cases.
- Explicit provider/customer roles remain stable when either party files; public filings prove committed lineage, trusted adapter calls cannot be spoofed through message fields, and active legacy billing workflows migrate as canonical claims.
- Resolution validates exact per-denom conservation, remains held through appeal, and persists exactly-once payout/refund/platform/slash/reservation/projection effects; retries converge without duplicate transfers.
- Fraud reports and HPC disputes open/merge canonical claims; review adds content-hash/reputation evidence; legacy billing/HPC/fraud money decisions are fenced or converted to canonical escalation requests after activation.
- Added `v1.7.0` migrations/reconciliation, terminal-finance preservation, duplicate/orphan quarantine, genesis import/export, consensus-version increments, invariants, representative reconciliation artifact and operator remediation runbook.
- Pinned generation completed repeatedly with zero drift: descriptor SHA-256 `38862f7cc92e1721d41585a243a65a664a1afaf363bc497cd8ab4f033fa733ea`; inventory SHA-256 `ff2c8d463de137b1400fdc5d07aecbace87fae82788f938821e46eda287f7193`; final before/after manifest SHA-256 `0e9c6f944e43d656786b3ec5f6ffb758afe421590220854c1fb8d1c43a7a0baa`; OpenAPI has 234 paths.
- Affected tests, 1,000-case allocation property test, WSL race, vet, lint (`0 issues`), binaries, repository-wide Go compilation, TypeScript build and all 36 suites/1101 tests, determinism, module/vendor policy, docs, Task preflight and mandatory repository preflight pass.
- Completion evidence and release-only process limitations are in `_docs/audits/task-84d-completion-report-2026-07-22.md`.

### 84C: Decide the Canonical Market Lifecycle and Unify Reservations

**Completed:** 2026-07-21
**Checkout:** `99973c3af6c084a915e2913016d5d08899dcdaa6`
**Commit:** Not created per orchestration instruction

- ADR-007 selects `x/market` as the only mutable Order→Bid→Lease/financial owner and retains `mktplace` as an offering/catalog compatibility surface with lifecycle writes disabled after activation.
- Added the generated authoritative `x/resources` reservation aggregate, explicit transitions, exact-retry idempotency, checked integer capacity, ordered expiration, bounded events, lineage queries, genesis, migrations, and invariants.
- Canonical market lease creation is atomic with reserve/escrow/lease/activate; standalone HPC owns one reservation while market-backed HPC reuses the canonical lease reservation without double-reserving capacity.
- Added `v1.6.0` with market `7→8`, mktplace `1→2`, resources `1→2`, and HPC `1→2` migrations; ambiguous active state is quarantined with zero synthetic capacity.
- Senior compatibility re-audit added explicit `resources→market→hpc→mktplace` migration order, reconciliation before non-owner activation, real-app upgrade-handler/idempotence tests, historical replay/current-genesis activation flags, deterministic request-digest reservation IDs, callback/usage write fencing, non-owner bid quarantine, terminal-genesis lineage/idempotency indexes, event-preserving genesis, paginated lineage, and linked/dangling/inconsistent legacy resource handling.
- Added fixed old-wire fixtures, randomized conservation, rollback, final-unit competition, fresh-genesis activation, migration, upgrade, and real-app keeper-wiring evidence.
- Final pinned generation is byte-stable; 224 OpenAPI routes, 418 descriptor methods, and 224 HTTP bindings are recorded under inventory hash `58d5236fde8755aef2c7d4d2f67a77f3712b3356c8d319ccdfb036c21a3d83e3`.
- Unit/integration/compatibility/upgrade tests, WSL race, vet, lint (`0 issues`), builds, determinism, generation verification, module verification, docs validation, preflight, and whitespace checks pass.
- Completion evidence and release-only process limitations are recorded in `_docs/audits/task-84c-completion-report-2026-07-21.md`.

### 84B: Cryptographically Authenticate Metering and Make Settlement Replay-Safe

**Completed:** 2026-07-21
**Requested base commit:** `e2d78c01a0b653baabc9d27d0ca26d8305dd54c6`
**Concurrent amended checkout:** `99973c3af6c084a915e2913016d5d08899dcdaa6`
**Commit:** Not created per orchestration instruction

- Added versioned, length-prefixed canonical provider/customer sign bytes with golden and legacy wire fixtures.
- Bound usage verification to immutable `x/provider` signing-key epochs and customer acknowledgment to supported `x/auth` account keys.
- Added exact-once sequence/nonce/idempotency/period indexes, cached-context atomicity, replay invariants, and durable daemon restart allocation.
- Added `v1.5.0`, settlement consensus version `2`, provider consensus version `4`, legacy-unverified migration, and fail-closed HPC pending accounting.
- Pinned generation completed twice with zero second-pass hash changes.
- Passed separate provider/customer helper-process execution, durable provider restart, local and chain exact-retry proofs, conflicting replay rejection, one usage/event/settlement/reward proof, and exact four-validator app-hash convergence at one committed height.
- Final recorded app-hash evidence: height `20`, all four validators `4C39E284A499E4EDC5BC363F541D8E2085E228E9391C75D9CEDEB951EC972E82`.
- Completion evidence and limitations are recorded in `_docs/audits/task-84b-completion-report-2026-07-21.md`.

### 84A: Eliminate Consensus Nondeterminism and Enforce Deterministic ABCI++ Admission

**Completed:** 2026-07-20
**Agent:** copilot_subagent with orchestrator verification
**Commit:** Not created per orchestration instruction

**Summary:**

- Added upgrade-gated bounded proposal admission with canonical SDK transactions and stable rejection codes.
- Added signed VEID vote-extension aggregation through one authenticated index-zero system transaction.
- Removed consensus wall-clock/runtime/external-service inputs and added a static determinism gate.
- Passed targeted tests, vet, lint, build, WSL race tests, and a four-validator 500-post-activation-block adversarial load test with a 5,000-transaction block and matching app hashes.
- Completion evidence is recorded in `_docs/audits/task-84a-completion-report-2026-07-20.md`.

## Inputs reviewed

- \_docs/ralph/ralph_patent_text.txt (AU 2024203136 A1 — 2121 lines, all 14 claims)
- Git history (279 commits since 2026-02-11, 5 since 2026-02-22 — all tooling/deps)
- Full x/ module audit (25 modules, LOC counts, keeper analysis, TODO/stub scan)
- Full pkg/provider\_daemon analysis (15+ files, adapter status, chain client/submitter audit)
- Full pkg/inference analysis (scorer, determinism, runtime, conformance tests)
- App wiring verification (app/app.go, app/modules.go — all 30+ modules registered)
- Proto definition inventory (sdk/proto/node — all modules have .proto files)
- Bosun scripts deep audit (2026-02-22): config.mjs, repo-root.mjs, monitor.mjs, agent-pool.mjs, task-executor.mjs, worktree-manager.mjs, codex-config.mjs, workspace-manager.mjs, setup.mjs — 9 files, 30k+ LOC
- 83A-83D implementation verification audit (2026-02-22): line-by-line acceptance criteria check
- Remote branch analysis (13 stale ve/ branches, none merged since Feb 11)

---

## 83A-83D Implementation Audit (2026-02-22)

### 83A: Workspace Isolation — ✅ FULLY IMPLEMENTED (8/8 criteria)

All acceptance criteria verified with code evidence:

| # | Criterion | Status | Evidence |
|---|-----------|--------|---------|
| 1 | Agents execute in workspace path | ✅ | `resolveAgentRepoRoot()` in config.mjs:867, agent-pool.mjs:61 |
| 2 | REPO\_ROOT no longer forces developer repo | ✅ | Priority inversion at config.mjs:868, repo-root.mjs:82 |
| 3 | `bosun --setup` auto-clones + verifies .git | ✅ | Interactive setup.mjs:4453 + non-interactive setup.mjs:4684 |
| 4 | Daemon startup auto-clones missing repos | ✅ | Bootstrap at monitor.mjs:559-594, pullWorkspaceRepos import at L209 |
| 5 | Sandbox writable roots include .git | ✅ | codex-config.mjs:437 adds .git per root |
| 6 | Worktrees under workspace .cache/worktrees/ | ✅ | worktree-manager.mjs:192 uses BOSUN\_AGENT\_REPO\_ROOT |
| 7 | Developer repo untouched | ✅ | All paths resolve via workspace root; REPO\_ROOT = "developer root for config only" |
| 8 | Graceful fallback to REPO\_ROOT | ✅ | monitor.mjs:547 chdir fallback, repo-root.mjs:96 resolution chain |

**Verdict: Task can be marked `done`.**

### 83B: Multi-Repo Task Routing — 🔶 ~70% IMPLEMENTED (3/6 full, 3/6 partial)

| # | Criterion | Status | Gap |
|---|-----------|--------|-----|
| 1 | Per-repo worktrees (waldur) | ✅ | `_resolveTaskRepoContext()` → per-repo WorktreeManager |
| 2 | Per-repo worktrees (virtengine) | ✅ | Same mechanism with primary fallback |
| 3 | Per-repo default branches | 🔶 | Global `branchRouting` works; **no per-repo `defaultBranch` in config schema** |
| 4 | Periodic workspace repo sync | 🔶 | Bootstrap-time sync exists; **no `setInterval` periodic re-sync or `workspace-sync.mjs`** |
| 5 | PR uses correct repo slug | 🔶 | Slug resolved per-task + env vars set; **`gh pr create` lacks `--repo` flag** |
| 6 | Backward-compatible single-repo | ✅ | Clean fallbacks when `repositories: []` |

**Remaining work (3 items):**
1. Add `defaultBranch` field to `repositories[]` config; wire into `resolveTaskBaseBranch()`
2. Add periodic workspace sync interval in monitor.mjs maintenance cycle
3. Pass `--repo ${executionRepoSlug}` to `gh pr create` when targeting secondary repos

### 83C: Codex Sandbox .git Resolution — 🔶 ~50% IMPLEMENTED (3/6 full, 1/6 partial, 2/6 missing)

| # | Criterion | Status | Gap |
|---|-----------|--------|-----|
| 1 | No "sandbox expects .git" errors | ✅ | `skipGitRepoCheck: true` at agent-pool.mjs:583,1958 |
| 2 | Worktree dirs have parent .git access | 🔶 | Parent dir added generically; **no `.git` file parsing for worktree refs** |
| 3 | Auto-detect full-write-access | ✅ | `disk-full-write-access` default; idempotent guard |
| 4 | `ensureGitAncestor()` for .git files | ❌ | **Function does not exist**; no `.git` file→gitdir resolution |
| 5 | Per-task writable roots cleanup | ❌ | **No `addTemporaryWritableRoot()`/cleanup**; TOML roots are permanent |
| 6 | Existing SDK launchers work | ✅ | All 3 launchers (Codex/Copilot/Claude) + dispatcher intact |

**Remaining work (3 items):**
1. Implement `ensureGitAncestor(dir)` — walk up, parse worktree `.git` files (`gitdir:` refs)
2. Implement `addTemporaryWritableRoot(path)` + cleanup in `executeTask()` post-completion
3. Add `worktreePath` parameter to `ensureSandboxWorkspaceWrite()`

**Note:** The `skipGitRepoCheck: true` workaround effectively mitigates the original error — agents CAN launch. The missing items are architectural hardening, not blockers.

### 83D: Workspace Health Dashboard — ❌ NOT IMPLEMENTED (0/7 full, 2/7 partial)

| # | Criterion | Status | Gap |
|---|-----------|--------|-----|
| 1 | `bosun workspace status` CLI | ❌ | No subcommand exists |
| 2 | `bosun workspace repair` CLI | ❌ | Worktree cleanup in maintenance.mjs but not exposed as CLI |
| 3 | `--json` output | ❌ | No JSON flag on workspace commands |
| 4 | UI Workspace tab | ❌ | Not in TAB\_CONFIG; no workspace tab file |
| 5 | REST API `/api/workspace/*` | ❌ | No routes exist |
| 6 | Telegram health alerts | 🔶 | Agent stuck alerts exist; no repo/worktree health alerts |
| 7 | Maintenance cycle health check | 🔶 | Worktree pruning exists via maintenance.mjs; no holistic health report |

**Foundation exists:** workspace-monitor.mjs (582 lines) tracks active agents for stuck detection; maintenance.mjs has `cleanupWorktrees()`. Neither constitutes the health dashboard described in 83D.

---

## Delta since 2026-02-11 (prior analysis)

### Summary

No chain module PRs merged since 2026-02-11. The 8 chain tasks (80A-82B) remain in `todo`. However, **83A is now fully implemented** — the workspace isolation blocker that prevented agent execution is resolved. Agents should now be able to execute chain tasks in workspace repos.

### Task execution status

**No chain tasks have been executed since the kanban migration** — all 8 chain tasks (80A-82B) remain in `todo`. However, the workspace isolation blocker (83A) is now **fully implemented**, meaning agents should be able to execute chain tasks going forward.

### Kanban state

11 draft (Bosun UI) + 12 todo (8 chain + 4 workspace) = 23 tasks. 83A should be marked `done`.

---

## CRITICAL BLOCKER: Workspace Isolation Failure

### Problem

Agents execute in `/home/jon/repos/virtengine` (the developer's active working directory) instead of `~/bosun/workspaces/virtengine/virtengine/` (the isolated workspace clone). This causes:

1. **Sandbox `.git` resolution failure** — Codex bwrap sandbox can't find `.git` when CWD is outside git repo
2. **Agent/developer conflicts** — agents and developer share the same git index, causing conflicts
3. **Worktree pollution** — worktrees created inside developer's repo pollute their working copy
4. **Non-functional workspace clone** — `~/bosun/workspaces/virtengine/virtengine/` exists but has NO `.git` (clone never completed)

### 12-Link Root Cause Chain (fully traced)

| # | File:Line | Issue |
|---|-----------|-------|
| 1 | `~/bosun/.env` | `REPO_ROOT=/home/jon/repos/virtengine` hardcoded |
| 2 | `repo-root.mjs:21` | `process.env.REPO_ROOT` checked first, always wins |
| 3 | `config.mjs:283` | `detectRepoRoot()` — env var first in priority |
| 4 | `config.mjs:831` | `repoRootOverride = process.env.REPO_ROOT` |
| 5 | `config.mjs:927` | `repoRoot = repoRootOverride \|\| selectedRepository?.path` — override wins |
| 6 | `monitor.mjs:533` | `process.chdir(repoRoot)` → developer's repo |
| 7 | `agent-pool.mjs:52` | `REPO_ROOT = resolveRepoRoot()` → developer's repo |
| 8 | `task-executor.mjs:3326` | `executionRepoRoot` → developer's repo |
| 9 | `worktree-manager.mjs:288` | worktrees under developer's `.cache/worktrees/` |
| 10 | `codex-config.mjs:1110` | sandbox writes only for developer's repo path |
| 11 | `monitor.mjs` (startup) | NO `pullWorkspaceRepos` call during daemon startup |
| 12 | `setup.mjs:4453` | Cloning only in interactive setup, not non-interactive |

### Resolution

Tasks 83A-83D created to fix this end-to-end (see Planned Tasks section below).

---

## Abandoned task branches (not merged, all stale)

| Branch | Ahead/Behind | Last Updated | Likely Task |
| --- | --- | --- | --- |
| ve/273a-xl-p0-feat-provi | +3 / -178 | 2026-02-12 | 72A: provider chain client |
| ve/6293-xl-p0-feat-provi | +2 / -166 | 2026-02-12 | 72B: provider chain submitter |
| ve/2bdf-xl-p1-feat-suppo | +8 / -173 | 2026-02-13 | 73C: support retention |
| ve/0f4b974d-xl-p1-feat-hpc-fairness | +2 / -179 | 2026-02-12 | 78C: HPC fairness |
| ve/c8939205-xl-p1-feat-economics | +0 / -184 | 2026-02-11 | 78D: economics |
| ve/8b0459ad-xl-p2-feat-ops-backup | +0 / -186 | 2026-02-12 | 79B: ops backup |
| ve/d5b90be5-xl-p2-feat-ibc | +0 / -186 | 2026-02-12 | 79A: IBC bridging |
| ve/161-security-remediation | +9 / -131 | 2026-02-14 | Security remediation |
| ve/753-testing-task-commit | local only | 2026-02-20 | Test task (agent debugging) |
| ve/759-test-task | local only | 2026-02-20 | Test task (agent debugging) |

All remote branches are stale (10+ days behind main) and must be recreated from scratch.

---

## Module Implementation Status (Full Audit 2026-02-22)

### Tier 1: Substantial Implementation (>10k LOC)

| Module | Source LOC | Test LOC | Status | Patent Coverage | Key Gap |
| --- | --- | --- | --- | --- | --- |
| x/veid | 70,048 | 42,775 | Complete | Claims 1-3, 12 | ML scorer defaults to stub; needs `mlruntime` build tag for real TF |
| x/market | 21,486 | 6,521 | Complete | Claims 1, 8, 11 | Query TODO stringers, pagination indexes |
| x/hpc | 20,474 | 5,655 | Complete | Claims 9, 13, 14 | Node-level proximity-weighted rewards implemented; SLURM automation and placement hardening remain |
| x/escrow | 18,754 | 5,798 | Complete | Claim 1 (billing) | None critical — mature module |
| x/mfa | 14,213 | 6,467 | Complete | Claims 1, 5 | None critical |
| x/settlement | 13,907 | 4,592 | Complete | Claims 1, 8 (billing) | Reward distribution uses placeholder recipient |

### Tier 2: Solid Implementation (3-10k LOC)

| Module | Source LOC | Test LOC | Status | Key Gap |
| --- | --- | --- | --- | --- |
| x/enclave | 7,553 | 2,660 | Complete | Heartbeat signature verification is placeholder |
| x/encryption | 7,308 | 5,534 | Complete | Ledger HID stub; post-quantum algo IDs are placeholders |
| x/delegation | 5,165 | 2,831 | Complete | None critical |
| x/fraud | 4,793 | 3,090 | Complete | Needs escrow/settlement integration |
| x/support | 4,488 | 1,526 | Complete | Retention queue is placeholder scan loop |
| x/benchmark | 4,235 | 1,261 | Complete | Hand-rolled proto stubs |
| x/staking | 3,827 | 1,699 | Complete | Deterministic stake-weighted rewards/slashing and delegator routing implemented |
| x/provider | 3,487 | 2,584 | Complete | None critical |

### Tier 3: Moderate Implementation (1-3k LOC)

| Module | Source LOC | Test LOC | Status | Key Gap |
| --- | --- | --- | --- | --- |
| x/config | 2,810 | 1,359 | Complete | Single test file |
| x/roles | 2,785 | 1,657 | Complete | Simulation genesis not implemented |
| x/deployment | 2,589 | 1,612 | Complete | None critical |
| x/review | 2,453 | 1,861 | Complete | Needs fraud module integration |
| x/audit | 1,992 | 754 | Complete | Needs cross-module hook wiring |
| x/oracle | 1,850 | 607 | Complete | No external oracle data ingestion |
| x/resources | 1,718 | 241 | **Partial** | **Lowest test coverage; minimal keeper** |
| x/cert | 1,518 | 1,209 | Complete | None critical |
| x/bme | 1,208 | 674 | Complete | Needs settlement/offramp integration |
| x/take | 508 | 833 | Complete | None — well tested |
| x/marketplace | 292 | 0 | Thin wrapper | Delegates to market; no own tests |

### Provider Daemon (`pkg/provider_daemon/`)

| Component | LOC | Status | Key Gap |
| --- | --- | --- | --- |
| Bid Engine | 1,005 | **Real** | None |
| Chain Client | 796 | **Mostly Real** | `GetProviderConfig` stub; HPC subscriptions noop |
| Chain Submitter | 840 | **Partial** | **`BroadcastTx` is a no-op** — cannot actually submit txns |
| Usage Meter | 667 | **Real** | None |
| K8s Adapter | 1,027 | **Real (interface)** | Needs concrete K8s client |
| OpenStack Adapter | 2,245 | **Real (interface)** | Needs concrete client |
| AWS Adapter | 2,451 | **Real (interface)** | Same |
| VMware Adapter | 1,968 | **Real (interface)** | Same |
| Ansible Adapter | 1,104 | **Real (exec)** | Runs actual ansible-playbook |
| Waldur Bridge | 1,000 | **Real** | Minor reconciler placeholder |

### ML Inference (`pkg/inference/`)

| Component | LOC | Status |
| --- | --- | --- |
| Scorer | 400 | **Real** — feature extraction, confidence, fallback |
| Determinism | 346 | **Real** — CPU-only, fixed seed, input/output hashing |
| TF Runtime | 702 | **Real** via sidecar (build tag `mlruntime`) |
| TF Runtime Stub | 524 | **Dev stub** (default build) |
| TF Serving Client | gRPC | **Real** — connects to TF Serving/ONNX |

### App Wiring: **Complete** — all 30+ modules registered in `app/app.go`, `app/modules.go`

---

## Patent Claims Coverage Assessment

| Claim | Description | Coverage | Key Gap |
| --- | --- | --- | --- |
| 1 | Core system (ID + auth + encrypt + cloud + billing) | **85%** | Billing reconciliation end-to-end untested |
| 2 | Mobile app for biometric capture | **70%** | Mobile app exists (`mobile/veid-capture-app/`) but integration untested |
| 3 | ML/AI identity scoring | **90%** | ML scorer fully implemented; defaults to stub without build tag |
| 4 | DEX services (crypto→fiat) | **40%** | Settlement offramp logic exists; no DEX integration |
| 5 | Multiple auth options (ledger, MFA, SSO) | **90%** | MFA module comprehensive |
| 6 | Encryption via third-party keys | **95%** | X25519-XSalsa20-Poly1305 envelope fully implemented |
| 7 | Proof-of-Stake consensus | **80%** | Deterministic rewards/slashing landed; resource-capacity integration remains |
| 8 | Cloud marketplace for computing | **75%** | Modules exist; provider can't submit txns (BroadcastTx no-op) |
| 9 | HPC pre-configured workloads | **80%** | HPC job templates exist; SLURM cluster automation incomplete |
| 10 | System combining 2+ of: ID, Cloud, HPC | **80%** | All three exist but integration untested |
| 11 | Method for cloud computing via blockchain | **70%** | Order→bid→lease flow exists; on-chain settlement incomplete |
| 12 | Method for decentralized identification | **90%** | VEID module is most complete |
| 13 | Decentralized supercomputer | **72%** | Node-level settlement rewards implemented; SLURM automation remains incomplete |
| 14 | Mini supercomputer clustering | **50%** | Topology-aware scheduling exists; placement engine incomplete |

---

## Critical Path to Patent Compliance

### Infrastructure Blocker (P0 — Blocks all agent execution)

0. **~~Workspace isolation failure~~ (83A ✅ RESOLVED, 83B/83C gaps remain)** — Core isolation (83A) is fully implemented: agents now use `~/bosun/workspaces/<ws>/<repo>` via `resolveAgentRepoRoot()`. Sandbox `.git` error mitigated via `skipGitRepoCheck: true`. Remaining 83B gaps: per-repo branch routing, periodic sync, PR `--repo` flag. Remaining 83C gaps: `ensureGitAncestor()`, per-task writable root cleanup. **These are hardening items, not blockers — chain tasks CAN proceed.**

### Chain Blockers (P0 — Must fix before any demo/audit)

1. **Chain Submitter BroadcastTx is a no-op** (pkg/provider\_daemon/chain\_submitter.go:461) — The entire marketplace cannot function if providers can't submit transactions
2. **Resources capacity integration remains incomplete** — staking economics are implemented, but downstream resource-capacity coupling is still pending for Claim 7 completeness
3. **Proto generation pipeline not automated** — 10 modules use hand-rolled stubs instead of generated code from existing .proto defs in sdk/proto/node/

### High Priority (P1 — Required for completeness)

4. **Resources module minimal** — 1.7k LOC, 241 test LOC, no capacity reservation
5. **HPC SLURM automation and placement hardening remain incomplete** — node-level rewards landed in `x/hpc/keeper/settlement.go` (PR #772)
6. **Oracle has no external data ingestion** — Price feeds defined but nothing pushes data in
7. **Fraud/review not integrated with escrow** — Dispute lifecycle exists but doesn't hold/release escrow
8. **End-to-end marketplace flow untested** — Individual modules work but no integration proof

### Planned Tasks (80A-83D) — Bosun task store

#### Phase 0: Workspace Isolation (P0 — blocks all agent execution)

| Order | Bosun ID | Title | Priority | Dependencies | Status |
| --- | --- | --- | --- | --- | --- |
| 83A | `91da8a7e` | feat(cli): workspace isolation — agents use ~/bosun/workspaces/ clone instead of developer repo | P0 | — | **✅ done** (8/8 criteria verified) |
| 83B | `5da8ca6a` | feat(cli): multi-repo task routing — worktrees per-repo with workspace-scoped execution | P0 | 83A | **🔶 ~70%** (3 gaps: per-repo branches, sync interval, PR --repo) |
| 83C | `412f8c9b` | feat(cli): Codex sandbox .git resolution + bwrap workspace mount boundaries | P0 | 83A | **🔶 ~50%** (3 gaps: ensureGitAncestor, temp roots, cleanup) |
| 83D | `efa475a5` | feat(cli): workspace health dashboard + diagnostic CLI | P1 | 83A, 83B | **❌ not started** (0/7 criteria) |

#### Phase 1: Chain Integration (P0 — parallel after 83A, 83C complete)

| Order | Bosun ID | Title | Priority | Dependencies | Status |
| --- | --- | --- | --- | --- | --- |
| 80A | `3ef33106` | feat(provider): chain submitter BroadcastTx + TxBuilder + chain client completion | P0 | 83A, 83C | **todo** |
| 80B | `84d2482e` | feat(staking): real stake-weighted rewards/slashing + delegation economics + resources capacity | P0 | 83A, 83C | **🔶 in progress** (staking economics complete; resource-capacity follow-up remains) |
| 80C | `1d5c01be` | feat(market): proto generation pipeline + query stack + end-to-end marketplace flow | P0 | 83A, 83C | **todo** |

#### Phase 2: Module Completion (P1 — parallel after Phase 1)

| Order | Bosun ID | Title | Priority | Dependencies | Status |
| --- | --- | --- | --- | --- | --- |
| 81A | `d163672e` | feat(hpc): SLURM integration + placement engine + SLA enforcement hardening | P1 | 80B | **todo** |
| 81B | `fefabac4` | feat(oracle): price feed ingestion + BME token ops + settlement FX offramp | P1 | 80C | **todo** |
| 81C | `f81211ee` | feat(fraud): dispute lifecycle + escrow integration + enclave attestation + review wiring | P1 | 80A | **todo** |

#### Phase 3: Cross-cutting (P1 — sequential after Phase 2)

| Order | Bosun ID | Title | Priority | Dependencies | Status |
| --- | --- | --- | --- | --- | --- |
| 82A | `8a8ff2d5` | feat(ibc): cross-chain settlement bridging + rate limits + interchain accounts | P1 | 81B | **todo** |
| 82B | `33a99163` | feat(roles): admin workflows + audit hooks + governance param updates + cert rotation | P1 | 81C | **todo** |

---

## Completed Tasks (historical, pre-kanban migration)

| Order | Title | Completed |
| --- | --- | --- |
| 67B | feat(settlement): CLI commands + HPC-settlement unification + escrow rec. | 2026-02-11 |
| 66D | feat(provider): KubernetesClient + chain submitter + key persistence | 2026-02-10 |
| 66C | feat(settlement): delegation-weighted rewards + EndBlocker | 2026-02-10 |
| 66B | feat(staking): MsgServer + Begin/EndBlocker | 2026-02-10 |
| 64C | ops: observability/telemetry pipeline for chain + provider daemon | 2026-02-10 |
| 58A | feat(encryption): enforce encrypted payload standard | 2026-02-10 |
| 56A | test(veid): end-to-end registration to verification tests | 2026-02-10 |
| 55C | feat(settlement): dynamic GPU fee burn + reward multipliers | 2026-02-09 |
| 53D | fix(security): bulk CodeQL/gosec cleanup + lint tuning | 2026-02-09 |
| 52B | fix(security): path traversal remediation + file path validator | 2026-02-09 |
| 52A | fix(security): command execution allowlist + exec.Command hardening | 2026-02-09 |
| 50B | feat(settlement): partial refunds + dispute workflow + payout adjustments | 2026-02-09 |
| 47B | feat(mfa): trusted browser scope to reduce MFA requirement | 2026-02-09 |
| 44B | feat(portal): provider onboarding wizard with adapter setup | 2026-02-09 |

---

## Execution Priority Sequence

Phase 0 (✅ RESOLVED): **83A done** → 83B, 83C have remaining gaps (hardening, not blocking)
Phase 1 (P0 parallel — **NOW UNBLOCKED**): 80A, 80B, 80C — Chain integration, staking economics, proto pipeline
Phase 2 (P1 parallel): 81A, 81B, 81C — HPC SLURM/placement hardening, oracle/BME, fraud/enclave
Phase 3 (P1 parallel): 82A, 82B — IBC bridging, roles/audit/governance

**83A is complete.** Agents now resolve to workspace repos via `resolveAgentRepoRoot()`. Chain tasks 80A-82B can proceed. The 83B/83C remaining gaps (per-repo branch config, periodic sync, ensureGitAncestor, temp writable roots) are hardening items that can be addressed in parallel with chain work.
