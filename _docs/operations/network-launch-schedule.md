# Network Launch Schedule and Promotion Gate

Last updated: 2026-11-04
Owner: Release Management (Ops)

## Current schedule

| Phase | Date | Control state |
| --- | --- | --- |
| TestNet launch | **20XX** (no date set) | Planned; launch date and launch approval not set. |
| Remediation and promotion review | **20XX** (no date set; after TestNet) | Reserved for observation, fixes, re-tests, evidence closure, release freeze, and final coordination. |
| MainNet launch | **20XX** (no date set; after remediation) | Planned; contingent on TestNet exit criteria and a fresh MainNet `GO` decision. |

**No launch date has been set.** `20XX` is a placeholder, not a target year or
month. Exact UTC start dates and control windows must be approved and recorded
in the relevant launch packet before activation.

## Environment boundary

### TestNet

TestNet is a public pre-production environment for multi-operator validation.
It is expected to exercise consensus, validator and provider onboarding,
upgrades, VEID, marketplace and settlement paths, observability, incident
response, backup, restore, and rollback procedures.

- State may be reset or replaced.
- TestNet tokens have no production value.
- TestNet balances and history are not guaranteed to migrate to MainNet.
- Availability and performance are validation targets, not production service
  commitments.

### MainNet

MainNet is the production environment based on approved release artifacts,
canonical genesis inputs, and production economic parameters. Its state is
intended to persist and support production activity, so activation requires a
separate, explicit approval.

TestNet launch or completion of an individual TestNet exercise must not be
treated as automatic MainNet authorization.

## Reason for the separation

The TestNet-to-MainNet separation creates an evidence-driven promotion gate.
The TestNet phase supplies public evidence. The interval that follows is
deliberately available for observation, defect remediation, security review,
repeated validation, artifact finalization, validator coordination, and a
release freeze. MainNet can launch only after those controls complete.

This sequencing reduces the chance that consensus, security, economic,
integration, capacity, or operating-model defects become embedded in
persistent production state.

## TestNet exit criteria for MainNet promotion

Release Management must record objective evidence for:

- stable consensus and agreed reliability/SLO results over the observation
  period;
- validator and provider onboarding across the intended launch cohort;
- successful upgrade, rollback, backup, restore, and incident-response drills;
- end-to-end VEID, marketplace, settlement, and governance validation;
- acceptable security, correctness, and performance results;
- closure and re-test of every launch-blocking finding;
- finalized release artifacts and canonical MainNet genesis inputs with
  published hashes;
- a fresh cross-functional MainNet readiness review and explicit `GO` record.

## Historical April 2026 records

The 2026-04-18 / 2026-04-19 MainNet windows did not proceed. Dated April 2026
control records, rehearsal reports, evidence audits, and communications drafts
remain historical evidence of the review conducted at that time. They do not
authorize either the TestNet launch or the MainNet launch, neither of which
has a date set.

This document supersedes month or window claims in those historical records
for current planning. The formal launch packets must be refreshed for each
environment before activation.
