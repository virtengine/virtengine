"""
Test fixtures for document preprocessing tests.

This module provides common fixtures and utilities for testing
the document preprocessing pipeline.
"""

import os
import pytest
import numpy as np
from typing import Tuple

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
def document_config():
    """Default document configuration for tests."""
    from ml.document_preprocessing.config import DocumentConfig
    return DocumentConfig()


@pytest.fixture
def standardizer(document_config):
    """Document standardizer instance."""
    from ml.document_preprocessing.standardization import DocumentStandardizer
    return DocumentStandardizer(document_config)


@pytest.fixture
def enhancer(document_config):
    """Document enhancer instance."""
    from ml.document_preprocessing.enhancement import DocumentEnhancer
    return DocumentEnhancer(document_config)


@pytest.fixture
def noise_reducer(document_config):
    """Noise reducer instance."""
    from ml.document_preprocessing.noise_reduction import NoiseReducer
    return NoiseReducer(document_config)


@pytest.fixture
def orientation_detector(document_config):
    """Orientation detector instance."""
    from ml.document_preprocessing.orientation import OrientationDetector
    return OrientationDetector(document_config)


@pytest.fixture
def perspective_corrector(document_config):
    """Perspective corrector instance."""
    from ml.document_preprocessing.perspective import PerspectiveCorrector
    return PerspectiveCorrector(document_config)


@pytest.fixture
def pipeline(document_config):
    """Document preprocessing pipeline instance."""
    from ml.document_preprocessing.pipeline import DocumentPreprocessingPipeline
    return DocumentPreprocessingPipeline(document_config)


@pytest.fixture
def sample_document_image(np_random) -> np.ndarray:
    """
    Create a sample document-like image for testing.
    
    Creates a synthetic ID card-like image with text-like patterns.
    """
    # Create a 1200x800 BGR image (portrait ID card size)
    width, height = 1200, 800
    image = np.ones((height, width, 3), dtype=np.uint8) * 240  # Light gray background
    
    # Add a border
    border_color = [50, 50, 50]
    cv2 = __import__('cv2')
    cv2.rectangle(image, (20, 20), (width - 20, height - 20), border_color, 3)
    
    # Add header area (darker)
    image[30:100, 30:width - 30] = [200, 220, 240]
    
    # Add simulated text lines (horizontal dark lines)
    for y in range(150, 600, 50):
        line_width = np_random.randint(200, 400)
        x_start = 50
        image[y:y+8, x_start:x_start + line_width] = [30, 30, 30]
    
    # Add simulated photo area (rectangle on right side)
    photo_x, photo_y = width - 280, 120
    photo_w, photo_h = 200, 250
    image[photo_y:photo_y + photo_h, photo_x:photo_x + photo_w] = [180, 180, 220]
    
    # Add some noise
    noise = np_random.randint(-5, 5, image.shape, dtype=np.int16)
    image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def small_document_image(np_random) -> np.ndarray:
    """Create a smaller document image for faster tests."""
    width, height = 800, 600
    image = np.ones((height, width, 3), dtype=np.uint8) * 245
    
    # Add some structure
    cv2 = __import__('cv2')
    cv2.rectangle(image, (10, 10), (width - 10, height - 10), [100, 100, 100], 2)
    
    # Add text-like lines
    for y in range(50, 400, 30):
        line_width = np_random.randint(100, 300)
        image[y:y+5, 30:30 + line_width] = [20, 20, 20]
    
    return image


@pytest.fixture
def grayscale_document_image(np_random) -> np.ndarray:
    """Create a grayscale document image."""
    width, height = 800, 600
    image = np.ones((height, width), dtype=np.uint8) * 240
    
    # Add text lines
    for y in range(50, 400, 30):
        line_width = np_random.randint(100, 300)
        image[y:y+5, 30:30 + line_width] = 30
    
    return image


@pytest.fixture
def dark_document_image(np_random) -> np.ndarray:
    """Create a dark document image for enhancement testing."""
    width, height = 800, 600
    # Dark gray background
    image = np.ones((height, width, 3), dtype=np.uint8) * 60
    
    # Add slightly lighter text
    for y in range(50, 400, 30):
        line_width = np_random.randint(100, 300)
        image[y:y+5, 30:30 + line_width] = [80, 80, 80]
    
    return image


@pytest.fixture
def bright_document_image(np_random) -> np.ndarray:
    """Create an overexposed document image."""
    width, height = 800, 600
    image = np.ones((height, width, 3), dtype=np.uint8) * 250
    
    # Add very light text
    for y in range(50, 400, 30):
        line_width = np_random.randint(100, 300)
        image[y:y+5, 30:30 + line_width] = [200, 200, 200]
    
    return image


@pytest.fixture
def noisy_document_image(np_random) -> np.ndarray:
    """Create a noisy document image."""
    width, height = 800, 600
    image = np.ones((height, width, 3), dtype=np.uint8) * 240
    
    # Add text
    for y in range(50, 400, 30):
        line_width = np_random.randint(100, 300)
        image[y:y+5, 30:30 + line_width] = [30, 30, 30]
    
    # Add significant noise
    noise = np_random.randint(-30, 30, image.shape, dtype=np.int16)
    image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    # Add salt and pepper noise
    salt = np_random.random(image.shape[:2]) < 0.02
    pepper = np_random.random(image.shape[:2]) < 0.02
    image[salt] = [255, 255, 255]
    image[pepper] = [0, 0, 0]
    
    return image


@pytest.fixture
def rotated_document_image(sample_document_image) -> np.ndarray:
    """Create a 90-degree rotated document image."""
    cv2 = __import__('cv2')
    return cv2.rotate(sample_document_image, cv2.ROTATE_90_CLOCKWISE)


@pytest.fixture
def perspective_distorted_image(np_random) -> np.ndarray:
    """Create a document image with perspective distortion."""
    import cv2
    
    # Create a clean document
    width, height = 800, 600
    image = np.ones((height, width, 3), dtype=np.uint8) * 240
    
    # Add border
    cv2.rectangle(image, (20, 20), (width - 20, height - 20), [50, 50, 50], 3)
    
    # Add text lines
    for y in range(80, 500, 40):
        image[y:y+8, 50:50 + np_random.randint(200, 500)] = [30, 30, 30]
    
    # Define source and destination points for perspective transform
    src_pts = np.float32([
        [0, 0],
        [width - 1, 0],
        [width - 1, height - 1],
        [0, height - 1]
    ])
    
    # Distort by moving corners
    dst_pts = np.float32([
        [50, 30],  # Top-left moved right and down
        [width - 80, 20],  # Top-right moved left and down
        [width - 40, height - 60],  # Bottom-right moved left and up
        [30, height - 40]  # Bottom-left moved right and up
    ])
    
    # Get perspective transform matrix
    matrix = cv2.getPerspectiveTransform(src_pts, dst_pts)
    
    # Apply transform (create larger canvas to show distortion)
    distorted = cv2.warpPerspective(
        image,
        matrix,
        (width + 100, height + 100),
        borderValue=(220, 220, 220)
    )
    
    return distorted


@pytest.fixture
def low_resolution_image(np_random) -> np.ndarray:
    """Create a low resolution document image."""
    width, height = 400, 300
    image = np.ones((height, width, 3), dtype=np.uint8) * 240
    
    for y in range(30, 200, 20):
        line_width = np_random.randint(50, 150)
        image[y:y+3, 20:20 + line_width] = [30, 30, 30]
    
    return image


@pytest.fixture
def high_resolution_image(np_random) -> np.ndarray:
    """Create a high resolution document image."""
    width, height = 3000, 2000
    image = np.ones((height, width, 3), dtype=np.uint8) * 240
    
    cv2 = __import__('cv2')
    cv2.rectangle(image, (50, 50), (width - 50, height - 50), [50, 50, 50], 5)
    
    for y in range(150, 1500, 80):
        line_width = np_random.randint(500, 1500)
        image[y:y+12, 100:100 + line_width] = [30, 30, 30]
    
    return image
