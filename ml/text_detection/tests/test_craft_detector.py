"""
Tests for CRAFT detector module.
"""

import pytest
import numpy as np
from unittest.mock import Mock, patch, MagicMock

from ml.text_detection.config import CRAFTConfig, DeviceType
from ml.text_detection.craft_detector import CRAFTDetector
from ml.text_detection.roi_types import TextROI, TextType


class TestCRAFTDetectorInit:
    """Tests for CRAFTDetector initialization."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        with patch.object(CRAFTDetector, '_load_model', return_value=Mock()):
            detector = CRAFTDetector()
            assert detector.config is not None
            assert detector.config.device == DeviceType.CPU
    
    def test_init_custom_config(self, craft_config):
        """Test initialization with custom config."""
        with patch.object(CRAFTDetector, '_load_model', return_value=Mock()):
            detector = CRAFTDetector(craft_config)
            assert detector.config == craft_config
    
    def test_model_version(self):
        """Test model version is set."""
        with patch.object(CRAFTDetector, '_load_model', return_value=Mock()):
            detector = CRAFTDetector()
            assert detector.model_version == "2.0.0"
    
    def test_deterministic_mode_enabled(self):
        """Test that deterministic mode is enabled by default."""
        config = CRAFTConfig(deterministic=True)
        with patch.object(CRAFTDetector, '_load_model', return_value=Mock()):
            with patch.object(CRAFTDetector, '_setup_deterministic') as mock_setup:
                detector = CRAFTDetector(config)
                mock_setup.assert_called_once()


class TestCRAFTDetectorDetect:
    """Tests for CRAFT detection functionality."""
    
    @pytest.fixture
    def mock_detector(self):
        """Create a detector with mocked model."""
        with patch.object(CRAFTDetector, '_load_model') as mock_load:
            mock_model = Mock()
            mock_result = Mock()
            mock_result.heatmaps = {
                'text_score_heatmap': np.random.rand(100, 150).astype(np.float32),
                'link_score_heatmap': np.random.rand(100, 150).astype(np.float32),
            }
            mock_model.detect_text = Mock(return_value=mock_result)
            mock_load.return_value = mock_model
            
            detector = CRAFTDetector()
            detector._model = mock_model
            return detector
    
    def test_detect_returns_score_maps(self, mock_detector, small_text_image):
        """Test that detect returns region and affinity score maps."""
        region_scores, affinity_scores = mock_detector.detect(small_text_image)
        
        assert isinstance(region_scores, np.ndarray)
        assert isinstance(affinity_scores, np.ndarray)
        assert region_scores.ndim == 2
        assert affinity_scores.ndim == 2
    
    def test_detect_invalid_image_none(self, mock_detector):
        """Test detect with None input."""
        with pytest.raises(ValueError, match="Invalid input image"):
            mock_detector.detect(None)
    
    def test_detect_invalid_image_empty(self, mock_detector):
        """Test detect with empty array."""
        with pytest.raises(ValueError, match="Invalid input image"):
            mock_detector.detect(np.array([]))
    
    def test_detect_grayscale_image(self, mock_detector, grayscale_text_image):
        """Test detect with grayscale input."""
        region_scores, affinity_scores = mock_detector.detect(grayscale_text_image)
        
        assert isinstance(region_scores, np.ndarray)
        assert isinstance(affinity_scores, np.ndarray)


class TestCRAFTDetectorBoundingBoxes:
    """Tests for bounding box extraction from score maps."""
    
    @pytest.fixture
    def detector_for_boxes(self):
        """Create detector for box extraction tests."""
        with patch.object(CRAFTDetector, '_load_model', return_value=Mock()):
            return CRAFTDetector()
    
    def test_get_bounding_boxes_empty_scores(self, detector_for_boxes):
        """Test with empty score maps."""
        empty_region = np.zeros((100, 100), dtype=np.float32)
        empty_affinity = np.zeros((100, 100), dtype=np.float32)
        
        boxes = detector_for_boxes.get_bounding_boxes(empty_region, empty_affinity)
        assert isinstance(boxes, list)
        assert len(boxes) == 0
    
    def test_get_bounding_boxes_returns_text_rois(
        self, detector_for_boxes, sample_region_scores, sample_affinity_scores
    ):
        """Test that boxes are returned as TextROI objects."""
        boxes = detector_for_boxes.get_bounding_boxes(
            sample_region_scores,
            sample_affinity_scores,
            text_threshold=0.5,
        )
        
        assert isinstance(boxes, list)
        for box in boxes:
            assert isinstance(box, TextROI)
            assert box.text_type == TextType.CHARACTER
    
    def test_get_bounding_boxes_respects_threshold(
        self, detector_for_boxes, sample_region_scores, sample_affinity_scores
    ):
        """Test that threshold affects number of detections."""
        boxes_low_thresh = detector_for_boxes.get_bounding_boxes(
            sample_region_scores,
            sample_affinity_scores,
            text_threshold=0.3,
        )
        
        boxes_high_thresh = detector_for_boxes.get_bounding_boxes(
            sample_region_scores,
            sample_affinity_scores,
            text_threshold=0.9,
        )
        
        # Higher threshold should produce fewer or equal boxes
        assert len(boxes_high_thresh) <= len(boxes_low_thresh)
    
    def test_get_bounding_boxes_has_scores(
        self, detector_for_boxes, sample_region_scores, sample_affinity_scores
    ):
        """Test that returned boxes have valid scores."""
        boxes = detector_for_boxes.get_bounding_boxes(
            sample_region_scores,
            sample_affinity_scores,
            text_threshold=0.5,
        )
        
        for box in boxes:
            assert 0.0 <= box.confidence <= 1.0
            assert 0.0 <= box.region_score <= 1.0
            assert 0.0 <= box.affinity_score <= 1.0


class TestCRAFTDetectorReproducibility:
    """Tests for reproducibility and determinism."""
    
    @pytest.fixture
    def mock_detector_reproducible(self):
        """Create detector with reproducible mock output."""
        with patch.object(CRAFTDetector, '_load_model') as mock_load:
            mock_model = Mock()
            
            # Fixed score maps for reproducibility
            np.random.seed(42)
            fixed_region = np.random.rand(100, 150).astype(np.float32)
            fixed_affinity = np.random.rand(100, 150).astype(np.float32)
            
            mock_result = Mock()
            mock_result.heatmaps = {
                'text_score_heatmap': fixed_region,
                'link_score_heatmap': fixed_affinity,
            }
            mock_model.detect_text = Mock(return_value=mock_result)
            mock_load.return_value = mock_model
            
            detector = CRAFTDetector(CRAFTConfig(deterministic=True))
            detector._model = mock_model
            return detector
    
    def test_same_input_same_output(self, mock_detector_reproducible, small_text_image):
        """Test that same input produces same output."""
        result1 = mock_detector_reproducible.detect(small_text_image.copy())
        result2 = mock_detector_reproducible.detect(small_text_image.copy())
        
        np.testing.assert_array_equal(result1[0], result2[0])
        np.testing.assert_array_equal(result1[1], result2[1])
    
    def test_model_version_recorded(self, mock_detector_reproducible):
        """Test that model version is accessible."""
        assert mock_detector_reproducible.model_version == "2.0.0"


class TestCRAFTDetectorCleanup:
    """Tests for model cleanup and memory management."""
    
    def test_unload_clears_model(self):
        """Test that unload clears the model."""
        with patch.object(CRAFTDetector, '_load_model', return_value=Mock()):
            detector = CRAFTDetector()
            detector._model = Mock()
            
            detector.unload()
            
            assert detector._model is None
    
    def test_unload_clears_hash(self):
        """Test that unload clears the model hash."""
        with patch.object(CRAFTDetector, '_load_model', return_value=Mock()):
            detector = CRAFTDetector()
            detector._model_hash = "test_hash"
            
            detector.unload()
            
            assert detector._model_hash is None
