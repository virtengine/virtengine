"""
U-Net face segmentation model.

This module provides the U-Net model wrapper for face region segmentation
from identity document images.
"""

import logging
import hashlib
import os
from dataclasses import dataclass
from typing import Optional, Tuple, Any
from pathlib import Path

import numpy as np

from ml.face_extraction.config import UNetConfig

logger = logging.getLogger(__name__)


# Set deterministic environment variables
os.environ["TF_DETERMINISTIC_OPS"] = "1"
os.environ["PYTHONHASHSEED"] = "42"


@dataclass
class SegmentationResult:
    """Result of face segmentation."""
    
    mask: np.ndarray
    confidence_map: np.ndarray
    success: bool
    mean_confidence: float
    max_confidence: float
    model_version: str
    model_hash: str
    error_message: Optional[str] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary for serialization."""
        return {
            "success": self.success,
            "mean_confidence": float(self.mean_confidence),
            "max_confidence": float(self.max_confidence),
            "model_version": self.model_version,
            "model_hash": self.model_hash[:16] + "..." if self.model_hash else None,
            "error_message": self.error_message,
        }


class UNetFaceSegmentor:
    """
    U-Net model for face region segmentation.
    
    This class provides face segmentation from identity documents using
    a U-Net architecture with support for deterministic execution.
    """
    
    def __init__(self, config: Optional[UNetConfig] = None):
        """
        Initialize the U-Net segmentor.
        
        Args:
            config: U-Net configuration. Uses defaults if not provided.
        """
        self.config = config or UNetConfig()
        self._model = None
        self._model_hash: Optional[str] = None
        self._initialized = False
        self.model_version = self.config.model_version
    
    def _ensure_determinism(self) -> None:
        """Ensure deterministic execution for consensus."""
        import random
        random.seed(self.config.random_seed)
        np.random.seed(self.config.random_seed)
        
        try:
            import tensorflow as tf
            tf.random.set_seed(self.config.random_seed)
            
            # Disable GPU if not wanted
            if not self.config.use_gpu:
                os.environ["CUDA_VISIBLE_DEVICES"] = "-1"
                tf.config.set_visible_devices([], 'GPU')
        except ImportError:
            pass
    
    def _lazy_init(self) -> None:
        """Lazy initialization of the model."""
        if self._initialized:
            return
        
        if self.config.ensure_determinism:
            self._ensure_determinism()
        
        try:
            self._model = self._build_or_load_model()
            self._model_hash = self._compute_weights_hash()
            self._initialized = True
            logger.info(f"U-Net model initialized. Version: {self.model_version}, "
                       f"Hash: {self._model_hash[:16]}...")
        except Exception as e:
            logger.error(f"Failed to initialize U-Net model: {e}")
            raise RuntimeError(f"U-Net initialization failed: {e}")
    
    def _build_or_load_model(self) -> Any:
        """Build or load the U-Net model."""
        try:
            import tensorflow as tf
            from tensorflow import keras
            from tensorflow.keras import layers
        except ImportError:
            raise RuntimeError("TensorFlow not available. Install with: pip install tensorflow")
        
        # Try to load pre-trained model if path provided
        if self.config.model_path and Path(self.config.model_path).exists():
            logger.info(f"Loading U-Net model from {self.config.model_path}")
            return keras.models.load_model(self.config.model_path)
        
        # Build U-Net architecture
        logger.info("Building U-Net model architecture")
        return self._build_unet(
            input_shape=(*self.config.input_size, self.config.input_channels),
            depth=self.config.encoder_depth,
            initial_filters=self.config.initial_filters,
            use_batch_norm=self.config.use_batch_norm,
        )
    
    def _build_unet(
        self,
        input_shape: Tuple[int, int, int],
        depth: int = 4,
        initial_filters: int = 64,
        use_batch_norm: bool = True,
    ) -> Any:
        """
        Build U-Net architecture for face segmentation.
        
        Args:
            input_shape: Input image shape (H, W, C)
            depth: Number of encoder/decoder levels
            initial_filters: Number of filters in first layer
            use_batch_norm: Whether to use batch normalization
            
        Returns:
            Keras model
        """
        import tensorflow as tf
        from tensorflow.keras import layers, Model
        
        def conv_block(x, filters: int, use_bn: bool = True):
            """Double convolution block."""
            x = layers.Conv2D(filters, 3, padding='same', kernel_initializer='he_normal')(x)
            if use_bn:
                x = layers.BatchNormalization()(x)
            x = layers.Activation('relu')(x)
            
            x = layers.Conv2D(filters, 3, padding='same', kernel_initializer='he_normal')(x)
            if use_bn:
                x = layers.BatchNormalization()(x)
            x = layers.Activation('relu')(x)
            return x
        
        def encoder_block(x, filters: int, use_bn: bool = True):
            """Encoder block with conv + pooling."""
            conv = conv_block(x, filters, use_bn)
            pool = layers.MaxPooling2D(pool_size=(2, 2))(conv)
            return conv, pool
        
        def decoder_block(x, skip, filters: int, use_bn: bool = True):
            """Decoder block with upsampling + skip connection."""
            x = layers.Conv2DTranspose(filters, 2, strides=2, padding='same')(x)
            x = layers.Concatenate()([x, skip])
            x = conv_block(x, filters, use_bn)
            return x
        
        # Input layer
        inputs = layers.Input(shape=input_shape)
        
        # Encoder path
        encoder_outputs = []
        x = inputs
        filters = initial_filters
        
        for i in range(depth):
            conv, x = encoder_block(x, filters, use_batch_norm)
            encoder_outputs.append(conv)
            filters *= 2
        
        # Bridge/bottleneck
        x = conv_block(x, filters, use_batch_norm)
        
        # Decoder path
        for i in range(depth - 1, -1, -1):
            filters //= 2
            x = decoder_block(x, encoder_outputs[i], filters, use_batch_norm)
        
        # Output layer (sigmoid for binary segmentation)
        outputs = layers.Conv2D(1, 1, activation='sigmoid')(x)
        
        model = Model(inputs=inputs, outputs=outputs, name='unet_face_segmentor')
        
        # Compile model
        model.compile(
            optimizer='adam',
            loss='binary_crossentropy',
            metrics=['accuracy']
        )
        
        return model
    
    def _compute_weights_hash(self) -> str:
        """
        Compute hash of model weights for determinism verification.
        
        Returns:
            SHA256 hash of model weights
        """
        if self._model is None:
            return ""
        
        try:
            weights = self._model.get_weights()
            weight_bytes = b''
            for w in weights:
                # Round to reduce floating-point precision issues
                w_rounded = np.round(w, decimals=6)
                weight_bytes += w_rounded.tobytes()
            
            return hashlib.sha256(weight_bytes).hexdigest()
        except Exception as e:
            logger.warning(f"Could not compute weights hash: {e}")
            return ""
    
    @property
    def model_hash(self) -> str:
        """Get the model weights hash."""
        if not self._initialized:
            self._lazy_init()
        return self._model_hash or ""
    
    def _preprocess_image(self, image: np.ndarray) -> np.ndarray:
        """
        Preprocess image for U-Net input.
        
        Args:
            image: Input image (H, W, C) in BGR format
            
        Returns:
            Preprocessed image ready for model input
        """
        import cv2
        
        # Convert BGR to RGB if needed
        if len(image.shape) == 3 and image.shape[2] == 3:
            image_rgb = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
        else:
            image_rgb = image
        
        # Resize to model input size
        target_size = self.config.input_size
        resized = cv2.resize(
            image_rgb, 
            (target_size[1], target_size[0]),  # (W, H)
            interpolation=cv2.INTER_LANCZOS4
        )
        
        # Normalize to [0, 1]
        normalized = resized.astype(np.float32) / 255.0
        
        # Add batch dimension
        batched = np.expand_dims(normalized, axis=0)
        
        return batched
    
    def _postprocess_mask(
        self, 
        mask: np.ndarray, 
        original_size: Tuple[int, int]
    ) -> np.ndarray:
        """
        Postprocess model output mask.
        
        Args:
            mask: Model output (1, H, W, 1)
            original_size: Original image size (H, W)
            
        Returns:
            Mask resized to original size (H, W)
        """
        import cv2
        
        # Remove batch and channel dimensions
        mask_2d = np.squeeze(mask)
        
        # Resize to original size
        resized = cv2.resize(
            mask_2d,
            (original_size[1], original_size[0]),  # (W, H)
            interpolation=cv2.INTER_LINEAR
        )
        
        return resized
    
    def segment(self, image: np.ndarray) -> np.ndarray:
        """
        Segment face region from document image.
        
        Args:
            image: Input document image (H, W, C) in BGR format
            
        Returns:
            Segmentation mask (H, W) with values in [0, 1]
        """
        result = self.segment_with_details(image)
        return result.mask
    
    def segment_with_details(self, image: np.ndarray) -> SegmentationResult:
        """
        Segment face region with detailed results.
        
        Args:
            image: Input document image (H, W, C) in BGR format
            
        Returns:
            SegmentationResult with mask, confidence, and metadata
        """
        if image is None or image.size == 0:
            return SegmentationResult(
                mask=np.zeros((1, 1), dtype=np.float32),
                confidence_map=np.zeros((1, 1), dtype=np.float32),
                success=False,
                mean_confidence=0.0,
                max_confidence=0.0,
                model_version=self.model_version,
                model_hash="",
                error_message="Invalid input image"
            )
        
        try:
            self._lazy_init()
            
            original_size = (image.shape[0], image.shape[1])
            
            # Preprocess
            preprocessed = self._preprocess_image(image)
            
            # Run inference
            raw_output = self._model.predict(preprocessed, verbose=0)
            
            # Postprocess mask
            mask = self._postprocess_mask(raw_output, original_size)
            
            # Compute confidence metrics
            mean_conf = float(np.mean(mask[mask > 0.5])) if np.any(mask > 0.5) else 0.0
            max_conf = float(np.max(mask))
            
            return SegmentationResult(
                mask=mask,
                confidence_map=mask.copy(),  # For U-Net, mask IS confidence
                success=True,
                mean_confidence=mean_conf,
                max_confidence=max_conf,
                model_version=self.model_version,
                model_hash=self.model_hash,
            )
            
        except Exception as e:
            logger.error(f"Segmentation failed: {e}")
            return SegmentationResult(
                mask=np.zeros(image.shape[:2], dtype=np.float32),
                confidence_map=np.zeros(image.shape[:2], dtype=np.float32),
                success=False,
                mean_confidence=0.0,
                max_confidence=0.0,
                model_version=self.model_version,
                model_hash=self.model_hash if self._initialized else "",
                error_message=str(e)
            )
    
    def get_confidence_map(self, image: np.ndarray) -> np.ndarray:
        """
        Get confidence map for segmentation.
        
        Args:
            image: Input document image (H, W, C)
            
        Returns:
            Confidence map (H, W) with values in [0, 1]
        """
        result = self.segment_with_details(image)
        return result.confidence_map
    
    def batch_segment(self, images: list) -> list:
        """
        Segment multiple images in a batch.
        
        Args:
            images: List of document images
            
        Returns:
            List of SegmentationResult objects
        """
        return [self.segment_with_details(img) for img in images]
