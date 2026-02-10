"""
Face detection and alignment module.

This module provides face detection, alignment, and cropping capabilities
using MTCNN or RetinaFace backends for accurate face localization.
"""

import logging
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Any
import hashlib

import numpy as np
import cv2

from ml.facial_verification.config import (
    VerificationConfig, 
    DetectionConfig,
    DetectorBackend
)
from ml.facial_verification.reason_codes import ReasonCodes

logger = logging.getLogger(__name__)


@dataclass
class FaceLandmarks:
    """Facial landmarks for alignment."""
    
    left_eye: Tuple[float, float]
    right_eye: Tuple[float, float]
    nose: Optional[Tuple[float, float]] = None
    mouth_left: Optional[Tuple[float, float]] = None
    mouth_right: Optional[Tuple[float, float]] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "left_eye": self.left_eye,
            "right_eye": self.right_eye,
            "nose": self.nose,
            "mouth_left": self.mouth_left,
            "mouth_right": self.mouth_right,
        }


@dataclass
class FaceDetection:
    """Result of face detection for a single face."""
    
    # Bounding box (x, y, width, height)
    bbox: Tuple[int, int, int, int]
    
    # Confidence score
    confidence: float
    
    # Landmarks (if available)
    landmarks: Optional[FaceLandmarks] = None
    
    # Face area for sorting
    area: int = 0
    
    # Detection metadata
    detector_backend: str = ""
    
    def __post_init__(self):
        """Calculate area from bbox."""
        if self.area == 0:
            x, y, w, h = self.bbox
            self.area = w * h
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "bbox": self.bbox,
            "confidence": self.confidence,
            "landmarks": self.landmarks.to_dict() if self.landmarks else None,
            "area": self.area,
            "detector_backend": self.detector_backend,
        }


@dataclass
class DetectionResult:
    """Result of face detection on an image."""
    
    faces: List[FaceDetection] = field(default_factory=list)
    success: bool = True
    error_code: Optional[ReasonCodes] = None
    processing_time_ms: float = 0.0
    
    @property
    def num_faces(self) -> int:
        """Number of detected faces."""
        return len(self.faces)
    
    @property
    def has_face(self) -> bool:
        """Whether at least one face was detected."""
        return len(self.faces) > 0
    
    def get_largest_face(self) -> Optional[FaceDetection]:
        """Get the face with the largest bounding box area."""
        if not self.faces:
            return None
        return max(self.faces, key=lambda f: f.area)
    
    def get_highest_confidence_face(self) -> Optional[FaceDetection]:
        """Get the face with the highest confidence score."""
        if not self.faces:
            return None
        return max(self.faces, key=lambda f: f.confidence)
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "faces": [f.to_dict() for f in self.faces],
            "success": self.success,
            "error_code": self.error_code.value if self.error_code else None,
            "num_faces": self.num_faces,
            "processing_time_ms": self.processing_time_ms,
        }


class FaceDetector:
    """
    Face detection and alignment using MTCNN or RetinaFace.
    
    This class provides deterministic face detection with support for
    multiple backends and consistent alignment for embedding extraction.
    """
    
    def __init__(self, config: Optional[VerificationConfig] = None):
        """
        Initialize the face detector.
        
        Args:
            config: Verification configuration. Uses defaults if not provided.
        """
        self.config = config or VerificationConfig()
        self.detection_config = self.config.detection
        self._detector = None
        self._initialized = False
    
    def _lazy_init(self) -> None:
        """Lazy initialization of the detector backend."""
        if self._initialized:
            return
        
        backend = self.detection_config.detector_backend
        
        try:
            if backend == DetectorBackend.MTCNN:
                from mtcnn import MTCNN
                self._detector = MTCNN(
                    min_face_size=self.detection_config.min_face_size,
                    scale_factor=self.detection_config.scale_factor
                )
            elif backend == DetectorBackend.RETINAFACE:
                # RetinaFace will be used via DeepFace
                self._detector = None
            elif backend == DetectorBackend.OPENCV:
                cascade_path = cv2.data.haarcascades + 'haarcascade_frontalface_default.xml'
                self._detector = cv2.CascadeClassifier(cascade_path)
            else:
                # Default to MTCNN
                from mtcnn import MTCNN
                self._detector = MTCNN(
                    min_face_size=self.detection_config.min_face_size,
                    scale_factor=self.detection_config.scale_factor
                )
            
            self._initialized = True
            logger.info(f"Initialized face detector with backend: {backend.value}")
            
        except ImportError as e:
            logger.error(f"Failed to import detector backend {backend.value}: {e}")
            raise RuntimeError(f"Detector backend {backend.value} not available: {e}")
    
    def detect(self, image: np.ndarray) -> DetectionResult:
        """
        Detect faces in an image.
        
        Args:
            image: Input image (BGR format)
            
        Returns:
            DetectionResult with list of detected faces
        """
        import time
        start_time = time.time()
        
        if image is None or image.size == 0:
            return DetectionResult(
                faces=[],
                success=False,
                error_code=ReasonCodes.INVALID_IMAGE_FORMAT
            )
        
        try:
            self._lazy_init()
            
            backend = self.detection_config.detector_backend
            
            if backend == DetectorBackend.MTCNN:
                faces = self._detect_mtcnn(image)
            elif backend == DetectorBackend.OPENCV:
                faces = self._detect_opencv(image)
            elif backend == DetectorBackend.RETINAFACE:
                faces = self._detect_retinaface(image)
            else:
                faces = self._detect_mtcnn(image)
            
            processing_time = (time.time() - start_time) * 1000
            
            # Filter by confidence threshold
            faces = [f for f in faces if f.confidence >= self.detection_config.confidence_threshold]
            
            if not faces:
                return DetectionResult(
                    faces=[],
                    success=True,
                    error_code=ReasonCodes.NO_FACE_DETECTED,
                    processing_time_ms=processing_time
                )
            
            # Check for multiple faces
            if len(faces) > 1 and not self.detection_config.allow_multiple_faces:
                if self.detection_config.select_largest_face:
                    faces = [max(faces, key=lambda f: f.area)]
                else:
                    return DetectionResult(
                        faces=faces,
                        success=False,
                        error_code=ReasonCodes.MULTIPLE_FACES,
                        processing_time_ms=processing_time
                    )
            
            return DetectionResult(
                faces=faces,
                success=True,
                processing_time_ms=processing_time
            )
            
        except Exception as e:
            logger.error(f"Face detection error: {e}")
            return DetectionResult(
                faces=[],
                success=False,
                error_code=ReasonCodes.DETECTION_ERROR,
                processing_time_ms=(time.time() - start_time) * 1000
            )
    
    def _detect_mtcnn(self, image: np.ndarray) -> List[FaceDetection]:
        """Detect faces using MTCNN."""
        # MTCNN expects RGB
        if len(image.shape) == 3 and image.shape[2] == 3:
            rgb_image = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
        else:
            rgb_image = image
        
        detections = self._detector.detect_faces(rgb_image)
        
        faces = []
        for det in detections:
            bbox = tuple(det['box'])  # x, y, width, height
            confidence = det['confidence']
            
            keypoints = det.get('keypoints', {})
            landmarks = None
            if keypoints:
                landmarks = FaceLandmarks(
                    left_eye=keypoints.get('left_eye', (0, 0)),
                    right_eye=keypoints.get('right_eye', (0, 0)),
                    nose=keypoints.get('nose'),
                    mouth_left=keypoints.get('mouth_left'),
                    mouth_right=keypoints.get('mouth_right'),
                )
            
            faces.append(FaceDetection(
                bbox=bbox,
                confidence=confidence,
                landmarks=landmarks,
                detector_backend=DetectorBackend.MTCNN.value
            ))
        
        return faces
    
    def _detect_opencv(self, image: np.ndarray) -> List[FaceDetection]:
        """Detect faces using OpenCV Haar Cascade."""
        gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY) if len(image.shape) == 3 else image
        
        detections = self._detector.detectMultiScale(
            gray,
            scaleFactor=1.1,
            minNeighbors=5,
            minSize=(self.detection_config.min_face_size, self.detection_config.min_face_size)
        )
        
        faces = []
        for (x, y, w, h) in detections:
            faces.append(FaceDetection(
                bbox=(x, y, w, h),
                confidence=0.99,  # OpenCV doesn't provide confidence
                landmarks=None,
                detector_backend=DetectorBackend.OPENCV.value
            ))
        
        return faces
    
    def _detect_retinaface(self, image: np.ndarray) -> List[FaceDetection]:
        """Detect faces using RetinaFace via DeepFace."""
        try:
            from deepface import DeepFace
            from deepface.detectors import FaceDetector as DFDetector
            
            # Use DeepFace's built-in detection
            face_objs = DeepFace.extract_faces(
                img_path=image,
                detector_backend='retinaface',
                enforce_detection=False,
                align=False
            )
            
            faces = []
            for face_obj in face_objs:
                region = face_obj.get('facial_area', {})
                x = region.get('x', 0)
                y = region.get('y', 0)
                w = region.get('w', 0)
                h = region.get('h', 0)
                confidence = face_obj.get('confidence', 0.99)
                
                faces.append(FaceDetection(
                    bbox=(x, y, w, h),
                    confidence=confidence,
                    landmarks=None,
                    detector_backend=DetectorBackend.RETINAFACE.value
                ))
            
            return faces
            
        except Exception as e:
            logger.warning(f"RetinaFace detection failed, falling back to MTCNN: {e}")
            return self._detect_mtcnn(image)
    
    def align(
        self, 
        image: np.ndarray, 
        detection: FaceDetection,
        target_size: Tuple[int, int] = (224, 224)
    ) -> np.ndarray:
        """
        Align face based on eye positions.
        
        This ensures consistent face orientation for embedding extraction
        by rotating the face so that the eyes are horizontal.
        
        Args:
            image: Input image
            detection: Face detection with landmarks
            target_size: Output size for aligned face
            
        Returns:
            Aligned face image
        """
        if detection.landmarks is None:
            # No landmarks available, just crop
            return self.crop(image, detection, target_size)
        
        left_eye = detection.landmarks.left_eye
        right_eye = detection.landmarks.right_eye
        
        # Calculate angle between eyes
        dY = right_eye[1] - left_eye[1]
        dX = right_eye[0] - left_eye[0]
        angle = np.degrees(np.arctan2(dY, dX))
        
        # Calculate center point between eyes
        eye_center = (
            (left_eye[0] + right_eye[0]) // 2,
            (left_eye[1] + right_eye[1]) // 2
        )
        
        # Get rotation matrix
        M = cv2.getRotationMatrix2D(eye_center, angle, 1.0)
        
        # Apply rotation
        h, w = image.shape[:2]
        aligned = cv2.warpAffine(
            image,
            M,
            (w, h),
            flags=cv2.INTER_LINEAR,
            borderMode=cv2.BORDER_REPLICATE,
        )
        
        # Crop the aligned face
        return self.crop(aligned, detection, target_size)
    
    def crop(
        self, 
        image: np.ndarray, 
        detection: FaceDetection,
        target_size: Optional[Tuple[int, int]] = None
    ) -> np.ndarray:
        """
        Crop face region with margin.
        
        Args:
            image: Input image
            detection: Face detection result
            target_size: Optional output size (will resize if provided)
            
        Returns:
            Cropped (and optionally resized) face image
        """
        x, y, w, h = detection.bbox
        margin = self.detection_config.crop_margin
        
        # Add margin
        margin_x = int(w * margin)
        margin_y = int(h * margin)
        
        # Calculate new bounding box
        x1 = max(0, x - margin_x)
        y1 = max(0, y - margin_y)
        x2 = min(image.shape[1], x + w + margin_x)
        y2 = min(image.shape[0], y + h + margin_y)
        
        # Crop
        face = image[y1:y2, x1:x2]
        
        # Resize if target size specified
        if target_size is not None and face.size > 0:
            face = cv2.resize(face, target_size, interpolation=cv2.INTER_AREA)
        
        return face
    
    def detect_and_align(
        self, 
        image: np.ndarray,
        target_size: Tuple[int, int] = (224, 224)
    ) -> Tuple[Optional[np.ndarray], DetectionResult]:
        """
        Detect, align, and crop the primary face in an image.
        
        This is a convenience method that combines detection, alignment,
        and cropping into a single call.
        
        Args:
            image: Input image
            target_size: Output size for aligned face
            
        Returns:
            Tuple of (aligned face image or None, detection result)
        """
        result = self.detect(image)
        
        if not result.success or not result.has_face:
            return None, result
        
        # Get primary face
        if self.detection_config.select_largest_face:
            primary_face = result.get_largest_face()
        else:
            primary_face = result.get_highest_confidence_face()
        
        if primary_face is None:
            result.error_code = ReasonCodes.NO_FACE_DETECTED
            return None, result
        
        # Align and crop
        try:
            if self.detection_config.align_faces and primary_face.landmarks:
                aligned_face = self.align(image, primary_face, target_size)
            else:
                aligned_face = self.crop(image, primary_face, target_size)
            
            return aligned_face, result
            
        except Exception as e:
            logger.error(f"Face alignment error: {e}")
            result.error_code = ReasonCodes.ALIGNMENT_ERROR
            return None, result
    
    def compute_detection_hash(self, result: DetectionResult) -> str:
        """
        Compute deterministic hash of detection result.
        
        Args:
            result: Detection result
            
        Returns:
            SHA256 hash of detection data
        """
        data = str(result.to_dict()).encode('utf-8')
        return hashlib.sha256(data).hexdigest()
