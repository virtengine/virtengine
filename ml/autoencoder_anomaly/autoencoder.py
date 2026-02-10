"""
Autoencoder model architecture for anomaly detection.

VE-924: Autoencoder anomaly detection - encoder/decoder architecture

This module provides the autoencoder neural network architecture for
anomaly detection in identity verification. It implements:
- Convolutional encoder for dimensionality reduction
- Bottleneck latent representation
- Convolutional decoder for reconstruction
- Feature extraction at multiple scales
"""

import logging
import hashlib
import time
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any

import numpy as np

from ml.autoencoder_anomaly.config import (
    AutoencoderAnomalyConfig,
    EncoderConfig,
    DecoderConfig,
)

logger = logging.getLogger(__name__)


@dataclass
class EncoderOutput:
    """Output from encoder network."""
    
    # Latent representation
    latent_vector: np.ndarray
    
    # Intermediate features (for multi-scale analysis)
    layer_features: Dict[str, np.ndarray] = field(default_factory=dict)
    
    # Statistics
    latent_mean: float = 0.0
    latent_std: float = 0.0
    latent_min: float = 0.0
    latent_max: float = 0.0
    
    # Hash for verification
    feature_hash: str = ""
    
    def to_dict(self) -> dict:
        """Convert to dictionary (without raw arrays)."""
        return {
            "latent_dim": len(self.latent_vector) if self.latent_vector is not None else 0,
            "latent_mean": float(self.latent_mean),
            "latent_std": float(self.latent_std),
            "latent_min": float(self.latent_min),
            "latent_max": float(self.latent_max),
            "feature_hash": self.feature_hash,
            "num_layers": len(self.layer_features),
        }


@dataclass
class DecoderOutput:
    """Output from decoder network."""
    
    # Reconstructed image
    reconstruction: np.ndarray
    
    # Processing time
    processing_time_ms: float = 0.0
    
    def to_dict(self) -> dict:
        """Convert to dictionary (without raw arrays)."""
        return {
            "reconstruction_shape": self.reconstruction.shape if self.reconstruction is not None else None,
            "processing_time_ms": self.processing_time_ms,
        }


@dataclass
class AutoencoderOutput:
    """Complete output from autoencoder forward pass."""
    
    # Encoder output
    encoder_output: EncoderOutput
    
    # Decoder output
    decoder_output: DecoderOutput
    
    # Model info
    model_version: str = "1.0.0"
    model_hash: str = ""
    
    # Timing
    total_time_ms: float = 0.0
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "encoder": self.encoder_output.to_dict(),
            "decoder": self.decoder_output.to_dict(),
            "model_version": self.model_version,
            "model_hash": self.model_hash,
            "total_time_ms": self.total_time_ms,
        }


class ConvolutionalEncoder:
    """
    Convolutional encoder for dimensionality reduction.
    
    Architecture:
    - Multiple convolutional blocks with increasing filters
    - Each block: Conv2D -> BatchNorm -> LeakyReLU -> (optional Dropout)
    - Stride 2 for downsampling
    - Flatten and FC layer to latent vector
    
    Usage:
        config = EncoderConfig()
        encoder = ConvolutionalEncoder(config)
        output = encoder.encode(image)
    """
    
    MODEL_VERSION = "1.0.0"
    
    def __init__(self, config: Optional[EncoderConfig] = None):
        """
        Initialize the encoder.
        
        Args:
            config: Encoder configuration.
        """
        self.config = config or EncoderConfig()
        self._model_hash = self._compute_model_hash()
        
        # Initialize weights deterministically
        self._initialize_weights()
    
    def _compute_model_hash(self) -> str:
        """Compute deterministic hash of model configuration."""
        config_str = (
            f"input_size:{self.config.input_size},"
            f"input_channels:{self.config.input_channels},"
            f"layer_filters:{self.config.layer_filters},"
            f"latent_dim:{self.config.latent_dim},"
            f"version:{self.MODEL_VERSION}"
        )
        return hashlib.sha256(config_str.encode()).hexdigest()[:32]
    
    @property
    def model_hash(self) -> str:
        """Get model hash."""
        return self._model_hash
    
    def _initialize_weights(self) -> None:
        """Initialize network weights deterministically."""
        np.random.seed(42)
        
        self._weights = {}
        
        in_channels = self.config.input_channels
        
        # Convolutional layer weights
        for i, filters in enumerate(self.config.layer_filters):
            # Conv weights (He initialization)
            fan_in = in_channels * self.config.kernel_size * self.config.kernel_size
            std = np.sqrt(2.0 / fan_in)
            self._weights[f"conv{i+1}"] = np.random.randn(
                filters, in_channels,
                self.config.kernel_size, self.config.kernel_size
            ).astype(np.float32) * std
            
            # Batch norm parameters
            if self.config.use_batch_norm:
                self._weights[f"bn{i+1}_gamma"] = np.ones(filters, dtype=np.float32)
                self._weights[f"bn{i+1}_beta"] = np.zeros(filters, dtype=np.float32)
                self._weights[f"bn{i+1}_mean"] = np.zeros(filters, dtype=np.float32)
                self._weights[f"bn{i+1}_var"] = np.ones(filters, dtype=np.float32)
            
            in_channels = filters
        
        # Calculate flattened size after convolutions
        h, w = self.config.input_size
        for _ in self.config.layer_filters:
            h = (h + 1) // 2  # Stride 2
            w = (w + 1) // 2
        
        flattened_size = h * w * self.config.layer_filters[-1]
        
        # FC layer to latent
        std = np.sqrt(2.0 / flattened_size)
        self._weights["fc_latent"] = np.random.randn(
            self.config.latent_dim, flattened_size
        ).astype(np.float32) * std
        self._weights["fc_latent_bias"] = np.zeros(
            self.config.latent_dim, dtype=np.float32
        )
        
        self._spatial_size = (h, w)
        self._flattened_size = flattened_size
    
    def _conv2d(
        self,
        x: np.ndarray,
        weights: np.ndarray,
        stride: int = 2
    ) -> np.ndarray:
        """
        Simple 2D convolution (for inference).
        
        Note: In production, this would use optimized libraries.
        This is a simplified implementation for deterministic behavior.
        """
        out_channels, in_channels, kh, kw = weights.shape
        batch, _, h, w = x.shape
        
        # Calculate output size
        oh = (h + stride - 1) // stride
        ow = (w + stride - 1) // stride
        
        # Pad input
        pad_h = max(0, (oh - 1) * stride + kh - h)
        pad_w = max(0, (ow - 1) * stride + kw - w)
        
        if pad_h > 0 or pad_w > 0:
            x = np.pad(
                x,
                ((0, 0), (0, 0), (0, pad_h), (0, pad_w)),
                mode='constant'
            )
        
        # Simplified convolution
        output = np.zeros((batch, out_channels, oh, ow), dtype=np.float32)
        
        for b in range(batch):
            for oc in range(out_channels):
                for i in range(oh):
                    for j in range(ow):
                        ih = i * stride
                        jw = j * stride
                        patch = x[b, :, ih:ih+kh, jw:jw+kw]
                        output[b, oc, i, j] = np.sum(patch * weights[oc])
        
        return output
    
    def _batch_norm(
        self,
        x: np.ndarray,
        gamma: np.ndarray,
        beta: np.ndarray,
        mean: np.ndarray,
        var: np.ndarray,
        eps: float = 1e-5
    ) -> np.ndarray:
        """Apply batch normalization."""
        # Reshape for broadcasting
        gamma = gamma.reshape(1, -1, 1, 1)
        beta = beta.reshape(1, -1, 1, 1)
        mean = mean.reshape(1, -1, 1, 1)
        var = var.reshape(1, -1, 1, 1)
        
        return gamma * (x - mean) / np.sqrt(var + eps) + beta
    
    def _leaky_relu(self, x: np.ndarray, alpha: float = 0.2) -> np.ndarray:
        """Apply leaky ReLU activation."""
        return np.where(x > 0, x, alpha * x)
    
    def _preprocess(self, image: np.ndarray) -> np.ndarray:
        """
        Preprocess image for encoder input.
        
        Args:
            image: Input image (BGR format, uint8).
        
        Returns:
            Preprocessed image tensor (NCHW format, float32).
        """
        # Convert BGR to RGB
        if len(image.shape) == 3:
            image = image[:, :, ::-1]
        
        # Resize to input size
        h, w = image.shape[:2]
        target_h, target_w = self.config.input_size
        
        if h != target_h or w != target_w:
            # Simple resize using nearest neighbor for speed
            y_ratio = h / target_h
            x_ratio = w / target_w
            
            y_indices = (np.arange(target_h) * y_ratio).astype(int)
            x_indices = (np.arange(target_w) * x_ratio).astype(int)
            
            y_indices = np.clip(y_indices, 0, h - 1)
            x_indices = np.clip(x_indices, 0, w - 1)
            
            image = image[y_indices][:, x_indices]
        
        # Normalize to [0, 1]
        image = image.astype(np.float32) / 255.0
        
        # Convert to NCHW format
        if len(image.shape) == 3:
            image = np.transpose(image, (2, 0, 1))  # HWC -> CHW
            image = np.expand_dims(image, 0)  # Add batch dimension
        
        return image
    
    def encode(
        self,
        image: np.ndarray,
        return_features: bool = False
    ) -> EncoderOutput:
        """
        Encode image to latent representation.
        
        Args:
            image: Input image (BGR format, uint8).
            return_features: Whether to return intermediate features.
        
        Returns:
            EncoderOutput with latent vector and optional features.
        """
        start_time = time.time()
        
        # Preprocess
        x = self._preprocess(image)
        
        layer_features = {}
        
        # Forward through convolutional layers
        for i, filters in enumerate(self.config.layer_filters):
            layer_name = f"layer{i+1}"
            
            # Convolution
            x = self._conv2d(x, self._weights[f"conv{i+1}"])
            
            # Batch normalization
            if self.config.use_batch_norm:
                x = self._batch_norm(
                    x,
                    self._weights[f"bn{i+1}_gamma"],
                    self._weights[f"bn{i+1}_beta"],
                    self._weights[f"bn{i+1}_mean"],
                    self._weights[f"bn{i+1}_var"],
                )
            
            # Activation
            x = self._leaky_relu(x, self.config.leaky_relu_alpha)
            
            # Store intermediate features
            if return_features:
                layer_features[layer_name] = x.copy()
        
        # Flatten
        x_flat = x.reshape(x.shape[0], -1)
        
        # FC to latent
        latent = np.dot(x_flat, self._weights["fc_latent"].T)
        latent = latent + self._weights["fc_latent_bias"]
        
        # Squeeze batch dimension
        latent = latent.squeeze(0)
        
        # Compute statistics
        latent_mean = float(np.mean(latent))
        latent_std = float(np.std(latent))
        latent_min = float(np.min(latent))
        latent_max = float(np.max(latent))
        
        # Compute feature hash
        feature_hash = hashlib.sha256(latent.tobytes()).hexdigest()[:16]
        
        processing_time = (time.time() - start_time) * 1000
        
        return EncoderOutput(
            latent_vector=latent,
            layer_features=layer_features if return_features else {},
            latent_mean=latent_mean,
            latent_std=latent_std,
            latent_min=latent_min,
            latent_max=latent_max,
            feature_hash=feature_hash,
        )


class ConvolutionalDecoder:
    """
    Convolutional decoder for image reconstruction.
    
    Architecture:
    - FC layer from latent to spatial
    - Multiple transposed convolutional blocks with decreasing filters
    - Each block: ConvTranspose2D -> BatchNorm -> LeakyReLU
    - Final conv to output channels with sigmoid
    
    Usage:
        config = DecoderConfig()
        decoder = ConvolutionalDecoder(config, spatial_size=(8, 8))
        output = decoder.decode(latent_vector)
    """
    
    MODEL_VERSION = "1.0.0"
    
    def __init__(
        self,
        config: Optional[DecoderConfig] = None,
        latent_dim: int = 128,
        spatial_size: Tuple[int, int] = (8, 8)
    ):
        """
        Initialize the decoder.
        
        Args:
            config: Decoder configuration.
            latent_dim: Dimension of latent vector.
            spatial_size: Spatial size after encoder.
        """
        self.config = config or DecoderConfig()
        self.latent_dim = latent_dim
        self.spatial_size = spatial_size
        self._model_hash = self._compute_model_hash()
        
        # Initialize weights deterministically
        self._initialize_weights()
    
    def _compute_model_hash(self) -> str:
        """Compute deterministic hash of model configuration."""
        config_str = (
            f"layer_filters:{self.config.layer_filters},"
            f"latent_dim:{self.latent_dim},"
            f"spatial_size:{self.spatial_size},"
            f"output_channels:{self.config.output_channels},"
            f"version:{self.MODEL_VERSION}"
        )
        return hashlib.sha256(config_str.encode()).hexdigest()[:32]
    
    @property
    def model_hash(self) -> str:
        """Get model hash."""
        return self._model_hash
    
    def _initialize_weights(self) -> None:
        """Initialize network weights deterministically."""
        np.random.seed(43)  # Different seed from encoder
        
        self._weights = {}
        
        # FC from latent to spatial
        first_filters = self.config.layer_filters[0]
        spatial_size = self.spatial_size[0] * self.spatial_size[1] * first_filters
        
        std = np.sqrt(2.0 / self.latent_dim)
        self._weights["fc_spatial"] = np.random.randn(
            spatial_size, self.latent_dim
        ).astype(np.float32) * std
        self._weights["fc_spatial_bias"] = np.zeros(
            spatial_size, dtype=np.float32
        )
        
        # Transposed convolutional layer weights
        in_channels = first_filters
        
        for i, filters in enumerate(self.config.layer_filters[1:]):
            fan_in = in_channels * self.config.kernel_size * self.config.kernel_size
            std = np.sqrt(2.0 / fan_in)
            # Shape: (out_channels, in_channels, kh, kw)
            self._weights[f"deconv{i+1}"] = np.random.randn(
                filters, in_channels,
                self.config.kernel_size, self.config.kernel_size
            ).astype(np.float32) * std
            
            if self.config.use_batch_norm:
                self._weights[f"bn{i+1}_gamma"] = np.ones(filters, dtype=np.float32)
                self._weights[f"bn{i+1}_beta"] = np.zeros(filters, dtype=np.float32)
                self._weights[f"bn{i+1}_mean"] = np.zeros(filters, dtype=np.float32)
                self._weights[f"bn{i+1}_var"] = np.ones(filters, dtype=np.float32)
            
            in_channels = filters
        
        # Final conv to output channels
        last_filters = self.config.layer_filters[-1]
        fan_in = last_filters * self.config.kernel_size * self.config.kernel_size
        std = np.sqrt(2.0 / fan_in)
        self._weights["conv_out"] = np.random.randn(
            self.config.output_channels, last_filters,
            self.config.kernel_size, self.config.kernel_size
        ).astype(np.float32) * std
    
    def _upsample(self, x: np.ndarray, scale: int = 2) -> np.ndarray:
        """Upsample using nearest neighbor."""
        batch, channels, h, w = x.shape
        new_h, new_w = h * scale, w * scale
        
        # Repeat along height and width
        x = np.repeat(x, scale, axis=2)
        x = np.repeat(x, scale, axis=3)
        
        return x
    
    def _conv2d_same(
        self,
        x: np.ndarray,
        weights: np.ndarray
    ) -> np.ndarray:
        """2D convolution with same padding."""
        out_channels, in_channels, kh, kw = weights.shape
        batch, _, h, w = x.shape
        
        # Pad for same output size
        pad_h = kh // 2
        pad_w = kw // 2
        
        x = np.pad(
            x,
            ((0, 0), (0, 0), (pad_h, pad_h), (pad_w, pad_w)),
            mode='constant'
        )
        
        output = np.zeros((batch, out_channels, h, w), dtype=np.float32)
        
        for b in range(batch):
            for oc in range(out_channels):
                for i in range(h):
                    for j in range(w):
                        patch = x[b, :, i:i+kh, j:j+kw]
                        output[b, oc, i, j] = np.sum(patch * weights[oc])
        
        return output
    
    def _batch_norm(
        self,
        x: np.ndarray,
        gamma: np.ndarray,
        beta: np.ndarray,
        mean: np.ndarray,
        var: np.ndarray,
        eps: float = 1e-5
    ) -> np.ndarray:
        """Apply batch normalization."""
        gamma = gamma.reshape(1, -1, 1, 1)
        beta = beta.reshape(1, -1, 1, 1)
        mean = mean.reshape(1, -1, 1, 1)
        var = var.reshape(1, -1, 1, 1)
        
        return gamma * (x - mean) / np.sqrt(var + eps) + beta
    
    def _leaky_relu(self, x: np.ndarray, alpha: float = 0.2) -> np.ndarray:
        """Apply leaky ReLU activation."""
        return np.where(x > 0, x, alpha * x)
    
    def _sigmoid(self, x: np.ndarray) -> np.ndarray:
        """Apply sigmoid activation."""
        return 1.0 / (1.0 + np.exp(-np.clip(x, -500, 500)))
    
    def decode(self, latent_vector: np.ndarray) -> DecoderOutput:
        """
        Decode latent vector to reconstructed image.
        
        Args:
            latent_vector: Latent representation from encoder.
        
        Returns:
            DecoderOutput with reconstructed image.
        """
        start_time = time.time()
        
        # Ensure batch dimension
        if len(latent_vector.shape) == 1:
            latent_vector = np.expand_dims(latent_vector, 0)
        
        # FC to spatial
        x = np.dot(latent_vector, self._weights["fc_spatial"].T)
        x = x + self._weights["fc_spatial_bias"]
        
        # Reshape to spatial
        first_filters = self.config.layer_filters[0]
        x = x.reshape(
            x.shape[0],
            first_filters,
            self.spatial_size[0],
            self.spatial_size[1]
        )
        
        # Forward through deconv layers
        for i, filters in enumerate(self.config.layer_filters[1:]):
            # Upsample
            x = self._upsample(x, scale=2)
            
            # Convolution (same padding)
            x = self._conv2d_same(x, self._weights[f"deconv{i+1}"])
            
            # Batch normalization
            if self.config.use_batch_norm:
                x = self._batch_norm(
                    x,
                    self._weights[f"bn{i+1}_gamma"],
                    self._weights[f"bn{i+1}_beta"],
                    self._weights[f"bn{i+1}_mean"],
                    self._weights[f"bn{i+1}_var"],
                )
            
            # Activation
            x = self._leaky_relu(x, self.config.leaky_relu_alpha)
        
        # Final upsample and conv
        x = self._upsample(x, scale=2)
        x = self._conv2d_same(x, self._weights["conv_out"])
        
        # Sigmoid activation for output
        x = self._sigmoid(x)
        
        # Squeeze batch and convert to HWC
        x = x.squeeze(0)  # Remove batch
        x = np.transpose(x, (1, 2, 0))  # CHW -> HWC
        
        processing_time = (time.time() - start_time) * 1000
        
        return DecoderOutput(
            reconstruction=x,
            processing_time_ms=processing_time,
        )


class Autoencoder:
    """
    Complete autoencoder for anomaly detection.
    
    Combines encoder and decoder for end-to-end reconstruction.
    
    Usage:
        config = AutoencoderAnomalyConfig()
        autoencoder = Autoencoder(config)
        output = autoencoder.forward(image)
    """
    
    MODEL_VERSION = "1.0.0"
    
    def __init__(self, config: Optional[AutoencoderAnomalyConfig] = None):
        """
        Initialize the autoencoder.
        
        Args:
            config: Autoencoder configuration.
        """
        self.config = config or AutoencoderAnomalyConfig()
        
        # Initialize encoder
        self.encoder = ConvolutionalEncoder(self.config.encoder)
        
        # Initialize decoder with matching dimensions
        self.decoder = ConvolutionalDecoder(
            config=self.config.decoder,
            latent_dim=self.config.encoder.latent_dim,
            spatial_size=self.encoder._spatial_size,
        )
        
        # Compute combined model hash
        self._model_hash = self._compute_model_hash()
    
    def _compute_model_hash(self) -> str:
        """Compute deterministic hash of complete model."""
        combined = f"{self.encoder.model_hash}:{self.decoder.model_hash}"
        return hashlib.sha256(combined.encode()).hexdigest()[:32]
    
    @property
    def model_hash(self) -> str:
        """Get model hash."""
        return self._model_hash
    
    def forward(
        self,
        image: np.ndarray,
        return_features: bool = False
    ) -> AutoencoderOutput:
        """
        Forward pass through autoencoder.
        
        Args:
            image: Input image (BGR format, uint8).
            return_features: Whether to return intermediate features.
        
        Returns:
            AutoencoderOutput with latent representation and reconstruction.
        """
        start_time = time.time()
        
        # Encode
        encoder_output = self.encoder.encode(image, return_features)
        
        # Decode
        decoder_output = self.decoder.decode(encoder_output.latent_vector)
        
        total_time = (time.time() - start_time) * 1000
        
        return AutoencoderOutput(
            encoder_output=encoder_output,
            decoder_output=decoder_output,
            model_version=self.MODEL_VERSION,
            model_hash=self._model_hash,
            total_time_ms=total_time,
        )
    
    def encode(
        self,
        image: np.ndarray,
        return_features: bool = False
    ) -> EncoderOutput:
        """
        Encode image to latent representation.
        
        Args:
            image: Input image.
            return_features: Whether to return intermediate features.
        
        Returns:
            EncoderOutput with latent vector.
        """
        return self.encoder.encode(image, return_features)
    
    def decode(self, latent_vector: np.ndarray) -> DecoderOutput:
        """
        Decode latent vector to image.
        
        Args:
            latent_vector: Latent representation.
        
        Returns:
            DecoderOutput with reconstructed image.
        """
        return self.decoder.decode(latent_vector)


def create_autoencoder(
    config: Optional[AutoencoderAnomalyConfig] = None
) -> Autoencoder:
    """
    Factory function to create an autoencoder.
    
    Args:
        config: Autoencoder configuration.
    
    Returns:
        Initialized Autoencoder instance.
    """
    return Autoencoder(config)
