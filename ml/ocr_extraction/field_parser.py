"""
Document field parser for structured extraction.

This module parses OCR results into structured fields
for various document types (ID cards, passports, driver's licenses).

Supports three parsing modes:
1. Label-based: Extracts fields using regex patterns on OCR text
2. Center-matching: Matches U-Net mask regions to OCR text boxes by center position
   (ported from RMIT OCR_Document_Scan intern project, VE-3041)
3. Config-based: Uses configurable document type definitions from YAML
   (VE-3047 - Generalize Document Type Configuration)

The config-based mode allows adding new document types without code changes
by defining field patterns, label text, and validation rules in document_types.yaml.
"""

import re
import logging
from typing import Dict, List, Optional, Any, Tuple
from dataclasses import dataclass, field
from enum import Enum
from datetime import datetime, date

import numpy as np

from ml.ocr_extraction.config import FieldParserConfig, DocumentType, CenterMatchingConfig
from ml.ocr_extraction.tesseract_wrapper import OCRResult
from ml.ocr_extraction.postprocessing import OCRPostProcessor
from ml.ocr_extraction.center_matching import (
    BoundingBox,
    MaskRegion,
    CenterMatcher,
    CenterMatchingConfig as CenterMatcherConfig,
    MatchResult,
    extract_mask_regions_from_unet,
    convert_ocr_boxes,
)
from ml.ocr_extraction.document_config import (
    DocumentConfigLoader,
    DocumentTypeDefinition,
    FieldDefinition,
)

logger = logging.getLogger(__name__)


class ValidationStatus(str, Enum):
    """Validation status for parsed fields."""
    VALID = "valid"
    UNCERTAIN = "uncertain"
    INVALID = "invalid"
    NOT_FOUND = "not_found"


@dataclass
class ParsedField:
    """
    A parsed document field with value and metadata.
    
    Contains the extracted value, confidence scores,
    and validation information.
    """
    field_name: str
    value: str
    confidence: float  # 0.0 - 1.0
    source_roi_ids: List[str]
    validation_status: ValidationStatus
    raw_value: str = ""  # Original OCR text before processing
    validation_message: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "field_name": self.field_name,
            "value": self.value,
            "confidence": self.confidence,
            "source_roi_ids": self.source_roi_ids,
            "validation_status": self.validation_status.value,
            "raw_value": self.raw_value,
            "validation_message": self.validation_message,
        }
    
    @property
    def is_valid(self) -> bool:
        """Check if field is valid."""
        return self.validation_status == ValidationStatus.VALID
    
    @property
    def is_usable(self) -> bool:
        """Check if field is usable (valid or uncertain)."""
        return self.validation_status in (
            ValidationStatus.VALID,
            ValidationStatus.UNCERTAIN
        )


@dataclass
class MRZData:
    """Parsed MRZ (Machine Readable Zone) data."""
    document_type: str  # P for passport
    country_code: str
    surname: str
    given_names: str
    document_number: str
    nationality: str
    date_of_birth: str  # YYMMDD
    sex: str
    expiry_date: str  # YYMMDD
    personal_number: str
    check_digits_valid: bool
    raw_lines: List[str]
    
    def to_fields(self) -> Dict[str, ParsedField]:
        """Convert to ParsedField dictionary."""
        confidence = 0.9 if self.check_digits_valid else 0.5
        status = ValidationStatus.VALID if self.check_digits_valid else ValidationStatus.UNCERTAIN
        
        return {
            "document_type": ParsedField(
                field_name="document_type",
                value=self.document_type,
                confidence=confidence,
                source_roi_ids=[],
                validation_status=status,
            ),
            "country_code": ParsedField(
                field_name="country_code",
                value=self.country_code,
                confidence=confidence,
                source_roi_ids=[],
                validation_status=status,
            ),
            "surname": ParsedField(
                field_name="surname",
                value=self.surname,
                confidence=confidence,
                source_roi_ids=[],
                validation_status=status,
            ),
            "given_names": ParsedField(
                field_name="given_names",
                value=self.given_names,
                confidence=confidence,
                source_roi_ids=[],
                validation_status=status,
            ),
            "document_number": ParsedField(
                field_name="document_number",
                value=self.document_number,
                confidence=confidence,
                source_roi_ids=[],
                validation_status=status,
            ),
            "nationality": ParsedField(
                field_name="nationality",
                value=self.nationality,
                confidence=confidence,
                source_roi_ids=[],
                validation_status=status,
            ),
            "date_of_birth": ParsedField(
                field_name="date_of_birth",
                value=self._format_date(self.date_of_birth),
                confidence=confidence,
                source_roi_ids=[],
                validation_status=status,
            ),
            "sex": ParsedField(
                field_name="sex",
                value=self.sex,
                confidence=confidence,
                source_roi_ids=[],
                validation_status=status,
            ),
            "expiry_date": ParsedField(
                field_name="expiry_date",
                value=self._format_date(self.expiry_date),
                confidence=confidence,
                source_roi_ids=[],
                validation_status=status,
            ),
        }
    
    def _format_date(self, yymmdd: str) -> str:
        """Convert YYMMDD to YYYY-MM-DD."""
        if len(yymmdd) != 6:
            return yymmdd
        
        yy = int(yymmdd[:2])
        mm = yymmdd[2:4]
        dd = yymmdd[4:6]
        
        # Assume 2000s for years < 50, 1900s otherwise
        year = 2000 + yy if yy < 50 else 1900 + yy
        
        return f"{year}-{mm}-{dd}"


class DocumentFieldParser:
    """
    Parser for extracting structured fields from OCR results.
    
    Supports multiple document types with specialized
    parsing logic for each.
    """
    
    # Field label patterns for automatic detection
    FIELD_PATTERNS = {
        "surname": [
            r"surname[:\s]*(.*)",
            r"last\s*name[:\s]*(.*)",
            r"family\s*name[:\s]*(.*)",
        ],
        "given_names": [
            r"given\s*names?[:\s]*(.*)",
            r"first\s*name[:\s]*(.*)",
            r"forename[:\s]*(.*)",
        ],
        "full_name": [
            r"name[:\s]*(.*)",
            r"full\s*name[:\s]*(.*)",
        ],
        "date_of_birth": [
            r"date\s*of\s*birth[:\s]*(.*)",
            r"d\.?o\.?b\.?[:\s]*(.*)",
            r"birth\s*date[:\s]*(.*)",
            r"born[:\s]*(.*)",
        ],
        "expiry_date": [
            r"expiry[:\s]*(.*)",
            r"expires?[:\s]*(.*)",
            r"expiration[:\s]*(.*)",
            r"valid\s*until[:\s]*(.*)",
        ],
        "issue_date": [
            r"issue\s*date[:\s]*(.*)",
            r"issued[:\s]*(.*)",
            r"date\s*of\s*issue[:\s]*(.*)",
        ],
        "document_number": [
            r"document\s*no\.?[:\s]*(.*)",
            r"id\s*no\.?[:\s]*(.*)",
            r"number[:\s]*(.*)",
            r"license\s*no\.?[:\s]*(.*)",
            r"passport\s*no\.?[:\s]*(.*)",
        ],
        "nationality": [
            r"nationality[:\s]*(.*)",
            r"citizen(ship)?[:\s]*(.*)",
        ],
        "sex": [
            r"sex[:\s]*(.*)",
            r"gender[:\s]*(.*)",
        ],
        "address": [
            r"address[:\s]*(.*)",
            r"residence[:\s]*(.*)",
        ],
    }
    
    def __init__(
        self,
        config: Optional[FieldParserConfig] = None,
        center_matching_config: Optional[CenterMatchingConfig] = None,
    ):
        """
        Initialize field parser.
        
        Args:
            config: Parser configuration. Uses defaults if None.
            center_matching_config: Center matching configuration for U-Net mode.
                                   If provided, enables center matching mode.
        """
        self.config = config or FieldParserConfig()
        self.postprocessor = OCRPostProcessor()
        self.center_matching_config = center_matching_config
        
        # Initialize document config loader for config-based parsing (VE-3047)
        try:
            self._doc_config = DocumentConfigLoader.get_instance()
        except FileNotFoundError:
            logger.warning(
                "document_types.yaml not found, config-based parsing disabled"
            )
            self._doc_config = None
        
        # Initialize center matcher if config provided
        if center_matching_config and center_matching_config.enabled:
            self._center_matcher = CenterMatcher(CenterMatcherConfig(
                distance_threshold=center_matching_config.distance_threshold,
                merge_threshold=center_matching_config.merge_threshold,
                min_confidence=center_matching_config.min_confidence,
                use_weighted_center=center_matching_config.use_weighted_center,
                neighbor_distance_pixels=center_matching_config.neighbor_distance_pixels,
            ))
        else:
            self._center_matcher = None
    
    def parse(
        self,
        ocr_results: List[OCRResult],
        document_type: DocumentType
    ) -> Dict[str, ParsedField]:
        """
        Parse OCR results for the specified document type.
        
        Args:
            ocr_results: List of OCR results from ROIs
            document_type: Type of document being parsed
            
        Returns:
            Dictionary of parsed fields
        """
        parsers = {
            DocumentType.ID_CARD: self.parse_id_card,
            DocumentType.PASSPORT: self.parse_passport,
            DocumentType.DRIVERS_LICENSE: self.parse_drivers_license,
            DocumentType.UNKNOWN: self.parse_generic,
        }
        
        parser = parsers.get(document_type, self.parse_generic)
        return parser(ocr_results)
    
    def parse_with_config(
        self,
        ocr_results: List[OCRResult],
        document_type_id: str,
    ) -> Dict[str, ParsedField]:
        """
        Parse OCR results using configurable document type definitions.
        
        This method uses document_types.yaml for field definitions,
        allowing new document types to be added without code changes.
        
        Task Reference: VE-3047 - Generalize Document Type Configuration
        
        Args:
            ocr_results: List of OCR results from ROIs
            document_type_id: Document type identifier from config
                             (e.g., "passport", "turkish_id", "us_drivers_license")
        
        Returns:
            Dictionary of parsed fields
        
        Raises:
            ValueError: If document type is not found in configuration
        
        Example:
            parser = DocumentFieldParser()
            fields = parser.parse_with_config(ocr_results, "turkish_id")
        """
        if self._doc_config is None:
            raise ValueError(
                "Document configuration not loaded. "
                "Ensure document_types.yaml exists."
            )
        
        doc_type_def = self._doc_config.get_document_type(document_type_id)
        if doc_type_def is None:
            available = self._doc_config.list_document_types()
            raise ValueError(
                f"Unknown document type: '{document_type_id}'. "
                f"Available types: {available}"
            )
        
        fields = {}
        all_text = self._combine_text(ocr_results)
        
        # Extract fields using label-based matching from config
        for field_name, field_def in doc_type_def.fields.items():
            result = self._extract_field_with_config(
                field_def, all_text, ocr_results
            )
            if result:
                fields[field_name] = result
        
        # Apply heuristic fallbacks for common fields if not found
        fields = self._apply_heuristic_fallbacks(fields, ocr_results, doc_type_def)
        
        return fields
    
    def _extract_field_with_config(
        self,
        field_def: FieldDefinition,
        all_text: str,
        ocr_results: List[OCRResult],
    ) -> Optional[ParsedField]:
        """
        Extract a field using its configuration definition.
        
        Uses label text from config to find field values in OCR text.
        
        Args:
            field_def: Field definition from config
            all_text: Combined OCR text
            ocr_results: Original OCR results
        
        Returns:
            ParsedField if found, None otherwise
        """
        # Try label-based extraction if labels are defined
        if field_def.label_text:
            for label in field_def.label_text:
                # Build pattern from label
                escaped_label = re.escape(label)
                pattern = rf"{escaped_label}[:\s]*(.+?)(?:\n|$)"
                
                match = re.search(pattern, all_text, re.IGNORECASE)
                if match:
                    value = match.group(1).strip()
                    if value:
                        return self._create_parsed_field(
                            field_def, value, ocr_results
                        )
        
        # Try pattern matching if pattern is defined
        if field_def.pattern:
            # Search for pattern in text
            matches = re.findall(field_def.pattern, all_text)
            if matches:
                value = matches[0] if isinstance(matches[0], str) else matches[0][0]
                return self._create_parsed_field(
                    field_def, value, ocr_results
                )
        
        return None
    
    def _create_parsed_field(
        self,
        field_def: FieldDefinition,
        raw_value: str,
        ocr_results: List[OCRResult],
    ) -> ParsedField:
        """
        Create a ParsedField from extracted value and field definition.
        
        Args:
            field_def: Field definition from config
            raw_value: Extracted raw value
            ocr_results: Original OCR results for confidence calculation
        
        Returns:
            ParsedField with validation applied
        """
        # Post-process value
        processed_value = self.postprocessor.process(raw_value, field_def.name)
        
        # Validate against pattern
        if field_def.pattern and field_def.matches_pattern(processed_value):
            validation_status = ValidationStatus.VALID
        elif field_def.pattern:
            validation_status = ValidationStatus.UNCERTAIN
        else:
            validation_status = ValidationStatus.VALID
        
        # Find source ROIs
        source_rois = self._find_source_rois(raw_value, ocr_results)
        confidence = self._calculate_confidence(raw_value, ocr_results)
        
        return ParsedField(
            field_name=field_def.name,
            value=processed_value,
            confidence=confidence,
            source_roi_ids=source_rois,
            validation_status=validation_status,
            raw_value=raw_value,
        )
    
    def _apply_heuristic_fallbacks(
        self,
        fields: Dict[str, ParsedField],
        ocr_results: List[OCRResult],
        doc_type_def: DocumentTypeDefinition,
    ) -> Dict[str, ParsedField]:
        """
        Apply heuristic extraction for missing fields.
        
        Args:
            fields: Already extracted fields
            ocr_results: OCR results
            doc_type_def: Document type definition
        
        Returns:
            Updated fields dictionary
        """
        # Check for name fields
        name_fields = {"full_name", "name", "surname", "given_names", "adi", "soyadi"}
        has_name = any(f in fields for f in name_fields)
        
        if not has_name:
            name_field = self._extract_name_heuristic(ocr_results)
            if name_field:
                fields["full_name"] = name_field
        
        # Check for date of birth
        dob_fields = {"date_of_birth", "dogum_tarihi", "dob", "dbb"}
        has_dob = any(f in fields for f in dob_fields)
        
        if not has_dob:
            dob_field = self._extract_date_heuristic(ocr_results, "date_of_birth")
            if dob_field:
                fields["date_of_birth"] = dob_field
        
        # Check for document number
        id_fields = {"document_number", "tc_kimlik_no", "licence_number", "daq"}
        has_id = any(f in fields for f in id_fields)
        
        if not has_id:
            id_field = self._extract_id_number_heuristic(ocr_results)
            if id_field:
                fields["document_number"] = id_field
        
        return fields
    
    def get_available_document_types(self) -> List[str]:
        """
        Get list of available document types from configuration.
        
        Returns:
            List of document type identifiers
        """
        if self._doc_config is None:
            return []
        return self._doc_config.list_document_types()
    
    def validate_extracted_fields(
        self,
        document_type_id: str,
        fields: Dict[str, ParsedField],
    ) -> Tuple[bool, List[str]]:
        """
        Validate extracted fields against document type configuration.
        
        Args:
            document_type_id: Document type identifier
            fields: Dictionary of extracted fields
        
        Returns:
            Tuple of (all_valid, list_of_errors)
        """
        if self._doc_config is None:
            return True, []
        
        field_values = {name: f.value for name, f in fields.items()}
        return self._doc_config.validate_document(document_type_id, field_values)
    
    def parse_id_card(
        self,
        ocr_results: List[OCRResult]
    ) -> Dict[str, ParsedField]:
        """
        Extract fields from an ID card.
        
        Common fields: surname, given_names, date_of_birth,
        document_number, nationality, expiry_date
        
        Args:
            ocr_results: OCR results from detected text regions
            
        Returns:
            Dictionary of parsed fields
        """
        fields = {}
        all_text = self._combine_text(ocr_results)
        
        # Try label-based extraction first
        for field_name, patterns in self.FIELD_PATTERNS.items():
            result = self._extract_by_label(all_text, patterns, ocr_results)
            if result:
                fields[field_name] = result
        
        # Fill in missing fields with heuristics
        if "full_name" not in fields and "surname" not in fields:
            name_field = self._extract_name_heuristic(ocr_results)
            if name_field:
                fields["full_name"] = name_field
        
        if "date_of_birth" not in fields:
            dob_field = self._extract_date_heuristic(ocr_results, "date_of_birth")
            if dob_field:
                fields["date_of_birth"] = dob_field
        
        if "document_number" not in fields:
            id_field = self._extract_id_number_heuristic(ocr_results)
            if id_field:
                fields["document_number"] = id_field
        
        return fields
    
    def parse_passport(
        self,
        ocr_results: List[OCRResult]
    ) -> Dict[str, ParsedField]:
        """
        Extract fields from a passport.
        
        Primarily uses MRZ (Machine Readable Zone) for extraction.
        
        Args:
            ocr_results: OCR results from detected text regions
            
        Returns:
            Dictionary of parsed fields
        """
        fields = {}
        
        # Look for MRZ lines
        mrz_lines = self._find_mrz_lines(ocr_results)
        
        if len(mrz_lines) >= 2:
            # Parse MRZ
            mrz_data = self._parse_mrz(mrz_lines)
            if mrz_data:
                fields = mrz_data.to_fields()
        
        # Also try visual zone extraction
        all_text = self._combine_text(ocr_results)
        
        for field_name, patterns in self.FIELD_PATTERNS.items():
            if field_name not in fields:
                result = self._extract_by_label(all_text, patterns, ocr_results)
                if result:
                    fields[field_name] = result
        
        return fields
    
    def parse_drivers_license(
        self,
        ocr_results: List[OCRResult]
    ) -> Dict[str, ParsedField]:
        """
        Extract fields from a driver's license.
        
        Common fields: full_name, date_of_birth, address,
        document_number, expiry_date, class/category
        
        Args:
            ocr_results: OCR results from detected text regions
            
        Returns:
            Dictionary of parsed fields
        """
        fields = {}
        all_text = self._combine_text(ocr_results)
        
        # Standard label-based extraction
        for field_name, patterns in self.FIELD_PATTERNS.items():
            result = self._extract_by_label(all_text, patterns, ocr_results)
            if result:
                fields[field_name] = result
        
        # Driver's license specific: class/category
        class_patterns = [
            r"class[:\s]*([A-Z0-9]+)",
            r"category[:\s]*([A-Z0-9]+)",
            r"cat\.?[:\s]*([A-Z0-9]+)",
        ]
        class_result = self._extract_by_label(all_text, class_patterns, ocr_results)
        if class_result:
            fields["license_class"] = class_result
        
        # Heuristic fallbacks
        if "full_name" not in fields:
            name_field = self._extract_name_heuristic(ocr_results)
            if name_field:
                fields["full_name"] = name_field
        
        if "document_number" not in fields:
            id_field = self._extract_id_number_heuristic(ocr_results)
            if id_field:
                fields["document_number"] = id_field
        
        return fields
    
    def parse_generic(
        self,
        ocr_results: List[OCRResult]
    ) -> Dict[str, ParsedField]:
        """
        Extract fields from an unknown document type.
        
        Uses heuristics and pattern matching without
        document-specific knowledge.
        
        Args:
            ocr_results: OCR results from detected text regions
            
        Returns:
            Dictionary of parsed fields
        """
        fields = {}
        all_text = self._combine_text(ocr_results)
        
        # Try all patterns
        for field_name, patterns in self.FIELD_PATTERNS.items():
            result = self._extract_by_label(all_text, patterns, ocr_results)
            if result:
                fields[field_name] = result
        
        # Extract all dates found
        dates = self._extract_all_dates(all_text)
        for i, date_str in enumerate(dates):
            if f"date_{i}" not in fields:
                fields[f"date_{i}"] = ParsedField(
                    field_name=f"date_{i}",
                    value=date_str,
                    confidence=0.6,
                    source_roi_ids=[],
                    validation_status=ValidationStatus.UNCERTAIN,
                )
        
        return fields
    
    def parse_with_center_matching(
        self,
        unet_mask: np.ndarray,
        ocr_boxes: List[BoundingBox],
        img_width: int,
        img_height: int,
        class_names: Optional[List[str]] = None,
    ) -> Dict[str, ParsedField]:
        """
        Parse fields using U-Net mask to OCR box center matching.
        
        This is an alternative to label-based parsing that uses spatial
        matching instead of text pattern matching. Ported from RMIT
        OCR_Document_Scan intern project (2023).
        
        The algorithm:
        1. Extract mask regions from U-Net output
        2. Convert mask centers to normalized ratios
        3. Convert OCR box centers to normalized ratios
        4. Match each mask region to nearest OCR box
        5. Return matched text as field values
        
        Args:
            unet_mask: Binary mask from U-Net segmentation model
            ocr_boxes: List of detected OCR text boxes with labels
            img_width: Image width in pixels
            img_height: Image height in pixels
            class_names: Optional list of field names for each mask class
                        (e.g., ["surname", "given_names", "date_of_birth"])
        
        Returns:
            Dictionary of parsed fields
            
        Raises:
            RuntimeError: If center matching is not enabled
        
        Task Reference: VE-3041 - Port Center-Matching Algorithm
        """
        if self._center_matcher is None:
            raise RuntimeError(
                "Center matching not enabled. Initialize parser with "
                "center_matching_config to use this method."
            )
        
        fields = {}
        
        # Extract mask regions from U-Net output
        max_regions = (
            self.center_matching_config.max_mask_regions
            if self.center_matching_config
            else 10
        )
        min_area = (
            self.center_matching_config.min_mask_area
            if self.center_matching_config
            else 100
        )
        
        mask_regions = extract_mask_regions_from_unet(
            unet_mask,
            class_names=class_names,
            min_area=min_area,
            max_regions=max_regions
        )
        
        if not mask_regions:
            return fields
        
        # Optionally merge adjacent OCR boxes
        merged_boxes = self._center_matcher.merge_adjacent_boxes(
            ocr_boxes, img_width, img_height
        )
        
        # Match mask regions to OCR boxes
        match_results = self._center_matcher.match_masks_to_boxes(
            mask_regions,
            merged_boxes,
            img_width,
            img_height
        )
        
        # Convert matches to ParsedField objects
        for result in match_results:
            if result.is_matched and result.matched_box:
                field_name = result.mask_region.class_name
                text_value = result.matched_box.label or ""
                
                # Post-process the extracted text
                processed_value = self.postprocessor.process(text_value, field_name)
                
                # Validate the field
                validation = self._validate_field(field_name, processed_value)
                
                fields[field_name] = ParsedField(
                    field_name=field_name,
                    value=processed_value,
                    confidence=result.confidence * result.matched_box.confidence,
                    source_roi_ids=[],
                    validation_status=validation,
                    raw_value=text_value,
                )
        
        return fields
    
    def parse_with_mask_regions(
        self,
        mask_regions: List[MaskRegion],
        ocr_boxes: List[BoundingBox],
        img_width: int,
        img_height: int,
    ) -> Dict[str, ParsedField]:
        """
        Parse fields using pre-extracted mask regions.
        
        Use this method when you already have MaskRegion objects
        (e.g., from a custom mask extraction pipeline).
        
        Args:
            mask_regions: Pre-extracted mask regions with class names
            ocr_boxes: List of detected OCR text boxes
            img_width: Image width in pixels
            img_height: Image height in pixels
            
        Returns:
            Dictionary of parsed fields
        """
        if self._center_matcher is None:
            raise RuntimeError(
                "Center matching not enabled. Initialize parser with "
                "center_matching_config to use this method."
            )
        
        fields = {}
        
        # Merge adjacent boxes
        merged_boxes = self._center_matcher.merge_adjacent_boxes(
            ocr_boxes, img_width, img_height
        )
        
        # Match and convert
        match_results = self._center_matcher.match_masks_to_boxes(
            mask_regions,
            merged_boxes,
            img_width,
            img_height
        )
        
        for result in match_results:
            if result.is_matched and result.matched_box:
                field_name = result.mask_region.class_name
                text_value = result.matched_box.label or ""
                processed_value = self.postprocessor.process(text_value, field_name)
                validation = self._validate_field(field_name, processed_value)
                
                fields[field_name] = ParsedField(
                    field_name=field_name,
                    value=processed_value,
                    confidence=result.confidence * result.matched_box.confidence,
                    source_roi_ids=[],
                    validation_status=validation,
                    raw_value=text_value,
                )
        
        return fields

    def _combine_text(self, ocr_results: List[OCRResult]) -> str:
        """Combine all OCR results into single text."""
        return "\n".join(r.text for r in ocr_results if r.text)
    
    def _extract_by_label(
        self,
        text: str,
        patterns: List[str],
        ocr_results: List[OCRResult]
    ) -> Optional[ParsedField]:
        """Extract field value using label patterns."""
        text_lower = text.lower()
        
        for pattern in patterns:
            match = re.search(pattern, text_lower, re.IGNORECASE)
            if match:
                value = match.group(1).strip() if match.lastindex else ""
                
                if not value:
                    continue
                
                # Find source ROI
                source_rois = self._find_source_rois(value, ocr_results)
                confidence = self._calculate_confidence(value, ocr_results)
                
                # Process value based on pattern type
                field_name = patterns[0].split("[")[0].replace("\\s*", "_").strip()
                processed_value = self.postprocessor.process(value, field_name)
                
                validation = self._validate_field(field_name, processed_value)
                
                return ParsedField(
                    field_name=field_name,
                    value=processed_value,
                    confidence=confidence,
                    source_roi_ids=source_rois,
                    validation_status=validation,
                    raw_value=value,
                )
        
        return None
    
    def _extract_name_heuristic(
        self,
        ocr_results: List[OCRResult]
    ) -> Optional[ParsedField]:
        """Extract name using heuristics."""
        # Look for text that looks like a name (Title Case, 2+ words)
        for result in ocr_results:
            text = result.text.strip()
            words = text.split()
            
            if len(words) >= 2 and len(words) <= 5:
                # Check if looks like a name (mostly letters)
                if all(word[0].isupper() for word in words if word):
                    alpha_ratio = sum(c.isalpha() for c in text) / len(text) if text else 0
                    
                    if alpha_ratio > 0.8:
                        processed = self.postprocessor.process(text, "name")
                        
                        return ParsedField(
                            field_name="full_name",
                            value=processed,
                            confidence=result.normalized_confidence * 0.7,
                            source_roi_ids=[result.roi_id] if result.roi_id else [],
                            validation_status=ValidationStatus.UNCERTAIN,
                            raw_value=text,
                        )
        
        return None
    
    def _extract_date_heuristic(
        self,
        ocr_results: List[OCRResult],
        field_name: str = "date"
    ) -> Optional[ParsedField]:
        """Extract date using heuristics."""
        date_pattern = r"\d{1,2}[/\-\.]\d{1,2}[/\-\.]\d{2,4}"
        
        for result in ocr_results:
            match = re.search(date_pattern, result.text)
            if match:
                date_str = match.group(0)
                processed = self.postprocessor.process(date_str, "date")
                
                validation = self._validate_date(processed)
                
                return ParsedField(
                    field_name=field_name,
                    value=processed,
                    confidence=result.normalized_confidence * 0.8,
                    source_roi_ids=[result.roi_id] if result.roi_id else [],
                    validation_status=validation,
                    raw_value=date_str,
                )
        
        return None
    
    def _extract_id_number_heuristic(
        self,
        ocr_results: List[OCRResult]
    ) -> Optional[ParsedField]:
        """Extract ID number using heuristics."""
        # Look for alphanumeric strings of appropriate length
        id_pattern = r"[A-Z0-9]{6,20}"
        
        for result in ocr_results:
            text_upper = result.text.upper()
            matches = re.findall(id_pattern, text_upper)
            
            for match in matches:
                # Should have at least some digits
                digit_count = sum(c.isdigit() for c in match)
                if digit_count >= 3:
                    return ParsedField(
                        field_name="document_number",
                        value=match,
                        confidence=result.normalized_confidence * 0.7,
                        source_roi_ids=[result.roi_id] if result.roi_id else [],
                        validation_status=ValidationStatus.UNCERTAIN,
                        raw_value=match,
                    )
        
        return None
    
    def _extract_all_dates(self, text: str) -> List[str]:
        """Extract all date-like strings from text."""
        patterns = [
            r"\d{1,2}[/\-\.]\d{1,2}[/\-\.]\d{2,4}",
            r"\d{4}[/\-\.]\d{1,2}[/\-\.]\d{1,2}",
        ]
        
        dates = []
        for pattern in patterns:
            matches = re.findall(pattern, text)
            dates.extend(matches)
        
        return dates
    
    def _find_mrz_lines(self, ocr_results: List[OCRResult]) -> List[str]:
        """Find MRZ lines in OCR results."""
        mrz_lines = []
        
        for result in ocr_results:
            text = result.text.upper().replace(" ", "")
            
            # MRZ lines are 44 chars for TD1, 36 for TD2, 44 for TD3 (passport)
            # and contain only A-Z, 0-9, <
            if len(text) >= 30:
                # Check if looks like MRZ
                mrz_chars = set("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<")
                if all(c in mrz_chars for c in text):
                    mrz_lines.append(text)
        
        return mrz_lines
    
    def _parse_mrz(self, lines: List[str]) -> Optional[MRZData]:
        """Parse MRZ lines into structured data."""
        if len(lines) < 2:
            return None
        
        # Handle TD3 format (passport, 2 lines of 44 chars each)
        line1 = lines[0][:44] if len(lines[0]) >= 44 else lines[0]
        line2 = lines[1][:44] if len(lines[1]) >= 44 else lines[1]
        
        try:
            # Line 1: P<ISSUING_COUNTRY<<SURNAME<<GIVEN<NAMES
            doc_type = line1[0]
            country = line1[2:5].replace("<", "")
            
            name_part = line1[5:]
            name_parts = name_part.split("<<")
            surname = name_parts[0].replace("<", " ").strip() if name_parts else ""
            given_names = name_parts[1].replace("<", " ").strip() if len(name_parts) > 1 else ""
            
            # Line 2: DOC_NUM<CHECK<NATIONALITY<DOB<CHECK<SEX<EXPIRY<CHECK<PERSONAL_NUM<CHECK<FINAL_CHECK
            doc_number = line2[0:9].replace("<", "")
            nationality = line2[10:13].replace("<", "")
            dob = line2[13:19]
            sex = line2[20]
            expiry = line2[21:27]
            personal_number = line2[28:42].replace("<", "")
            
            # Validate check digits (simplified - full implementation would verify each)
            check_digits_valid = True  # Simplified for now
            
            return MRZData(
                document_type=doc_type,
                country_code=country,
                surname=surname,
                given_names=given_names,
                document_number=doc_number,
                nationality=nationality,
                date_of_birth=dob,
                sex=sex,
                expiry_date=expiry,
                personal_number=personal_number,
                check_digits_valid=check_digits_valid,
                raw_lines=lines,
            )
        except (IndexError, ValueError):
            return None
    
    def _find_source_rois(
        self,
        value: str,
        ocr_results: List[OCRResult]
    ) -> List[str]:
        """Find ROI IDs that contain the value."""
        source_rois = []
        value_lower = value.lower()
        
        for result in ocr_results:
            if value_lower in result.text.lower():
                if result.roi_id:
                    source_rois.append(result.roi_id)
        
        return source_rois
    
    def _calculate_confidence(
        self,
        value: str,
        ocr_results: List[OCRResult]
    ) -> float:
        """Calculate confidence for extracted value."""
        value_lower = value.lower()
        confidences = []
        
        for result in ocr_results:
            if value_lower in result.text.lower():
                confidences.append(result.normalized_confidence)
        
        if confidences:
            return sum(confidences) / len(confidences)
        
        return 0.5
    
    def _validate_field(
        self,
        field_name: str,
        value: str
    ) -> ValidationStatus:
        """Validate a field value."""
        if not value:
            return ValidationStatus.NOT_FOUND
        
        # Length check
        if len(value) < self.config.min_name_length:
            return ValidationStatus.INVALID
        
        if "name" in field_name and len(value) > self.config.max_name_length:
            return ValidationStatus.INVALID
        
        # Date validation
        if "date" in field_name:
            return self._validate_date(value)
        
        return ValidationStatus.VALID
    
    def _validate_date(self, value: str) -> ValidationStatus:
        """Validate a date value."""
        for fmt in self.config.date_formats:
            try:
                parsed = datetime.strptime(value, fmt)
                # Check reasonable date range
                if 1900 <= parsed.year <= 2100:
                    return ValidationStatus.VALID
            except ValueError:
                continue
        
        # Check if it at least has date-like structure
        if re.match(r"\d{1,2}[/\-\.]\d{1,2}[/\-\.]\d{2,4}", value):
            return ValidationStatus.UNCERTAIN
        
        return ValidationStatus.INVALID
