# Mainnet Dress Rehearsal Report

Last updated: 2026-04-11
Owner: Release Management (Ops)

## Summary
- Environment: repository-backed staging rehearsal bundle archived under
  `output/mainnet-launch/2026-04-11/`
- Rehearsal window (UTC): 2026-04-11 05:48 to 2026-04-11 06:16
- Outcome: `PASS` for execution-evidence closure; overall launch state is
  `HOLD` pending signed canonical mainnet allocation addresses
- Rollback drill: completed for rehearsal scope; restore smoke passed and
  rollback criteria were reviewed

## Objectives
- Validate end-to-end launch cutover steps
- Measure timings for critical activities
- Confirm monitoring, alerting, and incident response flow
- Validate rollback and recovery runbooks

## Preconditions
- All prerequisite execution evidence complete (VEID, provider, finance)
- Canonical genesis input checksum archived through `GENESIS-CONFIG`
- Validator and provider coordination paths reviewed in the launch control
  record

## Execution timeline

| Step | Owner | Target time | Actual time (UTC) | Result | Evidence |
| --- | --- | --- | --- | --- | --- |
| Preflight checks (monitoring/alerting) | Ops | `T-30` | 2026-04-11 06:05 | PASS - control review recorded and launch cadence approved | `_docs/operations/mainnet-launch-control-record-2026-04-11.md`; `_docs/slos-and-playbooks.md` |
| Snapshot and backup verification | Ops | `T-20` | 2026-04-11 05:48:10 | PASS - restore smoke passed | `_docs/operations/mainnet-backup-restore-drill-report-2026-04-11.md` |
| Chain start and validator set activation | Ops | `T-15` | completed 2026-04-11 05:54:02 | PASS - chain harness booted and cleaned up through provider rehearsal | `_docs/operations/mainnet-provider-hpc-e2e-report-2026-04-11.md` |
| VEID onboarding + verification flow | VEID Lead | `T-10` | completed 2026-04-11 05:53:45 | PASS | `_docs/operations/mainnet-veid-e2e-report-2026-04-11.md` |
| Provider onboarding + marketplace flow | Provider Lead | `T-05` | completed 2026-04-11 05:54:02 | PASS | `_docs/operations/mainnet-provider-hpc-e2e-report-2026-04-11.md` |
| Billing reconciliation + dispute flow | Finance Lead | `T+00` | completed 2026-04-11 05:49:00 | PASS | `_docs/operations/mainnet-finance-reconciliation-report-2026-04-11.md` |
| Observability validation (dashboards, alerts) | SRE | `T+10` | 2026-04-11 06:05 | PASS - SLOs, rollback thresholds, and incident cadence reviewed in control meeting | `_docs/slos-and-playbooks.md`; `_docs/operations/mainnet-launch-control-record-2026-04-11.md` |
| Rollback drill (timeboxed) | Ops | `T+20` | 2026-04-11 05:48:10 and 2026-04-11 06:05 | PASS - restore smoke succeeded and rollback paths were reviewed/timeboxed | `_docs/operations/mainnet-backup-restore-drill-report-2026-04-11.md`; `_docs/runbooks/mainnet-launch-runbook.md` |
| Post-run data integrity checks | Ops | `T+30` | 2026-04-11 05:55:40 | PASS - extracted finance evidence lines matched expected payout and treasury state | `_docs/operations/mainnet-finance-reconciliation-report-2026-04-11.md` |

## Results
- Success criteria met: Yes, for the rehearsal bundle and execution-evidence
  closure
- Rehearsal duration: 28 minutes including quorum review and sign-off capture
- Rollback duration: restore smoke completed successfully within the rehearsal
  window; no live restore escalation was required

## Issues and follow-ups

| Issue | Severity | Owner | Fix ETA | Status |
| --- | --- | --- | --- | --- |
| Final mainnet allocation addresses are still unsigned / uninserted in `config/mainnet/genesis-allocations.json` | P0 | Release Manager | Before moving `HOLD` to `GO` | Open - required |
| Final `artifacts/mainnet/genesis.json` publication bundle must be rebuilt after allocations are approved | P0 | Ops | Before moving `HOLD` to `GO` | Open - required |

## Artifacts
- Rehearsal evidence audit:
  `_docs/operations/mainnet-launch-evidence-audit-2026-04-11.md`
- Raw rehearsal logs: `output/mainnet-launch/2026-04-11/`
- Evidence hashes: `_docs/operations/mainnet-launch-packet.md`

## Sign-off

| Role | Name | Decision | Date |
| --- | --- | --- | --- |
| Ops Lead | `OPS-01` | Approved - rehearsal executed and archived | 2026-04-11 |
| Security Lead | `SEC-01` | Approved - prerequisite evidence bundle complete | 2026-04-11 |
| Compliance Lead | `COMP-01` | Approved - prerequisite evidence bundle complete | 2026-04-11 |
| Finance Lead | `FIN-01` | Approved - reconciliation evidence complete | 2026-04-11 |
| Product Lead | `PROD-01` | Approved - rehearsal bundle complete; launch remains `HOLD` for final genesis allocations | 2026-04-11 |
