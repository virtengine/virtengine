"""
Tests for OCR post-processing functionality.
"""

import pytest

from ml.ocr_extraction.postprocessing import (
    OCRPostProcessor,
    CorrectionResult,
    CHARACTER_CONFUSIONS,
    MRZ_CORRECTIONS,
    create_name_corrector,
    create_id_corrector,
    create_mrz_corrector,
)
from ml.ocr_extraction.config import PostProcessingConfig


class TestOCRPostProcessor:
    """Tests for OCRPostProcessor class."""
    
    def test_initialization_default(self):
        """Test default initialization."""
        processor = OCRPostProcessor()
        assert processor.config is not None
    
    def test_initialization_custom(self):
        """Test custom initialization."""
        config = PostProcessingConfig(convert_to_uppercase=True)
        processor = OCRPostProcessor(config)
        assert processor.config.convert_to_uppercase is True
    
    def test_process_basic(self):
        """Test basic processing."""
        processor = OCRPostProcessor()
        result = processor.process("Hello World")
        assert result == "Hello World"
    
    def test_process_normalize_whitespace(self):
        """Test whitespace normalization."""
        config = PostProcessingConfig(normalize_whitespace=True)
        processor = OCRPostProcessor(config)
        
        result = processor.process("Hello    World")
        assert result == "Hello World"
    
    def test_process_strip_leading_trailing(self):
        """Test stripping leading/trailing whitespace."""
        config = PostProcessingConfig(strip_leading_trailing=True)
        processor = OCRPostProcessor(config)
        
        result = processor.process("  Hello World  ")
        assert result == "Hello World"
    
    def test_process_remoVIRTENGINE_newlines(self):
        """Test newline removal."""
        config = PostProcessingConfig(remoVIRTENGINE_newlines=True)
        processor = OCRPostProcessor(config)
        
        result = processor.process("Hello\nWorld\r\n")
        assert "\n" not in result
        assert "\r" not in result
    
    def test_process_uppercase(self):
        """Test uppercase conversion."""
        config = PostProcessingConfig(convert_to_uppercase=True)
        processor = OCRPostProcessor(config)
        
        result = processor.process("Hello World")
        assert result == "HELLO WORLD"
    
    def test_process_lowercase(self):
        """Test lowercase conversion."""
        config = PostProcessingConfig(convert_to_lowercase=True)
        processor = OCRPostProcessor(config)
        
        result = processor.process("Hello World")
        assert result == "hello world"
    
    def test_correct_common_errors_numeric(self):
        """Test numeric context error correction."""
        processor = OCRPostProcessor()
        
        # O -> 0 in numeric context
        result = processor.correct_common_errors("12O45", "numeric")
        assert result.corrected == "12045"
        assert result.was_corrected is True
    
    def test_correct_common_errors_alpha(self):
        """Test alphabetic context error correction."""
        processor = OCRPostProcessor()
        
        # 0 -> O in alpha context
        result = processor.correct_common_errors("HELL0", "alpha")
        assert result.corrected == "HELLO"
    
    def test_correct_common_errors_general(self):
        """Test general error correction."""
        processor = OCRPostProcessor()
        
        # Em-dash to hyphen
        result = processor.correct_common_errors("2020—2021", "general")
        assert result.corrected == "2020-2021"
    
    def test_correct_in_context_automatic(self):
        """Test automatic context detection for corrections."""
        config = PostProcessingConfig(enable_confusion_correction=True)
        processor = OCRPostProcessor(config)
        
        # Mixed context: O surrounded by digits should become 0
        result = processor.correct_in_context("12O45")
        assert result == "12045"
    
    def test_normalize_whitespace(self):
        """Test whitespace normalization method."""
        processor = OCRPostProcessor()
        
        result = processor.normalize_whitespace("  hello   world  ")
        assert result == " hello world "
    
    def test_normalize_whitespace_tabs(self):
        """Test tab normalization."""
        processor = OCRPostProcessor()
        
        result = processor.normalize_whitespace("hello\tworld")
        assert result == "hello world"
    
    def test_apply_field_patterns_date(self):
        """Test date field pattern processing."""
        processor = OCRPostProcessor()
        
        result = processor.apply_field_patterns("l5/O3/l985", "date")
        assert result == "15/03/1985"
    
    def test_apply_field_patterns_name(self):
        """Test name field pattern processing."""
        processor = OCRPostProcessor()
        
        result = processor.apply_field_patterns("john smith", "name")
        assert result == "John Smith"
    
    def test_apply_field_patterns_id_number(self):
        """Test ID number field pattern processing."""
        processor = OCRPostProcessor()
        
        result = processor.apply_field_patterns("abc-123-456", "id_number")
        assert result == "ABC-123-456"
    
    def test_apply_field_patterns_mrz(self):
        """Test MRZ field pattern processing."""
        processor = OCRPostProcessor()
        
        result = processor.apply_field_patterns("P<GBR SMITH<<JOHN<", "mrz")
        assert " " not in result
        assert result == "P<GBRSMITH<<JOHN<"
    
    def test_extract_numbers(self):
        """Test number extraction."""
        processor = OCRPostProcessor()
        
        result = processor.extract_numbers("ABC123DEF456")
        assert result == "123456"
    
    def test_extract_letters(self):
        """Test letter extraction."""
        processor = OCRPostProcessor()
        
        result = processor.extract_letters("ABC123DEF456")
        assert result == "ABCDEF"
    
    def test_split_by_pattern(self):
        """Test pattern splitting."""
        processor = OCRPostProcessor()
        
        result = processor.split_by_pattern("ABC123DEF456", r"[A-Z]+")
        assert result == ["ABC", "DEF"]
    
    def test_date_processing_various_formats(self):
        """Test date processing with various formats."""
        processor = OCRPostProcessor()
        
        # DD/MM/YYYY
        assert "15/03/1985" in processor.apply_field_patterns("15/03/1985", "date")
        
        # With OCR errors
        result = processor.apply_field_patterns("l5/O3/85", "date")
        assert "15/03" in result
    
    def test_name_processing_mc_names(self):
        """Test McXxx name processing."""
        processor = OCRPostProcessor()
        
        result = processor.apply_field_patterns("mcdonald", "name")
        assert result == "Mcdonald"  # Title case first
    
    def test_mrz_processing_spaces(self):
        """Test MRZ removes spaces."""
        processor = OCRPostProcessor()
        
        result = processor.apply_field_patterns("P< GBR SMITH", "mrz")
        assert result == "P<GBRSMITH"
    
    def test_custom_replacements(self):
        """Test custom replacement rules."""
        config = PostProcessingConfig(
            custom_replacements={"foo": "bar", "baz": "qux"}
        )
        processor = OCRPostProcessor(config)
        
        result = processor.process("foo and baz")
        assert result == "bar and qux"


class TestCorrectionResult:
    """Tests for CorrectionResult dataclass."""
    
    def test_was_corrected_true(self):
        """Test was_corrected when corrections made."""
        result = CorrectionResult(
            original="12O45",
            corrected="12045",
            corrections_made=[(2, "O", "0")],
        )
        assert result.was_corrected is True
    
    def test_was_corrected_false(self):
        """Test was_corrected when no corrections."""
        result = CorrectionResult(
            original="12345",
            corrected="12345",
            corrections_made=[],
        )
        assert result.was_corrected is False


class TestFactoryFunctions:
    """Tests for corrector factory functions."""
    
    def test_create_name_corrector(self):
        """Test name corrector factory."""
        corrector = create_name_corrector()
        
        assert corrector.config.enable_confusion_correction is True
        assert corrector.config.convert_to_uppercase is False
    
    def test_create_id_corrector(self):
        """Test ID corrector factory."""
        corrector = create_id_corrector()
        
        assert corrector.config.convert_to_uppercase is True
        assert corrector.config.remoVIRTENGINE_special_chars is True
    
    def test_create_mrz_corrector(self):
        """Test MRZ corrector factory."""
        corrector = create_mrz_corrector()
        
        assert corrector.config.convert_to_uppercase is True
        assert corrector.config.normalize_whitespace is False


class TestCharacterConfusions:
    """Tests for character confusion mappings."""
    
    def test_numeric_confusions_exist(self):
        """Test numeric confusions are defined."""
        assert "O" in CHARACTER_CONFUSIONS["numeric"]
        assert CHARACTER_CONFUSIONS["numeric"]["O"] == "0"
    
    def test_alpha_confusions_exist(self):
        """Test alpha confusions are defined."""
        assert "0" in CHARACTER_CONFUSIONS["alpha"]
        assert CHARACTER_CONFUSIONS["alpha"]["0"] == "O"
    
    def test_general_confusions_exist(self):
        """Test general confusions are defined."""
        assert "—" in CHARACTER_CONFUSIONS["general"]
        assert CHARACTER_CONFUSIONS["general"]["—"] == "-"
    
    def test_mrz_corrections_exist(self):
        """Test MRZ corrections are defined."""
        assert " " in MRZ_CORRECTIONS
        assert MRZ_CORRECTIONS[" "] == ""


class TestEdgeCases:
    """Tests for edge cases in post-processing."""
    
    def test_empty_string(self):
        """Test processing empty string."""
        processor = OCRPostProcessor()
        result = processor.process("")
        assert result == ""
    
    def test_whitespace_only(self):
        """Test processing whitespace-only string."""
        processor = OCRPostProcessor()
        result = processor.process("   ")
        assert result == ""
    
    def test_special_characters_only(self):
        """Test processing special characters."""
        processor = OCRPostProcessor()
        result = processor.process("@#$%")
        assert result is not None
    
    def test_unicode_characters(self):
        """Test processing unicode characters."""
        processor = OCRPostProcessor()
        result = processor.process("Café résumé")
        assert result is not None
    
    def test_mixed_line_endings(self):
        """Test processing mixed line endings."""
        config = PostProcessingConfig(remoVIRTENGINE_newlines=True)
        processor = OCRPostProcessor(config)
        
        result = processor.process("Line1\nLine2\r\nLine3\rLine4")
        assert "\n" not in result
        assert "\r" not in result
