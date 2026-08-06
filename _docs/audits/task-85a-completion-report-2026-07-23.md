# Task 85A Completion Report - 2026-07-23

**Status:** Locally complete for the single-provider, single-active-submitter Task 85A boundary. Task 85B is next. Task 85C retains multi-replica fencing and distributed lease evidence.

## Implementation Summary

- Added a versioned generic provider mutation envelope and registry under `pkg/provider_daemon/provider_mutation.go`.
- Added durable queue storage in `pkg/provider_daemon/provider_mutation_store.go`, including atomic put-if-absent/update, path locking, independent envelope copies and the public `QueueStore` alias.
- Added `SubmitterLease` plus local single-process fencing for Task 85A and explicit Task 85C extension points.
- Routed production provider writes through `ProviderMutationSubmitter` from bid placement, usage/settlement submission, HPC status/accounting/node metadata, resource heartbeats, provider domain confirmation, Waldur callbacks and support writes.
- Removed production generated `MsgClient` mutation use from provider-daemon paths; disconnected mutation callers now receive `ErrProviderMutationUnavailable` or `ErrProviderMutationNotReady` instead of success.
- Added SDK direct-sign transaction building with KeyManager Ed25519 keys, account sequence resolution, gas simulation, gas adjustment, tx bytes/hash persistence, Comet RPC broadcast, inclusion confirmation and finality checks.
- Added reconciliation for crash/restart, response-lost, sequence mismatch, timeout, out-of-gas, mempool rejection, replacement/reorg and dead-letter outcomes.
- Tightened the partial implementation with checked gas adjustment/bump arithmetic, terminal-aware mutable idempotency, deterministic support metadata digests, lease-guarded state transitions, required tx hash/height/block-hash confirmation evidence, locked queue file replacement, and production fail-closed composition for callbacks, domain verification, HPC node metadata and legacy usage submission.
- Kept Task 84B authenticated usage proof allocation intact; `ChainUsageSubmitter` delegates final signed transaction delivery to the generalized mutation submitter when configured.

## Contract Tests

- Every default registered provider mutation kind encodes/decodes through the registry.
- Every default registered provider mutation kind is signed into a decodable SDK transaction by the configured provider key.
- Customer/owner-signed `MsgCreateLease` and `MsgCloseLease` fail closed as unknown provider queue kinds.
- Tests cover unavailable/readiness errors, dead letters, response-loss reconciliation, restart recovery, timeout ambiguity, reorg retry, lease loss during confirmation, missing/wrong confirmation evidence, terminal mutable update replay, active mutable idempotency conflicts, metadata-map deterministic digesting, store reopen persistence, gas overflow and bounded classification.

## Evidence

- `go test ./pkg/provider_daemon -run ProviderMutation -count=1` passed.
- `go test ./pkg/provider_daemon -count=1` passed.
- `go test -race ./pkg/provider_daemon -run ProviderMutation -count=1` passed under WSL. Native Windows race is unavailable on this host because cgo cannot find `gcc` on `PATH`.
- `go vet ./pkg/provider_daemon` passed.
- `golangci-lint run ./pkg/provider_daemon --timeout 10m` passed with `0 issues`.
- `go build ./cmd/provider-daemon` passed.
- `go build ./cmd/virtengine` passed.
- `node scripts/validate-agents-docs.mjs` passed.
- `pwsh scripts/task85a-preflight.ps1` passed, including WSL race fallback, vet, lint, command builds, AGENTS validator and direct `MsgClient` scan.
- `pwsh scripts/agent-preflight.ps1` passed for the dirty worktree's detected checks.
- Static search over `pkg/provider_daemon` and `cmd/provider-daemon` found no production generated `NewMsgClient` mutation calls. Remaining `BroadcastTx` references are in the durable submitter path and tests.

## Release-Only Gaps

- A long-lived real localnet/testnet run with committed state/events for every externally enabled provider profile is still release evidence, not a local Task 85A code blocker.
- Task 85C must replace the local lease with distributed fencing before multi-daemon/multi-replica provider operation.
- Some registry-ready operator/provider lifecycle mutations have no current production daemon caller; they are registered to avoid future direct MsgClient reintroduction.
- Full proto regeneration was not rerun for Task 85A because this task did not change proto contracts and the checkout already contains unrelated Task 84 generated/proto layers.
- Native Windows race execution remains blocked by the missing cgo C compiler; WSL race passed and is the local race evidence for this host.
