# VirtEngine VEID Pipeline: Validator Operations Guide

**Version:** 1.0.0  
**Date:** 2026-01-26  
**Task Reference:** VE-219

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Installation](#installation)
4. [Verification of Image Integrity](#verification-of-image-integrity)
5. [Running the Pipeline](#running-the-pipeline)
6. [Conformance Testing](#conformance-testing)
7. [Troubleshooting](#troubleshooting)
8. [Upgrading the Pipeline](#upgrading-the-pipeline)
9. [Security Considerations](#security-considerations)

---

## Overview

The VirtEngine Identity Verification (VEID) pipeline is a deterministic ML inference system that computes identity trust scores for users. As a validator, you **MUST** run the exact same pipeline version as other validators to participate in consensus verification.

### Key Concepts

- **Deterministic Execution**: All validators must produce identical outputs for identical inputs
- **Pinned Container Image**: The pipeline runs in a versioned OCI container with fixed dependencies
- **Model Weight Hashes**: All ML model weights are verified by SHA256 hash
- **Consensus Verification**: Validators recompute verification results and vote on consensus

### Why Determinism Matters

VirtEngine validators recompute identity verification results during consensus. If your pipeline produces different outputs than other validators:

1. Your votes will be rejected
2. You may be slashed for invalid attestations
3. The network cannot reach consensus on identity scores

---

## Prerequisites

### System Requirements

| Resource | Minimum | Recommended |
|----------|---------|-------------|
| CPU | 4 cores (x86_64) | 8+ cores |
| RAM | 8 GB | 16+ GB |
| Disk | 50 GB | 100+ GB |
| Network | 100 Mbps | 1 Gbps |

**Important**: GPU execution is **disabled** to ensure determinism. All inference runs on CPU.

### Software Requirements

- Docker Engine 24.0+ or containerd 1.7+
- VirtEngine node software (latest release)
- `sha256sum` or equivalent for hash verification

### Network Access

The pipeline container requires no outbound network access during operation. All models and dependencies are bundled in the image.

---

## Installation

### Step 1: Pull the Pipeline Image

```bash
# Pull the specific versioned image
docker pull ghcr.io/virtengine/veid-pipeline:v1.0.0

# Verify the pull was successful
docker images ghcr.io/virtengine/veid-pipeline:v1.0.0
```

### Step 2: Verify Image Hash

**CRITICAL**: Always verify the image hash matches the on-chain registered hash before running.

```bash
# Get the image digest (SHA256)
docker inspect --format='{{.RepoDigests}}' ghcr.io/virtengine/veid-pipeline:v1.0.0

# Expected output format:
# [ghcr.io/virtengine/veid-pipeline@sha256:a1b2c3d4e5f6...]
```

Query the current active pipeline version from the chain:

```bash
# Using virtengine CLI
virtengine query veid active-pipeline-version

# Expected output:
# version: "1.0.0"
# image_hash: "sha256:a1b2c3d4e5f6..."
# model_manifest_hash: "abc123..."
# status: "active"
```

**Verify the image hash from `docker inspect` matches the `image_hash` from the chain query.**

### Step 3: Verify Model Weights

The pipeline includes pre-baked model weights. Verify their hashes:

```bash
# Run the hash verification command inside the container
docker run --rm ghcr.io/virtengine/veid-pipeline:v1.0.0 \
  python -m ml.verify_models

# Expected output:
# Model: deepface_facenet512
#   Version: 1.0.0
#   Hash: sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2 ✓
# Model: craft_text_detection
#   Version: 1.0.0
#   Hash: sha256:b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3 ✓
# Model: unet_face_extraction
#   Version: 1.0.0
#   Hash: sha256:c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4 ✓
# 
# All model hashes verified ✓
```

### Step 4: Configure the Validator Node

Add the pipeline configuration to your validator's `config.toml`:

```toml
[veid]
# Enable VEID pipeline
enabled = true

# Pipeline sidecar mode
pipeline_mode = "sidecar"

# Sidecar gRPC address
sidecar_address = "localhost:50051"

# Pipeline version (must match chain-registered version)
pipeline_version = "1.0.0"

# Expected image hash (from chain query)
expected_image_hash = "sha256:a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2"

# Expected model manifest hash (from chain query)
expected_model_manifest_hash = "abc123def456..."

# Timeout for pipeline operations (milliseconds)
pipeline_timeout_ms = 5000
```

---

## Running the Pipeline

### Option A: Docker Compose (Recommended)

Create a `docker-compose.veid.yml`:

```yaml
version: '3.8'

services:
  veid-pipeline:
    image: ghcr.io/virtengine/veid-pipeline:v1.0.0
    container_name: veid-pipeline
    restart: unless-stopped
    ports:
      - "50051:50051"
    environment:
      # Determinism settings (already set in image, but explicit here)
      - TF_DETERMINISTIC_OPS=1
      - CUDA_VISIBLE_DEVICES=-1
      - OMP_NUM_THREADS=1
      - PYTHONHASHSEED=42
    volumes:
      # Optional: mount logs directory
      - ./logs/veid:/app/logs
    healthcheck:
      test: ["CMD", "python", "-c", "import grpc; print('healthy')"]
      interval: 30s
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
        reservations:
          cpus: '2'
          memory: 4G
```

Start the pipeline:

```bash
docker compose -f docker-compose.veid.yml up -d

# Check logs
docker compose -f docker-compose.veid.yml logs -f
```

### Option B: Direct Docker Run

```bash
docker run -d \
  --name veid-pipeline \
  --restart unless-stopped \
  -p 50051:50051 \
  -e TF_DETERMINISTIC_OPS=1 \
  -e CUDA_VISIBLE_DEVICES=-1 \
  -e OMP_NUM_THREADS=1 \
  -e PYTHONHASHSEED=42 \
  --cpus 4 \
  --memory 8g \
  ghcr.io/virtengine/veid-pipeline:v1.0.0
```

### Option C: Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: veid-pipeline
  labels:
    app: veid-pipeline
spec:
  replicas: 1
  selector:
    matchLabels:
      app: veid-pipeline
  template:
    metadata:
      labels:
        app: veid-pipeline
    spec:
      containers:
      - name: veid-pipeline
        image: ghcr.io/virtengine/veid-pipeline:v1.0.0
        ports:
        - containerPort: 50051
        env:
        - name: TF_DETERMINISTIC_OPS
          value: "1"
        - name: CUDA_VISIBLE_DEVICES
          value: "-1"
        - name: OMP_NUM_THREADS
          value: "1"
        - name: PYTHONHASHSEED
          value: "42"
        resources:
          limits:
            cpu: "4"
            memory: 8Gi
          requests:
            cpu: "2"
            memory: 4Gi
        livenessProbe:
          grpc:
            port: 50051
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          grpc:
            port: 50051
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: veid-pipeline
spec:
  selector:
    app: veid-pipeline
  ports:
  - port: 50051
    targetPort: 50051
```

---

## Verification of Image Integrity

### Full Integrity Check

Run the complete integrity verification:

```bash
# Pull the verification script
virtengine veid verify-pipeline --version 1.0.0

# Or manually:
docker run --rm ghcr.io/virtengine/veid-pipeline:v1.0.0 \
  python -m ml.verify_integrity --full

# Expected output:
# ========================================
# VirtEngine VEID Pipeline Integrity Check
# ========================================
# 
# Pipeline Version: 1.0.0
# Image Hash: sha256:a1b2c3d4e5f6... ✓
# 
# Model Verification:
#   deepface_facenet512: ✓
#   craft_text_detection: ✓
#   unet_face_extraction: ✓
#   identity_scorer_v1: ✓
# 
# Manifest Hash: abc123def456... ✓
# 
# Determinism Settings:
#   TF_DETERMINISTIC_OPS: 1 ✓
#   CUDA_VISIBLE_DEVICES: -1 ✓
#   OMP_NUM_THREADS: 1 ✓
# 
# ========================================
# All integrity checks passed ✓
# ========================================
```

### Comparing with Chain State

```bash
# Query chain for expected values
virtengine query veid active-pipeline-version -o json | jq .

# Verify local image matches
virtengine veid compare-pipeline-hash \
  --local-image ghcr.io/virtengine/veid-pipeline:v1.0.0 \
  --chain-version 1.0.0
```

---

## Conformance Testing

### Running the Full Test Suite

Before participating in consensus, run the conformance test suite to verify your pipeline produces deterministic outputs:

```bash
# Run conformance tests
virtengine veid run-conformance-tests --version 1.0.0

# Or using Docker directly:
docker run --rm \
  -v $(pwd)/conformance-results:/app/results \
  ghcr.io/virtengine/veid-pipeline:v1.0.0 \
  python -m ml.conformance_runner \
    --suite default \
    --output /app/results/conformance.json
```

### Expected Output

```
========================================
VirtEngine VEID Conformance Test Suite
Pipeline Version: 1.0.0
========================================

Running 8 test vectors...

[1/8] face_detect_001: Standard Face Detection
      Input Hash:  a1b2c3d4... ✓
      Output Hash: b2c3d4e5... ✓
      PASSED

[2/8] face_embed_001: Face Embedding Generation
      Input Hash:  c3d4e5f6... ✓
      Output Hash: d4e5f6a1... ✓
      Intermediate Hashes: ✓
      PASSED

... (remaining tests)

========================================
CONFORMANCE TEST RESULTS
========================================
Total:  8
Passed: 8
Failed: 0

All conformance tests passed ✓
Your pipeline is ready for consensus participation.
========================================
```

### Interpreting Failures

If any tests fail:

1. **Hash Mismatch**: Your image or models differ from the expected version
2. **Output Mismatch**: Non-deterministic execution detected
3. **Missing Intermediate**: Pipeline stage not running correctly

**Do not participate in consensus if conformance tests fail.**

---

## Troubleshooting

### Common Issues

#### Issue: Image Hash Mismatch

```
Error: image hash mismatch
Expected: sha256:a1b2c3d4...
Actual:   sha256:ffffffff...
```

**Solution**: Re-pull the correct image version:
```bash
docker rmi ghcr.io/virtengine/veid-pipeline:v1.0.0
docker pull ghcr.io/virtengine/veid-pipeline:v1.0.0
```

#### Issue: Model Hash Verification Failed

```
Error: model weights hash mismatch for deepface_facenet512
```

**Solution**: The image may be corrupted. Re-pull and verify:
```bash
docker rmi ghcr.io/virtengine/veid-pipeline:v1.0.0
docker pull ghcr.io/virtengine/veid-pipeline:v1.0.0
docker run --rm ghcr.io/virtengine/veid-pipeline:v1.0.0 python -m ml.verify_models
```

#### Issue: Conformance Test Output Mismatch

```
[3/8] ocr_extract_001: OCR Field Extraction
      Output Hash: MISMATCH
      Expected: abcdef12...
      Actual:   99999999...
      FAILED
```

**Possible Causes**:
1. Running on GPU (should be disabled)
2. Multi-threading enabled
3. Different CPU architecture
4. Memory issues causing numerical instability

**Solution**: Verify determinism settings:
```bash
docker run --rm ghcr.io/virtengine/veid-pipeline:v1.0.0 env | grep -E "TF_|OMP_|CUDA_"
```

#### Issue: Connection Refused on Port 50051

**Solution**: Verify the container is running and healthy:
```bash
docker ps
docker logs veid-pipeline
```

---

## Upgrading the Pipeline

When a new pipeline version is activated on-chain:

### Step 1: Check for Upgrade

```bash
# Query upcoming pipeline version
virtengine query veid pending-pipeline-versions

# Or subscribe to upgrade proposals
virtengine query gov proposals --status voting_period
```

### Step 2: Download New Version

```bash
# Pull the new version (example: v1.1.0)
docker pull ghcr.io/virtengine/veid-pipeline:v1.1.0

# Verify integrity
docker run --rm ghcr.io/virtengine/veid-pipeline:v1.1.0 \
  python -m ml.verify_integrity --full
```

### Step 3: Run Conformance Tests

```bash
virtengine veid run-conformance-tests --version 1.1.0
```

### Step 4: Update Configuration

Update `config.toml`:
```toml
[veid]
pipeline_version = "1.1.0"
expected_image_hash = "sha256:new_hash..."
expected_model_manifest_hash = "new_manifest_hash..."
```

### Step 5: Switch Containers

```bash
# Stop old container
docker stop veid-pipeline

# Start new container
docker run -d \
  --name veid-pipeline-new \
  -p 50051:50051 \
  ... (same flags as before) \
  ghcr.io/virtengine/veid-pipeline:v1.1.0

# Verify new container is healthy
docker logs veid-pipeline-new

# Remove old container
docker rm veid-pipeline
docker rename veid-pipeline-new veid-pipeline
```

### Step 6: Restart Validator

```bash
systemctl restart virtengine
```

---

## Security Considerations

### Image Verification

**Always verify image hashes before running.** A compromised image could:
- Produce invalid verification results
- Exfiltrate sensitive data
- Cause consensus failures and slashing

### Network Isolation

The pipeline container should have:
- No outbound internet access
- Only inbound connections from the validator node on port 50051
- No access to host filesystem (except logs if mounted)

### Resource Limits

Set strict resource limits to prevent DoS:
```bash
docker run --cpus 4 --memory 8g ...
```

### Audit Logging

Enable audit logging for all pipeline invocations:
```toml
[veid]
audit_logging = true
audit_log_path = "/var/log/virtengine/veid-audit.log"
```

---

## Support

For issues with the VEID pipeline:

1. Check the troubleshooting section above
2. Search existing GitHub issues
3. Open a new issue with:
   - Pipeline version
   - Image hash
   - Conformance test output
   - Relevant logs

**Never include sensitive data (identity documents, biometric data) in support requests.**
