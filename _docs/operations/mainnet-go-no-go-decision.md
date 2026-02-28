# Mainnet Go/No-Go Decision Record

Last updated: 2026-04-11
Owner: Release Management (Ops)

## Meeting details
- Meeting date/time (UTC): 2026-04-11 08:18 to 2026-04-11 08:28
- Attendees: `RM-01`, `SEC-01`, `COMP-01`, `OPS-01`, `FIN-01`, `PROD-01`,
  `VREL-01`, `PLAT-01`, `SCRIBE-01`
- Chair: `RM-01`

## Inputs reviewed
- Readiness checklist: `_docs/operations/mainnet-launch-readiness-checklist.md`
- Dress rehearsal report: `_docs/operations/mainnet-dress-rehearsal-report.md`
- Evidence packet: `_docs/operations/mainnet-launch-packet.md`
- Evidence audit: `_docs/operations/mainnet-launch-evidence-audit-2026-04-11.md`
- Launch control record:
  `_docs/operations/mainnet-launch-control-record-2026-04-11.md`
- Allocation control record:
  `_docs/operations/mainnet-allocation-control-record-2026-04-11.md`
- Final genesis publication bundle:
  `artifacts/mainnet/genesis.json`,
  `artifacts/mainnet/genesis.sha256`,
  `artifacts/mainnet/gentx.sha256`,
  `artifacts/mainnet/ceremony-manifest.json`,
  `artifacts/mainnet/ceremony-manifest.sha256`
- Final readiness log:
  `output/mainnet-launch/2026-04-11/prelaunch-checklist-pass.log`

## Decision
- Decision: `GO`
- Conditions (if GO):
  - All P0 items PASS
  - P1 items have assigned owners and timeboxed mitigation
  - Rollback window and criteria confirmed
  - Signed canonical treasury, community-pool, team-vesting, and validator
    self-delegation addresses are inserted into
    `config/mainnet/genesis-allocations.json`
  - Final `artifacts/mainnet/genesis.json` bundle is rebuilt and published with
    current hashes
  - `scripts/mainnet/prelaunch-checklist.sh` passes without allow flags
  - Public launch messaging stays in “scheduled for approved UTC window” state
    until the 2026-04-18 launch window begins

## Checklist scoring summary

| Category | P0 status | P1 status | Notes |
| --- | --- | --- | --- |
| Security | PASS | Clear | Named approver archived and VEID/security evidence accepted. |
| Compliance | PASS | Clear | Named approver archived and privacy/compliance references accepted. |
| Operations | PASS | Clear | Launch window, freeze window, dress rehearsal, DR drill evidence, and final genesis publication evidence accepted. |
| Finance | PASS | Clear | Reconciliation, payout, dispute-path evidence, and canonical treasury allocations accepted. |
| Product/Release | PASS | Clear | Comms packet, quorum review, and scheduled launch posture accepted. |

## Sign-offs

| Role | Name | Decision | Date |
| --- | --- | --- | --- |
| Security Lead | `SEC-01` | Approved - final genesis publication does not reopen a launch-domain security blocker | 2026-04-11 |
| Compliance Lead | `COMP-01` | Approved - compliance launch posture remains clear after final allocation insertion | 2026-04-11 |
| Ops Lead | `OPS-01` | Approved - dress rehearsal, DR evidence, and final genesis publication bundle complete | 2026-04-11 |
| Finance Lead | `FIN-01` | Approved - reconciliation evidence and canonical treasury/community-pool/team allocations complete | 2026-04-11 |
| Product Lead | `PROD-01` | Approved - communications packet complete and launch state is `GO` for the scheduled UTC window | 2026-04-11 |

## Follow-up actions

| Action | Owner | Due date | Status |
| --- | --- | --- | --- |
| Insert signed canonical treasury, community-pool, and team-vesting addresses into `config/mainnet/genesis-allocations.json` | Release Manager | 2026-04-11 | Closed - completed 2026-04-11 07:44 UTC |
| Rebuild final mainnet genesis bundle and publish current hashes | Ops Lead | 2026-04-11 | Closed - completed 2026-04-11 |
| Re-run `scripts/mainnet/prelaunch-checklist.sh` after final genesis publication and record the control-window result | Release Manager | 2026-04-11 | Closed - completed 2026-04-11 |
