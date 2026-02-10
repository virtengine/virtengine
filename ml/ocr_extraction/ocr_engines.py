"""
OCR Engine abstraction layer with fallback chain support.

This module provides:
- Abstract OCREngine base class for consistent interface
- EasyOCREngine wrapper for deep learning-based OCR
- TesseractEngine wrapper using existing TesseractOCR
- FallbackChainEngine that tries engines in sequence

The fallback chain is useful for production environments where
multiple OCR engines can improve extraction accuracy through
redundancy and confidence-based fallback.

VE-3043: Integrate EasyOCR with Fallback Chain
"""

import logging
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any, Callable
from enum import Enum

import numpy as np

from ml.ocr_extraction.tesseract_wrapper import (
    TesseractOCR,
    OCRResult as TesseractOCRResult,
)
from ml.ocr_extraction.config import TesseractConfig

# Try importing EasyOCR
try:
    import easyocr
    EASYOCR_AVAILABLE = True
except ImportError:
    EASYOCR_AVAILABLE = False
    easyocr = None


logger = logging.getLogger(__name__)


class OCREngineType(str, Enum):
    """Supported OCR engine types."""
    EASYOCR = "easyocr"
    TESSERACT = "tesseract"
    FALLBACK_CHAIN = "fallback_chain"


@dataclass
class BoundingBox:
    """Bounding box for text region."""
    x: int
    y: int
    width: int
    height: int
    
    def to_tuple(self) -> Tuple[int, int, int, int]:
        """Return as (x, y, width, height) tuple."""
        return (self.x, self.y, self.width, self.height)
    
    def to_xyxy(self) -> Tuple[int, int, int, int]:
        """Return as (x1, y1, x2, y2) tuple."""
        return (self.x, self.y, self.x + self.width, self.y + self.height)
    
    @classmethod
    def from_xyxy(cls, x1: int, y1: int, x2: int, y2: int) -> "BoundingBox":
        """Create from (x1, y1, x2, y2) coordinates."""
        return cls(x=x1, y=y1, width=x2 - x1, height=y2 - y1)
    
    @classmethod
    def from_points(cls, points: List[List[int]]) -> "BoundingBox":
        """Create from list of corner points [[x1,y1], [x2,y2], ...]."""
        if len(points) < 4:
            raise ValueError("Need at least 4 points for bounding box")
        
        xs = [p[0] for p in points]
        ys = [p[1] for p in points]
        x_min, x_max = min(xs), max(xs)
        y_min, y_max = min(ys), max(ys)
        
        return cls(x=x_min, y=y_min, width=x_max - x_min, height=y_max - y_min)


@dataclass
class OCRResult:
    """
    Unified OCR result structure.
    
    Provides a consistent interface for OCR results regardless
    of which engine produced them.
    """
    text: str
    confidence: float  # 0.0 - 1.0 normalized
    bounding_box: Optional[BoundingBox] = None
    engine_name: str = ""
    language: str = "en"
    raw_result: Any = None  # Engine-specific raw result
    
    @property
    def confidence_percent(self) -> float:
        """Get confidence as percentage (0-100)."""
        return self.confidence * 100.0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        result = {
            "text": self.text,
            "confidence": self.confidence,
            "confidence_percent": self.confidence_percent,
            "engine_name": self.engine_name,
            "language": self.language,
        }
        if self.bounding_box:
            result["bounding_box"] = {
                "x": self.bounding_box.x,
                "y": self.bounding_box.y,
                "width": self.bounding_box.width,
                "height": self.bounding_box.height,
            }
        return result


@dataclass 
class OCREngineConfig:
    """Base configuration for OCR engines."""
    confidence_threshold: float = 0.5  # 0.0 - 1.0
    default_language: str = "en"
    gpu: bool = False
    

@dataclass
class EasyOCRConfig(OCREngineConfig):
    """Configuration specific to EasyOCR."""
    model_storage_directory: Optional[str] = None
    download_enabled: bool = True
    detector: bool = True  # Enable text detection
    recognizer: bool = True  # Enable text recognition
    paragraph: bool = False  # Merge into paragraphs
    detail: int = 1  # 0 = simple output, 1 = detailed
    batch_size: int = 1  # For GPU processing
    
    
@dataclass
class TesseractEngineConfig(OCREngineConfig):
    """Configuration specific to Tesseract engine wrapper."""
    tesseract_config: Optional[TesseractConfig] = None


@dataclass
class FallbackChainConfig(OCREngineConfig):
    """Configuration for fallback chain."""
    # Confidence thresholds for fallback decisions
    primary_confidence_threshold: float = 0.7
    fallback_confidence_threshold: float = 0.3
    # Whether to combine results from multiple engines
    combine_results: bool = False
    # Maximum number of engines to try
    max_engines: int = 3


class OCREngine(ABC):
    """
    Abstract base class for OCR engines.
    
    Provides a consistent interface for different OCR backends
    including EasyOCR, Tesseract, and others.
    
    Implementations should:
    - Handle initialization and model loading
    - Normalize confidence scores to 0.0-1.0 range
    - Return consistent OCRResult structures
    """
    
    def __init__(self, config: Optional[OCREngineConfig] = None):
        """
        Initialize the OCR engine.
        
        Args:
            config: Engine configuration. Uses defaults if None.
        """
        self.config = config or OCREngineConfig()
        self._initialized = False
    
    @property
    @abstractmethod
    def engine_name(self) -> str:
        """Return the engine identifier name."""
        pass
    
    @property
    @abstractmethod
    def is_available(self) -> bool:
        """Check if the engine backend is available."""
        pass
    
    @abstractmethod
    def recognize(
        self,
        image: np.ndarray,
        language: str = "en"
    ) -> List[OCRResult]:
        """
        Recognize text in an entire image.
        
        Args:
            image: Input image as numpy array (BGR or grayscale)
            language: Language code for recognition (e.g., "en", "de")
            
        Returns:
            List of OCRResult objects for each detected text region
        """
        pass
    
    @abstractmethod
    def recognize_region(
        self,
        image: np.ndarray,
        bbox: BoundingBox,
        language: str = "en"
    ) -> str:
        """
        Recognize text in a specific region of the image.
        
        Args:
            image: Input image as numpy array (BGR or grayscale)
            bbox: Bounding box defining the region to recognize
            language: Language code for recognition
            
        Returns:
            Recognized text string from the specified region
        """
        pass
    
    @property
    def confidence_threshold(self) -> float:
        """
        Get the minimum confidence threshold for accepting results.
        
        Returns:
            Confidence threshold as float between 0.0 and 1.0
        """
        return self.config.confidence_threshold
    
    def filter_by_confidence(
        self,
        results: List[OCRResult],
        threshold: Optional[float] = None
    ) -> List[OCRResult]:
        """
        Filter results by confidence threshold.
        
        Args:
            results: List of OCR results to filter
            threshold: Override threshold (uses config if None)
            
        Returns:
            Filtered list of results meeting threshold
        """
        min_conf = threshold if threshold is not None else self.confidence_threshold
        return [r for r in results if r.confidence >= min_conf]
    
    def get_combined_text(
        self,
        results: List[OCRResult],
        separator: str = " "
    ) -> str:
        """
        Combine all result texts into single string.
        
        Args:
            results: List of OCR results
            separator: String to join results with
            
        Returns:
            Combined text string
        """
        return separator.join(r.text for r in results if r.text)


class EasyOCREngine(OCREngine):
    """
    EasyOCR engine wrapper.
    
    EasyOCR is a deep learning-based OCR package that supports
    80+ languages and provides both text detection and recognition.
    
    Features:
    - GPU acceleration support
    - Multiple language support
    - Built-in text detection
    - Character-level confidence
    """
    
    LANGUAGE_MAP = {
        "en": "en",
        "eng": "en",
        "de": "de",
        "deu": "de", 
        "fr": "fr",
        "fra": "fr",
        "es": "es",
        "spa": "es",
        "it": "it",
        "ita": "it",
        "pt": "pt",
        "por": "pt",
        "nl": "nl",
        "nld": "nl",
        "ja": "ja",
        "jpn": "ja",
        "ko": "ko",
        "kor": "ko",
        "zh": "ch_sim",
        "chi_sim": "ch_sim",
        "chi_tra": "ch_tra",
        "ar": "ar",
        "ara": "ar",
        "ru": "ru",
        "rus": "ru",
    }
    
    def __init__(self, config: Optional[EasyOCRConfig] = None):
        """
        Initialize EasyOCR engine.
        
        Args:
            config: EasyOCR configuration
            
        Raises:
            ImportError: If easyocr is not installed
        """
        super().__init__(config or EasyOCRConfig())
        self._reader: Optional[Any] = None
        self._loaded_languages: List[str] = []
        
    @property
    def config(self) -> EasyOCRConfig:
        """Get typed config."""
        return self._config
    
    @config.setter
    def config(self, value: OCREngineConfig) -> None:
        """Set config with type conversion."""
        if isinstance(value, EasyOCRConfig):
            self._config = value
        else:
            self._config = EasyOCRConfig(
                confidence_threshold=value.confidence_threshold,
                default_language=value.default_language,
                gpu=value.gpu,
            )
    
    @property
    def engine_name(self) -> str:
        """Return engine identifier."""
        return "easyocr"
    
    @property
    def is_available(self) -> bool:
        """Check if EasyOCR is available."""
        return EASYOCR_AVAILABLE
    
    def _normalize_language(self, language: str) -> str:
        """Convert language code to EasyOCR format."""
        return self.LANGUAGE_MAP.get(language.lower(), language)
    
    def _get_reader(self, language: str) -> Any:
        """
        Get or create EasyOCR reader for language.
        
        Args:
            language: Language code
            
        Returns:
            EasyOCR Reader instance
        """
        if not self.is_available:
            raise ImportError(
                "EasyOCR is not installed. Install with: pip install easyocr"
            )
        
        easyocr_lang = self._normalize_language(language)
        
        # Reuse reader if language matches
        if self._reader is not None and easyocr_lang in self._loaded_languages:
            return self._reader
        
        # Create new reader
        logger.info(f"Initializing EasyOCR reader for language: {easyocr_lang}")
        
        reader_kwargs = {
            "lang_list": [easyocr_lang],
            "gpu": self.config.gpu,
            "download_enabled": self.config.download_enabled,
        }
        
        if self.config.model_storage_directory:
            reader_kwargs["model_storage_directory"] = self.config.model_storage_directory
        
        self._reader = easyocr.Reader(**reader_kwargs)
        self._loaded_languages = [easyocr_lang]
        self._initialized = True
        
        return self._reader
    
    def recognize(
        self,
        image: np.ndarray,
        language: str = "en"
    ) -> List[OCRResult]:
        """
        Recognize text in image using EasyOCR.
        
        Args:
            image: Input image (BGR or grayscale numpy array)
            language: Language code for recognition
            
        Returns:
            List of OCRResult for each detected text region
        """
        reader = self._get_reader(language)
        
        # EasyOCR expects RGB, convert if needed
        if len(image.shape) == 3 and image.shape[2] == 3:
            # Assume BGR input (OpenCV convention)
            import cv2
            image_rgb = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
        else:
            image_rgb = image
        
        # Run OCR
        raw_results = reader.readtext(
            image_rgb,
            detail=self.config.detail,
            paragraph=self.config.paragraph,
            batch_size=self.config.batch_size,
        )
        
        results = []
        for raw in raw_results:
            if self.config.detail == 1:
                # Detail format: (bbox_points, text, confidence)
                bbox_points, text, confidence = raw
                bbox = BoundingBox.from_points(bbox_points)
            else:
                # Simple format: just text
                text = raw
                confidence = 1.0
                bbox = None
            
            results.append(OCRResult(
                text=text,
                confidence=float(confidence),
                bounding_box=bbox,
                engine_name=self.engine_name,
                language=language,
                raw_result=raw,
            ))
        
        return results
    
    def recognize_region(
        self,
        image: np.ndarray,
        bbox: BoundingBox,
        language: str = "en"
    ) -> str:
        """
        Recognize text in a specific image region.
        
        Args:
            image: Input image
            bbox: Region bounding box
            language: Language code
            
        Returns:
            Recognized text from the region
        """
        # Crop the region
        x1, y1, x2, y2 = bbox.to_xyxy()
        y1 = max(0, y1)
        x1 = max(0, x1)
        y2 = min(image.shape[0], y2)
        x2 = min(image.shape[1], x2)
        
        cropped = image[y1:y2, x1:x2]
        
        if cropped.size == 0:
            return ""
        
        results = self.recognize(cropped, language)
        return self.get_combined_text(results)


class TesseractEngine(OCREngine):
    """
    Tesseract OCR engine wrapper.
    
    Wraps the existing TesseractOCR class to conform to
    the OCREngine interface for use in fallback chains.
    """
    
    def __init__(self, config: Optional[TesseractEngineConfig] = None):
        """
        Initialize Tesseract engine.
        
        Args:
            config: Tesseract engine configuration
        """
        super().__init__(config or TesseractEngineConfig())
        self._tesseract: Optional[TesseractOCR] = None
    
    @property
    def config(self) -> TesseractEngineConfig:
        """Get typed config."""
        return self._config
    
    @config.setter
    def config(self, value: OCREngineConfig) -> None:
        """Set config with type conversion."""
        if isinstance(value, TesseractEngineConfig):
            self._config = value
        else:
            self._config = TesseractEngineConfig(
                confidence_threshold=value.confidence_threshold,
                default_language=value.default_language,
                gpu=value.gpu,
            )
    
    @property
    def engine_name(self) -> str:
        """Return engine identifier."""
        return "tesseract"
    
    @property
    def is_available(self) -> bool:
        """Check if Tesseract is available."""
        try:
            from ml.ocr_extraction.tesseract_wrapper import TESSERACT_AVAILABLE
            if not TESSERACT_AVAILABLE:
                return False
            # Also check if binary is installed
            import pytesseract
            pytesseract.get_tesseract_version()
            return True
        except Exception:
            return False
    
    def _get_tesseract(self) -> TesseractOCR:
        """Get or create TesseractOCR instance."""
        if self._tesseract is None:
            tesseract_config = self.config.tesseract_config or TesseractConfig()
            self._tesseract = TesseractOCR(tesseract_config)
            self._initialized = True
        return self._tesseract
    
    def _convert_result(
        self,
        tess_result: TesseractOCRResult,
        language: str
    ) -> OCRResult:
        """Convert TesseractOCRResult to unified OCRResult."""
        # Tesseract confidence is 0-100, normalize to 0-1
        normalized_confidence = tess_result.confidence / 100.0
        
        return OCRResult(
            text=tess_result.text,
            confidence=normalized_confidence,
            bounding_box=None,  # Tesseract wrapper doesn't include bbox
            engine_name=self.engine_name,
            language=language,
            raw_result=tess_result,
        )
    
    def recognize(
        self,
        image: np.ndarray,
        language: str = "en"
    ) -> List[OCRResult]:
        """
        Recognize text using Tesseract.
        
        Args:
            image: Input image
            language: Language code
            
        Returns:
            List with single OCRResult containing all recognized text
        """
        tesseract = self._get_tesseract()
        
        # Update language in config if different
        if language != "en":
            lang_map = {"en": "eng", "de": "deu", "fr": "fra", "es": "spa"}
            tess_lang = lang_map.get(language, language)
            config = TesseractConfig(
                lang=tess_lang,
                **{k: v for k, v in self.config.tesseract_config.__dict__.items() 
                   if k != "lang"} if self.config.tesseract_config else {}
            )
            tess_result = tesseract.recognize(image, config_override=config)
        else:
            tess_result = tesseract.recognize(image)
        
        return [self._convert_result(tess_result, language)]
    
    def recognize_region(
        self,
        image: np.ndarray,
        bbox: BoundingBox,
        language: str = "en"
    ) -> str:
        """
        Recognize text in a specific image region.
        
        Args:
            image: Input image
            bbox: Region bounding box
            language: Language code
            
        Returns:
            Recognized text from the region
        """
        # Crop the region
        x1, y1, x2, y2 = bbox.to_xyxy()
        y1 = max(0, y1)
        x1 = max(0, x1)
        y2 = min(image.shape[0], y2)
        x2 = min(image.shape[1], x2)
        
        cropped = image[y1:y2, x1:x2]
        
        if cropped.size == 0:
            return ""
        
        results = self.recognize(cropped, language)
        return self.get_combined_text(results)


@dataclass
class FallbackResult:
    """Result from fallback chain execution."""
    results: List[OCRResult]
    engine_used: str
    engines_tried: List[str]
    fallback_occurred: bool
    combined_confidence: float
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "results": [r.to_dict() for r in self.results],
            "engine_used": self.engine_used,
            "engines_tried": self.engines_tried,
            "fallback_occurred": self.fallback_occurred,
            "combined_confidence": self.combined_confidence,
        }


class FallbackChainEngine(OCREngine):
    """
    OCR engine that chains multiple engines with fallback logic.
    
    Tries engines in sequence until one returns results above
    the confidence threshold. Useful for production environments
    where accuracy is critical.
    
    Features:
    - Configurable confidence thresholds
    - Engine priority ordering
    - Logging of which engine succeeded
    - Optional result combining
    
    Example:
        primary = EasyOCREngine(EasyOCRConfig(gpu=True))
        fallback = TesseractEngine()
        
        chain = FallbackChainEngine(
            engines=[primary, fallback],
            config=FallbackChainConfig(primary_confidence_threshold=0.7)
        )
        
        results = chain.recognize(image)
    """
    
    def __init__(
        self,
        engines: Optional[List[OCREngine]] = None,
        config: Optional[FallbackChainConfig] = None
    ):
        """
        Initialize fallback chain.
        
        Args:
            engines: Ordered list of engines to try (first = primary)
            config: Fallback chain configuration
        """
        super().__init__(config or FallbackChainConfig())
        self._engines = engines or []
        self._last_result: Optional[FallbackResult] = None
        
    @property
    def config(self) -> FallbackChainConfig:
        """Get typed config."""
        return self._config
    
    @config.setter
    def config(self, value: OCREngineConfig) -> None:
        """Set config with type conversion."""
        if isinstance(value, FallbackChainConfig):
            self._config = value
        else:
            self._config = FallbackChainConfig(
                confidence_threshold=value.confidence_threshold,
                default_language=value.default_language,
                gpu=value.gpu,
            )
    
    @property
    def engine_name(self) -> str:
        """Return engine identifier."""
        engine_names = [e.engine_name for e in self._engines]
        return f"fallback_chain[{','.join(engine_names)}]"
    
    @property
    def is_available(self) -> bool:
        """Check if at least one engine is available."""
        return any(e.is_available for e in self._engines)
    
    @property
    def engines(self) -> List[OCREngine]:
        """Get the list of engines."""
        return self._engines
    
    @property
    def last_result(self) -> Optional[FallbackResult]:
        """Get the last fallback result with metadata."""
        return self._last_result
    
    def add_engine(self, engine: OCREngine) -> None:
        """
        Add an engine to the chain.
        
        Args:
            engine: OCR engine to add
        """
        self._engines.append(engine)
    
    def _calculate_combined_confidence(
        self,
        results: List[OCRResult]
    ) -> float:
        """Calculate combined confidence from multiple results."""
        if not results:
            return 0.0
        
        # Weight by text length
        total_weight = 0
        weighted_sum = 0.0
        
        for r in results:
            weight = max(1, len(r.text))
            weighted_sum += r.confidence * weight
            total_weight += weight
        
        return weighted_sum / total_weight if total_weight > 0 else 0.0
    
    def recognize(
        self,
        image: np.ndarray,
        language: str = "en"
    ) -> List[OCRResult]:
        """
        Recognize text using fallback chain.
        
        Tries each engine in sequence until one returns
        results above the confidence threshold.
        
        Args:
            image: Input image
            language: Language code
            
        Returns:
            List of OCRResult from the first successful engine
        """
        engines_tried = []
        fallback_occurred = False
        best_results: List[OCRResult] = []
        best_confidence = 0.0
        best_engine = ""
        
        for i, engine in enumerate(self._engines[:self.config.max_engines]):
            if not engine.is_available:
                logger.debug(f"Skipping unavailable engine: {engine.engine_name}")
                continue
            
            engines_tried.append(engine.engine_name)
            
            try:
                results = engine.recognize(image, language)
                confidence = self._calculate_combined_confidence(results)
                
                logger.debug(
                    f"Engine {engine.engine_name} returned confidence: {confidence:.3f}"
                )
                
                # Determine threshold based on position in chain
                if i == 0:
                    threshold = self.config.primary_confidence_threshold
                else:
                    threshold = self.config.fallback_confidence_threshold
                    fallback_occurred = True
                
                # Accept if above threshold
                if confidence >= threshold:
                    logger.info(
                        f"OCR succeeded with {engine.engine_name} "
                        f"(confidence: {confidence:.3f})"
                    )
                    
                    self._last_result = FallbackResult(
                        results=results,
                        engine_used=engine.engine_name,
                        engines_tried=engines_tried,
                        fallback_occurred=fallback_occurred,
                        combined_confidence=confidence,
                    )
                    return results
                
                # Track best result in case all fail threshold
                if confidence > best_confidence:
                    best_results = results
                    best_confidence = confidence
                    best_engine = engine.engine_name
                    
            except Exception as e:
                logger.warning(
                    f"Engine {engine.engine_name} failed: {e}"
                )
                continue
        
        # Return best result even if below threshold
        if best_results:
            logger.warning(
                f"No engine met threshold, using best result from {best_engine} "
                f"(confidence: {best_confidence:.3f})"
            )
            
            self._last_result = FallbackResult(
                results=best_results,
                engine_used=best_engine,
                engines_tried=engines_tried,
                fallback_occurred=len(engines_tried) > 1,
                combined_confidence=best_confidence,
            )
            return best_results
        
        # All engines failed
        logger.error("All OCR engines failed")
        self._last_result = FallbackResult(
            results=[],
            engine_used="",
            engines_tried=engines_tried,
            fallback_occurred=len(engines_tried) > 1,
            combined_confidence=0.0,
        )
        return []
    
    def recognize_region(
        self,
        image: np.ndarray,
        bbox: BoundingBox,
        language: str = "en"
    ) -> str:
        """
        Recognize text in a specific region using fallback chain.
        
        Args:
            image: Input image
            bbox: Region bounding box
            language: Language code
            
        Returns:
            Recognized text from the region
        """
        # Crop and delegate to recognize
        x1, y1, x2, y2 = bbox.to_xyxy()
        y1 = max(0, y1)
        x1 = max(0, x1)
        y2 = min(image.shape[0], y2)
        x2 = min(image.shape[1], x2)
        
        cropped = image[y1:y2, x1:x2]
        
        if cropped.size == 0:
            return ""
        
        results = self.recognize(cropped, language)
        return self.get_combined_text(results)


def create_default_chain(
    gpu: bool = False,
    primary_threshold: float = 0.7,
    fallback_threshold: float = 0.3,
) -> FallbackChainEngine:
    """
    Create a default fallback chain with EasyOCR primary and Tesseract fallback.
    
    Args:
        gpu: Enable GPU for EasyOCR
        primary_threshold: Confidence threshold for primary engine
        fallback_threshold: Confidence threshold for fallback engines
        
    Returns:
        Configured FallbackChainEngine
    """
    engines: List[OCREngine] = []
    
    # Add EasyOCR if available
    if EASYOCR_AVAILABLE:
        engines.append(EasyOCREngine(EasyOCRConfig(gpu=gpu)))
    
    # Add Tesseract
    try:
        tess_engine = TesseractEngine()
        if tess_engine.is_available:
            engines.append(tess_engine)
    except Exception:
        pass
    
    if not engines:
        raise RuntimeError(
            "No OCR engines available. Install easyocr or tesseract."
        )
    
    return FallbackChainEngine(
        engines=engines,
        config=FallbackChainConfig(
            primary_confidence_threshold=primary_threshold,
            fallback_confidence_threshold=fallback_threshold,
        ),
    )
