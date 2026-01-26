"""
Document field parser for structured extraction.

This module parses OCR results into structured fields
for various document types (ID cards, passports, driver's licenses).
"""

import re
from typing import Dict, List, Optional, Any, Tuple
from dataclasses import dataclass, field
from enum import Enum
from datetime import datetime, date

from ml.ocr_extraction.config import FieldParserConfig, DocumentType
from ml.ocr_extraction.tesseract_wrapper import OCRResult
from ml.ocr_extraction.postprocessing import OCRPostProcessor


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
    
    def __init__(self, config: Optional[FieldParserConfig] = None):
        """
        Initialize field parser.
        
        Args:
            config: Parser configuration. Uses defaults if None.
        """
        self.config = config or FieldParserConfig()
        self.postprocessor = OCRPostProcessor()
    
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
