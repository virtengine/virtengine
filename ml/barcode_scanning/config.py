"""
Configuration classes for the barcode scanning module.

This module defines configuration parameters for:
- PDF417 barcode scanning (driver licenses)
- MRZ parsing (passports, ID cards)
- Cross-validation settings
- Scoring thresholds
"""

from dataclasses import dataclass, field
from typing import Dict, List, Optional, Set
from enum import Enum


class BarcodeType(str, Enum):
    """Supported barcode types."""
    PDF417 = "pdf417"
    QR_CODE = "qr_code"
    DATA_MATRIX = "data_matrix"
    AZTEC = "aztec"
    UNKNOWN = "unknown"


class MRZFormat(str, Enum):
    """ICAO Machine Readable Zone formats."""
    TD1 = "td1"  # 3 lines x 30 chars (ID cards)
    TD2 = "td2"  # 2 lines x 36 chars (older ID cards, some visas)
    TD3 = "td3"  # 2 lines x 44 chars (passports)
    UNKNOWN = "unknown"


class DocumentCategory(str, Enum):
    """Document categories for barcode scanning."""
    DRIVERS_LICENSE = "drivers_license"
    PASSPORT = "passport"
    ID_CARD = "id_card"
    VISA = "visa"
    UNKNOWN = "unknown"


@dataclass
class PDF417Config:
    """Configuration for PDF417 barcode parsing."""
    
    # Enable AAMVA (American Association of Motor Vehicle Administrators) parsing
    enable_aamva: bool = True
    
    # Minimum version to support (older versions have fewer fields)
    min_aamva_version: int = 1
    
    # Maximum version (for forward compatibility)
    max_aamva_version: int = 10
    
    # Fields considered critical for identity verification
    critical_fields: Set[str] = field(default_factory=lambda: {
        "DAC",  # First name
        "DCS",  # Last name
        "DAG",  # Street address
        "DAI",  # City
        "DAJ",  # State
        "DAK",  # Postal code
        "DBB",  # Date of birth
        "DBC",  # Sex
        "DAQ",  # Document number
        "DBA",  # Expiry date
    })
    
    # Fields that are optional but useful
    optional_fields: Set[str] = field(default_factory=lambda: {
        "DAD",  # Middle name
        "DDB",  # Issue date
        "DDD",  # Document discriminator
        "DCG",  # Country
        "DAY",  # Suffix
        "DAU",  # Height
        "DAW",  # Weight
        "DAZ",  # Hair color
        "DAY",  # Eye color
    })
    
    # Confidence threshold for barcode detection (0.0 - 1.0)
    detection_confidence_threshold: float = 0.5
    
    # Enable automatic orientation detection
    auto_rotate: bool = True
    
    # Maximum image dimension for processing (larger images are resized)
    max_image_dimension: int = 4096


@dataclass
class MRZConfig:
    """Configuration for MRZ (Machine Readable Zone) parsing."""
    
    # Supported MRZ formats
    supported_formats: Set[MRZFormat] = field(default_factory=lambda: {
        MRZFormat.TD1,
        MRZFormat.TD2,
        MRZFormat.TD3,
    })
    
    # Enable check digit validation
    validate_check_digits: bool = True
    
    # Strict mode: reject if any check digit fails
    strict_validation: bool = False
    
    # MRZ character set (ICAO Doc 9303 standard)
    valid_characters: str = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<"
    
    # Filler character
    filler: str = "<"
    
    # Minimum line length to consider as MRZ
    min_line_length: int = 28
    
    # OCR preprocessing for MRZ region
    preprocess_mrz_region: bool = True
    
    # Confidence threshold for character recognition
    char_confidence_threshold: float = 0.7
    
    # Enable O/0 and I/1 correction
    enable_ocr_correction: bool = True


@dataclass 
class CrossValidationConfig:
    """Configuration for cross-validation between barcode and OCR data."""
    
    # Minimum similarity score for string matching (0.0 - 1.0)
    string_similarity_threshold: float = 0.8
    
    # Exact match bonus (added when strings match exactly)
    exact_match_bonus: float = 0.1
    
    # Field weights for scoring (higher = more important)
    field_weights: Dict[str, float] = field(default_factory=lambda: {
        "full_name": 1.0,
        "surname": 0.9,
        "given_names": 0.9,
        "date_of_birth": 1.0,
        "document_number": 1.0,
        "expiry_date": 0.8,
        "sex": 0.5,
        "nationality": 0.6,
        "address": 0.7,
    })
    
    # Fields that must match for validation to pass
    required_fields: Set[str] = field(default_factory=lambda: {
        "date_of_birth",
    })
    
    # Minimum number of matching fields
    min_matching_fields: int = 2
    
    # Enable fuzzy matching for names
    enable_fuzzy_matching: bool = True
    
    # Levenshtein distance threshold for fuzzy matching
    max_edit_distance: int = 2
    
    # Case sensitivity for string comparison
    case_sensitive: bool = False
    
    # Date format tolerance (allow different formats)
    flexible_date_matching: bool = True


@dataclass
class ScoringConfig:
    """Configuration for VEID scoring integration."""
    
    # Base score for successful barcode read
    barcode_read_score: float = 0.15
    
    # Bonus for valid check digits
    check_digit_bonus: float = 0.05
    
    # Cross-validation score weight
    cross_validation_weight: float = 0.20
    
    # Penalty for failed cross-validation
    cross_validation_penalty: float = -0.10
    
    # Maximum score contribution from barcode scanning
    max_score_contribution: float = 0.40
    
    # Minimum score to consider barcode valid
    min_valid_score: float = 0.10


@dataclass
class BarcodeScanningConfig:
    """Main configuration for barcode scanning pipeline."""
    
    # Component configurations
    pdf417: PDF417Config = field(default_factory=PDF417Config)
    mrz: MRZConfig = field(default_factory=MRZConfig)
    cross_validation: CrossValidationConfig = field(default_factory=CrossValidationConfig)
    scoring: ScoringConfig = field(default_factory=ScoringConfig)
    
    # General settings
    model_version: str = "1.0.0"
    
    # Enable debug mode (logs processing steps, NOT PII)
    debug_mode: bool = False
    
    # Timeout for scanning operations (seconds)
    timeout_seconds: float = 30.0
    
    # Enable parallel processing for multiple barcodes
    parallel_processing: bool = True
    
    # Maximum barcodes to process per image
    max_barcodes: int = 5
    
    # Image preprocessing
    enable_preprocessing: bool = True
    grayscale_conversion: bool = True
    contrast_enhancement: bool = True
    
    def validate(self) -> List[str]:
        """
        Validate configuration values.
        
        Returns:
            List of validation error messages (empty if valid).
        """
        errors = []
        
        if self.scoring.max_score_contribution > 1.0:
            errors.append("max_score_contribution cannot exceed 1.0")
        
        if self.cross_validation.string_similarity_threshold < 0.5:
            errors.append("string_similarity_threshold should be at least 0.5")
        
        if self.timeout_seconds <= 0:
            errors.append("timeout_seconds must be positive")
        
        return errors
