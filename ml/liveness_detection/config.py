"""
Configuration classes for the liveness detection module.

VE-901: Liveness detection - anti-spoofing configuration

This module defines all configuration parameters for:
- Active challenge detection (blink, smile, head turn)
- Passive analysis (texture, depth, motion)
- Spoof detection (photo, screen, mask)
- Score computation and thresholds
"""

from dataclasses import dataclass, field
from typing import Tuple, Optional, List
from enum import Enum


class ChallengeType(str, Enum):
    """Types of active liveness challenges."""
    BLINK = "blink"
    SMILE = "smile"
    HEAD_TURN_LEFT = "head_turn_left"
    HEAD_TURN_RIGHT = "head_turn_right"
    HEAD_NOD = "head_nod"
    RAISE_EYEBROWS = "raise_eyebrows"


class PassiveFeatureType(str, Enum):
    """Types of passive liveness features to analyze."""
    TEXTURE = "texture"
    DEPTH = "depth"
    MOTION = "motion"
    REFLECTION = "reflection"
    MOIRE_PATTERN = "moire_pattern"
    FREQUENCY = "frequency"


class SpoofType(str, Enum):
    """Types of spoofing attacks to detect."""
    PHOTO_PRINT = "photo_print"
    PHOTO_SCREEN = "photo_screen"
    VIDEO_REPLAY = "video_replay"
    MASK_2D = "mask_2d"
    MASK_3D = "mask_3d"
    DEEPFAKE = "deepfake"


@dataclass
class ActiveChallengeConfig:
    """Configuration for active liveness challenge detection."""
    
    # Blink detection settings
    blink_ear_threshold: float = 0.21  # Eye aspect ratio threshold
    blink_consecutiVIRTENGINE_frames: int = 2  # Frames eye must be closed
    blink_min_duration_ms: int = 50
    blink_max_duration_ms: int = 500
    
    # Smile detection settings
    smile_threshold: float = 0.5  # Smile probability threshold
    smile_min_duration_ms: int = 200
    smile_lip_corner_ratio: float = 1.3  # Mouth width ratio for smile
    
    # Head turn detection settings
    head_turn_angle_threshold: float = 15.0  # Degrees
    head_turn_min_duration_ms: int = 300
    head_turn_max_angle: float = 45.0  # Maximum valid turn angle
    
    # Head nod detection settings
    head_nod_angle_threshold: float = 10.0  # Degrees
    head_nod_min_duration_ms: int = 250
    
    # Eyebrow raise detection settings
    eyebrow_raise_threshold: float = 0.15  # Distance ratio change
    eyebrow_raise_min_duration_ms: int = 200
    
    # Challenge timing
    challenge_timeout_seconds: float = 5.0
    min_frames_for_challenge: int = 10
    
    # Quality requirements
    min_face_confidence: float = 0.8
    max_face_angle: float = 30.0  # Max allowed face rotation


@dataclass
class PassiveAnalysisConfig:
    """Configuration for passive liveness analysis."""
    
    # Texture analysis (LBP - Local Binary Patterns)
    texture_lbp_radius: int = 1
    texture_lbp_points: int = 8
    texture_variance_threshold: float = 0.15
    texture_uniformity_threshold: float = 0.4
    
    # Depth estimation settings
    depth_gradient_threshold: float = 0.1
    depth_flatness_threshold: float = 0.3  # Too flat = photo
    depth_min_variation: float = 0.05
    
    # Motion analysis settings
    motion_optical_flow_threshold: float = 2.0
    motion_consistency_threshold: float = 0.7
    motion_min_frames: int = 5
    motion_natural_variation_min: float = 0.1
    motion_natural_variation_max: float = 10.0
    
    # Reflection analysis
    reflection_specular_threshold: float = 0.6
    reflection_glare_detection: bool = True
    
    # Moire pattern detection
    moire_frequency_bands: Tuple[int, int] = (10, 100)
    moire_energy_threshold: float = 0.2
    
    # Frequency analysis (FFT-based)
    frequency_high_threshold: float = 0.3
    frequency_low_threshold: float = 0.1
    
    # Weights for passive score combination
    texture_weight: float = 0.25
    depth_weight: float = 0.20
    motion_weight: float = 0.25
    reflection_weight: float = 0.15
    moire_weight: float = 0.15


@dataclass
class SpoofDetectionConfig:
    """Configuration for spoof attack detection."""
    
    # Photo print detection
    photo_print_texture_threshold: float = 0.3
    photo_print_color_saturation_min: float = 0.2
    photo_print_edge_sharpness_threshold: float = 0.4
    
    # Screen display detection (phone, tablet, monitor)
    screen_moire_threshold: float = 0.25
    screen_color_banding_threshold: float = 0.3
    screen_reflection_pattern_threshold: float = 0.4
    screen_frequency_signature_threshold: float = 0.35
    
    # Video replay detection
    video_temporal_consistency_threshold: float = 0.8
    video_frame_rate_analysis: bool = True
    video_compression_artifact_threshold: float = 0.3
    
    # 2D mask detection
    mask_2d_edge_threshold: float = 0.35
    mask_2d_skin_texture_threshold: float = 0.4
    mask_2d_symmetry_threshold: float = 0.7
    
    # 3D mask detection
    mask_3d_depth_uniformity_threshold: float = 0.5
    mask_3d_skin_response_threshold: float = 0.4
    mask_3d_micro_expression_threshold: float = 0.3
    
    # Deepfake detection
    deepfake_artifact_threshold: float = 0.25
    deepfake_temporal_coherence_threshold: float = 0.8
    deepfake_face_boundary_threshold: float = 0.3
    
    # Overall thresholds
    spoof_score_threshold: float = 0.5  # Above = likely spoof
    high_confidence_threshold: float = 0.8


@dataclass
class ScoreConfig:
    """Configuration for liveness score computation."""
    
    # Component weights (must sum to 1.0)
    actiVIRTENGINE_challenge_weight: float = 0.40
    passiVIRTENGINE_analysis_weight: float = 0.35
    spoof_detection_weight: float = 0.25
    
    # Thresholds
    pass_threshold: float = 0.75  # Minimum score to pass
    high_confidence_threshold: float = 0.90
    low_confidence_threshold: float = 0.50
    
    # Penalties
    single_challenge_fail_penalty: float = 0.15
    multiple_challenge_fail_penalty: float = 0.30
    spoof_detected_penalty: float = 0.50
    
    # Bonuses
    all_challenges_pass_bonus: float = 0.05
    natural_motion_bonus: float = 0.03


@dataclass
class LivenessConfig:
    """Master configuration for liveness detection."""
    
    # Sub-configurations
    active: ActiveChallengeConfig = field(default_factory=ActiveChallengeConfig)
    passive: PassiveAnalysisConfig = field(default_factory=PassiveAnalysisConfig)
    spoof: SpoofDetectionConfig = field(default_factory=SpoofDetectionConfig)
    score: ScoreConfig = field(default_factory=ScoreConfig)
    
    # General settings
    min_frames_required: int = 10
    max_frames_allowed: int = 300
    target_fps: float = 30.0
    
    # Face detection requirements
    min_face_size: int = 80  # Minimum face size in pixels
    max_face_size: int = 1000
    face_detection_confidence: float = 0.8
    
    # Required challenges (subset of ChallengeType)
    required_challenges: List[str] = field(default_factory=lambda: [
        ChallengeType.BLINK.value,
    ])
    optional_challenges: List[str] = field(default_factory=lambda: [
        ChallengeType.SMILE.value,
        ChallengeType.HEAD_TURN_LEFT.value,
    ])
    
    # Determinism settings
    random_seed: int = 42
    use_deterministic_ops: bool = True
    
    # Debug settings
    debug_mode: bool = False
    saVIRTENGINE_debug_frames: bool = False
    
    def validate(self) -> bool:
        """Validate configuration values."""
        # Check weight sums
        passiVIRTENGINE_weights = (
            self.passive.texture_weight +
            self.passive.depth_weight +
            self.passive.motion_weight +
            self.passive.reflection_weight +
            self.passive.moire_weight
        )
        if abs(passiVIRTENGINE_weights - 1.0) > 0.001:
            raise ValueError(f"Passive weights must sum to 1.0, got {passiVIRTENGINE_weights}")
        
        score_weights = (
            self.score.actiVIRTENGINE_challenge_weight +
            self.score.passiVIRTENGINE_analysis_weight +
            self.score.spoof_detection_weight
        )
        if abs(score_weights - 1.0) > 0.001:
            raise ValueError(f"Score weights must sum to 1.0, got {score_weights}")
        
        # Check threshold ranges
        if not 0.0 <= self.score.pass_threshold <= 1.0:
            raise ValueError("pass_threshold must be between 0.0 and 1.0")
        
        if self.min_frames_required < 1:
            raise ValueError("min_frames_required must be >= 1")
        
        return True


def get_default_config() -> LivenessConfig:
    """Get the default liveness detection configuration."""
    return LivenessConfig()


def get_strict_config() -> LivenessConfig:
    """Get a strict liveness detection configuration for high-security scenarios."""
    config = LivenessConfig()
    
    # More required challenges
    config.required_challenges = [
        ChallengeType.BLINK.value,
        ChallengeType.SMILE.value,
        ChallengeType.HEAD_TURN_LEFT.value,
    ]
    
    # Higher thresholds
    config.score.pass_threshold = 0.85
    config.spoof.spoof_score_threshold = 0.4
    
    # More frames required
    config.min_frames_required = 20
    
    return config


def get_permissiVIRTENGINE_config() -> LivenessConfig:
    """Get a permissive configuration for accessibility or low-risk scenarios."""
    config = LivenessConfig()
    
    # Fewer required challenges
    config.required_challenges = [ChallengeType.BLINK.value]
    config.optional_challenges = []
    
    # Lower thresholds
    config.score.pass_threshold = 0.60
    config.spoof.spoof_score_threshold = 0.6
    
    # Fewer frames required
    config.min_frames_required = 5
    
    return config
