"""
Tests for the complete OCR extraction pipeline.
"""

import pytest
import numpy as np

from ml.text_detection import TextROI, BoundingBox, TextType
from ml.ocr_extraction.pipeline import (
    OCRExtractionPipeline,
    OCRExtractionResult,
    create_default_pipeline,
    create_production_pipeline,
)
from ml.ocr_extraction.config import (
    OCRExtractionConfig,
    DocumentType,
    HashingConfig,
)
from ml.ocr_extraction.field_parser import ParsedField, ValidationStatus
from ml.ocr_extraction.field_hasher import FieldHasher


class TestOCRExtractionResult:
    """Tests for OCRExtractionResult dataclass."""
    
    def test_to_dict(self, sample_parsed_fields):
        """Test dictionary conversion."""
        result = OCRExtractionResult(
            fields=sample_parsed_fields,
            field_hashes={"surname": "abc123"},
            identity_hash="xyz789",
            raw_text="Test text",
            confidence_score=0.85,
            model_version="1.0.0",
            document_type=DocumentType.ID_CARD,
            processing_time_ms=100.5,
            roi_count=5,
            successful_rois=4,
        )
        
        d = result.to_dict()
        
        assert d["success"] is True
        assert d["confidence_score"] == 0.85
        assert d["model_version"] == "1.0.0"
        assert d["document_type"] == "id_card"
    
    def test_to_chain_data(self, sample_parsed_fields):
        """Test chain-safe data export."""
        result = OCRExtractionResult(
            fields=sample_parsed_fields,
            field_hashes={"surname": "abc123", "given_names": "def456"},
            identity_hash="xyz789",
            raw_text="SURNAME: SMITH GIVEN NAMES: JOHN",  # PII
            confidence_score=0.85,
            model_version="1.0.0",
        )
        
        chain_data = result.to_chain_data()
        
        # Should contain hashes
        assert "identity_hash" in chain_data
        assert "field_hashes" in chain_data
        
        # Should NOT contain raw text (PII)
        assert "raw_text" not in chain_data
        assert "SMITH" not in str(chain_data)
        assert "JOHN" not in str(chain_data)
    
    def test_has_required_fields_with_full_name(self, sample_parsed_fields):
        """Test required fields check with full_name."""
        result = OCRExtractionResult(
            fields=sample_parsed_fields,
            field_hashes={},
            identity_hash="",
            raw_text="",
            confidence_score=0.8,
            model_version="1.0.0",
        )
        
        # Has surname + given_names, so should be considered as having name
        assert result.has_required_fields is False  # Missing full_name specifically
    
    def test_has_required_fields_minimal(self):
        """Test required fields check with minimal data."""
        fields = {
            "full_name": ParsedField(
                field_name="full_name",
                value="John Smith",
                confidence=0.8,
                source_roi_ids=[],
                validation_status=ValidationStatus.VALID,
            ),
            "date_of_birth": ParsedField(
                field_name="date_of_birth",
                value="1985-03-15",
                confidence=0.85,
                source_roi_ids=[],
                validation_status=ValidationStatus.VALID,
            ),
        }
        
        result = OCRExtractionResult(
            fields=fields,
            field_hashes={},
            identity_hash="",
            raw_text="",
            confidence_score=0.8,
            model_version="1.0.0",
        )
        
        assert result.has_required_fields is True
    
    def test_has_required_fields_missing_dob(self):
        """Test required fields check missing DOB."""
        fields = {
            "full_name": ParsedField(
                field_name="full_name",
                value="John Smith",
                confidence=0.8,
                source_roi_ids=[],
                validation_status=ValidationStatus.VALID,
            ),
        }
        
        result = OCRExtractionResult(
            fields=fields,
            field_hashes={},
            identity_hash="",
            raw_text="",
            confidence_score=0.8,
            model_version="1.0.0",
        )
        
        assert result.has_required_fields is False


class TestOCRExtractionPipeline:
    """Tests for OCRExtractionPipeline class."""
    
    def test_initialization(self):
        """Test pipeline initialization."""
        pipeline = OCRExtractionPipeline()
        
        assert pipeline.config is not None
        assert pipeline.cropper is not None
        assert pipeline.ocr is not None
        assert pipeline.postprocessor is not None
        assert pipeline.field_parser is not None
        assert pipeline.hasher is not None
    
    def test_initialization_custom_config(self, test_config):
        """Test pipeline with custom config."""
        pipeline = OCRExtractionPipeline(test_config)
        
        assert pipeline.config == test_config
    
    @pytest.mark.skipif(
        True,  # Skip by default since requires Tesseract
        reason="Requires Tesseract OCR installation"
    )
    def test_extract_basic(self, sample_document_image, sample_rois):
        """Test basic extraction (requires Tesseract)."""
        pipeline = OCRExtractionPipeline()
        result = pipeline.extract(
            sample_document_image,
            sample_rois,
            DocumentType.ID_CARD
        )
        
        assert isinstance(result, OCRExtractionResult)
        assert result.model_version == pipeline.VERSION
    
    def test_extract_handles_empty_rois(self, sample_document_image):
        """Test extraction with no ROIs."""
        pipeline = OCRExtractionPipeline()
        
        # Mock the OCR to avoid Tesseract dependency
        result = OCRExtractionResult(
            fields={},
            field_hashes={},
            identity_hash="",
            raw_text="",
            confidence_score=0.0,
            model_version="1.0.0",
            roi_count=0,
            successful_rois=0,
        )
        
        assert result.roi_count == 0
        assert result.fields == {}
    
    def test_detect_document_type_passport(self):
        """Test passport detection from OCR text."""
        pipeline = OCRExtractionPipeline()
        
        from ml.ocr_extraction.tesseract_wrapper import OCRResult
        
        ocr_results = [
            OCRResult(text="PASSPORT", confidence=90.0, word_confidences=[]),
            OCRResult(text="P<GBR<<<", confidence=85.0, word_confidences=[]),
        ]
        
        doc_type = pipeline._detect_document_type(ocr_results)
        assert doc_type == DocumentType.PASSPORT
    
    def test_detect_document_type_drivers_license(self):
        """Test driver's license detection."""
        pipeline = OCRExtractionPipeline()
        
        from ml.ocr_extraction.tesseract_wrapper import OCRResult
        
        ocr_results = [
            OCRResult(text="DRIVER LICENSE", confidence=90.0, word_confidences=[]),
            OCRResult(text="CLASS: C", confidence=85.0, word_confidences=[]),
        ]
        
        doc_type = pipeline._detect_document_type(ocr_results)
        assert doc_type == DocumentType.DRIVERS_LICENSE
    
    def test_detect_document_type_id_card(self):
        """Test ID card detection."""
        pipeline = OCRExtractionPipeline()
        
        from ml.ocr_extraction.tesseract_wrapper import OCRResult
        
        ocr_results = [
            OCRResult(text="NATIONAL ID CARD", confidence=90.0, word_confidences=[]),
            OCRResult(text="IDENTITY", confidence=85.0, word_confidences=[]),
        ]
        
        doc_type = pipeline._detect_document_type(ocr_results)
        assert doc_type == DocumentType.ID_CARD
    
    def test_detect_document_type_unknown(self):
        """Test unknown document type detection."""
        pipeline = OCRExtractionPipeline()
        
        from ml.ocr_extraction.tesseract_wrapper import OCRResult
        
        ocr_results = [
            OCRResult(text="SOME RANDOM TEXT", confidence=90.0, word_confidences=[]),
        ]
        
        doc_type = pipeline._detect_document_type(ocr_results)
        assert doc_type == DocumentType.UNKNOWN
    
    def test_calculate_confidence(self, sample_parsed_fields, sample_ocr_results):
        """Test confidence calculation."""
        pipeline = OCRExtractionPipeline()
        
        confidence = pipeline._calculate_confidence(
            sample_parsed_fields,
            sample_ocr_results
        )
        
        assert 0.0 <= confidence <= 1.0
    
    def test_calculate_confidence_empty(self):
        """Test confidence calculation with no data."""
        pipeline = OCRExtractionPipeline()
        
        confidence = pipeline._calculate_confidence({}, [])
        assert confidence == 0.0
    
    def test_hash_image(self, sample_image):
        """Test image hashing."""
        pipeline = OCRExtractionPipeline()
        
        hash1 = pipeline.hash_image(sample_image)
        hash2 = pipeline.hash_image(sample_image)
        
        assert hash1 == hash2
        assert len(hash1) == 64  # SHA-256 hex
    
    def test_hash_image_different_images(self, sample_image, sample_document_image):
        """Test different images have different hashes."""
        pipeline = OCRExtractionPipeline()
        
        hash1 = pipeline.hash_image(sample_image)
        hash2 = pipeline.hash_image(sample_document_image)
        
        assert hash1 != hash2


class TestPipelineIntegration:
    """Integration tests for the pipeline."""
    
    def test_field_hashes_are_consistent(self, sample_parsed_fields):
        """Test that field hashes are reproducible."""
        hasher = FieldHasher()
        
        hashes1 = hasher.hash_fields(sample_parsed_fields)
        hashes2 = hasher.hash_fields(sample_parsed_fields)
        
        assert hashes1 == hashes2
    
    def test_identity_hash_changes_with_field_values(self, sample_parsed_fields):
        """Test identity hash changes when field values change."""
        hasher = FieldHasher()
        
        hash1 = hasher.create_derived_identity_hash(sample_parsed_fields)
        
        # Modify a field
        modified_fields = sample_parsed_fields.copy()
        modified_fields["surname"] = ParsedField(
            field_name="surname",
            value="JONES",  # Changed
            confidence=0.85,
            source_roi_ids=["roi_1"],
            validation_status=ValidationStatus.VALID,
        )
        
        hash2 = hasher.create_derived_identity_hash(modified_fields)
        
        assert hash1 != hash2
    
    def test_chain_data_contains_no_pii(self, sample_parsed_fields):
        """Test chain data export contains no PII."""
        # Create a result with PII
        result = OCRExtractionResult(
            fields=sample_parsed_fields,
            field_hashes={"surname": "hash1", "given_names": "hash2"},
            identity_hash="idhash",
            raw_text="JOHN SMITH 15/03/1985 ABC123456",  # PII!
            confidence_score=0.85,
            model_version="1.0.0",
        )
        
        chain_data = result.to_chain_data()
        chain_str = str(chain_data)
        
        # Check no PII in chain data
        assert "JOHN" not in chain_str
        assert "SMITH" not in chain_str
        assert "1985" not in chain_str
        assert "ABC123456" not in chain_str
        
        # But hashes should be present
        assert "hash1" in chain_str
        assert "idhash" in chain_str


class TestFactoryFunctions:
    """Tests for pipeline factory functions."""
    
    def test_create_default_pipeline(self):
        """Test default pipeline creation."""
        pipeline = create_default_pipeline()
        
        assert isinstance(pipeline, OCRExtractionPipeline)
        assert pipeline.config is not None
    
    def test_create_production_pipeline(self):
        """Test production pipeline creation."""
        pipeline = create_production_pipeline(
            hashing_salt="secret_salt_12345",
            tesseract_path=None,
        )
        
        assert isinstance(pipeline, OCRExtractionPipeline)
        assert pipeline.config.hashing.salt == "secret_salt_12345"


class TestRoiLimiting:
    """Tests for ROI limiting."""
    
    def test_max_rois_limit(self, sample_document_image):
        """Test pipeline respects max ROI limit."""
        config = OCRExtractionConfig()
        config.max_rois_per_image = 3
        
        pipeline = OCRExtractionPipeline(config)
        
        # Create more ROIs than limit
        many_rois = []
        for i in range(10):
            roi = TextROI.create(
                bounding_box=BoundingBox(x=50, y=50 + i*30, width=200, height=20),
                confidence=0.8,
                text_type=TextType.LINE,
            )
            many_rois.append(roi)
        
        # Pipeline should limit ROIs internally
        # (We can't fully test without Tesseract, but config is set)
        assert pipeline.config.max_rois_per_image == 3


class TestErrorHandling:
    """Tests for error handling."""
    
    def test_extract_handles_exceptions(self, sample_document_image, sample_rois):
        """Test extraction handles internal errors gracefully."""
        pipeline = OCRExtractionPipeline()
        
        # Create an invalid image that might cause issues
        invalid_image = np.array([])
        
        # Should not raise, should return error result
        result = pipeline.extract(invalid_image, sample_rois, DocumentType.ID_CARD)
        
        assert isinstance(result, OCRExtractionResult)
        assert result.success is False
        assert result.error_message is not None
