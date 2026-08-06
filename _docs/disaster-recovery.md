# VirtEngine Disaster Recovery Plan

**Owner:** SRE + Infrastructure
**Last Updated:** 2026-04-11

This document is the disaster-recovery control sheet for the repo-owned
platform. It aligns the A18 infrastructure model, the B18 recovery scripts, and
the live regional, TEE, and database alerts.

## Scope

This plan covers:

- validator and full-node recovery,
- provider-daemon state recovery,
- multi-region traffic and platform failover,
- TEE attestation and hardware failover,
- CockroachDB replication and restore,
- evidence-producing rehearsal and live drills.

## Recovery Targets

| Surface | RTO target | RPO target | Primary automation |
| --- | --- | --- | --- |
| Validator node | `15m` for single-node recovery | zero chain-state loss beyond the selected restore point | `scripts/dr/backup-chain-state.sh` |
| Provider daemon | `30m` | latest verified provider backup | `scripts/dr/backup-provider-state.sh` |
| Regional failover | `30m` rehearsal target | no more than the verified replication lag window | `infra/dr/run-failover-drill.sh` |
| TEE signing and attestation | `15m` to restore healthy signing path | zero unauthorized attestation acceptance | `deploy/tee/failure-recovery-drill.sh` |
| CockroachDB replicated state | `30m` for single-node loss, `60m` for cluster repair | replication lag below `300s` | platform restore plus Cockroach node repair |

## Entry Conditions

Start the DR process when any of the following fire or an equivalent manual
incident is declared:

- `RegionUnhealthy`
- `CrossRegionLatencyHigh`
- `StateSyncProviderDown`
- `StateSyncSnapshotStale`
- `PrimaryTEEPlatformDown`
- `TEEFailoverFailed`
- `AllEnclavesDown`
- `CockroachDBNodeDown`
- `CockroachDBHighReplicationLag`
- `CockroachDBLowDiskSpace`
- `CockroachDBHighMemoryUsage`
- a confirmed chain halt that cannot be resolved through standard validator
  recovery.

## Canonical Commands

### Backup and readiness checks

```bash
./scripts/dr/backup-chain-state.sh
./scripts/dr/backup-provider-state.sh
./scripts/dr/backup-keys.sh --type all
./scripts/dr/dr-test.sh --environment staging --report
```

### Drill execution

```bash
./infra/dr/run-failover-drill.sh --mode rehearsal --output-dir output/drill/rehearsal
./infra/dr/run-failover-drill.sh --mode live --live-validation --output-dir output/drill/live
```

### Rollback and traffic recovery

```bash
./scripts/rollback/argocd-rollback.sh virtengine-platform-prod --dry-run
./scripts/rollback/blue-green-switch.sh provider-daemon green --mode gradual --yes
./scripts/rollback/terraform-rollback.sh prod --steps-back 1 --dry-run
```

## Standard DR Workflow

1. Declare incident severity and freeze unrelated production changes.
2. Capture current evidence:
   - active alerts,
   - Grafana and Alertmanager URLs,
   - `kubectl get pods -A`,
   - node and provider logs,
   - latest backup metadata.
3. Run the narrowest recovery action that restores service safely.
4. If the event crosses a regional or attestation boundary, move to the
   matching scenario below.
5. Keep all generated artifacts under `output/` or the script default paths.

## Scenario Playbooks

### Single Node Recovery

Use this for `NodeDown`, `NodeOutOfSync`, `NodeBehind`, or a failed validator
pod that does not require cross-region work.

1. Confirm the node-specific failure from monitoring and logs.
2. Verify backup freshness:
   ```bash
   ./scripts/dr/dr-test.sh --environment staging --test backup
   ```
3. Restore state if needed:
   ```bash
   ./scripts/dr/backup-chain-state.sh --restore latest
   ```
4. If the host is a provider node, verify provider backup integrity:
   ```bash
   ./scripts/dr/backup-provider-state.sh --verify
   ```
5. Rejoin and verify peer count, block height, and readiness.

### Region Failover

Use this for `RegionUnhealthy`, `CrossRegionLatencyHigh`, or a cloud-provider
outage that affects the active region.

1. Confirm the region impact from cloud status and `up{job=~"virtengine-.*"}`.
2. Stop manual deploys and set incident command.
3. Run the region failover drill wrapper in rehearsal mode first if time allows:
   ```bash
   ./infra/dr/run-failover-drill.sh --mode rehearsal --output-dir output/drill/rehearsal
   ```
4. If a live failover is required, shift traffic and record the evidence bundle:
   ```bash
   ./infra/dr/run-failover-drill.sh --mode live --live-validation --output-dir output/drill/live
   ```
5. If Argo or Terraform state must be reverted, use the repo rollback wrappers
   instead of ad hoc cloud-console changes.

### TEE Failover and Attestation Recovery

Use this for:

- `TEEEnclaveReplicasUnavailable`
- `TEEEnclaveSignatureFailures`
- `TEEEnclaveAttestationFailures`
- `TEEEnclaveStaleAttestations`
- `AttestationVerificationFailed`
- `PrimaryTEEPlatformDown`
- `TEEFailoverFailed`

1. Check enclave health and replica status:
   ```bash
   kubectl -n virtengine get deploy,pods -l app.kubernetes.io/name=tee-enclave
   ```
2. Review the TEE-specific operational controls:
   - `deploy/tee/validate.sh`
   - `deploy/tee/failure-recovery-drill.sh`
3. Run the dry-run recovery drill first unless signatures are actively failing:
   ```bash
   ./deploy/tee/failure-recovery-drill.sh --dry-run sgx
   ```
4. If the primary platform is unavailable, promote the healthy platform per the
   deployment and verification runbooks, then confirm attestation freshness and
   signature recovery.
5. If an unknown measurement or debug enclave is detected, quarantine the
   affected deployment immediately and treat it as a security incident as well
   as an ops incident.

### CockroachDB and Replicated State Recovery

Use this for:

- `CockroachDBNodeDown`
- `CockroachDBHighReplicationLag`
- `CockroachDBLowDiskSpace`
- `CockroachDBHighMemoryUsage`

1. Confirm the failure from Cockroach metrics and Kubernetes events.
2. Check whether replication lag is below or above the `300s` DR target.
3. If the node is unavailable but quorum remains healthy, replace or restart
   the failed node without regional cutover.
4. If disk or memory pressure is the cause, stabilize the node class before
   rebalancing traffic.
5. If lag breaches the target and coincides with regional degradation, move to
   region failover and keep the replication-lag evidence in the drill bundle.

### Backup Restore

Use this only when the service cannot be recovered in-place.

1. Select the restore point from verified backup metadata.
2. Restore chain state:
   ```bash
   ./scripts/dr/backup-chain-state.sh --restore latest
   ```
3. Restore provider state if the incident is off-chain:
   ```bash
   ./scripts/dr/backup-provider-state.sh --restore latest
   ```
4. Restore encrypted key material only with the required passphrase or shares:
   ```bash
   ./scripts/dr/backup-keys.sh --restore BUNDLE_NAME
   ```
5. Re-run `dr-test.sh --report` after recovery.

## Drill Cadence

| Drill | Minimum cadence | Evidence required |
| --- | --- | --- |
| Backup verification | daily | `dr-test.sh` output or JSON report |
| Rehearsal failover drill | weekly | `failover-drill.log`, summary, evidence JSON |
| Live failover validation | before launch window and quarterly after launch | evidence bundle under `output/drill/live` |
| TEE recovery drill | weekly dry-run and before production changes | `failure-recovery-drill.sh` output |
| Rollback dry-runs | before each production rollout | script output plus selected target revision |

## Evidence Requirements

For any real or rehearsal DR event, keep:

- the script command line,
- the output directory path,
- SHA or version of the code under test,
- affected regions and services,
- start and end timestamps,
- final service state,
- any rollback decision taken.

## Exit Criteria

Do not close the DR event until:

1. the triggering alerts have cleared or are knowingly suppressed,
2. the recovered surface passes its readiness check,
3. the latest drill or restore artifacts are attached,
4. the owner has assigned follow-up work for any residual risk.
