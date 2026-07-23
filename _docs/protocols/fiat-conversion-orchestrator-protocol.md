# Fiat Conversion Orchestrator Protocol

**Protocol version:** 1
**Task 85B status:** `engineering_complete_external_blocked`

This status records deterministic local engineering/fixture conformance only. No real Osmosis testnet swap or payout-provider sandbox execution is claimed. Those real external conformance runs, followed by separate production certification evidence, remain required by the Task 85B plan.

## Boundary

`x/settlement` is the deterministic authority for intent, value holds, limits, compliance/profile commitments, observation lineage and terminal accounting. It performs no DEX, custody, payout-provider, webhook or secret-manager I/O. The provider daemon owns those external side effects and reports only authenticated, privacy-safe observations.

A conversion starts only when governance has enabled fiat conversion and committed exact DEX/payout profile IDs and SHA-256 digests. The request is provider-signed, linked to one pending payout and settlement, blocked by active canonical financial cases, bounded by per-conversion/daily limits, and bound to an encrypted destination envelope plus current compliance digest.

## Six authenticated observation stages

Every happy-path observation is a generated `MsgRecordFiatConversionObservation` submitted through the Task 85A durable signed mutation queue. SDK transaction authentication binds `sender` to the conversion provider. The keeper also requires the next sequence, a 32-byte idempotency key, immutable accepted profile IDs/digests and compliance digest, bounded block-time-relative observation time and stage-specific evidence. A stage that authorizes a new external side effect additionally requires current governance enablement, current certified profile commitments, current compliance and no active financial-case hold. Submission/finality stages reconcile an already-crossed irreversible boundary against immutable accepted commitments even if current governance, profile, compliance or hold state later changes.

1. **QUOTE_ACCEPTED** — commits the canonical DEX quote digest, expiry and positive minimum stable output; transitions `created` to `swap_pending`. After expiry and only before swap submission, a distinct replacement quote can append another `QUOTE_ACCEPTED` without releasing the payout hold or daily quota.
2. **SWAP_SUBMITTED** — binds the target-chain transaction hash to the accepted, unexpired quote and minimum output; transitions to `swap_submitted`.
3. **SWAP_FINALIZED** — commits transaction hash, height, block hash, confirmation count, finality hash and stable amount at or above minimum output; transitions to `swap_settled`.
4. **PAYOUT_QUOTED** — requires the linked pending payout and commits off-ramp quote ID, digest and expiry; transitions through off-ramp pending to `payout_pending`. After expiry and only before provider initiation, a distinct replacement quote can append another `PAYOUT_QUOTED` while retaining all immutable conversion commitments and reservations.
5. **PAYOUT_SUBMITTED** — binds provider payout ID and privacy-safe reference hash to the unexpired payout quote; transitions to `payout_submitted`.
6. **PAYOUT_COMPLETED** — requires provider completion plus authenticated webhook agreement, positive fiat amount, finality hash and reference hash; atomically completes the linked payout, moves the native net amount from the settlement module account to the internal-only fiat-custody module account, and records exactly-once platform-fee, validator-fee and holdback treasury entries. It performs no synthetic provider-account or external bank/chain transfer.

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
- After restart, a durable nonterminal payout can restore a fresh adapter's volatile status/webhook binding only if one metadata/correlation lookup exactly matches provider, payout, quote, reference, fiat/crypto amounts, fee, initiation time, status and durable daily-reservation identity. No match, ambiguity or mismatch fails closed.
- Attempt exhaustion, profile/intent divergence, reorg, finality mismatch, webhook conflict or payout binding mismatch moves work to failed/manual review rather than guessing success.

Task 85B uses a local single-process lease with renewal and a monotonically fenced token. Loss of the lease stops side effects and updates. Distributed multi-replica leasing/fencing is Task 85C work; multiple active workers for one provider account are not supported by this protocol version.

## Custody and signed execution

The provider `KeyManager` signs an authorization over the exact canonical quote execution payload. A distinct injected target-chain custody backend supplies the target-chain address and constructs/verifies actual Cosmos `TxRaw` bytes. Provider-chain keys are never reinterpreted as DEX-chain keys. Osmosis reserve payloads must come from an injected authenticated pool provider bound to response height, block hash and source node identity and agree with independent chain evidence.

The DEX adapter verifies that the signed transaction is bound to the canonical payload before broadcast. The production constructor rejects test-only custody, absent custody, detached payloads, malformed `TxRaw`, missing signatures, missing hashes and unsafe transport.

## Finality

DEX quotes bind exact pool reserves, fee, chain ID, height, block hash and oracle comparison. Execution completion requires a confirmed transaction, profile-defined confirmation floor, block identity, actual token-out event and minimum output. The orchestrator independently recomputes a canonical finality hash from chain ID, transaction hash, height, block hash, confirmations and output.

Payout completion requires all of the following: provider status `completed`, matching provider/quote/correlation/economic fields, completion timestamp, matching verified webhook status, privacy-safe reference hash and payout finality hash. Consensus records external payout finalization and performs the deterministic native module-to-module custody-sink movement described above, but performs no synthetic provider-account or external bank/chain transfer.

Platform fee, validator fee and holdback accounting are keyed by payout ID plus treasury-entry type. An exact retry observes the same record and does not add balance again; a conflicting amount or lineage fails the cached transaction. Genesis import/export preserves and validates those records, their aggregate balance and the fiat-custody module-account balance.

## Financial-case and irreversible-boundary policy

Before target-chain signing, broadcast or payout initiation, current canonical financial-case holds stop the new external side effect. At swap submission, the fiat protocol may release only its canonical payout hold and only when another local canonical incident hold remains. Once swap submission, a target transaction, provider payout ID or custody movement establishes the irreversible boundary, a later financial case is evidence/incident control only: it cannot cancel the conversion, allocate the linked payout exposure, transfer provider/customer/platform value or otherwise double-allocate funds. Finalization remains blocked pending governed external reconciliation.

## Webhook and profile trust

Production webhooks require exact profile/API/key versions, HMAC-SHA256 over timestamp plus raw body, bounded clock skew/body size, immutable payout binding, durable replay storage and durable verified-event storage. Exact duplicate payloads are harmless; the same event ID with different bytes is rejected.

Profile files use a strict schema and canonical JSON SHA-256 digest. A certified production row must carry an Ed25519 signature verified against a trust root configured independently from the file. Runtime authorizers accept only the exact loaded bytes. Self-declared evidence text, a matching profile ID, or backend identifier cannot authorize production.

## Privacy and fail-closed behavior

Destination and compliance resolvers expose clear references only for the duration of one provider call; the worker clears local copies and persists only digests. Operational metadata rejects bank, beneficiary, account, routing, credential and secret fields. On-chain observations contain bounded IDs, amounts, statuses and hashes, never raw beneficiary or compliance evidence.

Production construction requires certified independently trusted profiles, non-test custody, durable stores and all external dependencies. The current command has no reviewed in-binary factories for those backends and intentionally fails startup. Engineering external-blocked mode validates profile structure and exits without executing. This is the required behavior until the external certification ledger is complete.
