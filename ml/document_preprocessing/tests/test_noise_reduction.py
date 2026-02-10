"""
Tests for the noise reduction module.
"""

import pytest
import numpy as np
import cv2

from ml.document_preprocessing.noise_reduction import NoiseReducer
from ml.document_preprocessing.config import (
    DocumentConfig,
    NoiseReductionConfig,
    NoiseReductionMethod,
)


class TestNoiseReducer:
    """Tests for NoiseReducer class."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        reducer = NoiseReducer()
        assert reducer.config is not None
        assert reducer.nr_config is not None
        assert reducer.nr_config.enabled is True
    
    def test_init_custom_config(self, document_config):
        """Test initialization with custom config."""
        reducer = NoiseReducer(document_config)
        assert reducer.config == document_config
    
    def test_denoise_default_method(self, noise_reducer, noisy_document_image):
        """Test denoising with default method (bilateral)."""
        result = noise_reducer.denoise(noisy_document_image)
        
        assert result is not None
        assert result.shape == noisy_document_image.shape
        assert result.dtype == noisy_document_image.dtype
    
    def test_denoise_disabled(self, document_config, noisy_document_image):
        """Test that denoising can be disabled."""
        document_config.noise_reduction.enabled = False
        reducer = NoiseReducer(document_config)
        
        result = reducer.denoise(noisy_document_image)
        
        assert np.array_equal(result, noisy_document_image)
    
    def test_gaussian_blur(self, noise_reducer, noisy_document_image):
        """Test Gaussian blur."""
        result = noise_reducer.gaussian_blur(noisy_document_image)
        
        assert result is not None
        assert result.shape == noisy_document_image.shape
        # Blur should reduce variance
        assert np.std(result) <= np.std(noisy_document_image)
    
    def test_gaussian_blur_custom_params(self, noise_reducer, noisy_document_image):
        """Test Gaussian blur with custom parameters."""
        result = noise_reducer.gaussian_blur(noisy_document_image, ksize=7, sigma=2.0)
        
        assert result is not None
    
    def test_gaussian_blur_even_ksize_corrected(self, noise_reducer, noisy_document_image):
        """Test that even kernel size is corrected to odd."""
        # Even kernel size should be corrected to odd
        result = noise_reducer.gaussian_blur(noisy_document_image, ksize=4)
        
        assert result is not None  # Should not raise error
    
    def test_median_blur(self, noise_reducer, noisy_document_image):
        """Test median blur."""
        result = noise_reducer.median_blur(noisy_document_image)
        
        assert result is not None
        assert result.shape == noisy_document_image.shape
    
    def test_median_blur_removes_salt_pepper(self, noise_reducer, np_random):
        """Test that median blur removes salt and pepper noise."""
        # Create image with salt and pepper noise
        image = np.ones((200, 300, 3), dtype=np.uint8) * 128
        
        # Add salt and pepper
        salt = np_random.random(image.shape[:2]) < 0.05
        pepper = np_random.random(image.shape[:2]) < 0.05
        image[salt] = [255, 255, 255]
        image[pepper] = [0, 0, 0]
        
        result = noise_reducer.median_blur(image)
        
        # Median filter should reduce extreme values
        assert result.min() > 0 or result.max() < 255
    
    def test_bilateral_filter(self, noise_reducer, noisy_document_image):
        """Test bilateral filtering."""
        result = noise_reducer.bilateral_filter(noisy_document_image)
        
        assert result is not None
        assert result.shape == noisy_document_image.shape
    
    def test_bilateral_filter_custom_params(self, noise_reducer, noisy_document_image):
        """Test bilateral filter with custom parameters."""
        result = noise_reducer.bilateral_filter(
            noisy_document_image,
            d=5,
            sigma_color=50.0,
            sigma_space=50.0
        )
        
        assert result is not None
    
    def test_bilateral_preserves_edges(self, noise_reducer, small_document_image):
        """Test that bilateral filter preserves edges."""
        # Add some noise
        noisy = small_document_image.copy()
        noise = np.random.randint(-20, 20, noisy.shape, dtype=np.int16)
        noisy = np.clip(noisy.astype(np.int16) + noise, 0, 255).astype(np.uint8)
        
        result = noise_reducer.bilateral_filter(noisy)
        
        # Calculate edge strength
        gray_orig = cv2.cvtColor(small_document_image, cv2.COLOR_RGB2GRAY)
        gray_result = cv2.cvtColor(result, cv2.COLOR_RGB2GRAY)
        
        edges_orig = cv2.Canny(gray_orig, 50, 150)
        edges_result = cv2.Canny(gray_result, 50, 150)
        
        # Edges should be reasonably preserved
        orig_edge_count = np.sum(edges_orig > 0)
        result_edge_count = np.sum(edges_result > 0)
        
        # Allow some edge loss but not too much
        assert result_edge_count >= orig_edge_count * 0.5
    
    def test_denoise_with_method_string(self, noise_reducer, noisy_document_image):
        """Test denoising with method specified as string."""
        result_gauss = noise_reducer.denoise(noisy_document_image, method="gaussian")
        result_median = noise_reducer.denoise(noisy_document_image, method="median")
        result_bilateral = noise_reducer.denoise(noisy_document_image, method="bilateral")
        
        assert result_gauss is not None
        assert result_median is not None
        assert result_bilateral is not None
        
        # Different methods should produce different results
        assert not np.array_equal(result_gauss, result_median)
    
    def test_denoise_unknown_method_fallback(self, noise_reducer, noisy_document_image):
        """Test that unknown method falls back to bilateral."""
        result = noise_reducer.denoise(noisy_document_image, method="unknown")
        
        assert result is not None


class TestNoiseEstimation:
    """Tests for noise level estimation."""
    
    def test_estimate_noise_level(self, noise_reducer, noisy_document_image):
        """Test noise level estimation."""
        noise_level = noise_reducer.estimate_noise_level(noisy_document_image)
        
        assert isinstance(noise_level, float)
        assert noise_level >= 0
    
    def test_estimate_noise_clean_image(self, noise_reducer, small_document_image):
        """Test noise estimation on clean image."""
        noise_level = noise_reducer.estimate_noise_level(small_document_image)
        
        # Clean image should have low noise
        assert noise_level < 50  # Reasonable threshold
    
    def test_estimate_noise_noisy_image(self, noise_reducer, noisy_document_image):
        """Test noise estimation on noisy image."""
        noise_level = noise_reducer.estimate_noise_level(noisy_document_image)
        
        # Noisy image should have higher noise level
        assert noise_level > 5
    
    def test_estimate_noise_grayscale(self, noise_reducer, grayscale_document_image):
        """Test noise estimation on grayscale image."""
        noise_level = noise_reducer.estimate_noise_level(grayscale_document_image)
        
        assert isinstance(noise_level, float)
        assert noise_level >= 0


class TestAdaptiveDenoising:
    """Tests for adaptive denoising."""
    
    def test_adaptive_denoise_low_noise(self, noise_reducer, small_document_image):
        """Test adaptive denoising on low-noise image."""
        result = noise_reducer.adaptive_denoise(small_document_image)
        
        assert result is not None
        # Low noise should result in minimal denoising
    
    def test_adaptive_denoise_high_noise(self, noise_reducer, noisy_document_image):
        """Test adaptive denoising on high-noise image."""
        result = noise_reducer.adaptive_denoise(noisy_document_image)
        
        assert result is not None
        # High noise should result in stronger denoising
        assert np.std(result) < np.std(noisy_document_image)


class TestNonLocalMeans:
    """Tests for Non-Local Means denoising."""
    
    def test_non_local_means(self, noise_reducer, noisy_document_image):
        """Test Non-Local Means denoising."""
        result = noise_reducer.non_local_means(noisy_document_image)
        
        assert result is not None
        assert result.shape == noisy_document_image.shape
    
    def test_non_local_means_custom_params(self, noise_reducer, noisy_document_image):
        """Test NLM with custom parameters."""
        result = noise_reducer.non_local_means(
            noisy_document_image,
            h=5.0,
            template_window_size=5,
            search_window_size=15
        )
        
        assert result is not None
    
    def test_non_local_means_reduces_noise(self, noise_reducer, noisy_document_image):
        """Test that NLM effectively reduces noise."""
        result = noise_reducer.non_local_means(noisy_document_image, h=10.0)
        
        # NLM should reduce variance (noise)
        orig_var = np.var(noisy_document_image)
        result_var = np.var(result)
        
        assert result_var < orig_var


class TestMorphologicalCleaning:
    """Tests for morphological cleaning."""
    
    def test_morphological_cleaning(self, document_config, noisy_document_image):
        """Test morphological cleaning option."""
        document_config.noise_reduction.apply_morphological = True
        reducer = NoiseReducer(document_config)
        
        result = reducer.denoise(noisy_document_image)
        
        assert result is not None
    
    def test_morphological_disabled(self, document_config, noisy_document_image):
        """Test with morphological cleaning disabled."""
        document_config.noise_reduction.apply_morphological = False
        reducer = NoiseReducer(document_config)
        
        result = reducer.denoise(noisy_document_image)
        
        assert result is not None


class TestDenosingDeterminism:
    """Tests for deterministic denoising results."""
    
    def test_gaussian_deterministic(self, noise_reducer, noisy_document_image):
        """Test that Gaussian blur is deterministic."""
        result1 = noise_reducer.gaussian_blur(noisy_document_image.copy())
        result2 = noise_reducer.gaussian_blur(noisy_document_image.copy())
        
        assert np.array_equal(result1, result2)
    
    def test_bilateral_deterministic(self, noise_reducer, noisy_document_image):
        """Test that bilateral filter is deterministic."""
        result1 = noise_reducer.bilateral_filter(noisy_document_image.copy())
        result2 = noise_reducer.bilateral_filter(noisy_document_image.copy())
        
        assert np.array_equal(result1, result2)
    
    def test_median_deterministic(self, noise_reducer, noisy_document_image):
        """Test that median blur is deterministic."""
        result1 = noise_reducer.median_blur(noisy_document_image.copy())
        result2 = noise_reducer.median_blur(noisy_document_image.copy())
        
        assert np.array_equal(result1, result2)
