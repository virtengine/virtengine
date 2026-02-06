# Mainnet Launch Runbook

Last updated: 2026-02-06
Owner: Operations

## Purpose
Operational runbook for executing the mainnet launch, including readiness checks, cutover, and rollback procedures.

## References
- Readiness checklist: _docs/operations/mainnet-launch-readiness-checklist.md
- Dress rehearsal report: _docs/operations/mainnet-dress-rehearsal-report.md
- Go/No-Go decision record: _docs/operations/mainnet-go-no-go-decision.md
- Launch evidence packet: _docs/operations/mainnet-launch-packet.md
- DR and continuity: _docs/disaster-recovery.md, _docs/business-continuity.md

## Roles and contact points
- Release Manager: <name>
- Ops Lead: <name>
- Security Lead: <name>
- Compliance Lead: <name>
- Finance Lead: <name>
- Product Lead: <name>

## Pre-launch checklist (T-7 days to T-0)
- [ ] Confirm prerequisite tasks complete (E2E, security, billing)
- [ ] Review readiness checklist and sign-off status
- [ ] Validate backups and restore drill results
- [ ] Confirm validator set and chain config
- [ ] Prepare comms plan and status page posts

## Launch day steps
1. Validate monitoring dashboards and alerting
2. Verify genesis file checksums and config signatures
3. Start chain and confirm validator participation
4. Execute VEID onboarding and verification smoke tests
5. Execute provider marketplace smoke tests
6. Validate billing reconciliation snapshot
7. Declare GO/NO-GO after stability window

## Rollback procedure
- Trigger rollback when rollback criteria in readiness checklist are met
- Execute rollback runbook steps (timeboxed)
- Communicate status updates at 15-min cadence
- Record all evidence in launch packet

## Post-launch
- Monitor stability window (minimum 24h)
- Capture final metrics and SLO compliance
- Close out go/no-go record and publish retrospective

