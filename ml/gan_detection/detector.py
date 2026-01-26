"""
Main GAN detection module.

VE-923: GAN fraud detection - main detector implementation

This module provides the main GANDetector class that combines:
- CNN discriminator for synthetic image detection
- Deepfake detection
- Artifact analysis
- VEID score integration
"""

import logging
import hashlib
import time
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any

import numpy as np

from ml.gan_detection.config import (
    GANDetectionConfig,
    DetectionMode,
    SyntheticImageType,
)
from ml.gan_detection.discriminator import (
    CNNDiscriminator,
    DiscriminatorResult,
    create_discriminator,
)
from ml.gan_detection.deepfake_detection import (
    DeepfakeDetector,
    DeepfakeResult,
    create_deepfake_detector,
)
from ml.gan_detection.artifact_analysis import (
    ArtifactAnalyzer,
    ArtifactResult,
    create_artifact_analyzer,
)
from ml.gan_detection.reason_codes import (
    GANReasonCodes,
    ReasonCodeDetails,
    aggregate_reason_codes,
)

logger = logging.getLogger(__name__)


@dataclass
class GANDetectionResult:
    """Complete result of GAN fraud detection."""
    
    # Overall decision
    is_synthetic: bool
    decision: str  # "authentic", "synthetic", "suspicious", "uncertain"
    
    # Scores (0.0 = authentic, 1.0 = synthetic)
    overall_score: float
    confidence: float
    
    # Component scores
    discriminator_score: float = 0.0
    deepfake_score: float = 0.0
    artifact_score: float = 0.0
    
    # Detection type
    detected_type: Optional[SyntheticImageType] = None
    
    # VEID integration
    veid_penalty: int = 0  # Basis points penalty for VEID score
    veid_adjusted_score: int = 10000  # Score after penalty (max 10000)
    
    # Model info
    model_version: str = "1.0.0"
    model_hash: str = ""
    
    # Reason codes
    reason_codes: List[str] = field(default_factory=list)
    
    # Hashes for consensus verification
    result_hash: str = ""
    
    # Timing
    processing_time_ms: float = 0.0
    frames_analyzed: int = 0
    
    # Component results (optional)
    discriminator_result: Optional[Dict[str, Any]] = None
    deepfake_result: Optional[Dict[str, Any]] = None
    artifact_result: Optional[Dict[str, Any]] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "is_synthetic": self.is_synthetic,
            "decision": self.decision,
            "overall_score": self.overall_score,
            "confidence": self.confidence,
            "discriminator_score": self.discriminator_score,
            "deepfake_score": self.deepfake_score,
            "artifact_score": self.artifact_score,
            "detected_type": self.detected_type.value if self.detected_type else None,
            "veid_penalty": self.veid_penalty,
            "veid_adjusted_score": self.veid_adjusted_score,
            "model_version": self.model_version,
            "model_hash": self.model_hash,
            "reason_codes": self.reason_codes,
            "result_hash": self.result_hash,
            "processing_time_ms": self.processing_time_ms,
            "frames_analyzed": self.frames_analyzed,
        }
    
    def to_veid_record(self) -> dict:
        """
        Convert to VEID scoring record format.
        
        This format is used for integration with the on-chain
        VEID scoring model.
        """
        return {
            "gan_score": int((1 - self.overall_score) * 10000),  # Inverted: higher = more authentic
            "is_synthetic": self.is_synthetic,
            "decision": self.decision,
            "confidence": int(self.confidence * 100),
            "detected_type": self.detected_type.value if self.detected_type else None,
            "veid_penalty": self.veid_penalty,
            "veid_adjusted_score": self.veid_adjusted_score,
            "model_version": self.model_version,
            "model_hash": self.model_hash[:16] if self.model_hash else "",
            "result_hash": self.result_hash,
            "reason_codes": self.reason_codes[:5],
        }


class GANDetector:
    """
    Main GAN fraud detection pipeline.
    
    This class orchestrates the complete GAN detection process:
    1. CNN discriminator classification (real vs synthetic)
    2. Deepfake detection (face swap, expression manipulation)
    3. Artifact analysis (frequency, checkerboard, blending)
    4. Score computation and VEID integration
    
    Usage:
        config = GANDetectionConfig()
        detector = GANDetector(config)
        
        # Analyze single image
        result = detector.detect(image, face_region)
        
        # Analyze video frames
        result = detector.detect_sequence(
            frames=frame_sequence,
            face_regions=detected_faces,
            landmarks=landmark_sequence
        )
        
        # Get VEID record
        veid_record = result.to_veid_record()
    """
    
    MODEL_VERSION = "1.0.0"
    
    def __init__(self, config: Optional[GANDetectionConfig] = None):
        """
        Initialize the GAN detector.
        
        Args:
            config: GAN detection configuration.
        """
        self.config = config or GANDetectionConfig()
        
        # Set determinism
        if self.config.enforce_determinism:
            np.random.seed(self.config.random_seed)
        
        # Initialize component detectors
        self._discriminator = create_discriminator(self.config)
        self._deepfake_detector = create_deepfake_detector(self.config)
        self._artifact_analyzer = create_artifact_analyzer(self.config)
        
        # Compute model hash
        self._model_hash = self._compute_model_hash()
        
        logger.info(
            f"GANDetector initialized: version={self.MODEL_VERSION}, "
            f"mode={self.config.mode.value}, hash={self._model_hash[:8]}"
        )
    
    def _compute_model_hash(self) -> str:
        """Compute deterministic hash of model configuration."""
        config_str = (
            f"version:{self.MODEL_VERSION},"
            f"mode:{self.config.mode.value},"
            f"discriminator:{self._discriminator.model_hash},"
            f"deepfake:{self._deepfake_detector._model_hash},"
            f"seed:{self.config.random_seed}"
        )
        return hashlib.sha256(config_str.encode()).hexdigest()[:32]
    
    def detect(
        self,
        image: np.ndarray,
        face_region: Optional[Tuple[int, int, int, int]] = None,
        landmarks: Optional[Any] = None,
        include_details: bool = False
    ) -> GANDetectionResult:
        """
        Perform GAN detection on a single image.
        
        Args:
            image: Input image (BGR format, uint8).
            face_region: Optional face bounding box (x1, y1, x2, y2).
            landmarks: Optional facial landmarks.
            include_details: Include detailed component results.
        
        Returns:
            GANDetectionResult with detection outcome.
        """
        return self.detect_sequence(
            frames=[image],
            face_regions=[face_region] if face_region else None,
            landmarks=[landmarks] if landmarks else None,
            include_details=include_details,
        )
    
    def detect_sequence(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]] = None,
        landmarks: Optional[List[Any]] = None,
        include_details: bool = False
    ) -> GANDetectionResult:
        """
        Perform GAN detection on a sequence of frames.
        
        Args:
            frames: List of image frames (BGR format).
            face_regions: Optional list of face bounding boxes.
            landmarks: Optional list of facial landmarks.
            include_details: Include detailed component results.
        
        Returns:
            GANDetectionResult with detection outcome.
        """
        start_time = time.time()
        
        reason_codes_detail = []
        
        # Validate input
        if not frames:
            return GANDetectionResult(
                is_synthetic=False,
                decision="uncertain",
                overall_score=0.0,
                confidence=0.0,
                reason_codes=[GANReasonCodes.INVALID_INPUT.value],
                model_version=self.MODEL_VERSION,
                model_hash=self._model_hash,
            )
        
        # Limit frames
        if len(frames) > self.config.max_frames_per_sequence:
            frames = frames[:self.config.max_frames_per_sequence]
            if face_regions:
                face_regions = face_regions[:self.config.max_frames_per_sequence]
            if landmarks:
                landmarks = landmarks[:self.config.max_frames_per_sequence]
        
        # Validate image size
        h, w = frames[0].shape[:2]
        min_h, min_w = self.config.min_image_size
        max_h, max_w = self.config.max_image_size
        
        if h < min_h or w < min_w:
            return GANDetectionResult(
                is_synthetic=False,
                decision="uncertain",
                overall_score=0.0,
                confidence=0.0,
                reason_codes=[GANReasonCodes.IMAGE_TOO_SMALL.value],
                model_version=self.MODEL_VERSION,
                model_hash=self._model_hash,
            )
        
        if h > max_h or w > max_w:
            return GANDetectionResult(
                is_synthetic=False,
                decision="uncertain",
                overall_score=0.0,
                confidence=0.0,
                reason_codes=[GANReasonCodes.IMAGE_TOO_LARGE.value],
                model_version=self.MODEL_VERSION,
                model_hash=self._model_hash,
            )
        
        # Run discriminator on first frame
        discriminator_result = self._discriminator.classify(frames[0])
        discriminator_score = discriminator_result.synthetic_probability
        
        for rc in discriminator_result.reason_codes:
            reason_codes_detail.append(ReasonCodeDetails(
                code=rc,
                description=str(rc.value),
                confidence=discriminator_result.confidence,
                severity=GANReasonCodes.get_severity(rc),
                details={},
            ))
        
        # Run deepfake detection
        deepfake_result = self._deepfake_detector.detect(
            frames=frames,
            face_regions=face_regions,
            landmarks=landmarks,
            include_details=include_details,
        )
        deepfake_score = deepfake_result.deepfake_score
        
        for rc in deepfake_result.reason_codes:
            reason_codes_detail.append(ReasonCodeDetails(
                code=rc,
                description=str(rc.value),
                confidence=deepfake_result.confidence,
                severity=GANReasonCodes.get_severity(rc),
                details={},
            ))
        
        # Run artifact analysis
        first_region = face_regions[0] if face_regions else None
        artifact_result = self._artifact_analyzer.analyze(frames[0], first_region)
        artifact_score = artifact_result.artifact_score
        
        for rc in artifact_result.reason_codes:
            reason_codes_detail.append(ReasonCodeDetails(
                code=rc,
                description=str(rc.value),
                confidence=artifact_result.confidence,
                severity=GANReasonCodes.get_severity(rc),
                details={},
            ))
        
        # Compute overall score
        overall_score = self._compute_overall_score(
            discriminator_score, deepfake_score, artifact_score
        )
        
        # Determine decision
        is_synthetic, decision = self._determine_decision(overall_score)
        
        # Compute confidence
        confidence = abs(overall_score - 0.5) * 2
        
        # Determine detected type
        detected_type = self._determine_detection_type(
            discriminator_result, deepfake_result, artifact_result
        )
        
        # Compute VEID penalty
        veid_penalty, veid_adjusted = self._compute_veid_penalty(
            overall_score, is_synthetic, confidence
        )
        
        # Add final decision reason code
        if is_synthetic:
            if confidence > 0.7:
                reason_codes_detail.append(ReasonCodeDetails(
                    code=GANReasonCodes.SYNTHETIC_IMAGE_DETECTED,
                    description="Synthetic image detected with high confidence",
                    confidence=confidence,
                    severity=3,
                    details={},
                ))
            else:
                reason_codes_detail.append(ReasonCodeDetails(
                    code=GANReasonCodes.GAN_DETECTED,
                    description="GAN-generated content detected",
                    confidence=confidence,
                    severity=2,
                    details={},
                ))
        elif decision == "suspicious":
            reason_codes_detail.append(ReasonCodeDetails(
                code=GANReasonCodes.GAN_LOW_CONFIDENCE,
                description="Possible synthetic content detected",
                confidence=confidence,
                severity=1,
                details={},
            ))
        else:
            reason_codes_detail.append(ReasonCodeDetails(
                code=GANReasonCodes.IMAGE_AUTHENTIC,
                description="Image appears authentic",
                confidence=confidence,
                severity=0,
                details={},
            ))
        
        # Aggregate reason codes
        reason_codes = aggregate_reason_codes(reason_codes_detail)
        
        # Compute result hash
        result_hash = self._compute_result_hash(
            overall_score, discriminator_score, deepfake_score, artifact_score
        )
        
        processing_time = (time.time() - start_time) * 1000
        
        result = GANDetectionResult(
            is_synthetic=is_synthetic,
            decision=decision,
            overall_score=overall_score,
            confidence=confidence,
            discriminator_score=discriminator_score,
            deepfake_score=deepfake_score,
            artifact_score=artifact_score,
            detected_type=detected_type,
            veid_penalty=veid_penalty,
            veid_adjusted_score=veid_adjusted,
            model_version=self.MODEL_VERSION,
            model_hash=self._model_hash,
            reason_codes=reason_codes,
            result_hash=result_hash,
            processing_time_ms=processing_time,
            frames_analyzed=len(frames),
        )
        
        if include_details:
            result.discriminator_result = discriminator_result.to_dict()
            result.deepfake_result = deepfake_result.to_dict()
            result.artifact_result = artifact_result.to_dict()
        
        return result
    
    def _compute_overall_score(
        self,
        discriminator: float,
        deepfake: float,
        artifact: float
    ) -> float:
        """Compute weighted overall GAN detection score."""
        weights = self.config.veid_integration
        
        # Normalize weights
        total_weight = (
            weights.gan_detection_weight +
            weights.deepfake_weight +
            weights.artifact_weight
        )
        
        if total_weight == 0:
            return 0.0
        
        score = (
            weights.gan_detection_weight * discriminator +
            weights.deepfake_weight * deepfake +
            weights.artifact_weight * artifact
        ) / total_weight
        
        return min(1.0, max(0.0, score))
    
    def _determine_decision(self, score: float) -> Tuple[bool, str]:
        """Determine detection decision based on score."""
        threshold = self.config.veid_integration.synthetic_detection_threshold
        high_threshold = self.config.veid_integration.high_confidence_threshold
        low_threshold = self.config.veid_integration.low_confidence_threshold
        
        if score >= high_threshold:
            return True, "synthetic"
        elif score >= threshold:
            return True, "synthetic"
        elif score >= low_threshold:
            return False, "suspicious"
        else:
            return False, "authentic"
    
    def _determine_detection_type(
        self,
        discriminator_result: DiscriminatorResult,
        deepfake_result: DeepfakeResult,
        artifact_result: ArtifactResult
    ) -> Optional[SyntheticImageType]:
        """Determine the most likely type of synthetic content."""
        # Check deepfake types first
        if deepfake_result.detected_type:
            return deepfake_result.detected_type
        
        # Check discriminator
        if discriminator_result.is_synthetic and discriminator_result.confidence > 0.7:
            return SyntheticImageType.GAN_GENERATED
        
        # Check artifacts
        if artifact_result.has_artifacts:
            if artifact_result.checkerboard_score > 0.5:
                return SyntheticImageType.GAN_GENERATED
            elif artifact_result.blending_score > 0.5:
                return SyntheticImageType.MANIPULATED
        
        return None
    
    def _compute_veid_penalty(
        self,
        score: float,
        is_synthetic: bool,
        confidence: float
    ) -> Tuple[int, int]:
        """
        Compute VEID score penalty.
        
        Returns:
            Tuple of (penalty_amount, adjusted_score) in basis points.
        """
        veid_config = self.config.veid_integration
        
        if not is_synthetic:
            # No penalty for authentic images
            return 0, 10000
        
        # Base penalty based on confidence
        if confidence > 0.7:
            penalty = veid_config.high_confidence_penalty
        elif score > veid_config.synthetic_detection_threshold:
            penalty = veid_config.synthetic_detected_penalty
        else:
            penalty = veid_config.deepfake_detected_penalty
        
        # Scale penalty by confidence
        penalty = int(penalty * confidence)
        
        # Clamp to limits
        penalty = max(
            veid_config.min_gan_score_impact,
            min(penalty, veid_config.max_gan_score_impact)
        )
        
        adjusted = max(0, 10000 - penalty)
        
        return penalty, adjusted
    
    def _compute_result_hash(
        self,
        overall: float,
        discriminator: float,
        deepfake: float,
        artifact: float
    ) -> str:
        """Compute deterministic hash of detection result."""
        data = (
            f"{overall:.6f},"
            f"{discriminator:.6f},"
            f"{deepfake:.6f},"
            f"{artifact:.6f},"
            f"{self._model_hash}"
        )
        return hashlib.sha256(data.encode()).hexdigest()[:32]
    
    def detect_single_frame(
        self,
        frame: np.ndarray,
        face_region: Optional[Tuple[int, int, int, int]] = None
    ) -> Tuple[float, List[str]]:
        """
        Quick single-frame GAN detection.
        
        Args:
            frame: Single image frame.
            face_region: Optional face bounding box.
        
        Returns:
            Tuple of (synthetic_score, reason_codes).
        """
        result = self.detect(frame, face_region)
        return result.overall_score, result.reason_codes
    
    @property
    def model_hash(self) -> str:
        """Get model hash for verification."""
        return self._model_hash


def create_detector(config: Optional[GANDetectionConfig] = None) -> GANDetector:
    """
    Factory function to create a GAN detector.
    
    Args:
        config: GAN detection configuration.
    
    Returns:
        Configured GANDetector instance.
    """
    return GANDetector(config)
