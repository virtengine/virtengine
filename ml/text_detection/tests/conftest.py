"""
Test fixtures for text detection tests.

This module provides common fixtures and utilities for testing
the text detection pipeline.
"""

import os
import pytest
import numpy as np
from typing import Tuple, List

# Ensure deterministic behavior in tests
os.environ["CUDA_VISIBLE_DEVICES"] = "-1"
os.environ["TF_DETERMINISTIC_OPS"] = "1"


@pytest.fixture
def seed():
    """Fixed random seed for reproducibility."""
    return 42


@pytest.fixture
def np_random(seed):
    """NumPy random generator with fixed seed."""
    return np.random.RandomState(seed)


@pytest.fixture
def craft_config():
    """Default CRAFT configuration for tests."""
    from ml.text_detection.config import CRAFTConfig, DeviceType
    return CRAFTConfig(
        device=DeviceType.CPU,
        deterministic=True,
        text_threshold=0.7,
        link_threshold=0.4,
    )


@pytest.fixture
def postprocessing_config():
    """Default post-processing configuration for tests."""
    from ml.text_detection.config import PostProcessingConfig
    return PostProcessingConfig()


@pytest.fixture
def text_detection_config(craft_config, postprocessing_config):
    """Complete text detection configuration for tests."""
    from ml.text_detection.config import TextDetectionConfig
    return TextDetectionConfig(
        craft=craft_config,
        postprocessing=postprocessing_config,
    )


@pytest.fixture
def postprocessor(postprocessing_config):
    """Text post-processor instance."""
    from ml.text_detection.postprocessing import TextPostProcessor
    return TextPostProcessor(postprocessing_config)


@pytest.fixture
def sample_document_image(np_random) -> np.ndarray:
    """
    Create a sample document-like image with text regions for testing.
    
    Creates a synthetic ID card-like image with text-like patterns.
    """
    # Create a 1200x800 BGR image (landscape ID card size)
    width, height = 1200, 800
    image = np.ones((height, width, 3), dtype=np.uint8) * 240  # Light gray background
    
    # Import cv2 for drawing
    cv2 = __import__('cv2')
    
    # Add a border
    border_color = [50, 50, 50]
    cv2.rectangle(image, (20, 20), (width - 20, height - 20), border_color, 3)
    
    # Add header area (darker)
    image[30:100, 30:width - 30] = [200, 220, 240]
    
    # Add simulated text lines (horizontal dark regions)
    for y in range(150, 600, 50):
        line_width = np_random.randint(200, 500)
        x_start = 50
        # Add individual "characters" with gaps
        x = x_start
        while x < x_start + line_width:
            char_width = np_random.randint(8, 20)
            char_height = np_random.randint(15, 25)
            if x + char_width < width - 50:
                image[y:y+char_height, x:x+char_width] = [30, 30, 30]
            x += char_width + np_random.randint(2, 8)  # Gap between chars
    
    # Add simulated photo area (rectangle on right side)
    photo_x, photo_y = width - 280, 120
    photo_w, photo_h = 200, 250
    image[photo_y:photo_y + photo_h, photo_x:photo_x + photo_w] = [180, 180, 220]
    
    # Add some noise
    noise = np_random.randint(-5, 5, image.shape, dtype=np.int16)
    image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def small_text_image(np_random) -> np.ndarray:
    """Create a smaller image with clear text-like regions for faster tests."""
    width, height = 400, 300
    image = np.ones((height, width, 3), dtype=np.uint8) * 245
    
    cv2 = __import__('cv2')
    
    # Add border
    cv2.rectangle(image, (10, 10), (width - 10, height - 10), [100, 100, 100], 2)
    
    # Add clear text-like blocks
    for y in range(40, 250, 40):
        x = 30
        while x < 350:
            char_w = np_random.randint(10, 18)
            char_h = np_random.randint(18, 28)
            image[y:y+char_h, x:x+char_w] = [20, 20, 20]
            x += char_w + np_random.randint(3, 8)
    
    return image


@pytest.fixture
def grayscale_text_image(np_random) -> np.ndarray:
    """Create a grayscale document image."""
    width, height = 400, 300
    image = np.ones((height, width), dtype=np.uint8) * 240
    
    # Add text-like lines
    for y in range(40, 250, 40):
        x = 30
        while x < 350:
            char_w = np_random.randint(10, 18)
            char_h = np_random.randint(18, 28)
            image[y:y+char_h, x:x+char_w] = 30
            x += char_w + np_random.randint(3, 8)
    
    return image


@pytest.fixture
def sample_region_scores(np_random) -> np.ndarray:
    """Create sample region score map for testing post-processing."""
    height, width = 300, 400
    scores = np.zeros((height, width), dtype=np.float32)
    
    # Add some high-score regions (simulating detected text)
    regions = [
        (50, 30, 20, 25),    # (x, y, w, h)
        (80, 30, 18, 25),
        (110, 30, 22, 25),
        (50, 80, 20, 25),
        (78, 80, 20, 25),
        (200, 30, 25, 25),
        (230, 30, 20, 25),
    ]
    
    for x, y, w, h in regions:
        # Create gaussian-like peak
        center_val = 0.85 + np_random.uniform(-0.1, 0.1)
        scores[y:y+h, x:x+w] = center_val
        # Add some falloff
        if y > 2:
            scores[y-2:y, x:x+w] = center_val * 0.5
        if y + h + 2 < height:
            scores[y+h:y+h+2, x:x+w] = center_val * 0.5
    
    return scores


@pytest.fixture
def sample_affinity_scores(np_random) -> np.ndarray:
    """Create sample affinity score map for testing character linking."""
    height, width = 300, 400
    scores = np.zeros((height, width), dtype=np.float32)
    
    # Add affinity between adjacent character positions
    links = [
        (68, 35, 12, 15),   # Between first two chars
        (98, 35, 12, 15),   # Between chars 2 and 3
        (68, 85, 10, 15),   # Between chars on line 2
    ]
    
    for x, y, w, h in links:
        link_val = 0.6 + np_random.uniform(-0.1, 0.1)
        scores[y:y+h, x:x+w] = link_val
    
    return scores


@pytest.fixture
def sample_text_rois() -> List:
    """Create sample TextROI objects for testing."""
    from ml.text_detection.roi_types import TextROI, BoundingBox, TextType
    
    rois = [
        TextROI.create(
            bounding_box=BoundingBox(x=50, y=30, width=20, height=25),
            confidence=0.9,
            text_type=TextType.CHARACTER,
            region_score=0.85,
            affinity_score=0.1,
        ),
        TextROI.create(
            bounding_box=BoundingBox(x=75, y=30, width=18, height=25),
            confidence=0.85,
            text_type=TextType.CHARACTER,
            region_score=0.82,
            affinity_score=0.6,
        ),
        TextROI.create(
            bounding_box=BoundingBox(x=98, y=30, width=22, height=25),
            confidence=0.88,
            text_type=TextType.CHARACTER,
            region_score=0.84,
            affinity_score=0.55,
        ),
        TextROI.create(
            bounding_box=BoundingBox(x=50, y=80, width=20, height=25),
            confidence=0.82,
            text_type=TextType.CHARACTER,
            region_score=0.78,
            affinity_score=0.5,
        ),
        TextROI.create(
            bounding_box=BoundingBox(x=75, y=80, width=20, height=25),
            confidence=0.80,
            text_type=TextType.CHARACTER,
            region_score=0.76,
            affinity_score=0.1,
        ),
    ]
    
    return rois


@pytest.fixture
def overlapping_rois() -> List:
    """Create overlapping ROIs for NMS testing."""
    from ml.text_detection.roi_types import TextROI, BoundingBox, TextType
    
    return [
        TextROI.create(
            bounding_box=BoundingBox(x=50, y=50, width=30, height=30),
            confidence=0.95,
            text_type=TextType.CHARACTER,
            region_score=0.9,
        ),
        TextROI.create(
            bounding_box=BoundingBox(x=55, y=52, width=28, height=30),  # Overlaps first
            confidence=0.85,
            text_type=TextType.CHARACTER,
            region_score=0.82,
        ),
        TextROI.create(
            bounding_box=BoundingBox(x=150, y=50, width=30, height=30),  # No overlap
            confidence=0.80,
            text_type=TextType.CHARACTER,
            region_score=0.78,
        ),
    ]
