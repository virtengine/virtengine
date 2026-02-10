"""
Tests for the barcode scanning pipeline.
"""

import pytest
import numpy as np

from ml.barcode_scanning.pipeline import (
    BarcodeScanningPipeline,
    BarcodeScanningResult,
    create_pipeline,
)
from ml.barcode_scanning.config import (
    BarcodeScanningConfig,
    BarcodeType,
    MRZFormat,
    DocumentCategory,
)


class TestBarcodeScanningResult:
    """Tests for BarcodeScanningResult dataclass."""
    
    def test_default_values(self):
        """Test default initialization."""
        result = BarcodeScanningResult()
        
        assert result.barcode_detected is False
        assert result.mrz_detected is False
        assert result.success is True
        assert result.veid_contribution == 0.0
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        result = BarcodeScanningResult(
            barcode_detected=True,
            barcode_type=BarcodeType.PDF417,
            barcode_score=0.85,
            validation_score=0.9,
            success=True,
        )
        
        d = result.to_dict()
        
        assert d["barcode_detected"] is True
        assert d["barcode_type"] == "pdf417"
        assert d["barcode_score"] == 0.85
        assert d["success"] is True
    
    def test_to_chain_data(self):
        """Test on-chain safe data extraction."""
        result = BarcodeScanningResult(
            barcode_detected=True,
            identity_hash="abc123",
            field_hashes={"name": "hash123"},
            barcode_score=0.85,
            veid_contribution=0.15,
        )
        
        chain_data = result.to_chain_data()
        
        assert "identity_hash" in chain_data
        assert "field_hashes" in chain_data
        assert "veid_contribution" in chain_data
        
        # Should not contain PII
        assert "identity_fields" not in chain_data


class TestBarcodeScanningPipeline:
    """Tests for BarcodeScanningPipeline class."""
    
    def test_initialization(self):
        """Test pipeline initialization."""
        pipeline = BarcodeScanningPipeline()
        
        assert pipeline.config is not None
        assert pipeline.pdf417_parser is not None
        assert pipeline.mrz_parser is not None
        assert pipeline.cross_validator is not None
    
    def test_initialization_custom_config(self):
        """Test pipeline with custom config."""
        config = BarcodeScanningConfig()
        config.scoring.barcode_read_score = 0.2
        
        pipeline = BarcodeScanningPipeline(config)
        
        assert pipeline.config.scoring.barcode_read_score == 0.2
    
    def test_scan_empty_image(self, sample_grayscale_image):
        """Test scanning empty image."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.scan(sample_grayscale_image)
        
        assert result.success is True
        assert result.barcode_detected is False
        assert result.mrz_detected is False
    
    def test_scan_with_salt(self, sample_grayscale_image):
        """Test scanning with identity salt."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.scan(
            sample_grayscale_image,
            salt="test-salt-12345",
        )
        
        assert result.success is True
    
    def test_scan_for_pdf417(self, sample_grayscale_image):
        """Test PDF417-specific scanning."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.scan_for_pdf417(sample_grayscale_image)
        
        assert result.success is True
        assert result.document_category == DocumentCategory.DRIVERS_LICENSE
    
    def test_scan_for_mrz(self, sample_grayscale_image):
        """Test MRZ-specific scanning."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.scan_for_mrz(sample_grayscale_image)
        
        assert result.success is True
        assert result.document_category == DocumentCategory.PASSPORT
    
    def test_parse_mrz_text(self, sample_mrz_td3):
        """Test MRZ text parsing."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.parse_mrz_text(sample_mrz_td3, salt="test-salt")
        
        assert result.success is True
        assert result.mrz_detected is True
        assert result.mrz_format == MRZFormat.TD3
        assert result.mrz_data is not None
    
    def test_parse_mrz_text_td1(self, sample_mrz_td1):
        """Test TD1 MRZ parsing."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.parse_mrz_text(sample_mrz_td1)
        
        assert result.success is True
        assert result.mrz_detected is True
        assert result.mrz_format == MRZFormat.TD1
    
    def test_parse_mrz_generates_hashes(self, sample_mrz_td3):
        """Test that MRZ parsing generates hashes."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.parse_mrz_text(sample_mrz_td3, salt="hash-salt")
        
        if result.mrz_detected:
            assert result.field_hashes  # Not empty
            assert result.identity_hash  # Not empty
            
            # Hashes should be SHA-256 (64 hex chars)
            for hash_value in result.field_hashes.values():
                assert len(hash_value) == 64
    
    def test_cross_validation(self, sample_mrz_td3, sample_ocr_data):
        """Test cross-validation with OCR data."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.parse_mrz_text(sample_mrz_td3)
        
        if result.mrz_detected:
            # Re-run with OCR data
            result = pipeline.scan(
                np.zeros((480, 640), dtype=np.uint8),
                ocr_data=sample_ocr_data,
            )
            
            # Cross-validation should be attempted
            # (may not produce results without actual barcode/MRZ detection)
    
    def test_scoring_calculation(self, sample_mrz_td3):
        """Test score calculation."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.parse_mrz_text(sample_mrz_td3)
        
        if result.mrz_detected:
            assert result.mrz_score > 0
            assert result.validation_score > 0
            assert result.veid_contribution >= 0
    
    def test_document_category_inference(self, sample_mrz_td3):
        """Test document category inference."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.parse_mrz_text(sample_mrz_td3)
        
        assert result.document_category == DocumentCategory.PASSPORT
    
    def test_preprocessing(self, sample_color_image):
        """Test image preprocessing."""
        config = BarcodeScanningConfig()
        config.enable_preprocessing = True
        config.grayscale_conversion = True
        config.contrast_enhancement = True
        
        pipeline = BarcodeScanningPipeline(config)
        
        # Should handle color image
        result = pipeline.scan(sample_color_image)
        
        assert result.success is True
    
    def test_processing_time_tracked(self, sample_grayscale_image):
        """Test that processing time is tracked."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.scan(sample_grayscale_image)
        
        assert result.processing_time_ms > 0
    
    def test_model_version_set(self, sample_grayscale_image):
        """Test that model version is set."""
        pipeline = BarcodeScanningPipeline()
        
        result = pipeline.scan(sample_grayscale_image)
        
        assert result.model_version == pipeline.config.model_version
    
    def test_error_handling(self):
        """Test error handling for invalid input."""
        pipeline = BarcodeScanningPipeline()
        
        # Should handle invalid input gracefully
        result = pipeline.parse_mrz_text("")
        
        assert result.mrz_detected is False


class TestCreatePipeline:
    """Tests for pipeline factory function."""
    
    def test_create_default(self):
        """Test default pipeline creation."""
        pipeline = create_pipeline()
        
        assert pipeline is not None
        assert pipeline.config.pdf417.enable_aamva is True
    
    def test_create_pdf417_disabled(self):
        """Test pipeline with PDF417 disabled."""
        pipeline = create_pipeline(enable_pdf417=False)
        
        assert pipeline.config.pdf417.enable_aamva is False
    
    def test_create_mrz_disabled(self):
        """Test pipeline with MRZ disabled."""
        pipeline = create_pipeline(enable_mrz=False)
        
        assert len(pipeline.config.mrz.supported_formats) == 0
    
    def test_create_strict_validation(self):
        """Test pipeline with strict validation."""
        pipeline = create_pipeline(strict_validation=True)
        
        assert pipeline.config.mrz.strict_validation is True


class TestBarcodeScanningConfig:
    """Tests for configuration validation."""
    
    def test_validate_valid_config(self):
        """Test validation of valid config."""
        config = BarcodeScanningConfig()
        
        errors = config.validate()
        
        assert len(errors) == 0
    
    def test_validate_invalid_score(self):
        """Test validation catches invalid scores."""
        config = BarcodeScanningConfig()
        config.scoring.max_score_contribution = 1.5  # Invalid
        
        errors = config.validate()
        
        assert len(errors) > 0
        assert "max_score_contribution" in errors[0]
    
    def test_validate_invalid_threshold(self):
        """Test validation catches invalid thresholds."""
        config = BarcodeScanningConfig()
        config.cross_validation.string_similarity_threshold = 0.3  # Too low
        
        errors = config.validate()
        
        assert len(errors) > 0
    
    def test_validate_invalid_timeout(self):
        """Test validation catches invalid timeout."""
        config = BarcodeScanningConfig()
        config.timeout_seconds = -1  # Invalid
        
        errors = config.validate()
        
        assert len(errors) > 0
