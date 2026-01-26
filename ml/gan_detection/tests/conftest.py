"""
Test fixtures for GAN detection tests.

VE-923: Test fixtures and utilities for GAN detection module.
"""

import os
import pytest
import numpy as np
from typing import List, Tuple

# Ensure deterministic behavior in tests
os.environ["CUDA_VISIBLE_DEVICES"] = "-1"


@pytest.fixture
def seed():
    """Fixed random seed for reproducibility."""
    return 42


@pytest.fixture
def np_random(seed):
    """NumPy random generator with fixed seed."""
    return np.random.RandomState(seed)


@pytest.fixture
def gan_config():
    """Default GAN detection configuration for tests."""
    from ml.gan_detection.config import GANDetectionConfig
    return GANDetectionConfig()


@pytest.fixture
def discriminator_config():
    """Default discriminator configuration for tests."""
    from ml.gan_detection.config import DiscriminatorConfig
    return DiscriminatorConfig()


@pytest.fixture
def deepfake_config():
    """Default deepfake configuration for tests."""
    from ml.gan_detection.config import DeepfakeConfig
    return DeepfakeConfig()


@pytest.fixture
def artifact_config():
    """Default artifact analysis configuration for tests."""
    from ml.gan_detection.config import ArtifactAnalysisConfig
    return ArtifactAnalysisConfig()


@pytest.fixture
def sample_face_frame(np_random) -> np.ndarray:
    """
    Create a sample face-like frame for testing.
    
    Creates a synthetic BGR image with face-like features.
    """
    # Create a 224x224 BGR image
    image = np.zeros((224, 224, 3), dtype=np.uint8)
    
    # Fill with skin-tone background
    image[:, :] = [180, 200, 220]  # BGR
    
    # Add ellipse for face shape
    center = (112, 112)
    axes = (60, 80)
    
    y, x = np.ogrid[:224, :224]
    face_mask = ((x - center[0])**2 / axes[0]**2 + (y - center[1])**2 / axes[1]**2) <= 1
    image[face_mask] = [170, 190, 210]
    
    # Add "eyes" as dark spots
    left_eye = (80, 95)
    right_eye = (144, 95)
    for eye in [left_eye, right_eye]:
        eye_mask = ((x - eye[0])**2 + (y - eye[1])**2) <= 100
        image[eye_mask] = [50, 50, 50]
    
    # Add natural noise
    noise = np_random.randint(-10, 10, image.shape, dtype=np.int16)
    image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def sample_frame_sequence(sample_face_frame, np_random) -> List[np.ndarray]:
    """Create a sequence of frames with slight variations."""
    frames = [sample_face_frame.copy()]
    
    for i in range(19):  # 20 frames total
        frame = sample_face_frame.copy()
        
        # Add slight brightness variation
        brightness_shift = np_random.randint(-3, 3)
        frame = np.clip(frame.astype(np.int16) + brightness_shift, 0, 255).astype(np.uint8)
        
        # Add noise for natural variation
        noise = np_random.randint(-2, 2, frame.shape, dtype=np.int16)
        frame = np.clip(frame.astype(np.int16) + noise, 0, 255).astype(np.uint8)
        
        frames.append(frame)
    
    return frames


@pytest.fixture
def static_frame_sequence(sample_face_frame) -> List[np.ndarray]:
    """Create a sequence of identical frames (photo-like)."""
    return [sample_face_frame.copy() for _ in range(20)]


@pytest.fixture
def sample_face_region() -> Tuple[int, int, int, int]:
    """Sample face bounding box."""
    return (52, 32, 172, 192)


@pytest.fixture
def sample_face_regions(sample_face_region) -> List[Tuple[int, int, int, int]]:
    """Sample face regions for frame sequence."""
    return [sample_face_region for _ in range(20)]


@pytest.fixture
def synthetic_frame(np_random) -> np.ndarray:
    """
    Create a frame simulating GAN-generated characteristics.
    
    Features checkerboard artifacts and unnatural patterns.
    """
    image = np.zeros((224, 224, 3), dtype=np.uint8)
    
    # Base skin tone
    image[:, :] = [180, 200, 220]
    
    # Add checkerboard pattern (GAN artifact)
    for y in range(0, 224, 4):
        for x in range(0, 224, 4):
            if (y // 4 + x // 4) % 2 == 0:
                image[y:y+4, x:x+4] = np.clip(
                    image[y:y+4, x:x+4].astype(np.int16) + 5,
                    0, 255
                ).astype(np.uint8)
    
    # Add very smooth face region (unnatural)
    center = (112, 112)
    axes = (60, 80)
    
    y, x = np.ogrid[:224, :224]
    face_mask = ((x - center[0])**2 / axes[0]**2 + (y - center[1])**2 / axes[1]**2) <= 1
    image[face_mask] = [168, 188, 208]  # Very uniform
    
    return image


@pytest.fixture
def blended_frame(sample_face_frame, np_random) -> np.ndarray:
    """Create a frame with visible blending boundary."""
    image = sample_face_frame.copy()
    
    # Add sharp color transition at face boundary
    h, w = image.shape[:2]
    center_y, center_x = h // 2, w // 2
    
    y, x = np.ogrid[:h, :w]
    
    # Create boundary with color shift
    boundary_mask = (
        (np.abs((x - center_x)**2 / 60**2 + (y - center_y)**2 / 80**2 - 1) < 0.1)
    )
    
    image[boundary_mask] = np.clip(
        image[boundary_mask].astype(np.int16) + 30,
        0, 255
    ).astype(np.uint8)
    
    return image


@pytest.fixture
def deepfake_frame(np_random) -> np.ndarray:
    """Create a frame simulating deepfake characteristics."""
    image = np.zeros((224, 224, 3), dtype=np.uint8)
    
    # Base
    image[:, :] = [180, 200, 220]
    
    # Face with different color temperature (color mismatch)
    center = (112, 112)
    y, x = np.ogrid[:224, :224]
    face_mask = ((x - center[0])**2 / 60**2 + (y - center[1])**2 / 80**2) <= 1
    image[face_mask] = [200, 190, 180]  # Different color temp
    
    # Blending boundary
    boundary = (
        (np.abs((x - center[0])**2 / 60**2 + (y - center[1])**2 / 80**2 - 1) < 0.15)
    )
    image[boundary] = [190, 195, 200]  # Transition color
    
    return image


@pytest.fixture
def high_quality_frame(np_random) -> np.ndarray:
    """Create a high-quality natural frame."""
    image = np.zeros((512, 512, 3), dtype=np.uint8)
    
    # Natural gradient background
    for y in range(512):
        for x in range(512):
            image[y, x] = [
                int(150 + 50 * (y / 512)),
                int(180 + 40 * (y / 512)),
                int(200 + 30 * (y / 512)),
            ]
    
    # Add natural noise
    noise = np_random.randint(-5, 5, image.shape, dtype=np.int16)
    image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    # Add face-like features with texture
    center = (256, 256)
    y, x = np.ogrid[:512, :512]
    face_mask = ((x - center[0])**2 / 100**2 + (y - center[1])**2 / 130**2) <= 1
    
    # Add textured skin
    face_noise = np_random.randint(-8, 8, (512, 512, 3), dtype=np.int16)
    face_base = np.array([160, 185, 210], dtype=np.int16)
    face_pixels = np.clip(face_base + face_noise, 0, 255).astype(np.uint8)
    image[face_mask] = face_pixels[face_mask]
    
    return image


@pytest.fixture
def small_frame() -> np.ndarray:
    """Create a frame smaller than minimum size."""
    return np.zeros((32, 32, 3), dtype=np.uint8)


@pytest.fixture
def large_frame() -> np.ndarray:
    """Create a frame larger than maximum size."""
    return np.zeros((5000, 5000, 3), dtype=np.uint8)
