# Task 84C Completion Report — Canonical Market and Unified Reservations

## Status

**LOCALLY COMPLETE after senior upgrade/compatibility re-audit.** Task 84C now has one mutable financial market owner, one authoritative capacity ledger, deterministic migration/quarantine behavior, additive generated contracts, and passing targeted local quality gates. Release/process evidence remains explicitly open below.

The work was performed in the existing mixed checkout at `99973c3af6c084a915e2913016d5d08899dcdaa6` on `stable-virtengine-beta`. No commit, push, stage, unstage, reset, checkout, revert, amend, or other git-index/history mutation was performed.

## ADR decision

ADR-007 selects `x/market` as the only mutable Order→Bid→Lease and financial lifecycle owner. The evidence matrix compared both registered lifecycles before naming an owner:

- `x/market` has the deployed generated market contracts, escrow payment lifecycle, deployment/manifest IDs, provider-daemon consumers, portal/SDK clients, and mature store migration history.
- `mktplace` has richer offering, VEID/MFA/provider-policy, encryption, and Waldur-oriented catalog semantics but no equivalent escrow/deployment financial authority.
- Promoting `mktplace` would require replacing market, escrow, deployment, provider-daemon, portal, SDK, and historical lease contracts.
- The representative checked-in mainnet genesis contains both modules but has zero market orders/bids/leases, zero non-owner orders/bids/allocations, zero resources allocations/reservations, and zero HPC jobs.

`mktplace` is retained as a supply/offering and compatibility catalog. Historical genesis with the additive activation flags absent/false remains valid for startup and pre-upgrade replay; the `v1.6.0` migrations activate stable rejection for independent order, bid, allocation, callback, usage-request, legacy resource-allocation, and allocation-lifecycle writes. New chains initialized from the current binary's default genesis start directly in the current protocol with both flags true, and explicit post-upgrade exports preserve true flags. The upgrade reconciles legacy state before activating the fences. Historical records remain queryable and ambiguous upgrade records are quarantined.

## Implemented protocol

### Authoritative reservation aggregate

`x/resources` now owns a generated, versioned `Reservation` aggregate with stable request, requester, provider, inventory, consumer, canonical order/bid/lease, HPC job, escrow, collateral, legacy source, and legacy reference lineage.

The transition table is explicit:

- none → Pending via reserve;
- Pending → Active via activate/link;
- Active → Consumed;
- Pending/Active/Consumed → Released;
- Pending → Expired through ordered time/height indexes;
- Pending/Active/Consumed → Quarantined;
- Pending/Active/Consumed/Quarantined → Slashed.

Released, Expired, and Slashed are final. Exact retries return the original reservation; same idempotency key with a different payload fails. Reservation IDs are deterministic request-digest identifiers rather than caller-controlled idempotency keys. Capacity arithmetic is checked signed-integer arithmetic with no floating-point consensus decisions. GPU type requirements fail closed.

### Market and HPC orchestration

Canonical lease creation reserves capacity and activates the reservation in one cached Cosmos context with escrow and market mutations. Close paths propagate reservation release failures and use cached contexts, preventing silent partial lifecycle mutation.

Standalone HPC submission reserves, activates, and persists one dedicated reservation atomically. Queued/running transitions require a valid active reservation. Terminal standalone jobs release exactly once; disputes quarantine. A market-backed HPC job reuses, rather than duplicates, its active canonical lease reservation after provider, requester/customer, lease, and capacity validation. Market remains the release/quarantine owner for shared lease reservations.

Fresh genesis validation rejects active market leases or executable HPC jobs without reservation lineage. Resources genesis rebuilds indexes and immediately checks capacity conservation.

### Expiration, invariants, and observability

Pending reservation expiration uses ordered big-endian block-time and block-height indexes and processes at most 1,000 due entries per EndBlock. Provider heartbeat expiry expires pending reservations and quarantines active/consumed reservations.

Registered and executable checks cover:

- available plus nonterminal reserved capacity equals declared inventory total;
- no negative/overflowing dimensions;
- unique active consumer ownership;
- final-state non-reactivation;
- active market lease lineage;
- executable HPC provider/customer plus job-or-lease lineage;
- no mutable non-owner lifecycle after activation.

Stable bounded transition events and generated gRPC/gateway/OpenAPI/TypeScript queries expose reservation-by-ID, order, bid, lease, job, consumer, provider, and complete transition lineage without workload/manifests.

## Upgrade and reconciliation

Software upgrade `v1.6.0` runs module migrations:

| Module | Version |
| --- | --- |
| market | 7 → 8 |
| mktplace | 1 → 2 |
| resources | 1 → 2 |
| hpc | 1 → 2 |

Migration rules are deterministic:

- completed records remain preserved;
- already linked records remain linked only when state, provider, requester, consumer, and lineage agree; inconsistent references are quarantined;
- unambiguous derived indexes are rebuilt;
- ambiguous active market, HPC, resources, or non-owner order/bid/allocation records are quarantined;
- quarantined ambiguous records claim zero authoritative capacity until operator reconciliation;
- migration never guesses or synthesizes capacity;
- exact migration retry does not duplicate reservations.

The final senior FIX pass additionally closed concrete compatibility gaps: dangling or payload-inconsistent pre-linked resource allocations now become deterministic zero-capacity quarantine instead of bypassing migration; valid pre-linked allocations are counted without duplication; pending legacy allocations receive deterministic expiry; all migrated states emit lineage evidence; terminal reservations rebuild query/idempotency indexes at genesis; reservation events survive export/import; lineage events are paginated and malformed events fail closed; and open non-owner bids are validated, reconciled, and covered by the post-upgrade invariant alongside orders and allocations.

The handler reconciles cross-module state before the `mktplace` activation migration. Explicit module order is `resources -> market -> hpc -> mktplace`, followed by remaining modules with `auth` last. Active/pending resource allocations link only when one inventory matches and `available + all linked assignments = total`; otherwise migration records zero-capacity quarantine. A real-app handler test verifies versions, activation, post-upgrade invariants, premature-activation rejection, and retry idempotence.

Checked-in mainnet genesis preflight result:

Machine-readable evidence is stored in `artifacts/mainnet/task84c-reconciliation.json` with a repository-relative source path and binds these counts to genesis SHA-256 `a8d8a4a4f19882503265482c9433a6646d8dbbfe62f81c5945e81c32da9e6032`.

| Surface | Count |
| --- | ---: |
| market orders / bids / leases | 0 / 0 / 0 |
| non-owner offerings / orders / bids / allocations | 0 / 0 / 0 / 0 |
| resources inventories / allocations / reservations | 0 / 0 / 0 |
| HPC jobs | 0 |

Rollback below `v1.6.0` after activation is unsupported because older binaries permit competing lifecycle owners and ignore reservation indexes.

## Generated contract evidence

The pinned Task 86A WSL generator was run by output group after the final schema correction. Go, descriptor, OpenAPI, and TypeScript outputs were regenerated. The final repeated TypeScript/inventory pass reported `TASK84C_GENERATION_DRIFT=0`; TypeScript build/package-export validation passed under pinned Node 24.12.0.

- OpenAPI paths: 224.
- Descriptor methods: 418.
- HTTP bindings: 224.
- Inventory SHA-256: `58d5236fde8755aef2c7d4d2f67a77f3712b3356c8d319ccdfb036c21a3d83e3`.
- Descriptor SHA-256: `1cb9ae24edc689853e21215f0456e525515c993b81274051e269363a8b2f37be`.
- Generation verification: 3/3 inventory/route/module tests passed.
- TypeScript SDK generation build and package-export validation passed in the pinned WSL environment.

Compatibility tests use fixed historical field bytes, verify old fields decode with new reservation/activation/event fields absent, and guard additive reservation, genesis-event, and lineage-pagination field numbers.

## Changed-file inventory

Task 84C source and handwritten test/documentation surfaces:

- `_docs/INDEX.md`, `_docs/adr/ADR-007-canonical-market-reservations.md`, `_docs/audits/task-84c-completion-report-2026-07-21.md`, `_docs/protocols/resource-reservation-state-machine.md`, `_docs/runbooks/reservation-quarantine-remediation.md`, `_docs/ralph/progress.md`;
- `app/app.go`, `app/app_configure.go`, `app/app_configure_task84c_test.go`, `app/modules.go`, `app/types/app.go`, `app/types/reservation_eligibility.go`, `app/types/reservation_invariants.go`;
- `scripts/task84c-preflight.ps1`, `artifacts/mainnet/task84c-reconciliation.json`;
- `sdk/proto/node/virtengine/{hpc,market,resources}/...` and their generated Go, gateway, TypeScript, descriptor, inventory, checksum, and OpenAPI outputs;
- `tests/compatibility/task84c_wire_compatibility_test.go`, `tests/integration/marketplace/*`, `tests/upgrade/{registry_test.go,test-cases.json,workers_test.go}`;
- `upgrades/types/types.go`, `upgrades/upgrades.go`, `upgrades/upgrades_test.go`, `upgrades/software/v1.6.0/*`;
- `x/resources/{genesis.go,module.go,types/*,keeper/{allocation.go,grpc_query.go,inventory.go,keeper.go,migration.go,migration_test.go,reservation.go,reservation_test.go}}`;
- `x/market/{genesis.go,genesis_reservation_test.go,module.go,handler/{keepers.go,server.go,reservation_test.go},keeper/{keeper.go,migration.go},types/marketplace/{errors.go,genesis.go,keeper/*}}`;
- `x/marketplace/{genesis.go,module.go,module_test.go}`;
- `x/hpc/{genesis_test.go,module.go,types/{genesis.go,genesis_reservation_test.go,job.go},keeper/{keeper.go,migration.go,msg_server.go,reservation_test.go,scheduling.go}}`.

The checkout also retains protected pre-existing Task 84B completion/follow-up edits in its original layered staged/unstaged state, including settlement/provider process evidence, settlement event/immutability fixes, and repository preflight improvements. Task 84C did not stage, unstage, reset, or revert those files.

## Commands and results

Passed:

- `go test ./x/resources/... ./x/market/... ./x/marketplace ./x/hpc/... ./app ./app/types ./upgrades/software/v1.6.0 ./upgrades ./tests/compatibility ./tests/upgrade -count=1`
- `go test -tags='e2e.integration' ./tests/integration/marketplace -count=1`
- WSL race: `go test -race ./x/resources/keeper ./x/market/handler ./x/hpc/keeper -count=1`
- `go vet` over all affected Task 84C packages.
- `golangci-lint` over all affected Task 84C packages — `0 issues`.
- `go build ./cmd/virtengine ./cmd/provider-daemon`.
- `go run ./scripts/consensusdeterminism -root .` — `0` unapproved and `25` narrowly allowlisted findings.
- Pinned WSL protobuf Go/descriptor/OpenAPI/TypeScript/inventory full generation twice after the final source change; pre-run-to-pass-one drift `0` and pass-one-to-pass-two drift `0`.
- Pinned WSL generation verify — `3/3` tests passed.
- TypeScript SDK build/package exports in pinned WSL — passed.
- Git for Windows Bash `./scripts/verify-modules.sh` — all modules and policy passed.
- `node scripts/validate-agents-docs.mjs` — 9 files passed.
- `pwsh -NoProfile -File scripts/agent-preflight.ps1` — passed.
- `git diff --check` — passed.
- `pwsh -NoProfile -File scripts/task84c-preflight.ps1 -Genesis artifacts/mainnet/genesis.json -Output artifacts/mainnet/task84c-reconciliation.json` — counts and digest above with repository-relative source evidence.

Environment-qualified observation:

- The Windows TypeScript build cannot reuse `node_modules` populated by the pinned Linux/WSL generator because native `esbuild` is platform-specific. The authoritative pinned WSL TypeScript build passed.
- VS Code's broad `sdk/ts/tsconfig.json` project reports pre-existing `rootDir` diagnostics because it includes examples, scripts, and tests outside `src`; the package's authoritative `tsconfig.build.json` build and export validation pass.
- Native Windows Go race instrumentation is unavailable with CGO disabled; the focused WSL race suite passed.

## Acceptance matrix

| Requirement | Result | Evidence |
| --- | --- | --- |
| Evidence-first ADR compares both owners | PASS | ADR-007 evidence matrix and rejected alternatives |
| Exactly one mutable lifecycle owner | PASS | `x/market`; post-activation non-owner lifecycle writes rejected |
| Authoritative generated reservation aggregate | PASS | `x/resources` proto, keeper, genesis, queries, events |
| Market cannot create executable lease without reservation | PASS | cached reserve/escrow/lease/activate orchestration |
| HPC cannot become executable without active reservation | PASS | submission/scheduling/status guards and app wiring |
| Market and HPC cannot double-sell final capacity | PASS | randomized invariant, final-unit unit/app integration, shared reservation reuse |
| Terminal workflows release or quarantine once | PASS | market/HPC close, timeout, cancel, failure, dispute behavior |
| Indexed deterministic expiration | PASS | ordered time/height indexes and bounded EndBlock processing |
| Provider eligibility and optional profiles fail closed | PASS | app provider adapter; requested benchmark/attestation/collateral reject when unavailable |
| Migration handles linked/orphaned/duplicate/inconsistent state | PASS (local fixtures) | linked resources require exact inventory conservation; orphaned/inconsistent market/HPC/non-owner state is deterministic zero-capacity quarantine; real-app handler retry tested |
| Generated lineage queries and client contracts | PASS | ID/order/bid/lease/job/consumer/provider/lineage routes |
| Wire compatibility and deterministic generation | PASS | fixed old-wire decode fixtures, additive guards, zero generation drift |
| Capacity/collateral invariants | PARTIAL | capacity conservation is executable and randomized; collateral is fail-closed when requested, but a production collateral reader/profile remains later-task scope |
| Full process-boundary order→bid→lease→deployment→usage→release | RELEASE-ONLY | local tests cross real app keeper wiring but do not run that complete transaction workflow on separate validator/provider processes |

## Remaining release-only evidence

Task 84C has no remaining local code-quality gate failure. The following external/process evidence is intentionally not claimed:

- separately deployed multi-validator nodes executing real signed market and HPC transactions;
- a real provider daemon observing an order, bidding, winning a lease, provisioning infrastructure, running a workload, accounting usage, and cleaning up;
- live provider suspension, collateral-drop, benchmark/attestation expiry, and heartbeat-loss processes against governed production profiles;
- production Waldur/Kubernetes/SLURM lifecycle certification and long-duration oversubscription soak;
- Task 84D canonical dispute authority deciding whether quarantined shared lease capacity is released or slashed.

Those are release/testnet or later-task gates. They do not reopen the local canonical-owner decision, reservation state machine, capacity conservation, migration/quarantine behavior, or generated-contract compatibility delivered by Task 84C.
