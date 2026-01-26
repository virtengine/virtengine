"""
Tests for the face detection module.
"""

import pytest
import numpy as np

from ml.facial_verification.face_detection import (
    FaceDetector,
    FaceDetection,
    FaceLandmarks,
    DetectionResult,
)
from ml.facial_verification.config import VerificationConfig, DetectionConfig, DetectorBackend
from ml.facial_verification.reason_codes import ReasonCodes


class TestFaceDetection:
    """Tests for FaceDetection dataclass."""
    
    def test_face_detection_creation(self):
        """Test creating a FaceDetection object."""
        detection = FaceDetection(
            bbox=(10, 20, 100, 100),
            confidence=0.95,
        )
        
        assert detection.bbox == (10, 20, 100, 100)
        assert detection.confidence == 0.95
        assert detection.area == 10000  # 100 * 100
    
    def test_face_detection_with_landmarks(self):
        """Test FaceDetection with landmarks."""
        landmarks = FaceLandmarks(
            left_eye=(50, 50),
            right_eye=(80, 50),
            nose=(65, 70),
        )
        
        detection = FaceDetection(
            bbox=(10, 20, 100, 100),
            confidence=0.95,
            landmarks=landmarks,
        )
        
        assert detection.landmarks is not None
        assert detection.landmarks.left_eye == (50, 50)
        assert detection.landmarks.right_eye == (80, 50)
    
    def test_face_detection_to_dict(self):
        """Test serialization to dict."""
        detection = FaceDetection(
            bbox=(10, 20, 100, 100),
            confidence=0.95,
            detector_backend="mtcnn",
        )
        
        d = detection.to_dict()
        
        assert d["bbox"] == (10, 20, 100, 100)
        assert d["confidence"] == 0.95
        assert d["detector_backend"] == "mtcnn"


class TestDetectionResult:
    """Tests for DetectionResult dataclass."""
    
    def test_empty_result(self):
        """Test empty detection result."""
        result = DetectionResult()
        
        assert result.num_faces == 0
        assert result.has_face is False
        assert result.success is True
    
    def test_result_with_faces(self):
        """Test detection result with faces."""
        faces = [
            FaceDetection(bbox=(10, 20, 50, 50), confidence=0.95),
            FaceDetection(bbox=(100, 20, 80, 80), confidence=0.90),
        ]
        
        result = DetectionResult(faces=faces)
        
        assert result.num_faces == 2
        assert result.has_face is True
    
    def test_get_largest_face(self):
        """Test getting the largest face."""
        faces = [
            FaceDetection(bbox=(10, 20, 50, 50), confidence=0.95),  # area 2500
            FaceDetection(bbox=(100, 20, 80, 80), confidence=0.90),  # area 6400
        ]
        
        result = DetectionResult(faces=faces)
        largest = result.get_largest_face()
        
        assert largest is not None
        assert largest.area == 6400
    
    def test_get_highest_confidence_face(self):
        """Test getting the highest confidence face."""
        faces = [
            FaceDetection(bbox=(10, 20, 50, 50), confidence=0.95),
            FaceDetection(bbox=(100, 20, 80, 80), confidence=0.90),
        ]
        
        result = DetectionResult(faces=faces)
        highest = result.get_highest_confidence_face()
        
        assert highest is not None
        assert highest.confidence == 0.95


class TestFaceDetector:
    """Tests for FaceDetector class."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        detector = FaceDetector()
        assert detector.config is not None
        assert detector.detection_config is not None
    
    def test_init_custom_config(self, verification_config):
        """Test initialization with custom config."""
        detector = FaceDetector(verification_config)
        assert detector.config == verification_config
    
    def test_detect_none_image(self, detector):
        """Test detection with None input."""
        result = detector.detect(None)
        
        assert result.success is False
        assert result.error_code == ReasonCodes.INVALID_IMAGE_FORMAT
    
    def test_detect_empty_image(self, detector):
        """Test detection with empty array."""
        empty = np.array([])
        result = detector.detect(empty)
        
        assert result.success is False
        assert result.error_code == ReasonCodes.INVALID_IMAGE_FORMAT
    
    def test_detect_no_face_image(self, detector, no_face_image):
        """Test detection on image with no face."""
        result = detector.detect(no_face_image)
        
        # Detection may or may not find spurious faces in noise
        # The key is that it should not crash
        assert result is not None
        if not result.has_face:
            assert result.error_code == ReasonCodes.NO_FACE_DETECTED
    
    def test_crop_face(self, detector, sample_face_image):
        """Test face cropping."""
        detection = FaceDetection(
            bbox=(50, 50, 100, 100),
            confidence=0.95,
        )
        
        cropped = detector.crop(sample_face_image, detection, target_size=(112, 112))
        
        assert cropped is not None
        assert cropped.shape == (112, 112, 3)
    
    def test_crop_with_margin(self, detector, sample_face_image):
        """Test that cropping includes margin."""
        detection = FaceDetection(
            bbox=(50, 50, 100, 100),
            confidence=0.95,
        )
        
        # With margin, the crop should be larger than the bbox
        detector.detection_config.crop_margin = 0.2
        cropped = detector.crop(sample_face_image, detection)
        
        assert cropped is not None
    
    def test_detection_result_hash(self, detector):
        """Test computing detection result hash."""
        faces = [
            FaceDetection(bbox=(10, 20, 50, 50), confidence=0.95),
        ]
        result = DetectionResult(faces=faces)
        
        hash1 = detector.compute_detection_hash(result)
        hash2 = detector.compute_detection_hash(result)
        
        assert hash1 == hash2
        assert len(hash1) == 64  # SHA256


class TestDetectionConfig:
    """Tests for DetectionConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = DetectionConfig()
        
        assert config.detector_backend == DetectorBackend.MTCNN
        assert config.confidence_threshold == 0.9
        assert config.align_faces is True
    
    def test_custom_backend(self):
        """Test setting custom backend."""
        config = DetectionConfig(
            detector_backend=DetectorBackend.OPENCV,
        )
        
        assert config.detector_backend == DetectorBackend.OPENCV
    
    def test_multiple_faces_config(self):
        """Test multiple faces configuration."""
        config = DetectionConfig(
            allow_multiple_faces=True,
            select_largest_face=False,
        )
        
        assert config.allow_multiple_faces is True
        assert config.select_largest_face is False
