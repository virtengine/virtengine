# Fiat Conversion Orchestrator Protocol

**Protocol version:** 1
**Task 85B status:** `engineering_complete_external_blocked`

## Boundary

`x/settlement` is the deterministic authority for intent, value holds, limits, compliance/profile commitments, observation lineage and terminal accounting. It performs no DEX, custody, payout-provider, webhook or secret-manager I/O. The provider daemon owns those external side effects and reports only authenticated, privacy-safe observations.

A conversion starts only when governance has enabled fiat conversion and committed exact DEX/payout profile IDs and SHA-256 digests. The request is provider-signed, linked to one pending payout and settlement, blocked by active canonical financial cases, bounded by per-conversion/daily limits, and bound to an encrypted destination envelope plus current compliance digest.

## Six authenticated observation stages

Every happy-path observation is a generated `MsgRecordFiatConversionObservation` submitted through the Task 85A durable signed mutation queue. SDK transaction authentication binds `sender` to the conversion provider. The keeper also requires the next sequence, a 32-byte idempotency key, exact current profile IDs/digests, current compliance digest, bounded block-time-relative observation time and stage-specific evidence.

1. **QUOTE_ACCEPTED** — commits the canonical DEX quote digest, expiry and positive minimum stable output; transitions `created` to `swap_pending`.
2. **SWAP_SUBMITTED** — binds the target-chain transaction hash to the accepted, unexpired quote and minimum output; transitions to `swap_submitted`.
3. **SWAP_FINALIZED** — commits transaction hash, height, block hash, confirmation count, finality hash and stable amount at or above minimum output; transitions to `swap_settled`.
4. **PAYOUT_QUOTED** — requires the linked pending payout and commits off-ramp quote ID, digest and expiry; transitions through off-ramp pending to `payout_pending`.
5. **PAYOUT_SUBMITTED** — binds provider payout ID and privacy-safe reference hash to the unexpired payout quote; transitions to `payout_submitted`.
6. **PAYOUT_COMPLETED** — requires provider completion plus authenticated webhook agreement, positive fiat amount, finality hash and reference hash; completes the linked payout without an on-chain coin transfer.

`FAILED` and `CANCELLED` are additional bounded terminal controls. Cancellation is allowed only before swap submission. Failure/manual-review policy keeps value held rather than inferring a refund or completion.

Exact duplicate idempotency keys or sequences with the same message digest are accepted as duplicates. Any conflicting reuse is rejected. Each observation appends a digest and chained lineage digest, and all replay/sequence indexes are covered by invariants.

## Off-chain durable state

The orchestrator snapshots immutable, privacy-safe intent fields into a schema-versioned `FileFiatConversionStore`. Atomic replacement, exclusive path locking, monotonic transitions, bounded attempts and a fencing token protect each work item. It persists quote/finality commitments, external IDs, hashes and pending observations, but not credentials, beneficiary tokens, identity evidence, raw execution payloads or signed transactions.

A separate `FileFiatRepository` durably stores privacy-safe payout records, daily-limit reservations, webhook replay digests, immutable bindings and verified webhook events. A verified callback is persisted before HTTP acknowledgement and is consumed only after the matching completion observation is final on-chain.

Each observation is first persisted as pending, then submitted through `ProviderMutationSubmitter`. Local work advances only after the durable mutation is final or the on-chain sequence/digest proves an exact prior commit. Response loss cannot create a second logical observation.

## Idempotency and restart

- Claiming the same unchanged on-chain intent returns the existing work item; changed intent bytes force manual review.
- Swap and payout operations use stable conversion-derived correlation and idempotency keys.
- Restart converts signing/broadcast work to ambiguous swap reconciliation and payout-quote/submitted work to ambiguous payout reconciliation.
- A swap is never blindly rebuilt after an uncertain broadcast. The worker reconciles target-chain evidence; if needed, the custody backend must recover the exact signed `TxRaw` matching the stored hash before rebroadcast.
- An uncertain payout is found by immutable metadata before any new initiation attempt. Daily-limit reservation remains held while the result is ambiguous.
- Attempt exhaustion, profile/intent divergence, reorg, finality mismatch, webhook conflict or payout binding mismatch moves work to failed/manual review rather than guessing success.

Task 85B uses a local single-process lease with renewal and a monotonically fenced token. Loss of the lease stops side effects and updates. Distributed multi-replica leasing/fencing is Task 85C work; multiple active workers for one provider account are not supported by this protocol version.

## Custody and signed execution

The provider `KeyManager` signs an authorization over the exact canonical quote execution payload. A distinct injected target-chain custody backend supplies the target-chain address and constructs/verifies actual Cosmos `TxRaw` bytes. Provider-chain keys are never reinterpreted as DEX-chain keys.

The DEX adapter verifies that the signed transaction is bound to the canonical payload before broadcast. The production constructor rejects test-only custody, absent custody, detached payloads, malformed `TxRaw`, missing signatures, missing hashes and unsafe transport.

## Finality

DEX quotes bind exact pool reserves, fee, chain ID, height, block hash and oracle comparison. Execution completion requires a confirmed transaction, profile-defined confirmation floor, block identity, actual token-out event and minimum output. The orchestrator independently recomputes a canonical finality hash from chain ID, transaction hash, height, block hash, confirmations and output.

Payout completion requires all of the following: provider status `completed`, matching provider/quote/correlation/economic fields, completion timestamp, matching verified webhook status, privacy-safe reference hash and payout finality hash. Consensus records external payout finalization but performs no synthetic bank or chain transfer.

## Webhook and profile trust

Production webhooks require exact profile/API/key versions, HMAC-SHA256 over timestamp plus raw body, bounded clock skew/body size, immutable payout binding, durable replay storage and durable verified-event storage. Exact duplicate payloads are harmless; the same event ID with different bytes is rejected.

Profile files use a strict schema and canonical JSON SHA-256 digest. A certified production row must carry an Ed25519 signature verified against a trust root configured independently from the file. Runtime authorizers accept only the exact loaded bytes. Self-declared evidence text, a matching profile ID, or backend identifier cannot authorize production.

## Privacy and fail-closed behavior

Destination and compliance resolvers expose clear references only for the duration of one provider call; the worker clears local copies and persists only digests. Operational metadata rejects bank, beneficiary, account, routing, credential and secret fields. On-chain observations contain bounded IDs, amounts, statuses and hashes, never raw beneficiary or compliance evidence.

Production construction requires certified independently trusted profiles, non-test custody, durable stores and all external dependencies. The current command has no reviewed in-binary factories for those backends and intentionally fails startup. Engineering external-blocked mode validates profile structure and exits without executing. This is the required behavior until the external certification ledger is complete.
