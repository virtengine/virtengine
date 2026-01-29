"""
Validate document type configuration schema.

This module provides validation for document_types.yaml configuration
to catch errors before runtime.

Task Reference: VE-3047 - Generalize Document Type Configuration

Example Usage:
    from ml.ocr_extraction.config_validator import validate_config
    from pathlib import Path
    
    config_path = Path("document_types.yaml")
    valid, errors = validate_config(config_path)
    
    if not valid:
        for error in errors:
            print(f"Config Error: {error}")
"""

import re
import logging
from pathlib import Path
from typing import List, Tuple, Dict, Any, Set, Optional

import yaml

logger = logging.getLogger(__name__)


# Required top-level keys
REQUIRED_TOP_LEVEL_KEYS = {"document_types"}

# Optional top-level keys
OPTIONAL_TOP_LEVEL_KEYS = {
    "version",
    "schema_version", 
    "validation",
    "position_hints",
}

# Required keys in document type definition
REQUIRED_DOC_TYPE_KEYS = {"name", "fields"}

# Optional keys in document type definition
OPTIONAL_DOC_TYPE_KEYS = {
    "standard",
    "has_mrz",
    "mrz_format",
    "has_pdf417",
    "has_qr_code",
    "barcode_format",
    "language",
    "description",
}

# Valid MRZ formats
VALID_MRZ_FORMATS = {"TD1", "TD2", "TD3"}

# Valid position hints
VALID_POSITION_HINTS = {
    "top",
    "top_third",
    "upper_third",
    "middle",
    "lower_third",
    "bottom_third",
    "bottom",
    "left_half",
    "right_half",
}

# Required keys in field definition
REQUIRED_FIELD_KEYS = {"required"}

# Optional keys in field definition
OPTIONAL_FIELD_KEYS = {
    "pattern",
    "format",
    "max_length",
    "label_text",
    "position_hint",
    "has_check_digit",
    "check_digit_algorithm",
    "description",
    "barcode_element",
}


class ConfigValidationError:
    """Represents a configuration validation error."""
    
    def __init__(
        self,
        path: str,
        message: str,
        severity: str = "error"
    ):
        """
        Initialize validation error.
        
        Args:
            path: Path in config (e.g., "passport.fields.surname")
            message: Error message
            severity: "error" or "warning"
        """
        self.path = path
        self.message = message
        self.severity = severity
    
    def __str__(self) -> str:
        return f"[{self.severity.upper()}] {self.path}: {self.message}"


class ConfigValidator:
    """
    Validates document type configuration.
    
    Performs structural and semantic validation of document_types.yaml.
    """
    
    def __init__(self):
        self.errors: List[ConfigValidationError] = []
        self.warnings: List[ConfigValidationError] = []
    
    def validate(self, config_path: Path) -> Tuple[bool, List[str]]:
        """
        Validate configuration file.
        
        Args:
            config_path: Path to YAML configuration file
        
        Returns:
            Tuple of (is_valid, list_of_error_messages)
        """
        self.errors = []
        self.warnings = []
        
        # Check file exists
        if not config_path.exists():
            self.errors.append(ConfigValidationError(
                str(config_path),
                "Configuration file not found",
            ))
            return False, [str(e) for e in self.errors]
        
        # Load and parse YAML
        try:
            with open(config_path, encoding="utf-8") as f:
                config = yaml.safe_load(f)
        except yaml.YAMLError as e:
            self.errors.append(ConfigValidationError(
                str(config_path),
                f"YAML parse error: {e}",
            ))
            return False, [str(e) for e in self.errors]
        
        if config is None:
            self.errors.append(ConfigValidationError(
                str(config_path),
                "Empty configuration file",
            ))
            return False, [str(e) for e in self.errors]
        
        # Validate structure
        self._validate_top_level(config)
        
        # Validate document types
        for type_id, type_def in config.get("document_types", {}).items():
            self._validate_document_type(type_id, type_def)
        
        # Validate validation section
        if "validation" in config:
            self._validate_validation_section(config["validation"])
        
        # Validate position hints
        if "position_hints" in config:
            self._validate_position_hints(config["position_hints"])
        
        # Compile results
        is_valid = len(self.errors) == 0
        all_messages = [str(e) for e in self.errors] + [str(w) for w in self.warnings]
        
        return is_valid, [str(e) for e in self.errors]
    
    def _validate_top_level(self, config: Dict[str, Any]) -> None:
        """Validate top-level configuration structure."""
        # Check required keys
        for key in REQUIRED_TOP_LEVEL_KEYS:
            if key not in config:
                self.errors.append(ConfigValidationError(
                    "root",
                    f"Missing required key: '{key}'",
                ))
        
        # Check for unknown keys
        all_known_keys = REQUIRED_TOP_LEVEL_KEYS | OPTIONAL_TOP_LEVEL_KEYS
        for key in config.keys():
            if key not in all_known_keys:
                self.warnings.append(ConfigValidationError(
                    "root",
                    f"Unknown key: '{key}'",
                    severity="warning",
                ))
        
        # Validate version format
        if "version" in config:
            version = config["version"]
            if not isinstance(version, str):
                self.errors.append(ConfigValidationError(
                    "version",
                    "Version must be a string",
                ))
        
        # Validate schema version
        if "schema_version" in config:
            schema_version = config["schema_version"]
            if not isinstance(schema_version, int) or schema_version < 1:
                self.errors.append(ConfigValidationError(
                    "schema_version",
                    "Schema version must be a positive integer",
                ))
    
    def _validate_document_type(
        self,
        type_id: str,
        type_def: Dict[str, Any]
    ) -> None:
        """Validate a single document type definition."""
        path = f"document_types.{type_id}"
        
        if not isinstance(type_def, dict):
            self.errors.append(ConfigValidationError(
                path,
                "Document type must be a dictionary",
            ))
            return
        
        # Check required keys
        for key in REQUIRED_DOC_TYPE_KEYS:
            if key not in type_def:
                self.errors.append(ConfigValidationError(
                    path,
                    f"Missing required key: '{key}'",
                ))
        
        # Validate name
        if "name" in type_def and not isinstance(type_def["name"], str):
            self.errors.append(ConfigValidationError(
                f"{path}.name",
                "Name must be a string",
            ))
        
        # Validate MRZ settings
        if type_def.get("has_mrz", False):
            mrz_format = type_def.get("mrz_format")
            if mrz_format and mrz_format not in VALID_MRZ_FORMATS:
                self.errors.append(ConfigValidationError(
                    f"{path}.mrz_format",
                    f"Invalid MRZ format: '{mrz_format}'. "
                    f"Must be one of: {VALID_MRZ_FORMATS}",
                ))
        
        # Validate fields
        fields = type_def.get("fields", {})
        if not isinstance(fields, dict):
            self.errors.append(ConfigValidationError(
                f"{path}.fields",
                "Fields must be a dictionary",
            ))
        else:
            for field_name, field_def in fields.items():
                self._validate_field(type_id, field_name, field_def)
    
    def _validate_field(
        self,
        type_id: str,
        field_name: str,
        field_def: Dict[str, Any]
    ) -> None:
        """Validate a single field definition."""
        path = f"document_types.{type_id}.fields.{field_name}"
        
        if not isinstance(field_def, dict):
            self.errors.append(ConfigValidationError(
                path,
                "Field definition must be a dictionary",
            ))
            return
        
        # Check required keys
        for key in REQUIRED_FIELD_KEYS:
            if key not in field_def:
                self.errors.append(ConfigValidationError(
                    path,
                    f"Missing required key: '{key}'",
                ))
        
        # Validate 'required' is boolean
        if "required" in field_def and not isinstance(field_def["required"], bool):
            self.errors.append(ConfigValidationError(
                f"{path}.required",
                "Required must be a boolean",
            ))
        
        # Validate pattern is valid regex
        if "pattern" in field_def:
            pattern = field_def["pattern"]
            if not isinstance(pattern, str):
                self.errors.append(ConfigValidationError(
                    f"{path}.pattern",
                    "Pattern must be a string",
                ))
            else:
                try:
                    re.compile(pattern)
                except re.error as e:
                    self.errors.append(ConfigValidationError(
                        f"{path}.pattern",
                        f"Invalid regex pattern: {e}",
                    ))
        
        # Validate max_length is positive integer
        if "max_length" in field_def:
            max_length = field_def["max_length"]
            if not isinstance(max_length, int) or max_length <= 0:
                self.errors.append(ConfigValidationError(
                    f"{path}.max_length",
                    "Max length must be a positive integer",
                ))
        
        # Validate label_text is list of strings
        if "label_text" in field_def:
            label_text = field_def["label_text"]
            if not isinstance(label_text, list):
                self.errors.append(ConfigValidationError(
                    f"{path}.label_text",
                    "Label text must be a list",
                ))
            elif not all(isinstance(lt, str) for lt in label_text):
                self.errors.append(ConfigValidationError(
                    f"{path}.label_text",
                    "All label text entries must be strings",
                ))
        
        # Validate position_hint
        if "position_hint" in field_def:
            hint = field_def["position_hint"]
            if hint not in VALID_POSITION_HINTS:
                self.warnings.append(ConfigValidationError(
                    f"{path}.position_hint",
                    f"Unknown position hint: '{hint}'",
                    severity="warning",
                ))
        
        # Validate has_check_digit is boolean
        if "has_check_digit" in field_def:
            if not isinstance(field_def["has_check_digit"], bool):
                self.errors.append(ConfigValidationError(
                    f"{path}.has_check_digit",
                    "has_check_digit must be a boolean",
                ))
    
    def _validate_validation_section(
        self,
        validation: Dict[str, Any]
    ) -> None:
        """Validate the validation section."""
        path = "validation"
        
        if not isinstance(validation, dict):
            self.errors.append(ConfigValidationError(
                path,
                "Validation section must be a dictionary",
            ))
            return
        
        # Validate date_formats
        if "date_formats" in validation:
            date_formats = validation["date_formats"]
            if not isinstance(date_formats, dict):
                self.errors.append(ConfigValidationError(
                    f"{path}.date_formats",
                    "Date formats must be a dictionary",
                ))
            else:
                for fmt_name, fmt_def in date_formats.items():
                    if isinstance(fmt_def, dict):
                        if "strptime" not in fmt_def:
                            self.warnings.append(ConfigValidationError(
                                f"{path}.date_formats.{fmt_name}",
                                "Missing 'strptime' format string",
                                severity="warning",
                            ))
        
        # Validate check_digit_algorithms
        if "check_digit_algorithms" in validation:
            algorithms = validation["check_digit_algorithms"]
            if not isinstance(algorithms, dict):
                self.errors.append(ConfigValidationError(
                    f"{path}.check_digit_algorithms",
                    "Check digit algorithms must be a dictionary",
                ))
    
    def _validate_position_hints(
        self,
        hints: Dict[str, Any]
    ) -> None:
        """Validate position hints section."""
        path = "position_hints"
        
        if not isinstance(hints, dict):
            self.errors.append(ConfigValidationError(
                path,
                "Position hints must be a dictionary",
            ))
            return
        
        for hint_name, hint_def in hints.items():
            if not isinstance(hint_def, dict):
                continue
            
            # Check for y_range or x_range
            if "y_range" in hint_def:
                y_range = hint_def["y_range"]
                if not isinstance(y_range, list) or len(y_range) != 2:
                    self.errors.append(ConfigValidationError(
                        f"{path}.{hint_name}.y_range",
                        "y_range must be a list of two numbers",
                    ))
                elif not all(isinstance(v, (int, float)) for v in y_range):
                    self.errors.append(ConfigValidationError(
                        f"{path}.{hint_name}.y_range",
                        "y_range values must be numbers",
                    ))
            
            if "x_range" in hint_def:
                x_range = hint_def["x_range"]
                if not isinstance(x_range, list) or len(x_range) != 2:
                    self.errors.append(ConfigValidationError(
                        f"{path}.{hint_name}.x_range",
                        "x_range must be a list of two numbers",
                    ))
                elif not all(isinstance(v, (int, float)) for v in x_range):
                    self.errors.append(ConfigValidationError(
                        f"{path}.{hint_name}.x_range",
                        "x_range values must be numbers",
                    ))


def validate_config(config_path: Path) -> Tuple[bool, List[str]]:
    """
    Validate document type configuration file.
    
    This is the main entry point for configuration validation.
    
    Args:
        config_path: Path to YAML configuration file
    
    Returns:
        Tuple of (is_valid, list_of_error_messages)
    
    Example:
        valid, errors = validate_config(Path("document_types.yaml"))
        if not valid:
            print("Configuration errors:")
            for error in errors:
                print(f"  - {error}")
    """
    validator = ConfigValidator()
    return validator.validate(config_path)


def validate_config_string(yaml_content: str) -> Tuple[bool, List[str]]:
    """
    Validate document type configuration from string.
    
    Useful for testing or validating config before writing to file.
    
    Args:
        yaml_content: YAML configuration as string
    
    Returns:
        Tuple of (is_valid, list_of_error_messages)
    """
    import tempfile
    
    with tempfile.NamedTemporaryFile(
        mode="w",
        suffix=".yaml",
        delete=False,
        encoding="utf-8"
    ) as f:
        f.write(yaml_content)
        temp_path = Path(f.name)
    
    try:
        return validate_config(temp_path)
    finally:
        temp_path.unlink()


def check_document_type_exists(
    config_path: Path,
    type_id: str
) -> Tuple[bool, Optional[str]]:
    """
    Check if a specific document type exists in configuration.
    
    Args:
        config_path: Path to YAML configuration file
        type_id: Document type identifier to check
    
    Returns:
        Tuple of (exists, error_message_if_not)
    """
    try:
        with open(config_path, encoding="utf-8") as f:
            config = yaml.safe_load(f)
    except Exception as e:
        return False, f"Failed to load config: {e}"
    
    if not config:
        return False, "Empty configuration"
    
    doc_types = config.get("document_types", {})
    if type_id in doc_types:
        return True, None
    else:
        return False, f"Document type '{type_id}' not found"


def get_all_document_type_ids(config_path: Path) -> List[str]:
    """
    Get all document type IDs from configuration.
    
    Args:
        config_path: Path to YAML configuration file
    
    Returns:
        List of document type identifiers
    """
    try:
        with open(config_path, encoding="utf-8") as f:
            config = yaml.safe_load(f)
    except Exception:
        return []
    
    if not config:
        return []
    
    return list(config.get("document_types", {}).keys())
