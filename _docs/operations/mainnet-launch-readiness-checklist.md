# Mainnet Launch Readiness Checklist

Last updated: 2026-02-06
Owner: Release Management (Ops)

## Purpose
Provide a single, auditable checklist for mainnet launch readiness and go/no-go decisioning. This checklist is the authoritative gate for production launch approval.

## Prerequisites (must be complete before go/no-go)
- 6B test(veid): E2E VEID onboarding + verification flows
- 6C test(provider): E2E HPC/marketplace provider flow
- 6D test(veid): security review for ML/verification services
- 6E feat(escrow): billing reconciliation + dispute workflow
- 22B docs(mainnet): genesis config + validator onboarding + ceremony tooling

## A. Required sign-offs

| Sign-off area | Required approver role | Status | Evidence link(s) |
| --- | --- | --- | --- |
| Security | Security Lead | Pending | _docs/security-review-8d-ml-verification.md; _docs/security-checklist.md; _docs/threat-model.md |
| Compliance/Privacy | Compliance Lead | Pending | GDPR_COMPLIANCE.md; DATA_PROCESSING_AGREEMENT.md; DATA_INVENTORY.md; PRIVACY_POLICY.md; BIOMETRIC_DATA_ADDENDUM.md |
| Operations/SRE | Ops Lead | Pending | _docs/runbooks/mainnet-launch-runbook.md; _docs/disaster-recovery.md; _docs/slos-and-playbooks.md |
| Finance/Billing | Finance Lead | Pending | _docs/runbooks/finance-reconciliation-runbook.md; _docs/hpc-billing-rules.md; _docs/billing-policy.md |
| Product/Release | Product Lead | Pending | _docs/operations/mainnet-launch-packet.md |

## B. Sign-off to artifacts/tests mapping

| Sign-off area | Required artifacts/tests | Owner role | Validation method |
| --- | --- | --- | --- |
| Security | ML/verification security review, threat model update, vuln scan results, dependency audit | Security Lead | Review evidence bundle + approve findings remediation |
| Compliance/Privacy | GDPR/PII compliance evidence, consent flows, data retention policy, biometric addendum | Compliance Lead | Compliance review + legal sign-off |
| Operations/SRE | DR runbook validation, backup restore test, monitoring/alerting coverage, incident response test | Ops Lead | Staging rehearsal + drill sign-off |
| Finance/Billing | Billing reconciliation report, dispute workflow test, treasury controls | Finance Lead | Reconciliation evidence + finance sign-off |
| Product/Release | Go/no-go scorecard, launch comms readiness, status page prep | Product Lead | Release review meeting |

## C. Launch windows, freeze periods, rollback criteria

### Launch windows (UTC)
- Primary window: YYYY-MM-DD HH:MM–HH:MM UTC
- Backup window: YYYY-MM-DD HH:MM–HH:MM UTC
- Blackout windows: YYYY-MM-DD HH:MM–HH:MM UTC (document external conflicts)

### Code/config freeze
- Freeze starts: YYYY-MM-DD HH:MM UTC
- Freeze ends: YYYY-MM-DD HH:MM UTC
- Allowed changes during freeze: security hotfixes + release manager approved critical fixes only

### Rollback criteria (trigger any = NO-GO or ROLLBACK)
- Consensus failure: >3 consecutive halted blocks in 10 min window
- Validator participation: <67% voting power for >10 min
- Chain health: sustained block time >2x target for >20 min
- Critical service: VEID verification error rate >2% over 15 min
- Billing integrity: reconciliation mismatch >0.5% or dispute workflow failing

## D. Upgrade procedures, backups, and DR validation

Checklist:
- [ ] Upgrade runbook reviewed and timeboxed
- [ ] Genesis file checksum verified and stored
- [ ] Validator upgrade process tested on staging
- [ ] Backup/restore drill completed (RPO/RTO met)
- [ ] DR runbook validated with tabletop or live drill
- [ ] Key management and HSM procedures reviewed

Reference runbooks:
- _docs/disaster-recovery.md
- _docs/business-continuity.md
- _docs/verification-runbook.md
- _docs/operations/lifecycle-control.md
- _docs/tee-failover-strategy.md
- _docs/runbooks/validator-onboarding.md
- _docs/runbooks/mainnet-genesis-ceremony.md

## E. Dress rehearsal (staging) summary

- [ ] Full dress rehearsal executed in staging
- [ ] Cutover steps timed and recorded
- [ ] Rollback rehearsal completed
- [ ] Results captured in _docs/operations/mainnet-dress-rehearsal-report.md

## F. Communications plan

### Internal comms
- Incident channel: #ve-mainnet-launch (Slack)
- War room bridge: <insert video conference link>
- Pager rotation confirmed with on-call roster

### External comms
- Status page: status.virtengine.com (draft posts prepared)
- Customer notice: 7-day and 24-hour notices sent
- Partner notice: validator + provider announcement emails

### Status page templates
- Scheduled maintenance: "VirtEngine Mainnet Launch window scheduled for <DATE>. Expected impact: <IMPACT>."
- Incident update: "We are investigating issues impacting mainnet launch. Next update in <TIME>."
- Resolution: "Mainnet launch completed and services stable. Monitoring continues."

## G. Go/no-go checklist scoring and decision process

### Scoring
- P0 (must-pass): Any failure = NO-GO
- P1 (major): Up to 2 warnings allowed with mitigation owners
- P2 (minor): Document and track; does not block

### Decision process
1. Release Manager compiles scorecard + evidence packet
2. Sign-off owners review evidence and record approval
3. Go/No-Go meeting held with quorum (Security, Compliance, Ops, Finance, Product)
4. Final decision recorded in _docs/operations/mainnet-go-no-go-decision.md

## H. Launch packet evidence capture

- Evidence bundle location: _docs/operations/mainnet-launch-packet.md
- Hashing: SHA-256 for every artifact; include command output and timestamps

## I. Go/no-go meeting record

- Meeting record: _docs/operations/mainnet-go-no-go-decision.md
- Required attendees: Security Lead, Compliance Lead, Ops Lead, Finance Lead, Product Lead

## J. Publish final checklist + archive evidence

- [ ] Checklist finalized and signed
- [ ] Launch packet hashes verified and archived
- [ ] Runbook updated with final links
- [ ] Decision record published and indexed
