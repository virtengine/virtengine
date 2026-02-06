# Task 29A: ML Training + SavedModel Export

**ID:** 29A  
**Title:** feat(ml): Execute ML training pipeline + export SavedModel  
**Priority:** P0 (Critical Blocker)  
**Wave:** 1 (Parallel with 29C)  
**Estimated LOC:** ~500  
**Dependencies:** None  
**Blocking:** 29B (Model Hash Computation)  

---

## Problem Statement

The ML training infrastructure exists in the `ml/` directory with complete pipeline code, but **no trained model artifacts have been generated**. Without SavedModel artifacts:

1. VEID scoring returns zeros (stubbed inference)
2. Model hash governance has nothing to hash
3. Validators cannot verify they're using approved models
4. Production identity verification is non-functional

### Current State Analysis

```
ml/
├── facial_verification/
│   ├── train.py              ✅ Exists
│   ├── model.py              ✅ Exists
│   ├── requirements.txt      ✅ Exists
│   └── weights/              ❌ EMPTY (no .pt files)
├── liveness_detection/
│   ├── train.py              ✅ Exists
│   ├── model.py              ✅ Exists
│   └── weights/              ❌ EMPTY
└── ocr_extraction/
    ├── train.py              ✅ Exists
    ├── model.py              ✅ Exists
    └── weights/              ❌ EMPTY
```

---

## Acceptance Criteria

### AC-1: Facial Verification Model Training
- [ ] Execute `ml/facial_verification/train.py` with deterministic settings
- [ ] Achieve minimum accuracy threshold (>95% on validation set)
- [ ] Export model to TensorFlow SavedModel format
- [ ] Output to `ml/facial_verification/weights/facial_model/`
- [ ] Generate model metrics JSON with accuracy, F1, AUC, latency

### AC-2: Liveness Detection Model Training
- [ ] Execute `ml/liveness_detection/train.py` with deterministic settings
- [ ] Achieve minimum accuracy threshold (>98% for anti-spoofing)
- [ ] Export model to TensorFlow SavedModel format
- [ ] Output to `ml/liveness_detection/weights/liveness_model/`
- [ ] Generate model metrics JSON

### AC-3: OCR Extraction Model Training
- [ ] Execute `ml/ocr_extraction/train.py` with deterministic settings
- [ ] Achieve minimum accuracy threshold (>97% character accuracy)
- [ ] Export model to TensorFlow SavedModel format
- [ ] Output to `ml/ocr_extraction/weights/ocr_model/`
- [ ] Generate model metrics JSON

### AC-4: Determinism Verification
- [ ] All models trained with `TF_DETERMINISTIC_OPS=1`
- [ ] Fixed random seed (42) for all operations
- [ ] CPU-only training (no GPU variance) OR documented GPU determinism
- [ ] Same training produces identical model bytes (bit-exact reproducibility)
- [ ] Dependencies pinned via `requirements-deterministic.txt`

### AC-5: Documentation
- [ ] Update `ml/README.md` with training instructions
- [ ] Document hardware requirements (RAM, disk, GPU optional)
- [ ] Document training duration estimates
- [ ] Document how to verify model integrity

---

## Technical Requirements

### Determinism Configuration (Critical for Consensus)

```python
# Required settings in all training scripts
import os
import random
import numpy as np
import tensorflow as tf

# Environment
os.environ['TF_DETERMINISTIC_OPS'] = '1'
os.environ['TF_CUDNN_DETERMINISTIC'] = '1'
os.environ['PYTHONHASHSEED'] = '42'

# Seeds
SEED = 42
random.seed(SEED)
np.random.seed(SEED)
tf.random.set_seed(SEED)

# Force CPU if needed for determinism
os.environ['CUDA_VISIBLE_DEVICES'] = ''  # Optional: CPU-only
```

### SavedModel Export Format

```python
# Export format for TensorFlow Serving / Go inference
model.save(
    'weights/model_name',
    save_format='tf',
    signatures=tf.saved_model.DEFAULT_SERVING_SIGNATURE_DEF_KEY
)

# Also export frozen graph for hash verification
concrete_func = model.signatures[tf.saved_model.DEFAULT_SERVING_SIGNATURE_DEF_KEY]
frozen_func = convert_variables_to_constants_v2(concrete_func)
tf.io.write_graph(
    frozen_func.graph.as_graph_def(),
    'weights',
    'model_name_frozen.pb',
    as_text=False
)
```

### Model Metrics Output

```json
{
  "model_name": "facial_verification",
  "version": "1.0.0",
  "training_date": "2026-02-06T12:00:00Z",
  "framework": "tensorflow",
  "framework_version": "2.15.0",
  "metrics": {
    "accuracy": 0.9623,
    "precision": 0.9587,
    "recall": 0.9612,
    "f1_score": 0.9599,
    "auc_roc": 0.9891
  },
  "inference_latency_ms": {
    "p50": 12.3,
    "p95": 18.7,
    "p99": 24.1
  },
  "training_config": {
    "epochs": 100,
    "batch_size": 32,
    "learning_rate": 0.001,
    "optimizer": "adam",
    "seed": 42
  },
  "dataset": {
    "train_samples": 50000,
    "validation_samples": 10000,
    "test_samples": 10000
  },
  "determinism": {
    "tf_deterministic_ops": true,
    "random_seed": 42,
    "cpu_only": true
  }
}
```

---

## Files to Create/Modify

### New Files
| Path | Description |
|------|-------------|
| `ml/facial_verification/weights/facial_model/` | SavedModel directory |
| `ml/facial_verification/weights/facial_model_frozen.pb` | Frozen graph for hashing |
| `ml/facial_verification/weights/metrics.json` | Training metrics |
| `ml/liveness_detection/weights/liveness_model/` | SavedModel directory |
| `ml/liveness_detection/weights/liveness_model_frozen.pb` | Frozen graph |
| `ml/liveness_detection/weights/metrics.json` | Training metrics |
| `ml/ocr_extraction/weights/ocr_model/` | SavedModel directory |
| `ml/ocr_extraction/weights/ocr_model_frozen.pb` | Frozen graph |
| `ml/ocr_extraction/weights/metrics.json` | Training metrics |
| `ml/requirements-deterministic.txt` | Pinned dependencies |

### Files to Modify
| Path | Changes |
|------|---------|
| `ml/facial_verification/train.py` | Add determinism settings, export logic |
| `ml/liveness_detection/train.py` | Add determinism settings, export logic |
| `ml/ocr_extraction/train.py` | Add determinism settings, export logic |
| `ml/README.md` | Training documentation |

---

## Implementation Steps

### Step 1: Prepare Deterministic Environment
```bash
cd ml/
python -m venv .venv
source .venv/bin/activate  # or .venv\Scripts\activate on Windows
pip install -r requirements-deterministic.txt
```

### Step 2: Update Training Scripts
Add determinism configuration to each `train.py`:
- Set all random seeds
- Enable TF_DETERMINISTIC_OPS
- Configure CPU-only if needed
- Add SavedModel export at end of training

### Step 3: Execute Training
```bash
# Facial verification (~2-4 hours)
cd facial_verification
TF_DETERMINISTIC_OPS=1 python train.py --seed 42 --export-format savedmodel

# Liveness detection (~1-2 hours)
cd ../liveness_detection
TF_DETERMINISTIC_OPS=1 python train.py --seed 42 --export-format savedmodel

# OCR extraction (~3-5 hours)
cd ../ocr_extraction
TF_DETERMINISTIC_OPS=1 python train.py --seed 42 --export-format savedmodel
```

### Step 4: Verify Determinism
```bash
# Train again and compare hashes
sha256sum facial_verification/weights/facial_model_frozen.pb
# Should match first run exactly
```

### Step 5: Generate Metrics Reports
```bash
python generate_metrics.py --model facial_verification
python generate_metrics.py --model liveness_detection
python generate_metrics.py --model ocr_extraction
```

---

## Validation Checklist

- [ ] All three models exported to SavedModel format
- [ ] Frozen graphs (.pb) generated for each model
- [ ] Metrics JSON files with accuracy, latency data
- [ ] Determinism verified (re-training produces identical bytes)
- [ ] Model loads successfully in Go inference package
- [ ] Inference returns non-zero scores on test images
- [ ] Training documentation updated

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Training data not available | Use synthetic data or public datasets for initial training |
| GPU determinism issues | Train CPU-only for consensus requirements |
| Long training times | Use smaller model architectures or transfer learning |
| Model accuracy too low | Iterative hyperparameter tuning, data augmentation |

---

## Related Documents

- [_docs/ml-determinism-requirements.md](_docs/ml-determinism-requirements.md) - Determinism specifications
- [_docs/veid-flow-spec.md](_docs/veid-flow-spec.md) - VEID verification flow
- [pkg/inference/](pkg/inference/) - Go inference package that will load these models

---

## Vibe-Kanban Task ID

`d556f2bc-90de-446c-a8a7-7ff9245a8a8a`
