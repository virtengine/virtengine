"""
Trust score model architecture.

Defines the neural network architecture for predicting trust scores
from combined feature vectors.
"""

import logging
from typing import List, Optional, Tuple

from ml.training.config import ModelConfig

logger = logging.getLogger(__name__)

# Attempt to import TensorFlow
try:
    import tensorflow as tf
    from tensorflow import keras
    from tensorflow.keras import layers, regularizers
    TF_AVAILABLE = True
except ImportError:
    TF_AVAILABLE = False
    logger.warning("TensorFlow not available, model functionality will be limited")


def create_trust_score_model(config: ModelConfig) -> "tf.keras.Model":
    """
    Create a trust score model with the specified configuration.
    
    Args:
        config: Model configuration
        
    Returns:
        Compiled Keras model
    """
    if not TF_AVAILABLE:
        raise ImportError("TensorFlow is required for model creation")
    
    # Input layer
    inputs = keras.Input(shape=(config.input_dim,), name="features")
    
    x = inputs
    
    # Hidden layers
    for i, units in enumerate(config.hidden_layers):
        # Dense layer with L2 regularization
        x = layers.Dense(
            units,
            kernel_regularizer=regularizers.l2(config.l2_regularization),
            name=f"dense_{i}"
        )(x)
        
        # Batch normalization
        if config.use_batch_norm:
            x = layers.BatchNormalization(name=f"bn_{i}")(x)
        
        # Activation
        if config.activation == "relu":
            x = layers.ReLU(name=f"relu_{i}")(x)
        elif config.activation == "leaky_relu":
            x = layers.LeakyReLU(alpha=0.1, name=f"leaky_relu_{i}")(x)
        elif config.activation == "elu":
            x = layers.ELU(name=f"elu_{i}")(x)
        else:
            x = layers.Activation(config.activation, name=f"activation_{i}")(x)
        
        # Dropout
        if config.dropout_rate > 0:
            x = layers.Dropout(config.dropout_rate, name=f"dropout_{i}")(x)
    
    # Output layer
    if config.output_activation == "sigmoid":
        # Sigmoid outputs 0-1, scale to 0-100
        raw_output = layers.Dense(1, activation="sigmoid", name="raw_output")(x)
        outputs = layers.Lambda(
            lambda x: x * config.output_scale,
            name="trust_score"
        )(raw_output)
    else:
        # Direct output (for regression)
        outputs = layers.Dense(
            1,
            activation=config.output_activation,
            name="trust_score"
        )(x)
    
    model = keras.Model(inputs=inputs, outputs=outputs, name="trust_score_model")
    
    return model


class TrustScoreModel:
    """
    Trust score prediction model.
    
    Wraps a Keras model with training, evaluation, and prediction methods.
    
    Architecture:
    - Input: Combined feature vector (face embeddings, doc features, OCR, metadata)
    - Hidden layers: MLP with batch normalization and dropout
    - Output: Trust score 0-100
    """
    
    def __init__(self, config: Optional[ModelConfig] = None):
        """
        Initialize the trust score model.
        
        Args:
            config: Model configuration
        """
        if not TF_AVAILABLE:
            raise ImportError("TensorFlow is required for TrustScoreModel")
        
        self.config = config or ModelConfig()
        self.model = self._build_model(self.config)
        self._is_compiled = False
        self._history = None
    
    def _build_model(self, config: ModelConfig) -> "tf.keras.Model":
        """
        Build the Keras model.
        
        Args:
            config: Model configuration
            
        Returns:
            Uncompiled Keras model
        """
        return create_trust_score_model(config)
    
    def compile(self, learning_rate: Optional[float] = None) -> None:
        """
        Compile the model with optimizer and loss.
        
        Args:
            learning_rate: Learning rate (uses config if not provided)
        """
        lr = learning_rate or self.config.learning_rate
        
        # Create optimizer
        if self.config.optimizer == "adam":
            optimizer = keras.optimizers.Adam(
                learning_rate=lr,
                beta_1=self.config.adam_beta_1,
                beta_2=self.config.adam_beta_2,
                epsilon=self.config.adam_epsilon,
            )
        elif self.config.optimizer == "sgd":
            optimizer = keras.optimizers.SGD(learning_rate=lr, momentum=0.9)
        elif self.config.optimizer == "rmsprop":
            optimizer = keras.optimizers.RMSprop(learning_rate=lr)
        else:
            optimizer = keras.optimizers.Adam(learning_rate=lr)
        
        # Compile with MSE loss for regression
        self.model.compile(
            optimizer=optimizer,
            loss="mse",
            metrics=["mae"],
        )
        
        self._is_compiled = True
        logger.info(f"Model compiled with {self.config.optimizer} optimizer, lr={lr}")
    
    def train(
        self,
        train_features,
        train_labels,
        val_features=None,
        val_labels=None,
        epochs: Optional[int] = None,
        batch_size: Optional[int] = None,
        verbose: int = 1,
    ) -> dict:
        """
        Train the model.
        
        Args:
            train_features: Training feature array
            train_labels: Training labels (trust scores)
            val_features: Validation feature array
            val_labels: Validation labels
            epochs: Number of epochs (uses config if not provided)
            batch_size: Batch size (uses config if not provided)
            verbose: Verbosity level (0, 1, or 2)
            
        Returns:
            Training history dictionary
        """
        if not self._is_compiled:
            self.compile()
        
        epochs = epochs or self.config.epochs
        batch_size = batch_size or self.config.batch_size
        
        # Build callbacks
        callbacks = self._build_callbacks()
        
        # Prepare validation data
        validation_data = None
        if val_features is not None and val_labels is not None:
            validation_data = (val_features, val_labels)
        
        # Train
        logger.info(f"Starting training for {epochs} epochs")
        self._history = self.model.fit(
            train_features,
            train_labels,
            epochs=epochs,
            batch_size=batch_size,
            validation_data=validation_data,
            callbacks=callbacks,
            verbose=verbose,
        )
        
        return self._history.history
    
    def _build_callbacks(self) -> list:
        """Build training callbacks."""
        callbacks = []
        
        # Early stopping
        if self.config.early_stopping:
            callbacks.append(
                keras.callbacks.EarlyStopping(
                    monitor="val_loss" if True else "loss",
                    patience=self.config.early_stopping_patience,
                    min_delta=self.config.early_stopping_min_delta,
                    restore_best_weights=True,
                    verbose=1,
                )
            )
        
        # Learning rate scheduler
        if self.config.use_lr_schedule:
            callbacks.append(
                keras.callbacks.ReduceLROnPlateau(
                    monitor="val_loss" if True else "loss",
                    factor=self.config.lr_decay_factor,
                    patience=self.config.lr_decay_patience,
                    min_lr=self.config.min_lr,
                    verbose=1,
                )
            )
        
        # Model checkpointing
        if self.config.checkpoint_dir:
            import os
            os.makedirs(self.config.checkpoint_dir, exist_ok=True)
            callbacks.append(
                keras.callbacks.ModelCheckpoint(
                    filepath=os.path.join(
                        self.config.checkpoint_dir,
                        "model_{epoch:03d}_{val_loss:.4f}.keras"
                    ),
                    monitor="val_loss",
                    saVIRTENGINE_best_only=self.config.saVIRTENGINE_best_only,
                    verbose=1,
                )
            )
        
        # TensorBoard logging
        if self.config.tensorboard_log_dir:
            import os
            os.makedirs(self.config.tensorboard_log_dir, exist_ok=True)
            callbacks.append(
                keras.callbacks.TensorBoard(
                    log_dir=self.config.tensorboard_log_dir,
                    histogram_freq=1,
                    update_freq=self.config.log_every_n_steps,
                )
            )
        
        return callbacks
    
    def predict(self, features) -> "tf.Tensor":
        """
        Predict trust scores for features.
        
        Args:
            features: Feature array
            
        Returns:
            Predicted trust scores
        """
        return self.model.predict(features, verbose=0)
    
    def evaluate(self, features, labels) -> dict:
        """
        Evaluate the model on test data.
        
        Args:
            features: Test feature array
            labels: True labels
            
        Returns:
            Dictionary with loss and metrics
        """
        results = self.model.evaluate(features, labels, verbose=0)
        return {
            "loss": results[0],
            "mae": results[1],
        }
    
    def summary(self) -> str:
        """Get model summary as string."""
        import io
        stream = io.StringIO()
        self.model.summary(print_fn=lambda x: stream.write(x + "\n"))
        return stream.getvalue()
    
    def get_weights(self) -> list:
        """Get model weights."""
        return self.model.get_weights()
    
    def set_weights(self, weights: list) -> None:
        """Set model weights."""
        self.model.set_weights(weights)
    
    def save(self, filepath: str) -> None:
        """Save the model."""
        self.model.save(filepath)
        logger.info(f"Model saved to {filepath}")
    
    @classmethod
    def load(cls, filepath: str, config: Optional[ModelConfig] = None) -> "TrustScoreModel":
        """Load a saved model."""
        if not TF_AVAILABLE:
            raise ImportError("TensorFlow is required for loading models")
        
        instance = cls.__new__(cls)
        instance.config = config or ModelConfig()
        instance.model = keras.models.load_model(filepath)
        instance._is_compiled = True
        instance._history = None
        
        logger.info(f"Model loaded from {filepath}")
        return instance
    
    @property
    def history(self) -> Optional[dict]:
        """Get training history."""
        if self._history is not None:
            return self._history.history
        return None
    
    @property
    def keras_model(self) -> "tf.keras.Model":
        """Get the underlying Keras model."""
        return self.model
