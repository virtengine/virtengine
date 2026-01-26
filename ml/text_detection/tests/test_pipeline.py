"""
Tests for the main text detection pipeline.
"""

import pytest
import numpy as np
from unittest.mock import Mock, patch, MagicMock
import time

from ml.text_detection.pipeline import TextDetectionPipeline, SUITE_VERSION
from ml.text_detection.config import TextDetectionConfig, CRAFTConfig, PostProcessingConfig
from ml.text_detection.roi_types import TextDetectionResult, TextROI, TextType, BoundingBox


class TestTextDetectionPipelineInit:
    """Tests for TextDetectionPipeline initialization."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        with patch('ml.text_detection.craft_detector.CRAFTDetector._load_model', return_value=Mock()):
            pipeline = TextDetectionPipeline()
            assert pipeline.config is not None
            assert pipeline.detector is not None
            assert pipeline.postprocessor is not None
    
    def test_init_custom_config(self, text_detection_config):
        """Test initialization with custom config."""
        with patch('ml.text_detection.craft_detector.CRAFTDetector._load_model', return_value=Mock()):
            pipeline = TextDetectionPipeline(text_detection_config)
            assert pipeline.config == text_detection_config


class TestTextDetectionPipelineDetect:
    """Tests for detection functionality."""
    
    @pytest.fixture
    def mock_pipeline(self, text_detection_config):
        """Create a pipeline with mocked detector."""
        with patch('ml.text_detection.craft_detector.CRAFTDetector._load_model', return_value=Mock()):
            pipeline = TextDetectionPipeline(text_detection_config)
            
            # Mock detector methods
            mock_region = np.random.rand(100, 150).astype(np.float32)
            mock_affinity = np.random.rand(100, 150).astype(np.float32)
            pipeline.detector.detect = Mock(return_value=(mock_region, mock_affinity))
            
            # Create mock ROIs
            mock_rois = [
                TextROI.create(
                    bounding_box=BoundingBox(x=10, y=10, width=20, height=25),
                    confidence=0.9,
                    text_type=TextType.CHARACTER,
                ),
                TextROI.create(
                    bounding_box=BoundingBox(x=35, y=10, width=18, height=25),
                    confidence=0.85,
                    text_type=TextType.CHARACTER,
                ),
            ]
            pipeline.detector.get_bounding_boxes = Mock(return_value=mock_rois)
            
            return pipeline
    
    def test_detect_valid_image(self, mock_pipeline, small_text_image):
        """Test detection on a valid image."""
        result = mock_pipeline.detect(small_text_image)
        
        assert isinstance(result, TextDetectionResult)
        assert result.success is True
        assert result.error_message is None
    
    def test_detect_returns_rois(self, mock_pipeline, small_text_image):
        """Test that detection returns ROIs."""
        result = mock_pipeline.detect(small_text_image)
        
        assert isinstance(result.rois, list)
    
    def test_detect_records_metadata(self, mock_pipeline, small_text_image):
        """Test that metadata is recorded."""
        result = mock_pipeline.detect(small_text_image)
        
        assert result.model_version is not None
        assert result.suite_version == SUITE_VERSION
        assert result.processing_time_ms > 0
        assert result.thresholds_used is not None
    
    def test_detect_computes_hash(self, mock_pipeline, small_text_image):
        """Test that image hash is computed."""
        result = mock_pipeline.detect(small_text_image, compute_hash=True)
        
        assert result.image_hash is not None
        assert len(result.image_hash) == 64  # SHA256 hex length
    
    def test_detect_hash_disabled(self, mock_pipeline, small_text_image):
        """Test that hash can be disabled."""
        result = mock_pipeline.detect(small_text_image, compute_hash=False)
        
        assert result.image_hash == ""
    
    def test_detect_records_image_size(self, mock_pipeline, small_text_image):
        """Test that image size is recorded."""
        result = mock_pipeline.detect(small_text_image)
        
        height, width = small_text_image.shape[:2]
        assert result.image_size == (width, height)
    
    def test_detect_none_image(self, mock_pipeline):
        """Test detection with None input."""
        result = mock_pipeline.detect(None)
        
        assert result.success is False
        assert "None" in result.error_message
    
    def test_detect_empty_image(self, mock_pipeline):
        """Test detection with empty array."""
        result = mock_pipeline.detect(np.array([]))
        
        assert result.success is False
        assert "empty" in result.error_message
    
    def test_detect_invalid_type(self, mock_pipeline):
        """Test detection with invalid input type."""
        result = mock_pipeline.detect("not an array")
        
        assert result.success is False
        assert "numpy array" in result.error_message


class TestTextDetectionPipelineFilters:
    """Tests for filtered detection methods."""
    
    @pytest.fixture
    def mock_pipeline_with_levels(self, text_detection_config):
        """Create pipeline with all ROI levels."""
        with patch('ml.text_detection.craft_detector.CRAFTDetector._load_model', return_value=Mock()):
            pipeline = TextDetectionPipeline(text_detection_config)
            
            mock_region = np.random.rand(100, 150).astype(np.float32)
            mock_affinity = np.random.rand(100, 150).astype(np.float32)
            pipeline.detector.detect = Mock(return_value=(mock_region, mock_affinity))
            
            # Mock character ROIs
            mock_chars = [
                TextROI.create(
                    bounding_box=BoundingBox(x=10, y=10, width=15, height=20),
                    confidence=0.9,
                    text_type=TextType.CHARACTER,
                ),
                TextROI.create(
                    bounding_box=BoundingBox(x=28, y=10, width=15, height=20),
                    confidence=0.85,
                    text_type=TextType.CHARACTER,
                ),
            ]
            pipeline.detector.get_bounding_boxes = Mock(return_value=mock_chars)
            
            return pipeline
    
    def test_detect_characters_only(self, mock_pipeline_with_levels, small_text_image):
        """Test character-only detection."""
        result = mock_pipeline_with_levels.detect_characters_only(small_text_image)
        
        assert result.success is True
        # Should only have character type ROIs
        for roi in result.rois:
            assert roi.text_type == TextType.CHARACTER
    
    def test_detect_words_only(self, mock_pipeline_with_levels, small_text_image):
        """Test word-only detection."""
        result = mock_pipeline_with_levels.detect_words_only(small_text_image)
        
        assert result.success is True
        # Should only have word type ROIs (or empty if no merging)
        for roi in result.rois:
            assert roi.text_type == TextType.WORD
    
    def test_detect_lines_only(self, mock_pipeline_with_levels, small_text_image):
        """Test line-only detection."""
        result = mock_pipeline_with_levels.detect_lines_only(small_text_image)
        
        assert result.success is True
        # Should only have line type ROIs (or empty if no grouping)
        for roi in result.rois:
            assert roi.text_type == TextType.LINE


class TestTextDetectionPipelineReproducibility:
    """Tests for reproducibility and determinism."""
    
    @pytest.fixture
    def deterministic_pipeline(self, text_detection_config):
        """Create a deterministic pipeline."""
        text_detection_config.craft.deterministic = True
        
        with patch('ml.text_detection.craft_detector.CRAFTDetector._load_model', return_value=Mock()):
            pipeline = TextDetectionPipeline(text_detection_config)
            
            # Fixed output for reproducibility
            np.random.seed(42)
            fixed_region = np.random.rand(100, 150).astype(np.float32)
            fixed_affinity = np.random.rand(100, 150).astype(np.float32)
            
            pipeline.detector.detect = Mock(return_value=(fixed_region, fixed_affinity))
            
            fixed_rois = [
                TextROI.create(
                    bounding_box=BoundingBox(x=10, y=10, width=20, height=25),
                    confidence=0.9,
                    text_type=TextType.CHARACTER,
                    region_score=0.85,
                    affinity_score=0.6,
                ),
            ]
            pipeline.detector.get_bounding_boxes = Mock(return_value=fixed_rois)
            
            return pipeline
    
    def test_same_input_same_hash(self, deterministic_pipeline, small_text_image):
        """Test that same input produces same hash."""
        result1 = deterministic_pipeline.detect(small_text_image.copy(), compute_hash=True)
        result2 = deterministic_pipeline.detect(small_text_image.copy(), compute_hash=True)
        
        assert result1.image_hash == result2.image_hash
    
    def test_thresholds_recorded(self, deterministic_pipeline, small_text_image):
        """Test that thresholds are recorded for reproducibility."""
        result = deterministic_pipeline.detect(small_text_image)
        
        assert "text_threshold" in result.thresholds_used
        assert "link_threshold" in result.thresholds_used
        assert "nms_iou_threshold" in result.thresholds_used
    
    def test_model_version_recorded(self, deterministic_pipeline, small_text_image):
        """Test that model version is recorded."""
        result = deterministic_pipeline.detect(small_text_image)
        
        assert result.model_version is not None
        assert result.model_version == "2.0.0"
    
    def test_suite_version_recorded(self, deterministic_pipeline, small_text_image):
        """Test that suite version is recorded."""
        result = deterministic_pipeline.detect(small_text_image)
        
        assert result.suite_version == SUITE_VERSION


class TestTextDetectionPipelineScoreMaps:
    """Tests for score map access."""
    
    @pytest.fixture
    def pipeline_with_scores(self, text_detection_config):
        """Create pipeline that returns score maps."""
        with patch('ml.text_detection.craft_detector.CRAFTDetector._load_model', return_value=Mock()):
            pipeline = TextDetectionPipeline(text_detection_config)
            
            mock_region = np.random.rand(100, 150).astype(np.float32)
            mock_affinity = np.random.rand(100, 150).astype(np.float32)
            pipeline.detector.detect = Mock(return_value=(mock_region, mock_affinity))
            pipeline.detector.get_bounding_boxes = Mock(return_value=[])
            
            return pipeline
    
    def test_get_score_maps(self, pipeline_with_scores, small_text_image):
        """Test getting raw score maps."""
        region_scores, affinity_scores = pipeline_with_scores.get_score_maps(small_text_image)
        
        assert isinstance(region_scores, np.ndarray)
        assert isinstance(affinity_scores, np.ndarray)
    
    def test_score_maps_in_result(self, pipeline_with_scores, small_text_image):
        """Test that score maps are included in result."""
        result = pipeline_with_scores.detect(small_text_image)
        
        assert result.region_score_map is not None
        assert result.affinity_score_map is not None


class TestTextDetectionResultSerialization:
    """Tests for result serialization."""
    
    def test_to_dict(self):
        """Test result to_dict method."""
        result = TextDetectionResult(
            image_hash="abc123",
            rois=[
                TextROI.create(
                    bounding_box=BoundingBox(x=10, y=10, width=20, height=25),
                    confidence=0.9,
                    text_type=TextType.CHARACTER,
                ),
            ],
            model_version="2.0.0",
            processing_time_ms=100.0,
            thresholds_used={"text_threshold": 0.7},
            image_size=(400, 300),
        )
        
        d = result.to_dict()
        
        assert d["success"] is True
        assert d["image_hash"] == "abc123"
        assert d["model_version"] == "2.0.0"
        assert d["roi_count"] == 1
        assert "rois" in d
    
    def test_roi_counts_by_type(self):
        """Test ROI counts by type in serialization."""
        result = TextDetectionResult(
            image_hash="abc123",
            rois=[
                TextROI.create(
                    bounding_box=BoundingBox(x=10, y=10, width=20, height=25),
                    confidence=0.9,
                    text_type=TextType.CHARACTER,
                ),
                TextROI.create(
                    bounding_box=BoundingBox(x=10, y=10, width=60, height=25),
                    confidence=0.85,
                    text_type=TextType.WORD,
                ),
            ],
            model_version="2.0.0",
            processing_time_ms=100.0,
            thresholds_used={},
        )
        
        d = result.to_dict()
        
        assert d["roi_counts_by_type"]["character"] == 1
        assert d["roi_counts_by_type"]["word"] == 1


class TestTextDetectionPipelineCleanup:
    """Tests for cleanup and memory management."""
    
    def test_unload(self, text_detection_config):
        """Test unload method."""
        with patch('ml.text_detection.craft_detector.CRAFTDetector._load_model', return_value=Mock()):
            pipeline = TextDetectionPipeline(text_detection_config)
            pipeline.detector.unload = Mock()
            
            pipeline.unload()
            
            pipeline.detector.unload.assert_called_once()
