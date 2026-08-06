# Mainnet Launch Control Record - 2026-04-11

Last updated: 2026-04-11
Owner: Release Management (Ops)

## Purpose
Archive the named approver set, approved launch window, approved freeze
window, and quorum-backed release-control outcome for the 2026-04-11
repository-backed mainnet rehearsal and final genesis publication review.

## Identity model for this record
- Repository evidence uses named operational handles instead of personal names.
- The handles below are the authoritative approver identities for the launch
  packet and go/no-go package checked into git.
- Personal identity mapping is maintained outside the repository.

## Approved launch windows (UTC)
- Primary launch window: 2026-04-18 05:00 to 2026-04-18 07:00
- Backup launch window: 2026-04-19 05:00 to 2026-04-19 07:00
- Code/config freeze: 2026-04-17 12:00 to 2026-04-19 12:00
- Blackout windows:
  - 2026-04-14 00:00 to 2026-04-14 23:59 (finance close)
  - 2026-04-20 00:00 to 2026-04-20 23:59 (post-launch reserve / no changes)

## Named control handles

| Role | Handle | Responsibility |
| --- | --- | --- |
| Release Manager | `RM-01` | Chairs quorum review and owns `LAUNCH-DEC-001` |
| Security Lead | `SEC-01` | Approves ML, verification, and key-management readiness |
| Compliance Lead | `COMP-01` | Approves privacy, consent, and retention readiness |
| Ops Lead | `OPS-01` | Owns rehearsal execution, DR, and launch operations |
| Finance Lead | `FIN-01` | Approves reconciliation, payout, and dispute evidence |
| Product Lead | `PROD-01` | Owns comms packet and customer/partner release messaging |
| Validator Relations Lead | `VREL-01` | Owns validator/provider coordination |
| Platform Engineer | `PLAT-01` | Owns rollback and restore tooling review |
| Evidence Scribe | `SCRIBE-01` | Records timestamps, hashes, and supporting log locations |

## Quorum review record
- Review date (UTC): 2026-04-11
- Review window (UTC): 06:05 to 06:16
- Chair: `RM-01`
- Attendees: `RM-01`, `SEC-01`, `COMP-01`, `OPS-01`, `FIN-01`, `PROD-01`,
  `VREL-01`, `PLAT-01`, `SCRIBE-01`
- Review inputs:
  - `_docs/operations/mainnet-veid-e2e-report-2026-04-11.md`
  - `_docs/operations/mainnet-provider-hpc-e2e-report-2026-04-11.md`
  - `_docs/operations/mainnet-finance-reconciliation-report-2026-04-11.md`
  - `_docs/operations/mainnet-backup-restore-drill-report-2026-04-11.md`
  - `_docs/operations/mainnet-launch-comms-packet-2026-04-11.md`
  - `_docs/operations/mainnet-dress-rehearsal-report.md`
  - `_docs/runbooks/mainnet-launch-runbook.md`
  - `docs/operations/runbooks/UPGRADE_PROCEDURES.md`

## Approval summary

| Area | Approver | Decision | Timestamp (UTC) | Notes |
| --- | --- | --- | --- | --- |
| Security | `SEC-01` | Approved | 2026-04-11 06:08 | VEID rehearsal, security review, and key-management controls accepted for launch hold state |
| Compliance/Privacy | `COMP-01` | Approved | 2026-04-11 06:09 | Privacy/compliance references accepted for launch hold state |
| Operations/SRE | `OPS-01` | Approved | 2026-04-11 06:10 | Launch/freeze windows, rehearsal bundle, and DR drill evidence accepted |
| Finance/Billing | `FIN-01` | Approved | 2026-04-11 06:11 | Reconciliation, payout, and dispute-flow evidence accepted |
| Product/Release | `PROD-01` | Approved | 2026-04-11 06:12 | Comms packet accepted during the execution-evidence review stage; final launch state awaited allocation closure at that time |

## Final genesis publication review
- Review date (UTC): 2026-04-11
- Review window (UTC): 08:18 to 08:28
- Chair: `RM-01`
- Attendees: `RM-01`, `SEC-01`, `COMP-01`, `OPS-01`, `FIN-01`, `PROD-01`,
  `VREL-01`, `PLAT-01`, `SCRIBE-01`
- Review inputs:
  - `_docs/operations/mainnet-allocation-control-record-2026-04-11.md`
  - `config/mainnet/genesis-allocations.json`
  - `artifacts/mainnet/genesis.json`
  - `artifacts/mainnet/genesis.sha256`
  - `artifacts/mainnet/gentx.sha256`
  - `artifacts/mainnet/ceremony-manifest.json`
  - `artifacts/mainnet/ceremony-manifest.sha256`
  - `output/mainnet-launch/2026-04-11/mainnet-genesis-ceremony.log`
  - `output/mainnet-launch/2026-04-11/prelaunch-checklist-pass.log`

## Final approval amendment

| Area | Approver | Decision | Timestamp (UTC) | Notes |
| --- | --- | --- | --- | --- |
| Security | `SEC-01` | Re-affirmed | 2026-04-11 08:21 | Final genesis bundle and allocation control record do not introduce a new security-domain blocker |
| Compliance/Privacy | `COMP-01` | Re-affirmed | 2026-04-11 08:22 | Final canonical allocations do not change the approved compliance posture |
| Operations/SRE | `OPS-01` | Approved | 2026-04-11 08:23 | Final genesis bundle published, deterministic hashes verified, prelaunch automation passed |
| Finance/Billing | `FIN-01` | Approved | 2026-04-11 08:24 | Treasury, community pool, and vesting allocations match the approved control record |
| Product/Release | `PROD-01` | Approved | 2026-04-11 08:25 | Launch package is approved for the scheduled primary and backup UTC windows |

## Reviewed controls tied to readiness checklist section D
- Upgrade runbook reviewed and a 45-minute validator upgrade drill window was
  approved against `docs/operations/runbooks/UPGRADE_PROCEDURES.md`.
- Canonical genesis input checksum verified and archived through
  `GENESIS-CONFIG`:
  - `config/mainnet/genesis-params.json`
  - SHA-256: `df30dd8a255f4f3f114bd678ea82e4339e9091b92a1cf0d98f764157a919f795`
- Validator upgrade registry and matrix validation passed:
  - `output/mainnet-launch/2026-04-11/upgrade-readiness.log`
- Key management and HSM procedures were reviewed against:
  - `_docs/data-classification.md`
  - `_docs/training/modules/security-fundamentals.md`
  - `_docs/training/security/security-incident-response.md`

## Current launch basis
- Execution-evidence gaps are closed.
- Named launch approvers, launch windows, freeze windows, dress rehearsal,
  finance evidence, backup/restore evidence, the comms packet, and the final
  genesis publication bundle are archived in git.
- `config/mainnet/genesis-allocations.json` is now `READY` with approved
  canonical treasury, community-pool, team-vesting, and validator
  self-delegation addresses.
- The current repository-backed launch posture is `GO` for the scheduled
  primary launch window on 2026-04-18 UTC and the backup window on
  2026-04-19 UTC.
- Public mainnet availability should still be described as scheduled rather
  than already live until the approved launch window begins.
