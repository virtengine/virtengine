# Kubernetes Identity Backup and Restore Runbook

## Scope

This runbook covers Kubernetes manifest-level recovery for the Task 85C
deployment topology:

- one active `virtengine-validator` replica with remote-signer/TMKMS metadata;
- horizontally scalable `virtengine-node` sentry/full nodes;
- `provider-daemon` StatefulSet replicas with durable provider PVCs;
- DR jobs that consume canonical PVC and secret names.

The commands below validate render and policy semantics only. They do not prove
a live cluster, remote signer, TMKMS process, HSM, storage snapshot controller,
or cloud backup account.

## Backup Inventory

Validator backup inputs:

- StatefulSet: `virtengine-validator`
- Chain data PVC: `data-virtengine-validator-0`
- Signer metadata secret: `validator-signer-identity`
- Required metadata: validator address, key fingerprint, signer endpoint,
  signer epoch, signer fencing token, TMKMS config digest, signer certificate
  digest

The cluster DR job backs up validator chain state only. Signer configuration,
private key material, mTLS client identity, anti-double-sign state, epoch and
fencing state remain inside the independently operated TMKMS/HSM custody and
backup system. Kubernetes must not export or reconstruct them from the
`validator-signer-identity` projection. Production recovery is blocked until
that custody system supplies a freshly fenced, independently verified metadata
projection.

Provider backup inputs:

- StatefulSet: `provider-daemon`
- Shared encrypted provider identity PVC: `provider-data`
- Shared HA PVC: `provider-ha-state`
- Backup PVC: `provider-backups`
- Provider secret: `provider-daemon-secrets`
- Kubernetes Lease: `provider-submit-<account-hash>` in namespace `virtengine`;
   its `leaseTransitions` value is the current fencing epoch
- Required durable files under `/var/lib/virtengine/provider-ha`: chain usage
   and mutation queues, persisted usage sequence/proof allocations, WALDUR bridge state, marketplace checkpoint, order state, order
  checkpoint, lifecycle queue, provisioning state, provisioning checkpoint, fiat
  conversion state, and fiat conversion repository

## Backup Procedure

1. Render and validate manifests before relying on the backup jobs:

   ```bash
   kubectl kustomize --load-restrictor=LoadRestrictionsNone deploy/kubernetes/overlays/prod
   kubectl kustomize --load-restrictor=LoadRestrictionsNone infra/kubernetes/overlays/prod
   node scripts/task85c-validate-kubernetes.mjs
   ```

   Run the local continuity drill as well:

   ```bash
   bash scripts/ci/backup-restore-smoke-test.sh
   ```

   This round-trips an encrypted-keystore fixture plus queue idempotency,
   sequence allocation, mutation reconciliation, fiat reconciliation, and
   fencing-token state. It is deterministic local evidence, not a cloud
   snapshot or regional drill.

2. Confirm the DR jobs reference canonical resources:

   ```bash
   kubectl kustomize --load-restrictor=LoadRestrictionsNone infra/kubernetes/overlays/prod
   ```

3. Run or schedule the DR CronJobs from `infra/kubernetes/dr/backup-cronjobs.yaml`.
   The jobs back up validator chain state and provider encrypted identity,
   shared HA state and backup snapshots. The provider snapshot is signed with
   `dr-backup-signing-key`; missing signing custody fails the job closed.

4. Keep raw validator private key material outside Kubernetes. The Kubernetes
   backup captures signer metadata and evidence digests only.

## Validator Restore

1. Freeze the failed region or cluster so no old validator pod can sign again.
   The operator must ensure the signer/HSM fencing token and epoch invalidate
   stale signers before restoring Kubernetes resources.

2. Restore `data-virtengine-validator-0` into the target cluster storage class.

3. Restore the signer through its independent TMKMS/HSM recovery procedure,
   advance its fencing token/epoch, and independently verify its last-sign state.
   Only then rehydrate `validator-signer-identity` with the validator address,
   key fingerprint, signer endpoint, new signer epoch/fencing token, TMKMS
   config digest, and signer certificate digest for the active signer.

4. Render the production overlay and confirm it still has exactly one
   `virtengine-validator` replica:

   ```bash
   node scripts/task85c-validate-kubernetes.mjs
   ```

5. Apply the canonical production overlay only after the old validator is fenced.
   Do not scale `virtengine-validator` above one replica.

6. Confirm readiness through the validator service and chain status endpoint in
   the target cluster.

## Provider Restore

1. Stop or fence provider replicas in the failed region to prevent concurrent
   writes to restored queues and checkpoints.

2. Restore shared PVCs first:

   - `provider-ha-state`
   - `provider-backups`

3. Restore the shared encrypted provider identity PVC `provider-data`.

4. Rehydrate `provider-daemon-secrets` with provider address, key fingerprint,
   encrypted-keystore passphrase file, key backup passphrase, WALDUR chain
   keyring passphrase, and portal auth secret.

5. Apply the canonical production overlay. The provider StatefulSet should
   recreate pods using stable pod identities and durable paths under
   `/var/lib/virtengine/provider` and `/var/lib/virtengine/provider-ha`.

6. Verify `/health` and `/ready`, confirm one Kubernetes Lease holder and a
   fencing token greater than or equal to the backup epoch, then check that
   queue, account/usage sequence, reconciliation, and checkpoints advance
   monotonically before admitting external traffic. A restored stale pod must
   remain unready and unable to broadcast.

## Regional Recovery

1. Fence the old validator signer and provider write path.
2. Restore validator chain state and signer metadata.
3. Restore provider identity, HA-state and backup PVCs.
4. Render and validate:

   ```bash
   kubectl kustomize --load-restrictor=LoadRestrictionsNone deploy/kubernetes/overlays/prod
   kubectl kustomize --load-restrictor=LoadRestrictionsNone infra/kubernetes/overlays/prod
   node scripts/task85c-validate-kubernetes.mjs
   node scripts/validate-agents-docs.mjs
   ```

5. Apply the canonical production overlay.
6. Confirm validator readiness, sentry readiness, provider readiness, and
   provider queue/checkpoint continuity.

Target recovery objectives from the Task 85C planning scope are 15 minutes for
validator signer/state recovery and 30 minutes for provider/regional recovery.
Those targets require a live cluster drill and signer/storage backend evidence
before they can be claimed as achieved.

## Failure Rules

- Do not restore the same validator signer identity into two active validator
  pods.
- Do not scale `virtengine-validator` above one replica.
- Do not restore provider queues without restoring the distributed lease/fencing
  metadata.
- Do not bypass immutable image or persistence validation to speed up recovery.
- Do not treat render validation as proof that the live signer, HSM, storage
  snapshots, or cloud backup account are healthy.
- Do not call the software `SoftHSMProvider` a PKCS#11 hardware certification.
   A real vendor/device profile, mTLS continuity, outage/failover drill, and
   retained evidence remain release-only requirements.
