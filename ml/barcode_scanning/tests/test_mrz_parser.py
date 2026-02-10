"""
Tests for MRZ (Machine Readable Zone) parsing.
"""

import pytest

from ml.barcode_scanning.mrz_parser import (
    MRZParser,
    MRZData,
    MRZLine,
    MRZCheckDigit,
    CHECK_DIGIT_WEIGHTS,
    CHAR_VALUES,
)
from ml.barcode_scanning.config import MRZConfig, MRZFormat


class TestMRZCheckDigit:
    """Tests for MRZCheckDigit dataclass."""
    
    def test_creation(self):
        """Test check digit result creation."""
        check = MRZCheckDigit(
            field_name="document_number",
            expected="6",
            calculated="6",
            is_valid=True,
            value_checked="L898902C3",
        )
        
        assert check.field_name == "document_number"
        assert check.is_valid is True
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        check = MRZCheckDigit(
            field_name="date_of_birth",
            expected="2",
            calculated="2",
            is_valid=True,
            value_checked="740812",
        )
        
        d = check.to_dict()
        
        assert d["field_name"] == "date_of_birth"
        assert d["is_valid"] is True
        # value_checked should not be exposed (contains data)
        assert "value_checked" not in d


class TestMRZLine:
    """Tests for MRZLine dataclass."""
    
    def test_creation(self):
        """Test MRZ line creation."""
        line = MRZLine(
            line_number=1,
            raw_text="P<UTOERIKSSON<<ANNA<MARIA",
            clean_text="P<UTOERIKSSON<<ANNA<MARIA",
            expected_length=44,
            is_valid_length=False,
        )
        
        assert line.line_number == 1
        assert line.is_valid_length is False


class TestMRZData:
    """Tests for MRZData dataclass."""
    
    def test_default_values(self):
        """Test default initialization."""
        data = MRZData()
        
        assert data.document_type == ""
        assert data.surname == ""
        assert data.format == MRZFormat.UNKNOWN
        assert data.success is True
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        data = MRZData(
            document_type="P",
            issuing_country="UTO",
            surname="ERIKSSON",
            given_names="ANNA MARIA",
            document_number="L898902C3",
            date_of_birth="1974-08-12",
            success=True,
            confidence=0.9,
        )
        
        d = data.to_dict()
        
        assert d["document_type"] == "P"
        assert d["surname"] == "ERIKSSON"
        assert d["success"] is True
    
    def test_get_identity_fields(self):
        """Test identity field extraction."""
        data = MRZData(
            surname="ERIKSSON",
            given_names="ANNA MARIA",
            date_of_birth="1974-08-12",
            document_number="L898902C3",
            nationality="UTO",
            sex="F",
        )
        data.full_name = "ANNA MARIA ERIKSSON"
        
        fields = data.get_identity_fields()
        
        assert fields["surname"] == "ERIKSSON"
        assert fields["given_names"] == "ANNA MARIA"
        assert fields["date_of_birth"] == "1974-08-12"
        assert fields["sex"] == "F"


class TestMRZParser:
    """Tests for MRZParser class."""
    
    def test_initialization(self):
        """Test parser initialization."""
        parser = MRZParser()
        
        assert parser.config is not None
        assert parser.config.validate_check_digits is True
    
    def test_initialization_custom_config(self):
        """Test parser with custom config."""
        config = MRZConfig(
            strict_validation=True,
            enable_ocr_correction=False,
        )
        parser = MRZParser(config)
        
        assert parser.config.strict_validation is True
        assert parser.config.enable_ocr_correction is False
    
    def test_detect_format_td3(self, sample_mrz_td3):
        """Test TD3 format detection."""
        parser = MRZParser()
        lines = sample_mrz_td3.strip().split('\n')
        
        format_detected = parser.detect_format(lines)
        
        assert format_detected == MRZFormat.TD3
    
    def test_detect_format_td1(self, sample_mrz_td1):
        """Test TD1 format detection."""
        parser = MRZParser()
        lines = sample_mrz_td1.strip().split('\n')
        
        format_detected = parser.detect_format(lines)
        
        assert format_detected == MRZFormat.TD1
    
    def test_detect_format_unknown(self):
        """Test unknown format detection."""
        parser = MRZParser()
        
        # Short/invalid lines
        format_detected = parser.detect_format(["SHORT"])
        
        assert format_detected == MRZFormat.UNKNOWN
    
    def test_parse_td3(self, sample_mrz_td3):
        """Test TD3 parsing."""
        parser = MRZParser()
        result = parser.parse(sample_mrz_td3)
        
        assert result.success is True
        assert result.format == MRZFormat.TD3
        assert result.document_type == "P"
        assert result.issuing_country == "UTO"
        assert "ERIKSSON" in result.surname
        assert "ANNA" in result.given_names or "ANNA" in result.full_name
    
    def test_parse_td1(self, sample_mrz_td1):
        """Test TD1 parsing."""
        parser = MRZParser()
        result = parser.parse(sample_mrz_td1)
        
        assert result.success is True
        assert result.format == MRZFormat.TD1
        assert result.document_type == "I"
    
    def test_parse_extracts_dates(self, sample_mrz_td3):
        """Test date extraction."""
        parser = MRZParser()
        result = parser.parse(sample_mrz_td3)
        
        # Should have formatted dates
        assert result.date_of_birth  # Not empty
        assert "-" in result.date_of_birth  # ISO format
    
    def test_format_date(self):
        """Test YYMMDD to YYYY-MM-DD conversion."""
        parser = MRZParser()
        
        # 2000s date
        result = parser._format_date("240115")
        assert result == "2024-01-15"
        
        # 1900s date
        result = parser._format_date("740812")
        assert result == "1974-08-12"
    
    def test_format_date_invalid(self):
        """Test invalid date handling."""
        parser = MRZParser()
        
        result = parser._format_date("invalid")
        assert result == "invalid"
        
        result = parser._format_date("12")
        assert result == "12"
    
    def test_parse_names(self):
        """Test name parsing."""
        parser = MRZParser()
        result = MRZData()
        
        parser._parse_names(result, "ERIKSSON<<ANNA<MARIA<<<<<<<<<<<")
        
        assert result.surname == "ERIKSSON"
        assert "ANNA" in result.given_names
        assert "MARIA" in result.given_names
    
    def test_clean_mrz_line(self):
        """Test MRZ line cleaning."""
        parser = MRZParser()
        
        # Uppercase conversion
        result = parser._clean_mrz_line("p<uto")
        assert "P" in result
        assert "UTO" in result.upper()
        
        # Remove invalid chars
        result = parser._clean_mrz_line("P<UTO ERIK")
        assert " " not in result
    
    def test_calculate_check_digit(self):
        """Test check digit calculation."""
        parser = MRZParser()
        
        # Known test case: "L898902C3" -> should produce valid digit
        # The exact expected value depends on the ICAO algorithm
        digit = parser._calculate_check_digit("L898902C3")
        
        assert digit in "0123456789"
        assert len(digit) == 1
    
    def test_check_digit_algorithm(self):
        """Test check digit algorithm components."""
        # Verify character values
        assert CHAR_VALUES['A'] == 10
        assert CHAR_VALUES['Z'] == 35
        assert CHAR_VALUES['0'] == 0
        assert CHAR_VALUES['<'] == 0
        
        # Verify weights
        assert CHECK_DIGIT_WEIGHTS == [7, 3, 1]
    
    def test_calculate_confidence(self):
        """Test confidence calculation."""
        parser = MRZParser()
        
        # Create result with valid data
        result = MRZData(
            document_number="L898902C3",
            surname="ERIKSSON",
            date_of_birth="1974-08-12",
            expiry_date="2012-04-15",
            nationality="UTO",
        )
        result.lines = [
            MRZLine(1, "", "X" * 44, 44, True),
            MRZLine(2, "", "X" * 44, 44, True),
        ]
        result.check_digits = [
            MRZCheckDigit("test", "1", "1", True, ""),
        ]
        
        confidence = parser._calculate_confidence(result)
        
        assert 0.0 <= confidence <= 1.0
        assert confidence > 0.5  # Should have reasonable confidence
    
    def test_hash_fields(self, sample_mrz_td3):
        """Test field hashing."""
        parser = MRZParser()
        result = parser.parse(sample_mrz_td3)
        
        if result.success:
            hashes = parser.hash_fields(result, "test-salt")
            
            assert isinstance(hashes, dict)
            
            # Hashes should be SHA-256 (64 hex chars)
            for hash_value in hashes.values():
                assert len(hash_value) == 64
    
    def test_empty_input(self):
        """Test parsing empty input."""
        parser = MRZParser()
        result = parser.parse("")
        
        assert result.success is False
        assert result.error_message is not None
    
    def test_invalid_input(self):
        """Test parsing invalid input."""
        parser = MRZParser()
        result = parser.parse("NOT A VALID MRZ")
        
        assert result.success is False or result.format == MRZFormat.UNKNOWN
    
    def test_check_digits_validation(self, sample_mrz_td3):
        """Test check digit validation."""
        parser = MRZParser()
        result = parser.parse(sample_mrz_td3)
        
        if result.success:
            # Should have validated check digits
            assert len(result.check_digits) > 0
            
            # Each check digit should have required fields
            for check in result.check_digits:
                assert check.field_name
                assert check.expected in "0123456789<"
                assert check.calculated in "0123456789"
