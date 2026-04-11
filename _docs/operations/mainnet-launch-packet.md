# Mainnet Launch Packet (Evidence + Hashes)

Last updated: 2026-04-11
Owner: Release Management (Ops)

## Purpose
Centralized evidence bundle for mainnet launch approval. All checked-in
artifacts must be stored with SHA-256 hashes for immutability. This packet now
closes the execution-evidence gap and records the current repository posture as
`HOLD` pending final signed allocation addresses and regenerated mainnet
genesis artifacts.

## Evidence manifest

| Evidence ID | Artifact | Owner | Location | SHA-256 |
| --- | --- | --- | --- | --- |
| LAUNCH-CHK-001 | Launch readiness checklist | Release Manager | _docs/operations/mainnet-launch-readiness-checklist.md | 9b7aabda88d5c19dc8fc5264b4213d7f9c52793268f26c4b4750861ae2f332c8 |
| LAUNCH-DR-001 | Dress rehearsal report | Ops Lead | _docs/operations/mainnet-dress-rehearsal-report.md | aa4f685245d8e4abc457f0a393aa09ff084e4d77d1538ae3a5840baceba8dbfc |
| LAUNCH-DEC-001 | Go/No-Go decision record | Release Manager | _docs/operations/mainnet-go-no-go-decision.md | 907aca6e2ae90a2d0818c174fe64f14346485e8daaabe4cc0bd45fa05de5672b |
| LAUNCH-AUDIT-001 | Launch evidence audit | Release Manager | _docs/operations/mainnet-launch-evidence-audit-2026-04-11.md | 0d6c98362db8ef3d3699123ea96736eb210270b95c7f303d6d28eec02e499927 |
| LAUNCH-CTRL-001 | Launch control record + approver set | Release Manager | _docs/operations/mainnet-launch-control-record-2026-04-11.md | b58ca54b1bc485d5ce1673b1309c8959fae5622030828b086b77e0ea65482db1 |
| LAUNCH-RUNBOOK | Mainnet launch runbook | Ops Lead | _docs/runbooks/mainnet-launch-runbook.md | 7c8a9bce3d731d216b14849bdcef09c4f4da5095687a379ff4ca20c19ff9d6e5 |
| GENESIS-CONFIG | Mainnet genesis parameters | Release Manager | config/mainnet/genesis-params.json | df30dd8a255f4f3f114bd678ea82e4339e9091b92a1cf0d98f764157a919f795 |
| GENESIS-RUNBOOK | Genesis ceremony runbook | Ops Lead | _docs/runbooks/mainnet-genesis-ceremony.md | 6d5c63fc15e832a619484c5c7a135f7a440b13aae3ad62d101879c0c15a2eff2 |
| VALIDATOR-ONBOARD | Validator onboarding runbook | Ops Lead | _docs/runbooks/validator-onboarding.md | 6af17de1ba6067e0f3897f7c399e6288c77492b1cca5e6c41b5ceaffb348832a |
| VALIDATOR-HW | Validator hardware requirements | Ops Lead | _docs/validators/hardware-requirements.md | 1ddf6822787e0088936d93ccb12330a08cc533a8c4d004af49a66dd0456ddcb6 |
| PRELAUNCH-AUTO | Pre-launch checklist pass log | Ops Lead | output/mainnet-launch/2026-04-11/prelaunch-checklist-pass.log | 24a8d01b85c8a0c9961b0c009d832741a52222805ac798353d9cf920814de660 |
| PRELAUNCH-SCRIPT | Pre-launch checklist automation source | Ops Lead | scripts/mainnet/prelaunch-checklist.sh | 9a4c31dc186baa8685c2a904e07d098c74c1aac0cf136d2dadd3d2384ba29b2c |
| UPGRADE-READINESS | Upgrade readiness validation log | Ops Lead | output/mainnet-launch/2026-04-11/upgrade-readiness.log | 8b80649fa66e310ed349fc6bbb05fe587dbfdb8747a5db0f4ee44edd0a81462c |
| PREREQ-VEID-E2E | VEID E2E onboarding results | VEID Lead | _docs/operations/mainnet-veid-e2e-report-2026-04-11.md | e9ba3d4715b9e7630c559389824ee36b3161f6963226eecd4af9cc27cac6f11b |
| PREREQ-PROVIDER-E2E | Provider marketplace E2E results | Provider Lead | _docs/operations/mainnet-provider-hpc-e2e-report-2026-04-11.md | ff7b1420ec9a1b02fe09ca90df67f692b248045a56f3ccc5eb9787843ee4a949 |
| PREREQ-SEC-ML | ML/verification security review | Security Lead | _docs/security-review-8d-ml-verification.md | 3a70c85b01a3505e67861b920d042487fd1a189cfba552904cb50564ee8b780b |
| PREREQ-BILL-REC | Billing reconciliation report | Finance Lead | _docs/operations/mainnet-finance-reconciliation-report-2026-04-11.md | 903e0eecc4b08c082dee4776506f84b18dfaa4fde26054a6b4088e357ac6951b |
| DR-RUNBOOK | Disaster recovery runbook | Ops Lead | _docs/disaster-recovery.md | 4b3f6dd9bdcb94aa6ebd7540008ef4bdace31dd1cd78f55b458a195e9d215fb6 |
| BACKUP-RESTORE | Backup/restore drill report | Ops Lead | _docs/operations/mainnet-backup-restore-drill-report-2026-04-11.md | fd3477008dea9f994ccd403fba5fcc66330593ed01780971fd6711dc1f3f770f |
| COMMS-PLAN | External comms plan + status page drafts | Product Lead | _docs/operations/mainnet-launch-comms-packet-2026-04-11.md | 3fd94304163538bb57813e9e4a05646ce598044871c267c28df19fecda0a0cda |

## Hashing procedure

Run from repo root:

```bash
sha256sum _docs/operations/mainnet-launch-readiness-checklist.md
sha256sum _docs/operations/mainnet-dress-rehearsal-report.md
sha256sum _docs/operations/mainnet-go-no-go-decision.md
sha256sum _docs/operations/mainnet-launch-evidence-audit-2026-04-11.md
sha256sum _docs/operations/mainnet-launch-control-record-2026-04-11.md
sha256sum _docs/runbooks/mainnet-launch-runbook.md
sha256sum config/mainnet/genesis-params.json
sha256sum _docs/runbooks/mainnet-genesis-ceremony.md
sha256sum _docs/runbooks/validator-onboarding.md
sha256sum _docs/validators/hardware-requirements.md
sha256sum output/mainnet-launch/2026-04-11/prelaunch-checklist-pass.log
sha256sum scripts/mainnet/prelaunch-checklist.sh
sha256sum output/mainnet-launch/2026-04-11/upgrade-readiness.log
sha256sum _docs/operations/mainnet-veid-e2e-report-2026-04-11.md
sha256sum _docs/operations/mainnet-provider-hpc-e2e-report-2026-04-11.md
sha256sum _docs/security-review-8d-ml-verification.md
sha256sum _docs/operations/mainnet-finance-reconciliation-report-2026-04-11.md
sha256sum _docs/disaster-recovery.md
sha256sum _docs/operations/mainnet-backup-restore-drill-report-2026-04-11.md
sha256sum _docs/operations/mainnet-launch-comms-packet-2026-04-11.md
```

Record the hash outputs in the Evidence manifest table.

## Storage
- Store all evidence artifacts in the release evidence archive (immutable
  storage)
- Update the launch runbook to reference this packet
- Launch remains `HOLD` until signed allocation addresses are inserted into
  `config/mainnet/genesis-allocations.json`, the final mainnet genesis bundle
  is rebuilt, and `LAUNCH-DEC-001` is moved from `HOLD` to `GO`
