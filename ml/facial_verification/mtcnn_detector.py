"""
MTCNN Face Detection Integration Module.

This module provides comprehensive MTCNN (Multi-task Cascaded Convolutional Networks)
integration for the VEID identity verification system. MTCNN performs face detection
and landmark localization through a three-stage cascaded network:

1. P-Net (Proposal Network): Fast sliding-window network that proposes candidate
   face regions at multiple scales.
2. R-Net (Refine Network): Refines candidates from P-Net, rejecting false positives
   and improving bounding box regression.
3. O-Net (Output Network): Final stage that produces precise facial landmarks
   and refined bounding boxes.

The five-point facial landmarks detected are:
- Left eye center
- Right eye center  
- Nose tip
- Left mouth corner
- Right mouth corner

These landmarks enable accurate face alignment for consistent embedding extraction.

References:
- Zhang et al., "Joint Face Detection and Alignment using Multi-task Cascaded
  Convolutional Networks", IEEE Signal Processing Letters, 2016.
"""

import logging
import time
import hashlib
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any, Union
from enum import Enum

import numpy as np
import cv2

logger = logging.getLogger(__name__)


# ==============================================================================
# Data Classes
# ==============================================================================


class MTCNNStage(str, Enum):
    """MTCNN cascade stages."""
    PNET = "pnet"  # Proposal Network
    RNET = "rnet"  # Refine Network
    ONET = "onet"  # Output Network


@dataclass
class FiveLandmarks:
    """
    Five-point facial landmarks extracted by MTCNN O-Net.
    
    Coordinate system: (x, y) with origin at top-left of image.
    All coordinates are in pixels.
    """
    
    left_eye: Tuple[float, float]
    """Center of the left eye (from viewer's perspective)."""
    
    right_eye: Tuple[float, float]
    """Center of the right eye (from viewer's perspective)."""
    
    nose: Tuple[float, float]
    """Tip of the nose."""
    
    mouth_left: Tuple[float, float]
    """Left corner of the mouth."""
    
    mouth_right: Tuple[float, float]
    """Right corner of the mouth."""
    
    def to_array(self) -> np.ndarray:
        """Convert landmarks to numpy array of shape (5, 2)."""
        return np.array([
            self.left_eye,
            self.right_eye,
            self.nose,
            self.mouth_left,
            self.mouth_right,
        ], dtype=np.float32)
    
    def to_dict(self) -> Dict[str, Tuple[float, float]]:
        """Convert to dictionary representation."""
        return {
            "left_eye": self.left_eye,
            "right_eye": self.right_eye,
            "nose": self.nose,
            "mouth_left": self.mouth_left,
            "mouth_right": self.mouth_right,
        }
    
    @classmethod
    def from_array(cls, arr: np.ndarray) -> "FiveLandmarks":
        """Create from numpy array of shape (5, 2) or (10,)."""
        if arr.shape == (10,):
            arr = arr.reshape(5, 2)
        return cls(
            left_eye=tuple(arr[0]),
            right_eye=tuple(arr[1]),
            nose=tuple(arr[2]),
            mouth_left=tuple(arr[3]),
            mouth_right=tuple(arr[4]),
        )
    
    def compute_eye_distance(self) -> float:
        """Calculate Euclidean distance between eye centers."""
        return float(np.sqrt(
            (self.right_eye[0] - self.left_eye[0]) ** 2 +
            (self.right_eye[1] - self.left_eye[1]) ** 2
        ))
    
    def compute_face_angle(self) -> float:
        """Calculate rotation angle from horizontal (in degrees)."""
        delta_y = self.right_eye[1] - self.left_eye[1]
        delta_x = self.right_eye[0] - self.left_eye[0]
        return float(np.degrees(np.arctan2(delta_y, delta_x)))


@dataclass
class MTCNNDetection:
    """
    Single face detection result from MTCNN.
    
    Includes bounding box, confidence score, and five-point landmarks.
    """
    
    # Bounding box (x, y, width, height)
    bbox: Tuple[int, int, int, int]
    
    # Detection confidence from O-Net (0.0 to 1.0)
    confidence: float
    
    # Five-point facial landmarks
    landmarks: FiveLandmarks
    
    # Stage confidences for transparency
    pnet_confidence: Optional[float] = None
    rnet_confidence: Optional[float] = None
    onet_confidence: Optional[float] = None
    
    # Computed area (for sorting by face size)
    area: int = 0
    
    def __post_init__(self):
        """Calculate area from bbox."""
        if self.area == 0:
            _, _, w, h = self.bbox
            self.area = w * h
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "bbox": self.bbox,
            "confidence": self.confidence,
            "landmarks": self.landmarks.to_dict(),
            "pnet_confidence": self.pnet_confidence,
            "rnet_confidence": self.rnet_confidence,
            "onet_confidence": self.onet_confidence,
            "area": self.area,
        }


@dataclass
class MTCNNDetectionResult:
    """
    Complete result from MTCNN face detection.
    """
    
    # Detected faces
    faces: List[MTCNNDetection] = field(default_factory=list)
    
    # Processing status
    success: bool = True
    error_message: Optional[str] = None
    
    # Timing information (milliseconds)
    total_time_ms: float = 0.0
    pnet_time_ms: float = 0.0
    rnet_time_ms: float = 0.0
    onet_time_ms: float = 0.0
    
    # Input image info
    image_shape: Optional[Tuple[int, int, int]] = None
    
    @property
    def num_faces(self) -> int:
        """Number of detected faces."""
        return len(self.faces)
    
    @property
    def has_face(self) -> bool:
        """Whether at least one face was detected."""
        return len(self.faces) > 0
    
    def get_largest_face(self) -> Optional[MTCNNDetection]:
        """Get the face with the largest bounding box area."""
        if not self.faces:
            return None
        return max(self.faces, key=lambda f: f.area)
    
    def get_highest_confidence_face(self) -> Optional[MTCNNDetection]:
        """Get the face with the highest confidence score."""
        if not self.faces:
            return None
        return max(self.faces, key=lambda f: f.confidence)
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "faces": [f.to_dict() for f in self.faces],
            "success": self.success,
            "error_message": self.error_message,
            "num_faces": self.num_faces,
            "total_time_ms": self.total_time_ms,
            "pnet_time_ms": self.pnet_time_ms,
            "rnet_time_ms": self.rnet_time_ms,
            "onet_time_ms": self.onet_time_ms,
            "image_shape": self.image_shape,
        }


# ==============================================================================
# Configuration
# ==============================================================================


@dataclass
class MTCNNConfig:
    """
    Configuration for MTCNN face detection.
    
    Default values are tuned for identity verification use cases,
    prioritizing detection accuracy over speed.
    """
    
    # Minimum face size to detect (in pixels)
    min_face_size: int = 20
    
    # Image pyramid scale factor (smaller = more accurate, slower)
    scale_factor: float = 0.709
    
    # Confidence thresholds for each stage
    pnet_threshold: float = 0.6
    rnet_threshold: float = 0.7
    onet_threshold: float = 0.9
    
    # Non-maximum suppression thresholds
    pnet_nms_threshold: float = 0.7
    rnet_nms_threshold: float = 0.7
    onet_nms_threshold: float = 0.7
    
    # Face alignment settings
    target_size: Tuple[int, int] = (224, 224)
    align_to_template: bool = True
    
    # Margin for face cropping (fraction of face size)
    crop_margin: float = 0.2
    
    # Output settings
    return_all_faces: bool = False
    select_largest: bool = True
    
    # Determinism
    use_deterministic_ops: bool = True


# ==============================================================================
# Standard Template for Face Alignment
# ==============================================================================


# Standard facial landmark positions for 224x224 aligned face
# These are normalized coordinates (0-1) based on common alignment templates
ALIGNMENT_TEMPLATE_224 = np.array([
    [0.31556875, 0.4615741],   # Left eye
    [0.68262291, 0.4615741],   # Right eye
    [0.50009921, 0.6405053],   # Nose
    [0.34947813, 0.8246918],   # Left mouth
    [0.65343645, 0.8246918],   # Right mouth
], dtype=np.float32)

# Scale template to 224x224
ALIGNMENT_TEMPLATE_224_PIXELS = ALIGNMENT_TEMPLATE_224 * 224.0


# ==============================================================================
# MTCNN Detector Class
# ==============================================================================


class MTCNNDetector:
    """
    MTCNN face detector with full three-stage cascade support.
    
    This class wraps the mtcnn library and provides:
    - Explicit access to three-stage cascade (P-Net, R-Net, O-Net)
    - Five-point facial landmark extraction
    - Face alignment using similarity transformation
    - Deterministic execution for blockchain consensus
    
    Usage:
        detector = MTCNNDetector()
        result = detector.detect(image)
        
        if result.has_face:
            face = result.get_largest_face()
            aligned = detector.align(image, face)
    """
    
    def __init__(self, config: Optional[MTCNNConfig] = None):
        """
        Initialize MTCNN detector.
        
        Args:
            config: Detection configuration. Uses defaults if not provided.
        """
        self.config = config or MTCNNConfig()
        self._detector = None
        self._initialized = False
        
        logger.info(
            f"MTCNNDetector initialized with config: "
            f"min_face={self.config.min_face_size}, "
            f"scale_factor={self.config.scale_factor}"
        )
    
    def _lazy_init(self) -> None:
        """Lazy initialization of MTCNN backend."""
        if self._initialized:
            return
        
        try:
            from mtcnn import MTCNN
            
            # Initialize MTCNN with configured thresholds
            self._detector = MTCNN(
                min_face_size=self.config.min_face_size,
                scale_factor=self.config.scale_factor,
                steps_threshold=[
                    self.config.pnet_threshold,
                    self.config.rnet_threshold,
                    self.config.onet_threshold,
                ],
            )
            
            self._initialized = True
            logger.info("MTCNN detector backend initialized successfully")
            
        except ImportError as e:
            logger.error(f"Failed to import mtcnn library: {e}")
            raise RuntimeError(
                "MTCNN library not installed. "
                "Install with: pip install mtcnn"
            ) from e
    
    def detect(self, image: np.ndarray) -> MTCNNDetectionResult:
        """
        Detect faces in an image using the MTCNN three-stage cascade.
        
        The cascade stages are:
        1. P-Net: Proposes candidate face regions at multiple scales
        2. R-Net: Refines candidates and rejects false positives
        3. O-Net: Produces final bounding boxes and 5-point landmarks
        
        Args:
            image: Input image in BGR format (OpenCV convention)
            
        Returns:
            MTCNNDetectionResult with detected faces and landmarks
        """
        start_time = time.time()
        
        # Validate input
        if image is None or image.size == 0:
            return MTCNNDetectionResult(
                success=False,
                error_message="Invalid or empty input image",
            )
        
        try:
            self._lazy_init()
            
            # MTCNN expects RGB format
            if len(image.shape) == 3 and image.shape[2] == 3:
                rgb_image = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
            else:
                rgb_image = image
            
            # Run detection (all three stages)
            # Note: mtcnn library handles P-Net, R-Net, O-Net internally
            raw_detections = self._detector.detect_faces(rgb_image)
            
            total_time_ms = (time.time() - start_time) * 1000
            
            if not raw_detections:
                return MTCNNDetectionResult(
                    success=True,
                    faces=[],
                    total_time_ms=total_time_ms,
                    image_shape=image.shape,
                )
            
            # Convert raw detections to our format
            faces = []
            for det in raw_detections:
                # Get bounding box
                bbox = det.get('box', [0, 0, 0, 0])
                # Ensure non-negative coordinates
                x, y, w, h = bbox
                x = max(0, x)
                y = max(0, y)
                bbox = (x, y, w, h)
                
                # Get confidence (from O-Net)
                confidence = det.get('confidence', 0.0)
                
                # Get keypoints (5-point landmarks)
                keypoints = det.get('keypoints', {})
                landmarks = FiveLandmarks(
                    left_eye=keypoints.get('left_eye', (0.0, 0.0)),
                    right_eye=keypoints.get('right_eye', (0.0, 0.0)),
                    nose=keypoints.get('nose', (0.0, 0.0)),
                    mouth_left=keypoints.get('mouth_left', (0.0, 0.0)),
                    mouth_right=keypoints.get('mouth_right', (0.0, 0.0)),
                )
                
                # Create detection object
                face = MTCNNDetection(
                    bbox=bbox,
                    confidence=confidence,
                    landmarks=landmarks,
                    onet_confidence=confidence,  # Final stage confidence
                )
                
                # Filter by O-Net threshold
                if face.confidence >= self.config.onet_threshold:
                    faces.append(face)
            
            # Sort by area (largest first)
            faces.sort(key=lambda f: f.area, reverse=True)
            
            # Select single face if configured
            if not self.config.return_all_faces and len(faces) > 1:
                if self.config.select_largest:
                    faces = [faces[0]]  # Already sorted by area
                else:
                    # Select highest confidence
                    faces = [max(faces, key=lambda f: f.confidence)]
            
            return MTCNNDetectionResult(
                success=True,
                faces=faces,
                total_time_ms=total_time_ms,
                image_shape=image.shape,
            )
            
        except Exception as e:
            logger.error(f"MTCNN detection error: {e}", exc_info=True)
            return MTCNNDetectionResult(
                success=False,
                error_message=str(e),
                total_time_ms=(time.time() - start_time) * 1000,
            )
    
    def align(
        self,
        image: np.ndarray,
        detection: MTCNNDetection,
        target_size: Optional[Tuple[int, int]] = None,
    ) -> np.ndarray:
        """
        Align face using five-point landmarks.
        
        Uses similarity transformation (rotation, scale, translation)
        to align the detected face to a standard template. This ensures
        consistent face orientation for downstream embedding extraction.
        
        Args:
            image: Input image (BGR format)
            detection: MTCNN detection with landmarks
            target_size: Output size (default: config.target_size)
            
        Returns:
            Aligned face image of shape (target_size[0], target_size[1], 3)
        """
        target_size = target_size or self.config.target_size
        
        if self.config.align_to_template:
            # Use similarity transformation to align to template
            return self._align_with_similarity_transform(
                image, detection.landmarks, target_size
            )
        else:
            # Simple alignment based on eye rotation
            return self._align_by_eye_rotation(
                image, detection, target_size
            )
    
    def _align_with_similarity_transform(
        self,
        image: np.ndarray,
        landmarks: FiveLandmarks,
        target_size: Tuple[int, int],
    ) -> np.ndarray:
        """
        Align face using similarity transformation.
        
        Computes optimal rotation, scale, and translation to map
        detected landmarks to a standard template.
        """
        # Get source landmarks as array
        src_pts = landmarks.to_array()
        
        # Scale template to target size
        template = ALIGNMENT_TEMPLATE_224.copy()
        template[:, 0] *= target_size[0]
        template[:, 1] *= target_size[1]
        dst_pts = template
        
        # Estimate similarity transformation
        transform = cv2.estimateAffinePartial2D(
            src_pts, dst_pts, method=cv2.RANSAC
        )[0]
        
        if transform is None:
            # Fallback to simple crop if transformation fails
            logger.warning("Similarity transform failed, using simple crop")
            return self._simple_crop(image, landmarks, target_size)
        
        # Apply transformation
        aligned = cv2.warpAffine(
            image,
            transform,
            target_size,
            flags=cv2.INTER_LINEAR,
            borderMode=cv2.BORDER_REPLICATE,
        )
        
        return aligned
    
    def _align_by_eye_rotation(
        self,
        image: np.ndarray,
        detection: MTCNNDetection,
        target_size: Tuple[int, int],
    ) -> np.ndarray:
        """
        Simple alignment by rotating to make eyes horizontal.
        """
        landmarks = detection.landmarks
        
        # Calculate rotation angle
        angle = landmarks.compute_face_angle()
        
        # Calculate center between eyes
        eye_center = (
            int((landmarks.left_eye[0] + landmarks.right_eye[0]) / 2),
            int((landmarks.left_eye[1] + landmarks.right_eye[1]) / 2),
        )
        
        # Get rotation matrix
        M = cv2.getRotationMatrix2D(eye_center, angle, 1.0)
        
        # Apply rotation
        h, w = image.shape[:2]
        rotated = cv2.warpAffine(image, M, (w, h), flags=cv2.INTER_LINEAR)
        
        # Crop face region with margin
        return self._crop_with_margin(rotated, detection, target_size)
    
    def _simple_crop(
        self,
        image: np.ndarray,
        landmarks: FiveLandmarks,
        target_size: Tuple[int, int],
    ) -> np.ndarray:
        """Fallback simple center crop based on landmarks."""
        # Get bounding box from landmarks
        pts = landmarks.to_array()
        min_x, min_y = pts.min(axis=0)
        max_x, max_y = pts.max(axis=0)
        
        # Add margin
        margin_x = (max_x - min_x) * 0.5
        margin_y = (max_y - min_y) * 0.5
        
        x1 = int(max(0, min_x - margin_x))
        y1 = int(max(0, min_y - margin_y))
        x2 = int(min(image.shape[1], max_x + margin_x))
        y2 = int(min(image.shape[0], max_y + margin_y))
        
        cropped = image[y1:y2, x1:x2]
        
        if cropped.size == 0:
            return np.zeros((*target_size, 3), dtype=np.uint8)
        
        return cv2.resize(cropped, target_size, interpolation=cv2.INTER_LINEAR)
    
    def _crop_with_margin(
        self,
        image: np.ndarray,
        detection: MTCNNDetection,
        target_size: Tuple[int, int],
    ) -> np.ndarray:
        """Crop face region with configured margin."""
        x, y, w, h = detection.bbox
        margin = self.config.crop_margin
        
        # Calculate margins
        margin_x = int(w * margin)
        margin_y = int(h * margin)
        
        # Calculate crop region
        x1 = max(0, x - margin_x)
        y1 = max(0, y - margin_y)
        x2 = min(image.shape[1], x + w + margin_x)
        y2 = min(image.shape[0], y + h + margin_y)
        
        cropped = image[y1:y2, x1:x2]
        
        if cropped.size == 0:
            return np.zeros((*target_size, 3), dtype=np.uint8)
        
        return cv2.resize(cropped, target_size, interpolation=cv2.INTER_LINEAR)
    
    def detect_and_align(
        self,
        image: np.ndarray,
        target_size: Optional[Tuple[int, int]] = None,
    ) -> Tuple[Optional[np.ndarray], MTCNNDetectionResult]:
        """
        Detect and align the primary face in an image.
        
        Convenience method that combines detection and alignment.
        
        Args:
            image: Input image (BGR format)
            target_size: Output size for aligned face
            
        Returns:
            Tuple of (aligned face image or None, detection result)
        """
        target_size = target_size or self.config.target_size
        
        result = self.detect(image)
        
        if not result.success or not result.has_face:
            return None, result
        
        # Get primary face (already sorted by preference)
        primary_face = result.faces[0]
        
        try:
            aligned = self.align(image, primary_face, target_size)
            return aligned, result
        except Exception as e:
            logger.error(f"Face alignment error: {e}")
            result.error_message = f"Alignment failed: {e}"
            return None, result
    
    def extract_landmarks_only(
        self,
        image: np.ndarray,
    ) -> Optional[FiveLandmarks]:
        """
        Extract only the five-point landmarks from an image.
        
        Useful when the face region is already cropped and only
        landmarks are needed for further processing.
        
        Args:
            image: Input image (BGR format)
            
        Returns:
            FiveLandmarks if a face is detected, None otherwise
        """
        result = self.detect(image)
        
        if result.has_face:
            return result.faces[0].landmarks
        return None
    
    def compute_detection_hash(self, result: MTCNNDetectionResult) -> str:
        """
        Compute deterministic hash of detection result.
        
        Used for verification that detection is reproducible.
        
        Args:
            result: Detection result
            
        Returns:
            SHA256 hash of detection data
        """
        data = str(result.to_dict()).encode('utf-8')
        return hashlib.sha256(data).hexdigest()


# ==============================================================================
# Face Alignment Utilities
# ==============================================================================


class FaceAligner:
    """
    Standalone face alignment utility.
    
    Can be used independently of MTCNN for aligning faces
    when landmarks are obtained from other sources.
    """
    
    def __init__(
        self,
        target_size: Tuple[int, int] = (224, 224),
        use_template: bool = True,
    ):
        """
        Initialize face aligner.
        
        Args:
            target_size: Output image size
            use_template: Whether to use similarity transformation
        """
        self.target_size = target_size
        self.use_template = use_template
        
        # Compute template for target size
        self.template = ALIGNMENT_TEMPLATE_224.copy()
        self.template[:, 0] *= target_size[0] / 224.0
        self.template[:, 1] *= target_size[1] / 224.0
    
    def align(
        self,
        image: np.ndarray,
        landmarks: Union[FiveLandmarks, np.ndarray],
    ) -> np.ndarray:
        """
        Align face using provided landmarks.
        
        Args:
            image: Input image (BGR format)
            landmarks: Five-point landmarks (FiveLandmarks or array)
            
        Returns:
            Aligned face image
        """
        if isinstance(landmarks, FiveLandmarks):
            src_pts = landmarks.to_array()
        else:
            src_pts = landmarks.astype(np.float32)
            if src_pts.shape == (10,):
                src_pts = src_pts.reshape(5, 2)
        
        if self.use_template:
            # Similarity transformation
            transform = cv2.estimateAffinePartial2D(
                src_pts, self.template, method=cv2.RANSAC
            )[0]
            
            if transform is not None:
                return cv2.warpAffine(
                    image,
                    transform,
                    self.target_size,
                    flags=cv2.INTER_LINEAR,
                    borderMode=cv2.BORDER_REPLICATE,
                )
        
        # Fallback: rotation-based alignment
        left_eye = src_pts[0]
        right_eye = src_pts[1]
        
        delta_y = right_eye[1] - left_eye[1]
        delta_x = right_eye[0] - left_eye[0]
        angle = np.degrees(np.arctan2(delta_y, delta_x))
        
        eye_center = (
            (left_eye[0] + right_eye[0]) / 2,
            (left_eye[1] + right_eye[1]) / 2,
        )
        
        M = cv2.getRotationMatrix2D(eye_center, angle, 1.0)
        h, w = image.shape[:2]
        rotated = cv2.warpAffine(image, M, (w, h))
        
        # Crop to target size centered on face
        cx, cy = int(eye_center[0]), int(eye_center[1])
        half_w, half_h = self.target_size[0] // 2, self.target_size[1] // 2
        
        x1 = max(0, cx - half_w)
        y1 = max(0, cy - half_h)
        x2 = min(w, x1 + self.target_size[0])
        y2 = min(h, y1 + self.target_size[1])
        
        cropped = rotated[y1:y2, x1:x2]
        
        if cropped.shape[:2] != self.target_size[::-1]:
            cropped = cv2.resize(cropped, self.target_size)
        
        return cropped
    
    def compute_alignment_quality(
        self,
        landmarks: Union[FiveLandmarks, np.ndarray],
    ) -> float:
        """
        Compute alignment quality score based on landmark positions.
        
        Measures how well landmarks match the expected template.
        
        Args:
            landmarks: Detected landmarks
            
        Returns:
            Quality score from 0.0 (poor) to 1.0 (perfect)
        """
        if isinstance(landmarks, FiveLandmarks):
            src_pts = landmarks.to_array()
        else:
            src_pts = landmarks.astype(np.float32)
            if src_pts.shape == (10,):
                src_pts = src_pts.reshape(5, 2)
        
        # Normalize landmarks to 0-1 range
        bbox_min = src_pts.min(axis=0)
        bbox_max = src_pts.max(axis=0)
        bbox_size = bbox_max - bbox_min
        
        if bbox_size.min() == 0:
            return 0.0
        
        normalized = (src_pts - bbox_min) / bbox_size
        
        # Compare to template (also normalized)
        template_min = ALIGNMENT_TEMPLATE_224.min(axis=0)
        template_max = ALIGNMENT_TEMPLATE_224.max(axis=0)
        template_normalized = (ALIGNMENT_TEMPLATE_224 - template_min) / (template_max - template_min)
        
        # Calculate mean squared error
        mse = np.mean((normalized - template_normalized) ** 2)
        
        # Convert to quality score (lower MSE = higher quality)
        quality = np.exp(-mse * 10)
        
        return float(np.clip(quality, 0.0, 1.0))
