"""
VirtEngine ML Training Pipeline

This module provides the complete training pipeline for the trust score model
used in VEID identity verification. It supports:
- Dataset ingestion with multiple document types and face samples
- PII anonymization and secure handling
- Feature extraction (face embeddings, document features, OCR signals, metadata)
- Model training with configurable architectures
- Evaluation with reproducible metrics
- Export to TensorFlow SavedModel for Go inference

Usage:
    from ml.training import (
        TrainingConfig,
        DatasetIngestion,
        FeatureExtractor,
        TrustScoreModel,
        ModelExporter,
    )
    
    # Load config and dataset
    config = TrainingConfig.from_yaml("config.yaml")
    dataset = DatasetIngestion(config.dataset).load()
    
    # Extract features
    extractor = FeatureExtractor(config.features)
    features = extractor.extract_batch(dataset)
    
    # Train model
    model = TrustScoreModel(config.model)
    model.train(features.train, features.val)
    
    # Evaluate and export
    metrics = model.evaluate(features.test)
    ModelExporter().export_savedmodel(model, config.export_path, config.version)
"""

from ml.training.config import (
    TrainingConfig,
    DatasetConfig,
    FeatureConfig,
    ModelConfig,
    ExportConfig,
)
from ml.training.dataset.ingestion import DatasetIngestion
from ml.training.dataset.preprocessing import DatasetPreprocessor
from ml.training.dataset.augmentation import DataAugmentation
from ml.training.dataset.anonymization import PIIAnonymizer
from ml.training.features.feature_combiner import FeatureExtractor, FeatureVector
from ml.training.model.architecture import TrustScoreModel
from ml.training.model.training import ModelTrainer
from ml.training.model.evaluation import ModelEvaluator, EvaluationMetrics
from ml.training.model.export import ModelExporter, ExportResult

__version__ = "1.0.0"
__all__ = [
    # Config
    "TrainingConfig",
    "DatasetConfig",
    "FeatureConfig",
    "ModelConfig",
    "ExportConfig",
    # Dataset
    "DatasetIngestion",
    "DatasetPreprocessor",
    "DataAugmentation",
    "PIIAnonymizer",
    # Features
    "FeatureExtractor",
    "FeatureVector",
    # Model
    "TrustScoreModel",
    "ModelTrainer",
    "ModelEvaluator",
    "EvaluationMetrics",
    # Export
    "ModelExporter",
    "ExportResult",
    # Version
    "__version__",
]
