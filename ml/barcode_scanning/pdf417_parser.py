"""
PDF417 Barcode Parser for Driver Licenses.

This module implements parsing of PDF417 barcodes found on driver licenses,
following the AAMVA (American Association of Motor Vehicle Administrators)
standard for encoding license data.

The AAMVA standard defines a structured format with:
- Header containing version and jurisdiction info
- Subfile designators for different data sections
- Data elements with standard field codes (DAQ, DCS, etc.)

Security Note:
    Never log raw barcode data as it contains PII.
    All field values should be hashed before on-chain storage.
"""

import re
import logging
import hashlib
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Any, Tuple, TYPE_CHECKING
from datetime import datetime, date
from enum import Enum

if TYPE_CHECKING:
    import numpy as np

from ml.barcode_scanning.config import PDF417Config, BarcodeType

# Intentionally minimal logging - no PII
logger = logging.getLogger(__name__)


class AAMVAVersion(int, Enum):
    """AAMVA DL/ID Card Design Standard versions."""
    V1 = 1   # Pre-2000
    V2 = 2   # 2000
    V3 = 3   # 2005
    V4 = 4   # 2009
    V5 = 5   # 2010
    V6 = 6   # 2011
    V7 = 7   # 2012
    V8 = 8   # 2013
    V9 = 9   # 2016
    V10 = 10  # 2020


@dataclass
class AAMVAField:
    """A parsed AAMVA data element."""
    code: str           # 3-char field code (e.g., "DAQ")
    value: str          # Decoded value
    description: str    # Human-readable description
    is_required: bool   # Whether field is required in standard
    raw_value: str = "" # Original encoded value
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "code": self.code,
            "value": self.value,
            "description": self.description,
            "is_required": self.is_required,
        }


@dataclass
class PDF417Data:
    """
    Parsed PDF417 barcode data from driver license.
    
    Contains structured fields following AAMVA standard.
    """
    # Identification
    document_number: str = ""          # DAQ
    issuing_jurisdiction: str = ""     # DCG (country) + DAJ (state)
    
    # Personal info
    first_name: str = ""               # DAC
    last_name: str = ""                # DCS
    middle_name: str = ""              # DAD
    suffix: str = ""                   # DCU
    full_name: str = ""                # Constructed
    
    # Demographics
    date_of_birth: str = ""            # DBB (MMDDCCYY or CCYYMMDD)
    sex: str = ""                      # DBC (1=M, 2=F, 9=Not specified)
    height: str = ""                   # DAU
    weight: str = ""                   # DAW
    eye_color: str = ""                # DAY
    hair_color: str = ""               # DAZ
    
    # Address
    street_address: str = ""           # DAG
    street_address_2: str = ""         # DAH
    city: str = ""                     # DAI
    state: str = ""                    # DAJ
    postal_code: str = ""              # DAK
    country: str = ""                  # DCG
    
    # Document info
    issue_date: str = ""               # DBD
    expiry_date: str = ""              # DBA
    document_class: str = ""           # DCA
    restrictions: str = ""             # DCB
    endorsements: str = ""             # DCD
    
    # Metadata
    aamva_version: int = 0
    jurisdiction_version: int = 0
    issuer_id: str = ""
    
    # All parsed fields
    fields: Dict[str, AAMVAField] = field(default_factory=dict)
    
    # Parsing success
    success: bool = True
    error_message: Optional[str] = None
    confidence: float = 0.0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "document_number": self.document_number,
            "first_name": self.first_name,
            "last_name": self.last_name,
            "middle_name": self.middle_name,
            "full_name": self.full_name,
            "date_of_birth": self.date_of_birth,
            "sex": self.sex,
            "street_address": self.street_address,
            "city": self.city,
            "state": self.state,
            "postal_code": self.postal_code,
            "country": self.country,
            "issue_date": self.issue_date,
            "expiry_date": self.expiry_date,
            "aamva_version": self.aamva_version,
            "success": self.success,
            "confidence": self.confidence,
        }
    
    def get_identity_fields(self) -> Dict[str, str]:
        """
        Get fields relevant for identity verification.
        
        Returns normalized field names for cross-validation.
        """
        return {
            "full_name": self.full_name or f"{self.first_name} {self.last_name}".strip(),
            "surname": self.last_name,
            "given_names": f"{self.first_name} {self.middle_name}".strip(),
            "date_of_birth": self.date_of_birth,
            "document_number": self.document_number,
            "expiry_date": self.expiry_date,
            "sex": self._normalize_sex(self.sex),
            "address": self._format_address(),
        }
    
    def _normalize_sex(self, code: str) -> str:
        """Normalize sex code to standard format."""
        if code == "1":
            return "M"
        elif code == "2":
            return "F"
        elif code in ("M", "F"):
            return code
        return ""
    
    def _format_address(self) -> str:
        """Format full address."""
        parts = [
            self.street_address,
            self.street_address_2,
            self.city,
            self.state,
            self.postal_code,
        ]
        return ", ".join(p for p in parts if p)


class PDF417Parser:
    """
    Parser for PDF417 barcodes on driver licenses.
    
    Implements AAMVA DL/ID Card Design Standard parsing
    for versions 1-10.
    """
    
    # AAMVA field codes and descriptions
    FIELD_DEFINITIONS = {
        # Required fields
        "DAQ": ("Document Number", True),
        "DCS": ("Last Name", True),
        "DAC": ("First Name", True),
        "DBB": ("Date of Birth", True),
        "DBC": ("Sex", True),
        "DAJ": ("State", True),
        "DBA": ("Expiry Date", True),
        
        # Common optional fields
        "DAD": ("Middle Name", False),
        "DCU": ("Suffix", False),
        "DAG": ("Street Address", False),
        "DAH": ("Street Address 2", False),
        "DAI": ("City", False),
        "DAK": ("Postal Code", False),
        "DCG": ("Country", False),
        "DAU": ("Height", False),
        "DAW": ("Weight", False),
        "DAY": ("Eye Color", False),
        "DAZ": ("Hair Color", False),
        "DBD": ("Issue Date", False),
        "DCA": ("Document Class", False),
        "DCB": ("Restrictions", False),
        "DCD": ("Endorsements", False),
        "DDD": ("Document Discriminator", False),
        "DDK": ("Organ Donor", False),
        "DDL": ("Veteran", False),
        
        # Name variations (older versions)
        "DCT": ("First Name (alt)", False),
        "DBN": ("Last Name (alt)", False),
        "DBP": ("First Name (alt2)", False),
    }
    
    # Compliance indicator
    COMPLIANCE_INDICATOR = "@"
    
    # Record separator
    RECORD_SEPARATOR = "\x1e"
    
    # Segment terminator
    SEGMENT_TERMINATOR = "\r"
    
    # Data element separator
    DATA_SEPARATOR = "\n"
    
    def __init__(self, config: Optional[PDF417Config] = None):
        """
        Initialize PDF417 parser.
        
        Args:
            config: Parser configuration. Uses defaults if None.
        """
        self.config = config or PDF417Config()
        self._barcode_reader = None
    
    def _get_reader(self):
        """Lazy-load pyzbar barcode reader."""
        if self._barcode_reader is None:
            try:
                from pyzbar import pyzbar
                self._barcode_reader = pyzbar
            except ImportError:
                logger.error("pyzbar not installed. Install with: pip install pyzbar")
                raise ImportError("pyzbar is required for PDF417 parsing")
        return self._barcode_reader
    
    def scan_image(self, image: Any) -> List[Tuple[bytes, Any]]:
        """
        Scan image for PDF417 barcodes.
        
        Args:
            image: Input image as numpy array (BGR or grayscale).
            
        Returns:
            List of (barcode_data, barcode_info) tuples.
        """
        pyzbar = self._get_reader()
        
        # Convert to grayscale if needed
        if len(image.shape) == 3:
            import cv2
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        else:
            gray = image
        
        # Decode barcodes
        barcodes = pyzbar.decode(gray, symbols=[pyzbar.ZBarSymbol.PDF417])
        
        results = []
        for barcode in barcodes:
            results.append((barcode.data, barcode))
        
        logger.debug(f"Found {len(results)} PDF417 barcode(s)")
        return results
    
    def parse(self, raw_data: bytes) -> PDF417Data:
        """
        Parse PDF417 barcode data into structured fields.
        
        Args:
            raw_data: Raw bytes from barcode scan.
            
        Returns:
            PDF417Data with parsed fields.
        """
        result = PDF417Data()
        
        try:
            # Decode bytes to string
            try:
                data = raw_data.decode('utf-8')
            except UnicodeDecodeError:
                data = raw_data.decode('latin-1')
            
            # Parse AAMVA header
            header_info = self._parse_header(data)
            if header_info:
                result.aamva_version = header_info.get("version", 0)
                result.jurisdiction_version = header_info.get("jur_version", 0)
                result.issuer_id = header_info.get("issuer_id", "")
            
            # Parse data elements
            fields = self._parse_data_elements(data)
            result.fields = fields
            
            # Map to structured fields
            self._map_fields(result, fields)
            
            # Calculate confidence
            result.confidence = self._calculate_confidence(result)
            result.success = True
            
        except Exception as e:
            logger.error(f"PDF417 parsing failed: {type(e).__name__}")
            result.success = False
            result.error_message = f"Parsing error: {type(e).__name__}"
            result.confidence = 0.0
        
        return result
    
    def _parse_header(self, data: str) -> Optional[Dict[str, Any]]:
        """
        Parse AAMVA file header.
        
        Header format:
        @[ANSI/AAMVA indicator][IIN][AAMVAVersion][JurVersion][...]
        """
        if not data.startswith(self.COMPLIANCE_INDICATOR):
            # Try to find header in data
            idx = data.find(self.COMPLIANCE_INDICATOR)
            if idx == -1:
                return None
            data = data[idx:]
        
        if len(data) < 21:
            return None
        
        try:
            # Skip @ and look for ANSI/AAMVA
            header_match = re.match(
                r'@\s*ANSI\s+(\d{6})(\d{2})(\d{2})',
                data
            )
            
            if not header_match:
                # Try alternate format
                header_match = re.match(
                    r'@\s*AAMVA\s+(\d{6})(\d{2})(\d{2})',
                    data
                )
            
            if header_match:
                return {
                    "issuer_id": header_match.group(1),
                    "version": int(header_match.group(2)),
                    "jur_version": int(header_match.group(3)),
                }
        except (ValueError, IndexError):
            pass
        
        return None
    
    def _parse_data_elements(self, data: str) -> Dict[str, AAMVAField]:
        """
        Parse data elements from barcode data.
        
        Each element is: [3-char code][value][separator]
        """
        fields = {}
        
        # Split by common separators
        lines = re.split(r'[\r\n]+', data)
        
        for line in lines:
            # Skip header lines
            if line.startswith('@') or 'ANSI' in line or 'AAMVA' in line:
                continue
            
            # Look for field codes (3 uppercase letters or DAQ, etc.)
            matches = re.findall(r'([A-Z]{2,3})([^\r\n]*)', line)
            
            for code, value in matches:
                if code in self.FIELD_DEFINITIONS:
                    desc, required = self.FIELD_DEFINITIONS[code]
                    value = value.strip()
                    
                    if value:  # Only add non-empty fields
                        fields[code] = AAMVAField(
                            code=code,
                            value=self._clean_value(value),
                            description=desc,
                            is_required=required,
                            raw_value=value,
                        )
        
        return fields
    
    def _clean_value(self, value: str) -> str:
        """Clean field value of control characters."""
        # Remove control characters
        value = re.sub(r'[\x00-\x1f\x7f]', '', value)
        # Normalize whitespace
        value = ' '.join(value.split())
        return value.strip()
    
    def _map_fields(self, result: PDF417Data, fields: Dict[str, AAMVAField]) -> None:
        """Map parsed fields to result structure."""
        
        # Document number
        if "DAQ" in fields:
            result.document_number = fields["DAQ"].value
        
        # Names
        if "DCS" in fields:
            result.last_name = fields["DCS"].value
        elif "DBN" in fields:
            result.last_name = fields["DBN"].value
        
        if "DAC" in fields:
            result.first_name = fields["DAC"].value
        elif "DCT" in fields:
            result.first_name = fields["DCT"].value
        elif "DBP" in fields:
            result.first_name = fields["DBP"].value
        
        if "DAD" in fields:
            result.middle_name = fields["DAD"].value
        
        if "DCU" in fields:
            result.suffix = fields["DCU"].value
        
        # Construct full name
        name_parts = [result.first_name, result.middle_name, result.last_name]
        if result.suffix:
            name_parts.append(result.suffix)
        result.full_name = " ".join(p for p in name_parts if p)
        
        # Demographics
        if "DBB" in fields:
            result.date_of_birth = self._parse_date(fields["DBB"].value)
        
        if "DBC" in fields:
            result.sex = fields["DBC"].value
        
        if "DAU" in fields:
            result.height = fields["DAU"].value
        
        if "DAW" in fields:
            result.weight = fields["DAW"].value
        
        if "DAY" in fields:
            result.eye_color = fields["DAY"].value
        
        if "DAZ" in fields:
            result.hair_color = fields["DAZ"].value
        
        # Address
        if "DAG" in fields:
            result.street_address = fields["DAG"].value
        
        if "DAH" in fields:
            result.street_address_2 = fields["DAH"].value
        
        if "DAI" in fields:
            result.city = fields["DAI"].value
        
        if "DAJ" in fields:
            result.state = fields["DAJ"].value
        
        if "DAK" in fields:
            result.postal_code = fields["DAK"].value
        
        if "DCG" in fields:
            result.country = fields["DCG"].value
        
        # Document info
        if "DBD" in fields:
            result.issue_date = self._parse_date(fields["DBD"].value)
        
        if "DBA" in fields:
            result.expiry_date = self._parse_date(fields["DBA"].value)
        
        if "DCA" in fields:
            result.document_class = fields["DCA"].value
        
        if "DCB" in fields:
            result.restrictions = fields["DCB"].value
        
        if "DCD" in fields:
            result.endorsements = fields["DCD"].value
    
    def _parse_date(self, value: str) -> str:
        """
        Parse AAMVA date format to ISO format.
        
        AAMVA uses MMDDCCYY or CCYYMMDD depending on version.
        """
        value = value.strip()
        
        if len(value) == 8:
            # Try MMDDCCYY first
            try:
                dt = datetime.strptime(value, "%m%d%Y")
                return dt.strftime("%Y-%m-%d")
            except ValueError:
                pass
            
            # Try CCYYMMDD
            try:
                dt = datetime.strptime(value, "%Y%m%d")
                return dt.strftime("%Y-%m-%d")
            except ValueError:
                pass
        
        # Return as-is if parsing fails
        return value
    
    def _calculate_confidence(self, result: PDF417Data) -> float:
        """
        Calculate parsing confidence score.
        
        Based on presence and validity of fields.
        """
        if not result.fields:
            return 0.0
        
        # Count required fields present
        required_present = sum(
            1 for code in self.config.critical_fields
            if code in result.fields
        )
        required_total = len(self.config.critical_fields)
        
        # Base score from required field coverage
        base_score = required_present / required_total if required_total > 0 else 0.0
        
        # Bonus for optional fields
        optional_present = sum(
            1 for code in self.config.optional_fields
            if code in result.fields
        )
        optional_bonus = min(0.1, optional_present * 0.02)
        
        # Penalty for missing critical fields
        if not result.document_number:
            base_score -= 0.2
        if not result.last_name:
            base_score -= 0.15
        if not result.date_of_birth:
            base_score -= 0.15
        
        return max(0.0, min(1.0, base_score + optional_bonus))
    
    def hash_fields(self, result: PDF417Data, salt: str) -> Dict[str, str]:
        """
        Create hashed versions of fields for on-chain storage.
        
        Args:
            result: Parsed PDF417 data.
            salt: Salt for hashing (should be unique per identity).
            
        Returns:
            Dictionary of field name to hash.
        """
        hashes = {}
        identity_fields = result.get_identity_fields()
        
        for field_name, value in identity_fields.items():
            if value:
                # SHA-256 hash with salt
                salted = f"{salt}:{field_name}:{value}".encode('utf-8')
                hash_value = hashlib.sha256(salted).hexdigest()
                hashes[f"barcode_{field_name}"] = hash_value
        
        return hashes
