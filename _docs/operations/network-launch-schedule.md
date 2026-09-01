# Network Launch Schedule and Promotion Gate

Last updated: 2026-09-01
Owner: Release Management (Ops)

## Current schedule

| Phase | Window | Control state |
| --- | --- | --- |
| TestNet launch | **January 2027** | Planned; exact UTC date and launch approval pending. |
| Remediation and promotion review | **February 2027** | Reserved for observation, fixes, re-tests, evidence closure, release freeze, and final coordination. |
| MainNet launch | **March 2027** | Planned; contingent on TestNet exit criteria and a fresh MainNet `GO` decision. |

The windows specify target months only. Exact UTC start dates and control
windows must be recorded in the relevant launch packet before activation.

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

The January-to-March separation creates an evidence-driven promotion gate.
January supplies public TestNet evidence. February is deliberately available
for observation, defect remediation, security review, repeated validation,
artifact finalization, validator coordination, and a release freeze. March is
the earliest MainNet launch window after those controls complete.

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
authorize either the January 2027 TestNet launch or the March 2027 MainNet
launch.

This document supersedes month or window claims in those historical records
for current planning. The formal launch packets must be refreshed for each
2027 environment before activation.
