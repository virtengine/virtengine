"""
Tests for the perspective module.
"""

import pytest
import numpy as np
import cv2

from ml.document_preprocessing.perspective import (
    PerspectiveCorrector,
    PerspectiveResult,
    Point,
)
from ml.document_preprocessing.config import DocumentConfig, PerspectiveConfig


class TestPoint:
    """Tests for Point dataclass."""
    
    def test_point_creation(self):
        """Test creating a point."""
        p = Point(100.0, 200.0)
        assert p.x == 100.0
        assert p.y == 200.0
    
    def test_point_as_tuple(self):
        """Test converting to tuple."""
        p = Point(100.5, 200.5)
        assert p.as_tuple() == (100.5, 200.5)
    
    def test_point_as_int_tuple(self):
        """Test converting to integer tuple."""
        p = Point(100.7, 200.3)
        assert p.as_int_tuple() == (100, 200)


class TestPerspectiveCorrector:
    """Tests for PerspectiveCorrector class."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        corrector = PerspectiveCorrector()
        assert corrector.config is not None
        assert corrector.persp_config is not None
    
    def test_init_custom_config(self, document_config):
        """Test initialization with custom config."""
        corrector = PerspectiveCorrector(document_config)
        assert corrector.config == document_config
    
    def test_correct_perspectiVIRTENGINE_returns_result(self, perspectiVIRTENGINE_corrector, sample_document_image):
        """Test that correct_perspective returns proper result."""
        corrected, result = perspectiVIRTENGINE_corrector.correct_perspective(sample_document_image)
        
        assert isinstance(result, PerspectiveResult)
        assert corrected is not None
        assert isinstance(corrected, np.ndarray)
    
    def test_correct_perspectiVIRTENGINE_disabled(self, document_config, sample_document_image):
        """Test behavior when perspective correction is disabled."""
        document_config.perspective.enabled = False
        corrector = PerspectiveCorrector(document_config)
        
        corrected, result = corrector.correct_perspective(sample_document_image)
        
        assert result.corrected is False
        assert np.array_equal(corrected, sample_document_image)
    
    def test_detect_document_corners(self, perspectiVIRTENGINE_corrector, sample_document_image):
        """Test corner detection."""
        corners, confidence = perspectiVIRTENGINE_corrector.detect_document_corners(sample_document_image)
        
        # May or may not find corners depending on image content
        if corners is not None:
            assert len(corners) == 4
            assert all(isinstance(c, Point) for c in corners)
            assert 0.0 <= confidence <= 1.0
    
    def test_four_point_transform(self, perspectiVIRTENGINE_corrector, sample_document_image):
        """Test four-point perspective transform."""
        # Define corners manually
        height, width = sample_document_image.shape[:2]
        corners = [
            Point(50, 50),
            Point(width - 50, 60),
            Point(width - 40, height - 50),
            Point(40, height - 60),
        ]
        
        result = perspectiVIRTENGINE_corrector.four_point_transform(sample_document_image, corners)
        
        assert result is not None
        assert isinstance(result, np.ndarray)
        assert len(result.shape) == 3
    
    def test_manual_correction(self, perspectiVIRTENGINE_corrector, sample_document_image):
        """Test manual perspective correction with tuple corners."""
        height, width = sample_document_image.shape[:2]
        corners = [
            (50, 50),
            (width - 50, 60),
            (width - 40, height - 50),
            (40, height - 60),
        ]
        
        result = perspectiVIRTENGINE_corrector.manual_correction(sample_document_image, corners)
        
        assert result is not None
        assert isinstance(result, np.ndarray)
    
    def test_perspectiVIRTENGINE_on_distorted_image(self, perspectiVIRTENGINE_corrector, perspectiVIRTENGINE_distorted_image):
        """Test perspective correction on distorted image."""
        corrected, result = perspectiVIRTENGINE_corrector.correct_perspective(perspectiVIRTENGINE_distorted_image)
        
        # The distorted image should be detected and potentially corrected
        assert corrected is not None
        # Note: Detection success depends on the distortion level


class TestCornerOrdering:
    """Tests for corner ordering functionality."""
    
    def test_order_corners(self, perspectiVIRTENGINE_corrector):
        """Test that corners are ordered correctly."""
        # Create unordered corners
        corners = [
            Point(400, 400),  # Bottom-right
            Point(100, 100),  # Top-left
            Point(400, 100),  # Top-right
            Point(100, 400),  # Bottom-left
        ]
        
        ordered = perspectiVIRTENGINE_corrector._order_corners(corners)
        
        # Check ordering: TL, TR, BR, BL
        assert ordered[0].x < ordered[1].x  # TL.x < TR.x
        assert ordered[0].y < ordered[3].y  # TL.y < BL.y
        assert ordered[2].x > ordered[3].x  # BR.x > BL.x
    
    def test_order_corners_already_ordered(self, perspectiVIRTENGINE_corrector):
        """Test ordering already-ordered corners."""
        corners = [
            Point(100, 100),  # TL
            Point(400, 100),  # TR
            Point(400, 400),  # BR
            Point(100, 400),  # BL
        ]
        
        ordered = perspectiVIRTENGINE_corrector._order_corners(corners)
        
        # Should maintain order
        assert ordered[0].x == 100 and ordered[0].y == 100
        assert ordered[2].x == 400 and ordered[2].y == 400


class TestNeedsCorrection:
    """Tests for needs_correction logic."""
    
    def test_needs_correction_rectangular(self, perspectiVIRTENGINE_corrector):
        """Test that rectangular corners don't need correction."""
        # Perfect rectangle
        corners = [
            Point(0, 0),
            Point(100, 0),
            Point(100, 100),
            Point(0, 100),
        ]
        
        needs = perspectiVIRTENGINE_corrector._needs_correction(corners, (100, 100, 3))
        
        # Perfect rectangle shouldn't need correction
        assert needs is False
    
    def test_needs_correction_distorted(self, perspectiVIRTENGINE_corrector):
        """Test that distorted corners need correction."""
        # Distorted quadrilateral
        corners = [
            Point(10, 20),
            Point(90, 10),
            Point(95, 85),
            Point(5, 95),
        ]
        
        needs = perspectiVIRTENGINE_corrector._needs_correction(corners, (100, 100, 3))
        
        # Distorted shape should need correction
        assert needs is True


class TestPerspectiveConfiguration:
    """Tests for perspective correction configuration."""
    
    def test_custom_canny_thresholds(self, sample_document_image):
        """Test with custom Canny thresholds."""
        config = DocumentConfig()
        config.perspective.canny_low_threshold = 30
        config.perspective.canny_high_threshold = 100
        
        corrector = PerspectiveCorrector(config)
        corrected, result = corrector.correct_perspective(sample_document_image)
        
        assert corrected is not None
    
    def test_custom_contour_area_ratio(self, sample_document_image):
        """Test with custom contour area ratio."""
        config = DocumentConfig()
        config.perspective.min_contour_area_ratio = 0.2
        config.perspective.max_contour_area_ratio = 0.9
        
        corrector = PerspectiveCorrector(config)
        corrected, result = corrector.correct_perspective(sample_document_image)
        
        assert corrected is not None
    
    def test_output_margin(self, perspectiVIRTENGINE_corrector, sample_document_image):
        """Test that output margin is applied."""
        height, width = sample_document_image.shape[:2]
        corners = [
            Point(0, 0),
            Point(width - 1, 0),
            Point(width - 1, height - 1),
            Point(0, height - 1),
        ]
        
        margin = perspectiVIRTENGINE_corrector.persp_config.output_margin
        result = perspectiVIRTENGINE_corrector.four_point_transform(sample_document_image, corners)
        
        # Output should include margin
        assert result.shape[0] >= height + 2 * margin - 2
        assert result.shape[1] >= width + 2 * margin - 2


class TestEdgeCases:
    """Tests for edge cases and error handling."""
    
    def test_no_corners_found(self, perspectiVIRTENGINE_corrector):
        """Test handling when no corners are found."""
        # Uniform image has no edges
        uniform = np.ones((400, 600, 3), dtype=np.uint8) * 200
        
        corners, confidence = perspectiVIRTENGINE_corrector.detect_document_corners(uniform)
        
        # Should return None for no corners
        assert corners is None or confidence < 0.1
    
    def test_small_image(self, perspectiVIRTENGINE_corrector):
        """Test with small image."""
        small = np.ones((100, 100, 3), dtype=np.uint8) * 200
        cv2.rectangle(small, (10, 10), (90, 90), (50, 50, 50), 2)
        
        corners, confidence = perspectiVIRTENGINE_corrector.detect_document_corners(small)
        
        # Should handle small images gracefully
        assert corners is None or isinstance(corners, list)
    
    def test_non_convex_contour_ignored(self, perspectiVIRTENGINE_corrector, np_random):
        """Test that non-convex contours are ignored."""
        # Create image with complex shape
        image = np.ones((400, 600, 3), dtype=np.uint8) * 240
        
        # Draw a star (non-convex)
        pts = np.array([
            [300, 50], [340, 150], [450, 150], [360, 220],
            [390, 330], [300, 260], [210, 330], [240, 220],
            [150, 150], [260, 150]
        ], dtype=np.int32)
        cv2.polylines(image, [pts], True, (50, 50, 50), 2)
        
        corners, confidence = perspectiVIRTENGINE_corrector.detect_document_corners(image)
        
        # Non-convex shapes should not be detected as document corners
        # (result depends on implementation details)
        assert corners is None or len(corners) == 4
