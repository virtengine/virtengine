"""
Tests for document field parsing functionality.
"""

import pytest

from ml.ocr_extraction.field_parser import (
    DocumentFieldParser,
    ParsedField,
    ValidationStatus,
    MRZData,
)
from ml.ocr_extraction.tesseract_wrapper import OCRResult
from ml.ocr_extraction.config import FieldParserConfig, DocumentType


class TestParsedField:
    """Tests for ParsedField dataclass."""
    
    def test_is_valid(self):
        """Test is_valid property."""
        field = ParsedField(
            field_name="test",
            value="value",
            confidence=0.9,
            source_roi_ids=["roi_1"],
            validation_status=ValidationStatus.VALID,
        )
        assert field.is_valid is True
    
    def test_is_valid_false(self):
        """Test is_valid returns False for non-valid status."""
        field = ParsedField(
            field_name="test",
            value="value",
            confidence=0.5,
            source_roi_ids=["roi_1"],
            validation_status=ValidationStatus.UNCERTAIN,
        )
        assert field.is_valid is False
    
    def test_is_usable(self):
        """Test is_usable property."""
        # Valid is usable
        valid_field = ParsedField(
            field_name="test",
            value="value",
            confidence=0.9,
            source_roi_ids=[],
            validation_status=ValidationStatus.VALID,
        )
        assert valid_field.is_usable is True
        
        # Uncertain is usable
        uncertain_field = ParsedField(
            field_name="test",
            value="value",
            confidence=0.5,
            source_roi_ids=[],
            validation_status=ValidationStatus.UNCERTAIN,
        )
        assert uncertain_field.is_usable is True
        
        # Invalid is not usable
        invalid_field = ParsedField(
            field_name="test",
            value="x",
            confidence=0.3,
            source_roi_ids=[],
            validation_status=ValidationStatus.INVALID,
        )
        assert invalid_field.is_usable is False
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        field = ParsedField(
            field_name="surname",
            value="SMITH",
            confidence=0.85,
            source_roi_ids=["roi_1", "roi_2"],
            validation_status=ValidationStatus.VALID,
            raw_value="SMlTH",
        )
        d = field.to_dict()
        
        assert d["field_name"] == "surname"
        assert d["value"] == "SMITH"
        assert d["confidence"] == 0.85
        assert d["validation_status"] == "valid"


class TestDocumentFieldParser:
    """Tests for DocumentFieldParser class."""
    
    def test_initialization(self):
        """Test parser initialization."""
        parser = DocumentFieldParser()
        assert parser.config is not None
    
    def test_initialization_custom_config(self):
        """Test parser with custom config."""
        config = FieldParserConfig(min_field_confidence=0.7)
        parser = DocumentFieldParser(config)
        assert parser.config.min_field_confidence == 0.7
    
    def test_parse_dispatches_to_document_type(self, sample_ocr_results):
        """Test parse method dispatches to correct parser."""
        parser = DocumentFieldParser()
        
        # ID card
        fields = parser.parse(sample_ocr_results, DocumentType.ID_CARD)
        assert isinstance(fields, dict)
        
        # Passport
        fields = parser.parse(sample_ocr_results, DocumentType.PASSPORT)
        assert isinstance(fields, dict)
        
        # Driver's license
        fields = parser.parse(sample_ocr_results, DocumentType.DRIVERS_LICENSE)
        assert isinstance(fields, dict)
    
    def test_parse_id_card(self, sample_ocr_results):
        """Test ID card parsing."""
        parser = DocumentFieldParser()
        fields = parser.parse_id_card(sample_ocr_results)
        
        assert isinstance(fields, dict)
        # Should extract surname
        assert "surname" in fields
        assert fields["surname"].value == "SMITH"
    
    def test_parse_id_card_given_names(self, sample_ocr_results):
        """Test ID card extracts given names."""
        parser = DocumentFieldParser()
        fields = parser.parse_id_card(sample_ocr_results)
        
        assert "given_names" in fields
        assert "JOHN" in fields["given_names"].value
    
    def test_parse_id_card_dob(self, sample_ocr_results):
        """Test ID card extracts date of birth."""
        parser = DocumentFieldParser()
        fields = parser.parse_id_card(sample_ocr_results)
        
        assert "date_of_birth" in fields
        assert "15" in fields["date_of_birth"].value
        assert "1985" in fields["date_of_birth"].value
    
    def test_parse_id_card_document_number(self, sample_ocr_results):
        """Test ID card extracts document number."""
        parser = DocumentFieldParser()
        fields = parser.parse_id_card(sample_ocr_results)
        
        assert "document_number" in fields
        assert "ABC123456" in fields["document_number"].value
    
    def test_parse_passport_mrz(self, sample_passport_ocr_results):
        """Test passport MRZ parsing."""
        parser = DocumentFieldParser()
        fields = parser.parse_passport(sample_passport_ocr_results)
        
        # Should extract from MRZ
        assert "surname" in fields
        assert "SMITH" in fields["surname"].value
    
    def test_parse_passport_given_names_from_mrz(self, sample_passport_ocr_results):
        """Test passport extracts given names from MRZ."""
        parser = DocumentFieldParser()
        fields = parser.parse_passport(sample_passport_ocr_results)
        
        assert "given_names" in fields
        assert "JOHN" in fields["given_names"].value
    
    def test_parse_drivers_license(self, sample_ocr_results):
        """Test driver's license parsing."""
        parser = DocumentFieldParser()
        fields = parser.parse_drivers_license(sample_ocr_results)
        
        assert isinstance(fields, dict)
    
    def test_parse_generic(self, sample_ocr_results):
        """Test generic document parsing."""
        parser = DocumentFieldParser()
        fields = parser.parse_generic(sample_ocr_results)
        
        assert isinstance(fields, dict)
    
    def test_empty_ocr_results(self):
        """Test handling of empty OCR results."""
        parser = DocumentFieldParser()
        fields = parser.parse([], DocumentType.ID_CARD)
        
        assert fields == {}
    
    def test_low_confidence_ocr_results(self):
        """Test handling of low confidence results."""
        low_conf_results = [
            OCRResult(
                text="SURNAME: SMITH",
                confidence=20.0,  # Very low
                word_confidences=[20.0, 20.0],
            )
        ]
        
        parser = DocumentFieldParser()
        fields = parser.parse_id_card(low_conf_results)
        
        # Should still extract but with low confidence
        if "surname" in fields:
            assert fields["surname"].confidence < 0.5


class TestMRZData:
    """Tests for MRZData class."""
    
    def test_mrz_data_to_fields(self):
        """Test MRZ data conversion to fields."""
        mrz = MRZData(
            document_type="P",
            country_code="GBR",
            surname="SMITH",
            given_names="JOHN WILLIAM",
            document_number="AB1234567",
            nationality="GBR",
            date_of_birth="850315",
            sex="M",
            expiry_date="251231",
            personal_number="",
            check_digits_valid=True,
            raw_lines=["line1", "line2"],
        )
        
        fields = mrz.to_fields()
        
        assert "surname" in fields
        assert fields["surname"].value == "SMITH"
        assert "date_of_birth" in fields
        assert "1985" in fields["date_of_birth"].value
    
    def test_mrz_date_formatting(self):
        """Test MRZ date formatting."""
        mrz = MRZData(
            document_type="P",
            country_code="GBR",
            surname="SMITH",
            given_names="JOHN",
            document_number="AB1234567",
            nationality="GBR",
            date_of_birth="850315",  # 15 March 1985
            sex="M",
            expiry_date="251231",  # 31 Dec 2025
            personal_number="",
            check_digits_valid=True,
            raw_lines=[],
        )
        
        fields = mrz.to_fields()
        
        assert fields["date_of_birth"].value == "1985-03-15"
        assert fields["expiry_date"].value == "2025-12-31"


class TestFieldParserHeuristics:
    """Tests for field parser heuristic methods."""
    
    def test_name_heuristic(self):
        """Test name extraction heuristic."""
        parser = DocumentFieldParser()
        
        results = [
            OCRResult(
                text="John William Smith",
                confidence=85.0,
                word_confidences=[85.0, 85.0, 85.0],
                roi_id="roi_name",
            )
        ]
        
        field = parser._extract_name_heuristic(results)
        
        assert field is not None
        assert "John" in field.value or "JOHN" in field.value.upper()
    
    def test_date_heuristic(self):
        """Test date extraction heuristic."""
        parser = DocumentFieldParser()
        
        results = [
            OCRResult(
                text="15/03/1985",
                confidence=90.0,
                word_confidences=[90.0],
                roi_id="roi_date",
            )
        ]
        
        field = parser._extract_date_heuristic(results, "date_of_birth")
        
        assert field is not None
        assert "15" in field.value
        assert "1985" in field.value
    
    def test_id_number_heuristic(self):
        """Test ID number extraction heuristic."""
        parser = DocumentFieldParser()
        
        results = [
            OCRResult(
                text="AB12345678",
                confidence=88.0,
                word_confidences=[88.0],
                roi_id="roi_id",
            )
        ]
        
        field = parser._extract_id_number_heuristic(results)
        
        assert field is not None
        assert "AB12345678" in field.value
    
    def test_find_mrz_lines(self):
        """Test MRZ line detection."""
        parser = DocumentFieldParser()
        
        results = [
            OCRResult(text="Some random text", confidence=80.0, word_confidences=[]),
            OCRResult(
                text="P<GBRSMITH<<JOHN<WILLIAM<<<<<<<<<<<<<<<<<<<<<",
                confidence=90.0,
                word_confidences=[],
            ),
            OCRResult(
                text="AB1234567<8GBR8503151M2512319<<<<<<<<<<<<<<02",
                confidence=90.0,
                word_confidences=[],
            ),
        ]
        
        mrz_lines = parser._find_mrz_lines(results)
        
        assert len(mrz_lines) >= 2


class TestValidation:
    """Tests for field validation."""
    
    def test_validate_date_valid(self):
        """Test valid date validation."""
        parser = DocumentFieldParser()
        
        status = parser._validate_date("15/03/1985")
        assert status == ValidationStatus.VALID
    
    def test_validate_date_uncertain(self):
        """Test uncertain date validation."""
        parser = DocumentFieldParser()
        
        # Has date-like structure but invalid
        status = parser._validate_date("99/99/9999")
        assert status == ValidationStatus.UNCERTAIN
    
    def test_validate_date_invalid(self):
        """Test invalid date validation."""
        parser = DocumentFieldParser()
        
        status = parser._validate_date("not a date")
        assert status == ValidationStatus.INVALID
    
    def test_validate_field_empty(self):
        """Test empty field validation."""
        parser = DocumentFieldParser()
        
        status = parser._validate_field("name", "")
        assert status == ValidationStatus.NOT_FOUND
    
    def test_validate_field_too_short(self):
        """Test too short field validation."""
        parser = DocumentFieldParser()
        
        status = parser._validate_field("name", "A")
        assert status == ValidationStatus.INVALID


class TestEdgeCases:
    """Tests for edge cases in field parsing."""
    
    def test_multiple_dates_in_document(self):
        """Test handling multiple dates."""
        parser = DocumentFieldParser()
        
        results = [
            OCRResult(
                text="DOB: 15/03/1985 Exp: 31/12/2025",
                confidence=85.0,
                word_confidences=[],
            )
        ]
        
        fields = parser.parse_generic(results)
        
        # Should extract dates
        dates_found = [k for k in fields.keys() if "date" in k]
        assert len(dates_found) >= 1
    
    def test_noisy_ocr_input(self):
        """Test handling noisy OCR input."""
        parser = DocumentFieldParser()
        
        results = [
            OCRResult(
                text="SURNAM3: SM1TH",
                confidence=60.0,
                word_confidences=[60.0, 60.0],
            )
        ]
        
        fields = parser.parse_id_card(results)
        
        # Should still attempt extraction
        assert isinstance(fields, dict)
    
    def test_mixed_case_labels(self):
        """Test handling mixed case field labels."""
        parser = DocumentFieldParser()
        
        results = [
            OCRResult(
                text="SurName: SMITH",
                confidence=85.0,
                word_confidences=[85.0, 85.0],
            )
        ]
        
        fields = parser.parse_id_card(results)
        
        assert "surname" in fields
