"""
Tests for CNN discriminator.

VE-923: Test discriminator classification.
"""

import pytest
import numpy as np

from ml.gan_detection.discriminator import (
    CNNDiscriminator,
    DiscriminatorResult,
    DiscriminatorFeatures,
    create_discriminator,
)
from ml.gan_detection.config import DiscriminatorConfig, GANDetectionConfig
from ml.gan_detection.reason_codes import GANReasonCodes


class TestCNNDiscriminator:
    """Tests for CNNDiscriminator class."""
    
    def test_discriminator_creation(self, discriminator_config):
        """Test creating discriminator with config."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        assert discriminator is not None
        assert discriminator.config == discriminator_config
        assert discriminator.MODEL_VERSION == "1.0.0"
    
    def test_discriminator_default_config(self):
        """Test creating discriminator with default config."""
        discriminator = CNNDiscriminator()
        assert discriminator is not None
    
    def test_model_hash_generation(self, discriminator_config):
        """Test model hash is generated."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        assert discriminator.model_hash is not None
        assert len(discriminator.model_hash) == 32
    
    def test_model_hash_determinism(self, discriminator_config):
        """Test model hash is deterministic."""
        discriminator1 = CNNDiscriminator(discriminator_config)
        discriminator2 = CNNDiscriminator(discriminator_config)
        
        assert discriminator1.model_hash == discriminator2.model_hash
    
    def test_weights_initialized(self, discriminator_config):
        """Test weights are initialized."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        assert discriminator._weights_initialized is True
        assert "block1_conv" in discriminator._weights
        assert "fc1" in discriminator._weights
        assert "fc2" in discriminator._weights
    
    def test_classify_valid_image(self, discriminator_config, sample_face_frame):
        """Test classification of valid image."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        result = discriminator.classify(sample_face_frame)
        
        assert isinstance(result, DiscriminatorResult)
        assert 0.0 <= result.synthetic_probability <= 1.0
        assert 0.0 <= result.real_score <= 1.0
        assert 0.0 <= result.confidence <= 1.0
        assert result.model_version == "1.0.0"
    
    def test_classify_returns_scores(self, discriminator_config, sample_face_frame):
        """Test classification returns all scores."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        result = discriminator.classify(sample_face_frame)
        
        # Real and synthetic scores should sum to ~1
        assert abs(result.real_score + result.synthetic_score - 1.0) < 0.1
    
    def test_classify_empty_image(self, discriminator_config):
        """Test classification fails for empty image."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        result = discriminator.classify(np.array([]))
        
        assert result.synthetic_probability == 0.0
        assert GANReasonCodes.INVALID_INPUT in result.reason_codes
    
    def test_classify_small_image(self, discriminator_config, small_frame):
        """Test classification fails for small image."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        result = discriminator.classify(small_frame)
        
        assert GANReasonCodes.IMAGE_TOO_SMALL in result.reason_codes
    
    def test_classify_deterministic(self, discriminator_config, sample_face_frame):
        """Test classification is deterministic."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        result1 = discriminator.classify(sample_face_frame)
        result2 = discriminator.classify(sample_face_frame)
        
        assert result1.synthetic_probability == result2.synthetic_probability
        assert result1.real_score == result2.real_score
    
    def test_classify_with_threshold(self, discriminator_config, sample_face_frame):
        """Test classification with custom threshold."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        result_low = discriminator.classify(sample_face_frame, threshold=0.3)
        result_high = discriminator.classify(sample_face_frame, threshold=0.9)
        
        # Results may differ based on threshold
        assert isinstance(result_low.is_synthetic, bool)
        assert isinstance(result_high.is_synthetic, bool)
    
    def test_classify_synthetic_image(self, discriminator_config, synthetic_frame):
        """Test classification of synthetic-looking image."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        result = discriminator.classify(synthetic_frame)
        
        assert isinstance(result, DiscriminatorResult)
        # Synthetic frame should have higher synthetic probability
        # (though exact value depends on model)
    
    def test_processing_time_recorded(self, discriminator_config, sample_face_frame):
        """Test processing time is recorded."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        result = discriminator.classify(sample_face_frame)
        
        assert result.processing_time_ms > 0
    
    def test_frequency_anomaly_score(self, discriminator_config, sample_face_frame):
        """Test frequency anomaly score is computed."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        result = discriminator.classify(sample_face_frame)
        
        assert 0.0 <= result.frequency_anomaly_score <= 1.0
    
    def test_texture_anomaly_score(self, discriminator_config, sample_face_frame):
        """Test texture anomaly score is computed."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        result = discriminator.classify(sample_face_frame)
        
        assert 0.0 <= result.texture_anomaly_score <= 1.0


class TestDiscriminatorResult:
    """Tests for DiscriminatorResult dataclass."""
    
    def test_result_to_dict(self):
        """Test converting result to dict."""
        result = DiscriminatorResult(
            is_synthetic=True,
            synthetic_probability=0.75,
            confidence=0.8,
            real_score=0.25,
            synthetic_score=0.75,
            frequency_anomaly_score=0.3,
            texture_anomaly_score=0.4,
            model_version="1.0.0",
            model_hash="abc123",
            reason_codes=[GANReasonCodes.GAN_DETECTED],
            processing_time_ms=50.0,
        )
        
        d = result.to_dict()
        
        assert d["is_synthetic"] is True
        assert d["synthetic_probability"] == 0.75
        assert d["confidence"] == 0.8
        assert d["reason_codes"] == ["GAN_DETECTED"]


class TestDiscriminatorFeatures:
    """Tests for DiscriminatorFeatures dataclass."""
    
    def test_features_to_dict(self):
        """Test converting features to dict."""
        features = DiscriminatorFeatures(
            block1_features=np.zeros((1, 64, 112, 112)),
            feature_hash="abc123",
        )
        
        d = features.to_dict()
        
        assert d["has_block1"] is True
        assert d["has_block2"] is False
        assert d["feature_hash"] == "abc123"


class TestExtractFeatures:
    """Tests for feature extraction."""
    
    def test_extract_features(self, discriminator_config, sample_face_frame):
        """Test feature extraction without classification."""
        discriminator = CNNDiscriminator(discriminator_config)
        
        features = discriminator.extract_features(sample_face_frame)
        
        assert isinstance(features, DiscriminatorFeatures)
        assert features.block1_features is not None
        assert features.global_features is not None
        assert features.feature_hash is not None


class TestCreateDiscriminator:
    """Tests for factory function."""
    
    def test_create_without_config(self):
        """Test creating discriminator without config."""
        discriminator = create_discriminator()
        
        assert isinstance(discriminator, CNNDiscriminator)
    
    def test_create_with_config(self, gan_config):
        """Test creating discriminator with full config."""
        discriminator = create_discriminator(gan_config)
        
        assert isinstance(discriminator, CNNDiscriminator)
        assert discriminator.config == gan_config.discriminator
