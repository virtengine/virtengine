# Trust Score Model Card

## Model Details

| Field | Value |
|-------|-------|
| **Model Name** | VirtEngine Trust Score Model |
| **Version** | v1.0.0 |
| **Model Type** | trust_score |
| **Framework** | TensorFlow 2.15.x |
| **Format** | SavedModel |
| **License** | VirtEngine Proprietary |

## Model Description

The Trust Score Model predicts identity verification trust scores in the range 0-100 based on combined feature vectors from document analysis, facial verification, and capture metadata.

### Purpose
- Provides a unified trust score for VEID identity verification
- Enables consensus-safe scoring across all validator nodes
- Supports deterministic inference for blockchain operations

### Architecture
```
Input: 768-dimensional feature vector
├── Hidden Layer 1: 512 units (ReLU, BatchNorm, Dropout=0.3)
├── Hidden Layer 2: 256 units (ReLU, BatchNorm, Dropout=0.3)
├── Hidden Layer 3: 128 units (ReLU, BatchNorm, Dropout=0.3)
├── Hidden Layer 4: 64 units (ReLU, BatchNorm, Dropout=0.3)
└── Output: 1 unit (Sigmoid × 100 = Trust Score 0-100)
```

### Input Feature Vector (768 dimensions)
| Component | Dimensions | Description |
|-----------|------------|-------------|
| Face Embedding | 512 | Combined document/selfie face embedding |
| Face Scalars | 7 | Confidence, quality, detection flags |
| Document Features | ~50 | Quality metrics, security features |
| OCR Features | 9 | Field confidence and validation |
| Metadata Features | ~190 | Device, session, capture metadata |

## Training Data

### Dataset Characteristics
- **Size**: 100,000+ identity verification samples
- **Document Types**: ID cards, passports, driver's licenses, residence permits, national IDs
- **Geographic Coverage**: Global (190+ countries)
- **Time Period**: 2023-2024

### Data Preprocessing
- All PII anonymized via SHA-256 hashing
- Face embeddings extracted (no raw images stored)
- Quality filtering applied (min confidence thresholds)
- Class balancing via stratified sampling

### Data Splits
| Split | Percentage | Purpose |
|-------|------------|---------|
| Training | 80% | Model training |
| Validation | 10% | Hyperparameter tuning |
| Test | 10% | Final evaluation |

## Performance Metrics

### Target Thresholds (Must Pass All)
| Metric | Threshold | Description |
|--------|-----------|-------------|
| R² | ≥ 0.85 | Coefficient of determination |
| MAE | ≤ 8.0 | Mean Absolute Error |
| RMSE | ≤ 10.0 | Root Mean Squared Error |
| Accuracy@5 | ≥ 60% | Predictions within ±5 points |
| Accuracy@10 | ≥ 80% | Predictions within ±10 points |
| Accuracy@20 | ≥ 95% | Predictions within ±20 points |
| P95 Error | ≤ 15.0 | 95th percentile error |
| Mean Bias | ≤ ±2.0 | Systematic prediction bias |

## Determinism Guarantees

This model is designed for **consensus-critical blockchain operations** and provides the following guarantees:

1. **CPU-Only Execution**: No GPU variance
2. **Fixed Random Seed**: All operations seeded with value 42
3. **Deterministic TensorFlow Ops**: `tf.config.experimental.enable_op_determinism()`
4. **Single-Threaded Execution**: No parallel operation variance
5. **Pinned Dependencies**: Exact versions in requirements-deterministic.txt

### Reproducibility Verification
```bash
# Verify deterministic inference
python -m ml.conformance.generate_test_vectors --model models/trust_score/v1.0.0/model
python -m ml.conformance.verify_go_output --vectors output/test_vectors.json
```

## Usage

### Python Inference
```python
import tensorflow as tf
import numpy as np

# Load model
model = tf.saved_model.load("models/trust_score/v1.0.0/model")
serving_fn = model.signatures["serving_default"]

# Create input
features = np.random.randn(1, 768).astype(np.float32)

# Predict
result = serving_fn(tf.constant(features))
trust_score = result["trust_score"].numpy()[0][0]  # 0-100
```

### Go Inference (VirtEngine Node)
```go
import tf "github.com/tensorflow/tensorflow/tensorflow/go"

model, _ := tf.LoadSavedModel("models/trust_score/v1.0.0/model", []string{"serve"}, nil)
// See pkg/inference/scorer.go for full implementation
```

## Model Governance

### Update Process
1. Train new model with updated config
2. Verify all evaluation thresholds pass
3. Generate governance proposal
4. Submit proposal via `virtengine tx gov submit-proposal`
5. Community votes (7-day voting period)
6. If approved, validators update within sync grace period

### Rollback Procedure
1. Query previous model version from history
2. Submit rollback governance proposal
3. Reference previous model hash for activation

## Ethical Considerations

### Intended Use
- Identity verification for VirtEngine platform
- Trust score calculation for secure access control

### Out-of-Scope Uses
- Standalone identity verification (requires full pipeline)
- Real-time video analysis (designed for still captures)
- Age or demographic prediction

### Bias Mitigation
- Balanced training data across demographics
- Regular evaluation on stratified subgroups
- Continuous monitoring for distribution drift

## Version History

| Version | Date | Hash | Notes |
|---------|------|------|-------|
| v1.0.0 | 2024-XX-XX | See MODEL_HASH.txt | Initial release |

## Contact

- **Maintainer**: VirtEngine Core Team
- **Repository**: github.com/virtengine/virtengine
- **Issues**: github.com/virtengine/virtengine/issues
