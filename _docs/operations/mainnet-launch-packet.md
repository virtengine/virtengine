# Mainnet Launch Packet (Evidence + Hashes)

> **Historical baseline:** This packet was assembled for the April 2026
> MainNet window, which did not proceed. It must be rebuilt with January 2027
> TestNet exit evidence and fresh approvals before the March 2027 MainNet
> launch. See `network-launch-schedule.md`.
>
> **Integrity note:** The manifest hashes below are the values recorded when
> the April 2026 packet closed. Supersession notices added to the historical
> Markdown records are intentionally not represented by those old hashes. The
> 2027 launch packets must generate a new manifest and new hashes; this packet
> must not be reused as a current integrity manifest.

Last updated: 2026-04-11
Owner: Release Management (Ops)

## Purpose
Centralized evidence bundle for mainnet launch approval. All checked-in
artifacts must be stored with SHA-256 hashes for immutability. At the time this
packet was closed, it
closes the execution-evidence gap, records the approved canonical allocation
set, and archives the checked-in MainNet genesis publication bundle. It
recorded `GO` for the scheduled 2026-04-18 UTC launch window, but that window
did not proceed and the record is not current launch authority.

## Evidence manifest

| Evidence ID | Artifact | Owner | Location | SHA-256 |
| --- | --- | --- | --- | --- |
| LAUNCH-CHK-001 | Launch readiness checklist | Release Manager | _docs/operations/mainnet-launch-readiness-checklist.md | 9de30c90ba8f89e04d12b4060444a5914778a5cff863ee1ac7abd4ae12f07c71 |
| LAUNCH-DR-001 | Dress rehearsal report | Ops Lead | _docs/operations/mainnet-dress-rehearsal-report.md | cc2bfdeb620554c1c15dc389593300698f027d0cdd4973ac58e6c67ab18ac120 |
| LAUNCH-DEC-001 | Go/No-Go decision record | Release Manager | _docs/operations/mainnet-go-no-go-decision.md | ed08773fd0b2157efd78e96bffb569c01f45c28fae21b6814ba5d7816b7f4291 |
| LAUNCH-AUDIT-001 | Launch evidence audit | Release Manager | _docs/operations/mainnet-launch-evidence-audit-2026-04-11.md | daa23f294f7d3af96248f4f815b258eb58b00c7842256c5a61d269e883880b93 |
| LAUNCH-CTRL-001 | Launch control record + approver set | Release Manager | _docs/operations/mainnet-launch-control-record-2026-04-11.md | bc96fcb9f40ba84ba7b06e6b13c44de9a19191eed16dc729916af9f5e04e1157 |
| ALLOC-CTRL-001 | Canonical allocation control record | Release Manager | _docs/operations/mainnet-allocation-control-record-2026-04-11.md | 5bc18819e1cf9365334f4a5f735d7a2051498d863ab6d4c1371ed7fb8548cc2f |
| LAUNCH-RUNBOOK | Mainnet launch runbook | Ops Lead | _docs/runbooks/mainnet-launch-runbook.md | 7c8a9bce3d731d216b14849bdcef09c4f4da5095687a379ff4ca20c19ff9d6e5 |
| GENESIS-CONFIG | Mainnet genesis parameters | Release Manager | config/mainnet/genesis-params.json | e1a516786a9b2644fbf1008886c13854be86110cf834fa6d0b65f950e639df88 |
| GENESIS-ALLOC | Mainnet genesis allocations | Release Manager | config/mainnet/genesis-allocations.json | 497623c2686ff0cb7524d090171cbc4aa1dc20da882ec8ade0caee6cd8d487a2 |
| GENESIS-RUNBOOK | Genesis ceremony runbook | Ops Lead | _docs/runbooks/mainnet-genesis-ceremony.md | 33ffdfa69b2d4e52f6853fc431d17aec3f4e03614515d6a39cca7a133f4c25df |
| VALIDATOR-ONBOARD | Validator onboarding runbook | Ops Lead | _docs/runbooks/validator-onboarding.md | 6af17de1ba6067e0f3897f7c399e6288c77492b1cca5e6c41b5ceaffb348832a |
| VALIDATOR-HW | Validator hardware requirements | Ops Lead | _docs/validators/hardware-requirements.md | 1ddf6822787e0088936d93ccb12330a08cc533a8c4d004af49a66dd0456ddcb6 |
| GENESIS-PUB-LOG | Mainnet genesis ceremony publication log | Ops Lead | output/mainnet-launch/2026-04-11/mainnet-genesis-ceremony.log | 78752fb55be6db7242d88720a12e6126c9ba205b1c1d32d0a09bc7557c185d9f |
| GENESIS-FINAL | Final mainnet genesis bundle | Ops Lead | artifacts/mainnet/genesis.json | a8d8a4a4f19882503265482c9433a6646d8dbbfe62f81c5945e81c32da9e6032 |
| GENESIS-FINAL-SHA | Final mainnet genesis checksum file | Ops Lead | artifacts/mainnet/genesis.sha256 | dedc0277172b4ef660799ffafaf71240e395a98fc43ded06592d85ac99c757d6 |
| GENTX-FINAL-SHA | Final gentx checksum file | Ops Lead | artifacts/mainnet/gentx.sha256 | b1ea3138d8ebd0963f4fd47e936e8e48b60cf2426b5f06e2497ce9cf10fa3a10 |
| GENESIS-MANIFEST | Final ceremony manifest | Ops Lead | artifacts/mainnet/ceremony-manifest.json | f7a4dceb7289c3e6affc50193cec1499f0f18b62fc480f16e665e66b6da1a0d9 |
| GENESIS-MANIFEST-SHA | Final ceremony manifest checksum file | Ops Lead | artifacts/mainnet/ceremony-manifest.sha256 | 30a2e984c9eec29a540b5553998b6f06477efd15c2fbbae038059982de0839ba |
| PRELAUNCH-AUTO | Pre-launch checklist pass log | Ops Lead | output/mainnet-launch/2026-04-11/prelaunch-checklist-pass.log | 24a8d01b85c8a0c9961b0c009d832741a52222805ac798353d9cf920814de660 |
| PRELAUNCH-SCRIPT | Pre-launch checklist automation source | Ops Lead | scripts/mainnet/prelaunch-checklist.sh | 9a4c31dc186baa8685c2a904e07d098c74c1aac0cf136d2dadd3d2384ba29b2c |
| UPGRADE-READINESS | Upgrade readiness validation log | Ops Lead | output/mainnet-launch/2026-04-11/upgrade-readiness.log | 8b80649fa66e310ed349fc6bbb05fe587dbfdb8747a5db0f4ee44edd0a81462c |
| PREREQ-VEID-E2E | VEID E2E onboarding results | VEID Lead | _docs/operations/mainnet-veid-e2e-report-2026-04-11.md | e9ba3d4715b9e7630c559389824ee36b3161f6963226eecd4af9cc27cac6f11b |
| PREREQ-PROVIDER-E2E | Provider marketplace E2E results | Provider Lead | _docs/operations/mainnet-provider-hpc-e2e-report-2026-04-11.md | ff7b1420ec9a1b02fe09ca90df67f692b248045a56f3ccc5eb9787843ee4a949 |
| PREREQ-SEC-ML | ML/verification security review | Security Lead | _docs/security-review-8d-ml-verification.md | 3a70c85b01a3505e67861b920d042487fd1a189cfba552904cb50564ee8b780b |
| PREREQ-BILL-REC | Billing reconciliation report | Finance Lead | _docs/operations/mainnet-finance-reconciliation-report-2026-04-11.md | 903e0eecc4b08c082dee4776506f84b18dfaa4fde26054a6b4088e357ac6951b |
| DR-RUNBOOK | Disaster recovery runbook | Ops Lead | _docs/disaster-recovery.md | 4b3f6dd9bdcb94aa6ebd7540008ef4bdace31dd1cd78f55b458a195e9d215fb6 |
| BACKUP-RESTORE | Backup/restore drill report | Ops Lead | _docs/operations/mainnet-backup-restore-drill-report-2026-04-11.md | fd3477008dea9f994ccd403fba5fcc66330593ed01780971fd6711dc1f3f770f |
| COMMS-PLAN | External comms plan + status page drafts | Product Lead | _docs/operations/mainnet-launch-comms-packet-2026-04-11.md | a5cdf93cfc28939ec0d3393d15122d31aa0efea2ad283be0ff8f08d85ece3ff9 |

## Hashing procedure

Run from repo root:

```bash
sha256sum _docs/operations/mainnet-launch-readiness-checklist.md
sha256sum _docs/operations/mainnet-dress-rehearsal-report.md
sha256sum _docs/operations/mainnet-go-no-go-decision.md
sha256sum _docs/operations/mainnet-launch-evidence-audit-2026-04-11.md
sha256sum _docs/operations/mainnet-launch-control-record-2026-04-11.md
sha256sum _docs/operations/mainnet-allocation-control-record-2026-04-11.md
sha256sum _docs/runbooks/mainnet-launch-runbook.md
sha256sum config/mainnet/genesis-params.json
sha256sum config/mainnet/genesis-allocations.json
sha256sum _docs/runbooks/mainnet-genesis-ceremony.md
sha256sum _docs/runbooks/validator-onboarding.md
sha256sum _docs/validators/hardware-requirements.md
sha256sum output/mainnet-launch/2026-04-11/mainnet-genesis-ceremony.log
sha256sum artifacts/mainnet/genesis.json
sha256sum artifacts/mainnet/genesis.sha256
sha256sum artifacts/mainnet/gentx.sha256
sha256sum artifacts/mainnet/ceremony-manifest.json
sha256sum artifacts/mainnet/ceremony-manifest.sha256
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
- The packet historically recorded `GO` for the scheduled 2026-04-18 UTC
  launch window and the 2026-04-19 UTC backup window; neither window proceeded
- Current public documentation must use the January 2027 TestNet and March
  2027 MainNet schedule in `network-launch-schedule.md`
