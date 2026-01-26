"""
Configuration classes for the text detection pipeline.

This module defines all configuration parameters for:
- CRAFT model settings (thresholds, device, model paths)
- Post-processing settings (NMS thresholds, merging parameters)
- Pipeline execution settings
"""

from dataclasses import dataclass, field
from typing import Tuple, Optional, List
from enum import Enum


class DeviceType(str, Enum):
    """Compute device types for inference."""
    CPU = "cpu"
    CUDA = "cuda"
    MPS = "mps"  # Apple Silicon


class LinkRefinerMode(str, Enum):
    """Link refiner modes for CRAFT."""
    NONE = "none"
    REFINER = "refiner"


@dataclass
class CRAFTConfig:
    """Configuration for CRAFT text detector."""
    
    # Model settings
    model_path: Optional[str] = None  # Path to custom weights, None uses default
    refiner_path: Optional[str] = None  # Path to link refiner weights
    use_refiner: bool = False  # Whether to use link refiner
    
    # Inference device
    device: DeviceType = DeviceType.CPU
    
    # Input preprocessing
    canvas_size: int = 1280  # Maximum image dimension for inference
    mag_ratio: float = 1.5  # Magnification ratio for small text
    
    # Detection thresholds
    text_threshold: float = 0.7  # Threshold for text region score
    link_threshold: float = 0.4  # Threshold for link/affinity score
    low_text_threshold: float = 0.4  # Low threshold for connected components
    
    # Polygon settings
    poly: bool = False  # Return polygons instead of boxes
    
    # Reproducibility
    deterministic: bool = True  # Force deterministic operations
    
    def __post_init__(self):
        """Validate configuration values."""
        if self.canvas_size < 256:
            raise ValueError("canvas_size must be at least 256")
        if not 0.0 <= self.text_threshold <= 1.0:
            raise ValueError("text_threshold must be between 0.0 and 1.0")
        if not 0.0 <= self.link_threshold <= 1.0:
            raise ValueError("link_threshold must be between 0.0 and 1.0")
        if not 0.0 <= self.low_text_threshold <= 1.0:
            raise ValueError("low_text_threshold must be between 0.0 and 1.0")
        if self.mag_ratio <= 0:
            raise ValueError("mag_ratio must be positive")


@dataclass
class PostProcessingConfig:
    """Configuration for post-processing text detections."""
    
    # NMS settings
    nms_iou_threshold: float = 0.5  # IoU threshold for NMS
    
    # Score thresholding
    min_confidence: float = 0.5  # Minimum confidence to keep ROI
    
    # Character to word merging
    char_merge_x_threshold: float = 0.5  # X-distance threshold (fraction of char width)
    char_merge_y_threshold: float = 0.3  # Y-distance threshold (fraction of char height)
    
    # Word to line grouping
    line_y_threshold: float = 0.5  # Y-overlap threshold for same line
    line_x_gap_threshold: float = 2.0  # Max X gap between words (as fraction of avg word height)
    
    # Minimum sizes
    min_char_width: int = 5  # Minimum character box width in pixels
    min_char_height: int = 8  # Minimum character box height in pixels
    min_word_width: int = 15  # Minimum word box width in pixels
    min_line_width: int = 30  # Minimum line box width in pixels
    
    # Affinity score settings
    affinity_score_threshold: float = 0.3  # Min affinity to link characters
    
    def __post_init__(self):
        """Validate configuration values."""
        if not 0.0 <= self.nms_iou_threshold <= 1.0:
            raise ValueError("nms_iou_threshold must be between 0.0 and 1.0")
        if not 0.0 <= self.min_confidence <= 1.0:
            raise ValueError("min_confidence must be between 0.0 and 1.0")


@dataclass
class TextDetectionConfig:
    """Complete configuration for text detection pipeline."""
    
    # Component configs
    craft: CRAFTConfig = field(default_factory=CRAFTConfig)
    postprocessing: PostProcessingConfig = field(default_factory=PostProcessingConfig)
    
    # Pipeline settings
    compute_hash: bool = True  # Compute hash of input image
    return_character_boxes: bool = True  # Include character-level ROIs
    return_word_boxes: bool = True  # Include word-level ROIs
    return_line_boxes: bool = True  # Include line-level ROIs
    
    # Timing
    record_timing: bool = True  # Record processing time
    
    # Version tracking
    record_version: bool = True  # Record model version in output
