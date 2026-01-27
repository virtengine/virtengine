# VirtEngine Facial Verification Pipeline

**Task:** VE-211  
**Version:** 1.0.0  
**Status:** Complete

## Overview

This package provides deterministic facial verification for the VirtEngine VEID identity verification system. It enables validators to verify that a selfie/live capture matches either:
- A face extracted from an ID document
- A previously enrolled face reference

All operations are designed for blockchain consensus - same inputs produce identical outputs across validators.

## Features

- **Face Detection**: MTCNN, RetinaFace, or OpenCV backends
- **Face Alignment**: Eye-based alignment for consistent embeddings
- **Face Embedding**: VGG-Face, Facenet, Facenet512, ArcFace, Dlib, SFace
- **Verification**: Configurable thresholds with match/borderline/no-match decisions
- **Determinism**: CPU-only execution, fixed seeds, model hash verification
- **Reason Codes**: Detailed explanations for all verification outcomes

## Installation

```bash
pip install -r requirements.txt
```

## Quick Start

```python
from ml.facial_verification import FaceVerifier, VerificationConfig
import cv2

# Load images
selfie = cv2.imread("selfie.jpg")
id_face = cv2.imread("id_document_face.jpg")

# Create verifier with default config
verifier = FaceVerifier()

# Verify
result = verifier.verify(selfie, id_face)

print(f"Match: {result.match}")
print(f"Decision: {result.decision}")
print(f"Similarity: {result.similarity_score:.2%}")
print(f"Reason codes: {result.reason_codes}")
```

## Configuration

### Default Thresholds

| Threshold | Value | Meaning |
|-----------|-------|---------|
| Match | ≥90% | Definite match |
| Borderline | 85-90% | Inconclusive, may require secondary verification |
| Reject | <70% | Definite no-match |

### Custom Configuration

```python
from ml.facial_verification import VerificationConfig, ModelName

config = VerificationConfig(
    # Model settings
    model_name=ModelName.ARCFACE,
    model_version="1.0.0",
    
    # Thresholds
    match_threshold=0.95,      # High security
    borderline_lower=0.90,
    reject_threshold=0.80,
    
    # Preprocessing
    preprocessing=PreprocessingConfig(
        target_resolution=(224, 224),
        apply_clahe=True,
        noise_reduction=True,
    ),
    
    # Determinism
    determinism=DeterminismConfig(
        force_cpu=True,
        seed=42,
        deterministic_ops=True,
    ),
)

verifier = FaceVerifier(config)
```

### Pre-defined Configurations

```python
from ml.facial_verification.config import (
    DEFAULT_CONFIG,       # Balanced defaults
    HIGH_SECURITY_CONFIG, # Strict thresholds (95%+)
    PERMISSIVE_CONFIG,    # Relaxed thresholds (85%+)
)
```

## Supported Models

| Model | Embedding Dim | LFW Accuracy | Speed |
|-------|---------------|--------------|-------|
| VGG-Face | 2622 | 97.62% | Slow |
| Facenet | 128 | 99.63% | Medium |
| Facenet512 | 512 | 99.65% | Medium |
| ArcFace | 512 | 99.82% | Fast |
| Dlib | 128 | 99.38% | Fast |
| SFace | 128 | 99.40% | Very Fast |

## Verification Result

```python
@dataclass
class VerificationResult:
    match: bool                  # True if faces match
    decision: str                # "match", "no_match", "borderline"
    similarity_score: float      # 0.0 - 1.0
    confidence_percent: float    # 0 - 100
    model_name: str              # e.g., "VGG-Face"
    model_version: str           # e.g., "1.0.0"
    model_hash: str              # SHA256 of model weights
    reason_codes: List[str]      # Explanation codes
    embeddings_hash: str         # For consensus verification
    result_hash: str             # Complete result hash
    processing_time_ms: float    # Processing time
```

## Reason Codes

### Success Codes
- `MATCH_CONFIRMED` - Verification successful
- `HIGH_CONFIDENCE_MATCH` - Very high similarity (≥95%)

### Detection Issues
- `NO_FACE_DETECTED` - No face found in image
- `MULTIPLE_FACES` - More than one face detected
- `FACE_TOO_SMALL` - Face region too small
- `FACE_PARTIALLY_VISIBLE` - Face not fully in frame

### Quality Issues
- `LOW_QUALITY_IMAGE` - Overall low quality
- `IMAGE_TOO_DARK` - Underexposed
- `IMAGE_TOO_BRIGHT` - Overexposed
- `IMAGE_BLURRY` - Motion or focus blur
- `LOW_RESOLUTION` - Resolution too low

### Verification Issues
- `BORDERLINE_MATCH` - Score in borderline range
- `LOW_SIMILARITY_SCORE` - Below match threshold
- `EMBEDDING_MISMATCH` - Faces don't match

## Determinism for Consensus

The pipeline ensures identical results across validators:

```python
from ml.facial_verification import DeterminismController

controller = DeterminismController(config)
controller.ensure_deterministic()

# All validators will now produce identical results
```

### Determinism Controls

1. **Random Seeds**: Fixed at 42 (configurable)
2. **CPU-Only**: GPU disabled via `CUDA_VISIBLE_DEVICES=-1`
3. **Deterministic Ops**: TensorFlow deterministic operations enabled
4. **Model Hash**: Weights hash verified before verification
5. **Result Hash**: SHA256 of canonical result for consensus

## Testing

```bash
# Run all tests
pytest ml/facial_verification/tests/ -v

# Run with coverage
pytest ml/facial_verification/tests/ -v --cov=ml.facial_verification

# Run specific test file
pytest ml/facial_verification/tests/test_verification.py -v
```

## Architecture

```
ml/facial_verification/
├── __init__.py           # Package exports
├── config.py             # Configuration classes
├── preprocessing.py      # Image preprocessing pipeline
├── face_detection.py     # Face detection and alignment
├── embeddings.py         # Face embedding extraction
├── verification.py       # Main verification logic
├── determinism.py        # Determinism controls
├── reason_codes.py       # Reason code definitions
├── requirements.txt      # Dependencies
├── models/
│   ├── __init__.py
│   └── model_registry.py # Model version registry
└── tests/
    ├── __init__.py
    ├── conftest.py       # Test fixtures
    ├── test_preprocessing.py
    ├── test_detection.py
    ├── test_verification.py
    └── test_determinism.py
```

## Integration with VEID

This pipeline is used by validators during identity verification:

1. User uploads identity scope (encrypted selfie + ID document)
2. Validator decrypts scope using their private key
3. Validator extracts face from ID document (VE-216)
4. **Validator runs facial verification (this pipeline)**
5. Verification result contributes to VEID trust score
6. Consensus reached when validators agree on result hash

## Security Considerations

- Model weights are hashed and verified before use
- All intermediate results are hashed for audit
- Deterministic execution prevents validator disagreements
- Reason codes provide transparency for users

## License

Proprietary - VirtEngine Project
