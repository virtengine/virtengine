"""
Barcode Scanning Pipeline for Identity Document Validation.

This module provides the main pipeline that combines:
- PDF417 barcode detection and parsing (driver licenses)
- MRZ detection and parsing (passports, ID cards)
- Cross-validation with OCR data
- VEID scoring integration

The pipeline ensures secure handling of PII and provides
deterministic results for blockchain verification.

Security Note:
    This pipeline handles sensitive PII.
    - No raw PII is logged
    - All values are hashed before on-chain storage
    - Memory is cleared after processing
"""

import time
import logging
import hashlib
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any, Union, Tuple, TYPE_CHECKING

if TYPE_CHECKING:
    import numpy as np

from ml.barcode_scanning.config import (
    BarcodeScanningConfig,
    BarcodeType,
    MRZFormat,
    DocumentCategory,
)
from ml.barcode_scanning.pdf417_parser import PDF417Parser, PDF417Data
from ml.barcode_scanning.mrz_parser import MRZParser, MRZData
from ml.barcode_scanning.cross_validator import (
    CrossValidator,
    CrossValidationResult,
    calculate_combined_identity_hash,
)

# Intentionally minimal logging - no PII
logger = logging.getLogger(__name__)


@dataclass
class BarcodeScanningResult:
    """
    Complete result from barcode scanning pipeline.
    
    Contains parsed data, validation results, and scoring
    contributions for VEID.
    """
    # Detection results
    barcode_detected: bool = False
    barcode_type: BarcodeType = BarcodeType.UNKNOWN
    mrz_detected: bool = False
    mrz_format: MRZFormat = MRZFormat.UNKNOWN
    
    # Parsed data (NOT for on-chain storage)
    pdf417_data: Optional[PDF417Data] = None
    mrz_data: Optional[MRZData] = None
    
    # Normalized identity fields (NOT for on-chain storage)
    identity_fields: Dict[str, str] = field(default_factory=dict)
    
    # Hashed fields (safe for on-chain storage)
    field_hashes: Dict[str, str] = field(default_factory=dict)
    
    # Combined identity hash (safe for on-chain)
    identity_hash: str = ""
    
    # Cross-validation results
    cross_validation: Optional[CrossValidationResult] = None
    
    # Scoring
    barcode_score: float = 0.0       # Score from barcode parsing
    mrz_score: float = 0.0           # Score from MRZ parsing
    validation_score: float = 0.0    # Combined validation score
    veid_contribution: float = 0.0   # Contribution to VEID score
    
    # Check digit validation
    check_digits_valid: bool = False
    
    # Metadata
    document_category: DocumentCategory = DocumentCategory.UNKNOWN
    model_version: str = ""
    processing_time_ms: float = 0.0
    
    # Success/error handling
    success: bool = True
    error_message: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "success": self.success,
            "barcode_detected": self.barcode_detected,
            "barcode_type": self.barcode_type.value,
            "mrz_detected": self.mrz_detected,
            "mrz_format": self.mrz_format.value,
            "document_category": self.document_category.value,
            "barcode_score": self.barcode_score,
            "mrz_score": self.mrz_score,
            "validation_score": self.validation_score,
            "veid_contribution": self.veid_contribution,
            "check_digits_valid": self.check_digits_valid,
            "processing_time_ms": self.processing_time_ms,
            "model_version": self.model_version,
            "error_message": self.error_message,
            "cross_validation": self.cross_validation.to_dict() if self.cross_validation else None,
        }
    
    def to_chain_data(self) -> Dict[str, Any]:
        """
        Get data safe for on-chain storage.
        
        Returns only hashed values and scores, never plaintext PII.
        """
        chain_data = {
            "identity_hash": self.identity_hash,
            "field_hashes": self.field_hashes,
            "barcode_score": self.barcode_score,
            "mrz_score": self.mrz_score,
            "validation_score": self.validation_score,
            "veid_contribution": self.veid_contribution,
            "check_digits_valid": self.check_digits_valid,
            "barcode_detected": self.barcode_detected,
            "mrz_detected": self.mrz_detected,
        }
        
        if self.cross_validation:
            chain_data["cross_validation"] = self.cross_validation.to_chain_data()
        
        return chain_data


class BarcodeScanningPipeline:
    """
    Main pipeline for barcode/MRZ scanning and validation.
    
    This class orchestrates:
    1. Barcode detection (PDF417, QR, etc.)
    2. MRZ detection and parsing
    3. Data extraction and normalization
    4. Cross-validation with OCR data
    5. VEID score contribution calculation
    """
    
    def __init__(self, config: Optional[BarcodeScanningConfig] = None):
        """
        Initialize the barcode scanning pipeline.
        
        Args:
            config: Pipeline configuration. Uses defaults if not provided.
        """
        self.config = config or BarcodeScanningConfig()
        
        # Validate configuration
        errors = self.config.validate()
        if errors:
            logger.warning(f"Configuration warnings: {errors}")
        
        # Initialize parsers
        self.pdf417_parser = PDF417Parser(self.config.pdf417)
        self.mrz_parser = MRZParser(self.config.mrz)
        self.cross_validator = CrossValidator(self.config.cross_validation)
    
    def scan(
        self,
        image: Any,
        document_category: DocumentCategory = DocumentCategory.UNKNOWN,
        salt: str = "",
        ocr_data: Optional[Dict[str, Any]] = None,
    ) -> BarcodeScanningResult:
        """
        Scan image for barcodes and MRZ, parse and validate.
        
        Args:
            image: Input image as numpy array (BGR or grayscale).
            document_category: Expected document type (optional).
            salt: Salt for hashing identity fields.
            ocr_data: OCR extraction results for cross-validation.
            
        Returns:
            BarcodeScanningResult with all parsing and validation results.
        """
        start_time = time.time()
        result = BarcodeScanningResult()
        result.model_version = self.config.model_version
        result.document_category = document_category
        
        try:
            # Preprocess image if needed
            if self.config.enable_preprocessing:
                image = self._preprocess_image(image)
            
            # Detect and parse barcodes
            if document_category in (DocumentCategory.DRIVERS_LICENSE, DocumentCategory.UNKNOWN):
                self._scan_pdf417(image, result)
            
            # Detect and parse MRZ
            if document_category in (DocumentCategory.PASSPORT, DocumentCategory.ID_CARD, DocumentCategory.UNKNOWN):
                self._scan_mrz(image, result)
            
            # Infer document category if not specified
            if document_category == DocumentCategory.UNKNOWN:
                result.document_category = self._infer_document_category(result)
            
            # Extract identity fields
            self._extract_identity_fields(result)
            
            # Generate hashes for on-chain storage
            if salt:
                self._generate_hashes(result, salt)
            
            # Cross-validate with OCR if provided
            if ocr_data:
                self._cross_validate(result, ocr_data)
            
            # Calculate scores
            self._calculate_scores(result)
            
            result.success = True
            
        except Exception as e:
            logger.error(f"Barcode scanning failed: {type(e).__name__}")
            result.success = False
            result.error_message = f"Scanning error: {type(e).__name__}"
        
        finally:
            result.processing_time_ms = (time.time() - start_time) * 1000
        
        return result
    
    def scan_for_pdf417(
        self,
        image: Any,
        salt: str = ""
    ) -> BarcodeScanningResult:
        """
        Scan specifically for PDF417 barcodes (driver licenses).
        
        Args:
            image: Input image as numpy array.
            salt: Salt for hashing.
            
        Returns:
            BarcodeScanningResult with PDF417 data.
        """
        return self.scan(
            image,
            document_category=DocumentCategory.DRIVERS_LICENSE,
            salt=salt,
        )
    
    def scan_for_mrz(
        self,
        image: Any,
        salt: str = ""
    ) -> BarcodeScanningResult:
        """
        Scan specifically for MRZ (passports, ID cards).
        
        Args:
            image: Input image as numpy array.
            salt: Salt for hashing.
            
        Returns:
            BarcodeScanningResult with MRZ data.
        """
        return self.scan(
            image,
            document_category=DocumentCategory.PASSPORT,
            salt=salt,
        )
    
    def parse_mrz_text(self, mrz_text: str, salt: str = "") -> BarcodeScanningResult:
        """
        Parse MRZ from text directly (without image processing).
        
        Args:
            mrz_text: Raw MRZ text (2-3 lines).
            salt: Salt for hashing.
            
        Returns:
            BarcodeScanningResult with parsed MRZ data.
        """
        result = BarcodeScanningResult()
        result.model_version = self.config.model_version
        result.document_category = DocumentCategory.PASSPORT
        
        start_time = time.time()
        
        try:
            mrz_data = self.mrz_parser.parse(mrz_text)
            
            if mrz_data.success:
                result.mrz_detected = True
                result.mrz_data = mrz_data
                result.mrz_format = mrz_data.format
                result.check_digits_valid = mrz_data.all_check_digits_valid
                
                self._extract_identity_fields(result)
                
                if salt:
                    self._generate_hashes(result, salt)
                
                self._calculate_scores(result)
            
            result.success = True
            
        except Exception as e:
            logger.error(f"MRZ parsing failed: {type(e).__name__}")
            result.success = False
            result.error_message = f"Parsing error: {type(e).__name__}"
        
        finally:
            result.processing_time_ms = (time.time() - start_time) * 1000
        
        return result
    
    def _preprocess_image(self, image: Any) -> Any:
        """Preprocess image for barcode detection."""
        try:
            import cv2
        except ImportError:
            return image
        
        # Convert to grayscale if needed
        if self.config.grayscale_conversion and len(image.shape) == 3:
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        else:
            gray = image if len(image.shape) == 2 else cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        
        # Enhance contrast
        if self.config.contrast_enhancement:
            clahe = cv2.createCLAHE(clipLimit=2.0, tileGridSize=(8, 8))
            gray = clahe.apply(gray)
        
        return gray
    
    def _scan_pdf417(self, image: Any, result: BarcodeScanningResult) -> None:
        """Scan for PDF417 barcodes."""
        try:
            barcodes = self.pdf417_parser.scan_image(image)
            
            if barcodes:
                # Parse the first (best) barcode
                raw_data, barcode_info = barcodes[0]
                pdf417_data = self.pdf417_parser.parse(raw_data)
                
                if pdf417_data.success:
                    result.barcode_detected = True
                    result.barcode_type = BarcodeType.PDF417
                    result.pdf417_data = pdf417_data
                    
                    logger.debug("PDF417 barcode parsed successfully")
                    
        except ImportError:
            logger.warning("pyzbar not available, PDF417 scanning disabled")
        except Exception as e:
            logger.debug(f"PDF417 scanning error: {type(e).__name__}")
    
    def _scan_mrz(self, image: Any, result: BarcodeScanningResult) -> None:
        """Scan for MRZ zones."""
        try:
            mrz_data = self.mrz_parser.parse_from_image(image)
            
            if mrz_data.success and mrz_data.format != MRZFormat.UNKNOWN:
                result.mrz_detected = True
                result.mrz_data = mrz_data
                result.mrz_format = mrz_data.format
                result.check_digits_valid = mrz_data.all_check_digits_valid
                
                logger.debug(f"MRZ parsed successfully (format: {mrz_data.format.value})")
                
        except Exception as e:
            logger.debug(f"MRZ scanning error: {type(e).__name__}")
    
    def _infer_document_category(self, result: BarcodeScanningResult) -> DocumentCategory:
        """Infer document category from scan results."""
        if result.barcode_detected and result.barcode_type == BarcodeType.PDF417:
            return DocumentCategory.DRIVERS_LICENSE
        
        if result.mrz_detected:
            if result.mrz_format == MRZFormat.TD3:
                return DocumentCategory.PASSPORT
            elif result.mrz_format in (MRZFormat.TD1, MRZFormat.TD2):
                return DocumentCategory.ID_CARD
        
        return DocumentCategory.UNKNOWN
    
    def _extract_identity_fields(self, result: BarcodeScanningResult) -> None:
        """Extract normalized identity fields from parsed data."""
        fields = {}
        
        # Prefer barcode data if available
        if result.pdf417_data and result.pdf417_data.success:
            fields.update(result.pdf417_data.get_identity_fields())
        
        # Add/override with MRZ data
        if result.mrz_data and result.mrz_data.success:
            mrz_fields = result.mrz_data.get_identity_fields()
            for key, value in mrz_fields.items():
                if value and (key not in fields or not fields[key]):
                    fields[key] = value
        
        result.identity_fields = fields
    
    def _generate_hashes(self, result: BarcodeScanningResult, salt: str) -> None:
        """Generate hashed versions of identity fields."""
        hashes = {}
        
        # Hash each identity field
        for field_name, value in result.identity_fields.items():
            if value:
                salted = f"{salt}:{field_name}:{value}".encode('utf-8')
                hash_value = hashlib.sha256(salted).hexdigest()
                hashes[f"barcode_{field_name}"] = hash_value
        
        result.field_hashes = hashes
        
        # Generate combined identity hash
        if result.identity_fields:
            sorted_items = sorted(result.identity_fields.items())
            combined = '|'.join(f"{k}:{v}" for k, v in sorted_items if v)
            salted = f"{salt}:barcode_identity:{combined}".encode('utf-8')
            result.identity_hash = hashlib.sha256(salted).hexdigest()
    
    def _cross_validate(
        self,
        result: BarcodeScanningResult,
        ocr_data: Dict[str, Any]
    ) -> None:
        """Cross-validate barcode data against OCR data."""
        if not result.identity_fields:
            return
        
        cross_result = self.cross_validator.validate(
            result.identity_fields,
            ocr_data
        )
        
        result.cross_validation = cross_result
    
    def _calculate_scores(self, result: BarcodeScanningResult) -> None:
        """Calculate all scoring metrics."""
        scoring = self.config.scoring
        
        # Barcode score
        if result.pdf417_data and result.pdf417_data.success:
            result.barcode_score = result.pdf417_data.confidence
        
        # MRZ score
        if result.mrz_data and result.mrz_data.success:
            result.mrz_score = result.mrz_data.confidence
            
            # Bonus for valid check digits
            if result.check_digits_valid:
                result.mrz_score = min(1.0, result.mrz_score + scoring.check_digit_bonus)
        
        # Combined validation score
        scores = []
        if result.barcode_score > 0:
            scores.append(result.barcode_score)
        if result.mrz_score > 0:
            scores.append(result.mrz_score)
        
        if scores:
            result.validation_score = sum(scores) / len(scores)
        
        # VEID contribution
        veid_contribution = 0.0
        
        # Base score from successful read
        if result.barcode_detected or result.mrz_detected:
            veid_contribution += scoring.barcode_read_score
        
        # Check digit bonus
        if result.check_digits_valid:
            veid_contribution += scoring.check_digit_bonus
        
        # Cross-validation contribution
        if result.cross_validation:
            veid_contribution += result.cross_validation.veid_score_contribution
        
        # Scale by validation score
        veid_contribution *= result.validation_score
        
        # Clamp to max
        result.veid_contribution = min(
            veid_contribution,
            scoring.max_score_contribution
        )


def create_pipeline(
    enable_pdf417: bool = True,
    enable_mrz: bool = True,
    strict_validation: bool = False,
) -> BarcodeScanningPipeline:
    """
    Factory function to create a configured pipeline.
    
    Args:
        enable_pdf417: Enable PDF417 barcode scanning.
        enable_mrz: Enable MRZ scanning.
        strict_validation: Use strict MRZ check digit validation.
        
    Returns:
        Configured BarcodeScanningPipeline.
    """
    config = BarcodeScanningConfig()
    
    if not enable_pdf417:
        config.pdf417.enable_aamva = False
    
    if not enable_mrz:
        config.mrz.supported_formats = set()
    
    config.mrz.strict_validation = strict_validation
    
    return BarcodeScanningPipeline(config)
