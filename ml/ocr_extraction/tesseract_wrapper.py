"""
Tesseract OCR wrapper for text recognition.

This module provides a high-level interface to Tesseract OCR
with support for configurable character whitelists, page
segmentation modes, and detailed confidence information.
"""

import cv2
import numpy as np
from typing import List, Optional, Dict, Any, Tuple
from dataclasses import dataclass, field
import re

try:
    import pytesseract
    from pytesseract import Output
    TESSERACT_AVAILABLE = True
except ImportError:
    TESSERACT_AVAILABLE = False
    Output = None

from ml.ocr_extraction.config import TesseractConfig, PageSegmentationMode


@dataclass
class CharacterDetail:
    """Details for a single recognized character."""
    char: str
    confidence: float  # 0.0 - 100.0
    x: int
    y: int
    width: int
    height: int
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "char": self.char,
            "confidence": self.confidence,
            "x": self.x,
            "y": self.y,
            "width": self.width,
            "height": self.height,
        }


@dataclass
class WordDetail:
    """Details for a recognized word."""
    text: str
    confidence: float  # 0.0 - 100.0
    x: int
    y: int
    width: int
    height: int
    characters: List[CharacterDetail] = field(default_factory=list)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "text": self.text,
            "confidence": self.confidence,
            "x": self.x,
            "y": self.y,
            "width": self.width,
            "height": self.height,
            "characters": [c.to_dict() for c in self.characters],
        }


@dataclass
class OCRResult:
    """Basic OCR result with text and confidence."""
    text: str
    confidence: float  # Average confidence (0.0 - 100.0)
    word_confidences: List[float]  # Per-word confidences
    roi_id: Optional[str] = None
    
    @property
    def normalized_confidence(self) -> float:
        """Get confidence normalized to 0.0 - 1.0 range."""
        return self.confidence / 100.0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "text": self.text,
            "confidence": self.confidence,
            "word_confidences": self.word_confidences,
            "roi_id": self.roi_id,
        }


@dataclass
class DetailedOCRResult:
    """
    Detailed OCR result with character-level information.
    
    Contains per-character confidence scores and bounding boxes
    for quality assessment and error localization.
    """
    text: str
    confidence: float  # Average confidence
    words: List[WordDetail]
    raw_data: Dict[str, Any]  # Raw pytesseract output
    roi_id: Optional[str] = None
    
    @property
    def word_confidences(self) -> List[float]:
        """Get list of word confidences."""
        return [w.confidence for w in self.words]
    
    @property
    def character_count(self) -> int:
        """Get total character count."""
        return sum(len(w.characters) for w in self.words)
    
    @property
    def low_confidence_chars(self) -> List[CharacterDetail]:
        """Get characters with confidence below 50%."""
        result = []
        for word in self.words:
            for char in word.characters:
                if char.confidence < 50:
                    result.append(char)
        return result
    
    def to_basic_result(self) -> OCRResult:
        """Convert to basic OCRResult."""
        return OCRResult(
            text=self.text,
            confidence=self.confidence,
            word_confidences=self.word_confidences,
            roi_id=self.roi_id,
        )
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "text": self.text,
            "confidence": self.confidence,
            "words": [w.to_dict() for w in self.words],
            "roi_id": self.roi_id,
        }


class TesseractOCR:
    """
    Tesseract OCR wrapper with configurable settings.
    
    Provides high-level methods for text recognition with
    support for whitelists and detailed output.
    """
    
    def __init__(self, config: Optional[TesseractConfig] = None):
        """
        Initialize Tesseract OCR.
        
        Args:
            config: Tesseract configuration. Uses defaults if None.
            
        Raises:
            ImportError: If pytesseract is not installed.
        """
        if not TESSERACT_AVAILABLE:
            raise ImportError(
                "pytesseract is not installed. "
                "Install with: pip install pytesseract"
            )
        
        self.config = config or TesseractConfig()
        self._validate_tesseract()
    
    def _validate_tesseract(self) -> None:
        """Validate Tesseract installation."""
        try:
            pytesseract.get_tesseract_version()
        except pytesseract.TesseractNotFoundError:
            raise RuntimeError(
                "Tesseract is not installed or not in PATH. "
                "Install Tesseract-OCR and ensure it's accessible."
            )
    
    def recognize(
        self,
        image: np.ndarray,
        roi_id: Optional[str] = None,
        config_override: Optional[TesseractConfig] = None
    ) -> OCRResult:
        """
        Run Tesseract OCR on an image.
        
        Args:
            image: Input image (grayscale or BGR)
            roi_id: Optional ROI identifier
            config_override: Override default configuration
            
        Returns:
            OCRResult with recognized text and confidence
        """
        config = config_override or self.config
        config_str = config.build_config_string()
        
        # Ensure proper image format
        if len(image.shape) == 3:
            # Convert BGR to RGB for pytesseract
            image = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
        
        # Get text and word data
        try:
            data = pytesseract.image_to_data(
                image,
                lang=config.lang,
                config=config_str,
                output_type=Output.DICT
            )
        except Exception as e:
            return OCRResult(
                text="",
                confidence=0.0,
                word_confidences=[],
                roi_id=roi_id,
            )
        
        # Extract text and confidences
        words = []
        confidences = []
        
        for i, conf in enumerate(data['conf']):
            text = data['text'][i].strip()
            if text and conf != -1:  # -1 means no confidence available
                words.append(text)
                confidences.append(float(conf))
        
        full_text = " ".join(words)
        avg_confidence = sum(confidences) / len(confidences) if confidences else 0.0
        
        return OCRResult(
            text=full_text,
            confidence=avg_confidence,
            word_confidences=confidences,
            roi_id=roi_id,
        )
    
    def recognize_with_details(
        self,
        image: np.ndarray,
        roi_id: Optional[str] = None,
        config_override: Optional[TesseractConfig] = None
    ) -> DetailedOCRResult:
        """
        Run Tesseract OCR with character-level details.
        
        Args:
            image: Input image (grayscale or BGR)
            roi_id: Optional ROI identifier
            config_override: Override default configuration
            
        Returns:
            DetailedOCRResult with character-level confidence
        """
        config = config_override or self.config
        config_str = config.build_config_string()
        
        # Ensure proper image format
        if len(image.shape) == 3:
            image = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
        
        # Get detailed data
        try:
            data = pytesseract.image_to_data(
                image,
                lang=config.lang,
                config=config_str,
                output_type=Output.DICT
            )
        except Exception as e:
            return DetailedOCRResult(
                text="",
                confidence=0.0,
                words=[],
                raw_data={},
                roi_id=roi_id,
            )
        
        # Parse into structured format
        words = []
        all_texts = []
        all_confidences = []
        
        current_word = None
        
        for i in range(len(data['text'])):
            text = data['text'][i].strip()
            conf = data['conf'][i]
            level = data['level'][i]
            
            if not text or conf == -1:
                continue
            
            # Level 5 is word level in Tesseract
            if level == 5:
                word_detail = WordDetail(
                    text=text,
                    confidence=float(conf),
                    x=data['left'][i],
                    y=data['top'][i],
                    width=data['width'][i],
                    height=data['height'][i],
                )
                
                # Add character details (approximated from word)
                # Tesseract doesn't provide char-level boxes by default
                char_width = word_detail.width / len(text) if text else 0
                for j, char in enumerate(text):
                    char_detail = CharacterDetail(
                        char=char,
                        confidence=float(conf),  # Use word confidence
                        x=int(word_detail.x + j * char_width),
                        y=word_detail.y,
                        width=int(char_width),
                        height=word_detail.height,
                    )
                    word_detail.characters.append(char_detail)
                
                words.append(word_detail)
                all_texts.append(text)
                all_confidences.append(float(conf))
        
        full_text = " ".join(all_texts)
        avg_confidence = sum(all_confidences) / len(all_confidences) if all_confidences else 0.0
        
        return DetailedOCRResult(
            text=full_text,
            confidence=avg_confidence,
            words=words,
            raw_data=data,
            roi_id=roi_id,
        )
    
    def recognize_line(
        self,
        image: np.ndarray,
        roi_id: Optional[str] = None
    ) -> OCRResult:
        """
        Recognize a single line of text.
        
        Uses PSM 7 (single text line) for better results on
        cropped text lines.
        
        Args:
            image: Image containing a single text line
            roi_id: Optional ROI identifier
            
        Returns:
            OCRResult
        """
        line_config = TesseractConfig(
            lang=self.config.lang,
            psm=PageSegmentationMode.SINGLE_LINE,
            oem=self.config.oem,
            whitelist=self.config.whitelist,
            blacklist=self.config.blacklist,
            tessdata_path=self.config.tessdata_path,
        )
        return self.recognize(image, roi_id, line_config)
    
    def recognize_word(
        self,
        image: np.ndarray,
        roi_id: Optional[str] = None
    ) -> OCRResult:
        """
        Recognize a single word.
        
        Uses PSM 8 (single word) for better results on
        cropped single words.
        
        Args:
            image: Image containing a single word
            roi_id: Optional ROI identifier
            
        Returns:
            OCRResult
        """
        word_config = TesseractConfig(
            lang=self.config.lang,
            psm=PageSegmentationMode.SINGLE_WORD,
            oem=self.config.oem,
            whitelist=self.config.whitelist,
            blacklist=self.config.blacklist,
            tessdata_path=self.config.tessdata_path,
        )
        return self.recognize(image, roi_id, word_config)
    
    def recognize_mrz(
        self,
        image: np.ndarray,
        roi_id: Optional[str] = None
    ) -> OCRResult:
        """
        Recognize Machine Readable Zone (MRZ) text.
        
        Uses specialized settings for passport MRZ lines.
        
        Args:
            image: Image containing MRZ text
            roi_id: Optional ROI identifier
            
        Returns:
            OCRResult
        """
        mrz_config = TesseractConfig(
            lang=self.config.lang,
            psm=PageSegmentationMode.SINGLE_LINE,
            oem=self.config.oem,
            whitelist="ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789<",
            tessdata_path=self.config.tessdata_path,
        )
        return self.recognize(image, roi_id, mrz_config)
    
    def recognize_numeric(
        self,
        image: np.ndarray,
        roi_id: Optional[str] = None
    ) -> OCRResult:
        """
        Recognize numeric text only.
        
        Args:
            image: Image containing numeric text
            roi_id: Optional ROI identifier
            
        Returns:
            OCRResult
        """
        numeric_config = TesseractConfig(
            lang=self.config.lang,
            psm=self.config.psm,
            oem=self.config.oem,
            whitelist="0123456789",
            tessdata_path=self.config.tessdata_path,
        )
        return self.recognize(image, roi_id, numeric_config)
    
    def recognize_alpha(
        self,
        image: np.ndarray,
        include_space: bool = True,
        roi_id: Optional[str] = None
    ) -> OCRResult:
        """
        Recognize alphabetic text only.
        
        Args:
            image: Image containing alphabetic text
            include_space: Include space character
            roi_id: Optional ROI identifier
            
        Returns:
            OCRResult
        """
        whitelist = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
        if include_space:
            whitelist += " "
        
        alpha_config = TesseractConfig(
            lang=self.config.lang,
            psm=self.config.psm,
            oem=self.config.oem,
            whitelist=whitelist,
            tessdata_path=self.config.tessdata_path,
        )
        return self.recognize(image, roi_id, alpha_config)
    
    def recognize_date(
        self,
        image: np.ndarray,
        separator: str = "/",
        roi_id: Optional[str] = None
    ) -> OCRResult:
        """
        Recognize date text.
        
        Args:
            image: Image containing date text
            separator: Date separator character
            roi_id: Optional ROI identifier
            
        Returns:
            OCRResult
        """
        whitelist = f"0123456789{separator}"
        
        date_config = TesseractConfig(
            lang=self.config.lang,
            psm=PageSegmentationMode.SINGLE_LINE,
            oem=self.config.oem,
            whitelist=whitelist,
            tessdata_path=self.config.tessdata_path,
        )
        return self.recognize(image, roi_id, date_config)
    
    def batch_recognize(
        self,
        images: List[np.ndarray],
        roi_ids: Optional[List[str]] = None
    ) -> List[OCRResult]:
        """
        Recognize text in multiple images.
        
        Args:
            images: List of input images
            roi_ids: Optional list of ROI identifiers
            
        Returns:
            List of OCRResults
        """
        if roi_ids is None:
            roi_ids = [None] * len(images)
        
        return [
            self.recognize(img, roi_id)
            for img, roi_id in zip(images, roi_ids)
        ]
