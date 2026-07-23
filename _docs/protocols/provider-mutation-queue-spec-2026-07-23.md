# Provider Mutation Queue Spec - 2026-07-23

Task 85A defines one durable provider-originated transaction pipeline for all production chain writes emitted by `cmd/provider-daemon`.

## Envelope

Every queue item stores:

- `schema_version`: currently `1`.
- `kind`: stable provider mutation kind.
- `type_url`: registered SDK message type URL.
- `message_bytes` and `message_digest`: replayable protobuf bytes and SHA-256 digest. Map-free messages digest the protobuf bytes directly; `MsgUpdateSupportRequest` uses a deterministic logical digest for its public-metadata map while still storing protobuf bytes for decode/replay.
- `signer` and `idempotency_key`: configured provider account and deterministic logical key.
- `state`: `pending`, `building`, `built`, `broadcasting`, `ambiguous`, `included`, `confirmed`, `retry`, or `dead_letter`.
- `account_number`, `sequence`, `gas_limit`, `tx_bytes`, `tx_digest`, and `tx_hash`.
- bounded attempts with sequence, gas, hash, outcome and error classification.
- confirmation height/block hash, finality height, reconciliation state/count and dead-letter reason.

Unknown kinds or type URL mismatches fail before persistence. Replayed idempotency keys with the same digest return the existing envelope. A different digest for the same logical key returns `ErrProviderMutationConflict` while an earlier envelope is active, but a new envelope is permitted after the prior envelope is terminal (`confirmed` or `dead_letter`) so legitimate mutable resource updates do not conflict forever.

## Registry

The default registry admits only provider-signed production mutations. It validates each concrete SDK message, derives the expected signer and logical idempotency key, and rejects mismatched kind/type pairs.

Customer/owner-signed market lease mutations are outside this provider-originated queue until a customer-side durable submitter exists.

## Submission Flow

1. Caller builds a registered SDK message and calls `ProviderMutationSubmitter.Submit`.
2. Readiness verifies the worker is started, the queue store is open, local lease is held, key manager is unlocked, and chain transport/confirmation are available.
3. The registry encodes canonical message bytes and writes the envelope with put-if-absent semantics.
4. The submitter resolves account number/sequence, builds a direct-sign-mode SDK transaction through `KeyManager`, simulates gas, applies gas adjustment, rebuilds and persists transaction bytes/hash.
5. Broadcast uses Comet RPC. Accepted, lost or uncertain responses become explicit persisted states.
6. Inclusion confirmation requires the expected transaction hash, non-zero height and block hash before recording inclusion, then waits for configured finality blocks.
7. Crash/restart recovery marks `building`, `built`, `broadcasting` and `included` states as `ambiguous` before rebuilding.
8. Ambiguous outcomes reconcile by tx hash first, then by message-specific logical chain state before any replacement transaction.

Every persisted state transition re-checks the local submitter lease and envelope lease token. Loss of lease fails closed with `ErrProviderMutationNotReady` rather than continuing a stale broadcast/finality path.

## Production Fail-Closed Rules

- `ChainUsageSubmitter` requires `ProviderMutationSubmitter` in production. The legacy injected `ChainSubmitterClient` path is available only when tests set `AllowTestLegacyChainClient`.
- Waldur/provisioning callbacks require `NewDurableChainCallbackSink`; missing callback sink is a startup error.
- HPC node metadata submission requires an injected `HPCNodeChainReporter`; raw HTTP/no-op submission is not used.
- Domain verification requires the production chain client implementing the generalized mutation backend; the old standalone RPC confirmation backend is disabled.

## Errors And Metrics

Error classes are bounded: `unavailable`, `sequence_mismatch`, `out_of_gas`, `mempool_reject`, `timeout`, `replacement`, `invalid`, `unauthorized`, and `unknown`.

Metrics expose queue depth, oldest pending age, max attempts, ambiguous items, dead letters, confirmation latency and last success/failure without payload-derived labels.

## Task 85C Boundary

Task 85A supplies the `QueueStore`/`ProviderMutationStore` seam and `SubmitterLease` fencing interface with a local single-process lease. Multi-replica fencing, distributed lease ownership and cross-process sequence coordination remain Task 85C.
