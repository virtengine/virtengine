"""
Model components for trust score training.

This module provides:
- Model architecture (TrustScoreModel)
- Training loop (ModelTrainer)
- Evaluation (ModelEvaluator)
- Export for TensorFlow-Go inference (ModelExporter)
"""

from ml.training.model.architecture import TrustScoreModel, create_trust_score_model
from ml.training.model.training import ModelTrainer, TrainingResult
from ml.training.model.evaluation import ModelEvaluator, EvaluationMetrics
from ml.training.model.export import ModelExporter, ExportResult

__all__ = [
    # Architecture
    "TrustScoreModel",
    "create_trust_score_model",
    # Training
    "ModelTrainer",
    "TrainingResult",
    # Evaluation
    "ModelEvaluator",
    "EvaluationMetrics",
    # Export
    "ModelExporter",
    "ExportResult",
]
