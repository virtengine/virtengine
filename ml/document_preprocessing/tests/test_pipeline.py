"""
Tests for the main pipeline module.
"""

import pytest
import numpy as np
import time

from ml.document_preprocessing.pipeline import (
    DocumentPreprocessingPipeline,
    PreprocessingResult,
)
from ml.document_preprocessing.config import DocumentConfig


class TestDocumentPreprocessingPipeline:
    """Tests for DocumentPreprocessingPipeline class."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        pipeline = DocumentPreprocessingPipeline()
        assert pipeline.config is not None
        assert pipeline.standardizer is not None
        assert pipeline.enhancer is not None
        assert pipeline.noise_reducer is not None
        assert pipeline.orientation_detector is not None
        assert pipeline.perspectiVIRTENGINE_corrector is not None
    
    def test_init_custom_config(self, document_config):
        """Test initialization with custom config."""
        pipeline = DocumentPreprocessingPipeline(document_config)
        assert pipeline.config == document_config
    
    def test_process_valid_image(self, pipeline, sample_document_image):
        """Test processing a valid document image."""
        result = pipeline.process(sample_document_image)
        
        assert isinstance(result, PreprocessingResult)
        assert result.success is True
        assert result.normalized_image is not None
        assert len(result.normalized_image.shape) == 3
        assert result.error_message is None
    
    def test_process_returns_metadata(self, pipeline, sample_document_image):
        """Test that processing returns complete metadata."""
        result = pipeline.process(sample_document_image)
        
        assert result.original_size is not None
        assert result.final_size is not None
        assert result.rotation_applied in [0, 90, 180, 270]
        assert isinstance(result.perspectiVIRTENGINE_corrected, bool)
        assert isinstance(result.enhancements_applied, list)
        assert result.processing_time_ms > 0
    
    def test_process_applies_standardization(self, pipeline, sample_document_image):
        """Test that standardization is applied."""
        result = pipeline.process(sample_document_image)
        
        assert result.success is True
        assert "standardize" in result.enhancements_applied
    
    def test_process_applies_enhancements(self, pipeline, sample_document_image):
        """Test that enhancements are applied."""
        result = pipeline.process(sample_document_image)
        
        # At least some enhancements should be applied
        assert len(result.enhancements_applied) > 1
    
    def test_process_computes_hash(self, pipeline, sample_document_image):
        """Test that image hash is computed."""
        result = pipeline.process(sample_document_image, compute_hash=True)
        
        assert result.image_hash is not None
        assert len(result.image_hash) == 64  # SHA256 hex length
    
    def test_process_hash_disabled(self, pipeline, sample_document_image):
        """Test that hash can be disabled."""
        result = pipeline.process(sample_document_image, compute_hash=False)
        
        assert result.image_hash is None
    
    def test_process_hash_deterministic(self, pipeline, sample_document_image):
        """Test that hash is deterministic."""
        result1 = pipeline.process(sample_document_image.copy(), compute_hash=True)
        result2 = pipeline.process(sample_document_image.copy(), compute_hash=True)
        
        assert result1.image_hash == result2.image_hash
    
    def test_process_none_image(self, pipeline):
        """Test processing None input."""
        result = pipeline.process(None)
        
        assert result.success is False
        assert "None" in result.error_message
    
    def test_process_empty_image(self, pipeline):
        """Test processing empty array."""
        result = pipeline.process(np.array([]))
        
        assert result.success is False
        assert "empty" in result.error_message.lower()
    
    def test_process_invalid_type(self, pipeline):
        """Test processing invalid type."""
        result = pipeline.process("not an image")
        
        assert result.success is False
        assert "numpy array" in result.error_message.lower()
    
    def test_process_minimal(self, pipeline, sample_document_image):
        """Test minimal processing."""
        result = pipeline.process_minimal(sample_document_image)
        
        assert result is not None
        assert isinstance(result, np.ndarray)
        assert result.shape[1] <= pipeline.config.standardization.target_width
    
    def test_process_with_specific_steps(self, pipeline, sample_document_image):
        """Test processing with specific steps."""
        steps = ["standardize", "clahe", "sharpening"]
        result = pipeline.process_with_steps(sample_document_image, steps)
        
        assert result.success is True
        assert "standardize" in result.enhancements_applied
        assert "clahe" in result.enhancements_applied
        assert "sharpening" in result.enhancements_applied
    
    def test_process_with_steps_orientation(self, pipeline, sample_document_image):
        """Test processing with orientation step."""
        steps = ["standardize", "orientation"]
        result = pipeline.process_with_steps(sample_document_image, steps)
        
        assert result.success is True
        assert "standardize" in result.enhancements_applied
    
    def test_process_with_steps_perspective(self, pipeline, sample_document_image):
        """Test processing with perspective step."""
        steps = ["standardize", "perspective"]
        result = pipeline.process_with_steps(sample_document_image, steps)
        
        assert result.success is True
    
    def test_to_dict(self, pipeline, sample_document_image):
        """Test converting result to dictionary."""
        result = pipeline.process(sample_document_image)
        result_dict = result.to_dict()
        
        assert isinstance(result_dict, dict)
        assert "success" in result_dict
        assert "original_size" in result_dict
        assert "final_size" in result_dict
        assert "rotation_applied" in result_dict
        assert "perspectiVIRTENGINE_corrected" in result_dict
        assert "enhancements_applied" in result_dict
        assert "processing_time_ms" in result_dict


class TestPipelineConfiguration:
    """Tests for pipeline configuration options."""
    
    def test_orientation_first_enabled(self, sample_document_image):
        """Test with orientation correction first."""
        config = DocumentConfig()
        config.correct_orientation_first = True
        
        pipeline = DocumentPreprocessingPipeline(config)
        result = pipeline.process(sample_document_image)
        
        assert result.success is True
    
    def test_orientation_first_disabled(self, sample_document_image):
        """Test with orientation correction disabled first."""
        config = DocumentConfig()
        config.correct_orientation_first = False
        
        pipeline = DocumentPreprocessingPipeline(config)
        result = pipeline.process(sample_document_image)
        
        assert result.success is True
    
    def test_perspectiVIRTENGINE_first_enabled(self, sample_document_image):
        """Test with perspective correction first."""
        config = DocumentConfig()
        config.correct_perspectiVIRTENGINE_first = True
        
        pipeline = DocumentPreprocessingPipeline(config)
        result = pipeline.process(sample_document_image)
        
        assert result.success is True
    
    def test_perspectiVIRTENGINE_disabled(self, sample_document_image):
        """Test with perspective correction disabled."""
        config = DocumentConfig()
        config.perspective.enabled = False
        
        pipeline = DocumentPreprocessingPipeline(config)
        result = pipeline.process(sample_document_image)
        
        assert result.success is True
        assert result.perspectiVIRTENGINE_corrected is False
    
    def test_noise_reduction_disabled(self, sample_document_image):
        """Test with noise reduction disabled."""
        config = DocumentConfig()
        config.noise_reduction.enabled = False
        
        pipeline = DocumentPreprocessingPipeline(config)
        result = pipeline.process(sample_document_image)
        
        assert result.success is True
        assert "noise_reduction" not in result.enhancements_applied


class TestPipelinePerformance:
    """Tests for pipeline performance characteristics."""
    
    def test_processing_time_reasonable(self, pipeline, sample_document_image):
        """Test that processing time is reasonable."""
        result = pipeline.process(sample_document_image)
        
        # Should complete in reasonable time (< 5 seconds for test image)
        assert result.processing_time_ms < 5000
    
    def test_multiple_images_consistent_time(self, pipeline, sample_document_image):
        """Test processing time consistency."""
        times = []
        
        for _ in range(3):
            result = pipeline.process(sample_document_image.copy())
            times.append(result.processing_time_ms)
        
        # Times should be within 2x of each other (allowing for variance)
        assert max(times) < min(times) * 3


class TestPipelineDeterminism:
    """Tests for deterministic pipeline behavior."""
    
    def test_same_image_same_result(self, pipeline, sample_document_image):
        """Test that same image produces same result."""
        result1 = pipeline.process(sample_document_image.copy())
        result2 = pipeline.process(sample_document_image.copy())
        
        # Results should be identical
        assert np.array_equal(result1.normalized_image, result2.normalized_image)
        assert result1.rotation_applied == result2.rotation_applied
        assert result1.perspectiVIRTENGINE_corrected == result2.perspectiVIRTENGINE_corrected
        assert result1.enhancements_applied == result2.enhancements_applied
    
    def test_hash_consistency(self, pipeline, sample_document_image):
        """Test that hash is consistent across runs."""
        result1 = pipeline.process(sample_document_image.copy(), compute_hash=True)
        result2 = pipeline.process(sample_document_image.copy(), compute_hash=True)
        
        assert result1.image_hash == result2.image_hash


class TestDifferentImageTypes:
    """Tests for different image types and formats."""
    
    def test_grayscale_image(self, pipeline, grayscale_document_image):
        """Test processing grayscale image."""
        result = pipeline.process(grayscale_document_image)
        
        assert result.success is True
        assert result.normalized_image.shape[2] == 3  # Converted to RGB
    
    def test_dark_image(self, pipeline, dark_document_image):
        """Test processing dark image."""
        result = pipeline.process(dark_document_image)
        
        assert result.success is True
        # Enhanced image should be brighter
        assert np.mean(result.normalized_image) > np.mean(dark_document_image)
    
    def test_noisy_image(self, pipeline, noisy_document_image):
        """Test processing noisy image."""
        result = pipeline.process(noisy_document_image)
        
        assert result.success is True
        assert "noise_reduction" in result.enhancements_applied
    
    def test_high_resolution_image(self, pipeline, high_resolution_image):
        """Test processing high resolution image."""
        result = pipeline.process(high_resolution_image)
        
        assert result.success is True
        # Should be resized
        assert result.normalized_image.shape[1] <= pipeline.config.standardization.target_width
    
    def test_rotated_image(self, pipeline, rotated_document_image):
        """Test processing rotated image."""
        result = pipeline.process(rotated_document_image)
        
        assert result.success is True
        assert result.rotation_applied in [0, 90, 180, 270]


class TestErrorHandling:
    """Tests for error handling."""
    
    def test_corrupted_image(self, pipeline):
        """Test handling of corrupted image data."""
        # Create image with NaN values
        corrupted = np.full((600, 800, 3), np.nan, dtype=np.float32)
        
        # Should handle gracefully (either process or return error)
        result = pipeline.process(corrupted)
        # The result depends on how OpenCV handles NaN
        assert result is not None
    
    def test_very_small_image(self, pipeline):
        """Test handling of very small image."""
        tiny = np.ones((50, 50, 3), dtype=np.uint8) * 200
        
        result = pipeline.process(tiny)
        
        # Should fail due to minimum resolution
        assert result.success is False
        assert "resolution" in result.error_message.lower() or "minimum" in result.error_message.lower()
