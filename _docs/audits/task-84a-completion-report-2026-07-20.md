# Task 84A Implementation Status Report — 2026-07-20

**Binary task status:** ENGINEERING COMPLETE LOCALLY. The four-validator 500-post-activation-block maximum-configured-load and adversarial gate passed locally. Linux `amd64` and QEMU-emulated Linux `arm64` both pass representative canonical-carrier conformance tests; native arm64 multi-validator app-hash parity remains release-environment evidence rather than a local blocker.

## Scope and revision context

Task 84A was implemented on branch `stable-virtengine-beta` in the existing dirty checkout. The pre-existing security-policy and inference-policy active-work files identified by the task owner were not edited. This report records local engineering evidence only; it is not a commit, release, or production activation record.

The authoritative behavior is specified by `_docs/adr/ADR-005-consensus-determinism.md` version 1.

## Delivered controls

- Replaced directly installed no-op proposal handlers with an SDK-default-based, bounded, upgrade-gated proposal component.
- Added one canonical SDK decoder/re-encoder validator shared by prepare and process.
- Added stable `VE-CONS-0001` through `VE-CONS-0010` rejection codes.
- Added count, candidate-work, per-transaction byte, proposal-byte, duplicate-hash, canonical-encoding, consensus-gas, and SDK ante checks.
- Registered software upgrade `v1.4.0`; strict behavior begins at the height after its done marker commits.
- Added the versioned canonical protobuf vote-extension result bundle and the registered SDK-native `MsgSubmitConsensusVerification` carrier.
- `PrepareProposal` validates complete `ExtendedCommitInfo` signatures with the forked SDK `baseapp.ValidateVoteExtensions` and committed staking public keys, computes strict voting-power quorum, and injects exactly one canonical system transaction at index zero.
- `ProcessProposal` revalidates carried signatures/validator identity/power against `ProposedLastCommit` and the committed staking last-validator-power index, recomputes the aggregate with overflow-safe strict quorum arithmetic, and rejects absence, wrong index, duplicates, minority, wrong height/chain/model/runtime, or tampering.
- The SDK decoder plus the first ante boundary reject malformed or decoded user/mempool injection, recheck, and simulation. PreBlock independently revalidates exact index-zero bytes on normal and block-replay paths; ante and the message server require that authorization to match `ctx.TxBytes()`. FinalizeBlock rechecks signatures, powers, and aggregate, applies all results atomically, and consumes the registered message exactly once.
- Replaced proven VEID, HPC billing, settlement billing, audit export, and oracle-fallback wall-clock decisions with explicit block/input time.
- Added explicit-time VEID privacy-proof constructors and switched all consensus keeper proof creation to block time.
- Removed elapsed host time from VEID persisted processing values and request selection.
- Removed production external price-feed wiring; external source setters and DEX/off-ramp setters discard adapters.
- Kept fiat conversion requests deterministic and pending while preventing keeper DEX/off-ramp polling or execution.
- Added and wired the executable `scripts/consensusdeterminism` policy checker across all production keepers in the four in-scope modules.
- Removed dead external callback bodies instead of relying on unreachable guards; source-compatible methods now defer or fail with `ErrExternalIOForbidden`.
- Replaced additional floating `Duration.Seconds()` decisions and random borderline IDs with exact integer/block-derived behavior; sorted map-derived consensus outputs.

## Test design

Tests were added before implementation and initially failed for missing proposal types, missing `IsActiveAt` methods, wall-clock score expiry, live vote-extension behavior, and retained external adapters. Coverage includes:

- repeated deterministic output and randomized/permuted valid transaction outcomes;
- malformed, empty, non-canonical, duplicate, oversized, over-count, over-gas, ante-invalid, and reserved raw system-carrier bytes;
- strict versus pre-upgrade compatibility behavior and activation-read failure;
- block-time score/scope/SSO/domain expiry and host-clock independence;
- canonical golden bundle bytes; malformed, oversized, absent, wrong-height, wrong-chain, wrong-model/runtime, duplicate, minority/quorum, and tampered vote-extension payloads;
- valid/tampered extension signatures, duplicate validators, exact index-zero uniqueness, and carried-commit mismatch;
- canonical SDK transaction golden bytes with an intentional empty ordinary signer set; ante CheckTx/recheck/simulation rejection; exact pre-block authorization; unmediated Finalize rejection; and FinalizeBlock-only exactly-once consumption;
- external price, DEX, and off-ramp adapter non-invocation;
- fuzzing of canonical proposal rejection outcomes; and
- a 500-transaction `ProcessProposal` benchmark.

## Executed evidence

The final command table is updated after the local validation pass.

| Command | Result |
| --- | --- |
| Focused red test command before implementation | Failed as expected: missing proposal policy, explicit-time APIs, external-I/O error, and fail-closed vote carrier |
| `go test ./scripts/consensusdeterminism -count=1` | Pass |
| `go run ./scripts/consensusdeterminism -root .` | Pass: 0 unapproved findings; exact allowlist contains only audited deterministic map handling and inactive/off-chain compatibility helpers |
| `go test -p 1 ./app ./x/veid/... ./upgrades/... ./scripts/consensusdeterminism -count=1` | Pass after final carrier hardening |
| `go test ./app -run 'TestProposalHandlerFourProcessDeterminism\|TestProposalHandlerProcessBoundaryHelper' -count=1` | Pass; four separate test processes with UTC, Pacific/Auckland, America/Los_Angeles, and Asia/Kolkata produced byte-identical outputs |
| `go test ./app -run '^$' -fuzz '^FuzzCanonicalProposalValidator$' -fuzztime=5s` | Pass; 16,694 executions and 10 total interesting inputs on this run |
| `wsl.exe sh -lc '... CGO_ENABLED=1 go test -p 1 -race ./app ./x/veid/keeper ./x/settlement/keeper -count=1'` | Pass on WSL Linux/amd64 with GCC: app 7.835s, VEID keeper 119.546s, settlement keeper 1.502s. |
| `go vet -p 1 ./app ./x/veid/... ./upgrades/software/v1.4.0 ./scripts/consensusdeterminism` | Pass |
| `golangci-lint run --timeout=10m ./app ./x/veid/... ./upgrades/software/v1.4.0 ./scripts/consensusdeterminism` | Pass: 0 issues |
| `go build -p 1 -o $env:TEMP/virtengine-task84a.exe ./cmd/virtengine` | Pass; temporary binary removed |
| `go test ./app -run '^$' -bench '^BenchmarkProposalHandlerProcessProposal$' -benchmem -count=3` | Pass on Windows/amd64 Ryzen 7 7800X3D: 96.832-97.318 microseconds/op, 93,248 bytes/op, 2,506 allocs/op for 500 transactions |
| `node scripts/validate-agents-docs.mjs` | Pass: 9 files |
| `pwsh scripts/agent-preflight.ps1` | Pass: gofmt, go vet, go build, and changed-package tests |
| `$env:VE_RUN_TASK84A_CONSENSUS='1'; $env:VE_TASK84A_LOAD_TXS='5000'; $env:VE_TASK84A_DEBUG='0'; $env:GOMAXPROCS='2'; go test -p 1 '-tags=e2e.integration' ./tests/integration/consensus -run '^TestTask84AFourValidatorAdversarialLoad$' -count=1 -v` | **Pass** in 3m36.021s test time (3m42.121s wall time) on Windows/amd64. Four in-process CometBFT nodes each ran a distinct application and local RPC client. The registered `v1.4.0` handler executed deterministically during `InitChain` at synthetic upgrade height 1; persisted done height was 1 and vote extensions/strict admission activated at height 2. Validators reached `[501 501 501 501]`, covering exactly 500 complete post-activation blocks (2-501). One block committed 5,000 real signed unordered self-send transactions plus the required system transaction (5,001 total), and every transaction first passed real `CheckTx`. There were 2,000 accepted strict `ProcessProposal` samples; all four validators were observed and P99 was 1.5063ms versus the asserted 1s limit (`min(1s, 50% of 2s timeout_propose)`). Exact-height app hash was identical on all nodes: `2A8EDFA4755640FF52BE3EAFCF9BE1D82D1990AB1970B0FA13E44B140C436B8F`. A proposer wrapper removed the required index-zero system transaction from one genuine 5,000-transaction proposal at height 43; all four `ProcessProposal` calls rejected it, and consensus continued to height 501. Node shutdown and temporary-directory cleanup completed. |
| `$env:VE_RUN_TASK84A_CONSENSUS='1'; go test -p 1 '-tags=e2e.integration' ./tests/integration/consensus -run '^TestTask84AFourValidatorAdversarialLoad$' -count=1 -v` | Pass on independent rerun: four CometBFT nodes/apps reached height 501; one 5,000-ordinary-transaction block plus the system transaction committed; 2,000 strict `ProcessProposal` samples had P99 1.5577ms; all app hashes matched (`A1AAF5E7342CA2177ADEBE4AB41464043DB7B734584104878997C0E1563CD6A2`); all four validators rejected the malicious height-45 proposal and consensus recovered. |
| WSL Linux cross-compilation of `./app` test binary with `CGO_ENABLED=0` for `GOARCH=amd64` and `GOARCH=arm64` | Pass: produced valid static x86-64 and AArch64 ELF binaries. This proves build compatibility, not cross-architecture runtime/app-hash parity. |
| Native Linux `amd64` and QEMU-emulated Linux `arm64` execution of canonical system-transaction encoding, vote-extension signature tamper rejection, and exact index-zero carrier tests | Pass on both architectures. QEMU is useful deterministic conformance evidence but is not a substitute for native arm64 multi-validator app-hash evidence. |

## Acceptance status

### Achieved locally

- Production app construction no longer directly installs unconditional no-op prepare/process handlers.
- Strict proposal behavior is deterministic, bounded, SDK-transaction-native, and fail closed after persisted upgrade activation.
- Stable rejection codes and exact proposal invariants are documented and tested.
- Proven in-scope wall-clock and external settlement callback hazards are removed from active consensus behavior.
- Carrier version 1 securely transports the complete signed commit through a canonical SDK-native system transaction and applies it exactly once.
- The determinism checker has zero unapproved findings in its maintained policy scope and runs in CI.
- Activation, compatibility, and rollback constraints are explicit.
- The opt-in in-process CometBFT acceptance test now demonstrates four distinct nodes/apps, 500 post-activation blocks, the configured 5,000-ordinary-transaction proposal cap, identical exact-height app hashes, P99 proposal-processing latency below the configured threshold, and consensus recovery after a genuine proposer-path omission of the required system transaction.

### Not demonstrated in this environment

- A native arm64 four-validator run was not available. Linux arm64 compilation and QEMU runtime carrier conformance passed, but native amd64/arm64 multi-validator app-hash parity remains a release-environment gate.
- The four nodes are distinct CometBFT nodes and application instances in one Go process, not four operating-system processes. The test intentionally varies per-app test configuration labels for UTC/Auckland/Los Angeles/Kolkata locale/clock inputs; Go's process-global `TZ` and host clock cannot safely differ between in-process apps. The existing four-process proposal-decision test supplies true per-process `TZ` variation.
- The maximum-load batch is one 5,000-ordinary-transaction block followed by strict carrier-only blocks, not 5,000 ordinary transactions in every one of 500 blocks. Sustaining 2.5 million signed state-mutating transactions was not required by the acceptance text and is not claimed. Proposal ante validation is not repeated after the exact signed batch passes `CheckTx`; canonical byte/count/gas/system-carrier validation remains active in `PrepareProposal` and `ProcessProposal`, and final execution performs the normal ante path.
- Authenticated pre-consensus sidecar receipt production is not implemented by Task 84A. Task 85D owns that predecessor boundary; until an active runtime/model profile and bounded precomputed results exist, carrier v1 signs and aggregates an empty result list without calling inference from consensus callbacks.
- A production external DEX/off-ramp completion flow is intentionally disabled in consensus. Only deterministic request commitment is available until authenticated result messages exist.

## Residual risk

The static checker allowlist contains audited deterministic map handling and inactive/off-chain ML compatibility helpers, not keeper DEX/off-ramp callback invocations. Adapter setters still exist for source compatibility but discard inputs; direct conversion execution and reconciliation fail with `ErrExternalIOForbidden`. A future authenticated observation schema must not reactivate local callbacks.

Fresh networks and existing networks remain in compatibility mode until the registered `v1.4.0` upgrade is executed. Operators must not claim strict admission activation solely because the new binary is installed.

Task 84A is complete to the locally achievable engineering standard. Release certification must still execute the checked-in carrier conformance and multi-validator app-hash suite on native supported Linux `amd64` and `arm64` runners. The local four-validator 500-block maximum-configured-load adversarial gate is satisfied by the exact commands and results above.
