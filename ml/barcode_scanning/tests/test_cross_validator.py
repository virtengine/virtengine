"""
Tests for cross-validation between barcode and OCR data.
"""

import pytest

from ml.barcode_scanning.cross_validator import (
    CrossValidator,
    CrossValidationResult,
    FieldMatch,
    MatchType,
    calculate_combined_identity_hash,
)
from ml.barcode_scanning.config import CrossValidationConfig
from ml.barcode_scanning.mrz_parser import MRZData
from ml.barcode_scanning.pdf417_parser import PDF417Data


class TestFieldMatch:
    """Tests for FieldMatch dataclass."""
    
    def test_creation(self):
        """Test field match creation."""
        match = FieldMatch(
            field_name="surname",
            barcode_value="ERIKSSON",
            ocr_value="ERIKSSON",
            match_type=MatchType.EXACT,
            similarity_score=1.0,
            weight=0.9,
            contribution=0.9,
        )
        
        assert match.field_name == "surname"
        assert match.match_type == MatchType.EXACT
        assert match.similarity_score == 1.0
    
    def test_to_dict_no_pii(self):
        """Test dictionary conversion excludes PII."""
        match = FieldMatch(
            field_name="surname",
            barcode_value="ERIKSSON",
            ocr_value="ERIKSSON",
            match_type=MatchType.EXACT,
            similarity_score=1.0,
            weight=0.9,
            contribution=0.9,
        )
        
        d = match.to_dict()
        
        # Should NOT contain actual values
        assert "barcode_value" not in d
        assert "ocr_value" not in d
        assert "ERIKSSON" not in str(d)
        
        # Should contain metadata
        assert d["field_name"] == "surname"
        assert d["match_type"] == "exact"


class TestCrossValidationResult:
    """Tests for CrossValidationResult dataclass."""
    
    def test_default_values(self):
        """Test default initialization."""
        result = CrossValidationResult()
        
        assert result.score == 0.0
        assert result.is_valid is False
        assert result.success is True
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        result = CrossValidationResult(
            score=0.85,
            is_valid=True,
            confidence=0.9,
            total_fields_compared=5,
            exact_matches=4,
            fuzzy_matches=1,
        )
        
        d = result.to_dict()
        
        assert d["score"] == 0.85
        assert d["is_valid"] is True
        assert d["exact_matches"] == 4
    
    def test_to_chain_data(self):
        """Test on-chain safe data extraction."""
        result = CrossValidationResult(
            score=0.85,
            is_valid=True,
            exact_matches=4,
            total_fields_compared=5,
            veid_score_contribution=0.15,
        )
        
        chain_data = result.to_chain_data()
        
        assert "cross_validation_score" in chain_data
        assert "cross_validation_valid" in chain_data
        assert "veid_contribution" in chain_data


class TestMatchType:
    """Tests for MatchType enum."""
    
    def test_enum_values(self):
        """Test enum value strings."""
        assert MatchType.EXACT.value == "exact"
        assert MatchType.FUZZY.value == "fuzzy"
        assert MatchType.PARTIAL.value == "partial"
        assert MatchType.MISMATCH.value == "mismatch"
        assert MatchType.MISSING_BARCODE.value == "missing_barcode"
        assert MatchType.MISSING_OCR.value == "missing_ocr"


class TestCrossValidator:
    """Tests for CrossValidator class."""
    
    def test_initialization(self):
        """Test validator initialization."""
        validator = CrossValidator()
        
        assert validator.config is not None
        assert validator.config.enable_fuzzy_matching is True
    
    def test_initialization_custom_config(self):
        """Test validator with custom config."""
        config = CrossValidationConfig(
            string_similarity_threshold=0.9,
            max_edit_distance=1,
        )
        validator = CrossValidator(config)
        
        assert validator.config.string_similarity_threshold == 0.9
        assert validator.config.max_edit_distance == 1
    
    def test_validate_exact_match(self, sample_ocr_data):
        """Test validation with exact matching data."""
        validator = CrossValidator()
        
        # Create matching barcode data
        barcode_data = {
            "full_name": "ANNA MARIA ERIKSSON",
            "surname": "ERIKSSON",
            "given_names": "ANNA MARIA",
            "date_of_birth": "1974-08-12",
            "document_number": "L898902C3",
        }
        
        result = validator.validate(barcode_data, sample_ocr_data)
        
        assert result.success is True
        assert result.score > 0.8
        assert result.exact_matches > 0
        assert result.is_valid is True
    
    def test_validate_fuzzy_match(self, sample_ocr_data_mismatch):
        """Test validation with fuzzy matching data."""
        validator = CrossValidator()
        
        barcode_data = {
            "full_name": "ANNA MARIA ERIKSSON",
            "surname": "ERIKSSON",
            "given_names": "ANNA MARIA",
            "date_of_birth": "1974-08-12",
        }
        
        result = validator.validate(barcode_data, sample_ocr_data_mismatch)
        
        assert result.success is True
        # Should still have some matches due to fuzzy matching
        assert result.exact_matches + result.fuzzy_matches > 0
    
    def test_validate_mismatch(self):
        """Test validation with mismatching data."""
        validator = CrossValidator()
        
        barcode_data = {
            "full_name": "JOHN SMITH",
            "date_of_birth": "1980-05-20",
        }
        
        ocr_data = {
            "full_name": "JANE DOE",
            "date_of_birth": "1990-10-15",
        }
        
        result = validator.validate(barcode_data, ocr_data)
        
        assert result.success is True
        assert result.mismatches > 0
        assert result.score < 0.5
    
    def test_validate_missing_fields(self):
        """Test validation with missing fields."""
        validator = CrossValidator()
        
        barcode_data = {
            "full_name": "ANNA ERIKSSON",
            "date_of_birth": "1974-08-12",
        }
        
        ocr_data = {
            "full_name": "ANNA ERIKSSON",
            # date_of_birth missing
        }
        
        result = validator.validate(barcode_data, ocr_data)
        
        assert result.success is True
        assert result.missing_fields > 0
    
    def test_validate_empty_barcode(self):
        """Test validation with empty barcode data."""
        validator = CrossValidator()
        
        result = validator.validate({}, {"full_name": "JOHN"})
        
        assert result.success is False
        assert "No barcode fields" in result.error_message
    
    def test_validate_empty_ocr(self):
        """Test validation with empty OCR data."""
        validator = CrossValidator()
        
        result = validator.validate({"full_name": "JOHN"}, {})
        
        assert result.success is False
        assert "No OCR fields" in result.error_message
    
    def test_validate_with_mrz_data(self, sample_ocr_data):
        """Test validation with MRZData object."""
        validator = CrossValidator()
        
        mrz_data = MRZData(
            surname="ERIKSSON",
            given_names="ANNA MARIA",
            date_of_birth="1974-08-12",
            document_number="L898902C3",
        )
        mrz_data.full_name = "ANNA MARIA ERIKSSON"
        
        result = validator.validate(mrz_data, sample_ocr_data)
        
        assert result.success is True
        assert result.score > 0.5
    
    def test_validate_with_pdf417_data(self, sample_ocr_data):
        """Test validation with PDF417Data object."""
        validator = CrossValidator()
        
        pdf417_data = PDF417Data(
            first_name="ANNA",
            middle_name="MARIA",
            last_name="ERIKSSON",
            date_of_birth="1974-08-12",
            document_number="L898902C3",
        )
        
        result = validator.validate(pdf417_data, sample_ocr_data)
        
        assert result.success is True
    
    def test_compare_values_exact(self):
        """Test exact value comparison."""
        validator = CrossValidator()
        
        match_type, similarity = validator._compare_values(
            "surname", "SMITH", "SMITH"
        )
        
        assert match_type == MatchType.EXACT
        assert similarity == 1.0
    
    def test_compare_values_case_insensitive(self):
        """Test case-insensitive comparison."""
        config = CrossValidationConfig(case_sensitive=False)
        validator = CrossValidator(config)
        
        match_type, similarity = validator._compare_values(
            "surname", "smith", "SMITH"
        )
        
        assert match_type == MatchType.EXACT
    
    def test_compare_values_fuzzy_name(self):
        """Test fuzzy name matching."""
        validator = CrossValidator()
        
        # One character difference
        match_type, similarity = validator._compare_values(
            "surname", "ERIKSSON", "ERIKSON"
        )
        
        assert match_type in (MatchType.FUZZY, MatchType.PARTIAL)
        assert similarity > 0.7
    
    def test_compare_values_date_flexible(self):
        """Test flexible date matching."""
        config = CrossValidationConfig(flexible_date_matching=True)
        validator = CrossValidator(config)
        
        # Different formats, same date
        match_type, similarity = validator._compare_values(
            "date_of_birth", "19740812", "1974-08-12"
        )
        
        # Should recognize as same date
        assert similarity > 0.5
    
    def test_normalize_value_sex(self):
        """Test sex/gender normalization."""
        validator = CrossValidator()
        
        assert validator._normalize_value("sex", "MALE") == "M"
        assert validator._normalize_value("sex", "FEMALE") == "F"
        assert validator._normalize_value("sex", "1") == "M"
        assert validator._normalize_value("sex", "2") == "F"
    
    def test_levenshtein_distance(self):
        """Test Levenshtein distance calculation."""
        validator = CrossValidator()
        
        # Same string
        assert validator._levenshtein_distance("SMITH", "SMITH") == 0
        
        # One character different
        assert validator._levenshtein_distance("SMITH", "SMYTH") == 1
        
        # Multiple differences
        distance = validator._levenshtein_distance("ERIKSSON", "ERIKSON")
        assert distance == 1
    
    def test_calculate_similarity(self):
        """Test Jaccard similarity calculation."""
        validator = CrossValidator()
        
        # Identical strings
        sim = validator._calculate_similarity("SMITH", "SMITH")
        assert sim == 1.0
        
        # Similar strings
        sim = validator._calculate_similarity("ERIKSSON", "ERIKSON")
        assert sim > 0.7
        
        # Different strings
        sim = validator._calculate_similarity("JOHN", "MARY")
        assert sim < 0.3
    
    def test_required_fields_validation(self):
        """Test required fields enforcement."""
        config = CrossValidationConfig(
            required_fields={"date_of_birth"},
        )
        validator = CrossValidator(config)
        
        # Missing required field in OCR
        barcode_data = {"date_of_birth": "1990-01-01"}
        ocr_data = {"full_name": "JOHN"}
        
        result = validator.validate(barcode_data, ocr_data)
        
        assert result.required_fields_matched is False
        assert "date_of_birth" in result.required_fields_missing
    
    def test_veid_contribution_valid(self, sample_ocr_data):
        """Test VEID score contribution for valid result."""
        validator = CrossValidator()
        
        barcode_data = {
            "full_name": "ANNA MARIA ERIKSSON",
            "surname": "ERIKSSON",
            "date_of_birth": "1974-08-12",
        }
        
        result = validator.validate(barcode_data, sample_ocr_data)
        
        if result.is_valid:
            assert result.veid_score_contribution > 0


class TestCalculateCombinedIdentityHash:
    """Tests for combined identity hash function."""
    
    def test_hash_generation(self):
        """Test combined hash generation."""
        mrz_data = MRZData(
            surname="ERIKSSON",
            given_names="ANNA",
            date_of_birth="1974-08-12",
        )
        mrz_data.full_name = "ANNA ERIKSSON"
        
        ocr_data = {
            "full_name": "ANNA ERIKSSON",
            "nationality": "UTO",
        }
        
        hash_value = calculate_combined_identity_hash(
            mrz_data, ocr_data, "test-salt"
        )
        
        # Should be SHA-256 (64 hex chars)
        assert len(hash_value) == 64
        assert all(c in "0123456789abcdef" for c in hash_value)
    
    def test_hash_deterministic(self):
        """Test that hash is deterministic."""
        mrz_data = MRZData(surname="ERIKSSON")
        ocr_data = {"surname": "ERIKSSON"}
        
        hash1 = calculate_combined_identity_hash(mrz_data, ocr_data, "salt")
        hash2 = calculate_combined_identity_hash(mrz_data, ocr_data, "salt")
        
        assert hash1 == hash2
    
    def test_hash_salt_sensitive(self):
        """Test that hash changes with different salt."""
        mrz_data = MRZData(surname="ERIKSSON")
        ocr_data = {"surname": "ERIKSSON"}
        
        hash1 = calculate_combined_identity_hash(mrz_data, ocr_data, "salt1")
        hash2 = calculate_combined_identity_hash(mrz_data, ocr_data, "salt2")
        
        assert hash1 != hash2
