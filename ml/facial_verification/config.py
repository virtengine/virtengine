"""
Configuration classes for the facial verification pipeline.

This module defines all configuration parameters for:
- Model selection and versioning
- Verification thresholds
- Preprocessing settings
- Determinism controls
"""

from dataclasses import dataclass, field
from typing import Tuple, Optional, List
from enum import Enum


class ModelName(str, Enum):
    """Supported face recognition models."""
    VGG_FACE = "VGG-Face"
    FACENET = "Facenet"
    FACENET512 = "Facenet512"
    ARCFACE = "ArcFace"
    DLIB = "Dlib"
    SFACE = "SFace"


class DetectorBackend(str, Enum):
    """Supported face detection backends."""
    MTCNN = "mtcnn"
    RETINAFACE = "retinaface"
    OPENCV = "opencv"
    SSD = "ssd"
    DLIB = "dlib"


class DistanceMetric(str, Enum):
    """Distance metrics for face comparison."""
    COSINE = "cosine"
    EUCLIDEAN = "euclidean"
    EUCLIDEAN_L2 = "euclidean_l2"


@dataclass
class PreprocessingConfig:
    """Configuration for image preprocessing pipeline."""
    
    # Resolution settings
    target_resolution: Tuple[int, int] = (224, 224)
    min_resolution: Tuple[int, int] = (112, 112)
    max_resolution: Tuple[int, int] = (1024, 1024)
    
    # Color settings
    use_grayscale: bool = False
    normalize_color: bool = True
    
    # Enhancement settings
    apply_clahe: bool = True
    clahe_clip_limit: float = 2.0
    clahe_tile_grid_size: Tuple[int, int] = (8, 8)
    
    # Noise reduction
    noise_reduction: bool = True
    noise_reduction_method: str = "bilateral"  # "bilateral" or "gaussian"
    bilateral_d: int = 9
    bilateral_sigma_color: float = 75.0
    bilateral_sigma_space: float = 75.0
    gaussian_kernel_size: Tuple[int, int] = (5, 5)
    
    # Normalization
    pixel_normalization: str = "minmax"  # "minmax", "zscore", or "fixed"
    normalize_mean: Tuple[float, float, float] = (0.485, 0.456, 0.406)
    normalize_std: Tuple[float, float, float] = (0.229, 0.224, 0.225)


@dataclass
class DetectionConfig:
    """Configuration for face detection."""
    
    # Detection backend
    detector_backend: DetectorBackend = DetectorBackend.MTCNN
    
    # Detection parameters
    min_face_size: int = 20
    scale_factor: float = 0.709
    confidence_threshold: float = 0.9
    
    # Alignment
    align_faces: bool = True
    
    # Cropping
    crop_margin: float = 0.2  # 20% margin around detected face
    expand_percentage: int = 10
    
    # Multi-face handling
    allow_multiple_faces: bool = False
    select_largest_face: bool = True


@dataclass
class DeterminismConfig:
    """Configuration for deterministic execution."""
    
    # CPU enforcement
    force_cpu: bool = True
    
    # Random seed
    seed: int = 42
    
    # Deterministic operations
    deterministic_ops: bool = True
    disable_cudnn_benchmark: bool = True
    
    # Model verification
    verify_model_hash: bool = True
    expected_model_hash: Optional[str] = None
    
    # Result hashing
    hash_algorithm: str = "sha256"
    include_embeddings_in_hash: bool = True


@dataclass
class VerificationConfig:
    """Main configuration for face verification."""
    
    # Model settings
    model_name: ModelName = ModelName.VGG_FACE
    model_version: str = "1.0.0"
    model_weights_hash: str = ""
    distance_metric: DistanceMetric = DistanceMetric.COSINE
    
    # Decision thresholds
    match_threshold: float = 0.90  # >= 90% similarity = match
    borderline_lower: float = 0.85  # 85-90% = borderline
    reject_threshold: float = 0.70  # < 70% = definite no-match
    
    # Component configs
    preprocessing: PreprocessingConfig = field(default_factory=PreprocessingConfig)
    detection: DetectionConfig = field(default_factory=DetectionConfig)
    determinism: DeterminismConfig = field(default_factory=DeterminismConfig)
    
    # Quality thresholds
    min_image_quality_score: float = 0.3
    min_face_confidence: float = 0.9
    
    # Batch processing
    max_batch_size: int = 10
    
    # Logging
    log_embeddings: bool = False
    log_intermediate_results: bool = False
    
    def __post_init__(self):
        """Validate configuration after initialization."""
        self._validate_thresholds()
        self._validate_resolution()
    
    def _validate_thresholds(self) -> None:
        """Ensure thresholds are logically consistent."""
        if not (0.0 <= self.reject_threshold <= self.borderline_lower <= self.match_threshold <= 1.0):
            raise ValueError(
                f"Thresholds must satisfy: 0 <= reject ({self.reject_threshold}) "
                f"<= borderline ({self.borderline_lower}) <= match ({self.match_threshold}) <= 1"
            )
    
    def _validate_resolution(self) -> None:
        """Ensure resolution settings are valid."""
        min_w, min_h = self.preprocessing.min_resolution
        max_w, max_h = self.preprocessing.max_resolution
        target_w, target_h = self.preprocessing.target_resolution
        
        if not (min_w <= target_w <= max_w and min_h <= target_h <= max_h):
            raise ValueError(
                f"Target resolution {self.preprocessing.target_resolution} must be "
                f"between min {self.preprocessing.min_resolution} and max {self.preprocessing.max_resolution}"
            )
    
    def to_dict(self) -> dict:
        """Convert configuration to dictionary."""
        return {
            "model_name": self.model_name.value,
            "model_version": self.model_version,
            "model_weights_hash": self.model_weights_hash,
            "distance_metric": self.distance_metric.value,
            "match_threshold": self.match_threshold,
            "borderline_lower": self.borderline_lower,
            "reject_threshold": self.reject_threshold,
            "preprocessing": {
                "target_resolution": self.preprocessing.target_resolution,
                "use_grayscale": self.preprocessing.use_grayscale,
                "apply_clahe": self.preprocessing.apply_clahe,
                "noise_reduction": self.preprocessing.noise_reduction,
            },
            "detection": {
                "detector_backend": self.detection.detector_backend.value,
                "confidence_threshold": self.detection.confidence_threshold,
                "align_faces": self.detection.align_faces,
            },
            "determinism": {
                "force_cpu": self.determinism.force_cpu,
                "seed": self.determinism.seed,
                "deterministic_ops": self.determinism.deterministic_ops,
            },
        }
    
    @classmethod
    def from_dict(cls, data: dict) -> "VerificationConfig":
        """Create configuration from dictionary."""
        preprocessing = PreprocessingConfig(
            target_resolution=tuple(data.get("preprocessing", {}).get("target_resolution", (224, 224))),
            use_grayscale=data.get("preprocessing", {}).get("use_grayscale", False),
            apply_clahe=data.get("preprocessing", {}).get("apply_clahe", True),
            noise_reduction=data.get("preprocessing", {}).get("noise_reduction", True),
        )
        
        detection = DetectionConfig(
            detector_backend=DetectorBackend(
                data.get("detection", {}).get("detector_backend", "mtcnn")
            ),
            confidence_threshold=data.get("detection", {}).get("confidence_threshold", 0.9),
            align_faces=data.get("detection", {}).get("align_faces", True),
        )
        
        determinism = DeterminismConfig(
            force_cpu=data.get("determinism", {}).get("force_cpu", True),
            seed=data.get("determinism", {}).get("seed", 42),
            deterministic_ops=data.get("determinism", {}).get("deterministic_ops", True),
        )
        
        return cls(
            model_name=ModelName(data.get("model_name", "VGG-Face")),
            model_version=data.get("model_version", "1.0.0"),
            model_weights_hash=data.get("model_weights_hash", ""),
            distance_metric=DistanceMetric(data.get("distance_metric", "cosine")),
            match_threshold=data.get("match_threshold", 0.90),
            borderline_lower=data.get("borderline_lower", 0.85),
            reject_threshold=data.get("reject_threshold", 0.70),
            preprocessing=preprocessing,
            detection=detection,
            determinism=determinism,
        )


# Default configurations for different use cases
DEFAULT_CONFIG = VerificationConfig()

HIGH_SECURITY_CONFIG = VerificationConfig(
    model_name=ModelName.ARCFACE,
    match_threshold=0.95,
    borderline_lower=0.90,
    reject_threshold=0.80,
    detection=DetectionConfig(
        confidence_threshold=0.95,
        allow_multiple_faces=False,
    ),
)

BALANCED_CONFIG = VerificationConfig(
    model_name=ModelName.FACENET512,
    match_threshold=0.90,
    borderline_lower=0.85,
    reject_threshold=0.70,
)

PERMISSIVIRTENGINE_CONFIG = VerificationConfig(
    model_name=ModelName.VGG_FACE,
    match_threshold=0.85,
    borderline_lower=0.75,
    reject_threshold=0.60,
)
