"""
MRZ (Machine Readable Zone) Parser for Passports and ID Cards.

This module implements parsing of ICAO Doc 9303 compliant MRZ zones:
- TD1: 3 lines x 30 characters (ID cards)
- TD2: 2 lines x 36 characters (older ID cards, some visas)
- TD3: 2 lines x 44 characters (passports)

The parser validates check digits and extracts structured identity data.

Security Note:
    Never log raw MRZ data as it contains PII.
    All field values should be hashed before on-chain storage.
"""

import re
import logging
import hashlib
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any, Tuple, TYPE_CHECKING
from datetime import datetime
from enum import Enum

if TYPE_CHECKING:
    import numpy as np

from ml.barcode_scanning.config import MRZConfig, MRZFormat

# Intentionally minimal logging - no PII
logger = logging.getLogger(__name__)


# ICAO check digit weights
CHECK_DIGIT_WEIGHTS = [7, 3, 1]

# Character to value mapping for check digits
CHAR_VALUES = {
    '<': 0,
    '0': 0, '1': 1, '2': 2, '3': 3, '4': 4,
    '5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
    'A': 10, 'B': 11, 'C': 12, 'D': 13, 'E': 14,
    'F': 15, 'G': 16, 'H': 17, 'I': 18, 'J': 19,
    'K': 20, 'L': 21, 'M': 22, 'N': 23, 'O': 24,
    'P': 25, 'Q': 26, 'R': 27, 'S': 28, 'T': 29,
    'U': 30, 'V': 31, 'W': 32, 'X': 33, 'Y': 34,
    'Z': 35,
}


@dataclass
class MRZCheckDigit:
    """Result of check digit validation."""
    field_name: str
    expected: str
    calculated: str
    is_valid: bool
    value_checked: str
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "field_name": self.field_name,
            "expected": self.expected,
            "calculated": self.calculated,
            "is_valid": self.is_valid,
        }


@dataclass
class MRZLine:
    """A single line of MRZ data."""
    line_number: int
    raw_text: str
    clean_text: str
    expected_length: int
    is_valid_length: bool
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "line_number": self.line_number,
            "length": len(self.clean_text),
            "expected_length": self.expected_length,
            "is_valid_length": self.is_valid_length,
        }


@dataclass
class MRZData:
    """
    Parsed MRZ data from passport or ID card.
    
    Contains structured fields following ICAO Doc 9303 standard.
    """
    # Document info
    document_type: str = ""        # P = passport, I = ID card, etc.
    document_subtype: str = ""     # Optional second character
    issuing_country: str = ""      # 3-letter country code
    document_number: str = ""      # Document/passport number
    
    # Personal info
    surname: str = ""
    given_names: str = ""
    full_name: str = ""            # Constructed
    
    # Demographics
    date_of_birth: str = ""        # YYMMDD -> YYYY-MM-DD
    sex: str = ""                  # M, F, or <
    nationality: str = ""          # 3-letter country code
    
    # Document validity
    expiry_date: str = ""          # YYMMDD -> YYYY-MM-DD
    
    # Optional data
    personal_number: str = ""      # Optional field
    optional_data_1: str = ""
    optional_data_2: str = ""
    
    # MRZ format
    format: MRZFormat = MRZFormat.UNKNOWN
    
    # Raw MRZ lines
    lines: List[MRZLine] = field(default_factory=list)
    
    # Check digit results
    check_digits: List[MRZCheckDigit] = field(default_factory=list)
    all_check_digits_valid: bool = False
    
    # Parsing success
    success: bool = True
    error_message: Optional[str] = None
    confidence: float = 0.0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "document_type": self.document_type,
            "issuing_country": self.issuing_country,
            "document_number": self.document_number,
            "surname": self.surname,
            "given_names": self.given_names,
            "full_name": self.full_name,
            "date_of_birth": self.date_of_birth,
            "sex": self.sex,
            "nationality": self.nationality,
            "expiry_date": self.expiry_date,
            "format": self.format.value,
            "all_check_digits_valid": self.all_check_digits_valid,
            "success": self.success,
            "confidence": self.confidence,
        }
    
    def get_identity_fields(self) -> Dict[str, str]:
        """
        Get fields relevant for identity verification.
        
        Returns normalized field names for cross-validation.
        """
        return {
            "full_name": self.full_name,
            "surname": self.surname,
            "given_names": self.given_names,
            "date_of_birth": self.date_of_birth,
            "document_number": self.document_number,
            "expiry_date": self.expiry_date,
            "sex": self.sex,
            "nationality": self.nationality,
        }


class MRZParser:
    """
    Parser for Machine Readable Zone data.
    
    Implements ICAO Doc 9303 parsing for TD1, TD2, and TD3 formats.
    """
    
    # Format specifications
    FORMAT_SPECS = {
        MRZFormat.TD1: {
            "lines": 3,
            "chars_per_line": 30,
            "total_chars": 90,
        },
        MRZFormat.TD2: {
            "lines": 2,
            "chars_per_line": 36,
            "total_chars": 72,
        },
        MRZFormat.TD3: {
            "lines": 2,
            "chars_per_line": 44,
            "total_chars": 88,
        },
    }
    
    # Valid MRZ characters
    VALID_CHARS = set("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<")
    
    def __init__(self, config: Optional[MRZConfig] = None):
        """
        Initialize MRZ parser.
        
        Args:
            config: Parser configuration. Uses defaults if None.
        """
        self.config = config or MRZConfig()
    
    def detect_format(self, lines: List[str]) -> MRZFormat:
        """
        Detect MRZ format from lines.
        
        Args:
            lines: List of MRZ text lines.
            
        Returns:
            Detected MRZ format.
        """
        if not lines:
            return MRZFormat.UNKNOWN
        
        num_lines = len(lines)
        
        # Clean and measure first line
        first_line = self._clean_mrz_line(lines[0])
        line_length = len(first_line)
        
        if num_lines >= 3 and 28 <= line_length <= 32:
            return MRZFormat.TD1
        elif num_lines >= 2 and 34 <= line_length <= 38:
            return MRZFormat.TD2
        elif num_lines >= 2 and 42 <= line_length <= 46:
            return MRZFormat.TD3
        
        return MRZFormat.UNKNOWN
    
    def parse(self, mrz_text: str) -> MRZData:
        """
        Parse MRZ text into structured data.
        
        Args:
            mrz_text: Raw MRZ text (may contain multiple lines).
            
        Returns:
            MRZData with parsed fields.
        """
        result = MRZData()
        
        try:
            # Split into lines
            lines = self._extract_lines(mrz_text)
            
            if not lines:
                result.success = False
                result.error_message = "No valid MRZ lines found"
                return result
            
            # Detect format
            result.format = self.detect_format(lines)
            
            if result.format == MRZFormat.UNKNOWN:
                result.success = False
                result.error_message = "Could not detect MRZ format"
                return result
            
            # Clean and validate lines
            spec = self.FORMAT_SPECS[result.format]
            mrz_lines = []
            
            for i, line in enumerate(lines[:spec["lines"]]):
                clean = self._clean_mrz_line(line)
                mrz_line = MRZLine(
                    line_number=i + 1,
                    raw_text=line,
                    clean_text=clean,
                    expected_length=spec["chars_per_line"],
                    is_valid_length=len(clean) == spec["chars_per_line"],
                )
                mrz_lines.append(mrz_line)
            
            result.lines = mrz_lines
            
            # Parse based on format
            if result.format == MRZFormat.TD1:
                self._parse_td1(result, mrz_lines)
            elif result.format == MRZFormat.TD2:
                self._parse_td2(result, mrz_lines)
            elif result.format == MRZFormat.TD3:
                self._parse_td3(result, mrz_lines)
            
            # Validate check digits
            self._validate_check_digits(result)
            
            # Calculate confidence
            result.confidence = self._calculate_confidence(result)
            result.success = True
            
        except Exception as e:
            logger.error(f"MRZ parsing failed: {type(e).__name__}")
            result.success = False
            result.error_message = f"Parsing error: {type(e).__name__}"
            result.confidence = 0.0
        
        return result
    
    def parse_from_image(self, image: Any) -> MRZData:
        """
        Detect and parse MRZ from image.
        
        Uses OCR to extract MRZ text, then parses it.
        
        Args:
            image: Input image as numpy array.
            
        Returns:
            MRZData with parsed fields.
        """
        # Use Tesseract for MRZ OCR
        try:
            import pytesseract
            import cv2
            
            # Convert to grayscale
            if len(image.shape) == 3:
                gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
            else:
                gray = image
            
            # Apply thresholding
            _, binary = cv2.threshold(gray, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
            
            # OCR with MRZ-optimized settings
            text = pytesseract.image_to_string(
                binary,
                config='--psm 6 -c tessedit_char_whitelist=ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<'
            )
            
            return self.parse(text)
            
        except ImportError:
            result = MRZData()
            result.success = False
            result.error_message = "pytesseract not installed"
            return result
    
    def _extract_lines(self, text: str) -> List[str]:
        """Extract potential MRZ lines from text."""
        lines = []
        
        for line in text.split('\n'):
            line = line.strip()
            
            # Check if line looks like MRZ (mostly valid chars)
            if len(line) >= self.config.min_line_length:
                clean = self._clean_mrz_line(line)
                valid_ratio = sum(1 for c in clean if c in self.VALID_CHARS) / len(clean) if clean else 0
                
                if valid_ratio > 0.9:
                    lines.append(line)
        
        return lines
    
    def _clean_mrz_line(self, line: str) -> str:
        """Clean and normalize MRZ line."""
        # Uppercase
        line = line.upper()
        
        # Filter to valid MRZ characters only
        # Note: We do NOT convert O->0 or I->1 here as they are both valid
        # MRZ characters. OCR correction should happen at the OCR stage.
        result = []
        for c in line:
            if c in self.VALID_CHARS:
                result.append(c)
            elif c == ' ':
                continue  # Skip spaces
            # Skip other invalid chars
        
        return ''.join(result)
    
    def _parse_td1(self, result: MRZData, lines: List[MRZLine]) -> None:
        """
        Parse TD1 format (3 lines x 30 chars).
        
        Line 1: Document type, issuing state, document number
        Line 2: Birth date, sex, expiry date, nationality, optional
        Line 3: Names
        """
        if len(lines) < 3:
            return
        
        line1 = lines[0].clean_text.ljust(30, '<')
        line2 = lines[1].clean_text.ljust(30, '<')
        line3 = lines[2].clean_text.ljust(30, '<')
        
        # Line 1
        result.document_type = line1[0:1].replace('<', '')
        result.document_subtype = line1[1:2].replace('<', '')
        result.issuing_country = line1[2:5].replace('<', '')
        result.document_number = line1[5:14].replace('<', '')
        # Check digit at 14
        result.optional_data_1 = line1[15:30].replace('<', '')
        
        # Line 2
        dob_raw = line2[0:6]
        result.date_of_birth = self._format_date(dob_raw)
        # Check digit at 6
        result.sex = line2[7:8].replace('<', '')
        expiry_raw = line2[8:14]
        result.expiry_date = self._format_date(expiry_raw)
        # Check digit at 14
        result.nationality = line2[15:18].replace('<', '')
        result.optional_data_2 = line2[18:29].replace('<', '')
        # Overall check digit at 29
        
        # Line 3: Names
        names_raw = line3[0:30]
        self._parse_names(result, names_raw)
    
    def _parse_td2(self, result: MRZData, lines: List[MRZLine]) -> None:
        """
        Parse TD2 format (2 lines x 36 chars).
        
        Line 1: Document type, issuing state, names
        Line 2: Document number, nationality, birth, sex, expiry, optional
        """
        if len(lines) < 2:
            return
        
        line1 = lines[0].clean_text.ljust(36, '<')
        line2 = lines[1].clean_text.ljust(36, '<')
        
        # Line 1
        result.document_type = line1[0:1].replace('<', '')
        result.document_subtype = line1[1:2].replace('<', '')
        result.issuing_country = line1[2:5].replace('<', '')
        names_raw = line1[5:36]
        self._parse_names(result, names_raw)
        
        # Line 2
        result.document_number = line2[0:9].replace('<', '')
        # Check digit at 9
        result.nationality = line2[10:13].replace('<', '')
        dob_raw = line2[13:19]
        result.date_of_birth = self._format_date(dob_raw)
        # Check digit at 19
        result.sex = line2[20:21].replace('<', '')
        expiry_raw = line2[21:27]
        result.expiry_date = self._format_date(expiry_raw)
        # Check digit at 27
        result.optional_data_1 = line2[28:35].replace('<', '')
        # Overall check digit at 35
    
    def _parse_td3(self, result: MRZData, lines: List[MRZLine]) -> None:
        """
        Parse TD3 format (2 lines x 44 chars).
        
        This is the standard passport format.
        
        Line 1: Document type, issuing state, names
        Line 2: Passport number, nationality, birth, sex, expiry, personal number
        """
        if len(lines) < 2:
            return
        
        line1 = lines[0].clean_text.ljust(44, '<')
        line2 = lines[1].clean_text.ljust(44, '<')
        
        # Line 1
        result.document_type = line1[0:1].replace('<', '')
        result.document_subtype = line1[1:2].replace('<', '')
        result.issuing_country = line1[2:5].replace('<', '')
        names_raw = line1[5:44]
        self._parse_names(result, names_raw)
        
        # Line 2
        result.document_number = line2[0:9].replace('<', '')
        # Check digit at 9
        result.nationality = line2[10:13].replace('<', '')
        dob_raw = line2[13:19]
        result.date_of_birth = self._format_date(dob_raw)
        # Check digit at 19
        result.sex = line2[20:21].replace('<', '')
        expiry_raw = line2[21:27]
        result.expiry_date = self._format_date(expiry_raw)
        # Check digit at 27
        result.personal_number = line2[28:42].replace('<', '')
        # Check digit at 42
        # Overall check digit at 43
    
    def _parse_names(self, result: MRZData, names_raw: str) -> None:
        """Parse name field (SURNAME<<GIVEN<NAMES)."""
        # Split on double filler
        parts = names_raw.split('<<')
        
        if len(parts) >= 1:
            result.surname = parts[0].replace('<', ' ').strip()
        
        if len(parts) >= 2:
            result.given_names = parts[1].replace('<', ' ').strip()
        
        # Construct full name
        if result.given_names and result.surname:
            result.full_name = f"{result.given_names} {result.surname}"
        elif result.surname:
            result.full_name = result.surname
        else:
            result.full_name = result.given_names
    
    def _format_date(self, yymmdd: str) -> str:
        """
        Convert YYMMDD to YYYY-MM-DD.
        
        Uses ICAO rules: YY < 50 = 20YY, YY >= 50 = 19YY
        """
        if len(yymmdd) != 6:
            return yymmdd
        
        try:
            yy = int(yymmdd[0:2])
            mm = yymmdd[2:4]
            dd = yymmdd[4:6]
            
            # ICAO rule
            year = 2000 + yy if yy < 50 else 1900 + yy
            
            return f"{year}-{mm}-{dd}"
        except ValueError:
            return yymmdd
    
    def _validate_check_digits(self, result: MRZData) -> None:
        """Validate MRZ check digits."""
        if not self.config.validate_check_digits:
            result.all_check_digits_valid = True
            return
        
        check_results = []
        
        if result.format == MRZFormat.TD3 and len(result.lines) >= 2:
            line2 = result.lines[1].clean_text.ljust(44, '<')
            
            # Document number check (position 9)
            doc_check = self._validate_single_check_digit(
                "document_number",
                line2[0:9],
                line2[9:10]
            )
            check_results.append(doc_check)
            
            # Date of birth check (position 19)
            dob_check = self._validate_single_check_digit(
                "date_of_birth",
                line2[13:19],
                line2[19:20]
            )
            check_results.append(dob_check)
            
            # Expiry date check (position 27)
            exp_check = self._validate_single_check_digit(
                "expiry_date",
                line2[21:27],
                line2[27:28]
            )
            check_results.append(exp_check)
            
            # Personal number check (position 42)
            pn_check = self._validate_single_check_digit(
                "personal_number",
                line2[28:42],
                line2[42:43]
            )
            check_results.append(pn_check)
            
            # Overall check (position 43)
            composite = line2[0:10] + line2[13:20] + line2[21:43]
            overall_check = self._validate_single_check_digit(
                "composite",
                composite,
                line2[43:44]
            )
            check_results.append(overall_check)
        
        elif result.format == MRZFormat.TD1 and len(result.lines) >= 2:
            line1 = result.lines[0].clean_text.ljust(30, '<')
            line2 = result.lines[1].clean_text.ljust(30, '<')
            
            # Document number check (position 14 of line 1)
            doc_check = self._validate_single_check_digit(
                "document_number",
                line1[5:14],
                line1[14:15]
            )
            check_results.append(doc_check)
            
            # Date of birth check (position 6 of line 2)
            dob_check = self._validate_single_check_digit(
                "date_of_birth",
                line2[0:6],
                line2[6:7]
            )
            check_results.append(dob_check)
            
            # Expiry date check (position 14 of line 2)
            exp_check = self._validate_single_check_digit(
                "expiry_date",
                line2[8:14],
                line2[14:15]
            )
            check_results.append(exp_check)
        
        result.check_digits = check_results
        result.all_check_digits_valid = all(c.is_valid for c in check_results)
    
    def _validate_single_check_digit(
        self,
        field_name: str,
        value: str,
        expected_digit: str
    ) -> MRZCheckDigit:
        """Validate a single check digit."""
        calculated = self._calculate_check_digit(value)
        
        return MRZCheckDigit(
            field_name=field_name,
            expected=expected_digit,
            calculated=calculated,
            is_valid=calculated == expected_digit,
            value_checked=value,
        )
    
    def _calculate_check_digit(self, value: str) -> str:
        """
        Calculate ICAO check digit.
        
        Algorithm: Sum of (char_value * weight) mod 10
        Weights cycle: 7, 3, 1
        """
        total = 0
        
        for i, char in enumerate(value):
            char_val = CHAR_VALUES.get(char.upper(), 0)
            weight = CHECK_DIGIT_WEIGHTS[i % 3]
            total += char_val * weight
        
        return str(total % 10)
    
    def _calculate_confidence(self, result: MRZData) -> float:
        """Calculate parsing confidence score."""
        if not result.lines:
            return 0.0
        
        # Base score from line validity
        valid_lines = sum(1 for line in result.lines if line.is_valid_length)
        line_score = valid_lines / len(result.lines) if result.lines else 0.0
        
        # Check digit bonus
        if result.check_digits:
            valid_checks = sum(1 for c in result.check_digits if c.is_valid)
            check_score = valid_checks / len(result.check_digits)
        else:
            check_score = 0.5  # No validation = uncertain
        
        # Field presence bonus
        field_score = 0.0
        if result.document_number:
            field_score += 0.2
        if result.surname:
            field_score += 0.2
        if result.date_of_birth:
            field_score += 0.2
        if result.expiry_date:
            field_score += 0.1
        if result.nationality:
            field_score += 0.1
        
        # Weighted combination
        confidence = (
            line_score * 0.3 +
            check_score * 0.4 +
            min(1.0, field_score) * 0.3
        )
        
        return max(0.0, min(1.0, confidence))
    
    def hash_fields(self, result: MRZData, salt: str) -> Dict[str, str]:
        """
        Create hashed versions of fields for on-chain storage.
        
        Args:
            result: Parsed MRZ data.
            salt: Salt for hashing.
            
        Returns:
            Dictionary of field name to hash.
        """
        hashes = {}
        identity_fields = result.get_identity_fields()
        
        for field_name, value in identity_fields.items():
            if value:
                salted = f"{salt}:{field_name}:{value}".encode('utf-8')
                hash_value = hashlib.sha256(salted).hexdigest()
                hashes[f"mrz_{field_name}"] = hash_value
        
        return hashes
