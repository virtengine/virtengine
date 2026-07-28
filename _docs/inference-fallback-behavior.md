# VEID Inference Fallback Behavior

This document describes the development fallback strategy and health check behavior for the VEID ML inference system (VE-205).

## Overview

The VEID identity verification system uses ML inference to compute trust scores. Production consensus uses signed inference receipts and must fail closed when production inference evidence is unavailable or invalid. Local stub behavior remains an explicit development and test path, not a production automatic fallback.

## Scorer Selection Hierarchy

The system selects an ML scorer in the following order:

1. **TensorFlow Scorer** (preferred for production)
   - Used when `VEID_USE_TENSORFLOW=true` or `VEID_INFERENCE_ENABLED=true`
   - Requires model to be available at configured path
   - Uses embedded TensorFlow-Go or gRPC sidecar

2. **Sidecar Client** (transport client for pre-consensus workers)
   - Used when `VEID_INFERENCE_USE_SIDECAR=true`
   - Connects to external inference sidecar via gRPC
   - Provides better isolation and GPU support
   - Production inference-sidecar transport is mTLS-only and uses projected transport certs, not consensus keys
   - TLS mode requires a complete client certificate/key/server-CA tuple; partial or one-way TLS configuration fails before dialing
   - Authenticated score responses must match the model version and digest pinned during model-info exchange
   - The current production keeper consensus path does not select scoring behavior from these environment variables; it consumes signed receipts

3. **Stub Scorer** (fallback/development)
   - Used when TensorFlow is disabled or unavailable
   - Provides deterministic scores based on feature values
   - Not production consensus evidence and not an automatic production fallback
   - The sidecar escape hatch `--allow-fallback-to-stub` is for local dev/test only
   - A runtime backend value of `local_stub` in staging or production is an incident, not a tolerated steady state

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VEID_USE_TENSORFLOW` | `false` | Enable TensorFlow-based inference |
| `VEID_INFERENCE_ENABLED` | `false` | Alternative flag to enable inference |
| `VEID_DISABLE_TENSORFLOW` | `false` | Force disable TensorFlow (overrides enable) |
| `VEID_INFERENCE_MODEL_PATH` | `models/trust_score` | Path to TensorFlow SavedModel |
| `VEID_INFERENCE_MODEL_HASH` | `` | Expected SHA256 hash of model |
| `VEID_INFERENCE_USE_SIDECAR` | `false` | Use gRPC sidecar instead of embedded TF |
| `VEID_INFERENCE_SIDECAR_ADDR` | `localhost:50051` | Sidecar gRPC address |
| `VEID_INFERENCE_SIDECAR_TLS` | `false` | Require TLS for the sidecar client |
| `VEID_INFERENCE_SIDECAR_TLS_CERT_FILE` | `` | Client transport certificate path |
| `VEID_INFERENCE_SIDECAR_TLS_KEY_FILE` | `` | Client transport private key path |
| `VEID_INFERENCE_SIDECAR_TLS_CA_FILE` | `` | Server CA path |
| `VEID_INFERENCE_DETERMINISTIC` | `true` | Enforce deterministic mode |
| `VEID_INFERENCE_FORCE_CPU` | `true` | Force CPU-only execution |

## Development Fallback Triggers

In explicit non-production development or test configuration, the system may fall back to `StubMLScorer` when:

1. **Model not found**: `VEID_INFERENCE_MODEL_PATH` doesn't exist
2. **Hash mismatch**: Model hash doesn't match `VEID_INFERENCE_MODEL_HASH`
3. **Initialization failure**: TensorFlow session creation fails
4. **Sidecar unavailable**: gRPC connection to sidecar fails
5. **Health check failure**: Scorer reports unhealthy

Production policy:
- Production and staging deployment assets must not enable `--allow-fallback-to-stub`
- Production mTLS startup requires a non-empty primary serving endpoint and rejects `--serving-fallback-url`, CPU drift, and random-seed drift
- HTTP readiness fails closed when the configured serving backend is unhealthy even if the local bundle verification succeeded
- CI validates deployment, workflow, Kubernetes render, and public-doc deployable examples to keep the fallback disabled outside dev/test
- Consensus keeper production behavior must be based on signed inference receipts, not env-selected fallback or sidecar flags

## Evidence Boundaries

Current local evidence covers:

1. Inference sidecar startup and gRPC behavior with ephemeral generated CA, server, and client certificates.
2. Kustomize rendering for canonical `deploy/kubernetes` and compatibility `infra/kubernetes` resources.
3. Python deployment policy unit tests and rendered policy CLI checks.
4. Task 85C Kubernetes semantic validator compatibility.

Current evidence does not claim:

1. Live cert-manager issuance.
2. Live External Secrets backend synchronization.
3. Private CA operations or certificate rotation.
4. Production model or TensorFlow Serving runtime execution.
5. Named production cluster storage behavior.
6. Workload identity behavior.
7. Real cluster traffic between a deployed validator, sidecar, monitoring stack, and runtime backend.

Before production readiness is declared, operators must collect the live evidence listed in `_docs/runbooks/production-inference-sidecar.md`.

## Consensus Safety

### Determinism Requirements

For blockchain consensus, all validators must produce identical scores for the same inputs:

- **Same model version**: All validators must use the same model version
- **Same model hash**: SHA256 hash of model weights must match
- **CPU-only execution**: GPU execution can introduce non-determinism
- **Single-threaded**: Inter-op and intra-op parallelism set to 1
- **Fixed random seed**: Random seed fixed at 42
- **Input/output hashes**: All operations compute hashes for verification

### Verification Process

1. Pre-consensus inference workers produce model/version/hash evidence over authenticated transport.
2. Production keeper consensus verifies signed receipts and receipt metadata.
3. Score, model version, and hashes must match the governed receipt policy.
4. Missing or mismatched production evidence causes rejection or an explicit non-production path, not automatic stub scoring.

## Health Checks

### Passive Health Check

```go
scorer.IsHealthy() // Returns true if scorer is operational
```

### Active Health Check

```go
healthChecker.Check() // Runs test inference and verifies determinism
```

### Health Status Fields

| Field | Description |
|-------|-------------|
| `Healthy` | Scorer is fully operational |
| `Degraded` | Scorer operational but with issues |
| `ModelLoaded` | Model successfully loaded |
| `ModelVersion` | Current model version string |
| `ModelHash` | SHA256 of model weights |
| `DeterminismVerified` | Determinism test passed |
| `LastInferenceLatencyMs` | Latency of last inference |

## Prometheus Metrics

| Metric | Description |
|--------|-------------|
| `veid_inference_health_check_total` | Total health checks by result |
| `veid_inference_health_check_failures_total` | Failed health checks |
| `veid_inference_determinism_verified` | Determinism status (1=verified) |
| `veid_inference_health_check_latency_seconds` | Health check latency |
| `veid_inference_score_total` | Total scoring requests by status |
| `veid_inference_score_duration_seconds` | Scoring latency |

## Model Hot-Reload

The system supports model updates without restart:

```go
reloadManager.Reload(newModelPath)
```

Features:
- Atomic model swap (no dropped requests)
- Hash verification before activation
- Automatic rollback on failure
- Optional file watcher for auto-reload

## Troubleshooting

### Scorer Falls Back to Stub

Check:
1. Model path exists: `ls $VEID_INFERENCE_MODEL_PATH`
2. Model hash matches: Compare with `VEID_INFERENCE_MODEL_HASH`
3. TensorFlow enabled: `VEID_USE_TENSORFLOW=true`
4. View logs for initialization errors

### Consensus Mismatch

Check:
1. All validators have same model version
2. Model hashes match across validators
3. Determinism settings match (CPU-only, single-threaded)
4. Compare input/output hashes in vote extensions

### High Inference Latency

Check:
1. CPU-only mode may be slower than GPU
2. Model size and complexity
3. Sidecar network latency
4. Consider increasing timeout

## See Also

- [VEID Flow Specification](veid-flow-spec.md)
- [Validator VEID Pipeline Guide](validator-veid-pipeline-guide.md)
- [Verification Deployment](verification-deployment.md)
- [ML Feature Schema](veid-ml-feature-schema.md)
