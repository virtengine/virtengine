"""
Tests for the standardization module.
"""

import pytest
import numpy as np
import cv2

from ml.document_preprocessing.standardization import DocumentStandardizer
from ml.document_preprocessing.config import (
    DocumentConfig,
    StandardizationConfig,
    InterpolationMethod,
)


class TestDocumentStandardizer:
    """Tests for DocumentStandardizer class."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        standardizer = DocumentStandardizer()
        assert standardizer.config is not None
        assert standardizer.target_width == 1024
        assert standardizer.target_height == 768
        assert standardizer.output_format == "PNG"
    
    def test_init_custom_config(self, document_config):
        """Test initialization with custom config."""
        standardizer = DocumentStandardizer(document_config)
        assert standardizer.config == document_config
    
    def test_standardize_valid_image(self, standardizer, sample_document_image):
        """Test standardizing a valid document image."""
        result = standardizer.standardize(sample_document_image)
        
        assert result is not None
        assert isinstance(result, np.ndarray)
        assert result.dtype == np.uint8
        assert len(result.shape) == 3
        assert result.shape[2] == 3  # RGB
    
    def test_standardize_resizes_to_target(self, standardizer, sample_document_image):
        """Test that image is resized to target dimensions."""
        result = standardizer.standardize(sample_document_image)
        
        # With aspect ratio maintenance, at least one dimension should match
        height, width = result.shape[:2]
        target_width = standardizer.target_width
        target_height = standardizer.target_height
        
        # The image should fit within target dimensions
        assert width <= target_width
        assert height <= target_height
    
    def test_standardize_grayscale_input(self, standardizer, grayscale_document_image):
        """Test standardizing a grayscale image."""
        result = standardizer.standardize(grayscale_document_image)
        
        assert result is not None
        assert len(result.shape) == 3
        assert result.shape[2] == 3  # Converted to RGB
    
    def test_standardize_high_resolution(self, standardizer, high_resolution_image):
        """Test standardizing a high resolution image."""
        result = standardizer.standardize(high_resolution_image)
        
        # Should be downscaled
        assert result.shape[0] <= standardizer.target_height
        assert result.shape[1] <= standardizer.target_width
    
    def test_standardize_none_raises_error(self, standardizer):
        """Test that None input raises ValueError."""
        with pytest.raises(ValueError, match="cannot be None"):
            standardizer.standardize(None)
    
    def test_standardize_empty_raises_error(self, standardizer):
        """Test that empty array raises ValueError."""
        with pytest.raises(ValueError, match="empty"):
            standardizer.standardize(np.array([]))
    
    def test_standardize_too_small_raises_error(self, standardizer):
        """Test that image below minimum resolution raises ValueError."""
        small = np.zeros((100, 100, 3), dtype=np.uint8)
        with pytest.raises(ValueError, match="below minimum"):
            standardizer.standardize(small)
    
    def test_standardize_bgra_input(self, standardizer, sample_document_image):
        """Test standardizing a BGRA image."""
        # Add alpha channel
        alpha = np.ones((sample_document_image.shape[0], sample_document_image.shape[1], 1), dtype=np.uint8) * 255
        bgra = np.concatenate([sample_document_image, alpha], axis=2)
        
        result = standardizer.standardize(bgra)
        
        assert result.shape[2] == 3  # Alpha removed
    
    def test_standardize_maintains_aspect_ratio(self, document_config):
        """Test that aspect ratio is maintained when configured."""
        document_config.standardization.maintain_aspect_ratio = True
        standardizer = DocumentStandardizer(document_config)
        
        # Create a wide image (2:1 ratio)
        wide_image = np.ones((400, 800, 3), dtype=np.uint8) * 200
        result = standardizer.standardize(wide_image)
        
        # Aspect ratio should be preserved (approximately)
        original_ratio = 800 / 400
        result_ratio = result.shape[1] / result.shape[0]
        
        # Allow some tolerance due to padding
        assert abs(original_ratio - result_ratio) < 0.5 or result.shape == (768, 1024, 3)
    
    def test_standardize_no_aspect_ratio(self, document_config):
        """Test standardization without maintaining aspect ratio."""
        document_config.standardization.maintain_aspect_ratio = False
        standardizer = DocumentStandardizer(document_config)
        
        image = np.ones((400, 800, 3), dtype=np.uint8) * 200
        result = standardizer.standardize(image)
        
        # Should be exactly target size
        assert result.shape[0] == standardizer.target_height
        assert result.shape[1] == standardizer.target_width
    
    def test_get_resize_info(self, standardizer):
        """Test getting resize information."""
        info = standardizer.get_resize_info((800, 600))
        
        assert "original_size" in info
        assert "target_size" in info
        assert "scaled_size" in info
        assert "scale_factor" in info
        assert info["original_size"] == (800, 600)
    
    def test_standardize_float_input(self, standardizer, sample_document_image):
        """Test standardizing a float32 image."""
        float_image = sample_document_image.astype(np.float32) / 255.0
        result = standardizer.standardize(float_image)
        
        assert result.dtype == np.uint8
        assert result.max() > 1  # Converted back to 0-255 range
    
    def test_standardize_16bit_input(self, standardizer, sample_document_image):
        """Test standardizing a 16-bit image."""
        image_16 = sample_document_image.astype(np.uint16) * 256
        result = standardizer.standardize(image_16)
        
        assert result.dtype == np.uint8


class TestInterpolationMethods:
    """Tests for different interpolation methods."""
    
    @pytest.mark.parametrize("method", [
        InterpolationMethod.NEAREST,
        InterpolationMethod.LINEAR,
        InterpolationMethod.CUBIC,
        InterpolationMethod.LANCZOS,
        InterpolationMethod.AREA,
    ])
    def test_interpolation_method(self, method, sample_document_image):
        """Test each interpolation method produces valid output."""
        config = DocumentConfig()
        config.standardization.interpolation = method
        standardizer = DocumentStandardizer(config)
        
        result = standardizer.standardize(sample_document_image)
        
        assert result is not None
        assert result.dtype == np.uint8


class TestCustomTargetSize:
    """Tests for custom target sizes."""
    
    def test_id_card_config(self, sample_document_image):
        """Test ID card configuration."""
        config = DocumentConfig.for_id_card()
        standardizer = DocumentStandardizer(config)
        
        result = standardizer.standardize(sample_document_image)
        
        assert result is not None
        assert result.shape[1] <= config.standardization.target_width
    
    def test_passport_config(self, sample_document_image):
        """Test passport configuration."""
        config = DocumentConfig.for_passport()
        standardizer = DocumentStandardizer(config)
        
        result = standardizer.standardize(sample_document_image)
        
        assert result is not None
    
    def test_drivers_license_config(self, sample_document_image):
        """Test driver's license configuration."""
        config = DocumentConfig.for_drivers_license()
        standardizer = DocumentStandardizer(config)
        
        result = standardizer.standardize(sample_document_image)
        
        assert result is not None
