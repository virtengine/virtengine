# Reservation Quarantine Remediation

## Purpose

Quarantine retains capacity when a dispute, provider outage, or ambiguous migration prevents safe release. Operators must not edit stores directly or invent capacity.

## Diagnosis

1. Query reservation lineage by reservation ID, order, lease, job, consumer, or provider.
2. Confirm the provider inventory ID and heartbeat sequence.
3. Compare order/bid/lease, HPC job, escrow, settlement, and deployment terminal states.
4. Inspect `legacy_source`, `legacy_reference`, `reason`, and transition history.
5. Check for another active reservation with the same consumer.

## Resolution classes

- **Valid active workload:** wait for the canonical dispute/settlement adapter; do not release.
- **Proven terminal workload:** submit the authorized release adapter once.
- **Provider non-fulfillment:** submit the authorized slash adapter; it records evidence and restores capacity once.
- **Ambiguous migration:** reconcile against a signed provider inventory snapshot and canonical market/HPC lineage. If evidence does not uniquely identify capacity, leave quarantined.
- **Zero-capacity legacy quarantine:** do not interpret zero as available or reserved capacity. It is lineage-only evidence created because migration could not prove an inventory claim; remediation must atomically establish a governed inventory baseline and canonical consumer link before any capacity becomes sellable.
- **Duplicate legacy link:** preserve both historical records, select neither by guess, and use governance remediation after financial reconciliation.

## Safety checks

After remediation, run the resources conservation invariant and app lineage invariant. Available capacity must never exceed declared total. A terminal reservation must remain terminal on retry.

## Escalation

Task 84D supplies the canonical dispute resolution authority. Task 90B/87A will certify benchmark and attestation profiles. Until then, those mandatory profiles remain fail closed.

## Rollback

Do not roll back a chain below `v1.6.0` after the upgrade block. Older binaries permit the non-owner lifecycle writes and ignore reservation indexes, so replaying them can create two mutable owners or oversell inventory. The supported recovery is restore the pre-upgrade consensus snapshot on all validators, correct the upgrade binary/plan, and re-run the upgrade deterministically. Exported post-upgrade genesis must retain canonical activation flags and quarantine lineage.