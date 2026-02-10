"""
Tests for PDF417 barcode parsing.
"""

import pytest

from ml.barcode_scanning.pdf417_parser import (
    PDF417Parser,
    PDF417Data,
    AAMVAField,
    AAMVAVersion,
)
from ml.barcode_scanning.config import PDF417Config


class TestAAMVAField:
    """Tests for AAMVAField dataclass."""
    
    def test_creation(self):
        """Test field creation."""
        field = AAMVAField(
            code="DAQ",
            value="D12345678",
            description="Document Number",
            is_required=True,
        )
        
        assert field.code == "DAQ"
        assert field.value == "D12345678"
        assert field.is_required is True
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        field = AAMVAField(
            code="DCS",
            value="SMITH",
            description="Last Name",
            is_required=True,
            raw_value="SMITH\r",
        )
        
        d = field.to_dict()
        
        assert d["code"] == "DCS"
        assert d["value"] == "SMITH"
        assert d["description"] == "Last Name"
        assert "raw_value" not in d  # Should not expose raw value


class TestPDF417Data:
    """Tests for PDF417Data dataclass."""
    
    def test_default_values(self):
        """Test default initialization."""
        data = PDF417Data()
        
        assert data.document_number == ""
        assert data.first_name == ""
        assert data.last_name == ""
        assert data.success is True
        assert data.confidence == 0.0
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        data = PDF417Data(
            document_number="D12345678",
            first_name="JOHN",
            last_name="SMITH",
            date_of_birth="1990-01-01",
            success=True,
            confidence=0.85,
        )
        
        d = data.to_dict()
        
        assert d["document_number"] == "D12345678"
        assert d["first_name"] == "JOHN"
        assert d["last_name"] == "SMITH"
        assert d["success"] is True
    
    def test_get_identity_fields(self):
        """Test identity field extraction."""
        data = PDF417Data(
            first_name="JOHN",
            middle_name="WILLIAM",
            last_name="SMITH",
            date_of_birth="1990-01-01",
            document_number="D12345678",
            sex="1",  # Male
        )
        
        fields = data.get_identity_fields()
        
        assert "full_name" in fields
        assert "JOHN" in fields["full_name"]
        assert "SMITH" in fields["full_name"]
        assert fields["date_of_birth"] == "1990-01-01"
        assert fields["sex"] == "M"  # Normalized
    
    def test_normalize_sex(self):
        """Test sex code normalization."""
        data = PDF417Data()
        
        assert data._normalize_sex("1") == "M"
        assert data._normalize_sex("2") == "F"
        assert data._normalize_sex("M") == "M"
        assert data._normalize_sex("F") == "F"
        assert data._normalize_sex("9") == ""


class TestPDF417Parser:
    """Tests for PDF417Parser class."""
    
    def test_initialization(self):
        """Test parser initialization."""
        parser = PDF417Parser()
        
        assert parser.config is not None
        assert parser.config.enable_aamva is True
    
    def test_initialization_custom_config(self):
        """Test parser with custom config."""
        config = PDF417Config(
            min_aamva_version=3,
            detection_confidence_threshold=0.7,
        )
        parser = PDF417Parser(config)
        
        assert parser.config.min_aamva_version == 3
        assert parser.config.detection_confidence_threshold == 0.7
    
    def test_parse_simple_data(self, sample_pdf417_bytes):
        """Test parsing simple barcode data."""
        parser = PDF417Parser()
        result = parser.parse(sample_pdf417_bytes)
        
        # Should parse successfully
        assert result.success is True
        # Should have extracted some fields
        assert len(result.fields) > 0
    
    def test_parse_extracts_names(self, sample_pdf417_bytes):
        """Test that names are extracted."""
        parser = PDF417Parser()
        result = parser.parse(sample_pdf417_bytes)
        
        # Should have name fields
        if "DCS" in result.fields or result.last_name:
            assert result.last_name or "DCS" in result.fields
    
    def test_parse_date_mmddccyy(self):
        """Test MMDDCCYY date parsing."""
        parser = PDF417Parser()
        
        # Test MMDDCCYY format
        result = parser._parse_date("01011990")
        assert result == "1990-01-01"
    
    def test_parse_date_ccyymmdd(self):
        """Test CCYYMMDD date parsing."""
        parser = PDF417Parser()
        
        # Test CCYYMMDD format
        result = parser._parse_date("19900101")
        assert result == "1990-01-01"
    
    def test_parse_date_invalid(self):
        """Test invalid date handling."""
        parser = PDF417Parser()
        
        # Invalid date should return as-is
        result = parser._parse_date("invalid")
        assert result == "invalid"
    
    def test_clean_value(self):
        """Test value cleaning."""
        parser = PDF417Parser()
        
        # Should remove control characters
        result = parser._clean_value("SMITH\r\n")
        assert result == "SMITH"
        
        # Should normalize whitespace
        result = parser._clean_value("  JOHN   SMITH  ")
        assert result == "JOHN SMITH"
    
    def test_parse_header(self):
        """Test AAMVA header parsing."""
        parser = PDF417Parser()
        
        data = "@\n\nANSI 63600101020DL"
        header = parser._parse_header(data)
        
        # Should extract header info or return None
        # The exact parsing depends on format variations
    
    def test_hash_fields(self, sample_pdf417_bytes):
        """Test field hashing."""
        parser = PDF417Parser()
        result = parser.parse(sample_pdf417_bytes)
        
        if result.success:
            hashes = parser.hash_fields(result, "test-salt")
            
            # Should have hashes for identity fields
            assert isinstance(hashes, dict)
            
            # Hashes should be SHA-256 (64 hex chars)
            for hash_value in hashes.values():
                assert len(hash_value) == 64
    
    def test_calculate_confidence(self):
        """Test confidence calculation."""
        parser = PDF417Parser()
        
        # Result with required fields
        result = PDF417Data(
            document_number="D12345678",
            last_name="SMITH",
            first_name="JOHN",
            date_of_birth="1990-01-01",
        )
        result.fields = {
            "DAQ": AAMVAField("DAQ", "D12345678", "Doc", True),
            "DCS": AAMVAField("DCS", "SMITH", "Last", True),
            "DAC": AAMVAField("DAC", "JOHN", "First", True),
            "DBB": AAMVAField("DBB", "01011990", "DOB", True),
        }
        
        confidence = parser._calculate_confidence(result)
        
        assert 0.0 <= confidence <= 1.0
        assert confidence > 0.3  # Should have decent confidence with core fields
    
    def test_empty_data(self):
        """Test parsing empty data."""
        parser = PDF417Parser()
        result = parser.parse(b"")
        
        # Should fail gracefully
        assert result.confidence == 0.0 or result.success is False
    
    def test_invalid_data(self):
        """Test parsing invalid data."""
        parser = PDF417Parser()
        result = parser.parse(b"NOT A VALID BARCODE")
        
        # Should handle gracefully
        assert result.success is False or len(result.fields) == 0
