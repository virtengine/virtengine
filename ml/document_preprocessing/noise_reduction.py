"""
Noise reduction module.

This module provides various noise reduction techniques for ID documents:
- Gaussian blur
- Median blur
- Bilateral filtering
- Morphological operations
"""

import logging
from typing import Optional

import numpy as np
import cv2

from ml.document_preprocessing.config import (
    DocumentConfig,
    NoiseReductionConfig,
    NoiseReductionMethod,
)

logger = logging.getLogger(__name__)


class NoiseReducer:
    """
    Reduces noise in document images.
    
    This class provides multiple noise reduction methods optimized
    for document images, with bilateral filtering as the default
    for its edge-preserving properties.
    """
    
    def __init__(self, config: Optional[DocumentConfig] = None):
        """
        Initialize the noise reducer.
        
        Args:
            config: Document configuration. Uses defaults if not provided.
        """
        self.config = config or DocumentConfig()
        self.nr_config = self.config.noise_reduction
    
    def denoise(
        self,
        image: np.ndarray,
        method: Optional[str] = None
    ) -> np.ndarray:
        """
        Apply noise reduction using the specified or configured method.
        
        Args:
            image: Input image (RGB format)
            method: Noise reduction method ('gaussian', 'median', 'bilateral').
                   Uses config default if not specified.
                   
        Returns:
            Denoised image
        """
        if not self.nr_config.enabled:
            return image
        
        if method is None:
            method = self.nr_config.method.value
        
        # Convert string to enum if needed
        if isinstance(method, str):
            method = method.lower()
        
        if method == NoiseReductionMethod.GAUSSIAN.value or method == "gaussian":
            result = self.gaussian_blur(image)
        elif method == NoiseReductionMethod.MEDIAN.value or method == "median":
            result = self.median_blur(image)
        elif method == NoiseReductionMethod.BILATERAL.value or method == "bilateral":
            result = self.bilateral_filter(image)
        else:
            logger.warning(f"Unknown noise reduction method: {method}, using bilateral")
            result = self.bilateral_filter(image)
        
        # Apply morphological noise reduction if enabled
        if self.nr_config.apply_morphological:
            result = self._morphological_clean(result)
        
        return result
    
    def gaussian_blur(
        self,
        image: np.ndarray,
        ksize: Optional[int] = None,
        sigma: Optional[float] = None
    ) -> np.ndarray:
        """
        Apply Gaussian blur for noise reduction.
        
        Gaussian blur is fast and effective for general noise reduction,
        but it also blurs edges.
        
        Args:
            image: Input image
            ksize: Kernel size (must be odd). Uses config if not specified.
            sigma: Gaussian standard deviation. Uses config if not specified.
            
        Returns:
            Blurred image
        """
        if ksize is None:
            ksize = self.nr_config.gaussian_kernel_size
        if sigma is None:
            sigma = self.nr_config.gaussian_sigma
        
        # Ensure kernel size is odd
        if ksize % 2 == 0:
            ksize += 1
        
        return cv2.GaussianBlur(image, (ksize, ksize), sigma)
    
    def median_blur(
        self,
        image: np.ndarray,
        ksize: Optional[int] = None
    ) -> np.ndarray:
        """
        Apply median blur for noise reduction.
        
        Median blur is particularly effective for salt-and-pepper noise
        and preserves edges better than Gaussian blur.
        
        Args:
            image: Input image
            ksize: Kernel size (must be odd). Uses config if not specified.
            
        Returns:
            Blurred image
        """
        if ksize is None:
            ksize = self.nr_config.median_kernel_size
        
        # Ensure kernel size is odd
        if ksize % 2 == 0:
            ksize += 1
        
        return cv2.medianBlur(image, ksize)
    
    def bilateral_filter(
        self,
        image: np.ndarray,
        d: Optional[int] = None,
        sigma_color: Optional[float] = None,
        sigma_space: Optional[float] = None
    ) -> np.ndarray:
        """
        Apply bilateral filtering for edge-preserving noise reduction.
        
        Bilateral filtering smooths the image while preserving edges,
        making it ideal for document images where text and graphics
        need to remain sharp.
        
        Args:
            image: Input image
            d: Diameter of each pixel neighborhood. Uses config if not specified.
            sigma_color: Filter sigma in the color space.
            sigma_space: Filter sigma in the coordinate space.
            
        Returns:
            Filtered image
        """
        if d is None:
            d = self.nr_config.bilateral_d
        if sigma_color is None:
            sigma_color = self.nr_config.bilateral_sigma_color
        if sigma_space is None:
            sigma_space = self.nr_config.bilateral_sigma_space
        
        return cv2.bilateralFilter(image, d, sigma_color, sigma_space)
    
    def _morphological_clean(self, image: np.ndarray) -> np.ndarray:
        """
        Apply morphological operations for additional cleaning.
        
        This applies opening (erosion followed by dilation) to remove
        small noise while preserving the document structure.
        
        Args:
            image: Input image
            
        Returns:
            Cleaned image
        """
        ksize = self.nr_config.morphological_kernel_size
        kernel = cv2.getStructuringElement(
            cv2.MORPH_ELLIPSE,
            (ksize, ksize)
        )
        
        # Apply opening (erosion + dilation) to remove noise
        return cv2.morphologyEx(image, cv2.MORPH_OPEN, kernel)
    
    def estimate_noise_level(self, image: np.ndarray) -> float:
        """
        Estimate the noise level in an image.
        
        Uses the Laplacian method to estimate Gaussian noise.
        
        Args:
            image: Input image (RGB or grayscale)
            
        Returns:
            Estimated noise standard deviation
        """
        # Convert to grayscale if needed
        if len(image.shape) == 3:
            gray = cv2.cvtColor(image, cv2.COLOR_RGB2GRAY)
        else:
            gray = image
        
        # Compute Laplacian
        laplacian = cv2.Laplacian(gray, cv2.CV_64F)
        
        # Estimate noise using MAD (Median Absolute Deviation)
        # sigma = median(|laplacian|) / 0.6745
        sigma = np.median(np.abs(laplacian)) / 0.6745
        
        return float(sigma)
    
    def adaptiVIRTENGINE_denoise(self, image: np.ndarray) -> np.ndarray:
        """
        Apply adaptive noise reduction based on estimated noise level.
        
        Automatically selects the appropriate method and parameters
        based on the estimated noise in the image.
        
        Args:
            image: Input image
            
        Returns:
            Denoised image
        """
        noise_level = self.estimate_noise_level(image)
        
        logger.debug(f"Estimated noise level: {noise_level:.2f}")
        
        if noise_level < 5:
            # Very low noise, minimal denoising
            return self.gaussian_blur(image, ksize=3, sigma=0.5)
        elif noise_level < 15:
            # Moderate noise, bilateral filter
            return self.bilateral_filter(image)
        elif noise_level < 30:
            # High noise, stronger bilateral
            return self.bilateral_filter(
                image,
                d=11,
                sigma_color=100,
                sigma_space=100
            )
        else:
            # Very high noise, combination approach
            # First median to remove salt-and-pepper
            denoised = self.median_blur(image, ksize=3)
            # Then bilateral for remaining noise
            return self.bilateral_filter(
                denoised,
                d=11,
                sigma_color=100,
                sigma_space=100
            )
    
    def non_local_means(
        self,
        image: np.ndarray,
        h: float = 10.0,
        template_window_size: int = 7,
        search_window_size: int = 21
    ) -> np.ndarray:
        """
        Apply Non-Local Means denoising.
        
        NLM is more computationally expensive but provides excellent
        noise reduction while preserving fine details.
        
        Args:
            image: Input image (RGB format)
            h: Filter strength. Higher h removes more noise but loses detail.
            template_window_size: Size of template patch (should be odd).
            search_window_size: Size of area where search is performed (should be odd).
            
        Returns:
            Denoised image
        """
        # OpenCV's fastNlMeansDenoisingColored expects BGR
        bgr = cv2.cvtColor(image, cv2.COLOR_RGB2BGR)
        
        denoised_bgr = cv2.fastNlMeansDenoisingColored(
            bgr,
            None,
            h,
            h,
            template_window_size,
            search_window_size
        )
        
        # Convert back to RGB
        return cv2.cvtColor(denoised_bgr, cv2.COLOR_BGR2RGB)
