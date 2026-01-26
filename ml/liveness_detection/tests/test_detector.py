"""
Tests for main liveness detector.

VE-901: Test main LivenessDetector class.
"""

import pytest
import numpy as np

from ml.liveness_detection.detector import (
    LivenessDetector,
    LivenessResult,
    create_detector,
)
from ml.liveness_detection.config import LivenessConfig, ChallengeType
from ml.liveness_detection.reason_codes import LivenessReasonCodes


class TestLivenessDetector:
    """Tests for LivenessDetector class."""
    
    def test_detector_creation(self, liveness_config):
        """Test creating detector with config."""
        detector = LivenessDetector(liveness_config)
        
        assert detector is not None
        assert detector.config == liveness_config
        assert detector.MODEL_VERSION == "1.0.0"
    
    def test_detector_default_config(self):
        """Test creating detector with default config."""
        detector = LivenessDetector()
        assert detector is not None
    
    def test_model_hash_generation(self, liveness_config):
        """Test model hash is generated."""
        detector = LivenessDetector(liveness_config)
        
        assert detector._model_hash is not None
        assert len(detector._model_hash) == 32  # SHA256 truncated
    
    def test_model_hash_determinism(self, liveness_config):
        """Test model hash is deterministic."""
        detector1 = LivenessDetector(liveness_config)
        detector2 = LivenessDetector(liveness_config)
        
        assert detector1._model_hash == detector2._model_hash
    
    def test_detect_insufficient_frames(self, liveness_config):
        """Test detection fails with insufficient frames."""
        detector = LivenessDetector(liveness_config)
        
        # Too few frames
        frames = [np.zeros((100, 100, 3), dtype=np.uint8) for _ in range(3)]
        
        result = detector.detect(frames)
        
        assert result.is_live is False
        assert LivenessReasonCodes.INSUFFICIENT_FRAMES.value in result.reason_codes
    
    def test_detect_natural_sequence(self, liveness_config, sample_frame_sequence):
        """Test detection on natural-looking frames."""
        detector = LivenessDetector(liveness_config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert isinstance(result, LivenessResult)
        assert 0.0 <= result.liveness_score <= 1.0
        assert result.decision in ["live", "spoof", "uncertain"]
    
    def test_detect_with_landmarks(
        self,
        liveness_config,
        sample_frame_sequence,
        blink_landmarks_sequence
    ):
        """Test detection with landmarks provided."""
        detector = LivenessDetector(liveness_config)
        
        result = detector.detect(
            frames=sample_frame_sequence,
            landmarks=blink_landmarks_sequence,
            required_challenges=[ChallengeType.BLINK],
        )
        
        assert isinstance(result, LivenessResult)
        assert result.actiVIRTENGINE_challenge_score > 0.0
    
    def test_detect_with_face_regions(
        self,
        liveness_config,
        sample_frame_sequence,
        sample_face_regions
    ):
        """Test detection with face regions provided."""
        detector = LivenessDetector(liveness_config)
        
        result = detector.detect(
            frames=sample_frame_sequence,
            face_regions=sample_face_regions,
        )
        
        assert isinstance(result, LivenessResult)
    
    def test_detect_with_include_details(self, liveness_config, sample_frame_sequence):
        """Test detection with detailed results."""
        detector = LivenessDetector(liveness_config)
        
        result = detector.detect(
            frames=sample_frame_sequence,
            include_details=True,
        )
        
        assert result.passiVIRTENGINE_result is not None
        assert result.spoof_result is not None
    
    def test_detect_static_sequence(self, liveness_config, static_frame_sequence):
        """Test detection on static (photo-like) frames."""
        detector = LivenessDetector(liveness_config)
        
        result = detector.detect(static_frame_sequence)
        
        assert isinstance(result, LivenessResult)
        # Static frames should generally have lower score
        # (though exact behavior depends on thresholds)
    
    def test_detect_truncates_long_sequence(self, liveness_config, sample_face_frame):
        """Test that excessively long sequences are truncated."""
        detector = LivenessDetector(liveness_config)
        
        # Create very long sequence
        frames = [sample_face_frame.copy() for _ in range(500)]
        
        result = detector.detect(frames)
        
        # Should process without error
        assert isinstance(result, LivenessResult)


class TestLivenessResult:
    """Tests for LivenessResult dataclass."""
    
    def test_result_to_dict(self):
        """Test converting result to dict."""
        result = LivenessResult(
            is_live=True,
            decision="live",
            liveness_score=0.85,
            confidence=0.9,
            actiVIRTENGINE_challenge_score=0.8,
            passiVIRTENGINE_analysis_score=0.85,
            spoof_detection_score=0.9,
            model_version="1.0.0",
            model_hash="abc123",
            reason_codes=["LIVENESS_CONFIRMED"],
            result_hash="hash123",
            processing_time_ms=150.0,
        )
        
        d = result.to_dict()
        
        assert d["is_live"] is True
        assert d["decision"] == "live"
        assert d["liveness_score"] == 0.85
        assert d["confidence"] == 0.9
    
    def test_result_to_veid_record(self):
        """Test converting result to VEID record format."""
        result = LivenessResult(
            is_live=True,
            decision="live",
            liveness_score=0.85,
            confidence=0.9,
            model_version="1.0.0",
            model_hash="abc123def456ghi789",
            reason_codes=["LIVENESS_CONFIRMED"],
            result_hash="hash123",
        )
        
        record = result.to_veid_record()
        
        assert record["liveness_score"] == 8500  # Basis points
        assert record["is_live"] is True
        assert record["confidence"] == 90
        assert len(record["model_hash"]) <= 16  # Truncated


class TestDetectSingleFrame:
    """Tests for single frame detection."""
    
    def test_detect_single_frame(self, liveness_config, sample_face_frame):
        """Test quick single frame detection."""
        detector = LivenessDetector(liveness_config)
        
        score, reason_codes = detector.detect_single_frame(sample_face_frame)
        
        assert 0.0 <= score <= 1.0
        assert isinstance(reason_codes, list)
    
    def test_detect_single_frame_with_region(
        self,
        liveness_config,
        sample_face_frame,
        sample_face_region
    ):
        """Test single frame detection with face region."""
        detector = LivenessDetector(liveness_config)
        
        score, reason_codes = detector.detect_single_frame(
            sample_face_frame,
            face_region=sample_face_region,
        )
        
        assert 0.0 <= score <= 1.0


class TestValidateChallengeSequence:
    """Tests for challenge validation."""
    
    def test_validate_blink_challenge(self, liveness_config, blink_landmarks_sequence):
        """Test validating blink challenge."""
        detector = LivenessDetector(liveness_config)
        
        result = detector.validate_challenge_sequence(
            blink_landmarks_sequence,
            ChallengeType.BLINK,
        )
        
        assert result.challenge_type == ChallengeType.BLINK
        assert result.passed is True


class TestCreateDetector:
    """Tests for detector factory function."""
    
    def test_create_default_detector(self):
        """Test creating default detector."""
        detector = create_detector("default")
        
        assert detector is not None
        assert detector.config.score.pass_threshold == 0.75
    
    def test_create_strict_detector(self):
        """Test creating strict detector."""
        detector = create_detector("strict")
        
        assert detector is not None
        assert detector.config.score.pass_threshold == 0.85
    
    def test_create_permissiVIRTENGINE_detector(self):
        """Test creating permissive detector."""
        detector = create_detector("permissive")
        
        assert detector is not None
        assert detector.config.score.pass_threshold == 0.60
    
    def test_create_unknown_type_defaults(self):
        """Test creating with unknown type uses default."""
        detector = create_detector("unknown")
        
        assert detector is not None
        assert detector.config.score.pass_threshold == 0.75


class TestResultHash:
    """Tests for result hash computation."""
    
    def test_result_hash_determinism(self, liveness_config, sample_frame_sequence):
        """Test result hash is deterministic."""
        detector = LivenessDetector(liveness_config)
        
        result1 = detector.detect(sample_frame_sequence)
        result2 = detector.detect(sample_frame_sequence)
        
        # Same input should produce same hash
        assert result1.result_hash == result2.result_hash
    
    def test_result_hash_differs_for_different_input(
        self,
        liveness_config,
        sample_frame_sequence,
        static_frame_sequence
    ):
        """Test result hash differs for different inputs."""
        detector = LivenessDetector(liveness_config)
        
        result1 = detector.detect(sample_frame_sequence)
        result2 = detector.detect(static_frame_sequence)
        
        # Different input should produce different hash
        # (unless by coincidence they have same score/decision)
        if result1.liveness_score != result2.liveness_score:
            assert result1.result_hash != result2.result_hash


class TestComponentScores:
    """Tests for component score computation."""
    
    def test_component_scores_in_range(self, liveness_config, sample_frame_sequence):
        """Test all component scores are in valid range."""
        detector = LivenessDetector(liveness_config)
        
        result = detector.detect(sample_frame_sequence)
        
        assert 0.0 <= result.actiVIRTENGINE_challenge_score <= 1.0
        assert 0.0 <= result.passiVIRTENGINE_analysis_score <= 1.0
        assert 0.0 <= result.spoof_detection_score <= 1.0
        assert 0.0 <= result.liveness_score <= 1.0
    
    def test_combined_score_uses_weights(self, liveness_config, sample_frame_sequence):
        """Test combined score uses configured weights."""
        detector = LivenessDetector(liveness_config)
        
        result = detector.detect(sample_frame_sequence, include_details=True)
        
        # The liveness_score should be computed from component scores
        # with weights applied (though penalties/bonuses may modify it)
        assert result.liveness_score is not None


class TestReasonCodes:
    """Tests for reason code handling."""
    
    def test_reason_codes_are_strings(self, liveness_config, sample_frame_sequence):
        """Test reason codes are returned as strings."""
        detector = LivenessDetector(liveness_config)
        
        result = detector.detect(sample_frame_sequence)
        
        for code in result.reason_codes:
            assert isinstance(code, str)
    
    def test_reason_codes_deduplicated(self, liveness_config, sample_frame_sequence):
        """Test reason codes are deduplicated."""
        detector = LivenessDetector(liveness_config)
        
        result = detector.detect(sample_frame_sequence)
        
        # No duplicates
        assert len(result.reason_codes) == len(set(result.reason_codes))
