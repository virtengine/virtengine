# Mainnet Launch Runbook

Last updated: 2026-04-11
Owner: Operations

## Purpose
Operational control sheet for mainnet launch execution, launch-day smoke
verification, evidence capture, and rollback. This runbook is executable only
when every referenced A2 evidence ID already has a destination in
`_docs/operations/mainnet-launch-packet.md` and the operator running the step
has access to the paired operational artifact listed in this runbook.

## Launch Window
- Control window: `T-30` through `T+120`
- Stability window: `T+120` through `T+1440`
- `T0` is the UTC minute when `RM` records `GO` in `LAUNCH-DEC-001`
- External update cadence: every 15 minutes during any customer-facing launch
  activity or rollback
- Internal war-room cadence: every 10 minutes from `T-30` through `T+30`,
  then every 15 minutes until `T+120`
- Rollback decision SLA: no later than 10 minutes after any P0 rollback trigger

## Named Roles
- `RM` — Release Manager and launch chair; owns `LAUNCH-DEC-001`
- `OPS` — Operations Lead; owns `LAUNCH-RUNBOOK`, `LAUNCH-DR-001`,
  `PRELAUNCH-AUTO`, and execution control
- `SEC` — Security Lead; owns `PREREQ-SEC-ML`
- `COMP` — Compliance Lead; approves privacy and compliance readiness in
  `LAUNCH-CHK-001` and `LAUNCH-DEC-001`
- `FIN` — Finance Lead; owns `PREREQ-BILL-REC`
- `PROD` — Product Lead and comms lead; owns `COMMS-PLAN`
- `VREL` — Validator Relations Lead; coordinates validator acknowledgements and
  provider communications
- `PLAT` — Platform and infrastructure engineer; executes DR and rollback
  scripts under `scripts/dr/` and `scripts/rollback/`
- `SCRIBE` — Evidence scribe; timestamps every control decision and stores raw
  command output under the packet IDs listed below

## Command Contract
- Run all launch validation from repo root.
- Never use `scripts/mainnet/prelaunch-checklist.sh` with any allow flag.
- Never run launch smoke tests with `-short`; the VEID launch suites contain
  short-mode skips and are not valid launch evidence when short mode is active.
- Stop progression immediately on any failed command, missing evidence file,
  hash mismatch, or quorum dissent. Record the stop event in `LAUNCH-DEC-001`
  and move to the rollback section if the stop condition is customer-facing or
  consensus-impacting.

## Evidence ID Map

| A2 evidence ID | Required capture at launch | Source artifact or command |
| --- | --- | --- |
| `LAUNCH-CHK-001` | final launch checklist state, launch chair sign-off notes, preflight pass/fail summary | `_docs/operations/mainnet-launch-readiness-checklist.md` |
| `LAUNCH-DR-001` | minute log, block-height checks, launch timings, rollback notes, screenshots | `_docs/operations/mainnet-dress-rehearsal-report.md` |
| `LAUNCH-DEC-001` | GO, HOLD, NO-GO, and ROLLBACK declarations with UTC times | `_docs/operations/mainnet-go-no-go-decision.md` |
| `LAUNCH-RUNBOOK` | current runbook hash and revision used during launch | `_docs/runbooks/mainnet-launch-runbook.md` |
| `GENESIS-CONFIG` | hash and validation transcript for launch configuration inputs | `config/mainnet/genesis-params.json` and `scripts/mainnet/genesis-validate.sh` |
| `GENESIS-RUNBOOK` | ceremony transcript and genesis-hash verification notes | `_docs/runbooks/mainnet-genesis-ceremony.md`, `scripts/mainnet/genesis-ceremony.sh`, `scripts/mainnet/genesis-hash.sh` |
| `VALIDATOR-ONBOARD` | validator acknowledgement checklist and chain-start coordination notes | `_docs/runbooks/validator-onboarding.md` |
| `VALIDATOR-HW` | validator hardware profile confirmation | `_docs/validators/hardware-requirements.md` |
| `PRELAUNCH-AUTO` | raw output of `scripts/mainnet/prelaunch-checklist.sh` | `scripts/mainnet/prelaunch-checklist.sh` |
| `PREREQ-VEID-E2E` | raw `go test` output for launch-day VEID verification | `tests/e2e/veid_e2e_test.go` |
| `PREREQ-PROVIDER-E2E` | raw `go test` output for provider and marketplace verification | `tests/e2e/hpc_marketplace_e2e_test.go` |
| `PREREQ-SEC-ML` | launch-day acknowledgement that the security review remains valid | `_docs/security-review-8d-ml-verification.md` |
| `PREREQ-BILL-REC` | settlement and payout verification logs plus finance reconciliation sign-off | `_docs/runbooks/finance-reconciliation-runbook.md`, `tests/integration/settlement/full_pipeline_test.go`, `tests/integration/settlement/payout_flow_test.go` |
| `DR-RUNBOOK` | DR decision log and restore notes | `_docs/disaster-recovery.md` |
| `BACKUP-RESTORE` | backup and restore smoke-test output and any restore transcript | `docs/operations/runbooks/BACKUP_RESTORE.md`, `scripts/dr/backup-chain-state.sh`, `scripts/ci/backup-restore-smoke-test.sh` |
| `COMMS-PLAN` | internal announcements, status-page updates, customer notices, partner notices | `docs/sre/COMMUNICATION_TEMPLATES.md`, `docs/sre/INCIDENT_RESPONSE.md`, `docs/operations/runbooks/INCIDENT_RESPONSE.md` |

## Preconditions at `T-30`
- `RM` has `LAUNCH-DEC-001` open for live editing.
- `SCRIBE` has writable destinations for `LAUNCH-DR-001`, `LAUNCH-DEC-001`,
  `PREREQ-VEID-E2E`, `PREREQ-PROVIDER-E2E`, `PREREQ-BILL-REC`,
  `BACKUP-RESTORE`, and `COMMS-PLAN`.
- `OPS` has shell access to the scripts under `scripts/mainnet/`,
  `scripts/dr/`, `scripts/ci/`, and `scripts/rollback/`.
- `PROD` has `status.virtengine.com` access and the templates in
  `docs/sre/COMMUNICATION_TEMPLATES.md`.
- `VREL` has validator contact paths and the onboarding flow in
  `VALIDATOR-ONBOARD`.

## Minute-by-Minute Launch Playbook

| Minute | Lead | Action | A2 evidence IDs | Real operational artifact or command | Success rule |
| --- | --- | --- | --- | --- | --- |
| `T-30` | `RM`, `SCRIBE` | Open war room, start launch log, confirm role attendance, and declare the active command window. | `LAUNCH-DEC-001`, `LAUNCH-DR-001`, `COMMS-PLAN` | `docs/sre/COMMUNICATION_TEMPLATES.md`, `docs/operations/runbooks/INCIDENT_RESPONSE.md` | War room is active, `SCRIBE` is recording UTC timestamps, and all required roles are present. |
| `T-25` | `OPS` | Run the hard gate with no allow flags and archive the full output. | `LAUNCH-CHK-001`, `PRELAUNCH-AUTO`, `LAUNCH-DEC-001` | `bash scripts/mainnet/prelaunch-checklist.sh` | Command exits `0`; otherwise `RM` records `NO-GO`. |
| `T-20` | `OPS`, `VREL` | Validate genesis inputs, confirm validator onboarding packet, and verify the ceremony path still matches launch artifacts. | `GENESIS-CONFIG`, `GENESIS-RUNBOOK`, `VALIDATOR-ONBOARD`, `VALIDATOR-HW` | `bash scripts/mainnet/genesis-validate.sh --genesis artifacts/mainnet/genesis.json --checks config/mainnet/genesis-checks.json`, `bash scripts/mainnet/genesis-hash.sh --genesis artifacts/mainnet/genesis.json`, `_docs/runbooks/mainnet-genesis-ceremony.md`, `_docs/runbooks/validator-onboarding.md` | Validation succeeds, ceremony hash matches the approved value, and validators confirm readiness. |
| `T-15` | `SEC`, `COMP`, `FIN` | Reconfirm security, compliance, and finance sign-off inputs before any chain action. | `PREREQ-SEC-ML`, `PREREQ-BILL-REC`, `LAUNCH-CHK-001`, `LAUNCH-DEC-001` | `_docs/security-review-8d-ml-verification.md`, `GDPR_COMPLIANCE.md`, `DATA_PROCESSING_AGREEMENT.md`, `DATA_INVENTORY.md`, `PRIVACY_POLICY.md`, `BIOMETRIC_DATA_ADDENDUM.md`, `_docs/runbooks/finance-reconciliation-runbook.md` | No owner raises a blocking issue; any objection stops the launch. |
| `T-10` | `OPS`, `PLAT` | Verify DR readiness and rollback tooling before `GO`. | `DR-RUNBOOK`, `BACKUP-RESTORE`, `LAUNCH-DEC-001` | `bash scripts/ci/backup-restore-smoke-test.sh`, `scripts/dr/backup-chain-state.sh`, `scripts/rollback/argocd-rollback.sh`, `scripts/rollback/blue-green-switch.sh`, `scripts/rollback/terraform-rollback.sh`, `.github/workflows/ci.yaml` | Smoke test exits `0`, rollback scripts are present, and CI shell-lint coverage remains the reference static review path. |
| `T-05` | `PROD`, `SCRIBE` | Publish the scheduled-maintenance internal announcement and load the external update template with the launch timestamp. | `COMMS-PLAN`, `LAUNCH-DEC-001` | `docs/sre/COMMUNICATION_TEMPLATES.md`, `docs/sre/INCIDENT_RESPONSE.md` | Internal comms are posted and the next external update time is queued for `T+15`. |
| `T-02` | `RM` | Run final quorum poll and confirm rollback authority. | `LAUNCH-DEC-001`, `LAUNCH-CHK-001` | `_docs/operations/mainnet-go-no-go-decision.md`, `_docs/operations/mainnet-launch-readiness-checklist.md` | Unanimous `GO` or explicit `NO-GO`; no silent assent. |
| `T+00` | `RM`, `OPS` | Declare `GO`, freeze non-launch changes, and begin chain activation under launch control. | `LAUNCH-DEC-001`, `LAUNCH-DR-001`, `LAUNCH-RUNBOOK` | `_docs/runbooks/validator-onboarding.md`, `_docs/runbooks/mainnet-genesis-ceremony.md` | `GO` timestamp is written to `LAUNCH-DEC-001` and chain start actions begin immediately. |
| `T+05` | `OPS` | Confirm first healthy block production and baseline chain health. | `LAUNCH-DR-001`, `LAUNCH-DEC-001` | `virtengine status | jq`, `_docs/slos-and-playbooks.md` | New blocks are produced and no rollback trigger is active. |
| `T+10` | `VREL`, `OPS` | Verify validator participation, peer health, and coordination-channel acknowledgements. | `LAUNCH-DR-001`, `VALIDATOR-ONBOARD` | `_docs/runbooks/validator-onboarding.md`, `_docs/slos-and-playbooks.md` | Voting power and peer-health targets remain above rollback thresholds. |
| `T+15` | `PROD` | Publish first external update and first internal summary. | `COMMS-PLAN`, `LAUNCH-DEC-001`, `LAUNCH-DR-001` | `docs/sre/COMMUNICATION_TEMPLATES.md`, `docs/operations/runbooks/INCIDENT_RESPONSE.md` | Status page and war room both reflect the same UTC state. |
| `T+20` | `OPS` | Re-run targeted genesis and configuration verification after chain start. | `GENESIS-CONFIG`, `GENESIS-RUNBOOK`, `LAUNCH-DR-001` | `bash scripts/mainnet/genesis-hash.sh --genesis artifacts/mainnet/genesis.json`, `bash scripts/mainnet/genesis-validate.sh --genesis artifacts/mainnet/genesis.json --checks config/mainnet/genesis-checks.json` | App hash, genesis hash, and validation transcript still align with the approved launch inputs. |
| `T+25` | `OPS`, `SEC` | Run the deterministic VEID launch-grade suite without short mode and archive the full transcript. | `PREREQ-VEID-E2E`, `LAUNCH-DR-001` | `go test -tags=\"e2e.integration\" ./tests/e2e -run \"^TestVEIDE2E$\" -count=1`, `tests/e2e/veid_e2e_test.go` | Command exits `0`; any failure triggers `HOLD` and a rollback decision if customer-facing. |
| `T+35` | `OPS`, `VREL` | Run the deterministic provider/HPC marketplace launch-grade suite and archive the transcript. | `PREREQ-PROVIDER-E2E`, `LAUNCH-DR-001` | `go test -tags=\"e2e.integration\" ./tests/e2e -run \"^TestHPCMarketplaceE2E$\" -count=1`, `tests/e2e/hpc_marketplace_e2e_test.go` | Command exits `0`; provider onboarding and marketplace flow are healthy. |
| `T+45` | `FIN`, `OPS` | Run settlement and payout launch smoke tests and capture finance sign-off. | `PREREQ-BILL-REC`, `LAUNCH-DR-001`, `LAUNCH-DEC-001` | `go test -tags=\"e2e.integration\" ./tests/integration/settlement -run \"TestFullPipelineTestSuite|TestSettlementPayoutOffRampFlow|TestDisputeArbitrationRefundFlow\" -count=1`, `_docs/runbooks/finance-reconciliation-runbook.md`, `tests/integration/settlement/full_pipeline_test.go`, `tests/integration/settlement/payout_flow_test.go` | Finance confirms reconciliation and payout flows are healthy. |
| `T+60` | `RM`, `OPS`, `PROD` | Run the one-hour checkpoint, review all evidence IDs, and either continue the launch window or declare rollback. | `LAUNCH-DEC-001`, `LAUNCH-DR-001`, `COMMS-PLAN` | `_docs/operations/mainnet-go-no-go-decision.md`, `docs/sre/COMMUNICATION_TEMPLATES.md` | No P0 rollback trigger is active and customer-facing services are stable. |
| `T+75` | `PROD` | Publish the scheduled 75-minute update and confirm the next update time. | `COMMS-PLAN`, `LAUNCH-DR-001` | `docs/sre/COMMUNICATION_TEMPLATES.md` | External and internal updates remain on the 15-minute cadence. |
| `T+90` | `OPS`, `PLAT` | Validate DR posture remains intact after launch traffic begins. | `DR-RUNBOOK`, `BACKUP-RESTORE`, `LAUNCH-DR-001` | `_docs/disaster-recovery.md`, `docs/operations/runbooks/BACKUP_RESTORE.md`, `scripts/dr/backup-chain-state.sh` | Backup posture is intact and no restore action is required. |
| `T+105` | `RM`, `SEC`, `COMP`, `FIN`, `PROD` | Final launch-window quorum review before handoff to the stability window. | `LAUNCH-DEC-001`, `LAUNCH-CHK-001` | `_docs/operations/mainnet-go-no-go-decision.md`, `_docs/operations/mainnet-launch-readiness-checklist.md` | All role owners agree the launch can move to steady-state monitoring. |
| `T+120` | `RM`, `SCRIBE` | Close the active control window and enter 24-hour stability watch. | `LAUNCH-DEC-001`, `LAUNCH-DR-001`, `COMMS-PLAN` | `_docs/operations/mainnet-go-no-go-decision.md`, `_docs/operations/mainnet-dress-rehearsal-report.md` | `GO` remains in force, the stability watch owner is recorded, and the next update time is documented. |

## Escalation Rules

| Severity | Trigger | Immediate owner | Escalation path | Required evidence updates |
| --- | --- | --- | --- | --- |
| `SEV-1` | consensus halt, validator participation below threshold for more than 10 minutes, genesis-hash mismatch, confirmed security compromise, or unrecoverable settlement corruption | `OPS` | `RM`, `SEC`, `PLAT`, `PROD` join war room immediately; status page update within 15 minutes; rollback decision within 10 minutes | `LAUNCH-DEC-001`, `LAUNCH-DR-001`, `COMMS-PLAN`, plus the owning prerequisite ID |
| `SEV-2` | customer-facing launch regression while chain remains live, provider onboarding failure, VEID smoke-test failure, backup/restore smoke-test failure | `OPS` | `RM` and owning role join within 5 minutes; decide hold vs rollback within 15 minutes | `LAUNCH-DEC-001`, `LAUNCH-DR-001`, and the failing prerequisite ID |
| `SEV-3` | documentation mismatch, non-customer-facing tooling warning, comms delay without service impact | `RM` | owning role corrects within the current cadence window; do not advance to the next phase until resolved | `LAUNCH-DEC-001` and affected ID |

## Rollback Triggers
- `scripts/mainnet/prelaunch-checklist.sh` returns non-zero at `T-25`
- `scripts/mainnet/genesis-validate.sh` fails or `scripts/mainnet/genesis-hash.sh`
  disagrees with the approved launch hash
- no healthy block production is observed by `T+05`
- validator participation stays below 67 percent voting power for more than
  10 minutes
- sustained block time exceeds 2x target for more than 20 minutes
- VEID verification error rate exceeds 2 percent over 15 minutes
- settlement reconciliation fails or finance cannot confirm payout integrity
- any required launch smoke-test command exits non-zero with customer-facing
  impact
- `SEC` declares an active security event that invalidates `PREREQ-SEC-ML`

## Minute-by-Minute Rollback Playbook

| Rollback minute | Lead | Action | A2 evidence IDs | Real operational artifact or command | Exit rule |
| --- | --- | --- | --- | --- | --- |
| `R+00` | `RM`, `OPS`, `SCRIBE` | Declare `ROLLBACK`, freeze forward changes, timestamp the trigger, and post the first internal incident update. | `LAUNCH-DEC-001`, `LAUNCH-DR-001`, `COMMS-PLAN` | `docs/sre/COMMUNICATION_TEMPLATES.md`, `docs/operations/runbooks/INCIDENT_RESPONSE.md` | Rollback authority is explicit and all launch changes are frozen. |
| `R+05` | `OPS`, `PLAT` | Classify the failure domain and select exactly one rollback path: deployment, traffic switch, or infrastructure state. | `DR-RUNBOOK`, `BACKUP-RESTORE`, `LAUNCH-DEC-001` | `scripts/rollback/argocd-rollback.sh`, `scripts/rollback/blue-green-switch.sh`, `scripts/rollback/terraform-rollback.sh`, `_docs/disaster-recovery.md` | Correct rollback path is chosen and recorded before execution. |
| `R+10` | `PLAT` | Execute the selected rollback path and capture the exact command transcript. | `LAUNCH-DR-001`, `BACKUP-RESTORE`, `DR-RUNBOOK` | `scripts/rollback/argocd-rollback.sh`, `scripts/rollback/blue-green-switch.sh`, `scripts/rollback/terraform-rollback.sh` | Rollback command exits `0`; if it fails, advance immediately to restore planning. |
| `R+15` | `OPS`, `PLAT` | Validate chain health and, if needed, prepare or execute state restore using the DR procedures. | `BACKUP-RESTORE`, `DR-RUNBOOK`, `LAUNCH-DR-001` | `docs/operations/runbooks/BACKUP_RESTORE.md`, `scripts/dr/backup-chain-state.sh`, `_docs/disaster-recovery.md` | Either the service is healthy after rollback or a restore path is started with evidence captured. |
| `R+20` | `PROD` | Publish rollback status-page update and partner notice, then confirm the next update time. | `COMMS-PLAN`, `LAUNCH-DEC-001` | `docs/sre/COMMUNICATION_TEMPLATES.md`, `docs/sre/INCIDENT_RESPONSE.md` | External update is live within the 15-minute cadence. |
| `R+25` | `VREL`, `FIN`, `SEC` | Notify validators, providers, finance, and security owners of the rollback state and any follow-up checks they own. | `LAUNCH-DR-001`, `PREREQ-PROVIDER-E2E`, `PREREQ-BILL-REC`, `PREREQ-SEC-ML` | `_docs/runbooks/validator-onboarding.md`, `_docs/runbooks/finance-reconciliation-runbook.md`, `_docs/security-review-8d-ml-verification.md` | All owner groups acknowledge the rollback state and start their post-rollback checks. |
| `R+30` | `RM`, `OPS` | Decide whether the system is stable enough for observation or whether the incident converts into a full DR event. | `LAUNCH-DEC-001`, `DR-RUNBOOK`, `LAUNCH-DR-001` | `_docs/disaster-recovery.md`, `docs/operations/runbooks/INCIDENT_RESPONSE.md` | Stable rollback enters incident monitoring; unstable rollback enters DR execution. |

## Evidence Capture Instructions
- `SCRIBE` records every control action with UTC timestamp, acting role,
  command line, exit code, and destination evidence ID.
- Every launch-day command transcript is copied verbatim into the evidence file
  named by its A2 ID before the next timed step begins.
- External status-page messages, Slack incident posts, and partner notices are
  copied into `COMMS-PLAN` with the exact published UTC timestamp.
- VEID smoke-test evidence goes only to `PREREQ-VEID-E2E`.
- Provider and marketplace smoke-test evidence goes only to
  `PREREQ-PROVIDER-E2E`.
- Finance settlement and payout evidence goes only to `PREREQ-BILL-REC`.
- DR and backup evidence goes only to `DR-RUNBOOK` and `BACKUP-RESTORE`.
- `RM` updates `LAUNCH-DEC-001` at every control decision: `GO`, `HOLD`,
  `NO-GO`, `ROLLBACK`, and control-window close.
- At control-window close, compute and record the current runbook hash so
  `LAUNCH-RUNBOOK` can be refreshed in the launch packet:

```bash
sha256sum _docs/runbooks/mainnet-launch-runbook.md
```

## Exit Criteria
- `GO` path: `T+120` reached with no active rollback trigger, all timed
  checkpoints satisfied, and all launch-day evidence stored under the A2 IDs
  listed in this runbook.
- `NO-GO` path: any pre-`GO` hard gate fails and `RM` records the stop in
  `LAUNCH-DEC-001`.
- `ROLLBACK` path: any rollback trigger fires after `T0` and the rollback table
  above is executed without skipping evidence capture.
