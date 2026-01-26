"""
Configuration classes for the document preprocessing pipeline.

This module defines all configuration parameters for:
- Document standardization (format, resolution)
- Image enhancement (brightness, contrast, sharpness)
- Noise reduction settings
- Orientation detection parameters
- Perspective correction settings
"""

from dataclasses import dataclass, field
from typing import Tuple, Optional, List
from enum import Enum


class NoiseReductionMethod(str, Enum):
    """Supported noise reduction methods."""
    GAUSSIAN = "gaussian"
    MEDIAN = "median"
    BILATERAL = "bilateral"


class OutputFormat(str, Enum):
    """Supported output image formats."""
    PNG = "PNG"
    JPEG = "JPEG"
    TIFF = "TIFF"


class InterpolationMethod(str, Enum):
    """Interpolation methods for resizing."""
    NEAREST = "nearest"
    LINEAR = "linear"
    CUBIC = "cubic"
    LANCZOS = "lanczos"
    AREA = "area"


@dataclass
class StandardizationConfig:
    """Configuration for document standardization."""
    
    # Target resolution for ID documents
    target_width: int = 1024
    target_height: int = 768
    
    # Output format
    output_format: OutputFormat = OutputFormat.PNG
    
    # Minimum acceptable resolution
    min_width: int = 640
    min_height: int = 480
    
    # Maximum resolution (will downscale if larger)
    max_width: int = 4096
    max_height: int = 3072
    
    # Interpolation method for resizing
    interpolation: InterpolationMethod = InterpolationMethod.LANCZOS
    
    # Maintain aspect ratio when resizing
    maintain_aspect_ratio: bool = True
    
    # Padding color if aspect ratio maintained (BGR format)
    padding_color: Tuple[int, int, int] = (255, 255, 255)


@dataclass
class EnhancementConfig:
    """Configuration for image enhancement."""
    
    # CLAHE settings
    apply_clahe: bool = True
    clahe_clip_limit: float = 2.0
    clahe_tile_grid_size: Tuple[int, int] = (8, 8)
    
    # Brightness/contrast normalization
    apply_brightness_contrast: bool = True
    target_brightness: float = 128.0
    contrast_factor: float = 1.2
    
    # Auto-levels (histogram stretching)
    apply_auto_levels: bool = True
    auto_levels_percentile: float = 1.0  # Clip percentile
    
    # Sharpening (unsharp mask)
    apply_sharpening: bool = True
    sharpening_amount: float = 1.5
    sharpening_radius: int = 1
    sharpening_threshold: int = 0
    
    # Gamma correction
    apply_gamma: bool = False
    gamma_value: float = 1.0


@dataclass
class NoiseReductionConfig:
    """Configuration for noise reduction."""
    
    # Default method
    method: NoiseReductionMethod = NoiseReductionMethod.BILATERAL
    
    # Enable/disable noise reduction
    enabled: bool = True
    
    # Gaussian blur settings
    gaussian_kernel_size: int = 5
    gaussian_sigma: float = 1.0
    
    # Median blur settings
    median_kernel_size: int = 5
    
    # Bilateral filter settings (best for documents)
    bilateral_d: int = 9
    bilateral_sigma_color: float = 75.0
    bilateral_sigma_space: float = 75.0
    
    # Morphological noise reduction
    apply_morphological: bool = False
    morphological_kernel_size: int = 3


@dataclass
class OrientationConfig:
    """Configuration for orientation detection."""
    
    # Enable orientation detection
    enabled: bool = True
    
    # Rotation angles to try (degrees)
    rotation_angles: Tuple[int, ...] = (0, 90, 180, 270)
    
    # Detection method: "text" or "face" or "combined"
    detection_method: str = "combined"
    
    # Confidence threshold to accept detected orientation
    confidence_threshold: float = 0.5
    
    # Text detection settings (for orientation scoring)
    text_detection_scale: float = 0.5  # Scale down for faster detection
    
    # Face detection for orientation (detect upright face)
    use_face_detection: bool = True
    face_detection_confidence: float = 0.7


@dataclass
class PerspectiveConfig:
    """Configuration for perspective correction."""
    
    # Enable perspective correction
    enabled: bool = True
    
    # Edge detection settings
    canny_low_threshold: int = 50
    canny_high_threshold: int = 150
    
    # Contour detection settings
    min_contour_area_ratio: float = 0.1  # Min area as ratio of image
    max_contour_area_ratio: float = 0.95  # Max area as ratio of image
    
    # Corner detection tolerance
    corner_epsilon_ratio: float = 0.02  # For approxPolyDP
    
    # Margin to add around detected document
    output_margin: int = 10
    
    # Morphological operations for edge enhancement
    apply_morphology: bool = True
    morphology_kernel_size: int = 5
    
    # Gaussian blur before edge detection
    blur_kernel_size: int = 5


@dataclass
class DocumentConfig:
    """Master configuration for document preprocessing pipeline."""
    
    # Sub-configurations
    standardization: StandardizationConfig = field(
        default_factory=StandardizationConfig
    )
    enhancement: EnhancementConfig = field(
        default_factory=EnhancementConfig
    )
    noise_reduction: NoiseReductionConfig = field(
        default_factory=NoiseReductionConfig
    )
    orientation: OrientationConfig = field(
        default_factory=OrientationConfig
    )
    perspective: PerspectiveConfig = field(
        default_factory=PerspectiveConfig
    )
    
    # Pipeline order control
    correct_orientation_first: bool = True
    correct_perspectiVIRTENGINE_first: bool = True
    
    # Debug/logging
    debug_mode: bool = False
    saVIRTENGINE_intermediate_images: bool = False
    
    # Determinism settings for blockchain consensus
    random_seed: int = 42
    deterministic_mode: bool = True
    
    @classmethod
    def for_id_card(cls) -> "DocumentConfig":
        """Create config optimized for ID card processing."""
        config = cls()
        config.standardization.target_width = 1024
        config.standardization.target_height = 640
        return config
    
    @classmethod
    def for_passport(cls) -> "DocumentConfig":
        """Create config optimized for passport processing."""
        config = cls()
        config.standardization.target_width = 1024
        config.standardization.target_height = 720
        return config
    
    @classmethod
    def for_drivers_license(cls) -> "DocumentConfig":
        """Create config optimized for driver's license processing."""
        config = cls()
        config.standardization.target_width = 1024
        config.standardization.target_height = 640
        return config
