"""
Document standardization module.

This module handles format and resolution standardization for ID documents:
- Convert to consistent format (PNG)
- Resize to target resolution
- Ensure RGB color space
- Handle aspect ratio preservation
"""

import logging
from typing import Tuple, Optional

import numpy as np
import cv2

from ml.document_preprocessing.config import (
    DocumentConfig,
    StandardizationConfig,
    InterpolationMethod,
)

logger = logging.getLogger(__name__)


# Mapping from config interpolation to OpenCV constants
INTERPOLATION_MAP = {
    InterpolationMethod.NEAREST: cv2.INTER_NEAREST,
    InterpolationMethod.LINEAR: cv2.INTER_LINEAR,
    InterpolationMethod.CUBIC: cv2.INTER_CUBIC,
    InterpolationMethod.LANCZOS: cv2.INTER_LANCZOS4,
    InterpolationMethod.AREA: cv2.INTER_AREA,
}


class DocumentStandardizer:
    """
    Standardizes document images to consistent format and resolution.
    
    This class ensures all document images are processed to a consistent
    format (PNG), resolution, and color space for downstream processing.
    """
    
    def __init__(self, config: Optional[DocumentConfig] = None):
        """
        Initialize the standardizer.
        
        Args:
            config: Document configuration. Uses defaults if not provided.
        """
        self.config = config or DocumentConfig()
        self.std_config = self.config.standardization
        
    @property
    def target_width(self) -> int:
        """Target width in pixels."""
        return self.std_config.target_width
    
    @property
    def target_height(self) -> int:
        """Target height in pixels."""
        return self.std_config.target_height
    
    @property
    def output_format(self) -> str:
        """Output format string."""
        return self.std_config.output_format.value
    
    def standardize(self, image: np.ndarray) -> np.ndarray:
        """
        Standardize document image format and resolution.
        
        Args:
            image: Input image as numpy array (BGR format expected from OpenCV)
            
        Returns:
            Standardized image as numpy array
            
        Raises:
            ValueError: If image is invalid or too small
        """
        # Validate input
        self._validate_input(image)
        
        # Convert to RGB if needed (our standard internal format)
        image = self._ensure_rgb(image)
        
        # Resize to target resolution
        image = self._resize(image)
        
        # Ensure 8-bit depth
        image = self._ensure_8bit(image)
        
        logger.debug(
            f"Standardized image to {image.shape[1]}x{image.shape[0]} "
            f"{self.output_format}"
        )
        
        return image
    
    def _validate_input(self, image: np.ndarray) -> None:
        """
        Validate input image.
        
        Args:
            image: Image to validate
            
        Raises:
            ValueError: If image is invalid
        """
        if image is None:
            raise ValueError("Image cannot be None")
        
        if not isinstance(image, np.ndarray):
            raise ValueError(f"Expected numpy array, got {type(image)}")
        
        if image.size == 0:
            raise ValueError("Image is empty")
        
        if len(image.shape) < 2:
            raise ValueError(f"Invalid image dimensions: {image.shape}")
        
        height, width = image.shape[:2]
        
        if width < self.std_config.min_width or height < self.std_config.min_height:
            raise ValueError(
                f"Image resolution {width}x{height} is below minimum "
                f"{self.std_config.min_width}x{self.std_config.min_height}"
            )
    
    def _ensure_rgb(self, image: np.ndarray) -> np.ndarray:
        """
        Ensure image is in RGB format.
        
        Args:
            image: Input image (may be grayscale, BGR, BGRA, etc.)
            
        Returns:
            RGB image
        """
        if len(image.shape) == 2:
            # Grayscale to RGB
            return cv2.cvtColor(image, cv2.COLOR_GRAY2RGB)
        
        channels = image.shape[2]
        
        if channels == 4:
            # BGRA to RGB
            return cv2.cvtColor(image, cv2.COLOR_BGRA2RGB)
        elif channels == 3:
            # Assume BGR, convert to RGB
            return cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
        else:
            raise ValueError(f"Unsupported number of channels: {channels}")
    
    def _resize(self, image: np.ndarray) -> np.ndarray:
        """
        Resize image to target resolution.
        
        Args:
            image: Input image
            
        Returns:
            Resized image
        """
        height, width = image.shape[:2]
        target_width = self.std_config.target_width
        target_height = self.std_config.target_height
        
        # Check if resize needed
        if width == target_width and height == target_height:
            return image
        
        # Get interpolation method
        interp = INTERPOLATION_MAP.get(
            self.std_config.interpolation,
            cv2.INTER_LANCZOS4
        )
        
        # Use INTER_AREA for downscaling (better quality)
        if width > target_width or height > target_height:
            interp = cv2.INTER_AREA
        
        if self.std_config.maintain_aspect_ratio:
            return self._resize_with_aspect_ratio(
                image, target_width, target_height, interp
            )
        else:
            return cv2.resize(image, (target_width, target_height), interpolation=interp)
    
    def _resize_with_aspect_ratio(
        self,
        image: np.ndarray,
        target_width: int,
        target_height: int,
        interpolation: int
    ) -> np.ndarray:
        """
        Resize image while maintaining aspect ratio, padding if needed.
        
        Args:
            image: Input image
            target_width: Target width
            target_height: Target height
            interpolation: OpenCV interpolation constant
            
        Returns:
            Resized and padded image
        """
        height, width = image.shape[:2]
        
        # Calculate scaling factor
        scale = min(target_width / width, target_height / height)
        
        # Calculate new dimensions
        new_width = int(width * scale)
        new_height = int(height * scale)
        
        # Resize
        resized = cv2.resize(
            image,
            (new_width, new_height),
            interpolation=interpolation
        )
        
        # Create padded output
        channels = image.shape[2] if len(image.shape) > 2 else 1
        output = np.full(
            (target_height, target_width, channels),
            self.std_config.padding_color,
            dtype=image.dtype
        )
        
        # Center the resized image
        x_offset = (target_width - new_width) // 2
        y_offset = (target_height - new_height) // 2
        
        output[y_offset:y_offset + new_height, x_offset:x_offset + new_width] = resized
        
        return output
    
    def _ensure_8bit(self, image: np.ndarray) -> np.ndarray:
        """
        Ensure image has 8-bit depth.
        
        Args:
            image: Input image
            
        Returns:
            8-bit image
        """
        if image.dtype == np.uint8:
            return image
        
        if image.dtype == np.float32 or image.dtype == np.float64:
            # Assume 0-1 range for float
            if image.max() <= 1.0:
                return (image * 255).astype(np.uint8)
            else:
                return np.clip(image, 0, 255).astype(np.uint8)
        
        if image.dtype == np.uint16:
            # 16-bit to 8-bit
            return (image / 256).astype(np.uint8)
        
        # Try to convert
        return image.astype(np.uint8)
    
    def get_resize_info(
        self,
        original_size: Tuple[int, int]
    ) -> dict:
        """
        Get information about how an image would be resized.
        
        Args:
            original_size: Original (width, height)
            
        Returns:
            Dictionary with resize information
        """
        width, height = original_size
        target_width = self.std_config.target_width
        target_height = self.std_config.target_height
        
        if self.std_config.maintain_aspect_ratio:
            scale = min(target_width / width, target_height / height)
            new_width = int(width * scale)
            new_height = int(height * scale)
            padding = (
                target_width - new_width,
                target_height - new_height
            )
        else:
            scale = (target_width / width, target_height / height)
            new_width = target_width
            new_height = target_height
            padding = (0, 0)
        
        return {
            "original_size": original_size,
            "target_size": (target_width, target_height),
            "scaled_size": (new_width, new_height),
            "scale_factor": scale,
            "padding": padding,
            "aspect_ratio_maintained": self.std_config.maintain_aspect_ratio,
        }
