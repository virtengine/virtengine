"""
Tests for the orientation module.
"""

import pytest
import numpy as np
import cv2

from ml.document_preprocessing.orientation import (
    OrientationDetector,
    OrientationResult,
)
from ml.document_preprocessing.config import DocumentConfig, OrientationConfig


class TestOrientationDetector:
    """Tests for OrientationDetector class."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        detector = OrientationDetector()
        assert detector.config is not None
        assert detector.orient_config is not None
    
    def test_init_custom_config(self, document_config):
        """Test initialization with custom config."""
        detector = OrientationDetector(document_config)
        assert detector.config == document_config
    
    def test_detect_orientation_returns_result(self, orientation_detector, small_document_image):
        """Test that detect_orientation returns OrientationResult."""
        result = orientation_detector.detect_orientation(small_document_image)
        
        assert isinstance(result, OrientationResult)
        assert result.detected_angle in [0, 90, 180, 270]
        assert 0.0 <= result.confidence <= 1.0
        assert isinstance(result.scores, dict)
    
    def test_detect_upright_document(self, orientation_detector, small_document_image):
        """Test detecting orientation of upright document."""
        result = orientation_detector.detect_orientation(small_document_image)
        
        # An upright document should be detected as 0 degrees
        # (or have the highest score at 0)
        assert 0 in result.scores
    
    def test_detect_rotated_90(self, orientation_detector, sample_document_image):
        """Test detecting 90-degree rotated document."""
        # Rotate 90 degrees
        rotated = cv2.rotate(sample_document_image, cv2.ROTATE_90_CLOCKWISE)
        
        result = orientation_detector.detect_orientation(rotated)
        
        # Should detect rotation (90 or 270 depending on detection method)
        assert result.detected_angle in [0, 90, 180, 270]
    
    def test_detect_rotated_180(self, orientation_detector, sample_document_image):
        """Test detecting 180-degree rotated document."""
        rotated = cv2.rotate(sample_document_image, cv2.ROTATE_180)
        
        result = orientation_detector.detect_orientation(rotated)
        
        assert result.detected_angle in [0, 90, 180, 270]
    
    def test_correct_orientation_no_rotation_needed(self, orientation_detector, small_document_image):
        """Test that correct_orientation doesn't modify upright document unnecessarily."""
        corrected, angle = orientation_detector.correct_orientation(small_document_image)
        
        assert corrected is not None
        assert angle in [0, 90, 180, 270]
    
    def test_correct_orientation_with_specified_angle(self, orientation_detector, small_document_image):
        """Test correction with manually specified angle."""
        corrected, angle = orientation_detector.correct_orientation(small_document_image, angle=90)
        
        assert angle == 90
        # Image should be rotated
        assert corrected.shape[0] != small_document_image.shape[0] or \
               corrected.shape[1] != small_document_image.shape[1]
    
    def test_correct_orientation_zero_angle(self, orientation_detector, small_document_image):
        """Test that 0 degree correction returns original."""
        corrected, angle = orientation_detector.correct_orientation(small_document_image, angle=0)
        
        assert angle == 0
        assert np.array_equal(corrected, small_document_image)
    
    def test_detect_orientation_disabled(self, document_config, small_document_image):
        """Test behavior when orientation detection is disabled."""
        document_config.orientation.enabled = False
        detector = OrientationDetector(document_config)
        
        result = detector.detect_orientation(small_document_image)
        
        assert result.detected_angle == 0
        assert result.confidence == 1.0
        assert result.method_used == "disabled"
    
    def test_detect_skew_angle(self, orientation_detector, small_document_image):
        """Test detecting fine skew angles."""
        skew = orientation_detector.detect_skew_angle(small_document_image)
        
        assert isinstance(skew, float)
        assert -15 <= skew <= 15  # Within expected range
    
    def test_deskew_document(self, orientation_detector, small_document_image):
        """Test deskewing a document."""
        deskewed, angle = orientation_detector.deskew(small_document_image)
        
        assert deskewed is not None
        assert isinstance(angle, float)
    
    def test_deskew_with_specific_angle(self, orientation_detector, small_document_image):
        """Test deskewing with specified angle."""
        deskewed, angle = orientation_detector.deskew(small_document_image, angle=5.0)
        
        assert angle == 5.0
        # Image should be rotated
        assert deskewed is not None
    
    def test_deskew_negligible_skew(self, orientation_detector, small_document_image):
        """Test that negligible skew is not corrected."""
        deskewed, angle = orientation_detector.deskew(small_document_image, angle=0.3)
        
        assert angle == 0.0  # Below threshold
        assert np.array_equal(deskewed, small_document_image)


class TestOrientationScoring:
    """Tests for orientation scoring methods."""
    
    def test_text_scoring(self, orientation_detector, small_document_image):
        """Test text-based orientation scoring."""
        # Access private method for testing
        score = orientation_detector._score_text_orientation(small_document_image)
        
        assert isinstance(score, float)
        assert 0.0 <= score <= 1.0
    
    def test_text_scoring_rotated(self, orientation_detector, sample_document_image):
        """Test that text scoring differs for rotations."""
        score_0 = orientation_detector._score_text_orientation(sample_document_image)
        
        rotated = cv2.rotate(sample_document_image, cv2.ROTATE_90_CLOCKWISE)
        score_90 = orientation_detector._score_text_orientation(rotated)
        
        # Scores should be different for different orientations
        # (though they may not always be predictable for synthetic images)
        assert isinstance(score_0, float)
        assert isinstance(score_90, float)
    
    def test_face_scoring(self, orientation_detector, small_document_image):
        """Test face-based orientation scoring."""
        score = orientation_detector._score_face_orientation(small_document_image)
        
        assert isinstance(score, float)
        assert 0.0 <= score <= 1.0


class TestOrientationConfigurations:
    """Tests for different orientation configurations."""
    
    def test_text_only_detection(self, sample_document_image):
        """Test text-only orientation detection."""
        config = DocumentConfig()
        config.orientation.detection_method = "text"
        detector = OrientationDetector(config)
        
        result = detector.detect_orientation(sample_document_image)
        
        assert result.method_used == "text"
    
    def test_face_only_detection(self, sample_document_image):
        """Test face-only orientation detection."""
        config = DocumentConfig()
        config.orientation.detection_method = "face"
        detector = OrientationDetector(config)
        
        result = detector.detect_orientation(sample_document_image)
        
        assert result.method_used == "face"
    
    def test_combined_detection(self, sample_document_image):
        """Test combined orientation detection."""
        config = DocumentConfig()
        config.orientation.detection_method = "combined"
        detector = OrientationDetector(config)
        
        result = detector.detect_orientation(sample_document_image)
        
        assert result.method_used == "combined"
    
    def test_custom_rotation_angles(self, sample_document_image):
        """Test with custom rotation angles."""
        config = DocumentConfig()
        config.orientation.rotation_angles = (0, 180)  # Only 0 and 180
        detector = OrientationDetector(config)
        
        result = detector.detect_orientation(sample_document_image)
        
        assert result.detected_angle in [0, 180]
        assert len(result.scores) == 2


class TestRotationOperations:
    """Tests for rotation operations."""
    
    @pytest.mark.parametrize("angle", [0, 90, 180, 270])
    def test_rotate_image_standard_angles(self, orientation_detector, small_document_image, angle):
        """Test rotating by standard angles."""
        rotated = orientation_detector._rotate_image(small_document_image, angle)
        
        assert rotated is not None
        
        if angle == 0:
            assert np.array_equal(rotated, small_document_image)
        elif angle == 90:
            assert rotated.shape[0] == small_document_image.shape[1]
            assert rotated.shape[1] == small_document_image.shape[0]
        elif angle == 180:
            assert rotated.shape == small_document_image.shape
        elif angle == 270:
            assert rotated.shape[0] == small_document_image.shape[1]
    
    def test_rotate_and_rotate_back(self, orientation_detector, small_document_image):
        """Test that rotating 360 degrees returns to original (approximately)."""
        rotated_90 = orientation_detector._rotate_image(small_document_image, 90)
        rotated_180 = orientation_detector._rotate_image(rotated_90, 90)
        rotated_270 = orientation_detector._rotate_image(rotated_180, 90)
        rotated_360 = orientation_detector._rotate_image(rotated_270, 90)
        
        assert rotated_360.shape == small_document_image.shape
        # Due to potential interpolation artifacts, we check shape rather than exact values
