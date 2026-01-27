"""
Configuration classes for the OCR extraction pipeline.

This module defines all configuration parameters for:
- Tesseract OCR settings (language, PSM, character whitelists)
- ROI cropping settings (margins, preprocessing)
- Post-processing settings (error correction)
- Field parsing settings (patterns, validation)
- Hashing settings (algorithm, salting)
"""

from dataclasses import dataclass, field
from typing import Dict, List, Optional, Set
from enum import Enum


class DocumentType(str, Enum):
    """Supported document types for field parsing."""
    ID_CARD = "id_card"
    PASSPORT = "passport"
    DRIVERS_LICENSE = "drivers_license"
    UNKNOWN = "unknown"


class PageSegmentationMode(int, Enum):
    """Tesseract Page Segmentation Modes (PSM)."""
    OSD_ONLY = 0  # Orientation and script detection only
    AUTO_OSD = 1  # Automatic page segmentation with OSD
    AUTO_ONLY = 2  # Automatic page segmentation, no OSD
    FULLY_AUTO = 3  # Fully automatic page segmentation (default)
    SINGLE_COLUMN = 4  # Assume a single column of text
    SINGLE_BLOCK_VERT = 5  # Assume a single uniform block of vertically aligned text
    SINGLE_BLOCK = 6  # Assume a single uniform block of text
    SINGLE_LINE = 7  # Treat the image as a single text line
    SINGLE_WORD = 8  # Treat the image as a single word
    CIRCLE_WORD = 9  # Treat the image as a single word in a circle
    SINGLE_CHAR = 10  # Treat the image as a single character
    SPARSE_TEXT = 11  # Find as much text as possible in no particular order
    SPARSE_TEXT_OSD = 12  # Sparse text with OSD
    RAW_LINE = 13  # Treat the image as a single text line (no hacks)


class ThresholdingMethod(str, Enum):
    """Thresholding methods for OCR preprocessing."""
    NONE = "none"
    BINARY = "binary"
    BINARY_INV = "binary_inv"
    OTSU = "otsu"
    ADAPTIVE_MEAN = "adaptive_mean"
    ADAPTIVE_GAUSSIAN = "adaptive_gaussian"


@dataclass
class TesseractConfig:
    """Configuration for Tesseract OCR engine."""
    
    # Language settings
    lang: str = "eng"  # Tesseract language code
    
    # Page segmentation mode
    psm: PageSegmentationMode = PageSegmentationMode.SINGLE_BLOCK
    
    # OEM (OCR Engine Mode)
    # 0 = Legacy only, 1 = Neural nets LSTM only, 2 = Legacy + LSTM, 3 = Default
    oem: int = 3
    
    # Character whitelist (empty = all characters)
    whitelist: str = ""
    
    # Character blacklist
    blacklist: str = ""
    
    # Custom Tesseract config options
    custom_config: str = ""
    
    # Tesseract data path (None = use default)
    tessdata_path: Optional[str] = None
    
    # Minimum confidence threshold (0-100)
    min_confidence: float = 30.0
    
    def build_config_string(self) -> str:
        """Build Tesseract configuration string."""
        config_parts = [
            f"--psm {self.psm.value}",
            f"--oem {self.oem}",
        ]
        
        if self.tessdata_path:
            config_parts.append(f"--tessdata-dir {self.tessdata_path}")
        
        tesseract_config = []
        if self.whitelist:
            tesseract_config.append(f"-c tessedit_char_whitelist={self.whitelist}")
        if self.blacklist:
            tesseract_config.append(f"-c tessedit_char_blacklist={self.blacklist}")
        
        if self.custom_config:
            tesseract_config.append(self.custom_config)
        
        config_parts.extend(tesseract_config)
        return " ".join(config_parts)


@dataclass
class CropperConfig:
    """Configuration for ROI cropping."""
    
    # Margin to add around ROI (pixels)
    margin_pixels: int = 5
    
    # Margin as fraction of ROI size
    margin_fraction: float = 0.05
    
    # Use sub-pixel interpolation for rotated ROIs
    use_subpixel_crop: bool = True
    
    # Minimum crop size (pixels)
    min_crop_width: int = 20
    min_crop_height: int = 10
    
    # Maximum crop size (pixels) - will resize if larger
    max_crop_width: int = 2000
    max_crop_height: int = 500
    
    # Preprocessing for OCR
    convert_grayscale: bool = True
    thresholding: ThresholdingMethod = ThresholdingMethod.OTSU
    threshold_value: int = 127  # For binary thresholding
    adaptive_block_size: int = 11  # For adaptive thresholding
    adaptive_c: int = 2  # Constant subtracted for adaptive
    
    # Deskew small rotations
    deskew_enabled: bool = True
    deskew_max_angle: float = 5.0  # Maximum angle to correct (degrees)
    
    # Scale up small text
    scale_to_height: Optional[int] = 32  # Target height for scaling (None = no scaling)


# Character whitelists for common field types
FIELD_WHITELISTS: Dict[str, str] = {
    "alpha_only": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz",
    "alpha_space": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz ",
    "alpha_hyphen": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-' ",
    "numeric_only": "0123456789",
    "alphanumeric": "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789",
    "date_slash": "0123456789/",
    "date_dash": "0123456789-",
    "date_dot": "0123456789.",
    "mrz": "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<",
    "id_number": "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-",
}


@dataclass
class PostProcessingConfig:
    """Configuration for OCR post-processing."""
    
    # Common character confusion corrections
    enable_confusion_correction: bool = True
    
    # Whitespace normalization
    normalize_whitespace: bool = True
    strip_leading_trailing: bool = True
    
    # Case normalization
    convert_to_uppercase: bool = False
    convert_to_lowercase: bool = False
    
    # Remove specific characters
    remove_newlines: bool = True
    remove_special_chars: bool = False
    allowed_special_chars: str = "-/. '"
    
    # Custom replacements (applied after other corrections)
    custom_replacements: Dict[str, str] = field(default_factory=dict)


@dataclass
class FieldParserConfig:
    """Configuration for document field parsing."""
    
    # Field detection settings
    min_field_confidence: float = 0.5
    
    # Date format patterns (for parsing)
    date_formats: List[str] = field(default_factory=lambda: [
        "%d/%m/%Y",
        "%m/%d/%Y",
        "%Y-%m-%d",
        "%d-%m-%Y",
        "%d.%m.%Y",
        "%d %b %Y",
        "%d %B %Y",
    ])
    
    # ID number patterns by country
    id_patterns: Dict[str, str] = field(default_factory=lambda: {
        "default": r"[A-Z0-9]{5,20}",
        "us_ssn": r"\d{3}-\d{2}-\d{4}",
        "uk_nino": r"[A-Z]{2}\d{6}[A-Z]",
        "passport_mrz_line1": r"P<[A-Z]{3}[A-Z<]+",
        "passport_mrz_line2": r"[A-Z0-9<]{44}",
    })
    
    # Name field settings
    max_name_length: int = 100
    min_name_length: int = 2


@dataclass
class HashingConfig:
    """Configuration for field hashing."""
    
    # Hashing algorithm
    algorithm: str = "sha256"
    
    # Salt for hashing (should be set from environment in production)
    salt: str = ""
    
    # Normalize values before hashing
    normalize_before_hash: bool = True
    uppercase_before_hash: bool = True
    strip_whitespace: bool = True
    
    # Include field name in hash input
    include_field_name: bool = True
    
    # Hash output encoding
    output_encoding: str = "hex"  # "hex" or "base64"


@dataclass
class OCRExtractionConfig:
    """Master configuration for OCR extraction pipeline."""
    
    tesseract: TesseractConfig = field(default_factory=TesseractConfig)
    cropper: CropperConfig = field(default_factory=CropperConfig)
    postprocessing: PostProcessingConfig = field(default_factory=PostProcessingConfig)
    field_parser: FieldParserConfig = field(default_factory=FieldParserConfig)
    hashing: HashingConfig = field(default_factory=HashingConfig)
    
    # Pipeline settings
    max_rois_per_image: int = 100
    parallel_processing: bool = False  # Process ROIs in parallel
    
    # Debug settings
    save_intermediate_crops: bool = False
    intermediate_crop_path: str = "/tmp/ocr_crops"
    
    # Document type specific Tesseract configs
    document_configs: Dict[DocumentType, TesseractConfig] = field(
        default_factory=lambda: {
            DocumentType.ID_CARD: TesseractConfig(
                psm=PageSegmentationMode.SINGLE_BLOCK,
                whitelist=FIELD_WHITELISTS["alphanumeric"] + "-/. ",
            ),
            DocumentType.PASSPORT: TesseractConfig(
                psm=PageSegmentationMode.SINGLE_LINE,
                whitelist=FIELD_WHITELISTS["mrz"],
            ),
            DocumentType.DRIVERS_LICENSE: TesseractConfig(
                psm=PageSegmentationMode.SINGLE_BLOCK,
                whitelist=FIELD_WHITELISTS["alphanumeric"] + "-/. ",
            ),
        }
    )
    
    def get_tesseract_config(self, doc_type: DocumentType) -> TesseractConfig:
        """Get Tesseract config for document type."""
        return self.document_configs.get(doc_type, self.tesseract)
