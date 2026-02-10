"""
Configuration classes for the GAN fraud detection module.

VE-923: GAN fraud detection - synthetic image detection configuration

This module defines all configuration parameters for:
- Discriminator model settings
- Deepfake detection thresholds
- Artifact analysis parameters
- VEID score integration weights
"""

from dataclasses import dataclass, field
from typing import Tuple, Optional, List
from enum import Enum


class SyntheticImageType(str, Enum):
    """Types of synthetic/generated images to detect."""
    GAN_GENERATED = "gan_generated"
    DEEPFAKE_FACESWAP = "deepfake_faceswap"
    DEEPFAKE_EXPRESSION = "deepfake_expression"
    AI_GENERATED = "ai_generated"
    MORPHED = "morphed"
    MANIPULATED = "manipulated"
    UNKNOWN = "unknown"


class ArtifactType(str, Enum):
    """Types of GAN artifacts to detect."""
    FREQUENCY_ANOMALY = "frequency_anomaly"
    CHECKERBOARD = "checkerboard"
    BLENDING_BOUNDARY = "blending_boundary"
    TEXTURE_INCONSISTENCY = "texture_inconsistency"
    COLOR_MISMATCH = "color_mismatch"
    COMPRESSION_ARTIFACT = "compression_artifact"
    UPSAMPLING_ARTIFACT = "upsampling_artifact"
    EYE_ASYMMETRY = "eye_asymmetry"
    HAIR_BOUNDARY = "hair_boundary"
    BACKGROUND_INCONSISTENCY = "background_inconsistency"


class DetectionMode(str, Enum):
    """Detection mode for GAN detector."""
    FULL = "full"  # All detection methods
    FAST = "fast"  # Quick checks only
    ACCURATE = "accurate"  # Higher accuracy, slower


@dataclass
class DiscriminatorConfig:
    """Configuration for CNN discriminator model."""
    
    # Model architecture
    input_size: Tuple[int, int] = (224, 224)
    input_channels: int = 3
    base_filters: int = 64
    num_blocks: int = 4
    use_batch_norm: bool = True
    use_dropout: bool = True
    dropout_rate: float = 0.3
    
    # Feature extraction
    feature_extraction_layers: List[str] = field(
        default_factory=lambda: ["block1", "block2", "block3", "block4"]
    )
    use_multi_scale: bool = True
    
    # Classification head
    use_global_pool: bool = True
    fc_hidden_units: int = 512
    num_classes: int = 2  # Real vs Synthetic
    
    # Model path
    model_path: Optional[str] = None
    model_version: str = "1.0.0"


@dataclass
class DeepfakeConfig:
    """Configuration for deepfake detection."""
    
    # Face swap detection
    faceswap_threshold: float = 0.65
    faceswap_min_confidence: float = 0.7
    analyze_face_boundaries: bool = True
    boundary_smoothness_threshold: float = 0.3
    
    # Expression manipulation detection
    expression_threshold: float = 0.6
    analyze_eye_movement: bool = True
    analyze_mouth_movement: bool = True
    temporal_consistency_window: int = 5  # frames
    
    # Facial feature analysis
    analyze_eye_reflection: bool = True
    eye_reflection_threshold: float = 0.5
    analyze_skin_texture: bool = True
    skin_texture_threshold: float = 0.4
    
    # Temporal analysis (for video)
    use_temporal_analysis: bool = True
    min_frames_for_temporal: int = 10
    frame_rate: float = 30.0
    
    # Blinking analysis
    analyze_blink_pattern: bool = True
    natural_blink_rate_min: float = 10.0  # blinks per minute
    natural_blink_rate_max: float = 30.0
    
    # Audio-visual sync (if audio available)
    analyze_lip_sync: bool = False  # Disabled by default


@dataclass
class ArtifactAnalysisConfig:
    """Configuration for GAN artifact analysis."""
    
    # Frequency domain analysis
    use_frequency_analysis: bool = True
    fft_size: int = 256
    frequency_threshold: float = 0.4
    
    # Checkerboard artifact detection
    detect_checkerboard: bool = True
    checkerboard_threshold: float = 0.35
    
    # Blending boundary detection
    detect_blending: bool = True
    blending_kernel_sizes: List[int] = field(default_factory=lambda: [3, 5, 7])
    blending_threshold: float = 0.5
    
    # Texture consistency analysis
    analyze_texture_consistency: bool = True
    texture_patch_size: int = 32
    texture_threshold: float = 0.4
    
    # Color analysis
    analyze_color_consistency: bool = True
    color_histogram_bins: int = 64
    color_threshold: float = 0.45
    
    # Compression artifact detection
    detect_compression_artifacts: bool = True
    jpeg_quality_threshold: int = 70
    
    # Upsampling artifact detection
    detect_upsampling: bool = True
    upsampling_threshold: float = 0.4


@dataclass
class VEIDIntegrationConfig:
    """Configuration for VEID score integration."""
    
    # Score weights
    gan_detection_weight: float = 0.30
    deepfake_weight: float = 0.35
    artifact_weight: float = 0.20
    consistency_weight: float = 0.15
    
    # Thresholds for VEID impact
    synthetic_detection_threshold: float = 0.6  # Above this = likely synthetic
    high_confidence_threshold: float = 0.85
    low_confidence_threshold: float = 0.4
    
    # Score penalties
    synthetic_detected_penalty: int = 5000  # Basis points penalty
    deepfake_detected_penalty: int = 7500
    high_confidence_penalty: int = 10000  # Full rejection
    
    # Minimum score modifiers
    min_gan_score_impact: int = 0
    max_gan_score_impact: int = 10000  # Basis points


@dataclass
class GANDetectionConfig:
    """Main configuration for GAN fraud detection."""
    
    # Component configs
    discriminator: DiscriminatorConfig = field(default_factory=DiscriminatorConfig)
    deepfake: DeepfakeConfig = field(default_factory=DeepfakeConfig)
    artifact: ArtifactAnalysisConfig = field(default_factory=ArtifactAnalysisConfig)
    veid_integration: VEIDIntegrationConfig = field(default_factory=VEIDIntegrationConfig)
    
    # Detection mode
    mode: DetectionMode = DetectionMode.FULL
    
    # General settings
    enable_gpu: bool = False  # CPU-only for determinism
    batch_size: int = 1
    num_workers: int = 0  # No multiprocessing for determinism
    
    # Determinism settings
    random_seed: int = 42
    enforce_determinism: bool = True
    
    # Logging
    enable_detailed_logging: bool = False
    log_artifacts: bool = False
    
    # Quality requirements
    min_image_size: Tuple[int, int] = (64, 64)
    max_image_size: Tuple[int, int] = (4096, 4096)
    min_face_size: int = 50
    
    # Processing limits
    max_frames_per_sequence: int = 100
    processing_timeout_seconds: float = 30.0
    
    @classmethod
    def fast_mode(cls) -> "GANDetectionConfig":
        """Create config optimized for speed."""
        config = cls(mode=DetectionMode.FAST)
        config.deepfake.use_temporal_analysis = False
        config.artifact.use_frequency_analysis = False
        config.discriminator.use_multi_scale = False
        return config
    
    @classmethod
    def accurate_mode(cls) -> "GANDetectionConfig":
        """Create config optimized for accuracy."""
        config = cls(mode=DetectionMode.ACCURATE)
        config.deepfake.temporal_consistency_window = 10
        config.artifact.blending_kernel_sizes = [3, 5, 7, 9, 11]
        config.discriminator.num_blocks = 5
        return config
