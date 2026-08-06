# Financial Case Duplicate and Quarantine Remediation

## Purpose

Recover ambiguous `v1.7.0` financial subjects without guessing allocation or releasing held value. Settlement is authoritative; fraud, HPC, review and escrow billing records are evidence projections only.

## Safety rules

1. Never use legacy payout-release, billing-resolution, HPC-resolution or fraud-refund operations after activation.
2. Never delete a hold, case, claim, transition or source record manually.
3. Never infer a split from one source module. Reconcile all order/invoice/usage/job/escrow/reservation aliases and balances.
4. Keep the case `QUARANTINED` until a governance proposal names one conserved allocation.
5. Do not put evidence, narratives, PII, manifests, biometrics, keys or decrypted payloads in proposals/logs. Use hashes and encrypted references.

## Detection

Alert on:

- `financial_case_quarantined` event;
- canonical financial-case invariant failure;
- active subject aliases mapping to multiple cases;
- held payout/reservation with missing or terminal case;
- active value-bearing case with zero/mismatched hold count;
- incomplete terminal effect marker;
- migration report `quarantined` or `malformed_orphans` above zero.

## Triage

1. Pause operator automation for the affected order/invoice/job only.
2. Query the canonical case by case ID and each known alias.
3. Record privacy-safe IDs and hashes: case, claims, order, invoice, settlement, payout, escrow, usage, job, reservation and source report IDs.
4. Verify payout is `HELD`, escrow is `DISPUTED`, and reservation is `DISPUTED` where references exist.
5. Compare module-account balances with `original_held` per denomination.
6. Verify every source record is nonterminal or historically terminal. A completed historical transfer is immutable and must not be reversed by migration.
7. Compute the proposed allocation per denomination and prove:

$$provider + customer + platform + slashWitness = originalHeld$$

## Duplicate active roots

When two active roots share any alias:

1. quarantine both roots and preserve all holds;
2. choose the deterministic root whose canonical subject group matches the deployed order/invoice lineage; do not choose by creation time or local sequence;
3. submit a governance remediation that adds every source record as a typed claim on the selected root;
4. mark the other root as a quarantined migration projection with no allocation/effects;
5. rebuild subject/party/status/escrow indexes through the governed migration handler;
6. run invariants before allowing resolution.

No duplicate root may be deleted, finalized or released as part of deduplication.

## Orphan hold

For a payout/reservation held by an unknown case ID:

1. quarantine the record and the matching financial lineage;
2. locate legacy dispute/source records by invoice/order/job and their hashes;
3. create one migrated canonical root with migration claims;
4. rebind the hold only in one cached governance migration;
5. retain the old ID in `source_reference` and transition audit.

If parties or original exposure cannot be proven, leave the root quarantined. Do not fabricate addresses or allocation.

## Malformed or missing evidence reference

A malformed evidence record is never copied into canonical state. Add a migration claim containing only:

- source module and source record ID;
- SHA-256 hash of the legacy record/reference;
- `migration://...` encrypted/reference locator;
- `MIGRATION` claim type.

Quarantine when integrity or party binding cannot be established.

## Resolution and verification

1. Governance submits a valid multi-denom terminal allocation.
2. Wait through the committed appeal window; do not manually release holds.
3. Finalize once. If it fails, inspect the stable effect error and retry; do not run a legacy transfer.
4. Confirm all effects are `APPLIED`, hold count is zero, payout/escrow/reservation state is consistent and source projections show the canonical final status.
5. Run settlement financial-case, resources capacity and app lineage invariants.
6. Archive the migration/reconciliation digest, case ID, allocation hash, proposal ID, heights and command outputs. Do not archive sensitive evidence.

## Rollback boundary

Before activation height, operators may roll back the binary according to the normal software-upgrade procedure. After `v1.7.0` commits cases or new reservation states, do not run an older binary: it cannot enforce canonical holds or decode all persisted state. Recovery is forward-only through an audited patch/upgrade.
