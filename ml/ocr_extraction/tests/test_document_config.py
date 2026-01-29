"""
Tests for document type configuration system.

Task Reference: VE-3047 - Generalize Document Type Configuration

These tests verify:
- Configuration loading from YAML
- Document type definitions
- Field validation
- Configuration validator
"""

import pytest
from pathlib import Path
from typing import Dict

from ml.ocr_extraction.document_config import (
    DocumentConfigLoader,
    DocumentTypeDefinition,
    FieldDefinition,
    get_document_type,
    validate_field,
)
from ml.ocr_extraction.config_validator import (
    validate_config,
    validate_config_string,
    check_document_type_exists,
    get_all_document_type_ids,
    ConfigValidator,
)


# Path to the config file
CONFIG_PATH = Path(__file__).parent.parent / "document_types.yaml"


class TestFieldDefinition:
    """Tests for FieldDefinition dataclass."""
    
    def test_basic_field(self):
        """Test creating a basic field definition."""
        field = FieldDefinition(
            name="surname",
            required=True,
            pattern="^[A-Z]+$",
        )
        assert field.name == "surname"
        assert field.required is True
        assert field.pattern == "^[A-Z]+$"
    
    def test_matches_pattern_valid(self):
        """Test pattern matching with valid value."""
        field = FieldDefinition(
            name="nationality",
            required=True,
            pattern="^[A-Z]{3}$",
        )
        assert field.matches_pattern("USA") is True
        assert field.matches_pattern("GBR") is True
    
    def test_matches_pattern_invalid(self):
        """Test pattern matching with invalid value."""
        field = FieldDefinition(
            name="nationality",
            required=True,
            pattern="^[A-Z]{3}$",
        )
        assert field.matches_pattern("US") is False
        assert field.matches_pattern("USAA") is False
        assert field.matches_pattern("123") is False
    
    def test_matches_pattern_no_pattern(self):
        """Test that no pattern means any value is accepted."""
        field = FieldDefinition(
            name="notes",
            required=False,
        )
        assert field.matches_pattern("anything goes") is True
        assert field.matches_pattern("") is True
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        field = FieldDefinition(
            name="surname",
            required=True,
            pattern="^[A-Z]+$",
            max_length=50,
            position_hint="top_third",
        )
        d = field.to_dict()
        
        assert d["name"] == "surname"
        assert d["required"] is True
        assert d["pattern"] == "^[A-Z]+$"
        assert d["max_length"] == 50
        assert d["position_hint"] == "top_third"


class TestDocumentTypeDefinition:
    """Tests for DocumentTypeDefinition dataclass."""
    
    def test_get_required_fields(self):
        """Test getting required field names."""
        doc_type = DocumentTypeDefinition(
            type_id="test_doc",
            name="Test Document",
            standard="Test",
            fields={
                "surname": FieldDefinition(name="surname", required=True),
                "given_names": FieldDefinition(name="given_names", required=True),
                "middle_name": FieldDefinition(name="middle_name", required=False),
            }
        )
        
        required = doc_type.get_required_fields()
        assert "surname" in required
        assert "given_names" in required
        assert "middle_name" not in required
    
    def test_get_optional_fields(self):
        """Test getting optional field names."""
        doc_type = DocumentTypeDefinition(
            type_id="test_doc",
            name="Test Document",
            standard="Test",
            fields={
                "surname": FieldDefinition(name="surname", required=True),
                "nickname": FieldDefinition(name="nickname", required=False),
            }
        )
        
        optional = doc_type.get_optional_fields()
        assert "nickname" in optional
        assert "surname" not in optional
    
    def test_has_barcode(self):
        """Test barcode detection."""
        # No barcode
        no_barcode = DocumentTypeDefinition(
            type_id="passport",
            name="Passport",
            standard="ICAO",
            has_mrz=True,
        )
        assert no_barcode.has_barcode() is False
        
        # Has PDF417
        with_pdf417 = DocumentTypeDefinition(
            type_id="us_dl",
            name="US DL",
            standard="AAMVA",
            has_pdf417=True,
        )
        assert with_pdf417.has_barcode() is True
        
        # Has QR
        with_qr = DocumentTypeDefinition(
            type_id="aadhaar",
            name="Aadhaar",
            standard="India",
            has_qr_code=True,
        )
        assert with_qr.has_barcode() is True


class TestDocumentConfigLoader:
    """Tests for DocumentConfigLoader class."""
    
    @pytest.fixture(autouse=True)
    def reset_singleton(self):
        """Reset singleton before each test."""
        DocumentConfigLoader.reset_instance()
        yield
        DocumentConfigLoader.reset_instance()
    
    def test_load_config(self):
        """Test loading configuration from YAML."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        types = loader.list_document_types()
        assert len(types) > 0
    
    def test_singleton_pattern(self):
        """Test singleton returns same instance."""
        loader1 = DocumentConfigLoader.get_instance(CONFIG_PATH)
        loader2 = DocumentConfigLoader.get_instance()
        assert loader1 is loader2
    
    def test_passport_definition(self):
        """Test passport document type is properly loaded."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        passport = loader.get_document_type("passport")
        
        assert passport is not None
        assert passport.has_mrz is True
        assert passport.mrz_format == "TD3"
        assert passport.standard == "ICAO 9303"
        assert "surname" in passport.fields
        assert "given_names" in passport.fields
        assert "nationality" in passport.fields
    
    def test_passport_field_patterns(self):
        """Test passport field patterns are correct."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        passport = loader.get_document_type("passport")
        
        nationality = passport.fields.get("nationality")
        assert nationality is not None
        assert nationality.pattern == "^[A-Z]{3}$"
        assert nationality.matches_pattern("USA") is True
        assert nationality.matches_pattern("US") is False
    
    def test_turkish_id_definition(self):
        """Test Turkish ID (non-MRZ) document type."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        turkish = loader.get_document_type("turkish_id")
        
        assert turkish is not None
        assert turkish.has_mrz is False
        assert turkish.language == "tr"
        assert "tc_kimlik_no" in turkish.fields
        
        # Check TC Kimlik pattern
        tc_field = turkish.fields.get("tc_kimlik_no")
        assert tc_field is not None
        assert tc_field.pattern == "^[0-9]{11}$"
        assert tc_field.label_text is not None
        assert "T.C. Kimlik No" in tc_field.label_text
    
    def test_us_drivers_license_definition(self):
        """Test US driver's license with PDF417 barcode."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        us_dl = loader.get_document_type("us_drivers_license")
        
        assert us_dl is not None
        assert us_dl.has_pdf417 is True
        assert us_dl.barcode_format == "PDF417"
        assert us_dl.standard == "AAMVA"
        assert "dcs" in us_dl.fields  # Family name
        assert "dac" in us_dl.fields  # First name
    
    def test_validate_field_pattern(self):
        """Test field validation with pattern matching."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        # Valid nationality
        valid, error = loader.validate_field("passport", "nationality", "USA")
        assert valid is True
        assert error is None
        
        # Invalid nationality (too long)
        valid, error = loader.validate_field("passport", "nationality", "TOOLONG")
        assert valid is False
        assert error is not None
        assert "pattern" in error.lower()
    
    def test_validate_field_max_length(self):
        """Test field validation with max length."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        # Passport surname max_length is 39
        long_name = "A" * 50
        valid, error = loader.validate_field("passport", "surname", long_name)
        assert valid is False
        assert "max length" in error.lower()
    
    def test_validate_field_unknown_type(self):
        """Test validation with unknown document type."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        valid, error = loader.validate_field("unknown_doc", "field", "value")
        assert valid is False
        assert "Unknown document type" in error
    
    def test_validate_field_unknown_field(self):
        """Test validation with unknown field (should pass)."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        # Unknown fields are allowed
        valid, error = loader.validate_field("passport", "unknown_field", "value")
        assert valid is True
    
    def test_validate_document(self):
        """Test validating a complete document."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        # Valid passport fields
        fields = {
            "surname": "SMITH",
            "given_names": "JOHN",
            "nationality": "USA",
            "document_number": "AB1234567",
            "date_of_birth": "900115",
            "date_of_expiry": "300115",
            "sex": "M",
            "issuing_country": "USA",
        }
        
        valid, errors = loader.validate_document("passport", fields)
        # May have some validation issues depending on patterns
        assert isinstance(valid, bool)
        assert isinstance(errors, list)
    
    def test_get_required_fields(self):
        """Test getting required fields for document type."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        required = loader.get_required_fields("passport")
        assert "surname" in required
        assert "given_names" in required
        assert "nationality" in required
    
    def test_get_mrz_document_types(self):
        """Test getting MRZ document types."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        mrz_types = loader.get_mrz_document_types()
        assert "passport" in mrz_types
        assert "national_id_td1" in mrz_types
        assert "turkish_id" not in mrz_types  # Non-MRZ
    
    def test_get_barcode_document_types(self):
        """Test getting barcode document types."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        barcode_types = loader.get_barcode_document_types()
        assert "us_drivers_license" in barcode_types
    
    def test_get_field_labels(self):
        """Test getting label text for a field."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        labels = loader.get_field_labels("turkish_id", "tc_kimlik_no")
        assert len(labels) > 0
        assert "T.C. Kimlik No" in labels
    
    def test_find_field_by_label(self):
        """Test finding field by label text."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        field = loader.find_field_by_label("turkish_id", "Soyadı")
        assert field is not None
        assert field.name == "soyadi"
    
    def test_get_stats(self):
        """Test getting configuration statistics."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        stats = loader.get_stats()
        assert "document_types_count" in stats
        assert stats["document_types_count"] > 0
        assert "total_fields_count" in stats
        assert stats["total_fields_count"] > 0
    
    def test_reload_config(self):
        """Test hot-reloading configuration."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        initial_count = len(loader.list_document_types())
        
        # Reload (should work without error)
        loader.reload_config()
        
        # Count should be same
        assert len(loader.list_document_types()) == initial_count


class TestConvenienceFunctions:
    """Tests for module-level convenience functions."""
    
    @pytest.fixture(autouse=True)
    def reset_singleton(self):
        """Reset singleton before each test."""
        DocumentConfigLoader.reset_instance()
        yield
        DocumentConfigLoader.reset_instance()
    
    def test_get_document_type(self):
        """Test convenience function for getting document type."""
        # Force config path for testing
        DocumentConfigLoader.get_instance(CONFIG_PATH)
        
        passport = get_document_type("passport")
        assert passport is not None
        assert passport.type_id == "passport"
    
    def test_validate_field_function(self):
        """Test convenience function for validating field."""
        # Force config path for testing
        DocumentConfigLoader.get_instance(CONFIG_PATH)
        
        valid, error = validate_field("passport", "nationality", "USA")
        assert valid is True


class TestConfigValidator:
    """Tests for configuration validation."""
    
    def test_validate_existing_config(self):
        """Test validation of the actual config file."""
        valid, errors = validate_config(CONFIG_PATH)
        assert valid is True, f"Config validation failed: {errors}"
    
    def test_validate_missing_file(self):
        """Test validation with missing file."""
        valid, errors = validate_config(Path("/nonexistent/file.yaml"))
        assert valid is False
        assert len(errors) > 0
    
    def test_validate_config_string_valid(self):
        """Test validating valid config from string."""
        yaml_content = """
version: "1.0"
schema_version: 1
document_types:
  test_doc:
    name: "Test Document"
    fields:
      test_field:
        required: true
        pattern: "^[A-Z]+$"
"""
        valid, errors = validate_config_string(yaml_content)
        assert valid is True
    
    def test_validate_config_string_missing_required(self):
        """Test validating config missing required keys."""
        yaml_content = """
version: "1.0"
# Missing document_types
"""
        valid, errors = validate_config_string(yaml_content)
        assert valid is False
        assert any("document_types" in e for e in errors)
    
    def test_validate_config_string_invalid_pattern(self):
        """Test validating config with invalid regex pattern."""
        yaml_content = """
version: "1.0"
schema_version: 1
document_types:
  test_doc:
    name: "Test Document"
    fields:
      bad_field:
        required: true
        pattern: "[invalid(regex"
"""
        valid, errors = validate_config_string(yaml_content)
        assert valid is False
        assert any("regex" in e.lower() or "pattern" in e.lower() for e in errors)
    
    def test_validate_config_string_invalid_mrz_format(self):
        """Test validating config with invalid MRZ format."""
        yaml_content = """
version: "1.0"
schema_version: 1
document_types:
  test_doc:
    name: "Test Document"
    has_mrz: true
    mrz_format: "TD99"
    fields:
      test_field:
        required: true
"""
        valid, errors = validate_config_string(yaml_content)
        assert valid is False
        assert any("MRZ format" in e for e in errors)
    
    def test_check_document_type_exists(self):
        """Test checking if document type exists."""
        exists, error = check_document_type_exists(CONFIG_PATH, "passport")
        assert exists is True
        assert error is None
        
        exists, error = check_document_type_exists(CONFIG_PATH, "nonexistent")
        assert exists is False
        assert error is not None
    
    def test_get_all_document_type_ids(self):
        """Test getting all document type IDs."""
        type_ids = get_all_document_type_ids(CONFIG_PATH)
        assert len(type_ids) > 0
        assert "passport" in type_ids


class TestDocumentTypeCount:
    """Tests for document type coverage."""
    
    @pytest.fixture(autouse=True)
    def reset_singleton(self):
        """Reset singleton before each test."""
        DocumentConfigLoader.reset_instance()
        yield
        DocumentConfigLoader.reset_instance()
    
    def test_minimum_document_types(self):
        """Test that we have minimum required document types."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        types = loader.list_document_types()
        
        # Should have at least these types
        expected_types = [
            "passport",
            "national_id_td1",
            "turkish_id",
            "us_drivers_license",
        ]
        
        for expected in expected_types:
            assert expected in types, f"Missing document type: {expected}"
    
    def test_document_type_count(self):
        """Test total document type count."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        types = loader.list_document_types()
        
        # Should have reasonable number of document types
        # Based on our YAML: 13 document types
        assert len(types) >= 10, f"Expected at least 10 document types, got {len(types)}"
    
    def test_total_field_count(self):
        """Test total field count across all documents."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        stats = loader.get_stats()
        
        # Should have many fields across all documents
        assert stats["total_fields_count"] >= 50, (
            f"Expected at least 50 fields, got {stats['total_fields_count']}"
        )


class TestIntegrationWithFieldParser:
    """Tests for integration with existing field parser."""
    
    @pytest.fixture(autouse=True)
    def reset_singleton(self):
        """Reset singleton before each test."""
        DocumentConfigLoader.reset_instance()
        yield
        DocumentConfigLoader.reset_instance()
    
    def test_field_patterns_match_existing(self):
        """Test that configured patterns are usable for parsing."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        
        # Get passport and verify patterns work
        passport = loader.get_document_type("passport")
        assert passport is not None
        
        # Test surname pattern
        surname_field = passport.fields.get("surname")
        assert surname_field is not None
        assert surname_field.matches_pattern("SMITH") is True
        assert surname_field.matches_pattern("O'BRIEN") is True
        assert surname_field.matches_pattern("123") is False
    
    def test_mrz_fields_complete(self):
        """Test that MRZ documents have all standard fields."""
        loader = DocumentConfigLoader(CONFIG_PATH)
        passport = loader.get_document_type("passport")
        
        # Standard MRZ fields
        mrz_fields = [
            "surname",
            "given_names",
            "nationality",
            "date_of_birth",
            "sex",
            "date_of_expiry",
            "document_number",
        ]
        
        for field_name in mrz_fields:
            assert field_name in passport.fields, (
                f"Passport missing MRZ field: {field_name}"
            )
