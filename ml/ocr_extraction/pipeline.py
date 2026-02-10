"""
OCR Extraction Pipeline for document identity verification.

This module provides the main pipeline that combines:
- ROI cropping and preprocessing
- Tesseract OCR text recognition
- Post-processing and error correction
- Structured field parsing
- Secure field hashing for on-chain storage

The pipeline ensures no plaintext PII is stored on-chain.
"""

import time
import hashlib
from typing import Dict, List, Optional, Any, Tuple
from dataclasses import dataclass, field
import numpy as np

from ml.text_detection import TextROI
from ml.ocr_extraction.config import (
    OCRExtractionConfig,
    DocumentType,
    TesseractConfig,
)
from ml.ocr_extraction.roi_cropper import ROICropper, CropResult
from ml.ocr_extraction.tesseract_wrapper import TesseractOCR, OCRResult
from ml.ocr_extraction.postprocessing import OCRPostProcessor
from ml.ocr_extraction.field_parser import DocumentFieldParser, ParsedField
from ml.ocr_extraction.field_hasher import FieldHasher


@dataclass
class OCRExtractionResult:
    """
    Complete result from OCR extraction pipeline.
    
    Contains parsed fields, hashes for on-chain storage,
    and metadata for audit and verification.
    """
    # Parsed fields (NOT for on-chain storage)
    fields: Dict[str, ParsedField]
    
    # Hashed fields (safe for on-chain storage)
    field_hashes: Dict[str, str]
    
    # Combined identity hash (safe for on-chain storage)
    identity_hash: str
    
    # Raw OCR text (NOT for on-chain storage)
    raw_text: str
    
    # Overall confidence score
    confidence_score: float
    
    # Model/pipeline version for reproducibility
    model_version: str
    
    # Additional metadata
    document_type: DocumentType = DocumentType.UNKNOWN
    processing_time_ms: float = 0.0
    roi_count: int = 0
    successful_rois: int = 0
    
    # Per-ROI results (NOT for on-chain storage)
    roi_results: List[OCRResult] = field(default_factory=list)
    
    # Success flag
    success: bool = True
    error_message: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "success": self.success,
            "document_type": self.document_type.value,
            "confidence_score": self.confidence_score,
            "model_version": self.model_version,
            "processing_time_ms": self.processing_time_ms,
            "roi_count": self.roi_count,
            "successful_rois": self.successful_rois,
            "error_message": self.error_message,
            "fields": {
                name: field.to_dict()
                for name, field in self.fields.items()
            },
            "field_hashes": self.field_hashes,
            "identity_hash": self.identity_hash,
        }
    
    def to_chain_data(self) -> Dict[str, Any]:
        """
        Get data safe for on-chain storage.
        
        Returns only hashed values, never plaintext PII.
        """
        return {
            "identity_hash": self.identity_hash,
            "field_hashes": self.field_hashes,
            "confidence_score": self.confidence_score,
            "model_version": self.model_version,
            "document_type": self.document_type.value,
        }
    
    @property
    def has_required_fields(self) -> bool:
        """Check if required fields were extracted."""
        required = ["full_name", "date_of_birth"]
        # Check for name (could be full_name, or surname + given_names)
        has_name = (
            "full_name" in self.fields or
            ("surname" in self.fields and "given_names" in self.fields)
        )
        has_dob = "date_of_birth" in self.fields
        
        return has_name and has_dob


class OCRExtractionPipeline:
    """
    Main OCR extraction pipeline.
    
    Orchestrates the full extraction process from text ROIs
    to hashed field values ready for on-chain storage.
    """
    
    VERSION = "1.0.0"
    
    def __init__(self, config: Optional[OCRExtractionConfig] = None):
        """
        Initialize the OCR extraction pipeline.
        
        Args:
            config: Pipeline configuration. Uses defaults if None.
        """
        self.config = config or OCRExtractionConfig()
        
        # Initialize components
        self.cropper = ROICropper(self.config.cropper)
        self.ocr = TesseractOCR(self.config.tesseract)
        self.postprocessor = OCRPostProcessor(self.config.postprocessing)
        self.field_parser = DocumentFieldParser(self.config.field_parser)
        self.hasher = FieldHasher(self.config.hashing)
    
    def extract(
        self,
        image: np.ndarray,
        rois: List[TextROI],
        document_type: DocumentType = DocumentType.UNKNOWN
    ) -> OCRExtractionResult:
        """
        Extract structured fields from document image.
        
        This is the main entry point for the OCR pipeline.
        
        Args:
            image: Document image (BGR or grayscale)
            rois: Text ROIs from text detection
            document_type: Type of document for specialized parsing
            
        Returns:
            OCRExtractionResult with parsed and hashed fields
        """
        start_time = time.time()
        
        try:
            # Limit ROI count
            rois = rois[:self.config.max_rois_per_image]
            
            # Step 1: Crop and preprocess ROIs
            crop_results = self.cropper.crop_all_rois(image, rois)
            
            # Step 2: Run OCR on each crop
            ocr_results = self._run_ocr_batch(crop_results, document_type)
            
            # Step 3: Post-process OCR results
            processed_results = self._postprocess_results(ocr_results)
            
            # Step 4: Parse structured fields
            fields = self.field_parser.parse(processed_results, document_type)
            
            # Step 5: Hash fields for on-chain storage
            field_hashes = self.hasher.hash_fields(fields)
            identity_hash = self.hasher.create_derived_identity_hash(fields)
            
            # Calculate overall confidence
            confidence_score = self._calculate_confidence(fields, processed_results)
            
            # Combine raw text
            raw_text = "\n".join(r.text for r in processed_results if r.text)
            
            processing_time = (time.time() - start_time) * 1000
            
            return OCRExtractionResult(
                fields=fields,
                field_hashes=field_hashes,
                identity_hash=identity_hash,
                raw_text=raw_text,
                confidence_score=confidence_score,
                model_version=self.VERSION,
                document_type=document_type,
                processing_time_ms=processing_time,
                roi_count=len(rois),
                successful_rois=sum(1 for r in processed_results if r.text),
                roi_results=processed_results,
                success=True,
            )
            
        except Exception as e:
            processing_time = (time.time() - start_time) * 1000
            
            return OCRExtractionResult(
                fields={},
                field_hashes={},
                identity_hash="",
                raw_text="",
                confidence_score=0.0,
                model_version=self.VERSION,
                document_type=document_type,
                processing_time_ms=processing_time,
                roi_count=len(rois),
                successful_rois=0,
                success=False,
                error_message=str(e),
            )
    
    def extract_with_auto_detection(
        self,
        image: np.ndarray,
        rois: List[TextROI]
    ) -> OCRExtractionResult:
        """
        Extract fields with automatic document type detection.
        
        Analyzes the OCR text to determine document type,
        then applies specialized parsing.
        
        Args:
            image: Document image
            rois: Text ROIs
            
        Returns:
            OCRExtractionResult
        """
        # First pass: get raw OCR text
        crop_results = self.cropper.crop_all_rois(image, rois[:20])
        ocr_results = self._run_ocr_batch(crop_results, DocumentType.UNKNOWN)
        
        # Detect document type
        document_type = self._detect_document_type(ocr_results)
        
        # Full extraction with detected type
        return self.extract(image, rois, document_type)
    
    def _run_ocr_batch(
        self,
        crop_results: List[CropResult],
        document_type: DocumentType
    ) -> List[OCRResult]:
        """Run OCR on batch of crop results."""
        results = []
        
        # Get appropriate Tesseract config
        tess_config = self.config.get_tesseract_config(document_type)
        
        for crop in crop_results:
            if not crop.success:
                # Skip failed crops
                results.append(OCRResult(
                    text="",
                    confidence=0.0,
                    word_confidences=[],
                    roi_id=crop.roi_id,
                ))
                continue
            
            try:
                result = self.ocr.recognize(
                    crop.processed_crop,
                    roi_id=crop.roi_id,
                    config_override=tess_config,
                )
                results.append(result)
            except Exception as e:
                results.append(OCRResult(
                    text="",
                    confidence=0.0,
                    word_confidences=[],
                    roi_id=crop.roi_id,
                ))
        
        return results
    
    def _postprocess_results(
        self,
        ocr_results: List[OCRResult]
    ) -> List[OCRResult]:
        """Post-process OCR results."""
        processed = []
        
        for result in ocr_results:
            if not result.text:
                processed.append(result)
                continue
            
            # Apply post-processing
            corrected_text = self.postprocessor.process(result.text)
            
            # Create new result with corrected text
            processed.append(OCRResult(
                text=corrected_text,
                confidence=result.confidence,
                word_confidences=result.word_confidences,
                roi_id=result.roi_id,
            ))
        
        return processed
    
    def _detect_document_type(
        self,
        ocr_results: List[OCRResult]
    ) -> DocumentType:
        """Detect document type from OCR text."""
        combined_text = " ".join(r.text.upper() for r in ocr_results if r.text)
        
        # Look for MRZ patterns (passport)
        if "P<" in combined_text or "<<<" in combined_text:
            return DocumentType.PASSPORT
        
        # Look for driver's license indicators
        dl_keywords = ["DRIVER", "LICENSE", "LICENCE", "DL", "DRIVING"]
        if any(kw in combined_text for kw in dl_keywords):
            return DocumentType.DRIVERS_LICENSE
        
        # Look for ID card indicators
        id_keywords = ["IDENTITY", "ID CARD", "NATIONAL ID", "CITIZEN"]
        if any(kw in combined_text for kw in id_keywords):
            return DocumentType.ID_CARD
        
        return DocumentType.UNKNOWN
    
    def _calculate_confidence(
        self,
        fields: Dict[str, ParsedField],
        ocr_results: List[OCRResult]
    ) -> float:
        """Calculate overall confidence score."""
        if not fields and not ocr_results:
            return 0.0
        
        scores = []
        
        # Field confidences (weighted higher)
        for field in fields.values():
            scores.append(field.confidence * 1.5)
        
        # OCR confidences
        for result in ocr_results:
            if result.text:
                scores.append(result.normalized_confidence)
        
        if not scores:
            return 0.0
        
        # Clamp to 0-1
        return min(1.0, sum(scores) / len(scores))
    
    def process_single_roi(
        self,
        image: np.ndarray,
        roi: TextROI,
        field_type: Optional[str] = None
    ) -> Tuple[str, float]:
        """
        Process a single ROI and return text with confidence.
        
        Useful for targeted field extraction when ROI purpose
        is already known.
        
        Args:
            image: Document image
            roi: Single text ROI
            field_type: Optional field type for specialized processing
            
        Returns:
            Tuple of (text, confidence)
        """
        # Crop and preprocess
        crop_result = self.cropper.crop_and_prepare(image, roi)
        
        if not crop_result.success:
            return ("", 0.0)
        
        # Run OCR
        ocr_result = self.ocr.recognize(
            crop_result.processed_crop,
            roi_id=roi.roi_id,
        )
        
        # Post-process
        processed_text = self.postprocessor.process(
            ocr_result.text,
            field_type=field_type,
        )
        
        return (processed_text, ocr_result.normalized_confidence)
    
    def hash_image(self, image: np.ndarray) -> str:
        """
        Compute SHA-256 hash of image for audit.
        
        Args:
            image: Input image
            
        Returns:
            Hex-encoded hash
        """
        return hashlib.sha256(image.tobytes()).hexdigest()


def create_default_pipeline() -> OCRExtractionPipeline:
    """Create pipeline with default configuration."""
    return OCRExtractionPipeline()


def create_production_pipeline(
    hashing_salt: str,
    tesseract_path: Optional[str] = None
) -> OCRExtractionPipeline:
    """
    Create pipeline configured for production use.
    
    Args:
        hashing_salt: Secret salt for field hashing
        tesseract_path: Optional path to Tesseract data
        
    Returns:
        Configured OCRExtractionPipeline
    """
    from ml.ocr_extraction.config import HashingConfig, TesseractConfig
    
    config = OCRExtractionConfig()
    
    # Configure hashing with salt
    config.hashing = HashingConfig(
        salt=hashing_salt,
        normalize_before_hash=True,
        uppercase_before_hash=True,
        include_field_name=True,
    )
    
    # Configure Tesseract path if provided
    if tesseract_path:
        config.tesseract = TesseractConfig(
            tessdata_path=tesseract_path,
        )
    
    return OCRExtractionPipeline(config)
