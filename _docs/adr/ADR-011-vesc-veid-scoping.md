# ADR-011: VESC-Scoped VEID Condition

## Status

Proposed

## Date

2026-09-27

## Context

VESC is a planned but currently undeveloped and unpublished part of the
VirtEngine estate. A repository and site search conducted on 2026-09-26 found
no VESC references in code, configuration, or published documentation, and
that absence is the correct current state: there is no VESC system to
describe, deploy, or gate on today.

This record exists so that if VESC is ever developed and published, one design
rule is already settled in advance: any identity condition attached to VESC
belongs to VESC alone. It scopes a future possibility; it does not describe an
existing system, and nothing in this document is a guarantee about what any
shipped system does.

The marketplace already governs identity on a per-offer opt-in basis: a
provider may choose to require identity verification for their own listing,
and listings without such an opt-in are unaffected. That per-offer model is
independent of VESC and is not changed by anything decided here.

## Decision

If VESC's eventual purpose requires a VirtEngine identity (VEID) condition,
that condition applies to that specific VESC cluster only, is disclosed to
users verbatim before it can affect them, and ships with a working
appeal and recovery path. A VESC VEID condition is never inherited by, and
never implied for, any other surface.

Why a VEID condition could be needed there: VESC, as currently conceived, is
a curated compute environment whose purpose may include workloads or data
handling for which the operator must know who ran what. If that purpose is
confirmed during development, a VEID check at the point of entry to that
cluster is the mechanism that ties usage of that cluster to an accountable
identity. The condition exists to serve VESC's own purpose, not as a general
statement about who may use VirtEngine.

The exact disclosure text to be shown to users, verbatim-ready, is:

Access to this cluster requires a verified VirtEngine identity (VEID). This
is a condition of this cluster only. Other clusters, and marketplace listings
that do not opt in to identity verification, are not affected by this
requirement. If you cannot complete verification, see the appeal and recovery
path below instead of retrying entry.

The appeal and recovery path for a user who cannot satisfy the condition is:

1. The entry denial names the failed check in plain language and points to
   this document and to the reverification flow, so the user can tell a
   fixable verification problem from a final ineligibility decision.
2. A fixable failure (expired evidence, mismatched attestation, incomplete
   verification session) returns the user to reverification; no separate
   appeal is needed and no penalty attaches to the account elsewhere.
3. A contested or final ineligibility decision can be appealed to the
   designated review contact published alongside the cluster, with the denial
   reference included. While the appeal is pending, the user loses nothing
   outside this cluster: existing marketplace access and per-offer
   participation continue unchanged.
4. Where appeal is unavailable or fails, recovery is redirection, not
   exclusion: the user is directed to equivalent clusters or marketplace
   offers that do not carry the VESC condition.

## Non-goals

This proposal is not a marketplace-wide requirement; marketplace listings keep per-offer opt-in. No marketplace listing gains, loses, or inherits a
VEID condition because of anything decided here.

The operator's rule from the assessment of 2026-09-26 governs all future work
under this record, quoted in full and verbatim:

"If VESC's eventual purpose requires VEID, that is a clearly disclosed condition for THAT specific cluster, with its own explanation and appeal/recovery path."

"VESC's future rule must never quietly become a marketplace-wide requirement."

## Consequences

- If VESC is developed with a VEID condition, reviewers have a fixed
  reference for what compliant scoping looks like: cluster-specific,
  disclosed verbatim, with appeal and recovery, and with the non-goal
  statement preserved.
- Any future proposal that extends, reuses, or cites a VESC VEID condition
  outside VESC scope contradicts this record and must be rejected or
  re-scoped, regardless of implementation convenience.
- If VESC is developed without any VEID condition, this record costs nothing:
  it imposes no code, no configuration, and no user-facing text.
- If VESC is never developed, this record remains a true statement about an
  undeveloped surface and requires no maintenance beyond that fact.

## Validation

- The repository and site search of 2026-09-26 confirms the precondition:
  no code, configuration, or published material presents a VESC VEID
  condition as an existing system guarantee.
- Any future pull request introducing a VESC VEID condition must keep it
  strictly inside VESC scope, preserve the non-goal statement above, and
  introduce no code path that gates the marketplace on VESC.
- This record contains no code and no deployment instructions, so there is
  nothing to execute; conformance is checked by reading future proposals
  against the Decision and Non-goals sections.
