# Production Inference Sidecar Runbook

## Purpose

This runbook covers the VEID production inference-sidecar deployment slice. It is scoped to Kubernetes render, policy, local mTLS startup behavior, and operator checks for one sidecar replica per validator identity.

## Production Invariants

- The canonical source is `deploy/kubernetes`; `infra/kubernetes` is an import-only compatibility shim and must render identically for base and production overlays.
- `StatefulSet/inference-sidecar` has exactly one replica.
- The sidecar image is immutable and pinned with `@sha256:<64 hex>`.
- gRPC listens on TCP 50051 and requires mTLS.
- HTTP readiness, liveness, startup, and metrics use `/health` or `/metrics` on TCP 9092.
- Production mTLS startup requires a remote serving endpoint, rejects serving fallback and local stub execution, and fixes CPU-only execution with random seed 42.
- `/health` is ready only when both the pinned local bundle and the configured serving backend are healthy.
- The sidecar requires absolute server cert, server key, and client CA paths.
- The server leaf certificate must be current, non-CA, digital-signature capable, and authorized for server authentication.
- The sidecar uses `reloader.stakater.com/auto: "true"` so mounted transport and metadata Secret rotation restarts the pod through the existing reloader convention.
- The model/runtime bundle is read-only and pinned by explicit model, runtime, and manifest digest metadata.
- No deployable staging or production runtime enables local stub fallback.
- No ExternalSecret contains consensus private keys, validator private keys, inference receipt signing keys, or other raw signing material.

## Required Secrets

The sidecar server identity uses:

- `inference-sidecar-server-mtls`: `tls.crt`, `tls.key`
- `inference-sidecar-client-ca`: `ca.crt`

The validator transport client identity uses:

- `inference-sidecar-client-mtls`: `tls.crt`, `tls.key`, `ca.crt`

The model metadata uses:

- `inference-sidecar-model-metadata`: `model-version`, `model-digest`, `runtime-digest`, `manifest-digest`

These files are transport identity only. Consensus key material must stay in the existing consensus key management path and must not be added to these resources.

## Render Checks

Run all canonical and compatibility renders before promoting a manifest change:

```bash
kubectl kustomize --load-restrictor=LoadRestrictionsNone deploy/kubernetes/base
kubectl kustomize --load-restrictor=LoadRestrictionsNone deploy/kubernetes/overlays/dev
kubectl kustomize --load-restrictor=LoadRestrictionsNone deploy/kubernetes/overlays/staging
kubectl kustomize --load-restrictor=LoadRestrictionsNone deploy/kubernetes/overlays/prod
kubectl kustomize --load-restrictor=LoadRestrictionsNone infra/kubernetes/base
kubectl kustomize --load-restrictor=LoadRestrictionsNone infra/kubernetes/overlays/staging
kubectl kustomize --load-restrictor=LoadRestrictionsNone infra/kubernetes/overlays/prod
```

Then run the focused policy gates:

```bash
python .github/tests/test_inference_deployment_policy.py
python .github/scripts/validate_inference_deployment_policy.py
node scripts/task85c-validate-kubernetes.mjs
```

## Preflight Checklist

Before rollout:

- Confirm prod overlay ExternalSecret remote references point at prod transport certs and prod model/bundle metadata.
- Confirm the sidecar image digest matches the promoted build artifact.
- Confirm the model bundle digest values match governance-approved model metadata.
- Confirm NetworkPolicy allows validator-to-sidecar TCP 50051 and monitoring-to-sidecar TCP 9092 only.
- Confirm the validator pod has read-only client cert/key/CA mounts and sidecar endpoint metadata.
- Confirm the sidecar pod security context remains non-root, read-only root filesystem, no privilege escalation, RuntimeDefault seccomp, and dropped capabilities.
- Confirm the sidecar pod has read-only mounts at `/var/run/secrets/virtengine/inference-sidecar/server`, `/var/run/secrets/virtengine/inference-sidecar/client-ca`, `/var/run/secrets/virtengine/inference-sidecar/model-metadata`, and `/models/trust-score/v1`.

## Runtime Checks

After rollout, collect live evidence before declaring production readiness:

- `kubectl -n virtengine rollout status statefulset/inference-sidecar`
- `kubectl -n virtengine get pods,svc,networkpolicy,pdb,serviceaccount -l app.kubernetes.io/name=inference-sidecar`
- `kubectl -n virtengine describe externalsecret inference-sidecar-server-mtls`
- `kubectl -n virtengine describe externalsecret inference-sidecar-client-ca`
- `kubectl -n virtengine describe externalsecret inference-sidecar-client-mtls`
- `kubectl -n virtengine describe externalsecret inference-sidecar-model-metadata`
- Sidecar `/health` on port 9092 from an allowed monitoring workload.
- mTLS gRPC health or inference request from the validator workload using the projected client cert/key and CA.
- Failed plaintext gRPC attempt from an allowed diagnostic workload.
- Failed mTLS attempt with a client certificate outside the configured client CA.

The validator readiness probe checks environment values and verifies that the projected mTLS files are readable. It is not mTLS handshake evidence. The immutable validator image in this slice has no documented TLS-capable gRPC probe binary, so operators must collect the live mTLS success and failure evidence above separately before declaring the sidecar production-ready.

## Incident Triggers

Treat any of these as deployment incidents:

- The sidecar starts without required mTLS.
- The sidecar accepts plaintext gRPC.
- A client without a valid client certificate reaches gRPC 50051 successfully.
- Sidecar replicas drift away from one for a validator identity.
- A public Service, ingress, NodePort, or LoadBalancer exposes sidecar gRPC or metrics.
- Runtime args include local fallback scoring for staging or production.
- ExternalSecret or Secret content references raw consensus, validator, or signing private keys.
- The canonical and infra compatibility renders differ.

## Evidence Limits

This slice claims only local evidence from:

- Go mTLS tests with ephemeral generated CA, server, and client certificates.
- Kustomize render checks for canonical and compatibility manifests.
- The Python deployment policy unit test and rendered policy CLI.
- Task 85C Kubernetes validator compatibility checks.

This slice does not claim live evidence for cert-manager, an External Secrets backend, private CA issuance, certificate rotation, production model or TensorFlow Serving runtime, named cluster storage, workload identity, validator readiness as an mTLS handshake, or real cluster traffic.
