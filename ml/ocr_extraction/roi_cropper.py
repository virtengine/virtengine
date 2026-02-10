"""
ROI Cropper for precise text region extraction.

This module provides sub-pixel accurate cropping of text regions
from document images, with preprocessing optimized for OCR.
"""

import cv2
import numpy as np
from typing import List, Tuple, Optional
from dataclasses import dataclass
import math

from ml.text_detection import TextROI, BoundingBox, Point
from ml.ocr_extraction.config import CropperConfig, ThresholdingMethod


@dataclass
class CropResult:
    """Result of cropping a single ROI."""
    roi_id: str
    original_crop: np.ndarray  # Original cropped image
    processed_crop: np.ndarray  # Preprocessed for OCR
    applied_rotation: float  # Rotation applied (degrees)
    scale_factor: float  # Scale factor applied
    success: bool = True
    error_message: Optional[str] = None


class ROICropper:
    """
    Precise ROI cropping with sub-pixel accuracy.
    
    Handles rotated text regions and applies preprocessing
    optimized for Tesseract OCR.
    """
    
    def __init__(self, config: Optional[CropperConfig] = None):
        """
        Initialize ROI cropper.
        
        Args:
            config: Cropper configuration. Uses defaults if None.
        """
        self.config = config or CropperConfig()
    
    def crop_roi(self, image: np.ndarray, roi: TextROI) -> np.ndarray:
        """
        Crop a text ROI from the image with sub-pixel accuracy.
        
        For rotated text regions, uses the polygon points to compute
        an optimal rotation and crops with perspective transform.
        
        Args:
            image: Input image (BGR or grayscale)
            roi: Text ROI with bounding box and polygon
            
        Returns:
            Cropped image region
        """
        if len(roi.polygon) == 4 and self.config.use_subpixel_crop:
            return self._crop_rotated_roi(image, roi)
        else:
            return self._crop_axis_aligned(image, roi)
    
    def _crop_axis_aligned(self, image: np.ndarray, roi: TextROI) -> np.ndarray:
        """Crop using axis-aligned bounding box."""
        h, w = image.shape[:2]
        bbox = roi.bounding_box
        
        # Calculate margin
        margin_x = max(
            self.config.margin_pixels,
            int(bbox.width * self.config.margin_fraction)
        )
        margin_y = max(
            self.config.margin_pixels,
            int(bbox.height * self.config.margin_fraction)
        )
        
        # Apply margin with bounds checking
        x1 = max(0, int(bbox.x) - margin_x)
        y1 = max(0, int(bbox.y) - margin_y)
        x2 = min(w, int(bbox.x2) + margin_x)
        y2 = min(h, int(bbox.y2) + margin_y)
        
        return image[y1:y2, x1:x2].copy()
    
    def _crop_rotated_roi(self, image: np.ndarray, roi: TextROI) -> np.ndarray:
        """
        Crop rotated ROI using perspective transform.
        
        Uses cv2.getRectSubPix equivalent via perspective warp
        for sub-pixel accuracy.
        """
        # Get polygon points
        pts = np.array([[p.x, p.y] for p in roi.polygon], dtype=np.float32)
        
        # Compute rotation angle from polygon
        angle = self._compute_rotation_angle(pts)
        
        # Get rotated rectangle
        rect = cv2.minAreaRect(pts)
        center, (width, height), rect_angle = rect
        
        # Ensure width > height for text
        if height > width:
            width, height = height, width
            angle += 90
        
        # Add margin
        margin_x = max(
            self.config.margin_pixels,
            int(width * self.config.margin_fraction)
        )
        margin_y = max(
            self.config.margin_pixels,
            int(height * self.config.margin_fraction)
        )
        
        width_with_margin = int(width) + 2 * margin_x
        height_with_margin = int(height) + 2 * margin_y
        
        # Destination points (axis-aligned rectangle)
        dst_pts = np.array([
            [margin_x, margin_y],
            [width_with_margin - margin_x, margin_y],
            [width_with_margin - margin_x, height_with_margin - margin_y],
            [margin_x, height_with_margin - margin_y]
        ], dtype=np.float32)
        
        # Order source points correctly
        pts_ordered = self._order_points(pts)
        
        # Compute perspective transform
        M = cv2.getPerspectiveTransform(pts_ordered, dst_pts)
        
        # Apply transform
        cropped = cv2.warpPerspective(
            image,
            M,
            (width_with_margin, height_with_margin),
            flags=cv2.INTER_LINEAR,
            borderMode=cv2.BORDER_REPLICATE
        )
        
        return cropped
    
    def _compute_rotation_angle(self, pts: np.ndarray) -> float:
        """Compute rotation angle from polygon points."""
        # Use the top edge to compute angle
        if len(pts) >= 2:
            dx = pts[1][0] - pts[0][0]
            dy = pts[1][1] - pts[0][1]
            angle = math.atan2(dy, dx) * 180 / math.pi
            return angle
        return 0.0
    
    def _order_points(self, pts: np.ndarray) -> np.ndarray:
        """
        Order points in clockwise order: top-left, top-right, 
        bottom-right, bottom-left.
        """
        # Sort by y-coordinate
        sorted_by_y = pts[np.argsort(pts[:, 1])]
        
        # Get top two and bottom two points
        top_pts = sorted_by_y[:2]
        bottom_pts = sorted_by_y[2:]
        
        # Sort by x-coordinate
        top_left, top_right = top_pts[np.argsort(top_pts[:, 0])]
        bottom_left, bottom_right = bottom_pts[np.argsort(bottom_pts[:, 0])]
        
        return np.array([top_left, top_right, bottom_right, bottom_left], dtype=np.float32)
    
    def prepare_for_ocr(self, crop: np.ndarray) -> np.ndarray:
        """
        Prepare cropped image for OCR.
        
        Applies:
        - Grayscale conversion
        - Optional deskewing
        - Thresholding
        - Scaling
        
        Args:
            crop: Cropped image region
            
        Returns:
            Preprocessed image optimized for OCR
        """
        if crop is None or crop.size == 0:
            raise ValueError("Empty crop provided")
        
        result = crop.copy()
        
        # Convert to grayscale if needed
        if self.config.convert_grayscale and len(result.shape) == 3:
            result = cv2.cvtColor(result, cv2.COLOR_BGR2GRAY)
        
        # Deskew if enabled
        if self.config.deskew_enabled:
            result = self._deskew(result)
        
        # Scale to target height if specified
        if self.config.scale_to_height is not None:
            result = self._scale_to_height(result, self.config.scale_to_height)
        
        # Apply thresholding
        result = self._apply_threshold(result)
        
        return result
    
    def _deskew(self, image: np.ndarray) -> np.ndarray:
        """Deskew image by detecting text angle."""
        # Ensure grayscale
        if len(image.shape) == 3:
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        else:
            gray = image
        
        # Invert for better edge detection on white background
        inverted = cv2.bitwise_not(gray)
        
        # Apply threshold
        _, thresh = cv2.threshold(inverted, 0, 255, cv2.THRESH_BINARY + cv2.THRESH_OTSU)
        
        # Find coordinates of non-zero pixels
        coords = np.column_stack(np.where(thresh > 0))
        
        if len(coords) < 10:
            return image
        
        # Compute minimum rotated rectangle
        rect = cv2.minAreaRect(coords)
        angle = rect[2]
        
        # Adjust angle
        if angle < -45:
            angle = 90 + angle
        elif angle > 45:
            angle = angle - 90
        
        # Only correct small angles
        if abs(angle) > self.config.deskew_max_angle:
            return image
        
        # Rotate
        h, w = image.shape[:2]
        center = (w // 2, h // 2)
        M = cv2.getRotationMatrix2D(center, angle, 1.0)
        rotated = cv2.warpAffine(
            image,
            M,
            (w, h),
            flags=cv2.INTER_LINEAR,
            borderMode=cv2.BORDER_REPLICATE
        )
        
        return rotated
    
    def _scale_to_height(self, image: np.ndarray, target_height: int) -> np.ndarray:
        """Scale image to target height maintaining aspect ratio."""
        h, w = image.shape[:2]
        
        if h == target_height:
            return image
        
        scale = target_height / h
        new_width = int(w * scale)
        
        # Limit width
        if new_width > self.config.max_crop_width:
            new_width = self.config.max_crop_width
            scale = new_width / w
            target_height = int(h * scale)
        
        return cv2.resize(
            image,
            (new_width, target_height),
            interpolation=cv2.INTER_LINEAR if scale > 1 else cv2.INTER_AREA
        )
    
    def _apply_threshold(self, image: np.ndarray) -> np.ndarray:
        """Apply thresholding for OCR."""
        method = self.config.thresholding
        
        if method == ThresholdingMethod.NONE:
            return image
        
        # Ensure grayscale
        if len(image.shape) == 3:
            gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        else:
            gray = image
        
        if method == ThresholdingMethod.BINARY:
            _, result = cv2.threshold(
                gray,
                self.config.threshold_value,
                255,
                cv2.THRESH_BINARY
            )
        elif method == ThresholdingMethod.BINARY_INV:
            _, result = cv2.threshold(
                gray,
                self.config.threshold_value,
                255,
                cv2.THRESH_BINARY_INV
            )
        elif method == ThresholdingMethod.OTSU:
            _, result = cv2.threshold(
                gray,
                0,
                255,
                cv2.THRESH_BINARY + cv2.THRESH_OTSU
            )
        elif method == ThresholdingMethod.ADAPTIVE_MEAN:
            result = cv2.adaptiveThreshold(
                gray,
                255,
                cv2.ADAPTIVE_THRESH_MEAN_C,
                cv2.THRESH_BINARY,
                self.config.adaptive_block_size,
                self.config.adaptive_c
            )
        elif method == ThresholdingMethod.ADAPTIVE_GAUSSIAN:
            result = cv2.adaptiveThreshold(
                gray,
                255,
                cv2.ADAPTIVE_THRESH_GAUSSIAN_C,
                cv2.THRESH_BINARY,
                self.config.adaptive_block_size,
                self.config.adaptive_c
            )
        else:
            result = gray
        
        return result
    
    def crop_and_prepare(
        self,
        image: np.ndarray,
        roi: TextROI
    ) -> CropResult:
        """
        Crop ROI and prepare for OCR in one operation.
        
        Args:
            image: Input image
            roi: Text ROI to crop
            
        Returns:
            CropResult with original and processed crops
        """
        try:
            original = self.crop_roi(image, roi)
            
            # Validate crop size
            h, w = original.shape[:2]
            if w < self.config.min_crop_width or h < self.config.min_crop_height:
                return CropResult(
                    roi_id=roi.roi_id,
                    original_crop=original,
                    processed_crop=original,
                    applied_rotation=0.0,
                    scale_factor=1.0,
                    success=False,
                    error_message=f"Crop too small: {w}x{h}"
                )
            
            processed = self.prepare_for_ocr(original)
            
            # Calculate scale factor
            scale_factor = processed.shape[0] / original.shape[0] if original.shape[0] > 0 else 1.0
            
            return CropResult(
                roi_id=roi.roi_id,
                original_crop=original,
                processed_crop=processed,
                applied_rotation=0.0,  # Could be computed from deskew
                scale_factor=scale_factor,
                success=True
            )
            
        except Exception as e:
            # Return empty arrays on error
            empty = np.zeros((1, 1), dtype=np.uint8)
            return CropResult(
                roi_id=roi.roi_id,
                original_crop=empty,
                processed_crop=empty,
                applied_rotation=0.0,
                scale_factor=1.0,
                success=False,
                error_message=str(e)
            )
    
    def crop_all_rois(
        self,
        image: np.ndarray,
        rois: List[TextROI]
    ) -> List[CropResult]:
        """
        Crop all ROIs from an image.
        
        Args:
            image: Input image
            rois: List of TextROIs to crop
            
        Returns:
            List of CropResults
        """
        return [self.crop_and_prepare(image, roi) for roi in rois]
