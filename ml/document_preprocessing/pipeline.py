"""
Main document preprocessing pipeline.

This module orchestrates the complete document preprocessing workflow:
1. Standardize format and resolution
2. Detect and correct orientation
3. Detect and correct perspective
4. Apply CLAHE enhancement
5. Apply noise reduction
6. Apply sharpening
"""

import logging
import time
import hashlib
from dataclasses import dataclass, field
from typing import Tuple, List, Optional

import numpy as np

from ml.document_preprocessing.config import DocumentConfig
from ml.document_preprocessing.standardization import DocumentStandardizer
from ml.document_preprocessing.enhancement import DocumentEnhancer
from ml.document_preprocessing.noise_reduction import NoiseReducer
from ml.document_preprocessing.orientation import OrientationDetector, OrientationResult
from ml.document_preprocessing.perspective import PerspectiveCorrector, PerspectiveResult

logger = logging.getLogger(__name__)


@dataclass
class PreprocessingResult:
    """Result of document preprocessing."""
    
    # Output image
    normalized_image: np.ndarray
    
    # Original image info
    original_size: Tuple[int, int]  # (width, height)
    
    # Transformations applied
    rotation_applied: int  # Degrees (0, 90, 180, 270)
    perspective_corrected: bool
    enhancements_applied: List[str]
    
    # Timing
    processing_time_ms: float
    
    # Additional details
    success: bool = True
    error_message: Optional[str] = None
    orientation_result: Optional[OrientationResult] = None
    perspective_result: Optional[PerspectiveResult] = None
    final_size: Tuple[int, int] = (0, 0)  # (width, height)
    image_hash: Optional[str] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary for serialization."""
        return {
            "success": self.success,
            "original_size": self.original_size,
            "final_size": self.final_size,
            "rotation_applied": self.rotation_applied,
            "perspective_corrected": self.perspective_corrected,
            "enhancements_applied": self.enhancements_applied,
            "processing_time_ms": self.processing_time_ms,
            "error_message": self.error_message,
            "image_hash": self.image_hash,
        }


class DocumentPreprocessingPipeline:
    """
    Complete document preprocessing pipeline.
    
    This class orchestrates all preprocessing steps to produce a
    normalized document image suitable for OCR and face extraction.
    
    Processing Steps:
    1. Standardize format/resolution - Convert to RGB, resize to target
    2. Orientation detection/correction - Detect and fix 90° rotations
    3. Perspective correction - Detect corners and correct distortion
    4. CLAHE enhancement - Improve local contrast
    5. Noise reduction - Remove noise while preserving edges
    6. Sharpening - Enhance text and edge clarity
    """
    
    def __init__(self, config: Optional[DocumentConfig] = None):
        """
        Initialize the preprocessing pipeline.
        
        Args:
            config: Document configuration. Uses defaults if not provided.
        """
        self.config = config or DocumentConfig()
        
        # Initialize components
        self.standardizer = DocumentStandardizer(config)
        self.enhancer = DocumentEnhancer(config)
        self.noise_reducer = NoiseReducer(config)
        self.orientation_detector = OrientationDetector(config)
        self.perspective_corrector = PerspectiveCorrector(config)
    
    def process(
        self,
        image: np.ndarray,
        compute_hash: bool = True
    ) -> PreprocessingResult:
        """
        Apply complete preprocessing pipeline to a document image.
        
        This is the main entry point for document preprocessing.
        
        Args:
            image: Input image as numpy array (BGR or RGB format)
            compute_hash: Whether to compute hash of the output image
            
        Returns:
            PreprocessingResult with normalized image and metadata
        """
        start_time = time.perf_counter()
        
        # Track applied enhancements
        enhancements_applied = []
        rotation_applied = 0
        perspective_corrected = False
        orientation_result = None
        perspective_result = None
        
        # Validate input
        if image is None:
            return self._error_result(
                "Input image is None",
                start_time
            )
        
        if not isinstance(image, np.ndarray):
            return self._error_result(
                f"Expected numpy array, got {type(image)}",
                start_time
            )
        
        if image.size == 0:
            return self._error_result(
                "Input image is empty",
                start_time
            )
        
        original_size = (image.shape[1], image.shape[0])  # (width, height)
        
        try:
            # Step 1: Standardize format and initial resize
            logger.debug("Step 1: Standardizing format and resolution")
            image = self.standardizer.standardize(image)
            enhancements_applied.append("standardize")
            
            # Step 2: Orientation detection and correction
            if self.config.correct_orientation_first:
                logger.debug("Step 2: Detecting and correcting orientation")
                image, rotation_applied = self.orientation_detector.correct_orientation(image)
                if rotation_applied != 0:
                    enhancements_applied.append(f"rotate_{rotation_applied}")
                orientation_result = self.orientation_detector.detect_orientation(image)
            
            # Step 3: Perspective correction
            if self.config.correct_perspective_first:
                logger.debug("Step 3: Detecting and correcting perspective")
                image, perspective_result = self.perspective_corrector.correct_perspective(image)
                perspective_corrected = perspective_result.corrected
                if perspective_corrected:
                    enhancements_applied.append("perspective_correction")
                    # Re-standardize after perspective correction (size may have changed)
                    image = self.standardizer.standardize(image)
            
            # Step 4: Apply CLAHE enhancement
            logger.debug("Step 4: Applying image enhancements")
            image, enhancement_list = self.enhancer.enhance(image)
            enhancements_applied.extend(enhancement_list)
            
            # Step 5: Noise reduction
            logger.debug("Step 5: Applying noise reduction")
            if self.config.noise_reduction.enabled:
                image = self.noise_reducer.denoise(image)
                enhancements_applied.append("noise_reduction")
            
            # Compute hash for determinism verification
            image_hash = None
            if compute_hash:
                image_hash = self._compute_image_hash(image)
            
            # Calculate processing time
            processing_time_ms = (time.perf_counter() - start_time) * 1000
            
            final_size = (image.shape[1], image.shape[0])
            
            logger.info(
                f"Preprocessing complete: {original_size} -> {final_size}, "
                f"rotation={rotation_applied}°, perspective={perspective_corrected}, "
                f"time={processing_time_ms:.2f}ms"
            )
            
            return PreprocessingResult(
                normalized_image=image,
                original_size=original_size,
                rotation_applied=rotation_applied,
                perspective_corrected=perspective_corrected,
                enhancements_applied=enhancements_applied,
                processing_time_ms=processing_time_ms,
                success=True,
                orientation_result=orientation_result,
                perspective_result=perspective_result,
                final_size=final_size,
                image_hash=image_hash,
            )
            
        except ValueError as e:
            return self._error_result(str(e), start_time)
        except Exception as e:
            logger.exception("Preprocessing failed with unexpected error")
            return self._error_result(f"Unexpected error: {e}", start_time)
    
    def process_minimal(self, image: np.ndarray) -> np.ndarray:
        """
        Apply minimal preprocessing (just standardization).
        
        Useful when you want just format/size standardization without
        other enhancements.
        
        Args:
            image: Input image
            
        Returns:
            Standardized image
        """
        return self.standardizer.standardize(image)
    
    def process_with_steps(
        self,
        image: np.ndarray,
        steps: List[str]
    ) -> PreprocessingResult:
        """
        Apply only specified preprocessing steps.
        
        Available steps:
        - "standardize": Format and resolution standardization
        - "orientation": Orientation detection and correction
        - "perspective": Perspective correction
        - "clahe": CLAHE histogram equalization
        - "brightness_contrast": Brightness and contrast normalization
        - "sharpening": Unsharp masking
        - "noise_reduction": Noise reduction (bilateral)
        
        Args:
            image: Input image
            steps: List of step names to apply
            
        Returns:
            PreprocessingResult with processed image
        """
        start_time = time.perf_counter()
        
        original_size = (image.shape[1], image.shape[0])
        enhancements_applied = []
        rotation_applied = 0
        perspective_corrected = False
        
        try:
            if "standardize" in steps:
                image = self.standardizer.standardize(image)
                enhancements_applied.append("standardize")
            
            if "orientation" in steps:
                image, rotation_applied = self.orientation_detector.correct_orientation(image)
                if rotation_applied != 0:
                    enhancements_applied.append(f"rotate_{rotation_applied}")
            
            if "perspective" in steps:
                image, result = self.perspective_corrector.correct_perspective(image)
                perspective_corrected = result.corrected
                if perspective_corrected:
                    enhancements_applied.append("perspective_correction")
            
            if "clahe" in steps:
                image = self.enhancer.apply_clahe(image)
                enhancements_applied.append("clahe")
            
            if "brightness_contrast" in steps:
                image = self.enhancer.adjust_brightness_contrast(image)
                enhancements_applied.append("brightness_contrast")
            
            if "sharpening" in steps:
                image = self.enhancer.sharpen(image)
                enhancements_applied.append("sharpening")
            
            if "noise_reduction" in steps:
                image = self.noise_reducer.denoise(image)
                enhancements_applied.append("noise_reduction")
            
            processing_time_ms = (time.perf_counter() - start_time) * 1000
            
            return PreprocessingResult(
                normalized_image=image,
                original_size=original_size,
                rotation_applied=rotation_applied,
                perspective_corrected=perspective_corrected,
                enhancements_applied=enhancements_applied,
                processing_time_ms=processing_time_ms,
                success=True,
                final_size=(image.shape[1], image.shape[0]),
            )
            
        except Exception as e:
            return self._error_result(str(e), start_time)
    
    def _error_result(
        self,
        message: str,
        start_time: float
    ) -> PreprocessingResult:
        """
        Create an error result.
        
        Args:
            message: Error message
            start_time: Processing start time
            
        Returns:
            PreprocessingResult with error
        """
        processing_time_ms = (time.perf_counter() - start_time) * 1000
        
        logger.error(f"Preprocessing failed: {message}")
        
        return PreprocessingResult(
            normalized_image=np.array([]),
            original_size=(0, 0),
            rotation_applied=0,
            perspective_corrected=False,
            enhancements_applied=[],
            processing_time_ms=processing_time_ms,
            success=False,
            error_message=message,
        )
    
    def _compute_image_hash(self, image: np.ndarray) -> str:
        """
        Compute SHA256 hash of image for determinism verification.
        
        Args:
            image: Image to hash
            
        Returns:
            Hex-encoded SHA256 hash
        """
        # Ensure consistent byte representation
        image_bytes = image.astype(np.uint8).tobytes()
        return hashlib.sha256(image_bytes).hexdigest()
