# Fiat Conversion Incident and Recovery Runbook

**Applies to:** Task 85B DEX/off-ramp engineering and any later certified deployment
**Current production status:** `engineering_complete_external_blocked`

Only deterministic local fixture conformance is complete. No real Osmosis testnet or payout-provider sandbox operation is represented by this runbook's current evidence; real testnet/sandbox and production exercises remain external gates.

## Immediate controls

1. Stop new conversion intents by pausing the affected profile and/or disabling fiat conversion through governed change control.
2. Preserve the provider mutation queue, fiat conversion state file, fiat repository, chain heights/hashes, profile files/digests and relevant provider/custody references. Do not edit state files or pending observation bytes.
3. Ensure exactly one worker holds the provider lease. If ownership is uncertain, stop all workers and restart one after lease expiry and evidence capture.
4. Keep settlement value held. Never mark a conversion, payout or queue item complete from logs, timeout assumptions or a webhook alone.
5. Record conversion, settlement, payout, quote, transaction, profile and observation IDs. Do not copy beneficiary, bank, credential, KYC or sanctions data into tickets.

## Quote expiry

### Before swap submission

- Confirm the on-chain stage has not advanced to `SWAP_SUBMITTED`.
- Confirm block time has reached the recorded quote expiry. Before expiry, a replacement is invalid.
- Let the orchestrator clear the expired local DEX quote and obtain a new quote under the same committed profiles and intent. It must append a contiguous replacement `QUOTE_ACCEPTED` observation; it must not reset or rewrite the existing lineage.
- Verify new pool height/block, oracle comparison, minimum output and expiry.
- Do not reuse an expired quote or its signed payload.

Before target-chain signing, target-chain broadcast, or payout initiation, the worker must freshly query governed settlement parameters and canonical financial-case holds. A disabled conversion flag, rotated/non-certified current profile, revoked/current compliance failure, active hold, malformed response, or unavailable query means no new external side effect; retain held value and retry or enter manual review.

### After swap or payout submission

- Do not re-quote or initiate a replacement side effect.
- Treat the outcome as ambiguous and follow the relevant reconciliation procedure below.
- An expired provider quote after confirmed initiation does not prove payout failure.

### After swap finality but before payout submission

- If the committed payout quote expires while the conversion is `PAYOUT_PENDING`, confirm no provider payout ID, provider reference, or privacy-safe reference hash has been committed.
- Obtain a new quote under the same conversion sequence, payout profile, compliance decision, settlement lineage and stable amount. The replacement quote ID and digest must both differ from the expired quote.
- Submit the replacement as the next contiguous `PAYOUT_QUOTED` observation. The daily conversion reservation and linked pending payout remain unchanged; do not release or create a second reservation.
- A pre-expiry replacement, same-ID/digest replacement, or replacement after provider submission is invalid and must enter reconciliation/manual review instead of retrying initiation.

## Stale or reorged DEX evidence

1. Pause the route.
2. Compare the accepted quote's chain ID, observation height/block hash and age/lag limits with independent target-chain RPC/archive evidence.
3. For a submitted swap, query transaction inclusion, block hash, confirmations, token-out amount and canonical finality hash.
4. If the block was reorged, finality hash mismatches, output is below minimum, or evidence cannot be independently obtained, move the conversion to manual review and keep value held.
5. Resume only after the route profile/oracle/RPC evidence is current and the incident owner signs the reconciliation record. Never rewrite a recorded observation lineage.

## Ambiguous swap

1. Identify the stored quote digest, payload hash, signed-transaction hash, target-chain hash if known, and deterministic swap correlation ID.
2. Reconcile by transaction hash and, where supported, custody/client correlation. Require chain ID, non-zero height, block hash, confirmations and output.
3. If no hash was returned, ask the custody backend to recover only the exact `TxRaw` matching the stored signed hash and canonical payload. Verify it before any rebroadcast.
4. If the exact transaction is final, submit/reconcile the missing `SWAP_SUBMITTED` or `SWAP_FINALIZED` observation through the durable mutation queue.
5. If non-execution cannot be proven, do not build a new transaction. Escalate to manual review.

## Ambiguous payout

1. Retain the daily-limit reservation and existing idempotency key.
2. Search the provider by immutable metadata/correlation before calling initiation again.
3. Reconcile provider, quote ID, payout ID, amounts, fee, beneficiary token reference, status timestamps and provider reference.
4. Persist the immutable webhook binding as soon as a payout ID is recovered.
5. A timeout, retryable HTTP error or conflict is not a failure. Do not release quota or initiate with a new key until the provider proves no payout exists.
6. Completion still requires provider status and an authenticated matching webhook; otherwise remain pending/manual review.

On process restart, a durable bridge record may exist while a fresh HTTP adapter has no volatile payout binding. If status polling returns not-found, the bridge performs one explicit binding restore by immutable metadata/correlation. The provider response must exactly reproduce provider, payout ID, quote ID, correlation/metadata, fiat and crypto amounts, fee, reference, initiation time, status monotonicity and the durable daily-reservation identity. No match, multiple/ambiguous match, or any mismatched field is rejected. Only after exact validation may the adapter restore its status/webhook binding and poll again; restoration never authorizes a second initiation.

## Governance, compliance or hold change after an irreversible boundary

1. Determine whether target-chain signing/broadcast, swap submission, a provider payout ID or custody movement has crossed an irreversible boundary.
2. Before the boundary, current governance/profile/compliance/financial-case authorization controls and may stop the next side effect.
3. After the boundary, do not abandon reconciliation because the current enable flag, profile, compliance status or hold changed. Reconcile submission/finality using the immutable profile and compliance commitments accepted with the intent.
4. Do not use this rule to initiate another swap or payout. It permits observation and reconciliation of an existing side effect only.
5. A financial case opened after the boundary is an incident/evidence hold. It must not allocate the linked payout exposure, cancel the conversion, transfer provider/customer/platform amounts or release a second payout. Escalate to the governed external-reconciliation path.

## Custody and treasury accounting verification

1. On authenticated payout completion, expect one deterministic native transfer from the settlement module account to the internal-only fiat-custody module account for the conversion's net crypto amount. Do not expect a synthetic provider-account or external bank/chain transfer from consensus.
2. Verify one treasury record per non-zero platform fee, validator fee and holdback, keyed by payout ID and type. An exact retry must not change treasury balance; any conflicting amount/lineage is an incident.
3. Reconcile the custody module-account balance to all completed fiat conversion custody effects and verify each effect hash and linked payout finality hash.
4. During genesis export/import or recovery, preserve treasury records, aggregate treasury balance and declared custody balance together. A mismatch must stop import rather than be repaired by manual store edits.

## Webhook replay or conflict

- **Exact duplicate:** the durable replay store returns `duplicate_exact`; persist/acknowledge idempotently and do not apply a second state change.
- **Same event ID, different bytes:** reject as `duplicate_conflicting`, pause the corridor, retain raw-body digest and key ID/version, and open a security/provider incident.
- **Binding mismatch:** reject callbacks whose provider, payout, quote or correlation differs from the durable binding.
- **Status conflict:** if verified webhook events disagree, or webhook status differs from provider status, do not complete. Reconcile directly with the provider and move to manual review if disagreement persists.

Never bypass signature, timestamp, key-version, replay-store or binding checks to clear a backlog.

## Profile rotation, pause or mismatch

1. Pause new work before rotating route, provider, webhook or authority keys.
2. Record old and new canonical profile digests, authority signature/key version, chain governance height and activation time.
3. Existing work remains bound to its original profile digest. Do not replace its profile ID/digest in local state.
4. If local files, on-chain parameters or an intent disagree, startup/processing must remain blocked. Resolve by reviewed profile publication and governance, not by editing the queue.
5. Keep both webhook key versions only for an explicitly bounded overlap; remove the old version after all old-profile payouts reconcile.
6. A `paused`, `unsupported` or `engineering_complete_external_blocked` row is non-executable in production.

## Lease loss

1. Treat lease loss as an immediate stop signal; do not continue signing, broadcasting, payout initiation or local state updates.
2. Determine the current fencing token and whether another process owns the path lock/lease.
3. Reconcile every item left in signing, swap-broadcast, payout-quote or payout-submitted state before restarting one worker.
4. Do not run multiple active provider workers. Task 85B provides local fencing only; distributed failover is not supported until Task 85C evidence exists.

## Provider mutation backlog

1. Check mutation submitter readiness, queue depth/age, pending observation, account sequence, tx hash/inclusion and dead-letter classification.
2. Restore key manager, chain query, broadcast and confirmation transports before processing due work.
3. If an observation transaction response was lost, compare on-chain observation sequence and digest. Exact commitment finalizes local pending state; divergence requires manual review.
4. Do not delete/recreate an observation or alter its idempotency key. Follow `_docs/runbooks/provider-mutation-queue-recovery-2026-07-23.md` for queue-level recovery.
5. Alert on increasing oldest intent age, ambiguous swap/payout counts, manual review, dead letters or finality latency.

## Manual review and independent reconciliation

Build a privacy-safe evidence bundle containing:

- conversion/request digest, profile IDs/digests and governance state;
- observation sequences/digests and mutation queue transaction evidence;
- DEX quote/pool/oracle commitments, custody authorization reference, transaction inclusion/finality and balance deltas;
- payout quote/idempotency/correlation, provider status/settlement report, webhook digest/key version and destination receipt reference;
- settlement/payout lineage, held value, fees and daily-limit reservation; and
- incident decisions, reviewers and timestamps.

Two independent reviewers must reconcile DEX input/output/fees, the native settlement-to-custody module movement, exactly-once platform/validator/holdback treasury records, target-chain custody balances, stable amount, provider debit/fee, fiat destination receipt and on-chain terminal record. Any unexplained delta, missing finality, changed compliance decision or duplicate native/external movement keeps the item in manual review.

Recovery must use normal authenticated observations or a separately governed remediation path. Never mutate terminal records, replay indexes, external finality hashes or accounting files by hand. Production may resume only after root cause, balance conservation, profile state and external approvals are all revalidated.
