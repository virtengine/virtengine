"""
Tests for barcode scanning configuration.
"""

import pytest

from ml.barcode_scanning.config import (
    BarcodeScanningConfig,
    PDF417Config,
    MRZConfig,
    CrossValidationConfig,
    ScoringConfig,
    BarcodeType,
    MRZFormat,
    DocumentCategory,
)


class TestBarcodeType:
    """Tests for BarcodeType enum."""
    
    def test_enum_values(self):
        """Test enum value strings."""
        assert BarcodeType.PDF417.value == "pdf417"
        assert BarcodeType.QR_CODE.value == "qr_code"
        assert BarcodeType.DATA_MATRIX.value == "data_matrix"
        assert BarcodeType.AZTEC.value == "aztec"
        assert BarcodeType.UNKNOWN.value == "unknown"


class TestMRZFormat:
    """Tests for MRZFormat enum."""
    
    def test_enum_values(self):
        """Test enum value strings."""
        assert MRZFormat.TD1.value == "td1"
        assert MRZFormat.TD2.value == "td2"
        assert MRZFormat.TD3.value == "td3"
        assert MRZFormat.UNKNOWN.value == "unknown"


class TestDocumentCategory:
    """Tests for DocumentCategory enum."""
    
    def test_enum_values(self):
        """Test enum value strings."""
        assert DocumentCategory.DRIVERS_LICENSE.value == "drivers_license"
        assert DocumentCategory.PASSPORT.value == "passport"
        assert DocumentCategory.ID_CARD.value == "id_card"
        assert DocumentCategory.VISA.value == "visa"
        assert DocumentCategory.UNKNOWN.value == "unknown"


class TestPDF417Config:
    """Tests for PDF417Config dataclass."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = PDF417Config()
        
        assert config.enable_aamva is True
        assert config.min_aamva_version == 1
        assert config.max_aamva_version == 10
        assert config.detection_confidence_threshold == 0.5
        assert config.auto_rotate is True
    
    def test_critical_fields(self):
        """Test critical fields set."""
        config = PDF417Config()
        
        assert "DAQ" in config.critical_fields  # Document number
        assert "DCS" in config.critical_fields  # Last name
        assert "DBB" in config.critical_fields  # Date of birth
    
    def test_optional_fields(self):
        """Test optional fields set."""
        config = PDF417Config()
        
        assert "DAD" in config.optional_fields  # Middle name
        assert "DAU" in config.optional_fields  # Height


class TestMRZConfig:
    """Tests for MRZConfig dataclass."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = MRZConfig()
        
        assert config.validate_check_digits is True
        assert config.strict_validation is False
        assert config.enable_ocr_correction is True
        assert config.min_line_length == 28
    
    def test_supported_formats(self):
        """Test supported formats set."""
        config = MRZConfig()
        
        assert MRZFormat.TD1 in config.supported_formats
        assert MRZFormat.TD2 in config.supported_formats
        assert MRZFormat.TD3 in config.supported_formats
    
    def test_valid_characters(self):
        """Test valid character set."""
        config = MRZConfig()
        
        assert "A" in config.valid_characters
        assert "Z" in config.valid_characters
        assert "0" in config.valid_characters
        assert "9" in config.valid_characters
        assert "<" in config.valid_characters


class TestCrossValidationConfig:
    """Tests for CrossValidationConfig dataclass."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = CrossValidationConfig()
        
        assert config.string_similarity_threshold == 0.8
        assert config.exact_match_bonus == 0.1
        assert config.enable_fuzzy_matching is True
        assert config.max_edit_distance == 2
        assert config.case_sensitive is False
    
    def test_field_weights(self):
        """Test field weights dictionary."""
        config = CrossValidationConfig()
        
        assert "full_name" in config.field_weights
        assert "date_of_birth" in config.field_weights
        assert config.field_weights["date_of_birth"] == 1.0  # High weight
    
    def test_required_fields(self):
        """Test required fields set."""
        config = CrossValidationConfig()
        
        assert "date_of_birth" in config.required_fields


class TestScoringConfig:
    """Tests for ScoringConfig dataclass."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = ScoringConfig()
        
        assert config.barcode_read_score == 0.15
        assert config.check_digit_bonus == 0.05
        assert config.cross_validation_weight == 0.20
        assert config.max_score_contribution == 0.40
        assert config.min_valid_score == 0.10
    
    def test_score_ranges(self):
        """Test that default scores are in valid range."""
        config = ScoringConfig()
        
        assert 0.0 <= config.barcode_read_score <= 1.0
        assert 0.0 <= config.check_digit_bonus <= 1.0
        assert 0.0 <= config.cross_validation_weight <= 1.0
        assert 0.0 <= config.max_score_contribution <= 1.0


class TestBarcodeScanningConfig:
    """Tests for BarcodeScanningConfig dataclass."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = BarcodeScanningConfig()
        
        assert config.model_version == "1.0.0"
        assert config.debug_mode is False
        assert config.timeout_seconds == 30.0
        assert config.parallel_processing is True
        assert config.max_barcodes == 5
    
    def test_component_configs(self):
        """Test that component configs are initialized."""
        config = BarcodeScanningConfig()
        
        assert config.pdf417 is not None
        assert config.mrz is not None
        assert config.cross_validation is not None
        assert config.scoring is not None
    
    def test_preprocessing_options(self):
        """Test preprocessing configuration."""
        config = BarcodeScanningConfig()
        
        assert config.enable_preprocessing is True
        assert config.grayscale_conversion is True
        assert config.contrast_enhancement is True
    
    def test_validate_valid(self):
        """Test validation of valid configuration."""
        config = BarcodeScanningConfig()
        
        errors = config.validate()
        
        assert len(errors) == 0
    
    def test_validate_invalid_max_score(self):
        """Test validation catches invalid max score."""
        config = BarcodeScanningConfig()
        config.scoring.max_score_contribution = 1.5
        
        errors = config.validate()
        
        assert len(errors) == 1
        assert "max_score_contribution" in errors[0]
    
    def test_validate_invalid_threshold(self):
        """Test validation catches invalid threshold."""
        config = BarcodeScanningConfig()
        config.cross_validation.string_similarity_threshold = 0.3
        
        errors = config.validate()
        
        assert len(errors) == 1
        assert "string_similarity_threshold" in errors[0]
    
    def test_validate_invalid_timeout(self):
        """Test validation catches invalid timeout."""
        config = BarcodeScanningConfig()
        config.timeout_seconds = 0
        
        errors = config.validate()
        
        assert len(errors) == 1
        assert "timeout_seconds" in errors[0]
    
    def test_validate_multiple_errors(self):
        """Test validation returns multiple errors."""
        config = BarcodeScanningConfig()
        config.scoring.max_score_contribution = 2.0
        config.cross_validation.string_similarity_threshold = 0.2
        config.timeout_seconds = -5
        
        errors = config.validate()
        
        assert len(errors) == 3
    
    def test_custom_configuration(self):
        """Test creating custom configuration."""
        config = BarcodeScanningConfig()
        
        # Modify PDF417 config
        config.pdf417.enable_aamva = False
        config.pdf417.detection_confidence_threshold = 0.8
        
        # Modify MRZ config
        config.mrz.strict_validation = True
        config.mrz.validate_check_digits = False
        
        # Modify cross-validation config
        config.cross_validation.string_similarity_threshold = 0.9
        config.cross_validation.max_edit_distance = 1
        
        # Verify changes
        assert config.pdf417.enable_aamva is False
        assert config.mrz.strict_validation is True
        assert config.cross_validation.max_edit_distance == 1
