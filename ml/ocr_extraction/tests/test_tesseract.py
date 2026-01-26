"""
Tests for Tesseract OCR wrapper.

Note: These tests require Tesseract OCR to be installed.
Tests are skipped if Tesseract is not available.
"""

import pytest
import numpy as np
import cv2

try:
    import pytesseract
    TESSERACT_AVAILABLE = True
    try:
        pytesseract.get_tesseract_version()
        TESSERACT_INSTALLED = True
    except pytesseract.TesseractNotFoundError:
        TESSERACT_INSTALLED = False
except ImportError:
    TESSERACT_AVAILABLE = False
    TESSERACT_INSTALLED = False

from ml.ocr_extraction.tesseract_wrapper import (
    TesseractOCR,
    OCRResult,
    DetailedOCRResult,
    CharacterDetail,
    WordDetail,
)
from ml.ocr_extraction.config import TesseractConfig, PageSegmentationMode


pytestmark = pytest.mark.skipif(
    not TESSERACT_INSTALLED,
    reason="Tesseract OCR not installed"
)


class TestTesseractConfig:
    """Tests for TesseractConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = TesseractConfig()
        assert config.lang == "eng"
        assert config.psm == PageSegmentationMode.SINGLE_BLOCK
        assert config.oem == 3
        assert config.whitelist == ""
    
    def test_build_config_string_basic(self):
        """Test basic config string building."""
        config = TesseractConfig()
        config_str = config.build_config_string()
        
        assert "--psm 6" in config_str
        assert "--oem 3" in config_str
    
    def test_build_config_string_with_whitelist(self):
        """Test config string with character whitelist."""
        config = TesseractConfig(whitelist="0123456789")
        config_str = config.build_config_string()
        
        assert "tessedit_char_whitelist=0123456789" in config_str
    
    def test_build_config_string_with_blacklist(self):
        """Test config string with character blacklist."""
        config = TesseractConfig(blacklist="!@#$%")
        config_str = config.build_config_string()
        
        assert "tessedit_char_blacklist=!@#$%" in config_str


class TestOCRResult:
    """Tests for OCRResult dataclass."""
    
    def test_normalized_confidence(self):
        """Test confidence normalization."""
        result = OCRResult(
            text="Test",
            confidence=85.0,
            word_confidences=[85.0],
        )
        assert result.normalized_confidence == 0.85
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        result = OCRResult(
            text="Hello World",
            confidence=90.0,
            word_confidences=[92.0, 88.0],
            roi_id="roi_001",
        )
        d = result.to_dict()
        
        assert d["text"] == "Hello World"
        assert d["confidence"] == 90.0
        assert d["roi_id"] == "roi_001"


class TestDetailedOCRResult:
    """Tests for DetailedOCRResult dataclass."""
    
    def test_word_confidences(self):
        """Test word confidence extraction."""
        words = [
            WordDetail(text="Hello", confidence=90.0, x=0, y=0, width=50, height=20),
            WordDetail(text="World", confidence=85.0, x=60, y=0, width=50, height=20),
        ]
        result = DetailedOCRResult(
            text="Hello World",
            confidence=87.5,
            words=words,
            raw_data={},
        )
        
        assert result.word_confidences == [90.0, 85.0]
    
    def test_to_basic_result(self):
        """Test conversion to basic OCRResult."""
        words = [
            WordDetail(text="Test", confidence=88.0, x=0, y=0, width=40, height=20),
        ]
        detailed = DetailedOCRResult(
            text="Test",
            confidence=88.0,
            words=words,
            raw_data={},
            roi_id="roi_001",
        )
        
        basic = detailed.to_basic_result()
        
        assert isinstance(basic, OCRResult)
        assert basic.text == "Test"
        assert basic.roi_id == "roi_001"


@pytest.mark.skipif(not TESSERACT_INSTALLED, reason="Tesseract not installed")
class TestTesseractOCR:
    """Tests for TesseractOCR class."""
    
    def test_initialization(self):
        """Test OCR initialization."""
        ocr = TesseractOCR()
        assert ocr.config is not None
    
    def test_initialization_custom_config(self):
        """Test OCR initialization with custom config."""
        config = TesseractConfig(lang="eng", psm=PageSegmentationMode.SINGLE_LINE)
        ocr = TesseractOCR(config)
        assert ocr.config.psm == PageSegmentationMode.SINGLE_LINE
    
    def test_recognize_simple_image(self):
        """Test OCR on simple generated image."""
        # Create a simple image with text-like content
        image = np.ones((50, 200), dtype=np.uint8) * 255
        
        ocr = TesseractOCR()
        result = ocr.recognize(image)
        
        assert isinstance(result, OCRResult)
        # Result may be empty on blank image, but should not error
    
    def test_recognize_with_roi_id(self):
        """Test OCR preserves ROI ID."""
        image = np.ones((50, 200), dtype=np.uint8) * 255
        
        ocr = TesseractOCR()
        result = ocr.recognize(image, roi_id="test_roi_123")
        
        assert result.roi_id == "test_roi_123"
    
    def test_recognize_with_details(self):
        """Test detailed OCR recognition."""
        image = np.ones((50, 200), dtype=np.uint8) * 255
        
        ocr = TesseractOCR()
        result = ocr.recognize_with_details(image)
        
        assert isinstance(result, DetailedOCRResult)
    
    def test_recognize_line(self):
        """Test single line recognition mode."""
        image = np.ones((30, 200), dtype=np.uint8) * 255
        
        ocr = TesseractOCR()
        result = ocr.recognize_line(image)
        
        assert isinstance(result, OCRResult)
    
    def test_recognize_word(self):
        """Test single word recognition mode."""
        image = np.ones((30, 80), dtype=np.uint8) * 255
        
        ocr = TesseractOCR()
        result = ocr.recognize_word(image)
        
        assert isinstance(result, OCRResult)
    
    def test_recognize_numeric(self):
        """Test numeric-only recognition."""
        image = np.ones((30, 100), dtype=np.uint8) * 255
        
        ocr = TesseractOCR()
        result = ocr.recognize_numeric(image)
        
        assert isinstance(result, OCRResult)
        # Any text recognized should only contain digits
        cleaned = result.text.replace(" ", "")
        assert all(c.isdigit() for c in cleaned) if cleaned else True
    
    def test_recognize_alpha(self):
        """Test alphabetic-only recognition."""
        image = np.ones((30, 100), dtype=np.uint8) * 255
        
        ocr = TesseractOCR()
        result = ocr.recognize_alpha(image, include_space=False)
        
        assert isinstance(result, OCRResult)
    
    def test_recognize_date(self):
        """Test date recognition."""
        image = np.ones((30, 100), dtype=np.uint8) * 255
        
        ocr = TesseractOCR()
        result = ocr.recognize_date(image, separator="/")
        
        assert isinstance(result, OCRResult)
    
    def test_recognize_mrz(self):
        """Test MRZ recognition mode."""
        image = np.ones((30, 400), dtype=np.uint8) * 255
        
        ocr = TesseractOCR()
        result = ocr.recognize_mrz(image)
        
        assert isinstance(result, OCRResult)
    
    def test_batch_recognize(self):
        """Test batch recognition."""
        images = [
            np.ones((30, 100), dtype=np.uint8) * 255,
            np.ones((30, 150), dtype=np.uint8) * 255,
            np.ones((30, 120), dtype=np.uint8) * 255,
        ]
        roi_ids = ["roi_1", "roi_2", "roi_3"]
        
        ocr = TesseractOCR()
        results = ocr.batch_recognize(images, roi_ids)
        
        assert len(results) == 3
        assert all(isinstance(r, OCRResult) for r in results)
        assert results[0].roi_id == "roi_1"
    
    def test_color_image_conversion(self):
        """Test OCR handles color images."""
        # BGR color image
        image = np.ones((50, 200, 3), dtype=np.uint8) * 255
        
        ocr = TesseractOCR()
        result = ocr.recognize(image)
        
        assert isinstance(result, OCRResult)


class TestCharacterDetail:
    """Tests for CharacterDetail dataclass."""
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        char = CharacterDetail(
            char="A",
            confidence=95.0,
            x=10,
            y=5,
            width=15,
            height=20,
        )
        d = char.to_dict()
        
        assert d["char"] == "A"
        assert d["confidence"] == 95.0
        assert d["x"] == 10


class TestWordDetail:
    """Tests for WordDetail dataclass."""
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        word = WordDetail(
            text="Hello",
            confidence=88.0,
            x=0,
            y=0,
            width=50,
            height=20,
        )
        d = word.to_dict()
        
        assert d["text"] == "Hello"
        assert d["confidence"] == 88.0
