# Inference Sidecar Deployment Guide

## Overview

The VEID inference sidecar serves deterministic ML inference to one validator identity. Production deployments are fail-closed: gRPC requires mutual TLS, the model/runtime bundle is pinned by digest metadata, and local stub fallback is not part of deployable runtime configuration.

Canonical Kubernetes assets live under `deploy/kubernetes`. `infra/kubernetes` is an import-only compatibility shim and must render identically to the canonical tree.

## Production Topology

```
Validator pod
  pre-consensus inference worker transport identity
    mTLS client cert/key/server CA at /var/run/secrets/virtengine/inference-sidecar-client
    VEID_INFERENCE_SIDECAR_ADDR=inference-sidecar.virtengine.svc.cluster.local:50051

Inference sidecar StatefulSet
  gRPC: 50051, mTLS required
  HTTP readiness/liveness/metrics: /health and /metrics on 9092
  server cert/key at /var/run/secrets/virtengine/inference-sidecar/server
  client CA at /var/run/secrets/virtengine/inference-sidecar/client-ca
  model metadata at /var/run/secrets/virtengine/inference-sidecar/model-metadata
  read-only model/runtime bundle at /models/trust-score/v1
```

The sidecar Service is `ClusterIP` only. There is no public ingress, no HPA, and the production overlay keeps exactly one sidecar replica aligned to one validator identity.

## Canonical Kubernetes

Render the canonical base or overlays with Kustomize:

```bash
kubectl kustomize --load-restrictor=LoadRestrictionsNone deploy/kubernetes/base
kubectl kustomize --load-restrictor=LoadRestrictionsNone deploy/kubernetes/overlays/staging
kubectl kustomize --load-restrictor=LoadRestrictionsNone deploy/kubernetes/overlays/prod
```

The canonical set includes:

- `StatefulSet/inference-sidecar` with one replica, immutable digest image, non-root security context, read-only root filesystem, dropped capabilities, bounded writable tmp, anti-affinity, topology spread, and `/health` probes on port 9092.
- `Service/inference-sidecar` as an internal `ClusterIP` with gRPC 50051 and metrics/readiness 9092.
- `ServiceAccount/inference-sidecar`, `PodDisruptionBudget/inference-sidecar-pdb`, `NetworkPolicy/inference-sidecar`, and a read-only model bundle claim.
- ExternalSecret definitions for the server transport cert/key, client CA, validator client transport cert/key/server CA, and model/bundle metadata. These secrets are transport and model metadata only; they do not contain consensus keys or inference receipt signing keys.
- `StatefulSet/inference-sidecar` uses `reloader.stakater.com/auto: "true"` so transport certificate and metadata Secret rotation restarts the pod through the existing reloader convention.

Production images must use immutable digests:

```yaml
image: ghcr.io/virtengine/inference-sidecar@sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd
```

Replace placeholder digest values with governed, environment-approved image and model/runtime digests before live deployment.

## Sidecar Arguments

Production sidecar args must include mTLS and deterministic runtime settings:

```yaml
args:
  - --grpc-addr=:50051
  - --metrics-addr=:9092
  - --require-mtls=true
  - --tls-cert-file=/var/run/secrets/virtengine/inference-sidecar/server/tls.crt
  - --tls-key-file=/var/run/secrets/virtengine/inference-sidecar/server/tls.key
  - --tls-client-ca-file=/var/run/secrets/virtengine/inference-sidecar/client-ca/ca.crt
  - --model-path=/models/trust-score/v1/model
  - --model-version=v1.0.0
  - --expected-hash=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  - --manifest-path=/models/trust-score/v1/release_manifest.json
  - --serving-url=http://tf-serving.virtengine-runtime.svc.cluster.local:8501
  - --force-cpu=true
  - --random-seed=42
```

The sidecar refuses to start with required mTLS unless all cert paths are non-empty absolute paths to readable regular files and the server certificate, key, and client CA parse successfully. The server leaf must be within its validity window, allow digital signatures, and carry a server-authentication EKU. The client CA bundle must contain at least one CA certificate authorized for certificate signing; leaf-only or malformed trailing PEM is rejected. TLS is configured with TLS 1.3 minimum and client certificate verification. Production mTLS mode also requires the remote `--serving-url`, rejects a secondary serving endpoint and local stub fallback, and requires CPU-only execution with random seed 42.

The validator-side client is also fail-closed: enabling sidecar TLS requires an explicit client certificate, matching private key, and server CA bundle. One-way TLS and partial mTLS file configuration are rejected before dialing, and every score response must retain the model version and digest established by the authenticated model-info exchange.

Development plaintext is available only by explicitly passing `--require-mtls=false` for local testing. That opt-out is non-production and does not enable fallback scoring.

## Validator Configuration

The validator deployment projects the mTLS client identity into:

```text
/var/run/secrets/virtengine/inference-sidecar-client/tls.crt
/var/run/secrets/virtengine/inference-sidecar-client/tls.key
/var/run/secrets/virtengine/inference-sidecar-client/ca.crt
```

The validator environment records the sidecar endpoint and transport paths:

```bash
VEID_INFERENCE_ENABLED=true
VEID_INFERENCE_USE_SIDECAR=true
VEID_INFERENCE_SIDECAR_ADDR=inference-sidecar.virtengine.svc.cluster.local:50051
VEID_INFERENCE_SIDECAR_TLS=true
VEID_INFERENCE_SIDECAR_TLS_CERT_FILE=/var/run/secrets/virtengine/inference-sidecar-client/tls.crt
VEID_INFERENCE_SIDECAR_TLS_KEY_FILE=/var/run/secrets/virtengine/inference-sidecar-client/tls.key
VEID_INFERENCE_SIDECAR_TLS_CA_FILE=/var/run/secrets/virtengine/inference-sidecar-client/ca.crt
```

These values are transport identity metadata projected for the pre-consensus inference worker. The current production keeper consensus path uses signed receipts and must not select consensus scoring behavior from these environment variables. Consensus identity and signing material must remain outside the inference sidecar transport secrets.

## Network Policy

The canonical policy allows:

- validator workloads to call sidecar TCP 50051;
- monitoring workloads to scrape sidecar TCP 9092;
- sidecar DNS egress;
- sidecar egress to TensorFlow Serving only when that runtime dependency is used.

The validator NetworkPolicy also allows egress to the sidecar service on TCP 50051. Do not expose sidecar gRPC or metrics through a public Service or ingress.

## Monitoring

The sidecar exposes:

| Endpoint | Port | Purpose |
| --- | --- | --- |
| `/health` | 9092 | bundle verification plus live serving-backend startup, readiness, and liveness checks |
| `/metrics` | 9092 | Prometheus metrics |
| gRPC health | 50051 | internal service health |

Expected metrics include inference request counts, latency histograms, model info, and health status. Scraping must come from the monitoring namespace or pods selected by the canonical NetworkPolicy.

## Validation

Run the deployment policy gate locally before proposing changes:

```bash
python .github/tests/test_inference_deployment_policy.py
python .github/scripts/validate_inference_deployment_policy.py
node scripts/task85c-validate-kubernetes.mjs
```

The gate rejects mutable images, missing mTLS args, plaintext production-like configuration, fallback/stub runtime tokens, public services, missing NetworkPolicy/PDB/security controls, replica drift, duplicate named resources, cross-environment inference secret references, raw signing or consensus key references, and canonical-vs-infra render drift across base, staging, and production.

The validator readiness probe in the current immutable validator image proves required environment values and mounted transport files are present and readable. It does not prove a successful mTLS handshake to the sidecar. Until a known TLS-capable probe binary is included in that image, production readiness must include separately collected live mTLS probe evidence from the validator workload or an approved diagnostic workload.

## Operational Notes

- Rotate transport certificates through the environment External Secrets backend and the private CA process approved for the target cluster.
- Coordinate model/runtime bundle digest changes through governance and update the model metadata ExternalSecret references in the target overlay.
- Treat any production render that disables mTLS, exposes a public sidecar Service, changes sidecar replicas away from one, or configures fallback scoring as a deployment incident.
