"""
Test fixtures for liveness detection tests.

VE-901: Test fixtures and utilities for liveness detection module.
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
def photo_attack_frame(np_random) -> np.ndarray:
    """
    Create a frame simulating a printed photo attack.
    
    Characteristics: uniform texture, low saturation, paper-like.
    """
    # Create uniform, low-saturation image
    image = np.zeros((224, 224, 3), dtype=np.uint8)
    
    # Muted colors (low saturation)
    image[:, :] = [150, 155, 160]  # Almost grayscale
    
    # Add face region with minimal variation
    center = (112, 112)
    y, x = np.ogrid[:224, :224]
    face_mask = ((x - center[0])**2 + (y - center[1])**2) <= 4000
    image[face_mask] = [145, 150, 155]
    
    # Add paper texture (high-frequency noise)
    paper_noise = np_random.randint(0, 5, image.shape, dtype=np.uint8)
    image = np.clip(image.astype(np.int16) + paper_noise, 0, 255).astype(np.uint8)
    
    return image


@pytest.fixture
def screen_attack_frame(np_random) -> np.ndarray:
    """
    Create a frame simulating a screen display attack.
    
    Characteristics: moire patterns, color banding.
    """
    # Create image with periodic patterns (moire-like)
    image = np.zeros((224, 224, 3), dtype=np.uint8)
    
    # Base color
    image[:, :] = [180, 200, 220]
    
    # Add periodic pattern (simulating screen pixels)
    for i in range(0, 224, 3):
        image[i, :] = [175, 195, 215]
    for j in range(0, 224, 3):
        image[:, j] = [175, 195, 215]
    
    # Add color banding
    for i in range(0, 224, 10):
        band_color = [170 + (i % 20), 190 + (i % 15), 210 + (i % 10)]
        image[i:i+5, :] = band_color
    
    return image


@pytest.fixture
def sample_face_region() -> Tuple[int, int, int, int]:
    """Sample face bounding box."""
    return (50, 30, 124, 164)  # x, y, width, height


@pytest.fixture
def sample_face_regions(sample_face_region) -> List[Tuple[int, int, int, int]]:
    """List of face regions for frame sequence."""
    return [sample_face_region] * 20


@pytest.fixture
def sample_landmarks():
    """Create sample 68-point facial landmarks."""
    from ml.liveness_detection.actiVIRTENGINE_challenges import LandmarkData
    
    # Generate synthetic 68-point landmarks
    landmarks = np.zeros((68, 2), dtype=np.float32)
    
    # Face outline (points 0-16)
    for i in range(17):
        angle = np.pi * (0.5 + i * 0.0625)
        landmarks[i] = [112 + 60 * np.cos(angle), 112 + 80 * np.sin(angle)]
    
    # Left eyebrow (17-21)
    for i, offset in enumerate(range(17, 22)):
        landmarks[offset] = [70 + i * 8, 70]
    
    # Right eyebrow (22-26)
    for i, offset in enumerate(range(22, 27)):
        landmarks[offset] = [130 + i * 8, 70]
    
    # Nose (27-35)
    for i, offset in enumerate(range(27, 36)):
        landmarks[offset] = [112, 80 + i * 8]
    
    # Left eye (36-41)
    left_eye_center = (80, 95)
    for i, offset in enumerate(range(36, 42)):
        angle = i * np.pi / 3
        landmarks[offset] = [
            left_eye_center[0] + 10 * np.cos(angle),
            left_eye_center[1] + 5 * np.sin(angle)
        ]
    
    # Right eye (42-47)
    right_eye_center = (144, 95)
    for i, offset in enumerate(range(42, 48)):
        angle = i * np.pi / 3
        landmarks[offset] = [
            right_eye_center[0] + 10 * np.cos(angle),
            right_eye_center[1] + 5 * np.sin(angle)
        ]
    
    # Outer mouth (48-59)
    mouth_center = (112, 160)
    for i, offset in enumerate(range(48, 60)):
        angle = i * np.pi / 6
        landmarks[offset] = [
            mouth_center[0] + 20 * np.cos(angle),
            mouth_center[1] + 8 * np.sin(angle)
        ]
    
    # Inner mouth (60-67)
    for i, offset in enumerate(range(60, 68)):
        angle = i * np.pi / 4
        landmarks[offset] = [
            mouth_center[0] + 10 * np.cos(angle),
            mouth_center[1] + 4 * np.sin(angle)
        ]
    
    return LandmarkData(
        full_landmarks=landmarks,
        pose_pitch=0.0,
        pose_yaw=0.0,
        pose_roll=0.0,
        confidence=0.95,
        frame_index=0,
        timestamp_ms=0.0,
    )


@pytest.fixture
def sample_landmarks_sequence(sample_landmarks, np_random):
    """Create a sequence of landmarks with variations."""
    from ml.liveness_detection.actiVIRTENGINE_challenges import LandmarkData
    
    sequence = []
    base_landmarks = sample_landmarks.full_landmarks.copy()
    
    for i in range(20):
        # Add slight variation
        landmarks = base_landmarks.copy()
        noise = np_random.randn(68, 2) * 0.5
        landmarks = landmarks + noise
        
        sequence.append(LandmarkData(
            full_landmarks=landmarks,
            pose_pitch=np_random.randn() * 2,
            pose_yaw=np_random.randn() * 3,
            pose_roll=np_random.randn() * 1,
            confidence=0.95,
            frame_index=i,
            timestamp_ms=i * 33.33,  # ~30 FPS
        ))
    
    return sequence


@pytest.fixture
def blink_landmarks_sequence(sample_landmarks, np_random):
    """Create landmarks sequence with a blink."""
    from ml.liveness_detection.actiVIRTENGINE_challenges import LandmarkData
    
    sequence = []
    base_landmarks = sample_landmarks.full_landmarks.copy()
    
    for i in range(20):
        landmarks = base_landmarks.copy()
        
        # Simulate blink at frames 8-10
        if 8 <= i <= 10:
            # Close eyes by moving top and bottom eyelid points together
            # Left eye: 36-41, Right eye: 42-47
            for eye_start in [36, 42]:
                # Move points 1,2 (top) down
                landmarks[eye_start + 1, 1] += 4
                landmarks[eye_start + 2, 1] += 4
                # Move points 4,5 (bottom) up
                landmarks[eye_start + 4, 1] -= 4
                landmarks[eye_start + 5, 1] -= 4
        
        sequence.append(LandmarkData(
            full_landmarks=landmarks,
            pose_pitch=0.0,
            pose_yaw=0.0,
            pose_roll=0.0,
            confidence=0.95,
            frame_index=i,
            timestamp_ms=i * 33.33,
        ))
    
    return sequence


@pytest.fixture
def smile_landmarks_sequence(sample_landmarks, np_random):
    """Create landmarks sequence with a smile."""
    from ml.liveness_detection.actiVIRTENGINE_challenges import LandmarkData
    
    sequence = []
    base_landmarks = sample_landmarks.full_landmarks.copy()
    
    for i in range(20):
        landmarks = base_landmarks.copy()
        
        # Simulate smile at frames 5-15
        if 5 <= i <= 15:
            # Widen mouth (move corners outward)
            # Outer mouth points 48-59
            landmarks[48, 0] -= 5  # Left corner
            landmarks[54, 0] += 5  # Right corner
            # Also slightly raise corners
            landmarks[48, 1] -= 2
            landmarks[54, 1] -= 2
        
        sequence.append(LandmarkData(
            full_landmarks=landmarks,
            pose_pitch=0.0,
            pose_yaw=0.0,
            pose_roll=0.0,
            confidence=0.95,
            frame_index=i,
            timestamp_ms=i * 33.33,
        ))
    
    return sequence


@pytest.fixture
def head_turn_landmarks_sequence(sample_landmarks, np_random):
    """Create landmarks sequence with a head turn."""
    from ml.liveness_detection.actiVIRTENGINE_challenges import LandmarkData
    
    sequence = []
    
    for i in range(20):
        # Simulate head turn left at frames 5-15
        if 5 <= i <= 15:
            yaw = -20.0  # Turn left
        else:
            yaw = 0.0
        
        sequence.append(LandmarkData(
            full_landmarks=sample_landmarks.full_landmarks.copy(),
            pose_pitch=0.0,
            pose_yaw=yaw,
            pose_roll=0.0,
            confidence=0.95,
            frame_index=i,
            timestamp_ms=i * 33.33,
        ))
    
    return sequence


@pytest.fixture
def liveness_config():
    """Default liveness configuration for tests."""
    from ml.liveness_detection.config import LivenessConfig
    return LivenessConfig()


@pytest.fixture
def strict_config():
    """Strict liveness configuration for tests."""
    from ml.liveness_detection.config import get_strict_config
    return get_strict_config()


@pytest.fixture
def permissiVIRTENGINE_config():
    """Permissive liveness configuration for tests."""
    from ml.liveness_detection.config import get_permissiVIRTENGINE_config
    return get_permissiVIRTENGINE_config()
