# Mainnet Launch Packet (Evidence + Hashes)

Last updated: 2026-02-06
Owner: Release Management (Ops)

## Purpose
Centralized evidence bundle for mainnet launch approval. All artifacts must be stored with SHA-256 hashes for immutability.

## Evidence manifest

| Evidence ID | Artifact | Owner | Location | SHA-256 |
| --- | --- | --- | --- | --- |
| LAUNCH-CHK-001 | Launch readiness checklist | Release Manager | _docs/operations/mainnet-launch-readiness-checklist.md | bec1ecfebd5b77a9c847f413edf08773f4f96d0a5b5b91e4c0796f3ed9494c20 |
| LAUNCH-DR-001 | Dress rehearsal report | Ops Lead | _docs/operations/mainnet-dress-rehearsal-report.md | af7287c4ce382c81301f98c7cc1f21b013b536c65c3c9789c884b942ceb8d930 |
| LAUNCH-DEC-001 | Go/No-Go decision record | Release Manager | _docs/operations/mainnet-go-no-go-decision.md | fce1db0df03ad26a748589db49108886401314286febf507e7877de2b374e8e4 |
| LAUNCH-RUNBOOK | Mainnet launch runbook | Ops Lead | _docs/runbooks/mainnet-launch-runbook.md | 08d151dccac89e0e6ca6ee54a46d10872fef410506e527a2e0b46b0f57ad8016 |
| PREREQ-VEID-E2E | VEID E2E onboarding results | VEID Lead | <path> | Pending |
| PREREQ-PROVIDER-E2E | Provider marketplace E2E results | Provider Lead | <path> | Pending |
| PREREQ-SEC-ML | ML/verification security review | Security Lead | _docs/security-review-8d-ml-verification.md | Pending |
| PREREQ-BILL-REC | Billing reconciliation report | Finance Lead | <path> | Pending |
| DR-RUNBOOK | Disaster recovery runbook | Ops Lead | _docs/disaster-recovery.md | Pending |
| BACKUP-RESTORE | Backup/restore drill report | Ops Lead | <path> | Pending |
| COMMS-PLAN | External comms plan + status page drafts | Product Lead | <path> | Pending |

## Hashing procedure

Run from repo root:

```
sha256sum _docs/operations/mainnet-launch-readiness-checklist.md
sha256sum _docs/operations/mainnet-dress-rehearsal-report.md
sha256sum _docs/operations/mainnet-go-no-go-decision.md
sha256sum _docs/runbooks/mainnet-launch-runbook.md
```

Record the hash outputs in the Evidence manifest table.

## Storage
- Store all evidence artifacts in the release evidence archive (immutable storage)
- Update the launch runbook to reference this packet
