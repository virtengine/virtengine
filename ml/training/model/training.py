"""
Model training for trust score prediction.

Provides the training loop with proper logging, checkpointing,
and reproducibility.
"""

import logging
import time
import os
import json
import hashlib
from dataclasses import dataclass, field
from typing import List, Optional, Dict, Any, Tuple
from pathlib import Path

import numpy as np

from ml.training.config import TrainingConfig, ModelConfig
from ml.training.features.feature_combiner import FeatureDataset
from ml.training.model.architecture import TrustScoreModel

logger = logging.getLogger(__name__)

# TensorFlow import
try:
    import tensorflow as tf
    TF_AVAILABLE = True
except ImportError:
    TF_AVAILABLE = False


@dataclass
class TrainingResult:
    """Result of model training."""
    
    # Model reference
    model: Optional[TrustScoreModel] = None
    
    # Training history
    history: Dict[str, List[float]] = field(default_factory=dict)
    
    # Best metrics
    best_val_loss: float = float("inf")
    best_val_mae: float = float("inf")
    best_epoch: int = 0
    
    # Final metrics
    final_train_loss: float = 0.0
    final_train_mae: float = 0.0
    final_val_loss: float = 0.0
    final_val_mae: float = 0.0
    
    # Training info
    total_epochs: int = 0
    early_stopped: bool = False
    training_time_seconds: float = 0.0
    
    # Model info
    model_hash: str = ""
    config_hash: str = ""
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "best_val_loss": self.best_val_loss,
            "best_val_mae": self.best_val_mae,
            "best_epoch": self.best_epoch,
            "final_train_loss": self.final_train_loss,
            "final_train_mae": self.final_train_mae,
            "final_val_loss": self.final_val_loss,
            "final_val_mae": self.final_val_mae,
            "total_epochs": self.total_epochs,
            "early_stopped": self.early_stopped,
            "training_time_seconds": self.training_time_seconds,
            "model_hash": self.model_hash,
            "config_hash": self.config_hash,
        }
    
    def save_history(self, filepath: str) -> None:
        """Save training history to JSON."""
        Path(filepath).parent.mkdir(parents=True, exist_ok=True)
        with open(filepath, 'w') as f:
            json.dump({
                "history": self.history,
                "result": self.to_dict(),
            }, f, indent=2)


class ModelTrainer:
    """
    Handles model training with proper configuration and logging.
    
    Features:
    - Reproducible training with seed setting
    - Automatic checkpointing
    - Early stopping
    - Learning rate scheduling
    - TensorBoard logging
    - Training result tracking
    """
    
    def __init__(self, config: Optional[TrainingConfig] = None):
        """
        Initialize the trainer.
        
        Args:
            config: Training configuration
        """
        self.config = config or TrainingConfig()
        self._setup_reproducibility()
    
    def _setup_reproducibility(self) -> None:
        """Set up reproducibility settings."""
        seed = self.config.random_seed
        
        # Set Python random seed
        import random
        random.seed(seed)
        
        # Set NumPy seed
        np.random.seed(seed)
        
        # Set TensorFlow seed
        if TF_AVAILABLE:
            tf.random.set_seed(seed)
            
            # Enable deterministic operations if requested
            if self.config.deterministic:
                try:
                    tf.config.experimental.enable_op_determinism()
                except Exception:
                    logger.warning("Could not enable TensorFlow determinism")
        
        logger.info(f"Set random seed to {seed}")
    
    def train(
        self,
        feature_dataset: FeatureDataset,
        model: Optional[TrustScoreModel] = None,
    ) -> TrainingResult:
        """
        Train a model on the feature dataset.
        
        Args:
            feature_dataset: Dataset with extracted features
            model: Model to train (creates new if not provided)
            
        Returns:
            TrainingResult with trained model and metrics
        """
        if not TF_AVAILABLE:
            raise ImportError("TensorFlow is required for training")
        
        start_time = time.time()
        
        # Create model if not provided
        if model is None:
            model = TrustScoreModel(self.config.model)
        
        # Compile model
        model.compile()
        
        # Get training data
        train_features, train_labels = feature_dataset.get_arrays("train")
        val_features, val_labels = feature_dataset.get_arrays("validation")
        
        logger.info(
            f"Training on {len(train_labels)} samples, "
            f"validating on {len(val_labels)} samples"
        )
        
        # Create output directories
        self._setup_directories()
        
        # Train
        history = model.train(
            train_features,
            train_labels,
            val_features,
            val_labels,
            epochs=self.config.model.epochs,
            batch_size=self.config.model.batch_size,
            verbose=1,
        )
        
        # Compute result
        training_time = time.time() - start_time
        result = self._build_result(model, history, training_time)
        
        # Save training result
        self._save_result(result)
        
        return result
    
    def train_from_arrays(
        self,
        train_features: np.ndarray,
        train_labels: np.ndarray,
        val_features: Optional[np.ndarray] = None,
        val_labels: Optional[np.ndarray] = None,
        model: Optional[TrustScoreModel] = None,
    ) -> TrainingResult:
        """
        Train a model on numpy arrays.
        
        Args:
            train_features: Training features
            train_labels: Training labels
            val_features: Validation features
            val_labels: Validation labels
            model: Model to train
            
        Returns:
            TrainingResult with trained model and metrics
        """
        if not TF_AVAILABLE:
            raise ImportError("TensorFlow is required for training")
        
        start_time = time.time()
        
        # Create model if not provided
        if model is None:
            model = TrustScoreModel(self.config.model)
        
        # Compile model
        model.compile()
        
        # Create output directories
        self._setup_directories()
        
        # Train
        history = model.train(
            train_features,
            train_labels,
            val_features,
            val_labels,
            epochs=self.config.model.epochs,
            batch_size=self.config.model.batch_size,
            verbose=1,
        )
        
        # Compute result
        training_time = time.time() - start_time
        result = self._build_result(model, history, training_time)
        
        # Save training result
        self._save_result(result)
        
        return result
    
    def _setup_directories(self) -> None:
        """Create necessary output directories."""
        dirs = [
            self.config.model.checkpoint_dir,
            self.config.model.tensorboard_log_dir,
        ]
        
        for dir_path in dirs:
            if dir_path:
                Path(dir_path).mkdir(parents=True, exist_ok=True)
    
    def _build_result(
        self,
        model: TrustScoreModel,
        history: Dict[str, List[float]],
        training_time: float,
    ) -> TrainingResult:
        """Build training result from history."""
        result = TrainingResult(
            model=model,
            history=history,
            training_time_seconds=training_time,
            total_epochs=len(history.get("loss", [])),
        )
        
        # Find best validation metrics
        if "val_loss" in history:
            val_losses = history["val_loss"]
            result.best_epoch = int(np.argmin(val_losses))
            result.best_val_loss = float(val_losses[result.best_epoch])
            
            if "val_mae" in history:
                result.best_val_mae = float(history["val_mae"][result.best_epoch])
        
        # Get final metrics
        if "loss" in history and len(history["loss"]) > 0:
            result.final_train_loss = float(history["loss"][-1])
        if "mae" in history and len(history["mae"]) > 0:
            result.final_train_mae = float(history["mae"][-1])
        if "val_loss" in history and len(history["val_loss"]) > 0:
            result.final_val_loss = float(history["val_loss"][-1])
        if "val_mae" in history and len(history["val_mae"]) > 0:
            result.final_val_mae = float(history["val_mae"][-1])
        
        # Check for early stopping
        result.early_stopped = (
            result.total_epochs < self.config.model.epochs
        )
        
        # Compute model hash
        result.model_hash = self._compute_model_hash(model)
        result.config_hash = self._compute_config_hash()
        
        return result
    
    def _compute_model_hash(self, model: TrustScoreModel) -> str:
        """Compute hash of model weights."""
        weights = model.get_weights()
        
        # Concatenate all weights into single array
        flat_weights = np.concatenate([w.flatten() for w in weights])
        
        # Compute hash
        return hashlib.sha256(flat_weights.tobytes()).hexdigest()[:16]
    
    def _compute_config_hash(self) -> str:
        """Compute hash of training configuration."""
        config_str = json.dumps(self.config.to_dict(), sort_keys=True)
        return hashlib.sha256(config_str.encode()).hexdigest()[:16]
    
    def _save_result(self, result: TrainingResult) -> None:
        """Save training result to disk."""
        output_dir = Path(self.config.model.checkpoint_dir or "training_output")
        output_dir.mkdir(parents=True, exist_ok=True)
        
        # Save history
        history_path = output_dir / "training_history.json"
        result.save_history(str(history_path))
        
        logger.info(f"Training completed in {result.training_time_seconds:.2f}s")
        logger.info(f"Best validation loss: {result.best_val_loss:.4f} at epoch {result.best_epoch}")


def train_model(
    config: TrainingConfig,
    feature_dataset: FeatureDataset,
) -> TrainingResult:
    """
    Convenience function to train a model.
    
    Args:
        config: Training configuration
        feature_dataset: Feature dataset
        
    Returns:
        TrainingResult
    """
    trainer = ModelTrainer(config)
    return trainer.train(feature_dataset)
