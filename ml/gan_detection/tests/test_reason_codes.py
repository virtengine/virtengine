"""
Tests for GAN detection reason codes.

VE-923: Test reason codes functionality.
"""

import pytest

from ml.gan_detection.reason_codes import (
    GANReasonCodes,
    ReasonCodeDetails,
    REASON_CODE_DESCRIPTIONS,
    aggregate_reason_codes,
)


class TestGANReasonCodes:
    """Tests for GANReasonCodes enum."""
    
    def test_success_codes(self):
        """Test success reason codes."""
        assert GANReasonCodes.IMAGE_AUTHENTIC.value == "IMAGE_AUTHENTIC"
        assert GANReasonCodes.HIGH_CONFIDENCE_REAL.value == "HIGH_CONFIDENCE_REAL"
        assert GANReasonCodes.ALL_CHECKS_PASSED.value == "ALL_CHECKS_PASSED"
    
    def test_gan_detection_codes(self):
        """Test GAN detection reason codes."""
        assert GANReasonCodes.GAN_DETECTED.value == "GAN_DETECTED"
        assert GANReasonCodes.GAN_HIGH_CONFIDENCE.value == "GAN_HIGH_CONFIDENCE"
        assert GANReasonCodes.GAN_LOW_CONFIDENCE.value == "GAN_LOW_CONFIDENCE"
        assert GANReasonCodes.SYNTHETIC_IMAGE_DETECTED.value == "SYNTHETIC_IMAGE_DETECTED"
    
    def test_deepfake_codes(self):
        """Test deepfake detection reason codes."""
        assert GANReasonCodes.DEEPFAKE_DETECTED.value == "DEEPFAKE_DETECTED"
        assert GANReasonCodes.FACESWAP_DETECTED.value == "FACESWAP_DETECTED"
        assert GANReasonCodes.EXPRESSION_MANIPULATION_DETECTED.value == "EXPRESSION_MANIPULATION_DETECTED"
        assert GANReasonCodes.MORPHED_FACE_DETECTED.value == "MORPHED_FACE_DETECTED"
    
    def test_artifact_codes(self):
        """Test artifact detection reason codes."""
        assert GANReasonCodes.FREQUENCY_ANOMALY.value == "FREQUENCY_ANOMALY"
        assert GANReasonCodes.CHECKERBOARD_ARTIFACT.value == "CHECKERBOARD_ARTIFACT"
        assert GANReasonCodes.BLENDING_BOUNDARY_DETECTED.value == "BLENDING_BOUNDARY_DETECTED"
        assert GANReasonCodes.TEXTURE_INCONSISTENCY.value == "TEXTURE_INCONSISTENCY"
    
    def test_quality_codes(self):
        """Test quality issue reason codes."""
        assert GANReasonCodes.IMAGE_TOO_SMALL.value == "IMAGE_TOO_SMALL"
        assert GANReasonCodes.IMAGE_TOO_LARGE.value == "IMAGE_TOO_LARGE"
        assert GANReasonCodes.FACE_NOT_DETECTED.value == "FACE_NOT_DETECTED"
    
    def test_system_codes(self):
        """Test system issue reason codes."""
        assert GANReasonCodes.PROCESSING_ERROR.value == "PROCESSING_ERROR"
        assert GANReasonCodes.MODEL_ERROR.value == "MODEL_ERROR"
        assert GANReasonCodes.INVALID_INPUT.value == "INVALID_INPUT"


class TestGetCategory:
    """Tests for get_category method."""
    
    def test_success_category(self):
        """Test success codes are categorized correctly."""
        assert GANReasonCodes.get_category(GANReasonCodes.IMAGE_AUTHENTIC) == "success"
        assert GANReasonCodes.get_category(GANReasonCodes.HIGH_CONFIDENCE_REAL) == "success"
    
    def test_gan_detection_category(self):
        """Test GAN detection codes are categorized correctly."""
        assert GANReasonCodes.get_category(GANReasonCodes.GAN_DETECTED) == "gan_detection"
        assert GANReasonCodes.get_category(GANReasonCodes.SYNTHETIC_IMAGE_DETECTED) == "gan_detection"
    
    def test_deepfake_category(self):
        """Test deepfake codes are categorized correctly."""
        assert GANReasonCodes.get_category(GANReasonCodes.DEEPFAKE_DETECTED) == "deepfake_detection"
        assert GANReasonCodes.get_category(GANReasonCodes.FACESWAP_DETECTED) == "deepfake_detection"
    
    def test_artifact_category(self):
        """Test artifact codes are categorized correctly."""
        assert GANReasonCodes.get_category(GANReasonCodes.FREQUENCY_ANOMALY) == "artifact_detection"
        assert GANReasonCodes.get_category(GANReasonCodes.CHECKERBOARD_ARTIFACT) == "artifact_detection"
    
    def test_quality_category(self):
        """Test quality codes are categorized correctly."""
        assert GANReasonCodes.get_category(GANReasonCodes.IMAGE_TOO_SMALL) == "quality"
        assert GANReasonCodes.get_category(GANReasonCodes.FACE_NOT_DETECTED) == "quality"
    
    def test_system_category(self):
        """Test system codes are categorized correctly."""
        assert GANReasonCodes.get_category(GANReasonCodes.PROCESSING_ERROR) == "system"
        assert GANReasonCodes.get_category(GANReasonCodes.MODEL_ERROR) == "system"


class TestIsRejectionCode:
    """Tests for is_rejection_code method."""
    
    def test_rejection_codes(self):
        """Test rejection codes are identified correctly."""
        assert GANReasonCodes.is_rejection_code(GANReasonCodes.GAN_DETECTED) is True
        assert GANReasonCodes.is_rejection_code(GANReasonCodes.GAN_HIGH_CONFIDENCE) is True
        assert GANReasonCodes.is_rejection_code(GANReasonCodes.DEEPFAKE_DETECTED) is True
        assert GANReasonCodes.is_rejection_code(GANReasonCodes.FACESWAP_DETECTED) is True
    
    def test_non_rejection_codes(self):
        """Test non-rejection codes are identified correctly."""
        assert GANReasonCodes.is_rejection_code(GANReasonCodes.IMAGE_AUTHENTIC) is False
        assert GANReasonCodes.is_rejection_code(GANReasonCodes.FREQUENCY_ANOMALY) is False
        assert GANReasonCodes.is_rejection_code(GANReasonCodes.IMAGE_TOO_SMALL) is False


class TestGetSeverity:
    """Tests for get_severity method."""
    
    def test_critical_severity(self):
        """Test critical codes have severity 3."""
        assert GANReasonCodes.get_severity(GANReasonCodes.GAN_HIGH_CONFIDENCE) == 3
        assert GANReasonCodes.get_severity(GANReasonCodes.DEEPFAKE_DETECTED) == 3
        assert GANReasonCodes.get_severity(GANReasonCodes.FACESWAP_DETECTED) == 3
    
    def test_error_severity(self):
        """Test error codes have severity 2."""
        assert GANReasonCodes.get_severity(GANReasonCodes.GAN_DETECTED) == 2
        assert GANReasonCodes.get_severity(GANReasonCodes.SYNTHETIC_IMAGE_DETECTED) == 2
    
    def test_warning_severity(self):
        """Test warning codes have severity 1."""
        assert GANReasonCodes.get_severity(GANReasonCodes.GAN_LOW_CONFIDENCE) == 1
        assert GANReasonCodes.get_severity(GANReasonCodes.FREQUENCY_ANOMALY) == 1
    
    def test_info_severity(self):
        """Test informational codes have severity 0."""
        assert GANReasonCodes.get_severity(GANReasonCodes.IMAGE_AUTHENTIC) == 0
        assert GANReasonCodes.get_severity(GANReasonCodes.IMAGE_TOO_SMALL) == 0


class TestReasonCodeDetails:
    """Tests for ReasonCodeDetails dataclass."""
    
    def test_to_dict(self):
        """Test converting details to dict."""
        details = ReasonCodeDetails(
            code=GANReasonCodes.GAN_DETECTED,
            description="GAN-generated content detected",
            confidence=0.85,
            severity=2,
            details={"score": 0.85, "threshold": 0.6},
        )
        
        d = details.to_dict()
        
        assert d["code"] == "GAN_DETECTED"
        assert d["description"] == "GAN-generated content detected"
        assert d["confidence"] == 0.85
        assert d["severity"] == 2
        assert d["details"]["score"] == 0.85


class TestReasonCodeDescriptions:
    """Tests for reason code descriptions."""
    
    def test_all_codes_have_descriptions(self):
        """Test all reason codes have descriptions."""
        for code in GANReasonCodes:
            assert code in REASON_CODE_DESCRIPTIONS, f"Missing description for {code}"
    
    def test_descriptions_are_strings(self):
        """Test all descriptions are non-empty strings."""
        for code, description in REASON_CODE_DESCRIPTIONS.items():
            assert isinstance(description, str)
            assert len(description) > 0


class TestAggregateReasonCodes:
    """Tests for aggregate_reason_codes function."""
    
    def test_aggregate_single_code(self):
        """Test aggregating single code."""
        codes = [
            ReasonCodeDetails(
                code=GANReasonCodes.GAN_DETECTED,
                description="test",
                confidence=0.8,
                severity=2,
                details={},
            )
        ]
        
        result = aggregate_reason_codes(codes)
        
        assert result == ["GAN_DETECTED"]
    
    def test_aggregate_multiple_codes(self):
        """Test aggregating multiple codes."""
        codes = [
            ReasonCodeDetails(
                code=GANReasonCodes.FREQUENCY_ANOMALY,
                description="test",
                confidence=0.5,
                severity=1,
                details={},
            ),
            ReasonCodeDetails(
                code=GANReasonCodes.GAN_HIGH_CONFIDENCE,
                description="test",
                confidence=0.9,
                severity=3,
                details={},
            ),
            ReasonCodeDetails(
                code=GANReasonCodes.GAN_DETECTED,
                description="test",
                confidence=0.8,
                severity=2,
                details={},
            ),
        ]
        
        result = aggregate_reason_codes(codes)
        
        # Should be sorted by severity, then confidence
        assert result[0] == "GAN_HIGH_CONFIDENCE"  # severity 3
        assert result[1] == "GAN_DETECTED"  # severity 2
    
    def test_aggregate_respects_max_codes(self):
        """Test aggregation respects max_codes limit."""
        codes = [
            ReasonCodeDetails(
                code=GANReasonCodes.GAN_DETECTED,
                description="test",
                confidence=0.8,
                severity=2,
                details={},
            )
            for _ in range(10)
        ]
        
        result = aggregate_reason_codes(codes, max_codes=3)
        
        assert len(result) == 3
    
    def test_aggregate_empty_list(self):
        """Test aggregating empty list."""
        result = aggregate_reason_codes([])
        
        assert result == []
    
    def test_aggregate_default_max(self):
        """Test default max_codes is 5."""
        codes = [
            ReasonCodeDetails(
                code=GANReasonCodes.GAN_DETECTED,
                description="test",
                confidence=0.8,
                severity=2,
                details={},
            )
            for _ in range(10)
        ]
        
        result = aggregate_reason_codes(codes)
        
        assert len(result) == 5
