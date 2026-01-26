"""
Document enhancement module.

This module provides image enhancement for ID documents:
- CLAHE (Contrast Limited Adaptive Histogram Equalization)
- Brightness/contrast normalization
- Sharpening (unsharp masking)
- Gamma correction
- Auto-levels (histogram stretching)
"""

import logging
from typing import Tuple, Optional, List

import numpy as np
import cv2

from ml.document_preprocessing.config import DocumentConfig, EnhancementConfig

logger = logging.getLogger(__name__)


class DocumentEnhancer:
    """
    Enhances document images for better OCR and visual quality.
    
    This class applies various enhancement techniques to improve
    readability and consistency of document images.
    """
    
    def __init__(self, config: Optional[DocumentConfig] = None):
        """
        Initialize the enhancer.
        
        Args:
            config: Document configuration. Uses defaults if not provided.
        """
        self.config = config or DocumentConfig()
        self.enh_config = self.config.enhancement
        self._clahe = None
    
    @property
    def clahe(self) -> cv2.CLAHE:
        """Lazy initialization of CLAHE."""
        if self._clahe is None:
            self._clahe = cv2.createCLAHE(
                clipLimit=self.enh_config.clahe_clip_limit,
                tileGridSize=self.enh_config.clahe_tile_grid_size
            )
        return self._clahe
    
    def enhance(self, image: np.ndarray) -> Tuple[np.ndarray, List[str]]:
        """
        Apply all configured enhancements to the image.
        
        Args:
            image: Input image (RGB format expected)
            
        Returns:
            Tuple of (enhanced image, list of applied enhancements)
        """
        applied = []
        
        # Apply auto-levels first (histogram stretching)
        if self.enh_config.apply_auto_levels:
            image = self.apply_auto_levels(image)
            applied.append("auto_levels")
        
        # Apply CLAHE for local contrast enhancement
        if self.enh_config.apply_clahe:
            image = self.apply_clahe(image)
            applied.append("clahe")
        
        # Adjust brightness and contrast
        if self.enh_config.apply_brightness_contrast:
            image = self.adjust_brightness_contrast(image)
            applied.append("brightness_contrast")
        
        # Apply gamma correction if enabled
        if self.enh_config.apply_gamma:
            image = self.apply_gamma(image)
            applied.append("gamma")
        
        # Apply sharpening last
        if self.enh_config.apply_sharpening:
            image = self.sharpen(image)
            applied.append("sharpening")
        
        logger.debug(f"Applied enhancements: {applied}")
        
        return image, applied
    
    def apply_clahe(self, image: np.ndarray) -> np.ndarray:
        """
        Apply CLAHE histogram equalization.
        
        CLAHE improves local contrast while limiting amplification
        of noise, making it ideal for document images with varying
        lighting conditions.
        
        Args:
            image: Input RGB image
            
        Returns:
            CLAHE-enhanced image
        """
        # Convert to LAB color space
        lab = cv2.cvtColor(image, cv2.COLOR_RGB2LAB)
        
        # Split channels
        l, a, b = cv2.split(lab)
        
        # Apply CLAHE to L channel only
        l_clahe = self.clahe.apply(l)
        
        # Merge channels back
        lab_clahe = cv2.merge([l_clahe, a, b])
        
        # Convert back to RGB
        result = cv2.cvtColor(lab_clahe, cv2.COLOR_LAB2RGB)
        
        return result
    
    def adjust_brightness_contrast(
        self,
        image: np.ndarray,
        brightness: Optional[float] = None,
        contrast: Optional[float] = None
    ) -> np.ndarray:
        """
        Normalize brightness and contrast.
        
        Args:
            image: Input RGB image
            brightness: Target brightness (0-255). Uses config default if None.
            contrast: Contrast factor (1.0 = no change). Uses config default if None.
            
        Returns:
            Brightness/contrast adjusted image
        """
        if brightness is None:
            brightness = self.enh_config.target_brightness
        if contrast is None:
            contrast = self.enh_config.contrast_factor
        
        # Calculate current brightness
        current_brightness = np.mean(image)
        
        # Calculate brightness adjustment
        brightness_delta = brightness - current_brightness
        
        # Apply formula: output = contrast * (input - 127.5) + 127.5 + brightness_delta
        # This adjusts contrast around mid-gray and adds brightness offset
        image = image.astype(np.float32)
        
        image = contrast * (image - 127.5) + 127.5 + brightness_delta
        
        # Clip to valid range
        image = np.clip(image, 0, 255).astype(np.uint8)
        
        return image
    
    def sharpen(
        self,
        image: np.ndarray,
        amount: Optional[float] = None,
        radius: Optional[int] = None,
        threshold: Optional[int] = None
    ) -> np.ndarray:
        """
        Apply unsharp masking for sharpening.
        
        Unsharp masking enhances edges by subtracting a blurred version
        of the image from the original.
        
        Args:
            image: Input RGB image
            amount: Sharpening strength (1.0 = 100% of difference)
            radius: Blur radius for unsharp mask
            threshold: Minimum difference to apply sharpening
            
        Returns:
            Sharpened image
        """
        if amount is None:
            amount = self.enh_config.sharpening_amount
        if radius is None:
            radius = self.enh_config.sharpening_radius
        if threshold is None:
            threshold = self.enh_config.sharpening_threshold
        
        # Create blurred version
        kernel_size = 2 * radius + 1
        blurred = cv2.GaussianBlur(image, (kernel_size, kernel_size), 0)
        
        # Calculate difference
        image_float = image.astype(np.float32)
        blurred_float = blurred.astype(np.float32)
        
        # Unsharp mask formula: sharpened = original + amount * (original - blurred)
        difference = image_float - blurred_float
        
        # Apply threshold if specified
        if threshold > 0:
            # Only sharpen where difference exceeds threshold
            mask = np.abs(difference) > threshold
            difference = difference * mask
        
        sharpened = image_float + amount * difference
        
        # Clip to valid range
        sharpened = np.clip(sharpened, 0, 255).astype(np.uint8)
        
        return sharpened
    
    def apply_auto_levels(
        self,
        image: np.ndarray,
        percentile: Optional[float] = None
    ) -> np.ndarray:
        """
        Apply auto-levels (histogram stretching).
        
        Stretches the histogram to use the full dynamic range,
        clipping a small percentage of pixels at each end.
        
        Args:
            image: Input RGB image
            percentile: Percentile of pixels to clip at each end
            
        Returns:
            Auto-leveled image
        """
        if percentile is None:
            percentile = self.enh_config.auto_levels_percentile
        
        # Process each channel separately
        result = np.zeros_like(image)
        
        for i in range(3):
            channel = image[:, :, i]
            
            # Calculate percentile values
            low = np.percentile(channel, percentile)
            high = np.percentile(channel, 100 - percentile)
            
            # Avoid division by zero
            if high <= low:
                result[:, :, i] = channel
                continue
            
            # Stretch histogram
            stretched = (channel.astype(np.float32) - low) * 255.0 / (high - low)
            result[:, :, i] = np.clip(stretched, 0, 255).astype(np.uint8)
        
        return result
    
    def apply_gamma(
        self,
        image: np.ndarray,
        gamma: Optional[float] = None
    ) -> np.ndarray:
        """
        Apply gamma correction.
        
        Gamma < 1 brightens the image, gamma > 1 darkens it.
        
        Args:
            image: Input RGB image
            gamma: Gamma value (1.0 = no change)
            
        Returns:
            Gamma-corrected image
        """
        if gamma is None:
            gamma = self.enh_config.gamma_value
        
        if gamma == 1.0:
            return image
        
        # Build lookup table for efficiency
        inv_gamma = 1.0 / gamma
        table = np.array([
            ((i / 255.0) ** inv_gamma) * 255
            for i in np.arange(256)
        ]).astype(np.uint8)
        
        return cv2.LUT(image, table)
    
    def analyze_enhancement_needs(self, image: np.ndarray) -> dict:
        """
        Analyze image to determine what enhancements are needed.
        
        Args:
            image: Input RGB image
            
        Returns:
            Dictionary with analysis results and recommendations
        """
        # Calculate statistics
        brightness = np.mean(image)
        contrast = np.std(image)
        
        # Calculate histogram spread
        gray = cv2.cvtColor(image, cv2.COLOR_RGB2GRAY)
        hist = cv2.calcHist([gray], [0], None, [256], [0, 256])
        hist = hist.flatten() / hist.sum()
        
        # Find used dynamic range
        nonzero_indices = np.where(hist > 0.001)[0]
        if len(nonzero_indices) > 0:
            dynamic_range = nonzero_indices[-1] - nonzero_indices[0]
        else:
            dynamic_range = 0
        
        # Calculate sharpness using Laplacian variance
        laplacian = cv2.Laplacian(gray, cv2.CV_64F)
        sharpness = laplacian.var()
        
        # Make recommendations
        recommendations = []
        
        if brightness < 80:
            recommendations.append("increase_brightness")
        elif brightness > 180:
            recommendations.append("decrease_brightness")
        
        if contrast < 40:
            recommendations.append("increase_contrast")
        
        if dynamic_range < 200:
            recommendations.append("apply_auto_levels")
        
        if sharpness < 100:
            recommendations.append("apply_sharpening")
        
        return {
            "brightness": float(brightness),
            "contrast": float(contrast),
            "dynamic_range": int(dynamic_range),
            "sharpness": float(sharpness),
            "recommendations": recommendations,
            "needs_enhancement": len(recommendations) > 0,
        }
