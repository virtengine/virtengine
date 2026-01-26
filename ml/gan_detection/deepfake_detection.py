"""
Deepfake detection module.

VE-923: GAN fraud detection - deepfake detection implementation

This module provides detection for various deepfake manipulation types:
- Face swap detection
- Expression manipulation detection
- Morphed face detection
- Temporal consistency analysis
- Artifact analysis specific to deepfakes
"""

import logging
import hashlib
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any

import numpy as np

from ml.gan_detection.config import (
    GANDetectionConfig,
    DeepfakeConfig,
    SyntheticImageType,
)
from ml.gan_detection.reason_codes import (
    GANReasonCodes,
    ReasonCodeDetails,
)

logger = logging.getLogger(__name__)


@dataclass
class FaceSwapResult:
    """Result of face swap detection."""
    
    is_faceswap: bool
    confidence: float
    
    # Analysis scores
    boundary_score: float = 0.0
    blending_score: float = 0.0
    color_match_score: float = 0.0
    
    # Detected regions
    suspicious_regions: List[Tuple[int, int, int, int]] = field(default_factory=list)
    
    # Reason codes
    reason_codes: List[GANReasonCodes] = field(default_factory=list)
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "is_faceswap": self.is_faceswap,
            "confidence": self.confidence,
            "boundary_score": self.boundary_score,
            "blending_score": self.blending_score,
            "color_match_score": self.color_match_score,
            "suspicious_region_count": len(self.suspicious_regions),
            "reason_codes": [rc.value for rc in self.reason_codes],
        }


@dataclass
class ExpressionManipulationResult:
    """Result of expression manipulation detection."""
    
    is_manipulated: bool
    confidence: float
    
    # Analysis scores
    temporal_consistency_score: float = 0.0
    muscle_movement_score: float = 0.0
    micro_expression_score: float = 0.0
    
    # Detected anomalies
    anomaly_frames: List[int] = field(default_factory=list)
    
    # Reason codes
    reason_codes: List[GANReasonCodes] = field(default_factory=list)
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "is_manipulated": self.is_manipulated,
            "confidence": self.confidence,
            "temporal_consistency_score": self.temporal_consistency_score,
            "muscle_movement_score": self.muscle_movement_score,
            "micro_expression_score": self.micro_expression_score,
            "anomaly_frame_count": len(self.anomaly_frames),
            "reason_codes": [rc.value for rc in self.reason_codes],
        }


@dataclass
class DeepfakeResult:
    """Complete result of deepfake detection."""
    
    # Overall decision
    is_deepfake: bool
    decision: str  # "authentic", "deepfake", "suspicious"
    
    # Scores
    deepfake_score: float  # 0.0 = authentic, 1.0 = deepfake
    confidence: float
    
    # Detection type
    detected_type: Optional[SyntheticImageType] = None
    
    # Component results
    faceswap_score: float = 0.0
    expression_manipulation_score: float = 0.0
    morphing_score: float = 0.0
    artifact_score: float = 0.0
    
    # Temporal analysis (video)
    temporal_score: float = 0.0
    blink_analysis_score: float = 0.0
    
    # Model info
    model_version: str = "1.0.0"
    model_hash: str = ""
    
    # Reason codes
    reason_codes: List[GANReasonCodes] = field(default_factory=list)
    
    # Processing info
    processing_time_ms: float = 0.0
    frames_analyzed: int = 0
    
    # Detailed results (optional)
    faceswap_result: Optional[Dict[str, Any]] = None
    expression_result: Optional[Dict[str, Any]] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "is_deepfake": self.is_deepfake,
            "decision": self.decision,
            "deepfake_score": self.deepfake_score,
            "confidence": self.confidence,
            "detected_type": self.detected_type.value if self.detected_type else None,
            "faceswap_score": self.faceswap_score,
            "expression_manipulation_score": self.expression_manipulation_score,
            "morphing_score": self.morphing_score,
            "artifact_score": self.artifact_score,
            "temporal_score": self.temporal_score,
            "blink_analysis_score": self.blink_analysis_score,
            "model_version": self.model_version,
            "reason_codes": [rc.value for rc in self.reason_codes],
            "processing_time_ms": self.processing_time_ms,
            "frames_analyzed": self.frames_analyzed,
        }
    
    def to_veid_record(self) -> dict:
        """Convert to VEID scoring record format."""
        return {
            "deepfake_score": int((1 - self.deepfake_score) * 10000),  # Invert: higher = better
            "is_deepfake": self.is_deepfake,
            "decision": self.decision,
            "confidence": int(self.confidence * 100),
            "detected_type": self.detected_type.value if self.detected_type else None,
            "model_version": self.model_version,
            "model_hash": self.model_hash[:16] if self.model_hash else "",
            "reason_codes": [rc.value for rc in self.reason_codes[:5]],
        }


class DeepfakeDetector:
    """
    Detector for deepfake manipulations.
    
    This class provides comprehensive deepfake detection including:
    - Face swap detection
    - Expression manipulation detection
    - Face morphing detection
    - Temporal consistency analysis
    - Artifact analysis specific to deepfakes
    
    Usage:
        config = DeepfakeConfig()
        detector = DeepfakeDetector(config)
        result = detector.detect(frames, face_regions, landmarks)
    """
    
    MODEL_VERSION = "1.0.0"
    
    def __init__(self, config: Optional[DeepfakeConfig] = None):
        """
        Initialize the deepfake detector.
        
        Args:
            config: Deepfake detection configuration.
        """
        self.config = config or DeepfakeConfig()
        self._model_hash = self._compute_model_hash()
    
    def _compute_model_hash(self) -> str:
        """Compute deterministic model hash."""
        config_str = (
            f"faceswap_threshold:{self.config.faceswap_threshold},"
            f"expression_threshold:{self.config.expression_threshold},"
            f"version:{self.MODEL_VERSION}"
        )
        return hashlib.sha256(config_str.encode()).hexdigest()[:32]
    
    def detect(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]] = None,
        landmarks: Optional[List[Any]] = None,
        include_details: bool = False
    ) -> DeepfakeResult:
        """
        Perform deepfake detection on image frames.
        
        Args:
            frames: List of image frames (BGR format).
            face_regions: Optional list of face bounding boxes.
            landmarks: Optional facial landmarks for each frame.
            include_details: Include detailed component results.
        
        Returns:
            DeepfakeResult with detection outcome.
        """
        import time
        start_time = time.time()
        
        reason_codes = []
        
        # Validate input
        if not frames:
            return DeepfakeResult(
                is_deepfake=False,
                decision="uncertain",
                deepfake_score=0.0,
                confidence=0.0,
                reason_codes=[GANReasonCodes.INVALID_INPUT],
            )
        
        if len(frames) < 1:
            return DeepfakeResult(
                is_deepfake=False,
                decision="uncertain",
                deepfake_score=0.0,
                confidence=0.0,
                reason_codes=[GANReasonCodes.INSUFFICIENT_FRAMES],
            )
        
        # Detect face swap
        faceswap_result = self._detect_faceswap(frames, face_regions)
        faceswap_score = faceswap_result.confidence if faceswap_result.is_faceswap else 0.0
        reason_codes.extend(faceswap_result.reason_codes)
        
        # Detect expression manipulation (requires multiple frames)
        expression_result = self._detect_expression_manipulation(frames, landmarks)
        expression_score = expression_result.confidence if expression_result.is_manipulated else 0.0
        reason_codes.extend(expression_result.reason_codes)
        
        # Detect morphing
        morphing_score = self._detect_morphing(frames[0], face_regions[0] if face_regions else None)
        if morphing_score > self.config.faceswap_threshold:
            reason_codes.append(GANReasonCodes.MORPHED_FACE_DETECTED)
        
        # Artifact analysis
        artifact_score = self._analyze_deepfake_artifacts(frames, face_regions)
        
        # Temporal analysis (if enough frames)
        temporal_score = 0.0
        blink_score = 0.0
        if len(frames) >= self.config.min_frames_for_temporal and self.config.use_temporal_analysis:
            temporal_score = self._analyze_temporal_consistency(frames, landmarks)
            if temporal_score > 0.5:
                reason_codes.append(GANReasonCodes.TEMPORAL_INCONSISTENCY)
            
            if self.config.analyze_blink_pattern and landmarks:
                blink_score = self._analyze_blink_pattern(landmarks)
                if blink_score > 0.5:
                    reason_codes.append(GANReasonCodes.UNNATURAL_BLINK_PATTERN)
        
        # Compute overall deepfake score
        deepfake_score = self._compute_overall_score(
            faceswap_score, expression_score, morphing_score,
            artifact_score, temporal_score
        )
        
        # Determine decision
        is_deepfake = deepfake_score > self.config.faceswap_threshold
        
        if is_deepfake:
            decision = "deepfake"
            reason_codes.append(GANReasonCodes.DEEPFAKE_DETECTED)
        elif deepfake_score > 0.3:
            decision = "suspicious"
        else:
            decision = "authentic"
            if deepfake_score < 0.2:
                reason_codes.append(GANReasonCodes.IMAGE_AUTHENTIC)
        
        # Determine detected type
        detected_type = self._determine_deepfake_type(
            faceswap_score, expression_score, morphing_score
        )
        
        # Compute confidence
        confidence = abs(deepfake_score - 0.5) * 2
        
        processing_time = (time.time() - start_time) * 1000
        
        result = DeepfakeResult(
            is_deepfake=is_deepfake,
            decision=decision,
            deepfake_score=deepfake_score,
            confidence=confidence,
            detected_type=detected_type,
            faceswap_score=faceswap_score,
            expression_manipulation_score=expression_score,
            morphing_score=morphing_score,
            artifact_score=artifact_score,
            temporal_score=temporal_score,
            blink_analysis_score=blink_score,
            model_version=self.MODEL_VERSION,
            model_hash=self._model_hash,
            reason_codes=list(set(reason_codes)),  # Deduplicate
            processing_time_ms=processing_time,
            frames_analyzed=len(frames),
        )
        
        if include_details:
            result.faceswap_result = faceswap_result.to_dict()
            result.expression_result = expression_result.to_dict()
        
        return result
    
    def _detect_faceswap(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]] = None
    ) -> FaceSwapResult:
        """
        Detect face swap manipulation.
        
        Args:
            frames: Image frames.
            face_regions: Face bounding boxes.
        
        Returns:
            FaceSwapResult with detection outcome.
        """
        reason_codes = []
        
        if not frames:
            return FaceSwapResult(
                is_faceswap=False,
                confidence=0.0,
            )
        
        # Analyze first frame for face swap indicators
        frame = frames[0]
        region = face_regions[0] if face_regions else None
        
        # Boundary analysis
        boundary_score = self._analyze_face_boundary(frame, region)
        
        # Blending analysis
        blending_score = self._analyze_blending(frame, region)
        
        # Color match analysis
        color_score = self._analyze_color_consistency(frame, region)
        
        # Combine scores
        combined_score = (boundary_score + blending_score + color_score) / 3
        
        is_faceswap = combined_score > self.config.faceswap_threshold
        
        if is_faceswap:
            reason_codes.append(GANReasonCodes.FACESWAP_DETECTED)
            if blending_score > 0.5:
                reason_codes.append(GANReasonCodes.BLENDING_BOUNDARY_DETECTED)
            if color_score > 0.5:
                reason_codes.append(GANReasonCodes.COLOR_MISMATCH)
        
        return FaceSwapResult(
            is_faceswap=is_faceswap,
            confidence=combined_score,
            boundary_score=boundary_score,
            blending_score=blending_score,
            color_match_score=color_score,
            reason_codes=reason_codes,
        )
    
    def _analyze_face_boundary(
        self,
        frame: np.ndarray,
        region: Optional[Tuple[int, int, int, int]]
    ) -> float:
        """Analyze face boundary for swap artifacts."""
        if region is None:
            # Assume center of image contains face
            h, w = frame.shape[:2]
            margin = min(h, w) // 4
            region = (margin, margin, w - margin, h - margin)
        
        x1, y1, x2, y2 = region
        
        # Extract boundary region
        boundary_width = 10
        
        # Get inner and outer boundary regions
        inner_y1 = max(0, y1 + boundary_width)
        inner_y2 = min(frame.shape[0], y2 - boundary_width)
        inner_x1 = max(0, x1 + boundary_width)
        inner_x2 = min(frame.shape[1], x2 - boundary_width)
        
        if inner_y2 <= inner_y1 or inner_x2 <= inner_x1:
            return 0.0
        
        # Extract regions
        inner_region = frame[inner_y1:inner_y2, inner_x1:inner_x2]
        
        # Compute gradient at boundary
        gray = np.mean(frame, axis=2) if len(frame.shape) == 3 else frame
        
        # Sobel-like gradient computation
        gradient_y = np.abs(np.diff(gray, axis=0))
        gradient_x = np.abs(np.diff(gray, axis=1))
        
        # Check for sharp boundaries
        boundary_gradient = 0.0
        
        # Top boundary
        if y1 > 0 and y1 < gradient_y.shape[0]:
            boundary_gradient += np.mean(gradient_y[y1, x1:x2])
        
        # Bottom boundary
        if y2 > 0 and y2 < gradient_y.shape[0]:
            boundary_gradient += np.mean(gradient_y[y2-1, x1:x2])
        
        # Left boundary
        if x1 > 0 and x1 < gradient_x.shape[1]:
            boundary_gradient += np.mean(gradient_x[y1:y2, x1])
        
        # Right boundary
        if x2 > 0 and x2 < gradient_x.shape[1]:
            boundary_gradient += np.mean(gradient_x[y1:y2, x2-1])
        
        # Normalize
        boundary_gradient /= 4
        
        # Convert to anomaly score (higher gradient = more suspicious)
        score = min(1.0, boundary_gradient / 50.0)
        
        return float(score)
    
    def _analyze_blending(
        self,
        frame: np.ndarray,
        region: Optional[Tuple[int, int, int, int]]
    ) -> float:
        """Analyze blending artifacts at face boundary."""
        if frame is None or frame.size == 0:
            return 0.0
        
        h, w = frame.shape[:2]
        
        if region is None:
            region = (w // 4, h // 4, 3 * w // 4, 3 * h // 4)
        
        x1, y1, x2, y2 = region
        
        # Ensure valid region
        x1 = max(0, min(x1, w - 1))
        x2 = max(x1 + 1, min(x2, w))
        y1 = max(0, min(y1, h - 1))
        y2 = max(y1 + 1, min(y2, h))
        
        face_region = frame[y1:y2, x1:x2]
        
        if face_region.size == 0:
            return 0.0
        
        # Analyze color channel variance at edges
        edge_width = min(5, min(face_region.shape[0], face_region.shape[1]) // 4)
        
        if edge_width < 1:
            return 0.0
        
        # Get edge strips
        top_edge = face_region[:edge_width, :]
        bottom_edge = face_region[-edge_width:, :]
        left_edge = face_region[:, :edge_width]
        right_edge = face_region[:, -edge_width:]
        
        # Compute variance of each edge
        edge_variances = [
            np.var(top_edge),
            np.var(bottom_edge),
            np.var(left_edge),
            np.var(right_edge),
        ]
        
        # High variance at edges can indicate blending
        avg_variance = np.mean(edge_variances)
        
        # Normalize to 0-1 score
        score = min(1.0, avg_variance / 1000.0)
        
        return float(score)
    
    def _analyze_color_consistency(
        self,
        frame: np.ndarray,
        region: Optional[Tuple[int, int, int, int]]
    ) -> float:
        """Analyze color consistency between face and surrounding area."""
        if frame is None or frame.size == 0:
            return 0.0
        
        h, w = frame.shape[:2]
        
        if region is None:
            region = (w // 4, h // 4, 3 * w // 4, 3 * h // 4)
        
        x1, y1, x2, y2 = region
        
        # Ensure valid region
        x1 = max(0, min(x1, w - 1))
        x2 = max(x1 + 1, min(x2, w))
        y1 = max(0, min(y1, h - 1))
        y2 = max(y1 + 1, min(y2, h))
        
        face_region = frame[y1:y2, x1:x2]
        
        # Get surrounding region (expand bounds)
        expand = 20
        surr_x1 = max(0, x1 - expand)
        surr_y1 = max(0, y1 - expand)
        surr_x2 = min(w, x2 + expand)
        surr_y2 = min(h, y2 + expand)
        
        # Create mask for surrounding region (excluding face)
        surrounding = np.zeros((surr_y2 - surr_y1, surr_x2 - surr_x1, frame.shape[2] if len(frame.shape) == 3 else 1), dtype=frame.dtype)
        
        # Just use the expanded region
        surrounding = frame[surr_y1:surr_y2, surr_x1:surr_x2]
        
        if face_region.size == 0 or surrounding.size == 0:
            return 0.0
        
        # Compare mean color
        face_mean = np.mean(face_region, axis=(0, 1))
        surr_mean = np.mean(surrounding, axis=(0, 1))
        
        # Color difference
        color_diff = np.sqrt(np.sum((face_mean - surr_mean) ** 2))
        
        # Normalize (large difference = suspicious)
        score = min(1.0, color_diff / 100.0)
        
        return float(score)
    
    def _detect_expression_manipulation(
        self,
        frames: List[np.ndarray],
        landmarks: Optional[List[Any]]
    ) -> ExpressionManipulationResult:
        """Detect expression manipulation in video sequence."""
        reason_codes = []
        
        if len(frames) < 3:
            return ExpressionManipulationResult(
                is_manipulated=False,
                confidence=0.0,
            )
        
        # Temporal consistency check
        temporal_score = self._analyze_temporal_consistency(frames, landmarks)
        
        # Muscle movement analysis (simulated)
        muscle_score = self._analyze_muscle_movement(frames, landmarks)
        
        # Micro expression analysis (simulated)
        micro_score = 0.0
        if landmarks:
            micro_score = self._analyze_micro_expressions(landmarks)
        
        # Combine scores
        combined_score = (temporal_score + muscle_score + micro_score) / 3
        
        is_manipulated = combined_score > self.config.expression_threshold
        
        if is_manipulated:
            reason_codes.append(GANReasonCodes.EXPRESSION_MANIPULATION_DETECTED)
            if temporal_score > 0.5:
                reason_codes.append(GANReasonCodes.EXPRESSION_DISCONTINUITY)
        
        return ExpressionManipulationResult(
            is_manipulated=is_manipulated,
            confidence=combined_score,
            temporal_consistency_score=temporal_score,
            muscle_movement_score=muscle_score,
            micro_expression_score=micro_score,
            reason_codes=reason_codes,
        )
    
    def _analyze_temporal_consistency(
        self,
        frames: List[np.ndarray],
        landmarks: Optional[List[Any]]
    ) -> float:
        """Analyze temporal consistency across frames."""
        if len(frames) < 2:
            return 0.0
        
        # Compute frame-to-frame differences
        differences = []
        for i in range(len(frames) - 1):
            diff = np.mean(np.abs(
                frames[i].astype(np.float32) - frames[i + 1].astype(np.float32)
            ))
            differences.append(diff)
        
        # Check for sudden jumps in difference (inconsistency)
        diff_array = np.array(differences)
        diff_variance = np.var(diff_array)
        
        # High variance in differences = temporal inconsistency
        score = min(1.0, diff_variance / 100.0)
        
        return float(score)
    
    def _analyze_muscle_movement(
        self,
        frames: List[np.ndarray],
        landmarks: Optional[List[Any]]
    ) -> float:
        """Analyze facial muscle movement patterns."""
        if not landmarks or len(landmarks) < 3:
            return 0.0
        
        # Simulate landmark movement analysis
        movements = []
        for i in range(len(landmarks) - 1):
            if landmarks[i] is not None and landmarks[i + 1] is not None:
                # Compute movement magnitude (simulated)
                movement = np.random.uniform(0, 1)  # Placeholder
                movements.append(movement)
        
        if not movements:
            return 0.0
        
        # Check for unnatural movement patterns
        movement_variance = np.var(movements)
        
        return min(1.0, movement_variance * 2)
    
    def _analyze_micro_expressions(self, landmarks: List[Any]) -> float:
        """Analyze micro expressions for manipulation signs."""
        # Simulated analysis
        return 0.0
    
    def _detect_morphing(
        self,
        frame: np.ndarray,
        region: Optional[Tuple[int, int, int, int]]
    ) -> float:
        """Detect face morphing manipulation."""
        if frame is None or frame.size == 0:
            return 0.0
        
        # Analyze texture patterns for morphing artifacts
        gray = np.mean(frame, axis=2) if len(frame.shape) == 3 else frame
        
        # Compute local texture variance
        patch_size = 16
        h, w = gray.shape
        
        variances = []
        for y in range(0, h - patch_size, patch_size):
            for x in range(0, w - patch_size, patch_size):
                patch = gray[y:y+patch_size, x:x+patch_size]
                variances.append(np.var(patch))
        
        if not variances:
            return 0.0
        
        # Uniformity in variance can indicate morphing
        variance_of_variance = np.var(variances)
        
        # Low variance of variance = suspicious (too uniform)
        score = 1.0 - min(1.0, variance_of_variance / 100.0)
        
        return float(score) * 0.5  # Scale down
    
    def _analyze_deepfake_artifacts(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]]
    ) -> float:
        """Analyze frames for deepfake-specific artifacts."""
        if not frames:
            return 0.0
        
        artifact_scores = []
        
        for i, frame in enumerate(frames[:10]):  # Limit to first 10 frames
            region = face_regions[i] if face_regions and i < len(face_regions) else None
            
            # Eye reflection analysis
            if self.config.analyze_eye_reflection:
                eye_score = self._analyze_eye_reflections(frame, region)
                artifact_scores.append(eye_score)
            
            # Skin texture analysis
            if self.config.analyze_skin_texture:
                skin_score = self._analyze_skin_texture(frame, region)
                artifact_scores.append(skin_score)
        
        if not artifact_scores:
            return 0.0
        
        return float(np.mean(artifact_scores))
    
    def _analyze_eye_reflections(
        self,
        frame: np.ndarray,
        region: Optional[Tuple[int, int, int, int]]
    ) -> float:
        """Analyze eye reflections for inconsistencies."""
        # Simulated - in practice would detect and compare eye reflections
        return 0.0
    
    def _analyze_skin_texture(
        self,
        frame: np.ndarray,
        region: Optional[Tuple[int, int, int, int]]
    ) -> float:
        """Analyze skin texture for synthetic patterns."""
        if frame is None or frame.size == 0:
            return 0.0
        
        h, w = frame.shape[:2]
        
        if region is None:
            region = (w // 4, h // 4, 3 * w // 4, 3 * h // 4)
        
        x1, y1, x2, y2 = region
        x1 = max(0, min(x1, w - 1))
        x2 = max(x1 + 1, min(x2, w))
        y1 = max(0, min(y1, h - 1))
        y2 = max(y1 + 1, min(y2, h))
        
        face_region = frame[y1:y2, x1:x2]
        
        if face_region.size == 0:
            return 0.0
        
        # Convert to grayscale
        gray = np.mean(face_region, axis=2) if len(face_region.shape) == 3 else face_region
        
        # Compute high-frequency content (texture detail)
        high_freq = np.abs(np.diff(np.diff(gray, axis=0), axis=1))
        
        if high_freq.size == 0:
            return 0.0
        
        # Low high-frequency = suspicious (too smooth)
        avg_high_freq = np.mean(high_freq)
        
        # Invert: smooth skin = potentially synthetic
        score = 1.0 - min(1.0, avg_high_freq / 10.0)
        
        return float(score) * 0.3  # Scale down
    
    def _analyze_blink_pattern(self, landmarks: List[Any]) -> float:
        """Analyze blinking pattern for naturalness."""
        # Would analyze eye aspect ratio over time
        # Unnatural patterns: no blinks, too regular, too fast/slow
        return 0.0
    
    def _compute_overall_score(
        self,
        faceswap: float,
        expression: float,
        morphing: float,
        artifact: float,
        temporal: float
    ) -> float:
        """Compute weighted overall deepfake score."""
        weights = {
            "faceswap": 0.30,
            "expression": 0.25,
            "morphing": 0.15,
            "artifact": 0.15,
            "temporal": 0.15,
        }
        
        score = (
            weights["faceswap"] * faceswap +
            weights["expression"] * expression +
            weights["morphing"] * morphing +
            weights["artifact"] * artifact +
            weights["temporal"] * temporal
        )
        
        return min(1.0, max(0.0, score))
    
    def _determine_deepfake_type(
        self,
        faceswap: float,
        expression: float,
        morphing: float
    ) -> Optional[SyntheticImageType]:
        """Determine the most likely deepfake type."""
        threshold = 0.5
        
        max_score = max(faceswap, expression, morphing)
        
        if max_score < threshold:
            return None
        
        if faceswap == max_score:
            return SyntheticImageType.DEEPFAKE_FACESWAP
        elif expression == max_score:
            return SyntheticImageType.DEEPFAKE_EXPRESSION
        elif morphing == max_score:
            return SyntheticImageType.MORPHED
        
        return SyntheticImageType.MANIPULATED
    
    def detect_single_frame(
        self,
        frame: np.ndarray,
        face_region: Optional[Tuple[int, int, int, int]] = None,
        landmarks: Optional[Any] = None
    ) -> Tuple[float, List[GANReasonCodes]]:
        """
        Quick single-frame deepfake detection.
        
        Args:
            frame: Single image frame.
            face_region: Face bounding box.
            landmarks: Facial landmarks.
        
        Returns:
            Tuple of (deepfake_score, reason_codes).
        """
        result = self.detect(
            frames=[frame],
            face_regions=[face_region] if face_region else None,
            landmarks=[landmarks] if landmarks else None,
        )
        
        return result.deepfake_score, result.reason_codes


def create_deepfake_detector(
    config: Optional[GANDetectionConfig] = None
) -> DeepfakeDetector:
    """
    Factory function to create a deepfake detector.
    
    Args:
        config: Full GAN detection config.
    
    Returns:
        Configured DeepfakeDetector instance.
    """
    if config is None:
        return DeepfakeDetector()
    return DeepfakeDetector(config.deepfake)
