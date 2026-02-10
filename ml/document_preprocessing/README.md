# VirtEngine Document Preprocessing Pipeline

ID document preprocessing for OCR and face extraction. This module provides:

- **Standardization**: Convert to consistent format (PNG), resize to target resolution, ensure RGB
- **Enhancement**: CLAHE histogram equalization, brightness/contrast normalization, sharpening
- **Noise Reduction**: Gaussian blur, median blur, bilateral filtering
- **Orientation Detection**: Detect and correct 0°/90°/180°/270° rotations using text/face scoring
- **Perspective Correction**: Corner detection and four-point transform for distortion correction

## Installation

```bash
pip install -r requirements.txt
```

## Usage

```python
from ml.document_preprocessing import (
    DocumentPreprocessingPipeline,
    DocumentConfig,
)

# Create pipeline with default config
config = DocumentConfig()
pipeline = DocumentPreprocessingPipeline(config)

# Process document image
import cv2
image = cv2.imread("id_card.jpg")
result = pipeline.process(image)

if result.success:
    normalized = result.normalized_image
    print(f"Processed in {result.processing_time_ms:.2f}ms")
    print(f"Rotation applied: {result.rotation_applied}°")
    print(f"Perspective corrected: {result.perspective_corrected}")
    print(f"Enhancements: {result.enhancements_applied}")
```

## Configuration

```python
from ml.document_preprocessing.config import (
    DocumentConfig,
    StandardizationConfig,
    EnhancementConfig,
)

# Custom configuration
config = DocumentConfig()

# Standardization settings
config.standardization.target_width = 1024
config.standardization.target_height = 768
config.standardization.maintain_aspect_ratio = True

# Enhancement settings
config.enhancement.apply_clahe = True
config.enhancement.clahe_clip_limit = 2.0
config.enhancement.apply_sharpening = True

# Preset configurations for document types
id_card_config = DocumentConfig.for_id_card()
passport_config = DocumentConfig.for_passport()
drivers_license_config = DocumentConfig.for_drivers_license()
```

## Pipeline Steps

1. **Standardize** - Convert to RGB, resize to target resolution
2. **Orientation Detection** - Detect 90° rotation using text/face scoring
3. **Perspective Correction** - Detect corners, apply four-point transform
4. **CLAHE Enhancement** - Improve local contrast
5. **Noise Reduction** - Bilateral filtering for edge preservation
6. **Sharpening** - Unsharp masking for text clarity

## Testing

```bash
pytest ml/document_preprocessing/tests/ -v
```

## Security Note

The `normalized_document_image` output is intended for processing only.
It should **NOT** be stored in plaintext on-chain. Only encrypted versions
or derived features (embeddings, hashes) should be persisted.
