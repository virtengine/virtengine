# VEID Registry and Issuance Policy Protocol

**Version:** 1.0.0  
**Date:** 2026-04-10  
**Status:** Implemented  

## Overview

`x/veidregistry` and `x/issuancepolicy` are the governed control plane for the VEID verification pipeline.

- `x/veidregistry` owns verifier artifact identity, approval state, validator readiness, and deterministic activation.
- `x/issuancepolicy` owns proof-to-issuance eligibility, mint-rate limits, pause state, and proof mint records.
- `x/veid` consumes both modules at verification time and fails closed when governed verifier artifacts are stale or unauthorized.

This document describes the implemented behavior in:

- [`x/veidregistry`](../../x/veidregistry)
- [`x/issuancepolicy`](../../x/issuancepolicy)
- [`x/veid`](../../x/veid)

## `x/veidregistry`

### Consensus state

Registry state is keyed by verifier identifier and contains:

- immutable verifier metadata: `verifier_id`, `spec_version`, `weights_sha256`, optional spec/test-vector/build hashes, optional pipeline image hash, optional model manifest hash
- scheduled activation height
- governance proposal id
- lifecycle status: `proposed`, `approved`, `active`, `deprecated`, `retired`, `cancelled`
- validator readiness reports: `validator_address`, `implementation_id`, `organization`, `conformance_passed`, `reported_height`
- a single active verifier pointer: `verifier_id`, `spec_version`, `activated_at_height`
- registry params: `minimum_ready_validators`, `minimum_independent_implementations`, `allow_legacy_mirroring`

### Governance update semantics

All mutating governance operations are authority-gated and execute through the module Msg server.

- `MsgUpsertVerifierVersion` creates or updates a verifier in `proposed` state only.
- `MsgApproveVerifierVersion` moves a proposed verifier into `approved` state, records the governance proposal id, and sets the scheduled activation height.
- `MsgCancelVerifierVersion` moves a `proposed` or `approved` verifier into `cancelled`.
- `MsgRetireVerifierVersion` moves a `deprecated` verifier into `retired`.
- `MsgUpdateParams` updates registry params.
- `MsgReportValidatorReadiness` is validator-operated and records conformance readiness for a specific verifier.

Direct governance activation is not part of the production flow. Activation is automatic once the verifier is both due and eligible.

### Status transitions

The implemented lifecycle is:

```text
proposed -> approved -> active -> deprecated -> retired
proposed -> cancelled
approved -> cancelled
```

The module rejects invalid transitions, including:

- activating a verifier that is not `approved`
- mutating a `retired` or `cancelled` verifier back into an earlier state
- retiring an `active` verifier directly without first deprecating it through activation of a successor

### Deterministic activation

`BeginBlock` calls `ActivateReadyVerifiers`.

A verifier becomes eligible only when all of the following are true:

- status is `approved`
- `activation_height <= current block height`
- `ready validator count >= minimum_ready_validators`
- `independent implementation count >= minimum_independent_implementations`

When more than one verifier is eligible in the same block, the module chooses a single canonical candidate by:

1. lowest `activation_height`
2. lexical `verifier_id` tie-break

That ordering is now covered by keeper tests.

When a verifier activates:

- the new verifier becomes `active`
- the previous active verifier becomes `deprecated`
- the active verifier pointer is updated atomically
- an activation event is emitted

When an approved verifier is due but still short on readiness, the module emits readiness-shortfall events instead of activating partially governed state.

### Queries

The implemented query surface exposes:

- `Verifier`
- `Verifiers`
- `QueuedVerifiers`
- `EligibleVerifiers`
- `ActiveVerifier`
- `ValidatorReadiness`
- `Params`

### Legacy mirror semantics

`allow_legacy_mirroring` remains for compatibility with legacy VEID pipeline version storage. When enabled, VEID pipeline registration and activation can mirror into registry state. When disabled, only registry-governed state is authoritative.

Mainnet production should treat governed registry state as canonical and keep compatibility mirroring disabled unless an explicit migration step requires it.

## `x/issuancepolicy`

### Consensus state

Issuance policy state contains:

- policy metadata: `policy_id`, `status`, `active_verifier_scope`, `mint_units_per_proof`, `daily_cap`, `epoch_cap`, `minting_paused`, `created_at_height`, optional `governance_proposal_id`
- a single active policy id
- issuance counters: `day_index`, `minted_today`, `epoch_index`, `minted_this_epoch`
- proof mint records keyed by `proof_id`
- issuance params: `epoch_length_blocks`, `max_mint_units_per_proof`, `max_daily_cap`, `max_epoch_cap`, `emergency_pause_enabled`

### Governance update semantics

All policy administration is authority-gated and handled through the Msg server.

- `MsgUpsertPolicy` creates or updates a governed policy.
- `MsgSetActivePolicy` switches the active policy and deprecates the previous active policy.
- `MsgPausePolicy` pauses a specific policy or the active policy when no policy id is supplied.
- `MsgResumePolicy` resumes a specific policy or the active policy when no policy id is supplied.
- `MsgDeprecatePolicy` permanently deprecates a policy and clears the active pointer when that policy was active.
- `MsgUpdateParams` updates issuance params.

### Policy semantics

The implemented module uses deterministic fixed-unit issuance per successful proof.

Before recording issuance, the keeper enforces:

- proof idempotency: an existing `proof_id` returns the previously recorded result
- active policy presence
- pause state
- verifier-scope match
- daily cap
- epoch cap

If issuance is rejected, the module still records the proof outcome with an explicit status instead of silently dropping the event.

Implemented proof statuses are:

- `recorded`
- `paused`
- `cap_exceeded`
- `verifier_scope_mismatch`
- `duplicate`
- `no_active_policy`

### Queries

The implemented query surface exposes:

- `Policy`
- `Policies`
- `ProofMintRecords`
- `ActivePolicy`
- `Counters`
- `ProofMintRecord`
- `Params`

### Genesis behavior

Issuance genesis now round-trips complete proof state:

- policies
- active policy id
- counters
- proof mint records

Proof mint records were previously omitted from export; that gap is now closed and covered by tests.

## `x/veid` integration

`x/veid` consumes registry and issuance policy state through keeper interfaces wired at app startup.

### Verifier enforcement

During successful verification application, `x/veid` now:

1. resolves the active verifier from `x/veidregistry`
2. verifies that the request block is at or after the governed activation height
3. verifies that the scored model version matches the governed verifier spec version
4. verifies that the active VEID pipeline version is active and usable
5. verifies that the active registry artifact hash matches the active VEID pipeline manifest or image hash

Any mismatch is rejected as stale or unauthorized artifact state before issuance is considered.

### Issuance enforcement

After verifier checks succeed, `x/veid` calls `x/issuancepolicy.RecordVerifiedProof`.

That means a proof may verify successfully at the VEID layer while still recording:

- `paused`
- `verifier_scope_mismatch`
- `cap_exceeded`
- `no_active_policy`

The module does not bypass issuance policy state, and it does not mint by inference alone.

## Test coverage

The implemented coverage now includes:

- registry Msg/query lifecycle coverage
- registry deterministic activation-order coverage
- registry genesis init/export round-trip coverage
- issuance Msg/query lifecycle coverage
- issuance proof-record and active-policy behavior coverage
- issuance genesis init/export round-trip coverage
- cross-module VEID integration tests proving:
  - verified results record issuance only when governed registry and policy state agree
  - stale or unauthorized verifier artifact state is rejected before serving as a valid proof
  - policy verifier-scope mismatches prevent recorded issuance entitlement

## Operational expectations

For launch operations, the authoritative evidence should come from:

- registry verifier hashes and statuses on-chain
- issuance policy and proof mint records on-chain
- VEID pipeline manifest and active pipeline state
- the launch packet evidence index maintained under [`_docs/operations`](../operations)

The production rule is fail-closed:

- no unapproved verifier activates
- no stale artifact state verifies
- no paused or mismatched policy records issuance as successful

*Maintained by the VirtEngine protocol and launch engineering tracks.*
