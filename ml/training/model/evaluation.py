"""
Model evaluation for trust score prediction.

Computes comprehensive metrics for model performance assessment.
"""

import logging
import json
import time
from dataclasses import dataclass, field
from typing import List, Optional, Dict, Any, Tuple
from pathlib import Path

import numpy as np

from ml.training.config import ModelConfig
from ml.training.model.architecture import TrustScoreModel

logger = logging.getLogger(__name__)

# TensorFlow import
try:
    import tensorflow as tf
    TF_AVAILABLE = True
except ImportError:
    TF_AVAILABLE = False


@dataclass
class EvaluationMetrics:
    """Comprehensive evaluation metrics for the trust score model."""
    
    # Regression metrics
    mae: float = 0.0                 # Mean Absolute Error
    mse: float = 0.0                 # Mean Squared Error
    rmse: float = 0.0                # Root Mean Squared Error
    mape: float = 0.0                # Mean Absolute Percentage Error
    r2: float = 0.0                  # R-squared (coefficient of determination)
    
    # Accuracy within thresholds
    accuracy_5: float = 0.0          # Within 5 points
    accuracy_10: float = 0.0         # Within 10 points
    accuracy_15: float = 0.0         # Within 15 points
    accuracy_20: float = 0.0         # Within 20 points
    
    # Distribution analysis
    mean_prediction: float = 0.0
    std_prediction: float = 0.0
    mean_error: float = 0.0          # Mean signed error (bias)
    std_error: float = 0.0
    
    # Percentile errors
    p50_error: float = 0.0           # Median absolute error
    p90_error: float = 0.0           # 90th percentile error
    p95_error: float = 0.0           # 95th percentile error
    p99_error: float = 0.0           # 99th percentile error
    
    # Confusion matrix for binned scores (0-25, 25-50, 50-75, 75-100)
    confusion_matrix: np.ndarray = field(
        default_factory=lambda: np.zeros((4, 4))
    )
    
    # Sample counts
    num_samples: int = 0
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "mae": float(self.mae),
            "mse": float(self.mse),
            "rmse": float(self.rmse),
            "mape": float(self.mape),
            "r2": float(self.r2),
            "accuracy_5": float(self.accuracy_5),
            "accuracy_10": float(self.accuracy_10),
            "accuracy_15": float(self.accuracy_15),
            "accuracy_20": float(self.accuracy_20),
            "mean_prediction": float(self.mean_prediction),
            "std_prediction": float(self.std_prediction),
            "mean_error": float(self.mean_error),
            "std_error": float(self.std_error),
            "p50_error": float(self.p50_error),
            "p90_error": float(self.p90_error),
            "p95_error": float(self.p95_error),
            "p99_error": float(self.p99_error),
            "confusion_matrix": self.confusion_matrix.tolist(),
            "num_samples": self.num_samples,
        }
    
    def save(self, filepath: str) -> None:
        """Save metrics to JSON file."""
        Path(filepath).parent.mkdir(parents=True, exist_ok=True)
        with open(filepath, 'w') as f:
            json.dump(self.to_dict(), f, indent=2)
    
    @classmethod
    def load(cls, filepath: str) -> "EvaluationMetrics":
        """Load metrics from JSON file."""
        with open(filepath, 'r') as f:
            data = json.load(f)
        
        metrics = cls()
        for key, value in data.items():
            if key == "confusion_matrix":
                setattr(metrics, key, np.array(value))
            else:
                setattr(metrics, key, value)
        return metrics
    
    def summary(self) -> str:
        """Get a summary string of key metrics."""
        return (
            f"MAE: {self.mae:.4f}, RMSE: {self.rmse:.4f}, R²: {self.r2:.4f}\n"
            f"Accuracy@5: {self.accuracy_5:.1%}, @10: {self.accuracy_10:.1%}, "
            f"@20: {self.accuracy_20:.1%}\n"
            f"Samples: {self.num_samples}"
        )


class ModelEvaluator:
    """
    Evaluates trust score model performance.
    
    Computes:
    - Regression metrics (MAE, MSE, RMSE, R²)
    - Accuracy within various thresholds
    - Error distribution analysis
    - Confusion matrix for score bins
    """
    
    def __init__(self, config: Optional[ModelConfig] = None):
        """
        Initialize the evaluator.
        
        Args:
            config: Model configuration
        """
        self.config = config or ModelConfig()
        self._bin_edges = [0, 25, 50, 75, 100]
    
    def evaluate(
        self,
        model: TrustScoreModel,
        features: np.ndarray,
        labels: np.ndarray,
    ) -> EvaluationMetrics:
        """
        Evaluate model on test data.
        
        Args:
            model: Trained model
            features: Test features
            labels: True labels (trust scores)
            
        Returns:
            EvaluationMetrics with all computed metrics
        """
        if len(features) == 0 or len(labels) == 0:
            logger.warning("Empty test set, returning default metrics")
            return EvaluationMetrics()
        
        # Get predictions
        predictions = model.predict(features)
        predictions = predictions.flatten()
        
        # Compute metrics
        metrics = self._compute_metrics(predictions, labels)
        
        logger.info(f"Evaluation complete:\n{metrics.summary()}")
        
        return metrics
    
    def evaluate_predictions(
        self,
        predictions: np.ndarray,
        labels: np.ndarray,
    ) -> EvaluationMetrics:
        """
        Evaluate predictions directly (without model).
        
        Args:
            predictions: Predicted trust scores
            labels: True labels
            
        Returns:
            EvaluationMetrics
        """
        predictions = np.array(predictions).flatten()
        labels = np.array(labels).flatten()
        
        return self._compute_metrics(predictions, labels)

    def benchmark_latency(
        self,
        model: TrustScoreModel,
        input_dim: int,
        batch_size: int = 1,
        warmup_runs: int = 10,
        timed_runs: int = 50,
        seed: int = 42,
    ) -> Dict[str, float]:
        """
        Benchmark inference latency for the model.
        
        Returns:
            Dictionary with p50/p95/p99 latency in milliseconds.
        """
        if not TF_AVAILABLE:
            return {}
        
        rng = np.random.RandomState(seed)
        sample = rng.randn(batch_size, input_dim).astype(np.float32)
        
        # Warmup
        for _ in range(warmup_runs):
            model.predict(sample)
        
        # Timed runs
        durations_ms = []
        for _ in range(timed_runs):
            start = time.perf_counter()
            model.predict(sample)
            durations_ms.append((time.perf_counter() - start) * 1000.0)
        
        if not durations_ms:
            return {}
        
        return {
            "p50": float(np.percentile(durations_ms, 50)),
            "p95": float(np.percentile(durations_ms, 95)),
            "p99": float(np.percentile(durations_ms, 99)),
        }
    
    def _compute_metrics(
        self,
        predictions: np.ndarray,
        labels: np.ndarray,
    ) -> EvaluationMetrics:
        """Compute all metrics from predictions and labels."""
        n = len(labels)
        
        # Basic checks
        if n == 0:
            return EvaluationMetrics()
        
        # Errors
        errors = predictions - labels
        abs_errors = np.abs(errors)
        
        # Regression metrics
        mae = np.mean(abs_errors)
        mse = np.mean(errors ** 2)
        rmse = np.sqrt(mse)
        
        # MAPE (avoid division by zero)
        with np.errstate(divide='ignore', invalid='ignore'):
            ape = np.where(labels != 0, abs_errors / np.abs(labels), 0)
            mape = np.mean(ape) * 100
        
        # R-squared
        ss_res = np.sum(errors ** 2)
        ss_tot = np.sum((labels - np.mean(labels)) ** 2)
        r2 = 1 - (ss_res / ss_tot) if ss_tot != 0 else 0
        
        # Accuracy within thresholds
        accuracy_5 = np.mean(abs_errors <= 5)
        accuracy_10 = np.mean(abs_errors <= 10)
        accuracy_15 = np.mean(abs_errors <= 15)
        accuracy_20 = np.mean(abs_errors <= 20)
        
        # Distribution metrics
        mean_prediction = np.mean(predictions)
        std_prediction = np.std(predictions)
        mean_error = np.mean(errors)
        std_error = np.std(errors)
        
        # Percentile errors
        p50_error = np.percentile(abs_errors, 50)
        p90_error = np.percentile(abs_errors, 90)
        p95_error = np.percentile(abs_errors, 95)
        p99_error = np.percentile(abs_errors, 99)
        
        # Confusion matrix for binned scores
        confusion_matrix = self._compute_confusion_matrix(predictions, labels)
        
        return EvaluationMetrics(
            mae=mae,
            mse=mse,
            rmse=rmse,
            mape=mape,
            r2=r2,
            accuracy_5=accuracy_5,
            accuracy_10=accuracy_10,
            accuracy_15=accuracy_15,
            accuracy_20=accuracy_20,
            mean_prediction=mean_prediction,
            std_prediction=std_prediction,
            mean_error=mean_error,
            std_error=std_error,
            p50_error=p50_error,
            p90_error=p90_error,
            p95_error=p95_error,
            p99_error=p99_error,
            confusion_matrix=confusion_matrix,
            num_samples=n,
        )
    
    def _compute_confusion_matrix(
        self,
        predictions: np.ndarray,
        labels: np.ndarray,
    ) -> np.ndarray:
        """Compute confusion matrix for score bins."""
        n_bins = len(self._bin_edges) - 1
        confusion = np.zeros((n_bins, n_bins), dtype=np.int32)
        
        # Bin predictions and labels
        pred_bins = np.digitize(predictions, self._bin_edges[1:-1])
        label_bins = np.digitize(labels, self._bin_edges[1:-1])
        
        # Clip to valid range
        pred_bins = np.clip(pred_bins, 0, n_bins - 1)
        label_bins = np.clip(label_bins, 0, n_bins - 1)
        
        # Count
        for pred, label in zip(pred_bins, label_bins):
            confusion[label, pred] += 1
        
        return confusion
    
    def generate_report(
        self,
        metrics: EvaluationMetrics,
        output_path: Optional[str] = None,
    ) -> str:
        """
        Generate a detailed evaluation report.
        
        Args:
            metrics: Evaluation metrics
            output_path: Optional path to save report
            
        Returns:
            Report string
        """
        lines = [
            "=" * 60,
            "TRUST SCORE MODEL EVALUATION REPORT",
            "=" * 60,
            "",
            "REGRESSION METRICS",
            "-" * 40,
            f"  Mean Absolute Error (MAE):     {metrics.mae:.4f}",
            f"  Mean Squared Error (MSE):      {metrics.mse:.4f}",
            f"  Root Mean Squared Error:       {metrics.rmse:.4f}",
            f"  R² Score:                      {metrics.r2:.4f}",
            "",
            "ACCURACY WITHIN THRESHOLDS",
            "-" * 40,
            f"  Within ±5 points:              {metrics.accuracy_5:.1%}",
            f"  Within ±10 points:             {metrics.accuracy_10:.1%}",
            f"  Within ±15 points:             {metrics.accuracy_15:.1%}",
            f"  Within ±20 points:             {metrics.accuracy_20:.1%}",
            "",
            "ERROR DISTRIBUTION",
            "-" * 40,
            f"  Mean Error (bias):             {metrics.mean_error:.4f}",
            f"  Std Error:                     {metrics.std_error:.4f}",
            f"  Median Absolute Error:         {metrics.p50_error:.4f}",
            f"  90th Percentile Error:         {metrics.p90_error:.4f}",
            f"  95th Percentile Error:         {metrics.p95_error:.4f}",
            f"  99th Percentile Error:         {metrics.p99_error:.4f}",
            "",
            "PREDICTION DISTRIBUTION",
            "-" * 40,
            f"  Mean Prediction:               {metrics.mean_prediction:.4f}",
            f"  Std Prediction:                {metrics.std_prediction:.4f}",
            "",
            "CONFUSION MATRIX (Score Bins: 0-25, 25-50, 50-75, 75-100)",
            "-" * 40,
            "              Predicted",
            "Actual    0-25  25-50  50-75  75-100",
        ]
        
        bin_labels = ["0-25", "25-50", "50-75", "75-100"]
        for i, label in enumerate(bin_labels):
            row = metrics.confusion_matrix[i]
            row_str = "  ".join(f"{int(v):5d}" for v in row)
            lines.append(f"  {label:6s}  {row_str}")
        
        lines.extend([
            "",
            f"Total Samples: {metrics.num_samples}",
            "=" * 60,
        ])
        
        report = "\n".join(lines)
        
        if output_path:
            Path(output_path).parent.mkdir(parents=True, exist_ok=True)
            with open(output_path, 'w') as f:
                f.write(report)
            logger.info(f"Report saved to {output_path}")
        
        return report


def evaluate_model(
    model: TrustScoreModel,
    test_features: np.ndarray,
    test_labels: np.ndarray,
    report_path: Optional[str] = None,
) -> EvaluationMetrics:
    """
    Convenience function to evaluate a model.
    
    Args:
        model: Trained model
        test_features: Test features
        test_labels: Test labels
        report_path: Optional path to save evaluation report
        
    Returns:
        EvaluationMetrics
    """
    evaluator = ModelEvaluator()
    metrics = evaluator.evaluate(model, test_features, test_labels)
    
    if report_path:
        evaluator.generate_report(metrics, report_path)
    
    return metrics
