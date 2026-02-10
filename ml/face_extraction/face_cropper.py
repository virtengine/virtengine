"""
Face region cropping module.

This module provides face bounding box extraction and cropping
from segmentation masks.
"""

import logging
from dataclasses import dataclass
from typing import Tuple, Optional, List

import numpy as np
import cv2

from ml.face_extraction.config import CropperConfig, InterpolationMethod

logger = logging.getLogger(__name__)


# OpenCV interpolation mapping
CV2_INTERPOLATION = {
    InterpolationMethod.NEAREST: cv2.INTER_NEAREST,
    InterpolationMethod.LINEAR: cv2.INTER_LINEAR,
    InterpolationMethod.CUBIC: cv2.INTER_CUBIC,
    InterpolationMethod.LANCZOS: cv2.INTER_LANCZOS4,
    InterpolationMethod.AREA: cv2.INTER_AREA,
}


@dataclass
class BoundingBox:
    """Bounding box for face region."""
    
    x: int  # Top-left x coordinate
    y: int  # Top-left y coordinate
    width: int
    height: int
    
    @property
    def x1(self) -> int:
        """Left edge."""
        return self.x
    
    @property
    def y1(self) -> int:
        """Top edge."""
        return self.y
    
    @property
    def x2(self) -> int:
        """Right edge."""
        return self.x + self.width
    
    @property
    def y2(self) -> int:
        """Bottom edge."""
        return self.y + self.height
    
    @property
    def center(self) -> Tuple[int, int]:
        """Center point (cx, cy)."""
        return (self.x + self.width // 2, self.y + self.height // 2)
    
    @property
    def area(self) -> int:
        """Box area in pixels."""
        return self.width * self.height
    
    @property
    def aspect_ratio(self) -> float:
        """Width / Height ratio."""
        return self.width / self.height if self.height > 0 else 0.0
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "x": self.x,
            "y": self.y,
            "width": self.width,
            "height": self.height,
            "center": self.center,
            "area": self.area,
            "aspect_ratio": self.aspect_ratio,
        }
    
    def expand(self, margin: float) -> 'BoundingBox':
        """
        Expand bounding box by a margin percentage.
        
        Args:
            margin: Margin as fraction (0.1 = 10%)
            
        Returns:
            New expanded BoundingBox
        """
        margin_x = int(self.width * margin)
        margin_y = int(self.height * margin)
        
        return BoundingBox(
            x=self.x - margin_x,
            y=self.y - margin_y,
            width=self.width + 2 * margin_x,
            height=self.height + 2 * margin_y,
        )
    
    def clip_to_image(self, image_width: int, image_height: int) -> 'BoundingBox':
        """
        Clip bounding box to image boundaries.
        
        Args:
            image_width: Image width
            image_height: Image height
            
        Returns:
            New clipped BoundingBox
        """
        x1 = max(0, self.x)
        y1 = max(0, self.y)
        x2 = min(image_width, self.x2)
        y2 = min(image_height, self.y2)
        
        return BoundingBox(
            x=x1,
            y=y1,
            width=x2 - x1,
            height=y2 - y1,
        )


@dataclass
class FaceExtractionResult:
    """Result of face extraction."""
    
    face_image: Optional[np.ndarray]  # Cropped face (None if data minimization)
    bounding_box: BoundingBox
    mask: np.ndarray  # Segmentation mask
    confidence: float
    model_version: str
    model_hash: str
    success: bool = True
    error_message: Optional[str] = None
    face_percentage: float = 0.0  # Face area as percentage of image
    
    def to_dict(self) -> dict:
        """Convert to dictionary for serialization."""
        return {
            "success": self.success,
            "bounding_box": self.bounding_box.to_dict(),
            "confidence": self.confidence,
            "model_version": self.model_version,
            "model_hash": self.model_hash[:16] + "..." if self.model_hash else None,
            "face_percentage": self.face_percentage,
            "error_message": self.error_message,
            "has_face_image": self.face_image is not None,
        }


class FaceCropper:
    """
    Face region extraction from segmentation masks.
    
    This class extracts bounding boxes from masks and crops
    face regions from document images.
    """
    
    def __init__(self, config: Optional[CropperConfig] = None):
        """
        Initialize the face cropper.
        
        Args:
            config: Cropper configuration. Uses defaults if not provided.
        """
        self.config = config or CropperConfig()
    
    def extract_bounding_box(self, mask: np.ndarray) -> Optional[BoundingBox]:
        """
        Compute bounding box from segmentation mask.
        
        Args:
            mask: Binary mask (H, W)
            
        Returns:
            BoundingBox or None if no face region found
        """
        if mask is None or mask.size == 0:
            return None
        
        # Ensure mask is binary
        if mask.dtype != np.uint8:
            mask = (mask > 0).astype(np.uint8) * 255
        
        # Find contours
        contours, _ = cv2.findContours(
            mask, 
            cv2.RETR_EXTERNAL, 
            cv2.CHAIN_APPROX_SIMPLE
        )
        
        if not contours:
            return None
        
        # Find bounding rect of all contours combined
        # or just the largest contour
        if len(contours) == 1:
            x, y, w, h = cv2.boundingRect(contours[0])
        else:
            # Combine all contours
            all_points = np.vstack(contours)
            x, y, w, h = cv2.boundingRect(all_points)
        
        if w <= 0 or h <= 0:
            return None
        
        return BoundingBox(x=x, y=y, width=w, height=h)
    
    def extract_min_enclosing_rect(
        self, 
        mask: np.ndarray
    ) -> Tuple[Optional[BoundingBox], float]:
        """
        Extract minimum enclosing rectangle (may be rotated).
        
        Args:
            mask: Binary mask (H, W)
            
        Returns:
            Tuple of (BoundingBox, rotation_angle)
        """
        if mask is None or mask.size == 0:
            return None, 0.0
        
        if mask.dtype != np.uint8:
            mask = (mask > 0).astype(np.uint8) * 255
        
        contours, _ = cv2.findContours(
            mask,
            cv2.RETR_EXTERNAL,
            cv2.CHAIN_APPROX_SIMPLE
        )
        
        if not contours:
            return None, 0.0
        
        # Combine all contours
        all_points = np.vstack(contours)
        
        # Get minimum enclosing rotated rectangle
        rect = cv2.minAreaRect(all_points)
        center, size, angle = rect
        
        # Convert to axis-aligned bounding box
        w, h = size
        x = int(center[0] - w / 2)
        y = int(center[1] - h / 2)
        
        return BoundingBox(x=x, y=y, width=int(w), height=int(h)), angle
    
    def _get_margin(self, side: str) -> float:
        """Get margin for a specific side."""
        margin_map = {
            'top': self.config.margin_top,
            'bottom': self.config.margin_bottom,
            'left': self.config.margin_left,
            'right': self.config.margin_right,
        }
        return margin_map.get(side) or self.config.margin
    
    def crop_face(
        self, 
        image: np.ndarray, 
        bbox: BoundingBox,
        margin: Optional[float] = None
    ) -> np.ndarray:
        """
        Crop face region from image.
        
        Args:
            image: Input image (H, W, C)
            bbox: Face bounding box
            margin: Margin as fraction (overrides config if provided)
            
        Returns:
            Cropped face image
        """
        if image is None or image.size == 0:
            raise ValueError("Invalid input image")
        
        if bbox is None:
            raise ValueError("Invalid bounding box")
        
        h, w = image.shape[:2]
        
        # Calculate margins
        if margin is not None:
            margin_top = margin_bottom = margin_left = margin_right = margin
        else:
            margin_top = self._get_margin('top')
            margin_bottom = self._get_margin('bottom')
            margin_left = self._get_margin('left')
            margin_right = self._get_margin('right')
        
        # Apply margins
        margin_x_left = int(bbox.width * margin_left)
        margin_x_right = int(bbox.width * margin_right)
        margin_y_top = int(bbox.height * margin_top)
        margin_y_bottom = int(bbox.height * margin_bottom)
        
        x1 = max(0, bbox.x - margin_x_left)
        y1 = max(0, bbox.y - margin_y_top)
        x2 = min(w, bbox.x2 + margin_x_right)
        y2 = min(h, bbox.y2 + margin_y_bottom)
        
        # Crop
        cropped = image[y1:y2, x1:x2].copy()
        
        # Resize if output size specified
        if self.config.output_size:
            target_h, target_w = self.config.output_size
            
            if self.config.maintain_aspect_ratio:
                cropped = self._resize_with_padding(
                    cropped, 
                    target_w, 
                    target_h
                )
            else:
                interp = CV2_INTERPOLATION.get(
                    self.config.interpolation, 
                    cv2.INTER_LANCZOS4
                )
                cropped = cv2.resize(
                    cropped, 
                    (target_w, target_h), 
                    interpolation=interp
                )
        
        return cropped
    
    def _resize_with_padding(
        self, 
        image: np.ndarray, 
        target_w: int, 
        target_h: int
    ) -> np.ndarray:
        """
        Resize image maintaining aspect ratio with padding.
        
        Args:
            image: Input image
            target_w: Target width
            target_h: Target height
            
        Returns:
            Resized image with padding
        """
        h, w = image.shape[:2]
        
        # Calculate scale
        scale = min(target_w / w, target_h / h)
        new_w = int(w * scale)
        new_h = int(h * scale)
        
        # Resize
        interp = CV2_INTERPOLATION.get(
            self.config.interpolation, 
            cv2.INTER_LANCZOS4
        )
        resized = cv2.resize(image, (new_w, new_h), interpolation=interp)
        
        # Create padded image
        if len(image.shape) == 3:
            result = np.full(
                (target_h, target_w, image.shape[2]), 
                self.config.padding_color, 
                dtype=np.uint8
            )
        else:
            result = np.full(
                (target_h, target_w), 
                self.config.padding_color[0], 
                dtype=np.uint8
            )
        
        # Center the resized image
        x_offset = (target_w - new_w) // 2
        y_offset = (target_h - new_h) // 2
        result[y_offset:y_offset + new_h, x_offset:x_offset + new_w] = resized
        
        return result
    
    def validate_face_region(
        self, 
        bbox: BoundingBox, 
        image_shape: Tuple[int, int]
    ) -> Tuple[bool, str]:
        """
        Validate that face region meets quality criteria.
        
        Args:
            bbox: Face bounding box
            image_shape: Image shape (H, W)
            
        Returns:
            Tuple of (is_valid, reason)
        """
        h, w = image_shape
        
        # Check minimum size
        if bbox.width < self.config.min_face_width:
            return False, f"Face width {bbox.width} below minimum {self.config.min_face_width}"
        
        if bbox.height < self.config.min_face_height:
            return False, f"Face height {bbox.height} below minimum {self.config.min_face_height}"
        
        # Check face percentage of image
        image_area = h * w
        face_percentage = bbox.area / image_area if image_area > 0 else 0
        
        if face_percentage < self.config.min_face_percentage:
            return False, f"Face percentage {face_percentage:.2%} below minimum {self.config.min_face_percentage:.2%}"
        
        if face_percentage > self.config.max_face_percentage:
            return False, f"Face percentage {face_percentage:.2%} above maximum {self.config.max_face_percentage:.2%}"
        
        return True, "Valid"
    
    def extract(
        self, 
        image: np.ndarray, 
        mask: np.ndarray,
        confidence: float = 0.0,
        model_version: str = "",
        model_hash: str = "",
        return_face_image: bool = True
    ) -> FaceExtractionResult:
        """
        Extract face from image using segmentation mask.
        
        Args:
            image: Input document image (H, W, C)
            mask: Segmentation mask (H, W)
            confidence: Segmentation confidence
            model_version: Model version string
            model_hash: Model weights hash
            return_face_image: Whether to include face image in result
            
        Returns:
            FaceExtractionResult with cropped face and metadata
        """
        if image is None or image.size == 0:
            return FaceExtractionResult(
                face_image=None,
                bounding_box=BoundingBox(0, 0, 0, 0),
                mask=np.zeros((1, 1), dtype=np.uint8),
                confidence=0.0,
                model_version=model_version,
                model_hash=model_hash,
                success=False,
                error_message="Invalid input image"
            )
        
        try:
            # Extract bounding box
            bbox = self.extract_bounding_box(mask)
            
            if bbox is None:
                return FaceExtractionResult(
                    face_image=None,
                    bounding_box=BoundingBox(0, 0, 0, 0),
                    mask=mask,
                    confidence=confidence,
                    model_version=model_version,
                    model_hash=model_hash,
                    success=False,
                    error_message="No face region found in mask"
                )
            
            # Validate face region
            is_valid, reason = self.validate_face_region(bbox, image.shape[:2])
            
            if not is_valid:
                return FaceExtractionResult(
                    face_image=None,
                    bounding_box=bbox,
                    mask=mask,
                    confidence=confidence,
                    model_version=model_version,
                    model_hash=model_hash,
                    success=False,
                    error_message=reason
                )
            
            # Crop face
            face_image = None
            if return_face_image:
                face_image = self.crop_face(image, bbox)
            
            # Calculate face percentage
            image_area = image.shape[0] * image.shape[1]
            face_percentage = bbox.area / image_area if image_area > 0 else 0
            
            return FaceExtractionResult(
                face_image=face_image,
                bounding_box=bbox,
                mask=mask,
                confidence=confidence,
                model_version=model_version,
                model_hash=model_hash,
                success=True,
                face_percentage=face_percentage,
            )
            
        except Exception as e:
            logger.error(f"Face extraction failed: {e}")
            return FaceExtractionResult(
                face_image=None,
                bounding_box=BoundingBox(0, 0, 0, 0),
                mask=mask if mask is not None else np.zeros((1, 1), dtype=np.uint8),
                confidence=confidence,
                model_version=model_version,
                model_hash=model_hash,
                success=False,
                error_message=str(e)
            )
