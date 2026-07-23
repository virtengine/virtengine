# Provider Mutation Queue Recovery Runbook - 2026-07-23

Use this runbook when the provider daemon reports failed readiness, queued provider mutations, ambiguous transactions, or dead letters.

## Readiness

1. Check daemon startup output for `Provider Mutation Submitter: started (durable signed pipeline)`.
2. If readiness fails, inspect the reason:
   - `submitter not started`: daemon composition failed before starting the submitter.
   - `queue store unavailable`: queue file path is inaccessible or malformed.
   - `submitter lease not held`: another local process owns the signer lease or the worker lost renewal.
   - `key manager unavailable or locked`: provider key initialization/unlock failed.
   - `chain transport unavailable`: gRPC/Comet RPC transport is absent.
   - `confirmation transport unavailable`: Comet RPC confirmation path is absent.

## Queue File

Default queue path is `.cache/provider_daemon/provider_mutation_queue.json`; `cmd/provider-daemon` composes it from `--chain-usage-queue-file` with `.mutations`.

Do not edit message bytes, digests, transaction bytes or idempotency keys. If inspection is required, copy the file first and decode only the copy.

## Ambiguous Items

An item is ambiguous after response loss, timeout, restart during build/broadcast/inclusion, sequence uncertainty, or reorg detection.

Recovery steps:

1. Restore RPC connectivity and the provider key.
2. Start exactly one daemon process for the provider account.
3. Let `ProcessDue` reconcile by tx hash and logical state.
4. If the chain committed the transaction with matching tx hash, non-zero height and block hash, the item moves through `included` to `confirmed`.
5. If not committed and account sequence is still safe, the submitter clears tx bytes/hash and rebuilds.
6. If logical state conflicts, the item becomes `dead_letter` with `replacement`.

If reconciliation can prove only logical state but cannot provide tx hash, height and block hash, the item remains ambiguous. Do not manually mark it confirmed; restore chain query/indexer evidence or resubmit only after confirming the original transaction did not commit.

## Dead Letters

Dead letters are terminal local outcomes. Common reasons:

- `registry_validation_failed`: persisted message bytes no longer match a registered type or digest.
- `broadcast_rejected`: chain rejected the signed transaction as invalid or unauthorized.
- `attempts_exhausted`: retries reached `MaxAttempts`.
- `logical_idempotency_conflict`: chain state already contains a different logical result.

Operational response:

1. Record the envelope ID, kind, signer, classification, attempts and dead-letter reason.
2. Verify the provider account, key epoch and chain ID.
3. For invalid/unauthorized classes, fix the caller or account state before resubmitting a new logical mutation.
4. For conflicts, reconcile business state first; do not replay the same idempotency key with different payload bytes.

Terminal envelopes do not permanently reserve their logical idempotency key. A later mutable update to the same resource creates a new envelope after the earlier one is `confirmed` or `dead_letter`; concurrent different payloads for the same logical key still conflict.

## Startup Fail-Closed Checks

- Usage submission must have a generalized `ProviderMutationSubmitter`; the old `ChainSubmitterClient` path is test-only.
- Waldur bridge/provisioning callbacks must have a durable chain callback sink.
- HPC node metadata submission must have an injected chain reporter.
- Domain verification must have the production generalized chain backend.
- If any of these are missing, fix daemon composition or flags instead of substituting file sinks or raw RPC broadcast calls.

## Release-Only Evidence

Task 85A local engineering gates prove signed SDK transactions, durable restart, timeout/reconcile, lease-loss fail-closed behavior, metadata-map digest determinism and failure classification. Release certification must still run against a long-lived localnet/testnet with real committed events for every externally enabled provider profile and a Task 85C multi-replica fencing proof.
