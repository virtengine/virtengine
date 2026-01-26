# VE-924: Autoencoder Anomaly Detection

Autoencoder-based anomaly detection for identity verification in the VirtEngine VEID system.

## Overview

This module provides anomaly detection capabilities using autoencoders to identify potentially fraudulent or manipulated identity documents and selfie images. The autoencoder learns to reconstruct normal identity verification inputs, and anomalies are detected when reconstruction error is high.

## Architecture

### Autoencoder Model

```
Input Image (128x128x3)
    │
    ▼
┌─────────────────────┐
│  Convolutional      │
│  Encoder            │
│  ──────────────     │
│  Conv 32 → 64 →     │
│  128 → 256          │
│  + BatchNorm        │
│  + LeakyReLU        │
└─────────────────────┘
    │
    ▼
┌─────────────────────┐
│  Latent Space       │
│  (128 dimensions)   │
└─────────────────────┘
    │
    ▼
┌─────────────────────┐
│  Convolutional      │
│  Decoder            │
│  ──────────────     │
│  256 → 128 → 64 →   │
│  32 → 3             │
│  + BatchNorm        │
│  + Upsample         │
└─────────────────────┘
    │
    ▼
Reconstructed Image (128x128x3)
```

### Anomaly Detection Pipeline

1. **Encoding**: Image is encoded to latent representation
2. **Reconstruction**: Latent vector is decoded back to image
3. **Error Analysis**: Multiple metrics compare original vs reconstructed
4. **Latent Analysis**: Outlier detection in latent space
5. **Score Fusion**: Combined anomaly score for VEID integration

## Usage

### Basic Usage

```python
from ml.autoencoder_anomaly import AnomalyDetector, AutoencoderAnomalyConfig

# Create detector with default config
detector = AnomalyDetector()

# Analyze image (BGR format, uint8)
result = detector.detect(image)

# Check result
if result.is_anomaly:
    print(f"Anomaly detected!")
    print(f"Level: {result.anomaly_level.value}")
    print(f"Score: {result.overall_score:.3f}")
    print(f"Types: {result.detected_types}")
```

### Custom Configuration

```python
from ml.autoencoder_anomaly import (
    AnomalyDetector,
    AutoencoderAnomalyConfig,
    DetectionMode,
)

# Strict mode for higher sensitivity
config = AutoencoderAnomalyConfig.strict_mode()

# Or customize specific settings
config = AutoencoderAnomalyConfig()
config.scoring.normal_threshold = 0.25
config.scoring.anomaly_threshold = 0.65

detector = AnomalyDetector(config)
```

### VEID Integration

```python
# Get VEID scoring record
veid_record = result.to_veid_record()

# Fields:
# - anomaly_score: int (0-10000, higher = more normal)
# - is_anomaly: bool
# - decision: str ("normal", "suspicious", "anomaly", "uncertain")
# - veid_penalty: int (basis points)
# - veid_adjusted_score: int (10000 - penalty)
```

## Reconstruction Error Metrics

| Metric | Description | Threshold |
|--------|-------------|-----------|
| MSE | Mean Squared Error | 0.05 |
| MAE | Mean Absolute Error | 0.08 |
| SSIM | Structural Similarity | 0.85 |
| PSNR | Peak Signal-to-Noise Ratio | N/A |

## Latent Space Analysis

- **Euclidean Distance**: Distance from reference centroid
- **Mahalanobis Distance**: Distribution-aware distance
- **Z-Score Analysis**: Per-dimension outlier detection

## Anomaly Levels

| Level | Score Range | VEID Penalty |
|-------|-------------|--------------|
| NONE | < 0.3 | 0 bp |
| LOW | 0.3 - 0.5 | 500 bp |
| MEDIUM | 0.5 - 0.7 | 1500 bp |
| HIGH | 0.7 - 0.85 | 3500 bp |
| CRITICAL | > 0.85 | 7500 bp |

## Reason Codes

Common reason codes for audit trail:

- `NO_ANOMALY_DETECTED` - Input is normal
- `HIGH_RECONSTRUCTION_ERROR` - Overall reconstruction failed
- `MSE_ABOVIRTENGINE_THRESHOLD` - Mean squared error exceeded
- `SSIM_BELOW_THRESHOLD` - Structural similarity too low
- `LATENT_OUTLIER` - Latent representation is an outlier
- `MULTI_METRIC_ANOMALY` - Multiple metrics indicate anomaly

## Security Considerations

- **No Raw Data Logging**: Never log actual image data
- **Deterministic Execution**: Fixed seeds for reproducibility
- **CPU-Only Mode**: Consistent results across environments
- **Result Hashing**: Cryptographic verification of outputs

## Files

- `config.py` - Configuration dataclasses
- `autoencoder.py` - Encoder/decoder architecture
- `anomaly_scorer.py` - Reconstruction error and scoring
- `detector.py` - Main detector pipeline
- `reason_codes.py` - Audit trail codes
- `tests/` - Comprehensive test suite

## Dependencies

- numpy

## Version

1.0.0 (VE-924)
