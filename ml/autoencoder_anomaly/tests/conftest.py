"""
Test fixtures for autoencoder anomaly detection tests.

VE-924: Test fixtures and utilities for autoencoder anomaly detection module.
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
def anomaly_config():
    """Default autoencoder anomaly configuration for tests."""
    from ml.autoencoder_anomaly.config import AutoencoderAnomalyConfig
    return AutoencoderAnomalyConfig()


@pytest.fixture
def encoder_config():
    """Default encoder configuration for tests."""
    from ml.autoencoder_anomaly.config import EncoderConfig
    return EncoderConfig()


@pytest.fixture
def decoder_config():
    """Default decoder configuration for tests."""
    from ml.autoencoder_anomaly.config import DecoderConfig
    return DecoderConfig()


@pytest.fixture
def reconstruction_config():
    """Default reconstruction configuration for tests."""
    from ml.autoencoder_anomaly.config import ReconstructionConfig
    return ReconstructionConfig()


@pytest.fixture
def scoring_config():
    """Default scoring configuration for tests."""
    from ml.autoencoder_anomaly.config import AnomalyScoringConfig
    return AnomalyScoringConfig()


@pytest.fixture
def sample_image(np_random) -> np.ndarray:
    """
    Create a sample face-like image for testing.
    
    Creates a synthetic BGR image with face-like features.
    """
    # Create a 128x128 BGR image
    image = np.zeros((128, 128, 3), dtype=np.uint8)
    
    # Fill with skin-tone background
    image[:, :] = [180, 200, 220]  # BGR
    
    # Add ellipse for face shape
    center = (64, 64)
    axes = (40, 50)
    
    y, x = np.ogrid[:128, :128]
    face_mask = ((x - center[0])**2 / axes[0]**2 + (y - center[1])**2 / axes[1]**2) <= 1
    image[face_mask] = [170, 190, 210]
    
    # Add "eyes" as dark spots
    left_eye = (48, 55)
    right_eye = (80, 55)
    for eye in [left_eye, right_eye]:
        eye_mask = ((x - eye[0])**2 + (y - eye[1])**2) <= 36
        image[eye_mask] = [50, 50, 50]
    
    # Add natural noise
    noise = np_random.randint(-10, 10, image.shape, dtype=np.int16)
    image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def sample_image_normalized(sample_image) -> np.ndarray:
    """Return normalized version of sample image."""
    # Convert BGR to RGB and normalize
    image = sample_image[:, :, ::-1].astype(np.float32) / 255.0
    return image


@pytest.fixture
def anomalous_image(np_random) -> np.ndarray:
    """
    Create an anomalous image for testing anomaly detection.
    
    This image has unusual patterns that should trigger anomaly detection.
    """
    # Create a 128x128 BGR image with anomalous patterns
    image = np.zeros((128, 128, 3), dtype=np.uint8)
    
    # Add checkerboard pattern (GAN artifact)
    for i in range(0, 128, 8):
        for j in range(0, 128, 8):
            if (i // 8 + j // 8) % 2 == 0:
                image[i:i+8, j:j+8] = [200, 200, 200]
            else:
                image[i:i+8, j:j+8] = [50, 50, 50]
    
    # Add random noise
    noise = np_random.randint(-20, 20, image.shape, dtype=np.int16)
    image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def small_image(np_random) -> np.ndarray:
    """Create an image that's too small."""
    return np_random.randint(0, 256, (32, 32, 3), dtype=np.uint8)


@pytest.fixture
def large_image(np_random) -> np.ndarray:
    """Create an image that's too large."""
    return np_random.randint(0, 256, (5000, 5000, 3), dtype=np.uint8)


@pytest.fixture
def sample_latent_vector(np_random) -> np.ndarray:
    """Create a sample latent vector."""
    return np_random.randn(128).astype(np.float32)


@pytest.fixture
def outlier_latent_vector(np_random) -> np.ndarray:
    """Create an outlier latent vector."""
    # Normal vector with some extreme values
    latent = np_random.randn(128).astype(np.float32)
    # Add outliers
    latent[0:10] = 10.0  # High values
    latent[50:60] = -10.0  # Low values
    return latent


@pytest.fixture
def identical_images() -> Tuple[np.ndarray, np.ndarray]:
    """Create two identical normalized images."""
    image = np.random.RandomState(42).rand(128, 128, 3).astype(np.float32)
    return image, image.copy()


@pytest.fixture
def different_images(np_random) -> Tuple[np.ndarray, np.ndarray]:
    """Create two different normalized images."""
    image1 = np_random.rand(128, 128, 3).astype(np.float32)
    image2 = np_random.rand(128, 128, 3).astype(np.float32)
    return image1, image2
