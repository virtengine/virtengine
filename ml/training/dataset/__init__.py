"""
Dataset handling for trust score training.

This module provides:
- Dataset ingestion from multiple sources
- Data preprocessing and validation
- Data augmentation for training
- PII anonymization for security
"""

from ml.training.dataset.ingestion import (
    DatasetIngestion,
    IdentitySample,
    Dataset,
    DatasetSplit,
)
from ml.training.dataset.preprocessing import (
    DatasetPreprocessor,
    PreprocessedSample,
)
from ml.training.dataset.augmentation import (
    DataAugmentation,
    AugmentedSample,
)
from ml.training.dataset.anonymization import (
    PIIAnonymizer,
    AnonymizationResult,
)

__all__ = [
    # Ingestion
    "DatasetIngestion",
    "IdentitySample",
    "Dataset",
    "DatasetSplit",
    # Preprocessing
    "DatasetPreprocessor",
    "PreprocessedSample",
    # Augmentation
    "DataAugmentation",
    "AugmentedSample",
    # Anonymization
    "PIIAnonymizer",
    "AnonymizationResult",
]
