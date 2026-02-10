"""
Perspective correction module.

This module provides perspective correction for ID documents using:
- Edge detection to find document boundaries
- Corner detection using contour approximation
- Four-point perspective transform to correct perspective distortion
"""

import logging
from typing import List, Tuple, Optional
from dataclasses import dataclass

import numpy as np
import cv2

from ml.document_preprocessing.config import DocumentConfig, PerspectiveConfig

logger = logging.getLogger(__name__)


@dataclass
class Point:
    """A 2D point."""
    x: float
    y: float
    
    def as_tuple(self) -> Tuple[float, float]:
        """Convert to tuple."""
        return (self.x, self.y)
    
    def as_int_tuple(self) -> Tuple[int, int]:
        """Convert to integer tuple."""
        return (int(self.x), int(self.y))


@dataclass
class PerspectiveResult:
    """Result of perspective correction."""
    
    corrected: bool  # Whether correction was applied
    corners_detected: bool  # Whether 4 corners were found
    corners: Optional[List[Point]]  # Detected corners
    confidence: float  # Detection confidence


class PerspectiveCorrector:
    """
    Corrects perspective distortion in document images.
    
    This class detects the four corners of a document using edge
    detection and contour analysis, then applies a perspective
    transform to correct any distortion.
    """
    
    def __init__(self, config: Optional[DocumentConfig] = None):
        """
        Initialize the perspective corrector.
        
        Args:
            config: Document configuration. Uses defaults if not provided.
        """
        self.config = config or DocumentConfig()
        self.persp_config = self.config.perspective
    
    def correct_perspective(
        self,
        image: np.ndarray
    ) -> Tuple[np.ndarray, PerspectiveResult]:
        """
        Auto-detect document corners and correct perspective.
        
        Args:
            image: Input image (RGB format)
            
        Returns:
            Tuple of (corrected image, PerspectiveResult)
        """
        if not self.persp_config.enabled:
            return image, PerspectiveResult(
                corrected=False,
                corners_detected=False,
                corners=None,
                confidence=0.0
            )
        
        # Detect corners
        corners, confidence = self._detect_document_corners(image)
        
        if corners is None or len(corners) != 4:
            logger.info("Could not detect document corners, skipping perspective correction")
            return image, PerspectiveResult(
                corrected=False,
                corners_detected=False,
                corners=None,
                confidence=0.0
            )
        
        # Check if correction is needed
        if not self._needs_correction(corners, image.shape):
            logger.info("Document appears flat, skipping perspective correction")
            return image, PerspectiveResult(
                corrected=False,
                corners_detected=True,
                corners=corners,
                confidence=confidence
            )
        
        # Apply perspective transform
        corrected = self.four_point_transform(image, corners)
        
        return corrected, PerspectiveResult(
            corrected=True,
            corners_detected=True,
            corners=corners,
            confidence=confidence
        )
    
    def detect_document_corners(
        self,
        image: np.ndarray
    ) -> Tuple[Optional[List[Point]], float]:
        """
        Detect the four corners of a document in the image.
        
        Args:
            image: Input image (RGB format)
            
        Returns:
            Tuple of (list of 4 corner Points or None, confidence score)
        """
        return self._detect_document_corners(image)
    
    def _detect_document_corners(
        self,
        image: np.ndarray
    ) -> Tuple[Optional[List[Point]], float]:
        """
        Internal corner detection implementation.
        
        Args:
            image: Input image
            
        Returns:
            Tuple of (corners, confidence)
        """
        height, width = image.shape[:2]
        image_area = width * height
        
        # Convert to grayscale
        gray = cv2.cvtColor(image, cv2.COLOR_RGB2GRAY)
        
        # Apply blur to reduce noise
        blurred = cv2.GaussianBlur(
            gray,
            (self.persp_config.blur_kernel_size, self.persp_config.blur_kernel_size),
            0
        )
        
        # Edge detection
        edges = cv2.Canny(
            blurred,
            self.persp_config.canny_low_threshold,
            self.persp_config.canny_high_threshold
        )
        
        # Apply morphological operations to close gaps
        if self.persp_config.apply_morphology:
            kernel = cv2.getStructuringElement(
                cv2.MORPH_RECT,
                (self.persp_config.morphology_kernel_size,
                 self.persp_config.morphology_kernel_size)
            )
            edges = cv2.dilate(edges, kernel, iterations=1)
            edges = cv2.erode(edges, kernel, iterations=1)
        
        # Find contours
        contours, _ = cv2.findContours(
            edges,
            cv2.RETR_EXTERNAL,
            cv2.CHAIN_APPROX_SIMPLE
        )
        
        if not contours:
            return None, 0.0
        
        # Find the largest contour that could be a document
        min_area = image_area * self.persp_config.min_contour_area_ratio
        max_area = image_area * self.persp_config.max_contour_area_ratio
        
        best_corners = None
        best_confidence = 0.0
        
        for contour in contours:
            area = cv2.contourArea(contour)
            
            if area < min_area or area > max_area:
                continue
            
            # Approximate the contour to a polygon
            peri = cv2.arcLength(contour, True)
            epsilon = self.persp_config.corner_epsilon_ratio * peri
            approx = cv2.approxPolyDP(contour, epsilon, True)
            
            # We want exactly 4 corners (quadrilateral)
            if len(approx) == 4:
                # Check if it's convex
                if cv2.isContourConvex(approx):
                    # Calculate confidence based on area
                    confidence = area / image_area
                    
                    if confidence > best_confidence:
                        corners = [
                            Point(float(pt[0][0]), float(pt[0][1]))
                            for pt in approx
                        ]
                        # Order corners: top-left, top-right, bottom-right, bottom-left
                        corners = self._order_corners(corners)
                        best_corners = corners
                        best_confidence = confidence
        
        return best_corners, best_confidence
    
    def _order_corners(self, corners: List[Point]) -> List[Point]:
        """
        Order corners as: top-left, top-right, bottom-right, bottom-left.
        
        Args:
            corners: Unordered list of 4 corner points
            
        Returns:
            Ordered list of corners
        """
        # Convert to numpy array for easier manipulation
        pts = np.array([(p.x, p.y) for p in corners], dtype=np.float32)
        
        # Sum of coordinates: top-left has smallest, bottom-right has largest
        s = pts.sum(axis=1)
        top_left_idx = np.argmin(s)
        bottom_right_idx = np.argmax(s)
        
        # Difference of coordinates: top-right has smallest, bottom-left has largest
        d = np.diff(pts, axis=1)
        top_right_idx = np.argmin(d)
        bottom_left_idx = np.argmax(d)
        
        ordered = [
            Point(pts[top_left_idx][0], pts[top_left_idx][1]),
            Point(pts[top_right_idx][0], pts[top_right_idx][1]),
            Point(pts[bottom_right_idx][0], pts[bottom_right_idx][1]),
            Point(pts[bottom_left_idx][0], pts[bottom_left_idx][1]),
        ]
        
        return ordered
    
    def _needs_correction(
        self,
        corners: List[Point],
        image_shape: Tuple[int, ...]
    ) -> bool:
        """
        Check if perspective correction is needed.
        
        If the document corners are already close to rectangular,
        skip correction to preserve quality.
        
        Args:
            corners: Detected corners
            image_shape: Image shape (height, width, ...)
            
        Returns:
            True if correction is needed
        """
        height, width = image_shape[:2]
        
        # Check if corners form a reasonable rectangle
        # by comparing side lengths
        pts = np.array([c.as_tuple() for c in corners], dtype=np.float32)
        
        # Calculate side lengths
        top = np.linalg.norm(pts[1] - pts[0])
        right = np.linalg.norm(pts[2] - pts[1])
        bottom = np.linalg.norm(pts[3] - pts[2])
        left = np.linalg.norm(pts[0] - pts[3])
        
        # Calculate ratios
        width_ratio = min(top, bottom) / max(top, bottom)
        height_ratio = min(left, right) / max(left, right)
        
        # Check angles (should be close to 90 degrees)
        def angle_at_corner(p1, corner, p2):
            v1 = np.array([p1[0] - corner[0], p1[1] - corner[1]])
            v2 = np.array([p2[0] - corner[0], p2[1] - corner[1]])
            cos_angle = np.dot(v1, v2) / (np.linalg.norm(v1) * np.linalg.norm(v2) + 1e-6)
            return np.degrees(np.arccos(np.clip(cos_angle, -1, 1)))
        
        angles = [
            angle_at_corner(pts[3], pts[0], pts[1]),
            angle_at_corner(pts[0], pts[1], pts[2]),
            angle_at_corner(pts[1], pts[2], pts[3]),
            angle_at_corner(pts[2], pts[3], pts[0]),
        ]
        
        # Check how close angles are to 90 degrees
        angle_deviation = max(abs(a - 90) for a in angles)
        
        # If document is reasonably rectangular, skip correction
        if width_ratio > 0.95 and height_ratio > 0.95 and angle_deviation < 5:
            return False
        
        return True
    
    def four_point_transform(
        self,
        image: np.ndarray,
        corners: List[Point]
    ) -> np.ndarray:
        """
        Apply four-point perspective transform.
        
        Args:
            image: Input image
            corners: Ordered list of 4 corners (TL, TR, BR, BL)
            
        Returns:
            Perspective-corrected image
        """
        pts = np.array([c.as_tuple() for c in corners], dtype=np.float32)
        
        # Compute the width of the new image
        width_top = np.linalg.norm(pts[1] - pts[0])
        width_bottom = np.linalg.norm(pts[2] - pts[3])
        max_width = int(max(width_top, width_bottom))
        
        # Compute the height of the new image
        height_left = np.linalg.norm(pts[3] - pts[0])
        height_right = np.linalg.norm(pts[2] - pts[1])
        max_height = int(max(height_left, height_right))
        
        # Add margin
        margin = self.persp_config.output_margin
        
        # Destination points (rectangle)
        dst = np.array([
            [margin, margin],
            [max_width - 1 + margin, margin],
            [max_width - 1 + margin, max_height - 1 + margin],
            [margin, max_height - 1 + margin]
        ], dtype=np.float32)
        
        # Compute perspective transform matrix
        matrix = cv2.getPerspectiveTransform(pts, dst)
        
        # Apply transform
        output_width = max_width + 2 * margin
        output_height = max_height + 2 * margin
        
        warped = cv2.warpPerspective(
            image,
            matrix,
            (output_width, output_height),
            flags=cv2.INTER_CUBIC,
            borderMode=cv2.BORDER_CONSTANT,
            borderValue=(255, 255, 255)
        )
        
        logger.info(f"Applied perspective correction to {output_width}x{output_height}")
        
        return warped
    
    def manual_correction(
        self,
        image: np.ndarray,
        corners: List[Tuple[float, float]]
    ) -> np.ndarray:
        """
        Apply perspective correction with manually specified corners.
        
        Args:
            image: Input image
            corners: List of 4 corner tuples (x, y) in order: TL, TR, BR, BL
            
        Returns:
            Perspective-corrected image
        """
        corner_points = [Point(x, y) for x, y in corners]
        return self.four_point_transform(image, corner_points)
