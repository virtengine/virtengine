"""
PCA-based skew detection and correction module.

This module implements Principal Component Analysis (PCA) for detecting
and correcting document skew. The algorithm:
1. Extracts edge or text contours from the document
2. Applies PCA to find the dominant orientation
3. Calculates the skew angle from principal components
4. Rotates the image to correct the skew

Based on the RMIT intern project PCA implementation.
"""

import logging
from dataclasses import dataclass
from math import atan2, cos, sin, sqrt, pi
from typing import Optional, Tuple, List, Union

import cv2
import numpy as np

logger = logging.getLogger(__name__)


@dataclass
class PCASkewConfig:
    """Configuration for PCA-based skew correction."""
    
    # Enable/disable skew correction
    enabled: bool = True
    
    # Minimum angle (in degrees) to consider for correction
    # Angles below this threshold will be ignored
    min_angle_threshold: float = 0.5
    
    # Maximum angle (in degrees) to correct
    # Angles above this are likely 90-degree rotations, not skew
    max_angle_threshold: float = 45.0
    
    # Edge detection settings
    canny_low_threshold: int = 50
    canny_high_threshold: int = 150
    
    # Morphological operations for contour extraction
    dilation_kernel_size: int = 3
    dilation_iterations: int = 2
    erosion_kernel_size: int = 3
    erosion_iterations: int = 1
    
    # Minimum contour area as percentage of image area
    min_contour_area_ratio: float = 0.01
    
    # Use text-based contours (True) or edge-based (False)
    use_text_contours: bool = True
    
    # Expand image canvas when rotating to prevent cropping
    expand_canvas: bool = True
    
    # Background color for expanded regions (grayscale or BGR)
    background_color: Union[int, Tuple[int, int, int]] = 255


@dataclass
class PCASkewResult:
    """Result of PCA-based skew detection."""
    
    # Detected skew angle in degrees (positive = counter-clockwise)
    skew_angle: float
    
    # Whether the skew was actually corrected
    corrected: bool
    
    # Reason if not corrected
    reason: str
    
    # PCA analysis details
    center: Optional[Tuple[int, int]] = None
    eigenvectors: Optional[np.ndarray] = None
    eigenvalues: Optional[np.ndarray] = None
    
    # Confidence score (0.0 to 1.0)
    confidence: float = 0.0


class PCASkewCorrector:
    """
    PCA-based document skew detection and correction.
    
    This class uses Principal Component Analysis to detect the dominant
    orientation of a document and correct any skew by rotating the image.
    
    The algorithm works best on documents with clear text lines or edges
    that define a dominant orientation.
    
    Example:
        >>> from ml.document_preprocessing.pca_skew import PCASkewCorrector
        >>> corrector = PCASkewCorrector()
        >>> corrected_image, result = corrector.correct_skew(document_image)
        >>> print(f"Corrected skew of {result.skew_angle:.2f} degrees")
    """
    
    def __init__(self, config: Optional[PCASkewConfig] = None):
        """
        Initialize the PCA skew corrector.
        
        Args:
            config: PCA skew configuration. Uses defaults if not provided.
        """
        self.config = config or PCASkewConfig()
    
    def detect_skew_angle(self, image: np.ndarray) -> PCASkewResult:
        """
        Detect the skew angle of a document using PCA.
        
        Args:
            image: Input image (grayscale or BGR/RGB color)
            
        Returns:
            PCASkewResult containing the detected skew angle and analysis details
        """
        if not self.config.enabled:
            return PCASkewResult(
                skew_angle=0.0,
                corrected=False,
                reason="PCA skew detection disabled",
                confidence=0.0
            )
        
        # Convert to grayscale if needed
        gray = self._to_grayscale(image)
        
        # Extract contours for PCA analysis
        contours = self._extract_contours(gray)
        
        if not contours:
            logger.warning("No contours found for PCA analysis")
            return PCASkewResult(
                skew_angle=0.0,
                corrected=False,
                reason="No contours found",
                confidence=0.0
            )
        
        # Find the largest contour(s) for analysis
        min_area = gray.shape[0] * gray.shape[1] * self.config.min_contour_area_ratio
        valid_contours = [c for c in contours if cv2.contourArea(c) >= min_area]
        
        if not valid_contours:
            logger.warning("No valid contours above minimum area threshold")
            return PCASkewResult(
                skew_angle=0.0,
                corrected=False,
                reason="No contours above area threshold",
                confidence=0.0
            )
        
        # Combine all valid contours for global orientation
        all_points = np.vstack(valid_contours)
        
        # Perform PCA analysis
        angle, center, eigenvectors, eigenvalues = self._compute_pca_orientation(all_points)
        
        # Calculate confidence based on eigenvalue ratio
        if eigenvalues is not None and len(eigenvalues) >= 2:
            # Higher ratio = more linear structure = more confident
            ratio = eigenvalues[0] / max(eigenvalues[1], 1e-10)
            confidence = min(1.0, ratio / 10.0)  # Scale to 0-1 range
        else:
            confidence = 0.5
        
        logger.debug(
            f"PCA skew detection: angle={angle:.2f}°, "
            f"confidence={confidence:.2f}, center={center}"
        )
        
        # Determine if angle is within correctable range
        abs_angle = abs(angle)
        
        if abs_angle < self.config.min_angle_threshold:
            return PCASkewResult(
                skew_angle=angle,
                corrected=False,
                reason=f"Angle {angle:.2f}° below threshold {self.config.min_angle_threshold}°",
                center=center,
                eigenvectors=eigenvectors,
                eigenvalues=eigenvalues,
                confidence=confidence
            )
        
        if abs_angle > self.config.max_angle_threshold:
            return PCASkewResult(
                skew_angle=angle,
                corrected=False,
                reason=f"Angle {angle:.2f}° above maximum threshold {self.config.max_angle_threshold}°",
                center=center,
                eigenvectors=eigenvectors,
                eigenvalues=eigenvalues,
                confidence=confidence
            )
        
        return PCASkewResult(
            skew_angle=angle,
            corrected=True,
            reason="",
            center=center,
            eigenvectors=eigenvectors,
            eigenvalues=eigenvalues,
            confidence=confidence
        )
    
    def correct_skew(
        self,
        image: np.ndarray,
        angle: Optional[float] = None
    ) -> Tuple[np.ndarray, PCASkewResult]:
        """
        Detect and correct document skew using PCA.
        
        Args:
            image: Input image (grayscale or BGR/RGB color)
            angle: Optional specific angle to correct. If not provided,
                   the angle will be detected using PCA.
        
        Returns:
            Tuple of (corrected_image, PCASkewResult)
        """
        # Detect skew if angle not provided
        if angle is not None:
            result = PCASkewResult(
                skew_angle=angle,
                corrected=True,
                reason="Manual angle provided",
                confidence=1.0
            )
        else:
            result = self.detect_skew_angle(image)
        
        # Return original if no correction needed
        if not result.corrected and angle is None:
            return image.copy(), result
        
        # Apply the correction
        corrected = self._rotate_image(image, result.skew_angle)
        
        return corrected, result
    
    def _to_grayscale(self, image: np.ndarray) -> np.ndarray:
        """
        Convert image to grayscale if needed.
        
        Args:
            image: Input image (grayscale or color)
            
        Returns:
            Grayscale image
        """
        if len(image.shape) == 2:
            return image
        elif image.shape[2] == 3:
            return cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        elif image.shape[2] == 4:
            return cv2.cvtColor(image, cv2.COLOR_BGRA2GRAY)
        else:
            raise ValueError(f"Unsupported image shape: {image.shape}")
    
    def _extract_contours(self, gray: np.ndarray) -> List[np.ndarray]:
        """
        Extract contours from a grayscale image.
        
        Args:
            gray: Grayscale input image
            
        Returns:
            List of contour arrays
        """
        if self.config.use_text_contours:
            # Text-based approach: threshold and find text regions
            blurred = cv2.GaussianBlur(gray, (5, 5), 0)
            _, thresh = cv2.threshold(
                blurred, 0, 255,
                cv2.THRESH_BINARY_INV + cv2.THRESH_OTSU
            )
        else:
            # Edge-based approach
            thresh = cv2.Canny(
                gray,
                self.config.canny_low_threshold,
                self.config.canny_high_threshold
            )
        
        # Apply morphological operations to connect nearby elements
        kernel_dilate = np.ones(
            (self.config.dilation_kernel_size, self.config.dilation_kernel_size),
            np.uint8
        )
        dilated = cv2.dilate(
            thresh, kernel_dilate,
            iterations=self.config.dilation_iterations
        )
        
        kernel_erode = np.ones(
            (self.config.erosion_kernel_size, self.config.erosion_kernel_size),
            np.uint8
        )
        eroded = cv2.erode(
            dilated, kernel_erode,
            iterations=self.config.erosion_iterations
        )
        
        # Find contours
        contours, _ = cv2.findContours(
            eroded, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE
        )
        
        return list(contours)
    
    def _compute_pca_orientation(
        self,
        points: np.ndarray
    ) -> Tuple[float, Tuple[int, int], np.ndarray, np.ndarray]:
        """
        Compute the orientation angle using PCA.
        
        Based on the RMIT implementation using cv2.PCACompute2.
        
        Args:
            points: Array of points from contours
            
        Returns:
            Tuple of (angle_degrees, center, eigenvectors, eigenvalues)
        """
        # Reshape points for PCA if needed
        if len(points.shape) == 3:
            # Contour format: (N, 1, 2)
            data_pts = points.reshape(-1, 2).astype(np.float64)
        else:
            data_pts = points.astype(np.float64)
        
        # Perform PCA analysis using OpenCV
        mean = np.empty((0))
        mean, eigenvectors, eigenvalues = cv2.PCACompute2(data_pts, mean)
        
        # Extract center
        center = (int(mean[0, 0]), int(mean[0, 1]))
        
        # Calculate orientation angle from first principal component
        # The first eigenvector points in the direction of maximum variance
        angle_rad = atan2(eigenvectors[0, 1], eigenvectors[0, 0])
        angle_deg = np.rad2deg(angle_rad)
        
        # Normalize angle to be within -45 to 45 degrees
        # (skew is typically within this range)
        while angle_deg > 45:
            angle_deg -= 90
        while angle_deg < -45:
            angle_deg += 90
        
        return angle_deg, center, eigenvectors, eigenvalues.flatten()
    
    def _rotate_image(self, image: np.ndarray, angle: float) -> np.ndarray:
        """
        Rotate the image by the specified angle.
        
        Args:
            image: Input image (grayscale or color)
            angle: Rotation angle in degrees (positive = counter-clockwise)
            
        Returns:
            Rotated image
        """
        h, w = image.shape[:2]
        center = (w // 2, h // 2)
        
        if self.config.expand_canvas:
            # Calculate new dimensions to prevent cropping
            return self._rotate_bound(image, angle)
        else:
            # Simple rotation (may crop corners)
            M = cv2.getRotationMatrix2D(center, angle, 1.0)
            return cv2.warpAffine(
                image, M, (w, h),
                borderMode=cv2.BORDER_CONSTANT,
                borderValue=self._get_border_value(image)
            )
    
    def _rotate_bound(self, image: np.ndarray, angle: float) -> np.ndarray:
        """
        Rotate image with canvas expansion to prevent cropping.
        
        Based on the RMIT implementation (rotate_bound function).
        
        Args:
            image: Input image
            angle: Rotation angle in degrees
            
        Returns:
            Rotated image with expanded canvas
        """
        h, w = image.shape[:2]
        cx, cy = w // 2, h // 2
        
        # Get the rotation matrix
        M = cv2.getRotationMatrix2D((cx, cy), angle, 1.0)
        cos_val = np.abs(M[0, 0])
        sin_val = np.abs(M[0, 1])
        
        # Compute new bounding dimensions
        new_w = int((h * sin_val) + (w * cos_val))
        new_h = int((h * cos_val) + (w * sin_val))
        
        # Adjust the rotation matrix to account for translation
        M[0, 2] += (new_w / 2) - cx
        M[1, 2] += (new_h / 2) - cy
        
        # Perform the rotation
        return cv2.warpAffine(
            image, M, (new_w, new_h),
            borderMode=cv2.BORDER_CONSTANT,
            borderValue=self._get_border_value(image)
        )
    
    def _get_border_value(
        self,
        image: np.ndarray
    ) -> Union[int, Tuple[int, int, int]]:
        """
        Get the appropriate border value based on image type.
        
        Args:
            image: Input image
            
        Returns:
            Border value suitable for the image type
        """
        if len(image.shape) == 2:
            # Grayscale
            if isinstance(self.config.background_color, tuple):
                return self.config.background_color[0]
            return self.config.background_color
        else:
            # Color
            if isinstance(self.config.background_color, int):
                return (
                    self.config.background_color,
                    self.config.background_color,
                    self.config.background_color
                )
            return self.config.background_color
    
    def draw_pca_axes(
        self,
        image: np.ndarray,
        result: PCASkewResult,
        scale: float = 0.02,
        axis_colors: Tuple[Tuple[int, ...], Tuple[int, ...]] = ((0, 255, 0), (255, 255, 0))
    ) -> np.ndarray:
        """
        Draw PCA principal axes on the image for visualization.
        
        Args:
            image: Input image (will be copied)
            result: PCASkewResult from detect_skew_angle
            scale: Scale factor for axis visualization
            axis_colors: Colors for first and second principal axes (BGR)
            
        Returns:
            Image with PCA axes drawn
        """
        if result.center is None or result.eigenvectors is None or result.eigenvalues is None:
            logger.warning("Cannot draw axes: missing PCA data in result")
            return image.copy()
        
        # Ensure we have a color image for drawing
        if len(image.shape) == 2:
            viz = cv2.cvtColor(image, cv2.COLOR_GRAY2BGR)
        else:
            viz = image.copy()
        
        center = result.center
        eigenvectors = result.eigenvectors
        eigenvalues = result.eigenvalues
        
        # Draw center point
        cv2.circle(viz, center, 5, (255, 0, 255), -1)
        
        # Draw first principal axis (dominant direction)
        p1 = (
            int(center[0] + scale * eigenvectors[0, 0] * eigenvalues[0]),
            int(center[1] + scale * eigenvectors[0, 1] * eigenvalues[0])
        )
        self._draw_axis(viz, center, p1, axis_colors[0], scale=1.0)
        
        # Draw second principal axis (perpendicular)
        p2 = (
            int(center[0] - scale * eigenvectors[1, 0] * eigenvalues[1]),
            int(center[1] - scale * eigenvectors[1, 1] * eigenvalues[1])
        )
        self._draw_axis(viz, center, p2, axis_colors[1], scale=5.0)
        
        return viz
    
    def _draw_axis(
        self,
        image: np.ndarray,
        p_start: Tuple[int, int],
        p_end: Tuple[int, int],
        color: Tuple[int, ...],
        scale: float = 1.0
    ) -> None:
        """
        Draw an axis line with arrow heads.
        
        Based on the RMIT drawAxis function.
        
        Args:
            image: Image to draw on (modified in place)
            p_start: Start point (center)
            p_end: End point
            color: Line color (BGR)
            scale: Arrow scale factor
        """
        p = list(p_start)
        q = list(p_end)
        
        # Calculate angle and hypotenuse
        angle = atan2(p[1] - q[1], p[0] - q[0])
        hypotenuse = sqrt((p[1] - q[1])**2 + (p[0] - q[0])**2)
        
        # Lengthen the arrow by scale factor
        q[0] = int(p[0] - scale * hypotenuse * cos(angle))
        q[1] = int(p[1] - scale * hypotenuse * sin(angle))
        
        # Draw main line
        cv2.line(image, (int(p[0]), int(p[1])), (int(q[0]), int(q[1])), color, 2, cv2.LINE_AA)
        
        # Draw arrow hooks
        hook_size = 9
        p_hook1 = (
            int(q[0] + hook_size * cos(angle + pi / 4)),
            int(q[1] + hook_size * sin(angle + pi / 4))
        )
        cv2.line(image, p_hook1, (int(q[0]), int(q[1])), color, 2, cv2.LINE_AA)
        
        p_hook2 = (
            int(q[0] + hook_size * cos(angle - pi / 4)),
            int(q[1] + hook_size * sin(angle - pi / 4))
        )
        cv2.line(image, p_hook2, (int(q[0]), int(q[1])), color, 2, cv2.LINE_AA)


def detect_and_correct_skew(
    image: np.ndarray,
    min_angle: float = 0.5,
    max_angle: float = 45.0,
    expand_canvas: bool = True
) -> Tuple[np.ndarray, float]:
    """
    Convenience function to detect and correct document skew.
    
    Args:
        image: Input document image (grayscale or color)
        min_angle: Minimum skew angle to correct (degrees)
        max_angle: Maximum skew angle to correct (degrees)
        expand_canvas: Whether to expand canvas to prevent cropping
        
    Returns:
        Tuple of (corrected_image, detected_angle)
    
    Example:
        >>> corrected, angle = detect_and_correct_skew(document_image)
        >>> print(f"Corrected {angle:.2f} degrees of skew")
    """
    config = PCASkewConfig(
        min_angle_threshold=min_angle,
        max_angle_threshold=max_angle,
        expand_canvas=expand_canvas
    )
    
    corrector = PCASkewCorrector(config)
    corrected, result = corrector.correct_skew(image)
    
    return corrected, result.skew_angle
