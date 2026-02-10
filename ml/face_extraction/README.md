# Face Extraction Pipeline

This module provides U-Net-based face extraction from identity documents for the VirtEngine VEID identity verification system.

## Overview

The face extraction pipeline segments and extracts face regions from ID documents using a U-Net deep learning model, then optionally derives embeddings for comparison with selfie photos.

### Key Features

- **U-Net Segmentation**: Pixel-level face region detection
- **Mask Post-Processing**: Morphological operations for clean masks
- **Face Cropping**: Bounding box extraction and face region cropping
- **Data Minimization**: By default, face images are not retained - only embeddings
- **Deterministic Execution**: Consistent results for blockchain consensus

## Architecture

```
┌──────────────────┐    ┌─────────────────┐    ┌────────────────┐
│  Document Image  │───▶│ U-Net Segmentor │───▶│ Segmentation   │
│                  │    │                 │    │ Mask           │
└──────────────────┘    └─────────────────┘    └───────┬────────┘
                                                       │
                                                       ▼
                                               ┌───────────────┐
                                               │ Mask Processor│
                                               │ - Threshold   │
                                               │ - Morphology  │
                                               │ - Components  │
                                               └───────┬───────┘
                                                       │
                                                       ▼
┌──────────────────┐    ┌─────────────────┐    ┌───────────────┐
│ Face Embedding   │◀───│  Face Cropper   │◀───│ Cleaned Mask  │
│ (VE-211)         │    │                 │    │               │
└────────┬─────────┘    └─────────────────┘    └───────────────┘
         │
         ▼
┌──────────────────┐
│ Embedding Result │
│ - vector         │
│ - hash           │
│ - confidence     │
└──────────────────┘
```

## Usage

### Basic Usage (Data Minimization)

```python
from ml.face_extraction import FaceExtractionPipeline, FaceExtractionConfig

# Initialize pipeline
config = FaceExtractionConfig()
pipeline = FaceExtractionPipeline(config)

# Extract embedding only (recommended - data minimization)
result = pipeline.extract_embedding(document_image)

if result.success:
    embedding = result.embedding
    embedding_hash = result.embedding_hash
    confidence = result.confidence
```

### Extract Face Image (When Explicitly Needed)

```python
# Only use when face image is explicitly required
result = pipeline.extract_face(document_image)

if result.success:
    face_image = result.face_image
    bounding_box = result.bounding_box
    confidence = result.confidence
```

### Full Pipeline with Intermediate Results

```python
config = FaceExtractionConfig(return_intermediate_results=True)
pipeline = FaceExtractionPipeline(config)

result = pipeline.extract_full(document_image)

if result.success:
    print(f"Stages completed: {result.stages_completed}")
    print(f"Processing time: {result.processing_time_ms:.1f}ms")
```

## Components

### UNetFaceSegmentor

Segments face regions using a U-Net architecture:

```python
from ml.face_extraction import UNetFaceSegmentor, UNetConfig

config = UNetConfig(input_size=(256, 256))
segmentor = UNetFaceSegmentor(config)

# Get segmentation mask
mask = segmentor.segment(document_image)

# Get detailed result
result = segmentor.segment_with_details(document_image)
print(f"Confidence: {result.mean_confidence:.2f}")
print(f"Model hash: {result.model_hash}")
```

### MaskProcessor

Post-processes segmentation masks:

```python
from ml.face_extraction import MaskProcessor, MaskProcessingConfig

config = MaskProcessingConfig(
    threshold=0.5,
    apply_morphology=True,
    use_largest_component=True
)
processor = MaskProcessor(config)

# Full processing
result = processor.process(raw_mask)
cleaned_mask = result.processed_mask
```

### FaceCropper

Extracts face regions from masks:

```python
from ml.face_extraction import FaceCropper, CropperConfig

config = CropperConfig(
    margin=0.15,
    output_size=(224, 224)
)
cropper = FaceCropper(config)

# Extract bounding box
bbox = cropper.extract_bounding_box(mask)

# Crop face
face_image = cropper.crop_face(document_image, bbox)
```

### DocumentFaceEmbeddingExtractor

Extracts embeddings with data minimization:

```python
from ml.face_extraction import DocumentFaceEmbeddingExtractor

extractor = DocumentFaceEmbeddingExtractor()

# Extract embedding (face image is discarded)
result = extractor.extract_embedding_only(face_image)
```

## Configuration

### FaceExtractionConfig

Main configuration class with nested configs:

```python
config = FaceExtractionConfig(
    # U-Net settings
    unet=UNetConfig(
        input_size=(256, 256),
        use_gpu=False,
        ensure_determinism=True
    ),
    
    # Mask processing
    mask=MaskProcessingConfig(
        threshold=0.5,
        apply_morphology=True
    ),
    
    # Face cropping
    cropper=CropperConfig(
        margin=0.15,
        output_size=(224, 224)
    ),
    
    # Embedding
    embedding=EmbeddingConfig(
        discard_face_image=True  # Data minimization
    ),
    
    # Pipeline settings
    fallback_to_detection=True,
    min_extraction_confidence=0.6
)
```

## Data Minimization Policy

This module follows privacy-by-design principles:

1. **Default behavior**: Only embeddings are returned, not face images
2. **Explicit consent**: Face images returned only when `return_face_image=True`
3. **Immediate discard**: Face images are discarded after embedding computation
4. **No storage**: Face images are never written to disk by the pipeline

## Testing

Run tests with pytest:

```bash
cd ml/face_extraction
pytest tests/ -v
```

Run with coverage:

```bash
pytest tests/ -v --cov=. --cov-report=html
```

## Dependencies

See `requirements.txt` for dependencies. Key packages:

- TensorFlow 2.13.0 (U-Net model)
- OpenCV 4.8.0 (image processing)
- NumPy 1.24.3 (numerical operations)
- DeepFace 0.0.79 (embedding extraction)

## Related Modules

- `ml.facial_verification` (VE-211): Face embedding and verification
- `ml.document_preprocessing` (VE-213): Document preprocessing

## Version

1.0.0 (VE-216)
