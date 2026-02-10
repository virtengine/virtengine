"""
Tests for the MTCNN face detection integration module.

This module provides comprehensive tests for:
- MTCNN three-stage cascade detection (P-Net, R-Net, O-Net)
- Five-point facial landmark extraction
- Face alignment using similarity transformation
- Deterministic execution
"""

import pytest
import numpy as np
import cv2
from typing import Tuple

from ml.facial_verification.mtcnn_detector import (
    MTCNNDetector,
    MTCNNConfig,
    MTCNNDetection,
    MTCNNDetectionResult,
    FiveLandmarks,
    FaceAligner,
    ALIGNMENT_TEMPLATE_224,
    MTCNNStage,
)


# Check if TensorFlow is available for MTCNN tests
def _has_tensorflow():
    """Check if TensorFlow is available."""
    try:
        import tensorflow
        return True
    except ImportError:
        return False


requires_tensorflow = pytest.mark.skipif(
    not _has_tensorflow(),
    reason="TensorFlow not installed - MTCNN requires TensorFlow"
)


# ==============================================================================
# Fixtures
# ==============================================================================


@pytest.fixture
def seed():
    """Fixed random seed for reproducibility."""
    return 42


@pytest.fixture
def np_random(seed):
    """NumPy random generator with fixed seed."""
    return np.random.RandomState(seed)


@pytest.fixture
def default_config():
    """Default MTCNN configuration."""
    return MTCNNConfig()


@pytest.fixture
def detector(default_config):
    """MTCNN detector with default config."""
    return MTCNNDetector(default_config)


@pytest.fixture
def sample_face_image(np_random) -> np.ndarray:
    """
    Create a synthetic face-like image for testing.
    
    This creates an image with basic face features that MTCNN
    might recognize in a test environment.
    """
    image = np.zeros((480, 640, 3), dtype=np.uint8)
    
    # Background
    image[:, :] = [200, 200, 200]
    
    # Face ellipse (skin tone)
    center = (320, 240)
    cv2.ellipse(image, center, (80, 100), 0, 0, 360, (180, 200, 220), -1)
    
    # Eyes
    left_eye = (280, 210)
    right_eye = (360, 210)
    cv2.circle(image, left_eye, 12, (255, 255, 255), -1)
    cv2.circle(image, right_eye, 12, (255, 255, 255), -1)
    cv2.circle(image, left_eye, 5, (50, 50, 50), -1)
    cv2.circle(image, right_eye, 5, (50, 50, 50), -1)
    
    # Nose
    nose = (320, 250)
    cv2.circle(image, nose, 8, (170, 190, 210), -1)
    
    # Mouth
    mouth_left = (290, 290)
    mouth_right = (350, 290)
    cv2.line(image, mouth_left, mouth_right, (150, 100, 100), 3)
    
    return image


@pytest.fixture
def sample_landmarks():
    """Sample five-point landmarks."""
    return FiveLandmarks(
        left_eye=(280.0, 210.0),
        right_eye=(360.0, 210.0),
        nose=(320.0, 250.0),
        mouth_left=(290.0, 290.0),
        mouth_right=(350.0, 290.0),
    )


@pytest.fixture
def sample_detection(sample_landmarks):
    """Sample MTCNN detection result."""
    return MTCNNDetection(
        bbox=(200, 140, 240, 280),
        confidence=0.98,
        landmarks=sample_landmarks,
    )


@pytest.fixture
def blank_image():
    """Blank test image."""
    return np.zeros((480, 640, 3), dtype=np.uint8)


@pytest.fixture
def small_image():
    """Very small image for edge case testing."""
    return np.zeros((10, 10, 3), dtype=np.uint8)


# ==============================================================================
# FiveLandmarks Tests
# ==============================================================================


class TestFiveLandmarks:
    """Tests for FiveLandmarks dataclass."""
    
    def test_landmarks_creation(self, sample_landmarks):
        """Test creating FiveLandmarks object."""
        assert sample_landmarks.left_eye == (280.0, 210.0)
        assert sample_landmarks.right_eye == (360.0, 210.0)
        assert sample_landmarks.nose == (320.0, 250.0)
        assert sample_landmarks.mouth_left == (290.0, 290.0)
        assert sample_landmarks.mouth_right == (350.0, 290.0)
    
    def test_landmarks_to_array(self, sample_landmarks):
        """Test conversion to numpy array."""
        arr = sample_landmarks.to_array()
        
        assert arr.shape == (5, 2)
        assert arr.dtype == np.float32
        np.testing.assert_array_equal(arr[0], [280.0, 210.0])
        np.testing.assert_array_equal(arr[1], [360.0, 210.0])
    
    def test_landmarks_from_array(self):
        """Test creation from numpy array."""
        arr = np.array([
            [100.0, 100.0],
            [150.0, 100.0],
            [125.0, 130.0],
            [110.0, 160.0],
            [140.0, 160.0],
        ], dtype=np.float32)
        
        landmarks = FiveLandmarks.from_array(arr)
        
        assert landmarks.left_eye == (100.0, 100.0)
        assert landmarks.right_eye == (150.0, 100.0)
        assert landmarks.nose == (125.0, 130.0)
    
    def test_landmarks_from_flat_array(self):
        """Test creation from flattened array."""
        arr = np.array([
            100.0, 100.0, 150.0, 100.0, 125.0, 130.0,
            110.0, 160.0, 140.0, 160.0
        ], dtype=np.float32)
        
        landmarks = FiveLandmarks.from_array(arr)
        
        assert landmarks.left_eye == (100.0, 100.0)
        assert landmarks.nose == (125.0, 130.0)
    
    def test_landmarks_to_dict(self, sample_landmarks):
        """Test conversion to dictionary."""
        d = sample_landmarks.to_dict()
        
        assert "left_eye" in d
        assert "right_eye" in d
        assert "nose" in d
        assert "mouth_left" in d
        assert "mouth_right" in d
        assert d["left_eye"] == (280.0, 210.0)
    
    def test_compute_eye_distance(self, sample_landmarks):
        """Test eye distance calculation."""
        distance = sample_landmarks.compute_eye_distance()
        
        # Distance between (280, 210) and (360, 210) = 80
        assert distance == pytest.approx(80.0)
    
    def test_compute_face_angle_horizontal(self, sample_landmarks):
        """Test face angle for horizontally aligned eyes."""
        angle = sample_landmarks.compute_face_angle()
        
        # Eyes are at same y-coordinate, so angle should be 0
        assert angle == pytest.approx(0.0, abs=0.01)
    
    def test_compute_face_angle_tilted(self):
        """Test face angle for tilted face."""
        # Create tilted landmarks (right eye 10 pixels higher)
        landmarks = FiveLandmarks(
            left_eye=(280.0, 220.0),
            right_eye=(360.0, 210.0),
            nose=(320.0, 250.0),
            mouth_left=(290.0, 290.0),
            mouth_right=(350.0, 290.0),
        )
        
        angle = landmarks.compute_face_angle()
        
        # Should be negative (tilted counter-clockwise)
        assert angle < 0
        # arctan2(-10, 80) ≈ -7.125 degrees
        assert angle == pytest.approx(-7.125, abs=0.1)


# ==============================================================================
# MTCNNDetection Tests
# ==============================================================================


class TestMTCNNDetection:
    """Tests for MTCNNDetection dataclass."""
    
    def test_detection_creation(self, sample_detection):
        """Test creating MTCNNDetection object."""
        assert sample_detection.bbox == (200, 140, 240, 280)
        assert sample_detection.confidence == 0.98
        assert sample_detection.landmarks is not None
    
    def test_detection_area_calculation(self, sample_detection):
        """Test automatic area calculation."""
        # Area = 240 * 280 = 67200
        assert sample_detection.area == 67200
    
    def test_detection_to_dict(self, sample_detection):
        """Test serialization to dict."""
        d = sample_detection.to_dict()
        
        assert d["bbox"] == (200, 140, 240, 280)
        assert d["confidence"] == 0.98
        assert "landmarks" in d
        assert d["area"] == 67200


# ==============================================================================
# MTCNNDetectionResult Tests
# ==============================================================================


class TestMTCNNDetectionResult:
    """Tests for MTCNNDetectionResult dataclass."""
    
    def test_empty_result(self):
        """Test empty detection result."""
        result = MTCNNDetectionResult()
        
        assert result.num_faces == 0
        assert result.has_face is False
        assert result.success is True
        assert result.get_largest_face() is None
    
    def test_result_with_faces(self, sample_detection):
        """Test result with detected faces."""
        result = MTCNNDetectionResult(faces=[sample_detection])
        
        assert result.num_faces == 1
        assert result.has_face is True
    
    def test_get_largest_face(self, sample_landmarks):
        """Test getting largest face."""
        small = MTCNNDetection(
            bbox=(100, 100, 50, 50),
            confidence=0.99,
            landmarks=sample_landmarks,
        )
        large = MTCNNDetection(
            bbox=(200, 200, 100, 100),
            confidence=0.90,
            landmarks=sample_landmarks,
        )
        
        result = MTCNNDetectionResult(faces=[small, large])
        largest = result.get_largest_face()
        
        assert largest is not None
        assert largest.area == 10000  # 100 * 100
    
    def test_get_highest_confidence_face(self, sample_landmarks):
        """Test getting highest confidence face."""
        low_conf = MTCNNDetection(
            bbox=(200, 200, 100, 100),
            confidence=0.85,
            landmarks=sample_landmarks,
        )
        high_conf = MTCNNDetection(
            bbox=(100, 100, 50, 50),
            confidence=0.99,
            landmarks=sample_landmarks,
        )
        
        result = MTCNNDetectionResult(faces=[low_conf, high_conf])
        best = result.get_highest_confidence_face()
        
        assert best is not None
        assert best.confidence == 0.99
    
    def test_result_to_dict(self, sample_detection):
        """Test serialization to dict."""
        result = MTCNNDetectionResult(
            faces=[sample_detection],
            success=True,
            total_time_ms=15.5,
            image_shape=(480, 640, 3),
        )
        
        d = result.to_dict()
        
        assert d["num_faces"] == 1
        assert d["success"] is True
        assert d["total_time_ms"] == 15.5
        assert d["image_shape"] == (480, 640, 3)


# ==============================================================================
# MTCNNConfig Tests
# ==============================================================================


class TestMTCNNConfig:
    """Tests for MTCNNConfig dataclass."""
    
    def test_default_config(self):
        """Test default configuration values."""
        config = MTCNNConfig()
        
        assert config.min_face_size == 20
        assert config.scale_factor == 0.709
        assert config.pnet_threshold == 0.6
        assert config.rnet_threshold == 0.7
        assert config.onet_threshold == 0.9
        assert config.target_size == (224, 224)
    
    def test_custom_config(self):
        """Test custom configuration."""
        config = MTCNNConfig(
            min_face_size=40,
            scale_factor=0.8,
            onet_threshold=0.95,
            target_size=(160, 160),
        )
        
        assert config.min_face_size == 40
        assert config.scale_factor == 0.8
        assert config.onet_threshold == 0.95
        assert config.target_size == (160, 160)


# ==============================================================================
# MTCNNDetector Tests
# ==============================================================================


class TestMTCNNDetector:
    """Tests for MTCNNDetector class."""
    
    def test_detector_initialization(self):
        """Test detector initialization."""
        detector = MTCNNDetector()
        
        assert detector.config is not None
        assert detector._initialized is False
    
    def test_detector_with_custom_config(self):
        """Test detector with custom config."""
        config = MTCNNConfig(min_face_size=40)
        detector = MTCNNDetector(config)
        
        assert detector.config.min_face_size == 40
    
    @requires_tensorflow
    def test_detect_empty_image(self, detector, blank_image):
        """Test detection on blank image."""
        result = detector.detect(blank_image)
        
        assert result.success is True
        assert result.num_faces == 0
    
    def test_detect_none_image(self, detector):
        """Test detection with None input."""
        result = detector.detect(None)
        
        assert result.success is False
        assert "Invalid" in result.error_message
    
    def test_detect_empty_array(self, detector):
        """Test detection with empty array."""
        result = detector.detect(np.array([]))
        
        assert result.success is False
    
    @requires_tensorflow
    def test_detect_returns_result_type(self, detector, sample_face_image):
        """Test that detect returns correct type."""
        result = detector.detect(sample_face_image)
        
        assert isinstance(result, MTCNNDetectionResult)
        assert result.total_time_ms >= 0
    
    @requires_tensorflow
    def test_detect_includes_timing(self, detector, sample_face_image):
        """Test that timing information is included."""
        result = detector.detect(sample_face_image)
        
        assert result.total_time_ms > 0
    
    @requires_tensorflow
    def test_detect_includes_image_shape(self, detector, sample_face_image):
        """Test that image shape is recorded."""
        result = detector.detect(sample_face_image)
        
        assert result.image_shape == sample_face_image.shape
    
    def test_compute_detection_hash(self, detector, sample_detection):
        """Test detection hash computation."""
        result = MTCNNDetectionResult(faces=[sample_detection])
        hash1 = detector.compute_detection_hash(result)
        hash2 = detector.compute_detection_hash(result)
        
        # Same result should produce same hash
        assert hash1 == hash2
        assert len(hash1) == 64  # SHA256 hex length


# ==============================================================================
# Face Alignment Tests
# ==============================================================================


class TestFaceAlignment:
    """Tests for face alignment functionality."""
    
    def test_align_returns_correct_shape(
        self, detector, sample_face_image, sample_detection
    ):
        """Test that align returns image of correct shape."""
        aligned = detector.align(sample_face_image, sample_detection)
        
        assert aligned.shape == (224, 224, 3)
    
    def test_align_custom_size(
        self, detector, sample_face_image, sample_detection
    ):
        """Test alignment with custom target size."""
        aligned = detector.align(
            sample_face_image, sample_detection, target_size=(160, 160)
        )
        
        assert aligned.shape == (160, 160, 3)
    
    @requires_tensorflow
    def test_detect_and_align_with_no_face(self, detector, blank_image):
        """Test detect_and_align when no face is found."""
        aligned, result = detector.detect_and_align(blank_image)
        
        assert aligned is None
        assert result.success is True
        assert result.num_faces == 0
    
    @requires_tensorflow
    def test_extract_landmarks_only_no_face(self, detector, blank_image):
        """Test landmark extraction when no face present."""
        landmarks = detector.extract_landmarks_only(blank_image)
        
        assert landmarks is None


# ==============================================================================
# FaceAligner Tests
# ==============================================================================


class TestFaceAligner:
    """Tests for standalone FaceAligner class."""
    
    def test_aligner_initialization(self):
        """Test aligner initialization."""
        aligner = FaceAligner()
        
        assert aligner.target_size == (224, 224)
        assert aligner.use_template is True
    
    def test_aligner_custom_size(self):
        """Test aligner with custom size."""
        aligner = FaceAligner(target_size=(160, 160))
        
        assert aligner.target_size == (160, 160)
    
    def test_align_with_landmarks_object(
        self, sample_face_image, sample_landmarks
    ):
        """Test alignment with FiveLandmarks object."""
        aligner = FaceAligner()
        aligned = aligner.align(sample_face_image, sample_landmarks)
        
        assert aligned.shape == (224, 224, 3)
    
    def test_align_with_array(self, sample_face_image):
        """Test alignment with numpy array landmarks."""
        aligner = FaceAligner()
        landmarks = np.array([
            [280.0, 210.0],
            [360.0, 210.0],
            [320.0, 250.0],
            [290.0, 290.0],
            [350.0, 290.0],
        ], dtype=np.float32)
        
        aligned = aligner.align(sample_face_image, landmarks)
        
        assert aligned.shape == (224, 224, 3)
    
    def test_alignment_quality_good(self, sample_landmarks):
        """Test alignment quality for well-positioned landmarks."""
        aligner = FaceAligner()
        quality = aligner.compute_alignment_quality(sample_landmarks)
        
        # Should be reasonable quality (0.5-1.0)
        assert 0.0 <= quality <= 1.0
    
    def test_alignment_quality_with_array(self):
        """Test alignment quality with numpy array."""
        aligner = FaceAligner()
        landmarks = np.array([
            [100.0, 100.0],
            [150.0, 100.0],
            [125.0, 130.0],
            [110.0, 160.0],
            [140.0, 160.0],
        ], dtype=np.float32)
        
        quality = aligner.compute_alignment_quality(landmarks)
        
        assert 0.0 <= quality <= 1.0


# ==============================================================================
# Template Tests
# ==============================================================================


class TestAlignmentTemplate:
    """Tests for alignment template."""
    
    def test_template_shape(self):
        """Test template has correct shape."""
        assert ALIGNMENT_TEMPLATE_224.shape == (5, 2)
    
    def test_template_range(self):
        """Test template values are in 0-1 range."""
        assert ALIGNMENT_TEMPLATE_224.min() >= 0.0
        assert ALIGNMENT_TEMPLATE_224.max() <= 1.0
    
    def test_template_symmetry(self):
        """Test template is roughly symmetric."""
        left_eye = ALIGNMENT_TEMPLATE_224[0]
        right_eye = ALIGNMENT_TEMPLATE_224[1]
        
        # Eyes should be at same y-coordinate
        assert left_eye[1] == pytest.approx(right_eye[1], abs=0.01)
        
        # Eyes should be symmetric around center (0.5)
        assert (left_eye[0] + right_eye[0]) / 2 == pytest.approx(0.5, abs=0.02)
    
    def test_template_mouth_symmetry(self):
        """Test mouth corners are symmetric."""
        mouth_left = ALIGNMENT_TEMPLATE_224[3]
        mouth_right = ALIGNMENT_TEMPLATE_224[4]
        
        # Mouth corners at same y-coordinate
        assert mouth_left[1] == pytest.approx(mouth_right[1], abs=0.01)


# ==============================================================================
# MTCNNStage Enum Tests
# ==============================================================================


class TestMTCNNStage:
    """Tests for MTCNNStage enum."""
    
    def test_stage_values(self):
        """Test stage enum values."""
        assert MTCNNStage.PNET.value == "pnet"
        assert MTCNNStage.RNET.value == "rnet"
        assert MTCNNStage.ONET.value == "onet"
    
    def test_stage_count(self):
        """Test there are exactly 3 stages."""
        stages = list(MTCNNStage)
        assert len(stages) == 3


# ==============================================================================
# Integration Tests
# ==============================================================================


class TestMTCNNIntegration:
    """Integration tests for MTCNN pipeline."""
    
    @requires_tensorflow
    def test_full_pipeline_synthetic_image(
        self, detector, sample_face_image
    ):
        """Test full detection pipeline with synthetic image."""
        # Run detection
        result = detector.detect(sample_face_image)
        
        # Should at least run without error
        assert result.success is True
        assert result.total_time_ms > 0
    
    @requires_tensorflow
    def test_detect_and_align_pipeline(
        self, detector, sample_face_image
    ):
        """Test combined detect and align pipeline."""
        aligned, result = detector.detect_and_align(sample_face_image)
        
        # Result should be valid
        assert result.success is True
        
        # If face was detected, aligned should be present
        if result.has_face:
            assert aligned is not None
            assert aligned.shape == (224, 224, 3)
    
    @requires_tensorflow
    def test_landmarks_extraction_pipeline(
        self, detector, sample_face_image
    ):
        """Test landmark extraction pipeline."""
        landmarks = detector.extract_landmarks_only(sample_face_image)
        
        # May or may not find face in synthetic image
        if landmarks is not None:
            assert isinstance(landmarks, FiveLandmarks)


# ==============================================================================
# Edge Case Tests
# ==============================================================================


class TestEdgeCases:
    """Tests for edge cases and error handling."""
    
    @requires_tensorflow
    def test_very_small_image(self, detector, small_image):
        """Test detection on very small image."""
        result = detector.detect(small_image)
        
        # Should handle gracefully
        assert result.success is True
        assert result.num_faces == 0
    
    @requires_tensorflow
    def test_grayscale_image(self, detector):
        """Test detection on grayscale image."""
        gray = np.zeros((480, 640), dtype=np.uint8)
        result = detector.detect(gray)
        
        # Should handle gracefully
        assert isinstance(result, MTCNNDetectionResult)
    
    @requires_tensorflow
    def test_rgba_image(self, detector):
        """Test detection on RGBA image."""
        rgba = np.zeros((480, 640, 4), dtype=np.uint8)
        result = detector.detect(rgba)
        
        # Should handle gracefully
        assert isinstance(result, MTCNNDetectionResult)
    
    @requires_tensorflow
    def test_high_resolution_image(self, detector):
        """Test detection on high resolution image."""
        large = np.zeros((2000, 3000, 3), dtype=np.uint8)
        result = detector.detect(large)
        
        # Should complete without error
        assert result.success is True
    
    def test_negative_bbox_handling(self, sample_landmarks):
        """Test handling of negative bbox values."""
        # MTCNN sometimes returns negative coordinates
        detection = MTCNNDetection(
            bbox=(-10, -5, 100, 100),
            confidence=0.95,
            landmarks=sample_landmarks,
        )
        
        # Area should still be calculated
        assert detection.area == 10000
    
    @requires_tensorflow
    def test_config_with_high_threshold(self, sample_face_image):
        """Test with very high confidence threshold."""
        config = MTCNNConfig(onet_threshold=0.999)
        detector = MTCNNDetector(config)
        
        result = detector.detect(sample_face_image)
        
        # Should filter out low-confidence detections
        assert result.success is True


# ==============================================================================
# Determinism Tests
# ==============================================================================


class TestDeterminism:
    """Tests for deterministic execution."""
    
    @requires_tensorflow
    def test_same_input_same_hash(self, detector, sample_face_image):
        """Test that same input produces same hash."""
        result1 = detector.detect(sample_face_image.copy())
        result2 = detector.detect(sample_face_image.copy())
        
        hash1 = detector.compute_detection_hash(result1)
        hash2 = detector.compute_detection_hash(result2)
        
        # Same input should produce same results
        assert hash1 == hash2
    
    def test_landmark_array_determinism(self, sample_landmarks):
        """Test landmark array conversion is deterministic."""
        arr1 = sample_landmarks.to_array()
        arr2 = sample_landmarks.to_array()
        
        np.testing.assert_array_equal(arr1, arr2)

    def test_alignment_determinism_fixed_input(self):
        """Test that alignment output is deterministic for fixed inputs."""
        # Synthetic deterministic image
        image = np.zeros((256, 256, 3), dtype=np.uint8)
        image[:, :] = [128, 128, 128]
        cv2.circle(image, (96, 112), 8, (255, 255, 255), -1)
        cv2.circle(image, (160, 112), 8, (255, 255, 255), -1)
        cv2.circle(image, (128, 144), 6, (200, 200, 200), -1)
        cv2.line(image, (104, 176), (152, 176), (180, 120, 120), 2)

        landmarks = FiveLandmarks(
            left_eye=(96.0, 112.0),
            right_eye=(160.0, 112.0),
            nose=(128.0, 144.0),
            mouth_left=(108.0, 176.0),
            mouth_right=(148.0, 176.0),
        )

        aligner = FaceAligner(target_size=(224, 224), use_template=True)
        aligned1 = aligner.align(image.copy(), landmarks)
        aligned2 = aligner.align(image.copy(), landmarks)

        np.testing.assert_array_equal(aligned1, aligned2)
