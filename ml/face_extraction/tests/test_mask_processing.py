"""
Tests for mask post-processing module.

This module tests mask processing operations including:
- Thresholding (binary and adaptive)
- Morphological operations
- Connected component analysis
- Contour smoothing
- Full processing pipeline
"""

import pytest
import numpy as np
import cv2


class TestMaskProcessor:
    """Tests for MaskProcessor class."""
    
    def test_initialization(self, mask_config):
        """Test processor initialization."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        assert processor is not None
        assert processor.config == mask_config
    
    def test_default_initialization(self):
        """Test processor with default config."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor()
        
        assert processor is not None
        assert processor.config is not None


class TestThresholding:
    """Tests for thresholding operations."""
    
    def test_binary_threshold(self, mask_config, sample_face_mask):
        """Test binary thresholding."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        binary = processor.threshold(sample_face_mask)
        
        # Should be binary (0 or 255)
        unique_values = np.unique(binary)
        assert len(unique_values) <= 2
        assert all(v in [0, 255] for v in unique_values)
    
    def test_threshold_with_custom_value(self, mask_config):
        """Test thresholding with custom threshold value."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # Create gradient mask
        gradient = np.linspace(0, 1, 100).reshape(10, 10).astype(np.float32)
        
        # Threshold at 0.7
        binary = processor.threshold(gradient, threshold=0.7)
        
        # Only values >= 0.7 should be 255
        high_region = gradient >= 0.7
        assert np.all(binary[high_region] == 255)
        assert np.all(binary[~high_region] == 0)
    
    def test_threshold_normalized_input(self, mask_config):
        """Test thresholding handles already normalized input."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # Create mask with values in [0, 1]
        mask = np.array([[0.2, 0.6], [0.4, 0.8]], dtype=np.float32)
        binary = processor.threshold(mask, threshold=0.5)
        
        expected = np.array([[0, 255], [0, 255]], dtype=np.uint8)
        np.testing.assert_array_equal(binary, expected)
    
    def test_threshold_uint8_input(self, mask_config):
        """Test thresholding handles uint8 input."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # Create mask with values 0-255
        mask = np.array([[50, 150], [100, 200]], dtype=np.uint8)
        binary = processor.threshold(mask, threshold=0.5)
        
        # Should normalize and threshold
        assert binary.dtype == np.uint8
        assert np.all((binary == 0) | (binary == 255))


class TestMorphologicalOperations:
    """Tests for morphological operations."""
    
    def test_erosion(self, mask_config):
        """Test erosion operation."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # Create a simple binary mask with a filled square
        mask = np.zeros((100, 100), dtype=np.uint8)
        mask[30:70, 30:70] = 255
        
        eroded = processor.erode(mask)
        
        # Eroded mask should have smaller foreground
        original_area = np.sum(mask > 0)
        eroded_area = np.sum(eroded > 0)
        assert eroded_area < original_area
    
    def test_dilation(self, mask_config):
        """Test dilation operation."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # Create a simple binary mask
        mask = np.zeros((100, 100), dtype=np.uint8)
        mask[40:60, 40:60] = 255
        
        dilated = processor.dilate(mask)
        
        # Dilated mask should have larger foreground
        original_area = np.sum(mask > 0)
        dilated_area = np.sum(dilated > 0)
        assert dilated_area > original_area
    
    def test_opening_removes_noise(self, mask_config):
        """Test that opening removes small noise."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # Create mask with main region and small noise spots
        mask = np.zeros((100, 100), dtype=np.uint8)
        mask[30:70, 30:70] = 255  # Main region
        mask[5:8, 5:8] = 255  # Small noise
        mask[90:93, 90:93] = 255  # Another noise
        
        opened = processor.opening(mask)
        
        # Small noise should be removed, main region preserved
        assert np.sum(opened[30:70, 30:70] > 0) > 0  # Main region preserved
        assert np.sum(opened[5:8, 5:8] > 0) == 0  # Noise removed
    
    def test_closing_fills_holes(self, mask_config):
        """Test that closing fills small holes."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # Create mask with a hole
        mask = np.zeros((100, 100), dtype=np.uint8)
        mask[20:80, 20:80] = 255
        mask[45:55, 45:55] = 0  # Small hole
        
        closed = processor.closing(mask)
        
        # Hole should be filled
        hole_region = closed[45:55, 45:55]
        assert np.mean(hole_region) > 0
    
    def test_morphological_cleanup_sequence(self, mask_config, noisy_face_mask):
        """Test full morphological cleanup."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # First threshold the noisy mask
        binary = processor.threshold(noisy_face_mask)
        
        # Apply cleanup
        cleaned = processor.morphological_cleanup(binary)
        
        assert cleaned.shape == binary.shape
        assert cleaned.dtype == np.uint8


class TestConnectedComponents:
    """Tests for connected component analysis."""
    
    def test_largest_connected_component(self, mask_config, fragmented_mask):
        """Test keeping only largest connected component."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # Threshold and get components
        binary = processor.threshold(fragmented_mask)
        
        # Count components before
        num_before, _, _, _ = cv2.connectedComponentsWithStats(binary)
        
        # Keep largest
        result = processor.largest_connected_component(binary)
        
        # Count components after
        num_after, _, _, _ = cv2.connectedComponentsWithStats(result)
        
        # Should have only 1 component (plus background)
        assert num_after <= 2
        assert num_after < num_before
    
    def test_get_connected_components(self, mask_config, fragmented_mask):
        """Test getting component statistics."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        binary = processor.threshold(fragmented_mask)
        
        num_labels, labels, stats, centroids = processor.get_connected_components(binary)
        
        assert num_labels > 1  # At least background + 1 component
        assert labels.shape == binary.shape
        assert stats.shape[0] == num_labels
        assert centroids.shape[0] == num_labels
    
    def test_min_component_area_filter(self, mask_config):
        """Test minimum component area filtering."""
        from ml.face_extraction.mask_processing import MaskProcessor
        from ml.face_extraction.config import MaskProcessingConfig
        
        config = MaskProcessingConfig(min_component_area=1000)
        processor = MaskProcessor(config)
        
        # Create mask with small components
        mask = np.zeros((100, 100), dtype=np.uint8)
        mask[10:15, 10:15] = 255  # 25 pixels - below threshold
        mask[40:70, 40:70] = 255  # 900 pixels - also below 1000
        
        result = processor.largest_connected_component(mask)
        
        # Both components are below threshold, should keep largest
        assert np.sum(result > 0) > 0


class TestContourSmoothing:
    """Tests for contour smoothing."""
    
    def test_smooth_contours(self, mask_config):
        """Test contour smoothing."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # Create jagged mask
        mask = np.zeros((100, 100), dtype=np.uint8)
        mask[20:80, 20:80] = 255
        # Add jagged edges
        for i in range(20, 80, 2):
            if i % 4 == 0:
                mask[i:i+2, 18:20] = 255
                mask[i:i+2, 80:82] = 255
        
        smoothed = processor.smooth_contours(mask)
        
        assert smoothed.shape == mask.shape
        assert np.sum(smoothed > 0) > 0
    
    def test_fill_holes(self, mask_config):
        """Test hole filling."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # Create mask with holes
        mask = np.zeros((100, 100), dtype=np.uint8)
        mask[20:80, 20:80] = 255
        mask[40:60, 40:60] = 0  # Hole
        
        original_foreground = np.sum(mask > 0)
        
        filled = processor.fill_holes(mask)
        
        filled_foreground = np.sum(filled > 0)
        
        # Should have more foreground pixels after filling
        assert filled_foreground >= original_foreground


class TestFullProcessingPipeline:
    """Tests for complete mask processing pipeline."""
    
    def test_process_returns_result(self, mask_config, sample_face_mask):
        """Test that process returns proper result object."""
        from ml.face_extraction.mask_processing import MaskProcessor, MaskProcessingResult
        
        processor = MaskProcessor(mask_config)
        result = processor.process(sample_face_mask)
        
        assert isinstance(result, MaskProcessingResult)
        assert result.success
        assert result.processed_mask is not None
        assert result.original_mask is not None
        assert len(result.operations_applied) > 0
    
    def test_process_noisy_mask(self, mask_config, noisy_face_mask):
        """Test processing of noisy mask."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        result = processor.process(noisy_face_mask)
        
        assert result.success
        # Processed mask should be cleaner (fewer components)
        assert result.num_components <= 5
    
    def test_process_fragmented_mask(self, mask_config, fragmented_mask):
        """Test processing of fragmented mask."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        result = processor.process(fragmented_mask)
        
        assert result.success
        # Should keep only largest component
        assert result.num_components == 1
        assert result.largest_component_area > 0
    
    def test_process_invalid_input(self, mask_config):
        """Test processing of invalid input."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        # None input
        result = processor.process(None)
        assert not result.success
        
        # Empty input
        result = processor.process(np.array([]))
        assert not result.success
    
    def test_operations_tracking(self, mask_config, sample_face_mask):
        """Test that operations are tracked correctly."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        result = processor.process(sample_face_mask)
        
        # Check expected operations
        assert "threshold" in result.operations_applied
        
        if mask_config.apply_morphology:
            assert "morphological_cleanup" in result.operations_applied
        
        if mask_config.use_largest_component:
            assert "largest_component" in result.operations_applied
    
    def test_to_dict_serialization(self, mask_config, sample_face_mask):
        """Test result serialization."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        result = processor.process(sample_face_mask)
        
        result_dict = result.to_dict()
        
        assert isinstance(result_dict, dict)
        assert "success" in result_dict
        assert "num_components" in result_dict
        assert "operations_applied" in result_dict


class TestEdgeCases:
    """Tests for edge cases and error handling."""
    
    def test_all_zeros_mask(self, mask_config):
        """Test processing of empty (all zeros) mask."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        empty_mask = np.zeros((100, 100), dtype=np.float32)
        result = processor.process(empty_mask)
        
        assert result.success
        assert result.num_components == 0
    
    def test_all_ones_mask(self, mask_config):
        """Test processing of full (all ones) mask."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        full_mask = np.ones((100, 100), dtype=np.float32)
        result = processor.process(full_mask)
        
        assert result.success
        assert result.num_components >= 1
    
    def test_single_pixel_mask(self, mask_config):
        """Test processing of single pixel mask."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        single_pixel = np.zeros((100, 100), dtype=np.float32)
        single_pixel[50, 50] = 1.0
        
        result = processor.process(single_pixel)
        
        assert result.success
    
    def test_very_small_mask(self, mask_config):
        """Test processing of very small mask."""
        from ml.face_extraction.mask_processing import MaskProcessor
        
        processor = MaskProcessor(mask_config)
        
        small_mask = np.array([[0.0, 1.0], [1.0, 0.0]], dtype=np.float32)
        result = processor.process(small_mask)
        
        assert result.success or result.error_message is not None
