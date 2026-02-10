# OCR Extraction Pipeline

VirtEngine OCR extraction module for identity document field extraction using Tesseract OCR.

## Overview

This module provides:
- **Precise ROI cropping** with sub-pixel accuracy for rotated text regions
- **Tesseract OCR integration** with configurable character whitelists and page segmentation modes
- **Post-processing** for common OCR error correction (O/0, I/1/l, S/5 confusions)
- **Structured field parsing** for ID cards, passports, and driver's licenses
- **Secure field hashing** for on-chain storage (never plaintext PII)

## Installation

### Prerequisites

1. **Tesseract OCR** must be installed on your system:
   - Windows: Download from [UB-Mannheim](https://github.com/UB-Mannheim/tesseract/wiki)
   - macOS: `brew install tesseract`
   - Ubuntu: `sudo apt install tesseract-ocr`

2. Install Python dependencies:
   ```bash
   pip install -r requirements.txt
   ```

## Usage

### Basic Usage

```python
from ml.text_detection import TextDetectionPipeline
from ml.ocr_extraction import (
    OCRExtractionPipeline,
    OCRExtractionConfig,
    DocumentType,
)

# Load document image
image = cv2.imread("document.jpg")

# Detect text regions (from VE-214)
text_pipeline = TextDetectionPipeline()
detection_result = text_pipeline.detect(image)

# Extract text and parse fields
ocr_pipeline = OCRExtractionPipeline()
result = ocr_pipeline.extract(
    image,
    detection_result.rois,
    document_type=DocumentType.ID_CARD
)

# Access parsed fields
for field_name, field in result.fields.items():
    print(f"{field_name}: {field.value} (confidence: {field.confidence})")

# Get chain-safe data (hashes only, no PII)
chain_data = result.to_chain_data()
print(f"Identity hash: {chain_data['identity_hash']}")
```

### Document Types

Supported document types with specialized parsing:

- `DocumentType.ID_CARD` - National ID cards
- `DocumentType.PASSPORT` - Passports with MRZ support
- `DocumentType.DRIVERS_LICENSE` - Driver's licenses
- `DocumentType.UNKNOWN` - Generic document parsing

### Configuration

```python
from ml.ocr_extraction import (
    OCRExtractionConfig,
    TesseractConfig,
    CropperConfig,
    HashingConfig,
    PageSegmentationMode,
    ThresholdingMethod,
)

config = OCRExtractionConfig(
    # Tesseract settings
    tesseract=TesseractConfig(
        lang="eng",
        psm=PageSegmentationMode.SINGLE_BLOCK,
        whitelist="ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
    ),
    
    # ROI cropping settings
    cropper=CropperConfig(
        margin_pixels=5,
        thresholding=ThresholdingMethod.OTSU,
        scale_to_height=32,
    ),
    
    # Hashing settings (for production)
    hashing=HashingConfig(
        salt="your-secret-salt",  # Load from env
        normalize_before_hash=True,
        uppercase_before_hash=True,
    ),
)

pipeline = OCRExtractionPipeline(config)
```

### MRZ (Machine Readable Zone) Support

Passport MRZ is automatically detected and parsed:

```python
result = ocr_pipeline.extract(
    passport_image,
    rois,
    document_type=DocumentType.PASSPORT
)

# MRZ fields are extracted automatically
print(f"Surname: {result.fields['surname'].value}")
print(f"Given names: {result.fields['given_names'].value}")
print(f"Document number: {result.fields['document_number'].value}")
```

### Secure Hashing

**CRITICAL**: Never store plaintext PII on-chain. Use hashed values:

```python
from ml.ocr_extraction import FieldHasher

# For production, use a secret salt
hasher = FieldHasher(HashingConfig(salt="secret-from-env"))

# Hash individual field
field_hash = hasher.hash_field("surname", "SMITH")

# Hash all fields
all_hashes = hasher.hash_fields(result.fields)

# Create combined identity hash
identity_hash = hasher.create_derived_identity_hash(result.fields)

# Verify a field value against stored hash
is_valid = hasher.verify_field_hash("surname", "SMITH", stored_hash)
```

## Module Structure

```
ml/ocr_extraction/
├── __init__.py           # Package exports
├── config.py             # Configuration classes
├── roi_cropper.py        # Precise ROI cropping
├── tesseract_wrapper.py  # Tesseract OCR interface
├── postprocessing.py     # OCR error correction
├── field_parser.py       # Structured field extraction
├── field_hasher.py       # Secure field hashing
├── pipeline.py           # Main OCR pipeline
├── requirements.txt      # Dependencies
├── README.md             # This file
└── tests/
    ├── __init__.py
    ├── conftest.py       # Test fixtures
    ├── test_cropper.py
    ├── test_tesseract.py
    ├── test_postprocessing.py
    ├── test_field_parser.py
    └── test_pipeline.py
```

## Extracted Fields

Common fields extracted by document type:

### ID Card
- `surname` - Family name
- `given_names` - First/middle names
- `date_of_birth` - DOB
- `document_number` - ID number
- `expiry_date` - Expiration date
- `nationality` - Nationality

### Passport
- All ID card fields plus:
- `country_code` - Issuing country
- `sex` - Gender (M/F)

### Driver's License
- All ID card fields plus:
- `license_class` - License category
- `address` - Address (if present)

## Error Correction

The post-processor handles common OCR errors:

| Context | Confusion | Correction |
|---------|-----------|------------|
| Numeric | O → 0 | `12O45` → `12045` |
| Numeric | I/l → 1 | `l23` → `123` |
| Numeric | S → 5 | `1S0` → `150` |
| Alpha | 0 → O | `HELL0` → `HELLO` |
| General | Em-dash → hyphen | `2020—2021` → `2020-2021` |

## Security Considerations

1. **Never store plaintext PII on-chain**
   - Use `result.to_chain_data()` for blockchain storage
   - Only hashed values are included

2. **Use secure hashing salt**
   - Load salt from environment/secrets manager
   - Same salt must be used across all nodes for verification

3. **Confidence thresholds**
   - Low confidence fields (< 50%) should trigger manual review
   - Check `field.validation_status` before trusting values

## Running Tests

```bash
# Run all tests
pytest ml/ocr_extraction/tests/ -v

# Run with coverage
pytest ml/ocr_extraction/tests/ --cov=ml.ocr_extraction

# Skip Tesseract-dependent tests if not installed
pytest ml/ocr_extraction/tests/ -v -k "not test_tesseract"
```

## Integration with VirtEngine

This module integrates with:
- **VE-214**: Text ROI detection (CRAFT) - provides ROIs as input
- **VE-211**: Facial verification - combined identity scoring
- **VE-200/202**: VEID module - stores hashed results on-chain

## Version

- Package version: 1.0.0
- Tesseract compatibility: 4.x, 5.x
