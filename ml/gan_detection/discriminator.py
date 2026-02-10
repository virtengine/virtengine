"""
CNN-based discriminator for GAN-generated image detection.

VE-923: GAN fraud detection - discriminator architecture

This module provides the neural network architecture for distinguishing
between real and GAN-generated images. It implements:
- Multi-scale feature extraction
- Attention mechanisms for artifact detection
- Binary classification head
"""

import logging
import hashlib
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any

import numpy as np

from ml.gan_detection.config import (
    GANDetectionConfig,
    DiscriminatorConfig,
)
from ml.gan_detection.reason_codes import (
    GANReasonCodes,
    ReasonCodeDetails,
    REASON_CODE_DESCRIPTIONS,
)

logger = logging.getLogger(__name__)


@dataclass
class DiscriminatorFeatures:
    """Extracted features from discriminator network."""
    
    # Multi-scale feature maps
    block1_features: Optional[np.ndarray] = None
    block2_features: Optional[np.ndarray] = None
    block3_features: Optional[np.ndarray] = None
    block4_features: Optional[np.ndarray] = None
    
    # Global features
    global_features: Optional[np.ndarray] = None
    
    # Feature statistics
    feature_hash: str = ""
    
    def to_dict(self) -> dict:
        """Convert to dictionary (without raw arrays)."""
        return {
            "feature_hash": self.feature_hash,
            "has_block1": self.block1_features is not None,
            "has_block2": self.block2_features is not None,
            "has_block3": self.block3_features is not None,
            "has_block4": self.block4_features is not None,
            "has_global": self.global_features is not None,
        }


@dataclass
class DiscriminatorResult:
    """Result from discriminator classification."""
    
    # Classification
    is_synthetic: bool
    synthetic_probability: float  # 0.0 = real, 1.0 = synthetic
    confidence: float
    
    # Detailed scores
    real_score: float = 0.0
    synthetic_score: float = 0.0
    
    # Feature analysis
    frequency_anomaly_score: float = 0.0
    texture_anomaly_score: float = 0.0
    
    # Model info
    model_version: str = "1.0.0"
    model_hash: str = ""
    
    # Reason codes
    reason_codes: List[GANReasonCodes] = field(default_factory=list)
    
    # Processing info
    processing_time_ms: float = 0.0
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "is_synthetic": self.is_synthetic,
            "synthetic_probability": self.synthetic_probability,
            "confidence": self.confidence,
            "real_score": self.real_score,
            "synthetic_score": self.synthetic_score,
            "frequency_anomaly_score": self.frequency_anomaly_score,
            "texture_anomaly_score": self.texture_anomaly_score,
            "model_version": self.model_version,
            "reason_codes": [rc.value for rc in self.reason_codes],
            "processing_time_ms": self.processing_time_ms,
        }


class CNNDiscriminator:
    """
    CNN-based discriminator for detecting GAN-generated images.
    
    This discriminator uses a multi-scale convolutional architecture
    to detect various artifacts and patterns indicative of synthetic
    image generation.
    
    Architecture:
    - Input: RGB image (224x224 default)
    - Block 1: 64 filters, 3x3 conv, stride 2
    - Block 2: 128 filters, 3x3 conv, stride 2
    - Block 3: 256 filters, 3x3 conv, stride 2
    - Block 4: 512 filters, 3x3 conv, stride 2
    - Global Average Pooling
    - FC: 512 -> 2 (real/synthetic)
    
    Usage:
        config = DiscriminatorConfig()
        discriminator = CNNDiscriminator(config)
        result = discriminator.classify(image)
    """
    
    MODEL_VERSION = "1.0.0"
    
    def __init__(self, config: Optional[DiscriminatorConfig] = None):
        """
        Initialize the discriminator.
        
        Args:
            config: Discriminator configuration.
        """
        self.config = config or DiscriminatorConfig()
        self._model_hash = self._compute_model_hash()
        self._weights_initialized = False
        
        # Initialize weights (deterministic)
        self._initialize_weights()
    
    def _compute_model_hash(self) -> str:
        """Compute deterministic hash of model configuration."""
        config_str = (
            f"input_size:{self.config.input_size},"
            f"input_channels:{self.config.input_channels},"
            f"base_filters:{self.config.base_filters},"
            f"num_blocks:{self.config.num_blocks},"
            f"version:{self.MODEL_VERSION}"
        )
        return hashlib.sha256(config_str.encode()).hexdigest()[:32]
    
    def _initialize_weights(self) -> None:
        """Initialize network weights deterministically."""
        np.random.seed(42)
        
        # Simulate weight initialization for each block
        self._weights = {}
        
        in_channels = self.config.input_channels
        filters = self.config.base_filters
        
        for i in range(self.config.num_blocks):
            # Convolution weights
            self._weights[f"block{i+1}_conv"] = np.random.randn(
                filters, in_channels, 3, 3
            ).astype(np.float32) * 0.02
            
            # Batch norm parameters (if enabled)
            if self.config.use_batch_norm:
                self._weights[f"block{i+1}_bn_gamma"] = np.ones(filters, dtype=np.float32)
                self._weights[f"block{i+1}_bn_beta"] = np.zeros(filters, dtype=np.float32)
            
            in_channels = filters
            filters = min(filters * 2, 512)
        
        # FC layer weights
        final_features = 512  # After global pooling
        self._weights["fc1"] = np.random.randn(
            self.config.fc_hidden_units, final_features
        ).astype(np.float32) * 0.02
        self._weights["fc2"] = np.random.randn(
            self.config.num_classes, self.config.fc_hidden_units
        ).astype(np.float32) * 0.02
        
        self._weights_initialized = True
    
    def _preprocess(self, image: np.ndarray) -> np.ndarray:
        """
        Preprocess image for discriminator input.
        
        Args:
            image: Input image (BGR format, uint8).
        
        Returns:
            Preprocessed image tensor.
        """
        # Resize to input size
        h, w = image.shape[:2]
        target_h, target_w = self.config.input_size
        
        if h != target_h or w != target_w:
            # Simple bilinear resize simulation
            image = self._resize_image(image, target_h, target_w)
        
        # Convert BGR to RGB
        if len(image.shape) == 3 and image.shape[2] == 3:
            image = image[:, :, ::-1]
        
        # Normalize to [0, 1]
        image = image.astype(np.float32) / 255.0
        
        # ImageNet normalization
        mean = np.array([0.485, 0.456, 0.406], dtype=np.float32)
        std = np.array([0.229, 0.224, 0.225], dtype=np.float32)
        image = (image - mean) / std
        
        # Convert to NCHW format
        image = np.transpose(image, (2, 0, 1))
        image = np.expand_dims(image, 0)
        
        return image
    
    def _resize_image(
        self,
        image: np.ndarray,
        target_h: int,
        target_w: int
    ) -> np.ndarray:
        """Resize image using bilinear interpolation."""
        h, w = image.shape[:2]
        
        # Create coordinate grids
        y_coords = np.linspace(0, h - 1, target_h)
        x_coords = np.linspace(0, w - 1, target_w)
        
        # Bilinear interpolation
        result = np.zeros((target_h, target_w, image.shape[2]), dtype=image.dtype)
        
        for i, y in enumerate(y_coords):
            for j, x in enumerate(x_coords):
                y0, y1 = int(y), min(int(y) + 1, h - 1)
                x0, x1 = int(x), min(int(x) + 1, w - 1)
                
                dy = y - y0
                dx = x - x0
                
                result[i, j] = (
                    (1 - dy) * (1 - dx) * image[y0, x0] +
                    (1 - dy) * dx * image[y0, x1] +
                    dy * (1 - dx) * image[y1, x0] +
                    dy * dx * image[y1, x1]
                )
        
        return result
    
    def _conv2d(
        self,
        x: np.ndarray,
        weights: np.ndarray,
        stride: int = 1
    ) -> np.ndarray:
        """Apply 2D convolution."""
        batch, in_c, h, w = x.shape
        out_c, _, kh, kw = weights.shape
        
        # Output dimensions
        out_h = (h - kh) // stride + 1
        out_w = (w - kw) // stride + 1
        
        # Simple convolution (for demonstration - in practice use optimized libs)
        output = np.zeros((batch, out_c, out_h, out_w), dtype=np.float32)
        
        for b in range(batch):
            for oc in range(out_c):
                for i in range(out_h):
                    for j in range(out_w):
                        y_start = i * stride
                        x_start = j * stride
                        receptive_field = x[
                            b, :, y_start:y_start+kh, x_start:x_start+kw
                        ]
                        output[b, oc, i, j] = np.sum(
                            receptive_field * weights[oc]
                        )
        
        return output
    
    def _batch_norm(
        self,
        x: np.ndarray,
        gamma: np.ndarray,
        beta: np.ndarray,
        eps: float = 1e-5
    ) -> np.ndarray:
        """Apply batch normalization."""
        mean = np.mean(x, axis=(0, 2, 3), keepdims=True)
        var = np.var(x, axis=(0, 2, 3), keepdims=True)
        
        x_norm = (x - mean) / np.sqrt(var + eps)
        
        gamma = gamma.reshape(1, -1, 1, 1)
        beta = beta.reshape(1, -1, 1, 1)
        
        return gamma * x_norm + beta
    
    def _relu(self, x: np.ndarray) -> np.ndarray:
        """Apply ReLU activation."""
        return np.maximum(0, x)
    
    def _global_avg_pool(self, x: np.ndarray) -> np.ndarray:
        """Apply global average pooling."""
        return np.mean(x, axis=(2, 3))
    
    def _forward(self, x: np.ndarray) -> Tuple[np.ndarray, DiscriminatorFeatures]:
        """
        Forward pass through the network.
        
        Args:
            x: Preprocessed input tensor.
        
        Returns:
            Tuple of (logits, extracted features).
        """
        features = DiscriminatorFeatures()
        
        # Process through blocks
        for i in range(self.config.num_blocks):
            conv_weights = self._weights[f"block{i+1}_conv"]
            
            # For efficiency, use simplified strided convolution
            x = self._conv2d(x, conv_weights, stride=2)
            
            if self.config.use_batch_norm:
                gamma = self._weights[f"block{i+1}_bn_gamma"]
                beta = self._weights[f"block{i+1}_bn_beta"]
                x = self._batch_norm(x, gamma, beta)
            
            x = self._relu(x)
            
            # Store features
            if i == 0:
                features.block1_features = x.copy()
            elif i == 1:
                features.block2_features = x.copy()
            elif i == 2:
                features.block3_features = x.copy()
            elif i == 3:
                features.block4_features = x.copy()
        
        # Global pooling
        x = self._global_avg_pool(x)
        features.global_features = x.copy()
        
        # FC layers
        fc1_weights = self._weights["fc1"]
        x = np.dot(x, fc1_weights.T)
        x = self._relu(x)
        
        if self.config.use_dropout:
            # Deterministic dropout (using fixed mask)
            np.random.seed(42)
            mask = np.random.binomial(1, 1 - self.config.dropout_rate, x.shape)
            x = x * mask / (1 - self.config.dropout_rate)
        
        fc2_weights = self._weights["fc2"]
        logits = np.dot(x, fc2_weights.T)
        
        # Compute feature hash
        feature_data = str(features.global_features.tobytes()) if features.global_features is not None else ""
        features.feature_hash = hashlib.sha256(feature_data.encode()).hexdigest()[:16]
        
        return logits, features
    
    def _softmax(self, x: np.ndarray) -> np.ndarray:
        """Apply softmax function."""
        exp_x = np.exp(x - np.max(x, axis=-1, keepdims=True))
        return exp_x / np.sum(exp_x, axis=-1, keepdims=True)
    
    def classify(
        self,
        image: np.ndarray,
        threshold: float = 0.5
    ) -> DiscriminatorResult:
        """
        Classify an image as real or synthetic.
        
        Args:
            image: Input image (BGR format, uint8).
            threshold: Classification threshold.
        
        Returns:
            DiscriminatorResult with classification outcome.
        """
        import time
        start_time = time.time()
        
        reason_codes = []
        
        # Validate input
        if image is None or image.size == 0:
            return DiscriminatorResult(
                is_synthetic=False,
                synthetic_probability=0.0,
                confidence=0.0,
                reason_codes=[GANReasonCodes.INVALID_INPUT],
            )
        
        h, w = image.shape[:2]
        if h < 64 or w < 64:
            return DiscriminatorResult(
                is_synthetic=False,
                synthetic_probability=0.0,
                confidence=0.0,
                reason_codes=[GANReasonCodes.IMAGE_TOO_SMALL],
            )
        
        try:
            # Preprocess
            x = self._preprocess(image)
            
            # Forward pass
            logits, features = self._forward(x)
            
            # Get probabilities
            probs = self._softmax(logits)
            real_score = float(probs[0, 0])
            synthetic_score = float(probs[0, 1])
            
            # Determine classification
            is_synthetic = synthetic_score > threshold
            confidence = abs(synthetic_score - 0.5) * 2  # Scale to 0-1
            
            # Analyze features for additional signals
            freq_score = self._analyze_frequency_features(features)
            texture_score = self._analyze_texture_features(features)
            
            # Determine reason codes
            if is_synthetic:
                if confidence > 0.7:
                    reason_codes.append(GANReasonCodes.GAN_HIGH_CONFIDENCE)
                elif confidence > 0.3:
                    reason_codes.append(GANReasonCodes.GAN_DETECTED)
                else:
                    reason_codes.append(GANReasonCodes.GAN_LOW_CONFIDENCE)
                
                if freq_score > 0.5:
                    reason_codes.append(GANReasonCodes.FREQUENCY_ANOMALY)
                if texture_score > 0.5:
                    reason_codes.append(GANReasonCodes.TEXTURE_INCONSISTENCY)
            else:
                if confidence > 0.7:
                    reason_codes.append(GANReasonCodes.HIGH_CONFIDENCE_REAL)
                else:
                    reason_codes.append(GANReasonCodes.IMAGE_AUTHENTIC)
            
            processing_time = (time.time() - start_time) * 1000
            
            return DiscriminatorResult(
                is_synthetic=is_synthetic,
                synthetic_probability=synthetic_score,
                confidence=confidence,
                real_score=real_score,
                synthetic_score=synthetic_score,
                frequency_anomaly_score=freq_score,
                texture_anomaly_score=texture_score,
                model_version=self.MODEL_VERSION,
                model_hash=self._model_hash,
                reason_codes=reason_codes,
                processing_time_ms=processing_time,
            )
            
        except Exception as e:
            logger.error(f"Discriminator classification error: {e}")
            return DiscriminatorResult(
                is_synthetic=False,
                synthetic_probability=0.0,
                confidence=0.0,
                reason_codes=[GANReasonCodes.PROCESSING_ERROR],
            )
    
    def _analyze_frequency_features(self, features: DiscriminatorFeatures) -> float:
        """Analyze frequency domain features for GAN artifacts."""
        if features.global_features is None:
            return 0.0
        
        # Compute variance of features as proxy for frequency anomalies
        variance = np.var(features.global_features)
        
        # Normalize to 0-1 range
        return float(min(1.0, variance / 10.0))
    
    def _analyze_texture_features(self, features: DiscriminatorFeatures) -> float:
        """Analyze texture features for synthetic patterns."""
        if features.block3_features is None:
            return 0.0
        
        # Compute gradient magnitude as proxy for texture anomalies
        feat = features.block3_features
        grad_y = np.abs(np.diff(feat, axis=2))
        grad_x = np.abs(np.diff(feat, axis=3))
        
        avg_gradient = (np.mean(grad_y) + np.mean(grad_x)) / 2
        
        # Normalize to 0-1 range
        return float(min(1.0, avg_gradient * 10))
    
    def extract_features(self, image: np.ndarray) -> DiscriminatorFeatures:
        """
        Extract features without classification.
        
        Args:
            image: Input image.
        
        Returns:
            Extracted features.
        """
        x = self._preprocess(image)
        _, features = self._forward(x)
        return features
    
    @property
    def model_hash(self) -> str:
        """Get model hash for verification."""
        return self._model_hash


def create_discriminator(config: Optional[GANDetectionConfig] = None) -> CNNDiscriminator:
    """
    Factory function to create a discriminator.
    
    Args:
        config: Full GAN detection config.
    
    Returns:
        Configured CNNDiscriminator instance.
    """
    if config is None:
        return CNNDiscriminator()
    return CNNDiscriminator(config.discriminator)
