"""
Tests for liveness detection configuration.

VE-901: Test configuration classes and validation.
"""

import pytest
from ml.liveness_detection.config import (
    LivenessConfig,
    ActiveChallengeConfig,
    PassiveAnalysisConfig,
    SpoofDetectionConfig,
    ScoreConfig,
    ChallengeType,
    get_default_config,
    get_strict_config,
    get_permissive_config,
)


class TestLivenessConfig:
    """Tests for LivenessConfig class."""
    
    def test_default_config_creation(self):
        """Test creating default configuration."""
        config = LivenessConfig()
        
        assert config is not None
        assert isinstance(config.active, ActiveChallengeConfig)
        assert isinstance(config.passive, PassiveAnalysisConfig)
        assert isinstance(config.spoof, SpoofDetectionConfig)
        assert isinstance(config.score, ScoreConfig)
    
    def test_default_config_validation(self):
        """Test that default configuration is valid."""
        config = LivenessConfig()
        assert config.validate() is True
    
    def test_config_required_challenges(self):
        """Test default required challenges."""
        config = LivenessConfig()
        
        assert ChallengeType.BLINK.value in config.required_challenges
        assert len(config.required_challenges) >= 1
    
    def test_config_score_weights_sum_to_one(self):
        """Test that score weights sum to 1.0."""
        config = LivenessConfig()
        
        total = (
            config.score.active_challenge_weight +
            config.score.passive_analysis_weight +
            config.score.spoof_detection_weight
        )
        
        assert abs(total - 1.0) < 0.001
    
    def test_config_passive_weights_sum_to_one(self):
        """Test that passive weights sum to 1.0."""
        config = LivenessConfig()
        
        total = (
            config.passive.texture_weight +
            config.passive.depth_weight +
            config.passive.motion_weight +
            config.passive.reflection_weight +
            config.passive.moire_weight
        )
        
        assert abs(total - 1.0) < 0.001
    
    def test_invalid_score_weights(self):
        """Test validation fails with invalid score weights."""
        config = LivenessConfig()
        config.score.active_challenge_weight = 0.5
        config.score.passive_analysis_weight = 0.5
        config.score.spoof_detection_weight = 0.5  # Sum = 1.5
        
        with pytest.raises(ValueError, match="Score weights must sum to 1.0"):
            config.validate()
    
    def test_invalid_passive_weights(self):
        """Test validation fails with invalid passive weights."""
        config = LivenessConfig()
        config.passive.texture_weight = 0.5
        config.passive.depth_weight = 0.5
        config.passive.motion_weight = 0.5
        config.passive.reflection_weight = 0.5
        config.passive.moire_weight = 0.5  # Sum = 2.5
        
        with pytest.raises(ValueError, match="Passive weights must sum to 1.0"):
            config.validate()
    
    def test_invalid_threshold_range(self):
        """Test validation fails with invalid threshold."""
        config = LivenessConfig()
        config.score.pass_threshold = 1.5  # Invalid: > 1.0
        
        with pytest.raises(ValueError, match="pass_threshold must be between"):
            config.validate()
    
    def test_invalid_min_frames(self):
        """Test validation fails with invalid min frames."""
        config = LivenessConfig()
        config.min_frames_required = 0  # Invalid: < 1
        
        with pytest.raises(ValueError, match="min_frames_required must be >= 1"):
            config.validate()


class TestActiveChallengeConfig:
    """Tests for ActiveChallengeConfig class."""
    
    def test_default_blink_settings(self):
        """Test default blink detection settings."""
        config = ActiveChallengeConfig()
        
        assert config.blink_ear_threshold == 0.21
        assert config.blink_consecutive_frames == 2
        assert config.blink_min_duration_ms == 50
        assert config.blink_max_duration_ms == 500
    
    def test_default_smile_settings(self):
        """Test default smile detection settings."""
        config = ActiveChallengeConfig()
        
        assert config.smile_threshold == 0.5
        assert config.smile_lip_corner_ratio == 1.3
    
    def test_default_head_turn_settings(self):
        """Test default head turn detection settings."""
        config = ActiveChallengeConfig()
        
        assert config.head_turn_angle_threshold == 15.0
        assert config.head_turn_max_angle == 45.0


class TestSpoofDetectionConfig:
    """Tests for SpoofDetectionConfig class."""
    
    def test_default_thresholds(self):
        """Test default spoof detection thresholds."""
        config = SpoofDetectionConfig()
        
        assert config.spoof_score_threshold == 0.5
        assert config.high_confidence_threshold == 0.8
    
    def test_photo_detection_settings(self):
        """Test photo detection settings."""
        config = SpoofDetectionConfig()
        
        assert config.photo_print_texture_threshold > 0
        assert config.photo_print_color_saturation_min > 0


class TestConfigPresets:
    """Tests for configuration presets."""
    
    def test_get_default_config(self):
        """Test default config preset."""
        config = get_default_config()
        
        assert config.validate() is True
        assert config.score.pass_threshold == 0.75
    
    def test_get_strict_config(self):
        """Test strict config preset."""
        config = get_strict_config()
        
        assert config.validate() is True
        assert config.score.pass_threshold == 0.85
        assert len(config.required_challenges) >= 2
    
    def test_get_permissive_config(self):
        """Test permissive config preset."""
        config = get_permissive_config()
        
        assert config.validate() is True
        assert config.score.pass_threshold == 0.60
        assert len(config.required_challenges) == 1


class TestChallengeType:
    """Tests for ChallengeType enum."""
    
    def test_all_challenge_types(self):
        """Test all challenge types exist."""
        assert ChallengeType.BLINK.value == "blink"
        assert ChallengeType.SMILE.value == "smile"
        assert ChallengeType.HEAD_TURN_LEFT.value == "head_turn_left"
        assert ChallengeType.HEAD_TURN_RIGHT.value == "head_turn_right"
        assert ChallengeType.HEAD_NOD.value == "head_nod"
        assert ChallengeType.RAISE_EYEBROWS.value == "raise_eyebrows"
    
    def test_challenge_type_string_conversion(self):
        """Test challenge type string conversion."""
        assert str(ChallengeType.BLINK) == "ChallengeType.BLINK"
        assert ChallengeType.BLINK.value == "blink"
