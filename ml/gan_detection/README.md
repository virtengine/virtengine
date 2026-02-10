# GAN Fraud Detection Module

**VE-923: GAN-based synthetic image detection for identity fraud prevention**

This module provides comprehensive detection of GAN-generated and manipulated images for the VirtEngine VEID identity verification system.

## Features

### CNN Discriminator
- Multi-scale convolutional neural network for real vs synthetic classification
- Feature extraction at multiple resolution levels
- Binary classification with confidence scores

### Deepfake Detection
- **Face Swap Detection**: Identifies face replacement manipulations
- **Expression Manipulation**: Detects altered facial expressions
- **Face Morphing**: Identifies blended/morphed faces
- **Temporal Analysis**: Analyzes video sequences for consistency

### Artifact Analysis
- **Frequency Domain Analysis**: Detects unnatural spectral patterns
- **Checkerboard Detection**: Identifies transposed convolution artifacts
- **Blending Boundary**: Detects manipulation boundaries
- **Texture Consistency**: Analyzes texture patterns
- **Color Consistency**: Compares face/background color distributions

### VEID Integration
- Score computation compatible with on-chain VEID scoring model
- Penalty calculation for synthetic image detection
- Deterministic execution for blockchain consensus

## Installation

The module uses dependencies from the main ML requirements:

```bash
pip install -r ml/requirements-deterministic.txt
```

## Usage

### Basic Detection

```python
from ml.gan_detection import GANDetector, GANDetectionConfig

# Create detector with default config
config = GANDetectionConfig()
detector = GANDetector(config)

# Detect single image
result = detector.detect(image, face_region=(x1, y1, x2, y2))

# Check result
if result.is_synthetic:
    print(f"Synthetic image detected: {result.decision}")
    print(f"Detection type: {result.detected_type}")
    print(f"Confidence: {result.confidence}")
else:
    print("Image appears authentic")

# Get VEID scoring record
veid_record = result.to_veid_record()
```

### Video Sequence Detection

```python
# Analyze video frames
result = detector.detect_sequence(
    frames=frame_list,
    face_regions=face_region_list,
    landmarks=landmark_list,
)

print(f"Frames analyzed: {result.frames_analyzed}")
print(f"Temporal score: {result.deepfake_score}")
```

### Detection Modes

```python
# Fast mode (lower latency)
config = GANDetectionConfig.fast_mode()
fast_detector = GANDetector(config)

# Accurate mode (higher accuracy)
config = GANDetectionConfig.accurate_mode()
accurate_detector = GANDetector(config)
```

### Detailed Results

```python
result = detector.detect(image, include_details=True)

# Access component results
print(f"Discriminator result: {result.discriminator_result}")
print(f"Deepfake result: {result.deepfake_result}")
print(f"Artifact result: {result.artifact_result}")
```

## Configuration

### GANDetectionConfig

| Parameter | Default | Description |
|-----------|---------|-------------|
| `mode` | `FULL` | Detection mode: FULL, FAST, or ACCURATE |
| `enable_gpu` | `False` | Use GPU acceleration (disabled for determinism) |
| `enforce_determinism` | `True` | Ensure deterministic execution |
| `random_seed` | `42` | Random seed for reproducibility |
| `min_image_size` | `(64, 64)` | Minimum input image size |
| `max_image_size` | `(4096, 4096)` | Maximum input image size |

### DiscriminatorConfig

| Parameter | Default | Description |
|-----------|---------|-------------|
| `input_size` | `(224, 224)` | Model input resolution |
| `base_filters` | `64` | Base number of convolutional filters |
| `num_blocks` | `4` | Number of convolutional blocks |
| `use_batch_norm` | `True` | Enable batch normalization |
| `dropout_rate` | `0.3` | Dropout rate for regularization |

### DeepfakeConfig

| Parameter | Default | Description |
|-----------|---------|-------------|
| `faceswap_threshold` | `0.65` | Face swap detection threshold |
| `expression_threshold` | `0.6` | Expression manipulation threshold |
| `use_temporal_analysis` | `True` | Enable temporal analysis for video |
| `min_frames_for_temporal` | `10` | Minimum frames for temporal analysis |

### VEIDIntegrationConfig

| Parameter | Default | Description |
|-----------|---------|-------------|
| `gan_detection_weight` | `0.30` | Weight for discriminator score |
| `deepfake_weight` | `0.35` | Weight for deepfake score |
| `artifact_weight` | `0.20` | Weight for artifact score |
| `synthetic_detected_penalty` | `5000` | VEID penalty (basis points) |
| `high_confidence_penalty` | `10000` | Full rejection penalty |

## Reason Codes

The module provides detailed reason codes for audit trails:

### Success Codes
- `IMAGE_AUTHENTIC`: Image verified as authentic
- `HIGH_CONFIDENCE_REAL`: High confidence real image
- `ALL_CHECKS_PASSED`: All detection checks passed

### Detection Codes
- `GAN_DETECTED`: GAN-generated image detected
- `GAN_HIGH_CONFIDENCE`: High confidence GAN detection
- `DEEPFAKE_DETECTED`: Deepfake manipulation detected
- `FACESWAP_DETECTED`: Face swap detected
- `FREQUENCY_ANOMALY`: Unusual frequency patterns
- `CHECKERBOARD_ARTIFACT`: GAN checkerboard pattern

## VEID Scoring Integration

The module integrates with the on-chain VEID scoring model:

```python
# Get VEID-compatible record
record = result.to_veid_record()

# Record format:
# {
#     "gan_score": 8500,        # 0-10000, higher = more authentic
#     "is_synthetic": False,
#     "decision": "authentic",
#     "confidence": 85,          # 0-100
#     "veid_penalty": 0,         # Basis points penalty
#     "veid_adjusted_score": 10000,
#     "model_version": "1.0.0",
#     "model_hash": "abc123...",
#     "result_hash": "def456...",
#     "reason_codes": ["IMAGE_AUTHENTIC"]
# }
```

## Testing

Run tests with pytest:

```bash
# Run all tests
pytest ml/gan_detection/tests/ -v

# Run specific test file
pytest ml/gan_detection/tests/test_detector.py -v

# Run with coverage
pytest ml/gan_detection/tests/ --cov=ml/gan_detection
```

## Architecture

```
ml/gan_detection/
├── __init__.py              # Public API exports
├── config.py                # Configuration classes
├── detector.py              # Main GANDetector class
├── discriminator.py         # CNN discriminator
├── deepfake_detection.py    # Deepfake detection
├── artifact_analysis.py     # Artifact analysis
├── reason_codes.py          # Reason codes for audit
├── README.md                # This file
├── requirements.txt         # Module dependencies
└── tests/
    ├── __init__.py
    ├── conftest.py          # Test fixtures
    ├── test_config.py
    ├── test_detector.py
    ├── test_discriminator.py
    ├── test_deepfake_detection.py
    ├── test_artifact_analysis.py
    └── test_reason_codes.py
```

## Security Considerations

- **No Sensitive Data Logging**: The module never logs raw image data
- **Deterministic Execution**: All operations are deterministic for blockchain consensus
- **CPU-Only Mode**: GPU is disabled by default to ensure reproducibility
- **Hash Verification**: Model and result hashes enable verification across nodes

## Performance

| Mode | Single Image | 20-Frame Sequence |
|------|--------------|-------------------|
| Fast | ~50ms | ~200ms |
| Full | ~150ms | ~500ms |
| Accurate | ~300ms | ~1000ms |

*Benchmarks on Intel Core i7, single thread*

## License

See project LICENSE file.
