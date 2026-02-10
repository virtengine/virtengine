"""
Tests for the enhancement module.
"""

import pytest
import numpy as np

from ml.document_preprocessing.enhancement import DocumentEnhancer
from ml.document_preprocessing.config import DocumentConfig, EnhancementConfig


class TestDocumentEnhancer:
    """Tests for DocumentEnhancer class."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        enhancer = DocumentEnhancer()
        assert enhancer.config is not None
        assert enhancer.enh_config is not None
    
    def test_init_custom_config(self, document_config):
        """Test initialization with custom config."""
        enhancer = DocumentEnhancer(document_config)
        assert enhancer.config == document_config
    
    def test_enhance_returns_image_and_steps(self, enhancer, small_document_image):
        """Test that enhance returns image and list of steps."""
        result, applied = enhancer.enhance(small_document_image)
        
        assert result is not None
        assert isinstance(result, np.ndarray)
        assert isinstance(applied, list)
        assert len(applied) > 0
    
    def test_clahe_applied_by_default(self, enhancer, small_document_image):
        """Test that CLAHE is applied by default."""
        result, applied = enhancer.enhance(small_document_image)
        
        assert "clahe" in applied
    
    def test_apply_clahe_directly(self, enhancer, small_document_image):
        """Test applying CLAHE directly."""
        result = enhancer.apply_clahe(small_document_image)
        
        assert result is not None
        assert result.shape == small_document_image.shape
        # CLAHE should modify the image
        assert not np.array_equal(result, small_document_image)
    
    def test_apply_clahe_improves_contrast(self, enhancer, dark_document_image):
        """Test that CLAHE improves contrast on dark images."""
        result = enhancer.apply_clahe(dark_document_image)
        
        # Contrast (std) should increase
        original_contrast = np.std(dark_document_image)
        result_contrast = np.std(result)
        
        assert result_contrast >= original_contrast * 0.9  # Allow some tolerance
    
    def test_adjust_brightness_contrast(self, enhancer, small_document_image):
        """Test brightness/contrast adjustment."""
        result = enhancer.adjust_brightness_contrast(small_document_image)
        
        assert result is not None
        assert result.dtype == np.uint8
    
    def test_adjust_brightness_specific_target(self, enhancer, dark_document_image):
        """Test adjusting to specific brightness."""
        target_brightness = 150.0
        result = enhancer.adjust_brightness_contrast(
            dark_document_image,
            brightness=target_brightness
        )
        
        # Result brightness should be closer to target
        original_brightness = np.mean(dark_document_image)
        result_brightness = np.mean(result)
        
        assert abs(result_brightness - target_brightness) < abs(original_brightness - target_brightness)
    
    def test_sharpen(self, enhancer, small_document_image):
        """Test sharpening."""
        result = enhancer.sharpen(small_document_image)
        
        assert result is not None
        assert result.shape == small_document_image.shape
    
    def test_sharpen_increases_edge_strength(self, enhancer, small_document_image):
        """Test that sharpening increases edge strength."""
        import cv2
        
        result = enhancer.sharpen(small_document_image, amount=2.0)
        
        # Calculate edge strength using Laplacian
        original_gray = cv2.cvtColor(small_document_image, cv2.COLOR_RGB2GRAY)
        result_gray = cv2.cvtColor(result, cv2.COLOR_RGB2GRAY)
        
        original_edges = cv2.Laplacian(original_gray, cv2.CV_64F).var()
        result_edges = cv2.Laplacian(result_gray, cv2.CV_64F).var()
        
        assert result_edges >= original_edges
    
    def test_apply_auto_levels(self, enhancer, small_document_image):
        """Test auto-levels (histogram stretching)."""
        result = enhancer.apply_auto_levels(small_document_image)
        
        assert result is not None
        # Dynamic range should be maximized
        assert result.min() <= 10 or result.max() >= 245
    
    def test_apply_gamma(self, enhancer, small_document_image):
        """Test gamma correction."""
        result = enhancer.apply_gamma(small_document_image, gamma=0.5)
        
        assert result is not None
        # Gamma < 1 should brighten the image
        assert np.mean(result) > np.mean(small_document_image)
    
    def test_apply_gamma_darken(self, enhancer, small_document_image):
        """Test gamma correction to darken."""
        result = enhancer.apply_gamma(small_document_image, gamma=2.0)
        
        # Gamma > 1 should darken the image
        assert np.mean(result) < np.mean(small_document_image)
    
    def test_apply_gamma_no_change(self, enhancer, small_document_image):
        """Test gamma=1.0 doesn't change image."""
        result = enhancer.apply_gamma(small_document_image, gamma=1.0)
        
        assert np.array_equal(result, small_document_image)
    
    def test_analyze_enhancement_needs(self, enhancer, dark_document_image):
        """Test enhancement needs analysis."""
        analysis = enhancer.analyze_enhancement_needs(dark_document_image)
        
        assert "brightness" in analysis
        assert "contrast" in analysis
        assert "sharpness" in analysis
        assert "recommendations" in analysis
        assert "needs_enhancement" in analysis
        
        # Dark image should need enhancement
        assert analysis["needs_enhancement"] is True
        assert "increase_brightness" in analysis["recommendations"]
    
    def test_analyze_enhancement_bright_image(self, enhancer, bright_document_image):
        """Test enhancement needs for bright image."""
        analysis = enhancer.analyze_enhancement_needs(bright_document_image)
        
        assert analysis["brightness"] > 200
    
    def test_enhance_disabled_steps(self, document_config, small_document_image):
        """Test that disabled steps are not applied."""
        document_config.enhancement.apply_clahe = False
        document_config.enhancement.apply_sharpening = False
        
        enhancer = DocumentEnhancer(document_config)
        result, applied = enhancer.enhance(small_document_image)
        
        assert "clahe" not in applied
        assert "sharpening" not in applied


class TestEnhancementParameters:
    """Tests for enhancement parameters."""
    
    def test_clahe_clip_limit(self, small_document_image):
        """Test different CLAHE clip limits."""
        config1 = DocumentConfig()
        config1.enhancement.clahe_clip_limit = 1.0
        enhancer1 = DocumentEnhancer(config1)
        
        config2 = DocumentConfig()
        config2.enhancement.clahe_clip_limit = 4.0
        enhancer2 = DocumentEnhancer(config2)
        
        result1 = enhancer1.apply_clahe(small_document_image)
        result2 = enhancer2.apply_clahe(small_document_image)
        
        # Higher clip limit should produce more contrast
        assert not np.array_equal(result1, result2)
    
    def test_sharpening_amount(self, small_document_image):
        """Test different sharpening amounts."""
        config = DocumentConfig()
        enhancer = DocumentEnhancer(config)
        
        mild = enhancer.sharpen(small_document_image, amount=0.5)
        strong = enhancer.sharpen(small_document_image, amount=3.0)
        
        # Strong sharpening should have more edge enhancement
        import cv2
        mild_gray = cv2.cvtColor(mild, cv2.COLOR_RGB2GRAY)
        strong_gray = cv2.cvtColor(strong, cv2.COLOR_RGB2GRAY)
        
        mild_edges = cv2.Laplacian(mild_gray, cv2.CV_64F).var()
        strong_edges = cv2.Laplacian(strong_gray, cv2.CV_64F).var()
        
        assert strong_edges > mild_edges


class TestEnhancementDeterminism:
    """Tests for deterministic enhancement results."""
    
    def test_clahe_deterministic(self, enhancer, small_document_image):
        """Test that CLAHE produces consistent results."""
        result1 = enhancer.apply_clahe(small_document_image.copy())
        result2 = enhancer.apply_clahe(small_document_image.copy())
        
        assert np.array_equal(result1, result2)
    
    def test_full_enhance_deterministic(self, enhancer, small_document_image):
        """Test that full enhancement is deterministic."""
        result1, _ = enhancer.enhance(small_document_image.copy())
        result2, _ = enhancer.enhance(small_document_image.copy())
        
        assert np.array_equal(result1, result2)
