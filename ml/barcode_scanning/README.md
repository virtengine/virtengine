# VirtEngine Barcode Scanning Module

## Overview

This module provides barcode and MRZ (Machine Readable Zone) scanning capabilities
for identity document validation in the VirtEngine VEID (Verified Electronic Identity)
system.

## Features

### PDF417 Barcode Parsing
- Full AAMVA (American Association of Motor Vehicle Administrators) standard support
- Driver license data extraction (name, DOB, address, document number, etc.)
- Versions 1-10 compatibility
- Automatic date format normalization

### MRZ Parsing
- ICAO Doc 9303 compliant parsing
- TD1 format (3 lines × 30 chars) - ID cards
- TD2 format (2 lines × 36 chars) - older ID cards, visas
- TD3 format (2 lines × 44 chars) - passports
- Check digit validation
- Automatic OCR error correction (O/0, I/1)

### Cross-Validation
- Field-by-field comparison with OCR data
- Fuzzy matching for names (Levenshtein distance)
- Flexible date format matching
- Weighted scoring based on field importance
- Required field enforcement

### VEID Scoring Integration
- Contributes to overall identity trust score
- Bonus for valid check digits
- Cross-validation score contribution
- Configurable scoring weights

## Security

⚠️ **CRITICAL**: This module handles Personally Identifiable Information (PII).

- **Never log raw barcode/MRZ data**
- **Never store plaintext PII on-chain**
- Use `to_chain_data()` methods for blockchain storage
- All identity fields are hashed with salt before on-chain storage
- Clear sensitive data from memory after processing

## Usage

### Basic Scanning

```python
from ml.barcode_scanning import BarcodeScanningPipeline, BarcodeScanningConfig

# Initialize pipeline
config = BarcodeScanningConfig()
pipeline = BarcodeScanningPipeline(config)

# Scan an image
result = pipeline.scan(
    image=document_image,
    document_category=DocumentCategory.DRIVERS_LICENSE,
    salt="unique-identity-salt",
)

# Check results
if result.success and result.barcode_detected:
    print(f"Barcode score: {result.barcode_score}")
    print(f"VEID contribution: {result.veid_contribution}")
    
    # Get on-chain safe data
    chain_data = result.to_chain_data()
```

### MRZ Parsing

```python
from ml.barcode_scanning import MRZParser

parser = MRZParser()

# Parse from text
mrz_text = """
P<UTOERIKSSON<<ANNA<MARIA<<<<<<<<<<<<<<<<<<<
L898902C36UTO7408122F1204159ZE184226B<<<<<10
"""

result = parser.parse(mrz_text)

if result.success:
    print(f"Name: {result.full_name}")
    print(f"DOB: {result.date_of_birth}")
    print(f"Check digits valid: {result.all_check_digits_valid}")
```

### Cross-Validation

```python
from ml.barcode_scanning import CrossValidator

validator = CrossValidator()

# Compare barcode data with OCR data
cross_result = validator.validate(
    barcode_data=result.identity_fields,
    ocr_data=ocr_extraction_result.fields,
)

if cross_result.is_valid:
    print(f"Cross-validation score: {cross_result.score}")
    print(f"Exact matches: {cross_result.exact_matches}")
```

## Configuration

### Scoring Configuration

```python
from ml.barcode_scanning.config import ScoringConfig

scoring = ScoringConfig(
    barcode_read_score=0.15,      # Base score for successful read
    check_digit_bonus=0.05,       # Bonus for valid check digits
    cross_validation_weight=0.20, # Weight for cross-validation
    max_score_contribution=0.40,  # Max VEID contribution
)
```

### Cross-Validation Configuration

```python
from ml.barcode_scanning.config import CrossValidationConfig

config = CrossValidationConfig(
    string_similarity_threshold=0.8,  # Minimum similarity for fuzzy match
    enable_fuzzy_matching=True,       # Enable name fuzzy matching
    max_edit_distance=2,              # Max Levenshtein distance
    required_fields={"date_of_birth"}, # Fields that must match
)
```

## API Reference

### BarcodeScanningPipeline

Main entry point for barcode scanning.

**Methods:**
- `scan(image, document_category, salt, ocr_data)` - Full scanning pipeline
- `scan_for_pdf417(image, salt)` - Scan specifically for driver license barcodes
- `scan_for_mrz(image, salt)` - Scan specifically for passport MRZ
- `parse_mrz_text(mrz_text, salt)` - Parse MRZ from text directly

### PDF417Parser

Parser for PDF417 barcodes on driver licenses.

**Methods:**
- `scan_image(image)` - Detect PDF417 barcodes in image
- `parse(raw_data)` - Parse raw barcode bytes to structured data
- `hash_fields(result, salt)` - Generate hashed field values

### MRZParser

Parser for ICAO Machine Readable Zone.

**Methods:**
- `detect_format(lines)` - Detect MRZ format from text lines
- `parse(mrz_text)` - Parse MRZ text to structured data
- `parse_from_image(image)` - Detect and parse MRZ from image
- `hash_fields(result, salt)` - Generate hashed field values

### CrossValidator

Cross-validator for barcode vs OCR data.

**Methods:**
- `validate(barcode_data, ocr_data)` - Compare and validate data sources

## Testing

Run unit tests:

```bash
cd ml/barcode_scanning
pytest tests/ -v
```

## Dependencies

- `pyzbar` - PDF417 barcode detection
- `opencv-python-headless` - Image processing
- `pytesseract` - MRZ OCR (optional, for image-based MRZ detection)
- `numpy` - Array operations
- `Pillow` - Image handling

## License

Copyright © 2026 VirtEngine. All rights reserved.
