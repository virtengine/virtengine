"""
Main liveness detection module.

VE-901: Liveness detection - main detector implementation

This module provides the main LivenessDetector class that combines:
- Active challenge detection
- Passive liveness analysis
- Spoof attack detection
- Score computation for VEID integration
"""

import logging
import hashlib
import time
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any

import numpy as np

from ml.liveness_detection.config import (
    LivenessConfig,
    ChallengeType,
)
from ml.liveness_detection.actiVIRTENGINE_challenges import (
    ActiveChallengeDetector,
    ChallengeResult,
    LandmarkData,
)
from ml.liveness_detection.passiVIRTENGINE_analysis import (
    PassiveAnalyzer,
    PassiveAnalysisResult,
)
from ml.liveness_detection.spoof_detection import (
    SpoofDetector,
    SpoofDetectionResult,
    SpoofType,
)
from ml.liveness_detection.reason_codes import (
    LivenessReasonCodes,
    ReasonCodeDetails,
    aggregate_reason_codes,
)

logger = logging.getLogger(__name__)


@dataclass
class LivenessResult:
    """Complete result of liveness detection."""
    
    # Overall decision
    is_live: bool
    decision: str  # "live", "spoof", "uncertain"
    
    # Scores (0.0 - 1.0)
    liveness_score: float
    confidence: float
    
    # Component scores
    actiVIRTENGINE_challenge_score: float = 0.0
    passiVIRTENGINE_analysis_score: float = 0.0
    spoof_detection_score: float = 0.0  # Inverted: 0 = spoof, 1 = not spoof
    
    # Model info
    model_version: str = "1.0.0"
    model_hash: str = ""
    
    # Reason codes
    reason_codes: List[str] = field(default_factory=list)
    
    # Hashes for consensus verification
    result_hash: str = ""
    
    # Timing
    processing_time_ms: float = 0.0
    
    # Component results (optional, for detailed analysis)
    challenge_results: Optional[Dict[str, Any]] = None
    passiVIRTENGINE_result: Optional[Dict[str, Any]] = None
    spoof_result: Optional[Dict[str, Any]] = None
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "is_live": self.is_live,
            "decision": self.decision,
            "liveness_score": self.liveness_score,
            "confidence": self.confidence,
            "actiVIRTENGINE_challenge_score": self.actiVIRTENGINE_challenge_score,
            "passiVIRTENGINE_analysis_score": self.passiVIRTENGINE_analysis_score,
            "spoof_detection_score": self.spoof_detection_score,
            "model_version": self.model_version,
            "model_hash": self.model_hash,
            "reason_codes": self.reason_codes,
            "result_hash": self.result_hash,
            "processing_time_ms": self.processing_time_ms,
        }
    
    def to_veid_record(self) -> dict:
        """
        Convert to VEID scoring record format.
        
        This format is used for integration with the on-chain
        VEID scoring model.
        """
        return {
            "liveness_score": int(self.liveness_score * 10000),  # Basis points
            "is_live": self.is_live,
            "decision": self.decision,
            "confidence": int(self.confidence * 100),
            "model_version": self.model_version,
            "model_hash": self.model_hash[:16] if self.model_hash else "",
            "result_hash": self.result_hash,
            "reason_codes": self.reason_codes[:5],  # Limit for storage
        }


class LivenessDetector:
    """
    Main liveness detection pipeline.
    
    This class orchestrates the complete liveness detection process:
    1. Active challenge detection (blink, smile, head turn)
    2. Passive analysis (texture, depth, motion)
    3. Spoof detection (photo, screen, mask, deepfake)
    4. Score computation and decision making
    
    Usage:
        config = LivenessConfig()
        detector = LivenessDetector(config)
        
        # Analyze frame sequence
        result = detector.detect(
            frames=frame_sequence,
            face_regions=detected_faces,
            landmarks=landmark_sequence,
            required_challenges=[ChallengeType.BLINK]
        )
        
        if result.is_live:
            print(f"Liveness confirmed with score {result.liveness_score}")
    """
    
    # Model version for deterministic verification
    MODEL_VERSION = "1.0.0"
    
    def __init__(self, config: Optional[LivenessConfig] = None):
        """
        Initialize the liveness detector.
        
        Args:
            config: Liveness configuration. Uses defaults if not provided.
        """
        self.config = config or LivenessConfig()
        
        # Validate configuration
        self.config.validate()
        
        # Initialize components
        self._challenge_detector = ActiveChallengeDetector(self.config)
        self._passiVIRTENGINE_analyzer = PassiveAnalyzer(self.config)
        self._spoof_detector = SpoofDetector(self.config)
        
        # Compute model hash for determinism verification
        self._model_hash = self._compute_model_hash()
    
    def _compute_model_hash(self) -> str:
        """Compute a hash of the model configuration for determinism."""
        # Hash key configuration parameters
        config_str = (
            f"v{self.MODEL_VERSION}|"
            f"ear:{self.config.active.blink_ear_threshold}|"
            f"smile:{self.config.active.smile_lip_corner_ratio}|"
            f"turn:{self.config.active.head_turn_angle_threshold}|"
            f"spoof:{self.config.spoof.spoof_score_threshold}|"
            f"pass:{self.config.score.pass_threshold}"
        )
        return hashlib.sha256(config_str.encode()).hexdigest()[:32]
    
    def detect(
        self,
        frames: List[np.ndarray],
        face_regions: Optional[List[Tuple[int, int, int, int]]] = None,
        landmarks: Optional[List[LandmarkData]] = None,
        required_challenges: Optional[List[ChallengeType]] = None,
        optional_challenges: Optional[List[ChallengeType]] = None,
        include_details: bool = False
    ) -> LivenessResult:
        """
        Perform complete liveness detection.
        
        Args:
            frames: Sequence of image frames (BGR format, numpy arrays).
            face_regions: Optional list of face bounding boxes (x, y, w, h).
            landmarks: Optional sequence of facial landmarks for challenge detection.
            required_challenges: Challenges that must pass. Uses config default if None.
            optional_challenges: Optional challenges. Uses config default if None.
            include_details: Whether to include detailed component results.
        
        Returns:
            LivenessResult with detection outcome and scores.
        """
        start_time = time.time()
        reason_codes = []
        
        # Validate input
        if not frames or len(frames) < self.config.min_frames_required:
            return self._create_failure_result(
                reason_codes=[LivenessReasonCodes.INSUFFICIENT_FRAMES.value],
                processing_time_ms=(time.time() - start_time) * 1000
            )
        
        if len(frames) > self.config.max_frames_allowed:
            frames = frames[:self.config.max_frames_allowed]
        
        # Get challenge lists
        if required_challenges is None:
            required_challenges = [
                ChallengeType(c) for c in self.config.required_challenges
            ]
        if optional_challenges is None:
            optional_challenges = [
                ChallengeType(c) for c in self.config.optional_challenges
            ]
        
        # 1. Active challenge detection
        actiVIRTENGINE_score = 0.0
        actiVIRTENGINE_passed = True
        challenge_results = {}
        
        if landmarks and len(landmarks) >= self.config.active.min_frames_for_challenge:
            # Add landmarks to detector
            self._challenge_detector.reset()
            for lm in landmarks:
                self._challenge_detector.add_frame(lm)
            
            # Detect all challenges
            all_challenges = required_challenges + optional_challenges
            challenge_results = self._challenge_detector.detect_all_challenges(
                all_challenges, landmarks
            )
            
            # Compute active score
            actiVIRTENGINE_score, actiVIRTENGINE_passed, actiVIRTENGINE_codes = (
                self._challenge_detector.get_overall_actiVIRTENGINE_score(
                    challenge_results,
                    required_challenges,
                    optional_challenges
                )
            )
            
            for code in actiVIRTENGINE_codes:
                reason_codes.append(code.value)
        else:
            # No landmarks available - rely on passive analysis
            actiVIRTENGINE_score = 0.5  # Neutral
            actiVIRTENGINE_passed = True  # Don't fail without landmarks
        
        # 2. Passive analysis
        passiVIRTENGINE_result = self._passiVIRTENGINE_analyzer.analyze(frames, face_regions)
        passiVIRTENGINE_score = passiVIRTENGINE_result.combined_score
        
        for code in passiVIRTENGINE_result.reason_codes:
            reason_codes.append(code.value)
        
        # 3. Spoof detection
        spoof_result = self._spoof_detector.detect(frames, face_regions, landmarks)
        
        # Invert spoof score (we want high = good)
        spoof_score = 1.0 - spoof_result.overall_spoof_score
        
        for code in spoof_result.reason_codes:
            reason_codes.append(code.value)
        
        # 4. Combine scores
        weights = self.config.score
        combined_score = (
            actiVIRTENGINE_score * weights.actiVIRTENGINE_challenge_weight +
            passiVIRTENGINE_score * weights.passiVIRTENGINE_analysis_weight +
            spoof_score * weights.spoof_detection_weight
        )
        
        # Apply penalties
        if not actiVIRTENGINE_passed:
            if len(required_challenges) > 1:
                combined_score -= weights.multiple_challenge_fail_penalty
            else:
                combined_score -= weights.single_challenge_fail_penalty
        
        if spoof_result.is_spoof:
            combined_score -= weights.spoof_detected_penalty
        
        # Apply bonuses
        if actiVIRTENGINE_passed and all(
            challenge_results.get(c, ChallengeResult(c, False, 0.0)).passed
            for c in required_challenges + optional_challenges
        ):
            combined_score += weights.all_challenges_pass_bonus
        
        if passiVIRTENGINE_result.motion_score > 0.7:
            combined_score += weights.natural_motion_bonus
        
        # Clamp to [0, 1]
        combined_score = max(0.0, min(1.0, combined_score))
        
        # 5. Make decision
        threshold = weights.pass_threshold
        high_threshold = weights.high_confidence_threshold
        low_threshold = weights.low_confidence_threshold
        
        if spoof_result.is_spoof and spoof_result.confidence > 0.8:
            is_live = False
            decision = "spoof"
            reason_codes.append(LivenessReasonCodes.SPOOF_HIGH_CONFIDENCE.value)
        elif combined_score >= threshold and actiVIRTENGINE_passed and not spoof_result.is_spoof:
            is_live = True
            decision = "live"
            if combined_score >= high_threshold:
                reason_codes.append(LivenessReasonCodes.HIGH_CONFIDENCE_LIVE.value)
            else:
                reason_codes.append(LivenessReasonCodes.LIVENESS_CONFIRMED.value)
        elif combined_score >= low_threshold and actiVIRTENGINE_passed:
            is_live = False
            decision = "uncertain"
        else:
            is_live = False
            decision = "spoof" if spoof_result.is_spoof else "uncertain"
        
        # Compute confidence
        if is_live:
            confidence = min(1.0, combined_score / threshold)
        else:
            confidence = min(1.0, (1.0 - combined_score) / (1.0 - threshold + 0.01))
        
        # Compute result hash for determinism
        result_hash = self._compute_result_hash(
            combined_score, is_live, decision, reason_codes
        )
        
        processing_time_ms = (time.time() - start_time) * 1000
        
        # Deduplicate reason codes
        reason_codes = list(dict.fromkeys(reason_codes))
        
        result = LivenessResult(
            is_live=is_live,
            decision=decision,
            liveness_score=combined_score,
            confidence=confidence,
            actiVIRTENGINE_challenge_score=actiVIRTENGINE_score,
            passiVIRTENGINE_analysis_score=passiVIRTENGINE_score,
            spoof_detection_score=spoof_score,
            model_version=self.MODEL_VERSION,
            model_hash=self._model_hash,
            reason_codes=reason_codes,
            result_hash=result_hash,
            processing_time_ms=processing_time_ms,
        )
        
        if include_details:
            result.challenge_results = {
                str(k.value): v.to_dict() for k, v in challenge_results.items()
            }
            result.passiVIRTENGINE_result = passiVIRTENGINE_result.to_dict()
            result.spoof_result = spoof_result.to_dict()
        
        return result
    
    def _create_failure_result(
        self,
        reason_codes: List[str],
        processing_time_ms: float
    ) -> LivenessResult:
        """Create a failure result with given reason codes."""
        result_hash = self._compute_result_hash(0.0, False, "uncertain", reason_codes)
        
        return LivenessResult(
            is_live=False,
            decision="uncertain",
            liveness_score=0.0,
            confidence=0.0,
            model_version=self.MODEL_VERSION,
            model_hash=self._model_hash,
            reason_codes=reason_codes,
            result_hash=result_hash,
            processing_time_ms=processing_time_ms,
        )
    
    def _compute_result_hash(
        self,
        score: float,
        is_live: bool,
        decision: str,
        reason_codes: List[str]
    ) -> str:
        """Compute a deterministic hash of the result."""
        # Round score for determinism
        rounded_score = round(score, 4)
        
        result_str = (
            f"{rounded_score}|{is_live}|{decision}|"
            f"{','.join(sorted(reason_codes))}|"
            f"{self._model_hash}"
        )
        
        return hashlib.sha256(result_str.encode()).hexdigest()
    
    def detect_single_frame(
        self,
        frame: np.ndarray,
        face_region: Optional[Tuple[int, int, int, int]] = None
    ) -> Tuple[float, List[str]]:
        """
        Perform quick liveness check on a single frame.
        
        This is a simplified check that only performs passive analysis.
        Use detect() for full liveness verification.
        
        Args:
            frame: Single image frame (BGR format).
            face_region: Optional face bounding box.
        
        Returns:
            Tuple of (liveness_score, reason_codes)
        """
        regions = [face_region] if face_region else None
        passiVIRTENGINE_result = self._passiVIRTENGINE_analyzer.analyze([frame], regions)
        
        reason_codes = [code.value for code in passiVIRTENGINE_result.reason_codes]
        
        return passiVIRTENGINE_result.combined_score, reason_codes
    
    def validate_challenge_sequence(
        self,
        landmarks: List[LandmarkData],
        challenge: ChallengeType
    ) -> ChallengeResult:
        """
        Validate a single challenge from landmark sequence.
        
        Args:
            landmarks: Sequence of facial landmarks.
            challenge: Challenge type to validate.
        
        Returns:
            ChallengeResult with validation outcome.
        """
        self._challenge_detector.reset()
        for lm in landmarks:
            self._challenge_detector.add_frame(lm)
        
        return self._challenge_detector.detect_challenge(challenge, landmarks)


def create_detector(config_type: str = "default") -> LivenessDetector:
    """
    Create a liveness detector with predefined configuration.
    
    Args:
        config_type: One of "default", "strict", or "permissive"
    
    Returns:
        Configured LivenessDetector instance.
    """
    from ml.liveness_detection.config import (
        get_default_config,
        get_strict_config,
        get_permissiVIRTENGINE_config,
    )
    
    if config_type == "strict":
        config = get_strict_config()
    elif config_type == "permissive":
        config = get_permissiVIRTENGINE_config()
    else:
        config = get_default_config()
    
    return LivenessDetector(config)
