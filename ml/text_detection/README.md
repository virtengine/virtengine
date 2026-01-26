# VirtEngine Text Detection Pipeline

Text region of interest (ROI) detection for identity documents using CRAFT (Character-Region Awareness for Text detection).

## Overview

This module provides text detection capabilities for the VirtEngine identity verification pipeline. It identifies text regions in preprocessed document images and produces structured, versioned ROI outputs suitable for OCR processing.

## Features

- **CRAFT Integration**: Character-level text detection with region and affinity scores
- **Hierarchical ROIs**: Character → Word → Line grouping
- **Post-processing**: Thresholding, NMS, and merging operations
- **Reproducibility**: Deterministic execution with version tracking
- **Configurable**: Adjustable thresholds and processing parameters

## Installation

```bash
pip install -r requirements.txt
```

## Quick Start

```python
from ml.text_detection import (
    TextDetectionPipeline,
    TextDetectionConfig,
)

# Initialize pipeline with default config
config = TextDetectionConfig()
pipeline = TextDetectionPipeline(config)

# Detect text regions in a document image
result = pipeline.detect(preprocessed_image)

# Access detected ROIs
for roi in result.word_rois:
    print(f"Word at {roi.bounding_box} with confidence {roi.confidence:.2f}")

# Check processing metadata
print(f"Model version: {result.model_version}")
print(f"Processing time: {result.processing_time_ms:.1f}ms")
print(f"Image hash: {result.image_hash}")
```

## Architecture

### Components

1. **CRAFTDetector** (`craft_detector.py`)
   - Wraps the CRAFT model for text detection
   - Produces region and affinity score maps
   - Extracts character-level bounding boxes

2. **TextPostProcessor** (`postprocessing.py`)
   - Score thresholding
   - Non-maximum suppression (NMS)
   - Character → Word merging
   - Word → Line grouping

3. **TextDetectionPipeline** (`pipeline.py`)
   - Orchestrates detection and post-processing
   - Produces versioned, reproducible outputs

### Data Types

- **TextROI**: Individual text region with bounding box, polygon, and scores
- **TextDetectionResult**: Complete detection output with metadata

## Configuration

### CRAFT Settings

```python
from ml.text_detection import CRAFTConfig

craft_config = CRAFTConfig(
    text_threshold=0.7,      # Character detection threshold
    link_threshold=0.4,      # Affinity threshold for linking
    canvas_size=1280,        # Max image dimension
    deterministic=True,      # Enable deterministic mode
)
```

### Post-processing Settings

```python
from ml.text_detection import PostProcessingConfig

postproc_config = PostProcessingConfig(
    nms_iou_threshold=0.5,   # NMS overlap threshold
    min_confidence=0.5,      # Minimum confidence to keep
    min_char_width=5,        # Minimum character size
)
```

## ROI Types

The pipeline produces three levels of text ROIs:

| Type | Description | Use Case |
|------|-------------|----------|
| CHARACTER | Individual character boxes | Fine-grained analysis |
| WORD | Merged character groups | OCR input |
| LINE | Grouped word sequences | Document structure |

## Reproducibility

All outputs include versioning information for audit and reproducibility:

```python
result = pipeline.detect(image)

print(result.model_version)      # CRAFT model version
print(result.suite_version)      # Detection suite version
print(result.thresholds_used)    # Thresholds applied
print(result.image_hash)         # SHA256 of input image
```

## Integration with Document Pipeline

This module is designed to work with the document preprocessing pipeline:

```python
from ml.document_preprocessing import DocumentPreprocessingPipeline
from ml.text_detection import TextDetectionPipeline

# Preprocess document
doc_pipeline = DocumentPreprocessingPipeline()
preprocessed = doc_pipeline.process(raw_image)

# Detect text regions
text_pipeline = TextDetectionPipeline()
text_result = text_pipeline.detect(preprocessed.normalized_image)

# Pass ROIs to OCR (VE-215)
for roi in text_result.word_rois:
    # Extract region for OCR processing
    x, y, w, h = roi.bounding_box.to_int_xyxy()
    word_region = preprocessed.normalized_image[y:y+h, x:x+w]
```

## Testing

```bash
# Run all tests
pytest ml/text_detection/tests/ -v

# Run with coverage
pytest ml/text_detection/tests/ --cov=ml/text_detection --cov-report=html
```

## Dependencies

- PyTorch 2.0.1
- craft-text-detector 0.4.3
- OpenCV 4.8.0
- NumPy 1.24.3

## Version History

- **1.0.0** (2026-01-24): Initial release with CRAFT integration

## Related Tasks

- VE-213: Document preprocessing (dependency)
- VE-215: OCR extraction (consumer)
- VE-216: Face extraction (parallel)
