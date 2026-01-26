"""
VirtEngine OCR Extraction Pipeline

This module provides OCR extraction for identity documents using Tesseract.
It supports:
- Precise ROI cropping with sub-pixel accuracy
- Tesseract OCR with configurable character whitelists
- Post-processing for common OCR error correction
- Structured field parsing for various document types
- Secure field hashing for on-chain storage (never plaintext PII)

Usage:
    from ml.ocr_extraction import (
        OCRExtractionPipeline,
        OCRExtractionConfig,
        OCRExtractionResult,
        ParsedField,
    )
    
    config = OCRExtractionConfig()
    pipeline = OCRExtractionPipeline(config)
    result = pipeline.extract(image, rois, document_type="id_card")
    
    for field_name, field in result.fields.items():
        print(f"{field_name}: confidence={field.confidence}")
        print(f"Hash: {result.field_hashes[field_name]}")
"""

from ml.ocr_extraction.config import (
    OCRExtractionConfig,
    TesseractConfig,
    CropperConfig,
    PostProcessingConfig,
    FieldParserConfig,
    HashingConfig,
    DocumentType,
)
from ml.ocr_extraction.roi_cropper import ROICropper
from ml.ocr_extraction.tesseract_wrapper import (
    TesseractOCR,
    OCRResult,
    DetailedOCRResult,
    CharacterDetail,
)
from ml.ocr_extraction.postprocessing import OCRPostProcessor
from ml.ocr_extraction.field_parser import (
    DocumentFieldParser,
    ParsedField,
    ValidationStatus,
)
from ml.ocr_extraction.field_hasher import FieldHasher
from ml.ocr_extraction.pipeline import OCRExtractionPipeline, OCRExtractionResult

__version__ = "1.0.0"
__all__ = [
    # Config
    "OCRExtractionConfig",
    "TesseractConfig",
    "CropperConfig",
    "PostProcessingConfig",
    "FieldParserConfig",
    "HashingConfig",
    "DocumentType",
    # ROI Cropper
    "ROICropper",
    # Tesseract
    "TesseractOCR",
    "OCRResult",
    "DetailedOCRResult",
    "CharacterDetail",
    # Post-processing
    "OCRPostProcessor",
    # Field Parsing
    "DocumentFieldParser",
    "ParsedField",
    "ValidationStatus",
    # Hashing
    "FieldHasher",
    # Pipeline
    "OCRExtractionPipeline",
    "OCRExtractionResult",
]
