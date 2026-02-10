"""
Tests for autoencoder anomaly detection configuration.

VE-924: Test configuration classes.
"""

import pytest

from ml.autoencoder_anomaly.config import (
    AutoencoderAnomalyConfig,
    EncoderConfig,
    DecoderConfig,
    ReconstructionConfig,
    AnomalyScoringConfig,
    VEIDIntegrationConfig,
    DetectionMode,
    AnomalyType,
    AnomalyLevel,
)


class TestEncoderConfig:
    """Tests for EncoderConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = EncoderConfig()
        
        assert config.input_size == (128, 128)
        assert config.input_channels == 3
        assert config.latent_dim == 128
        assert len(config.layer_filters) == 4
    
    def test_custom_values(self):
        """Test custom configuration values."""
        config = EncoderConfig(
            input_size=(256, 256),
            latent_dim=256,
            layer_filters=[64, 128, 256, 512],
        )
        
        assert config.input_size == (256, 256)
        assert config.latent_dim == 256
        assert config.layer_filters == [64, 128, 256, 512]


class TestDecoderConfig:
    """Tests for DecoderConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = DecoderConfig()
        
        assert config.output_channels == 3
        assert config.output_activation == "sigmoid"
        assert len(config.layer_filters) == 4
    
    def test_mirrors_encoder(self):
        """Test decoder mirrors encoder architecture."""
        encoder_config = EncoderConfig()
        decoder_config = DecoderConfig()
        
        # Decoder filters should be reverse of encoder
        expected = list(reversed(encoder_config.layer_filters))
        assert decoder_config.layer_filters == expected


class TestReconstructionConfig:
    """Tests for ReconstructionConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = ReconstructionConfig()
        
        assert config.use_mse is True
        assert config.use_mae is True
        assert config.use_ssim is True
        assert config.mse_threshold > 0
        assert config.ssim_threshold < 1.0
    
    def test_weight_sum(self):
        """Test that weights are reasonable."""
        config = ReconstructionConfig()
        
        total_weight = (
            config.mse_weight +
            config.mae_weight +
            config.ssim_weight +
            config.perceptual_weight
        )
        
        assert 0.9 <= total_weight <= 1.1  # Allow some tolerance


class TestAnomalyScoringConfig:
    """Tests for AnomalyScoringConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = AnomalyScoringConfig()
        
        assert config.normal_threshold < config.suspicious_threshold
        assert config.suspicious_threshold < config.anomaly_threshold
        assert 0 < config.normal_threshold < 1
        assert 0 < config.anomaly_threshold < 1
    
    def test_weight_sum(self):
        """Test that component weights sum to 1."""
        config = AnomalyScoringConfig()
        
        total_weight = (
            config.reconstruction_weight +
            config.latent_weight
        )
        
        assert abs(total_weight - 1.0) < 0.01


class TestVEIDIntegrationConfig:
    """Tests for VEIDIntegrationConfig."""
    
    def test_default_values(self):
        """Test default configuration values."""
        config = VEIDIntegrationConfig()
        
        assert config.low_anomaly_penalty >= 0
        assert config.medium_anomaly_penalty > config.low_anomaly_penalty
        assert config.high_anomaly_penalty > config.medium_anomaly_penalty
        assert config.critical_anomaly_penalty > config.high_anomaly_penalty
        assert config.max_veid_impact == 10000
    
    def test_penalty_bounds(self):
        """Test that penalties are within bounds."""
        config = VEIDIntegrationConfig()
        
        assert config.low_anomaly_penalty <= config.max_veid_impact
        assert config.medium_anomaly_penalty <= config.max_veid_impact
        assert config.high_anomaly_penalty <= config.max_veid_impact
        assert config.critical_anomaly_penalty <= config.max_veid_impact


class TestAutoencoderAnomalyConfig:
    """Tests for main AutoencoderAnomalyConfig."""
    
    def test_default_config(self):
        """Test default configuration."""
        config = AutoencoderAnomalyConfig()
        
        assert config.mode == DetectionMode.FULL
        assert config.enable_gpu is False
        assert config.enforce_determinism is True
        assert config.random_seed == 42
    
    def test_fast_mode(self):
        """Test fast mode factory method."""
        config = AutoencoderAnomalyConfig.fast_mode()
        
        assert config.mode == DetectionMode.FAST
        assert config.reconstruction.use_ssim is False
        assert config.reconstruction.use_patch_analysis is False
        assert config.scoring.use_latent_distance is False
    
    def test_strict_mode(self):
        """Test strict mode factory method."""
        config = AutoencoderAnomalyConfig.strict_mode()
        
        assert config.mode == DetectionMode.STRICT
        assert config.scoring.normal_threshold < AutoencoderAnomalyConfig().scoring.normal_threshold
        assert config.scoring.anomaly_threshold < AutoencoderAnomalyConfig().scoring.anomaly_threshold
    
    def test_nested_configs(self):
        """Test that nested configs are created."""
        config = AutoencoderAnomalyConfig()
        
        assert isinstance(config.encoder, EncoderConfig)
        assert isinstance(config.decoder, DecoderConfig)
        assert isinstance(config.reconstruction, ReconstructionConfig)
        assert isinstance(config.scoring, AnomalyScoringConfig)
        assert isinstance(config.veid_integration, VEIDIntegrationConfig)


class TestEnums:
    """Tests for enum types."""
    
    def test_detection_modes(self):
        """Test detection mode values."""
        assert DetectionMode.FULL.value == "full"
        assert DetectionMode.FAST.value == "fast"
        assert DetectionMode.STRICT.value == "strict"
    
    def test_anomaly_types(self):
        """Test anomaly type values."""
        assert AnomalyType.RECONSTRUCTION_HIGH.value == "reconstruction_high"
        assert AnomalyType.LATENT_OUTLIER.value == "latent_outlier"
    
    def test_anomaly_levels(self):
        """Test anomaly level values."""
        assert AnomalyLevel.NONE.value == "none"
        assert AnomalyLevel.LOW.value == "low"
        assert AnomalyLevel.MEDIUM.value == "medium"
        assert AnomalyLevel.HIGH.value == "high"
        assert AnomalyLevel.CRITICAL.value == "critical"
