# Mainnet Backup Restore Drill Report - 2026-04-11

Last updated: 2026-04-11
Owner: Ops Lead

## Scope
Repository-backed launch rehearsal evidence for backup, restore, and rollback
readiness used by the mainnet dress rehearsal bundle.

## Command

```bash
./scripts/ci/backup-restore-smoke-test.sh
```

## Evidence artifact
- Raw log: `output/mainnet-launch/2026-04-11/backup-restore-smoke.log`
- Completed (UTC): 2026-04-11 05:48:10
- SHA-256: `f14d808dabfae137ea9369bd21698ab5165f918584530191d76b4fdf2631181a`

## Result
- Status: `PASS`
- Recorded output: `backup/restore smoke test passed`

## Rehearsal interpretation
- The repository-backed restore path executed successfully for rehearsal scope.
- `_docs/disaster-recovery.md` remains the source of truth for production RPO
  and RTO policy.
- The launch checklist now treats this artifact as completion of the restore
  smoke and DR validation path required for the rehearsal bundle.

## Source references
- `_docs/disaster-recovery.md`
- `docs/operations/runbooks/BACKUP_RESTORE.md`
- `scripts/ci/backup-restore-smoke-test.sh`
- `output/mainnet-launch/2026-04-11/backup-restore-smoke.log`
