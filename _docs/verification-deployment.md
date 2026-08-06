# VEID Verification TEE Deployment

This guide is the production deployment contract for the TEE-backed VEID verification path defined in `deploy/tee/`.

## Scope

The deployable package is:

- `deploy/tee/kustomization.yaml`
- `deploy/tee/common-resources.yaml`
- `deploy/tee/config.yaml`
- `deploy/tee/secrets.yaml`
- `deploy/tee/monitoring.yaml`
- `deploy/tee/sgx-deployment.yaml`
- `deploy/tee/sev-deployment.yaml`
- `deploy/tee/validate.sh`
- `deploy/tee/failure-recovery-drill.sh`

The package deploys two hardware-backed verifier pools in the `virtengine` namespace:

- `tee-enclave-sgx`
- `tee-enclave-sev-snp`

Shared objects created by the package:

- `ServiceAccount/tee-enclave`
- `Service/tee-enclave`
- `Service/tee-enclave-headless`
- `PodDisruptionBudget/tee-enclave-pdb`
- `ConfigMap/tee-enclave-config`
- `ExternalSecret/tee-attestation-material`
- `ExternalSecret/tee-measurement-allowlist`
- `ServiceMonitor/tee-enclave`
- `PrometheusRule/tee-enclave`

## Secret Sources

The deployment fails closed until External Secrets sync the required material from AWS Secrets Manager.

`ExternalSecret/tee-attestation-material` reads from `virtengine/prod/tee/attestation` and must publish:

- `attestation-client-cert.pem`
- `attestation-client-key.pem`
- `attestation-ca.pem`
- `signer-public.pem`
- `rotation-bundle.json`

`ExternalSecret/tee-measurement-allowlist` reads from `virtengine/prod/tee/measurements` and must publish:

- `measurements.json`

The SGX and SEV init containers refuse startup when any required file is missing, empty, or still carries forbidden launch-token text.

## Readiness Model

Each platform deployment uses three layers of protection:

1. `tee-secret-preflight` validates synced secret files and non-empty measurement allowlists before the main container starts.
2. `tee-hardware-preflight` validates platform hardware access:
   - SGX: `/dev/sgx_enclave` or `/dev/sgx/enclave`, plus `/dev/sgx_provision` or `/dev/sgx/provision`
   - SEV-SNP: `/dev/sev-guest`
3. The main container exposes:
   - `startupProbe` on `GET /healthz`
   - `livenessProbe` on `GET /healthz`
   - `readinessProbe` on `GET /readyz`

The operational meaning is:

- no secret sync: pod never starts
- bad hardware mapping: pod never starts
- stale or broken runtime: pod starts but never becomes ready
- repeated runtime failure: pod is restarted by liveness

## Deployment Procedure

1. Render and validate the package locally:

```bash
./deploy/tee/validate.sh
```

2. Confirm the secret backends contain the current attestation and measurement artifacts:

- `virtengine/prod/tee/attestation`
- `virtengine/prod/tee/measurements`

3. Apply the package:

```bash
kubectl apply -k deploy/tee
```

4. Wait for External Secrets to reconcile:

```bash
kubectl get externalsecret -n virtengine tee-attestation-material tee-measurement-allowlist
kubectl get secret -n virtengine tee-attestation-material tee-measurement-allowlist
```

5. Wait for both verifier pools:

```bash
kubectl rollout status deployment/tee-enclave-sgx -n virtengine --timeout=10m
kubectl rollout status deployment/tee-enclave-sev-snp -n virtengine --timeout=10m
```

6. Check services and monitoring objects:

```bash
kubectl get svc -n virtengine tee-enclave tee-enclave-headless
kubectl get servicemonitor -n virtengine tee-enclave
kubectl get prometheusrule -n virtengine tee-enclave
```

## Rotation Contract

Signer and attestation trust material is rotated by updating the backing Secrets Manager entries and forcing an External Secret refresh. The rollout order is:

1. Update `virtengine/prod/tee/attestation`
2. Update `virtengine/prod/tee/measurements` if the new signer bundle changes approved measurements
3. Force `ExternalSecret` refresh
4. Restart one platform deployment at a time
5. Wait for `rollout status`
6. Verify no signer or stale-attestation alerts remain firing

Use the drill script to execute the restart and reconciliation flow:

```bash
./deploy/tee/failure-recovery-drill.sh sgx
./deploy/tee/failure-recovery-drill.sh sev-snp
```

## Alert Expectations

`PrometheusRule/tee-enclave` defines the required operator alerts:

- `TEEEnclaveReplicasUnavailable`
- `TEEEnclaveSignatureFailures`
- `TEEEnclaveAttestationFailures`
- `TEEEnclaveStaleAttestations`
- `TEEEnclaveUnhealthyValidators`

Operator response rules:

- `TEEEnclaveSignatureFailures`: treat as critical signer-path failure; rotate or roll back signer material before re-enabling traffic.
- `TEEEnclaveAttestationFailures`: treat as critical attestation-path failure; check cert chain, trust roots, and hardware quote freshness.
- `TEEEnclaveStaleAttestations`: refresh attestation material and restart the affected pool before the 24-hour freshness budget is exceeded further.
- `TEEEnclaveReplicasUnavailable`: restore at least one SGX replica and one SEV-SNP replica before closing the incident.

## Recovery Expectations

Operators should use the runbook in `_docs/verification-runbook.md` for:

- signer rotation
- stale attestation recovery
- External Secret sync failure
- SGX node loss
- SEV-SNP node loss
- platform-by-platform rollback

The deployment is considered production-ready only when both platform deployments pass probes, both External Secrets have synced successfully, and the drill script has completed cleanly for the active platform.
