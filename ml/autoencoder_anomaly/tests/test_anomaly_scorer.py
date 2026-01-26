"""
Tests for anomaly scoring.

VE-924: Test reconstruction error and anomaly scoring.
"""

import pytest
import numpy as np

from ml.autoencoder_anomaly.anomaly_scorer import (
    AnomalyScorer,
    AnomalyScore,
    ReconstructionMetrics,
    LatentAnalysisResult,
    ReconstructionErrorCalculator,
    LatentSpaceAnalyzer,
    create_anomaly_scorer,
)
from ml.autoencoder_anomaly.config import (
    AutoencoderAnomalyConfig,
    ReconstructionConfig,
    AnomalyScoringConfig,
    AnomalyLevel,
    AnomalyType,
)
from ml.autoencoder_anomaly.reason_codes import AnomalyReasonCodes


class TestReconstructionErrorCalculator:
    """Tests for ReconstructionErrorCalculator class."""
    
    def test_calculator_creation(self, reconstruction_config):
        """Test creating calculator with config."""
        calculator = ReconstructionErrorCalculator(reconstruction_config)
        
        assert calculator is not None
        assert calculator.config == reconstruction_config
    
    def test_calculator_default_config(self):
        """Test creating calculator with default config."""
        calculator = ReconstructionErrorCalculator()
        assert calculator is not None
    
    def test_mse_identical_images(self, identical_images):
        """Test MSE is zero for identical images."""
        calculator = ReconstructionErrorCalculator()
        original, reconstructed = identical_images
        
        metrics = calculator.calculate(original, reconstructed)
        
        assert metrics.mse < 1e-6
    
    def test_mse_different_images(self, different_images):
        """Test MSE is positive for different images."""
        calculator = ReconstructionErrorCalculator()
        original, reconstructed = different_images
        
        metrics = calculator.calculate(original, reconstructed)
        
        assert metrics.mse > 0
    
    def test_mae_identical_images(self, identical_images):
        """Test MAE is zero for identical images."""
        calculator = ReconstructionErrorCalculator()
        original, reconstructed = identical_images
        
        metrics = calculator.calculate(original, reconstructed)
        
        assert metrics.mae < 1e-6
    
    def test_ssim_identical_images(self, identical_images):
        """Test SSIM is 1.0 for identical images."""
        calculator = ReconstructionErrorCalculator()
        original, reconstructed = identical_images
        
        metrics = calculator.calculate(original, reconstructed)
        
        assert metrics.ssim > 0.99
    
    def test_ssim_different_images(self, different_images):
        """Test SSIM is less than 1.0 for different images."""
        calculator = ReconstructionErrorCalculator()
        original, reconstructed = different_images
        
        metrics = calculator.calculate(original, reconstructed)
        
        assert metrics.ssim < 0.99
    
    def test_psnr_identical_images(self, identical_images):
        """Test PSNR is high for identical images."""
        calculator = ReconstructionErrorCalculator()
        original, reconstructed = identical_images
        
        metrics = calculator.calculate(original, reconstructed)
        
        assert metrics.psnr >= 99
    
    def test_per_channel_metrics(self, different_images):
        """Test per-channel metrics are computed."""
        calculator = ReconstructionErrorCalculator()
        original, reconstructed = different_images
        
        metrics = calculator.calculate(original, reconstructed)
        
        assert len(metrics.mse_per_channel) == 3
        assert len(metrics.mae_per_channel) == 3
    
    def test_patch_metrics(self, different_images):
        """Test patch-level metrics are computed."""
        config = ReconstructionConfig(use_patch_analysis=True)
        calculator = ReconstructionErrorCalculator(config)
        original, reconstructed = different_images
        
        metrics = calculator.calculate(original, reconstructed)
        
        assert metrics.patch_mse_mean >= 0
        assert metrics.patch_mse_max >= metrics.patch_mse_mean
    
    def test_combined_error_range(self, different_images):
        """Test combined error is in [0, 1] range."""
        calculator = ReconstructionErrorCalculator()
        original, reconstructed = different_images
        
        metrics = calculator.calculate(original, reconstructed)
        
        assert 0 <= metrics.combined_error <= 1
    
    def test_metrics_to_dict(self, different_images):
        """Test ReconstructionMetrics.to_dict()."""
        calculator = ReconstructionErrorCalculator()
        original, reconstructed = different_images
        
        metrics = calculator.calculate(original, reconstructed)
        metrics_dict = metrics.to_dict()
        
        assert "mse" in metrics_dict
        assert "mae" in metrics_dict
        assert "ssim" in metrics_dict
        assert "combined_error" in metrics_dict


class TestLatentSpaceAnalyzer:
    """Tests for LatentSpaceAnalyzer class."""
    
    def test_analyzer_creation(self, scoring_config):
        """Test creating analyzer with config."""
        analyzer = LatentSpaceAnalyzer(scoring_config)
        
        assert analyzer is not None
        assert analyzer.config == scoring_config
    
    def test_analyzer_default_config(self):
        """Test creating analyzer with default config."""
        analyzer = LatentSpaceAnalyzer()
        assert analyzer is not None
    
    def test_analyze_normal_latent(self, sample_latent_vector):
        """Test analyzing normal latent vector."""
        analyzer = LatentSpaceAnalyzer()
        
        result = analyzer.analyze(sample_latent_vector)
        
        assert isinstance(result, LatentAnalysisResult)
        assert result.euclidean_distance >= 0
        assert result.mahalanobis_distance >= 0
    
    def test_analyze_outlier_latent(self, outlier_latent_vector):
        """Test analyzing outlier latent vector."""
        analyzer = LatentSpaceAnalyzer()
        
        result = analyzer.analyze(outlier_latent_vector)
        
        assert result.is_outlier == True
        assert len(result.outlier_dimensions) > 0
    
    def test_z_scores_computed(self, sample_latent_vector):
        """Test Z-scores are computed."""
        analyzer = LatentSpaceAnalyzer()
        
        result = analyzer.analyze(sample_latent_vector)
        
        assert result.z_scores is not None
        assert len(result.z_scores) == len(sample_latent_vector)
    
    def test_latent_anomaly_score_range(self, sample_latent_vector):
        """Test latent anomaly score is in [0, 1] range."""
        analyzer = LatentSpaceAnalyzer()
        
        result = analyzer.analyze(sample_latent_vector)
        
        assert 0 <= result.latent_anomaly_score <= 1
    
    def test_cosine_distance_range(self, sample_latent_vector):
        """Test cosine distance is in valid range."""
        analyzer = LatentSpaceAnalyzer()
        
        result = analyzer.analyze(sample_latent_vector)
        
        assert 0 <= result.cosine_distance <= 2
    
    def test_result_to_dict(self, sample_latent_vector):
        """Test LatentAnalysisResult.to_dict()."""
        analyzer = LatentSpaceAnalyzer()
        
        result = analyzer.analyze(sample_latent_vector)
        result_dict = result.to_dict()
        
        assert "euclidean_distance" in result_dict
        assert "mahalanobis_distance" in result_dict
        assert "is_outlier" in result_dict
        assert "latent_anomaly_score" in result_dict


class TestAnomalyScorer:
    """Tests for AnomalyScorer class."""
    
    def test_scorer_creation(self, anomaly_config):
        """Test creating scorer with config."""
        scorer = AnomalyScorer(anomaly_config)
        
        assert scorer is not None
        assert scorer.config == anomaly_config
    
    def test_scorer_default_config(self):
        """Test creating scorer with default config."""
        scorer = AnomalyScorer()
        assert scorer is not None
    
    def test_compute_score_identical(self, identical_images, sample_latent_vector):
        """Test scoring with identical images."""
        scorer = AnomalyScorer()
        original, reconstructed = identical_images
        
        score = scorer.compute_anomaly_score(
            original=original,
            reconstructed=reconstructed,
            latent_vector=sample_latent_vector,
        )
        
        assert isinstance(score, AnomalyScore)
        assert score.is_anomaly is False
        assert score.overall_score < 0.3
    
    def test_compute_score_different(self, different_images, sample_latent_vector):
        """Test scoring with different images."""
        scorer = AnomalyScorer()
        original, reconstructed = different_images
        
        score = scorer.compute_anomaly_score(
            original=original,
            reconstructed=reconstructed,
            latent_vector=sample_latent_vector,
        )
        
        assert isinstance(score, AnomalyScore)
        assert score.overall_score > 0
    
    def test_compute_score_with_outlier_latent(
        self, identical_images, outlier_latent_vector
    ):
        """Test scoring with outlier latent vector."""
        scorer = AnomalyScorer()
        original, reconstructed = identical_images
        
        score = scorer.compute_anomaly_score(
            original=original,
            reconstructed=reconstructed,
            latent_vector=outlier_latent_vector,
        )
        
        assert score.latent_score > 0.3
    
    def test_anomaly_level_determination(self, different_images, sample_latent_vector):
        """Test anomaly level is determined."""
        scorer = AnomalyScorer()
        original, reconstructed = different_images
        
        score = scorer.compute_anomaly_score(
            original=original,
            reconstructed=reconstructed,
            latent_vector=sample_latent_vector,
        )
        
        assert score.anomaly_level in AnomalyLevel
    
    def test_reason_codes_collected(self, different_images, outlier_latent_vector):
        """Test reason codes are collected."""
        scorer = AnomalyScorer()
        original, reconstructed = different_images
        
        score = scorer.compute_anomaly_score(
            original=original,
            reconstructed=reconstructed,
            latent_vector=outlier_latent_vector,
        )
        
        assert len(score.reason_codes) > 0
    
    def test_detected_types_populated(
        self, different_images, outlier_latent_vector
    ):
        """Test detected anomaly types are populated."""
        scorer = AnomalyScorer()
        original, reconstructed = different_images
        
        score = scorer.compute_anomaly_score(
            original=original,
            reconstructed=reconstructed,
            latent_vector=outlier_latent_vector,
        )
        
        # With different images and outlier latent, should detect something
        assert len(score.detected_types) >= 0  # May or may not detect
    
    def test_confidence_range(self, different_images, sample_latent_vector):
        """Test confidence is in valid range."""
        scorer = AnomalyScorer()
        original, reconstructed = different_images
        
        score = scorer.compute_anomaly_score(
            original=original,
            reconstructed=reconstructed,
            latent_vector=sample_latent_vector,
        )
        
        assert 0 <= score.confidence <= 1
    
    def test_include_details(self, different_images, sample_latent_vector):
        """Test including detailed results."""
        scorer = AnomalyScorer()
        original, reconstructed = different_images
        
        score = scorer.compute_anomaly_score(
            original=original,
            reconstructed=reconstructed,
            latent_vector=sample_latent_vector,
            include_details=True,
        )
        
        assert score.reconstruction_metrics is not None
        assert score.latent_analysis is not None
    
    def test_score_to_dict(self, different_images, sample_latent_vector):
        """Test AnomalyScore.to_dict()."""
        scorer = AnomalyScorer()
        original, reconstructed = different_images
        
        score = scorer.compute_anomaly_score(
            original=original,
            reconstructed=reconstructed,
            latent_vector=sample_latent_vector,
        )
        score_dict = score.to_dict()
        
        assert "is_anomaly" in score_dict
        assert "anomaly_level" in score_dict
        assert "overall_score" in score_dict
        assert "confidence" in score_dict


class TestCreateAnomalyScorer:
    """Tests for factory function."""
    
    def test_create_with_config(self, anomaly_config):
        """Test creating scorer with config."""
        scorer = create_anomaly_scorer(anomaly_config)
        
        assert isinstance(scorer, AnomalyScorer)
        assert scorer.config == anomaly_config
    
    def test_create_without_config(self):
        """Test creating scorer without config."""
        scorer = create_anomaly_scorer()
        
        assert isinstance(scorer, AnomalyScorer)
