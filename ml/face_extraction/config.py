"""
Configuration classes for the face extraction pipeline.

This module defines all configuration parameters for:
- U-Net model settings
- Mask post-processing parameters
- Face cropping settings
- Embedding extraction configuration
"""

from dataclasses import dataclass, field
from typing import Tuple, Optional
from enum import Enum


class MorphologicalOperation(str, Enum):
    """Morphological operations for mask cleanup."""
    ERODE = "erode"
    DILATE = "dilate"
    OPEN = "open"
    CLOSE = "close"


class InterpolationMethod(str, Enum):
    """Interpolation methods for resizing."""
    NEAREST = "nearest"
    LINEAR = "linear"
    CUBIC = "cubic"
    LANCZOS = "lanczos"
    AREA = "area"


@dataclass
class UNetConfig:
    """Configuration for U-Net face segmentation model."""
    
    # Model paths
    model_path: Optional[str] = None
    weights_path: Optional[str] = None
    
    # Input configuration
    input_size: Tuple[int, int] = (256, 256)  # (H, W)
    input_channels: int = 3
    
    # Model architecture
    encoder_depth: int = 4
    initial_filters: int = 64
    use_batch_norm: bool = True
    
    # Inference settings
    batch_size: int = 1
    use_gpu: bool = False
    gpu_memory_limit: Optional[int] = None  # MB
    
    # Determinism
    ensure_determinism: bool = True
    random_seed: int = 42
    
    # Model version tracking
    model_version: str = "1.0.0"
    
    # Confidence thresholds
    min_confidence: float = 0.5
    high_confidence: float = 0.85


@dataclass
class MaskProcessingConfig:
    """Configuration for mask post-processing."""
    
    # Thresholding
    threshold: float = 0.5
    use_adaptive_threshold: bool = False
    adaptive_block_size: int = 11
    adaptive_constant: int = 2
    
    # Morphological operations
    apply_morphology: bool = True
    erosion_kernel_size: Tuple[int, int] = (3, 3)
    dilation_kernel_size: Tuple[int, int] = (5, 5)
    erosion_iterations: int = 1
    dilation_iterations: int = 2
    
    # Opening/closing for noise removal
    apply_opening: bool = True
    opening_kernel_size: Tuple[int, int] = (5, 5)
    apply_closing: bool = True
    closing_kernel_size: Tuple[int, int] = (7, 7)
    
    # Connected component analysis
    use_largest_component: bool = True
    min_component_area: int = 500  # pixels
    
    # Contour smoothing
    smooth_contours: bool = True
    smoothing_epsilon: float = 0.01  # Fraction of arc length
    
    # Fill holes in mask
    fill_holes: bool = True


@dataclass
class CropperConfig:
    """Configuration for face cropping."""
    
    # Margin around face (percentage)
    margin: float = 0.15
    margin_top: Optional[float] = None  # Override for top margin
    margin_bottom: Optional[float] = None
    margin_left: Optional[float] = None
    margin_right: Optional[float] = None
    
    # Output size
    output_size: Optional[Tuple[int, int]] = (224, 224)  # (H, W)
    maintain_aspect_ratio: bool = True
    
    # Padding
    padding_color: Tuple[int, int, int] = (0, 0, 0)  # BGR
    
    # Interpolation
    interpolation: InterpolationMethod = InterpolationMethod.LANCZOS
    
    # Minimum face size (percentage of image)
    min_face_percentage: float = 0.02
    max_face_percentage: float = 0.8
    
    # Quality checks
    min_face_width: int = 50
    min_face_height: int = 50


@dataclass
class EmbeddingConfig:
    """Configuration for embedding extraction."""
    
    # Use existing facial verification config
    use_facial_verification: bool = True
    
    # Model selection (if not using facial verification)
    model_name: str = "ArcFace"
    
    # Embedding normalization
    normalize_embedding: bool = True
    
    # Data minimization
    discard_face_image: bool = True
    
    # Embedding hash
    compute_hash: bool = True
    hash_algorithm: str = "sha256"


@dataclass
class FaceExtractionConfig:
    """Main configuration for the face extraction pipeline."""
    
    # Component configurations
    unet: UNetConfig = field(default_factory=UNetConfig)
    mask: MaskProcessingConfig = field(default_factory=MaskProcessingConfig)
    cropper: CropperConfig = field(default_factory=CropperConfig)
    embedding: EmbeddingConfig = field(default_factory=EmbeddingConfig)
    
    # Pipeline settings
    preprocess_document: bool = True
    return_intermediate_results: bool = False
    
    # Logging and debugging
    debug_mode: bool = False
    save_debug_images: bool = False
    debug_output_dir: Optional[str] = None
    
    # Quality thresholds
    min_extraction_confidence: float = 0.6
    min_face_quality: float = 0.5
    
    # Error handling
    raise_on_failure: bool = False
    fallback_to_detection: bool = True  # Use face detection if segmentation fails
