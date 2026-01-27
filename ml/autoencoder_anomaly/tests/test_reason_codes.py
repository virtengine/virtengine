"""
Tests for reason codes.

VE-924: Test reason codes for audit trail.
"""

import pytest

from ml.autoencoder_anomaly.reason_codes import (
    AnomalyReasonCodes,
    ReasonCodeDetails,
    REASON_CODE_DESCRIPTIONS,
    aggregate_reason_codes,
    get_total_score_impact,
)


class TestAnomalyReasonCodes:
    """Tests for AnomalyReasonCodes enum."""
    
    def test_success_codes_exist(self):
        """Test success reason codes exist."""
        assert AnomalyReasonCodes.NO_ANOMALY_DETECTED
        assert AnomalyReasonCodes.RECONSTRUCTION_NORMAL
        assert AnomalyReasonCodes.LATENT_NORMAL
        assert AnomalyReasonCodes.ALL_CHECKS_PASSED
    
    def test_reconstruction_codes_exist(self):
        """Test reconstruction error codes exist."""
        assert AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR
        assert AnomalyReasonCodes.MSE_ABOVE_THRESHOLD
        assert AnomalyReasonCodes.MAE_ABOVE_THRESHOLD
        assert AnomalyReasonCodes.SSIM_BELOW_THRESHOLD
    
    def test_latent_codes_exist(self):
        """Test latent space codes exist."""
        assert AnomalyReasonCodes.LATENT_OUTLIER
        assert AnomalyReasonCodes.LATENT_DISTANCE_HIGH
        assert AnomalyReasonCodes.MAHALANOBIS_DISTANCE_HIGH
    
    def test_quality_codes_exist(self):
        """Test quality issue codes exist."""
        assert AnomalyReasonCodes.IMAGE_TOO_SMALL
        assert AnomalyReasonCodes.IMAGE_TOO_LARGE
        assert AnomalyReasonCodes.INVALID_INPUT
    
    def test_system_codes_exist(self):
        """Test system error codes exist."""
        assert AnomalyReasonCodes.PROCESSING_ERROR
        assert AnomalyReasonCodes.MODEL_ERROR
        assert AnomalyReasonCodes.TIMEOUT_ERROR
    
    def test_get_category_success(self):
        """Test get_category for success codes."""
        category = AnomalyReasonCodes.get_category(
            AnomalyReasonCodes.NO_ANOMALY_DETECTED
        )
        assert category == "success"
    
    def test_get_category_reconstruction(self):
        """Test get_category for reconstruction codes."""
        category = AnomalyReasonCodes.get_category(
            AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR
        )
        assert category == "reconstruction"
    
    def test_get_category_latent(self):
        """Test get_category for latent codes."""
        category = AnomalyReasonCodes.get_category(
            AnomalyReasonCodes.LATENT_OUTLIER
        )
        assert category == "latent"
    
    def test_get_category_quality(self):
        """Test get_category for quality codes."""
        category = AnomalyReasonCodes.get_category(
            AnomalyReasonCodes.IMAGE_TOO_SMALL
        )
        assert category == "quality"
    
    def test_get_category_system(self):
        """Test get_category for system codes."""
        category = AnomalyReasonCodes.get_category(
            AnomalyReasonCodes.PROCESSING_ERROR
        )
        assert category == "system"
    
    def test_is_anomaly_code_true(self):
        """Test is_anomaly_code returns True for anomaly codes."""
        assert AnomalyReasonCodes.is_anomaly_code(
            AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR
        ) is True
        assert AnomalyReasonCodes.is_anomaly_code(
            AnomalyReasonCodes.LATENT_OUTLIER
        ) is True
    
    def test_is_anomaly_code_false_for_success(self):
        """Test is_anomaly_code returns False for success codes."""
        assert AnomalyReasonCodes.is_anomaly_code(
            AnomalyReasonCodes.NO_ANOMALY_DETECTED
        ) is False
        assert AnomalyReasonCodes.is_anomaly_code(
            AnomalyReasonCodes.ALL_CHECKS_PASSED
        ) is False
    
    def test_is_anomaly_code_false_for_quality(self):
        """Test is_anomaly_code returns False for quality codes."""
        assert AnomalyReasonCodes.is_anomaly_code(
            AnomalyReasonCodes.IMAGE_TOO_SMALL
        ) is False
    
    def test_is_anomaly_code_false_for_system(self):
        """Test is_anomaly_code returns False for system codes."""
        assert AnomalyReasonCodes.is_anomaly_code(
            AnomalyReasonCodes.PROCESSING_ERROR
        ) is False


class TestReasonCodeDetails:
    """Tests for ReasonCodeDetails dataclass."""
    
    def test_details_creation(self):
        """Test creating reason code details."""
        details = ReasonCodeDetails(
            code=AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR,
            description="Test description",
            severity="error",
            category="reconstruction",
            score_impact=2000,
        )
        
        assert details.code == AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR
        assert details.description == "Test description"
        assert details.severity == "error"
        assert details.category == "reconstruction"
        assert details.score_impact == 2000
    
    def test_to_dict(self):
        """Test to_dict() conversion."""
        details = ReasonCodeDetails(
            code=AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR,
            description="Test description",
            severity="error",
            category="reconstruction",
            score_impact=2000,
        )
        
        details_dict = details.to_dict()
        
        assert "code" in details_dict
        assert "description" in details_dict
        assert "severity" in details_dict
        assert "category" in details_dict
        assert "score_impact" in details_dict


class TestReasonCodeDescriptions:
    """Tests for REASON_CODE_DESCRIPTIONS dictionary."""
    
    def test_descriptions_populated(self):
        """Test descriptions dictionary is populated."""
        assert len(REASON_CODE_DESCRIPTIONS) > 0
    
    def test_all_descriptions_valid(self):
        """Test all descriptions are valid ReasonCodeDetails."""
        for code, details in REASON_CODE_DESCRIPTIONS.items():
            assert isinstance(details, ReasonCodeDetails)
            assert details.code == code
            assert details.description
            assert details.severity in ["info", "warning", "error"]
    
    def test_success_codes_have_zero_impact(self):
        """Test success codes have zero score impact."""
        success_codes = [
            AnomalyReasonCodes.NO_ANOMALY_DETECTED,
            AnomalyReasonCodes.RECONSTRUCTION_NORMAL,
            AnomalyReasonCodes.LATENT_NORMAL,
            AnomalyReasonCodes.ALL_CHECKS_PASSED,
        ]
        
        for code in success_codes:
            if code in REASON_CODE_DESCRIPTIONS:
                assert REASON_CODE_DESCRIPTIONS[code].score_impact == 0
    
    def test_anomaly_codes_have_positive_impact(self):
        """Test anomaly codes have positive score impact."""
        if AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR in REASON_CODE_DESCRIPTIONS:
            assert REASON_CODE_DESCRIPTIONS[
                AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR
            ].score_impact > 0


class TestAggregateReasonCodes:
    """Tests for aggregate_reason_codes function."""
    
    def test_empty_list(self):
        """Test aggregating empty list."""
        result = aggregate_reason_codes([])
        
        assert result == [AnomalyReasonCodes.NO_ANOMALY_DETECTED.value]
    
    def test_single_code(self):
        """Test aggregating single code."""
        codes = [AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR]
        
        result = aggregate_reason_codes(codes)
        
        assert len(result) == 1
        assert result[0] == AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR.value
    
    def test_multiple_codes(self):
        """Test aggregating multiple codes."""
        codes = [
            AnomalyReasonCodes.MSE_ABOVE_THRESHOLD,
            AnomalyReasonCodes.LATENT_OUTLIER,
            AnomalyReasonCodes.SSIM_BELOW_THRESHOLD,
        ]
        
        result = aggregate_reason_codes(codes)
        
        assert len(result) == 3
    
    def test_max_codes_limit(self):
        """Test max codes limit is respected."""
        codes = [
            AnomalyReasonCodes.MSE_ABOVE_THRESHOLD,
            AnomalyReasonCodes.MAE_ABOVE_THRESHOLD,
            AnomalyReasonCodes.SSIM_BELOW_THRESHOLD,
            AnomalyReasonCodes.LATENT_OUTLIER,
            AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR,
            AnomalyReasonCodes.MULTI_METRIC_ANOMALY,
        ]
        
        result = aggregate_reason_codes(codes, max_codes=3)
        
        assert len(result) == 3
    
    def test_duplicates_removed(self):
        """Test duplicate codes are removed."""
        codes = [
            AnomalyReasonCodes.MSE_ABOVE_THRESHOLD,
            AnomalyReasonCodes.MSE_ABOVE_THRESHOLD,
            AnomalyReasonCodes.MSE_ABOVE_THRESHOLD,
        ]
        
        result = aggregate_reason_codes(codes)
        
        assert len(result) == 1
    
    def test_severity_ordering(self):
        """Test codes are ordered by severity (error first)."""
        codes = [
            AnomalyReasonCodes.NO_ANOMALY_DETECTED,  # info
            AnomalyReasonCodes.MSE_ABOVE_THRESHOLD,  # warning
            AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR,  # error
        ]
        
        result = aggregate_reason_codes(codes)
        
        # Error should come first
        assert result[0] == AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR.value


class TestGetTotalScoreImpact:
    """Tests for get_total_score_impact function."""
    
    def test_empty_list(self):
        """Test impact of empty list."""
        result = get_total_score_impact([])
        
        assert result == 0
    
    def test_success_codes_zero_impact(self):
        """Test success codes have zero impact."""
        codes = [
            AnomalyReasonCodes.NO_ANOMALY_DETECTED,
            AnomalyReasonCodes.ALL_CHECKS_PASSED,
        ]
        
        result = get_total_score_impact(codes)
        
        assert result == 0
    
    def test_anomaly_codes_positive_impact(self):
        """Test anomaly codes have positive impact."""
        codes = [AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR]
        
        result = get_total_score_impact(codes)
        
        assert result > 0
    
    def test_cumulative_impact(self):
        """Test impacts are cumulative."""
        single = get_total_score_impact(
            [AnomalyReasonCodes.MSE_ABOVE_THRESHOLD]
        )
        double = get_total_score_impact([
            AnomalyReasonCodes.MSE_ABOVE_THRESHOLD,
            AnomalyReasonCodes.MAE_ABOVE_THRESHOLD,
        ])
        
        assert double > single
    
    def test_impact_capped(self):
        """Test impact is capped at 10000."""
        # Add many high-impact codes
        codes = [
            AnomalyReasonCodes.HIGH_RECONSTRUCTION_ERROR,
            AnomalyReasonCodes.LATENT_OUTLIER,
            AnomalyReasonCodes.MULTI_METRIC_ANOMALY,
            AnomalyReasonCodes.MAHALANOBIS_DISTANCE_HIGH,
            AnomalyReasonCodes.MSE_ABOVE_THRESHOLD,
            AnomalyReasonCodes.SSIM_BELOW_THRESHOLD,
        ]
        
        result = get_total_score_impact(codes)
        
        assert result <= 10000
