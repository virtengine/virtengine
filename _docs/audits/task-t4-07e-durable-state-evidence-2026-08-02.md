# T4-07E Durable-State Remediation Evidence - 2026-08-02

## Status

`durable-state` remains **unverified**. T4-07E does not complete Task 88C and does not retire `_build/helm/slurm-cluster`.

Helm and kubectl are unavailable in the evidence environment. No Kubernetes cluster, CSI driver, StorageClass, PV reclaim policy, VolumeSnapshot implementation, or live backup destination was exercised. The evidence below is source, schema, rendered-fixture, and offline-drill evidence only.

## Remediation

- Controller `/var/spool/slurm`, slurmdbd `/var/spool/slurm`, and MariaDB `/var/lib/mysql` remain PVC-backed in their primary containers.
- Every component supports either a generated `volumeClaimTemplates` claim or an explicit `persistence.existingClaim`; the two forms are mutually exclusive in each template.
- Existing claims require exactly one replica. HA controller and slurmdbd configurations use generated per-ordinal claims; `persistence.accessMode` does not claim shared-state safety for an existing PVC.
- Generated claims declare `persistentVolumeClaimRetentionPolicy.whenDeleted=Retain` and `whenScaled=Retain`.
- Static and rendered checks reject an authoritative path backed by `emptyDir` and require generated claims to retain on delete and scale.
- The runbook requires independent StorageClass/PV reclaim review and documents complete backup inputs, SHA-256 manifests, isolated logical MariaDB import into a new PVC, staged startup order, and whole-claim-set rollback.
- The executable offline drill validates all archive checksums and members before creating fresh staging, restricts every archive to its component root, rejects links/special entries and unsafe manifest names, preserves pre-existing staging/targets, and atomically promotes only a complete restore.
- Backup code explicitly enforces bundle mode `0700` and archive/manifest mode `0600` independent of umask.

## Executed Evidence

| Command | Result | Scope |
| --- | --- | --- |
| `python scripts/validate_slurm_chart_semantics_test.py -v` | PASS, 53 tests | Static, schema, rendered fixtures, and existing-claim replica safety |
| `python scripts/hpc/slurm-durable-state-drill-test.py -v` | PASS, 10 tests; 2 POSIX/symlink source checks skipped on Windows | Round trip, checksum tamper, traversal, cross-component entries, links/special files, modes, and pre-existing staging/target preservation |
| `python scripts/hpc/slurm-durable-state-drill.py` | PASS | Offline fixture only |
| `node scripts/validate-slurm-chart-inventory.cjs` | PASS, blocked executable-semantic | Inventory remains conservative |
| `node --test scripts/validate-slurm-chart-inventory.test.cjs` | PASS, 11 internal scenarios | Inventory controls |
| `go test -tags=e2e.integration -count=1 ./tests/integration/hpc -run TestDurableStateOfflineContracts` | PASS | Focused tagged Go source contract |
| `git diff --check` | PASS | Worktree whitespace integrity |

## Remaining Live Gates

- Render and validate both generated-claim and `existingClaim` configurations with Helm.
- Verify PVC retention through StatefulSet deletion and scale operations against the selected Kubernetes version.
- Record StorageClass and bound PV reclaim policy, encryption, snapshot/export, and deletion-protection behavior.
- Exercise controller, slurmdbd, and MariaDB restart/failover with state continuity.
- Perform an isolated backup and fail-closed restore using approved storage and backup systems, verify checksums from the final destination, and validate `mariadb-check`, slurmdbd connectivity, `sacct`, and `scontrol ping`.

Until those gates are executed, the semantic diagnostic and inventory must continue to report `durable-state: unverified` and overall completion blocked.