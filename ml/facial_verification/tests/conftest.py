"""
Test fixtures for facial verification tests.

This module provides common fixtures and utilities for testing
the facial verification pipeline.
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
def sample_face_image(np_random) -> np.ndarray:
    """
    Create a sample face-like image for testing.
    
    This creates a synthetic image with basic face-like features.
    """
    # Create a 224x224 BGR image
    image = np.zeros((224, 224, 3), dtype=np.uint8)
    
    # Fill with skin-tone-ish background
    image[:, :] = [180, 200, 220]  # BGR for light skin tone
    
    # Add ellipse for face shape
    center = (112, 112)
    axes = (60, 80)
    
    # Create face region (slightly darker)
    y, x = np.ogrid[:224, :224]
    face_mask = ((x - center[0])**2 / axes[0]**2 + (y - center[1])**2 / axes[1]**2) <= 1
    image[face_mask] = [170, 190, 210]
    
    # Add simple "eyes" as dark spots
    left_eye = (80, 95)
    right_eye = (144, 95)
    for eye in [left_eye, right_eye]:
        y, x = np.ogrid[:224, :224]
        eye_mask = ((x - eye[0])**2 + (y - eye[1])**2) <= 100
        image[eye_mask] = [50, 50, 50]
    
    # Add noise for realism
    noise = np_random.randint(-10, 10, image.shape, dtype=np.int16)
    image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def sample_face_image_variant(sample_face_image, np_random) -> np.ndarray:
    """
    Create a variant of the sample face image (similar but not identical).
    
    This simulates a different photo of the same person.
    """
    # Start with the original image
    variant = sample_face_image.copy()
    
    # Add slight brightness variation
    brightness_shift = np_random.randint(-5, 5)
    variant = np.clip(variant.astype(np.int16) + brightness_shift, 0, 255).astype(np.uint8)
    
    # Add additional noise
    noise = np_random.randint(-5, 5, variant.shape, dtype=np.int16)
    variant = np.clip(variant.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    return variant


@pytest.fixture
def different_face_image(np_random) -> np.ndarray:
    """
    Create a different face image (different person).
    """
    # Create a 224x224 BGR image with different features
    image = np.zeros((224, 224, 3), dtype=np.uint8)
    
    # Different skin tone
    image[:, :] = [140, 160, 180]
    
    # Different face shape (rounder)
    center = (112, 112)
    axes = (70, 70)  # More circular
    
    y, x = np.ogrid[:224, :224]
    face_mask = ((x - center[0])**2 / axes[0]**2 + (y - center[1])**2 / axes[1]**2) <= 1
    image[face_mask] = [130, 150, 170]
    
    # Different eye positions
    left_eye = (85, 100)
    right_eye = (139, 100)
    for eye in [left_eye, right_eye]:
        y, x = np.ogrid[:224, :224]
        eye_mask = ((x - eye[0])**2 + (y - eye[1])**2) <= 80
        image[eye_mask] = [40, 40, 40]
    
    # Add noise
    noise = np_random.randint(-10, 10, image.shape, dtype=np.int16)
    image = np.clip(image.astype(np.int16) + noise, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def no_face_image(np_random) -> np.ndarray:
    """Create an image with no face (random noise)."""
    return np_random.randint(0, 255, (224, 224, 3), dtype=np.uint8)


@pytest.fixture
def low_quality_image(np_random) -> np.ndarray:
    """Create a low quality blurry image."""
    # Very small image that will be low resolution
    return np_random.randint(0, 255, (50, 50, 3), dtype=np.uint8)


@pytest.fixture
def dark_image(np_random) -> np.ndarray:
    """Create a very dark image."""
    image = np_random.randint(0, 30, (224, 224, 3), dtype=np.uint8)
    return image


@pytest.fixture
def bright_image(np_random) -> np.ndarray:
    """Create a very bright/overexposed image."""
    image = np_random.randint(225, 255, (224, 224, 3), dtype=np.uint8)
    return image


@pytest.fixture
def multiple_faces_image(sample_face_image, np_random) -> np.ndarray:
    """Create an image with multiple face-like regions."""
    # Create a wider image with two "faces"
    image = np.zeros((224, 448, 3), dtype=np.uint8)
    image[:, :] = [180, 200, 220]
    
    # First face on the left
    for center_x in [100, 348]:
        center = (center_x, 112)
        axes = (50, 70)
        
        y, x = np.ogrid[:224, :448]
        face_mask = ((x - center[0])**2 / axes[0]**2 + (y - center[1])**2 / axes[1]**2) <= 1
        image[face_mask] = [170, 190, 210]
        
        # Add eyes
        left_eye = (center_x - 20, 95)
        right_eye = (center_x + 20, 95)
        for eye in [left_eye, right_eye]:
            y, x = np.ogrid[:224, :448]
            eye_mask = ((x - eye[0])**2 + (y - eye[1])**2) <= 64
            image[eye_mask] = [50, 50, 50]
    
    return image


@pytest.fixture
def verification_config():
    """Default verification configuration for tests."""
    from ml.facial_verification.config import VerificationConfig
    return VerificationConfig()


@pytest.fixture
def high_security_config():
    """High security verification configuration."""
    from ml.facial_verification.config import HIGH_SECURITY_CONFIG
    return HIGH_SECURITY_CONFIG


@pytest.fixture
def permissive_config():
    """Permissive verification configuration."""
    from ml.facial_verification.config import PERMISSIVE_CONFIG
    return PERMISSIVE_CONFIG


@pytest.fixture
def preprocessor(verification_config):
    """FacePreprocessor instance."""
    from ml.facial_verification.preprocessing import FacePreprocessor
    return FacePreprocessor(verification_config)


@pytest.fixture
def detector(verification_config):
    """FaceDetector instance."""
    from ml.facial_verification.face_detection import FaceDetector
    return FaceDetector(verification_config)


@pytest.fixture
def embedder(verification_config):
    """FaceEmbedder instance."""
    from ml.facial_verification.embeddings import FaceEmbedder
    return FaceEmbedder(verification_config)


@pytest.fixture
def verifier(verification_config):
    """FaceVerifier instance."""
    from ml.facial_verification.verification import FaceVerifier
    return FaceVerifier(verification_config)


@pytest.fixture
def determinism_controller(verification_config):
    """DeterminismController instance."""
    from ml.facial_verification.determinism import DeterminismController
    return DeterminismController(verification_config)


def create_embedding_pair(
    similarity: float, 
    dimension: int = 512,
    seed: int = 42
) -> Tuple[np.ndarray, np.ndarray]:
    """
    Create a pair of embeddings with a specific similarity.
    
    Args:
        similarity: Desired cosine similarity (0-1)
        dimension: Embedding dimension
        seed: Random seed
        
    Returns:
        Tuple of (embedding1, embedding2)
    """
    np.random.seed(seed)
    
    # Create first embedding (random unit vector)
    e1 = np.random.randn(dimension)
    e1 = e1 / np.linalg.norm(e1)
    
    # Create second embedding with desired similarity
    # e2 = similarity * e1 + sqrt(1 - similarity^2) * orthogonal
    orthogonal = np.random.randn(dimension)
    orthogonal = orthogonal - np.dot(orthogonal, e1) * e1
    orthogonal = orthogonal / np.linalg.norm(orthogonal)
    
    e2 = similarity * e1 + np.sqrt(1 - similarity**2) * orthogonal
    e2 = e2 / np.linalg.norm(e2)
    
    return e1, e2
