"""
VirtEngine Barcode Scanning Module

This module provides barcode and MRZ scanning for identity document validation.
It supports:
- PDF417 barcode parsing for driver licenses (AAMVA standard)
- MRZ parsing for passports (ICAO TD1, TD2, TD3 formats)
- Cross-validation of barcode data against OCR extracted data
- Integration with VEID scoring pipeline

Usage:
    from ml.barcode_scanning import (
        BarcodeScanningPipeline,
        BarcodeScanningConfig,
        BarcodeScanningResult,
        PDF417Parser,
        MRZParser,
        CrossValidator,
    )
    
    config = BarcodeScanningConfig()
    pipeline = BarcodeScanningPipeline(config)
    result = pipeline.scan(image, document_type="drivers_license")
    
    # Access parsed data
    print(f"Name: {result.parsed_data.get('full_name')}")
    print(f"Validation score: {result.validation_score}")
    
    # Cross-validate with OCR data
    validator = CrossValidator()
    cross_result = validator.validate(result, ocr_extraction_result)
    print(f"Cross-validation score: {cross_result.score}")

Security Note:
    This module handles PII (Personally Identifiable Information).
    - Never log raw barcode/MRZ data
    - Use hashed values for on-chain storage
    - Clear sensitive data from memory when done
"""

from ml.barcode_scanning.config import (
    BarcodeScanningConfig,
    PDF417Config,
    MRZConfig,
    CrossValidationConfig,
    BarcodeType,
    MRZFormat,
)
from ml.barcode_scanning.pdf417_parser import (
    PDF417Parser,
    PDF417Data,
    AAMVAField,
)
from ml.barcode_scanning.mrz_parser import (
    MRZParser,
    MRZData,
    MRZLine,
    MRZCheckDigit,
)
from ml.barcode_scanning.cross_validator import (
    CrossValidator,
    CrossValidationResult,
    FieldMatch,
    MatchType,
)
from ml.barcode_scanning.pipeline import (
    BarcodeScanningPipeline,
    BarcodeScanningResult,
)

__all__ = [
    # Config
    "BarcodeScanningConfig",
    "PDF417Config",
    "MRZConfig",
    "CrossValidationConfig",
    "BarcodeType",
    "MRZFormat",
    # PDF417
    "PDF417Parser",
    "PDF417Data",
    "AAMVAField",
    # MRZ
    "MRZParser",
    "MRZData",
    "MRZLine",
    "MRZCheckDigit",
    # Cross-validation
    "CrossValidator",
    "CrossValidationResult",
    "FieldMatch",
    "MatchType",
    # Pipeline
    "BarcodeScanningPipeline",
    "BarcodeScanningResult",
]

__version__ = "1.0.0"
