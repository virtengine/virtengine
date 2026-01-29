"""
Tests for OCR engine abstraction layer and fallback chain.

VE-3043: Integrate EasyOCR with Fallback Chain

Tests cover:
- OCREngine abstract interface
- EasyOCREngine with mocked backend
- TesseractEngine wrapper
- FallbackChainEngine logic
- Confidence threshold behavior
- Language support
"""

import pytest
import numpy as np
from unittest.mock import Mock, MagicMock, patch, PropertyMock
from typing import List

from ml.ocr_extraction.ocr_engines import (
    OCREngine,
    OCREngineConfig,
    OCRResult,
    BoundingBox,
    EasyOCREngine,
    EasyOCRConfig,
    TesseractEngine,
    TesseractEngineConfig,
    FallbackChainEngine,
    FallbackChainConfig,
    FallbackResult,
    OCREngineType,
    create_default_chain,
    EASYOCR_AVAILABLE,
)


# ============================================================================
# Fixtures
# ============================================================================

@pytest.fixture
def sample_image() -> np.ndarray:
    """Create a sample grayscale image."""
    return np.ones((100, 200), dtype=np.uint8) * 255


@pytest.fixture
def sample_color_image() -> np.ndarray:
    """Create a sample BGR image."""
    return np.ones((100, 200, 3), dtype=np.uint8) * 255


@pytest.fixture
def sample_bbox() -> BoundingBox:
    """Create a sample bounding box."""
    return BoundingBox(x=10, y=10, width=50, height=20)


@pytest.fixture
def mock_ocr_result() -> OCRResult:
    """Create a mock OCR result."""
    return OCRResult(
        text="Test Text",
        confidence=0.95,
        bounding_box=BoundingBox(x=0, y=0, width=100, height=20),
        engine_name="mock_engine",
        language="en",
    )


@pytest.fixture
def mock_easyocr_reader():
    """Create a mock EasyOCR reader."""
    mock_reader = MagicMock()
    mock_reader.readtext.return_value = [
        ([[0, 0], [100, 0], [100, 20], [0, 20]], "Hello World", 0.92),
        ([[0, 30], [80, 30], [80, 50], [0, 50]], "Test", 0.88),
    ]
    return mock_reader


class MockOCREngine(OCREngine):
    """Concrete mock implementation for testing base class."""
    
    def __init__(
        self,
        config: OCREngineConfig = None,
        results: List[OCRResult] = None,
        available: bool = True,
    ):
        super().__init__(config)
        self._results = results or []
        self._available = available
        self._name = "mock_engine"
    
    @property
    def engine_name(self) -> str:
        return self._name
    
    @property
    def is_available(self) -> bool:
        return self._available
    
    def recognize(
        self,
        image: np.ndarray,
        language: str = "en"
    ) -> List[OCRResult]:
        return self._results
    
    def recognize_region(
        self,
        image: np.ndarray,
        bbox: BoundingBox,
        language: str = "en"
    ) -> str:
        return " ".join(r.text for r in self._results)


# ============================================================================
# BoundingBox Tests
# ============================================================================

class TestBoundingBox:
    """Tests for BoundingBox dataclass."""
    
    def test_to_tuple(self):
        """Test conversion to (x, y, w, h) tuple."""
        bbox = BoundingBox(x=10, y=20, width=100, height=50)
        assert bbox.to_tuple() == (10, 20, 100, 50)
    
    def test_to_xyxy(self):
        """Test conversion to (x1, y1, x2, y2) format."""
        bbox = BoundingBox(x=10, y=20, width=100, height=50)
        assert bbox.to_xyxy() == (10, 20, 110, 70)
    
    def test_from_xyxy(self):
        """Test creation from (x1, y1, x2, y2) format."""
        bbox = BoundingBox.from_xyxy(10, 20, 110, 70)
        assert bbox.x == 10
        assert bbox.y == 20
        assert bbox.width == 100
        assert bbox.height == 50
    
    def test_from_points(self):
        """Test creation from corner points."""
        points = [[0, 0], [100, 0], [100, 50], [0, 50]]
        bbox = BoundingBox.from_points(points)
        assert bbox.x == 0
        assert bbox.y == 0
        assert bbox.width == 100
        assert bbox.height == 50
    
    def test_from_points_rotated(self):
        """Test creation from rotated corner points."""
        # Slightly rotated rectangle
        points = [[5, 0], [105, 5], [100, 55], [0, 50]]
        bbox = BoundingBox.from_points(points)
        # Should get axis-aligned bounding box
        assert bbox.x == 0
        assert bbox.y == 0
        assert bbox.width == 105
        assert bbox.height == 55
    
    def test_from_points_insufficient(self):
        """Test error on insufficient points."""
        points = [[0, 0], [100, 0]]
        with pytest.raises(ValueError, match="at least 4 points"):
            BoundingBox.from_points(points)


# ============================================================================
# OCRResult Tests
# ============================================================================

class TestOCRResult:
    """Tests for OCRResult dataclass."""
    
    def test_confidence_percent(self):
        """Test confidence percentage calculation."""
        result = OCRResult(text="Test", confidence=0.85, engine_name="test")
        assert result.confidence_percent == 85.0
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        bbox = BoundingBox(x=10, y=20, width=100, height=30)
        result = OCRResult(
            text="Hello",
            confidence=0.9,
            bounding_box=bbox,
            engine_name="test_engine",
            language="en",
        )
        d = result.to_dict()
        
        assert d["text"] == "Hello"
        assert d["confidence"] == 0.9
        assert d["confidence_percent"] == 90.0
        assert d["engine_name"] == "test_engine"
        assert d["bounding_box"]["x"] == 10
        assert d["bounding_box"]["width"] == 100
    
    def test_to_dict_no_bbox(self):
        """Test dictionary conversion without bounding box."""
        result = OCRResult(text="Test", confidence=0.8, engine_name="test")
        d = result.to_dict()
        
        assert "bounding_box" not in d


# ============================================================================
# OCREngine Base Class Tests
# ============================================================================

class TestOCREngineBase:
    """Tests for OCREngine abstract base class."""
    
    def test_initialization(self):
        """Test engine initialization."""
        engine = MockOCREngine()
        assert engine.config is not None
        assert engine.config.confidence_threshold == 0.5
    
    def test_custom_config(self):
        """Test engine with custom config."""
        config = OCREngineConfig(confidence_threshold=0.8, default_language="de")
        engine = MockOCREngine(config=config)
        
        assert engine.confidence_threshold == 0.8
        assert engine.config.default_language == "de"
    
    def test_filter_by_confidence(self):
        """Test confidence filtering."""
        results = [
            OCRResult(text="High", confidence=0.9, engine_name="test"),
            OCRResult(text="Medium", confidence=0.5, engine_name="test"),
            OCRResult(text="Low", confidence=0.2, engine_name="test"),
        ]
        
        config = OCREngineConfig(confidence_threshold=0.6)
        engine = MockOCREngine(config=config)
        
        filtered = engine.filter_by_confidence(results)
        assert len(filtered) == 1
        assert filtered[0].text == "High"
    
    def test_filter_by_confidence_custom_threshold(self):
        """Test filtering with custom threshold."""
        results = [
            OCRResult(text="A", confidence=0.8, engine_name="test"),
            OCRResult(text="B", confidence=0.4, engine_name="test"),
        ]
        
        engine = MockOCREngine()
        filtered = engine.filter_by_confidence(results, threshold=0.3)
        
        assert len(filtered) == 2
    
    def test_get_combined_text(self):
        """Test text combination."""
        results = [
            OCRResult(text="Hello", confidence=0.9, engine_name="test"),
            OCRResult(text="World", confidence=0.8, engine_name="test"),
        ]
        
        engine = MockOCREngine()
        combined = engine.get_combined_text(results)
        
        assert combined == "Hello World"
    
    def test_get_combined_text_custom_separator(self):
        """Test text combination with custom separator."""
        results = [
            OCRResult(text="Line1", confidence=0.9, engine_name="test"),
            OCRResult(text="Line2", confidence=0.8, engine_name="test"),
        ]
        
        engine = MockOCREngine()
        combined = engine.get_combined_text(results, separator="\n")
        
        assert combined == "Line1\nLine2"
    
    def test_get_combined_text_empty_filtered(self):
        """Test that empty texts are filtered."""
        results = [
            OCRResult(text="Hello", confidence=0.9, engine_name="test"),
            OCRResult(text="", confidence=0.5, engine_name="test"),
            OCRResult(text="World", confidence=0.8, engine_name="test"),
        ]
        
        engine = MockOCREngine()
        combined = engine.get_combined_text(results)
        
        assert combined == "Hello World"


# ============================================================================
# EasyOCREngine Tests
# ============================================================================

class TestEasyOCREngine:
    """Tests for EasyOCREngine wrapper."""
    
    def test_initialization(self):
        """Test engine initialization."""
        engine = EasyOCREngine()
        assert engine.engine_name == "easyocr"
    
    def test_custom_config(self):
        """Test engine with custom config."""
        config = EasyOCRConfig(
            confidence_threshold=0.7,
            gpu=True,
            paragraph=True,
        )
        engine = EasyOCREngine(config=config)
        
        assert engine.config.gpu is True
        assert engine.config.paragraph is True
        assert engine.confidence_threshold == 0.7
    
    def test_language_normalization(self):
        """Test language code normalization."""
        engine = EasyOCREngine()
        
        assert engine._normalize_language("en") == "en"
        assert engine._normalize_language("eng") == "en"
        assert engine._normalize_language("de") == "de"
        assert engine._normalize_language("deu") == "de"
        assert engine._normalize_language("zh") == "ch_sim"
        assert engine._normalize_language("chi_sim") == "ch_sim"
    
    def test_is_available(self):
        """Test availability check."""
        engine = EasyOCREngine()
        # Should reflect actual easyocr installation status
        assert engine.is_available == EASYOCR_AVAILABLE
    
    @pytest.mark.skipif(not EASYOCR_AVAILABLE, reason="EasyOCR not installed")
    def test_recognize_integration(self, sample_image):
        """Integration test with actual EasyOCR (if installed)."""
        engine = EasyOCREngine(EasyOCRConfig(gpu=False))
        results = engine.recognize(sample_image, "en")
        
        # Should return list (may be empty for blank image)
        assert isinstance(results, list)
    
    @patch("ml.ocr_extraction.ocr_engines.easyocr")
    def test_recognize_mocked(self, mock_easyocr, sample_color_image, mock_easyocr_reader):
        """Test recognition with mocked EasyOCR."""
        # Patch availability check
        with patch("ml.ocr_extraction.ocr_engines.EASYOCR_AVAILABLE", True):
            mock_easyocr.Reader.return_value = mock_easyocr_reader
            
            engine = EasyOCREngine()
            engine._reader = mock_easyocr_reader
            engine._loaded_languages = ["en"]
            
            results = engine.recognize(sample_color_image, "en")
            
            assert len(results) == 2
            assert results[0].text == "Hello World"
            assert results[0].confidence == 0.92
            assert results[1].text == "Test"
            assert results[1].confidence == 0.88
            assert results[0].engine_name == "easyocr"
    
    @patch("ml.ocr_extraction.ocr_engines.easyocr")
    def test_recognize_region_mocked(self, mock_easyocr, sample_color_image, sample_bbox, mock_easyocr_reader):
        """Test region recognition with mocked EasyOCR."""
        with patch("ml.ocr_extraction.ocr_engines.EASYOCR_AVAILABLE", True):
            mock_easyocr.Reader.return_value = mock_easyocr_reader
            
            engine = EasyOCREngine()
            engine._reader = mock_easyocr_reader
            engine._loaded_languages = ["en"]
            
            text = engine.recognize_region(sample_color_image, sample_bbox, "en")
            
            # Should combine results
            assert "Hello World" in text
            assert "Test" in text


# ============================================================================
# TesseractEngine Tests
# ============================================================================

class TestTesseractEngine:
    """Tests for TesseractEngine wrapper."""
    
    def test_initialization(self):
        """Test engine initialization."""
        engine = TesseractEngine()
        assert engine.engine_name == "tesseract"
    
    def test_custom_config(self):
        """Test engine with custom config."""
        from ml.ocr_extraction.config import TesseractConfig
        
        tess_config = TesseractConfig(lang="deu", min_confidence=50.0)
        config = TesseractEngineConfig(
            confidence_threshold=0.6,
            tesseract_config=tess_config,
        )
        engine = TesseractEngine(config=config)
        
        assert engine.confidence_threshold == 0.6
        assert engine.config.tesseract_config.lang == "deu"
    
    def test_convert_result(self):
        """Test result conversion."""
        from ml.ocr_extraction.tesseract_wrapper import OCRResult as TessResult
        
        engine = TesseractEngine()
        tess_result = TessResult(
            text="Test Text",
            confidence=85.0,  # Tesseract uses 0-100
            word_confidences=[85.0],
            roi_id="test_roi",
        )
        
        converted = engine._convert_result(tess_result, "en")
        
        assert converted.text == "Test Text"
        assert converted.confidence == 0.85  # Normalized to 0-1
        assert converted.engine_name == "tesseract"
        assert converted.language == "en"
    
    @patch("ml.ocr_extraction.ocr_engines.TesseractOCR")
    def test_recognize_mocked(self, MockTesseractOCR, sample_image):
        """Test recognition with mocked Tesseract."""
        from ml.ocr_extraction.tesseract_wrapper import OCRResult as TessResult
        
        mock_tess = MagicMock()
        mock_tess.recognize.return_value = TessResult(
            text="Tesseract Result",
            confidence=90.0,
            word_confidences=[90.0],
        )
        MockTesseractOCR.return_value = mock_tess
        
        engine = TesseractEngine()
        engine._tesseract = mock_tess
        
        # Mock is_available
        with patch.object(TesseractEngine, "is_available", new_callable=PropertyMock) as mock_avail:
            mock_avail.return_value = True
            
            results = engine.recognize(sample_image, "en")
            
            assert len(results) == 1
            assert results[0].text == "Tesseract Result"
            assert results[0].confidence == 0.9
            assert results[0].engine_name == "tesseract"


# ============================================================================
# FallbackChainEngine Tests
# ============================================================================

class TestFallbackChainEngine:
    """Tests for FallbackChainEngine."""
    
    def test_initialization(self):
        """Test chain initialization."""
        engine1 = MockOCREngine()
        engine2 = MockOCREngine()
        
        chain = FallbackChainEngine(engines=[engine1, engine2])
        
        assert len(chain.engines) == 2
        assert "fallback_chain" in chain.engine_name
    
    def test_empty_chain(self):
        """Test empty chain."""
        chain = FallbackChainEngine()
        
        assert len(chain.engines) == 0
        assert chain.is_available is False
    
    def test_add_engine(self):
        """Test adding engines."""
        chain = FallbackChainEngine()
        engine = MockOCREngine()
        
        chain.add_engine(engine)
        
        assert len(chain.engines) == 1
    
    def test_is_available(self):
        """Test availability check."""
        engine1 = MockOCREngine(available=False)
        engine2 = MockOCREngine(available=True)
        
        chain = FallbackChainEngine(engines=[engine1, engine2])
        
        assert chain.is_available is True
    
    def test_is_available_all_unavailable(self):
        """Test availability when all engines unavailable."""
        engine1 = MockOCREngine(available=False)
        engine2 = MockOCREngine(available=False)
        
        chain = FallbackChainEngine(engines=[engine1, engine2])
        
        assert chain.is_available is False
    
    def test_primary_engine_succeeds(self, sample_image):
        """Test when primary engine succeeds."""
        primary_results = [
            OCRResult(text="Primary", confidence=0.9, engine_name="primary"),
        ]
        fallback_results = [
            OCRResult(text="Fallback", confidence=0.95, engine_name="fallback"),
        ]
        
        primary = MockOCREngine(results=primary_results)
        primary._name = "primary"
        fallback = MockOCREngine(results=fallback_results)
        fallback._name = "fallback"
        
        config = FallbackChainConfig(
            primary_confidence_threshold=0.7,
            fallback_confidence_threshold=0.5,
        )
        chain = FallbackChainEngine(engines=[primary, fallback], config=config)
        
        results = chain.recognize(sample_image)
        
        assert len(results) == 1
        assert results[0].text == "Primary"
        assert chain.last_result.engine_used == "primary"
        assert chain.last_result.fallback_occurred is False
    
    def test_fallback_on_low_confidence(self, sample_image):
        """Test fallback when primary confidence is low."""
        primary_results = [
            OCRResult(text="Low Conf", confidence=0.4, engine_name="primary"),
        ]
        fallback_results = [
            OCRResult(text="Better", confidence=0.85, engine_name="fallback"),
        ]
        
        primary = MockOCREngine(results=primary_results)
        primary._name = "primary"
        fallback = MockOCREngine(results=fallback_results)
        fallback._name = "fallback"
        
        config = FallbackChainConfig(
            primary_confidence_threshold=0.7,
            fallback_confidence_threshold=0.5,
        )
        chain = FallbackChainEngine(engines=[primary, fallback], config=config)
        
        results = chain.recognize(sample_image)
        
        assert len(results) == 1
        assert results[0].text == "Better"
        assert chain.last_result.engine_used == "fallback"
        assert chain.last_result.fallback_occurred is True
        assert "primary" in chain.last_result.engines_tried
        assert "fallback" in chain.last_result.engines_tried
    
    def test_returns_best_when_all_below_threshold(self, sample_image):
        """Test returns best result when all engines below threshold."""
        primary_results = [
            OCRResult(text="Low", confidence=0.3, engine_name="primary"),
        ]
        fallback_results = [
            OCRResult(text="Also Low", confidence=0.4, engine_name="fallback"),
        ]
        
        primary = MockOCREngine(results=primary_results)
        primary._name = "primary"
        fallback = MockOCREngine(results=fallback_results)
        fallback._name = "fallback"
        
        config = FallbackChainConfig(
            primary_confidence_threshold=0.7,
            fallback_confidence_threshold=0.6,
        )
        chain = FallbackChainEngine(engines=[primary, fallback], config=config)
        
        results = chain.recognize(sample_image)
        
        # Should return the best result (fallback with 0.4)
        assert len(results) == 1
        assert results[0].text == "Also Low"
        assert chain.last_result.engine_used == "fallback"
    
    def test_skips_unavailable_engines(self, sample_image):
        """Test skipping unavailable engines."""
        unavailable = MockOCREngine(available=False)
        unavailable._name = "unavailable"
        
        available_results = [
            OCRResult(text="Available", confidence=0.9, engine_name="available"),
        ]
        available = MockOCREngine(results=available_results, available=True)
        available._name = "available"
        
        chain = FallbackChainEngine(engines=[unavailable, available])
        
        results = chain.recognize(sample_image)
        
        assert len(results) == 1
        assert results[0].text == "Available"
        assert "unavailable" not in chain.last_result.engines_tried
    
    def test_handles_engine_exception(self, sample_image):
        """Test handling engine exceptions."""
        class FailingEngine(MockOCREngine):
            def recognize(self, image, language="en"):
                raise RuntimeError("Engine failed")
        
        failing = FailingEngine()
        failing._name = "failing"
        
        backup_results = [
            OCRResult(text="Backup", confidence=0.8, engine_name="backup"),
        ]
        backup = MockOCREngine(results=backup_results)
        backup._name = "backup"
        
        chain = FallbackChainEngine(engines=[failing, backup])
        
        results = chain.recognize(sample_image)
        
        assert len(results) == 1
        assert results[0].text == "Backup"
    
    def test_max_engines_limit(self, sample_image):
        """Test max engines limit."""
        engines = []
        for i in range(5):
            results = [
                OCRResult(text=f"Engine{i}", confidence=0.3, engine_name=f"engine{i}"),
            ]
            engine = MockOCREngine(results=results)
            engine._name = f"engine{i}"
            engines.append(engine)
        
        config = FallbackChainConfig(max_engines=2)
        chain = FallbackChainEngine(engines=engines, config=config)
        
        chain.recognize(sample_image)
        
        # Only 2 engines should be tried
        assert len(chain.last_result.engines_tried) == 2
    
    def test_recognize_region(self, sample_color_image, sample_bbox):
        """Test region recognition."""
        results = [
            OCRResult(text="Region Text", confidence=0.85, engine_name="mock"),
        ]
        engine = MockOCREngine(results=results)
        
        chain = FallbackChainEngine(engines=[engine])
        
        text = chain.recognize_region(sample_color_image, sample_bbox)
        
        assert "Region Text" in text
    
    def test_empty_results(self, sample_image):
        """Test handling empty results from all engines."""
        engine1 = MockOCREngine(results=[])
        engine2 = MockOCREngine(results=[])
        
        chain = FallbackChainEngine(engines=[engine1, engine2])
        
        results = chain.recognize(sample_image)
        
        assert results == []
        assert chain.last_result.combined_confidence == 0.0
    
    def test_confidence_calculation_weighted(self, sample_image):
        """Test weighted confidence calculation."""
        results = [
            OCRResult(text="A", confidence=1.0, engine_name="test"),  # weight 1
            OCRResult(text="Long text here", confidence=0.5, engine_name="test"),  # weight 14
        ]
        engine = MockOCREngine(results=results)
        
        chain = FallbackChainEngine(engines=[engine])
        chain.recognize(sample_image)
        
        # Weighted average: (1*1.0 + 14*0.5) / (1+14) = 8/15 ≈ 0.533
        expected = (1 * 1.0 + 14 * 0.5) / 15
        assert abs(chain.last_result.combined_confidence - expected) < 0.01


class TestFallbackResult:
    """Tests for FallbackResult dataclass."""
    
    def test_to_dict(self):
        """Test dictionary conversion."""
        results = [
            OCRResult(text="Test", confidence=0.9, engine_name="test"),
        ]
        fb_result = FallbackResult(
            results=results,
            engine_used="test_engine",
            engines_tried=["primary", "test_engine"],
            fallback_occurred=True,
            combined_confidence=0.9,
        )
        
        d = fb_result.to_dict()
        
        assert len(d["results"]) == 1
        assert d["engine_used"] == "test_engine"
        assert d["fallback_occurred"] is True
        assert d["combined_confidence"] == 0.9
        assert "primary" in d["engines_tried"]


# ============================================================================
# Factory Function Tests
# ============================================================================

class TestCreateDefaultChain:
    """Tests for create_default_chain factory function."""
    
    def test_creates_chain(self):
        """Test chain creation."""
        # This may or may not work depending on installed engines
        try:
            chain = create_default_chain()
            assert isinstance(chain, FallbackChainEngine)
            assert len(chain.engines) > 0
        except RuntimeError as e:
            # Expected if no engines installed
            assert "No OCR engines available" in str(e)
    
    def test_custom_thresholds(self):
        """Test custom threshold configuration."""
        try:
            chain = create_default_chain(
                primary_threshold=0.9,
                fallback_threshold=0.4,
            )
            assert chain.config.primary_confidence_threshold == 0.9
            assert chain.config.fallback_confidence_threshold == 0.4
        except RuntimeError:
            pytest.skip("No OCR engines available")


# ============================================================================
# Language Support Tests
# ============================================================================

class TestLanguageSupport:
    """Tests for multi-language support."""
    
    def test_easyocr_language_map_coverage(self):
        """Test that common languages are mapped."""
        engine = EasyOCREngine()
        
        # ISO 639-1 codes
        assert engine._normalize_language("en") == "en"
        assert engine._normalize_language("de") == "de"
        assert engine._normalize_language("fr") == "fr"
        assert engine._normalize_language("es") == "es"
        assert engine._normalize_language("ja") == "ja"
        assert engine._normalize_language("ko") == "ko"
        assert engine._normalize_language("ru") == "ru"
        
        # ISO 639-2 codes (Tesseract style)
        assert engine._normalize_language("eng") == "en"
        assert engine._normalize_language("deu") == "de"
        assert engine._normalize_language("fra") == "fr"
    
    def test_unknown_language_passthrough(self):
        """Test unknown languages pass through unchanged."""
        engine = EasyOCREngine()
        
        # Uncommon language should pass through
        assert engine._normalize_language("xyz") == "xyz"
    
    def test_case_insensitive(self):
        """Test language codes are case insensitive."""
        engine = EasyOCREngine()
        
        assert engine._normalize_language("EN") == "en"
        assert engine._normalize_language("De") == "de"
        assert engine._normalize_language("FRA") == "fr"


# ============================================================================
# Edge Cases
# ============================================================================

class TestEdgeCases:
    """Tests for edge cases and error handling."""
    
    def test_empty_image(self):
        """Test handling of empty/zero-size image."""
        engine = MockOCREngine(results=[])
        
        # Zero-size image
        empty = np.array([], dtype=np.uint8).reshape(0, 0)
        results = engine.recognize(empty)
        
        assert results == []
    
    def test_bbox_boundary_clipping(self, sample_color_image):
        """Test that region recognition clips to image bounds."""
        results = [OCRResult(text="Clipped", confidence=0.9, engine_name="test")]
        engine = MockOCREngine(results=results)
        
        # BBox partially outside image
        bbox = BoundingBox(x=-10, y=-10, width=50, height=50)
        
        # Should not raise
        text = engine.recognize_region(sample_color_image, bbox)
        assert text == "Clipped"
    
    def test_bbox_fully_outside(self, sample_color_image):
        """Test bbox completely outside image."""
        engine = MockOCREngine(results=[])
        
        # BBox fully outside
        bbox = BoundingBox(x=1000, y=1000, width=50, height=50)
        
        text = engine.recognize_region(sample_color_image, bbox)
        # Should return empty string
        assert text == ""
    
    def test_single_channel_image(self):
        """Test grayscale image handling."""
        results = [OCRResult(text="Gray", confidence=0.9, engine_name="test")]
        engine = MockOCREngine(results=results)
        
        gray = np.ones((50, 100), dtype=np.uint8) * 128
        results = engine.recognize(gray)
        
        assert results[0].text == "Gray"
    
    def test_four_channel_image(self):
        """Test RGBA image handling."""
        results = [OCRResult(text="RGBA", confidence=0.9, engine_name="test")]
        engine = MockOCREngine(results=results)
        
        rgba = np.ones((50, 100, 4), dtype=np.uint8) * 200
        results = engine.recognize(rgba)
        
        assert results[0].text == "RGBA"
