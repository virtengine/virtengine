# Inference Sidecar Deployment Guide

## Overview

The VEID inference sidecar provides deterministic ML inference for identity scoring
in a blockchain consensus environment. This document describes the deployment topology
and resource requirements.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Validator Node                          │
│  ┌─────────────────┐     gRPC      ┌──────────────────────┐ │
│  │   VirtEngine    │◄─────────────►│  Inference Sidecar   │ │
│  │   (Chain Node)  │  localhost:   │  (TensorFlow Model)  │ │
│  │                 │    50051      │                      │ │
│  └─────────────────┘               └──────────────────────┘ │
│         │                                   │               │
│         ▼                                   ▼               │
│  ┌─────────────────┐               ┌──────────────────────┐ │
│  │   Prometheus    │◄──────────────│  /metrics (9090)     │ │
│  │   Metrics       │               │                      │ │
│  └─────────────────┘               └──────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Deployment Modes

### 1. Sidecar Mode (Recommended for Production)

The inference sidecar runs as a separate process alongside the validator node.
This provides:
- Memory isolation between chain and inference
- Independent scaling and resource management
- Easier model updates without chain restarts

**Configuration:**
```bash
# On the validator node
export VEID_INFERENCE_ENABLED=true
export VEID_INFERENCE_USE_SIDECAR=true
export VEID_INFERENCE_SIDECAR_ADDR=localhost:50051
export VEID_INFERENCE_MODEL_HASH=<expected-model-hash>
```

**Starting the sidecar:**
```bash
./inference-sidecar \
    --grpc-addr=:50051 \
    --metrics-addr=:9090 \
    --model-path=/models/trust_score \
    --model-version=v1.0.0 \
    --expected-hash=<model-hash> \
    --force-cpu=true \
    --random-seed=42
```

### 2. Embedded Mode (Development/Testing)

TensorFlow is embedded directly in the chain node process.
Only recommended for development and testing.

**Configuration:**
```bash
export VEID_USE_TENSORFLOW=true
export VEID_INFERENCE_MODEL_PATH=/models/trust_score
export VEID_INFERENCE_MODEL_HASH=<expected-model-hash>
```

## Resource Requirements

### Minimum Requirements (per validator)

| Resource    | Minimum   | Recommended |
|-------------|-----------|-------------|
| CPU         | 2 cores   | 4 cores     |
| RAM         | 2 GB      | 4 GB        |
| Disk        | 1 GB      | 5 GB        |
| Network     | 1 Gbps    | 10 Gbps     |

### Sidecar-Specific Requirements

| Resource    | Minimum   | Recommended |
|-------------|-----------|-------------|
| CPU         | 1 core    | 2 cores     |
| RAM         | 512 MB    | 1 GB        |
| Model Size  | ~50 MB    | ~50 MB      |

### Latency Requirements

| Metric              | Requirement |
|---------------------|-------------|
| P99 Latency         | < 500 ms    |
| P95 Latency         | < 200 ms    |
| Timeout             | 2 seconds   |

## Determinism Configuration

**CRITICAL:** All validators MUST use identical determinism settings for consensus.

### Required Environment Variables

```bash
# TensorFlow Determinism
export CUDA_VISIBLE_DEVICES=-1      # Disable GPU
export TF_DETERMINISTIC_OPS=1       # Use deterministic ops
export TF_CUDNN_DETERMINISTIC=1     # cuDNN determinism
export TF_USE_CUDNN_AUTOTUNE=0      # Disable auto-tuning
export TF_ENABLE_ONEDNN_OPTS=0      # Disable oneDNN
export OMP_NUM_THREADS=1            # Single thread
export PYTHONHASHSEED=42            # Fixed Python hash seed

# VirtEngine Inference
export VEID_INFERENCE_DETERMINISTIC=true
export VEID_INFERENCE_FORCE_CPU=true
```

### Fixed Configuration Values

| Setting               | Required Value |
|-----------------------|----------------|
| Random Seed           | 42             |
| Inter-Op Parallelism  | 1              |
| Intra-Op Parallelism  | 1              |
| CPU Only              | true           |
| Deterministic Ops     | true           |

## Model Management

### Model Version Requirements

All validators in the active set MUST use:
1. The same model version
2. The same model weights (verified by hash)
3. The same TensorFlow version

### Model Hash Verification

```bash
# Compute model hash
sha256sum /models/trust_score/saved_model.pb

# Verify in sidecar
./inference-sidecar --expected-hash=<hash> ...
```

### Model Updates

1. Propose model update via governance
2. All validators download new model
3. Verify hash matches governance proposal
4. Coordinate upgrade at specific block height

## Monitoring

### Prometheus Metrics

The sidecar exposes metrics at `/metrics`:

| Metric                          | Type      | Description                    |
|---------------------------------|-----------|--------------------------------|
| veid_inference_total            | Counter   | Total inference requests       |
| veid_inference_latency_seconds  | Histogram | Inference latency              |
| veid_inference_model_info       | Gauge     | Model version and hash         |
| veid_inference_sidecar_healthy  | Gauge     | Sidecar health status          |
| veid_inference_score_distribution | Histogram | Score distribution           |

### Health Check Endpoints

| Endpoint    | Port  | Description          |
|-------------|-------|----------------------|
| /health     | 9090  | HTTP health check    |
| gRPC Health | 50051 | gRPC health service  |

### Alerting Rules

```yaml
groups:
  - name: inference-sidecar
    rules:
      - alert: InferenceSidecarDown
        expr: up{job="inference-sidecar"} == 0
        for: 1m
        labels:
          severity: critical

      - alert: HighInferenceLatency
        expr: histogram_quantile(0.99, veid_inference_latency_seconds) > 1
        for: 5m
        labels:
          severity: warning

      - alert: InferenceErrors
        expr: rate(veid_inference_total{status="error"}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
```

## Docker Deployment

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o inference-sidecar ./cmd/inference-sidecar

FROM alpine:3.18
RUN apk add --no-cache libc6-compat
COPY --from=builder /app/inference-sidecar /usr/local/bin/
COPY models/trust_score /models/trust_score

ENV CUDA_VISIBLE_DEVICES=-1
ENV TF_DETERMINISTIC_OPS=1
ENV OMP_NUM_THREADS=1

EXPOSE 50051 9090

ENTRYPOINT ["inference-sidecar"]
CMD ["--grpc-addr=:50051", "--metrics-addr=:9090", "--model-path=/models/trust_score"]
```

### Docker Compose

```yaml
version: '3.8'

services:
  inference-sidecar:
    image: virtengine/inference-sidecar:v1.0.0
    ports:
      - "50051:50051"
      - "9090:9090"
    environment:
      - CUDA_VISIBLE_DEVICES=-1
      - TF_DETERMINISTIC_OPS=1
      - OMP_NUM_THREADS=1
    volumes:
      - ./models:/models:ro
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '1'
          memory: 512M
    healthcheck:
      test: ["CMD", "wget", "-q", "--spider", "http://localhost:9090/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## Kubernetes Deployment

### Sidecar Container

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: validator
spec:
  containers:
    - name: virtengine
      image: virtengine/virtengine:latest
      env:
        - name: VEID_INFERENCE_ENABLED
          value: "true"
        - name: VEID_INFERENCE_USE_SIDECAR
          value: "true"
        - name: VEID_INFERENCE_SIDECAR_ADDR
          value: "localhost:50051"

    - name: inference-sidecar
      image: virtengine/inference-sidecar:v1.0.0
      ports:
        - containerPort: 50051
        - containerPort: 9090
      env:
        - name: CUDA_VISIBLE_DEVICES
          value: "-1"
        - name: TF_DETERMINISTIC_OPS
          value: "1"
      resources:
        limits:
          cpu: "2"
          memory: "1Gi"
        requests:
          cpu: "1"
          memory: "512Mi"
      livenessProbe:
        httpGet:
          path: /health
          port: 9090
        initialDelaySeconds: 10
        periodSeconds: 30
      readinessProbe:
        grpc:
          port: 50051
        initialDelaySeconds: 5
        periodSeconds: 10
```

## Troubleshooting

### Common Issues

1. **Hash Mismatch**
   - Verify model file integrity
   - Check TensorFlow version matches
   - Ensure deterministic export

2. **High Latency**
   - Check CPU resources
   - Verify single-threaded mode
   - Monitor memory usage

3. **Connection Refused**
   - Verify sidecar is running
   - Check port configuration
   - Verify network policies

### Debug Mode

```bash
./inference-sidecar \
    --log-level=debug \
    --enable-reflection=true \
    ...
```

### Testing Determinism

```bash
# Run determinism verification
grpcurl -plaintext localhost:50051 \
    inference.v1.InferenceService/VerifyDeterminism \
    -d '{"test_vector_id": "high_quality_verification"}'
```
