"""
Image preprocessing pipeline for facial verification.

This module implements the preprocessing steps required before face detection
and embedding extraction:
1. Resolution standardization
2. Color space conversion (optional grayscale)
3. CLAHE histogram equalization
4. Noise reduction
5. Pixel normalization
"""

import logging
from dataclasses import dataclass
from typing import Tuple, Optional, List
import hashlib

import numpy as np
import cv2

from ml.facial_verification.config import PreprocessingConfig, VerificationConfig
from ml.facial_verification.reason_codes import ReasonCodes

logger = logging.getLogger(__name__)


@dataclass
class PreprocessingResult:
    """Result of image preprocessing."""
    
    image: np.ndarray
    original_size: Tuple[int, int]
    final_size: Tuple[int, int]
    applied_steps: List[str]
    success: bool
    error_code: Optional[ReasonCodes] = None
    image_hash: Optional[str] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "original_size": self.original_size,
            "final_size": self.final_size,
            "applied_steps": self.applied_steps,
            "success": self.success,
            "error_code": self.error_code.value if self.error_code else None,
            "image_hash": self.image_hash,
        }


class FacePreprocessor:
    """
    Image preprocessing pipeline for facial verification.
    
    This class implements a deterministic preprocessing pipeline that prepares
    images for face detection and embedding extraction. All operations are
    designed to produce consistent results for blockchain consensus.
    """
    
    def __init__(self, config: Optional[VerificationConfig] = None):
        """
        Initialize the preprocessor.
        
        Args:
            config: Verification configuration. Uses defaults if not provided.
        """
        self.config = config or VerificationConfig()
        self.preprocessing_config = self.config.preprocessing
        self._clahe = None
        
    @property
    def clahe(self) -> cv2.CLAHE:
        """Lazy initialization of CLAHE."""
        if self._clahe is None:
            self._clahe = cv2.createCLAHE(
                clipLimit=self.preprocessing_config.clahe_clip_limit,
                tileGridSize=self.preprocessing_config.clahe_tile_grid_size
            )
        return self._clahe
    
    def preprocess(
        self, 
        image: np.ndarray,
        compute_hash: bool = True
    ) -> PreprocessingResult:
        """
        Apply full preprocessing pipeline to an image.
        
        Args:
            image: Input image as numpy array (BGR or RGB format)
            compute_hash: Whether to compute hash of preprocessed image
            
        Returns:
            PreprocessingResult with preprocessed image and metadata
        """
        applied_steps = []
        
        # Validate input
        validation_result = self._validate_image(image)
        if validation_result is not None:
            return PreprocessingResult(
                image=image,
                original_size=(0, 0),
                final_size=(0, 0),
                applied_steps=[],
                success=False,
                error_code=validation_result
            )
        
        original_size = (image.shape[1], image.shape[0])  # width, height
        
        try:
            # Step 1: Resize to standard resolution
            image = self._resize(image)
            applied_steps.append("resize")
            
            # Step 2: Optional grayscale conversion
            if self.preprocessing_config.use_grayscale:
                image = self._to_grayscale(image)
                applied_steps.append("grayscale")
            
            # Step 3: CLAHE histogram equalization
            if self.preprocessing_config.apply_clahe:
                image = self._apply_clahe(image)
                applied_steps.append("clahe")
            
            # Step 4: Noise reduction
            if self.preprocessing_config.noise_reduction:
                image = self._reduce_noise(image)
                applied_steps.append("noise_reduction")
            
            # Step 5: Normalize pixel values
            if self.preprocessing_config.normalize_color:
                image = self._normalize(image)
                applied_steps.append("normalize")
            
            final_size = (image.shape[1], image.shape[0])
            
            # Compute hash for determinism verification
            image_hash = None
            if compute_hash:
                image_hash = self._compute_image_hash(image)
            
            return PreprocessingResult(
                image=image,
                original_size=original_size,
                final_size=final_size,
                applied_steps=applied_steps,
                success=True,
                image_hash=image_hash
            )
            
        except Exception as e:
            logger.error(f"Preprocessing error: {e}")
            return PreprocessingResult(
                image=image,
                original_size=original_size,
                final_size=(0, 0),
                applied_steps=applied_steps,
                success=False,
                error_code=ReasonCodes.PREPROCESSING_ERROR
            )
    
    def _validate_image(self, image: np.ndarray) -> Optional[ReasonCodes]:
        """Validate input image."""
        if image is None:
            return ReasonCodes.INVALID_IMAGE_FORMAT
        
        if not isinstance(image, np.ndarray):
            return ReasonCodes.INVALID_IMAGE_FORMAT
        
        if image.size == 0:
            return ReasonCodes.CORRUPT_IMAGE_DATA
        
        if len(image.shape) < 2:
            return ReasonCodes.INVALID_IMAGE_FORMAT
        
        height, width = image.shape[:2]
        min_w, min_h = self.preprocessing_config.min_resolution
        
        if width < min_w or height < min_h:
            return ReasonCodes.LOW_RESOLUTION
        
        return None
    
    def _resize(self, image: np.ndarray) -> np.ndarray:
        """
        Resize image to target resolution while maintaining aspect ratio.
        
        Args:
            image: Input image
            
        Returns:
            Resized image
        """
        target_w, target_h = self.preprocessing_config.target_resolution
        height, width = image.shape[:2]
        
        # Calculate aspect ratio
        aspect = width / height
        target_aspect = target_w / target_h
        
        if aspect > target_aspect:
            # Image is wider - fit to width
            new_width = target_w
            new_height = int(target_w / aspect)
        else:
            # Image is taller - fit to height
            new_height = target_h
            new_width = int(target_h * aspect)
        
        # Resize using INTER_AREA for downscaling, INTER_LINEAR for upscaling
        if new_width < width or new_height < height:
            interpolation = cv2.INTER_AREA
        else:
            interpolation = cv2.INTER_LINEAR
        
        resized = cv2.resize(image, (new_width, new_height), interpolation=interpolation)
        
        # Pad to exact target size if needed
        if new_width != target_w or new_height != target_h:
            padded = np.zeros((target_h, target_w, 3), dtype=image.dtype)
            y_offset = (target_h - new_height) // 2
            x_offset = (target_w - new_width) // 2
            if len(resized.shape) == 2:
                padded = np.zeros((target_h, target_w), dtype=image.dtype)
            padded[y_offset:y_offset+new_height, x_offset:x_offset+new_width] = resized
            return padded
        
        return resized
    
    def _to_grayscale(self, image: np.ndarray) -> np.ndarray:
        """
        Convert image to grayscale.
        
        Args:
            image: Input BGR image
            
        Returns:
            Grayscale image (still 3-channel for compatibility)
        """
        if len(image.shape) == 2:
            return image
        
        gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        # Keep 3 channels for model compatibility
        return cv2.cvtColor(gray, cv2.COLOR_GRAY2BGR)
    
    def _apply_clahe(self, image: np.ndarray) -> np.ndarray:
        """
        Apply Contrast Limited Adaptive Histogram Equalization.
        
        CLAHE improves local contrast and reduces the effect of
        non-uniform lighting conditions.
        
        Args:
            image: Input image (BGR or grayscale)
            
        Returns:
            CLAHE-enhanced image
        """
        if len(image.shape) == 2:
            return self.clahe.apply(image)
        
        # Convert to LAB color space
        lab = cv2.cvtColor(image, cv2.COLOR_BGR2LAB)
        l, a, b = cv2.split(lab)
        
        # Apply CLAHE to L channel only
        l = self.clahe.apply(l)
        
        # Merge and convert back to BGR
        lab = cv2.merge([l, a, b])
        return cv2.cvtColor(lab, cv2.COLOR_LAB2BGR)
    
    def _reduce_noise(self, image: np.ndarray) -> np.ndarray:
        """
        Apply noise reduction filter.
        
        Args:
            image: Input image
            
        Returns:
            Denoised image
        """
        method = self.preprocessing_config.noise_reduction_method
        
        if method == "bilateral":
            return cv2.bilateralFilter(
                image,
                d=self.preprocessing_config.bilateral_d,
                sigmaColor=self.preprocessing_config.bilateral_sigma_color,
                sigmaSpace=self.preprocessing_config.bilateral_sigma_space
            )
        elif method == "gaussian":
            return cv2.GaussianBlur(
                image, 
                self.preprocessing_config.gaussian_kernel_size, 
                0
            )
        else:
            return image
    
    def _normalize(self, image: np.ndarray) -> np.ndarray:
        """
        Normalize pixel values.
        
        Args:
            image: Input image
            
        Returns:
            Normalized image (float32)
        """
        image = image.astype(np.float32)
        method = self.preprocessing_config.pixel_normalization
        
        if method == "minmax":
            # Normalize to [0, 1]
            min_val = image.min()
            max_val = image.max()
            if max_val - min_val > 0:
                image = (image - min_val) / (max_val - min_val)
            else:
                image = np.zeros_like(image)
        
        elif method == "zscore":
            # Z-score normalization
            mean = image.mean()
            std = image.std()
            if std > 0:
                image = (image - mean) / std
            else:
                image = np.zeros_like(image)
        
        elif method == "fixed":
            # Use ImageNet normalization
            image = image / 255.0
            mean = np.array(self.preprocessing_config.normalize_mean)
            std = np.array(self.preprocessing_config.normalize_std)
            if len(image.shape) == 3:
                image = (image - mean) / std
        
        return image
    
    def _compute_image_hash(self, image: np.ndarray) -> str:
        """
        Compute deterministic hash of image.
        
        Args:
            image: Input image
            
        Returns:
            SHA256 hash of image bytes
        """
        # Convert to contiguous array with consistent dtype
        if image.dtype != np.float32:
            image = image.astype(np.float32)
        
        # Round to reduce floating-point precision issues
        image = np.round(image, decimals=6)
        
        # Compute hash
        return hashlib.sha256(image.tobytes()).hexdigest()
    
    def check_image_quality(self, image: np.ndarray) -> Tuple[float, List[ReasonCodes]]:
        """
        Assess image quality and return score with issues.
        
        Args:
            image: Input image
            
        Returns:
            Tuple of (quality_score 0-1, list of issue codes)
        """
        issues = []
        scores = []
        
        # Check brightness
        brightness_score, brightness_issue = self._check_brightness(image)
        scores.append(brightness_score)
        if brightness_issue:
            issues.append(brightness_issue)
        
        # Check blur
        blur_score, blur_issue = self._check_blur(image)
        scores.append(blur_score)
        if blur_issue:
            issues.append(blur_issue)
        
        # Check resolution
        resolution_score, resolution_issue = self._check_resolution(image)
        scores.append(resolution_score)
        if resolution_issue:
            issues.append(resolution_issue)
        
        # Calculate overall quality
        overall_quality = sum(scores) / len(scores) if scores else 0.0
        
        if overall_quality < self.config.min_image_quality_score:
            issues.append(ReasonCodes.LOW_QUALITY_IMAGE)
        
        return overall_quality, issues
    
    def _check_brightness(self, image: np.ndarray) -> Tuple[float, Optional[ReasonCodes]]:
        """Check image brightness."""
        if len(image.shape) == 3:
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        else:
            gray = image
        
        mean_brightness = gray.mean()
        
        if mean_brightness < 40:
            return 0.3, ReasonCodes.IMAGE_TOO_DARK
        elif mean_brightness > 215:
            return 0.3, ReasonCodes.IMAGE_TOO_BRIGHT
        
        # Normalize to 0-1 score
        if mean_brightness < 100:
            score = mean_brightness / 100
        elif mean_brightness > 155:
            score = (255 - mean_brightness) / 100
        else:
            score = 1.0
        
        return min(1.0, score), None
    
    def _check_blur(self, image: np.ndarray) -> Tuple[float, Optional[ReasonCodes]]:
        """Check image blur using Laplacian variance."""
        if len(image.shape) == 3:
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        else:
            gray = image
        
        # Laplacian variance as blur metric
        laplacian_var = cv2.Laplacian(gray, cv2.CV_64F).var()
        
        # Threshold determined empirically
        if laplacian_var < 100:
            return 0.3, ReasonCodes.IMAGE_BLURRY
        
        # Normalize to 0-1 score
        score = min(1.0, laplacian_var / 500)
        
        return score, None
    
    def _check_resolution(self, image: np.ndarray) -> Tuple[float, Optional[ReasonCodes]]:
        """Check image resolution."""
        height, width = image.shape[:2]
        min_w, min_h = self.preprocessing_config.min_resolution
        
        if width < min_w or height < min_h:
            return 0.3, ReasonCodes.LOW_RESOLUTION
        
        # Score based on how much above minimum
        width_ratio = width / min_w
        height_ratio = height / min_h
        score = min(1.0, (width_ratio + height_ratio) / 4)
        
        return score, None
