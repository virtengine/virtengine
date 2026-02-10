"""
Document type configuration loader.

Loads document type definitions from YAML configuration and provides
utilities for field validation and extraction.

This module enables configurable document field definitions instead of
hardcoded field patterns, supporting any country's identity documents.

Task Reference: VE-3047 - Generalize Document Type Configuration

Example Usage:
    from ml.ocr_extraction.document_config import DocumentConfigLoader
    
    # Get singleton instance
    loader = DocumentConfigLoader.get_instance()
    
    # Get document type definition
    passport = loader.get_document_type("passport")
    print(passport.has_mrz)  # True
    
    # Validate a field
    is_valid, error = loader.validate_field("passport", "nationality", "USA")
    
    # List all available document types
    types = loader.list_document_types()
"""

import re
import logging
from pathlib import Path
from typing import Dict, List, Optional, Any, Tuple, Set
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum

import yaml

logger = logging.getLogger(__name__)


class PositionHint(str, Enum):
    """Position hints for field location on document."""
    TOP = "top"
    TOP_THIRD = "top_third"
    UPPER_THIRD = "upper_third"
    MIDDLE = "middle"
    LOWER_THIRD = "lower_third"
    BOTTOM_THIRD = "bottom_third"
    BOTTOM = "bottom"
    LEFT_HALF = "left_half"
    RIGHT_HALF = "right_half"


class MRZFormat(str, Enum):
    """MRZ format types per ICAO 9303."""
    TD1 = "TD1"  # 3 lines x 30 chars (ID cards)
    TD2 = "TD2"  # 2 lines x 36 chars (older IDs)
    TD3 = "TD3"  # 2 lines x 44 chars (passports)


@dataclass
class FieldDefinition:
    """
    Definition of a document field.
    
    Attributes:
        name: Field identifier
        required: Whether field is mandatory
        pattern: Regex pattern for validation
        format: Date/time format string
        max_length: Maximum character length
        label_text: List of label strings to look for in OCR
        position_hint: Hint for field location on document
        has_check_digit: Whether field includes check digit
        check_digit_algorithm: Algorithm for check digit validation
        description: Human-readable description
        barcode_element: AAMVA barcode element code
    """
    name: str
    required: bool = False
    pattern: Optional[str] = None
    format: Optional[str] = None
    max_length: Optional[int] = None
    label_text: Optional[List[str]] = None
    position_hint: Optional[str] = None
    has_check_digit: bool = False
    check_digit_algorithm: Optional[str] = None
    description: Optional[str] = None
    barcode_element: Optional[str] = None
    
    # Compiled pattern for efficient matching
    _compiled_pattern: Optional[re.Pattern] = field(default=None, repr=False)
    
    def __post_init__(self):
        """Compile regex pattern if provided."""
        if self.pattern:
            try:
                self._compiled_pattern = re.compile(self.pattern)
            except re.error as e:
                logger.warning(f"Invalid pattern for field {self.name}: {e}")
                self._compiled_pattern = None
    
    def matches_pattern(self, value: str) -> bool:
        """Check if value matches the field pattern."""
        if not self._compiled_pattern:
            return True  # No pattern = any value accepted
        return bool(self._compiled_pattern.match(value))
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "name": self.name,
            "required": self.required,
            "pattern": self.pattern,
            "format": self.format,
            "max_length": self.max_length,
            "label_text": self.label_text,
            "position_hint": self.position_hint,
            "has_check_digit": self.has_check_digit,
            "description": self.description,
        }


@dataclass
class DocumentTypeDefinition:
    """
    Definition of a document type.
    
    Represents a complete document type (e.g., passport, driver's license)
    with all its fields and parsing metadata.
    
    Attributes:
        type_id: Unique identifier for document type
        name: Human-readable name
        standard: Standard followed (e.g., "ICAO 9303", "AAMVA")
        has_mrz: Whether document has machine-readable zone
        mrz_format: MRZ format type (TD1, TD2, TD3)
        has_pdf417: Whether document has PDF417 barcode
        has_qr_code: Whether document has QR code
        barcode_format: Type of barcode if present
        language: Primary language code (e.g., "en", "tr")
        description: Human-readable description
        fields: Dictionary of field definitions
    """
    type_id: str
    name: str
    standard: str
    has_mrz: bool = False
    mrz_format: Optional[str] = None
    has_pdf417: bool = False
    has_qr_code: bool = False
    barcode_format: Optional[str] = None
    language: Optional[str] = None
    description: Optional[str] = None
    fields: Dict[str, FieldDefinition] = field(default_factory=dict)
    
    def get_field(self, field_name: str) -> Optional[FieldDefinition]:
        """Get field definition by name."""
        return self.fields.get(field_name)
    
    def get_required_fields(self) -> List[str]:
        """Get list of required field names."""
        return [f.name for f in self.fields.values() if f.required]
    
    def get_optional_fields(self) -> List[str]:
        """Get list of optional field names."""
        return [f.name for f in self.fields.values() if not f.required]
    
    def get_fields_by_position(self, position_hint: str) -> List[FieldDefinition]:
        """Get fields matching a position hint."""
        return [f for f in self.fields.values() if f.position_hint == position_hint]
    
    def get_fields_with_label(self, label: str) -> List[FieldDefinition]:
        """Find fields that match a label text."""
        label_lower = label.lower()
        matching = []
        for field_def in self.fields.values():
            if field_def.label_text:
                for lt in field_def.label_text:
                    if label_lower in lt.lower():
                        matching.append(field_def)
                        break
        return matching
    
    def has_barcode(self) -> bool:
        """Check if document has any barcode."""
        return self.has_pdf417 or self.has_qr_code or bool(self.barcode_format)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "type_id": self.type_id,
            "name": self.name,
            "standard": self.standard,
            "has_mrz": self.has_mrz,
            "mrz_format": self.mrz_format,
            "has_pdf417": self.has_pdf417,
            "has_qr_code": self.has_qr_code,
            "barcode_format": self.barcode_format,
            "language": self.language,
            "description": self.description,
            "fields": {k: v.to_dict() for k, v in self.fields.items()},
        }


class DateFormatConfig:
    """Date format configuration from YAML."""
    
    def __init__(self, raw_config: Dict[str, Any]):
        self._formats: Dict[str, str] = {}
        for fmt_name, fmt_def in raw_config.items():
            if isinstance(fmt_def, dict):
                self._formats[fmt_name] = fmt_def.get("strptime", "")
            else:
                self._formats[fmt_name] = str(fmt_def)
    
    def get_strptime(self, format_name: str) -> Optional[str]:
        """Get strptime format string for a named format."""
        return self._formats.get(format_name)
    
    def parse_date(self, value: str, format_name: str) -> Optional[datetime]:
        """Parse a date string using named format."""
        strptime_fmt = self.get_strptime(format_name)
        if not strptime_fmt:
            return None
        try:
            return datetime.strptime(value, strptime_fmt)
        except ValueError:
            return None


class DocumentConfigLoader:
    """
    Loads and provides access to document type configurations.
    
    This class implements a singleton pattern to efficiently load and cache
    document type definitions from YAML configuration.
    
    Thread Safety:
        The singleton instance is created lazily. In multi-threaded scenarios,
        call get_instance() once during initialization.
    
    Example:
        loader = DocumentConfigLoader.get_instance()
        passport = loader.get_document_type("passport")
        
        # Validate field value
        is_valid, error = loader.validate_field("passport", "nationality", "USA")
    """
    
    _instance: Optional["DocumentConfigLoader"] = None
    _config: Dict[str, DocumentTypeDefinition]
    _config_path: Path
    _date_formats: Optional[DateFormatConfig]
    _validation_config: Dict[str, Any]
    _position_hints: Dict[str, Any]
    _config_version: str
    _schema_version: int
    
    @classmethod
    def get_instance(cls, config_path: Optional[Path] = None) -> "DocumentConfigLoader":
        """
        Get singleton instance of configuration loader.
        
        Args:
            config_path: Optional path to config file. Only used for first instantiation.
        
        Returns:
            Singleton DocumentConfigLoader instance
        """
        if cls._instance is None:
            cls._instance = cls(config_path)
        return cls._instance
    
    @classmethod
    def reset_instance(cls) -> None:
        """Reset singleton instance (mainly for testing)."""
        cls._instance = None
    
    def __init__(self, config_path: Optional[Path] = None):
        """
        Initialize configuration loader.
        
        Args:
            config_path: Path to YAML configuration file.
                        Defaults to document_types.yaml in same directory.
        """
        self._config = {}
        self._date_formats = None
        self._validation_config = {}
        self._position_hints = {}
        self._config_version = ""
        self._schema_version = 0
        
        if config_path:
            self._config_path = config_path
        else:
            self._config_path = Path(__file__).parent / "document_types.yaml"
        
        self._load_config()
    
    def _load_config(self) -> None:
        """Load configuration from YAML file."""
        if not self._config_path.exists():
            raise FileNotFoundError(
                f"Document config not found: {self._config_path}"
            )
        
        with open(self._config_path, encoding="utf-8") as f:
            raw_config = yaml.safe_load(f)
        
        if not raw_config:
            raise ValueError(f"Empty configuration file: {self._config_path}")
        
        # Load metadata
        self._config_version = raw_config.get("version", "unknown")
        self._schema_version = raw_config.get("schema_version", 0)
        
        # Load document types
        self._config = {}
        for type_id, type_def in raw_config.get("document_types", {}).items():
            self._config[type_id] = self._parse_document_type(type_id, type_def)
        
        # Load validation config
        self._validation_config = raw_config.get("validation", {})
        
        # Load date formats
        date_formats_raw = self._validation_config.get("date_formats", {})
        self._date_formats = DateFormatConfig(date_formats_raw)
        
        # Load position hints
        self._position_hints = raw_config.get("position_hints", {})
        
        logger.info(
            f"Loaded {len(self._config)} document types from {self._config_path}"
        )
    
    def _parse_document_type(
        self,
        type_id: str,
        type_def: Dict[str, Any]
    ) -> DocumentTypeDefinition:
        """Parse a document type definition from raw config."""
        fields = {}
        for field_name, field_def in type_def.get("fields", {}).items():
            fields[field_name] = FieldDefinition(
                name=field_name,
                required=field_def.get("required", False),
                pattern=field_def.get("pattern"),
                format=field_def.get("format"),
                max_length=field_def.get("max_length"),
                label_text=field_def.get("label_text"),
                position_hint=field_def.get("position_hint"),
                has_check_digit=field_def.get("has_check_digit", False),
                check_digit_algorithm=field_def.get("check_digit_algorithm"),
                description=field_def.get("description"),
                barcode_element=field_def.get("barcode_element"),
            )
        
        return DocumentTypeDefinition(
            type_id=type_id,
            name=type_def.get("name", type_id),
            standard=type_def.get("standard", ""),
            has_mrz=type_def.get("has_mrz", False),
            mrz_format=type_def.get("mrz_format"),
            has_pdf417=type_def.get("has_pdf417", False),
            has_qr_code=type_def.get("has_qr_code", False),
            barcode_format=type_def.get("barcode_format"),
            language=type_def.get("language"),
            description=type_def.get("description"),
            fields=fields,
        )
    
    def get_document_type(self, type_id: str) -> Optional[DocumentTypeDefinition]:
        """
        Get document type definition by ID.
        
        Args:
            type_id: Document type identifier (e.g., "passport", "turkish_id")
        
        Returns:
            DocumentTypeDefinition or None if not found
        """
        return self._config.get(type_id)
    
    def list_document_types(self) -> List[str]:
        """
        List all available document type IDs.
        
        Returns:
            List of document type identifiers
        """
        return list(self._config.keys())
    
    def get_document_types_by_standard(self, standard: str) -> List[str]:
        """
        Get document types following a specific standard.
        
        Args:
            standard: Standard name (e.g., "ICAO 9303", "AAMVA")
        
        Returns:
            List of matching document type IDs
        """
        return [
            type_id for type_id, doc_type in self._config.items()
            if doc_type.standard == standard
        ]
    
    def get_mrz_document_types(self) -> List[str]:
        """Get all document types that have MRZ."""
        return [
            type_id for type_id, doc_type in self._config.items()
            if doc_type.has_mrz
        ]
    
    def get_barcode_document_types(self) -> List[str]:
        """Get all document types that have barcodes."""
        return [
            type_id for type_id, doc_type in self._config.items()
            if doc_type.has_barcode()
        ]
    
    def validate_field(
        self,
        type_id: str,
        field_name: str,
        value: str
    ) -> Tuple[bool, Optional[str]]:
        """
        Validate a field value against document type definition.
        
        Args:
            type_id: Document type identifier
            field_name: Field name to validate
            value: Value to validate
        
        Returns:
            Tuple of (is_valid, error_message)
            error_message is None if valid
        """
        doc_type = self.get_document_type(type_id)
        if not doc_type:
            return False, f"Unknown document type: {type_id}"
        
        field_def = doc_type.fields.get(field_name)
        if not field_def:
            # Unknown field - allow by default
            return True, None
        
        # Check pattern
        if field_def.pattern:
            if not field_def.matches_pattern(value):
                return False, (
                    f"Field '{field_name}' does not match pattern "
                    f"'{field_def.pattern}'"
                )
        
        # Check max length
        if field_def.max_length and len(value) > field_def.max_length:
            return False, (
                f"Field '{field_name}' exceeds max length "
                f"({len(value)} > {field_def.max_length})"
            )
        
        # Check date format
        if field_def.format and self._date_formats:
            parsed = self._date_formats.parse_date(value, field_def.format)
            if parsed is None:
                return False, (
                    f"Field '{field_name}' does not match date format "
                    f"'{field_def.format}'"
                )
        
        return True, None
    
    def validate_document(
        self,
        type_id: str,
        fields: Dict[str, str]
    ) -> Tuple[bool, List[str]]:
        """
        Validate all fields for a document.
        
        Args:
            type_id: Document type identifier
            fields: Dictionary of field_name -> value
        
        Returns:
            Tuple of (all_valid, list_of_errors)
        """
        doc_type = self.get_document_type(type_id)
        if not doc_type:
            return False, [f"Unknown document type: {type_id}"]
        
        errors = []
        
        # Check required fields
        for req_field in doc_type.get_required_fields():
            if req_field not in fields or not fields[req_field]:
                errors.append(f"Missing required field: {req_field}")
        
        # Validate each provided field
        for field_name, value in fields.items():
            is_valid, error = self.validate_field(type_id, field_name, value)
            if not is_valid and error:
                errors.append(error)
        
        return len(errors) == 0, errors
    
    def get_required_fields(self, type_id: str) -> List[str]:
        """
        Get list of required field names for a document type.
        
        Args:
            type_id: Document type identifier
        
        Returns:
            List of required field names (empty if type not found)
        """
        doc_type = self.get_document_type(type_id)
        if not doc_type:
            return []
        return doc_type.get_required_fields()
    
    def get_field_labels(self, type_id: str, field_name: str) -> List[str]:
        """
        Get label text patterns for a field.
        
        Useful for label-based OCR extraction.
        
        Args:
            type_id: Document type identifier
            field_name: Field name
        
        Returns:
            List of label text patterns (empty if not found)
        """
        doc_type = self.get_document_type(type_id)
        if not doc_type:
            return []
        
        field_def = doc_type.fields.get(field_name)
        if not field_def or not field_def.label_text:
            return []
        
        return field_def.label_text
    
    def find_field_by_label(
        self,
        type_id: str,
        label: str
    ) -> Optional[FieldDefinition]:
        """
        Find a field definition that matches a label.
        
        Args:
            type_id: Document type identifier
            label: Label text from OCR
        
        Returns:
            Matching FieldDefinition or None
        """
        doc_type = self.get_document_type(type_id)
        if not doc_type:
            return None
        
        matches = doc_type.get_fields_with_label(label)
        return matches[0] if matches else None
    
    def get_position_hint_config(
        self,
        hint_name: str
    ) -> Optional[Dict[str, Any]]:
        """
        Get position hint configuration.
        
        Args:
            hint_name: Position hint name (e.g., "top_third")
        
        Returns:
            Position hint config with y_range/x_range or None
        """
        return self._position_hints.get(hint_name)
    
    def get_date_format(self, format_name: str) -> Optional[str]:
        """
        Get strptime format string for a named date format.
        
        Args:
            format_name: Format name (e.g., "YYMMDD", "DD.MM.YYYY")
        
        Returns:
            strptime format string or None
        """
        if self._date_formats:
            return self._date_formats.get_strptime(format_name)
        return None
    
    def reload_config(self) -> None:
        """
        Hot-reload configuration from disk.
        
        Call this to reload configuration without restarting the service.
        Useful for dynamic configuration updates.
        """
        logger.info(f"Reloading document config from {self._config_path}")
        self._load_config()
    
    @property
    def config_version(self) -> str:
        """Get configuration version string."""
        return self._config_version
    
    @property
    def schema_version(self) -> int:
        """Get configuration schema version."""
        return self._schema_version
    
    def get_stats(self) -> Dict[str, Any]:
        """
        Get statistics about loaded configuration.
        
        Returns:
            Dictionary with config statistics
        """
        total_fields = sum(
            len(doc.fields) for doc in self._config.values()
        )
        mrz_types = len(self.get_mrz_document_types())
        barcode_types = len(self.get_barcode_document_types())
        
        return {
            "config_version": self._config_version,
            "schema_version": self._schema_version,
            "document_types_count": len(self._config),
            "total_fields_count": total_fields,
            "mrz_document_types": mrz_types,
            "barcode_document_types": barcode_types,
            "config_path": str(self._config_path),
        }


def get_document_type(type_id: str) -> Optional[DocumentTypeDefinition]:
    """
    Convenience function to get document type.
    
    Uses singleton loader instance.
    
    Args:
        type_id: Document type identifier
    
    Returns:
        DocumentTypeDefinition or None
    """
    return DocumentConfigLoader.get_instance().get_document_type(type_id)


def validate_field(
    type_id: str,
    field_name: str,
    value: str
) -> Tuple[bool, Optional[str]]:
    """
    Convenience function to validate a field.
    
    Uses singleton loader instance.
    
    Args:
        type_id: Document type identifier
        field_name: Field name
        value: Value to validate
    
    Returns:
        Tuple of (is_valid, error_message)
    """
    return DocumentConfigLoader.get_instance().validate_field(
        type_id, field_name, value
    )
