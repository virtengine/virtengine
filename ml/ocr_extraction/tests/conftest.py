"""
Pytest fixtures for OCR extraction tests.
"""

import pytest
import numpy as np
from typing import List

from ml.text_detection import TextROI, BoundingBox, Point, TextType
from ml.ocr_extraction.config import (
    OCRExtractionConfig,
    TesseractConfig,
    CropperConfig,
    PostProcessingConfig,
    FieldParserConfig,
    HashingConfig,
    DocumentType,
    ThresholdingMethod,
)
from ml.ocr_extraction.tesseract_wrapper import OCRResult
from ml.ocr_extraction.field_parser import ParsedField, ValidationStatus


@pytest.fixture
def sample_image() -> np.ndarray:
    """Create a sample grayscale image."""
    # 200x100 white image with some gray areas
    image = np.ones((100, 200), dtype=np.uint8) * 255
    # Add some text-like dark region
    image[20:40, 20:180] = 50
    return image


@pytest.fixture
def sample_color_image() -> np.ndarray:
    """Create a sample BGR color image."""
    # 200x100 white image
    image = np.ones((100, 200, 3), dtype=np.uint8) * 255
    # Add some text-like dark region
    image[20:40, 20:180] = [50, 50, 50]
    return image


@pytest.fixture
def sample_document_image() -> np.ndarray:
    """Create a larger document-like image."""
    # 800x600 document image
    image = np.ones((600, 800, 3), dtype=np.uint8) * 240
    
    # Add some "text" regions
    for y in range(50, 550, 50):
        x_start = 50 + (y % 100)
        image[y:y+20, x_start:x_start+400] = [30, 30, 30]
    
    return image


@pytest.fixture
def sample_roi() -> TextROI:
    """Create a sample text ROI."""
    bbox = BoundingBox(x=20, y=20, width=160, height=20)
    return TextROI.create(
        bounding_box=bbox,
        confidence=0.9,
        text_type=TextType.LINE,
        affinity_score=0.8,
        region_score=0.85,
    )


@pytest.fixture
def sample_rois() -> List[TextROI]:
    """Create multiple sample ROIs."""
    rois = []
    for i in range(5):
        bbox = BoundingBox(
            x=50 + (i * 10),
            y=50 + (i * 50),
            width=400,
            height=20
        )
        roi = TextROI.create(
            bounding_box=bbox,
            confidence=0.8 + (i * 0.02),
            text_type=TextType.LINE,
        )
        rois.append(roi)
    return rois


@pytest.fixture
def rotated_roi() -> TextROI:
    """Create a rotated text ROI with polygon."""
    # Slightly rotated rectangle
    polygon = [
        Point(20, 25),
        Point(180, 20),
        Point(182, 40),
        Point(22, 45),
    ]
    bbox = BoundingBox.from_points(polygon)
    
    return TextROI(
        roi_id="roi_rotated_001",
        bounding_box=bbox,
        confidence=0.85,
        text_type=TextType.LINE,
        polygon=polygon,
        affinity_score=0.7,
        region_score=0.8,
    )


@pytest.fixture
def default_config() -> OCRExtractionConfig:
    """Create default OCR extraction config."""
    return OCRExtractionConfig()


@pytest.fixture
def test_config() -> OCRExtractionConfig:
    """Create config optimized for testing."""
    return OCRExtractionConfig(
        cropper=CropperConfig(
            margin_pixels=5,
            use_subpixel_crop=True,
            thresholding=ThresholdingMethod.OTSU,
        ),
        hashing=HashingConfig(
            salt="test_salt_12345",
            normalize_before_hash=True,
            uppercase_before_hash=True,
        ),
    )


@pytest.fixture
def sample_ocr_results() -> List[OCRResult]:
    """Create sample OCR results."""
    return [
        OCRResult(
            text="SURNAME: SMITH",
            confidence=85.0,
            word_confidences=[90.0, 80.0],
            roi_id="roi_001",
        ),
        OCRResult(
            text="GIVEN NAMES: JOHN WILLIAM",
            confidence=82.0,
            word_confidences=[88.0, 78.0, 80.0],
            roi_id="roi_002",
        ),
        OCRResult(
            text="DATE OF BIRTH: 15/03/1985",
            confidence=88.0,
            word_confidences=[90.0, 85.0, 90.0, 87.0],
            roi_id="roi_003",
        ),
        OCRResult(
            text="DOCUMENT NO: ABC123456",
            confidence=90.0,
            word_confidences=[92.0, 88.0],
            roi_id="roi_004",
        ),
    ]


@pytest.fixture
def sample_passport_ocr_results() -> List[OCRResult]:
    """Create sample passport OCR results with MRZ."""
    return [
        OCRResult(
            text="PASSPORT",
            confidence=95.0,
            word_confidences=[95.0],
            roi_id="roi_001",
        ),
        OCRResult(
            text="UNITED KINGDOM",
            confidence=90.0,
            word_confidences=[90.0, 90.0],
            roi_id="roi_002",
        ),
        OCRResult(
            text="P<GBRSMITH<<JOHN<WILLIAM<<<<<<<<<<<<<<<<<<<<<",
            confidence=92.0,
            word_confidences=[92.0],
            roi_id="roi_mrz1",
        ),
        OCRResult(
            text="AB1234567<8GBR8503151M2512319<<<<<<<<<<<<<<02",
            confidence=90.0,
            word_confidences=[90.0],
            roi_id="roi_mrz2",
        ),
    ]


@pytest.fixture
def sample_parsed_fields() -> dict:
    """Create sample parsed fields."""
    return {
        "surname": ParsedField(
            field_name="surname",
            value="SMITH",
            confidence=0.85,
            source_roi_ids=["roi_001"],
            validation_status=ValidationStatus.VALID,
            raw_value="SMITH",
        ),
        "given_names": ParsedField(
            field_name="given_names",
            value="JOHN WILLIAM",
            confidence=0.82,
            source_roi_ids=["roi_002"],
            validation_status=ValidationStatus.VALID,
            raw_value="JOHN WILLIAM",
        ),
        "date_of_birth": ParsedField(
            field_name="date_of_birth",
            value="15/03/1985",
            confidence=0.88,
            source_roi_ids=["roi_003"],
            validation_status=ValidationStatus.VALID,
            raw_value="15/03/1985",
        ),
        "document_number": ParsedField(
            field_name="document_number",
            value="ABC123456",
            confidence=0.90,
            source_roi_ids=["roi_004"],
            validation_status=ValidationStatus.VALID,
            raw_value="ABC123456",
        ),
    }


@pytest.fixture
def noisy_ocr_text() -> str:
    """Create OCR text with common errors."""
    return "JOHN SM1TH  0OB: 15/O3/l985  lD: ABC12345G"


@pytest.fixture
def mrz_lines() -> List[str]:
    """Create sample MRZ lines."""
    return [
        "P<GBRSMITH<<JOHN<WILLIAM<<<<<<<<<<<<<<<<<<<<<",
        "AB1234567<8GBR8503151M2512319<<<<<<<<<<<<<<02",
    ]
