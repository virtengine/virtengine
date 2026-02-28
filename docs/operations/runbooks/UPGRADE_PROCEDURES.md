# Upgrade Procedures

**Owner:** Release Management + SRE
**Last Updated:** 2026-04-11

This runbook covers chain, provider, VEID, TEE, and infrastructure rollouts
using the current repository automation. It is written to match the repository's
prelaunch and staged-rollout posture: upgrades are executable, but production
promotion still requires the checked-in verification and launch evidence.

## Upgrade Sources of Truth

Before any production-facing rollout, review:

- `RELEASE.md`
- `VERIFICATION.md`
- `docs/COMPATIBILITY.md`
- `_docs/operations/mainnet-go-no-go-decision.md`
- `docs/operations/DEPLOYMENT_GUIDE.md`
- `infra/scripts/check-environment-parity.sh`
- `infra/scripts/terraform-run.sh`

## Upgrade Classes

| Upgrade class | Primary trigger | Primary rollback |
| --- | --- | --- |
| Chain binary upgrade | approved release tag or governance-driven upgrade | binary rollback plus `virtengine rollback --hard` if needed |
| Provider-daemon upgrade | provider rollout or bug fix | `scripts/rollback/argocd-rollback.sh` or `blue-green-switch.sh` |
| VEID model upgrade | approved manifest and matching governance state | revert to previous verified model bundle |
| TEE platform or attestation update | TCB or attestation requirement | TEE platform failover plus previous attestation material |
| Infra change | reviewed Terraform plan | `scripts/rollback/terraform-rollback.sh` |

## Preflight Gate

Do not begin rollout until all applicable checks pass:

```bash
./infra/scripts/check-environment-parity.sh
./scripts/dr/dr-test.sh --environment staging --report
./scripts/rollback/argocd-rollback.sh virtengine-platform-staging --dry-run
./scripts/rollback/terraform-rollback.sh staging --steps-back 1 --dry-run
```

Also confirm the relevant CI lanes are green:

- `quality-gate`
- `compatibility`
- `security`
- `supply-chain`
- `staging-e2e`
- `smoke-test`

## Chain Binary Upgrade

### Prepare the candidate

```bash
RELEASE_TAG="${RELEASE_TAG:?set approved release tag}"
./script/semver.sh validate "$RELEASE_TAG"
./script/is_prerelease.sh "$RELEASE_TAG"
./script/mainnet-from-tag.sh "$RELEASE_TAG"
```

### Backup before rollout

```bash
./scripts/dr/backup-chain-state.sh
./scripts/dr/backup-keys.sh --type all
```

### Upgrade

```bash
curl -L "https://github.com/virtengine/virtengine/releases/download/${RELEASE_TAG}/virtengine_linux_amd64.tar.gz" -o /tmp/virtengine.tar.gz
tar -xzf /tmp/virtengine.tar.gz -C /tmp
sudo install -m 0755 /tmp/virtengine /usr/local/bin/virtengine
virtengine version
sudo systemctl restart virtengine
```

### Post-upgrade watchlist

Watch for:

- `ConsensusStalled`
- `ValidatorDown`
- `ChainHalted`
- `BlockProductionStalled`
- `LowVotingPowerParticipation`
- `SLOTxConfirmationLatencyBreach`

If any Tier 0 consensus alert persists beyond one review window, move to the
rollback section.

## Provider Rollout

Use Argo or blue/green tooling rather than manual ad hoc pod edits.

Dry-run rollback first:

```bash
./scripts/rollback/argocd-rollback.sh virtengine-platform-prod --dry-run
PROMETHEUS_URL="${PROMETHEUS_URL:?set prometheus url}" \
./scripts/rollback/blue-green-switch.sh provider-daemon green --mode gradual --yes
```

Watch for:

- `ProviderHighFailureRate`
- `ProviderDaemonScalingFailure`
- `ProviderDaemonHighPendingOrders`
- `ProviderDaemonClaimConflictsHigh`
- `ProviderDaemonBidLatencyHigh`
- `MarketplaceUnavailable`

If pending orders, bid latency, or claim conflicts regress during rollout,
pause promotion and revert to the previous healthy revision.

## VEID Model Upgrade

Preconditions:

- approved model manifest exists,
- sidecars verify the manifest successfully,
- governed model state matches the bundle,
- the previous verified model remains available for rollback.

Checks:

```bash
virtengine query veid model-status
curl -s http://localhost:26660/metrics | grep -E "veid_|inference_"
```

Abort immediately on:

- `VEIDModelVersionMismatch`
- `VEIDModelNotLoaded`
- `VEIDNonDeterministicInference`
- `VEIDInferenceNonDeterministic`

Rollback rule:

1. revert to the previous verified model manifest and bundle,
2. restart the affected scoring or validator processes,
3. confirm the on-chain or governed expected version again before resuming.

## TEE and Attestation Upgrade

Before changing enclave material:

```bash
./deploy/tee/validate.sh
./deploy/tee/failure-recovery-drill.sh --dry-run sgx
./deploy/tee/failure-recovery-drill.sh --dry-run sev
```

Watch for:

- `TEEEnclaveReplicasUnavailable`
- `TEEEnclaveSignatureFailures`
- `TEEEnclaveAttestationFailures`
- `TEEEnclaveStaleAttestations`
- `PrimaryTEEPlatformDown`
- `TEEFailoverFailed`

If signatures or attestations fail, stop the rollout and restore the previous
attestation or enclave material before continuing.

## Infrastructure Upgrade

Use the reviewed plan wrapper only:

```bash
./infra/scripts/terraform-run.sh plan infra/terraform/environments/staging output/infra/staging-plan
./infra/scripts/terraform-run.sh apply infra/terraform/environments/staging output/infra/staging-plan
```

For production, confirm the plan artifact and checksum were reviewed before
apply.

Watch for:

- `RegionUnhealthy`
- `CrossRegionLatencyHigh`
- `ClusterCapacityInsufficient`
- `ClusterNodePressure`
- `CockroachDBNodeDown`
- `CockroachDBHighReplicationLag`

## Rollback Decision Rules

Roll back immediately when:

1. a Tier 0 consensus alert persists after targeted remediation,
2. error budget burn turns `critical` during rollout,
3. attestation or model determinism fails,
4. provider rollout breaks order fulfillment materially,
5. regional or database health breaches the DR boundary.

## Rollback Paths

### ArgoCD revision rollback

```bash
./scripts/rollback/argocd-rollback.sh virtengine-platform-prod --yes
```

### Blue or green traffic rollback

```bash
PROMETHEUS_URL="${PROMETHEUS_URL:?set prometheus url}" \
./scripts/rollback/blue-green-switch.sh provider-daemon blue --mode instant --yes
```

### Terraform rollback

```bash
./scripts/rollback/terraform-rollback.sh prod --steps-back 1 --yes
```

### Chain binary rollback

```bash
sudo systemctl stop virtengine
sudo cp /usr/local/bin/virtengine.bak /usr/local/bin/virtengine
virtengine rollback --hard
sudo systemctl start virtengine
```

## Exit Criteria

Do not mark the rollout complete until:

1. the target service version is confirmed,
2. the rollout watchlist is clear,
3. rollback commands still succeed in dry-run mode,
4. any generated drill or rollback artifacts are attached to the release
   evidence.
