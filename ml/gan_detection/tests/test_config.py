"""
Tests for GAN detection configuration.

VE-923: Test configuration classes.
"""

import pytest

from ml.gan_detection.config import (
    GANDetectionConfig,
    DiscriminatorConfig,
    DeepfakeConfig,
    ArtifactAnalysisConfig,
    VEIDIntegrationConfig,
    DetectionMode,
    SyntheticImageType,
    ArtifactType,
)


class TestDetectionMode:
    """Tests for DetectionMode enum."""
    
    def test_detection_modes(self):
        """Test detection mode values."""
        assert DetectionMode.FULL.value == "full"
        assert DetectionMode.FAST.value == "fast"
        assert DetectionMode.ACCURATE.value == "accurate"


class TestSyntheticImageType:
    """Tests for SyntheticImageType enum."""
    
    def test_synthetic_types(self):
        """Test synthetic image type values."""
        assert SyntheticImageType.GAN_GENERATED.value == "gan_generated"
        assert SyntheticImageType.DEEPFAKE_FACESWAP.value == "deepfake_faceswap"
        assert SyntheticImageType.DEEPFAKE_EXPRESSION.value == "deepfake_expression"
        assert SyntheticImageType.MORPHED.value == "morphed"


class TestArtifactType:
    """Tests for ArtifactType enum."""
    
    def test_artifact_types(self):
        """Test artifact type values."""
        assert ArtifactType.FREQUENCY_ANOMALY.value == "frequency_anomaly"
        assert ArtifactType.CHECKERBOARD.value == "checkerboard"
        assert ArtifactType.BLENDING_BOUNDARY.value == "blending_boundary"


class TestDiscriminatorConfig:
    """Tests for DiscriminatorConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = DiscriminatorConfig()
        
        assert config.input_size == (224, 224)
        assert config.input_channels == 3
        assert config.base_filters == 64
        assert config.num_blocks == 4
        assert config.use_batch_norm is True
        assert config.use_dropout is True
        assert config.dropout_rate == 0.3
        assert config.num_classes == 2
    
    def test_custom_values(self):
        """Test custom configuration values."""
        config = DiscriminatorConfig(
            input_size=(128, 128),
            base_filters=32,
            num_blocks=3,
            dropout_rate=0.5,
        )
        
        assert config.input_size == (128, 128)
        assert config.base_filters == 32
        assert config.num_blocks == 3
        assert config.dropout_rate == 0.5


class TestDeepfakeConfig:
    """Tests for DeepfakeConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = DeepfakeConfig()
        
        assert config.faceswap_threshold == 0.65
        assert config.expression_threshold == 0.6
        assert config.analyze_face_boundaries is True
        assert config.analyze_eye_reflection is True
        assert config.use_temporal_analysis is True
        assert config.min_frames_for_temporal == 10
    
    def test_blink_analysis_defaults(self):
        """Test blink analysis configuration."""
        config = DeepfakeConfig()
        
        assert config.analyze_blink_pattern is True
        assert config.natural_blink_rate_min == 10.0
        assert config.natural_blink_rate_max == 30.0


class TestArtifactAnalysisConfig:
    """Tests for ArtifactAnalysisConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = ArtifactAnalysisConfig()
        
        assert config.use_frequency_analysis is True
        assert config.fft_size == 256
        assert config.detect_checkerboard is True
        assert config.detect_blending is True
        assert config.analyze_texture_consistency is True
    
    def test_kernel_sizes(self):
        """Test blending kernel sizes."""
        config = ArtifactAnalysisConfig()
        
        assert config.blending_kernel_sizes == [3, 5, 7]


class TestVEIDIntegrationConfig:
    """Tests for VEIDIntegrationConfig."""
    
    def test_default_weights(self):
        """Test default weight values."""
        config = VEIDIntegrationConfig()
        
        assert config.gan_detection_weight == 0.30
        assert config.deepfake_weight == 0.35
        assert config.artifact_weight == 0.20
        assert config.consistency_weight == 0.15
    
    def test_penalty_values(self):
        """Test penalty configuration."""
        config = VEIDIntegrationConfig()
        
        assert config.synthetic_detected_penalty == 5000
        assert config.deepfake_detected_penalty == 7500
        assert config.high_confidence_penalty == 10000
    
    def test_threshold_values(self):
        """Test threshold configuration."""
        config = VEIDIntegrationConfig()
        
        assert config.synthetic_detection_threshold == 0.6
        assert config.high_confidence_threshold == 0.85
        assert config.low_confidence_threshold == 0.4


class TestGANDetectionConfig:
    """Tests for main GANDetectionConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = GANDetectionConfig()
        
        assert config.mode == DetectionMode.FULL
        assert config.enable_gpu is False
        assert config.enforce_determinism is True
        assert config.random_seed == 42
    
    def test_component_configs(self):
        """Test component configurations are created."""
        config = GANDetectionConfig()
        
        assert isinstance(config.discriminator, DiscriminatorConfig)
        assert isinstance(config.deepfake, DeepfakeConfig)
        assert isinstance(config.artifact, ArtifactAnalysisConfig)
        assert isinstance(config.veid_integration, VEIDIntegrationConfig)
    
    def test_fast_mode(self):
        """Test fast mode configuration."""
        config = GANDetectionConfig.fast_mode()
        
        assert config.mode == DetectionMode.FAST
        assert config.deepfake.use_temporal_analysis is False
        assert config.artifact.use_frequency_analysis is False
        assert config.discriminator.use_multi_scale is False
    
    def test_accurate_mode(self):
        """Test accurate mode configuration."""
        config = GANDetectionConfig.accurate_mode()
        
        assert config.mode == DetectionMode.ACCURATE
        assert config.deepfake.temporal_consistency_window == 10
        assert config.discriminator.num_blocks == 5
        assert len(config.artifact.blending_kernel_sizes) == 5
    
    def test_image_size_limits(self):
        """Test image size limit configuration."""
        config = GANDetectionConfig()
        
        assert config.min_image_size == (64, 64)
        assert config.max_image_size == (4096, 4096)
        assert config.min_face_size == 50
    
    def test_processing_limits(self):
        """Test processing limit configuration."""
        config = GANDetectionConfig()
        
        assert config.max_frames_per_sequence == 100
        assert config.processing_timeout_seconds == 30.0
