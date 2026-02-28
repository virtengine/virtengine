# VEID Verification TEE Runbook

This runbook operates the `deploy/tee` package and assumes the resources in `_docs/verification-deployment.md` are already applied.

## Primary Objects

- `Deployment/tee-enclave-sgx`
- `Deployment/tee-enclave-sev-snp`
- `ExternalSecret/tee-attestation-material`
- `ExternalSecret/tee-measurement-allowlist`
- `Service/tee-enclave`
- `PrometheusRule/tee-enclave`
- `ServiceMonitor/tee-enclave`

## Steady-State Checks

Run these before and after any change:

```bash
kubectl get pods -n virtengine -l app.kubernetes.io/name=tee-enclave -o wide
kubectl get externalsecret -n virtengine tee-attestation-material tee-measurement-allowlist
kubectl rollout status deployment/tee-enclave-sgx -n virtengine --timeout=5m
kubectl rollout status deployment/tee-enclave-sev-snp -n virtengine --timeout=5m
```

Healthy state means:

- both deployments have at least one ready replica
- both External Secrets are synced
- no active `TEEEnclaveSignatureFailures`
- no active `TEEEnclaveStaleAttestations`

## Signer Rotation

1. Update the backing secret `virtengine/prod/tee/attestation`.
2. If measurement approvals changed, update `virtengine/prod/tee/measurements`.
3. Force External Secret refresh:

```bash
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
kubectl annotate externalsecret tee-attestation-material -n virtengine "force-sync=${STAMP}" --overwrite
kubectl annotate externalsecret tee-measurement-allowlist -n virtengine "force-sync=${STAMP}" --overwrite
```

4. Restart the first platform:

```bash
kubectl rollout restart deployment/tee-enclave-sgx -n virtengine
kubectl rollout status deployment/tee-enclave-sgx -n virtengine --timeout=10m
```

5. Verify no signer or stale-attestation alert is firing.
6. Repeat for `tee-enclave-sev-snp`.

The supported command-backed path is:

```bash
./deploy/tee/failure-recovery-drill.sh sgx
./deploy/tee/failure-recovery-drill.sh sev-snp
```

## Stale Attestation Alert

Trigger: `TEEEnclaveStaleAttestations`

1. Check whether `tee-measurement-allowlist` or `tee-attestation-material` missed a sync.
2. Force External Secret refresh.
3. Restart only the affected platform deployment.
4. Confirm `rollout status`.
5. Confirm the alert clears.

Commands:

```bash
kubectl get externalsecret tee-attestation-material tee-measurement-allowlist -n virtengine
kubectl rollout restart deployment/tee-enclave-sgx -n virtengine
kubectl rollout status deployment/tee-enclave-sgx -n virtengine --timeout=10m
```

If the alert persists after a refresh and rollout, stop the rollout and treat the problem as bad attestation material, not a transient platform issue.

## Signer Failure Alert

Trigger: `TEEEnclaveSignatureFailures`

1. Freeze further rotations until the failing signer bundle is identified.
2. Check recent pod restarts and readiness:

```bash
kubectl get pods -n virtengine -l app.kubernetes.io/name=tee-enclave
kubectl describe pod -n virtengine <pod-name>
kubectl logs -n virtengine <pod-name> --all-containers --tail=200
```

3. Verify `tee-attestation-material` contains the expected `signer-public.pem` and `rotation-bundle.json`.
4. Re-sync External Secret and restart the affected deployment.
5. If failures continue, roll back to the last known-good secret version in Secrets Manager and restart again.

Close the incident only after signer failures stop increasing and both platform pools are ready.

## External Secret Sync Failure

Symptoms:

- init container `tee-secret-preflight` crash loops
- no ready pods after rollout
- `kubectl get secret` does not show the expected target objects

Recovery:

```bash
kubectl describe externalsecret tee-attestation-material -n virtengine
kubectl describe externalsecret tee-measurement-allowlist -n virtengine
kubectl get secret -n virtengine tee-attestation-material tee-measurement-allowlist
```

1. Fix the backing secret content or permissions.
2. Force a new sync with `force-sync`.
3. Delete the stuck pods so they restart after the sync succeeds:

```bash
kubectl delete pod -n virtengine -l virtengine.com/tee-platform=sgx
kubectl delete pod -n virtengine -l virtengine.com/tee-platform=sev-snp
```

## SGX Hardware Failure

Symptoms:

- `tee-hardware-preflight` fails on SGX pods
- `TEEEnclaveReplicasUnavailable` for `tee-enclave-sgx`

Recovery:

1. Verify SGX-capable nodes still advertise:
   - `virtengine.com/tee-platform=sgx`
   - `virtengine.com/enclave-ready=true`
2. Cordon or drain the bad node if required.
3. Restore at least one healthy SGX node.
4. Restart the SGX deployment if scheduling does not recover automatically.

Commands:

```bash
kubectl get nodes -L virtengine.com/tee-platform,virtengine.com/enclave-ready
kubectl rollout restart deployment/tee-enclave-sgx -n virtengine
kubectl rollout status deployment/tee-enclave-sgx -n virtengine --timeout=10m
```

## SEV-SNP Hardware Failure

Symptoms:

- `tee-hardware-preflight` fails on SEV-SNP pods
- `TEEEnclaveReplicasUnavailable` for `tee-enclave-sev-snp`

Recovery mirrors SGX:

```bash
kubectl get nodes -L virtengine.com/tee-platform,virtengine.com/enclave-ready
kubectl rollout restart deployment/tee-enclave-sev-snp -n virtengine
kubectl rollout status deployment/tee-enclave-sev-snp -n virtengine --timeout=10m
```

## Rollback Rule

Rollback the active platform deployment when any of the following is true:

- `tee-secret-preflight` fails after one forced External Secret refresh
- `TEEEnclaveSignatureFailures` continues increasing after a restart
- `TEEEnclaveStaleAttestations` stays firing for more than 15 minutes after refresh
- the replacement pod never becomes ready within the rollout timeout

Rollback action:

1. Restore the last known-good Secrets Manager values.
2. Force External Secret refresh.
3. Restart the affected deployment.
4. Wait for `rollout status`.

## Drill Completion Criteria

A failure-recovery drill is complete only when:

- `./deploy/tee/failure-recovery-drill.sh <platform>` exits successfully
- the deployment returns to ready state
- External Secrets show synced status
- no `TEEEnclaveSignatureFailures` alert is firing
- no `TEEEnclaveStaleAttestations` alert is firing
