"""
Orientation detection and correction module.

This module detects and corrects document orientation by:
- Scanning multiple rotation angles (0, 90, 180, 270 degrees)
- Scoring each orientation using text and/or face detection
- Selecting the best orientation and applying correction
"""

import logging
from typing import Tuple, Optional, List
from dataclasses import dataclass

import numpy as np
import cv2

from ml.document_preprocessing.config import DocumentConfig, OrientationConfig

logger = logging.getLogger(__name__)


@dataclass
class OrientationResult:
    """Result of orientation detection."""
    
    detected_angle: int  # 0, 90, 180, or 270 degrees
    confidence: float  # 0.0 to 1.0
    scores: dict  # Scores for each angle
    method_used: str  # "text", "face", or "combined"


class OrientationDetector:
    """
    Detects and corrects document orientation.
    
    This class analyzes the document at multiple rotation angles
    and selects the one with the highest probability of being correct
    based on text detection patterns and/or face detection.
    """
    
    def __init__(self, config: Optional[DocumentConfig] = None):
        """
        Initialize the orientation detector.
        
        Args:
            config: Document configuration. Uses defaults if not provided.
        """
        self.config = config or DocumentConfig()
        self.orient_config = self.config.orientation
    
    def detect_orientation(self, image: np.ndarray) -> OrientationResult:
        """
        Detect the orientation of a document image.
        
        Args:
            image: Input image (RGB format)
            
        Returns:
            OrientationResult with detected angle and confidence
        """
        if not self.orient_config.enabled:
            return OrientationResult(
                detected_angle=0,
                confidence=1.0,
                scores={0: 1.0},
                method_used="disabled"
            )
        
        angles = self.orient_config.rotation_angles
        scores = {}
        
        method = self.orient_config.detection_method
        
        for angle in angles:
            rotated = self._rotate_image(image, angle)
            
            if method == "text":
                score = self._score_text_orientation(rotated)
            elif method == "face":
                score = self._score_face_orientation(rotated)
            else:  # combined
                text_score = self._score_text_orientation(rotated)
                face_score = self._score_face_orientation(rotated)
                # Weight text more heavily for documents
                score = 0.7 * text_score + 0.3 * face_score
            
            scores[angle] = score
            logger.debug(f"Angle {angle}: score = {score:.4f}")
        
        # Find best angle
        best_angle = max(scores, key=scores.get)
        best_score = scores[best_angle]
        
        # Calculate confidence (difference between best and second best)
        sorted_scores = sorted(scores.values(), reverse=True)
        if len(sorted_scores) > 1:
            confidence = (sorted_scores[0] - sorted_scores[1]) / max(sorted_scores[0], 0.001)
            confidence = min(1.0, confidence + 0.5)  # Scale up
        else:
            confidence = 1.0
        
        return OrientationResult(
            detected_angle=best_angle,
            confidence=confidence,
            scores=scores,
            method_used=method
        )
    
    def correct_orientation(
        self,
        image: np.ndarray,
        angle: Optional[int] = None
    ) -> Tuple[np.ndarray, int]:
        """
        Correct the orientation of a document image.
        
        Args:
            image: Input image (RGB format)
            angle: Rotation angle to apply. If None, auto-detects.
            
        Returns:
            Tuple of (corrected image, angle applied)
        """
        if angle is None:
            result = self.detect_orientation(image)
            angle = result.detected_angle
        
        if angle == 0:
            return image, 0
        
        rotated = self._rotate_image(image, angle)
        logger.info(f"Corrected orientation by rotating {angle} degrees")
        
        return rotated, angle
    
    def _rotate_image(self, image: np.ndarray, angle: int) -> np.ndarray:
        """
        Rotate image by the specified angle.
        
        Args:
            image: Input image
            angle: Rotation angle (0, 90, 180, 270)
            
        Returns:
            Rotated image
        """
        if angle == 0:
            return image
        elif angle == 90:
            return cv2.rotate(image, cv2.ROTATE_90_CLOCKWISE)
        elif angle == 180:
            return cv2.rotate(image, cv2.ROTATE_180)
        elif angle == 270:
            return cv2.rotate(image, cv2.ROTATE_90_COUNTERCLOCKWISE)
        else:
            # For arbitrary angles, use affine transform
            height, width = image.shape[:2]
            center = (width // 2, height // 2)
            matrix = cv2.getRotationMatrix2D(center, -angle, 1.0)
            
            # Calculate new bounding box size
            cos = np.abs(matrix[0, 0])
            sin = np.abs(matrix[0, 1])
            new_width = int(height * sin + width * cos)
            new_height = int(height * cos + width * sin)
            
            # Adjust transformation matrix
            matrix[0, 2] += (new_width / 2) - center[0]
            matrix[1, 2] += (new_height / 2) - center[1]
            
            return cv2.warpAffine(image, matrix, (new_width, new_height))
    
    def _score_text_orientation(self, image: np.ndarray) -> float:
        """
        Score the orientation based on text detection characteristics.
        
        Uses horizontal edge projection and text line detection to
        determine if the text is correctly oriented.
        
        Args:
            image: Input image (RGB format)
            
        Returns:
            Score (higher is better)
        """
        # Scale down for faster processing
        scale = self.orient_config.text_detection_scale
        height, width = image.shape[:2]
        scaled = cv2.resize(
            image,
            (int(width * scale), int(height * scale)),
            interpolation=cv2.INTER_AREA
        )
        
        # Convert to grayscale
        gray = cv2.cvtColor(scaled, cv2.COLOR_RGB2GRAY)
        
        # Apply edge detection
        edges = cv2.Canny(gray, 50, 150)
        
        # Apply morphological closing to connect text elements
        kernel = cv2.getStructuringElement(cv2.MORPH_RECT, (15, 1))
        closed = cv2.morphologyEx(edges, cv2.MORPH_CLOSE, kernel)
        
        # Calculate horizontal projection (sum of pixels per row)
        h_projection = np.sum(closed, axis=1).astype(np.float32)
        
        # Calculate variance of horizontal projection
        # Higher variance indicates clearer text lines (correct orientation)
        h_variance = np.var(h_projection)
        
        # Also check vertical projection
        kernel_v = cv2.getStructuringElement(cv2.MORPH_RECT, (1, 15))
        closed_v = cv2.morphologyEx(edges, cv2.MORPH_CLOSE, kernel_v)
        v_projection = np.sum(closed_v, axis=0).astype(np.float32)
        v_variance = np.var(v_projection)
        
        # Text documents should have higher horizontal variance (text lines)
        # than vertical variance
        score = h_variance / max(v_variance, 1.0)
        
        # Normalize score
        return min(1.0, score / 10.0)
    
    def _score_face_orientation(self, image: np.ndarray) -> float:
        """
        Score the orientation based on face detection.
        
        ID documents typically have an upright face photo. Detecting
        a face suggests correct orientation.
        
        Args:
            image: Input image (RGB format)
            
        Returns:
            Score (higher is better)
        """
        try:
            # Use OpenCV's Haar cascade for fast face detection
            face_cascade = cv2.CascadeClassifier(
                cv2.data.haarcascades + 'haarcascade_frontalface_default.xml'
            )
            
            gray = cv2.cvtColor(image, cv2.COLOR_RGB2GRAY)
            
            # Detect faces
            faces = face_cascade.detectMultiScale(
                gray,
                scaleFactor=1.1,
                minNeighbors=5,
                minSize=(30, 30)
            )
            
            if len(faces) == 0:
                return 0.0
            
            # Score based on face size and position
            height, width = gray.shape
            best_score = 0.0
            
            for (x, y, w, h) in faces:
                # Face area relative to image
                face_area_ratio = (w * h) / (width * height)
                
                # Face should be in reasonable position (not at edge)
                center_x = (x + w / 2) / width
                center_y = (y + h / 2) / height
                position_score = 1.0 - abs(0.5 - center_x) - abs(0.5 - center_y)
                
                # Combined score
                score = (face_area_ratio * 10 + position_score) / 2
                best_score = max(best_score, score)
            
            return min(1.0, best_score)
            
        except Exception as e:
            logger.warning(f"Face detection failed: {e}")
            return 0.0
    
    def detect_skew_angle(
        self,
        image: np.ndarray,
        angle_range: float = 15.0
    ) -> float:
        """
        Detect fine skew angle for deskewing.
        
        This detects small rotation angles (not 90-degree rotations)
        that may occur when scanning documents.
        
        Args:
            image: Input image (RGB format)
            angle_range: Maximum angle to consider (in degrees)
            
        Returns:
            Skew angle in degrees
        """
        # Convert to grayscale
        gray = cv2.cvtColor(image, cv2.COLOR_RGB2GRAY)
        
        # Edge detection
        edges = cv2.Canny(gray, 50, 150, apertureSize=3)
        
        # Detect lines using Hough transform
        lines = cv2.HoughLinesP(
            edges,
            rho=1,
            theta=np.pi / 180,
            threshold=100,
            minLineLength=100,
            maxLineGap=10
        )
        
        if lines is None or len(lines) == 0:
            return 0.0
        
        # Calculate angles of detected lines
        angles = []
        for line in lines:
            x1, y1, x2, y2 = line[0]
            angle = np.degrees(np.arctan2(y2 - y1, x2 - x1))
            
            # Normalize to -90 to 90 range
            if angle < -45:
                angle += 90
            elif angle > 45:
                angle -= 90
            
            # Only consider angles within range
            if abs(angle) <= angle_range:
                angles.append(angle)
        
        if not angles:
            return 0.0
        
        # Use median angle to avoid outliers
        return float(np.median(angles))
    
    def deskew(
        self,
        image: np.ndarray,
        angle: Optional[float] = None
    ) -> Tuple[np.ndarray, float]:
        """
        Deskew a document image by correcting small rotation angles.
        
        Args:
            image: Input image (RGB format)
            angle: Skew angle to correct. If None, auto-detects.
            
        Returns:
            Tuple of (deskewed image, angle corrected)
        """
        if angle is None:
            angle = self.detect_skew_angle(image)
        
        if abs(angle) < 0.5:
            # Negligible skew
            return image, 0.0
        
        height, width = image.shape[:2]
        center = (width // 2, height // 2)
        
        # Get rotation matrix
        matrix = cv2.getRotationMatrix2D(center, angle, 1.0)
        
        # Rotate
        rotated = cv2.warpAffine(
            image,
            matrix,
            (width, height),
            flags=cv2.INTER_CUBIC,
            borderMode=cv2.BORDER_REPLICATE
        )
        
        logger.info(f"Deskewed image by {angle:.2f} degrees")
        
        return rotated, angle
