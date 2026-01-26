"""
VirtEngine Document Preprocessing Pipeline

This module provides document preprocessing for ID documents before OCR and
face extraction. It supports:
- Format and resolution standardization
- Brightness/contrast normalization (CLAHE)
- Noise reduction (Gaussian, median, bilateral)
- Orientation detection and correction
- Perspective correction using corner detection
- Deterministic execution for blockchain consensus

Usage:
    from ml.document_preprocessing import (
        DocumentPreprocessingPipeline,
        DocumentConfig,
    )
    
    config = DocumentConfig()
    pipeline = DocumentPreprocessingPipeline(config)
    result = pipeline.process(document_image)
"""

from ml.document_preprocessing.config import (
    DocumentConfig,
    StandardizationConfig,
    EnhancementConfig,
    NoiseReductionConfig,
    OrientationConfig,
    PerspectiveConfig,
)
from ml.document_preprocessing.standardization import DocumentStandardizer
from ml.document_preprocessing.enhancement import DocumentEnhancer
from ml.document_preprocessing.noise_reduction import NoiseReducer
from ml.document_preprocessing.orientation import OrientationDetector
from ml.document_preprocessing.perspective import PerspectiveCorrector
from ml.document_preprocessing.pipeline import (
    DocumentPreprocessingPipeline,
    PreprocessingResult,
)

__version__ = "1.0.0"
__all__ = [
    # Config
    "DocumentConfig",
    "StandardizationConfig",
    "EnhancementConfig",
    "NoiseReductionConfig",
    "OrientationConfig",
    "PerspectiveConfig",
    # Components
    "DocumentStandardizer",
    "DocumentEnhancer",
    "NoiseReducer",
    "OrientationDetector",
    "PerspectiveCorrector",
    # Pipeline
    "DocumentPreprocessingPipeline",
    "PreprocessingResult",
]
