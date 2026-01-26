"""
Active liveness challenge detection module.

VE-901: Liveness detection - active challenge implementation

This module provides detection for active liveness challenges:
- Blink detection using eye aspect ratio (EAR)
- Smile detection using mouth geometry and classification
- Head turn detection using pose estimation
- Head nod detection using vertical pose changes
- Eyebrow raise detection using landmark analysis
"""

import logging
import time
from dataclasses import dataclass, field
from typing import List, Optional, Tuple, Dict, Any
from enum import Enum

import numpy as np

from ml.liveness_detection.config import (
    LivenessConfig,
    ActiveChallengeConfig,
    ChallengeType,
)
from ml.liveness_detection.reason_codes import (
    LivenessReasonCodes,
    ReasonCodeDetails,
)

logger = logging.getLogger(__name__)


@dataclass
class ChallengeResult:
    """Result of a single challenge detection."""
    
    challenge_type: ChallengeType
    passed: bool
    confidence: float  # 0.0 - 1.0
    detected_at_frame: Optional[int] = None
    duration_ms: float = 0.0
    reason_codes: List[LivenessReasonCodes] = field(default_factory=list)
    details: Dict[str, Any] = field(default_factory=dict)
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "challenge_type": self.challenge_type.value,
            "passed": self.passed,
            "confidence": self.confidence,
            "detected_at_frame": self.detected_at_frame,
            "duration_ms": self.duration_ms,
            "reason_codes": [rc.value for rc in self.reason_codes],
            "details": self.details,
        }


@dataclass
class LandmarkData:
    """Facial landmark data for a single frame."""
    
    # Eye landmarks (6 points per eye)
    left_eye: Optional[np.ndarray] = None  # Shape: (6, 2)
    right_eye: Optional[np.ndarray] = None  # Shape: (6, 2)
    
    # Mouth landmarks (20 points)
    mouth: Optional[np.ndarray] = None  # Shape: (20, 2)
    
    # Eyebrow landmarks (5 points per eyebrow)
    left_eyebrow: Optional[np.ndarray] = None  # Shape: (5, 2)
    right_eyebrow: Optional[np.ndarray] = None  # Shape: (5, 2)
    
    # Nose landmarks
    nose: Optional[np.ndarray] = None  # Shape: (9, 2)
    
    # Full face landmarks (68 or 468 points)
    full_landmarks: Optional[np.ndarray] = None
    
    # Pose estimation (pitch, yaw, roll in degrees)
    pose_pitch: float = 0.0  # Up/down
    pose_yaw: float = 0.0    # Left/right
    pose_roll: float = 0.0   # Tilt
    
    # Detection confidence
    confidence: float = 0.0
    
    # Frame metadata
    frame_index: int = 0
    timestamp_ms: float = 0.0


class ActiveChallengeDetector:
    """
    Detector for active liveness challenges.
    
    This class analyzes sequences of facial landmarks to detect
    whether the user has completed requested challenges.
    """
    
    # Eye landmark indices for EAR calculation (68-point model)
    LEFT_EYE_INDICES = [36, 37, 38, 39, 40, 41]
    RIGHT_EYE_INDICES = [42, 43, 44, 45, 46, 47]
    
    # Mouth landmark indices
    OUTER_MOUTH_INDICES = [48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59]
    INNER_MOUTH_INDICES = [60, 61, 62, 63, 64, 65, 66, 67]
    
    # Eyebrow landmark indices
    LEFT_EYEBROW_INDICES = [17, 18, 19, 20, 21]
    RIGHT_EYEBROW_INDICES = [22, 23, 24, 25, 26]
    
    def __init__(self, config: Optional[LivenessConfig] = None):
        """
        Initialize the active challenge detector.
        
        Args:
            config: Liveness configuration. Uses defaults if not provided.
        """
        self.config = config or LivenessConfig()
        self.actiVIRTENGINE_config = self.config.active
        
        # State tracking
        self._landmark_history: List[LandmarkData] = []
        self._challenge_states: Dict[ChallengeType, Dict[str, Any]] = {}
        
    def reset(self) -> None:
        """Reset detector state for a new session."""
        self._landmark_history.clear()
        self._challenge_states.clear()
    
    def add_frame(self, landmarks: LandmarkData) -> None:
        """
        Add a frame's landmarks to the history.
        
        Args:
            landmarks: Facial landmark data for this frame.
        """
        self._landmark_history.append(landmarks)
    
    def detect_challenge(
        self,
        challenge_type: ChallengeType,
        landmarks_sequence: Optional[List[LandmarkData]] = None
    ) -> ChallengeResult:
        """
        Detect if a specific challenge was completed.
        
        Args:
            challenge_type: Type of challenge to detect.
            landmarks_sequence: Sequence of landmarks to analyze.
                              Uses internal history if not provided.
        
        Returns:
            ChallengeResult with detection outcome.
        """
        sequence = landmarks_sequence or self._landmark_history
        
        if len(sequence) < self.actiVIRTENGINE_config.min_frames_for_challenge:
            return ChallengeResult(
                challenge_type=challenge_type,
                passed=False,
                confidence=0.0,
                reason_codes=[LivenessReasonCodes.INSUFFICIENT_FRAMES],
            )
        
        if challenge_type == ChallengeType.BLINK:
            return self._detect_blink(sequence)
        elif challenge_type == ChallengeType.SMILE:
            return self._detect_smile(sequence)
        elif challenge_type in (ChallengeType.HEAD_TURN_LEFT, ChallengeType.HEAD_TURN_RIGHT):
            return self._detect_head_turn(sequence, challenge_type)
        elif challenge_type == ChallengeType.HEAD_NOD:
            return self._detect_head_nod(sequence)
        elif challenge_type == ChallengeType.RAISE_EYEBROWS:
            return self._detect_eyebrow_raise(sequence)
        else:
            return ChallengeResult(
                challenge_type=challenge_type,
                passed=False,
                confidence=0.0,
                reason_codes=[LivenessReasonCodes.INVALID_INPUT],
            )
    
    def detect_all_challenges(
        self,
        challenge_types: List[ChallengeType],
        landmarks_sequence: Optional[List[LandmarkData]] = None
    ) -> Dict[ChallengeType, ChallengeResult]:
        """
        Detect multiple challenges.
        
        Args:
            challenge_types: List of challenges to detect.
            landmarks_sequence: Sequence of landmarks to analyze.
        
        Returns:
            Dictionary mapping challenge types to results.
        """
        results = {}
        for challenge in challenge_types:
            results[challenge] = self.detect_challenge(challenge, landmarks_sequence)
        return results
    
    def _compute_eye_aspect_ratio(self, eye_landmarks: np.ndarray) -> float:
        """
        Compute the Eye Aspect Ratio (EAR) for blink detection.
        
        EAR = (||p2-p6|| + ||p3-p5||) / (2 * ||p1-p4||)
        
        Args:
            eye_landmarks: 6 eye landmark points.
        
        Returns:
            Eye aspect ratio value.
        """
        if eye_landmarks is None or len(eye_landmarks) < 6:
            return 1.0  # Default to open eye
        
        # Vertical distances
        v1 = np.linalg.norm(eye_landmarks[1] - eye_landmarks[5])
        v2 = np.linalg.norm(eye_landmarks[2] - eye_landmarks[4])
        
        # Horizontal distance
        h = np.linalg.norm(eye_landmarks[0] - eye_landmarks[3])
        
        if h == 0:
            return 1.0
        
        ear = (v1 + v2) / (2.0 * h)
        return float(ear)
    
    def _detect_blink(self, sequence: List[LandmarkData]) -> ChallengeResult:
        """
        Detect blink using Eye Aspect Ratio (EAR).
        
        A blink is detected when EAR drops below threshold for
        consecutive frames, then returns above threshold.
        """
        threshold = self.actiVIRTENGINE_config.blink_ear_threshold
        min_frames = self.actiVIRTENGINE_config.blink_consecutiVIRTENGINE_frames
        
        # Compute EAR for each frame
        ear_values = []
        for landmarks in sequence:
            if landmarks.left_eye is not None and landmarks.right_eye is not None:
                left_ear = self._compute_eye_aspect_ratio(landmarks.left_eye)
                right_ear = self._compute_eye_aspect_ratio(landmarks.right_eye)
                avg_ear = (left_ear + right_ear) / 2.0
            elif landmarks.full_landmarks is not None:
                # Extract from 68-point model
                left_eye = landmarks.full_landmarks[self.LEFT_EYE_INDICES]
                right_eye = landmarks.full_landmarks[self.RIGHT_EYE_INDICES]
                left_ear = self._compute_eye_aspect_ratio(left_eye)
                right_ear = self._compute_eye_aspect_ratio(right_eye)
                avg_ear = (left_ear + right_ear) / 2.0
            else:
                avg_ear = 1.0  # Assume open if no landmarks
            
            ear_values.append(avg_ear)
        
        # Find blink patterns
        blink_detected = False
        blink_frame = None
        blink_duration = 0.0
        consecutiVIRTENGINE_closed = 0
        blink_start_frame = None
        
        for i, ear in enumerate(ear_values):
            if ear < threshold:
                if consecutiVIRTENGINE_closed == 0:
                    blink_start_frame = i
                consecutiVIRTENGINE_closed += 1
            else:
                if consecutiVIRTENGINE_closed >= min_frames:
                    # Blink completed
                    blink_detected = True
                    blink_frame = blink_start_frame
                    
                    # Calculate duration
                    if blink_start_frame is not None:
                        start_ts = sequence[blink_start_frame].timestamp_ms
                        end_ts = sequence[i].timestamp_ms
                        blink_duration = end_ts - start_ts
                    break
                consecutiVIRTENGINE_closed = 0
        
        # Validate blink duration
        reason_codes = []
        if blink_detected:
            min_duration = self.actiVIRTENGINE_config.blink_min_duration_ms
            max_duration = self.actiVIRTENGINE_config.blink_max_duration_ms
            
            if blink_duration < min_duration:
                blink_detected = False
                reason_codes.append(LivenessReasonCodes.BLINK_TOO_FAST)
            elif blink_duration > max_duration:
                blink_detected = False
                reason_codes.append(LivenessReasonCodes.BLINK_TOO_SLOW)
        else:
            reason_codes.append(LivenessReasonCodes.BLINK_NOT_DETECTED)
        
        # Compute confidence based on EAR variation
        if len(ear_values) > 0:
            ear_range = max(ear_values) - min(ear_values)
            confidence = min(1.0, ear_range / 0.3)  # Normalize to expected range
        else:
            confidence = 0.0
        
        return ChallengeResult(
            challenge_type=ChallengeType.BLINK,
            passed=blink_detected,
            confidence=confidence if blink_detected else confidence * 0.5,
            detected_at_frame=blink_frame,
            duration_ms=blink_duration,
            reason_codes=reason_codes,
            details={
                "ear_min": min(ear_values) if ear_values else 0.0,
                "ear_max": max(ear_values) if ear_values else 0.0,
                "ear_threshold": threshold,
            },
        )
    
    def _compute_mouth_aspect_ratio(self, mouth_landmarks: np.ndarray) -> float:
        """Compute mouth aspect ratio for smile detection."""
        if mouth_landmarks is None or len(mouth_landmarks) < 12:
            return 0.0
        
        # Horizontal distance (mouth corners)
        width = np.linalg.norm(mouth_landmarks[0] - mouth_landmarks[6])
        
        # Vertical distances
        height1 = np.linalg.norm(mouth_landmarks[2] - mouth_landmarks[10])
        height2 = np.linalg.norm(mouth_landmarks[4] - mouth_landmarks[8])
        avg_height = (height1 + height2) / 2.0
        
        if avg_height == 0:
            return float('inf')
        
        return float(width / avg_height)
    
    def _detect_smile(self, sequence: List[LandmarkData]) -> ChallengeResult:
        """
        Detect smile using mouth geometry.
        
        A smile is detected when the mouth width to height ratio
        exceeds the threshold for sufficient duration.
        """
        ratio_threshold = self.actiVIRTENGINE_config.smile_lip_corner_ratio
        min_duration = self.actiVIRTENGINE_config.smile_min_duration_ms
        
        # Compute mouth ratios
        ratios = []
        for landmarks in sequence:
            if landmarks.mouth is not None:
                ratio = self._compute_mouth_aspect_ratio(landmarks.mouth)
            elif landmarks.full_landmarks is not None:
                mouth = landmarks.full_landmarks[self.OUTER_MOUTH_INDICES]
                ratio = self._compute_mouth_aspect_ratio(mouth)
            else:
                ratio = 0.0
            ratios.append(ratio)
        
        # Find smile patterns
        smile_detected = False
        smile_frame = None
        smile_duration = 0.0
        smile_start_frame = None
        
        for i, ratio in enumerate(ratios):
            if ratio > ratio_threshold:
                if smile_start_frame is None:
                    smile_start_frame = i
            else:
                if smile_start_frame is not None:
                    start_ts = sequence[smile_start_frame].timestamp_ms
                    end_ts = sequence[i].timestamp_ms
                    duration = end_ts - start_ts
                    
                    if duration >= min_duration:
                        smile_detected = True
                        smile_frame = smile_start_frame
                        smile_duration = duration
                        break
                    
                    smile_start_frame = None
        
        # Check if smile is ongoing at end of sequence
        if not smile_detected and smile_start_frame is not None:
            start_ts = sequence[smile_start_frame].timestamp_ms
            end_ts = sequence[-1].timestamp_ms
            duration = end_ts - start_ts
            
            if duration >= min_duration:
                smile_detected = True
                smile_frame = smile_start_frame
                smile_duration = duration
        
        reason_codes = []
        if not smile_detected:
            if max(ratios) > ratio_threshold * 0.8:
                reason_codes.append(LivenessReasonCodes.SMILE_INSUFFICIENT)
            else:
                reason_codes.append(LivenessReasonCodes.SMILE_NOT_DETECTED)
        
        confidence = min(1.0, max(ratios) / (ratio_threshold * 1.5)) if ratios else 0.0
        
        return ChallengeResult(
            challenge_type=ChallengeType.SMILE,
            passed=smile_detected,
            confidence=confidence if smile_detected else confidence * 0.5,
            detected_at_frame=smile_frame,
            duration_ms=smile_duration,
            reason_codes=reason_codes,
            details={
                "max_ratio": max(ratios) if ratios else 0.0,
                "ratio_threshold": ratio_threshold,
            },
        )
    
    def _detect_head_turn(
        self,
        sequence: List[LandmarkData],
        challenge_type: ChallengeType
    ) -> ChallengeResult:
        """
        Detect head turn using pose estimation.
        
        Detects left or right head turn based on yaw angle changes.
        """
        angle_threshold = self.actiVIRTENGINE_config.head_turn_angle_threshold
        max_angle = self.actiVIRTENGINE_config.head_turn_max_angle
        min_duration = self.actiVIRTENGINE_config.head_turn_min_duration_ms
        
        is_left = challenge_type == ChallengeType.HEAD_TURN_LEFT
        
        # Get yaw angles
        yaw_angles = [lm.pose_yaw for lm in sequence]
        
        if not yaw_angles:
            return ChallengeResult(
                challenge_type=challenge_type,
                passed=False,
                confidence=0.0,
                reason_codes=[LivenessReasonCodes.HEAD_TURN_NOT_DETECTED],
            )
        
        # Find baseline (initial head position)
        baseline_yaw = np.median(yaw_angles[:min(5, len(yaw_angles))])
        
        # Find turn events
        turn_detected = False
        turn_frame = None
        turn_duration = 0.0
        turn_start_frame = None
        max_turn_angle = 0.0
        
        for i, yaw in enumerate(yaw_angles):
            delta = yaw - baseline_yaw
            
            # Check direction
            correct_direction = (is_left and delta < 0) or (not is_left and delta > 0)
            abs_delta = abs(delta)
            
            if abs_delta > max_turn_angle:
                max_turn_angle = abs_delta
            
            if correct_direction and abs_delta >= angle_threshold:
                if turn_start_frame is None:
                    turn_start_frame = i
            else:
                if turn_start_frame is not None:
                    start_ts = sequence[turn_start_frame].timestamp_ms
                    end_ts = sequence[i].timestamp_ms
                    duration = end_ts - start_ts
                    
                    if duration >= min_duration:
                        turn_detected = True
                        turn_frame = turn_start_frame
                        turn_duration = duration
                        break
                    
                    turn_start_frame = None
        
        reason_codes = []
        if not turn_detected:
            if max_turn_angle >= angle_threshold:
                # Check if wrong direction
                max_idx = yaw_angles.index(max(yaw_angles, key=abs))
                delta = yaw_angles[max_idx] - baseline_yaw
                if (is_left and delta > 0) or (not is_left and delta < 0):
                    reason_codes.append(LivenessReasonCodes.HEAD_TURN_WRONG_DIRECTION)
                elif max_turn_angle > max_angle:
                    reason_codes.append(LivenessReasonCodes.HEAD_TURN_TOO_EXTREME)
                else:
                    reason_codes.append(LivenessReasonCodes.HEAD_TURN_NOT_DETECTED)
            else:
                reason_codes.append(LivenessReasonCodes.HEAD_TURN_NOT_DETECTED)
        
        confidence = min(1.0, max_turn_angle / (angle_threshold * 2.0))
        
        return ChallengeResult(
            challenge_type=challenge_type,
            passed=turn_detected,
            confidence=confidence if turn_detected else confidence * 0.5,
            detected_at_frame=turn_frame,
            duration_ms=turn_duration,
            reason_codes=reason_codes,
            details={
                "max_turn_angle": max_turn_angle,
                "baseline_yaw": baseline_yaw,
                "angle_threshold": angle_threshold,
            },
        )
    
    def _detect_head_nod(self, sequence: List[LandmarkData]) -> ChallengeResult:
        """
        Detect head nod using pose estimation.
        
        Detects vertical head movement based on pitch angle changes.
        """
        angle_threshold = self.actiVIRTENGINE_config.head_nod_angle_threshold
        min_duration = self.actiVIRTENGINE_config.head_nod_min_duration_ms
        
        # Get pitch angles
        pitch_angles = [lm.pose_pitch for lm in sequence]
        
        if not pitch_angles:
            return ChallengeResult(
                challenge_type=ChallengeType.HEAD_NOD,
                passed=False,
                confidence=0.0,
                reason_codes=[LivenessReasonCodes.HEAD_NOD_NOT_DETECTED],
            )
        
        # Find baseline
        baseline_pitch = np.median(pitch_angles[:min(5, len(pitch_angles))])
        
        # Look for down-up or up-down motion
        nod_detected = False
        nod_frame = None
        max_pitch_change = 0.0
        
        # State machine: neutral -> down -> up -> neutral (or reverse)
        state = "neutral"
        state_start_frame = 0
        
        for i, pitch in enumerate(pitch_angles):
            delta = pitch - baseline_pitch
            abs_delta = abs(delta)
            
            if abs_delta > max_pitch_change:
                max_pitch_change = abs_delta
            
            if state == "neutral":
                if abs_delta >= angle_threshold:
                    state = "moved"
                    state_start_frame = i
            elif state == "moved":
                # Check if returned to neutral
                if abs_delta < angle_threshold * 0.5:
                    start_ts = sequence[state_start_frame].timestamp_ms
                    end_ts = sequence[i].timestamp_ms
                    duration = end_ts - start_ts
                    
                    if duration >= min_duration:
                        nod_detected = True
                        nod_frame = state_start_frame
                        break
                    
                    state = "neutral"
        
        reason_codes = []
        if not nod_detected:
            reason_codes.append(LivenessReasonCodes.HEAD_NOD_NOT_DETECTED)
        
        confidence = min(1.0, max_pitch_change / (angle_threshold * 2.0))
        
        return ChallengeResult(
            challenge_type=ChallengeType.HEAD_NOD,
            passed=nod_detected,
            confidence=confidence if nod_detected else confidence * 0.5,
            detected_at_frame=nod_frame,
            duration_ms=0.0,
            reason_codes=reason_codes,
            details={
                "max_pitch_change": max_pitch_change,
                "angle_threshold": angle_threshold,
            },
        )
    
    def _detect_eyebrow_raise(self, sequence: List[LandmarkData]) -> ChallengeResult:
        """
        Detect eyebrow raise using landmark distances.
        
        Measures the vertical distance between eyebrows and eyes.
        """
        threshold = self.actiVIRTENGINE_config.eyebrow_raise_threshold
        min_duration = self.actiVIRTENGINE_config.eyebrow_raise_min_duration_ms
        
        # Compute eyebrow-eye distances
        distances = []
        
        for landmarks in sequence:
            if landmarks.full_landmarks is not None:
                left_eyebrow = landmarks.full_landmarks[self.LEFT_EYEBROW_INDICES]
                right_eyebrow = landmarks.full_landmarks[self.RIGHT_EYEBROW_INDICES]
                left_eye = landmarks.full_landmarks[self.LEFT_EYE_INDICES]
                right_eye = landmarks.full_landmarks[self.RIGHT_EYE_INDICES]
                
                # Average vertical distance
                left_dist = np.mean(left_eyebrow[:, 1]) - np.mean(left_eye[:, 1])
                right_dist = np.mean(right_eyebrow[:, 1]) - np.mean(right_eye[:, 1])
                avg_dist = (left_dist + right_dist) / 2.0
                distances.append(avg_dist)
            else:
                distances.append(0.0)
        
        if not distances or all(d == 0.0 for d in distances):
            return ChallengeResult(
                challenge_type=ChallengeType.RAISE_EYEBROWS,
                passed=False,
                confidence=0.0,
                reason_codes=[LivenessReasonCodes.EYEBROW_RAISE_NOT_DETECTED],
            )
        
        # Find baseline (initial distance)
        baseline = np.median(distances[:min(5, len(distances))])
        
        if baseline == 0:
            baseline = np.mean([d for d in distances if d != 0])
        
        # Find raise events
        raise_detected = False
        raise_frame = None
        max_change = 0.0
        
        for i, dist in enumerate(distances):
            if baseline != 0:
                relatiVIRTENGINE_change = (dist - baseline) / abs(baseline)
            else:
                relatiVIRTENGINE_change = 0.0
            
            if abs(relatiVIRTENGINE_change) > max_change:
                max_change = abs(relatiVIRTENGINE_change)
            
            # Eyebrow raise increases the distance (more negative in image coords)
            if relatiVIRTENGINE_change > threshold:
                raise_detected = True
                raise_frame = i
                break
        
        reason_codes = []
        if not raise_detected:
            reason_codes.append(LivenessReasonCodes.EYEBROW_RAISE_NOT_DETECTED)
        
        confidence = min(1.0, max_change / (threshold * 2.0))
        
        return ChallengeResult(
            challenge_type=ChallengeType.RAISE_EYEBROWS,
            passed=raise_detected,
            confidence=confidence if raise_detected else confidence * 0.5,
            detected_at_frame=raise_frame,
            duration_ms=0.0,
            reason_codes=reason_codes,
            details={
                "max_change": max_change,
                "threshold": threshold,
                "baseline": baseline,
            },
        )
    
    def get_overall_actiVIRTENGINE_score(
        self,
        results: Dict[ChallengeType, ChallengeResult],
        required_challenges: List[ChallengeType],
        optional_challenges: List[ChallengeType]
    ) -> Tuple[float, bool, List[LivenessReasonCodes]]:
        """
        Compute overall active liveness score from challenge results.
        
        Args:
            results: Challenge detection results.
            required_challenges: Challenges that must pass.
            optional_challenges: Challenges that boost score if passed.
        
        Returns:
            Tuple of (score, passed, reason_codes)
        """
        reason_codes = []
        
        # Check required challenges
        required_passed = 0
        required_total = len(required_challenges)
        
        for challenge in required_challenges:
            if challenge in results and results[challenge].passed:
                required_passed += 1
            else:
                if challenge in results:
                    reason_codes.extend(results[challenge].reason_codes)
        
        # All required must pass
        all_required_passed = required_passed == required_total
        
        # Count optional challenges
        optional_passed = 0
        optional_total = len(optional_challenges)
        
        for challenge in optional_challenges:
            if challenge in results and results[challenge].passed:
                optional_passed += 1
        
        # Compute score
        if required_total > 0:
            required_score = required_passed / required_total
        else:
            required_score = 1.0
        
        if optional_total > 0:
            optional_score = optional_passed / optional_total
        else:
            optional_score = 1.0
        
        # Weight required higher
        total_score = (required_score * 0.7) + (optional_score * 0.3)
        
        # Add bonus for passing all challenges
        if all_required_passed and optional_passed == optional_total:
            total_score = min(1.0, total_score + 0.05)
        
        return total_score, all_required_passed, reason_codes
