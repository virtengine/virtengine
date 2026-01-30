# Trust Score Model Release Runbook

VE-3A: Complete runbook for training, releasing, and rolling back the VEID trust score model.

## Overview

This document covers:
1. Training a new model version
2. Evaluating model quality
3. Publishing artifacts
4. Submitting governance proposal
5. Rolling back to a previous version

## Prerequisites

- Python 3.11+ with dependencies from `ml/requirements-deterministic.txt`
- Go 1.21+ for inference testing
- Access to the VEID dataset (set `VEID_DATASET_PATH` environment variable)
- VirtEngine CLI (`virtengine`) for governance proposals

## 1. Training a New Model

### 1.1 Prepare Environment

```bash
# Install dependencies with pinned versions for reproducibility
pip install -r ml/requirements-deterministic.txt

# Set environment variables for determinism
export PYTHONHASHSEED=42
export TF_DETERMINISTIC_OPS=1
export TF_CUDNN_DETERMINISTIC=1
export CUDA_VISIBLE_DEVICES=""  # Force CPU
```

### 1.2 Validate Configuration

```bash
# Dry run to validate config without training
python -m ml.training.run_training \
    --config ml/training/configs/trust_score_v1.yaml \
    --dry-run
```

### 1.3 Run Training

```bash
# Full training run
python -m ml.training.run_training \
    --config ml/training/configs/trust_score_v1.yaml \
    --output-dir output/trust_score_$(date +%Y%m%d) \
    --version v$(date +%Y%m%d_%H%M%S)
```

### 1.4 Training Outputs

After training completes, the following artifacts are created:

```
output/trust_score_YYYYMMDD/
├── checkpoints/           # Model checkpoints during training
├── logs/tensorboard/      # TensorBoard logs
├── exported_models/
│   └── vYYYYMMDD_HHMMSS/
│       ├── model/         # TensorFlow SavedModel
│       ├── manifest.json  # Model manifest with hash
│       └── export_metadata.json
├── evaluation_report.txt  # Human-readable metrics
├── evaluation_metrics.json # Machine-readable metrics
└── training_config.json   # Configuration snapshot
```

## 2. Evaluating Model Quality

### 2.1 Evaluation Thresholds

The model must meet these thresholds (defined in `trust_score_v1.yaml`):

| Metric | Threshold | Description |
|--------|-----------|-------------|
| R² | ≥ 0.85 | Coefficient of determination |
| MAE | ≤ 8.0 | Mean Absolute Error |
| RMSE | ≤ 10.0 | Root Mean Squared Error |
| Accuracy@5 | ≥ 60% | Predictions within ±5 points |
| Accuracy@10 | ≥ 80% | Predictions within ±10 points |
| Accuracy@20 | ≥ 95% | Predictions within ±20 points |
| P95 Error | ≤ 15 | 95th percentile error |
| Mean Bias | ≤ ±2 | Mean signed error |

### 2.2 Review Evaluation Report

```bash
cat output/trust_score_YYYYMMDD/evaluation_report.txt
```

### 2.3 Verify Determinism

```bash
# Run inference multiple times and compare
python -c "
import tensorflow as tf
import numpy as np

model = tf.saved_model.load('output/trust_score_YYYYMMDD/exported_models/vXXX/model')
serve = model.signatures['serving_default']

np.random.seed(42)
test_input = tf.constant(np.random.randn(1, 768).astype(np.float32))

results = [serve(test_input)['trust_score'].numpy() for _ in range(5)]
assert all(np.allclose(results[0], r) for r in results), 'Non-deterministic!'
print('Determinism verified')
"
```

## 3. Publishing Artifacts

### 3.1 Local Registry

```bash
# Publish to local artifact registry
python -c "
from ml.training.model.publish import ArtifactPublisher, PublishConfig

config = PublishConfig(
    registry_type='local',
    local_registry_path='artifacts/models',
    enable_immutability=True,
)

publisher = ArtifactPublisher(config)
result = publisher.publish(
    model_path='output/trust_score_YYYYMMDD/exported_models/vXXX/model',
    manifest_path='output/trust_score_YYYYMMDD/exported_models/vXXX/manifest.json',
    version='vYYYYMMDD_HHMMSS',
)
print(f'Published to: {result.artifact_url}')
print(f'Hash: {result.artifact_hash}')
"
```

### 3.2 S3 Registry (Production)

```bash
# Set AWS credentials
export AWS_ACCESS_KEY_ID=xxx
export AWS_SECRET_ACCESS_KEY=xxx
export AWS_REGION=us-east-1

python -c "
from ml.training.model.publish import ArtifactPublisher, PublishConfig

config = PublishConfig(
    registry_type='s3',
    s3_bucket='virtengine-ml-models',
    s3_prefix='models/trust_score',
    enable_immutability=True,
)

publisher = ArtifactPublisher(config)
result = publisher.publish(
    model_path='output/trust_score_YYYYMMDD/exported_models/vXXX/model',
    manifest_path='output/trust_score_YYYYMMDD/exported_models/vXXX/manifest.json',
    version='vYYYYMMDD_HHMMSS',
)
print(f'Published to: {result.artifact_url}')
"
```

## 4. Governance Proposal

### 4.1 Generate Proposal

```bash
python -c "
from ml.training.model.manifest import ManifestGenerator
from ml.training.model.governance import GovernanceProposalGenerator

# Load manifest
gen = ManifestGenerator()
manifest = gen.load('output/trust_score_YYYYMMDD/exported_models/vXXX/manifest.json')

# Generate proposal
proposal_gen = GovernanceProposalGenerator()
proposal = proposal_gen.generate(
    manifest=manifest,
    model_url='s3://virtengine-ml-models/models/trust_score/vXXX/trust_score_vXXX.tar.gz',
)

# Save proposal
proposal_gen.save_proposal(proposal, 'output/proposal.json')
print('Proposal saved to output/proposal.json')
"
```

### 4.2 Submit Proposal

```bash
# Submit to governance
virtengine tx gov submit-proposal output/proposal.json \
    --from=validator \
    --chain-id=virtengine-1 \
    --gas=auto \
    --gas-adjustment=1.5 \
    --fees=5000uvirt \
    -y

# Get proposal ID from output
PROPOSAL_ID=<from output>

# Query proposal status
virtengine query gov proposal $PROPOSAL_ID
```

### 4.3 Vote on Proposal

```bash
# Validators vote on the proposal
virtengine tx gov vote $PROPOSAL_ID yes \
    --from=validator \
    --chain-id=virtengine-1 \
    --gas=auto \
    -y

# Check voting results
virtengine query gov votes $PROPOSAL_ID
virtengine query gov tally $PROPOSAL_ID
```

## 5. Rollback Procedure

### 5.1 When to Rollback

Rollback if any of the following occur:
- Model produces incorrect scores in production
- Determinism failures on validator nodes
- Performance regression beyond thresholds
- Security vulnerability in model weights

### 5.2 Emergency Rollback (Validators)

Validators can immediately switch to a previous model version:

```bash
python -c "
from ml.training.model.rollback import RollbackManager

manager = RollbackManager(
    registry_path='artifacts/models',
    active_model_path='models/trust_score/active',
)

# List available versions
versions = manager.list_available_versions()
for v in versions:
    print(f'{v.version}: {v.model_hash[:16]}... (metrics: R²={v.metrics.get(\"r2\", \"N/A\")})')

# Rollback to previous version
result = manager.rollback(versions[1])  # Second newest
print(f'Rollback: {result.from_version} -> {result.to_version}')
print(f'Success: {result.success}')
"
```

### 5.3 Governance Rollback

For coordinated network-wide rollback:

```bash
# Generate rollback proposal
python -c "
from ml.training.model.rollback import RollbackManager
import json

manager = RollbackManager(registry_path='artifacts/models')
versions = manager.list_available_versions()
target = versions[1]  # Previous version

proposal = manager.generate_rollback_proposal(
    target=target,
    reason='Performance regression detected after v20240115 deployment',
)

with open('rollback_proposal.json', 'w') as f:
    json.dump(proposal, f, indent=2)
print('Rollback proposal saved')
"

# Submit rollback proposal
virtengine tx gov submit-proposal rollback_proposal.json \
    --from=validator \
    --chain-id=virtengine-1 \
    --gas=auto \
    -y
```

### 5.4 Verify Rollback

```bash
# Verify the active model
python -c "
from ml.training.model.rollback import RollbackManager

manager = RollbackManager(
    registry_path='artifacts/models',
    active_model_path='models/trust_score/active',
)

current = manager.get_current_version()
print(f'Active version: {current.version}')
print(f'Hash: {current.model_hash}')

# Verify hash
verified, msg = manager.verify_version(current)
print(f'Verification: {msg}')
"
```

## 6. CI/CD Integration

### 6.1 Automated Verification

The ML Model Verification workflow (`.github/workflows/ml-model-verify.yaml`) runs:
- Training pipeline tests
- Go inference tests
- SavedModel verification
- Manifest hash verification

### 6.2 Manual Verification Trigger

```bash
# Trigger verification workflow manually
gh workflow run ml-model-verify.yaml \
    --field model_path=artifacts/models/vXXX/model \
    --field model_version=vXXX
```

## 7. Troubleshooting

### 7.1 Model Load Failures

```bash
# Debug TensorFlow loading
python -c "
import tensorflow as tf
import logging
logging.getLogger('tensorflow').setLevel(logging.DEBUG)

try:
    model = tf.saved_model.load('path/to/model')
    print('Loaded successfully')
    print(f'Signatures: {list(model.signatures.keys())}')
except Exception as e:
    print(f'Load failed: {e}')
"
```

### 7.2 Hash Mismatches

```bash
# Recompute hash
python -c "
import hashlib
import os
from pathlib import Path

model_path = 'path/to/model'
h = hashlib.sha256()

files = sorted([
    str(f) for f in Path(model_path).rglob('*')
    if f.is_file() and f.name != 'export_metadata.json'
])

for f in files:
    with open(f, 'rb') as fp:
        h.update(fp.read())

print(f'Computed hash: {h.hexdigest()}')
"
```

### 7.3 Determinism Failures

1. Verify TensorFlow version matches `requirements-deterministic.txt`
2. Check environment variables are set
3. Ensure CPU-only execution (`CUDA_VISIBLE_DEVICES=""`)
4. Verify random seeds are set before model load

## 8. Security Considerations

1. **Model Signing**: Future versions will support Ed25519 signatures on manifests
2. **Hash Verification**: Always verify model hash before inference
3. **Immutability**: Published artifacts cannot be modified (object lock enabled)
4. **Audit Trail**: All rollbacks are logged in `artifacts/models/rollback_log.jsonl`

## Appendix: Configuration Reference

See `ml/training/configs/trust_score_v1.yaml` for the complete training configuration.

Key sections:
- `model`: Architecture and hyperparameters
- `evaluation`: Pass/fail thresholds
- `determinism`: Reproducibility settings
- `export`: SavedModel export settings
