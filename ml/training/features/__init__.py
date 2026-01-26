"""
Feature extraction for trust score training.

This module provides:
- Face embedding feature extraction
- Document quality feature extraction
- OCR-based feature extraction
- Metadata feature extraction
- Feature combination and normalization
"""

from ml.training.features.face_features import (
    FaceFeatureExtractor,
    FaceFeatures,
)
from ml.training.features.doc_features import (
    DocumentFeatureExtractor,
    DocumentFeatures,
)
from ml.training.features.ocr_features import (
    OCRFeatureExtractor,
    OCRFeatures,
)
from ml.training.features.metadata_features import (
    MetadataFeatureExtractor,
    MetadataFeatures,
)
from ml.training.features.feature_combiner import (
    FeatureExtractor,
    FeatureVector,
    FeatureDataset,
)

__all__ = [
    # Face features
    "FaceFeatureExtractor",
    "FaceFeatures",
    # Document features
    "DocumentFeatureExtractor",
    "DocumentFeatures",
    # OCR features
    "OCRFeatureExtractor",
    "OCRFeatures",
    # Metadata features
    "MetadataFeatureExtractor",
    "MetadataFeatures",
    # Combiner
    "FeatureExtractor",
    "FeatureVector",
    "FeatureDataset",
]
