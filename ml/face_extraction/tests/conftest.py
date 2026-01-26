"""
Test fixtures for face extraction tests.

This module provides common fixtures and utilities for testing
the face extraction pipeline.
"""

import os
import pytest
import numpy as np
from typing import Tuple

# Ensure deterministic behavior in tests
os.environ["CUDA_VISIBLE_DEVICES"] = "-1"
os.environ["TF_DETERMINISTIC_OPS"] = "1"
os.environ["PYTHONHASHSEED"] = "42"


@pytest.fixture
def seed():
    """Fixed random seed for reproducibility."""
    return 42


@pytest.fixture
def np_random(seed):
    """NumPy random generator with fixed seed."""
    return np.random.RandomState(seed)


@pytest.fixture
def sample_document_image(np_random) -> np.ndarray:
    """
    Create a sample ID document image with a face region.
    
    This creates a synthetic image simulating an ID document
    with a face photo area.
    """
    # Create a 768x1024 BGR image (typical ID document size)
    image = np.ones((768, 1024, 3), dtype=np.uint8) * 240  # Light gray background
    
    # Add border (document edge)
    image[10:758, 10:1014] = [250, 250, 250]  # Slightly lighter interior
    
    # Add "face photo" region (upper left quadrant)
    face_region_x = 50
    face_region_y = 100
    face_region_w = 200
    face_region_h = 250
    
    # Background of face photo area
    image[face_region_y:face_region_y+face_region_h, 
          face_region_x:face_region_x+face_region_w] = [220, 220, 230]
    
    # Add ellipse for face shape
    center_x = face_region_x + face_region_w // 2
    center_y = face_region_y + face_region_h // 2 - 20
    
    y, x = np.ogrid[:768, :1024]
    face_mask = ((x - center_x)**2 / 60**2 + (y - center_y)**2 / 80**2) <= 1
    image[face_mask] = [170, 190, 210]  # Skin tone
    
    # Add simple "eyes" as dark spots
    for eye_x in [center_x - 25, center_x + 25]:
        eye_y = center_y - 15
        eye_mask = ((x - eye_x)**2 + (y - eye_y)**2) <= 64
        image[eye_mask] = [50, 50, 50]
    
    # Add "text" regions (horizontal lines)
    for i in range(5):
        text_y = 400 + i * 40
        image[text_y:text_y+10, 300:900] = [60, 60, 60]
    
    # Add noise for realism
    noise = np_random.randint(-5, 5, image.shape, dtype=np.int16)
    image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def sample_face_mask() -> np.ndarray:
    """
    Create a sample face segmentation mask.
    """
    # Create 768x1024 mask
    mask = np.zeros((768, 1024), dtype=np.float32)
    
    # Face region (ellipse)
    center_x = 150
    center_y = 205
    
    y, x = np.ogrid[:768, :1024]
    face_region = ((x - center_x)**2 / 60**2 + (y - center_y)**2 / 80**2) <= 1
    mask[face_region] = 0.95
    
    return mask


@pytest.fixture
def noisy_face_mask(sample_face_mask, np_random) -> np.ndarray:
    """
    Create a noisy version of the face mask.
    """
    mask = sample_face_mask.copy()
    
    # Add random noise
    noise = np_random.uniform(-0.2, 0.2, mask.shape).astype(np.float32)
    mask = np.clip(mask + noise, 0, 1)
    
    # Add some random spots
    for _ in range(20):
        cx = np_random.randint(0, mask.shape[1])
        cy = np_random.randint(0, mask.shape[0])
        radius = np_random.randint(3, 15)
        
        y, x = np.ogrid[:mask.shape[0], :mask.shape[1]]
        spot = ((x - cx)**2 + (y - cy)**2) <= radius**2
        mask[spot] = np_random.uniform(0.3, 0.7)
    
    return mask


@pytest.fixture
def fragmented_mask(np_random) -> np.ndarray:
    """
    Create a fragmented mask with multiple disconnected components.
    """
    mask = np.zeros((768, 1024), dtype=np.float32)
    
    # Main face region
    center_x = 150
    center_y = 205
    y, x = np.ogrid[:768, :1024]
    face_region = ((x - center_x)**2 / 60**2 + (y - center_y)**2 / 80**2) <= 1
    mask[face_region] = 0.9
    
    # Add smaller disconnected components
    for _ in range(5):
        cx = np_random.randint(300, 900)
        cy = np_random.randint(100, 600)
        radius = np_random.randint(10, 30)
        
        spot = ((x - cx)**2 + (y - cy)**2) <= radius**2
        mask[spot] = 0.8
    
    return mask


@pytest.fixture
def low_quality_document(sample_document_image, np_random) -> np.ndarray:
    """
    Create a low quality (noisy, low contrast) document image.
    """
    image = sample_document_image.copy()
    
    # Reduce contrast
    image = (image.astype(np.float32) * 0.5 + 64).astype(np.uint8)
    
    # Add heavy noise
    noise = np_random.randint(-30, 30, image.shape, dtype=np.int16)
    image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def bright_document(sample_document_image) -> np.ndarray:
    """
    Create an overexposed (bright) document image.
    """
    image = sample_document_image.copy()
    
    # Increase brightness
    image = np.clip(image.astype(np.int16) + 60, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def dark_document(sample_document_image) -> np.ndarray:
    """
    Create an underexposed (dark) document image.
    """
    image = sample_document_image.copy()
    
    # Decrease brightness
    image = np.clip(image.astype(np.int16) - 80, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def small_face_document(np_random) -> np.ndarray:
    """
    Create a document with a very small face region.
    """
    image = np.ones((768, 1024, 3), dtype=np.uint8) * 240
    
    # Small face region
    center_x = 100
    center_y = 150
    
    y, x = np.ogrid[:768, :1024]
    face_mask = ((x - center_x)**2 / 20**2 + (y - center_y)**2 / 25**2) <= 1
    image[face_mask] = [170, 190, 210]
    
    return image


@pytest.fixture
def face_extraction_config():
    """Default face extraction configuration for tests."""
    from ml.face_extraction.config import FaceExtractionConfig
    
    return FaceExtractionConfig(
        debug_mode=True,
        return_intermediate_results=True,
    )


@pytest.fixture
def unet_config():
    """U-Net configuration for tests."""
    from ml.face_extraction.config import UNetConfig
    
    return UNetConfig(
        input_size=(256, 256),
        use_gpu=False,
        ensure_determinism=True,
        random_seed=42,
    )


@pytest.fixture
def mask_config():
    """Mask processing configuration for tests."""
    from ml.face_extraction.config import MaskProcessingConfig
    
    return MaskProcessingConfig(
        threshold=0.5,
        apply_morphology=True,
        use_largest_component=True,
        smooth_contours=True,
    )


@pytest.fixture
def cropper_config():
    """Cropper configuration for tests."""
    from ml.face_extraction.config import CropperConfig
    
    return CropperConfig(
        margin=0.15,
        output_size=(224, 224),
        maintain_aspect_ratio=True,
        min_face_width=30,
        min_face_height=30,
        min_face_percentage=0.005,
    )
