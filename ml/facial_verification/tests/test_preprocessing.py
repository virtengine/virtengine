"""
Tests for the preprocessing module.
"""

import pytest
import numpy as np

from ml.facial_verification.preprocessing import (
    FacePreprocessor,
    PreprocessingResult,
)
from ml.facial_verification.config import VerificationConfig, PreprocessingConfig
from ml.facial_verification.reason_codes import ReasonCodes


class TestFacePreprocessor:
    """Tests for FacePreprocessor class."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        preprocessor = FacePreprocessor()
        assert preprocessor.config is not None
        assert preprocessor.preprocessing_config is not None
    
    def test_init_custom_config(self, verification_config):
        """Test initialization with custom config."""
        preprocessor = FacePreprocessor(verification_config)
        assert preprocessor.config == verification_config
    
    def test_preprocess_valid_image(self, preprocessor, sample_face_image):
        """Test preprocessing a valid image."""
        result = preprocessor.preprocess(sample_face_image)
        
        assert result.success is True
        assert result.error_code is None
        assert result.image is not None
        assert len(result.applied_steps) > 0
        assert "resize" in result.applied_steps
    
    def test_preprocess_applies_clahe(self, sample_face_image):
        """Test that CLAHE is applied when configured."""
        config = VerificationConfig()
        config.preprocessing.apply_clahe = True
        preprocessor = FacePreprocessor(config)
        
        result = preprocessor.preprocess(sample_face_image)
        
        assert result.success is True
        assert "clahe" in result.applied_steps
    
    def test_preprocess_applies_noise_reduction(self, sample_face_image):
        """Test that noise reduction is applied when configured."""
        config = VerificationConfig()
        config.preprocessing.noise_reduction = True
        preprocessor = FacePreprocessor(config)
        
        result = preprocessor.preprocess(sample_face_image)
        
        assert result.success is True
        assert "noise_reduction" in result.applied_steps
    
    def test_preprocess_grayscale_conversion(self, sample_face_image):
        """Test grayscale conversion when configured."""
        config = VerificationConfig()
        config.preprocessing.use_grayscale = True
        preprocessor = FacePreprocessor(config)
        
        result = preprocessor.preprocess(sample_face_image)
        
        assert result.success is True
        assert "grayscale" in result.applied_steps
    
    def test_preprocess_normalization(self, sample_face_image):
        """Test pixel normalization."""
        config = VerificationConfig()
        config.preprocessing.normalize_color = True
        preprocessor = FacePreprocessor(config)
        
        result = preprocessor.preprocess(sample_face_image)
        
        assert result.success is True
        assert "normalize" in result.applied_steps
        assert result.image.dtype == np.float32
    
    def test_preprocess_resize_to_target(self, preprocessor):
        """Test that images are resized to target resolution."""
        # Create a large image
        large_image = np.random.randint(0, 255, (512, 512, 3), dtype=np.uint8)
        
        result = preprocessor.preprocess(large_image)
        
        assert result.success is True
        target = preprocessor.preprocessing_config.target_resolution
        assert result.final_size == target or result.image.shape[:2] == target[::-1]
    
    def test_preprocess_none_image(self, preprocessor):
        """Test preprocessing with None input."""
        result = preprocessor.preprocess(None)
        
        assert result.success is False
        assert result.error_code == ReasonCodes.INVALID_IMAGE_FORMAT
    
    def test_preprocess_empty_image(self, preprocessor):
        """Test preprocessing with empty array."""
        empty = np.array([])
        result = preprocessor.preprocess(empty)
        
        assert result.success is False
        assert result.error_code in [
            ReasonCodes.INVALID_IMAGE_FORMAT,
            ReasonCodes.CORRUPT_IMAGE_DATA
        ]
    
    def test_preprocess_low_resolution_image(self, preprocessor, low_quality_image):
        """Test preprocessing with low resolution image."""
        result = preprocessor.preprocess(low_quality_image)
        
        # Should still succeed but may flag low resolution
        assert result.error_code == ReasonCodes.LOW_RESOLUTION
    
    def test_preprocess_computes_hash(self, preprocessor, sample_face_image):
        """Test that image hash is computed."""
        result = preprocessor.preprocess(sample_face_image, compute_hash=True)
        
        assert result.success is True
        assert result.image_hash is not None
        assert len(result.image_hash) == 64  # SHA256 hex length
    
    def test_preprocess_hash_deterministic(self, preprocessor, sample_face_image):
        """Test that hash is deterministic."""
        result1 = preprocessor.preprocess(sample_face_image.copy(), compute_hash=True)
        result2 = preprocessor.preprocess(sample_face_image.copy(), compute_hash=True)
        
        assert result1.image_hash == result2.image_hash
    
    def test_check_image_quality_good_image(self, preprocessor, sample_face_image):
        """Test quality check on good image."""
        quality, issues = preprocessor.check_image_quality(sample_face_image)
        
        assert 0.0 <= quality <= 1.0
        # Good image should have minimal issues
        assert ReasonCodes.LOW_QUALITY_IMAGE not in issues
    
    def test_check_image_quality_dark_image(self, preprocessor, dark_image):
        """Test quality check on dark image."""
        quality, issues = preprocessor.check_image_quality(dark_image)
        
        assert ReasonCodes.IMAGE_TOO_DARK in issues
        assert quality < 1.0
    
    def test_check_image_quality_bright_image(self, preprocessor, bright_image):
        """Test quality check on bright image."""
        quality, issues = preprocessor.check_image_quality(bright_image)
        
        assert ReasonCodes.IMAGE_TOO_BRIGHT in issues
        assert quality < 1.0
    
    def test_preprocessing_result_to_dict(self, preprocessor, sample_face_image):
        """Test PreprocessingResult serialization."""
        result = preprocessor.preprocess(sample_face_image)
        
        result_dict = result.to_dict()
        
        assert "original_size" in result_dict
        assert "final_size" in result_dict
        assert "applied_steps" in result_dict
        assert "success" in result_dict


class TestPreprocessingConfig:
    """Tests for PreprocessingConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = PreprocessingConfig()
        
        assert config.target_resolution == (224, 224)
        assert config.use_grayscale is False
        assert config.apply_clahe is True
        assert config.noise_reduction is True
    
    def test_custom_values(self):
        """Test custom configuration values."""
        config = PreprocessingConfig(
            target_resolution=(160, 160),
            use_grayscale=True,
            apply_clahe=False,
        )
        
        assert config.target_resolution == (160, 160)
        assert config.use_grayscale is True
        assert config.apply_clahe is False
    
    def test_bilateral_filter_config(self):
        """Test bilateral filter configuration."""
        config = PreprocessingConfig(
            noise_reduction_method="bilateral",
            bilateral_d=5,
            bilateral_sigma_color=50.0,
        )
        
        assert config.noise_reduction_method == "bilateral"
        assert config.bilateral_d == 5
        assert config.bilateral_sigma_color == 50.0
    
    def test_gaussian_filter_config(self):
        """Test Gaussian filter configuration."""
        config = PreprocessingConfig(
            noise_reduction_method="gaussian",
            gaussian_kernel_size=(3, 3),
        )
        
        assert config.noise_reduction_method == "gaussian"
        assert config.gaussian_kernel_size == (3, 3)
