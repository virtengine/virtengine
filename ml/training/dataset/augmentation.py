"""
Data augmentation for training samples.

This module provides augmentation techniques to increase training data
diversity and improve model robustness.
"""

import logging
import random
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Callable

import numpy as np

from ml.training.config import AugmentationConfig, AugmentationType
from ml.training.dataset.preprocessing import PreprocessedSample

logger = logging.getLogger(__name__)


@dataclass
class AugmentedSample:
    """An augmented training sample."""
    
    # Original sample reference
    original_sample_id: str
    augmentation_id: str  # Unique ID for this augmentation
    
    # Augmented images
    document_image: Optional[np.ndarray] = None
    selfie_image: Optional[np.ndarray] = None
    
    # Augmentations applied
    augmentations_applied: List[str] = field(default_factory=list)
    
    # Original sample data
    original_sample: Optional[PreprocessedSample] = None
    
    # Label (same as original)
    trust_score: float = 0.0


class DataAugmentation:
    """
    Data augmentation for training samples.
    
    Applies various augmentations to increase dataset diversity:
    - Brightness adjustment
    - Contrast adjustment
    - Small rotations
    - Gaussian blur
    - Gaussian noise
    - Perspective transforms
    - JPEG compression artifacts
    """
    
    def __init__(self, config: Optional[AugmentationConfig] = None):
        """
        Initialize the data augmentation.
        
        Args:
            config: Augmentation configuration
        """
        self.config = config or AugmentationConfig()
        self._rng = random.Random()
        
        # Build augmentation functions
        self._augmentation_funcs = self._build_augmentation_funcs()
    
    def _build_augmentation_funcs(self) -> dict:
        """Build dictionary of augmentation functions."""
        return {
            AugmentationType.BRIGHTNESS.value: self._augment_brightness,
            AugmentationType.CONTRAST.value: self._augment_contrast,
            AugmentationType.ROTATION.value: self._augment_rotation,
            AugmentationType.BLUR.value: self._augment_blur,
            AugmentationType.NOISE.value: self._augment_noise,
            AugmentationType.PERSPECTIVE.value: self._augment_perspective,
            AugmentationType.JPEG_ARTIFACT.value: self._augment_jpeg,
        }
    
    def augment_batch(
        self,
        samples: List[PreprocessedSample],
        seed: Optional[int] = None
    ) -> List[AugmentedSample]:
        """
        Augment a batch of samples.
        
        Args:
            samples: List of preprocessed samples
            seed: Random seed for reproducibility
            
        Returns:
            List of augmented samples (original + augmented copies)
        """
        if not self.config.enabled:
            # Return samples as-is (wrapped in AugmentedSample)
            return [
                AugmentedSample(
                    original_sample_id=s.sample_id,
                    augmentation_id=f"{s.sample_id}_orig",
                    document_image=s.document_image,
                    selfie_image=s.selfie_image,
                    augmentations_applied=[],
                    original_sample=s,
                    trust_score=s.original_sample.trust_score if s.original_sample else 0.0,
                )
                for s in samples
            ]
        
        if seed is not None:
            self._rng.seed(seed)
        
        augmented = []
        
        for sample in samples:
            # Add original sample
            augmented.append(
                AugmentedSample(
                    original_sample_id=sample.sample_id,
                    augmentation_id=f"{sample.sample_id}_orig",
                    document_image=sample.document_image,
                    selfie_image=sample.selfie_image,
                    augmentations_applied=[],
                    original_sample=sample,
                    trust_score=sample.original_sample.trust_score if sample.original_sample else 0.0,
                )
            )
            
            # Add augmented copies
            for i in range(self.config.num_augmented_copies):
                aug_sample = self._augment_sample(sample, i)
                augmented.append(aug_sample)
        
        logger.info(
            f"Augmented {len(samples)} samples to {len(augmented)} "
            f"({self.config.num_augmented_copies} copies each)"
        )
        
        return augmented
    
    def _augment_sample(
        self,
        sample: PreprocessedSample,
        copy_index: int
    ) -> AugmentedSample:
        """Augment a single sample."""
        augmentations_applied = []
        
        # Augment document image
        document_image = sample.document_image
        if document_image is not None:
            document_image, doc_augs = self._apply_augmentations(
                document_image.copy()
            )
            augmentations_applied.extend([f"doc_{a}" for a in doc_augs])
        
        # Augment selfie image
        selfie_image = sample.selfie_image
        if selfie_image is not None:
            selfie_image, selfie_augs = self._apply_augmentations(
                selfie_image.copy()
            )
            augmentations_applied.extend([f"selfie_{a}" for a in selfie_augs])
        
        return AugmentedSample(
            original_sample_id=sample.sample_id,
            augmentation_id=f"{sample.sample_id}_aug{copy_index}",
            document_image=document_image,
            selfie_image=selfie_image,
            augmentations_applied=augmentations_applied,
            original_sample=sample,
            trust_score=sample.original_sample.trust_score if sample.original_sample else 0.0,
        )
    
    def _apply_augmentations(
        self,
        image: np.ndarray
    ) -> Tuple[np.ndarray, List[str]]:
        """Apply random augmentations to an image."""
        applied = []
        
        for aug_type in self.config.augmentation_types:
            if self._rng.random() < self.config.augmentation_probability:
                if aug_type in self._augmentation_funcs:
                    image = self._augmentation_funcs[aug_type](image)
                    applied.append(aug_type)
        
        return image, applied
    
    def _augment_brightness(self, image: np.ndarray) -> np.ndarray:
        """Adjust brightness."""
        min_val, max_val = self.config.brightness_range
        factor = self._rng.uniform(min_val, max_val)
        
        # Handle normalized images
        if image.dtype == np.float32:
            return np.clip(image * factor, -3.0, 3.0)
        else:
            return np.clip(image * factor, 0, 255).astype(np.uint8)
    
    def _augment_contrast(self, image: np.ndarray) -> np.ndarray:
        """Adjust contrast."""
        min_val, max_val = self.config.contrast_range
        factor = self._rng.uniform(min_val, max_val)
        
        mean = np.mean(image)
        
        if image.dtype == np.float32:
            return np.clip((image - mean) * factor + mean, -3.0, 3.0)
        else:
            return np.clip((image - mean) * factor + mean, 0, 255).astype(np.uint8)
    
    def _augment_rotation(self, image: np.ndarray) -> np.ndarray:
        """Apply small rotation."""
        min_angle, max_angle = self.config.rotation_range
        angle = self._rng.uniform(min_angle, max_angle)
        
        try:
            from scipy.ndimage import rotate
            return rotate(image, angle, reshape=False, mode='nearest')
        except ImportError:
            # Fallback without scipy
            return image
    
    def _augment_blur(self, image: np.ndarray) -> np.ndarray:
        """Apply Gaussian blur."""
        min_k, max_k = self.config.blur_kernel_range
        kernel_size = self._rng.randrange(min_k, max_k + 1, 2)  # Must be odd
        
        try:
            from scipy.ndimage import gaussian_filter
            sigma = kernel_size / 6.0
            if len(image.shape) == 3:
                # Apply per channel
                for c in range(image.shape[2]):
                    image[:, :, c] = gaussian_filter(image[:, :, c], sigma=sigma)
            else:
                image = gaussian_filter(image, sigma=sigma)
            return image
        except ImportError:
            return image
    
    def _augment_noise(self, image: np.ndarray) -> np.ndarray:
        """Add Gaussian noise."""
        min_std, max_std = self.config.noise_std_range
        std = self._rng.uniform(min_std, max_std)
        
        noise = np.random.normal(0, std, image.shape).astype(image.dtype)
        
        if image.dtype == np.float32:
            return np.clip(image + noise, -3.0, 3.0)
        else:
            return np.clip(image + noise * 255, 0, 255).astype(np.uint8)
    
    def _augment_perspective(self, image: np.ndarray) -> np.ndarray:
        """Apply perspective transform."""
        # Simplified perspective - just use slight skew
        strength = self.config.perspective_strength
        
        try:
            from scipy.ndimage import affine_transform
            
            # Create slight perspective matrix
            skew_x = self._rng.uniform(-strength, strength)
            skew_y = self._rng.uniform(-strength, strength)
            
            matrix = np.array([
                [1, skew_x, 0],
                [skew_y, 1, 0],
            ])
            
            if len(image.shape) == 3:
                result = np.zeros_like(image)
                for c in range(image.shape[2]):
                    result[:, :, c] = affine_transform(
                        image[:, :, c], matrix[:, :2], offset=matrix[:, 2]
                    )
                return result
            else:
                return affine_transform(image, matrix[:, :2], offset=matrix[:, 2])
        except ImportError:
            return image
    
    def _augment_jpeg(self, image: np.ndarray) -> np.ndarray:
        """Simulate JPEG compression artifacts."""
        min_q, max_q = self.config.jpeg_quality_range
        quality = self._rng.randint(min_q, max_q)
        
        try:
            from PIL import Image
            import io
            
            # Convert to uint8 for JPEG compression
            if image.dtype == np.float32:
                # Denormalize
                img_uint8 = ((image + 3.0) / 6.0 * 255).astype(np.uint8)
            else:
                img_uint8 = image
            
            # Compress and decompress
            pil_image = Image.fromarray(img_uint8)
            buffer = io.BytesIO()
            pil_image.save(buffer, format='JPEG', quality=quality)
            buffer.seek(0)
            compressed = Image.open(buffer)
            result = np.array(compressed)
            
            # Re-normalize if needed
            if image.dtype == np.float32:
                result = result.astype(np.float32) / 255.0 * 6.0 - 3.0
            
            return result
        except ImportError:
            return image
