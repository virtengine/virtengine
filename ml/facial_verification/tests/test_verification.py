"""
Tests for the verification module.
"""

import pytest
import numpy as np

from ml.facial_verification.verification import (
    FaceVerifier,
    VerificationResult,
    VerificationDecisionEngine,
)
from ml.facial_verification.config import VerificationConfig, ModelName
from ml.facial_verification.reason_codes import ReasonCodes
from ml.facial_verification.tests.conftest import create_embedding_pair


class TestVerificationResult:
    """Tests for VerificationResult dataclass."""
    
    def test_match_result(self):
        """Test creating a match result."""
        result = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            reason_codes=["MATCH_CONFIRMED"],
        )
        
        assert result.match is True
        assert result.decision == "match"
        assert result.similarity_score == 0.95
    
    def test_no_match_result(self):
        """Test creating a no-match result."""
        result = VerificationResult(
            match=False,
            decision="no_match",
            similarity_score=0.40,
            confidence_percent=40.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            reason_codes=["LOW_SIMILARITY_SCORE"],
        )
        
        assert result.match is False
        assert result.decision == "no_match"
        assert "LOW_SIMILARITY_SCORE" in result.reason_codes
    
    def test_borderline_result(self):
        """Test creating a borderline result."""
        result = VerificationResult(
            match=False,
            decision="borderline",
            similarity_score=0.87,
            confidence_percent=87.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
            reason_codes=["BORDERLINE_MATCH"],
        )
        
        assert result.match is False
        assert result.decision == "borderline"
    
    def test_to_dict(self):
        """Test serialization to dict."""
        result = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95,
            confidence_percent=95.0,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123",
        )
        
        d = result.to_dict()
        
        assert d["match"] is True
        assert d["decision"] == "match"
        assert d["similarity_score"] == 0.95
        assert d["model_name"] == "VGG-Face"
    
    def test_to_verification_record(self):
        """Test conversion to verification record for on-chain storage."""
        result = VerificationResult(
            match=True,
            decision="match",
            similarity_score=0.95123456,
            confidence_percent=95.12345,
            model_name="VGG-Face",
            model_version="1.0.0",
            model_hash="abc123def456",
            embeddings_hash="xyz789abc012",
            reason_codes=["MATCH_CONFIRMED", "HIGH_CONFIDENCE_MATCH"],
        )
        
        record = result.to_verification_record()
        
        # Should be truncated for storage
        assert len(record["model_hash"]) == 16
        assert len(record["embeddings_hash"]) == 16
        # Should be rounded
        assert record["similarity_score"] == 0.9512


class TestFaceVerifier:
    """Tests for FaceVerifier class."""
    
    def test_init_default_config(self):
        """Test initialization with default config."""
        verifier = FaceVerifier()
        assert verifier.config is not None
    
    def test_init_custom_config(self, verification_config):
        """Test initialization with custom config."""
        verifier = FaceVerifier(verification_config)
        assert verifier.config == verification_config
    
    def test_get_model_info(self, verifier):
        """Test getting model information."""
        info = verifier.get_model_info()
        
        assert "model_name" in info
        assert "model_version" in info
        assert "thresholds" in info
        assert "match" in info["thresholds"]
        assert "borderline" in info["thresholds"]
        assert "reject" in info["thresholds"]


class TestVerificationDecisionEngine:
    """Tests for VerificationDecisionEngine."""
    
    def test_evaluate_match(self, verification_config):
        """Test evaluation of a match."""
        engine = VerificationDecisionEngine(verification_config)
        
        decision, confidence, codes = engine.evaluate(0.95)
        
        assert decision == "match"
        assert ReasonCodes.MATCH_CONFIRMED in codes
    
    def test_evaluate_no_match(self, verification_config):
        """Test evaluation of no match."""
        engine = VerificationDecisionEngine(verification_config)
        
        decision, confidence, codes = engine.evaluate(0.40)
        
        assert decision == "no_match"
        assert ReasonCodes.EMBEDDING_MISMATCH in codes
    
    def test_evaluate_borderline(self, verification_config):
        """Test evaluation of borderline match."""
        engine = VerificationDecisionEngine(verification_config)
        
        decision, confidence, codes = engine.evaluate(0.87)
        
        assert decision == "borderline"
        assert ReasonCodes.BORDERLINE_MATCH in codes
    
    def test_evaluate_at_match_threshold(self, verification_config):
        """Test evaluation exactly at match threshold."""
        engine = VerificationDecisionEngine(verification_config)
        threshold = verification_config.match_threshold
        
        decision, _, _ = engine.evaluate(threshold)
        
        assert decision == "match"
    
    def test_evaluate_just_below_match_threshold(self, verification_config):
        """Test evaluation just below match threshold."""
        engine = VerificationDecisionEngine(verification_config)
        threshold = verification_config.match_threshold
        
        decision, _, _ = engine.evaluate(threshold - 0.01)
        
        assert decision == "borderline"
    
    def test_evaluate_at_borderline_threshold(self, verification_config):
        """Test evaluation at borderline lower threshold."""
        engine = VerificationDecisionEngine(verification_config)
        threshold = verification_config.borderline_lower
        
        decision, _, _ = engine.evaluate(threshold)
        
        assert decision == "borderline"
    
    def test_evaluate_just_below_borderline(self, verification_config):
        """Test evaluation just below borderline threshold."""
        engine = VerificationDecisionEngine(verification_config)
        threshold = verification_config.borderline_lower
        
        decision, _, _ = engine.evaluate(threshold - 0.01)
        
        assert decision == "no_match"
    
    def test_evaluate_at_reject_threshold(self, verification_config):
        """Test evaluation at reject threshold."""
        engine = VerificationDecisionEngine(verification_config)
        threshold = verification_config.reject_threshold
        
        decision, _, _ = engine.evaluate(threshold)
        
        assert decision == "no_match"
        # Should still be above complete rejection
    
    def test_evaluate_below_reject_threshold(self, verification_config):
        """Test evaluation below reject threshold."""
        engine = VerificationDecisionEngine(verification_config)
        threshold = verification_config.reject_threshold
        
        decision, _, codes = engine.evaluate(threshold - 0.1)
        
        assert decision == "no_match"
        assert ReasonCodes.EMBEDDING_MISMATCH in codes
    
    def test_evaluate_with_low_quality(self, verification_config):
        """Test evaluation with low quality images."""
        engine = VerificationDecisionEngine(verification_config)
        
        # High similarity but low quality should become borderline
        decision, adjusted, codes = engine.evaluate(
            similarity=0.92,
            probe_quality=0.3,
            reference_quality=0.3,
        )
        
        assert decision == "borderline"
        assert ReasonCodes.LOW_QUALITY_IMAGE in codes
        assert adjusted < 0.92  # Adjusted confidence should be lower
    
    def test_evaluate_quality_factor(self, verification_config):
        """Test quality factor calculation."""
        engine = VerificationDecisionEngine(verification_config)
        
        # Same similarity, different quality
        _, conf_high, _ = engine.evaluate(0.80, probe_quality=1.0, reference_quality=1.0)
        _, conf_low, _ = engine.evaluate(0.80, probe_quality=0.5, reference_quality=0.5)
        
        assert conf_high > conf_low
    
    def test_get_threshold_info(self, verification_config):
        """Test getting threshold information."""
        engine = VerificationDecisionEngine(verification_config)
        
        info = engine.get_threshold_info()
        
        assert "match_threshold" in info
        assert "match_description" in info
        assert "borderline_lower" in info
        assert "reject_threshold" in info


class TestBoundaryThresholds:
    """Tests for boundary threshold behavior."""
    
    @pytest.mark.parametrize("similarity,expected_decision", [
        (1.00, "match"),
        (0.95, "match"),
        (0.90, "match"),
        (0.899, "borderline"),
        (0.85, "borderline"),
        (0.849, "no_match"),
        (0.70, "no_match"),
        (0.50, "no_match"),
        (0.00, "no_match"),
    ])
    def test_decision_boundaries(self, verification_config, similarity, expected_decision):
        """Test decision boundaries with default thresholds."""
        engine = VerificationDecisionEngine(verification_config)
        
        decision, _, _ = engine.evaluate(similarity)
        
        assert decision == expected_decision
    
    def test_high_security_thresholds(self, high_security_config):
        """Test high security configuration thresholds."""
        engine = VerificationDecisionEngine(high_security_config)
        
        # 0.92 is below high-security match threshold (0.95)
        decision, _, _ = engine.evaluate(0.92)
        assert decision == "borderline"
        
        # 0.96 should be a match
        decision, _, _ = engine.evaluate(0.96)
        assert decision == "match"
    
    def test_permissiVIRTENGINE_thresholds(self, permissiVIRTENGINE_config):
        """Test permissive configuration thresholds."""
        engine = VerificationDecisionEngine(permissiVIRTENGINE_config)
        
        # 0.86 should be a match with permissive config
        decision, _, _ = engine.evaluate(0.86)
        assert decision == "match"
