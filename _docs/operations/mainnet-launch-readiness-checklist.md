# Mainnet Launch Readiness Checklist

Last updated: 2026-04-11
Owner: Release Management (Ops)

## Purpose
Provide a single, auditable checklist for mainnet launch readiness and
go/no-go decisioning. This checklist is the authoritative gate for production
launch approval.

## Prerequisites (must be complete before go/no-go)
- `6B test(veid)`: complete - see
  `_docs/operations/mainnet-veid-e2e-report-2026-04-11.md`
- `6C test(provider)`: complete - see
  `_docs/operations/mainnet-provider-hpc-e2e-report-2026-04-11.md`
- `6D test(veid)`: complete - see `_docs/security-review-8d-ml-verification.md`
- `6E feat(escrow)`: complete - see
  `_docs/operations/mainnet-finance-reconciliation-report-2026-04-11.md`
- `22B docs(mainnet)`: complete - launch control, ceremony tooling, validator
  onboarding, canonical allocation approval, and final mainnet genesis
  publication are complete

## A. Required sign-offs

| Sign-off area | Required approver role | Status | Evidence link(s) |
| --- | --- | --- | --- |
| Security | Security Lead | PASS - `SEC-01` approved the rehearsal bundle on 2026-04-11; no launch-domain security blocker remains in repo evidence | `_docs/security-review-8d-ml-verification.md`; `_docs/security-checklist.md`; `_docs/threat-model.md`; `_docs/operations/mainnet-veid-e2e-report-2026-04-11.md`; `_docs/operations/mainnet-launch-control-record-2026-04-11.md` |
| Compliance/Privacy | Compliance Lead | PASS - `COMP-01` approved the checked-in privacy/compliance pack on 2026-04-11 | `GDPR_COMPLIANCE.md`; `DATA_PROCESSING_AGREEMENT.md`; `DATA_INVENTORY.md`; `PRIVACY_POLICY.md`; `BIOMETRIC_DATA_ADDENDUM.md`; `_docs/operations/mainnet-launch-control-record-2026-04-11.md` |
| Operations/SRE | Ops Lead | PASS - `OPS-01` approved launch/freeze windows, dress rehearsal bundle, and restore drill evidence on 2026-04-11 | `_docs/runbooks/mainnet-launch-runbook.md`; `_docs/disaster-recovery.md`; `_docs/slos-and-playbooks.md`; `_docs/operations/mainnet-dress-rehearsal-report.md`; `_docs/operations/mainnet-backup-restore-drill-report-2026-04-11.md`; `_docs/operations/mainnet-launch-control-record-2026-04-11.md` |
| Finance/Billing | Finance Lead | PASS - `FIN-01` approved reconciliation, payout, and dispute-flow evidence on 2026-04-11 | `_docs/runbooks/finance-reconciliation-runbook.md`; `_docs/hpc-billing-rules.md`; `_docs/billing-policy.md`; `_docs/operations/mainnet-finance-reconciliation-report-2026-04-11.md`; `_docs/operations/mainnet-launch-control-record-2026-04-11.md` |
| Product/Release | Product Lead | PASS - `PROD-01` approved the launch comms packet, final genesis publication package, and scheduled `GO` state on 2026-04-11 | `_docs/operations/mainnet-launch-packet.md`; `_docs/operations/mainnet-go-no-go-decision.md`; `_docs/operations/mainnet-launch-comms-packet-2026-04-11.md`; `_docs/operations/mainnet-launch-control-record-2026-04-11.md`; `_docs/operations/mainnet-allocation-control-record-2026-04-11.md` |

## B. Sign-off to artifacts/tests mapping

| Sign-off area | Required artifacts/tests | Owner role | Validation method |
| --- | --- | --- | --- |
| Security | ML/verification security review, threat model update, VEID rehearsal evidence, key-management review | Security Lead | Review evidence bundle + approve findings remediation |
| Compliance/Privacy | GDPR/PII compliance evidence, consent flows, data retention policy, biometric addendum | Compliance Lead | Compliance review + sign-off record |
| Operations/SRE | DR runbook validation, backup restore smoke, monitoring/alerting review, dress rehearsal | Ops Lead | Staging rehearsal + drill sign-off |
| Finance/Billing | Billing reconciliation report, payout evidence, dispute workflow test, treasury controls | Finance Lead | Reconciliation evidence + finance sign-off |
| Product/Release | Go/no-go scorecard, launch comms readiness, status page prep | Product Lead | Release review meeting |

## C. Launch windows, freeze periods, rollback criteria

### Launch windows (UTC)
- Primary window: 2026-04-18 05:00 to 2026-04-18 07:00
- Backup window: 2026-04-19 05:00 to 2026-04-19 07:00
- Blackout windows:
  - 2026-04-14 00:00 to 2026-04-14 23:59 (finance close)
  - 2026-04-20 00:00 to 2026-04-20 23:59 (post-launch reserve / no changes)

### Code/config freeze
- Freeze starts: 2026-04-17 12:00
- Freeze ends: 2026-04-19 12:00
- Allowed changes during freeze: security hotfixes + release manager approved
  critical fixes only

### Rollback criteria (trigger any = NO-GO or ROLLBACK)
- Consensus failure: >3 consecutive halted blocks in 10 min window
- Validator participation: <67% voting power for >10 min
- Chain health: sustained block time >2x target for >20 min
- Critical service: VEID verification error rate >2% over 15 min
- Billing integrity: reconciliation mismatch >0.5% or dispute workflow failing

## D. Upgrade procedures, backups, and DR validation

Checklist:
- [x] Upgrade runbook reviewed and rehearsal window timeboxed
- [x] Canonical genesis input checksum verified and stored
- [x] Canonical allocation addresses approved and archived
- [x] Validator upgrade registry and matrix validation passed
- [x] Backup/restore drill completed for rehearsal scope
- [x] DR runbook validated with restore smoke and control review
- [x] Key management and HSM procedures reviewed
- [x] Final genesis publication bundle rebuilt and hashed

Reference runbooks and evidence:
- `_docs/disaster-recovery.md`
- `_docs/business-continuity.md`
- `_docs/verification-runbook.md`
- `_docs/operations/lifecycle-control.md`
- `_docs/runbooks/validator-onboarding.md`
- `_docs/runbooks/mainnet-genesis-ceremony.md`
- `docs/operations/runbooks/UPGRADE_PROCEDURES.md`
- `_docs/operations/mainnet-backup-restore-drill-report-2026-04-11.md`
- `_docs/operations/mainnet-allocation-control-record-2026-04-11.md`
- `_docs/operations/mainnet-launch-control-record-2026-04-11.md`
- `artifacts/mainnet/genesis.json`
- `artifacts/mainnet/genesis.sha256`
- `artifacts/mainnet/gentx.sha256`
- `artifacts/mainnet/ceremony-manifest.json`
- `artifacts/mainnet/ceremony-manifest.sha256`
- `output/mainnet-launch/2026-04-11/upgrade-readiness.log`
- `output/mainnet-launch/2026-04-11/mainnet-genesis-ceremony.log`
- `output/mainnet-launch/2026-04-11/prelaunch-checklist-pass.log`

## E. Dress rehearsal (staging) summary

- [x] Full dress rehearsal executed in staging harnesses
- [x] Cutover steps timed and recorded
- [x] Rollback rehearsal completed for repository smoke scope
- [x] Results captured in `_docs/operations/mainnet-dress-rehearsal-report.md`

## F. Communications plan

### Internal comms
- Incident channel: `#ve-mainnet-launch`
- War room bridge owner: `PROD-01`; live bridge reference is attached through
  the product-control packet, not stored directly in git
- Pager rotation confirmed with on-call roster

### External comms
- Status page: `status.virtengine.com`
- Customer notice: scheduled-maintenance draft prepared in
  `_docs/operations/mainnet-launch-comms-packet-2026-04-11.md`
- Partner notice: validator + provider hold/go-live drafts prepared in
  `_docs/operations/mainnet-launch-comms-packet-2026-04-11.md`

### Status page templates
- Scheduled maintenance: populated in
  `_docs/operations/mainnet-launch-comms-packet-2026-04-11.md`
- Hold update: populated in
  `_docs/operations/mainnet-launch-comms-packet-2026-04-11.md`
- Rollback update: populated in
  `_docs/operations/mainnet-launch-comms-packet-2026-04-11.md`

## G. Go/no-go checklist scoring and decision process

### Scoring
- P0 (must-pass): Any failure = NO-GO
- P1 (major): Up to 2 warnings allowed with mitigation owners
- P2 (minor): Document and track; does not block

### Decision process
1. Release Manager compiles scorecard + evidence packet
2. Sign-off owners review evidence and record approval
3. Go/No-Go meeting held with quorum (Security, Compliance, Ops, Finance,
   Product)
4. Final decision recorded in `_docs/operations/mainnet-go-no-go-decision.md`

## H. Launch packet evidence capture

- Evidence bundle location: `_docs/operations/mainnet-launch-packet.md`
- Hashing: SHA-256 for every artifact; include command output and timestamps

## I. Go/no-go meeting record

- Meeting record: `_docs/operations/mainnet-go-no-go-decision.md`
- Required attendees: Security Lead, Compliance Lead, Ops Lead, Finance Lead,
  Product Lead

## J. Publish final checklist + archive evidence

- [x] Checklist finalized and signed
- [x] Launch packet hashes verified and archived
- [x] Runbook updated with final links
- [x] Decision record published and indexed

Current launch state: `GO` for the scheduled primary launch window on
2026-04-18 05:00 to 07:00 UTC, with backup on 2026-04-19 05:00 to 07:00 UTC.
Repository evidence shows the final canonical allocations and genesis
publication bundle are complete; public mainnet availability should still be
described as scheduled rather than already live until the window begins.
