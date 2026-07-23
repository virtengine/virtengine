# Task 84D Canonical Financial Cases Completion Report

**Date:** 2026-07-22  
**Branch:** `stable-virtengine-beta`  
**Checkout:** `99973c3af6c084a915e2913016d5d08899dcdaa6`  
**Commit:** not created by instruction  
**Status:** locally complete to available process boundary; release-only evidence is listed below

## Decision

ADR-008 selects `x/settlement` as the sole mutable financial-case owner. Evidence compared settlement, escrow billing disputes, fraud, HPC and review across stores, APIs, existing money authority, lineage and dependency direction. Settlement already owns authenticated usage, settlements, payout holds/refunds, claimable rewards and the ledger from which exact conservation can be enforced. Fraud, HPC, review and billing are adapters/projections after `v1.7.0`; they cannot independently release/refund/transfer financial value.

Normative artifacts:

- `_docs/adr/ADR-008-canonical-financial-cases.md`;
- `_docs/protocols/financial-case-state-machine.md`;
- `_docs/runbooks/financial-case-duplicate-quarantine-remediation.md`;
- `_docs/billing-policy.md`.

## Implementation

### Generated aggregate and API

Settlement v1 contracts now define:

- canonical subject types for order, invoice, usage, HPC job and settlement plus full alias lineage;
- repeated typed claims and privacy-safe evidence hash/encrypted reference;
- claimant/respondent, multi-denom escrow/payout/reward exposure, reservation hold, statuses/deadlines/resolver, terminal allocation, bounded appeals, migration/quarantine and append-only transitions/effects;
- explicit canonical provider/customer financial roles separate from claimant/respondent filing roles, preventing provider-filed cases from reversing allocation recipients;
- open, add-claim, submit-review, escalate, resolve, appeal, cancel and finalize messages with signer annotations;
- queries by ID, canonical subject, order, invoice, usage, job, escrow, status, party and paginated privacy-safe lineage;
- generated opened/claim-added/held/reviewed/escalated/resolved/appealed/finalized/effect/quarantined/expired event contracts.

All existing field numbers and message type URLs remain unchanged. Additions used new field numbers. Generated Go, gateway, OpenAPI, descriptor, inventory and TypeScript outputs are checked in.

Descriptor SHA-256: `38862f7cc92e1721d41585a243a65a664a1afaf363bc497cd8ab4f033fa733ea`.  
Inventory SHA-256: `ff2c8d463de137b1400fdc5d07aecbace87fae82788f938821e46eda287f7193`.  
Final zero-drift generated manifest SHA-256: `0e9c6f944e43d656786b3ec5f6ffb758afe421590220854c1fb8d1c43a7a0baa` (before = after; drift `0`).  
Generated OpenAPI SHA-256: `6004f5c8df7a90cafcc3b6ea26a44b9a82d80bc65d3a511342c75cc28299f595`.  
OpenAPI path count: 234.

### Lifecycle and holds

- Case/claim/appeal IDs are SHA-256 domain-separated length-prefixed canonical inputs; no wall time, random input or local sequence affects identity.
- Exact retries return original case/claim; a key reused with changed bytes fails.
- Any shared active order/invoice/usage/job/escrow alias merges into one root.
- Cached contexts establish payout `HELD`, settlement escrow `DISPUTED` and Task 84C reservation `DISPUTED` before case commit.
- Settlement, payout creation/execution/retry, escrow release/refund/expiry, claimable rewards, HPC accounting and HPC reward creation check active canonical indexes.
- Review/content collection, governance escalation, bounded appeal and deterministic timeout escalation leave value held.
- Resolver must be governance authority and cannot be claimant/respondent.
- Public filings prove subject-party ownership from payout/invoice/settlement/usage/escrow/reservation state; only application-wired adapters use the trusted boundary, so caller-supplied `source_module` cannot bypass authorization.
- Legacy settlement escrow-dispute messages project into canonical cases, and settlement/escrow payout release/refund/execute paths are fenced after activation.

### Conservation and exactly-once resolution

For every denomination:

$$provider + customer + platform + slashWitness = originalHeld$$

Allocations use `sdk.Coins`; invalid/negative/extra denominations and cross-denom netting fail before writes. Resolution enters `RESOLVED_PENDING_APPEAL`. Finalization persists deterministic provider/customer/platform/slash-witness/reservation/projection effect markers and executes in one cached context. A failed boundary commits neither transfer nor marker. A terminal retry observes `APPLIED` markers and performs no duplicate transfer. Fraud-confirmed resolution can terminally slash the held reservation; other finals release it. Cancellation restores the exact pre-dispute reservation state.

Explicit provider/customer fields route allocation roles independently of who filed. Reward exposure is added to `original_held` only when it is an independently claimable ledger balance, then consumed by the reward effect exactly once; payout-backed settlement value is not counted again as escrow exposure.

### Adapters

- Fraud financial reports create/merge `FRAUD` claims and project canonical ID/status. Resolution/escalation is a recommendation and canonical escalation request after activation.
- HPC `FlagDispute` stores a hash reference, creates/merges an `HPC` claim and returns canonical ID/status. `ResolveDispute` is an escalation request after activation. HPC accounting/rewards and capacity remain held.
- Review adds only content hash and bounded `rating:N` recommendation to an existing case; it never transfers value.
- Escrow billing initiation projects an invoice claim after activation. Legacy billing resolution is fenced and requests canonical escalation. Historical terminal workflows remain queryable.

## Migration and reconciliation

`v1.7.0` follows `v1.6.0` and registers:

- settlement `2→3`;
- fraud `1→2`;
- HPC `2→3`;
- review `1→2`;
- escrow `3→4`;
- resources `2→3`.

Settlement migration imports existing held payouts/disputed escrows first, preserves terminal finance, builds indexes/effect defaults and persists an idempotent report. The upgrade then adds pending fraud/HPC records as claims and activates non-owner fencing only after migration/invariants. Multiple aliases merge. Missing/ambiguous lineage stays held and quarantined; migration never releases value or fabricates allocation.

The final hardening pass also imports active escrow billing workflows, restores payout records and payout sequence through settlement genesis export/import, validates referenced payout/escrow holds at genesis, uses bounded digest idempotency for maximum-length source IDs, and propagates usage/reward failures instead of committing partial accounting state.

Checked-in representative reconciliation (`artifacts/mainnet/task84d-reconciliation.json`):

| Counter | Value |
| --- | ---: |
| payouts scanned | 0 |
| escrows scanned | 0 |
| cases created | 0 |
| claims merged | 0 |
| quarantined | 0 |
| terminal preserved | 0 |
| already migrated | 0 |
| malformed/orphan | 0 |

Counter digest: `f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b`.

This is representative empty/current-genesis evidence. No validator-approved mainnet export was supplied locally; release must rerun reconciliation against the exact approved export and publish actual counts/digest.

## Invariants

Registered settlement application/crisis invariant verifies:

- canonical case bytes, deterministic IDs and claim root;
- one active subject root;
- active value case has exact payout/escrow hold count;
- terminal case has no holds, conserved allocation and all effects applied;
- contiguous append-only transition sequence;
- subject/status/party/order/invoice/usage/job/escrow indexes resolve;
- malformed records fail closed.

Resources conservation includes disputed reservations and verifies case binding/prior state. Existing authenticated-usage and escrow-settlement invariants remain registered.

## Test evidence

TDD record:

- **RED:** focused financial-case tests failed on absent deterministic ID, state-machine, query/message and migration implementation.
- **GREEN:** kernel, holds, allocation, adapters, migration, queries and events made the suite pass.
- **REFACTOR:** active alias merging, sorted audit roots, terminal reservation effect and historical v1.6.0 target scoping were added without changing behavior outside activation.

Coverage includes state transitions, authorization, exact retry/conflict, bounded evidence, resolver conflict, appeal, timeout escalation, multi-denom conservation, terminal effect retry, migration activation/idempotence and 1,000 deterministic randomized allocations. Adapter package suites cover existing fraud/HPC/review/billing compatibility.

## Commands and results

| Command | Result |
| --- | --- |
| `go test ./x/settlement/... ./x/escrow/... ./x/fraud/... ./x/hpc/... ./x/review/... ./x/resources/... ./upgrades/software/v1.7.0 ./tests/upgrade -count=1` | PASS |
| `go test ./... -run '^$' -count=1` | PASS; all Go packages compile, including integration fixtures |
| `go test ./tests/compatibility -run Task84D -count=1` | PASS; frozen legacy dispute wire, additive case fields and type URLs |
| `go test ./upgrades/software/v1.6.0 ./upgrades/software/v1.7.0 ./tests/upgrade -count=1` | PASS; historical v1.6.0 remains at resources/HPC v2 |
| WSL `go test -race` on six affected keepers | PASS |
| `go vet` affected packages | PASS |
| `golangci-lint run` affected packages | PASS, 0 issues |
| `go build ./cmd/virtengine ./cmd/provider-daemon` | PASS |
| checked-in descriptor/inventory/OpenAPI hash and descriptor contract audit | PASS; signer annotations, field numbers, type URLs, REST routes, sidecars and 234 OpenAPI paths verified |
| pinned `proto-generate-wsl.sh all`, hash all outputs, repeat and diff | Previous implementation pass: PASS, zero drift; this FIX pass did not change schemas and therefore did not regenerate |
| WSL TypeScript `npm run build` | Previous implementation pass: PASS; this FIX pass added public export coverage, but local reruns were interrupted by the host terminal after startup |
| WSL TypeScript Jest with pinned Buf and 4 GiB heap | Previous implementation pass: PASS, 36/36 suites, 1101/1101 tests, 16 snapshots; this FIX pass did not claim a new full Jest completion |
| Type-check `sdk/ts/examples/financial-case.ts` | PASS |
| `go run ./scripts/consensusdeterminism` | PASS, 0 unapproved; 25 narrow allowlists |
| WSL `./scripts/verify-modules.sh` with pinned Node | PASS |
| `node scripts/validate-agents-docs.mjs` | PASS, 9 files |
| `pwsh scripts/task84d-preflight.ps1` | PASS |
| `pwsh scripts/agent-preflight.ps1` | PASS |
| `gofmt -l` changed Go files, `git diff --check`, conflict-marker scan | PASS |

Windows `go test -race` correctly reported that CGO was disabled; the same affected suite passed under WSL race. Initial Windows TypeScript build used Linux `node_modules` and failed platform detection; execution under pinned WSL was green. The first final TypeScript test attempt used an incorrect mixed Windows/WSL path; the next pinned run reached the default 2 GiB Node heap and failed out-of-memory. The authoritative retry used pinned Node 24.12.0, pinned Buf and the documented 4 GiB heap and passed all 36 suites/1101 tests. Module verification initially inherited Windows Node through WSL and was rerun successfully through Git for Windows Bash.

### Review/fix pass verification on 2026-07-22

This review pass fixed appeal replay reconstruction after genesis-style rebuild, strengthened subject-index and appeal-replay invariants, corrected the invoice-paid payout callback argument order, and made HPC terminal status reporting rollback when canonical reward distribution is blocked by an active financial case.

| Command | Result |
| --- | --- |
| `go test ./x/settlement/keeper -run "FinancialCase\|Audit\|Payout" -count=1` | PASS |
| `go test ./x/hpc/keeper -run "FinancialCase\|ReportJobStatus" -count=1` | PASS |
| `go test ./x/settlement -run "Genesis" -count=1` | PASS |
| `go test ./x/settlement/keeper -count=1` | PASS |
| `go test ./x/hpc/keeper -count=1` | PASS |
| `go test ./tests/compatibility -run "Task84D\|Wire\|Compat\|Financial" -count=1` | PASS |
| `go test ./x/settlement/... ./x/hpc/... ./upgrades/software/v1.7.0/... -count=1` | PASS |
| `go vet ./x/settlement/... ./x/hpc/... ./upgrades/software/v1.7.0/...` | PASS |
| WSL `CGO_ENABLED=1 go test -race ./x/settlement/keeper -run FinancialCase -count=1` | PASS |
| `golangci-lint run ./x/settlement/... ./x/hpc/... ./upgrades/software/v1.7.0/... --timeout 10m --allow-parallel-runners` | PASS, 0 issues |
| `go build ./cmd/...` | PASS |
| Git for Windows Bash `./scripts/verify-modules.sh` | PASS |
| `node scripts/validate-agents-docs.mjs` | PASS, 9 files |
| `pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/task84d-preflight.ps1` | PASS |
| `pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/agent-preflight.ps1` | PASS |
| WSL pinned proto hash/generate/check wrapper, pass 1 | PASS, zero hash drift; inventory tests and `git diff --check` passed |
| WSL pinned proto hash/generate/check wrapper, pass 2 | PASS, zero hash drift; inventory tests and `git diff --check` passed |
| Windows `npm -C sdk/ts run build` after lockfile `npm ci` | PASS |
| Windows `npm -C sdk/ts test -- --runInBand` after lockfile `npm ci` | PASS, 37/37 suites, 1102/1102 tests, 16 snapshots |
| `git diff --check` | PASS |

Docker-based `scripts/verify-proto-generation.sh` could not run because Docker Desktop's Linux engine pipe was unavailable locally. The WSL pinned generator path was used instead and completed two hash-checked zero-drift passes.

## Acceptance matrix

| Outcome | Local status |
| --- | --- |
| Evidence-based owner decision/spec/runbook | PASS |
| Generated aggregate, lifecycle messages, REST/gRPC and queries | PASS |
| Deterministic duplicate merge and deadlines/authorization | PASS |
| Atomic payout/escrow/reservation holds and execution guards | PASS |
| Multi-denom exact conservation and appeal-before-release | PASS |
| Persisted exactly-once effects/restart-safe retry semantics | PASS at keeper/store boundary |
| Fraud/HPC/review/billing adapters and post-activation fencing | PASS at module/keeper boundary |
| Privacy-safe hashes/references and bounded generated events | PASS |
| v1.7.0 migration, genesis, quarantine and reconciliation artifact | PASS with representative counts |
| Registered invariants and randomized property test | PASS |
| Pinned generation twice with zero drift | PASS |
| Affected tests/race/vet/lint/build/TS/modules/docs/preflight | PASS |

## Honest remaining release-only evidence

- No approved mainnet genesis/export was available, so actual production migration counts and digest remain a release operation.
- No separately running multi-validator network was started for Task 84D; Task 84B already supplies four-validator metering convergence, but this report does not relabel isolated keeper/app tests as a Task 84D multi-process balance proof.
- No destructive process kill was injected between real database commits. Cached-context tests and persisted effect-marker exact retries prove local/store convergence; a node process restart/failure campaign remains release certification.
- No production governance proposal, real customer/provider funded escrow or external reputation consumer was available. Release must execute low-value provider/customer/partial/fraud/timeout scenarios through signed transactions and independently reconcile balances/app hashes.
- These limitations do not change local protocol ownership, schema, deterministic migration, guards or invariants; they constrain claims to locally demonstrated boundaries.
