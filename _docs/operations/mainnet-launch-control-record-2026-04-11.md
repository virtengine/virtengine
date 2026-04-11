# Mainnet Launch Control Record - 2026-04-11

Last updated: 2026-04-11
Owner: Release Management (Ops)

## Purpose
Archive the named approver set, approved launch window, approved freeze
window, and quorum-backed release-control outcome for the 2026-04-11
repository-backed mainnet rehearsal.

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
| Product/Release | `PROD-01` | Approved | 2026-04-11 06:12 | Comms packet accepted; final public launch remains on hold pending genesis allocations |

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

## Current hold basis
- Execution-evidence gaps are closed.
- Named launch approvers, launch windows, freeze windows, dress rehearsal,
  finance evidence, backup/restore evidence, and the comms packet are now
  archived in git.
- Final `GO` is withheld because `config/mainnet/genesis-allocations.json`
  still intentionally fails closed until signed canonical treasury,
  community-pool, and team-vesting addresses are inserted and the final
  `artifacts/mainnet/genesis.json` bundle is rebuilt.
