"""
Tests for artifact analysis.

VE-923: Test artifact analysis functionality.
"""

import pytest
import numpy as np

from ml.gan_detection.artifact_analysis import (
    ArtifactAnalyzer,
    ArtifactResult,
    create_artifact_analyzer,
)
from ml.gan_detection.config import ArtifactAnalysisConfig, GANDetectionConfig, ArtifactType
from ml.gan_detection.reason_codes import GANReasonCodes


class TestArtifactAnalyzer:
    """Tests for ArtifactAnalyzer class."""
    
    def test_analyzer_creation(self, artifact_config):
        """Test creating analyzer with config."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        assert analyzer is not None
        assert analyzer.config == artifact_config
    
    def test_analyzer_default_config(self):
        """Test creating analyzer with default config."""
        analyzer = ArtifactAnalyzer()
        assert analyzer is not None
    
    def test_analyze_valid_image(self, artifact_config, sample_face_frame):
        """Test analysis of valid image."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame)
        
        assert isinstance(result, ArtifactResult)
        assert 0.0 <= result.artifact_score <= 1.0
        assert 0.0 <= result.confidence <= 1.0
    
    def test_analyze_with_face_region(
        self,
        artifact_config,
        sample_face_frame,
        sample_face_region
    ):
        """Test analysis with face region."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame, sample_face_region)
        
        assert isinstance(result, ArtifactResult)
    
    def test_analyze_empty_image(self, artifact_config):
        """Test analysis fails for empty image."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(np.array([]))
        
        assert result.artifact_score == 0.0
        assert GANReasonCodes.INVALID_INPUT in result.reason_codes
    
    def test_analyze_synthetic_image(self, artifact_config, synthetic_frame):
        """Test analysis of synthetic image with checkerboard."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(synthetic_frame)
        
        assert isinstance(result, ArtifactResult)
        # Synthetic frame with checkerboard should have artifacts
    
    def test_frequency_analysis(self, artifact_config, sample_face_frame):
        """Test frequency analysis is performed."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame)
        
        assert 0.0 <= result.frequency_score <= 1.0
    
    def test_checkerboard_detection(self, artifact_config, synthetic_frame):
        """Test checkerboard detection."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(synthetic_frame)
        
        assert 0.0 <= result.checkerboard_score <= 1.0
    
    def test_blending_detection(self, artifact_config, blended_frame):
        """Test blending boundary detection."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(blended_frame)
        
        assert 0.0 <= result.blending_score <= 1.0
    
    def test_texture_analysis(self, artifact_config, sample_face_frame):
        """Test texture consistency analysis."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame)
        
        assert 0.0 <= result.texture_score <= 1.0
    
    def test_color_analysis(self, artifact_config, sample_face_frame):
        """Test color consistency analysis."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame)
        
        assert 0.0 <= result.color_score <= 1.0
    
    def test_compression_detection(self, artifact_config, sample_face_frame):
        """Test compression artifact detection."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame)
        
        assert 0.0 <= result.compression_score <= 1.0
    
    def test_upsampling_detection(self, artifact_config, sample_face_frame):
        """Test upsampling artifact detection."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame)
        
        assert 0.0 <= result.upsampling_score <= 1.0
    
    def test_processing_time_recorded(self, artifact_config, sample_face_frame):
        """Test processing time is recorded."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame)
        
        assert result.processing_time_ms > 0
    
    def test_detected_artifacts_list(self, artifact_config, sample_face_frame):
        """Test detected artifacts are listed."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame)
        
        assert isinstance(result.detected_artifacts, list)
    
    def test_high_quality_frame(self, artifact_config, high_quality_frame):
        """Test analysis of high quality frame."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(high_quality_frame)
        
        assert isinstance(result, ArtifactResult)


class TestArtifactResult:
    """Tests for ArtifactResult dataclass."""
    
    def test_result_to_dict(self):
        """Test converting result to dict."""
        result = ArtifactResult(
            has_artifacts=True,
            artifact_score=0.65,
            confidence=0.8,
            frequency_score=0.4,
            checkerboard_score=0.7,
            blending_score=0.3,
            texture_score=0.5,
            color_score=0.2,
            compression_score=0.1,
            upsampling_score=0.15,
            detected_artifacts=[
                ArtifactType.CHECKERBOARD,
                ArtifactType.TEXTURE_INCONSISTENCY,
            ],
            reason_codes=[
                GANReasonCodes.CHECKERBOARD_ARTIFACT,
                GANReasonCodes.TEXTURE_INCONSISTENCY,
            ],
            processing_time_ms=25.0,
        )
        
        d = result.to_dict()
        
        assert d["has_artifacts"] is True
        assert d["artifact_score"] == 0.65
        assert d["checkerboard_score"] == 0.7
        assert "checkerboard" in d["detected_artifacts"]
        assert "CHECKERBOARD_ARTIFACT" in d["reason_codes"]
    
    def test_result_no_artifacts(self):
        """Test result with no artifacts."""
        result = ArtifactResult(
            has_artifacts=False,
            artifact_score=0.1,
            confidence=0.9,
            detected_artifacts=[],
            reason_codes=[GANReasonCodes.ALL_CHECKS_PASSED],
        )
        
        d = result.to_dict()
        
        assert d["has_artifacts"] is False
        assert len(d["detected_artifacts"]) == 0


class TestFrequencyAnalysis:
    """Tests for frequency domain analysis."""
    
    def test_frequency_on_natural_image(self, artifact_config, sample_face_frame):
        """Test frequency analysis on natural image."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame)
        
        # Natural image should have low frequency anomaly
        assert result.frequency_score < 0.8
    
    def test_frequency_on_synthetic_image(self, artifact_config, synthetic_frame):
        """Test frequency analysis on synthetic image."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(synthetic_frame)
        
        # Synthetic may have frequency anomalies
        assert 0.0 <= result.frequency_score <= 1.0


class TestCheckerboardDetection:
    """Tests for checkerboard artifact detection."""
    
    def test_checkerboard_on_clean_image(self, artifact_config, sample_face_frame):
        """Test checkerboard detection on clean image."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame)
        
        # Checkerboard score should be in valid range [0, 1]
        assert 0.0 <= result.checkerboard_score <= 1.0
    
    def test_checkerboard_on_pattern_image(self, artifact_config, np_random):
        """Test checkerboard detection on pattern image."""
        # Create image with strong checkerboard pattern
        image = np.zeros((224, 224, 3), dtype=np.uint8)
        for y in range(224):
            for x in range(224):
                if (y // 2 + x // 2) % 2 == 0:
                    image[y, x] = [100, 100, 100]
                else:
                    image[y, x] = [150, 150, 150]
        
        analyzer = ArtifactAnalyzer(artifact_config)
        result = analyzer.analyze(image)
        
        # Should detect checkerboard pattern
        assert result.checkerboard_score > 0.0


class TestBlendingDetection:
    """Tests for blending boundary detection."""
    
    def test_blending_on_clean_image(self, artifact_config, sample_face_frame):
        """Test blending detection on clean image."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(sample_face_frame)
        
        assert 0.0 <= result.blending_score <= 1.0
    
    def test_blending_on_manipulated_image(self, artifact_config, blended_frame):
        """Test blending detection on manipulated image."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(blended_frame)
        
        # Blended frame should have detectable boundary
        assert 0.0 <= result.blending_score <= 1.0


class TestTextureAnalysis:
    """Tests for texture consistency analysis."""
    
    def test_texture_on_natural_image(self, artifact_config, high_quality_frame):
        """Test texture analysis on natural image."""
        analyzer = ArtifactAnalyzer(artifact_config)
        
        result = analyzer.analyze(high_quality_frame)
        
        # Natural texture should be consistent
        assert result.texture_score < 0.8


class TestCreateArtifactAnalyzer:
    """Tests for factory function."""
    
    def test_create_without_config(self):
        """Test creating analyzer without config."""
        analyzer = create_artifact_analyzer()
        
        assert isinstance(analyzer, ArtifactAnalyzer)
    
    def test_create_with_config(self, gan_config):
        """Test creating analyzer with full config."""
        analyzer = create_artifact_analyzer(gan_config)
        
        assert isinstance(analyzer, ArtifactAnalyzer)
        assert analyzer.config == gan_config.artifact
