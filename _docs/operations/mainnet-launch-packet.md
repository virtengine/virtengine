# Mainnet Launch Packet (Evidence + Hashes)

Last updated: 2026-02-06
Owner: Release Management (Ops)

## Purpose
Centralized evidence bundle for mainnet launch approval. All artifacts must be stored with SHA-256 hashes for immutability.

## Evidence manifest

| Evidence ID | Artifact | Owner | Location | SHA-256 |
| --- | --- | --- | --- | --- |
| LAUNCH-CHK-001 | Launch readiness checklist | Release Manager | _docs/operations/mainnet-launch-readiness-checklist.md | 538ee2e9e070f0fd858f0bdaa8154b1c750b365eeb38f1af0adb7f027b4534d3 |
| LAUNCH-DR-001 | Dress rehearsal report | Ops Lead | _docs/operations/mainnet-dress-rehearsal-report.md | af7287c4ce382c81301f98c7cc1f21b013b536c65c3c9789c884b942ceb8d930 |
| LAUNCH-DEC-001 | Go/No-Go decision record | Release Manager | _docs/operations/mainnet-go-no-go-decision.md | fce1db0df03ad26a748589db49108886401314286febf507e7877de2b374e8e4 |
| LAUNCH-RUNBOOK | Mainnet launch runbook | Ops Lead | _docs/runbooks/mainnet-launch-runbook.md | 80e5d90a7a7f063af273b908a0552f24714d4534d31cfbeb5756a7bcb50b2103 |
| GENESIS-CONFIG | Mainnet genesis parameters | Release Manager | config/mainnet/genesis-params.json | df30dd8a255f4f3f114bd678ea82e4339e9091b92a1cf0d98f764157a919f795 |
| GENESIS-RUNBOOK | Genesis ceremony runbook | Ops Lead | _docs/runbooks/mainnet-genesis-ceremony.md | 90781cf39aa9de3cf94001a8e04baff0403fba14c74b05c94f08596ce4e31ab6 |
| VALIDATOR-ONBOARD | Validator onboarding runbook | Ops Lead | _docs/runbooks/validator-onboarding.md | b58207beb2351f363de1147bc7339e7f60db083ecfc6ce05d12403e2c8d7fbc5 |
| VALIDATOR-HW | Validator hardware requirements | Ops Lead | _docs/validators/hardware-requirements.md | 1ddf6822787e0088936d93ccb12330a08cc533a8c4d004af49a66dd0456ddcb6 |
| PRELAUNCH-AUTO | Pre-launch checklist automation | Ops Lead | scripts/mainnet/prelaunch-checklist.sh | c47e8805c1e8bd03be0cd674c6ed1cb6f2aeddedef3eb1a1a708075479c17243 |
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
sha256sum config/mainnet/genesis-params.json
sha256sum _docs/runbooks/mainnet-genesis-ceremony.md
sha256sum _docs/runbooks/validator-onboarding.md
sha256sum _docs/validators/hardware-requirements.md
sha256sum scripts/mainnet/prelaunch-checklist.sh
```

Record the hash outputs in the Evidence manifest table.

## Storage
- Store all evidence artifacts in the release evidence archive (immutable storage)
- Update the launch runbook to reference this packet
