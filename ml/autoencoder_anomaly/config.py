"""
Configuration classes for the autoencoder anomaly detection module.

VE-924: Autoencoder anomaly detection - configuration

This module defines all configuration parameters for:
- Autoencoder model architecture
- Reconstruction error thresholds
- Anomaly scoring parameters
- VEID score integration weights
"""

from dataclasses import dataclass, field
from typing import Tuple, Optional, List
from enum import Enum


class AnomalyType(str, Enum):
    """Types of anomalies that can be detected."""
    RECONSTRUCTION_HIGH = "reconstruction_high"
    LATENT_OUTLIER = "latent_outlier"
    FEATURE_ANOMALY = "feature_anomaly"
    STATISTICAL_OUTLIER = "statistical_outlier"
    UNKNOWN = "unknown"


class AnomalyLevel(str, Enum):
    """Severity level of detected anomaly."""
    NONE = "none"
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class DetectionMode(str, Enum):
    """Detection mode for autoencoder."""
    FULL = "full"  # Complete analysis
    FAST = "fast"  # Quick checks only
    STRICT = "strict"  # Lower thresholds, more sensitive


@dataclass
class EncoderConfig:
    """Configuration for encoder network."""
    
    # Input specifications
    input_size: Tuple[int, int] = (128, 128)
    input_channels: int = 3
    
    # Architecture
    layer_filters: List[int] = field(
        default_factory=lambda: [32, 64, 128, 256]
    )
    kernel_size: int = 3
    stride: int = 2
    use_batch_norm: bool = True
    activation: str = "leaky_relu"
    leaky_relu_alpha: float = 0.2
    
    # Bottleneck
    latent_dim: int = 128
    use_variational: bool = False  # VAE mode
    
    # Regularization
    use_dropout: bool = True
    dropout_rate: float = 0.2


@dataclass
class DecoderConfig:
    """Configuration for decoder network."""
    
    # Architecture (mirrors encoder)
    layer_filters: List[int] = field(
        default_factory=lambda: [256, 128, 64, 32]
    )
    kernel_size: int = 3
    stride: int = 2
    use_batch_norm: bool = True
    activation: str = "leaky_relu"
    leaky_relu_alpha: float = 0.2
    
    # Output
    output_channels: int = 3
    output_activation: str = "sigmoid"


@dataclass
class ReconstructionConfig:
    """Configuration for reconstruction error calculation."""
    
    # Error metrics
    use_mse: bool = True
    use_mae: bool = True
    use_ssim: bool = True
    use_perceptual: bool = False  # Feature-based loss
    
    # Metric weights
    mse_weight: float = 0.4
    mae_weight: float = 0.2
    ssim_weight: float = 0.3
    perceptual_weight: float = 0.1
    
    # Per-channel analysis
    analyze_per_channel: bool = True
    
    # Patch-based analysis
    use_patch_analysis: bool = True
    patch_size: int = 16
    patch_stride: int = 8
    
    # Statistical thresholds
    mse_threshold: float = 0.05  # Above this is anomalous
    mae_threshold: float = 0.08
    ssim_threshold: float = 0.85  # Below this is anomalous


@dataclass
class AnomalyScoringConfig:
    """Configuration for anomaly scoring."""
    
    # Score thresholds
    normal_threshold: float = 0.3  # Below this = normal
    suspicious_threshold: float = 0.5  # Between normal and this = suspicious
    anomaly_threshold: float = 0.7  # Above this = definite anomaly
    
    # Score computation
    use_percentile_scoring: bool = True
    reference_percentile: float = 95.0  # 95th percentile of training data
    
    # Multi-modal scoring
    combine_reconstruction_latent: bool = True
    reconstruction_weight: float = 0.6
    latent_weight: float = 0.4
    
    # Latent space analysis
    use_latent_distance: bool = True
    latent_distance_method: str = "mahalanobis"  # "euclidean", "mahalanobis"
    
    # Adaptive thresholds
    use_adaptiVIRTENGINE_threshold: bool = False
    adaptation_rate: float = 0.01


@dataclass
class VEIDIntegrationConfig:
    """Configuration for VEID score integration."""
    
    # Score impact weights
    anomaly_detection_weight: float = 0.25
    reconstruction_weight: float = 0.35
    latent_analysis_weight: float = 0.25
    consistency_weight: float = 0.15
    
    # Anomaly impact on VEID
    low_anomaly_penalty: int = 500  # Basis points
    medium_anomaly_penalty: int = 1500
    high_anomaly_penalty: int = 3500
    critical_anomaly_penalty: int = 7500
    
    # Score adjustments
    min_veid_impact: int = 0
    max_veid_impact: int = 10000  # Basis points
    
    # Confidence weighting
    apply_confidence_scaling: bool = True
    min_confidence_for_impact: float = 0.6


@dataclass
class AutoencoderAnomalyConfig:
    """Main configuration for autoencoder anomaly detection."""
    
    # Component configs
    encoder: EncoderConfig = field(default_factory=EncoderConfig)
    decoder: DecoderConfig = field(default_factory=DecoderConfig)
    reconstruction: ReconstructionConfig = field(default_factory=ReconstructionConfig)
    scoring: AnomalyScoringConfig = field(default_factory=AnomalyScoringConfig)
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
    log_reconstructions: bool = False  # Never log actual image data
    
    # Quality requirements
    min_image_size: Tuple[int, int] = (64, 64)
    max_image_size: Tuple[int, int] = (4096, 4096)
    
    # Processing limits
    processing_timeout_seconds: float = 15.0
    
    # Model paths
    model_path: Optional[str] = None
    reference_stats_path: Optional[str] = None
    
    @classmethod
    def fast_mode(cls) -> "AutoencoderAnomalyConfig":
        """Create config optimized for speed."""
        config = cls(mode=DetectionMode.FAST)
        config.reconstruction.use_ssim = False
        config.reconstruction.use_patch_analysis = False
        config.scoring.use_latent_distance = False
        config.encoder.layer_filters = [32, 64, 128]
        config.decoder.layer_filters = [128, 64, 32]
        return config
    
    @classmethod
    def strict_mode(cls) -> "AutoencoderAnomalyConfig":
        """Create config with stricter anomaly detection."""
        config = cls(mode=DetectionMode.STRICT)
        config.scoring.normal_threshold = 0.2
        config.scoring.suspicious_threshold = 0.4
        config.scoring.anomaly_threshold = 0.6
        config.reconstruction.mse_threshold = 0.03
        config.reconstruction.ssim_threshold = 0.9
        return config
