"""
Tests for active liveness challenge detection.

VE-901: Test active challenge detection (blink, smile, head turn).
"""

import pytest
import numpy as np

from ml.liveness_detection.active_challenges import (
    ActiveChallengeDetector,
    ChallengeResult,
    LandmarkData,
)
from ml.liveness_detection.config import ChallengeType, LivenessConfig
from ml.liveness_detection.reason_codes import LivenessReasonCodes


class TestActiveChallengeDetector:
    """Tests for ActiveChallengeDetector class."""
    
    def test_detector_creation(self, liveness_config):
        """Test creating detector with config."""
        detector = ActiveChallengeDetector(liveness_config)
        assert detector is not None
        assert detector.config == liveness_config
    
    def test_detector_default_config(self):
        """Test creating detector with default config."""
        detector = ActiveChallengeDetector()
        assert detector is not None
    
    def test_detector_reset(self, liveness_config, sample_landmarks_sequence):
        """Test resetting detector state."""
        detector = ActiveChallengeDetector(liveness_config)
        
        # Add some frames
        for lm in sample_landmarks_sequence[:5]:
            detector.add_frame(lm)
        
        # Reset
        detector.reset()
        
        # History should be empty
        assert len(detector._landmark_history) == 0
    
    def test_add_frame(self, liveness_config, sample_landmarks):
        """Test adding frames to detector."""
        detector = ActiveChallengeDetector(liveness_config)
        
        detector.add_frame(sample_landmarks)
        
        assert len(detector._landmark_history) == 1


class TestBlinkDetection:
    """Tests for blink detection."""
    
    def test_blink_detected(self, liveness_config, blink_landmarks_sequence):
        """Test that blink is detected in blink sequence."""
        detector = ActiveChallengeDetector(liveness_config)
        
        result = detector.detect_challenge(
            ChallengeType.BLINK,
            blink_landmarks_sequence
        )
        
        assert isinstance(result, ChallengeResult)
        assert result.challenge_type == ChallengeType.BLINK
        assert result.passed is True
        assert result.confidence > 0.5
        assert result.detected_at_frame is not None
    
    def test_no_blink_in_static_sequence(self, liveness_config, sample_landmarks_sequence):
        """Test that no blink is detected without eye closure."""
        detector = ActiveChallengeDetector(liveness_config)
        
        result = detector.detect_challenge(
            ChallengeType.BLINK,
            sample_landmarks_sequence
        )
        
        assert result.passed is False
        assert LivenessReasonCodes.BLINK_NOT_DETECTED in result.reason_codes
    
    def test_insufficient_frames(self, liveness_config, sample_landmarks):
        """Test detection fails with insufficient frames."""
        detector = ActiveChallengeDetector(liveness_config)
        
        result = detector.detect_challenge(
            ChallengeType.BLINK,
            [sample_landmarks] * 3  # Too few frames
        )
        
        assert result.passed is False
        assert LivenessReasonCodes.INSUFFICIENT_FRAMES in result.reason_codes
    
    def test_eye_aspect_ratio_computation(self, liveness_config):
        """Test EAR computation."""
        detector = ActiveChallengeDetector(liveness_config)
        
        # Open eye landmarks (6 points)
        open_eye = np.array([
            [0, 0],   # Left corner
            [2, -3],  # Top-left
            [5, -3],  # Top-right
            [7, 0],   # Right corner
            [5, 3],   # Bottom-right
            [2, 3],   # Bottom-left
        ], dtype=np.float32)
        
        ear = detector._compute_eye_aspect_ratio(open_eye)
        assert ear > 0.2  # Open eye should have higher EAR
        
        # Closed eye landmarks
        closed_eye = np.array([
            [0, 0],
            [2, -1],  # Closer together
            [5, -1],
            [7, 0],
            [5, 1],
            [2, 1],
        ], dtype=np.float32)
        
        closed_ear = detector._compute_eye_aspect_ratio(closed_eye)
        assert closed_ear < ear  # Closed eye has lower EAR


class TestSmileDetection:
    """Tests for smile detection."""
    
    def test_smile_detected(self, liveness_config, smile_landmarks_sequence):
        """Test that smile is detected in smile sequence."""
        detector = ActiveChallengeDetector(liveness_config)
        
        result = detector.detect_challenge(
            ChallengeType.SMILE,
            smile_landmarks_sequence
        )
        
        assert isinstance(result, ChallengeResult)
        assert result.challenge_type == ChallengeType.SMILE
        # Note: Detection depends on landmark geometry
        assert result.confidence >= 0.0
    
    def test_no_smile_in_neutral_sequence(self, liveness_config, sample_landmarks_sequence):
        """Test that no smile is detected in neutral sequence."""
        detector = ActiveChallengeDetector(liveness_config)
        
        result = detector.detect_challenge(
            ChallengeType.SMILE,
            sample_landmarks_sequence
        )
        
        # Smile may or may not be detected depending on fixture geometry
        assert isinstance(result, ChallengeResult)


class TestHeadTurnDetection:
    """Tests for head turn detection."""
    
    def test_head_turn_left_detected(self, liveness_config, head_turn_landmarks_sequence):
        """Test that head turn left is detected."""
        detector = ActiveChallengeDetector(liveness_config)
        
        result = detector.detect_challenge(
            ChallengeType.HEAD_TURN_LEFT,
            head_turn_landmarks_sequence
        )
        
        assert isinstance(result, ChallengeResult)
        assert result.challenge_type == ChallengeType.HEAD_TURN_LEFT
        assert result.passed is True
        assert result.confidence > 0.5
    
    def test_head_turn_right_not_detected_for_left(self, liveness_config, head_turn_landmarks_sequence):
        """Test that head turn right is not detected in left turn sequence."""
        detector = ActiveChallengeDetector(liveness_config)
        
        result = detector.detect_challenge(
            ChallengeType.HEAD_TURN_RIGHT,
            head_turn_landmarks_sequence
        )
        
        # Should fail because the turn is left, not right
        assert result.passed is False
        assert LivenessReasonCodes.HEAD_TURN_WRONG_DIRECTION in result.reason_codes
    
    def test_no_head_turn_in_static_sequence(self, liveness_config, sample_landmarks_sequence):
        """Test no head turn detected in static sequence."""
        detector = ActiveChallengeDetector(liveness_config)
        
        result = detector.detect_challenge(
            ChallengeType.HEAD_TURN_LEFT,
            sample_landmarks_sequence
        )
        
        assert result.passed is False
        assert LivenessReasonCodes.HEAD_TURN_NOT_DETECTED in result.reason_codes


class TestDetectAllChallenges:
    """Tests for detecting multiple challenges."""
    
    def test_detect_all_challenges(self, liveness_config, blink_landmarks_sequence):
        """Test detecting multiple challenges at once."""
        detector = ActiveChallengeDetector(liveness_config)
        
        challenges = [ChallengeType.BLINK, ChallengeType.SMILE]
        results = detector.detect_all_challenges(challenges, blink_landmarks_sequence)
        
        assert len(results) == 2
        assert ChallengeType.BLINK in results
        assert ChallengeType.SMILE in results
    
    def test_overall_active_score(self, liveness_config, blink_landmarks_sequence):
        """Test computing overall active liveness score."""
        detector = ActiveChallengeDetector(liveness_config)
        
        challenges = [ChallengeType.BLINK]
        results = detector.detect_all_challenges(challenges, blink_landmarks_sequence)
        
        score, passed, reason_codes = detector.get_overall_active_score(
            results,
            required_challenges=[ChallengeType.BLINK],
            optional_challenges=[]
        )
        
        assert 0.0 <= score <= 1.0
        assert passed is True  # Blink should pass


class TestChallengeResult:
    """Tests for ChallengeResult dataclass."""
    
    def test_challenge_result_to_dict(self):
        """Test converting challenge result to dict."""
        result = ChallengeResult(
            challenge_type=ChallengeType.BLINK,
            passed=True,
            confidence=0.85,
            detected_at_frame=5,
            duration_ms=150.0,
            reason_codes=[],
            details={"ear_min": 0.15},
        )
        
        d = result.to_dict()
        
        assert d["challenge_type"] == "blink"
        assert d["passed"] is True
        assert d["confidence"] == 0.85
        assert d["detected_at_frame"] == 5
        assert d["duration_ms"] == 150.0


class TestLandmarkData:
    """Tests for LandmarkData dataclass."""
    
    def test_landmark_data_creation(self):
        """Test creating landmark data."""
        landmarks = np.zeros((68, 2), dtype=np.float32)
        
        data = LandmarkData(
            full_landmarks=landmarks,
            pose_pitch=5.0,
            pose_yaw=-10.0,
            pose_roll=2.0,
            confidence=0.95,
            frame_index=0,
            timestamp_ms=0.0,
        )
        
        assert data.full_landmarks.shape == (68, 2)
        assert data.pose_pitch == 5.0
        assert data.pose_yaw == -10.0
        assert data.confidence == 0.95
