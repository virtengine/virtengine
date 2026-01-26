"""
Mask post-processing module.

This module provides post-processing operations for segmentation masks:
- Thresholding (binary and adaptive)
- Morphological operations (erosion, dilation, opening, closing)
- Connected component analysis
- Contour smoothing
"""

import logging
from typing import Tuple, Optional, List
from dataclasses import dataclass

import numpy as np
import cv2

from ml.face_extraction.config import MaskProcessingConfig

logger = logging.getLogger(__name__)


@dataclass
class MaskProcessingResult:
    """Result of mask processing."""
    
    processed_mask: np.ndarray
    original_mask: np.ndarray
    num_components: int
    largest_component_area: int
    operations_applied: List[str]
    success: bool
    error_message: Optional[str] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary for serialization."""
        return {
            "success": self.success,
            "num_components": self.num_components,
            "largest_component_area": self.largest_component_area,
            "operations_applied": self.operations_applied,
            "error_message": self.error_message,
        }


class MaskProcessor:
    """
    Mask post-processing for face segmentation.
    
    This class provides morphological and other operations to clean
    up segmentation masks for reliable face extraction.
    """
    
    def __init__(self, config: Optional[MaskProcessingConfig] = None):
        """
        Initialize the mask processor.
        
        Args:
            config: Mask processing configuration. Uses defaults if not provided.
        """
        self.config = config or MaskProcessingConfig()
    
    def threshold(
        self, 
        mask: np.ndarray, 
        threshold: Optional[float] = None
    ) -> np.ndarray:
        """
        Apply binary threshold to mask.
        
        Args:
            mask: Input mask (H, W) with values in [0, 1]
            threshold: Threshold value (uses config default if not provided)
            
        Returns:
            Binary mask (H, W) with values 0 or 255
        """
        thresh = threshold if threshold is not None else self.config.threshold
        
        # Ensure mask is in correct format
        if mask.dtype != np.float32:
            mask = mask.astype(np.float32)
        
        # Normalize if needed
        if mask.max() > 1.0:
            mask = mask / 255.0
        
        if self.config.use_adaptiVIRTENGINE_threshold:
            # Convert to 8-bit for adaptive thresholding
            mask_8bit = (mask * 255).astype(np.uint8)
            binary = cv2.adaptiveThreshold(
                mask_8bit,
                255,
                cv2.ADAPTIVIRTENGINE_THRESH_GAUSSIAN_C,
                cv2.THRESH_BINARY,
                self.config.adaptiVIRTENGINE_block_size,
                self.config.adaptiVIRTENGINE_constant
            )
        else:
            # Simple binary threshold
            binary = (mask >= thresh).astype(np.uint8) * 255
        
        return binary
    
    def erode(
        self, 
        mask: np.ndarray,
        kernel_size: Optional[Tuple[int, int]] = None,
        iterations: Optional[int] = None
    ) -> np.ndarray:
        """
        Apply erosion to mask.
        
        Args:
            mask: Binary mask (H, W)
            kernel_size: Erosion kernel size
            iterations: Number of iterations
            
        Returns:
            Eroded mask
        """
        ksize = kernel_size or self.config.erosion_kernel_size
        iters = iterations if iterations is not None else self.config.erosion_iterations
        
        kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, ksize)
        return cv2.erode(mask, kernel, iterations=iters)
    
    def dilate(
        self, 
        mask: np.ndarray,
        kernel_size: Optional[Tuple[int, int]] = None,
        iterations: Optional[int] = None
    ) -> np.ndarray:
        """
        Apply dilation to mask.
        
        Args:
            mask: Binary mask (H, W)
            kernel_size: Dilation kernel size
            iterations: Number of iterations
            
        Returns:
            Dilated mask
        """
        ksize = kernel_size or self.config.dilation_kernel_size
        iters = iterations if iterations is not None else self.config.dilation_iterations
        
        kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, ksize)
        return cv2.dilate(mask, kernel, iterations=iters)
    
    def opening(
        self, 
        mask: np.ndarray,
        kernel_size: Optional[Tuple[int, int]] = None
    ) -> np.ndarray:
        """
        Apply morphological opening (erosion then dilation).
        
        Removes small objects from foreground.
        
        Args:
            mask: Binary mask (H, W)
            kernel_size: Kernel size for operation
            
        Returns:
            Opened mask
        """
        ksize = kernel_size or self.config.opening_kernel_size
        kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, ksize)
        return cv2.morphologyEx(mask, cv2.MORPH_OPEN, kernel)
    
    def closing(
        self, 
        mask: np.ndarray,
        kernel_size: Optional[Tuple[int, int]] = None
    ) -> np.ndarray:
        """
        Apply morphological closing (dilation then erosion).
        
        Fills small holes in foreground.
        
        Args:
            mask: Binary mask (H, W)
            kernel_size: Kernel size for operation
            
        Returns:
            Closed mask
        """
        ksize = kernel_size or self.config.closing_kernel_size
        kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, ksize)
        return cv2.morphologyEx(mask, cv2.MORPH_CLOSE, kernel)
    
    def morphological_cleanup(self, mask: np.ndarray) -> np.ndarray:
        """
        Apply standard morphological cleanup sequence.
        
        Sequence: opening -> closing -> erosion -> dilation
        
        Args:
            mask: Binary mask (H, W)
            
        Returns:
            Cleaned mask
        """
        result = mask.copy()
        
        # Opening removes small noise
        if self.config.apply_opening:
            result = self.opening(result)
        
        # Closing fills small holes
        if self.config.apply_closing:
            result = self.closing(result)
        
        # Additional erosion/dilation for fine-tuning
        if self.config.erosion_iterations > 0:
            result = self.erode(result)
        
        if self.config.dilation_iterations > 0:
            result = self.dilate(result)
        
        return result
    
    def largest_connected_component(self, mask: np.ndarray) -> np.ndarray:
        """
        Keep only the largest connected component in the mask.
        
        Args:
            mask: Binary mask (H, W)
            
        Returns:
            Mask with only largest component
        """
        # Ensure mask is 8-bit
        if mask.dtype != np.uint8:
            mask = mask.astype(np.uint8)
        
        # Find connected components
        num_labels, labels, stats, centroids = cv2.connectedComponentsWithStats(
            mask, connectivity=8
        )
        
        if num_labels <= 1:
            # No foreground components
            return mask
        
        # Find largest component (excluding background at label 0)
        largest_label = 1
        largest_area = 0
        
        for i in range(1, num_labels):
            area = stats[i, cv2.CC_STAT_AREA]
            if area > largest_area and area >= self.config.min_component_area:
                largest_area = area
                largest_label = i
        
        # Create mask with only largest component
        result = np.zeros_like(mask)
        result[labels == largest_label] = 255
        
        return result
    
    def get_connected_components(
        self, 
        mask: np.ndarray
    ) -> Tuple[int, np.ndarray, np.ndarray, np.ndarray]:
        """
        Get all connected components with statistics.
        
        Args:
            mask: Binary mask (H, W)
            
        Returns:
            Tuple of (num_labels, labels, stats, centroids)
        """
        if mask.dtype != np.uint8:
            mask = mask.astype(np.uint8)
        
        return cv2.connectedComponentsWithStats(mask, connectivity=8)
    
    def smooth_contours(self, mask: np.ndarray) -> np.ndarray:
        """
        Smooth mask contours using polygon approximation.
        
        Args:
            mask: Binary mask (H, W)
            
        Returns:
            Mask with smoothed contours
        """
        if mask.dtype != np.uint8:
            mask = mask.astype(np.uint8)
        
        # Find contours
        contours, hierarchy = cv2.findContours(
            mask, 
            cv2.RETR_EXTERNAL, 
            cv2.CHAIN_APPROX_SIMPLE
        )
        
        if not contours:
            return mask
        
        # Create new mask with smoothed contours
        result = np.zeros_like(mask)
        
        for contour in contours:
            # Calculate arc length for epsilon
            arc_length = cv2.arcLength(contour, True)
            epsilon = self.config.smoothing_epsilon * arc_length
            
            # Approximate polygon
            approx = cv2.approxPolyDP(contour, epsilon, True)
            
            # Draw filled polygon
            cv2.fillPoly(result, [approx], 255)
        
        return result
    
    def fill_holes(self, mask: np.ndarray) -> np.ndarray:
        """
        Fill holes in the mask.
        
        Args:
            mask: Binary mask (H, W)
            
        Returns:
            Mask with holes filled
        """
        if mask.dtype != np.uint8:
            mask = mask.astype(np.uint8)
        
        # Find contours
        contours, hierarchy = cv2.findContours(
            mask,
            cv2.RETR_EXTERNAL,
            cv2.CHAIN_APPROX_SIMPLE
        )
        
        # Draw filled contours
        result = np.zeros_like(mask)
        cv2.drawContours(result, contours, -1, 255, cv2.FILLED)
        
        return result
    
    def process(self, mask: np.ndarray) -> MaskProcessingResult:
        """
        Apply full processing pipeline to mask.
        
        Args:
            mask: Input mask (H, W) with values in [0, 1]
            
        Returns:
            MaskProcessingResult with processed mask and metadata
        """
        if mask is None or mask.size == 0:
            return MaskProcessingResult(
                processed_mask=np.zeros((1, 1), dtype=np.uint8),
                original_mask=np.zeros((1, 1), dtype=np.float32),
                num_components=0,
                largest_component_area=0,
                operations_applied=[],
                success=False,
                error_message="Invalid input mask"
            )
        
        operations = []
        
        try:
            original = mask.copy()
            
            # Step 1: Threshold
            result = self.threshold(mask)
            operations.append("threshold")
            
            # Step 2: Morphological cleanup
            if self.config.apply_morphology:
                result = self.morphological_cleanup(result)
                operations.append("morphological_cleanup")
            
            # Step 3: Largest connected component
            if self.config.use_largest_component:
                result = self.largest_connected_component(result)
                operations.append("largest_component")
            
            # Step 4: Fill holes
            if self.config.fill_holes:
                result = self.fill_holes(result)
                operations.append("fill_holes")
            
            # Step 5: Smooth contours
            if self.config.smooth_contours:
                result = self.smooth_contours(result)
                operations.append("smooth_contours")
            
            # Get component statistics
            num_labels, labels, stats, _ = self.get_connected_components(result)
            num_components = num_labels - 1  # Exclude background
            
            largest_area = 0
            if num_components > 0:
                for i in range(1, num_labels):
                    area = stats[i, cv2.CC_STAT_AREA]
                    largest_area = max(largest_area, area)
            
            return MaskProcessingResult(
                processed_mask=result,
                original_mask=original,
                num_components=num_components,
                largest_component_area=largest_area,
                operations_applied=operations,
                success=True,
            )
            
        except Exception as e:
            logger.error(f"Mask processing failed: {e}")
            return MaskProcessingResult(
                processed_mask=np.zeros_like(mask, dtype=np.uint8),
                original_mask=mask,
                num_components=0,
                largest_component_area=0,
                operations_applied=operations,
                success=False,
                error_message=str(e)
            )
