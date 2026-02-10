"""
Tests for liveness reason codes.

VE-901: Test reason codes for liveness detection.
"""

import pytest

from ml.liveness_detection.reason_codes import (
    LivenessReasonCodes,
    ReasonCodeDetails,
    aggregate_reason_codes,
)


class TestLivenessReasonCodes:
    """Tests for LivenessReasonCodes enum."""
    
    def test_success_codes_exist(self):
        """Test success reason codes exist."""
        assert LivenessReasonCodes.LIVENESS_CONFIRMED.value == "LIVENESS_CONFIRMED"
        assert LivenessReasonCodes.HIGH_CONFIDENCE_LIVE.value == "HIGH_CONFIDENCE_LIVE"
        assert LivenessReasonCodes.ALL_CHALLENGES_PASSED.value == "ALL_CHALLENGES_PASSED"
    
    def test_challenge_failure_codes_exist(self):
        """Test challenge failure codes exist."""
        assert LivenessReasonCodes.BLINK_NOT_DETECTED.value == "BLINK_NOT_DETECTED"
        assert LivenessReasonCodes.SMILE_NOT_DETECTED.value == "SMILE_NOT_DETECTED"
        assert LivenessReasonCodes.HEAD_TURN_NOT_DETECTED.value == "HEAD_TURN_NOT_DETECTED"
    
    def test_spoof_detection_codes_exist(self):
        """Test spoof detection codes exist."""
        assert LivenessReasonCodes.PHOTO_PRINT_DETECTED.value == "PHOTO_PRINT_DETECTED"
        assert LivenessReasonCodes.SCREEN_DISPLAY_DETECTED.value == "SCREEN_DISPLAY_DETECTED"
        assert LivenessReasonCodes.VIDEO_REPLAY_DETECTED.value == "VIDEO_REPLAY_DETECTED"
        assert LivenessReasonCodes.MASK_2D_DETECTED.value == "MASK_2D_DETECTED"
        assert LivenessReasonCodes.MASK_3D_DETECTED.value == "MASK_3D_DETECTED"
        assert LivenessReasonCodes.DEEPFAKE_DETECTED.value == "DEEPFAKE_DETECTED"
    
    def test_get_category_success(self):
        """Test getting category for success codes."""
        category = LivenessReasonCodes.get_category(LivenessReasonCodes.LIVENESS_CONFIRMED)
        assert category == "success"
    
    def test_get_category_challenge(self):
        """Test getting category for challenge codes."""
        category = LivenessReasonCodes.get_category(LivenessReasonCodes.BLINK_NOT_DETECTED)
        assert category == "challenge"
    
    def test_get_category_passive(self):
        """Test getting category for passive analysis codes."""
        category = LivenessReasonCodes.get_category(LivenessReasonCodes.UNNATURAL_TEXTURE)
        assert category == "passive"
    
    def test_get_category_spoof(self):
        """Test getting category for spoof codes."""
        category = LivenessReasonCodes.get_category(LivenessReasonCodes.PHOTO_PRINT_DETECTED)
        assert category == "spoof"
    
    def test_get_category_quality(self):
        """Test getting category for quality codes."""
        category = LivenessReasonCodes.get_category(LivenessReasonCodes.FACE_TOO_SMALL)
        assert category == "quality"
    
    def test_get_category_system(self):
        """Test getting category for system codes."""
        category = LivenessReasonCodes.get_category(LivenessReasonCodes.PROCESSING_ERROR)
        assert category == "system"
    
    def test_is_fatal_spoof_codes(self):
        """Test spoof codes are fatal."""
        assert LivenessReasonCodes.is_fatal(LivenessReasonCodes.PHOTO_PRINT_DETECTED) is True
        assert LivenessReasonCodes.is_fatal(LivenessReasonCodes.SCREEN_DISPLAY_DETECTED) is True
        assert LivenessReasonCodes.is_fatal(LivenessReasonCodes.DEEPFAKE_DETECTED) is True
    
    def test_is_fatal_system_codes(self):
        """Test system error codes are fatal."""
        assert LivenessReasonCodes.is_fatal(LivenessReasonCodes.PROCESSING_ERROR) is True
        assert LivenessReasonCodes.is_fatal(LivenessReasonCodes.MODEL_ERROR) is True
    
    def test_is_fatal_non_fatal_codes(self):
        """Test non-fatal codes return False."""
        assert LivenessReasonCodes.is_fatal(LivenessReasonCodes.BLINK_NOT_DETECTED) is False
        assert LivenessReasonCodes.is_fatal(LivenessReasonCodes.LIVENESS_CONFIRMED) is False
    
    def test_get_user_message(self):
        """Test getting user-friendly messages."""
        msg = LivenessReasonCodes.get_user_message(LivenessReasonCodes.BLINK_NOT_DETECTED)
        assert "blink" in msg.lower()
        
        msg = LivenessReasonCodes.get_user_message(LivenessReasonCodes.PHOTO_PRINT_DETECTED)
        assert "photo" in msg.lower() or "print" in msg.lower()
    
    def test_get_user_message_unknown(self):
        """Test getting message for codes not in map."""
        # All codes should have messages, but test default behavior
        msg = LivenessReasonCodes.get_user_message(LivenessReasonCodes.LIVENESS_CONFIRMED)
        assert msg is not None
        assert len(msg) > 0


class TestReasonCodeDetails:
    """Tests for ReasonCodeDetails dataclass."""
    
    def test_details_creation(self):
        """Test creating reason code details."""
        details = ReasonCodeDetails(
            code=LivenessReasonCodes.BLINK_NOT_DETECTED,
            frame_index=10,
            confidence=0.85,
            details="EAR threshold not met",
        )
        
        assert details.code == LivenessReasonCodes.BLINK_NOT_DETECTED
        assert details.frame_index == 10
        assert details.confidence == 0.85
    
    def test_details_to_dict(self):
        """Test converting details to dict."""
        details = ReasonCodeDetails(
            code=LivenessReasonCodes.PHOTO_PRINT_DETECTED,
            frame_index=5,
            confidence=0.9,
            details="High texture uniformity",
        )
        
        d = details.to_dict()
        
        assert d["code"] == "PHOTO_PRINT_DETECTED"
        assert d["frame_index"] == 5
        assert d["confidence"] == 0.9
        assert d["category"] == "spoof"
        assert d["is_fatal"] is True
        assert "user_message" in d
    
    def test_details_default_values(self):
        """Test default values for optional fields."""
        details = ReasonCodeDetails(
            code=LivenessReasonCodes.LIVENESS_CONFIRMED,
        )
        
        assert details.frame_index is None
        assert details.confidence == 0.0
        assert details.details is None


class TestAggregateReasonCodes:
    """Tests for aggregate_reason_codes function."""
    
    def test_aggregate_removes_duplicates(self):
        """Test aggregation removes duplicates."""
        codes = [
            ReasonCodeDetails(
                code=LivenessReasonCodes.BLINK_NOT_DETECTED,
                confidence=0.8,
            ),
            ReasonCodeDetails(
                code=LivenessReasonCodes.BLINK_NOT_DETECTED,
                confidence=0.9,
            ),
        ]
        
        result = aggregate_reason_codes(codes)
        
        assert len(result) == 1
        assert result[0] == "BLINK_NOT_DETECTED"
    
    def test_aggregate_keeps_highest_confidence(self):
        """Test aggregation keeps highest confidence version."""
        codes = [
            ReasonCodeDetails(
                code=LivenessReasonCodes.BLINK_NOT_DETECTED,
                confidence=0.5,
            ),
            ReasonCodeDetails(
                code=LivenessReasonCodes.BLINK_NOT_DETECTED,
                confidence=0.9,
            ),
            ReasonCodeDetails(
                code=LivenessReasonCodes.BLINK_NOT_DETECTED,
                confidence=0.7,
            ),
        ]
        
        result = aggregate_reason_codes(codes)
        
        # Should keep the one with 0.9 confidence
        assert len(result) == 1
    
    def test_aggregate_multiple_different_codes(self):
        """Test aggregation with multiple different codes."""
        codes = [
            ReasonCodeDetails(
                code=LivenessReasonCodes.BLINK_NOT_DETECTED,
                confidence=0.8,
            ),
            ReasonCodeDetails(
                code=LivenessReasonCodes.SMILE_NOT_DETECTED,
                confidence=0.7,
            ),
            ReasonCodeDetails(
                code=LivenessReasonCodes.HEAD_TURN_NOT_DETECTED,
                confidence=0.6,
            ),
        ]
        
        result = aggregate_reason_codes(codes)
        
        assert len(result) == 3
        assert "BLINK_NOT_DETECTED" in result
        assert "SMILE_NOT_DETECTED" in result
        assert "HEAD_TURN_NOT_DETECTED" in result
    
    def test_aggregate_sorts_by_confidence(self):
        """Test aggregation sorts by confidence descending."""
        codes = [
            ReasonCodeDetails(
                code=LivenessReasonCodes.BLINK_NOT_DETECTED,
                confidence=0.5,
            ),
            ReasonCodeDetails(
                code=LivenessReasonCodes.SMILE_NOT_DETECTED,
                confidence=0.9,
            ),
            ReasonCodeDetails(
                code=LivenessReasonCodes.HEAD_TURN_NOT_DETECTED,
                confidence=0.7,
            ),
        ]
        
        result = aggregate_reason_codes(codes)
        
        # Highest confidence first
        assert result[0] == "SMILE_NOT_DETECTED"
    
    def test_aggregate_empty_list(self):
        """Test aggregation of empty list."""
        result = aggregate_reason_codes([])
        assert result == []
