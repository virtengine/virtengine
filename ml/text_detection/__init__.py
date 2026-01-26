"""
VirtEngine Text Detection Pipeline

This module provides text region of interest (ROI) detection for identity documents
using CRAFT (Character-Region Awareness for Text detection). It supports:
- Character-level text detection with region scores
- Word and line-level ROI aggregation via affinity scores
- Post-processing with thresholding and non-max suppression
- Versioned, reproducible detection outputs

Usage:
    from ml.text_detection import (
        TextDetectionPipeline,
        TextDetectionConfig,
        TextROI,
        TextDetectionResult,
    )
    
    config = TextDetectionConfig()
    pipeline = TextDetectionPipeline(config)
    result = pipeline.detect(preprocessed_document_image)
    
    for roi in result.rois:
        print(f"Text at {roi.bounding_box} with confidence {roi.confidence}")
"""

from ml.text_detection.config import (
    TextDetectionConfig,
    CRAFTConfig,
    PostProcessingConfig,
)
from ml.text_detection.roi_types import (
    Point,
    BoundingBox,
    TextROI,
    TextDetectionResult,
    TextType,
)
from ml.text_detection.craft_detector import CRAFTDetector
from ml.text_detection.postprocessing import TextPostProcessor
from ml.text_detection.pipeline import TextDetectionPipeline

__version__ = "1.0.0"
__all__ = [
    # Config
    "TextDetectionConfig",
    "CRAFTConfig",
    "PostProcessingConfig",
    # ROI Types
    "Point",
    "BoundingBox",
    "TextROI",
    "TextDetectionResult",
    "TextType",
    # Components
    "CRAFTDetector",
    "TextPostProcessor",
    # Pipeline
    "TextDetectionPipeline",
    # Version
    "__version__",
]
